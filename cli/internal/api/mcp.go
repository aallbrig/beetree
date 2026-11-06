package api

import (
	"encoding/json"
	"fmt"
	"github.com/aallbrig/beetree-cli/internal/differ"
	"net/http"
	"strings"
)

// mcpProtocolVersion is the MCP revision this server implements (Streamable
// HTTP transport, JSON responses, no server-initiated streams).
const mcpProtocolVersion = "2025-03-26"

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// JSON-RPC error codes used by the server.
const (
	rpcParseError     = -32700
	rpcInvalidRequest = -32600
	rpcMethodNotFound = -32601
	rpcInvalidParams  = -32602
)

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
	// Annotations tell clients which tools are safe to auto-approve.
	Annotations mcpAnnotations `json:"annotations"`
	// needsAuth tools act on the caller's account.
	needsAuth bool
}

// mcpAnnotations are the MCP tool hints. Only create_tree and update_tree
// write; update_tree overwrites the caller's own tree, so it is marked destructive.
type mcpAnnotations struct {
	ReadOnlyHint    bool `json:"readOnlyHint"`
	DestructiveHint bool `json:"destructiveHint"`
	OpenWorldHint   bool `json:"openWorldHint"`
}

// mcpSearchWindow is how many of the newest public trees search_trees scans.
const mcpSearchWindow = 2000

func objSchema(required []string, props map[string]any) map[string]any {
	return map[string]any{"type": "object", "properties": props, "required": required}
}

var strProp = func(desc string) map[string]any { return map[string]any{"type": "string", "description": desc} }

func mcpTools() []mcpTool {
	spec := strProp("Behavior tree specification as .beetree.yaml text")
	tools := []mcpTool{
		{Name: "list_node_types", Description: "List the node types, decorators, render formats and code generation engines BeeTree supports.", InputSchema: objSchema(nil, map[string]any{})},
		{Name: "validate_tree", Description: "Validate a .beetree.yaml spec. Returns whether it is valid and every error found.", InputSchema: objSchema([]string{"spec"}, map[string]any{"spec": spec})},
		{Name: "render_tree", Description: "Render a spec as ascii, sigil, mermaid or dot.", InputSchema: objSchema([]string{"spec"}, map[string]any{"spec": spec, "format": strProp("ascii | sigil | mermaid | dot (default ascii)")})},
		{Name: "simulate_tree", Description: "Dry-run a spec and return the execution trace. Overrides force named nodes to SUCCESS, FAILURE or RUNNING.", InputSchema: objSchema([]string{"spec"}, map[string]any{"spec": spec, "overrides": map[string]any{"type": "object", "description": "node name -> SUCCESS|FAILURE|RUNNING", "additionalProperties": map[string]any{"type": "string"}}})},
		{Name: "simulate_start", Description: "Begin a step-through simulation. Returns the first pending leaf; answer it with simulate_answer.", InputSchema: objSchema([]string{"spec"}, map[string]any{"spec": spec})},
		{Name: "simulate_answer", Description: "Answer the pending leaf of a step-through simulation with SUCCESS, FAILURE or RUNNING. Pass the same spec and every earlier answer in order (the server keeps no state). Returns the next pending leaf, or the final status and trace.", InputSchema: objSchema([]string{"spec", "status"}, map[string]any{"spec": spec, "status": strProp("SUCCESS | FAILURE | RUNNING for the pending leaf"), "answers": map[string]any{"type": "array", "description": "earlier answers, oldest first (empty after simulate_start)", "items": map[string]any{"type": "string"}}})},
		{Name: "diff_trees", Description: "Compare two specs and list what was added, removed or modified, by node path. Use after editing a tree to confirm only the intended change happened.", InputSchema: objSchema([]string{"spec", "other"}, map[string]any{"spec": strProp("the earlier .beetree.yaml text"), "other": strProp("the later .beetree.yaml text")})},
		{Name: "generate_code", Description: "Generate engine code (unity, unreal or godot) from a spec.", InputSchema: objSchema([]string{"spec", "engine"}, map[string]any{"spec": spec, "engine": strProp("unity | unreal | godot")})},
		{Name: "search_trees", Description: "Search public behavior trees shared by the community.", InputSchema: objSchema(nil, map[string]any{"query": strProp("text to match against name, description and tags"), "tag": strProp("exact tag"), "limit": map[string]any{"type": "integer"}})},
		{Name: "get_tree", Description: "Fetch a tree, including its spec, by id. Private trees need the owner's token.", InputSchema: objSchema([]string{"id"}, map[string]any{"id": strProp("tree id")})},
		{Name: "create_tree", Description: "Save a new tree to the signed-in user's account. Requires a bearer token.", needsAuth: true, InputSchema: objSchema([]string{"spec"}, map[string]any{"spec": spec, "name": strProp("display name (defaults to spec metadata.name)"), "description": strProp("short description"), "visibility": strProp("public | private (default private)"), "tags": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}})},
		{Name: "update_tree", Description: "Replace the spec of a tree you own. Omitted name, description, visibility and tags keep their current values. Requires a bearer token.", needsAuth: true, InputSchema: objSchema([]string{"id", "spec"}, map[string]any{"id": strProp("id of the tree to update"), "spec": spec, "name": strProp("new display name"), "description": strProp("new short description"), "visibility": strProp("public | private"), "tags": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}})},
		{Name: "delete_tree", Description: "Permanently delete a tree you own. Requires a bearer token.", needsAuth: true, InputSchema: objSchema([]string{"id"}, map[string]any{"id": strProp("id of the tree to delete")})},
		{Name: "get_scene", Description: "Fetch an Explore scene, including its layout and agents, by id. Private scenes need the owner's token; list_my_scenes gives ids.", InputSchema: objSchema([]string{"id"}, map[string]any{"id": strProp("scene id")})},
		{Name: "list_my_scenes", Description: "List the signed-in user's Explore scenes. Requires a bearer token.", needsAuth: true, InputSchema: objSchema(nil, map[string]any{})},
		{Name: "list_my_trees", Description: "List the signed-in user's trees. Requires a bearer token.", needsAuth: true, InputSchema: objSchema(nil, map[string]any{})},
	}
	for i := range tools {
		n := tools[i].Name
		tools[i].Annotations = mcpAnnotations{ReadOnlyHint: n != "create_tree" && n != "update_tree" && n != "delete_tree", DestructiveHint: n == "update_tree" || n == "delete_tree"}
	}
	return tools
}

