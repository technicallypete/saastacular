# Budgets and proportional process

Budgets detect thrashing, scope growth and failing strategies. They are maxima, never productivity targets, and
never a reason to weaken validation, security or acceptance criteria. The adopted policy file is instantiated from
`templates/management/BUDGET_POLICY.md.tmpl`; the default class is the `DEFAULT_BUDGET` parameter.

## The unit: a Goal-level execution attempt

- One attempt is one Goal-level try at the outcome. **Commands, test batches, inner implement/test/fix loops,
  rebases, handoffs, checkpoints, context compaction and session restarts do not consume attempts.**
- The initial authorized implementation is attempt 1. An attempt ends when a review-ready Result is submitted,
  when the attempt is reported unable to satisfy the Goal, or when its material strategy is abandoned.
- A new attempt is consumed by a retry after an unsuccessful attempt or by a supervisor-authorized correction
  round. Calling an abandoned strategy "the same attempt" is not allowed.
- **Workers share the allowance.** Subagents and parallel workers contribute to the parent attempt; nobody gets
  an independent counter, and their loops are not summed into the total.
- Planning consumes no implementation attempt; neither do Plan corrections.

## Classes and checkpoints

| Class | Maximum attempts | Checkpoint | Typical use | Typical delegation |
|---|---|---|---|---|
| B3 | 3 | after attempt 2 | narrow configuration, small bug fixes, isolated presentation work, tooling | 0–1 |
| B5 | 5 | after attempt 3 | ordinary features, bounded infrastructure, integrations | 0–2 |
| B7 | 7 | after attempt 4 | authorization, isolation-sensitive work, migrations, cross-service work | 1–3 |
| B15 (optional, exceptional) | 15 | after attempts 8 and 12 | intentionally coupled foundation work | 2–4 |

The supervisor assigns the class in the Goal; the implementer may recommend a change during planning but never
changes an approved budget. A checkpoint is a reassessment, not an automatic stop: criteria done and remaining,
failures and repeated patterns, uncertainty, convergence and whether the Goal is still well bounded.

## Exhaustion and extensions

On exhaustion before all criteria pass: stop, preserve the working state, keep every failing criterion, update the
Result (attempts used, contributions per attempt, criteria satisfied and remaining, blockers, attempted approaches,
recommended next action) and return control to the supervisor.

An extension is explicit and exceptional. Its record states: previous budget, new budget, attempts already used
(historical counts preserved), reason, remaining work and the approving supervisor. **Counters are never reset**,
and an extension never authorizes scope, architecture or acceptance changes.

## Result ledger

Every Result carries an Execution Budget section (see `templates/management/release/results/RESULT.md.tmpl`):

```text
Budget class: B3   Maximum attempts: 3   Attempts used: 2 (second active)
Attempt log:
1. Initial implementation; batch B1 failed twice (fixture path, parser edge case), fixed and rerun;
   worker review contributed in this attempt; Result submitted; supervisor returned CORRECTION LOOP.
2. Authorized correction; rerun of affected batches; resubmitted.
Validation history (not counted): every failed, skipped and interrupted run with its evidence path.
```

## Effort

Effort is a **configurable recommendation** by role: `COORDINATOR_EFFORT` (default `high`) for orchestration,
status and handoffs; `PLANNING_EFFORT` (default `xhigh`) for planning workers, and for a coordinator that plans;
`IMPLEMENTATION_EFFORT` (default `high`) for implementation workers. Set worker effort explicitly through supported
controls where available instead of assuming inheritance. The owner may raise complex debugging or security work
to xhigh. Before a phase, state the intended effort and remind the operator; honor an already stated
choice. Verify actual runtime effort only where session metadata exposes it, and otherwise record "unknown". A
handoff's wording never sets or proves the runtime setting, and effort changes never reset budgets or completed
work.

## Workload-based model selection

Model choice is independent of effort and budget. The owner-approved Claude
default is **Fable 5.1 high for planning and test corrections**, including failure
diagnosis and corrective test/fix work. Default to **Opus 5.5 high for
implementation and ordinary test execution/reporting**. Default the Claude
coordinator to **Sonnet 5.5 high**; the owner may choose **Opus 5.5 high** for
coordinator work. The owner
may choose Opus for planning/test corrections; then use **xhigh**. Fable xhigh is
selective for especially difficult work. Other providers may use equivalent
owner-selected tiers; verify availability before adopting examples elsewhere.

Apply the configured role effort separately. State model/effort before a new phase,
honor owner choices and verify actual configuration where supported, otherwise
record unknown. Never assume an instruction changed a worker's runtime. Do not
interrupt an active batch, redo completed work, reset budgets or weaken gates
solely to change models. A stronger model does not accelerate the underlying
Docker/build/test process. These recommendations add no provider dependency or
automatic configuration change.

## Proportional process

An owner range/epic mandate never combines, resets or extends individual Goal
budgets. Assign each newly specified Goal its normal class; exhaustion and
owner work/phase/endpoint stops still hold while a broader mandate is unfinished.

Scale Plan and evidence detail to scope and risk ([risk tiers](adoption.md#risk-tiers)). Reference unchanged
accepted controls by path and version instead of rebuilding them. After the initial reading, re-read only changed
sections and unresolved findings. A new execution framework or orchestration library is separate scope, not an
incidental prerequisite.

## Model-aware effort precedence — owner direction, 2026-10-15

For Claude Fable 5.1 use **high** for planning, reviews, diagnosis and correction
loops; use **xhigh** selectively for especially difficult architecture, security
or concurrency work. When the owner chooses **Opus 5.5 for planning or test
corrections**, use **xhigh**, including failure diagnosis and corrective test/fix
work. Opus routine implementation and ordinary test execution/reporting remain
**high**. Coordinator default is Sonnet 5.5 high; the owner may choose Opus 5.5 high. A correction phase is distinct from
merely running an unchanged test command.

These model-aware values supersede blanket planning-xhigh or implementation-high
wording for the affected phases, without changing generic parameter fallbacks.
Owner choices take precedence. Use supported runtime controls and verify actual
or inherited model/effort, otherwise report unknown; prose does not configure a
worker. Apply at the next phase or safe boundary, never by interrupting a running
batch or restarting completed work. Preserve budgets, evidence, permissions and
bindings; do not switch the current G013 worker mid-batch.

Fable test-correction workers may escalate from high to xhigh when diagnosis is
uncertain, a fix fails again, or the correction becomes substantially more
complex. This owner-authorized effort increase needs no repeat approval. Apply
through supported controls at the next safe boundary and record the reason and
actual setting (or unknown) in the Result. Do not interrupt a running batch.
Existing repeated-issue escalation and budget stops still apply; higher effort
does not reset attempts or expand scope.


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
