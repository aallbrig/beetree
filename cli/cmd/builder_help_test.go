package cmd

import (
	"testing"

	"github.com/aallbrig/beetree-cli/internal/tui"
	"github.com/stretchr/testify/assert"
)

func TestBuilderHelpListsEveryKeyBinding(t *testing.T) {
	for _, kb := range tui.KeyBindings {
		assert.Contains(t, builderCmd.Long, kb.Action, "mode %s key %s", kb.Mode, kb.Key)
	}
}
