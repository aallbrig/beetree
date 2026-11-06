---
title: "REST API"
description: "Create, validate, share and discuss behavior trees over HTTPS."
weight: 40
---

Base URL: `https://api.beetreecraft.com`

- **[Interactive reference (Swagger UI)](/docs/api/reference/)**: try calls in your browser
- **[OpenAPI 3 document](https://api.beetreecraft.com/v1/openapi.json)**: generate a client in any language

Everything the website does is an API call. Specs travel as `.beetree.yaml` text inside JSON.

## Get an API key

The API and the MCP server need an account. Every request carries an **API key**.

1. [Sign up](/app/#/signup) and confirm your email, then sign in.
2. Open your [profile](/app/#/profile), find **API keys**, give the key a name and press **Create API key**.
3. Copy the key now. It starts with `btc_` and is shown **once**; BeeTreeCraft stores only a hash of it.
4. Keep it out of source control: put it in an environment variable or a secrets manager.

```bash
export BEETREE_API_KEY="btc_<paste your key here>"
```

You can have up to 5 keys. Revoke one on the same page if it leaks; it stops working immediately. Only a signed-in session can create or revoke keys, so a leaked key cannot mint new ones.

Send the key in either header:

```bash
curl -s https://api.beetreecraft.com/v1/me -H "X-API-Key: $BEETREE_API_KEY"
curl -s https://api.beetreecraft.com/v1/me -H "Authorization: Bearer $BEETREE_API_KEY"
```

`GET /v1/health` and the OpenAPI document need no key.

## Quick start

Validate a tree:

```bash
curl -s https://api.beetreecraft.com/v1/validate \
  -H "X-API-Key: $BEETREE_API_KEY" \
  -H 'content-type: application/json' \
  -d '{"spec":"version: \"1.0\"\nmetadata:\n  name: demo\ntree:\n  type: action\n  name: hello\n"}'
```

Render, simulate and generate use the same body shape:

| Endpoint | Extra field | Returns |
|---|---|---|
| `POST /v1/validate` | | `{valid, errors, nodeCount}` |
| `POST /v1/render` | `format`: `ascii` `sigil` `mermaid` `dot` | `{output}` |
| `POST /v1/simulate` | `overrides`: `{"node_name": "FAILURE"}` | `{status, trace, steps}` |
| `POST /v1/generate` | `engine`: `unity` `unreal` `godot` | `{files: [{path, content, isStub}]}` |

Browse public trees: `GET /v1/trees?q=patrol&sort=popular`.

## Authentication

Use an [API key](#get-an-api-key) for scripts, servers and MCP clients. A sign-in token from Amazon Cognito also works, lasts about an hour, and is the only credential that can manage API keys.

**A short-lived token from a script:**

```bash
aws cognito-idp initiate-auth --region us-east-1 \
  --client-id 7i6a8ab6aidb93b4g91sisa2ev \
  --auth-flow USER_PASSWORD_AUTH \
  --auth-parameters USERNAME=<you>,PASSWORD=<password> \
  --query AuthenticationResult.IdToken --output text
```

Tokens last about an hour. Use the returned `RefreshToken` with the `REFRESH_TOKEN_AUTH` flow to get a new one.

```bash
curl -s https://api.beetreecraft.com/v1/me -H "Authorization: Bearer $TOKEN"
```

Manage keys over the API with a sign-in token (a key cannot do this):

```bash
curl -s https://api.beetreecraft.com/v1/keys -H "Authorization: Bearer $TOKEN"
curl -s -X POST https://api.beetreecraft.com/v1/keys -H "Authorization: Bearer $TOKEN" \
  -H 'content-type: application/json' -d '{"name":"laptop"}'
curl -s -X DELETE https://api.beetreecraft.com/v1/keys/<id> -H "Authorization: Bearer $TOKEN"
```

## Save and share a tree

```bash
curl -s https://api.beetreecraft.com/v1/trees \
  -H "X-API-Key: $BEETREE_API_KEY" -H 'content-type: application/json' \
  -d '{"name":"My guard","visibility":"public","tags":["enemy"],"spec":"..."}'
```

Trees are **private by default**. A private tree is invisible, even by id, to everyone but its owner.

## Social

| Action | Request |
|---|---|
| Like / unlike | `PUT` / `DELETE /v1/trees/{id}/like` (idempotent) |
| Comment | `POST /v1/trees/{id}/comments` with `{"body": "..."}` |
| Forum | `GET/POST /v1/forum/threads`, `POST /v1/forum/threads/{id}/posts` |
| Your stats | `GET /v1/me/stats` |

## Errors

Errors are JSON: `{"error": {"code": "invalid_spec", "message": "..."}}`.

| Status | Meaning |
|---|---|
| 400 | Bad request body |
| 401 | Missing, invalid, revoked or expired key or token |
| 403 | Not allowed (for example deleting someone else's comment, or managing API keys with an API key) |
| 404 | Not found, or private and not yours |
| 413 | Request body too large (512 KB limit) |
| 422 | The spec is invalid; the message lists why |

## Limits

Specs are limited to 128 KB and 500 nodes. Requests are rate limited. Browsers may call the API from any origin (CORS is open) because credentials travel in the `Authorization` header, never cookies.

## MCP

Prefer to let an AI agent drive? See the [MCP server](/docs/mcp/).
