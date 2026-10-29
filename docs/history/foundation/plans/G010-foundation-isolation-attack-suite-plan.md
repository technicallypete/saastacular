> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G010 — Foundation isolation attack suite: Implementation Plan

Revision: r1, 2026-10-14 UTC. Status: **review-ready for Codex Plan re-review**. This is not approval, not an execution
binding and not implementation. Implementation waits for final G009 ACCEPT, Codex Plan ACCEPT and a fresh Git-first
execution binding (§0).

r1 is the document-only revision for the r0 Codex review (not included) (CORRECTION LOOP,
SHA-256 `26ff94f1…a08a3a5b`). r0 is archived byte-identically as
G010-foundation-isolation-attack-suite-plan-r0.md (not included) (SHA-256
`487e500cff1c7580a8e587b18e54a4603351df2e87967363e9f4c3fd557463fe`). B5, the ten ACs and every gate are unchanged.

**r1 correction map:**

| Review item | Change | Where |
|---|---|---|
| 1 (A5) | r0 A5 is removed. A typed zero `PutObjectRequest.Group` means an Organization object; it is not invalid input. Every remaining selector case is inherited and cited: zero Group = Organization scope (P01–P07, `putZeroGroup`), zero ObjectID on the four operations (U-OBJ-INPUT), absent versus foreign object (R45/R63), string-selector parsing (IT-CTX-5), and a foreign Group under an Organization (R44). | §2 A |
| 1 (denial rule) | The §1 rule is split. The zero-I/O, unchanged-state rule applies only to pre-provider denials. Staged denials (E1, E3) verify the accepted staged effects and a separate working authorized control; there is no lifecycle redesign. | §1, §2 E, §7 |
| 2 | The §0 table is marked illustrative. A G009 correction requires assessing every relevant context, policy, liveness, RLS/helper, storage, contract and fixture change. | §0 |
| 3 | B3 is cited to IT-PROFILE-1. A4's Sign and Delete faults are cited to R54–R59, P11, D54 and D55; only Put B4/B5/B6 and cross-request Delete (B5) stay new. The deduplication rule is applied now to every step: B2, r0 D1 (policy), r0 D5 (liveness), the C2 scope negatives and E2's signer-revocation leg become citations; r0 D2–D4 merge into one cross-surface step. NEW steps drop from about 21 to 10. The oracle is a flat literal list with explicit per-step test code, not a case language or dispatcher. | §1, §2, §3.2, §3.3 |
| 4 | The standalone b3 is removed. bC also proves the wiring (it must fail for the flipped case only). The forecast is about 60–65 min. | §5, §7, §9 |
| Q1–Q4 | Rulings recorded: minimal B-OBJ-IT wiring; an AUTHORIZATION.md section; inherited expiry, cancellation and poison proof; one local AC8 control. | §3.4, §3.5, §2 D, §7 |

- **Goal / Spec:** [Goal](../goals/G010-foundation-isolation-attack-suite.md) (SHA-256 `1b4940eb…5a710d2`) and
  [Spec](../specs/G010-foundation-isolation-attack-suite-spec.md) (SHA-256 `b07827c3…2d2eb0`), both with the 2026-10-14
  owner amendment; planning authorization (not included) (SHA-256 `bdf88496…a340fd`).
- **Planner:** task G010, role plan, native worktree `.claude/worktrees/agent-<id>` under the
  authorization's single-planner exception; branch `g010-plan`. Effort xhigh as instructed; the runtime setting cannot
  be verified from inside the worker (unknown).
- **Baseline (Git-first):**
  - r0 was written on `origin/staging` `83a5784a35af8b62cf484dc8b20a602ae76b11e5`.
  - r1 is rebased onto `37efdd66e735fef118c52c246b5bc8763c627483`, which adds only the r0 review and a PROGRESS entry.
  - Every non-management byte equals G009 merge `e087f2c`.
  - No snapshot, probe, build, test or Docker command was used for r0 or r1. Read-only inspection confirmed the cited
    inherited cases.
- **Service guidance:** core has no service-specific CLAUDE.md or skill. `docs/SECURITY.md` is empty at baseline
  (an observation; this Plan does not change it).

## 0. Gates and dependency reconciliation

| Gate | State at planning | Effect |
|---|---|---|
| G009 final ACCEPT | Pending. Staging pipeline `2946564278` was `running` on `e087f2c` when checked during r0 | Implementation gate. The Plan targets the integrated code. |
| Codex Plan ACCEPT and execution binding | Pending | Implementation starts only after both. |
| G021 product barrier | Closed | Unchanged. No product route, UI or workflow. |

**G009 reconciliation.** At binding, run the exact-delta assessment of `e087f2c` → the then-current staging. **The
table below is illustrative, not exhaustive.** A G009 correction requires assessing every relevant change, even where
its path is absent from the table:
- context binding;
- policy and storage-action decisions;
- identity and liveness;
- RLS, helpers and grants;
- storage lifecycle;
- contracts and docs;
- fixtures and harness pins.

Unchanged design is not reopened, and exact-delta reuse applies to demonstrably unaffected items.

