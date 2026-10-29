> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G017 implementation review — 2026-10-17 UTC

**CORRECTION LOOP — two bounded web corrections before merge.**
Reviewed MR !40 candidate `3d715bc24f72258766c07cdfb9bd6d07fc1038b7`.
Owner requests merge; existing pre-integration gates remain. Required MR checks
2958686504 SUCCESS and actual account selector patch verified. Owner visual
feedback is positive, not a claim of manual functional upload execution.

Two bounded read-only source reviews covered core/security and web/BFF. Codex
inspected the affected code, Result, scope and count evidence. No tests/builds or
application writes by reviewers; reported passes remain attributable evidence,
not independently rerun. Prior isolated dropdown acceptance stays valid.

## Required corrections

1. **Unconfirmed session rendered as signed in.** web/app/page.tsx:40 tests
   indicator kind != anonymous, which includes unavailable from the existing
   session error/malformed/network branch. It then renders “You are signed in.”
   and Account. Require confirmed member or none explicitly; unavailable must
   not assert authenticated identity. Preserve normal anonymous/member/none UX.
   Reuse the existing context-indicator unit cases and inspect the four-state
   predicate; no new helper/framework or whole browser campaign is required.
2. **Valid wide logo can overflow management UI.**
   web/components/branding/logo-forms.tsx:37–39 fixes preview height but not width.
   A supported2048×16 logo scales far wider than the320px viewport. Bound box/image
   width with containment, preserving aspect ratio and accessible alt behavior.
   Reuse one existing affected E-BRD logo/reflow case with a legal extreme aspect
   ratio or equivalent focused actual-render evidence; do not run an umbrella suite.

Both production paths and relevant existing test files are already in scope.
Use the current sole writer, preserve completed work/ledger and count the round
under existing B5/checkpoint3 rules without a reset. Return exact correction diff,
focused outcomes, updated Result and completed required self/security review.
Codex delta review and current MR checks/Duo disposition precede auto-merge.

## Settled source review and evidence dispositions

No material core/security blocker found in the reviewed normalization, versioned
guarded operations, unknown-commit retention, public projection/assets, migration
RLS/grants/definer boundaries, audit or origin-bound session/membership delta.
No further material web/BFF issue found in CSRF/origin, permission forwarding,
bounded decoding/upload and cache/privacy boundaries. Preserve known accepted
24h cached public logo-byte retention and PG-008's non-production estimate limit.

Duo reports no findings, but saw only about24% of the diff, excluding web/BFF.
It is supplementary; the two source reviews cover the omitted surfaces. Its count
query is resolved: actual it_count counts every successful TestName/ subtest line,
excluding the parent. Existing catalog log SHA63cc9ec3c8d3ff84fc11c23c6b9324da05f123a8ad46758922594df92287ecf4
has3 such passes and no failure; branding log
9edf737491497b1bb63741bb8b3cc1cad3779d105814f50e767448585c9c12f7
has23 (14 top-level +9 nested), no failure. The3/23 pins preserve actual coverage;
4/15 summaries include parent/different granularity. No CI meta-test rerun needed.

Design notes4/6: accept factual lookup-ID documentation in existing WEB-SHELL;
no OPERATIONS scope addition is needed. Supplemental read IDs are diagnostic only,
not identity/auth authority, and must not falsely claim one shared request ID.
Note7: logo evidence in E-BRD-5 rather than E-BRD-1 is equivalent covered behavior.
Note8: PRG303 with closed status messages is acceptable; core refuses mutations
with existing errors, session/CSRF/origin failures keep fixed refusal behavior.
Note9: owner-approved existing lucide icons supersede filled-icon preference.
These are bounded dispositions, not new product/API/security contracts.

G017 is not accepted or authorized to merge while these two corrections and final
review/evidence gates remain. Keep current auto-merge intent and enable it only
after current-source pre-integration ACCEPT; final Goal ACCEPT follows staging
qualification/closeout. G018 execution and G019/G058 phase holds remain.

## Security self-review D-1 — correction added, 2026-10-17 UTC

Verified self-review SHA55e53467c058a1788072a29306bf07098a9c4ed6f05394682881a40cc5932e58
and actual SetBrandingLogo/facade order. **D-1 requires correction in the current
round.** The pre hook calls NormalizeLogo before live-session/capability checks;
a plain member can cause bounded but expensive decode/re-encode before refusal,
contrary to the accepted owner/admin resource-containment premise.

