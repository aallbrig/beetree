package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// LocalClient implements Client using the local filesystem.
// Trees are stored in a directory with a JSON index for metadata.
type LocalClient struct {
	dir   string
	token string
	index *registryIndex
}

type registryIndex struct {
	Entries []TreeEntry `json:"entries"`
}

// specMetadata extracts the name from a .beetree.yaml spec.
type specMetadata struct {
	Metadata struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	} `yaml:"metadata"`
}

// NewLocalClient creates a file-backed registry client rooted at dir.
func NewLocalClient(dir string) *LocalClient {
	c := &LocalClient{dir: dir}
	c.loadIndex()
	c.loadToken()
	return c
}

func (c *LocalClient) indexPath() string {
	return filepath.Join(c.dir, "index.json")
}

func (c *LocalClient) tokenPath() string {
	return filepath.Join(c.dir, ".token")
}

func (c *LocalClient) loadIndex() {
	c.index = &registryIndex{}
	data, err := os.ReadFile(c.indexPath())
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, c.index)
}

func (c *LocalClient) saveIndex() error {
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c.index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.indexPath(), data, 0644)
}

func (c *LocalClient) loadToken() {
	data, err := os.ReadFile(c.tokenPath())
	if err != nil {
		return
	}
	c.token = strings.TrimSpace(string(data))
}

func (c *LocalClient) Login(_ context.Context, token string) error {
	c.token = token
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(c.tokenPath(), []byte(token), 0600)
}

func (c *LocalClient) Logout(_ context.Context) error {
	c.token = ""
	_ = os.Remove(c.tokenPath())
	return nil
}

func (c *LocalClient) IsAuthenticated() bool {
	return c.token != ""
}

func (c *LocalClient) Browse(_ context.Context, opts BrowseOptions) ([]TreeEntry, error) {
	var results []TreeEntry
	for _, e := range c.index.Entries {
		if !e.Public {
			continue
		}
		if opts.Tag != "" && !containsTag(e.Tags, opts.Tag) {
			continue
		}
		results = append(results, e)
	}

	switch opts.Sort {
	case "popular":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Downloads > results[j].Downloads
		})
	case "name":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
	default: // "recent" or empty
		sort.Slice(results, func(i, j int) bool {
			return results[i].UpdatedAt.After(results[j].UpdatedAt)
		})
	}

	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}
	return results, nil
}

func (c *LocalClient) Search(_ context.Context, opts SearchOptions) ([]TreeEntry, error) {
	query := strings.ToLower(opts.Query)
	var results []TreeEntry
	for _, e := range c.index.Entries {
		if !e.Public {
			continue
		}
		if opts.Tag != "" && !containsTag(e.Tags, opts.Tag) {
			continue
		}
		nameMatch := strings.Contains(strings.ToLower(e.Name), query)
		descMatch := strings.Contains(strings.ToLower(e.Description), query)
		tagMatch := false
		for _, tag := range e.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				tagMatch = true
				break
			}
		}
		if nameMatch || descMatch || tagMatch {
			results = append(results, e)
		}
	}
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}
	return results, nil
}

func (c *LocalClient) Pull(_ context.Context, fullName string) ([]byte, error) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid name format, expected owner/name")
	}
	owner, name := parts[0], parts[1]
	specPath := filepath.Join(c.dir, "trees", owner, name+".beetree.yaml")
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("tree %q not found", fullName)
	}
	return data, nil
}

func (c *LocalClient) Push(_ context.Context, specData []byte, opts PushOptions) (*TreeEntry, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated — run 'beetree registry login' first")
	}

	var meta specMetadata
	if err := yaml.Unmarshal(specData, &meta); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}
	if meta.Metadata.Name == "" {
		return nil, fmt.Errorf("spec must have metadata.name")
	}

	owner := c.token
	name := meta.Metadata.Name
	now := time.Now()

	// Save spec file
	treeDir := filepath.Join(c.dir, "trees", owner)
	if err := os.MkdirAll(treeDir, 0755); err != nil {
		return nil, err
	}
	specPath := filepath.Join(treeDir, name+".beetree.yaml")
	if err := os.WriteFile(specPath, specData, 0644); err != nil {
		return nil, err
	}

	// Update or create index entry
	entry := c.findEntry(owner, name)
	if entry != nil {
		entry.Description = opts.Description
		entry.Tags = opts.Tags
		entry.Public = opts.Public
		entry.UpdatedAt = now
	} else {
		newEntry := TreeEntry{
			Owner:       owner,
			Name:        name,
			Description: opts.Description,
			Tags:        opts.Tags,
			Public:      opts.Public,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		c.index.Entries = append(c.index.Entries, newEntry)
		entry = &c.index.Entries[len(c.index.Entries)-1]
	}

	if err := c.saveIndex(); err != nil {
		return nil, err
	}
	return entry, nil
}

func (c *LocalClient) findEntry(owner, name string) *TreeEntry {
	for i := range c.index.Entries {
		if c.index.Entries[i].Owner == owner && c.index.Entries[i].Name == name {
			return &c.index.Entries[i]
		}
	}
	return nil
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}
