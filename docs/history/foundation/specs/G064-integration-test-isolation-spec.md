> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G064 — Integration-test isolation and per-test evidence — Work Spec

State: Supervisor-owned definition, 2026-10-21; [Goal and authority](../goals/G064-integration-test-isolation.md). Claude owns the Plan. No implementation/runtime release follows from publication alone.

## Outcome and existing seams

Independent Go integration packages must not inherit another package's database mutations. Tests that represent independent behavior must support direct selectors with their own bounded prerequisites. Deliberately ordered scenarios remain coherent, documented proofs. Existing validation must expose actual test/subtest outcomes, not merely a green suite label.

Ground the Plan in these current seams:

| Seam | Relevant current paths |
|---|---|
| Database/role coupling and provisioning | `core/internal/migrate/runner.go`, `infra/postgres/init/10-roles.sh`, `infra/postgres/provision-roles.sql` |
| Destructive migration/catalog lifecycles | `core/internal/migrate/integration_test.go` and sibling `*_integration_test.go` |
| Package-wide environments and cleanup | `core/internal/identity/integration_test.go`, `core/internal/signup/signup_integration_test.go` (`TestMain`) |
| Per-test migration/setup and shared logs | `core/internal/tenancy/integration_test.go`, `core/internal/tenancy/access/integration_test.go`, `core/internal/mcp/env_integration_test.go` |
| Ordered recovery and fault evidence | `core/internal/tenancy/access/recovery_integration_test.go`, ordering/attack/fault siblings, `infra/recovery/recovery.sh` |
| Existing invocation/reporting | `infra/validation/{run.sh,lib.sh,required-checks.tsv}`, `infra/ci/{validate.sh,lib.sh}`, `web/playwright.config.ts` |

Current identity cleanup releases fixture slugs before later `infra/identity/mcp-seed.sql` use. A package database change must reconcile that consumer explicitly, keeping the browser fixture database attributable and correctly seeded. The `36d1189b` retention repair preserves scanned files in local `--evidence-dir`; G064 must add the bounded per-test capture/plumbing it does not provide.

## R1 — Honest package and test-lifecycle inventory

Inventory all current harness PostgreSQL integration packages: `migrate`, `identity`, `tenancy`, `tenancy/access`, `mcp` and `signup`, plus any discovered executable consumer in the approved source. Map existing required IDs, selectors/subtests, setup/fixture side effects, role/catalog mutations, log/canary paths, external fixture consumers and resource ownership. State which groups become independently selectable, which already are independent, and which deliberately require a whole scenario. Do not call package isolation proof of test independence or silently omit a difficult group.

Use the existing Plan/Result inventory, not another registry. For each changed group identify reuse/extend/new tests and the concrete uncovered behavior. Any proposed scope exclusion needs a reason and supervisor disposition before Plan acceptance; implementation must deliver the reviewed inventory, not a convenient subset.

## R2 — Correct templates and private package databases

Default ordinary package runs to distinct disposable databases cloned from a correctly provisioned, migrated and verified template for the bound source. Identify migration content/current version, provisioned roles/ACLs and fixture baseline; refuse stale, incomplete or dirty templates. Prepare bounded baseline fixtures once at the appropriate lifecycle, rather than rebuilding all migrations for every ordinary case. Do not template live credentials, provider bindings or another test's residue.

Close migration/setup connections and pools before cloning; the source must have no active sessions and remain closed to test writers. Use a separate maintenance connection for lifecycle operations. Namespace templates, clones and logs by run/package; validate and quote identifiers within PostgreSQL/repository length limits. Register exact ownership before use, avoid reuse/collision with live runs, and restore/verify database-level CONNECT/TEMP/CREATE grants and settings that cloning does not reliably establish.

**Resolve role-name binding explicitly in the Plan:** `migrate.Config.Database` also prefixes every role, while a clone preserves original ownership/ACL role identities. Merely changing `DATABASE_NAME` is insufficient. Choose a bounded test-only template/role-family or connection-lifecycle arrangement that preserves actual production-equivalent roles and migration checks. Do not rewrite production role derivation, broadly reassign ownership/ACLs or grant runtime owner/CREATEDB/superuser/BYPASSRLS. Roles, memberships and role settings are cluster-global and are not copied by database cloning; isolate their mutations with private role families/clusters or retain justified serialization. Prove that runtime connections cannot reach another package's private data/database under the selected boundary.

