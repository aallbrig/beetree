package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aallbrig/beetree-cli/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	// Reset flags to defaults between test runs
	generateDryRun = false
	generateOverwrite = false
	generateOutput = ""
	generateAll = false
	renderFormat = "ascii"
	browseTag = ""
	browseSort = "recent"
	pushPublic = true
	pushPrivate = false
	pushTags = nil
	pushDesc = ""
	simulateOverrides = nil
	nodeAddType = "action"
	nodeAddNode = ""
	nodeAddDecorator = ""
	nodeMoveDest = ""
	nodeFilter = ""

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

var generateTestYAML = `
version: "1.0"
metadata:
  name: "test-tree"
  description: "Test behavior tree"
blackboard:
  - name: "target"
    type: "object"
  - name: "health"
    type: "float"
    default: 100.0
tree:
  type: "selector"
  name: "root"
  children:
    - type: "condition"
      name: "has_target"
      node: "HasTarget"
    - type: "action"
      name: "attack"
      node: "Attack"
    - type: "action"
      name: "patrol"
      node: "Patrol"
`

func TestGenerateCommand_Unity(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "test.beetree.yaml")
	os.WriteFile(specFile, []byte(generateTestYAML), 0644)
	outDir := filepath.Join(dir, "output")

	output, err := executeCommand("generate", "unity", specFile, "--output", outDir)
	require.NoError(t, err)
	assert.Contains(t, output, "Generated")

	// Verify files were written
	assert.FileExists(t, filepath.Join(outDir, "TestTreeBlackboard.cs"))
	assert.FileExists(t, filepath.Join(outDir, "TestTreeTreeDefinition.cs"))
	assert.FileExists(t, filepath.Join(outDir, "AttackAction.cs"))
	assert.FileExists(t, filepath.Join(outDir, "PatrolAction.cs"))
	assert.FileExists(t, filepath.Join(outDir, "HasTargetCondition.cs"))
}

func TestGenerateCommand_Unreal(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "test.beetree.yaml")
	os.WriteFile(specFile, []byte(generateTestYAML), 0644)
	outDir := filepath.Join(dir, "output")

	output, err := executeCommand("generate", "unreal", specFile, "--output", outDir)
	require.NoError(t, err)
	assert.Contains(t, output, "Generated")

	assert.FileExists(t, filepath.Join(outDir, "TestTreeBlackboard.h"))
	assert.FileExists(t, filepath.Join(outDir, "BTTask_Attack.h"))
	assert.FileExists(t, filepath.Join(outDir, "BTTask_Attack.cpp"))
}

func TestGenerateCommand_Godot(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "test.beetree.yaml")
	os.WriteFile(specFile, []byte(generateTestYAML), 0644)
	outDir := filepath.Join(dir, "output")

	output, err := executeCommand("generate", "godot", specFile, "--output", outDir)
	require.NoError(t, err)
	assert.Contains(t, output, "Generated")

	assert.FileExists(t, filepath.Join(outDir, "test_tree_blackboard.gd"))
	assert.FileExists(t, filepath.Join(outDir, "attack_action.gd"))
	assert.FileExists(t, filepath.Join(outDir, "patrol_action.gd"))
}

func TestGenerateCommand_DryRun(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "test.beetree.yaml")
	os.WriteFile(specFile, []byte(generateTestYAML), 0644)
	outDir := filepath.Join(dir, "output")

	output, err := executeCommand("generate", "unity", specFile, "--output", outDir, "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, output, "Dry run")

	// Output dir should NOT be created
	_, statErr := os.Stat(outDir)
	assert.True(t, os.IsNotExist(statErr), "dry-run should not create files")
}

func TestGenerateCommand_SkipExistingStubs(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "test.beetree.yaml")
	os.WriteFile(specFile, []byte(generateTestYAML), 0644)
	outDir := filepath.Join(dir, "output")

	// First generate
	_, err := executeCommand("generate", "unity", specFile, "--output", outDir)
	require.NoError(t, err)

	// Modify a stub
	stubPath := filepath.Join(outDir, "AttackAction.cs")
	os.WriteFile(stubPath, []byte("// my custom code"), 0644)

	// Second generate — should skip existing stubs
	output, err := executeCommand("generate", "unity", specFile, "--output", outDir)
	require.NoError(t, err)
	assert.Contains(t, output, "skipped")

	// Stub should still have custom content
	content, _ := os.ReadFile(stubPath)
	assert.Contains(t, string(content), "my custom code")
}

