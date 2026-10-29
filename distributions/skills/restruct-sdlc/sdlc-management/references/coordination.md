# Coordinator, workers and messages

## Local qualification after merge

Under the local-after-merge default, the existing implementer coordinator automatically launches and owns affected local qualification of the exact merged integration SHA. Full local runs on the integration tip are owner-requested or supervisor-recommended; record why and freeze that tip. A merge handoff identifies that SHA and required run; no extra owner prompt or persistent coordinator is needed. Track its actual state and evidence in the existing Result, with one resource-blocker report if queued. Preserve immutable inputs, shared locks and cleanup. Ready successor source work may overlap after merge; predecessor final ACCEPT still gates successor integration.

Transport is not here. Session discovery, routing, pinning, receipts and creation gates belong to the companion
[claude-session-handoff](../../claude-session-handoff/SKILL.md) skill; supervision belongs to
[repository-supervisor](../../repository-supervisor/SKILL.md). This reference states the management rules those
tools serve.

## One coordinator, one owner per task

- Exactly **one active, owner-selected persistent coordinator** per repository. It owns worker dispatch and
  integration of delegated work. The owner selects it; the helper's private pin records that selection.
- **One verified owner per task**, with non-overlapping file ownership. Multi-session routing stays available:
  an explicit `--session FULL_UUID` target always works and leaves the pin unchanged.
- **Recover before replacing.** Before any replacement, check surviving children, `claude agents --json`,
  `git worktree list` and the Result and PROGRESS records. A surviving worker is reused. Missing metadata or an
  unknown/ambiguous discovery result **never** authorizes creating a coordinator or a duplicate worker.
- Creating a coordinator follows the companion's read-only `status --creation-check` gate and a recorded owner
  creation authority; on any refusal: reuse, restart, select or escalate.
- Service work explicitly reads the service root's scoped instructions and skills. A shell `cd` does not refresh
  skill discovery.

## Default to safe parallel work lanes

When authoring a Goal/Spec, the supervisor records **high** or **low** priority
and the owner instruction or supervisor rationale that selected it. The owner
should identify low-priority work explicitly. If priority is unspecified, the
supervisor may determine it or ask the owner when the distinction materially
affects scheduling; normally default to **high** without a routine question or
specification hold. An explicit owner choice takes precedence. Attribute a
supervisor determination accurately and reflect later changes at a safe handoff.

The Plan carries that priority into the implementation worker allocation:

| Priority | Target parallel implementation workers, whenever feasible |
|---|---|
| High (usual default) | 3–5 |
| Low | 1–3 |

The counts are concurrent implementation workers during dependency-ready phases;
the planner, coordinator and reviewers do not count toward them. Use useful
independent work within approved scope, exclusive write ownership, native/resource
capacity and budget. Record concrete constraints when fewer workers are feasible;
the targets do not justify duplicate work, resource-limit changes or idle agents.
The planner itself may be one agent. Parallel validation alone does not replace
allocating implementation workers where independent source work exists.

Use the existing Plan's delegation and sequence sections to record worker counts
by phase, each worker's responsibility and writable paths, input SHAs, dependency
order, integration owner, evidence and shared-resource reservations. State a
concrete dependency, ownership or capacity reason when the allocation falls below
its priority target. Do not manufacture work or duplicate an artifact's
writer to fill slots. Keep one accountable task owner and one writer per artifact;
use approved delegation and separate owned checkouts for concurrent writes.

At implementation kickoff, check actual source ownership, dependency readiness,
native slots and resource reservations, then dispatch all ready workers specified
in the accepted Plan. Kickoff follows the accepted allocation and sequence;
it is not a fresh planning phase or a reason to reduce workers for convenience.
As dependencies clear, dispatch the next planned workers without another routine
owner prompt. Record an actual allocation change and its concrete reason in the
existing Result; material scope, design, budget or gate changes retain the normal
review/escalation procedure.

Low-priority work may use one implementation worker within its 1–3 target
and yields capacity to the current high-priority implementation. Normally
concentrate available workers within that task. Concurrent Goals retain existing
authorization, dependency and isolation gates; supervisor management updates may
proceed. Priority grants no new task, phase or resource authority.

