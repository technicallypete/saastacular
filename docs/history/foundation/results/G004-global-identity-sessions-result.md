> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G004 — Global identity and sessions: Result

**Status: SUBMITTED FOR CODEX IMPLEMENTATION REVIEW. G004 is not accepted; only Codex may accept it.**

**Final state (2026-10-08; supersedes both status texts below, which are kept as history).** All three gates ran on the same application tree (§ Final evidence on the C7 candidate):
- **Focused manual qualification on the frozen C7 candidate `0f5a5dc`:** MQ-0, MQ-1, MQ-5, MQ-6 and MQ-6B pass. That closes forced reauthentication (AC7) and CAP-4b on both origins (AC6).
- **Final local `all`:** S8d (run 8a9014) passed **77/77**. The run before it, S8c, failed one intermittent, unchanged G048 browser test; both runs are recorded.
- **Consolidated hosted pipeline 2928428140** (owner-started, source `web`, protected `identity-qualification`, `4a46da9`): **success.** `fast` passed 33/33, the full gate passed 77/77, and `identity-staging` passed 5/5 against real WorkOS staging.

**Budget:** C7, the last B7 cycle, was used; **7/7 used and no extension**. The owned local resources are cleaned up. The operator-workstation cleanup is still pending: the browser profiles and the exact-leaf ESET exception (§ Cleanup).

**Earlier state (history; superseded).** The C7 correction (§ C7 correction) is implemented and its covering runs pass. The focused manual retests on the frozen C7 candidate come next, then the final S8 `all` and one consolidated hosted pipeline on that same candidate.

**Earlier status (history).** On the final delivery `e73e47b`:
- **Hosted:** pipeline 2926625444 passed `fast`, the full gate and `identity-staging`.
- **Local:** S8b (run e655ec) passed.
- **Operator session (CAP-4a/MQ):** 7 pass, **2 fail**. MQ-5 failed on an observer timing defect. MQ-6 failed on a real-provider return-contract mismatch, so **CAP-4b is unmet**.

Both failures await Codex direction (§ CAP-4a and MQ). The work in the tree:
- implementation S1–S7;
- the checkpoint-C4 corrections;
- the logging amendment (item 3, resolved by Codex).

The covering run dev6 passed, and **the final local `all` run (S8, run 831a70) passed: 77/77 required IDs**.

The first consolidated hosted pipeline (2925002671, on `0657d88`) passed `fast` and the full gate. Its `identity-staging` job failed closed on a staging-file delivery defect, which is corrected in C6 (§ Hosted correction C6). **The rerun of the final local `all` on the corrected tree (S8b, run e655ec) passed: 77/77.**

Still remaining:
- the operator-attended CAP-4a/4b and MQ-0..7 session;
- the consolidated hosted run (S7 staging with S9 full) on the final tree.

Only Codex may accept G004. This record makes no acceptance claim.

## Bindings

