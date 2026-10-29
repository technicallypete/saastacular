# Roles, artifacts and lifecycle

This reference defines who owns each management artifact, what each artifact contains and how a Goal moves from
proposal to final acceptance. Actor names are parameters (`OWNER_ROLE`, `ARCHITECTURE_AUTHORITY`,
`SUPERVISOR_AGENT`, `IMPLEMENTER_AGENT`); the adopted repository's governance files hold the concrete names.

## Precedence

1. Harness and platform instructions, then explicit owner direction.
2. The adopting repository's canonical governance, architecture and accepted decisions.
3. This skill's references and the instantiated templates.

A skill or template never overrides a higher-priority instruction. A policy change records its scope and its
authorizing actor truthfully; nobody widens their own authority by editing a policy file.

## Ownership matrix

| Artifact or responsibility | Owner | Purpose |
|---|---|---|
| Product and architecture direction | Owner (with the architecture authority) | Canonical decisions; major changes are decided here and never silently by an agent. |
| Roadmap, Goal, Work Spec | Supervisor | Dependency-ordered Goals, bounded outcomes, requirements and acceptance criteria. |
| Plan review, implementation acceptance, progress | Supervisor | Review against scope, architecture, security, budget and every criterion; maintain PROGRESS. |
| Implementation Plan, implementation, tests, Result | Coordinator / implementer | Concrete approach, code, validation and the evidence record. |
| Tactical workers | Coordinator / implementer | Bounded, non-overlapping delegated work; the implementer integrates and validates it. |

Workers, reviewers and subagents never hold owner or supervisor authority. A delegated opinion does not replace
supervisor acceptance or an owner decision.

## Owner-selected supervisor

The owner may select Claude Code as supervisor. In this mode, Claude owns orchestration, Plan review, implementation acceptance, progress and closeout through the existing workflow; ChatGPT/Codex contributes only Work Spec writing and architecture reviews. Do not retain a second ChatGPT/Codex supervisor, routine approval gate or monitoring/dispatch duty. This explicitly overrides the default supervisor-owned Spec row above. Major product/architecture decisions stay with the owner. Record the selected assignment in existing repository governance/handoff; documenting the option alone does not activate it or restart/rebind active work.

## Artifacts

All Goal artifacts live under `management/<release>/` with a stable Goal ID (default form `G###`) and a
descriptive stem: `goals/<ID>-<stem>.md`, `specs/<ID>-<stem>-spec.md`, `plans/<ID>-<stem>-plan.md`,
`results/<ID>-<stem>-result.md`, and review records such as `results/<ID>-implementation-codex-review.md`.
IDs are never renumbered or reused. Revisions append dated records or create linked revision files
(`<stem>-plan-r1.md`); prior approvals, failures and evidence are never erased.

**Goal: eleven fields, in this order.** 1. Goal ID and title. 2. Objective. 3. In Scope. 4. Out of Scope.
5. Dependencies. 6. Completion Definition. 7. Execution Policy (including the risk tier). 8. Iteration Budget.
9. Context / Scope Budget. 10. Suggested Subagents / Delegation. 11. Stop / Escalation Conditions.

**Work Spec:** requirements, architecture constraints, security invariants, expected artifacts, acceptance
criteria with verification, required tests and validation, and explicit non-goals. A Spec states what must be
true, not how to code it.

Record high/low priority in the Goal/Spec with its source/rationale. Honor an
explicit owner choice; the owner should identify low-priority work. If unspecified,
the supervisor determines it or asks when needed, normally defaulting to high.

