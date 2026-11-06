package api

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Scopes are resource:action strings. Actions chain upward: delete implies
// write implies read, so a key granted trees:delete may also write and read trees.
var scopeResources = map[string][]string{
	"trees":   {"read", "write", "delete"},
	"scenes":  {"read", "write", "delete"},
	"forum":   {"read", "write", "delete"},
	"profile": {"read", "write"},
	"keys":    {"read", "write"},
}

// Presets offered when creating a key. keys:* is never part of a preset.
var scopePresets = map[string][]string{
	"read-only":  {"trees:read", "scenes:read", "forum:read", "profile:read"},
	"read-write": {"trees:write", "scenes:write", "forum:write", "profile:write"},
	"full":       {"trees:delete", "scenes:delete", "forum:delete", "profile:write"},
}

// Key expiry choices in days; a key made without a choice lasts 90 days.
var expiryChoices = map[string]int{"30d": 30, "90d": 90, "365d": 365, "never": 0}

// lastUsedGranularity limits last-used writes to one per key per interval.
const lastUsedGranularity = time.Hour

func actionRank(resource, action string) int {
	for i, a := range scopeResources[resource] {
		if a == action {
			return i
		}
	}
	return -1
}

func splitScope(sc string) (resource, action string, ok bool) {
	resource, action, ok = strings.Cut(sc, ":")
	if !ok || actionRank(resource, action) < 0 {
		return "", "", false
	}
	return resource, action, true
}

// validScope reports whether sc names a real resource and action.
func validScope(sc string) bool {
	_, _, ok := splitScope(sc)
	return ok
}

// scopeAllows reports whether the granted scopes cover required, applying the
// delete > write > read implication chain.
func scopeAllows(granted []string, required string) bool {
	res, act, ok := splitScope(required)
	if !ok {
		return false
	}
	need := actionRank(res, act)
	for _, g := range granted {
		gr, ga, ok := splitScope(g)
		if ok && gr == res && actionRank(gr, ga) >= need {
			return true
		}
	}
	return false
}

// normalizeScopes drops duplicates and scopes already implied by a higher one,
// and sorts the rest. Unknown scopes are the caller's error to catch first.
func normalizeScopes(in []string) []string {
	var out []string
	for _, sc := range in {
		if !validScope(sc) {
			continue
		}
		implied := false
		for _, other := range in {
			if other != sc && scopeAllows([]string{other}, sc) {
				implied = true
			}
		}
		if !implied && !containsStr(out, sc) {
			out = append(out, sc)
		}
	}
	sort.Strings(out)
	return out
}

// effectiveScopes is what a stored key may do. Keys made before scopes existed
// record none and are treated as Full minus keys:* until they are revoked.
func (k APIKey) effectiveScopes() []string {
	if len(k.Scopes) == 0 {
		return scopePresets["full"]
	}
	return k.Scopes
}

// expired reports whether the key's expiry has passed. No expiry means never.
func (k APIKey) expired(now time.Time) bool {
	if k.ExpiresAt == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339Nano, k.ExpiresAt)
	return err != nil || !now.Before(t)
}

type keyScopesCtx struct{}

// keyScopes returns the scopes a key-authenticated request is limited to.
// Signed-in sessions are not scoped (ok is false).
func keyScopes(ctx context.Context) (scopes []string, ok bool) {
	scopes, ok = ctx.Value(keyScopesCtx{}).([]string)
	return scopes, ok
}

// checkScope returns true when the request may use required. Sessions always
// may; keys need a covering scope. The owner's own rights are enforced by the
// handlers themselves (ownership, moderator group), so a key's effective right
// is its scopes intersected with what its owner can do right now.
func checkScope(ctx context.Context, required string) (granted []string, ok bool) {
	if required == "" || required == scopeMCP {
		return nil, true
	}
	granted, limited := keyScopes(ctx)
	if !limited {
		return nil, true
	}
	return granted, scopeAllows(granted, required)
}

func writeInsufficientScope(w http.ResponseWriter, required string, granted []string) {
	if granted == nil {
		granted = []string{}
	}
	writeJSON(w, http.StatusForbidden, map[string]any{"error": map[string]any{
		"code":     "insufficient_scope",
		"message":  "this API key lacks the " + required + " scope; create a key that includes it",
		"required": required,
		"granted":  granted,
	}})
}

// scopeMCP marks the MCP routes, which check a scope per tool instead.
const scopeMCP = "mcp"