func (s *Server) mcpGet(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Allow", "POST")
	writeErr(w, http.StatusMethodNotAllowed, "method_not_allowed", "this MCP server answers JSON-RPC over POST only")
}

func (s *Server) mcp(w http.ResponseWriter, r *http.Request) {
	var req rpcRequest
	if !readRPC(w, r, &req) {
		return
	}
	if req.JSONRPC != "2.0" || req.Method == "" {
		writeRPC(w, rpcResponse{ID: req.ID, Error: &rpcError{rpcInvalidRequest, "invalid JSON-RPC request"}})
		return
	}
	// notifications carry no id and get no body
	if len(req.ID) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	switch req.Method {
	case "initialize":
		writeRPC(w, rpcResponse{ID: req.ID, Result: jsonRaw(map[string]any{
			"protocolVersion": mcpProtocolVersion,
			"capabilities": map[string]any{
				"tools":     map[string]any{"listChanged": false},
				"resources": map[string]any{"listChanged": false},
				"prompts":   map[string]any{"listChanged": false},
			},
			"serverInfo":   map[string]string{"name": "beetreecraft", "version": s.Version},
			"instructions": "BeeTree authors engine-agnostic behavior trees. Pass specs as .beetree.yaml text. Read beetree://schema/beetree for the format and beetree://examples/patrol for a working tree. Workflow: validate_tree before simulate_start; simulate_answer keeps no state, so replay every earlier answer in order each call. Tools that touch an account need an Authorization: Bearer token.",
		})})
	case "ping":
		writeRPC(w, rpcResponse{ID: req.ID, Result: jsonRaw(map[string]any{})})
	case "tools/list":
		writeRPC(w, rpcResponse{ID: req.ID, Result: jsonRaw(map[string]any{"tools": mcpTools()})})
	case "tools/call":
		s.mcpCall(w, r, req)
	case "resources/list":
		writeRPC(w, rpcResponse{ID: req.ID, Result: jsonRaw(map[string]any{"resources": mcpResources()})})
	case "resources/read":
		var p struct {
			URI string `json:"uri"`
		}
		_ = json.Unmarshal(req.Params, &p)
		text, mime, ok := readMCPResource(p.URI)
		if !ok {
			writeRPC(w, rpcResponse{ID: req.ID, Error: &rpcError{rpcInvalidParams, "unknown resource: " + p.URI}})
			return
		}
		writeRPC(w, rpcResponse{ID: req.ID, Result: jsonRaw(map[string]any{"contents": []map[string]string{{"uri": p.URI, "mimeType": mime, "text": text}}})})
	case "prompts/list":
		writeRPC(w, rpcResponse{ID: req.ID, Result: jsonRaw(map[string]any{"prompts": mcpPrompts()})})
	case "prompts/get":
		var p struct {
			Name      string            `json:"name"`
			Arguments map[string]string `json:"arguments"`
		}
		_ = json.Unmarshal(req.Params, &p)
		text, err := promptText(p.Name, p.Arguments)
		if err != nil {
			writeRPC(w, rpcResponse{ID: req.ID, Error: &rpcError{rpcInvalidParams, err.Error()}})
			return
		}
		writeRPC(w, rpcResponse{ID: req.ID, Result: jsonRaw(map[string]any{"messages": []map[string]any{{"role": "user", "content": map[string]string{"type": "text", "text": text}}}})})
	default:
		writeRPC(w, rpcResponse{ID: req.ID, Error: &rpcError{rpcMethodNotFound, "method not found: " + req.Method}})
	}
}

