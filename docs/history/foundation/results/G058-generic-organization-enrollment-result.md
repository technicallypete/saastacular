> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G058 Result — Generic Organization enrollment and self-signup

## Status

**Initial Result (R0′ state) — no execution attempt started; execution release NOT yet recorded.** This is a
readiness/binding record submitted for Codex review (`review-ready` for the readiness correction), not an
implementation submission and not acceptance. Author: Claude (model **Fable 5.1**; effort as configured, actual
**unknown**). 2026-10-18 UTC.

| Item | Value |
|---|---|
| Checkout owner | Claude planning/Result worker, native isolated worktree `agent-<id>` (`.claude/worktrees/agent-<id>`, not renamed; AGENTS "Task-prefixed Claude worktrees" fallback); coordinator session `019EKrFpe3tJhYgzBhHthyD6` |
| Branch / HEAD at writing | `g058-generic-organization-enrollment`, created from `origin/staging` `1055c93da1e1f4f827d1984930b92d07d59303ba` (= HEAD before this commit; branch and HEAD asserted after the switch) |
| Target (baseline) | `origin/staging` `1055c93da1e1f4f827d1984930b92d07d59303ba`; delta from the readiness SHA `36f31917bb7abe0bf525969da1639726f06bb78f` = 7 management files only (PROGRESS, SUPERVISOR-HANDOFF, readiness R0, scope amendment 01 + archive, selected-design review); no `core/`, `web/`, `infra/`, `docs/` change |
| Predecessor | G019 merged `8f39d72f` (reviewed head `0cddd0df`); both verified ancestors of HEAD (`git merge-base --is-ancestor`) |
| G019 staging qualification | pipeline `#2960419227` **RUNNING** at writing (coordinator-observed; not PASS, not canceled). Not a start gate for isolated overlap; **final G019 ACCEPT gates G058 integration/final ACCEPT** |
| Budget | provisional **B7, checkpoint after attempt 4**; attempts used **0**; forecast **4** (Plan r3 §11.1, addendum §A.7, correction-01 §C.7, readiness correction 01 §3 unchanged) |
| Model/effort for execution | Opus 5.5 high for implementation/routine test execution (CLAUDE.md); Fable 5.1 high for test corrections; actual values to be recorded at dispatch |

## Bound inputs (SHA-256 verified on `1055c93d`)

| Input | Path | SHA-256 |
|---|---|---|
| Goal | `management/v0/goals/G058-generic-organization-enrollment.md` | `7e71b6b7d837a8d625bda5cc94beaf4752cd35c71b9465462988843f7b934c62` |
| Spec | `management/v0/specs/G058-generic-organization-enrollment-spec.md` | `201fffe7818e2d5f24cda9f5628755ad739d12e484717b72deb9674518daf7a7` |
| Plan r3 (immutable) | `management/v0/plans/G058-generic-organization-enrollment-plan-r3.md` | `1753b0fe624a9d763a88187b8e1f2a28a3c1ec7b29aa9e45841effe2c5b6ca8d` |
| Selected-design addendum r0 (immutable) | `management/v0/plans/G058-selected-design-addendum-r0.md` | `c251477029d90900bb97ed7b80c953834fb8b0a1ee03acc60d1bac86309c4ace` |
| Addendum correction-01 (immutable) | `management/v0/plans/G058-selected-design-addendum-correction-01.md` | `2cecc812ef74eaec91c8d16cae13f78cb661d87e8bb7df20650f0307a8736d0b` |
| Readiness R0 | `management/v0/plans/G058-readiness-r0-after-G019.md` | `cd81d09898acdf500561b443793e0e97faa95b4c07a73042ce49b902ea7d15de` |
| Codex selected-design review (incl. "Integrated-G019 readiness R0 — bounded correction") | `management/v0/results/G058-selected-design-codex-review.md` | `8481aeb641c89b67526e3145f19c2aead05089e9acc0a1fc3d19d935351c2fa2` |
| Scope amendment 01 (46 patterns; `plans/` copy byte-identical) | `management/v0/results/G058-scope-allowlist.txt` | `d69470b5e334eda44745804d746d8be9d4b69d681ffaa3377cc5ccee4b03f4a2` |
| **Readiness correction 01 (this round)** | `management/v0/plans/G058-readiness-correction-01.md` (150 lines) | `5a950784717ecc099012ac5c83496a6261f4ee89c8f0603d399cefbf7fb77755` |

Owner decisions D-4r1 / D-15(i) / D-16(c) selected (Goal/Spec 2026-10-18 sections); D-17/D-18 recorded; not re-asked.

## Implementation summary and files changed

None yet. This round changed only the two management files above (this Result and the readiness correction).
Migrations/schema: none applied (00016 bound as definite, not written). Contracts/API/MCP: none changed. Tests added:
none. No application, migration, test, infra, harness, Docker, DB, provider or network action beyond `git fetch`/`push`.

## Readiness corrections bound this round (summary; full text in readiness correction 01)

- **D-3:** identity's `attestFromEvidence` (`identity/enroll.go`) refuses zero / unverified / issuer-subject-mismatched /
  zero-time evidence **before** `AttestVerifiedAddress` (`core/internal/identity/attest.go:49`), then requires
  `VerifiedAddress.UserID() == EnrollmentHandle.UserID`; outcomes stay within the closed set
  (`recipient_proof_unavailable`, `identity_conflict`); `AdmitSelector.Address`/`admitInvitation` carry
  `identity.VerifiedAddress`; pins: `TestAccessBoundary` text pin (`access/boundary_test.go:220-263`), new
  `TestAddressEvidenceUnforgeable`, `TestVerifiedAddress` (`identity/attest_test.go:57`) reused.
- **verifiedAt:** `workos.Adapter.verify` observation time `a.cfg.Now()` after successful `AuthenticateWithCode` +
  token validation (`adapter.go:100-127`), never `claims.AuthTime` (`:134`), never a provider timestamp (none exists);
  fake path tested through the shared harness clock; adapter-level pin needs `workos/adapter_test.go`, **not in scope**
  — open reconciliation item for Codex (one-pattern amendment or accept fake-path + review).
- **D-11 (required):** `finalizeRetirement` (`identity/retire.go:67-106`) gains the same
  `UPDATE app.enrollment_sessions … revoke_reason = 'disabled' WHERE user_id = $1 AND revoked_at IS NULL` as
  `DisableUser`, via one unexported `store.go` helper shared by `revokeAll`/`DisableUser`/`finalizeRetirement`, placed
  after the sessions UPDATE and before `revocationGap`; tests: ENR-19 extension retire → re-enable → old credential
  unusable (`identity/service_test.go`), unit statement-order/rollback (`retire_test.go` with `recTx`), fault-seam
  rollback (`store_test.go` with `faultDB`), affected reuse `'^TestMembership$/^MEM-14_'`; no new API/role/grant/pool/
  schema; no count/pin change.
- Mechanical R0′ items (fixture `grace` → `hana`, line/count re-pins, migration 00016 definite, target/predecessor
  binding) listed in correction 01 §4.

## Selected early-risk slice (prove first; Plan r0 §1.2/§8 slice 1 narrowed)

**Slice 1a — first-time-invitee admission path + D-11 revocation (core, no HTTP):**
1. `00016_organization_enrollment.sql` tables/functions/policies needed by admission (`enrollment_sessions`,
   `enrollment_restrictions`, `auth_resolve_enrollment`, `enrollment_admit`, lock helper, owner INSERT policy) +
   CT-ENR-1 catalog assertions (`migrate/enrollment_integration_test.go`).
2. `identity.go` `AddressEvidence`/`EvidenceMinter`/`ExchangeResult`; `workos/adapter.go` + `identitytest/fake_workos.go`
   minting; `identity/enroll.go` callback branch (T1) + `attestFromEvidence` (D-3); `access/enrollment.go`
   `AdmitAtCallback`/`admitInvitation` (claim + `lift` in one transaction).
3. `store.go` + `retire.go` revocation paths (D-11, D-18 scoped paths).

**Minimum focused checks for slice 1a (selectors bound at implementation; Docker-first, isolated unit checks lock-free,
DB checks under `local-harness.lock`):**
- unit (no DB): `go test ./internal/identity/ -run '^(TestAddressEvidenceUnforgeable|TestFinalizeRetirementRefusesBeforeSQL|TestRetirementCheckView|TestUserStatusRejectsNonCanonicalIDsBeforeDatabaseWork)$'`
  plus the new `retire_test.go` statement-order case;
- integration (`-tags integration`, fresh disposable DB): identity `-run '^TestEnrollment'` (ENR-7/8 first-time
  invitee via the fake, D-3 refusals, ENR-19 incl. retire → re-enable, ENR-21 bootstrap), `-run '^TestDisableAndEnable$'`
  unchanged reuse; access `'^TestIsolationAttacks$/^(ISO-ATTEST|ENR)'`, `'^TestAccessOrdering$/^O-ENR-8'`
  (claim vs departure), `'^TestMembership$/^MEM-14_'` (affected reuse); migrate CT-ENR-1.
- Required final coverage shape unchanged (affected checks only; no default `full`; existing CI trigger matrix).

Later slices (HTTP, web, seeds/browser, docs/M1) follow Plan r0 §8 order; the first-time path is exercised only after
the SDK qualification below is recorded as PRESENT (it is).

## SDK qualification outcome (executed by the coordinator, offline, read-only; recorded verbatim — not performed by this worker)

- Environment: `beginify-core-check:latest` (go1.27.2, `GOMODCACHE` `/go/pkg/mod`) with Docker volume `g019-a1-gomod`
  mounted read-only at `/go/pkg/mod`, `--network none`, plain separate `docker run --rm` commands; no install, network
  or provider call; no harness refusals.
- Integrity: `/go/pkg/mod/cache/download/github.com/workos/workos-go/v10/@v/v10.5.0.ziphash` =
  `h1:eClqevZs8AihL0kQOuX1Zzi2WT+8PYrYAuXkGszQdN4=`, equal to `core/go.sum:75`; pin `core/go.mod:15` `v10.5.0`
  (both lines re-verified in-repo on `1055c93d`).
- v10.5.0 has no `pkg/usermanagement` package (only `pkg/events`); types live in the root package, imported by the
  adapter as `sdk "github.com/workos/workos-go/v10"` (`adapter.go:18`, verified in-repo). The readiness §6
  `go doc …/pkg/usermanagement.*` path was therefore wrong; inspection used the root package files.
- `user_management.go:152` `func (s *UserManagementService) AuthenticateWithCode(...) (*AuthenticateResponse, error)`;
  `user_management.go:960` `Get(ctx, id) (*User, error)`.
- `models.go:6021` `type AuthenticateResponse struct`; `models.go:6023` `User *User \`json:"user"\``.
- `models.go:1020` `type User = EmailChangeConfirmationUser`; `models.go:7085` `type EmailChangeConfirmationUser struct`;
  `models.go:7099` `Email string \`json:"email"\``; `models.go:7101` `EmailVerified bool \`json:"email_verified"\``
  (non-pointer bool: an absent field decodes to false → fail-closed).
- No verification timestamp exists: recursive case-insensitive grep for `verified_?at|VerifiedAt` over the whole
  v10.5.0 module returned no matches (exit 1). User timestamps are only `CreatedAt`, `UpdatedAt`, `LastSignInAt` (strings).
- **Verdict:** projection PRESENT on the authenticate response → Plan §1.8 qualified for the pinned SDK shape;
  `verifiedAt` = adapter observation time of this exchange (not `auth_time`, not a provider verification timestamp).
  This is SDK-shape qualification only, not live provider behaviour proof; emulator/fake evidence is fixture behaviour.

## Resource readiness (coordinator-observed; recorded, not run; cross-checked against the repository)

| Resource | Observed locally | Repository reference on `1055c93d` | Gap / note |
|---|---|---|---|
| `beginify-core-check:latest` | present, built 2 days ago | compose build target `core-check` (`infra/compose/local.yml:187-193`, `infra/docker/core.Dockerfile` target `check`); harness builds `<project>-core-check` per run (`run.sh:2394-2398`) | **must be rebuilt from the task checkout before any batch; not treated as bound** |
| `golang:1.27.2-alpine` | present | `core.Dockerfile:3`, `migrate.Dockerfile:5`, `minio.Dockerfile:6`, `validation-tools.Dockerfile:4`; `run.sh:5642` image list | ok |
| `ghcr.io/workos/emulate` (untagged local) | present, untagged | `infra/compose/identity.yml:46` `ghcr.io/workos/emulate:0.14.0@sha256:767021483ed1…` (digest-pinned) | **gap:** local untagged image may not be the pinned digest; compose resolves by digest (pull under the lock if absent) |
| `mcr.microsoft.com/playwright:v1.63.0-noble` | present | `web-browser.Dockerfile:16` `…v1.63.0-noble@sha256:eff16c30…` (digest-pinned; `lib.sh:557` pin check) | ok if the local tag matches the digest; otherwise pull |
| `pgvector/pgvector:0.8.7-pg17` | present | `local.yml:10`; `run.sh:4026` recovery drill | ok |
| `minio/minio:latest` | present | **not referenced**: MinIO is built from `infra/docker/minio.Dockerfile` (golang + `alpine:3.22` + distroless); `run.sh:727-729` forbids `:latest` | **gap:** stray local image, not a repo input; ignore (candidate for closeout cleanup) |
| `zricethezav/gitleaks:v8.30.1` | present | `local.yml:215`; `infra/ci/lib.sh:9` | ok |
| not observed locally | — | `redis:8.10.2-alpine` (`local.yml:32`), `nginx:1.30.5-alpine3.24@sha256:…` (`identity.yml:131`), `gcr.io/distroless/static-debian12:nonroot`, `alpine:3.22`, `node:24-slim`, `python:3.13.16-slim`, `ghcr.io/astral-sh/uv:0.12.23`, `docker:29.8.2-cli@…`, `buildpack-deps:trixie-scm@…`, `busybox:1.37.0` | **gap:** needed by Compose/identity/web/CI-context builds; first pull/build happens under `local-harness.lock` (dependency downloads are not lock-free) |
| Module cache | volume `g019-a1-gomod` (G019-owned) | — | G058 prepares its own cache volume under the lock before batches; read-only mounts thereafter (CLAUDE.md isolated-unit-check rule) |
| Locks | `<home>/.local/state/platform-evidence/local-harness.lock` exists | — | DB/Compose/browser batches take it; isolated unit/static checks run lock-free |
| Evidence root | `<home>/.local/state/platform-evidence/g058-a1/` **not yet created** | owner override 2026-10-18 (home `platform-evidence/<task>-<run>`) | create at attempt-1 dispatch with private permissions |
| Docker Desktop/WSL safeguard | — | `infra/ci/docker-desktop-mounts.sh` present | run before each batch |
| Q1 persisted-DB inventory | held (G019 context) | — | **not required** (fresh disposable fixtures, Plan §6.3/§6.4) |

## Validation performed this round

| Command (plain, separate) | Outcome |
|---|---|
| `git fetch origin`; `git rev-parse origin/staging` | `1055c93d…` (expected head; staging had not moved) |
| `git switch -c g058-generic-organization-enrollment 1055c93d…`; `git rev-parse --abbrev-ref HEAD`; `git rev-parse HEAD` | branch and HEAD asserted before any write |
| `sha256sum` of the nine bound inputs | all equal to the dispatch values (table above) |
| `git diff --stat 36f31917 1055c93d` | 7 management files, 0 source files |
| `git merge-base --is-ancestor 8f39d72f HEAD`; `… 0cddd0df HEAD` | both ancestors |
| read-only inspections | `retire.go`, `attest.go`, `adapter.go`, `store.go` (afterRevocation/revokeSession/revokeAll/DisableUser/EnableUser), `service.go:437-448`, `identity.go`, `retire_test.go`, `identity_test.go` (`recDB`/`recTx`), `rls_pool_integration_test.go` (`faultDB`), `logout_test.go:223`, `service_test.go:330-360`, `access/retirement.go`, `access/boundary_test.go` pins, `access/membership_integration_test.go:18-33`, compose/Dockerfile image references, `required-checks.tsv` (125 rows), `core/migrations/` (00015 highest), `docs/foundation/OPERATIONS.md` (exists), emulator/MCP seeds |
| tests / builds / Docker / DB / provider | **none executed** (planning-phase rule) |

## Acceptance-criteria checklist

AC1–AC8: **not yet evidenced** (no attempt executed). Mapping stands as Plan r3 §9 + addendum §A.8 with readiness
correction 01 additions: AC2 += ENR-19 retire → re-enable (D-11); AC3 += `attestFromEvidence` refusals and the
verifiedAt observation-time assertion (D-3).

## Architecture deviations

None planned beyond the accepted Plan package. D-11 adds one statement to an existing identity transaction (no new
authority, grant, role, pool, schema or public API). D-3 is a DTO-shape reconciliation under the review's stated
condition. Flagged for Codex visibility: Plan §5.1 lines 626/636/649 wording changes (contract lines), applied at R0′
as mechanical text under the review's "DTO reconciliation only if…" clause.

## Unresolved issues, blockers, holds, attempt history

- **Execution release NOT recorded.** Required before any application/migration/test write: Codex disposition of
  readiness correction 01 (D-3 condition, verifiedAt, D-11), supervisor release line under AGENTS "Automatic
  reassessment and execution release".
- Open reconciliation item (not a blocker): adapter-level `verifiedAt`/projection unit pin needs
  `core/internal/identity/workos/adapter_test.go`, outside the 46-pattern scope — Codex to choose a one-pattern scope
  amendment or accept the in-scope fake-path assertion plus review of the adapter lines.
- Holds: no live provider/DNS/TLS/deployment action (G057 owns live-host qualification); Q1 not required; native
  permission holds standing (plain literal-path commands only); **stop before G020** (no G020 work of any kind).
- G019: staging `#2960419227` RUNNING; G019 retains its staging repairs/cleanup in parallel; a G019 repair touching
  shared files requires rebase before the next G058 batch; final G019 ACCEPT gates G058 integration/final ACCEPT.
- Attempt history: none (attempts 0 of B7; checkpoint review after attempt 4).
- Tool refusals this round: **none**.

## Suggested follow-up (not authorization)

- Closeout cleanup candidate: stray local `minio/minio:latest` image (not a repository input).
- If Codex prefers the adapter-level pin, import `^core/internal/identity/workos/adapter_test[.]go$` as scope
  amendment 02 (47 patterns) at the same time as this correction's disposition.

---

## Attempt 1 — current state (appended 2026-10-18 UTC; the submitted snapshot above is preserved unchanged)

**Attempt 1 started** (released early-risk slice 1a). Implementation owner: Claude, model **Opus 5.5**, effort dispatched
"high" (runtime value not verifiable from inside the worker: **unknown/inherited**). Budget **B7, checkpoint after
attempt 4, forecast 4, no reset**; attempts used: **1 (in progress)**.

| Item | Value |
|---|---|
| Release | `management/v0/results/G058-selected-design-codex-review.md` §"Readiness correction01 ACCEPT and early-risk execution release", SHA-256 `f6f93a2601e50c53ce1b8cda91fbc4e50d79f8dee4c405310ce9caa20de578db` |
| Scope | amendment 02, `management/v0/results/G058-scope-allowlist.txt` (= `plans/` copy) SHA-256 `f08ac956b0654d2216a6b83e2449fa5bfa8ba9940bf738cb1892ce5df446e3ff`, 47 executable patterns |
| Immutable inputs (re-verified on the rebased branch) | Plan r3 `1753b0fe…a8d`; readiness R0 `cd81d098…15de`; readiness correction 01 `5a950784…7755`; initial Result `9ab8aaa0…0cd3` (this file before this append) |
| Worktree / agent dir | `<source-checkout>` (native isolated worker `agent-<id>`, not renamed; the only writer of this branch) |
| Branch / HEAD | `g058-generic-organization-enrollment`; fetched `origin/g058-generic-organization-enrollment` = `00a39599…` (asserted), rebased onto `origin/staging` `4de485b1f184ed5cff4cdaabb7bb69920a843348`: commit `00a39599` dropped as already upstream (byte-identical); HEAD = `4de485b1…` at the start of attempt 1; `origin/staging` verified ancestor; force-pushed with lease `00a39599 → 4de485b1` after branch/HEAD assertion |
| Target / source SHAs | target `origin/staging` `4de485b1f184ed5cff4cdaabb7bb69920a843348`; source before rebase `00a395996e265400e216ce91f2c97a1903b9e9ce` |
| Selected slice | **1a** — migration 00016 admission subset + catalog, sealed identity projection (D-3, adapter minting, observation clock), first-time claim + lift (`access.admitInvitation`), D-11 retirement revocation; core only, no HTTP routes, no UI |
| Evidence root | `<home>/.local/state/platform-evidence/g058-a1/` (mode 0700) |
| Latest round (index) | **Round 1a-5** (end of file): stale-invitation selector correction, ENR-HOST/AUD/LOST/REMOVED/INTENT + O-ENR-8 DB rows, identity/migrate/logrecord focused runs, change-impact mapping |
| Latest round (index, r6) | **Round 1a-6** (end of file): O-ENR-8 victim-attribution correction; slice 1b SQL (policy write, readiness guard, operator setter/backfill, captures) + CT-ENR-1 readiness; transfer state |
| Latest round (index, r7) | **Round 1a-7** (end of file): policy facade (Get/UpdateEnrollmentPolicy, ErrEnrollmentNotReady) + ENR-1/ENR-3 facade rows; AT-INV 42/42 + B-AUDIT pins 42/36 + AUDIT.md T34–T36 (known-red cleared); request slice not started; transfer state |
| Latest round (index, r13) | **Round 1a-13** (end of file): HTTP/identity callback slice — enrollment routes (10 pairs / 9 paths; count note), identity login/callback fields and answers, request_on_arrival T2 request kind, EnrollmentCredential, PublicEnrollment, ENR-H units, ENR-1/ENR-INTENT DB rows, B-MCP-SOURCE/logrecord parity |

### Round 1a-1 — sealed identity projection (implemented, unit-verified); rest of slice 1a STOPPED on scope

**Stop reason (scope expansion needed; reported before any 00016/D-11/access source write).** Two required mechanical
edits fall outside the 47 patterns, and every remaining slice-1a validation chain depends on them:

| # | File (not in scope) | Why 00016 breaks it | Needed change |
|---|---|---|---|
| S-1 | `core/internal/tenancy/access/integration_test.go:236` (`setupIT`) | asserts migrate output ends `current_version\t15`; with 00016 every access integration test (claim+lift ENR rows, `TestAccessOrdering` O-ENR, `TestIsolationAttacks` ENR, reuse `'^TestMembership$/^MEM-14_'`) fails in setup | pin 16 (+ comment :233-235) |
| S-2 | `core/internal/tenancy/access/invitations_integration_test.go:227-235` (MEM-10, `mem10ConsumeOrganizationInvitation`) | asserts the runtime pool is refused `invitation_claim_organization` (D-15 pending); the selected D-15(i) runtime EXECUTE grant in 00016 inverts it | runtime pool now consumes (or assert the owner-only state only through `subsetFS(15)`); comment :21-25 |

Proposed amendment 03 (2 patterns → 49): `^core/internal/tenancy/access/integration_test[.]go$`,
`^core/internal/tenancy/access/invitations_integration_test[.]go$`. All other 00016 re-pins found by the pin inventory
are in scope (migrations_test, migrate runner/integration (`reset`/`rlsReset` drop lists, lifecycle 15→16,
`metadataThroughVersion16Only`), tenancy/integration_test `setup` 16-branch, access/recovery_integration_test:92,
mcp/env_integration_test:254, identity/store_test catalog, access/boundary_test SECURITY DEFINER per-file predicate,
run.sh B-MIGRATE-SET/CATALOG/DOCS/IDENTITY-CATALOG/version greps). `attack_expected_test.go` F2-12 still yields 42501
after D-15(i) (the claim refuses an Organization context, `invitation_context`), so no oracle change is needed; its
source text names the G019 EXECUTE denial (noted for review).

**Further consolidated early-risk findings (design-level, need Codex disposition before the affected code):**
- F-1 `rls_invitation_pending` cannot be `<db>_policy_reader`-owned as planned (Plan r3 §4.3): `pr` has no grant or
  policy on `app.invitations` (00015 postcondition 3). Smallest options: (a) own it by `<db>_claim_writer` (already
  `SELECT … USING (true)` on invitations; no new privilege; cw-owned 6 → 7), or (b) +1 `pr` column grant and policy
  `invitations_policy_reader` (policies +21 → +22). Recommendation (a).
- F-2 The readiness-correction fault-seam row (store_test.go + `faultDB` + `identity.NewStore(faultDB)`) is not
  compilable: `faultDB` is in package `identity` (`rls_pool_integration_test.go`, integration tag) while store_test.go is
  `identity_test`, and `NewStore` takes `*pgxpool.Pool` (only unexported `newStore` takes `storeDB`). Options: (a) put the
  fault-seam test in `enroll_test.go` as package `identity` with the integration tag (then the two unit tests below move
  into the integration binary or another in-scope package-`identity` file), or (b) +1 pattern for a package-`identity`
  integration file (e.g. `^core/internal/identity/enroll_integration_test[.]go$`). Recommendation (b) (keeps the unit tests
  lock-free and the existing `faultDB` seam).
- F-3 Exchange return shape (deviation taken, flagged): Plan r3 §5.1 `Provider.Exchange` → `ExchangeResult` would break
  four out-of-scope `Provider` implementers/callers (`relay_test.go` hookProvider, `rls_pool_integration_test.go`
  rlsStubProvider, `workos/staging_test.go` verifiedProvider + `a.verify(ctx, resp)`, `manual_qualification_test.go`).
  Implemented instead: the `Exchange` input carries the sealed per-call `EvidenceMinter`; the adapter mints once after
  `verify`; identity reads the one minted value back (`minted()`; zero for none or more than one mint). Same trust
  properties (identity-only construction, per call, issuer/subject bound, no client proof); `VerifiedIdentity` field pin
  unchanged.
- F-4 `identity/store_test.go:69-74` `TestIdentityCatalog` table list appears stale already at version 15 (no
  `invitations`; `rls != 13`) — inference from the inventory, not executed; will be reconciled with the 00016 re-pin.
- F-5 Interim audit gap for slice 1a (design note): enrollment refusals (`enrollment_closed`, `invitation_mismatch`, …)
  need the audit catalog (reasons 58 → 62) that arrives with the request/decision slice; until then they would be
  recorded as `auth.sign_in failed identity_unavailable` by the callback's post-consume recorder unless excluded.
  Proposed: exclude enrollment codes from that recorder in 1a; record `access.denied` when the catalog lands.

**Implemented this round (in scope, lines 6/20/22/23/24):**
- `core/internal/identity/identity.go`: `Exchange.Minter`; sealed `AddressEvidence` (unexported fields, `IsZero`);
  `EvidenceMinter` (unexported sink, `newEvidenceMinter`, `Mint` once — a second mint poisons, `minted()`);
  `EnrollmentHandle`; `ErrRecipientProofUnavailable` (409).
- `core/internal/identity/workos/adapter.go`: `Exchange` = unexported `exchange` + one `in.Minter.Mint`; `exchange`
  returns the projection `resp.User.Email` / `resp.User.EmailVerified` (non-pointer; absent → false) with `observedAt =
  a.cfg.Now().UTC()` read after `verify` (token + impersonation checks) — never `claims.AuthTime`. `verify(ctx, resp)`
  signature unchanged (out-of-scope staging_test callers).
- `core/internal/identity/identitytest/fake_workos.go`: `Grant.EmailUnverified` (false) / `OmitEmailVerified` (absent);
  default stays `true` (existing fixtures unchanged).
- `core/internal/identity/enroll.go`: `exchangeWithEvidence` (fresh minter per exchange) and `attestFromEvidence`
  (zero / unverified / issuer-subject mismatch / zero time → `recipient_proof_unavailable` **before**
  `AttestVerifiedAddress`; then attestation; `UserID() != handle` → `identity_conflict`). The access boundary pin now has
  its first production `AttestVerifiedAddress` call site under identity.
- Not yet written: 00016, callback T1/T2, `access/enrollment.go`, D-11 store/retire changes, catalog/re-pins.

**Validation round 1a-1 (lock-free; `docker run --rm --network none`, ro source, ro task module cache `g058-a1-gomod`,
`GOFLAGS=-mod=readonly`, `GOPROXY=off`, tmpfs `GOCACHE`; image `golang:1.27.2-alpine`
`sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d`, pinned base, unchanged). Pre-batch:
`docker-desktop-mounts.sh` 14404/100000 (14%); `git fetch`; `origin/staging` = `4de485b1` (no move).**

| Selector | Guarantee | Outcome |
|---|---|---|
| `gofmt -l internal/identity internal/tenancy/access` | format | clean |
| `go vet ./internal/identity/... ./internal/tenancy/access/` and `-tags integration,workos_staging ./internal/identity/... ./internal/tenancy/... ./cmd/...` | compile incl. every out-of-scope `Provider` implementer | clean |
| workos `^(TestExchangeAddressProjection\|TestExchangeValidIdentity\|TestExchangeRejections)$` | **new**: real adapter projection verified/false/absent, observedAt = adapter clock ≠ auth_time, refused/impersonated exchange → no projection, zero-minter Exchange unchanged; reuse: identity field pin, rejections | 3/3 PASS |
| identity `^(TestAddressEvidenceUnforgeable\|TestAttestFromEvidenceRefusesBeforeAttestation\|TestVerifiedAddress\|TestFinalizeRetirementRefusesBeforeSQL\|TestRetirementCheckView)$` | **new**: zero/literal/poisoned minter, one-mint read-back, Mint call-site pin (adapter only, exactly once), minter/evidence construction pin; refusal order before attestation (nil store proves no attestation reached); reuse: G019 attest + retirement units | 5/5 PASS |
| access `^TestAccessBoundary$` | `AttestVerifiedAddress` call sites only under identity (now non-vacuous) | 1/1 PASS |

