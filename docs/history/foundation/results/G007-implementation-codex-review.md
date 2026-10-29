> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G007 implementation readiness review — 2026-10-11

Outcome: **ACCEPT** — actual-code and local-evidence readiness, including attempt-2 corrections C1/C2. This is not final Goal ACCEPT or owner merge authorization.

## Reviewed inputs

- Candidate `c151407e1e196d3bf38a17955016326c930c4ec1`, clean in `g007-agent-5c6f436`; tested implementation `12d2d8ec1b1a81d9bbbff02042242f274b596b8e`, base `da1ed80`. Changes since the tested implementation are governance/management records, not application or harness content.
- Current Goal/Spec r3, approved Plan r1 SHA-256 `b6bf98f461cc20fc3f4b55d2d808c93702607505ae5c48d1183f68440a179e60`, prior correction review and dispositions, actual production diff and representative independent-oracle/context/boundary/transaction/ordering tests.
- Result imported byte-exactly from the candidate: SHA-256 `99e60f486b1dc1e1356d0062c91a980f8da7a90b43ffca8025cb44a25900ab0f`. Its early in-progress header and missing final checklist remain to be completed at S9; imported bytes preserve that history.

## Findings and evidence

No further blocking authorization-code finding. The opaque request-bound context derives identity from the accepted session boundary; each operation revalidates session, Organization membership and selected Group. RoleBinding remains the sole role authority; profile authority derives from current same-Organization Group links or Organization roles. The closed matrix, selected-Group limits, owner-only policy update and optional-MFA semantics match the current Spec. No public adapter, product workflow, RLS or independent profile-role store is added.

Guarded writes evaluate locked facts and mutate in one transaction. Ordered user/session/membership/Group locks compose with the inspected predecessor writers and the explicit unlink Group-lock amendment. Reads use a fresh repeatable-read snapshot. The accepted slug-key lock upgrade, possible lifecycle deadlock with safe unavailable response, and small-fixture PostgreSQL scan observations remain disclosed limitations; no production-scale performance guarantee is inferred. Future adapters still need workload/deadline/rate-limit assessment under PG-001.

C1 retains exact independent cardinality expectations: full fast/gate/union 33/60/89, integration unchanged 33/42/71. C2 conservatively distinguishes ordinary command ERROR/ROLLBACK from FATAL/PANIC/unknown/lost COMMIT replies, preserving public unavailable, bounded private reasons and no automatic retry. The 21 classifier and transaction cases assert exact reasons, one attempt and driver-text exclusion. Real definitive-abort and durable-lost-reply cases reran.

Independently verified the attempt-2 report bytes:

- `a2-c1c2-1/report/results.tsv`: `90e4036f28fe0ad44a3d1c750d8e12f9194ce5fa124cf81d64d07b8ded96df77`, 34 PASS, no FAIL/UNMET; harness exit 0 and wrapper postcheck passed.
- `ci-controls.log`: `ac97d60a64d9d4ed5c5cf4e0cd413201b6c5813d1632f45feaaee11aa159899c`; 439 controls pass, including covering full-count and union controls.
- `policy-epg.jsonl`: `e6fc938d74ba31c1bd086b0224324adf49ef6f76bfb4f9e6b44a288565a17c67`; catalog/migration/policy/ordering coverage is reported in the Result. Leftover inventory is zero bytes.

Accept the documented content/input-based reuse of unaffected passing `s8-final-3` checks. Its historical run remains failed. Combining those passing checks with the covering attempt-2 replacements yields 89 distinct PASS obligations, no remaining failing obligation; this is composite evidence, not a claim that the failed full run passed. No further unchanged full local rerun is requested.

## S9 release and remaining gates

Claude may finish S9 and open the G007 MR to staging under standing authority after fetching/rebasing/reconciling the actual target. This review does not authorize a merge. Preserve G051's scheduling priority when granted; do not alter the shared lock.

Complete these bounded handback items within the already authorized attempt 2, without another implementation correction cycle:

1. Correct only OPERATIONS.md's full-profile Required IDs cell from 84 to 88 (the full coverage row already says 33/60/89). This exact documentation-site extension is authorized; no control logic or new scope. Verify the two full-count rows agree and append the amendment/evidence.
2. Complete the Result's current submission status and AC1–AC14 evidence checklist, preserving prior failures and history. State AC13/AC14 hosted/integration/final acceptance as pending, with the lock-upgrade and measured-footprint limits recorded under AC8. Record attempt 2 accounting and the evidence reuse above.
3. Reconcile any actual new staging content before MR opening; retain G007's four count reconciliations when merging G051 controls changes. Retest only affected content/validation inputs under the accepted procedure; a changed SHA alone is not a whole-suite rerun trigger. Push the completed checkpoint and use glab to open a non-draft MR targeting staging with source deletion enabled; collect automatic MR fast evidence.

The current Spec r3/Plan contract requires applicable post-integration staging disposition, not a separate live WorkOS or second hosted-full campaign. Final Goal ACCEPT requires reviewed completed Result, MR evidence, owner-authorized integration and its qualifying staging disposition with no known failed required checks. G008 remains unreleased and G021 remains closed. Codex changed management records only; no application files or tests were modified or run during review.

