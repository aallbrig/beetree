package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const goodSpec = `version: "1.0"
metadata:
  name: patrol
  description: walk around
  tags: [Patrol, demo]
tree:
  type: sequence
  name: root
  children:
    - type: condition
      name: has_waypoint
    - type: action
      name: move
`

// fakeAuth treats the token as the user id: "token-alice" -> sub "alice".
type fakeAuth struct{}

func (fakeAuth) Verify(_ context.Context, token string) (Claims, error) {
	if !strings.HasPrefix(token, "token-") {
		return Claims{}, ErrUnauthenticated
	}
	n := strings.TrimPrefix(token, "token-")
	return Claims{Sub: n, Username: n, Email: n + "@example.com"}, nil
}

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	n := 0
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := &Server{
		Store: NewMemStore(), Auth: fakeAuth{}, Version: "test",
		Now:   func() time.Time { clock = clock.Add(time.Second); return clock },
		NewID: func() string { n++; return "id" + string(rune('a'+n)) },
	}
	return s.Handler()
}

type resp struct {
	Code int
	Body []byte
}

func (r resp) JSON(t *testing.T) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(r.Body, &m), string(r.Body))
	return m
}

func do(h http.Handler, method, path, user string, body any) resp {
	var rdr *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	} else {
		rdr = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rdr)
	if user != "" {
		req.Header.Set("Authorization", "Bearer token-"+user)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return resp{w.Code, w.Body.Bytes()}
}

// TestOpenAPIMatchesRoutes keeps the published spec honest: every route has an
// operation and every operation has a route.
func TestOpenAPIMatchesRoutes(t *testing.T) {
	var doc struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(openapiYAML, &doc))

	inSpec := map[string]bool{}
	for path, ops := range doc.Paths {
		for method := range ops {
			if method == "parameters" {
				continue
			}
			inSpec[strings.ToUpper(method)+" "+path] = true
		}
	}
	inCode := map[string]bool{}
	for _, r := range (&Server{}).Routes() {
		inCode[r.Method+" "+r.Path] = true
	}
	for k := range inCode {
		assert.True(t, inSpec[k], "route %s is missing from openapi.yaml", k)
	}
	for k := range inSpec {
		assert.True(t, inCode[k], "openapi.yaml documents %s but no such route exists", k)
	}
}

func TestOpenAPIServedAsJSON(t *testing.T) {
	r := do(newTestServer(t), "GET", "/v1/openapi.json", "", nil)
	require.Equal(t, 200, r.Code)
	assert.Equal(t, "3.0.3", r.JSON(t)["openapi"])
}

