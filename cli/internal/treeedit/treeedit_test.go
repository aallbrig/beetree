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

// --- UpdateNode ---

func TestUpdateNode_Rename(t *testing.T) {
	s := sampleSpec()
	err := UpdateNode(&s.Tree, "patrol", NodeUpdates{Name: "guard"})
	require.NoError(t, err)

	assert.Nil(t, FindNode(&s.Tree, "patrol"))
	node := FindNode(&s.Tree, "guard")
	require.NotNil(t, node)
	assert.Equal(t, "action", node.Type)
	assert.Equal(t, "Patrol", node.Node)
}

func TestUpdateNode_RenameDuplicate(t *testing.T) {
	s := sampleSpec()
	err := UpdateNode(&s.Tree, "patrol", NodeUpdates{Name: "combat"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestUpdateNode_ChangeType(t *testing.T) {
	s := sampleSpec()
	newType := "condition"
	err := UpdateNode(&s.Tree, "patrol", NodeUpdates{Type: &newType})
	require.NoError(t, err)

	node := FindNode(&s.Tree, "patrol")
	assert.Equal(t, "condition", node.Type)
}

func TestUpdateNode_ChangeNodeClass(t *testing.T) {
	s := sampleSpec()
	newClass := "GuardPatrol"
	err := UpdateNode(&s.Tree, "patrol", NodeUpdates{NodeClass: &newClass})
	require.NoError(t, err)

	node := FindNode(&s.Tree, "patrol")
	assert.Equal(t, "GuardPatrol", node.Node)
}

func TestUpdateNode_ChangeDecorator(t *testing.T) {
	s := sampleSpec()
	dec := "negate"
	err := UpdateNode(&s.Tree, "patrol", NodeUpdates{Decorator: &dec})
	require.NoError(t, err)

	node := FindNode(&s.Tree, "patrol")
	assert.Equal(t, "negate", node.Decorator)
}

func TestUpdateNode_NotFound(t *testing.T) {
	s := sampleSpec()
	err := UpdateNode(&s.Tree, "nonexistent", NodeUpdates{Name: "x"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdateNode_MultipleFields(t *testing.T) {
	s := sampleSpec()
	newType := "condition"
	newClass := "CheckPatrol"
	err := UpdateNode(&s.Tree, "patrol", NodeUpdates{
		Name:      "check_patrol",
		Type:      &newType,
		NodeClass: &newClass,
	})
	require.NoError(t, err)

	node := FindNode(&s.Tree, "check_patrol")
	require.NotNil(t, node)
	assert.Equal(t, "condition", node.Type)
	assert.Equal(t, "CheckPatrol", node.Node)
}

func TestCloneNode_Leaf(t *testing.T) {
	node := &model.NodeSpec{
		Type: "action", Name: "attack", Node: "Attack",
		Parameters: map[string]interface{}{"damage": 10, "range": 5.5},
	}
	existing := map[string]bool{"attack": true}

	clone := CloneNode(node, existing)
	assert.Equal(t, "attack_copy", clone.Name)
	assert.Equal(t, "action", clone.Type)
	assert.Equal(t, "Attack", clone.Node)
	assert.Equal(t, 10, clone.Parameters["damage"])

	// Mutating clone params shouldn't affect original
	clone.Parameters["damage"] = 99
	assert.Equal(t, 10, node.Parameters["damage"])
}

func TestCloneNode_Subtree(t *testing.T) {
	s := sampleSpec()
	combat := FindNode(&s.Tree, "combat")
	require.NotNil(t, combat)

	existing := CollectNames(&s.Tree)
	clone := CloneNode(combat, existing)

	assert.Equal(t, "combat_copy", clone.Name)
	require.Len(t, clone.Children, 2)
	assert.Equal(t, "has_target_copy", clone.Children[0].Name)
	assert.Equal(t, "attack_copy", clone.Children[1].Name)

	// All cloned names should be registered in existing
	assert.True(t, existing["combat_copy"])
	assert.True(t, existing["has_target_copy"])
	assert.True(t, existing["attack_copy"])
}

func TestCloneNode_UniqueNameCollision(t *testing.T) {
	node := &model.NodeSpec{Type: "action", Name: "attack"}
	existing := map[string]bool{"attack": true, "attack_copy": true}

	clone := CloneNode(node, existing)
	assert.Equal(t, "attack_copy2", clone.Name)
}