func TestGenerateCommand_Overwrite(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "test.beetree.yaml")
	os.WriteFile(specFile, []byte(generateTestYAML), 0644)
	outDir := filepath.Join(dir, "output")

	// First generate
	_, err := executeCommand("generate", "unity", specFile, "--output", outDir)
	require.NoError(t, err)

	// Modify a stub
	stubPath := filepath.Join(outDir, "AttackAction.cs")
	os.WriteFile(stubPath, []byte("// my custom code"), 0644)

	// Second generate with --overwrite
	_, err = executeCommand("generate", "unity", specFile, "--output", outDir, "--overwrite")
	require.NoError(t, err)

	// Stub should be regenerated
	content, _ := os.ReadFile(stubPath)
	assert.Contains(t, string(content), "EDIT THIS FILE")
}

func TestGenerateCommand_NoArgs(t *testing.T) {
	_, err := executeCommand("generate", "unity")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "specify a .beetree.yaml file or use --all")
}

func TestGenerateCommand_InvalidEngine(t *testing.T) {
	_, err := executeCommand("generate", "invalid", "test.yaml")
	assert.Error(t, err)
}

func TestGenerateCommand_AllFlag(t *testing.T) {
	tmpDir := t.TempDir()
	treesDir := filepath.Join(tmpDir, "trees")
	require.NoError(t, os.MkdirAll(treesDir, 0755))

	specContent := `version: "1.0"
metadata:
  name: patrol
tree:
  type: action
  name: do_patrol
  node: DoPatrol
`
	require.NoError(t, os.WriteFile(filepath.Join(treesDir, "patrol.beetree.yaml"), []byte(specContent), 0644))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	output, err := executeCommand("generate", "unity", "--all", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, output, "Dry run")
}

func TestGenerateCommand_AllFlagNoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	_, err := executeCommand("generate", "unity", "--all")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no .beetree.yaml files found")
}

// --- Registry command tests ---

func setupTestRegistry(t *testing.T) *registry.LocalClient {
	t.Helper()
	tmpDir := t.TempDir()
	client := registry.NewLocalClient(filepath.Join(tmpDir, "registry"))
	registryClientOverride = client
	t.Cleanup(func() { registryClientOverride = nil })
	return client
}

func TestRegistryLogin(t *testing.T) {
	tmpDir := t.TempDir()
	registryClientOverride = registry.NewLocalClient(filepath.Join(tmpDir, "reg"))
	t.Cleanup(func() { registryClientOverride = nil })
	output, err := executeCommand("registry", "login", "test-token")
	require.NoError(t, err)
	assert.Contains(t, output, "Logged in")
}

func TestRegistryLogout(t *testing.T) {
	tmpDir := t.TempDir()
	registryClientOverride = registry.NewLocalClient(filepath.Join(tmpDir, "reg"))
	t.Cleanup(func() { registryClientOverride = nil })
	output, err := executeCommand("registry", "logout")
	require.NoError(t, err)
	assert.Contains(t, output, "Logged out")
}

func TestRegistryBrowseEmpty(t *testing.T) {
	setupTestRegistry(t)
	output, err := executeCommand("registry", "browse")
	require.NoError(t, err)
	assert.Contains(t, output, "No trees found")
}

func TestRegistryBrowseWithTrees(t *testing.T) {
	client := setupTestRegistry(t)
	ctx := context.Background()
	require.NoError(t, client.Login(ctx, "user"))
	_, err := client.Push(ctx, []byte(`version: "1.0"
metadata:
  name: patrol-ai
tree:
  type: action
  name: do_it
  node: DoIt
`), registry.PushOptions{Public: true, Description: "A patrol tree", Tags: []string{"ai"}})
	require.NoError(t, err)

	output, err := executeCommand("registry", "browse")
	require.NoError(t, err)
	assert.Contains(t, output, "patrol-ai")
	assert.Contains(t, output, "1 tree(s) found")
}

func TestRegistrySearch(t *testing.T) {
	client := setupTestRegistry(t)
	ctx := context.Background()
	require.NoError(t, client.Login(ctx, "user"))
	_, err := client.Push(ctx, []byte(`version: "1.0"
metadata:
  name: enemy-combat
tree:
  type: action
  name: fight
  node: Fight
`), registry.PushOptions{Public: true})
	require.NoError(t, err)

	output, err := executeCommand("registry", "search", "combat")
	require.NoError(t, err)
	assert.Contains(t, output, "enemy-combat")
}

