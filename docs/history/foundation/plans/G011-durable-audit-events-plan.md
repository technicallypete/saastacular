> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G011 — Durable audit and event conventions: Implementation Plan

Revision: r2, 2026-10-14 UTC. Status: **provisional, review-ready for Codex Plan re-review**. This is not approval, not
an execution binding and not implementation. Execution waits for **final G010 ACCEPT**, explicit Codex Plan ACCEPT and a
fresh Git-first execution binding (§0). Tier: architecture/security/data (Goal Execution Policy).

r2 is the one-item revision for the r1 review (not included) (CORRECTION LOOP, SHA-256
`7ed6483a12e71203255024d824533cfe206742e7123cf06741046825da547989`). r1 is archived byte-identically as
G011-durable-audit-events-plan-r1.md (not included) (SHA-256
`dd4406b57a4df90c2b198eb3fcb738c3b43abbb6caf176a9fd19e915100166c6`). The r1 review resolved the other five r0
correction areas.

**r2 change map:**

| r1 review item | Change | Where |
|---|---|---|
| Skip the empty strict audit statement | The strict `superseded` INSERT runs only when the transition tables contain a `superseded` transition. The guard inspects transition tables only and never touches `audit.events`, so an audit privilege fault cannot abort a logout/disable before the M5 exception block. The renewal-cost text is updated to match | §5.3, §11 |
| Minimum execution validation | AT-ATOM gains an audit column-INSERT privilege fault (revocation UPDATE privileges intact, grants restored and re-verified afterwards): logout/disable commit and report `audit.gap` `failed`; sign-in and `superseded` rotation fail atomically. This is a test scenario executed only after implementation release | §9 |
| Forecast (non-blocking) | The local machine-time range is 64–105 min (about 1.1–1.75 h) | §10.1 |

r1 itself was the bounded revision for the r0 review (not included) (CORRECTION LOOP, SHA-256
`968ed5ad8ef3a8c3eafe48dbf39d88beb099cae3c9e70b758af84d35abef635f`). r0 is archived byte-identically as
G011-durable-audit-events-plan-r0.md (not included) (SHA-256
`76929a2a931d9c6947d9d4ff71c2d727fa5f2ebf528f8c75f81e5ba4e916301a`). Only the affected sections change. Twelve ACs and
B7 with a checkpoint after attempt 4 are unchanged; no execution attempt is spent.

**r1 correction map (retained for context; resolved per the r1 review):**

| Review item | Change | Where |
|---|---|---|
| 1 Group attribution | A direct-append Group must be a row of `app.groups` in the event's Organization, read under the recording transaction's RLS. Organization owners/admins see every Group of their Organization; foreign and absent Groups are refused. New AT-FMT negatives and an owner/admin positive. | §4.4, §5.3, §9, §11 |
| 2 Bounded identity recorder | The identity recorder gets its own begin, a bounded rollback and connection disposal (`recordAuditEvents`). Unrelated identity transactions are unchanged. A unit case in the existing `identity_test.go` and an AT-ATOM stall case reuse the G009 freezer. | §4.3, §6, §9 |
| 3 Owning-transaction helper | `Store.inTxAudited`, the smallest extension of `Store.inTx`, owns `signIn`'s mutation and its required append in one place for both success branches. | §6 |
| 4 Revocation priority (owner decision) | Narrow trigger-level protection for revocation transitions only, a transaction-local gap marker, a post-COMMIT bounded independent record (G6 `revocation.unaudited`) and an `audit.gap` diagnostic. All other mutations stay strictly atomic. Replaces r0 E7. | §2, §3, §4.1, §4.2, §5.3, §5.4, §6, §9, §14 |
| 5 No duplicate b3 | b3 is folded into b4; b1/b2, the security review and affected failure reruns are kept. Forecast updated. | §10.1, §12 |
| 6 Fan-out and costs | `revokeAll`/DisableUser revoke every `revoked_at IS NULL` session, including expired-but-unrevoked rows, with no count cap. Corrected round trips and reader calls; AT-ID counts an expired-unrevoked example. | §2, §11, §9 |
| Rulings Q1, Q3–Q6, Q8 | Applied as written: retention `USAGE` deferred; `system` only from the lifecycle producer's context. Q2 is decided by item 4. Q7 awaits the owner. | §2, §5.4, §15 |
| G010 integration | G010 is merged at `c961e56` (final ACCEPT pending). A-G010-1..4 are reconciled against the integrated bytes. | §0 |

- **Goal / Spec / ADR (canonical, updated 2026-10-14):**
  - [Goal](../goals/G011-durable-audit-events.md) SHA-256
    `a2e25e39a58ca49af4bb77cadb57f0d2da64751176398795b49bb406eae27eb5`;
  - [Spec](../specs/G011-durable-audit-events-spec.md) SHA-256
    `c8aacafbbf9657d1051702b8d01fd30c33e80a90945a3c882b3fae5d1e892784` (R4/AC5 revocation exception);
  - [ADR-006](../../../adr/ADR-006-audit-schema-boundary.md) SHA-256
    `dc7eda774dc06c8ca1eafc52e723772d4e4c8f32d3141187759b9c47cadabc2a` (owner clarification).
  - The r0 inputs were Goal `a4e8d5ac…` and Spec `209ab924…`.
- **Planner:** task G011, role plan; native isolated worktree `.claude/worktrees/agent-<id>` (the Goal's
  one-planner native-path exception; isolated workers cannot switch worktrees); branch `g011-plan`. Effort xhigh as
  instructed; the runtime setting cannot be verified from inside the worker (unknown).
- **Baseline (Git-first):** r2 is rebased onto `origin/staging` `102fd16d6eaf949c04d734e692bc39bc6d528557`. Its delta
  from r1's base `74a3ffcdc0efeddcf659f9d1409045952ecdd2c2` is management/risk only: the r1 review, PERFORMANCE and
  PROGRESS. The base:
  - This includes the G010 merge `c961e562f1041b51d5f45d0e9bbf172813c030d1`, whose core, infra and foundation-doc bytes
    equal the reviewed candidate `71c8c47`.
  - It also includes the r0 review record with the Goal/Spec/ADR/PERFORMANCE amendments.
  - r0 was written on `1849bd4`, dispatched on `31963ae`.
  - No snapshot, probe, build, test or Docker command was used.
- **Service guidance:** core has no service-specific CLAUDE.md or skill; `docs/SECURITY.md` is empty (no added policy).
  Read: ARCHITECTURE §3/§21/§21.1/§56/§63, ADR-003/004/006, DECISIONS, PERFORMANCE PG-001..004, ROADMAP, PROGRESS,
  BUDGET_POLICY, foundation AUTHORIZATION/DATABASE-ISOLATION/OBJECT-STORAGE/IDENTITY/TENANCY/MIGRATIONS, the G009/G010
  Plans, and the code of migrations 00001–00007, `tenancy/access` (tx/facade/context/errors/log/objects), `tenancy`
  (facts/store/service) and `identity` (service/store/logout), plus the harness pins they touch.

## 0. Gates and G010 reconciliation (provisional assumptions)

| Gate | Status at r2 |
|---|---|
| G010 final ACCEPT (EPIC-003 disposition) | **pending**; merged to staging as `c961e56` (MR !26, conditional integration); post-merge staging gate, cleanup and epic closeout pending |
| Codex Plan ACCEPT | pending (this r2) |
| Owner PG-004 (Q7) disposition on the §11 estimate | pending |
| Git-first execution binding, allow-list (§10.6) and initial Result imported by Codex | pending |

**Provisional G010 assumptions (label: A-G010-n).** r1 reconciles them against the **integrated** staging bytes
(`c961e56`, equal to candidate `71c8c47` for core, infra and foundation docs). They are re-verified at binding against
G010's final accepted state, including any correction.

- **Integrated paths:** `attack_expected_test.go`, `attack_integration_test.go`; the `run.sh` `obj_it` second
  invocation (`it_count == 17`, `ISO-ORACLE {"cases":82,"disagreements":0}`, `rls_logcheck` over both logs);
  AUTHORIZATION.md "Isolation attack suite (G010)"; one line each in OBJECT-STORAGE.md and FOLLOW-UP.md.
- **How G011 avoids breaking it:**
  - G011 edits none of these files.
  - The suite reaches version 8 only through `setupObjects` → `access/integration_test.go`'s latest-version check, a
    §10.2 RC row, so it runs with triggers active.
  - Its `B-OBJ-IT` PASS text "at version 7" becomes "8" in the same RC row.
  - It reads no log records itself. `rls_logcheck` validates its log against events.json, which G011 extends with
    `audit.record_failed` and `audit.gap`.
  - Its staged E1/E3 chains gain only database-side G4/G5 records: no provider call, and no `app`/MinIO state they
    probe.
- A-G010-1: G010 adds no migration; the accepted predecessor is version 7, so G011's migration is
  `00008_audit_events.sql`. At `71c8c47` there is no schema or production change. If G010 adds a migration, G011 takes
  the next number with no other change.
- A-G010-2: G010 changes only two new tests (`attack_expected_test.go`, `attack_integration_test.go`), `run.sh` `obj_it`
  (+1 invocation, 17 subtests, `ISO-ORACLE` 82/0), AUTHORIZATION/OBJECT-STORAGE/FOLLOW-UP prose and its Result. There is
  no new check ID. G011's harness edits are in other functions; textual conflicts are reconciled by rebase.
- A-G010-3: G010's state probes (`e.state`, `objectRows`, `allObjects`) cover only `app` tables and MinIO. Its `Tx`
  effect is boolean (begun / not begun), so an extra transaction after a denial that already began one does not change
  it. Denials with `Tx: not begun` are binding/input failures, which G011 never records. Its single-connection D2 case
  checks the backend PID and the four `app.*` settings only.
  - G011 adds one short audit transaction after a denied guarded/storage call releases its connection (§4.3). On a
    one-connection pool it reuses the same backend and leaves the settings empty.
  - Any G010 pin this changes is reconciled under §10.2's RC rule (guarantee preserved, never weakened) or returned to
    Codex.
  - r1 reconfirms all of this on the integrated test bytes (`atkSnap`, `atkTx`, D2's `atkBackend`).
- A-G010-4: G010 corrections do not change context, policy, liveness, RLS readers, storage lifecycle or fixtures.
  Any correction made before final ACCEPT is assessed against this Plan before binding.

## 1. Implementation approach

**One durable boundary, two recording mechanisms, no new role or definer.**

1. **Schema.** ADR-006's `audit` schema holds one append-only table, `audit.events`, owned by `<db>_owner` (never a
   runtime role), with FORCE row-level security, a closed vocabulary and no foreign keys.
   - The owner lacks `CREATE` on the database, so **provisioning** creates the empty schema
     (`10-roles.sh`, `provision-roles.sql`). Migration 00008 creates every object in it (escalation E1).
2. **Database triggers own every committed row change** (owner preference). Eighteen capture triggers on eight `app`
   tables (17 `AFTER ROW` and one `AFTER STATEMENT` with transition tables for session revocations) call owner-owned
   SECURITY INVOKER functions. Each copies allow-listed fields of the changed row into `audit.events` in the same
   transaction. WHEN clauses and `UPDATE OF` column lists skip no-op updates and unaudited columns on the row triggers.
   - **Revocation priority (owner decision, r1).** Only the local revocation transitions (session `revoked_at`
     NULL→set for `logout|logout_all|provider_logout|disabled`, and user `active→disabled`) wrap their audit insert in
     a bounded exception block. An audit-only fault then rolls back just that insert, sets a transaction-local gap
     marker and lets the revocation commit.
   - Identity reads the marker before COMMIT and, after a successful COMMIT, attempts an independent G6 record and
     emits an `audit.gap` diagnostic.
   - Every other mutation, including sign-in rotation (`superseded`), enable and grants, stays strictly atomic.
