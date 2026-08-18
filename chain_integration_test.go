package dregistry_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/streamingfast/dregistry"
	pbregistry "github.com/streamingfast/dregistry/pb/sf/registry/v1"
	"github.com/streamingfast/dregistry/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	if err := plugins.Registers("json", "controlplane", "passthrough"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestNew_ThreeTierChainFallsThrough(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stores.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"from-json":"json.example.com:9000"}`), 0o644))

	resolver, err := dregistry.New("json://"+path+",passthrough://", zap.NewNop())
	require.NoError(t, err)

	got, err := resolver.Resolve(t.Context(), "from-json")
	require.NoError(t, err)
	assert.Equal(t, "json.example.com:9000", got.Address)

	got, err = resolver.Resolve(t.Context(), "passthrough.example.com:443")
	require.NoError(t, err)
	assert.Equal(t, &dregistry.Endpoint{
		Address: "passthrough.example.com:443",
		TLS:     true,
	}, got)
}

func TestNew_JSONThenControlPlaneThenPassthrough(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stores.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"from-json":"json.example.com:9000"}`), 0o644))

	addr := startFakeRegistry(t, map[string]*pbregistry.FoundationStoreEntry{
		"from-control-plane": {
			DeploymentId:     "from-control-plane",
			Endpoint:         "public.example.com:443",
			Tls:              true,
			InternalEndpoint: "internal.example.com:9000",
			InternalTls:      false,
			AuthRequired:     true,
		},
	})

	resolver, err := dregistry.New(
		"json://"+path+",controlplane://"+addr+"?insecure=true,passthrough://",
		zap.NewNop(),
	)
	require.NoError(t, err)

	got, err := resolver.Resolve(t.Context(), "from-json")
	require.NoError(t, err)
	assert.Equal(t, "json.example.com:9000", got.Address)

	got, err = resolver.Resolve(t.Context(), "from-control-plane")
	require.NoError(t, err)
	assert.Equal(t, &dregistry.Endpoint{
		Address:      "internal.example.com:9000",
		AuthRequired: true,
	}, got)

	got, err = resolver.Resolve(t.Context(), "direct.example.com:443")
	require.NoError(t, err)
	assert.Equal(t, &dregistry.Endpoint{Address: "direct.example.com:443", TLS: true}, got)
}

type fakeRegistry struct {
	pbregistry.UnimplementedFoundationStoreRegistryServiceServer
	entries map[string]*pbregistry.FoundationStoreEntry
}

func (f *fakeRegistry) GetFoundationStore(_ context.Context, req *pbregistry.GetFoundationStoreRequest) (*pbregistry.GetFoundationStoreResponse, error) {
	entry, ok := f.entries[req.DeploymentId]
	if !ok {
		return &pbregistry.GetFoundationStoreResponse{Found: false}, nil
	}
	return &pbregistry.GetFoundationStoreResponse{Found: true, Entry: entry}, nil
}

func startFakeRegistry(t *testing.T, entries map[string]*pbregistry.FoundationStoreEntry) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	pbregistry.RegisterFoundationStoreRegistryServiceServer(server, &fakeRegistry{entries: entries})
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	return listener.Addr().String()
}
