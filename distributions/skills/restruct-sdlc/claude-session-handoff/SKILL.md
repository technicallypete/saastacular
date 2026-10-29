---
name: claude-session-handoff
description: Discover and check existing local Claude Code sessions, send authorized Codex-to-Claude handoffs, and verify acknowledgment. Use for requests such as check Claude, send corrections to Claude, or inspect a running Claude session.
---

# Claude session handoff

Use shell tools to communicate with an existing local Claude Code session. Codex collaboration tools address Codex agents, not Claude Code sessions.

## Select the procedure

Read the active repository's `AGENTS.md` and `docs/agents/CLAUDE-SESSION-HANDOFF.md` when present. That repository runbook is authoritative for its local interface and workflow. Otherwise read [the bundled reference](references/local-session-handoff.md), adapted from the originating repository's runbook verified with Claude Code 2.1.290/2.1.291 and 2.1.296 on Linux/WSL. Its internal socket interface is protocol-checked; release versions are diagnostic baselines, not transport allowlists. Resolve repository paths against the target checkout, not this skill's installation directory.

## Check progress

Use [scripts/handoff.py](scripts/handoff.py) for routine checks and authorized
messages. It uses the standard library and the runbook's Linux protocol, currently
baselined with Claude Code 2.1.290/2.1.291/2.1.296. Protocol 1, current registration,
peer identity and key shape remain required; other release versions produce
`compatibility_warnings`, not a transport block. An incompatible protocol still stops delivery. Run scripts without reading their whole
source into conversation context. Resolve the script path from this skill folder.

```sh
python3 .agents/skills/restruct-sdlc/claude-session-handoff/scripts/handoff.py --repo "$PWD" status
python3 .agents/skills/restruct-sdlc/claude-session-handoff/scripts/handoff.py --repo "$PWD" check --match G006
python3 .agents/skills/restruct-sdlc/claude-session-handoff/scripts/handoff.py --repo "$PWD" check --session FULL_UUID --match G006
```

`status` reports the preferred-coordinator pin and duplicate-safe discovery (see
below). `discover` lists registrations, which may be stale, plus the same combined
discovery; `check` verifies process start and checkout. Use `--child CHILD_ID` for a known child transcript and `--limit 1`
for a short update. Output contains bounded assistant text, not raw tool results;
it is reported progress, not independently validated test evidence. Redaction is
best effort: never ask the helper to print credential files or authentication logs.

Discover sessions with `claude agents --json --all --cwd "$PWD"` from the target repository. Match the full session UUID and checkout; names may be stale. A resumed/forked
conversation can have a new UUID while its predecessor process remains alive.
Check recent transcript/task activity and `newer_live_sessions_check_target`;
never treat a live remembered PID as proof it is still the owner's active task.
When newer peers exist, `send` blocks unless `--confirm-target` is supplied after
verifying the selected task owner. Never automatically route to the newest peer. Confirm the current registration and process identity before using a remembered PID/socket. Do not select the first session by default.

Check the listing's `kind`: a stopped background entry and a live interactive
entry may share one UUID. `claude attach` and `claude logs` target background
sessions; they do not attach to the live interactive instance. For an interactive
session, use its verified inbox or ask the owner to deliver the prompt in its
existing terminal. Do not resume another copy to work around this distinction.
An unavailable host PID namespace can prevent helper verification; report that
access limitation instead of concluding the session stopped.

Use `claude logs <background-id>` or the exact session's JSONL transcript. Read a bounded tail of assistant text, timestamps and tool descriptions, keeping credentials and raw authentication/tool payloads out of output. Distinguish reported progress from independently inspected validation evidence. A busy session or expired monitor does not prove success or failure. Report the current stage, known blockers and remaining work; avoid continuous polling unless requested.

## Waiting-state prompt check

`waiting` with no unanswered transcript question does not prove there is no native
prompt. Before reporting no open questions, inspect the verified background
session's existing terminal with `claude attach <id>`; interactive sessions need
their existing terminal. If inspection is unavailable, report prompt state as
unverified. Socket messages may queue behind the dialog. Answer only under
existing owner authority using the native procedure below; verify the recorded
response, then detach without stopping the session. No automatic approval or
permission/settings change is implied.

## Preferred coordinator and routing

One existing Claude session may be pinned as the repository's preferred
coordinator. The pin is private machine-local metadata at
`~/.local/state/claude-session-handoff/<sha256(git common dir)>/pin.json` (0700
directories, 0600 file, no symlinks; shared by every worktree of the repository).
It stores the full session UUID, the checkout, name and CLI version at pin time,
the owner-selection record `{path, sha256}`, the time and any replaced UUID. It
holds no PID, socket, token or transcript data, never enters Git and grants no
authority.

