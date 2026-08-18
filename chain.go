package dregistry

import (
	"context"
	"errors"
)

type chain []Resolver

// Chain returns a resolver that tries each resolver in order.
// [ErrNotFound] continues to the next resolver; any other error stops the chain.
func Chain(resolvers ...Resolver) Resolver {
	return chain(resolvers)
}

func (c chain) Resolve(ctx context.Context, identifier string) (*Endpoint, error) {
	var lastNotFound error
	for _, resolver := range c {
		endpoint, err := resolver.Resolve(ctx, identifier)
		if err == nil {
			return endpoint, nil
		}
		if errors.Is(err, ErrNotFound) {
			lastNotFound = err
			continue
		}
		return nil, err
	}
	if lastNotFound != nil {
		return nil, lastNotFound
	}
	return nil, fmtNotFound(identifier)
}
