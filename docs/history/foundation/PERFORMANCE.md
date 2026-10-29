> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](INDEX.md).

# Performance concerns and decisions

Repository register for concrete performance concerns identified during Specs,
Plans and reviews. See the architecture rule (not included).
Use stable entry IDs. Distinguish measured regressions from design hypotheses;
retain resolved entries with their disposition and evidence. Record affected
Goal/operation/query/index, workload assumptions, impact, evidence, evaluation,
responsibility and owner decision. Credentials and client data do not belong here.

## PG-004 — G011 durable audit write and read cost

- **Status:** owner accepts proceeding with the Plan r2 cost risk and bounded
  E-AUDIT-1..3 measurements (2026-10-14); no measured performance acceptance or
  production latency guarantee. G010 final acceptance and execution reconciliation
  remain separate gates.
- **Responsible Goal:** G011; Claude Plan assessment, Codex review and owner
  disposition if the proposed costs affect architecture, scope or feasibility.
- **Source:** [G011 Spec](specs/G011-durable-audit-events-spec.md),
  ADR-006 and the accepted identity/authorization/storage transaction boundaries.
- **Concern:** durable audit inserts/indexes amplify existing mutation writes;
  independent refusal recording adds transaction/pool work. Tenant-read RLS and
  pagination can scan or repeat helper work as event volume grows. New foreign
  keys/locks can extend authorization lock lifetime or introduce ordering conflicts.
  Provider I/O must remain outside new audit-held database transactions.
- **Assumptions/evidence:** source assessment only; audit event rate, retention,
  concurrency and production latency targets are unknown. Existing PG-001–003
  dispositions do not accept this new footprint. No numeric SLO is invented.
- **Smallest evaluation:** Plan states per-operation statements/round trips,
  index/query shape, bounded failure-recording budgets and lock order. During
  approved execution collect only attributable representative query/lock evidence
  needed for these paths, with fixture distribution/environment recorded. Reuse
  existing fixtures; no benchmark campaign, queue, partitioning or broad tuning.
- **Owner direction, 2026-10-14:** prefer trigger-based committed-row auditing
  plus Go recording for non-row/denied/external events. The Plan may justify a
  combined SQL function with the same guarantees. Avoid unnecessary extra client
  round trips, while retaining audit insert/index/durability cost and explicit
  transaction/failure semantics. This is a mechanism preference, not measured
  performance acceptance or permission for a benchmark comparison campaign.
- **Escalation:** flag unexplained tenant-wide scans, helper fan-out, blocking
  network-held transactions or feasibility-threatening write/pool cost to the
  owner before proceeding with a materially changed design. Preserve current
  authority, least privilege and revocation; no stale authorization cache.

### PG-004 provisional Plan r0 assessment — 2026-10-14

Plan r0 proposes four client round trips for independent refusal records and
signed-URL issuance, plus trigger/index work. Counts and Group validation require
correction before accepting the estimate. Existing logout-all/DisableUser update
all `revoked_at IS NULL` sessions, including expired but not revoked rows, with
no count cap; “bounded by live sessions” is not established. Correct the workload
assumption and document fan-out in the revised Plan. No session pruning/index/
benchmark campaign is authorized by that observation. The owner's security-only
revocation-priority decision is not performance acceptance. Corrected costs and
bounded evidence/owner disposition remain pending before execution.

### PG-004 provisional Plan r1 assessment — 2026-10-14

r1 review (not included) verifies the
corrected estimate in Plan `dd4406b5…`: independent append and signed issuance
add four client round trips; revocation adds a gap-marker read; Group attribution
adds an RLS-checked Group lookup (up to three role-reader calls in the independent
transaction). Session statement triggers add work to every session UPDATE,
including renewals. Logout-all/disable audit fan-out includes expired-but-unrevoked
sessions and is uncapped; event growth remains unbounded until G013.

This is a source estimate, not measured latency or production acceptance. One
remaining empty strict-INSERT correction must be reflected in renewal work.
Owner Q7 disposition remains pending on proceeding with bounded E-AUDIT-1..3
evidence and unknown workload limits. Existing performance waivers and the
revocation-priority ruling do not grant it. No new benchmark, index, pruning,
queue or tuning scope is authorized.

### PG-004 Plan r2 design review — 2026-10-14

Plan r2 ACCEPT (not included) approves
the implementation design, with execution held for the owner Q7 risk decision,
final G010 acceptance and Git-first binding. The empty strict INSERT is now
guarded by transition-table flags: a renewal executes no audit-table statement;
the transition capture and aggregate-query overhead remain. The other r1 cost
estimates remain. Recommend proceeding with bounded E-AUDIT-1..3 observations,
unknown workload limits and the existing escalation triggers; this recommendation
is not owner risk acceptance or measured performance evidence.

### PG-004 owner disposition — 2026-10-14

After Codex explained the corrected audit overhead and recommended bounded
measurements, the owner replied **"Proceed."** This accepts continuing with
Plan r2's disclosed design risk: audit write/index amplification, additional
round trips, renewal statement-trigger work and uncapped historical unrevoked-
session fan-out. Collect the existing bounded E-AUDIT-1..3 evidence; stop and
escalate unexpected scans, fan-out, helper/lock changes or feasibility-threatening
cost before changing the reviewed design. Workload limits remain unknown.

This is risk acceptance for approved execution and measurements, not measured
production acceptance, a numeric SLO, authorization for a benchmark campaign or
permission to weaken isolation/revocation. It closes Q7; no Plan revision is
required. G010 final ACCEPT and current-target execution reconciliation remain.

## PG-001 — G007 authorization lock/query footprint

- **Status:** flagged to owner, unmeasured design risk; no measured regression.
- **Responsible Goal:** G007 trusted context/policy; Claude Plan, Codex review.
- **Source:** first Plan r0, planning commit
  `49cf3e368ef9eed22423b7ad61c2f0b1bc5c3f36`, §4.1 and §3.7;
  Codex review (not included).
