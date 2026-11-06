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
	p.Moderator = c.IsModerator()
	writeJSON(w, http.StatusOK, p)
}

type profileRequest struct {
	DisplayName string `json:"displayName"`
	Bio         string `json:"bio"`
	// Avatar, when present, picks a preset picture ("" clears it).
	Avatar *string `json:"avatar"`
	// MarketingOptIn, when present, records or withdraws consent to email campaigns.
	MarketingOptIn *bool `json:"marketingOptIn"`
	// MarketingSource says where the choice was made: "signup" or "profile".
	MarketingSource string `json:"marketingSource"`
}

var avatarRE = regexp.MustCompile(`^a(0[0-9][1-9]|0[1-9]0|100)$`)

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
	if req.Avatar != nil && *req.Avatar != "" && !avatarRE.MatchString(*req.Avatar) {
		writeErr(w, http.StatusBadRequest, "invalid_request", "avatar must be one of a001 to a100")
		return
	}
	p.DisplayName = req.DisplayName
	p.Bio = clip(req.Bio, 500)
	if req.Avatar != nil {
		p.Avatar = *req.Avatar
	}
	if req.MarketingOptIn != nil && *req.MarketingOptIn != p.MarketingOptIn {
		p.MarketingOptIn = *req.MarketingOptIn
		p.MarketingAt = s.stamp()
		p.MarketingSource = "profile"
		if req.MarketingSource == "signup" {
			p.MarketingSource = "signup"
		}
	}
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
	scenes, err := s.Store.ListScenesByOwner(r.Context(), id)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	pubScenes := []Scene{}
	for _, sc := range scenes {
		if sc.Visibility == VisibilityPublic {
			pubScenes = append(pubScenes, sc)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"profile": p.Public(), "trees": listOf(summarize(pub))["items"], "scenes": pubScenes})
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
	title, body := plainText(req.Title, 120, true), plainText(req.Body, 5000, false)
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

// reportThread files a report against a thread, or against one of its posts
// when postId is given. Reporting twice is accepted and counted once.
func (s *Server) reportThread(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	var req struct {
		PostID string `json:"postId"`
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
	id := r.PathValue("id")
	if req.PostID != "" {
		posts, err := s.Store.ListPosts(r.Context(), id)
		if err != nil {
			s.storeErr(w, err)
			return
		}
		found := false
		for _, p := range posts {
			found = found || p.ID == req.PostID
		}
		if !found {
			writeErr(w, http.StatusNotFound, "not_found", "post not found")
			return
		}
	}
	rep := Report{ID: s.NewID(), ThreadID: id, PostID: req.PostID, ReporterID: c.Sub, Reason: reason, CreatedAt: s.stamp()}
	if err := s.Store.AddReport(r.Context(), rep); err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rep)
}

// listReports is the moderators-only review queue for one thread.
func (s *Server) listReports(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	if !c.IsModerator() {
		writeErr(w, http.StatusForbidden, "forbidden", "moderators only")
		return
	}
	id := r.PathValue("id")
	if _, err := s.Store.GetThread(r.Context(), id); err != nil {
		s.storeErr(w, err)
		return
	}
	reps, err := s.Store.ListReports(r.Context(), id)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	if reps == nil {
		reps = []Report{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": reps})
}

// listRecentReports is the moderators-only review queue across all threads.
func (s *Server) listRecentReports(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	if !c.IsModerator() {
		writeErr(w, http.StatusForbidden, "forbidden", "moderators only")
		return
	}
	reps, err := s.Store.ListRecentReports(r.Context(), 100)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	if reps == nil {
		reps = []Report{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": reps})
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
	body := plainText(req.Body, 5000, false)
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