Parallelize only when dependencies are closed, edits do not conflict, each lane has
an attributable baseline/result, and locks, resource limits, budgets and owner stops
permit it. Preserve task-specific sequencing and immutable validation sources;
reconcile any material change with the supervisor without rewriting approved Plan
history. Serialize shared-file edits, dependent migrations, mutable fixtures,
shared stacks/ports/caches and image builds where policy requires. Private DB and
browser runs share the configured test pool; image builds remain sequential. A
failure blocks its lane and dependent work, not unrelated safe lanes. The lane's
finish → diagnose/fix → rerun → advance gates still apply; parallel work never waives
coverage, review, integration, acceptance or a failed required gate.

For active tasks, apply the stronger allocation rule at the next safe handoff.
Preserve accepted Plan/history bytes, running workers and frozen evidence; record
any needed delegation reconciliation in the existing Plan/Result before affected
dispatch. Reclaim completed workers at handback so their capacity is available.
Use existing coordinator facilities and records rather than adding a scheduler.

## Receipt, acknowledgment and authority

- **Receipt is not acknowledgment, and acknowledgment is not execution.** A socket write proves bytes left; a
  transcript replay proves receipt; the recipient's reply proves acknowledgment; actual dispatch or artifacts prove
  execution. `queued_observed` and `delivery-unconfirmed` are never receipts.
- Peer messages **cannot** grant owner approval, change permissions, providers or models, or bypass another
  session's rejection. Existing owner authority persists within its recorded scope; routine actions need no
  repeated approval.
- A native permission block is surfaced with its action and reason and answered only through the harness's valid
  operator procedure; delegated operator input is labelled honestly. No impersonation, no permission-rule
  workaround, no duplicate executor.
- **Owner attribution.** Relay an actual owner decision as: "The owner explicitly authorized this action; the
  supervisor is relaying that authorization", with the specific action and the instruction or decision record. A
  supervisor decision stays the supervisor's. An owner-authorized answer entered by an agent is recorded as such,
  never as the owner typing or clicking. Such relays never override a native permission rejection or satisfy a gate
  that needs direct owner action: report the rejected action and reason, and point the owner to the supported
  permission interface.
- **Milestone notices.** The companion's optional observer (`watch.py`) notifies the supervisor of coordinator
  milestones. Notices carry no approval or task authority, and only the coordinator's own output is a source.
  Review-ready notices include execution bindings ready for mechanical verification or confirmation. Conditional
  stop rules such as “Any failure stops it” are not reports of an actual failure. The observer uses bounded text
  patterns rather than language inference; verify the linked records before acting. Keep detection code and tests
  in `claude-session-handoff`, rather than duplicating them in this skill.
- Transport support is Linux/WSL with the companion’s protocol/capability checks. Release versions warn rather
  than gate messaging; unsupported protocols/options, ambiguous identity and uncertain delivery still stop.
  Refresh the same session/socket; one pre-write missing/refused reconnect is allowed, never an uncertain resend.
  Session creation keeps its separate verified-version gate; there is no blanket future-compatibility promise.

## Supervisor status observation — owner clarification, 2026-10-20

This clarifies the existing communication workflow; it introduces no new transport,
review, notification or response-waiting stage.

The supervisor does not poll the coordinator or its sessions for progress. The
Claude session-handoff watcher already polls and delivers milestone notices; rely
on those notices unless the owner explicitly requests a status check. Do not repeat
session checks, ping the watcher or reread unchanged state. When a new notice
arrives, verify the linked repository records directly. Send needed coordinator decisions, corrections or requests immediately using
nonblocking delivery. Continue supervisor work without waiting for acknowledgment
or polling the coordinator transcript; the watcher delivers responses asynchronously.
Receipt is not acknowledgment or execution. This changes observation cadence only;
it does not suppress required reviews, failure reports, owner decisions or gates.

The Claude session-handoff watcher observes Claude session output and milestone
notices; it does not inspect GitLab MRs, pipelines or jobs. Do not treat a watcher
notice, general task ownership, an auto-merge setting or a review request as
proof that a GitLab pipeline is being monitored. Each required MR pipeline has
one explicitly named monitoring owner (normally the coordinator), recorded with
the MR/pipeline ID and immutable source SHA when the MR is opened or the pipeline
is created. That owner verifies the expected pipeline started, tracks it to a
terminal state, reports the outcome, and routes the first failure for diagnosis.
If ownership changes, record a direct handoff and acknowledgment so there is no
unowned interval. Monitoring is event-driven and concise: report meaningful state
changes, terminal outcome and actionable failures, not repetitive unchanged
polls. Auto-merge does not replace monitoring or change review and merge authority.

## Apply workflow updates to live repository guidance

