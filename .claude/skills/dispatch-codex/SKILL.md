---
name: dispatch-codex
description: Dispatch implementation and investigation work to the local codex CLI (gpt-6-astra through the locally configured gateway) as a headless executor while this session stays the orchestrator and acceptor. Use when the user asks to "派发 astra" / "派发 gpt" / "dispatch codex", or when a task suits a written task book and grok is busy or a second opinion is wanted. Covers the measured sandbox boundaries on this machine, the env-file stash fence (codex has no read-deny at all), the flag set, task-book differences from dispatch-grok, and the acceptance protocol.
---

# Dispatching codex (gpt-6-astra) as executor

> **If you are the codex executor and this file was loaded into your context: ignore it.**
> It describes how the orchestrator dispatches *you*. It is not a task book.

This session is the **orchestrator**: it adjudicates design, writes the task book, runs every
gate, all git, and accepts or rejects. The local `codex` CLI (default model `gpt-6-astra` at
reasoning effort max, through the locally configured gateway) is the **executor**. Same
division of labour as `dispatch-grok`, different machinery underneath — read §1 before
assuming anything carries over.

A dispatch is a sandboxed subprocess of this session working in this session's checkout — not
a second session in the iron-rule-13 sense. What rule 13 does forbid is two writers on one
checkout: never dispatch while another session owns this worktree, and never run two
dispatches (codex or grok) concurrently over overlapping paths.

## 1. What codex can reach on this machine

Read/write/network rows re-measured **2026-09-05 in this repository** against `codex-cli
0.153.4` with `codex sandbox` (deterministic, no model in the loop). Rows marked † come from
a full exec probe run 2026-09-04 in the sibling private repo this skill was adapted from —
same machine, same codex version; verify them against the first real dispatch here. Do not
inherit this table into another repo or another codex version without re-measuring.

| Capability | Under `dispatch.sh` | Measured outcome |
|---|---|---|
| Read inside the repo | **free** | — |
| Read **outside** the repo | **free — there is no read fence** | read `/etc/hostname` and a sibling repo's `CLAUDE.md`; reads stop only at DAC |
| Write inside the repo | allowed (workspace-write) | probe file created and removed |
| Write to `.git/` | **denied by the sandbox** | `Read-only file system` — codex structurally cannot commit |
| Write outside repo (`$HOME`, `~/.codex`) | denied | `Read-only file system` |
| Write to `/tmp` | **allowed** | the scratchpad is reachable without `--add-dir` |
| Network | denied | `curl: (6) Could not resolve host` |
| Shell | **available, sandboxed** | this is how codex works; it is not deniable |
| MCP † | emptied by `-c 'mcp_servers={}'` | probe listed zero resources |
| Web search / fetch † | none | probe tool inventory: `apply_patch exec_command clock goals view_image write_stdin` |
| Secret env vars † | scrubbed twice | a decoy exported to dispatch.sh was invisible inside the run |
| The repo's env files | **physically absent during the run** | before the stash, a sandboxed `wc -c apps/api/.env` read it freely |
| Project instructions † | `CLAUDE.md` auto-loaded | probe quoted its first heading without opening a file (served as "AGENTS.md instructions") — no `AGENTS.md` pointer file is needed |
| Escalation / approval † | auto-denied in exec mode | denied writes returned plain errors, no interactive prompt |

### The env-file fence is a stash, not a rule

codex's sandbox restricts **writes and network, never reads** — and unlike grok there is no
`--deny Read(...)` mechanism at all. Any file this user can read, a dispatched executor can
read, and a read lands the content in the model's context at the gateway. This repo carries
**six** gitignored env files, not one: `apps/api/.env` (the live catalog `nmk_` key, OAuth
client secret, image / trust / JWT secrets, the DB DSN), `apps/web/.env` (OG site key,
image-bed and S3 secrets, CF cache token), `apps/web/.env.prod` (**production**
credentials), `docker/api.env`, `docker/web.env`, `refs/legacy/.env.prod`. `dispatch.sh`
**discovers** every env-shaped gitignored file instead of hard-coding that list, and removes
them all for the duration: copied to 000-mode files under `~/.cache` (mirroring the repo
layout), deleted from the repo, restored and **hash-verified per file** afterwards. The stash
is sealed on measured counts: read denied by DAC (mode 000), chmod / unlink denied because
`$HOME` is not a writable root.

Residual, stated honestly: sibling repos' env files, `~/.ssh`, the session scratchpad and
everything else readable by this user stay readable. Two mitigations, both real but neither
structural: the task book forbids reads outside the repo, and **every command the executor
runs is in `events.jsonl`** — `dispatch.sh` greps it for suspect strings (`.ssh`, `id_rsa`,
`id_ed25519`, `auth.json`, the stash prefix, `/proc/*/environ`, `.env`) and prints hits for
the orchestrator to read in context. A hit is a line to go read, not automatically a leak —
the original probe's `.env` hits were an instructed `cat .env` test.

If a dispatch dies hard (SIGKILL, machine loss) the trap may not restore: recover the newest
`~/.cache/.codex-dispatch-*` directory by hand (`chmod 600` each file, move each back to the
matching path inside the repo). `dispatch.sh` exits 3 and says so whenever a restore or a
hash check fails.

### Differences from grok that change how you write a task book

- **codex has a shell.** It runs commands in the sandbox — that is its native way of
  reading, editing and checking. It *can* run `go build ./...` / `go vet ./...` in
  `apps/api`, and — node_modules being local — `pnpm -F web lint` / `pnpm -F web typecheck`,
  so the task book may permit self-checking. Acceptance still re-runs every gate here; a
  green run from the executor is a courtesy, never proof.