```sh
H=".agents/skills/restruct-sdlc/claude-session-handoff/scripts/handoff.py"
python3 "$H" --repo "$PWD" pin --session FULL_UUID --selection-ref management/v0/PROGRESS.md
python3 "$H" --repo "$PWD" status
python3 "$H" --repo "$PWD" unpin --session FULL_UUID
```

- `pin` runs the same live verification as `send` (one live registration for the
  UUID in this repository, matching process start and integer protocol 1). Versions are diagnostics.
  Newer live peers require `--confirm-target` after verifying them; replacing a
  different pin requires `--replace`. It never sends or spawns.
- `status` is one of `unpinned`, `pin-unsafe`, `pinned-live-verified`,
  `pinned-not-live`, `pinned-ambiguous`, `pinned-unverified` or
  `discovery-unknown`, with live candidates and any change of name or checkout
  since the pin (provenance only, not blocking).
- Without `--session`, `check` and `send` route only when the status is
  `pinned-live-verified`; every other status blocks with no send or spawn.
  `--session FULL_UUID` routes exactly as before and leaves the pin unchanged,
  so multi-session routing remains available. Names are never targets.
- `send --task GNNN` labels each delivery record; records also carry `route`
  (`preferred` or `explicit`) and any peers confirmed with `--confirm-target`.
  `receipt --handoff-id ID` takes the session from that handoff's log record;
  pass `--session` when the record is absent or ambiguous. `queued_observed`
  reports an enqueue row without a replay (for example a message held or queued
  behind an open dialog); it never confirms delivery.

Discovery combines the registry with `claude agents --json --all --cwd <main
checkout>`. Duplicate names, duplicate registrations, stopped/done rows and stale
registration files are reported, never acted on or deleted. A failed listing, an
unresolvable live cwd, a CLI-listed live PID the helper cannot see, or an
unreadable or unverifiable registration (bad PID, permission-denied process or
`/proc` entry) that is not known to belong to another repository makes discovery
`unknown`, never "no coordinator". A live session whose cwd is an existing,
listable ordinary directory that Git reports, exactly, as "not a git repository"
is outside this repository, as is one in another repository; any other Git
outcome for a live cwd (missing or inaccessible path, dubious ownership, invalid
Git context, other errors) stays unresolved and keeps discovery `unknown`. A pinned UUID that returns under a
new PID revalidates automatically; a changed UUID (copy or fork) shows as
`pinned-not-live` plus the new candidate and needs an explicit re-pin.

**Claude ↔ Claude:** Claude sessions use native `ListAgents`/`SendMessage` (or
`/list-agents`) and short identifiers. This helper remains the Codex → Claude route.

## Compatibility and stale connections

The helper refreshes the same UUID's live registration/socket before connecting.
Only a pre-write missing/refused connection may trigger one rediscovery and retry;
recheck repository, process start, protocol, key and authenticated peer. Never
choose a different UUID or create a session. Once `sendall` starts, its failure is
uncertain: no automatic retry, even if the receipt has not appeared. Socket writes
and queue success still require the existing receipt/acknowledgment checks.

Capability checks establish the interface in use, not guaranteed compatibility
with every future release. Changed protocol/schema/options or uncertain delivery
stop the transport. The separate read-only **session-creation** gate retains its
verified-version restriction; this change grants no creation/permission authority.
Legacy suspended version blocks need one explicit verified rebind; retained
attempted/uncertain sends are never released by that recovery.

## Recovery and delegation

Reserve `0-agent-<id>` for the one active persistent coordinator; task workers prefer
their task prefixes when natively supported. Preserve the existing agent ID when available. Naming is
organizational, not identity/authority: verify discovery and the preferred pin.
Existing coordinator relocation waits for a supported safe point; do not move a
directory beneath a live session or restart/create duplicates just for naming.

Before dispatching any replacement, the coordinator checks surviving children
(`check --child`, `claude agents --json`), `git worktree list` and the Result and
PROGRESS records. One owner per active task; prefer task/role-prefixed worktrees
when natively supported. Unsupported naming is a Claude harness bug/limitation,
never a dispatch blocker: use its default isolated `agent-<id>` worktree without
a task-specific exception or naming approval prompt. Before task writes, verify
and record actual task/role/agent/path/branch/HEAD ownership and isolation. Native
permissions and locks remain binding; preserve active paths and never move a live
worker or duplicate sessions for naming. Handoffs carry Goal/Spec/Plan references,
phase, effort and file ownership. Pin, check and send dispatch nothing.

