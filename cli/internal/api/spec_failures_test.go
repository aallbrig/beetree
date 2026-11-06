package api

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Every failure a stranger can trigger on any route (no credentials, a missing
// resource, a body that is not JSON) must come back with a status the OpenAPI
// document declares for that operation.
func TestOpenAPIDeclaresEveryProbedFailureStatus(t *testing.T) {
	var doc struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(openapiYAML, &doc))
	h := newTestServer(t)

	declared := func(method, path string) map[string]bool {
		op, _ := doc.Paths[path][strings.ToLower(method)].(map[string]any)
		out := map[string]bool{"2xx": true}
		resp, _ := op["responses"].(map[string]any)
		for k := range resp {
			out[k] = true
		}
		return out
	}

	for _, rt := range (&Server{}).Routes() {
		if rt.Path == "/mcp" {
			continue
		}
		url := strings.NewReplacer("{id}", "does-not-exist", "{commentId}", "nope").Replace(rt.Path)
		var body any
		if rt.Method == "POST" || rt.Method == "PUT" {
			body = map[string]any{}
		}
		for _, user := range []string{"", "alice"} {
			r := do(h, rt.Method, url, user, body)
			if r.Code >= 200 && r.Code < 300 {
				continue
			}
			ok := declared(rt.Method, rt.Path)[strconv.Itoa(r.Code)]
			assert.True(t, ok, "%s %s as %q returned %d, which openapi.yaml does not declare: %s",
				rt.Method, rt.Path, user, r.Code, r.Body)
		}
	}
}
