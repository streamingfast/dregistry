package noop

import (
	"context"
	"fmt"

	"github.com/streamingfast/dregistry"
	"go.uber.org/zap"
)

// Register installs the noop:// scheme. Call this explicitly; importing
// the package does not register the plugin.
func Register() {
	dregistry.Register("noop", func(string, *zap.Logger) (dregistry.Resolver, error) {
		return resolver{}, nil
	})
}

type resolver struct{}

func (resolver) Resolve(_ context.Context, identifier string) (*dregistry.Endpoint, error) {
	return nil, fmt.Errorf("%w: %q", dregistry.ErrNotFound, identifier)
}
