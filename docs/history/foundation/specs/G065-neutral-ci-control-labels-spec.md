> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G065 — Goal-neutral CI-control labels — Work Spec

Revision: r6, 2026-10-23 — owner selects hosted full pre-merge qualification; r5 snapshot preserved
Goal: [G065](../goals/G065-neutral-ci-control-labels.md)

## Outcome and scope

Mechanically rename the four existing CI-control Docker label keys, preserving
all current values and meaning. Stop affected creators, verify cleanup of their
old-label resources, update their writers/readers together and cut over to the
new keys. Reuse the accepted G064 harness, registry and cleanup mechanisms.

**The baseline rename requirements are R1 and R8. The bounded RC8 repair in the
r5 addendum is also current. R2–R7 remain dropped and are not scheduled.** R9's
general cleanup obligation remains existing governance, not a separate G065
implementation or proof campaign. The earlier r1 (not included)
and r2 (not included) definitions are historical;
r4 is authoritative for the target mapping; the prior r3 snapshot (not included) preserves the approved mapping and timing history.

## R1 — Rename the existing CI-control label keys

| Existing key | New key |
|---|---|
| `g003.controls.run` | `sdlc.ci.run` |
| `g003.control` | `sdlc.ci.control` |
| `g003.outer` | `sdlc.ci.outer` |
| `g003.controls.owner` | `sdlc.ci.owner` |

Preserve values, run/control/nesting identity and owner-key interpretation.
Update the actual producers, readers, queries, current fixtures and runbook
references for these four keys as one bounded change. Existing Compose fixtures
using those keys receive the same rename; do not expand this into adding labels
to every repository resource. Replace the exact keys, not every `g003` string:
unrelated environment variables, storage paths, helpers and historical Goal
identifiers are outside scope. Historical evidence retains its original labels.

After the R8 boundary, active writers and readers use only the new keys. No
dual-label writes, legacy readers, mixed-generation compatibility or subsequent
reader-retirement phase is required.

## R8 — Verified cleanup and coordinated cutover

Use one bounded maintenance boundary after G064 final ACCEPT:

1. Identify and stop affected old-label creators sharing the relevant Docker
   daemon, including nested helpers. Preserve unrelated projects and isolated
   hosted jobs; do not stop ordinary source work that cannot create those
   resources.
2. Use the accepted exact-ID cleanup/recovery mechanism to remove their owned
   disposable resources. Reconcile the existing inventory, pending recovery
   records and queries for the four old keys. Verify no affected old resource
   or live creator remains; record the receipt in the existing Result.
3. Change all inventoried label writers/readers and current references together.
   Do not restart old writers against the shared daemon after the new readers
   are enabled.
4. Verify the new emitted labels, affected checks and cleanup, then resume
   workloads. Record the cutover SHA and outcomes.

Failed queries, uncertain ownership or required live retention hold cutover.
Keep protected/shared/foreign resources safe; an unresolved resource does not
justify deletion or silently adding compatibility support. Old-key absence
queries belong to the one-off cutover check, not a new runtime migration layer.

## Existing safeguards and non-goals

Follow AGENTS.md, CLAUDE.md, architecture/security decisions and the accepted
G064 ownership, resource limits, recovery, cleanup and evidence contracts.
Labels remain discovery metadata; they do not grant deletion authority. Preserve
exact IDs, anonymous/shared-resource protection and failed-run evidence.

This task does not implement the dropped common project/owner/lifecycle label
schema, a repository-wide Compose project-name rollout, additional labels on all
Compose/direct resources, or a new label-based discovery mechanism. Except for
the narrowly specified RC8 K1–K3 repairs below, it does not redesign cleanup,
isolation, scheduling, reporting, recovery or resource limits; create a
framework/database/manifest; rebuild shared images for metadata; or rewrite
historical evidence. Existing governance still applies to task-created resources,
without making the dropped rollout an acceptance requirement.

## Acceptance and validation

Keep the existing criterion IDs; AC3–AC7 from r2 are retired with the dropped
scope and must not become a validation campaign under another heading.

| ID | Required outcome | Evidence |
|---|---|---|
| AC1 | All actual writers/readers of the four keys use the agreed mapping with existing values and semantics. | Bounded source inventory/search, mechanical diff review and existing emitted-label assertions. |
| AC2 | Affected creators are stopped and old resources are verified gone before the coordinated cutover. | Existing exact-ID cleanup receipt, old-key absence query and cutover source; unknown/live retained resources hold cutover. |
| AC8 | Renamed paths retain current ownership/cleanup behavior and leave no task-created leftovers. | Applicable existing fast checks and affected harness cases, source/Result review, exact-SHA qualification and cleanup receipt. Preserve historical evidence and unchanged passes. |

Use existing checks and fixtures for changed paths. No new metadata rollout,
compatibility matrix, resource-type coverage expansion or generalized proof
suite is required. The Claude Plan records concise selectors and why the rename
affects them. A requested existing full run may supply that coverage; do not
add duplicate focused campaigns over unchanged mechanisms. Otherwise apply the
current affected-suite policy. Full local runs remain owner-requested or
supervisor-recommended. The earlier RC8 hosted qualification remains FAILED;
any later candidate has its own full hosted qualification gate.

