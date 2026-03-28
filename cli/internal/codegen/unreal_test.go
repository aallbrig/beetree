package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnrealGenerator_Engine(t *testing.T) {
	gen := NewUnrealGenerator()
	assert.Equal(t, "unreal", gen.Engine())
}

func TestUnrealGenerator_Generate(t *testing.T) {
	spec := sampleSpec()
	gen := NewUnrealGenerator()

	files, err := gen.Generate(spec)
	require.NoError(t, err)
	require.NotEmpty(t, files)

	fileNames := make([]string, len(files))
	for i, f := range files {
		fileNames[i] = f.Path
	}

	// Integration README
	assert.Contains(t, fileNames, "README.md")
	// Blackboard header
	// Tree definition
	assert.Contains(t, fileNames, "EnemyAiTreeDefinition.h")
	assert.Contains(t, fileNames, "EnemyAiTreeDefinition.cpp")
	// Task stubs (header + source per action)
	assert.Contains(t, fileNames, "BTTask_Attack.h")
	assert.Contains(t, fileNames, "BTTask_Attack.cpp")
	assert.Contains(t, fileNames, "BTTask_Patrol.h")
	assert.Contains(t, fileNames, "BTTask_Patrol.cpp")
	// Condition decorators
	assert.Contains(t, fileNames, "BTDecorator_HasTarget.h")
	assert.Contains(t, fileNames, "BTDecorator_HasTarget.cpp")
}

func TestUnrealGenerator_BlackboardContent(t *testing.T) {
	spec := sampleSpec()
	gen := NewUnrealGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	var bbFile *GeneratedFile
	for i := range files {
		if files[i].Path == "EnemyAiBlackboard.h" {
			bbFile = &files[i]
			break
		}
	}
	require.NotNil(t, bbFile)
	assert.False(t, bbFile.IsStub)
	assert.Contains(t, bbFile.Content, "AUTO-GENERATED")
	assert.Contains(t, bbFile.Content, "Target")
	assert.Contains(t, bbFile.Content, "Health")
	assert.Contains(t, bbFile.Content, "UPROPERTY")
}

func TestUnrealGenerator_TaskStubContent(t *testing.T) {
	spec := sampleSpec()
	gen := NewUnrealGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	var taskH, taskCpp *GeneratedFile
	for i := range files {
		if files[i].Path == "BTTask_Attack.h" {
			taskH = &files[i]
		}
		if files[i].Path == "BTTask_Attack.cpp" {
			taskCpp = &files[i]
		}
	}
	require.NotNil(t, taskH)
	require.NotNil(t, taskCpp)
	assert.True(t, taskH.IsStub)
	assert.True(t, taskCpp.IsStub)
	assert.Contains(t, taskH.Content, "EDIT THIS FILE")
	assert.Contains(t, taskH.Content, "UBTTaskNode")
	assert.Contains(t, taskH.Content, "BTTask_Attack")
	assert.Contains(t, taskCpp.Content, "ExecuteTask")
}

func TestUnrealGenerator_NoDuplicateStubs(t *testing.T) {
	spec := sampleSpec()
	spec.Tree.Children[0].Children = append(spec.Tree.Children[0].Children,
		spec.Tree.Children[0].Children[1])

	gen := NewUnrealGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	attackCount := 0
	for _, f := range files {
		if f.Path == "BTTask_Attack.h" {
			attackCount++
		}
	}
	assert.Equal(t, 1, attackCount)
}

func TestUnrealGenerator_ValidCpp(t *testing.T) {
	spec := sampleSpec()
	gen := NewUnrealGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	for _, f := range files {
		if strings.HasSuffix(f.Path, ".h") {
			assert.Contains(t, f.Content, "#pragma once", "header %s should have pragma once", f.Path)
		}
		if strings.HasSuffix(f.Path, ".cpp") {
			assert.Contains(t, f.Content, "#include", "source %s should have includes", f.Path)
		}
	}
}
