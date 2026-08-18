package dregistry

import (
	"context"
	"errors"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNew_UnknownScheme(t *testing.T) {
	_, err := New("unknown://example", zap.NewNop())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown")
}

func TestNew_InvalidURL(t *testing.T) {
	_, err := New("://bad", zap.NewNop())
	require.Error(t, err)
}

func TestNew_EmptyConfig(t *testing.T) {
	_, err := New("", zap.NewNop())
	require.Error(t, err)
}

func TestRegisterAndNew(t *testing.T) {
	restore := snapshotRegistry(t)

	Register("stub", func(config string, logger *zap.Logger) (Resolver, error) {
		assert.Equal(t, "stub://ok", config)
		assert.NotNil(t, logger)
		return staticResolver{ep: &Endpoint{Address: "example.com:9000"}}, nil
	})

	resolver, err := New("stub://ok", zap.NewNop())
	require.NoError(t, err)

	got, err := resolver.Resolve(t.Context(), "any")
	require.NoError(t, err)
	assert.Equal(t, &Endpoint{Address: "example.com:9000"}, got)

	restore()
}

func TestNew_CommaSeparatedChain(t *testing.T) {
	restore := snapshotRegistry(t)

	Register("miss", func(string, *zap.Logger) (Resolver, error) {
		return notFoundResolver{}, nil
	})
	Register("hit", func(string, *zap.Logger) (Resolver, error) {
		return staticResolver{ep: &Endpoint{Address: "hit.example.com:443", TLS: true}}, nil
	})

	resolver, err := New("miss://first,hit://second", zap.NewNop())
	require.NoError(t, err)

	got, err := resolver.Resolve(t.Context(), "id")
	require.NoError(t, err)
	assert.Equal(t, "hit.example.com:443", got.Address)
	assert.True(t, got.TLS)

	restore()
}

func TestSplitDSNs_KeepsCommasInsideInlineJSON(t *testing.T) {
	got := splitDSNs(`json://{"a":"x:1","b":"y:2"},passthrough://`)
	assert.Equal(t, []string{`json://{"a":"x:1","b":"y:2"}`, "passthrough://"}, got)
}

func TestChain_FallsThroughNotFound(t *testing.T) {
	resolver := Chain(
		notFoundResolver{},
		staticResolver{ep: &Endpoint{Address: "second.example.com:9000"}},
	)

	got, err := resolver.Resolve(t.Context(), "id")
	require.NoError(t, err)
	assert.Equal(t, "second.example.com:9000", got.Address)
}

func TestChain_StopsOnHardError(t *testing.T) {
	hard := errors.New("boom")
	resolver := Chain(
		errResolver{err: hard},
		staticResolver{ep: &Endpoint{Address: "should-not-run.example.com:9000"}},
	)

	_, err := resolver.Resolve(t.Context(), "id")
	require.ErrorIs(t, err, hard)
}

func TestChain_AllMissReturnsNotFound(t *testing.T) {
	resolver := Chain(notFoundResolver{}, notFoundResolver{})

	_, err := resolver.Resolve(t.Context(), "missing")
	require.ErrorIs(t, err, ErrNotFound)
	assert.True(t, IsNotFound(err))
}

type staticResolver struct {
	ep *Endpoint
}

func (s staticResolver) Resolve(context.Context, string) (*Endpoint, error) {
	return s.ep, nil
}

type notFoundResolver struct{}

func (notFoundResolver) Resolve(_ context.Context, identifier string) (*Endpoint, error) {
	return nil, fmtNotFound(identifier)
}

type errResolver struct {
	err error
}

func (e errResolver) Resolve(context.Context, string) (*Endpoint, error) {
	return nil, e.err
}

func snapshotRegistry(t *testing.T) func() {
	t.Helper()
	original := maps.Clone(registry)
	return func() {
		registry = original
	}
}
