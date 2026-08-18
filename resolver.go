package dregistry

import "context"

// Endpoint is a resolved foundational-store gRPC target.
type Endpoint struct {
	Address      string
	TLS          bool
	AuthRequired bool
}

// Resolver maps an identifier to a store endpoint.
type Resolver interface {
	Resolve(ctx context.Context, identifier string) (*Endpoint, error)
}
