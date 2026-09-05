#!/usr/bin/env bash
# Dispatch one grok executor run. See SKILL.md in this directory.
#
#   GROK_OUT_ROOT=<dir under /tmp> dispatch.sh <slug> [extra grok args...]
#
# Expects <GROK_OUT_ROOT>/<slug>/task.md to exist. Writes run.json, stderr.log and
# debug.log beside it. Extra arguments are passed straight through to grok — that is
# where the per-task --allow 'Write(<glob>)' / 'Edit(<glob>)' rules go.
#
# Run this in the background: a real dispatch takes minutes.
set -euo pipefail

slug="${1:?usage: dispatch.sh <slug> [extra grok args...]}"
shift

root="${GROK_OUT_ROOT:?set GROK_OUT_ROOT to a writable directory under /tmp}"
out="$root/$slug"
[ -f "$out/task.md" ] || { echo "dispatch: missing $out/task.md" >&2; exit 2; }

# The deny rules below are the fence, not a courtesy. Measured on this machine
# (grok 1.0.13; machine pins re-checked 2026-09-05): /etc/grok/requirements.toml
# pins no sandbox profile, so the executor reads sibling repositories freely,
# and ~/.grok/config.toml sets permission_mode = "always-approve". Nothing but
# these rules stands between a task book and the shell, an MCP server, or the
# env files.
#
# The env files matter most here: apps/api/.env holds the live catalog nmk_
# key, the OAuth client secret and the DB DSN; apps/web/.env.prod holds
# production credentials; docker/*.env and refs/legacy/.env.prod round out six
# files. grok is a third-party service — a read of any of them is an
# exfiltration. Five patterns cover all six files; **/.env.* is the one the
# sibling repo's four-pattern set was missing (it left .env.prod readable). A
# denied read answers `Denied by permission policy: deny rule on read matching
# "..."` and names the pattern that fired, so a miss is visible, not silent.
grok --prompt-file "$out/task.md" \
     --allow "Write($out/**)" \
     --allow "Read($out/**)" \
     --deny 'Bash(*)' \
     --deny 'mcp__*' \
     --deny 'Read(.env)' \
     --deny 'Read(.env.*)' \
     --deny 'Read(**/.env)' \
     --deny 'Read(**/.env.*)' \
     --deny 'Read(**/*.env)' \
     --output-format json \
     --max-turns "${GROK_MAX_TURNS:-200}" \
     --debug-file "$out/debug.log" \
     "$@" \
     >"$out/run.json" 2>"$out/stderr.log" || true

if [ ! -s "$out/run.json" ]; then
  echo "dispatch: grok produced no JSON; see $out/stderr.log" >&2
  tail -c 2000 "$out/stderr.log" >&2 || true
  exit 1
fi

stop=$(jq -r '.stopReason // "?"' "$out/run.json")
turns=$(jq -r '.num_turns // "?"' "$out/run.json")
cost=$(jq -r '.total_cost_usd // "?"' "$out/run.json")
printf 'stopReason=%s turns=%s cost_usd=%s out=%s\n' "$stop" "$turns" "$cost" "$out"

[ -s "$out/stderr.log" ] && { echo '--- stderr ---' >&2; tail -c 2000 "$out/stderr.log" >&2; }

if [ "$stop" != "end_turn" ]; then
  # stopReason "cancelled" covers a denied tool call, a tool that fell through to
  # an ask floor, and turn exhaustion. Only the debug log separates them.
  if rg -q 'PermissionCancelled' "$out/debug.log" 2>/dev/null; then
    echo 'dispatch: a tool call was CANCELLED — an ask floor, not one of the deny rules' >&2
    rg -o 'permission policy: .*tool="[^"]*"' "$out/debug.log" | sort -u >&2 || true
  elif [ "$turns" -ge "${GROK_MAX_TURNS:-200}" ] 2>/dev/null; then
    echo 'dispatch: turn budget exhausted — raise GROK_MAX_TURNS or split the task' >&2
  else
    echo "dispatch: run ended as '$stop'; inspect $out/debug.log" >&2
  fi
  exit 1
fi