// routeScopes maps every route to the scope it needs. "" means no scope
// (public or stateless). A test fails when Routes() has an entry missing here.
var routeScopes = map[string]string{
	"GET /v1/health":         "",
	"GET /v1/openapi.json":   "",
	"GET /v1/public/trees":   "",
	"GET /v1/node-types":     "",
	"GET /v1/templates":      "",
	"POST /v1/validate":      "",
	"POST /v1/simulate":      "",
	"POST /v1/simulate/step": "",
	"POST /v1/render":        "",
	"POST /v1/generate":      "",

	"GET /v1/trees":                              "trees:read",
	"POST /v1/trees":                             "trees:write",
	"GET /v1/trees/{id}":                         "trees:read",
	"PUT /v1/trees/{id}":                         "trees:write",
	"DELETE /v1/trees/{id}":                      "trees:delete",
	"POST /v1/trees/{id}/remix":                  "trees:write",
	"POST /v1/trees/{id}/report":                 "trees:write",
	"PUT /v1/trees/{id}/like":                    "trees:write",
	"DELETE /v1/trees/{id}/like":                 "trees:write",
	"GET /v1/trees/{id}/comments":                "trees:read",
	"POST /v1/trees/{id}/comments":               "trees:write",
	"DELETE /v1/trees/{id}/comments/{commentId}": "trees:delete",

	"GET /v1/me":         "profile:read",
	"PUT /v1/me":         "profile:write",
	"GET /v1/me/trees":   "trees:read",
	"GET /v1/me/stats":   "profile:read",
	"GET /v1/users/{id}": "profile:read",

	"GET /v1/forum/threads":              "forum:read",
	"POST /v1/forum/threads":             "forum:write",
	"GET /v1/forum/threads/{id}":         "forum:read",
	"POST /v1/forum/threads/{id}/posts":  "forum:write",
	"POST /v1/forum/threads/{id}/report": "forum:write",
	"GET /v1/forum/reports":              "forum:read",
	"GET /v1/forum/threads/{id}/reports": "forum:read",

	"GET /v1/scenes":             "scenes:read",
	"POST /v1/scenes":            "scenes:write",
	"GET /v1/scenes/{id}":        "scenes:read",
	"PUT /v1/scenes/{id}":        "scenes:write",
	"POST /v1/scenes/{id}/remix": "scenes:write",
	"DELETE /v1/scenes/{id}":     "scenes:delete",

	"GET /v1/keys":         "keys:read",
	"POST /v1/keys":        "keys:write",
	"DELETE /v1/keys/{id}": "keys:write",

	"POST /mcp": scopeMCP,
	"GET /mcp":  scopeMCP,
}

// toolScopes maps every MCP tool to the scope it needs ("" for stateless or
// public tools). A test fails when a tool is missing here.
var toolScopes = map[string]string{
	"list_node_types": "",
	"get_template":    "",
	"validate_tree":   "",
	"render_tree":     "",
	"simulate_tree":   "",
	"simulate_start":  "",
	"simulate_answer": "",
	"diff_trees":      "",
	"generate_code":   "",
	"search_trees":    "",
	"get_tree":        "trees:read",
	"create_tree":     "trees:write",
	"update_tree":     "trees:write",
	"delete_tree":     "trees:delete",
	"remix_tree":      "trees:write",
	"remix_scene":     "scenes:write",
	"get_scene":       "scenes:read",
	"list_my_scenes":  "scenes:read",
	"list_my_trees":   "trees:read",
}

// scoped wraps a route handler with its scope check.
func scoped(required string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if granted, ok := checkScope(r.Context(), required); !ok {
			writeInsufficientScope(w, required, granted)
			return
		}
		h(w, r)
	}
}

// keyRequest is the body of POST /v1/keys.
type keyRequest struct {
	Name   string   `json:"name"`
	Preset string   `json:"preset"`
	Scopes []string `json:"scopes"`
	Expiry string   `json:"expiry"`
}

// resolveScopes turns a key request into the scopes to store, or an error message.
func resolveScopes(req keyRequest) ([]string, string) {
	var scopes []string
	switch {
	case req.Preset != "" && len(req.Scopes) > 0:
		return nil, "send either preset or scopes, not both"
	case req.Preset != "":
		p, ok := scopePresets[req.Preset]
		if !ok {
			return nil, "unknown preset; use read-only, read-write or full"
		}
		scopes = p
	case len(req.Scopes) > 0:
		for _, sc := range req.Scopes {
			if !validScope(sc) {
				return nil, "unknown scope: " + sc
			}
		}
		scopes = req.Scopes
	default:
		return nil, "choose a preset (read-only, read-write, full) or list scopes; keys have no scopes by default"
	}
	return normalizeScopes(scopes), ""
}

// resolveExpiry returns the RFC 3339 expiry for the choice ("" for never).
func resolveExpiry(choice string, now time.Time) (string, bool) {
	if choice == "" {
		choice = "90d"
	}
	days, ok := expiryChoices[choice]
	if !ok {
		return "", false
	}
	if days == 0 {
		return "", true
	}
	return now.UTC().AddDate(0, 0, days).Format(time.RFC3339Nano), true
}
