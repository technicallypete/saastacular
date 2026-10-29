> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G011 — Durable audit and event conventions work specification

Goal: [G011](../goals/G011-durable-audit-events.md)  
Epic: EPIC-004 — Audit and compliance foundation  
State: r0 + owner amendments — provisional planning released; implementation gated

## Rationale, authority and baseline

Operational logs describe decisions but are not the durable audit ledger.
G011 implements [ADR-006](../../../adr/ADR-006-audit-schema-boundary.md):
the existing database gains a separately owned `audit` schema with controlled
append, current-authority tenant reads and restricted maintenance. This provides
engineering evidence, not SOC 2 certification or immunity from privileged DBAs.

Inspection baseline is staging `908b84c`, after G009 final ACCEPT at `325c136`.
G010 Plan r1/binding is approved; G010 implementation/final acceptance is pending.
This is not an execution binding. Reconcile its final diff, fixtures, catalog and
reviews before implementation; no G010 isolation guarantee is assumed delivered.

Read current AGENTS/CLAUDE, architecture/security/decisions, accepted ADR-003/004/006,
ROADMAP/PROGRESS, BUDGET_POLICY and accepted G004–G010 evidence. SECURITY.md is
empty at inspection; it supplies no additional policy. Relevant contracts are
[AUTHORIZATION](../../../foundation/AUTHORIZATION.md),
[DATABASE-ISOLATION](../../../foundation/DATABASE-ISOLATION.md),
[OBJECT-STORAGE](../../../foundation/OBJECT-STORAGE.md), and
[PERFORMANCE](../PERFORMANCE.md). Earlier follow-up/skill language
placing every denial in the mutation transaction is superseded by ADR-006's
rollback requirement, not an instruction to preserve a rolled-back audit row.

## Requirements

### R1 — Finite event inventory and existing producers

Claude's Plan pins a compact catalog with action/version, producer/call site,
actor/scope source, resource, outcome, persistence point, recording-failure policy
and tests for each existing category:

- Identity/session lifecycle: successful sign-in/session issue and relevant
  authentication failure, logout/session revocation, user enable/disable and
  existing administrative session actions. Distinguish local application facts
  from unverified provider logout or assurance; do not add live-provider features.
- Existing Organization/Group creation and metadata, membership/role changes,
  BusinessProfile creation/update/association and Organization policy mutations.
  Include the accepted trusted-only model/service paths; future administration
  or onboarding UI is not a prerequisite and must not be implemented here.
- Guarded security denials and accepted tenant/selected-Group context changes.
  A request resolving the same context repeatedly is not automatically a durable
  navigation/change event. Pin event granularity to actual transitions/decisions.
- Guarded object Put/Delete and signed-GET issuance, including staged failure,
  compensation, unknown outcomes and retries. Preserve the distinction between
  authorization/issuance and observed transfer completion.

Every category maps actual existing entry points, including no-op/idempotent
outcomes, or identifies a genuinely future operation with no current producer.
Do not label an implemented mutation future to omit its required audit evidence.
Trusted fixture/bootstrap direct SQL is separately identified; it is not evidence
that authenticated user attribution works. Operational `access.decision`,
`storage.operation` and identity records remain diagnostic records with their
existing privacy contracts; do not create a log parser or copy all log lines.
Auditing trusted-only creation/membership primitives preserves their existing
invocation boundary; do not restore application-runtime grants removed by G008
or expose new end-user operations merely to exercise an audit producer.

### R2 — Event identity, bounded data and truthful attribution

Each durable event has a server-generated canonical UUID, schema/event version,
server timestamp, closed action/outcome/reason values, resource type and canonical
UUID where known, trusted actor kind/UUID where established, Organization UUID
and optional Group UUID where established, and bounded request/trace correlation
when available. Distinguish event time from any later persistence time if needed.
The Plan pins field sizes, nullability, validation and duplicate/retry semantics.
No unbounded arbitrary JSON, raw error/body or caller-defined action namespace.

Derive actor and current scope from verified identity/context or an explicitly
trusted bootstrap/system invocation. Submitted UUIDs, provider claims without
verification, headers and diagnostic correlation never confer authority. A
rejected foreign resource selector is not proof of ownership in that tenant.
Unknown actor/tenant/resource stays unknown, not a fabricated UUID or tenant.
Global identity/pre-context events have no invented Organization; ordinary tenant
readers cannot access those records. System/maintenance producers are distinguishable
from human actors and cannot impersonate a user by filling an actor field.
Independent global/pre-context recording grants no global runtime SELECT path.
An invalid supplied UUID remains rejected input, not an authoritative resource ID.

