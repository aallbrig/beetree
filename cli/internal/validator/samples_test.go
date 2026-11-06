package validator_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/aallbrig/beetree-cli/internal/spec"
	"github.com/aallbrig/beetree-cli/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Each builder template is a `yaml: ` template literal in web/src/samples.ts.
var sampleYAML = regexp.MustCompile("(?s)id: '(sample:[a-z]+)',.*?yaml: `(.*?)`,")

// A template the real validator rejects would greet a first-time user with errors.
func TestBuilderTemplatesPassRealValidator(t *testing.T) {
	src, err := os.ReadFile("../../../web/src/samples.ts")
	require.NoError(t, err)

	matches := sampleYAML.FindAllStringSubmatch(string(src), -1)
	require.GreaterOrEqual(t, len(matches), 4, "expected the built-in templates in samples.ts")

	for _, m := range matches {
		id, yaml := m[1], m[2]
		t.Run(id, func(t *testing.T) {
			treeSpec, err := spec.ParseYAML([]byte(yaml))
			require.NoError(t, err)
			assert.Empty(t, validator.Validate(treeSpec))
		})
	}
}
