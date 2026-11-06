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

The server needs an [API key](/docs/api/#get-an-api-key). Create one on your [profile page](/app/#/profile), keep it in an environment variable, then:

```bash
claude mcp add --transport http beetreecraft https://api.beetreecraft.com/mcp \
  --header "X-API-Key: $BEETREE_API_KEY"
```

**Any MCP client** (JSON config)

```json
{
  "mcpServers": {
    "beetreecraft": {
      "type": "http",
      "url": "https://api.beetreecraft.com/mcp",
      "headers": { "X-API-Key": "<your API key>" }
    }
  }
}
```

Get a key from your [profile page](/app/#/profile) under **API keys**. Keys start with `btc_`, are shown once, and can be revoked there. Requests without a valid key get a 401. Never commit a key; use an environment variable or your client's secret store.

## Tools

Every tool needs a valid API key. Tools marked (account) also act on your account.

| Tool | What it does |
|---|---|
| `list_node_types` | Node types, decorators, render formats and engines |
| `validate_tree` | Validate a `.beetree.yaml` spec and list every problem |
| `render_tree` | Render as `ascii`, `sigil`, `mermaid` or `dot` |
| `simulate_start`, `simulate_answer` | Step through a simulation: start returns the first pending leaf; answer submits `SUCCESS`, `FAILURE` or `RUNNING` and returns the next leaf or the final trace |
| `diff_trees` | Compare two specs (`spec`, `other`) and list added, removed and modified nodes by path |
| `simulate_tree` | Dry-run with optional per-node `overrides` and return the trace |
| `generate_code` | Generate `unity`, `unreal` or `godot` code |
| `search_trees` | Search the community's public trees. Returns `{trees, truncated}`; `truncated: true` means only the newest 2000 public trees were scanned, so narrow the query or tag |
| `get_tree` | Fetch a tree and its spec by id (private trees only for their owner) |
| `create_tree` (account) | Save a tree to your account |
| `update_tree` (account) | Replace the spec of a tree you own; omitted fields keep their values |
| `delete_tree` (account) | Permanently delete a tree you own |
| `list_my_scenes` (account) | List your Explore scenes |
| `list_my_trees` (account) | List your trees |

Errors from a tool come back inside the result with `isError: true`, so the model can read them and try again.

## Resources and prompts

Resources (`resources/list`, `resources/read`):

| URI | Contents |
|---|---|
| `beetree://node-types` | Node types, decorators, render formats and engines as JSON |
| `beetree://schema/beetree` | The JSON Schema for `.beetree.yaml`, identical to `/schemas/beetree.schema.json` |
| `beetree://examples/patrol` | A small valid `.beetree.yaml` to copy the format from |
| `beetree://templates/<name>` | The builder templates (shopkeeper, guard, coward, boss, companion), each a valid `.beetree.yaml` |

Prompt `design_behavior_tree` (`prompts/get`) takes `character` and an optional `engine`. It walks a model through reading the resources, writing a spec, validating, simulating, rendering and generating code.

## Try it with curl

Both calls send your key in `X-API-Key`; without it the server answers 401.

```bash
curl -s https://api.beetreecraft.com/mcp \
  -H "X-API-Key: $BEETREE_API_KEY" \
  -H 'content-type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

```bash
curl -s https://api.beetreecraft.com/mcp \
  -H "X-API-Key: $BEETREE_API_KEY" \
  -H 'content-type: application/json' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"validate_tree","arguments":{"spec":"version: \"1.0\"\nmetadata:\n  name: demo\ntree:\n  type: action\n  name: hello\n"}}}'
```

## Limits

Specs are limited to 128 KB and 500 nodes. Requests are rate limited per client.
