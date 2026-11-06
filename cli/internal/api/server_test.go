package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	c := Claims{Sub: n, Username: n, Email: n + "@example.com"}
	if n == "mod" {
		c.Groups = []string{ModeratorGroup}
	}
	return c, nil
}

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	n := 0
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := &Server{
		Store: NewMemStore(), Auth: fakeAuth{}, Version: "test", AllowAnonymous: true,
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
	assert.NotContains(t, me, "moderator", "non-moderators get no moderator flag")
	mod := do(h, "GET", "/v1/me", "mod", nil).JSON(t)
	assert.Equal(t, true, mod["moderator"])

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

	rp := map[string]string{"reason": "spam"}
	assert.Equal(t, 401, do(h, "POST", "/v1/forum/threads/"+id+"/report", "", rp).Code)
	assert.Equal(t, 400, do(h, "POST", "/v1/forum/threads/"+id+"/report", "bob", map[string]string{}).Code)
	assert.Equal(t, 404, do(h, "POST", "/v1/forum/threads/nope/report", "bob", rp).Code)
	assert.Equal(t, 404, do(h, "POST", "/v1/forum/threads/"+id+"/report", "bob", map[string]string{"reason": "x", "postId": "nope"}).Code)
	assert.Equal(t, 201, do(h, "POST", "/v1/forum/threads/"+id+"/report", "bob", rp).Code)

	q := "/v1/forum/threads/" + id + "/reports"
	assert.Equal(t, 401, do(h, "GET", q, "", nil).Code)
	assert.Equal(t, 403, do(h, "GET", q, "bob", nil).Code, "members cannot read the queue")
	assert.Equal(t, 404, do(h, "GET", "/v1/forum/threads/nope/reports", "mod", nil).Code)
	queue := do(h, "GET", q, "mod", nil)
	require.Equal(t, 200, queue.Code)
	items := queue.JSON(t)["items"].([]any)
	require.Len(t, items, 1)
	assert.Equal(t, "spam", items[0].(map[string]any)["reason"])

	all := "/v1/forum/reports"
	assert.Equal(t, 401, do(h, "GET", all, "", nil).Code)
	assert.Equal(t, 403, do(h, "GET", all, "bob", nil).Code, "members cannot read the cross-thread queue")
	recent := do(h, "GET", all, "mod", nil)
	require.Equal(t, 200, recent.Code)
	ritems := recent.JSON(t)["items"].([]any)
	require.Len(t, ritems, 1)
	assert.Equal(t, id, ritems[0].(map[string]any)["threadId"])

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

func TestSceneShareConfirmation(t *testing.T) {
	h := newTestServer(t)
	mk := func(user, vis string) string {
		r := do(h, "POST", "/v1/trees", user, map[string]any{"spec": goodSpec, "visibility": vis})
		require.Equal(t, 201, r.Code, string(r.Body))
		return r.JSON(t)["id"].(string)
	}
	priv, pub, theirs := mk("alice", "private"), mk("alice", "public"), mk("bob", "private")
	agents := []map[string]any{
		{"name": "a", "treeId": priv, "x": 0, "z": 0},
		{"name": "b", "treeId": priv, "x": 1, "z": 0},
		{"name": "c", "treeId": pub, "x": 2, "z": 0},
		{"name": "d", "treeId": theirs, "x": 3, "z": 0},
		{"name": "e", "treeId": "sample:guard", "x": 4, "z": 0},
	}
	body := map[string]any{"name": "arena", "visibility": "public", "agents": agents}

	r := do(h, "POST", "/v1/scenes", "alice", body)
	require.Equal(t, 409, r.Code, string(r.Body))
	got := r.JSON(t)
	assert.Equal(t, "confirmation_required", got["error"].(map[string]any)["code"])
	trees := got["trees"].([]any)
	require.Len(t, trees, 1, "only my private trees, each once")
	assert.Equal(t, priv, trees[0].(map[string]any)["id"])
	assert.Equal(t, 404, do(h, "GET", "/v1/trees/"+priv, "bob", nil).Code, "dry run changes nothing")
	assert.Len(t, do(h, "GET", "/v1/scenes", "alice", nil).JSON(t)["items"], 0)

	body["confirm"] = true
	r = do(h, "POST", "/v1/scenes", "alice", body)
	require.Equal(t, 201, r.Code, string(r.Body))
	id := r.JSON(t)["id"].(string)
	assert.Equal(t, 200, do(h, "GET", "/v1/trees/"+priv, "bob", nil).Code, "confirmed save publishes the tree")
	assert.Equal(t, 404, do(h, "GET", "/v1/trees/"+theirs, "alice", nil).Code, "other people's trees are never touched")

	// Unsharing the scene leaves the trees public.
	body["visibility"] = "private"
	assert.Equal(t, 200, do(h, "PUT", "/v1/scenes/"+id, "alice", body).Code)
	assert.Equal(t, 200, do(h, "GET", "/v1/trees/"+priv, "bob", nil).Code)

	// Updating to public with a new private tree asks again.
	more := mk("alice", "private")
	body["visibility"] = "public"
	body["agents"] = append(agents, map[string]any{"name": "f", "treeId": more, "x": 5, "z": 0})
	body["confirm"] = false
	r = do(h, "PUT", "/v1/scenes/"+id, "alice", body)
	require.Equal(t, 409, r.Code)
	assert.Equal(t, more, r.JSON(t)["trees"].([]any)[0].(map[string]any)["id"])
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

func TestOpenAPIDocumentsRateLimitAndSizeResponses(t *testing.T) {
	var doc struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(openapiYAML, &doc))
	for path, ops := range doc.Paths {
		for method, raw := range ops {
			if method == "parameters" {
				continue
			}
			op, ok := raw.(map[string]any)
			require.True(t, ok, "%s %s", method, path)
			responses, _ := op["responses"].(map[string]any)
			assert.Contains(t, responses, "429", "%s %s must document 429", strings.ToUpper(method), path)
			if _, hasBody := op["requestBody"]; hasBody {
				assert.Contains(t, responses, "413", "%s %s takes a body and must document 413", strings.ToUpper(method), path)
			}
		}
	}
}

// TestListTreesSearchSeesOldTrees guards against a fixed scan window: with
// more than 200 public trees the oldest one must still be findable by query,
// tag and popularity.
func TestListTreesSearchSeesOldTrees(t *testing.T) {
	st := NewMemStore()
	s := &Server{Store: st, Auth: fakeAuth{}, Version: "test", AllowAnonymous: true}
	h := s.Handler()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 205; i++ {
		tr := Tree{ID: fmt.Sprintf("t%03d", i), OwnerID: "u1", Name: fmt.Sprintf("filler %d", i),
			Visibility: VisibilityPublic, CreatedAt: base.Add(time.Duration(i) * time.Minute).Format(time.RFC3339)}
		if i == 0 {
			tr.Name, tr.Tags, tr.LikeCount = "ancient guard", []string{"relic"}, 99
		}
		require.NoError(t, st.PutTree(context.Background(), tr))
	}
	first := func(q string) string {
		items := do(h, "GET", "/v1/trees?"+q, "", nil).JSON(t)["items"].([]any)
		require.NotEmpty(t, items, q)
		return items[0].(map[string]any)["id"].(string)
	}
	assert.Equal(t, "t000", first("q=ancient"))
	assert.Equal(t, "t000", first("tag=relic"))
	assert.Equal(t, "t000", first("sort=popular"))
}

func TestListTreesCursorPagination(t *testing.T) {
	h := newTestServer(t)
	for i := 0; i < 150; i++ {
		r := do(h, "POST", "/v1/trees", "alice", map[string]any{"name": "t", "visibility": "public", "spec": goodSpec})
		require.Equal(t, http.StatusCreated, r.Code, string(r.Body))
	}
	seen := map[string]bool{}
	path := "/v1/trees?limit=40"
	pages := 0
	for path != "" {
		r := do(h, "GET", path, "", nil)
		require.Equal(t, http.StatusOK, r.Code)
		m := r.JSON(t)
		for _, it := range m["items"].([]any) {
			id := it.(map[string]any)["id"].(string)
			assert.False(t, seen[id], "tree %s repeated", id)
			seen[id] = true
		}
		pages++
		path = ""
		if c, ok := m["next_cursor"].(string); ok {
			path = "/v1/trees?limit=40&cursor=" + c
		}
	}
	assert.Len(t, seen, 150)
	assert.Equal(t, 4, pages)

	for _, bad := range []string{"limit=0", "limit=abc", "limit=101", "cursor=!!"} {
		assert.Equal(t, http.StatusBadRequest, do(h, "GET", "/v1/trees?"+bad, "", nil).Code, bad)
	}
}

func TestListTemplates(t *testing.T) {
	h := newTestServer(t)
	r := do(h, "GET", "/v1/templates", "alice", nil)
	require.Equal(t, 200, r.Code)
	list, _ := r.JSON(t)["templates"].([]any)
	require.Len(t, list, 5)
	for _, it := range list {
		m := it.(map[string]any)
		assert.NotEmpty(t, m["name"])
		assert.NotEmpty(t, m["description"])
		v := do(h, "POST", "/v1/validate", "alice", map[string]any{"spec": m["spec"]})
		assert.Equal(t, true, v.JSON(t)["valid"], "template %v must validate", m["name"])
	}
}

func TestWriteRateLimitReturns429(t *testing.T) {
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := &Server{
		Store: NewMemStore(), Auth: fakeAuth{}, Version: "test", AllowAnonymous: true,
		Now: func() time.Time { return clock }, WriteLimit: 3, WriteWindow: time.Minute,
	}
	h := s.Handler()
	post := func(user string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/v1/trees", strings.NewReader("{}"))
		req.Header.Set("Authorization", "Bearer token-"+user)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	for i := 0; i < 3; i++ {
		assert.NotEqual(t, 429, post("alice").Code)
	}
	rec := post("alice")
	require.Equal(t, 429, rec.Code)
	assert.Equal(t, "60", rec.Header().Get("Retry-After"))
	assert.Contains(t, rec.Body.String(), "rate_limited")
	assert.NotEqual(t, 429, post("bob").Code, "limits are per user")
	get := httptest.NewRequest("GET", "/v1/health", nil)
	get.Header.Set("Authorization", "Bearer token-alice")
	grec := httptest.NewRecorder()
	h.ServeHTTP(grec, get)
	assert.Equal(t, 200, grec.Code)
	clock = clock.Add(61 * time.Second)
	assert.NotEqual(t, 429, post("alice").Code, "window resets")
}

func TestProfileAvatarAndMarketingConsent(t *testing.T) {
	h := newTestServer(t)
	put := func(body map[string]any) resp {
		body["displayName"] = "Alice_B"
		return do(h, "PUT", "/v1/me", "alice", body)
	}
	for _, bad := range []string{"a000", "a101", "a1", "../x", "<b>"} {
		assert.Equal(t, 400, put(map[string]any{"avatar": bad}).Code, bad)
	}
	for _, good := range []string{"a001", "a010", "a099", "a100"} {
		require.Equal(t, 200, put(map[string]any{"avatar": good}).Code, good)
	}
	me := do(h, "GET", "/v1/me", "alice", nil).JSON(t)
	assert.Equal(t, "a100", me["avatar"])
	assert.Equal(t, false, me["marketingOptIn"], "consent is opt-in, never assumed")

	require.Equal(t, 200, put(map[string]any{"marketingOptIn": true, "marketingSource": "signup"}).Code)
	me = do(h, "GET", "/v1/me", "alice", nil).JSON(t)
	assert.Equal(t, true, me["marketingOptIn"])
	assert.Equal(t, "signup", me["marketingSource"])
	assert.NotEmpty(t, me["marketingOptInAt"])
	assert.Equal(t, "a100", me["avatar"], "omitted fields keep their value")

	require.Equal(t, 200, put(map[string]any{"marketingOptIn": false}).Code)
	me = do(h, "GET", "/v1/me", "alice", nil).JSON(t)
	assert.Equal(t, false, me["marketingOptIn"], "consent can be withdrawn")
	assert.Equal(t, "profile", me["marketingSource"])

	pub := do(h, "GET", "/v1/users/alice", "", nil)
	assert.NotContains(t, string(pub.Body), "marketing", "the consent record is private")
	assert.Contains(t, string(pub.Body), "a100", "the avatar is public")
}

func TestPlainText(t *testing.T) {
	assert.Equal(t, "a\nb", plainText("a\r\nb", 100, false))
	assert.Equal(t, "a\n\nb", plainText("a\n\n\n\n\nb", 100, false), "blank-line runs collapse")
	assert.Equal(t, "hi there", plainText("hi\x00\x07 there", 100, false), "control characters go")
	assert.Equal(t, "ab", plainText("a\u200b\u202eb", 100, false), "invisible and direction tricks go")
	assert.Equal(t, "one two three", plainText("one\ntwo\tthree", 100, true), "titles stay on one line")
	assert.Equal(t, "<b>x</b>", plainText("<b>x</b>", 100, false), "markup is kept as literal text, never interpreted")
	assert.Equal(t, "abc", plainText("abcdef", 3, false))
}

func TestForumKeepsTextOnly(t *testing.T) {
	h := newTestServer(t)
	r := do(h, "POST", "/v1/forum/threads", "alice", map[string]string{"title": "Hello\nworld", "body": "line1\r\n\x00line2"})
	require.Equal(t, 201, r.Code)
	got := r.JSON(t)
	assert.Equal(t, "Hello world", got["title"])
	assert.Equal(t, "line1\nline2", got["body"])
}

func TestRemixTree(t *testing.T) {
	h := newTestServer(t)
	pub := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "name": "Patrol", "visibility": "public"}).JSON(t)
	id := pub["id"].(string)
	priv := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "name": "Hidden"}).JSON(t)

	assert.Equal(t, 401, do(h, "POST", "/v1/trees/"+id+"/remix", "", nil).Code)
	assert.Equal(t, 404, do(h, "POST", "/v1/trees/"+priv["id"].(string)+"/remix", "bob", nil).Code, "private tree is 404")
	assert.Equal(t, 404, do(h, "POST", "/v1/trees/"+priv["id"].(string)+"/remix", "alice", nil).Code)

	r := do(h, "POST", "/v1/trees/"+id+"/remix", "bob", nil)
	require.Equal(t, 201, r.Code, string(r.Body))
	cp := r.JSON(t)
	assert.Equal(t, "private", cp["visibility"])
	assert.Equal(t, "Patrol (Remix)", cp["name"])
	assert.Equal(t, "bob", cp["ownerId"])
	lin := cp["remixedFrom"].(map[string]any)
	assert.Equal(t, id, lin["id"])
	assert.Equal(t, "Patrol", lin["title"])
	assert.Equal(t, "alice", lin["ownerName"])
	assert.Equal(t, "public", lin["state"])
	cpID := cp["id"].(string)
	state := func() string {
		return do(h, "GET", "/v1/trees/"+cpID, "bob", nil).JSON(t)["remixedFrom"].(map[string]any)["state"].(string)
	}

	// lineage survives the original going private, flips back when re-shared
	do(h, "PUT", "/v1/trees/"+id, "alice", map[string]any{"spec": goodSpec, "name": "Renamed", "visibility": "private"})
	assert.Equal(t, "private", state())
	got := do(h, "GET", "/v1/trees/"+cpID, "bob", nil).JSON(t)["remixedFrom"].(map[string]any)
	assert.Equal(t, "Patrol", got["title"], "frozen title kept")
	do(h, "PUT", "/v1/trees/"+id, "alice", map[string]any{"spec": goodSpec, "name": "Renamed", "visibility": "public"})
	assert.Equal(t, "public", state())

	// remix of a remix records the chain
	do(h, "PUT", "/v1/trees/"+cpID, "bob", map[string]any{"spec": strings.Replace(goodSpec, "name: ", "name: x", 1), "name": "Mine", "visibility": "public"})
	r2 := do(h, "POST", "/v1/trees/"+cpID+"/remix", "carol", nil).JSON(t)["remixedFrom"].(map[string]any)
	assert.Equal(t, "Mine", r2["title"])
	require.Len(t, r2["ancestors"], 1)
	assert.Equal(t, id, r2["ancestors"].([]any)[0].(map[string]any)["id"])

	// deleted original
	do(h, "DELETE", "/v1/trees/"+id, "alice", nil)
	assert.Equal(t, "deleted", state())
	assert.Equal(t, "Patrol", do(h, "GET", "/v1/trees/"+cpID, "bob", nil).JSON(t)["remixedFrom"].(map[string]any)["title"])
}

