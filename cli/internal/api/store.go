package api

import "context"

// Store is the persistence boundary. Implementations: MemStore (tests, local)
// and DynamoStore (production).
type Store interface {
	PutTree(ctx context.Context, t Tree) error
	GetTree(ctx context.Context, id string) (Tree, error)
	DeleteTree(ctx context.Context, id string) error
	ListPublicTrees(ctx context.Context, limit int) ([]Tree, error)
	ListTreesByOwner(ctx context.Context, ownerID string) ([]Tree, error)

	// SetLike records or removes a like and returns the new like count.
	// It is idempotent: liking twice counts once.
	SetLike(ctx context.Context, treeID, userID string, liked bool) (int, error)
	HasLiked(ctx context.Context, treeID, userID string) (bool, error)

	AddComment(ctx context.Context, c Comment) error
	ListComments(ctx context.Context, treeID string) ([]Comment, error)
	DeleteComment(ctx context.Context, treeID, commentID string) error
	GetComment(ctx context.Context, treeID, commentID string) (Comment, error)

	PutProfile(ctx context.Context, p Profile) error
	GetProfile(ctx context.Context, userID string) (Profile, error)

	PutThread(ctx context.Context, t Thread) error
	GetThread(ctx context.Context, id string) (Thread, error)
	ListThreads(ctx context.Context, limit int) ([]Thread, error)
	AddPost(ctx context.Context, p Post) error
	ListPosts(ctx context.Context, threadID string) ([]Post, error)
	// AddReport files a report. It is idempotent per reporter and target:
	// the second report of the same thread or post by the same user is a no-op.
	// ErrNotFound when the thread does not exist.
	AddReport(ctx context.Context, r Report) error
	ListReports(ctx context.Context, threadID string) ([]Report, error)
	// ListRecentReports lists reports across all threads, newest first.
	ListRecentReports(ctx context.Context, limit int) ([]Report, error)

	PutScene(ctx context.Context, s Scene) error
	GetScene(ctx context.Context, id string) (Scene, error)
	DeleteScene(ctx context.Context, id string) error
	ListScenesByOwner(ctx context.Context, ownerID string) ([]Scene, error)

	// API keys are stored by hash; the secret is never persisted.
	PutAPIKey(ctx context.Context, k APIKey) error
	GetAPIKeyByHash(ctx context.Context, hash string) (APIKey, error)
	ListAPIKeys(ctx context.Context, ownerID string) ([]APIKey, error)
	// DeleteAPIKey removes the owner's key with this id; ErrNotFound otherwise.
	DeleteAPIKey(ctx context.Context, ownerID, id string) error

	// Stats aggregates engagement for one user.
	Stats(ctx context.Context, userID string) (Stats, error)
}
