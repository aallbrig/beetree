package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxBodyBytes = 512 * 1024

// Server wires the Store and Authenticator into an http.Handler.
type Server struct {
	Store Store
	Auth  Authenticator
	Now   func() time.Time
	NewID func() string
	// Version is reported by /v1/health.
	Version string
}

// Route is one API operation. The route table drives both registration and the
// OpenAPI conformance test, so the spec cannot silently drift from the code.
type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

// Routes returns every REST and MCP route.
func (s *Server) Routes() []Route {
	return []Route{
		{"GET", "/v1/health", s.health},
		{"GET", "/v1/openapi.json", s.openapi},
		{"GET", "/v1/node-types", s.getNodeTypes},
		{"POST", "/v1/validate", s.validate},
		{"POST", "/v1/simulate", s.simulate},
		{"POST", "/v1/render", s.render},
		{"POST", "/v1/generate", s.generate},

		{"GET", "/v1/trees", s.listTrees},
		{"POST", "/v1/trees", s.createTree},
		{"GET", "/v1/trees/{id}", s.getTree},
		{"PUT", "/v1/trees/{id}", s.updateTree},
		{"DELETE", "/v1/trees/{id}", s.deleteTree},
		{"PUT", "/v1/trees/{id}/like", s.likeTree},
		{"DELETE", "/v1/trees/{id}/like", s.unlikeTree},
		{"GET", "/v1/trees/{id}/comments", s.listComments},
		{"POST", "/v1/trees/{id}/comments", s.addComment},
		{"DELETE", "/v1/trees/{id}/comments/{commentId}", s.deleteComment},

		{"GET", "/v1/me", s.getMe},
		{"PUT", "/v1/me", s.updateMe},
		{"GET", "/v1/me/trees", s.myTrees},
		{"GET", "/v1/me/stats", s.myStats},
		{"GET", "/v1/users/{id}", s.getUser},

		{"GET", "/v1/forum/threads", s.listThreads},
		{"POST", "/v1/forum/threads", s.createThread},
		{"GET", "/v1/forum/threads/{id}", s.getThread},
		{"POST", "/v1/forum/threads/{id}/posts", s.addPost},

		{"GET", "/v1/scenes", s.listScenes},
		{"POST", "/v1/scenes", s.createScene},
		{"GET", "/v1/scenes/{id}", s.getScene},
		{"PUT", "/v1/scenes/{id}", s.updateScene},
		{"DELETE", "/v1/scenes/{id}", s.deleteScene},

		{"POST", "/mcp", s.mcp},
		{"GET", "/mcp", s.mcpGet},
	}
}

// Handler builds the http.Handler with CORS, auth extraction and body limits.
func (s *Server) Handler() http.Handler {
	if s.Now == nil {
		s.Now = time.Now
	}
	if s.NewID == nil {
		s.NewID = randomID
	}
	mux := http.NewServeMux()
	for _, r := range s.Routes() {
		mux.HandleFunc(r.Method+" "+r.Path, r.Handler)
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		writeErr(w, http.StatusNotFound, "not_found", "no such route")
	})
	return s.middleware(mux)
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Mcp-Session-Id")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		h.Set("Access-Control-Max-Age", "600")
		h.Set("X-Content-Type-Options", "nosniff")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		if tok := bearerToken(r); tok != "" && s.Auth != nil {
			c, err := s.Auth.Verify(r.Context(), tok)
			if err != nil {
				writeErr(w, http.StatusUnauthorized, "unauthenticated", "invalid or expired token")
				return
			}
			r = r.WithContext(WithClaims(r.Context(), c))
		}
		next.ServeHTTP(w, r)
	})
}

// ---- helpers ---------------------------------------------------------------

func randomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Server) stamp() string { return s.Now().UTC().Format(time.RFC3339Nano) }

type errBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	var b errBody
	b.Error.Code, b.Error.Message = code, msg
	writeJSON(w, status, b)
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, http.StatusRequestEntityTooLarge, "too_large", "request body too large")
		return false
	}
	if err := json.Unmarshal(data, v); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_json", "request body is not valid JSON")
		return false
	}
	return true
}

func (s *Server) storeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeErr(w, http.StatusNotFound, "not_found", "not found")
	case errors.Is(err, ErrConflict):
		writeErr(w, http.StatusConflict, "conflict", "conflict")
	default:
		writeErr(w, http.StatusInternalServerError, "internal", "internal error")
	}
}

// requireAuth returns the caller or writes a 401.
func requireAuth(w http.ResponseWriter, r *http.Request) (Claims, bool) {
	c, ok := ClaimsFrom(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthenticated", "sign in required")
	}
	return c, ok
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len([]rune(s)) > n {
		return string([]rune(s)[:n])
	}
	return s
}

func listOf[T any](items []T) map[string]any {
	if items == nil {
		items = []T{}
	}
	return map[string]any{"items": items}
}