## Post-integration Duo disposition — 2026-10-11 — CORRECTION LOOP

Reviewed integrated source `e5b1823` and Duo !18 note 3993788301 (repeated in
3993992512/3994034607). The availability concern is valid: policy writes request
the Organization's FOR NO KEY UPDATE lock before the capability decision;
selected/Organization-context Group writes can similarly lock the target Group
before denial. A live subject lacking the capability can consequently contend
with authorized writes. This does not permit an unauthorized mutation; ordinary
guarded reads are lock-free and are not blocked by these row locks. The warning
must describe the actual affected operations rather than all guarded reads.

Retain the reviewed locked-facts decision/mutation and lock order. A speculative
pre-decision is additional transaction/authorization engineering, not this
correction. Under PG-001's existing owner risk disposition and absence of public
tenant adapters, accept documentation as Duo's offered disposition: append to
AUTHORIZATION.md and PERFORMANCE.md that denied exclusive-write requests can
contend, and future adapters must provide per-subject admission/rate limits and
bounded request/transaction lifetime before exposing these operations. Do not
claim those mitigations already exist, introduce numeric performance guarantees,
or change authorization, SQL, schema or indexes.

Authorize bounded documentation correction attempt 3 within existing B7, no
Plan revision or new runtime tests. Owner explicitly directs direct staging
publication with no local test runs; hosted evidence will be reviewed later.
Record this exception and append Result evidence. Codex final Goal ACCEPT is
still pending applicable staging disposition and review of the completed handback.
Duo's earlier truncated no-finding reviews are retained, not comprehensive proof.

## Attempt-3 documentation correction — 2026-10-11 — ACCEPT

Accept correction readiness at `7c4e8a2` (documentation commit `2ac732e`),
on baseline `c845bdb`: exactly AUTHORIZATION.md, PERFORMANCE.md and the appended
Result. The warning identifies policy/Group operations, pre-decision exclusive
locks and contention against writes; it correctly excludes ordinary lock-free
reads. Required future per-subject admission/rate limits and bounded request/
transaction lifetime are explicit and are not claimed as implemented. No code,
schema, lock order or authority changes. Reviewed file hashes:
AUTHORIZATION.md `0363d9d177e06395e3f1db528081791cbeeb6a928d2b48c27eb5627bda884f22`;
PERFORMANCE.md `e5798a83da933b280debaffcf880c70343d650aca0018ecb900f5eadaf756db3`;
Result `f18462d6569cc9cba5a275bdde3853d590207a9042586d7b167d1502e2edb821`.

Local test runs were owner-waived, not passing evidence. The Result preserves
prior staging cancellation and attempt 3/B7. Claude may publish this reviewed
correction directly to staging under the owner's explicit exception, coordinated
with G051. This accepts the documentation correction, not final Goal completion;
reconcile the integrated content, staging disposition and final handback later.


## Final implementation acceptance — 2026-10-11 — ACCEPT

**ACCEPT G007: completed and integrated on staging.** This closes the previously
conditional integration and AC13/AC14 hosted/final gates. Owner-authorized !18
integration and MR fast success2936306602 are already recorded. The reviewed
implementation and documentation correction are retained on tested staging
`4bab0e796ad6a0bb0a599431cdd35d70e83bc165`; no core/contracts changes occurred
between the G007 merge and this qualification. AUTHORIZATION and PERFORMANCE
remain exactly the accepted attempt3 hashes0363d9d... and e5798a83....

Independently verified automatic staging pipeline2936644423 SUCCESS: fast
17097945983 and gate17097945984 both succeeded on the same commit. All93 harness
obligations PASS, no FAIL/UNMET/duplicate outcome IDs;441/441 controls PASS.
B-POLICY-CATALOG/MIGRATION/IT/ORDERING all PASS. Both ledgers MET,9/9 and69/69,
zero unknown and no unexcepted duplicates. Final job traces show every wrapper
phase0, full publication, registry removed, owned_remaining0 and scrub0.
The pre-publication artifact's pending phases are historical snapshots; final
job exits/traces establish successful finalization. No additional run required.

Prior actual-diff/local evidence acceptance and the AC1–AC14 checklist remain
part of this acceptance. Preserve PG-001's recorded lock/fairness limitations,
denied-write contention and required admission/rate/deadline measures before
future public adapters. No claim those future mitigations exist. All failed,
canceled and correction runs remain historical. G007 closes at attempt3/B7.

Claude may append the final status/hosted reference to its Result and complete
accepted task/worktree/session-marker cleanup after checking live ownership;
no application edits or new tests required. The source remote branch was
already deleted after !18. G008's G007 dependency is satisfied, but this record
does not authorize new G008 implementation. G021 remains closed. G051's separate
45/90 timing criterion does not revoke G007's successful staging gate.


### Final evidence count clarification — 2026-10-11

The qualifying pipeline has93PASSoutcome rows across both jobs, representing89distinct checkIDs because four final checks intentionally occur in each partition. Candidate's pipeline-union confirms this exact overlap. The prior acceptance's “93harness obligations” wording denotes outcome rows; G007finalACCEPT and all criteria are unchanged.