func readRPC(w http.ResponseWriter, r *http.Request, req *rpcRequest) bool {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		writeRPC(w, rpcResponse{ID: json.RawMessage("null"), Error: &rpcError{rpcParseError, "parse error"}})
		return false
	}
	return true
}

func writeRPC(w http.ResponseWriter, resp rpcResponse) {
	resp.JSONRPC = "2.0"
	if len(resp.ID) == 0 {
		resp.ID = json.RawMessage("null")
	}
	writeJSON(w, http.StatusOK, resp)
}

func toolText(text string, isError bool) json.RawMessage {
	return jsonRaw(map[string]any{
		"content": []map[string]string{{"type": "text", "text": text}},
		"isError": isError,
	})
}

func toolJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return toolText(string(b), false)
}

type callParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type toolArgs struct {
	Spec        string            `json:"spec"`
	Other       string            `json:"other"`
	Format      string            `json:"format"`
	Engine      string            `json:"engine"`
	Overrides   map[string]string `json:"overrides"`
	Answers     []string          `json:"answers"`
	Status      string            `json:"status"`
	Query       string            `json:"query"`
	Tag         string            `json:"tag"`
	Limit       int               `json:"limit"`
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Visibility  string            `json:"visibility"`
	Tags        []string          `json:"tags"`
}

func (s *Server) mcpCall(w http.ResponseWriter, r *http.Request, req rpcRequest) {
	var p callParams
	if err := json.Unmarshal(req.Params, &p); err != nil || p.Name == "" {
		writeRPC(w, rpcResponse{ID: req.ID, Error: &rpcError{rpcInvalidParams, "tools/call needs a tool name"}})
		return
	}
	var tool *mcpTool
	for _, t := range mcpTools() {
		if t.Name == p.Name {
			t := t
			tool = &t
		}
	}
	if tool == nil {
		writeRPC(w, rpcResponse{ID: req.ID, Error: &rpcError{rpcInvalidParams, "unknown tool: " + p.Name}})
		return
	}
	var a toolArgs
	if len(p.Arguments) > 0 {
		if err := json.Unmarshal(p.Arguments, &a); err != nil {
			writeRPC(w, rpcResponse{ID: req.ID, Error: &rpcError{rpcInvalidParams, "arguments do not match the tool's input schema"}})
			return
		}
	}
	claims, authed := ClaimsFrom(r.Context())
	if tool.needsAuth && !authed {
		writeRPC(w, rpcResponse{ID: req.ID, Result: toolText("This tool needs a signed-in user. Send an Authorization: Bearer token.", true)})
		return
	}
	writeRPC(w, rpcResponse{ID: req.ID, Result: s.runTool(r, p.Name, a, claims)})
}

