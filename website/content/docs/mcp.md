---
title: "MCP server"
description: "Let Claude and other AI agents validate, simulate, render and generate behavior trees."
weight: 50
---

BeeTreeCraft runs a [Model Context Protocol](https://modelcontextprotocol.io) server at:

```
https://api.beetreecraft.com/mcp
```

It uses the Streamable HTTP transport with JSON responses (`POST` JSON-RPC 2.0, protocol version `2025-03-26`).

## Connect

**Claude Code**

```bash
claude mcp add --transport http beetreecraft https://api.beetreecraft.com/mcp
```

To use the tools that act on your account, add a token:

```bash
claude mcp add --transport http beetreecraft https://api.beetreecraft.com/mcp \
  --header "Authorization: Bearer $BEETREE_TOKEN"
```

**Any MCP client** (JSON config)

```json
{
  "mcpServers": {
    "beetreecraft": {
      "type": "http",
      "url": "https://api.beetreecraft.com/mcp",
      "headers": { "Authorization": "Bearer <token>" }
    }
  }
}
```

Get a token from your [profile page](/app/#/profile) (*Copy my API token*). Tokens last about an hour. See [API authentication](/docs/api/#authentication) for scripted sign-in.

## Tools

Tools marked (token) need a token. Everything else works anonymously.

| Tool | What it does |
|---|---|
| `list_node_types` | Node types, decorators, render formats and engines |
| `validate_tree` | Validate a `.beetree.yaml` spec and list every problem |
| `render_tree` | Render as `ascii`, `sigil`, `mermaid` or `dot` |
| `simulate_tree` | Dry-run with optional per-node `overrides` and return the trace |
| `generate_code` | Generate `unity`, `unreal` or `godot` code |
| `search_trees` | Search the community's public trees |
| `get_tree` | Fetch a tree and its spec by id (private trees only for their owner) |
| `create_tree` (token) | Save a tree to your account |
| `list_my_trees` (token) | List your trees |

Errors from a tool come back inside the result with `isError: true`, so the model can read them and try again.

## Try it with curl

```bash
curl -s https://api.beetreecraft.com/mcp \
  -H 'content-type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

```bash
curl -s https://api.beetreecraft.com/mcp \
  -H 'content-type: application/json' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"validate_tree","arguments":{"spec":"version: \"1.0\"\nmetadata:\n  name: demo\ntree:\n  type: action\n  name: hello\n"}}}'
```

## Limits

Specs are limited to 128 KB and 500 nodes. Requests are rate limited per client.