3. **Go owns non-row events:** the authentication decision at sign-in, security denials, signed-URL issuance and
   external storage outcomes. Go appends through one primitive, `audit.Append` (one `INSERT`), either inside the owning
   transaction (sign-in success) or in a short independent audit transaction after the business transaction ended.
4. **Controlled append in the database.** Runtime receives `INSERT` on 12 descriptive columns only. A `BEFORE INSERT`
   trigger sets every server-owned field (id, version, time, actor, correlation) from the transaction context and
   enforces per-action shapes. It also enforces **source separation** with `pg_trigger_depth()`: trigger-owned actions
   are accepted only from a capture trigger, and Go-owned actions only from a direct insert, so Go can neither forge nor
   duplicate a trigger-owned event (escalation E2).
5. **Trusted attribution** comes only from the G008 context statement. It gains two transaction-local correlation
   settings and the reset of the revocation gap marker, all in the same round trip. The actor is never a supplied column
   (escalation E4).
6. **Reads.** One internal guarded operation, `ReadAuditEvents`, uses the existing `audit.read` capability (current
   `org_owner` of the active Organization). RLS enforces the same rule. Global rows are never tenant-visible.

**Rejected alternatives (recorded):**
- Combined mutation/audit SQL functions: the same atomicity as triggers, but each existing mutation statement and call
  site would be rewritten. Triggers add zero client round trips and change no call site.
- A SECURITY DEFINER writer role: it needs a new provisioned role, owner→role SET chains, an exception to the
  "SECURITY DEFINER only in 00006" boundary rule and harness PROV pin churn, for no extra guarantee against the trusted
  runtime writer (Spec R2 trust limit).
- Go appends for row changes: this contradicts the owner preference and adds a round trip per mutation.

## 2. Event catalog (finite; R1, R2)

**Fields** (`audit.events`): `id` uuid (server); `schema_version` 1; `occurred_at` (server clock at append); `action`;
`outcome` ∈ {success, denied, failed, unknown}; `reason`; `operation` (`access.denied` only); `actor_kind` ∈ {user,
system, unknown} and `actor_user_id` (set exactly for `user`); `organization_id` and `group_id` (Group only with an
Organization); `resource_type` ∈ {organization, group, role_binding, business_profile, user, session, stored_object} and
`resource_id` (both or neither); `subject_user_id`; `role`/`previous_role` (six role codes); `ttl_seconds` 1–300;
`request_id` `^[A-Za-z0-9._-]{8,64}$` and `trace_id` `^[0-9a-f]{32}$` (diagnostic, sanitized to NULL when malformed).
No names, slugs, emails, provider subjects, keys, URLs, content types, sizes, payloads or error text.

**Actor rule (BEFORE trigger; Q3 ruling applied):**
- **Lifecycle actions** are `user.disabled`, `user.enabled`, `session.revoked` with reason `disabled`, and
  `revocation.unaudited` with reason `disabled`. They are recorded as `system` **only** when produced inside the
  explicitly defined trusted lifecycle producer: `DisableUser`/`EnableUser`, which run in the target's IDENTITY shape,
  so `ctx_user` is set and is the target.
  - The affected user is never recorded as the initiator.
  - With no context (privileged/bootstrap SQL), these actions are `unknown`.
- Otherwise `user` = `app.ctx_user()` when set; else `unknown`. `unknown` covers privileged/bootstrap SQL with no
  context and anonymous authentication failures. No authenticated actor is invented from missing context.

| # | Action | Mechanism (producer) | Outcomes · reasons | Resource · scope · subject |
|---|---|---|---|---|
| T1 | organization.created | trigger `app.organizations` INSERT | success | organization · org |
| T2 | organization.renamed | trigger UPDATE OF name (changed) | success | organization · org |
| T3 | organization.policy_changed | trigger UPDATE OF flag (changed) | success · `group_admin_profile_edit_enabled\|disabled` | organization · org |
| T4 | group.created | trigger `app.groups` INSERT | success | group · org, group |
| T5 | group.renamed | trigger UPDATE OF name (changed) | success | group · org, group |
| T6 | group.slug_changed | trigger UPDATE OF slug (changed) | success | group · org, group |
| T7 | membership.granted | trigger `app.role_bindings` INSERT | success | role_binding · org, group? · subject, role |
| T8 | membership.role_changed | trigger UPDATE OF role (changed) | success | role_binding · org, group? · subject, role, previous_role |
| T9 | membership.removed | trigger DELETE | success | role_binding · org, group? · subject, role (old) |
| T10 | business_profile.created | trigger `app.business_profiles` INSERT | success | business_profile · org |
| T11 | business_profile.updated | trigger UPDATE OF 5 context columns (row changed) | success | business_profile · org |
| T12 | business_profile.linked | trigger `app.group_business_profiles` INSERT | success | business_profile · org, group |
| T13 | business_profile.unlinked | trigger DELETE | success | business_profile · org, group |
| T14 | user.created | trigger `app.users` INSERT | success | user · global · subject |
| T15 | user.disabled | trigger UPDATE OF status active→disabled (**revocation-priority path**) | success (actor system in the lifecycle producer) | user · global · subject |
| T16 | user.enabled | trigger UPDATE OF status disabled→active | success (actor system) | user · global · subject |
| T17 | session.issued | trigger `app.sessions` INSERT | success | session · global · subject |
| T18 | session.revoked | **statement-level** trigger, UPDATE revoked_at NULL→set (one row per revoked session, including expired-but-unrevoked sessions); `logout\|logout_all\|provider_logout\|disabled` on the revocation-priority path, `superseded` (sign-in rotation) strict | success · `logout\|logout_all\|provider_logout\|disabled\|superseded` | session · global · subject |
| T19 | session.provider_revocation_recorded | trigger UPDATE OF provider_revocation (changed) | success (`confirmed`) \| failed (`failed`): provider-**reported**, not verified provider state | session · global · subject |
| T20 | object.upload_started | trigger `app.stored_objects` INSERT (pending) | success | stored_object · org, group? |
| T21 | object.ready | trigger UPDATE OF state pending→ready | success | stored_object · org, group? |
| T22 | object.delete_started | trigger state pending\|ready→deleting (logical deletion) | success | stored_object · org, group? |
| T23 | object.deleted | trigger state deleting→deleted (physical absence confirmed, tombstone) | success | stored_object · org, group? |
| G1 | auth.sign_in | Go: success **atomic in `signIn`'s transaction**; denied/failed independent | success · `signed_in\|session_rotated`; denied · `transaction_expired\|login_canceled\|code_rejected\|impersonation_refused\|mfa_unverified\|stale_authentication\|identity_conflict\|user_disabled`; failed · `provider_error\|provider_timeout\|identity_namespace_mismatch\|identity_unavailable` | global · subject when verified |
| G2 | access.denied | Go independent, after the denied transaction | denied · `organization_not_permitted\|group_not_selectable\|target_not_found\|target_kind_mismatch\|target_outside_context\|role_not_permitted\|no_linked_path\|edit_policy_disabled\|not_creator\|object_not_ready\|capability_unknown\|fact_inconsistent`; `operation` ∈ the 15 operations of §3 | target only when **established**; org only when established, else global |
| G3 | object.url_issued | Go audit transaction after offline presign, **before** the URL is returned | success · `ttl_seconds` | stored_object · org, group? |
| G4 | object.upload_failed | Go independent, after a post-T1 failure and its compensation attempt | failed \| unknown · `compensated\|compensation_failed\|not_compensated` | stored_object · org, group? |
| G5 | object.deletion_pending | Go independent, when physical removal or the tombstone is not confirmed | failed \| unknown · `provider_unavailable\|tombstone_failed\|tombstone_unknown` | stored_object · org, group? |
| G6 | revocation.unaudited | Go independent (identity), **only** after a revocation transaction COMMITTED with its gap marker set (r1, owner decision) | success · `logout\|logout_all\|provider_logout\|disabled`: attests the committed revocation whose row-change event(s) could not be written | session (`logout`, `provider_logout`) or user (`logout_all`, `disabled`) · global · subject |

**Duplicates.** Each catalog event has exactly one owner. `auth.sign_in` (the authentication decision) and
`session.issued`/`session.revoked` (row transitions) are distinct catalog events, as the Spec requires; they are not a
duplicate pair. The database refuses a Go insert of any T-action and a trigger insert of any G-action.

`revocation.unaudited` is not a substitute copy of `session.revoked`/`user.disabled`. It exists only when those trigger
events are absent because of a caught audit-only fault, and it names that gap.

**Retry identity.** There are no idempotency keys and no automatic retries. A trigger event exists exactly when its row
change committed. A Go event is attempted once. A caller's retry is a new operation with new events.

**Event time.** `occurred_at` is the server clock at append. For independent records it is the recording time, which
follows the decision by at most the audit budget (§4.3). No separate persistence time is kept.

## 3. Producer inventory (R1, AC1)

Every current entry point at version 7. "Live" means the `identity` package, which the server binary links. `access`
and the `tenancy` primitives are linked into no binary (dormant until adapters exist).

| Entry point (call site) | Catalog events | Notes / no-op and refusal granularity |
|---|---|---|
| `identity.Service.Login` → `store.insertTransaction` | — (exempt) | anonymous protocol state (expired-row cleanup, login row); `auth.login.started` stays operational |
| `Callback` → `store.consumeTransaction` | — (exempt) | one-use protocol state |
| `Callback` → `store.signIn` | T14 (first sign-in), T18 superseded (rotation), T17, **G1 success** | binding upsert exempt (one-time namespace bootstrap); principal row is part of T14 (subject strings excluded); display-name refresh not audited; `already_signed_in` commits nothing → no event |
| `Callback` failures **after** the one-use transaction was consumed | G1 denied/failed | `user_disabled` attributed to the verified mapped subject, `identity_conflict` to the live-session subject; all others `unknown`. Pre-consume refusals (ambiguous parameter, unconfigured, origin, namespace, invalid transaction) stay operational: anonymous, no state change |
| `revokeExtraProviderSession` (conflict) | — | best-effort provider call; operational `auth.provider.revocation` only |
| `Session` → `store.resolve` renewal | — | heartbeat; trigger not fired (`UPDATE OF` excludes renewal columns) |
| `Logout` current/all → `revokeSession`/`revokeAll` | T18 `logout`/`logout_all`: one per revoked session, including expired-but-unrevoked rows for `logout_all`, with no count cap. On an audit-only fault: revocation commits, then G6 + `audit.gap` (§4.2) | idempotent current logout with no live session → no event; origin/unauthenticated/CSRF refusals operational (transport checks, no state change) |
| `providerLogout` → `revokeSession`, provider call, `recordProviderLogout` | T18 `provider_logout` (revocation-priority path; G6 on gap), T19 (strict: provider bookkeeping after the local revocation committed) | redirect URL never recorded; a T19 audit fault fails only the bookkeeping step (`identity_unavailable`, as for any failure of that step today) and never undoes the committed revocation |
| `ProviderLogoutReturn` → `consumeLogoutRef` | — | protocol bookkeeping (`provider_logout_returned_at`) |
| `DisableUser` | T15 (system), T18 `disabled` per revoked session, including expired-but-unrevoked rows (system), all on the revocation-priority path; G6 `disabled` + `audit.gap` on a gap | repeat on a disabled user changes no status → no event |
| `EnableUser` | T16 (system) | **strict**: an audit fault fails the enable (owner ruling) |
| `LockLiveSession`/`CheckLiveSession` | — | reads |
| tenancy `CreateOrganization`/`RenameOrganization` | T1/T2 | trusted-only; at v6+ only privileged pools can run structural primitives |
| tenancy `CreateGroup` | T4 + T7 (creator `group_owner`) | |
| tenancy `RenameGroup`/`ChangeGroupSlug` (raw) and `*Tx` writers | T5/T6 | raw: actor unknown; via facade: user |
| tenancy `Grant*Role`/`Change*Role`/`Remove*Membership` | T7/T8/T9 | runtime cannot run these at v6+ (G008); trigger records any privileged run |
| tenancy `CreateBusinessProfile`, `UpdateBusinessProfile` (raw and `Tx`) | T10/T11 | |
| tenancy `Link`/`UnlinkGroupBusinessProfile` | T12/T13 | |
| trusted fixture/bootstrap direct SQL | same triggers, actor `unknown` | separately identified; never evidence of user attribution |
| `access.BeginRequest` | — | authenticating an existing session is not a durable event (operational `access.decision`) |
| `access.Resolve` | denial → G2 (`access.resolve`) | success is per-request selection: no persisted context exists, so there is no "change" producer (G014 owns persisted Group selection and will produce through this catalog). Binding/scope misuse (B1–B6, nested scope) stays operational: an in-process programming fault |
| `GetOrganization`, `GetOrganizationPolicy`, `GetGroup`, `GetBusinessProfile` | denial → G2 | successful reads not audited |
| `SetOrganizationPolicy` | T3; denial → G2 | absolute set: a no-op changes no row → no event |
| `RenameGroup`, `ChangeGroupSlug` | T5/T6; denial → G2 | no-op → no event; `slug_conflict`/`invalid_input` → no event |
| `UpdateBusinessProfile` | T11; denial → G2 | identical full replace → no event |
| `ReadAuditEvents` (new) | denial → G2 (`audit.read`) | successful audit reads not audited |
| `PutObject` | T20 (T1), T21 (T2); T1 denial → G2 (`object.put`); post-T1 failure → G4 | reconciled-ready emits nothing extra (T21 committed with T2); a simulated crash seam records nothing |
| `DeleteObject` | T22 (D1), T23 (D4); denial → G2; unconfirmed removal → G5 | repeat on a tombstone changes no row → no event |
| `SignedGetURL` | G3; denial → G2 | issuance is not evidence of download |
| `GetObject`, `ObjectMetadata` | denial → G2 | streaming reads never write per chunk |