Migration/up/down, dirty/refusal, catalog/provisioning and recovery/restore tests that intentionally change schema, roles or database state need fresh lifecycle databases or equivalent existing isolated resources at their required starting version. Do not run them on an immutable migrated template or replace their transitions with a pre-migrated success path. Document and preserve seed → restore/invalidate → verify/repair chains and required ordering.

## R3 — Individually runnable tests without erasing proofs

For independent groups, fixtures and prerequisite state must be established by the selected test, not an earlier package/test. Prefer existing bounded fixtures and scoped cleanup; use fresh clones/resets at test/group boundaries when commits, global counts, provider-binding changes or injected catalog faults make rollback insufficient. A transaction rollback alone cannot isolate services using multiple pools/connections. Close pools before database teardown, restore modified state on failure, and isolate output/canary files and other writable fixtures as well as rows.

Keep intentional sequence, concurrency, attack, lock-order, audit-atomicity, recovery and commit/fault scenarios grouped and serial where their guarantee requires it. A direct selector for such a group must establish its baseline and execute the complete proof. Record the supported selection boundary; do not promise arbitrary nested subtests are independent. Preserve required runtime-role attestations, privileged-assertion negatives, RLS/ACL/catalog checks, audit completeness/attribution and secret scans. No blanket `t.Parallel`, assertion removal, skip, weakened count or hidden rerun; no-retry and commit-fault semantics remain explicit.

## R4 — Actual per-test outcomes in retained evidence

Capture Go's existing JSON test event stream for affected integration invocations, retaining package, test/subtest name, final status, elapsed time and failure diagnostics alongside source SHA, command/selector and run/check binding. Reuse existing Playwright JSON/reporters for browser outcomes; no new browser runner or format is required. Retain raw bounded sanitized output and existing human-readable diagnostics through the established evidence path, scanners and explicit local `--evidence-dir` behavior. Do not leak DSNs, passwords, tokens, canaries or fixture payloads, or upload secret/env files.

Preserve process exit status through pipes/conversion. Package setup/build/panic failures, malformed/truncated output, zero selected tests, missing expected tests, disallowed skips and interrupted runs must remain non-green or incomplete. Keep required-check IDs and completeness authoritative: per-test output supplements the check ledger, not a substitute for missing controls. Reconcile existing text PASS parsers if invocations move to JSON; prove that successful output handling cannot hide an earlier test/process failure. Historical failures remain failures.

## R5 — Resource safety, integration and boundaries

Reuse Docker-first pinned tools, existing run inventory/cleanup and the shared two-run private DB/browser limit; template databases inside an owned run do not grant additional independently active stacks. Keep image/cache preparation serialized and shared mutations locked. Unique names alone do not isolate roles, ports, writable files, object-store fixtures or teardown. G063 qualification remains frozen and independently owned; do not switch its source to obtain G064 evidence. G020's paused unmerged migration is absent from the staging baseline and must not enter templates implicitly.

