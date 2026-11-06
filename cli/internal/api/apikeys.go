package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
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
	// Scopes are immutable after creation. Empty means a key made before scopes
	// existed: Full minus keys:* (see effectiveScopes).
	Scopes     []string `json:"scopes,omitempty"`
	ExpiresAt  string   `json:"expiresAt,omitempty"`
	LastUsedAt string   `json:"lastUsedAt,omitempty"`
}

// APIKeyInfo is what the API returns about a key; it never includes the hash.
type APIKeyInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	CreatedAt string `json:"createdAt"`
	// Scopes is the effective list (legacy keys show Full minus keys:*).
	Scopes     []string `json:"scopes"`
	Legacy     bool     `json:"legacy,omitempty"`
	ExpiresAt  string   `json:"expiresAt,omitempty"`
	LastUsedAt string   `json:"lastUsedAt,omitempty"`
}

func (k APIKey) info() APIKeyInfo {
	return APIKeyInfo{
		ID: k.ID, Name: k.Name, Prefix: k.Prefix, CreatedAt: k.CreatedAt,
		Scopes: k.effectiveScopes(), Legacy: len(k.Scopes) == 0,
		ExpiresAt: k.ExpiresAt, LastUsedAt: k.LastUsedAt,
	}
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
func publicPath(p string) bool {
	return p == "/v1/health" || p == "/v1/openapi.json" || p == "/v1/public/trees"
}

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
	var req keyRequest
	if !readJSON(w, r, &req) {
		return
	}
	name := clip(req.Name, 60)
	if name == "" {
		writeErr(w, http.StatusBadRequest, "invalid", "name is required")
		return
	}
	scopes, msg := resolveScopes(req)
	if msg != "" {
		writeErr(w, http.StatusBadRequest, "invalid", msg)
		return
	}
	expires, ok := resolveExpiry(req.Expiry, s.Now())
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid", "expiry must be 30d, 90d, 365d or never")
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
		Scopes: scopes, ExpiresAt: expires,
	}
	if err := s.Store.PutAPIKey(r.Context(), k); err != nil {
		s.storeErr(w, err)
		return
	}
	created := struct {
		APIKeyInfo
		Key  string `json:"key"`
		Note string `json:"note"`
	}{k.info(), secret, "Copy this key now; it is shown only once."}
	writeJSON(w, http.StatusCreated, created)
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

// touchKey records last use at most once an hour per key. It is best effort: a
// failure never blocks the request.
func (s *Server) touchKey(ctx context.Context, k APIKey) {
	now := s.Now().UTC()
	if last, err := time.Parse(time.RFC3339Nano, k.LastUsedAt); err == nil && now.Sub(last) < lastUsedGranularity {
		return
	}
	_ = s.Store.TouchAPIKey(ctx, k.Hash, now.Format(time.RFC3339Nano))
}