**Plan:** approach; files/modules; schema/migration changes (or none); contracts and public-interface
implications; tests and validation batches; delegation and worker allocation; sequence; a
mapping of every acceptance criterion to concrete changes and evidence; risks and escalations. Scale detail to
the change and reference accepted controls by path instead of copying them.
Plans target 3–5 parallel implementation workers for high priority or 1–3 for low
whenever independent work, ownership, native/resource capacity and budget permit;
record concrete constraints reducing the count. Exclude planner/coordinator/
reviewers. Kickoff follows the accepted allocation/sequence after readiness checks.
The planner may remain one agent. See [parallel scheduling](coordination.md#default-to-safe-parallel-work-lanes).

**Result:** status (submission is not acceptance); summary and files changed; schema and contract changes; tests
added; exact commands with outcomes, including failed, skipped and interrupted checks; the acceptance checklist
with evidence; deviations (or none); blockers, attempted approaches and the attempt ledger
([budgets](budgets.md)); an outstanding-findings checklist; suggested follow-up, which authorizes nothing.

**PROGRESS:** the mutable index of active, completed, blocked and upcoming Goals with links and review outcomes.
**BACKLOG:** proposals, not authorization. **ROADMAP:** dependency order and any delivery barrier.

## Four distinct states

| State | Meaning | Who |
|---|---|---|
| Plan approval | The Plan may be implemented within its scope and budget. | Supervisor (ACCEPT on the Plan) |
| Readiness | The candidate passed its required local evidence and review; it may be integrated. | Supervisor pre-integration review |
| Conditional integration | Merged to the integration branch with required post-merge evidence still pending. | Supervisor executes under the adopting owner's recorded merge policy |
| Final ACCEPT | The Goal is complete; dependents may proceed. | Supervisor, after the required integration disposition |

Only final ACCEPT completes a Goal. A merge, a passing pipeline or a delivered message is never acceptance.

## Lifecycle

1. The supervisor selects a dependency-ready Goal and writes the Goal and Spec. Major questions go to the owner.
2. The implementer inspects the repository and writes the Plan, unless the Goal explicitly authorizes the direct
   tier ([adoption](adoption.md#risk-tiers)). If the Goal exceeds reliable context, the Plan recommends a split.
3. The supervisor reviews and returns exactly one outcome: **ACCEPT**, **CORRECTION LOOP** or
   **ARCHITECTURE DECISION REQUIRED**. The accepted Plan bytes are imported into the integration branch.
4. The implementer implements, validates ([validation and integration](validation-integration.md)), self-reviews
   and records the Result. Escalation conditions override autonomy.
5. The supervisor reviews the actual diff and evidence. Passing tests alone are insufficient.
6. Corrections fix only the named findings and append to the Result. The same material issue surviving two
   correction cycles returns to Goal/Spec reassessment or escalation.
7. After the integration disposition, final ACCEPT; then closeout ([closeout and risks](closeout-risks.md)).

**Mandatory escalations:** architecture contradictions, scope expansion, security-model changes, destructive
migrations, public API or tool-contract changes, new major infrastructure or vendors, irreversible data-model
decisions. Plan approval never covers an unreviewed material change discovered afterward.
The standing mechanical-correction authority below covers routine implementation consistency.

## Autonomous mechanical corrections

After the approved task's execution release, the implementer diagnoses and fixes
mechanical inconsistencies within its approved functional scope, file allow-list
and remaining budget without a separate supervisor approval or Plan-revision
round. This includes stale test selectors/fixtures, schema-version and count pins,
expected inventories/catalog parity for already-approved values, and missing
imports or harness file-list entries for already-approved behavior.

The intended behavior must be unambiguous in the Goal/Spec, architecture and
accepted contracts. Fix the cause, retain meaningful assertions and negative
controls, and run the smallest failed/affected selectors. Do not update an expected
value merely to match a regression or declare a failure unrelated without evidence.
Simple Plan wording mistakes may be reconciled in the existing Result/addendum;
preserve accepted Plan/history bytes and frozen running inputs. Record the cause,
exact paths, focused evidence and any reconciliation at the existing safe boundary;
continue the current task/worker rather than waiting for another routine proceed.

Escalate affected work for functional/file scope expansion, ambiguous intended
behavior, security/authority/tenancy changes, public API/MCP contract changes,
unapproved schema/constraint/grant/data or risky migration changes, major vendors/
infrastructure, irreversible decisions, budget exhaustion, explicit owner holds or
native refusals. A catalog label does not make a production constraint change
mechanical. New production semantics still require the existing review.

Supervisor initial Goal/Spec/Plan decisions and final actual-diff/evidence review
remain. Required CI/manual checks, merge authority, attempt accounting and stop
boundaries are unchanged. A fixed case is reported as fixed, not self-accepted;
routine repair progress uses working status, and only real decisions/review-ready
submissions interrupt the supervisor. No new artifact framework or test campaign.

## Bounded specification authority and owner stop boundaries

Automatic continuation does not authorize new Goals/Specs. Without an explicit
recorded owner mandate, stop at a missing/draft Spec and report the gap. The
supervisor retains Goal/Spec ownership; implementers do not invent scope.

The owner may authorize specification through an inclusive named task/Spec,
an explicit task set/range or a named epic. Record the owner instruction/reference,
allowed phases, start, exact task membership, inclusive/exclusive endpoint and
stop/resumption disposition in existing PROGRESS/handoff records. Pin the
roadmap commit: dependency order, not numeric IDs, governs; later inserted tasks
do not silently enlarge the mandate. Do not create an authorization registry.

Specification-only authority is a phase stop: it does not automatically permit
planning/implementation when the Spec becomes ready. A full-epic mandate may
explicitly include missing Goals/Specs and normal delivery, retaining separate
Goal budgets, reviews, dependencies, bindings and acceptance. Other separately
authorized tasks remain unaffected. Missing prerequisites outside the mandate
or owner-owned architecture decisions still escalate.

An independent owner stop overrides specification, planning, reconciliation
and execution, even with an existing Spec/accepted Plan. Check it before every
automatic transition, including advance planning while current CI/closeout runs.
Examples only, not current authorizations:
- "Auto-spec through G013 inclusive": the recorded specification range only.
- "Finish EPIC-004 including missing Goals/Specs, then stop": the pinned epic
  and normal gated delivery, not a rolling authorization for later epics.
- "Stop before G013": no G013 specification/planning/execution.
- "Finish G013, then stop": complete G013 acceptance/closeout, with no successor
  specification or planning during its CI overlap, even if a Spec exists.
- "Plan G013, then stop": no implementation under that phase-limited mandate.

Record the interpretation of "stop at"/"next task". If identity or before/after
meaning is ambiguous, hold before crossing it and clarify, preserving work
within clear bounds. Cancel queued continuation past a new stop; pause existing
out-of-bounds work at a safe point, preserving artifacts, cleanup and truthful
status. "Stop now" remains immediate. The earlier/more restrictive endpoint or
stop wins until explicitly changed by the owner; retain decision history and
never infer resumption from a merge/observer event or skip to a later task.
Native permissions, budget exhaustion and delivery barriers still govern.

<a id="automatic-planning-after-integration"></a>
## Automatic planning at implementation MR creation

When a predecessor's implementation MR is first opened against the configured
integration branch, automatically plan the earliest eligible next roadmap task
with an existing ready supervisor-owned Goal/Spec. Draft implementation MRs qualify;
planning-only, placeholder and probe-only MRs do not. Verify the actual MR,
source/target and candidate revision. Do not wait for CI/Duo, MR ready status,
merge, predecessor acceptance or cleanup to begin provisional planning.

Check existing task/planner ownership and recorded owner phase/stop bounds first;
reuse the assigned planner and avoid duplicate dispatches on later pushes. Missing
Specs require an existing explicit bounded specification mandate. Planning does
not invent scope or bypass unresolved architecture/owner decisions. Record the MR,
candidate baseline and pending gates in existing PROGRESS/handoff records.

Reconcile relevant predecessor revisions and integrated content before
execution; if its MR closes without integration or materially changes, reassess
readiness. Integration remains a fallback discovery checkpoint if the first MR
was missed, not the primary trigger or a second dispatch. This supersedes the
former integration-only trigger; legacy links to that heading remain valid.

Use the verified coordinator and current model-aware planning effort under the
existing single-owner/worktree rules. Scope is inspection and Plan/history only,
not application/test writes, builds, installs, shared validation resources or
unauthorized probes. Execution, including direct-tier work, follows the release
policy below; planning dispatch alone never grants implementation authority.
No new artifact, watcher, scheduler, transport or permission override is added.

## Automatic reassessment and execution release

Default: once a predecessor's reviewed source is merged, its existing coordinator
must automatically launch and own affected local qualification of that exact merged
revision. Reassess the next authorized task's approved provisional Plan against
that integration and current Goal/Spec. Isolated implementation may overlap the
local run and closeout; do not require another owner prompt or wait for final
predecessor ACCEPT solely to start it. Record merged revision and pending local
qualification in existing bindings/Result/PROGRESS. Protect frozen test inputs
and shared locks. A merge or existing Plan alone is not successor release.

Before release, verify Plan ACCEPT (or explicit direct-tier authority), current
input/scope binding, owner decisions, budget, isolated task branch/checkout and
resource readiness. Other required dependencies must be satisfied. For material
changed assumptions, request bounded implementer-owned Plan updates and review;
otherwise record reconciliation without rewriting immutable approved Plan bytes
or repeating unchanged validation. Record and relay supervisor execution release;
no additional routine owner proceed prompt is needed. Reuse the task owner and
current model-aware effort, with no budget reset.

Predecessor final ACCEPT and its required qualification disposition still gate the next
task's integration and final ACCEPT. Pending or failed checks remain pending or
failed. Unresolved security, architecture or contract changes that affect the next
task block its affected implementation. Explicit task/owner holds, manual gates,
native permissions and delivery barriers override this default; do not silently
rewrite an active Goal/Spec's stricter binding or expand a phase mandate.

If predecessor qualification fails, diagnose attributable evidence and pause only
affected or unsafe dependent work; preserve unaffected work. The predecessor's
assigned implementation owner retains responsibility for diagnosing/fixing its
staging failures, on its repair branch and existing attempt ledger. Run those
repairs in parallel with successor implementation; do not transfer them to the
successor's workers by default. Any worker reassignment must be explicit and keep
predecessor ownership, exact repair scope and focused validation. Keep repair and
successor changes attributable, prevent concurrent writers/shared-resource clashes,
and reconcile changed interfaces before dependent work resumes. No silent scope
expansion, automatic retry campaign or failed-to-pass conversion.

Planning still starts at the first eligible implementation MR; implementation
starts only after reviewed source integration and recorded release. No new watcher,
scheduler, artifact framework or blanket specification authority is introduced.

## Epic-closeout progress checkpoint

The supervisor records a PROGRESS compaction disposition at every epic closeout,
before marking closeout complete: archive superseded updates with working links,
record none needed, or explain an active-binding deferral and next safe checkpoint.
Keep current dispositions, dependencies, blockers and acceptance/evidence links visible.
Preserve history under [closeout policy](closeout-risks.md#progress-compaction);
this housekeeping alone needs no tests or new tracking framework.

## Supervisor progress publication (option, enabled by default)

`SUPERVISOR_AUTO_PUBLISH` (default `true`) with an explicit `INTEGRATION_BRANCH`. When it is enabled, the
**supervisor**, **whenever it completes a PROGRESS update**, commits and pushes that update and its associated
management records (imported Results and status, reviews, decisions) to the integration branch from the main
integration checkout, without another owner confirmation.
- It verifies imported bytes, stages exact owned paths and excludes secrets and machine-local logs.
- It fetches and reconciles first and never force-pushes. A conflict, protected-branch rule or permission gate is
  reported, not bypassed.
- Publication is not review or ACCEPT. The implementer delivers Results and status and keeps its task branch and
  merge requests; this grants no direct implementer push.
- **Opt-out:** the owner may set it to `false` for the repository, or disable it for an explicit task or session;
  the override is recorded.
- It is an agent workflow, not a watcher, commit hook or scheduler. The supervisor assesses whether the update
  touches active task bindings and notifies the coordinator at a safe point.

Routine management-only acceptance/cleanup closeout uses the same direct supervisor
publication route. The coordinator hands back its Result and receipt; the supervisor
reviews/imports them and records the final disposition. Do not create another
acceptance or Result-only MR, automated-review request or CI run just to publish it.
See [direct closeout publication](closeout-risks.md#direct-supervisor-publication-of-acceptance-and-closeout)
for actual branch-protection/owner exceptions and source-repair boundaries.

## Supervisor checkout, delegated review and automatic relay

- The supervisor normally works in the **main integration checkout**. Specifications and read-only reviews need
  no separate worktree. A supervisor worker that genuinely needs isolation follows the worker-root rules in
  [coordination](coordination.md#naming-dispatch-and-relocation-gate).
- Prefer **bounded subagents** for Spec drafting/research and formal Plan/Result reviews. Reviewers are
  read-only by default; one writer per artifact. The supervisor integrates findings, owns the disposition and
  records it. A child reviewer cannot approve anything.
- **Review-ready submissions.** When an existing handoff or status check shows a review-ready Plan or Result, the
  supervisor starts a delegated review **without waiting for another owner prompt**:
  1. Verify the exact artifact (path and SHA-256) and its phase. Unfinished work is not reviewed as final.
  2. If the same bytes already received a completed review, record "duplicate unchanged review skipped" and
     point at the existing record.
  3. If delegation is unavailable, record that limitation and review directly rather than leaving authorized work
     idle.
  4. Record the disposition in a review file (template `results/REVIEW.md.tmpl`).
- **Asynchronous completion.** For a bounded delegated review, the supervisor keeps its turn open using the
  harness's supported asynchronous agent wait/notification interface, and the owner may message or redirect
  meanwhile. It collects the findings, then finalizes, records and relays the decision **before** its final reply.
  It uses bounded waits and avoids repeated inference-based polling. Waiting uses no model tokens; reviewer work and
  result processing do. Automatic wake-up after an ended turn is not guaranteed; state that limit rather than
  promise it. No scheduler or extra monitoring agent is implied.
- **Interruption is not stop-all.** An interruption of the supervisor's turn alone does not stop delegated agents.
  On resuming, inspect the children's actual status, preserve completed findings and authorized running work, and do
  not cancel or re-dispatch automatically. An explicit stop-all instruction, or a scope, security or owner direction
  that makes their work unsafe or obsolete, still governs. Do not assume a client (for example mobile) cancellation
  reached the child agents; check.
- **Automatic relay.** After recording, the supervisor relays the decision, reviewed inputs, permitted next
  action and remaining gates to the verified coordinator, through the companion messaging skill
  ([claude-session-handoff](../../claude-session-handoff/SKILL.md)). Receipt and acknowledgment are verified and
  recorded; pending or unavailable delivery is reported, never auto-resent, and never answered by creating a
  session.
- The owner may interrupt or redirect at any time. Mandatory escalations and owner merge authority remain. This
  adds no background watcher, no transport and no implementation authority before Plan approval.


## Continue routine transitions without owner nudges

Within recorded authority, continue from a verified ready submission through review,
correction/approval relay, MR and automated-review disposition, authorized merge,
staging disposition, final acceptance and routine closeout. Reassess the next task
under the existing planning/execution rules. Do not end a handoff by asking the
owner to repeat “review”, “continue” or “tell the coordinator” for an already
covered action. Stop at actual external waits, native permissions, unresolved owner
product/risk decisions, budgets or mandate/phase limits; name the exact wait.
The observer supplies notices, not authority, and cannot guarantee wake-up of an
ended supervisor turn. This adds no daemon, scheduler or permission bypass.

At normal closeout, briefly note repeated validation, harness flakes or avoidable
owner interventions in the existing Result if material. Existing timestamps and
records suffice; no telemetry system, new metric schema or benchmark campaign.