- **Concern:** Organization-only profile writes may load and lock every candidate
  Group membership/Group for the subject before reading associations. Query work
  and locks can grow with Group count. Shared Organization-row locks also
  serialize with policy changes; long transactions may increase revocation or
  policy-update waiting. The proposed predicate/index access and resulting
  contention need assessment against expected cardinalities.
- **Assumptions/evidence:** identified from the proposed lock/read sequence, not
  executed SQL or measurements. Typical/upper Group counts per subject, concurrency
  and latency targets are not established. Existing indexes are not evidence
  of an efficient plan for the new queries.
- **Smallest next step:** in the corrected G007 Plan, describe the expected query
  footprint and relevant existing indexes, identify workload assumptions and
  whether focused query-plan/lock-wait evidence is needed during approved
  execution. Escalate if the estimate threatens feasibility or requires a broader
  design change. Do not start a pre-approval benchmark campaign or add unrelated
  indexes. Preserve the authorization/write-ordering guarantee.
- **Owner disposition:** awareness requested by this register; no performance
  waiver, numeric target, architectural change or new implementation authorized.

### PG-001 review update — 2026-10-10

G007 Plan r1 review (not included) accepts the query/index/lock assessment and bounded E-PG-1..3 execution evidence. Additional hypotheses include shared-lock writer delay and a sign-in rotation deadlock with a safe definitive-abort test. No measured regression or latency bound exists. Workload assumptions and owner disposition of potential revocation/policy waiting remain pending before implementation; no index or lock-strategy mitigation is authorized.

### PG-001 owner disposition — 2026-10-10

Owner accepts proceeding with unknown workload limits and the recorded query/lock/revocation-wait risk for G007, collecting E-PG-1..3 without a maximum latency promise. Authorization ordering and escalation triggers remain mandatory. Revisit transaction deadlines and revocation latency before public tenant adapters. This is an explicit risk disposition, not measured performance acceptance or permission for new indexes/frameworks.

### PG-001 execution observation disposition — 2026-10-10

G007 blocker review (not included) disposes reported scratch E-PG-1 scans at statements 6b/7 on 260-/200-row tables as a bounded observation under the existing owner risk decision. Membership queries remained subject-scoped in that fixture; this is not attributable final qualification or production scalability acceptance. E-PG-2 lock-set assertions and required E-PG-1..3 evidence remain. Reported compatible readers passing queued policy/revocation writers preserve safety but give no fairness or maximum-wait guarantee. No index/query tuning or larger benchmark campaign is authorized. New broad membership scans, unexpected locking behavior or latency requirements still escalate; revisit workload/deadline assumptions before public tenant adapters.

### PG-001 additional execution observation — 2026-10-11

Follow-up review (not included) disposes G007 F7 at statement 6a (Gt): tenant-and-UUID point lookup/lock for group.rename scanned 264 fixture rows, returned one, hit four scan buffers and took 0.032/0.029 ms in the two modes. Membership work remained subject-scoped. Accepted as a small-fixture observation under the existing unknown-workload disposition, not production scalability acceptance; keep exact triggers and full evidence. No index/query/lock tuning authorized. Revisit real cardinalities, latency and contention before public adapters; unexplained broad membership work or locking changes still escalate.

### PG-001 F8/F9 observation disposition — 2026-10-11

G007 findings review (not included) accepts additional small-fixture sequential scans at statements 2/3/8/9 as observations under the existing unknown-workload owner disposition: 8 sessions, 6 Organizations, 104 profiles, 264 Groups; one result, shared-buffer hits only, sub-millisecond recorded times. No production scalability or latency guarantee, query/index tuning or trigger relaxation follows. A real slug change upgrades the Group lock to FOR UPDATE and can delay FK key-share inserts; the existing unchanged-slug footprint case did not measure that branch. Correct documentation, retain ordering guarantees and revisit real workload/deadlines before public adapters.

### PG-001 post-integration Duo disposition — 2026-10-11

[G007 post-integration review](results/G007-implementation-codex-review.md) (Duo !18 note 3993788301) records that denied exclusive-write requests can contend. A policy write takes the Organization row FOR NO KEY UPDATE before the capability decision, and a Group rename or slug change in Organization-only or selected-Group context can lock the target Group row before denial. A live subject lacking the capability can therefore delay authorized writes of that Organization or Group until its request is denied. No unauthorized mutation is possible, and ordinary guarded reads take no row locks and are not blocked. The locked-facts decision and the lock order are retained under the existing owner risk disposition; no authorization, SQL, schema, index or lock-order change. Future adapters must provide per-subject admission or rate limits and bounded request and transaction lifetimes before exposing these operations. Neither exists yet, and no numeric guarantee is given.

## PG-002 — G008 incremental RLS authority-lookup and lock cost

- **Status:** owner-visible planning risk; hypothesis, no measured regression.
- **Responsible Goal:** G008 PostgreSQL RLS; Claude Plan, Codex review.
- **Inspection source:** staging `b6306e6`, accepted G007 guarded facade and
  `tenancy/facts.go`/`identity/store.go`; [reconciled Spec](specs/G008-postgresql-rls-spec.md).
- **Concern:** row policies and recursive-reader helpers can add repeated
  membership/profile joins to existing subject-scoped reads and locking statements.
  Per-row predicate work, helper fan-out, tenant selectivity, changed index use and
  additional lock interactions may increase transaction lifetime or revocation
  waiting. Current Group-count/concurrency/latency bounds are unknown.
- **Evidence:** accepted source inspection only; RLS is not implemented and no
  timings or query plans have been collected for the proposed paths. PG-001's
  small-fixture observations and owner disposition are not RLS scaling evidence.
