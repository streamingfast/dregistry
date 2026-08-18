package json

import (
	"cmp"
	"context"
	stdjson "encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/streamingfast/dregistry"
	"github.com/streamingfast/dregistry/internal/addr"
	"go.uber.org/zap"
)

// Register installs the json:// scheme. Call this explicitly; importing
// the package does not register the plugin.
//
// The DSN is either a file path (`json:///etc/stores.json`) or inline JSON
// (`json://{"id":"stores.example.com:443"}`).
func Register() {
	dregistry.Register("json", func(config string, _ *zap.Logger) (dregistry.Resolver, error) {
		return newResolver(config)
	})
}

type resolver struct {
	endpoints map[string]*dregistry.Endpoint
}

func newResolver(config string) (*resolver, error) {
	payload, err := loadJSON(config)
	if err != nil {
		return nil, err
	}

	endpoints, err := parseEndpoints(payload)
	if err != nil {
		return nil, err
	}
	return &resolver{endpoints: endpoints}, nil
}

func loadJSON(config string) ([]byte, error) {
	rest, ok := strings.CutPrefix(config, "json://")
	if !ok {
		return nil, fmt.Errorf("invalid json registry DSN %q", config)
	}
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return nil, fmt.Errorf("json registry DSN %q is missing a file path or inline object", config)
	}
	if strings.HasPrefix(rest, "{") {
		return []byte(rest), nil
	}
	data, err := os.ReadFile(rest)
	if err != nil {
		return nil, fmt.Errorf("read json registry %q: %w", rest, err)
	}
	return data, nil
}

func parseEndpoints(data []byte) (map[string]*dregistry.Endpoint, error) {
	var raw map[string]stdjson.RawMessage
	if err := stdjson.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse json registry: %w", err)
	}

	out := make(map[string]*dregistry.Endpoint, len(raw))
	for id, msg := range raw {
		endpoint, err := parseEndpoint(msg)
		if err != nil {
			return nil, fmt.Errorf("json registry entry %q: %w", id, err)
		}
		out[id] = endpoint
	}
	return out, nil
}

type richEndpoint struct {
	Address      string `json:"address"`
	Endpoint     string `json:"endpoint"`
	TLS          *bool  `json:"tls"`
	AuthRequired bool   `json:"auth_required"`
}

func parseEndpoint(msg stdjson.RawMessage) (*dregistry.Endpoint, error) {
	var asString string
	if err := stdjson.Unmarshal(msg, &asString); err == nil {
		return addr.Endpoint(asString, false)
	}

	var obj richEndpoint
	if err := stdjson.Unmarshal(msg, &obj); err != nil {
		return nil, err
	}
	address := cmp.Or(obj.Address, obj.Endpoint)
	if address == "" {
		return nil, fmt.Errorf("missing address")
	}
	endpoint, err := addr.Endpoint(address, obj.AuthRequired)
	if err != nil {
		return nil, err
	}
	if obj.TLS != nil {
		endpoint.TLS = *obj.TLS
	}
	return endpoint, nil
}

func (r *resolver) Resolve(_ context.Context, identifier string) (*dregistry.Endpoint, error) {
	endpoint, ok := r.endpoints[identifier]
	if !ok {
		return nil, fmt.Errorf("%w: %q", dregistry.ErrNotFound, identifier)
	}
	cp := *endpoint
	return &cp, nil
}
