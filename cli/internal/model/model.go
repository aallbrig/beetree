package model

import "fmt"

// Status represents the return status of a behavior tree node tick.
type Status int

const (
	StatusSuccess Status = iota
	StatusFailure
	StatusRunning
)

func (s Status) String() string {
	switch s {
	case StatusSuccess:
		return "Success"
	case StatusFailure:
		return "Failure"
	case StatusRunning:
		return "Running"
	default:
		return fmt.Sprintf("Unknown(%d)", int(s))
	}
}

// TreeSpec is the top-level structure for a .beetree.yaml specification.
type TreeSpec struct {
	Version     string          `yaml:"version" json:"version"`
	Metadata    Metadata        `yaml:"metadata" json:"metadata"`
	Blackboard  []BlackboardVar `yaml:"blackboard,omitempty" json:"blackboard,omitempty"`
	CustomNodes []CustomNodeDef `yaml:"custom_nodes,omitempty" json:"custom_nodes,omitempty"`
	Notation    NotationConfig  `yaml:"notation,omitempty" json:"notation,omitempty"`
	Tree        NodeSpec        `yaml:"tree" json:"tree"`
	Subtrees    []SubtreeRef    `yaml:"subtrees,omitempty" json:"subtrees,omitempty"`
}

// NotationConfig allows users to override sigils for node types and specific node classes.
type NotationConfig struct {
	TypeSigils map[string]string `yaml:"type_sigils,omitempty" json:"type_sigils,omitempty"`
	NodeSigils map[string]string `yaml:"node_sigils,omitempty" json:"node_sigils,omitempty"`
}

// Metadata holds descriptive information about a behavior tree spec.
type Metadata struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Author      string   `yaml:"author,omitempty" json:"author,omitempty"`
	License     string   `yaml:"license,omitempty" json:"license,omitempty"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// BlackboardVar defines a typed variable on the blackboard.
type BlackboardVar struct {
	Name        string      `yaml:"name" json:"name"`
	Type        string      `yaml:"type" json:"type"`
	Default     interface{} `yaml:"default,omitempty" json:"default,omitempty"`
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
}

// NodeSpec defines a single node within the behavior tree.
type NodeSpec struct {
	Type          string                 `yaml:"type" json:"type"`
	Name          string                 `yaml:"name" json:"name"`
	Node          string                 `yaml:"node,omitempty" json:"node,omitempty"`
	Children      []NodeSpec             `yaml:"children,omitempty" json:"children,omitempty"`
	Parameters    map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Decorator     string                 `yaml:"decorator,omitempty" json:"decorator,omitempty"`
	SuccessPolicy string                 `yaml:"success_policy,omitempty" json:"success_policy,omitempty"`
	FailurePolicy string                 `yaml:"failure_policy,omitempty" json:"failure_policy,omitempty"`
	Ref           string                 `yaml:"ref,omitempty" json:"ref,omitempty"`
	File          string                 `yaml:"file,omitempty" json:"file,omitempty"`
}

// CustomNodeDef defines a user-created node type.
type CustomNodeDef struct {
	Name             string         `yaml:"name" json:"name"`
	Type             string         `yaml:"type" json:"type"`
	Sigil            string         `yaml:"sigil,omitempty" json:"sigil,omitempty"`
	Description      string         `yaml:"description,omitempty" json:"description,omitempty"`
	Category         string         `yaml:"category,omitempty" json:"category,omitempty"`
	Parameters       []ParameterDef `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	BlackboardReads  []string       `yaml:"blackboard_reads,omitempty" json:"blackboard_reads,omitempty"`
	BlackboardWrites []string       `yaml:"blackboard_writes,omitempty" json:"blackboard_writes,omitempty"`
}

// ParameterDef defines a typed parameter for a custom node.
type ParameterDef struct {
	Name    string      `yaml:"name" json:"name"`
	Type    string      `yaml:"type" json:"type"`
	Default interface{} `yaml:"default,omitempty" json:"default,omitempty"`
}

// SubtreeRef references an external subtree file.
type SubtreeRef struct {
	Name string `yaml:"name" json:"name"`
	File string `yaml:"file" json:"file"`
}
