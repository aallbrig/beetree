// Package treeedit provides functions to programmatically modify
// behavior tree specs: add, remove, and move nodes.
package treeedit

import (
	"fmt"
	"os"

	"github.com/aallbrig/beetree-cli/internal/model"
	"gopkg.in/yaml.v3"
)

// FindNode recursively searches for a node by name, returning a pointer to it.
func FindNode(root *model.NodeSpec, name string) *model.NodeSpec {
	if root.Name == name {
		return root
	}
	for i := range root.Children {
		if found := FindNode(&root.Children[i], name); found != nil {
			return found
		}
	}
	return nil
}

// CollectNames returns all node names in the tree.
func CollectNames(root *model.NodeSpec) map[string]bool {
	names := make(map[string]bool)
	collectNamesRecursive(root, names)
	return names
}

func collectNamesRecursive(node *model.NodeSpec, names map[string]bool) {
	names[node.Name] = true
	for i := range node.Children {
		collectNamesRecursive(&node.Children[i], names)
	}
}

// AddNode appends a child node to the parent identified by parentName.
func AddNode(root *model.NodeSpec, parentName string, child model.NodeSpec) error {
	parent := FindNode(root, parentName)
	if parent == nil {
		return fmt.Errorf("parent node %q not found", parentName)
	}

	if model.IsLeafType(parent.Type) {
		return fmt.Errorf("cannot add children to %s node %q (leaf type)", parent.Type, parent.Name)
	}

	// Check for duplicate names
	names := CollectNames(root)
	if names[child.Name] {
		return fmt.Errorf("node %q already exists in tree", child.Name)
	}

	parent.Children = append(parent.Children, child)
	return nil
}

// RemoveNode removes a node (and its subtree) from the tree by name.
func RemoveNode(root *model.NodeSpec, name string) error {
	if root.Name == name {
		return fmt.Errorf("cannot remove root node")
	}
	return removeNodeRecursive(root, name)
}

func removeNodeRecursive(parent *model.NodeSpec, name string) error {
	for i := range parent.Children {
		if parent.Children[i].Name == name {
			parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
			return nil
		}
		if err := removeNodeRecursive(&parent.Children[i], name); err == nil {
			return nil
		}
	}
	return fmt.Errorf("node %q not found", name)
}

// MoveNode detaches a node from its current location and attaches it to newParentName.
func MoveNode(root *model.NodeSpec, nodeName, newParentName string) error {
	node := FindNode(root, nodeName)
	if node == nil {
		return fmt.Errorf("node %q not found", nodeName)
	}

	newParent := FindNode(root, newParentName)
	if newParent == nil {
		return fmt.Errorf("target parent %q not found", newParentName)
	}

	if model.IsLeafType(newParent.Type) {
		return fmt.Errorf("cannot add children to %s node %q (leaf type)", newParent.Type, newParent.Name)
	}

	// Copy the node before removing (remove invalidates the pointer)
	nodeCopy := *node
	if err := RemoveNode(root, nodeName); err != nil {
		return err
	}

	newParent.Children = append(newParent.Children, nodeCopy)
	return nil
}

// UpdateNode applies field changes to a node identified by name.
// If newName is non-empty and different, the node is renamed (checking for duplicates).
func UpdateNode(root *model.NodeSpec, name string, updates NodeUpdates) error {
	node := FindNode(root, name)
	if node == nil {
		return fmt.Errorf("node %q not found", name)
	}

	if updates.Name != "" && updates.Name != node.Name {
		names := CollectNames(root)
		if names[updates.Name] {
			return fmt.Errorf("node %q already exists in tree", updates.Name)
		}
		node.Name = updates.Name
	}
	if updates.Type != nil {
		node.Type = *updates.Type
	}
	if updates.NodeClass != nil {
		node.Node = *updates.NodeClass
	}
	if updates.Decorator != nil {
		node.Decorator = *updates.Decorator
	}
	return nil
}

// NodeUpdates holds optional field changes for UpdateNode.
// Pointer fields are only applied when non-nil.
type NodeUpdates struct {
	Name      string  // new name (empty = keep current)
	Type      *string // new type
	NodeClass *string // new node class
	Decorator *string // new decorator
}

// SaveSpec writes a TreeSpec back to a YAML file.
func SaveSpec(s *model.TreeSpec, path string) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal spec: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
