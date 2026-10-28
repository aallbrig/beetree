package api

import (
	"math"
	"net/http"
)

const (
	maxSceneAgents = 24
	sceneHalfSize  = 50.0
)

type sceneRequest struct {
	Name       string       `json:"name"`
	Visibility string       `json:"visibility"`
	Agents     []SceneAgent `json:"agents"`
}

func (s *Server) buildScene(w http.ResponseWriter, req sceneRequest, base Scene) (Scene, bool) {
	sc := base
	sc.Name = clip(req.Name, 80)
	if sc.Name == "" {
		writeErr(w, http.StatusBadRequest, "invalid_request", "name is required")
		return Scene{}, false
	}
	if len(req.Agents) > maxSceneAgents {
		writeErr(w, http.StatusBadRequest, "invalid_request", "too many agents in one scene")
		return Scene{}, false
	}
	sc.Visibility = VisibilityPrivate
	if req.Visibility == VisibilityPublic {
		sc.Visibility = VisibilityPublic
	}
	sc.Agents = make([]SceneAgent, 0, len(req.Agents))
	for _, a := range req.Agents {
		if math.IsNaN(a.X) || math.IsNaN(a.Z) || math.Abs(a.X) > sceneHalfSize || math.Abs(a.Z) > sceneHalfSize {
			writeErr(w, http.StatusBadRequest, "invalid_request", "agent position is outside the scene")
			return Scene{}, false
		}
		a.Name, a.Model, a.TreeID = clip(a.Name, 40), clip(a.Model, 40), clip(a.TreeID, 40)
		sc.Agents = append(sc.Agents, a)
	}
	return sc, true
}

func (s *Server) listScenes(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	sc, err := s.Store.ListScenesByOwner(r.Context(), c.Sub)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, listOf(sc))
}

func (s *Server) createScene(w http.ResponseWriter, r *http.Request) {
	c, ok := requireAuth(w, r)
	if !ok {
		return
	}
	var req sceneRequest
	if !readJSON(w, r, &req) {
		return
	}
	sc, ok := s.buildScene(w, req, Scene{})
	if !ok {
		return
	}
	sc.ID, sc.OwnerID = s.NewID(), c.Sub
	sc.CreatedAt = s.stamp()
	sc.UpdatedAt = sc.CreatedAt
	if err := s.Store.PutScene(r.Context(), sc); err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sc)
}

func (s *Server) getScene(w http.ResponseWriter, r *http.Request) {
	sc, err := s.Store.GetScene(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	if c, ok := ClaimsFrom(r.Context()); sc.Visibility != VisibilityPublic && (!ok || c.Sub != sc.OwnerID) {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	writeJSON(w, http.StatusOK, sc)
}

func (s *Server) ownedScene(w http.ResponseWriter, r *http.Request) (Scene, bool) {
	c, ok := requireAuth(w, r)
	if !ok {
		return Scene{}, false
	}
	sc, err := s.Store.GetScene(r.Context(), r.PathValue("id"))
	if err != nil || sc.OwnerID != c.Sub {
		writeErr(w, http.StatusNotFound, "not_found", "not found")
		return Scene{}, false
	}
	return sc, true
}

func (s *Server) updateScene(w http.ResponseWriter, r *http.Request) {
	old, ok := s.ownedScene(w, r)
	if !ok {
		return
	}
	var req sceneRequest
	if !readJSON(w, r, &req) {
		return
	}
	sc, ok := s.buildScene(w, req, old)
	if !ok {
		return
	}
	sc.UpdatedAt = s.stamp()
	if err := s.Store.PutScene(r.Context(), sc); err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sc)
}

func (s *Server) deleteScene(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.ownedScene(w, r)
	if !ok {
		return
	}
	if err := s.Store.DeleteScene(r.Context(), sc.ID); err != nil {
		s.storeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
