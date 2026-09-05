#!/usr/bin/env bash
# Dispatch one codex (gpt-6-astra) executor run. See SKILL.md in this directory.
#
#   CODEX_OUT_ROOT=<dir under the session scratchpad> dispatch.sh <slug> [extra codex exec args...]
#
# Expects <CODEX_OUT_ROOT>/<slug>/task.md. Writes events.jsonl, last-message.txt
# and stderr.log beside it. Extra arguments pass straight through to `codex exec`
# (e.g. -c model_reasoning_effort='"medium"').
#
# Run this in the background: a real dispatch takes minutes.
set -euo pipefail

slug="${1:?usage: dispatch.sh <slug> [extra codex exec args...]}"
shift

root="${CODEX_OUT_ROOT:?set CODEX_OUT_ROOT to a writable directory outside the repo}"
out="$root/$slug"
[ -f "$out/task.md" ] || { echo "dispatch: missing $out/task.md" >&2; exit 2; }

repo="${CODEX_REPO:-$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../../.." && pwd)}"
cd "$repo"

# codex has no read fence at all (measured 2026-09-05 in this repo, codex-cli
# 0.153.4): the sandbox restricts writes and network, never reads — a sandboxed
# command read /etc/hostname, a sibling repo and apps/api/.env without a murmur.
# codex is a third-party service; the env files below hold the live catalog
# nmk_ key, the OAuth client secret, the image/trust/JWT secrets and — in
# apps/web/.env.prod — production credentials, and a read of any of them lands
# the value in the model's context upstream. So the files are physically absent
# for the run: copied 000-mode under ~/.cache, removed, then restored and
# hash-verified afterwards. The stash is sealed against the sandbox: read
# denied by DAC (mode 000), chmod/unlink denied because $HOME is not a writable
# root. Only an escalated command could touch it, and exec mode auto-denies
# escalation.
#
# Discovery, not a hard-coded list: the repo carries six env files today
# (apps/api/.env, apps/web/.env, apps/web/.env.prod, docker/api.env,
# docker/web.env, refs/legacy/.env.prod) and a hard-coded list rots the day a
# seventh appears.
mapfile -t envfiles < <(find . \
  \( -name node_modules -o -name .git -o -name .nuxt -o -name .output \
     -o -name dist -o -name build -o -name tmp \) -prune -o \
  -type f \( -name '.env' -o -name '.env.*' -o -name '*.env' \) \
  ! -name '*.example' -print | sort)

# Two scrub layers, and only one would do alone: shell_environment_policy
# governs what codex hands to the commands it spawns (default is inherit=all —
# a decoy leaked straight through, measured in the sibling repo), and the
# env -u strip keeps the values out of the codex process itself. The unset list
# is every key NAME in the env files (names only — values are never read here)
# plus the launcher vocabulary from iron rules 13/14.
unset_args=()
while IFS= read -r v; do unset_args+=(-u "$v"); done < <(
  { if [ "${#envfiles[@]}" -gt 0 ]; then
      grep -hoE '^[A-Za-z_][A-Za-z0-9_]*' "${envfiles[@]}" || true
    fi
    env | grep -oE '^(TEST_DATABASE|CODEX_SESSION)_[A-Za-z0-9_]*' || true
  } | sort -u)

stash_dir=""
declare -A env_sha=()
if [ "${#envfiles[@]}" -gt 0 ]; then
  stash_dir="$(mktemp -d "$HOME/.cache/.codex-dispatch-XXXXXXXX")"
  for f in "${envfiles[@]}"; do
    env_sha["$f"]="$(sha256sum "$f" | cut -d' ' -f1)"
    mkdir -p "$stash_dir/$(dirname "$f")"
    cp "$f" "$stash_dir/$f"
    chmod 000 "$stash_dir/$f"
    rm -f "$f"
  done
fi
restore_env() {
  [ -n "$stash_dir" ] || return 0
  local f
  for f in "${!env_sha[@]}"; do
    if [ -e "$stash_dir/$f" ]; then
      chmod 600 "$stash_dir/$f"
      mv -f "$stash_dir/$f" "$repo/$f"
    fi
  done
  # Directories only — a file still here means a failed restore, and rm -rf
  # would destroy the one copy of a secret.
  find "$stash_dir" -depth -type d -empty -delete 2>/dev/null || true
}
trap restore_env EXIT

# -c values are parsed as TOML, so strings need their own quotes.
# mcp_servers={} and plugins={} empty the tables from config.toml: no remote
# MCP server or plugin has a place in a fenced executor run. Never add
# --approve-for-me or --dangerously-bypass-* here — exec's default of
# auto-denying escalation is what makes the sandbox the fence.
codex_args=(exec
  --cd "$repo"
  --sandbox workspace-write
  --add-dir "$out"
  --color never
  --json
  -o "$out/last-message.txt"
  --ignore-rules
  -c 'mcp_servers={}'
  -c 'plugins={}'
  -c 'shell_environment_policy.inherit="core"'
)

code=0
timeout --signal=INT --kill-after=30 "${CODEX_TIMEOUT:-2700}" \
  env "${unset_args[@]}" codex "${codex_args[@]}" "$@" - <"$out/task.md" \
  >"$out/events.jsonl" 2>"$out/stderr.log" || code=$?

restore_env
trap - EXIT
bad=0
for f in "${!env_sha[@]}"; do
  if [ ! -f "$f" ] || [ "$(sha256sum "$f" | cut -d' ' -f1)" != "${env_sha[$f]}" ]; then
    echo "dispatch: $f DID NOT RESTORE CLEANLY — recover from $stash_dir by hand" >&2
    bad=1
  fi
done
[ "$bad" -eq 0 ] || exit 3

if [ "$code" -eq 124 ]; then
  echo "dispatch: timed out after ${CODEX_TIMEOUT:-2700}s; partial events in $out/events.jsonl" >&2
elif [ "$code" -ne 0 ]; then
  echo "dispatch: codex exited $code; see $out/stderr.log" >&2
  tail -c 2000 "$out/stderr.log" >&2 || true
fi

tokens="$(jq -rs '[.[] | select(.type=="turn.completed") | .usage]
  | if length==0 then "?"
    else "in=\([.[].input_tokens]|add) (cached \([.[].cached_input_tokens]|add)) out=\([.[].output_tokens]|add)" end' \
  "$out/events.jsonl" 2>/dev/null || echo '?')"
cmds="$(jq -rs '[.[] | select(.type=="item.completed" and .item.type=="command_execution")] | length' \
  "$out/events.jsonl" 2>/dev/null || echo '?')"
printf 'exit=%s tokens=%s commands=%s out=%s\n' "$code" "$tokens" "$cmds" "$out"

# Every command the executor ran is in the event log; reads are unrestricted, so
# this audit is the read fence's second half. A hit is not proof of a leak — it
# is a line the orchestrator must go read in context.
suspects="$(grep -aoiE 'codex-dispatch|\.ssh|id_rsa|id_ed25519|auth\.json|/proc/[0-9a-z]+/environ|\.env\b' "$out/events.jsonl" 2>/dev/null | sort | uniq -c | sort -rn || true)"
if [ -n "$suspects" ]; then
  echo '--- SUSPECT strings in the event log (go read them in context) ---'
  echo "$suspects"
fi

echo '--- git status (acceptance check one) ---'
git status --porcelain || true

exit "$code"
