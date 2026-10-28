package api

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type treeRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Visibility  string   `json:"visibility"`
	Tags        []string `json:"tags"`
	Spec        string   `json:"spec"`
}

func normalizeTags(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, t := range in {
		t = strings.ToLower(clip(t, 24))
		if t != "" && !seen[t] && len(out) < 8 {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}

// ownerName resolves the display name stored on content authored by c.
func (s *Server) ownerName(r *http.Request, c Claims) string {
	p, err := s.ensureProfile(r, c)
	if err != nil || p.DisplayName == "" {
		return c.Username
	}
	return p.DisplayName
}

// canRead reports whether the caller may see t.
func canRead(r *http.Request, t Tree) bool {
	if t.Visibility == VisibilityPublic {
		return true
	}
	c, ok := ClaimsFrom(r.Context())
	return ok && c.Sub == t.OwnerID
}

func summarize(ts []Tree) []Tree {
	out := make([]Tree, len(ts))
	for i, t := range ts {
		t.Spec = ""
		out[i] = t
	}
	return out
}

func (s *Server) listTrees(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if n, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && n > 0 && n <= 100 {
		limit = n
	}
	ts, err := s.Store.ListPublicTrees(r.Context(), 200)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	q := strings.ToLower(r.URL.Query().Get("q"))
	tag := strings.ToLower(r.URL.Query().Get("tag"))
	var out []Tree
	for _, t := range ts {
		if q != "" && !strings.Contains(strings.ToLower(t.Name+" "+t.Description+" "+strings.Join(t.Tags, " ")), q) {
			continue
		}
		if tag != "" && !containsStr(t.Tags, tag) {
			continue
		}
		out = append(out, t)
	}
	if r.URL.Query().Get("sort") == "popular" {
		sort.SliceStable(out, func(i, j int) bool { return out[i].LikeCount > out[j].LikeCount })
	}
	if len(out) > limit {
		out = out[:limit]
	}
	writeJSON(w, http.StatusOK, listOf(summarize(out)))
}

func containsStr(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func (s *Server) createTree(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	var req treeRequest
	if !readJSON(w, r, &req) {
		return
	}
	t, err := s.createTreeFor(r, c, req)
	if err != nil {
		var ve *validationError
		if errors.As(err, &ve) {
			writeErr(w, ve.status, ve.code, ve.msg)
			return
		}
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

// createTreeFor validates req and stores a new tree owned by c.
func (s *Server) createTreeFor(r *http.Request, c Claims, req treeRequest) (Tree, error) {
	t, err := s.buildTree(req, Tree{})
	if err != nil {
		return Tree{}, err
	}
	t.ID = s.NewID()
	t.OwnerID = c.Sub
	t.OwnerName = s.ownerName(r, c)
	t.CreatedAt = s.stamp()
	t.UpdatedAt = t.CreatedAt
	if err := s.Store.PutTree(r.Context(), t); err != nil {
		return Tree{}, err
	}
	return t, nil
}

type validationError struct {
	status    int
	code, msg string
}

func (e *validationError) Error() string { return e.msg }

// buildTree validates req and merges it over base.
func (s *Server) buildTree(req treeRequest, base Tree) (Tree, error) {
	sp, err := validSpec(req.Spec)
	if err != nil {
		return Tree{}, &validationError{http.StatusUnprocessableEntity, "invalid_spec", err.Error()}
	}
	t := base
	t.Spec = req.Spec
	t.Name = clip(req.Name, 80)
	if t.Name == "" {
		t.Name = clip(sp.Metadata.Name, 80)
	}
	if t.Name == "" {
		return Tree{}, &validationError{http.StatusBadRequest, "invalid_request", "name is required"}
	}
	t.Description = clip(req.Description, 500)
	if t.Description == "" {
		t.Description = clip(sp.Metadata.Description, 500)
	}
	switch req.Visibility {
	case "", VisibilityPrivate:
		t.Visibility = VisibilityPrivate
	case VisibilityPublic:
		t.Visibility = VisibilityPublic
	default:
		return Tree{}, &validationError{http.StatusBadRequest, "invalid_request", "visibility must be public or private"}
	}
	tags := req.Tags
	if len(tags) == 0 {
		tags = sp.Metadata.Tags
	}
	t.Tags = normalizeTags(tags)
	return t, nil
}

func (s *Server) getTree(w http.ResponseWriter, r *http.Request) {
	t, err := s.Store.GetTree(r.Context(), r.PathValue("id"))
	if err != nil || !canRead(r, t) {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	if c, ok := ClaimsFrom(r.Context()); ok {
		t.LikedByMe, _ = s.Store.HasLiked(r.Context(), t.ID, c.Sub)
	}
	writeJSON(w, http.StatusOK, t)
}

// ownedTree loads a tree the caller owns, writing the error response otherwise.
func (s *Server) ownedTree(w http.ResponseWriter, r *http.Request) (Tree, Claims, bool) {
	c, ok := requireAuth(w, r)
	if !ok {
		return Tree{}, c, false
	}
	t, err := s.Store.GetTree(r.Context(), r.PathValue("id"))
	if err != nil || t.OwnerID != c.Sub {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return Tree{}, c, false
	}
	return t, c, true
}

func (s *Server) updateTree(w http.ResponseWriter, r *http.Request) {
	old, _, ok := s.ownedTree(w, r)
	if !ok {
		return
	}
	var req treeRequest
	if !readJSON(w, r, &req) {
		return
	}
	t, err := s.buildTree(req, old)
	if err != nil {
		var ve *validationError
		if errors.As(err, &ve) {
			writeErr(w, ve.status, ve.code, ve.msg)
			return
		}
		s.storeErr(w, err)
		return
	}
	t.UpdatedAt = s.stamp()
	if err := s.Store.PutTree(r.Context(), t); err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) deleteTree(w http.ResponseWriter, r *http.Request) {
	t, _, ok := s.ownedTree(w, r)
	if !ok {
		return
	}
	if err := s.Store.DeleteTree(r.Context(), t.ID); err != nil {
		s.storeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) setLike(w http.ResponseWriter, r *http.Request, liked bool) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	t, err := s.Store.GetTree(r.Context(), r.PathValue("id"))
	if err != nil || !canRead(r, t) {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	n, err := s.Store.SetLike(r.Context(), t.ID, c.Sub, liked)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"liked": liked, "likeCount": n})
}

func (s *Server) likeTree(w http.ResponseWriter, r *http.Request)   { s.setLike(w, r, true) }
func (s *Server) unlikeTree(w http.ResponseWriter, r *http.Request) { s.setLike(w, r, false) }

func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	t, err := s.Store.GetTree(r.Context(), r.PathValue("id"))
	if err != nil || !canRead(r, t) {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	cs, err := s.Store.ListComments(r.Context(), t.ID)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, listOf(cs))
}

type bodyRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (s *Server) addComment(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	t, err := s.Store.GetTree(r.Context(), r.PathValue("id"))
	if err != nil || !canRead(r, t) {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	var req bodyRequest
	if !readJSON(w, r, &req) {
		return
	}
	body := clip(req.Body, 2000)
	if body == "" {
		writeErr(w, http.StatusBadRequest, "invalid_request", "body is required")
		return
	}
	cm := Comment{ID: s.NewID(), TreeID: t.ID, AuthorID: c.Sub, Author: s.ownerName(r, c), Body: body, CreatedAt: s.stamp()}
	if err := s.Store.AddComment(r.Context(), cm); err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, cm)
}

func (s *Server) deleteComment(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	treeID, commentID := r.PathValue("id"), r.PathValue("commentId")
	t, err := s.Store.GetTree(r.Context(), treeID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	cm, err := s.Store.GetComment(r.Context(), treeID, commentID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	// the comment's author or the tree's owner may remove it
	if cm.AuthorID != c.Sub && t.OwnerID != c.Sub {
		writeErr(w, http.StatusForbidden, "forbidden", "not allowed to delete this comment")
		return
	}
	if err := s.Store.DeleteComment(r.Context(), treeID, commentID); err != nil {
		s.storeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
