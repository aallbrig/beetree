package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runStoreContract exercises the behavior every Store must provide. It is run
// against MemStore in CI and against DynamoStore when BEETREE_TEST_TABLE is set.
func runStoreContract(t *testing.T, s Store) {
	t.Helper()
	ctx := context.Background()

	t.Run("trees", func(t *testing.T) {
		pub := Tree{ID: "t1", OwnerID: "u1", Name: "pub", Visibility: VisibilityPublic, CreatedAt: "2026-01-02T00:00:00Z", Tags: []string{"a"}}
		priv := Tree{ID: "t2", OwnerID: "u1", Name: "priv", Visibility: VisibilityPrivate, CreatedAt: "2026-01-01T00:00:00Z"}
		require.NoError(t, s.PutTree(ctx, pub))
		require.NoError(t, s.PutTree(ctx, priv))

		got, err := s.GetTree(ctx, "t1")
		require.NoError(t, err)
		assert.Equal(t, "pub", got.Name)
		assert.Equal(t, []string{"a"}, got.Tags)

		_, err = s.GetTree(ctx, "missing")
		assert.ErrorIs(t, err, ErrNotFound)

		public, err := s.ListPublicTrees(ctx, 10)
		require.NoError(t, err)
		require.Len(t, public, 1)
		assert.Equal(t, "t1", public[0].ID)

		mine, err := s.ListTreesByOwner(ctx, "u1")
		require.NoError(t, err)
		assert.Len(t, mine, 2)

		// making a tree private removes it from the public listing
		pub.Visibility = VisibilityPrivate
		require.NoError(t, s.PutTree(ctx, pub))
		public, err = s.ListPublicTrees(ctx, 10)
		require.NoError(t, err)
		assert.Empty(t, public)
		pub.Visibility = VisibilityPublic
		require.NoError(t, s.PutTree(ctx, pub))
	})

	t.Run("likes are idempotent", func(t *testing.T) {
		n, err := s.SetLike(ctx, "t1", "u2", true)
		require.NoError(t, err)
		assert.Equal(t, 1, n)
		n, err = s.SetLike(ctx, "t1", "u2", true)
		require.NoError(t, err)
		assert.Equal(t, 1, n)
		n, err = s.SetLike(ctx, "t1", "u3", true)
		require.NoError(t, err)
		assert.Equal(t, 2, n)

		ok, err := s.HasLiked(ctx, "t1", "u2")
		require.NoError(t, err)
		assert.True(t, ok)

		n, err = s.SetLike(ctx, "t1", "u2", false)
		require.NoError(t, err)
		assert.Equal(t, 1, n)
		n, err = s.SetLike(ctx, "t1", "u2", false)
		require.NoError(t, err)
		assert.Equal(t, 1, n)

		got, err := s.GetTree(ctx, "t1")
		require.NoError(t, err)
		assert.Equal(t, 1, got.LikeCount)
	})

	t.Run("comments", func(t *testing.T) {
		c := Comment{ID: "c1", TreeID: "t1", AuthorID: "u2", Author: "bob", Body: "nice", CreatedAt: "2026-01-03T00:00:00Z"}
		require.NoError(t, s.AddComment(ctx, c))
		list, err := s.ListComments(ctx, "t1")
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, "nice", list[0].Body)

		got, err := s.GetTree(ctx, "t1")
		require.NoError(t, err)
		assert.Equal(t, 1, got.CommentCount)

		require.NoError(t, s.DeleteComment(ctx, "t1", "c1"))
		list, err = s.ListComments(ctx, "t1")
		require.NoError(t, err)
		assert.Empty(t, list)
		assert.ErrorIs(t, s.DeleteComment(ctx, "t1", "c1"), ErrNotFound)
	})

	t.Run("profiles", func(t *testing.T) {
		_, err := s.GetProfile(ctx, "nobody")
		assert.ErrorIs(t, err, ErrNotFound)
		require.NoError(t, s.PutProfile(ctx, Profile{UserID: "u1", DisplayName: "alice", Bio: "hi"}))
		p, err := s.GetProfile(ctx, "u1")
		require.NoError(t, err)
		assert.Equal(t, "alice", p.DisplayName)
	})

	t.Run("forum", func(t *testing.T) {
		th := Thread{ID: "th1", Title: "Hello", AuthorID: "u1", Author: "alice", Body: "first", CreatedAt: "2026-01-01T00:00:00Z", LastActivity: "2026-01-01T00:00:00Z"}
		require.NoError(t, s.PutThread(ctx, th))
		require.NoError(t, s.AddPost(ctx, Post{ID: "p1", ThreadID: "th1", AuthorID: "u2", Author: "bob", Body: "reply", CreatedAt: "2026-01-02T00:00:00Z"}))

		got, err := s.GetThread(ctx, "th1")
		require.NoError(t, err)
		assert.Equal(t, 1, got.PostCount)
		assert.Equal(t, "2026-01-02T00:00:00Z", got.LastActivity)

		posts, err := s.ListPosts(ctx, "th1")
		require.NoError(t, err)
		assert.Len(t, posts, 1)

		threads, err := s.ListThreads(ctx, 10)
		require.NoError(t, err)
		require.Len(t, threads, 1)

		assert.ErrorIs(t, s.AddPost(ctx, Post{ID: "p2", ThreadID: "nope"}), ErrNotFound)
	})

	t.Run("scenes", func(t *testing.T) {
		sc := Scene{ID: "s1", OwnerID: "u1", Name: "arena", Visibility: VisibilityPrivate, CreatedAt: "2026-01-01T00:00:00Z",
			Agents: []SceneAgent{{Name: "a", TreeID: "t1", Model: "character-a", X: 1, Z: 2}}}
		require.NoError(t, s.PutScene(ctx, sc))
		got, err := s.GetScene(ctx, "s1")
		require.NoError(t, err)
		require.Len(t, got.Agents, 1)
		assert.InDelta(t, 2.0, got.Agents[0].Z, 0.001)
		list, err := s.ListScenesByOwner(ctx, "u1")
		require.NoError(t, err)
		assert.Len(t, list, 1)
		require.NoError(t, s.DeleteScene(ctx, "s1"))
		assert.ErrorIs(t, s.DeleteScene(ctx, "s1"), ErrNotFound)
	})

	t.Run("stats", func(t *testing.T) {
		st, err := s.Stats(ctx, "u1")
		require.NoError(t, err)
		assert.Equal(t, 2, st.Trees)
		assert.Equal(t, 1, st.PublicTrees)
		assert.Equal(t, 1, st.Likes)
		assert.Equal(t, 1, st.ForumThreads)
	})

	t.Run("delete tree", func(t *testing.T) {
		require.NoError(t, s.DeleteTree(ctx, "t2"))
		_, err := s.GetTree(ctx, "t2")
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestMemStoreContract(t *testing.T) {
	runStoreContract(t, NewMemStore())
}
