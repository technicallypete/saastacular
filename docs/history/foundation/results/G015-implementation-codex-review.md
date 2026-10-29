> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G015 implementation review — 2026-10-16 UTC

**Outcome: ACCEPT — pre-integration source and attributable local evidence.**
Candidate `c7ce0bb3eed25c32199ef531fb4efaa364212464`, base73a4274, branch
`g015-dynamic-capability-exposure`. This is not final Goal ACCEPT; required hosted
MR/Duo disposition, integration, staging qualification and closeout remain.
Supervisor and two bounded read-only reviewers inspected actual source, current
contracts, access/policy invariants and retained focused evidence. No review tests,
builds or provider probes performed. All18 changed paths match approved scope.
Exact [Result](G015-dynamic-capability-exposure-result.md) imported, SHA-256
`70e3ac47e3573e521e68bfe483c6d6de29cea29ed20e15ae7a720b70cb4d69f4`.

## Source and deviations accepted

Requirement-aware per-request registration keeps the five production controls and
their own authorization checks, including explicit clear/logout after access loss.
Hidden and absent names share one discovery snapshot, SDK envelope and truthful
records across context denial/outage. Names-only AuthorizeCatalog returns no effect
authority. Capability fixtures remain test-only/effect-free; no authority cache,
new production tool, schema/index, migration, identity namespace or audit vocabulary.
Discovery follows the accepted role/link/policy matrix and independent Organization
owner/admin paths. Current evaluator/guarded effect checks and tenant-scoped facts
remain. Snapshot ordering permits an in-flight read on its snapshot, then rechecks
new requests. Notifications remain truthfully listChanged:false.

Accept the reported bounded deviations: production catalog validation at package
initialization fails closed; conditional fact-statement skipping is below the
accepted maximum; a liveness loss after the transport check yields the same
non-authorizing unknown-tool response for hidden and absent names; an infrastructure
failure retains truthful failed decision logs without inventing an audit success.
No extra owner decision or technical correction is required for these deviations.

**Result wording clarification:** catalogSpec validates the input schema's object
shape, not output schema validity/closedness. Current production/fixture outputs
are fixed, closed and pinned by tests. The Result's claim that missing/non-object
"schemas" are refused is overly broad; interpret it as input schema validation.
Coordinator qualifies that sentence in its next Result update; no runtime defect
or additional test campaign is inferred. Imported exact submission remains history.

## Attributable evidence

Retained b1 unit/static selectors pass. b2 meta/log evidence reports transport7/7,
lifecycle6/6, audit1/1, recovery1/1, discovery5/5, ordering10/10, footprint1/1,
policy38/38 and policy ordering17/17; five E-MCP records, zero triggers/canary hits,
owned teardown successful, leftover empty. E-MCP-4 records production/context
statement equality6/6 Organization-only and8/8 selected; capability shape10.
Additional Group/profile sets do not increase statement counts; this is bounded
fixture evidence, not a production latency guarantee.

After S-review fixes, b3 reran changed TestMCPDiscovery5/5 (36s), changed-file gofmt
and integration vet, with zero canary hits and no leftovers. The final core tree
matches b3 across the management-only rebase. Earlier b1/b2 passes remain valid
for unchanged code; no whole local suite was restarted. Static-r3 shows all17
required IDs once as PASS, no FAIL/UNMET, source/canary and18-path scope passes.
Full migration-driver/logcheck and identity/MCP E2E execution remain hosted-only
under Q9. Harness regex/count/text were inspected against actual selected outcomes.
This review does not claim a passing full local harness or executed hosted tests.
One/B7 attempt used, checkpoint4; one internal S-review correction cycle, no reset.

## MR readiness and next gates

Verified actual draft implementation MR !33, sourceg015-dynamic-capability-exposure,
targetstaging, candidatec7ce0bb:
https://gitlab.example/<namespace>/<project>/-/merge_requests/33
The source is based on73a4274; latest f649aca adds management-only early-MR/review
records. Coordinator reconciles current remote target before ready/integration,
preserving accepted executable content and attributable evidence. No local full
rerun for the management-only delta. Qualify the Result wording, record actual
reconciliation and return substantive changes for affected review if any.

