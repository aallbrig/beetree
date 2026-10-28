package api

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"gopkg.in/yaml.v3"
)

//go:embed openapi.yaml
var openapiYAML []byte

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": s.Version})
}

func (s *Server) openapi(w http.ResponseWriter, _ *http.Request) {
	var doc any
	if err := yaml.Unmarshal(openapiYAML, &doc); err != nil {
		writeErr(w, http.StatusInternalServerError, "internal", "bad embedded spec")
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, doc)
}

func (s *Server) getNodeTypes(w http.ResponseWriter, _ *http.Request) {
	types, decs := nodeTypes()
	writeJSON(w, http.StatusOK, map[string]any{"nodeTypes": types, "decorators": decs, "engines": Engines, "renderFormats": RenderFormats})
}

type specRequest struct {
	Spec      string            `json:"spec"`
	Format    string            `json:"format"`
	Engine    string            `json:"engine"`
	Overrides map[string]string `json:"overrides"`
}

func (s *Server) validate(w http.ResponseWriter, r *http.Request) {
	var req specRequest
	if !readJSON(w, r, &req) {
		return
	}
	res, _ := validateYAML(req.Spec)
	if res.Errors == nil {
		res.Errors = []string{}
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) simulate(w http.ResponseWriter, r *http.Request) {
	var req specRequest
	if !readJSON(w, r, &req) {
		return
	}
	res, err := simulateYAML(req.Spec, req.Overrides)
	if err != nil {
		writeErr(w, http.StatusUnprocessableEntity, "invalid_spec", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) render(w http.ResponseWriter, r *http.Request) {
	var req specRequest
	if !readJSON(w, r, &req) {
		return
	}
	out, err := renderYAML(req.Spec, req.Format)
	if err != nil {
		writeErr(w, http.StatusUnprocessableEntity, "invalid_request", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"format": req.Format, "output": out})
}

func (s *Server) generate(w http.ResponseWriter, r *http.Request) {
	var req specRequest
	if !readJSON(w, r, &req) {
		return
	}
	files, err := generateYAML(req.Spec, req.Engine)
	if err != nil {
		writeErr(w, http.StatusUnprocessableEntity, "invalid_request", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"engine": req.Engine, "files": files})
}

// jsonRaw is a tiny helper for building JSON-RPC results.
func jsonRaw(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
