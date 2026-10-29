> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G058 Implementation Plan — Generic Organization enrollment and self-signup

Revision **r3 (provisional; two bounded SQL corrections of r2 plus the owner package)**. 2026-10-18 UTC. Author:
Claude (model **Fable 5.1**; planning effort "xhigh (requested)" — the runtime setting is not verifiable from inside
the worker, actual value **unknown**). Native isolated worktree, generated name `agent-<id>` (recreated at
the same path for r3; AGENTS "Task-prefixed Claude worktrees" fallback; task G058, role plan r3 correction; not
renamed; ownership recorded here). Branch `g058-plan-r3` from `origin/staging`
`2e209db11bc68590d0622fbf82305d89a76ab94b` (verified ancestor of HEAD). Coordinator session
`019EKrFpe3tJhYgzBhHthyD6`.

**Inputs of this revision (r3):** Codex review `management/v0/results/G058-plan-codex-review-r2.md` (SHA-256
`45336e915f033f48a9937e3435210da571b226948fa91723b00c4460fb429d86`) — **binding**; its two SQL corrections and the
owner-package instruction are mapped in §0.C. The r2 archives `plans/G058-generic-organization-enrollment-plan-r2.md`
(`97145547…a1e`, 808 lines) and `plans/G058-scope-allowlist-r2.txt` (`50a045c4…2a9`, 45 patterns) are the base of this
text and are **unchanged**. r2's inputs, preserved: Codex review r1 (`6a7d7aad…13fc02`) and the preserved archives
`plans/G058-generic-organization-enrollment-plan-r0.md` (`a022c35d…2f2b`, 744 lines),
`plans/G058-generic-organization-enrollment-plan-r1.md` (`f82875d4…079e`, 563 lines), `plans/G058-scope-allowlist-r0.txt`
(`b8fd6d2e…545a8`, 43 patterns) and `plans/G058-scope-allowlist-r1.txt` (`bbacc370…9df0`, 44 patterns), all
**unchanged**; the r0 review (`dd0d22e7…9d100b`); the G058 Goal/Spec; targeted read-only source facts at `c5f83fbf`
(cited file:line below) and the provisional G019 branch `origin/g019-scoped-invitations-membership-ownership` (draft
MR !43; `00015_membership_invitations.sql` read at the branch head) — **provisional, not accepted**. No implementation,
test, build, Docker, install, database or provider probe, and no MR. No G020 planning and no Spec authoring (hard
stop). Tier: architecture/security/data. Budget **provisional B7, checkpoint after attempt 4, unspent** (§11.1);
planning spends no attempt. r2 reading was capped (≈48k tokens: the review, the cited r1 sections and the code spots
each correction needs); r3 reading was bounded to the r2 review, the r2 Plan lines it cites and the one SQL precedent
(`rls_user_enabled_lock` body and the 00015:877-881 EXECUTE grants); no unrestricted repository reread.

**Trigger, required reading, reused non-G019 interfaces and the P19-1..P19-12 table are unchanged from r0 §0** and are
not repeated here except where a correction touches them (P19-1/2/9/10/11 are re-cited with the exact 00015 lines read
at the branch head: claim policies `:642-657`, cw column grants `:633-639`, `rls_user_enabled_lock` self branch
`:692-694`, `invitation_claim_organization` `:796-875`, EXECUTE grants `:877-881`, `role_bindings_authority_guard`
`:945-976`, `invitations_bound` Organization lock `:485`). The r1 tool refusal is preserved verbatim in §12; **no
refusal occurred during r2**. The SDK `EmailVerified` projection stays **unqualified and held** (§1.8): the module
cache was not read and the refusal was not routed around. **Every r0 correction resolved by the r1 review ("Resolved
directions") is preserved**: audit-derived authority and the cooldown stay withdrawn, durable restrictions stay
escalated (D-4r1), lock conflicts, the separated applicant audit path, approval revocation, continuity-first login,
cookie path/intent binding, the bootstrap reader and honest staged outcomes are unchanged except where §0.B says
otherwise.

## 0.C Correction map (review r2 → r3)

| Review item | r3 section(s) | Change |
|---|---|---|
| 1 Self-leave cannot depend on authority after deleting it | §1.4 (departure body order), §4.3 (**`app.enrollment_lock_organization`** helper; departure/decision bodies), §4.4 (EXECUTE), §6.2 ENR-22, §11.2 | r2's departure body re-read the Organization through the owner-shared tenant policies (`organizations_update USING (id = ctx_org() AND rls_org_role() IS NOT NULL)`, 00006:434) **after** the binding was gone: on a self-leave of the last membership `rls_org_role()` is already NULL, the row is invisible, `FOUND` fails and `LeaveScope` rolls back. r3 pins a **narrow cw-owned helper that stays valid after deletion**: `app.enrollment_lock_organization(p_org uuid) RETURNS boolean` SECURITY DEFINER owned by `<db>_claim_writer`, body `PERFORM 1 FROM app.organizations WHERE id = p_org FOR NO KEY UPDATE; RETURN FOUND;` — it locks through the r2 lock-only pair `organizations_claim_lock` (`USING (true) WITH CHECK (false)`) + cw `UPDATE (enrollment_mode)`, so visibility never depends on the caller's remaining membership and no tenant row read is widened (the helper returns a boolean, reads no column). EXECUTE revoked from PUBLIC and granted to **`<db>_owner` only** (the executing role of the two owner-owned trigger bodies; the `invitation_claim_organization` grant shape, 00015:881): the runtime cannot call it, so direct-SQL authority protection is unchanged. Both trigger bodies call it and RAISE if it returns false (the FK makes that unreachable). The `ctx_org()`-conditional of r2 is removed; the replay path takes the same lock. Tests: only the existing G019 self-leave case (last Organization membership → one `departed` row, `LeaveScope` commits) and ENR-22's rollback case are extended |
| 2 Approval authority version must be authoritative in SQL | §1.4 (approvals), §4.4 (rt UPDATE columns), §4.5 (`enrollment_requests_immutable`), §6.2 ENR-23, §11.2 | r2 let the facade compute `max(seq)` and let rt write `authority_seq` with only a non-negative/immutable-after check, so an inflated watermark could pre-authorize later restrictions. r3: **rt holds no column privilege on `authority_seq`** (removed from the rt UPDATE grant; a direct `SET authority_seq = …` is 42501 before any trigger runs), and the BEFORE UPDATE guard `enrollment_requests_immutable` **derives** it: on `pending → approved` it first takes `PERFORM 1 FROM app.organizations WHERE id = NEW.organization_id FOR NO KEY UPDATE` (invoker = the deciding rt session, which passes `organizations_update` as a current MGR member and already holds the lock from (3′) — re-locking in the same transaction is a no-op), then assigns `NEW.authority_seq := coalesce((SELECT max(seq) FROM app.enrollment_restrictions WHERE organization_id = NEW.organization_id AND user_id = NEW.user_id), 0)` through rt's `enrollment_restrictions_select` (`organization_id = O AND MGR` branch), ignoring any supplied value; in every other transition `NEW.authority_seq IS DISTINCT FROM OLD.authority_seq` raises (first-decision immutability preserved; `denied` stays 0; cw's consumption UPDATE never touches it). The facade no longer computes the watermark. Tests: one direct-SQL inflated-watermark negative (42501) plus the ordinary approval (derived value equals the greatest committed `seq`, and a later restriction is refused at admission) |
| Owner package | §10 | D-4r1, D-15 and **D-16 all UNSELECTED** in one coherent package; D-16 now presents Codex root's recommendation (fail-closed open-mode readiness for Organizations with unresolved legacy history) alongside the options, the "start empty" default recommendation is **withdrawn**, and the exact pre-00016 recovery consequences stay visible; D-17 and D-18 unchanged |
| Pins/scope | §11.2, §11.3 | SECURITY DEFINER allowance 00016 **8** (cw 3); cw-owned functions 3 → **6**; rt `enrollment_requests` UPDATE(**5**); EXECUTE grants to ow +1; scope unchanged (**45** patterns; the helper lives in 00016) |

## 0.B Correction map (review r1 → r2) — preserved from r2

| Review item | r2 section(s) | Change |
|---|---|---|
| 1 Executable applicant lock grants | §4.4 (`organizations` cw), §6.2 CT-ENR-1 / ENR-21 | cw gains **`GRANT UPDATE (enrollment_mode) ON app.organizations`** — the column privilege `FOR SHARE` requires (one UPDATE-able column) — paired with the unchanged `organizations_claim_lock` `USING (true) WITH CHECK (false)` (the exact `users_claim_lock` precedent: `GRANT UPDATE (status)` 00015:638 + `WITH CHECK (false)` :653). SELECT stays `(id, enrollment_mode)`. New catalog/lock negatives: cw `UPDATE enrollment_mode` → 42501 policy violation; cw `UPDATE name`/`SELECT name` → 42501 no privilege; cw `SELECT id, enrollment_mode … FOR SHARE` → one row. No runtime tenant read widened |
| 2 FORCE ledger insertion | §4.3 (trigger bodies), §4.4 (`enrollment_restrictions_owner_insert`), §4.5, §6.2 ENR-22 | Owner-only, trigger-bound **INSERT policy** `enrollment_restrictions_owner_insert FOR INSERT TO ow WITH CHECK (pg_trigger_depth() >= 1 AND lifted_at IS NULL AND lifted_by_user_id IS NULL AND lifted_via IS NULL AND kind IN (…) AND recorded_via = 'trigger')` (the owner is bound by FORCE RLS, 00015:108-127; `TO … ow` policies are the 00006/00008 shape). Both DEFINER bodies assert `TG_WHEN='AFTER'`, `TG_TABLE_SCHEMA/NAME`, `TG_OP`, depth ≥ 1 and insert only values copied from `OLD`/`NEW`. No runtime, cw, pr or ar INSERT exists; an owner INSERT at depth 0 is a policy violation. ENR-22 proves the real paths (`RemoveMember`, `LeaveScope`, G013 replay delete, `denied`, `revoked`) and atomic rollback (policy dropped inside the test transaction → the deleting/deciding statement fails → binding/request unchanged, zero ledger rows) |
| 3 Replay no-op and legacy state | §1.4 (restrictions), §3, §6.2 ENR-16, §10 **D-16**, §11.3 (+1 scope) | r1's "no-op replay leaves the restored row" was only true when the checkpoint held the row. r2 adds **bounded replay materialization** to `infra/recovery/sql/replay-restrict.sql` (one guarded statement after the delete, `membership.remove` + `scope_type='organization'` only, `recorded_via='replay'`, skipped when the ledger table is absent in the restored checkpoint; a third keyed output line `ledger<TAB>…` that `recovery.sh`'s `kv()` tolerates, `recovery.sh:822`). Scope +1 pattern (exact). Existing V8 negatives retained; ENR-16 gains the "checkpoint predates grant and removal" subcase. **Legacy departures before 00016 get no ledger row** (no inference from audit history) → **D-16 product choice escalated**, with the readiness assumptions stated |
| 4 Fresh reinstatement authority | §1.4 (**Reinstatement authority**), §4.2 (`seq`, `authority_seq`), §4.3 (`enrollment_admit` steps 5/6/9), §6.2 ENR-10 O-ENR-6..8, ENR-23 | **Authority version for approvals**: `enrollment_restrictions.seq` (identity column) and `enrollment_requests.authority_seq` written at approval = the greatest committed `seq` of the subject's rows for the Organization; admission lifts only rows with `seq <= authority_seq` and refuses (`enrollment_restricted`, no binding) when any unlifted row is newer. **Chronology for invitations**: `lift` compares `enrollment_restrictions.created_at` (`clock_timestamp()` stamped under the Organization `FOR NO KEY UPDATE` lock the departure body takes) with `invitations.created_at` (00015:144; G019's insert holds the same lock, :485); a newer restriction makes `lift` **RAISE** so the claim in the same transaction rolls back (invitation stays pending; re-invite lifts). No G019 row is modified; no restriction is reset. Races covered in ENR-10 |
| 5 T3 transaction/principal binding | §1.3 steps 4–6, §4.2 (columns), §5.1, §5.2, §6.2 ENR-11/19, §10 **D-17** | **Route chosen: the explicitly permitted fresh ordinary sign-in after admission** (D-17). `IssueAfterAdmission`, `auth_time`, `consumed_at`, `issued_session_id` and the `admitted_sign_in_required` outcome are **removed**; T2's `admitted_at`/`admitted_binding_id` remain the durable proof; `admitted` answers `{"next":"sign_in"}`, the status page starts the unchanged ordinary login, and `signIn`'s `issuanceGuard` (membership) is the only issuance authority. Unknown-commit handling stays truthful (T1/T2 by read; T3 = existing `identity_unavailable`). Justification in §1.3 |
| 6 Cycle-free DTO ownership | §5.1, §11.2 (importers) | `Admitter`'s selector and result DTOs are **identity-owned** (`identity.AdmitSelector`, `identity.AdmitOutcome`, `identity.InvitationRef = [16]byte`); access converts to/from `tenancy.InvitationID`/`enrollment.*` and identity never imports access **or tenancy** (identity's current import set is unchanged). `EvidenceMinter` is an exported identity struct with an unexported seal, constructed only by identity and **passed per call** in the existing `Exchange` input (`identity.go:24`); no unexported interface method, no constructor wiring in `main.go`, no new `tenancy` importer |
| 7 Preserve logout scope | §1.3 step 5, §4.4 (`enrollment_sessions` rt UPDATE), §6.2 ENR-19, §10 **D-18** | **Bound paths** (`revokeSession`: logout and provider logout, `store.go:852`; supersession in `signIn`, `:663-665`) revoke only enrollment rows with the revoked session's `provider_session_id` (`app.sessions.provider_session_id`, 00002:57). **Broader paths** (`revokeAll`: logout-all, `:872`; `DisableUser`, `:958-985`) revoke every unrevoked row of the user, as they already do for sessions. Each UPDATE locks the rows it touches and therefore serializes with `enrollment_admit` step (2) (`FOR UPDATE`, EvalPlanQual recheck of `revoked_at IS NULL`); T3 (ordinary sign-in) touches no enrollment row. The one semantic flagged for review is in D-18 |
| Pins/scope | §11.2, §11.3 | policies +**21**; grants recomputed; `enrollment_sessions` 13 columns; scope **45** patterns (+`infra/recovery/sql/replay-restrict[.]sql`); importers of `identity` and `access` +1 each (`enrollment/api`); D-4r1 and D-15 **unselected** |

## 0.A Correction map (review r0 → r1) — preserved from r1