- **Smallest evaluation:** the G008 Plan identifies affected predicates/queries,
  existing indexes, workload assumptions and any bounded representative plan/
  lock evidence needed during approved execution. Reuse existing fixtures;
  no pre-approval benchmark campaign, new latency target or unrelated indexes.
- **Decision:** no mitigation or performance waiver selected. Escalate unexpected
  broad scans, lock cycles or feasibility-threatening costs to the owner and
  append attributable evidence here. Preserve authorization/linearization,
  runtime least privilege and all RLS guarantees. Future public-adapter admission,
  rate/deadline obligations from PG-001 remain; they are not implemented in G008.

### PG-002 G008 Plan r0 review — 2026-10-13

Plan review (not included) identifies
r0's proposed Session transaction as **+3 round trips per authenticated Session
resolution** and a longer renewal-lock interval. This is a source/Plan estimate,
not measured latency or an accepted workload limit. Scoped identity visibility
and authentication protection require a revised design, so its final query/
round-trip and pool-lifetime cost must be re-estimated. Owner risk disposition
is required before execution; PG-001's G007 disposition does not automatically
cover PG-002. No benchmark, batching optimization, new target or waiver selected.

### PG-002 G008 Plan r1 estimate — 2026-10-13

r1 review (not included) supersedes the
r0 per-request estimate: non-renewing Session lookup remains one round trip via
its bounded reader; r0's +3 on every lookup is withdrawn. Guarded transactions
add one context round trip plus policy/helper work. Renewal/login/logout/consume/
Enable add their planned transaction costs, generally three round trips; helper
calls and increased pool/lock lifetime remain unmeasured. No pool-pressure or
latency guarantee. The owner has been asked about proceeding with bounded
E-RLS evidence/escalations versus setting targets first; disposition is pending.

### PG-002 owner disposition — 2026-10-13

Owner says **“Proceed with measurements”**, accepting the recorded revised
G008 risk with unknown workload/latency/pool limits and bounded E-RLS-1..3
measurements plus escalation. See exact disposition (not included).
Normal non-renewing Session resolution remains one round trip; extra guarded/
renewal/login/logout transaction/helper work stays within the documented design.
No measured performance acceptance, numeric target, benchmark, optimization or
security waiver follows. Materially changed costs require review. Plan corrections,
Plan ACCEPT and execution binding remain prerequisites; no G008 work starts here.

### PG-002 G008 accepted Plan r3 cost clarification — 2026-10-13

Plan r3 ACCEPT (not included) folds
login-start cleanup into its guarded transaction before mutation. Login start
now adds four rather than three round trips and retains expired-row locks until
commit. This is the bounded cost of the required prior-context safety guard;
existing owner measured-risk disposition stands. Normal Session lookup remains
one round trip. Record actual costs/lock duration under E-RLS; no numeric latency
or pool guarantee, new optimization or performance waiver is introduced.

## PG-003 — G009 storage I/O and metadata authorization lifetime

- **Status:** owner-visible planning risk; unmeasured hypothesis, no regression.
- **Responsible Goal:** G009; Claude Plan, Codex review.
- **Source:** [G009 draft Spec R7/R8](specs/G009-tenant-object-storage-spec.md),
  accepted G007 ordered authorization locks and G008's planned context/readers.
- **Concern:** holding a database transaction/authorization lock throughout a
  slow upload/download/provider retry can increase pool occupancy and revocation
  waiting. New metadata policies can repeat membership/helper lookups or lose
  point-lookup selectivity. Provider time and database decision time are separate.
- **Evidence/assumptions:** source/design inspection only; no storage implementation
  or measurements. Payload/concurrency/latency bounds require Plan pins.
- **Smallest evaluation:** pin bounded I/O deadlines and transaction/lock lifetime,
  assess exact metadata queries/indexes, and collect only necessary attributable
  query/lock/transfer evidence in owned fixtures during approved execution.
  Prefer bounded staged authorization/finalization where needed without claiming
  database/provider atomicity. No benchmark campaign, new pooler/queue or numeric
  performance promise is authorized.
- **Disposition:** pending Plan assessment. Escalate unexplained fan-out, broad
  scans, lock-held network waits or feasibility-threatening costs; preserve
  current-fact authorization, failure safety and all RLS/credential invariants.
  Prior G007/G008 owner dispositions are not new storage workload acceptance.

### PG-002 corrected E-RLS observations — 2026-10-14

Escalation-03 review (not included)
accepts same-transaction function-counter deltas and records defective historical
alert output without erasing associated evidence. Fixture server 17.11; Group
writes use 17 reader calls over eight statements for both 1/50-binding subjects.
Profile group-reader growth is +94/+97 over 49 added bindings; this is observed
fixture scaling, not a total-call or production latency guarantee. Forty-four
reader alerts remain; per-policy attribution is inferred, not proven, and the
all-scanned-row denominator is only descriptive. Owner disposition on continuing
under this characterized risk is pending. No mitigation, threshold relaxation,
security change or broad benchmark authorized.

### PG-002 owner accepts characterized risk — 2026-10-14

Owner selected **“Option 1 is good”**: continue G008 and its required validation
without performance tuning. Disposition (not included).
PG-002 stays open as an accepted risk; production latency/scaling and the proposed
fixed-overhead explanation remain unproven. Preserve corrected warnings and
escalate materially different or feasibility-threatening cost; no threshold or
security waiver, extra benchmark campaign or automatic Goal acceptance.

### PG-003 Plan r0 assessment — 2026-10-14

G009 Plan review (not included)
requires exact-scope read discovery (no all-binding fallback), interruptible
upload-source handling and a genuinely finite transaction-cleanup budget before
approval. The staged I/O/short-transaction design and proposed point-lookup
evidence otherwise fit the existing bounded assessment. Approve collection
of E-OBJ-1..4 during later authorized execution, not a benchmark or numeric
latency/scaling acceptance. Proposed reader-call bounds remain hypotheses.
No new owner performance waiver is inferred from prior Goals; unexpected
fan-out, broad scans, lock-held I/O or feasibility-threatening cost still escalates.
No storage implementation or measurements exist yet.