Preferred Claude names are `<task-id>-agent-<agent-id>-<role>` (`plan`, `check`,
`impl`); with native default names, record the role in the existing dispatch record.
Codex workers needing isolation use `.codex/worktrees/`; supervisors normally
stay in the main integration checkout. This does not rename existing live paths.

If the pinned session is stopped, the owner may restart it under the same UUID
with `claude attach <short-id>` or `claude respawn <short-id>`. Never use
`--resume … --bg` (it may start a copy) or a name-based resume. Fallbacks remain:
explicit `--session`, native Claude messaging, agent-view peek reply or
`claude attach <short-id>`, and the owner's terminal for interactive sessions.

## Creating a coordinator (documented procedure; the helper never spawns)

There is no `create` command. A new coordinator is created only when complete
discovery shows no reusable coordinator and no unknown or ambiguous candidate, and
Codex has verified a recorded owner creation authority. Run the read-only gate
immediately before creating:

```sh
python3 "$H" --repo "$PWD" status --creation-check \
  --coordinator-uuid FULL_UUID \
  --authority-ref management/v0/results/RECORD.md --authority-sha256 HEX \
  --authority-text "exact quoted owner decision" \
  --creation-name NAME --creation-prompt-file /path/outside-checkout/prompt.txt
```

Pass every coordinator UUID found in management records with
`--coordinator-uuid`; the pinned UUID and its replaced UUID are added. The gate
reports every failing predicate: `discovery-unknown`,
`reusable-live-coordinator`, `reusable-stopped-coordinator` (stopped/done row or
stored transcript), `ambiguous-candidates` (a live session not named by full UUID
in the quoted decision itself, or a live row with an empty or invalid session ID),
`coordinator-candidate-conflict` (a coordinator UUID named in the quote),
`missing-authority` (absent, outside `management/`, reached through a symlink,
untracked, modified, hash mismatch or missing quote), `unverified-cli-version`
(only 2.1.296; a change is a bounded Codex escalation, never widened
automatically) and `creation-command-arguments-invalid`. The helper checks only
those bindings; whether the record is a real creation decision, dated after the
discovery and naming every live candidate, is Codex's recorded judgement.

On any refusal, stop: reuse, restart, select or escalate. Only on a clean gate
does the operator run the printed command once from the main checkout
(`claude --bg --name NAME --effort xhigh PROMPT`, with no permission, model or
settings flags), re-run `status` and require exactly one new live session. The
2.1.296 help documents no `--` option terminator, so none is added; instead the
gate refuses a prompt that starts with `-` or is a single token, either of which
could be parsed as an option or a subcommand.
Anything else stops and is reported, with no killing. Then verify, `pin` it
explicitly and confirm a receipt.

## Task-closeout compaction (archive before prune)

`compact` moves a verified-terminal task's delivery records into a private archive.
It is dry-run by default and writes nothing; apply needs the dry-run's selection
digest. Invoke it, like `send`, against the **main checkout** only; linked
worktrees are refused.

1. **Verify closure first (Codex or owner).** Identify the standing terminal record
   on the integration branch: for `completed`, the task's final
   `results/<TASK>-implementation-codex-review*.md` ACCEPT after the qualifying
   staging disposition; for `canceled`/`superseded`, the PROGRESS entry or review
   recording it (and the successor). Confirm no later record reopens the task and
   quote the exact sentence. A task label, handoff ID, receipt or idle PID is not
   closure. The closure reference must be the real file: a path through a
   symlink is refused.
2. **Dry-run** and review the selection, retained reasons and prerequisite evidence:

   ```sh
   python3 "$H" --repo MAIN_CHECKOUT compact --task GNNN --disposition completed \
     --closure-ref management/v0/results/GNNN-implementation-codex-review.md \
     --closure-sha256 HEX --closure-text "quoted ACCEPT sentence" --verified-by codex \
     --writer-sha256 REVIEWED_INTEGRATED_HELPER_SHA256 [--handoff-id UUID ...]
   ```

3. **Apply** with `--apply --selection-sha256 DIGEST` only during a quiet
   maintenance window: start no other helper invocation and run from the host
   shell (a separate PID namespace makes visibility unproven and refuses). Invoke
   `python3 …/handoff.py` directly; a wrapper process whose own arguments name
   `handoff.py` (for example `timeout`) counts as another invocation and refuses.