| Example G009 change | G010 items affected |
|---|---|
| `objects_expected_test.go` IDs or bytes (r2 `e1d26df1…`) | the oracle's `Inherited` citations (§3.2) |
| `obj_it` in `run.sh` (pins 10 / 297 / 85, text) | the textual merge of the §3.4 edit |
| `objEnv`, `setupObjects`, `stageService`, `put`, `singlePool`, `objSpy`, the stage names | the chain glue (§3.3) |
| Put T2 or Delete D4 semantics (for example D4 re-authorization failure → `deletion_pending` with a nil error) | E1 and E3 expectations |
| `OBJECT-STORAGE.md` | the §3.5 doc edit |

G008 is final. Its `rls_expected_test.go` r2 (`4d76e68a…`), pool contract and residuals are relied on unchanged; the
same full assessment applies if any of them changes.

## 1. Implementation approach

G010 adds **one** integrated test, `TestIsolationAttacks`, with chains A–E. It is driven by **one** new independent
oracle that holds only the missing chain expectations. It reuses the accepted `access` integration environment: real
PostgreSQL at version 7 under the attested runtime login, real G004 sessions through the fake WorkOS, and private
MinIO through the scoped credential.

Each chain crosses at least two of: the session/context boundary, the guarded facade, raw runtime SQL and the actual
provider. Across the five chains, all four surfaces are connected.

Reuse rules:
- **Inherited coverage is cited, not duplicated.** Inherited means the G007–G009 single-boundary tests and the two
  accepted oracles. Citations live in the oracle's `Inherited` field and in the doc table (§3.5).
  - G010 adds no discretionary duplicate case or run.
  - Mandatory selected suites (bC, b4, hosted) still execute their existing tests.
- **NEW steps exist only where a composition guarantee is missing (§2).** The deduplication rule was applied in r1
  against the cited sources: an existing case asserting the same actor, context, target and outcome is a citation, not
  a step. If implementation finds another such duplicate, the step becomes a citation. The suite can only shrink.
- **No production code change is planned or allow-listed.** A failing attack that exposes a conformance defect stops
  that step. Claude records the finding and requests a bounded Codex amendment naming the exact path(s) and a
  regression test (§9, E3).
- **One wiring edit:** a second named invocation inside the existing B-OBJ-IT check. There is no new check ID, row,
  control count, framework, registry or snapshot protocol (§3.4).

### Denial evidence rules (AC2–AC8)

**Pre-provider denials** (every step except the staged ones below) count only when all of these hold:
1. The external `access` code equals the oracle's literal code (`not_found`, `forbidden`, `context_denied` or
   `unauthenticated`). `unavailable`, `internal`, `nil` or any other code fails the step.
2. A paired authorized control on the same world succeeded earlier in the chain. That proves the fixture exists and
   the provider is reachable.
3. For storage denials, the `objSpy` call delta is exactly 0, Head presence is unchanged, and the `objectRows` hash is
   unchanged.
4. For write denials, the scoped owner-SQL hash of the touched relations is unchanged.

**Staged denials (E1, E3)** legitimately perform database and provider work before final authorization fails. They
assert the accepted staged effects instead of zero I/O or unchanged state:
- **E1:**
  - the pending row exists, is never published and is never readable;
  - the spy shows the provider Put and Head, then the compensating Delete and the Head that confirms absence;
  - the final code is the oracle's literal;
  - a separate authorized Put through the same staged path completes and is readable (the working control).
- **E3:**
  - D2/D3 removed the bytes and the row stays `deleting`;
  - the result is `deletion_pending` with a nil error;
  - an authorized repeat Delete completes (`deleted`, tombstone).

A missing prerequisite is `t.Fatal`, per the existing `setupIT` convention. `it_count` fails on any SKIP or FAIL. An
infrastructure failure therefore cannot become a denial pass.

## 2. Chains: NEW versus INHERITED

**World.** It is built from the oracle fixture with fresh UUIDs and distinct names. It uses only the existing builders
`e.org/group/profile/link/grantOrg/grantGroup/newUser` (privileged `e.model`, superuser) and `e.put`/`osvc` for objects.
It is not `objWorld`, whose seeded UUIDs are fixed.

The fixture is the complete R1 world, even where a role's behaviour is proven only by citation.

- Organizations: A (flag on) and B.
- In A: g1 and g2 are siblings sharing profile p1; g3 has p2; p3 is unlinked.
- In B: gB, deliberately with g1's slug, and pB.
- Actors:

| Actor | Bindings | Used by |
|---|---|---|
| aOwner | org_owner of A | E1–E3 |
| aAdmin, admin2 | org_admin of A | E1 |
| aMember | org_member of A | AC1 matrix row only; behaviour cited from IT-MATRIX-1 and objects R02/R27/P03 |
| gOwner | org_member of A plus group_owner of g1 | AC1 matrix row only; behaviour cited from IT-PROFILE-1 and objects R23/D24 |
| gAdmin, gMember, gMember2, gMember3 | org_member of A plus that role in g1 | B1, C1, C2, D1, E3 |
| dual | org_member; group_admin of g1 and g2 | D1 |
| multi | org_member and group_admin of g3 in A; org_owner of B (one global identity, different grants) | A1, A2, D2 |
| outsider | valid session, no binding | A2 |