func (s *Server) runTool(r *http.Request, name string, a toolArgs, claims Claims) json.RawMessage {
	ctx := r.Context()
	fail := func(err error) json.RawMessage { return toolText(err.Error(), true) }
	switch name {
	case "list_node_types":
		types, decs := nodeTypes()
		return toolJSON(map[string]any{"nodeTypes": types, "decorators": decs, "engines": Engines, "renderFormats": RenderFormats})
	case "validate_tree":
		res, _ := validateYAML(a.Spec)
		return toolJSON(res)
	case "render_tree":
		out, err := renderYAML(a.Spec, a.Format)
		if err != nil {
			return fail(err)
		}
		return toolText(out, false)
	case "simulate_tree":
		res, err := simulateYAML(a.Spec, a.Overrides)
		if err != nil {
			return fail(err)
		}
		return toolText(res.Trace, false)
	case "simulate_start":
		res, err := stepYAML(a.Spec, nil)
		if err != nil {
			return fail(err)
		}
		return toolJSON(res)
	case "simulate_answer":
		res, err := stepYAML(a.Spec, append(append([]string{}, a.Answers...), a.Status))
		if err != nil {
			return fail(err)
		}
		return toolJSON(res)
	case "diff_trees":
		a1, err := validSpec(a.Spec)
		if err != nil {
			return fail(fmt.Errorf("spec: %w", err))
		}
		b1, err := validSpec(a.Other)
		if err != nil {
			return fail(fmt.Errorf("other: %w", err))
		}
		changes := differ.Diff(a1, b1)
		out := make([]map[string]string, 0, len(changes))
		for _, c := range changes {
			out = append(out, map[string]string{"type": string(c.Type), "path": c.Path, "old": c.Old, "new": c.New})
		}
		return toolJSON(map[string]any{"identical": len(changes) == 0, "changes": out})
	case "generate_code":
		files, err := generateYAML(a.Spec, a.Engine)
		if err != nil {
			return fail(err)
		}
		var sb strings.Builder
		for _, f := range files {
			fmt.Fprintf(&sb, "=== %s ===\n%s\n\n", f.Path, f.Content)
		}
		return toolText(sb.String(), false)
	case "search_trees":
		ts, err := s.Store.ListPublicTrees(ctx, mcpSearchWindow)
		if err != nil {
			return fail(fmt.Errorf("search failed"))
		}
		limit := a.Limit
		if limit <= 0 || limit > 50 {
			limit = 20
		}
		q, tag := strings.ToLower(a.Query), strings.ToLower(a.Tag)
		var out []Tree
		full := false
		for _, t := range ts {
			if q != "" && !strings.Contains(strings.ToLower(t.Name+" "+t.Description+" "+strings.Join(t.Tags, " ")), q) {
				continue
			}
			if tag != "" && !containsStr(t.Tags, tag) {
				continue
			}
			if out = append(out, t); len(out) == limit {
				full = true
				break
			}
		}
		// truncated: the scan covered only the newest window and ended without
		// filling the limit, so older public trees may match but were not seen.
		return toolJSON(map[string]any{"trees": summarize(out), "truncated": !full && len(ts) >= mcpSearchWindow})
	case "get_tree":
		t, err := s.Store.GetTree(ctx, a.ID)
		if err != nil || !canRead(r, t) {
			return fail(fmt.Errorf("tree not found"))
		}
		return toolJSON(t)
	case "get_scene":
		sc, err := s.Store.GetScene(ctx, a.ID)
		if err != nil || (sc.Visibility != VisibilityPublic && claims.Sub != sc.OwnerID) {
			return fail(fmt.Errorf("scene not found; use list_my_scenes for your scene ids"))
		}
		return toolJSON(sc)
	case "create_tree":
		t, err := s.createTreeFor(r, claims, treeRequest{Name: a.Name, Description: a.Description, Visibility: a.Visibility, Tags: a.Tags, Spec: a.Spec})
		if err != nil {
			return fail(err)
		}
		return toolJSON(t)
	case "update_tree":
		old, err := s.Store.GetTree(ctx, a.ID)
		if err != nil || old.OwnerID != claims.Sub {
			return fail(fmt.Errorf("tree not found among your trees; use list_my_trees for ids"))
		}
		req := treeRequest{Name: a.Name, Description: a.Description, Visibility: a.Visibility, Tags: a.Tags, Spec: a.Spec}
		if req.Name == "" {
			req.Name = old.Name
		}
		if req.Description == "" {
			req.Description = old.Description
		}
		if req.Visibility == "" {
			req.Visibility = old.Visibility
		}
		if len(req.Tags) == 0 {
			req.Tags = old.Tags
		}
		t, err := s.buildTree(req, old)
		if err != nil {
			return fail(err)
		}
		t.UpdatedAt = s.stamp()
		if err := s.Store.PutTree(ctx, t); err != nil {
			return fail(fmt.Errorf("save failed"))
		}
		return toolJSON(t)
	case "delete_tree":
		old, err := s.Store.GetTree(ctx, a.ID)
		if err != nil || old.OwnerID != claims.Sub {
			return fail(fmt.Errorf("tree not found among your trees; use list_my_trees for ids"))
		}
		if err := s.Store.DeleteTree(ctx, old.ID); err != nil {
			return fail(fmt.Errorf("delete failed"))
		}
		return toolJSON(map[string]string{"deleted": old.ID})
	case "list_my_scenes":
		sc, err := s.Store.ListScenesByOwner(ctx, claims.Sub)
		if err != nil {
			return fail(fmt.Errorf("lookup failed"))
		}
		return toolJSON(listOf(sc))
	case "list_my_trees":
		ts, err := s.Store.ListTreesByOwner(ctx, claims.Sub)
		if err != nil {
			return fail(fmt.Errorf("lookup failed"))
		}
		return toolJSON(summarize(ts))
	}
	return fail(fmt.Errorf("unknown tool"))
}
