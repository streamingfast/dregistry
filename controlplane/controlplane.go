package controlplane

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/streamingfast/dgrpc"
	"github.com/streamingfast/dregistry"
	pbregistry "github.com/streamingfast/dregistry/pb/sf/registry/v1"
	"go.uber.org/zap"
)

// Register installs the controlplane:// scheme. Call this explicitly; importing
// the package does not register the plugin.
//
// DSN: controlplane://host:port[?insecure=true|false]
//
// The registry connection is insecure unless the URL port is 443.
// Set insecure=true/false to override.
func Register() {
	dregistry.Register("controlplane", func(config string, logger *zap.Logger) (dregistry.Resolver, error) {
		return newResolver(config, logger)
	})
}

type config struct {
	endpoint string
	insecure bool
}

type resolver struct {
	client pbregistry.FoundationStoreRegistryServiceClient
	logger *zap.Logger
}

func newResolver(configURL string, logger *zap.Logger) (*resolver, error) {
	cfg, err := parseConfig(configURL)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = zlog
	}

	creds, err := dgrpc.WithAutoTransportCredentials(false, cfg.insecure, false)
	if err != nil {
		return nil, fmt.Errorf("control-plane transport credentials: %w", err)
	}

	conn, err := dgrpc.NewClientConn(cfg.endpoint, creds)
	if err != nil {
		return nil, fmt.Errorf("dial control-plane registry %q: %w", cfg.endpoint, err)
	}

	logger.Info("setting up control-plane registry resolver",
		zap.String("endpoint", cfg.endpoint),
		zap.Bool("insecure", cfg.insecure),
	)

	return &resolver{
		client: pbregistry.NewFoundationStoreRegistryServiceClient(conn),
		logger: logger,
	}, nil
}

func parseConfig(configURL string) (*config, error) {
	u, err := url.Parse(configURL)
	if err != nil {
		return nil, fmt.Errorf("parse control-plane DSN: %w", err)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("control-plane DSN %q is missing host:port", configURL)
	}

	cfg := &config{
		endpoint: u.Host,
		insecure: u.Port() != "443",
	}

	if raw := u.Query().Get("insecure"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid insecure query param %q: %w", raw, err)
		}
		cfg.insecure = parsed
	}

	return cfg, nil
}

func (r *resolver) Resolve(ctx context.Context, identifier string) (*dregistry.Endpoint, error) {
	resp, err := r.client.GetFoundationStore(ctx, &pbregistry.GetFoundationStoreRequest{
		DeploymentId: identifier,
	})
	if err != nil {
		return nil, fmt.Errorf("control-plane GetFoundationStore %q: %w", identifier, err)
	}
	if resp == nil || !resp.Found || resp.Entry == nil {
		return nil, fmt.Errorf("%w: %q", dregistry.ErrNotFound, identifier)
	}

	entry := resp.Entry
	address := entry.Endpoint
	useTLS := entry.Tls
	if entry.InternalEndpoint != "" {
		address = entry.InternalEndpoint
		useTLS = entry.InternalTls
	}
	if address == "" {
		return nil, fmt.Errorf("%w: %q", dregistry.ErrNotFound, identifier)
	}

	endpoint := &dregistry.Endpoint{
		Address:      address,
		TLS:          useTLS,
		AuthRequired: entry.AuthRequired,
	}
	r.logger.Debug("resolved foundation store via control plane",
		zap.String("identifier", identifier),
		zap.String("address", endpoint.Address),
		zap.Bool("tls", endpoint.TLS),
		zap.Bool("auth_required", endpoint.AuthRequired),
	)
	return endpoint, nil
}
