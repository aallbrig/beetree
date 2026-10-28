package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	assert.Contains(t, text, `"valid": true`)

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