On success, test failure, setup failure and interruption, close/reap owned clients and remove only attributable clones/templates and disposable resources. Preserve scanned evidence and failure receipts under existing retention policy; never terminate unrelated sessions or drop unknown databases/roles. Record any held resource and why cleanup could not finish. Assess the bounded unmeasured harness risks in [PG-010](../PERFORMANCE.md#pg-010--g064-test-database-cloning-cost-and-connection-contention) through correctness observations; no repeated benchmark campaign. Escalate material changes through that register.

No production migration, application behavior, qualification/CI trigger change, G061 orchestration implementation or new infrastructure is authorized. Preserve the G021 barrier. Update only necessary current harness/runbook consumers; accepted historical evidence is immutable.

## Acceptance criteria and validation

| ID | Required evidence |
|---|---|
| AC1 | Reviewed complete package/group/check inventory maps every existing relevant selector and required ID to its lifecycle, isolation treatment and retained intentional dependencies; Result accounts for every row and exception. |
| AC2 | Real PostgreSQL evidence shows correct template schema/migration/role binding and private ordinary package clones; stale/invalid template and active-source-session refusal controls fail safely. A marker/fault in one package cannot affect another; actual runtime-role denial and privileged-assertion negative controls remain meaningful. |
| AC3 | Every group classified independent runs by its documented direct selector from clean prerequisites; bounded representative repeat and shuffled/reversed selections cover each changed isolation mechanism and risky fixture class. Record commands/seeds and repeat at least one formerly coupled pair in opposite order; justify reuse rather than run a combinatorial campaign. |
| AC4 | Focused unchanged/extended migration transition, recovery chain, ordering/concurrency, isolation-attack, audit-fault and no-retry/commit-fault proofs pass under their reviewed lifecycle; no required assertion/control is removed or made accidentally unreachable. |
| AC5 | Actual successful and failing Go test/subtest and existing Playwright results remain readable and retained; controlled setup/process failure, zero match, skip, truncated/interrupted evidence and parser/exit propagation cannot create PASS. Required-check completeness and privacy scans still fail closed. |
| AC6 | Focused setup/test-failure and interruption cleanup proves only owned databases/resources removed, no cross-package/run teardown and required evidence retained; inventory shows resource/lock/cap ownership with no displacement of frozen G063 work. |
| AC7 | Focused pre-merge validation and source/security/Duo/hosted fast gates clear, followed by normal authorized staging merge and one complete coordinator-owned local full qualification on the exact merged SHA, complete Result/failure ledger, runbook and cleanup receipt. No mandatory manual M1; Codex final ACCEPT remains separate. |

Claude's Plan selects the smallest sufficient executable checks with explicit AC mapping and current input/scope bindings. Use focused local reproductions/corrections before merge; after merge run the normal full profile, preserving failed/interrupted evidence and attributable reuse/targeted rerun policy. A complete run may satisfy pending G063 and G064 only on the same reviewed candidate with all required bindings and coverage; do not duplicate unchanged qualification or reinterpret G063's frozen earlier source as G064 coverage. The retention repair's owner no-full exception is not inherited. Do not expand this task into full-suite parallel scheduling or repeated whole-platform timing runs.

## Supervisor definition review — 2026-10-21

Ready for Claude planning. Source-grounded review confirms that template role/ACL binding, cluster-global state, browser seed consumers, true per-test independence and fail-closed reporter plumbing must be resolved in the Plan. Definition approval authorizes planning only; implementation follows reviewed scope/input bindings, Plan ACCEPT and the supervisor's normal execution release without another routine owner prompt.

## Bounded lifecycle clarification — Plan r0 review, 2026-10-21

For R2/AC2, serial recreation of the canonical database name from a verified, write-closed run-private template is an approved alternative to distinct simultaneously existing package database names. Existing production migrations and role checks bind role identities to that name. Each independent invocation must receive a fresh disposable lifetime with no earlier invocation's rows/schema/settings residue; no old package clients may remain connected. Shared roles remain serialized with complete relevant membership/attribute/settings residue detection. Prove runtime denial to the closed template and preserve existing role/ACL isolation negatives. This alternative does not claim separate package credentials or parallel package isolation, change production provisioning, or authorize G061 orchestration. Stale/dirty template refusal, independent selectors, evidence and cleanup guarantees remain unchanged. See the r0 review (not included); implementation is not yet released.

## Owner extension — batch-owned infrastructure, 2026-10-22

The [batch-resource reuse addendum](G064-batch-resource-reuse-spec-addendum-20261022.md)
is part of this open G064 remediation. Reuse compatible batch-owned Docker stacks
and runners while preserving fresh test state, individual test/subtest reporting
and isolated reruns. Claude prepares a bounded append-only Plan delta immediately.
Do not finish the existing repairs with a full run before adding this work;
consolidate final qualification after the combined changes. Existing AC1–AC7,
failed evidence, budgets, review gates and production boundaries remain.