**Operations of G2** (closed): `access.resolve`, `organization.get`, `organization.get_policy`,
`organization.set_policy`, `group.get`, `group.rename`, `group.change_slug`, `business_profile.get`,
`business_profile.update`, `object.put`, `object.get`, `object.metadata`, `object.sign`, `object.delete`, `audit.read`.

**Recordable denial** = a policy refusal after a valid binding check: `context_denied` with
`organization_not_permitted`/`group_not_selectable`, or `not_found`/`forbidden`.
- **Not recorded:** unauthenticated (session invalid), binding misuse, invalid input, conflict, and
  unavailable/internal failures.
- **"Established" Organization:** the denied transaction's facts passed the Organization gate (`organizationRole`
  returned no reason). This is carried on the failure and re-validated by the database (§5.3).
- **"Established" target:** only `forbidden` (readable on the same facts), so a rejected foreign selector never
  becomes a recorded UUID. Unknown and foreign targets record identically (no new timing or enumeration channel
  between them).

**Future (no current producer):** persisted context select/clear (G014); membership/onboarding administration
(G018/G019; its row changes are captured automatically); Organization/Group/profile deletion (no path); export/UI;
retention deletion (G013).

## 4. Transaction, failure and attribution policy (R3, R4, R5)

### 4.1 Mechanisms

| Mechanism | Persistence point | Audit-append failure | Unknown COMMIT |
|---|---|---|---|
| **M1 trigger** (T1–T23 except the M5 transitions) | same statement/transaction as the row change (AFTER ROW, or AFTER STATEMENT for revocations, before COMMIT) | the statement fails, so the whole transaction rolls back: no unaudited success. Facade `internal` (or `unavailable` for store/cancel), identity `identity_unavailable`; nothing changed | unchanged existing semantics: change and event both durable or neither; no retry, no Go re-append |
| **M2 Go in-transaction** (G1 success) | `signIn`'s transaction through `Store.inTxAudited` (r1): after `fn` (mapping/rotation/issue), before COMMIT | sign-in fails `identity_unavailable`; no session, no events | unchanged identity semantics (`identity_unavailable`, may be durable); no retry |
| **M3 Go independent** (G1 denied/failed, G2, G4, G5) | a new short audit transaction **after** the business transaction ended and released its connection | the caller's documented outcome is returned unchanged; one sanitized `audit.record_failed` log record; no durable row is claimed | no retry; the event may or may not exist; the outcome is unchanged |
| **M5 revocation priority** (T15 `active→disabled`; T18 `logout\|logout_all\|provider_logout\|disabled`; then G6) | same transaction as the revocation, inside a bounded exception block; G6 after COMMIT | only the audit insert rolls back; the revocation commits; the marker is set; after COMMIT a G6 record is attempted (identity recorder, §4.3) and one `audit.gap` diagnostic (`independent: recorded\|failed`) is emitted; the call returns its normal success. Never claimed: the missing T-event | revocation database failure: `identity_unavailable` as today. COMMIT failed or unknown with the marker set: `identity_unavailable`, no G6 attempt, `audit.gap` with `independent: not_attempted`. No retry |
| **M4 issuance** (G3) | read-guard transaction (unchanged), then offline presign outside any transaction, then the audit transaction | **no URL** is returned (`unavailable`, `storage.operation` stage `audit`) | no URL is returned; at most one `object.url_issued` may be durable ("issued, not delivered"); no retry |

**Invariants:**
- No provider or network I/O inside any audit-holding transaction. G009's staged I/O is unchanged; presign stays
  outside; identity provider calls stay between transactions.
- No request holds two pooled connections.
- No audit statement takes row locks on `app` tables: no foreign keys; the insert takes only `RowExclusiveLock` on
  `audit.events`. Lock order is unchanged.
- No OLD/NEW row dumps: capture functions copy named columns only, and CT-AUD-4 asserts that their bodies contain no
  `row_to_json`, `to_jsonb`, `hstore` or `NEW.*`/`OLD.*` expansion.
- An effect is never retried to obtain a record. When an effect succeeded but its later independent record fails (for
  example, bytes removed but the D4 tombstone and G5 both failed), the earlier trigger events (`object.delete_started`)
  remain the durable record and the diagnostic names the missing action.

### 4.2 Per-operation pins

| Operation | Ordering / return | Revocation and failure truth |
|---|---|---|
| Guarded writes (policy, rename, slug, profile) | G007 steps 1–9 unchanged; the event is written inside step 9's statement; COMMIT | denied → G2 after rollback; failed write → no event |
| Logout / logout-all / provider logout / DisableUser | identity transactions as today (DisableUser: users `FOR UPDATE` → status → sessions); then **one added statement before COMMIT** reads the gap marker | **revocation priority (owner decision, replaces r0 E7):** an audit-only fault (any error raised by the audit insert, its BEFORE trigger or its RLS check) cannot prevent or undo the local revocation. M5 commits it, then attempts G6 and reports the gap with `audit.gap`. A failure of the revocation's own database work, or an unknown COMMIT, is still `identity_unavailable`, never success. `query_canceled` is not caught, so a canceled revocation fails truthfully |
| EnableUser, sign-in (incl. `superseded` rotation), grants and all other mutations | unchanged | **strictly atomic** (M1/M2): an audit fault fails the operation and nothing commits |
| Sign-in | exchange (outside transaction) → `signIn` transaction: mapping/rotation/issue (T14/T18/T17) → G1 success → COMMIT | failure after consume → G1 denied/failed (M3) |
| Put | T1 (T20) → provider → V → T2 (T21) → R as G009 | post-T1 failure: compensate (result **captured**: every branch except size mismatch currently discards it) → G4 `failed` + compensation result. T2 unknown and R fails → G4 `unknown`/`not_compensated` (row may be ready). If the post-T1 failure is a T2 policy denial, G4 is recorded **without** Organization (scope no longer established) |
| Delete | D1 (T22) → provider delete + Head → D4 (T23) | unconfirmed removal → `deletion_pending` (unchanged) + G5 (`provider_unavailable`; D4 definitive `tombstone_failed`; D4 unknown `tombstone_unknown`, where T23 may also exist: both are truthful) |
| Sign | read guard → presign (`providerControlTimeout`) → audit transaction (G3) → return the URL only after COMMIT succeeds | URL, signature and key are never stored |

### 4.3 Independent recording budget, cancellation and disposal

- `auditTx` runs on `context.WithTimeout(context.WithoutCancel(ctx), auditRecordBudget = 2s)` for BEGIN..COMMIT, with
  `auditCleanupBudget = 2s` for the deferred rollback. These are the G009 storage-budget mechanics (`inTx` cleanup
  parameter), with fixed constants.
- Request cancellation therefore does not suppress recording a decision that was already made, and recording is
  bounded.
- Poisoned connections are inherited from `inContextTx`: prior context → discard → recording fails → diagnostic.
- **The identity recorder is bounded (r1, item 2).** `(*identity.Store).recordAuditEvents` does not use `Store.inTx`:
  - it opens its own transaction (`s.begin`) on `context.WithTimeout(context.WithoutCancel(ctx), auditRecordBudget)`;
  - its deferred rollback runs on a fresh `context.WithTimeout(context.WithoutCancel(ctx), auditCleanupBudget)`;
  - an expired cleanup budget makes pgx close the connection while it is held, so pgxpool destroys it on release (the
    G009 mechanism);
  - a prior context found by its initializer runs the existing `discard` (close while held), returns
    `identity_unavailable` and appends nothing;
  - caller cancellation does not cancel it; only its own budgets bound it;
  - there is no retry. Every other identity transaction keeps `Store.inTx` and its existing unbounded-rollback
    convention, unchanged.
- Recording happens **before** the call's operational record is written, so the existing "last record is
  access.decision" test convention and the one-record-per-call contract hold.
- If the database is unavailable, no durable failure evidence can exist. The diagnostic is the only trace (stated
  limit).

### 4.4 Trusted attribution and trust boundary

- **Context statement** (`sqlInitContext`, byte-identical in `tenancy` and `identity`) adds
  `set_config('audit.request_id', $5, true)` and `set_config('audit.trace_id', $6, true)` from
  `observability.RequestID/TraceID(ctx)`.
  - It also resets `set_config('audit.revocation_gap', '', true)`, a constant, so a stale session-level value can never
    be read as a gap (r1).
  - They are **not** part of the prior-context guard: they confer no authority, and identity's EMPTY→LOGIN and
    savepoint double initializations must keep working.
  - The four `app.*` fields, their guard, the parsers and the RLS shapes are unchanged.
  - Correlation is never compared or used in a decision; AUTHORIZATION.md's "never stored" wording becomes "stored only
    as bounded diagnostic audit correlation".
- **Trigger events:** Organization and Group come from the changed row (a database fact). For runtime writes, RLS
  already forces the row's Organization to equal `ctx_org()`; the INSERT policy re-checks it. The actor comes from
  context per §2's rule.
