> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G071 source review — Codex

Date: 2026-10-27 UTC  
Disposition: **ACCEPT for MR creation and hosted pre-merge qualification**  
Reviewed candidate: `origin/g071` `0b284596158ee98de5ff7995ccd161cb55e7c5f4`, tree `80de7c04290824753d4e009accfccd4c7aa5904e`  
Base: staging `acc3f7d414e17f4c9c26aba0fad9d2c124aca411` (MR !111 merged)  
Plan: accepted r4 plus owner-directed r4.2 validation sequence; Work Spec unchanged.

## Review

The implementation stays within G071's accepted boundary. Core resolves the configured existing slug once at startup; the migration adds a narrowly scoped auth-reader function with an explicit search path, limited `EXECUTE` grants, and strict pre/postconditions. Database failures do not become an unknown-tenant fallback. The web gate aliases only the exact platform root, keeps the tenant identity and signup paths on the existing authorization boundaries, and builds the temporary redirect from Core's configured `signup_origin`; it does not use request Host/Origin/query data for the destination.

The AC8 staging-observation change remains in the test-only helper and rejects dot segments and decoded backslashes. E-DT-5/6's GET-body allowance is acceptable as a pinned Next.js adapter behavior: G071 constructs a bodyless `Response`, but the adapter exposes the exact fixed `Location` text for GET. The test allows only empty or that exact URL, keeps HEAD empty, and retains the status, exact Location, no-store and no-cookie assertions; no request value is reflected. The coordinator should preserve this narrow detail in the Result. The prior focused results and failure history are reported in the candidate Result. Codex did not run tests; `git diff --check acc3f7d4..0b284596` is clean.

## Required hosted gate and remaining holds

The owner-directed sequence is no full local run for G071, and a full hosted `validate:integration` gate on the final G071 MR before merge. The selector classifies this diff as `full` because it changes foundation paths including `core/migrations/` and `infra/validation/`; `ci_controls_required` also selects `ci-controls` because `infra/*` changed. The Plan r4.2 phrase “no `ci-controls`” must be clarified to mean **no local `ci-controls` run**; it must not suppress `ci-controls` from the requested full hosted profile. This is a documentation-only clarification; it does not require restarting implementation or rerunning focused checks.

`validate:integration` is a manual, allow-failure MR job. The coordinator must play it and verify its full-profile result before enabling auto-merge or otherwise merging. G069 final ACCEPT remains a gate to G071 merge/integration and final acceptance. The Goal says this gate applies to integration; it does not prevent opening the reviewed G071 MR or starting the requested hosted pre-merge qualification. Do not arm auto-merge until both the G069 gate and the hosted qualification are clear.

The RC17 candidate is not qualified: MR !110 is closed, pipeline `2992392293` failed, `validate:checks` job `17489187669` passed, and full `validate:integration` job `17489187670` failed. G069's Result still says that pipeline was running and lists cleanup as pending; update the closeout record and verify cleanup before G069 final ACCEPT. Do not treat the failed RC17 run as a pass or reopen it as part of this review.

This ACCEPT covers the reviewed source for MR creation and qualification only. It is not hosted qualification, merge authorization, or G071 final acceptance.