- The Group-only binding is produced dynamically in C2. The static case is cited (IT-CTX-6).
- p2, p3 and pB are cited targets (IT-PROFILE-1/2/4, IT-FACADE-5).
- Objects with exact creators:
  - an Organization object in A (aAdmin);
  - g1 objects (gAdmin, gMember, gMember2, gMember3);
  - a g2 object (dual);
  - an Organization object in B (multi).

  E1 and E2 create their own.
- About 13 real sign-ins. No Cartesian product. The oracle maps every role to its NEW steps or its cited cases (AC1).

Legend: **N** = new composition step (a subtest of `TestIsolationAttacks`). **I** = inherited proof cited by ID.
The codes shown below are planning illustrations from the contracts. The oracle's independently written literals
govern, and the oracle delegate receives this step list with the outcomes withheld (§3.2).

### A — Tenant/subject substitution (AC2)

| Step | Content | N/I |
|---|---|---|
| A1 | One request scope, two tenant contexts with different roles, three surfaces. multi resolves ctxA (A, selected g3) and ctxB (B, Organization-only). Authorized controls: PutObject g3; PutObject in B and SignedGetURL of it. ctxB with A's targets (GetGroup g1 with gB's slug; Sign/Delete of a g1 object; PutObject into g1) and ctxA with pB and B's object: `not_found`, zero provider calls, unchanged state. An attested direct runtime connection (`rlsRuntime`/`rlsAttest`) with both shapes: unfiltered `SELECT id` from `stored_objects` and `groups` returns exactly the oracle's sets for these live rows. Under the ctxB shape, `UPDATE stored_objects … WHERE id = <g1 object>` affects 0 rows. | N: alternating live contexts of one scope across facade, storage and SQL. Single operations: I IT-FLAG-3, IT-FACADE-5, objects R40–R43/R63/P35–P37/D48–D51, `rls:` crossUser, objects RS/RU |
| A2 | Binding faults not yet proven for storage writes: PutObject with ctxA after its scope ended (B4), under a second scope of the same credential (B5) and under outsider's scope (B6); DeleteObject with ctxA under a second same-credential scope (B5). Each gives `context_denied`, 0 BeginTx and 0 provider calls. | N (Put B4/B5/B6, Delete B5 only). I R54–R59 (all six faults × Metadata/Get/Sign), P11 (Put forged), D54 (Delete ended), D55 (Delete another subject), IT-CTX-7/8/10, U-CTX-1..9 |
| — | Selectors: a typed zero Group is an Organization object (authorized for org_owner/org_admin, `forbidden` for others); zero ObjectID; absent versus foreign object; malformed or slug string selectors at Resolve; a foreign Group under an Organization | I P01–P07 with `putZeroGroup`, U-OBJ-INPUT, R45/R63, IT-CTX-5, R44, RLS-CTX, 13 rejected oracle shapes |
| — | Names, slugs, provider and model assertions confer nothing (there is no input path). Optional MFA, unknown assurance and the accepted verified-MFA-required denial. | I IT-CTX-4/5, IT-MFA-1, identity `RequireVerifiedMFA` (`service_test.go`) |

### B — Shared-profile escalation (AC3)

| Step | Content | N/I |
|---|---|---|
| B1 | gAdmin, Organization-only: GetBusinessProfile p1 and UpdateBusinessProfile p1 succeed (the authorized control). The same context then tries the sibling g2: GetGroup and RenameGroup give `not_found`; Sign/Delete of the g2 object and PutObject into g2 give `not_found`. The same actor and shape in raw SQL: `groups`, `group_business_profiles`, `role_bindings` and `stored_objects` return exactly the oracle's sets (sibling rows and link hidden). | N: the shared profile gives no path to the sibling across facade, storage and SQL. Single surfaces: I IT-PROFILE-5, `rls:` "gOwnerA1 org-only: … sibling A2 and its link hidden …", objects R28/R29/P31/D38 |
| — | Selected versus Organization-only, an alternate attached-Group path, both flag values | I IT-PROFILE-1 (either attached Group), IT-PROFILE-2 (selected Group restricts; Organization-only allows), IT-FLAG-4 (false denies only Group-inherited updates; read and selection kept), IT-MATRIX-1 (selected × flag on database-loaded facts) |
| — | Derived management confers no sensitive authority (privileged settings, policy, audit, membership administration; refused SetOrganizationPolicy) | I IT-PROFILE-1 (`integration_test.go:1313`), IT-FLAG-3 |
| — | Unrelated (p2), unlinked (p3) and cross-tenant (pB) profiles; member read versus manage | I IT-PROFILE-1/2/4, IT-FACADE-5 |

### C — Uploader privilege bleed (AC4)