- **Go events (depth 1):** the actor is always `ctx_user()`.
  - An Organization is accepted only when it equals `ctx_org()` **and** `app.rls_org_role()` is non-NULL now.
  - **A Group is accepted only when it is proven to belong to that Organization (r1, item 1):**
    `EXISTS (SELECT 1 FROM app.groups WHERE id = NEW.group_id AND organization_id = NEW.organization_id)`.
    - The check runs as the invoker under the recording transaction's TENANT-ORG RLS (`groups_select`).
    - An Organization owner/admin sees every Group of the active Organization (the `MGR` branch), so a legitimate
      owner/admin needs no individual Group membership.
    - A Group-role actor sees only its own Groups.
    - A foreign Group (another Organization), a nonexistent UUID or an invisible Group yields no row, and the insert
      is refused. With no foreign key, this is the only way a direct append may name a Group.
    - Trigger rows (depth ≥ 2) take their Group from the changed row, a database fact, and are not re-read.
  - Otherwise the insert is refused, never re-attributed.
  - Org-scoped records use the TENANT-ORG shape (no selected Group), so Group validation uses the actor's own role.
    Org-less records use the IDENTITY shape (or EMPTY for anonymous authentication failures).
- **Trusted-only primitives and bootstrap** (no request): recorded `actor_kind = unknown`, with the row's
  Organization/Group. They can never claim a user, because actor fields are derived and never insertable. A privileged
  writer that sets an application context is attributed to that context; privileged administration is outside the
  claim.
- **Not claimed:** that a superuser, the owner, or a holder of the runtime credential with arbitrary SQL cannot fabricate
  or suppress events. That holder can set any context and insert Go-owned actions for it, as G008's trust limit already
  states for every table. G008 D1 residual: a raw-runtime self re-enable would be recorded as `system`, and a
  self un-revocation (`revoked_at` set→NULL) is not an existing code path and is not captured. A privileged session
  that revokes with an audit fault commits with no reader of the gap marker (privileged administration is outside the
  claim).
- **The database guarantees:**
  - no runtime UPDATE/DELETE/TRUNCATE/DDL/GRANT;
  - no runtime-supplied actor, time, id or correlation;
  - trigger-owned events only from real row changes;
  - closed shapes;
  - Organization attribution consistent with the established context;
  - a direct-append Group that belongs to the event Organization;
  - the revocation exception limited to the five pinned transitions, owner-owned trigger code, with no runtime
    trigger-disable or bypass privilege.

## 5. Database: provisioning and migration `00008_audit_events.sql`

### 5.1 Provisioning (escalation E1)

- `infra/postgres/init/10-roles.sh` (registered customization): `CREATE SCHEMA audit AUTHORIZATION :"owner_role";
  REVOKE ALL ON SCHEMA audit FROM PUBLIC;`.
- `infra/postgres/provision-roles.sql`: add a read-only refusal "schema audit exists but is not owned by <db>_owner",
  then `CREATE SCHEMA IF NOT EXISTS audit AUTHORIZATION owner` and, as the owner, `REVOKE ALL ON SCHEMA audit FROM
  PUBLIC`. A repeat run changes nothing (PROV-4). The CREATEROLE-admin path already holds a temporary SET on the owner.
- No new role, membership, password or environment variable.
- `rls_prov_catalog` adds `audit` to its schema set; PROV-1 then proves 10-roles.sh ≡ provision-roles.sql for the new
  schema, and PROV-6's pinned delta gains `schema:audit:DB_owner:-`.
- **Deploy order:** provision (existing volumes re-run provision-roles.sql, the G008 precedent) → migrate to 8 → deploy
  code. Old code against v8 works (triggers record; correlation NULL). New code against v7 fails closed at sign-in
  (append refused).

### 5.2 Relation (DDL)

```sql
CREATE TABLE audit.events (
  id              uuid        NOT NULL DEFAULT gen_random_uuid(),
  schema_version  smallint    NOT NULL DEFAULT 1 CONSTRAINT events_schema_version CHECK (schema_version = 1),
  occurred_at     timestamptz NOT NULL DEFAULT clock_timestamp(),
  action          text        NOT NULL CONSTRAINT events_action_valid CHECK (action IN (/* the 29 actions of §2 */)),
  outcome         text        NOT NULL CONSTRAINT events_outcome_valid CHECK (outcome IN ('success','denied','failed','unknown')),
  reason          text        CONSTRAINT events_reason_valid CHECK (reason IS NULL OR reason IN (/* §2 reasons */)),
  operation       text        CONSTRAINT events_operation_valid CHECK (operation IS NULL OR operation IN (/* §3, 15 */)),
  actor_kind      text        NOT NULL CONSTRAINT events_actor_kind_valid CHECK (actor_kind IN ('user','system','unknown')),
  actor_user_id   uuid,
  organization_id uuid,
  group_id        uuid,
  resource_type   text        CONSTRAINT events_resource_type_valid CHECK (resource_type IS NULL OR resource_type IN
                    ('organization','group','role_binding','business_profile','user','session','stored_object')),
  resource_id     uuid,
  subject_user_id uuid,
  role            text        CONSTRAINT events_role_valid CHECK (role IS NULL OR role IN (/* 6 role codes */)),
  previous_role   text        CONSTRAINT events_previous_role_valid CHECK (previous_role IS NULL OR previous_role IN (/* 6 */)),
  ttl_seconds     integer     CONSTRAINT events_ttl_valid CHECK (ttl_seconds IS NULL OR ttl_seconds BETWEEN 1 AND 300),
  request_id      text        CONSTRAINT events_request_id_valid CHECK (request_id IS NULL OR request_id ~ '^[A-Za-z0-9._-]{8,64}$'),
  trace_id        text        CONSTRAINT events_trace_id_valid CHECK (trace_id IS NULL OR trace_id ~ '^[0-9a-f]{32}$'),
  CONSTRAINT events_pkey PRIMARY KEY (id),
  CONSTRAINT events_actor_consistent CHECK ((actor_kind = 'user') = (actor_user_id IS NOT NULL)),
  CONSTRAINT events_group_needs_org CHECK (group_id IS NULL OR organization_id IS NOT NULL),
  CONSTRAINT events_resource_pair CHECK ((resource_type IS NULL) = (resource_id IS NULL)),
  CONSTRAINT events_ids_nonzero CHECK (/* no zero UUID in id, actor, org, group, resource, subject */)
);
CREATE INDEX events_tenant_time_idx ON audit.events (organization_id, occurred_at, id) WHERE organization_id IS NOT NULL;
```

No foreign keys (R7): deleting a source row never deletes history, and attribution stays historical. Three relations.

### 5.3 Functions and triggers (all owner-owned, SECURITY INVOKER, `search_path = pg_catalog, pg_temp`, PUBLIC EXECUTE revoked, no grants)

**`audit.events_attribute()`**, `BEFORE INSERT … FOR EACH ROW`:

```
d := pg_trigger_depth()            -- 1: direct INSERT (Go); >= 2: issued by a capture trigger
u := app.ctx_user(); o := app.ctx_org()   -- parsers: malformed/illegal context raises (G008)
NEW.id := gen_random_uuid(); NEW.schema_version := 1; NEW.occurred_at := clock_timestamp()
NEW.request_id / NEW.trace_id := audit.* settings when they match the column patterns, else NULL
IF (d >= 2) <> (NEW.action is a T-action) THEN RAISE 42501 (audit_event_source)
(actor_kind, actor_user_id) := §2 actor rule
IF d = 1 THEN r := app.rls_org_role();
   org given and (org <> o or r IS NULL)                                → RAISE 42501 (audit_event_scope)
   group given and not EXISTS (app.groups row: id = group AND organization_id = org, under RLS) → RAISE 42501 (r1)
per-action shape (outcome, reason set, operation, resource type, role fields, ttl, scope presence) → else RAISE 23514 (audit_event_shape)
RETURN NEW
```

Rows issued by capture triggers make no reader calls, so the expected business-statement reader counts are unchanged;
only descriptive parser counts grow. A direct append makes one `rls_org_role` call (organization), plus, with a Group,
one PK read of `app.groups` whose `groups_select` policy adds one `rls_org_role` init-plan and at most one
`rls_group_role` call (§11).

**`audit.events_refuse()`**: `BEFORE UPDATE OR DELETE … FOR EACH ROW` and `BEFORE TRUNCATE … FOR EACH STATEMENT`. It
raises `42501 'audit events are append-only'` for every role, superuser included, unless a privileged administrator
disables the trigger.

**Eight row-level capture functions, `audit.capture_<table>()`**, plus `audit.capture_session_revocations()`, all
`RETURN NULL`. Each inserts the §2 row with outcome `success`,
except T19's provider outcome:

| Table | Triggers (`audit_capture_insert` / `_update` / `_delete`) | Fields copied |
|---|---|---|
| organizations | INSERT; UPDATE OF name, group_admins_edit_business_profiles WHEN changed | id; flag→T3 reason |
| groups | INSERT; UPDATE OF name, slug WHEN changed (one event per changed attribute) | id, organization_id |
| role_bindings | INSERT; UPDATE OF role WHEN changed; DELETE | id, organization_id, group_id, subject_user_id, role (and OLD.role) |
| business_profiles | INSERT; UPDATE OF the 5 context columns WHEN `ROW(old) IS DISTINCT FROM ROW(new)` | id, organization_id |
| group_business_profiles | INSERT; DELETE | business_profile_id, organization_id, group_id |
| users | INSERT; UPDATE OF status WHEN changed. `active→disabled` is the **M5 branch**: `BEGIN INSERT … EXCEPTION WHEN OTHERS THEN` add `user.disabled` to `audit.revocation_gap` `END`. `disabled→active` stays strict | id |
| sessions | row INSERT (T17); row `UPDATE OF provider_revocation` WHEN changed (T19, strict); **statement-level** `AFTER UPDATE REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows` → `audit.capture_session_revocations()` (T18) | id, user_id, revoke_reason, provider_revocation |
| stored_objects | INSERT; UPDATE OF state WHEN changed (after `stored_objects_guard`) | id, organization_id, group_id, state |

**`audit.capture_session_revocations()`** (r1, M5) runs once per sessions UPDATE statement:

```
T  := new_rows JOIN old_rows USING (id) WHERE old.revoked_at IS NULL AND new.revoked_at IS NOT NULL
SELECT bool_or(reason = 'superseded'), bool_or(reason IN ('logout','logout_all','provider_logout','disabled'))
  INTO has_strict, has_priority FROM T                -- transition tables only; never touches audit.events (r2)
IF has_strict THEN
  INSERT the 'superseded' rows of T (strict, no exception block)                   -- sign-in rotation stays atomic
END IF;
IF has_priority THEN
  BEGIN INSERT those rows (one INSERT … SELECT, N rows)                             -- one subtransaction per statement
  EXCEPTION WHEN OTHERS THEN add 'session.revoked' to audit.revocation_gap (set_config(…, true))
  END;
END IF;
```

- PostgreSQL forbids a column list on an UPDATE trigger with transition tables, so this statement trigger fires on
  every sessions UPDATE, including renewal. A renewal pays one function call and one aggregate query over its 1-row
  transition tables. Both flags are false, so it executes **no** audit statement and enters no exception block.
- **r2:** an INSERT that would return zero rows still checks INSERT privilege and locks `audit.events`, so no audit
  statement runs unless its flag is set.
  - A logout/disable statement (only priority reasons) therefore reaches `audit.events` only inside the M5 block.
  - A sign-in rotation statement (`superseded`) still runs its strict INSERT, so an audit fault there fails the sign-in
    atomically.
  - The `users` M5 branch already executes only the branch-specific INSERT: its strict enable INSERT is in the
    `disabled→active` branch.
- Uncapped revocation fan-out costs one subtransaction per statement, never one per row, so no subtransaction
  overflow (§11).
- Errors other than the audit insert are not inside the block. `WHEN OTHERS` does not catch `query_canceled`.

That is 18 triggers on 8 app tables (17 row, 1 statement) plus 3 on `audit.events`; 11 functions in `audit`. The capture functions'
`INSERT INTO audit.events` runs as the invoking role, so runtime writes need the column grant below.
`stored_objects_guard` and the `*_keys_immutable` triggers are unchanged.