func TestRemixRateLimit(t *testing.T) {
	h := newTestServer(t)
	id := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "visibility": "public"}).JSON(t)["id"].(string)
	for i := 0; i < RemixLimit; i++ {
		require.Equal(t, 201, do(h, "POST", "/v1/trees/"+id+"/remix", "bob", nil).Code, "remix %d", i)
	}
	assert.Equal(t, 429, do(h, "POST", "/v1/trees/"+id+"/remix", "bob", nil).Code)
	assert.Equal(t, 201, do(h, "POST", "/v1/trees/"+id+"/remix", "carol", nil).Code, "limit is per user")
}

func TestPublishUnchangedRemixBlocked(t *testing.T) {
	h := newTestServer(t)
	id := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "name": "Patrol", "visibility": "public"}).JSON(t)["id"].(string)
	cp := do(h, "POST", "/v1/trees/"+id+"/remix", "bob", nil).JSON(t)
	cpID := cp["id"].(string)
	put := func(spec, vis string) int {
		return do(h, "PUT", "/v1/trees/"+cpID, "bob", map[string]any{"spec": spec, "name": "Mine", "visibility": vis}).Code
	}
	assert.Equal(t, 200, put(goodSpec, "private"), "private copy may stay identical")
	r := do(h, "PUT", "/v1/trees/"+cpID, "bob", map[string]any{"spec": goodSpec, "name": "Mine", "visibility": "public"})
	assert.Equal(t, 422, r.Code)
	assert.Contains(t, string(r.Body), "unchanged_remix")
	assert.Equal(t, 422, put("\n"+goodSpec+"\n\n", "public"), "whitespace-only edits do not count")
	changed := strings.Replace(goodSpec, "name: ", "name: x", 1)
	assert.Equal(t, 200, put(changed, "public"), "a real change publishes")
}

