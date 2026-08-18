package json

import (
	"os"
	"path/filepath"
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

func TestResolve_HitStringValueInfersTLS(t *testing.T) {
	path := writeJSON(t, `{
		"store-plain": "stores.example.com:9000",
		"store-tls": "stores.example.com:443",
		"store-grpcs": "grpcs://stores.example.com:9000"
	}`)

	resolver, err := dregistry.New("json://"+path, zap.NewNop())
	require.NoError(t, err)

	got, err := resolver.Resolve(t.Context(), "store-plain")
	require.NoError(t, err)
	assert.Equal(t, &dregistry.Endpoint{Address: "stores.example.com:9000"}, got)

	got, err = resolver.Resolve(t.Context(), "store-tls")
	require.NoError(t, err)
	assert.Equal(t, &dregistry.Endpoint{Address: "stores.example.com:443", TLS: true}, got)

	got, err = resolver.Resolve(t.Context(), "store-grpcs")
	require.NoError(t, err)
	assert.Equal(t, &dregistry.Endpoint{Address: "stores.example.com:9000", TLS: true}, got)
}

func TestResolve_HitRichObject(t *testing.T) {
	path := writeJSON(t, `{
		"store-a": {
			"address": "internal.example.com:9000",
			"tls": false,
			"auth_required": true
		}
	}`)

	resolver, err := dregistry.New("json://"+path, zap.NewNop())
	require.NoError(t, err)

	got, err := resolver.Resolve(t.Context(), "store-a")
	require.NoError(t, err)
	assert.Equal(t, &dregistry.Endpoint{
		Address:      "internal.example.com:9000",
		TLS:          false,
		AuthRequired: true,
	}, got)
}

func TestResolve_Miss(t *testing.T) {
	path := writeJSON(t, `{"store-a": "stores.example.com:9000"}`)

	resolver, err := dregistry.New("json://"+path, zap.NewNop())
	require.NoError(t, err)

	_, err = resolver.Resolve(t.Context(), "missing")
	require.ErrorIs(t, err, dregistry.ErrNotFound)
}

func TestNew_InlineJSON(t *testing.T) {
	resolver, err := dregistry.New(`json://{"store-a":"stores.example.com:9000"}`, zap.NewNop())
	require.NoError(t, err)

	got, err := resolver.Resolve(t.Context(), "store-a")
	require.NoError(t, err)
	assert.Equal(t, "stores.example.com:9000", got.Address)
}

func TestNew_MissingFile(t *testing.T) {
	_, err := dregistry.New("json:///no/such/foundational-stores.json", zap.NewNop())
	require.Error(t, err)
}

func writeJSON(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "stores.json")
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	return path
}