Resources: volume `g058-a1-gomod` (task-owned, label `g058=a1`, created and populated under
`local-harness.lock` — holder pids 56044/56045 killed after `go mod download && go mod verify` → "all modules
verified"); containers all `--rm` (none left); no image built or pulled; `g019-a1-gomod` untouched; no DB/Compose.

**Tool refusal (verbatim, not retried in that form):** one compound `docker run … sh -c "gofmt … ; go vet … ; go test …
| grep …"` was refused: "This agent is isolated in the worktree
<source-checkout>, but this command hands sh
the text gofmt -l internal/identity…, which runs go with the text ^(TestExchangeAddressProjection|TestExc… in a plain
command, so what it runs cannot be shown not to be git. Refusing to run it — a worktree-isolated agent's git operations
must target its own worktree. Run the plain command from
<source-checkout>" Disposition: the batch was
re-issued as separate plain `docker run … go …` commands (no `sh -c`, no pipes), as the guard text and the dispatch
hygiene rule require. Narrowest supported rule if the compound form is ever needed: none requested.

Attempt accounting: attempt 1 in progress (not complete); no budget change.

### Round 1a-2 — schema subset, access lift, D-11/D-18, Codex S1/S2/F1–F5 (appended 2026-10-18 UTC)

**Owner/inputs.** Writer: Claude Opus 5.5, dispatched "high" (runtime value **unknown/inherited**); replaces the round-1
writer (context rotation). Worktree `<source-checkout>`
(native, not renamed; single writer). Delegation (same worktree, disjoint file ownership, integrated/reviewed by the
writer): chunk A schema/catalog/re-pins, chunk B D-11/D-18, chunk C access admission, chunk D pins + focused DB batch.
Release/limits: codex review §"Slice1a round1 and scope amendment03" (review SHA-256 `e37c393e…0221`); scope amendment 03
(`4f9dc632…b1f`, 50 patterns; every written file matches). Branch: fetched `1ca50bcc` (asserted), rebased onto
`origin/staging` `6a8f6712` (management-only delta; no conflict; Result bytes = staging import `c5927cfd…`), staging
verified ancestor, force-pushed with lease `1ca50bcc → 940b8793`. Checkpoint `4f183916` pushed. Staging later moved to
`8ba3e3bb` (management-only: SUPERVISOR-HANDOFF, PROGRESS, a G019 review) — exact-delta exception, no rebase.

**Codex dispositions applied.**
- **S1** `access/integration_test.go:233-236` setup pin `current_version\t15` → `16` (+comment).
- **S2** `access/invitations_integration_test.go` MEM-10: obsolete runtime-EXECUTE refusal replaced by the D-15(i)
  current-runtime valid first-time claim (same consumption/intended-user/one-binding/audit-pair `check` as the owner
  path); runtime replay of the consumed invitation returns G019's `already_member` with the same binding and writes
  nothing (binding count 1, consumer/intended unchanged, audit pair unchanged — 00015 claim body :748-755 semantics);
  MEM-11 negatives untouched and re-run green. Header comments updated.
- **F1** `app.rls_invitation_pending` cw-owned, fixed SQL, boolean only, `id = p_id AND organization_id = p_org AND
  scope_type = 'organization'` + pending (unconsumed, unrevoked, unexpired), NULL → false, search_path set, PUBLIC
  revoked, EXECUTE rt+ow; uses cw's existing invitation SELECT (no new grant/policy on invitations).
- **F2** fault seam in new `identity/enroll_integration_test.go` (package identity, integration tag, existing `faultDB`,
  unexported `newStore`); `NewStore` unchanged; no production change for tests.
- **F3** recorded deviation: Exchange input carries the sealed per-call `EvidenceMinter` (sink shared by copies, mutex
  counter, second mint poisons) and `Provider.Exchange` still returns `VerifiedIdentity` — **internal-contract deviation
  replacing Plan r3 §5.1 `ExchangeResult`** (accepted by Codex). New assertions in `TestAddressEvidenceUnforgeable`:
  copied minter mint after original → both read back zero; two concurrent completed mints joined before read-back → zero.
- **F4** `identity/store_test.go:71-73` `TestIdentityCatalog` derived from actual 00016: table list + `invitations`
  (pre-existing G019 staleness) + `enrollment_sessions`, `enrollment_restrictions`; RLS count 13 → 16 (13→14 was the
  G019 staleness, +2 ours). No security assertion removed. Executed green.
- **F5** catalog lands in 1a before any callback use: reasons 58 → 62 (`enrollment_closed`, `enrollment_restricted`,
  `approval_revoked`, `invitation_mismatch`), operations 53 → 63 (ten `enrollment.*`), `audit.go` + `audit_test.go`
  parity; `audit.events_attribute()` body replaced (00015 body + the 4 reasons on the access.denied branch; Down restores
  00015). Truthful recording wired in access (below). No identity_unavailable label, no new event family.

**00016 slice-1a subset** (`core/migrations/00016_organization_enrollment.sql`, 1083 lines; later slices extend it):
precondition DO (version 15, three helper roles); `organizations.enrollment_mode` (default `invite_only`) + addendum
§A.1/§A.3 readiness columns/3 constraints/defaults + postcondition DO; pr `SELECT(id,enrollment_mode)` +
`organizations_policy_reader`; cw `SELECT(id,enrollment_mode)`, `UPDATE(enrollment_mode)` + `organizations_claim_select`,
`organizations_claim_lock` (WITH CHECK false); `login_transactions` + `enrollment_intent`, `invitation_id`,
`request_on_arrival` (rt already table-level; `return_path` also allows `/join/status`); `enrollment_sessions` (13
cols, FORCE RLS; rt SELECT/INSERT(9 issuing cols)/UPDATE(revoked_at,revoke_reason)/DELETE + `_select/_insert/_update/
_expire`; ar SELECT + `_auth_reader`; cw SELECT/UPDATE(admitted_*) + `_claim_select/_claim_update`);
`enrollment_restrictions` (**12** columns — the list in Plan r3 §4.2 has 12 though the text says 13; pinned 12;
`recorded_via` ∈ trigger/replay/backfill; `source_request_id` without FK until the request slice; rt `_select`, cw
`_claim_select/_claim_lift`, ow `_owner_insert`; no INSERT/DELETE for any session role; rt sequence grant revoked).
Policies +14 in 1a (organizations 3, sessions 7, restrictions 4). **DEFINER: 6 of the 8-function allowance** —
`rls_enrollment_mode` (pr, STABLE, rt+ow), `rls_invitation_pending` (cw, STABLE, rt+ow), `auth_resolve_enrollment`
(ar, STABLE, rt+ow), `enrollment_lock_organization` (cw, VOLATILE STRICT, **ow only**), `enrollment_admit` (cw,
VOLATILE, rt+ow; kinds `lift`/`open`, `request` → 22023 `enrollment_kind_unsupported` until the request slice),
`enrollment_restrict_on_departure` (owner trigger body, guards, AFTER DELETE trigger on role_bindings for
organization/membership rows). Owned functions after 1a: cw 3 → 6 (final 7 with `enrollment_request`), pr +1, ar +1,
owner trigger bodies +1 (final +2). D-15(i): `GRANT EXECUTE invitation_claim_organization … TO rt`. Down refuses with
rows present / enrollment-shaped audit, restores 00015 lists. Not in 1a (later slices): enrollment_requests,
`enrollment_request`, decision trigger, rt mode UPDATE/guard/captures, readiness setter/backfill.

**access (chunk C).** `tenancy/enrollment.go` (`AdmitKind`, closed `AdmitOutcome` set, `AdmitEnrollmentTx`; P0001
`enrollment_restricted` → `CodeEnrollmentRestricted`; unknown outcome fails closed; SQL constant lives in
`tenancy/store.go`). `access/enrollment.go`: `admitInvitation(ctx, ev identity.VerifiedAddress, h, id, token)` —
pre-statement refusals (zero evidence, non-canonical handle, `ev.UserID() != h.UserID`, token shape), then ONE
IDENTITY-shape transaction: G019 `ClaimInvitationTx` → claimed Organization must equal the handle's →
`enrollment_admit(digest, org, 'lift', id)`; every refusal rolls back the whole transaction (claim included); only
`admitted`/`already_member` commit. Exported: `AdmitInvitationAtCallback`, `AdmitOpen`, `EnrollmentAdmission`.
Denials recorded after rollback through G019's independent `recordDenial` (NULL Organization, op `enrollment.admit`,
attributed): `target_not_found`, `enrollment_closed`, `enrollment_restricted`; not recorded (no reason invented):
invalid input, `recipient_ineligible`, `enrollment_expired`, unavailable. `access.decision` log keeps catalogued reason
`write_conflict` for closed/restricted (events.json has no enrollment reasons) — **review item**.

**identity (chunk B; D-11/D-18).** `store.go` helper `revokeEnrollmentSessionsAll` = exactly `UPDATE
app.enrollment_sessions SET revoked_at = $2, revoke_reason = $3 WHERE user_id = $1 AND revoked_at IS NULL`, called by
`revokeAll` (logout_all), `DisableUser` and `finalizeRetirement` (`disabled`), each after the sessions UPDATE and before
the gap read (`revocationGap` stays last before COMMIT; unknown COMMIT → identity_unavailable unchanged). Bound D-18
statement (inline: the boundary pin map forbids a helper naming `app.sessions`) in `revokeSession` (logout,
provider_logout) and the `signIn` supersession branch: only rows with the revoked session's `provider_session_id`. E5:
`DELETE FROM app.enrollment_sessions` in `insertLogin` in the EMPTY shape before `initLogin`. `EnableUser` touches no
enrollment row.

**Not implemented in round 2 — STOP for Codex disposition (callback T1/T2 branch, `Service.Login` intent, `Admitter`
wiring in main.go):**
1. **Audit vocabulary gap:** the closed outcome `recipient_proof_unavailable` (attestFromEvidence refusals) has no
   approved audit reason (the approved +4 do not include it) and is not an `auth.sign_in` reason; recording it would need
   a new catalog reason or a mapping decision — F5 forbids identity_unavailable mislabel and omission. Same for
   `recipient_ineligible`/`enrollment_expired` if Codex wants them audited (currently log-only).
2. **Host recheck location:** access has no host resolver; `AdmitInvitationAtCallback` trusts the handle's
   Organization (resolved by identity when the credential was minted). Plan §1.4 places a `host_changed` recheck inside
   the admission transaction — needs either identity-side recheck in T2 or an access-side resolver (contract choice).
Everything else in 1a (schema, claim+lift, open admission, D-11/D-18, catalog) is implemented and DB-verified.

**Mechanical re-pins (standing authority).** `migrations_test.go:44/47/54` (+00016); `migrate/runner_test.go:117/127/
132`, `:307-312` (15→16); `migrate/integration_test.go` reset/rlsReset drop lists, `metadataThroughVersion16Only`,
lifecycle 05/07/08 15→16, 09 +16→15 step, relation-order query `COLLATE "C"` (identity sequence relkind S);
`access/recovery_integration_test.go:92`, `mcp/env_integration_test.go:253` 15→16; `tenancy/integration_test.go` 16→15
purge step; `access/boundary_test.go` :25 accessFiles +enrollment.go, :67 +EnrollmentAdmission, :77-78 +2 methods,
:601 00016 DEFINER predicate (6), :657 inContextTx 16→18 (admitInvitation, AdmitOpen), :786 insertLogin order
`initEmpty,admit,Exec,unavailable,Exec,unavailable,initLogin,Exec`; `tenancy/boundary_test.go` :39-41
+AdmitEnrollmentTx, :94 files 18→19, :197 statements 74→75; `logrecord/events.json` :402 + :447 `enrollment.admit`
(no-capability rule), `logrecord_test.go` :143-145 +2 accepted records; `run.sh` B-MIGRATE-SET, root-migrate +00016,
mig_step 2–7 → 00016, object counts 56/50 → 66/56, current_version 15→16 (MIGRATE-CLI, RECOVERY-DRILL, PERSIST,
DBRENAME, TOGGLE-OFF, IDENTITY-CATALOG, MEMBERSHIP-IT texts), `identity_catalog()` +ENROLLMENT_OBJECTS (10), count
56→66, ORDER BY `COLLATE "C"`; `bash -n run.sh` exit 0. Left for the G019 repair owner (coordinator): `ow>other`
detectors, ci-controls 86/117 vs 88/119, B-TENANCY-IT 35/37. Not edited: `docs/foundation/MIGRATIONS.md` v16 row
(docs slice).

**Validation round 1a-2.** Pre-batch each time: `docker-desktop-mounts.sh` 14% (exit 0); `git fetch`. Lock-free:
`docker run --rm --network none`, ro source, ro `g058-a1-gomod`, `GOFLAGS=-mod=readonly`, `GOPROXY=off`, tmpfs GOCACHE,
`golang:1.27.2-alpine`. Logs in `<home>/.local/state/platform-evidence/g058-a1/` (0700).

| Run | Selector / guarantee | Exit | Counts | Log (SHA-256) |
|---|---|---|---|---|
| unit | identity `^TestAddressEvidenceUnforgeable$` — F3 copied/concurrent minter | 0 | 1/1 | `r2-unit-f3-minter.log` `6b69395d…14ff` |
| static | `gofmt -l`, `go vet ./...`, `go vet -tags integration,workos_staging ./...` | 0 | clean | `r2-gofmt*.log`, `r2-vet*.log` (empty, `e3b0c442…b855`) |
| unit | tenancy boundary/SQL/fixture pins + `TestScanAdmit`, `TestAdmitEnrollmentTxRefusesBeforeSQL`, `TestClassifyEnrollmentRaise` | 0 | 6/6 | `r2-unit-tenancy.log` `29f99655…149d` |
| unit | access `TestAccessBoundary`, `TestRLSBoundary`, `TestProductionImporters`, `TestAdmissionRefusesBeforeSQL`, `TestAdmissionOutcomeMapping` | 0 | 5/5 | `r2-unit-access.log` `9e527f36…a2a7` |
| unit | identity minter/attest refusal order, `TestFinalizeRetirementRevokesEnrollmentInOrder` (users→sessions→enrollment, gap last, rollback on each exec error, refusal before SQL), retirement/status units | 0 | 6/6 | `r2-unit-identity.log` `06c6a7f7…fffb` |
| unit | audit catalog parity (62/63), migrations set, migrate units | 0 | 22/22 | `r2-unit-audit-migrations-migrate.log` `50fa4d73…0d44` |
| DB S1 | migrate `^TestEnrollmentCatalog$` — CT-ENR-1 subset: tables/columns/FORCE, policies, grants, ledger no-INSERT/DELETE, 6 DEFINER owners/volatility/search_path/EXECUTE, cw column grants, CHECK lists, owner depth-0 INSERT/rt INSERT/direct trigger body/rt lock helper/cw mode UPDATE negatives, 15→16 legacy postcondition, up/down | 0 | 8/8 | `r2db-migrate-enrollmentcatalog.log` `0c0a29fe…31d3` |
| DB S1 | migrate `^TestMigrationLifecycle$` — version-16 lifecycle re-pins | 0 | 10/10 | `r2db-migrate-lifecycle.log` `5babf6d7…5f1f` |
| DB S2 | identity `^(TestEnrollmentRevocationFaultRollback\|TestEnrollmentPendingCredentialRevocation\|TestDisableAndEnable\|TestIdentityCatalog)$` — fault at the added UPDATE → identity_unavailable, zero partial user/session/enrollment/binding/audit writes (retire, disable); ENR-19 bound/broad scope, retire→EnableUser old digests never resolve, admitted_at kept, E5 cleanup; reuse disable/enable; F4 catalog | 0 | 4/4 (14 subtests, 0 skip) | `r2db-identity.log` `878e636f…1e2c` |
| DB S3 | access `^(TestEnrollmentAdmission)$` — ENR-7 first-time claim+lift atomic (binding, consumption, admitted_*, audit pair, 0 denials); ENR-23 stale invitation → enrollment_restricted, claim rolled back (pending, 0 bindings, unlifted, 1 denial), re-invite lifts; ENR-8 wrong token/foreign evidence/replay; open closed/admitted/restricted; ENR-11 fault at `enrollment_admit` rolls back the claim | 0 | 5/5 | `r2db-access-enrollment.log` `7235be92…1026` |
| DB S4 | access `^TestMembership$/^MEM-1[014]_` | 1 | MEM-10 FAIL (test asserted replay error; G019 body answers already_member), MEM-11 8/8, MEM-14 7/7 | `r2db-access-membership.log` `7a5a40dd…0856` |
| DB S5 | same (first fix expected AcceptConsumed — wrong) | 1 | MEM-10 FAIL | `r2db-access-membership-rerun.log` `2cc92b8f…ab08` |
| DB S6 | same, corrected assertion | 0 | 18/18 | `r2db-access-membership-rerun2.log` `f8f5ca84…d4d8` |

Each DB package ran on its own fresh stack (coordinator rule), 0 relations at start, roles non-superuser/no BYPASSRLS.
No failure fell into a G019-owned category. **Not run:** logrecord pin test (`validationtools` needs
`github.com/santhosh-tekuri/jsonschema/v6`, absent from `g058-a1-gomod`; would need a networked vtools build —
events.json parses, gofmt clean); umbrella/ops profiles (not required).

**Resources.** Lock `local-harness.lock` held for the DB batch (holders 13218/13219 killed; probe free). Source:
`git stash create` snapshots (stash stack untouched; final `8029236b`, DB `44ad5b92` + overlaid test fix) archived to
`…/g058-a1/r2db/` (0700). Project `g058a1r2-44ad5b92`, ports 3858/8858/8859/5858/6858/9858/9859; six stacks each `down
-v` (containers, 2 volumes, network removed); images `g058a1r2-44ad5b92-{core-check,minio-init,minio}:latest` removed;
no dangling; container count 15 = start; `.env`, `it.env`, tar deleted. Kept: volume `g058-a1-gomod`. No prune; no g019-*
resource touched; `minio/minio:latest` untouched.

**Tool refusals (verbatim, not retried in that form).**
1. chunk B, heredoc append to retire_test.go: "This agent is isolated in the worktree <source-checkout>, but this command is too complex to verify that it stays inside the worktree. Refusing to run it — a worktree-isolated agent's git operations must target its own worktree. Split it into plain, separate commands and run them from <source-checkout>" — edit made with the Edit tool instead.
2. chunk A, `bash -n run.sh` combined with grep: "…but this command runs bash inside a construct too complex to verify; what it reads or is handed as shell text cannot be shown not to run git. Refusing to run it…" — later `bash -n` run alone (exit 0).
3. chunk D, awk list comparison: "…but this command runs awk with a program that can execute commands or that it cannot read in a plain command, so what it runs cannot be shown not to be git. Refusing to run it…" — output compared by reading instead. Narrowest rule if wanted: read-only `awk` inline program on platform-evidence files (not requested).
Hygiene note: chunk A ran one read-only `git -C <worktree> status --short` (no effect) contrary to the rule.

**Open risks for the early-risk security review.** (a) `enrollment_admit` lift: restriction rows locked then
`created_at >= inv.created_at` RAISE; chronology relies on the Organization lock held by departures/invitation inserts
(r3 §1.4) — ENR-23 proves the sequential case only; O-ENR-8 concurrency not yet run. (b) Departure trigger fires on
every Organization membership DELETE incl. superuser fixtures; FK RESTRICT to users/organizations may break fixtures or
operations that delete users/organizations after a removal (`tenancy/integration_test.go:815/824/1397` flagged, not
exercised here). (c) rt INSERT on enrollment_sessions column-limited; insert policy requires NULL Organization context
and unadmitted/unrevoked rows (beyond plan text). (d) `events_attribute` body replaced in 1a. (e) `access.decision` log
reason `write_conflict` for closed/restricted. (f) callback stop items 1–2 above. Attempt 1 still in progress; no budget
change.

### Round 1a-3 — partial, transfer state at the context handoff (appended 2026-10-18 UTC)

**Owner/inputs.** Writer Claude Opus 5.5, dispatched "high" (actual **unknown/inherited**); worktree
`.claude/worktrees/agent-<id>` (native, single writer). Authority: Codex review §"Slice1a round2 early-risk
review and callback decisions" (review SHA-256 `12421918…0750734`); scope amendment 03 unchanged (`4f9dc632…b1f`, every
written file matches). Handoff ordered by the coordinator at ~367k context; **no restart, no budget change** (attempt 1
in progress, B7/checkpoint 4).

**Rebase.** Fetched `366caf83` (asserted), rebased onto `origin/staging` `d8a1f563` (contains G019 repair merge `ce7ace65`);
one conflict: `core/internal/identity/store_test.go` `TestIdentityCatalog` — staging side (G019 repair): + `invitations`,
RLS 14; G058 side: + `invitations` + `enrollment_restrictions` + `enrollment_sessions`, RLS 16 → **kept the G058 side**
(superset; no assertion dropped). `run.sh` auto-merged (B-TENANCY-IT stays G019's 37; G058 adds no TestTenancyModel
subtest). G058 does not edit `rls_integration_test.go` (`cw` label, `ow>cw` kept exactly), `policy_integration_test.go`
(group_owner fixture) or `validate-controls.sh` (88/119 kept). Staging verified ancestor; force-pushed with lease
`366caf83 → 8f456ea5`.

**Implemented (commit `8f456ea5`, plus this Result commit).**
- Host recheck: `identity.(*Service).CheckEnrollmentHost(ctx, q Querier, h EnrollmentHandle)` (enroll.go) — existing
  `cfg.Hosts.ResolveOrigin` on the passed transaction, exact `h.Origin`, Routing + same Organization + platform/custom kind;
  else `host_changed`; resolver error → `identity_unavailable`; nil service/Querier fail closed. `access/enrollment.go`:
  `enrollmentIdentity` assertion on `s.id`; guard is the first statement after InitContext in both admission
  transactions (before claim/admit); non-implementer → unavailable. No new resolver/constructor/pool.
- Audit: reasons 62 → **65** (`recipient_proof_unavailable`, `recipient_ineligible`, `session_invalid`); operations 63.
  `events_attribute`: `recipient_proof_unavailable` on auth.sign_in denied; enrollment reasons + reused `host_changed` on
  access.denied only for enrollment ops with NULL org/group/resource; `enrollment.admit` only with its reasons
  (target_not_found, enrollment_closed/restricted, approval_revoked, recipient_ineligible, session_invalid, host_changed)
  and an actor. Down refusal counts the 3 new reasons. `audit.go`/`audit_test.go` 65/63. `signInFailure`:
  recipient_proof_unavailable → denied, attributed only when a subject is passed.
- access semantic reasons (no write_conflict) in access.decision and access.denied; `enrollment_expired` → `session_invalid`;
  claim `recipient_ineligible` → recorded; `host_changed` recorded. `events.json` access.decision +4 reasons with a rule
  restricting them to `enrollment.admit` deny; `logrecord_test.go` +6 accept / +4 reject rows.
- Admitter: identity DTOs `InvitationRef`, `AdmitKind` (invitation/open), `AdmitSelector{Address *VerifiedAddress}`,
  `AdmitOutcome`, status consts, `Admitter`; `Service.admitter` field; `admitAtCallback` (attest first, discard client
  Address, Admitter error → unavailable). `access.(*Service).AdmitAtCallback` implements it (known refusals → status + nil
  error; others → error); boundary pin + `Service.AdmitAtCallback`.
- Tests updated/added: migrate CT-ENR-1_audit (65, piecewise hash-diff to 00015, 14 runtime shape controls — **not yet
  run**); access `TestAdmissionOutcomeMapping` (semantic reasons), new `TestEnrollmentHostGuard`, `TestAdmitAtCallbackMapping`;
  identity new `TestCheckEnrollmentHost`, `TestAdmitAtCallbackOrchestration`.

**NOT implemented — open decisions (stop items for Codex).**
1. **T1 audit operation:** Codex says T1 policy/intent denials use "planned `enrollment.start`" access.denied, but no
   `enrollment.start` exists in the 63-operation catalog (Plan r0/r3/addendum list ten ops, none is start) and "operations
   stay 63". Callback T1 branch (`store.enroll`), `Service.Login` enrollment intent and the callback dispatch are therefore
   **not written** (F5 forbids omitting the denial record). Needs: add `enrollment.start` (ops 64) or name an existing op.
2. **main.go Admitter wiring:** `cmd/server/main.go` does not import `tenancy/access` (Plan §5.1 "main already imports
   both" is false; each api package builds its own facade). Wiring needs a new importer edge (MAIN_GO_IMPORTS,
   TestProductionImporters) or composition via the later `enrollment/api` package (recommended). Not written.
3. From read-only SQL inspection (not DB-confirmed): (a) a mismatched attested address makes the G019 claim answer
   `not_found` (00015:837-841), so `recipient_ineligible` from the claim only arises for a disabled user; (b) an
   already-admitted, still-live credential answers `already_member` with a stale `admitted_binding_id` after removal
   (00016:624-626, before mode/restriction checks; departure does not revoke enrollment sessions) — no write, but the status
   is misleading (Codex point 4); fresh credentials get `enrollment_restricted`. Needs disposition.
4. Unrecorded known refusals remain: consumed replay (`already_consumed`), other G019 claim conflicts
   (issuer_authority_lost etc.) return an error from `AdmitAtCallback` (→ identity_unavailable) — review item.

**Lock-order correction (replaces the implied whole-transaction order).** Plan §1.4's order is the `enrollment_admit` body
order only. For the invitation kind the admission transaction actually takes: host read (no lock) → G019 claim:
invitations row `FOR NO KEY UPDATE` (00015:835) → users `FOR SHARE` via `rls_user_enabled_lock` (:863/:715) → subject
role_bindings `FOR SHARE ORDER BY id` (:870) → binding INSERT (:891) → body: users `FOR SHARE` → enrollment_sessions
`FOR UPDATE` → organizations `FOR SHARE` → restrictions `FOR UPDATE` → admitted_* UPDATE. Departure: authority guard
(users `FOR SHARE`) → binding DELETE → AFTER trigger `enrollment_lock_organization` `FOR NO KEY UPDATE` → ledger INSERT.
Bindings-before-Organization in both; a class-40 victim (complete rollback, no retry) remains an allowed outcome; no G019
lock contract changed.

**Validation round 1a-3 (lock-free; `docker run --rm --network none`, ro source, ro `g058-a1-gomod`, `-mod=readonly`,
`GOPROXY=off`, tmpfs GOCACHE, `golang:1.27.2-alpine`; pre-batch mounts 14%, fetch, staging `d8a1f563`).** Logs in
`<home>/.local/state/platform-evidence/g058-a1/`:

| Run | Exit | Counts | Log (SHA-256) |
|---|---|---|---|
| `gofmt -l internal cmd migrations` (first: 1 file, enrOps alignment) / after fix | 0 / 0 | 1 / clean | `r3-gofmt.log` `9cb92d33…d651` / `r3-gofmt-fix1.log` (empty) |
| `go vet ./...`; `go vet -tags integration,workos_staging ./...` | 0; 0 | clean | `r3-vet.log`, `r3-vet-integration.log` (empty, `e3b0c442…b855`) |
| identity `^(TestCheckEnrollmentHost\|TestAdmitAtCallbackOrchestration\|TestAttestFromEvidenceRefusesBeforeAttestation\|TestAddressEvidenceUnforgeable\|TestIdentityAuditRecorder\|TestClosedErrorVocabulary)$` | 0 | 6/6 | `r3-unit-identity.log` `0f8f375b…a458` |
| access `^(TestAccessBoundary\|TestRLSBoundary\|TestProductionImporters\|TestAdmissionRefusesBeforeSQL\|TestAdmissionOutcomeMapping\|TestEnrollmentHostGuard\|TestAdmitAtCallbackMapping)$` | 0 | 7/7 | `r3-unit-access.log` `37bb5dcf…494e` |
| `./internal/audit/ ./migrations/` (catalog parity 65/63, migration set) | 0 | 14/14 | `r3-unit-audit-migrations.log` `82ff63d4…01db` |

**Not run (next writer):** no DB batch this round; logrecord tests (need jsonschema/v6 added to `g058-a1-gomod` under the
lock). Resources: no lock taken, no stack/image created this round; `g058-a1-gomod` kept; round-2 leftover
`g058-a1/r2db/src` (old source export) retained as evidence.