### 5.4 RLS, grants and schema privileges

- `ENABLE` + `FORCE ROW LEVEL SECURITY` on `audit.events`. Two permissive policies `TO <db>_runtime, <db>_owner`:
  - `events_select FOR SELECT USING (organization_id = (SELECT app.ctx_org()) AND (SELECT app.rls_org_role()) = 'org_owner')`;
  - `events_insert FOR INSERT WITH CHECK ((actor_kind <> 'user' OR actor_user_id = (SELECT app.ctx_user())) AND
    (organization_id IS NULL OR organization_id = (SELECT app.ctx_org())))`.
  - No UPDATE/DELETE policy, so the owner path affects zero rows. Retention and the helpers have no policy.
- Grants:
  - `USAGE ON SCHEMA audit TO <db>_runtime`. Retention gets nothing: the Q8 ruling defers its privilege until G013 needs
    it;
  - `SELECT ON audit.events TO <db>_runtime`;
  - `INSERT (action, outcome, reason, operation, resource_type, resource_id, subject_user_id, organization_id, group_id,
    role, previous_role, ttl_seconds) ON audit.events TO <db>_runtime`;
  - nothing else: no UPDATE/DELETE/TRUNCATE/REFERENCES/TRIGGER, no EXECUTE in `audit`, no PUBLIC entry, nothing to the
    helpers or retention. The runtime role already lacks TEMP/CREATE/TRIGGER and every role membership.
- **No `app` grant, policy, function or role changes.** Runtime and owner already hold EXECUTE on the parsers and
  readers. "SECURITY DEFINER appears only in 00006" stays true.
- **Global rows** (`organization_id IS NULL`) match no runtime/owner policy. Inspection is a privileged operator action
  (superuser, or the owner after an explicit `NO FORCE` maintenance step), documented in AUDIT.md. There is no product
  or platform-admin role (Q4).
- **Maintenance boundary (G013 handoff):** in G011 `<db>_retention` holds **no** `audit` privilege (Q8). G013's
  reviewed migration may add, all at once:
  - `USAGE`;
  - an expired-row DELETE grant and policy `TO <db>_retention`;
  - an `events_refuse` exception limited to that role and expired rows.

  Runtime and owner stay refused. No retention, scheduler, backup, legal hold or inactive grant here.

### 5.5 Up (one transaction): preconditions → objects → postconditions

**Preconditions** (fail closed):
- the current user owns `app`;
- the 00007 role checks, reused;
- schema `audit` exists, is owned by the owner, holds no objects and no non-owner ACL entry (else
  `run infra/postgres/provision-roles.sql`);
- `app.stored_objects` exists;
- the exact version-7 strings of 00007 postconditions 1–4 (RLS tables, policy inventory, ACLs, function allowlist),
  copied verbatim.

**Postconditions** (exact strings, symbolic role labels):
- the app inventories are unchanged;
- the `app` trigger inventory is the 18 new triggers (17 row, 1 statement with transition tables) plus the existing
  ones;
- `audit` relations (3), columns, constraints, RLS flags, policies (2), table and column ACLs, function allowlist (11)
  and its triggers (3);
- schema privileges: `audit` = `rt=USAGE`; `app` unchanged.

### 5.6 Down (development and test only; the runner refuses it in production)

1. `ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY`, so the owner's count is unfiltered.
2. **Refuse** while any row exists. A populated history is never silently discarded; the refusal rolls back.
3. Drop the 18 capture triggers, the table and the 11 functions; revoke the schema USAGE grant.
4. Postconditions: the exact version-7 inventories and an empty, owner-only `audit` schema (the provisioning object
   stays).

Disposable test databases that must step below 8 purge rows first through the attested superuser
(`DISABLE TRIGGER`/`TRUNCATE`/`ENABLE`, §10.2). That is a test-only privileged action, never production.

### 5.7 Migration risk

- Catalog-only, with no rewrite or backfill: history starts at version 8.
- `CREATE TRIGGER` takes brief `SHARE ROW EXCLUSIVE` locks on the eight tables; migrate in a quiescent window.
- 00001–00007 stay byte-identical (00007 digest newly pinned).
- **Irreversibility:** once rows exist, rolling back requires a deliberate privileged purge (escalation E10).

## 6. Go changes and the audit-aware helper (R4 amendment)

- **`core/internal/audit`** (new, imports pgx only; importable by `identity` and `access`):
  - closed Go constants for the 29 actions, outcomes, reasons, operations and resource types, with unit parity against
    the migration's lists;
  - `type Event` (Go-owned fields only);
  - `func Append(ctx context.Context, tx pgx.Tx, e Event) error`: one `INSERT INTO audit.events (…)` on the caller's
    transaction, never its own connection. It refuses a T-action in Go and returns `ErrRefused` (42501/23xxx) or
    `ErrUnavailable`, never driver text;
  - `type Record` and `func ReadTenant(ctx, tx pgx.Tx, org [16]byte, since, until time.Time, afterAt time.Time,
    afterID [16]byte, limit int) ([]Record, error)`: one statement
    `WHERE organization_id = $1 AND occurred_at >= $2 AND (occurred_at, id) < ($3, $4) ORDER BY occurred_at DESC,
    id DESC LIMIT $5`. Sentinels: since `-infinity`; no cursor = (until or `infinity`, zero UUID). This keeps the
    statement sargable for a backward scan of `events_tenant_time_idx`;
  - input codes `invalid_page_size`, `invalid_time_range`, `invalid_cursor` (used as `access.Error.Field`).
- **The audit-aware helpers (r1, item 3).** Each is the smallest extension of an existing transaction helper. There is
  no unit-of-work engine, no callback retry and no refactor of unrelated call sites.
  - **Owning-transaction helper (mutation plus required append, R4 amendment):** `(*identity.Store).inTxAudited(ctx
    context.Context, fn func(tx pgx.Tx) ([]audit.Event, error)) error`.
    - It is `Store.inTx`: same begin, deferred rollback and commit, so the existing identity semantics and unknown-COMMIT
      truth are unchanged.
    - `fn` performs the mutation and returns the events it requires. The helper appends them through `audit.Append` on
      the **same** transaction after `fn` succeeds and before COMMIT. An append failure is a failed transaction: nothing
      commits, `identity_unavailable`.
    - `fn` may return the sentinel `errNoCommit` to end without committing; the helper rolls back and returns it.
      `signIn` uses this for its existing rollback-only outcomes (`already_signed_in`, `identity_conflict`), which today
      return before `commit`.
    - `signIn` becomes one `inTxAudited` call. Both success branches (`signed_in`, `session_rotated`) return their G1
      success event from `fn`, so the append exists in exactly one place. Statement order, initializers
      (`initIdentity,initIdentity`) and the binding upsert are unchanged.
  - **Independent recorders (separate, bounded):**
    - `(*access.Service).auditTx(ctx context.Context, dbc dbContext, events ...audit.Event) *failure` in `tx.go` calls
      `inContextTx(write, dbc, auditCleanupBudget, fn)` under `auditRecordBudget`; `fn` appends each event through
      `audit.Append`. It inherits `inTx`'s COMMIT classification, discard and bounded cleanup. A zero Organization
      selects the IDENTITY shape.
    - `(*identity.Store).recordAuditEvents(ctx context.Context, subject string, events ...audit.Event) error` has its own
      `begin`, a **bounded** deferred rollback and disposal (§4.3), `initIdentity(subject)` or `initEmpty()`, then
      `audit.Append` and `commit`. It serves G1 denied/failed and G6.
  - Where a trigger owns the event, the transaction whose first statement now carries the trusted correlation is the
    whole mechanism: no Go append, so no duplicate.
- **`tenancy`:** `sqlInitContext` gains `$5/$6` and the constant gap-marker reset. `InitContext` passes correlation from
  ctx and accepts a zero Organization **only with** a zero Group, producing the IDENTITY shape for org-less audit records
  (escalation E5). A zero user is still `invalid_id`.
- **`identity`:**
  - `sqlInitContext` stays byte-identical, and `initContext` passes correlation;
  - `signIn` runs through `inTxAudited` (above) and returns the mapped subject with `ErrUserDisabled`;
  - `Callback` sets `consumed` after `consumeTransaction`, and on a post-consume error records G1 through
    `recordAuditEvents` (bounded, detached);
  - **revocation priority (r1, item 4):**
    - `revokeSession`, `revokeAll` and `DisableUser` keep `Store.inTx`. Their `fn` ends with one statement,
      `SELECT coalesce(pg_catalog.current_setting('audit.revocation_gap', true), '')` (closed values
      `session.revoked`, `user.disabled`), captured before COMMIT.
    - After a **successful** COMMIT with a non-empty marker, identity calls `recordAuditEvents` with one G6 (`reason` =
      the operation; resource = the session for `logout`/`provider_logout`, the user for `logout_all`/`disabled`) and
      logs `audit.gap` (`operation`, `independent: recorded|failed`). It still returns the normal success.
    - On a COMMIT failure with the marker set it returns `identity_unavailable` (as today) and logs `audit.gap`
      (`independent: not_attempted`).
    - EnableUser, `signIn` and `recordProviderLogout` never read or tolerate the marker; their events are strict;
  - a recording failure outside M5 → `audit.record_failed` (error level).
- **`access`:**
  - `failure` gains unexported scope fields (Organization established; established target/Group);
  - `guarded`, `resolve` and the storage readers set them from the facts;
  - `guard`, `Resolve` and each storage wrapper call `recordDenial` **before** writing their operational record;
  - `objects.go` captures the compensation results, records G4/G5 and runs `auditTx` for G3 between presign and return;
  - `log.go` adds `audit.read` to `access.decision`, `stage: audit` to `storage.operation`, and the
    `audit.record_failed` writer.
- **Guarded read** (internal Go only): `func (s *Service) ReadAuditEvents(ctx context.Context, tc *TenantContext, req
  AuditReadRequest) (AuditPage, error)`.
  - It runs through the existing `guard` (`audit.read`, target the active Organization, REPEATABLE READ READ ONLY,
    `organizationQuery`).
  - `pre` validates the request purely: Limit 0→50, 1..100, else `invalid_input`/`invalid_page_size`; `Until > Since`
    when both are set; the cursor's Organization and window must equal the request (`invalid_cursor`).
  - `act` calls `audit.ReadTenant` in the guard's snapshot, with limit+1 to compute `More`.
  - Types: `AuditReadRequest{Since, Until time.Time; Limit int; Cursor AuditCursor}`; opaque `AuditCursor` (unexported
    Organization, window and last key; zero means first page); `AuditPage{Events []AuditEvent; Next AuditCursor; More
    bool}`; `type AuditEvent = audit.Record`.
  - A selected Group grants nothing new: an owner in selected-Group context reads the whole Organization; Group roles
    read nothing.
  - Filter shapes: the time window only, newest first. Actor and resource filters are deferred, because each would need
    its own index decision under PG-004. There is no export, count or cross-tenant listing.
  - A page is a snapshot. An event committed later with an earlier timestamp may be absent from an ongoing traversal
    and appears in a fresh one.

**Enforcement by layer (R6):**

| Layer | Enforces |
|---|---|
| Guarded read | liveness (ended session, disabled user) and current role |
| RLS | current role, downgrade, membership removal, wrong/missing context and pool reuse, also under raw runtime SQL |
| Not added | a user-status predicate in RLS, keeping the G008 convention |

## 7. Files and modules