Groups move whole, by handoff ID. Only groups labelled with the task, or listed
with `--handoff-id` and unlabelled, from one session, are archived; failed,
unconfirmed and incomplete deliveries keep their bytes (incomplete ones are
flagged). Other tasks, conflicting labels, multi-session groups, unattributed
records and malformed or truncated lines stay live, even when listed.

Apply holds the stable maintenance lock exclusively (`send` holds it shared for its
whole invocation) and refuses unless: the running helper, the main checkout's
committed and working `handoff.py` and `--writer-sha256` are one hash; no other
same-user process runs a helper (any `handoff.py` argument, or a Python process
whose argument resolves, relative to its cwd and through symlinks, into a skill
`scripts/` directory of any worktree or the user copy) or holds the live log
open (a running process is never trusted from its current source path, since it
may have loaded older code); no same-user process (by `status` real Uid) is
uninspectable where a writer could hide (unreadable or malformed `status`,
unreadable or empty `cmdline` except zombies, a Python process with relative
arguments and an unreadable cwd, or an unlistable descriptor directory for a
Python or helper-named process); and every live PID in `claude agents --json` is
visible, with at least one. Only the Python helper's `send` appends to the log
(Claude sessions use native messaging), so a non-Python process that names no
helper but whose descriptors cannot be listed, such as the session's
`(sd-pam)`, is reported as `uninspectable_nonhelper` (PID and comm only) rather
than refused. Renamed copies outside the skill directories, `runpy` launches and
other non-helper descriptor holders are left to the quiet-window procedure, not
the scan. It writes `<state>/archive/<UTC>-<task>.jsonl` plus a
manifest (closure binding, counts, IDs, incomplete flags, writer hash,
prerequisite summary, hashes), verifies the archive, rescans, and only then
atomically replaces the live log. The manifest then records the outcome:
`pruned: true` with `prune_status: pruned`, or `pruned: false` with the refusal
reason; `pending` means the run stopped between the archive and the replace. Any
failure before the replace leaves the live file unchanged; a rerun is safe and
repeated closeout archives nothing. A failure after the replace (for example
`pruned-directory-fsync-failed`) is reported as having pruned. Record the archive basename and
manifest SHA-256 in the closeout summary. There is no scheduler, no compaction on
send and no deletion of archives, transcripts or approvals.

## Answer a pending routine question

When the owner authorizes answering a routine question in an existing background
session, use `claude attach FULL_OR_SHORT_ID` in an interactive PTY after verifying
the exact live identity. Inspect the displayed question/options, choose only the
answer supported by owner instructions, and verify the resulting AskUserQuestion
tool-result in that session's transcript. Do not infer an answer from socket
receipt: peer messages can remain queued while the dialog is open. On verified
Claude 2.1.296, Ctrl+Z detaches the view while the original session keeps running;
do not resume another copy or stop the worker to answer it.

This route is for authorized ordinary choices, not blanket approval of permissions,
architecture/scope changes or unrelated external actions. Never change settings or
fabricate transcript replies. If the displayed question differs or requires an
unresolved owner decision, report it rather than selecting blindly. Capture bounded
clean output; avoid replaying terminal redraws into conversation context.

## Send a handoff