## Plan, budget and execution boundary

A concise Claude Plan inventories the exact-key replacements and their current
inputs, identifies the cleanup/cutover boundary, and selects existing validation.
The Result records source, actual outcomes, failures, deviations and cleanup.
Use one task owner and the existing B3 budget; the original rename has spent
attempt 1, and the amendment does not reset the attempt ledger.

G064 final ACCEPT and the original G065 Plan gate are satisfied. The added RC8
repair code remains held until Claude's bounded Plan amendment is reviewed by
Codex and current bindings/resources are reconciled. Preserve the exact failed
RC8 evidence and immutable tag. Current sequencing is governed by the latest
owner amendment below and the Roadmap: G065 (including repair) precedes G020;
the historical RC8 snapshot remains FAILED, and a later RC candidate requires
full hosted qualification. A merge is not final Goal ACCEPT or release authority.

## Revision history

- **2026-10-23, r6 — Hosted full pre-merge qualification.** The owner waived
  G065's local pre-merge qualification and selected the hosted full profile,
  including manually triggered `validate:integration` on the exact MR SHA.
  The r5 Spec is preserved in the r5 snapshot (not included).
- **2026-10-23, r5 — RC8 harness repair added.** The owner directed the bounded
  RC8 findings into G065. The previous r4 Spec is preserved in
  the r4 snapshot (not included); the addendum at
  the end of this file requires a Plan amendment before added repair code.
- **2026-10-23, r4 — Product-neutral mapping.** `sdlc.ci.*` replaces the
  identity-bearing target namespace; preserved in the exact r4 snapshot.
- **2026-10-23, r3 — Scope narrowed by owner.** Verified direct exchange in
  coordinator session `<redacted-uuid>`: owner message
  `<redacted-uuid>` at 08:01:02.900Z selects the preceding
  proposal of R1 + R8 only; the 08:01:23 handoff explicitly drops R2–R7 without
  scheduling them. AC3–AC7 and the common-metadata deliverables are removed.
- **2026-10-23, r2 — Clean cutover.** Owner selected cleanup followed by a
  coordinated label change, replacing the compatibility period. Preserved in
  the r2 snapshot (not included).
- **2026-10-22, r1 — Original definition.** Broad metadata and compatibility
  requirements, preserved in the r1 snapshot (not included).


## Owner start clarification — RC8 independent, 2026-10-23

The owner states: “Rc8 qualification is not a blocker.” G064 FINAL ACCEPT is
complete. Claude planning is released now; Plan ACCEPT and current-input/resource
reconciliation remain the implementation gates. G065 may proceed independently
of RC8 qualification through its normal integration and final acceptance gates.
This supersedes the earlier timing/order language above; R1, R8, AC1, AC2, AC8
and the existing B3 budget are unchanged. Use existing affected checks with
explicit selectors and expected runtimes; no blanket local full run is added
for the rename. RC8 remains pinned to its original source.


## Owner validation-timing exception — local pre-merge qualification (superseded by r6)

Historical decision: the owner stated “This can also use pre merge instead of post merge.” For G065,
run the selected affected local checks on the reconciled candidate before merge;
that evidence may replace the post-merge local run. Record the tested SHA/tree,
commands, outcomes and cleanup, then verify the actual merged SHA/tree against
them. Matching trees require no duplicate post-merge execution. If the candidate
or integration changes after testing, inspect the actual delta and rerun only
affected checks; document any reuse of unchanged evidence, including management-
only differences. Do not silently transfer results to changed executable inputs.

This supersedes G065's default post-merge placement and the review's request for
a separate merged-SHA selection. It changes timing only: keep the focused
coverage, source/hosted review, cleanup and final acceptance gates. No local full
or 528-control run is added, and RC8 continues independently.


## Namespace correction — product-neutral keys, 2026-10-23

Revision r3's approved target namespace contained the current product identity,
which the existing `B-LITERALS` guard forbids in these source/runbook paths. Use
the corrected R1 table above (`sdlc.ci.run/control/outer/owner`) with the same
values and semantics. Keep the literal guard and architecture invariant intact.
The r3 snapshot (not included) preserves its original
bytes and failed mapping; do not rewrite prior implementation/failure evidence.

This is a supervisor-authorized metadata mapping correction within the released
G065 scope and existing B3 ledger. It supersedes only the target keys in accepted
Plan r2 `4b28d20e`; its inventory, cutover, focused pre-merge checks, cleanup and
release remain valid. Claude records the mechanical amendment in the existing
Plan/Result and continues without another planning/approval round. Reconcile
exact inputs at the safe boundary and rerun failed/affected checks using the
corrected keys, including escaped forms and the temporary accounting proof.
At cutover also reconcile any objects actually created with the intermediate
target keys using the same exact-ownership cleanup; no compatibility layer or
extra runtime reader is added. RC8 remains independent.


## Owner RC8 harness-repair amendment — r5, 2026-10-23