| Input | Value |
|---|---|
| Checkout / branch | `.claude/worktrees/g004-exec`, `g004-global-identity-sessions` |
| BASE | `201f5d1114bd43d978dca6afb26bf410e5f79da7` |
| Plan r5 | `fcbc863d…` (unchanged) |
| E4 snapshot | `approved-inputs.sha256` `31360f9c87ac60bfe68a3767c9d1777e27e75906cdcc9bb6b5988c2fbb499a07`, 42 entries (unchanged; every wrapper run's precheck verifies it) |
| Original binding | `binding.json` `9ee78cd9c3d0e1af70b0a84e3c52ace80af294f3b0769b76d783df09750ef40d` (unchanged; still verified by the launcher) |
| Scope (amended at checkpoint C4) | `scope-allowlist-c4.txt` → `management/v0/results/G004-scope-allowlist.txt`, 87 entries, `a2c019a33c73d6a791d5e42c8f684bd5bc3adbb7719bf7b12b1f0604c93bc749`; metadata `checkpoint-c4-binding.json` `96b26ea86d18686aae555c0acdf352276a6451faf8c5d32a1d8f8b8e1e1c0df9`. Only the source, digest and count are superseded. The copy was verified byte-for-byte after dev2 finished and before any application command used it |
| Checkpoint review | `G004-checkpoint-c4-codex-review.md` `a99a9610dd6e5029f0641c64c123c37d2a3a778c9534add83a857f2ff775998d` (CORRECTION LOOP) |
| Logging amendment | `G004-checkpoint-c4-logging-amendment.md` `fa8b879d3d91e992a7273cf3b1f64634c25d5954088584c4938110a08c504b47` (CORRECTION LOOP; supersedes only the item-3 schema prohibition and the separate pre-final S7 hosted timing) |
| Scope (current) | `scope-allowlist-c4-logging.txt` → `management/v0/results/G004-scope-allowlist.txt`, **88 entries**, `e46b9174e5072e93ad9fb997c2cb600c11deab4cd4dcda2ced4d9a8f38ca8126`; metadata `checkpoint-c4-logging-binding.json` `3471fd24d358ac1f72ac89257d8060c9f51a74b3a4be8e6c79c759cfea386f05`. The only addition is `completion.schema.json`. Copied only after confirming no validation run was active; `cmp` byte-identical. The 87-entry scope above is historical |
| Launcher | `$EV/bin/g004-vw.sh` verifies all three binding digests (original, C4, C4 logging) and takes the scope from the logging amendment. It checks the amendment against the binding (base, snapshot digest and count) before every wrapper run. `--check-only` after the update: precheck passed (42 inputs) |

## Cycle accounting (B7: at most 7 shared meaningful cycles)

| Cycle | Content | State |
|---|---|---|
| C1 | Schema, store, concurrency | done |
| C2 | Adapter and token contract; CAP-2 23/23; CAP-3 PASS | done |
| C3 | Emulator HTTPS E2E environment | done |
| C4 | Web, guard and variant reconciliation (dev1 → three harness defects fixed; dev2) | done; checkpoint C4 reported |
| C5 | Checkpoint-C4 corrections 1–3, disposition tests, CI-1 (adapter gate, staging profile, workflow rule 10, `identity-staging` job, controls), the logging amendment; covering runs dev3–dev6 and the fixes they found; S8 PASSED | done |
| C6 | Hosted correction: the staging-file delivery defect found by pipeline 2925002671; group G controls; S8b; second consolidated hosted run (2926625444, all success); MQ run 2 (7 pass, 2 fail) | done; MQ-5/MQ-6 went to Codex, which answered with the manual-qualification review a015b166 |
| C7 | The addendum correction (cookie-carried sign-out reference, the fixed 307 hand-over, observer freshness and SSO rules); covering runs c7-cover (one fix: the canary registration) and c7-cover2; focused MQ; S8c (one intermittent G048 failure) and S8d; hosted pipeline 2928428140 | **done** |

**Used: 7 of 7 (B7 exhausted). No extension or fresh budget was taken.** The earlier line "5 completed plus C6 in progress; C7 is the remaining reserve" is superseded.

## Checkpoint-C4 correction record (CORRECTION LOOP, pass 1)

### Correction 1: finite identity configuration names in exit diagnostics

**Code.** `core/internal/observability/exit.go`:
- `configKeys` gains exactly the twelve `IDENTITY_*` inputs that `config.LoadIdentity` reads, including `IDENTITY_WORKOS_API_KEY_FILE`.
- Names are an exact map lookup: there is no prefix or wildcard, and unknown names are dropped.
- The comment now names both loaders.

**Catalogue.** The `core.exit` `config_keys` enum in `events.json` gains the same twelve names, sorted and unique.

**Tests.** `exit_test.go`:
- **`TestConfigKeysAllowListMatchesConfigLoad`.** The allow-list must equal the names `config.Load` reads plus the names `config.LoadIdentity` reads on a *valid configured* identity path. The two loaders must not share a name.
- **`TestLogExitIdentityConfigKeys`** (new, 12 subtests). It runs `config.Load` with a valid base configuration and then `config.LoadIdentity`, as `cmd/server` does:
  - **Accepted paths (3):** unconfigured, configured, and configured with a key file in production. Each must read exactly the twelve inputs, against an independent list `identityKeys`.
  - **Rejected paths (9):** partial client-only; partial without a client plus a malformed issuer; a non-workos provider with both key sources; an unreadable key file; an empty key file; a split inline key; origins, lifetimes and algorithms; production refusals; and unconfigured with settings in production.
  - **Each rejection:** one `config_invalid` record whose `config_keys` equals the expected sorted names exactly. Every loader input that the message names is expected. No canary value, file name, directory, `sk_test_` key or message text appears.
  - **Coverage:** together, the rejections name all twelve.
- **`TestLogExitConfigKeysFromMessage`** adds `IDENTITY_`, `IDENTITY_WORKOS`, `IDENTITY_UNKNOWN`, `IDENTITY_WORKOS_API_KEY_FILES` and `XIDENTITY_PROVIDER`. Only the exact `IDENTITY_WORKOS_API_KEY_FILE` is reported.

**Validation.**
- `g004-dev.sh unit ./internal/observability/... ./internal/config/...` (gofmt, vet, test): ok.
- Verbose run: the 12 subtests PASS.

### Correction 2: the five identity events in the independent catalogue expectation

`logrecord_test.go` `TestCatalogueMatchesTheBrief`: the hand-written core list now holds exactly the five Plan §4.7 events (`auth.callback.completed`, `auth.login.started`, `auth.provider.revocation`, `auth.session.revoked`, `identity.namespace_mismatch`) alongside the four existing ones. The list is not derived from `events.json`.

**Accept cases added:** one per identity event, and a `core.exit` record with identity `config_keys`.

**Reject cases added:**
- an uncatalogued identity-like event (`auth.session.created`);
- an identity event on the web service;
- a near-miss `config_keys` name (`IDENTITY_WORKOS_API_KEY_FILES`);
- an unknown identity name.

The existing closed checks are unchanged.

**Validation.** `g004-dev.sh tools` (gofmt, vet, every validation-tools package): ok.

### Correction 3: compose the identity customization onto the disabled variant's own render

The `TOGGLE_RENDERED` exception is replaced in `infra/validation/run.sh` `variants_off`.

**Pinned runs.** Each toggle-rendered file has one contiguous identity run:
- **`.env.example`:** the rendered identity block (`ENV_IDENTITY_RENDERED`), replaced by the reviewed G004 block. That block runs from `ENV_IDENTITY_MARKER` to the end of the file and its SHA is pinned as `ENV_IDENTITY_BLOCK_SHA`.
- **`VENDORS.md`:** the rendered row (`VENDORS_IDENTITY_RENDERED`), replaced by the five reviewed rows. They start with `VENDORS_IDENTITY_PREFIX`, and their SHA is pinned as `VENDORS_IDENTITY_ROWS_SHA`.

**Composition and check.** For each of the two files:
- The root must be its render (root manifest hash) with exactly the reviewed run in place of the rendered one.
- The variant's own render must equal the variant manifest. The delivered variant file is that render with the reviewed run composed in, and is then verified the same way.
- Disabled inference configuration is preserved: the variant's own intelligence-specific render is kept everywhere outside the run.
- Both files must appear in the variant manifest.

**Detector controls.** In the harness on every run:
- the uncomposed render (customization omitted) must be rejected;
- a copy with the run's last line altered must be rejected.

**Focused test before the harness** (`tmp/c3-test.sh`; renders the real disabled variant, then applies `run.sh`'s helpers verbatim):
- root and variant both accepted;
- omitted, altered last line, altered middle line and duplicated run all rejected;
- the variant render differs from the root render for both files, as expected.

shellcheck is clean. Neither file is exempt any more, and no exception was broadened.

### Dispositions (checkpoint items 3, 4, 6–10)

**Item 3: logging scope.** Accepted as directed, but **one conflict is open** (detail below).
- `completion.schema.json` is unchanged and no web completion logging was added.
- `IDENTITY.md` § Logs states:
  - the web identity routes emit no `http.request.completed` record;
  - correlation travels to core's `auth.*` events;
  - these are operational logs, not a durable audit trail.

**Item 4: unconfigured before origin.** Kept for the disabled-provider path only.
- Configured Login and Callback enforce exact origins first.
- An ambiguous callback query still precedes the configured check.
- The unit tests show any origin returns `ErrUnconfigured` with no effects.
- The foreign-origin controls are unchanged.

**Item 6: 307 for anonymous GET `/account` only.** `IDENTITY.md` states this; every other redirect stays 303. The Plan text was not rewritten.

**Item 7: the double gate.** New `TestLoadIdentityCookieProfileGate`, in development and test:
- with core's profile unset or `https`, `http://localhost:3100` and `http://127.0.0.1:3100` are refused;
- an explicit `loopback-http` accepts exactly those, alongside https origins;
- under `loopback-http`, `http://org-a.localhost`, `http://localhost.example`, `http://127.0.0.2`, `http://[::1]` and `http://LOCALHOST` are refused;
- production refuses `loopback-http` with any origin.

Web's production refusal is existing coverage: `session.test.mjs` and B-IDENTITY-CONFIG. `IDENTITY.md` now says that configured HTTPS use outside production needs an explicit `IDENTITY_COOKIE_PROFILE=https` for web, as the overlay sets.

**Item 8: raw Cookie-header parsing.** Unchanged. The tests are kept.

**Item 9: staging floor.**
- `validate.sh --profile staging` checks `ci_time_ok staging` (2700 s) **first**, in the location stage. That is before the hosted-context check, the forwarded-file check, the preflight, the registry and the harness.
- `runner-preflight.sh` and `select-profile.sh` are unchanged.
- **Proof in C-SEC-9:**
  - 2699, `45m` and unset are each refused at location with no harness run;
  - exactly 2700 passes the floor and is then refused for the missing hosted context or forwarded file;
  - the harness staging suite records UNMET for a missing input.
- V37 adds the unit cases staging 2700 ok, 2699 fail and `45m` fail (17/17).

**Item 10: project ID plus a pinned path SHA-256.**
- The adapter compares the literal ID and the literal pinned SHA-256 of `CI_PROJECT_PATH`'s bytes.
- C-SEC-8 refuses, independently: a wrong ID with the approved path; the approved ID with a wrong path; and a missing path.
- The control reads the expected hash from the adapter's literal, never from the incoming path.

### Item 3: resolved by the logging amendment (2026-10-08 UTC)

**Change.** Exactly the five literals were added to the top-level `route` enum and to the core-service `route` enum of `completion.schema.json`:
- `/internal/v1/auth/session`
- `/internal/v1/auth/callback`
- `/internal/v1/auth/login`
- `/internal/v1/auth/logout`
- `/internal/v1/auth/provider-logout/return`

The web and intelligence enums, the closed fields and every conditional rule are unchanged. No wildcard, no middleware change and no new web completion logging.

**Tests.** `logrecord_test.go`:
- **Accept:** each of the five routes on core (GET 200, POST 401, POST 200, POST 503 warn, POST 200). The health cases are kept.
- **Reject** (each `completion-invalid`):
  - an unknown route, the prefix, a trailing slash, the wrong case, the partial `/provider-logout`, and a query-bearing route;
  - an auth route on web or on intelligence;
  - extra `user_id`, `session`, `code`, `email` or `authorization` fields.

The expected values are hand-written. `TestCatalogueMatchesTheBrief` is unchanged.

**Docs.**
- `OPERATIONS.md` (completion records) names the five core identity route patterns.
- `IDENTITY.md` § Logs says core's completion record names only the matched pattern.

**Validation.**
- `g004-dev.sh tools`: every validation-tools package ok; the three logrecord tests PASS.
- **dev6** (below): **PASSED**, including B-CONTRACT-TOOLS, B-UNIT-CHECKS, B-OBS-LOGS, **B-IDENTITY-PRIVACY** and B-SECRET-CANARY.

**Acknowledgment delivery.** The acknowledgment sent to the relaying session expired there unapproved, because of a permission-mode mismatch, and was not delivered. It is recorded in `$EV/CHECKPOINT-C4-LOGGING-ACK.md`.

**Docs correction found while planning S9.** `OPERATIONS.md` validate.sh § Location and § Attribution still described CI as fast/gate only and branch schedules as default-branch only. Both now describe `--profile staging` (the 2700 s floor, then the context and file checks) and the protected qualification schedule.

### Item 3 conflict as reported before the amendment (history)

Core's G003 middleware writes one `http.request.completed` record for every request through the mux. Its `route` is the matched pattern's path, and `correlation.go` is outside G004 scope.
- **The conflict.** The five core identity routes therefore produce records with `route: "/internal/v1/auth/…"`. The accepted `completion.schema.json` (core: `/healthz`, `/readyz`, `unmatched`) rejects these.
- **What fails.** B-IDENTITY-PRIVACY's strict core-log validation fails for this reason alone (see dev3). B-OBS-LOGS keeps its accepted meaning: the unconfigured probe now runs after G003's capture.
- **Options, as in the acknowledgment:**
  - **(a), recommended:** admit the five core paths in both route enums of `completion.schema.json`;
  - **(b):** relabel them "unmatched" by wrapping the mux, which mislabels matched routes;
  - **(c):** change `correlation.go`, which is outside scope.
- **Nothing was changed for this.**

## C7 correction (manual-qualification review a015b166; addendum approval 8485ea5a)

### Bindings

| Item | Value |
|---|---|
| Manual-qualification review | `G004-manual-qualification-codex-review.md` `a015b166944168556ba207228c4e60a3127390cf74de29606f362a4b10702674` (CORRECTION LOOP) |
| Addendum | `management/v0/plans/G004-global-identity-sessions-c7-addendum.md` `841faa15e835d5aec2195d3978e4651f64ed4bbf2f87d1510f03eb51cd0e9f21` (commit `46932db`). Bound by digest before and after each wrapper run (`--expect-hash`, plus the launcher's pre- and post-check) |
| Approval | `G004-c7-addendum-codex-review.md` `8485ea5a34b3f7b844511a3deaaa7fb5144f16bd14f8c95316da0ee376d5708f` (ACCEPT for the bounded correction). It approves the fixed 307 `GET /sign-in` → `GET /auth/logout-return` |
| Scope | `scope-allowlist-c7.txt` → `management/v0/results/G004-scope-allowlist.txt`, **91 entries**, `0001c719ee49c72711c440111be02c0a6af30003adb342a3f76d4111239c846e` (byte-identical copy). New paths: `web/app/auth/logout-return/route.ts`, `core/internal/identity/mq_eval_test.go`, the addendum |
| Metadata | `c7-correction-binding.json` `55dab965520b2c69a5cdd003313a7aba472c53786dc32c74caf180262f1fb3c3`, verified by the launcher with every earlier binding |
| Unchanged | BASE `201f5d1…`; Plan r5 `fcbc863d…`; the 42-input snapshot `31360f9c…` |

**Cycles:** C1–C6 closed; **C7 is the last cycle of B7**. No budget reset.

### What changed

**Core** (`logout.go`, `http.go`):
- The provider sign-out answer carries the one-use reference only in a separate `logout_ref` field.
- `return_to` is the fixed `/sign-in`, and the WorkOS logout `return_to` is exactly `<requesting origin>/sign-in`.
- No reference appears in any URL.
- Unchanged: local-before-provider revocation, the 5 s revocation, redirect only when confirmed, the one-use hash with its 10-minute TTL, and consume-once.

**Web:**
- **`lib/session.ts`:** cookie kind `logout` (`__Host-logout`; `HttpOnly; Secure; Path=/; SameSite=Lax`; no Domain; `Max-Age` at most 600), plus `hasCookie` (presence only).
- **`lib/identity-client.ts`:** validates `logout_ref`, the fixed return, and that the reference is absent from the provider URL.
- **`POST /auth/provider-logout`:** sets the cookie, then answers with a 303:
  - **confirmed:** to the provider;
  - **not confirmed:** directly to `/auth/logout-return`.

  CSRF and origin checks are unchanged.
- **New `GET /auth/logout-return`** (other methods 405):
  - consumes the outcome once through core and renders the closed-map no-store page;
  - clears the cookie in the same response;
  - a missing, malformed or duplicate cookie gives "There is no sign-out outcome to show here.";
  - a spent, forged or expired reference gives "Signed out of this application.";
  - a core outage gives 503 "Sign-out status unavailable";
  - it never changes a session.
- **`/sign-in`:** no longer reads `?logout_ref`. With the logout cookie present it hands over with the approved fixed 307.

**Observer** (`manual_qualification_test.go`; pure rules and their untagged unit tests in `mq_eval_test.go`):
- **MQ-5:**
  - the baseline is the latest signed authentication over all of user A's sessions, re-read before "ready";
  - a sign-in during the wait moves the baseline (counted), within a bounded wait;
  - it records the baseline, bound, ready time and new `auth_time`.
- **MQ-3:** single sign-on reuse is decided only by `mqSSOReuse`; a prompted sign-in fails the step.
- **MQ-6 / MQ-6B:**
  - an actual sign-in on that origin first establishes the session to revoke;
  - it then requires the local and confirmed provider revocation of that session, its reference consumed, and the attested same-origin destination and cleared cookie;
  - finally CAP-4b: a prompted next sign-in with an advanced signed `auth_time`.

**Docs:** `IDENTITY.md` (browser contract, core contract, sign-out section, MQ table). **Harness:** `run.sh`:
- B-WEB-BOUNDARY route and request-context lists gain `/auth/logout-return`;
- `identity_probes` gains the store-failure probe of the return route (PostgreSQL paused: 503, no provider claim).

### SSO evidence linked from run 2 (no-code inspection; approved interpretation)

| Fact | Database record |
|---|---|
| `org-a` session created 17:41:19 | locally revoked at 17:45:30 (`logout`) |
| Next `org-a` session | created 17:45:34, attested `no-prompt` |
| Same user | true |
| Signed `auth_time` | 17:44:05 kept: the latest prior signed authentication (MQ-4) |
| Provider `sid` | the same as the MQ-3 and MQ-4 sessions |

`mqSSOReuse` qualifies exactly this case; it is a unit-test vector.
- **MQ-5's freshness FAIL is kept.**
- **Run 2's MQ-3 is restated as "ordinary sign-in only":** it was prompted, so it is not SSO evidence.

### Validation

1. **Core:**
   - `g004-dev.sh unit` (gofmt, vet, every package): ok;
   - the new HTTP answer test and the `TestMQ*` freshness and SSO rules (positive and negative): pass;
   - `go vet` with tags `integration` and `workos_manual`: ok;
   - `g004-dev.sh it ./internal/identity/...` (real PostgreSQL), including `TestProviderLogout` with the second origin, expiry, consume-once and forged references: ok.
2. **Web:** `g004-dev.sh web` (lint, typecheck, unit): **124/124** (3 new).
3. **Identity E2E** (dev HTTPS environment): focused 6/6, then **56/56** in `identity-light` and `identity-dark`.
4. **Covering run `c7-cover`** (wrapper, run 4b95f8: lint unit root off ops):
   - **Passed:** every check except one, including B-WEB-BOUNDARY (six `/auth` routes), B-IDENTITY-GUARDS, B-UNIT-CHECKS, B-BROWSER-ROOT (141), B-BROWSER-OFF (141), B-TOGGLE-OFF, B-IDENTITY-IT (57), B-IDENTITY-E2E (56), B-IDENTITY-PROBES (with the new store-failure probe), B-IDENTITY-PRIVACY and B-IDENTITY-CONFIG.
   - **Failed: B-SECRET-CANARY.** The new E2E helper registered the deliberately malformed literal `not-a-reference`, which also appears in the repository's tests, as an observed secret. Fixed (`ced10c3`): only canonical references are registered.
   - Postcheck passed; addendum post-check passed.
5. **Rerun `c7-cover2` (ops):** run 3f2fc1, **PASSED** (wrapper 0, postcheck passed, addendum post-check passed). Every ops check passes, including B-IDENTITY-E2E (56) and **B-SECRET-CANARY** (130 secrets absent; the planted control detected).

### Supervisor validation-order amendment (2026-10-08, owner-authorized via Codex)

**The amendment.** The focused manual browser retests run **before** the final S8 `all` and the consolidated hosted fast/full/staging pipeline, on a frozen candidate.

**Rules that still hold:**
- every final local and hosted check runs against that same candidate afterwards;
- any later application, provider or observer change invalidates the affected browser evidence;
- final full CI stays mandatory.

**Not changed:** the Plan, the addendum bytes, the 91-path scope and the budget. This amendment is recorded here only.

### Focused manual qualification on the frozen C7 candidate (2026-10-08 19:55–20:11 UTC)

**Candidate and stack.**
- Candidate: commit `0f5a5dca6dff43a470de7b57c247cdeb80277868`, tree `19a26fc3…`. Its application subtrees and the observer, rules, addendum and scope hashes are recorded in `$EV/mq/candidate-c7.tsv`.
- The MQ stack was updated in place, keeping the per-run CA `d62d3085…` (PEM) and the leaf `cdabae4f…`. Its copy matches the candidate byte-for-byte (712 tracked files).
- The owner started on an explicit "ready". The browser was the same dedicated Firefox profile with the scoped ESET exception.

**Result.** `$EV/mq/result-1791489320.json`: `tree_sha256` `45055795…` (equal to the MQ copy's recomputed hash), `ca_sha256` `d62d3085…`, `config_sha256` `c4f2ee98…`.

| Step | Outcome | Attestation | Server-side observations |
|---|---|---|---|
| MQ-0 | **pass** | — | binding matches the configuration; tree and CA recorded |
| MQ-1 | **pass** | signed-in | session; `mfa_assurance` unknown; recent `auth_time` (19:56:56); provider session recorded |
| MQ-5 | **pass** | prompt | baseline 19:56:56 (latest signed authentication); bound 2m30s; ready 19:59:25; `baseline_moved` 0; new signed `auth_time` 20:01:54, advanced and within the bound |
| MQ-6 (`org-a`) | **pass** | revoked-message, same-origin-return, cookie-absent, prompt | the established `org-a` session (20:02:57) revoked locally and confirmed at WorkOS; return consumed at 20:03:48; next sign-in prompted with an advanced `auth_time` (20:07:39) |
| MQ-6B (`org-b`) | **pass** | revoked-message, same-origin-return, cookie-absent, prompt | the established `org-b` session (20:08:23) revoked locally and confirmed; return consumed at 20:09:03; next sign-in prompted with an advanced `auth_time` |

**What this closes:**
- **R7/AC7 freshness:** forced reauthentication after the maximum age.
- **R6/AC6 and CAP-4b on both origins:** the provider return reaching the same origin, the reference consumed once, the cookie cleared, and the next sign-in requiring credentials.

**Reused unchanged evidence** (approved mapping; login, callback and disable code unchanged in C7):
- run 2's CAP-4a, MQ-2, MQ-4 and MQ-7;
- the SSO-reuse evidence linked above (run 2's 17:45:34 session).

### Final local runs on the C7 candidate (commit `3a0ab1f`: the candidate plus this Result)

| Run | Outcome |
|---|---|
| **s8c-final** (run 3292a8, `all`) | **FAILED: 1 required ID.** Wrapper 1; postcheck passed; addendum post-check passed. **Only B-BROWSER-OFF failed**, through one test: `[shell-dark] [S-FIDELITY]` (`web/tests/browser/design.spec.ts:306`) on the intelligence-disabled variant. Its first two 320 px measurements read the desktop values (`padding-left` 40px against 20; H1 71.68px against 41.6). <br>**Assessment: intermittent, in an accepted and unchanged G048 test.** <br>• The file is unchanged since G048 (`40bfebf`), and G004 never touched it. <br>• Its `resize()` already waits for the viewport and two animation frames. <br>• In the same run it passed on root (light and dark) and on the off variant in light. <br>• It passed on the off variant in S8, S8b, c7-cover and S8d. <br>No change was made. Every other check passed, including B-CI-CONTROLS (224), B-BROWSER-ROOT (141), B-BROWSER-RENAME, every B-IDENTITY-* and B-SECRET-CANARY (180) |
| **s8d-final** (run 8a9014, `all`; unchanged rerun on the same commit) | **PASSED.** Wrapper 0; postcheck passed; addendum post-check passed. <br>• B-REQUIRED-CHECKS: 76 required IDs plus itself, **77/77**; <br>• B-CI-CONTROLS: 224; <br>• B-BROWSER-ROOT, B-BROWSER-OFF and B-BROWSER-RENAME: 141 each for root and off; <br>• every B-IDENTITY-* check; <br>• B-SECRET-CANARY: 180; <br>• B-SCOPE: 88 changed paths within the 91-entry allow-list. <br>Evidence: `$EV/s8d-final/` |

**Formerly pending (both now done):** the consolidated hosted pipeline is in § Final evidence on the C7 candidate, and the cleanup is in § Cleanup.

## Final evidence on the C7 candidate

### One application tree across all three gates

| Gate | Commit | Relation to the frozen candidate `0f5a5dc` |
|---|---|---|
| Focused MQ | `0f5a5dca6dff43a470de7b57c247cdeb80277868` | the candidate itself; the MQ copy's tree hash `45055795…` equals the result file's |
| S8c and S8d | `3a0ab1f` | the candidate plus this Result |
| Hosted pipeline 2928428140 | `4a46da96136b54fd816cfd5dca2cfebd92ae1662` (tree `6669f603…`) | the candidate plus this Result: `git diff --stat 0f5a5dc 4a46da9` lists only this file (36 insertions) |

No application, provider, observer or harness change came after the focused MQ, so the validation-order amendment invalidated none of its browser evidence. Against BASE, 88 paths changed, all within the 91-entry allow-list.

### Consolidated hosted pipeline 2928428140

**Run.** The owner started it in the GitLab UI: <https://gitlab.example/<namespace>/<project>/-/pipelines/2928428140>. Before monitoring I verified it read-only: source `web`, the protected ref `identity-qualification`, SHA `4a46da96…`.
- **Delivery.** `$EV/hosted/delivery.tsv`: `4a46da9` was pushed with `ci.skip` to the evidence branch `g004-global-identity-sessions` and to `identity-qualification`, a fast-forward from `e73e47b`.
- **What I did not do:** launch a duplicate pipeline, restart or cancel any job, or change any setting.
- **Logs.** The job logs are kept privately (0600) in `$EV/hosted/`. The lines quoted below are adapter and validate summary lines, which carry no values.

| Job | Result | Duration | Evidence |
|---|---|---|---|
| `fast` 17045592437 | **success** | 380 s | `profile pipeline=full executed=fast reason=web-explicit`; harness `lint provenance contracts unit`, exit 0; **33/33 checks PASS**; `final status 0; published=full`; disk incremental peak 2.75 GB; guard ok |
| `gate` 17045592438 | **success** | 3372 s | `profile pipeline=full executed=full`; harness `all`, exit 0; **77/77 PASS**, including B-CI-CONTROLS, B-SCOPE and B-REQUIRED-CHECKS; `final status 0; published=full`. Disk: baseline 40.0 GB, minimum 27.4 GB, incremental peak 12.65 GB (the durable 35 GiB floor met); guard ok |
| `identity-staging` 17045592439 | **success** | 142 s | `approved context; 3 of 3 inputs forwarded as one read-only file (values not shown)`; `profile pipeline=staging executed=staging reason=identity-staging-job`; harness `staging`, exit 0; **5/5 PASS**, including B-IDENTITY-STAGING (`TestStagingContract` against the real WorkOS staging API); validate cleanup verified; disk incremental peak 1.56 GB; guard ok |

**Integration profile.** As before, it is not separately selectable for this diff. A hosted gate selects full for any foundation path, and full ⊇ integration (70 ⊂ 77 IDs). This is an interpretation for Codex.

**Staging disk floor** (the provisional 15 GiB; Plan r5 §8.5 rule `max(15, ceil(1.5 × peak + 4))`):
- measured peaks: 1,561,268,224 bytes (2926625444) and 1,561,432,064 bytes (2928428140), about 1.45 GiB each;
- `ceil(1.5 × 1.45 + 4)` = 7, so the rule gives **15 GiB: the floor is unchanged** and `infra/ci/lib.sh` is not edited.

**GitLab settings** (AC11):
- **G1, G2 and G4a:** applied by Codex with the owner's authorization (`G004-gitlab-staging-setup-2026-10-07.md`).
- **G3:** the existing `no_one_allowed` setting was kept and verified.
- **Re-verified at RG-4:** all of them (`$EV/readiness/RG-STATUS.md`).
- **G4b:** the issuer variable was created after CAP-3 PASS from the private issuer file over stdin and read back as metadata only (`$EV/g4b/G4B.md`): three project variables, all protected, environment scope `identity-staging`, raw; only the API key is masked and hidden.
- **G5:** Free tier, recorded.
- **G6:** not used (it would need a paid tier; no purchase).
- **Cross-check:** the staging job forwarding exactly 3 of 3 inputs on the protected web pipeline is consistent with this.
- **Not done:** the settings were not re-read after the run. A read-only re-read is available if Codex wants one.

## Cleanup

**Removed (owned by this Goal):**
- **Docker:**
  - the MQ stack `g004mq-64f1c470` (`g004-mq.sh down`);
  - the dev database `g004dev-pg` and its network;
  - the image `g004dev-web-deps:64f1c470`;
  - the dev identity project (earlier).

  Verified afterwards: no container, volume, network or image named for `g004` or `g001` remains. Unrelated resources (`payments-assistant-*`, `needknow-*`, `outreacher-*`) and every `.env` file were untouched.
- **Derived secret copies:** the MQ stack's generated `manual.env` and `observer.env`.
- **Scratch trees:** `$EV/mq/repo`, `$EV/idn/repo` and `$EV/devdb`.
- **Real WorkOS per-run users:** `TestStagingContract`'s S-CLEAN deletes every per-run user it created and fails if any is still present. Every staging run that passed (2926625444 and 2928428140) therefore confirmed each user gone.
- **Hosted:** each job verified validate's cleanup and removed the registry. The adapter's `finish` and the job's `after_script` remove the staging forwarding volume (the C6 design, asserted by C-SEC-8). The job-private daemon ends with the job.

**Kept on purpose:**
- **Owner files under `platform-secrets/g004/`** (files 0600, directory 0700): `workos-staging.env`, `mq-operator.env`, `mq-ownership.env` and `cap3/`. The owner decides whether to keep them.
- **The persistent MQ users at WorkOS**, whose setup the owner approved. No preexisting or unrelated user was deleted.
- **Evidence under `$EV`:** run logs, MQ result files, hosted job logs (0600), and the run CA and server certificates (public parts only).
- **Branches:** `identity-qualification` stays at `4a46da9`, the qualified tree. The evidence branch carries this Result commit as well (`ci.skip`).

**Pending on the operator's workstation** (cannot be done from here):
1. Close Firefox, then delete the dedicated profile `C:\Temp\g004-mq-ff`.
2. Delete `C:\Temp\g004-mq-profile`, `C:\Temp\g004-mq-profile2` and `C:\Temp\g004-mq-profile3`, where they exist (the Chrome attempts).
3. Remove the ESET exception for the exact test leaf `cdabae4f4bb30ffc8c0ab65396c174c32b445b3ade1c333ebe562b52c2ec666e`.

The per-run CA was trusted only inside that Firefox profile; nothing was added to Windows trust.

## Acceptance criteria: final evidence (Spec r3 AC1–AC12; Plan r5 §9)

All evidence below is on the C7 application tree unless it is marked "run 2". Run 2 evidence is reused under the approved C7 mapping, because login, callback and disable code did not change in C7.

| AC | Evidence |
|---|---|
| AC1 | <ul><li>ADR-003; exact pins (RG-3).</li><li>Unit suites (S8d; hosted `fast`); CAP-3 PASS (C2).</li><li>B-IDENTITY-CONFIG (S8d; hosted full); MQ-1 (focused).</li><li>Real staging token contract, exact issuer and client: B-IDENTITY-STAGING (hosted 2928428140).</li><li>No real key rotation was observed. Rotation is covered by U-JWKS.</li></ul> |
| AC2 | <ul><li>B-IDENTITY-IT, including I-MAP-CONC and I-NS-1..3.</li><li>Forged UUID and header probes (B-IDENTITY-PROBES).</li><li>Real attribute behaviour, S-ATTR (staging).</li><li>Same user across origins: MQ-4 (run 2).</li></ul> |
| AC3 | <ul><li>Real wrong verifier rejected: CAP-4a (run 2).</li><li>MQ-1 (focused).</li><li>E-PKCE and E-CB-1..10 (B-IDENTITY-E2E 56/56); S-CODE (staging).</li></ul> |
| AC4 | <ul><li>I-LIFE and I-REPLICA (B-IDENTITY-IT).</li><li>E-FIX and E-LO-3 (B-IDENTITY-E2E).</li><li>B-IDENTITY-PROBES, including restart, Redis or PostgreSQL loss, and the C7 store-failure probe of `/auth/logout-return`.</li></ul> |
| AC5 | <ul><li>E-LOGIN cookie attributes on two HTTPS hostnames; E-CSRF, E-CT and E-CACHE (B-IDENTITY-E2E).</li><li>The `__Host-logout` attributes: unit tests (`session.test.mjs`). E-LO-6..8 show the cookie is cleared after the return.</li><li>Production `loopback-http` refusal (`TestLoadIdentityCookieProfileGate`; B-IDENTITY-CONFIG).</li><li>No CORS.</li></ul> |
| AC6 | <ul><li>E-LO-1..8, with E-LO-8 now on both origins.</li><li>I-DISABLE and I-ALL.</li><li>**CAP-4b on both origins:** MQ-6 and MQ-6B (focused), each with narrow `sid` revocation confirmed, the same-origin return consumed once, the cookie cleared, and the next sign-in prompted with an advanced `auth_time`.</li><li>MQ-7 (run 2).</li></ul> |
| AC7 | <ul><li>MFA challenge: MQ-1 (focused).</li><li>Cancellation leaves no session: MQ-2 (run 2).</li><li>**Forced reauthentication:** MQ-5 (focused).</li><li>SSO reuse, same UUID: the run-2 linkage under the approved `mqSSOReuse` interpretation. Run 2's MQ-3 is restated as an ordinary sign-in.</li><li>S-MFA (staging); A-1..A-8 (unit).</li></ul> |
| AC8 | <ul><li>Contracts and U-HTTP; E-A11Y; E-NET-1..3.</li><li>Inherited browser scenarios: B-BROWSER-ROOT, -OFF and -RENAME, 141 each (S8d; hosted full).</li><li>One intermittent G048 failure in S8c; passed on rerun (§ Final local runs).</li></ul> |
| AC9 | <ul><li>B-IDENTITY-CATALOG and the reconciled migration checks (S8d; hosted full).</li></ul> |
| AC10 | <ul><li>B-IDENTITY-PRIVACY (after the logging amendment).</li><li>B-SECRET-CANARY: 180 secrets absent, including the staging API key; the planted control detected.</li><li>P-NEG-1..3 and P-POS-1..2.</li></ul> |
| AC11 | <ul><li>Full wrapper `all`: S8d, 77/77.</li><li>Hosted `fast` 33/33 and full 77/77, plus protected `identity-staging` 5/5, on the same tree.</li><li>C-SEC-1..10 and CI-1 admission evidence (B-CI-CONTROLS 224).</li><li>Settings G1–G5 recorded (above); MQ on the same application tree.</li></ul> |
| AC12 | <ul><li>Bindings: snapshot `31360f9c…` (42 inputs); scope `0001c719…`; addendum `841faa15…`. Each verified by the wrapper or launcher before and after every run.</li><li>Cleanup (above, including WorkOS per-run users); CAP results; the cycle-4 checkpoint; this Result.</li><li>**Codex ACCEPT is pending.**</li></ul> |

## Return to Codex

G004 is returned for Codex's implementation review. The three outcomes are ACCEPT, CORRECTION LOOP and ARCHITECTURE DECISION REQUIRED.
- **Budget.** B7 is exhausted (7/7). Any further correction needs a Codex decision on budget, so none is started.
- **Points for review:**
  - the S8c intermittent G048 failure (an unchanged test; passed on rerun);
  - the integration-profile interpretation;
  - the reused run-2 evidence mapping;
  - the operator-workstation cleanup still pending.

## Hosted correction C6: staging-file delivery (2026-10-08)

### Pipeline 2925002671 (first consolidated hosted run)

**Run.** Started by the owner in the GitLab UI. Source `web`, ref `identity-qualification`, SHA `0657d887b8b890dcd68c875ae0f6bf23a1c3dd7c`. Recorded read-only in `$EV/hosted/`.

| Job | Outcome |
|---|---|
| `fast` 17019718394 | **success** (352 s) |
| `gate` 17019718395 | **success** (3393 s): `validate: profile pipeline=full executed=full reason=web-explicit`; harness exit 0; `final status 0; published=full`. Disk: baseline 40.0 GB, minimum 27.4 GB, incremental peak 12.7 GB (floor 35 GiB met); guard ok |
| `identity-staging` 17019718396 | **failed** (`script_failure`, 36 s), as below |

**What happened in `identity-staging`:**
- The adapter accepted the approved context: `identity staging: approved context; 3 of 3 inputs forwarded as one read-only file (values not shown)`. The credential proof was clean and the disk guard ok.
- `validate.sh` then stopped at location: `the staging profile has no forwarded identity staging file`, final status 16, published=none, registry=none. Nothing reached the harness, and no value appears in the log.

**Cause.** The adapter bind-mounted its job-side file (`-v /tmp/hosted-context-staging.*:/run/platform-identity-staging.env:ro`). The Docker CLI reaches the job-private DinD daemon over TCP, so a `-v` source path names the **daemon's** filesystem, not the job container's. Only the build volume and `/certs/client` are shared. The daemon therefore created an empty directory at that path, and `validate.sh`'s `-f` check refused, failing closed as designed.

**Why the controls missed it.** The group G shim runs the context inside the job container itself. It checked the argument shape, but it cannot model daemon-side paths.

### Correction (approved CI-1 files only; every stated security property kept)

**`infra/ci/hosted-context.sh`:**
- **Delivery.** After the context image is built, `stg_deliver` creates the labelled volume `platform-ci-stg-$CI_JOB_ID` on the job daemon. It fills the volume through a helper container's **stdin** (`sh -c 'umask 077 && cat >/d/env && test -s /d/env'`, `--network none`), so no value is ever in argv or env.
- **Mount.** The context mounts the file read-only: `--mount type=volume,src=…,dst=/run/platform-identity-staging.env,volume-subpath=env,readonly`.
- **Failure.** A failed delivery refuses with 16 before any context starts.
- **Cleanup.** `finish` removes the volume, and the staging job's `after_script` removes it as a backstop. No other job's calls change.
- **Unchanged:** the job-side `/tmp` file (mode 600, outside the build directory, removed at exit), the context gate and the credential patterns.

**`infra/ci/validate-controls.sh`:**
- **Shim.** It records the delivered key names (never values) from the helper's stdin and can inject a volume or helper fault. The allowed-call list gains the volume create/rm and the helper run.
- **C-SEC-8 asserts:**
  - the volume-subpath read-only mount and no job-side bind mount;
  - the order: volume created, then filled, then mounted, then removed;
  - exactly the three delivered names;
  - both injected faults → 16, no context, the volume still removed;
  - **the mechanism on the real daemon** (`g_stg_real`), with the adapter's exact helper command and mount form, which C-SEC-8 also checks against the adapter source: a regular mode-600 read-only file of exactly the three names, the key not in the output, and the volume removed. No daemon-host path is involved, so this holds for the remote job-private daemon.

**`docs/foundation/OPERATIONS.md`:** the forwarding bullet describes the volume delivery.

### Validation

1. A throwaway real-daemon proof of the mechanism (`tmp/stg-mech-test.sh`).
2. shellcheck (warning, and SC2251 info): clean.
3. **Group G, all of it:** `validate-controls.sh --group G` gives 32/32 PASS, G1–G7i and C-SEC-1..10 (`$EV/ctl2/`).
4. **S8b:** the final local `all` on the corrected tree `d26063b`, **run e655ec PASSED**. Wrapper 0, postcheck passed, 76 + 1 = 77 required IDs, B-CI-CONTROLS 224, B-SECRET-CANARY (174), and B-SCOPE (85 changed paths within the 88-entry allow-list).

The hosted evidence of the corrected tree comes from a second consolidated web pipeline, which the owner starts.

## Consolidated hosted run on `e73e47b` (history; superseded by pipeline 2928428140)

**Pipeline 2926625444:** started by the owner in the GitLab UI; source `web`; ref `identity-qualification`; SHA `e73e47ba243e41636d9dd677ec24129448f9057a` (the delivery commit). Recorded read-only in `$EV/hosted/`.

| Job | Result | Duration |
|---|---|---|
| `fast` 17031501925 | success | 307 s |
| `gate` 17031501926 | success, full profile | 3561 s |
| `identity-staging` 17031501927 | **success** | 134 s |

The staging job ran the real WorkOS staging contract with the volume-delivered credential file from C6. Job logs are kept privately (0600). October usage after pipeline 2925002671 was 2482 compute minutes.

## CAP-4a and MQ (operator session on `e73e47b`, 2026-10-08 17:30–17:55 UTC)

**Set-up:**
- **Observer:** `TestManualQualification` (tag `workos_manual`).
- **Result file:** `$EV/mq/result-1791480626.json`. Its `tree_sha256` is `73ca9ad9…fe92`, which equals the recomputed hash of the MQ copy of `e73e47b` (709 tracked files byte-identical; the generated `.env` is the only extra). `ca_sha256` `d62d3085…` is the PEM file hash; the CA DER fingerprint is `33a39521…eb74`.
- **Full report:** `$EV/mq/MQ-RUN2-REPORT.md`. Interruptions and the operator's words, verbatim, are in `$EV/mq/attempts.tsv`.

**Browser deviation.** ESET on the operator workstation intercepted TLS and re-signed the test leaf, so Chrome's SPKI-pin launch failed even though the flags were present. The run therefore used a dedicated, disposable Firefox profile that trusts only the per-run CA. That matches Plan §8.6 ("run CA trusted only in a dedicated browser profile"). A scoped ESET exception covered the exact test leaf `cdabae4f…666e`. **Limitation:** the per-run CA has no name constraints, so its trust is local to that profile. CAP-4a attempt 1 was stopped before any attestation (certificate interstitial, then the Network log was not retained).

| Step | Outcome | Attestation | Server-side observations |
|---|---|---|---|
| CAP-4a | **pass** | pasted | wrong verifier → `code_rejected`; positive control `code_valid_with_own_verifier`; control session revoked; no application session |
| MQ-0 | **pass** | — | tree and CA recorded; binding absent |
| MQ-1 | **pass** | signed-in | session; `mfa_assurance` unknown; recent `auth_time`; provider session recorded. Landing `/?status=signed_in` is the reviewed default return |
| MQ-2 | **pass** | abandoned | no cancel control at TOTP; no session or user created |
| MQ-3 | **pass** | prompt | same user; previous session revoked by logout; `auth_time` changed (no SSO-reuse claim) |
| MQ-4 | **pass** | signed-in | same user across origins |
| MQ-5 | **FAIL** | no-prompt | "freshness not shown" |
| MQ-6 | **FAIL** | (none) | "no attestation within five minutes"; MQ-6b not reached |
| MQ-7 | **pass** | refused, signed-in | no session while disabled; all revoked; new session after enable; old stay revoked |

**MQ-5: observer timing defect.**
- **Defect.** `lastAuth` is set only by MQ-1, MQ-5 and MQ-6, not by MQ-3 or MQ-4, so the observer printed "ready" immediately.
- **Database timeline.** The MQ-5 session (17:45:34) reused MQ-4's authentication (17:44:05): 89 s old, inside the 2-minute bound. WorkOS correctly did not prompt and core correctly accepted it.
- **Not proven:** forced reauthentication after the maximum age.

**MQ-6: real-provider return-contract mismatch.**
- **What worked, at 17:48:06** (outcome labels only): local revocation (`provider_logout`); `RevokeSession(sid)` confirmed (`provider_session_revoked`); redirect to the hosted logout (`provider_logout_redirected`); `logout_ref` stored.
- **What broke.** No `provider-logout/return` call reached core, so the hosted logout returned to `/sign-in` without `?logout_ref=…`. The emulator echoes `return_to`, which is why E-LO-8 passed.
- **Consequences.** The outcome message cannot appear with real WorkOS, and **CAP-4b (next sign-in requires credentials) is unmet**. Nothing is inferred.

**Bounded proposal for Codex (nothing started):**
1. **MQ-5 rerun, no code change.** Observer `MQ-0,MQ-1,MQ-5` on the same tree and stack; within C6.
2. **MQ-6 needs a decision:**
   - **(a) Recommended:** carry the logout reference in a short-lived `__Host-` cookie instead of the query (web routes and page, the session module, tests, IDENTITY.md, the B-IDENTITY-GUARDS cookie list). That needs S8 again, one more consolidated hosted run and an MQ-6 rerun, using C7, the last cycle.
   - **(b)** Make the outcome message best-effort (a contract decision).
   - **(c)** A provider setting (not authorized).

**Retained environment (history).** Pending that direction, these stayed up: the MQ stack, the operator's Firefox profile and the exact-leaf ESET exception. They were reused for the C7 focused MQ. The MQ stack is now removed; the two workstation items are in § Cleanup.

## CI-1 (S7) — implemented

- **`infra/ci/hosted-context.sh`.** The staging forwarding gate:
  - **Approved context:** project ID `<gitlab-project-id>` plus the pinned path SHA; branch `identity-qualification` with `CI_COMMIT_REF_PROTECTED=true`; a `web` or `schedule` pipeline without an MR; job and environment `identity-staging`.
  - **Forwarding:** one mode-600 file outside the build directory, mounted read-only at `/run/platform-identity-staging.env`.
  - **Secrets:** only the API key joins the credential patterns.
  - **Refusals:** `--profile staging` anywhere else is refused (16), and any `IDENTITY_STAGING_WORKOS_*` visible to another profile fails closed.
  - **Context:** `CI_COMMIT_REF_PROTECTED` is forwarded.
- **`infra/ci/validate.sh`.**
  - Staging job kind: the floor first, then the hosted context, then the forwarded file.
  - E-1 attribution: a protected qualification schedule counts as a branch schedule.
  - Staging selection; the preflight reuses the fast thresholds; the API key goes to the strict registry.
  - Staging completeness branch; profile regexes.
- **`infra/ci/lib.sh`.** The staging profile: suite `staging`, reserve 300, cap 1800, minimum budget 900, job floor 2700, and 2 CPU / 4 GiB / 15 GiB (provisional). The hosted floor applies to integration and full only.
- **`infra/validation/run.sh`.**
  - The `staging` suite (`suite_staging`, `TestStagingContract`; a missing input is UNMET).
  - B-CI-CONFIG: the ten workflow rules, the `identity-staging` job pin, and the staging profile count.
- **`.gitlab-ci.yml`.**
  - Rule 10: a protected qualification schedule.
  - The `identity-staging` job: stage gate, no needs, 45m (40m+4m), the pinned CLI/DinD and tag, `environment: identity-staging`.
- **`infra/validation/required-checks.tsv`.** `B-IDENTITY-STAGING staging`: 79 rows. Profiles: fast 33, integration 70, full 77, staging 5.
- **`infra/ci/validate-controls.sh`.**
  - New controls: N12q, F-schedule-qualification, F-schedule-other, C-SEC-1..10, and the stub modes for C-SEC-10.
  - Updated: VP1 (33), the VP8 message, the V37 staging floor, and `G_ALLOW` + `CI_COMMIT_REF_PROTECTED`.

**Focused controls** (`bash infra/ci/validate-controls.sh --only VP1,VP8,V37,N12q,F-schedule-qualification,F-schedule-other,C-SEC-1..C-SEC-10`):
**16 passed, 0 failed** (evidence `$EV/ctl1/`). The full matrix runs in the final `controls` suite.

## Validation to date

| Run | Suites | Outcome |
|---|---|---|
| dev1 (wrapper, run e578ee) | lint migrations ops | three harness defects found and fixed (B-WEB-BOUNDARY `cookies()` in a test; `relkind` concat under `set -e`; shellcheck SC2128/SC2318) |
| dev2 (wrapper, run 22cf54) | lint migrations ops | **Lint:** all PASS. **Migrations:** all PASS, including B-IDENTITY-CATALOG 4/4. **Ops:** B-IDENTITY-IT (53) and B-IDENTITY-E2E (56) PASS. **Failures:** B-OBS-LOGS FAIL (the probe ordering, fixed), then an abort on a `chmod` of foreign-owned handoff files (fixed). Postcheck passed |
| ctl1 | 16 focused CI controls | 16/16 PASS |
| dev3-c4 (wrapper, run f9d110) | lint provenance contracts unit off ops | **FAILED: 5 required IDs.** Precheck: 42 inputs, snapshot `31360f9c…`. B-SCOPE passed: 83 changed paths, all within the allow-list. <br>**Passed:** <br>• all of lint, including B-CI-CONFIG (ten rules plus the identity-staging pin), B-ENV-EXAMPLE, B-IDENTITY-GUARDS, B-SHELL and B-MCP-SOURCE; <br>• provenance; <br>• all of contracts, including B-CONTRACT-TOOLS (the new logrecord tests); <br>• B-TOGGLE-OFF-ADDITIONS (composition and its controls); <br>• B-OBS-OFF; <br>• B-BROWSER-OFF (141 tests); <br>• B-OBS-* and B-OBS-LOGS (the ordering fix holds); <br>• B-OUTAGE; <br>• B-PERSIST (version 2); <br>• B-IDENTITY-IT (53); <br>• B-IDENTITY-E2E (56); <br>• B-IDENTITY-CONFIG. <br>**Failures:** <br>• **B-UNIT-CHECKS and B-TOGGLE-OFF.** The README and GETTING-STARTED secret scan (gitleaks `stripe-access-token`) matched two fake `sk_test_…` test literals that I added (`identity_config_test.go:17`, `exit_test.go:250`). Fixed by assembling them as the repo's fake already does. gitleaks then reports 0 findings. <br>• **B-IDENTITY-PROBES.** This was its first execution. The light project's handoff session for Alice was already revoked by later identity-dark tests (sign out everywhere, disable) under `workers: 1`. E-HANDOFF now uses Carol, whom no later test signs out everywhere or disables. <br>• **B-SECRET-CANARY.** `identity_privacy`'s table dump legitimately holds the hostile user's registered display canary and stayed in the work tree. It is now deleted right after the email check. <br>• **B-IDENTITY-PRIVACY.** Only the item-3 conflict: all 445 unclassified core lines are `http.request.completed` records for the five `/internal/v1/auth/*` routes; the web log validates; the scans and P-NEG/P-POS pass. |
| dev4-c4 (wrapper, run 93e63a) | lint unit off ops | **FAILED: 2 required IDs; postcheck passed.** <br>**Now passing:** <br>• B-UNIT-CHECKS (with the secret scan); <br>• B-TOGGLE-OFF-ADDITIONS; <br>• B-TOGGLE-OFF; <br>• B-IDENTITY-IT (53); <br>• B-IDENTITY-E2E (56, Carol handoff); <br>• **B-IDENTITY-PROBES**: live and revoked on core and the replica; restart; Redis stopped; E-LO-5 with PostgreSQL paused; PostgreSQL stopped and recovered; forged headers; CSRF across a restart; the unconfigured stack; <br>• B-IDENTITY-CONFIG. <br>**Failures:** <br>• **B-IDENTITY-PRIVACY:** only item 3 (449 `completion-invalid` lines, all auth routes). <br>• **B-SECRET-CANARY:** a live session's `GET /session` answer carries its CSRF token, a registered value, and the probe outputs stayed in the work tree. The harness now deletes `idn-probes/p*.json` once they are evaluated. **Process note:** that `run.sh` edit landed after the harness printed RESULT but before the wrapper's postcheck finished. `run.sh` is not among the 42 snapshot inputs and the postcheck passed, but it breaks my own no-edits-during-runs rule. |
| **s8b-final (wrapper, run e655ec)** | **all**, on the C6-corrected tree `d26063b` | **PASSED** (wrapper 0, postcheck passed). Same scope as s8-final: 77/77 required IDs; B-CI-CONTROLS 224 (with the new C-SEC-8); B-SECRET-CANARY 174; B-SCOPE 85. Evidence: `$EV/s8b-final/` |
| **s8-final (wrapper, run 831a70)** | **all** (lint controls provenance contracts root migrations ops rename off) | **PASSED** (wrapper 0, postcheck passed; tree `7583044`). <br>• **B-REQUIRED-CHECKS:** 76 required IDs of the executed suites, each recorded exactly once as PASS, plus B-REQUIRED-CHECKS itself: 77, the full profile. <br>• **CI controls:** B-CI-CONTROLS 224 pass, including C-SEC-1..10, N12q and F-schedule-qualification/-other. <br>• **Identity:** B-IDENTITY-GUARDS, -CATALOG, -IT, -E2E, -PROBES, -PRIVACY and -CONFIG. <br>• **Browsers:** B-BROWSER-ROOT and B-BROWSER-OFF, 141 tests each. <br>• **Variant:** B-TOGGLE-OFF-ADDITIONS (composition) and B-TOGGLE-OFF. <br>• **Logs:** B-OBS-*. <br>• **Final checks:** B-SECRET-CANARY (174 secrets absent; the planted control detected); B-SCOPE (85 changed paths within the 88-entry allow-list). <br>Evidence: `$EV/s8-final/` |
| dev6-c4log (wrapper, run 301a12) | lint contracts unit ops | **PASSED** (wrapper 0, postcheck passed). The new 88-entry scope and digest are in force. B-SCOPE: 85 changed paths, all within the allow-list. **B-IDENTITY-PRIVACY PASS:** core, replica and web logs fully catalogued; scans clean; P-NEG/P-POS hold. B-SECRET-CANARY PASS: 140 secrets absent; the planted control detected. Every lint, contracts, unit and ops ID PASS |
| dev5-c4 (wrapper, run a9d5e9) | lint ops | **FAILED: 1 required ID; postcheck passed.** Every lint ID PASS. Every ops ID PASS, including: <br>• B-OBS-LOGS; <br>• B-IDENTITY-IT (53); <br>• B-IDENTITY-E2E (56); <br>• B-IDENTITY-PROBES; <br>• B-IDENTITY-CONFIG; <br>• **B-SECRET-CANARY** (132 generated secrets absent; the planted control detected). <br>**Failure:** **B-IDENTITY-PRIVACY**, for the open item-3 conflict only. |

**Fixes found by the covering runs (C5).** These are all within scope. Each fix's covering suite was rerun.

1. **Fake key literals.** The two key-shaped fake literals in my new tests are now assembled at runtime, as the repo's fake already does. gitleaks reports 0 findings. B-UNIT-CHECKS and B-TOGGLE-OFF pass.
2. **E-HANDOFF user.** E-HANDOFF signs in Carol for the live session. B-IDENTITY-PROBES passes.
3. **Table dump.** `identity_privacy` deletes its identity-table dump after the email check.
4. **Probe outputs.** `identity_probes` deletes its probe outputs after evaluation.

After fixes 3 and 4, B-SECRET-CANARY passes.

**Also verified:**
- web check: lint, typecheck and 121 unit tests;
- core unit: every package, with gofmt and vet;
- shellcheck: every project script.

**Process notes (kept as history):**
- **The dev4 slip:** a `run.sh` edit landed after the harness's RESULT line, during the wrapper's postcheck. The postcheck passed, and `run.sh` is not a snapshot input.
- **dev1–dev5 failures:** all recorded above with their causes.

Pending at that time (history; both are now done, see § Final evidence on the C7 candidate):
- **one consolidated hosted pipeline on the final tree:** a web pipeline on the protected `identity-qualification` branch, giving `fast`, `gate` (full) and `identity-staging` on the same commit. The evidence branch `g004-global-identity-sessions` is pushed with `ci.skip`.
  - **Who starts it.** The owner starts it in the GitLab UI. The API creates an `api`-source pipeline, which the staging job and forwarding gate refuse by design.
  - **Integration profile (interpretation for Codex).** It is not separately selectable for this diff. A hosted gate selects full for any foundation path, and full ⊇ integration; the local wrapper runs cover the integration suites;
- CAP-4a/4b and MQ-0..7 with the operator.

## Files changed (worktree, against BASE)

88 paths changed against BASE, all within the 91-entry allow-list (B-SCOPE). Changed or added:
- **Core:**
  - `core/internal/identity/**`;
  - `core/internal/config/{config.go,identity_config_test.go}`;
  - `core/internal/observability/{exit.go,exit_test.go}`;
  - `core/cmd/server/main.go`;
  - `core/migrations/00002_identity.sql` and its tests;
  - `core/go.{mod,sum}`.
- **Web:**
  - `web/app/{account,auth,sign-in,api/account}`;
  - `web/lib/{session,identity-client}.ts`;
  - `web/components/account`;
  - `web/components/layout/app-shell.tsx`;
  - `web/playwright.config.ts`;
  - the tests.
- **Infra:**
  - `infra/compose/identity.yml`;
  - `infra/identity/**`;
  - `infra/validation/{run.sh,lib.sh,required-checks.tsv,render-customizations.tsv}`;
  - `infra/validation/tools/{cmd/identitysetup,internal/logrecord/{events.json,completion.schema.json,logrecord_test.go}}`;
  - `infra/docker/validation-tools.Dockerfile`;
  - `infra/ci/{hosted-context.sh,validate.sh,lib.sh,validate-controls.sh}`;
  - `.gitlab-ci.yml`.
- **Docs and configuration:**
  - `docs/foundation/{IDENTITY,VENDORS,MIGRATIONS,OPERATIONS,FOLLOW-UP,WEB-SHELL}.md`;
  - `.env.example`.

## Architecture deviations

None beyond the recorded and accepted dispositions (items 4, 6–10), the logging amendment (item 3) and the approved C7 addendum (the cookie-carried sign-out reference and the fixed 307 hand-over).