Move NormalizeLogo after the existing successful guarded authorization and
expected-version pre-check, before stageP/PutObject and outside any database
transaction. Keep cheap bounded body/length validation and existing closed
validation errors; do not hold locks during image work or relax later upload/
association reauthorization, audit or unknown-commit handling. No global guard
refactor or new quota/semaphore is requested.

Existing-scope evidence: BRD-9 member+malformed image returns forbidden rather
than image-validation failure; existing authorized invalid-input refusal and
normal normalization remain covered by affected BRD-2/8/9 cases only. No full
migrations/ops/controls or broad integration rerun. Same current correction
owner/ledger; no reset or additional framework. Return the minimal delta with
the two web corrections, then focused source/security delta review before merge.

H-3 (normal account navigation emits a permission-probe audit denial) remains
under bounded disposition; no audit suppression, new capability contract or
show-management-to-all change is authorized by this D-1 release. H-1's JPEG CPU
claim is unverified, not grounds to silently restrict accepted image formats.
Other suggestions/known limits are assessed separately without blocking D-1 work.

## Remaining self-review notes disposed — 2026-10-17 UTC

**H-3: nonblocking audit-noise follow-up.** Normal member /account navigation uses
a guarded GetBranding probe to hide the management link, producing a durable
denial. Current session/public projection exposes no owner/admin hint. Showing
the link universally conflicts with accepted visibility; suppressing GetBranding
audit would weaken the current contract. Preserve current authorization/audit
behavior. Record bounded G021 hardening assessment for a fresh core-derived
read-only display hint through a reviewed response; no route/helper/grant
expansion is authorized now. This is operational signal/volume noise, not a
confirmed unauthorized grant or data leak, and does not block current correction.

