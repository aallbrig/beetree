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
	RemixedFrom  *Lineage `json:"remixedFrom,omitempty"`
	// UnchangedCopy is derived on single reads: a public remix whose spec still matches its original.
	UnchangedCopy bool `json:"unchangedCopy,omitempty"`
}

// LineageRef is a frozen record of an original at the moment it was remixed.
// It is never refreshed from the live original, so credit survives the
// original going private or being deleted.
type LineageRef struct {
	Kind      string `json:"kind"`
	ID        string `json:"id"`
	Title     string `json:"title"`
	OwnerID   string `json:"ownerId"`
	OwnerName string `json:"ownerName"`
	At        string `json:"at"`
}

// Lineage is the remix record on a copy. State is derived on read from the
// live original (public, private or deleted) and is never stored.
type Lineage struct {
	LineageRef
	State     string       `json:"state,omitempty"`
	Ancestors []LineageRef `json:"ancestors,omitempty"`
}

// Lineage states.
const (
	LineagePublic  = "public"
	LineagePrivate = "private"
	LineageDeleted = "deleted"
)

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
	Moderator   bool   `json:"moderator,omitempty"`
	// Avatar is the id of one of the 100 preset pictures (a001 to a100); empty means the default.
	Avatar string `json:"avatar,omitempty"`
	// MarketingOptIn records consent to receive product news and campaign email. It is
	// private: shown to the owner only, kept so the list can move to a marketing platform.
	MarketingOptIn  bool   `json:"marketingOptIn"`
	MarketingAt     string `json:"marketingOptInAt,omitempty"`
	MarketingSource string `json:"marketingSource,omitempty"`
	CreatedAt       string `json:"createdAt"`
}

// PublicProfile is the profile as other people may see it: no email, no consent record.
type PublicProfile struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Bio         string `json:"bio"`
	Avatar      string `json:"avatar,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

// Public returns the fields safe to show to anyone.
func (p Profile) Public() PublicProfile {
	return PublicProfile{p.UserID, p.DisplayName, p.Bio, p.Avatar, p.CreatedAt}
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

// Report flags a thread (PostID empty) or one of its posts for review.
type Report struct {
	ID       string `json:"id"`
	ThreadID string `json:"threadId"`
	PostID   string `json:"postId,omitempty"`
	// Kind is "tree" when ThreadID holds a tree id; empty for forum reports.
	Kind       string `json:"kind,omitempty"`
	ReporterID string `json:"reporterId"`
	Reason     string `json:"reason"`
	CreatedAt  string `json:"createdAt"`
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
	// RemixedFrom is set on copies made by the remix route.
	RemixedFrom *Lineage `json:"remixedFrom,omitempty"`
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