| Step | Content | N/I |
|---|---|---|
| C1 | Live uploads: gMember and gMember2 each PutObject g1, and the stored creator is the authenticated subject. gMember deletes its own: `deleted`, tombstone, Head absent. gMember deletes gMember2's: `forbidden`, spy 0, state equal. gAdmin deletes gMember2's (the authorized control): `deleted`. | N. The oracle's D20–D22/D28/D29 use seeded creators; no test chains a guarded member Put into the own-versus-other Delete decision |
| C2 | gMember3 uploads and keeps a selected-g1 context. `RemoveOrganizationMembership` commits and leaves the g1 binding, so gMember3 is Group-only. With the retained context, DeleteObject and SignedGetURL of its own object give `context_denied`. Raw SQL under that shape returns zero tenant rows, including its own `stored_objects` row; only the actor's own `users` row is visible (D-2). Row and bytes are unchanged, spy 0. | N for storage and SQL (Group-only is absent from the objects oracle). Facade Resolve: I IT-REVOKE-8, IT-CTX-6 |
| — | Scope negatives: sibling or unselected Group put, Organization-scope put by a member, owner/admin deletion of others | I P07/P26/P31, D23–D26 |
| — | Downgrade keeps own-delete and loses others' deletion | I OOR-4 (real ChangeGroupRole), D30–D32 |
| — | Group removal; session liveness | I D33/R36/D35, OOR-1, R50/D36, OOR-5 |

### D — Current facts across reuse (AC5)

| Step | Content | N/I |
|---|---|---|
| D1 | dual keeps one request scope with Organization-only and selected-g1 contexts, its own g1 object and gAdmin's g1 object. Production writers commit before each check: (1) `UnlinkGroupBusinessProfile(g2, p1)`; (2) `ChangeGroupRole(dual, g1 → member)`; (3) `RemoveGroupMembership(dual, g1)`. | N: one retained context observes profile and storage authority together; each surface alone is inherited |
| D1 (1) | Organization-only p1 update is still ok through g1 (an independent path). ObjectMetadata and SignedGetURL of the g2 object still succeed through dual's group_admin role in g2: a link change leaves storage authority unchanged. | |
| D1 (2) | Selected g1: p1 update `forbidden` (last management path gone) and read ok; deleting gAdmin's object `forbidden`; deleting its own object `deleted`. A Group-role change alters both surfaces consistently. | |
| D1 (3) | Selected g1 gives `context_denied` for facade and storage. Organization-only: g1 and p1 `not_found`. | |
| D2 | One observed backend on `e.singlePool` (MaxConns 1), behind one `hookDB` and one `e.objService` service that serves both facade and storage operations. multi selected-g3 PutObject (storage, TENANT-GROUP); a context-free probe (same `pg_backend_pid`, settings `\|\|\|`); multi ctxB GetOrganization(B), exact. An injected storage statement failure (hookDB `failOn`) gives `unavailable` and rollback; the next facade operation on the same PID is exact and `\|\|\|`. | N: storage ↔ facade reuse. Facade only: I P1/P2/P5 |
| — | Policy change on a retained context | I IT-REVOKE-3, IT-FLAG-4, O-11, OR-3 |
| — | Logout, logout-all, disablement and expiry on retained contexts, facade and storage | I IT-REVOKE-4/5/6/7, O-5, OOR-5/6, objects R50/R52/R53/P12/P14/P39/D15/D36/D56 |
| — | Last-path removal with an independent path, member read retained, Organization path | I IT-REVOKE-8, OR-1, O-13, `rls:` last-path |
| — | Cancellation and poisoned initialization (clean reset versus the required destruction/replacement under the accepted pgx v5.11.0 discard contract). No unsafe backend is ever retained to prove reuse. | I P6/P7/P8, I1–I5, E5(f), OBJ-RLS poison, FX-CANCEL, FX-STALL (Q3 ruling) |

### E — Staged storage and bearer residual (AC6)

| Step | Content | N/I |
|---|---|---|
| E1 | Staged denial (§1). aAdmin PutObject of an A Organization object, paused at stage `t2` (bytes landed, V passed). `ChangeOrganizationRole(aAdmin → org_member)` commits, then release. The final code is the oracle's literal; per the contract, a put denial on a readable Organization scope is `forbidden`, as in P03. The row never becomes ready and is never readable; compensation leaves Head absent. Working control: admin2's identical staged Organization Put completes and is readable. | N. OOR-1/OOR-5 cover only Group removal and logout at T2 |
| E2 | aOwner signs a URL for a g1 object. aOwner's DeleteObject pauses at the Delete `provider` stage (after D1): the issued URL still returns 200 with exact bytes (logical delete only), new SignedGetURL is `not_found`, and the row is `deleting`. Release: `deleted`, a tombstone, Head absent, and the URL now returns non-2xx before its expiry. | N. No test covers an issued URL across logical → physical delete |
| E3 | Staged denial (§1). gMember's own Delete pauses at the Delete `t2` stage (bytes removed). `RemoveGroupMembership(gMember, g1)`, release: `deletion_pending` with a nil error; the row stays `deleting`, Head absent. aOwner's repeat Delete completes: `deleted`. | N. Revocation between D1 and D4 is untested |
| — | Issued URL survives signer revocation until expiry; new issuance denied; tampered, expired, anonymous and PUT URLs; real 2 s expiry; former Organization admin | I OOR-7, OBJ-SIGN, OBJ-SIGN-HOST, R05/D05 |

