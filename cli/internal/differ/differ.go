// Package differ provides structural diffing of behavior tree specs.
package differ

import (
	"fmt"
	"strings"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/aallbrig/beetree-cli/internal/renderer"
)

// ChangeType describes the kind of difference.
type ChangeType string

const (
	ChangeAdded    ChangeType = "added"
	ChangeRemoved  ChangeType = "removed"
	ChangeModified ChangeType = "modified"
)

// Change represents a single difference between two specs.
type Change struct {
	Type ChangeType
	Path string
	Old  string
	New  string
}

// Diff compares two tree specs and returns a list of structural changes.
func Diff(a, b *model.TreeSpec) []Change {
	var changes []Change
	changes = append(changes, diffMetadata(a, b)...)
	changes = append(changes, diffBlackboard(a.Blackboard, b.Blackboard)...)
	changes = append(changes, diffNode(&a.Tree, &b.Tree, "tree")...)
	return changes
}

func diffMetadata(a, b *model.TreeSpec) []Change {
	var changes []Change
	if a.Version != b.Version {
		changes = append(changes, Change{ChangeModified, "version", a.Version, b.Version})
	}
	if a.Metadata.Name != b.Metadata.Name {
		changes = append(changes, Change{ChangeModified, "metadata.name", a.Metadata.Name, b.Metadata.Name})
	}
	if a.Metadata.Description != b.Metadata.Description {
		changes = append(changes, Change{ChangeModified, "metadata.description", a.Metadata.Description, b.Metadata.Description})
	}
	if a.Metadata.Author != b.Metadata.Author {
		changes = append(changes, Change{ChangeModified, "metadata.author", a.Metadata.Author, b.Metadata.Author})
	}
	return changes
}

func diffBlackboard(a, b []model.BlackboardVar) []Change {
	var changes []Change
	aMap := make(map[string]model.BlackboardVar)
	for _, v := range a {
		aMap[v.Name] = v
	}
	bMap := make(map[string]model.BlackboardVar)
	for _, v := range b {
		bMap[v.Name] = v
	}

	for name, av := range aMap {
		bv, exists := bMap[name]
		if !exists {
			changes = append(changes, Change{ChangeRemoved, "blackboard." + name, fmt.Sprintf("%s (%v)", av.Type, av.Default), ""})
			continue
		}
		if av.Type != bv.Type {
			changes = append(changes, Change{ChangeModified, "blackboard." + name + ".type", av.Type, bv.Type})
		}
		if fmt.Sprintf("%v", av.Default) != fmt.Sprintf("%v", bv.Default) {
			changes = append(changes, Change{ChangeModified, "blackboard." + name + ".default", fmt.Sprintf("%v", av.Default), fmt.Sprintf("%v", bv.Default)})
		}
	}

	for name, bv := range bMap {
		if _, exists := aMap[name]; !exists {
			changes = append(changes, Change{ChangeAdded, "blackboard." + name, "", fmt.Sprintf("%s (%v)", bv.Type, bv.Default)})
		}
	}
	return changes
}

func diffNode(a, b *model.NodeSpec, path string) []Change {
	var changes []Change
	if a.Type != b.Type {
		changes = append(changes, Change{ChangeModified, path + ".type", fmt.Sprintf("%s %s", renderer.TypeSigil(a.Type), a.Type), fmt.Sprintf("%s %s", renderer.TypeSigil(b.Type), b.Type)})
	}
	if a.Name != b.Name {
		changes = append(changes, Change{ChangeModified, path + ".name", a.Name, b.Name})
	}
	if a.Node != b.Node {
		changes = append(changes, Change{ChangeModified, path + ".node", a.Node, b.Node})
	}
	if a.Decorator != b.Decorator {
		changes = append(changes, Change{ChangeModified, path + ".decorator", a.Decorator, b.Decorator})
	}

	// Diff children by matching on name
	aChildren := make(map[string]*model.NodeSpec)
	for i := range a.Children {
		aChildren[a.Children[i].Name] = &a.Children[i]
	}
	bChildren := make(map[string]*model.NodeSpec)
	for i := range b.Children {
		bChildren[b.Children[i].Name] = &b.Children[i]
	}

	for name, ac := range aChildren {
		bc, exists := bChildren[name]
		if !exists {
			changes = append(changes, Change{ChangeRemoved, path + ".children." + name, fmt.Sprintf("%s %s (%s)", renderer.TypeSigil(ac.Type), name, ac.Type), ""})
			continue
		}
		childPath := path + ".children." + name
		changes = append(changes, diffNode(ac, bc, childPath)...)
	}

	for name, bc := range bChildren {
		if _, exists := aChildren[name]; !exists {
			changes = append(changes, Change{ChangeAdded, path + ".children." + name, "", fmt.Sprintf("%s %s (%s)", renderer.TypeSigil(bc.Type), name, bc.Type)})
		}
	}

	return changes
}

// FormatDiff produces a human-readable diff output.
func FormatDiff(changes []Change) string {
	if len(changes) == 0 {
		return "No differences found.\n"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d change(s):\n", len(changes)))
	for _, c := range changes {
		switch c.Type {
		case ChangeAdded:
			b.WriteString(fmt.Sprintf("  + %s: %s\n", c.Path, c.New))
		case ChangeRemoved:
			b.WriteString(fmt.Sprintf("  - %s: %s\n", c.Path, c.Old))
		case ChangeModified:
			b.WriteString(fmt.Sprintf("  ~ %s: %s → %s\n", c.Path, c.Old, c.New))
		}
	}
	return b.String()
}
