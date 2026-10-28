package api

import (
	"context"
	"sort"
	"sync"
)

// MemStore is an in-memory Store for tests and local development.
type MemStore struct {
	mu       sync.Mutex
	trees    map[string]Tree
	likes    map[string]map[string]bool
	comments map[string][]Comment
	profiles map[string]Profile
	threads  map[string]Thread
	posts    map[string][]Post
	scenes   map[string]Scene
}

// NewMemStore returns an empty MemStore.
func NewMemStore() *MemStore {
	return &MemStore{
		trees:    map[string]Tree{},
		likes:    map[string]map[string]bool{},
		comments: map[string][]Comment{},
		profiles: map[string]Profile{},
		threads:  map[string]Thread{},
		posts:    map[string][]Post{},
		scenes:   map[string]Scene{},
	}
}

func (m *MemStore) PutTree(_ context.Context, t Tree) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if old, ok := m.trees[t.ID]; ok {
		t.LikeCount, t.CommentCount = old.LikeCount, old.CommentCount
	}
	m.trees[t.ID] = t
	return nil
}

func (m *MemStore) GetTree(_ context.Context, id string) (Tree, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.trees[id]
	if !ok {
		return Tree{}, ErrNotFound
	}
	return t, nil
}

func (m *MemStore) DeleteTree(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.trees[id]; !ok {
		return ErrNotFound
	}
	delete(m.trees, id)
	delete(m.likes, id)
	delete(m.comments, id)
	return nil
}

func (m *MemStore) filterTrees(keep func(Tree) bool, limit int) []Tree {
	var out []Tree
	for _, t := range m.trees {
		if keep(t) {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func (m *MemStore) ListPublicTrees(_ context.Context, limit int) ([]Tree, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.filterTrees(func(t Tree) bool { return t.Visibility == VisibilityPublic }, limit), nil
}

func (m *MemStore) ListTreesByOwner(_ context.Context, ownerID string) ([]Tree, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.filterTrees(func(t Tree) bool { return t.OwnerID == ownerID }, 0), nil
}

func (m *MemStore) SetLike(_ context.Context, treeID, userID string, liked bool) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.trees[treeID]
	if !ok {
		return 0, ErrNotFound
	}
	if m.likes[treeID] == nil {
		m.likes[treeID] = map[string]bool{}
	}
	if liked {
		m.likes[treeID][userID] = true
	} else {
		delete(m.likes[treeID], userID)
	}
	t.LikeCount = len(m.likes[treeID])
	m.trees[treeID] = t
	return t.LikeCount, nil
}

func (m *MemStore) HasLiked(_ context.Context, treeID, userID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.likes[treeID][userID], nil
}

func (m *MemStore) AddComment(_ context.Context, c Comment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.trees[c.TreeID]
	if !ok {
		return ErrNotFound
	}
	m.comments[c.TreeID] = append(m.comments[c.TreeID], c)
	t.CommentCount = len(m.comments[c.TreeID])
	m.trees[c.TreeID] = t
	return nil
}

func (m *MemStore) ListComments(_ context.Context, treeID string) ([]Comment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Comment(nil), m.comments[treeID]...), nil
}

func (m *MemStore) GetComment(_ context.Context, treeID, commentID string) (Comment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.comments[treeID] {
		if c.ID == commentID {
			return c, nil
		}
	}
	return Comment{}, ErrNotFound
}

func (m *MemStore) DeleteComment(_ context.Context, treeID, commentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	list := m.comments[treeID]
	for i, c := range list {
		if c.ID == commentID {
			m.comments[treeID] = append(list[:i:i], list[i+1:]...)
			t := m.trees[treeID]
			t.CommentCount = len(m.comments[treeID])
			m.trees[treeID] = t
			return nil
		}
	}
	return ErrNotFound
}

func (m *MemStore) PutProfile(_ context.Context, p Profile) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.profiles[p.UserID] = p
	return nil
}

func (m *MemStore) GetProfile(_ context.Context, userID string) (Profile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.profiles[userID]
	if !ok {
		return Profile{}, ErrNotFound
	}
	return p, nil
}

func (m *MemStore) PutThread(_ context.Context, t Thread) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if old, ok := m.threads[t.ID]; ok {
		t.PostCount = old.PostCount
	}
	m.threads[t.ID] = t
	return nil
}

func (m *MemStore) GetThread(_ context.Context, id string) (Thread, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.threads[id]
	if !ok {
		return Thread{}, ErrNotFound
	}
	return t, nil
}

func (m *MemStore) ListThreads(_ context.Context, limit int) ([]Thread, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Thread, 0, len(m.threads))
	for _, t := range m.threads {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastActivity > out[j].LastActivity })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *MemStore) AddPost(_ context.Context, p Post) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.threads[p.ThreadID]
	if !ok {
		return ErrNotFound
	}
	m.posts[p.ThreadID] = append(m.posts[p.ThreadID], p)
	t.PostCount = len(m.posts[p.ThreadID])
	t.LastActivity = p.CreatedAt
	m.threads[p.ThreadID] = t
	return nil
}

func (m *MemStore) ListPosts(_ context.Context, threadID string) ([]Post, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Post(nil), m.posts[threadID]...), nil
}

func (m *MemStore) PutScene(_ context.Context, s Scene) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scenes[s.ID] = s
	return nil
}

func (m *MemStore) GetScene(_ context.Context, id string) (Scene, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.scenes[id]
	if !ok {
		return Scene{}, ErrNotFound
	}
	return s, nil
}

func (m *MemStore) DeleteScene(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.scenes[id]; !ok {
		return ErrNotFound
	}
	delete(m.scenes, id)
	return nil
}

func (m *MemStore) ListScenesByOwner(_ context.Context, ownerID string) ([]Scene, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Scene
	for _, s := range m.scenes {
		if s.OwnerID == ownerID {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	return out, nil
}

func (m *MemStore) Stats(_ context.Context, userID string) (Stats, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var s Stats
	for _, t := range m.trees {
		if t.OwnerID != userID {
			continue
		}
		s.Trees++
		if t.Visibility == VisibilityPublic {
			s.PublicTrees++
		}
		s.Likes += t.LikeCount
		s.Comments += t.CommentCount
	}
	for _, t := range m.threads {
		if t.AuthorID == userID {
			s.ForumThreads++
		}
	}
	for _, ps := range m.posts {
		for _, p := range ps {
			if p.AuthorID == userID {
				s.ForumPosts++
			}
		}
	}
	for _, sc := range m.scenes {
		if sc.OwnerID == userID {
			s.ScenesCreated++
		}
	}
	return s, nil
}