### PG-003 measured INSERT RETURNING observation — 2026-10-14

Codex ruling (not included) reviews
G009 b12 at `780ca45` against staging `2485ede`: upload INSERT RETURNING incurs
4 Organization-reader and 2 Group-reader calls, identical at 1 and 50 memberships,
against Plan §10 hypotheses 2/1. PostgreSQL additionally checks returned rows
against SELECT policies. The bounded observation explains the extra calls;
retain RETURNING and key verification before upload. Preserve original hypotheses
and raw trigger evidence; no silent threshold increase or trigger suppression.
Other measured statements meet their hypotheses, and no runtime transaction is
idle during blocked transfer. This permits continued approved execution, not a
production latency/scaling promise or new owner waiver. PG-003 stays open;
unexplained growth, scans, locks and feasibility-threatening cost still escalate.

### PG-004 Group attribution deviation disposition — 2026-10-14

Codex's scoped review (not included)
accepts the Group-path attribution optimization in migration 00008 at candidate
`eb7826f`: use the same-statement invoker RLS Group lookup's current Organization
role requirement instead of a redundant separate role-reader call. The accepted
runtime/provisioned-owner security boundary is preserved, with privileged DBA/
BYPASS and arbitrary-runtime-context limits unchanged.

The representative observation was four role-reader calls before, three after.
Correct the estimate: two Organization-reader init-plans plus at most one Group-
reader call inside the Group lookup, with no additional outer call. Round-trip,
index and transaction boundaries are unchanged. Final-candidate focused evidence
and separate escalation-line inspection remain required; no production latency
or workload guarantee, new benchmark or additional owner risk waiver.

The implementer changed the design before the required cost-trigger escalation.
That process violation remains recorded, not retroactively authorized. Future
unexplained scans/fan-out/helper/lock changes return before design changes, as
PG-004 already requires. The scoped disposition releases final coverage only
after F1's missing runtime-denial cases and covering focused evidence pass.

## PG-005 — G013 retention and paired-recovery maintenance cost

- **Status:** owner accepts proceeding with bounded measurements and no
  speculative indexes (2026-10-15, explicit “Yes” to the PG-005 question).
  This accepts the characterized implementation risk, not measured performance,
  a production SLO or an index/scaling guarantee.
- **Responsible Goal:** G013; Claude plans/measures, Codex reviews, owner decides
  feasibility-affecting index/cost tradeoffs. Source: Plan r0 c7116c5, SHA4822a143…,
  review (not included).
- **Concern:** repeated bounded retention batches may scan/sort many audit or
  tombstone rows; per-row eligibility parses trusted settings. Global audit rows
  lack the partial tenant-time index path. Organization query predicates and
  eligibility-function optimizer behavior must be explicit before predicting
  index use. Delete/WAL/autovacuum and row locks add maintenance load; paired
  backup fingerprints scan/sort twice, extending a closed maintenance window;
  single-transaction restore and restriction sync add WAL/transaction/scan cost.
- **Assumptions:** synthetic two-Organization fixture only; real volumes, event
  rates, deadlines and concurrency are unknown. No timing/query evidence yet.
  Batch/timeout bounds limit work per invocation, not total table scan complexity.
- **Smallest evaluation:** corrected Plan pins query shapes and fixture row counts/
  distribution; E-RET-1..3 record representative EXPLAIN (ANALYZE, BUFFERS), actual
  index/scan paths and lock scope. Record one drill's sizes/durations/window.
  No benchmark campaign, speculative new index, partitioning or new service.
- **Recommendation / stop:** proceed without new indexes only after owner risk
  disposition; escalate unexpected scans/helper work/locks or feasibility-
  threatening cost. Routine bounded tuning stays reviewed and preserves RLS,
  audit, hold and replay guarantees. Production targets remain owner decisions.

### PG-005 G013 Plan r1 clarification — 2026-10-15

Plan57c58e8 adds explicit scope/time predicates, representative fixture counts
and the actual E-RET query shapes; index selection stays unknown until observation.
This resolves the source-assumption issue, not the cost risk. No new index or
benchmark is authorized. Owner answer on proceeding with bounded measurements
remains pending; capture-lifecycle corrections are separately in the r1 review.

## PG-006 — G014 MCP request authority and Group-list work

- **Status:** definition-stage risk flagged (2026-10-15), not a measured
  regression or approved implementation. Owner approved G014 definition and
  provisional planning; concrete query/workload assessment remains for the Plan.
- **Responsible Goal:** G014; Claude proposes query/workload shape during planning,
  Codex reviews and owner decides material architecture/feasibility tradeoffs.
- **Concern:** external MCP invokes PostgreSQL-authoritative identity/context
  checks on each protected call; protocol polling/reconnect and several clients
  may multiply existing G007/G008/G011 work. A new selectable-Group list can scan
  or generate per-Group query fan-out; connection-state/audit writes may add
  contention. No concrete new query, cache, index or throughput target is chosen.
- **Basis:** dormant MCP today; request-scoped access envelopes cannot be reused
  as a connection authority cache. Existing per-operation facts/locks remain
  required; prior performance decisions do not approve an unknown new footprint.
- **Smallest evaluation:** Plan inventories added statements/round trips/lock
  order, output/page bounds and Group-list predicates/index paths. Reuse the
  finite two-Organization fixture for attributable work/query observations if
  needed; flag unexplained scans, fan-out or contention before implementation.
  No benchmark campaign, speculative indexes/Redis authority cache, weakened
  revalidation or new performance SLO is required by this draft.

