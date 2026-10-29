> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G007 — Trusted context and shared policy evaluation: implementation plan (r1)

- **Plan revision:** r1 (correction of r0), 2026-10-10 UTC. Author: Claude Code (planning subagent). Reviewer: Codex.
  Major decisions: product owner.
- **Status: PLAN DRAFT for Codex review. Not implementation-ready** until the gates in §2 pass. Nothing was
  implemented, built, tested, migrated, run in Docker, pushed or merged. **Zero of seven B7 cycles consumed.**
- **Governing inputs:** [Goal r3](../goals/G007-trusted-context-policy.md) (SHA-256 `a9240953…2fac`),
  [Spec r3](../specs/G007-trusted-context-policy-spec.md) (`a48118fb…0c06`),
  reassessment (not included) (`99339b5d…dea0`); ARCHITECTURE §21 and §53.1–53.7;
  ADR-003 (optional MFA, unknown assurance); ADR-004 (derived profile access). **Added in r1:**
  r0 Codex review (not included) (`06199f49…b9b5`), ARCHITECTURE §63 (file
  `ff30e55f…ad72`) and [PERFORMANCE.md PG-001](../PERFORMANCE.md#pg-001--g007-authorization-lockquery-footprint)
  (`b433ab67…9591`), read from `origin/staging` (§0).
- **Planning provenance:** isolated worktree `.claude/worktrees/agent-<id>`, branch `g007-plan` created
  at `6889273cf5eca65ebf979c6299be290ff5f61a1f` (tree `ca21d224ea719b0e476ed609bde489aa4170978b`), clean before
  writing. `origin/staging` had advanced to `8414107` (PROGRESS and SESSION-HANDOFFS only); it is not used. Every
  relied-on input's SHA-256 is in G007-planning-inputs.json (not included). This is planning provenance,
  not an execution binding.

## 0. Revision r1 (Codex CORRECTION LOOP on r0)

**Provenance.** r0 (planning commit `49cf3e368ef9eed22423b7ad61c2f0b1bc5c3f36`) was produced by the first planning
session at `high` effort; the review notes this was not independently verified as xhigh. r1 is produced by a new
planner session at the owner's xhigh setting. It continues from the committed r0, which it does not rebase, and it is
committed on top of r0 in the same worktree and branch. r0 is preserved byte-for-byte as
G007-trusted-context-policy-plan-r0.md (not included) (`7bca07fb…bbd4`) and its inventory as
G007-planning-inputs-r0.json (not included) (`554acf77…6bf2`).
- The new inputs were read with `git show` from `origin/staging` at `43af76e0a5eb259ccce4bced32db5e139821cc5b`
  (fetched for r1):
  - the review, committed in `ff32076`;
  - ARCHITECTURE §63 and PG-001, both committed in `378f870`.
- Staging's CI and session-tracking changes and the pending G051 changes are **not** adopted; E2 reconciles them.
- Planning used no implementation cycle and no Docker, build, test, migration, push or pipeline action.

**C1: trusted per-request binding.** r0 bound a context to its issuing Service and to `observability.RequestID(ctx)`.
That value is a diagnostic: it may come from a well-formed inbound `X-Request-Id`, distinct requests can share it, and
in-process callers have none. So an old valid context could pass under another request.
- r1 replaces this binding with an **authenticated request scope**. `BeginRequest(ctx, credential)` validates the
  credential through the accepted `identity.Service.Session`. It returns a `ctx` carrying a new, unique, unexported
  scope object that records the authenticated session and subject.
- `Resolve` no longer takes a credential. It takes the subject from the scope in `ctx`, and the context it returns
  is bound to that scope object by pointer identity.
- Every guarded operation, before any SQL, requires all of these:
  - the `ctx` scope exists, was issued by the same Service and has not ended;
  - the context's scope is that exact object;
  - the context's subject and session equal the scope's.

  Correlation IDs are never compared and are no longer stored in the context.
- Changed sections:
  - §1 item 1;
  - §3.2: the concrete API and the rejection conditions;
  - the §3.5 surface;
  - the §3.6 reasons and event;
  - §6.1 U-CTX, §6.3 IT-CTX and IT-REVOKE;
  - §9 AC1;
  - §12 Q9, which is now answered, and Q11.
- No route, public adapter, provider-claim input or request framework is added.

**Performance addendum (ARCHITECTURE §63, PG-001).** New §4.7 links PG-001 and covers:
- each guarded operation's statements, rows read, rows locked and lock modes, including the Organization-only profile
  write that locks all of the subject's candidate Group bindings and Groups;
- growth with Group count, and how long-held shared locks delay policy and revocation writers;
- concrete index coverage from migrations 00002–00004;
- workload assumptions, which are explicitly unresolved;
- the bounded execution evidence E-PG-1..3 (§6.7). It sets no performance target and claims no efficient plan.

The addendum yields owner-visible questions Q13 and Q14 and Codex questions Q16 and Q17; Q15 comes from the
consistency fix below. No index, schema, grant or lock-mode change is made.

**Consistency fixes forced by the above:**
- §3.6: the `access.decision` catalog represents the absent capability of `BeginRequest` and `Resolve` explicitly.
  The field is forbidden for those operations, and their success outcome is `established`, never `allow` (review
  routine item).
- §4.3/§4.4: while enumerating conflicting writers for §4.7, r1 found that r0 omitted sign-in's conditional
  `users` display-name UPDATE. It can form one rare wait cycle with a guarded write. r1 records that fact and adds the
  proposed O-15, but leaves the disposition to Q15. §4.3 also lists the brief waits of session renewal and
  provider-logout bookkeeping. No writer or lock order changes.
- §3.6's cross-reference for the lost commit reply now points to §4.5 (r0 said §4.6).
- Other sections touched to carry C1 and the addendum through:
  - §2 E3;
  - §6.4: hookDB counters, O-15 and E-PG-3;
  - §6.5: catalog records;
  - §6.6 RC-9;
  - §8: one new test file, already covered by the allow-list;
  - §9: AC1, AC8 and a PG-001 row;
  - §10 S3/S4, and §11: the C3 note and stops.

  Each such change is marked "(r1)".

**Unchanged and settled by the review:**
- the Q8(a)–(d) matrix pins;
- package placement;
- the transaction composition seam;
- the Organization boolean column and its default;
- the Group-row Unlink lock surrogate;
- REPEATABLE READ read snapshots;
- error privacy;
- the unexported manageable list;
- the commit-ambiguity contract;
- the runtime projection-lock incompatibility stop;
- exact runtime grants and read-only views;
- B7 with the checkpoint after cycle 4;
- the CI profile and registry counts;
- gates E1, E2, E4, E5 and S0;
- the named external actions X1/X2.

## 1. Approach summary

1. **Trusted context.**
   - `BeginRequest(ctx, credential)` validates the opaque credential through the accepted `identity.Service.Session`
     boundary. It returns a `ctx` carrying a new, unique authenticated request scope.
   - `Resolve(ctx, organization selector, optional Group selector)` takes the subject from that scope. It checks, in
     one REPEATABLE READ READ ONLY snapshot, the live session, the current Organization membership and Group
     selectability. It returns an immutable, opaque `*TenantContext` bound to that exact scope.
   - The context holds no roles, capabilities or allow results.
   - Every guarded operation first rejects a context that is not bound to the current `ctx` scope and its subject.
     It then revalidates the context completely inside its own transaction (§3.2, §4).
2. **Closed evaluator.** A pure function over raw facts decides exactly 11 capabilities × 6 roles (Spec R3) and is
   deny-by-default. It has no MFA, provider, name, hostname or claim input.
3. **One flag.** The Organization column `app.organizations.group_admins_edit_business_profiles boolean NOT NULL
   DEFAULT true` is added by additive migration `00005_organization_policy.sql`. Only `org_owner` can update it, through
   the facade.
4. **Bounded guarded facade.** Eight exported operations plus `BeginRequest` and `Resolve` (§3.5). Each operation
   authorizes and acts in **one database transaction on one connection**. Writes use READ COMMITTED with an ordered
   FOR SHARE lock set on every relied-on mutable fact; reads use one REPEATABLE READ snapshot. The query and lock
   footprint is in §4.7 ([PG-001](../PERFORMANCE.md#pg-001--g007-authorization-lockquery-footprint)).
5. **Minimum predecessor changes.**
   - identity: two transaction-scoped session checks.
   - tenancy: one fact reader and four transaction-scoped writes that reuse the existing SQL, validation and error
     mapping.
   - `UnlinkGroupBusinessProfile` takes its Group row lock first, so it conflicts with guarded writers. This is the
     only change to the behaviour of an existing writer, and it is not visible to its callers.

   No other writer, signature, projection or accepted migration changes.
6. **Evidence:**
   - unit evaluator and boundary tests in the existing `unit` suite;
   - four new `migrations`-suite IDs: catalog, migration, model and ordering, all against real PostgreSQL as the
     runtime role. The model and ordering IDs also carry the bounded PG-001 evidence (§6.7), so no registry ID is
     added;
   - mechanical version-5 reconciliation of the existing sites;
   - final local full coverage before the MR, the automatic MR `fast` job and the automatic staging gate after
     conditional integration.

   No CI mapping change.

## 2. Gates and baseline

| Gate | Requirement | State |
|---|---|---|
| E1 | G004–G006 implementation ACCEPT and integration | Satisfied: G006 final review, merge `4486b41` |
| E2 | Reconcile against the execution baseline: equal inputs need nothing; mechanical differences are re-derived and recorded; semantic changes to identity/tenancy/migrations/harness return to Codex. G051 changes to `run.sh`/registry/controls are expected to be mechanical | At execution start |
| E3 | Codex ACCEPT of this Plan, including the open §12 questions (Q13–Q17). The owner disposes of Q13/Q14 before implementation; "proceed with the recorded PG-001 risk" is a valid answer. The r0 review settled the Q8 pins without an owner decision | Pending |
| E4 | Fresh committed baseline, approved-input snapshot/digest, exact scope allow-list (§8), fresh isolated worktree | Pending (Codex) |
| E5 | Named external actions X1/X2 (§11) authorized | Pending (owner) |
| S0 | Start gate (read-only, no cycle): E1–E5 recorded; `run-approved.sh --check-only` passes; HEAD = bound base, NUL-safe status clean, root `.env` absent; `bash infra/ci/docker-desktop-mounts.sh` passes; ports and `local-harness.lock` free; no foreign owned Docker resources; GitLab compute-minute usage read before X1 | Pending |

Observed baseline facts at `6889273` that this Plan relies on:

- Every tenancy model method opens its own READ COMMITTED transaction (`inTx`). Session resolution is a separate
  autocommit read plus renewal. Nothing spans an authorization check and a later write.
- Writers and the row locks they take:
  - `revokeSession` and `revokeAll`: UPDATE `sessions`.
  - `DisableUser`: `users FOR UPDATE`, then UPDATE `sessions`.
  - sign-in rotation: UPDATE the old `sessions` row, then `users FOR SHARE`.
  - Role writers: UPDATE or DELETE one `role_bindings` row.
  - `Unlink`: DELETE one `group_business_profiles` row, after plain existence reads.
  - `Link`: INSERT, which takes FK `FOR KEY SHARE` on `organizations`, `groups` and `business_profiles`.
- Runtime privileges that matter for row locks: FOR SHARE requires UPDATE on at least one column.
  - Locks are possible on `users`, `sessions` (UPDATE), `organizations(name)`, `groups(name, slug)`,
    `role_bindings(role)` (also through the two simple `security_invoker` projections) and `business_profiles`.
  - **Not possible on `group_business_profiles`**: runtime has only SELECT/INSERT/DELETE there. This is why §4 uses
    the Group row as the association's lock surrogate.
- `sessions.mfa_assurance` is CHECK-constrained to `'unknown'` and no writer changes it.
- Tenancy U-BOUNDARY pins 10 files, 29 Service methods and 31/12 statements, and keeps all SQL in `store.go`. Its
  import allow-list excludes identity.
- B-MCP-SOURCE rejects any binary linking `core/internal/tenancy(/.*)?`. `profile-paths.tsv` classes
  `core/internal/tenancy/` as `foundation`, and `APPLICABLE` already includes `core/internal/tenancy`.
- PostgreSQL 17 (`pgvector/pgvector:0.8.7-pg17`); pgx v5.11.0.

## 3. Contracts

### 3.1 Placement and dependency direction

New package **`core/internal/tenancy/access`** (Go package `access`):
- It imports `tenancy`, `identity`, `observability` and pgx. It contains **no SQL**.
- tenancy keeps its allow-list: it never imports identity or access. identity imports neither.
- The subpackage placement means:
  - the existing `foundation` row and `APPLICABLE` entry already cover it;
  - B-MCP-SOURCE already proves that no binary links it.

  So there is **no CI mapping change** (Q1).
- No binary, route, handler, CLI, environment variable or bootstrap is wired in G007.

### 3.2 Request scope, TenantContext and resolver (R1)

**Authenticated request scope (r1, C1).** The scope is the trusted per-request binding. It is in-process and
unforgeable, and it never depends on request IDs, traces, time or any client-supplied value.
- `type requestScope struct{ … }` is **unexported**. Its fields:
  - issuer (`*access.Service`);
  - the authenticated session UUID;
  - the subject user UUID;
  - an `ended` flag (`atomic.Bool`).
- It is attached to a `context.Context` under an unexported key type. Only `BeginRequest` creates one, so code
  outside the package can neither construct a scope nor place one in a `ctx`.
- Each `BeginRequest` allocates a fresh scope. Its identity is pointer identity, so two requests never share a scope,
  even with the same credential.
- The scope holds no credential, CSRF token, display name, provider session, assurance, role or allow result.
- It is a **binding, not an authority snapshot** (R6): every operation rechecks the session, user and facts in its own
  transaction.

`BeginRequest(ctx, credential string) (context.Context, func(), error)`:
1. If `ctx` already carries any request scope, from this or another Service, it returns `context_denied` with reason
   `context_scope_nested`. It does this before identity is called, so one request chain has exactly one
   authenticated subject.
2. `identity.Service.Session(ctx, credential)`: the accepted G004 path, including renewal.
   - Failure gives `unauthenticated` (`session_invalid`).
   - Identity unavailable gives `unavailable` (`store_unavailable`).
3. It returns a derived `ctx` carrying the new scope, and an `end` function.
   - `end` marks the scope ended. It is idempotent and safe for concurrent use.
   - The request owner defers it: a future adapter, or the tests in G007.
   - Derived contexts (deadline, cancellation, values) keep the scope, so legitimate same-request calls work.

access reaches identity through an unexported interface (`Session`, `LockLiveSession`, `CheckLiveSession`).
`*identity.Service` satisfies it, and `NewService` takes `*identity.Service`. The interface lets U-CTX prove which
rejections happen before identity or the database is called.

**TenantContext.** `type TenantContext struct{ … }` has unexported fields only:
- issuer (`*access.Service`);
- scope (`*requestScope`);
- subject user UUID;
- session UUID;
- Organization UUID;
- Group UUID (zero means Organization-only).

It stores **no request ID or other correlation value**. Read-only accessors return the subject, Organization and
optional Group. They confer nothing. Roles, capabilities, assurance and allow results are never stored in it.

`Resolve(ctx, ResolveRequest{OrganizationID, GroupID string}) (*TenantContext, error)`:
1. Scope check, which is pure and comes before any SQL. The `ctx` scope must exist, must have been issued by this
   Service and must not have ended. Otherwise it returns `context_denied` (`context_scope_missing` or
   `context_scope_ended`).
2. Parse the selectors with `tenancy.Parse*ID`. Malformed, braced, unhyphenated, empty or all-zero values give
   `invalid_input` with reason `selector_invalid`. An empty Group selector means Organization-only.
3. A REPEATABLE READ READ ONLY transaction:
   - `identity.CheckLiveSession(tx, scope.session, scope.subject)`. Failure gives `unauthenticated`, so a session
     revoked after `BeginRequest` is caught here;
   - `tenancy.ReadFacts`, which requires an Organization row **and** a current Organization binding of the subject;
   - with a Group selector, the Group must exist in that Organization and the subject must hold `org_owner`/
     `org_admin` or a binding in that exact Group.

   An unknown or foreign Organization or Group, no Organization membership, or a Group binding without an
   Organization binding all give the **same** `context_denied` (internal reasons `organization_not_permitted` and
   `group_not_selectable`). No row is written and there is no default Organization.
4. It returns a context whose subject and session are copied from the scope, bound to that exact scope object.
   Several `Resolve` calls in one scope are allowed, for example Organization-only and then selected-Group. Each
   returned context is bound to the same scope.

**Binding check.** Every guarded operation runs this check. It is pure and runs before `BeginTx` or any SQL. The
conditions are evaluated in this order, the first failure decides the reason, and every failure gives external
`context_denied`:

| # | Rejected when | Reason |
|---|---|---|
| B1 | the context is nil or zero-valued, or caller-built (`issuer == nil` or `scope == nil`) | `context_forged` |
| B2 | the context was issued by another `Service` instance | `context_forged` |
| B3 | `ctx` carries no scope, or a scope issued by another `Service` | `context_scope_missing` |
| B4 | the `ctx` scope has ended | `context_scope_ended` |
| B5 | the context's scope is not the `ctx` scope (pointer identity) | `context_scope_mismatch` |
| B6 | the context's subject or session differs from the `ctx` scope's | `context_subject_mismatch` |

B6 cannot fail when B5 passes, because contexts are immutable and built only from their own scope. B6 is kept as an
independent defence-in-depth check, and U-CTX exercises it directly through in-package construction.

The binding gives these results:
- **Cross-request substitution fails B5.** This holds even when both requests have the same subject and session
  (one credential used twice), carry equal well-formed inbound `X-Request-Id` values, or have no request ID at all.
- **Cross-subject substitution fails B5 and B6:** user A's context fails under user B's authenticated scope.
- **Use outside an authenticated request fails B3**, for example under `context.Background()` or a fresh in-process
  `ctx`.
- **Use after the request ends fails B4**, even through `context.WithoutCancel`.
- **Same-request reuse passes the binding but stays subject to §4.** A context reused within its own request still
  observes every committed session, user, membership, link and policy change. The scope never preserves authority.

**The binding's boundary.** A scope and its contexts are usable only together, and only until `end`. The check is in
process: a component inside the trusted core process can always misuse its own request's pair. That is the same trust
boundary as the raw model primitives (§3.7 bypass control). Even then, §4 applies only that subject's current
authority.

**Correlation stays diagnostic.** Only `log.go` reads `observability.RequestID` and `TraceID`, to annotate records
(U-ACCESS-BOUNDARY). They are never compared, stored or used in a decision. Hostnames, headers, cookies, provider
claims, email, names, slugs and model data have no input path. Transport checks such as origin and CSRF remain the
obligation of future adapters; G007 wires no adapter. Q9 is answered by this section.

### 3.3 Capabilities and evaluator (R2, R3)

`decide(cap Capability, ctx contextSelection, f Facts, t Target) (allow bool, reason Reason)` is a pure, closed
`switch`. Unknown capabilities, roles, sources or target kinds, and inconsistent facts, return deny with reason
`capability_unknown` or `fact_inconsistent`. There is no ranking, prefix matching or global authority. Gates are
applied in order:

1. **Org gate:** Facts carry the active Organization and exactly one current Organization binding with source
   `membership`, else `organization_not_permitted`.
2. **Target gate:** target kind matches the capability.
   - Organization capabilities and `audit.read` target the active Organization.
   - `group.*` targets an exact Group of it.
   - `business_profile.*` targets an exact profile of it.

   A mismatch or a foreign/unknown target gives `target_not_found` or `target_kind_mismatch`.
3. **Context gate:** in selected-Group context, a `group.*` target must equal the selected Group
   (`target_outside_context`).
4. **Matrix** (Spec R3, unchanged):
   - `all` means an Organization owner/admin path to any Group of the active Organization.
   - `own` means the subject's binding in the exact target Group.
   - `linked` means a candidate Group `G` with the required Group role and `association(G) = profile`. Candidate Groups
     are `{selected Group}` in selected context, else every Group of the subject in the active Organization.
   - `linked+flag` additionally requires `group_admins_edit_business_profiles = true`.
   - Organization owner/admin profile paths are independent of the selected Group and include unlinked profiles.
   - `select_for_onboarding` ignores the flag. Member-only paths allow read only.
   - `privileged_settings.manage`, `organization.manage_policy` and `audit.read` have no Group path.

**Interpretation pins for review (Q8):**
- (a) Every cell for a Group-role subject is tested with the Organization gate held constant by an `org_member`
  binding. The expected result is the independent combination (e.g. `group_owner` + `org_member` may
  `organization.read` through `org_member` only).
- (b) An org owner/admin's Group `all` path is restricted by a selected Group context to that Group.
- (c) Policy read is `organization.read`; policy update is `organization.manage_policy`.
- (d) The manageable-profile set is `{P in active Organization : decide(select_for_onboarding, P) = allow}` under the
  current context. Org-only context therefore deduplicates across attached Groups. Selected context limits only the
  Group-derived paths.

### 3.4 Policy flag (R4)

- **Representation:** a column on the Organization row.
  - Every existing row reads `true` after `ADD COLUMN … DEFAULT true` (PostgreSQL fast default, no rewrite).
  - Every new Organization gets `true` from the column default.
  - `NOT NULL` means there is **no missing-row or malformed state**. `true` is the configured default, never a
    fallback: a read error is `unavailable`, never an allow.
- **Grants:** runtime gets `UPDATE (group_admins_edit_business_profiles)` only, with no INSERT on the column, so a
  creator cannot pre-set it.
- **Internal contract:**
  - `GetOrganizationPolicy(ctx, tc) (tenancy.OrganizationPolicy, error)`, which requires `organization.read`;
  - `SetOrganizationPolicy(ctx, tc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles bool})`, which
    requires `organization.manage_policy`. It is an absolute set, so setting the current value is a successful no-op.
- **Malformed values** are unrepresentable in the typed contract. AC5's malformed evidence is therefore database-level:
  runtime and owner `NULL`/text assignments are rejected (23502/22P02) with the row unchanged (CT-POL-3).
- Request data, environment, claims and other Organizations cannot override it. There is no policy map, DSL, editor,
  MFA policy or public endpoint.

### 3.5 Guarded facade (R5): exact surface

| Method | Capability | Target | Transaction | Write |
|---|---|---|---|---|
| `BeginRequest(credential)` | none (authentication via `identity.Service.Session`) | — | none in access (identity's accepted resolve and renewal) | — |
| `Resolve(selectors)` | none (scope check + context gate) | Organization (+ Group) | RR READ ONLY | — |
| `GetOrganization(tc)` | organization.read | active Organization | RR RO | — |
| `GetOrganizationPolicy(tc)` | organization.read | active Organization | RR RO | — |
| `SetOrganizationPolicy(tc, p)` | organization.manage_policy | active Organization | RC RW + locks | `UPDATE app.organizations SET flag` |
| `GetGroup(tc, g)` | group.read | exact Group | RR RO | — |
| `RenameGroup(tc, g, name)` | group.update | exact Group | RC RW + locks | accepted rename SQL |
| `ChangeGroupSlug(tc, g, slug)` | group.update | exact Group | RC RW + locks | accepted slug SQL |
| `GetBusinessProfile(tc, p)` | business_profile.profile.read | exact profile | RR RO | — |
| `UpdateBusinessProfile(tc, p, c)` | business_profile.profile.update | exact profile | RC RW + locks | accepted full-replace SQL |

- Every guarded operation and `Resolve` runs the §3.2 scope/binding check first, before `BeginTx`.
- `BeginRequest` and `Resolve` have **no capability**. They establish authentication and context; they are not
  policy decisions (§3.6 represents this explicitly).
- `manageableProfiles(tc)` is unexported. It is used by in-package tests for AC4/AC7 (Q7).
- `group.manage_members`, `business_profile.select_for_onboarding`, `business_profile.privileged_settings.manage` and
  `audit.read` are evaluated (unit and DB-loaded facts) but have **no** operation. No allow result, capability check or
  evaluator is exported.
- Payloads keep the accepted contracts: `tenancy.ValidateName`, `ValidateSlug` and `ValidateBusinessContext` run
  **before** authorization. They are pure and reveal nothing about targets.
- A shared-profile read returns only the profile, never its associations, access paths or sibling memberships.

### 3.6 Errors, reasons and logging (R5, R9)

`*access.Error{Code, Field}` prints `access: <code>`, never IDs or driver text. The external codes are closed:

| Code | When |
|---|---|
| `unauthenticated` | absent, invalid, expired or revoked session; disabled user |
| `context_denied` | the scope/binding check (§3.2 B1–B6, nested `BeginRequest`), the context gate, or the context's Organization membership or Group selectability no longer holds |
| `not_found` | unknown, foreign or kind-mismatched target, a target outside the selected Group, **or** a target the subject cannot read |
| `forbidden` | the target is readable by the subject but the requested capability is denied (including flag off) |
| `invalid_input` | malformed selector or payload; `Field` carries the accepted tenancy code (e.g. `invalid_slug`) |
| `conflict` | `slug_conflict` |
| `unavailable` | identity/database unavailable, a definitive abort, cancellation, or a **lost commit reply** (§4.5) |
| `internal` | unexpected tenancy `internal` |

`forbidden` versus `not_found` uses the read capability of the same kind, evaluated on the same facts, so an
inaccessible target is indistinguishable from an unknown one.

**Internal reasons** (closed; r1 replaces `context_request_mismatch` with the five scope reasons and adds
`context_established`):
- `allowed`, `context_established`, `session_invalid`, `selector_invalid`;
- `context_forged`, `context_scope_missing`, `context_scope_ended`, `context_scope_mismatch`,
  `context_subject_mismatch`, `context_scope_nested`;
- `organization_not_permitted`, `group_not_selectable`, `target_not_found`, `target_kind_mismatch`,
  `target_outside_context`;
- `capability_unknown`, `role_not_permitted`, `no_linked_path`, `edit_policy_disabled`, `fact_inconsistent`;
- `input_invalid`, `write_conflict`, `store_unavailable`, `aborted`, `canceled`, `commit_outcome_unknown`.

**Event:** one catalogued `access.decision` record per call, including `BeginRequest` and `Resolve`.
- **Fields:**
  - `operation`: 10 closed names. These are `access.begin_request` and `access.resolve` (the two **context
    operations**), plus `organization.get`, `organization.get_policy`, `organization.set_policy`, `group.get`,
    `group.rename`, `group.change_slug`, `business_profile.get` and `business_profile.update`;
  - `capability`: 11 closed values;
  - `outcome`;
  - `reason`;
  - `request_id`/`trace_id` when present.
- **Capability absence is explicit** (review routine item). The catalog rules are:
  - For a context operation, `capability` is **forbidden**. `outcome` ∈ {`established` (info), `deny` (warn),
    `failed` (error)}, and `reason` is never `allowed`.
  - For the 8 guarded operations, `capability` is **required** and must be the operation's single §3.5 mapping.
    `outcome` ∈ {`allow` (info), `deny` (warn), `failed` (error)}.
  - `allowed` occurs only with `allow`; `context_established` occurs only with `established`.
- No eleventh mapping, sentinel capability value or policy allow is recorded for a context operation.

Records never contain UUIDs, names, slugs, roles, profile/policy values, tokens, claims or driver text. Guarded writes
do not also emit `tenancy.operation`, because the tx-scoped helpers do not log.

### 3.7 Exact shared-file changes

**identity:**
- Two `Service` methods using the service clock:
  - `LockLiveSession(ctx, tx pgx.Tx, sessionID, userID string) error`:
    `SELECT … FROM app.users WHERE id=$1 AND status='active' FOR SHARE`, then
    `SELECT … FROM app.sessions WHERE id=$1 AND user_id=$2 AND revoked_at IS NULL AND idle_expires_at>$3 AND
    absolute_expires_at>$3 FOR SHARE`.
  - `CheckLiveSession(...)`: the same two statements without `FOR SHARE`, for RR snapshots.
- Both return `ErrUnauthenticated` or `ErrUnavailable`.
- No renewal, schema or existing-method change.

**tenancy** (new file `facts.go`; statements in `store.go`; Service surface stays 29):
- `ReadFacts(ctx, tx pgx.Tx, q FactQuery) (Facts, error)`. It performs the ordered reads of §4.1 steps 3–7, in lock
  mode or snapshot mode.
  - `FactQuery` holds the Organization and user, the candidate Groups (nil = all of the subject's Groups in the
    Organization), the optional exclusive target Group, the optional profile, an all-profiles flag (for the manageable
    set), `Lock bool` and `ExclusiveOrganization bool`.
  - `Facts` holds the Organization row and flag, the Organization role, the Group roles, the Group rows, the
    associations, and the target or all profiles.
  - Membership facts are read **only through the two projections**. Association facts are read only from
    `group_business_profiles` with the Organization predicate.
- `RenameGroupTx`, `ChangeGroupSlugTx`, `UpdateBusinessProfileTx` and `SetOrganizationPolicyTx`. The first three reuse
  the existing SQL, validation and `classify`. `RenameGroup`, `ChangeGroupSlug` and `UpdateBusinessProfile` call the
  same unexported bodies, with behaviour and logging unchanged.
- `UnlinkGroupBusinessProfile`:
  - Its Group precheck becomes `SELECT id FROM app.groups WHERE organization_id=$1 AND id=$2 FOR NO KEY UPDATE`;
    a missing row is still `group_not_found`. This runs before the DELETE.
  - Codes, ordering of codes, results and the single-statement DELETE are unchanged.
- U-BOUNDARY pins: 11 files; 29 methods; an explicit pinned list of exported package-level functions (the existing
  `Parse*`/`Validate*`/`CodeOf`/`ReservedSlugs`/`NewService` plus `ReadFacts` and the 4 `*Tx`); the new statement
  counts; locking reads are still `SELECT`s carrying `organization_id = $1` (or `id = $1` for `organizations`).
- **Bypass control** (Q2): the tx helpers are as raw as the accepted Service methods. The guarantee is the production
  import boundary. U-ACCESS-BOUNDARY walks every non-test `.go` file under `core/`; only `tenancy/access` may import
  `core/internal/tenancy`, and nothing may import `tenancy/access` in G007. B-MCP-SOURCE keeps both out of every
  binary.

## 4. Authorization/write ordering contract (R6, AC6, AC8)

### 4.1 Guarded write

**Write transaction:** one READ COMMITTED read-write transaction on one connection. Statements run in this order:

1. `users` row: `FOR SHARE` with `status='active'`. This is first, matching `DisableUser`'s user-then-sessions order.
2. `sessions` row: `FOR SHARE` with the id/user/revoked/expiry predicates at service-clock `now`.
3. `organizations` row:
   - `FOR SHARE` for every write except `SetOrganizationPolicy`;
   - `FOR NO KEY UPDATE` for `SetOrganizationPolicy`, taken up front so a shared lock is never upgraded.
4. The subject's Organization binding, `FOR SHARE` through `organization_memberships`.
5. The subject's Group bindings for the candidate set, `FOR SHARE` through `group_memberships`, ordered by
   `group_id`. The candidate set is:
   - the target Group, for Group writes;
   - the selected Group, for any write in selected context;
   - every subject Group in the Organization, for an Organization-only profile write.
6. Group rows:
   - the update target first, `FOR NO KEY UPDATE`;
   - then the remaining candidates `FOR SHARE`, ordered by `id`.
7. Associations of the locked Groups: a plain read in a **new statement**, so it sees every unlink committed before
   step 6 and no unlink can commit after it.
8. Profile/Group target read; `decide`.
9. The write.
10. `COMMIT`.

**Linearization point:** the completion of step 7.
- Every relied-on mutable fact is row-locked from its own statement until commit or rollback.
- Under READ COMMITTED, a statement that waited on a concurrent writer re-evaluates its predicates against the
  committed row version (PostgreSQL's re-check of updated rows):
  - a revoked or expired session, a disabled user, or a deleted or changed binding is excluded or re-read;
  - an unlink committed before step 6 is invisible to step 7.
- `decide` uses **only rows returned by steps 1–7**. A fact that appears later (a concurrent grant or link) is never
  relied on, so it needs no lock.

**Conflicting writers:** any later writer of a locked fact blocks until this transaction ends. The authorized write
therefore commits **before** the conflicting change, or it observes the change and denies. Either way, no write
commits after a revocation that it did not see.

### 4.2 Guarded read and resolve

- **Mode:** one REPEATABLE READ READ ONLY transaction, using `CheckLiveSession` and `ReadFacts` in snapshot mode.
  PostgreSQL forbids row locks in read-only transactions.
- **Linearization point:** the snapshot. All facts and the returned row come from one consistent state.
- **Guarantee:** a read that starts after a revocation commits always observes it.

### 4.3 Compatibility with the accepted writers

| Writer (unchanged unless noted) | Lock it takes | Conflict with guarded write | Result |
|---|---|---|---|
| `Logout` current / provider (`revokeSession`) | UPDATE `sessions` row | step 2 FOR SHARE | either the guard commits first or the session is revoked and denied |
| `Logout` all (`revokeAll`) | UPDATE user's `sessions` | step 2 | same |
| `DisableUser` | `users FOR UPDATE`, then UPDATE `sessions` | steps 1–2 (same order) | same; `EnableUser` never revives a revoked session |
| Sign-in rotation (superseded) | UPDATE old session, then `users FOR SHARE`, then **(r1) `UPDATE users` display name only when it changed** | step 2 (and step 1 for the display-name UPDATE) | same. **r1:** with a changed display name, one rare wait cycle with steps 1–2 is possible (§4.4, Q15) |
| (r1) Fresh sign-in display-name refresh | `users FOR SHARE`, then conditional `UPDATE users` | step 1 | brief wait only; a guard that waited re-checks `status='active'` |
| (r1) Session renewal (`resolve`, at most once per 60 s); provider-logout bookkeeping (`recordProviderLogout`, `consumeLogoutRef`) | single-row UPDATE `sessions` | step 2 | brief wait only; none of them revokes or revives a session |
| Session expiry (time) | none | predicate at step 2 | evaluated once at the lock point; a session expiring after it does not affect the in-flight write (ordered before expiry) |
| `ChangeOrganizationRole` / `RemoveOrganizationMembership` | UPDATE/DELETE binding | step 4 | same |
| `ChangeGroupRole` / `RemoveGroupMembership` | UPDATE/DELETE binding | step 5 | same |
| `UnlinkGroupBusinessProfile` (**changed**: Group row `FOR NO KEY UPDATE` first) | Group row, then DELETE association | step 6 FOR SHARE | same |
| `SetOrganizationPolicy` (new guarded op) | `organizations FOR NO KEY UPDATE`, UPDATE | step 3 | same |
| Grants, `Link`, `CreateGroup`, `CreateBusinessProfile` | INSERTs, FK `FOR KEY SHARE` | none (compatible) | new authority, never relied on (§4.1) |
| `RenameOrganization`, raw `RenameGroup`/`ChangeGroupSlug`/`UpdateBusinessProfile` | UPDATE one row | brief wait only | not authority sources |

MFA has no writer: assurance is constant `unknown`, so no MFA race exists. Callers adopt no lock convention: every
writer keeps its signature and caller contract.

### 4.4 Deadlock freedom

- Guarded writes take shared locks plus at most two exclusive locks:
  - the Organization row (policy update), at step 3;
  - one target Group row, at step 6.

  They take them in a fixed global order. No shared lock is upgraded.
- Every existing writer holds at most one contended exclusive row lock before its single conflicting statement:
  - `DisableUser` takes user then sessions, which is the guard's order;
  - Unlink takes only the Group row.
- **Exception found in r1 (Q15).** r0 overlooked sign-in rotation. With a changed provider display name, rotation
  takes two contended locks in the order opposite to the guard's:
  - it updates the old `sessions` row S;
  - it takes `users FOR SHARE`, which is compatible with the guard;
  - it then runs `UPDATE users`, which conflicts with the guard's step-1 share lock.

  A guard holding `users` U (step 1) and then requesting S FOR SHARE (step 2) waits for the rotation. The rotation's
  display-name UPDATE waits for the guard. PostgreSQL detects the cycle and aborts one transaction with `40P01`:
  - an aborted guard returns `unavailable` (`aborted`) with state unchanged;
  - an aborted rotation rolls back and its callback fails `identity_unavailable`, issuing no session; S stays live and
    the guard proceeds.

  No unauthorized write results. The cycle needs three things at once: the guard's context session is the browser
  session being superseded, a concurrent additional-provider-session callback for the same user, and a changed
  display name. Reversing the guard's order would instead create a cycle with `DisableUser`, so r1 keeps the r0
  order and changes no writer. Disposition is Q15; the proposed evidence is O-15 (§6.4).
- Apart from that exception, no wait cycle is possible. AC8 also runs both orders of every guarded-guarded pair that
  can conflict (O-14). An unexpected `40P01`/`40001` is still a definitive abort: `unavailable`, state unchanged, no
  retry.

### 4.5 What a caller can conclude: definitive abort versus lost commit reply

| Failure | How detected | Returned | Caller may conclude |
|---|---|---|---|
| Validation, deny, statement error, cancellation before `COMMIT` is sent | error before Commit; deferred rollback | the pinned code; `unavailable` for failures | **nothing changed** (tested by row-set hashes) |
| Server-reported `COMMIT` failure (a `PgError` from Commit) | definitive rollback | `unavailable` (`aborted`) | nothing changed |
| `COMMIT` sent and reply lost (connection error, EOF, cancellation during Commit) | non-`PgError` from Commit | `unavailable` (`commit_outcome_unknown`) | **nothing**: the write may be durable. Reconcile through a guarded read, which is subject to current authority. Rollback is never claimed |

- **No automatic retry and no idempotency key.**
- All four guarded writes are absolute assignments (name, slug, full context, boolean). Re-applying the same request
  therefore cannot accumulate effects. Setting the current slug or value is a successful no-op.
- Limits of a caller retry:
  - It is a **new** operation, authorized at retry time. It may deny after a revocation even if the first attempt
    committed.
  - Under the accepted last-commit-wins semantics, it may overwrite an intervening concurrent write.
- No "exactly once" or double-apply-free claim is made beyond this.

### 4.6 Not claimed

- No database-wide isolation or RLS (G008).
- No protection against holders of the runtime credential running arbitrary SQL.
- No serializable cross-table guarantee for unguarded raw model calls.
- No ordering for reads beyond their snapshot.
- (r1) No performance target, efficient query plan or bounded lock-wait time (§4.7).

### 4.7 Query and lock footprint (r1; ARCHITECTURE §63, PG-001)

Register entry: [PG-001](../PERFORMANCE.md#pg-001--g007-authorization-lockquery-footprint) (on
`origin/staging`; the link resolves once the Plan is imported). The Plan links the entry rather than duplicating it.
r1 does not edit the register, which is outside the planning write set; Codex updates its status on review, and the
Result links it with the §6.7 evidence.
- **Status:** design hypotheses derived from the planned statements of §3.7/§4.1. **Nothing is measured.**
- **No target:** no performance target exists and none is set here. Existing indexes are mapped below; they are not
  evidence of an efficient plan.
- **Design:** the footprint is r0's §4.1 design, unchanged.

**Notation.** U, S and O are the subject's `users` row, its `sessions` row and the active `organizations` row.
- OB is the subject's Organization binding.
- GB is the subject's Group bindings in the candidate set, and G is the candidate Group rows.
- A is the candidate Groups' associations, and P is the target profile.
- "sel" means the selected Group and "t" a target Group.
- Cardinalities:
  - n_g: the subject's Group bindings in the active Organization;
  - N_g: the subject's Group bindings across all Organizations;
  - k: the subject's Organizations;
  - M_org: the Organization-scope bindings of the active Organization;
  - B_org: all Group bindings in the active Organization.

#### 4.7.1 Per-operation footprint

Each set is fetched by one `= ANY($n)` statement, so the **statement count per call is constant, with no per-Group
query (no N+1)**. Rows row-locked are exactly the rows the locking statements return (FOR SHARE unless noted). E-PG-2
checks this.

| Operation | Transaction | Statements (excluding BEGIN/COMMIT) | Rows read | Rows row-locked | Written |
|---|---|---|---|---|---|
| `BeginRequest` | identity's autocommit | 1 SELECT, plus 1 renewal UPDATE at most once per 60 s per session | 2 (session ⋈ user) | renewal: S (UPDATE) | S (renewal only) |
| `Resolve`, Organization-only | RR RO | 4: U, S, O, OB | 4 | none | — |
| `Resolve`, selected Group | RR RO | 6: adds G(sel) and GB(sel) | ≤ 6 | none | — |
| `GetOrganization`, `GetOrganizationPolicy` | RR RO (+2 statements in selected context) | 4: U, S, O (carries name and flag), OB | 4 (+2) | none | — |
| `GetGroup` | RR RO | 6: U, S, O, OB, GB(t), G(t) | ≤ 6 | none | — |
| `GetBusinessProfile`, selected | RR RO | 8: U, S, O, OB, GB(sel), G(sel), A(sel), P | ≤ 8 | none | — |
| `GetBusinessProfile`, Organization-only | RR RO | 8 | 5 + 3·n_g | none | — |
| `SetOrganizationPolicy` | RC RW | 5: U, S, O, OB, UPDATE O (+2 in selected context) | 4 (+2) | U, S, OB; **O FOR NO KEY UPDATE** (+GB(sel), G(sel)) | O |
| `RenameGroup`, `ChangeGroupSlug` | RC RW | ≤ 8: U, S, O, OB, GB(t), G(t), target read, UPDATE G | ≤ 6 | U, S, O, OB, GB(t); **G(t) FOR NO KEY UPDATE** | G(t) |
| `UpdateBusinessProfile`, selected | RC RW | 9: U, S, O, OB, GB(sel), G(sel), A(sel), P, UPDATE P | ≤ 8 | ≤ 6 (U, S, O, OB, GB(sel), G(sel)), plus P through the UPDATE | P |
| `UpdateBusinessProfile`, Organization-only | RC RW | 9 | 5 + 3·n_g | **4 + 2·n_g** (U, S, O, OB, n_g GB, n_g G), plus P through the UPDATE | P |
| `UnlinkGroupBusinessProfile` (changed) | RC RW (accepted) | accepted prechecks with the Group precheck now FOR NO KEY UPDATE, then the accepted DELETE | unchanged | **G FOR NO KEY UPDATE** (new) | A (deleted) |
| `manageableProfiles` (unexported, tests only) | RR RO | Organization-only reads plus one all-profiles read | 4 + 3·n_g + profiles in the Organization | none | — |

The r0 design takes the Organization-only candidate set as every subject Group in the Organization, whatever the
subject's Organization role. So the 4 + 2·n_g lock count applies to every subject with n_g Group bindings, including
an Organization owner or admin whose decision uses only the Organization path.

#### 4.7.2 Existing index coverage (migrations 00002–00004)

| # | Statement | Base-relation predicate | Existing indexes | Expected access (hypothesis) | Index entries examined |
|---|---|---|---|---|---|
| s0 | `BeginRequest` session resolve | `sessions.credential_hash = $1`, join `users` by id | `sessions_credential_hash_key` (UNIQUE), `users_pkey` | unique probe + PK probe | 2 |
| s0r | renewal UPDATE | `sessions.id = $1` | `sessions_pkey` | PK probe; updated columns unindexed, so HOT-eligible | 1 |
| 1 | `users` lock/check | `id = $1 AND status = 'active'` | `users_pkey` | PK probe | 1 |
| 2 | `sessions` lock/check | `id = $1 AND user_id = $2 AND revoked_at IS NULL AND expiries > $3` | `sessions_pkey` (`sessions_user_live_idx` also exists) | PK probe | 1 |
| 3 | `organizations` lock/read | `id = $1` | `organizations_pkey` | PK probe | 1 |
| 4 | OB through `organization_memberships` | `role_bindings`: `organization_id = $1 AND subject_user_id = $2 AND scope_type = 'organization' AND source = 'membership'` | `role_bindings_subject_scope_key` (subject_user_id, scope_type, scope_id) prefix; `role_bindings_scope_idx` (organization_id, scope_type, scope_id) prefix | **H-1**: subject path, then filter on Organization; or Organization path, then filter on subject | k, or M_org |
| 5a | GB, explicit candidates (Group ops, selected context) | as 4 with `scope_type = 'group' AND group_id = ANY($3)` | `role_bindings_group_idx` (organization_id, group_id) WHERE group_id IS NOT NULL; `role_bindings_subject_scope_key` prefix | **H-1**: the candidate Groups' members, or the subject's Group bindings | members of the candidates, or N_g |
| 5b | GB, all subject Groups (Organization-only profile ops) | as 4 with `scope_type = 'group'` and no Group predicate | `role_bindings_subject_scope_key` prefix; `role_bindings_scope_idx` prefix; `role_bindings_group_idx` prefix | **H-1**: subject path (N_g), or an **Organization-wide** path (B_org) | N_g, or B_org |
| 6a | target Group | `organization_id = $1 AND id = $2` | `groups_pkey`, `groups_organization_id_key` | point | 1 |
| 6b | candidate Groups | `organization_id = $1 AND id = ANY($2) ORDER BY id` | `groups_organization_id_key` (organization_id, id) | one probe per candidate; index order equals lock order | n_g, or 1 |
| 7 | associations | `organization_id = $1 AND group_id = ANY($2)` | `group_business_profiles_group_key` (organization_id, group_id) UNIQUE | one probe per candidate | ≤ n_g |
| 8 | profile | `organization_id = $1 AND id = $2` | `business_profiles_pkey`, `business_profiles_organization_id_key` | point | 1 |
| 9 | writes | by `id` or `(organization_id, id)` | the same PK/UNIQUE indexes | point. Profile and flag columns are unindexed, so those updates are HOT-eligible. A slug change maintains `groups_organization_slug_key`, as accepted in G005 | 1 |
| U1 | Unlink Group precheck (changed) | `organization_id = $1 AND id = $2 FOR NO KEY UPDATE` | `groups_organization_id_key` | point | 1 |
| M | `manageableProfiles` (tests) | `business_profiles.organization_id = $1` | `business_profiles_organization_id_key` prefix | range | profiles in the Organization |

**Why H-1 is open.**
- No existing index leads with both `organization_id` and `subject_user_id`.
- The projections expose `organization_id`, `user_id` and `group_id` but not `scope_id`, so statements 4/5 cannot
  supply the third key column of `role_bindings_subject_scope_key`.
- The CHECK `role_bindings_scope_consistent` implies `scope_id = organization_id` for Organization scope, but the
  planner does not derive index conditions from CHECK constraints.
- The planner's choice depends on statistics, and r1 cannot determine it without plans (E-PG-1).

#### 4.7.3 Growth with Group count

- **Organization-only profile operations grow linearly in n_g:**
  - rows read: 5 + 3·n_g;
  - for writes, rows locked: 4 + 2·n_g;
  - one probe per candidate on statements 6b and 7;
  - statement 5b examines N_g entries on the subject path, or B_org on an Organization-wide path (H-1);
  - statement count does not grow.
- **All other operations are O(1) in Group count,** apart from H-1's choice on statements 4/5a.
- **Row-lock write cost (PostgreSQL behaviour, documented):**
  - Each FOR SHARE lock records the locker in the tuple header. That dirties the heap page and is WAL-logged
    (PostgreSQL docs, "Row-Level Locks": locking a row may cause a disk write).
  - Concurrent shared lockers of one row are represented by a MultiXact (docs, "Multixacts and Wraparound").
  - Hot shared rows: O for every guarded write in the Organization; U and S for the subject's writes; and G for
    widely shared Groups, which every Organization-only profile write by any member of those Groups locks.
  - Write amplification therefore grows with rows locked per write and per-Organization write concurrency.
  - Row locks create no new tuple version and touch no index.
- **Guarded reads take no row locks.** A REPEATABLE READ snapshot holds back the vacuum horizon only for its own
  short duration.

#### 4.7.4 Contention and long transactions

| Hot row | Shared-locked by | Writers that wait for in-flight guarded writes |
|---|---|---|
| O | every guarded write in the Organization | `SetOrganizationPolicy` (FOR NO KEY UPDATE), `RenameOrganization` (UPDATE) |
| U | every guarded write of the subject | `DisableUser` (FOR UPDATE), sign-in display-name UPDATE |
| S | every guarded write under the session | `revokeSession` (logout, provider logout), `revokeAll`, `DisableUser`'s sessions UPDATE, sign-in rotation, renewal, provider-logout bookkeeping |
| OB | the subject's own guarded writes | `ChangeOrganizationRole`, `RemoveOrganizationMembership` for that subject |
| GB | the subject's guarded writes (candidate set) | `ChangeGroupRole`, `RemoveGroupMembership` for that subject |
| G | guarded writes whose candidate set includes the Group | `UnlinkGroupBusinessProfile` (FOR NO KEY UPDATE), raw or guarded rename and slug change |

These are never delayed: grants, `Link`, `CreateGroup` and `CreateBusinessProfile`, because FK FOR KEY SHARE is
compatible.

- **Duration.** A waiting writer waits for the remaining duration of the in-flight guarded transactions.
  - A guarded transaction contains only its fixed statements, with no provider or network call inside. The
    `BeginRequest` identity call precedes it and runs outside it.
  - Its duration is unmeasured.
  - No server-side bound applies. Neither compose nor core configures `lock_timeout`, `statement_timeout` or
    `idle_in_transaction_session_timeout`, and core uses pgxpool defaults (`cmd/server/main.go`; G007 wires nothing).
  - The bounds are the caller's `ctx` deadline (a definitive abort, §4.5) and connection loss.
  - A guard blocked behind a writer keeps its own earlier locks. Because writers hold at most one contended exclusive
    lock, chains stop at writer → guard → writer (except Q15's cycle).
  - A blocked guard also holds its pool connection, which is relevant to future pool sizing and not to G007.
- **H-4: possible writer delay beyond one transaction (hypothesis).** This is the behaviour of PostgreSQL's row-lock
  code as r1 understands its source (`heap_lock_tuple`'s shared-lock path). It is unverified here.
  - When a row carries only shared locks, a new FOR SHARE request may proceed without queuing behind a writer
    already waiting for that row.
  - If so, sustained overlapping guarded writes could delay a queued writer for longer than one guarded transaction:
    - at Organization scope: a policy update or Organization rename behind O;
    - at subject scope: removal, downgrade, `DisableUser` or logout behind the same subject's own overlapping
      guarded writes.
  - Ordering safety is unaffected, because each guarded write commits before the change or observes it (R6). The
    exposure is the latency and liveness of revocation and policy writes.
  - G007 has no adapter, so no untrusted caller can generate this concurrency.
  - E-PG-3 characterizes the behaviour, and Q14 asks the owner for a disposition.

#### 4.7.5 Workload assumptions (explicitly unresolved)

No reviewed source states any of the following: ARCHITECTURE, the ADRs, Goal/Spec r3, ROADMAP and PG-001 all leave
them open.
- **Group cardinalities:**
  - Groups per Organization;
  - n_g, N_g and k;
  - members per Organization or Group;
  - Groups per shared profile.
- **Traffic:** guarded-operation rates, and per-Organization or per-subject concurrency.
- **Targets and configuration:** latency or lock-wait targets, pool size and server timeouts for the future adapters.

The only structural facts:
- ADR-004 lets a global user hold roles in many Organizations and Groups, with no stated bound.
- A Group has at most one profile, and a profile may be shared by many Groups.
- The G007 fixtures use 1–3 Groups per Organization.

r1 assumes no values; Q13 asks the owner.

#### 4.7.6 Assessment

- **Constant statements and point lookups.** Statement counts are constant per operation, and every point statement
  has a direct existing index.
- **H-1 (statements 4/5):** the access path is unknown, and an Organization-wide path for 5b is possible. E-PG-1
  settles it.
- **H-2:** Organization-only profile operations are linear in n_g. Whether that is acceptable depends on Q13.
- **H-3:** hot shared rows (O and widely shared G) bring MultiXact/WAL write amplification. They serialize policy
  updates, Organization renames, unlinks and Group renames with in-flight guarded writes. The magnitude depends on
  unknown concurrency.
- **H-4:** possible writer delay beyond one transaction, characterized by E-PG-3 and disposed of by Q14.
- **Migration cost.**
  - 00005's `ADD COLUMN … NOT NULL DEFAULT true` changes the catalog only, with no table rewrite (PostgreSQL ≥ 11).
  - It needs a brief ACCESS EXCLUSIVE lock on `organizations`. That lock waits for in-flight users of the table and
    blocks new ones while queued.
  - This does not matter for disposable validation databases, and production rollout is outside G007.
- **Feasibility.**
  - For small n_g, nothing found threatens feasibility.
  - No architecture, schema, index, grant or scope change is proposed.
  - Q13 and Q14 are raised before implementation, and §11 adds conditional stops.
- **Available mitigations, not adopted (Q17).** Both preserve the §4.1 ordering, least privilege and the grants, and
  neither needs a schema change.
  - **M1:** skip steps 5–7 when the locked OB role alone decides the operation. That is an Organization owner or admin
    on every profile and Group operation, keeping only the selected or target Group's row.
  - **M2:** restrict profile candidates to the subject's Groups linked to P:
    - read P's associations through `group_business_profiles_business_profile_idx`;
    - lock those Groups' GB and G rows;
    - re-read their associations in a new statement;
    - `decide` uses only the re-read. A concurrent link can then cause only a safe deny.

  Either changes §4.1 steps 5–7 and the affected O cases, so r1 keeps r0's lock set unless directed otherwise.
- **Evidence during approved execution: yes, bounded** (§6.7):
  - E-PG-1, plan summaries;
  - E-PG-2, the locked-row inventory;
  - E-PG-3, lock-queue characterization.

  There is no benchmark and no timing target. EXPLAIN ANALYZE times are recorded only as environment-labelled
  observations.

## 5. Schema: `core/migrations/00005_organization_policy.sql` (low risk, additive)

- **Up:**
  - the G005 preflight (schema owner, runtime role exists);
  - `ALTER TABLE app.organizations ADD COLUMN group_admins_edit_business_profiles boolean NOT NULL DEFAULT true`;
  - `GRANT UPDATE (group_admins_edit_business_profiles) ON app.organizations TO <db>_runtime`;
  - the G005-style postcondition: runtime table privileges on `organizations` are exactly `SELECT`; column privileges
    are exactly `name: INSERT, UPDATE` and the flag `UPDATE`; no other grantee and no PUBLIC entry.
- No new relation, function, trigger, index, RLS or SECURITY DEFINER. `organizations_keys_immutable` is unaffected.
- **Down:** `ALTER TABLE app.organizations DROP COLUMN group_admins_edit_business_profiles`.
  - This destroys only flag values and restores the version-4 behaviour and grants.
  - PostgreSQL leaves a dropped-attribute slot, so catalog comparisons exclude `attisdropped`, as the existing ones do.
  - Production down/reapply refusal is unchanged.
- Migrations 00001–00004 stay byte-identical; B-MIGRATE-SET pins the 00004 digest `f73bc9df…a677`.
- The relation count stays 35 at v4 and v5. `TestIdentityCatalog`'s table list is unchanged.

## 6. Tests and evidence (designed now; built and run only after S0)

### 6.1 Unit (`unit` suite; no harness change)

| ID | Content |
|---|---|
| U-POL-MATRIX | `matrix_test.go`, **written by the oracle delegate from Spec R3 text only** (§10): a literal 11×6 expected table plus own/all/linked/flag and selected-Group variants. It is never generated from `decide` |
| U-POL-GATE | No Organization binding denies all 11; Group binding without Organization binding; unknown capability, role, source and kind; empty, nil and inconsistent facts; target kind mismatch |
| U-POL-SCOPE | Selected-Group restriction; independent Organization paths; unlinked profile; new unattached Group cannot select an unrelated profile; manageable set rules |
| U-POL-MFA | `decide`'s inputs carry no assurance/provider field (reflection over the types); identical decisions |
| U-CTX (r1) | `context_test.go`, with a spy DB counting `BeginTx`/statements and a spy identity, so every case also asserts **zero `BeginTx`** and the pinned reason. **1:** a nil, zero-valued or caller-built context gives B1; another issuer's context gives B2. **2:** no scope (`context.Background()`), or another Service's scope, gives B3 for `Resolve` and all 8 operations. **3:** two scopes for the same subject and session, under `ctx`s from `observability.Middleware` with **identical** inbound `X-Request-Id`, and again with **absent** IDs: the context from scope 1 under scope 2 gives B5, and the reverse too. **4:** contexts for subjects A and B, with equal and with absent IDs, swapped: B5. **5:** an in-package context carrying the `ctx` scope pointer but another subject, or another session, gives B6, which proves the check live on its own. **6:** after `end`, `Resolve` and every operation give B4. `end` is idempotent, and ending one scope leaves another usable. A `context.WithoutCancel` child still gives B4. **7:** nested `BeginRequest` gives `context_scope_nested` with **zero identity calls**. **8:** a derived `ctx` (deadline or value child) of the same scope passes the binding and reaches `BeginTx`. **9:** reflection: neither `TenantContext` nor `requestScope` has a field holding a request ID, trace or credential |
| U-ACCESS-BOUNDARY | access imports allow-list; no SQL literal; no `ServeHTTP`/`net/http` in non-test files; pinned exported surface (r0's plus `BeginRequest`; `requestScope` and its key unexported); core-wide production-importer walk (§3.7); (r1) `observability.RequestID`/`TraceID`/`FromContext` referenced only from `log.go` |
| U-BOUNDARY | Updated tenancy pins (§3.7) |

### 6.2 B-POLICY-CATALOG and B-POLICY-MIGRATION

These run in `core/internal/migrate/policy_integration_test.go`, against real PostgreSQL, as runtime/owner.

**Catalog** (`TestPolicyCatalog`):
- CT-POL-1: the v4→v5 delta is exactly one column (type, NOT NULL, default) and one column ACL; every other catalog row
  is unchanged.
- CT-POL-2: the exact runtime privilege matrix; no INSERT on the flag; no PUBLIC.
- CT-POL-3 (rolled back): runtime and owner NULL/text assignments give 23502/22P02. Runtime `FOR SHARE` on
  `group_business_profiles` gives 42501, which documents §2's lock-surrogate reason. Runtime `FOR SHARE` succeeds
  through both projections and on `users`, `sessions`, `organizations` and `groups`.
- CT-POL-4: still no RLS; no new relation, function or trigger; the authority detector inventory is unchanged.
- CT-POL-5: runtime DDL, role and metadata denials.

**Migration** (`TestPolicyMigration`, plus a persist seed/verify around a PostgreSQL stop/start):
- MG-POL-1: empty database to v5.
- MG-POL-2: repeated up and status.
- MG-POL-3: a populated v4 database upgrades applying only 00005. Existing Organizations read `true`; every identity,
  Organization, Group, binding, profile and link row is unchanged.
- MG-POL-4: reapply.
- MG-POL-5: down 5→4 with flags set restores the v4 catalog. Predecessor rows are unchanged.
- MG-POL-6: a `false` flag and a `true` flag survive the restart.

### 6.3 B-POLICY-IT (`TestAccessPolicy`, `core/internal/tenancy/access/integration_test.go`)

The fixture:
- two Organizations A and B; Groups g1, g2 and g3 in A, gB in B;
- profiles: P1 shared by g1 and g2, P2 on g3, P3 unlinked, PB in B;
- users covering all six roles, mixed multi-Organization users, a Group-binding-only user and a new-unattached-Group
  owner;
- sessions issued by the real G004 `Login`/`Callback` against `identitytest.Fake` and the WorkOS adapter. The fake's
  default claims include `org_id`, `role: admin` and `permissions: [all]`.

All operations run through a runtime-role pool. State assertions compare scoped row-set hashes, and failure messages
print hashes only (`h()` style).
- (r1) The access Service's pool is wrapped by the counting `hookDB` (§6.4), so binding cases can assert that no
  `BeginTx` or statement was issued.
- Request `ctx`s with chosen correlation values come from the real `observability.Middleware` through
  `httptest`, in test files only.
- Every case obtains its request scope from `BeginRequest` with a real session credential and defers `end`.

| Cases | AC | Content |
|---|---|---|
| IT-CTX-1..6 | AC1, AC2 | Positives for each role through `BeginRequest` → `Resolve`. Absent, garbage, expired, revoked, logout-all and disabled sessions give `unauthenticated` at `BeginRequest`. A session revoked after `BeginRequest` gives `unauthenticated` at `Resolve` in the same scope. Provider claims and role-like display names grant nothing. Malformed, zero, unknown, foreign and non-member selectors give `invalid_input`/`context_denied` (identical for unknown, foreign and non-member). Org admin may select any Group of A; a member only Groups they belong to; a foreign Group is denied. A Group binding without an Organization binding is denied. No row is minted |
| IT-CTX-7 (r1) | AC1 | **Cross-request, same subject.** Two `BeginRequest` calls with one credential give scopes R1 and R2. Run it under two `ctx`s with **equal** inbound `X-Request-Id`, and again under `context.Background()` (**absent** IDs). A context resolved in R1 and used under R2 for each of the 8 operations gives `context_denied` (`context_scope_mismatch`), and the reverse too. Each case has zero `BeginTx` and unchanged row-set hashes |
| IT-CTX-8 (r1) | AC1 | **Cross-subject.** Users A (`org_owner`) and B (`group_member`) begin scopes with equal and with absent IDs. A's context under B's scope gives `context_denied`, for reads and for a write A alone may perform, with zero `BeginTx` and state unchanged; B's under A's likewise. Each context still succeeds under its own scope |
| IT-CTX-9 (r1) | AC1, AC6 | **Legitimate same request.** In one scope: `Resolve` Organization-only, then `GetOrganization`, `UpdateBusinessProfile` and `GetBusinessProfile`. Then `Resolve` selected-Group in the same scope, then `RenameGroup` and `GetGroup`. One call runs under a deadline-derived child `ctx`. All succeed with the expected rows |
| IT-CTX-10 (r1) | AC1 | **Outside or after the request.** A context used under `context.Background()` gives `context_scope_missing`. After `end`, `Resolve` and every operation give `context_scope_ended`. Nested `BeginRequest` gives `context_scope_nested` and does not touch the session row (`renewed_at` unchanged). All have zero `BeginTx` |
| IT-MATRIX-1 | AC3 | The same independent 11×6 table on **DB-loaded** facts (A, with B as foreign), in both context modes; unavailable-store and cancellation injections deny |
| IT-PROFILE-1..5 | AC4 | Owner/admin through g1 or g2 update/select P1; member reads only; no Group role reaches privileged settings, policy or audit. A selected g3 context cannot reach P1 through g1. The manageable set in A excludes PB and member-only profiles and deduplicates P1. A new unattached Group cannot authorize P1. A P1 read exposes no g1/g2 data |
| IT-FLAG-1..5 | AC5 | Default `true` on pre-existing and new Organizations. Owner sets `false` then `true`. Admin, member and Group-role attempts give `forbidden`; cross-tenant gives `context_denied`/`not_found`; in all of these the row is unchanged. `false` denies only Group-inherited updates; Organization paths, reads and selection are unchanged; binding and link row sets are unchanged |
| IT-FACADE-1..6 | AC6 | Every method's allow and deny path, with state unchanged on deny. Definitive aborts (write-statement failure, server-rejected COMMIT, cancellation before COMMIT) leave state unchanged. A lost COMMIT reply (real commit, then a connection error) returns `unavailable`; the authorized row is persisted and reconciled; exactly one attempt is made. A method→capability mapping table test. The bypass/importer evidence is U-ACCESS-BOUNDARY |
| IT-REVOKE-1..8 | AC7, AC1 | (r1) Each case keeps **one unchanged, un-ended request scope**: the reused context and `ctx` pass the binding check, so the denial comes from §4 revalidation. It covers a reused context and session after: owner/admin removal or downgrade, unlink, flag `false`, disable, logout, logout-all, expiry. A fresh `Resolve` in that same scope also denies after each session or user revocation, and after each Organization or selected-Group membership loss. Links and the flag do not affect context resolution, so those cases deny at the operation. Losing the last management path denies update/select and removes the profile from the manageable set, while a g2 or Organization path still allows. Member read is retained. No profile row is updated; no binding is created or synchronized. `EnableUser` does not revive a revoked session. A Group regrant does not create Organization membership |
| IT-MFA-1 | AC9 | Users signed in with different provider methods and MFA-like claims all have `unknown` assurance and identical decisions across all roles. `identity.RequireVerifiedMFA` still denies |
| IT-LOG-1 | AC12 | Allow, deny and failure records are `access.decision` only, with the canaries planted in names, slugs, profile context and display names. (r1) The `BeginRequest`/`Resolve` records carry **no `capability`** and use `established`/`deny`/`failed`. Records from equal-ID requests carry that diagnostic `request_id` without affecting any decision |

### 6.4 B-POLICY-ORDERING (`TestAccessOrdering`, deterministic, no sleeps)

**Barriers:**
- `hookDB`: a test wrapper, like tenancy's `faultDB`, with before/after statement-prefix hooks and two commit modes:
  definitive (rollback + `PgError 40001`) and lost-reply (real commit, then `io.ErrUnexpectedEOF`). (r1) It also
  counts `BeginTx` and statements, which IT-CTX uses, and the rows each locking statement returns, which E-PG-2 uses.
- Lock waits are proven by the accepted `waitBlocked` query (`pg_locks` ⋈ `pg_stat_activity`).
- `DisableUser` is paused inside its transaction by a `pgx.QueryTracer` on a dedicated identity pool. The tracer
  blocks after its sessions UPDATE.

**Orderings:**
- **A (guard first):** the guard pauses after step 7, the writer blocks, the guard commits, then the writer commits.
  Both changes are persisted and the guard's commit precedes the writer's.
- **B (writer first):** the guard pauses after BEGIN and before step 1, the writer commits, then the guard denies with
  its pinned code. The target row-set is unchanged.
- **C (writer in flight):** the writer pauses before COMMIT, holding its row lock; the guard blocks; the writer
  commits; the guard re-checks and denies. The state is unchanged.

| Case | Writer | Guarded op | Orders |
|---|---|---|---|
| O-1 | Logout current | RenameGroup | A, B (single statement: C = A or B) |
| O-2 | Logout all | UpdateBusinessProfile | A, B |
| O-3 | DisableUser | RenameGroup | A, B, C |
| O-4 | Sign-in rotation | RenameGroup | A, B |
| O-5 | Expiry | RenameGroup | clock before or after expiry at the lock point |
| O-6 | RemoveOrganizationMembership | UpdateBusinessProfile (Organization path) | A, B, C |
| O-7 | ChangeOrganizationRole owner→member | SetOrganizationPolicy | A, B, C |
| O-8 | RemoveGroupMembership | UpdateBusinessProfile (Group path) | A, B, C |
| O-9 | ChangeGroupRole admin→member | RenameGroup | A, B, C |
| O-10 | UnlinkGroupBusinessProfile | UpdateBusinessProfile (Group path) | A, B, C |
| O-11 | SetOrganizationPolicy(false) | UpdateBusinessProfile (Group path) | A, B, C |
| O-12 | Removal of the selected-Group binding | UpdateBusinessProfile in selected context | A, B |
| O-13 | Unlink g1 while g2 path holds | UpdateBusinessProfile on P1 | allow in both orders |
| O-14 | Policy update × profile update; rename × profile update | — | both orders; no `40P01` |
| O-15 (r1, proposed under Q15) | Sign-in rotation of the guard's own session with a changed display name | RenameGroup | The rotation pauses after its session UPDATE (identity-pool tracer, as O-3). The guard runs step 1, then blocks at step 2, proven by `waitBlocked`. The rotation then resumes into its display-name UPDATE. Expect exactly one `40P01` victim. If the guard is aborted: `unavailable` (`aborted`), Group row unchanged, rotation committed. If the rotation is aborted: no new session, old session live, guard commits. Either way no unauthorized write. Detection takes PostgreSQL's `deadlock_timeout` (default 1 s; the runtime role cannot lower it). That is a server-side wait, not a test sleep |
| E-PG-3a (r1) | `SetOrganizationPolicy(false)` queued behind guard T1's O share | a second guarded profile write T2 by another member | Characterization (§6.7). Records whether T2's step 3 waited or passed the queued writer, and whether the policy write committed before T2. Asserts only safety: every committed guarded write precedes the policy commit or observes `false` and denies |
| E-PG-3b (r1) | `RemoveOrganizationMembership(m)` queued behind m's guard T1 (OB share) | a second guard T2 of m in another request scope | Characterization, as E-PG-3a. Afterwards a new guard of m gives `context_denied` |

Each case records its interleaving and outcome. The Result names the linearization point and accepted-writer
compatibility (AC8).

### 6.5 Privacy and forced failure

- **Privacy:** B-POLICY-IT and B-POLICY-ORDERING write their logs through `TENANCY_LOG_FILE`. `logcheck validate`,
  canary `scan` and a planted-canary control run as in G006.
- **Catalog:** `events.json` gains `access.decision` with its closed enums. `logrecord_test.go` gains accepted and
  rejected records: unknown operation or reason, an extra field, a UUID-shaped value.
  - (r1) The §3.6 conditional rules use the catalog's existing `rules` (if/then/else) support.
  - Accepted records: context operations without `capability`, and guarded operations with their mapped
    `capability`.
  - Rejected records:
    - a context operation carrying any `capability`;
    - a context operation with outcome `allow` or reason `allowed`;
    - a guarded operation without `capability`, or with outcome `established`.
- **F-1 forced failure:** in a scratch clone, flip one U-POL-MATRIX cell and one IT-MATRIX expectation. `W(unit)` and
  `W(migrations)` must FAIL with exactly those IDs, combined with the inherited B-CI-CONTROLS gate controls.

### 6.6 Mechanical reconciliation sites (same step as the change)

| Site | Change |
|---|---|
| RC-1 `run.sh` B-MIGRATE-SET | Add 00005; pin 00004 digest |
| RC-2 `run.sh` B-DOCS-EXEC-ROOT, `mig_step`s, B-MIGRATE-CLI/CATALOG | Version sequence ending at 5. Reapply and rollback act on 00005 (rollback to v4). Relation counts stay 35. Texts |
| RC-3 `run.sh` B-PERSIST, B-DBRENAME, off variant | Version 5 and `00005 applied` |
| RC-4 migrate `integration_test.go` | `reset()` and `TestMigrationLifecycle` at v5 (subtest count unchanged) |
| RC-5 migrate `business_profile_integration_test.go` | G006 catalog, migration and persist tests pinned with `runnerFor(subsetFS(4))`; every assertion unchanged |
| RC-6 tenancy `integration_test.go` `setup()` | `current_version\t4` → `\t5` (shared with `TestBusinessProfileModel`) |
| RC-7 migrate `runner_test.go` | Normal set adds 00005 |
| RC-8 `run.sh` `TENANCY_PROBE_CORE/WEB` | Add `/internal/v1/policy`, `/internal/v1/context`, `/internal/v1/access`, `/policy`, `/api/policy`, `/api/context` (still 404) |
| RC-9 `run.sh` new functions `policy_migrations`, `policy_model`, `policy_ordering` after `business_profile_model`; `APPLICABLE` adds `docs/foundation/AUTHORIZATION.md` | Same pattern as G006. (r1) `policy_model` runs `TestAccessPolicy` and `TestAccessFootprint` (§6.7) under the B-POLICY-IT ID; `policy_ordering` includes the E-PG-3 cases |
| RC-10 `required-checks.tsv` | +4 `migrations` rows: 86 → 90 rows; full 84 → 88; fast 33 and integration 71 unchanged |
| RC-11 `events.json`, `logrecord_test.go` | §6.5 |
| RC-12 (only if a pin requires it) tenancy `business_profile_integration_test.go` | If a G006 `recordDB` assertion enumerates Unlink's exact statements, add the locking Group precheck to that pin. The foreign/unknown identical-code and Organization-predicate assertions must pass unchanged (allow-listed for this purpose only) |

Any change to a predecessor test's **meaning** stops for escalation. This covers anything beyond its version pin,
tolerated object set, statement counts or the Unlink lock.

### 6.7 PG-001 focused evidence (r1; built and run only after S0)

This is bounded characterization, not a benchmark or a performance-acceptance campaign. It adds no required-check
ID, no harness change and no target. `TestAccessFootprint` lives in `access/footprint_integration_test.go` and runs
in the B-POLICY-IT step; E-PG-3 lives in `TestAccessOrdering`.

**Dataset.** It is seeded with owner SQL in the disposable test database, in its own Organizations X and Y, then
`ANALYZE`d as the owner. The runtime role cannot ANALYZE; every measured statement still runs as the runtime role.
- X has 200 Groups, 400 Organization members and 4,000 Group bindings.
- Subject s1 holds n_g = 1 Group binding in X, and subject s50 holds n_g = 50. Each also holds 50 Group bindings in
  Y, so N_g = n_g + 50.
- An X admin also holds 50 Group bindings in X.
- The sessions are real G004 sessions.
- These **characterization points only distinguish subject-scoped from Organization-wide access.** They are not
  representative cardinalities and not targets; the owner may replace them through Q13. The dataset is about 5,000
  rows.

**E-PG-1: plan summaries (recorded; no pass/fail threshold).**
- Every statement 4–9 of §4.7.2 is run in both context modes, at s1 and s50, as the runtime role, under
  `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)` inside a transaction that is rolled back. ANALYZE executes the locking
  statements too, and their locks end with the rollback.
- Recorded per statement:
  - node types, relation and index names;
  - actual rows, loops, and rows removed by filter or recheck;
  - shared buffers hit and read;
  - planning and execution time, labelled with the environment (host type, PostgreSQL image digest).
- Conditions, constants and parameter values are dropped, so summaries hold no UUIDs or fixture values (R9).
- Each summary is emitted as one JSON line on the test's stdout, through the accepted report stream, with no new bind
  mount.
- **Escalation trigger:** at the s50 point, statement 4 examines ≥ M_org entries, or statement 5a/5b ≥ B_org entries
  (Organization-wide), or any of statements 4–7 uses a sequential scan. Claude then records it, finishes the current
  step, and **stops before adding any index or reshaping SQL**, reporting to Codex and the owner (§11).

**E-PG-2: locked-row inventory (asserted).** The check covers each guarded write in both context modes for s1, s50
and the X admin.
- The `hookDB` row counts of the locking statements must equal §4.7.1: U, S, O and OB once each; GB and G equal to the
  candidate set; plus the target.
- An Organization-only profile write locks exactly 4 + 2·n_g rows plus the updated profile.
- No statement other than the planned ones carries a locking clause.

This verifies the Plan's footprint claim. It is not a performance target.

**E-PG-3: lock-queue characterization (deterministic; behaviour recorded, safety asserted).** E-PG-3a and E-PG-3b
are in §6.4.
- Each records whether a later compatible guard waited, or passed a queued policy or revocation writer (H-4).
- It asserts only R6 safety and final-state consistency.
- `waitBlocked` and hooks are the only synchronization; there are no sleeps.

**Timing.** No timing campaign runs. The only times collected are those EXPLAIN ANALYZE reports in E-PG-1, recorded as
environment-specific observations. The Result records all three with PG-001's link, and Codex or the owner set any
disposition or follow-up in the register.

## 7. Validation, CI profile and hosted evidence

- **Before every local batch** (development, focused or final):
  - `git fetch origin`;
  - verify `origin/staging` is an ancestor of HEAD, or rebase and reconcile bindings first;
  - `bash infra/ci/docker-desktop-mounts.sh` (do not launch on failure; copy needed `/tmp` evidence out first);
  - run under `flock <home>/.local/state/platform-evidence/local-harness.lock` (the accepted shared lock);
  - record target and candidate SHAs.

  Never edit the repository during a run (B-RENDER-PLAN).
- **Command form:** the accepted wrapper, unchanged (G006 Plan §10):
  `bash infra/validation/run-approved.sh --snapshot <S> --snapshot-sha256 <D> --count <N> --path-re '<RE>'
  --plan management/v0/plans/G007-trusted-context-policy-plan.md -- --base <BASE> --allow-file
  management/v0/results/G007-scope-allowlist.txt --report-dir <EVID>/<step> <suites>`.
  `controls` always runs with a secret-generating suite, and every run includes `final`. Ownership cleanup uses
  `infra/ci/cleanup-owned.sh` and the harness labels, as accepted.
- **Focused inner loop:** `W(unit)` for the evaluator and boundary. `W(migrations)` for catalog, model and ordering.
- **CI profile actually selected:** **full**. Changed paths are `core/migrations/`, `core/internal/migrate/`,
  `core/internal/tenancy/**` (foundation), `core/internal/identity/` (runtime) and `infra/` (foundation); `docs/**/*.md`
  and `management/**` are docs. The union is full, and no mapping row changes.
- **No manual-testing or operator-feedback stage:** there is no UI, browser workflow or operator qualification. The
  final qualification therefore has no feedback gate.
- **Final local coverage** of the rebased candidate before the MR: `W(all)` plus `W(unit)`. Reuse unchanged evidence
  only with byte-identity provenance.
- **Hosted:**
  - The MR `fast` job runs automatically. The MR `gate` is manual/optional and is not started.
  - The automatic staging `gate` (full) after an owner-authorized conditional integration supplies hosted evidence,
    judged by the owner status rule. `canceled` is recorded as canceled, never as a test pass; any failed job or known
    failing required check fails.
- **Not required:** WorkOS, `identity-qualification` or second hosted-full campaigns.
- **G051:** independent. Whichever integrates second reconciles `run.sh`, `required-checks.tsv` and controls
  mechanically. Its unaccepted proposals are not assumed.

## 8. File inventory and scope allow-list

**Create:**
- `core/migrations/00005_organization_policy.sql`;
- `core/internal/tenancy/facts.go`;
- `core/internal/tenancy/access/{access,context,policy,facade,tx,errors,log}.go`;
- `core/internal/tenancy/access/{policy,matrix,context,boundary,integration,ordering_integration,footprint_integration}_test.go`
  (r1 adds `footprint_integration`; the existing allow-list regex already covers it);
- `core/internal/migrate/policy_integration_test.go`;
- `docs/foundation/AUTHORIZATION.md`;
- `management/v0/results/G007-scope-allowlist.txt`;
- `management/v0/results/G007-trusted-context-policy-result.md`.

**Modify:**
- `core/internal/identity/{service,store}.go`;
- `core/internal/tenancy/{store,service,business_profile,association}.go`;
- `core/internal/tenancy/{boundary,integration}_test.go` (and `business_profile_integration_test.go` only under RC-12);
- `core/internal/migrate/{integration,business_profile_integration,runner}_test.go`;
- `infra/validation/{run.sh,required-checks.tsv}`;
- `infra/validation/tools/internal/logrecord/{events.json,logrecord_test.go}`;
- `docs/foundation/{TENANCY,MIGRATIONS,IDENTITY,FOLLOW-UP}.md`.

**Untouched:**
- `core/cmd/**` and every binary;
- web, intelligence and contracts;
- migrations 00001–00004;
- `.gitlab-ci.yml` and `infra/ci/**`;
- `infra/validation/{lib.sh,run-approved*.sh,render-customizations.tsv}`;
- `skills/**`;
- governance, ADRs, Goal/Spec, ROADMAP/PROGRESS and SECURITY.md.

```text
^core/migrations/00005_organization_policy\.sql$
^core/internal/identity/(service|store)\.go$
^core/internal/tenancy/(facts|store|service|business_profile|association)\.go$
^core/internal/tenancy/(boundary|integration|business_profile_integration)_test\.go$
^core/internal/tenancy/access/[a-z_]+\.go$
^core/internal/migrate/(integration|business_profile_integration|runner|policy_integration)_test\.go$
^infra/validation/(run\.sh|required-checks\.tsv)$
^infra/validation/tools/internal/logrecord/(events\.json|logrecord_test\.go)$
^docs/foundation/(AUTHORIZATION|TENANCY|MIGRATIONS|IDENTITY|FOLLOW-UP)\.md$
^management/v0/results/G007-(scope-allowlist\.txt|trusted-context-policy-result\.md)$
```

`AUTHORIZATION.md` covers:
- the matrix and interpretation pins;
- context and selector semantics, including (r1) the request-scope contract that future adapters must honour:
  one `BeginRequest` per request, `end` deferred, and correlation IDs never used as a binding;
- the operation→capability mapping;
- the §4 ordering and failure contract, and (r1) the §4.7 footprint summary with the PG-001 link;
- errors, reasons and the event;
- the **G008 handoff**:
  - tenant relations: `organizations` (now including the flag), `groups`, `role_bindings` and both projections,
    `business_profiles`, `group_business_profiles` and `business_profile_access_paths`;
  - the global identity tables read inside tenant transactions (`users`, `sessions`);
  - every statement used by `ReadFacts`, the `*Tx` writers and the identity checks;
  - the single transaction-owning call site in `access/tx.go` where G008 initializes transaction-local context;
  - that raw Service methods remain trusted-only.

## 9. Acceptance-criteria mapping

| AC | Evidence |
|---|---|
| AC1 | IT-CTX-1..10 (r1: 7 cross-request with equal and absent IDs, 8 cross-subject, 9 legitimate same-request, 10 outside or after the request and nested), U-CTX-1..9 (zero `BeginTx` on every rejection), IT-REVOKE in an unchanged scope, U-ACCESS-BOUNDARY (no header or claim input path; correlation only in `log.go`) |
| AC2 | IT-CTX-3..6, row-set hashes (no membership minted), U-POL-GATE |
| AC3 | U-POL-MATRIX (independent literal) + IT-MATRIX-1 (DB facts), U-POL-GATE/SCOPE; store-error and cancel denials |
| AC4 | IT-PROFILE-1..5, U-POL-SCOPE; there is no profile-role input type |
| AC5 | §3.4; IT-FLAG-1..5; CT-POL-2/3; MG-POL-3/6 (restart) |
| AC6 | IT-FACADE-1..6 (definitive vs lost-reply), §4.5, mapping table test, U-ACCESS-BOUNDARY + B-MCP-SOURCE (no bypass wiring) |
| AC7 | IT-REVOKE-1..8 |
| AC8 | B-POLICY-ORDERING O-1..O-14 (+ O-15 if Q15 accepts it); E-PG-2 locked-row inventory; Result states the §4.1 linearization and §4.3 compatibility |
| AC9 | IT-MFA-1, U-POL-MFA; inherited B-IDENTITY-* verified-MFA and cancellation controls unchanged |
| AC10 | CT-POL-1..5, MG-POL-1..6, CT-POL-4 detector (no copied grants), reconciled B-MIGRATE-*/B-PERSIST/B-DBRENAME/B-TOGGLE-OFF, 00001–00004 digests |
| AC11 | B-TENANCY-HTTP (RC-8), B-MCP-SOURCE, U-ACCESS-BOUNDARY, CT-POL-4 (no RLS or new object), AUTHORIZATION.md G008 inventory; RLS isolation not claimed |
| AC12 | IT-LOG-1 + logcheck validate/scan/planted, B-SECRET-CANARY, inherited suites (auth, browser, theme, scoped tokens, contracts, variants, dependencies) on the final tree |
| AC13 | Wrapper on the real root and hash-identical copies; final local `W(all)` + `W(unit)` on the rebased candidate; F-1 + B-CI-CONTROLS; MR `fast`; staging disposition under the owner rule |
| AC14 | E4 binding, wrapper pre/post, NUL-safe status, allow-list, owned cleanup, failed/interrupted runs, rebase records, B7 accounting, Result, Codex ACCEPT; no G008/G010/G021 claim |
| PG-001 (ARCHITECTURE §63; not an acceptance criterion or target) | §4.7 assessment; E-PG-1..3 recorded in the Result with the PG-001 link; owner dispositions of Q13/Q14 |

## 10. Sequence and delegation

| Step | Work (tests and wiring land with their change) | Batch |
|---|---|---|
| S0 | Start gate (§2) | — |
| S1 | Allow-list copy; migration 00005; CT-POL/MG-POL; RC-1..5, RC-7; 2 registry rows | `W(unit lint controls migrations)`, then `W(root ops rename off)` |
| S2 | identity `Lock/CheckLiveSession`; tenancy `facts.go`, `*Tx`, Unlink lock, U-BOUNDARY, RC-6 | `W(unit)`, then `W(migrations)` (G005/G006 models still pass) |
| S3 | access package (including the r1 request scope and B1–B6 binding) and unit tests (U-CTX-1..9); delegate's `matrix_test.go` integrated | `W(unit)` |
| S4 | B-POLICY-IT (including IT-CTX-7..10 and `TestAccessFootprint` E-PG-1/2), then B-POLICY-ORDERING (including E-PG-3, and O-15 if accepted); RC-8..11; 2 registry rows | `W(unit migrations)`, then `W(lint controls contracts root ops off)` |
| — | (r1) **E-PG-1 trigger check** (§6.7): on a trigger, stop and report before any index or SQL reshaping | — |
| — | **Cycle-4 checkpoint** (expected after S4) | §11 |
| S5 | Read-only security/SQL review subagent; bounded corrections | covering suites |
| S6 | Docs (AUTHORIZATION, TENANCY, MIGRATIONS, IDENTITY, FOLLOW-UP STAGE-POLICY annotations) | `W(lint provenance migrations off)` |
| S7 | F-1 forced failure in a scratch clone | expected FAIL recorded |
| S8 | Fetch/rebase/reconcile; then final `W(all)` + `W(unit)` | all 88 full IDs + B-UNIT-CHECKS PASS |
| S9 | Target re-check; X1 push + MR; `fast`; Result commit | `fast` PASS |
| S10 | Codex review → X2 conditional integration → staging gate → Result update | owner status rule; Codex ACCEPT |

**Delegation** (shares B7; Claude main owns integration and every other file):
- **Oracle delegate:**
  - Owns only `access/matrix_test.go`.
  - Writes it from Spec R3, §3.3's pins and the pinned `decide`/`Facts` signatures, **without reading `policy.go`**,
    so that U-POL-MATRIX and IT-MATRIX are independent of the implementation (AC3).
  - Disagreements are resolved by Claude against the Spec, or escalated. Neither side is edited to match the other.
- **Read-only security/SQL reviewer** after S4. It reviews lock order, predicates, privilege use, error
  non-disclosure, log fields and the bypass boundary, and (r1) the scope binding and the E-PG evidence. It has zero
  cycles unless its findings need corrections.
- **Declined:** an implementation delegate for tenancy/identity/access code. The ordering contract couples those files.

## 11. Budget, checkpoint, stops and external actions

**B7: at most seven shared cycles, checkpoint after cycle 4. Forecast 3–5; feasible without a split.** Calibration:
G005 (B7) and G006 (B5) each consumed 1 cycle with similar harness ripple. The new risk is concentrated in locking
semantics and deterministic barriers. r1 adds an in-memory binding check, one footprint test file and two or three
deterministic ordering cases. None of them changes the forecast, though E-PG and O-15 add to C3's load.

| Cycle | Likely material issue |
|---|---|
| C1 | Version-5 ripple: lifecycle/reset, G006 pins, `mig_step` texts |
| C2 | `ReadFacts` locking: FOR SHARE through projections (privilege/plan), predicate re-check, Unlink lock effect on G006 tests |
| C3 | Ordering barriers: tracer pause, hook prefixes, `waitBlocked` attribution; (r1) E-PG dataset/ANALYZE and the O-15 deadlock victim handling |
| C4 | logcheck/events, security-review corrections |
| C5 | Rebase reconciliation with integrated G051 harness changes |
| C6–C7 | Reserve |

**Checkpoint after C4:** record the ACs met and remaining, repeated failures and convergence.
- If AC8 is not converging, stop and propose the split: **G007a** (migration, evaluator, request scope and resolver,
  guarded reads, policy read/update) and **G007b** (guarded Group/profile writes, the Unlink lock, O-1..O-14, plus
  r1's O-15 and E-PG-3).
- No split is recommended now: the guarded writes and their ordering proof are one guarantee.

**Stop and escalate** (in addition to the Goal's list):
- The projections cannot be row-locked by the runtime role. The fallback of locking `role_bindings` directly is a
  U-BOUNDARY rule change that needs review.
- Any writer other than Unlink needs a change.
- A predecessor test changes meaning.
- A `40P01` appears in a case §4.4 claims impossible. r1 lists one possible cycle; any other `40P01` stops.
- A need for a grant widening on `group_business_profiles`, a table lock, SERIALIZABLE retries, an invalidation
  service, RLS or a public surface.
- (r1) The binding needs anything beyond §3.2: a transport adapter, request middleware, a correlation-ID comparison,
  a provider claim or a stored authority snapshot.
- (r1) An E-PG-1 trigger (§6.7). Also stop if the evidence or the owner's Q13/Q14 answers call for:
  - a new or changed index;
  - a lock-mode, grant or schema change;
  - adopting M1/M2 when Codex has not directed it;
  - a bounded revocation-latency guarantee.

  These are architecture, schema or scope questions and are not resolved within G007.
- B7 exhaustion or thrashing.

**Named external actions** (none performed; each separately authorized):
- **X1:** push the rebased branch (proposed `g007-trusted-context-policy`) and
  `glab mr create --target-branch staging`, keeping source-branch deletion. The manual MR gate is not started.
- **X2:** owner-authorized conditional integration after Codex pre-integration review; verify branch deletion; the
  staging gate follows.
- No WorkOS, protected-ref, schedule, settings or runner action.

## 12. Questions for Codex

**Status in r1:**
- The r0 review settled Q1–Q8 and Q10, and r1 keeps them unchanged; Q8 needs no owner sign-off.
- Q9 is answered by §3.2, and Q11 is updated for `BeginRequest`.
- The review did not address Q12, which is unchanged apart from a note.
- Q13–Q17 are new. Q13 and Q14 are owner-visible PG-001 questions; Q15–Q17 are for Codex.

| ID | Question | Proposal |
|---|---|---|
| Q1 | Package placement | `core/internal/tenancy/access`: existing CI row, `APPLICABLE` and B-MCP-SOURCE cover it, so no mapping change. Alternative: top-level `core/internal/access` plus a profile row, `APPLICABLE` and a B-MCP-SOURCE regex change |
| Q2 | Composition seam | tenancy exports `ReadFacts` + four `*Tx` writers that reuse the accepted SQL and codes; identity exports two tx-scoped session checks. Bypass control is the importer boundary test plus B-MCP-SOURCE, not Go visibility. Alternative: the facade inside `tenancy`, with identity behind an injected interface (larger change to tenancy's pinned purpose and imports) |
| Q3 | Flag storage | A NOT NULL DEFAULT true column on `organizations` (no missing-row state; one column grant). Alternative: an `organization_policies` table (row materialization, phantom-insert ordering, a new RLS relation) |
| Q4 | Association lock surrogate | Runtime cannot row-lock `group_business_profiles`. Unlink therefore locks its Group row `FOR NO KEY UPDATE` and the guard locks it `FOR SHARE`. Rejected alternatives: granting UPDATE on an association column (widens write privilege); `LOCK TABLE` (global serialization); SERIALIZABLE (unguarded writers stay READ COMMITTED) |
| Q5 | Read path | REPEATABLE READ READ ONLY snapshot without row locks; linearized at the snapshot |
| Q6 | Error contract | §3.6: `forbidden` only when the target is readable, else `not_found`; uniform `context_denied`; payload validation before authorization |
| Q7 | Manageable list | Unexported `manageableProfiles` (evidence for AC4/AC7 only); no exported list until G018. Semantics: §3.3(d) |
| Q8 | Matrix interpretation | Pins (a)–(d) in §3.3. **Settled by the r0 review:** they follow Spec R3 with no new product decision |
| Q9 | Context request binding | **Answered in r1 (C1), §3.2.** The binding is the unforgeable authenticated request scope: one `BeginRequest` per request, a fresh unexported scope in `ctx`, pointer identity, and `end`. Each context is bound to its scope and carries the scope's subject and session. B1–B6 run before `BeginTx`. Correlation IDs are diagnostic only and are never compared or stored. Per-operation revalidation (§4) still decides authority. r0's request-ID equality is withdrawn |
| Q10 | Commit ambiguity | §4.5: a `PgError` on COMMIT is definitive; anything else is unknown. Both are external `unavailable`. No retry or idempotency key; absolute-assignment convergence is stated with its last-commit-wins limit |
| Q11 | Session renewal | (r1) `BeginRequest` renews through `Session`, once per request at most, under G004's 60 s throttle. `Resolve` and guarded operations check liveness and expiry but do not renew |
| Q12 | Required IDs | B-POLICY-CATALOG, -MIGRATION, -IT and -ORDERING in `migrations`; registry 90, full 88. r1's E-PG evidence adds no ID |
| Q13 (owner, PG-001) | Workload assumptions (§4.7.5) | Typical and upper values are needed for: Groups per Organization; a subject's Group bindings per Organization (n_g) and across Organizations; Organizations per subject; members per Group; per-Organization and per-subject guarded-write concurrency; any latency or lock-wait target. **Proposal:** answering "unknown for now" is acceptable. G007 then proceeds with the §6.7 characterization points, sets no target, and records the gap in PG-001. If the owner expects n_g in the hundreds or more, Codex should weigh Q17 before implementation |
| Q14 (owner, PG-001) | H-4: shared-lock writer delay (§4.7.4) | Sustained overlapping guarded writes may delay a queued policy, removal, disable or logout write beyond one transaction, including through the revoked subject's own concurrent writes. Ordering safety (R6) holds; latency is unbounded by G007. **Proposal:** accept for G007 as a recorded PG-001 risk with E-PG-3 characterization. G007 has no adapter or untrusted concurrency source, and future adapter Goals own request concurrency, rate limits and deadlines. Requiring bounded revocation latency instead would be a Spec/design change: exclusive locks, timeouts or predecessor-writer changes. That would return to planning |
| Q15 (Codex) | Sign-in rotation deadlock (§4.4) | r0's "no wait cycle" claim missed the rotation's conditional display-name UPDATE. **Proposal (a):** accept it as a rare definitive abort, as the safety analysis shows, and add O-15. Rejected alternatives: (b) reorder the sign-in writer, a predecessor change outside R6's minimum; (c) reverse the guard's lock order, which creates a cycle with `DisableUser` |
| Q16 (Codex) | PG-001 evidence form (§6.7) | Accept E-PG-1 (recorded plan summaries with an escalation trigger), E-PG-2 (asserted locked-row inventory) and E-PG-3 (deterministic characterization). Codex may narrow them, for example by dropping E-PG-3 if the owner accepts Q14 without characterization. None is a benchmark or target |
| Q17 (Codex, conditional) | Footprint mitigations M1/M2 (§4.7.6) | **Not adopted in r1.** Both are internal query-strategy choices: no schema, grant or ordering change, and least privilege kept. They would change §4.1 steps 5–7 and the affected O cases. **Proposal:** keep r0's lock set unless Q13's answer or an E-PG-1 trigger shows the n_g growth matters. Then adopt M2 (and optionally M1) through a reviewed Plan amendment |

**Context and scope:** one bounded Goal within B7. G021 stays closed. No product, UI, route, RLS or provider work.
