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

func TestInitCommand(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	output, err := executeCommand("init", "--name", "my-project")
	require.NoError(t, err)
	assert.Contains(t, output, "my-project")

	// Check created files and directories
	assert.FileExists(t, filepath.Join(dir, "beetree.yaml"))
	assert.DirExists(t, filepath.Join(dir, "trees"))
	assert.DirExists(t, filepath.Join(dir, "subtrees"))

	// Verify manifest content
	data, err := os.ReadFile(filepath.Join(dir, "beetree.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "my-project")
	assert.Contains(t, string(data), "version:")
}

func TestInitCommand_DefaultName(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	output, err := executeCommand("init")
	require.NoError(t, err)
	assert.Contains(t, output, "Initialized")
	assert.FileExists(t, filepath.Join(dir, "beetree.yaml"))
}

func TestInitCommand_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile(filepath.Join(dir, "beetree.yaml"), []byte("existing"), 0644)

	_, err := executeCommand("init")
	assert.Error(t, err)
}

func TestNewCommand(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.MkdirAll(filepath.Join(dir, "trees"), 0755)
	os.Chdir(dir)
	defer os.Chdir(origDir)

	output, err := executeCommand("new", "patrol-ai")
	require.NoError(t, err)
	assert.Contains(t, output, "patrol-ai")
	assert.FileExists(t, filepath.Join(dir, "trees", "patrol-ai.beetree.yaml"))

	data, err := os.ReadFile(filepath.Join(dir, "trees", "patrol-ai.beetree.yaml"))
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "patrol-ai")
	assert.Contains(t, content, "version:")
	assert.Contains(t, content, "selector")
}

func TestNewCommand_NoArgs(t *testing.T) {
	_, err := executeCommand("new")
	assert.Error(t, err)
}

func TestNewCommand_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	treesDir := filepath.Join(dir, "trees")
	os.MkdirAll(treesDir, 0755)
	os.WriteFile(filepath.Join(treesDir, "existing.beetree.yaml"), []byte("x"), 0644)
	os.Chdir(dir)
	defer os.Chdir(origDir)

	_, err := executeCommand("new", "existing")
	assert.Error(t, err)
}

func TestNodeListCommand(t *testing.T) {
	output, err := executeCommand("node", "list")
	require.NoError(t, err)

	// Should list all core types
	assert.Contains(t, output, "action")
	assert.Contains(t, output, "condition")
	assert.Contains(t, output, "sequence")
	assert.Contains(t, output, "selector")
	assert.Contains(t, output, "parallel")
	assert.Contains(t, output, "decorator")

	// Should list extensions
	assert.Contains(t, output, "utility_selector")
}

func TestNodeListCommand_FilterCore(t *testing.T) {
	output, err := executeCommand("node", "list", "--filter", "core")
	require.NoError(t, err)
	assert.Contains(t, output, "action")
	assert.NotContains(t, output, "utility_selector")
}

func TestNodeListCommand_FilterExtension(t *testing.T) {
	output, err := executeCommand("node", "list", "--filter", "extension")
	require.NoError(t, err)
	assert.Contains(t, output, "utility_selector")
	assert.NotContains(t, output, "CORE")
}

func TestRenderCommand_YAML(t *testing.T) {
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
      name: "attack"
      node: "Attack"
    - type: "action"
      name: "patrol"
      node: "Patrol"
`
	os.WriteFile(specFile, []byte(yamlContent), 0644)

	output, err := executeCommand("render", specFile)
	require.NoError(t, err)
	assert.Contains(t, output, "SEL")
	assert.Contains(t, output, "root")
	assert.Contains(t, output, "attack")
	assert.Contains(t, output, "patrol")
}

func TestRenderCommand_Mermaid(t *testing.T) {
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
      name: "attack"
      node: "Attack"
`
	os.WriteFile(specFile, []byte(yamlContent), 0644)

	output, err := executeCommand("render", specFile, "--format", "mermaid")
	require.NoError(t, err)
	assert.Contains(t, output, "graph TD")
	assert.Contains(t, output, "-->")
}
