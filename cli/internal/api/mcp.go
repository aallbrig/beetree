package api

import (
	"encoding/json"
	"fmt"
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
	// needsAuth tools act on the caller's account.
	needsAuth bool
}

func objSchema(required []string, props map[string]any) map[string]any {
	return map[string]any{"type": "object", "properties": props, "required": required}
}

var strProp = func(desc string) map[string]any { return map[string]any{"type": "string", "description": desc} }

func mcpTools() []mcpTool {
	spec := strProp("Behavior tree specification as .beetree.yaml text")
	return []mcpTool{
		{Name: "list_node_types", Description: "List the node types, decorators, render formats and code generation engines BeeTree supports.", InputSchema: objSchema(nil, map[string]any{})},
		{Name: "validate_tree", Description: "Validate a .beetree.yaml spec. Returns whether it is valid and every error found.", InputSchema: objSchema([]string{"spec"}, map[string]any{"spec": spec})},
		{Name: "render_tree", Description: "Render a spec as ascii, sigil, mermaid or dot.", InputSchema: objSchema([]string{"spec"}, map[string]any{"spec": spec, "format": strProp("ascii | sigil | mermaid | dot (default ascii)")})},
		{Name: "simulate_tree", Description: "Dry-run a spec and return the execution trace. Overrides force named nodes to SUCCESS, FAILURE or RUNNING.", InputSchema: objSchema([]string{"spec"}, map[string]any{"spec": spec, "overrides": map[string]any{"type": "object", "description": "node name -> SUCCESS|FAILURE|RUNNING", "additionalProperties": map[string]any{"type": "string"}}})},
		{Name: "generate_code", Description: "Generate engine code (unity, unreal or godot) from a spec.", InputSchema: objSchema([]string{"spec", "engine"}, map[string]any{"spec": spec, "engine": strProp("unity | unreal | godot")})},
		{Name: "search_trees", Description: "Search public behavior trees shared by the community.", InputSchema: objSchema(nil, map[string]any{"query": strProp("text to match against name, description and tags"), "tag": strProp("exact tag"), "limit": map[string]any{"type": "integer"}})},
		{Name: "get_tree", Description: "Fetch a tree, including its spec, by id. Private trees need the owner's token.", InputSchema: objSchema([]string{"id"}, map[string]any{"id": strProp("tree id")})},
		{Name: "create_tree", Description: "Save a new tree to the signed-in user's account. Requires a bearer token.", needsAuth: true, InputSchema: objSchema([]string{"spec"}, map[string]any{"spec": spec, "name": strProp("display name (defaults to spec metadata.name)"), "description": strProp("short description"), "visibility": strProp("public | private (default private)"), "tags": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}})},
		{Name: "list_my_trees", Description: "List the signed-in user's trees. Requires a bearer token.", needsAuth: true, InputSchema: objSchema(nil, map[string]any{})},
	}
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
			"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
			"serverInfo":      map[string]string{"name": "beetreecraft", "version": s.Version},
			"instructions":    "BeeTree authors engine-agnostic behavior trees. Pass specs as .beetree.yaml text. Tools that touch an account need an Authorization: Bearer token.",
		})})
	case "ping":
		writeRPC(w, rpcResponse{ID: req.ID, Result: jsonRaw(map[string]any{})})
	case "tools/list":
		writeRPC(w, rpcResponse{ID: req.ID, Result: jsonRaw(map[string]any{"tools": mcpTools()})})
	case "tools/call":
		s.mcpCall(w, r, req)
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
	b, _ := json.MarshalIndent(v, "", "  ")
	return toolText(string(b), false)
}

type callParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type toolArgs struct {
	Spec        string            `json:"spec"`
	Format      string            `json:"format"`
	Engine      string            `json:"engine"`
	Overrides   map[string]string `json:"overrides"`
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
		ts, err := s.Store.ListPublicTrees(ctx, 200)
		if err != nil {
			return fail(fmt.Errorf("search failed"))
		}
		limit := a.Limit
		if limit <= 0 || limit > 50 {
			limit = 20
		}
		q, tag := strings.ToLower(a.Query), strings.ToLower(a.Tag)
		var out []Tree
		for _, t := range ts {
			if q != "" && !strings.Contains(strings.ToLower(t.Name+" "+t.Description+" "+strings.Join(t.Tags, " ")), q) {
				continue
			}
			if tag != "" && !containsStr(t.Tags, tag) {
				continue
			}
			if out = append(out, t); len(out) == limit {
				break
			}
		}
		return toolJSON(summarize(out))
	case "get_tree":
		t, err := s.Store.GetTree(ctx, a.ID)
		if err != nil || !canRead(r, t) {
			return fail(fmt.Errorf("tree not found"))
		}
		return toolJSON(t)
	case "create_tree":
		t, err := s.createTreeFor(r, claims, treeRequest{Name: a.Name, Description: a.Description, Visibility: a.Visibility, Tags: a.Tags, Spec: a.Spec})
		if err != nil {
			return fail(err)
		}
		return toolJSON(t)
	case "list_my_trees":
		ts, err := s.Store.ListTreesByOwner(ctx, claims.Sub)
		if err != nil {
			return fail(fmt.Errorf("lookup failed"))
		}
		return toolJSON(summarize(ts))
	}
	return fail(fmt.Errorf("unknown tool"))
}
