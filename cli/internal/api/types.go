// Package api implements the BeeTree HTTP API: REST endpoints, the MCP server,
// and the storage and auth abstractions they share. It is transport-agnostic
// (net/http) so the same handler runs under Lambda and in tests.
package api

import "errors"

// Visibility controls who can read a tree.
const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
)

// Sentinel errors returned by Store implementations.
var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

// Tree is a stored behavior tree. Spec holds the .beetree.yaml document.
type Tree struct {
	ID           string   `json:"id"`
	OwnerID      string   `json:"ownerId"`
	OwnerName    string   `json:"ownerName"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Visibility   string   `json:"visibility"`
	Tags         []string `json:"tags"`
	Spec         string   `json:"spec,omitempty"`
	LikeCount    int      `json:"likeCount"`
	CommentCount int      `json:"commentCount"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
	LikedByMe    bool     `json:"likedByMe,omitempty"`
}

// Comment is a comment on a tree.
type Comment struct {
	ID        string `json:"id"`
	TreeID    string `json:"treeId"`
	AuthorID  string `json:"authorId"`
	Author    string `json:"author"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}

// Profile is a user's public profile.
type Profile struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Bio         string `json:"bio"`
	Email       string `json:"email,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

// Thread is a forum thread.
type Thread struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	AuthorID     string `json:"authorId"`
	Author       string `json:"author"`
	Body         string `json:"body"`
	PostCount    int    `json:"postCount"`
	CreatedAt    string `json:"createdAt"`
	LastActivity string `json:"lastActivity"`
}

// Post is a reply in a forum thread.
type Post struct {
	ID        string `json:"id"`
	ThreadID  string `json:"threadId"`
	AuthorID  string `json:"authorId"`
	Author    string `json:"author"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}

// SceneAgent places one tree-driven agent in a 3D scene.
type SceneAgent struct {
	Name   string  `json:"name"`
	TreeID string  `json:"treeId,omitempty"`
	Model  string  `json:"model"`
	X      float64 `json:"x"`
	Z      float64 `json:"z"`
}

// Scene is a saved arrangement of agents in the 3D world.
type Scene struct {
	ID         string       `json:"id"`
	OwnerID    string       `json:"ownerId"`
	Name       string       `json:"name"`
	Agents     []SceneAgent `json:"agents"`
	Visibility string       `json:"visibility"`
	CreatedAt  string       `json:"createdAt"`
	UpdatedAt  string       `json:"updatedAt"`
}

// Stats is the engagement summary shown on a user's dashboard.
type Stats struct {
	Trees         int `json:"trees"`
	PublicTrees   int `json:"publicTrees"`
	Likes         int `json:"likes"`
	Comments      int `json:"comments"`
	ForumThreads  int `json:"forumThreads"`
	ForumPosts    int `json:"forumPosts"`
	ScenesCreated int `json:"scenes"`
}
