package passthrough

import (
	"context"
	"fmt"

	"github.com/streamingfast/dregistry"
	"github.com/streamingfast/dregistry/internal/addr"
	"go.uber.org/zap"
)

// Register installs the passthrough:// scheme. Call this explicitly; importing
// the package does not register the plugin.
// The identifier itself is treated as the endpoint address.
func Register() {
	dregistry.Register("passthrough", func(string, *zap.Logger) (dregistry.Resolver, error) {
		return resolver{}, nil
	})
}

type resolver struct{}

func (resolver) Resolve(_ context.Context, identifier string) (*dregistry.Endpoint, error) {
	if identifier == "" {
		return nil, fmt.Errorf("passthrough: empty identifier")
	}
	return addr.Endpoint(identifier, false)
}
