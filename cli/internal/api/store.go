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

	PutScene(ctx context.Context, s Scene) error
	GetScene(ctx context.Context, id string) (Scene, error)
	DeleteScene(ctx context.Context, id string) error
	ListScenesByOwner(ctx context.Context, ownerID string) ([]Scene, error)

	// Stats aggregates engagement for one user.
	Stats(ctx context.Context, userID string) (Stats, error)
}