| Path | Change |
|---|---|
| `core/migrations/00008_audit_events.sql` | **new** (§5) |
| `core/internal/audit/{audit,store}.go`, `audit_test.go` | **new** primitive, reader, catalog; unit U-AUDIT |
| `core/internal/tenancy/access/audit.go` | **new**: `recordDenial`, storage outcome records, `ReadAuditEvents`, cursor/types |
| `core/internal/tenancy/access/{tx,facade,context,errors,objects,log}.go` | `auditTx`; failure scope; recording calls; compensation capture; issuance transaction; log additions |
| `core/internal/tenancy/{store,facts}.go` | context statement `$5/$6`; `InitContext` correlation and zero-Organization IDENTITY shape |
| `core/internal/identity/{store,service}.go` | correlation; `inTxAudited` and `signIn` through it; bounded `recordAuditEvents`; post-consume recording; revocation gap read, G6 and `audit.gap` |
| `core/internal/identity/identity_test.go` | unit U-IDN-AUDIT (r1): recorder budgets, blocked rollback, poison, gap decision (§9) |
| `core/internal/migrate/audit_integration_test.go` | **new**: TestAuditCatalog, TestAuditMigration, TestAuditPersistSeed/Verify |
| `core/internal/tenancy/access/audit_integration_test.go` | **new**: TestAudit |
| `core/migrations/migrations_test.go` | set through 00008, 00007 pinned; TestAuditMigrationShape (static) |
| RC test edits (§10.2) | `migrate/{integration,objects_integration,rls_integration,runner}_test.go`, `tenancy/integration_test.go`, `access/{integration,boundary}_test.go`; conditional: grep-inventory hits |
| `infra/postgres/{init/10-roles.sh,provision-roles.sql}` | audit schema (§5.1) |
| `infra/validation/run.sh` | B-AUDIT-* functions; version-8 pins; `MIGRATION_00007_SHA`; `rls_prov_catalog` + PROV-6 delta; `APPLICABLE` += `core/internal/audit docs/foundation/AUDIT.md` |
| `infra/validation/{required-checks.tsv,render-customizations.tsv}` | +3 IDs and totals; the 10-roles.sh reason gains G011 |
| `infra/validation/tools/internal/logrecord/{events.json,logrecord_test.go}` | `audit.read`, stage `audit`, `audit.record_failed`, `audit.gap` (`operation` ∈ logout\|logout_all\|provider_logout\|disabled, `independent` ∈ recorded\|failed\|not_attempted) |
| `infra/ci/{validate-controls.sh,profile-paths.tsv}` | VA7/VU1/J8 count pins 69→72, 98→101; `foundation dir core/internal/audit/` |
| `docs/foundation/AUDIT.md` | **new** guide: catalog, schema/grants, trust limits, failure semantics, read API, operator inspection, G013 contract |
| `docs/foundation/{DATABASE-ISOLATION,AUTHORIZATION,OBJECT-STORAGE,IDENTITY,MIGRATIONS,TENANCY,FOLLOW-UP,OPERATIONS}.md` | version 8 and behavior deltas only; OPERATIONS totals 97→100 / 69→72 / 98→101 |

**Unchanged:** `contracts/`, `.gitlab-ci.yml`, compose, `.env.example`, `go.mod`/`go.sum`, migrations 00001–00007,
`core/cmd/*` (server `main.go` digest pin), GETTING-STARTED/VENDORS (rendered), the parsers/readers and every policy in
`app`.

## 8. Contracts and interfaces

- No public HTTP/BFF/MCP contract.
- New internal Go API: `ReadAuditEvents` and its types, and `audit.Append`/`ReadTenant`.
- New operational log events `audit.record_failed` and `audit.gap`.
- Live behavior change (server via `identity`, E13):
  - sign-in, logout and disable now write audit rows;
  - sign-in, enable and every other mutation fail closed if they cannot;
  - local logout (current/all/provider) and disable complete despite an audit-only fault, then attempt G6 and report
    `audit.gap` (owner decision).

## 9. Tests (minimal, reusing existing fixtures, real runtime logins, fake WorkOS and private MinIO)

**TestAuditCatalog (B-AUDIT-CATALOG, 7 subtests, superuser setup / attested runtime and retention assertions):**
- CT-AUD-1: relation, columns, constraints, index.
- CT-AUD-2: RLS flags, policy inventory and expressions.
- CT-AUD-3: every ACL (schema, table, column, function) for every role and PUBLIC; `audit` schema = runtime `USAGE`
  only (retention none, Q8); app ACLs unchanged.
- CT-AUD-4: 18+3 triggers (exact definitions incl. the one statement-level transition-table trigger, enabled),
  function allowlist (11), no row-dump constructs. The only `EXCEPTION` blocks are the two M5 branches, and their
  bodies hold only the audit insert.
- CT-AUD-5: runtime and retention refusals of UPDATE/DELETE/TRUNCATE/ALTER/DROP/CREATE in audit/GRANT/SET ROLE/DISABLE
  TRIGGER/`row_security=off`, each with rows and catalog unchanged.
- CT-AUD-6: owner FORCE (owner-context rows only; UPDATE/DELETE affect 0 rows; TRUNCATE refused).
- CT-AUD-NEG: N1–N8, each yielding exactly its detector code and rolled back: NO FORCE, UPDATE grant, actor-column
  INSERT grant, dropped refuse trigger, disabled capture trigger, permissive probe policy, PUBLIC schema USAGE, definer
  capture function.

**TestAuditMigration (B-AUDIT-MIGRATION, 6 subtests), plus TestAuditPersistSeed/Verify around a PostgreSQL
stop/start:**
- MG-AUD-1: empty → 8.
- MG-AUD-2: populated G008/G009 world 7 → 8, every app row and ACL identical, repeat up.
- MG-AUD-3: missing or unclean schema refused, nothing changed.
- MG-AUD-4: populated audit refuses down, unchanged.
- MG-AUD-5: empty down restores the exact v7 catalog, then up (reapply).
- MG-AUD-6: rolled-back up restores the owner role and the v7 catalog.

**TestAudit (B-AUDIT-IT, 9 subtests).** AT-ATTEST reuses `rlsAttest` and the absence of MinIO root.

| ID | Proves |
|---|---|
| AT-INV | Every one of the 29 actions observed from real calls (`AUDIT-INV {"actions":29,"observed":29}`; G6 through the injected revocation-gap fault); trusted-primitive/bootstrap events labelled `unknown`, kept apart from user attribution |
| AT-FMT | exact durable fields/nullability for user, system and unknown events; negatives: forged actor column (42501), foreign or unestablished Organization (42501), **Group of another Organization and nonexistent Group UUID (42501), with the Organization owner/admin as appender** (r1), T-action direct insert (42501), bad action/reason/shape (23514), malformed correlation → NULL, foreign target → no resource, pre-context failure unattributed; positive: **an Organization owner/admin's G3 for an object of a Group they hold no binding in** carries that Group (r1); lifecycle actions with no context → `unknown` (Q3) |
| AT-ATOM | M1 (guarded rename) and M2 (sign-in via `inTxAudited`, both branches) success with `xmin` equality (same transaction) and exactly the catalogued events. A superuser-planted `CHECK (…) NOT VALID` fault: rename, sign-in, **rotation (`superseded` strict)**, **enable** and grants fail with neither effect, then succeed after removal. **Revocation priority (r1):** with the fault on `session.revoked`/`user.disabled` only, logout current/all/provider and disable **commit** (sessions revoked, user disabled), G6 is recorded and `audit.gap` reports `recorded`; with `CHECK (false)` the revocation still commits and `audit.gap` reports `failed`, with no audit row claimed. **Privilege fault (r2):** the attested superuser revokes the runtime's column INSERT privilege on `audit.events`, leaving the `app.sessions`/`app.users` UPDATE privileges intact. Logout current/all (including an expired-but-unrevoked session) and provider logout and disable then **commit**: the guarded trigger statements reach `audit.events` only inside M5, and `audit.gap` reports `failed` because G6 is refused too. Sign-in and `superseded` rotation fail atomically, with no session issued or superseded. The grants are then restored and re-checked against the CT-AUD-3 ACL inventory. A frozen identity pool (the G009 `newFreezer` plus a pgx tracer that freezes after the failed recorder insert) bounds the stalled rollback within `auditCleanupBudget` and the connection is not reused. Business rollback (slug conflict) → no event; a forbidden rename leaves G2 while the group is unchanged; a recording fault → denial unchanged plus one diagnostic |
| AT-ID | sign-in, rotation, logout current/all/provider (confirmed and failed), Disable/Enable (system); **fan-out with an expired-but-unrevoked session: logout-all and disable count one `session.revoked` per `revoked_at IS NULL` row, expired ones included** (r1); post-consume failures (unattributed `login_canceled`; attributed `user_disabled`, `identity_conflict`); pre-consume → none; global rows invisible to every tenant reader; a revocation on an unavailable (frozen) database → `identity_unavailable`, never success, no G6; lost COMMIT reply on a guarded write → at most one event, no retry |
| AT-OBJ | Put T20/T21, Delete T22/T23, Sign G3 + TTL with no URL/signature/key in any row; issuance fault → no URL; provider Put fault after T1 → G4 compensated; provider Delete fault → G5; T2 lost reply → reconcile → T21 or G4 (reusing the G009 `objSpy`/hook seams and residual evidence) |
| AT-READ | owner-only positive; admin, member and all Group roles forbidden plus zero SQL rows; multi-tenant owner of A / member of B; cross-tenant cursor → `invalid_cursor`; selected Group and shared profile confer nothing; revoked session → unauthenticated; downgrade → forbidden; pooled reuse; exact visible sets under raw runtime SQL per context |
| AT-LIM | page size and time-range bounds; cursor binding; deterministic continuation (strictly decreasing keys, no duplicates) under concurrent appends; E-AUDIT-1/2 lines (§11) |
| AT-PRIV | canary scan of every audit row after the success/deny/fault producers and of `ReadAuditEvents` output; a planted canary (rolled-back superuser row) is detected; no deletion path exists |

**Unit:**
- U-AUDIT: catalog parity Go ↔ migration ↔ AUDIT.md table; `Append` refuses T-actions; statement texts.
- U-ACCESS-BOUNDARY: files, imports and exported surface.
- U-RLS-BOUNDARY: 7 `set_config` (4 guarded `app.*`, 2 correlation, 1 constant gap reset); 5 `inContextTx` sites;
  `recordAuditEvents` initializers `initIdentity,initEmpty`.
- U-IDN-AUDIT (r1, existing unit file `identity_test.go`, the `countingDB` fake style):
  - the recorder's statement deadline equals `auditRecordBudget`;
  - its rollback context carries `auditCleanupBudget`, is not canceled at start, and a rollback blocked until its
    context ends returns within the budget;
  - caller cancellation does not cancel recording;
  - a context statement returning no row (prior context) discards the connection, returns `identity_unavailable` and
    sends no append;
  - the gap decision: marker + COMMIT ok → G6 attempted and `audit.gap` `recorded\|failed`; marker + COMMIT error →
    `not_attempted`, no G6; no marker → nothing.
- TestMigrationSet and TestAuditMigrationShape.

**Not added:** an independent oracle. The read rule is a single predicate over a small explicit case table, and the
catalog is a literal list. No second helper framework, and no repetition of predecessor role/action matrices: the
inherited suites run at version 8 instead (§10).

### 9.1 Acceptance-criteria mapping

