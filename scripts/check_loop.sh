#!/usr/bin/env bash
# Guard the loop's prompt-injection defenses: only the trusted author's issues reach the agent.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
grep -q -- '--author "$TRUSTED"' "$ROOT/scripts/agent-loop.sh" || { echo "agent-loop.sh must fetch the backlog with --author" >&2; exit 1; }
grep -q 'untrusted data' "$ROOT/scripts/loop-prompt.md" || { echo "loop-prompt.md needs the Trust section" >&2; exit 1; }
if grep -nE 'gh issue (list|view)' "$ROOT/scripts/loop-prompt.md" | grep -v -- '--author' | grep -vE 'list no issues|--json title,body,author'; then
  echo "loop-prompt.md must not tell the agent to read issues without an author filter" >&2; exit 1
fi
echo "loop trust checks ok"