Mark the reviewed MR ready, request/verify current-head Duo and passing fast checks,
reply to actual findings and relay them for disposition. Once all pre-merge holds
close, arm/complete auto-merge with the exact reconciled candidate and source-branch
deletion under standing authority; no new owner merge prompt. Do not request the
optional MR full gate. Required staging gate and cleanup precede final G015 ACCEPT.

## Verified G016 advance-planning trigger

MR !33 qualifies as the actual predecessor implementation MR, including draft
status. Existing ready G016 Goal/Spec and the recorded EPIC-005/006 owner mandate
allow provisional planning now. No existing G016 Plan was found. Coordinator
reuses any assigned planner, otherwise dispatches one bounded planner under current
model-aware policy; no duplicate coordinator/worker or naming exception.

Record MR33/candidatec7ce0bb and current target as provisional predecessor baseline.
Plan inspection/history only: no application/test implementation, install, build,
shared test resources, live DNS/TLS/WorkOS probe or provider settings change.
Compare bounded DNS/TLS/edge, browser callback/handoff and management-authority
choices; escalate owner-owned vendor/security choices before Plan ACCEPT.
Include the owner-requested bounded sqlc assessment for one new domain-query
package; adoption or concrete deferral, no mandatory tool or performance claim.
Reconcile eventual accepted G015 and predecessor changes before execution.
G016 implementation waits final dependency ACCEPTs, accepted Plan/bindings,
owner decisions, resource/budget readiness and recorded release. G021 and the
G017-inclusive mandate endpoint remain; no G018 implementation is authorized.


## Duo disposition and bounded correction release — 2026-10-16 UTC

**Outcome: CORRECTION LOOP — one bounded correction round, attempt2/B7.**
Actual MR !33 head154ba827a4c81d0b5083763f174e0600dbe473c1 has passing fast
pipeline2954646033. Executable source is unchanged from acceptedc7ce0bb; delta is
management only. Read both actual unresolved Duo threads; retain merge hold until
the scoped correction returns for review. Duo disclosed truncated run.sh coverage;
Codex's earlier source/evidence review remains separate.

### D1: target-existence wording and literal coverage

Thread222379f262618344338ced0f7fe4bac71c1e58e8 identifies a real wording/test gap,
not a live capability-tool authorization bypass. Pin the existing Organization
owner/admin path as a role-level discovery hint without an existence assertion,
including an empty Organization. No profile-only existence gate, extra fact query
or change to decide/effect authorization. Current context/selection validation,
owner-only privileged settings and Group-derived candidate/link/policy checks stay.
The supervisor Spec clarification above is accepted within R1's target-semantics
pin, without changing roles, effects or adding a production tool. Align code/doc/
contract descriptions as necessary and add a small literal empty-world test using
existing fixtures. Do not clone the matrix or create a new oracle.

### D2: remove global fixture lock from production requests

Thread5214a10f79e526900f4ab81f596b0896b132ba5e identifies an avoidable per-request
global lock/cross-handler fixture coupling. Use a private per-Handler catalog
override, nil in production, with the fixed production fallback. No exported
configuration, runtime flag, environment setting or production fixture setter.
Catalogs are immutable after publication. Existing tests toggle fixtures on a
serving handler: use a race-safe per-instance pointer (for example atomic.Pointer)
or restructure bounded fixtures to publish only before serving. An ordinary pointer
with concurrent unsynchronized writes is not approved. Remove package-global
fixture state and its mutex; isolate reset/cleanup to the owning handler. Do not
copy Handler values containing synchronization state.

The schema-comment nit remains documentation-only: catalogSpec currently checks
input shape, not output validity/closedness. Qualify that claim; fixed production
schemas and existing pins remain. Do not expand validation machinery here.

### Scope, checks and handback

Approve only core/internal/mcp/handler.go as an additional source path for that
private field; existing catalog/fixture/tests/policy-comments/docs/contract paths
already cover the correction. Scope hash before `eeb84eb34f0d983be41e788f4c3f307ca99b18435aaa389dc3cf09ef40bc1986`;
new hash `f01feafc05b8e24c23156eef7557afc43ba94172849819919d4212d763ea39d4` (ten patterns, only handler addition).
Bind updated Spec and scope/review at the next safe writer boundary; preserve
original Plan r4 bytes and prior binding history. No full replan or extra owner
approval loop. Coordinator keeps one writer within350k context policy, records
actual model/effort and attempt2/B7; no budget reset.

