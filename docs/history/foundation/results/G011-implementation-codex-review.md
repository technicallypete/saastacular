> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G011 implementation — Codex review

## Source and local-evidence review — 2026-10-15 UTC

**ACCEPT — implementation/source and local evidence for MR integration.**
This is not final Goal ACCEPT: required MR/Duo, staging disposition and cleanup
remain. Owner's bounded autopilot mandate authorizes the normal integration loop.

Reviewed candidate `ce3184546a5bdc7a4b91abc2ecaff49b34e1cfe2` on staging base
`96fd0fb3ed83b866c0af8c970444edd47c536d68`, exact submitted Result SHA-256
`506e8217a97c978a512354064521ddc2839774a35ca6db3888b3224767a4b524`.
Imported the Result byte-exactly into the integration checkout. Plan r2 remains
`72042f5ce1682d7ba5b6ca2cce8851a56c49c70117880ab73ec4d82ba5907d3a`, scope
`e57d7a675a7de7d0f515fab85e7fe9bb25265d8341ccb9216760587cc07144e0`.
Migration 00008 remains the accepted deviation's bytes:
`b1dedf1f56f28c97a9f8c59796ff1b615c1ed79e6b8ab8f2cfb602098c597233`.

Codex inspected actual diffs, contracts and Result, with two bounded independent
read-only reviews for source/security and attributable validation. No material
blocking findings. No implementation rewrites, new tests or broad runs by Codex.

### Findings and criterion coverage

- AC1–AC3: finite 29-action producer inventory, constrained event shapes and
  server attribution; controlled descriptive-column append, source separation,
  runtime immutability and FORCE RLS. Prior migrations and public contracts are
  unchanged; no new role, vendor or SECURITY DEFINER function.
- AC4/AC5: ordinary mutation/sign-in events share the mutation transaction.
  Revocation's two narrowly contained audit-insert exception blocks preserve
  the owner-approved security priority; gap recording follows commit and is
  bounded. Failed/unknown commit never becomes success or an automatic retry.
  Cancellation/poisoned-context disposal retain existing safeguards.
- AC6: signed issuance commits its audit before returning a URL; staged provider
  effects, compensation and uncertain outcomes remain distinct. Provider calls
  remain outside audit transactions, with no claim of distributed atomicity.
- AC7/AC8: owner-only current-authority read through existing guarded/RLS paths;
  tenant/window-bound cursor and bounded pages. Inspected E-AUDIT observations:
  Group append 3 reader calls; logout-all one capture / three rows / one marker
  read; renewal no audit rows; tenant-time index used on two tenants/220 rows.
  No E-AUDIT-TRIGGER in b4/b5/b6. PG-004 observations are not production SLOs.
- AC9/AC10: privacy/canary checks, restricted future maintenance handoff, additive
  populated/empty migration cases, populated rollback refusal and restart
  persistence. No runtime retention/deletion path or source-audit cascade added.
- AC11: independently reconstructed exactly 101 required PASS IDs: 74 retained
  from b4, 5 retained from b5, 22 from b6; no missing/extra/duplicate IDs. All
  cited log/results hashes verified. The seven b4 failures and five b5 failures
  remain historical failures, not an overall passing b4/b5 claim. Rebases only
  changed management content; stale unit-count fix was separately covered;
  final commit is Result-only with a zero-finding scan. Reuse is attributable.
  Three new audit checks, minimal path mapping and expected-count pins preserve
  covering checks and 441 controls; no control weakening found.
- AC12: read-only security review and F1 fixes verified. F2 accepted deviation
  remains closed with its prior process violation and privilege/trust limits
  recorded. Final integration/cleanup/acceptance obligations remain outstanding.

All 43 implementation paths plus the Result fit approved scope. B-PROTECTED's
536-file evidence remains bound to its original base, with subsequent staging
changes reconciled separately. All three run leftover inventories are empty;
this does not mean held focus images, cache volume and evidence are closed out.
Attempt 1/B7; checkpoint after attempt 4. No new attempt or budget extension.

### Authorized next actions and remaining gates

1. Coordinator fetches/reconciles current staging, verifies scope/content/input
   provenance and opens the feature MR to staging via glab, retaining source
   deletion. Management-only review import does not require duplicate broad
   validation; remember whole-repository scanners read management content too.
2. Obtain MR fast and Duo findings; Codex dispositions material findings before
   integration. Owner's autopilot authority permits merge/auto-merge once these
   review gates pass. Do not merge through a known failed required check.
3. Required post-integration staging result and exact merged-content provenance
   precede final Goal ACCEPT. Preserve failed/canceled status honestly; current
   missing/running staging evidence is not acceptance.
4. Finish compact Result/cleanup receipt: owned focus images, g011-gomod, raw
   temporary evidence/credentials/scripts and supported worker/branch cleanup.
   Preserve native-locked held paths, unrelated services and unproven volumes.
5. G012 planning becomes eligible after verified G011 integration, provisionally
   while final staging is pending. G012 execution requires final G011 ACCEPT,
   reviewed Plan/current-input/budget/resource release. G013 then follows its
   dependencies. Hard stop after G014 Goal/Spec for owner input remains.

G011 is not yet marked completed. G021 remains closed. This review authorizes
no G014 planning, production deployment or independent held Goal work.

## MR fast and Duo disposition — 2026-10-15 UTC