**Exact next steps.** (1) Fetch/rebase/reconcile, mounts script. (2) Write access integration rows in
`enrollment_integration_test.go` (TestEnrollmentAdmission): ENR-HOST (spy HostResolver via
`identity.NewService(identity.NewStore(e.pool), identity.Config{Hosts: spy, Now: e.clk.Now})` + `e.service(id)`; spy reads
`current_setting('app.user_id')` on q and checks `*trackedTx`; reassigned/non-routing → host_changed, zero writes, one
denial; resolver error → unavailable, zero denials; `hookDB.capture` proves no `invitation_claim_organization(`/
`enrollment_admit(` ran), ENR-AUD (revoked credential → session_invalid; disabled user → recipient_ineligible; Admitter
status), ENR-LOST (`commitLostReply`, 1 BeginTx, read-only reconciliation, second admit already_member), ENR-REMOVED
(no restored binding, membership false, same-handle outcome logged pending decision 3b, fresh handle restricted),
O-ENR-8 claim-first + in-flight both directions (pauseTracer on `app.invitation_claim_organization(` via
`e.runtimePool(t, tracer)`; departure pause `hookDB.on("after","D",…)`; 20 s bounds; invariants: never a lifted
restriction, consumption iff admission committed, no partial writes). (3) Under the lock, fresh stacks: migrate
`^(TestEnrollmentCatalog|TestMigrationLifecycle|TestRLSCatalog|TestRLSMigration|TestRetentionCatalog|TestPolicyCatalog|TestPolicyMigration)$`;
identity `^(TestEnrollmentRevocationFaultRollback|TestEnrollmentPendingCredentialRevocation|TestDisableAndEnable|TestIdentityCatalog)$`;
access `^TestEnrollmentAdmission$` + `'^TestMembership$/^MEM-1[014]_'`; logrecord
`^(TestClassifyAccepts|TestClassifyRejects|TestCatalogueMatchesTheBrief)$`. (4) After Codex decisions 1–2: T1/Login
intent/callback dispatch and wiring. Known later-slice gaps: docs AUDIT.md catalog (B-AUDIT-CATALOG) not updated.

**Tool refusals this round:** none. Delegation: one read-only design subagent (no edits, no runs, no resources).

### Round 1a-4 — callback T1/T2 core wiring, round-3 decisions (partial; checkpoint at the context boundary)

**Owner/inputs.** Writer Claude Opus 5.5, dispatched "high" (actual **unknown/inherited**); worktree
`.claude/worktrees/agent-<id>` (native, single writer, no subagents). Authority: review §"Round3 bounded
callback decisions…" (review SHA-256 `807ac918…6a68`); scope amendment 03 unchanged (`4f9dc632…b1f`; every written file
matches lines 3/5/6/7/8/19/20/27/38/39/49/55). Branch: fetched `28ab7d7d` (asserted), rebased onto `origin/staging`
`4655cd2b` (delta since `d8a1f563` = management/governance docs only; `28ab7d7d` dropped as already upstream), staging
verified ancestor, force-pushed with lease `28ab7d7d → 5155a962`. Attempt 1 in progress; no budget change.

**Implemented (round-3 decisions).**
- D1 `enrollment.start`: operations 63 → **64** (`audit.go`, `audit_test.go` parity 65/64; 00016 up constraint and Down
  refusal list); reasons stay 65.
- Operation-level NULL scope: `events_attribute` — `enrollment.start` and `enrollment.admit` never carry Organization/
  Group/resource for any reason (incl. `target_not_found`); start only `enrollment_closed`/`invitation_mismatch` (actor
  optional: attributed only when the mapped subject exists); admit unchanged reasons + actor. CT-ENR-1_audit narrowing
  constant re-derived; +7 shape rows (3 start positives; admit `target_not_found` with an Organization, start with an
  Organization, start with `session_invalid`/`target_not_found` → 23514).
- D3 `enrollment_admit` step (2): an admitted credential answers `already_member` only while `admitted_binding_id` is
  still the subject's current Organization membership binding; gone/replaced → `enrollment_restricted`, no write, proof
  not rewritten.
- D4 `enrollmentClaimFailure`: `issuer_authority_lost` → `not_found`/`target_not_found` (recorded); other claim failures
  unchanged; consumed replay keeps `already_consumed` (no new code).
- D2 composition: new `core/internal/enrollment/api/http.go` (no routes, no `Register`): own access facade, `admitter`
  facade delegating `AdmitAtCallback`, `Bind()` → new `identity.(*Service).BindAdmitter` (nil/second binding refused;
  unbound → every admission `identity_unavailable`). `main.go` imports only `enrollmentapi`, binds before serving
  (failure stops startup). Pins: `TestProductionImporters` (+enrollment/api may import access; only cmd/server imports
  enrollment/api); `run.sh` `MAIN_GO_IMPORTS` +`core/internal/enrollment/api`, `MAIN_GO_G016_SHA` → `d183b0e3…a24c`.
- T1/T2 (identity): `LoginInput{EnrollmentIntent, Invitation, Request}` (facts without intent → `invalid_request` before
  any read) stored in `login_transactions` (`enrollment_intent`, `invitation_id`, `request_on_arrival`; return path
  `/join/status`); `consumeLogin` returns them. `CallbackInput{Invitation, InvitationToken}` (both-or-neither, G019 token
  form, refused **pre-consume**). Enrollment-intent exchange uses `exchangeWithEvidence`. `store.enroll` (enroll.go, one
  tx): initEmpty → namespace binding → E2 subject → (a) live continuity (other user → `identity_conflict`, same → ordinary
  signIn) → (b) forwarded invitation == stored else `invitation_mismatch` → (c) host on the tx → (d)
  `rls_enrollment_mode`/`rls_invitation_pending` else `enrollment_closed` → current member → ordinary signIn → (e)
  createMapping / initIdentity → (g) INSERT `enrollment_sessions` (30 min, provider_session_id only) → COMMIT. T1 policy/
  intent denials recorded once as `access.denied enrollment.start` (no auth.sign_in: `recorded` flag); T2 via the bound
  Admitter (invitation kind iff the stored invitation was pending, open iff mode open, else `pending` with credential);
  confirmed → `EnrollmentOutcome{admitted|already_member, next sign_in}` (no credential); known refusal → status +
  credential (Admitter recorded it; no auth.sign_in); proof failure/identity_conflict → auth.sign_in denied attributed to
  the handle user; other → identity_unavailable. New errors `enrollment_closed` (403), `invitation_mismatch` (409);
  callback outcome `enrollment_status`; `events.json` auth.callback.completed +4 outcomes; logrecord_test +4 accept/+1
  reject rows (**not run**).

**Validation round 1a-4 (lock-free; `docker run --rm --network none`, ro source, ro `g058-a1-gomod`,
`GOFLAGS=-mod=readonly`, `GOPROXY=off`, tmpfs GOCACHE, `golang:1.27.2-alpine`; pre-batch mounts 14448/100000 (14%), fetch,
staging `4655cd2b` unchanged).**

| Run | Exit | Counts | Log (SHA-256) |
|---|---|---|---|
| `gofmt -l internal cmd migrations` | 0 | clean | (stdout empty) |
| `go vet ./...`; `go vet -tags integration,workos_staging ./...` | 0; 0 | clean | (stdout empty) |
| identity `^(TestEnrollmentCallbackInputs\|TestCheckEnrollmentHost\|TestAdmitAtCallbackOrchestration\|TestAttestFromEvidenceRefusesBeforeAttestation\|TestAddressEvidenceUnforgeable\|TestIdentityAuditRecorder\|TestClosedErrorVocabulary\|TestMutationsAcceptJSONOnly\|TestBodyRules)$` (new: binding once/nil/second, invitation id/token form, pre-consume/pre-read refusals with 0 DB/resolver calls, start-denial shape) | 0 | 9/9 | `r4-unit-identity.log` `1d6cbef0…71ee8` |
| access `^(TestAccessBoundary\|TestRLSBoundary\|TestProductionImporters\|TestAdmissionRefusesBeforeSQL\|TestAdmissionOutcomeMapping\|TestEnrollmentHostGuard\|TestAdmitAtCallbackMapping)$` (extended: issuer_authority_lost → not_found/target_not_found; pass-through of other claim failures) | 0 | 7/7 | `r4-unit-access.log` `f1100ad0…c54c6` |
| `./internal/audit/ ./migrations/ ./internal/enrollment/...` (parity 65/64, migration set) | 0 | 14/14 | `r4-unit-audit-migrations.log` `332cf40a…c97a7` |

**Not run / pending (no claim of pass):** every DB batch (no lock taken, no stack/image built this round); logrecord
tests (jsonschema/v6 not in `g058-a1-gomod`). Resources: none created; `g058-a1-gomod` kept; no g019-* touched.

**Exact next steps (next writer).** (1) Fetch/rebase, mounts script. (2) Write rows in
`access/enrollment_integration_test.go` `TestEnrollmentAdmission` (designs): **ENR-HOST** spy `identity.HostResolver`
in `identity.NewService(identity.NewStore(e.pool), identity.Config{Hosts: spy, Now: e.clk.Now})` + `NewService(e.hook,
spyID, …)`, `e.hook.capture = true`; spy reads `current_setting('app.user_id', true)` on its `q` (= user ⇒ same admission
tx); non-routing/reassigned → host_changed, invitation pending, 0 bindings, not admitted, 1 denial, no captured
`invitation_claim_organization(`/`enrollment_admit(`; resolver error → unavailable, 0 denials; routing → admitted.
**ENR-AUD** via `e.svc.AdmitAtCallback`: revoked credential (setup UPDATE revoked_at) + open mode → status
`enrollment_expired`, denial `session_invalid`; disabled user (setup) → `recipient_ineligible`; invitation issued by a
granted admin whose binding is then deleted (setup) → status `not_found`, denial `target_not_found`; each nil error, NULL
Organization. **ENR-LOST** `e.hook.commit = commitLostReply` on one `AdmitOpen` → unavailable, `e.hook.begins == 1`;
reset; setup reads: 1 binding, proof names it, 1 `membership.granted`, 0 denials; second `AdmitOpen` → already_member same
binding, counts unchanged. **ENR-REMOVED** open-admit h → B; owner `RemoveMember` → `AdmitOpen(h)` =
enrollment_restricted, 0 bindings, `is_organization_member` false, proof still B, 1 denial; fresh invitation admits h2 →
B2 (restriction lifted); `AdmitOpen(h)` still enrollment_restricted (replaced), `AdmitOpen(h2)` already_member B2.
**O-ENR-8** (`ordering_integration_test.go`): claim-first sequential (admitted, then RemoveMember → departed newer → open
refused); departure-first = ENR-23 (cite); in-flight: (i) claim paused after its statement (pauseTracer on
`app.invitation_claim_organization(` via `e.runtimePool(t, tracer)`) while RemoveMember(U) → not_found, admission commits,
no restriction; (ii) existing member U with an `org_admin` upgrade invitation, departure paused after `D`
(`hookDB.on("after","D",…)`) holding the Organization lock while U's admission runs → claim waits, after the departure
commits the lift RAISEs → enrollment_restricted, full rollback, invitation pending; 20 s bounds; class-40 victim with full
rollback allowed. (3) Under the lock, fresh stack per package: migrate `^(TestEnrollmentCatalog|TestMigrationLifecycle)$`
(+ CT-ENR-1_audit), identity `^(TestEnrollmentRevocationFaultRollback|TestEnrollmentPendingCredentialRevocation|TestDisableAndEnable|TestIdentityCatalog)$`
plus new identity T1 integration rows (service_test: enrollment login → callback → credential row; closed → one
enrollment.start denial, 0 auth.sign_in, nothing created; mismatch; member → signIn; live other user → identity_conflict),
access `^TestEnrollmentAdmission$`, `'^TestMembership$/^MEM-1[014]_'`; add jsonschema/v6 to `g058-a1-gomod` under the lock,
then logrecord `^(TestClassifyAccepts|TestClassifyRejects|TestCatalogueMatchesTheBrief)$`.

**Remaining early-risk review items.** Callback T1 host refusal records auth.sign_in `host_changed` (existing ownership,
not enrollment.start) — confirm. `request_access` without an invitation answers `pending` with a credential and no
admission until the request slice. A disabled existing user still receives a T1 credential (T2 refuses
`recipient_ineligible`). Identity HTTP login/callback parsing of the new fields is held with HTTP composition. Tool
refusals this round: none.

### Round 1a-5 — stale-invitation selector correction and focused runtime proofs (appended 2026-10-18 UTC)

**Owner/inputs.** Writer Claude Opus 5.5, dispatched "high" (actual **unknown/inherited**); worktree
`.claude/worktrees/agent-<id>` (native, single writer, no subagents). Authority: review §"Round4 partial
callback source review and focused proof continuation" (review SHA-256 `afa6a43b…4ca2`, staging `68c56b77`); Result
input `700943e8…f6f3`; scope 50 (`4f9dc632…b1f`). Attempt 1 / B7 / checkpoint 4 unchanged.
- Branch: fetched `2f02b572` (branch + HEAD asserted), rebased onto `origin/staging` `68c56b77` (delta = 4 management
  files; the Result was byte-identical on both sides), ancestor verified, force-pushed with lease `2f02b572 → 50895614`.
- Later staging advances `bd0179e7`, `7ac4d3c0`: governance/management docs only (`git diff --stat 68c56b77
  origin/staging -- core infra web intelligence` empty); exact-delta exception before batches; rebase before the
  safekeeping push.

**Source correction (Codex round-4).** `core/internal/identity/enroll.go`: T1 keeps the mode-or-pending gate unchanged
but records `enrollResult.Invitation = StoredInvitation != ""` (explicit stored intent, pending or not). T2 kind moved to
pure `callbackSelector(r, in)`: invitation kind whenever stored intent exists, else open on mode open, else none. A stale
(expired/revoked/consumed) invitation on an open Organization is now adjudicated by the claim and answers its existing
closed refusal (`not_found`) with the enrollment credential, never an open admission. No new error, no policy change.

**Tests added/extended.**
- Unit `TestAdmitAtCallbackOrchestration` (extended table): stale expired/revoked/consumed on open → invitation kind with
  intent forwarded; stale on request_access, pending on invite_only → invitation; no invitation + open → open (positive);
  no invitation + request_access → none.