### PG-005 G013 Plan r3 design accepted — 2026-10-15

Plan6492a14/SHA4c936221 is accepted as design, B7/checkpoint4. Bounded retention
observations and no speculative index remain the recommendation. Owner answer
is still pending, so no risk acceptance or implementation release follows from
Plan approval. Current binding/resource gates are separately required.

### PG-005 owner disposition — 2026-10-15

Owner answered **“Yes”** to: may Claude proceed with bounded query/lock
measurements, without adding indexes upfront, with unexpected cost or locking
issues returning for review and no production performance approval. This closes
the pending owner gate. It authorizes the accepted G013 Plan's E-RET-1..3 and
single-drill observations, not a benchmark campaign, security relaxation or
unreviewed index/schema/platform change. Preserve prior pending assessments as
history; escalate material unexpected scan/fan-out/helper/lock cost as planned.

### PG-006 owner disposition — 2026-10-15

Owner accepted the recommendations presented with G014 Plan r3, including bounded
per-request query-plan/lock observations and escalation for unexpected scans,
fan-out, contention or cost. This accepts proceeding with the measured-risk
approach, not a production latency/throughput guarantee. Retain current live
checks, existing fixtures, rate limits/request deadlines; no benchmark campaign,
speculative index or authority cache. Plan r3 remains gated on actual WorkOS
qualification, final review and G013 acceptance. See the owner record (not included).


## PG-007 — G016 hostname lookup and domain lifecycle cost

Owner design/risk approval, 2026-10-16: **“Approve those details”** includes the
DB-backed hostname lookup with bounded query/index measurements. This authorizes
proceeding with characterized risk, not acceptance of measured production latency.
[G016 Plan r3 §5.5](plans/G016-hostname-resolution-domain-lifecycle-plan.md)
pins expected execution and the existing E-DOM-F1..F4 observations.

Standalone MCP/metadata/anonymous-host/relay checks add an indexed autocommit lookup
and pool acquisition; login, pre-exchange callback and final issuance use existing
caller transactions. Record session-origin/sign-out costs under the dated
baseline correction below. Count SQL, transactions and HTTP round trips separately.
Checkpoint observation, 2026-10-16: G016 draft `94a3a6a` instead uses separate
pool/autocommit lookups for login, pre-exchange callback and session-origin checks.
Those add acquisitions and statement transactions despite unchanged SELECT counts;
this departure is not accepted or measured evidence. The supervisor requests the
approved caller-transaction placement, or a bounded reviewed cost amendment if a
concrete constraint prevents it. Track actual E-DOM-F observations at delivery;
no new benchmark campaign or latency claim follows from this observation.

Reviewed baseline correction, 2026-10-16: the proposal at `d1601ca4…` and source
through `e00150d` confirm `Store.resolve` uses a pool/autocommit read in the common
path, with an explicit renewal transaction only occasionally. Accept one separate
indexed hostname lookup for session-origin and sign-out; relay was already
standalone. Each added lookup costs one acquisition, one statement/DB round trip
and one implicit autocommit transaction, with zero extra explicit BEGIN/COMMIT;
HTTP calls remain unchanged. Login/callback checks are restored to existing
transactions. No forced common-path transaction, combined reader or cache is
required. This is characterized-risk approval with existing bounded measurements,
not a measured latency result. See the proposal disposition (not included).

Bounded observations at delivery (G016 branch, local focused run of `TestDomains` DOM-10 at `701c8418`, PASS; the
`migrations` suite records them again as `domains-edom.jsonl`). No latency guarantee and no cache follow from them:

- **F1 plans:** on the small test tables the planner chose sequential scans for both probes (`organization_domains`
  filtered on `hostname` and `state <> 'removed'`; `organizations` on `slug`). The partial unique indexes
  `organization_domains_hostname_live` and `organizations_slug_key` exist and back uniqueness; index use at production
  row counts is expected but **not shown** by this evidence.
- **F2:** `ResolveWith` on a transaction runs 1 statement; `Resolve` on the pool runs 1 statement and 0 BEGIN/COMMIT.
- **F3:** statements per operation, constant between 10 and 40 domain rows of the Organization: claim 6, list 6,
  set slug 6, verify write 7, remove 7, retry 7 (the Plan's expectations were about 5–6; the difference is the
  context/liveness statements of the guarded transaction).
- **F4:** the issuance transaction holds only `AccessShareLock` on `organization_domains` and `role_bindings` (no row
  lock: snapshot reads); the verify write holds row-share locks on `organization_domains`, `organizations`,
  `role_bindings`, `sessions` and `users`.

Per-request cost as built: +1 autocommit statement per anonymous page, relay, MCP and metadata request and TLS `ask`;
+1 statement inside the existing login and pre-exchange callback transactions and +2 inside the issuance transaction;
+1 separate indexed lookup for session-origin and sign-out (the reviewed baseline correction above).

Domain lifecycle checks perform network work outside locks, then guarded writes;
record Organization/domain lock contention and slug-update conflicts with fact
reads. Check query/index shape, statement-count fan-out and unexpected waits with
existing fixtures. Actual results belong in the Result and this record at delivery.
Escalate material unexpected cost; no speculative indexes, authority cache or new
benchmark campaign. [Owner decision](../../adr/ADR-007-verified-domain-routing-and-authentication.md).


## PG-008 — G017 branding projection and logo lifecycle cost

- **Status:** provisional Plan r0 source estimate, corrections pending; no measured
  regression, production SLO or owner performance acceptance inferred.
- **Responsible Goal:** G017; Claude corrects query/pin inventory, Codex reviews;
  owner escalation if cost changes architecture/scope/feasibility.
- **Source:** r0 review (not included).
  Proposed per-page branding call adds hostname + selected public-projection reads;
  session membership adds a read inside the existing core call. Logo misses add
  projection reads and a bounded provider GET; cache hits change those costs.
