package registry

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTreeEntry_FullName(t *testing.T) {
	e := TreeEntry{Owner: "alice", Name: "patrol-ai"}
	assert.Equal(t, "alice/patrol-ai", e.FullName())
}

func TestLocalClient_LoginLogout(t *testing.T) {
	tmpDir := t.TempDir()
	client := NewLocalClient(filepath.Join(tmpDir, "registry"))

	ctx := context.Background()
	assert.False(t, client.IsAuthenticated())

	require.NoError(t, client.Login(ctx, "gh_test_token_123"))
	assert.True(t, client.IsAuthenticated())

	require.NoError(t, client.Logout(ctx))
	assert.False(t, client.IsAuthenticated())
}

func TestLocalClient_PushAndPull(t *testing.T) {
	tmpDir := t.TempDir()
	client := NewLocalClient(filepath.Join(tmpDir, "registry"))
	ctx := context.Background()
	require.NoError(t, client.Login(ctx, "token"))

	specData := []byte(`version: "1.0"
metadata:
  name: patrol-ai
tree:
  type: action
  name: patrol
  node: Patrol
`)

	entry, err := client.Push(ctx, specData, PushOptions{
		Public:      true,
		Description: "Simple patrol behavior",
		Tags:        []string{"patrol", "ai"},
	})
	require.NoError(t, err)
	assert.Equal(t, "patrol-ai", entry.Name)
	assert.True(t, entry.Public)

	pulled, err := client.Pull(ctx, entry.FullName())
	require.NoError(t, err)
	assert.Equal(t, specData, pulled)
}

func TestLocalClient_Browse(t *testing.T) {
	tmpDir := t.TempDir()
	client := NewLocalClient(filepath.Join(tmpDir, "registry"))
	ctx := context.Background()
	require.NoError(t, client.Login(ctx, "token"))

	for _, name := range []string{"patrol-ai", "combat-ai", "stealth-ai"} {
		spec := []byte(`version: "1.0"
metadata:
  name: ` + name + `
tree:
  type: action
  name: do_it
  node: DoIt
`)
		_, err := client.Push(ctx, spec, PushOptions{Public: true, Tags: []string{"ai"}})
		require.NoError(t, err)
	}

	entries, err := client.Browse(ctx, BrowseOptions{Limit: 10})
	require.NoError(t, err)
	assert.Len(t, entries, 3)
}

func TestLocalClient_BrowseByTag(t *testing.T) {
	tmpDir := t.TempDir()
	client := NewLocalClient(filepath.Join(tmpDir, "registry"))
	ctx := context.Background()
	require.NoError(t, client.Login(ctx, "token"))

	specs := []struct {
		name string
		tags []string
	}{
		{"patrol-ai", []string{"patrol", "ai"}},
		{"combat-ai", []string{"combat", "ai"}},
		{"stealth-ai", []string{"stealth"}},
	}
	for _, s := range specs {
		data := []byte(`version: "1.0"
metadata:
  name: ` + s.name + `
tree:
  type: action
  name: do_it
  node: DoIt
`)
		_, err := client.Push(ctx, data, PushOptions{Public: true, Tags: s.tags})
		require.NoError(t, err)
	}

	entries, err := client.Browse(ctx, BrowseOptions{Tag: "ai", Limit: 10})
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

func TestLocalClient_Search(t *testing.T) {
	tmpDir := t.TempDir()
	client := NewLocalClient(filepath.Join(tmpDir, "registry"))
	ctx := context.Background()
	require.NoError(t, client.Login(ctx, "token"))

	for _, name := range []string{"patrol-ai", "combat-ai", "patrol-stealth"} {
		spec := []byte(`version: "1.0"
metadata:
  name: ` + name + `
tree:
  type: action
  name: do_it
  node: DoIt
`)
		_, err := client.Push(ctx, spec, PushOptions{Public: true})
		require.NoError(t, err)
	}

	entries, err := client.Search(ctx, SearchOptions{Query: "patrol", Limit: 10})
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

func TestLocalClient_PullNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	client := NewLocalClient(filepath.Join(tmpDir, "registry"))
	ctx := context.Background()

	_, err := client.Pull(ctx, "nobody/nonexistent")
	assert.Error(t, err)
}

func TestLocalClient_PushRequiresAuth(t *testing.T) {
	tmpDir := t.TempDir()
	client := NewLocalClient(filepath.Join(tmpDir, "registry"))
	ctx := context.Background()

	_, err := client.Push(ctx, []byte("data"), PushOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not authenticated")
}

func TestLocalClient_PersistsIndex(t *testing.T) {
	tmpDir := t.TempDir()
	regDir := filepath.Join(tmpDir, "registry")
	client := NewLocalClient(regDir)
	ctx := context.Background()
	require.NoError(t, client.Login(ctx, "token"))

	spec := []byte(`version: "1.0"
metadata:
  name: persist-test
tree:
  type: action
  name: do_it
  node: DoIt
`)
	_, err := client.Push(ctx, spec, PushOptions{Public: true})
	require.NoError(t, err)

	// Create a new client instance — should load from disk
	client2 := NewLocalClient(regDir)
	entries, err := client2.Browse(ctx, BrowseOptions{Limit: 10})
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "persist-test", entries[0].Name)

	// Verify the spec file exists on disk
	_, err = os.Stat(filepath.Join(regDir, "trees", client.token, "persist-test.beetree.yaml"))
	assert.NoError(t, err)
}