For trusted-only mutation primitives without an authenticated request, the Plan
must define the trusted caller/attribution contract and cover it independently;
adding a logger does not supply missing identity. Document the internal-process/
database trust limit. Do not claim that a compromised superuser or trusted core
writer cannot fabricate/suppress events. No new public append endpoint is allowed.

### R3 — Database ownership, append and runtime immutability

Add the audit schema through a reviewed additive migration after the actual
accepted predecessor version. Preserve existing migrations and application data.
Use separate audit object ownership and explicit least privileges. Ordinary
runtime roles own no audit schema/table/function and cannot UPDATE, DELETE,
TRUNCATE, alter/drop objects, disable controls, grant privileges or assume a
privileged writer/owner/maintenance identity. Preserve runtime non-superuser,
NOBYPASSRLS and accepted pooled-context protections.

The controlled append boundary validates allowed event shapes and trusted
attribution; raw runtime SQL must not bypass it to mutate existing rows or
assign another tenant/actor outside the documented authority. The Plan pins
exact table/function/role/grant/RLS inventory and writer trust boundary. A narrow
privileged function, if proposed, needs explicit ownership/search_path/EXECUTE
and input validation review; no broad definer or runtime owner connection.
Tenant audit tables require RLS including owner-path defense in depth as in the
accepted database convention. Privileged administration remains a stated exception.

#### Hybrid recording preference — owner amendment, 2026-10-14 UTC

Prefer database triggers to append audit records for committed row changes, with
Go supplying trusted actor/scope/request context. Go owns explicit recording for
authorization denials, authentication decisions, signed-URL issuance and external
storage outcomes that cannot be inferred from a committed row change. A trigger
can cover an actual session-row change; it does not by itself establish the
authentication decision or provider outcome behind that change.

The Plan maps each catalog entry to its recording mechanism and explains the
smallest choice. A combined mutation/audit SQL function is an acceptable reviewed
alternative to a trigger where it retains the guarantees and avoids a separate
client round trip. Do not assume triggers are faster: audit inserts, indexes and
durability still cost work. Compare statement/round-trip shape under PG-004;
no benchmark project or automatic implementation release follows.

Propagate trusted context through the existing transaction/pool safety boundary;
missing or stale attribution cannot silently become another actor's event.
Define explicit trusted system/bootstrap behavior where no request actor exists.
Trigger output is allow-listed event data, never a generic OLD/NEW row dump.
Assign responsibility for each event so a trigger and Go do not duplicate the
same catalog event; distinct row transitions and business outcomes remain distinct
where the catalog requires them. No automatic per-read/chunk event expansion.

### R4 — Atomic success and independent refusal/failure evidence

Except for the explicit security-revocation exception below, a successful
single-database mutation and its required audit append commit
in the same transaction: both durable effects or neither. Audit append failure
must not yield a successful unaudited mutation. Preserve existing transaction
cleanup, locks, current-fact authorization and bounded cancellation behavior.

#### Reusable transaction helper — owner amendment, 2026-10-14 UTC

Use a small Go core-internal audit-aware helper built on the existing transaction
and cleanup mechanisms for application-recorded atomic mutations. Claude's Plan
chooses the name/signature and whether this is an extension or narrow wrapper;
do not introduce an external utility dependency, generalized unit-of-work engine
or unrelated transaction refactor merely to unify different existing call sites.

The helper must operate within the owning transaction's lifecycle: establish or
validate trusted actor/scope/request attribution, execute the mutation and its
required application-owned append in that same transaction, and commit or roll
back both. Existing current-fact authorization, lock order, cancellation,
connection disposal and unknown-commit semantics remain authoritative. A callback
must not silently use another pool connection/transaction for the required append,
nest an independent commit or receive a cached authorization bypass. Do not
automatically retry callbacks, mutations or provider effects on ambiguous commit.

Where a trigger owns the catalog event, the helper supplies the required trusted
context and transaction handling without a duplicate Go append. Explicitly retain
separate bounded paths for post-rollback denial recording, signing/issuance and
external-provider outcomes; their ordering follows R4/R5, not a blanket mutation
callback. No provider I/O is placed inside a new audit-held transaction. A common
small append primitive may be reused without pretending these flows are atomic.

Cover the helper through the existing AC4/AC5 transaction/fault tests and AC6
issuance/storage cases, including append failure, rollback, unknown commit and
duplicate prevention. No second helper test framework or redundant suite is needed.