- **Risks/unknowns:** actual repeated host/session reads, indexed orphan-ledger
  listing, new FK/unique-index creation on existing stored_objects, publication/
  deletion locks, compensation and uncertain commit work; production cardinality,
  concurrency and workload are unknown. Column grants do not establish query shape.
- **Smallest evaluation:** corrected Plan inventories actual statements/round
  trips, indices, lock order, storage bounds and cache semantics; extend its
  existing focused branding fixture counts where needed. No benchmark campaign,
  speculative cache service or duplicate validation suite. Preserve guard/RLS/
  revocation correctness. Escalate unexpected scans/fan-out/locks or feasibility-
  threatening cost before changing the reviewed design. Prior PG waivers do not
  silently accept this new workload.


PG-008 r1 reconciliation (2026-10-16): the revised proposal removes the new
stored_objects unique constraint/FKs and the uploads ledger. Those existing-table
index-build and permanent-retention-block hypotheses are withdrawn for r1, not
measured passes. Indexed finite image listing, helper-role projection policies,
per-page/session reads, bounded full-image processing and uncertain-commit handling
remain source-estimated work; r1 review (not included)
requests small safety corrections without a benchmark or new gate.


PG-008 r2 design reconciliation (2026-10-16): the
accepted provisional Plan (not included)
retains the r1 query footprint and adds bounded decode/re-encode normalization.
Its per-format decoded-image/canvas/output estimates are source estimates, not
measured peak process memory, a request-concurrency bound or a latency guarantee.
No benchmark, new semaphore or performance approval gate is added; unexpected
cost/feasibility concerns retain the normal escalation rule.


PG-008 G017 implementation source estimate (2026-10-17; Claude, for Codex review;
no benchmark, gate or owner acceptance inferred). Plan r2 §5.6 as implemented, with
the focused `TestBranding` BRD-12 statement counts (existing hook mechanism, one
local run):

| Path | Round trips / statements |
|---|---|
| Tenant page (any) | +1 web→core public projection: host lookup (1 unique-index probe) + `public_branding` (BRD-12: 0 transactions / 1 statement) |
| Authenticated tenant page | +1 `is_organization_member` probe inside the existing session call (no extra round trip) |
| MCP page indicator | +1 session read (lookup request ID) beside the page's own list call |
| `/account` on a tenant host | +1 management read for the branding link (BRD-12 get: 1 transaction / 8 statements, incl. the indexed upload listing, LIMIT 100) |
| `/organization/branding` | session read + management read (1 / 8) |
| Update / reset / remove logo | one guarded transaction (update: 1 / 8; a labelled update adds one statement) |
| Logo upload | 4 transactions / 28 statements (pre-check, the unchanged G009 Put footprint, association) plus one provider PUT ≤ 512 KiB; bounded decode/re-encode in core |
| Logo origin fetch (cache miss) | 0 transactions / 1 statement + one provider GET ≤ 512 KiB; fresh cache hits reach no origin |

Migration 00013 adds one table and no index on an existing table. Production
cardinality, concurrency and latency remain unknown; unexpected scans, fan-out or
lock amplification keep the normal escalation rule.


## PG-009 — G018 Group creation, onboarding completion and list cost

- **Status:** bounded design ([G018 Plan r5 §1.5](plans/G018-group-business-profile-onboarding-plan.md))
  with focused observations from the G018 branch; no benchmark, production SLO, latency claim or owner performance
  acceptance is inferred. Codex reviews at delivery.
- **Responsible Goal:** G018 (Claude records, Codex reviews; owner escalation only if cost changes
  architecture/scope/feasibility).
- **Operations and queries:** guarded `CreateGroup` (guard steps 1–5, one keyed `groups` INSERT whose AFTER INSERT
  trigger writes one `role_bindings` row, then one read of the new row), `CompleteGroupOnboardingWithNewProfile` (steps
  1–7 with the target Group `FOR NO KEY UPDATE`, then the profile and association INSERTs),
  `CompleteGroupOnboardingWithProfile` (steps 1–7 with the subject's other Groups locked `FOR SHARE` in ascending id
  around the target, at most one extra lock statement when they straddle it, then one association INSERT), `ListGroups`
  (facts, one keyset page statement, one associations statement) and `ListBusinessProfiles` (facts, one keyset page
  statement). Pages are 1..100 rows (`limit+1` read); no per-row round trip. The primary keys,
  `groups_organization_slug_key` and `group_business_profiles_business_profile_idx` serve the lookups and cursors; no
  index is added. The eligibility reader `app.rls_profile_eligible` is evaluated by the association INSERT policy (one
  profile probe, then at most one association/binding join); its cost is not separately measured.
- **Observations** (`TestAccessFootprint` "E-PG-2 onboarding writes (G018)", focused local runs recorded in the G018
  Result; dataset: Organization X with 200 Groups, 400 members and 4,000 Group bindings; subjects s1 with 1 Group
  binding, s50 with 50, and an X admin; Organization-only context). Asserted: each write issues its keyed INSERTs exactly
  once, no INSERT or unrecognized statement carries a locking clause, and s1 and s50 issue the same number of
  statements (existing-profile completion may differ by the straddle statement):

  | Operation | Statements | Locked rows |
  |---|---|---|
  | `group.create` | 7 (`C U S O OB IG Gt`) | user, session, Organization and Organization binding, one each, for every subject |
  | `group.onboarding.complete_new` | 10 (`C U S O OB GB Gt A IP IA`) | as above plus one Group binding and the target Group |
  | `group.onboarding.complete_existing` | 11 (s50, X admin) or 12 (s1, with the straddle lock statement) | with 50 Group bindings: 51 Group rows and 52 Group-binding rows (subject-scoped) |
  | `group.list` | 9 (`C U S O OB GB G ? A`) | none |
  | `business_profile.list` | 9 (`C U S O OB GB G A ?`) | none |

