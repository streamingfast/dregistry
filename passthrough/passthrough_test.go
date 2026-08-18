package passthrough

import (
	"os"
	"testing"

	"github.com/streamingfast/dregistry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	Register()
	os.Exit(m.Run())
}

func TestResolve_InfersTLSFromPortAndScheme(t *testing.T) {
	resolver, err := dregistry.New("passthrough://", zap.NewNop())
	require.NoError(t, err)

	tests := []struct {
		identifier string
		want       *dregistry.Endpoint
	}{
		{
			identifier: "stores.example.com:9000",
			want:       &dregistry.Endpoint{Address: "stores.example.com:9000"},
		},
		{
			identifier: "stores.example.com:443",
			want:       &dregistry.Endpoint{Address: "stores.example.com:443", TLS: true},
		},
		{
			identifier: "grpcs://stores.example.com:9000",
			want:       &dregistry.Endpoint{Address: "stores.example.com:9000", TLS: true},
		},
		{
			identifier: "grpc://stores.example.com:9000",
			want:       &dregistry.Endpoint{Address: "stores.example.com:9000"},
		},
		{
			identifier: "https://stores.example.com",
			want:       &dregistry.Endpoint{Address: "stores.example.com", TLS: true},
		},
	}

	for _, test := range tests {
		t.Run(test.identifier, func(t *testing.T) {
			got, err := resolver.Resolve(t.Context(), test.identifier)
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestResolve_EmptyIdentifier(t *testing.T) {
	resolver, err := dregistry.New("passthrough://", zap.NewNop())
	require.NoError(t, err)

	_, err = resolver.Resolve(t.Context(), "")
	require.Error(t, err)
	require.NotErrorIs(t, err, dregistry.ErrNotFound)
}