The owner instructed “Fix rc8 issue with 065” and clarified “Add it to G065.”
This amendment adds the bounded harness repairs below to the accepted four-key
rename and verified cutover. It supersedes only the earlier statement that RC8
is independent of G065's repair scope; G065 may proceed without waiting for
RC8, but its final acceptance now includes these repairs. The exact r4 bytes are
preserved in the r4 snapshot (not included).

The retained RC8 candidate Result (not included)
and repair proposal identify these bounded code corrections:

1. **K1 — fail-closed image cleanup/accounting.** In the existing CI-control
   image listing, count and cleanup helpers, use a consistent all-image listing
   that includes untagged images. A failed listing or image-removal command must
   become a recorded control failure/hold, never an empty listing or success.
2. **K2 — record Compose-built test images.** After the existing VC8/VC9 Compose
   build paths create a labelled image, inspect its exact image ID and record it
   in that control's ownership inventory before cleanup/finalization. This closes
   the observed gap where the final sweep found an image with no inventory row.
3. **K3 — preserve results and clean up on persistence-start abort.** If the
   persistence startup command fails, record the affected check as failed but
   still join/collect background checks, export available suite results, and
   stop/reap the run witness on the abort path. Do not let an unguarded command
   discard completed results or leave a task-owned witness for wrapper cleanup.

The registry/frontend authorization `DeadlineExceeded` in the RC8 hosted log is
a separate, unproven infrastructure symptom: the same frontend resolution
succeeded elsewhere in that job. Do not add retries, change timeouts, remove or
pin the Dockerfile frontend, or otherwise patch that symptom without new evidence.
The RC8 tag and failed qualification remain immutable and FAILED; this amendment
does not qualify, retag or promote it.

## Added acceptance criterion

| ID | Required outcome | Evidence |
|---|---|---|
| AC9 | K1 reports listing/removal errors as failures; K2 records and cleans the Compose-built images; K3 preserves completed results and removes its witness after a failed persistence start. Successful cleanup leaves no unrecorded image or task-owned abort-path resource. | Plan-bound focused selector outcomes, exact image/registry records, forced-start-failure proof, resulting per-check ledger, and verified Docker inventory cleanup on the tested candidate. |

The Plan amendment must identify the executable selectors and reproduce the
observed Docker 29.8.2/Compose 5.5.1 image-listing/build path in a bounded
comparison. Run the proposal's affected selector sets: `VC1,VC4,VC8,VC9,VB2,V18n`
for K1; `VC8,VC9` for K2; and `VB2,VB5` plus a disposable forced persistence-
start failure for K3. It must specify cleanup/retention evidence for every run.
Use a focused Docker-in-Docker 29.8.2 reproduction to match hosted daemon behavior
if the local daemon cannot reproduce that path. No 528-control campaign or blanket
local full suite is added. The current pre-merge timing exception applies:
selected affected checks run on the reconciled candidate, and the merged tree is
compared before reusing evidence.

Claude prepares the Plan amendment and Codex reviews it before the added code
starts. Existing B3 budget and attempt ledger remain binding: one attempt was
already spent on the rename; no reset or implicit budget expansion is authorized.
If the Plan determines the remaining attempts cannot contain the bounded work,
stop for an owner decision. Keep MR !87 draft until amended Plan acceptance,
implementation, affected checks, hosted fast pipeline, review disposition and
input/tree reconciliation are complete. A subsequent RC candidate receives its
own full hosted qualification; this G065 amendment does not authorize a tag or
main promotion.


## Owner hosted-qualification amendment — r6, 2026-10-23

The owner directs: “waive the local pre merge run, and just trigger
`validate:integration` on CI.” This supersedes the local pre-merge selector
requirements in the r4/r5 timing sections for G065. Do not require or launch a
local pre-merge suite for G065 qualification. Existing local evidence for the
original rename remains historical and reusable only for unchanged source; it
does not qualify the added K1–K3 repair code.

G065's pre-merge qualification is the hosted full profile on one exact final MR
SHA: the automatically run `validate:checks` must pass, then the coordinator
manually starts the optional `validate:integration` job in that same MR pipeline
and verifies it passes on that same SHA. The two jobs together supply the full
profile selected for the `infra/ci` and `infra/validation` changes. Though GitLab
marks `validate:integration` optional/manual on an ordinary staging MR, it is a
required G065 acceptance check under this owner instruction. Do not treat the
existing fast-only pipeline on the pre-repair !87 head as qualification for the
expanded G065 scope.

After merge, verify that the merged executable trees match the hosted-qualified
SHA; do not run a duplicate local post-merge suite on a match. Documentation or
management-only reconciliation is recorded without rerunning executable tests.
An executable change after the hosted full run changes its exact input and must
be requalified before merge under the accepted CI policy. Local tests may be used
for diagnosis when needed, but they do not replace this hosted result. The full
hosted run belongs after the K1–K3 code is implemented/reviewed and before MR
merge, not on the current incomplete head.

The owner also directed closing RC8 promotion MR !86. It is closed; its tag,
failed qualification and evidence remain unchanged. RC8's separate failure is
being repaired within G065, and no tag or main promotion is authorized by this
qualification amendment.
