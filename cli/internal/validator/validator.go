package validator

import (
	"fmt"

	"github.com/aallbrig/beetree-cli/internal/model"
)

// Validate checks a TreeSpec for structural and semantic errors.
// Returns a slice of errors found (empty if valid).
func Validate(spec *model.TreeSpec) []error {
	var errs []error

	if spec.Version == "" {
		errs = append(errs, fmt.Errorf("version is required"))
	}
	if spec.Metadata.Name == "" {
		errs = append(errs, fmt.Errorf("metadata.name is required"))
	}

	errs = append(errs, validateBlackboard(spec.Blackboard)...)
	errs = append(errs, validateCustomNodes(spec.CustomNodes)...)
	errs = append(errs, validateSubtrees(spec)...)
	errs = append(errs, validateNode(&spec.Tree, "tree")...)

	return errs
}

func validateBlackboard(vars []model.BlackboardVar) []error {
	var errs []error
	seen := make(map[string]bool)
	for _, v := range vars {
		if seen[v.Name] {
			errs = append(errs, fmt.Errorf("duplicate blackboard variable: %q", v.Name))
		}
		seen[v.Name] = true
	}
	return errs
}

func validateCustomNodes(nodes []model.CustomNodeDef) []error {
	var errs []error
	for _, n := range nodes {
		if n.Type != "action" && n.Type != "condition" {
			errs = append(errs, fmt.Errorf("custom node %q: type must be 'action' or 'condition', got %q", n.Name, n.Type))
		}
	}
	return errs
}

func validateSubtrees(spec *model.TreeSpec) []error {
	var errs []error
	seen := make(map[string]bool)
	for _, st := range spec.Subtrees {
		if st.Name == "" {
			errs = append(errs, fmt.Errorf("subtree: name is required"))
		}
		if st.File == "" {
			errs = append(errs, fmt.Errorf("subtree %q: file is required", st.Name))
		}
		if seen[st.Name] {
			errs = append(errs, fmt.Errorf("subtree %q: duplicate name", st.Name))
		}
		seen[st.Name] = true
	}

	// Validate subtree node references match defined subtrees
	subtreeNames := make(map[string]bool)
	for _, st := range spec.Subtrees {
		subtreeNames[st.Name] = true
	}
	validateSubtreeRefs(&spec.Tree, subtreeNames, "tree", &errs)

	return errs
}

func validateSubtreeRefs(node *model.NodeSpec, subtreeNames map[string]bool, path string, errs *[]error) {
	if node.Type == "subtree" && node.Ref != "" {
		if !subtreeNames[node.Ref] {
			*errs = append(*errs, fmt.Errorf("%s: subtree ref %q not found in subtrees list", path, node.Ref))
		}
	}
	for i := range node.Children {
		childPath := fmt.Sprintf("%s.children[%d]", path, i)
		validateSubtreeRefs(&node.Children[i], subtreeNames, childPath, errs)
	}
}

func validateNode(node *model.NodeSpec, path string) []error {
	var errs []error

	if node.Name == "" {
		errs = append(errs, fmt.Errorf("%s: name is required", path))
	}

	if node.Type == "" {
		errs = append(errs, fmt.Errorf("%s: type is required", path))
		return errs
	}

	if !model.IsValidNodeType(node.Type) {
		errs = append(errs, fmt.Errorf("%s: unknown node type %q", path, node.Type))
		return errs
	}

	if model.IsLeafType(node.Type) && len(node.Children) > 0 {
		errs = append(errs, fmt.Errorf("%s: leaf node %q cannot have children", path, node.Type))
	}

	if model.IsCompositeType(node.Type) && len(node.Children) == 0 {
		errs = append(errs, fmt.Errorf("%s: composite %q must have at least one child", path, node.Type))
	}

	if node.Type == "decorator" {
		if len(node.Children) != 1 {
			errs = append(errs, fmt.Errorf("%s: decorator must have exactly one child, got %d", path, len(node.Children)))
		}
	}

	for i := range node.Children {
		childPath := fmt.Sprintf("%s.children[%d]", path, i)
		errs = append(errs, validateNode(&node.Children[i], childPath)...)
	}

	return errs
}
