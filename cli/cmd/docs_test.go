package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteReferenceCoversEveryVisibleCommand(t *testing.T) {
	var buf bytes.Buffer
	assert.NoError(t, writeReference(&buf, rootCmd))
	out := buf.String()
	assert.Contains(t, out, "title: \"CLI reference\"")
	for _, name := range []string{"beetree validate", "beetree simulate", "beetree generate", "beetree render", "beetree builder"} {
		assert.Contains(t, out, "`"+name+"`", name)
	}
	assert.NotContains(t, out, "`beetree docs`", "hidden commands stay out of the reference")
	assert.NotContains(t, out, "`beetree completion`")
}