| AC | Evidence |
|---|---|
| AC1 | §2/§3 inventory; AT-INV; CT-AUD-4; U-AUDIT parity |
| AC2 | AT-FMT; AT-ID unattributed cases; CHECK and shape constraints |
| AC3 | TestAuditCatalog CT-AUD-1..6 and NEG; AT-FMT forged negatives |
| AC4 | AT-ATOM (`inTxAudited`/trigger/xmin, fault, rollback, duplicates, independent denial, sanitized failure; strict paths stay atomic) |
| AC5 | AT-ID; AT-ATOM revocation-priority faults (committed revocation, G6/`audit.gap`, DB failure truthful); U-IDN-AUDIT (bounded recorder, poison, cancellation, unknown COMMIT); inherited B-RLS-IT pool and poison cases at v8 |
| AC6 | AT-OBJ; inherited B-OBJ-IT/ORDERING/FAULTS at v8 |
| AC7 | AT-READ; RLS `events_select` |
| AC8 | AT-LIM; E-AUDIT-1..3; PG-004 assessment (§11) |
| AC9 | AT-PRIV; CT-AUD-5; AUDIT.md G013 contract |
| AC10 | TestAuditMigration; persistence seed/restart/verify; B-MIGRATE-SET (00007 pin) |
| AC11 | §10 wiring, b4 (the only broad run), B-REQUIRED-CHECKS counts, no skips, Result provenance |
| AC12 | S-review, Result, B7 ledger, cleanup receipt, staging disposition, G012/G013 handoffs |

## 10. Validation

### 10.1 Batches

Before every batch:
1. `git fetch origin`.
2. Apply the exact-delta rule (record SHAs and reason), or rebase/reconcile.
3. `bash infra/ci/docker-desktop-mounts.sh` (stop on failure).
4. Take the harness flock.
5. Record target and candidate SHAs.

Always reconcile before b4 and before the MR. **r1 (item 5):** there is no separate covering batch. Standard form: `bash infra/validation/run.sh --base <target> --allow-file
management/v0/results/G011-scope-allowlist.txt --report-dir <EVID>/<batch> <suites>`. The Result stays at base bytes
while validating (B-PROTECTED).

