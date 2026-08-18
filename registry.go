package dregistry

import (
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// FactoryFunc constructs a [Resolver] from a DSN and logger.
type FactoryFunc func(config string, logger *zap.Logger) (Resolver, error)

var registry = make(map[string]FactoryFunc)

// Register associates a DSN scheme with a resolver factory.
func Register(scheme string, factory FactoryFunc) {
	registry[scheme] = factory
}

// New builds a [Resolver] from a DSN. Multiple DSNs may be comma-separated;
// they are tried in order and [ErrNotFound] falls through to the next.
//
// Commas that sit inside a later `scheme://` token are treated as separators;
// commas inside inline JSON are left alone.
func New(config string, logger *zap.Logger) (Resolver, error) {
	if strings.TrimSpace(config) == "" {
		return nil, fmt.Errorf("empty registry DSN")
	}
	if logger == nil {
		logger = zlog
	}

	parts := splitDSNs(config)
	resolvers := make([]Resolver, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("empty registry DSN in chain")
		}
		resolver, err := newOne(part, logger)
		if err != nil {
			return nil, err
		}
		resolvers = append(resolvers, resolver)
	}
	if len(resolvers) == 1 {
		return resolvers[0], nil
	}
	return Chain(resolvers...), nil
}

func newOne(config string, logger *zap.Logger) (Resolver, error) {
	scheme, _, ok := strings.Cut(config, "://")
	if !ok || scheme == "" {
		return nil, fmt.Errorf("registry DSN %q is missing a scheme", config)
	}

	factory := registry[scheme]
	if factory == nil {
		return nil, fmt.Errorf("no registry plugin named %q is currently registered", scheme)
	}
	return factory(config, logger)
}

var dsnSplitRe = regexp.MustCompile(`,([a-zA-Z][a-zA-Z0-9+.-]*://)`)

func splitDSNs(config string) []string {
	matches := dsnSplitRe.FindAllStringIndex(config, -1)
	if matches == nil {
		return []string{config}
	}

	parts := make([]string, 0, len(matches)+1)
	last := 0
	for _, match := range matches {
		parts = append(parts, config[last:match[0]])
		last = match[0] + 1
	}
	return append(parts, config[last:])
}
