package tui

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestNavigateHintKeepsHelpAndQuitAt80Columns(t *testing.T) {
	visible := navigateHint[:80]
	assert.Contains(t, visible, "[?]help")
	assert.Contains(t, visible, "[q]uit")
	assert.Contains(t, visible, "[s]ave")
	assert.Contains(t, visible, "[r]un")
}

func TestStatusBarKeysSurviveTviewColourTags(t *testing.T) {
	tv := tview.NewTextView().SetDynamicColors(true)
	tv.SetText(tview.Escape(navigateHint))
	got := tv.GetText(true)
	for _, want := range []string{"[?]help", "[q]uit", "[s]ave", "[r]un", "[u]ndo", "[a]dd"} {
		assert.Contains(t, got, want)
	}
}

func TestSimulateStatusKeysSurviveTviewColourTags(t *testing.T) {
	tv := tview.NewTextView().SetDynamicColors(true)
	tv.SetText("[aqua]SIM[-] [green]" + tview.Escape("[S]") + "[-]uccess  " + tview.Escape("[Esc]") + " stop")
	got := tv.GetText(true)
	assert.Contains(t, got, "[S]uccess")
	assert.Contains(t, got, "[Esc] stop")
}