When the supervisor updates Restruct SDLC workflow rules, it also updates the
corresponding live governance files in the adopting repository in the same change
batch, then commits and pushes to that repository's configured integration branch
under its publication policy. Updating reusable references or adoption templates
alone is insufficient. Preserve repository-specific settings and existing task
bindings; record any conflict or explicit owner deferral rather than silently
leaving the live guidance stale. This applies to every adopting repository.


## Naming, dispatch and relocation gate

This procedure preserves ownership/isolation and relocation safety. Naming is a preference, never a dispatch blocker.

1. **Coordinator name.** `0-agent-<id>` is reserved for the single active persistent coordinator, for example
   `.claude/worktrees/0-agent-7f3a2c1`. Its existing agent ID is kept when available. No worker or second session
   uses the `0-` prefix.
2. **One active coordinator,** verified through complete session discovery and the owner-selected pin (the
   helper's `status` is `pinned-live-verified`). An unknown or ambiguous result never justifies another
   coordinator.
3. **A name grants no authority.** Directory names are organizational. They never establish session identity,
   select a routing target, grant authority or justify a new or replacement coordinator. Routing and ownership come
   from discovery, the verified pin and recorded task ownership.
4. **Worker names and roots.** Prefer a task-prefixed isolated worktree named
   `<task-id>-agent-<agent-id>-<role>` in lowercase, with a concise role such as `plan`, `check` or `impl`:
   for example `.claude/worktrees/g002-agent-4be91d0-impl`. Claude workers use `.claude/worktrees/`; Codex workers
   that genuinely need isolation use `.codex/worktrees/` with the same convention. Read-only supervisor workers and
   reviewers need no separate checkout. Generated worktree and session state stays private and gitignored.
5. **Before dispatch,** record the task ID, role and preferred path when supported; use native isolated-worker
   creation if custom naming is unavailable.
6. **Before any task write,** verify and record the actual worktree path, branch, HEAD, agent/session owner and
   task/role. Wrong ownership, branch or unsafe isolation fails; a generated basename alone does not.
7. **Unsupported naming.** Treat Claude's inability to set task/role prefixes as a harness naming bug/limitation
   pending resolution. Continue immediately with its native generated `agent-<id>` worktree and the verified
   task/role mapping. No task-specific exception, extra approval or operator prompt is needed solely for naming.
   Do not create another coordinator, force a move, pre-create an inaccessible checkout, restart a session, add
   hooks or bypass permissions to enforce the convention. Preserve active bound paths. This supersedes older
   mandatory-prefix/exception gates without changing actual isolation, ownership or native permission checks.
8. **Relocation refusal.** No worktree is ever moved beneath a live process or session, and no session is restarted
   or duplicated just for naming. The default is **defer**. Relocating an existing coordinator to `0-agent-<id>`,
   or any rename, proceeds only when **every** precondition is *positively* verified; anything unknown, ambiguous,
   inaccessible or unverified defers:
   - **Complete discovery:** the helper's `status` reports complete discovery, never `discovery-unknown`.
   - **Positively verified stopped ownership:** `status` is `pinned-not-live` for the pinned UUID **and** the CLI
     listing shows that UUID only as a stopped/done row, and no live registration (under any UUID, including a copy
     or fork) has its checkout at or under the path. `pinned-ambiguous`, `pinned-unverified`, `pin-unsafe`,
     `unpinned`, `discovery-unknown`, an invisible PID, a restricted namespace or the mere absence of
     `pinned-live-verified` is not proof; a missing positive stopped row also defers.
   - **No cwd users and no open-file users, established separately:** *cwd* by reading `/proc/<pid>/cwd` and
     `/proc/<pid>/root` for every process; *open files and mappings* by `/proc/<pid>/fd/*` and `/proc/<pid>/maps`,
     or `lsof +D <path>` where installed. A cwd scan alone never proves the absence of open-file users. Either check
     is unproven, and the move defers, if any process belongs to another user, any entry is unreadable, a PID is
     invisible, a container could mount the path, or `lsof` reports access errors.
   - **Quiet state:** no validation run or helper invocation is active; work and bindings are clean or preserved.
   - **Supported CLI:** the current CLI's same-UUID restart from the moved path is verified locally and recorded in
     the local session-handoff verification record (`docs/agents/CLAUDE-SESSION-HANDOFF.md`).
9. **Relocation procedure,** only after item 8 passes in full:
   1. `git worktree move <old> <new>`, then verify with `git worktree list --porcelain`.
   2. Restart the same UUID from the new path through the recorded supported route.
   3. Reconcile the session path and pin provenance: `status` reports the checkout change since the pin; re-verify
      or re-pin, recording the replaced provenance.

   Any failure stops and is reported, with no further moves, restarts or session creation.

The gate adds no naming framework, lock bypass, process-scanner code or helper change. Verification uses Git,
standard `/proc`/`lsof` inspection and the helper's existing `status`.

## Handoff contents

### Coordinator final status line

For ordinary replies to direct owner prompts, omit the `COORDINATOR_STATUS` footer unless the owner explicitly asks for supervisor relay. Exception: an owner decision that changes active work must immediately emit a `decision` footer, including when the same message answers the owner. Do not wait for another turn, worker handback or review-ready commit. Autonomous handoffs and required reviews, failure records and blockers still notify; existing legacy-prose/native-question detection is unchanged.

Finish autonomous coordinator-to-supervisor handoffs with one plain final line (outside Markdown/code blocks):

```text
COORDINATOR_STATUS: review-ready | G008 | Plan ready for supervisor review.
```

Use a nonsecret task ID and a nonempty single-line summary of at most 240 characters, without `|`.
This is an authoring limit, not a notification rejection rule: the observer classifies the full summary within
its existing transcript-row limit, then truncates the delivered excerpt to 240 characters. Oversized summaries
must not silently lose their notification or hide pending handoffs beyond the excerpt cutoff.
Classes: `working` for routine progress, acknowledgment or recovery (no milestone notice);
`decision` for a new direct owner decision affecting active scope, priority, inputs, validation, gates or authority;
`review-ready` for a Plan, Result or execution binding needing review; `blocked` for an unavailable
prerequisite or native refusal; `question` for a specific supervisor/owner decision; `failed` for a
new validation failure needing attention; `completed` for verified terminal closeout, never self-acceptance.
For concurrent work, pending owner/supervisor decisions (`question`) and submitted reviews (`review-ready`)
take priority over routine work: name the task needing attention, not the task whose tests are still running,
and summarize other tasks in the body. Use `working` only when no actionable handoff is pending.
The observer also catches explicit pending decision/review wording in a `working` marker's full summary;
it does not scan historical body prose through a valid `working` marker. This fallback prevents a mixed-task
summary from hiding a handoff; it does not replace truthful classification.
Report failed history honestly in the body without labelling a recovery instruction as a new failure.

Emit `decision` as soon as a direct owner instruction changes active work; do not
hide it under `working` or wait for a review-ready commit. In the existing task
record/body, include the instruction, source message ID/time or committed decision
reference, what prior assumption/acceptance it supersedes and the supervisor's
needed reconciliation. Use a compact final marker, for example:

```text
COORDINATOR_STATUS: decision | TASK-ID | Owner changed validation; ref owner-message-id; supervisor reconcile the current gate.
```

This event reports a decision already made; `question` still requests a decision,
and `review-ready` submits an actual artifact. Record the current working state
in the body. Send the decision nonblocking and continue within actual authority;
acknowledgment is not a new gate. The supervisor verifies the source, reconciles
Goal/Spec/Plan/progress and stale assumptions, then sends any needed action without
polling. The notice itself grants no authority and bypasses no native gate.

The companion observer otherwise prefers a valid final marker; missing/invalid markers retain legacy prose detection.
Native question detection and source/echo exclusions remain. Markers are reports, not authority or proof;
the supervisor verifies records before acting. Keep parsing/tests in `claude-session-handoff`.

A handoff carries Goal/Spec/Plan references and revisions, phase, effort, file ownership, allowed writes, budget,
required validation and stop conditions, and asks for acknowledgment. It never copies newer inputs into a bound
checkout outside the accepted reconciliation procedure. Durable context goes into management records, not chat.

## Compact repair handoff

For a failed batch, correction or escalated diagnosis, send one short handoff using
existing Plan/Result/review records. Include only the context needed to act:

- Task, repair owner, current Goal-level attempt and remaining authority/stop gates.
- Failed check/command, expected versus actual outcome, and whether the suspected
  fault is application, test, environment or still unknown.
- Exact candidate/target revisions and the relevant diff or patch reference;
  identify material configuration/environment differences when relevant.
- Focused sanitized error/trace excerpts and links to the complete hosted artifacts.
- Fixes already tried with outcomes, plus the current hypothesis clearly marked
  as a hypothesis rather than a confirmed cause.
- Next smallest reproducer or discriminating test, the proposed bounded repair
  and the required guarantee/retest. Identify any actual scope decision needed.

Reference unchanged context rather than copying it. Unknown facts stay unknown;
do not collect unrelated inventories to fill this checklist. Keep full logs out
of chat/handoff payloads and follow existing hosted-artifact/local-cleanup rules.
The receiving agent uses the same Goal budget and file/permission scope; diagnosis
does not authorize wider self-healing or automatically require a Spec amendment.
No new JSON manifest, packet file, worker role or workflow stage is required.


## Concise operational messages

### Mandatory short reports — owner direction, 2026-10-18

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

Apply equally to coordinator and supervisor. Required formal evidence belongs in
artifacts, not long announcement messages. No new timer, monitor or parser change.

### Update frequency — owner refinement, 2026-10-18

Default to event-driven updates: a meaningful completed milestone, changed blocker,
new failure, decision request or material ETA change. Answer owner status questions
promptly. During long active work, give brief progress at the harness-required
cadence; do not turn that into a detailed multi-task report. Routine receipts,
policy acknowledgments, dispatch confirmations and unchanged waiting states need
no separate owner-facing report. Batch related changes into one update; state
only the delta and link settled context. Keep actionable failures/decisions explicit
and preserve requested detail in the durable Result or on request. Do not repeat
full hashes, inventories, unchanged task rosters or already-recorded gates in every
message. Required review handbacks and machine-readable status remain truthful;
this does not suppress failures or change observer parsing. Apply to coordinator
and supervisor alike, without a new timer, monitor or reporting framework.

## Enforce existing usage limits

When allowance use is excessive, inspect available coordinator/worker request usage
and cache counters before attributing cost. Generated tokens include reasoning;
cache totals and context sizes are not the provider's billed allowance percentage.
Use existing Results and handoffs, not a second monitoring framework.

- Enforce the existing 150k checkpoint and 200k safe handoff threshold. A new
  substantial phase above the threshold needs compaction/transfer or a specific
  completion-imminent exception. Preserve a safe running batch and its evidence.
- Default to one implementation worker. Nested delegation needs a concrete
  independent benefit and bounded inputs; do not duplicate diagnosis or reporting.
- After mandatory initial reading, continuations read relevant changed sections,
  source and open findings. Do not reload complete historical Results/reviews.
  A reading-only phase must produce a patch, minimal reproducer or explicit blocker
  before another substantial exploration phase is started.
- After a verified clean terminal handback, ignore duplicate completion notices
  unless there is evidence of new post-close writes. Do not repeat host inventories
  or policy acknowledgments for unchanged state. Preserve requested owner detail.
- Record each disposition once in its review/Result. PROGRESS and the supervisor
  handoff retain current state and links; archive superseded history instead of
  copying successive summaries into every current-state document.

Apply to active work at safe boundaries without resetting attempts, rerunning
unchanged checks, weakening required proof or changing native permissions/models.
If compaction is owner-only, report the precise supported action once; do not
claim that a policy note or checkpoint compacted the actual session.

## Coordinator context compaction at safe boundaries

At task handoffs/closeout and when usage shows conversation history dominating
context, assess whether obsolete messages/tool output outweigh the current working
state. Prefer targeted, harness-supported compaction at the next safe boundary;
do not wait solely for the context window to fill. Avoid frequent tiny compactions
or claiming a guaranteed quota saving from the context meter alone.

First preserve concise current state in existing management records: task/owner,
checkout and candidate SHAs, approved Plan/input references, decisions, open
findings, active batches, queued handoffs and remaining gates. Keep raw logs out
of routine coordinator messages and reference focused reports instead of repeating
them. Resume from those references rather than reloading every unchanged artifact.

Compaction keeps the same session UUID, checkout, coordinator pin and observer
binding/cursor/dedup state. Use supported summarization; do not clear/replace the
session merely to compact, purge its JSONL, reset budgets or alter approvals. Preserve a running
validation batch and its bindings; compaction alone requires no rebase or retest.
If supported compaction is unavailable, report that limitation and leave state
intact. No new watcher, recurring model poll or compaction daemon is needed.

Context composition is not usage attribution: investigate actual coordinator/
worker usage and cache behavior before changing models or pruning useful skills.
Deferred tool definitions shown separately are not the active message-history total.


## Coordinator checkpoints and retirement

Use **150,000 active context tokens** as the coordinator checkpoint target and
**200,000** as the compaction-or-handoff threshold, respecting smaller harness
limits and recorded owner overrides. Measure current working context, not total
transcript bytes or cumulative/cache-read tokens. Check at existing task/phase
boundaries; do not add polling. If measurement is unavailable, record unknown and
use the qualitative history assessment above.

At 150k, safeguard concise state and arrange supported compaction at the next safe
boundary. At or above 200k, do not begin another substantial phase until compaction
or a controlled replacement has been assessed. Preserve a safe running batch;
record a short completion-imminent exception when appropriate. This is not a hard
kill or automatic retirement solely for crossing a counter. Reassess at epic
closeout. Prefer retirement when compaction leaves excessive obsolete working
context, repeated stale decisions/missed handoffs persist, or the session/harness
is unstable. A healthy compacted coordinator can continue under the same UUID.

For retirement, reuse the existing supervisor handoff/Result: current authorities
and stops, task/checkout/SHAs, Plan/input references, budget, open decisions, active
agents/resources, pending messages and next gates. Finish surviving child work,
or verify a supported explicit transfer before changing its parent; never assume
workers can be adopted by another session. Establish one quiescent old coordinator
before granting a replacement ownership. Preserve unique work and existing history.

Automatic replacement is allowed only under recorded session-creation authority
and a qualified native launch/cutover procedure. Verify complete discovery and
one coordinator, the replacement's role/checkout/model/effort and acknowledgment;
update the verified coordinator pin and explicitly rebind the observer to the new
UUID while retaining delivery history and deduplication. Do not replay old notices
as new work. Do not fork the full oversized transcript, bypass a creation gate or
native prompt, or automatically resume the retired coordinator. If native child
transfer, launch or observer cutover is unsupported/uncertain, defer replacement
and report the specific supported owner action once. Preserve independent running
work. No new daemon, automatic privilege change, test restart or budget reset.

This policy documents the lifecycle; it does not implement automatic session
replacement or retire the current coordinator. Until the adopting repository has
qualified that route and recorded its authority, use supported compaction and
prepare a handoff rather than claiming automatic cutover is available.


## Supervisor session continuity

The same checkpoint, compaction-first and safe-retirement principles apply to
supervisor sessions, including Codex supervisors and Claude supervisors. The
150k/200k defaults apply only when the selected model/harness supports that working
window; honor an earlier native compaction boundary or smaller limit. Never raise
or invent a context-window setting to fit policy numbers. Record unknown occupancy
as unknown. Use native automatic/manual compaction where supported; do not infer
that a long saved transcript means the active context exceeds a threshold.

Before a supervisor handoff, preserve current mandates/stops, reviews and exact
SHAs, budgets, pending deliveries/decisions and active delegates in the existing
supervisor handoff and progress records. Complete or explicitly account for
surviving delegates; a new supervisor does not automatically adopt them. Transfer
one supervisor role and explicitly verify/rebind the observer's supervisor thread
before retiring the old route. Preserve coordinator identity when only the
supervisor changes. Resume/fork is not assumed to produce a fresh bounded context;
new sessions receive concise durable state rather than the complete transcript.
Do not self-launch a second supervisor or claim a current client can switch its
own thread unless that native route and authority are verified. Report a required
owner/client action once; no new monitoring agent or forced live restart.

A recorded owner transition exception can defer policy-driven retirement through
a named task/phase, even above the normal threshold. Preserve the active session,
children and work; use only non-disruptive preparation until that boundary. At the
first safe boundary afterwards, reassess actual context and supported compaction
before deciding on retirement. The boundary is the earliest reassessment point,
not an instruction to kill the session immediately. Exceptions do not waive native
context limits, security stops, task budgets or acceptance gates.


Owner/client navigation may be a manual final step: after a replacement thread
and its handoff/observer route are verified, the owner can open that chat in the
client. Do not make automatic UI switching a prerequisite or claim it is supported
without verification. Opening the chat grants no additional task authority.


## Worker context and phase handoffs

Default worker checkpoint target: **150,000 active context tokens**; upper handoff
threshold: **200,000 active context tokens**, configurable by explicit owner
direction or a reviewed Goal/dispatch override. Around 150k, use the next safe
step/batch boundary to safeguard work and prepare a concise handoff, leaving room
to transfer before 200k rather than beginning another large phase near the limit.
Checkpointing alone does not require immediate rotation or a new review. This measures
current working context, not cumulative input/output/cache-read usage or transcript
file size. Respect smaller model/harness context limits. At an observed threshold,
finish the current safe step or running test batch, safeguard the checkpoint and
hand the remaining work to a fresh bounded worker before more substantive work.
A fresh worker receives concise state/references, not a full-transcript fork that
recreates the oversized context. Rotation alone neither spends nor resets a Goal
attempt; actual implementation/correction rounds retain their normal accounting.

An earlier useful handoff is allowed. If completion is imminent, record a short
exception bounded to the finishing step/phase in the existing Result/dispatch;
do not carry the exception into another lengthy phase. Check available native
context metadata at existing phase/batch boundaries; no new polling agent or
monitor. If unavailable, record unknown and use the qualitative context assessment
below rather than claiming the cap is enforced or blocking otherwise safe work.
The persistent coordinator retains its separate compaction policy and identity.

At a completed implementation phase or prolonged validation/diagnosis boundary,
assess whether the worker is carrying substantially obsolete history. When useful,
hand the next bounded role to a focused worker using the existing compact repair
handoff and approved task records, rather than forwarding the full transcript.
First safeguard scoped work and record candidate/target SHAs, approved inputs,
findings, completed/failed checks, resource ownership and next action. Verify one
write owner, checkout/isolation and any unfinished resource handover before dispatch.
Reuse attributable evidence; a replacement worker does not restart tests or budgets.

Do not require fresh agents per phase or batch, rotate a useful ongoing owner solely
because time passed, or add review rounds. Agent setup/re-reading also costs usage.
Keep the single coordinator; supported compaction of a continuing worker may suffice.
Use the existing phase-specific model/effort policy and supported controls. A saved
model default or dispatch instruction is not independent runtime verification;
record actual metadata when available and label owner-reported/unknown settings.
If compaction or effort selection needs an owner-only native command, report that
specific action once at a safe boundary. Do not bypass it, repeatedly prompt, or
block otherwise authorized unrelated work merely for that optimization.

Combine this with existing concise messages, targeted reruns, delta-only source
reviews and duplicate-CI safeguards; do not create another framework or checklist
artifact. Inspect available usage attribution/cache statistics when consumption is
unexpectedly high. Context occupancy, elapsed test time and tool-call counts alone
do not establish weekly allowance shares. Compaction consumes usage itself, does
not restore allowance and has no guaranteed saving. Record useful observations in
the normal Result/closeout; no new benchmark campaign or reporting requirement.


## Lower context thresholds and coordination economy — owner direction, 2026-10-20

Use **150k active context tokens for checkpointing and 200k for safe handoff/compaction assessment**, superseding the former300k/350k defaults for workers and coordinator checkpoints and the shared supervisor default. Smaller native limits still prevail. Measure current working context, not transcript bytes or cumulative/cached token totals; report unavailable measurements as unknown. Adopt at the next safe step/batch boundary; preserve running validation, exact source/evidence, current task owner and budgets. No forced interruption, full-transcript fork, duplicate writer or retest merely to adopt the limits. Record only a bounded completion-imminent exception when needed.

**Claude and Codex may compact their own sessions at safe points without another owner prompt**, using a supported native agent-accessible control or harness-managed automatic compaction. Safeguard concise state first. If the actual interface exposes only an operator control, report that limitation once; authorization does not create a tool or permit fabricated owner input. Preserve active children and one coordinator; no automatic session replacement or new monitor.

Reduce avoidable supervisor and coordinator overhead:

- Handle routine policy prose, mechanical references and bounded path changes directly with focused checks; do not create a delegated review solely because a small edit exists. Keep independent review for substantive architecture/security/data changes and required Plan/Result/source gates. Batch related findings into one handoff rather than repeated review rounds over unchanged material.
- Batch independent searches/reads and related configuration/documentation edits. After initial required reading, inspect relevant deltas/open findings; do not repeatedly reload whole Results, transcripts or unchanged governance. Give an independently useful delegate only bounded task inputs and necessary references, avoiding full-history inheritance when unnecessary.
- Check a running task only at a meaningful event, failure, owner request or agreed deadline. Use existing asynchronous completion; no repeated unchanged polls, duplicate notice acknowledgments, receipt/status recaps or extra monitoring agent. A failed/pending required check remains visible and receives prompt action.
- Report new outcome, next action and an actionable blocker once under the short-message rule; keep exact evidence in existing artifacts. Publication, receipt, acknowledgment and actual execution remain distinct. A link-only status is not a substitute for explaining a material failure.

Judge savings from observed usage/cost where available; context-size/cache counters or API-equivalent estimates alone do not establish subscription depletion or actual bills. This is an operating rule, not a new telemetry system, blanket review waiver, model change or reduction in required validation.


## Authorized self-compaction at safe points — owner amendment, 2026-10-20

Owner prefers Claude and Codex compact themselves. This supersedes the earlier repository owner-only `/compact` rule. Standing permission covers native self-compaction and ordinary harness automatic compaction; no routine confirmation is needed when the agent's interface supports it. Keep150k checkpoint/200k compaction-or-handoff assessment and smaller native limits; these are workflow thresholds, not a claim that the harness has been configured to compact at those exact values.

At a completed step/phase boundary, safeguard current authority/stops, exact checkout/source/input references, reviews, budgets, open findings, active children/resources, pending deliveries and next actions in existing records. Use only the native control exposed for that session. Preserve running tests/evidence and one owner; do not interrupt a blocking test to compact or launch a replacement agent merely for context reduction. Resume by reading the compact state and relevant deltas, not the whole old transcript.

Documented product capability is distinct from this live session's tools: Claude Code has automatic compaction and an interactive `/compact`; Codex app-server documents `thread/compact/start`. Neither proves every model-facing interface exposes a self-compaction tool. If unavailable or natively refused, state that concrete limitation once and use supported automatic compaction or operator action. No transcript deletion/edit, `/clear`, hidden API/session injection, permission bypass, duplicate session or pretend owner keystrokes. Do not report “compacted” without an observed native completion; a written checkpoint alone is not compaction. Existing session replacement, observer identity, failure/review/cleanup gates and budgets remain.

References: [Claude Code context](https://code.claude.com/docs/en/how-claude-code-works#the-context-window), [Codex app-server compaction](https://learn.chatgpt.com/docs/app-server#trigger-thread-compaction). These document native capabilities, not a new local controller or execution claim.


## Codex supervisor model and delegation defaults — owner direction, 2026-10-20

The owner may use a lower-cost model in the main Codex supervisor session while delegated Goal/Spec drafting and substantive architecture/security/Plan/Result reviews default to **gpt-6.1-sol**. Use supported repository-scoped Codex configuration:

```toml
[agents]
default_subagent_model = "gpt-6.1-sol"
```

Keep the main model and reasoning effort owner-controlled; setting a child default does not change them. Explicit approved per-agent model overrides take precedence. In a live harness that does not reload project configuration or whose full-history forks inherit the parent model, explicitly select Sol for a fresh bounded delegate where supported; do not assume or claim an effective model change from a file edit alone. Verify actual dispatched model when available, otherwise report unknown. Native availability/trust/managed constraints remain; no extra persistent Codex session or artificial test spawn is required. Account usage remains shared, not an independent allowance per agent.

**Routine supervisor work uses no subagent by default.** Handle status, progress/handoff publication, receipt inspection, simple policy prose, configuration and mechanical path updates directly in the main session with batched focused checks. Delegate Goal/Spec drafting and substantive independent reviews when they provide the required quality/independence; avoid a second main-session re-review of unchanged work. Preserve supervisor ownership of final dispositions and mandatory security/validation gates. Give delegates only necessary inputs/references and bounded writable scope for drafting; reviewers remain read-only. Reuse useful existing findings, never duplicate writers or fork a huge transcript just to dispatch a small task.

This assigns a model to appropriate delegation, not a requirement to create agents for every task. Portable adopters choose their supported child model explicitly; the model ID above is this owner's selected example. It is not a bundled account credential, permission grant, automatic model migration or authorization to replace a supervisor. Claude implementation models remain independently governed; Claude interim supervision is not activated merely by this setting.

Official configuration: [subagent model default](https://learn.chatgpt.com/docs/config-file/config-sample), [project configuration/trust](https://learn.chatgpt.com/docs/config-file/config-basic).

## Continuous cleanup checkpoints — 2026-10-22

Track owned resources at allocation and removal in the existing run/task inventory;
reconcile before startup/resume/dispatch, at every handback and batch end, and after
failure/cancellation. Reclaim finished worker trees and scratch immediately after
preserving their handback. Any unresolved abandoned resources block new resource-
consuming work until repaired; verify actual absence, including nested resources
and anonymous volumes. Follow [continuous accountability](closeout-risks.md#continuous-resource-accountability--mandatory).
This adds no coordinator polling or duplicate inventory framework.
