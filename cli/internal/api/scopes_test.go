package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func keyDo(h http.Handler, method, path, secret string, body any) resp {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("X-API-Key", secret)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return resp{rec.Code, rec.Body.Bytes()}
}

func makeKey(t *testing.T, h http.Handler, body map[string]any) string {
	t.Helper()
	body["name"] = "k"
	r := do(h, "POST", "/v1/keys", "alice", body)
	require.Equal(t, 201, r.Code, string(r.Body))
	return r.JSON(t)["key"].(string)
}

func TestScopeImplicationChain(t *testing.T) {
	cases := []struct {
		granted  []string
		required string
		want     bool
	}{
		{[]string{"trees:delete"}, "trees:write", true},
		{[]string{"trees:delete"}, "trees:read", true},
		{[]string{"trees:write"}, "trees:read", true},
		{[]string{"trees:write"}, "trees:delete", false},
		{[]string{"trees:read"}, "trees:write", false},
		{[]string{"trees:delete"}, "scenes:read", false},
		{[]string{"profile:write"}, "profile:read", true},
		{[]string{"keys:write"}, "keys:read", true},
		{nil, "trees:read", false},
		{[]string{"trees:read"}, "bogus:read", false},
		{[]string{"trees:bogus"}, "trees:read", false},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, scopeAllows(c.granted, c.required), "%v -> %s", c.granted, c.required)
	}
	assert.Equal(t, []string{"scenes:read", "trees:write"}, normalizeScopes([]string{"trees:read", "trees:write", "scenes:read", "trees:read"}))
}

func TestPresetsContainOnlyRealScopesAndNoKeys(t *testing.T) {
	for name, scopes := range scopePresets {
		for _, sc := range scopes {
			assert.True(t, validScope(sc), "%s: %s", name, sc)
			assert.NotContains(t, sc, "keys:", name)
		}
	}
	assert.False(t, scopeAllows(scopePresets["read-write"], "trees:delete"))
	assert.False(t, scopeAllows(scopePresets["read-only"], "trees:write"))
	assert.True(t, scopeAllows(scopePresets["full"], "forum:delete"))
}

// A new route or MCP tool must say which scope it needs, even if the answer is none.
func TestEveryRouteAndToolHasAScope(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range (&Server{}).Routes() {
		k := r.Method + " " + r.Path
		seen[k] = true
		sc, ok := routeScopes[k]
		require.True(t, ok, "route %s has no entry in routeScopes (scopes.go)", k)
		assert.True(t, sc == "" || sc == scopeMCP || validScope(sc), "%s: bad scope %q", k, sc)
	}
	for k := range routeScopes {
		assert.True(t, seen[k], "routeScopes lists %s but no such route exists", k)
	}
	tools := map[string]bool{}
	for _, tl := range mcpTools() {
		tools[tl.Name] = true
		sc, ok := toolScopes[tl.Name]
		require.True(t, ok, "MCP tool %s has no entry in toolScopes (scopes.go)", tl.Name)
		assert.True(t, sc == "" || validScope(sc), "%s: bad scope %q", tl.Name, sc)
	}
	for name := range toolScopes {
		assert.True(t, tools[name], "toolScopes lists %s but no such tool exists", name)
	}
}

func TestReadOnlyKeyReadsButCannotWrite(t *testing.T) {
	h, _ := newSecureServer(t)
	secret := makeKey(t, h, map[string]any{"preset": "read-only"})

	assert.Equal(t, 200, keyDo(h, "GET", "/v1/trees", secret, nil).Code)
	assert.Equal(t, 200, keyDo(h, "GET", "/v1/node-types", secret, nil).Code, "stateless tools need no scope")

	r := keyDo(h, "POST", "/v1/trees", secret, map[string]any{"name": "x"})
	assert.Equal(t, 403, r.Code)
	e := r.JSON(t)["error"].(map[string]any)
	assert.Equal(t, "insufficient_scope", e["code"])
	assert.Equal(t, "trees:write", e["required"])
	assert.Contains(t, e["granted"], "trees:read")

	r = keyDo(h, "DELETE", "/v1/trees/abc", secret, nil)
	assert.Equal(t, 403, r.Code)
	assert.Equal(t, "trees:delete", r.JSON(t)["error"].(map[string]any)["required"])

	assert.Equal(t, 403, keyDo(h, "POST", "/v1/forum/threads", secret, map[string]any{}).Code)
	assert.Equal(t, 401, keyDo(h, "GET", "/v1/trees", "btc_bogus", nil).Code, "bad keys stay 401, not 403")
}

