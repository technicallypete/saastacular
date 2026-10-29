> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G018 Implementation Plan — Group creation and create-or-select Business Profile onboarding

Revision **r5** (provisional; one-paragraph fixture correction of r4 under the Codex review
`management/v0/results/G018-plan-codex-review-r4.md`, SHA-256 `2f741b02d433654e8012acfcd4edd8f0675a724cfb1f27b9f03c134e7afc6532`;
same branch `g018-plan-r4`, r4 commit `898c7efd` + `origin/staging` `091be3d6` (the r4 review; management-only)
merged; r4 archived byte-exact as `…-plan-r4.md` (SHA-256
`55bc7821a7b723c28fd47d7ee2579ddf66b2ee9affc035162e1bd046ff00b583`); scope allow-list unchanged; same worker,
worktree `agent-<id>`). r4 was the bounded correction of r3 under the Codex review
`management/v0/results/G018-plan-codex-review-r3.md`, SHA-256 `5567da3ca2db4242298fd3b8bce6fe4feefec90ea72fb94616a879d4ad5b2e86`;
branch `g018-plan-r4`, created at the r3 commit `bab9d32070ba52384abc24933abb5f9a7a6e0675` (`origin/g018-plan-r2`) with
`origin/staging` `1131651c` (the r3 review; management-only) merged; r3 archived byte-exact as `…-plan-r3.md`
(SHA-256 `963a97cabbe86de16017430cfedbf9fd12a19c53c1d834fcd860d4aa0d348c3e`); the scope allow-list is **byte-identical
to r3** (SHA-256 `5a94f3ac6cc585bfd437e3b5cf7b577bec147ca0c2d3d816d9e8e8d267a84a7e`; r4 adds no pattern, §11.3, so no
`-r3` copy is made and its header comment still reads "r3"). Worktree for r4: generated name `agent-<id>`
(recorded per AGENTS "Task-prefixed Claude worktrees"; no rename); author Claude (model Fable 5.1; planning effort
"high (requested)", actual value **unknown**). r3 was the bounded correction of r2 under the Codex review
`management/v0/results/G018-plan-codex-review-r2.md`, SHA-256 `423f8107a980f65c1a1b04e0121b9bc57b8bc54facb8dbd6b6ef0a72a4926ebc`;
same branch `g018-plan-r2`, r2 commit `087b2e4c` + `origin/staging` `4eef5983` (management-only) merged; r2 archived
byte-exact as `…-plan-r2.md` (SHA-256 `120625d108637ddc0766da0bceb234400523b96be91e722e483ba2d4bac030c4`) and
`G018-scope-allowlist-r2.txt` (SHA-256 `70eff624fccd0d9159297e451b4c44abdfe9503656daffc7ad22d8aadcfbb22b`). r2 was
the bounded correction of r1 under the Codex review
`management/v0/results/G018-plan-codex-review-r1.md`, SHA-256 `ec3855ed39cca99c7e10e49209781ca260e42b6763dc6c411f2b244f835f3b45`,
reconciled in the same revision with the owner's **superseding D-1** recorded in the Spec at staging `efd4295c`
(Spec SHA-256 `784feedbeb79d975b95a51f382f107e1bd6be380ba9a255abc3f977f6d52e653`, section "Superseding D-1")),
2026-10-17 UTC. Author of r0–r3: Claude (model Fable 5.1; planning effort "high (requested)" — the runtime setting is not
verifiable from inside the worker, actual value **unknown**). Worktree for r2/r3: generated name `agent-<id>`
(recorded per AGENTS "Task-prefixed Claude worktrees"; no rename). Branch `g018-plan-r2`, created at the r1 commit
`0eed3ad567d550bbe03a9b0054c79a4077a15cf0` with `origin/staging` `68cb0d1e` (the r1 review) and then `efd4295c`
(the superseding D-1), both management-only, merged. r1 is archived byte-exact as
`management/v0/plans/G018-group-business-profile-onboarding-plan-r1.md` (SHA-256
`e9ac8c83c0780abca4b99a154276238c0514ae5dd3a4a5148ae7d6ef042e05eb`) and
`management/v0/plans/G018-scope-allowlist-r1.txt` (SHA-256
`ef4ce188f7c89bf29c3cb33de4ebb4c292748aec6be55ab83f91c65119baae3c`); r0 stays archived as `…-plan-r0.md`
(SHA-256 `c7a64649f35692e3b985d292026cc4e21649d293ad8e9ec0d86eb98718574c59`) and `G018-scope-allowlist-r0.txt`
(SHA-256 `8e40e9cb5b84a200c07d4284ff536db923e0c5980f710d6953c252d8137013e0`). Status: **proposed for Codex review;
provisional on unmerged G017 (now observed at `e9d35f3f`, still in progress); not an execution binding; no
implementation, build, test, probe, container or MR.** Tier: architecture/security/data (Goal). Budget **B7,
checkpoint after attempt 4**, unspent; planning spends no attempt. Forecast: 3 attempts.

**Provisional note.** G017 at `e9d35f3f` (the observed MR !40 head) adds, relative to `862db18d`, the web branding
client (`web/lib/branding.ts`), `tenant-frame`, `brand-mark`, `context-indicator`, a `branding/logo` BFF route, shell
changes and ten harness lines; `ShellNavItem.href` is still `` `#${string}` ``, no shared session-scope package exists,
and `required-checks.tsv` still holds 119 IDs (G017 later slices are expected to reach 121). Every pin in §11.2 is
read at `862db18d`/`e9d35f3f` and must be re-read at R0 (§0.3) against the integrated G017. Nothing below assumes
G017 web code.

