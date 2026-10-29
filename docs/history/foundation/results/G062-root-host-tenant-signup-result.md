> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G062 — Root-host tenant provisioning signup: Result (in progress)

Goal: [G062](../goals/G062-root-host-tenant-signup.md) · Spec: [G062](../specs/G062-root-host-tenant-signup-spec.md) ·
Plan r5 (accepted): [G062 Plan](../plans/G062-root-host-tenant-signup-plan.md) · Review/release:
G062-plan-codex-review.md (not included).

Status: **attempt 1; slice 1 merged to staging as `c5688bf94e1fa4080145f5f6de22005047853ec2` and qualified by the exact
merged-SHA full local run; slice 2 execution released by Codex (staging `50982d14`, see round 13); slice 2 round A (core
API) implemented on `g062-impl2` with stack-free tests (round 14), rounds B/C and their proofs pending; final G062 ACCEPT
held.** Not self-accepted.

## 1. Bindings and rebinding (recorded before the first write)

- Writer: Opus 5.5 high (runtime effort not observable from inside the agent), single write owner; worktree
  `.claude/worktrees/agent-<id>`, branch `g062-impl` created from `origin/staging` after `git fetch origin
  staging`; upstream tracking removed so a push can only target `g062-impl`.
- **Immutable original target SHA:** `origin/staging` = `51e8c57c51ea709bc62a6eedc03843b1aba28ce8` (rebound from the R0′
  release head `8a66ae51`; management-only delta). Source/Plan baseline `47df73f5`: `git diff --stat 47df73f5 HEAD --
  core web infra contracts` is empty at `51e8c57c` (code trees byte-identical).
- SHA-256 verified at `51e8c57c` (all match the release record):

| File | SHA-256 |
|---|---|
| `management/v0/plans/G062-root-host-tenant-signup-plan.md` (r5) | `b2a612a2c50b4a36505346c2c18214cd785d7d62277bdb35c993db49a3a5a3e2` |
| `management/v0/plans/G062-planning-inputs.json` | `f89b44ed7ee9d50fab464029cc838c964b2727f0b2f0cbcac665db8ef5a0549d` |
| `management/v0/results/G062-scope-allowlist.txt` (40 ERE patterns) | `7f279a42c97256b5b1abffafd77a770c1e6d2effaf3b8717de96b9263dea073f` |
| `management/v0/goals/G062-root-host-tenant-signup.md` | `e747807c83875d8928b8c505fff83a4a2fe65d716b15d5ed64f24eb72560118a` |
| `management/v0/specs/G062-root-host-tenant-signup-spec.md` | `3c90ccb6385768e598d0cc61afdbf3b33d79999b94217e01988c2a2316154bb3` |
| `management/v0/results/G062-plan-codex-review.md` | `41c47517d27cb8ac60e3b184b6dd11c0b53ca0c399c95f679373bdec5397030c` |
| `management/v0/PROGRESS.md` (rebound at `51e8c57c`) | `fb891dd9de57891e8418f9ac2f46f7d72c5071149ad94b20b7050c4d3af3978b` |
| `management/v0/results/G062-r0prime-readiness-20261020.md` | `8602add90098cc78c73889687b62fea71000fe7310682b22060503bec118a5c5` |

- Mandatory r5 ACCEPT clarification adopted: the `provision_writer` SELECT/UPDATE policies on `app.signup_sessions`
  admit the active user's **pending** row (first create) and **provisioned** rows (retry/continuation); body checks
  (IDENTITY own-credential; LOGIN exact reference/origin/Organization), the proof-reset WITH CHECK and the restricted
  column grants are kept; no runtime access is broadened.

- **Later management-only rebinding (coordinator notice):** staging advanced to `8a35f5c8` (management-only; owner RC
  qualification direction). `management/v0/PROGRESS.md` is now
  `1219f22477177e77443fe46002c1d56b4388ae2d45e800172ce3af09a870b9ef` (was `fb891dd9…` at `51e8c57c`). Code, Goal, Spec,
  Plan, inputs and scope bindings are unchanged; the immutable original target stays `51e8c57c`.

## 2. Attempt accounting

- B5, checkpoint after attempt 3. Attempts before this run: 0. **This run: attempt 1, slice 1 only.**

## 3. Slice 1 work log

### Round 1 (attempt 1): role bootstrap, migration 00018, migration-set and reset parity

Files (all match the 40-pattern allow-list at `7f279a42…`; `grep -vEf <patterns>` over `git diff --name-only 51e8c57c`
is empty):

| Path | Change |
|---|---|
| `infra/postgres/init/10-roles.sh` | fourth helper `<db>_provision_writer` (NOLOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION), owner its only member `INHERIT FALSE, SET TRUE`; header |
| `infra/postgres/provision-roles.sql` | `provision_writer_role`; unsafe-attribute and membership allow-lists; `mk_provision_writer`/`g_provision_writer` branches; header (a pre-G062 volume gains only the fourth helper) |
| `infra/validation/run.sh` | PROV-1 admin-residual string, PROV-6 delta string, B-RLS-MIGRATION PASS text (four helpers), B-DBRENAME role list/text and `current_version 18` |
| `core/migrations/00018_platform_signup.sql` | new (Plan r5 §3): precondition (four helpers, version 17); `login_transactions` signup columns/CHECKs and the four-member return-path CHECK; `app.signup_sessions`; runtime, auth-reader and provision-writer grants/policies (r5 clarification: the pw SELECT/UPDATE policies admit any row of an **active** user, pending or provisioned; proof WITH CHECK kept); 3 pw + 1 ar DEFINER bodies; audit delta (reasons 70→79, operations 65→73) and `events_attribute` (00017 body plus the signup clauses); `events_insert_provision`; postconditions; down (history-preserving refusal; restores the 00017 body/CHECKs verbatim) |
| `core/migrations/migrations_test.go` | `TestMigrationSet` 00001–00018 |
| `core/internal/migrate/integration_test.go` | `reset`/`rlsReset` drop the 00018 objects (signup table first; pw bodies as pw; `auth_resolve_signup` as ar; pw schema USAGE) |
| this Result | bindings, log, checkpoint |

Generation note: 00018's audit-constraint and `events_attribute` texts were produced from the accepted 00017 text (up:
targeted insertions; down: verbatim) by a private generator under `~/.local/state/platform-evidence/g062-a1/gen/` (not
committed). The committed SQL is the reviewable artifact.

Commands. Pre-batch: `git fetch origin staging` → `51e8c57c`; ancestor check exit 0; non-management delta empty, so no
rebase. `bash infra/ci/docker-desktop-mounts.sh` exit 0 (17%). Containers use the existing `golang:1.27.2-alpine`, the
existing module volume read-only, `GOFLAGS=-mod=readonly GOPROXY=off`, a disposable `GOCACHE` and source copied with
`docker cp` (no bind mount).

| # | Command | Exit |
|---|---|---|
| 1 | unit (`--network none`): `go test -count=1 -run '^TestMigrationSet$' ./migrations/` | 0 |
| 2 | unit: `gofmt -l ./migrations ./internal/migrate` (empty) + `go vet -tags integration ./internal/migrate/ ./migrations/` | 0 |
| 3 | private DB (pool slot 2; existing pinned `pgvector/pgvector:0.8.7-pg17`; internal network `platform-test-g062-net`; no published ports): fresh-volume `infra/postgres/init/10-roles.sh` → 8 roles, `<db>_provision_writer` NOLOGIN/NOINHERIT/safe attributes, owner its only member (inherit f, set t, admin f) | 0 |
| 4 | same DB: repeat `provision-roles.sql` as superuser (no refusal, nothing changed) | 0 |
| 5 | `go run ./cmd/migrate up` → `current_version 18` (00018 postconditions pass) | 0 |
| 6 | `go run ./cmd/migrate down` → 17, then `up` → 18 (reapply) | 0 / 0 |
| 7 | focused SQL probe at version 18, rolled back: 18/18 assertions pass (list below) | 0 |
| 8 | after `down`: `md5(prosrc)` of `audit.events_attribute` equals the md5 of the 00017 body text (`bd213b18…`); no signup/provision policy remains; pw has no `app` USAGE | 0 |

Probe 7 assertions: runtime org INSERT 42501; **pending-row first create** `created` g1; identical retry `already_created`
g2; changed payload `intent_used`; runtime UPDATE of `ref_hash`/`completed_at` 42501; TENANT shape refused (42501
`signup_context`); challenge issue: superseded R1 and a foreign Organization refused, R2 once, replay refused; prove under
another user's context matches nothing, own prove updates 1 row; pw non-null `proof_hash` 42501 (WITH CHECK); pw
`login_transactions` INSERT 42501; complete: wrong P refused with nothing spent, C+P once, replay refused; **re-mint after
completion** `already_created` g3 with every continuation column NULL; exactly one org, one `org_owner`, one
`organization.created` and one `membership.granted`, all attributed to the creator; second user, same slug `slug_taken`;
foreign credential `signup_invalid`; disabled user `user_disabled`; `signup_handoff` without `expected_user_id` 23514;
`/signup/details` admitted; `/signup`, `/signup/details/`, `//evil`, `https://evil`, `/%2Fevil`, `/signup/../account`
23514.

Resource receipt: containers `platform-test-g062-{pg,it,unit}` removed with `docker rm -f -v` (the PG anonymous volume
`ccbe0bad…` went with it; verified absent); network `platform-test-g062-net` removed; the label filter
`platform-test=g062-a1` lists nothing; the mode-600 credential env file is deleted. No image was built or pulled. No
`local-harness.lock`, Compose stack or published port was used. The shared module volume was mounted read-only only.

Deferred (these need the harness, the core-check image or more slice-1 code):
- `bash infra/validation/run.sh migrations` rows B-RLS-MIGRATION (PROV-1..6 incl. the four-helper delta), B-DBRENAME
  and B-MIGRATE-SET/CLI/CATALOG. The last three still pin version 17 / 71 relations and need the mechanical 18/new-count
  parity edit (take the count from a version-18 catalog). B-PERSIST, B-TOGGLE-OFF, the B-RECOVERY-DRILL setup,
  B-IDENTITY-CATALOG and the "at version 17" PASS texts need the same edit.
- `go test -tags integration` suites whose version-17 pins move with 00018 (`metadataThroughVersion17` callers,
  `TestMigrationLifecycle`, `TestRLSCatalog` (`rlsLabel` gains `pw`), the `TestMembershipCatalog` role list,
  `TestAuditCatalog`/`TestCatalogParity` mig18). These parity edits are not made yet.

### Checkpoint (context ~150k): transfer state for the next writer

Remaining slice-1 items (Plan §7 step 2), none started:
- `TestSignupCatalog` (`core/internal/migrate/signup_integration_test.go`) and the migrate-package version-18 parity above.
- `core/internal/audit/audit.go` catalog delta and `TestCatalogParity` mig18.
- `core/internal/signup/{signup,store,slug,names_copy,errors,log}.go` and their tests (`TestSignupProvisioning` SUP-7
  a/b/c and concurrent retries, `TestSignupSlug`/`TestSignupSlugRace`, `TestSignupContinuation`).
- identity: `LoginInput.SignupIntent`, `platformRootOrigin`, `loginTxn.{SignupIntent,SignupHandoff,ExpectedUserID}`, the
  callback T1 branch, `LoginPinned`, the pinned-callback user/Organization checks and the `tenant_created` outcome,
  `returnPaths` + `/signup/details`, `SignupCredential`, revocation in the `store.go` logout/logout-all/disable paths
  (next to the enrollment revocations, ~lines 706/900/927/1028), the expired sweep next to `sqlExpireEnrollmentSessions`
  (EMPTY shape, `DELETE FROM app.signup_sessions`), and the root-callback limiter.
- `infra/recovery/sql/invalidate.sql` fold and the `TestRecoveryPost` subcase.

Scope note: `identity/retire.go` (FinalizeRetirement) also revokes enrollment credentials but is **not** allow-listed.
The Plan names only logout, logout-all and disable, so it stays unchanged unless Codex widens scope.

Contract facts fixed by round 1 for the Go code:
- Provisioning: `app.signup_provision_organization(cred_hash, name, slug, intent_hash, ref_hash, target_origin, now)` in
  the IDENTITY shape returns `(outcome, organization_id, slug, generation)`. `outcome` is one of `created`,
  `already_created`, `intent_used`, `slug_taken`, `user_disabled`, `signup_invalid`. Grammar/reserved violations surface
  as SQLSTATE 23514 (Go validates first).
- Challenge and complete need the LOGIN shape (`app.login_binding` set, no user or org).
- Step-3 prove is runtime SQL on `nonce_consumed_at, proof_hash` under `signup_sessions_prove`.
- T1 inserts `(id, user_id, origin, credential_hash, csrf_token, provider_session_id, created_at, expires_at)` in the
  IDENTITY shape.
- `auth_resolve_signup(hash, now)` returns session_id, user_id, origin, csrf_token, expires_at, provider_session_id,
  organization_id, slug, target_origin, provisioned_at, continuation_generation, completed_at.

Deviations: none from the Plan's SQL contract. Two readings are recorded for review:
1. pw `SELECT (id, slug)` on organizations is granted as Plan §3.3 lists, but no pw SELECT policy on organizations is
   created (§3.5 lists none). The grant is therefore inert, and `already_created` returns the intent-bound slug.
2. `role_bindings_insert_owner` also pins `scope_id = organization_id` and `ctx_org() IS NULL` (narrower than §3.5).

### Round 2 (attempt 1, continuation writer): audit delta, identity T1/pinned callback/LoginPinned, revocation, limiter, recovery fold

Continuation bindings: second writer (Opus 5.5 high; runtime effort not observable), single write owner, worktree
`.claude/worktrees/agent-<id>`, branch `g062-impl` switched from `origin/g062-impl`, HEAD asserted
`f1d29cf3ac54db94712dbe11a39380f497586650` before any edit. Plan, inputs, allow-list, Goal and Spec SHA-256 re-verified
from `origin/staging` (all match §1). Original target unchanged (`51e8c57c`, ancestor of HEAD). **Later management-only
rebinding:** staging advanced `8a35f5c8` → `8f0396bb` → `ca725c8c` (certifi review records, secret-scan ADR-007
proposal/review, PROGRESS; no code, Goal/Spec/Plan/inputs/scope change; certifi MR !64 not merged at this point);
`management/v0/PROGRESS.md` is now `587c10169e04df2a3b353f091bd262da5b58c809f3c6fa003fd199ef50854438`. No rebase (the
delta is management-only). Round-1 passes are reused and attributed to `f1d29cf3`; none of the files they exercised
(roles, 00018, migrations_test, migrate reset helpers) changed in this round.

Files (all match the 40-pattern allow-list; `git diff --name-only 51e8c57c` filtered by the patterns is empty):

| Path | Change |
|---|---|
| `core/internal/audit/audit.go` | 9 reasons (79) and 8 signup operations (73), Plan §3.6 |
| `core/internal/audit/audit_test.go` | `TestCatalogParity` reads `mig18` (actions, reasons, operations, trigger-owned); 79/73; the 00017 lists stay contained |
| `core/internal/identity/identity.go` | errors `signup_origin_rejected` 403, `signup_user_mismatch` 403, `continuation_invalid` 404, `rate_limited` 429; `SignupHandle` (unexported fields, minted only by `SignupCredential`), `SignupCredentialInfo`, `SignupOutcome`, `SignupLifetime`; outcomes `signup`, `tenant_created` |
| `core/internal/identity/allowlist.go` (+`_test`) | `returnPaths` = `/`, `/account`, `/signup/details`; `TestReturnPathAllowList` extended (positive + 13 look-alike negatives) |
| `core/internal/identity/service.go` | `Config.RootOrigin`; `LoginInput.SignupIntent` (exact root origin only, no enrollment facts, no credential check, return path `/signup/details`, root `/auth/callback`); callback: signup rows checked against the root at consumption (`signup_origin_rejected`), T1 branch (`signupCallback`), pinned handoff (`ExpectedUser`); `signInFailure` gains the two denials; `SignupCredential` (`auth_resolve_signup`, root origin only), `SignupCSRFMatches`, `SignupIdentityTx` (IDENTITY of a resolved handle), `SignupContinuationTx` (LOGIN shape + runtime host resolution, hands fn the request Organization), `LoginPinned` (tx 1: `signup_challenge_complete`; tx 2: `insertLogin` with re-resolution, runtime `is_organization_member`, pinned row `signup_handoff`/`expected_user_id`/`return_path '/'`), `RecordSignupDenial` (closed signup ops/reasons; handle-attributed) |
| `core/internal/identity/store.go` | `loginTxn` signup fields; `insertLogin` sweeps `DELETE FROM app.signup_sessions` (EMPTY, after the enrollment sweep) and inserts the three columns; `consumeLogin` returns them; `signInParams.ExpectedUser` (mismatch refused before any write; success outcome/audit reason `tenant_created`); bound signup revocation in `revokeSession` and supersession, all-rows revocation in `revokeAll` and `DisableUser`; `signupT1` (namespace binding, E2 mapping or `createMapping`, active check FOR SHARE, signup row, `auth.sign_in` `signup_authenticated` in the same tx); `signupIdentityTx`, `loginTx` |
| `core/internal/identity/http.go` (+`_test`) | intent `signup` (no enrollment facts); root-start global bucket (2/s, 20); root-origin callback per-binding (1/s, 5; key `sha256("signup/callback/"+binding)`) and global (5/s, 50) buckets, both always charged; 202 `{"signup":{"credential":{value,max_age},"next":"/signup/details"}}`; `TestRootCallbackRateLimited` (start bucket, per-binding, churn vs global, tenant callbacks unlimited, signup answer shape, enrollment facts refused) |
| `core/internal/identity/limiter_copy.go` (+`_test`) | byte-pinned copy of the `enrollment/api` limiter (generated from the source file; `TestLimiterCopyIsExact`, `TestLimiterBounds`) |
| `core/internal/tenancy/access/boundary_test.go` | mechanical pin parity: 00018 DEFINER allowance (exactly the four bodies), `signupT1`/`signupIdentityTx`/`loginTx` in the identity init inventory, `createMapping` callers `signIn,signupT1`, `insertLogin` call order with the signup sweep, `sqlExpireSignupSessions` text pinned |
| `infra/recovery/sql/invalidate.sql` | fold: unrevoked signup rows expire now (`expires_at`, and `continuation_expires_at` only when non-NULL), count `signup_sessions_expired`, postcondition extended. **Mechanical correction to the Plan §2 text:** `least(NULL, now())` is `now()` in PostgreSQL, which would violate `signup_sessions_generation_valid` on generation-0 rows, so the continuation expiry keeps NULL |

Commands (pre-batch: `git fetch origin staging`; ancestor check exit 0; non-management delta empty; `bash
infra/ci/docker-desktop-mounts.sh` exit 0, 17%). Unit containers: existing `golang:1.27.2-alpine`, `--network none`,
module volume `g058-a1-gomod` read-only, `GOFLAGS=-mod=readonly GOPROXY=off`, disposable `GOCACHE`, source by `docker cp`,
label `platform-test=g062-a1`; script `~/.local/state/platform-evidence/g062-a1/gen/unit.sh` (gofmt -l, vet, vet
`-tags integration`, `go test -count=1`).

| # | Command | Exit |
|---|---|---|
| 9 | unit: identity + audit (gofmt, vet, vet `-tags integration` incl. `tenancy/access`, `migrate`; tests) | 3 (gofmt: http.go, identity.go) → `gofmt -w` in a disposable container, copied back |
| 10 | same, rerun | 0 (`identity` ok, `audit` ok) |
| 11 | unit: `tenancy` + `tenancy/access` | 1: `TestRLSBoundary` (00018 DEFINER allowance; new identity init helpers; `createMapping` callers; `insertLogin` order; `RecordSignupDenial(userID)` violated "no bootstrap API accepts a user UUID") |
| 12 | after the boundary pin parity and `RecordSignupDenial(SignupHandle)`: identity + `tenancy/access` | 0 |

Resource receipt (round 2): containers `platform-test-g062-unit` (four runs) and `platform-test-g062-fmt` removed with
`docker rm -f -v`; `docker ps -a --filter label=platform-test=g062-a1` lists nothing. No database run, no image build or
pull, no `local-harness.lock`, Compose stack or published port. The module volume was mounted read-only only.

Review notes (round 2, for Codex; no redesign):
3. `identity/retire.go` (`finalizeRetirement`) revokes enrollment credentials but not signup credentials; it is not
   allow-listed and the Plan names only logout/logout-all/disable, so it is unchanged. Signup rows of a retired user
   still expire within 30 minutes and are refused for a disabled user by `auth_resolve_signup` and every body.
4. Signup revocation is also added on supersession (beside the enrollment bound revocation in `signIn`), which the
   previous transfer state listed (`store.go` ~706) though the Plan prose names logout/logout-all/disable.
5. Because `set_config` may appear only in `identity/store.go` and `tenancy/store.go` (`TestRLSBoundary`), identity owns
   every signup transaction shape: `signup` will call `SignupIdentityTx`/`SignupContinuationTx` with its own SQL, and the
   completion SQL lives in `identity.LoginPinned`.
