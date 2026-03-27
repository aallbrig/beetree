// Package registry provides a client interface for the BeeTree registry,
// enabling browsing, searching, pulling, and pushing behavior tree specs.
package registry

import (
	"context"
	"time"
)

// TreeEntry represents a published behavior tree in the registry.
type TreeEntry struct {
	Owner       string    `json:"owner" yaml:"owner"`
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description" yaml:"description"`
	Tags        []string  `json:"tags" yaml:"tags"`
	Version     string    `json:"version" yaml:"version"`
	Downloads   int       `json:"downloads" yaml:"downloads"`
	Stars       int       `json:"stars" yaml:"stars"`
	Public      bool      `json:"public" yaml:"public"`
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" yaml:"updated_at"`
}

// FullName returns the owner/name identifier.
func (e TreeEntry) FullName() string {
	return e.Owner + "/" + e.Name
}

// BrowseOptions controls listing behavior.
type BrowseOptions struct {
	Tag    string
	Sort   string // "popular", "recent", "name"
	Limit  int
	Offset int
}

// SearchOptions controls search behavior.
type SearchOptions struct {
	Query  string
	Tag    string
	Limit  int
	Offset int
}

// PushOptions controls publish behavior.
type PushOptions struct {
	Public      bool
	Description string
	Tags        []string
}

// Client defines the registry API contract.
type Client interface {
	// Login authenticates with the registry and stores credentials.
	Login(ctx context.Context, token string) error

	// Logout removes stored credentials.
	Logout(ctx context.Context) error

	// IsAuthenticated checks if valid credentials exist.
	IsAuthenticated() bool

	// Browse lists trees with optional filters.
	Browse(ctx context.Context, opts BrowseOptions) ([]TreeEntry, error)

	// Search finds trees matching a query.
	Search(ctx context.Context, opts SearchOptions) ([]TreeEntry, error)

	// Pull downloads a tree spec by owner/name.
	Pull(ctx context.Context, fullName string) ([]byte, error)

	// Push publishes a tree spec to the registry.
	Push(ctx context.Context, specData []byte, opts PushOptions) (*TreeEntry, error)
}
