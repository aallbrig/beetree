package api

import (
	"net/http"
	"reflect"
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
	var vis string
	var err error
	if l.Kind == "scene" {
		var sc Scene
		sc, err = s.Store.GetScene(r.Context(), l.ID)
		vis = sc.Visibility
	} else {
		var orig Tree
		orig, err = s.Store.GetTree(r.Context(), l.ID)
		vis = orig.Visibility
	}
	switch {
	case err != nil:
		out.State = LineageDeleted
	case vis == VisibilityPublic:
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
	t := s.copyTree(orig, c.Sub, s.ownerName(r, c), s.stamp())
	if err := s.Store.PutTree(r.Context(), t); err != nil {
		return Tree{}, err
	}
	t.RemixedFrom = s.withLineageState(r, t.RemixedFrom)
	return t, nil
}

// ancestorsOf returns the chain to record on a copy of an item whose own
// lineage is prev, capped at maxAncestors.
func ancestorsOf(prev *Lineage) []LineageRef {
	if prev == nil {
		return nil
	}
	out := append(append([]LineageRef{}, prev.Ancestors...), prev.LineageRef)
	if len(out) > maxAncestors {
		out = out[len(out)-maxAncestors:]
	}
	return out
}

// copyTree builds the private copy of orig for a new owner.
func (s *Server) copyTree(orig Tree, ownerID, ownerName, now string) Tree {
	ref := LineageRef{Kind: "tree", ID: orig.ID, Title: orig.Name, OwnerID: orig.OwnerID, OwnerName: orig.OwnerName, At: now}
	// Name is suffixed, then clipped, so a long title cannot push the copy past the limit.
	return Tree{
		ID: s.NewID(), OwnerID: ownerID, OwnerName: ownerName,
		Name: clip(orig.Name, 80-len(" (Remix)")) + " (Remix)", Description: orig.Description,
		Visibility: VisibilityPrivate, Tags: orig.Tags, Spec: orig.Spec, CreatedAt: now, UpdatedAt: now,
		RemixedFrom: &Lineage{LineageRef: ref, Ancestors: ancestorsOf(orig.RemixedFrom)},
	}
}

const unchangedRemixMsg = "This remix is identical to the original. Change the tree before sharing it."

// unchangedRemix reports whether t is a remix whose spec matches its live
// original. Specs are compared parsed, so whitespace and comment edits do not
// count as a change. A missing original (private or deleted) cannot be compared.
func (s *Server) unchangedRemix(r *http.Request, t Tree) bool {
	if t.RemixedFrom == nil {
		return false
	}
	orig, err := s.Store.GetTree(r.Context(), t.RemixedFrom.ID)
	if err != nil {
		return false
	}
	a, errA := validSpec(orig.Spec)
	b, errB := validSpec(t.Spec)
	return errA == nil && errB == nil && reflect.DeepEqual(a, b)
}

// sceneRemix is a scene copy with the trees that were copied alongside it.
type sceneRemix struct {
	Scene
	Trees []Tree `json:"trees"`
}

// remixScene copies a public scene and the public trees its agents use into the
// caller's account as private items with lineage. One call counts as one remix.
func (s *Server) remixScene(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	orig, err := s.Store.GetScene(r.Context(), r.PathValue("id"))
	if err != nil || orig.Visibility != VisibilityPublic {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	if ok, wait := s.remixLimiter.allow(c.Sub, s.Now(), RemixLimit, RemixWindow); !ok {
		w.Header().Set("Retry-After", retryAfter(wait))
		writeErr(w, http.StatusTooManyRequests, "rate_limited", "remix limit reached; try again later")
		return
	}
	now := s.stamp()
	owner := s.ownerName(r, c)
	ownerName := ""
	if p, err := s.Store.GetProfile(r.Context(), orig.OwnerID); err == nil {
		ownerName = p.DisplayName
	}
	copies := map[string]string{} // original tree id -> copy id
	var trees []Tree
	agents := make([]SceneAgent, len(orig.Agents))
	copy(agents, orig.Agents)
	for i, a := range agents {
		if a.TreeID == "" {
			continue
		}
		if id, done := copies[a.TreeID]; done {
			agents[i].TreeID = id
			continue
		}
		t, err := s.Store.GetTree(r.Context(), a.TreeID)
		if err != nil {
			// A sample or deleted tree has nothing to copy; the reference stays.
			continue
		}
		if t.Visibility != VisibilityPublic {
			// Never expose a private tree through a public scene.
			agents[i].TreeID = ""
			continue
		}
		cp := s.copyTree(t, c.Sub, owner, now)
		if err := s.Store.PutTree(r.Context(), cp); err != nil {
			s.storeErr(w, err)
			return
		}
		cp.RemixedFrom = s.withLineageState(r, cp.RemixedFrom)
		copies[a.TreeID] = cp.ID
		agents[i].TreeID = cp.ID
		trees = append(trees, cp)
	}
	ref := LineageRef{Kind: "scene", ID: orig.ID, Title: orig.Name, OwnerID: orig.OwnerID, OwnerName: ownerName, At: now}
	sc := Scene{
		ID: s.NewID(), OwnerID: c.Sub, Name: clip(orig.Name, 80-len(" (Remix)")) + " (Remix)",
		Agents: agents, Visibility: VisibilityPrivate, CreatedAt: now, UpdatedAt: now,
		RemixedFrom: &Lineage{LineageRef: ref, Ancestors: ancestorsOf(orig.RemixedFrom)},
	}
	if err := s.Store.PutScene(r.Context(), sc); err != nil {
		s.storeErr(w, err)
		return
	}
	sc.RemixedFrom = s.withLineageState(r, sc.RemixedFrom)
	writeJSON(w, http.StatusCreated, sceneRemix{Scene: sc, Trees: trees})
}

// reportTree files a report against a public tree, for example an unchanged
// copy of someone else's work. Reporting twice is counted once.
func (s *Server) reportTree(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	t, err := s.Store.GetTree(r.Context(), r.PathValue("id"))
	if err != nil || t.Visibility != VisibilityPublic {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	reason := clip(req.Reason, 500)
	if reason == "" {
		writeErr(w, http.StatusBadRequest, "invalid_request", "reason is required")
		return
	}
	rep := Report{ID: s.NewID(), ThreadID: "tree:" + t.ID, Kind: "tree", ReporterID: c.Sub, Reason: reason, CreatedAt: s.stamp()}
	if err := s.Store.AddReport(r.Context(), rep); err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rep)
}
