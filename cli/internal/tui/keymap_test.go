package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocsPageListsEveryKeyBinding(t *testing.T) {
	page, err := os.ReadFile(filepath.Join("..", "..", "..", "website", "content", "docs", "tui.md"))
	require.NoError(t, err)
	for _, kb := range KeyBindings {
		row := "| " + kb.Mode + " | `" + kb.Key + "` | " + kb.Action + " |"
		assert.Contains(t, string(page), row, "docs/tui.md is missing the %s key %s", kb.Mode, kb.Key)
	}
}

func TestHelpOverlayMentionsEveryKeyBinding(t *testing.T) {
	for _, kb := range KeyBindings {
		if kb.Mode == "Simulate" && kb.Key == "q" {
			continue // only offered once a run completes
		}
		key := strings.ReplaceAll(kb.Key, "/", " ")
		if strings.Contains(kb.Key, "/") && len(kb.Key) > 1 {
			key = map[string]string{"Up/Down": "↑ ↓", "Left/Right": "← →"}[kb.Key]
		}
		assert.Contains(t, helpText, key, "help overlay is missing %s", kb.Key)
	}
}