**Size:** 10 NEW steps (A1, A2, B1, C1, C2, D1, D2, E1, E2, E3), plus an attestation subtest and a citation subtest,
under 5 chain parents. That is about 50–60 literal expectations. The exact `it_count` and case count are pinned once
(§3.4).

## 3. Files and modules

| Path | Change | Owner |
|---|---|---|
| `core/internal/tenancy/access/attack_expected_test.go` | **new**, `//go:build integration`. Revision header with input SHA-256 pins and authorship; `atkExpectedFixture()` and `atkExpected()` | oracle delegate |
| `core/internal/tenancy/access/attack_integration_test.go` | **new**, `//go:build integration`. `TestIsolationAttacks` with explicit per-step code | implementer |
| `infra/validation/run.sh` | `obj_it` only (§3.4) | implementer |
| `docs/foundation/AUTHORIZATION.md` | a new section, "Isolation attack suite (G010)" (§3.5) | implementer |
| `docs/foundation/OBJECT-STORAGE.md` | the B-OBJ-IT validation row names the attack suite | implementer |
| `docs/foundation/FOLLOW-UP.md` | annotate the STAGE-RLS "runtime-role attack suite" verify line | implementer |
| `management/v0/results/G010-foundation-isolation-attack-suite-result.md`, `G010-scope-allowlist.txt` | Result and binding allow-list | implementer (initial bytes imported by Codex at binding) |

**Unchanged:** production Go, migrations, `required-checks.tsv` (104 rows), `validate-controls.sh`, `profile-paths.tsv`,
`events.json`, `OPERATIONS.md`, `go.mod`/`go.sum`, and every existing test file, oracle and count pin.

### 3.1 Schema and migrations
None. Version 7 is unchanged.

### 3.2 Oracle (R1, AC1)
- **Shape.** Plain data; no case language, interpreter or dispatcher.
  - `atkExpectedFixture()` lists Organizations, Groups (slug), profiles, links, actors with bindings, and objects with
    creator and scope keys.
  - `atkExpected()` is a flat list of literal expectations. Each records:
    - step ID, operation and target key;
    - the code, or the visible key set for an SQL read;
    - the effects that apply: row state, bytes present, provider calls zero or nonzero, URL 200 / non-2xx;
    - `Inherited` citations;
    - a `Source` contract reference.
  - Each step's test code is written explicitly and looks up its literals by step and operation.
- **Independence.** The author uses canonical contracts only:
  - Spec R1–R3;
  - AUTHORIZATION.md: context, matrix, storage mapping and errors;
  - OBJECT-STORAGE.md: authorization, lifecycle and signed URLs;
  - DATABASE-ISOLATION.md: context validity, trust limit and residuals;
  - ADR-004; ARCHITECTURE §40 storage action policy and §53.7;
  - the two accepted oracles, read-only, solely to cite IDs.

  The author receives §2 with the outcome text withheld: actors, contexts, events, operations, targets and citations
  only. The author has **no access** to `policy.go`, `objects_policy.go`, `objects.go`, `facade.go`, migrations, the
  glue file, observed output or this Plan's illustrative codes.
- **Output and rulings.** The glue prints `ISO-ORACLE {"cases":K,"disagreements":0}` and
  `ISO-ORACLE-DISAGREEMENT <step/op>`. It prints no UUIDs, URLs or payloads. The oracle is never edited to match
  observations: disagreements go to Codex, and originals are preserved (the G008/G009 RU05 precedent).
- **Citation check.** One subtest checks that every `objects:<ID>` / `rls:<Name>` citation resolves in
  `objectsExpectedCases()` / `rlsExpectedCases()`. This is an assertion, not a meta-test.

### 3.3 Glue (`attack_integration_test.go`)
- **Setup.** `setupObjects(t)` provides the shared `itEnv`, the scoped store, `objSpy` and `osvc`. The world is built
  from `atkExpectedFixture()`. An attestation subtest checks the runtime pool and connection with `rlsAttest` and that
  `MINIO_ROOT_*` is absent (the OBJ-ATTEST form).
- **Mechanics, all existing:**
  - context and facade: `begin`, `resolve`, `wantCode`/`CodeOf`;
  - lifecycle writers: the `e.model` writers;
  - state probes: `e.present`, `e.rowState`, `e.objectRows`, `e.state`;
  - staging and pausing: `e.stageService` with `newPause` (E1–E3);
  - pool: `e.singlePool` plus `e.objService` over a `hookDB` (D2);
  - logs: `lastStorage`/`lastReason`.
- **New code.** Only small local `atk*` helpers: an expectation lookup, a backend-PID and settings probe, and a URL GET
  that keeps neither the URL nor the body. This avoids the package-level names `orgA`, `g1`, `begin`, `pause`,
  `eqSet`, …. No existing test file is edited.

