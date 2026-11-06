package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"
)

const (
	apiKeyPrefix   = "btc_"
	maxKeysPerUser = 5
	docsURL        = "https://beetreecraft.com/docs/api/"
)

// APIKey is the stored record for a key. The secret itself is never stored,
// only its SHA-256 hash, so a database leak does not reveal usable keys.
type APIKey struct {
	ID        string `json:"id"`
	OwnerID   string `json:"ownerId"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	Hash      string `json:"hash"`
	CreatedAt string `json:"createdAt"`
}

// APIKeyInfo is what the API returns about a key; it never includes the hash.
type APIKeyInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	CreatedAt string `json:"createdAt"`
}

func (k APIKey) info() APIKeyInfo {
	return APIKeyInfo{ID: k.ID, Name: k.Name, Prefix: k.Prefix, CreatedAt: k.CreatedAt}
}

// newAPIKeySecret returns a fresh key: the btc_ prefix plus 32 random bytes.
func newAPIKeySecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return apiKeyPrefix + base64.RawURLEncoding.EncodeToString(b), nil
}

func hashAPIKey(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

type viaKeyCtx struct{}

// viaAPIKey reports whether the caller authenticated with an API key rather
// than a signed-in session.
func viaAPIKey(ctx context.Context) bool {
	v, _ := ctx.Value(viaKeyCtx{}).(bool)
	return v
}

// requireSession is requireAuth for key management: only a signed-in session
// qualifies, so a leaked API key cannot mint or revoke keys.
func requireSession(w http.ResponseWriter, r *http.Request) (Claims, bool) {
	c, ok := requireAuth(w, r)
	if !ok {
		return c, false
	}
	if viaAPIKey(r.Context()) {
		writeErr(w, http.StatusForbidden, "session_required", "managing API keys needs a signed-in session, not an API key")
		return c, false
	}
	return c, true
}

// publicPath lists the routes that need no credentials.
func publicPath(p string) bool { return p == "/v1/health" || p == "/v1/openapi.json" }

// credential extracts the caller's token from Authorization: Bearer or X-API-Key.
func credential(r *http.Request) string {
	if t := bearerToken(r); t != "" {
		return t
	}
	return strings.TrimSpace(r.Header.Get("X-API-Key"))
}

func (s *Server) listKeys(w http.ResponseWriter, r *http.Request) {
	c, ok := requireSession(w, r)
	if !ok {
		return
	}
	keys, err := s.Store.ListAPIKeys(r.Context(), c.Sub)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	out := make([]APIKeyInfo, 0, len(keys))
	for _, k := range keys {
		out = append(out, k.info())
	}
	writeJSON(w, http.StatusOK, listOf(out))
}

func (s *Server) createKey(w http.ResponseWriter, r *http.Request) {
	c, ok := requireSession(w, r)
	if !ok {
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	name := clip(req.Name, 60)
	if name == "" {
		writeErr(w, http.StatusBadRequest, "invalid", "name is required")
		return
	}
	existing, err := s.Store.ListAPIKeys(r.Context(), c.Sub)
	if err != nil {
		s.storeErr(w, err)
		return
	}
	if len(existing) >= maxKeysPerUser {
		writeErr(w, http.StatusConflict, "key_limit", "you can have at most 5 API keys; revoke one first")
		return
	}
	secret, err := newAPIKeySecret()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal", "internal error")
		return
	}
	k := APIKey{
		ID: s.NewID(), OwnerID: c.Sub, Username: c.Username, Name: name,
		Prefix: secret[:8], Hash: hashAPIKey(secret), CreatedAt: s.stamp(),
	}
	if err := s.Store.PutAPIKey(r.Context(), k); err != nil {
		s.storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": k.ID, "name": k.Name, "prefix": k.Prefix, "createdAt": k.CreatedAt,
		"key": secret, "note": "Copy this key now; it is shown only once.",
	})
}

func (s *Server) deleteKey(w http.ResponseWriter, r *http.Request) {
	c, ok := requireSession(w, r)
	if !ok {
		return
	}
	if err := s.Store.DeleteAPIKey(r.Context(), c.Sub, r.PathValue("id")); err != nil {
		s.storeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
