package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSecureServer(t *testing.T) (http.Handler, *MemStore) {
	t.Helper()
	store := NewMemStore()
	n := 0
	s := &Server{
		Store: store, Auth: fakeAuth{}, Version: "test",
		Now:   func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
		NewID: func() string { n++; return "id" + string(rune('a'+n)) },
	}
	return s.Handler(), store
}

func withKey(h http.Handler, method, path, key string) resp {
	req := httptest.NewRequest(method, path, strings.NewReader(`{"spec":""}`))
	req.Header.Set("X-API-Key", key)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return resp{Code: rec.Code, Body: rec.Body.Bytes()}
}

func TestAnonymousCallsAreRejectedExceptHealthAndSpec(t *testing.T) {
	h, _ := newSecureServer(t)
	for _, p := range []string{"/v1/trees", "/v1/node-types", "/v1/forum/threads", "/v1/users/x"} {
		r := do(h, "GET", p, "", nil)
		assert.Equal(t, 401, r.Code, p)
		assert.Contains(t, string(r.Body), docsURL)
	}
	assert.Equal(t, 401, do(h, "POST", "/mcp", "", map[string]any{"jsonrpc": "2.0", "id": 1, "method": "ping"}).Code, "MCP needs credentials too")
	assert.Equal(t, 200, do(h, "GET", "/v1/health", "", nil).Code)
	assert.Equal(t, 200, do(h, "GET", "/v1/openapi.json", "", nil).Code)
}

func TestAPIKeyLifecycle(t *testing.T) {
	h, store := newSecureServer(t)

	created := do(h, "POST", "/v1/keys", "alice", map[string]string{"name": "laptop"})
	require.Equal(t, 201, created.Code, string(created.Body))
	body := created.JSON(t)
	secret, _ := body["key"].(string)
	require.True(t, strings.HasPrefix(secret, "btc_"))
	assert.Equal(t, secret[:8], body["prefix"])

	// only the hash is stored
	keys, err := store.ListAPIKeys(t.Context(), "alice")
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, hashAPIKey(secret), keys[0].Hash)
	assert.NotContains(t, keys[0].Hash, secret)

	// listing never reveals secrets
	list := do(h, "GET", "/v1/keys", "alice", nil)
	require.Equal(t, 200, list.Code)
	assert.NotContains(t, string(list.Body), secret)
	assert.NotContains(t, string(list.Body), keys[0].Hash)

	// the key works as X-API-Key and as a bearer token
	assert.Equal(t, 200, withKey(h, "GET", "/v1/node-types", secret).Code)
	req := httptest.NewRequest("GET", "/v1/trees/mine", nil)
	req.Header.Set("Authorization", "Bearer "+secret)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.NotEqual(t, 401, rec.Code)

	// a key cannot manage keys
	assert.Equal(t, 403, withKey(h, "POST", "/v1/keys", secret).Code)
	assert.Equal(t, 403, withKey(h, "GET", "/v1/keys", secret).Code)

	// a made-up key is rejected, a revoked key stops working
	assert.Equal(t, 401, withKey(h, "GET", "/v1/node-types", "btc_notarealkey").Code)
	assert.Equal(t, 204, do(h, "DELETE", "/v1/keys/"+keys[0].ID, "alice", nil).Code)
	assert.Equal(t, 401, withKey(h, "GET", "/v1/node-types", secret).Code)
}

func TestAPIKeyRules(t *testing.T) {
	h, _ := newSecureServer(t)
	assert.Equal(t, 401, do(h, "POST", "/v1/keys", "", map[string]string{"name": "x"}).Code)
	assert.Equal(t, 400, do(h, "POST", "/v1/keys", "alice", map[string]string{"name": "  "}).Code)
	for i := 0; i < maxKeysPerUser; i++ {
		require.Equal(t, 201, do(h, "POST", "/v1/keys", "alice", map[string]string{"name": "k"}).Code)
	}
	assert.Equal(t, 409, do(h, "POST", "/v1/keys", "alice", map[string]string{"name": "k"}).Code)
	assert.Equal(t, 201, do(h, "POST", "/v1/keys", "bob", map[string]string{"name": "k"}).Code, "limit is per user")
	assert.Equal(t, 404, do(h, "DELETE", "/v1/keys/idz", "bob", nil).Code)
}

func TestKeyGenerationIsUniqueAndHashed(t *testing.T) {
	a, err := newAPIKeySecret()
	require.NoError(t, err)
	b, _ := newAPIKeySecret()
	assert.NotEqual(t, a, b)
	assert.Len(t, a, len("btc_")+43)
	assert.Len(t, hashAPIKey(a), 64)
}
