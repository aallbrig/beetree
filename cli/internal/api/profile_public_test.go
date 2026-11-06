package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A profile is readable signed out, and shows only public trees and scenes.
func TestAnonymousProfileLeaksNothingPrivate(t *testing.T) {
	s := &Server{Store: NewMemStore(), Auth: fakeAuth{}, Version: "test"}
	h := s.Handler()

	require.Equal(t, 401, do(h, "GET", "/v1/trees/x", "", nil).Code, "other routes stay closed")
	for _, vis := range []string{"public", "private"} {
		require.Equal(t, 201, do(h, "POST", "/v1/scenes", "alice", map[string]any{"name": vis + "-scene", "visibility": vis}).Code)
		require.Equal(t, 201, do(h, "POST", "/v1/trees", "alice", map[string]any{
			"name": vis + "-tree", "visibility": vis,
			"spec": goodSpec,
		}).Code)
	}
	r := do(h, "GET", "/v1/users/alice", "", nil)
	require.Equal(t, 200, r.Code, string(r.Body))
	body := string(r.Body)
	assert.Contains(t, body, "public-scene")
	assert.NotContains(t, body, "private-scene")
	assert.NotContains(t, body, "private-tree")
	assert.NotContains(t, body, "alice@example.com")
	assert.Equal(t, 404, do(h, "GET", "/v1/users/ghost", "", nil).Code)
}
