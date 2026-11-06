package validator_test

import (
	"path/filepath"
	"testing"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/aallbrig/beetree-cli/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func countNodes(n *model.NodeSpec) int {
	total := 1
	for i := range n.Children {
		total += countNodes(&n.Children[i])
	}
	return total
}

// The Explore combat scene builds on these four trees; the builder shows each whole.
func TestCombatSceneExamplesValidateAndStayUnder25Nodes(t *testing.T) {
	for _, name := range []string{"trooper", "skirmisher", "shieldbearer", "commander"} {
		t.Run(name, func(t *testing.T) {
			treeSpec, err := spec.ParseFile(filepath.Join("..", "..", "..", "examples", name+".beetree.yaml"))
			require.NoError(t, err)
			assert.Empty(t, validator.Validate(treeSpec))
			assert.LessOrEqual(t, countNodes(&treeSpec.Tree), 25)
		})
	}
}
