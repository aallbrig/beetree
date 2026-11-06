package api

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The "Save and share a tree" curl body in the API docs must really be accepted,
// and the error table must list every status the OpenAPI document declares.
func TestAPIDocsSaveExampleRuns(t *testing.T) {
	src, err := os.ReadFile("../../../website/content/docs/api/_index.md")
	require.NoError(t, err)
	doc := string(src)

	m := regexp.MustCompile(`(?s)/v1/trees \\\n.*?-d '(\{"name":"My guard".*?\})'\n`).FindStringSubmatch(doc)
	require.Len(t, m, 2, "save example not found in docs")
	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(m[1]), &body))

	h := newTestServer(t)
	r := do(h, "POST", "/v1/trees", "alice", body)
	assert.Equal(t, 201, r.Code, string(r.Body))

	for _, code := range []string{"400", "401", "403", "404", "413", "422", "429"} {
		assert.Contains(t, doc, "| "+code+" |", "error table lacks %s", code)
	}
}