Denials and rolled-back operations cannot rely on an audit insert inside the
aborted transaction, including a trigger's insert. Use a bounded reviewed
independent recording path after safe cleanup where needed; it must not retain
poisoned context or reauthorize
the rejected business action. Pin timeout/cancellation and audit-unavailable
behavior per catalog entry. If recording also fails, do not claim a durable row;
emit a sanitized diagnostic and return the documented outcome. An unavailable
database cannot provide guaranteed durable failure evidence; state that limit.

An audit failure cannot permit a denied action, revive a revoked session or
undo an already committed security revocation.

**Owner decision, 2026-10-14:** when local revocation can commit but its audit
append fails, complete logout/session revocation or user disablement. Attempt
bounded independent recording and emit a sanitized diagnostic for any remaining
audit gap; never claim the missing record exists. This is a narrow exception to
atomic success recording for these security-reducing operations, not a general
best-effort audit policy. It covers local logout-current/all/provider-logout
revocation and disablement's session/user changes; sign-in, enabling a user,
role grants and other mutations retain strict atomic audit requirements. Failure
of the revocation database work itself still fails; unknown COMMIT remains
unknown. The Plan must make trigger/helper behavior preserve revocation rather
than allowing a failing audit trigger to abort it, with an explicit gap diagnostic
and no broad runtime trigger-disable or audit-bypass privilege.

 Preserve unknown-COMMIT truth:
do not report a definite rollback/success without evidence or blindly repeat a
mutation to obtain an audit record. The Plan must resolve per-operation ordering,
status/return behavior and retry identity before approval. Any conflict with an
accepted security/error contract returns to Codex; no blanket best-effort policy
for successful mutations or blanket audit-induced denial-of-service policy.

### R5 — External storage and signed capability boundaries

Preserve G009's short transactions, staged Put/Delete, unknown-commit reconciliation,
compensation and bearer/in-flight residuals. No PostgreSQL/MinIO atomic-commit
claim, network transfer inside a new audit-held authorization transaction, or
provider-effect replay on an ambiguous audit result.

Record metadata transitions in their owning transaction where applicable and
distinguish pending, ready, logical deletion, physical deletion/compensation and
unknown outcomes. Audit signing/issuance successfully before delivering a signed
URL to the caller; a failed required issuance append returns no URL. Never store
the URL, signature or provider credential. State what is known when issuance or
transfer succeeds but a later record fails. Ordinary streaming reads do not add
one audit write per chunk, and issuance is not evidence the consumer downloaded.

### R6 — Tenant read authorization and bounded queries

Provide an internal guarded core read operation using the existing `audit.read`
capability: **current Organization owner only**, target active Organization.
Organization admin/member, all Group roles, shared-profile access and other
tenants confer no audit authority. A selected Group is not a new audit grant;
it does not alter the canonical Organization-target capability. A global user
owning A but only a member in B cannot read B's audit through A context.

RLS plus the guarded read enforce current authority, including downgrade,
membership removal, disablement, ended sessions, wrong/missing context and pool
reuse. Global/unattributed identity rows are not exposed to tenant owners; define
their restricted inspection path without inventing a platform admin product role.
Reading an event never reveals hidden tenant/resource existence through errors,
pagination or counts. Preserve denied-target non-enumeration conventions.

Bound page size, filter shapes and ordering; use deterministic continuation with
tenant/filter binding and no unbounded scan/export API. The Plan pins limits,
ordering/cursor behavior, relevant indexes and behavior during concurrent appends.
No full audit console, search platform or cross-tenant runtime listing.

### R7 — Privacy, maintenance and migration boundary

Persist only the allow-listed audit fields. Exclude credentials, bearer/session/
login material, signed queries/URLs, object keys/content, business-context payloads,
emails/provider subjects, names/slugs and raw driver/provider errors. UUID
attribution belongs only where explicitly required; operational logs do not
silently gain durable-audit fields. Correlation remains diagnostic and bounded.
Test denied/fault paths and a planted forbidden-value detector, reusing existing
canaries rather than creating another scanning framework.

Describe restricted maintenance ownership and future G013 retention interface;
ordinary runtime append/read cannot delete old rows or change retention policy.
Do not implement retention scheduling/deletion, backup or privileged-administration
immunity. Deleting a source user/resource must not cascade-delete its audit
history; event attribution remains historical and does not confer current access.
Do not add resource foreign keys that prevent the accepted lifecycle merely to
make historical attribution look current. An existing populated audit table must not be silently discarded by a
down migration. Demonstrate the reviewed empty rollback/reapply and populated
rollback refusal or separately reviewed preservation route on disposable data.
Audit records survive the documented stop/start persistence cycle.