### 3.4 Minimum wiring (`run.sh` `obj_it`; escalation E1; Q1 accepted)
- The new test needs `OBJ_ENV` (MinIO), and no check runs the whole package, so the existing B-OBJ-IT host function
  gains a **second** `rls_go` invocation: `'^TestIsolationAttacks$'` with its own log.
  - It runs after `TestObjects`, so `objWorld`'s fixed UUIDs and that log's pins (10 / 297 / 85) are untouched.
  - Pins: `it_count == N`, and the `ISO-ORACLE` line must equal `{"cases":K,"disagreements":0}`. A regex miss
    ("no tests to run") therefore fails.
  - `rls_logcheck` scans both logs, with the canary and planted-canary checks.
  - The PASS/FAIL detail text is extended.
  - About 8 changed lines.
- The ID is unchanged and there is no `required-checks.tsv` row, so the VA7/VU1/J8 and OPERATIONS counts and
  B-TOGGLE-OFF `APPLICABLE` are unchanged. The edit preserves current IDs and fail-closed semantics.
- This minimum edit selects the existing CI controls (§5). No other control suite is added, and those mandatory
  controls are not claimed omitted.
- **Rejected alternatives:**
  - a new B-ISO-ATTACK row;
  - subtests inside an existing test (they break the accepted exact pins);
  - the identity package's auto-run B-IDENTITY-IT (no MinIO or canary environment, and no access to the `access`
    fixtures).

### 3.5 Documentation (AC1, AC7; Q2 accepted)
The new AUTHORIZATION.md section (about 60 lines) contains:
- the actor roster;
- the A–E step table with N/I references;
- both denial evidence rules;
- trust limits and residuals, referenced and not restated as new defects: G008 D1, X1–X7, the bucket-level scoped S3
  credential, and the runtime-credential holder;
- the B-OBJ-IT run.

OBJECT-STORAGE.md and FOLLOW-UP.md get one-line updates. No penetration-test or compliance claim is made.

## 4. Contracts and interfaces
- There are no public API, BFF, MCP, route, capability, event, configuration or schema changes.
- Tests call only the existing internal `access` operations and model writers.

## 5. Test strategy and validation

**Before every batch:**
1. `git fetch origin`.
2. Apply the exact-delta rule, recording both SHAs and the reason, or rebase and reconcile.
3. `bash infra/ci/docker-desktop-mounts.sh`; stop on failure.
4. Take the harness flock.
5. Record the target and candidate SHAs.

Always reconcile before b4 and before the MR.

Standard form: `bash infra/validation/run.sh --base <target> --allow-file management/v0/results/G010-scope-allowlist.txt
--report-dir <EVID>/<batch> <suites>`. While validating, the Result stays at its base bytes (B-PROTECTED; the
G008/G009 precedent).

| Batch | Content | Purpose | Expected wall |
|---|---|---|---|
| b1 focus-go | gofmt, `go vet -tags integration ./internal/tenancy/access/` (the only early compile of the integration files), core unit tests | inner loop | ~1–2 min |
| b2 focus-db | `go test -tags integration -run '^TestIsolationAttacks$'` on the G009 focus stack (pinned PostgreSQL plus the built MinIO; container, volume and network IDs recorded in the run inventory at creation) | inner loop; repeat as needed | ~1–3 min per run (G009 ~35 s for 4 tests) |
| S-review | read-only security review of the diff (§6) | findings fixed; affected b1/b2 rerun | — |
| bC control (AC8; Q4 accepted) | one disposable, never-pushed commit flipping one literal (§7), then `migrations` | B-OBJ-IT must FAIL **only** because of the flipped expectation (`ISO-ORACLE` disagreements 1, naming that step/op); the run exits 1. This also proves that the new invocation executes through `run.sh`. Afterwards, reset to the candidate and verify `HEAD`, a clean tree and that the oracle bytes equal the candidate before b4. | ~5 min |
| b4 final | reconcile, then `all unit` | full profile, which includes every suite of the removed b3. **ci-controls is selected**, because `infra/validation/run.sh` changes (`ci_controls_required`); it runs in the background in parallel. B-REQUIRED-CHECKS: every required ID PASS once | ~47–50 min wall (G009 b18 2788 s; ci-controls ~2550 s in parallel) |

- **No standalone covering batch.** Focused b1/b2, the security review, bC and b4 cover the changes. A targeted
  regression run (for example `migrations`) is added only when a defect or change justifies it, never as a ritual.
- **Run-time forecast.** The new test adds about 15–45 s to B-OBJ-IT: about 11 sign-ins, about 50–60 small operations
  and deterministic pauses, with no sleeps; expiry is inherited. Total local machine time is about 60–65 min without
  failures. A broad run is never repeated per chain, per management-only delta or for a harmless SHA change; only
  failed or affected batches rerun.
- **CI selection (actual paths):**
  - `core/internal/tenancy/` is foundation, so **full**.
  - `infra/validation/run.sh` is full **plus ci-controls**. This is the disclosed, unavoidable cost of minimum wiring:
    no existing check can run a new `access` test without a `run.sh` edit.
  - The allow-list `.txt` is imported into staging at binding, so it is in the base and not a change.
  - `docs/foundation`/management Markdown alone would be fast.
  - No manual or operator stage (no UI, no live provider).
