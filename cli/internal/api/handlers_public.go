package api

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	publicTTL      = 24 * time.Hour
	publicMaxItems = 48
)

// publicCache holds the visitor showcase: a once-a-day snapshot of public trees. Visitors
// without an account see this; signed-in users read the live /v1/trees. It is the only
// unauthenticated data route, it carries no tree specs, and it never queries the store
// more than once per TTL however many visitors arrive.
type publicCache struct {
	mu    sync.Mutex
	at    time.Time
	trees []Tree
}

func (s *Server) publicSnapshot(r *http.Request) ([]Tree, time.Time, error) {
	ttl := s.PublicTTL
	if ttl <= 0 {
		ttl = publicTTL
	}
	s.public.mu.Lock()
	defer s.public.mu.Unlock()
	if !s.public.at.IsZero() && s.Now().Sub(s.public.at) < ttl {
		return s.public.trees, s.public.at, nil
	}
	ts, err := s.Store.ListPublicTrees(r.Context(), 200)
	if err != nil {
		if !s.public.at.IsZero() {
			return s.public.trees, s.public.at, nil // serve the stale snapshot rather than fail
		}
		return nil, time.Time{}, err
	}
	s.public.trees, s.public.at = summarize(ts), s.Now()
	return s.public.trees, s.public.at, nil
}

func (s *Server) publicTrees(w http.ResponseWriter, r *http.Request) {
	trees, at, err := s.publicSnapshot(r)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	q := strings.ToLower(r.URL.Query().Get("q"))
	var out []Tree
	for _, t := range trees {
		if q != "" && !strings.Contains(strings.ToLower(t.Name+" "+t.Description+" "+strings.Join(t.Tags, " ")), q) {
			continue
		}
		out = append(out, t)
	}
	if r.URL.Query().Get("sort") == "popular" {
		sort.SliceStable(out, func(i, j int) bool { return out[i].LikeCount > out[j].LikeCount })
	}
	if len(out) > publicMaxItems {
		out = out[:publicMaxItems]
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	writeJSON(w, http.StatusOK, map[string]any{
		"items":       listOf(out)["items"],
		"generatedAt": at.UTC().Format(time.RFC3339),
		"note":        "A snapshot refreshed daily. Sign in for live results, likes, comments and forking.",
	})
}