**H-1: unresolved resource-hardening assessment, not a demonstrated runtime exploit.**
A bounded read-only check of exact Go1.27.2 JPEG sources confirms repeated SOS
dispatch and per-scan block traversal, without an evident global scan-count cap:
[dispatch](https://raw.githubusercontent.com/golang/go/go1.27.2/src/image/jpeg/reader.go),
[traversal](https://raw.githubusercontent.com/golang/go/go1.27.2/src/image/jpeg/scan.go).
Practical cost for an accepted512KiB/2048² input is unmeasured. D-1 removes plain
member access to expensive decoding; existing input/dimension/output caps remain.
Record focused assessment before G021/real deployment. Do not silently impose
an arbitrary64-scan input-contract restriction or launch a new campaign now.
No image-CPU production-performance qualification is claimed by G017/PG-008.

**H-2:** same-Organization authorized deletion/current-logo races resolve through
the accepted finite live-object fallback contract; preserve it, not a new deletion
transaction system. **H-4:** ready filtering in the definer contains auth-reader
metadata exposure, and blank/reset reason wording is cosmetic. No change required.
These notes add no blocker beyond the two web corrections and confirmed D-1.
The current correction must still return source/evidence for delta review.

## Correction delta — pre-integration source ACCEPT — 2026-10-17 UTC

**ACCEPT source for conditional integration**, exact candidate
`c762f2cde41a0181b36c6d4bcc46920302a426fb`; owner explicitly requested merge.
Two bounded read-only delta reviews plus Codex inspection close both web findings
and S-review D-1. Exactly six non-management paths changed from3d715bc2:
entry predicate, logo containment, legal2048×16 fixture/one E-BRD-5 extension,
normalization order and BRD-9 member negatives. No unrelated application delta.

Normalization follows current role/liveness/version authorization, outside DBtx
and before upload; later checks/audit/P/A/R/C remain. Confirmed member/none only
can render signed-in copy. Actual CSS/markup wide render at320px reports box0–320,
image10–310, contain and scrollWidth320; reviewers read those metrics. Source
E-BRD-5 now tests the real2048×16 fixture end to end. Its local run stopped before
the new case due to stale preview emulator subjects, and is NOT a local PASS.
The equivalent isolated render is sufficient for this bounded presentation fix;
required post-merge staging evidence must exercise the amended real path. No
permission refusal was bypassed or live preview DB rewritten.

Imported exact submitted Result SHA
`1b79d03606c09d12a256eeee500fa82f3ace1c0159e6521c8fbd57e5b0dfcebc`;
previous local Result preserved as linked result-r0. Reported focused evidence:
web lint/types/unit164PASS and exact predicate truth table; BRD-2/8/9 PASS3/3,
core build/gofmt/vet; prior affected appearance/account evidence retained. No
reviewer reran tests. M1 actual screenshot feedback accepted; no invented manual
functional actions. Preview auth failure does not undo earlier bound evidence
or authorize another repair/test campaign. Owned preview cleanup stays required.

Duo's existing no-findings review is supplementary and truncated; Codex's complete
core/web and exact current-delta reviews supply coverage it omitted. Count query
and all design/hardening notes are disposed above; no open defect is inferred
closed merely from Duo or CI. Reinspect latest discussions before scheduling merge.
No second blanket Duo request is necessary for this fully reviewed bounded delta
without a new finding; do not claim Duo reviewed more than its actual source.

**Integration authority:** mark MR !40 ready and enable SHA-bound auto-merge with
source deletion under owner/standing authority once the existing native required
checks permit it. Pipeline2958730636 was running at review; do not manufacture a
pass or start the optional MR integration job. Current target's management-only
review/import changes require reconciliation, not an application rebase/test
restart. Preserve this source SHA unless a material conflict appears. Staging
integration runs automatically after merge and remains required for final ACCEPT;
G018 execution and G019/G058 phase holds remain until their actual gates close.

## MR !40 ready; SHA-bound auto-merge verified — 2026-10-17 UTC

Posted source/evidence/Duo disposition note4035265591. Marked MR ready through
the supported CLI; scheduled auto-merge for exact reviewedc762f2cd with source
deletion. GitLab independently reports draft=false, merge_when_pipeline_succeeds=true,
should_remove_source_branch=true, detailed_merge_status=ci_still_running and
headpipeline2958730636 running atc762f2cd. MR is OPEN, not claimed merged.

The installed CLI's first raw JSON note request returned400 without a created
note; corrected through supported file-backed body field, then ready/auto-merge
commands succeeded. No policy/permission refusal was bypassed. No optional
MR integration job or duplicate pipeline was started. Current source/review/input
gates are complete for conditional merge; required CI must finish successfully.
Notify coordinator to watch terminal state and stop/reassess for an actual new
finding or source change. After merge, verify source deletion and attributable
automatic staging checks before cleanup/finalG017ACCEPT and any G018 release.

## Staging qualification failed; predecessor diagnosis — 2026-10-17 UTC

Independently verified pipeline `2958742766` on merge `3cc1822826754ff70cf6346809f62b417080832c`:
checks passed; integration job `17256744765` failed after 3873 seconds. Published
artifacts are available. Six distinct required rows fail, plus the aggregate:
B-BROWSER-ROOT, B-IDENTITY-CATALOG, B-IDENTITY-E2E, B-OBS-OFF, B-BROWSER-OFF,
B-TOGGLE-OFF; B-REQUIRED-CHECKS is consequential. No final Goal ACCEPT.

Direct artifact/source observations, not a complete repair disposition:

- Root browser failures are S-BOUNDARY in both themes; its source still clicks
  text Dark/System rather than using the newly selected appearance control.
- Identity-catalog Go test passes all four subtests; harness identity-object
  inventory omits G017 branding relations. Full diagnosis must confirm exact delta.
- E-BRD-5 fails document horizontal-scroll assertion at320px. Prior local actual-logo
  render evidence did not qualify the complete page; preserve this failure and
  diagnose the actual overflowing element, retaining the assertion.
- Disabled variant start fails its production build: missing `@/lib/branding`
  imported by copied G017 pages. OBS-OFF/browser-off failures follow missing startup.
- Disabled check log separately fails TestDNSSystemChecker (CNAME chain returns
  terminal target instead of expected first edge). This is an in-process DNS fixture,
  not external DNS evidence; G016 source is unchanged. Do not label it a transient
  or weaken the edge bound without the source/security disposition.

Coordinator has already dispatched one Fable-high read-only G017 diagnosis; reuse
it, do not duplicate it. Return consolidated cause/scope/focused repair proposal
and current G017 ledger for Codex disposition. No blind gate rerun or blanket
product/harness rewrite authorized by this observation. G017 owner retains repairs
on its branch/ledger in parallel; G018 continues approved unaffected work. Pause
only work shown unsafe by relevant changed contracts/resources. G017 final ACCEPT
continues to gate G018 integration; later autopilot stops before G020 planning.
No source edits, tests, builds or additional pipeline were run by Codex.

## Post-integration qualification verified — 2026-10-17 UTC

Supervisor independently verified MR !42 merged source173d0511 into staging
`a33029dd3f50a51ff8158a17a93d35d8d9d8fbc0`, source branch deleted. Core, web and
run.sh blobs match accepted0ead89db. Later staging policy/ignore/OS001 changes
contain no branding/DNS runtime delta and do not invalidate this evidence.

Pipeline 2959201147 (not included)
qualifies that exact merge/tree5273b764. Both jobs independently report success:
checks17259440630 (35 PASS, B-DEPSCAN zero blocking unexcepted, one existing visible
exception) and integration17259440631 (84 PASS/0 FAIL, selected required set exactly
once,441 CI controls PASS). The integration artifact summary is explicitly
pre-publication/partitioned; successful terminal job/API status and both attributable
partitions establish final pipeline success, not the summary file alone.
Downloaded integration archive SHA-256
`722e4bce730b72fa89f129521a42d6b917cc25ffb04a10b7040c0f462f99febe`;
raw duplicate is disposable after this review, hosted artifacts currently available.
Attribution accepted exactmerge; permitted render-cache volume is the only hosted
leftover. Previously failing browser-root/off, identity-catalog/E2E, obs-off and
toggle-off rows now pass. No new supervisor test, build, probe or rerun.

Latest MR !42 Duo review has zero findings and no unresolved resolvable threads.
Dispose its advisory about narrow loopback-test timing windows as non-blocking:
current local/hosted tests pass and no defect is demonstrated. Keep truthful timing
limits; a future evidenced failure gets a targeted mechanical correction, not a
speculative broad campaign. B-DEPSCAN and prior module/source review cover the
checksums that Duo did not verify. Independent bounded reviewer concurs source and
qualification clear; updated Result and cleanup receipt remain required by AC10.

This qualification disposition is not yet final Goal ACCEPT. The coordinator has
dispatched one management-only G017 closeout worker under standing cleanup authority.
Preserve live G018 M1/forwarder and coordinator resources; do not infer M1 completion.
Final acceptance and EPIC-006 closeout follow the reviewed Result/receipt. No G020 planning.

## Final implementation ACCEPT and EPIC-006 closeout — 2026-10-17 UTC

**ACCEPT — G017 is finally accepted.** Accepted implementation and repairs are
integrated at `a33029dd3f50a51ff8158a17a93d35d8d9d8fbc0`; successful staging
pipeline2959201147 and the preceding source/evidence reviews satisfy AC1–AC10.
The original failed gate remains historical; it is not converted to a pass.
Budget finishes at correction attempt2/B5, checkpoint3 unchanged.

Imported exact closeout Result from `54069a09b66be90ac8bd985763d1585674ea5e72`,
SHA-256 `5d774d88d82f5a10d35bb4704a5df232a198495fb11eae8419e214f7b0840177`,
retaining all prior bytes. Compact cleanup receipt (not included)
records inspected full receipt046788eb, task-resource removal and verified current
absence of G017 containers/volumes/images/worktrees. Independent reviewer confirms
all14 selected proof copies match originals and all3 branch bundles match receipt
hashes. Unique management history is preserved; coordinator/G018 M1 resources
remain active and protected. No remaining source, validation or safety blocker.

This ACCEPT releases the temporary review hold on duplicate/raw local G017 proof.
Coordinator completes normal post-import branch/scratch/proof cleanup without
another permission or qualification loop, preserving only the recorded unique-
history bundle/ref holds until their exact management content is reconciled.
Keep committed Result/review/receipt and hosted links. Report terminal removal,
not a claim that deleted raw bytes remain available. No additional tests required.

**EPIC-006 is closed.** Required compaction checkpoint performed:28 superseded
G017 progress entries plus the now-superseded qualification-pending entry are
archived with original decisions/failures, adjusted relative links and current
acceptance links. Active G018/G019 records/inputs remain; notify coordinator of
this management-only advance, not another application validation campaign.

G017's dependency gate for G018 is closed. Actual owner G018 M1 feedback and its
remaining MR/review/CI gates still precede G018 integration/final acceptance.
G019's provisional design stays accepted but implementation awaits G018 reviewed
integration/current-target reassessment, bindings and recorded release. G058's
planning trigger remains the first real G019 implementation MR. Stop before any
G020 planning and the G021 foundation barrier remain; no owner stop is waived.