func TestHealthAndCORS(t *testing.T) {
	h := newTestServer(t)
	r := do(h, "GET", "/v1/health", "", nil)
	assert.Equal(t, 200, r.Code)
	assert.Equal(t, "ok", r.JSON(t)["status"])

	req := httptest.NewRequest("OPTIONS", "/v1/trees", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, 204, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

	assert.Equal(t, 404, do(h, "GET", "/nope", "", nil).Code)
}

func TestBadTokenIsRejected(t *testing.T) {
	h := newTestServer(t)
	req := httptest.NewRequest("GET", "/v1/trees", nil)
	req.Header.Set("Authorization", "Bearer garbage")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestStatelessTools(t *testing.T) {
	h := newTestServer(t)

	r := do(h, "POST", "/v1/validate", "", map[string]string{"spec": goodSpec})
	assert.Equal(t, true, r.JSON(t)["valid"])

	r = do(h, "POST", "/v1/validate", "", map[string]string{"spec": "tree:\n  type: sequence\n  name: x\n"})
	j := r.JSON(t)
	assert.Equal(t, false, j["valid"])
	assert.NotEmpty(t, j["errors"])

	r = do(h, "POST", "/v1/validate", "", map[string]string{"spec": ":::not yaml"})
	assert.Equal(t, 200, r.Code)
	assert.Equal(t, false, r.JSON(t)["valid"])

	r = do(h, "POST", "/v1/simulate", "", map[string]any{"spec": goodSpec, "overrides": map[string]string{"move": "failure"}})
	require.Equal(t, 200, r.Code, string(r.Body))
	assert.Equal(t, "FAILURE", r.JSON(t)["status"])

	r = do(h, "POST", "/v1/simulate", "", map[string]any{"spec": goodSpec, "overrides": map[string]string{"move": "maybe"}})
	assert.Equal(t, 422, r.Code)

	for _, f := range RenderFormats {
		r = do(h, "POST", "/v1/render", "", map[string]string{"spec": goodSpec, "format": f})
		assert.Equal(t, 200, r.Code, f)
		assert.NotEmpty(t, r.JSON(t)["output"], f)
	}
	assert.Equal(t, 422, do(h, "POST", "/v1/render", "", map[string]string{"spec": goodSpec, "format": "pdf"}).Code)

	for _, e := range Engines {
		r = do(h, "POST", "/v1/generate", "", map[string]string{"spec": goodSpec, "engine": e})
		require.Equal(t, 200, r.Code, e+string(r.Body))
		assert.NotEmpty(t, r.JSON(t)["files"], e)
	}
	assert.Equal(t, 422, do(h, "POST", "/v1/generate", "", map[string]string{"spec": goodSpec, "engine": "cryengine"}).Code)

	r = do(h, "GET", "/v1/node-types", "", nil)
	assert.Contains(t, string(r.Body), "sequence")
}

func TestRejectsOversizedSpec(t *testing.T) {
	h := newTestServer(t)
	r := do(h, "POST", "/v1/validate", "", map[string]string{"spec": strings.Repeat("a", maxSpecBytes+1)})
	assert.Equal(t, false, r.JSON(t)["valid"])
}

func TestTreeLifecycleAndOwnership(t *testing.T) {
	h := newTestServer(t)

	assert.Equal(t, 401, do(h, "POST", "/v1/trees", "", map[string]any{"spec": goodSpec}).Code)
	assert.Equal(t, 422, do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": "garbage: ["}).Code)
	assert.Equal(t, 400, do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "visibility": "secret"}).Code)

	r := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "visibility": "public"})
	require.Equal(t, 201, r.Code, string(r.Body))
	tree := r.JSON(t)
	id := tree["id"].(string)
	assert.Equal(t, "patrol", tree["name"], "name defaults to spec metadata")
	assert.Equal(t, []any{"patrol", "demo"}, tree["tags"], "tags default to spec metadata, lowercased")
	assert.Equal(t, "alice", tree["ownerName"])

	priv := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "name": "secret one"}).JSON(t)
	privID := priv["id"].(string)
	assert.Equal(t, "private", priv["visibility"], "trees default to private")

	// public listing hides private trees and omits spec bodies
	list := do(h, "GET", "/v1/trees", "", nil).JSON(t)["items"].([]any)
	require.Len(t, list, 1)
	assert.Nil(t, list[0].(map[string]any)["spec"])

	// private tree: visible to owner, 404 to everyone else
	assert.Equal(t, 200, do(h, "GET", "/v1/trees/"+privID, "alice", nil).Code)
	assert.Equal(t, 404, do(h, "GET", "/v1/trees/"+privID, "bob", nil).Code)
	assert.Equal(t, 404, do(h, "GET", "/v1/trees/"+privID, "", nil).Code)

	// only the owner may modify
	assert.Equal(t, 404, do(h, "PUT", "/v1/trees/"+id, "bob", map[string]any{"spec": goodSpec}).Code)
	assert.Equal(t, 404, do(h, "DELETE", "/v1/trees/"+id, "bob", nil).Code)
	r = do(h, "PUT", "/v1/trees/"+id, "alice", map[string]any{"spec": goodSpec, "name": "renamed", "visibility": "public"})
	require.Equal(t, 200, r.Code)
	assert.Equal(t, "renamed", r.JSON(t)["name"])

	mine := do(h, "GET", "/v1/me/trees", "alice", nil).JSON(t)["items"].([]any)
	assert.Len(t, mine, 2)

	assert.Equal(t, 204, do(h, "DELETE", "/v1/trees/"+id, "alice", nil).Code)
	assert.Equal(t, 404, do(h, "GET", "/v1/trees/"+id, "", nil).Code)
}

func TestSearchAndSort(t *testing.T) {
	h := newTestServer(t)
	a := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "name": "Zombie AI", "visibility": "public", "tags": []string{"enemy"}}).JSON(t)
	do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "name": "Shopkeeper", "visibility": "public", "tags": []string{"npc"}})
	do(h, "PUT", "/v1/trees/"+a["id"].(string)+"/like", "bob", nil)

	items := func(q string) []any { return do(h, "GET", "/v1/trees?"+q, "", nil).JSON(t)["items"].([]any) }
	assert.Len(t, items("q=zombie"), 1)
	assert.Len(t, items("tag=npc"), 1)
	assert.Len(t, items("q=nothing"), 0)
	assert.Equal(t, "Zombie AI", items("sort=popular")[0].(map[string]any)["name"])
	assert.Len(t, items("limit=1"), 1)
}

