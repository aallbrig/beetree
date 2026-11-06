package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNavigateHintKeepsHelpAndQuitAt80Columns(t *testing.T) {
	visible := navigateHint[:80]
	assert.Contains(t, visible, "[?]help")
	assert.Contains(t, visible, "[q]uit")
	assert.Contains(t, visible, "[s]ave")
	assert.Contains(t, visible, "[r]un")
}
