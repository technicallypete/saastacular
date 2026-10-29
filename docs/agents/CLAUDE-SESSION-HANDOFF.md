# Codex ↔ Claude Code session handoffs

Operational note, verified locally on 2026-10-06 with Claude Code 2.1.290/2.1.291 and reverified on 2026-10-10 with 2.1.296 on Linux/WSL. Codex uses shell tools to discover, inspect and message Claude Code sessions. Codex's own collaboration tools address Codex agents; they do not address an existing Claude Code session.

This runbook records the existing handoff mechanism, not a new execution framework. It does not authorize extra implementation, permission changes, session restarts, commits or external actions. Follow [AGENTS.md](../../AGENTS.md) and the current Goal's approved scope/budget/binding.

## Find the intended session

From the repository root:

```sh
claude agents --json --all --cwd "$PWD"
```

Match session name, full session UUID and checkout path. Several sessions may belong to this project; some are stopped/completed or work in separate worktrees. Do not pick the first entry or reuse a remembered PID. The short background ID is useful for CLI commands; the full UUID identifies the conversation.

Inspect the chosen registration at `~/.claude/sessions/<pid>.json`. Relevant fields include `pid`, `sessionId`, `cwd`, `name`, `status`, `version`, `peerProtocol`, `procStart` and `messagingSocketPath`. After a restart, rediscover them; the process/socket may change while the conversation UUID is retained.

Check `kind` in the CLI listing. A stopped background entry and a live
interactive entry can share the same UUID; select the verified live process.
`claude attach` and `claude logs` operate on background sessions, not an
interactive instance already running in another terminal. For interactive
sessions, use the verified inbox or the owner's existing terminal. Do not
start a duplicate resume. Missing host process visibility (for example, a
restricted PID namespace) prevents verification and does not prove shutdown.

## 2026-10-10 compatibility verification

Installed CLI 2.1.296 still has no native `agents send` command. Its registered
protocol-1 inbox accepted an authenticated, peer-verified G051 handoff
(handoff ID redacted); transcript replay and the recipient
acknowledgment were verified. This supports the same narrow helper protocol,
not a claim of a new public messaging API. That exact-version restriction is superseded for transport by the compatibility policy below.

## Check status and progress

Try `claude logs <background-id>` for recent output. If the local daemon/log bridge fails, read the session's JSONL transcript under `~/.claude/projects/<encoded-checkout>/<session-uuid>.jsonl`. Find it by exact UUID and confirm its checkout; do not guess the directory encoding. Subagent transcripts live under `<session-uuid>/subagents/`.

Read a bounded tail and extract assistant text, timestamps and tool descriptions. Avoid printing raw tool inputs/results, environment dumps, credential files or entire transcripts. Combine session status with actual test/run logs and Result existence; `busy` or an updated transcript alone does not prove a gate passed. A monitor expiry is not a test failure or completion signal.

Prefer user-requested checks or checks at handoffs. Continuous polling is unnecessary. If the session is running tests, inspect progress without launching competing stacks or editing its checkout.

Follow AGENTS.md's **Context and token discipline**: filter terminal control
sequences/redraws before returning output, inspect only relevant bounded excerpts,
and reference the committed review and exact binding in concise handoffs instead
of repeating settled controls. Keep receipt/dispatch verification; reduced output
is not permission to infer acknowledgment or completion.

### Waiting-state prompt gap — verified 2026-10-12

The coordinator reported `waiting` while bounded parent/worker transcripts showed
no unresolved question or tool. Attaching its existing terminal revealed a G051
cleanup confirmation dialog. The operator's existing cleanup authorization
supported selecting "Yes, remove it"; its native response was verified afterward.

When a status check reports `waiting`, inspect the verified existing background
terminal before claiming there are no open questions. Use the existing terminal
for an interactive session. If inspection is unavailable, report prompt state as
unverified. Queued inbox delivery is not a dialog response. Answer only within
recorded owner authority via the accepted native procedure, verify the response
and detach without stopping the worker; unresolved decisions remain with the
owner. This adds prompt inspection, not automatic approval or settings changes.

## Send instructions to a running session