Run affected eligibility/catalog unit cases, a small two-handler isolation/reset
regression with the race detector where supported, and affected TestMCPDiscovery
real-HTTP cases. Use existing runtime fixture/teardown and narrowed selectors,
changed-package vet/gofmt and necessary source/secret/scope checks. No local
unit/migrations/ops umbrella, full controls, benchmark or duplicate suite. Preserve
all unchanged prior passes, record attributable deltas and return exact code/Result
for Codex review. Normal MR fast/Duo applies after push; auto-merge stays held.

G016 provisional planning continues; relay relevant discovery/seam clarification
at its next safe planning boundary, not an implementation restart. Final G015
ACCEPT still requires qualifying staging and closeout. G021 remains closed.


## Attempt2 Duo correction acceptance — 2026-10-16 UTC

**Outcome: ACCEPT — pre-integration correction source and local evidence.**
Candidate `2ef5fcde271339df6179266bee2b9d2e61187eb9`, baseb219ef4. Exact updated
[Result](G015-dynamic-capability-exposure-result.md) imported, SHA-256
`e4a6310a5a83a553f5e402c52d88d52bffacc7c66d0b213fa65ccc90f689f9a3`.
Supervisor and bounded read-only review found no blocking correction defect.
This accepts attempt2/B7; checkpoint4 and final Goal gates remain unchanged.

### Findings closed

D1 clarifies Organization-level discovery hints without claiming target existence;
13 literal empty-Organization cases cover owner/admin, owner-only privilege,
non-admin and context refusal. Effect evaluation, fact options and queries remain
unchanged. D2 removes package-global fixture state/mutex: Handler has a private
atomic.Pointer[catalog], nil in production; lookup atomically loads an immutable
catalog, and test store/reset acts on that handler only. All affected clear calls
use the correct handler, including the store-down fixture. No production setter,
configuration toggle or Handler copying. Input-schema-only wording is corrected.

**Explicit boundary-pin disposition: ACCEPT sync/atomic in TestPackageBoundary.**
The private Handler field was already approved, handler.go and mcp_test.go are
within scope, and this is a required standard-library synchronization primitive,
not a dependency or broader access-layer import allowance. Only that import was
added to the MCP production-import pin; actual architecture boundary stays intact.
All19 committed cumulative paths match scope; static qualification counted18
before the final Result-only update, not an unexplained extra source path.

### Evidence and race-detector limitation

Independently read b4-r1/meta: changed-package vets/gofmt and empty-Organization
eligibility PASS, discovery5/5, zero canary hits, teardown rc0/leftover empty.
That batch retains the original boundary-pin failure; do not call it wholly green.
The following formatting failure is likewise retained. Final b4b-r2/meta verifies
fixed gofmt/vet and all3 selected MCP unit tests PASS, no SKIP/FAIL. Static-r4
has no FAIL/UNMET,17 required IDs once as PASS, source/canary/boundary/scope passes,
empty leftovers. Final code/docs/infra match904de79; affected-only reruns cover the
actual delta, retaining unchanged earlier attributable coverage.

Race instrumentation did **not** execute: initial command reports requires cgo;
retained support log shows CGO_ENABLED=0 and no gcc. It is unavailable, not PASS.
Accept this scoped limitation under the approved "where supported" condition:
source uses standard atomic pointer publication, immutable contents and handler-
local reset, while TestCatalogFixtureIsolation checks concurrent reads and
cross-handler isolation. The concurrent test passing is not a claim of race-
detector proof. No new image/tool install or broad campaign is required solely
for this correction. Hosted-only evidence and required staging gates remain.

### Reconciliation and integration handoff

b219ef4→b8ceeab changes only PROGRESS and the G016 review; reconcile latest remote
target, including this management-only import/review. Preserve reviewed executable
bytes, scope and evidence; no local rerun for unchanged source. A substantive new
delta returns for affected review. Record actual rebased SHA and source equivalence.