### R8 — Performance and proportionate validation

Assess new per-operation write/index amplification, transaction/pool/lock lifetime,
bounded refusal recording and tenant-read query/index selectivity in
[PG-004](../PERFORMANCE.md#pg-004--g011-durable-audit-write-and-read-cost).
No production latency/throughput SLO or prior-risk waiver is implied. The Plan
states expected statement/round-trip work and workload assumptions, plus only
the smallest query-plan/locking observation needed to resolve concrete risk.
Do not add unrequested benchmarks, partitioning, queues or tuning frameworks.
Flag materially new cost/feasibility concerns to the owner before proceeding.

Reuse accepted identity/RLS/storage fixtures, actual PostgreSQL runtime credentials,
private MinIO and existing test entry points. Use one representative success,
denial and rollback/fault per distinct persistence mechanism plus inventory
completeness checks; do not repeat every predecessor role/action cross-product.
Every new assertion proves an audit guarantee, not the unchanged harness.

## Acceptance criteria and verification

| ID | Required outcome | Verification/evidence |
|---|---|---|
| AC1 | R1 catalog covers actual existing foundation producers. | Finite catalog maps action, call site, required outcomes, actor/scope, transaction/failure policy and tests. Compare to current identity, guarded/model tenancy and storage entry points; future/exempt fixture paths explicit. Representative real calls for every category produce the expected events. |
| AC2 | R2 event format and attribution are enforced. | Valid user/system/global events; missing/invalid/oversized fields, forged actor/tenant, foreign resource and untrusted correlation negatives; compare exact durable fields and nullability, not operational logs alone. No invented actor/Organization for pre-context failures. |
| AC3 | R3 database inventory and controlled append are least privileged. | Versioned schema/grant/RLS/function/trigger inventory through existing catalog checks; actual runtime append positive and forbidden direct/forged append cases; UPDATE/DELETE/TRUNCATE/DDL/role escalation refused with rows/catalog unchanged. Separate setup owner from assertion role. |
| AC4 | R4 single-database mutation success is atomic with required audit. | Real successful mutation plus exactly the catalogued events through the chosen helper/trigger/function path; injected append failure and business rollback leave neither success effect. Verify application-owned mutation/append use the same transaction, no nested independent commit or duplicate same-event Go/trigger append; rejected/rolled-back event survives independently where storage is available. Exact row/outcome assertions and sanitized recording-failure result. |
| AC5 | Identity/global and unknown/failure paths remain truthful and secure. | Accepted local session/fake-provider lifecycle positives; invalid/absent identity remains unattributed, global rows hidden from tenant readers; audit-only fault does not prevent committed logout/disable, audit gap diagnosed and bounded recording attempted; revocation not undone; cancel/poisoned context and unknown commit handled per pinned policy without duplicate business effects or false durability. |
| AC6 | R5 storage records match real staged effects and issuance. | Actual MinIO Put/Delete and signed issuance with exact event states; one representative staged/compensation or ambiguous case per distinct audit mechanism; issuance audit failure returns no URL, provider/metadata effects reconciled and no URL/content leak. Reuse unchanged G009 residual evidence. |
| AC7 | R6 current-authority read isolation holds in core and SQL. | Owner-only positive; Organization admin/member and Group-role negatives, global multi-tenant user, cross-tenant filters/cursors, selected Group/shared profile, revoked session/role and pooled-context reuse. Exact visible row sets under attested runtime credentials; global rows excluded. |
| AC8 | Read limits and new query/lock costs are assessed. | Limit/filter/cursor boundary and concurrent-append ordering tests; bounded query shape/index/statement and client-round-trip assessment of chosen recording mechanisms linked to PG-004, attributable observations where justified, unexpected scans/fan-out/locks escalated. No arbitrary benchmark target or production claim. |
| AC9 | R7 privacy and maintenance boundary hold. | Audit-row and permitted-output canary scans on success/deny/fault with planted-leak detection; runtime maintenance/delete refusal; documented restricted future G013 path, no active retention/deletion scheduler. |
| AC10 | Additive migration and persistence preserve predecessor data. | Disposable empty/populated upgrade against actual accepted G010 version, repeat up, previous migration bytes unchanged; snapshot old application rows/grants, stop/start event readback; empty rollback/reapply and nonempty preservation/refusal. No production destructive test. |
| AC11 | Required validation is integrated and attributable without new harness scope. | Focused audit tests and relevant inherited regressions, minimal existing entry-point wiring, required-check execution counts/absence of skips, actual path-selected final local/hosted disposition. Record exact commands/source/target/inputs/tools and failures/interruption/reuse; no separate unchanged-control campaign. |
| AC12 | Review, budget, cleanup and handoffs are complete. | Result maps AC1–AC11, actual diff/independent security review, B7/checkpoint4, limitations, compact cleanup receipt and required staging disposition; Codex final ACCEPT, G012/G013 contracts and PROGRESS compaction assessment. G021 remains closed. |

## Expected artifacts, validation and provenance

Deliver a concise Claude Plan with the event/producer and transaction/failure
inventory, exact schema/roles/grants and append/read trust boundary, affected
modules/migration, bounded query assessment, AC mapping, delegation, named test
batches and timing/profile estimate. Resolve actual mechanism questions before
approval; do not build a parallel audit/validation implementation during planning.
Expected implementation is core audit code, minimum producer integration,
additive SQL, relevant tests and foundation audit documentation; extend existing
authorization/database/storage/follow-up documentation where behavior changes.
Internal events do not become public cross-service protocols merely to use
`contracts/`; public API/MCP contract changes require separate review.

Record the accepted G010 source/Result/corrections and versioned catalog, committed
Goal/Spec/Plan/scope, source/target/test inputs, actual role/tool/image versions,
commands, statuses and event-to-test mapping. Git-first provenance and existing
wrappers/resources apply; no new external snapshot framework. Separate real
SQL/provider evidence from fault injection and limits of each trust boundary.
Retain failure history and compact required evidence without secrets.

Sequence implementation, focused batches, independent security review/fixes,
then required final coverage/MR. No manual operator/live-provider ceremony is
expected; if a concrete need arises, obtain it before final qualification.
New migrations/authorization/identity/shared harness paths currently select full
coverage; Plan states the actual accepted map and cannot promise affected-only
hosted coverage. CI meta-tests run only for their actual changed dependencies
under existing rules. Do not run full per category or repeat passing evidence
for management-only/SHA changes with unchanged content and inputs.

Track owned resources at creation and perform one scoped cleanup pass with a
compact receipt after completion/interruption; hold shared/uncertain resources
without global prune. Apply current safekeeping, rebase/delta reconciliation,
MR/Duo/auto-merge and post-integration status policies without reproducing them.

## Permitted choices, unresolved questions and gates

Internal package layout, field/index layout, migration identifier and narrow
append mechanism belong to Claude's reviewed Plan. The Plan must resolve:

1. Exact current event inventory, trusted-only/system attribution and no-op/retry
   granularity; no deferred implemented category or unrestricted metadata bag.
2. Narrow helper reuse and per-operation append transaction/failure/unknown-commit
   behavior, including revocation, pre-context denials and external storage.
   Escalate any accepted contract conflict before implementation; recording
   success cannot be guessed.
3. Exact trigger/function/Go mechanism per event, trusted context propagation,
   append/read/maintenance privileges, global-event inspection boundary, schema
   constraints and current-authority read/pagination implementation.
4. PG-004 cost assessment and any owner decision required by concrete findings.
   No numeric SLO exists; prior G007–G009 dispositions do not accept new costs.

These are Plan approval gates, not authorization for probes/framework work or
silent architecture decisions. No requirement weakening or major deviation is
preapproved. Planning dispatch needs its own release; execution additionally
requires final G010 ACCEPT/integration, explicit Plan ACCEPT and fresh binding.
Completion requires AC1–AC12, actual-diff ACCEPT and applicable staging evidence.
G012/G013 remain dependency-bound and product implementation remains barred
through G021. No G012 detailed Spec or G011 implementation is created here.

## Planning release — 2026-10-14 UTC

The [Goal owner planning release](../goals/G011-durable-audit-events.md#owner-planning-release--2026-10-14-utc)
supersedes the planning-dispatch hold above. One xhigh planner may prepare the
Plan while G010 finishes; explicitly provisional assumptions must be reconciled
against final accepted G010. This releases no implementation, tests or builds.
Twelve ACs, B7/checkpoint4, Plan approval/binding, PG-004 and G021 are unchanged.

## Security-revocation policy amendment — 2026-10-14 UTC

Owner selected “Complete revocation; attempt bounded failure recording and report
any audit gap.” R4 and AC5 above incorporate this explicit exception; it supersedes
any proposed policy making audit-only faults block logout or disablement. Existing
AC count/B7, dependency, review and execution gates remain unchanged.
