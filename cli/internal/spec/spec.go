package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/aallbrig/beetree-cli/internal/model"
	"gopkg.in/yaml.v3"
)

// ParseYAML parses a YAML byte slice into a TreeSpec.
func ParseYAML(data []byte) (*model.TreeSpec, error) {
	if len(data) == 0 {
		return nil, errors.New("empty input")
	}

	var spec model.TreeSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	if spec.Version == "" && spec.Metadata.Name == "" && spec.Tree.Type == "" {
		return nil, errors.New("empty or invalid spec: missing version, metadata, and tree")
	}

	return &spec, nil
}

// ParseJSON parses a JSON byte slice into a TreeSpec.
func ParseJSON(data []byte) (*model.TreeSpec, error) {
	if len(data) == 0 {
		return nil, errors.New("empty input")
	}

	var spec model.TreeSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return &spec, nil
}

// ParseFile reads and parses a .beetree.yaml or .beetree.json file.
func ParseFile(path string) (*model.TreeSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}

	ext := fileExtension(path)
	switch ext {
	case ".json":
		return ParseJSON(data)
	default:
		return ParseYAML(data)
	}
}

// NodeCount returns the total number of nodes in a tree (including the root).
func NodeCount(node *model.NodeSpec) int {
	if node == nil {
		return 0
	}
	count := 1
	for i := range node.Children {
		count += NodeCount(&node.Children[i])
	}
	return count
}

func fileExtension(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}