func TestRemixScene(t *testing.T) {
	h := newTestServer(t)
	tr := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "name": "Patrol", "visibility": "public"}).JSON(t)["id"].(string)
	priv := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "name": "Hidden"}).JSON(t)["id"].(string)
	agents := []any{
		map[string]any{"name": "a", "treeId": tr, "model": "bee", "x": 1, "z": 1},
		map[string]any{"name": "b", "treeId": tr, "model": "bee", "x": 2, "z": 2},
		map[string]any{"name": "c", "treeId": priv, "model": "bee", "x": 3, "z": 3},
	}
	pubScene := do(h, "POST", "/v1/scenes", "alice", map[string]any{"name": "Meadow", "visibility": "public", "agents": agents})
	require.Equal(t, 409, pubScene.Code, "private tree needs confirm")
	sceneID := do(h, "POST", "/v1/scenes", "alice", map[string]any{"name": "Meadow", "visibility": "public", "agents": agents[:2]}).JSON(t)["id"].(string)
	privScene := do(h, "POST", "/v1/scenes", "alice", map[string]any{"name": "Hid", "agents": []any{}}).JSON(t)["id"].(string)

	assert.Equal(t, 401, do(h, "POST", "/v1/scenes/"+sceneID+"/remix", "", nil).Code)
	assert.Equal(t, 404, do(h, "POST", "/v1/scenes/"+privScene+"/remix", "bob", nil).Code)

	r := do(h, "POST", "/v1/scenes/"+sceneID+"/remix", "bob", nil)
	require.Equal(t, 201, r.Code, string(r.Body))
	got := r.JSON(t)
	assert.Equal(t, "Meadow (Remix)", got["name"])
	assert.Equal(t, "private", got["visibility"])
	assert.Equal(t, "bob", got["ownerId"])
	lin := got["remixedFrom"].(map[string]any)
	assert.Equal(t, "scene", lin["kind"])
	assert.Equal(t, sceneID, lin["id"])
	assert.Equal(t, "public", lin["state"])
	trees := got["trees"].([]any)
	require.Len(t, trees, 1, "a tree shared by two agents is copied once")
	cp := trees[0].(map[string]any)
	assert.Equal(t, "private", cp["visibility"])
	assert.Equal(t, tr, cp["remixedFrom"].(map[string]any)["id"])
	ag := got["agents"].([]any)
	require.Len(t, ag, 2)
	for _, a := range ag {
		assert.Equal(t, cp["id"], a.(map[string]any)["treeId"])
	}
	// the copy is stored and readable by its owner only
	assert.Equal(t, 200, do(h, "GET", "/v1/scenes/"+got["id"].(string), "bob", nil).Code)
	assert.Equal(t, 404, do(h, "GET", "/v1/scenes/"+got["id"].(string), "carol", nil).Code)

	// original goes private: lineage state follows
	do(h, "PUT", "/v1/scenes/"+sceneID, "alice", map[string]any{"name": "Meadow", "visibility": "private", "agents": agents[:2]})
	assert.Equal(t, "private", do(h, "GET", "/v1/scenes/"+got["id"].(string), "bob", nil).JSON(t)["remixedFrom"].(map[string]any)["state"])
}