func TestReadWriteKeyCannotDelete(t *testing.T) {
	h, _ := newSecureServer(t)
	secret := makeKey(t, h, map[string]any{"preset": "read-write"})
	assert.Equal(t, 200, keyDo(h, "GET", "/v1/trees", secret, nil).Code)
	assert.Equal(t, 403, keyDo(h, "DELETE", "/v1/trees/abc", secret, nil).Code)
	assert.Equal(t, 403, keyDo(h, "DELETE", "/v1/scenes/abc", secret, nil).Code)
	assert.NotEqual(t, 403, keyDo(h, "PUT", "/v1/trees/abc/like", secret, nil).Code, "write scope reaches the handler")
}

func TestFullKeyReachesDeleteButNeverKeys(t *testing.T) {
	h, _ := newSecureServer(t)
	secret := makeKey(t, h, map[string]any{"preset": "full"})
	assert.Equal(t, 404, keyDo(h, "DELETE", "/v1/trees/abc", secret, nil).Code, "past the scope check, the tree is just missing")
	r := keyDo(h, "POST", "/v1/keys", secret, map[string]any{"name": "x", "preset": "full"})
	assert.Equal(t, 403, r.Code, "even a key granted keys:write cannot mint keys")
	assert.Equal(t, 403, keyDo(h, "GET", "/v1/keys", secret, nil).Code)
}

func TestExplicitScopesAndValidation(t *testing.T) {
	h, store := newSecureServer(t)
	secret := makeKey(t, h, map[string]any{"scopes": []string{"trees:read", "trees:write", "scenes:read"}})
	keys, _ := store.ListAPIKeys(t.Context(), "alice")
	assert.Equal(t, []string{"scenes:read", "trees:write"}, keys[0].Scopes, "implied scopes are folded")
	assert.Equal(t, 403, keyDo(h, "GET", "/v1/forum/threads", secret, nil).Code)
	assert.Equal(t, 200, keyDo(h, "GET", "/v1/scenes", secret, nil).Code)

	bad := []map[string]any{
		{"name": "k"},
		{"name": "k", "preset": "admin"},
		{"name": "k", "scopes": []string{"trees:fly"}},
		{"name": "k", "preset": "full", "scopes": []string{"trees:read"}},
		{"name": "k", "preset": "full", "expiry": "forever"},
	}
	for _, b := range bad {
		assert.Equal(t, 400, do(h, "POST", "/v1/keys", "alice", b).Code, "%v", b)
	}
}

func TestKeyExpiry(t *testing.T) {
	h, store := newSecureServer(t)
	def := do(h, "POST", "/v1/keys", "alice", map[string]any{"name": "d", "preset": "read-only"}).JSON(t)
	assert.Equal(t, "2026-04-01T00:00:00Z", def["expiresAt"], "default is 90 days")
	never := do(h, "POST", "/v1/keys", "alice", map[string]any{"name": "n", "preset": "read-only", "expiry": "never"}).JSON(t)
	assert.Nil(t, never["expiresAt"])
	y := do(h, "POST", "/v1/keys", "alice", map[string]any{"name": "y", "preset": "read-only", "expiry": "365d"}).JSON(t)
	assert.Equal(t, "2027-01-01T00:00:00Z", y["expiresAt"])

	secret := "btc_expiredexpiredexpiredexpiredexpiredexp"
	require.NoError(t, store.PutAPIKey(t.Context(), APIKey{
		ID: "old", OwnerID: "alice", Hash: hashAPIKey(secret), Scopes: []string{"trees:read"},
		CreatedAt: "2025-01-01T00:00:00Z", ExpiresAt: "2025-12-31T00:00:00Z",
	}))
	r := keyDo(h, "GET", "/v1/trees", secret, nil)
	assert.Equal(t, 401, r.Code)
	assert.Contains(t, string(r.Body), "expired")
}