**r0 → r1 correction record (review §"Required bounded corrections" 1–6 and dispositions).** C1 creator bootstrap →
§1.1, §4.3–4.4, §5.6, ONB-2/ONB-3b/CT-ONB-1 (the subject-filtered count reader is removed; D-9 redesigned). C2
database-enforced profile eligibility → §4.4 (`app.rls_profile_eligible`), ONB-9, RLS oracle r3. C3 executable
INSERT/return sequences under actual grants → §4.7, §5.5 corrected. C4 combined-facts lock order → §1.4
"Lock order", O-ONB-D. C5 real inventory, selectors and focused validation → §3, §6.1–6.3, §11.2 reconciled counts.
C6 bounded lists and honest recovery → §1.5, §1.4 retry table, §1.6, §5.3. D-3 accepted in principle with positive
security and closed-route coverage (§6.2 ONB-H, §5.3). D-5 pinned copy retained (§1.1). `can_update` display hint
only (§5.3). **D-1 resolved by the owner on 2026-10-17 (staging `2d535708`; Spec §"Owner decision D-1"): `group.create`
is held by current `org_owner`/`org_admin` membership in the active Organization only — option (a). The (b)
keyed-insert alternative is dropped; creator-owner atomicity, no Organization grant and no Group-role-only creation
are preserved. This resolution is not Plan acceptance or execution release.** *(r2: that answer is superseded by the
owner's later same-day decision below; the paragraph is kept as history.)*

**r1 → r2 correction record (review r1 §"Remaining bounded corrections" 1–4, plus the superseding D-1).** C2′ `xmin`
is not creation proof → §1.1, §1.4, §4.2/§4.4/§4.7 (the admission proof is now "inserted by this transaction": `xmin`
**and** `created_at = transaction_timestamp()`, with `created_at` pinned non-updatable at runtime), ONB-9, CT-ONB-1
and the single deterministic ordering negative **O-ONB-E**. Reader context → §4.4 (the boolean reader validates the
current Organization/selected-Group context itself through `app.rls_org_role()`), ONB-9e direct-call negatives and
attack row C3. Request budgets → §1.1, §1.6, §5.3 (computed whole-payload budgets: 96 KiB for the two profile-bearing
JSON routes and the two profile forms, 8 KiB/16 KiB elsewhere; no field limit shrinks), ONB-H9, the `session.ts`
`maxBytes` option (scope +2 patterns). Executable validation → §6.3 (tools-module working directory and package path,
`TestRLSFootprint` no longer cited as an oracle r3 target, pinned affected subcases, the bounded wider-batch rule).
**Superseding D-1 (owner, Spec `efd4295c`):** `group.create` is held by every current Organization membership role
(`org_owner`/`org_admin`/`org_member`) in Organization-only context; the earlier same-day owner/admin-only answer
(`2d535708`) is **superseded, kept as history** in §0.1/§10/§13. Applied to §1.1, §1.3, §1.4, §4.2/§4.4/§4.7 (the
member-safe keyed Group insert without RETURNING followed by the read after the bootstrap trigger), §5.3, §5.6,
§6.2 (ONB-1/2/13, E-ONB-4), §9, §10, §11.2. Creator-owner atomicity, no Organization grant, no other-Group authority,
Group-role-only creation refused, pre-existing eligibility for existing-profile selection, and the settled bootstrap,
profile-insert, lock and paging designs are unchanged. No public endpoint, implicit membership or staff privilege is
added (G042/G043). Planning spends no attempt; B7/checkpoint 4 unchanged.

**r2 → r3 correction record (review r2 §"Two bounded corrections" 1–2; everything else closed and untouched).**
(1) Timestamp equality is not insertion provenance: the r2 `xmin` + `created_at = transaction_timestamp()` predicate
is **withdrawn** and replaced by an insertion-only provenance column `business_profiles.created_xid xid8 NOT NULL
DEFAULT pg_current_xact_id()` — assigned once by the column default at INSERT, outside every runtime column grant,
immutable by the extended `business_profiles_keys_immutable` trigger, never repeating (64-bit epoch-qualified xid)
and never current again once committed; reader branch (a) is `bp.created_xid = pg_current_xact_id()`. No `xmin` and no
timestamp is read anywhere. O-ONB-E retained and extended with the synthetic equal-timestamp negative (c); CT-ONB-1
pins the column; scope +1 (`core/internal/migrate/business_profile_integration_test.go`, column/privilege inventory).
§1.1, §1.4, §4.1/§4.2/§4.4/§4.5/§4.6, §4 risk, §6.2, §10, §11.2, §11.3, §12. (2) O-ONB-D's fixture is made
executable: two **incomplete** targets g1/g2 plus a separate qualifying attached Group g3 (profile p3), both targets
in both actors' candidate sets; lock design and interleavings unchanged. Member-safe insert/read, reader context,
body budgets and selectors are closed and not reopened.

**r3 → r4 correction record (review r3 §"One bounded correction"; everything else closed and untouched).** The
accepted recovery path (`infra/recovery/recovery.sh:369` custom-format `pg_dump`, `:556/:562` target system
identifier must differ from the source, `:614–617` single-transaction `pg_restore` as the superuser, `:621–625`
source/target data+catalog fingerprint equality) copies a committed `created_xid` into a **different cluster's**
transaction-id namespace, where it can equal a later `pg_current_xact_id()`. r4 pins **option (a), cluster-aware
provenance**: a second insertion-only column `business_profiles.created_sysid bigint NOT NULL DEFAULT
(pg_control_system()).system_identifier` (the cluster's 64-bit system identifier, the value the recovery path already
requires to differ between source and target), assigned by its default at INSERT, outside every runtime grant,
immutable by the same replaced `business_profiles_keys_immutable` body, and reader branch (a) becomes
`(bp.created_xid, bp.created_sysid) = (pg_current_xact_id(), (pg_control_system()).system_identifier)`. A restored row
keeps its source pair, so in the target it never equals the target's current pair; nothing is rewritten at restore,
so the fingerprint equality `:621–625` holds unchanged and no recovery script, SQL, grant or integrity check is
touched. Option (b) (rewriting restored `created_xid` during restore) is rejected: it would either break the data
fingerprint diff or require excluding the column from it (a weakened integrity check), add a trusted mutation to
`recovery.sh`, and leave a manual `pg_restore` outside the script unprotected. Applied to §1.1, §1.4, §4.1/§4.2/§4.4/
§4.5/§4.6, §4 evidence paragraph, §6.2 (CT-ONB-1 extended; new deterministic restore-collision negative **R-ONB-1**
and its fresh-insert positive), §6.3 (no drill rerun; the hosted migrations suite's existing B-RECOVERY-DRILL is the
real-restore evidence), §10, §11.2, §11.3 (**no new scope pattern**; the carrying patterns are listed), §12, §13.
No `xmin`, no timestamp, no new function, no SECURITY DEFINER change (allowance stays 1). Planning spends no attempt;
B7/checkpoint 4 unchanged; G019/G058 stops unchanged.

**r4 → r5 correction record (review r4: one fixture paragraph only; the design is accepted and untouched).**
R-ONB-1's negative inspected the synthetic restored row as the FORCE-bound owner under u's ordinary-member context
with only an incomplete gN, where the unchanged `business_profiles_select` policy hides the unattached p_r (the
collision SELECT returned no row and the UPDATE affected zero rows). r5 applies the review's option (a): the
collision/provenance inspection and the `xmin`-refreshing UPDATE run as the package's separately attested superuser
fixture connection (the real restore actor, RLS not applicable), then `SET LOCAL ROLE <db>_runtime` with u's
transaction-local context performs the reader and association denial in the **same** transaction, so the colliding
id is live for the reader; the positive is unchanged except that its post-commit provenance read uses the
privileged connection. SELECT policy, role boundaries and grants are untouched; no schema, scope or selector change
(§6.3's `^TestOnboardingCatalog$/^R-ONB-1` selector still applies). Only §6.2's R-ONB-1 row, this record, the
revision line and §13 change.

## 0. Gates, provisional G017 interfaces and R0

### 0.1 Gates

| Gate | Status | Owner |
|---|---|---|
| G017 final ACCEPT, hosted qualification and integration into `staging` (MR !40, observed head `e9d35f3f`) | **open** | Codex / owner |
| **D-1** `group.create` role mapping (§10) | **RESOLVED — superseded (r2)**: the owner's first answer on 2026-10-17 (owners and admins only, `2d535708`, Spec SHA-256 `6d86d694…6a09`) is superseded by the owner's later same-day decision (Spec §"Superseding D-1", staging `efd4295c`, SHA-256 `784feedb…2e653`): every current Organization membership role (`org_owner`/`org_admin`/`org_member`) in Organization-only context; the verified creator becomes the initial `group_owner` atomically; no Organization-role grant, actor selector, extra owner store or other-Group privilege; a Group role without current Organization membership stays insufficient; existing-profile selection keeps needing pre-existing eligibility. Applied in §1.3, §1.4, §4.2/§4.4/§4.7, §5.6, §6.2; resolves D-1 only; no public endpoint or implicit membership (G042/G043) | owner (done) |
| **D-3** retiring the four absence probes for the deliberately introduced routes | accepted in principle, conditional on the positive coverage in §5.3/§6.2 (ONB-H) and preserving every other probe | Codex (review of r1) |
| **D-9** | rejected as written in r0; **redesigned** (§4.3–4.4): no count reader, no SECURITY DEFINER mutation surface, no broad grant | Codex (review of r1) |
| R0 reconciliation (§0.3) after G017 integration; re-pin §11.2; rebind §2 | after integration | Claude; Codex reviews |
| Codex Plan ACCEPT; allow-list import (`management/v0/results/G018-scope-allowlist.txt`); initial Result; Git-first binding; B7 ledger; resource readiness; recorded execution release | after the above | Codex |
| Owner check-in: "call for me when G018 planning is finished" — the Plan accepted and §5.6 interfaces settled, **before** G019 Spec reassessment | after Plan ACCEPT | Codex → owner |
| G021 barrier closed; no G019 planning/implementation; no G020+ | — | — |

### 0.2 Provisional G017 and accepted G016 interfaces relied on (P-1..10, file:line at `862db18d`; `e9d35f3f` deltas noted)

| ID | Interface | Where |
|---|---|---|
| P-1 | session-scope behaviour `begin`/`scope`/`originHost`/`failure` (one `Authorization: Session` header, exact `https://` `origin`, resolver lookup, G016 Organization binding, CSRF for mutations, `BeginRequest` then Organization-only `Resolve`), copied byte-pinned from `domains/api` and pinned by `TestDomainsCopyIsExact` | `core/internal/branding/api/http.go:282-351`, `:115-138` (codes), `:151` (`refuse`), `:467` (`mutation`); `http_test.go:557` |
| P-2 | thirteen capabilities, `capOrganizationManageBranding` last | `core/internal/tenancy/access/policy.go:12-26`; `matrix_test.go:731`, `integration_test.go:1469` (286 cells) |
| P-3 | audit catalog: 36 actions (30 trigger, 6 core), 58 reasons, 29 `access.denied` operations; `TestCatalogParity` reads `00013` | `core/internal/audit/audit.go:65,168,190`; `audit_test.go:69-95` |
| P-4 | migration set `00001`–`00013`; version-13 suffix pins | `core/migrations/migrations_test.go:50-54` (`TestMigrationSet`); `core/internal/migrate/runner_test.go:118-129` (`TestNormalSetIsTheReviewedSet`), `integration_test.go:131-141,578-598,737-814,900` (`TestMigrationLifecycle`); `core/internal/tenancy/integration_test.go:127-143` (`TestTenancyModel` purge/down steps); `access/{integration_test.go:233-235,recovery_integration_test.go:92}`; `core/internal/mcp/env_integration_test.go:253` (shared `mcpEnv` fixture of the five `TestMCP*` entry points, not a test of its own); `infra/validation/run.sh:1209-1230,1698-1728,2259-2294,3890,3917,4445-4451,4498-4500,4611-4652,4764` |
| P-5 | `00013` goose style (precondition `DO`, StatementBegin/End per function, postcondition, refused/guarded down) | `core/migrations/00013_organization_branding.sql:1-21,301-342,453+` |
| P-6 | importer boundary: only `access` imports `tenancy`; only `internal/mcp`, `domains/api`, `branding/api` import `access`; 11 `inContextTx` call sites; exported-method pin; SECURITY DEFINER allowance per named migration (`boundary_test.go:478-519`, 00013 = exactly 1) | `core/internal/tenancy/access/boundary_test.go:44-66,197,478-519,556` (`TestAccessBoundary`, `TestProductionImporters`, `TestRLSBoundary`); `core/internal/tenancy/boundary_test.go:25,74-76` (14 production files, pinned package functions; `TestPackageBoundary`, `TestSQLBoundary`) |
| P-7 | logrecord enums: `/internal/v1/*` route enum (two occurrences), `access.decision` operation and capability enums, `branding.request` event shape | `infra/validation/tools/internal/logrecord/{completion.schema.json:38-44,95-97; events.json:172-237,304,313,398-403; logrecord_test.go:71-72,137-180,339-350}` |
| P-8 | harness: `rls_go` runner (`run.sh:2618-2626`), `domains_checks` template (`:3196-3214`: catalog + IT rows, `it_count` subtest pins, `rls_logcheck`), `required-checks.tsv` header arithmetic (119 at `e9d35f3f`), `WEB_ROUTES` (`:614`, basenames under `web/app/`), B-TENANCY-HTTP probe arrays (`:2152-2181`), `profile-paths.tsv` | as cited |
| P-9 | web BFF pattern (G014): `force-dynamic` server page, `readCookie`/`ownOrigin`, hidden `csrf_token`, `readForm` (form encoding, exact Origin, 16 KiB `MAX_FORM_BYTES` at `session.ts:31`; r2 adds an optional per-call `maxBytes` option, default unchanged, §1.6), `seeOther`/`statusPage`/`methodNotAllowed`; `mcp-client.ts` strict-shape client with `readBounded`, `MAX_RESPONSE_BYTES` 64 KiB and the per-endpoint `successBytes` arithmetic (`MCP_LIST_MAX_RESPONSE_BYTES`, `MAX_CONNECTIONS` 100); `ShellNavItem.href` anchors only (`app-shell.tsx:16` at `e9d35f3f`); S-LANDMARKS "one Account link" | `web/app/connections/mcp/{page.tsx,create/route.ts}`, `web/lib/{mcp-client,session,identity-client,request-context}.ts:13-24,70-110`, `web/components/layout/app-shell.tsx`, `web/tests/browser/shell.spec.ts:32` |
| P-10 | seed and identity-mode projects: `mcp-seed.sql` (Org A/B/C; Research, Design, Operations; Dave group_member of Research; everyone else `org_member` only; no owner/admin, no profile); Playwright `IDENTITY_E2E=1` projects `identity-*`, `mcp-*` at `https://org-a.localhost:543` (`playwright.config.ts:102-107`) | `infra/identity/mcp-seed.sql`, `web/playwright.config.ts`, `web/tests/browser/mcp.spec.ts` |
| P-11 | bounded keyset list pattern (G014): `tenancy.SelectableGroupsPage(org, user, manager, after, limit)` reads `limit+1` rows after a UUID cursor; `access.ListSelectableGroups` validates `limit` 1..`MaxSelectableGroupsPage` (`invalid_page_size`) and the cursor (`invalid_cursor`); `TestMCPListBounds` | `core/internal/tenancy/mcp.go:18-41`, `core/internal/tenancy/access/mcp.go:325-336`, `access/mcp_test.go:115` |

### 0.3 R0 after G017 integration

Re-read every P-row against integrated `staging`; re-pin §11.2 (migrations, capabilities, audit operations,
required-checks count, route enums, `WEB_ROUTES`, `components/ui` set, S-LANDMARKS wording, G017's `ShellNavItem`/
area-navigation decision if a later slice changed it); rebind §2; record drift in the Result. Material drift (a G017
change to the session-scope helper, the shell navigation model, the seed, or the harness row pattern) returns as a
Plan revision. If G017 introduces area navigation or a shared scope helper, G018 **reuses** it and drops its own
(§1.1, D-5).

## 1. Approach

### 1.1 Trust boundaries first (what changes, what a malicious actor can now control)

| Boundary | Change | Containment |
|---|---|---|
| **Runtime creation grants** | 00006 revoked every runtime INSERT on `groups`, `role_bindings`, `business_profiles`, `group_business_profiles` and DELETE on `role_bindings`/`group_business_profiles` (`00006:397-404`); at version 13 no runtime or FORCE-owner statement can create a Group, binding, profile or association (the trusted primitives are exercised below version 6 and by the superuser seed). 00014 restores **INSERT only**, column-limited, on the four tables (the `groups` grant includes `id` for the member-safe keyed insert, §4.7); **no DELETE, no UPDATE change, no TRIGGER privilege** | Four FOR INSERT policies (§4.4) bind each insert to the active Organization (`ctx_org()`), a valid current Organization membership (`rls_org_role()` non-NULL; superseding D-1), Organization-only context for creation and the exact row shape. **Creator bootstrap (C1):** the creator's `group_owner` binding is written by a plain (SECURITY INVOKER) AFTER INSERT trigger on `groups` in the same statement as the Group; the `role_bindings` INSERT policy admits a binding only from inside that trigger (`pg_trigger_depth() > 0`) with the fixed shape `(ctx_user(), 'group', NEW.id, NEW.id, 'group_owner', 'membership')`. A direct runtime INSERT on `role_bindings` (depth 0) is refused whatever its content; an existing Group can never be claimed (the trigger fires only for a newly inserted Group row) and two bootstraps cannot race (there is no separate bootstrap statement). No reader of other subjects' bindings exists. Unlink stays impossible at runtime (no DELETE grant, no DELETE policy) |
| **Profile eligibility at the database (C2)** | the `group_business_profiles` INSERT policy also requires `app.rls_profile_eligible(business_profile_id)` for non-Organization-admin actors: the profile was **inserted by this transaction in this cluster** (r3/r4: its insertion-provenance pair `(created_xid, created_sysid)` equals `(pg_current_xact_id(), (pg_control_system()).system_identifier)`; both columns are assigned once by their defaults at INSERT, are in no runtime column grant, are immutable by trigger; the xid never repeats within a cluster and is never current again after commit, and the cluster identifier is what a supported logical restore into another cluster cannot reproduce (r4: a restored row carries the source pair, review r3) — an UPDATE cannot produce the pair, whatever the clock or the restore; the r2 timestamp predicate is withdrawn) **or** is attached to a Group on which the context subject currently holds `group_owner`/`group_admin` (selected-Group context restricts to that Group); the reader itself first requires a valid current context (`app.rls_org_role()` non-NULL, which also validates a selected Group) and returns false otherwise, so a direct call outside the policy discloses nothing after Organization membership removal | one STABLE SECURITY DEFINER boolean reader owned by `<db>_policy_reader` (00006 pattern, §4.3), reading only `(id, organization_id, created_xid, created_sysid)` of `business_profiles` and `(organization_id, group_id, business_profile_id)` of `group_business_profiles` under helper policies limited to `ctx_org()`, joined with the subject's own bindings (`role_bindings_policy_reader`). It returns no data. Same-tenant foreign keys are not the eligibility proof; the single deterministic ordering negative O-ONB-E (§6.2) proves that UPDATE-then-revocation cannot attach a profile |
| **First HTTP exposure of tenancy mutations** | a new bounded adapter `core/internal/onboarding/api` registers **eight method/path pairs on six paths** under `/internal/v1/groups*` and `/internal/v1/business-profiles*` (§5.3); B-TENANCY-HTTP's absence probes for exactly those two served core paths and the two served web paths are retired and replaced by positive closed-route coverage (D-3) | P-1 session scope copied byte-pinned (D-5, `TestDomainsCopyIsExact` pattern); every route calls a guarded facade operation; raw `tenancy.Service` stays unreachable (importer pins); JSON bodies strict with per-route whole-payload budgets computed from the existing field maxima (8 KiB; 96 KiB for the two profile-bearing routes, §5.3), closed error codes, no names/UUIDs of foreign objects in any error; list answers bounded (§1.5) |
| **Who may create a Group** | new Organization-scope capability `group.create` held by every current Organization membership role `org_owner`/`org_admin`/`org_member` (superseding D-1), in Organization-only context; the creator becomes the Group's initial `group_owner` atomically and gains nothing else | matrix + RLS policy agree; Group-role-only subjects (no current Organization binding), non-members, foreign-Organization, disabled or revoked subjects and selected-Group contexts receive `role_not_permitted`/`context_denied`/42501 twice (facade, then policy); a member's new Group grants no authority in any other Group and no existing-profile eligibility (§4.4) |
| **Who may complete onboarding** | new Group-scope capability `group.complete_onboarding` (`all` for org owner/admin, `own` for group_owner/group_admin; the context gate restricts selected-Group context to that Group) | the target Group is row-locked `FOR NO KEY UPDATE` in the global Group-id lock order (§1.4) so two completions serialize; the association unique key is the second line |
| **Existing-profile selection** | `business_profile.select_for_onboarding` evaluated in the **same** transaction and on the **same facts** as `group.complete_onboarding`; candidates are the subject's attached Groups (Organization-only) — the new unattached Group never qualifies | facade decision and the database reader agree (§4.4); unreadable → `not_found` (same as unknown/foreign); readable but unlinked path → `forbidden` (`no_linked_path`) recorded as `access.denied` |
| **Atomicity** | Group + owner binding in one statement; profile + association commit together inside the guarded transaction; no new status column | failure before `COMMIT` leaves nothing; `unavailable` after `COMMIT` may be durable and is reconciled by a guarded read (existing contract) |
| **Browser** | new form pages; first text inputs in `web/` | server-rendered, escaped data; `readForm` exact-Origin + CSRF with the computed 96 KiB form budget for the two profile forms (§1.6; the 16 KiB default stays for every other form); no client-side authority; no selected-Group concept in web (none exists; `Resolve` stays Organization-only); no MCP connection is touched by navigation (contract rule 8) |

### 1.2 Riskiest slice first

Slice 1 (§8) proves, against real PostgreSQL as the **runtime role** and as the FORCE-bound owner, the complete
creation and both completion paths plus their principal denials (including the fixed-actor runtime SQL negatives of
C1/C2) and the injected mid-transaction failures, before any HTTP or UI exists. Everything after it is thin composition.

### 1.3 Operation and capability matrix (R1 pin; the finite set)

| Guarded operation (new unless marked) | Capability | Target | Tx | Writes |
|---|---|---|---|---|
| `SuggestGroupSlug(tc, name)` | `group.create` | active Organization | RR RO | — (advisory; reserves nothing) |
| `CreateGroup(tc, name, slug)` | `group.create` | active Organization (TENANT-ORG only) | RC RW | one keyed `groups` INSERT (no RETURNING); the bootstrap trigger writes the creator's `group_owner` binding in that statement; then one read of the new row (§4.7 S-G1/S-G2) |
| `ListGroups(tc, after, limit)` | `group.read` per row | each Group of the active Organization in context, one page | RR RO | — |
| `GetGroupOnboarding(tc, g)` | `group.read` | exact Group | RR RO | — |
| `CompleteGroupOnboardingWithNewProfile(tc, g, ctx)` | `group.complete_onboarding` | exact Group (must have no association) | RC RW | `business_profiles` row (keyed), then `group_business_profiles` row |
| `CompleteGroupOnboardingWithProfile(tc, g, p)` | `group.complete_onboarding` on `g` **and** `business_profile.select_for_onboarding` on `p` | exact Group and exact profile | RC RW | `group_business_profiles` row |
| `ListBusinessProfiles(tc, after, limit)` | `business_profile.select_for_onboarding` per row (page form of `manageableProfiles`, `facade.go:309`) | each profile of the active Organization, one page | RR RO | — |
| `GetBusinessProfile` (existing) | `business_profile.profile.read` | exact profile | RR RO | — |
| `UpdateBusinessProfile` (existing) | `business_profile.profile.update` (`linked+flag`) | exact profile | RC RW | full replace, last commit wins (disclosed, R4) |

Matrix rows added (Organization-only; flag irrelevant):

| Capability | org_owner | org_admin | org_member | group_owner | group_admin | group_member |
|---|---|---|---|---|---|---|
| `group.create` | Y | Y | Y (superseding D-1, Spec `efd4295c`; Organization-only context) | — | — | — |
| `group.complete_onboarding` | all | all | — | own | own | — |

Nothing else changes in the matrix. `group.manage_members`, `privileged_settings.manage` and `audit.read` keep no
operation. No BusinessProfile role, no third tier, no standalone profile creation (D-7), no unlink/replace, no Group
rename/slug-change route (D-4; the existing guarded `RenameGroup`/`ChangeGroupSlug` stay Go-only).

### 1.4 Creation and completion contracts (R1, R2, R5)

**CreateGroup.** `pre`: `ValidateName`; empty slug → `suggestBase` (deterministic, no suffix search, as the raw
primitive); `ValidateSlug` (grammar, 24 reserved). Guard steps 1–5 (user, session, Organization `FOR SHARE`,
Organization binding, no Group candidates), decide `group.create` (any current Organization membership role,
superseding D-1), then `act` = new `tenancy.CreateGroupTx(ctx, tx, org, id, name, slug)`: **one writing statement**
(§4.7 S-G1, keyed by a Go-generated UUID v4 `tenancy.NewGroupID()`, **no RETURNING**) followed by the read of the row
just written (S-G2, `sqlSelectGroupByID`, visible to the creator because the bootstrap binding now exists). The
bootstrap trigger (§4.3) inserts the creator's `group_owner` binding inside S-G1 under the
`role_bindings_insert_bootstrap` policy; audit T4 `group.created` then T7 `membership.granted` fire from the triggers
with actor `ctx_user()` (`audit_capture_insert` sorts before `groups_bootstrap_owner`, so T4 precedes T7). Returns
`GroupOnboarding{Group, BusinessProfileID: nil, Complete: false}`. S-G2 returning no row is `unavailable` (the
transaction rolls back; it cannot happen under the pinned policies and is asserted by ONB-1, not designed around).
A primary-key collision of the generated id (23505 on `groups_pkey`) is `unavailable`, never a slug conflict.
Failure semantics are the existing table (AUTHORIZATION "What a caller can conclude"): `slug_conflict`→`conflict`
(field `slug`), `slug_reserved`/`slug_required`/`invalid_*`→`invalid_input`, denial→`forbidden` (Organization
established), `unavailable` possibly durable. **No Organization binding is written or changed; no second owner row;
no owner column** (amendment c73f4833). The creator is always `ctx_user()` of the transaction: there is no creator
parameter anywhere (Go, SQL or trigger).

**Completion, new profile.** `pre`: `ValidateBusinessContext`. Facts: `groupQuery` with `TargetGroup = g` locked
`FOR NO KEY UPDATE` (step 6) and associations read after the lock (step 7, the linearization point). Decide
`group.complete_onboarding`. State check on the locked facts: an existing association → `conflict` (field
`association_exists`; new `writeFailure` mapping) with zero writes. `act` = `tenancy.CreateBusinessProfileTx` (keyed
insert, no RETURNING, §4.7 S-P1) then `tenancy.LinkGroupBusinessProfileTx` (the existing association statement, S-P2).
T10 and T12 fire. A profile is never committed without its association: the association insert failing (unique key,
policy, fault) rolls back the profile. At the database the association policy admits the new profile because it was
**inserted** by the current transaction in this cluster (`(created_xid, created_sysid) = (pg_current_xact_id(),
(pg_control_system()).system_identifier)`, §4.4 (a)); the same policy refuses any pre-existing profile, including
one this transaction merely UPDATEd (its `created_xid` is a committed, never again current transaction id that
nothing at runtime can change) or one logically restored from another cluster (its `created_sysid` is the source
cluster's identifier, which the target's never equals, r4), and any unlinked profile.

**Completion, existing profile.** Facts: one combined `FactQuery{TargetGroup: g, Groups: nil, Profile: p, Lock: true}`
(`onboardingQuery(g, p)` in `access/onboarding.go`): the subject's Group bindings (locked, ascending `group_id`), the
target Group and the other candidate Groups (locked in the order below), their associations and the target profile.
Decide `group.complete_onboarding(g)` **then** `select_for_onboarding(p)`; either denial is recorded with the
established target. `act` = `LinkGroupBusinessProfileTx` (S-P2). No check-then-link gap: every relied-on
binding/association row is locked or read after the Group locks in the same transaction (existing O-ordering proof
extended, §6).

**Lock order (C4).** Today `ReadFacts` locks the target Group first (`sqlFactTargetGroupLock`, `FOR NO KEY UPDATE`)
and then the other candidate Groups (`sqlFactGroupsLock`, `ORDER BY id FOR SHARE`; `tenancy/facts.go:139-160`). The
existing shapes never combine both (`groupQuery(g)` passes `Groups: {g}`, so "others" is empty; `profileQuery(p)` has
no target), but the combined completion shape would: completion A (target g1, candidate g2) and completion B (target
g2, candidate g1) take X(g1)→S(g2) and X(g2)→S(g1) and deadlock. r1 pins **one rule for every Group row lock**: Group
rows are locked in ascending Group id, each with its own strength. `ReadFacts` with `Lock` and a target partitions
the other candidates around the target id and issues, in order: others with `id < target` (`sqlFactGroupsLock`,
`ORDER BY id FOR SHARE`), the target (`FOR NO KEY UPDATE`), others with `id > target` (`ORDER BY id FOR SHARE`); the
no-target shape (one ascending statement) and the single-target shapes are unchanged and already obey the rule.
Because every transaction acquires its Group locks in strictly ascending id, no wait-for cycle on Group rows can
form between completions, profile-target writes and single-target writes (S/S never conflicts; X/S and X/X conflicts
only wait on a lower-or-equal id held by a transaction that is itself only waiting on higher ids). Group binding
locks (`ORDER BY group_id FOR SHARE`) and the Organization share precede the Group locks as today. Cost: at most one
extra statement for the combined shape (E-PG row). No lock-order framework, no advisory locks, no new index.

**Concurrency and revocation.** Two completions of one Group: the second waits on the Group row lock, re-reads the
association at step 7 and answers `conflict`; the association unique key is the backstop (IT-BP-LINK-3 pattern).
A revocation (binding delete/downgrade, policy change, session revoke, user disable) that commits before the guard's
locking statement is observed and denies; one that waits behind the held Group/binding lock is not reported as having
won (existing orderings A/B/C, `ordering_integration_test.go:3-16`). A completion racing an unlink is impossible at
runtime (no DELETE); the trusted-unlink ordering case O-13 stays as the owner-pool proof.

**Retry / double submit / unknown commit (R5, AC7) — the pinned contract, no intent key, no claim of fresh creation
on recovery (C6):**

| Operation | Duplicate logical request | Lost response after `COMMIT` | Changed payload |
|---|---|---|---|
| Create Group | second `CreateGroup` with the same slug → `conflict` (field `slug`). The web BFF then reads `GET /internal/v1/groups?slug=<slug>` (exact-slug filter of the list route, ≤ 1 row, `group.read` applies) and shows an **authorized conflict**: "A Group with the slug `<slug>` already exists" with the readable Group's name and link, and the sentence "Nothing was created by this request; if an earlier attempt was interrupted, this may be that Group". If no readable Group matches: "The slug `<slug>` is unavailable in this Organization" (no name, no existence beyond what `slug_conflict` already states). The BFF never labels a slug match as "created" or "recovered" | the BFF shows an honest **unknown** state ("Your request may or may not have completed; check your Groups", link to the list); it never retries on its own | a different slug is a distinct intentional creation (names are not identity); same slug with a different name → `conflict`, nothing overwritten |
| Complete (either branch) | `conflict` (`association_exists`); the BFF re-reads `GetGroupOnboarding` and shows the **current** committed profile (same scoped UUIDs) with "This Group already has a Business Profile; the current state is shown" — not "completed by this request" | same | a different profile/context after a commit → `conflict`; the winner is never overwritten or duplicated |
| Suggest slug | idempotent read | — | — |

Browser double-submit is additionally bounded by the existing `pending-submit` component (disable on submit); the
server contract above is the guarantee. No idempotency service, no stored intent, no generalized retry framework.
The honest-conflict wording is proven by `http_test`/unit tests (the BFF status page never contains "created" on the
conflict path) and E-ONB-6.

### 1.5 Lists, context and the shared profile (R3, R4) — bounded (C6)

- Both lists reuse the G014 keyset pattern (P-11): `after` (UUID cursor; zero = first page), `limit` 1..100 (default
  50 in web; `invalid_page_size`/`invalid_cursor` → `invalid_input`), deterministic `ORDER BY id`, `limit+1` rows read
  to compute `next` (the last returned id, or absent). Traversal is complete and terminates: ids strictly ascend and
  every row visible to the subject appears exactly once. No generic pagination subsystem; no silent cap.
- `ListGroups(tc, after, limit)`: facts (steps 1–5, no lock) + one page statement over `app.groups`
  (`WHERE organization_id = $1 AND id > $2 ORDER BY id LIMIT $3`, RLS `groups_select` already restricts to readable
  rows: owner/admin all, others own) + one associations statement for the page's ids (`sqlFactAssociations`).
  `BusinessProfileID` (nullable) derives completion (D-2): no status column, no browser marker. Selected-Group context
  lists only that Group. Optional exact filter `slug=<slug>` (validated by `ValidateSlug`) for conflict recovery (§1.4).
- `ListBusinessProfiles(tc, after, limit)` is the Organization-only "My Business Profiles": the page statement scans
  RLS-visible profiles ascending from the cursor (`limit+1` rows); each scanned row is decided with
  `select_for_onboarding` on the facts (owner/admin: all profiles including unlinked; Group owner/admin: profiles
  attached to their Groups); member-only readable profiles are **not** listed and **not** selectable. A page may
  therefore hold fewer than `limit` admitted rows while `next` (the last **scanned** id) is present; the client
  follows `next` until absent. List rows are **summaries** `{id, display_name, industry}`; the full record is read by
  id. Nothing in a list is authority: every later write re-decides on fresh facts.
- Body bounds (pinned in `onboarding-client.ts` as `successBytes`, the `MCP_LIST_MAX_RESPONSE_BYTES` arithmetic):
  Group row ≤ 36+36 UUIDs, name ≤ 200 code points (≤ 1,200 escaped bytes), slug ≤ 63, flags ≈ 160 fixed → ≤ 1,500
  bytes; 100 rows + envelope < 160 KiB. Profile summary row ≤ 36 + 2×(255×6) + 60 → ≤ 3,200 bytes; 100 rows + envelope
  < 330 KiB. Single profile ≤ 3×1,530 + 24,000 + 12,288 + 200 < 42 KiB. Every other answer ≤ `MAX_RESPONSE_BYTES`
  (64 KiB). A client-side over-bound answer is `malformed`, never a silently truncated list.
- The shared editor is the existing `GetBusinessProfile`/`UpdateBusinessProfile`; a Group page's **Edit Business
  Profile** is a link to `/business-profiles/<uuid>`; no Group-owned copy. Members see **View Business Profile** and a
  read-only page (update → `forbidden` → "not permitted" status, nothing changed). Flag off: Group-derived edits
  `forbidden`, selection and reads unchanged (existing `decide`). The profile page never renders associations, sibling
  Groups or people (existing read contract).
- Web resolves Organization-only context only (no web selected-Group exists; P-9/P-10). Selected-Group semantics are
  proven at the facade (§6, extend `TestAccessPolicy` selected-context cases), not invented in the UI.
- Bounded queries (R5, PG-009): create = steps 1–5 + 1 insert; completion (new) = steps 1–7 + 2 inserts; completion
  (existing) = steps 1–7 (+1 lock statement when candidates straddle the target) + 1 insert; `ListGroups` = facts + 1
  page + 1 associations statement; `ListBusinessProfiles` = facts (bindings, Groups, associations) + 1 page statement
  — no per-row round trips. `group_business_profiles_business_profile_idx`, `groups_organization_slug_key` and the
  primary keys serve the lookups and cursors; no new index. Record one E-PG-style observation per operation in the
  Result (`TestAccessFootprint` rows, §6.3 b2), not a benchmark.

### 1.6 Web and BFF (R4, AC8)

Pattern P-9 (G014 MCP), server-rendered, progressive forms, no client authority:

| Route | Purpose |
|---|---|
| `GET /groups?after=` | list page (server order, `Next` link when `next` is present): name, slug, state (Incomplete / Complete), actions **Open** / **Resume onboarding**; **New Group** |
| `GET /groups/new?name=` | step 1 name → step 2 (same page) with the **server-suggested** editable slug prefilled; submit |
| `POST /groups/create` | BFF → core `POST /internal/v1/groups`; 303 to `/groups/<uuid>/onboarding`; `conflict` → the honest authorized-conflict page (§1.4); transport failure → the unknown-state page |
| `GET /groups/<uuid>` | Group page; incomplete → resume CTA; complete → profile summary + **Edit/View Business Profile** link |
| `GET /groups/<uuid>/onboarding?after=` | two explicit choices: **Create a new Business Profile** (five fields) or **Select an existing Business Profile** (radio list from one `ListBusinessProfiles` page, `Next` when `next` is present; empty state when none) |
| `POST /groups/<uuid>/onboarding/complete` | `mode=new` with context fields or `mode=existing` with `business_profile_id`; BFF → core; 303 to `/groups/<uuid>?status=completed`; `conflict` → re-read and show the current state (§1.4) |
| `GET /business-profiles?after=` | My Business Profiles (summaries, deduplicated by construction); each → shared page |
| `GET /business-profiles/<uuid>` | canonical shared page (read or edit per the `can_update` display hint, re-decided on submit); `POST /business-profiles/<uuid>/update` → `PUT /internal/v1/business-profiles/{id}`; last-commit-wins **disclosed on the page** (R4) |

Nine web routes (`WEB_ROUTES` +9: `groups/page.tsx groups/new/page.tsx groups/create/route.ts groups/[id]/page.tsx
groups/[id]/onboarding/page.tsx groups/[id]/onboarding/complete/route.ts business-profiles/page.tsx
business-profiles/[id]/page.tsx business-profiles/[id]/update/route.ts`). Shell: add **Groups** and **Business
Profiles** as sibling area links (widen `ShellNavItem.href` to `` `#${string}` | `/${string}` `` or add an `areas`
slot; whichever G017 did not already decide — R0). Components: adapt kit `input` and `textarea` (plain elements,
semantic tokens, lineage header, `data-slot`), native `<label>`, `<fieldset>` + native radios; **no new npm package**
(D-8). Error messages are associated by `aria-describedby`; focus moves to the first error; 320 px reflow; reduced
motion untouched. Inputs are rendered as escaped data; server echoes values only from the request it validated.

**Form budgets (r2, review item 3).** `readForm` keeps its 16 KiB default (`MAX_FORM_BYTES`, `session.ts:31`) for
`POST /groups/create` (name ≤ 200 code points → ≤ 2,400 form bytes, slug ≤ 63, `csrf_token`, `mode`) and every
existing form; the two profile forms (`POST /groups/<uuid>/onboarding/complete` with `mode=new`, `POST
/business-profiles/<uuid>/update`) call `readForm(request, { maxBytes: PROFILE_FORM_BYTES })`, a new optional second
parameter of `readForm`/`readBounded` whose absence means the current constant (existing callers and
`session.test.mjs` behaviour unchanged). `PROFILE_FORM_BYTES = 96 * 1024` is derived in `onboarding-client.ts` next
to `successBytes`: `application/x-www-form-urlencoded` encodes one code point as at most 12 bytes (an astral code
point is four UTF-8 bytes, each percent-encoded as three bytes; a BMP code point at most nine; a textarea newline is
normalized by the browser to `%0D%0A`, six); the five profile fields total 255+255+4000+255+2048 = 6,813 code points
→ ≤ 81,756 bytes; names/separators (`display_name=&legal_name=&description=&industry=&website=&mode=new&csrf_token=`)
and the token ≤ 1 KiB → ≤ 82,780 bytes < 98,304. The BFF normalizes `\r\n` → `\n` in the description before
validation/forwarding (the browser's submission normalization is not part of the stored value), then sends JSON
(`JSON.stringify`: ≤ 4 bytes per code point raw UTF-8, 6 for an escaped control character) ≤ 40,878 + envelope,
inside core's 96 KiB budget (§5.3). No field maximum is reduced; the unit test `onboarding-forms.test.mjs` pins the
arithmetic from the exported maxima and a max-field all-astral submission (accepted, forwarded intact) plus a
`PROFILE_FORM_BYTES + 1` submission (400, nothing forwarded).

### 1.7 Configuration

None expected: no new environment variable, product label (labels resolve through existing configuration) or vendor.

## 2. Source binding (provisional; rebind at R0)

| Input | Revision | SHA-256 |
|---|---|---|
| Base (r2): r1 commit `0eed3ad5` (on r0 `b8c8b8a8`, G017 `862db18d`, `staging` `b0873d12`, amendment `c73f4833`, `b4250a23`, `2d535708`) + `origin/staging` `68cb0d1e` and `efd4295c` merged (both management-only) | — | — |
| Goal `management/v0/goals/G018-group-business-profile-onboarding.md` (superseding D-1 clarification; the `2d535708` D-1 addendum `35b9f195…5517` is history) | `efd4295c` | `a58616dd95fbc6ee8f80860136e31255411a09c224816a1e24eb19bb165550b1` |
| Spec `management/v0/specs/G018-group-business-profile-onboarding-spec.md` (§"Superseding D-1 — Generic self-service Group onboarding"; the `2d535708` revision `6d86d694…6a09` is history) | `efd4295c` | `784feedbeb79d975b95a51f382f107e1bd6be380ba9a255abc3f977f6d52e653` |
| Codex review r1 `management/v0/results/G018-plan-codex-review-r1.md` (the binding correction list for r2) | `68cb0d1e` | `ec3855ed39cca99c7e10e49209781ca260e42b6763dc6c411f2b244f835f3b45` |
| ADR-004 `docs/adr/ADR-004-group-business-profile-cardinality.md` | `862db18d` | `51a487844d74e913d195c963095c2a381deeb293d1829085f2f026f060955a68` |
| Codex review r0 `management/v0/results/G018-plan-codex-review-r0.md` (reviewed at `b4250a23` = `edc1a240…63f0`; D-1 addendum at `2d535708`) | `2d535708` | `b545ebc02b290f314daf462babe4c3a318b345d4e80be3dcf3c03e635f1ee663` |
| G017 provisional head `origin/g017-organization-branding-logged-out-shell` | `e9d35f3f` (observed) | — |
| Allow-list proposal `management/v0/plans/G018-scope-allowlist.txt` | this commit | §11.3 |

## 3. Files (touch-list; actual caller/pin inventory, C5)

**Create**

| Path | Content |
|---|---|
| `core/migrations/00014_group_onboarding.sql` | §4 |
| `core/internal/tenancy/onboarding.go` + `onboarding_test.go` | `CreateGroupTx` (S-G1), `CreateBusinessProfileTx` (S-P1, keyed), `LinkGroupBusinessProfileTx` (S-P2), `GroupsPage`/`GroupAssociations` and `ProfilesPage` readers (keyset, P-11 pattern); production files 14 → **15**, pinned package functions +5 |
| `core/internal/tenancy/access/onboarding.go` + `onboarding_test.go` + `onboarding_integration_test.go` | the seven guarded operations (§1.3), `GroupOnboarding`/`BusinessProfileSummary`/`GroupPage`/`ProfilePage` types, `onboardingQuery(g, p)` combined fact query, page validation (unit, `TestOnboardingPageBounds` as `TestMCPListBounds`); `TestOnboarding` ONB-1..12 (§6.2) |
| `core/internal/migrate/onboarding_integration_test.go` | `TestOnboardingCatalog` (grants, policies, trigger, reader, postcondition, down) |
| `core/internal/onboarding/api/http.go` + `http_test.go` | routes §5.3; bounded copy of P-1 with the equivalence pin test (D-5, `TestDomainsCopyIsExact` pattern against `branding/api`) |
| `web/lib/onboarding-client.ts`, `web/tests/unit/onboarding-client.test.mjs`, `web/tests/unit/onboarding-forms.test.mjs` | strict-shape client with per-endpoint `successBytes` (as `mcp-client.ts`) and `PROFILE_FORM_BYTES` (§1.6); form/status-page unit tests (error association, honest conflict wording, form budget boundary) |
| `web/lib/session.ts` + `web/tests/unit/session.test.mjs` (r2, **modify**) | `readForm(request, options?: { maxBytes })` and `readBounded(request, maxBytes = MAX_FORM_BYTES)`: optional per-call budget, default and every existing caller unchanged; `session.test.mjs` gains the default-unchanged and explicit-budget cases |
| `web/app/groups/{page.tsx,new/page.tsx,create/route.ts}`, `web/app/groups/[id]/{page.tsx,onboarding/page.tsx,onboarding/complete/route.ts}`, `web/app/business-profiles/{page.tsx,[id]/page.tsx,[id]/update/route.ts}` | §1.6 |
| `web/components/ui/{input,textarea}.tsx`, `web/components/onboarding/*.tsx` | kit-adapted fields; form sections; state badge |
| `web/tests/browser/onboarding.spec.ts` | E-ONB-1..7 (identity-mode projects `onboarding-light/dark`) |
| `management/v0/results/G018-group-business-profile-onboarding-result.md`, `G018-scope-allowlist.txt` | Result; imported allow-list |

**Modify (production)**: `core/cmd/server/main.go` (register the handler after branding); `core/internal/tenancy/
facts.go` (the ascending-id partition of "other" candidate Groups around a locked target, §1.4 — the only existing
reader; r0's `access/facts.go` does not exist); `core/internal/tenancy/store.go` (new constants
`sqlInsertBusinessProfileKeyed`, `sqlFactGroupsPage`, `sqlFactProfilesPage`, reuse of `sqlSelectBySlug`;
`TestSQLBoundary` parses them); `core/internal/tenancy/access/{policy,errors,log,facade}.go` (two capabilities, matrix
rows, `association_exists`→`conflict`, seven `access.decision` operations, `manageableProfiles` page form);
`core/internal/audit/audit.go` (+7 operations; actions/reasons unchanged); `web/components/layout/app-shell.tsx` (area
links); `infra/identity/mcp-seed.sql` (Alice `org_admin` of A; Bob `group_owner` of Research with a seeded profile
attached to Research; Carol stays `org_member`; one profile attached to Design shared with Research; one profile in B
— inserted by the superuser seed as today, below the trigger's role condition).

**Modify (tests, pins, harness, docs)** — actual names: `core/internal/tenancy/{boundary,parity,integration,
business_profile_integration}_test.go` (`TestPackageBoundary` 15 files and +5 functions, `TestSQLBoundary` new
constants, `TestTenancyModel`/`TestBusinessProfileModel` version-14 purge/down steps);
`core/internal/tenancy/access/{boundary,matrix,policy,rls_expected,attack_expected,integration,ordering_integration,
rls_integration,rls_ordering_integration,attack_integration,audit_integration,footprint_integration,
recovery_integration}_test.go` (`TestAccessBoundary` exported methods +7 and `inContextTx` 11 → **17**
(`ListBusinessProfiles` reuses `manageableProfiles`' site), SECURITY DEFINER allowance for 00014 = exactly 1,
`TestProductionImporters` +1 importer; `TestOracleTables`/`TestOracleMatrix`/`TestOracleManageable` 15 capabilities and
cells; `TestPolicyGates`/`TestPolicyScope`; `rls_expected_test.go` r3 and `TestRLSIsolation` (four INSERT policies'
admitted/denied rows per context incl. the depth-0 binding negative, the unlinked/unrelated/member-only profile
negatives and the owner-bound FORCE cases); `attack_expected_test.go` r2 and `TestIsolationAttacks`;
`TestAccessOrdering` O-ONB-1..4 incl. O-ONB-D; `TestRLSOrdering` unchanged unless a lock statement text changes;
`TestAudit` AT-INV; `TestAccessFootprint` E-PG rows; `TestRecoverySeed` version pin); `core/internal/audit/
audit_test.go` (`TestCatalogParity` operations 29 → 36; source stays 00013 — 00014 touches no audit constraint);
version-14 pins (`core/migrations/migrations_test.go` `TestMigrationSet`; `core/internal/migrate/runner_test.go`
`TestNormalSetIsTheReviewedSet`; `core/internal/migrate/integration_test.go` `TestMigrationLifecycle` drops/
`metadataThroughVersion14Only`/down-from-14 step; `core/internal/migrate/rls_integration_test.go` `TestRLSCatalog`
policy/function counts if it pins the whole catalog; `core/internal/mcp/env_integration_test.go:253` fixture suffix);
`core/internal/identity/store_test.go` **unchanged** (no new RLS table); `web/tests/browser/{shell,mcp}.spec.ts`
(S-LANDMARKS/S-NAV extended for two area links), `web/tests/browser/support/{assets,modes}.ts` (seeded users),
`web/playwright.config.ts` (two projects), `infra/validation/run.sh` (migration pins, `onboarding_checks()` two rows
after `branding_checks()`, fallback FAIL list, `WEB_ROUTES` +9, B-TENANCY-HTTP arrays −2 core/−2 web plus the positive
401 pin (D-3), APPLICABLE path), `infra/validation/required-checks.tsv` (two rows, header arithmetic),
`infra/validation/lib.sh` (B-WEB-COMPONENTS `data-slot` set; `BROWSER_MATRIX_IDENTITY_E2E`),
`infra/ci/validate-controls.sh` (count pins only if fixture counts move), `infra/ci/profile-paths.tsv` (two foundation
rows: `web/app/groups/`, `web/app/business-profiles/`), `infra/validation/tools/internal/logrecord/{events.json,
completion.schema.json,logrecord_test.go}` (six core paths in both route enums, one `onboarding.request` event, seven
operations and two capabilities in the `access.decision` enums), docs `docs/foundation/{TENANCY,AUTHORIZATION,
DATABASE-ISOLATION,AUDIT,MIGRATIONS,WEB-SHELL,FOLLOW-UP,RECOVERY}.md`, `docs/{SECURITY,PERFORMANCE}.md` (PG-009),
`docs/ARCHITECTURE.md` only on Codex request.

**Not touched**: `core/internal/tenancy/{service,association,business_profile}.go` primitives (unchanged, still
trusted-only; their behaviour under the new policies is stated honestly in §5.5); `domains/api`, `branding/api`,
identity; `role_bindings` UPDATE/DELETE policies; `stored_objects`; MCP tool catalog (five tools; no capability-mapped
production tool — §5.3); the audit trigger functions; `tenancy/mcp.go`.

## 4. Schema and migration `00014_group_onboarding.sql` (P-5 style; C1–C3)

1. **Precondition** `DO`: owner role; version-13 shape (`app.organization_branding` exists; none of the four INSERT
   policies, the two helper policies, `app.groups_bootstrap_owner()`/its trigger or `app.rls_profile_eligible` exist;
   runtime holds no INSERT on the four tables; the policy reader holds nothing on `business_profiles`/
   `group_business_profiles`; `app.business_profiles` has exactly the eight 00004 columns and neither `created_xid`
   nor `created_sysid`; `pg_control_system()` is executable by the runtime and policy-reader roles — PostgreSQL's
   default PUBLIC grant, r4 — else the migration refuses, §10).
2. **Grants (runtime, column-limited, INSERT only)**: `groups (id, organization_id, name, slug)` — `id` is included
   because the superseding D-1 admits creators who cannot see the Group before the bootstrap trigger has run, so the
   insert is keyed and has no RETURNING (§4.7 S-G1); `role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)` —
   needed because the SECURITY INVOKER trigger inserts as the statement's role; `business_profiles (id,
   organization_id, display_name, legal_name, description, industry, website)` — `id` is client-generated so the
   insert needs no RETURNING (§4.7); `group_business_profiles (organization_id, group_id, business_profile_id)`.
   **No DELETE anywhere; no UPDATE change; no TRIGGER privilege; `created_at` and the binding/association keys stay
   database-generated** (the `*_keys_immutable` trigger inventory is untouched; one function body is replaced, below).
   Policy reader: `SELECT (id, organization_id, created_xid, created_sysid) ON app.business_profiles` and
   `SELECT (organization_id, group_id, business_profile_id) ON app.group_business_profiles`. No system-column grant:
   the r2 `xmin` reading and its grant alternative are withdrawn.
   **Insertion provenance (r3, review r2 item 1; r4 cluster qualification, review r3):** `ALTER TABLE
   app.business_profiles ADD COLUMN created_xid xid8 NOT NULL DEFAULT pg_current_xact_id(), ADD COLUMN created_sysid
   bigint NOT NULL DEFAULT (pg_control_system()).system_identifier` (one statement, one rewrite). Existing rows
   receive the migrating cluster's identifier and the migration transaction's id (committed before any runtime
   transaction, hence never current again). Properties that make the pair `(created_xid, created_sysid) =
   (pg_current_xact_id(), (pg_control_system()).system_identifier)` an insertion-only proof independent of wall
   clocks **and of logical restore**: (i) both values are assigned by the column defaults exactly at INSERT, evaluated
   as the inserting role; runtime's INSERT grant on `business_profiles` stays `(id, organization_id, display_name,
   legal_name, description, industry, website)` and its UPDATE grant exactly `(description, display_name, industry,
   legal_name, website)` (00007 `:295-297` ACL string), so no runtime statement can name either column;
   (ii) `app.business_profiles_keys_immutable()` is replaced (`CREATE OR REPLACE`, same owner/attributes/trigger) to
   raise `integrity_constraint_violation` when `(NEW.id, NEW.organization_id, NEW.created_xid, NEW.created_sysid) IS
   DISTINCT FROM (OLD.id, OLD.organization_id, OLD.created_xid, OLD.created_sysid)`, so even the FORCE-bound owner
   cannot move either value; (iii) `xid8` is the 64-bit epoch-qualified transaction id — it never wraps or repeats
   within one cluster's lifetime, and a committed transaction's id is never `pg_current_xact_id()` of a later
   transaction **of the same cluster**; (iv) `pg_control_system().system_identifier` is the 64-bit identifier `initdb`
   assigns to a cluster; the accepted recovery path already requires the target's to differ from the source's
   (`recovery.sh:556,562`, `target_alias`), so a row restored by `pg_restore` — which writes the dumped values and
   never evaluates the defaults — carries the **source** pair and can never equal the target's current pair, whatever
   transaction id the target is at. A physical copy or PITR carries the identifier and the transaction counter
   together, so (iii) continues to hold there. Therefore equality holds only for a row inserted by the evaluating
   transaction in the evaluating cluster; an UPDATEd row keeps its original pair whatever `xmin`, any timestamp or
   any restore says. `created_at` is no longer part of any proof. The reader's `pg_control_system()` call reads the
   control file once per evaluation (no table, no RLS; executable by PUBLIC by PostgreSQL default — pinned, not
   assumed, precondition 1 and CT-ONB-1); no wrapper function is added, so the function allowlist and the SECURITY
   DEFINER allowance are unchanged by r4. CT-ONB-1 asserts both columns (type, NOT NULL, default expressions via
   `pg_get_expr`), the two runtime grant strings, the replaced function body (`md5(prosrc)` pin, the 00010 `:254`
   pattern), the immutability refusal as owner for each of the two columns, the `has_function_privilege` pins for
   `pg_control_system()`, and the backfilled rows' `created_xid <> pg_current_xact_id()` with `created_sysid =
   (pg_control_system()).system_identifier`.
3. **Creator bootstrap (C1; replaces r0's count reader)**: `app.groups_bootstrap_owner() RETURNS trigger LANGUAGE
   plpgsql SET search_path = pg_catalog` (SECURITY INVOKER, owned by the owner, like the `*_keys_immutable` and
   `audit.capture_*` trigger functions — 00008 header: "no new role and no SECURITY DEFINER function"). Body: if
   `current_user` is neither `<db>_runtime` nor `<db>_owner` (names resolved at migration time as 00006 resolves `rt`/
   `ow`) or `app.ctx_user()` is NULL, return (the superuser seed and the pre-RLS trusted paths keep their explicit
   binding inserts); otherwise `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id,
   group_id, role, source) VALUES (NEW.organization_id, app.ctx_user(), 'group', NEW.id, NEW.id, 'group_owner',
   'membership')`. Trigger `groups_bootstrap_owner AFTER INSERT ON app.groups FOR EACH ROW`. Properties: the binding
   is created in the statement that creates the Group (atomic by construction; no second statement to lose or race);
   it fires only for a new Group row, so an existing Group cannot be claimed; its creator is always the transaction's
   context subject. It is not a SECURITY DEFINER mutation surface and grants nothing: the insert passes the
   `role_bindings_insert_bootstrap` policy as the statement's own role.
4. **Policies** (`TO <db>_runtime, <db>_owner`; U/O/G/R/MGR as in 00006):
   - `groups_insert` FOR INSERT WITH CHECK `organization_id = O AND G IS NULL AND R IS NOT NULL` (superseding D-1:
     `R = app.rls_org_role()` is non-NULL only for a current `org_owner`/`org_admin`/`org_member` membership binding
     of `U` in `O`; a Group role alone (no Organization binding) is NULL and never qualifies; `G IS NULL` pins
     Organization-only context). The earlier `MGR` form (r1, D-1 (a)) is superseded.
   - `role_bindings_insert_bootstrap` FOR INSERT WITH CHECK `pg_trigger_depth() > 0 AND organization_id = O AND G IS
     NULL AND R IS NOT NULL AND subject_user_id = U AND scope_type = 'group' AND group_id = scope_id AND
     role = 'group_owner' AND source = 'membership'`. The depth predicate admits a binding only from inside a trigger;
     the only trigger that inserts bindings is §4.3, and runtime cannot create triggers (no TRIGGER privilege; ACL
     pinned). A direct runtime INSERT with any content, including the creator's own owner shape on another Group,
     fails 42501 (fixed-actor negative ONB-2d). The composite foreign key proves the Group belongs to O. No second
     owner, no admin binding, no Organization binding, no grant to another subject (all G019).
   - `business_profiles_insert` FOR INSERT WITH CHECK `organization_id = O AND R IS NOT NULL AND (MGR OR EXISTS
     (SELECT 1 FROM app.role_bindings rb WHERE rb.subject_user_id = U AND rb.organization_id = O AND rb.scope_type =
     'group' AND rb.source = 'membership' AND rb.role IN ('group_owner','group_admin') AND (G IS NULL OR rb.scope_id =
     G)))` — the subject's own bindings are visible under `role_bindings_select`; the policy graph stays acyclic.
   - `group_business_profiles_insert` FOR INSERT WITH CHECK `organization_id = O AND R IS NOT NULL AND (G IS NULL OR
     group_id = G) AND (MGR OR (app.rls_group_role(group_id) IN ('group_owner','group_admin') AND
     app.rls_profile_eligible(business_profile_id)))` (C2). The unique key `group_business_profiles_group_key` keeps
     at-most-one; same-Organization is the composite foreign keys; eligibility is the reader.
   - **Reader** `app.rls_profile_eligible(p_profile uuid) RETURNS boolean` — `LANGUAGE plpgsql STABLE SECURITY DEFINER
     SET search_path = pg_catalog, pg_temp`, owned by `<db>_policy_reader` through the 00006 SET chain, `REVOKE ALL
     FROM PUBLIC`, `EXECUTE` to runtime and owner. **Context first (r2, review item 2):** `u := app.ctx_user()`,
     `o := app.ctx_org()`, `g := app.ctx_group()`, `r := app.rls_org_role()`; if `u`, `o` or `r` is NULL the reader
     returns false before reading any row. `rls_org_role()` (00006 `:248-278`, same owner, callable by its owner)
     is exactly the context validity every 00006 policy relies on: a current `membership` Organization binding of
     `u` in `o`, and, when `g` is set, `g` a Group of `o` on which `u` holds a Group binding (or `u` is an
     Organization owner/admin). A retained Group binding after Organization membership removal, a forged or
     foreign selected Group, or no context therefore yields false from the reader itself, not only from the calling
     policy; the boolean is disclosed to nobody without a valid current context. Then true iff
     (a) `EXISTS (SELECT 1 FROM app.business_profiles bp WHERE bp.id = p_profile AND bp.organization_id = o AND
     bp.created_xid = pg_current_xact_id() AND bp.created_sysid = (pg_control_system()).system_identifier)` — the
     row was **inserted** by this transaction **in this cluster** (§4.2 insertion provenance; r4 adds the cluster
     term): an UPDATE in the same transaction refreshes `xmin` but cannot touch either column, no timestamp is
     consulted, and a row logically restored from another cluster carries that cluster's identifier, so
     UPDATE-then-attach and restored-row-collision-then-attach are both refused by (a) whatever the clock or the
     target's transaction counter, and decided by (b) on the statement's fresh snapshot. (The r2 form `xmin` + `created_at = transaction_timestamp()` is withdrawn: finite timestamp
     precision and repeatable wall clocks allow a committed row to carry an equal timestamp, review r2 item 1.) Or
     (b) `EXISTS (SELECT 1
     FROM app.group_business_profiles a JOIN app.role_bindings rb ON rb.organization_id = a.organization_id AND
     rb.scope_type = 'group' AND rb.scope_id = a.group_id AND rb.source = 'membership' WHERE a.business_profile_id =
     p_profile AND a.organization_id = o AND rb.subject_user_id = u AND rb.role IN ('group_owner','group_admin') AND (g
     IS NULL OR a.group_id = g))`. It returns no data. The r1 caveat ("an UPDATE also refreshes `xmin`, but updaters
     are eligible under (b) anyway") was **unsound** across a concurrent revocation (review C2: the updater's last
     qualifying binding can be deleted by the trusted path `tenancy/service.go:387-407,465-467` between the UPDATE
     and the attachment, since neither takes the other's locks); r2 replaces it with the insertion proof above, proven
     by O-ONB-E (§6.2), now including its synthetic equal-timestamp case (c). No broader profile visibility, no new
     trust channel, no new grant beyond the policy reader's `created_xid`/`created_sysid` columns.
   - **Helper policies** (TO `<db>_policy_reader`): `business_profiles_policy_reader` FOR SELECT USING
     `organization_id = app.ctx_org()`; `group_business_profiles_policy_reader` FOR SELECT USING `organization_id =
     app.ctx_org()`. Both are reachable only through the boolean reader (the helper role is NOLOGIN, NOINHERIT,
     member of nothing; its only member is the owner by SET).
5. **Postcondition** `DO`: exact ACL of the four tables and the reader grants (00006 `:532-554` pattern, INSERT columns
   added, no DELETE/TRIGGER), policy inventory (+4 runtime/owner, +2 helper), function allowlist (+2:
   `groups_bootstrap_owner:ow:false:false:v:search_path=pg_catalog:-` and `rls_profile_eligible:pr:true:false:s:
   search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE`), trigger inventory of `groups` (+1), `business_profiles`
   column inventory (**ten** columns, `created_xid xid8 NOT NULL DEFAULT pg_current_xact_id()` and `created_sysid
   bigint NOT NULL DEFAULT (pg_control_system()).system_identifier`, no NULLs, every row's `created_sysid` equal to
   the live identifier) and the replaced `business_profiles_keys_immutable` body hash, refuse to commit otherwise.
6. **Down**: drop the trigger and its function, the six policies and the reader (through the SET chain), revoke the
   runtime and policy-reader grants, restore the 00004 body of `business_profiles_keys_immutable` and drop the
   `created_xid` and `created_sysid` columns, restore the version-13 ACL/function/policy/column postconditions.
   Reversible: no new relation or row shape; the two added columns carry only provenance (their loss on down
   discards nothing of the model; rows created at runtime remain ordinary tenancy rows). Data migration: the
   backfill is the ADD COLUMN defaults (one rewrite of the small `business_profiles` table), not a separate
   statement. **Restore/reopen interaction (r4):** the accepted recovery path needs no change — `pg_restore` writes
   both columns verbatim, the BEFORE UPDATE immutability trigger does not fire on restore and is itself restored
   from the set, the data fingerprint (`infra/recovery/sql/fingerprint-data.sql`, md5 of each table's row text)
   includes both columns on both sides and stays equal because nothing is rewritten, the catalog fingerprint is
   equal because both sides are at version 14, and `toc_check`/the target-identifier checks are untouched. No
   `infra/recovery/**` file is in scope and none is needed; `docs/foundation/RECOVERY.md` (already in scope) records
   the boundary: restored profiles are selectable only through branch (b) in the target, never through (a). Audit constraints untouched
   (no new action/reason); `TestCatalogParity` keeps reading 00013.
7. **Executable statement sequences under actual grants (C3)** — the facade writers use exactly these:
   - **S-G1** `CreateGroupTx` (superseding D-1; the member-safe sequence, reinstating the r1 `f3f1ed1e` (b) delta):
     `sqlInsertGroupKeyed` = `INSERT INTO app.groups (id, organization_id, name, slug) VALUES ($1, $2, $3, $4)` —
     **no RETURNING**, because `groups_select` is evaluated for RETURNING rows before AFTER triggers run and an
     `org_member` creator holds no binding on the new Group at that point (the row would be invisible and the
     statement refused); the id is a Go-generated UUID v4 (`tenancy.NewGroupID()`, as `NewBusinessProfileID()`). The
     bootstrap trigger writes the owner binding inside this statement. Then **S-G2** `sqlSelectGroupByID` =
     `SELECT id, organization_id, name, slug, created_at FROM app.groups WHERE id = $1` under `groups_select`: the
     creator now holds `group_owner` on the row, so it is visible to every creator role; an Organization owner/admin
     saw it already. Slug uniqueness (23505 `groups_organization_slug_key` → `slug_conflict`) is unchanged; a
     `groups_pkey` 23505 is `unavailable`. Runtime's `groups` INSERT column grant is exactly `(id, organization_id,
     name, slug)` (CT-ONB-1, ACL postcondition). The existing `sqlInsertGroup` (RETURNING, database-generated id)
     stays the trusted primitive below version 6 and in the seed; it is not used by the guarded path.
   - **S-P1** `CreateBusinessProfileTx`: `sqlInsertBusinessProfileKeyed` = `INSERT INTO app.business_profiles (id,
     organization_id, display_name, legal_name, description, industry, website) VALUES ($1..$7)` — **no RETURNING**,
     because `business_profiles_select` hides an unlinked profile from a Group-derived actor and a RETURNING row must
     pass the SELECT policy. The id is a Go-generated UUID v4 (`tenancy.NewBusinessProfileID()`); `created_at` is not
     part of `GroupOnboarding`.
   - **S-P2** `LinkGroupBusinessProfileTx`: `sqlInsertAssociation` verbatim (RETURNING `id, organization_id, group_id,
     business_profile_id, created_at`): runtime holds table SELECT on `group_business_profiles` and the row is visible
     to the target's owner/admin and to MGR under `group_business_profiles_select`.
   - No runtime statement touches `role_bindings`; `sqlInsertBinding` (RETURNING `id`/`created_at`, not granted) is
     not used at runtime.
   - CT-ONB-1 and ONB-1/5 execute each sequence as the real runtime role and as the FORCE-bound owner with context;
     ONB-3/6 inject faults between S-P1 and S-P2 (rollback proof).

Risk handling: additive grants only; no data migration; no new relation (relcounts 50/48 and the forced-RLS table
count stay); the policies are tested by the independent `rls_expected_test.go` oracle extension (r3), by
`TestOnboardingCatalog` (grants, 42501 for excluded/foreign/disabled actors and for depth-0 binding inserts, no
DELETE/TRIGGER, trigger and reader ownership/search_path, the insertion predicate true for a row inserted in the
same transaction and false for a previously committed row, for a previously committed row UPDATEd in the same
transaction and for a row carrying a colliding transaction id from another cluster (R-ONB-1, r4), the reader
returning false without context, the `created_xid`/`created_sysid` column/grant/immutability pins of §4.2, down/up
idempotence) and by the fixed-actor runtime SQL negatives in `TestIsolationAttacks`.
Unexpected 42P17 or lock-order findings stop the slice (§10).

## 5. Contracts and interfaces

### 5.1 Go surfaces (`core/internal/tenancy/access`)

```go
type GroupOnboarding struct {
    Group             tenancy.Group
    BusinessProfileID *tenancy.BusinessProfileID // nil ⇔ onboarding incomplete
    Complete          bool                        // derived: BusinessProfileID != nil
}
type GroupPage struct { Groups []GroupOnboarding; Next *tenancy.GroupID }
type BusinessProfileSummary struct { ID tenancy.BusinessProfileID; DisplayName string; Industry *string }
type ProfilePage struct { Profiles []BusinessProfileSummary; Next *tenancy.BusinessProfileID } // Next = last scanned id
func (s *Service) SuggestGroupSlug(ctx context.Context, tc *TenantContext, name string) (string, error)
func (s *Service) CreateGroup(ctx context.Context, tc *TenantContext, name, slug string) (GroupOnboarding, error)
func (s *Service) ListGroups(ctx context.Context, tc *TenantContext, after string, limit int, slug string) (GroupPage, error)
func (s *Service) GetGroupOnboarding(ctx context.Context, tc *TenantContext, g tenancy.GroupID) (GroupOnboarding, error)
func (s *Service) CompleteGroupOnboardingWithNewProfile(ctx context.Context, tc *TenantContext, g tenancy.GroupID, c tenancy.BusinessContext) (GroupOnboarding, error)
func (s *Service) CompleteGroupOnboardingWithProfile(ctx context.Context, tc *TenantContext, g tenancy.GroupID, p tenancy.BusinessProfileID) (GroupOnboarding, error)
func (s *Service) ListBusinessProfiles(ctx context.Context, tc *TenantContext, after string, limit int) (ProfilePage, error)
```
`tenancy`: `CreateGroupTx`, `CreateBusinessProfileTx`, `LinkGroupBusinessProfileTx`, `GroupsPage`, `ProfilesPage`
(caller's transaction; authorize nothing; `access` is their only caller). Existing `GetGroup` keeps its signature (MCP
uses it). `ReadFacts` keeps its signature; only its Group lock ordering for the target-plus-others shape changes (§1.4).

### 5.2 Capabilities (13 → 15) and `access.denied` operations (29 → 36)

`group.create` (Organization kind), `group.complete_onboarding` (Group kind). Operations: `group.suggest_slug`,
`group.create`, `group.list`, `group.onboarding.get`, `group.onboarding.complete_new`,
`group.onboarding.complete_existing`, `business_profile.list`. Audit actions (36) and reasons (58) unchanged; the
trigger-owned T4/T7/T10/T12 record the row changes with the request actor.

### 5.3 HTTP — `core/internal/onboarding/api` (all authenticated; P-1 scope; JSON; `no-store`)

Eight method/path pairs on six paths (the route enums take the six paths; `WEB_ROUTES` takes the nine web files):

| Method and path | Body / query | Answer |
|---|---|---|
| `GET /internal/v1/groups?origin=&after=&limit=&slug=` | — | `{"groups":[{id,name,slug,business_profile_id|null,complete}], "next": id|null, "csrf_token"}` (≤ 100 rows) |
| `POST /internal/v1/groups/slug-suggestion` | `{origin, csrf_token, name}` | `{"slug"}` or `invalid_request` |
| `POST /internal/v1/groups` | `{origin, csrf_token, name, slug?}` | `201 {"group":{…}}`; `409 conflict {"field":"slug"}` (no name, no id of the holder) |
| `GET /internal/v1/groups/{id}?origin=` | — | `{"group":{…}}` (onboarding view) |
| `POST /internal/v1/groups/{id}/business-profile` | `{origin, csrf_token, mode:"new", profile:{display_name, legal_name?, description?, industry?, website?}}` or `{…, mode:"existing", business_profile_id}` | `200 {"group":{…}}`; `409 conflict {"field":"association_exists"}` |
| `GET /internal/v1/business-profiles?origin=&after=&limit=` | — | `{"business_profiles":[{id, display_name, industry|null}], "next": id|null, "csrf_token"}` (≤ 100 summaries) |
| `GET /internal/v1/business-profiles/{id}?origin=` | — | `{"business_profile":{id, display_name, legal_name, description, industry, website}, "can_update": bool}` — `can_update` is a **display hint** from a second `decide` on the same snapshot; the page renders the editor or the read-only view from it, and `PUT` always reauthorizes on fresh facts; it implies nothing about sibling visibility |
| `PUT /internal/v1/business-profiles/{id}` | `{origin, csrf_token, profile:{…}}` | `200 {"business_profile":{…}}` |

**Request budgets (r2, review item 3; P-9's 8 KiB `maxBody` is not copied unchanged).** The adapter decodes each JSON
body through `io.LimitReader(r.Body, budget+1)` with `DisallowUnknownFields` and rejects a body over its budget as
`invalid_request` before any guard or write. Budgets are whole-payload bounds computed from the existing field
contract so that every value the validators accept is admitted in **any** valid JSON encoding: one code point costs
at most 12 bytes (an astral code point escaped as a surrogate pair `😀`; raw UTF-8 is ≤ 4, a BMP escape
6); `ValidateBusinessContext` maxima are code points (`business_context.go:13-17`: 255, 255, 4000, 255, 2048 =
6,813 → ≤ 81,756 bytes), `ValidateName` ≤ 200 (≤ 2,400), `ValidateSlug` ≤ 63; keys, structure, `mode`, `origin`
(≤ 512) and `csrf_token` (≤ 256) are allowed 1 KiB. Pinned: `maxBody = 8 << 10` for `POST /groups/slug-suggestion`,
`POST /groups` (≤ 3,487 bytes worst case) and every GET; `maxProfileBody = 96 << 10` (98,304 ≥ 82,780) for `POST
/groups/{id}/business-profile` and `PUT /business-profiles/{id}`. Insignificant whitespace counts toward the budget
(the budget admits the canonical encoding; it is not a whitespace allowance). A body within budget whose field
exceeds its code-point maximum is still `invalid_request` with the tenancy `field` code — the byte budget never
replaces or shrinks field validation. `TestOnboardingBodyBudgets` (unit, `http_test.go`) recomputes both budgets
from the exported tenancy maxima (so a changed maximum fails the pin), sends the max-field all-astral payload raw
and fully escaped (both accepted and validated intact), a 4,001-code-point description (`invalid_request`, field
`description`, zero writes) and a `maxProfileBody + 1` payload (`invalid_request`, nothing decoded further).
Response bounds stay as §1.5 (Go's encoder: ≤ 6 bytes per code point).

Closed codes: `invalid_request` 400 (with `field` = tenancy code, incl. `invalid_page_size`/`invalid_cursor`),
`ambiguous_parameter` 400, `origin_rejected` 403, `unauthenticated` 401, `csrf_failed` 403, `not_available` 403
(`context_denied`), `forbidden` 403, `not_found` 404, `method_not_allowed` 405 (closed method set per path),
`conflict` 409, `onboarding_unavailable` 503 (any `unavailable`, possibly durable — the client reconciles by a read).
Unknown and foreign UUIDs are `not_found`; unknown sub-paths are 404; errors carry no names. One `onboarding.request`
log event per request (operation, outcome, closed reason, 12-hex host tag; never UUIDs/names/payloads). **No MCP
change**: no capability-mapped production tool is added; the five-tool catalog, discovery and `connection-context.md`
are untouched. The `/internal/v1` routes are web-BFF-only internal contracts (same class as branding/domains), not
public API. **D-3 positive coverage** (replacing the four retired absence probes): for every pair, unauthenticated →
401 closed body; mutation without CSRF/with a foreign Origin → 403; every non-listed method → 405; sub-path probing
(`/internal/v1/groups/{id}/x`, `/internal/v1/business-profiles/x/y`) → 404; a Group-role-only subject (no current
Organization binding) creating → 403 `forbidden` with zero rows, and an `org_member` creating → 201 with exactly the
Group and its own owner binding (superseding D-1); cross-Organization ids → 404; `ListGroups`/`ListBusinessProfiles` answers never exceed the pinned
bounds; B-TENANCY-HTTP keeps its remaining 16 core and 10 web absence probes and adds that the two served core paths
answer 401 (not 404) without a session.

### 5.4 Web — §1.6. `onboarding-client.ts` exports `listGroups`, `suggestSlug`, `createGroup`, `getGroup`,
`completeOnboarding`, `listProfiles`, `getProfile`, `updateProfile`; strict shapes with per-endpoint `successBytes`
(§1.5), anything unexpected → `malformed`.

### 5.5 Completion/association mutation-path inventory (R2, AC5) — corrected (C3)

| Path | Can it attach / remove a completed Group's profile? |
|---|---|
| guarded `CompleteGroupOnboarding*` | attach only to a Group **without** association; `conflict` otherwise; never replaces |
| guarded `UpdateBusinessProfile` | edits context; never touches associations |
| raw `tenancy.Service` primitives (`CreateGroup`, `CreateBusinessProfile`, `LinkGroupBusinessProfile`, `UnlinkGroupBusinessProfile`) | trusted-only by the importer pins; they are **not** a privilege boundary: SQL policy applies equally to identical statements. Under the runtime role with a context: `UnlinkGroupBusinessProfile` fails 42501 (no DELETE grant); `LinkGroupBusinessProfile` is admitted exactly when the guarded association policy admits it (eligible profile, Group without association); `CreateBusinessProfile` (RETURNING) is refused by the RETURNING visibility check for a Group-derived actor and admitted for an Organization owner/admin (an unlinked profile only that actor can read); `CreateGroup`'s Group statement is admitted (the trigger bootstraps the owner) and its explicit binding statement is refused at depth 0, so the transaction fails and nothing commits. Below version 6 and as the superuser they behave as today |
| runtime SQL (credential holder with a fixed context) | INSERT admitted only by the four policies (never a second association: unique key; never a direct binding: depth 0; never an ineligible profile: reader); DELETE impossible |
| owner pool / migrator / DBA | FORCE-bound to the same policies with a context; without a context nothing is visible or writable; the superuser is the documented restricted repair path (DATABASE-ISOLATION, RECOVERY) — no application path |
| MCP tools | none touch associations (`context.*`, `logout`, reads) |
| browser | no route; a stale "complete" page re-reads on every load (no localStorage/step flag) |

### 5.6 Settled Group-creation and permission interfaces (for the owner check-in and the G019 reassessment)

These are the interfaces G018 fixes and G019 is expected to build on; they become settled at Plan ACCEPT (D-1 is
resolved):

1. **Creation is one guarded operation** `access.CreateGroup(tc, name, slug)` under capability `group.create`
   (Organization kind; every current Organization membership role `org_owner`/`org_admin`/`org_member` per the
   owner's superseding D-1 (Spec `efd4295c`); no Organization grant is written and a Group role alone — without a
   current Organization binding — never creates), in Organization-only context, executing one keyed `groups`
   INSERT without RETURNING whose AFTER INSERT trigger writes the verified creator's `role_bindings` row
   `(scope_type='group', scope_id=group_id, group_id, role='group_owner', source='membership', subject =
   ctx_user())`, followed by one read of the new row (§4.7 S-G1/S-G2). The creator is always the authenticated
   context subject; no actor parameter, no provider claim, no Organization binding written; the new Group confers
   no authority in any other Group and no eligibility for any existing profile. Failure leaves neither row. Public
   signup, enrollment, invitation and membership provisioning are not part of this interface (G042/G043).
2. **Ownership cardinality**: one or more owners, at least one. G018 adds **no** unique-owner index, owner column,
   `created_by`, parallel owner store or admin binding; `role_bindings_subject_scope_key` (one binding per subject
   and scope) is the only uniqueness. The schema and API therefore already admit later additional owners.
3. **Runtime write surface on `role_bindings` after G018**: INSERT admitted **only** from inside the bootstrap trigger
   (`role_bindings_insert_bootstrap`: self, `group_owner`, the Group being inserted); UPDATE(role) stays lock-only
   (WITH CHECK false); no DELETE. G019 adds its own guarded promotion/transfer/removal operations and policies (each
   with its own admission predicate; the depth predicate is specific to the bootstrap policy), the last-owner guard
   (database and runtime), and the legacy bootstrap reconciliation for ownerless Groups (the trigger never fires for
   an existing Group; such repair is an owner-pool/G019 path); G018 neither pre-empts nor blocks them.
4. **`group.manage_members`** remains without an operation and is not authority over `group_owner` bindings
   (AUTHORIZATION matrix note); G019 owns membership administration UI/operations.
5. **Completion** is derived: a Group is complete ⇔ its single `group_business_profiles` row exists; no status column.
   `GroupOnboarding{BusinessProfileID, Complete}` is the read contract (`GET /internal/v1/groups/{id}`).
6. **Capability set after G018 (15)**: the thirteen of `862db18d` plus `group.create` and `group.complete_onboarding`.
   `business_profile.select_for_onboarding` gains an operation (existing-profile completion and the manageable list)
   with its existing semantics (flag-independent; Organization-only candidates are all of the subject's Groups), now
   also enforced at the database by `app.rls_profile_eligible` for the association insert.
7. **Group row lock order**: ascending Group id with per-row strength (§1.4) is the rule for every multi-Group lock in
   `ReadFacts`; G019 membership writes that lock Groups follow it.
8. **Emergency suspension/revocation** keeps its current semantics (binding retained, access denied via session/user
   checks at guard steps 1–2); G018 changes nothing there.

## 6. Validation and test mapping

### 6.1 Existing coverage inspected (reuse unless stated; actual names)

`core/internal/tenancy/integration_test.go` `TestTenancyModel` IT-OWN-1..6 (creator owner atomicity, IT-OWN-3 injected
failure, IT-OWN-5 concurrent same slug), IT-SLUG-1..6 (below version 6); `business_profile_integration_test.go`
`TestBusinessProfileModel` IT-BP-LINK-1..5, IT-BP-ATOM-1/2; `tenancy_test.go` (`TestServiceRejectsBeforeSQL`),
`business_context_test.go`, `slug_test.go`, `names_test.go`, `parity_test.go` (`TestMigrationParity`); `access/
integration_test.go` `TestAccessPolicy` (two-Organization `itWorld`: g1/g2 share p1, p2 on g3, p3 unlinked, gNew
unattached, gB/pB — exactly the Spec's fixture shape), `manageableProfiles` cases (`:1588-1615, 1717, 2142`),
selected-context cases, 286-cell matrix; `matrix_test.go` (`TestOracleTables/Matrix/Scope/Manageable/Gates`),
`policy_test.go` (`TestPolicyGates/Scope`); `ordering_integration_test.go` `TestAccessOrdering` (orderings A/B/C,
O-5, O-13 unlink race, O-14 both-orders deadlock-freedom, O-15, E-PG-3a/b), `rls_ordering_integration_test.go`
`TestRLSOrdering`; `rls_expected_test.go` + `rls_integration_test.go` `TestRLSIsolation` (independent oracle,
visible/update rows per table), `TestRLSPool`, `TestRLSFootprint`; `attack_expected_test.go` + `attack_integration_test.go`
`TestIsolationAttacks` (D1-01 trusted unlink); `footprint_integration_test.go` `TestAccessFootprint`;
`audit_integration_test.go` `TestAudit` AT-INV; `migrate/*_integration_test.go` catalogs (`TestRLSCatalog`,
`TestDomainsCatalog`, `TestBrandingCatalog`); `access/mcp_test.go` `TestMCPListBounds`; `branding/api/http_test.go`
`TestDomainsCopyIsExact`; `web/tests/browser/{shell,accessibility,overlay,mcp,appearance,tokens}.spec.ts`;
`web/tests/unit/mcp-connections.test.mjs`.

### 6.2 Scenario groups → tests (reuse / extend / new; the distinct failure each new case catches)

| Group | Case | Kind | Distinct failure caught |
|---|---|---|---|
| (1) create/owner/slug | IT-OWN-1..6, IT-SLUG-1..6, `TestServiceRejectsBeforeSQL` | **reuse** | primitive atomicity/validation (unchanged) |
| | ONB-1 permitted creator (`org_owner`, `org_admin` **and `org_member`**, superseding D-1) creates; both rows from the one keyed statement, S-G2 reads the row as the creator (the member case proves the member-safe sequence under actual grants); T4 then T7 with actor; no Organization binding change; `Complete=false`; also as the FORCE-bound owner with context | **new** (`TestOnboarding`, runtime role) | guarded path, trigger and policies admit the legitimate case, including the member whose RETURNING would have been refused |
| | ONB-2 (a) a `group_owner`/`group_admin` **without a current Organization binding** (IT-CTX-6 shape), non-member, foreign Organization, disabled user, revoked session, forged/ended scope, and any creator in **selected-Group context**: denied, zero rows, `access.denied` where recordable; (b) an `org_member`'s new Group grants nothing elsewhere: the member cannot read, complete or attach to another member's Group (`not_found`/`forbidden`, 42501 at the policy) and `ListGroups` shows only its own Groups; **(d) fixed-actor runtime SQL**: with the context of a legitimate creator (owner and member variants), `INSERT INTO app.role_bindings` of its own owner shape on another owner's existing Group, on its own existing Group and on a Group of another Organization → 42501 each, zero rows; a keyed `INSERT INTO app.groups` with a foreign `organization_id`, in selected-Group context, or as a Group-role-only subject → 42501 | **new** | facade denial and 42501 at the policy agree; the depth-0 negative (C1); the superseding-D-1 negatives (other Group, no current Organization membership) |
| | ONB-3 (a) injected failure inside S-G1 (hookDB after the Group statement, before COMMIT): neither row; **(b) deterministic bootstrap race**: two transactions creating different slugs and two creating the same slug, interleaved at the statement seam — each winner holds exactly one owner binding, the same-slug loser gets `conflict`, no Group ever exists without its owner binding at any observation point (owner-pool read after each commit) | **extend** IT-OWN-3/IT-OWN-5 pattern into the guarded tx | guarded transaction atomicity and race-safety of the bootstrap (C1) |
| | ONB-4 same-slug concurrent creates: one winner, one `conflict`; same slug in Organization B independent; rename/slug change keeps UUID, binding, association (existing guarded ops) | **extend** IT-OWN-5/IT-SLUG-3 + `TestAccessPolicy` rename cases | HTTP-level/guarded conflict code and cross-tenant independence under RLS |
| | CT-ONB-1 `TestOnboardingCatalog`: grants exact (incl. `groups.id` INSERT, policy-reader columns `(id, organization_id, created_xid, created_sysid)`, runtime INSERT/UPDATE on `business_profiles` without either provenance column), no DELETE/TRIGGER, six policies, trigger + function + reader ownership/search_path/volatility, the `created_xid` and `created_sysid` columns (type/NOT NULL/default expressions; backfilled rows not current, `created_sysid` equal to the live `pg_control_system().system_identifier`), `has_function_privilege` of `pg_control_system()` for the runtime and policy-reader roles, the replaced `business_profiles_keys_immutable` body hash and its refusal of a `created_xid` change and of a `created_sysid` change as owner, insertion predicate true for a same-transaction insert and false for a committed row and for a committed row UPDATEd in the same transaction, reader false without context, down/up (both columns dropped, 00004 body restored) | **new** (`migrate`) | migration shape and least privilege; the insertion-provenance pins (r3, r4) |
| | **R-ONB-1** (r4, review r3; fixture corrected in r5, review r4 — option (a): privileged inspection, then the fixed runtime actor; the deterministic restored-row/target-id collision pair, subcases `CT-ONB-1/R-ONB-1-neg` and `CT-ONB-1/R-ONB-1-pos` of `TestOnboardingCatalog`, same file; connections: the package's separately attested **superuser fixture connection** (`DB_SUPERUSER`, required from version 6, `rls_integration_test.go:44`) and the runtime connection). Fixture: Organization o, user u with a current `org_member` binding and `group_owner` of an incomplete Group gN only (no other path); the unchanged `business_profiles_select` policy therefore hides the unattached p_r from u, so u never inspects it — the inspection is privileged, the denial is runtime. **Negative**, one transaction T on the superuser fixture connection (the real restore actor: RLS does not apply, every column writable, no default evaluated — the privilege shape of `pg_restore`'s COPY; no grant is widened): (1) as the superuser, `INSERT INTO app.business_profiles (id, organization_id, display_name, …, created_xid, created_sysid) VALUES (p_r, o, …, pg_current_xact_id(), (pg_control_system()).system_identifier + 1)` — a row whose stored transaction id **equals** the evaluating transaction's id (the exact target collision the review describes, produced deterministically because T itself is the colliding transaction) under a foreign cluster identifier; (2) still as the superuser, assert `SELECT created_xid = pg_current_xact_id() FROM app.business_profiles WHERE id = p_r` is **true** (one row; the collision is real) and `UPDATE app.business_profiles SET description = … WHERE id = p_r` affects **one** row (the `xmin` refresh the review r3 names, now performed by the actor who can see the row); (3) `SET LOCAL ROLE <db>_runtime` and u's Organization-only context through the package's transaction-local `set_config(…, true)` statements (`rls_integration_test.go:94-95`; the fixed runtime actor, policies in force, no superuser bypass from here on): `SELECT app.rls_profile_eligible(p_r)` → **false**, and `INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES (o, gN, p_r)` → **42501**, zero rows; (4) `ROLLBACK` (the 42501 aborts T; nothing is committed). One connection, one transaction, so the reader evaluates against the live colliding id; no sleep, no second connection, no restore run. **Positive**, a fresh transaction T′ on the runtime connection (u, same context, no role switch): S-P1 keyed insert of p_f (defaults evaluated), inside T′ `SELECT app.rls_profile_eligible(p_f)` → **true** and the association INSERT `(o, gN, p_f)` admitted (one row), COMMIT; afterwards, on the superuser fixture connection, `created_sysid = (pg_control_system()).system_identifier` and `created_xid <> pg_current_xact_id()` for p_f. The negative fails on the r3 design (xid-only branch (a) would admit p_r); the positive fails if the cluster term ever refused a legitimate insert | **new** subcases of CT-ONB-1 (`migrate`) | the supported logical-restore boundary: a restored row's transaction id colliding with the target's current id (review r3); the legitimate fresh-insertion path under the cluster term |
| (2) completion | IT-BP-LINK-2/3, IT-BP-ATOM-1/2 | **reuse** | primitive at-most-one and atomicity |
| | ONB-5 new-profile completion (S-P1 + S-P2): one coherent state; T10+T12; `Complete=true`; as Group owner (flag off and on) and as org admin | **new** | the combined guarded write under actual grants (C3) |
| | ONB-6 faults: validation (zero writes), hookDB fault at S-P2 → no orphan profile; audit-store fault (strict trigger path) → rollback | **new** | profile-without-association orphan |
| | ONB-7 deterministic precommit revocation (binding delete / `group_admin`→`group_member` / user disable / session revoke) that commits before the Group lock: denied, Group incomplete; revocation waiting behind the lock: not reported as winning (ordering B/C) | **extend** `TestAccessOrdering` (O-ONB-1..3 on the hookDB seams) | revocation ordering on the new write path |
| | **O-ONB-D** (fixture corrected in r3, review r2 item 2) subject u holds `group_owner` on three Groups with ascending ids g1 < g2 < g3; g1 and g2 are **incomplete** (no association), g3 is attached to profile p3 (u's qualifying path). Completion A: target g1 ← p3 (existing-profile branch; combined `onboardingQuery(g1, p3)` locks u's Groups in ascending id with per-row strength: X(g1), S(g2), S(g3)); completion B: target g2 ← p3 (S(g1), X(g2), S(g3)). Both targets are in both actors' candidate sets, so without the ascending rule A would take X(g1)→S(g2) while B takes X(g2)→S(g1) (the 40P01 shape). Both interleavings at the lock seams: both commit, no 40P01, g1 and g2 each end with exactly one association to p3 (a shared profile, ADR-004), g3 unchanged | **extend** `TestAccessOrdering` (O-14 pattern) | combined-facts lock order (C4); same-target ONB-8 does not cover it |
| | **O-ONB-E** (r2, review item 1; the single deterministic runtime ordering negative) in `TestAccessOrdering` on two connections: fixture u owns gA and gB, p1 attached to gA, flag on. T1 (runtime, u, Organization-only) `UPDATE app.business_profiles SET description = … WHERE id = p1` (admitted; `xmin` now T1's). T2 (trusted `tenancy/service.go:387-407` binding removal, the O-13/E-PG-3b pattern) deletes u's gA binding and **commits** without blocking (no shared lock). T1 then `INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES (o, gB, p1)` → **42501**, zero rows, T1 rolled back; `SELECT app.rls_profile_eligible(p1)` inside T1 after T2's commit is false. Mirror positive in the same case: T1′ inserts a new profile pN (S-P1) and attaches it to gB after the identical T2 → admitted (legitimate new-profile path retained, ONB-5). **(c) synthetic equal-timestamp negative (r3, review r2 item 1):** fixture p_eq attached to gA; T1 (runtime, u) first reads its own `transaction_timestamp()`; the owner connection then sets `p_eq.created_at` to exactly that value and commits (the owner may write `created_at`; runtime cannot); T1 `UPDATE`s p_eq (so `xmin` is T1's **and** `created_at` equals T1's timestamp — the r2 predicate would have passed), T2 revokes u's gA binding and commits, T1 attaches p_eq to gB → **42501**, and `SELECT app.rls_profile_eligible(p_eq)` in T1 is false, because `created_xid` is the fixture transaction's id. Deterministic: no clock dependence, no timing loop. All three branches are step-ordered by the existing two-connection harness | **extend** `TestAccessOrdering` | UPDATE-refreshed `xmin` (and an equal timestamp) bypassing fresh eligibility after a concurrent revocation (the r1/r2 C2 holes) |
| | ONB-13 (superseding D-1) an `org_member` creates its Group, completes it with a new profile (ONB-5 as member) and with an existing profile only where it already holds `group_owner`/`group_admin` on an attached Group; without such a path the new Group alone never qualifies (`not_found`/`forbidden`, 42501) | **new** | the member self-service positive and the pre-existing-eligibility rule under the widened creation role |
| | ONB-8 two concurrent completions of one Group (new×new, new×existing): exactly one association, loser `conflict`, no second profile | **new** | serialization on the Group lock + unique key under the guard |
| | ONB-9 existing selection at the facade **and** as fixed-actor runtime SQL against the policy: Group owner/admin of an attached Group (g1 owner selects p1 for gNew) allowed; org owner/admin allowed incl. unlinked p3; **unrelated** (p2 for a g1-only owner), **member-only** (p1 for a g1 member who owns gNew), **unlinked** (p3 for a Group owner), foreign (pB) denied with `not_found`/`forbidden` at the facade and 42501 at the policy; the new unattached Group alone never qualifies; new-profile branch admitted only for the profile **inserted** by the same transaction (an UPDATEd committed profile is refused, r3 row); last qualifying path removed before the lock → denied, zero effect; a second remaining path still works; **(e) direct reader calls (r2, review item 2)** as the runtime role: `SELECT app.rls_profile_eligible(p1)` with no context → false; with u's Organization binding removed by the trusted path while its gA Group binding and a selected-Group context on gA are retained (attack C2 shape) → false, and the association INSERT → 42501; with a selected Group of another Organization or one u is not bound to → false; the legitimate context → true (the only positive) | **extend** `TestAccessPolicy` manageable/selected cases + **new** ONB write cases + `rls_expected` r3 rows + attack row C3 | manufactured access through the new Group, through runtime SQL, or through the reader called outside its policy (C2) |
| (3) lists/edits/context | `TestOracleManageable`, `TestAccessPolicy` profile read/update/flag cases | **reuse** | matrix semantics |
| | ONB-10 `ListGroups`/`ListBusinessProfiles`: complete ascending traversal across pages (fixture with 2×limit+1 rows; member-only profiles skipped yet `next` present; final page `next` absent), `invalid_page_size`/`invalid_cursor`, `slug=` exact filter, dedup by construction, foreign absent, selected-Group context restricts, flag off keeps selection/read and denies Group-derived update; no associations/sibling data in profile answers; body bound arithmetic (`TestOnboardingPageBounds`, unit) | **extend** matrix/`TestAccessPolicy` (list cells) + **new** list assertions | bounded deterministic lists (C6) and leakage |
| (4) retry/resume/UX | ONB-11 duplicate create (same slug) → `conflict`; lost-response simulation (hookDB after COMMIT) then retry → `conflict`, and `ListGroups(slug=)` finds the committed Group as an authorized conflict (never labelled created); duplicate completion → `conflict` and `GetGroupOnboarding` shows the committed profile; changed payload after commit → `conflict`, nothing overwritten | **new** | pinned retry/unknown-commit contract (C6) |
| | ONB-12 `can_update` hint equals the matrix cell for every role/flag combination and never changes the `PUT` outcome (hint true + revoked between read and write → `forbidden`) | **new** | display hint vs authority |
| | `http_test` ONB-H1..H8 (one per pair): closed codes, strict body, CSRF, origin binding, 405/404 closure, Group-role-only negative and `org_member` positive, list bounds, equivalence pin of the copied scope table to `branding/api`; **ONB-H9** `TestOnboardingBodyBudgets` (§5.3: budgets recomputed from the tenancy maxima; max-field raw and fully escaped accepted; 4,001-code-point description rejected by validation; `maxProfileBody + 1` rejected by the budget; `maxBody + 1` on the 8 KiB routes rejected) | **new** | adapter composition; D-3 positive coverage; the whole-payload budget admits the field contract (r2) |
| | unit `onboarding-client.test.mjs` (shapes, `malformed`, `successBytes` per endpoint), `onboarding-forms.test.mjs` (error association; the conflict/unknown status pages never claim creation; `PROFILE_FORM_BYTES` arithmetic, max-field all-astral form accepted and CRLF-normalized, `PROFILE_FORM_BYTES + 1` → 400), `session.test.mjs` (default 16 KiB unchanged; explicit `maxBytes` honoured) | **new** (pattern `mcp-connections.test.mjs`) + **extend** | BFF shape validation, honest wording and the form budget boundary (r2) |
| | E-ONB-1 sign-in (seeded admin) → New Group → suggested slug → edit → create → onboarding → new profile → complete; E-ONB-2 resume an incomplete Group, select existing profile (seeded owner); E-ONB-3 both paths to the shared editor show the same record; member sees read-only; E-ONB-4 (superseding D-1) a seeded `org_member` sees **New Group**, creates and completes its own Group with a new profile, sees no other member's Group in the list or by direct URL (status page, nothing changed); a Group-role-only seeded user (no Organization binding) sees no create action and a direct POST → status page, nothing created; E-ONB-5 keyboard-only run of E-ONB-1 + axe at 1280/320 (pattern M-A11Y); E-ONB-6 double submit → one Group, the second answer is the authorized-conflict page; E-ONB-7 no MCP context change (connection context unchanged across navigation, via the existing connections page) | **new** `onboarding.spec.ts`; S-LANDMARKS/S-NAV **extend** | end-to-end UX, accessibility, security of the thin adapter |
| (5) RLS/audit/queries/integration | `TestRLSIsolation` oracle r3: four INSERT policies' admitted/denied rows per context (owner/admin/member/group roles; selected vs Organization-only; foreign; depth-0 binding; unlinked/unrelated/member-only profile; FORCE-bound owner), r3-U update refusal for member/unlinked, DELETE still denied on associations | **extend** `rls_expected_test.go` | database-level least privilege independent of the facade |
| | `TestIsolationAttacks` r2: ATK rows for creation/completion via the new routes and runtime SQL (forged creator selector ignored; direct binding insert; new Group selecting p3; completed Group unlink attempt → 42501) | **extend** `attack_expected_test.go` | attack-suite coverage of the new surface |
| | AT-INV (dynamic) and `TestCatalogParity` (operations 36) | **extend** pins | audit invariants |
| | `TestAccessFootprint` E-PG rows for create/complete (both)/lists (statement count, locks) → PG-009 | **extend** | fan-out/lock regression |
| | importer/route pins: `TestAccessBoundary` (methods +7, `inContextTx` 17, definer allowance 00014 = 1), `TestProductionImporters` (+1), `TestPackageBoundary`/`TestSQLBoundary`, B-TENANCY-HTTP (−2/−2 probes, 401 pin), `WEB_ROUTES`, logrecord pins | **extend** | closed surfaces |

### 6.3 Batches — focused executable selectors (C5); no default full local profile

Selectors are the actual entry points; a selector that selects zero tests is a batch failure, not evidence.
Go `-run` patterns select subcases with `/` where noted.

- **b1** (unit, no database; README `platform:check` core line = B-UNIT-CORE whole-package run; the `-run` form is the
  local focus): `go test ./internal/tenancy/... ./internal/audit/... ./internal/onboarding/... ./internal/branding/...
  -run '^(TestServiceRejectsBeforeSQL|TestPackageBoundary|TestSQLBoundary|TestAccessBoundary|TestProductionImporters|
  TestRLSBoundary|TestOracleTables|TestOracleMatrix|TestOracleManageable|TestOracleGates|TestPolicyGates|
  TestPolicyScope|TestCatalogParity|TestMigrationParity|TestOnboardingPageBounds|TestOnboardingHTTP|
  TestOnboardingCopyIsExact|TestDomainsCopyIsExact|TestOnboardingBodyBudgets)$'` plus `go test ./migrations/ -run
  '^TestMigrationSet$'` and `go test ./internal/migrate/ -run '^TestNormalSetIsTheReviewedSet$'` (working directory
  `core/`, module `core`); the logrecord pin (r2, review item 4): working directory **`infra/validation/tools/`**
  (module `validationtools`), `go test ./internal/logrecord/ -run '^(TestCatalogueMatchesTheBrief|TestClassifyAccepts|
  TestClassifyRejects|TestValidate|TestCanaries)$'` — the five tests that read `events.json`/`completion.schema.json`
  (route enums, event, operations, capabilities); the package has ten tests and no database, so the whole package
  (`go test ./internal/logrecord/`) is the equally bounded alternative when a selector would select zero tests. The
  r1 path `./infra/validation/tools/internal/logrecord/...` was wrong from that module and is withdrawn.
- **b2** (runtime PostgreSQL, `rls_go`, harness rows **B-ONBOARDING-CATALOG** and **B-ONBOARDING-IT** in
  `onboarding_checks()` on the `domains_checks` template with `it_count` subtest pins):
  `./internal/migrate/ '^TestOnboardingCatalog$'`; `./internal/tenancy/access/ '^TestOnboarding$'`. Affected existing
  rows rerun with **pinned subcase selectors** (r2, review item 4; Go subtest names with spaces are matched with
  `_`, e.g. `-run '^TestRLSIsolation$/^RLS-SCOPE'`), each with its changed input; nothing else in those suites runs by
  default:
  - B-RLS-CATALOG `./internal/migrate/ -run '^TestRLSCatalog$/^(CT-RLS-2|CT-RLS-3|CT-RLS-5)'` — policy inventory and
    expressions, ACLs/functions, runtime denials (the four INSERT grants and the reader change exactly these);
    CT-RLS-1/4/6 and MG-RLS-* have no changed input.
  - B-RLS-IT `./internal/tenancy/access/ -run '^TestRLSIsolation$/^(RLS-SCOPE|RLS-CTX|RLS-ORG|RLS-IMMUT|RLS-PROJ)'` —
    oracle r3 rows (SCOPE), context variation incl. the reader context negative (CTX), fixed-actor Organization
    rows (ORG), the mutation negatives that version 14 turns into policy-gated inserts (IMMUT), helper/reader
    projections (PROJ). RLS-ATTEST/RLS-AUTH, `TestRLSPool` and `TestRLSFootprint` are **not** oracle r3 targets and are
    not rerun unless slice 1's grep of their statements finds one of the four tables' INSERT/DELETE or the helper
    functions (the Result records the grep and the final selection).
  - B-RLS-ORDERING `'^TestRLSOrdering$'` only if a lock statement text changes.
  - access rows: `./internal/tenancy/access/ -run '^TestAccessOrdering$/^(O-ONB|O-13|O-14|E-PG-3)'` (new cases
    O-ONB-1..3/D/E; O-13 trusted unlink, O-14 both-order writers and E-PG-3a/b exercise the changed `ReadFacts` lock
    path); `-run '^TestAccessFootprint$/^(E-PG-1|E-PG-2)'` (the E-PG rows extended there — **not** `TestRLSFootprint`);
    `-run '^TestIsolationAttacks$/^(ISO-ATTEST|B|C|ISO-CITATIONS)'` (B1 shared-profile path, C2 retained context after
    membership removal, new rows B2/C3 and their citations); `-run '^TestAccessPolicy$/^(IT-CTX-6|IT-MATRIX-1|
    IT-PROFILE-|IT-FLAG-4|IT-FACADE-)'` (matrix +2 capabilities, profile/manageable cases, facade allow/deny for the
    seven new methods, mapping, payload rules); `-run '^TestAudit$/^(AT-FMT|AT-ATOM|AT-INV)'` (trigger attribution,
    same-statement T4/T7, catalog actions from real calls). Focused loops: `-run '^TestAccessOrdering$/^O-ONB'`.
  - **Wider covering batch (bounded, for supervisor disposition, not default):** if slice 1 changes the semantics of
    a shared reader that every policy consumes — i.e. `app.rls_org_role()`/`app.rls_group_role()` bodies or the
    `role_bindings_select`/`*_policy_reader` policies (not planned; the reader only *calls* `rls_org_role()`) — the
    concrete guarantee "every existing RLS oracle row still holds under the changed helper" needs the whole
    `'^TestRLSIsolation$'` once (one invocation, ≈ the existing B-RLS-IT row). Changed grants alone do not trigger it.
  - version pins `./internal/tenancy/ '^(TestTenancyModel|TestBusinessProfileModel)$'`,
  `./internal/migrate/ '^TestMigrationLifecycle$'`, `./internal/tenancy/access/ '^TestRecoverySeed$'` and the shared
  MCP fixture through one existing entry point `./internal/mcp/ '^TestMCPTransport$'` (the fixture at
  `env_integration_test.go:253` has no entry point of its own). Rows whose inputs do not change (B-OBJ-*, B-AUDIT-*,
  B-RETENTION-*, B-MCP-CATALOG, B-DOMAINS-*, B-BRANDING-*) are **not** rerun without a changed input.
  - **Restore boundary (r4):** R-ONB-1 runs inside the B-ONBOARDING-CATALOG selector above (`^TestOnboardingCatalog$`
    selects its subcases); the local loop is `./internal/migrate/ -run '^TestOnboardingCatalog$/^R-ONB-1'`.
    **B-RECOVERY-DRILL is not rerun locally** and no restore campaign is added: its inputs (`infra/recovery/**`,
    `TestRecoverySeed`/`TestRecoveryVerify`) do not change; `TestRecoverySeed` is already in the version-pin list.
    The drill belongs to the `migrations` suite that the path-based staging gate runs for `core/migrations/` after an
    authorized conditional merge, so the real source→target logical restore of the version-14 schema (distinct system
    identifiers, verbatim rows, fingerprint equality) is collected there as existing hosted evidence; the Result
    cites it, and a drill failure on the provenance columns is a G018 finding, not a recovery change.
- **b3** (web unit): `npm run test:unit` (B-UNIT-WEB; whole suite, seconds).
- **b4** (browser, identity mode): `IDENTITY_E2E=1 npx playwright test --project=onboarding-light
  --project=onboarding-dark --project=shell-light --project=shell-dark --grep 'E-ONB|S-LANDMARKS|S-NAV|S-A11Y'`;
  `BROWSER_MATRIX_IDENTITY_E2E` extended; B-TENANCY-HTTP and B-WEB-COMPONENTS via their harness rows.
- **Required final local coverage before the MR** (AGENTS 2026-10-16 rule): the union of affected checks only —
  B-UNIT-CORE, B-UNIT-WEB, B-ONBOARDING-CATALOG/IT, B-RLS-CATALOG, B-RLS-IT, the access rows above, B-MCP-IT's
  fixture entry point, B-TENANCY-HTTP, B-WEB-COMPONENTS, the browser rows for the onboarding/shell projects,
  B-SCOPE/B-REQUIRED-CHECKS (final) — each with changed-input provenance in the Result. **No default `full` profile and
  no whole legacy oracle campaign.** A broader run needs its coverage reason and a supervisor disposition. Hosted
  evidence: the MR fast pipeline runs automatically; the path-based staging gate (foundation rows for the two web
  areas, migrations suite for `core/migrations/`) supplies the required hosted profile after an authorized conditional
  merge; its security gates are retained. CI meta-tests (`validate-controls.sh`) only if a count pin or control logic
  changes (Q4: mapping edits do not justify the campaign).

### 6.4 Manual review (M1), readiness and the draft MR

Same preview environment as G017 §6.4 (L2 emulator overlay, `IDENTITY_PROXY_PUBLISH=127.0.0.1:543`, seeded users,
run CA in a dedicated browser profile; no WorkOS, no hosts-file edit). Readiness: candidate SHA, seed digest, CA path,
`browser/users.json` pointer; the operator confirms readiness before any timer. Feedback items: New Group with
suggested/edited slug; a duplicate slug (the authorized-conflict page); resume; both completion branches; paging with
more than one page of Groups; Groups and Business Profiles sibling pages; the shared editor from both entry points;
member read-only; Light/Dark/System at 1280 and 320; keyboard-only completion; error association on an invalid website.
Findings → fixes → rerun affected E-ONB/S-*/unit selectors → repeat M1 for changed surfaces only; cleanup
`identity:stop` (`down -v`) with the owned-resource receipt. **Draft MR opens when slice 2 compiles and b1/b2 pass**;
M1 completes before final local qualification, Codex pre-integration acceptance and any merge.

## 7. Delegation

One implementer owns core and web (overlapping edits, one write owner). Two optional read-only reviewers: R-A
(creation/completion authorization, the trigger/policies/reader, orderings) after slice 1; R-B (UX/accessibility of
slices 3–4) after slice 4. Delegated instructions carry the task-prefixed worktree rule (generated `agent-<id>`
accepted) and the read-only boundary; Claude validates and integrates all work and keeps the one B7 ledger.

## 8. Sequence

0. R0 after G017 integration (§0.3); rebind; Codex Plan ACCEPT (incl. D-3/D-9 r1 dispositions);
   allow-list import; execution release.
1. **Slice 1 (riskiest)**: migration 00014 + `TestOnboardingCatalog`; two capabilities, matrix, `rls_expected` r3;
   `tenancy/onboarding.go` Tx writers + `facts.go` lock partition + boundary pins; guarded `CreateGroup`, both
   `Complete*`, `GetGroupOnboarding`; ONB-1..9 (incl. the member positive and 9e), ONB-13, O-ONB-1..3, O-ONB-D,
   O-ONB-E; version-14 pins; audit operations; the §6.3 subcase-selection grep. b1, b2.
   R-A review.
2. **Slice 2**: `ListGroups`, `ListBusinessProfiles` (paged), `SuggestGroupSlug`; `onboarding/api` + `http_test`;
   server registration; logrecord enums; importer/route/B-TENANCY-HTTP pins; `onboarding_checks()` rows; ONB-10..12;
   attack r2; footprint rows. b1, b2. **Open the draft MR.**
3. **Slice 3**: `onboarding-client.ts`, fields, pages/routes (incl. the conflict/unknown status pages and paging),
   shell area links; unit tests; `WEB_ROUTES`, `profile-paths`, B-WEB-COMPONENTS. b1, b3.
4. **Slice 4**: seed additions; `onboarding.spec.ts`, projects, matrix pins; S-LANDMARKS/S-NAV; b4. R-B review.
5. **Slice 5**: docs (TENANCY, AUTHORIZATION, DATABASE-ISOLATION, AUDIT, MIGRATIONS, WEB-SHELL, FOLLOW-UP, RECOVERY,
   SECURITY, PG-009); M1; reconcile; required final local coverage (affected only, §6.3); Result; safekeeping push;
   MR to ready after Codex pre-integration acceptance; checkpoint review after attempt 4 if reached.

## 9. Acceptance-criteria mapping

| AC | Evidence |
|---|---|
| AC1 | ONB-1..3 (incl. the `org_member` positive of ONB-1, 2a/2b/2d and 3b), ONB-13, CT-ONB-1 (incl. the `groups.id` grant), matrix/policy pins (15), `rls_expected` r3 rows for `groups_insert`/`role_bindings_insert_bootstrap` (all three Organization roles admitted; Group-role-only and selected-Group refused); §5.6 items 1–3 (no unique-owner constraint, no second binding); AUTHORIZATION rows; E-ONB-4 |
| AC2 | IT-SLUG-1..6 + `parity_test` (reuse), ONB-4 (guarded conflict, cross-tenant independence, rename keeps UUID/binding/association), E-ONB-1 (suggested then edited slug through the real form) |
| AC3 | ONB-5..8, O-ONB-1..3, O-ONB-D, **O-ONB-E**, hookDB fault seams; AT-INV; §1.4 ordering text in AUTHORIZATION |
| AC4 | ONB-9 (facade, runtime SQL and the direct reader calls 9e), ONB-13, extended `TestAccessPolicy`/`TestOracleManageable`, `rls_expected` r3 profile rows (incl. the UPDATEd-committed-profile refusal), attack rows B2/C3, CT-ONB-1 creation-proof pins, E-ONB-2 |
| AC5 | §5.5 inventory (documented in TENANCY/AUTHORIZATION), `rls_expected` DELETE-denied rows, attack unlink row, ONB-8/11 raced/stale cases, E-ONB-2 resume without false completion; no DELETE grant (CT-ONB-1) |
| AC6 | ONB-10 (traversal, member-only absent, selected-context restriction, flag off), ONB-12, existing read-contract cases, E-ONB-3 (both entry points, member read-only), canary absence of sibling data in HTTP bodies (`http_test`) |
| AC7 | ONB-11 (duplicate/lost-response/changed-payload; authorized conflict, never "created"), E-ONB-6, `http_test` conflict mapping, `onboarding-forms` wording test, web recovery read (§1.4) |
| AC8 | E-ONB-1..7, S-LANDMARKS/S-NAV/S-A11Y extensions, `onboarding-client` unit, M1 record; no signup/invite/product flow; E-ONB-7 |
| AC9 | CT-ONB-1, `TestAccessBoundary`/`TestProductionImporters`/route pins, B-TENANCY-HTTP revision with positive pin (D-3), `TestCatalogParity` 36 operations, `TestAccessFootprint` E-PG rows → PG-009, SECURITY §5 row for the runtime role updated; ONB-H9 and the form-budget unit cases (whole-payload budgets admit the field contract; nothing shrinks); no new role/tier/vendor/framework; one new SECURITY DEFINER reader (boolean, policy-reader-owned) and one SECURITY INVOKER trigger, both pinned |
| AC10 | Result with commands/outcomes/provenance (Git-first), M1 record, MR/staging disposition, cleanup receipt; Codex ACCEPT separate from merge |

## 10. Decisions, risks and escalations

| ID | Item | Recommendation / status |
|---|---|---|
| **D-1** | Which Organization roles hold `group.create`. **History:** the owner first answered "Yes" to Codex's recommendation (a) `org_owner` + `org_admin` only on 2026-10-17 (Spec §"Owner decision D-1", staging `2d535708`; r1 applied it). **Superseded the same day** (Spec §"Superseding D-1 — Generic self-service Group onboarding", staging `efd4295c`, SHA-256 `784feedb…2e653`): a Beginify Group is a client account or team a user signs up for and sets up, so `group.create` admits every current Organization membership role **`org_owner`/`org_admin`/`org_member`** in trusted Organization-only context; the verified creator becomes the initial `group_owner` atomically; no actor selector, Organization-role grant, extra owner store or other-Group privilege; a Group role without current Organization membership is insufficient; existing-profile selection keeps needing pre-existing eligibility. r2 applies it to the matrix (§1.3), `groups_insert`/`role_bindings_insert_bootstrap` (`R IS NOT NULL`, §4.4), the member-safe keyed insert + read (S-G1/S-G2, §4.2/§4.7, reinstating the r1 `f3f1ed1e` (b) delta under review), §5.3, §5.6 item 1, ONB-1/2/13, E-ONB-4. Bootstrap, profile-insert, lock and paging designs unchanged. Public signup, enrollment and membership provisioning stay in G042/G043; no public endpoint, implicit membership or staff privilege is added. Resolves D-1 only — not Plan acceptance or execution release | **RESOLVED (superseding) — owner** |
| D-2 | Completion representation: derived from the association (no `onboarding_completed_at` column) | acceptable in principle (review) |
| **D-3** | First HTTP exposure of tenancy mutations: retire `/internal/v1/groups` and `/internal/v1/business-profiles` from `TENANCY_PROBE_CORE` and `/groups`, `/business-profiles` from `TENANCY_PROBE_WEB`; keep the remaining 16/10 probes; add the positive closed-route coverage of §5.3 and the 401 pin in B-TENANCY-HTTP | accepted in principle, conditional on that coverage (review) |
| D-4 | Group rename/slug-change over HTTP: **not** in G018 (existing guarded ops stay Go-only; AC2 rename evidence at the facade) | acceptable in principle |
| D-5 | Session-scope helper: bounded byte-pinned copy of P-1 (third copy) with the equivalence test (`TestOnboardingCopyIsExact` against `branding/api`); reuse an accepted shared helper instead if G017 supplies one at R0; extraction is not a prerequisite | accepted (review) |
| D-6 | Web placement: foundation sibling areas under `web/app/{groups,business-profiles}` with two `profile-paths` foundation rows, not a G051 feature area | acceptable in principle |
| D-7 | No standalone profile creation in the Business Profiles area (creation only through onboarding; prevents orphan profiles, matches ADR-004) | acceptable in principle |
| D-8 | Form controls: kit-adapted plain `input`/`textarea`, native label/fieldset/radio; **no** `@radix-ui/react-label`/`react-radio-group` dependency | acceptable in principle |
| **D-9 (redesigned)** | r0's SECURITY DEFINER count reader `app.rls_group_binding_count` is withdrawn (subject-filtered through `role_bindings_policy_reader`, hence blind to other owners, and a count cannot serialize concurrent inserts). r1 bootstraps the creator's binding from a SECURITY INVOKER AFTER INSERT trigger admitted by a `pg_trigger_depth() > 0` policy (§4.3–4.4), and enforces profile eligibility with one boolean SECURITY DEFINER reader `app.rls_profile_eligible` in the 00006 policy-reader pattern (no data returned, helper policies bounded to `ctx_org()`). Neither is a SECURITY DEFINER mutation surface, a new trust channel or a broad grant; if Codex judges the trigger-admitted policy or the `created_xid`/`created_sysid` insertion-provenance pair (r3/r4; two defaulted, immutable, non-granted columns on `business_profiles`, reversible on down) a material architecture change, it is escalated rather than implemented | security review item (r1; column r3; cluster term r4) |
| **D-10 (r4)** | Restore-namespace closure, review r3: **option (a) cluster-aware provenance** — `created_sysid bigint DEFAULT (pg_control_system()).system_identifier`, compared with the live identifier in reader branch (a) — chosen over **option (b)** restore/reopen reconciliation. (b) rejected because rewriting or invalidating `created_xid` during the restore would make `recovery.sh:621–625`'s data-fingerprint diff fail unless the column were excluded from `fingerprint-data.sql` (a weakened integrity check), would add a trusted data mutation to the recovery path, and would protect only restores performed through the script; (a) needs no recovery change, keeps every source/target check and fingerprint byte-identical in effect, and holds for a manual `pg_restore` too. Rejected variants of (a): a wrapper function `app.cluster_sysid()` (adds a function-allowlist entry for no gain; the composite field access is a valid column default and reader expression) and a SECURITY DEFINER wrapper (would change the 00014 definer allowance; only needed if `pg_control_system()` were not PUBLIC-executable, which precondition 1 pins instead). Residual, recorded not addressed: `pg_resetwal` moving a cluster's transaction counter backwards is an OS-level destructive operation outside the supported recovery path and the model | proposed for Codex disposition (r4) |
| Q3 | `can_update`: display hint from a second decide on the same snapshot; the mutation always reauthorizes; no sibling-visibility implication | accepted (review) |
| Q4 | B-ONBOARDING-* rows and the probe-list edit are mapping changes; no CI-control campaign; actual control-logic changes are escalated separately | accepted (review) |

**Escalation risks (stop, do not design around):** G017 drift at R0 (shell navigation, scope helper, seed, harness
row pattern); any widening of creation beyond current Organization membership (public signup, enrollment, implicit
membership, a Group-role-only or cross-Organization creator — G042/G043 owner decisions; never a silent change); a
RETURNING/visibility refusal of S-G1/S-G2 that the pinned grants cannot explain; any need for a status column with its own write path, an unlink/replace route, a standalone
profile creation route, a new package/font/primitive, an MCP tool, or a change to `role_bindings` UPDATE/DELETE
policies (all out of scope or G019); a 42P17 policy cycle, any path by which `created_xid` or `created_sysid` can be
written or changed at runtime or the replaced immutability body not taking effect (would void the insertion proof;
stop, do not widen the trigger silently), `pg_control_system()` found not executable by the runtime or policy-reader
role in the pinned image (precondition 1 refuses; the SECURITY DEFINER wrapper alternative changes the definer
allowance and needs review — stop, do not grant around it), any need to touch `infra/recovery/**` or the fingerprint
SQL to make R-ONB-1 or the hosted drill pass (would be option (b) by the back door; stop), an `ADD COLUMN` rewrite
that cannot complete within the migration transaction's normal bounds, a lock-order cycle or unexpected scan in b2; an ownerless legacy Group appearing in
seeds/fixtures (the trigger never re-claims it; recorded residual for G019's reconciliation); budget exhaustion or
context overrun (handoff at 300k).

## 11. Budget, forecast, pins and scope

### 11.1 Budget — B7, checkpoint after attempt 4; forecast 3 attempts; unspent.

### 11.2 Count pins (provisional at `862db18d`/`e9d35f3f`; re-pin at R0 against integrated G017; reconciled, C5)

Migrations 13 → **14** (all P-4 sites; relcounts 50/48 unchanged; forced-RLS table count unchanged); capabilities 13 →
**15**; matrix cells 286 → computed and recorded (+22 for `group.create`, three Organization-role cells `Y` under the
superseding D-1; Group-kind cells per the oracle's formula); runtime `groups` INSERT columns **4** (`id` included) and
policy-reader `business_profiles` columns **4** (`id, organization_id, created_xid, created_sysid`; no system
column); runtime UPDATE columns on `business_profiles` **5** and INSERT columns **7** (neither names a provenance
column); `business_profiles` columns 8 → **10** (`created_xid xid8`, `created_sysid bigint`; r4), relation count
unchanged, trigger inventory unchanged, function allowlist unchanged by r4 (no wrapper function), one replaced
trigger-function body (hash pinned); `has_function_privilege(<role>, 'pg_control_system()', 'EXECUTE')` **true** for
the runtime and policy-reader roles (pinned, r4); JSON budgets **8 KiB / 96 KiB**, `PROFILE_FORM_BYTES` 96 KiB, `MAX_FORM_BYTES` 16 KiB
unchanged; `web/lib/session.ts` exported signatures unchanged except the optional `readForm` second parameter;
`access.denied` operations 29 → **36**; audit actions **36** and reasons **58** unchanged; policies +4 runtime/owner
and +2 helper; SECURITY DEFINER functions +1 (00014 allowance = 1; `rls_profile_eligible`), SECURITY INVOKER trigger
functions +1 (`groups_bootstrap_owner`), triggers on `groups` +1; `inContextTx` call sites 11 → **17**; exported
access methods +7; tenancy production files 14 → **15**, package functions +5; importers of `access` +1
(`onboarding/api`); `required-checks.tsv` 119 at `e9d35f3f` (+2 expected from G017) → **+2** (B-ONBOARDING-CATALOG,
B-ONBOARDING-IT); `completion.schema.json` route enums **+6 paths** (two occurrences); `events.json` +1 event, +7
operations, +2 capabilities; HTTP **8 method/path pairs** on those 6 paths; `WEB_ROUTES` **+9**; B-TENANCY-HTTP
probes −2 core/−2 web, +1 positive 401 pin; `components/ui` +2 (`input`, `textarea`) with their `data-slot` values;
Playwright projects +2 (`onboarding-light/dark`); `profile-paths.tsv` +2 rows; seed +4 bindings, +3 profiles, +2
associations; `SEMANTIC_COLOR_TOKENS`, fonts, `web/public`, `appearance-init` unchanged.

### 11.3 Scope allow-list — `management/v0/plans/G018-scope-allowlist.txt` (r2; literal `[.]` dots; Codex imports it
as `management/v0/results/G018-scope-allowlist.txt`): 43 executable patterns, 55 lines with comments; SHA-256
recorded in the hand-back and §13. **r4 adds no pattern** (the file is byte-identical to r3, SHA-256
`5a94f3ac6cc585bfd437e3b5cf7b577bec147ca0c2d3d816d9e8e8d267a84a7e`; its header comment still reads "r3" because no
byte changes). The cluster term is carried entirely by already-listed patterns, enumerated here so the absence of
additions is reviewable rather than silent: `^core/migrations/00014_group_onboarding[.]sql$` (the second column,
the trigger tuple, the reader term, precondition/postcondition/down); `^core/internal/migrate/onboarding_integration_test[.]go$`
(CT-ONB-1 pins and the R-ONB-1 subcases); `^core/internal/migrate/business_profile_integration_test[.]go$`
(column/column-privilege inventory 10/4); `^core/internal/tenancy/access/rls_expected_test[.]go$` (the oracle's
reader-column projection row, if it enumerates columns); `^docs/foundation/(…|DATABASE-ISOLATION|…|RECOVERY)[.]md$`
(the restore boundary). Explicitly **not** added and not needed: any `infra/recovery/**` pattern (`recovery.sh`,
`sql/fingerprint-data.sql`, `sql/fingerprint-catalog.sql`, `testdata/`), `infra/validation/run.sh` beyond the two
rows already listed (no drill rerun), and no grant or integrity-check relaxation anywhere. r3 added exactly one
pattern for the provenance column (§4.2):
`^core/internal/migrate/business_profile_integration_test[.]go$` (`TestBusinessProfileCatalog` column and
column-privilege inventories of `business_profiles`, if evaluated at head — R0 confirms; the pattern is listed so the
consequence is reviewable, not silent). r2 (SHA-256 `70eff624fccd0d9159297e451b4c44abdfe9503656daffc7ad22d8aadcfbb22b`,
archived as `G018-scope-allowlist-r2.txt`) added exactly two patterns for the form budget (§1.6):
`^web/lib/session[.]ts$` (the optional `maxBytes` parameter; default and existing callers unchanged) and
`^web/tests/unit/session[.]test[.]mjs$`; nothing is removed in r2 or r3. r1 (SHA-256 `ef4ce188f7c89bf29c3cb33de4ebb4c292748aec6be55ab83f91c65119baae3c`) added
`core/internal/tenancy/(facts|store).go`, `core/internal/tenancy/business_profile_integration_test.go`,
`core/internal/migrate/rls_integration_test.go`, the access `onboarding_test.go` unit file and
`web/tests/unit/onboarding-*.test.mjs`. r1 archived as `G018-scope-allowlist-r1.txt`, r0 as
`G018-scope-allowlist-r0.txt`. The superseding D-1 needs no new file: the keyed Group statement lives in the
already-listed `core/internal/tenancy/(store|onboarding).go`.

## 12. Open questions for Codex

1. (Closed, superseded) D-1 is resolved by the owner's superseding decision (`efd4295c`): all current Organization
   membership roles create; §5.6 item 1 and §4.7 S-G1/S-G2 are settled on it. Codex review point: the reinstated
   keyed insert/`id` grant and read-after-trigger sequence (§4.7) as the member-safe sequence under actual grants.
2. D-9 r3 disposition: the SECURITY INVOKER bootstrap trigger with the `pg_trigger_depth()` policy predicate, and the
   "inserted by this transaction" predicate `created_xid = pg_current_xact_id()` (§4.2 insertion-provenance column,
   §4.4 (a); the r2 timestamp form is withdrawn) inside the boolean reader, which validates context through
   `app.rls_org_role()` itself (closed in r2). Specifically for review: two added, defaulted, immutable,
   non-granted columns on `business_profiles` (`xid8` r3, `bigint` cluster identifier r4) and the replaced
   `business_profiles_keys_immutable` body.
6. (r4) D-10 disposition: option (a) cluster-aware provenance (`created_sysid` = `pg_control_system().system_identifier`,
   compared live in reader branch (a)) as the closure of the logical-restore namespace change, with no recovery
   change, no fingerprint exclusion and no scope addition (§11.3 lists the carrying patterns); R-ONB-1 as the
   deterministic restored-row/target-id collision negative plus fresh-insert positive; confirm that the PUBLIC
   executability of `pg_control_system()` pinned in precondition 1/CT-ONB-1 is an acceptable dependency (the
   alternative is a SECURITY DEFINER wrapper, which would move the definer allowance pin — not proposed).
5. §6.3's pinned subcase selectors and the bounded wider-batch rule (whole `TestRLSIsolation` only if a shared
   helper/policy body changes): confirm as the local-loop and final-coverage selection; the Result records the slice 1
   grep that finalizes the RLS-ATTEST/RLS-AUTH/`TestRLSPool`/`TestRLSFootprint` exclusion.
3. D-3 r1 disposition: the positive coverage of §5.3 as the replacement for the four retired absence probes.
4. Whether the ascending-id partition in `ReadFacts` (§1.4) may be applied to the shared reader as a general rule now,
   or only to the combined shape (both are bounded; the general rule is what §5.6 item 7 settles for G019).

## 13. Revision history

| Rev | Date (UTC) | Base | Change |
|---|---|---|---|
| r0 | 2026-10-17 | `862db18d` (G017 draft MR !40 head) on `g018-plan` | initial provisional Plan; allow-list r0 SHA-256 `8e40e9cb5b84a200c07d4284ff536db923e0c5980f710d6953c252d8137013e0` |
| r1 | 2026-10-17 | r0 commit `b8c8b8a8` + `origin/staging` `b4250a23` on `g018-plan-r1` (G017 observed at `e9d35f3f`) | bounded correction under the Codex review r0 (SHA-256 `edc1a240…63f0`): C1 trigger bootstrap replacing the count reader, C2 database eligibility reader, C3 executable statement sequences and corrected §5.5, C4 ascending-id Group lock order with O-ONB-D, C5 actual inventory/selectors/focused validation and reconciled counts, C6 bounded keyset lists and honest conflict/unknown recovery; D-3/D-5/Q3/Q4 dispositions applied; D-9 redesigned; D-1 left open at first commit `f3f1ed1e`. r0 archived as `…-plan-r0.md` (SHA-256 `c7a64649…74c59`) and `G018-scope-allowlist-r0.txt`; allow-list r1 SHA-256 `ef4ce188f7c89bf29c3cb33de4ebb4c292748aec6be55ab83f91c65119baae3c` |
| r1 + D-1 | 2026-10-17 | r1 `f3f1ed1e` + `origin/staging` `2d535708` (management-only) | bounded delta only: D-1 recorded as owner-resolved (a) `org_owner`/`org_admin` in §0.1/§1.1/§1.3/§4.2/§4.4/§4.7/§5.3/§5.6/§6.2/§8/§10/§12; the (b) keyed-insert alternative dropped; §2 rebound to the Goal/Spec/review D-1 addenda at `2d535708`. The six corrections and the scope allow-list are unchanged (scope SHA-256 unchanged) |
| r2 | 2026-10-17 | r1 `0eed3ad5` + `origin/staging` `68cb0d1e` (r1 review) and `efd4295c` (superseding D-1), both management-only, on `g018-plan-r2` (worktree `agent-<id>`) | bounded correction under the Codex review r1 (SHA-256 `ec3855ed…f35b45`): (1) creation proof = `xmin` + `created_at = transaction_timestamp()` with the `created_at` ACL pin, O-ONB-E; (2) the reader validates context via `rls_org_role()` itself, ONB-9e, attack C3; (3) computed whole-payload budgets 96 KiB (profile JSON routes and forms) with `readForm` `maxBytes`, ONB-H9 and form-boundary units; (4) tools-module working directory/path, `TestRLSFootprint` de-cited, pinned subcase selectors, bounded wider-batch rule. **Superseding D-1** (Spec `efd4295c`, SHA-256 `784feedb…2e653`) reconciled: `org_member` creates; member-safe keyed insert without RETURNING + read (S-G1/S-G2, `groups.id` grant); `R IS NOT NULL` policies; ONB-1/2/13, E-ONB-4; §0.1/§10 keep the superseded (a) answer as history. r1 archived as `…-plan-r1.md` (SHA-256 `e9ac8c83…e05eb`) and `G018-scope-allowlist-r1.txt` (SHA-256 `ef4ce188…aae3c`); allow-list r2 adds `web/lib/session.ts` and `web/tests/unit/session.test.mjs` (SHA-256 `70eff624…b22b`) |
| r3 | 2026-10-17 | r2 `087b2e4c` + `origin/staging` `4eef5983` (r2 review; management-only) on the same branch `g018-plan-r2` (worktree `agent-<id>`) | bounded correction under the Codex review r2 (SHA-256 `423f8107…26ebc`): (1) the timestamp element of the creation proof withdrawn; insertion provenance = `business_profiles.created_xid xid8 NOT NULL DEFAULT pg_current_xact_id()` (no runtime grant, immutable by the replaced `business_profiles_keys_immutable` body, reversible on down), reader (a) `created_xid = pg_current_xact_id()`, no `xmin`/system-column grant; O-ONB-E retained and extended with the synthetic equal-timestamp negative (c); CT-ONB-1/§4.1/§4.5/§4.6/§11.2 pins; scope +1 `core/internal/migrate/business_profile_integration_test.go`; (2) O-ONB-D fixture corrected to two incomplete targets g1/g2 plus the qualifying attached g3/p3, both targets in both candidate sets, lock design and interleavings unchanged. r2 archived as `…-plan-r2.md` (SHA-256 `120625d1…30c4`) and `G018-scope-allowlist-r2.txt` (SHA-256 `70eff624…b22b`); allow-list r3 SHA-256 `5a94f3ac6cc585bfd437e3b5cf7b577bec147ca0c2d3d816d9e8e8d267a84a7e` |
| r4 | 2026-10-17 | r3 `bab9d320` (`origin/g018-plan-r2`) + `origin/staging` `1131651c` (r3 review; management-only) on `g018-plan-r4` (worktree `agent-<id>`) | bounded correction under the Codex review r3 (SHA-256 `5567da3c…b2e86`): the supported logical restore (`recovery.sh:369,556/562,614–617,621–625`) moves `created_xid` into another cluster's transaction namespace; closed by **option (a)** cluster-aware provenance — `business_profiles.created_sysid bigint NOT NULL DEFAULT (pg_control_system()).system_identifier`, immutable by the same replaced trigger body, in the policy reader's column grant (4) and in reader branch (a) as the pair `(created_xid, created_sysid) = (pg_current_xact_id(), (pg_control_system()).system_identifier)`; no recovery/fingerprint/grant change, no new function, no scope pattern (§11.3 lists the carrying patterns); R-ONB-1 negative/positive subcases of CT-ONB-1; §6.3 drill note; D-10 and the `pg_control_system()` executability pin/escalation. r3 archived as `…-plan-r3.md` (SHA-256 `963a97ca…48c3e`); allow-list unchanged (SHA-256 `5a94f3ac…a7e`) |
| r5 | 2026-10-17 | r4 `898c7efd` + `origin/staging` `091be3d6` (r4 review; management-only) on the same branch `g018-plan-r4` (worktree `agent-<id>`) | one fixture paragraph under the Codex review r4 (SHA-256 `2f741b02…6532`): R-ONB-1's restored-row collision inspection and UPDATE moved to the superuser fixture connection, then `SET LOCAL ROLE` runtime with u's context for the reader/association denial in the same transaction; positive's post-commit provenance read on the privileged connection. Design, SELECT policy, grants, schema, scope and selectors unchanged. r4 archived as `…-plan-r4.md` (SHA-256 `55bc7821…b583`); allow-list unchanged (SHA-256 `5a94f3ac…a7e`) |