- **Unknown:** query plans of the two new page statements and of the reader at production cardinality (not measured),
  latency, lock-wait times, production concurrency and workload. Existing-profile completion locks grow with the
  subject's Group bindings (as G007's Organization-only profile writes, PG-001); creation and new-profile completion
  are constant in Group count.
- **Completion lifetime bounds (availability addendum):** a completion holds its target Group lock for at most its
  5-second decision budget (plus a 5-second bounded rollback). The completion route has a 10-second request and body
  read deadline. Each core process admits at most one completion per subject and 256 in all, without queueing
  ([AUTHORIZATION.md](../../foundation/AUTHORIZATION.md)).
  These are fixed safety bounds, not latency targets. `TestAccessOrdering` O-ONB-T checks that a bounded wait ends as
  unavailable with no orphan transaction. No throughput or latency figure is claimed.
- **Evaluation and escalation:** no index, cache, benchmark campaign or lock-strategy change follows from these
  observations. Unexpected scans, fan-out, lock amplification or a latency requirement return to review.

### PG-009 G019 membership, ownership and invitation operations — characterization pending

- **Design (Plan r1 §1.5, AC7):** the G019 guarded operations reuse G018's facts order and add no advisory lock; the
  only new indexes are on `app.invitations` (scope, pending-by-expiry, recipient). The database guard adds one scope
  row lock and one owner count per owner-role change or deletion; `invitations_bound` locks the scope row and counts
  pending invitations (at most 100 per scope); a claim locks the invitation row first. Member, invitation and candidate
  lists are keyset pages and hold no locks.