**ACCEPT — remaining pre-integration review hold cleared.** MR !27 head
`d27dfc2aab8c505d716830bdefefda1bb3d75b65` contains current staging `c3e0a85`;
its entire non-management tree is identical to reviewed `ce31845`. The imported
Result is already upstream byte-exactly. MR fast pipeline `2948805800` succeeded
on that head; optional/manual MR gate was not run and is not claimed passing.

Inspected current GitLab discussions: Duo note `4016564961` reports no defects
and no unresolved blocking thread. Its approximately 34% diff visibility remains
a stated limit, not comprehensive coverage or a substitute for Codex's prior
source/evidence review. Independently confirmed both requested call-site checks:

- The sole production caller of `tenancy.InitContext` is access/tx.go's private
  transaction initializer. Tenant operations use parsed Organization selectors
  and service-issued/bound contexts; storage uses those same validated fields.
  Explicit Organization-less contexts are limited to the reviewed independent
  audit recorders. A Group without an Organization still fails initialization.
- Both production callers of the two `sqlInitContext` constants supply all six
  parameters, including request/trace correlation; both scan seven results.
  No stale four-parameter caller remains. Tests and integration evidence cover
  the context boundary; no new test campaign is required for this confirmation.

No correction or owner architecture choice remains before integration. Owner's
standing auto-merge/bounded autopilot authority permits the exact-head merge to
staging with temporary source deletion. Post-merge staging, closeout and final
Goal ACCEPT remain required; G012 may begin provisional planning after merge.

## Conditional staging integration — 2026-10-15 UTC

MR !27 (not included)
merged under the cleared review/owner auto-merge authority at
`1af469634225e6e6527aca485b40aa55953227a8`. Verified the merge's entire tree is
identical to reviewed/reconciled source `d27dfc2`; remote temporary source branch
is absent. Main local staging fast-forwarded; no application edits by Codex.

Staging pipeline 2948834614 (not included)
is running on that exact merge. **Not final Goal ACCEPT.** Cleanup/final Result
and the qualifying staging disposition remain; no next-task execution release.
G012 provisional planning is now eligible and requested from the existing
coordinator at xhigh, binding actual integrated G011/current Goal/Spec and holding
implementation until final dependency ACCEPT and its own reviewed Plan/inputs.
No G054/G055/OS001 hold changes; G014 specification human stop remains.

## Final implementation acceptance — 2026-10-15 UTC

**ACCEPT — G011 completed and integrated.** All twelve acceptance criteria
satisfied. One of B7 attempts used; checkpoint after attempt 4 did not arise.
The prior source/security/local review and F2 disposition remain valid.

Codex and a bounded read-only reviewer independently checked GitLab/API and
artifacts in memory: staging pipeline
2948834614 (not included),
fast 17186812828 and gate 17186812829 all SUCCESS on exact merge
`1af469634225e6e6527aca485b40aa55953227a8`. Merged tree
`c312478966715a729d1e5431dd567f8c275b3cb6` equals reviewed/reconciled source.

- Fast 33 PASS plus gate 72 PASS/zero FAIL cover exactly 101 required IDs; only
  four intentional final-check overlaps, no omissions/extras/duplicates.
- B-AUDIT-CATALOG/MIGRATION/IT PASS, TestAudit 9/9, inventory 29/29, no actual
  E-AUDIT-TRIGGER lines; CI controls 441/441 PASS. E-RLS 40/E-OBJ 2 remain the
  separately accepted pre-existing observations, not new unexplained changes.
- Gate 2695.877278 seconds (44m56s); pipeline 3145 seconds (52m25s). GitLab job
  interval 00:21:40–01:06:36 UTC differs from the harness window quoted in the
  Result; neither interval is mistaken for the other.
- All four Result artifact digests match actual bytes. Artifacts expire
  2027-01-13; no local bulk mirror retained. Final job status/trace, rather than
  a pre-publication summary status, establishes success. Shared renderer cache
  in hosted inventory is not an owned G011 focus-resource leak.

Imported [final Result](G011-durable-audit-events-result.md) byte-exactly from
report-only `02b709962f13f932aaa0262faf2cae6c7880a43e`, SHA-256
`e62069daa0107d82c08bf0f15f63bba47a5e0d8a39c7f3382689e1e0e3b36572`;
its earlier pending labels and Closeout owed are superseded by this final
disposition and the separately dated receipt, not silently rewritten.
Cleanup receipt (not included) imported byte-exactly, SHA-256
`95e0cd46ef819e2a45c10435f376e837d052e2d828ee2578cc0b40f41940ce02`.
Independently confirmed G011 implementation/planner worktrees and local branches,
focus MinIO images, g011-gomod volume and both bulk evidence folders absent.
No raw-proof hold remains. Temporary remote closeout can be removed after this
import; remote plan history and two unproven small temporary review copies are
explicitly held. Native-locked G010 planner, G054 and unrelated/unproven resources
remain protected. No global prune or force-unlock is authorized.

Historical b4/b5 failures, the fixed staging prose scan and F2's prior missing
escalation remain accurately recorded. No new validation campaign was required
for report import or later management-only commits; application/validation trees
remain those tested at the merge.

G012's predecessor gate is now satisfied. Reconcile its pending r1 Plan against
this final accepted G011; remaining r0 corrections, feed disposition, Plan/input
and resource gates still precede execution. G013 receives audit immutability/
restricted maintenance, unknown-outcome and sensitive-backup responsibilities.
EPIC-004 remains open through G013; G021 remains closed. Owner's G014 Goal/Spec
human-input stop remains, with no G014 planning or implementation release.
