package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// authzFixture is a fresh server holding one private tree, scene, thread and
// comment owned by alice, plus two of her keys (read-only and full).
type authzFixture struct {
	h                                                   http.Handler
	tree, scene, thread, comment, keyID, roKey, fullKey string
}

func newAuthzFixture(t *testing.T) authzFixture {
	t.Helper()
	h := newTestServer(t)
	f := authzFixture{h: h}
	r := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "visibility": "private"})
	require.Equal(t, 201, r.Code, string(r.Body))
	f.tree = r.JSON(t)["id"].(string)
	r = do(h, "POST", "/v1/trees/"+f.tree+"/comments", "alice", map[string]any{"body": "hello"})
	require.Less(t, r.Code, 300, string(r.Body))
	f.comment = r.JSON(t)["id"].(string)
	r = do(h, "POST", "/v1/scenes", "alice", map[string]any{"name": "s", "agents": []any{}})
	require.Less(t, r.Code, 300, string(r.Body))
	f.scene = r.JSON(t)["id"].(string)
	r = do(h, "POST", "/v1/forum/threads", "alice", map[string]any{"title": "t", "body": "b"})
	require.Less(t, r.Code, 300, string(r.Body))
	f.thread = r.JSON(t)["id"].(string)
	r = do(h, "POST", "/v1/keys", "alice", map[string]any{"name": "ro", "preset": "read-only"})
	require.Equal(t, 201, r.Code, string(r.Body))
	f.roKey = r.JSON(t)["key"].(string)
	f.keyID = r.JSON(t)["id"].(string)
	f.fullKey = makeKey(t, h, map[string]any{"preset": "full"})
	return f
}

func (f authzFixture) path(p string) string {
	switch {
	case strings.HasPrefix(p, "/v1/trees/{id}/comments/{commentId}"):
		return strings.NewReplacer("{id}", f.tree, "{commentId}", f.comment).Replace(p)
	case strings.HasPrefix(p, "/v1/trees/"):
		return strings.Replace(p, "{id}", f.tree, 1)
	case strings.HasPrefix(p, "/v1/scenes/"):
		return strings.Replace(p, "{id}", f.scene, 1)
	case strings.HasPrefix(p, "/v1/forum/threads/"):
		return strings.Replace(p, "{id}", f.thread, 1)
	case strings.HasPrefix(p, "/v1/keys/"):
		return strings.Replace(p, "{id}", f.keyID, 1)
	case strings.HasPrefix(p, "/v1/users/"):
		return strings.Replace(p, "{id}", "alice", 1)
	}
	return p
}

func (f authzFixture) call(actor, method, path string, body any) int {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	switch actor {
	case "owner":
		req.Header.Set("Authorization", "Bearer token-alice")
	case "other":
		req.Header.Set("Authorization", "Bearer token-bob")
	case "readkey":
		req.Header.Set("X-API-Key", f.roKey)
	case "fullkey":
		req.Header.Set("X-API-Key", f.fullKey)
	}
	rec := httptest.NewRecorder()
	f.h.ServeHTTP(rec, req)
	return rec.Code
}

// Actors, in column order: anonymous, owner (session), other signed-in user,
// owner's read-only key, owner's full key (everything except keys:*).
var authzActors = []string{"anon", "owner", "other", "readkey", "fullkey"}