| Batch | Content | Purpose | Wall (forecast) |
|---|---|---|---|
| b1 focus-go | gofmt, `go vet` (with `-tags integration`), unit for `audit`, `tenancy`, `access`, `identity` (U-IDN-AUDIT), `migrations` in the core-check image | inner loop | ~2 min |
| b2 focus-db | the B-AUDIT tests plus the affected inherited tests (selected `-run`) on a focus stack (pinned PostgreSQL + built MinIO; IDs in the run inventory at creation) | inner loop; repeat | ~3–6 min per run |
| S-review | independent read-only security review (§12) | fix, rerun affected b1/b2 | — |
| b4 final | reconcile, then `all unit` | the **only** broad run. It carries every former b3 obligation (B-MIGRATE-SET and digests, every migrations-suite check at v8 incl. B-RLS-*, B-OBJ-* with G010's attack invocation, B-POLICY-*, B-AUDIT-*, unit) plus the rest of full. **ci-controls is selected** (infra/* changes) and runs in the background. B-REQUIRED-CHECKS: every ID PASS once. The new B-AUDIT wiring is proven here by its `it_count`/`AUDIT-INV` pins (a regex miss fails), with no separate wiring run. A b4 failure reruns only the failed suites and checks its fixes affect (finish → fix → rerun) | ~50–55 min (G009 b18 2788 s, plus G010's and G011's ~4 min) |

There is no manual or operator stage (no UI, no live provider). Total local machine time is about **64–105 min
(about 1.1–1.75 h)** without failures: b1 ~2 min, b2 about 4–8 runs of 3–6 min (12–48 min), b4 ~50–55 min. This is
the r2 correction of r1's understated 1.1–1.4 h; r0's estimate was 1.5–2 h with b3. Unchanged passing batches are never repeated for a SHA-only or management-only delta.

### 10.2 Version-8 reconciliation of accepted checks (RC; guarantees preserved)

| Location | Change |
|---|---|
| `run.sh` `static_migration_set` (~995) | set + 00008; add `MIGRATION_00007_SHA` |
| root migrate block (~1460) | 00008; `current_version 0/8/8` |
| B-MIGRATE-CLI/CATALOG (~2015) | reapply/rollback act on 00008; app relation count stays 38 (audit objects are outside `app`) |
| B-PERSIST (~3242), B-DBRENAME (~3293), B-TOGGLE-OFF (~3403) | version 8 |
| `migrate/integration_test.go` | latest-version outputs; `reset()` first drops audit objects (CASCADE) |
| `migrate/objects_integration_test.go`, `rls_integration_test.go` | version-7 checks use `subsetFS(t, 7)` (v7 catalogs stay v7 evidence) |
| `tenancy/integration_test.go` | step-down chain 8→7→6→5, with the disposable superuser audit purge first |
| `access/integration_test.go` | latest version is 8, so B-POLICY/B-RLS/B-OBJ ITs run **with triggers active** (inherited regression) |
| `access/boundary_test.go` | §9 unit pins |
| `run.sh` `obj_it` B-OBJ-IT PASS text | "at version 7" → "at version 8" only; G010's `TestIsolationAttacks` invocation and its 17 / 82-0 pins stay as integrated (`c961e56`) |
| `migrations_test.go`, `runner_test.go` | set through 00008 |

Expected **observation** changes, not regressions:
- E-RLS-1 parser counts grow for audited writes;
- E-PG-2/E-OBJ-2 lock inventories and business-statement reader counts are expected to be unchanged, because capture
  rows make no reader call;
- the new direct-append reader calls (§11) occur only in the separate audit transactions.

Per the r0 review, old counts are not required to stay identical where pinned audit work explains an increment. Any
increment is attributed to a §5/§11 mechanism in the Result, or escalated as an E-AUDIT-TRIGGER. The implementer inventories every pin with
`grep -nE '00007|version 7|current_version|stored_objects\.sql|97|69|98|104'` at binding and records each RC in the
Result.

### 10.3 CI selection (actual paths)

- `core/migrations/`, `core/internal/{tenancy,identity,migrate}/` and `infra/` select **full**; `infra/*` also selects
  **ci-controls**.
- `core/internal/audit/` is registered as foundation, as G009 did for `objectstore`, because its covering checks run
  only in `migrations`.
- Docs alone would be fast. No affected-only hosted run is promised.

### 10.4 Hosted

1. After Codex's pre-integration review: reconcile, push, `glab` MR to `staging` with source-branch deletion; auto-merge
   armed once the Codex/Duo holds clear.
2. MR fast runs automatically; the optional manual gate is not requested.
3. The post-merge staging pipeline (full gate plus ci-controls, ~41–45 min) is **required**: success or owner-canceled
   satisfies it (canceled recorded as canceled); a failed job fails; pending does not qualify.

### 10.5 Provenance and evidence

- Git-first. The Result records per batch: candidate/target commits, commands, outcomes (failed, interrupted and
  skipped explicit), core-check/PostgreSQL/MinIO image IDs, the Go version and CI links.
- Durable proof lives in the Result and CI artifacts; local raw output is temporary (2026-10-14 policy) and purged at
  closeout, keeping the compact record.
- Never retained: secrets, sessions, signed URLs, payloads.

### 10.6 Proposed B-SCOPE allow-list (`management/v0/results/G011-scope-allowlist.txt`, imported by Codex at binding)

```text
# Create
^core/migrations/00008_audit_events\.sql$
^core/internal/audit/(audit|store)(_test)?\.go$
^core/internal/tenancy/access/audit(_integration_test)?\.go$
^core/internal/migrate/audit_integration_test\.go$
^docs/foundation/AUDIT\.md$
^management/v0/results/G011-(scope-allowlist\.txt|durable-audit-events-result\.md)$
# Modify: production
^infra/postgres/(init/10-roles\.sh|provision-roles\.sql)$
^core/internal/tenancy/(store|facts)\.go$
^core/internal/identity/(store|service)\.go$
^core/internal/tenancy/access/(tx|facade|context|errors|objects|log)\.go$
# Modify: unit pins and RC (§10.2)
^core/internal/tenancy/access/(boundary|integration)_test\.go$
^core/internal/identity/identity_test\.go$
^core/migrations/migrations_test\.go$
^core/internal/migrate/(integration|objects_integration|rls_integration|runner)_test\.go$
^core/internal/tenancy/integration_test\.go$
# Modify: harness, catalogs and counts
^infra/validation/(run\.sh|required-checks\.tsv|render-customizations\.tsv)$
^infra/validation/tools/internal/logrecord/(events\.json|logrecord_test\.go)$
^infra/ci/(validate-controls\.sh|profile-paths\.tsv)$
# Modify: docs
^docs/foundation/(DATABASE-ISOLATION|AUTHORIZATION|OBJECT-STORAGE|IDENTITY|MIGRATIONS|TENANCY|FOLLOW-UP|OPERATIONS)\.md$
```

Conditional entries (added only with a Codex-reviewed amendment): grep-inventory RC hits outside this list and
`core/internal/tenancy/boundary_test.go` if its pins move. The integrated G010 attack tests (`attack_*_test.go`) need no
edit (§0) and are deliberately **not** listed.

## 11. PG-004 assessment, corrected in r1 ([register](../PERFORMANCE.md#pg-004--g011-durable-audit-write-and-read-cost); no benchmark, no numeric claim)

Round trips are client↔server statements, counting BEGIN and COMMIT. Reader calls mean `app.rls_org_role` /
`app.rls_group_role` invocations.

| Path | Added statements / round trips | Database work, readers and locks |
|---|---|---|
| Audited row change (M1 row triggers) | **+0**; the context statement carries correlation and the gap reset in the same trip | +1 audit insert per affected row: heap + PK (random UUID) + partial tenant index (tenant rows only). BEFORE trigger and INSERT policy: parser calls only, **no reader call**. WHEN/`UPDATE OF` skip no-ops and unaudited columns. `RowExclusiveLock` on `audit.events` only; no app-table lock; no FK |
| Every `app.sessions` UPDATE (statement trigger, r1) | +0 | transition-table capture of the updated row(s) plus one function call with one aggregate query over the transition tables. This includes renewal (≤ once per 60 s per active session), provider bookkeeping and logout-reference return. **No audit statement executes** (no privilege check or lock on `audit.events`) unless a `superseded` or priority revocation transition is present (r2). A deliberate cost of revocation priority with uncapped fan-out |
| Revocation (M5: logout current/all/provider, DisableUser) | **+1** (gap-marker read before COMMIT) | **uncapped fan-out (r1, item 6):** `revokeAll` and DisableUser update every `revoked_at IS NULL` session of the user, including expired-but-unrevoked rows (nothing prunes sessions), so one `session.revoked` row per such session. It is written by **one** `INSERT … SELECT` in **one** subtransaction per statement (DisableUser: +1 for the user row), never one per row, so there is no subtransaction-cache overflow. The revocation's existing N-row UPDATE and row locks are unchanged |
| Revocation with an audit gap (fault path) | **+4** (G6 through the bounded identity recorder) | rare; ≤ `auditRecordBudget` 2 s + `auditCleanupBudget` 2 s |
| Sign-in success (M2) | **+1** (G1 insert in the existing `signIn` transaction) | one insert, plus the M1 inserts of T14/T17/T18 |
| Independent record (M3: G1 denied/failed, G2, G4, G5) | **+4** (BEGIN, context, INSERT, COMMIT) on a second short pooled connection, **after** the business connection is released | reader calls: 0 when org-less; 1 (`rls_org_role`) with an Organization; with a Group (r1, item 1) +1 PK read of `app.groups` under RLS, whose policy adds one `rls_org_role` init-plan and ≤1 `rls_group_role`, so **≤3**. Budgets 2 s + 2 s; detached from caller cancellation |
| Issuance (M4, G3) | **+4** per signed URL (as M3, after the unchanged read guard and the offline presign) | as M3; ≤3 reader calls for a Group object |
| Identity recorder (G1 failures, G6) | **+4** (BEGIN, init, INSERT, COMMIT) | 0 reader calls (org-less); bounded 2 s + 2 s with disposal |
| Tenant read | one statement | RLS init-plans (`ctx_org`, `rls_org_role` once) + backward index range scan of `events_tenant_time_idx`, LIMIT ≤ 101 |

**Workload assumptions.** Event rate, retention, concurrency and per-user session history are unknown. Session rows
accumulate (no pruning exists), so logout-all/disable cost grows with a user's historical unrevoked sessions. Audit
growth is unbounded until G013.

**Flagged concerns for owner visibility:**
1. An extra 4-round-trip transaction per durable denial, per independent storage outcome and per signed URL, with brief
   occupancy of a second pooled connection.
2. Uncapped revocation fan-out: N audit rows and N PK entries per logout-all/disable, where N includes
   expired-but-unrevoked sessions.
3. Statement-trigger overhead on every `app.sessions` UPDATE, including renewal.
4. +1 round trip per revocation transaction.
5. Random-UUID PK write amplification.
6. Unbounded table growth before G013.
7. Anonymous authentication-failure events are bounded per consumed login transaction; authenticated denial floods are
   bounded only by future adapter admission/rate limits (the existing PG-001 obligation).

No session pruning, index, partitioning, queue or benchmark is proposed (none is authorized). The owner's
revocation-priority decision is a security decision, not performance acceptance.

**Smallest evidence during execution:**
- **E-AUDIT-1:** statement/transaction counts for one representative operation per mechanism. This includes the
  revocation marker read and a logout-all over ≥2 sessions, one of them expired-but-unrevoked, which must show one
  statement-level insert of N rows.
- **E-AUDIT-2:** `EXPLAIN (ANALYZE, BUFFERS)` of the tenant read on a two-tenant fixture, with the environment
  recorded.
- **E-AUDIT-3:** inherited E-RLS-1/E-OBJ-1 reader counts and E-PG-2/E-OBJ-2 lock inventories at v8, with any increment
  attributed to a pinned mechanism.

Unexplained scans, fan-out, reader growth or lock changes produce E-AUDIT-TRIGGER lines and return to Codex/owner before
any design change. **Owner disposition requested (Q7):** proceed with this bounded evidence and unknown workload limits,
or set targets first. No inherited PG-001–003 waiver applies.

## 12. Delegation and sequence

**Roles:**
- **Implementer:** one owner of all files (task G011, role impl; path `g011-agent-<id>-impl`, or the recorded
  native-tool exception). Default effort high; the owner may choose xhigh for the migration/attribution work, and the
  operator is asked at execution start.
- **No database/Go split:** the migration, triggers, Go recorders and tests are tightly coupled, and splitting them
  would overlap files.
- **One independent read-only reviewer** after b2 (role review, high effort): attribution and source separation,
  transaction/failure and unknown-commit truth, **the narrowness of the M5 revocation exception** (only the five pinned
  transitions, only the audit insert inside the block, strict everywhere else), the direct-append Group proof,
  grants/RLS/runtime-role tests, privacy, and the RC list's non-weakening. Findings are fixed by the implementer.
- **Shared stack:** the coordinator owns the shared stack and harness lock (G010 must be finally accepted first).

**Sequence:**
1. Binding: G010 final ACCEPT and reconciliation of any post-merge correction (A-G010-1..4), migration number,
   allow-list and Result import.
2. Provisioning + 00008 + TestAuditCatalog/Migration + migrate RC.
3. `audit` package + U-AUDIT.
4. Context statement and `InitContext` + U-RLS-BOUNDARY.
5. Identity (`inTxAudited`/`signIn`, bounded `recordAuditEvents`, post-consume recording, revocation gap) +
   U-IDN-AUDIT.
6. Access (failure scope, `recordDenial`, `auditTx`, storage outcomes, issuance, `ReadAuditEvents`, logs) +
   U-ACCESS-BOUNDARY.
7. TestAudit + remaining RC.
8. Harness wiring, catalogs, counts, profile path.
9. Docs.
10. b1/b2 loops → S-review → fixes and affected b1/b2 reruns → reconcile → b4 (the only broad run) → Result.
11. Safekeeping push after each round (no MR) → Codex review → MR.

## 13. Budget, forecast and scope

**B7**, shared attempts, with a checkpoint after attempt 4; planning spends none. **Forecast: 2 likely, 3 plausible:**

| Attempt | Expected content |
|---|---|
| 1 | implementation through b4, review-ready |
| 2 | one bounded Codex correction (likely, given the size and the security tier) |
| 3 | only hosted/ci-controls drift or a G010 reconciliation found late |
| 4+ | only if a settled ruling is reversed after implementation; the checkpoint is not expected to be reached |

**Context and scope:** one bounded boundary. The estimated diff (~4–5k lines including tests and harness) is comparable
to G009's single cycle. **No split is recommended.** If the checkpoint after attempt 4 shows B7 at risk, the proposed
split is G011a (schema, triggers, identity, read) and G011b (access denials, storage outcome events, issuance), with no
weakening of tests.

## 14. Escalation items for explicit review; risks

| # | Item | Class |
|---|---|---|
| E1 | Provisioning creates `audit` (the owner lacks database CREATE); existing volumes re-run provision-roles.sql; infra change → ci-controls | new infrastructure/security boundary |
| E2 | Runtime column-limited INSERT + BEFORE attribution trigger + `pg_trigger_depth()` source separation, instead of a SECURITY DEFINER writer role. **Accepted in principle (Q1)**, subject to the r1 Group proof and implementation review | security model |
| E3 | Actor rule: `system` only for lifecycle actions inside the lifecycle producer's context; `unknown` for no-context privileged writes and trusted primitives; D1 residuals recorded. **Q3 applied** | security model / attribution |
| E4 | Accepted G008 context statement gains two unguarded transaction-local correlation settings and a constant gap-marker reset; U-RLS-BOUNDARY pins updated; correlation is now stored (diagnostic only) | change to accepted artifact |
| E5 | `tenancy.InitContext` accepts zero Organization + zero Group (IDENTITY shape) for org-less audit records | change to accepted contract |
| E6 | Accepted-test and harness pin amendments (§10.2; boundary tests; +3 required IDs and count pins) | validation-scope change |
| E7 | **Superseded in r1 by the owner decision** (Goal, Spec R4/AC5, ADR-006 clarification). M5: an audit-only fault cannot prevent local logout (current/all/provider) or disable. The audit insert is rolled back inside a bounded exception block limited to five pinned transitions; the revocation commits; G6 is attempted after COMMIT; `audit.gap` reports the gap. A revocation database failure or unknown COMMIT stays `identity_unavailable`. Sign-in, rotation, enable, grants and all other mutations stay strictly atomic. No runtime trigger-disable or bypass privilege | security exception (owner-decided) |
| E8 | Durable-denial granularity (§3): no unauthenticated, binding-misuse, pre-consume, logout-transport or successful-Resolve events | catalog scope |
| E9 | No public surface; new internal guarded op and log event only | contract (none public) |
| E10 | No FKs; populated down refuses; history starts at v8; unbounded growth until G013 | irreversible data decision |
| E11 | PG-004 corrected costs (§11): 4-round-trip independent records and issuance, ≤3 reader calls with a Group, uncapped revocation fan-out, statement-trigger overhead on every sessions UPDATE, +1 round trip per revocation. Owner disposition pending (Q7) | performance risk |
| E12 | Provisional G010 assumptions (§0) | dependency |
| E13 | Live identity behavior: sign-in and enable depend on the audit append; logout/disable complete with a reported gap on an audit-only fault; deploy order provision → migrate → code | production behavior |

**Risks:**
- R-a: inherited ITs at v8 expose a trigger interaction. Mitigation: b2 runs the affected inherited tests early; a pin
  is fixed only with its guarantee preserved.
- R-b: `pg_trigger_depth()` semantics. Mitigation: AT-FMT negatives and CT-AUD; documented PostgreSQL behavior; the
  fixture server is 17.x.
- R-c: concurrent-append tests must be deterministic. Mitigation: barriers, no sleeps.
- R-d (r1): the transition-table statement trigger and the M5 exception blocks are new trigger forms. Mitigation:
  CT-AUD-4 pins their exact definitions and bodies; AT-ATOM proves both the strict and the protected branches; the
  fixture server is 17.x.

## 15. Rulings applied and open items (r1)

| # | r0 review ruling | Applied |
|---|---|---|
| Q1 | invoker / column grant / depth approach accepted in principle | §1, §5.3; subject to the r1 Group proof (§4.4) and implementation review |
| Q2 | decided by the owner's revocation-priority decision; not asked again | M5 (§4.1, §4.2, §5.3, §6); E7 superseded |
| Q3 | `unknown` where no verified initiator exists; `system` only from the defined lifecycle producer | §2 actor rule |
| Q4 | privileged-operator-only global inspection; no platform-admin role | §5.4 (unchanged) |
| Q5 | three check IDs | §10 (unchanged) |
| Q6 | distinct `auth.sign_in` and `session.issued`; one producer per event | §2 (unchanged); G6 is its own event |
| Q8 | defer retention `USAGE` until G013 | §5.4, §5.5, CT-AUD-3 |

**Open:** Q7, the PG-004 owner disposition on the corrected §11 estimate. Execution also still waits for G010 final
ACCEPT, Plan ACCEPT and binding.

## 16. Closeout and handoffs

- G011 opens EPIC-004 and does not close it, so there is no epic-closeout compaction checkpoint; Codex makes the
  ordinary post-ACCEPT assessment.
- Bounded cleanup: one scoped preflight, one removal pass and a compact receipt. Focus resources are recorded at
  creation. Uncertain items are held; native refusals are reported once.
- **G012** becomes dependency-ready.
- **G013** receives:
  - the `<db>_retention` maintenance contract (§5.4; no G011 privilege, Q8);
  - the pending/tombstone object linkage (G4/G5 events);
  - privileged inspection and backup expectations.
- G021 stays closed.

## Revision history

- r0, 2026-10-14 UTC: initial provisional Plan on `1849bd4` (dispatched on `31963ae`; the governance/management-only
  deltas were rebased before commit). It labels G010 assumptions, checked read-only against G010's pre-integration
  candidate `71c8c47`, and records the gates. No code, test, build, Docker or probe was run.
- r1, 2026-10-14 UTC: bounded revision for the r0 review (not included) CORRECTION LOOP,
  on staging `74a3ffc`, which includes the G010 merge `c961e56` and the updated Goal/Spec/ADR-006/PERFORMANCE.
  - Six corrections (map at top): the Group proof, the bounded identity recorder, the `inTxAudited` owning helper, M5
    revocation priority (owner decision; replaces r0 E7), b3 folded into b4, and corrected fan-out and PG-004 costs.
  - Rulings Q1, Q3–Q6 and Q8 applied; Q2 decided; Q7 pending the owner.
  - A-G010-1..4 reconciled against the integrated G010 bytes.
  - r0 archived byte-identically. No code, test, build, Docker or probe was run. B7/checkpoint4 and the twelve ACs are
    unchanged.
- r2, 2026-10-14 UTC: one-item revision for the r1 review (not included) CORRECTION LOOP,
  on staging `102fd16`.
  - The strict `superseded` INSERT, and every audit statement of the session-revocation trigger, is guarded by a
    transition-table-only flag, so an audit privilege fault cannot abort a revocation before M5.
  - The renewal-cost text is updated; AT-ATOM gains the column-INSERT privilege fault case.
  - The local forecast is corrected to 64–105 min.
  - r1 archived byte-identically (`dd4406b5…`); r0 archive unchanged. No other section, ruling or gate changed. No code,
    test, build, Docker or probe was run.
