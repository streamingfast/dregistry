package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisters_UnknownPlugin(t *testing.T) {
	err := Registers("not-a-plugin")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not-a-plugin")
}

func TestRegisters_KnownPlugins(t *testing.T) {
	require.NoError(t, Registers("noop", "json", "controlplane", "passthrough"))
}