func TestLegacyKeysAreFullMinusKeys(t *testing.T) {
	h, store := newSecureServer(t)
	secret := "btc_legacylegacylegacylegacylegacylegacylega"
	require.NoError(t, store.PutAPIKey(t.Context(), APIKey{ID: "leg", OwnerID: "alice", Name: "old", Hash: hashAPIKey(secret), CreatedAt: "2025-01-01T00:00:00Z"}))
	assert.Equal(t, 200, keyDo(h, "GET", "/v1/trees", secret, nil).Code)
	assert.Equal(t, 404, keyDo(h, "DELETE", "/v1/trees/abc", secret, nil).Code)
	assert.Equal(t, 403, keyDo(h, "GET", "/v1/keys", secret, nil).Code)

	list := do(h, "GET", "/v1/keys", "alice", nil).JSON(t)["items"].([]any)
	info := list[0].(map[string]any)
	assert.Equal(t, true, info["legacy"])
	assert.Contains(t, info["scopes"], "trees:delete")
}

func TestLastUsedIsRecordedAndThrottled(t *testing.T) {
	store := NewMemStore()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := &Server{Store: store, Auth: fakeAuth{}, Now: func() time.Time { return now }}
	h := s.Handler()
	secret := makeKey(t, h, map[string]any{"preset": "read-only"})
	get := func() APIKey {
		k, err := store.GetAPIKeyByHash(t.Context(), hashAPIKey(secret))
		require.NoError(t, err)
		return k
	}
	assert.Empty(t, get().LastUsedAt)

	keyDo(h, "GET", "/v1/trees", secret, nil)
	first := get().LastUsedAt
	assert.Equal(t, "2026-01-01T00:00:00Z", first)

	now = now.Add(10 * time.Minute)
	keyDo(h, "GET", "/v1/trees", secret, nil)
	assert.Equal(t, first, get().LastUsedAt, "within the hour: no write")

	now = now.Add(2 * time.Hour)
	keyDo(h, "GET", "/v1/trees", secret, nil)
	assert.NotEqual(t, first, get().LastUsedAt)
}

func TestMCPToolsRespectScopes(t *testing.T) {
	h, _ := newSecureServer(t)
	ro := makeKey(t, h, map[string]any{"preset": "read-only"})
	rw := makeKey(t, h, map[string]any{"preset": "read-write"})
	call := func(secret, name string, args map[string]any) (string, bool) {
		r := keyDo(h, "POST", "/mcp", secret, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": map[string]any{"name": name, "arguments": args}})
		require.Equal(t, 200, r.Code, string(r.Body))
		var out map[string]any
		require.NoError(t, json.Unmarshal(r.Body, &out))
		return toolResult(t, out)
	}
	text, isErr := call(ro, "create_tree", map[string]any{"spec": goodSpec})
	assert.True(t, isErr)
	assert.Contains(t, text, "insufficient_scope")
	assert.Contains(t, text, "trees:write")

	_, isErr = call(ro, "list_my_trees", nil)
	assert.False(t, isErr, "read scope lists trees")
	_, isErr = call(ro, "validate_tree", map[string]any{"spec": goodSpec})
	assert.False(t, isErr, "stateless tools need no scope")

	_, isErr = call(rw, "create_tree", map[string]any{"spec": goodSpec})
	assert.False(t, isErr)
	text, isErr = call(rw, "delete_tree", map[string]any{"id": "x"})
	assert.True(t, isErr)
	assert.Contains(t, text, "trees:delete")
}

// Every MCP tool that acts on an account must demand a scope, so a limited key
// cannot reach it by a missing table entry, and must refuse anonymous callers.
func TestMCPAccountToolsDemandScopeAndSignIn(t *testing.T) {
	h := newTestServer(t)
	for _, tl := range mcpTools() {
		if !tl.needsAuth {
			continue
		}
		assert.NotEmpty(t, toolScopes[tl.Name], "account tool %s must require a scope", tl.Name)
		_, out := rpc(h, "", 1, "tools/call", map[string]any{"name": tl.Name, "arguments": map[string]any{"id": "x", "spec": goodSpec}})
		text, isErr := toolResult(t, out)
		assert.True(t, isErr, "anonymous call to %s must be refused", tl.Name)
		assert.Contains(t, text, "signed-in user", tl.Name)
	}
}