func TestRemixSceneCountsOnce(t *testing.T) {
	h := newTestServer(t)
	tr := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "visibility": "public"}).JSON(t)["id"].(string)
	ag := []any{map[string]any{"name": "a", "treeId": tr, "model": "bee"}, map[string]any{"name": "b", "treeId": tr, "model": "bee"}}
	sc := do(h, "POST", "/v1/scenes", "alice", map[string]any{"name": "S", "visibility": "public", "agents": ag}).JSON(t)["id"].(string)
	for i := 0; i < RemixLimit; i++ {
		require.Equal(t, 201, do(h, "POST", "/v1/scenes/"+sc+"/remix", "bob", nil).Code, "remix %d", i)
	}
	assert.Equal(t, 429, do(h, "POST", "/v1/scenes/"+sc+"/remix", "bob", nil).Code)
}

func TestReportTreeAndUnchangedLabel(t *testing.T) {
	h := newTestServer(t)
	id := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec, "name": "Patrol", "visibility": "public"}).JSON(t)["id"].(string)
	cpID := do(h, "POST", "/v1/trees/"+id+"/remix", "bob", nil).JSON(t)["id"].(string)
	label := func() any { return do(h, "GET", "/v1/trees/"+cpID, "bob", nil).JSON(t)["unchangedCopy"] }

	assert.Nil(t, label(), "private copy is not labelled")
	do(h, "PUT", "/v1/trees/"+cpID, "bob", map[string]any{"spec": strings.Replace(goodSpec, "name: ", "name: x", 1), "name": "Mine", "visibility": "public"})
	assert.Nil(t, label(), "a changed public remix is not labelled")
	assert.Nil(t, do(h, "GET", "/v1/trees/"+id, "", nil).JSON(t)["unchangedCopy"], "originals are not labelled")

	// The original later adopts the remix's spec: the public copy now reads as unchanged.
	do(h, "PUT", "/v1/trees/"+id, "alice", map[string]any{"spec": strings.Replace(goodSpec, "name: ", "name: x", 1), "name": "Patrol", "visibility": "public"})
	assert.Equal(t, true, label())

	assert.Equal(t, 401, do(h, "POST", "/v1/trees/"+cpID+"/report", "", map[string]any{"reason": "copy"}).Code)
	assert.Equal(t, 400, do(h, "POST", "/v1/trees/"+cpID+"/report", "carol", map[string]any{}).Code)
	assert.Equal(t, 201, do(h, "POST", "/v1/trees/"+cpID+"/report", "carol", map[string]any{"reason": "exact copy"}).Code)
	assert.Equal(t, 201, do(h, "POST", "/v1/trees/"+cpID+"/report", "carol", map[string]any{"reason": "again"}).Code)
	priv := do(h, "POST", "/v1/trees", "alice", map[string]any{"spec": goodSpec}).JSON(t)["id"].(string)
	assert.Equal(t, 404, do(h, "POST", "/v1/trees/"+priv+"/report", "carol", map[string]any{"reason": "x"}).Code)
}