func TestRegistrySearchNoResults(t *testing.T) {
	setupTestRegistry(t)
	output, err := executeCommand("registry", "search", "nonexistent")
	require.NoError(t, err)
	assert.Contains(t, output, "No trees found")
}

func TestRegistryPushAndPull(t *testing.T) {
	client := setupTestRegistry(t)
	ctx := context.Background()
	require.NoError(t, client.Login(ctx, "testuser"))

	tmpDir := t.TempDir()
	specFile := filepath.Join(tmpDir, "test.beetree.yaml")
	require.NoError(t, os.WriteFile(specFile, []byte(`version: "1.0"
metadata:
  name: pull-test
tree:
  type: action
  name: do_it
  node: DoIt
`), 0644))

	// Push
	output, err := executeCommand("registry", "push", specFile, "--description", "test tree")
	require.NoError(t, err)
	assert.Contains(t, output, "Published")
	assert.Contains(t, output, "pull-test")

	// Pull
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	output, err = executeCommand("registry", "pull", "testuser/pull-test")
	require.NoError(t, err)
	assert.Contains(t, output, "Pulled")

	// Verify downloaded file
	pulled, err := os.ReadFile(filepath.Join(tmpDir, "trees", "pull-test.beetree.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(pulled), "pull-test")
}

func TestRegistryPushNotAuthenticated(t *testing.T) {
	setupTestRegistry(t)

	tmpDir := t.TempDir()
	specFile := filepath.Join(tmpDir, "test.beetree.yaml")
	require.NoError(t, os.WriteFile(specFile, []byte(`version: "1.0"
metadata:
  name: test
tree:
  type: action
  name: do_it
  node: DoIt
`), 0644))

	_, err := executeCommand("registry", "push", specFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not authenticated")
}

// --- Simulate command tests ---

func TestSimulateCommand(t *testing.T) {
	tmpDir := t.TempDir()
	specFile := filepath.Join(tmpDir, "test.beetree.yaml")
	require.NoError(t, os.WriteFile(specFile, []byte(`version: "1.0"
metadata:
  name: test
tree:
  type: sequence
  name: root
  children:
    - type: action
      name: patrol
      node: Patrol
`), 0644))

	output, err := executeCommand("simulate", specFile)
	require.NoError(t, err)
	assert.Contains(t, output, "root")
	assert.Contains(t, output, "SUCCESS")
	assert.Contains(t, output, "patrol")
}

func TestSimulateCommand_WithOverride(t *testing.T) {
	tmpDir := t.TempDir()
	specFile := filepath.Join(tmpDir, "test.beetree.yaml")
	require.NoError(t, os.WriteFile(specFile, []byte(`version: "1.0"
metadata:
  name: test
tree:
  type: sequence
  name: root
  children:
    - type: condition
      name: check
      node: Check
    - type: action
      name: do_it
      node: DoIt
`), 0644))

	output, err := executeCommand("simulate", specFile, "--override", "check=FAILURE")
	require.NoError(t, err)
	assert.Contains(t, output, "FAILURE")
}

// --- Diff command tests ---

func TestDiffCommand_Identical(t *testing.T) {
	tmpDir := t.TempDir()
	specA := filepath.Join(tmpDir, "a.beetree.yaml")
	require.NoError(t, os.WriteFile(specA, []byte(`version: "1.0"
metadata:
  name: test
tree:
  type: action
  name: patrol
  node: Patrol
`), 0644))

	output, err := executeCommand("diff", specA, specA)
	require.NoError(t, err)
	assert.Contains(t, output, "No differences")
}

func TestDiffCommand_WithChanges(t *testing.T) {
	tmpDir := t.TempDir()
	specA := filepath.Join(tmpDir, "a.beetree.yaml")
	require.NoError(t, os.WriteFile(specA, []byte(`version: "1.0"
metadata:
  name: test
  description: old
tree:
  type: action
  name: patrol
  node: Patrol
`), 0644))

	specB := filepath.Join(tmpDir, "b.beetree.yaml")
	require.NoError(t, os.WriteFile(specB, []byte(`version: "1.0"
metadata:
  name: test
  description: new
tree:
  type: action
  name: patrol
  node: Patrol
`), 0644))

	output, err := executeCommand("diff", specA, specB)
	require.NoError(t, err)
	assert.Contains(t, output, "metadata.description")
	assert.Contains(t, output, "old")
	assert.Contains(t, output, "new")
}

// --- Doctor command tests ---

func TestDoctorCommand(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	output, err := executeCommand("doctor")
	// Should warn about missing trees/ but not fail hard
	assert.Error(t, err) // 1 issue: missing trees/
	assert.Contains(t, output, "BeeTree Doctor")
	assert.Contains(t, output, "Go version")
}

func TestDoctorCommand_WithValidSpec(t *testing.T) {
	tmpDir := t.TempDir()
	treesDir := filepath.Join(tmpDir, "trees")
	require.NoError(t, os.MkdirAll(treesDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(treesDir, "test.beetree.yaml"), []byte(`version: "1.0"
metadata:
  name: test
tree:
  type: action
  name: patrol
  node: Patrol
`), 0644))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	output, err := executeCommand("doctor")
	require.NoError(t, err)
	assert.Contains(t, output, "All checks passed")
	assert.Contains(t, output, "test.beetree.yaml: valid")
}

// --- Node edit command tests ---

func writeTestSpec(t *testing.T) string {
t.Helper()
tmpDir := t.TempDir()
specFile := filepath.Join(tmpDir, "test.beetree.yaml")
require.NoError(t, os.WriteFile(specFile, []byte(`version: "1.0"
metadata:
  name: test
tree:
  type: selector
  name: root
  children:
    - type: sequence
      name: combat
      children:
        - type: condition
          name: has_target
          node: HasTarget
        - type: action
          name: attack
          node: Attack
    - type: action
      name: patrol
      node: Patrol
`), 0644))
return specFile
}

func TestNodeAdd(t *testing.T) {
specFile := writeTestSpec(t)
output, err := executeCommand("node", "add", specFile, "combat", "reload", "--type", "action", "--node", "Reload")
require.NoError(t, err)
assert.Contains(t, output, "Added")
assert.Contains(t, output, "reload")

// Verify the file was updated
data, err := os.ReadFile(specFile)
require.NoError(t, err)
assert.Contains(t, string(data), "reload")
assert.Contains(t, string(data), "Reload")
}

func TestNodeRemove(t *testing.T) {
specFile := writeTestSpec(t)
output, err := executeCommand("node", "remove", specFile, "patrol")
require.NoError(t, err)
assert.Contains(t, output, "Removed")

data, err := os.ReadFile(specFile)
require.NoError(t, err)
assert.NotContains(t, string(data), "patrol")
}

func TestNodeMove(t *testing.T) {
specFile := writeTestSpec(t)
output, err := executeCommand("node", "move", specFile, "patrol", "--to", "combat")
require.NoError(t, err)
assert.Contains(t, output, "Moved")
assert.Contains(t, output, "patrol")
assert.Contains(t, output, "combat")

data, err := os.ReadFile(specFile)
require.NoError(t, err)
// patrol should now be nested under combat in the YAML
assert.Contains(t, string(data), "patrol")
}

func TestNodeAdd_DuplicateFails(t *testing.T) {
specFile := writeTestSpec(t)
_, err := executeCommand("node", "add", specFile, "root", "patrol", "--type", "action")
assert.Error(t, err)
assert.Contains(t, err.Error(), "already exists")
}

func TestNodeRemove_RootFails(t *testing.T) {
specFile := writeTestSpec(t)
_, err := executeCommand("node", "remove", specFile, "root")
assert.Error(t, err)
assert.Contains(t, err.Error(), "cannot remove root")
}

func TestBuilderCommand_Help(t *testing.T) {
	output, err := executeCommand("builder", "--help")
	require.NoError(t, err)
	assert.Contains(t, output, "Launch the interactive behavior tree editor")
	assert.Contains(t, output, "Add child node")
	assert.Contains(t, output, "Edit selected node")
	assert.Contains(t, output, "Delete selected node")
	assert.Contains(t, output, "Move node")
	assert.Contains(t, output, "Undo last change")
	assert.Contains(t, output, "Save to file")
	assert.Contains(t, output, "confirms if unsaved")
	assert.Contains(t, output, "[file]")
}
