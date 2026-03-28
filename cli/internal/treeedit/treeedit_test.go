package treeedit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleSpec() *model.TreeSpec {
	return &model.TreeSpec{
		Version:  "1.0",
		Metadata: model.Metadata{Name: "test"},
		Tree: model.NodeSpec{
			Type: "selector",
			Name: "root",
			Children: []model.NodeSpec{
				{
					Type: "sequence",
					Name: "combat",
					Children: []model.NodeSpec{
						{Type: "condition", Name: "has_target", Node: "HasTarget"},
						{Type: "action", Name: "attack", Node: "Attack"},
					},
				},
				{Type: "action", Name: "patrol", Node: "Patrol"},
			},
		},
	}
}

// --- FindNode ---

func TestFindNode_Root(t *testing.T) {
	s := sampleSpec()
	node := FindNode(&s.Tree, "root")
	require.NotNil(t, node)
	assert.Equal(t, "selector", node.Type)
}

func TestFindNode_DeepChild(t *testing.T) {
	s := sampleSpec()
	node := FindNode(&s.Tree, "attack")
	require.NotNil(t, node)
	assert.Equal(t, "action", node.Type)
}

func TestFindNode_NotFound(t *testing.T) {
	s := sampleSpec()
	node := FindNode(&s.Tree, "nonexistent")
	assert.Nil(t, node)
}

// --- AddNode ---

func TestAddNode_ToComposite(t *testing.T) {
	s := sampleSpec()
	err := AddNode(&s.Tree, "combat", model.NodeSpec{
		Type: "action",
		Name: "reload",
		Node: "Reload",
	})
	require.NoError(t, err)

	parent := FindNode(&s.Tree, "combat")
	require.NotNil(t, parent)
	assert.Len(t, parent.Children, 3)
	assert.Equal(t, "reload", parent.Children[2].Name)
}

func TestAddNode_ToLeafFails(t *testing.T) {
	s := sampleSpec()
	err := AddNode(&s.Tree, "patrol", model.NodeSpec{
		Type: "action",
		Name: "scout",
		Node: "Scout",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot add children")
}

func TestAddNode_ParentNotFound(t *testing.T) {
	s := sampleSpec()
	err := AddNode(&s.Tree, "nonexistent", model.NodeSpec{
		Type: "action",
		Name: "scout",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAddNode_DuplicateName(t *testing.T) {
	s := sampleSpec()
	err := AddNode(&s.Tree, "root", model.NodeSpec{
		Type: "action",
		Name: "patrol", // already exists
		Node: "Patrol2",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// --- RemoveNode ---

func TestRemoveNode_Leaf(t *testing.T) {
	s := sampleSpec()
	err := RemoveNode(&s.Tree, "patrol")
	require.NoError(t, err)
	assert.Len(t, s.Tree.Children, 1)
}

func TestRemoveNode_Subtree(t *testing.T) {
	s := sampleSpec()
	err := RemoveNode(&s.Tree, "combat")
	require.NoError(t, err)
	assert.Len(t, s.Tree.Children, 1)
	assert.Equal(t, "patrol", s.Tree.Children[0].Name)
}

func TestRemoveNode_Root(t *testing.T) {
	s := sampleSpec()
	err := RemoveNode(&s.Tree, "root")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot remove root")
}

func TestRemoveNode_NotFound(t *testing.T) {
	s := sampleSpec()
	err := RemoveNode(&s.Tree, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// --- MoveNode ---

func TestMoveNode_ToNewParent(t *testing.T) {
	s := sampleSpec()
	err := MoveNode(&s.Tree, "patrol", "combat")
	require.NoError(t, err)

	// patrol should now be under combat
	combat := FindNode(&s.Tree, "combat")
	require.NotNil(t, combat)
	assert.Len(t, combat.Children, 3)
	assert.Equal(t, "patrol", combat.Children[2].Name)

	// root should have only combat
	assert.Len(t, s.Tree.Children, 1)
}

func TestMoveNode_ToLeafFails(t *testing.T) {
	s := sampleSpec()
	err := MoveNode(&s.Tree, "patrol", "attack")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot add children")
}

func TestMoveNode_NotFound(t *testing.T) {
	s := sampleSpec()
	err := MoveNode(&s.Tree, "nonexistent", "root")
	assert.Error(t, err)
}

// --- SaveSpec (roundtrip) ---

func TestSaveSpec_Roundtrip(t *testing.T) {
	s := sampleSpec()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.beetree.yaml")

	require.NoError(t, SaveSpec(s, path))

	loaded, err := spec.ParseFile(path)
	require.NoError(t, err)
	assert.Equal(t, "test", loaded.Metadata.Name)
	assert.Equal(t, "selector", loaded.Tree.Type)
	assert.Len(t, loaded.Tree.Children, 2)
}

func TestSaveSpec_AfterEdit(t *testing.T) {
	s := sampleSpec()
	require.NoError(t, AddNode(&s.Tree, "root", model.NodeSpec{
		Type: "action",
		Name: "flee",
		Node: "Flee",
	}))

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.beetree.yaml")
	require.NoError(t, SaveSpec(s, path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "flee")
	assert.Contains(t, string(data), "Flee")
}

// --- CollectNames ---

func TestCollectNames(t *testing.T) {
	s := sampleSpec()
	names := CollectNames(&s.Tree)
	assert.Contains(t, names, "root")
	assert.Contains(t, names, "combat")
	assert.Contains(t, names, "has_target")
	assert.Contains(t, names, "attack")
	assert.Contains(t, names, "patrol")
	assert.Len(t, names, 5)
}