| Review item | r1 section(s) | Change |
|---|---|---|
| 1 Durable restrictions, not audit history | §1.4, §4.2 (`enrollment_restrictions`), §4.3 (`enrollment_admit` kinds), §4.5 (two restriction triggers), §10 **D-4r1 (escalation)** | `rls_prior_membership` and the pr audit SELECT grant are **withdrawn** (ADR-006 read boundary untouched). A durable, append-only `app.enrollment_restrictions` ledger written by owner-owned trigger bodies on Organization-membership binding deletion (incl. G013 replay) and on denial/revocation; lifted only by explicit authority inside the admission transaction. **No seven-day lapse**: the cooldown is removed. Flagged as a genuinely new durable authority record → owner/Codex decision, not selected silently |
| 2 Applicant locking and privacy | §1.4 (lock order), §4.3, §4.4 (cw `organizations_claim_select/_lock`), §10 D-6 withdrawn | `FOR SHARE` **does** conflict with `FOR NO KEY UPDATE` (that is the serialization). The facade takes **no** Organization lock and the broad `organizations_enrollment_reader` is withdrawn; the lock-read is one `PERFORM … FOR SHARE` inside the cw body under a `users_claim_lock`-shaped helper policy (`USING (true) WITH CHECK (false)`) with a cw column grant limited to `(id, enrollment_mode)`. **One executable order** pinned in §1.4 = the body's statement order. SQL-level request/admission checks live in cw bodies and triggers, not only in the facade |
| 3 Controlled applicant audit path | §1.4, §4.3 (`enrollment_request` body), §4.4 (`events_insert_enrollment` TO cw), §4.5 | The applicant's request INSERT moves out of the runtime role into a cw DEFINER body; its capture fires at trigger depth ≥ 1 under a **new, separate** cw policy `events_insert_enrollment` (G019's `events_insert_claim` and the runtime `events_insert` are untouched, 00008:668-671). No GUC is trusted: the body resolves the applicant through the enrollment credential digest. One success event (`membership.granted`); a failed required append aborts the transaction |
| 4 Explicit approval revocation | §1.4 (revocation), §1.5 (`RevokeEnrollmentApproval`), §4.2 (`revoked` status), §4.5, §5.2 | New guarded operation, idempotent, serialized with consumption on the request row lock; `enrollment.revoked` audit + durable `revoked` restriction; consumed approvals are not revocable (membership is G019's `RemoveMember`) |
| 5 Truthful staged/unknown outcomes | §1.3 step 6 (**outcome table**), §4.2 (`admitted_at`, `admitted_binding_id`), §5.1 | Per-stage confirmed-commit / confirmed-abort / unknown-COMMIT table; reconciliation is a guarded read of durable state, never a repeated grant/consumption; `admitted_sign_in_required` only for confirmed membership; `unavailable` for possibly-durable unknowns |
| 6 Identity continuity and pending-credential revocation | §1.3 steps 2 and 5, §4.4 (`enrollment_sessions` revocation), §5.1 | `store.enroll` performs the existing Live-continuity check (`store.go:645-651`) **before** any mapping/enrollment row; a live userA with providerB is refused (`identity_conflict`, nothing created, A untouched). Deferred issuance **is** the existing `signIn` with the stored provider facts (rotation semantics inherited) and rechecks continuity. Logout, logout-all, disablement and supersession revoke the user's enrollment sessions in the same transaction; restrictions refuse at admission |
| 7 Invitation browser intent and concrete cookies | §1.6 (**cookie contract**), §4.1 (`invitation_id` in `login_transactions`), §5.2 (`invitation_mismatch`) | Cookie `__Host-join` / `join`, own serializer/regex (not the 43-char `SECRET`), **`Path=/` only**, 80-char value `<uuid>.<token>`, duplicate → absent, relay unchanged (cookie read at the tenant-origin callback), one-use clearing on every callback outcome, bound to the stored login transaction's `invitation_id` (mismatch → refused, nothing created). Raw token never in a server URL, log or table |
| 8 Implementable contracts | §5.1 (`AddressEvidence` construction contract; cycle-free orchestration), §4.3/§4.4 (`auth_resolve_enrollment`, ar-owned), §1.8 (SDK verification) | Trusted construction through an unexported interface method on `identity.Provider` implementations (`identity.NewAddressEvidence` is package-private; `workos`/`identitytest` receive a sealed constructor via `identity.EvidenceMinter`); access keeps importing identity (no reverse import; identity's enrollment orchestration calls access through a small interface it declares). Credential bootstrap = `app.auth_resolve_enrollment` owned by `<db>_auth_reader`, the `auth_resolve_session` pattern (00010:596-609). SDK projection: **unqualified** for the exchange response type (§1.8) |
| Validation reconciliation | §6.2 (ENR-15..ENR-21 negatives), §6.3 | Retention/purge + no-op replay survival; revoked/old denial; NULL-context applicant audit and minimal disclosure; lock/closure races; three-stage unknown commits; identityA/providerB; logout then deferred issuance; concurrent invitation intents; credential bootstrap negatives. No new framework, default full suite or cross-product |
| Fixture eligibility | §6.4 | Request-mode applicant = a seeded subject asserted nonmember of Organization C at setup; independent first-time recipients `erin` (open, B) and new account-less `grace` (invite-only, A) in `emulator-seed.yaml` (+1 scope pattern) |
| Expired-row sweep | §4.4 (`enrollment_sessions_expire`), §3 | Named implementation: the E5 pattern (`store.go:82` `DELETE FROM app.login_transactions` under `login_transactions_expire`, 00006:472-473, EMPTY shape) copied as `enrollment_sessions_expire`; executed by `Service.Login` alongside the login cleanup. No "maintenance catalog" claim |
| Pins and G018 lessons | §11.2, §11.3, §3 | Recomputed helper/function/policy/catalog/route counts; `APPLICABLE` (`run.sh:194`) gains the four enrollment paths so the disabled-copy variant composes (G018 repair review R2/copy); CI partition rows in `profile-paths.tsv` |
| D-15 | §10 | Both options restated accurately; one coherent recommendation for the owner; not selected |

## 0.1 Gates, provisional dependencies and reconciliation

Unchanged from r0 §0.1–§0.3 except: (a) the D-list gate now includes **D-4r1** (durable restriction ledger) as an explicit
owner/Codex decision alongside D-15; (b) R0′ additionally confirms the cw policy names at 00015:642-657 and the EXECUTE
grants at `:877-881` (G058 adds cw policies on new tables and on `organizations`, and never replaces a 00015 policy);
(c) material drift also includes any change to `store.signIn`'s Live-continuity branch (`store.go:645-676`); (d) r2 adds
to the drift list: `app.sessions.provider_session_id` (00002:57; the logout-scope key), `app.invitations.created_at`
(00015:144) and the Organization lock in `invitations_bound` (00015:485; the invitation chronology argument),
`role_bindings_authority_guard` (00015:945-976; the departure body runs after it), the two-line output contract of
`infra/recovery/sql/replay-restrict.sql` and `recovery.sh:822/894`, and `Provider.Exchange`'s input struct
(`identity.go:24`). R0′ also reconciles the importer pins (`TestProductionImporters`, `access/boundary_test.go:206`).

## 1. Approach

### 1.1 Trust boundaries (r0 table corrected)

| Boundary | Change | Containment |
|---|---|---|
| **Enrollment policy** | `app.organizations.enrollment_mode` NOT NULL DEFAULT `'invite_only'` CHECK ∈ {`invite_only`,`request_access`,`open`}; `UpdateEnrollmentPolicy` under **`organization.manage_policy`** (owner only) | closed at rest; unknown → `invite_only` in Go; the mode is read inside the cw admission/request bodies **after** the Organization row lock (§1.4), never from a page or credential |
| **Ordinary sign-in** | unchanged (`issuanceGuard` `store.go:581-603`) | the enrollment branch is selected only by the stored login transaction's `enrollment_intent` |
| **Enrollment credential** (`app.enrollment_sessions`) | created at callback for an enrollment-intent login the Organization currently permits; 30-minute absolute lifetime; cookie kind `enrollment` (`__Host-enroll`/`enroll`, `Path=/`); header `Authorization: Enrollment <credential>`; resolved by **`app.auth_resolve_enrollment`** (ar-owned, §4.3) before any context exists; revoked by the user's session-revocation paths **at their actual scope** (§1.3 step 5: bound current-session/provider paths touch only rows minted from the same provider session; logout-all and disablement touch all) | not a session: no ordinary reader queries the table; the four own-row operations only; `TenantContext` never constructed; **never exchanged for a session** (§1.3 step 4: admission is followed by the ordinary sign-in) |
| **Post-admission sign-in** (D-17) | after a confirmed T2 the browser starts the **unchanged** ordinary login; `signIn` (`store.go:623`) with `issuanceGuard` (`:581-603`, membership) issues the session | no transaction-sharing issuance primitive, no stored principal subject, no second issuer; the enrollment credential grants nothing after admission except the reconciliation read |
| **Local identity before membership** | `createMapping` only when the mode is `request_access`/`open` or `rls_invitation_pending(id, org)`; **and only after the Live-continuity check passes** (§1.3 step 2) | no mapping on a closed tenant; no mapping for a second principal in a browser that holds a live session |
| **Admission write** | one cw body `app.enrollment_admit(p_hash, p_org, p_kind, p_request)` for `open`, `request` and the post-invitation `lift`; the invitation grant itself is P19-1 | runtime still has no INSERT shape on `role_bindings`; the body writes `ctx_user()`'s own `org_member` binding under `role_bindings_insert_invited` (00015:646-647) |
| **Applicant request write** | one cw body `app.enrollment_request(p_hash, p_org)`; runtime has **no INSERT grant** on `enrollment_requests` | SQL-level mode/eligibility checks and the audit append live in the body and its triggers |
| **Restrictions** | durable `app.enrollment_restrictions` (D-4r1) written by owner-owned trigger bodies under an **owner-only, trigger-bound INSERT policy** (§4.4) and, for replayed departures only, by the G013 superuser replay statement (§1.4; `recorded_via='replay'`); read by the admission body; **authority is versioned** (`seq`/`authority_seq`, chronology for invitations) so a stale approval or invitation cannot lift a newer restriction | audit history is evidence, not authority; retention/purge and replay cannot reopen a restriction; a restriction is never reset by a mode change, time, a new login or a stale authorization |
| **Request decisions / revocation** | `DecideEnrollmentRequest` and `RevokeEnrollmentApproval` under **`organization.manage_members`** (MGR) | approval becomes membership only when the applicant returns and the body rechecks approval, decider authority, mode and host |
| **Public surface** | as r0 (`/internal/v1/enrollment/public`, `web/app/join`) | as r0 |
| **First-time invitee proof** (D-15) | as r0: fresh `AddressEvidence` from the same `Exchange`, in memory only | execution authority at SQL = D-15 (§10) |

### 1.2 Riskiest slice first — as r0 (slice 1 now also proves the restriction ledger, the applicant audit path, revocation and the three-stage outcome table).

### 1.3 Authentication is distinct from admission (R2) — the enrollment-intent login

1. **Start.** `POST /internal/v1/auth/login` gains `intent: "enroll"`, `invitation_id` (UUID) and `request: bool`.
   `Service.Login` stores `enrollment_intent`, `invitation_id`, `request_on_arrival` in `app.login_transactions` (§4.1);
   `return_path` for an enrollment login is `/join/status`. State, PKCE, binding cookie, exact origin, `tenantOrigin`
   resolution in the insert transaction and the 10-minute TTL are G016's unchanged code; the E5 expired-login cleanup
   (`store.go:82`) now also runs `DELETE FROM app.enrollment_sessions` under `enrollment_sessions_expire` (§4.4).
2. **Callback, transaction 1 (`store.enroll`, `identity/enroll.go`)** after `consumeLogin`, host recheck and
   `Provider.Exchange`; when `enrollment_intent` is false → existing `signIn`, unchanged. Otherwise, in order:
   (a) **Live-continuity check first** — the BFF forwards the browser's session cookie as today (`service.go:330-344`,
   `in.Credential` → `live`); `principalSubject(issuer, subject)`; if `live != nil` and (`mapped == ""` or
   `mapped != live.UserID`) → `identity_conflict` (`signInResult{Conflict:true}` semantics, `store.go:645-648`): nothing
   is created, A's session is untouched, the login transaction is consumed; if `live != nil` and `mapped == live.UserID`
   → the subject is already signed in on this host, which implies current membership (`issuanceGuard`) → fall through
   to the existing `signIn` (already_signed_in / session_rotated semantics inherited, `store.go:649-676`);
   (b) invitation context: when `txn.invitation_id` is set, the BFF-forwarded `{invitation_id, invitation_token}` must
   be present and `invitation_id` must equal `txn.invitation_id`, else `invitation_mismatch` (consumed, nothing created);
   (c) host recheck inside the transaction → `host_changed`; (d) `rls_enrollment_mode(org)` and, when set,
   `rls_invitation_pending(id, org)`; neither admits → `enrollment_closed` (403; `access.denied` with NULL Organization,
   D-6 shape; no mapping); (e) `createMapping` if unknown (existing); (f) if `is_organization_member` is already true
   (possible without a live session) → ordinary `signIn` issuance; (g) INSERT `enrollment_sessions` (`credential_hash`,
   `csrf_token`, exact origin, Organization, **`provider_session_id`** = `vi.ProviderSessionID` — the only provider fact
   stored, and only as the logout-scope key of step 5 — `expires_at`), COMMIT. No issuer/subject and no `auth_time`
   are stored: nothing in the row is later exchanged for a session (step 4).
3. **Admission, transaction 2** (same request; no provider call holds a lock): kind **invitation** → `AttestVerifiedAddress`
   (P19-3) → P19-4 `ConsumeOrganizationInvitation`, then in the **same transaction** `enrollment_admit(hash, org,
   'lift', NULL)` (lifts any unlifted restriction only when `is_organization_member` is now true, §4.3); kind **open**
   → `enrollment_admit(hash, org, 'open', NULL)`; kind **request** → `enrollment_request(hash, org)` when
   `request_on_arrival`, else no write (status `pending`). On `admitted`/`already_member` the answer is `200
   {"outcome": …, "next": "sign_in"}` with the enrollment cookie cleared (step 4); otherwise `202 {enrollment:
   {status, credential}}` and never a session.
4. **Issuance = the fresh ordinary sign-in after admission (correction 5; D-17).** There is **no T3 inside G058's
   code**: on `admitted`/`already_member` the status page immediately starts the unchanged ordinary login
   (`POST /internal/v1/auth/login` without `intent`, `return_path` `/`), the provider round trip repeats (silent in
   practice while the provider session is live — not assumed, not required), and `store.signIn` (`store.go:623-700`)
   issues the session under `issuanceGuard` (`:581-603`: host routing + `is_organization_member`, now true). Continuity,
   rotation, the active-status `FOR SHARE` and `issue` are the unchanged, already qualified code; the enrollment row is
   **not** read, locked, consumed or exchanged by `signIn`. **Why this route rather than a transaction-sharing
   primitive:** (a) `signIn` owns its transaction through `inTxAudited` (`:155-178`, `:623`); sharing it would require
   extracting an internal `signInInTx(ctx, tx, p)` and a second audited wrapper — an identity refactor on the
   qualified sign-in path for one caller; (b) issuance from stored facts needs the exchanged principal persisted on the
   enrollment row (`auth_principal_subject` maps issuer/subject → UUID only, 00006:319; no reverse reader exists and
   none is wanted) plus a row lock, revocation recheck and provider-binding recheck through issuance — all new
   security surface; (c) the ordinary sign-in already proves the exact principal (fresh `Exchange`), the current
   provider binding (`identity_provider_binding`), continuity, membership and status, so the only cost is one more
   provider redirect; (d) unknown-commit handling becomes two staged transactions plus the existing
   `identity_unavailable` semantics of `signIn` — nothing new to prove. Rejected alternative recorded as D-17.
   Consequences: `IssueAfterAdmission`, `auth_time`, `consumed_at`, `issued_session_id` and the
   `admitted_sign_in_required` outcome are removed (§4.2, §5.1, §5.2); the admitted enrollment row stays resolvable
   (reconciliation read only) until it expires, is abandoned or is revoked by step 5; a disabled or removed user is
   refused by `signIn` itself (`issue` status check; `issuanceGuard`).
5. **Revocation of pending credentials at the revoking path's actual scope (correction 7; D-18).** Each path adds one
   UPDATE in its own transaction (`initIdentity` already sets the user; rt policy `user_id = U`, §4.4):
   - **bound current-session paths** — `revokeSession` (logout and the provider-logout reference, `store.go:852-869`,
     `:896-916`) and the supersession branch of `signIn` (`:663-665`): `UPDATE app.enrollment_sessions SET revoked_at =
     $now, revoke_reason = $reason WHERE user_id = $1 AND revoked_at IS NULL AND provider_session_id = (SELECT
     s.provider_session_id FROM app.sessions s WHERE s.id = $2 AND s.user_id = $1)` — only the enrollment credentials
     minted from the **same provider session** as the local session being revoked (`app.sessions.provider_session_id`,
     00002:57; the own-row policy supplies the SELECT on `sessions`). Pending credentials of the user's other provider
     sessions (another browser/device) are untouched: a current-session logout never becomes user-wide revocation;
   - **broader paths** — `revokeAll` (logout-all, `:872-894`) and `DisableUser` (`:958-985`): `… WHERE user_id = $1 AND
     revoked_at IS NULL` (all rows), matching their existing all-sessions semantics; `auth_resolve_enrollment` also
     requires `users.status = 'active'`, so disablement refuses even before the UPDATE commits.
   **Serialization with admission:** every UPDATE above takes the row locks of the rows it touches; `enrollment_admit`
   and `enrollment_request` take the own row `FOR UPDATE` at step (2) with `revoked_at IS NULL` in the predicate. Under
   READ COMMITTED the loser re-evaluates the row (EvalPlanQual): a revocation that lands first makes the body answer
   `enrollment_expired`; an admission that lands first still gets revoked afterwards (the credential is spent; the
   confirmed `admitted_at` is untouched and the sign-in of step 4 does not depend on the credential). T3 (the ordinary
   sign-in) touches no enrollment row, so no serialization with it is needed; its own serialization (users `FOR SHARE`
   vs `DisableUser` `FOR UPDATE`; `issuanceGuard` vs `RemoveMember`) is the existing one. A restriction does not revoke
   the credential (the status page must be able to say `denied`/`enrollment_restricted`); it refuses at admission.
6. **Per-stage outcomes (correction 5, r2).**

| Stage | Confirmed COMMIT | Confirmed abort | Unknown (COMMIT in flight; error without a definite answer) |
|---|---|---|---|
| T1 mapping + enrollment session | continue to T2 | closed code or `unavailable`; nothing created | answer `unavailable` (503); the login transaction is consumed either way; a fresh enrollment login creates a new enrollment session (mapping is keyed by `(issuer, subject)`; the stale enrollment session, if committed, expires, is abandoned or is revoked by step 5 — never reused) |
| T2 admission (binding + consumption + lift + audit) | `admitted_at`/`admitted_binding_id` durable on the enrollment session; answer `admitted` + `next: sign_in`; the enrollment cookie is cleared | outcome row (`pending`, `denied`, `enrollment_closed`, `enrollment_restricted`, …) or `unavailable`; zero writes (a `lift` RAISE rolls back the claim too, §1.4) | answer `unavailable`; **no repeat of the grant**: the next `status`/`admit` call is reconciliation by **read** (`auth_resolve_enrollment` → `admitted_at`, `is_organization_member`) and reports `admitted` only when the row proves it; a second `admit` enters the body, which answers `already_member` without a write |
| T3 ordinary sign-in (unchanged identity code) | session cookie (`signed_in`/`session_rotated`) | existing codes (`organization_not_member` if membership was removed meanwhile, `user_disabled`, `host_changed`, `identity_conflict`) — the browser is told to sign in again; membership stands as T2 committed it | existing `identity_unavailable`: nothing is claimed; the possibly-issued session idles out; the browser retries the ordinary login (`signIn` is idempotent per provider session: `already_signed_in`/`session_rotated`) |

### 1.4 One membership transaction and truthful retries (R3)

**Admission kinds and their single write.**

| Kind | Precondition (rechecked inside the body) | Write |
|---|---|---|
| invitation | P19-1 preconditions (pending Organization invitation, principal → `ctx_user()`, byte-equal address, `intended_user_id`) | G019's claim; then `enrollment_admit(…,'lift',inv_id)` lifts the unlifted restrictions **older than the invitation** (`lifted_via='invitation'`) **in the same transaction**; any unlifted restriction **not older** makes the body RAISE `enrollment_restricted`, which rolls the claim back (invitation stays pending) |
| open | mode = `open`; subject active; **no unlifted row in `enrollment_restrictions`** for `(org, user)`; no current binding | `enrollment_admit(…,'open',NULL)`: INSERT `org_member` (+ `membership.granted`) |
| approved request | request `approved`, unconsumed, unexpired, for this subject/Organization; decider currently `org_owner`/`org_admin` (`source='membership'`); mode ∈ {`request_access`,`open`}; subject active; no current binding; **no unlifted restriction with `seq > request.authority_seq`** | `enrollment_admit(…,'request',id)`: INSERT binding + request consumption + lift of the rows with `seq <= authority_seq` (`lifted_via='approved_request'`) — one function, one transaction; a newer restriction → `enrollment_restricted`, zero writes |

**Executable lock order (one; equals the body's statement order, §4.3):**
(1) `app.users` self row `FOR SHARE` via `rls_user_enabled_lock(ctx_user(),'organization',p_org)` (00015:692-694) →
(2) `app.enrollment_sessions` own row `FOR UPDATE` (by `credential_hash = p_hash`) →
(3) `app.organizations` `FOR SHARE` (`PERFORM 1 … WHERE id = p_org FOR SHARE` under `organizations_claim_lock`) →
(4) `app.enrollment_requests` row `FOR UPDATE` (kind `request` only) →
(5) `app.enrollment_restrictions` rows `FOR UPDATE` (lift) or plain read (open check) →
(6) `app.role_bindings` INSERT → (7) `enrollment_requests` consumption UPDATE → (8) `enrollment_sessions` `admitted_*` UPDATE.
The facade (`access.Admit`) runs IDENTITY shape, resolves the host inside the transaction (`host_changed`) and calls the
body **once**; it takes no row lock itself. **Serialization facts (corrected):** `FOR SHARE` conflicts with `FOR NO KEY
UPDATE` and `FOR UPDATE`; the owner's guarded mode UPDATE (an UPDATE of a non-key column acquires `FOR NO KEY UPDATE`;
the facade's `ExclusiveOrganization` is `FOR NO KEY UPDATE`, `tenancy/store.go:83`) therefore **waits** for an in-flight
admission and vice versa, which is the intended serialization: the mode read after the lock sees the committed mode.
Two concurrent admissions both take `FOR SHARE` (compatible) and serialize on the enrollment-session row (2) and the
`role_bindings_subject_scope_key` unique index. A concurrent `RemoveMember`/`DisableUser`: the disable commits before
or after (1); a removal after (6) runs G019's delete path and the departure trigger records a restriction.
**Decision/revocation order (r2, r3):** `DecideEnrollmentRequest`/`RevokeEnrollmentApproval` take (3′) the facade's
`ExclusiveOrganization` (`FOR NO KEY UPDATE`, `tenancy/store.go:83`) → (4) the request row (UPDATE lock); on approval
the BEFORE UPDATE guard re-takes (3′) itself and **derives** `authority_seq` under it (§4.5; r3 correction 2), so the
watermark is authoritative in SQL whether or not the facade locked first. **Departure body order (r3, correction 1):**
G019's `role_bindings_authority_guard` (BEFORE, 00015:976) runs first; the AFTER DELETE body then takes (3″) through
**`app.enrollment_lock_organization(OLD.organization_id)`** (§4.3: a cw-owned DEFINER helper that locks `FOR NO KEY
UPDATE` under the lock-only policy `organizations_claim_lock USING (true)`), and RAISEs if it returns false. r2 took
this lock through the owner-shared tenant policies (`organizations_update USING (id = ctx_org() AND rls_org_role() IS
NOT NULL)`, 00006:434), which fails exactly on a **self-leave of the last Organization membership**: after the DELETE
`rls_org_role()` is NULL, the row is invisible, and `LeaveScope` would roll back. The helper's visibility does not
depend on the caller's remaining membership, reads no column, is callable by `<db>_owner` only, and is taken the same
way under the G013 superuser replay (no `ctx_org()` conditional remains). Then INSERT the ledger row with `created_at =
clock_timestamp()`. Every path
therefore orders Organization before request/ledger rows; the only exclusive Organization holders (decisions,
departures, G019's invitation insert `invitations_bound` 00015:485 and last-owner guard :961) serialize with each other
and with admissions' `FOR SHARE`; no path takes the request row before the Organization.

**Outcomes (closed set, r2):** `admitted` (always with `next: sign_in`), `already_member`, `pending`, `denied`,
`enrollment_closed`, `enrollment_restricted` (open kind with an unlifted row; request/lift kinds with a restriction
newer than the authority), `approval_revoked` (revoked, expired, decider authority lost), `already_consumed`,
`host_changed`, `recipient_ineligible`, `recipient_proof_unavailable`, `invitation_mismatch`, `identity_conflict`,
`not_found`, `enrollment_expired`, `unavailable`. `admitted_sign_in_required` is **removed** (D-17).

**Duplicate / lost response:** as r0, with reconciliation by read only (§1.3 step 6). `RequestEnrollment` is idempotent
per `(organization_id, user_id)` while pending (`enrollment_requests_one_pending`; the body returns the existing row).

**Restrictions and reinstatement (AC5; correction 1).** A row in `app.enrollment_restrictions (organization_id,
user_id, kind ∈ {departed, denied, revoked}, created_at, created_by_user_id NULL, lifted_at NULL, lifted_by_user_id
NULL, lifted_via NULL)` with `lifted_at IS NULL` refuses the **open** kind (`enrollment_restricted`). Rows are written
only by owner-owned trigger bodies (§4.5): on DELETE of an Organization-scope `source='membership'` binding (G019
`RemoveMember`, `LeaveScope`, ownership paths and the superuser G013 `replay-restrict.sql` `membership.remove` delete,
`infra/recovery/sql/replay-restrict.sql:8-9`) → `departed`; on a request `denied` → `denied`; on an approval `revoked`
→ `revoked`. They are lifted only inside `enrollment_admit` by explicit authority (invitation claim or approved
request) **that is not older than the restriction** (below); **mode changes, time and new logins never lift them** (no
cooldown). Durability: the table lives in `app` (not retention-eligible; restored with the data), so an audit purge
(`events_retention_delete`, 00009:297) cannot reopen membership. Disabled accounts fail `recipient_ineligible`. Audit
remains evidence (`membership.removed`, `enrollment.decided/revoked`), not authority. **This ledger is a new durable
authority record → D-4r1 escalation (§10).**

**Ledger insertion under FORCE RLS (correction 2).** The two DEFINER trigger bodies are owner-owned, and the owner is
bound by FORCE RLS (G019's legacy preflight has to lift it, 00015:108-127), so "DEFINER" alone inserts nothing. r2
pins the exact owner path: policy **`enrollment_restrictions_owner_insert`** (`FOR INSERT TO ow`, §4.4) whose WITH
CHECK requires `pg_trigger_depth() >= 1`, the lift columns NULL, `kind IN ('departed','denied','revoked')` and
`recorded_via = 'trigger'`; plus body guards (`TG_WHEN = 'AFTER'`, `TG_TABLE_SCHEMA = 'app'`, `TG_TABLE_NAME`, `TG_OP`,
values copied only from `OLD`/`NEW`). There is **no runtime, cw, pr or ar INSERT grant**; an owner INSERT at depth 0 is
a policy violation; the runtime never connects as owner. **Atomicity:** the INSERT runs inside the deleting/deciding
statement; a refused insert (policy, CHECK, FK, injected fault) raises inside that statement, so `RemoveMember`,
`LeaveScope`, the G013 replay delete, `DecideEnrollmentRequest(denied)` and `RevokeEnrollmentApproval` all roll back
as a whole — no binding is removed and no request is decided without its ledger row (ENR-22 proves each real path
and the rollback by dropping the owner policy inside the test transaction).

**Replay, no-op replay and legacy state (correction 3).** The G013 replay (`infra/recovery/sql/replay-restrict.sql`,
`membership.remove` CTE `r`, `:8-12`; driven by `recovery.sh:889-900`, which parses keyed output lines with `kv()`,
`:822`) deletes the binding as superuser (RLS bypassed; the AFTER DELETE trigger still runs and the owner body inserts
under the owner policy, depth ≥ 1). Three restored states exist for a replayed removal: (a) the checkpoint holds the
binding → the delete fires the trigger → `departed` row (`recorded_via='trigger'`); (b) the checkpoint holds the
restriction row (taken after the removal) → `updated 0`, state `none` → already applied, the row stays; (c) **the
checkpoint predates both the grant and the removal** → no binding, no restriction, no trigger — r1's claim that the
row "stays" was false here. r2 adds **bounded materialization** as one guarded statement after the CTE in the same
transaction: `INSERT INTO app.enrollment_restrictions (organization_id, user_id, kind, recorded_via) SELECT :'org',
:'subject', 'departed', 'replay' WHERE :'kind' = 'membership.remove' AND :'scope_type' = 'organization' AND EXISTS
(organization) AND EXISTS (user) AND NOT EXISTS (current binding) AND NOT EXISTS (unlifted departed row)` — it
runs only when the delete left no binding (state `none` or deleted), never for Group scope, never for `absent`
references (those remain conflicts, `recovery.sh:865`), never duplicates the trigger's row, and is wrapped in a
`to_regclass('app.enrollment_restrictions') IS NOT NULL` guard (psql variables passed through `set_config` to a `DO`
block, because psql does not interpolate inside dollar quotes) so checkpoints taken before 00016 still replay; a third
output line `ledger<TAB>materialized|present|absent|unavailable` is emitted and tolerated by `kv()` (first-match lookup
by key). `recovery.sh` is **not** modified. The materialized row carries `created_at = clock_timestamp()` at replay
time, so every approval/invitation restored from the checkpoint is older than it (reinstatement rule below): "replayed
restrictions deny their old authority" (V8) extends to enrollment. **Not inferred:** audit history, legacy
`membership.removed` events or replay entries of other kinds never create rows. **Legacy state:** memberships removed
**before 00016** have no ledger row (requests, denials and revocations do not exist before G058, so only departures
are legacy). r2 does not backfill them (that would be audit-derived authority, withdrawn in r1) and does not reset
anything; the choice is **escalated as D-16** with these **readiness assumptions**: (1) the ledger starts empty at
00016; (2) every Organization is `invite_only` at 00016 (column default; the §4.1 precondition `DO` asserts it), so
no open-signup exposure exists until an owner opts in after release; (3) checkpoints taken after 00016 carry the
ledger table, and a pre-00016 checkpoint replays without materialization (recorded `ledger absent`) — the operator
runbook (`docs/foundation/RECOVERY.md`, in scope) states that such a recovery leaves pre-G058 departures unrestricted
until an owner re-removes or the D-16 decision says otherwise; (4) `recovery.sh` does not migrate the restored
checkpoint before replay (no `migrate` call in `recovery.sh`), which is why the guard, not a migration order, carries
assumption (3).

**Reinstatement authority (correction 4).** Explicit authorization must be valid for the **current** restriction
state. Two mechanisms, no G019 row modified, no restriction reset:
- **Approvals — authority version (r3: derived in SQL).** `enrollment_restrictions.seq` is an identity column
  (monotonic, unique). `DecideEnrollmentRequest(approved)` sets only `status`/`decided_*`; the BEFORE UPDATE guard
  `enrollment_requests_immutable` (§4.5) takes the Organization `FOR NO KEY UPDATE` and assigns
  `NEW.authority_seq := coalesce((SELECT max(seq) FROM app.enrollment_restrictions WHERE organization_id =
  NEW.organization_id AND user_id = NEW.user_id), 0)` itself, ignoring any supplied value; rt has **no column
  privilege** on `authority_seq`, so a direct-SQL inflated watermark is refused (42501) before the trigger runs, and
  any other transition that changes the column raises. The facade computes nothing. At admission (kind `request`, step
  6′), any unlifted row with `seq > authority_seq` → `enrollment_restricted`
  (zero writes; the request stays `approved` and unusable until the admin re-decides a **new** request or revokes
  this one); the lift of step (9) updates only rows with `seq <= authority_seq`. Visibility makes this exact: a
  restriction committed after the approval read its watermark has a higher `seq` or was not yet visible — either way
  it is "newer"; one that was visible before the decision is covered by the watermark. No clock is involved.
- **Invitations — chronology under the shared Organization lock.** G019's rows carry `created_at = now()` (00015:144)
  and the insert holds the Organization `FOR NO KEY UPDATE` until commit (`invitations_bound`, :485). The departure
  body stamps `created_at = clock_timestamp()` while holding the same lock (3″) until commit; decision/revocation rows
  are stamped the same way under (3′). `lift` (kind `lift`, `p_request` = invitation id) lifts rows with `created_at <
  inv.created_at` and **RAISEs `enrollment_restricted`** (`P0001`, message `enrollment_restricted`) when any unlifted
  row has `created_at >= inv.created_at`, so the G019 claim already executed in the same transaction rolls back and
  the invitation stays pending. Because both writers hold the Organization lock from stamp/insert to commit, a
  restriction that commits after the invitation commits is always stamped after the invitation's `now()`: **no stale
  invitation can lift a newer restriction**. The converse inversion (an invitation whose transaction started before a
  departure committed but inserted after it) is refused although the invitation is logically newer — fail-closed; the
  admin re-invites (the new row's `now()` is later). Under replay, materialized rows are newer than everything restored.
- **Concurrency (ENR-10 O-ENR-6..8):** approval vs removal in both orders (removal first → watermark covers it → lifts;
  approval first → `seq` newer → refused); two tabs consuming one approval (request row lock → `already_member` for the
  loser, one binding); removal racing an in-flight admission (Organization lock → removal waits → deletes the fresh
  binding → `departed` row newer than the consumed approval); invitation claim vs departure in both orders (claim
  first → the departure restricts the new membership; departure first → the older invitation's `lift` raises, claim
  rolled back, invitation still pending, re-invite lifts).

**Approval revocation (correction 4).** `RevokeEnrollmentApproval(tc, id)` (MGR): `UPDATE enrollment_requests SET
status='revoked', revoked_at, revoked_by_user_id WHERE id=$1 AND organization_id=O AND status='approved' AND consumed_at
IS NULL`; 0 rows → re-read: `revoked` → `200` unchanged (idempotent, no second event); `consumed` → `409
already_consumed` (membership stands; removal is G019's `RemoveMember`); `pending`/`denied` → `409 not_approved`. The
UPDATE takes the row lock that `enrollment_admit` step (4) also takes, so revocation and consumption serialize: the
loser sees `approval_revoked` or `already_consumed`. Effects: `enrollment.revoked` audit (actor = revoker) and a
`revoked` restriction row (trigger); a later approval of a **new** request lifts it on admission.

### 1.5 Finite modes, authority and operations (R1)

| Operation | Authority | Tx | Writes |
|---|---|---|---|
| `PublicEnrollment(ctx, origin)` | none | RR RO | — |
| `GetEnrollmentPolicy(tc)` | `organization.read` | RR RO | — |
| `UpdateEnrollmentPolicy(tc, mode)` | **`organization.manage_policy`** | RC RW | UPDATE `enrollment_mode` + `enrollment.policy_changed` |
| `ListEnrollmentRequests(tc, after, limit, status)` | **`organization.manage_members`** | RR RO | — |
| `DecideEnrollmentRequest(tc, id, decision)` | `organization.manage_members` | RC RW | `ExclusiveOrganization` (§1.4 (3′)); UPDATE `pending → approved|denied` (r3: the BEFORE guard derives `authority_seq` on `approved`; the facade never sets it) + `enrollment.decided` (+ `denied` restriction) |
| **`RevokeEnrollmentApproval(tc, id)`** | `organization.manage_members` | RC RW | UPDATE `approved → revoked` + `enrollment.revoked` (+ `revoked` restriction) |
| `EnrollmentStatus(ctx, cred)` | own credential | RR RO | — (reconciliation read) |
| `RequestEnrollment(ctx, cred)` | own credential | RC RW | `enrollment_request` body: INSERT + `enrollment.requested` |
| `Admit(ctx, cred, kind, requestID)` | own credential | RC RW | `enrollment_admit` body |
| `AbandonEnrollment(ctx, cred)` | own credential | RC RW | UPDATE `revoked_at` (own row) |

Capabilities unchanged (19 after G019). Group roles hold neither capability.

### 1.6 Web and BFF (R5, AC7) — r0 routes retained; the cookie contract replaces r0 D-3

Routes and files as r0 §1.6 (nine web files; the admin page gains a **Revoke** action on approved rows; `WEB_ROUTES`
39 → 48). **Cookie contract (correction 7):**

| Cookie kind | Name (`https` / `loopback-http`) | Value | Attributes | Lifetime | Set / cleared |
|---|---|---|---|---|---|
| `login` (existing) | `__Host-login` / `login` | 43-char secret | `Path=/; HttpOnly; SameSite=Lax; Secure(https)` | ≤ 10 min | unchanged |
| **`enrollment`** | `__Host-enroll` / `enroll` | 43-char secret (`SECRET` regex) | same | ≤ 1800 s | set by the callback on `202`; cleared on `admitted`/`already_member` (the browser then starts the ordinary login, r2), `abandon` and any 401 from the status route |
| **`join`** (invitation intent) | `__Host-join` / `join` | `<invitation uuid 36>.<token 43>` = 80 chars, regex `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\.[A-Za-z0-9_-]{43}$` via a **separate** `setJoinCookie`/`readJoinCookie` pair (the existing `setCookie` keeps refusing anything but `SECRET`) | `Path=/` (the only legal path for a `__Host-` cookie; no `Domain`), `HttpOnly; SameSite=Lax; Secure(https)` | 600 s | set by `POST /join/start` in the same 303 as `__Host-login`; **cleared by `/auth/callback` on every outcome** (one use) |

Rules: a duplicate or malformed `join` cookie reads as **absent** (`readCookie`'s existing repeated-value rule,
`session.ts:76-80`); a login transaction that carries `invitation_id` with an absent or different cookie id is refused
`invitation_mismatch` (consumed; nothing created) — this binds the invitation intent to the stored transaction and the
browser flow; a concurrent start in another tab overwrites `__Host-join` and the earlier tab's callback is refused the
same way. The custom-domain relay is unchanged: the token never travels in the relay URL; the cookie is read at the
tenant-origin `/auth/callback` after the relay hop. The token appears only in the fragment, the one-use cookie and the
callback JSON body; never in a `Location`, server URL, log line or table (G019 stores the digest only). Sizes: `join`
value 80 bytes; JSON body budget 8 KiB.

### 1.7 Configuration and bounds — as r0 minus `EnrollmentDenialCooldown` (removed; no automatic lapse). Constants: `EnrollmentSessionLifetime = 30m`, `JoinCookieLifetime = 10m`, `EnrollmentRequestLifetime = 30d`, `MaxEnrollmentRequestsPage = 100`, limiter as r0.

### 1.8 Live qualification delta and SDK projection (correction 8)

Read-only in-repo evidence at `044cfd52`: the adapter uses `resp.User.{ID, FirstName, LastName}` of
`sdk.AuthenticateResponse` (`core/internal/identity/workos/adapter.go:111-136`, module `github.com/workos/workos-go/v10
v10.5.0`, `core/go.mod:15`); the staging test constructs `UserManagementCreateParams{…, EmailVerified: &verified}`
(`workos/staging_test.go:143-146`), proving the pinned SDK models `EmailVerified` on **user creation parameters**.
Whether the **authenticate response's `User`** carries `Email`/`EmailVerified` was **not** verified in this revision: the
one command that would have read the module cache was refused by the harness (§12) and was not retried. The projection
is therefore **honestly unqualified**; the fake's `email_verified` field is fixture behavior, not provider proof. G057
and attempt 1 (read-only grep of the module cache or vendored copy, recorded with file:line) must settle it before any
first-time path is exercised; if absent, D-15 (i)/(ii) are both unimplementable without a provider probe → return to
Codex. Nothing here authorizes WorkOS/DNS/TLS/deployment changes.

## 2. Source binding (provisional; rebind at R0′)

| Input | Value |
|---|---|
| Target | `origin/staging` `044cfd526daecd2ea035e4f03267cc439d2b8f5a` |
| Provisional predecessor | `origin/g019-scoped-invitations-membership-ownership` (draft MR !43; interfaces read at the branch head during this revision) |
| Migration name | `core/migrations/00016_organization_enrollment.sql` — **provisional** |
| Execution | after D-15 and D-4r1 dispositions, Codex Plan ACCEPT, scope import, initial Result, Git-first binding, G019 reviewed integration, resource readiness and recorded release |

## 3. Files (touch-list; r0 §3 with these deltas)

**Create** as r0, plus `core/internal/identity/enrollment_evidence.go` folded into `identity.go` (no new file).
**Modify (core production):** as r0 plus `core/internal/identity/store.go` revocation paths at their actual scope
(`revokeSession` + supersession: bound by `provider_session_id`; `revokeAll`, `DisableUser`: all rows; §1.3 step 5), the
E5 cleanup, and `identity.go` (`Exchange` input gains `Minter`; `ExchangeResult`; `Admitter`/`AdmitSelector`/
`AdmitOutcome`/`EnrollmentHandle` types). **No `IssueAfterAdmission` and no `signIn` refactor** (D-17). **Modify (web
production):** as r0; `web/lib/session.ts` gains `CookieKind` `enrollment` and `join`, `setJoinCookie`/`readJoinCookie`;
`web/app/join/status/page.tsx` starts the ordinary login on `admitted` (existing sign-in start; no new route).
**Modify (recovery, r2, +1 scope):** `infra/recovery/sql/replay-restrict.sql` — one guarded materialization statement
and one keyed output line after the existing CTE (§1.4); `recovery.sh`, `lib.sh` and the other replay scripts are
**not** touched. **Modify (fixtures):** `infra/identity/mcp-seed.sql`
(modes; a request-mode applicant asserted nonmember of C) and **`infra/identity/emulator-seed.yaml` (+1 account-less
user `grace`)**. **Modify (harness):** as r0 plus `infra/validation/run.sh` `APPLICABLE` (+4: `core/internal/enrollment`,
`web/lib/enrollment-client.ts`, `web/app/join`, `web/app/organization/enrollment`; G018 repair review R2/copy lesson:
server `main.go` and `web/components` import them, so the disabled-copy variant must carry them) and
`infra/ci/profile-paths.tsv` (+3 partition rows incl. `core/internal/enrollment/`). **Not touched:** as r0 (no
`infra/postgres/**`, no new role; `<db>_claim_writer`, `<db>_policy_reader`, `<db>_auth_reader` reused).

## 4. Schema and migration `00016_organization_enrollment.sql` (flagged for explicit review: §4.3 DEFINER bodies, §4.4 policies/grants, §4.5 triggers)

### 4.1 Precondition `DO` and columns — as r0, plus `login_transactions` gains `request_on_arrival boolean NOT NULL DEFAULT false`; the three helper roles are preconditions (00015:46-50 pattern).

### 4.2 Tables

```
app.enrollment_sessions (13 columns, r2): id uuid PK, user_id uuid NOT NULL FK users RESTRICT, organization_id uuid NOT NULL FK RESTRICT,
  origin text NOT NULL (≤ 300), credential_hash bytea NOT NULL UNIQUE (32), csrf_token bytea NOT NULL,
  provider_session_id text NOT NULL CHECK (char_length BETWEEN 1 AND 255)   -- logout-scope key only (§1.3 step 5; 00002:57 shape)
  created_at DEFAULT now(), expires_at NOT NULL,
  admitted_at timestamptz NULL, admitted_binding_id uuid NULL,            -- T2 durable proof (correction 5)
  revoked_at timestamptz NULL, revoke_reason text NULL
  -- removed in r2 (D-17): auth_time, consumed_at, issued_session_id — the credential is never exchanged for a session
  CHECK lifetime: expires_at > created_at AND expires_at <= created_at + interval '30 minutes'
  CHECK admission_shape: (admitted_at IS NULL) = (admitted_binding_id IS NULL)
  CHECK revocation_shape: (revoked_at IS NULL) = (revoke_reason IS NULL)
  INDEX (user_id, organization_id) WHERE revoked_at IS NULL; INDEX (user_id, provider_session_id) WHERE revoked_at IS NULL

app.enrollment_requests: id uuid PK, organization_id, user_id, status text NOT NULL CHECK IN ('pending','approved','denied','revoked'),
  created_at, expires_at, decided_at NULL, decided_by_user_id NULL FK, revoked_at NULL, revoked_by_user_id NULL FK,
  consumed_at NULL, consumed_binding_id NULL,
  authority_seq bigint NOT NULL DEFAULT 0            -- r2 (correction 4): greatest committed restriction seq covered by the approval
  CHECK decision_shape: (status = 'pending') = (decided_at IS NULL) AND (decided_at IS NULL) = (decided_by_user_id IS NULL)
  CHECK revocation_shape: (status = 'revoked') = (revoked_at IS NOT NULL) AND (revoked_at IS NULL) = (revoked_by_user_id IS NULL)
  CHECK consumption_shape: (consumed_at IS NULL) = (consumed_binding_id IS NULL) AND (consumed_at IS NULL OR status = 'approved')
  CHECK authority_shape: (authority_seq = 0 OR status IN ('approved','revoked'))
  UNIQUE INDEX one_pending (organization_id, user_id) WHERE status = 'pending'; scope and user indexes as r0

app.enrollment_restrictions: id uuid PK DEFAULT gen_random_uuid(),
  seq bigint NOT NULL GENERATED ALWAYS AS IDENTITY UNIQUE,                 -- r2: authority version (correction 4)
  organization_id uuid NOT NULL FK RESTRICT, user_id uuid NOT NULL FK users RESTRICT,
  kind text NOT NULL CHECK IN ('departed','denied','revoked'),
  recorded_via text NOT NULL CHECK IN ('trigger','replay'),                -- r2: 'replay' only from replay-restrict.sql (correction 3)
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),               -- r2: stamped under the Organization lock (§1.4)
  created_by_user_id uuid NULL, source_request_id uuid NULL,
  lifted_at timestamptz NULL, lifted_by_user_id uuid NULL, lifted_via text NULL CHECK IN ('invitation','approved_request')
  CHECK lift_shape: (lifted_at IS NULL) = (lifted_via IS NULL)
  CHECK replay_shape: (recorded_via = 'trigger' OR (kind = 'departed' AND source_request_id IS NULL AND created_by_user_id IS NULL))
  INDEX enrollment_restrictions_open_idx (organization_id, user_id) WHERE lifted_at IS NULL
```
All three `FORCE ROW LEVEL SECURITY`. No raw credential, address, token, principal subject or provider payload is
stored (`provider_session_id` is the non-bearer reference the sessions table already stores).

### 4.3 SECURITY DEFINER functions (allowance for 00016 = **8** (r3: +1 cw lock helper); every one `SET search_path = pg_catalog, pg_temp`, PUBLIC revoked; owner/volatility/body hash pinned in CT-ENR-1)

- `app.rls_enrollment_mode(p_org uuid) RETURNS text` STABLE (**pr**): as r0 (public mode; absent row → `'invite_only'`).
- `app.rls_invitation_pending(p_id uuid, p_org uuid) RETURNS boolean` STABLE (**pr**): as r0 (D-5).
- **`app.auth_resolve_enrollment(p_hash bytea, p_at timestamptz)`** STABLE (**ar**; EXECUTE runtime/owner) — the
  `auth_resolve_session` pattern (00010:596-609): `RETURNS TABLE (session_id, user_id, organization_id, origin,
  csrf_token, expires_at, admitted_at, admitted_binding_id)` (r2: no provider facts are returned — nothing is issued
  from them) from
  `app.enrollment_sessions es JOIN app.users u` WHERE `es.credential_hash = p_hash AND es.revoked_at IS NULL AND
  es.expires_at > p_at AND u.status = 'active'`. This is the **only** way the runtime finds an enrollment row before
  `ctx_user()` exists; runtime self-row policies never serve the bootstrap.
- **`app.enrollment_request(p_hash bytea, p_org uuid) RETURNS TABLE (outcome text, request_id uuid)`** VOLATILE (**cw**;
  EXECUTE runtime/owner — the same trust shape as `invitation_claim_group` and `rls_user_enabled_lock`, 00015:879-880):
  requires `ctx_user()` non-NULL and IDENTITY shape; (1) `rls_user_enabled_lock(ctx_user(),'organization',p_org)` →
  `recipient_ineligible`; (2) `SELECT … FROM app.enrollment_sessions WHERE credential_hash = p_hash AND user_id =
  ctx_user() AND organization_id = p_org AND revoked_at IS NULL AND expires_at > now() FOR UPDATE` → `enrollment_expired`;
  (3) `PERFORM 1 FROM app.organizations WHERE id = p_org FOR SHARE`; (4) mode read → `enrollment_closed` unless
  `request_access`; (5) existing binding → `already_member`; (6) existing pending → `pending` (existing id, no write);
  (7) INSERT `enrollment_requests (pending)` under `enrollment_requests_claim_insert` → the capture trigger appends
  `enrollment.requested` under `events_insert_enrollment` (depth ≥ 1). A prior `denied`/`revoked` restriction does
  **not** refuse a new request (an admin re-decides; open signup stays refused).
- **`app.enrollment_admit(p_hash bytea, p_org uuid, p_kind text, p_request uuid) RETURNS TABLE (outcome text, binding_id
  uuid)`** VOLATILE (**cw**; EXECUTE runtime/owner): statements in the §1.4 order, each a separate command: (1) user
  lock/status → `recipient_ineligible`; (2) enrollment session own-row `FOR UPDATE` by digest → `enrollment_expired`;
  if `admitted_at IS NOT NULL` → `already_member` (reconciliation, no write); (3) Organization `FOR SHARE`; (4) mode
  read; `open` requires `mode='open'`, `request` requires `mode IN ('request_access','open')`, `lift` requires
  nothing; (5) `request`: request row `FOR UPDATE` → `not_found` unless own/Organization match; `approval_revoked`
  unless `status='approved' AND consumed_at IS NULL AND expires_at > now()` and the decider currently holds
  `org_owner`/`org_admin` (`source='membership'`); `lift`: `p_request` is the claimed invitation id — `SELECT
  created_at FROM app.invitations WHERE id = p_request AND organization_id = p_org AND consumed_by_user_id = ctx_user()`
  (cw `invitations_claim_select`, 00015:642; the claim in the same transaction already set the consumption pair) →
  `not_found` otherwise; (6) `open`: unlifted restriction exists → `enrollment_restricted`; **(6′ r2, correction 4)**
  `request`: `EXISTS (unlifted row WITH seq > request.authority_seq)` → `enrollment_restricted` (zero writes);
  `lift`: `EXISTS (unlifted row WITH created_at >= inv.created_at)` → **`RAISE EXCEPTION 'enrollment_restricted' USING
  ERRCODE = 'P0001'`** (the claim executed earlier in this transaction must roll back; access maps `P0001` +
  this message to the outcome); (7) current binding exists → `already_member` (for `lift`: proceed to (9) only if it
  exists, else `not_found`); (8) `open`/`request`: INSERT `org_member` under `role_bindings_insert_invited`; `request`:
  consumption UPDATE under `enrollment_requests_claim_update`; (9) `UPDATE app.enrollment_restrictions SET
  lifted_at=now(), lifted_by_user_id=<decider | NULL for invitations>, lifted_via=… WHERE organization_id=p_org AND
  user_id=ctx_user() AND lifted_at IS NULL AND (`request`: seq <= authority_seq | `lift`: created_at < inv.created_at)`;
  (10) `UPDATE enrollment_sessions SET admitted_at=now(), admitted_binding_id=…`. Every outcome other than the `lift`
  RAISE is a returned row; the facade commits only on `admitted`/`already_member`/`pending` rows, and rolls back on
  every refusal row even though those paths wrote nothing.
- **`app.enrollment_lock_organization(p_org uuid) RETURNS boolean`** VOLATILE STRICT (**cw**; r3 correction 1; EXECUTE
  revoked from PUBLIC and granted to **`<db>_owner` only** — the `invitation_claim_organization` grant shape,
  00015:881; never to the runtime): body `PERFORM 1 FROM app.organizations o WHERE o.id = p_org FOR NO KEY UPDATE;
  RETURN FOUND;`. It locks through `organizations_claim_lock` (`USING (true) WITH CHECK (false)`) with the cw
  `UPDATE (enrollment_mode)` privilege, so it is valid after the caller's own membership binding is deleted and under
  the G013 superuser replay alike; it reads no column and returns no row data (no tenant read widened). Called only by
  the two trigger bodies below.
- **Trigger bodies (owner-owned SECURITY DEFINER, EXECUTE revoked from PUBLIC; r2 guards, correction 2):** each body
  first asserts `pg_trigger_depth() >= 1 AND TG_WHEN = 'AFTER' AND TG_LEVEL = 'ROW' AND TG_TABLE_SCHEMA = 'app'` plus
  its own `TG_TABLE_NAME`/`TG_OP`, raising `insufficient_privilege` otherwise (a direct `SELECT app.enrollment_restrict_
  on_departure()` has depth 0 and no TG_* context → refused before any INSERT). `app.enrollment_restrict_on_departure()`
  (AFTER DELETE ON `role_bindings` FOR EACH ROW WHEN `OLD.scope_type='organization' AND OLD.source='membership'`):
  takes the Organization lock (3″, §1.4) as `IF NOT app.enrollment_lock_organization(OLD.organization_id) THEN RAISE
  … 'organization_missing' END IF` (r3; unreachable while the FK holds), then `INSERT (organization_id, user_id, kind,
  recorded_via, created_by_user_id) VALUES (OLD.organization_id, OLD.subject_user_id, 'departed', 'trigger',
  app.ctx_user())` — `ctx_user()` is NULL under replay. `app.enrollment_restrict_on_decision()` (AFTER UPDATE OF status
  ON `enrollment_requests` FOR EACH ROW WHEN `NEW.status IN ('denied','revoked') AND OLD.status <> NEW.status`): the
  same helper call on `NEW.organization_id` (a no-op re-lock when (3′) is already held), then `INSERT (organization_id, user_id, kind,
  recorded_via, created_by_user_id, source_request_id) VALUES (NEW.organization_id, NEW.user_id, NEW.status, 'trigger',
  coalesce(NEW.revoked_by_user_id, NEW.decided_by_user_id), NEW.id)`. Both inserts pass **only** through
  `enrollment_restrictions_owner_insert` (§4.4); the runtime role holds **no write grant** on the ledger. The bodies
  are not idempotent by design (every departure/denial/revocation is a new row; `one_pending` and the status machine
  bound the count).
  CT-ENR-1 inventories cw-owned functions: exactly **6** after 00016 (`invitation_claim_group`,
  `invitation_claim_organization`, `rls_user_enabled_lock`, `enrollment_request`, `enrollment_admit`,
  **`enrollment_lock_organization`** — EXECUTE to ow only, asserted); ar-owned +1; pr-owned +2; owner-owned DEFINER
  trigger bodies +2.

### 4.4 Grants and policies (`rt`/`ow`/`pr`/`ar`/`cw` as 00015; `U/O/MGR` as 00006)

- `organizations`: rt `UPDATE (enrollment_mode)` under `organizations_enrollment_update` FOR UPDATE USING `id = O AND
  MGR` WITH CHECK (same) + BEFORE UPDATE OF `enrollment_mode` invoker trigger `organizations_enrollment_guard` requiring
  `rls_org_role() = 'org_owner'` for rt; rt `SELECT (enrollment_mode)` added to the runtime column list (facts stay
  behind the existing `organizations_select`, 00006:432). **pr**: `GRANT SELECT (id, enrollment_mode)` +
  `organizations_policy_reader` FOR SELECT USING (true). **cw** (r2, correction 1): `GRANT SELECT (id,
  enrollment_mode), UPDATE (enrollment_mode) ON app.organizations` + `organizations_claim_select` FOR SELECT USING
  (true) + **`organizations_claim_lock`** FOR UPDATE USING (true) WITH CHECK (false). `SELECT … FOR SHARE` requires the
  UPDATE privilege on at least one column **and** a passing FOR UPDATE policy: the column grant supplies the
  privilege (exactly G019's `GRANT UPDATE (status) ON app.users TO cw` + `users_claim_lock … WITH CHECK (false)`,
  00015:638/653), the `USING (true)` admits the lock-read, and `WITH CHECK (false)` makes every cw UPDATE a policy
  violation — no successful applicant UPDATE exists, and the SELECT columns stay `id`/`enrollment_mode`. The r0
  runtime-facing `organizations_enrollment_reader` is **withdrawn** (no applicant row lock from the runtime role; no
  exposure of the full runtime column list).
- `enrollment_sessions`: **rt** `SELECT`, `INSERT`, `UPDATE (revoked_at, revoke_reason)` (r2: the consumption pair is
  gone), `DELETE` under `enrollment_sessions_select/_insert/_update` (`user_id = U`) and **`enrollment_sessions_expire`** FOR
  DELETE USING `expires_at <= now() AND U IS NULL AND B IS NULL` (the 00006:472-473 shape; executed by `Service.Login`'s
  E5 cleanup, `DELETE FROM app.enrollment_sessions` in the EMPTY shape — this **is** the sweep: no catalog, no job);
  **ar** `GRANT SELECT` + `enrollment_sessions_auth_reader` USING (true); **cw** `GRANT SELECT, UPDATE (admitted_at,
  admitted_binding_id)` + `enrollment_sessions_claim_select` USING (true) + `enrollment_sessions_claim_update` USING
  (true) WITH CHECK `user_id = ctx_user()`.
- `enrollment_requests`: **rt** `SELECT` under `enrollment_requests_select` USING `user_id = U OR (organization_id = O
  AND MGR)`; `UPDATE (status, decided_at, decided_by_user_id, revoked_at, revoked_by_user_id)` (r3: **no privilege on
  `authority_seq`** — the BEFORE trigger assigns it, which needs no column privilege; a direct `SET authority_seq` is
  42501) under
  `enrollment_requests_decide` USING `organization_id = O AND MGR AND status IN ('pending','approved')` WITH CHECK
  `organization_id = O AND MGR AND status IN ('approved','denied','revoked')`; **no rt INSERT**. **cw**
  `INSERT (id, organization_id, user_id, status, expires_at)` under `enrollment_requests_claim_insert` WITH CHECK
  `user_id = ctx_user() AND status = 'pending'`; `SELECT` under `enrollment_requests_claim_select` USING (true);
  `UPDATE (consumed_at, consumed_binding_id)` under `enrollment_requests_claim_update` USING (true) WITH CHECK
  `user_id = ctx_user()`.
- `enrollment_restrictions`: **rt** `SELECT` under `enrollment_restrictions_select` USING `user_id = U OR
  (organization_id = O AND MGR)` (status page / admin list only; rt also reads `max(seq)` for the approval watermark
  through this policy); **cw** `SELECT` + `UPDATE (lifted_at, lifted_by_user_id, lifted_via)` under
  `enrollment_restrictions_claim_select` USING (true) and `enrollment_restrictions_claim_lift` USING (true) WITH CHECK
  `user_id = ctx_user() AND lifted_at IS NOT NULL`; **ow** (r2, correction 2): **`enrollment_restrictions_owner_insert`**
  `FOR INSERT TO ow WITH CHECK (pg_trigger_depth() >= 1 AND kind IN ('departed','denied','revoked') AND recorded_via =
  'trigger' AND lifted_at IS NULL AND lifted_by_user_id IS NULL AND lifted_via IS NULL)` — the owner needs no GRANT
  (table owner) but is bound by FORCE RLS, so this policy is the **only** INSERT path and it is unusable outside a
  trigger; no INSERT grant or policy for rt, cw, pr or ar; no DELETE for any role; no ow UPDATE/DELETE policy (the owner
  cannot lift or erase a row either). The G013 replay materialization runs as the superuser (RLS bypassed, `recorded_via
  = 'replay'`; §1.4) and is not a role grant.
- `audit.events`: **new cw policy `events_insert_enrollment`** FOR INSERT WITH CHECK `pg_trigger_depth() >= 1 AND
  actor_kind = 'user' AND actor_user_id = ctx_user() AND organization_id IS NOT NULL AND action = 'enrollment.requested'`.
  `events_insert_claim` (00015:655-657) and runtime `events_insert` (00008:668-671) are **not** modified; decision,
  revocation and policy events are inserted by runtime actors **with** Organization context and pass `events_insert`
  as today. **No audit SELECT grant to pr** (r0 D-4 withdrawn; ADR-006 tenant audit-read boundary unchanged).
- `role_bindings`: no new grant or policy. `is_organization_member`: already EXECUTE rt/ow/cw. **D-15 (i) only:**
  `GRANT EXECUTE ON FUNCTION app.invitation_claim_organization(uuid, bytea, text, text, text) TO <db>_runtime`.

### 4.5 Triggers and audit

Invoker trigger functions (owner-owned): `organizations_enrollment_guard` (BEFORE UPDATE OF `enrollment_mode`);
`enrollment_requests_immutable` (BEFORE UPDATE, SECURITY INVOKER, r3 correction 2: rt may move `pending →
approved|denied` once with `decided_*`; on `pending → approved` the guard itself executes `PERFORM 1 FROM
app.organizations WHERE id = NEW.organization_id FOR NO KEY UPDATE` (the deciding session is a current MGR member, so
`organizations_update` admits it; the facade's (3′) lock makes this a same-transaction no-op) and then assigns
`NEW.authority_seq := coalesce((SELECT max(seq) FROM app.enrollment_restrictions WHERE organization_id =
NEW.organization_id AND user_id = NEW.user_id), 0)` — the SELECT runs under rt's `enrollment_restrictions_select` MGR
branch — **ignoring whatever the statement supplied**; `denied` leaves it 0; for every other transition (`approved →
revoked`, cw's consumption pair, anything else) `NEW.authority_seq IS DISTINCT FROM OLD.authority_seq` raises
`insufficient_privilege`; `approved → revoked` once with `revoked_*` only while `consumed_at IS NULL`; cw may set the
consumption pair once on an `approved` row; everything else immutable). The watermark is therefore derived by SQL
under the Organization lock, immutable after the first decision, and unreachable by any runtime value. DEFINER trigger bodies: §4.3 (two). Audit captures: `audit.capture_enrollment_requests`
(AFTER INSERT → `enrollment.requested`; AFTER UPDATE OF status → `enrollment.decided` for `approved|denied`,
`enrollment.revoked` for `revoked`; `resource_type='enrollment_request'`, `subject_user_id` = applicant, actor =
`ctx_user()` — the applicant for `requested`, the decider/revoker otherwise; the consumption UPDATE is not captured);
`audit.capture_organization_enrollment` (AFTER UPDATE OF `enrollment_mode` → `enrollment.policy_changed`).
`audit.events_attribute()` replaced only to add the `enrollment.*` shape CASE (hash-diffed against the 00015 body).
Mandatory atomic append: every capture runs inside the writing transaction; a refused append (policy, constraint,
injected fault) raises and the transaction rolls back (no request, no decision, no binding). Catalog: trigger actions
33 → **37** (`enrollment.requested`, `enrollment.decided`, `enrollment.revoked`, `enrollment.policy_changed`),
resource types 10 → **11**, reasons 58 → **62** (`enrollment_closed`, `enrollment_restricted`, `approval_revoked`,
`invitation_mismatch`), operations 53 → **63**. **Down:** refuses when the three tables hold rows or `audit.events`
holds `enrollment.*` actions; otherwise drops objects (incl. the two restriction triggers, the owner INSERT policy, the
cw `UPDATE (enrollment_mode)` grant and the four cw/pr/ar policies on `organizations`/`audit.events`), restores the
00015 `events_attribute` body and catalog lists, drops the three `login_transactions` columns and the mode column.
Down does not touch `infra/recovery/sql/replay-restrict.sql` (its guard makes it correct with or without the ledger).

## 5. Contracts and interfaces

### 5.1 Go surfaces (deltas from r0)

```go
// core/internal/identity — identity imports neither access nor tenancy (its import set is unchanged at c5f83fbf).
type AddressEvidence struct { address string; verified bool; observedAt time.Time; issuer, subject string } // unexported, no JSON tags
// Trusted construction contract (r2, correction 6): EvidenceMinter is an EXPORTED struct with an unexported seal
// (`seal *struct{}`); only identity constructs a non-zero value (unexported newEvidenceMinter) and it is handed to the
// Provider PER CALL in the existing Exchange input (identity.go:24: `Exchange(ctx, in Exchange) (…)` gains
// `in.Minter EvidenceMinter` and returns ExchangeResult). Mint(vi VerifiedIdentity, address string, verified bool,
// at time.Time) AddressEvidence binds issuer/subject into the value; a zero EvidenceMinter mints a zero
// AddressEvidence, which attests nothing. No interface method is unexported, nothing is wired in cmd/server/main.go,
// workos.New/identitytest.NewFake signatures are unchanged. AttestVerifiedAddress(ctx, ev, vi) refuses evidence whose
// bound issuer/subject differ from the exchanged identity. TestAddressEvidenceUnforgeable: only identity/workos and
// identity/identitytest call Mint (package pin), and a zero or copied value cannot attest.
type EvidenceMinter struct { seal *struct{} }
type ExchangeResult struct { Identity VerifiedIdentity; Address AddressEvidence }
type InvitationRef [16]byte                      // identity-owned id value; access converts to/from tenancy.InvitationID
type LoginRequest struct { /* existing */; EnrollmentIntent bool; Invitation *InvitationRef; Request bool }
type CallbackRequest struct { /* existing, incl. the forwarded session Credential */; Invitation *InvitationRef; InvitationToken string }
type EnrollmentHandle struct { UserID, OrganizationID, Origin string; Digest [32]byte } // identity-owned, no provider facts
type AdmitKind string                             // "invitation" | "open" | "request"
type AdmitSelector struct { Kind AdmitKind; Invitation *InvitationRef; InvitationToken string; Address *AddressEvidence; Request *[16]byte }
type AdmitOutcome struct { Status string; BindingID string }  // identity-owned result DTO (closed outcome set of §1.4)
// Cycle-free orchestration: the callback branch calls an Admitter declared in identity with identity-owned DTOs only:
type Admitter interface { AdmitAtCallback(ctx context.Context, h EnrollmentHandle, sel AdmitSelector) (AdmitOutcome, error) }
// implemented by access.Service (which converts InvitationRef → tenancy.InvitationID and its internal enrollment
// result → identity.AdmitOutcome) and injected into identity.Service in cmd/server/main.go (main already imports both).
type EnrollmentOutcome struct { Status string; Credential string /* once */; Next string /* "sign_in" on admitted/already_member */ }
func (s *Service) EnrollmentCredential(ctx, credential, origin string) (EnrollmentHandle, error) // auth_resolve_enrollment
// removed in r2 (D-17): IssueAfterAdmission — issuance is the unchanged ordinary Login/Callback/signIn path.
// core/internal/tenancy/access
func (s *Service) RevokeEnrollmentApproval(ctx, tc, id [16]byte) (EnrollmentRequestView, error) // new
func (s *Service) AdmitAtCallback(ctx, h identity.EnrollmentHandle, sel identity.AdmitSelector) (identity.AdmitOutcome, error) // converts, then:
func (s *Service) admit(ctx, h identity.EnrollmentHandle, kind string, request *[16]byte) (enrollment.AdmitResult, error)      // body call only
func (s *Service) admitInvitation(ctx, ev identity.AddressEvidence, h identity.EnrollmentHandle, id tenancy.InvitationID, token string) (enrollment.AdmitResult, error) // P19-4 + lift, one tx; P0001 enrollment_restricted → rolled back
```
`core/internal/enrollment/api` imports `identity` (credential resolution) and `access` (operations) — importers of
each +1; `tenancy` importers unchanged. Raw-tenancy DTOs never cross into identity.

### 5.2 HTTP — `core/internal/enrollment/api` (**9 pairs on 8 paths**) plus two modified identity routes

As r0 §5.2 plus `POST /internal/v1/enrollment/requests/{id}/revoke` (Session, `organization.manage_members`; `200
{"request"}`; `409 already_consumed|not_approved`; idempotent on `revoked`). Callback answers add `409
invitation_mismatch` and `409 identity_conflict` (existing conflict mapping reused). The `admit` answer (r2) is `200
{"outcome":"admitted"|"already_member","next":"sign_in"}` on a confirmed T2 — **never a session**; the BFF clears the
enrollment cookie and the status page starts the ordinary login. `409 enrollment_restricted` also covers the stale
approval/invitation cases; `503 enrollment_unavailable` on T1/T2 unknown. `admitted_sign_in_required` is removed.

### 5.3 Web — as r0 (`enrollment-client.ts` exports **10**: + `revokeRequest`). 5.4 Downstream — as r0.

## 6. Validation and test mapping

### 6.1 Existing coverage inspected — as r0 §6.1.

### 6.2 Scenario groups → tests (r0 rows retained with the corrections below; new rows ENR-15..21)

- **CT-ENR-1**: adds the three tables (13/15/13 columns incl. `seq`, `recorded_via`, `authority_seq`), the ledger's
  exact grant/policy inventory (**only** `enrollment_restrictions_owner_insert` can insert; no rt/cw/pr/ar INSERT; no
  DELETE; no ow UPDATE/DELETE), the ar/cw/pr DEFINER inventory (**8**, r3: `enrollment_lock_organization` cw-owned with EXECUTE to ow only, body hash pinned; `enrollment_requests_immutable` body hash incl. the `authority_seq` derivation), `organizations_claim_lock` (WITH CHECK false)
  **and the cw column grants `SELECT (id, enrollment_mode)`/`UPDATE (enrollment_mode)`** (correction 1),
  `events_insert_enrollment` text, absence of any pr grant on `audit.events`, the two DEFINER trigger bodies (hash,
  owner, the TG_*/depth guard text), and `enrollment_sessions_expire`.
- **ENR-9** (request flow): adds `revoke` after approval → `approval_revoked` at admit; revoke idempotent; revoke after
  consumption → `already_consumed`; denial then `admit kind=open` on an `open` tenant **after any elapsed time** →
  `enrollment_restricted` (no cooldown); a new request after denial is allowed and its approval lifts the restriction;
  **r2:** approval, then a second request denied (newer `denied` row) → the first approval admits `enrollment_restricted`
  (zero writes, request still `approved`); the admin revokes it and approves a third request → admitted, both older
  rows lifted (`lifted_via='approved_request'`, `seq <= authority_seq`).
- **ENR-10** (lock seams; `TestAccessOrdering` O-ENR-2..8): mode closure vs admission in both orders; revocation vs
  consumption in both orders (one winner, zero partial writes); host reassignment; `RemoveMember` racing the
  post-admission ordinary sign-in (session refused `organization_not_member` by `issuanceGuard`, `departed` row
  recorded); **r2 O-ENR-6** approval vs removal in both orders (watermark covers / `seq` newer); **O-ENR-7** two
  consumers of one approval (one binding, loser `already_member`); **O-ENR-8** invitation claim vs departure in both
  orders (claim rolled back by the `lift` RAISE with the invitation still pending; or departure restricting the fresh
  membership); plus bound logout vs admission in both orders (`enrollment_expired` / revoked after admission with
  `admitted_at` intact).
- **ENR-11** (faults): adds T1-unknown (fault at COMMIT) → `unavailable`, no credential; T2-unknown → `unavailable`,
  then `status` reports by read only and a second `admit` answers `already_member` with no second binding; **r2:** the
  T3 row is the existing identity fault coverage (`identity_unavailable` on the ordinary sign-in; no enrollment row is
  read or written by it — asserted by a row-count canary around `signIn`).
- **ENR-12/13** (restrictions): sources now `enrollment_restrictions`; adds **ENR-15** retention/purge survival: delete
  the subject's `membership.removed` audit rows as `<db>_retention` (`events_retention_delete` eligibility path, no new
  mechanism) → open admit still `enrollment_restricted`; **ENR-16** replay (r2, three restored states of §1.4):
  (a) checkpoint with the binding → replay deletes → trigger row (`recorded_via='trigger'`); (b) checkpoint with the
  restriction → `updated 0`/`none`/`ledger present`, row intact; (c) **checkpoint predating grant and removal** →
  `updated 0`/`none`/`ledger materialized`, one `departed` row with `recorded_via='replay'`, `created_by_user_id` NULL,
  `created_at` newer than every restored approval/invitation → open admit refused, a restored approval refused, a
  restored invitation claim rolled back; Group-scope and `absent`-reference replays materialize nothing (conflict
  unchanged); a second replay of the same entry materializes nothing (`ledger present`). Existing V8 subcases
  (`recovery_integration_test.go:599-646`) are retained unchanged.
- **ENR-22** ledger insertion under FORCE (r2, correction 2): each real path writes exactly one row — `RemoveMember`,
  `LeaveScope`, the G013 replay delete (superuser), `DecideEnrollmentRequest(denied)`, `RevokeEnrollmentApproval`;
  negatives: rt/cw/pr/ar `INSERT` → 42501 (no privilege); owner `INSERT` at depth 0 (as `<db>_owner` in the migrate
  test harness) → 42501 policy violation; `SELECT app.enrollment_restrict_on_departure()` outside a trigger → refused
  by the guard; owner `UPDATE`/`DELETE` → no policy; **atomic rollback**: inside one test transaction as owner, `DROP
  POLICY enrollment_restrictions_owner_insert`, then as rt `RemoveMember` → the DELETE statement fails, the binding is
  still present and the ledger is empty after rollback (same for a denial: the request stays `pending`); **r3
  (correction 1)**: the existing G019 self-leave case is extended — a subject's `LeaveScope` of its **last** Organization
  membership commits, writes exactly one `departed` row and leaves the Organization row locked only until commit
  (the helper, not the tenant policy, supplied the lock); the rollback case above is re-run on that same self-leave
  path; `SELECT app.enrollment_lock_organization(…)` as rt → 42501 (no EXECUTE), as cw outside a body → not granted
  to any session role but the owner (asserted in CT-ENR-1).
- **ENR-23** reinstatement authority (r2, correction 4): stale approval (older `authority_seq`) → `enrollment_restricted`
  with zero writes and no lift; stale invitation (`created_at` ≤ newest unlifted row) → claim rolled back
  (`invitations.consumed_at` still NULL, no binding, no lift, outcome `enrollment_restricted`); fresh approval lifts
  only rows `<= authority_seq`; fresh invitation lifts only older rows; `authority_seq` immutable after approval
  (trigger); a `denied` decision cannot carry a non-zero `authority_seq` (CHECK); **r3 (correction 2)**: one focused
  direct-SQL inflated-watermark negative — as rt in a MGR context, `UPDATE app.enrollment_requests SET status =
  'approved', decided_at = now(), decided_by_user_id = ctx_user(), authority_seq = 9223372036854775807 WHERE id = …` →
  42501 (no column privilege; the row stays `pending`, no audit event) — plus the ordinary approval: the same UPDATE
  without `authority_seq` commits with `authority_seq` equal to the greatest committed `seq` of that `(org, user)` (0
  when none), a restriction recorded afterwards has a higher `seq` and the approval is refused at admission.
- **ENR-17** NULL-context applicant audit: `enrollment_request` as a nonmember with IDENTITY shape appends exactly one
  `enrollment.requested` with `organization_id` set; a direct runtime `INSERT INTO audit.events` with that shape and
  NULL `ctx_org()` → policy violation (00008:668-671 unchanged); a direct runtime `INSERT INTO enrollment_requests` →
  42501 (no grant); the applicant's `status`/list answers expose no other subject or tenant column (minimal
  disclosure canary).
- **ENR-18** identity continuity: live session for userA on the host, callback exchanging providerB with enrollment
  intent → `identity_conflict`, no mapping/enrollment/grant for B, A's session row unchanged; same user, new provider
  session → `session_rotated` via the ordinary path; at T3 a live session for a different user → refused.
- **ENR-19** pending-credential revocation at scope (r2, correction 7): two enrollment sessions for one user minted
  from provider sessions sid1 and sid2; logout of the local session bound to sid1 → only the sid1 row revoked, the sid2
  row still resolves; provider-logout reference for sid1 → same; supersession of the sid1 local session (`signIn`
  rotation) → sid1 row revoked, sid2 untouched; logout-all → both revoked; `DisableUser` → both revoked and
  `auth_resolve_enrollment` empty regardless; `admit` after each revocation → `enrollment_expired`; `departed`/`denied`
  restriction → credential valid, admit refused; an admitted row revoked after admission keeps `admitted_at`.
- **ENR-20** invitation intent binding: concurrent `/join/start` for invitations X then Y in one browser → X's callback
  `invitation_mismatch` (nothing created), Y's proceeds; duplicate `__Host-join` header → treated as absent →
  `invitation_mismatch`; stale cookie after a completed flow is cleared (unit: BFF sets `Max-Age=0` on every callback
  outcome); token never in `Location`/logs (canary).
- **ENR-21** credential bootstrap negatives: `auth_resolve_enrollment` with a revoked/expired/disabled-user row → no
  row; the runtime role cannot read `enrollment_sessions` in the EMPTY shape except through `_expire` DELETE eligibility
  (`TestRLSIsolation` rows); an enrollment credential against `auth_resolve_session` → no row.
- **ENR-H1..H9**: + revoke route; `invitation_mismatch`/`identity_conflict` mappings.
- Browser: **E-ENR-1** open tenant B with `erin`; **E-ENR-2** request tenant C with the asserted-nonmember seeded
  applicant, admin approve in a second browser, return → admitted; also admin **revoke** before return → status
  `approval_revoked`; **E-ENR-3** invite-only A, first-time invitation link for `grace` (emulator `email_verified`
  permitting; else recorded as fake-only); **E-ENR-4/5** as r0.
No role×mode×host×theme cross-product; no new control framework; no default full suite.

### 6.3 Batches — as r0 §6.3 with the new test names added to the b1/b2 selectors (`TestAddressEvidenceUnforgeable`, `TestEnrollmentCredentialBootstrap`, `TestJoinCookie`), `TestRecoveryPost` ENR-13/15/16 subcases (no drill rerun; ENR-16(c) restores a pre-grant checkpoint fixture through the existing `TestRecoverySeed` seed path), `'^TestIsolationAttacks$/^(ISO-ATTEST|ENR)'` covering ENR-17/21/22, and `TestAccessOrdering` O-ENR-2..8 (ENR-10/23). Required final local coverage unchanged in shape (affected checks only; no default `full` profile).

### 6.4 Manual review (M1), fixtures and readiness

As r0 §6.4, plus fixture eligibility pins: the request-mode applicant for E-ENR-2/ENR-9 is a seeded user asserted
`is_organization_member(C, user) = false` at setup (if no seeded user qualifies, `mcp-seed.sql` leaves one seeded
account without a C binding — never a subject whose normal login would already succeed on C); first-time recipients
are account-less emulator users: `erin` (E-ENR-1, open B) and new `grace` (E-ENR-3, invite-only A), so the two
first-time cases are independent. M1 adds: revoke an approval before the applicant returns; a denied applicant's open
retry after the admin flips B to `open` (still refused).

## 7. Delegation — as r0 §7 (R-A's checklist adds the restriction ledger, `events_insert_enrollment`, the cookie contract and the continuity branch).

## 8. Sequence — as r0 §8; step 0 adds the D-4r1 disposition; slice 1 includes the ledger, both cw bodies, `auth_resolve_enrollment`, revocation paths and ENR-15..21.

## 9. Acceptance-criteria mapping

| AC | Evidence |
|---|---|
| AC1 | ENR-1/2/3, CT-ENR-1 (as r0); revoke authority in ENR-3/H9 |
| AC2 | ENR-4/5, ENR-18 (continuity), ENR-19 (scoped pending-credential revocation), ENR-21 (bootstrap + cw lock negatives), ENR-22 (ledger insertion), `TestRLSIsolation` rows |
| AC3 | ENR-6/7/8/9 through G019's contract and the two cw bodies; `TestAddressEvidenceUnforgeable`; ENR-20 (intent binding); ENR-23 (authority validity) |
| AC4 | ENR-10 (closure/revocation/host/removal/authority races), ENR-8 duplicate, ENR-11 (T1/T2 abort and unknown; T3 = existing identity coverage), the §1.3 outcome table |
| AC5 | ENR-12 (ledger-based refusals), ENR-13/15/16 (restore, purge, three replay states incl. materialization), ENR-9 (no lapse; stale approval), ENR-22 (atomic insertion), D-4r1 pinned in CT-ENR-1; legacy departures = D-16 |
| AC6 | E-ENR-1, ENR-14 (as r0) |
| AC7 | ENR-H1..H9, ENR-17 (NULL-context audit path and minimal disclosure), ENR-20 cookie/log canaries, E-ENR-2..5, M1 |
| AC8 | as r0 |

## 10. Owner package (r3): three UNSELECTED decisions, two recorded behaviors, and the review items

**One coherent package for owner disposition, after the r3 corrections.** Three decisions are **UNSELECTED** and
nothing in this Plan selects them, infers them from an owner enabling a generic feature, or treats a Codex
recommendation as an owner decision: **D-4r1** (durable restriction authority), **D-15** (execution/provider-proof
trust boundary) and **D-16** (legacy policy). Two behaviors are recorded as stated and have no remaining technical
blocker in review r2: **D-17** (fresh ordinary sign-in after admission) and **D-18** (scoped revocation). The owner
package is complete when all three decisions carry an owner disposition; until then the §2 execution gate stays
closed (D-15, D-4r1 and D-16 undecided at release are escalation risks). The r0 design items D-1..D-9 remain review
items beneath the package.

| ID | Item | Recommendation / status |
|---|---|---|
| **D-15** | **Proof/execution authority for first-time-invitee public admission.** G019's `invitation_claim_organization` is EXECUTE-able by `<db>_owner` only (00015:881). The production consumer needs (a) an execution path and (b) a provider-proof source. **Neither option proves cryptographically that the address came from WorkOS**; the proof is the adapter's TLS-bound `Exchange` plus the code-path pins (`EvidenceMinter`, call-site pin, canary tests). **Option (i) runtime EXECUTE** (00016 `GRANT EXECUTE … TO <db>_runtime`): SQL verifies principal → `ctx_user()`, byte-equal normalized address, pending state and `intended_user_id`; what SQL cannot verify is that the runtime obtained the address from the provider — the same trust the runtime already holds over `users`/`identity_principals`/`sessions`, and the same EXECUTE shape G019 already grants the runtime for `invitation_claim_group`, `rls_user_enabled_lock` (00015:879-880) and this Plan proposes for `enrollment_request`/`enrollment_admit`. Cost: one grant; no new credential/pool/role. **Option (ii) dedicated execution credential/pool** holding EXECUTE alone: isolates execution of the claim from ordinary runtime connections (a compromised ordinary runtime path cannot call it), but the address still arrives from the same Go adapter over that connection — it does **not** establish provider evidence either; cost: new login role, pool, provisioning, secret handling, drills, `infra/postgres/**` scope, G013/G021 reconciliation; delays G058. **Dropping first-time support is not an option** (owner's 2026-10-17 decision; a new explicit owner requirement change would be needed). **Recommendation for the owner: (i)**, because both options rely on the same code-path trust for provenance and (ii) buys execution isolation at major infrastructure cost without closing the evidence gap; if the owner wants execution isolation as a standing property, (ii) should be scoped as its own Goal rather than inside G058. Review r2 confirms: no runtime EXECUTE and no new credential/pool/authority model is granted by review; the SDK `EmailVerified` projection remains a held, unqualified prerequisite (§1.8) for either option | **owner/security decision via Codex — UNSELECTED (the Plan's (i) preference is a recommendation only)** |
| **D-4r1** | **Durable enrollment restriction ledger** (`app.enrollment_restrictions`, §1.4/§4.2/§4.5) written only by owner-owned trigger bodies on departure/denial/revocation **through one owner-only, trigger-bound INSERT policy** (r2), by the G013 superuser replay for replayed departures (r2, `recorded_via='replay'`), and lifted only by explicit authority **valid for the current restriction state** (r2: `seq`/`authority_seq` for approvals, chronology under the Organization lock for invitations) inside admission. It replaces r0's audit-derived check (withdrawn) and the seven-day cooldown (withdrawn). It is a **new durable authority record** that gates open signup; the alternative that avoids a new record (no durable check) fails AC5 under retention/purge, and the audit-based check was rejected by review. Scope: one table (`seq`, `recorded_via`), two triggers, two cw policies, one rt SELECT policy, one ow INSERT policy, one guarded replay statement; no G019 code change (the trigger observes G019's deletes; G019's rows are read, never written). The ledger does not change who may grant membership. r3 closes the two SQL defects found in r2 review (self-leave lock via the cw helper; `authority_seq` derived in SQL) without changing the authority model. **Coherent proposal, still UNSELECTED** | **owner decision via Codex (security/data) — UNSELECTED** |
| **D-16** (r2 correction 3; r3 owner package) | **Legacy departures and pre-00016 checkpoints — a real legacy policy choice.** Memberships removed before 00016 have no ledger row; an empty ledger cannot reconstruct absent historical departures, and the Plan neither backfills from audit (audit-derived authority was withdrawn in r1) nor resets anything. **Exact recovery consequence (visible, unchanged):** a checkpoint restored from before 00016 lacks the table; `replay-restrict.sql` records `ledger absent` and materializes nothing, so after migration every pre-ledger departure of that Organization is unrestricted for open mode until it is re-recorded by an operator or re-removed — and a replayed `membership.remove` whose checkpoint postdates 00016 but predates the grant **is** materialized (`recorded_via='replay'`). **Options, none selected:** **(a) start empty** — the ledger begins empty, the §1.4 readiness assumptions hold (every Organization `invite_only` at 00016), and the adopter runbook states that pre-G058 departures are unrestricted until re-removed; *the r2 default recommendation of (a) is withdrawn*; **(b) operator-reviewed history/backfill** — explicit, operator-authored G013-style replay entries per legacy departure (authority supplied and reviewed by the operator, never inferred from audit), scoped as its own bounded task, not G058; **(c) fail-closed open-mode readiness (Codex root's recommendation, review r2)** — an Organization with unresolved legacy history cannot enable `open` (and, if the owner so decides, `request_access`) until either (b) has been completed for it or the absence of legacy history has been **verified** (e.g. the Organization was created after 00016, or an operator attests no pre-ledger removals); fresh installations have no legacy history and need **no fabricated backfill**; the readiness state would be an explicit, operator-set fact on the Organization, not an inference, and the same gate would re-close after a pre-00016 recovery (`ledger absent`). Cost of (c): one readiness fact + one check in `UpdateEnrollmentPolicy`/the mode guard and an operator procedure — bounded, but a scope addition that only an owner disposition authorizes. The choice also governs what the owner expects after the pre-00016 recovery above | **owner decision via Codex — UNSELECTED; Codex root recommends (c); this Plan recommends nothing** |
| **D-17** (r2, correction 5) | **Issuance after admission = fresh ordinary sign-in** (§1.3 step 4), the review's explicitly permitted route. Rejected alternative: a transaction-sharing `signInInTx` primitive with enrollment-row locking, revocation recheck through issuance and stored issuer/subject/provider facts on the enrollment row — more identity surface on the qualified sign-in path for one caller, and a stored principal reference the design otherwise avoids. Cost accepted: one extra provider redirect after admission (not assumed silent); the `admit` answer never carries a session | review item (recorded decision; reversible only by a new Plan revision) |
| **D-18** (r2, correction 7) | **Logout scope for pending enrollment credentials.** Bound: current-session logout, provider logout and supersession revoke only rows minted from the same `provider_session_id`; broader: logout-all and disablement revoke all. **Flagged semantics for explicit review:** a pending enrollment credential minted in browser 2 survives a logout in browser 1 (consistent with sessions: logout is per session, logout-all is user-wide), and a successful ordinary sign-in does **not** revoke pending credentials of other provider sessions. If the owner wants every logout to be user-wide for enrollment, that is a broader security semantic to decide here, not a WHERE clause to change silently | security review item |
| D-1 | Separate credential table/cookie/header — as r0 | security review item |
| D-2 | Mapping before membership only when the mode admits or a pending invitation id exists, and only after the continuity check — r0 amended | review item |
| D-3r1 | Invitation transport: fragment → `__Host-join` one-use cookie (§1.6 contract) → callback JSON body; bound to the stored login transaction | security review item |
| D-5 | `rls_invitation_pending` boolean reader — as r0 | review item |
| D-6 | **withdrawn** (no runtime-facing Organization lock policy); replaced by cw `organizations_claim_lock` inside the body | — |
| D-7 | Approval does not write membership — as r0 | acceptable in principle (proposed) |
| D-8r1 | `Provider.Exchange` returns `ExchangeResult`; `AddressEvidence` minted only through `EvidenceMinter` | review item |
| D-9 | `enrollment_request`/`enrollment_admit` EXECUTE to runtime (cw bodies, `invitation_claim_group` precedent) | review item |
| D-10 | **superseded by D-17** (no deferred issuance from stored facts) | — |

**Escalation risks (stop, do not design around):** D-15, D-4r1 or D-16 undecided at release; the SDK response type not
exposing `EmailVerified` (§1.8); material G019 drift at R0′ (§0.1); any need for a new database role, credential, pool,
email vendor, issuer or public Organization creation; widening of `role_bindings_select`, `rls_org_role()`, the depth-1
audit rule, `events_insert` or `events_insert_claim`; relaxing `issuanceGuard`; an unreviewed destructive migration; a
42P17 policy cycle; context/budget exhaustion (handoff at ~250k). **No split is recommended** (five slices, one cycle).

## 11. Budget, pins and scope

### 11.1 Budget — provisional **B7, checkpoint after attempt 4; unspent; recommendation: keep B7**. Reassessed forecast **4 attempts** (r1: 4; r2: 4): r2 removes the deferred-issuance primitive and adds the owner INSERT policy, the authority columns, the guarded replay statement and two scenario groups; r3 adds one cw lock helper and moves the watermark into the existing guard — net neutral. Planning spends no attempt; no reset, no new validation campaign, no duplicate tests.

### 11.2 Count pins (provisional; re-pin at R0′; r2 deltas marked)

Migrations 15 → **16**; capabilities **19 → 19**; operations 53 → **63** (`enrollment.public`, `.status`, `.request`,
`.admit`, `.abandon`, `.policy.get`, `.policy.update`, `.requests.list`, `.requests.decide`, `.requests.revoke`);
trigger actions 33 → **37**; core actions 6; reasons 58 → **62**; resource types 10 → **11**; SECURITY DEFINER +**8**
(r3; allowance 00016 = 8: pr 2, ar 1, cw **3** incl. `enrollment_lock_organization`, owner trigger bodies 2); cw-owned
functions 3 → **6**; EXECUTE grants to ow +1 (the helper; none to rt); SECURITY INVOKER
trigger functions +2 (`organizations_enrollment_guard`, `enrollment_requests_immutable`) + replaced `events_attribute`;
audit capture triggers +2; restriction triggers +2; policies +**21** (r2: `organizations` +4: `_enrollment_update` rt,
`_policy_reader` pr, `_claim_select` cw, `_claim_lock` cw; `enrollment_sessions` +7: rt 3 + `_expire` + ar 1 + cw 2;
`enrollment_requests` +5: rt 2 + cw 3; `enrollment_restrictions` +**4**: rt 1 + cw 2 + **ow 1
`_owner_insert`**; `audit.events` +1 cw `events_insert_enrollment`); grants (r2): rt `organizations` SELECT +1 /
UPDATE +1 column, pr `organizations` SELECT `(id, enrollment_mode)`, **cw `organizations` SELECT `(id,
enrollment_mode)` + UPDATE `(enrollment_mode)`**, ar `enrollment_sessions` SELECT, **rt `enrollment_sessions`
UPDATE(2: `revoked_at`, `revoke_reason`)**, cw `enrollment_sessions` SELECT/UPDATE(2), **rt `enrollment_requests`
UPDATE(5; r3: no `authority_seq` privilege)**, cw `enrollment_requests` INSERT(5)/SELECT/UPDATE(2), cw `enrollment_restrictions` SELECT/UPDATE(3),
**no INSERT grant on `enrollment_restrictions` for any role**, **no pr grant on `audit.events`**, **no rt INSERT on
`enrollment_requests`**; columns: `enrollment_sessions` **13**, `enrollment_requests` **15** (+`authority_seq`),
`enrollment_restrictions` **13** (+`seq`, `recorded_via`), `login_transactions` +3; **D-15 (i)**: +1 EXECUTE;
importers of `access` +1 and of `identity` +1 (`enrollment/api`), `tenancy` importers unchanged, identity's own
imports unchanged; HTTP **9 pairs / 8 paths** new + 2 modified auth routes; `CORE_ENROLLMENT_ROUTES` 9;
`CORE_AUTH_ROUTES` 6; `completion.schema.json` +8 paths; `events.json` +1 event, +10 operations; `required-checks.tsv`
125 → **127** (B-ENROLLMENT-CATALOG, B-ENROLLMENT-IT); `WEB_ROUTES` 39 → **48** (no new sign-in route); `static_web_boundary`
+5; `CookieKind` +2 (`enrollment`, `join`); `APPLICABLE` +4; `UI_FILES`/`WorkspaceArea` unchanged; Playwright projects
+2; `BROWSER_MATRIX_IDENTITY_E2E` +1; `profile-paths.tsv` +3; `validate-controls.sh` +1 `F-widen-*` row; recovery:
`replay-restrict.sql` +1 guarded statement, +1 keyed output line, `recovery.sh`/`lib.sh` unchanged (R0′ checks whether
`fingerprint-catalog.sql` pins table inventories; if it does, that is a pin update to flag, not a mechanism change);
seed: B `open`, C `request_access`, A default, one asserted-nonmember applicant; emulator seed +1 account-less user.
Budgets 8 KiB / 16 KiB.

### 11.3 Scope allow-list — `management/v0/plans/G058-scope-allowlist.txt` (r2 **proposed**, not current execution scope; Codex imports as `management/v0/results/G058-scope-allowlist.txt`): **45 executable patterns** (r1 44 + `^infra/recovery/sql/replay-restrict[.]sql$`, the one bounded recovery addition of correction 3). Exclusions as r0 §11.3 (no `infra/postgres/**`, no other `infra/recovery/**` file — `recovery.sh`, `lib.sh`, the other replay scripts and `fingerprint-catalog.sql` stay excluded — no `policy.go`, historical migrations, `docs/adr/**`, Codex-owned files).

## 12. Revision history and tool refusals

| Rev | Date (UTC) | Inputs | Summary |
|---|---|---|---|
| r0 | 2026-10-18 | `origin/staging` `1b5688cf`; G019 draft MR !43 at `702f1ea4`; branch `g058-plan`; worktree `agent-<id>` | initial provisional Plan (`a022c35d…2f2b`, 744 lines) and 43-pattern scope (`b8fd6d2e…545a8`); D-15 recommendation (i) |
| r1 | 2026-10-18 | Codex review r0 (`dd0d22e7…9d100b`); `origin/staging` `044cfd52`; G019 branch head read for 00015/`invitations.go` interfaces; branch `g058-plan-r1`; worktree `agent-<id>`; model Fable 5.1, effort "xhigh (requested)", actual unknown | bounded correction of review items 1–8 and the validation/scope reconciliation (§0.A map): durable restriction ledger (D-4r1, escalated), cw lock-read contract and one lock order, cw applicant audit path, approval revocation, per-stage outcome table, continuity + pending-credential revocation, `__Host-join` cookie contract, `EvidenceMinter`/cycle-free orchestration, `auth_resolve_enrollment` bootstrap, SDK projection marked unqualified; D-15 restated, not selected; B7 unspent, forecast 4; 44-pattern proposed scope (`f82875d4…079e`, 563 lines; `bbacc370…9df0`). No implementation, build, test, probe, Docker or MR; r0 archives unchanged |
| **r2** | 2026-10-18 | Codex review r1 (`6a7d7aad…13fc02`, 79 lines); `origin/staging` `c5f83fbf`; targeted code spots at `c5f83fbf` (`identity/store.go` transaction ownership and revocation paths, `identity.go:24`, 00002:57, 00006:319/432-434, 00008:668, 00009:291-297, `replay-restrict.sql`, `recovery.sh:822/865/894`, `recovery_integration_test.go:599-646`, `access/boundary_test.go:206`) and 00015 at the G019 branch head (`:108-127`, `:133-179`, `:485`, `:633-657`, `:945-976`); branch `g058-plan-r2`; worktree `agent-<id>`; model Fable 5.1, effort "xhigh (requested)", actual unknown; reading capped ≈48k tokens | bounded correction of review items 1–7 (§0.B map): cw lock-only `UPDATE (enrollment_mode)` grant; owner-only trigger-bound INSERT policy with body guards and atomic-rollback proof; bounded replay materialization (+1 recovery scope), legacy state and readiness assumptions, D-16 escalated; authority version (`seq`/`authority_seq`) and invitation chronology under the Organization lock, `lift` RAISE; T3 = fresh ordinary sign-in (D-17; deferred issuance, `auth_time`, consumption pair and `admitted_sign_in_required` removed); identity-owned Admitter DTOs, sealed per-call `EvidenceMinter`, no new tenancy importer; scoped logout revocation (D-18 flagged). D-4r1 and D-15 presented coherently, **unselected**; SDK projection still unqualified/held; B7 unspent, forecast 4; 45-pattern proposed scope. No implementation, build, test, probe, Docker or MR; r0/r1 archives, reviews, Goal, Spec, PROGRESS/ROADMAP and code unchanged |

| **r3** | 2026-10-18 | Codex review r2 (`45336e91…9d86`); `origin/staging` `2e209db1` (holds the r2 archives `97145547…a1e` / `50a045c4…2a9`); bounded reading: the review, the r2 Plan lines it cites (§1.4:205–208, 281–287; §4.3:484–487; §4.4:522–524; §4.5:550–554) and the `rls_user_enabled_lock`/EXECUTE-grant precedent (00015:680-700, 877-881); branch `g058-plan-r3`; worktree `agent-<id>` (recreated at the same path); model Fable 5.1, effort "xhigh (requested)", actual unknown | two bounded SQL corrections (§0.C): (1) self-leave — the departure/decision bodies lock the Organization through a new cw-owned DEFINER helper `app.enrollment_lock_organization` (EXECUTE to ow only; lock-only policy; valid after the caller's binding is deleted and under replay), the `ctx_org()` conditional removed, existing self-leave/rollback cases extended; (2) `authority_seq` derived by the BEFORE UPDATE guard under the Organization lock, rt column privilege removed, one direct-SQL inflated-watermark negative plus the ordinary approval. §10 rewritten as one owner package: D-4r1, D-15, D-16 UNSELECTED; D-16 presents Codex root's fail-closed open-mode readiness recommendation beside (a)/(b), withdraws the r2 "start empty" default, keeps the pre-00016 recovery consequences visible; D-17/D-18 unchanged. Pins: DEFINER allowance 8, cw-owned 6, rt `enrollment_requests` UPDATE(5); scope unchanged (45). No implementation, build, test, probe, Docker or MR; r0/r1/r2 archives, reviews, Goal, Spec, PROGRESS/ROADMAP and code unchanged |

**r2 and r3 tool refusals: none.** The r1 refusal below is preserved verbatim; its subject (the module-cache read) was not
retried in r2 and §1.8 stays unqualified.

**Tool refusal in r1 (verbatim, not retried, not routed around):** one read-only batch that would have listed the Go module
cache path for `workos-go/v10@v10.5.0` and grepped `EmailVerified` was refused: "This agent is isolated in the worktree
<source-checkout>, but this command names git
in a form too complex to verify that it stays inside the worktree. Refusing to run it — a worktree-isolated agent's git
operations must target its own worktree. Split it into plain, separate commands and run them from
<source-checkout>" The independent in-repo
reads in that batch were re-issued as plain commands; the module-cache read was not, so §1.8 records the projection as
unqualified.
