package controlplane

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/streamingfast/dregistry"
	pbregistry "github.com/streamingfast/dregistry/pb/sf/registry/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	Register()
	os.Exit(m.Run())
}

type fakeRegistry struct {
	pbregistry.UnimplementedFoundationStoreRegistryServiceServer
	byID map[string]*pbregistry.GetFoundationStoreResponse
}

func (f *fakeRegistry) GetFoundationStore(_ context.Context, req *pbregistry.GetFoundationStoreRequest) (*pbregistry.GetFoundationStoreResponse, error) {
	if resp, ok := f.byID[req.DeploymentId]; ok {
		return resp, nil
	}
	return &pbregistry.GetFoundationStoreResponse{Found: false, Message: "not found"}, nil
}

func startFakeRegistry(t *testing.T, svc *fakeRegistry) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	pbregistry.RegisterFoundationStoreRegistryServiceServer(server, svc)

	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(server.Stop)

	return listener.Addr().String()
}

func TestResolve_PrefersInternalEndpoint(t *testing.T) {
	addr := startFakeRegistry(t, &fakeRegistry{
		byID: map[string]*pbregistry.GetFoundationStoreResponse{
			"store-a": {
				Found: true,
				Entry: &pbregistry.FoundationStoreEntry{
					DeploymentId:     "store-a",
					Endpoint:         "public.example.com:443",
					Tls:              true,
					AuthRequired:     true,
					InternalEndpoint: "internal.example.com:9000",
					InternalTls:      false,
				},
			},
		},
	})

	resolver, err := dregistry.New("controlplane://"+addr+"?insecure=true", zap.NewNop())
	require.NoError(t, err)

	got, err := resolver.Resolve(t.Context(), "store-a")
	require.NoError(t, err)
	assert.Equal(t, &dregistry.Endpoint{
		Address:      "internal.example.com:9000",
		TLS:          false,
		AuthRequired: true,
	}, got)
}

func TestResolve_UsesPublicWhenInternalEmpty(t *testing.T) {
	addr := startFakeRegistry(t, &fakeRegistry{
		byID: map[string]*pbregistry.GetFoundationStoreResponse{
			"store-a": {
				Found: true,
				Entry: &pbregistry.FoundationStoreEntry{
					DeploymentId: "store-a",
					Endpoint:     "public.example.com:443",
					Tls:          true,
					AuthRequired: true,
				},
			},
		},
	})

	resolver, err := dregistry.New("controlplane://"+addr+"?insecure=true", zap.NewNop())
	require.NoError(t, err)

	got, err := resolver.Resolve(t.Context(), "store-a")
	require.NoError(t, err)
	assert.Equal(t, &dregistry.Endpoint{
		Address:      "public.example.com:443",
		TLS:          true,
		AuthRequired: true,
	}, got)
}

func TestResolve_NotFound(t *testing.T) {
	addr := startFakeRegistry(t, &fakeRegistry{})

	resolver, err := dregistry.New("controlplane://"+addr+"?insecure=true", zap.NewNop())
	require.NoError(t, err)

	_, err = resolver.Resolve(t.Context(), "missing")
	require.ErrorIs(t, err, dregistry.ErrNotFound)
}

func TestResolve_FoundWithoutEntry(t *testing.T) {
	addr := startFakeRegistry(t, &fakeRegistry{
		byID: map[string]*pbregistry.GetFoundationStoreResponse{
			"store-a": {Found: true},
		},
	})

	resolver, err := dregistry.New("controlplane://"+addr+"?insecure=true", zap.NewNop())
	require.NoError(t, err)

	_, err = resolver.Resolve(t.Context(), "store-a")
	require.ErrorIs(t, err, dregistry.ErrNotFound)
}

func TestResolve_DoesNotStripVersionSuffix(t *testing.T) {
	var gotID string
	addr := startRecordingRegistry(t, func(id string) *pbregistry.GetFoundationStoreResponse {
		gotID = id
		return &pbregistry.GetFoundationStoreResponse{Found: false}
	})

	resolver, err := dregistry.New("controlplane://"+addr+"?insecure=true", zap.NewNop())
	require.NoError(t, err)

	_, err = resolver.Resolve(t.Context(), "store-a@v1.2.3")
	require.ErrorIs(t, err, dregistry.ErrNotFound)
	assert.Equal(t, "store-a@v1.2.3", gotID)
}

type recordingRegistry struct {
	pbregistry.UnimplementedFoundationStoreRegistryServiceServer
	onGet func(string) *pbregistry.GetFoundationStoreResponse
}

func (r *recordingRegistry) GetFoundationStore(_ context.Context, req *pbregistry.GetFoundationStoreRequest) (*pbregistry.GetFoundationStoreResponse, error) {
	if r.onGet != nil {
		if resp := r.onGet(req.DeploymentId); resp != nil {
			return resp, nil
		}
	}
	return &pbregistry.GetFoundationStoreResponse{Found: false}, nil
}

func startRecordingRegistry(t *testing.T, onGet func(string) *pbregistry.GetFoundationStoreResponse) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	pbregistry.RegisterFoundationStoreRegistryServiceServer(server, &recordingRegistry{onGet: onGet})

	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(server.Stop)

	return listener.Addr().String()
}

func TestParseConfig_PortAndInsecure(t *testing.T) {
	tests := []struct {
		url      string
		insecure bool
	}{
		{url: "controlplane://registry.example.com:443", insecure: false},
		{url: "controlplane://registry.example.com:9000", insecure: true},
		{url: "controlplane://registry.example.com:1443", insecure: true},
		{url: "controlplane://[::1]:443", insecure: false},
		{url: "controlplane://registry.example.com:443?insecure=true", insecure: true},
		{url: "controlplane://registry.example.com:9000?insecure=false", insecure: false},
	}

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			cfg, err := parseConfig(test.url)
			require.NoError(t, err)
			assert.Equal(t, test.insecure, cfg.insecure)
		})
	}
}