In the locally verified CLI, `claude agents` discovers/manages sessions but has no documented `send` subcommand. We used Claude Code's authenticated local Unix-socket inbox instead of starting another copy of the running conversation.

**This inbox is an internal, protocol-sensitive interface.** Verify the current live registration, protocol, key and authenticated peer before messaging; release-version differences warn rather than gate transport. Do not assume the wire format, key layout or socket namespace is a stable public API. If it is unavailable or incompatible, use `claude attach <id>` for a background session or the owner's existing interactive terminal, or report that the handoff could not be delivered. Do not write a fabricated user entry into the transcript or job files.

Verified protocol 1 on this host:

For routine use, prefer the repository skill's
[Python helper](../../.agents/skills/restruct-sdlc/claude-session-handoff/scripts/handoff.py).
Its `check`, `send` and `receipt` commands produce compact JSON without replaying
the example below into each conversation. `send --document management/...`
references root management records by path/hash and appends delivery metadata to
`management/SESSION-HANDOFFS.jsonl`, a machine-local gitignored/untracked
file. Keep it locally; commit shared decisions/outcomes in management records,
not delivery metadata. Alternate delivery logs must also remain local/untracked.
This log is not approval, acceptance or
execution evidence; canonical reviews/Results/PROGRESS remain authoritative.
The helper never commits or silently refreshes active inputs. Read the skill for
command examples and version/permission boundaries.

1. Connect to the target registration's `messagingSocketPath`, normally `/run/user/<uid>/cc-socks/<pid>.sock` (another local socket location may be published).
2. Derive the local key filename as `<pid>.<sha256(canonical-socket-path)>.key` under `~/.claude/sessions/`. Read its `peerToken` directly into memory.
3. Send newline-delimited JSON: first `{"type":"auth","token":...}`, then a `type: user` frame with the exact target `session_id` and string `message.content`.
4. Use `priority: next` for routine supervisor handoffs so instructions are queued without requesting immediate interruption. Include a unique `uuid` and identify the sender as `codex-supervisor`; that label does not replace the socket authentication or repository approval record.

The following example sends one text file without putting the token into a shell argument or output. Replace the PID/UUID/path after discovery. It requires local Python only for the supervisor handoff; it does not change the application's Docker-first runtime policy.

```sh
python3 - TARGET_PID TARGET_SESSION_UUID /absolute/outside-checkout/handoff.txt <<'PY'
from pathlib import Path
import hashlib
import json
import os
import socket
import sys
import time
import uuid

pid = int(sys.argv[1])
expected_session = sys.argv[2]
content = Path(sys.argv[3]).read_text()
assert content.strip(), "Empty handoff"
registry = Path.home() / ".claude" / "sessions"
registration = json.loads((registry / f"{pid}.json").read_text())
assert registration["pid"] == pid
assert registration["sessionId"] == expected_session
assert registration["peerProtocol"] == 1, "Reverify changed protocol"
os.kill(pid, 0)  # Existence check only; sends no signal.
linux_start = Path(f"/proc/{pid}/stat").read_text().rsplit(")", 1)[1].split()[19]
assert str(registration["procStart"]) == linux_start, "Stale/reused PID"
address = registration["messagingSocketPath"]
assert address.startswith("/") and address.endswith(".sock")
assert os.path.realpath(address) == address, "Reverify noncanonical socket path"
key_path = registry / f"{pid}.{hashlib.sha256(address.encode()).hexdigest()}.key"
token = json.loads(key_path.read_text())["peerToken"]
frame = {
    "type": "user",
    "session_id": expected_session,
    "uuid": str(uuid.uuid4()),
    "from": "codex-supervisor",
    "priority": "next",
    "message": {"role": "user", "content": content},
}
with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as client:
    client.settimeout(5)
    client.connect(address)
    client.sendall((json.dumps({"type": "auth", "token": token}) + "\n"
                    + json.dumps(frame) + "\n").encode())
    time.sleep(0.4)  # Allow consumption; this is not an acknowledgment.
print("Sent to the local inbox; Claude acknowledgment is still required.")
PY
```

Never print/copy/commit the key or token. Do not include them in handoff files, task prompts, logs or evidence. They authenticate the local inbox and are distinct from GitLab/provider credentials. Use this only for the user's authorized local session communication; never redirect the secret or handoff to a remote socket/service.

