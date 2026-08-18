package noop

import (
	"testing"

	"github.com/streamingfast/dregistry"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestResolve_AlwaysNotFound(t *testing.T) {
	Register()

	resolver, err := dregistry.New("noop://", zap.NewNop())
	require.NoError(t, err)

	_, err = resolver.Resolve(t.Context(), "anything")
	require.ErrorIs(t, err, dregistry.ErrNotFound)
}