6. `core/cmd/server/main.go` does not yet set `identity.Config.RootOrigin`, so root signup is disabled (fail closed)
   until it is wired (with `signup/api`, slice 2, or as a one-line slice-1 addition at the next writer's choice).

### Checkpoint 2 (context ~170k): transfer state for the next writer

Done and unit-green at this commit: audit delta, identity signup intent/T1/pinned callback/`LoginPinned`/credential
resolution/revocation/sweep/limiters, `invalidate.sql` fold. **Not yet exercised against a database** (all integration
proofs are still open).

Remaining slice-1 items (Plan §7 step 2):
- `core/internal/signup/{signup,store,slug,names_copy,errors,log}.go` (no `tenancy` import): `Provision` (validate name via
  the pinned `tenancy.ValidateName` copy, slug via `hostname.SlugGrammar` + `hostname.IsReservedLabel(cfg)`; `intent_hash =
  SHA-256(name‖0x00‖slug)`; 256-bit R; `target_origin = hostname.Host{Name: slug+"."+base, Port: port}.Origin()`; inside
  `SignupIdentityTx`: `SELECT outcome, organization_id::text, slug, generation FROM app.signup_provision_organization($1..$7)`
  with the handle `Digest()`), `Challenge` (`SignupContinuationTx` → `SELECT ok FROM app.signup_challenge_issue(H(R), origin,
  org, H(C), H(N), now)`; returns C and `<root>/signup/prove?n=N`), `Prove` (`SignupIdentityTx` → the runtime UPDATE of Plan
  §3.4 step 3 with `user_id = app.ctx_user()`; zero rows `signup_context_mismatch`; returns `<target>/signup/complete?p=P`,
  target from `SignupCredentialInfo.TargetOrigin`), `Complete` (`identity.LoginPinned` with H(C), H(P)), `Abandon` (runtime
  revoke own row), `SuggestSlug` (advisory; candidates per the copied `suggestBase`/`slugCandidates`, availability via
  runtime `app.hostname_state('platform', slug)`), `RecordSignupDenial` on refusals; `signup.request` log event (check
  whether `infra/validation/tools/internal/logrecord/events.json` must list it before adding the name).
- Tests: `TestSignupCatalog` (`core/internal/migrate/signup_integration_test.go`; pending vs provisioned pw policy rows,
  14-column grant, non-grants, return-path CHECK, TENANT refusal, 00018 precondition without the role) and the migrate
  version-18 parity (round-1 deferred list); `TestSignupProvisioning` (SUP-7 a/b/c, concurrent retries, disable race),
  `TestSignupSlug`/`TestSignupSlugRace`, `TestSignupContinuation` (SUP-8), identity `TestCallbackSignup`, `TestCallbackPinned`
  (fake WorkOS in `identitytest`), revocation (SUP-10) and the recovery seed/verify subcase (old-auth signup credential
  refused after `invalidate-auth`).
- Then the private-DB batch (pool slot 2) for those integration selectors.

### Round 3 (attempt 1, writer #3): `signup` package, provisioning/slug/continuation integration proofs

Bindings: third writer (Opus 5.5 high; runtime effort not observable), single write owner, worktree
`.claude/worktrees/agent-<id>`, branch `g062-impl-w3` from `origin/g062-impl`, HEAD asserted
`dcaa6d9f4b8f7b5ac161d71004215dba7b27a66a` and `51e8c57c` ancestor before any edit. Plan r5 and allow-list SHA-256
re-verified (`b2a612a2…`, `7f279a42…`). Pre-batch: `git fetch origin staging` → `ca725c8c`; `git diff --stat 51e8c57c
ca725c8c -- core web infra contracts` empty (management-only), so no rebase; `bash infra/ci/docker-desktop-mounts.sh`
exit 0 (17%). Rounds 1-2 passes are reused; none of their covered files changed in this round (only new files under
`core/internal/signup/` were added). Resource scheduling follows the coordinator relay of Codex handoff c3401fdd (private
per-run stacks, two-run pool, no builds, no harness lock); `docker ps -a --filter name=platform-test` showed no live
private run before launch, so this run held one slot.

Files (all new; all match the allow-list; `git diff --name-only 51e8c57c` filtered by the 40 patterns is empty):

| Path | Change |
|---|---|
| `core/internal/signup/signup.go` | `Service` (`Config{Identity, Hostname, RootOrigin, Now}`; no tenancy import, no actor/role/tenant parameter): `Resolve` (identity `SignupCredential`), `StatusOf`, `SuggestSlug` (pinned base/candidates, reserved labels skipped, availability via runtime `app.hostname_state` in the credential's IDENTITY shape), `Provision` (pinned name rule; `hostname.SlugGrammar` + `IsReservedLabel(cfg)`; `Intent = SHA-256(name‖0x00‖slug)`; 256-bit R; target origin from configuration only; one `SignupIdentityTx`; outcome mapping; denials recorded for `intent_used`/`slug_taken`/`signup_invalid`), `Challenge` (`SignupContinuationTx` → `signup_challenge_issue`; C/N 256-bit, digests only), `Prove` (own-row runtime UPDATE; zero rows `signup_context_mismatch`), `Complete` (`identity.LoginPinned`), `Abandon` (own-row revoke); test-only `afterProvision` hook for the lost-acknowledgement case |
| `core/internal/signup/store.go` | the five statements (provision body, challenge body, prove UPDATE, abandon UPDATE, slug probe) and the driver-error mapping (23514 → `invalid_request` field; anything else `signup_unavailable`; no driver text leaves the package) |
| `core/internal/signup/{errors,names_copy,slug,log}.go` | closed codes; byte-pinned copies of `tenancy.ValidateName`/`MaxNameLength` and `suggestBase`/`cut`/`slugCandidates`/`MaxSlugLength`/`maxSuffix`; `ValidSlug`; `signup.request` record helper (`LogRequest`, closed operations/reasons; used by the slice-2 HTTP adapter) |
| `core/internal/signup/{signup,slug}_test.go` | `TestNameRuleCopyIsExact`, `TestSlugRuleCopyIsExact`, `TestSignupSlug` (unit part), `TestSignupInputsRefusedBeforeSQL`, `TestSignupUnconfigured`, `TestSignupIntentAndSecrets`, `TestStatusOf`, `TestLogRequestIsClosed` |
| `core/internal/signup/signup_integration_test.go` | `TestSignupProvisioning` (created + capture attribution; injected faults after the Organization INSERT and at the audit capture retain nothing; SUP-7 (a) lost acknowledgement → `already_created` g2, (b) after step 3 the re-mint resets `nonce_consumed_at`/`proof_hash` and kills g2's R/C/P, runtime re-mint 42501, changed name/slug `intent_used`, (c) abandoned/expired credential `unauthenticated`; four concurrent identical retries → generations 2-5 once each, exactly one live reference; disabled creator and a 5× disable race with one durable outcome), `TestSignupSlugRace` (5× two users one slug: one `created`, loser `slug_taken` with zero writes and a pending row, `hostname_state` → winner; loser suggestion skips the taken slug; reserved/nested refused), `TestSignupContinuation` (happy path to the pinned callback `tenant_created`/`return_to "/"` with a **new provider session** accepted and the session bound to the tenant origin and Organization; R/N/C+P replay; captured R/N/P with two browsers and a foreign provisioned credential, nothing consumed on refusals, "Try again" g2 completes; same-user other row, expired generation, revoked credential at `prove`; six wrong origins at `challenge` leave R unconsumed; completion at another tenant; account substitution `signup_user_mismatch` with no mapping; removed owner → `continuation_invalid` with the generation completed and no login row; slug moved before the callback → `host_changed`) |

Commands:

| # | Command | Exit |
|---|---|---|
| 13 | unit (`--network none`, `golang:1.27.2-alpine`, `g058-a1-gomod` ro, `-mod=readonly`, `GOPROXY=off`, disposable `GOCACHE`; `unit.sh` with `PKGS="./internal/signup/ ./internal/identity/ ./internal/tenancy/access/ ./internal/tenancy/"`) | 1: `TestSignupInputsRefusedBeforeSQL` reached the audit recorder (pre-SQL refusals recorded a denial) |
| 14 | after dropping denial records for pre-SQL refusals: same | 0 |
| 15 | `gofmt -l -w` (disposable `platform-test-g062-fmt`, no network) on `internal/signup`, copied back; same unit batch (vet `-tags integration` compiles the integration file) | 0 / 0 |
| 16 | private DB (`pgvector/pgvector:0.8.7-pg17`, internal network `platform-test-g062-net`, no published port, fresh volume with `10-roles.sh`, mode-600 env with superuser): identity regression `go test -tags integration -count=1 -p 1 -run '^(TestLoginStart\|TestCallbackParametersAndTransactions\|TestCallbackContinuity\|TestEnrollmentPendingCredentialRevocation\|TestDisableVersusSignInRace\|TestRLSIdentityPool\|TestRLSFirstSignInRace\|TestSignInIsAtomic)$' ./internal/identity/` | 0 |
| 17 | same DB: `go test -tags integration -count=1 -p 1 -v -run '^(TestSignupProvisioning\|TestSignupSlugRace\|TestSignupContinuation)$' ./internal/signup/` | 1: fixture subjects `user_signup_<hex>` violate the adapter's `^user_[A-Za-z0-9]+$` (provider_error); harness helpers failed the parent test |
| 18 | after the fixture fix (`user_01SUP<hex>`; helpers take the subtest's `t`) | 1: test expectation error only (`localhost` is a valid label under base `localhost`; it is not the platform domain) |
| 19 | after correcting the case list (`example.com`, `a.b`, `acme.localhost`): all three | **0** (13 subtests PASS) |
| 20 | final unit batch on the same four packages | 0 |

Resource receipt (round 3): containers `platform-test-g062-{unit,fmt,pg,it}` removed with `docker rm -f -v` (the PG
anonymous volume with them); network `platform-test-g062-net` removed; `docker ps -a` and `docker network ls` with
`--filter label=platform-test=g062-a1` list nothing; the env file `db3.env` is deleted. No image built or pulled; no
`local-harness.lock`, Compose stack, shared fixture or published port. Logs: `~/.local/state/platform-evidence/g062-a1/it-w3-run{1,2,3}.log`.

Deviations (round 3):
- None from the Plan's contracts. `signup_challenge_issue` returns `ok = false` (not zero rows) on a mismatch; Go maps
  it to `continuation_invalid`.
- Pre-SQL refusals (malformed nonce/challenge/proof, unprovisioned credential at `prove`) are refused without an audit
  row; SQL-determined refusals are recorded through `RecordSignupDenial` (closed ops/reasons).

Review notes (rounds 1-3 consolidated, for Codex; no redesign):
1. (round 1) Plan §3.3 grants pw `SELECT (id, slug)` on `app.organizations`, but §3.5 defines no pw SELECT policy on
   organizations, so the grant is inert under FORCE RLS; `already_created` therefore returns the intent-bound `p_slug`
   rather than reading the stored slug (equal by construction of the intent digest).
2. (round 1) `role_bindings_insert_owner` additionally pins `scope_id = organization_id` and `ctx_org() IS NULL`, which
   is narrower than §3.5's listed predicate.
3. (round 2) `identity/retire.go` does not revoke signup credentials (not allow-listed; unchanged).
4. (round 2) signup revocation was also added on supersession.
5. (round 2) identity owns every signup transaction shape; `signup` issues its statements inside them.
6. (round 2) `core/cmd/server/main.go` still does not set `identity.Config.RootOrigin`; **kept unwired in slice 1**
   (fail closed). Mounting `signup/api` and wiring the signup service/root origin is slice 2 (Plan §2 main.go row).
7. (round 3) Log catalog impact found while checking events.json: the identity events `auth.callback.completed`
   (`signup`, `tenant_created`, `signup_origin_rejected`, `signup_user_mismatch`, `rate_limited`) and
   `auth.login.started` (`continuation_invalid`, `host_changed`, `organization_not_member`, `rate_limited` from
   `LoginPinned`/the root limiter) can now emit outcomes the catalog enum does not list, and `signup.request` is not
   catalogued. These are allow-listed (`events.json`, `logrecord_test.go`) mechanical catalog parity edits that the log
   validators of B-IDENTITY-PRIVACY and full qualification will require; not yet made (see transfer state).
8. (round 3) For a single-label platform domain (`localhost`), the label equal to the domain (`localhost.localhost`) is
   a valid tenant label; Spec "`B` reserved" is satisfied by the grammar (a dotted `B` never matches a label).

Hash drift / rebinding: none in round 3 (Plan, inputs, allow-list, Goal, Spec unchanged; staging `ca725c8c` delta
management-only as recorded in round 2).

### Checkpoint 3 (context ~150k): transfer state for the next writer

Done and green at this commit: everything of rounds 1-2 plus the `signup` package and its provisioning, slug and
continuation integration proofs (SUP-5 (two injected faults), SUP-6, SUP-7 (a)/(b)/(c)+concurrency+disable race, SUP-8
core matrix through the pinned callback).

Remaining slice-1 items (Plan §7 step 2), in suggested order:
1. `events.json` + `logrecord_test.go` catalog parity (review note 7) — stack-free (`go test ./...` in
   `infra/validation/tools`).
2. `TestSignupCatalog` (`core/internal/migrate/signup_integration_test.go`: pw policy rows for pending vs provisioned,
   the 14-column grant, the non-grants and 42501 probes of §3.5, return-path CHECK, TENANT refusal, 00018 precondition
   without the role) and the migrate-package version-18 parity listed under round 1 "Deferred" (`metadataThroughVersion17`
   callers, `TestMigrationLifecycle`, `TestRLSCatalog` `rlsLabel` += `pw`, `TestMembershipCatalog` role list,
   `TestAuditCatalog`).
3. identity integration `TestCallbackSignup` (SUP-2) and `TestCallbackPinned` (provider-not-live subcases with the fake's
   interaction mode; much of the pinned behaviour is already proven through `TestSignupContinuation`), plus SUP-10
   revocation of signup rows on logout/logout-all/disable (`TestEnrollmentPendingCredentialRevocation` extension).
4. `TestRecoveryPost` subcase (`core/internal/tenancy/access/recovery_integration_test.go`).
5. Then one private-DB batch for 2-4 and the run.sh mechanical 18/count parity (round 1 "Deferred").

DEFERRED (shared-harness rows; not run by policy; exact commands for the coordinator):
- `bash infra/validation/run.sh migrations` (B-RLS-MIGRATION incl. PROV-1..6, B-DBRENAME, B-MIGRATE-SET/CLI/CATALOG —
  the last three still need the version-18/count parity edit before they can pass).
- `bash infra/validation/run.sh ops` for B-RECOVERY-DRILL (`rd_tool invalidate-auth`, `signup_sessions_expired` key)
  after item 4.
- B-SIGNUP-CATALOG/B-SIGNUP-IT harness rows do not exist yet (Plan §2 run.sh row); until added, the private-run
  commands above (#16/#19) are the attributable evidence.

### Round 4 (attempt 1, writer #4): log catalog, `TestSignupCatalog` + version-18 parity, identity signup/pinned/revocation tests, recovery subcase, run.sh 18 parity

Bindings: fourth writer (Opus 5.5 high; runtime effort not observable), single write owner, worktree
`.claude/worktrees/agent-<id>`, branch `g062-impl-w4` from `origin/g062-impl`; HEAD asserted
`45b891b78faddede6d44e5aadd95a24b0d413f0a` and `51e8c57c` ancestor before any edit. Plan r5, inputs, allow-list, Goal,
Spec re-verified (`b2a612a2…`, `f89b44ed…`, `7f279a42…`, `e747807c…`, `3c90ccb6…`); `origin/staging` = `ca725c8c`
(PROGRESS `587c1016…`, as round 2); `git diff --stat 51e8c57c ca725c8c -- core web infra contracts` empty, so no rebase.
`bash infra/ci/docker-desktop-mounts.sh` exit 0 (17%). Rounds 1-3 passes are reused: no non-test core file, migration,
role script or `invalidate.sql` changed in this round (only test files, the log catalog and run.sh strings).

Files (all match the 40-pattern allow-list; `git diff --name-only 51e8c57c` + untracked, filtered by the patterns, is empty):

| Path | Change |
|---|---|
| `infra/validation/tools/internal/logrecord/events.json` | `auth.login.started` += `continuation_invalid`, `host_changed`, `organization_not_member` (`LoginPinned`); `auth.callback.completed` += `signup`, `tenant_created`, `signup_origin_rejected`, `signup_user_mismatch`; new closed `signup.request` (operations/reasons exactly `signup/log.go`'s sets; failed only as `signup_unavailable`) |
| `infra/validation/tools/internal/logrecord/logrecord_test.go` | catalogue list += `signup.request`; accept cases for the new outcomes and `signup.request`; reject cases (credential/slug/user id keys, foreign or misplaced reasons, unknown operation, ok with a reason, outcomes on the wrong identity event) |
| `core/internal/migrate/signup_integration_test.go` (new) | `TestSignupCatalog`: CT-SUP-1 precondition (role missing / INHERIT refused at 17, nothing applied), tables, policies (pw policies admit pending and provisioned rows of an active user; proof WITH CHECK; no rt/ow INSERT path beyond the 00014 group_owner bootstrap), privileges (rt INSERT 8 / UPDATE 4, pw UPDATE 14, pw grants elsewhere, non-grants), definers (4 bodies; owned 3/8/7), negatives (TENANT/LOGIN/empty shapes 42501 with the named constraint; completion with another Organization 0 rows, right one once, replay 0; rt and ow Organization / org_owner INSERT 42501), provision writer (visible rows, WITH CHECK 42501, ungranted column/DELETE/login_transactions/hostname_state/is_member/audit SELECT/depth-0 audit INSERT 42501), return path (23514 with the named constraints), updown (refused with 3 rows, catalog unchanged; exact 17 catalog after down; exact 18 after up) |
| `core/internal/migrate/integration_test.go` | `signupRelations`, `metadataThroughVersion18` (79 relations; `auth_resolve_signup`, three pw bodies); `TestMigrationLifecycle` up/synthetic-down/reapply/down/up-again at 18, new down-from-18 step to the exact 17 form |
| `core/internal/migrate/rls_integration_test.go` | `rlsLabel` += `pw`; `rlsWantMemberships` += `ow>pw` (the fourth helper exists at every version since round 1) |
| `core/internal/migrate/runner_test.go` | `TestNormalSetIsTheReviewedSet` and the provider-source test at 18 (unit tests; see deviation 1) |
| `core/internal/identity/service_test.go` | `TestLoginStart` + SUP-1 block (root-only signup intent: www/tenant/auth/custom/port/scheme/path refused, enrollment facts `invalid_request`, unconfigured root refused, stored row shape); new `TestCallbackSignup` (SUP-2) and `TestCallbackPinned` (SUP-8 identity part); `TestEnrollmentPendingCredentialRevocation` extended with two signup credentials (SUP-10: bound logout/provider logout/supersession revoke only the sid1 row; logout-all and disable revoke both, re-enable revives neither; the expired signup row is swept by login start, the live one stays) |
| `core/internal/identity/store_test.go` | `TestIdentityCatalog`: table set += `signup_sessions`, forced tables 18 |
| `core/internal/tenancy/access/recovery_integration_test.go` | setup pin `current_version 18`; SUP-11 subcase: `TestRecoverySeed` seeds a live provisioned signup row of `memberA2` before checkpoint B1 (digest only in `old-auth.json`); `TestRecoveryVerify` V6 asserts the restored credential no longer resolves and the row is unrevoked with `expires_at`/`continuation_expires_at` ≤ verify start (see deviation 2) |
| `infra/validation/run.sh` | version-18 parity, strings/expected rows only: B-MIGRATE-SET set + text; root getting-started migrate block (00018 pending/applied/inspected, versions 18); B-MIGRATE-CLI steps (apply 79, reapply 00018, rollback to 17 = 71, inspect 00018 pending at 17, re-apply 79, production version 18) and text; B-MIGRATE-CATALOG unique rows 0/71/79 and text; B-AUDIT-CATALOG reads 00018's action constraint (still 46/39); B-RECOVERY-DRILL setup at 18, RD-A also requires `signup_sessions_expired=[1-9]` in the output and the receipt; B-PERSIST, B-TOGGLE-OFF at 18 with 00018 applied; B-IDENTITY-CATALOG `SIGNUP_OBJECTS` (79 in all); "at version 18" PASS texts of B-RLS-IT, B-AUDIT-IT, B-ONBOARDING-IT, B-MEMBERSHIP-IT, B-ENROLLMENT-IT. `bash -n` clean |

Commands (unit containers: existing `golang:1.27.2-alpine`, `--network none`, `g058-a1-gomod` read-only,
`GOFLAGS=-mod=readonly GOPROXY=off`, disposable `GOCACHE`, source by `docker cp`, label `platform-test=g062-a1`; scripts
under the worktree's ignored `.local/tmp/`; logs `~/.local/state/platform-evidence/g062-a1/{unit,it}-w4-*.log`):

| # | Command | Exit |
|---|---|---|
| 21 | tools module (`infra/validation/tools`): `gofmt -l internal/logrecord`, `go vet ./internal/logrecord/`, `go test -count=1 ./...` | 0 (7 packages ok) |
| 22 | `gofmt -l -w` on migrate/identity/tenancy/access/signup/migrations (disposable `platform-test-g062-fmt`); `service_test.go` reformatted and copied back | 0 |
| 23 | unit: gofmt + vet + vet `-tags integration` + `go test -count=1` on `./internal/migrate/ ./migrations/ ./internal/identity/ ./internal/tenancy/access/ ./internal/signup/` | 0 |
| 24 | private DB (pool slot 1 of 2; no other `platform-test-*` run live; `pgvector/pgvector:0.8.7-pg17`, internal network `platform-test-g062-net`, no published port, fresh volume with `10-roles.sh` → 8 roles incl. `g062db_provision_writer`): identity `-run '^(TestLoginStart\|TestCallbackSignup\|TestCallbackPinned\|TestEnrollmentPendingCredentialRevocation\|TestIdentityCatalog\|TestDisableAndEnable\|TestDisableVersusSignInRace\|TestRLSIdentityPool\|TestRLSFirstSignInRace)$'`, then migrate `-run '^(TestSignupCatalog\|TestMigrationLifecycle\|TestRLSCatalog\|TestRLSMigration\|TestMembershipCatalog\|TestAuditCatalog\|TestEnrollmentCatalog)$'` (`ENROLLMENT_REPLAY_SQL_B64` = the bound replay script) | 1: 14/16 top-level PASS; `TestCallbackPinned/expected_user` (test asserted `auth.sign_in` by Organization; it is an identity event without one), `TestSignupCatalog/{privileges,negatives}` (test expectations: `format('%s', bool)` prints `f`; owner's own ACL entry on its functions; LOGIN binding must be 64 hex) |
| 25 | same DB, after the three test-only corrections: `-run '^TestCallbackPinned$'` and `-run '^TestSignupCatalog$'` | **0** (4/4 and 9/9 subtests) |
| 26 | final unit batch as #23 | 0 |

Resource receipt (round 4): containers `platform-test-g062-{fmt,unit,pg,it}` removed with `docker rm -f -v` (the PG
anonymous volume with it); network `platform-test-g062-net` removed; `docker ps -a`, `docker network ls` and `docker volume
ls` with `--filter label=platform-test=g062-a1` list nothing; the mode-600 env file `db4.env` deleted. No image built or
pulled; no `local-harness.lock`, Compose stack, shared fixture or published port; the module volume read-only only.

Deviations (round 4):
1. `core/internal/migrate/runner_test.go` (unit, no build tag) still pinned the 17-file set; earlier unit batches ran
   `./migrations/` but not `./internal/migrate/` tests, so this round-1 parity gap went unseen. Fixed (allow-listed).
2. SUP-11 placement: the Plan names "one `TestRecoveryPost` subcase seeding a signup row before the checkpoint", but the
   drill runs Seed → backup B1 → Post → B2 → restore of B1, so a row written by Post is never restored. The fixture is
   therefore seeded in `TestRecoverySeed` (before B1) and asserted in `TestRecoveryVerify` V6 (same file, same drill row
   B-RECOVERY-DRILL). The seeded continuation lifetime is 29 min (fixture only) so the invalidation is observable.
3. SUP-8 "provider session not live" at L1: the fake WorkOS has no hosted-page model; `TestCallbackPinned` covers what the
   callback observes (a code with a new provider session and fresh `auth_time` → `tenant_created`; another user →
   `signup_user_mismatch`; `access_denied` → `login_canceled`, tenant and membership intact, later ordinary sign-in works).
   The real prompt is SUP-15 (slice 2 browser) and G057.
4. `rate_limited` is not added to the identity log enums: the root limiters answer in the HTTP adapter before the service
   logs (round-3 review note 7 listed it as a possibility).

Review notes (round 4, for Codex; no redesign):
9. **Scope blocker — version-18 pins outside the allow-list.** Four test files that are not allow-listed pin version 17 or
   the pre-G062 `invalidate.sql` and will fail with 00018 present (not edited, per the allow-list rule):
   - `core/internal/tenancy/access/integration_test.go:236` (`setupIT` Up expects `current_version\t17`): every access IT
     suite (B-RLS-IT, B-POLICY-IT, B-OBJ-IT, B-AUDIT-IT, B-ONBOARDING-IT, B-MEMBERSHIP-IT, B-ENROLLMENT-IT, and the Plan's b2
     selectors `TestEnrollment`/`TestEnrollmentAdmission`); needs `18`.
   - `core/internal/tenancy/integration_test.go:144` (teardown chain starts at 17): needs a version-18 branch (purge signup
     rows/audit history through the attested superuser, then down to 17) before the 17 branch.
   - `core/internal/mcp/env_integration_test.go:253` (`current_version\t17`): B-MCP-IT; needs `18`.
   - `core/internal/mcp/mcp_integration_test.go:1432` pins `invalidate.sql` SHA-256 `f198f41d…`; with the round-2 fold it is
     `931b4e066807fc2dfcb96d0949380c90b3a780965325c605c5d9fa504014e312` (TestMCPRecovery, B-MCP-IT).
   These are one-line mechanical pin edits; they need an allow-list widening (four exact paths) by Codex.
10. **B-SIGNUP-CATALOG/B-SIGNUP-IT harness rows not added.** Adding required-check IDs changes the per-suite counts that
    `infra/ci/validate-controls.sh` pins (e.g. VU1 "fast 35 + gate 90 -> 121", VU9, the RC/release-verify lines), which is
    outside the allow-list. The Plan's §7 slice-1 list names no B-SIGNUP harness row; the attributable evidence is the
    private-run selectors #19/#25 until the rows are added with the CI count pins (final-coverage stage, needs that path).
11. `identity/retire.go` (note 3) remains: a retired user's unexpired signup credential would resolve again after a
    re-enable (within its 30-minute lifetime); `TestEnrollmentPendingCredentialRevocation`'s retire subcase does not assert
    signup rows for that reason.
12. `TestRLSMigration` passed in the private run (#24); its PROV-1..6 half is harness-only (`rls_provision`).

Hash drift / rebinding: none in round 4. Expected drift recorded: `infra/recovery/sql/invalidate.sql` is
`931b4e06…` (planning-inputs binding `f198f41d…` is the pre-change baseline; the fold is the Plan §2 change).

### Slice-1 status after round 4

Slice-1 source items of the transfer state are done within the allow-list: catalog parity, `TestSignupCatalog` and the
migrate version-18 parity, identity `TestCallbackSignup`/`TestCallbackPinned`/SUP-10 revocation, the SUP-11 recovery
subcase, run.sh version-18 parity. RootOrigin stays unwired in `core/cmd/server/main.go` (root signup fails closed until
slice 2). **Open before slice-1 b2 can be complete:** review note 9 (allow-list widening for four pin files) and the
DEFERRED shared-harness rows below.

DEFERRED (shared harness, not run by policy; exact commands for the coordinator, from the repository root):
- `bash infra/validation/run.sh migrations` — B-MIGRATE-SET/CLI/CATALOG (now at 18/79), B-RLS-CATALOG/MIGRATION incl.
  PROV-1..6, B-AUDIT-CATALOG (00018 path), B-IDENTITY-CATALOG (79), B-ENROLLMENT-CATALOG/IT, B-RECOVERY-DRILL
  (Seed/Verify SUP-11 subcase, RD-A `signup_sessions_expired`), B-MCP-IT and the access IT rows — the last two groups fail
  until review note 9 is resolved.
- `bash infra/validation/run.sh ops` — B-PERSIST (at 18), B-IDENTITY-PRIVACY/IT (log catalog incl. the new identity
  outcomes); `bash infra/validation/run.sh off` — B-TOGGLE-OFF (at 18); `bash infra/validation/run.sh rename` — B-DBRENAME.

### Round 5 (attempt 1, writer #5): scope amendment 01, version-18 pins in four access/tenancy/MCP test files

**Scope-amendment rebinding (recorded before editing the four files).** Fifth writer (Opus 5.5 high; runtime effort not
observable), single write owner, worktree `.claude/worktrees/agent-<id>`, branch `g062-impl-w5` from
`origin/g062-impl`; HEAD asserted `e66f2bc42ae578424201ded3a0a9f775b7330f7b` and `51e8c57c` ancestor before any edit.
Codex disposition (staging `681e091b`, handoff 716f6668): the executable allow-list is widened from 40 to 44 exact ERE
patterns (amendment 01). Staging is now `51fc3c86` (management-only; no rebase, original target stays `51e8c57c`).

| Binding | Old | New |
|---|---|---|
| `management/v0/results/G062-scope-allowlist.txt` (staging copy) | `7f279a42c97256b5b1abffafd77a770c1e6d2effaf3b8717de96b9263dea073f` (40) | `f8eaf8ee4c1848688ec3a3b7030ebb3e3687c02cd35c58aa79cfeede0db67c53` (44) |
| `management/v0/PROGRESS.md` (staging `51fc3c86`) | `1219f224…` (last recorded) | `e24c88f472eb86f8bc332fa0b8b48e02755aa9416dad5c57ab488e2f38861e17` |

The four added patterns are exactly `^core/internal/tenancy/integration_test[.]go$`,
`^core/internal/tenancy/access/integration_test[.]go$`, `^core/internal/mcp/env_integration_test[.]go$`,
`^core/internal/mcp/mcp_integration_test[.]go$` (the only content delta besides the header line). The branch's own copy of
the allow-list file is left unchanged (still `7f279a42…`; Codex-owned); the scope check uses the verified staging copy saved
to the worktree's ignored `.local/tmp/G062-scope-allowlist-44.txt`. `infra/recovery/sql/invalidate.sql` on this branch is
`931b4e066807fc2dfcb96d0949380c90b3a780965325c605c5d9fa504014e312` (matches the pin requested in review note 9).

Pre-batch: `git fetch origin g062-impl staging`; `git diff --stat 51e8c57c origin/staging -- core web infra contracts` empty
(management-only), so no rebase. `bash infra/ci/docker-desktop-mounts.sh` exit 0 (17%). Rounds 1-4 passes are reused: no
non-test core file, migration, role script, `invalidate.sql` or harness file changed in this round.

Files (all four match the 44-pattern copy; `git diff --name-only 51e8c57c` + untracked = 41 paths, `grep -vEf` against
the 44 patterns prints nothing):

| Path | Change |
|---|---|
| `core/internal/tenancy/access/integration_test.go` | `setupIT` Up expects `current_version\t18` (comment: latest version 18, G062 added) |
| `core/internal/tenancy/integration_test.go` | setup teardown: new version-18 branch before the 17 branch — through the attested superuser deletes `app.signup_sessions` and the signup-shaped/`/signup/details` login transactions (inside the existing `purgeDomainState` call, which also clears the disposable domain state the 17 branch clears anyway) and purges the audit history, then `Down` must report `current_version\t17` |
| `core/internal/mcp/env_integration_test.go` | `setupMCP` Up expects `current_version\t18` |
| `core/internal/mcp/mcp_integration_test.go` | `TestMCPRecovery` pins `invalidate.sql` SHA-256 `931b4e06…e312` (verified with `sha256sum` on this branch) |

Commands (unit container as round 4: `golang:1.27.2-alpine`, `--network none`, `g058-a1-gomod` read-only,
`-mod=readonly`, `GOPROXY=off`, disposable `GOCACHE`, source by `docker cp`; private DB as round 4:
`pgvector/pgvector:0.8.7-pg17`, internal network `platform-test-g062-net`, no published port, fresh volume with
`10-roles.sh` (8 roles), mode-600 env with `MCP_INVALIDATE_SQL_B64` = this branch's `invalidate.sql` and three run
canaries; pool slot 1 of 2, no other `platform-test-*` container live before launch; label `platform-test=g062-a1`;
scripts `.local/tmp/{unit,it5,run5}.sh`; logs `~/.local/state/platform-evidence/g062-a1/{unit,it}-w5-*.log`):

| # | Command | Exit |
|---|---|---|
| 27 | unit: gofmt + vet + vet `-tags integration` + `go test -count=1` on `./internal/tenancy/ ./internal/tenancy/access/ ./internal/mcp/` | 0 |
| 28 | DB 1: access `-run '^(TestEnrollment\|TestEnrollmentAdmission\|TestEnrollmentRequests\|TestOnboarding\|TestMembership\|TestAuthority\|TestRLSIsolation\|TestRLSPool\|TestDomains\|TestRetention)$' ./internal/tenancy/access/`, then `-run '^TestSignupProvisioning$' ./internal/signup/` | **0** (10/10 top-level, 134 subtests; signup 6/6); DB then at 18 with 12 signup rows and 21 00018 audit rows |
| 29 | DB 1: `-run '^(TestTenancyModel\|TestBusinessProfileModel)$' ./internal/tenancy/` (one process) | 1: the setup step-down 18→5 through the new branch succeeded with those rows present; `TestBusinessProfileModel/IT-BP-REF-2` then failed `53300` (connection slots exhausted, note 13); `TestTenancyModel` PASS |
| 30 | DB 1 (at 5): `-run '^TestBusinessProfileModel$' ./internal/tenancy/` | 0 (35/35) |
| 31 | DB 1 (at 5): harness-order phase | 1: invalid ordering on my side (access `Up` from a populated version-5 tenancy world stops at 00015's legacy-owner preflight, pre-existing behaviour); not evidence, DB discarded |
| 32 | DB 2 (fresh): access `-run '^(TestEnrollment\|TestEnrollmentAdmission)$'` (→18), `TestSignupProvisioning`, then `TestTenancyModel` and `TestBusinessProfileModel` in separate processes (run.sh order) | 1: access 2/2 PASS; **`TestSignupProvisioning/disabled_creator_and_disable_race` FAIL `race 2: signup_unavailable`** (note 14); `TestTenancyModel` PASS stepping 18→5 with signup rows present; `TestBusinessProfileModel` PASS 35/35 |
| 33 | DB 3 (fresh, `ALTER DATABASE g062db SET plan_cache_mode = force_generic_plan`): `-run '^TestSignupProvisioning$/^created$' ./internal/signup/` | 1: deterministic reproduction of note 14 (`provision: signup_unavailable`; PG log `permission denied for table organizations` in `audit.events_attribute()`) |
| 34 | DB 4 (fresh): `-run '^(TestMCPTransport\|TestMCPLifecycle\|TestMCPAudit\|TestMCPDiscovery\|TestMCPRecovery)$' ./internal/mcp/` | **0** (7/6/1/5/1; 18 subtests) |

DEFERRED (MinIO/shared stack or harness-only; not run by policy): access suites needing MinIO (`setupObjects`/
`setupRecovery`: `TestAccessOrdering`, `TestAudit`, `TestBranding`, `TestObjects*`, `TestIsolationAttacks`,
`TestRecovery*`) and the unselected DB-only ordering/footprint suites (`TestRLSOrdering`, `TestRLSFootprint`,
`TestAccessFootprint`, `TestMCPOrdering`, `TestMCPFootprint`). Exact command: `bash infra/validation/run.sh migrations`
(B-POLICY-IT/ORDERING, B-RLS-IT/ORDERING, B-OBJ-*, B-AUDIT-IT, B-BRANDING-IT, B-MCP-IT incl. the facade half,
B-RECOVERY-DRILL), plus the round-4 DEFERRED list (unchanged).

Resource receipt (round 5): containers `platform-test-g062-{unit,pg,it}` removed with `docker rm -f -v` (each PG anonymous
volume with it; PG recreated three times within the same single slot); network `platform-test-g062-net` removed;
`docker ps -a`, `docker network ls`, `docker volume ls` with `--filter label=platform-test=g062-a1` list nothing; `db5.env`
deleted. No image built or pulled; no `local-harness.lock`, Compose stack, MinIO, shared fixture or published port.

B-SIGNUP-CATALOG / B-SIGNUP-IT rows (investigated, **not added**; Codex defers them to final local coverage):
- `infra/validation/run.sh` (allow-listed): a `signup_checks` function after `enrollment_checks` — B-SIGNUP-CATALOG
  `rls_go signup-catalog … ./internal/migrate/ '^TestSignupCatalog$'` (9 subtests); B-SIGNUP-IT `rls_go signup-it …
  ./internal/signup/ '^(TestSignupProvisioning|TestSignupSlugRace|TestSignupContinuation)$'` with `it_count` pins and the
  access logcheck; both IDs added to the MinIO-failure FAIL list (run.sh:2517) and the call added beside
  `enrollment_checks`. Identity `TestCallbackSignup`/`TestCallbackPinned` already run under B-IDENTITY-IT
  (`./internal/identity/...`).
- `infra/validation/required-checks.tsv` (allow-listed): `B-SIGNUP-CATALOG<TAB>migrations`, `B-SIGNUP-IT<TAB>migrations`
  after `B-ENROLLMENT-IT`; header `+ 2 G062 = 129 rows`, `G062: B-SIGNUP-CATALOG/IT migrations`, profile-line full count +2.
- **Scope-extension proposal (outside the 44 patterns):** `infra/ci/validate-controls.sh` count pins (the gate remainder
  grows by 2; `migrations` is in the full/RC gate only): `ctl_VA7` (:1146 `sum_has required_count 90` → `92`; :1149
  detail "90 required IDs" → "92"); `ctl_VU1` (:1268 `vu_split full 35 90 121` → `35 92 123` and detail); `ctl_VU10`
  (:1320 `vu_split rc 35 90 121` → `35 92 123`; :1326 detail); `ctl_J8` (:2317 `(( n != 90 ))` → `92` and message).
  `ctl_VU9` (integration 35/44/75) is unchanged: the integration profile has no `migrations` suite. B-CI-CONFIG counts are
  computed, not pinned; no other consumer found (precedent: G058 staging repair `2283a198` changed exactly these four
  controls). Not edited.

Deviations (round 5): none from the requested edits. The 18 branch reuses `purgeDomainState` (one superuser connection)
for the signup deletes instead of adding a helper.

Review notes (round 5, for Codex):
13. **Connection margin in `TestBusinessProfileModel` (pre-existing fragility).** When its own setup performs the full
    step-down from 18 in the same process, each step's superuser purge connection is held until test cleanup (`connect`
    registers `t.Cleanup`) and IT-BP-REF-2 opens 50 more owner connections, exceeding PostgreSQL's default
    `max_connections=100` (`53300`). The new 18 branch adds two held connections. In run.sh order (B-TENANCY-IT steps down
    first in its own process; B-BUSINESS-PROFILE-IT then starts at 5) both pass (#32). No change made; bounded test-only
    fix if wanted: close each purge connection at the end of its helper (same file).
14. **Blocker: provisioning fails under a generic plan (slice-1 defect in 00018).** `audit.events_attribute()` (00018:256)
    evaluates one `ok := CASE …` expression whose G060 `access.override_used` branch (00018:451) reads
    `app.organizations.group_management_override`. The provision writer holds only `SELECT (id, slug)` on
    `app.organizations`. With a custom plan the untaken branch is folded away; with a generic plan (`plan_cache_mode =
    auto` may switch after five executions in a session; `force_generic_plan` always) the executor permission check covers
    the subquery, so the `membership.granted` capture inside `signup_provision_organization` fails `42501 permission
    denied for table organizations` and `Provision` returns `signup_unavailable` (nothing retained; fail closed).
    Reproduced intermittently by the disable race (#32) and deterministically with `force_generic_plan` (#33). In
    production a pooled backend could refuse every signup once it uses the generic plan. Not fixed this round (outside
    the assigned pin work; touches the audit attribution body). Smallest options: (a) preferred, no privilege change —
    move the override lookup out of the CASE into a PL/pgSQL statement executed only when `NEW.action =
    'access.override_used'` (`IF … THEN SELECT … INTO ovr; END IF;`, CASE uses `ovr`) in 00018's up body, updating any
    body/definer pins in `TestSignupCatalog`/`TestAuditCatalog`; (b) grant the provision writer `SELECT
    (group_management_override)` (inert under FORCE RLS without a pw policy) — a grant-inventory change needing review.
    Regression proof: a `TestSignupProvisioning` subcase with `plan_cache_mode = force_generic_plan` on the provisioning
    connection (or more than five provisions on one connection).

Slice-1 status after round 5: the four version-18 pin files are done; their DB-only selectors pass (#28, tenancy #30/#32,
#34). **Slice 1 is not complete:** besides the Codex-deferred items (B-SIGNUP rows; `identity/retire.go`), review note 14
is an open slice-1 defect needing a disposition and a fix round.

### Round 6 (attempt 1, writer #6): round-5 correction — generic-plan audit attribution (review note 14)

**Rebinding (recorded before the first write).** Sixth writer (Opus 5.5 high; runtime effort not observable), single
write owner, worktree `.claude/worktrees/agent-<id>`, branch `g062-impl-w6` from `origin/g062-impl`; HEAD
asserted `0f39c4cd9c9678835cabd0c8e14f8a7e4535b92c` and `51e8c57c` ancestor before any edit. Bounded correction within
attempt 1 (budget unchanged). Staging is `29f6652c` (management-only: `git diff --stat 51e8c57c origin/staging -- core
web infra contracts` empty; no rebase, original target stays `51e8c57c`).

| Binding | Value (verified on `origin/staging`) |
|---|---|
| Codex disposition `management/v0/results/G062-round5-audit-generic-plan-review-20261020.md` | `4bfe09d3fdb3ea04a71bb4630b1b575bd34395289a4d787fc03d36178ec9f2b1` (option A; no grant change) |
| `management/v0/results/G062-scope-allowlist.txt` (staging copy) | `7d4db9a20aef0f6d7110b135fd50e7260780a8d0c3be56073987b9746186611a` (45 patterns; was `f8eaf8ee…` / 44). Amendment 02 adds `^core/internal/identity/retire(_test)?[.]go$` (identity retire files; **not touched** here, a separate correction is held for Codex review) |
| `management/v0/PROGRESS.md` (staging `29f6652c`) | `7ffa564a8a6c3136acde17969cdaaef15b07b6578508aa614625570a5d5b2b63` (was `e24c88f4…`) |

The branch's own allow-list copy stays unchanged (Codex-owned); the scope check uses the verified staging copy saved to the
worktree's ignored `.local/tmp/allowlist45.txt`.

**Change.** `core/migrations/00018_platform_signup.sql` up body of `audit.events_attribute()` only (the down body, the
exact 00017 restore, is unchanged; no grant, policy, constraint or inventory change). A new local `ovr boolean` is set
`false`, and only `IF NEW.action = 'access.override_used'` a separate statement assigns `ovr := app.rls_org_role() IN
('org_owner', 'org_admin') AND coalesce((SELECT og.group_management_override FROM app.organizations og WHERE og.id =
NEW.organization_id), false)`; the CASE's override branch uses `AND ovr` in place of those two conjuncts (the rest of the
branch unchanged). AND is commutative under three-valued logic and `ok` is consumed as `coalesce(ok, false)`, so the
branch's result is identical; other actions never execute the lookup. Header comment above the body notes the move.
00018 SHA-256 `a05d7b56…04e1` → `3d391f6b4c2502e3714c676acb1be8f5609d4111352ae4ec7a66c3a1e19fbe81`.

**Regression.** `core/internal/signup/signup_integration_test.go`: the harness gains `on(pool)` (same fake provider,
issuer binding and clock, another identity store pool) and `genericPlanPool` (one-connection runtime pool,
`plan_cache_mode=force_generic_plan` as a startup parameter, so also the RESET value). New last subtest
`TestSignupProvisioning/generic_plan_on_the_provisioning_backend`: the whole flow (root login, T1, resolve, provision) on
that single backend, six provisions; each asserts `OutcomeCreated`, then `assertOnePair` (one Organization, one org_owner
binding, one `organization.created`, one `membership.granted`), and that `pg_backend_pid()` and `plan_cache_mode` are
unchanged after each provision (same backend, generic mode).

Commands (unit container: `golang:1.27.2-alpine`, `--network none`, `g058-a1-gomod` read-only, `-mod=readonly`,
`GOPROXY=off`, disposable `GOCACHE`, source by `docker cp`; private DB: `pgvector/pgvector:0.8.7-pg17`, internal network
`platform-test-g062-net`, no published port, fresh anonymous volume per run with `10-roles.sh` (8 roles), mode-600 env;
pool slot 1 of 2, no other `platform-test-*` container live before each launch; label `platform-test=g062-a1`; scripts
`.local/tmp/{unit,rununit6,pg6,run6,it6}.sh`; logs `~/.local/state/platform-evidence/g062-a1/{unit,it,pg}-w6-*.log`).
Pre-batch: `git fetch origin g062-impl staging` (staging code-tree delta empty), `bash infra/ci/docker-desktop-mounts.sh`
exit 0 (17%).

| # | Command | Exit |
|---|---|---|
| 35 | unit: gofmt + vet + vet `-tags integration` + test `./internal/signup/` (new test, old body) | 0 |
| 36 | DB 1 (fresh, **old 00018 body**): `-run '^TestSignupProvisioning$' ./internal/signup/` | **1 (expected)**: 6 existing subtests PASS; `generic_plan_on_the_provisioning_backend` FAIL `provision 0 under a generic plan: … signup: signup_unavailable`; PG log `ERROR: permission denied for table organizations` (42501) in `PL/pgSQL assignment "ok := CASE …"` (`pg-w6-old.log`) |
| 37 | DB 2 (fresh, diagnostic variant: only the subquery moved, `app.rls_org_role()` left in the CASE) | 1: 7/8 FAIL incl. `created` in the default plan mode; PG `permission denied for function rls_org_role` (see deviation) |
| 38 | unit: gofmt + vet (plain and `-tags integration`) + test `./migrations/ ./internal/audit/ ./internal/migrate/ ./internal/signup/` (fixed body) | 0 |
| 39 | DB 3 (fresh, **fixed body**): access `-run '^(TestAuthority\|TestEnrollment\|TestEnrollmentAdmission\|TestMembership)$'` (G060 override cases MEM-OVR-*/O-OVR-* incl. `access.override_used` rows), then signup `-run '^(TestSignupProvisioning\|TestSignupSlugRace\|TestSignupContinuation)$'` | **0**: access 4/4 top-level (63 subtests); signup 3/3, `TestSignupProvisioning` 7/7 incl. **`generic_plan_on_the_provisioning_backend` PASS**; no `permission denied` for organizations/rls_org_role in `pg-w6-fix.log` |
| 40 | DB 4 (fresh, fixed body): migrate `-run '^(TestSignupCatalog\|TestMigrationLifecycle\|TestAuditCatalog\|TestRLSCatalog\|TestMembershipCatalog)$'` | **0** (5/5; 39 subtests) |

Body-hash/catalog pins: **none changed.** Derived from the code: no test or harness pins the 00018 `events_attribute`
body (the md5 pins are of the 00008/00013/00015/00016 bodies at their versions; `TestSignupCatalog` and
`TestMigrationLifecycle` compare live up/down/up snapshots, not a fixed digest; `TestAudit` (unit) and `run.sh` read only
the constraint lists of 00018; B-MIGRATE-SET digests only 00001-00010). Confirmed by #38/#40.

DEFERRED (unchanged policy; not run): MinIO/shared-stack access suites including `TestAudit` (B-AUDIT-IT, which also
exercises `access.override_used`) and the rest of the round-5 DEFERRED list; exact command `bash
infra/validation/run.sh migrations`. B-SIGNUP rows are still not added; when they are, B-SIGNUP-IT's `it_count` for
`TestSignupProvisioning` must count 7 subtests (this round adds one).

Resource receipt (round 6): containers `platform-test-g062-{unit,it,pg}` removed with `docker rm -f -v` (PG recreated four
times within one slot, each anonymous volume removed with it); network `platform-test-g062-net` removed; `docker ps -a`,
`docker network ls`, `docker volume ls` with `--filter label=platform-test=g062-a1` list nothing; `db6.env` deleted. No
image built or pulled; no `local-harness.lock`, Compose stack, MinIO, shared fixture, published port or `run.sh` suite.

Deviation (round 6, for Codex review):
15. **`app.rls_org_role()` moved with the override lookup.** The disposition names only the `group_management_override`
    lookup. Moving it alone (#37) turns the CASE into a PL/pgSQL *simple expression*, which PL/pgSQL always evaluates with a
    generic plan; expression initialisation then checks EXECUTE on every function of every branch, and the provision
    writer has no EXECUTE on `app.rls_org_role()` (00006 grants runtime/owner only). Every provision then failed, even in
    the default mode. The override branch's other privileged read therefore moves into the same `IF` statement with the
    same predicate. No grant change; behaviour of the override branch is identical (proved by #39's G060 override cases;
    B-AUDIT-IT deferred). Side effect for review: a role lacking those privileges that appends `access.override_used`
    directly now gets the privilege error deterministically instead of a plan-dependent shape refusal/privilege error.

Slice-1 status after round 6: review note 14 is corrected and regression-proved (old FAIL #36, fixed PASS #39). Remaining
slice-1 items are the Codex-held/deferred ones only (B-SIGNUP rows and their `validate-controls.sh` proposal;
`identity/retire.go` correction held for review; deferred MinIO/harness selectors).

### Round 7 (attempt 1, writer #7): retirement revokes pending signup credentials (AC6 / SUP-10)

**Rebinding (recorded before the first write).** Seventh writer (Opus 5.5 high; runtime effort not observable), single
write owner, worktree `.claude/worktrees/agent-<id>`, branch `g062-impl-w7` from `origin/g062-impl`; HEAD
asserted `1a7827cdd37dda9d64059c7e98993f38aa4e68b8` and `51e8c57c` ancestor before any edit. Bounded correction within
attempt 1 (budget unchanged). Staging is `bb1a8d90750b9c8cbc96e08b65e6537fdc0376a4`; no rebase (original target stays
`51e8c57c`). `git diff --stat 51e8c57c origin/staging -- core/internal/identity core/internal/tenancy core/migrations` is
empty. Reconciliation note for Codex: staging now also differs from `51e8c57c` in `infra/security/gitleaks.toml`,
`infra/validation/render-customizations.tsv` and `infra/validation/run.sh` (MAINT `bbf6f024`, `9e10e223`); this branch
also edits `run.sh`, so the pending staging reconciliation must merge that file. Not touched here.

| Binding | Value (verified) |
|---|---|
| Proposal `~/.local/state/platform-evidence/g062-a1/g062-retire-signup-revocation-proposal-20261020.md` | `58c4f3fb523f0c18da6082c04a2ec0694f7c2b6125a0fba5f5ad0aa5c0d167e7` (its "44 patterns / extension needed" text is stale; superseded by the review) |
| Codex review `management/v0/results/G062-retirement-revocation-proposal-review-20261020.md` (staging) | `fe3617b975418a12d322016f7808475928b8461b1a66c41bd16929ce2ce4ad68` (proposal ACCEPTED; bounded implementation RELEASED) |
| `management/v0/results/G062-scope-allowlist.txt` (staging copy) | `7d4db9a20aef0f6d7110b135fd50e7260780a8d0c3be56073987b9746186611a` (45 patterns incl. `^core/internal/identity/retire(_test)?[.]go$`; saved to the worktree's ignored `.local/tmp/allowlist.ere`) |
| Plan r5 `management/v0/plans/G062-root-host-tenant-signup-plan.md` | `b2a612a2c50b4a36505346c2c18214cd785d7d62277bdb35c993db49a3a5a3e2` (unchanged) |
| `management/v0/PROGRESS.md` (staging `bb1a8d90`) | `b1917fb537b9bb885320f5e0d6350ebd0b2067eecf74c3d688bf5fa85b656e65` (was `7ffa564a…` at round 6) |

**Change (write set exactly as released).**
- `core/internal/identity/retire.go`: `finalizeRetirement` calls the existing `revokeSignupSessionsAll(ctx, tx, userID,
  revokeDisabled, now)` (`store.go`, statement `sqlRevokeSignupAll`) right after the enrollment revocation and before
  `revocationGap`, in the same IDENTITY transaction, as `DisableUser` does. Header comments of the file, `FinalizeRetirement`
  and `finalizeRetirement` name the signup UPDATE. No migration, grant, policy, API, event or catalog change.
- `core/internal/identity/retire_test.go`: `TestFinalizeRetirementRevokesEnrollmentInOrder` pins execs 4, writes
  `users → sessions → sqlRevokeEnrollmentAll → sqlRevokeSignupAll`, `stmts[n-2] == sqlRevokeSignupAll` (gap read last), and a
  `"signup UPDATE": "UPDATE app.signup_sessions"` failAt case (identity_unavailable, zero commits).
- `core/internal/identity/service_test.go` (integration): the existing subcase `retire revokes both credentials and
  re-enable revives neither` asserts `f.signupBoth(t, "disabled")` after `FinalizeRetirement` and again after `EnableUser`,
  plus the positive control `seedSignup(t, f.user, "session_user_01E19RE_3")` resolves. No new top-level test.

Commands (unit container: `golang:1.27.2-alpine`, `--network none`, `g058-a1-gomod` read-only, `-mod=readonly`,
`GOPROXY=off`, disposable `GOCACHE`, source by `docker cp`; private DB: `pgvector/pgvector:0.8.7-pg17`, internal network
`platform-test-g062-net`, no published port, fresh anonymous volume per run with `10-roles.sh` (8 roles), mode-600 env
`db7.env`; pool slot 1 of 2, no other `platform-test-*` container live before each launch; label `platform-test=g062-a1`;
scripts `.local/tmp/{mkenv7,canaries7,pg7,run7,it7,rununit7,unit7}.sh`; logs
`~/.local/state/platform-evidence/g062-a1/{unit,it}-w7-*.log`). Pre-batch: `git fetch origin g062-impl staging`,
`bash infra/ci/docker-desktop-mounts.sh` exit 0 (17%). Integration commands are `go test -tags integration -count=1 -p 1
-run <selector> <pkg>` (with `-v -timeout 40m`).

| # | Command | Exit |
|---|---|---|
| 41 | DB 1 (fresh), **negative control**: only the `service_test.go` extension applied, `retire.go` unfixed; `-run '^TestEnrollmentPendingCredentialRevocation$' ./internal/identity/` (`it-w7-control-old.log`) | **1 (expected)**: 7/8 subtests PASS; `retire_revokes_both_credentials_and_re-enable_revives_neither` FAIL `signup credential {1,2}: reason "" (want "disabled"), resolves false` after retirement (`:1096`, the user is disabled) and `signup credential {1,2}: reason "" (want "disabled"), resolves true` after re-enable (`:1103`) |
| 42 | unit: `retire.go` fixed, `retire_test.go` **unextended**; `go test -count=1 -run '^TestFinalizeRetirementRevokesEnrollmentInOrder$' ./internal/identity/` (`unit-w7-fix-oldtest.log`) | **1 (expected)**: `execs 4 commits 1, want 3 and 1` |
| 43 | unit: `retire_test.go` extended, `retire.go` temporarily reverted (fix kept as a patch, re-applied after); same selector (`unit-w7-nofix-newtest.log`) | **1 (expected)**: `execs 3 commits 1, want 4 and 1` |
| 44 | unit (fixed): `gofmt -l`, `go vet`, `go vet -tags integration` on `./internal/identity/ ./internal/tenancy/access/` (`unit-w7-check.log`) | 0 |
| 45 | unit (fixed): `go test -count=1 -run '^(TestFinalizeRetirementRefusesBeforeSQL\|TestRetirementCheckView\|TestFinalizeRetirementRevokesEnrollmentInOrder)$' ./internal/identity/` (`unit-w7-identity.log`) | **0** (3/3) |
| 46 | unit (fixed): `go test -count=1 -run '^(TestRLSBoundary\|TestRetireUserFinalize\|TestRetireUserNeedsFinalizer)$' ./internal/tenancy/access/` (`unit-w7-access.log`) | **0** (3/3; `TestRLSBoundary` 6 subtests) |
| 47 | DB 1 reused (fixed source): `-run '^TestEnrollmentPendingCredentialRevocation$' ./internal/identity/` (`it-w7-control-fixed.log`) | 1, invalid (fixture, not product): every subcase `origin_rejected` at fixture sign-in — see deviation 16 |
| 48 | DB 2 (fresh, **fixed source**): same selector (`it-w7-control-fixed2.log`) | **0**: 8/8 subtests incl. **`retire_revokes_both_credentials_and_re-enable_revives_neither` PASS** |
| 49 | DB 3 (fresh, fixed): `-run '^(TestEnrollmentPendingCredentialRevocation\|TestDisableAndEnable\|TestDisableVersusSignInRace\|TestEnrollmentRevocationFaultRollback\|TestRLSIdentityPool)$' ./internal/identity/` (`it-w7-identity-regression.log`) | **0** (5/5 top-level; 16 subtests incl. `TestEnrollmentRevocationFaultRollback/{retire,disable}`) |
| 50 | DB 3 (fixed): `-run '^TestMembership$' ./internal/tenancy/access/` (`it-w7-access-membership.log`) | 1, invalid (harness setup): `canary file: at least three canaries are required` before any SQL (my env had two canaries) |
| 51 | DB 4 (fresh, fixed; env canary line replaced by four, secrets unchanged): same selector (`it-w7-access-membership2.log`) | **0** (32 subtests incl. all seven MEM-14 retire/finalize cases) |

FAIL → PASS pair: #41 (old `retire.go`, exit 1, the predicted `reason "" (want "disabled"), resolves true` ×2) → #48
(fixed, exit 0). Unit mirror: #42/#43 fail as predicted, #45 passes.

Scope proof: `git diff --name-only 51e8c57c HEAD | grep -vEf .local/tmp/allowlist.ere` prints nothing (exit 1) at the
pushed tip; this round's own delta is exactly the three released files plus this Result.

DEFERRED (unchanged policy; not run): B-SIGNUP rows and the `validate-controls.sh` proposal; MinIO/shared-stack access
suites and the rest of the round-5/6 DEFERRED list; `enroll_integration_test.go` (optional fault case, not touched);
B-IDENTITY-IT (`go test -tags integration ./internal/identity/...`) picks the extended subcase up at final coverage.

Resource receipt (round 7): containers `platform-test-g062-{unit,it,pg}` removed with `docker rm -f -v` (PG recreated four
times within one slot, each anonymous volume removed with it); network `platform-test-g062-net` removed; `docker ps -a`,
`docker network ls`, `docker volume ls` with `--filter label=platform-test=g062-a1` list nothing; `db7.env` deleted. No
image built or pulled; no `local-harness.lock`, Compose stack, MinIO, shared fixture, published port or `run.sh` suite.

Deviation (round 7, for Codex review):
16. **Integration reruns use a fresh DB, not the control's DB.** The release asked for the fixed-source integration "on the
    same private DB". The identity package is not rerunnable on one database: TestMain's `releaseFixtureSlugs` sets the
    fixture Organizations' slugs to NULL after a run and `seedTenants` re-inserts them with `ON CONFLICT (id) DO NOTHING`,
    so a second run on the same DB rejects every tenant origin (#47; slugs verified NULL). Each run therefore got a fresh PG
    from the identical stack definition (as rounds 5/6 did). No test or fixture change.

Slice-1 status after round 7: the AC6 / SUP-10 retirement gap is closed and regression-proved (control FAIL #41, fixed
PASS #48/#49/#51). Remaining slice-1 items are the Codex-held/deferred ones only (B-SIGNUP rows and their
`validate-controls.sh` proposal; deferred MinIO/harness selectors) plus the staging reconciliation noted above. Not
self-accepted.

### Round 8: signup harness rows

**Rebinding (recorded before the first write).** Eighth writer (Opus 5.5 high; runtime effort not observable), single write
owner, worktree `.claude/worktrees/agent-<id>`, branch `g062-impl-w8` created from `origin/g062-impl`
`9491e7eba2a1d1b9433bd72a69c76fcd7d4c2fec` (asserted after `git fetch origin staging g062-impl`). Bounded slice within
attempt 1 (budget unchanged).

| Binding | Value (verified) |
|---|---|
| Codex release `management/v0/results/G062-signup-harness-proposal-review-20261021.md` (staging) | `6f8303c491f535d6be44bd719a96b70c0d00699285a843c340a9a20861e3df23` (proposal ACCEPTED; D1-D6; bounded implementation RELEASED) |
| Proposal `~/.local/state/platform-evidence/g062-a1/g062-signup-harness-proposal-20261020.md` | `02c6492b58caf6ad8b978c4311195b361792fb6effe1a17fb33e1f3734887c55` (its 45-pattern / amendment-03-request text is superseded by the release) |
| `management/v0/results/G062-scope-allowlist.txt` (staging `4dfe9ed8`, in the merged tree) | `41c05378966fab11baabd516161df24f95cef5fdd0fe598be57711f8405aadcb` (46 ERE patterns; pattern 46 `^infra/ci/validate-controls[.]sh$`, amendment 03; copy and `#`-stripped pattern file in the worktree's ignored `.local/tmp/`) |
| Plan r5 `management/v0/plans/G062-root-host-tenant-signup-plan.md` | `b2a612a2c50b4a36505346c2c18214cd785d7d62277bdb35c993db49a3a5a3e2` (unchanged) |
| `management/v0/PROGRESS.md` (staging `4dfe9ed8`) | `7290a8cdb4fce84dcbcf1743ad6f061cc15f3d141769b3c9a4dc5d8b591a94c8` (was `b1917fb5…` at round 7) |

**Staging merge (D2: merge, not rebase; separate commit before any harness edit).** `git merge --no-ff --no-commit
origin/staging` with staging at `4dfe9ed8e294c5a37eafa89b781af9ad17f79239`: "Auto-merging infra/validation/run.sh —
Automatic merge went well", **zero conflicts**. 29 paths from staging only (the maintenance delta since `51e8c57c`: archetype
output `adda1ceba47bd16a` manifest rename, `AGENTS.md`, `docs/agents/GITLAB-BRANCH-WORKFLOW.md`,
`docs/foundation/{BOOTSTRAP,OPERATIONS}.md`, `infra/security/gitleaks.toml`, `infra/validation/render-customizations.tsv`,
`infra/validation/run.sh` (+72, S8 / ADR-007 block), `intelligence/uv.lock` and the template `uv.lock` (certifi),
`PROGRESS.md`, the archive and management Results incl. this release and the 46-pattern allow-list, two skill references).
Merged `run.sh`: `bash -n` exit 0; 16 `S8` lines and 17 `00018_platform_signup` lines present (staging hunks and G062 hunks
disjoint). Merge commit `2b1bac65ed6d44122f006e829034f2b621b66a07` (parents `9491e7eb` + `4dfe9ed8`). No runtime/SQL input
moved (`core/`, `web/`, `infra/{postgres,recovery,compose}` unchanged on staging), so the round 1-7 evidence bindings stand.
After the merge the harness `--base` must be the staging tip (`4dfe9ed8` or later), never `51e8c57c`.

**Change (write set exactly as released).**
- `infra/validation/run.sh`: new `signup_checks` (after `enrollment_checks`, before `# B-MCP-IT`), on the
  `enrollment_checks` template: B-SIGNUP-CATALOG `rls_go signup-catalog … ./internal/migrate/ '^TestSignupCatalog$'`
  first (its `reset()` rebuilds 17 -> 18), then B-SIGNUP-IT `rls_go signup-it … ./internal/signup/
  '^(TestSignupProvisioning|TestSignupSlugRace|TestSignupContinuation)$'` with `rls_logcheck` kept (D5). Wired into the
  `migrations` suite after `enrollment_checks` and before `recovery_drill` (line 2581); both IDs in the MinIO-failure FAIL
  list (+2 IDs). Usage text unchanged (lists suites, not IDs); no count pinned in `run.sh`.
- `infra/validation/required-checks.tsv`: `B-SIGNUP-CATALOG\tmigrations`, `B-SIGNUP-IT\tmigrations` after
  `B-ENROLLMENT-IT`; header comment edits only as tabulated below.
- `infra/ci/profile-paths.tsv`: one comment and `foundation\tdir\tcore/internal/signup/` after
  `core/internal/enrollment/`; no adjacent path changed.
- `infra/ci/validate-controls.sh`: only VA7, VU1, VU10, J8, R3a, R8a (9 lines; VU9 and everything else unchanged).

Pin derivation (real PASS lines; `it_count` counts `--- PASS: <Test>/` lines and fails on any SKIP/FAIL in the log):

| Selector | Private log (`~/.local/state/platform-evidence/g062-a1/`) | Static `t.Run` count at HEAD | Pin |
|---|---|---|---|
| `TestSignupCatalog` | `it-w6-fix-migrate.log`: 9 subtest PASS, top-level PASS, `ok core/internal/migrate` | `core/internal/migrate/signup_integration_test.go:130…552`: 9 | `n == 9` |
| `TestSignupProvisioning` | `it-w6-fix-signup.log`: 7 subtest PASS | `core/internal/signup/signup_integration_test.go:399…632`: 7 | `p == 7` |
| `TestSignupSlugRace` | same log: top-level `--- PASS: TestSignupSlugRace` only | `:665-742`: no `t.Run` | `grep -c '^--- PASS: TestSignupSlugRace '` `== 1` (TestMCPFootprint precedent, `run.sh:3564`) |
| `TestSignupContinuation` | same log: 5 subtest PASS; `ok core/internal/signup 1.300s` | `:747…904`: 5 | `c == 5` |

Both test files are unchanged since round 6 (`1a7827cd`), the logs' source.

Count before/after:

| Item | Before | After |
|---|---|---|
| `required-checks.tsv` data rows (`grep -cv '^#'`) | 127 | 129 (migrations 43 -> 45) |
| TSV header line 4 | `… + 2 G019 + 2 G058 = 127 rows` | `… + 2 G019 + 2 G058 + 2 G062 = 129 rows` |
| TSV header line 12 | `G058: B-ENROLLMENT-CATALOG/IT migrations).` | `…; G062: B-SIGNUP-CATALOG/IT migrations).` |
| TSV header line 17 (profiles; computed from `infra/ci/lib.sh` suite lists + 4 final) | `fast 35, integration 77, full 120, staging 5` | `fast 35, integration 75, full 122, staging 5` (77 was stale; computed 75 before and after) |
| VA7 `:1146` / `:1149` | `sum_has required_count 90` / `90 required IDs` | `92` / `92 required IDs` |
| VU1 `:1268` | `vu_split full 35 90 121` / `gate 90 -> 121` | `35 92 123` / `gate 92 -> 123` |
| VU9 `:1269` | `vu_split integration 35 44 75` | unchanged |
| VU10 `:1320` / `:1326` | `vu_split rc 35 90 121` / `gate 90 -> 121` | `35 92 123` / `gate 92 -> 123` |
| J8 `:2317` | `n != 90` / `want 90` | `n != 92` / `want 92` |
| R3a `:5217` / `:5218` | `remainder 90 -> 121` / `(35 + 90 -> 121` | `remainder 92 -> 123` / `(35 + 92 -> 123` |
| R8a `:5433` | `remainder 90 -> 121` | `remainder 92 -> 123` |

Other literal 90/121 hits in `validate-controls.sh` (BuildKit `#90`/`#91` vertices `:3366-3387`, `121 unknown_context` cap
`:3389`) are unrelated and unchanged.

Commands (worktree root; no `local-harness.lock`, no `run.sh` suite, no Compose stack, no build or pull):

| # | Command | Exit |
|---|---|---|
| 52 | `git fetch origin staging g062-impl`; `git rev-parse origin/g062-impl` = `9491e7eb…`; `git switch -c g062-impl-w8 9491e7eb…`; HEAD asserted | 0 |
| 53 | `git merge --no-ff --no-commit origin/staging` (`4dfe9ed8`) | 0, no conflicts |
| 54 | `bash -n infra/validation/run.sh` (merged, before commit); `git commit -F` (merge `2b1bac65`) | 0 |
| 55 | `bash -n infra/validation/run.sh infra/ci/validate-controls.sh` (edited) | 0 |
| 56 | `git diff --check` | 0 |
| 57 | `docker run --rm --pull never --network none -v <worktree>:/repo:ro -w /repo koalaman/shellcheck:v0.11.0 -S warning -x -P SCRIPTDIR infra/validation/run.sh infra/ci/validate-controls.sh` | 0 |
| 58 | same image, `-S info --include=SC2251 -x -P SCRIPTDIR infra/ci/validate-controls.sh` | 0 |
| 59 | `bash infra/ci/docker-desktop-mounts.sh` | 0 (17%) |
| 60 | `bash infra/ci/validate-controls.sh --only VA7,VU1,VU9,VU10,J8,R3a,R8a --log .local/tmp/ci-controls-focused.log` (copied to `~/.local/state/platform-evidence/g062-a1/ci-controls-w8-focused.log`, sha256 `ae362d29…6b52`) | **0: 7 passed, 0 failed (of 7 selected; 521 defined)**; VA7 23 s, VU1 27 s, VU9 22 s, VU10 23 s, J8 9 s, R3a 5 s, R8a 6 s |

Control-run safety (read before launch): the seven controls run the real `validate.sh` / `pipeline-union.sh` /
`release-verify.sh` against the stub harness in disposable scratch git trees under `/tmp/ci-controls-*` (outside the
checkout), fixture repositories and a local bare remote; the docker shim sets `SHIM_PULL_NOOP=1` and the only containers
are `--rm --network none` runs of the locally present pinned `mikefarah/yq:4.48.1` / `zricethezav/gitleaks:v8.30.1`,
labelled with the run token and swept by it. No Compose, published port, shared stack, build or shared fixture. After the
run: `docker ps -aq --filter label=g003.controls.run` empty; no `/tmp/ci-controls-*` left.

Scope proof (at the pushed tip): `git diff --name-only origin/staging HEAD | grep -vEf .local/tmp/g062-patterns.txt` prints
nothing; `git diff --name-only 51e8c57c HEAD | grep -vEf …` prints only paths that arrived through the staging merge (listed
in the handback). This round's own delta: the four released harness files plus this Result.

DEFERRED to the coordinator (shared harness lock, in order): `lint lint-base provenance unit migrations` (the first
attributable run of B-SIGNUP-CATALOG/IT; a pin mismatch there is a mechanical correction with `migrations` rerun only), then
`root ops rename off`; the full `ci-controls` suite and the exact merged-SHA full qualification remain later gates.

Resource receipt (round 8): the two shellcheck containers and the control-run containers were `--rm`; no labelled resource
remains; no image built or pulled; no `platform-test-*` container, network or volume; no lock taken.

Slice-1 status after round 8: the two signup selectors are registered and wired, with the six count controls repinned and
proved by the focused selection. Not self-accepted; Codex source review and the coordinator's lock-held suites are pending.

### Round 9: recovery verify ordering

**Rebinding.** Ninth writer (Opus 5.5 high; runtime effort not observable), single write owner, worktree
`.claude/worktrees/agent-<id>`, branch `g062-impl-w9` from `origin/g062-impl`
`33306fec0e08720a3d7f77f474f6b92251e22552` (asserted after `git fetch origin staging g062-impl`; `origin/staging` still
`4dfe9ed8`). Bounded test correction within attempt 1 (budget unchanged).

**Defect.** Coordinator batch 1 (run `d8067c`, `33306fec`): everything passed except B-RECOVERY-DRILL RD-G
(`TestRecoveryVerify/V6_old_authentication_refused`, `recovery_integration_test.go:610`: the `app.signup_sessions` read by
`credential_hash` returned `no rows in result set`; `recovery_logs_verify.log` sha256 `f76caed9…ecd24`); RD-O and
B-MAINTENANCE-DRILL are cascades. RD-A shows `signup_sessions_expired=1`, so the row was restored and expired.

**Diagnosis (verified by reading the code; confirmed, no product defect).** `infra/recovery/sql/invalidate.sql:27-29`
expires (does not revoke or delete) unrevoked signup rows: `expires_at = least(expires_at, now())`. Every identity login
start (`identity.Store.insertLogin`, `core/internal/identity/store.go:444`) runs `sqlExpireSignupSessions` (`DELETE FROM
app.signup_sessions`, `:90`) in the EMPTY shape, where policy `signup_sessions_expire`
(`core/migrations/00018_platform_signup.sql:527-528`) admits exactly the rows with `expires_at <= now()`. In
`TestRecoveryVerify`, the `fresh(...)` sign-ins before V2 (`signInAt` -> `id.Login`), V3's `fresh("memberA2")` and V6's own
callbacks all start logins, so the invalidated row is correctly swept before V6's read. The `no rows` outcome (not
`f|t|t`) also rules out a revocation path. The assertion was order-sensitive.

**Change (only `core/internal/tenancy/access/recovery_integration_test.go`, only `TestRecoveryVerify`).**
- Right after `one` is defined (before V1 and before any identity call), the signup digest is decoded once and the
  restored row is read through `e.setup` with `signupState` = `SELECT coalesce((SELECT format('%s|%s|%s', revoked_at IS
  NULL, expires_at <= $2, continuation_expires_at <= $2) FROM app.signup_sessions WHERE credential_hash = $1), 'swept')`;
  the binding expectation stays `t|t|t` (a missing row reads `swept` and fails).
- V6 keeps `app.auth_resolve_signup($1, now())` == 0 rows unchanged and reuses the decoded digest; its row read after the
  sign-ins now accepts `t|t|t` or `swept` (unchanged or swept, never revoked/live).

Commands (worktree root; no `local-harness.lock`, no `run.sh`, no stack, no build or pull; unit container: existing
`golang:1.27.2-alpine`, `--network none`, `g058-a1-gomod` read-only, `GOFLAGS=-mod=readonly`, `GOPROXY=off`, disposable
`GOCACHE`, source by `docker cp`, name `platform-test-g062-unit9`, label `platform-test=g062-a1`, no published port;
scripts `.local/tmp/{rununit9,unit9}.sh`; no `platform-test-*` container live before launch):

| # | Command | Exit |
|---|---|---|
| 61 | `git fetch origin staging g062-impl`; `git rev-parse` = `33306fec…` / `4dfe9ed8…`; `git switch -c g062-impl-w9 33306fec…`; HEAD asserted | 0 |
| 62 | `bash infra/ci/docker-desktop-mounts.sh` | 0 (17%) |
| 63 | unit container: `gofmt -l`, `go vet`, `go vet -tags integration`, `go test -tags integration -count=1 -run XXX_COMPILE_ONLY` on `./internal/tenancy/access/` (`~/.local/state/platform-evidence/g062-a1/unit-w9-check.log`, sha256 `89671d48…e9467`) | **0** (gofmt clean; vet clean; integration binary compiles, `[no tests to run]`) |
| 64 | `git diff --check` | 0 |

Deviations: none in scope. The corrected drill itself was **not** run here (it needs the stack and the shared lock).

DEFERRED to the coordinator (shared harness lock): rerun the failed `migrations` coverage that contains B-RECOVERY-DRILL
(RD-G, then the cascaded RD-O and B-MAINTENANCE-DRILL); retain batch 1's other attributable passes.

Resource receipt (round 9): the one unit container was removed (`docker ps -a --filter name=platform-test-` empty); no
image built or pulled; no network, volume or lock. Not self-accepted; Codex review pending.

### Round 10: identity test root origin

**Rebinding.** Tenth writer (Opus 5.5 high; runtime effort not observable), single write owner, worktree
`.claude/worktrees/agent-<id>`, branch `g062-impl-w10` from `origin/g062-impl`
`6a852b88cd1fd04b3bcae9f8de25f913c0b15a3c` (asserted after `git fetch origin g062-impl staging`; `origin/staging`
`3fe83311ec5d6979107c05ec9b1b5e74037dd77c`, not merged here). Bounded test correction within attempt 1 (budget unchanged).

**Defect.** Coordinator batch 2 (`6a852b88`): everything passed except B-IDENTITY-IT
`TestAssuranceIsAlwaysUnknown` (`core/internal/identity/service_test.go:195`: `root signup start: identity:
origin_rejected`; `~/.local/state/platform-evidence/g062-a1/batch2-6a852b88/evidence/diagnostics/identity-it.log`).

**Diagnosis (verified by reading the code; confirmed, no product defect).** The test built `newHarness(t)` without a
root origin, but its G062 SUP-1 block expects one: the positive `Login{Origin: rootOrigin, SignupIntent: true}` and the
`noRoot := h.cfg; noRoot.RootOrigin = ""` contrast. `newHarness` (`integration_test.go:335-352`) applies options to
the same `cfg` it stores as `h.cfg` and passes to `NewService`; `Service.Login` (`service.go:163`) refuses a signup
intent when `cfg.RootOrigin == ""`. `RootOrigin` is not read by tenant resolution (`relay.go:51-63`), so the A-1..A-6
tenant-origin flows and the SUP-1 refusals are unchanged by configuring it (as `TestLoginStart` already states).

**Change (only `core/internal/identity/service_test.go`, only `TestAssuranceIsAlwaysUnknown`).** `h := newHarness(t)`
becomes `h := newHarness(t, withRoot(rootOrigin))` with a one-line G062 comment. Other instances searched: every
`SignupIntent: true` / `signupStart` call in the package's `*_test.go` (only `service_test.go`) is in
`TestCallbackSignup` (`:280`) or `TestCallbackPinned` (`:475`), both already `withRoot(rootOrigin)`; no other genuine
instance.

Commands (worktree root; no `local-harness.lock`, no `run.sh`, no stack, no build or pull; unit container: existing
`golang:1.27.2-alpine`, `--network none`, `g058-a1-gomod` read-only, `GOFLAGS=-mod=readonly`, `GOPROXY=off`, tmpfs
`GOCACHE`, source by `docker cp`, name `platform-test-g062-unit10`, label `platform-test=g062-a1`, no published port;
script `.local/tmp/unit10.sh`; no `platform-test-*` container live before launch; one container at a time):

| # | Command | Exit |
|---|---|---|
| 65 | `git fetch origin g062-impl staging`; `git rev-parse` = `6a852b88…` / `3fe83311…`; `git switch -c g062-impl-w10 origin/g062-impl`; HEAD asserted | 0 |
| 66 | `bash infra/ci/docker-desktop-mounts.sh` | 0 (17%) |
| 67 | unit container, first pass: `gofmt -l` flagged a whitespace slip in this round's own edit (`variants :=[]…`); vet clean; `go test -tags integration -run XXX_COMPILE_ONLY` compiled but its `TestMain` requires the DB settings, so compile-only switched to `go test -tags integration -c` | 1 (fixed) |
| 68 | unit container, rerun after the fix: `gofmt -l internal/identity` (pre-change file also clean), `go vet`, `go vet -tags integration`, `go test -tags integration -c -o /tmp/identity.test ./internal/identity/` | **0** (gofmt clean; vet clean; integration binary compiles, not executed) |
| 69 | `git diff --check` | 0 |

Deviations: compile-only uses `go test -c` instead of `-run XXX_COMPILE_ONLY` (the package `TestMain` exits on missing
DB settings before test selection). The corrected test itself was **not** run here (it needs the stack and the lock).

DEFERRED to the coordinator (shared harness lock): rerun B-IDENTITY-IT; retain batch 2's other attributable passes.

Resource receipt (round 10): each unit container was removed (`docker ps -a --filter name=platform-test-` empty); no
image built or pulled; no network, volume or lock. Not self-accepted; Codex review pending.

### Round 11: staging reconcile and batch evidence

**Rebinding.** Eleventh writer (Opus 5.5 high; runtime effort not observable), single write owner, worktree
`.claude/worktrees/agent-<id>`, branch `g062-impl-w11` from `origin/g062-impl`
`717bc6f9713b557a19dd0bd17c0071b3a2298c98` (asserted after `git fetch origin staging g062-impl`). Bounded reconcile
round within attempt 1 (budget unchanged). No product, test or harness file is edited in this round.

**Reconcile.** `origin/staging` = `eb359bb361ba5004b246e9c8a115c899f5a87833` (merge of MAINT Go 1.27.3 MR !68; not
moved beyond the expected tip). Merge base with the branch: `4dfe9ed8` (the previous G062 staging merge). Staging delta
since `4dfe9ed8` (32 files): Go 1.27.3 pins (`core/go.mod`, `docs/SECURITY.md`, `docs/foundation/VENDORS.md`,
`infra/docker/{core,migrate,minio,validation-tools}.Dockerfile`, `infra/security/dependency-scan.sh` and its
`testdata/{exceptions/go-{expired,valid}.tsv,govulncheck/{broken,module}/go.mod}`,
`infra/validation/{lib.sh,render-customizations.tsv,run.sh,tools/go.mod}`), management records
(`management/v0/PROGRESS.md`, MAINT/RC2 results), SDLC skill docs under `skills/restruct-sdlc/` and the governance text
in `AGENTS.md`, `CLAUDE.md`, `docs/agents/{CLAUDE-SESSION-HANDOFF,GITLAB-BRANCH-WORKFLOW}.md`. The governance files were
not listed in the dispatch's expected delta but are owner/management text outside G062's scope and were taken from
staging unchanged. `G062-scope-allowlist.txt` is unchanged on staging.

`git merge --no-ff origin/staging` (ort) completed with **no conflicts**: `infra/validation/run.sh` auto-merged (staging's
Go-pin hunks and G062's signup hunks do not overlap). Merge commit
`165f12d84ebd0486c27796747b48c634f8e110a5` (parents `717bc6f9` and `eb359bb3`). Every management file other than this
Result equals staging (`git diff --name-only origin/staging HEAD -- management AGENTS.md CLAUDE.md docs/agents skills`
prints only this Result).

**Scope proof.** `git diff --name-only origin/staging HEAD` lists 46 paths; `grep -vEf` against the 46 G062 patterns
prints nothing. The non-`#` lines of `origin/staging:management/v0/results/G062-scope-allowlist.txt` are byte-identical to
the dispatch's 46-pattern file (`diff` empty), so the pattern set did not change.

**PROGRESS binding.** sha256 of `management/v0/PROGRESS.md` at merged staging `eb359bb3` (and at the merge commit):
`3dfec5625517ee7f97ece4f21065baef001f1aab3310eb2fcff177db2e8d8cf5`.

**Coordinator batches (lock-held local harness, base `4dfe9ed8`, allow46; recorded from the coordinator; evidence
read-only under `~/.local/state/platform-evidence/g062-a1/`).**
- Batch 1, run `d8067c` at `33306fec` (`lint lint-base provenance unit migrations`, `batch1-33306fec/evidence`): all PASS,
  including the first attributable B-SIGNUP-CATALOG 9/9 and B-SIGNUP-IT 7/1/5, except B-RECOVERY-DRILL RD-G
  (`TestRecoveryVerify` V6: the invalidated signup row was swept by the login-start `DELETE` before the read; RD-O and
  B-MAINTENANCE-DRILL were cascades). Fixed by round 9 (`6a852b88`).
- Batch 1b at `6a852b88` (`lint lint-base migrations`, `batch1b-6a852b88/evidence`): B-RECOVERY-DRILL,
  B-MAINTENANCE-DRILL, B-SIGNUP-CATALOG and B-SIGNUP-IT PASS. **B-DEPSCAN FAIL**, and the B-REQUIRED-CHECKS rollup FAIL it
  caused: the current-tree scan at 01:19Z reports blocking `Go:stdlib@1.27.2` GO-2026-6809 (CVE-2026-111943,
  reachability called, fixed in 1.27.3; `RESULT fail exit=1 findings=2`, the other finding the already-excepted dev
  `npm:braces@3.0.3`). The advisory was published 00:27:29Z, after batch 1's earlier B-DEPSCAN PASS. Preserved log
  `batch1b-6a852b88/evidence/depscan.log`, sha256 `a097e59c7f82b03066dbb4b1c0f521cdc3f1c465deec8ad9bc3debdba2251774`
  (re-verified this round). The failure is caused by GO-2026-6809 and is independent of G062; it is not excepted. The Go
  1.27.3 pins were integrated separately by MAINT MR !68 (staging `eb359bb3`), which this merge brings in. B-DEPSCAN on
  the merged tip is re-run after this round by the coordinator (lock-held); **it is not claimed to pass here**.
- Batch 2 at `6a852b88` (`root ops rename off`, `batch2-6a852b88/evidence`): root, rename, off and the ops checks PASS
  except B-IDENTITY-IT (`TestAssuranceIsAlwaysUnknown` built its harness without the root origin for its SUP-1 block;
  test fault). Fixed by round 10 (`717bc6f9`).
- Batch 2b at `717bc6f9` (`ops`, `batch2b-717bc6f9/evidence`): RESULT PASSED; B-IDENTITY-IT 93 top-level tests pass
  (`results.tsv` has no FAIL row).

These runs span rounds 9 and 10 (test-only changes), so **no single run covers one tip**. The exact merged-SHA full
local qualification is still required after the MR merge. G062 final ACCEPT is not claimed.

Commands (worktree root; stack-free; no `local-harness.lock`, no `run.sh` suite, no build, pull, stack or container):

| # | Command | Exit |
|---|---|---|
| 70 | `git fetch origin staging g062-impl`; `git rev-parse` = `717bc6f9…` / `eb359bb3…`; `git switch -c g062-impl-w11 origin/g062-impl`; branch/HEAD asserted | 0 |
| 71 | `git merge --no-ff -F <msg> origin/staging` (author/committer `<owner> <<owner-email>>`) | 0 (no conflicts) |
| 72 | `bash -n infra/validation/run.sh infra/ci/validate-controls.sh infra/security/dependency-scan.sh` | 0 |
| 73 | `git diff --check HEAD~1 HEAD` (merge) and `git diff --check` (this Result edit) | 0 |
| 74 | `git diff --name-only origin/staging HEAD \| grep -vEf <46 patterns>` | 1 (no output: every path in scope) |
| 75 | `diff <46 patterns> <non-# staging allow-list>` | 0 (identical) |
| 76 | `sha256sum` of staging `PROGRESS.md` and of `batch1b-6a852b88/evidence/depscan.log` | 0 (values above) |

Deviations: none. Resource receipt (round 11): no container, image, network, volume or lock used. Not self-accepted;
Codex review pending.

### Round 12: batch 3 evidence (static lint, provenance and dependency scan at a567d102)

**Rebinding.** Twelfth writer (Opus 5.5 high; runtime effort not observable), single write owner, worktree
`.claude/worktrees/agent-<id>`, branch `g062-impl-w12` from `origin/g062-impl`
`a567d1026130994b355261ab9eebef4d47ac41e5` (asserted after `git fetch origin g062-impl staging`). Result-only round
within attempt 1 (budget unchanged); no product, test, harness or other management file is edited. Every fact below is
read from the evidence directory
`<home>/.local/state/platform-evidence/g062-a1/batch3-a567d102/evidence/` and the saved coordinator console output
`.../<redacted-uuid>/tasks/bct6kofyz.output` (readable; sha256
`a858f87c2028fbd69359f9859be34523e6b784e771c2b22e1593c6f22966dbdb`).

**Invocation (coordinator, lock-held; checkout detached and clean at `a567d102`; recorded from the coordinator).** A prior
`bash infra/ci/docker-desktop-mounts.sh` check reported 17% mount use (coordinator report; not in the evidence directory).
```
flock -n <home>/.local/state/platform-evidence/local-harness.lock bash infra/validation/run.sh \
  --base eb359bb361ba5004b246e9c8a115c899f5a87833 \
  --allow-file <home>/.claude/jobs/580cdb2f/tmp/g062-allow46.txt \
  --report-dir <home>/.local/state/platform-evidence/g062-a1/batch3-a567d102/evidence \
  lint lint-base provenance
```
Run `2d354c`, base `eb359bb3`, suites `lint lint-base provenance`, areas none; started 22:49:06 local
(`timings.tsv` 1792550946 = 2026-10-21T02:49:06Z), ended 22:53:50 local (1792551230 = 02:53:50Z). Exit code 0,
`RESULT: PASSED`. No build (`builds.tsv` header only; BUILD-LEDGER rows 0, ac5_status MET); `inventory-leftover.txt` and
`deps-codes.txt` are empty.

**Check outcomes (`results.tsv`, identical to the console summary).** `required-checks.tsv` maps 22 rows to the executed
suites; B-REQUIRED-CHECKS reports "21 required IDs of the executed suites (lint lint-base provenance, final) each recorded
exactly once as PASS" (the 21 other IDs below).

| Suite | ID | Outcome (as recorded) |
|---|---|---|
| lint | B-SHELL, B-HADOLINT, B-COMPOSE, B-ENV-EXAMPLE, B-DEPS, B-MIGRATE-SET, B-DOCKERIGNORE, B-CI-CONFIG, B-SECRET-SOURCE | PASS (each) |
| lint-base | B-LITERALS, B-NAMING, B-IDENTITY-GUARDS | PASS (each) |
| lint-base | B-DEPSCAN | PASS: all six phases PASS; "1 findings, 0 blocking unexcepted"; 1 excepted finding kept visible |
| provenance | B-RENDER-ID (output `adda1ceba47bd16a`), B-RENDER-PLAN, B-RENDER-REPEAT (31 registered customizations) | PASS (each) |
| provenance | B-PROTECTED | PASS: 1010 protected files equal the base commit, 0 relocated, 0 expected-hash files |
| provenance | B-SKILL-UNCHANGED | PASS: skill roots identical to `eb359bb3`, 0 delta paths, 0 untracked |
| final | B-DOCKER-ONLY | PASS |
| final | B-SECRET-CANARY | PASS: 19 generated secrets absent; a planted secret is detected |
| final | B-SCOPE | PASS: 46 changed paths since `eb359bb3`, all within the allow-list |
| final | B-REQUIRED-CHECKS | PASS (21 IDs, above) |
| info | PREFLIGHT (docker 27.5.1, compose 2.32.4-desktop.1), PROBE (devtcp), DEPSCAN-PINS, BUILD-LEDGER, IMAGES | INFO (not PASS) |

There is no FAIL or UNMET row. hadolint printed three DL3066 `info` lines below its warning threshold.

**Dependency scan (`depscan.log`, `depscan-results.tsv`, `depscan-current-report.json`).** Scan date 2026-10-21 (UTC;
current-tree scan 02:53:23Z to 02:53:35Z). Pins: osv-scanner
`ghcr.io/google/osv-scanner:v2.6.0@sha256:afd838850ac1a0fcc15ff4a041dc9ba11123c3f0d2666217a5f0fcf9222b55fa`; Go
`golang:1.27.3-alpine@sha256:f2c872bcee949b7be9c490efd02b48a8a2fbf947cdf533f1c58525ec9d8170fd`; govulncheck
`golang.org/x/vuln@v1.8.0` (module hash equals the pin); `GO_STDLIB_VERSION` 1.27.3 (`go version go1.27.3`). Advisory
objects: Go, npm, PyPI OSV zips and `vuln.go.dev/vulndb.zip`, each md5 equal to the published md5; govulndb
`db_last_modified` 2026-10-21T00:27:29Z. Phases fetch, phase-a, vulnpolicy-build, fixtures-govulncheck, fixtures-osv and
current all PASS. Graphs: Go 37 identities (37 locations, toolchain go1.27.3 standard library), npm 472 (478 locations),
PyPI 29 (29 locations); no unsupported, missing or unexplained scanner-only entries; `problems` and `incomplete` empty.
Findings exactly as printed:
```
FINDING excepted npm:braces@3.0.3 ids=GHSA-vfj7-8cjw-p6xm aliases=CVE-2026-93687,GHSA-vfj7-8cjw-p6xm class=High reachability=- sources=osv-scanner groups=dev fixed=
RESULT pass exit=0 findings=1
```
The single finding is the dev-only braces advisory under its existing recorded exception; 0 blocking unexcepted; the Go
standard library has no finding. **No exception was added** by G062 or by this run:
`infra/security/dependency-exceptions.tsv` is byte-identical at `eb359bb3` and `a567d102` (blob
`9b6b511732e11246ede1a1fb001ee25c0232eccf`; `git diff --stat` empty).

Contrast (unaltered): batch 1b's preserved `batch1b-6a852b88/evidence/depscan.log` (sha256
`a097e59c7f82b03066dbb4b1c0f521cdc3f1c465deec8ad9bc3debdba2251774`, re-verified) still records
`FINDING BLOCKING Go:stdlib@1.27.2 ids=GO-2026-6809 ... fixed=v1.27.3`. The cause was fixed by the Go 1.27.3 pins integrated
through MAINT MR !68 (staging `eb359bb3`), which `a567d102` contains (`git merge-base --is-ancestor` exit 0).

**Hashes (sha256).**

| File | sha256 |
|---|---|
| `results.tsv` | `5e68ddaf074a1edd728133ebfbec228ad5fc29fc9824dafc8c79cffabc435bda` |
| `required-checks.tsv` | `a91ce21304a2ad6820d9da3a5fb8da4b3f51c36110b6564f86c787dce1d879d0` |
| `depscan.log` | `acc58472a41864dc3b46d72f4fb4d0dc9cbf6ffaeae187c790b9ee384e7aebdf` |
| `depscan-results.tsv` | `b907befceabcd78b230228d5e3be48a375782286c1b02ef422b944284b401480` |
| `depscan-current-report.json` | `b7a211754096e5c728a32d94b48493d1c22b2bfc0b30a14255f18fcab54f2cd0` |
| `timings.tsv` | `69bd1425694fb661e8a62a9aa4e7b31e79a9d1f66a4a6b7241fdce5fe7bfa423` |
| `ledger-summary.tsv` | `d825c88d7231cf9b1487905ba2134e7f2b488ae52871ab08f1827e174924b57c` |
| `builds.tsv` | `8d9dc31ab713f39f9954b6bff0ea0c8176bae7a6aa2db048744f6ff662c08682` |
| `images.txt` | `61deaddf1e773f7490b457663904b6690b04cf323c037d6c05e63255f086344e` |
| `plan.out` and `repeat-render.out` (identical) | `db0aac3cd4ad71a6fc28c844df6667e6a368903b7c896fea819e1f88ff69a941` |
| `deps-codes.txt` and `inventory-leftover.txt` (empty) | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| allow-file `g062-allow46.txt` | `060156a6b4550c41fdd09778070656acb46bc5b4d9e2a69f4dd7f7fff9a5bfbb` |
| `management/v0/PROGRESS.md` at `a567d102` | `3dfec5625517ee7f97ece4f21065baef001f1aab3310eb2fcff177db2e8d8cf5` |

**Scope of this evidence.** This is static lint, lint-base, provenance and dependency-scan evidence **only**, at
`a567d102`. It is not the deferred local suites (ci-controls, contracts, controls), not any stack suite, and not the final
exact merged-SHA full local qualification, which is still required after the MR merge. The historical failed-run entries
(batch 1 B-RECOVERY-DRILL failure, batch 1b B-DEPSCAN failure, batch 2 B-IDENTITY-IT failure) and their later reruns remain
as recorded in round 11. G062 final ACCEPT is not claimed.

Commands (worktree root; no `run.sh`, no lock, no container, build or pull):

| # | Command | Exit |
|---|---|---|
| 77 | `git fetch origin g062-impl staging`; `git rev-parse origin/g062-impl` = `a567d102…`; `git switch -c g062-impl-w12 origin/g062-impl`; branch/HEAD asserted | 0 |
| 78 | read evidence files and console output; `awk` over `required-checks.tsv` for the executed suites | 0 |
| 79 | `sha256sum` of the files above, allow-file, `PROGRESS.md` and the batch 1b `depscan.log` | 0 |
| 80 | `git diff --stat eb359bb3 a567d102 -- infra/security/dependency-exceptions.tsv`; `git merge-base --is-ancestor eb359bb3 a567d102` | 0 (empty diff; ancestor) |
| 81 | `git diff --check`; `git diff --name-only origin/g062-impl HEAD` | 0 (only this Result) |

Deviations: none. Resource receipt (round 12): no container, image, network, volume or lock used. Not self-accepted;
Codex review pending.

### Round 13: slice-1 integration, merged-SHA qualification and slice-2 status

**Rebinding.** Thirteenth writer (Opus 5.5 high; runtime effort not observable), Result-only round within attempt 1
(budget unchanged); worktree `.claude/worktrees/agent-<id>`, branch `g062-s1-result` created from
`origin/staging` after `git fetch origin staging`. The branch was first created at `4e0777b38c258fff704cc7ee138fa158a29b0694`;
a re-fetch showed staging had advanced to `50982d146534567597bf3503420d5639123f7303` (Codex's slice-2 execution release,
management-only), and the empty branch was fast-forwarded to it before any edit. No product, test, harness or other
management file is edited. Facts below are verified against git, the GitLab API (read-only) and the read-only evidence
under `<home>/.local/state/platform-evidence/g062-a1/`, except where marked as coordinator-reported.

**Feature MR !70 and integration.**

- MR !70 (`g062-impl` → `staging`), head `243f9ce6508799308dc91c36db11a3940af852ab`, opened 2026-10-21T03:01:32Z after Codex's
  bounded slice-1 source ACCEPT at `a567d1026130994b355261ab9eebef4d47ac41e5`
  (review record (not included), sha256
  `2302314b1d48bec50f293545ac0ca32ff4eabd6c4d0617b69cf73da3e757ea09`). `git diff --name-only a567d102 243f9ce6` lists only this
  Result (the round 12 evidence section).
- Hosted fast pipeline `2970186651` on the exact head: **success** (`validate:checks` job `17335643877` success;
  `validate:integration` job `17335643878` manual, not run, by design for feature MRs).
- GitLab Duo reviewed and approved: one non-resolvable summary note, "found nothing to comment on"; it states its diff was
  truncated to about 28% (web flow, signup API adapter and the later Result not seen). Resolvable/open threads: 0.
- Reconciliation without a pushed merge commit. The MR description recorded a dry-run against staging `d3eaf7ed` (tree
  `6feac2520c20d376dc1547e9d01da07398df523b`). Staging then advanced to `a6b5cddc7ddc89a19cd6df72f13a7f0ba50c0bab`; the
  staging-side delta since the merge base is management-only (`management/v0/PROGRESS.md` and the slice-1 review record),
  disjoint from the branch. `git merge-tree --write-tree a6b5cddc 243f9ce6` = `879def13432bac3e0e17dbc0c5d389dabe83f64d`
  (clean; re-verified this round). In an earlier round a subagent's attempt to commit a staging merge into the branch was
  refused by the auto-mode classifier; it was not retried or worked around, and the PROGRESS rebinding is recorded here
  instead (coordinator-reported; that unpushed commit is listed under held resources).
- Merged under the standing staging auto-merge rule (`glab mr merge 70 --auto-merge --sha 243f9ce6… --remove-source-branch`,
  coordinator-reported command) at 2026-10-21T03:12:17Z: merge commit `c5688bf94e1fa4080145f5f6de22005047853ec2`, parents
  `a6b5cddc7ddc89a19cd6df72f13a7f0ba50c0bab` and `243f9ce6508799308dc91c36db11a3940af852ab`, tree
  `879def13432bac3e0e17dbc0c5d389dabe83f64d` (equals the predicted merge tree). Source branch deleted
  (`git ls-remote origin g062-impl` empty).

**Merged-SHA full local qualification.**

- First launch (kept as history): `flock -n <home>/.local/state/platform-evidence/local-harness.lock bash
  infra/ci/validate.sh --profile full --evidence-dir .../g062-a1/postmerge-c5688bf9` was refused at preflight, exit 16
  (coordinator-reported): the frozen checkout held 5 ignored files in `.local/tmp` (the coordinator's own proposal/message
  copies). The harness did not run. Its `storage.tsv` (job `local-57a07c2cf631`, candidate `c5688bf9…`, sha256
  `08803d8f87cbb6f03f23aca937416a008c67b6ae8728232dc5f7d747da635bfb`) records `cleanup_status` 0 and `owned_remaining` 0. The
  five files were moved intact to `<home>/.local/state/platform-evidence/coordinator-local-tmp-20261021/tmp` (present
  this round: `certifi-repair-proposal-claude-20261020.md`, `g062-r0prime-readiness-20261020.md`,
  `g062-retire-signup-revocation-proposal-20261020.md`, `gitleaks-adr007-allowlist-proposal-20261020.md`,
  `maint-duo-correction-msg.txt`).
- Second launch, same command with `--evidence-dir .../g062-a1/postmerge-c5688bf9-r2`, on the clean detached checkout at
  `c5688bf9` (job `local-54a6260f0649`, 03:14:06Z to 04:27:09Z). `console.log`: `tested_content=local
  head=c5688bf94e1fa4080145f5f6de22005047853ec2`; profile pipeline/executed `full`; suites executed `[lint lint-base lint-web
  lint-core controls ci-controls provenance contracts root migrations ops rename off]`, coverage and partition complete;
  harness `infra/validation/run.sh` (deadline 14400s) exit status 0; every phase status 0 (location, recovery, preflight,
  state, harness, postcheck, cleanup, raw_scan, complete, stage, summary, junit, candidate_scan, report, scrub); junit
  summary "123 cases, 0 failing (supplementary)"; cleanup verified (0 resources removed), `owned_remaining=0`;
  `final status 0`.
- Check rows: 129 `validate: check` rows = **122 PASS** + 7 INFO (PREFLIGHT, PROBE, DEPSCAN-PINS, BACKGROUND ×2,
  BUILD-LEDGER, IMAGES); **no FAIL or UNMET**. Rows include B-REQUIRED-CHECKS, B-SCOPE, B-SECRET-CANARY, B-DOCKER-ONLY,
  B-SIGNUP-CATALOG, B-SIGNUP-IT, B-RECOVERY-DRILL, B-IDENTITY-IT, B-CI-CONTROLS, B-DEPS-CONTROLS, B-WRAPPER-CONTROLS and the
  six B-CONTRACT-* rows, each PASS. The previously deferred `ci-controls`, `contracts` and `controls` suites are covered
  by this run.
- Evidence limitation (as Codex recorded): the wrapper scrubbed raw/staged/state files (`published=none`), so the
  `junit.xml` and per-suite logs are not retained; only `console.log` and `storage.tsv` remain.
- Codex independently accepted this qualification **for slice 1 only**
  (post-merge review (not included), sha256
  `9e96e7d25a9158f418950167c3f77a9dd132de29f5af94b3c5949587bf272330`; its recorded console/storage hashes match below).

| File | sha256 |
|---|---|
| `postmerge-c5688bf9-r2/console.log` | `41750b691b32a6eb62c158a412dc0b772b97400aacd9324a61d6ca3bdef7c78a` |
| `postmerge-c5688bf9-r2/storage.tsv` | `3415c8bae3770212e5965ca62d4b2542f40008378ddd2bcdb5ee44acffa125ca` |
| `postmerge-c5688bf9/storage.tsv` (refused first launch) | `08803d8f87cbb6f03f23aca937416a008c67b6ae8728232dc5f7d747da635bfb` |

**What this does not cover.** AC1, AC7 and AC8 and the slice-2 scope below; other criteria still lack their required
end-to-end proof (per the post-merge review). G062 final ACCEPT is held; the Goal completion definition is not satisfied;
closeout/cleanup and G020 release are not authorized by this evidence.

**Slice 2 status and binding.** Correction to the dispatch premise for this round: staging `50982d14` adds Codex's
slice-2 execution release (not included) (sha256
`f88acb4ebe594bb1e229519131a36ecc769264bc42086f08a0242dc2df44bf4f`): "RELEASED. No planner delta or replan is required";
attempt 1 within B5, no reset or extension, Goal not accepted; bound to staging `4e0777b3`. Slice 2 has **not started**;
no slice-2 source exists on staging. Remaining scope, quoted from accepted Plan r5 §7 steps 4–6: "**Slice 2:**
`signup/api` + limiter copy + HTTP tests; `hosts/resolve`/proxy/host.ts/config removal of `MARKETING_SITE_URL` +
`.env.example` re-pin; landing/signup pages; overlay/playwright changes; E-DOM-4 rewrite + `signup.spec.ts`; b0, b3. 5.
M1; fixes; affected reruns. 6. Docs; PG-009; Result; final local coverage; fetch/rebase; push; MR (fast + Duo + Codex
source review) → conditional merge → merged-SHA full local → cleanup receipt → final ACCEPT." Open acceptance criteria:
AC1, AC7 (browser continuity depends on the web hops), AC8 (`[E-SUP-6]` and M1/manual feedback).

| Binding | Value |
|---|---|
| Attempt / budget | attempt 1 of B5, unchanged |
| Staging base of this round | `50982d146534567597bf3503420d5639123f7303` (release bound to `4e0777b38c258fff704cc7ee138fa158a29b0694`; code/infra diff from `c5688bf9` empty) |
| Plan r5 | `b2a612a2c50b4a36505346c2c18214cd785d7d62277bdb35c993db49a3a5a3e2` |
| Goal / Spec | `e747807c83875d8928b8c505fff83a4a2fe65d716b15d5ed64f24eb72560118a` / `3c90ccb6385768e598d0cc61afdbf3b33d79999b94217e01988c2a2316154bb3` |
| `G062-scope-allowlist.txt` (46 patterns) | `41c05378966fab11baabd516161df24f95cef5fdd0fe598be57711f8405aadcb` (unchanged) |
| `PROGRESS.md` at `4e0777b3` / at `50982d14` | `78bed970d3c85c51d00cccd831e175451f59ebfc1ca7da7b8d11208db64a468b` / `775ace50cfd62288b0cd3088345abf375164a810b88bfa3386d7a51f4f810dcc` |
| This Result before round 13 (at `4e0777b3` and `50982d14`) | `fa938d007bc10edb33ac4d78598b512bdcc39f23afa3508748daec7c05d66c50` (equals the release record's binding) |

Any slice-2 writer must branch from current staging, not from the deleted `g062-impl`, and reconcile these bindings first.

**Held resources (nothing deleted).** G062 writer and planner worktrees, local `g062-impl-w*` branches, the `g062-a1`
evidence and the coordinator's frozen checkout. Writer #12's worktree holds an unpushed local staging-merge commit
`9b38ddd5` plus a staged Result edit (classifier-refused push; coordinator-reported, not inspected this round);
discarding it needs the owner's in-session approval at closeout.

Commands (worktree root; no `run.sh`, no lock, no container, build or pull):

| # | Command | Exit |
|---|---|---|
| 82 | `git fetch origin staging`; `git switch -c g062-s1-result origin/staging` (`4e0777b3`); branch/HEAD asserted | 0 |
| 83 | `git log`/`git diff --name-only`/`git merge-tree --write-tree` for `c5688bf9`, `243f9ce6`, `a6b5cddc`, `d3eaf7ed`; `git ls-remote origin g062-impl`; `git merge-base --is-ancestor c5688bf9 origin/staging` | 0 |
| 84 | `glab api` reads of MR 70, pipeline 2970186651 jobs and MR discussions (read-only) | 0 |
| 85 | read and `sha256sum` evidence files and bound management files; `grep -c` check rows and allow-list patterns | 0 |
| 86 | `git fetch origin staging` (staging now `50982d14`); `git merge --ff-only 50982d14` on the empty branch; HEAD asserted; `git diff --stat c5688bf9 HEAD -- core web infra contracts intelligence` empty | 0 |
| 87 | `git diff --check`; `git diff --name-only origin/staging HEAD` | 0 (only this Result) |

Deviations: the dispatch asked to record slice 2 as not released; staging `50982d14` now records the Codex slice-2 release,
so the header and this section record that instead. Resource receipt (round 13): no container, image, network, volume or
lock used. Not self-accepted; Codex review pending.

### Round 14: slice 2 round A (core API)

**Rebinding.** Fourteenth writer (Opus 5.5 high; runtime effort not observable), single slice-2 implementation owner for
round A (core only), attempt 1 of B5 (no reset or extension); worktree `.claude/worktrees/agent-<id>`,
branch `g062-impl2` created from `origin/g062-s1-result` = `186b1ca0bb49578764732b8019c78776255935ec` after `git fetch
origin staging g062-s1-result`; branch and HEAD asserted before any edit. `origin/staging` = `50982d14` (the base's parent;
unchanged at the pre-batch re-fetch, so no reconcile). Verified: Plan r5 `b2a612a2…`, Goal `e747807c…`, Spec `3c90ccb6…`,
slice-2 release record `f88acb4e…`, `PROGRESS.md` `775ace50…`, this Result before round 14 `7d899548…`. The 46-pattern
extract used for the scope check (`grep -v '^#'` of `G062-scope-allowlist.txt`, file `41c05378…`) is byte-identical to the
coordinator's `g062-allow46.txt` (sha256 `060156a6…`; 46 lines).

**Inventory (existing slice-1 code reused, not redone).** `core/internal/signup` already held every service operation the
API needs (`Resolve`, `StatusOf`, `SuggestSlug`, `Provision`, `Challenge`, `Prove`, `Complete`, `Abandon`, closed codes,
`LogRequest` for `signup.request`); identity already held the slice-2 identity items of Plan §2/§4.2/§4.5: the `signup`
login intent (exact root origin only), the 202 signup callback answer, the pinned `tenant_created` callback,
`SignupCredential`/`SignupCSRFMatches`/`SignupIdentityTx`/`SignupContinuationTx`/`LoginPinned`/`RecordSignupDenial`, the
root login-start (2/s, 20) and root-origin callback (per binding 1/s, 5; global 5/s, 50) buckets with their limiter copy
and `TestRootCallbackRateLimited`, `returnPaths` + `/signup/details`. `events.json` already catalogues `signup.request`. No
identity file needed a change. Missing and added in this round: the HTTP adapter, its limiter copy, the root-origin
helpers, the `MARKETING_SITE_URL` removal, the `public_entry` contract change and the server wiring (identity's
`RootOrigin` was still unset, so root signup was disabled fail-closed until now — review note 6).

| Path | Change |
|---|---|
| `core/internal/signup/api/http.go` (new) | `/internal/v1/signup/*` (Plan §4.2): `GET status?origin=` → `{status, csrf_token, organization?: {slug, target_origin}}`; `POST slug {origin, csrf_token, name}` → `{suggested\|null, available}`; `POST organizations {origin, csrf_token, name, slug}` → `201 {outcome, organization: {id, slug}, continue}`, `409 conflict {field: slug_taken\|intent_used}`, `400 invalid_request {field: name\|slug}`, `403 csrf_failed\|user_disabled`, `401`, `429`, `503 signup_unavailable`; `POST prove {origin, nonce}` (Signup credential, no CSRF member: D-15) → `{complete}` / `403 signup_context_mismatch`; `POST challenge {origin, ref}` and `POST complete {origin, challenge, proof}` (no credential — an Authorization header is ignored; `X-Login-Binding` forwarded) → `{challenge: {value, max_age}, prove}` / the ordinary login answer `{authorization_url, binding: {value, max_age}}`, else `404 continuation_invalid`; `POST abandon {origin, csrf_token}` → `{}`. Exact root origin for root routes (`origin_rejected`), `Signup` scheme only (Session/Enrollment/other = no credential = 401), JSON ≤ 8 KiB, exactly one object with the route's exact members, one `application/json` (charset utf-8 only), no query on mutations; no-store, nosniff, no CORS; closed method/path answers; one `signup.request` record per request (conflicts logged as their field) |
| `core/internal/signup/api/limiter_copy.go` (new) | byte-pinned copy of the `enrollment/api` limiter; buckets (§4.5 (iii)-(iv)): per credential 5/s burst 10 on every root route, per user (`Handle.UserID()`) one create per 10 s burst 5 (slow-refill construction in `http.go`; `allow`/`reclaim` pinned), per tenant origin 5/s burst 20 for each hop separately; keys are scoped SHA-256 digests |
| `core/internal/signup/api/http_test.go` (new) | stack-free httptest + fake service: `TestSignupHTTP` (50 cases: shapes, closed codes, headers, call sequence, one closed log record without credential/secret/name/slug/origin), `TestSignupForwarding`, `TestSignupMediaType`, `TestSignupUnconfigured`, `TestSignupRoutesPinned`, `TestSignupCredentialBoundary` (identity session route refuses `Signup`; signup routes refuse `Session`), `TestSignupRateLimited` (every bucket incl. re-minted credentials of one user and per-hop/per-origin separation), `TestSignupLimiterChurn`, `TestSignupCopyIsExact` |
| `core/internal/signup/signup.go` (+`_test`) | `Service.RootOrigin()` accessor (nil-safe) so the adapter shares the service's root origin |
| `core/internal/hostname/hostname.go` (+`_test`) | `PlatformRootOrigin(cfg)` (`https://<B>[:port]`, `""` without B) and `IsPlatformRoot(h, cfg)` (B only, never `www.B`); `TestPlatformRoot` |
| `core/internal/config/domains.go`, `domains_config_test.go` | `Domains.RootOrigin()`; `MARKETING_SITE_URL` removed (field, `domainVariables`, production requirement, URL validation, doc comment, tests); `TestMarketingSiteURLRemoved` (production loads without it; a still-set value, well-formed or not, is ignored) |
| `core/internal/domains/api/http.go`, `http_test.go` | `hosts/resolve` `public_entry` → `{kind: "public_entry", signup_origin: <root origin>}` for both B and `www.B`; no `location`; `TestResolveRoute` cases for B, `www.B`, `http://`, uppercase `WWW` |
| `core/cmd/server/main.go` | `rootOrigin := domCfg.RootOrigin()` → `identity.Config.RootOrigin` (arms the signup intent and the root start/callback limiters of `identity.NewHandler`); signup service (`Identity`, `Hostname: domCfg.Hostnames()`, `RootOrigin`) and `signupapi.New(...).Register(mux)`; doc comment. New sha256 `db0e4c151015b8ce3fd3015a5ab968484ab1edc19ef681b240f98e5cbfb6797c` |

Code commit `79a038ad1fa3fce9ef0d6d2a0179f262ba5c8491` (tree `635880f0…`); every changed path matches the 46 patterns
(`git diff --name-only origin/g062-s1-result HEAD | grep -vEf allow46` empty).

Decisions and deviations (no contract change; for Codex):
1. `status` omits the Plan's optional `suggested_slug?` (no name exists at status time; `POST /signup/slug` supplies the
   suggestion).
2. The adapter's closed codes are the `signup.request` reason set: a repeated header/parameter, a wrong media type and a
   malformed body are all `invalid_request` (no `ambiguous_parameter`/`unsupported_media_type`, which events.json does
   not list); `conflict` is logged as its field.
3. The per-credential bucket covers all five root routes (Plan names provisioning, own-status and prove explicitly; slug
   and abandon share it). `challenge` and `complete` each have their own per-origin bucket (Plan: "per tenant-origin
   bucket" for both; separate buckets keep one signup's two hops from competing).
4. `IsPlatformRoot` takes the configuration (`IsPlatformRoot(h, cfg)`); it has no production caller in round A (the web
   host gate is round B).
5. **Question (contract gap, round B):** an `already_created` re-mint needs the identical `{name, slug}` (intent digest),
   but `status` returns only `{slug, target_origin}` and the stored Organization name is not readable through
   `auth_resolve_signup` (00018, accepted). The `/signup/details` "Try again" and the automatic re-mint of Plan §4.4 can
   therefore not reproduce the payload from status alone. Options for Codex: (a) web carries the submitted name (e.g. a
   short-lived host-only root cookie or hidden form state) — no core change; (b) a credential-keyed re-mint route —
   contract change; (c) status returns the name — needs a 00018 reader change. Round A implements the Plan contract
   unchanged.

Commands (worktree root; plain native Docker; one container at a time; `golang:1.27.3-alpine`
`sha256:f2c872bc…` (local, not pulled; `core/go.mod` `go 1.27.3`), `--network none`, `g058-a1-gomod` read-only at
`/go/pkg/mod`, `GOFLAGS=-mod=readonly`, `GOPROXY=off`, `GOTOOLCHAIN=local`, tmpfs `GOCACHE`, label
`platform-test=g062-a1`, source by `docker cp` of `core/`; script `~/.local/state/platform-evidence/g062-a1/gen/unit-w14.sh`:
`gofmt -l .`, `go vet ./...`, `go vet -tags integration ./...`, `go build ./cmd/server ./cmd/migrate`, `go test -count=1`
on `./internal/signup/... ./internal/identity/ ./internal/hostname/ ./internal/config/ ./internal/audit/
./internal/domains/... ./internal/tenancy/ ./internal/tenancy/access/ ./cmd/server/`):

| # | Command | Exit |
|---|---|---|
| 88 | `git fetch origin staging g062-s1-result`; `git switch -c g062-impl2 186b1ca0`; branch/HEAD asserted; allow-list extract `diff` | 0 |
| 89 | pre-batch: `git fetch origin staging` (`50982d14`, unchanged); `bash infra/ci/docker-desktop-mounts.sh` (18%); `docker ps -a --filter name=platform-test-` empty | 0 |
| 90 | unit batch run 1 (`unit-w14-run1.log`) | 3: `gofmt -l` listed `internal/signup/api/http_test.go` (one function to be split over lines) |
| 91 | after the gofmt fix: run 3 (`unit-w14-run3.log`; run 2 only printed the diff) | 0: every package `ok` |
| 92 | verbose listing of the round's packages (`unit-w14-run4-verbose.log`, sha256 `6e5746d1…`) | 0: 60 PASS, 0 FAIL (incl. the nine `signup/api` tests, `TestPlatformRoot`, `TestMarketingSiteURLRemoved`, `TestResolveRoute`) |
| 93 | final run on the committed tree `79a038ad` after a comment-only reflow (`unit-w14-run5.log`, sha256 `9dcbfbd1…`) | 0: gofmt clean, vet (plain and integration tag) clean, both binaries build, `signup`, `signup/api`, `identity`, `hostname`, `config`, `audit`, `domains`, `domains/api`, `tenancy`, `tenancy/access` ok |
| 94 | `git diff --check`; scope check of the changed names against the 46 patterns | 0 / empty |

Resource receipt (round 14): container `platform-test-g062-w14-unit` (two instances, sequential) removed with `docker rm -f
-v`; `docker ps -a --filter label=platform-test=g062-a1` lists nothing. No database, Compose, browser stack, image build or
pull, published port, `local-harness.lock` or `run.sh`; the module volume was mounted read-only only. No integration test
was written in this round (every new surface is stack-free; the allow-list admits no `signup/api` integration file) and
none was run.

**Proofs this round makes necessary (coordinator; names/selectors only; nothing edited here).**
- `B-MCP-SOURCE` (`static_mcp_source`, run.sh) **will fail until updated**: `MAIN_GO_G016_SHA` → `db0e4c15…` (re-pin to the
  final main.go); `MAIN_GO_IMPORTS` gains `core/internal/signup core/internal/signup/api` (after
  `core/internal/onboarding/api`); a `CORE_SIGNUP_ROUTES` pin of the seven pairs read from
  `core/internal/signup/api/http.go` (`GET /internal/v1/signup/status|POST /internal/v1/signup/abandon|POST
  /internal/v1/signup/challenge|POST /internal/v1/signup/complete|POST /internal/v1/signup/organizations|POST
  /internal/v1/signup/prove|POST /internal/v1/signup/slug`), an importer check (`core/internal/signup/api` → `core/cmd/server`
  only) and the PASS text.
- `unit`/`lint` (core-check `gofmt`/`go vet ./...`/`go test ./...`): picks up the new `signup/api`, `hostname`, `config` and
  `domains/api` tests automatically.
- `core_env` in run.sh still sets `MARKETING_SITE_URL` (now ignored by core; Plan §2 drops it) and `B-ENV-EXAMPLE` needs the
  `.env.example` re-pin once round B/C removes lines 80–86.
- The web host gate (`web/lib/host.ts`, `proxy.ts`) still expects `location`: until round B lands, `[E-DOM-4]` and any root
  landing check on this branch fail by design; the identity browser line (`identity:browser`, `domain-platform`) and
  `B-IDENTITY-E2E` must run after round B. With `RootOrigin` now configured, the overlay's core accepts the signup intent
  at `https://localhost:543` and charges the root callback buckets there for the first time (`B-IDENTITY-IT`/`-E2E`,
  `B-IDENTITY-PROBES`/`-PRIVACY` rerun after round B, with `ref`/`n`/`p` added to the redaction probe list per AC9).
- SUP-3 remainder (integration): the enrollment adapter's and MCP's refusal of a `Signup` credential (Plan: E `TestEnrollment`
  credential-as-session case, `core/internal/tenancy/access/enrollment_integration_test.go`; MCP env/integration tests) is
  not covered by a stack-free test here; confirm or extend under `B-ENROLLMENT-IT`/`B-RLS-IT`.
- Unchanged by this round (no service/SQL edit): `B-SIGNUP-CATALOG`, `B-SIGNUP-IT`, `B-RECOVERY-DRILL`, migrations rows.

**Remaining slice-2 work seen (rounds B/C):** web BFF (`proxy.ts`, `lib/host.ts` `signup_origin`, `session.ts`,
`identity-client.ts`, `signup-client.ts`, `/` landing, `/signup` + `details` + the six route handlers, callback 202 handling,
`sign-in` status, components, unit tests), overlay (`identity.yml` `--redirect-hosts`/`MARKETING_SITE_URL`/marketing host,
`proxy.conf`, `emulator-seed.yaml`), `.env.example` + run.sh re-pins (`ENV_DOMAIN_BLOCK_SHA`, `ENV_IDENTITY_DOMAIN_TAIL_SHA`,
`WEB_ROUTES` +8, `core_env`, B-MCP-SOURCE above, browser matrices, `identity:browser` projects), `playwright.config.ts`,
`signup.spec.ts` + `[E-DOM-4]` rewrite, docs (§9) and PG-009, then b0/b3, M1, final coverage, MR. The question in decision 5
should be settled before the web re-mint path is built.

Not self-accepted; Codex review pending.

### Round 15: slice 2 round B (web BFF)

**Rebinding.** Fifteenth writer (Opus 5.5 high; runtime effort not observable), single slice-2 implementation owner for
round B (web only), attempt 1 of B5 (no reset or extension); worktree `.claude/worktrees/agent-<id>`, branch
`g062-impl2-b` created from `origin/g062-impl2` = `7c24e0cc7da1d7849c9619149b3b655b51ad8505` after `git fetch origin
staging g062-impl2`; branch and HEAD asserted before any edit. Staging reconcile at the safe boundary before the first
write: `git merge --no-ff origin/staging` (`e5e9c7204335630b476d82573f02fe3ff11c0704`, management-only: Codex's Round 14
review record and PROGRESS) as its own commit `c36690a0075d4c019821a42d8ac4ad623dc2c96b`; `git diff --name-only HEAD~1
HEAD` = `management/v0/PROGRESS.md`, `management/v0/results/G062-round14-remint-contract-review-20261021.md` (no code).
Based on `PROGRESS.md` `16968365f3159c96cd5fbd2ec9a3ce1d6d5aa2c0382aa99b817191813b97b929`; Plan r5 `b2a612a2…`
(unchanged); this Result before round 15 `532ef9fc…`; Codex slice-2 release `f88acb4e…` and Round 14 decision
`1a9e3b4d…` read from `origin/staging`. The 46-pattern extract (`grep -v '^#'` of the staging allow-list) is
byte-identical to the coordinator's `g062-allow46.txt` (46 lines). Binding decision: Codex Round 14 **option A** (host-only
root form restoration cookie; no core, schema or API change).

**Inventory.** Already present (reused, not redone): core's seven `/internal/v1/signup/*` routes and closed codes (Round
14), identity's `signup` login intent, 202 signup callback answer and pinned `tenant_created` callback (slice 1),
`public_entry {kind, signup_origin}` (Round 14). On the web side nothing of the signup surface existed: no
`signup-client.ts`, no `/signup` pages or routes, no `components/signup`, `host.ts` still read `location` and `proxy.ts`
still 303'd the public entry to marketing. The callback route read every completion through G058's
`enrollment-client.ts completeCallback`, which refuses both new answers (the root's 202 `signup` and the tenant's
`tenant_created`) as malformed. `sign-in/page.tsx` needed no change (its closed statuses already cover the tenant
callback's `failed`/`canceled` outcomes; the root's `/sign-in` is the gate's 404).

| Path | Change |
|---|---|
| `web/lib/host.ts` | `public_entry` → `{kind, signupOrigin}` (exactly `signup_origin`, an exact lower-case https origin; `location` or any extra member is malformed; every kind now has its exact member set); `isHttpsOrigin`; gate path sets `publicEntryAllows(path, root)` (root: `/`, `/signup`, `/signup/details`, `/signup/start`, `/signup/create`, `/signup/abandon`, `/signup/prove`; www: `/` only; both the shell's static `/fonts/`, `/blueprint/`), `isSignupPath`, `tenantAllowsSignup` (only `/signup/continue`, `/signup/complete`, platform hosts only) |
| `web/proxy.ts` | marketing redirect removed; public entry served per the sets above (root = own origin equals `signup_origin`); `/signup/*` on tenant hosts 404 except the two hops on platform hosts |
| `web/lib/session.ts` | cookie kinds `signup` (`__Host-signup`/`signup`, ≤ 1800 s), `challenge` (`__Host-sct`/`sct`, ≤ 120 s), `signup_form` (`__Host-signup-form`/`signup-form`): `setSignupFormCookie`/`readSignupFormCookie` — value `v1.<base64url UTF-8 name>.<slug>`, name 1-800 bytes, well-formed, no NUL; slug empty or one host label; canonical base64url and fatal UTF-8 on read; exactly one cookie; Path=/, no Domain, HttpOnly, SameSite=Lax, Secure (https), Max-Age 1800 |
| `web/lib/identity-client.ts` | `login(..., {intent: "signup"})` (no `return_to` with it); `rate_limited` closed code |
| `web/lib/signup-client.ts` (new) | the seven core calls (`Signup` scheme on root routes; hops send only `X-Login-Binding`), strict answer shapes and closed refusals/fields; continuation URL rules (create → exactly the created slug's tenant origin `/signup/continue?ref=`; challenge → exactly the root `/signup/prove?n=`; prove → one platform-subdomain tenant of the root `/signup/complete?p=`); `authCallback` (G058's answers as `completeCallback` validates them + root 202 `signup` ≤ 1800 s with `next: /signup/details` + `tenant_created` with `return_to "/"`; `signup_user_mismatch`, `rate_limited` closed); root resolution `rootOfRootRequest` (root = public entry whose `signup_origin` is the own origin), `rootOfTenantRequest` (candidate `https://<PLATFORM_BASE_DOMAIN><own port>`, accepted only when core resolves it as that `signup_origin`), `publicEntryOf`; `hopSecret` (raw query exactly `<name>=<43-char secret>`); closed notices |
| `web/app/page.tsx` | https public entry (root and www) → `Landing` in the entry frame; tenant entry shows the closed "Your organization is ready." for `?status=tenant_created`; loopback generic page unchanged |
| `web/app/auth/callback/route.ts` | root (`signup_origin` = own origin): `authCallback` → 202 signup sets `__Host-signup` (≤ 1800 s) and 303 `/signup/details`; refusals 303 `/signup?status=<closed>`; any non-signup answer sets nothing; www 404 without a callback; tenant hosts: `tenant_created` → session cookie + 303 `/?status=tenant_created`; a stray signup answer → `/sign-in?status=failed`, nothing set; join-cookie clearing unchanged |
| `web/app/signup/page.tsx`, `details/page.tsx` (new) | root-only pages (fixed 404 elsewhere): start; details = pending name step → editable suggested address step (values only from the restoration cookie), provisioned state: no `?status=` ⇒ server-side identical re-mint and redirect into the fresh continuation; with a failure status ⇒ "Your organization is ready" + "Try again" (only when the record's slug equals the created slug) + manual tenant `/sign-in` link + abandon; dead credential ⇒ GET `/signup/abandon` |
| `web/app/signup/{start,create,abandon}/route.ts` (new, root) | start: signup-intent login, login binding set, old signup/form cookies cleared; create: `{csrf_token, name}` suggestion (record name + suggested slug), `{csrf_token, name, slug}` provisioning (record the exact submitted values on success/field refusal/slug_taken/rate limit/unknown outcome; not on intent_used, dead credential, CSRF/origin refusal), `{csrf_token, retry: "1"}` identical retry from the record (never rewritten; missing/stale ⇒ `retry_unavailable`, no core call); abandon POST revokes + clears both; abandon GET clears both only when core's status read is `unauthenticated` (never revokes) |
| `web/app/signup/{continue,complete}/route.ts` (new, tenant), `prove/route.ts` (new, root) | continue: challenge → `__Host-sct` (120 s) + 303 core's prove URL, else 303 `<root>/signup/details?status=continue_failed`; prove: original `__Host-signup` → 303 core's complete URL; no cookie ⇒ `/signup?status=expired`; 401 clears both; mismatch ⇒ `continue_failed`; complete: `__Host-sct` + `p` → clear `__Host-sct`, set `__Host-login`, 303 provider; failure clears `__Host-sct` ⇒ `continue_failed`; `rate_limited` ⇒ `?status=rate_limited`; every answer a no-store/no-referrer 303 |
| `web/components/signup/{landing,signup-form,details-form}.tsx` (new) | `Landing` (one Sign Up `<a data-testid="root-sign-up" href="${signup_origin}/signup">`), `SignupShell` + `PLATFORM_ENTRY` frame, `rootPage()`; `SignupForm`; `NameForm`, `DetailsForm`, `RetryForm`, `AbandonForm` — existing `AppShell`, `Section/Kicker/Heading/Text`, `FormField`, `PendingSubmit`, `StatusMessage`, `TextLink`, `Button`; no new tokens or design system |
| `web/tests/unit/host.test.mjs`, `session.test.mjs` | extended (above) |
| `web/tests/unit/signup-client.test.mjs`, `signup-forms.test.mjs` (new) | 17 and 19 tests (below) |

Commits: web code `99af0dd6f89ea1e2c801d0a5ad3d44c7be508988` (tree `eeeb6b14…`), tests
`d3a87914b15ea498707740670f2469ab7537c159` (tree `3a6c35b9…`). Every changed path outside the merged management files
matches the 46 patterns (`git diff --name-only origin/g062-impl2 HEAD -- . ':!management' | grep -vEf allow46` empty).

Decisions and deviations (for Codex; no core/API/schema change):
1. **Callback parser in `signup-client.ts`.** `enrollment-client.ts` (outside the 46 patterns) `completeCallback` refuses
   `tenant_created` and the 202 `signup` answer; the callback route now reads every completion through
   `signup-client.ts authCallback`, which re-validates G058's answers exactly as `completeCallback` does (same statuses
   via the exported `CALLBACK_STATUSES`) plus the two G062 answers. `completeCallback` stays (its unit tests unchanged) but
   has no route caller now. Smaller alternative if Codex amends scope: two lines in `completeCallback` and keep it.
2. **Form restoration cookie** (Codex option A): name `__Host-signup-form`, format above, Max-Age 1800 (the credential's
   30 minutes), recorded at the suggestion step too (name + core's suggested slug, shown for editing) and on field
   refusals/unknown outcomes; not rewritten by a retry or on `intent_used`; cleared on abandon, on any `unauthenticated`
   (create, prove, abandon, the abandon GET) and at a new `/signup/start`. A Server Component cannot set cookies, so the
   details page sends a dead credential to GET `/signup/abandon`, which clears only after re-reading `unauthenticated`.
3. **Automatic re-mint on `/signup/details`** (Plan §4.4 "otherwise it 303s straight into a fresh continuation"): done
   in the page render with the status read's CSRF token and the recorded payload, then Next's `redirect()` (a 307 for a
   GET render, not a route-handler 303 with explicit `no-referrer`; the tenant `continue` hop itself answers with
   no-store/no-referrer). Security note: a cross-site top-level GET of `/signup/details` can trigger this re-mint for a
   provisioned row (it only replaces the current generation and sends the browser into its own continuation; core's
   intent digest still decides). Codex may prefer the button-only variant (failure view for every provisioned visit).
4. **Root frame.** The generic `AppShell` header carries an Account link (a second, sign-in-like action on the root);
   `components/layout` is outside scope, so the root/www/signup pages use the entry frame with platform defaults (no
   label, logo or overrides; token scope label `tenant` with no overrides).
5. **Gate refinements.** The public entry also serves the shell's own static `/fonts/` and `/blueprint/` (else the landing
   loses its fonts); `/signup/continue|complete` pass only on platform hosts (core refuses custom hosts anyway).
6. Every `hosts/resolve` kind is now read with its exact member set (core sends exactly those; previously extra members
   were tolerated except `location`).

**Blocker for Codex (scope; not edited):** `web/tests/unit/enrollment-forms.test.mjs` (outside the 46 patterns), test
"a classified non-routing host (reserved, public entry) clears the join intent and reaches no callback" (lines 236-246),
feeds the retired answers `{kind: "public_entry"}` and `{kind: "public_entry", location: …}` and expects
`/sign-in?status=failed` with the join cookie cleared. Under the new contract both are malformed (unclassified ⇒
`/sign-in?status=unavailable`, no cookie touched), so this one subtest fails (observed, run 99). Minimal fix needs a scope
amendment admitting that file: replace the two answers with `{kind: "public_entry", signup_origin: "https://example.test"}`
(for host `org-c.localhost:543`, a non-root public entry) and expect the fixed 404 (or keep `/sign-in?status=failed` if
Codex prefers that over Plan §4.4's www 404). No other unit file regressed.

Commands (worktree root; plain native Docker; one container at a time; `node:24-slim`
`sha256:0e0ff40c39bc087845bfb27465a0df4ea419520094bc35842ff83dd8cbe6f9b6` (local, not pulled; the `web.Dockerfile` base;
Node v24.21.0), `--network none`, label `platform-test=g062-a1`, source by `docker cp` of `web/` into the container; no
published port):

| # | Command | Exit |
|---|---|---|
| 95 | `git fetch origin staging g062-impl2`; `origin/g062-impl2` = `7c24e0cc` asserted; `git switch -c g062-impl2-b origin/g062-impl2`; branch/HEAD asserted; allow-list extract `diff` | 0 / identical |
| 96 | `git merge --no-ff origin/staging -F <msg>` (author/committer <owner-email>); `git diff --name-only HEAD~1 HEAD` management only | 0 |
| 97 | pre-batch: `git fetch origin staging g062-impl2` (both unchanged); `bash infra/ci/docker-desktop-mounts.sh` (18%); `docker ps -a --filter name=platform-test-` empty | 0 |
| 98 | **diagnostic, not qualifying** (the `web-check` image and `node_modules` are not present locally, so typecheck/lint/`npm run test:unit` cannot run offline): container `platform-test-g062-w15-unit`, a stub `node_modules/server-only` (its `react-server` entry empty, as the real 0.0.1 package) inside the disposable copy only; `node --conditions=react-server --test tests/unit/host.test.mjs tests/unit/session.test.mjs tests/unit/signup-client.test.mjs tests/unit/signup-forms.test.mjs` on the tree later committed as `d3a87914` | 0: 86 tests, 86 pass (host 14, session 36, signup-client 17, signup-forms 19) |
| 99 | same container, the twelve other unit files | 1: `enrollment-forms.test.mjs` one subtest (blocker above); `onboarding-forms.test.mjs` cannot load (`typescript` package absent — environment, not this change); every other file passes |
| 100 | `git diff --check`; scope check of the changed names against the 46 patterns | 0 / empty |

Resource receipt (round 15): container `platform-test-g062-w15-unit` (one earlier create with a tmpfs layout removed
unused before start) removed with `docker rm -f -v`; `docker ps -a --filter label=platform-test=g062-a1` lists nothing.
No image pull or build, no database, Compose, browser stack, published port, `local-harness.lock` or `run.sh`; no
`npm install`. **Not run (missing offline):** `npm run lint`, `npm run typecheck`, the harness `unit-web` line — they need
the `web-check` image (or `node_modules`: `typescript` 5.9.3, `eslint` 9.39.5 + `eslint-config-next`, `server-only`,
`next`, `react`); TypeScript and lint were self-reviewed only.

**Proofs needed from the coordinator / round C (names and selectors; nothing edited here).**
- `unit-web` / `B-UNIT-WEB` and `lint-web` (README `web-check`: `npm run check` = lint + typecheck + `test:unit`) on the
  final tree; expect the `enrollment-forms.test.mjs` subtest above to fail until the scope amendment.
- `WEB_ROUTES` (run.sh) += 8: `/signup`, `/signup/details` pages and `/signup/{start,create,abandon,continue,prove,
  complete}` route handlers (`abandon` also answers GET); `static_web_boundary` must accept `web/lib/signup-client.ts`
  as a core client (it calls `/internal/v1/signup/*` and `/internal/v1/auth/callback`).
- `B-IDENTITY-PRIVACY`/`-PROBES`: add `ref`, `n`, `p` query values and the `__Host-sct`/`__Host-signup`/
  `__Host-signup-form` values to the redaction probe list (AC9); the form record must never appear in logs or URLs.
- `signup.spec.ts` (`signup-light/dark`, base `https://localhost:543`), selectors: landing `[data-testid=root-sign-up]`
  (href `https://localhost:543/signup` on both `localhost:543` and `www.localhost:543`; no Account/Sign in action);
  `/signup` `[data-testid=signup-start]`; details step 1 `[data-testid=signup-name-form]` (`#field-name`,
  `[data-testid=signup-name-submit]`), step 2 `[data-testid=signup-details-form]` (`#field-name`, `#field-slug`,
  `[data-testid=signup-create]`, `[data-testid=field-error]`), recovery `[data-testid=signup-retry]`,
  `[data-testid=signup-tenant-sign-in]`, `[data-testid=signup-abandon]`, notices `[data-testid=status-message]`; tenant
  entry `[data-testid=status-message]` "Your organization is ready." after `/?status=tenant_created`. Required browser
  proofs from the Codex Round 14 decision: `continue_failed` restore → "Try again" → identical re-mint → `tenant_created`
  with one org; `__Host-signup-form` attributes (Secure, HttpOnly, SameSite=Lax, Path=/, no Domain, Max-Age ≤ 1800, root
  host only, absent on the tenant host); a tampered record (edited cookie) → `intent_used` notice, no new org; a stale
  record after expiry → `/signup`; no name/slug in any URL (request log) or app log. Plus Plan SUP-15 `[E-SUP-1..6]`.
- `identity.spec.ts` `[E-DOM-4]` rewrite (root/www landing, no marketing hop) and the `playwright.config.ts` marketing
  host-resolver removal (round C).

**Remaining round C work:** overlay (`identity.yml` `--redirect-hosts`/`MARKETING_SITE_URL`/marketing host, `proxy.conf`,
`emulator-seed.yaml`), `.env.example` + run.sh re-pins (`ENV_DOMAIN_BLOCK_SHA`, `ENV_IDENTITY_DOMAIN_TAIL_SHA`, `core_env`,
`WEB_ROUTES` +8, B-MCP-SOURCE from Round 14, browser matrices, `identity:browser` projects), `playwright.config.ts`
(`signup-light/dark`), `signup.spec.ts`, `[E-DOM-4]`, docs (§9) and PG-009; then b0/b3, M1, final coverage, MR. The
`enrollment-forms.test.mjs` scope question and decisions 1 and 3 above should be settled first.

Not self-accepted; Codex review pending.

### Round 16: slice 2 round C1 (overlay, harness pins, browser specs)

**Rebinding.** Sixteenth writer (Opus 5.5 high; runtime effort not observable), single slice-2 implementation owner for
round C1, attempt 1 of B5 (no reset or extension); worktree `.claude/worktrees/agent-<id>`, branch
`g062-impl2-c1` created from `origin/g062-impl2` = `1fa3a30dd039a576218141cfad26610d44aac62b` after `git fetch origin
staging g062-impl2`; branch and HEAD asserted before any edit. Staging had advanced to
`e0cb63faf3b6ea746e173d7f6180a36d43f2c2eb` (management-only: Codex's Round 15 web disposition (not included),
sha256 `942a3ca980bb58be89aa74201c5426d1128fe8d86551ef5dacf23f21217def81`; scope amendment 04; PROGRESS). Per the
coordinator's instruction it was merged first as its own commit `6a48c1d779f9390b950d1e9d1c55efe16a55e58a` (`git merge
--no-ff origin/staging`; delta `management/v0/PROGRESS.md` sha256 `ed710e37…`, `G062-scope-allowlist.txt` sha256
`fd443933125ee0390225e1b88d6723cb2b738e50bd096f843f6afbc64508c4af`, the review record; no code). Scope authority from then
on: the staging allow-list's 47 non-`#` lines (amendment 04 adds `^web/tests/unit/enrollment-forms[.]test[.]mjs$`). Plan r5
`b2a612a2…` unchanged; this Result before round 16 `171a485c165369e1b09b4bed402b1b1b5178d83e98cda22381b5a5ad2e5cb49b`.

**Bounded security correction to Plan r5 §4.4 (Codex Round 15 decision 2): no re-mint on GET.** Recorded here, with the
edit, as the supersession record; the accepted Plan file is not edited.
- *Superseded text (Plan r5 §4.4):* "when the row is already provisioned the page shows the closed 'Your organization is
  ready' text with the 'Try again' POST (identical payload) only when it arrives with `?status=continue_failed` — otherwise
  it 303s straight into a fresh continuation (re-mint) with no button."
- *Old behaviour (Round B, `99af0dd6`):* `web/app/signup/details/page.tsx`, provisioned state without `?status=`, called
  `createOrganization` during the GET render with the status read's CSRF token and the recorded `{name, slug}`, then
  `redirect()`ed into the new continuation. A cross-site top-level navigation carries the `SameSite=Lax` root cookies, so a
  GET could trigger a state change (generation g+1) outside the POST route's Origin/CSRF check.
- *New behaviour:* every provisioned-state GET of `/signup/details` (with or without `?status=`) renders the recovery view
  without mutation: the closed "Your organization is ready" text, the user-triggered **Try again** POST (`RetryForm` →
  `POST /signup/create {csrf_token, retry: "1"}`: the original `__Host-signup` credential, the CSRF token and the exact
  recorded name/slug; core's intent digest stays authoritative) only when the `__Host-signup-form` record names the created
  slug, the manual tenant sign-in link and abandon. A missing, stale (empty slug) or mismatched record offers no Try again
  and triggers no core write; the page's only core call is the status read. The first successful create is unchanged: an
  automatic POST → 303 into the continuation. The decision is the pure `recoveryView(status, form, organization)` in
  `web/lib/signup-client.ts`; the page no longer imports `createOrganization`. No core/API/schema change; still exactly one
  Organization/owner/audit pair (core's re-mint semantics unchanged).
- *Evidence (stack-free, diagnostic run #103 below):* `signup-client.test.mjs` "recoveryView: a GET never calls core; Try
  again only for a record of the created address" (matching/mismatched/missing/empty-slug/no-organization cases; zero fetch
  calls); `signup-forms.test.mjs` "GET /signup/details never mutates" — (a) the page source imports exactly
  `DETAILS_NOTICES, fieldErrorOf, noticeOf, recoveryView, signupStatus` from the client, contains no
  `createOrganization`/`suggestSlug`/`abandonSignup`/hop call or `fetch(`, and redirects only to `/signup` and
  `/signup/abandon` (the page is TSX, not loadable by `node --test`, hence the source check); a GET of `/signup/create` is
  405 with no core call; (b) Try again is the identical POST (`Authorization: Signup <cred>`, body `{origin, csrf_token,
  name, slug}`) and continues; a foreign-Origin POST is 403 with no core call. The browser-level proof (direct GET → no
  generation change; Try again → `tenant_created` with one Organization) is written into `signup.spec.ts` `[E-SUP-5]`
  (below; written, not run).
- *Interpretation for Codex:* the retry POST route still forwards a record whose slug differs from the created slug to core
  (which refuses it as `intent_used`, nothing created) — the page never offers Try again for it; a pre-check would need a
  second core read in the POST route. Left unchanged unless Codex prefers that check.

**Scope amendment 04 applied** (`web/tests/unit/enrollment-forms.test.mjs`, only the stale fixture): the combined
"reserved, public entry" subtest is split. The reserved case keeps its refusal semantics unchanged
(`/sign-in?status=failed`, join cookie cleared, no callback). The new public-entry case feeds the exact G062 answer
`{kind: "public_entry", signup_origin: "https://example.test"}` (the test env's canonical root, `PLATFORM_BASE_DOMAIN=example.test`)
for the non-root origin `https://org-c.localhost:543` and expects the fixed 404 page (no Location), the cleared join cookie
(`Max-Age=0`) and no `/internal/v1/auth/callback` call. Nothing else in the file changed.

**Inventory (existing vs added).** Already present and reused: slice 1's `B-SIGNUP-CATALOG`/`B-SIGNUP-IT` rows and the
required-check count pins (92/123, amendment 03), `core/internal/signup/` in `profile-paths.tsv`, Round A's core API and
Round B's web surface and selectors. Missing before this round and added: every overlay/config change below, the
`.env.example` re-pin, `WEB_ROUTES` +8, the B-MCP-SOURCE signup pins, `core_env` without `MARKETING_SITE_URL`, the
B-WEB-BOUNDARY allow-list, the signup browser projects and matrix row, `web/app/signup/` as foundation, `signup.spec.ts`
and the `[E-DOM-4]` rewrite.

| Path | Change |
|---|---|
| `infra/compose/identity.yml` | `MARKETING_SITE_URL` dropped from core's env; `marketing.identity.test` dropped from `identity-setup --hosts`; emulator `--redirect-hosts` + `localhost,*.localhost` (deviation 1); header comments |
| `infra/identity/proxy.conf` | marketing stand-in server block and its comment removed; G062 note (public entry served by web; a new `<slug>.localhost` matches the existing pattern) |
| `infra/identity/emulator-seed.yaml` | comment only (no seed user added: fresh creators via the emulator API, existing identity = seeded alice) |
| `.env.example` | domain block: the marketing sentence and the commented `MARKETING_SITE_URL` line removed (block 17 → 16 lines) |
| `infra/validation/run.sh` | env pins; `WEB_ROUTES`; B-MCP-SOURCE pins/checks/PASS text; `core_env`; B-WEB-BOUNDARY allow-list, comment, log and PASS texts; B-IDENTITY-E2E and B-IDENTITY-PRIVACY PASS texts |
| `infra/validation/lib.sh` | `BROWSER_MATRIX_IDENTITY_E2E` + `"signup-light signup-dark|E-SUP-1 … E-SUP-6 S-ERRORS"` (Plan §11 pin) |
| `infra/ci/profile-paths.tsv` | `foundation dir web/app/signup/` (Plan §2) |
| `web/components/signup/landing.tsx`, `web/app/page.tsx`, `web/app/signup/page.tsx`, `web/app/signup/details/page.tsx` | B-WEB-BOUNDARY correction (deviation 2): the component no longer imports `@/lib/config`; `SignupShell` takes `productName`/`platformBaseDomain` from the three pages |
| `web/playwright.config.ts` | marketing host-resolver rule removed; `signup-light`/`signup-dark` (base `https://localhost:543`) |
| `web/tests/browser/identity.spec.ts` | `[E-DOM-4]` rewritten (root and www landing: one Sign Up to `https://localhost:543/signup`, no redirect hop, `/account`, `/sign-in`, `/connections/mcp`, `/groups` the fixed 404, `/api/health` 200, root `/signup` 200, www `/signup` 404); `MARKETING` constant and comment removed |
| `web/tests/browser/signup.spec.ts` (new) | SUP-15 `[E-SUP-1..6]` (below) |
| plus the correction commit above | `web/app/signup/details/page.tsx`, `web/lib/signup-client.ts`, three unit files |

Commits: `5570717e` (no re-mint on GET + amendment 04 + this section's first part), `46c047a6` (overlay, config, harness
pins, boundary correction), `46ddc678` (browser specs), then this Result commit. Every changed path since `1fa3a30d`
outside the two merged staging management files matches the 47 patterns.

**Pins: before → after, and how each was derived.**

| Pin | Before | After | Derivation |
|---|---|---|---|
| `ENV_DOMAIN_BLOCK_SHA` | `cae3a79015168389f74c7c34ec2d93f454d829fb454c0c91cd8e8f2355d92006` | `7d861980ea86918e4569ea87cc5578952af075494cfda6ddfd1c1d0e5b28d48d` | `awk -v d="# --- Hostnames and custom domains (docs/foundation/DOMAINS.md) -------------" '$0 == d {f = 1} f' .env.example \| sha256sum` (B-ENV-EXAMPLE's own range); the old value reproduced exactly on `git show HEAD:.env.example` first |
| `ENV_IDENTITY_DOMAIN_TAIL_SHA` | `aad6dbc7fa074415154a9299e89c1fd7a9785ecdf91ff4ed253d0013236065e7` | `7bc688ad72854b9b439d1787c3e13e9c1f0644ebb86f19f77c0626b5af66e0f5` | `awk -v m="# --- Identity provider: WorkOS AuthKit (docs/foundation/IDENTITY.md) --------" '$0 == m {f = 1} f' .env.example \| sha256sum` (variants_off's range; old reproduced first); tail 31 → 30 lines (comment updated) |
| `ENV_IDENTITY_BLOCK_SHA` | `3d55f30d…` | unchanged | between-markers awk range re-hashed: still `3d55f30dedcae4654c9c098557acdfc5f575e8e22249c0f96e1c51793bf300c2` |
| `WEB_ROUTES` | 50 entries | 58 entries (+`signup/abandon/route.ts signup/complete/route.ts signup/continue/route.ts signup/create/route.ts signup/details/page.tsx signup/page.tsx signup/prove/route.ts signup/start/route.ts`) | replay of `web_routes_ok`'s exact `find -printf '%P\0'` / non-`(features)` filter / `LC_ALL=C sort` / space join on the final tree (script `~/.local/state/platform-evidence/g062-a1/gen/w16-routes.sh`); GET `/signup/abandon` is the same file |
| `MAIN_GO_G016_SHA` | `b4fed81910017929b7b822430ecc179bda367449442947f6d5e05039d654723d` | `db0e4c151015b8ce3fd3015a5ab968484ab1edc19ef681b240f98e5cbfb6797c` | `sha256sum core/cmd/server/main.go` (equals Round 14's recorded digest; main.go unchanged since) |
| `MAIN_GO_IMPORTS` | … `core/internal/onboarding/api errors` … | … `core/internal/onboarding/api core/internal/signup core/internal/signup/api errors` … | `go list -f '{{join .Imports " "}}' ./cmd/server` in `golang:1.27.3-alpine` (local), `--network none`, read-only `g058-a1-gomod`, `GOFLAGS=-mod=readonly GOPROXY=off GOTOOLCHAIN=local` — output pasted verbatim |
| `CORE_SIGNUP_ROUTES` (new) | — | `GET /internal/v1/signup/status\|POST …/abandon\|POST …/challenge\|POST …/complete\|POST …/organizations\|POST …/prove\|POST …/slug` | the check's own `sed -nE 's/^[[:space:]]+Route[A-Za-z]+[[:space:]]+= "([^"]+)"$/\1/p' core/internal/signup/api/http.go \| LC_ALL=C sort \| paste -sd'\|'`; new check line beside the enrollment one |
| signup/api importer check (new) | — | `core/cmd/server` only | the check's own `importers` function replayed: `core/cmd/server`; the access, tenancy, objectstore and mcp importer sets replayed unchanged; literal and constant `mux.Handle*` route sets unchanged (Register is not a `mux.Handle` call) |
| B-WEB-BOUNDARY request-context allow-list | G004…G058 files | + `app/signup/(start\|create\|abandon\|continue\|prove\|complete)/route.ts`, `components/signup/landing.tsx`, `lib/signup-client.ts` | exactly the eight files of the coordinator's failing batch (`harness.log` lines 8-15, run 01:27); replay of the grep with the new regex: none outside; the component-configuration rule replayed: none (after the landing correction) |
| `core_env` | `MARKETING_SITE_URL=https://www.probe-marketing.example.net/` | line removed | Plan §2; core ignores the variable (Round 14 `TestMarketingSiteURLRemoved`); production probes (B-IDENTITY-CONFIG cc1-cc8, mcp_config_refusals) start without it |
| browser matrix | 5 rows | + `signup-light signup-dark\|E-SUP-1 E-SUP-2 E-SUP-3 E-SUP-4 E-SUP-5 E-SUP-6 S-ERRORS`; `domain-platform\|E-DOM-1 E-DOM-3 E-DOM-4 S-ERRORS` unchanged | Plan §11 |
| `required-checks.tsv`, `validate-controls.sh` count pins | 92 / 123 | unchanged | no check ID is added or removed in this round; the validate-controls route control extracts `WEB_ROUTES` from run.sh at run time (no pin) |
| `profile-paths.tsv` | — | `foundation dir web/app/signup/` | Plan §2; no row-count or hash pin exists |

**Redaction (AC9).** B-IDENTITY-PRIVACY/PROBES use one registry: the run's secrets plus the values the browser specs
register with `observe()` (`web/tests/browser/support/sanitize.ts`), scanned in core, replica, web and proxy logs and in
the browser evidence. `sanitize.ts` (its `SECRET_PARAMS`) is outside the 47 patterns and is not edited; `signup.spec.ts`
registers every `ref`, `n` and `p` it sees (each hop's request URL and Location, every held Location), the
`__Host-signup`, `__Host-sct`, `__Host-signup-form` and `__Host-login` values it reads, every rendered CSRF token, the
emulator state values, the fresh users' passwords and an organization-name canary per signup; it saves no URL or page
text as an attachment. The harness side needed no list change beyond the PASS text.

**`signup.spec.ts` (written, not run; selectors from Round B).** `[E-SUP-1]` root and www landing, one `root-sign-up` to
`https://localhost:543/signup`, no redirect, other pages 404, www `/signup` 404, `/signup/details` without cookie → `/signup`.
`[E-SUP-2]` fresh emulator user → `signup-start` → details (name, suggested slug edited) → `signup-create` → automatic
`continue`/`prove`/`complete` (each a 303 with `no-store` and `no-referrer`, recorded per hop) → `<slug>.localhost:543`
`/?status=tenant_created`; `__Host-sid` host-only on the tenant only; `__Host-signup` and `__Host-signup-form` host-only on
the root (Secure, HttpOnly, SameSite=Lax, Path=/, no Domain, ≤ 1800 s), the record equal to `v1.<base64url name>.<slug>`,
absent on the tenant; no name/slug in any request path or query. `[E-SUP-3]` seeded alice (org-a member) creates a second
tenant: same `/api/account` user ID, org-a indicator still `member`, distinct Organization. `[E-SUP-4]` keyboard-only
start, axe 1280/320 on landing, start, both details steps and each field error; `org-a` → slug_taken, `www`, `auth`,
`Bad_Slug` → invalid_slug; abandon → `/signup?status=abandoned`, both cookies cleared, no tenant (`hosts/resolve` 404).
`[E-SUP-5]` four tests: (1) **no re-mint on GET** — the create's 303 is held (R unused), a direct GET and a cross-site
top-level navigation from `attacker.localhost` of `/signup/details` render the recovery view with Try again and issue no
`/signup/create` and no tenant request, then the held R still completes (generation unchanged) into the same Organization;
(2) captured R in a second browser (R spent; its `__Host-sct` host-only ≤ 120 s; prove without `__Host-signup` → `/signup`,
no session), captured N proved by the original browser lands on `continue_failed` (no challenge cookie), the original R →
`continue_failed`; a tampered record (other name, same slug) → Try again → `intent_used`, record not rewritten, same
Organization; a stale record (other slug) and no record → no Try again, no create; the exact record → Try again →
`tenant_created`, same Organization; R/N replays → recovery page; the captured generation-1 P in the second browser →
refused, no session; (3) dropped create response → recovery view → Try again; every hop held and checked (challenge
cookie set then cleared, `__Host-login` host-only ≤ 600 s, `redirect_uri` = the tenant's `/auth/callback`); another
browser with only a copy of the tenant login binding signs in as bob → tenant `/sign-in?status=failed`, no session
anywhere; the creator recovers with Try again into the same Organization; (4) provider session not live (emulator cookies
cleared): the hosted prompt appears, the same user → `tenant_created`; a second creator cancels at the prompt → tenant
`/sign-in?status=canceled`, the tenant routes, a later ordinary tenant sign-in succeeds with membership. `[E-SUP-6]` the
new owner opens `/groups`, `/account`, `/groups/new` directly (200, no onboarding redirect).

**Deviations and questions for Codex.**
1. **`--redirect-hosts`.** The Plan says "+ `localhost`". In the pinned emulator 0.14.0 (`dist/workos/helpers.js`,
   inspected by copying `/app/dist` out of a never-started container) `localhost` is always allowed by default
   (`DEFAULT_ALLOWED_REDIRECT_HOSTS`) and host entries match exactly, while `*.example` matches subdomains. A tenant
   created during the run (`<slug>.localhost`) would therefore be refused at the final hop. Added `localhost` (per Plan,
   documentary) and `*.localhost`. Fixture-only: core alone builds redirect URIs from the resolved host; no existing test
   depends on the emulator refusing a host. Smaller alternative if refused: fixed slugs in `signup.spec.ts` listed
   explicitly (M1 would then need one of those slugs).
2. **B-WEB-BOUNDARY component rule.** The batch also failed "client or component code reaches server-only configuration"
   (`components/signup/landing.tsx:8` `@/lib/config`). Fixed in the source (props from the pages), not by widening that
   rule; the request-context list is extended for the eight files only, as G017 did for its tenant frame.
3. **`identity:browser` is a docs block** (`docs/foundation/IDENTITY.md`, round C2). Until C2 adds
   `--project=signup-light --project=signup-dark` to that line, B-IDENTITY-E2E fails the new matrix row by design.
4. **Retry with a mismatched record** (from the correction above): the page never offers it; the POST route still forwards
   it to core (`intent_used`, nothing created). Unchanged unless Codex wants a pre-check.
5. **Assumptions to verify in b3:** route.fetch's Set-Cookie reaches the browser context's store (asserted where relied on);
   Chromium's SPKI pin accepts a new `<slug>.localhost` whose name is not in the run leaf's SANs (identitysetup `--hosts`
   is a fixed list; the spec never calls a tenant origin from the Node client).

**Commands** (worktree root; plain native Docker, one container at a time, names `platform-test-g062-w16-*`, label
`platform-test=g062-a1`, `--network none`, sources by `docker cp`; no published port, database, Compose, browser stack,
image build or pull, `run.sh`, `local-harness.lock` or `npm install`):

| # | Command | Exit |
|---|---|---|
| 101 | `git fetch origin staging g062-impl2`; `origin/g062-impl2` = `1fa3a30d` asserted; `git switch -c g062-impl2-c1 origin/g062-impl2`; branch/HEAD asserted | 0 |
| 102 | `git fetch origin staging` (`e0cb63fa`); `git merge --no-ff origin/staging -F <msg>` → `6a48c1d7` (management only); 47-pattern extract regenerated from the staging scope file | 0 |
| 103 | pre-batch: `bash infra/ci/docker-desktop-mounts.sh` (18%); `docker ps -a --filter name=platform-test-` empty | 0 |
| 104 | **diagnostic** unit run (as round 15: `node:24-slim` `sha256:0e0ff40c…`, a stub `server-only` in the disposable copy; the `web-check` image and `node_modules` are absent): `node --conditions=react-server --test` on `host`, `session`, `signup-client`, `signup-forms`, `enrollment-forms` tests | 0: 102 tests, 102 pass |
| 105 | `git push origin g062-impl2-c1:g062-impl2` (`1fa3a30d..5570717e`, fast-forward) | 0 |
| 106 | `docker create` (never started) of the pinned emulator image + `docker cp /app/dist` out; `grep` of the redirect-host logic | 0 |
| 107 | `.env.example` pin derivations (old file from `git show HEAD:.env.example`, new file): awk ranges above `\| sha256sum`, `wc -l` | 0 |
| 108 | `bash w16-routes.sh <worktree>/web` (route inventory replay) | 0: 58 routes |
| 109 | `go list -f '{{join .Imports " "}}' ./cmd/server` (`golang:1.27.3-alpine`, offline module cache) | 0 |
| 110 | `bash w16-boundary-check.sh <worktree>` (request-context and component-configuration replays) | 0: none, none |
| 111 | `bash w16-mcpsource-check.sh <worktree>` (digest, signup routes, importer and route sets) | 0: all match |
| 112 | `bash -n infra/validation/run.sh`; `bash -n infra/validation/lib.sh` | 0 |
| 113 | `koalaman/shellcheck:v0.11.0 -S warning -x -P SCRIPTDIR infra/validation/run.sh infra/validation/lib.sh` (B-SHELL's flags) | 0 |
| 114 | `mikefarah/yq:4.48.1 e` on `identity.yml` (parses; emulator command, core env and identity-setup hosts as intended) | 0 |
| 115 | push `5570717e..46c047a6` | 0 |
| 116 | syntax-only check (Node's `stripTypeScriptTypes` then `node --check`) of `signup.spec.ts`, `identity.spec.ts`, `playwright.config.ts` — not a type check | 0: 3 ok |
| 117 | unit rerun on the final web tree (`w16-unit-run2.log`, sha256 `a1cf8637405fb99f9cbfd32bfc7a14d16f97ead81166776758ee7400021bbb2e`) | 0: 102 tests, 102 pass |
| 118 | `git diff --check`; `git diff --name-only 1fa3a30d HEAD` minus the two merged staging files `\| grep -vEf <47 patterns>` | 0 / empty |
| 119 | push `46c047a6..46ddc678`; `git fetch origin staging` before the final push: still `e0cb63fa` (scope file unchanged, 47 patterns) | 0 |

Resource receipt (round 16): containers `platform-test-g062-w16-unit` (three create/remove cycles while placing the copy;
one run), `-w16-emu` (never started), `-w16-golist`, `-w16-sc`, `-w16-yq`, `-w16-syntax` (two), `-w16-unit2`, each removed
with `docker rm -f -v`; `docker ps -a --filter label=platform-test=g062-a1` lists nothing. Scratch scripts and the copied
emulator `dist` are under `~/.local/state/platform-evidence/g062-a1/gen/` (`w16-*`).

**Not run (cannot run stack-free or offline here):** `npm run lint`/`typecheck` and `unit-web` (no `web-check` image or
`node_modules`); TypeScript type checking of the specs and the edited TSX; every `run.sh` suite; Playwright.

**Coordinator's earlier lock-held batch (recorded for Codex, coordinator-reported; evidence read-only here).** On
`g062-impl2` `1fa3a30d`, 01:27-01:29: `flock -n <home>/.local/state/platform-evidence/local-harness.lock bash
infra/validation/run.sh --base e5e9c7204335630b476d82573f02fe3ff11c0704 --allow-file <allow46> --report-dir
<home>/.local/state/platform-evidence/g062-a1/s2b-batch1-1fa3a30d/evidence lint-web unit-web`: B-WEB-TOKENS,
B-WEB-COMPONENTS, B-WEB-DESIGN, B-WEB-ASSET-CONTROLS PASS; B-SCOPE PASS (35 paths); B-SECRET-CANARY PASS; B-WEB-BOUNDARY
FAIL (`harness.log`: the landing's `@/lib/config` import, the eight request-context importers and the route inventory —
all three addressed above); B-UNIT-WEB FAIL (exactly the `enrollment-forms.test.mjs` stale fixture, fixed under amendment
04); RESULT FAILED as expected pending these fixes.

**Lock-held proofs the coordinator must run next (names only).**
- Rerun `lint-web unit-web` with the 47-pattern allow-list on the pushed tip: B-WEB-BOUNDARY (routes, request-context,
  component configuration), B-UNIT-WEB (incl. the new no-re-mint and amendment-04 tests), lint/typecheck.
- `lint`: B-ENV-EXAMPLE (new domain pin), B-SHELL, B-COMPOSE; `lint-base`: B-LITERALS, B-IDENTITY-GUARDS (with `lint`
  for B-SECRET-CANARY); every run: B-SCOPE with the 47-pattern allow-list; `lint-core`: B-MCP-SOURCE (digest, imports, signup routes, importer); `ci-controls` (the WEB_ROUTES control).
- `off`: B-TOGGLE-OFF (the new tail pin). `root`: B-CORE-ERROR-OUTPUT (`core_env`).
- After round C2 adds the signup projects to IDENTITY.md's `identity:browser` line: `ops` identity — B-IDENTITY-E2E
  (b3: `signup-light`, `signup-dark`, `domain-platform` incl. the new `[E-DOM-4]`, then the unchanged projects),
  B-IDENTITY-PROBES, B-IDENTITY-PRIVACY (the new registered values), B-IDENTITY-CONFIG and the MCP config refusals
  (`core_env`), B-MCP-E2E.
- Then M1 (Plan §5.5: owner walk-through on the overlay, root and www → Sign Up → emulator → details → new tenant, the
  existing-identity second tenant, slug taken, the not-live prompt; light/dark/keyboard), then final local coverage and MR.

**Remaining:** round C2 (docs: `docs/ARCHITECTURE.md`, `SECURITY.md`, `PERFORMANCE.md` PG-009, `docs/foundation/*`,
including IDENTITY.md's `identity:browser` projects and the supersession notes), the lock-held proofs above, M1, final
coverage, MR, merged-SHA full local qualification. Round B's source review (callback-parser duplication and other
deviations) remains Codex's.

Not self-accepted; Codex review pending.

### Round 17: slice 2 round C2 (documentation)

**Rebinding.** Seventeenth writer (Opus 5.5 high; runtime effort not observable), single slice-2 implementation owner for
round C2 (adopter documentation), attempt 1 of B5 (no reset or extension); worktree
`.claude/worktrees/agent-<id>`, branch `g062-impl2-c2` created from `origin/g062-impl2` =
`b707718ae9b9c03abf234a2405362b091c423341` after `git fetch origin staging g062-impl2`; branch and HEAD asserted before
any edit. `origin/staging` = `e0cb63faf3b6ea746e173d7f6180a36d43f2c2eb` (already merged into the base by round 16;
unchanged at the pre-commit re-fetch, so no reconcile commit). Verified: slice-2 release record `f88acb4e…`, Plan r5
`b2a612a2…` (not edited), staging scope file `fd443933…`; its 47 non-`#` lines are byte-identical to the coordinator's
`g062-allow47.txt` (that extract file itself hashes `0b8abb4c…`; the `fd443933…` digest is the full scope file with its
comments). This Result before round 17: `ba2837b73c901c7af4dd10fab9399be9afb0c61fde810405c7716370f9a4f812`. Decisions
documented as binding: Codex Round 14 option A (`__Host-signup-form`), Codex Round 15 decision 2 (no re-mint on GET;
supersedes Plan r5 §4.4 as recorded in round 16).

**Files (documentation only; no source, harness or test file).**

| Path | What it now says |
|---|---|
| `docs/ARCHITECTURE.md` (protected) | §51 `app.foundation.test` row only: "Planned G062 foundation landing" → the implemented landing with Sign Up only, `www` linking to the root, automatic continuation into the new tenant host, `MARKETING_SITE_URL` removed, live provider/edge qualification remaining G057's; links to DOMAINS.md and IDENTITY.md. No other line changed. |
| `docs/SECURITY.md` (protected) | Section 3 runtime-material table: three rows — `signup_sessions.credential_hash` (signup credential, SHA-256 only, ≤ 30 min), `signup_sessions.csrf_token` (recoverable 32 bytes), `ref_hash`/`challenge_hash`/`nonce_hash`/`proof_hash` (SHA-256 only, one generation of 120 s, all four invalidated by a re-mint). The dated "Planned platform-root provisioning boundary — G062" section is kept verbatim and gains one implementation note (G060 precedent): separate Signup bearer refused by session consumers, provision-writer-only creation, one-use hashed continuation secrets, no re-mint on GET, the restoration cookie is never authentication, no names/slugs/secrets in logs, revocation and restore expiry. |
| `docs/PERFORMANCE.md` (protected) | New subsection "PG-009 G062 platform root signup and tenant continuation — source estimate, not measured" under PG-009 (Plan D-10 names PG-009; PG-009 is the G018 record and G019 already appended a subsection there, so G062 follows that precedent rather than opening PG-010): sign-in sweep (+1 DELETE), callback (+1 INSERT, no session), credential resolution (one unique-index autocommit probe), creation (two row locks, two keyed INSERTs with their two trigger audit rows, one UPDATE; retry = locks + one UPDATE), hops (host probe + one unique-digest UPDATE; completion + one short LOGIN transaction), +1 UPDATE in each revocation path and in the restore invalidation, in-memory limiters; nothing measured, no index/cache change. |
| `docs/foundation/IDENTITY.md` | New section "Platform root signup (G062)": root origin (configuration, `signup_origin`, not an `IDENTITY_PUBLIC_ORIGINS` entry, apex provider entries); signup login intent (exact root origin only); T1 callback (`202 {signup: {credential, next}}`, no session, `__Host-signup`); credential boundary and revocation scope; the seven core routes and their answers; creation and intent digest; the four automatic 303 hops with `no-store`/`no-referrer`; pinned callback (`signup_user_mismatch`, issuance guard, `tenant_created`, not-live provider prompt); **read-only GET recovery** (Round 15 correction, supersedes Plan §4.4); cookies `__Host-signup` (1800 s), `__Host-sct` (120 s), `__Host-signup-form` (format, 1800 s, write/clear rules, never authentication); every rate-limit bucket with the per-process statement; privacy (no name/slug/ref/n/p in URLs other than each hop's own Location, or in logs; `signup.request`; no web `http.request.completed`); F/Q evidence boundary. Also: Logs sentence for `signup.request`; emulator origins (marketing stand-in removed); browser-projects bullet (`signup-light`/`signup-dark`); `identity:browser` block (below). |
| `docs/foundation/DOMAINS.md` | Host-model row for `B`/`www.B` (landing; root signup surface and `/auth/callback` on `B` only; everything else 404; probes exempt) plus a "Public entry (G062; supersedes G016's 303 to `MARKETING_SITE_URL`)" paragraph (`{kind, signup_origin}` with no `location`; exact path sets for root, `www`, platform and custom tenant hosts); configuration row `MARKETING_SITE_URL` struck through as removed and ignored; "Creation-time slugs (G062)" (platform-wide unique, no per-tenant DNS/CNAME); root signup sign-in note; edge inputs: query redaction extended to the three hop paths and the exact apex provider entries `https://B/auth/callback`, `https://B/sign-in` (G057 operator action); test overlay (`localhost` and `*.localhost` redirect hosts, marketing stand-in gone); evidence rows (E-DOM-4 rewrite, `signup-*`). |
| `docs/foundation/AUTHORIZATION.md` | New subsection "Platform root signup (G062; outside the facade)": no TenantContext, no capability, no `tenancy`/`access` import (D-9); authority = own signup credential + CSRF, the database decides through the provision writer; ordinary policy from the first tenant session; `access.denied` with `signup.*` operations. |
| `docs/foundation/DATABASE-ISOLATION.md` | Roles table row `<db>_provision_writer` (exact grants and non-grants); readers table rows `auth_resolve_signup` (auth reader) and the three provision-writer bodies (callers runtime and owner, per the 00018 GRANT); Provisioning: four helpers in both paths, a pre-G062 volume gains only the fourth, 00018 refuses until it exists, "both helpers" → "each helper it created"; new "Version 18 (G062)" record (table, 79 relations, runtime/owner, auth-reader and provision-writer policies incl. the proof WITH CHECK, `login_transactions` columns/CHECKs/return paths, unchanged `login_transactions_insert`, B-SIGNUP-CATALOG/PROV-1..6/B-DBRENAME). |
| `docs/foundation/AUDIT.md` | G1 row gains the four signup `auth.sign_in` reasons; G2 row gains the five signup refusals plus `host_changed` with `signup.*` and "73 operations"; the operations list gains the eight `signup.*` operations (65 → 73); new "Platform root signup (G062)" paragraph (no action or resource type; reasons 70 → 79; `events_insert_provision`; the creation's `organization.created`/`membership.granted` captures; down refusal). The catalog table rows are unchanged (46 actions, 39 trigger-owned; B-AUDIT-CATALOG's extraction re-run below). |
| `docs/foundation/TENANCY.md` | New "Platform root signup (G062)" section: the creator-owner binding as a third source beside the G019 Owners bullet (role/source CHECK unchanged; supersession note, the G019 text is not rewritten); defaults (`invite_only`, override false); intent idempotency and slug race; no runtime Organization/`org_owner` INSERT path. |
| `docs/foundation/WEB-SHELL.md` | Browser-projects table row `signup-light`/`signup-dark`; Logged-out-entry note (root and `www` are not tenant entries); new "Platform root landing and signup pages (G062)" route table (landing, start, details incl. the read-only recovery view, the root routes, the three hops, the tenant `tenant_created` notice; entry frame, existing components, `WEB_ROUTES` +8); Status line G062. |
| `docs/foundation/MIGRATIONS.md` | Current-set row 18 (`00018_platform_signup.sql`; history now ends at 18; the table's pre-existing gap at row 16 is left as it was); the "From version 18 (G062)" sentence; `down` from version 18 bullet (refusals, restoration, role kept). The `migrations:*` command blocks are unchanged. |
| `docs/foundation/RECOVERY.md` | Data-class row: `signup_sessions` in the identity class, 30-min/120-s lifetimes, restore expires signup credentials; `invalidate-auth` paragraph: the fold into the invoked `invalidate.sql` (`recovery.sh` unchanged), `signup_sessions_expired` receipt key, postcondition, restored secrets refused, signup login transactions and pinned handoffs deleted with the others. |

Commits on `g062-impl2-c2`: `2e6cd66d80da94d139cbd8dbaccb9ed6cfbd10ee` (the three protected docs),
`492ee1026f2bd65d2df37cdf16426efbfb6b8c51` (the nine foundation docs and the `identity:browser` block), then this Result
commit. `git diff --name-only b707718a HEAD` matches the 47 patterns (`grep -vEf g062-allow47.txt` empty).

**`identity:browser` block and harness pins.** The block is now (one line; still exactly one `docker compose … run --rm …`
line, as `identity_e2e` requires): `… npx playwright test --project=identity-light --project=identity-dark
--project=domain-platform --project=domain-custom --project=mcp-light --project=mcp-dark --project=onboarding-light
--project=onboarding-dark --project=membership-light --project=membership-dark --project=enrollment-light
--project=enrollment-dark --project=signup-light --project=signup-dark`. `domain-platform` was already present (it
carries the rewritten E-DOM-4); the change appends `--project=signup-light --project=signup-dark`, which the round-16
matrix row `signup-light signup-dark|E-SUP-1 … E-SUP-6 S-ERRORS` (lib.sh `BROWSER_MATRIX_IDENTITY_E2E`) and the
round-16 PASS text already expect. `web/playwright.config.ts` defines both projects (base `https://localhost:543`) in
`identityProjects()`.

| Item | Before | After | Derivation |
|---|---|---|---|
| `identity:browser` block text (sha256 of the extracted block; **not pinned anywhere**) | `cb1ac3797369f4bffd51d10efbdd887047b1e6de19bfbb71e315fab4064dfa63` | `8d95d92d88f3f70fbd4711d952b370c2161c3ab0c4b375cf4d6a2cb8c3f0471e` | `awk -v m="<!-- identity:browser -->" '$0==m{f=1; next} f && /^```sh/{c=1; next} c && /^```/{exit} c{print}' <file> \| sha256sum` (lib.sh `docs_block`'s own awk) on `git show b707718a:docs/foundation/IDENTITY.md` and on the new file |
| run.sh/lib.sh/validate-controls.sh/required-checks.tsv/profile-paths.tsv pins over the touched docs | none | none (unchanged) | `grep -rn` for each touched path, `identity:browser` and the three protected docs over those five files: only the `APPLICABLE` input lists (run.sh 211–261), the `.env.example` marker strings (705/709, `.env.example` content, not doc content), `docs_block`/`run_docs_block` consumers of IDENTITY.md `identity:*` and MIGRATIONS.md `migrations:*` (blocks other than `identity:browser` unchanged), the WEB-SHELL.md init-hash lookup (not touched), the AUDIT.md catalog extraction (below) and validate-controls' path-only profile fixtures `F-gov-architecture`/`F-gov-security`/`F-doc-mixed-runtime`. No sha256 over doc content exists for these files. |
| B-AUDIT-CATALOG doc extraction (`sed -nE 's/^\| [TG][0-9]+ \| `([a-z_.]+)` \|.*/\1/p'`) | 46 actions, 39 trigger-owned | 46, 39 (sorted list sha256 `2fd1efd7…` before and after) | run.sh 3350–3353 replayed on both versions |
| B-LITERALS | — | no manifest literal added | `git diff -U0 -- docs/foundation \| grep -ic beginify` = 0 (the manifest literals are the product name, slug, base domain, CNAME target and database name; none appears in the foundation diff) |

**Expect-hash list for B-PROTECTED** (docs/ except docs/foundation is protected; the final tree's SHA-256 of each
changed protected file, for `--expect-hash PATH=SHA256`):

- `docs/ARCHITECTURE.md=03c7b8313e62ea6c7f6c5b9bf63a3c8676259aae98353b1399de80f5807ab06e`
- `docs/SECURITY.md=bcb19e1045929aeffa3bece3dc301d21881ebece59a81d2e9120b85e04ce8f69`
- `docs/PERFORMANCE.md=46f498a26262d68bdf089e9999028db1c40b881a8ae4ea9b1baa84c1859fe617`
- plus this Result (`management/` is protected too) at its committed hash, as for every earlier Result round.

Foundation docs at the final tree (not protected; for binding): IDENTITY `dfd46bd1…`, DOMAINS `1b8a51b2…`, AUTHORIZATION
`3366e046…`, DATABASE-ISOLATION `12b63073…`, AUDIT `7d3e00eb…`, TENANCY `06a6b0c2…`, WEB-SHELL `29081824…`, MIGRATIONS
`4f4062eb…`, RECOVERY `4c6daee1…`.

**Commands** (worktree root; no container, image, network, volume, database, Compose, browser stack, `run.sh` or
`local-harness.lock`):

| # | Command | Exit |
|---|---|---|
| 120 | `git fetch origin staging g062-impl2`; `origin/g062-impl2` = `b707718a` asserted; `git checkout -b g062-impl2-c2 b707718a`; branch/HEAD asserted; staging scope file `sha256sum` (`fd443933…`) and `grep -v '^#'` extract `diff` against `g062-allow47.txt` (identical, 47 lines) | 0 |
| 121 | `git merge-base origin/staging origin/g062-impl2` = `e0cb63fa` (staging already merged; no reconcile) | 0 |
| 122 | read-only inspection of the code being documented (00018, `signup/`, `signup/api`, identity store/service/http, `web/lib/{session,signup-client,host}.ts`, `web/proxy.ts`, signup routes, `identity.yml`, `invalidate.sql`, `events.json`, `audit.go` diff from `47df73f5`) | 0 |
| 123 | doc-pin inventory greps (table above); `docs_block` awk on the old and new IDENTITY.md (one line, contains ` run --rm `) and `sha256sum` | 0 |
| 124 | B-AUDIT-CATALOG extraction replay on old/new AUDIT.md (46/39, equal sorted digests) | 0 |
| 125 | `git diff --check` | 0 |
| 126 | `git diff --name-only HEAD \| grep -vEf g062-allow47.txt` | 1 (no output: every path matches) |
| 127 | `git fetch origin staging g062-impl2` before committing: both unchanged (`e0cb63fa`, `b707718a`) | 0 |
| 128 | two commits (protected docs; foundation docs); `sha256sum` of the 12 docs | 0 |

**Not run (and why).** No suite: this round changed documentation only. Rendering of the Markdown was checked by reading
the changed sections, not by a renderer. One read-only inspection command chained two `sed`/`sha256sum` pipelines with
`;` (the AUDIT.md replay); no state changed. Resource receipt (round 17): nothing created; no container or label
`platform-test=g062-a1` used.

**Deviations and questions for Codex.**
1. **PG-009 placement.** Plan r5 D-10/§7 names "PG-009", which is the G018 record; G062 is added as a subsection under it
   (the G019 precedent) instead of a new PG-010. Renaming would be a one-heading change if Codex prefers PG-010.
2. **Apex `https://B/sign-in` provider entry.** Documented as Plan §5.5 lists it, although the root's `/sign-in` is the
   gate's fixed 404 (the signup credential has no sign-out return). Harmless to register; Codex may drop it.
3. **SECURITY.md G062 heading** keeps its dated "Planned …" title (history not rewritten); the implemented status is the
   appended note, exactly as G060's section did.
4. `docs/foundation/OPERATIONS.md`, `VENDORS.md` and `docs/adr/**` are outside the 47 patterns and were not edited; no
   `MARKETING_SITE_URL` or marketing-host reference remains in the foundation docs in scope (grep).

**Lock-held proofs the coordinator must run (names only).**
- Every run: B-PROTECTED with the three `--expect-hash` values above (plus the Result's), B-SCOPE with the 47-pattern
  allow-list, B-LITERALS (`lint-base`, with `lint` for B-SECRET-CANARY).
- `migrations`: B-AUDIT-CATALOG (AUDIT.md rows ↔ 00018: unchanged 46/39) and B-MIGRATE-CLI (MIGRATIONS.md blocks
  unchanged).
- `ops` identity: B-IDENTITY-E2E now executes the updated `identity:browser` line (`signup-light`, `signup-dark`,
  `domain-platform` with the new E-DOM-4, then the unchanged projects), with B-IDENTITY-PROBES/-PRIVACY/-CONFIG and
  B-MCP-E2E; plus the round-16 list (lint-web/unit-web rerun, B-ENV-EXAMPLE, B-SHELL, B-COMPOSE, B-MCP-SOURCE,
  ci-controls, B-TOGGLE-OFF, B-CORE-ERROR-OUTPUT).
- Then M1, final local coverage and the MR (Plan §7 steps 5–6).

**Remaining:** the lock-held proofs above (b3 browser run first), M1, any fixes and affected reruns, final local coverage,
MR (fast + Duo + Codex source review), merged-SHA full local qualification, cleanup receipt. Round B/C1 source review
items remain Codex's.

Not self-accepted; Codex review pending.

### Round 18: slice 2 round C3 (variant copy lists)

**Rebinding.** Eighteenth writer (Opus 5.5 high; runtime effort not observable), single slice-2 implementation owner for
round C3 (one bounded harness correction), attempt 1 of B5 (no reset or extension); worktree
`.claude/worktrees/agent-<id>`, branch `g062-impl2-c3` created from `origin/g062-impl2` =
`7cf0842e376b088a3a4822eda9e78090577bd96b` after `git fetch origin staging g062-impl2`; branch and HEAD asserted before
any edit. `origin/staging` = `e0cb63faf3b6ea746e173d7f6180a36d43f2c2eb`, an ancestor of the base (no reconcile commit;
unchanged at the pre-commit re-fetch). Plan r5 `b2a612a2…` and the slice-2 release `f88acb4e…` are unchanged and not
edited. This Result before round 18: `8120895ed3c536ae9df9fa05a966d2935b1e8477f636582c2164f9a6efcd36a2`.

**Diagnosis (the coordinator's first lock-held run, root + off at `b707718a`).** `root` passed every check (including
B-BROWSER-ROOT); `off` failed B-OBS-OFF, B-BROWSER-OFF and B-TOGGLE-OFF (start, ready, check). The off-browser
diagnostic shows Turbopack `Module not found: Can't resolve '@/lib/signup-client'` from `app/auth/callback/route.ts`,
`app/page.tsx` and `components/signup/landing.tsx`. Cause: `variants_off` builds the intelligence-disabled variant from
its own render (`templates/base` only) and then copies only the paths in the `APPLICABLE` array from the repository.
Slice 2 made three copied web files import `@/lib/signup-client`, and made the copied reviewed server main import
`core/internal/signup` and `core/internal/signup/api`, but none of those paths, nor `web/app/signup`, were in
`APPLICABLE`. So the variant's web build failed, and its core build would have failed the same way. The C2 commits
since `b707718a` (`b707718a..7cf0842e`) change documentation and this Result only, so the diagnosis still holds at the
base.

**Enumeration inventory (every list that decides what a disposable copy carries).**

| Site | Mechanism | G062 impact |
|---|---|---|
| `run.sh` `APPLICABLE` (211–266), used by `variants_off` (the copy loop and the B-TOGGLE-OFF-ADDITIONS byte-identity loop) | explicit path list over the disabled render | **missing the G062 additions: fixed** |
| `run.sh` `TOGGLE_RENDERED` (`.env.example`, `VENDORS.md`) | composed from the root's reviewed runs | none: B-TOGGLE-OFF-ADDITIONS passed in the failing run (the G062 `.env.example` change is inside the pinned identity+domain tail) |
| `lib.sh` `copy_repo` / `impl_tree_hash` / `impl_files` (used by `variants_ops`, `variants_rename`, `variants_identity`, B-SOURCE-HASH, the repeat render and the other copies) | `git ls-files --cached --others --exclude-standard`, so the whole implementation | none: `ops` and `rename` already carry every G062 file |
| `lib.sh` `DEPS_ADDITIONS` | npm dependency pins, not paths | none (G062 adds no dependency) |
| `run.sh`/`lib.sh`/`validate-controls.sh`/`required-checks.tsv` pins over `run.sh` content or over `APPLICABLE`'s length | none: no sha256 over `run.sh` and no count pin (the PASS text prints `${#APPLICABLE[@]}`, now 105 instead of 102) | none |

**Change (commit `8c5af644`, `infra/validation/run.sh` only; `lib.sh` unchanged).** Before: the `APPLICABLE` array's
last group before the G018 R3 Docker ignore files was the G058 group (`core/internal/enrollment
web/lib/enrollment-client.ts web/app/join web/app/organization/enrollment`). After: that group is followed by

```
  # G062 (Plan r5 slice 2): the platform root signup package and its HTTP adapter (both imported by the reviewed server
  # main), web's signup client (imported by the copied root landing page, auth callback and signup components) and the
  # signup pages and routes (migration 00018, the hostname, domain, identity and host files, web/components/signup and
  # the tests are already covered by core/migrations, the G016 and G004 additions, web/components and web/tests).
  core/internal/signup web/lib/signup-client.ts web/app/signup
```

Items went from 102 to 105. Every other G062-touched path was already covered: `core/internal/hostname`,
`core/internal/domains`, `core/internal/config/domains.go`, `web/lib/host.ts` and `web/proxy.ts` (G016);
`core/internal/identity`, `web/lib/session.ts`, `web/lib/identity-client.ts`, `web/app/auth`, `web/app/sign-in`,
`infra/compose/identity.yml` and `infra/identity` (G004); `core/migrations` (00018); `web/components` (signup components);
`web/tests` (specs and unit tests); `web/playwright.config.ts`; `core/cmd/server/main.go`. The documentation in
`APPLICABLE` is copied as before. Nothing else was added: no path the variant does not need.

**Stack-free closure check** (throwaway `.local/tmp/closure.py`, git-ignored, not committed; `python3 -I`). It builds
the variant tree statically, exactly as `variants_off` does. First the render: every file of
`skills/saas-foundation/templates/base`, using the renderer's own path mapping (`dot_` → `.`, `.tmpl` dropped;
`render.go` renders `templates/intelligence` only when intelligence is enabled). Then, for each `APPLICABLE` item parsed
from `run.sh`, the render's path is removed and the repository's path is copied in. Because `core.Dockerfile` runs
`go vet ./...` and `go test ./...`, and the web `tsconfig.json` includes `**/*.ts` and `**/*.tsx`, the check scans the
**whole** variant `core/` and `web/` trees, not only what is reachable from `main.go` or the pages:
(a) every `@/…` and relative import in every web `.ts/.tsx/.mjs/.js` file must resolve to a variant file;
(b) every module-local Go import (`core/…`) in every core `.go` file, test files included, must resolve to a variant
directory that holds a non-test `.go` file.
It also lists every tracked `core/` and `web/` file that is absent from the variant.

| Run | Items | Variant files | Web files / local imports | Go files / module imports | Absent tracked core/web files | Problems |
|---|---|---|---|---|---|---|
| before (base `7cf0842e`) | 102 | 493 | 159 / 553 | 230 / 265 | (not yet measured) | **12**: `main.go` → `core/internal/signup`, `core/internal/signup/api`; `@/lib/signup-client` from `app/auth/callback/route.ts`, `app/page.tsx`, `components/signup/landing.tsx` (exactly the diagnostic); `tests/unit/signup-client.test.mjs` → `lib/signup-client.ts`; `tests/unit/signup-forms.test.mjs` → the six `app/signup/*/route.ts` |
| after (`8c5af644`) | 105 | 514 | 168 / 593 | 242 / 283 | **0** | **0** |

Assumption: the files the render supplies, which `APPLICABLE` does not overlay, are `core/internal/config/config_test.go`,
`web/lib/config.ts`, `web/next-env.d.ts`, `web/next.config.ts`, `web/postcss.config.mjs` and `web/tsconfig.json`. None
is in `render-customizations.tsv`, so the repository copy equals the render (B-RENDER-PLAN enforces this), and G062
touched none of them (`git diff origin/staging HEAD` on them is empty). Template directives in `.tmpl` files do not
affect import lines. The closure covers static import resolution only. Symbol-level and type-level compilation inside a
package is covered by the absent-file inventory (every tracked core/web file is present in the variant) together with
root's passing builds, and is proven finally by the lock-held `off` build.

**Commands** (worktree root; one offline shellcheck container and nothing else; no suite, stack, image pull or build,
network or `local-harness.lock`):

| # | Command | Exit |
|---|---|---|
| 129 | `git fetch origin staging g062-impl2`; `origin/g062-impl2` = `7cf0842e` asserted; `git switch -c g062-impl2-c3 origin/g062-impl2`; branch/HEAD asserted; `git merge-base --is-ancestor origin/staging HEAD` | 0 |
| 130 | read-only inventory: `grep` of `APPLICABLE`/`variants_*`/`copy_repo`/`"$RENDER" render`/`cp -a "$REPO` in `run.sh` and `lib.sh`; `render.sh`/`render.go` template-set logic; `core.Dockerfile` build/vet/test lines; `web/tsconfig.json`; `grep` for `run.sh`/`APPLICABLE` pins in `infra/ci` and `infra/validation` | 0 |
| 131 | `python3 -I .local/tmp/closure.py <worktree>` before the edit | 1 (12 problems, above) |
| 132 | edit `APPLICABLE`; `python3 -I .local/tmp/closure.py <worktree>` (twice, the second with the absent-file inventory) | 0, 0 |
| 133 | `bash -n infra/validation/run.sh` | 0 |
| 134 | `docker ps -a --filter name=platform-test-` (empty); `docker image inspect koalaman/shellcheck:v0.11.0` (local, `sha256:61862eba…`) | 0 |
| 135 | `docker run --rm --name platform-test-g062-c3-shellcheck --label platform-test=g062-a1 --network none -v <worktree>:/repo:ro -w /repo koalaman/shellcheck:v0.11.0 -S warning -x -P SCRIPTDIR infra/validation/run.sh infra/validation/lib.sh` (B-SHELL's flags) | 0 (no findings) |
| 136 | `docker ps -a --filter name=platform-test-` after: empty (container removed by `--rm`) | 0 |
| 137 | `git diff --check` | 0 |
| 138 | `printf` of `infra/validation/run.sh` and this Result through `grep -vE -f g062-allow47.txt` | 1 (no output: both in scope) |
| 139 | commit `8c5af644`; `git fetch origin staging g062-impl2` before the Result commit: both unchanged (`e0cb63fa`, `7cf0842e`) | 0 |

`run.sh` sha256: base `15916865b2325ef6b417de703f655759e53c0cd2040317d099cd7597476842fd` → after
`5de497abb985d68777b91517ff6cf6bed7a67e67f9a1a62a3fa57dee08860f63` (pinned nowhere).

**Lock-held proofs the coordinator must run (names only).**
- `off`: B-TOGGLE-OFF-ADDITIONS (now 105 additions byte-identical), B-TOGGLE-OFF (start, ready, check),
  B-OBS-OFF, B-BROWSER-OFF: the three failures this round corrects.
- `rename`: B-SOURCE-HASH, B-RENAME, B-BROWSER-RENAME, B-DBRENAME. `copy_repo` is unaffected by this change, but this
  suite has not run on the slice-2 tree.
- `ops` with the identity browser: B-IDENTITY-E2E (the round-17 `identity:browser` line with `signup-light`/`signup-dark`),
  B-IDENTITY-PROBES/-PRIVACY/-CONFIG, B-MCP-E2E, B-OUTAGE, B-PERSIST, B-SECRET-ROTATION.
- `ci-controls`, plus `lint` with `lint-base` for B-SHELL (`run.sh` changed), B-SCOPE and B-PROTECTED (this Result's new
  hash). The round-16/17 lists otherwise stand; root's passing checks at `b707718a` remain attributable, because this
  change affects only the off variant's copy list.

**Deviations and questions for Codex.** None. The fix is a pure list addition within the 47 patterns. `lib.sh` is
untouched. Resource receipt (round 18): one `--rm` shellcheck container with the `platform-test-g062-` prefix and the
label `platform-test=g062-a1`, verified removed. No image, volume or network was created.

Not self-accepted; Codex review pending.

### Round 19: slice 2 round D (first browser run diagnosis and corrections)

**Rebinding.** Nineteenth writer (Fable 5.1 high, test-failure diagnosis and correction; runtime effort not observable),
single slice-2 implementation owner for round D, attempt 1 of B5 (no reset or extension); worktree
`.claude/worktrees/agent-<id>`, branch `g062-impl2-d` created from `origin/g062-impl2` =
`f076fe4513f1a188b5eb556604427891da37700e` after `git fetch origin staging g062-impl2`; branch and HEAD asserted before
any edit. `origin/staging` = `e0cb63faf3b6ea746e173d7f6180a36d43f2c2eb` = the merge base (no code or management
movement; no reconcile commit). The coordinator's 47-pattern extract equals the staging allow-list minus `#` lines
(`diff` empty). Plan r5 `b2a612a2…` and the slice-2 release `f88acb4e…` (re-read from `origin/staging`, digest
reproduced) are unchanged and not edited. This Result before round 19:
`03a14dde347e31a8b445142ea78c31bde8ca6e9d7e777def6e4b275122379acf`.

**Evidence read (read-only; `~/.local/state/platform-evidence/g062-a1/s2c2-batch1-7cf0842e/evidence/`).**
`diagnostics/idn-browser.log` (`fd95a57c…`: 137 passed, 14 unexpected, the stacks and call logs),
`browser-identity-tests.tsv` (`9f60bf17…`), `results.tsv` (`dab67553…`), `diagnostics/idn-logcheck-core.log`
(`9e717360…`), `idn-logcheck-web.log` (`5de580c5…`), `idn-scan-report.log` (`094cc516…`), `diagnostics/harness.log`
(`ede7100e…`, lines 100-130). The harness preserved no core/web/proxy/emulator log, screenshot or `error-context.md`
(the `$WORK/browser-identity/` report directory was not copied), so the flow diagnosis below rests on the logcheck
counts, the spec, the product code and the pinned emulator's source (round 16's copy of its `dist`,
`gen/w16-emu-dist/workos/routes/auth.js` `0215e687…`, `helpers.js` `991ce3ed…`). `[E-DOM-4]` (the round-16 rewrite):
**expected** in `domain-platform` with S-ERRORS (row 59); every other project passed; `[E-SUP-1]`/`[E-SUP-4]` expected in
both signup projects.

**Diagnosis (three independent defect classes, plus the harness catalogue).**

1. **`route.fetch` TLS (E-SUP-5 ×4 in both projects; certain).** `hold()` used `route.fetch({maxRedirects: 0})`, which
   runs on Playwright's Node client: it verifies the chain against Node's roots (the run CA is self-signed, hence
   "self-signed certificate in certificate chain") and would also refuse a wildcard `*.localhost` name (Node rejects
   `*.<tld>` wildcards), while only Chromium carries the SPKI pin. identity.spec.ts's only `route.fetch` targets the
   plain-http emulator, which is why it never failed. *Class: spec defect.* The request never left the test runner, so
   none of the eight failing E-SUP-5 runs created an Organization (consistent with the counts in 3).
2. **The emulator prompts on every authorize (E-SUP-3 ×2; certain from source).** The overlay runs the emulator with
   `--interactive-password`; its `GET /user_management/authorize` (auth.js 335-390) renders the hosted login page
   whenever the interactive flag is set and never consults a browser session, so the spec's premise "silent return
   while the provider session is live" is false in the fixture. identity.spec.ts `[E-ORIGINS]` already relies on the
   prompt appearing for a second sign-in of the same browser. E-SUP-3 clicked `signup-start` and waited for
   `/signup/details` while the browser sat on the prompt (S-ERRORS passed, no console error, zero signup-route calls:
   see 3). The same premise sits under every automatic continuation (`landedCreated` after a create, a retry or a
   `page.goto` of a held reference): the complete hop's 303 to the provider reaches the prompt and would have hung
   there once defect 3 is fixed. *Class: spec defect (and the documentation sentence that mirrored it).*
3. **web's `PLATFORM_BASE_DOMAIN` in the overlay (E-SUP-2/E-SUP-6 ×2 each; high confidence).** The harness builds the
   stack `.env` from `.env.example` (`make_env_file`, run.sh 290-301), so web ran with `PLATFORM_BASE_DOMAIN=foundation.test`:
   `identity.yml` overrides core's value to `localhost` but set nothing on `web`. `rootOfTenantRequest`
   (`web/lib/signup-client.ts` 461-469) builds the root candidate `https://<PLATFORM_BASE_DOMAIN><port>` and refuses the
   hop as `not_found` when the request origin is not one label under it, so every `/signup/continue` on a new
   `<slug>.localhost:543` tenant answered web's fixed 404 page: a main-frame 404 is exactly Chromium's
   "Failed to load resource: the server responded with a status of 404" console error, and the `waitForURL` then timed
   out. The core counts confirm it: the run's 64 signup-route completions equal E-SUP-4's 12 per project (status,
   slug, status, four refused creates with their four status reads, abandon) + E-SUP-5's 3 per project (status, slug,
   status; the held create never reached web) + E-SUP-2/E-SUP-6's 4 per project (status, slug, status, create) =
   24 + 24 + 16, leaving **no** challenge/prove/complete call and no E-SUP-3 call at all. The emulator's
   `*.localhost` redirect host (round 16) is matched by `helpers.js` 743-744 (`hostname.endsWith(pattern.slice(1))`),
   so the final hop is not where it stopped. *Class: harness/overlay defect (the documented `identity:browser` command
   has the same gap for an operator's own `.env`).* Not a product defect: with a correct domain the route derives the
   root from configuration as designed (Plan r5 §4.4), and the unit tests of round B/C1 cover it with
   `PLATFORM_BASE_DOMAIN=example.test`.
4. **B-IDENTITY-PRIVACY's `logcheck validate` failure (certain; independent of the browser outcome).**
   `completion.schema.json` enumerates every core route a completion record may carry and had no
   `/internal/v1/signup/*` entry; the 64 signup completions (`route` = the registered path, as `routeOf` derives it)
   were `completion-invalid`, so `logcheck validate --service core` failed and B-IDENTITY-PRIVACY would have failed
   even with every browser test green. The `POST unmatched …` buckets are pre-existing MCP transport records, not
   signup. *Class: harness catalogue defect (round A registered the routes; no round added them to the schema).*
5. **B-SECRET-CANARY / the browser-evidence scan (consequence of 1).** The 32 hits are the `cookie:` line of
   Playwright's call log in the eight E-SUP-5 `error-context.md` files and the same text in `results.json`
   (`idn-scan-report.log`): the `__Host-signup` value, which the harness redacts in `idn-browser.log` but not in the
   report directory. They clear once the held requests succeed. See the answer to the secret-leak question below.

Round 16's two "assumptions to verify in b3": (i) `route.fetch`'s Set-Cookie reaching the browser store is moot (the
spec now applies the held answer's cookies itself and asserts their attributes); (ii) Chromium's SPKI pin accepting a
new `<slug>.localhost` name is still unexercised by this run (the browser stopped at web's 404 after the TLS handshake
had already succeeded for the tenant host, which is weak positive evidence: the 404 page itself was served over TLS to
that host). It remains the one assumption the rerun proves or refutes.

**Changes (every path inside the 47 patterns; `git diff --name-only origin/g062-impl2 HEAD | grep -vEf allow47` empty).**

| Commit | Path | Before → after |
|---|---|---|
| `78c557d2` | `web/tests/browser/signup.spec.ts` (`976c9b5a…` → `c247f537…`) | `hold()` replays the held request from a driver to `WEB_DIRECT` (`http://127.0.0.1:3000`, web's own port in the shared namespace, as identity.spec.ts `[E-CB-8]`/`[E-LO-8]` already do) with the browser's method, `Host`, headers (minus `content-length`, `accept-encoding`, `cookie`), the context store's cookies for that URL and the post body, `maxRedirects: 0`; new `applySetCookies()` parses each `Set-Cookie` of the answer, asserts exactly `Path=/`, `Max-Age`, `HttpOnly`, `SameSite=Lax`, `Secure` (no `Domain`), registers the value, and clears or adds the host-only cookie on the browser context with that lifetime; a transport error is rethrown with its first line only. New `continueAtPrompt(page, who, i)` = `enterCredentials` + `landedCreated`, used at every automatic continuation (E-SUP-2, E-SUP-3, E-SUP-5 ×4, E-SUP-6); E-SUP-3 starts through `signupLogin(page, ALICE)` (prompt, then details); the fourth E-SUP-5 test is retitled as the prompt path and its two no-op emulator cookie clearings are removed. New `expectSecret(value, label)` replaces the five `toMatch(SECRET)` assertions on secret values. Header comments state the direct-web hold, the interactive emulator and the G057 boundary. |
| `fc5df9c9` | `infra/compose/identity.yml` (`363bd139…` → `51b468be…`) | `services.web.environment` gains `PLATFORM_BASE_DOMAIN: localhost` (with the reason); nothing else changes (`yq` parse: command, core env and hosts as before). |
| `fc5df9c9` | `docs/foundation/DOMAINS.md`, `docs/foundation/IDENTITY.md` | Test overlay: "runs core and web with `PLATFORM_BASE_DOMAIN=localhost` (web derives the platform root of a G062 tenant continuation hop from it)"; the browser-projects sentence names the emulator's interactive prompt on the way into the tenant instead of "the provider prompt when its session is not live". |
| `511885ff` | `infra/validation/tools/internal/logrecord/completion.schema.json` (`fe2cf6ea…` → `95dd6602…`), `logrecord_test.go` (→ `19fae697…`) | both `route` enums (the general one and the core branch) gain exactly the seven registered signup paths; the test gains one positive row per path (plus a 409 create, a 401 challenge, a 503 complete and a 405 other-method) and the negative rows (prefix, trailing slash, query, unknown path, the path on web). |

No product code changed (no `web/app`, `web/lib`, `core` or schema edit); no security-model, contract or Plan change.
The one overlay change is configuration the Plan already assumes (`PLATFORM_BASE_DOMAIN=localhost` for the identity
environment, Plan r5 §1.1/§5.6; DOMAINS.md's overlay paragraph stated it for core only).

**Secret-leak question (for Codex).** Playwright writes `error-context.md` only for a failed test, and `results.json`
carries a test's error messages only when it fails; both are scanned by the browser-evidence scan and the whole work
tree by B-SECRET-CANARY, so a registered secret cannot reach saved evidence on a PASS through these paths. On a FAIL two
sources can carry one: (a) Playwright's own call log of a failed API request (the `cookie:` header line), which the
new `hold()` truncates to the error's first line before rethrowing; (b) any failed assertion whose message prints a
secret-valued received value (`toMatch(SECRET)`, `toBe(<cookie value>)`), which `expectSecret` removes from this spec.
Not changed: identity.spec.ts and the other identity-environment specs keep `toMatch(SECRET)`-style assertions
(outside this round and the 47 patterns), so a future failure there could still print one; a failing
`expect(record!.value).toBe(formRecord(...))` (the form record, a registered but non-secret value) would print the
record. The scans are the control that catches any of this; none of it is a product leak.

**Stack-free proofs (plain native Docker, one container at a time, `--network none`, read-only mounts, names
`platform-test-g062-w19-*`, label `platform-test=g062-a1`; no pull, build, Compose, DB, browser stack, Playwright,
`run.sh`, lock or `npm install`).**

| # | Command | Exit |
|---|---|---|
| 140 | `git fetch origin staging g062-impl2`; `origin/g062-impl2` = `f076fe45` asserted; `git checkout -b g062-impl2-d f076fe45…`; branch/HEAD asserted; `git merge-base HEAD origin/staging` = `e0cb63fa` (= staging tip); allow-list `diff` empty | 0 |
| 141 | `docker ps -a --filter name=platform-test-` empty; `docker run --rm … -v g058-a1-gomod:/gomod:ro golang:1.27.3-alpine ls …/jsonschema/v6/@v …/x/text/@v` (v6.0.3 and v0.14.0 present) | 0 |
| 142 | `docker run --rm … node:24-slim node /w/gen/tscheck.mjs signup.spec.ts identity.spec.ts` (types stripped with `node:module` `stripTypeScriptTypes`, then `node --check`; syntax only, not a type check; script `gen/w19/tscheck.mjs`) | 0: `ok` ×2 |
| 143 | `docker run --rm … mikefarah/yq:4.48.1 e '.services.web' /identity.yml` (parses; the new key beside `IDENTITY_COOKIE_PROFILE`) | 0 |
| 144 | `docker run --rm … -w /src golang:1.27.3-alpine gofmt -l ./internal/logrecord/` | 0 (no output) |
| 145 | `docker run --rm … -v g058-a1-gomod:/gomod:ro -e GOFLAGS=-mod=readonly -e GOPROXY=file:///gomod/cache/download -e GOSUMDB=off -e GOTOOLCHAIN=local -e GOMODCACHE=/tmp/gomod -e GOCACHE=/tmp/gocache golang:1.27.3-alpine go test -count=1 ./internal/logrecord/` (the prepared cache read-only as a file proxy; extraction in the disposable container) | 0: `ok validationtools/internal/logrecord` |
| 146 | `git diff --check`; `git diff --name-only origin/g062-impl2 HEAD \| grep -vEf allow47` | 0 / empty |
| 147 | commits `78c557d2`, `fc5df9c9`, `511885ff`; `docker ps -a --filter name=platform-test-` after: empty | 0 |

**Not run here (cannot run stack-free):** TypeScript type checking of the spec (`web-check`/`node_modules` absent),
every `run.sh` suite, Playwright. The web unit suites are unaffected (no web source change).

**Lock-held rerun for the coordinator (names only) and what each proves.**
- `ops` (the identity environment at this tip): **B-IDENTITY-E2E** `signup-light`/`signup-dark` — E-SUP-2/E-SUP-6
  prove defect 3 (the continue hop answers 303, the chain reaches the prompt) and, through `continueAtPrompt`, defect 2
  and round 16's SPKI assumption (the pinned callback on the new host, `/?status=tenant_created`); E-SUP-3 proves
  defect 2 at the signup login; the four E-SUP-5 tests prove defect 1 (`held POST /signup/create` 303 through
  `WEB_DIRECT`, the held `__Host-signup-form`/`__Host-sct`/`__Host-login` cookies then present with the asserted
  attributes, the cleared `__Host-sct` absent) and the rest of their scenarios. The unchanged projects are attributable
  from the previous run and need no selection beyond the `identity:browser` line (the harness runs the whole line).
  **B-IDENTITY-PRIVACY** proves defect 4 (`logcheck validate --service core` with 0 failures; the browser-evidence scan
  clean once E-SUP-5 passes), **B-SECRET-CANARY** defect 5. B-IDENTITY-PROBES/-CONFIG and B-MCP-E2E are unchanged.
- `lint` (B-COMPOSE does not compose `identity.yml`; B-ENV-EXAMPLE unchanged; B-SCOPE with the 47 patterns on the new
  paths), `ci-controls` (`validate-controls.sh` does not read the schema). The `vtools` image rebuilds from
  `infra/validation/tools` on the next run, carrying the schema (B-TENANCY-IT and every other `logcheck validate` user
  only gains accepted routes).
- A second failure would mean: E-SUP-5 `held … : …` with a connection error → web's port or namespace differs from
  identity.spec.ts's `WEB_DIRECT` (same container, so unlikely); a `held … to be 303` with 403 → web's Origin/Host
  derivation of the direct request (then compare the browser's headers); E-SUP-2 stopping on the prompt → the emulator
  prompt's selectors; a tenant-host TLS error at the callback → the SPKI assumption (ii) above (the fix would be an
  `identitysetup --hosts` entry per fixed slug, outside this round); a `completion-invalid` → a route path outside the
  seven (none is registered).

**Deviations and questions for Codex.** (1) The fourth E-SUP-5 test's title changed (the scenario tag `E-SUP-5` and the
browser matrix row are unchanged; the harness pins tags, not titles): the "provider session not live" case is not
distinguishable in the fixture because the emulator never returns silently, so the test now states what it proves
(the prompt path and the cancel). (2) The direct-web hold bypasses the TLS edge for the held request only (the
browser's own hops still go through it); the edge adds nothing but `Host`, which the replay sets. If Codex prefers the
edge in the loop, the alternative is `ignoreHTTPSErrors: true` on a separate Node request context (weaker, and the
wildcard tenant name still fails Node's hostname check), so this was not chosen. (3) `web/tests/browser/support/*`
stays untouched (outside the patterns). Resource receipt (round 19): five `--rm` containers with the
`platform-test-g062-w19-` prefix and the label, each verified gone; no image, volume or network created; scratch
`gen/w19/tscheck.mjs` under the evidence root.

Not self-accepted; Codex review pending.

### Round 20: slice 2 round E (second browser run corrections)

**Rebinding.** Twentieth writer (Fable 5.1 high, test-failure diagnosis and correction; runtime effort not observable),
single slice-2 implementation owner for round E, attempt 1 of B5 (no reset or extension); worktree
`.claude/worktrees/agent-<id>`, branch `g062-impl2-e` created from `origin/g062-impl2` =
`cbfaf2242acb81cb4c6ac49cb38c5a1a65304920` after `git fetch origin staging g062-impl2`; branch and HEAD asserted before
any edit. `origin/staging` = `e0cb63faf3b6ea746e173d7f6180a36d43f2c2eb`, an ancestor of HEAD (`git diff --name-only
HEAD...origin/staging` empty: no code or management movement; no reconcile commit). The coordinator's 47-pattern
extract equals the staging allow-list minus `#` lines (`diff` empty). Plan r5 `b2a612a2…` not edited. This Result
before round 20: `f68ae119117baf3a5e4fa63b1c956a9f53fe875b11e00b058ec08830b7b7c057`; the spec before:
`c247f537d17d7a17090b10098f6265a70901890ffa2d5b4dd1a43cd9dbe1118e`.

**Evidence read (read-only; `~/.local/state/platform-evidence/g062-a1/s2d-batch1-cbfaf224/evidence/`, the coordinator's
lock-held `lint lint-base lint-web provenance ops` at `cbfaf224`).** `diagnostics/idn-browser.log` (`e5e3b355…`:
145 passed, 6 unexpected = 3 tests × 2 signup projects, the stacks and the `waitForURL` call log),
`diagnostics/idn-logcheck-core.log` (`9146400c…`: the per-route/status completion counts of core, `pass: true`),
`browser-identity-tests.tsv` (`72797307…`), `results.tsv` (`d550bce3…`), `images.txt` (`61deaddf…`). Passed this run in
both projects: E-SUP-1, E-SUP-3, E-SUP-4, E-SUP-5 GET-never-re-mints, E-SUP-5 hosted-prompt path, E-SUP-6 (so round 19's
defects 1-3 and round 16's SPKI assumption are proven: the held create through `WEB_DIRECT`, the continuation chain to
the prompt and the pinned callback on the new `<slug>.localhost` host); B-IDENTITY-PRIVACY and B-SECRET-CANARY are out
of the failure list. As in round 19 the report directory (screenshots, `error-context.md`) was not preserved and no raw
core/web log with timestamps is in the evidence, so timing below is inferred.

**Diagnosis (two defect classes, both in the spec; no product defect).**

1. **Cookie selection by `context.cookies(url)` (E-SUP-2 ×2 and E-SUP-5 "final hop" ×2; certain from the received
   values).** E-SUP-2's received array for `context.cookies("https://<slug>.localhost:543")` was
   `["__Host-login", "__Host-signup", "__Host-signup-form", "__Host-login", "__Host-sid"]`: two `__Host-login` entries,
   i.e. the root's host-only cookies merged with the tenant's; the final-hop test's `cookieAt(context, tenant,
   LOGIN_COOKIE)` then returned the root's binding (`domain` `localhost` where `<slug>.localhost` was expected). Playwright's
   URL filter (`filterCookies` in `playwright-core`'s `server/network`) prefixes every stored domain with a dot and keeps a
   cookie when `"." + hostname` ends with it, so a host-only `localhost` cookie matches `.<slug>.localhost`; Chromium
   itself never sends a host-only cookie to a subdomain (RFC 6265 §5.1.3). The pinned image
   `mcr.microsoft.com/playwright:v1.63.0-noble` carries no `playwright-core` (searched, two `find` runs) and no web
   `node_modules` or dependency volume is present offline, so the source of the installed version was not re-read in
   this round; the behaviour is established by the run's own received values (the diagnosis does not depend on the
   source). *Class: spec defect of the helpers, not a product leak.* The same helper built `hold()`'s replayed Cookie
   header, so a held tenant hop (continue, complete) carried the root's `__Host-login` (and would have carried a
   duplicate name, which web reads as absent, once the tenant had its own): identity treats the binding as "shape only,
   never authority" and mints one when none is presented (`SignupContinuationTx`, `LoginPinned`), which is why those
   holds still answered 303; the replay is now what the browser sends. identity.spec.ts reads `context.cookies(origin)`
   for `__Host-sid`/`__Host-login`/`__Host-logout` on `org-a`/`org-b` only (no root cookies exist in its projects); it is
   unaffected and outside this round.
2. **The per-credential limiter (E-SUP-5 "captured continuation secrets" ×2; bucket certain, timing inferred).** The
   core counts hold exactly `POST /internal/v1/signup/prove 429: 2` and `GET /internal/v1/signup/status 429: 2` (one
   pair per project; challenge and complete, the per-tenant-origin buckets, show no 429; the per-user bucket is checked
   only on create, which shows none). `prove` is limited by the per-credential bucket alone (`authenticate`,
   `credentialRate 5, credentialBurst 10`, Plan r5 §4.5 (iii)/(iv)). The spec's sequence for that credential before the
   final "Try again": status after the login, slug, status, create (held), prove (held N), status (complete without a
   challenge cookie → recovery), status (R replay → recovery), status (tampered record), create (intent_used), status,
   status ×2 (stale records), status (restored record) = 13; the retry's create is the 14th and the re-mint's prove the
   15th, which met 429 → web's `/signup/details?status=rate_limited`, whose status read was the 16th (the second 429)
   and whose page the `waitForURL` log shows; `continueAtPrompt` then waited for a prompt that never came (360 s). Every
   other signup count reconciles with the spec's sequence per project (status 32 = 2+2+6+4+9+3+4+2 across E-SUP-2/3/4/
   5×4/6; organizations 201 = 10, 409 = 2, 400 = 3; challenge 200 = 9 incl. the re-mint's hop, 404 = 2; prove 200 = 8;
   complete 200 = 7; abandon = 1), so no call is duplicated by web, the hold replay or the retry click. The 15th call
   is refused only if the 15 fall inside about one second (`10 + 5·t ≥ 15`); the sibling scenarios' total durations
   (E-SUP-5 GET 1.2 s for a comparable flow, E-SUP-6 1.1 s) make that pace plausible; no timestamped log is in the
   evidence to pin it. *Class: (a) spec defect — the scripted pace exceeds a legitimate budget no person reaches*; not
   (c): the limiter is keyed as the Plan states and no double submission exists. The only other scenario above the
   burst is E-SUP-4 (11 credential calls), spread over its axe audits (5.7 s) and passing; the final-hop test spends
   exactly 10 (status, slug, status, create, status, create, prove, status, create, prove) and passes at any pace
   (`10 − 9 + 5·t ≥ 1`); E-SUP-2/3/6 and the prompt test spend 5, the GET test 7. Per-user creates: at most 3 of burst 5.
   Global start/callback buckets: 9 root logins per project against burst 20 / 50.

**E-SUP-2 and E-SUP-5 tails reasoned against the code (after the fixes).** E-SUP-2: `cookiesSentTo(tenant)` lists the
tenant's `__Host-login` and `__Host-sid` only (the SPKI/host proof already passed in E-SUP-3/E-SUP-6 with the same
hops); `__Host-sid` is Path=/, HttpOnly, SameSite=Lax, Secure with `Max-Age` ≤ 86 400 (`serialize`, `issuedSecret(…,
86_400)`); the root keeps `__Host-signup`/`__Host-signup-form` (r4-2; present in the received array) and the record
equals `formRecord` (passed at the captured test's line 668); `urlsClean` passed in E-SUP-3/E-SUP-6 with the same
emulator `redirect_uri`. Final hop: the held complete's answer clears `__Host-sct` and sets `__Host-login` on the
tenant (applied by `applySetCookies` with the tenant host), so the binding assertion now selects that cookie (600 s);
BOB's callback with the copied binding is `signup_user_mismatch` → web's `SIGN_IN_STATUS` fallback
`/sign-in?status=failed`; the creator's recovery re-mints (already_created, the next generation) and the chain forwards
the tenant's stale binding, which identity uses as shape only. Captured test after the landing: `statusAt(r1)` →
challenge 404 → recovery (one status read); `statusAt(prove2)` → `prove` on the completed row is
`signup_context_mismatch` (`info.Provisioned`, the row's nonce spent) → recovery, not `/signup` (a completed credential
still resolves: `StatusOf` has `completed`); the other browser's `complete1` with its generation-1 `__Host-sct` →
`continuation_invalid` → root recovery → `/signup` (no credential), within the allowed pair. Those three reads come
after the refilled budget and the landing.

**Changes (one file; `git diff --name-only origin/g062-impl2 HEAD | grep -vEf allow47` empty).**

| Commit | Path | Before → after |
|---|---|---|
| `6d197cbb` | `web/tests/browser/signup.spec.ts` (`c247f537…` → intermediate) | new `cookiesSentTo(context, url)` = `context.cookies(url)` filtered by RFC 6265 domain matching (host-only: exact host; dotted domain: suffix); `cookieAt`, `expectCookiesWhereExpected`, the E-SUP-2 root/tenant name checks, the E-SUP-4 root name check and `hold()`'s replayed Cookie header read through it; `expectCookiesWhereExpected` adds the whole-store control that no `__Host-signup`/`__Host-signup-form`/`__Host-sct`/`__Host-login`/`__Host-sid` is a domain cookie; `expectHostOnly`'s exact-domain check is kept as the positive control (a `.localhost` cookie would be selected for the root and fail it). |
| `cbd16f55` | `web/tests/browser/signup.spec.ts` (→ `9cd13f3d666c037ecbc204e477d82a087c48da13fb48f91632e59962951d6d78`) | `CREDENTIAL_RATE`/`CREDENTIAL_BURST` (5/10, cited to §4.5 and `http.go`) and `credentialBudgetRefilled(page)` = `waitForTimeout(burst/rate = 2 s)`, called once in the captured-secrets test before the restored record's retry (after its 13th credential call), with the count in a comment. No assertion changed. |

No product code, overlay, schema, contract, limiter budget or Plan change. No security-relevant change.

**Stack-free proofs (plain native Docker, one container at a time, `--network none`, read-only mounts, names
`platform-test-g062-w20-*`, label `platform-test=g062-a1`; no pull, build, Compose, DB, browser stack, Playwright,
`run.sh`, lock or `npm install`).**

| # | Command | Exit |
|---|---|---|
| 148 | `git fetch origin staging g062-impl2`; `origin/g062-impl2` = `cbfaf224` asserted; `git checkout -b g062-impl2-e origin/g062-impl2`; branch/HEAD asserted; `git merge-base --is-ancestor origin/staging HEAD`; allow-list `diff` empty, 47 lines | 0 |
| 149 | `docker ps -a --filter name=platform-test-` empty (before); `docker run --rm … mcr.microsoft.com/playwright:v1.63.0-noble find / -name network.js -path '*playwright-core/lib/server/*'` and `find / -xdev -type d -name 'playwright-core*'` | 0, both empty (no source copy in the image) |
| 150 | `docker run --rm … -v gen/w19/tscheck.mjs:ro -v signup.spec.ts:ro node:24-slim node /w/tscheck.mjs /src/signup.spec.ts` (types stripped, `node --check`; syntax only) | 0: `ok` |
| 151 | `git diff --check`; `git diff --name-only origin/g062-impl2 HEAD \| grep -vEf allow47` | 0 / empty |
| 152 | commits `6d197cbb`, `cbd16f55`; `docker ps -a --filter name=platform-test-` after: empty | 0 |

**Not run here (cannot run stack-free):** TypeScript type checking of the spec (`node_modules` absent), every `run.sh`
suite, Playwright. No Go or web source changed, so no unit suite is affected.

**Lock-held rerun for the coordinator (names only) and what each proves.**
- `ops` (the identity environment at this tip): **B-IDENTITY-E2E** `signup-light`/`signup-dark` — E-SUP-2 proves
  defect 1 (`expectCookiesWhereExpected` with the filtered jar and the new whole-store control; the tenant `__Host-sid`
  host-only); the E-SUP-5 final-hop test proves defect 1 at the binding (`domain` `<slug>.localhost`) and the faithful
  replay of the held tenant hops (continue without a binding, complete with `__Host-sct` only), then its previously
  unreached tail (BOB refused, the creator's re-mint); the E-SUP-5 captured-secrets test proves defect 2 (no
  `rate_limited`; the re-mint's chain, the replays and one Organization). The other six scenarios are attributable from
  this run and unchanged. **B-IDENTITY-PRIVACY** (the browser-evidence scan clean once no test fails) and
  **B-SECRET-CANARY** rerun with the suite. B-IDENTITY-PROBES/-CONFIG and B-MCP-E2E unchanged.
- `lint` (B-SCOPE with the 47 patterns on the one path). `lint-base`, `lint-web`, `provenance` are unchanged by a test
  spec edit; the coordinator decides whether the batch repeats them.
- A further failure would mean: E-SUP-2 cookie names → a cookie set with a `Domain` attribute (the new control names it);
  captured test still `rate_limited` → the 2 s refill was consumed by a slower-than-modelled sequence (then the count in
  the comment is wrong and the raw core log with timestamps is the next evidence to preserve); final hop at BOB's
  refusal → the callback's status mapping (compare `SIGN_IN_STATUS`).

**Deviations and questions for Codex.** (1) The pacing wait is a fixed 2 s `page.waitForTimeout` (the spec previously
had none; no lint rule forbids it): it models a person's pace against a budget the Plan sets, and nothing is retried on
429. If Codex prefers no sleeps, the alternative is splitting the captured-secrets scenario across two credentials,
which would weaken its "one credential, one Organization" proof, so it was not chosen. (2) The installed Playwright's
`filterCookies` source was not re-read offline (no copy reachable); the diagnosis rests on the received values.
(3) `web/tests/browser/support/*` and identity.spec.ts untouched. Resource receipt (round 20): three `--rm` containers
with the `platform-test-g062-w20-` prefix and the label, each verified gone; no image, volume, network or scratch file
created; the two commit-message files are untracked in the worktree and removed after the commits.

Not self-accepted; Codex review pending.

### Round 21: slice 2 round F (B-WEB-BOUNDARY false positive on the whole-store cookie control)

**Rebinding.** Twenty-first writer (Fable 5.1 high, mechanical correction; runtime effort not observable), single
slice-2 implementation owner for round F, attempt 1 of B5 (no reset or extension); worktree
`.claude/worktrees/agent-<id>`, branch `g062-impl2-f` created from `origin/g062-impl2` =
`c87da5559a7ef45818ae35511fa0272073b28436` (asserted) after `git fetch origin g062-impl2 staging`; branch and HEAD
asserted before any edit. `origin/staging` = `5d12c667f15c65896516549892d3b216db5419dd`, one commit past the branch's
merge base `e0cb63fa` (`docs: clarify G062 M1 and MR sequence`; `git diff --name-only HEAD...origin/staging` =
`management/v0/PROGRESS.md` only, which the task branch does not touch): merged as `a2bb5e74` (`--no-ff`, no conflict)
so the scope check below runs against the current target. Plan r5 not edited. This Result before round 21:
`82b2da95b362d981417458a7e246eb6cd0dc3bc8b1f559af0373e583429f36b9`; the spec before:
`9cd13f3d666c037ecbc204e477d82a087c48da13fb48f91632e59962951d6d78`.

**Failure (coordinator's lock-held run `d4ed97`, `lint-web`).** B-WEB-BOUNDARY: `cookies() outside lib/session.ts:
web/tests/browser/signup.spec.ts:253: for (const c of await context.cookies()) {`. The guard (`static_web_boundary`,
`infra/validation/run.sh:962`) greps every web source except `lib/session.ts` for the literal `cookies\(\)` and drops
only comment-only lines; its purpose is G004's rule that `next/headers` `cookies()` is handled in `lib/session.ts` alone.
Round E's whole-store control in `expectCookiesWhereExpected` read Playwright's `BrowserContext.cookies()` with no
argument, which carries the same literal. The guard cannot tell the two APIs apart; this is a harness false positive on
a browser-test helper, not a product boundary defect (no `next/headers` use was added anywhere). That was the only hit
in `web/` (verified by the same expression over `web/`, `node_modules`/`.next`/lockfile excluded).

**Change (Codex-specified method; one file, commit `1dca793d`, spec `9cd13f3d…` →
`2330f2408110023dbf51cd316024213d921fbbab11d9fa1d717c26d915ff83b7`).** The read is now
`context.cookies([ROOT, tenant])`, the root and tenant origins the helper already holds. Playwright's URL selection
(round 20, defect 1: every stored domain is dotted and kept when `"." + hostname` ends with it; path and Secure are
checked too) lists for the two tested hosts every cookie whose domain covers either of them, dotted or not: a
`.localhost` or `.<slug>.localhost` domain cookie set by either host appears for the tenant URL (and `.localhost` for
the root), so the control still fails on any signup/login/session cookie carrying a `Domain` attribute that reaches
either tested host, which is the leak the control exists for. Cookies on unrelated hosts (none exist in these projects)
are no longer inspected. The comment now reads "on the root and tenant hosts" instead of "anywhere in the store" and
states why the unfiltered selection is the right reader there. No assertion, `cookiesSentTo`, `expectHostOnly`,
`run.sh` or guard change; no product code touched. The `storageState()` alternative was not used (Codex chose the URL
form; the installed Playwright 1.63 types are not reachable offline, so its return shape was not verified here).

**Other web-boundary literals scanned (report only; no change).** `cookies\(\)`: none outside `lib/session.ts` after the
change. `document\.cookie`: none. `NEXT_PUBLIC_`: `web/lib/config.ts:5` in a comment only (`code_lines` drops it).
`process\.env`: only in `playwright.config.ts`, `tests/**` and `lib/config.ts`, none under `"use client"` files or
`components/` (the guard's scope for that pattern). The guard's other rules are unaffected by a test-spec edit.

**Stack-free proofs (plain native Docker, one container, `--rm`, `--network none`, read-only mounts, name
`platform-test-g062-w21-tscheck`, label `platform-test=g062-a1`; no pull, build, Compose, DB, browser, Playwright,
`run.sh`, lock or `npm install`).**

| # | Command | Exit |
|---|---|---|
| 153 | `git fetch origin g062-impl2 staging`; `origin/g062-impl2` = `c87da555` asserted; `git switch -c g062-impl2-f c87da555…`; branch/HEAD asserted; `git merge-base --is-ancestor origin/staging HEAD` 1 (staging one management commit ahead); `docker ps -a --filter name=platform-test-` empty (before) | 0 / 1 / 0 |
| 154 | `docker run --rm … -v gen/w19/tscheck.mjs:ro -v signup.spec.ts:ro node:24-slim node /w/tscheck.mjs /src/signup.spec.ts` (image `sha256:0e0ff40c…`; types stripped, `node --check`; syntax only) | 0: `ok` |
| 155 | `git diff --check` | 0 |
| 156 | `grep -rnHE 'cookies\(\)' web --exclude-dir=node_modules --exclude-dir=.next --exclude=package-lock.json \| grep -v '^web/lib/session.ts:' \| grep -vE '^[^:]+:[0-9]+:[[:space:]]*//'` | 1 (no hit) |
| 157 | commit `1dca793d`; `git merge --no-ff origin/staging` → `a2bb5e74`; `git merge-base --is-ancestor origin/staging HEAD` | 0 |
| 158 | `git diff --name-only origin/staging HEAD` (60 paths) `grep -vEf allow47` | empty |
| 159 | `docker ps -a --filter name=platform-test-` after: empty | 0 |

**Not run here (cannot run stack-free):** TypeScript type checking of the spec (`node_modules` absent), `lint-web`
itself and every other `run.sh` suite, Playwright. No Go or web application source changed.

**Lock-held rerun for the coordinator.** `lint-web` (B-WEB-BOUNDARY now passes the literal scan; the other rules are
unchanged); `ops` as round 20 lists it if the batch had not reached it (`signup-light`/`signup-dark` E-SUP-2 and the
E-SUP-5 final-hop test exercise `expectCookiesWhereExpected`); `lint` B-SCOPE unchanged (one path). A further
B-WEB-BOUNDARY hit would mean another `cookies()` literal elsewhere (none exists at this tip).

**Deviations and questions for Codex.** (1) The staging merge `a2bb5e74` is management-only reconciliation (PROGRESS.md)
so the scope check binds to the current target; it is not part of the correction. (2) The control's coverage is now the
two tested hosts rather than every host in the context (the projects create cookies on no other host). Resource
receipt (round 21): one `--rm` container, verified gone; no image, volume, network or scratch file created under the
repository; the commit-message files are untracked and removed after the commits.

Not self-accepted; Codex review pending.

### Round 22: slice 2 round G (MR !71 fast failure: unused `context` fixture in E-SUP-3)

**Rebinding.** Twenty-second writer (Fable 5.1 high, mechanical correction; runtime effort not observable), single
slice-2 implementation owner for round G, attempt 1 of B5 (no reset or extension); worktree
`.claude/worktrees/agent-<id>`, branch `g062-impl2-g` created from `origin/g062-impl2` =
`182017a3d1ee59560a13650b0d5b3affe16c5873` (asserted) after `git fetch origin g062-impl2`; branch and HEAD asserted
before any edit. `origin/staging` = `57fb45cf8a47ff6f648fd6cd0b7876e69dbe1a76`, two management-only commits past the
branch's merge base `5d12c667` (`c7630c1c` rounds A-F source acceptance, `57fb45cf` MR !71 fast failure review;
`git diff --name-only 5d12c667 origin/staging` = `management/v0/PROGRESS.md` and the two review records, none touched
by this branch): not merged (no conflict forces it; the scope check below runs against the merge base and says so).
Plan r5 not edited. This Result before round 22: `14e889af84f893a6cb23f4f0835fab0cbb2b6323a76c62bf03c30f90e3ce2474`;
the spec before: `2330f2408110023dbf51cd316024213d921fbbab11d9fa1d717c26d915ff83b7`.

**Hosted failure (preserved; not retried).** MR !71 fast pipeline `2971437635`, job `validate:checks`
`17344669966`, on the exact source `182017a3d1ee59560a13650b0d5b3affe16c5873`: four failed JUnit rows, all kept here
with that pipeline/job/SHA: **B-UNIT-CHECKS** (the substantive one), **B-REQUIRED-CHECKS** (derived: a required ID
failed), **phase:harness** and **phase:complete** (derived from the harness exit). The coordinator's local `unit-web`
rerun fails **B-UNIT-WEB** the same way. The one finding, from `npm run lint` (`eslint --max-warnings=0 .`):
`web/tests/browser/signup.spec.ts` `531:134 warning 'context' is defined but never used
@typescript-eslint/no-unused-vars`. Codex's independent diagnosis
(`management/v0/results/G062-mr71-fast-failure-review-20261021.md` on staging) names the same single cause. Because
`npm run check` is `npm run lint && npm run typecheck && npm run test:unit` (`web/package.json`; the `check` target of
`infra/docker/web.Dockerfile` runs it), `tsc --noEmit` and the web unit tests did not run in that job: typecheck
coverage of the slice-2 web changes is still outstanding until the next `unit-web`/hosted run.

**Why the local `lint-web` lane did not catch it.** `lint-web` (`infra/validation/run.sh` `suite_lint_web`) is the
harness's own web boundary checks: B-WEB-BOUNDARY, B-WEB-TOKENS, B-WEB-COMPONENTS, B-WEB-DESIGN and
B-WEB-ASSET-CONTROLS, grep/structure controls over the web tree. It never runs eslint. eslint runs only inside the
`web-check` image (`npm run check`), which is B-UNIT-CHECKS (`unit`, hosted `validate:checks`) and its exact subset
B-UNIT-WEB (`unit-web`). Rounds D-F changed only the browser spec and ran `lint-web`/`ops` locally; no `unit-web` run
followed round E's helper changes, so the warning first surfaced in the MR's fast pipeline. (Round 19's and 21's
`tscheck` is a syntax check and cannot see an unused binding.)

**Cause and change (one file, commit `111089c7`, spec `2330f240…` →
`8b4520c18ba491df4c48ab02341bd5f92b5eeb32afe65de027bf29e9ab248d8b`).** `[E-SUP-3]` was declared
`async ({ page, context }, testInfo) =>`; its body (sign-in on org-a, `accountAt`, `intent`, `signupLogin`,
`nameAndAddress`, `continueAtPrompt`, `organizationOf`, `urlsClean`) uses `page` and `testInfo` only: round E moved the
cookie reads of the other tests behind `cookiesSentTo`/`cookieAt` and left this test's `context` with no reader.
Confirmed unused by a fresh read of the whole test; the destructure is now `async ({ page }, testInfo) =>`. No
assertion, cookie helper, rule, configuration or `run.sh` change; no product code touched.

**Scan of the same warning class and for type errors (read of the whole spec; report only, no further change).**
Every import is referenced (`APIResponse` 322, `Locator` 459, `Response` 271, `Route` 313, `Browser` 129, `Cookie`
217; the rest throughout); every module-level constant, type and function is used; every function parameter and
destructured fixture of the other eight tests is read (`context` in E-SUP-1/2/4/5 ×3, `browser` in E-SUP-5 ×3,
`testInfo` in every test that takes it); the only `catch (error)` reads `error`. Types of the round D/E/F helpers
against Playwright 1.63's public API as documented: `cookiesSentTo`/`cookieAt` (`context.cookies(url)` →
`Cookie[]`, exported from `@playwright/test`; `domain`, `name`, `value` strings), `expectHostOnly` (`Cookie`'s
`domain`/`path`/`httpOnly`/`secure`/`sameSite`/`expires`), `expectCookiesWhereExpected` (`cookies([ROOT, tenant])`
takes a string array), `hold` (`page.route` handler `(route: Route) => Promise<void>` in both branches;
`request.headers()` spread into `Record<string, string>`; `postDataBuffer(): Buffer | null` coalesced to `undefined`
for `data`; `maxRedirects: 0`; `headersArray()` → `{name, value}[]`), `applySetCookies` (`clearCookies({ name,
domain })`, 1.43+; `addCookies` with `domain`+`path`), `credentialBudgetRefilled` (`waitForTimeout(number)`),
`continueAtPrompt`, `expectSecret` (`observe(value: string | undefined | null)` matches the `string | null`
arguments). `tsconfig.json` is `strict` without `noUncheckedIndexedAccess`, so `h["location"] ?? ""` and
`req.headers()` indexing type as `string`. `eslint.config.mjs` is `eslint-config-next` core-web-vitals + typescript
(recommended, not type-checked), so `no-unused-vars` is the only rule this spec's shape touches; the hosted run reported
no other finding anywhere under `web/`. The `identity.spec.ts` and `playwright.config.ts` deltas of this branch
(round C1) were re-read: `MARKETING` removed with its only readers, `ROOT` added and used. This is a read, not a
`tsc` run.

**Stack-free proofs (plain native Docker, one container, `--rm`, `--network none`, read-only mounts, name
`platform-test-g062-w22-tscheck`, label `platform-test=g062-a1`; no pull, build, Compose, DB, browser, Playwright,
`run.sh`, lock or `npm install`; the round-19 script recreated as `/tmp/g062-w22/tscheck.mjs` outside the repository).**

| # | Command | Exit |
|---|---|---|
| 160 | `git fetch origin g062-impl2`; `origin/g062-impl2` = `182017a3` asserted; `git checkout -b g062-impl2-g origin/g062-impl2`; branch/HEAD asserted; `docker ps -a --filter name=platform-test-` empty (before) | 0 |
| 161 | `git diff --check` (after the one-line edit) | 0 |
| 162 | `docker run --rm --name platform-test-g062-w22-tscheck --label platform-test=g062-a1 --network none -v tscheck.mjs:ro -v signup.spec.ts:ro -v identity.spec.ts:ro node:24-slim node /w/tscheck.mjs /src/signup.spec.ts /src/identity.spec.ts` (image `0e0ff40c39bc`; types stripped, `node --check`; syntax only) | 0: `ok` ×2 |
| 163 | `git fetch origin staging` (`57fb45cf`); `git merge-base HEAD origin/staging` = `5d12c667`; `git merge-base --is-ancestor origin/staging HEAD` | 1 (staging two management commits ahead; not merged) |
| 164 | `git diff --name-only origin/staging HEAD \| grep -vEf allow47` → the three staging-only management paths; `git diff --name-only 5d12c667 HEAD` (60 paths) `\| grep -vEf allow47` | empty (merge base) |
| 165 | commit `111089c7` (spec); `docker ps -a --filter name=platform-test-` after: empty | 0 |

**Not run here (cannot run stack-free):** eslint, `tsc --noEmit`, the web unit tests (`node_modules` and the
`web-check` image are absent on this host), `lint-web`, `unit-web` and every other `run.sh` suite, Playwright. No Go or
web application source changed.

**Rerun for the coordinator.** `unit-web` (B-UNIT-WEB: eslint now clean on this finding, then `tsc --noEmit` and the
web unit tests run for the first time on the slice-2 web changes; a type error there would be the next, hidden
finding); the exact-head MR !71 fast pipeline on the pushed tip (B-UNIT-CHECKS and the three derived rows). `lint`
B-SCOPE unchanged (one allowed path). The browser projects are unaffected (a fixture no longer requested; E-SUP-3's
behaviour and the harness tag are the same).

**Deviations and questions for Codex.** (1) `origin/staging` was not merged into the branch this round: the two new
commits are management-only and conflict-free, the scope check binds to the merge base `5d12c667` as stated, and the MR
pipeline runs on the merge result anyway. (2) Nothing else changed in the spec: the scan found no second unused
binding and no type error by reading; typecheck proof remains the coordinator's `unit-web`/hosted run. Resource
receipt (round 22): one `--rm` container, verified gone; no image, volume, network or scratch file created under the
repository; the commit-message files and `tscheck.mjs` live under `/tmp/g062-w22/` and are removed after the push.

Not self-accepted; Codex review pending.

### Post-merge local full qualification ledger (merged 84cbab04)

**Binding.** Docs-only writer (Opus 5.5 high; runtime effort not observable), Result-only round within attempt 1 of B5
(no reset or extension); worktree `.claude/worktrees/agent-<id>`, branch `g062-s2-qualification-ledger`
from `origin/staging` = `7d62e8ffed4fa8617135d4edb3caa433da672961` after `git fetch origin staging`, fast-forwarded
before commit to `3e8b68d518be9fac6f7b2254f271cc1ddc1fdfd9` (management-only), which adds Codex's evidence-only
CORRECTION LOOP (qualification review (not included), sha256
`3319b7e514ceab0ac8a4e7e177d648c5ddf3434f9d553f0e5dc474dd94bad1b0`) that this section answers. No product, test,
harness or other management file edited; no build, test, container or lock used. Source: the coordinator's ledger
`g062-a1/postmerge-84cbab04-qualification-ledger.md`. Artifact-table paths are relative to
`<home>/.local/state/platform-evidence/`; commands are written in full. Merge facts, artifact hashes and the run 1/3b
check-row diff were re-verified read-only this round. Times are local (UTC-4), as in the ledger.

**Merge facts.**

- MR !71 merged: merge commit `84cbab04045e3a19056f5ac2a14eebc5c4b64abc`, parents `57fb45cf8a47ff6f648fd6cd0b7876e69dbe1a76`
  + `3f2ee030847de2fe29be58dcd36153a92ac08158`, tree `48c5244ab43a759eafe8bf7844db4e8c976e0318`. Source branch deleted
  (`git ls-remote origin g062-impl2` empty).
- Fast pipeline `2971500358` on source head `3f2ee030`: **SUCCESS** (`validate:checks` job `17345153225`;
  `validate:integration` manual, not run).
- Checkout under test: coordinator checkout detached at `84cbab04`, clean, no ignored files, no `platform-test-*`
  containers, Docker Desktop mounts 19% before each run.

**Runs.**

| # | Run | Exact command | Window | Result |
|---|---|---|---|---|
| 1 | Full qualification, first | `flock -n <home>/.local/state/platform-evidence/local-harness.lock bash infra/ci/validate.sh --profile full --evidence-dir <home>/.local/state/platform-evidence/g062-a1/postmerge-84cbab04/ci-evidence` | 06:49–08:03 | **FAILED, exit 1** |
| 2 | Diagnostic, `ci-controls` only | `flock -n <home>/.local/state/platform-evidence/local-harness.lock bash infra/validation/run.sh --base 57fb45cf8a47ff6f648fd6cd0b7876e69dbe1a76 --allow-file <home>/.claude/jobs/580cdb2f/tmp/g062-allow47.txt --report-dir <home>/.local/state/platform-evidence/g062-a1/postmerge-84cbab04-diag/evidence ci-controls` | 08:04–08:50 | exit 1: B-CI-CONTROLS PASS (521/521); only B-SECRET-CANARY UNMET → B-REQUIRED-CHECKS FAIL (known standalone artifact: no `lint` suite in the run) |
| 3a | Full rerun, aborted launch | `flock -n <home>/.local/state/platform-evidence/local-harness.lock bash infra/ci/validate.sh --profile full --evidence-dir <home>/.local/state/platform-evidence/g062-a1/postmerge-84cbab04-r2/ci-evidence` | seconds | exit 16: `validate: location: cannot create --evidence-dir` (parent dir missing). No tests ran. |
| 3b | Full rerun | `mkdir -p <home>/.local/state/platform-evidence/g062-a1/postmerge-84cbab04-r2`, then `flock -n <home>/.local/state/platform-evidence/local-harness.lock bash infra/ci/validate.sh --profile full --evidence-dir <home>/.local/state/platform-evidence/g062-a1/postmerge-84cbab04-r2/ci-evidence` | 08:51–10:04 | **PASS, exit 0** |

Both full runs (1, 3b): `tested_content=local head=84cbab04045e3a19056f5ac2a14eebc5c4b64abc`; `profile pipeline=full
executed=full reason=local-explicit base=none`; suites executed `lint lint-base lint-web lint-core controls ci-controls
provenance contracts root migrations ops rename off` (13, coverage=complete, partition=complete); harness deadline
14400s; cleanup verified (0 resources removed); bootstrap not-applicable; registry removed.

**Run 1 failure record (cause UNKNOWN).**

- Harness exit status 1; phase rows harness=1, complete=14, all others 0. `validate: complete: required check
  B-CI-CONTROLS is not PASS`. Check rows: 120 PASS, 2 FAIL (B-CI-CONTROLS and the derived B-REQUIRED-CHECKS), plus INFO
  rows. Wrapper console: junit.xml 125 cases, 4 failing (supplementary): those 2 checks + `phase:harness` +
  `phase:complete`.
- Every other required check PASS, including lint*, controls, provenance, contracts, root, migrations, ops
  (B-RECOVERY-DRILL, B-MAINTENANCE-DRILL, B-OBS-*, B-OUTAGE, B-PERSIST, B-SECRET-ROTATION,
  B-IDENTITY-IT/E2E/PROBES/PRIVACY/CONFIG, B-MCP-E2E, B-SOURCE-HASH), rename, off, B-DEPS-CONTROLS, B-WRAPPER-CONTROLS,
  B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE.
- **The failing control name(s) were NOT captured.** Both full runs, the failing run 1 and the passing run 3b, ended
  `published=none` and retained only `storage.tsv` in their evidence directory. In `infra/ci/validate.sh` `finalize`
  (lines 1814-1816) publication happens only when `CI_MODE` is set and a publish target exists, so a local profile run
  publishes nothing by design; this is not a failure-only policy. After run 1, `ci-controls.log`, `harness.log`/the
  results table, `summary.json`/`phase-status.tsv` and `junit.xml` were not retained in the evidence directory, and no
  harness work directory remained. Run 1's junit counts and phase statuses are known only from the wrapper's console
  lines (`validate-console.txt`).
- **Cause: UNKNOWN.** This is not a fixed or confirmed flake. *Circumstantial observations only, not an explanation:*
  load/timing was suspected because `ci-controls` ran as a supervised background child among other suites (run.sh
  ~5904); the same 521 controls passed standalone at `cbfaf224` (slice-2 candidate, 45 min) and again standalone on the
  merged tree (run 2); slice 1's full run passed `ci-controls`.

**Reason for the broader rerun (run 3b) after run 2.** Run 2 proved the controls pass on the merged tree when run alone,
but cannot reproduce the original configuration (background child under full-suite load) or prove a full-profile PASS.
A local profile run retains no per-suite detail (publication is CI-only, above), so a failure-capturing rerun was not
possible through the wrapper. The hypothesis under test
was load/concurrency; the only way to test it in the same configuration, and to obtain an attributable full-profile
result for the exact merged SHA, was the same `validate.sh --profile full` command on the same unchanged SHA. No
source, harness or input changed between runs 1 and 3b.

**Run 3b result and diff against run 1.** Harness exit 0; all 13 phases status 0; junit.xml 123 cases, 0 failing
(supplementary); 122 check rows PASS including B-CI-CONTROLS and B-REQUIRED-CHECKS; seven INFO rows: PREFLIGHT,
PROBE, DEPSCAN-PINS, BACKGROUND ×2 (it appears twice), BUILD-LEDGER, IMAGES; final status 0; `published=none` (local profile writes no published bundle; same as
run 1). Check-row diff (`grep '^validate: check'` of both consoles), exactly two rows; everything else identical:

```
122c122
< validate: check B-CI-CONTROLS FAIL
---
> validate: check B-CI-CONTROLS PASS
129c129
< validate: check B-REQUIRED-CHECKS FAIL
---
> validate: check B-REQUIRED-CHECKS PASS
```

A pass on rerun does not explain run 1's failure.

**Preserved artifacts (sha256; re-verified this round).**

| File | sha256 |
|---|---|
| `g062-a1/postmerge-84cbab04/validate-console.txt` (run 1 console) | `9dba83b775313484e6da7f7ea67b12faa40a256ccb0608d8c9d1b6956be36e0f` |
| `g062-a1/postmerge-84cbab04/ci-evidence/storage.tsv` (run 1) | `b774d549260ebd4dcec4125f5d29340a7eaf18def8726da2673ab5c1e22e01b1` |
| `g062-a1/postmerge-84cbab04-r2/validate-console.txt` (run 3b console) | `3fa2a53a0d94c11e7ad448d6ad8b91b6c94f1559df0b2518d2750a43f867d509` |
| `g062-a1/postmerge-84cbab04-r2/ci-evidence/storage.tsv` (run 3b) | `7f8b5ddf8b370eff828ad00478cdd5c4123e56429cc02deeae8af2e219f88a93` |
| `g062-a1/postmerge-84cbab04-r2/first-launch-evidence-dir-error.txt` (run 3a) | `b4da9b6f043912be0363e097791376e9977fda253d6e167c724f2d320b519531` |
| `g062-a1/postmerge-84cbab04-diag/run-console.txt` (run 2 console) | `5db75f823d18ed51905b17115b86a555a5c62a7b0f1d84b6cd648f272584bb74` |
| `g062-a1/postmerge-84cbab04-diag/evidence/ci-controls.log` (run 2, 521 PASS) | `b131f03cb24c13ef35496cc6902a975a7cf028881ca46e00794053511a36cef6` |
| `g062-a1/postmerge-84cbab04-diag/evidence/results.tsv` | `e894ad265f38b2f052416d7304521ceac5c80ab6a47b6f5bb9bfcb21afc24044` |
| `g062-a1/postmerge-84cbab04-diag/evidence/ci-controls-timings.tsv` | `87ce9b4b97e81c8813fcd280d8fb3f99268700e7bb05f63f78e1669fa31c023e` |

The console files are verbatim copies of the coordinator background-task outputs (the wrapper's own stdout).

**Run-1 failure record.** Run 1's B-CI-CONTROLS failure on the exact merged SHA is **unexplained** (failing control
not captured, cause UNKNOWN). Preserve its FAILED outcome. The owner later waived its qualification impact after the
complete same-SHA run 3b passed; the failure is not reclassified, and its cause is not claimed as fixed or understood.
Capturing `ci-controls.log` (or other per-suite diagnostics) for a local profile run needs a separate decision; none is
started.

**AC mapping for this run.** Rows cite check IDs that PASS in run 3b's console, placed per rounds 8 and 16-20; a PASS
is evidence for the automated portion only. Unit-level tests (`B-UNIT-CHECKS`) are not in the local full profile;
their evidence is fast pipeline `2971500358`.

| AC | Run 3b check IDs (PASS) | Not covered by this run |
|---|---|---|
| AC1 | B-IDENTITY-E2E (`domain-platform` incl. rewritten E-DOM-4; `signup-light`/`signup-dark`), B-ENV-EXAMPLE, B-RENAME, B-BROWSER-RENAME, B-DBRENAME | M1 owner walk-through PENDING; G057 PENDING: real apex/`www` edge serving of the landing and the apex callback registration (Plan 5.6) |
| AC2 | B-SIGNUP-IT, B-IDENTITY-IT, B-IDENTITY-E2E (emulator) | G057 PENDING: real WorkOS identity creation/sign-in on the apex callback (Plan 5.6) |
| AC3 | B-IDENTITY-IT, B-ENROLLMENT-IT, B-MCP-IT, B-MCP-E2E, B-TENANCY-IT, B-IDENTITY-E2E | none beyond final review |
| AC4 | B-SIGNUP-CATALOG, B-SIGNUP-IT, B-MIGRATE-SET, B-MIGRATE-CATALOG, B-MIGRATE-IT, B-AUDIT-IT, B-RLS-CATALOG, B-RECOVERY-DRILL | none beyond final review |
| AC5 | B-SIGNUP-IT, B-DOMAINS-CATALOG, B-DOMAINS-IT, B-IDENTITY-E2E | G057 PENDING: real wildcard DNS/ingress/TLS resolution of a new `<slug>.B` (Plan 5.6) |
| AC6 | B-SIGNUP-IT, B-IDENTITY-IT, B-RECOVERY-DRILL | none beyond final review |
| AC7 | B-IDENTITY-E2E (`signup-light`/`signup-dark` E-SUP-1..6, local HTTPS browser), B-IDENTITY-PROBES | G057 PENDING: real WorkOS silent code return while the provider session is live, real wildcard tenant callbacks, production edge callback-query redaction (Plan 5.6) |
| AC8 | B-IDENTITY-E2E (`signup-*` incl. S-ERRORS), B-WEB-BOUNDARY, B-WEB-TOKENS, B-WEB-COMPONENTS, B-WEB-DESIGN | M1 owner walk-through and changed-flow feedback PENDING |
| AC9 | B-IDENTITY-PRIVACY, B-IDENTITY-PROBES, B-OBS-PRIVACY, B-OBS-LOGS, B-SECRET-CANARY, B-SIGNUP-IT | G057 PENDING: production edge log redaction (Plan 5.6) |
| AC10 | B-REQUIRED-CHECKS, B-SCOPE, B-SOURCE-HASH, B-DOCKER-ONLY, B-CI-CONTROLS | Run-1 failure retained; owner waived its acceptance impact. Cleanup receipt, G057 reconciliation and Codex final ACCEPT: PENDING |

**Status.** The run 3b full-profile PASS on `84cbab04` is attributable merged-SHA qualification evidence; the owner has
waived the qualification impact of run 1 while its failure/cause remain recorded above. This is not final Goal
acceptance. M1 and Codex final ACCEPT remain pending, as do the cleanup receipt and G057.
Commands this round (worktree root): `git fetch origin staging`, `git checkout --detach origin/staging`,
`git switch -c g062-s2-qualification-ledger`, `git merge --ff-only origin/staging`, `git log -1 --format='%H %P %T' 84cbab04`,
`git ls-remote origin g062-impl2`, `sha256sum` of the nine artifacts, the check-row `diff` above; all exit 0 except the
`diff` (exit 1, differences found). Resource receipt: none. Not self-accepted; Codex review pending.

**Correction pass (coordinator request, MR !72; this section only).** Commands written in full; the run 1 explanation
now attributes the missing detail to CI-only publication (`infra/ci/validate.sh` 1814-1816, read this pass) rather than
a failure-only policy; the run 3b INFO rows now list all seven; the AC5, AC7, AC9 cells (and AC1, AC2) now name each Q item from
Plan §5.6/§8 (AC table lines 517-526); AC3, AC4, AC6 and AC8 have no Q item there. Artifact table and hashes unchanged.
Evidence directory listings (`ls -la` of both `ci-evidence` folders: `storage.tsv` only) and run 1's console
junit/phase lines were re-read. Then `git fetch origin staging` and `git merge origin/staging`.

### Owner waiver of run-1 qualification impact — 2026-10-21

The owner explicitly waives the acceptance impact of the unexplained run-1 B-CI-CONTROLS failure. The same unchanged
merged SHA `84cbab04` passed the complete local full profile on run 3b: 122/122 required check IDs, every phase exited
0, final status 0 and cleanup verified. Run 1 remains FAILED, its failing control remains uncaptured and its cause
UNKNOWN; the later pass does not explain, erase or retroactively pass it. Run 3b is the qualification basis. This
waiver removes the unknown run-1 disposition as an acceptance blocker; M1, cleanup receipt, G057 and final Codex Goal
ACCEPT remain pending. No additional run is requested by this disposition.

### Owner disposition: M1 preview after G063 — 2026-10-21

The owner directs that G062's manual M1 walkthrough not hold G062 final acceptance. M1 is deferred, not reported as
performed or passed. After G063 receives final ACCEPT, conduct one owner walkthrough against the then-current
integrated staging build, covering both G063's MCP connection management and G062's tenant signup. Any defects found
are recorded as separate follow-up work and do not reopen G062. G057 remains a separate live qualification required
before G021, not a G062 acceptance gate. G062 final acceptance remains subject only to the existing scoped cleanup
receipt and supervisor disposition; this decision does not erase the run-1 failure or its recorded owner waiver.

### Final implementation acceptance — 2026-10-21 — ACCEPT

Codex accepts G062 at merged SHA `84cbab04045e3a19056f5ac2a14eebc5c4b64abc`. The exact merged-SHA local full profile
passed run 3b with 122/122 required checks, all 13 phases successful, final status 0 and cleanup verified. The owner
waived only the acceptance impact of run 1; its FAILED/UNKNOWN disposition remains unchanged. Source review, the MR
fast check, and the corrected qualification ledger are linked in the [final Codex review](G062-final-codex-review-20261021.md).

The cleanup receipt (not included) is imported and verified. It records the removal of 23 task
worktrees and their `g062-*` refs, deletion of `origin/g062-s1-result`, and no remaining G062 containers, volumes or
images. Codex removed the 23 associated `worktree-agent-*` refs after confirming they were not checked out and were
reachable from `origin/main`; their commits remain available. The live-locked `g062-plan`, frozen coordinator
checkout/M1 `.env`, and machine-local evidence remain explicit cleanup holds with the owners and release conditions
listed in the receipt and final review. They do not reopen implementation acceptance. The combined M1 preview is a
separate follow-up after G063 final ACCEPT; G057 remains a separate live qualification required before G021.

This final disposition supersedes the earlier AC-table/status snapshot: M1 is deferred rather than pending for G062,
the run-1 impact is waived without changing its FAILED/UNKNOWN record, and the cleanup receipt is imported. G057's
remaining live-provider proof is a separate G021 prerequisite, not an open G062 acceptance criterion.


### Owner sequencing update: combined staging preview before RC8 — 2026-10-21

The owner supersedes the earlier “after G063” preview timing. The current sequence is G063 final ACCEPT, RC7 qualification, G020, then one combined owner walkthrough on the current integrated staging build before RC8. The walkthrough includes G062 tenant signup and G063 MCP connections and may review the integrated G020 label surfaces. It remains a separate manual follow-up, not a G062 or G063 final-acceptance gate; G063 may be accepted before the preview. Findings are recorded as separate follow-up work and do not reopen accepted Goals. RC8 is cut from the latest staging commit after the preview.


### Owner-confirmed preview/release timing — 2026-10-21

Latest delivery order: G063 → G064 → RC7 → G020 → combined staging preview → RC8 (if needed) → G057 → G021 → RC9. The preview remains a follow-up after G020 acceptance; it does not gate G062 or G063 acceptance. RC8 is optional after the preview; if cut, it requires normal candidate qualification. RC9 follows G021. This supersedes earlier preview/RC8 timing text without changing accepted Goal dispositions or historical evidence.
