package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGodotGenerator_Engine(t *testing.T) {
	gen := NewGodotGenerator()
	assert.Equal(t, "godot", gen.Engine())
}

func TestGodotGenerator_Generate(t *testing.T) {
	spec := sampleSpec()
	gen := NewGodotGenerator()

	files, err := gen.Generate(spec)
	require.NoError(t, err)
	require.NotEmpty(t, files)

	fileNames := make([]string, len(files))
	for i, f := range files {
		fileNames[i] = f.Path
	}

	// Integration README
	assert.Contains(t, fileNames, "README.md")
	// Blackboard
	// Tree definition
	assert.Contains(t, fileNames, "enemy_ai_tree_definition.gd")
	// Action stubs
	assert.Contains(t, fileNames, "attack_action.gd")
	assert.Contains(t, fileNames, "alert_action.gd")
	assert.Contains(t, fileNames, "patrol_action.gd")
	// Condition stubs
	assert.Contains(t, fileNames, "has_target_condition.gd")
	assert.Contains(t, fileNames, "detect_nearby_enemy_condition.gd")
}

func TestGodotGenerator_BlackboardContent(t *testing.T) {
	spec := sampleSpec()
	gen := NewGodotGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	var bbFile *GeneratedFile
	for i := range files {
		if files[i].Path == "enemy_ai_blackboard.gd" {
			bbFile = &files[i]
			break
		}
	}
	require.NotNil(t, bbFile)
	assert.False(t, bbFile.IsStub)
	assert.Contains(t, bbFile.Content, "AUTO-GENERATED")
	assert.Contains(t, bbFile.Content, "target")
	assert.Contains(t, bbFile.Content, "health")
	assert.Contains(t, bbFile.Content, "extends Node")
}

func TestGodotGenerator_ActionStubContent(t *testing.T) {
	spec := sampleSpec()
	gen := NewGodotGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	var actionFile *GeneratedFile
	for i := range files {
		if files[i].Path == "attack_action.gd" {
			actionFile = &files[i]
			break
		}
	}
	require.NotNil(t, actionFile)
	assert.True(t, actionFile.IsStub)
	assert.Contains(t, actionFile.Content, "EDIT THIS FILE")
	assert.Contains(t, actionFile.Content, "extends Node")
	assert.Contains(t, actionFile.Content, "func tick")
}

func TestGodotGenerator_ConditionStubContent(t *testing.T) {
	spec := sampleSpec()
	gen := NewGodotGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	var condFile *GeneratedFile
	for i := range files {
		if files[i].Path == "has_target_condition.gd" {
			condFile = &files[i]
			break
		}
	}
	require.NotNil(t, condFile)
	assert.True(t, condFile.IsStub)
	assert.Contains(t, condFile.Content, "EDIT THIS FILE")
	assert.Contains(t, condFile.Content, "func evaluate")
}

func TestGodotGenerator_NoDuplicateStubs(t *testing.T) {
	spec := sampleSpec()
	spec.Tree.Children[0].Children = append(spec.Tree.Children[0].Children,
		spec.Tree.Children[0].Children[1])

	gen := NewGodotGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	attackCount := 0
	for _, f := range files {
		if f.Path == "attack_action.gd" {
			attackCount++
		}
	}
	assert.Equal(t, 1, attackCount)
}

func TestGodotGenerator_ValidGDScript(t *testing.T) {
	spec := sampleSpec()
	gen := NewGodotGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	for _, f := range files {
		if strings.HasSuffix(f.Path, ".gd") {
			assert.Contains(t, f.Content, "extends", "file %s should extend a class", f.Path)
		}
	}
}