- **Hosted:**
  1. After Codex's pre-integration review: reconcile, push, then `glab` MR to `staging` with source-branch deletion.
  2. MR fast runs automatically; Duo findings are disposed before merge; eligible auto-merge.
  3. The post-merge staging pipeline (full gate plus ci-controls, about 41 min for the gate remainder) is required.
     Success or owner-canceled satisfies it (canceled is recorded as canceled); a failed job fails; pending does not
     qualify.
  4. No optional manual MR gate is requested.
- **Provenance (Git-first, AC9):**
  - The Result records, per batch: candidate and target commits, commands, outcomes (failed, interrupted and skipped
    explicit), core-check, PostgreSQL and MinIO image IDs/digests, the Go version and CI links.
  - Unchanged passing evidence is reused only with exact content and input provenance.
  - Never retained: secrets, sessions, signed URLs/queries, payloads, raw provider errors.
- **Performance:** no new production query. PG-001/002/003 are unchanged, and no E-* collection or benchmark is
  planned. Any material lock wait or scan seen in the chains is reported to the register, not tuned.
- **Cleanup:** bounded. Focus-stack resources go in the run inventory at creation; harness projects are already owned.
  Closeout is one scoped preflight, one removal pass and a compact receipt. Uncertain items are held; native refusals
  are reported once.
- **Safekeeping:** commit and push the implementation branch (for example `g010-foundation-isolation-attack-suite`)
  after every implementation or correction round (no MR). The bC commit is never pushed.

## 6. Delegation

| Role | Owns | Effort | Notes |
|---|---|---|---|
| Implementer (integration owner, single stack owner) | everything in §3 except the oracle | high (owner default) | Worktree `g010-agent-<id>-impl`, or a recorded Codex exception if the tooling cannot name it. Integrates and validates all work, and owns the disposable bC context. |
| Oracle delegate | `attack_expected_test.go` only | high | Inputs and isolation per §3.2. Dispatched first, in parallel with the glue, against the §3.2 shape and the outcome-redacted §2 step list. Its own `g010-agent-<id>-oracle` worktree; the implementer imports the file byte-identically. |
| Security reviewer | none (read-only) | high | Before bC/b4. Reviews both denial evidence rules, the no-overclaim residual text, fixture privilege separation, log/URL/UUID secrecy of new output lines, and that no production behaviour changed. Cannot grant acceptance. |

There are no other workers. Overlapping edits are impossible by ownership.

## 7. Implementation sequence and acceptance mapping

1. **Binding:**
   - after final G009 ACCEPT and Plan ACCEPT: fetch; the full G009 exact-delta assessment (§0);
   - create the implementation worktree and branch from staging; Codex imports the allow-list and initial Result;
   - record identity, path, branch and effort;
   - dispatch the oracle delegate.
2. Glue skeleton, world, attestation, then chains A → E, each with b1/b2 iterations and checkpoint pushes.
3. Import the oracle; reconcile disagreements through Codex; citation check.
4. The `obj_it` edit with exact pins; docs.
5. S-review and fixes (affected b1/b2); bC with verified restoration; reconcile; b4; Result; push; hand back
   review-ready.
6. After Codex pre-integration ACCEPT: MR, fast, Duo, auto-merge, staging disposition. Codex alone marks final ACCEPT
   and closes EPIC-003 (§8).

