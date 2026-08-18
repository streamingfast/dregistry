package plugins

import (
	"fmt"

	"github.com/streamingfast/dregistry/controlplane"
	"github.com/streamingfast/dregistry/json"
	"github.com/streamingfast/dregistry/noop"
	"github.com/streamingfast/dregistry/passthrough"
)

// Registers enables the named built-in plugins. Each id is a DSN scheme
// (`noop`, `json`, `controlplane`, `passthrough`).
func Registers(ids ...string) error {
	for _, id := range ids {
		switch id {
		case "noop":
			noop.Register()
		case "json":
			json.Register()
		case "controlplane":
			controlplane.Register()
		case "passthrough":
			passthrough.Register()
		default:
			return fmt.Errorf("unknown registry plugin %q", id)
		}
	}
	return nil
}