For repair/diagnosis handoffs, give the failed check, exact revision/relevant diff,
focused sanitized evidence, tried fixes, current hypothesis and next smallest
test, with the current attempt/scope and owner. Link existing records/artifacts
rather than sending full logs or creating a new packet. See the
[compact repair checklist](../sdlc-management/references/coordination.md#compact-repair-handoff)
when that companion is installed; this summary also supports standalone use.

For bounded specification/epic mandates or stop changes, cite the actual recorded owner instruction, exact task
membership/endpoint and permitted phases. Stops override existing Specs/Plans and automatic lookahead; verify
acknowledgment and safe pauses of affected work. No notice/receipt creates mandate or resume authority. See
[mandates and stops](../sdlc-management/references/roles-lifecycle.md#bounded-specification-authority-and-owner-stop-boundaries).

An adopting repository's automatic planning rule at initial implementation-MR creation can supply standing
authority for a bounded
planning dispatch (draft implementation MRs included; no CI/Duo/merge wait). Verify the real MR/candidate,
ready Goal/Spec and task owner, label pending dependencies provisional and relay the
remaining execution gates. Neither a merge notice nor this helper independently authorizes planning or execution;
see [lifecycle policy](../sdlc-management/references/roles-lifecycle.md#automatic-planning-after-integration).

For automatic execution dispatch, cite reviewed predecessor integration, pending staging qualification and the
supervisor's recorded reconciliation/release with closed start gates. Isolated implementation may overlap
predecessor CI; final predecessor ACCEPT still gates successor integration/final acceptance. Preserve explicit
holds and affected security/architecture/contract stops. Route staging repairs to the predecessor implementation
owner, with its branch/writer and ledger; successor implementation proceeds independently. Verify acknowledgment/dispatch; a merge, CI or observer notice alone grants no execution authority. See
[release policy](../sdlc-management/references/roles-lifecycle.md#automatic-reassessment-and-execution-release).

For in-scope mechanical repairs under standing execution authority, use working
status and the existing Result; do not request another permission/replan handoff.
Escalate actual scope/security/contract or unapproved production schema changes,
ambiguity and budget/native stops; transport never supplies that authority.

For an already-authorized instruction, write its concise text to a local file and
reference the actual repository root's management documents:

```sh
python3 .agents/skills/restruct-sdlc/claude-session-handoff/scripts/handoff.py --repo MAIN_CHECKOUT send \
  --message-file /tmp/authorized-handoff.txt --task GNNN \
  --document management/v0/PROGRESS.md \
  --document management/v0/results/GOAL-execution-binding-codex-confirmation.md
python3 .agents/skills/restruct-sdlc/claude-session-handoff/scripts/handoff.py --repo MAIN_CHECKOUT receipt \
  --handoff-id UUID_RETURNED_BY_SEND
```

Omitting `--session` uses the verified live pin; add `--session FULL_UUID` to
target another session explicitly. Use the main checkout as `--repo` for `send`
so delivery records land in its log, the one `compact` maintains.

Use the actual document name, not the example placeholder. `--document` can repeat;
it attaches paths and byte hashes, not document bodies. Direct the recipient to read
the referenced records for decisions, bindings, scope and remaining gates. Root
management references never authorize copying newer inputs into a bound checkout.

When a handoff needs context not already recorded, save a concise management
handoff document under the repository's existing conventions before sending it.
Include current phase, decisions, unresolved findings, next action and authoritative
review/binding links; pass that file as `--message-file` or `--document`. Do not
store the entire chat or secrets. This gives the next session durable context;
the delivery log alone cannot reconstruct a message from its hash.

The helper resolves `--repo` to its Git root, including when launched in a service
directory. By default, it appends minimal prepared/sent metadata to
`management/SESSION-HANDOFFS.jsonl`; `--record` selects another root-relative
management `.jsonl` file. The log stores session/checkout, unique handoff ID,
message hash and document references, never message bodies or peer credentials.
It records delivery attempts, not decisions or task completion. Canonical reviews,
Results and PROGRESS retain those responsibilities. The default delivery log is
machine-local, gitignored and must remain untracked; keep its local history,
not a committed copy. Alternate delivery logs must also stay local/untracked.
The helper never commits, pushes or changes approval records. No personal-
directory or execution-input copy is automatic.

`send` sends once and waits up to 20 seconds for its unique transcript replay
marker (`--wait-seconds 0..60`). It returns `receipt-confirmed`, or
`delivery-unconfirmed` with exit code 2. It holds the maintenance lock shared
from before its first record to after its final one; if the lock is not free
within `--lock-wait-seconds` (default 10, while a closeout applies) it blocks
without sending. `send` therefore needs the private, writable
`~/.local/state/claude-session-handoff/<repo-key>/locks/` directory; a sandboxed
Codex shell must allow writes there, or `send` fails closed. An unsafe state
directory (symlink, foreign owner, group/other access) blocks `send` as
`state-unsafe`; repair the modes rather than bypass it, and use the fallback
routes meanwhile. Each record starts on a new line. Socket-write success alone is not
delivery. Unconfirmed can mean a held/refused/queued message, a stale target or
replay outside the bounded tail; do not assert which without evidence. Inspect
the intended session/approval state rather than resending automatically.
The helper cannot approve a held message or restart/change permissions.
Use normal `claude attach` delivery only under existing authorization; `receipt`
verifies replay independently. Receipt is not model acknowledgment or dispatch:
check the recipient's subsequent assistant reply and actual Agent-tool dispatch.
For replies, inspect that exact session's bounded transcript with `check`; a
Codex session need not have a Claude peer socket to receive its assistant reply.

Use the repository runbook's authenticated local inbox procedure, or the bundled reference when no local runbook exists. Verify UUID, checkout, live process start, protocol and canonical socket path; read the peer token into memory without printing it. Send routine instructions with `priority: next` and a unique message UUID. Never fabricate transcript entries or launch a second copy of the running conversation to deliver a message.

State the authorized task, exact inputs/revisions and checkout, permitted outputs, scope/budget, required validation and stop conditions. Preserve active input bindings. Do not copy newer planning documents into an execution checkout without its approved reconciliation.

Socket delivery is not acknowledgment. Verify the intended transcript's user replay or Claude acknowledgment before claiming receipt. Require explicit evidence before claiming subagent dispatch or completion. If delivery remains pending during a tool run, report that state rather than repeatedly resending.

If the inbox is incompatible or unavailable, use the supported attach/input interface or report the blocker. Do not restart a live session, bypass a held permission prompt or change permission modes merely to make delivery work. Follow existing user authorization for any such action; this skill grants none.

## Coordinator milestone observer (opt-in here; an adopting repository may default-enable it after verified bindings)

For ordinary replies to direct owner prompts, omit the `COORDINATOR_STATUS` footer unless the owner explicitly asks for supervisor relay. Exception: an owner decision that changes active work must immediately emit a `decision` footer, including when the same message answers the owner. Do not wait for another turn, worker handback or review-ready commit. Autonomous handoffs and required reviews, failure records and blockers still notify; existing legacy-prose/native-question detection is unchanged.

**Coordinator response contract:** finish autonomous supervisor handoffs with one plain final line:
`COORDINATOR_STATUS: review-ready | G008 | Plan ready for supervisor review.`
Classes are `working`, `decision`, `review-ready`, `blocked`, `question`, `failed`, `completed`.
Use `decision` immediately for direct owner changes to active work, with a source
reference and the supervisor reconciliation needed; keep ongoing work in the body.
The receiver verifies the source; this report adds no acknowledgment gate or authority.
Use the actual nonsecret task ID and a nonempty summary of at most 240 characters
without `|`. Routine acknowledgments/progress and recovery use `working`, which
queues no milestone. Plans/Results/bindings awaiting review use `review-ready`;
`completed` means verified terminal closeout, never self-acceptance. Pending owner/
supervisor decisions (`question`) or submitted reviews (`review-ready`) take priority
over another task's routine work: name the task needing attention in the final
marker and put running tasks in the body. Never label such a handoff `working`.
As a safety fallback, an explicit pending decision/review in a `working` marker's
bounded summary still notifies; routine history elsewhere in the body stays silent.
A valid final marker otherwise overrides prose classification; native questions still notify. Markers
inside fences/quotes, nonterminal or malformed markers are not milestones;
other legacy prose remains detectable when no valid final marker exists.
Existing source/echo exclusions, deduplication and authority boundaries remain.

[scripts/watch.py](scripts/watch.py) notifies one existing Codex supervisor thread of the verified Claude
coordinator's milestones: decision, failure, review-ready, blocked or question, waiting, completed. Explicit decision/approval
requests qualify as questions; absent, expected or explicitly resolved failures do not qualify as failure alerts.
Detection uses bounded text patterns, not AI inference or direct CI polling. It is a stdlib polling loop (default
30 seconds) with no model inference. It belongs to the repository's coordinator–supervisor pair, not to one task:
once enabled it keeps running across Goals, reviews and closeouts until the operator stops it.

```sh
W=".agents/skills/restruct-sdlc/claude-session-handoff/scripts/watch.py"
python3 "$W" --repo "$PWD" bind --supervisor-file /private/path/supervisor.json   # or --supervisor FULL_UUID
python3 "$W" --repo "$PWD" start [--interval 30]
python3 "$W" --repo "$PWD" start --resume   # optionally catch up from the saved cursor after stopping
python3 "$W" --repo "$PWD" status
python3 "$W" --repo "$PWD" stop
```

- **Binding.** `bind` records the existing supervisor UUID (never a name) and the current pinned coordinator, which
  must be `pinned-live-verified`. `status` shows only a supervisor fingerprint. A different supervisor or
  coordinator needs an explicit `bind --replace` after verification; the observer never retargets.
- **Sources.** The only source is the pinned coordinator's own assistant output. Incoming handoffs and their
  replay, queue receipts, notices echoed back (`[observer-notice: …]`), subagent sidechains, repository files
  (including management records), Codex transcripts and the observer's own records are never sources, so notices
  cannot loop. "Waiting" is reported as *native prompt not inspected*; inspect it before answering.
- **Delivery.** Notices go through `codex queue --thread UUID`. Before sending, read-only checks require the
  queue command’s `--thread`/`--message` options and readable CLI/daemon version metadata. Codex CLI 0.160.0 and
  daemon 0.162.1/0.163.0 are baselines; other releases warn, without an exact- or major-version gate. An idle poll makes no queue or CLI call. The outcome is recorded as `queued-not-acknowledged`, never as
  acknowledgment. Before each `codex queue` call, the notice's intent and dedup key are written to private state;
  if that write fails, nothing is queued. An intent left unfinished by a crash or a failed final write is recorded
  as `uncertain-interrupted` and never retried, even after rebinding. Queue failure or uncertainty, incompatible
  capabilities, identity loss, a retired session or a changed pin suspends delivery, with no retry, until an explicit
  `bind --replace`.
  A capability rejection before any queue call records a known-unsent outcome without consuming event/advisory/
  waiting/smoke markers. After explicit verified rebinding, these may be delivered; use `start --resume` when
  recovering transcript events from the saved cursor (fresh `start` intentionally skips old backlog). Rebinding
  also releases older known-unsent markers only when retained history contains no attempted/uncertain send for
  that key. Queued or uncertain deliveries remain deduplicated. Smoke recovery requires another explicit `smoke`
  command; rebinding itself never queues anything.
- **Parsing limit.** A JSONL row longer than 1 MiB is not parsed. The observer scans past it in bounded chunks,
  resuming across polls while it grows, and records it under `status` `parse_limit`. If the row was coordinator
  output, one notice says it could not be assessed for milestones. Later rows are still delivered.
- **Lifecycle.** `start` first re-verifies the bound pair; if it cannot, it stays inactive and reports why. It then
  starts at the current transcript tail, including after an earlier activation, so old backlog is not replayed.
  Use `start --resume` after stopping when deliberately catching up from the saved cursor. Both modes preserve
  deduplication and advisory/smoke history; fresh start refuses unresolved delivery uncertainty. An already-running
  observer's cursor is never reset. Explicitly resolved or negated blockers do not produce blocked notices;
  another active blocker on the same line still qualifies. It
  launches one explicitly managed background process per repository (a lock plus a verified PID, start time and
  command line). A second `start` for a running observer is idempotent (`already-running`), and replacing the pair
  needs an explicit verified `bind --replace`. `stop` never signals a reused PID. There is no cron, service or boot
  startup. Task closeout, delivery-log compaction and worktree cleanup leave its state alone.
- **State.** Deduplication keys, transcript offsets, delivery records and the log live in 0700/0600 files under
  `~/.local/state/claude-session-handoff/<repo-key>/observer/`, outside Git. Unsafe modes or symlinks block.
- **Transcript-size advisory.** Each poll reads the coordinator transcript's size from file metadata only. At or
  above the threshold (default 100 MiB, `--size-threshold BYTES` at bind, or `--no-size-advisory` to opt out), one
  advisory with size, threshold and a safe handoff/archive recommendation goes to the supervisor. It is sent once
  per coordinator UUID and threshold, including on a first poll that is already large, and is kept across restarts,
  rotation, size decreases and closeout. `status` shows size, threshold and advisory state. It never prunes
  transcripts or retires, creates or rebinds sessions.
- **Adoption default.** When an adopting repository's activation option is enabled (the default) and it has
  explicitly established and verified the existing coordinator pin and supervisor UUID, run `bind` and then
  `start` (30-second default) with no further enablement question. A missing, ambiguous, unverified or retired
  binding leaves the observer inactive and reported. Never guess a UUID from a name, create a session or change
  permissions to satisfy it. The opt-out leaves it inactive.
- **Authority.** A notice carries no approval, task or permission authority. `smoke --label TEXT` sends exactly one
  labelled benign delivery to check the route.

## Effort for coordinators and workers

The persistent coordinator defaults to **high** for orchestration, status and handoffs, and uses **xhigh** when it
plans. Planning workers default to **xhigh** and implementation workers to **high**. Set worker effort explicitly
through supported dispatch or session controls where available; never assume inheritance, and never claim that
handoff prose sets it. Verify the actual or inherited value where metadata allows; otherwise report it as unknown.
The owner may raise affected work to xhigh for complex debugging or security decisions. Effort never changes
authority, budget or a running session.

## Owner authorization attribution

When relaying an actual owner decision, write: **"The owner explicitly authorized this action; Codex is relaying
that authorization."** Name the action and the owner instruction or decision record. A Codex decision remains a
Codex decision. An owner-authorized answer entered by Codex is recorded as such, never as the owner typing or
clicking. A relayed authorization does not override a native permission rejection or a gate that needs direct owner
action. Report the rejected action and reason, and point the owner to the supported permission interface. Never
change settings, fabricate input, or retry through another executor.

## Authority boundaries

Checking status does not authorize messaging or new implementation. A handoff conveys only work already authorized by the user and repository governance. It does not grant Plan approval, acceptance, scope expansion, commits/pushes, hosted actions, budget extensions or permission changes. Preserve the implementer's checkout and running resources while inspecting progress.

## Model-aware effort override — 2026-10-15

Default to Fable 5.1 high for planning and test corrections, including failure
diagnosis and corrective test/fix work; use xhigh selectively for especially
hard work. Default to Opus 5.5 high for implementation, routine test execution
and coordination. If the owner selects Opus 5.5 for planning or test
corrections/diagnosis, use xhigh; routine implementation, test execution and
coordination stay high. This overrides generic role-effort defaults for those
models/phases. Honor owner choices, verify actual settings or report unknown,
and apply only at the next phase/safe boundary without restarting completed work,
interrupting a batch or resetting budgets. Prose does not configure the runtime.


## Fable weekly-limit fallback — owner direction, 2026-10-15

When Fable's weekly allowance is confirmed exhausted, automatically use Opus 5.5
**xhigh for planning and test-failure diagnosis/corrections** until Fable resets.
Opus implementation, ordinary test execution/reporting and coordination remain
**high**. This standing fallback needs no repeat owner approval; honor an explicit
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


Use the adopting repository’s low-overhead delivery policy; when the companion is installed, see
[validation/failure handling](../sdlc-management/references/validation-integration.md#validation-economy-and-recurring-failures),
[routine transitions](../sdlc-management/references/roles-lifecycle.md#continue-routine-transitions-without-owner-nudges),
[concise messages](../sdlc-management/references/coordination.md#concise-operational-messages) and
[closeout receipts](../sdlc-management/references/closeout-risks.md#routine-closeout-receipt).
Preserve standalone operation and every existing authority/native-permission boundary.

Local validation handoffs specify focused executable selectors for changed
behavior from the first run, not umbrella suites by default. The coordinator
checks actual command scope before dispatch; broad exceptions need a technical
coverage reason and supervisor disposition, not convenience. Reuse approved
exceptions; preserve required hosted checks and existing authority/native gates.


## Mandatory task-acceptance cleanup — owner direction, 2026-10-17

At every final task ACCEPT, immediately remove all finished task-owned worktrees
and temporary resources under the [mandatory cleanup policy](../sdlc-management/references/closeout-risks.md#mandatory-cleanup-at-every-task-acceptance).
Unlock finished agent-tool worktrees; preserve unique commits and dirty/untracked
work before force removal when needed. Clean task-owned containers, unused images
(including attributable untagged IDs), volumes and disposable evidence without
another routine approval. Keep active coordinator/successor/preview resources and
specific recorded retention/native-permission holds. Verify a compact receipt;
no silent lock/dirty-work deferral, blanket host prune or extra test campaign.


## Private files for new tasks — 2026-10-17

Use the main integration checkout's ignored `.local/{evidence,secrets,tmp}/` for
new task-private files, shared across worker worktrees. Follow the [local storage
policy](../sdlc-management/references/adoption.md#repository-local-private-files): task-scoped paths, private permissions, Git/build/upload exclusions
and separate evidence/secret cleanup. Existing tasks keep current paths through
acceptance cleanup; do not migrate active bindings or native session state.

## Repository-scoped permission configuration

Keep shared project rules tracked and repository-specific personal exceptions
ignored; home settings are for intentional cross-project preferences. Verify the
actual loaded rules for the session/worker launch directory. Owner-authorized
configuration is distinct from task authority; native rejection gates remain.
Follow the adopting repository's permission placement/adoption procedure.

## Mandatory short owner reports

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

## Post-merge qualification handoff

When the adopting workflow uses local-after-merge qualification, an authorized merge handoff names the exact merged SHA, affected executable selectors and the local run owned automatically by the existing implementer coordinator. Full local tip runs need an owner request or supervisor recommendation, with a recorded reason and frozen source. Verify actual queued/running/terminal state and Result evidence; receipt alone is not execution or PASS. Do not require another owner launch prompt, create a second persistent coordinator or use hosted fast success as full qualification. Preserve resource locks, final acceptance holds and the adopting repository's current policy.