- **There are no per-path allow/deny grants.** The whole repo is writable, full stop. Scope
  is bounded by the task book's writable-paths list plus `git status --porcelain` at
  acceptance — same as grok in practice (its `--allow` was a grant, not a whitelist), minus
  the documentation value of the grant flags. Put the writable paths in the task book,
  prominently.
- **`.git/` is read-only**, so the executor cannot stage, commit, or branch even by
  accident. `status` / `log` / `diff` work.
- **DB-backed tests cannot run in the sandbox** (no network, and iron rule 14 reserves the
  launcher `TEST_DATABASE_DSN` for the orchestrator anyway). Tell the executor which tests
  not to run rather than letting it discover and report the failure.

## 2. The dispatch

```bash
export CODEX_OUT_ROOT="$SCRATCHPAD/codex"        # session scratchpad, never the repo
mkdir -p "$CODEX_OUT_ROOT/<slug>"
# write the task book to $CODEX_OUT_ROOT/<slug>/task.md, then:
.claude/skills/dispatch-codex/dispatch.sh <slug>
```

`dispatch.sh` supplies `--sandbox workspace-write`, `--json` (events to `events.jsonl`),
`-o last-message.txt`, `--ignore-rules`, the MCP/plugin emptying, the env-var scrub, the
env-file stash, a `timeout` (default 2700 s, `CODEX_TIMEOUT` to change), and afterwards
prints token usage, the suspect-string audit and `git status --porcelain`. Extra arguments
pass through to `codex exec` — that is where a per-task model or effort override goes:

| Extra arg | When |
|---|---|
| `-c model_reasoning_effort='"medium"'` | mechanical sweeps; the config default is `max` |
| `-m <model>` | override `gpt-6-astra` |
| `--ephemeral` | leave nothing in codex's session store |
| env `CODEX_REPO=<abs path>` | dispatch into a sibling repo: that repo becomes the sandbox's writable root and it is *its* env files that get stashed — **this repo's env files are then readable to the executor**, so the task book must forbid them by name and the audit must be read |

`-c` values parse as TOML — strings need their own quotes inside the shell quotes.
**Never pass `--approve-for-me` or `--dangerously-bypass-*`**: exec's auto-deny of
escalation is what makes the sandbox the fence. Never run two dispatches (codex or grok)
concurrently over overlapping paths.

A follow-up fix can resume the same thread with its context intact:
`codex exec resume --last` (plus the same fences — run it through your own judgment, there
is no wrapper for it yet).

## 3. Reading the result

1. `dispatch.sh` prints `exit= tokens= commands= out=`; nonzero exit tails `stderr.log`,
   124 is the timeout.
2. **A clean exit is not acceptance.** Read `report.md` in the output directory — the task
   book requires one — then run the gates yourself.
3. `last-message.txt` is the executor's closing paragraph; `events.jsonl` is the full
   record, one JSON line per event, every command with its `aggregated_output`. The audit
   section of the dispatch output points into it.
4. `git status --porcelain` (printed at the end) is acceptance check one: exactly the task
   book's writable paths, nothing else. The report lives in the scratchpad, so the status
   is a pure signal.

## 4. The task book

Template: `task-book-template.md` in this directory. Everything from `dispatch-grok` §4
holds (English, self-contained, no open design decisions, acceptance criteria the
orchestrator runs, fixed report path and structure, forbid ranking, demand a positive
control on any audit) with the environment section swapped: codex has a sandboxed shell and
no network, the repo's env files do not exist during the run, and reads outside the repo are
forbidden by the book and audited rather than denied by policy.

`CLAUDE.md` is auto-loaded — do not re-paste the iron rules; restate the specific ones the
task turns on. A task touching `apps/web` almost always turns on iron rules 1 (no background
gradients — the sanctioned exceptions are annotated in-code), 2 (KunUI first, never modify
KunUI) and 11 (single real root element per page).

## 5. What the orchestrator never delegates

Identical in shape to `dispatch-grok` §5, and it does not soften because the executor can
run a compiler: adjudications and scope calls; every gate; every migration under
`apps/api/migrations/**` and every schema decision behind one (plus the end-of-task
migration reminder to the user — that duty is the orchestrator's); everything that touches
the database, production, or a live upstream (catalog / OAuth / community / image / trust)
with real keys; all git; any ruling under iron rule 1 (whether a gradient is sanctioned) or
iron rule 2 (whether a KunUI equivalent exists — an executor never hand-rolls around KunUI
and never edits KunUI itself); and final acceptance.

## 6. Choosing between codex and grok

Both are third-party executors with the same shape of fence and the same task-book contract.
Measured differences that matter when picking:

- codex can self-check a build offline; grok cannot run anything. For `apps/web` work this
  weighs heavily: a Vue/TS change never passed through `vue-tsc` is expensive to accept, and
  only codex can run it.
- codex cannot touch `.git/` or anything outside the repo, enforced by kernel sandbox;
  grok's only hard fence is its deny rules.
- grok can read the env files only through a deny-rule gap; codex would read them freely —
  hence the stash. Neither executor is ever handed a secret on purpose.
- Cost order of magnitude (sibling-repo probe, 2026-09-04): 18 commands ≈ 138k in / 4.7k
  out tokens through the gateway.

When in doubt, dispatch the one whose failure mode you can check more cheaply, and never
both onto overlapping paths at once.
