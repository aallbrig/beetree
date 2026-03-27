package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestVersionCommand(t *testing.T) {
	output, err := executeCommand("version")
	require.NoError(t, err)
	assert.Contains(t, output, "beetree")
}

func TestValidateCommand_ValidFile(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "test.beetree.yaml")
	yamlContent := `
version: "1.0"
metadata:
  name: "test-tree"
tree:
  type: "selector"
  name: "root"
  children:
    - type: "action"
      name: "do_something"
      node: "DoSomething"
`
	err := os.WriteFile(specFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	output, err := executeCommand("validate", specFile)
	require.NoError(t, err)
	assert.Contains(t, output, "valid")
}

func TestValidateCommand_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "bad.beetree.yaml")
	yamlContent := `
version: "1.0"
metadata:
  name: "test-tree"
tree:
  type: "bogus"
  name: "root"
`
	err := os.WriteFile(specFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	output, _ := executeCommand("validate", specFile)
	assert.True(t, strings.Contains(output, "error") || strings.Contains(output, "Error") || strings.Contains(output, "unknown"))
}

func TestValidateCommand_NoArgs(t *testing.T) {
	_, err := executeCommand("validate")
	assert.Error(t, err)
}

func TestValidateCommand_FileNotFound(t *testing.T) {
	_, err := executeCommand("validate", "/nonexistent/file.yaml")
	assert.Error(t, err)
}