// authzMatrix is the expected status per route and actor. A route without a
// row fails TestAuthzMatrix, so a new route must state who may call it.
// Bodies are minimal, so 400/422 means "authorized, rejected for content".
var authzMatrix = map[string][5]int{
	"GET /v1/health":         {200, 200, 200, 200, 200},
	"GET /v1/openapi.json":   {200, 200, 200, 200, 200},
	"GET /v1/public/trees":   {200, 200, 200, 200, 200},
	"GET /v1/node-types":     {200, 200, 200, 200, 200},
	"GET /v1/templates":      {200, 200, 200, 200, 200},
	"POST /v1/validate":      {200, 200, 200, 200, 200},
	"POST /v1/simulate":      {200, 200, 200, 200, 200},
	"POST /v1/simulate/step": {200, 200, 200, 200, 200},
	"POST /v1/render":        {200, 200, 200, 200, 200},
	"POST /v1/generate":      {422, 422, 422, 422, 422},

	"GET /v1/trees":                              {200, 200, 200, 200, 200},
	"POST /v1/trees":                             {401, 201, 201, 403, 201},
	"GET /v1/trees/{id}":                         {404, 200, 404, 200, 200},
	"PUT /v1/trees/{id}":                         {401, 200, 404, 403, 200},
	"DELETE /v1/trees/{id}":                      {401, 204, 404, 403, 204},
	"POST /v1/trees/{id}/remix":                  {401, 404, 404, 403, 404},
	"POST /v1/trees/{id}/report":                 {401, 404, 404, 403, 404},
	"PUT /v1/trees/{id}/like":                    {401, 200, 404, 403, 200},
	"DELETE /v1/trees/{id}/like":                 {401, 200, 404, 403, 200},
	"GET /v1/trees/{id}/comments":                {404, 200, 404, 200, 200},
	"POST /v1/trees/{id}/comments":               {401, 201, 404, 403, 201},
	"DELETE /v1/trees/{id}/comments/{commentId}": {401, 204, 404, 403, 204},

	"GET /v1/me":         {401, 200, 200, 200, 200},
	"PUT /v1/me":         {401, 400, 400, 403, 400},
	"GET /v1/me/trees":   {401, 200, 200, 200, 200},
	"GET /v1/me/stats":   {401, 200, 200, 200, 200},
	"GET /v1/users/{id}": {200, 200, 200, 200, 200},

	"GET /v1/forum/threads":              {200, 200, 200, 200, 200},
	"POST /v1/forum/threads":             {401, 201, 201, 403, 201},
	"GET /v1/forum/threads/{id}":         {200, 200, 200, 200, 200},
	"POST /v1/forum/threads/{id}/posts":  {401, 201, 201, 403, 201},
	"POST /v1/forum/threads/{id}/report": {401, 400, 400, 403, 400},
	"GET /v1/forum/reports":              {401, 403, 403, 403, 403},
	"GET /v1/forum/threads/{id}/reports": {401, 403, 403, 403, 403},

	"GET /v1/scenes":             {401, 200, 200, 200, 200},
	"POST /v1/scenes":            {401, 201, 201, 403, 201},
	"GET /v1/scenes/{id}":        {404, 200, 404, 200, 200},
	"PUT /v1/scenes/{id}":        {401, 200, 404, 403, 200},
	"DELETE /v1/scenes/{id}":     {401, 204, 404, 403, 204},
	"POST /v1/scenes/{id}/remix": {401, 404, 404, 403, 404},

	"GET /v1/keys":         {401, 200, 200, 403, 403},
	"POST /v1/keys":        {401, 201, 201, 403, 403},
	"DELETE /v1/keys/{id}": {401, 204, 404, 403, 403},
}

func TestAuthzMatrix(t *testing.T) {
	body := map[string]any{"spec": goodSpec, "body": "x", "title": "t", "name": "n", "preset": "read-only"}
	seen := map[string]bool{}
	for _, r := range (&Server{}).Routes() {
		if r.Path == "/mcp" {
			continue // MCP checks a scope per tool; see TestEveryRouteAndToolHasAScope
		}
		k := r.Method + " " + r.Path
		seen[k] = true
		want, ok := authzMatrix[k]
		require.True(t, ok, "route %s has no row in authzMatrix (authz_matrix_test.go)", k)
		for i, actor := range authzActors {
			f := newAuthzFixture(t)
			assert.Equal(t, want[i], f.call(actor, r.Method, f.path(r.Path), body), "%s as %s", k, actor)
		}
	}
	for k := range authzMatrix {
		assert.True(t, seen[k], "authzMatrix lists %s but no such route exists", k)
	}
}