func TestLikesAndComments(t *testing.T) {
	h := newTestServer(t)
	id := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "visibility": "public"}).JSON(t)["id"].(string)

	assert.Equal(t, 401, do(h, "PUT", "/v1/trees/"+id+"/like", "", nil).Code)
	r := do(h, "PUT", "/v1/trees/"+id+"/like", "bob", nil)
	assert.Equal(t, float64(1), r.JSON(t)["likeCount"])
	r = do(h, "PUT", "/v1/trees/"+id+"/like", "bob", nil)
	assert.Equal(t, float64(1), r.JSON(t)["likeCount"], "double like counts once")
	assert.Equal(t, true, do(h, "GET", "/v1/trees/"+id, "bob", nil).JSON(t)["likedByMe"])
	r = do(h, "DELETE", "/v1/trees/"+id+"/like", "bob", nil)
	assert.Equal(t, float64(0), r.JSON(t)["likeCount"])

	assert.Equal(t, 400, do(h, "POST", "/v1/trees/"+id+"/comments", "bob", map[string]string{"body": "  "}).Code)
	cm := do(h, "POST", "/v1/trees/"+id+"/comments", "bob", map[string]string{"body": "great tree"})
	require.Equal(t, 201, cm.Code)
	cid := cm.JSON(t)["id"].(string)
	assert.Len(t, do(h, "GET", "/v1/trees/"+id+"/comments", "", nil).JSON(t)["items"], 1)
	assert.Equal(t, float64(1), do(h, "GET", "/v1/trees/"+id, "", nil).JSON(t)["commentCount"])

	// a third party cannot delete; the author and the tree owner can
	assert.Equal(t, 403, do(h, "DELETE", "/v1/trees/"+id+"/comments/"+cid, "carol", nil).Code)
	assert.Equal(t, 204, do(h, "DELETE", "/v1/trees/"+id+"/comments/"+cid, "alice", nil).Code)
	assert.Equal(t, 404, do(h, "DELETE", "/v1/trees/"+id+"/comments/"+cid, "bob", nil).Code)

	// cannot like or comment on a private tree you cannot see
	priv := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec}).JSON(t)["id"].(string)
	assert.Equal(t, 404, do(h, "PUT", "/v1/trees/"+priv+"/like", "bob", nil).Code)
	assert.Equal(t, 404, do(h, "POST", "/v1/trees/"+priv+"/comments", "bob", map[string]string{"body": "hi"}).Code)
}

func TestProfileAndStats(t *testing.T) {
	h := newTestServer(t)
	assert.Equal(t, 401, do(h, "GET", "/v1/me", "", nil).Code)

	me := do(h, "GET", "/v1/me", "alice", nil).JSON(t)
	assert.Equal(t, "alice", me["displayName"])
	assert.Equal(t, "alice@example.com", me["email"])

	assert.Equal(t, 400, do(h, "PUT", "/v1/me", "alice", map[string]string{"displayName": "x"}).Code)
	assert.Equal(t, 400, do(h, "PUT", "/v1/me", "alice", map[string]string{"displayName": "<script>alert(1)</script>"}).Code)
	r := do(h, "PUT", "/v1/me", "alice", map[string]string{"displayName": "Alice_B", "bio": "I like trees"})
	require.Equal(t, 200, r.Code)

	id := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "visibility": "public"}).JSON(t)
	assert.Equal(t, "Alice_B", id["ownerName"], "content carries the display name")
	do(h, "PUT", "/v1/trees/"+id["id"].(string)+"/like", "bob", nil)
	do(h, "POST", "/v1/forum/threads", "alice", map[string]string{"title": "hi", "body": "there"})

	st := do(h, "GET", "/v1/me/stats", "alice", nil).JSON(t)
	assert.Equal(t, float64(1), st["trees"])
	assert.Equal(t, float64(1), st["likes"])
	assert.Equal(t, float64(1), st["forumThreads"])

	u := do(h, "GET", "/v1/users/alice", "", nil)
	require.Equal(t, 200, u.Code)
	assert.NotContains(t, string(u.Body), "alice@example.com", "public profiles never expose email")
	assert.Equal(t, 404, do(h, "GET", "/v1/users/ghost", "", nil).Code)
}

