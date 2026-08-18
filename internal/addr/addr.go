package addr

import (
	"fmt"
	"strings"

	"github.com/streamingfast/dregistry"
)

// Endpoint builds an endpoint from a free-form address, inferring TLS from
// an https/grpcs scheme or a :443 port. grpc/http force plaintext.
func Endpoint(address string, authRequired bool) (*dregistry.Endpoint, error) {
	if address == "" {
		return nil, fmt.Errorf("empty address")
	}
	return &dregistry.Endpoint{
		Address:      HostPort(address),
		TLS:          UseTLS(address),
		AuthRequired: authRequired,
	}, nil
}

// UseTLS reports whether address should be dialed with TLS.
func UseTLS(address string) bool {
	lower := strings.ToLower(address)
	switch {
	case strings.HasPrefix(lower, "https://"), strings.HasPrefix(lower, "grpcs://"):
		return true
	case strings.HasPrefix(lower, "http://"), strings.HasPrefix(lower, "grpc://"):
		return false
	default:
		return strings.HasSuffix(address, ":443")
	}
}

// HostPort strips a known scheme and any path, leaving host[:port].
func HostPort(address string) string {
	rest := address
	lower := strings.ToLower(address)
	switch {
	case strings.HasPrefix(lower, "https://"):
		rest = address[len("https://"):]
	case strings.HasPrefix(lower, "http://"):
		rest = address[len("http://"):]
	case strings.HasPrefix(lower, "grpcs://"):
		rest = address[len("grpcs://"):]
	case strings.HasPrefix(lower, "grpc://"):
		rest = address[len("grpc://"):]
	default:
		return address
	}
	host, _, _ := strings.Cut(rest, "/")
	return host
}
