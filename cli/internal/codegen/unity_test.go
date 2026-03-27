package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnityGenerator_Engine(t *testing.T) {
	gen := NewUnityGenerator()
	assert.Equal(t, "unity", gen.Engine())
}

func TestUnityGenerator_Generate(t *testing.T) {
	spec := sampleSpec()
	gen := NewUnityGenerator()

	files, err := gen.Generate(spec)
	require.NoError(t, err)
	require.NotEmpty(t, files)

	// Should generate: blackboard, tree definition, + action stubs + condition stubs
	fileNames := make([]string, len(files))
	for i, f := range files {
		fileNames[i] = f.Path
	}

	// Blackboard
	assert.Contains(t, fileNames, "EnemyAiBlackboard.cs")
	// Tree definition (auto-regenerated)
	assert.Contains(t, fileNames, "EnemyAiTreeDefinition.cs")
	// Action stubs
	assert.Contains(t, fileNames, "AttackAction.cs")
	assert.Contains(t, fileNames, "AlertAction.cs")
	assert.Contains(t, fileNames, "PatrolAction.cs")
	// Condition stubs
	assert.Contains(t, fileNames, "HasTargetCondition.cs")
	assert.Contains(t, fileNames, "DetectNearbyEnemyCondition.cs")
}

func TestUnityGenerator_BlackboardContent(t *testing.T) {
	spec := sampleSpec()
	gen := NewUnityGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	var bbFile *GeneratedFile
	for i := range files {
		if files[i].Path == "EnemyAiBlackboard.cs" {
			bbFile = &files[i]
			break
		}
	}
	require.NotNil(t, bbFile)
	assert.False(t, bbFile.IsStub)
	assert.Contains(t, bbFile.Content, "AUTO-GENERATED")
	assert.Contains(t, bbFile.Content, "target")
	assert.Contains(t, bbFile.Content, "health")
	assert.Contains(t, bbFile.Content, "is_alerted")
}

func TestUnityGenerator_TreeDefinitionContent(t *testing.T) {
	spec := sampleSpec()
	gen := NewUnityGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	var tdFile *GeneratedFile
	for i := range files {
		if files[i].Path == "EnemyAiTreeDefinition.cs" {
			tdFile = &files[i]
			break
		}
	}
	require.NotNil(t, tdFile)
	assert.False(t, tdFile.IsStub)
	assert.Contains(t, tdFile.Content, "AUTO-GENERATED")
	assert.Contains(t, tdFile.Content, "EnemyAi")
	assert.Contains(t, tdFile.Content, "BTSelector")
}

func TestUnityGenerator_ActionStubContent(t *testing.T) {
	spec := sampleSpec()
	gen := NewUnityGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	var attackFile *GeneratedFile
	for i := range files {
		if files[i].Path == "AttackAction.cs" {
			attackFile = &files[i]
			break
		}
	}
	require.NotNil(t, attackFile)
	assert.True(t, attackFile.IsStub)
	assert.Contains(t, attackFile.Content, "EDIT THIS FILE")
	assert.Contains(t, attackFile.Content, "class AttackAction")
	assert.Contains(t, attackFile.Content, "BTAction")
	assert.Contains(t, attackFile.Content, "OnTick")
}

func TestUnityGenerator_ConditionStubContent(t *testing.T) {
	spec := sampleSpec()
	gen := NewUnityGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	var condFile *GeneratedFile
	for i := range files {
		if files[i].Path == "HasTargetCondition.cs" {
			condFile = &files[i]
			break
		}
	}
	require.NotNil(t, condFile)
	assert.True(t, condFile.IsStub)
	assert.Contains(t, condFile.Content, "EDIT THIS FILE")
	assert.Contains(t, condFile.Content, "class HasTargetCondition")
	assert.Contains(t, condFile.Content, "BTCondition")
	assert.Contains(t, condFile.Content, "Evaluate")
}

func TestUnityGenerator_NoDuplicateStubs(t *testing.T) {
	spec := sampleSpec()
	// Add a second reference to Attack
	spec.Tree.Children[0].Children = append(spec.Tree.Children[0].Children,
		spec.Tree.Children[0].Children[1]) // duplicate attack

	gen := NewUnityGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	attackCount := 0
	for _, f := range files {
		if f.Path == "AttackAction.cs" {
			attackCount++
		}
	}
	assert.Equal(t, 1, attackCount, "should not generate duplicate stubs")
}

func TestUnityGenerator_ValidCSharp(t *testing.T) {
	spec := sampleSpec()
	gen := NewUnityGenerator()
	files, err := gen.Generate(spec)
	require.NoError(t, err)

	for _, f := range files {
		if strings.HasSuffix(f.Path, ".cs") {
			assert.Contains(t, f.Content, "using", "file %s should have using statements", f.Path)
			assert.Contains(t, f.Content, "namespace", "file %s should have namespace", f.Path)
			// Ensure balanced braces
			opens := strings.Count(f.Content, "{")
			closes := strings.Count(f.Content, "}")
			assert.Equal(t, opens, closes, "file %s should have balanced braces", f.Path)
		}
	}
}