| AC | Tests | Evidence |
|---|---|---|
| AC1 | `attack_expected_test.go` (every step, `Source`, literal expectations, `Inherited`, authorship header), the §2 actor roster with cited role coverage, citation subtest, AUTHORIZATION.md table | `ISO-ORACLE` K/0; doc review |
| AC2 | A1, A2, plus the cited selector/substitution cases | authorized controls, exact SQL sets, spy 0, unchanged hashes, 0 BeginTx |
| AC3 | B1, plus cited IT-PROFILE-1/2/5, IT-FLAG-3/4, IT-MATRIX-1 | exact codes and sets across facade, storage and SQL |
| AC4 | C1, C2, plus cited OOR-4, D23–D33 | DB/MinIO before/after, creator binding, admin control |
| AC5 | D1, D2, plus cited IT-REVOKE-3..8, O-5, OOR-5/6, P6/P7/P8, I1–I5, FX-CANCEL | committed writers, cross-surface outcomes, same PID and `\|\|\|`, exact results |
| AC6 | E1–E3, plus cited OOR-7/OBJ-SIGN | staged effects (§1): never ready, compensated absence, working control; URL 200 → non-2xx across logical → physical; `deletion_pending` and repair |
| AC7 | attestation subtest, `e.model` superuser setup, spy counts, unchanged B-RLS-CATALOG/B-OBJ-CATALOG/U-ACCESS-BOUNDARY in b4, `rls_logcheck` canaries on the new log, §3.5 trust-limit text | b4 PASS, docs |
| AC8 | **bC: exactly one** assertion reversal (C1's "gMember deletes gMember2's object: `forbidden`" flipped to `deleted`) through the existing B-OBJ-IT gate; the §1 rules (no SKIP, exact codes, authorized controls, staged-effect checks) | bC report: B-OBJ-IT FAIL for that expectation only, exit 1, verified restoration; diff review shows no schema, contract or framework |
| AC9 | b1/b2, bC and b4 records, reconciliations, hosted links | Result |
| AC10 | Result (B5/checkpoint-3 ledger, cleanup receipt), staging disposition, Codex final ACCEPT, EPIC-003 compaction disposition, handoffs | Codex review |

## 8. Closeout
- **G010 completes EPIC-003** (G007–G010). Its final ACCEPT triggers the mandatory epic-closeout PROGRESS compaction
  checkpoint, owned by Codex: compact into `management/v0/archive/`, or record no-change, or a specific active-binding
  deferral with the next safe checkpoint. It also resolves the G008-deferred movement (PROGRESS 2026-10-14 entry).
- Claude does not edit PROGRESS. The Result states that the disposition is owed. Compaction triggers no test run.
- Bulky local evidence is purged under the closeout policy after final ACCEPT, keeping the compact record.
- **Handoffs:**
  - G011: map `access.decision`/`storage.operation` to the ADR-006 audit schema; the attack world can verify audit
    read isolation.
  - G013: tombstones and pending rows from E1/E3.
  - G014/G015: never trust adapter- or AI-supplied scope; PG-001 admission.
  - G018/G019: membership administration.
  - G021 stays closed.

## 9. Escalation items, risks and rulings

| ID | Item | Class |
|---|---|---|
| E1 | `run.sh` `obj_it` second invocation and pins. Selects the existing ci-controls (~42 min local in the background, plus the hosted gate). No new ID, row or control count. | bounded CI change (Q1 accepted) |
| E2 | New test-only world, oracle and glue in package `access`; reuses existing fixtures; no framework, case language or dispatcher | test scope |
| E3 | Conformance repairs: **none planned, no production path allow-listed**. A defect stops the step and goes to Codex for a bounded amendment (exact paths plus regression test). Changed capability/security semantics, schema, public contract, identity redesign or infrastructure → ARCHITECTURE DECISION REQUIRED. | scope/security |
| E4 | None of: schema, migration, contract, route, MCP, capability, dependency, vendor or infrastructure | — |
| E5 | The oracle delegate reads the two accepted oracles to cite IDs. They are independent artifacts, not implementation predicates. | independence |

**Risks:**
- R-a: fixture names colliding in the shared migrations database. Mitigation: distinct `Attack` names plus the canary
  suffix.
- R-b: the exact pins must be set once at stabilization; any later subtest change updates them in the same commit.
- R-c: oracle disagreements need Codex rulings, which may cost an attempt.
- R-d: a real conformance defect (E3).
- R-e: hosted ci-controls drift (as in G009).
- R-f: a G009 correction (§0, full assessment).
- R-g: the E1–E3 pauses rely on the existing stage seams. If a seam cannot express a step, report the smallest
  alternative before adding any mechanism.
- R-h: without a standalone covering batch, a wiring defect first appears in bC. It is fixed and only bC is rerun, a
  ~5 min cost.

**Budget:** B5, with a checkpoint after attempt 3. Planning spends none. **Forecast 2 likely, 3 possible:**
- attempt 1: implementation through b4, review-ready;
- attempt 2: one bounded Codex correction;
- attempt 3: only for hosted drift, a G009 reconciliation or an oracle-ruling cycle;
- attempts 4–5: only if E3 needs an authorized repair.

**Context:** one bounded test increment; no split recommended.

**Allow-list content for binding** (`G010-scope-allowlist.txt`, one ERE per line):
```
^core/internal/tenancy/access/attack_(expected|integration)_test\.go$
^infra/validation/run\.sh$
^docs/foundation/(AUTHORIZATION|OBJECT-STORAGE|FOLLOW-UP)\.md$
^management/v0/results/G010-(scope-allowlist\.txt|foundation-isolation-attack-suite-result\.md)$
```

**Rulings (r0 review):**
- Q1: minimal second invocation inside B-OBJ-IT, no new check ID or count campaign.
- Q2: extend AUTHORIZATION.md and the existing storage/follow-up docs.
- Q3: cited inherited expiry, cancellation and poisoned-pool proof plus the new composition coverage, preserving the
  clean-reuse versus required-discard distinction.
- Q4: one local-only, never-pushed assertion reversal through the existing gate, with candidate bytes restored and
  verified before b4.

No open questions in r1.

## Revision history

- r0, 2026-10-14 UTC: initial Plan on staging `83a5784` (G009 merged at `e087f2c`, final ACCEPT pending). Nothing was
  executed. Exact archive (not included); reviewed commit `a87ccab`.
- r1, 2026-10-14 UTC: document-only revision for the r0 review (not included) CORRECTION
  LOOP, rebased onto staging `37efdd6`. Four corrections (map at top) and the Q1–Q4 rulings. NEW steps reduced from
  about 21 to 10, and the standalone b3 removed. There is no new Spec, scope, framework, probe or test run. B5, the ten
  ACs, G021 and every gate are unchanged.
