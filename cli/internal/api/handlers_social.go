package api

import (
	"net/http"
	"regexp"
)

var displayNameRE = regexp.MustCompile(`^[A-Za-z0-9_-]{3,32}$`)

// ensureProfile returns the caller's profile, creating it on first use.
func (s *Server) ensureProfile(r *http.Request, c Claims) (Profile, error) {
	p, err := s.Store.GetProfile(r.Context(), c.Sub)
	if err == nil {
		return p, nil
	}
	if err != ErrNotFound {
		return Profile{}, err
	}
	p = Profile{UserID: c.Sub, DisplayName: c.Username, CreatedAt: s.stamp()}
	return p, s.Store.PutProfile(r.Context(), p)
}

func (s *Server) getMe(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	p, err := s.ensureProfile(r, c)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	p.Email = c.Email
	writeJSON(w, http.StatusOK, p)
}

type profileRequest struct {
	DisplayName string `json:"displayName"`
	Bio         string `json:"bio"`
}

func (s *Server) updateMe(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	var req profileRequest
	if !readJSON(w, r, &req) {
		return
	}
	if !displayNameRE.MatchString(req.DisplayName) {
		writeErr(w, http.StatusBadRequest, "invalid_request", "displayName must be 3-32 characters: letters, digits, _ or -")
		return
	}
	p, err := s.ensureProfile(r, c)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	p.DisplayName = req.DisplayName
	p.Bio = clip(req.Bio, 500)
	if err := s.Store.PutProfile(r.Context(), p); err != nil {
		s.storeErr(w, err)
		return
	}
	p.Email = c.Email
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) myTrees(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	ts, err := s.Store.ListTreesByOwner(r.Context(), c.Sub)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, listOf(summarize(ts)))
}

func (s *Server) myStats(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	st, err := s.Store.Stats(r.Context(), c.Sub)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// getUser returns a public profile plus that user's public trees.
func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := s.Store.GetProfile(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	ts, err := s.Store.ListTreesByOwner(r.Context(), id)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	var pub []Tree
	for _, t := range ts {
		if t.Visibility == VisibilityPublic {
			pub = append(pub, t)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"profile": p, "trees": listOf(summarize(pub))["items"]})
}

func (s *Server) listThreads(w http.ResponseWriter, r *http.Request) {
	ts, err := s.Store.ListThreads(r.Context(), 100)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, listOf(ts))
}

func (s *Server) createThread(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	var req bodyRequest
	if !readJSON(w, r, &req) {
		return
	}
	title, body := clip(req.Title, 120), clip(req.Body, 5000)
	if title == "" || body == "" {
		writeErr(w, http.StatusBadRequest, "invalid_request", "title and body are required")
		return
	}
	now := s.stamp()
	t := Thread{ID: s.NewID(), Title: title, Body: body, AuthorID: c.Sub, Author: s.ownerName(r, c), CreatedAt: now, LastActivity: now}
	if err := s.Store.PutThread(r.Context(), t); err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (s *Server) getThread(w http.ResponseWriter, r *http.Request) {
	t, err := s.Store.GetThread(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	posts, err := s.Store.ListPosts(r.Context(), t.ID)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"thread": t, "posts": listOf(posts)["items"]})
}

func (s *Server) addPost(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	var req bodyRequest
	if !readJSON(w, r, &req) {
		return
	}
	body := clip(req.Body, 5000)
	if body == "" {
		writeErr(w, http.StatusBadRequest, "invalid_request", "body is required")
		return
	}
	p := Post{ID: s.NewID(), ThreadID: r.PathValue("id"), AuthorID: c.Sub, Author: s.ownerName(r, c), Body: body, CreatedAt: s.stamp()}
	if err := s.Store.AddPost(r.Context(), p); err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}