## Confirm receipt and dispatch

A successful socket write proves only that bytes were sent. It does **not** prove acceptance into the queue, model consumption, subagent creation or task completion. Look for the handoff's user replay or Claude acknowledgment in the target transcript. Until then, report “sent; acknowledgment pending.” Do not repeatedly resend just because acknowledgment is delayed during a tool/test run; avoid duplicate tasks.

For delegation, require Claude to report the child ID, configured effort, isolated checkout and output path. Planning defaults to `xhigh`; implementation defaults to `high`, unless the owner explicitly chooses another setting. Asking for an effort in text is not evidence that it was configured. For a separate CLI agent, use the matching supported `--effort` option; for native subagents verify configured or inherited effort where possible. Report unverifiable effort as unknown. Only say the child started after verifying dispatch.

Completion requires the promised repository artifact, attributable validation where applicable, and a handback. A finished background process is not Codex Plan approval or implementation ACCEPT.

## What a supervisor handoff contains

For [owner mandates and stops](../../AGENTS.md#bounded-specification-authority-and-owner-stop-boundaries),
include the owner instruction/reference, exact task set/endpoint, allowed phases
and resume/hold disposition. Check limits even with existing Specs/Plans and
before lookahead planning. Verify acknowledgment and safe pauses; a queued
message, merge or observer notice never expands or resumes owner authority.

For provisional planning at initial implementation-MR creation (drafts included;
not planning-only/placeholder/probe MRs), cite the standing
[automatic-planning rule](../../AGENTS.md#automatic-planning-after-integration),
the existing Goal/Spec, verified MR/source/target/candidate, assigned planner and
pending gates. Do not wait for CI/Duo, ready status or merge to dispatch planning.
No extra owner prompt is needed for eligible planning; no execution release is
implied. Verify acknowledgment and task ownership through this same route.

For [automatic execution release](../../AGENTS.md#automatic-reassessment-and-execution-release),
cite reviewed predecessor integration and pending staging qualification, the
reconciled Plan, exact input/release record, scope/budget and task owner. Isolated
implementation may overlap predecessor CI; final predecessor ACCEPT still gates
successor integration/final acceptance. Explicit holds and affected security/
architecture/contract gaps remain stops. A recorded supervisor release removes the
routine owner-proceed prompt; a merge or watcher notice does not. Preserve
specific owner holds and native permissions, and verify acknowledgment/dispatch.

- Goal and exact Goal/Spec/Plan/ADR paths/revisions; baseline commit or approved snapshot/digest where execution is authorized.
- Whether the task is read-only research, planning, approved implementation or a bounded correction; exact allowed writes/output path.
- Requested effort, delegation boundary, shared budget/checkpoint, required validation and stop conditions.
- Protection of the active checkout, bound inputs, snapshots and developer/Docker resources; separate checkout for unrelated parallel planning.
- Required artifact save/return and acknowledgment; no implied commit/push/hosted authorization.

Write a new Plan into its goal-specific repository path in the designated planning checkout. Do not leave it only in Claude's internal Plan mode directory or chat. A side checkout does not bypass scope/approval. Superseding architecture/progress documents in main are not silently copied into an active execution checkout; reconcile after the accepted phase transition with a fresh binding.

If a stopped session must resume, use the current CLI's `--resume <session-uuid>` (and `--bg` if appropriate) only when authorized. Do not resume an already-running session as a messaging shortcut: the CLI may create a copy, causing duplicate work. Do not stop/restart a live session or change permission mode merely to send a handoff.

## 2026-10-10 resumed-session routing and receipt correction

A live predecessor PID is insufficient to identify the current conversation.
Owner's G007 work moved from one session UUID to a resumed
session with a new UUID (both redacted); the former process remained alive.
The earlier correction sent to the predecessor was unconfirmed. Do not repeat
it there or treat its older acknowledgment as confirmation of the new task.

Rediscover and verify the full UUID, recent task transcript and checkout at each
handoff. The helper surfaces newer live peers; it blocks a send to a potentially
stale target unless `--confirm-target` is supplied after actual task-owner
verification. Newer activity alone does not authorize choosing another session.

`send` now waits for its unique user replay marker, returning `receipt-confirmed`
or `delivery-unconfirmed` (exit 2), never interpreting successful socket writes
as delivery. Receipt scanning is bounded to the latest 2 MiB. Missing replay is
not proof of a permission hold, refusal or shutdown; inspect the actual state.
No retries, permission bypass or duplicate session starts are automatic.
After receipt, verify assistant acknowledgment and actual task dispatch. Read
replies through the exact recipient's bounded transcript; Claude need not find
a Codex peer in its own session listing for Codex to inspect those replies.

A fresh confirmation handoff (ID redacted) to the
current session replayed immediately. Its 17:12 UTC assistant reply confirms one
G007 correction planner at xhigh, preserving the existing r0 worktree/history,
addressing C1 and PG-001. This verifies communication with the current session,
not Plan approval, implementation release or completion.

## Answering pending routine questions — verified 2026-10-10

Queued inbox messages do not answer an open AskUserQuestion dialog. For an
owner-authorized routine answer in a verified background session, attach its
existing terminal with `claude attach <id>` through a PTY, inspect the displayed
question and select the authorized option. Verify the actual question tool-result
in its transcript; then detach with Ctrl+Z, leaving the original worker running.
Do not create another resume or kill/restart the session.

This was verified for a live coordinator session (ID redacted): the owner chose the isolated-agent naming
fallback, Codex selected that option, and the tool-result recorded exactly
"Allow agent-<id> name". Socket delivery alone had left the question unanswered.
Use this for bounded choices already authorized by the owner; actual permission,
architecture/scope and unapproved external-action decisions retain their gates.
No settings edit, transcript fabrication or blanket permission answer is allowed.

### Owner authorization attribution — 2026-10-12

When relaying an owner-authorized action or approval, attribute the decision to
the owner and the delivery to Codex: **"The owner explicitly authorized this
action; Codex is relaying that authorization."** Include the specific action and
the owner instruction or canonical decision reference. Use this wording only
when that authorization exists; a Codex review decision remains a Codex decision.

If Codex selects an authorized routine answer in the native terminal, record it
as an owner-authorized answer entered by Codex. Do not describe it as the owner
personally typing or clicking. A relayed owner instruction or AskUserQuestion
answer is not a native permission override. If the harness requires direct owner
approval, retain that gate; report the rejected action and reason and direct the
owner to the supported permission interface. Do not change settings, fabricate
human input, or retry through a different executor to clear a rejection.

## Preferred coordinator, recovery and closeout — 2026-10-10 (G052)

The handoff skill's helper now records one owner-selected, verified live Claude
session as this repository's preferred coordinator. Commands and refusal codes are
in the [skill](../../.agents/skills/restruct-sdlc/claude-session-handoff/SKILL.md); this section
records the operating rules.

- **Pin.** `pin --session FULL_UUID --selection-ref management/v0/PROGRESS.md`
  stores private metadata under `~/.local/state/claude-session-handoff/`, outside
  Git, keyed by the repository's Git common directory so all worktrees share it.
  It holds no PID, socket, token or transcript data and grants no authority. Live
  identifiers and pin bytes stay out of tracked files; Results record a UUID prefix
  and receipt ID only.
- **Routing.** Without `--session`, `check`/`send` route only when `status` is
  `pinned-live-verified`; anything else blocks with no send or spawn. Explicit
  `--session` routing, the newer-peer guard, authenticated inbox checks, single
  send, receipt wait and redaction are unchanged. Claude sessions message each
  other with native `ListAgents`/`SendMessage`; the helper stays Codex → Claude.
  `send` holds a shared maintenance lock in the private
  `~/.local/state/claude-session-handoff/<repo-key>/locks/` directory, which must
  be writable: a sandboxed Codex shell must allow it, or `send` fails closed.
- **Discovery.** Registry plus `claude agents --json --all --cwd <main checkout>`.
  Duplicates and stale registrations are reported, never acted on. An incomplete
  view (listing failure, unresolvable live cwd, invisible CLI-listed PID, or an
  unreadable or unverifiable registration not known to be another repository's)
  is `discovery-unknown`, never "no coordinator", so nothing is created from it.
- **Recovery.** A stopped pinned background session restarts under the same UUID
  with `claude attach <short-id>` or `claude respawn <short-id>`. This supersedes
  the earlier advice in this runbook to resume with `--resume <uuid>` and `--bg`:
  never use `--resume … --bg` or a name-based resume, which may start a copy.
  Before dispatching a replacement, inspect surviving children, `git worktree list`
  and Result/PROGRESS records.
- **Owner-directed relocation, verified 2026-10-13 / CLI 2.1.296.** At a quiet
  maintenance boundary, stop the exact verified session and verify stopped
  ownership/no path holders before `git worktree move`. `claude respawn <id>`
  preserves UUID/name/saved launch options. An absent old worktree makes resume
  fall back to its launch checkout; instruct the same session to use native
  `EnterWorktree` with the renamed existing path. Verify process/registry cwd
  and native `worktree-state`, refresh pin provenance and restore the observer
  from the current transcript tail. No metadata fabrication, permission change,
  fork or move beneath a live holder. Native gates remain binding. See the
  [relocation gate](../../distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#naming-dispatch-and-relocation-gate).
- **Creation.** Documented procedure only; the helper never starts a session. The
  read-only `status --creation-check` gate must report no reusable live or stopped
  coordinator, no unknown or ambiguous candidate, CLI 2.1.296 and a Codex-verified
  owner creation authority bound by path (no symlink), SHA-256 and quote; every
  live candidate must be named by full UUID in that quote. Nothing overrides a
  refusal. A different CLI version is a bounded escalation to Codex.
- **Closeout compaction.** After Codex or the owner verifies a standing terminal
  disposition from committed integration records, `compact` (dry-run first, then
  `--apply` bound to the reviewed selection digest) archives that task's whole
  delivery groups into the private archive before pruning them from the live log.
  Run it only against the main checkout, in a quiet window with no other helper
  invocation, from the host shell. It refuses unless the invocation path is the
  reviewed integrated helper (`--writer-sha256`), no other process runs a helper
  (any `handoff.py`, or a Python process whose argument resolves into a skill
  `scripts/` directory) or holds the log open, no same-user process is
  uninspectable where a writer could hide, and every CLI-listed live PID is
  visible. Only the Python helper's `send` appends to the log, so a non-Python,
  non-helper process whose descriptors cannot be listed (the session's
  `(sd-pam)`) is recorded as evidence, not refused; such holders, renamed copies
  outside the skill directories and `runpy` launches are covered by the quiet
  window, not the scan. Its manifest records whether the prune happened.
  The first live apply is owner- or Codex-operated; implementation grants no
  pruning authority.

## Coordinator milestone observer — 2026-10-12 (G053 AC13)

Failure wording (`failed`, `failing`, `failure`) and explicit decision/approval-needed requests qualify for
notifications alongside existing milestones. Absent/expected/resolved issues are filtered, with separate active
issues retained. Detection is bounded text matching of coordinator reports, not semantic AI or direct CI polling.

The handoff skill's `scripts/watch.py` can notify the owner's existing Codex supervisor thread when the verified
Claude coordinator reaches a milestone (decision, review-ready, blocked or question, waiting, completed). The standalone
companion is opt-in (`bind`, then `start`); this repository's adoption default activates it after verified bindings.
Commands, sources and refusals are in the
[skill](../../.agents/skills/restruct-sdlc/claude-session-handoff/SKILL.md). This section records the local operating rules.

This watcher reports Claude session output only. It does not inspect GitLab MR,
pipeline or job status and cannot serve as the monitoring owner for a required
pipeline. Assign and verify that owner separately under the
[GitLab branch workflow](GITLAB-BRANCH-WORKFLOW.md#required-mr-pipeline-monitoring-owner).

- **Pair lifetime.** The observer serves the repository's coordinator–supervisor pair across Goals, reviews and
  closeouts. Task closeout, `compact` and worktree removal do not stop it or reset its deduplication. The operator
  stops it with `stop`. There is no cron, service or boot startup. It polls every 30 seconds by default; an explicit `--interval`
  overrides that. Historical owner direction required G053 integration before continuous activation; that gate
  is satisfied. Current lifecycle/maintenance state is recorded in PROGRESS and the handoff record.
- **Binding.** It binds one explicit supervisor UUID and the pinned coordinator (`pinned-live-verified`). A retired
  or unverifiable session, a changed pin, incompatible Codex capabilities or an uncertain queue outcome suspends
  delivery until an explicit verified `bind --replace`. It never retargets by name or recency.
- **Sources.** Only the coordinator's own assistant output counts. Codex repository edits, management records,
  incoming handoffs and their replay, queue receipts, echoed notices and observer records cause no notice.
- **Compatibility.** Codex CLI 0.160.0 and daemon 0.162.1/0.163.0 are diagnostic baselines. Before delivery, require readable CLI/daemon version metadata and the queue command’s `--thread`/`--message` options. Other releases warn; unsupported capabilities stop delivery. No same-major compatibility promise is inferred.
- **Privacy.** The supervisor UUID and binding stay in the private 0700/0600 state directory. Tracked records cite
  only the fingerprint printed by `status`.
- **Transcript-size advisory.** It is a metadata-only check at the default 100 MiB threshold (configurable at bind,
  or opted out). One advisory recommends assessing a safe handoff or archive; nothing is pruned or rebound
  automatically.
- **Verified start.** `start` re-verifies the pair and otherwise stays inactive and reported. A repeated `start`
  on a stopped observer defaults to the current transcript tail; use `start --resume` for deliberate backlog
  catch-up. Existing dedup/advisory/smoke history is retained, uncertain intents still block fresh activation,
  and an already-running observer's cursor is never reset. Resolved/negated blockers are excluded from blocked
  notices without hiding another active blocker on the same line. A repeated start for a running observer
  is idempotent. The reusable adoption default starts it after verified bindings; G053 integration is complete.
- **Delivery safety.** Each notice's intent is persisted before `codex queue`, so an interrupted delivery is
  recorded as uncertain and never re-sent. Rows over 1 MiB are skipped with bounded progress and reported, never
  stalling later notices.
  A capability check that rejects before queueing leaves markers unconsumed. Explicit verified `bind --replace`
  permits recovery of known-unsent events, including older markers with no retained attempted/uncertain send for
  the key. Use `start --resume` for saved-cursor event recovery; fresh `start` skips backlog intentionally.
  Advisory/waiting conditions are rechecked, and smoke needs an explicit new command. Rebinding never queues;
  queued/uncertain outcomes remain protected against retry.
- **Authority.** Notices carry no approval or task authority, and do not satisfy any native permission gate.

## Explicit coordinator status — 2026-10-13

For ordinary replies to direct owner prompts, omit the `COORDINATOR_STATUS` footer unless the owner explicitly asks for supervisor relay. Exception: an owner decision that changes active work must immediately emit a `decision` footer, including when the same message answers the owner. Do not wait for another turn, worker handback or review-ready commit. Autonomous handoffs and required reviews, failure records and blockers still notify; existing legacy-prose/native-question detection is unchanged.

For autonomous supervisor handoffs, the coordinator ends with one plain `COORDINATOR_STATUS: CLASS | TASK-ID | SUMMARY`
line, following the [portable convention](../../distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#coordinator-final-status-line).
Owner-decision updates (`decision`), pending questions (`question`) and review handoffs (`review-ready`) take priority over
another task's running work. Name the actionable task in the final marker; keep routine task progress in the body.
The watcher prefers the valid final marker; `working` suppresses routine history/recovery notices, but an explicit
pending decision/review in that marker's summary still notifies. Classify the full summary within the existing
transcript-row limit, then truncate its delivered excerpt to 240 characters. A summary exceeding the recommended
240-character authoring limit must not invalidate the marker or silently discard its notification.
Historical body prose stays suppressed.
Native questions still notify. Absent/invalid markers retain legacy prose detection; source/echo,
deduplication, permission and approval boundaries are unchanged. Both supervisor backends use this classifier.

## Role effort defaults — 2026-10-12

The coordinator uses high for orchestration, status and handoffs, and xhigh when planning. Planning workers use
xhigh; implementation workers use high. Request worker effort explicitly through supported controls, verify the
actual or inherited value, and report unknown honestly. The owner may raise complex debugging or security work to
xhigh.


## Native isolated-worker names — owner direction, 2026-10-14

Unsupported task/role prefixes are treated as a Claude harness naming bug/
limitation pending resolution. Use the native isolated-worker tool and its
`agent-<id>` directory without another naming exception, approval or operator
prompt. Naming alone never delays an otherwise authorized dispatch. Record
actual task, role, agent/session, worktree path, branch and HEAD in the existing
record and verify intended ownership/isolation before writes.

This supersedes older task-specific naming-exception instructions, not native
permissions, locks, safe relocation, identity verification or Plan/input gates.
Never move a live checkout, pre-create an inaccessible path, create another
coordinator or install a workaround merely to force a preferred name. Preserve
existing active paths; prefer custom task/role names again when natively supported.

## Model-aware effort override — 2026-10-15

Default to Fable 5.1 high for planning and test corrections, including failure
diagnosis and corrective test/fix work; use xhigh selectively for especially
hard work. Default to Opus 5.5 high for implementation and routine test execution.
Default the Claude coordinator to Sonnet 5.5 high; the owner may choose Opus 5.5
high for coordinator work. If the owner selects Opus 5.5 for planning or test
corrections/diagnosis, use xhigh; routine implementation, test execution and
coordination stay high. This overrides generic role-effort defaults for those
models/phases. Honor owner choices, verify actual settings or report unknown,
and apply only at the next phase/safe boundary without restarting completed work,
interrupting a batch or resetting budgets. Prose does not configure the runtime.

## Compact repair context — 2026-10-15

Use the [repair checklist](../../distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#compact-repair-handoff)
for failed batches/corrections. Existing Goal/Plan/Result/review and artifact links
supply durable context; no separate packet, full-log payload or authority expansion.


## Fable weekly-limit fallback — owner direction, 2026-10-15

When Fable's weekly allowance is confirmed exhausted, automatically use Opus 5.5
**xhigh for planning and test-failure diagnosis/corrections** until Fable resets.
Opus implementation and ordinary test execution/reporting remain **high**.
Coordinator default remains Sonnet 5.5 high, with Opus 5.5 high available by
owner choice. This standing fallback needs no repeat owner approval; honor an explicit
owner override. Record the limit, fallback model/effort and actual setting (or
unknown) in the existing Result/dispatch record. A generic tool error is not
proof of a model limit.

Use supported controls at the next safe boundary, preserving the existing task,
artifacts, checkout, bindings, reviews and attempt count. Do not interrupt a
running validation batch, redo completed work, reset budgets or weaken gates.
If native switching is unavailable or refused, report the blocker once; do not
bypass permissions. Once Fable's allowance reset is verified, return to its normal
model/effort defaults at the next safe phase boundary, without restarting work or
repeatedly probing the allowance. This is model selection only, not scope,
execution-release or budget-extension authority.


For routine status/handoffs use [concise operational messages](../../distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#concise-operational-messages).
Report changes and actionable waits, reference the existing decision record and
continue authorized next actions; do not repeat settled context or approval prompts.

## Transport compatibility and stale sockets — 2026-10-15

Owner replaces exact-version transport allowlists with protocol/capability checks.
Claude still requires integer peerProtocol1, the same UUID/repository, live process
start, canonical Unix socket, private peer-key shape and matching PID/UID peer.
Each send rediscovers that UUID’s current registration. A pre-write ENOENT or
ECONNREFUSED allows one fresh lookup/reconnect; no other UUID, session creation,
endpoint guessing or cached direct supervisor socket is used. Once `sendall`
starts, failure is uncertain and must not trigger another send. Verify replay and
acknowledgment as before.

Codex delivery uses the installed CLI’s queue command, which resolves its daemon
endpoint, rather than a stored supervisor socket. Read-only help/version checks
verify the required command options/metadata; versions outside tested baselines
produce warnings in delivery records. They do not prove every future release’s
semantics. Missing capabilities, changed protocol, identity mismatch and uncertain
queue results still suspend. No automatic probe message or repeated retry is added.
Known-unsent capability blocks and historical version blocks can recover through
explicit verified rebind; attempted/uncertain sends retain dedup protection.
The separate coordinator-creation check remains version-restricted and authorized
independently. Restart a running observer at a safe point with `stop`, then
`start --resume` to load this code without resetting the pair/cursor/dedup.

## Coordinator retirement checkpoints — 2026-10-16

Follow [coordinator checkpoints and retirement](../../distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#coordinator-checkpoints-and-retirement):
150k active-context checkpoint, 200k compaction-or-handoff assessment, with safe
batch completion and compaction first. A new UUID requires verified single-owner
cutover, preserved pending work and explicit pin/observer rebinding. Existing
creation checks and native permissions are unchanged; no automatic replacement
script or live coordinator retirement is delivered by this policy amendment.


Codex supervisor
continuity follows the same durable-state/native-limit principles; replacing the
supervisor needs its own verified observer-thread rebind, not coordinator creation.

## Project rules and personal exceptions — 2026-10-18

Use [repository permission policy](CLAUDE-PERMISSIONS.md) for shared versus personal
rules and actual worker loading. A received handoff or settings hash verifies
neither loaded permissions nor execution authority. Report actual native holds
without settings self-modification, duplicate workers or speculative retries.

## Owner-readable status prose — 2026-10-18

Keep the useful detail in policy acknowledgments and multi-task reports. Write
short labeled bullets/fragments rather than repeated explanatory sentences;
[examples and boundaries](../../distributions/skills/restruct-sdlc/sdlc-management/references/coordination.md#concise-operational-messages).
No dashboard-only restriction, cryptic shorthand, authority omission or marker/parser change.


## Mandatory short coordinator reports — 2026-10-18

Routine owner-facing updates MUST be at most 80 words and three short bullets
(or one short paragraph), plus a status marker only when required for a supervisor handoff. Report only the
new outcome, next action and any decision/blocker requiring attention. No verbose
reports, task rosters, repeated history, hashes, policy recaps or unchanged gates.
Routine acknowledgments, dispatch receipts and unchanged waiting states get no
separate report. Put evidence, inventories and full handbacks in the existing
Result/review and link it; do not paste them into chat. Detailed chat is allowed
only when the owner explicitly asks for detail. A failure/decision must remain
clear within the limit; do not hide it in a link. Preserve truthful uncertainty,
source/final-ACCEPT distinctions and the exact status-marker syntax. This hard
default supersedes earlier guidance to preserve detailed status prose or avoid
fixed-length reports; it changes no technical, native-permission or review gate.

## Post-merge local qualification handoff — 2026-10-19 UTC

After an authorized merge, the existing Claude coordinator must automatically launch, own and track affected local staging batches/suites on the exact merged SHA under the [current policy](../../distributions/skills/restruct-sdlc/sdlc-management/references/validation-integration.md#post-merge-qualification-runs-locally). Name executable selectors and coverage rationale in the handoff. Full local staging-tip runs are owner-requested or supervisor-recommended, with the reason and frozen observed tip recorded; RC MRs retain full hosted qualification. A Codex merge handoff names that SHA; delivery/receipt is not proof that the run started or passed. Verify actual queued/running/terminal state and retain command, exit, logs and cleanup in the Result. Shared-resource contention queues the run visibly without an extra owner prompt, a new persistent coordinator or a hosted-full substitute. Successor implementation may overlap after merge; predecessor final ACCEPT still gates its integration.


## No automatic post-merge cloud checks — owner clarification

Automatic staging/main/tag push pipelines are disabled, including lightweight checks. After staging merge, the existing coordinator owns affected local qualification on the exact merged commit; feature MR hosted fast is the pre-merge gate, and RC promotion MRs carry required complete hosted release qualification. Keep explicit diagnostic/provider runs and owner holds separate; a skipped/canceled old-config pipeline is not passing qualification.


Owner context amendment,2026-10-20:150k checkpoint/200k safe handoff or compaction assessment at existing boundaries. Claude/Codex self-compaction is authorized through supported native controls at safe points; preserve active children/run and report any actual operator-only interface once. Batch handoffs and inspect only changed state; no unchanged polling/duplicate receipt recaps or full-history transfer. Native gates and required failure/review reporting remain.