- **Evidence:** `TestAccessFootprint` E-PG-MEM records one observation per operation (lists, issue, get, candidates,
  promotion, removal with dependents, role change, demotion, revocation, Group removal, departure, retirement
  eligibility, the recipient's list and Group accept, the Organization consume and the transfer): statement steps,
  claim-function calls and locked rows by step. It asserts only that each consumer issues exactly one claim call,
  nothing else issues one, and no list or get statement locks. NOT EXECUTED until DB qualification; no count, throughput
  or latency figure is claimed.
- **Evaluation and escalation:** unchanged from G018: unexpected scans, fan-out or lock amplification return to review.

### PG-009 G062 platform root signup and tenant continuation — source estimate, not measured

- **Status:** source estimate from G062 Plan r5 D-10 and the implemented code (Claude records, Codex reviews). Hypothesis:
  negligible against the existing sign-in and Group-creation paths. No statement-count observation, benchmark, latency,
  throughput or production-cardinality claim is made, and no index, cache or lock-strategy change follows.
- **Root sign-in:** the signup login start is the ordinary login-start transaction; its expired-row sweep gains one
  `DELETE FROM app.signup_sessions` (expired rows by the database clock, EMPTY shape). The signup callback is the
  ordinary consume/exchange/mapping path plus one `signup_sessions` INSERT, and creates no session row.
- **Credential resolution:** every root signup request resolves its credential with `app.auth_resolve_signup`, one
  autocommit probe of the unique `credential_hash` index joined to the user (and the created Organization's slug).
- **Creation:** one IDENTITY transaction calling `app.signup_provision_organization`: the user row `FOR SHARE`, the own
  signup row `FOR UPDATE`, one keyed `organizations` INSERT (the unique slug index decides a race: one winner, the loser
  retains no write), one `org_owner` binding INSERT with its two trigger-issued audit rows, and one signup-row UPDATE.
  An identical retry (`already_created`) is the two locks and one UPDATE. The slug suggestion is advisory reads only.
- **Continuation hops:** challenge and completion each resolve the tenant host once (the existing `app.hostname_state`
  probe) and run one row-level UPDATE matched by a unique digest index; completion then runs one more short LOGIN
  transaction (`app.is_organization_member` and the pinned `login_transactions` INSERT). The root proof is one own-row
  UPDATE. No hop holds a lock across a redirect or a provider call. The pinned login adds two `boolean NOT NULL DEFAULT
  false` columns and one nullable `uuid` to `login_transactions` (constant defaults, no rewrite).
- **Revocation and recovery:** logout, sign-out everywhere, provider logout, supersession, `DisableUser` and
  `FinalizeRetirement` each add one `signup_sessions` UPDATE to their existing transaction; the restore invalidation adds
  one UPDATE over unrevoked signup rows.
- **Web:** the landing and each signup page or hop keep the existing host-gate lookup and add one core call; every hop
  is one 303. The rate limits are in-memory buckets per core process ([IDENTITY.md](../../foundation/IDENTITY.md)).
- **Unknown:** query plans at production cardinality, lock waits, concurrency and real provider latency on the final
  WorkOS hop (G057). Unexpected scans, fan-out or lock amplification return to review.

## PG-010 — G064 test-database cloning cost and connection contention

- **Status:** observed on one private test instance (G064 S4 campaign, 2026-10-22), not benchmarked; **open** until the G064 `migrations`-suite and identity-window observations (S6) and Codex review. Test infrastructure only; no production workload or latency claim.
- **Responsible Goal:** [G064](goals/G064-integration-test-isolation.md); Claude proposes bounded setup/cleanup and Codex reviews feasibility.
- **Concern:** cloning migrated templates multiplies fixture storage and setup I/O; template connections can block cloning, and independently owned pools may exhaust the private PostgreSQL test service. Repeated per-test bootstrap may cost more than the state coupling it removes.
- **Assumptions:** existing two-run DB/browser capacity and sequential image builds remain; package/test counts and template sizes must be inventoried. A clone does not copy cluster-global roles/settings or rebind existing ownership/ACLs to a new database-derived role family.
- **Smallest evaluation:** record template preparation, connection draining, representative package/test clone timings, pool limits and cleanup for the focused isolation proofs already required by the Spec. Keep lifecycle/global-state scenarios separate. No full-suite benchmarking campaign, extra parallel workers or production-role widening is authorized. Escalate a material feasibility/scope change before implementation.
- **Observations (G064 S4, 2026-10-22).** One run of the private-instance direct-selector campaign at `03e50356` on this WSL2/Docker Desktop workstation (PostgreSQL `pgvector/pgvector:0.8.7-pg17`, default `max_connections` 100, no published port; [Result](results/G064-integration-test-isolation-result.md), S4 rerun). These are single observations, not a benchmark or a CI estimate.
  - Template: `pg_database_size` 9,822,735 bytes (about 9.4 MiB); `db_template_create` 1,921 ms, including the source binding, the pristine check, the copy, the flag check and the shared-catalog snapshot. (S1's private proof at `f58b33f2`: 2,114 ms, 9.7 MB.)
  - `db_fresh`: 109 recreations (95 campaign selectors, 12 order/repeat/shuffle invocations, 2 controls), all PASS, no residue. Total per recreation min 1,398 / median 1,478 / max 1,628 ms; of that the precondition checks take a median 545 ms, `DROP DATABASE` 144 ms, `CREATE DATABASE ... TEMPLATE` plus grants 145 ms and the postcondition checks 640 ms. Zero-session drain waits: none. Refusals: 6, all from the deliberate controls (T-1b, T-1c, T-3, T-5a-c).
  - Phases: campaign 459 s for 95 selectors, order/repeat/shuffle 118 s for 12 invocations, all phases 648 s (instance and image phases included; wall clock 655 s). Recreations sum to 141 s against 296 s of test invocations in the campaign. Per selector, `db_fresh` was longer than the single-test invocation it preceded for 5 of 95 selectors (by at most 73 ms; all invocations of 1.4-1.5 s with a warm build cache); with the bounded drain poll included, 23 of 95. No order/repeat/shuffle invocation (4.6-13.5 s) was shorter than its recreation. Whether the short-selector cases meet the Plan's stop condition ("template or recreation taking more than the invocation it precedes") is left to Codex review; the suite's own 50 reset points precede multi-test invocations and are not yet observed.
  - Connection contention: the first campaign run (`s4-8d712163`) failed `TestBusinessProfileModel` with `53300` (too many connections) because a fresh copy starts its version ladder at the top version, whose superuser purge connections stayed open until the test ended; the helpers now close them (S4 correction 1) and the rerun passed with no limit or pool change.
- **Observations: corrected S4 campaign at `17fdad66` (G064 S6, 2026-10-22).** The same campaign rerun once on the corrected head (same workstation and instance shape; [Result](results/G064-integration-test-isolation-result.md#s6-timing-disposition), S6 timing disposition): 95/95 selectors and 12/12 order/repeat/shuffle invocations PASS, cleanup clean. Single observations, not a benchmark or a CI estimate. **These per-selector figures supersede the `03e50356` comparison above (5 of 95, by at most 73 ms; 23 of 95 with the drain), which is kept as history.**
  - Template: `pg_database_size` 9,740,815 bytes; `db_template_create` 2,126 ms.
  - `db_fresh`: 109 recreations, all PASS, no residue; total min 1,441 / median 1,694 / max 2,997 ms (medians: checks 599, drop 159, create 166, postconditions 733 ms); 6 refusals, all from the deliberate controls.
  - Campaign: raw `db_fresh` min/median/max 1,441 / 1,663 / 2,997 ms, summing to 167,452 ms against 357,678 ms of single-test invocation time; with the bounded drain (`it_fresh`) 192,296 ms. Raw recreation was longer than the invocation it preceded for **12 of 95** selectors (by at most **948 ms**), with the drain for **20 of 95** (at most **1,141 ms**). No order/repeat/shuffle invocation was shorter than its recreation. Phases: campaign 551 s, order/repeat/shuffle 147 s, all phases 793 s (wall clock 805 s).
  - **Codex disposition (2026-10-22):** the Plan §10 stop is cleared by bounded continuation with this recorded **performance limitation**: per-test isolation costs about as much as a short single-test invocation. It is measured isolation cost, not a correctness failure and not permission to weaken resets; no optimization or benchmark campaign is required to continue. Materially different later observations return through the risk register. Status unchanged: **open**, observed, not benchmarked (the batch-4 identity-window observation and Codex review remain).
- **Batch Docker reuse (G064 addendum A, implemented 2026-10-22; not yet measured).** `migrations` now runs its eligible Go invocations as `docker exec` calls in one long-lived runner per batch, instead of one `docker run --rm` each ([OPERATIONS.md](../../foundation/OPERATIONS.md)). Each invocation still gets its own `db_fresh`, so the recreation figures above are unchanged by design.
  - **Basis, not a new measurement.** The Plan (addendum A §A1) recorded about 6-7 s of fixed cost per invocation on the container path. That figure comes from single-test catalog checks of run `79085025`. The S4 exec mode recorded about 2.0 s on the same basis (Result, S6 timing disposition: 357,678 ms of invocation time against 166,640 ms of test elapsed over 95 invocations). The Plan's resulting wall-clock and creation-count savings are **estimates** and are not repeated here as results.
  - **Unmeasured.** Neither the runner's actual saving nor the added cost of the accounting has been measured: the observation points, the witness markers (one `docker top` and one acknowledgment wait per witnessed call), the serialized ceiling queries and the per-helper cidfiles. The focused real-Docker proof (`management/v0/results/G064-br-proof-driver.sh`) and the consolidated qualification will report both. The accounting figures will come from their `resources.tsv`: creations and final leftovers measured exactly, maxima sampled (lower bounds), ceilings source-derived.
  - **Status unchanged.** The cost of `ci-controls` (the critical path when selected) is out of this slice's scope.
