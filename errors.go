package dregistry

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when a resolver has no endpoint for the identifier.
// Chain skips this error and tries the next resolver. Other errors stop the chain.
var ErrNotFound = errors.New("registry: not found")

// IsNotFound reports whether err is or wraps [ErrNotFound].
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func fmtNotFound(identifier string) error {
	return fmt.Errorf("%w: %q", ErrNotFound, identifier)
}
