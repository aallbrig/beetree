package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublicShowcaseIsOpenButLimited(t *testing.T) {
	store := NewMemStore()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := &Server{Store: store, Auth: fakeAuth{}, Version: "test", Now: func() time.Time { return now }, NewID: func() string { return "x" }}
	h := s.Handler()
	pub := Tree{ID: "t1", OwnerID: "u1", Name: "guard", Visibility: VisibilityPublic, CreatedAt: "2026-01-01T00:00:00Z", Spec: "secret spec body"}
	priv := Tree{ID: "t2", OwnerID: "u1", Name: "private", Visibility: VisibilityPrivate, CreatedAt: "2026-01-01T00:00:00Z"}
	require.NoError(t, store.PutTree(t.Context(), pub))
	require.NoError(t, store.PutTree(t.Context(), priv))

	r := do(h, "GET", "/v1/public/trees", "", nil)
	require.Equal(t, 200, r.Code, string(r.Body))
	body := r.JSON(t)
	items, _ := body["items"].([]any)
	require.Len(t, items, 1, "only public trees")
	assert.NotContains(t, string(r.Body), "secret spec body", "no spec bodies for visitors")
	assert.NotEmpty(t, body["generatedAt"])

	// the rest of the API still needs credentials
	assert.Equal(t, 401, do(h, "GET", "/v1/trees", "", nil).Code)
	assert.Equal(t, 401, do(h, "GET", "/v1/trees/t1", "", nil).Code)

	// the snapshot is stale by design: a tree added later is hidden until the TTL passes
	require.NoError(t, store.PutTree(t.Context(), Tree{ID: "t3", OwnerID: "u1", Name: "new", Visibility: VisibilityPublic, CreatedAt: "2026-01-01T01:00:00Z"}))
	now = now.Add(23 * time.Hour)
	items, _ = do(h, "GET", "/v1/public/trees", "", nil).JSON(t)["items"].([]any)
	assert.Len(t, items, 1, "still the old snapshot within 24 hours")
	now = now.Add(2 * time.Hour)
	items, _ = do(h, "GET", "/v1/public/trees", "", nil).JSON(t)["items"].([]any)
	assert.Len(t, items, 2, "refreshed after the TTL")
}

func TestPublicShowcaseSearchAndHeaders(t *testing.T) {
	store := NewMemStore()
	s := &Server{Store: store, Version: "test", Now: time.Now, NewID: func() string { return "x" }}
	h := s.Handler()
	require.NoError(t, store.PutTree(t.Context(), Tree{ID: "a", Name: "patrol guard", Visibility: VisibilityPublic, LikeCount: 1, CreatedAt: "2026-01-01T00:00:00Z"}))
	require.NoError(t, store.PutTree(t.Context(), Tree{ID: "b", Name: "boss fight", Visibility: VisibilityPublic, LikeCount: 9, CreatedAt: "2026-01-02T00:00:00Z"}))
	items, _ := do(h, "GET", "/v1/public/trees?q=boss", "", nil).JSON(t)["items"].([]any)
	assert.Len(t, items, 1)
	items, _ = do(h, "GET", "/v1/public/trees?sort=popular", "", nil).JSON(t)["items"].([]any)
	require.Len(t, items, 2)
	assert.Equal(t, "b", items[0].(map[string]any)["id"])
}