Reply to/resolve actual Duo D1/D2 threads with the accepted correction and honest
race limitation, verify current-head Duo disposition and required passing MR fast,
then arm/complete auto-merge for MR !33 with exact candidate/source deletion when
all native holds clear. No repeated owner merge prompt or optional MR full gate.
Required staging disposition/closeout precede final G015 ACCEPT. G016 provisional
planning continues its existing correction loop, reconciling relevant final G015
content at R0; no restart or G021 waiver.


## Final implementation acceptance — 2026-10-16 UTC

**Outcome: ACCEPT — G015 integrated and closed; EPIC-005 complete.**
Prior source/correction reviews plus actual qualifying staged source and closeout
satisfy AC1–AC10. MR !33 merge `7e0ac44c6f85f9de73a8e3c53d209296f9f90a6f`, source
`b80c28307c097c72041012a60385e82720734c08`, retains reviewed executable bytes from
2ef5fcd; remote source deletion was independently verified. GitLab pipeline
2954782898 SUCCESS on its first run: fast17229626722 SUCCESS, gate17229626723
SUCCESS (3067s). Read the actual gate artifacts:80 PASS,0 FAIL/UNMET,6 INFO,
79 required IDs once as PASS, all441 CI controls pass. Previously hosted-only
B-MCP-IT driver/logcheck, identity browser and MCP E2E all pass. results.tsv SHA
`40f5e4dc1433a3e9302b6eadc7af9f533611b11fea764a40a5567276ebb7fd47`.
Cleanup inventory removes four owned images; only project-archetype-render-cache
remains. No new test run performed for this acceptance.

Exact closeout source `6609658c4070abc00f4ba074eeb8b880b6801bc5` changes only the
Result and preserves earlier history. Imported exact
[Result](G015-dynamic-capability-exposure-result.md), SHA-256
`69235db5f96db2287ad85868d7e89935ad79d82df761b3b33021a365b3eae986`, and
receipt (not included), SHA-256
`edb4d484cabd159dd1dff4436b0db6885aae2c65625f9843a8dd7c1b1486d021`.
Reported local removals remain receipt evidence; current worktree list confirms
finishing/R0 checkouts absent. Two/B7 used, checkpoint4 not reached. Local race
instrumentation remains unavailable, not passed; accepted limitation preserved.

### Precise cleanup disposition

Independently verified old implementer checkout agent-<id>: all1286
tracked files match709f551, no untracked/ignored files, no process cwd references.
Its current branch b80c283 is merged. Clarify receipt wording: application patches
are preserved; two older Result-only commits (709f551,a8f04c1) are patch-distinct,
with their binding/implementation history carried in imported Results. Do not
claim entire-tree/commit patch equivalence with the later corrected delivery.

Authorize the existing coordinator to use **scoped git worktree remove --force**
for that exact old implementer checkout only after confirming native worker
completion/release and unchanged no-unique-work/private-state preflight, then
remove its merged local branch. This addresses ordinary Git stale-index refusal;
it does not authorize bypassing a native permission refusal, unlocking a live
agent, removing active G016 or sweeping other worktrees. If native release is
unavailable, record the precise nonblocking hold once and finish other cleanup.
Retired native-locked planner, historical remotes and prior unproven holds remain.
No additional owner prompt or validation campaign for routine scoped cleanup.

### Epic checkpoint and next gates

EPIC-005 contains accepted G014/G015 under the pinned mandate. Codex completed
the required PROGRESS compaction checkpoint: prior index text/dates/links archived
at EPIC-005 history (not included), current acceptance,
owner authority, unresolved decisions/holds and next gates retained in the concise
index. No Goal/Spec/Plan/Result review history was rewritten by compaction.
G016 planning is on owner hold with no active immutable implementation batch;
coordinator reconciles the management-only index at the next safe read/binding,
with no test rerun. Future EPIC-006 closeout retains its own checkpoint.

G015 dependency is satisfied for G016, but owner D1–D4 design package remains
unanswered. No full G016 Plan ACCEPT, live action or execution release is inferred.
Existing planner may perform bounded accepted-predecessor R0 inspection while
waiting; execution still needs owner choices, Plan/bindings/budget/resources and
recorded release. No new planner, test run or redundant Plan revision solely for
a SHA/index change. G017 trigger and G018 stop/G021 barrier remain unchanged.