- `access/enrollment_integration_test.go` `TestEnrollmentAdmission` rows (round-4 designs):
  - **ENR-HOST** spy resolver reads `app.user_id` on the handed querier (= the admitting user ⇒ admission tx);
    non-routing/reassigned → host_changed, pending, 0 bindings, not admitted, no captured claim/admit statement, 1 denial
    each; resolver error → unavailable, no denial; routing → admitted.
  - **ENR-AUD** via `AdmitAtCallback`: revoked credential (open) → `enrollment_expired`/`session_invalid`; disabled user
    (`DisableUser`, then credential) → `recipient_ineligible`, credential never bootstraps; issuer removed by G019
    `RemoveMember` → `not_found`/`target_not_found`, pending; every enrollment.admit denial has no Org/Group/resource.
  - **ENR-LOST** `commitLostReply` → unavailable, 1 transaction; committed effect 1 binding = proof, 1
    membership.granted, 0 denials; replay → already_member same binding, no writes.
  - **ENR-REMOVED** admitted → removed → `enrollment_restricted` (not member, proof unchanged, 1 denial); re-invite
    admits new binding (lifted); old credential still restricted (replaced); new credential already_member.
  - **ENR-INTENT** (new; covers the correction, disabled subject and the T1 Login/callback identity rows) end to end:
    `e.id.BindAdmitter(e.svc)`, enrollment Login + Callback over the fake provider. invite_only no invitation →
    `enrollment_closed`, +1 unattributed enrollment.start denial, 0 auth.sign_in, no principal; stored X vs forwarded Y →
    `invitation_mismatch`, 1 attributed start denial; open Org + stale expired / revoked / consumed-by-another
    invitation → `enrollment_status not_found` + credential, 0 bindings, credential not admitted, 1 `target_not_found`;
    open no invitation → `admitted`, next `sign_in`, no credential, 1 binding; same member → ordinary `signed_in`;
    another principal with that live session → `identity_conflict`; disabled existing user → `recipient_ineligible` +
    credential that does NOT resolve via `auth_resolve_enrollment`, 0 bindings, 0 sessions, 1 denial.
  - **O-ENR-8** (placed in `TestEnrollmentAdmission`, not `TestAccessOrdering`, so B-POLICY-ORDERING's 37-subtest pin
    is untouched; uses that file's barriers): claim_first sequential → departure → open refused (restricted);
    departure_first = ENR-23 (cited); claim_in_flight (tracer pause after `app.invitation_claim_organization(`) →
    observed seam `removal_done`, removal `not_found`, admission committed, no restriction (the serialized alternative
    is also accepted); departure_in_flight (pause after `D`) → observed `claim_blocked`, admission `conflict`
    enrollment_restricted, full rollback (pending, 0 bindings, not admitted).

**Validation round 1a-5.** Lock-free: `docker run --rm --network none`, ro source, ro `g058-a1-gomod`,
`GOFLAGS=-mod=readonly`, `GOPROXY=off`, tmpfs GOCACHE, `golang:1.27.2-alpine`. DB: pre-batch
`docker-desktop-mounts.sh` 14456/100000 (14%); `local-harness.lock` held (flock holder killed after; probe free); one
fresh `pgvector/pgvector:0.8.7-pg17` (tmpfs data, repo `infra/postgres/init` ro) per package on internal network
`g058a1r5-net`; test container = same golang image, ro source/cache, env file 0600 (deleted after).

| Run | Exit | Counts | Log (SHA-256) |
|---|---|---|---|
| identity units (round-4 9-test selector; affected by the source change) | 0 | 9/9 | `r5-unit-identity.log` `20840250…9ccdc` |
| `go vet -tags integration,workos_staging ./internal/tenancy/access/ ./internal/identity/`; `gofmt -l internal cmd migrations` | 0; 0 | clean (after one alignment fix in `enroll.go`, whitespace only) | stdout empty |
| DB-A access `^TestEnrollmentAdmission$` (fresh stack) | 0 | 11/11 rows (+3 O-ENR-8 subtests) | `r5db-access-enrollment.log` `346d1f2a…7c9e1` |
| DB-A rerun same DB `^TestEnrollmentAdmission$/^ENR-INTENT` (see failure 1) | 1 | 0/1 | `r5db-access-intent-negcontrol.log` `89dc2cc9…738bb` |
| DB-A rerun after fix, same non-fresh DB | 0 | 1/1 | `r5db-access-intent-rerun.log` `f55552b3…40b31` |
| DB-A access `'^TestMembership$/^MEM-1[014]_'` | 0 | 18 PASS lines (MEM-10, MEM-11 8/8, MEM-14 7/7) | `r5db-access-membership.log` `8b8807ab…7097` |
| DB-B identity `^(TestEnrollmentRevocationFaultRollback\|TestEnrollmentPendingCredentialRevocation\|TestDisableAndEnable\|TestIdentityCatalog\|TestLoginStart\|TestCallbackParametersAndTransactions\|TestCallbackContinuity\|TestCallbackRefusals\|TestNamespaceBinding)$` (fresh stack) | 0 | 9/9 (23 PASS lines, 0 skip) | `r5db-identity.log` `9ce6a907…f4a8bb` |
| DB-C migrate `^(TestEnrollmentCatalog\|TestMigrationLifecycle)$` (fresh stack; incl. CT-ENR-1_audit, CT-ENR-1_updown) | 0 | 8/8 + 10/10 | `r5db-migrate.log` `c67ee719…8ab8ce4b6` |
| logrecord `^(TestClassifyAccepts\|TestClassifyRejects\|TestCatalogueMatchesTheBrief)$` first try | 1 | setup failed (x/text absent, ro cache) | `r5-unit-logrecord.log` `642cbe76…54fe2` |
| logrecord same selector after cache prep | 0 | 3/3 | `r5-unit-logrecord-2.log` `df388aae…60062` |

Full log hashes: identity `20840250166b9b8383275753da4caccb4d6f48ec1b8dba9a6d158ab967d9ccdc`; access-enrollment
`346d1f2a8fd639b9e8a3bf4f5d05f421f50491d51b613bcac42fe208a9e7c9e1`; intent-negcontrol
`89dc2cc92e006cebcbef2d3a40491e6542969876450e373ff3219944898738bb`; intent-rerun
`f55552b3f4763b192601b2146c4836843a2f7abfd8b9e72d886376ab0e540b31`; membership
`8b8807abb483898666bd1c92775ab15b7e1235c80b0342f03dcf0ce1233f7097`; identity-DB
`9ce6a907efb4593b4766a3b4ab7694f7b61d3791e1b390505f9a9cb206f4a8bb`; migrate
`c67ee719ac19cf3c0ecd3524bcb21892474bd7180b2b6bf56c513150ab8ce4b6`; logrecord
`642cbe76f3b273c01975cdaa32ede6a9eaff6d1649ea5464d9f70ec42c454fe2` / `df388aae2713c74d7aa0caeb8f64dbfd89c2b13eff1160d15a927b2a2c660062`.

**Failures and fixes.**
1. ENR-INTENT closed case counted unattributed `enrollment.start enrollment_closed` denials globally; on a reused DB it
   saw 2. Fixed to a before/after delta; rerun on the same non-fresh DB passed (also proves rerun stability).
2. logrecord setup: jsonschema/v6 v6.0.3 imports its go.sum-pinned requirement `golang.org/x/text v0.14.0`. Under the
   lock, `go mod download` of `github.com/santhosh-tekuri/jsonschema/v6` and then `golang.org/x/text` into
   `g058-a1-gomod` (proxy, go.sum-verified; no go.mod/go.sum change). x/text is beyond the literal "jsonschema only"
   wording; recorded as the minimal closure needed — **supervisor disposition requested**.

**Negative control not executed (tool refusal).** A temporary revert of the selector fix to show ENR-INTENT fails on the
old behaviour was refused by the auto-mode classifier ("[Security Weaken]"); not retried or routed around. The source
was unchanged when the next run executed (that run is failure 1 above, not a negative control). The unit table pins
the selector; the DB rows pin the end-to-end outcome.

**Change-impact mapping (owner rule, staging `7b3564f1`).**
- Stored-invitation selector → only caller `enrollCallback` (grep): unit table + ENR-INTENT. No shared auth/authz or
  policy fixture edited (new helpers live in `enrollment_integration_test.go` only; `memExpire`/`RevokeInvitation`/
  `IssueInvitation` used unchanged). Membership rows still pass on the same package build (MEM-10/11/14 18/18).
- Ordinary login/callback (round-4 `login_transactions` columns, `consumeLogin`) → identity
  TestLoginStart/CallbackParametersAndTransactions/Continuity/Refusals/NamespaceBinding 9/9 above.
- Migration paths: 00016 up/down, operation 64 / reason 65 constraints → CT-ENR-1_audit + CT-ENR-1_updown +
  TestMigrationLifecycle above; no other migrate test pins the audit constraint (grep). Historical persist/catalog
  consumers (TestTenancyCatalog etc.): 00016 not changed this round; reuse round-2 hosted/local evidence.
- Catalog/count pins: audit parity 65/64 (round-4 unit 14/14, unchanged source); TestIdentityCatalog 1/1 above;
  `TestAccessOrdering` 37 pin untouched (O-ENR-8 placed elsewhere); RLS table count / run.sh subtest pins: no
  table/run.sh change this round; `MAIN_GO_IMPORTS` / `MAIN_GO_G016_SHA`: `core/cmd/server/main.go` SHA-256
  `d183b0e3…a24c` = pin (unchanged); run.sh `7562e523…75ae` unchanged.
- Startup variants: unbound admitter → identity_unavailable (round-4 unit, reused); bound admitter → ENR-INTENT.
- Concurrency/timing: unknown COMMIT → ENR-LOST; claim/departure interleavings → O-ENR-8.
- **Hosted staging only covers what the harness selects:** `TestEnrollmentAdmission` and `TestEnrollmentCatalog` have no
  run.sh/required-checks selector yet (grep), so these rows are local evidence only until the planned harness mapping;
  hosted staging will cover the unchanged G019/foundation suites (B-POLICY-*, B-MEM-*, identity/migrate batteries).

**Resources.** Created and removed: network `g058a1r5-net`, three `g058a1r5-pg` containers, test containers (`--rm`),
two cache-prep containers (`--rm`); no images built; no new dangling volume (newest dangling = `g058-a1-gomod`); secret
env dir deleted; evidence `…/g058-a1/r5*` + `r5db/{in,out}` kept. Kept: `g058-a1-gomod` (now + jsonschema/v6, x/text).
No g019-* resource touched.

**Remaining / open early-risk items.** HTTP/UI held (no HTTP parsing/composition). Hosted harness mapping for the
enrollment rows (see above). Confirm x/text cache prep. O-ENR-8 placement deviation (TestEnrollmentAdmission). Cycle:
round-4 CORRECTION LOOP → this round (correction cycle 1 for the selector issue).

### Round 1a-6 — O-ENR-8 attribution correction and slice 1b policy/readiness SQL (appended 2026-10-18 UTC)

**Owner/inputs.** Writer Claude Opus 5.5, dispatched "high" (actual **unknown/inherited**); worktree
`.claude/worktrees/agent-<id>` (native, single writer, no subagents). Authority: review §"Round5 early-risk
evidence and bounded core continuation" (review SHA-256 `b3687f98…dad7c`, staging `8abf2d2c`); Result input
`76262fe0…2f26`; scope 50. Attempt 1 / B7 / checkpoint 4 unchanged.
- Branch: fetched `afdc8d25` (branch + HEAD asserted). Rebase 1 onto `50350349` (delta `8abf2d2c..50350349` = 3
  management files), ancestor verified. Before the DB batch staging advanced to `8b833147` (delta incl. the G019 repair
  `run.sh` pins: B-RLS-IT 10, TestIsolationAttacks 22/142, B-AUDIT-CATALOG reads 00015 with 39/33, B-AUDIT-IT AUDIT-INV
  39); rebase 2 onto `8b833147`, no conflict; `MAIN_GO_G016_SHA` `d183b0e3…a24c` = `main.go` SHA-256, `MAIN_GO_IMPORTS`
  unchanged; ancestor verified. Final source/target SHAs in the handback.

**1. O-ENR-8 correction (test only, `access/enrollment_integration_test.go`).** departure_in_flight: (a) the seam is now
observed in PostgreSQL with the existing `e.waitWaiters(t, 1)` (an ungranted `pg_locks` row of another backend of this
database) while the admission has not returned, instead of a 2 s timer; truthful statement: this observes *a* lock wait
in the database during the paused departure, not a pid-attributed wait of the admission backend; the earlier timed
non-completion alone was not lock-wait evidence. (b) the admission's own `access.decision` record is read with the
existing `memOpRecords(from, enrollment.admit)` (exactly one); `lastReason` is not used because the concurrent departure
writes to the same log. The restricted branch requires `conflict/enrollment_restricted` **and** reason
`enrollment_restricted`; the class-40 branch requires `unavailable` **and** reason `aborted` (failureReason: SQLSTATE
class 40 or a COMMIT answered with ROLLBACK); `store_unavailable`, `canceled` and `commit_outcome_unknown` cannot satisfy
it. Pending-invitation, zero-binding and unadmitted assertions unchanged. Observed: `admission conflict`, reason
`enrollment_restricted`, seam `claim_lock_wait`.

**2. Slice 1b (Plan r3 §1.5/§4.4/§4.5, addendum §A.1/§A.2/§A.5, correction 01 §C.1/§C.2/§C.5) — SQL + catalog only.**
- `00016`: rt `UPDATE (enrollment_mode)` + `organizations_enrollment_update` (rt/ow, `id = O AND MGR`); owner-owned
  invoker `app.organizations_enrollment_guard` (BEFORE UPDATE OF mode + both readiness columns): readiness change only
  when superuser **and** marker ∈ {`enrollment_readiness_set`, `replay-restrict`}; a non-superuser mode change only by
  `org_owner` (superuser fixtures/seed not blocked; the CHECK stays the single open-requires-ready rule).
  `audit.capture_organization_enrollment` (`audit_capture_enrollment`: `enrollment.policy_changed`,
  `enrollment.readiness_changed` reason = basis | `history_unresolved`) and `audit.capture_enrollment_restrictions`
  (`audit_capture_backfill`, WHEN `recorded_via='backfill'`: `enrollment.restriction_backfilled`). Owner-owned SECURITY
  INVOKER `app.enrollment_readiness_set(uuid,text,text)` (resolved|closed|unchanged|absent|no_backfill; 42501 unless
  superuser; 22004 NULL org/ref; 22023 basis/grammar, `l<seq>.` required for `operator_backfilled`; Organization
  `FOR NO KEY UPDATE`; close flips open → invite_only in the same UPDATE) and `app.enrollment_history_backfill(uuid,uuid,
  text)` (inserted|present|member|absent|not_legacy; `l<seq>.` grammar; an insert closes readiness — C.2); both PUBLIC
  revoked, no grantee. Catalog: actions 39 → **42** (trigger-owned 33 → 36), reasons 65 → **68**; `events_attribute` +3
  trigger-owned actions and one shape branch (Organization as resource; readiness/backfill actor `unknown`). Down
  refuses on the new actions/reasons and drops everything; postconditions 3b/4 extended. DEFINER inventory unchanged (8
  allowance; 6 in 1a), cw-owned unchanged.
- Pin deviation (mechanical, recorded): reasons are **68**, not the Plan's 65 — 1a already used 65 with a different
  composition (Codex round-2 +3); the three readiness reasons add on top. Policy delta on `organizations`: +4.
- `audit.go`/`audit_test.go`: +3 trigger actions, +3 reasons; parity now reads 00016 for actions/trigger list.
- `migrate/enrollment_integration_test.go`: policies (+`organizations_enrollment_update`, count 4), privileges (rt mode
  UPDATE only), invoker functions owner/no-grantee/body-hash + trigger defs, audit counts 68/64/42 + delta strip pieces
  (`enrReadinessShape`), two trigger-owned direct-append negatives, cw WITH CHECK(false) now proved with a same-value
  write (the guard refuses a mode change first), new **CT-ENR-1_readiness** (ENR-3 SQL: no_backfill, resolved + audit
  row actor unknown / request_id, unchanged, open, close → two rows, absent; guard: superuser-without-marker / owner /
  rt readiness 42501, open-unresolved 23514 for rt owner, owner role and superuser, rt member mode 42501, rt owner
  request_access + `policy_changed`; EXECUTE rt/cw/pr/ar/owner 42501; 22004/22023 cases; ENR-12/13 backfill outcomes
  absent×2/not_legacy/member/inserted (+ live close, 3 audit rows with `l7.` ref)/present, re-attest; ENR-22 owner/rt
  backfill INSERT 42501). All rolled back. `migrate/integration_test.go`: `enrollmentDrops` +6, version-16 function and
  audit-object pins.

**Validation r6.** Lock-free as round 5. DB: mounts 14476 then 14488/100000 (14%); lock held (holders killed, probe
free); fresh `pgvector/pgvector:0.8.7-pg17` per run on internal `g058a1r6-net`; env file 0600 (deleted).

| Run | Exit | Counts | Log (SHA-256) |
|---|---|---|---|
| `go vet -tags integration,workos_staging ./internal/tenancy/access/`; gofmt | 0; 0 | clean | — |
| DB access `^TestEnrollmentAdmission$/^O-ENR-8` (fresh) | 0 | 1 row, 3/3 subtests | `r6db-access-oenr8.log` `4ccef273…94cbe9` |
| audit + migrations units (pre-rebase-2) | 0 | 14/14 | `r6-unit-audit-migrations.log` `829cb0c8…dcdf889` |
| DB migrate `^(TestEnrollmentCatalog\|TestMigrationLifecycle)$` run 1 | 1 | up failed 42601 | `r6db-migrate.log` `059e161a…595ac0` |
| run 2 | 1 | up failed (postcondition 3b text) | `r6db-migrate-2.log` `92a006cf…a2e6` |
| run 3 | 1 | tables/definers/readiness + lifecycle 05-09 | `r6db-migrate-3.log` `90886b6d…b615` |
| run 4 | 1 | tables (fn order) + lifecycle 05 | `r6db-migrate-4.log` `04e60a0a…15b4` |
| run 5 (fresh) | 0 | 9/9 + 10/10 | `r6db-migrate-5.log` `4a07524c…569` |
| audit + migrations units (final source) | 0 | 14/14 | `r6-unit-audit-migrations-2.log` `204d0b9f…41c` |
| `go vet -tags integration,workos_staging ./internal/migrate/ ./internal/audit/ ./migrations/`; gofmt | 0; 0 | clean | — |

Failures/fixes: (1) plpgsql `IF CASE … THEN … END THEN` cut at the inner THEN → parenthesized; (2) postcondition 3b
regprocedure text is search_path-dependent → namespace.proname; (3) format `%s` of booleans is `t/f`; reset lacked
the new objects (lifecycle 05 version-6 allowlist) → `enrollmentDrops`; version-16 function pin order. All fixed and
rerun on fresh stacks.

**Change-impact mapping.** O-ENR-8 change → only that row (rerun). 00016/audit catalog → CT-ENR-1 (all 9 subtests incl.
updown), TestMigrationLifecycle 10/10, audit parity 14/14. Not rerun (source unchanged since round 5, attributable):
access enrollment rows other than O-ENR-8, MEM-10/11/14, identity DB rows, logrecord. **Known red until the next
slice step (not yet run, will fail as-is):** access `TestAudit` AT-INV (3 new trigger-owned actions never observed;
B-AUDIT-IT pin `{"actions":39,"observed":39}`), and B-AUDIT-CATALOG (reads 00015, 39/33; AUDIT.md T-rows) — both need
the facade producer + operator calls in AT-INV, run.sh pins 42/36 reading 00016, and AUDIT.md rows; harness/docs in
scope. Hosted-only: everything harness-selected; no enrollment selector exists yet (B-ENROLLMENT-CATALOG/IT pending).

**Resources.** Created/removed: `g058a1r6-net`, five `g058a1r6-pg`, `--rm` test containers; env dir deleted. Kept:
`g058-a1-gomod`. No g019-* touched. **Tool refusals:** two worktree-guard refusals, not retried in the same form: a
`sed -i` with `r` file insertion ("…runs sed with the program 418r … cannot be shown not to be git") → done with Edit;
a `docker run … ; echo exit=$?` compound ("…construct too complex to verify") → rerun as a plain command.

**Transfer state (context checkpoint ~300k; next writer).** Remaining slice-1 core before any HTTP/UI, in order:
(a) Go policy facade: `tenancy/enrollment.go` policy read/update store functions; `access` `GetEnrollmentPolicy`
(`organization.read`; `{mode, open_ready, history_basis}`) and `UpdateEnrollmentPolicy` (`organization.manage_policy`,
`ExclusiveOrganization`, read readiness → `ErrEnrollmentNotReady` 409 `enrollment_not_ready`, no UPDATE/audit; 23514
on `organizations_open_requires_resolved_history` maps the same), `access/errors.go`/`log.go`, unit + DB rows (ENR-1,
ENR-3 facade part); (b) AT-INV extension (`UpdateEnrollmentPolicy` → policy_changed; operator `enrollment_readiness_set`
/ `enrollment_history_backfill` on a non-member subject → readiness_changed / restriction_backfilled), run.sh
B-AUDIT-IT 42/42 and B-AUDIT-CATALOG → 00016 42/36 + AUDIT.md T34–T36 rows; (c) request slice: `enrollment_requests`
table, `enrollment_request` DEFINER, `enrollment_restrict_on_decision`, `enrollment_requests_immutable` (authority_seq
derived under the lock), `events_insert_enrollment`, request/decision/revocation captures and catalog (+3 actions,
resource type `enrollment_request`), `enrollment_admit` kind `request`, facades (Request/Decide/Revoke/List/Status/
Abandon), ENR/O-ENR rows; (d) `replay-restrict.sql` materialization + §C.3 close and ENR-16; (e)
B-ENROLLMENT-CATALOG/IT run.sh selectors + required-checks mapping (disposition 3). Cycle: round-5 CORRECTION LOOP →
this round (correction cycle 1 for O-ENR-8).

### Round 1a-7 — policy facade, AT-INV/audit pins (known-red cleared); request slice not started (appended 2026-10-18 UTC)

**Owner/inputs.** Fresh writer Claude Opus 5.5, dispatched "high" (actual **unknown/inherited**); worktree
`.claude/worktrees/agent-<id>` (native, single writer, no subagents). Authority: review §"Round6 partial
policy/readiness source ACCEPT" (review SHA-256 `d4955353…3850`, staging `ef4eb278`); Result input `45b8a851…7451`;
scope `4f9dc632…b1f` (50 patterns). Attempt 1 / B7 / checkpoint 4 unchanged.
- Branch: fetched `d7cb93a9` (branch + HEAD asserted). `git diff --stat 8b833147 origin/staging -- core infra web
  intelligence` empty. Rebase 1 onto `a608145d`, ancestor verified, force-push with lease `d7cb93a9 → d96610d1`. Before
  the first batch staging advanced to `d6599e9b` (delta `a608145d..d6599e9b` = 9 management/skills files: G020/G055
  goal/spec, PROGRESS/ROADMAP, handoff, coordination ref — irrelevant to tested inputs, exact-delta exception); the
  checkpoint rebase landed on `7df1cd1a` (staging had advanced again: `d6599e9b..7df1cd1a` = G055 result/review/scope,
  G020 goal/spec, PROGRESS, handoff and `docs/ARCHITECTURE.md` — none a tested input of this round), merge-base verified
  = `7df1cd1a`, force-push with lease `d96610d1 → 592d4e80`. Tests ran on the `a608145d`-based tree; the whole delta
  `a608145d..7df1cd1a` touches no core/infra/web/intelligence path. Final SHAs in the handback.

**1. Policy facade (Plan r3 §1.5; addendum §A.5/§A.6).**
- `tenancy/enrollment.go`: `EnrollmentMode` (closed set, `Valid`), `ParseEnrollmentMode` (unknown → invite_only),
  `EnrollmentPolicy{Mode, OpenReady, HistoryBasis}`, `ParseEnrollmentPolicy` (ready only with both readiness columns
  set; half-set reads unresolved), `ReadEnrollmentPolicyTx`, `SetEnrollmentModeTx` (absolute set; zero org → invalid_id,
  bad mode → `invalid_enrollment_mode`, both before SQL). `store.go`: `sqlReadEnrollmentPolicy`, `sqlSetEnrollmentMode`
  (UPDATE … RETURNING the 3 columns). `errors.go`: `CodeInvalidEnrollmentMode`, `CodeEnrollmentNotReady`; checkCodes
  `organizations_enrollment_mode_valid`, `organizations_open_requires_resolved_history` (23514 → not ready).
- `access/enrollment.go`: `GetEnrollmentPolicy` (`organization.read`, `organizationQuery`) and `UpdateEnrollmentPolicy`
  (`organization.manage_policy`, `policyWriteQuery` = Organization `FOR NO KEY UPDATE` with the facts; for open the
  readiness pair is read in the same transaction → `notReady()` with no UPDATE and no audit; 23514 from the CHECK maps
  to the same; other errors `writeFailure`). `errors.go`: `fieldEnrollmentNotReady`/`reasonEnrollmentNotReady`
  (`enrollment_not_ready`), exported `ErrEnrollmentNotReady` (conflict + field; HTTP 409 mapping is the later adapter),
  `(*Error).Is` (code+field+object equality; no prior `*Error` sentinel existed, so no existing `errors.Is` changes).
  Not an access.denied reason (no `denied` scope). `log.go`: `enrollment.policy.get`, `enrollment.policy.update`.
- Pins: tenancy `pinnedPackageFunctions` +4; `TestSQLBoundary` 75/19 → 77/20; access `pinnedExportedDecls` +`var
  ErrEnrollmentNotReady`, methods +`Error.Is`, `Service.GetEnrollmentPolicy`, `Service.UpdateEnrollmentPolicy`;
  `events.json` access.decision +2 operations, +1 reason, rule `enrollment_not_ready` → only `enrollment.policy.update`
  deny, capability rules (get → organization.read; update → organization.manage_policy); `logrecord_test` +4 accept,
  +4 reject rows.
- Tests: units `TestEnrollmentPolicyParse`, `TestEnrollmentPolicyTxRefusesBeforeSQL`, `TestClassifyEnrollmentPolicyCheck`
  (tenancy), `TestEnrollmentNotReadyError` (access). DB: new entry `TestEnrollment` — **ENR-1** (fresh Organization
  invite_only, ready by construction `created_after_ledger`, read by owner and member, zero audit rows) and **ENR-3**
  facade part (admin/member forbidden + one access.denied each, mode unchanged; 3 bad modes invalid_input; owner
  request_access = 1 `enrollment.policy_changed`, repeat no row, open on ready; operator close → read `{invite_only,
  false, ""}`; owner open → `errors.Is(ErrEnrollmentNotReady)`, mode unchanged, **zero** new audit rows for the
  Organization, no denial; request_access still allowed (+1 row); operator `operator_verified_absent` → open succeeds
  with basis). The direct-SQL 23514/42501 negatives of ENR-3 stay in CT-ENR-1_readiness (round 6, unchanged).

**2. AT-INV and audit docs/counts (known-red consumers, now green).** `audit_integration_test.go` AT-INV: owner
`UpdateEnrollmentPolicy(request_access)`; superuser `enrollment_readiness_set(A, NULL, 'audit-inv.close')` = closed,
`enrollment_history_backfill(A, nonmember, 'l1.audit-inv')` = inserted, `enrollment_readiness_set(A,
'operator_backfilled', 'l2.audit-inv')` = resolved; exact rows asserted (policy_changed user/owner; readiness_changed
history_unresolved unknown; restriction_backfilled operator_backfilled unknown subject nonmember; readiness_changed
operator_backfilled unknown). `run.sh`: B-AUDIT-IT pin `{"actions":42,"observed":42}` + comment/PASS prose (42; G058
contributors); B-AUDIT-CATALOG doc half reads **00016** with **42/36** + comment/PASS prose. `AUDIT.md`: T34–T36 rows,
G2 row (+7 enrollment reasons, 64 operations), operations list +11 (53 → 64), G058 paragraph (36 + 6 = 42 actions,
reasons 58 → 68, operations 53 → 64).

**Validation r7.** Lock-free units as round 5/6 (`docker run --rm --network none`, ro source, ro `g058-a1-gomod`,
`GOFLAGS=-mod=readonly`, `GOPROXY=off`, tmpfs GOCACHE, `golang:1.27.2-alpine`). DB batch: lock held (holder flock 83839 +
sleep 83840 + shell, all killed; probe free); mounts 14496/100000 (14%); internal network `g058a1r7-net`; fresh
`pgvector/pgvector:0.8.7-pg17` (tmpfs data, repo `infra/postgres/init` ro) + MinIO (tmpfs) + repo `infra/minio/init.sh`
(bucket + scoped credential); env files 0600 from a generator script (no value printed), deleted after.

| Run | Exit | Counts | Log (SHA-256) |
|---|---|---|---|
| gofmt -l tenancy/audit (first) | 1 | 3 alignment-only files | stdout |
| gofmt -l (after fix) | 0 | clean | — |
| `go vet -tags integration,workos_staging` tenancy/access/audit | 0 | clean | — |
| units tenancy + access + audit (first) | 1 | 2 FAIL (below) | `r7-unit-tenancy-access-audit.log` `ed34198d…3634205` |
| units tenancy + access + audit (after fix) | 0 | 3/3 packages ok | `r7-unit-tenancy-access-audit-2.log` `041147a7…a9d238` |
| logrecord `./internal/logrecord/` | 0 | ok | `r7-unit-logrecord.log` `43dd442b…105584e` |
| B-AUDIT-CATALOG doc-half reproducer (run.sh sed lines on AUDIT.md / 00016) | 0 | doc 42 = migration 42 (identical lists), T-rows 36 | `r7-auditdoc-{doc,mig}.txt` `2e06ada4…cefcd6b`, `-trig.txt` `b56717bf…65d86e2` |
| `bash -n infra/validation/run.sh` | 0 | — | — |
| DB access `-run '^(TestEnrollment\|TestAudit)$'` (fresh stack) | 0 | TestAudit 9/9, `AUDIT-INV {"actions":42,"observed":42}`; TestEnrollment 2/2 | `r7db-access-enr-audit.log` `6385f7d9…5e998f4` |
| logcheck validate of the run's access.decision log | 0 | 210 records, 172 access.decision, 0 failures, pass | `r7db-logcheck.json` `6bbd8596…ab273` |
| logcheck canary scan (access log + PostgreSQL log) | 0 | 3 canaries, 0 hits | `r7db-logscan.log` `f37f11e8…6ef42d` |
| `go vet ./...` and `go vet -tags integration,workos_staging ./...` (core) | 0; 0 | clean | — |

Failures/fixes (all first-run, mechanical): (1) tenancy `TestSQLBoundary`: a column-list constant counted as a non-
statement → inlined; statement pin 75/19 → 77/20. (2) access `TestRLSBoundary/forbidden_production_SQL`: the 1a pin
counted 6 app functions in 00016, but round 6 added 3 owner-owned SECURITY INVOKER functions (known-red since round 6,
not run there) → pin 9 app functions, 6 DEFINER, 3 INVOKER, all 9 names. Rerun green.

**Change-impact mapping.** Policy facade → tenancy units (boundary/SQL pins, parse/classify), access units (exported
surface, RLS boundary, not-ready error), ENR-1/ENR-3 DB rows, logrecord schema + logcheck of real records. AT-INV/pins
→ TestAudit 9/9 with AUDIT-INV 42/42 and the doc-half reproducer. Compile impact on all importers → `go vet ./...` both
tag sets. Not rerun (inputs unchanged since round 6, attributable): migrate CT-ENR-1/TestMigrationLifecycle (no SQL
change), identity rows, TestEnrollmentAdmission rows, MEM rows. **Hosted-only:** the harness-selected B-AUDIT-IT and
B-AUDIT-CATALOG themselves (full harness, pinned MinIO build), required-checks/CI-controls (unchanged). Local deviation:
MinIO server image `minio/minio:latest` (RELEASE.2025-09-07; mc RELEASE.2025-08-13 = the pin) instead of the pinned
source build (RELEASE.2025-10-15), to avoid an unattributable source build; hosted B-AUDIT-IT uses the pinned build.

**Resources.** Created/removed: `g058a1r7-net`, `g058a1r7-pg`, `g058a1r7-minio`, `--rm` test/init containers; env files
deleted. No image built or pulled. Kept: `g058-a1-gomod`. No g019-* touched. **Tool refusals:** (a) a python heredoc
editing `boundary_test.go` — auto-mode classifier "[Auto-Mode Bypass]" → done with Edit; (b) `docker run … > log 2>&1;
echo exit=$?` — worktree guard "…construct too complex to verify…" → rerun as a plain command. Neither retried in form.

**Remaining (in Plan order, before any HTTP/UI).** (c) request slice — `enrollment_requests` table (15 columns incl.
`authority_seq`, CHECKs, `one_pending`), `enrollment_request` cw DEFINER, `enrollment_restrict_on_decision` owner
DEFINER trigger body, `enrollment_requests_immutable` invoker guard (authority_seq derived under the Organization lock;
immutable after), `events_insert_enrollment` cw policy, rt/cw grants/policies (§4.4), request captures (+3 trigger-owned
actions `enrollment.requested/decided/revoked`, resource type `enrollment_request` 10 → 11), `enrollment_admit` kind
`request` (§4.3 steps 5/6′/8/9), Go Request/Decide/Revoke/List/Status/Abandon facades, ENR/O-ENR rows; it is atomic with
its pins: DEFINER 6 → 8 (access `TestRLSBoundary` names/counts, CT-ENR-1), cw-owned 6 → 7, AT-INV +3 producers via the
facades, B-AUDIT-IT 45/45, B-AUDIT-CATALOG 45/39, AUDIT.md T37–T39, audit.go TriggerActions/ResourceTypes; (d)
`replay-restrict.sql` + §C.3 close + ENR-16; (e) B-ENROLLMENT-CATALOG/IT selectors (`TestEnrollment`,
`TestEnrollmentAdmission`, `TestEnrollmentCatalog`; actual recursive subtest counts) + required-checks mapping and
CI-controls counts if they pin the list. Refused source-revert negative control: not done, not retried.

**Transfer state (safe boundary; context ≈ 160k, handed off early because the request slice is atomic with its pins
and would not fit below 300k).** Branch head = this commit on `7df1cd1a`; all r7 checks green; no known-red consumer
from slices 1a/1b remains locally (hosted harness rows pending CI). Next writer starts at (c). Cycle count: no
correction loop this round (first-run mechanical pin fixes only).

### Round 1a-8 — request-slice SQL drafted and parked; context checkpoint before Go/pins (appended 2026-10-18 UTC)

**Owner/inputs.** Fresh writer Claude Opus 5.5, dispatched "high" (actual **unknown/inherited**); worktree
`.claude/worktrees/agent-<id>` (native, single writer, no subagents). Result input `4450ddd3…0ce`; scope 50
patterns. Attempt 1 / B7 / checkpoint 4 unchanged. Branch fetched `98d2ef9d` (branch + HEAD asserted); `origin/staging`
`7df1cd1a` is an ancestor (no rebase needed). No test batch ran this round (no pre-batch reconciliation required).

**Why stopped.** Active context reached the ~300k checkpoint after reading what the request slice needs (Plan r3
§1.4/§4.2–§4.5/§11.2, addendum §A.4–§A.7, correction 01 §C.3, the whole 00016 file and CT-ENR-1). The slice is atomic with
its pins, so nothing partial was put on the task branch: the task branch is unchanged except this section.

**Parked work (local branch, not pushed).** `g058-a1-r8-request-slice-wip` = `7ebe584d` (parent `98d2ef9d`), one file,
`core/migrations/00016_organization_enrollment.sql` (+413/−51), **not executed, no pin moved**. Next writer: `git switch
g058-a1-r8-request-slice-wip` (or cherry-pick `7ebe584d` onto the task branch), review, then continue. Content:
- `app.enrollment_requests` (created before the ledger; FORCE RLS): 13 columns exactly as Plan r3 §4.2 enumerates (the
  §11.2 "15" does not match its own list — the list is bound, as round 1a did for the ledger's 12 vs "13"; **pin
  deviation to record**); FKs org/user/decided_by/revoked_by RESTRICT; CHECKs status_valid, lifetime (≤ 30 days),
  decision/revocation/consumption/authority shapes (authority `>= 0` and 0 unless approved/revoked); indexes
  `enrollment_requests_one_pending` (unique partial), `_scope_idx`, `_user_idx`. Ledger gains
  `enrollment_restrictions_source_request_fkey`.
- Policies/grants (§4.4): rt SELECT + UPDATE(status, decided_at, decided_by_user_id, revoked_at, revoked_by_user_id);
  `enrollment_requests_select` (rt/ow: own or O+MGR), `enrollment_requests_decide` (rt/ow: O+MGR, USING pending|approved,
  CHECK approved|denied|revoked); cw SELECT, INSERT(id, organization_id, user_id, status, expires_at), UPDATE(consumed_at,
  consumed_binding_id); `_claim_insert` (user = ctx_user, pending), `_claim_select` (true), `_claim_update` (true / user =
  ctx_user); `events_insert_enrollment` on audit.events (cw, depth ≥ 1, user actor = ctx_user, org set, action
  `enrollment.requested`).
- Functions: cw DEFINER `app.enrollment_request(bytea, uuid)` (steps 1–7 of §4.3; outcomes recipient_ineligible,
  enrollment_expired, not_found, enrollment_closed (unless request_access), already_member, pending (existing or new id));
  `enrollment_admit` kind `request` (mode request_access|open; request row FOR UPDATE; pending → `pending`, denied →
  `denied`, consumed → `already_member` with a current binding else `approval_revoked`; revoked/expired/decider no longer
  org_owner|org_admin → `approval_revoked`; ledger rows FOR UPDATE; unlifted `seq > authority_seq` → `enrollment_restricted`
  zero writes; INSERT binding, consumption UPDATE, lift `seq <= authority_seq` with `lifted_by_user_id` = decider,
  `approved_request`); owner DEFINER trigger body `app.enrollment_restrict_on_decision()` (TG_* guards, lock helper,
  ledger row kind = new status, creator = revoker|decider, `source_request_id`) on trigger `enrollment_requests_restrict`
  (AFTER UPDATE OF status WHEN denied|revoked and changed); owner INVOKER `app.enrollment_requests_immutable()` (BEFORE
  UPDATE; immutable keys/timestamps; pending → approved|denied once; on approved locks the Organization FOR NO KEY UPDATE
  and derives `authority_seq` = max(seq) for the subject, ignoring any supplied value; approved → revoked while
  unconsumed; consumption pair once; any other change 42501 `enrollment_request_transition`); capture
  `audit.capture_enrollment_requests()` on `audit_capture_request_insert` / `audit_capture_request_status`.
- Catalog: actions 42 → 45 (trigger list +3), resource types 10 → 11 (`events_resource_type_valid` replaced up and restored
  down), `events_attribute` branch: requested/decided/revoked = success, no reason, plain, org + subject + resource
  `enrollment_request`, actor_kind `user`, requested only by the applicant itself.
- Pre/postconditions: absent-object checks +3; DEFINER inventory 8; cw-owned **7**; policy inventory incl. the five
  request policies and `events_insert_enrollment`; invoker list +2; privilege boundary +11 negatives. Down: refuses also on
  request rows and on the three actions / `enrollment_request` resource type (message gains `/ % requests`), drops the new
  triggers/functions/policy/table, restores the version-15 resource-type constraint.

**Design notes for review (within the Plan; flag, not decided here).** (1) `enrollment.decided` carries no reason (Plan r0
§4.5 text: success, reason NULL) — approve vs deny is read from the request row, not the audit row. (2) An expired pending
request still occupies `one_pending`, so `enrollment_request` returns it as `pending` (Plan step 6 as written); the
applicant needs an administrator decision to request again. (3) The request list's display name for a non-member
applicant is not readable by rt under the existing users policies; the facade will have to return it as absent (Plan r0
`DisplayName *string`) unless a reviewed read path is added — no policy was added.

**Remaining before HTTP/UI (in order; unchanged list, (c) now SQL-drafted).** (c) review/finish the parked SQL; Go:
tenancy `AdmitKindRequest` + outcomes `pending`/`denied`/`approval_revoked`, `RequestEnrollmentTx`, decide/revoke/list/
status/abandon store functions + error codes (not_pending, not_approved, already_consumed, request not found) and
`enrollment_request_transition`; access `RequestEnrollment`, `AdmitRequest`, `DecideEnrollmentRequest`,
`RevokeEnrollmentApproval`, `ListEnrollmentRequests`, `EnrollmentStatus`, `AbandonEnrollment` + ops in `log.go`
(`enrollment.request/.status/.abandon/.requests.list/.decide/.revoke`), events.json/logrecord rows; pins: CT-ENR-1
(tables/policies/privileges/definers incl. `enrollment_request` and the decision body, cw 7, owner DEFINER 2, audit
actions 45 + resource types 11 + strip pieces, updown message), migrate `enrollmentDrops`/`enrollmentRelations`/version-16
function list + captures, access `TestRLSBoundary` (12 app functions, 8 DEFINER, 4 INVOKER), identity `store_test` FORCE
table list, tenancy integration down purge (+requests, +three actions), audit.go TriggerActions/ResourceTypes + audit_test,
AT-INV +3 producers, run.sh B-AUDIT-IT 45/45 and B-AUDIT-CATALOG 45/39, AUDIT.md T37–T39; ENR/O-ENR rows; then (d)
`replay-restrict.sql` + §C.3 close + ENR-16; (e) B-ENROLLMENT-CATALOG/IT selectors + required-checks mapping.

**Resources.** None created (no container, network, env file or lock). Kept: `g058-a1-gomod`; local branch
`g058-a1-r8-request-slice-wip`. **Tool refusals:** one worktree-guard refusal — a `git diff --name-only $(git merge-base
HEAD origin/staging) HEAD | …` compound ("…names git in a form too complex to verify…"); rerun as a plain `git diff
--name-only origin/staging HEAD -- core infra web docs intelligence`. Not retried in form. Source-revert negative control:
not done.

### Round 1a-9 (c1) — request-slice SQL, catalog/grants/pins and focused real-PG proof (appended 2026-10-18 UTC)

**Owner/inputs.** Fresh single writer Claude Opus 5.5 (dispatched high; actual **unknown/inherited**), native worktree
`.claude/worktrees/agent-<id>`, no subagents. Codex round-8 dispositions (review SHA `77774b60…a0c1`) binding;
Result input `dc8b1209…6a12`; scope 50 patterns. Attempt 1 / B7 / checkpoint 4 unchanged. Branch fetched `58695c7d`
(branch + HEAD asserted), rebased onto `origin/staging` `8ede7dbf` (the r8 Result commit dropped as already upstream).
Pre-batch fetch found staging `6a494e24` (2 commits, `management/` only: G019 review/result, PROGRESS, handoff): exact-
delta exception, no bound input touched; rebased onto it before the push.

**SQL (00016) = parked `7ebe584d` + three corrections (+19/−4 vs the draft).** Reviewed critically against Plan r3
§4.2–§4.5 and the dispositions; the draft's 13 columns (D1), reason-NULL decided event (D2), expired-pending-stays-pending
(D3) were already right; no users-RLS change (D4). Deltas:
1. `enrollment_requests_immutable`: pending → approved|denied requires `decided_by_user_id = ctx_user()`, approved →
   revoked requires `revoked_by_user_id = ctx_user()`. Without it a manager could attribute the approval to another
   (e.g. the owner) and so defeat `enrollment_admit`'s "decider still org_owner/org_admin" check and split the event actor
   from the recorded decider. **Flag for review** (a tightening inside the Plan's own invariant; no new column/policy).
2. `enrollment_request`: the INSERT is `ON CONFLICT (organization_id, user_id) WHERE status = 'pending' DO NOTHING`, then
   the committed pending row is returned (fallback `40001 enrollment_request_conflict` if none): a concurrent first
   request from a second credential of the same subject resolves to the one pending row instead of a raw 23505. **Flag.**
3. Comments: D3 (expired pending kept, never renewed/consumable, admin resolves) and the decider rule.

**Pins moved (all the SQL determines).** CT-ENR-1 (`migrate/enrollment_integration_test.go`): tables (13 request columns,
11 constraints, FKs incl. `source_request_fkey`, one_pending definition, indexes, FORCE), policies (+5 request,
`events_insert_enrollment`), privileges (rt/cw column grants, pr/ar none, table grants), definers (8 DEFINER, cw 7, owner
DEFINER 2, body md5 for `enrollment_request`/decision body/state machine/capture, the 4 request triggers, request and
admission statement order; the removed `enrollment_kind_unsupported` fragment replaced), audit (actions 45, resource types
11, strip pieces incl. the request shape branch, direct appends of the 3 actions 42501), **new `CT-ENR-1_requests`**
(below), updown (message `/ 0 requests /`, FORCE incl. requests, + a refused down with a committed request row).
`migrate/integration_test.go`: `enrollmentDrops` (+requests table, 3 functions), `rlsReset` (+`enrollment_request`),
`enrollmentRelations` 10 → 15, version-16 function inventory (+3), audit functions (+capture). access `TestRLSBoundary`:
00016 = 12 app functions, 8 DEFINER, 4 INVOKER, all 12 names. identity `store_test`: FORCE tables + `enrollment_requests`
(rls 17). tenancy integration down purge: + `DELETE FROM app.enrollment_requests` after the ledger. `audit.go`: +3
trigger actions (39), `ResourceEnrollmentReq` (11); `audit_test`: 39+6=45, 11 resource types read from 00016 (+00015
types kept). run.sh: B-AUDIT-CATALOG 45/39 (+message), `ENROLLMENT_OBJECTS` 15, B-MIGRATE-CLI/CATALOG relation count
66 → 71, B-IDENTITY-CATALOG message 15/71. AUDIT.md: T37–T39 rows, G058 paragraph (45 actions, resource type 10 → 11,
request capture producers). **Not moved (c2-owned):** B-AUDIT-IT `{"actions":42,"observed":42}` pin and AT-INV producers;
`tenancy/enrollment.go` comment and `enrollment_test.go` classify case still name `enrollment_kind_unsupported` (a
message string only; harmless; c2 updates with `AdmitKindRequest`).

**CT-ENR-1_requests (real runtime/claim-writer paths, each case rolled back).** approve (idempotent pending, authority_seq
= the replay row's seq, consumption once, covered row lifted `approved_request` by the decider, one binding, events
requested+decided only, membership.granted 1, second admission `already_member`); deny (ledger denied/trigger/decider/
source, decided event reason NULL, admit `denied`, new request = new id); revoke (revoked ledger by the revoker, revoked
event, admit `approval_revoked`); newer restriction (admit `enrollment_restricted`, zero writes); departed decider
(`approval_revoked`); expired pending (D3: same id returned, one unrenewed row, approvable, admit `approval_revoked`);
negatives (rt INSERT/DELETE/authority_seq/consumption 42501, cw decision 42501, decision attributed to another manager
42501, pending → revoked 42501, request with Organization context 42501, decision body rt 42501 / direct 0A000, member
decide 0 rows, approved → denied / → pending 42501, revoke after consumption 42501); atomic append (capture policy
dropped → request 42501, 0 rows; planted `CHECK (action <> 'enrollment.decided')` → denial 23514, still pending, 0 ledger).

**Change-impact map → checks.** 00016 up/down/catalog → CT-ENR-1 + TestMigrationLifecycle (+ TestMembershipCatalog:
shared `reset`/`enrollmentDrops`); request transitions/audit/atomicity → CT-ENR-1_requests; `enrollment_admit` open/lift
re-structured → access TestEnrollmentAdmission + TestEnrollment; FORCE list → identity TestIdentityCatalog; tenancy down
purge (FK order) → TestTenancyModel from version 16 seeded with a request + dependent ledger row; catalog counts →
audit/access/migrate units, doc-half reproducer, `ENROLLMENT_OBJECTS` = `enrollmentRelations`; compile → `go vet` both
tag sets. Disabled/startup variants: no pin depends on the new objects (B-TOGGLE-OFF/B-PERSIST check versions only).
**Concurrency** (one_pending ON CONFLICT path, Organization lock with decisions/departures): statically pinned only; the
executable ordering proof is c2's O-ENR work. AT-INV → TestAudit run for the truthful mismatch.

**Validation r9.** Unit/static lock-free (`docker run --rm --network none`, ro source, ro `g058-a1-gomod`,
`GOFLAGS=-mod=readonly`, `GOPROXY=off`, tmpfs GOCACHE, `golang:1.27.2-alpine`). DB under the harness lock (background
`flock -n … sleep 5400`, stopped after; lock verified free), mounts 14512/100000; internal network `g058a1r9-net`, fresh
`pgvector/pgvector:0.8.7-pg17` (tmpfs, repo init ro) per package, MinIO (`minio/minio:latest`, r7 deviation kept) for the
access package; env files from `r9db/mkenv.sh` (0600, no value printed), deleted after. Logs `~/.local/state/platform-
evidence/g058-a1/r9/`.

| Check | Exit | Counts | Log SHA-256 |
|---|---|---|---|
| gofmt -l (first) | 1 | 1 file (2 whitespace lines) → fixed | stdout |
| gofmt -l; go vet `./...`; go vet `-tags integration,workos_staging ./...` | 0 | clean | — |
| units audit, tenancy/..., migrate, identity, migrations | 0 | 6/6 packages ok | `r9-unit.log` `5caa5443…3eae5fd` |
| bash -n run.sh; B-AUDIT-CATALOG doc half | 0 | doc 45 = migration 45, T-rows 39 | `r9-auditdoc-{doc,mig}.txt` `a0efa60e…d3cfc`, `-trig` `fbe16143…14823` |
| `ENROLLMENT_OBJECTS` vs `enrollmentRelations` | 0 | 15 = 15 | `r9-enr-objects-*.txt` |
| migrate `^(TestEnrollmentCatalog\|TestMigrationLifecycle\|TestMembershipCatalog)$` (first) | 1 | requests: 9 test-side `true`→`t` boolean-format pins (all outcomes as asserted); Lifecycle/06 env: random superuser name, no `postgres` role | `r9db-migrate.log` `c74a2977…71d1f2` |
| same, rerun (pins fixed; env superuser `postgres`) | 0 | 3/3; CT-ENR-1 10/10 subtests | `r9db-migrate-2.log` `0b0d662e…7d7885` |
| identity `^TestIdentityCatalog$` | 0 | 4/4 subtests | `r9db-identity.log` `8b74b54b…2b2fc2` |
| tenancy `^TestTenancyModel$` (fresh, version 0) | 0 | 37/37 | `r9db-tenancy.log` `04fbcc35…7fe3` |
| tenancy `^TestTenancyModel$` from version 16 with a request + dependent ledger row | 0 | 37/37; stepped down to 5 | `r9db-tenancy-stepdown.log` `fd7aae34…12884` |
| access `^(TestEnrollment\|TestAudit)$` | 1 | TestEnrollment 2/2; TestAudit 8/9 — **AT-INV `{"actions":45,"observed":42}` known mismatch** | `r9db-access-enr-audit.log` `a74022f7…be96` |
| access `^TestEnrollmentAdmission$` | 0 | 11/11 | `r9db-access-admission.log` `1afd2214…f36b3` |

**Known mismatch (explicit, not green).** AT-INV/B-AUDIT-IT: catalog 45, observed 42 — `enrollment.requested`,
`enrollment.decided`, `enrollment.revoked` have no Go producer until c2 (`RequestEnrollment`, `DecideEnrollmentRequest`,
`RevokeEnrollmentApproval`). B-AUDIT-IT pin left at its prior text (c2-owned); not lowered or skipped. The task is not
fully green.

**Resources.** Created and removed: `g058a1r9-net`, `g058a1r9-pg` (×5 fresh), `g058a1r9-minio`, `--rm` test/init
containers; env files and `/tmp` scratch deleted; lock released. Kept: `g058-a1-gomod`, `r9db/mkenv.sh` (no secrets).
**Tool refusals (worktree guard, not retried in form):** (a) an `awk` program reading a test file ("…awk with a program
that can execute commands…") → plain `grep`; (b) `docker run … sh -c "gofmt -l …; go vet …"` ("…hands sh the text…") →
two plain `docker run` commands. Source-revert negative control: not done (the requests subtest's negatives serve as
privilege/transition controls).

**c2 transfer state.** Go: tenancy `AdmitKindRequest` + outcomes `pending`/`denied`/`approval_revoked`, `RequestEnrollmentTx`
(map `40001 enrollment_request_conflict` to retry), decide/revoke/list/status/abandon store functions + error codes
(`enrollment_request_transition` 42501 incl. the decider-is-actor rule), access facades + `log.go` ops, events.json/
logrecord rows; AT-INV +3 producers and run.sh B-AUDIT-IT 45/45; ENR/O-ENR ordering/authority/expiry proofs incl. the
one_pending race; then (d) `replay-restrict.sql` + §C.3 close + ENR-16; (e) B-ENROLLMENT-CATALOG/IT selectors.

### Round 1a-10 (c2) — request Go facades, AT-INV 45/45 and real-PostgreSQL behavioral/race proofs (appended 2026-10-18 UTC)

**Owner/inputs.** Fresh single writer Claude Opus 5.5 (dispatched high; actual **unknown/inherited**), native worktree
`.claude/worktrees/agent-<id>`, no subagents. Codex "Request c1" + "Round8" sections (review `b03d1ee2…56e2`),
Result `230e1f36…f245f50`, scope 50 (0 outside before each push). Attempt 1 / B7 / checkpoint 4 unchanged. Branch fetched
`6a705da2` (branch + HEAD asserted). Staging moved `f98b6ca8` → `594979b3` → `8b38d03e` → `d7cd0ec1`, every delta
`management/` only (G055 result/review, handoff, PROGRESS): exact-delta exception for the test batches (no bound input
touched); rebased before each push. SQL 00016 **unchanged** in c2. Parked `g058-a1-r8-request-slice-wip` untouched.

**Rulings applied.** Decider/revoker = acting user: the decide/revoke statements pass `tc.subject` as
`decided_by_user_id`/`revoked_by_user_id` (the 00016 guard re-checks `= ctx_user()`); authority_seq never supplied.
one_pending: `RequestEnrollmentTx` calls the body once; its 40001 fallback stays tenancy `unavailable` (class 40) and the
facade reports `unavailable` with reason **aborted** (from the transaction's first driver error), distinct from
`commit_outcome_unknown` (inContextTx `commitFailure`); **no retry in SQL, tenancy or access**. Issued approvals: the
admission's current-role check is the 00016 body (role, not status); no stronger invalidation. D2 reason NULL, D3
expired pending kept/unconsumable, D4 `DisplayName` always nil (no users read added).

**Source (c2).**
- `tenancy`: `AdmitKindRequest`, `AdmitRequestTx` (shared `admitTx`), outcomes `pending`/`denied`/`approval_revoked`
  (closed set 10); `EnrollmentRequestID` + parser, `RequestStatusValid`, `EnrollmentRequest` (13 columns), `RequestEnrollmentTx`
  (closed request outcome set 6, fail closed), `EnrollmentRequestTx`, `DecideEnrollmentRequestTx` (0 rows → re-read:
  not found / `not_pending`), `RevokeEnrollmentApprovalTx` (idempotent on revoked, `already_consumed`, `not_approved`),
  `EnrollmentRequestsPage` (keyset ≤100, status filter), `EnrollmentStateTx` (credential T2 proof + latest request; dead
  credential `enrollment_expired`), `AbandonEnrollmentTx` (own row, `revoke_reason` `abandoned`, idempotent). Codes:
  `enrollment_request_not_found`, `not_pending`, `not_approved`, `already_consumed`, `enrollment_request_transition`
  (42501 CONSTRAINT, also the decider/revoker rule), `enrollment_expired`; `enrollment_context` 42501 → `context_required`.
  `store.go` +8 statements (+3 writes: decide, revoke on `enrollment_requests`; abandon on `enrollment_sessions`). Stale
  `enrollment_kind_unsupported` comment/test case replaced (`enrollment_kind_invalid`/`enrollment_request_invalid` 22023 → internal).
- `access`: `RequestEnrollment`, `AdmitRequest`, `EnrollmentStatus`, `AbandonEnrollment` (own credential, IDENTITY shape
  through one helper `enrollmentTx`; `AdmitOpen` now shares `admitOwn`, behavior unchanged), `ListEnrollmentRequests`,
  `DecideEnrollmentRequest`, `RevokeEnrollmentApproval` (guarded, `organization.manage_members`; writes lock the Organization
  `FOR NO KEY UPDATE` via `policyWriteQuery` before the request row). Views `EnrollmentRequestView`/`Page`, `EnrollmentStatus`
  (`admitted` | request state | `none`), `DecisionApprove`/`Deny`. Ops `enrollment.request/.status/.abandon/.requests.list/
  .decide/.revoke`; reason `approval_revoked` (recorded; enrollment.admit only). Recording: admit + request record their
  known denials (NULL Organization); status/abandon log only; `pending`/`denied`/`already_member` are status conflicts
  (write_conflict, never recorded). inContextTx call sites stay 18 (`enrollmentTx` replaces AdmitOpen's).
- `events.json`: +6 operations, +`approval_revoked`; rules: `enrollment_restricted`/`approval_revoked` → enrollment.admit deny;
  `enrollment_closed`/`recipient_ineligible`/`host_changed` → admit|request deny; no capability for request/status/abandon;
  manager ops → `organization.manage_members`. `logrecord_test` +12 accept / +7 reject rows.
- Pins: tenancy `pinnedPackageFunctions` (+10), `TestSQLBoundary` 77/20 → 85/23, `writeTargets` (+2); access exported
  decls (+7) and methods (+7). run.sh B-AUDIT-IT `{"actions":45,"observed":45}` (moved only after AT-INV observed it).
- Tests: `TestEnrollmentRequests` (11 rows, below); AT-INV produces requested/decided/revoked through the facades with exact
  rows; tenancy/access unit tables for every new closed set, pre-SQL refusal and failure mapping.

**Change-impact map → checks.** tenancy request primitives/codes/statements → tenancy units (closed sets, pre-SQL,
classify) + `TestPackageBoundary`/`TestSQLBoundary`/`TestTenancyFixturePrimitives`; access facades/log/errors → access units
+ access boundary pins + `TestEnrollmentRequests` (DB); `AdmitOpen` refactor (shared helper) → `TestEnrollmentAdmission`
(all 11 rows) + `TestEnrollment`; AT-INV producers → `TestAudit` (9/9, AUDIT-INV); events.json → logrecord units + logcheck
validate of the DB-run access logs; run.sh pin text → `bash -n` + the AT-INV line it reads. **Not affected (c1 evidence
reused):** 00016 SQL, CT-ENR-1 catalog/updown, migrate lifecycle, identity catalog, tenancy stepdown, access admission
SQL rows' body, RLS/isolation catalogs (no grant/policy/function change). No disabled/startup variant reads the new code.

**Validation r10.** Units lock-free (`docker run --rm --network none`, ro source, ro `g058-a1-gomod`, `GOFLAGS=-mod=readonly`,
`GOPROXY=off`, tmpfs GOCACHE, `golang:1.27.2-alpine`). DB under the harness lock (background `flock -n … sleep 7200`; flock
and its sleep child stopped after; probe free), mounts 14540/100000; internal `g058a1r10-net`, fresh
`pgvector/pgvector:0.8.7-pg17` (tmpfs, repo init ro) per run, MinIO `minio/minio:latest` + repo `infra/minio/init.sh` for
the access runs needing objects; env files from `r10/mkenv.sh` (0600, superuser `postgres`, no value printed), deleted after.
Logs `~/.local/state/platform-evidence/g058-a1/r10/`.

| Check | Exit | Counts | Log SHA-256 |
|---|---|---|---|
| gofmt -l internal (first) | 1 | 1 test file (map alignment) → fixed | stdout |
| gofmt -l internal cmd migrations; go vet `./internal/tenancy/...`; vet `-tags integration,workos_staging` | 0 | clean | — |
| units tenancy/... (first, all) | 1 | access `TestAdmissionOutcomeMapping`: the old row expecting `pending` → internal (now a closed outcome) → row updated | stdout |
| focused units (Enrollment\|Admission\|Request\|Admit\|Boundary\|SQL) | 0 | 2/2 packages | stdout |
| logrecord `./internal/logrecord/` | 0 | ok | stdout |
| access `^TestEnrollmentRequests$` (fresh) | 0 | 11/11 rows, 6 ORDERING lines | `r10db-access-requests.log` `42ec4347…5090f` |
| access `^(TestAudit\|TestEnrollmentAdmission\|TestEnrollment)$` (fresh + MinIO) | 0 | TestAudit 9/9 — **AUDIT-INV `{"actions":45,"observed":45}`**; TestEnrollmentAdmission 11/11; TestEnrollment 2/2 | `r10db-access-audit-admission.log` `6e148c74…11925` |
| units tenancy/... (full, after the integration edits) | 1 | `TestTenancyFixturePrimitives`: ENR-RQA called the fixture-only `ChangeOrganizationRole` → replaced by G019's guarded `ChangeMemberRole` (inventory not widened) | `r10-unit-tenancy.log` `74932cdc…3c86a` |
| units tenancy/... (rerun) | 0 | 2/2 packages | `r10-unit-tenancy-2.log` `0da074c0…7c5f` |
| access `^TestEnrollmentRequests$/^ENR-RQA_` (fresh; the changed row only) | 0 | 1/1 | `r10db-access-rqa-rerun.log` `1bc572f7…1b5683` |
| logcheck validate (core) on the 3 access logs | 0 | 141 / 315 / 30 records, 0 unclassified, pass | access logs `6e1919ae…`, `6bf4560c…`, `a8087e8d…` |
| logcheck scan (3 canaries) over access/test/PostgreSQL logs; planted-canary control | 0; 1 | 0 hits; control detected (1 hit) | stdout |

PostgreSQL log of the first DB run was not captured (container removed before `docker logs`; its access/test logs were
scanned); the second and third runs' PostgreSQL logs were scanned (`25e3089a…`, `96cda7bf…`).

**Race/ordering proofs (real concurrent transactions; seam observed in `pg_locks`, not inferred from time).**
- O-ENR-RQ one_pending: credential 1's transaction held open right after `enrollment_request(` inserted (traced pool);
  credential 2 (same subject) waits on the unique index (lock wait observed, call unfinished; transaction-id wait, no tuple
  lock); after release both answer the **same** pending id, 1 row, 1 `enrollment.requested`, no error (no 23505/40001).
- O-ENR-7: first consumer held after `enrollment_admit(` (request row FOR UPDATE); second waits (tuple lock
  `enrollment_requests`) → `admitted` then `already_member`, same binding, 1 binding, 1 `membership.granted`.
- O-ENR-RV consumption first: revocation's facts queue on the **Organization** row (tuple lock `organizations`) →
  `already_consumed`, nothing written; revocation first (held before COMMIT): admission queues on `organizations` →
  `approval_revoked`, no binding.
- O-ENR-6 removal first (held before COMMIT): approval queues on `organizations`; derived `authority_seq` = the departed
  row's seq → admitted, row lifted; approval first: removal queues on `organizations`; newer departed row →
  `enrollment_restricted`, zero writes. Every path took the Organization before the request/ledger row; no deadlock.
- Class-40 vs unknown COMMIT (ENR-RQL): lost COMMIT reply → `unavailable`/`commit_outcome_unknown`, 1 transaction, the row
  committed, repeat returns it (1 row, 1 event); COMMIT answered 40001 → `unavailable`/`aborted`, nothing committed, repeat
  inserts once.

**Behavioral rows.** ENR-9 (pending idempotent per subject across credentials, applicant-attributed requested event,
admin approval with decider = admin and reason-NULL decided event, `not_pending`, admitted + consumption + proof, status
`admitted`, second credential `already_member`, revoke consumed `already_consumed`, member request `already_member`
unrecorded); ENR-9D (denied ledger row by the denier, `denied` status conflict unrecorded, open admit
`enrollment_restricted`, new request new id, re-approval lifts); ENR-9R (revoke idempotent with one event/one revoked row by
the revoker, `approval_revoked` recorded once, `not_approved` on pending, approval + newer denial → `enrollment_restricted`
zero writes, revoke + fresh approval lifts all 3 rows by the decider); ENR-RQA (org_member forbidden on list/decide/revoke
with 3 recorded denials, foreign manager `not_found`, invalid decision/zero id/limit/status/cursor `invalid_input`,
keyset pages by id with filter and nil display names, decider demoted via guarded `ChangeMemberRole` → `approval_revoked`,
request outside request_access `enrollment_closed` recorded on enrollment.request); ENR-RQE (expired pending: same id,
unrenewed, status `Expired`, admit `pending`, approval allowed but `approval_revoked`, decision frees the slot); ENR-RQB
(planted CHECK per action: request/denial/revocation each fail whole — no row, no ledger row, no event; succeed after).

**Design notes for review (within approved scope; flag, not decided here).** (1) AdmitRequest `pending`/`denied` and
RequestEnrollment `already_member` are `conflict` + `write_conflict` (not recorded): the 00016 catalog has no such
access.denied reason and no log reason was invented. (2) `EnrollmentStatus` reports `admitted` from the credential's T2
proof only (current membership is not readable without a tenant context); it records no denial for a dead credential;
`AbandonEnrollment` does not withdraw a pending request. (3) The callback's `request_on_arrival` → `enrollment_request` wiring
in identity (Plan r3 §1.3 step 3) is not in c2 (no identity change); it belongs with the HTTP/callback slice.

**Known mismatches.** None open for AT-INV (45/45 observed; pin moved). Hosted B-AUDIT-IT not run (quota; no pipeline
triggered: branch pushes are `when: never`, no MR exists).

**Resources.** Created and removed: `g058a1r10-net`, `g058a1r10-pg` (×3 fresh), `g058a1r10-minio` (×2), `--rm`
test/init/tool containers; env files deleted; lock released (probe free). Kept: `g058-a1-gomod`, `r10/mkenv.sh` (no
secrets), r10 logs. No image pulled or built. **Tool refusals (worktree guard; not retried in form):** a compound cleanup
(`docker network rm …; rm …; kill …; docker ps … {{.Names}} …`) — "…runs docker with the text {{.Names}} inside a construct
too complex to verify…" → split into plain commands (`docker ps -a --filter name=… -q`). Process notes: one `cat >>`
heredoc appended the new access unit tests and one `sed -i` updated the sorted access pin lists (no refusal; later edits
with Edit); one `awk` line filter was used on a grep listing (read-only). Source-revert negative control: not done (the
race rows assert the observed wait and the refused/aborted outcomes; the planted-CHECK rows are the rollback controls).

**Remaining before HTTP/UI (in order).** (d) `replay-restrict.sql` materialization + §C.3 close + ENR-16; (e)
B-ENROLLMENT-CATALOG/IT selectors + required-checks mapping (incl. `TestEnrollmentRequests`); identity callback
`request_on_arrival` wiring with the HTTP slice; then HTTP/UI, final local coverage and hosted qualification after quota.

### Round 1a-11 (d+e) — replay materialization/close + ENR-16, B-ENROLLMENT selectors and required-checks (appended 2026-10-18 UTC)

**Owner/inputs.** Fresh single writer Claude Opus 5.5 (dispatched high; actual **unknown/inherited**), native worktree
`.claude/worktrees/agent-<id>`, no subagents. Codex "Request c2" section (review `829e1211…eb122`) + accepted
correction 01 (§C.1–§C.7), Result `e66e78db…0d482d38`, scope 50. Attempt 1 / B7 / checkpoint 4 unchanged. Branch fetched
`86a22c95` (branch + HEAD asserted). **Staging reconciliation:** `5ca95c08` (rebased at start; the c2 Result commit dropped as
already upstream — Codex's import is byte-identical) → `6d104586` (management only; exact-delta exception for the batches) →
`fa2c2831` (G019 MR !49 merge `2710dcaa`: run.sh PROV-6 delta pin +2 claim_writer entries and "three helpers" PASS text).
Rebased onto `fa2c2831` after the batches: **no conflict**, both run.sh changes present (verified), `bash -n` + shellcheck
(B-SHELL form) clean on the rebased file; the G019 delta touches no G058 selector, so no rerun. Ancestry verified.

**Rulings applied.** Unrecorded status conflicts unchanged; `admitted` stays historical T2; `request_on_arrival` stays
pending for the HTTP slice. **Minor correction done (comment only):** `tenancy.AbandonEnrollmentTx` and
`access.AbandonEnrollment` now say an expired but unrevoked own row is revoked too (safe; no authority widened); idempotent
on an already revoked row. No predicate change, no test (per review).

**Source (d).** `infra/recovery/sql/replay-restrict.sql` (SHA-256 `8ef8f9d9…aab665`): after the unchanged `updated`/`state`
statements, one `SELECT set_config(...) \g /dev/null` (psql variables → transaction-local `replay.*`), one guarded `DO`
block, one `ledger<TAB>…` line. Block: `to_regclass('app.enrollment_restrictions') IS NULL` → `absent`; Organization-scope
`membership.remove` with the Organization present (taken `FOR NO KEY UPDATE`, the §1.4 chronology lock): scope id ≠ org, subject
user absent, or any Organization-scope binding remaining → §C.3 close (marker `replay-restrict`, `resolved_at/basis` NULL,
`open`→`invite_only`, `WHERE resolved_at IS NOT NULL`, idempotent) → `closed`; else unlifted `departed` row present →
`present`; else INSERT `departed`/`replay` (no creator, `created_at` = replay clock) → `materialized`. Other kinds, Group
scope or Organization absent → `unavailable` (nothing written). `recovery.sh`/`lib.sh` unchanged (kv() ignores the line;
updated/state decision unchanged). Value set exactly correction 01 §C.7 (5 values). 00016 unchanged (guard already accepts
`replay-restrict`; backfill/setter already §C.1/§C.2/§C.5).
`core/internal/migrate/enrollment_integration_test.go`: `CT-ENR-1_replay_pre16` (before the 00016 up) and `CT-ENR-1_replay`
(7 nested rows) run the **bound** script, passed as `ENROLLMENT_REPLAY_SQL_B64` (MCP_INVALIDATE_SQL_B64 precedent; SHA-256
pinned in the test) through `psqlStatements`/`psqlRun` (psql's statement split, `:'var'` literal quoting, `\g /dev/null`;
BEGIN/COMMIT removed so each case runs in an always-rolled-back transaction; each replay starts with the operator
settings cleared, as a fresh psql session). Restored states are emulated in one database; a savepoint rolled back after
live operator steps stands for restoring an earlier checkpoint.

**ENR-16 mapping.** (a) binding in checkpoint → `updated 1/none/present`, one trigger row, rerun `0/none/present`; (b)
restored trigger row → `0/none/present`; (c) predates grant+removal, restored approval (authority 0) → `materialized`, row
newer than the approval (created_at, seq), request admission `enrollment_restricted` with zero writes, rerun `present`,
Group scope / absent Organization / `policy.disable` → `unavailable`; (d) version 15 → `1/none/absent`, no ledger, then the
up closes L (legacy_postcondition); (e) backfill+attest before checkpoint → `present`, readiness/events unchanged; (f)
checkpoint before them (savepoint restore) → `materialized`, still closed, open → 23514; (g) ready+open checkpoint, later
backfill `l9.` → live `inserted` + close + 3 events; restore → ready+open, no row; replay `materialized`, still ready/open,
X open admission `enrollment_restricted` (zero writes), V admitted (control), rerun `present`; (h) absent subject →
`0/absent/closed`, L NULL/invite_only, +policy_changed +readiness_changed(history_unresolved), actor unknown; rerun same, no new
events; role mismatch (`org_admin` vs `org_member`) → `0/org_member/closed`, binding kept, no row; (i) reused unchanged
`CT-ENR-1_readiness` (22004/22023/not_legacy). Restored invitation: same `created_at` rule as ENR-23 (reused).

**Source (e).** `infra/validation/run.sh` `enrollment_checks` after `membership_checks`, before the drill:
B-ENROLLMENT-CATALOG = `rls_go … ./internal/migrate/ '^TestEnrollmentCatalog$' -e ENROLLMENT_REPLAY_SQL_B64=…`,
`it_count == 19`; B-ENROLLMENT-IT = `'^(TestEnrollment|TestEnrollmentAdmission|TestEnrollmentRequests)$'` in access, pins
**2 / 14 / 15** (recursive: Admission 11 + 3 O-ENR-8; Requests 11 + 2 O-ENR-RV + 2 O-ENR-6) + `rls_logcheck`. Both IDs added
to the MinIO-failure FAIL list. `required-checks.tsv` 125 → **127** (`B-ENROLLMENT-CATALOG`/`-IT` migrations); header updated
(+2 G058; full-profile note 112 → 120 = 116 suite IDs + 4 final: the old 112 dated from G016 and missed G017–G019 +6; the
integration note "77" (computed 75) is pre-existing and left unchanged — flag only). CI controls read the registry
dynamically (no count pin) → unchanged. `profile-paths.tsv` unchanged (all touched paths already foundation → full).
`.gitlab-ci.yml` untouched; no CI triggered.

**Change-impact map → checks.** replay-restrict.sql → CT-ENR-1_replay/_pre16 (real PG, emulated psql) + real-psql parity
(4 entries) + drill compatibility review (drill entries are Group-scope or invalid-ref refusals → `unavailable`, no readiness
event, V9 unaffected; the hosted B-RECOVERY-DRILL executes the new block on every restrict entry); migrate test → itself +
gofmt/vet (integration tag); run.sh/required-checks → `bash -n`, shellcheck (B-SHELL form), exact selector run + it_count
grep reproduction, registry recount; abandon comments → none (comment-only). Not affected (c1/c2 evidence reused): 00016,
access/tenancy code, identity, AT-INV 45/45, logrecord.

**Validation r11.** Units lock-free (`--network none`, ro source, ro `g058-a1-gomod`, `-mod=readonly`, `GOPROXY=off`, tmpfs
GOCACHE, `golang:1.27.2-alpine`). DB under the harness lock (background `flock -n … sleep 5400`, stopped after; probe free),
mounts 14564/100000; internal `g058a1r11-net`, fresh `pgvector/pgvector:0.8.7-pg17` per run (tmpfs, repo init ro), MinIO
`minio/minio:latest` + repo `init.sh` for the IT run; env files from `r11/mkenv.sh` (0600, superuser `postgres`), deleted
after. Logs `~/.local/state/platform-evidence/g058-a1/r11/`.

| Check | Exit | Counts | Log SHA-256 |
|---|---|---|---|
| gofmt -l internal cmd migrations (first) | — | 1 file (comment `E''` rewrite, long closure) → fixed | stdout |
| gofmt -l; go vet `-tags integration ./internal/migrate/ ./internal/tenancy/...` | 0 | clean | — |
| migrate `^TestEnrollmentCatalog$` (fresh, B64 replay) | 0 | **19/19** recursive (12 top-level + 7 ENR-16 rows), first run | `r11db-migrate-catalog.log` `22b3c4ae…18b5`; PG `1e84e532…5a04` |
| real psql parity (same DB, `go run ./cmd/migrate up`, recovery.sh psql flags) | 0 | materialized / present / unavailable / closed, lines equal the emulation | `r11db-psql-parity.txt` `1f367dd8…f768` |
| access `^(TestEnrollment\|TestEnrollmentAdmission\|TestEnrollmentRequests)$` (fresh + MinIO; the exact hosted selector, one process) | 0 | it_count grep: **2 / 14 / 15** | `r11db-access-enrollment-it.log` `21035853…e9f7`; PG `be0b4ba5…8620` |
| logcheck validate (core) on the IT access log | 0 | 272 records, 0 unclassified, pass | `…access.log` `dd7d4938…a03a` |
| logcheck scan (3 canaries) over migrate/IT/PG/access logs + parity; planted control (canary file) | 0; 1 | 0 hits; control 3 hits | stdout |
| `bash -n` run.sh; shellcheck `-S warning -x -P SCRIPTDIR` run.sh lib.sh (before and after the fa2c2831 rebase) | 0 | clean | — |
| registry recount | — | 127 rows; full suites 116 (+4 final) | — |
| scope (50 patterns) vs `origin/staging..HEAD` | — | 45 paths, **0 outside** | `r11-changed-paths.txt` |

**Known mismatches / hosted-only.** B-ENROLLMENT-CATALOG/IT and the B-RECOVERY-DRILL execution of the new block are
hosted-only (quota; no pipeline triggered). ENR-16 runs on emulated restores (savepoint restore; psql emulation with a
real-psql parity spot check), not through `recovery.sh restore/replay` — the drill's restore/invalidation/replay gates are
G013's unchanged ones. ENR-17/21/22 `TestIsolationAttacks` rows remain absent (not part of d/e; final-coverage item).

**Open decision for review (not changed).** After restoring a checkpoint taken **before** an `operator_backfilled`
attestation (ENR-16 f), the replay recreates the departures as `replay` rows; re-running backfill returns `present`, and
`enrollment_readiness_set(…, 'operator_backfilled', …)` then returns `no_backfill` (it counts `recorded_via='backfill'` only).
The operator can still attest `operator_verified_absent` or backfill another departure; whether replay rows should satisfy
the backfill attestation is a setter-semantics question for Codex (00016 unchanged; correction 01 §C.4 row 4 implies plain
re-attestation).

**Deviations.** ENR-16 lives in `TestEnrollmentCatalog` (migrate, SQL-level, B-ENROLLMENT-CATALOG) instead of
`TestRecoveryPost` (drill phase), so it is focused and runnable without the full drill; replay-restrict.sql adds 3 top-level
statements (settings, `DO` with the two guarded writes, ledger line) where the Plan counted "+2 statements".

**Resources.** Created and removed: `g058a1r11-net` (×2), `g058a1r11-pg` (×2), `g058a1r11-minio`, `--rm` test/tool/init
containers; `r11db/` env files deleted; lock released (probe free); no round-11 container left. Kept: `g058-a1-gomod`,
`r11/mkenv.sh` (no secrets), r11 logs. No image pulled or built. **Tool refusal (worktree guard):** one
`docker run … sh -c "gofmt … && go vet …"` — "…hands sh the text gofmt -l internal cmd migrations && go…, which runs gofmt
with the text cmd in a plain command, so what it runs cannot be shown not to be git…" → my error against the plain-command
rule; re-run as two plain `docker run` commands (not retried in form). No other deviation from plain commands.

**Push sequence.** `a05bd9b0` (d; lease from `86a22c95`), then (e) `3f71aba7` + this Result.

**Remaining before final.** Identity callback `request_on_arrival` wiring (HTTP slice); HTTP/UI with its tests; final local
affected coverage; hosted qualification after quota (incl. B-ENROLLMENT-* and B-RECOVERY-DRILL); the open decision above.

### Round 1a-12 (re-attestation correction) — CORRECTION LOOP record (appended 2026-10-18 UTC)

**Owner/inputs.** Fresh single writer Claude Opus 5.5 (dispatched high; actual **unknown/inherited**), native worktree
`.claude/worktrees/agent-<id>`, no subagents. Codex "Replay/hosted-selector partial review; re-attestation
correction" (review `8ba388d0…5743`, origin/staging `57e4d283`); Result `3984c014…a093a003` (= this file before this
section). Scope 50, attempt 1 / B7 / checkpoint 4 unchanged; correction pass 1 for this issue. Branch fetched `370e47f7`
(branch + HEAD asserted). Staging `57e4d283` was not an ancestor; delta since `fa2c2831` management-only (handoff, PROGRESS,
this Result = Codex's byte-identical import, the review) → exact-delta exception for the batch, then rebased onto
`57e4d283` before this Result (the r11 Result commit dropped as already upstream; no conflict; source bytes unchanged by
the rebase; ancestry verified).

**Decision applied (resolves the r11 "Open decision for review").** `app.enrollment_readiness_set` step (5): the
`operator_backfilled` evidence is now ≥ 1 row `WHERE r.organization_id = p_org AND r.kind = 'departed' AND r.recorded_via
IN ('backfill', 'replay')` (was `recorded_via = 'backfill'`). `kind = 'departed'` states the review's wording; it is implied
for both provenances by `enrollment_restrictions_replay_shape`, so no backfill row loses standing. Unchanged: superuser
check first, NULL/basis/`l<seq>.<ref>` grammar, Organization `FOR NO KEY UPDATE`, idempotence, transaction-local
marker + audited UPDATE (one `enrollment.readiness_changed` row, operator reference as request_id), no write to
`enrollment_restrictions` (no provenance/seq/time rewrite), no automatic readiness or open, no audit-history inference.
Recovery guidance: after restoring a checkpoint taken before an attestation, the replay rebuilds the departures as
`replay` rows (readiness stays closed); the operator repeats the explicit `operator_backfilled` attestation with the
ledger reference. `operator_verified_absent` is not a substitute when departures are known. `replay-restrict.sql`,
backfill, guard, `recovery.sh` unchanged. No HTTP/UI, no route.

**Tests (extend).** `CT-ENR-1_definers`: setter body fragment pin (the new predicate + `RETURN 'no_backfill'`); the setter
md5 pin follows the reviewed body (computed from the migration; owner/invoker/no-grantee inventory unchanged).
`CT-ENR-1_readiness` (ENR-3 negatives): under a savepoint, Organization N's backfill + replay rows and L's own trigger row
→ L `no_backfill`; the empty-L `no_backfill` kept. `ENR-16f` after the existing replay/23514 asserts: repeated backfill →
`present`, readiness still `invite_only|-|t`, events unchanged; `operator_backfilled` `l6.ct-enr-replay` → `resolved`,
`invite_only|operator_backfilled|f`, exactly one new readiness row (operator_backfilled, actor unknown, the reference);
row `seq|replay|created_at|unlifted` identical before/after; runtime owner opens (1 row); X's open admission
`enrollment_restricted|-` with zero writes; V admitted (control); row still identical.

**Change-impact map.** 00016 setter → TestEnrollmentCatalog (real PG) + static units reading 00016 (migrations, audit,
access boundary). Not rerun (reuse c1/c2/r11): access/tenancy integration (AT-INV and the access helper attest only after a
backfill row or via verified_absent → same outcome), replay-restrict.sql and its real-psql parity (bytes unchanged,
SHA pin unchanged), B-ENROLLMENT-IT, selectors/registry (unchanged; catalog `it_count` stays 19: subtests extended, none
added).

**Validation r12.** Units lock-free (`--network none`, ro source, ro `g058-a1-gomod`, `-mod=readonly`, `GOPROXY=off`, tmpfs
GOCACHE, `golang:1.27.2-alpine`). DB under the harness lock (background `flock -n … sleep 3600`, stopped after; probe
free); `docker-desktop-mounts.sh` 14600/100000; internal `g058a1r12-net`, fresh `pgvector/pgvector:0.8.7-pg17` per run
(tmpfs, repo init ro, superuser `postgres`); env from `r12/mkenv.sh` (0600; no MinIO), deleted after. Logs
`~/.local/state/platform-evidence/g058-a1/r12/`.

| Check | Exit | Counts | Log SHA-256 |
|---|---|---|---|
| gofmt -l internal cmd migrations; go vet `-tags integration ./internal/migrate/` (rerun after the fix) | 0 | clean | — |
| static units `./migrations/ ./internal/audit/ ./internal/tenancy/access/` (non-integration) | 0 | 70 PASS, 0 FAIL | `r12-unit-static.log` `775c8431…faeec` |
| migrate `^TestEnrollmentCatalog$` run 1 (fresh, B64 replay) | 1 | 18/19; ENR-16f failed at its **last** read (test bug: still in the runtime role after V's admission → no rows); every earlier new assert passed, incl. `resolved` | `r12db-migrate-catalog-1.log` `5e635ae3…7ed6`; PG `3ea1c7e0…efb72` |
| migrate `^TestEnrollmentCatalog$` run 2 (fresh stack, `e.as` reset added) | 0 | **19/19** recursive (12 + 7 ENR-16), 0 fail/skip | `r12db-migrate-catalog.log` `bce621c7…951e`; PG `4e56d532…bdcde` |
| credential-term scan (`password|secret|PRIVATE KEY`, count only) over all r12 logs | — | 0 hits | stdout |
| scope (50 patterns) vs `origin/staging` (incl. this Result) | — | 46 paths, **0 outside** | `r12-changed-paths.txt` |

**Refusal (auto-mode classifier).** `grep PASSWORD …/r12db/in/pg.env | cut -d= -f2 > …/secret-values.txt` (a private
value file for an exact-value canary scan) — denied as "[Credential Materialization]". Not retried or routed around;
replaced by the count-only credential-term scan above (weaker than an exact-value scan; the r11 logcheck canary scan
covered the IT path, which this round did not run). Narrow remedy if exact-value scans are wanted: generate canaries
inside the env script (r11 pattern) instead of extracting passwords.

**Resources.** Created and removed: `g058a1r12-net`, `g058a1r12-pg` (×2), `--rm` test/tool containers; `r12db/` deleted;
lock released (probe free); no round-12 container left. Kept: `g058-a1-gomod`, `r12/mkenv.sh` (no secrets), r12 logs. No
image pulled or built. No CI triggered, no MR.

**Remaining before final.** Unchanged from r11 except the open decision (now applied, pending Codex review); HTTP
execution not released.

### Round 1a-13 (HTTP/identity callback slice) (appended 2026-10-18 UTC)

**Owner/inputs.** Fresh single writer Claude Opus 5.5 (dispatched high; actual **unknown/inherited**), native worktree
`.claude/worktrees/agent-<id>`, no subagents. Release: Codex "R12 re-attestation ACCEPT; HTTP execution
release" (review `1c1fb76f…f53d`, origin/staging `7bf7f59e`); Result `f56ed3e5…8fd38` (= this file before this section).
Plan r3 §1.3/§1.5/§1.6/§5.1/§5.2/§6.2, r0 §1.3/§1.7/§5.2/§6.2–6.3, selected-design addendum §A.5/§A.7. Scope 50,
attempt 1 / B7 / checkpoint 4 unchanged. Branch fetched `2db012c9` (branch + HEAD asserted). Staging `7bf7f59e` delta since
`57e4d283` management-only (handoff, PROGRESS, review, this Result = Codex's byte-identical import) → exact-delta exception
for the batches; rebased onto `7bf7f59e` after them (the r12 Result commit dropped as already upstream; post-rebase tree
differs from the tested tree only in staging's three management files; ancestry verified). 00016 SQL **unchanged**; no
migration, vendor, dependency (`golang.org/x/time` already in core/go.mod), CI, MR or UI change.

**Source.**
- `core/internal/enrollment/api/http.go`: the routes, registered by the existing composition handler (`Register`) after
  `Bind`; `Config` gains the shared `Resolver` (main passes it). Pairs: `GET public`, `GET status`, `POST request`,
  `POST admit`, `POST abandon`, `GET policy`, `POST policy`, `GET requests`, `POST requests/{id}/decide`,
  `POST requests/{id}/revoke` (all under `/internal/v1/enrollment`); other methods on served paths 405, the root and any
  other sub-path 404 (closed JSON). Trust shapes: public (no credential; resolver decision; 404 `unknown_host`); applicant
  (`Authorization: Enrollment <credential>` only; exact https origin that routes now, else `origin_rejected`; per-credential
  limiter keyed by a digest, before resolution; identity resolution at the exact origin; credential Organization must equal
  the Organization the origin routes to now, else 401; CSRF on mutations; never a tenant context); manager (session-scope
  `begin`/`scope`/`originHost`/`query`/`hostTag`/`writeJSON` = byte copy of onboarding/api, pinned). Answers: public
  `{mode,label}`; status `{status, request|null, member, csrf_token}` — `status` admitted is the credential's historical T2
  proof, `member` is the current `is_organization_member` re-read (an admitted-then-removed applicant is `member:false`;
  nothing restores access), `csrf_token` is the credential's own token for the BFF forms; request `200 {request}`; admit
  `200 {outcome, next:"sign_in"}` (never a session); abandon `200 {}`; policy `{mode, open_ready, history_basis|null}`;
  requests `{requests, next}`; decide/revoke `200 {request}`. Applicant request view: id, status, expired, consumed,
  created/expires/decided/revoked_at only (no subject, decider, revoker, binding, authority); manager view adds user_id and
  null display_name. Refusals: onboarding's closed set + `enrollment_unavailable` 503 (class-40 abort or unknown COMMIT —
  never a guessed success), `rate_limited` 429, `unknown_host` 404, `enrollment_closed` 403 (no field); conflicts carry the
  closed field (`enrollment_restricted`, `approval_revoked`, `pending`, `denied`, `host_changed`, `recipient_ineligible`,
  `already_member`, `not_pending`, `already_consumed`, `not_approved`, `enrollment_not_ready`). Limiter (r0 §1.7): 5/s
  burst 10 per credential, 20/s burst 40 per resolved Organization, ≤ 10 000 keys. One `enrollment.request` event per
  request (operation, outcome, reason, host tag).
- `core/internal/identity`: `EnrollmentCredential(ctx, credential, origin)` — pre-SQL refusal of a malformed credential /
  absent or overlong origin, then one pool statement through `app.auth_resolve_enrollment` (unrevoked, unexpired, active
  user); no row or a row minted for another origin → `unauthenticated`; returns `EnrollmentCredentialInfo{Handle, CSRFToken,
  ExpiresAt}` (deviation 2) — and `EnrollmentCSRFMatches` (constant-time, the `CSRFMatches` rule). `AdmitKindRequest`; T1
  carries the stored `request_on_arrival`; `callbackSelector`: invitation intent > open > request kind on `request_access`
  with `request_on_arrival` > none; `admitAtCallback` forwards no invitation/address for the request kind. HTTP: login body
  `intent` ("enroll" only), `invitation_id` (canonical, non-zero), `request`; callback body `invitation_id`,
  `invitation_token`; enrollment answers `200 {outcome, next}` (confirmed admission; no credential, no session) or
  `202 {enrollment:{status, credential:{value,max_age}}}`; refusals keep the closed vocabulary (403 `enrollment_closed`, 409
  `invitation_mismatch`/`identity_conflict`, 503 `identity_unavailable`). Identity routes still 6.
- `core/internal/tenancy/access`: `AdmitAtCallback` request kind → `RequestEnrollment` (status `pending`; recorded known
  refusals pass through as statuses; a current member's unrecorded `already_member` → status `already_member`, the ordinary
  sign-in decides; any other failure incl. unknown COMMIT → error → `identity_unavailable`). `PublicEnrollment(ctx, org)`
  (`EnrollmentPublic{Mode, Label}`): two pool statements (policy reader mode, G017 public branding label), no context, nothing
  recorded; malformed/zero Organization `not_found`.
- `core/internal/tenancy`: `ReadPublicEnrollmentMode` + `sqlPublicEnrollmentMode = SELECT app.rls_enrollment_mode($1)`.
- `core/cmd/server/main.go`: `Resolver` passed, `enh.Register(mux)`; imports unchanged.
- Pins: tenancy `pinnedPackageFunctions` +1, `TestSQLBoundary` 85/23 → **86/23**; access exported decls +1
  (`type EnrollmentPublic`) and methods +1 (`Service.PublicEnrollment`). `run.sh`: `MAIN_GO_G016_SHA` → `b4fed819…723d`,
  new `CORE_ENROLLMENT_ROUTES` (10) + its B-MCP-SOURCE check line, access importers + `core/internal/enrollment/api`
  (**pre-existing red pin since slice 1a**: enrollment/api already imported access while the pin listed five importers),
  `core/internal/enrollment/api` importers = `core/cmd/server`, B-MCP-SOURCE PASS text. Logrecord: `events.json` +1 event
  `enrollment.request` (closed operations/reasons), `completion.schema.json` +9 routes (both enums), `logrecord_test` +5
  accept / +5 reject / +2 completion rows and the catalogue pin.

**Tests (ENR-H and negatives).** New `enrollment/api/http_test.go` (unit, no database): `TestEnrollmentHTTP` (every pair:
missing/duplicate origin, unknown parameter, http/foreign origin, no credential, duplicate Authorization, not JSON, query on
a mutation, unknown and forged members `user_id/address/email_verified/role/session/organization_id`, trailing data,
malformed, 8 KiB + 1; 405 closure; root/sub-paths 404; no credential read and only resolver statements),
`TestEnrollmentPublic` (resolver + mode + branding statements only; exactly `{mode,label}`), `TestEnrollmentCredentialBinding`
(Session/Bearer/lowercase/blank schemes = no credential; dead 401; reassigned host 401; store failure 503; wrong/missing CSRF
403 before any facade; live credential reaches exactly one facade transaction → 503 `enrollment_unavailable` (fault: no
transaction can open); admit body closure `kind`/`request_id`), `TestEnrollmentRateLimited` (constants pinned; refill frozen:
burst then 429 without a credential read; separate bucket; public burst; key bound), `TestEnrollmentRefusalMapping`,
`TestEnrollmentAnswerShapes` (ENR-17 minimal disclosure; ≤ 600 bytes), `TestEnrollmentEventIsClosed` (no credential, origin,
host, Organization or user in the event), `TestEnrollmentRoutesPinned`, `TestEnrollmentCopyIsExact`, `TestEnrollmentNew…`.
Identity: `TestEnrollmentCredential` (pre-SQL refusals with zero statements; digest-only argument; another origin / no row
401; store failure 503; CSRF matcher), `TestLoginEnrollmentIntent`, `TestCallbackEnrollment` (forwarded intent; 200/202
shapes; credential only in its field, no session; 403/409/409/503; malformed/zero id and forged address fields refused before
the service; ordinary answer unchanged); `TestAdmitAtCallbackOrchestration` +3 request-kind rows (unknown-kind row now
`lift`) and +4 selector rows. Access unit: `TestAdmitAtCallbackMapping` (+request with malformed handle, unknown kind `lift`,
`PublicEnrollment` pre-SQL refusals). Tenancy unit: `TestReadPublicEnrollmentMode`. DB (extend, no new subtest; B-ENROLLMENT-IT
counts 2/14/15 unchanged): **ENR-1** += anonymous `PublicEnrollment` on the runtime pool (closed default `invite_only`, label
nil; follows the stored mode; unknown Organization `invite_only`; zero audit rows); **ENR-INTENT** += `request_on_arrival` on
`request_access` (status pending + credential, exactly one pending row and one applicant-attributed `enrollment.requested`
with the Organization, no binding, not admitted; repeat login idempotent — still 1/1; credential resolves at its own origin
with the T1 digest and a 32-byte CSRF token, not at another origin, not as a session, not after abandonment; login without
the flag writes no request).

**Change-impact map → checks.** enrollment/api (new routes) → its unit tests + static B-MCP-SOURCE reproduction; identity
enroll/http/identity → identity units (whole package) + ENR-INTENT (DB); access enrollment (`AdmitAtCallback`,
`PublicEnrollment`) → access units + boundary pins + ENR-INTENT/ENR-1 (DB); tenancy store/enrollment → tenancy units
(`TestSQLBoundary`, `TestPackageBoundary`, `TestReadPublicEnrollmentMode`) + ENR-1 (DB); main.go → digest/imports/route pins
(static reproduction; no runtime start in this slice); run.sh → `bash -n`, shellcheck, static reproduction; logrecord
catalog → logrecord units + logcheck validate of the batch's access log. **Not affected (reused, not rerun):** 00016 SQL and
its catalog/updown/replay proofs (r11/r12), TestEnrollmentRequests and the admission rows other than ENR-INTENT (invitation/open
branches unchanged; their facades untouched), AT-INV 45/45 (no new producer: request at callback reuses the c2 producer),
identity store/revocation proofs, recovery/replay, required-checks registry (127; no new check ID).

**Validation r13.** Units lock-free (`docker run --rm --network none`, ro source, ro `g058-a1-gomod`, `-mod=readonly`,
`GOPROXY=off`, tmpfs GOCACHE, `golang:1.27.2-alpine`). DB under the harness lock (background `flock -n … sleep 3600`; flock
and its sleep child stopped after; probe free); `docker-desktop-mounts.sh` 14604/100000; internal `g058a1r13-net`, fresh
`pgvector/pgvector:0.8.7-pg17` (tmpfs, repo `infra/postgres/init` ro, superuser `postgres`), env from `r13/mkenv.sh` (0600,
three generated canaries; no credential extracted into any file), deleted after. Logs `~/.local/state/platform-evidence/g058-a1/r13/`.

| Check | Exit | Counts | Log SHA-256 |
|---|---|---|---|
| gofmt -l internal cmd migrations (first) | 1 | 1 test file (map alignment) → fixed | stdout |
| gofmt -l (rerun, and after the integration-test edits); go vet `./internal/identity/... ./internal/tenancy/... ./internal/enrollment/... ./cmd/server/`; vet `-tags integration,workos_staging` (same packages; access re-vetted after the test edits) | 0 | clean | — |
| units `-v ./internal/identity/ ./internal/tenancy/... ./internal/enrollment/... ./cmd/server/` (first run) | 0 | 4 packages ok (cmd/server no test files); 144 top-level PASS, 0 FAIL, 0 SKIP; all new tests listed above PASS | `r13-unit-1.log` `37bc856a…5e19` |
| logrecord `-v ./internal/logrecord/` (validationtools) | 0 | 10/10 | `r13-logrecord-1.log` `6a8c81f0…65e0` |
| access `-tags integration -run '^(TestEnrollment\|TestEnrollmentAdmission)$/^(ENR-1_\|ENR-INTENT_)'` (fresh PG, first run) | 0 | 2/2 selected rows (ENR-1, ENR-INTENT) | `r13db-access-intent-public.log` `0f6e7c6e…e444`; PG `4571d35d…5952` |
| logcheck validate (core) on the batch access log | 0 | 27 records, 0 unclassified, pass | `…access.log` `2082834d…680c` |
| logcheck scan (3 generated canaries) over test/access/PostgreSQL logs; planted control (PG log + canaries) | 0; 1 | 0 hits; control 3 hits | stdout |
| static B-MCP-SOURCE reproduction (`r13/static-check.sh`: run.sh's own expressions) | 0 | main.go digest, access/enrollment/tenancy importers, routes 6/8/16/10, literal and constant mux routes: PASS | script `cf2b0f22…a9a` |
| `bash -n` run.sh; shellcheck v0.11.0 `-S warning -x -P SCRIPTDIR` run.sh lib.sh | 0 | clean | — |
| scope (50 patterns) vs `origin/staging` (before this Result) | — | 49 paths, **0 outside** | `r13-changed-paths.txt` |

**Deviations / review items (within approved scope; flagged, not decided).**
1. **Route count.** Plan r1–r3 and the release say "9 pairs on 8 paths" (`CORE_ENROLLMENT_ROUTES` 9), but the enumerated
   contract (r0 §5.2's table: public, status, request, admit, abandon, policy GET+POST, requests, decide — itself 9 pairs/8
   paths although its header says 8/7 — plus r1's revoke), the 10 §1.5 operations and the 10 `enrollment-client.ts` exports
   all give **10 pairs on 9 paths**; the count arithmetic omits the anonymous public route. Implemented the enumerated set
   (no extra route, none dropped) and pinned 10 (round-8 precedent: explicitly named items over inconsistent counts). If
   Codex intends a different set, one Route constant + the pin change.
2. **`EnrollmentCredential` returns `EnrollmentCredentialInfo{Handle, CSRFToken, ExpiresAt}`** instead of Plan §5.1's bare
   `EnrollmentHandle`, because the plan's applicant mutations carry `csrf_token` and the BFF needs the token (the session
   route's established pattern). The status answer adds `csrf_token` to r0's `{status, request, member}` for the same
   reason; `member` is the current re-read (Codex round-2 item 4, c2 disposition 2), never the historical status.
3. **`POST request` answers 200 for both a new and an existing pending request** (r0 says 201/200): the accepted 00016 body
   and `RequestEnrollment` do not report which happened; no SQL change was made to distinguish them.
4. Callback 202 carries the credential as `{value, max_age}` (the existing `session`/`binding` secret shape) so the BFF can
   set the ≤ 1800 s cookie; r0 wrote `credential` without its shape.
5. `golang.org/x/time/rate` is newly imported by enrollment/api (already a core dependency used by MCP); no go.mod change.

**Known mismatches / not done.** Runtime B-TENANCY-HTTP positive pins for `/internal/v1/enrollment` (r0 §3 harness) not added:
it needs a running stack; deferred to final coverage with the browser/web slice (open item). No end-to-end HTTP run against a
started server (units + facade DB rows only). Hosted B-MCP-SOURCE / B-ENROLLMENT-* / logcheck not run (no pipeline triggered).
ENR-17/21/22 attack rows, web/UI, E-ENR browser, M1 and hosted qualification remain.

**Refusals.** None this round (all plain commands; no `$(...)`, heredoc, `sed -i` or compound shell). Process note: the
first lock release killed only `flock`; its `sleep` child kept the lock until stopped explicitly (probe then free).

**Resources.** Created and removed: `g058a1r13-net`, `g058a1r13-pg`, `--rm` test/tool containers; `r13db/` deleted; lock
released (probe free); no round-13 container left. Kept: `g058-a1-gomod`, `r13/mkenv.sh` (no secrets), `r13/static-check.sh`,
r13 logs. No image pulled or built. No CI triggered, no MR.

**Remaining before final.** Codex review of this HTTP slice (deviations 1–4); web/BFF/UI slice (cookie contract, join pages,
admin page, `enrollment-client.ts` 10 exports) with B-TENANCY-HTTP positive pins and E-ENR browser rows; ENR-17/21/22; final
local affected coverage; hosted qualification after quota (B-ENROLLMENT-*, B-RECOVERY-DRILL, B-MCP-SOURCE).

### Round 1a-14 (web/BFF/UI slice and HTTP positives) (appended 2026-10-18 UTC) — PARTIAL, context checkpoint

**Owner/inputs.** Fresh single writer Claude Opus 5.5 (dispatched high; actual unknown/inherited), native worktree
`.claude/worktrees/agent-<id>`, no subagents. Release: Codex "R13 HTTP partial source ACCEPT; web slice release"
(review `07e9725c…0079`, staging `d83093ee`); Result `8daabb59…e1f4`. Vestive skill read from `web/.claude/skills/vestive-design/SKILL.md`
(MR48 layout not assumed). Branch fetched `07256a5e` (branch + HEAD asserted). Batches under the exact-delta exception
(staging `d83093ee`→`86e38923` management-only); rebased onto `86e38923` before push (r13 Result commit dropped as already
upstream; post-rebase tree = tested tree + staging's four management files). No migration/SQL, vendor, dependency, core or
public-contract change; no CI, MR or G020.

**Source (web).**
- `web/lib/enrollment-client.ts` (server-only): the ten route functions `publicEnrollment`, `enrollmentStatus`,
  `requestEnrollment`, `admit`, `abandon`, `getPolicy`, `updatePolicy`, `listRequests`, `decideRequest`, `revokeRequest`
  (strict shapes, per-endpoint byte bounds; `Enrollment` scheme for applicant routes, `Session` for manager/auth routes, none
  for public; status requires `csrf_token` and the current `member`; POST request accepts **200 only** for new and existing;
  policy rejects `open` with `open_ready=false`); the enrollment variants of the two modified auth routes
  `startEnrollment` (login `intent:"enroll"`, `invitation_id`, `request`) and `completeCallback` (ordinary 200 answers incl.
  the enrollment login's `/join/status` return path; 200 `{outcome,next:"sign_in"}` with no session; 202
  `{enrollment:{status,credential:{value,max_age≤1800}}}`); closed notices; form handlers `joinStartRoute`,
  `applicantRoute`, `manageActionRoute` (route files stay thin and import no request-context).
- `web/lib/session.ts`: `CookieKind` += `enrollment` (`__Host-enroll`/`enroll`, ≤ 1800 s via `setCookie`) and `join`
  (`__Host-join`/`join`, only via `setJoinCookie`/`readJoinCookie`, regex `<uuid>.<43>`, 600 s, duplicate/malformed = absent;
  `setCookie("join")` throws; `readCookie` never reads it).
- Routes: `join/{start,request,admit,abandon}/route.ts`, `organization/enrollment/action/route.ts` (POST; other methods 405).
  `/join/start` = login binding + join cookie (or join cleared) + enrollment cleared in one 303; token never in a Location.
  Applicant routes: 303 to `/join/status?status=<closed>`; confirmed admission clears the enrollment cookie, never sets a
  session; 401/no cookie → `expired` + cookie cleared; unknown outcome → `unknown` (never success).
- Pages: `join/page.tsx` (mode entry; invite-only has no form), `join/status/page.tsx` (reconciliation read every request;
  `admitted` → ordinary sign-in form; admitted-but-removed shows not current member; expired pending/approval stated as
  expired), `join/invitations/[id]/page.tsx` (fragment prefill via the G019 `TokenPrefill`), `organization/enrollment/page.tsx`
  (mode for members; owner radio with `open` disabled + generic not-ready copy; admin request list with Approve/Deny/Revoke).
  Components `components/enrollment/{join-form,status-card,policy-form,request-list}.tsx`.
- Modified: `auth/callback/route.ts` (uses `completeCallback`; forwards the join intent in the body; clears the join cookie on
  every outcome when present; 202 → enrollment cookie only; `enrollment_closed`/`invitation_mismatch` → `/join?status=`),
  `app/page.tsx` (logged-out Join link when mode ≠ invite_only), `sign-in/page.tsx` (`admitted` copy),
  `organization/members/page.tsx` (one Enrollment link via the frame).
- Harness: `WEB_ROUTES` +9 (39→48), `static_web_boundary` request-context allow +1 (`lib/enrollment-client.ts`),
  `APPLICABLE` +4 (`core/internal/enrollment` — missing since slice 1a — `web/lib/enrollment-client.ts`, `web/app/join`,
  `web/app/organization/enrollment`), `profile-paths.tsv` +3 foundation rows, **B-TENANCY-HTTP enrollment pins** (root and
  unserved sub-path 404 `not_found` on GET and POST; `/public`,`/status` POST 405; GET without origin 400
  `ambiguous_parameter` on `/public`,`/status`,`/policy`; `/request` GET 405; served mutations POST `{}` closed 400/401/403).

**Deviations (flagged, within scope).** (1) `lib/identity-client.ts` is outside scope, so the enrollment login/callback
variants live in `enrollment-client.ts` (two exports beyond the ten route functions) and the callback route uses
`completeCallback` for every completion (only core knows the stored intent); the ordinary branch mirrors identity-client's
validation (plus `/join/status`, mapped to `/`). (2) static_web_boundary +1 entry, not +5 (route handlers live in the
client). (3) The G019 Organization-invitation link still points to `/invitations/<id>#token` (membership-client and the G019
page are outside scope); first-time invitees need `/join/invitations/<id>#token` — open item. (4) B-TENANCY-HTTP mutation
POST pins accept 400/401/403 pending the first runtime run (decode/origin order not proven at runtime yet).

**Validation r14.** `docker-desktop-mounts.sh` 14636/100000; harness lock held (background flock sleep), released after (task
stopped, no sleep child, probe free). Harness run from the worktree, `--base origin/staging`.

| Check | Exit | Counts | Log SHA-256 |
|---|---|---|---|
| h1 `lint-web unit-web` (first) | 1 | lint-web 4/4 PASS; B-UNIT-WEB FAIL: eslint 1 warning (unused `readJoinCookie` import) → fixed | `r14/h1.log` |
| h2 `lint lint-web unit-web` | 1 | all PASS incl. B-SHELL, B-WEB-BOUNDARY, B-WEB-TOKENS/COMPONENTS/DESIGN/ASSET-CONTROLS, **B-UNIT-WEB** (lint 0 warnings, typecheck, `test:unit` **247/247**, incl. new 12 client + 9 forms cases), B-SECRET-CANARY, B-DOCKER-ONLY; B-SCOPE/B-REQUIRED-CHECKS FAIL **only** on staging's management delta (exact-delta artifact; cleared by the rebase) | `r14/h2.log` `221cb189…42fc`; `h2/diagnostics/unit-web-check.log` `51a2ccfa…8813` |
| `bash -n` run.sh | 0 | — | — |
| scope (50) vs `origin/staging` `86e38923` after rebase | — | 71 paths, **0 outside** (+ this Result) | `r14/r14-changed-paths.txt` |

**Change-impact map.** web lib/session/client/routes/pages/components → B-UNIT-WEB (unit+lint+typecheck) + lint-web; run.sh →
B-SHELL; B-TENANCY-HTTP pins → runtime root-suite run (**not run**); callback route ordinary branch → identity browser rows
(**not run**). Core/SQL/replay unchanged: r11–r13 proofs reused (units `37bc856a…`, DB `0f6e7c6e…`, catalog/replay r11/r12).

**Not done (transfer state for the next writer).** (a) Runtime B-TENANCY-HTTP (root suite, shared stack, lock) — tighten the
mutation POST codes once observed. (b) Actual successful HTTP route-to-facade positives: the access integration tests are
`package access` (cannot import `enrollment/api`) and only `http(_test).go` is in scope for `enrollment/api`, so the route
positives must come from the started identity-overlay stack: E-ENR browser rows driving status (member, csrf_token),
request/admit/abandon, policy GET/POST, list, decide, revoke through the BFF, plus wrong-authority negatives (member→list
403, admin→policy POST 403). (c) Seeds (in scope, untouched): `mcp-seed.sql` B `open`, C `request_access`, A closed via
`enrollment_readiness_set(A, NULL, 'mcp-seed.g058')`, dave as asserted C nonmember applicant; `emulator-seed.yaml` +1
account-less `grace` (note: mcp-seed already has a principal-less local "Grace Admin"). (d) `web/tests/browser/enrollment.spec.ts`
(E-ENR-1..5) + `playwright.config.ts` projects `enrollment-light/dark` + `lib.sh` `BROWSER_MATRIX_IDENTITY_E2E` +1. (e) M1,
ENR-17/21/22, final local affected coverage, hosted qualification.

**Refusals.** One: a compound `cd … && timeout 1800 bash run.sh … > log; echo; tail` was refused by the worktree-isolation
guard ("construct too complex to verify"); rerun as a plain single command (not routed around). **Resources.** Harness-owned
projects/images cleaned by the harness (`inventory-leftover.txt` empty); lock free; evidence `~/.local/state/platform-evidence/
g058-a1/r14/` (0700). No CI, no MR.

### Round 1a-15 (web slice continuation: runtime pins, fixtures, E-ENR positives) (appended 2026-10-18 UTC) — PARTIAL, frozen on coordinator instruction

**Owner/inputs.** Fresh single writer Claude Opus 5.5 (dispatched high; actual unknown/inherited), native worktree
`.claude/worktrees/agent-<id>`, no subagents; context rotation from 1a-14 (attempt 1, B7, checkpoint 4).
Release: R13 web slice release (review section unchanged); Result `f36cf4c7…0d55` (= this file before this section). Vestive
skill read from `web/.claude/skills/vestive-design/SKILL.md`. Branch fetched `a52ee651` (branch + HEAD asserted). Staging
`86e38923`→`5519d50b`→`801672b3`→`433e6048`, every delta management/process-only (handoffs, PROGRESS, reviews, G055 docs,
skills/workflow docs); rebased onto each before the batch / push. **Scope amendment04** (review section "Scope amendment04;
Organization invitation entry", `915344a1…c6e0`): scope 52, `G058-scope-allowlist.txt` `dc791b4b…ee23`; applied. Frozen at
the coordinator's "freeze G058 source now" instruction after batch h1 (no new batch started). No SQL/migration, core,
dependency, vendor or public-contract change; no CI, MR or G020.

**Source.**
- Amendment04: `web/lib/membership-client.ts` `invitationIssuedPage` selects the link by the issued invitation's own scope —
  Organization `${origin}/join/invitations/<id>#token`, Group unchanged `${origin}/invitations/<id>#token`; the obsolete
  "not available yet" sentence replaced ("The person opens the link, signs in (or creates an account with the invited
  address) and joins this workspace."). Same single fragment token, 200 no-store page, no Location, no title token. No other
  membership-client, G019 lifecycle/API or invitation-page change. `web/tests/unit/membership-forms.test.mjs`: the existing
  issued-page test now asserts the Organization path (join entry, new copy, no old path) and a Group invite through the
  Group action route (member path, no join entry/copy), each with one token occurrence, no Location, no title token.
  `web/tests/browser/membership.spec.ts` (directly corresponding): `LINK` accepts both paths; E-MEM-1 asserts the
  Organization link is `/join/invitations/…`, `inviteToGroup` asserts Group links stay `/invitations/…`.
- Fixtures: `infra/identity/mcp-seed.sql` — B `open`, C `request_access` (superuser UPDATE; B ready by construction), A
  history closed via `app.enrollment_readiness_set(A, NULL, 'mcp-seed.g058')` (asserted `closed`); on C carol `org_owner`
  and frank `org_admin` (replacing their plain C memberships; A/B unchanged), alice stays a plain C member, **dave and bob
  without a C binding** (bob's plain C membership removed; dave keeps his planning Group binding); a DO block asserts
  `is_organization_member` false for dave/bob on C, true for alice/carol/frank, and the three modes. Two applicants because
  the light and dark projects share one database and each needs its own single-use request applicant (deviation 1).
  `infra/identity/emulator-seed.yaml` +1 account-less `grace@identity.test` "Grace Invitee" (`email_verified: true`;
  distinct from the principal-less "Grace Admin").
- `web/tests/browser/enrollment.spec.ts` (new), projects `enrollment-light/dark` in `identityProjects()`, IDENTITY.md
  `identity:browser` line +2 projects, `lib.sh` `BROWSER_MATRIX_IDENTITY_E2E` +1 row (`E-ENR-1..5 S-ERRORS`, both
  projects), run.sh B-IDENTITY-E2E PASS text, IDENTITY.md/WEB-SHELL.md project rows. Per-project subjects (light: open
  erin, invited grace, applicant dave; dark: open grace, invited erin, applicant bob) so the rows are order-independent.
  Rows (all through the real BFF routes to core): **E-ENR-1** B open, keyboard-only on app pages: entry Join link →
  `/join` → callback admission → `/join/status?status=admitted` with no enrollment/join/session cookie → ordinary fresh
  sign-in → own Group + new Business Profile (G018 tail) → no owner/invitation controls, read-only mode, no request list.
  **E-ENR-3** A invite-only: no Join link/form; carol's Organization invitation link is `/join/invitations/<id>#token`
  (amendment04), prefill + fragment cleared, join cookie set before the callback and cleared after, admitted → fresh
  sign-in → member; token never in a request URL or Location; carol sees the member, invitation not pending.
  **E-ENR-4** C: request on arrival (status GET pending, enrollment cookie only, `/api/account` 401) → carol (owner)
  policy POST to invite_only → applicant POST `/join/request` → `closed` notice with core state still pending → carol
  restores request_access → frank denies (decide POST) → denied status; no tenant data; axe on pending/denied.
  **E-ENR-2** C: request on arrival → POST `/join/request` on the existing pending (idempotent 200 → `requested`) → frank
  policy GET + list GET → **wrong authority**: alice (member) decide POST → 403 "Not permitted", frank (admin) policy POST
  → 403 "Not permitted", nothing changed → approve → revoke (revoke POST) → status `revoked` → abandon POST (cookie
  cleared) → new request → approve → admit POST kind=request → `admitted`, cookies cleared → fresh sign-in → member;
  frank's list shows "Joined". **E-ENR-5** axe 1280/320 on the three join entries, status without credential, invitation
  landing, owner (A, open disabled + not-ready hint, `aria-describedby`), admin and member pages; keyboard: C entry →
  request start reaches the emulator; carol's members-page Enrollment link and two policy changes by arrow keys + Save.
  The applicant POST `/join/request` in states without a button is a same-origin DOM form submit of the existing route.
- run.sh B-TENANCY-HTTP: the enrollment POST `{}` pins tightened from `40[013]` to `403` for `/policy` and `/request`
  (from the handler order decode → origin → credential: an empty origin is `origin_rejected`); PASS text names the
  enrollment probes. **Not yet observed at runtime** (open item a).

**Deviations (flagged).** (1) Two request applicants (dave light, bob dark) instead of one, and carol/frank given C manager
roles: required for both themed projects to run every E-ENR row on one database; no earlier spec uses bob/carol/frank on C.
(2) The status page offers no button for POST `/join/request` once a request exists; the rows submit the existing route by a
same-origin form (positive: idempotent existing pending; negative: closed). (3) The open and invitation admissions happen at
the callback (T2), so POST `/admit` is exercised by the request kind (E-ENR-2) only.

**Validation r15.** `docker-desktop-mounts.sh` 14640/100000; harness lock held (background flock), released after (task
stopped, no sleep child, probe free). Harness from the worktree on HEAD `63b32063` (= tested tree), `--base origin/staging`.

| Check | Exit | Counts | Log SHA-256 |
|---|---|---|---|
| h1 `lint lint-web unit-web` | 1 | all PASS incl. B-SHELL (run.sh/lib.sh), B-HADOLINT, B-COMPOSE, B-ENV-EXAMPLE, B-DEPS, B-CI-CONFIG, B-SECRET-SOURCE, B-WEB-BOUNDARY/TOKENS/COMPONENTS/DESIGN/ASSET-CONTROLS, **B-UNIT-WEB** (eslint 0 warnings and `tsc` over `enrollment.spec.ts`/`membership.spec.ts`/config, `test:unit` **247/247**, 0 fail/skip, incl. the extended membership-forms issued-page test), B-DOCKER-ONLY, B-SECRET-CANARY (24 generated); B-SCOPE/B-REQUIRED-CHECKS FAIL **only** on staging's process files that landed during the run (`801672b3`→`433e6048`: `.agents/skills/*`, `CLAUDE.md`, `docs/agents/CLAUDE-SESSION-HANDOFF.md`, management, `skills/*`) — exact-delta artifact; after rebase onto `433e6048`: 82 paths, **0 outside** the 52 patterns | `r15/h1.log` `37553524…4766`; `h1/diagnostics/unit-web-check.log` `dbf9fbb4…9998` |
| `bash -n` run.sh after a comment-only edit (post-h1) | 0 | — | — |

**Not run (open items, transfer state).** (a) Runtime B-TENANCY-HTTP: `run.sh --base origin/staging --allow-file
management/v0/results/G058-scope-allowlist.txt root` under the lock (est. 30–45 min); confirm the `/policy` and `/request`
POST `{}` 403 pins. (b) Identity-overlay browser batch (prepared, not executed): `r15/idn.sh` (`up` → `seed` → `list`/`run`
per selector → `logs` → `scan` → `down`; git-archive copy of HEAD, project `g058a1r15-idn`, emulator IDs looked up inside the
emulator container, documented `mcp-seed.sql`), under the lock (est. 60–90 min): order (i) `--project=identity-light
--project=identity-dark --project=domain-platform --project=domain-custom` (callback ordinary branch, item d), (ii)
`--project=mcp-light --project=mcp-dark --project=onboarding-light --project=onboarding-dark --project=membership-light
--project=membership-dark --grep 'M-NONMEMBER|E-ONB-4|E-MEM-1|E-MEM-2'` (fixture impact + amendment04 link), (iii)
`--project=enrollment-light --project=enrollment-dark` (E-ENR-1..5); `--list` before each first run. Enrollment rows must run
after (i)/(ii) (they admit erin/grace/dave/bob). (c) ENR-17/21/22, M1, final local affected coverage, hosted qualification.

**Refusals.** None. **Resources.** Harness h1 resources cleaned (no `g001.run` container left); lock free; no stack started
for (a)/(b). Kept: `r15/` (0700) logs and scripts (`idn.sh`, `lookup.js`, `scan-in.sh`, `scope.sh`; no secrets).

### Round 1a-16 (runtime B-TENANCY-HTTP, identity overlay, Codex preliminary corrections) (appended 2026-10-18 UTC) — PARTIAL, blocker reported

**Owner/inputs.** Fresh single writer Claude Opus 5.5 (dispatched high; actual unknown/inherited), native worktree
`.claude/worktrees/agent-<id>`, no subagents. Branch fetched `71388d12` (asserted); Result in `aeb0f221…18e1`.
Staging `433e6048` (ancestor at start) → `3688f8d2` during check (a): delta exactly `management/SUPERVISOR-HANDOFF.md`,
`management/v0/PROGRESS.md`, `management/v0/results/G019-r0-binding-codex-review.md`,
`management/v0/results/G058-selected-design-codex-review.md` (management-only, exact-delta exception); rebased onto
`3688f8d2` after (a) → `a6ee4f96`. Codex "Frozen web checkpoint preliminary correction" (review section, 2026-10-18)
applied after (a) at a safe boundary → `694413f1`. No SQL/migration, core, dependency, vendor or public-contract change.
Draft MR !50 untouched; intermediate push with `ci.skip`; this handback push without it (coordinator instruction).

**Check (a) runtime B-TENANCY-HTTP.** No narrower entrypoint exists: `tenancy_http` is called only from `suite_root`
(`run.sh` line ~1799), so the root suite was run (coordinator instruction). `docker-desktop-mounts.sh` 14644/100000; lock held
by a background `flock` holder, released after. Command: `bash infra/validation/run.sh --base origin/staging --allow-file
management/v0/results/G058-scope-allowlist.txt --report-dir …/r16/a1 root` on HEAD `71388d12` (tested tree).

| Check | Exit | Counts | Log SHA-256 |
|---|---|---|---|
| a root | 1 | 19 PASS incl. **B-TENANCY-HTTP PASS** (86 probes; the enrollment `/policy` and `/request` POST `{}` **403** pins observed as tightened in r15, root/sub-path 404 `not_found`, `/public` `/status` 405 and 400 `ambiguous_parameter`, `/request` GET 405), B-BROWSER-ROOT 144 tests, B-CHECKS-ROOT, B-DOCS-EXEC-ROOT, B-SECRET-CANARY (14), boundaries, health, error outputs; **B-SCOPE / B-REQUIRED-CHECKS FAIL only** on the four staging management files above (exact-delta artifact, cleared by the rebase: h1 B-SCOPE PASS) | `r16/a1.log` `e1ab244e…2b62e`; `inventory-leftover.txt` empty, no `beginify` container, no `.env` left |

**Corrections (Codex preliminary review, both bounded).**
1. `web/app/auth/callback/route.ts`: the host is classified before the size check (the lookup sends the origin only,
   never the query). Shared auth host: no cookie on any answer (oversized → 400 without cookie, else relay unchanged).
   Unknown host / resolver failure (unclassified): no cookie read or written. Classified non-routing host (reserved,
   public entry): 303 `failed` + join intent cleared. Tenant host: oversized query → 400 + join intent cleared, no
   callback call; all other outcomes unchanged (`clearedJoin` helper). No relay, provider, session or token handling
   change; no token in Location/query/log.
2. `web/lib/enrollment-client.ts` `requestRowView` (pure; used by `components/enrollment/request-list.tsx`): expired
   pending → state "Expired, deny to resolve", actions `[deny]` (no Approve; approval never renews validity); valid
   pending → Approve + Deny; valid unconsumed approval → Revoke; expired approval → "Expired", none (non-consumable);
   consumed → "Joined". No SQL, expiration or credential change.
- Tests (extend existing files): `enrollment-forms.test.mjs` +3 (oversized query on a tenant host clears the intent and
  reaches no callback, the padded query never sent; reserved/public-entry clear the intent; auth host normal + oversized,
  unknown host normal + oversized, resolver 503: zero Set-Cookie, no callback; no token/secret in Location or cookies);
  `enrollment-client.test.mjs` +2 (`requestRowView` expired pending, and the other states).

| Check | Exit | Counts | Log SHA-256 |
|---|---|---|---|
| h1 `lint lint-web unit-web` on `694413f1`, base `3688f8d2` | 0 | 18 required IDs PASS incl. **B-UNIT-WEB** (eslint, tsc, `test:unit`; count not retained on a passing run, 247 + 5 new cases expected), B-WEB-BOUNDARY, B-SCOPE (82 paths, all within the 52 patterns), B-SECRET-CANARY (24), B-SHELL | `r16/h1.log` `2f618cf7…a76d` |

**Check (b) identity overlay.** `r16/idn.sh` (r15 script adapted: r16 paths, project `g058a1r16-idn`, source from a plain
`git archive -o r16/src.tar HEAD` of `694413f1`, `src.tar` `d8e73b04…5411`, so the script runs no git). up: configure 0
(.env 600), start 0, migrate 0 (version 16). seed: lookup 0 (alice bob carol dave frank hostile), seed 0; org-a
`invite_only`, org-b `open`, org-c `request_access` (`mcp-seed.sql` `6b8f9efc…058d`, `emulator-seed.yaml` `ad97a633…7c31`).

| Selector | List | Run exit | Counts | Log SHA-256 |
|---|---|---|---|---|
| (i) identity-light/dark, domain-platform, domain-custom | 67 | 0 | **67/67 PASS** | `pw-r1.log` `f6fc7d01…80c1`; `rep-r1/results.json` `b5f1b305…a23f` |
| (ii) mcp, onboarding, membership ×2 themes, grep `M-NONMEMBER\|E-ONB-4\|E-MEM-1\|E-MEM-2` | 8 (first list attempt exit 17: Docker Desktop registry proxy timeout resolving `docker/dockerfile:1` during `--build`, no test ran; one retry listed) | 0 | **8/8 PASS** (amendment04 Organization link, Group link unchanged) | `pw-r2.log` `d2d0090f…cfff`; `rep-r2/results.json` `c435c4a3…b84d` |
| (iii) enrollment-light/dark | 10 | 1 | **8 PASS, 2 FAIL**: E-ENR-1, -2, -4, -5 pass in both themes; **E-ENR-3 fails in both** at `enrollment.spec.ts:371` | `pw-r3.log` `359289f7…c5e7`; `rep-r3/results.json` `f877f8f0…6045` |

logs 0 (8294 lines, `idn-compose-logs.txt` `1b89b685…cdd`); scan: 127 patterns, **0 hits** across logs/reports, control 14;
down: stop 0, containers 0, volumes 0, images 0; lock released (probe free).

**Blocker (E-ENR-3), diagnosed, not fixed (outside the authorized change bound).** Everything up to line 371 passes (A
invite-only, `/join/invitations/<id>#token`, join cookie, callback admission, fresh sign-in, member). Then carol's
`/organization/members` shows "The invitation list cannot be shown right now." so the assertion that the recipient's
invitation row is not pending finds no row. Cause: core consumes an address-addressed Organization invitation by setting
`intended_user_id = consumer` (accepted G019 SQL `00015…sql` line ~895; CHECK `invitations_consumed_is_intended`), and
core's `invitationJSON` renders both `intended_user_id` and `intended_address` for the issuing administrators (core logs
`list_invitations ok`; DB rows for grace/erin: address + intended user + consumed). The web parser
`web/lib/membership-client.ts` `invitation()` rejects any row with both (`if (intended !== null && address !== null)
return undefined;`), so one consumed address invitation makes the whole list malformed. Latent G019 web defect; G058 is
the first path that consumes an address Organization invitation. Smallest fix (proposed, not applied): accept both forms
only when `status === "consumed" && scope === "organization" && intended === consumedBy`, plus one unit case. Not applied
because amendment04 limits `membership-client.ts` changes to the Organization link and copy ("no general membership-client
refactor, G019 lifecycle/authorization change"); the natural unit file `membership-client.test.mjs` is outside the 52
patterns (`membership-forms.test.mjs` is inside). The assertion is unchanged (not weakened). Needs a Codex scope decision;
then rerun only (iii) E-ENR-3 (both themes) plus `unit-web`.

**Change-impact map.** Callback route → B-UNIT-WEB new cases + identity rows (i) (callback ordinary branch, relay
E-DOM-2) PASS on the corrected tree; request-list → B-UNIT-WEB + E-ENR-2/-4/-5 PASS on the corrected tree; B-TENANCY-HTTP
unaffected by both corrections (no enrollment/web route set or refusal change), r16 (a) evidence reused.

**Not done.** E-ENR-3 (blocker above); ENR-17/21/22, M1, final local affected coverage, hosted qualification.

**Refusals.** Worktree guard refused three compound commands (`git merge-base …; echo`, `bash …mounts.sh; echo`, a
`docker ps --format '{{.Names}}'` pipeline: "construct too complex to verify"); each rerun as a plain single command, not
routed around. **Resources.** Root project and overlay fully removed; lock free; evidence `~/.local/state/platform-evidence/
g058-a1/r16/` (0700; `src/` removed; `src.tar`, scripts, logs, reports; no secrets, observed-value files none).

#### Round 1a-16 correction (scope amendment05; E-ENR-3 parser parity) (appended 2026-10-18 UTC)

**Inputs.** Codex scope amendment05 (review `843d0456…7318`): 53 patterns, `G058-scope-allowlist.txt` `ce8ff6b6…954e`
(+`web/tests/unit/membership-client.test.mjs`; `membership-client.ts` authorized for consumed-Organization parser parity).
Branch `d0093ab8` (asserted) rebased onto staging `0fae876e` (delta management-only: handoff, BACKLOG, PROGRESS, G019
closeout/review/result, G058 review and allowlist + pre-amendment05 copy, ONEOFF promotion record) → fix `bf4f9268`.

**Fix.** `web/lib/membership-client.ts` `invitation()`: both `intended_address` and `intended_user_id` are accepted only
when `scope === "organization" && status === "consumed" && intended_user_id === consumed_by` (core's 00015
`invitations_consumed_is_intended` consumption shape); every other ID/shape/bound/status check unchanged. No API, SQL or
authority change. `web/tests/unit/membership-client.test.mjs` +1 case: the consumed Organization positive (status,
address, intended user, consumer kept) and five negatives (Group scope, pending, revoked, expired, mismatched consumer) →
`malformed`. The existing "refuses inconsistent invitations" negatives (pending + intended user, consumed without
consumer) are unchanged.

| Check | Exit | Counts | Log SHA-256 |
|---|---|---|---|
| h2 `lint lint-web unit-web` on `bf4f9268`, base `0fae876e` | 0 | 18 required IDs PASS incl. B-UNIT-WEB (eslint, tsc, `test:unit`), B-WEB-BOUNDARY, **B-SCOPE** (83 paths, all within the 53 patterns), B-SECRET-CANARY (24) | `r16/h2.log` `aab30538…2849` |
| idn fresh stack (`src.tar` `cde0d91e…0aaa` = `bf4f9268`): up 0 (migrate 16), seed 0 (A invite_only, B open, C request_access); list `--project=enrollment-light --project=enrollment-dark --grep E-ENR-3` 2 | 0 | **E-ENR-3 2/2 PASS** (light, dark): carol's members list renders, the recipient's invitation row is not pending | `pw-r4.log` `7f7b83db…31fa`; `rep-r4/results.json` `33880f89…4682` |

logs 0 (564 lines, `idn-compose-logs.txt` `cdeb09a2…e6bc`; overwrites the first stack's file, which was scanned before its
teardown: 0 hits); scan 18 patterns, 0 hits, control 14; down: containers 0, volumes 0, images 0; lock released.

**Combined r16 browser state on the corrected tree.** (i) 67/67 and (ii) 8/8 on `694413f1`; (iii) E-ENR-1/-2/-4/-5 4×2 on
`694413f1` and E-ENR-3 2/2 on `bf4f9268`. Change-impact of the parser fix: membership list parsing only → B-UNIT-WEB +
E-ENR-3 (the only row consuming an address Organization invitation, then listing); E-MEM-1 (pending/revoked rows) and
other rows unaffected, their r16 evidence reused. **Still not done:** ENR-17/21/22, M1, final local affected coverage,
hosted qualification. **Refusals:** none in this correction.

### Round 1a-17 (ENR-17/21/22 final security subcases; M1 preview and checklist) (appended 2026-10-19 UTC)

**Owner/inputs.** Same r16 writer (Claude Opus 5.5, dispatched high; actual unknown/inherited), worktree
`agent-<id>`. Release: Codex "R16 runtime slice ACCEPT; final proof release" (review `3c47b84d…6826`).
Branch `d5d928ad` (asserted) rebased onto staging `813d2296` (delta: SUPERVISOR-HANDOFF, PROGRESS, this Result as
imported — blob identical to the branch's `cf1b8d8d` — and the G058 review; management-only). The Result-only rebase
conflicts were resolved with `-X theirs` and verified: the rebased tree differs from `d5d928ad` only by staging's three
other management files → `f899c839`. Bounded nested delegation (context limit; one writer at a time): one read-only
inventory agent, then one implementation agent writing only the five test files below with focused DB runs while this
writer held the harness lock and made no edits; integrated, reviewed and committed here (`839a0a4f`). Scope 53.

**Inventory (existing coverage reused, cited).** 17a `TestEnrollmentCatalog/CT-ENR-1_requests` approve + `enr9ApproveAdmit`;
17c `CT-ENR-1_requests` negatives (runtime INSERT 42501); 21a `TestEnrollmentPendingCredentialRevocation`
(`enrollResolves` after each revocation); 21c `enrAuditRefusals` (disabled user, fresh row, no resolve); 22a RemoveMember
`enr23StaleInvitation`, G013 superuser delete `CT-ENR-1_negatives`, deny `enr9DenyRelift`, revoke `enr9RevokeStale` /
`enrRequestRollback`; 22c owner depth-0 INSERT 42501 row-level security; 22d guard outside a trigger; 22h
`enrollment_lock_organization` rt 42501 + definer grant pin.

**Added (missing/partial only; extend existing tests, no new subtests).**
| Subcase | Test (file) | Assertion |
|---|---|---|
| 17b | `TestEnrollmentCatalog/CT-ENR-1_audit` (migrate) | rt in the IDENTITY shape (NULL `ctx_org()`) appending `enrollment.requested` / `access.denied` with `organization_id` set → 42501 from the guards `audit_event_source` / `audit_event_scope` (constraint names pinned); no audit row written |
| 17d | `TestEnrollmentAnswerShapes` (enrollment/api, unit) | applicant status envelope keys exactly `csrf_token,member,request,status`; no uid/org/host/user_id/organization_id/origin/address in the body (shared `fakeDB` gains an optional read-only `tx`; whole package rerun) |
| 21b | `TestEnrollmentPendingCredentialRevocation` "login start removes an expired enrollment credential only" (identity) | an expired, unrevoked credential has no revoke reason and resolves nothing |
| 21e | same subtest | the enrollment digest against `app.auth_resolve_session(…, now(), origin)` → 0 rows, while `auth_resolve_enrollment` resolves it (positive control) |
| 21d | `TestRLSIsolation` RLS-AUTH EMPTY block (access) | EMPTY shape: SELECT on `enrollment_sessions` 0 rows; DELETE by id of a live and of an expired row 0 rows (the SELECT policy applies); identity's unqualified expiry sweep deletes exactly the expired rows; both rows present after rollback |
| 22b | `CT-ENR-1_negatives` (migrate) | cw, pr, ar ledger INSERT → 42501 |
| 22e | `CT-ENR-1_negatives` | owner sees the departed row; owner UPDATE/DELETE affect 0 rows; row unchanged |
| 22f | `CT-ENR-1_requests` case "atomic without the owner policy" (closure, not a subtest) | with `enrollment_restrictions_owner_insert` dropped: runtime manager RemoveMember DELETE, the member's last-binding self-leave DELETE and the manager's denial UPDATE each 42501; after rollback `1|pending|0` (binding, pending, no ledger); positive controls with the policy present: each commits one row and writes exactly one trigger ledger row |
| 22g | `enr23StaleInvitation` (access, TestEnrollmentAdmission) | a subject with only A's Organization binding: facade `LeaveScope` commits, 0 bindings, exactly one unlifted `departed|trigger` row |

**Deviation.** Plan r3 §6.3 placed ENR-17/21/22 under `'^TestIsolationAttacks$/^(ISO-ATTEST|ENR)'`; that suite is pinned
at 22 subtests plus a 142-case oracle (run.sh `obj_it`), so new subtests there would need pin/oracle changes. The cases
live in the nearest existing tests above; every run.sh count pin is unchanged.

**Validation (Docker `golang:1.27.2-alpine`, read-only source and module cache, tmpfs GOCACHE, `-mod=readonly`,
`GOPROXY=off`; units `--network none`; DB batches on a disposable internal network with pgvector 0.8.7-pg17 on tmpfs,
generated env never printed; `docker-desktop-mounts.sh` 14712/100000; harness lock held throughout). Logs
`r16/enr/`.**
| Check | Exit | Counts | Log SHA-256 |
|---|---|---|---|
| gofmt (4 packages, again after fixes) / `go vet -tags integration` | 0 | clean | `vet*.log` empty |
| unit `^TestEnrollmentAnswerShapes$`; whole `./internal/enrollment/api/` | 0 / 0 | PASS / PASS | `83b26a38…dd0f` / `7c716c7c…69a7` |
| migrate `^TestEnrollmentCatalog$` run 1 | 1 | 18/19: new 17b assertion expected an RLS/source message; real refusers are the guard triggers (secure) → assertion pinned to the observed constraint names | `2e90e0bb…6f9b` |
| migrate `^TestEnrollmentCatalog$` rerun | 0 | **19/19** | `5ee865a4…ae29` |
| access `^(TestEnrollmentAdmission\|TestEnrollmentRequests)$` (fresh DB, migrate up 16) | 0 | **14 / 15** | `90e5dd3e…785e` |
| access `^TestRLSIsolation$` run 1 | 1 | 9/10: new 21d by-id delete expected 1 row; the SELECT policy hides the row in EMPTY (secure) → assertion corrected to the real sweep contract | `ce4626f2…3d0d` |
| access `^TestRLSIsolation$` rerun | 0 | **10/10** | `f4c37506…68f0` |
| identity `^TestEnrollmentPendingCredentialRevocation$` (fresh DB) | 0 | **8/8** | `78c36673…3943` |
Both first-run failures were wrong new assertions, not boundary failures; no existing assertion changed. Cleanup:
containers/network removed, no volumes, env files deleted, verified empty by label/name. Change-impact: test-only
delta; no production/SQL/run.sh change, so core/replay/browser/runtime proof is reused unchanged.

**M1 preparation (owner performs M1; not run by Claude).** Checklist `~/.local/state/platform-evidence/g058-a1/r16/
M1-checklist.md` `f33000ee…bcdea` (r0 §6.4 items + r3 additions: per-step user/host/expected; open signup → Group;
request → approve → revoke before return → re-request → admit; deny → closed notice → owner flips C to `open` → denied
applicant still refused (on C: B's owner has no sign-in principal); first-time invitation; owner/admin/member pages;
keyboard, error association, Light/Dark/System at 1280 and 320; privacy). Preview tooling: `r16/idn.sh` `0e627ab3…3611`
(+`preview-up`: proxy on 127.0.0.1:543 and the preview-only override `r16/preview.yml` `5e86c6e7…ea91` publishing the
emulator on 127.0.0.1:4100, because core's sign-in redirect targets `http://emulator:4100`; +`ca`: public CA only),
`r16/preview-smoke.sh` `33795128…928b`. Smoke on `f899c839`: org-a/b/c 200 with the run CA, auth host 404, redirect
host `http://emulator:4100`, emulator page 200 with the email field; then `down` (containers/volumes/images 0). Logins:
an owner-only command prints the disposable emulator's seeded email/password pairs in the owner's terminal (never the
API key; Claude never runs it). No stack left running.

**Not done.** M1 itself (owner), final local affected coverage, MR !50 checks to terminal, hosted qualification.
**Refusals.** One classifier denial: `git rev-parse HEAD HEAD:management/v0/results/G058-generic-organization-enrollment-result.md`
("Git Destructive") right after the rebase; not retried in that form (plain `git rev-parse HEAD` and a tree diff
verified the result). Nested worker: two worktree-guard refusals of compound commands, rerun as plain commands.

### Staging repair 1 — 2026-10-19

**Context.** G058 merged as `27ee7c20` (MR !50, reviewed source `4ec314b8`). Its required post-merge staging pipeline
2961169069 failed in job 17270568377 (`validate:integration`): 87 PASS / 3 FAIL / 6 INFO. Under the owner's
"Staging repair gate" (docs/agents/GITLAB-BRANCH-WORKFLOW.md), this is a single-writer repair on branch
`g058-staging-repair-1`. Target `origin/staging` = `baac320d`. Commits: `1c20ee0f` (F2-12), `5eaa71f4` (pins),
`b9febbb6` (JUnit timing), then this Result commit. Scope: 54-pattern allowlist (amendment07), 0 paths outside.

**Inventory completeness.** From the pipeline's results.tsv and junit.xml: there are no skipped cases. The non-PASS
rows are B-OBJ-IT, B-CI-CONTROLS and B-REQUIRED-CHECKS (derived), plus phase:harness/phase:complete, which are
derived from those rows. This was G058's first staging run, so no earlier failure is unresolved.

| Check | Failing run | Correction or disposition | Pre-merge proof (command, exit, counts, log SHA-256) |
|---|---|---|---|
| B-OBJ-IT (TestIsolationAttacks F2/F2-12; oracle 142 cases, 1 disagreement: observed `sqlstate 42501 invitation_context`, oracle `sqlstate 42501`) | 2961169069 / 17270568377 / `27ee7c20` | **Fixed.** `core/internal/tenancy/access/attack_expected_test.go`: F2-12 Effect now equals the hosted observation byte for byte. Source and the oracle comment cite D-15(i): the 00016 GRANT gives the runtime EXECUTE, so the Organization-context call reaches the body, and the 00015 context guard refuses it (`invitation_context`, insufficient_privilege). No other row, grant or guard changed. No-consumption/no-membership stays asserted: the F2 `membershipDigest` (role_bindings + invitations) is compared before and after, and the F2-24 control later consumes invOrg. | `repair1/attack-batch.sh` under `local-harness.lock`: fresh pgvector 0.8.7-pg17 + MinIO on an internal network, repo postgres/MinIO init, `go test -tags integration -run '^TestIsolationAttacks$' ./internal/tenancy/access/` (the run.sh obj_it selector), ro source and gomod, `-mod=readonly`, `GOPROXY=off`. Exit 0. 22/22 subtests, `ISO-ORACLE {"cases":142,"disagreements":0}`. `run1/attack.log` `f3791700…29f76a48`. Run on `808b0a66`; core/ is byte-identical at the final HEAD (`git diff 808b0a66 HEAD -- core` empty; oracle blob `18215cb3`). Cleanup verified: no containers or network left. |
| B-CI-CONTROLS (exit 1, 438 passing: VA7 `required_count is not 88`; VU1 `35/90/121, want 35/88/119`; J8 `90 check rows, want 88`) | same | **Fixed.** `infra/ci/validate-controls.sh` pins now take the hosted values (summary.json `required_count` 90): VA7 `sum_has required_count 90` plus its detail text, VU1 `vu_split full 35 90 121` (detail "35 + 90 -> 121"), J8 `n != 90`. Consumer inventory: a repo-wide `grep -w` for 88/119/86/117 found no other count consumer. The only hits are unrelated: contract fixture message numbers, a web layout px value, `boundary_test.go` 86 statements, and validate-controls 90/91/121, which are BuildKit vertex numbers and a context-line cap. The other summary pins (fast 35 at line 956, affected 20 at 1078, integration 35/44/75 in VU9) passed on the hosted run and are unaffected. profile-paths.tsv, README and OPERATIONS carry no 88/119 count. | Focused: `validate-controls.sh --only VA7,VU1,J8` exit 0, 3/3 (`ctl-pins.log` `4ffadecc…1aeb2e0a`). Suite: `flock -n local-harness.lock bash infra/validation/run.sh --base origin/staging --allow-file management/v0/results/G058-scope-allowlist.txt --report-dir …/repair1/ci-controls ci-controls` on HEAD `41720275`, the same code as `b9febbb6`, which is rebased over a management-only delta. Result: **B-CI-CONTROLS PASS, 441 passed / 0 failed of 441** (`ci-controls.log` `4448213c…404fe6a5`, `results.tsv` `aed85e6f…36d99bde`, run log `ci-controls-run.log` `2c2e6cf0…3cfb655a`). Wrapper rows in the same run: B-SCOPE FAIL and B-SECRET-CANARY UNMET. B-SCOPE failed because shared `origin/staging` advanced mid-run (17dae94d→baac320d, 9 management/skills files). That is an exact-delta artifact, cleared by the rebase: scope check 54 patterns, 0 outside. B-SECRET-CANARY needs the lint suite, which the narrow ci-controls suite does not run. Neither is a product defect. |
| B-REQUIRED-CHECKS | same | **Derived.** It aggregates the two rows above, and its inputs are now fixed. | Clears when B-OBJ-IT and B-CI-CONTROLS pass on the hosted run. |

**JUnit timing (owner item, released by amendment07).** The hosted integration JUnit summed 9565 s (overlapping
suite durations plus phase:harness 3830 s) for a 3934 s job, because GitLab sums case times. Correction:
- `infra/ci/junit.sh`: suite and phase durations are diagnostic only, never cases or case times. One explicitly
  named, always-passing case, `timing:validation-elapsed` (class `validate.JOB.timing`), carries the run's
  measured wall-clock elapsed seconds and replaces the former per-suite `suite:NAME` cases.
- `infra/ci/validate.sh`: `do_junit` writes the `elapsed validation $SECONDS` row only when executed-suite timing
  rows exist. JUNIT_SHAPE and the yq class check admit a time only on the timing case, and the self-check rejects
  more than one time attribute.
- Unchanged: failure cases and counts, and timings.tsv, which keeps the per-suite breakdown.
- `docs/foundation/OPERATIONS.md` documents the rule.

Proofs:
- J4b converter fixtures: overlapping suites 51 s + harness 40 s under 41 s elapsed give exactly one time, 41, on
  the timing case. Counts are pass 2/0, fail 3/2 and phase-only 2/1. Without an elapsed row there is no timing
  case. Invalid, repeated or misnamed rows return 19.
- J1/J5 through validate.sh: stub suites with overlapping timings (220 s summed) on pass and harness-fail runs give
  one self-checked timing case at 5 s and 3 s. The failure cases are unchanged.
- `validate-controls.sh --only J1,J2,J3,J4a,J4b,J4c,J5,J6,J7,J8`: exit 0, 10/10 (`ctl-junit2.log`
  `92524a67…5e10eea3`).
- The first J run failed J1 on byte determinism, before the design moved the elapsed time to the dedicated case.
- shellcheck v0.11.0 `-S warning -x` and `--include=SC2251` on validate.sh, junit.sh and validate-controls.sh:
  clean.
- The full 441/441 ci-controls run above includes these controls.

**Reversed placement work.** The local-only CI change (amendment06, since rescinded) was saved, not committed:
`repair1/local-only-ci-wip.patch` `00cad8b0…50686f80`. `.gitlab-ci.yml`, select-profile.sh and README are
byte-identical to staging. No full run was started.

**Evidence root.** `~/.local/state/platform-evidence/g058-a1/repair1/`. No secret values were printed, and env
files were deleted after use.

**Not done.** Hosted qualification of the repair (merge-gated); M1 (owner).

### Staging repair 1 — Duo correction (MR !52, 2026-10-19)

GitLab Duo requested two changes on MR !52 (head `e5ad8ee1`; MR fast pipeline 2961409283 success). Both fixed:
- `attack_expected_test.go`: added the header `r4` revision entry for F2-12 (G058 D-15(i) and the 00016 GRANT;
  updated by the repair author, not an independent oracle writer; F2-12 the only row changed). Comment only; no
  expectation, row or assertion changed, so the attack run above stands.
- `validate-controls.sh` `jtiming` (J1/J5): dropped the wall-clock `< 220 s` bound, which compared real runtime to
  the stub's summed timings and could fail on a slow runner. It now checks shape only (exactly one time value, on
  the timing case, no suite case). J4b keeps the fixed-elapsed non-additivity proof.

Proofs: `validate-controls.sh --only J1,J2,J3,J4a,J4b,J4c,J5,J6,J7,J8` exit 0, 10/10 (`ctl-junit3.log`
`5ac0009e…ff82c722`); shellcheck v0.11.0 `-S warning -x` on validate-controls.sh clean. Applied by the coordinator.

### G058 original-scope closeout record (appended 2026-10-19 UTC) — for Codex review, not self-accepted

**Authority and author.** The owner directs G058 to close on its original enrollment scope. Option B, the paused
Group-admin UI correction, is withdrawn and folded into G060 (review "Owner separates enrollment closeout from new
Group policy"; G060 Spec `21f4cf89…805e`). This record was written by the closeout writer, Claude Opus 5.5 (effort
inherited, actual value unknown). Worktree `agent-<id>`, branch `g058-closeout` from `origin/staging`
`70f61393`. Documentation only: no tests, builds, Docker commands or code changes, and nothing was deleted.
Pipeline statuses below were re-read through the read-only GitLab API while writing. Everything else is cited from
this Result, the review history and the coordinator checkpoint. Final ACCEPT belongs to Codex.

#### 1. Hosted qualification and source binding

| Pipeline | Ref / SHA | Jobs and outcome | Role |
|---|---|---|---|
| 2961066120, 2961087100, 2961101551 | MR !50 `71388d12`, `d0093ab8`, `d5d928ad` | success (validate:checks; integration manual, not run) | interim MR fast checks |
| 2961047560, 2961077071, 2961077103 | MR !50 `71388d12`, `694413f1` ×2 | skipped (no jobs) | none |
| **2961150120** | MR !50 `4ec314b8` (exact reviewed source) | success | MR !50 fast gate. Merged 00:43:51Z as **`27ee7c20`** |
| 2961169069 | staging push `27ee7c20` | checks 17270568376 success; **integration 17270568377 failed** (87 PASS / 3 FAIL / 6 INFO; `results.tsv` `25b92770…d292`) | first staging run; failed, see §3 |
| 2961400681 / 2961408831 | MR !52 `8851d4a1` | skipped / **canceled** | superseded, see §3 |
| 2961409283 | MR !52 `e5ad8ee1` | success | interim |
| **2961442501** | MR !52 `6c77b4f6` (exact reviewed source) | success | MR !52 fast gate. Duo approved; Codex delta ACCEPT. Merged 04:31:45Z as **`f7a28581`** (GitLab actor `<owner-account>`; the review records a Codex merge under standing authority) |
| **2961594265** | staging push **`f7a28581`** | validate:checks 17272937323 success (607.9 s); validate:integration 17272937324 success (3472.7 s); pipeline 4080 s. JUnit: integration 91/91 (90 checks + timing case, 3421 s), checks 36/36 | **qualifying hosted run** |

`f7a28581` contains `27ee7c20` (`git merge-base --is-ancestor` checked). Its tree is the reviewed repair `6c77b4f6`
plus management-only content. The local-only placement override was rescinded before the repair merged. G058 claims
no local full run. Later RC qualifications (0.1.0-rc.1, rc.2) are separate evidence and are not claimed here.

#### 2. Owner M1 provenance

- **Candidate.** Staging `7fafd1f490e910510a99d047fe0fb84bfe211eda`. It contains `f7a28581` and the RC1 limiter fix
  `3eec8f29` (ancestry checked). `src.tar` `822a583f…3fdd` is a git archive of that commit.
- **Preview receipts (coordinator-run).**
  - configure 0 (`.env` mode 600), start 0, migrate 0 (version 16).
  - Proxy 127.0.0.1:543; emulator 127.0.0.1:4100.
  - Seed 0: `mcp-seed.sql` `6b8f9efc…058d`, `emulator-seed.yaml` `ad97a633…7c31`. org-a `invite_only`, org-b
    `open`, org-c `request_access`.
  - CA `9c697cca…0ee7`, regenerated for this run.
  - Smoke: org-a/b/c 200, auth host 404, sign-in redirect host `http://emulator:4100`, emulator page 200 with the
    email field.
  - Harness lock held by task `bm5zwkyb1` for the preview's life.
- **Earlier previews (no M1 performed).** `4ec314b8` (CA `4e531851`) was stopped when the owner deferred M1 to the
  morning. `f7a28581` (CA `8eea5d6a`) was taken down at the owner's instruction. Both ended with 0 containers,
  volumes and images left, and the lock free.
- **Outcome is owner-reported, not independently re-executed.** The owner reported "everything works" around
  18:40Z. The coordinator did not watch each checklist step and infers nothing for steps the owner did not mention.
  The record `g058-a1/r16/M1-result-20261019.md` (`ca8225ad…0327`) was imported byte-exact as
  `results/G058-owner-M1-20261019.md`. Checklist: `r16/M1-checklist.md` `f33000ee…bcdea`.
- **Stop.** `idn.sh down`: stop-exit 0; containers 0, volumes 0, images 0. Lock holder `bm5zwkyb1` stopped; lock
  verified free. The record's "18:5xZ" is an invalid placeholder. **18:42:01Z** is the mtime of `idn-stop.log`
  (`18:42:01.54 +0000`, re-read here). It is observed completion evidence, not independent proof of the exact stop
  instant. No more precise time is asserted.

| Finding | Disposition |
|---|---|
| Obs 1: grace's A invitation opened while signed in as the seeded hostile account | Behaved correctly; not a finding. The read-only DB check showed the invitation consumed only by grace (`invitations_consumed_is_intended`). Optional UX (telling the other account the link is for someone else) is not released |
| Obs 2: frank (org_admin of A/C, no Group binding) gets 404 on A Group members pages | Triage (`plans/G058-M1-group-admin-404-correction-20261019.md`, `a3c45729…97d8`) traced it to `members-view` needing self-membership, while core policy grants org-admin Group read and manage. Option B was accepted, then put on HOLD, then **withdrawn**. Carried to **G060** (§5). Not an enrollment failure |
| Req 3: Group owners/admins inviting new people into the Organization and the Group | New scope, not an M1 failure. G060 Spec keeps combined admission separately scoped; nothing is released |
| Follow-up 4: styled privacy-preserving 404 page | Possible later bounded presentation follow-up. Not released. It must not hide an access bug |
| RC1 Duo limiter LOW (rate-cache rollover) | Owner disposition: BACKLOG item `2026-10-19-enrollment-rate-cache-rollover-hardening-rc1-follow-up`. No fix or PASS claimed |

#### 3. Failed and canceled runs

**Hosted.**
- 2961169069 / 17270568377 on `27ee7c20` failed on B-OBJ-IT (F2-12 oracle), B-CI-CONTROLS (VA7/VU1/J8 count pins
  88/119 → 90/121) and B-REQUIRED-CHECKS (derived). Fixed by Staging repair 1 (MR !52; Codex source ACCEPT `8851d4a1`
  and delta ACCEPT `6c77b4f6`). Qualified by 2961594265. The failed run stays recorded as failed.
- 2961408831 on `8851d4a1` was canceled by MR supersession after the management-only rebase to `e5ad8ee1`. Its
  successor 2961409283 succeeded.
- No other MR !50, MR !52 or staging pipeline for G058 failed or was canceled.

**Local.** Every failed local run below was a first-run failure, corrected and rerun green within its round. No
existing security assertion was weakened.

| Round (Result section) | Failed run | Disposition |
|---|---|---|
| 1a-2 | MEM-10 (S4, S5): wrong replay expectation | Assertion corrected to the G019 `already_member` body; S6 18/18 |
| 1a-5 | ENR-INTENT rerun 0/1; logrecord setup (x/text absent from the read-only cache) | Fixed; 1/1. Cache prepared; 3/3. The negative-control revert was refused by the classifier and never executed |
| 1a-6 | migrate runs 1–4 (plpgsql parse, postcondition text, boolean format, function order) | Run 5 fresh: 9/9 + 10/10 |
| 1a-7 | gofmt alignment; 2 unit pins (`TestSQLBoundary`, `TestRLSBoundary`) | Fixed; 3/3 packages; AUDIT-INV 42/42 |
| 1a-9 (c1) | migrate boolean pins / Lifecycle env; AT-INV 45/42 | Rerun 3/3. AT-INV was a declared known mismatch until c2: 45/45 in 1a-10 |
| 1a-10 (c2) | gofmt; `TestAdmissionOutcomeMapping` row; `TestTenancyFixturePrimitives` | Fixed; reruns green |
| 1a-12 | Catalog ENR-16f (test-role bug) 18/19 | Fixed; 19/19 (the round was also a re-attestation CORRECTION LOOP, accepted R12) |
| 1a-13 | gofmt alignment | Fixed |
| 1a-14 / 1a-15 / 1a-16 | h1 eslint warning; B-SCOPE / B-REQUIRED-CHECKS exact-delta failures from staging management drift; `root` exit 1 on the same artifact; registry-proxy list exit 17 (no test ran) | Fixed or cleared by rebase (0 outside scope); retry listed 8/8 |
| 1a-16 | E-ENR-3 failed in both themes (G019 parser) | Amendment05 parser fix; E-ENR-3 2/2 |
| 1a-17 | Catalog 17b 18/19; RLS 21d 9/10 (wrong new assertions) | Pinned to the observed secure refusers; 19/19 and 10/10 |
| Repair 1 | J1 byte determinism; wrapper B-SCOPE (target drift) and B-SECRET-CANARY UNMET (lint not in suite) | Design moved to the timing case, 10/10. The two wrapper rows stay explicit, with no green-run claim |

#### 4. Cleanup

**Completed (from records, and verified here where stated).**
- All round-writer worktrees are removed. `git worktree list` now shows only the three G058 entries listed under
  Retained and Proposed deletion.
- Remote temporary branches `g058-generic-organization-enrollment` and the repair source branch were deleted at
  merge (ls-remote shows them absent).
- All three preview stacks are down (0/0/0) and the lock is free.
- Generated env files were deleted after each DB batch.
- No `platform-secrets/g058*` exists.
- Private files are in the home-local root `<home>/.local/state/platform-evidence/`, not the repository
  `.local`.

**Retained, with reasons.**

| Path / resource | Reason |
|---|---|
| Branch `g058-group-admin-404` (local and remote `5f94a705`) and worktree `.claude/worktrees/agent-<id>` | G060 planner reuse disposition. Spec: fold only after design review, no wholesale cherry-pick |
| `g058-a1/r16/{idn.sh,preview.yml,preview-smoke.sh,lookup.js,M1-checklist.md}` plus `scan-in.sh` (an `idn.sh` dependency), 52 KiB including the two records below | Preview tooling for G060/M1 reuse |
| `g058-a1/r16/{M1-result-20261019.md,group-404-triage-20261019.md}` | Raw owner-feedback and triage provenance; byte-identical copies are tracked in the repo |
| This worktree `agent-<id>` / `g058-closeout` | Until Codex imports this record |
| G055, RC, G060-plan resources (`agent-ae071c53…`, `agent-ad080d71…`, `agent-a1f0e89b…`, `g055-plan-*`, `rc-release-20261019/`) | Active or other-task owned; not G058 |

**Proposed deletion (Codex decides; nothing performed).** Sizes are `du -sk`.

| Path | Size | Note |
|---|---|---|
| `platform-evidence/g058-a1/` except the retained r16 files above | 78,772 KiB (of 78,824) | Includes `r16/src.tar` 31,932 KiB (reproducible `git archive 7fafd1f4`), `r2db/src` 33,364 KiB, `r16/a1` 8.6 MiB, `repair1/` 340 KiB, and the r2–r15 / r16 logs, reports, `preview-*.pem` and messages that the Result hashes cite |
| `platform-evidence/g058-staging-2961169069/` | 64 KiB | `results.tsv` `25b92770…`, cited by the review |
| `platform-evidence/g058-review-r1` … `-r5` | 12 + 12 + 16 + 12 + 12 KiB | `tmp/handoff.txt` scratch |
| `platform-evidence/g058-plan/`, `g058-premerge/` | 8 KiB, 12 KiB | Commit message and MR ready/merge JSON |
| `/tmp/{g058_00015.sql,g058r2-00015.sql,g058-addendum-commit-msg.txt,g058-r8-scope.txt,g058c2-scope.txt}` | 188 KiB total | Planning scratch |
| Worktree `.claude/worktrees/agent-<id>` (locked; `g058-staging-repair-1` at `8851d4a1`) | — | Superseded by merged `6c77b4f6`. Unlock, check dirty state (not inspected; worktree guard), then remove |
| Local branches `g058-generic-organization-enrollment`, `g058-staging-repair-1-rebase` (both merged into staging) | — | Remove |
| Local branches `g058-staging-repair-1` (`8851d4a1`), `g058-a1-r8-request-slice-wip` (`7ebe584d`, parked WIP superseded by c1) | — | Bundle first, then remove |
| Local and remote `g058-plan`, `g058-plan-r1` … `-r5`, `g058-readiness-r0` | — | Content imported under `management/v0/plans/`. Bundle first, then remove |
| Placeholder branches `worktree-agent-<id>` and other G058 writers' placeholders (main-merge tips, no unique commits) | — | Optional; repository-wide placeholder hygiene is not G058-specific |
| Docker volume `g058-a1-gomod` (last recorded 2026-10-18; not re-inspected because this pass ran no Docker) | — | Verify, then remove if present and unused |

#### 5. Pending G060 carry-forward

- **Group authority and the members-page UI** move to G060: an Organization owner/admin without a Group binding
  currently gets a UI 404 while core grants org-admin Group read and manage. That 404 is current behavior, not an
  enforced policy. The server-side authority stays as implemented until G060 changes it.
- **Option B is withdrawn: not run, NOT PASS.** No test was executed on `5f94a705`, and the G058 Result was not
  changed by that writer. The option-B supersession of the E-MEM-5 oracle is withdrawn with it; original E-MEM-5 is
  historical evidence for G060 to reconcile.
- **Test-lock native refusal stays blocked.** The classifier denied a background `flock -n … sleep 14400` as
  [Interfere With Workloads], and the worktree guard refused a compound flock+run.sh command. Neither was routed
  around. G060 needs a supported rule or ownership pattern before it runs locked checks.
- **Holds.** G055 integration and completion wait for G060 final ACCEPT. G020 planning and OS001 execution stay held.
  Closing G058 releases no downstream task.
