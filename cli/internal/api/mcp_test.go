package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rpc(h http.Handler, user string, id any, method string, params any) (int, map[string]any) {
	body := map[string]any{"jsonrpc": "2.0", "method": method}
	if id != nil {
		body["id"] = id
	}
	if params != nil {
		body["params"] = params
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(string(b)))
	if user != "" {
		req.Header.Set("Authorization", "Bearer token-"+user)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func toolResult(t *testing.T, resp map[string]any) (string, bool) {
	t.Helper()
	res := resp["result"].(map[string]any)
	content := res["content"].([]any)[0].(map[string]any)
	return content["text"].(string), res["isError"].(bool)
}

func TestMCPHandshake(t *testing.T) {
	h := newTestServer(t)
	code, r := rpc(h, "", 1, "initialize", map[string]any{"protocolVersion": mcpProtocolVersion})
	require.Equal(t, 200, code)
	res := r["result"].(map[string]any)
	assert.Equal(t, mcpProtocolVersion, res["protocolVersion"])
	assert.Equal(t, "beetreecraft", res["serverInfo"].(map[string]any)["name"])

	ins := res["instructions"].(string)
	for _, want := range []string{"beetree://schema/beetree", "beetree://examples/patrol", "validate_tree", "simulate_answer"} {
		assert.Contains(t, ins, want)
	}

	code, _ = rpc(h, "", nil, "notifications/initialized", nil)
	assert.Equal(t, 202, code, "notifications get no body")

	_, r = rpc(h, "", 2, "ping", nil)
	assert.NotNil(t, r["result"])

	_, r = rpc(h, "", 3, "bogus/method", nil)
	assert.EqualValues(t, rpcMethodNotFound, r["error"].(map[string]any)["code"])

	req := httptest.NewRequest("POST", "/mcp", strings.NewReader("{broken"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Contains(t, w.Body.String(), "-32700")

	req = httptest.NewRequest("GET", "/mcp", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, 405, w.Code)
}

func TestMCPToolsList(t *testing.T) {
	_, r := rpc(newTestServer(t), "", 1, "tools/list", nil)
	tools := r["result"].(map[string]any)["tools"].([]any)
	names := map[string]bool{}
	for _, tl := range tools {
		m := tl.(map[string]any)
		names[m["name"].(string)] = true
		assert.NotEmpty(t, m["description"])
		ann, ok := m["annotations"].(map[string]any)
		require.True(t, ok, "%s needs annotations", m["name"])
		assert.Equal(t, m["name"] != "create_tree", ann["readOnlyHint"], m["name"])
		assert.Equal(t, false, ann["destructiveHint"], m["name"])
		assert.Equal(t, "object", m["inputSchema"].(map[string]any)["type"])
	}
	for _, want := range []string{"validate_tree", "render_tree", "simulate_tree", "generate_code", "search_trees", "get_tree", "create_tree", "list_my_trees", "list_node_types"} {
		assert.True(t, names[want], want)
	}
}

func TestMCPStatelessTools(t *testing.T) {
	h := newTestServer(t)
	call := func(name string, args map[string]any) (string, bool) {
		_, r := rpc(h, "", 1, "tools/call", map[string]any{"name": name, "arguments": args})
		return toolResult(t, r)
	}

	text, isErr := call("validate_tree", map[string]any{"spec": goodSpec})
	assert.False(t, isErr)
	assert.Contains(t, text, `"valid":true`)

	text, isErr = call("render_tree", map[string]any{"spec": goodSpec, "format": "mermaid"})
	assert.False(t, isErr)
	assert.Contains(t, text, "graph")

	text, isErr = call("simulate_tree", map[string]any{"spec": goodSpec})
	assert.False(t, isErr)
	assert.Contains(t, text, "Simulation")

	text, isErr = call("generate_code", map[string]any{"spec": goodSpec, "engine": "godot"})
	assert.False(t, isErr)
	assert.Contains(t, text, "===")

	_, isErr = call("generate_code", map[string]any{"spec": goodSpec, "engine": "nope"})
	assert.True(t, isErr, "tool failures are reported in-band with isError")

	text, _ = call("list_node_types", nil)
	assert.Contains(t, text, "selector")
}

func TestMCPAccountTools(t *testing.T) {
	h := newTestServer(t)
	call := func(user, name string, args map[string]any) (string, bool) {
		_, r := rpc(h, user, 1, "tools/call", map[string]any{"name": name, "arguments": args})
		return toolResult(t, r)
	}

	text, isErr := call("", "create_tree", map[string]any{"spec": goodSpec})
	assert.True(t, isErr)
	assert.Contains(t, text, "signed-in")

	text, isErr = call("alice", "create_tree", map[string]any{"spec": goodSpec, "visibility": "public", "name": "From MCP"})
	require.False(t, isErr, text)
	var tree Tree
	require.NoError(t, json.Unmarshal([]byte(text), &tree))
	assert.Equal(t, "From MCP", tree.Name)

	_, isErr = call("alice", "create_tree", map[string]any{"spec": "bad: ["})
	assert.True(t, isErr)

	text, _ = call("alice", "list_my_trees", nil)
	assert.Contains(t, text, "From MCP")

	text, _ = call("", "search_trees", map[string]any{"query": "mcp"})
	assert.Contains(t, text, "From MCP")
	assert.NotContains(t, text, "has_waypoint", "search results omit spec bodies")

	text, isErr = call("", "get_tree", map[string]any{"id": tree.ID})
	assert.False(t, isErr)
	assert.Contains(t, text, "has_waypoint")

	priv, _ := call("alice", "create_tree", map[string]any{"spec": goodSpec})
	var pt Tree
	require.NoError(t, json.Unmarshal([]byte(priv), &pt))
	_, isErr = call("bob", "get_tree", map[string]any{"id": pt.ID})
	assert.True(t, isErr, "private trees are hidden from other users")
}

func TestMCPBadCalls(t *testing.T) {
	h := newTestServer(t)
	_, r := rpc(h, "", 1, "tools/call", map[string]any{"name": "nope"})
	assert.EqualValues(t, rpcInvalidParams, r["error"].(map[string]any)["code"])
	_, r = rpc(h, "", 1, "tools/call", map[string]any{})
	assert.EqualValues(t, rpcInvalidParams, r["error"].(map[string]any)["code"])
	_, r = rpc(h, "", 1, "tools/call", map[string]any{"name": "validate_tree", "arguments": "not-an-object"})
	assert.EqualValues(t, rpcInvalidParams, r["error"].(map[string]any)["code"])
}

func TestMCPResourcesAndPrompts(t *testing.T) {
	h := newTestServer(t)

	_, r := rpc(h, "", 1, "resources/list", nil)
	list := r["result"].(map[string]any)["resources"].([]any)
	require.Len(t, list, 3+len(builderTemplateNames(t)))

	for _, uri := range []string{"beetree://node-types", "beetree://schema/beetree", "beetree://examples/patrol"} {
		_, r = rpc(h, "", 2, "resources/read", map[string]any{"uri": uri})
		c := r["result"].(map[string]any)["contents"].([]any)[0].(map[string]any)
		assert.Equal(t, uri, c["uri"])
		assert.NotEmpty(t, c["text"])
	}

	// the starter spec must pass the real validator
	res, _ := validateYAML(starterSpec)
	assert.True(t, res.Valid, "%v", res.Errors)

	_, r = rpc(h, "", 3, "resources/read", map[string]any{"uri": "beetree://nope"})
	assert.NotNil(t, r["error"])

	_, r = rpc(h, "", 4, "prompts/list", nil)
	assert.Len(t, r["result"].(map[string]any)["prompts"].([]any), 1)

	_, r = rpc(h, "", 5, "prompts/get", map[string]any{"name": "design_behavior_tree", "arguments": map[string]any{"character": "a guard dog", "engine": "godot"}})
	msg := r["result"].(map[string]any)["messages"].([]any)[0].(map[string]any)["content"].(map[string]any)["text"].(string)
	assert.Contains(t, msg, "a guard dog")
	assert.Contains(t, msg, "generate_code with engine godot")

	_, r = rpc(h, "", 6, "prompts/get", map[string]any{"name": "design_behavior_tree"})
	assert.NotNil(t, r["error"])
}

func TestMCPSchemaMatchesPublished(t *testing.T) {
	published, err := os.ReadFile("../../../website/static/schemas/beetree.schema.json")
	require.NoError(t, err)
	assert.Equal(t, string(published), schemaJSON, "copy website/static/schemas/beetree.schema.json to cli/internal/api/beetree.schema.json")
	assert.True(t, json.Valid([]byte(schemaJSON)))
}

func TestMCPSearchFindsOlderTrees(t *testing.T) {
	store := NewMemStore()
	s := &Server{Store: store, Auth: fakeAuth{}, Version: "test", AllowAnonymous: true}
	h := s.Handler()
	for i := 0; i < 250; i++ {
		name := "filler"
		created := "2026-02-01T00:00:00Z"
		if i == 0 {
			name, created = "ancient-patrol", "2025-01-01T00:00:00Z"
		}
		require.NoError(t, store.PutTree(context.Background(), Tree{ID: fmt.Sprintf("t%03d", i), Name: name, Visibility: VisibilityPublic, CreatedAt: created}))
	}
	_, r := rpc(h, "", 1, "tools/call", map[string]any{"name": "search_trees", "arguments": map[string]any{"query": "ancient-patrol"}})
	text, isErr := toolResult(t, r)
	require.False(t, isErr, text)
	assert.Contains(t, text, "ancient-patrol", "search must reach beyond the newest 200 public trees")
}

// builderTemplateNames reads the template ids from the web builder's samples.
func builderTemplateNames(t *testing.T) []string {
	t.Helper()
	src, err := os.ReadFile("../../../web/src/samples.ts")
	require.NoError(t, err)
	var names []string
	for _, m := range regexp.MustCompile(`id: 'sample:(\w+)'`).FindAllStringSubmatch(string(src), -1) {
		names = append(names, m[1])
	}
	require.NotEmpty(t, names)
	return names
}

func TestMCPTemplatesMatchBuilder(t *testing.T) {
	h := newTestServer(t)
	src, err := os.ReadFile("../../../web/src/samples.ts")
	require.NoError(t, err)
	for _, name := range builderTemplateNames(t) {
		uri := "beetree://templates/" + name
		_, r := rpc(h, "", 1, "resources/read", map[string]any{"uri": uri})
		require.Nil(t, r["error"], uri)
		text := r["result"].(map[string]any)["contents"].([]any)[0].(map[string]any)["text"].(string)
		assert.Contains(t, string(src), text, "%s drifted from web/src/samples.ts", name)
		res, _ := validateYAML(text)
		assert.True(t, res.Valid, "%s: %v", name, res.Errors)
	}
	_, r := rpc(h, "", 2, "resources/read", map[string]any{"uri": "beetree://templates/../schema"})
	assert.NotNil(t, r["error"])
}

// TestMCPDocsMatchServer keeps website/content/docs/mcp.md honest: every tool,
// resource and prompt the server lists is documented, and every name the
// tables document is real.
func TestMCPDocsMatchServer(t *testing.T) {
	src, err := os.ReadFile("../../../website/content/docs/mcp.md")
	require.NoError(t, err)
	doc := string(src)

	h := newTestServer(t)
	names := func(method, key, field string) []string {
		_, r := rpc(h, "", 1, method, nil)
		var out []string
		for _, it := range r["result"].(map[string]any)[key].([]any) {
			out = append(out, it.(map[string]any)[field].(string))
		}
		require.NotEmpty(t, out, method)
		return out
	}
	tools := names("tools/list", "tools", "name")
	resources := names("resources/list", "resources", "uri")
	prompts := names("prompts/list", "prompts", "name")

	for _, n := range append(append([]string{}, tools...), prompts...) {
		assert.Contains(t, doc, "`"+n+"`", "mcp.md does not mention %q", n)
	}
	for _, u := range resources {
		// templates/<name> style entries are documented with a placeholder
		if i := strings.LastIndex(u, "/"); strings.HasPrefix(u, "beetree://templates/") && i > 0 {
			u = u[:i+1]
		}
		assert.Contains(t, doc, u, "mcp.md does not mention resource %q", u)
	}

	// Reverse: names in the first column of a table row must exist.
	known := map[string]bool{}
	for _, n := range append(append([]string{}, tools...), prompts...) {
		known[n] = true
	}
	cell := regexp.MustCompile("`([a-z]+(?:_[a-z]+)+)`")
	for _, line := range strings.Split(doc, "\n") {
		if !strings.HasPrefix(line, "| `") {
			continue
		}
		first := strings.SplitN(line, "|", 3)[1]
		for _, m := range cell.FindAllStringSubmatch(first, -1) {
			assert.True(t, known[m[1]], "mcp.md documents %q but the server has no such tool or prompt", m[1])
		}
	}
}

func TestMCPToolsCompactJSON(t *testing.T) {
	h := newTestServer(t)
	for _, name := range []string{"validate_tree", "list_node_types"} {
		args := map[string]any{"spec": goodSpec}
		_, r := rpc(h, "", 1, "tools/call", map[string]any{"name": name, "arguments": args})
		text, isErr := toolResult(t, r)
		require.False(t, isErr, name)
		assert.NotContains(t, text, "\n", "%s result must be compact JSON to save tokens", name)
		assert.True(t, json.Valid([]byte(text)), name)
	}
}