func TestForum(t *testing.T) {
	h := newTestServer(t)
	assert.Equal(t, 401, do(h, "POST", "/v1/forum/threads", "", map[string]string{"title": "a", "body": "b"}).Code)
	assert.Equal(t, 400, do(h, "POST", "/v1/forum/threads", "alice", map[string]string{"title": "a"}).Code)

	th := do(h, "POST", "/v1/forum/threads", "alice", map[string]string{"title": "Best selector layout?", "body": "Thoughts?"})
	require.Equal(t, 201, th.Code)
	id := th.JSON(t)["id"].(string)

	assert.Equal(t, 201, do(h, "POST", "/v1/forum/threads/"+id+"/posts", "bob", map[string]string{"body": "Priority order."}).Code)
	assert.Equal(t, 404, do(h, "POST", "/v1/forum/threads/nope/posts", "bob", map[string]string{"body": "x"}).Code)

	got := do(h, "GET", "/v1/forum/threads/"+id, "", nil).JSON(t)
	assert.Len(t, got["posts"], 1)
	assert.Equal(t, float64(1), got["thread"].(map[string]any)["postCount"])
	assert.Len(t, do(h, "GET", "/v1/forum/threads", "", nil).JSON(t)["items"], 1)
}

func TestScenes(t *testing.T) {
	h := newTestServer(t)
	agents := []map[string]any{{"name": "a", "model": "character-a", "x": 1, "z": 2}}

	assert.Equal(t, 401, do(h, "GET", "/v1/scenes", "", nil).Code)
	assert.Equal(t, 400, do(h, "POST", "/v1/scenes", "alice", map[string]any{"name": ""}).Code)
	assert.Equal(t, 400, do(h, "POST", "/v1/scenes", "alice", map[string]any{"name": "far", "agents": []map[string]any{{"x": 999, "z": 0}}}).Code)

	r := do(h, "POST", "/v1/scenes", "alice", map[string]any{"name": "arena", "agents": agents})
	require.Equal(t, 201, r.Code, string(r.Body))
	id := r.JSON(t)["id"].(string)

	assert.Equal(t, 200, do(h, "GET", "/v1/scenes/"+id, "alice", nil).Code)
	assert.Equal(t, 404, do(h, "GET", "/v1/scenes/"+id, "bob", nil).Code, "private scenes stay private")
	assert.Equal(t, 404, do(h, "PUT", "/v1/scenes/"+id, "bob", map[string]any{"name": "x"}).Code)

	r = do(h, "PUT", "/v1/scenes/"+id, "alice", map[string]any{"name": "arena2", "visibility": "public", "agents": agents})
	assert.Equal(t, 200, r.Code)
	assert.Equal(t, 200, do(h, "GET", "/v1/scenes/"+id, "", nil).Code, "public scenes are readable by anyone")
	assert.Len(t, do(h, "GET", "/v1/scenes", "alice", nil).JSON(t)["items"], 1)

	many := make([]map[string]any, maxSceneAgents+1)
	for i := range many {
		many[i] = map[string]any{"x": 0, "z": 0}
	}
	assert.Equal(t, 400, do(h, "POST", "/v1/scenes", "alice", map[string]any{"name": "crowd", "agents": many}).Code)

	assert.Equal(t, 204, do(h, "DELETE", "/v1/scenes/"+id, "alice", nil).Code)
}

func TestBadJSON(t *testing.T) {
	h := newTestServer(t)
	req := httptest.NewRequest("POST", "/v1/validate", strings.NewReader("{nope"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestExamplesValidate(t *testing.T) {
	for _, f := range []string{"patrol", "combat", "full-enemy-ai"} {
		data, err := os.ReadFile("../../../examples/" + f + ".beetree.yaml")
		require.NoError(t, err)
		res, _ := validateYAML(string(data))
		assert.True(t, res.Valid, "%s: %v", f, res.Errors)
	}
}

func TestOpenAPIIsValidOpenAPI3(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(openapiYAML)
	require.NoError(t, err)
	require.NoError(t, doc.Validate(context.Background()))
	for path, item := range doc.Paths.Map() {
		for method, op := range item.Operations() {
			assert.NotEmpty(t, op.OperationID, "%s %s needs an operationId", method, path)
		}
	}
}
