package api

import (
	"net/http"
	"time"
)

// RemixLimit is the number of remixes one user may make per RemixWindow.
const (
	RemixLimit  = 20
	RemixWindow = 24 * time.Hour
)

const maxAncestors = 20

// withLineageState returns a copy of l with State derived from the live
// original. The frozen title and owner name are untouched.
func (s *Server) withLineageState(r *http.Request, l *Lineage) *Lineage {
	if l == nil {
		return nil
	}
	out := *l
	orig, err := s.Store.GetTree(r.Context(), l.ID)
	switch {
	case err != nil:
		out.State = LineageDeleted
	case orig.Visibility == VisibilityPublic:
		out.State = LineagePublic
	default:
		out.State = LineagePrivate
	}
	return &out
}

// remixError is a failed remix with the HTTP status and code the REST route reports.
type remixError struct {
	status int
	code   string
	msg    string
	wait   int
}

func (e *remixError) Error() string { return e.msg }

// remixTree copies a public tree into the caller's account as a private tree
// that records where it came from.
func (s *Server) remixTree(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	t, err := s.remixFor(r, c, r.PathValue("id"))
	if err != nil {
		if re, ok := err.(*remixError); ok {
			if re.wait > 0 {
				w.Header().Set("Retry-After", retryAfter(re.wait))
			}
			writeErr(w, re.status, re.code, re.msg)
			return
		}
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

// remixFor makes the copy for the REST route and the MCP remix_tree tool.
func (s *Server) remixFor(r *http.Request, c Claims, id string) (Tree, error) {
	orig, err := s.Store.GetTree(r.Context(), id)
	if err != nil || orig.Visibility != VisibilityPublic {
		return Tree{}, &remixError{status: http.StatusNotFound, code: "not_found", msg: "not found"}
	}
	if ok, wait := s.remixLimiter.allow(c.Sub, s.Now(), RemixLimit, RemixWindow); !ok {
		return Tree{}, &remixError{status: http.StatusTooManyRequests, code: "rate_limited", msg: "remix limit reached; try again later", wait: wait}
	}
	now := s.stamp()
	ref := LineageRef{Kind: "tree", ID: orig.ID, Title: orig.Name, OwnerID: orig.OwnerID, OwnerName: orig.OwnerName, At: now}
	var ancestors []LineageRef
	if orig.RemixedFrom != nil {
		ancestors = append(ancestors, orig.RemixedFrom.Ancestors...)
		ancestors = append(ancestors, orig.RemixedFrom.LineageRef)
		if len(ancestors) > maxAncestors {
			ancestors = ancestors[len(ancestors)-maxAncestors:]
		}
	}
	// Name is suffixed, then clipped, so a long title cannot push the copy past the limit.
	name := clip(orig.Name, 80-len(" (Remix)")) + " (Remix)"
	t := Tree{
		ID: s.NewID(), OwnerID: c.Sub, OwnerName: s.ownerName(r, c),
		Name: name, Description: orig.Description, Visibility: VisibilityPrivate,
		Tags: orig.Tags, Spec: orig.Spec, CreatedAt: now, UpdatedAt: now,
		RemixedFrom: &Lineage{LineageRef: ref, Ancestors: ancestors},
	}
	if err := s.Store.PutTree(r.Context(), t); err != nil {
		return Tree{}, err
	}
	t.RemixedFrom = s.withLineageState(r, t.RemixedFrom)
	return t, nil
}
