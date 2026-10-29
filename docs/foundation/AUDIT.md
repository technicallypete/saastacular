# Durable audit events

The durable, tenant-scoped audit record of foundation-level mutations, security denials and external storage outcomes.
Migration `core/migrations/00008_audit_events.sql` (version 8)
creates one append-only relation, `audit.events`, in the owner-owned `audit` schema that provisioning creates empty.
Row changes of eight audited `app` tables are recorded by database capture triggers in the same transaction; core
appends only the events that are not row changes. The increment is core-internal: there is no HTTP, BFF or MCP surface,
no export, no retention job and no product read. Operational logs stay operational evidence; this record is the audit.

## Components

| Part | Where | What it does |
|---|---|---|
| Relation | `audit.events` | One row per event: server UUID, schema version 1, `occurred_at` (database clock at append), the closed action, outcome and reason, the guarded operation of a denial, the derived actor (`user`, `system`, `unknown`), optional Organization and Group, the resource, the subject user, roles of membership changes, the TTL of a signed URL, and diagnostic request/trace correlation (NULL when malformed). No foreign keys: deleting a source row never deletes history. No names, slugs, emails, provider subjects, keys, URLs, content types, sizes, payloads or error text. |
| Capture triggers | 18 triggers on 8 `app` tables; from version 10, 20 on 9 (two `AFTER ROW` triggers on `app.mcp_connections`) | 17 `AFTER ROW` triggers and one `AFTER STATEMENT` trigger with transition tables (session revocations) call owner-owned SECURITY INVOKER functions that copy named columns of the changed row. `WHEN` clauses and `UPDATE OF` lists skip no-op updates and unaudited columns. |
| Attribution trigger | `events_attribute` (BEFORE INSERT) | Sets every server-owned field from the transaction context, enforces source separation (`pg_trigger_depth()`), validates a direct append's Organization and Group, and enforces each action's shape. |
| Append-only trigger | `events_refuse` (BEFORE UPDATE, DELETE, TRUNCATE) | Refuses every role, the owner and superusers included, unless a privileged administrator disables it. |
| Go primitive | `core/internal/audit` | The closed catalog constants, `Append` (one INSERT of the descriptive columns on the caller's transaction; trigger-owned actions refused before anything is sent), `ReadTenant` and the revocation-gap marker read. Imports pgx only. |
| Recorders | `identity` and `tenancy/access` | `Store.inTxAudited` (sign-in success, same transaction), the bounded identity recorder `recordAuditEvents`, and the facade's `auditTx` (independent records after the business transaction released its connection). |
| Guarded read | `access.Service.ReadAuditEvents` | The current `org_owner` of the active Organization (`audit.read`); internal Go only. |
| Diagnostics | `audit.record_failed`, `audit.gap` | Operational records when an independent record could not be written, or a committed revocation's row-change event was lost (events.json). |

## Catalog

The 35 actions (29 from version 8; T24 and T25 added at version 10 with `core/migrations/00010_mcp_connections.sql`; T26–T29 at version 11 with `core/migrations/00011_organization_domains.sql`). T-actions are recorded only by a capture trigger (depth 2 or more); G-actions only by `Append` (depth 1).
The database refuses a core insert of a T-action and a trigger insert of a G-action, so each event has exactly one
producer.

| # | Action | Producer | Outcome · reasons | Resource · scope · subject |
|---|---|---|---|---|
| T1 | `organization.created` | trigger: organizations INSERT | success | organization · Organization |
| T2 | `organization.renamed` | trigger: UPDATE OF name (changed) | success | organization · Organization |
| T3 | `organization.policy_changed` | trigger: UPDATE OF the flag (changed); UPDATE OF `group_management_override` (changed; the owner's guarded write) | success · `group_admin_profile_edit_enabled`, `group_admin_profile_edit_disabled`, `group_management_override_enabled`, `group_management_override_disabled` | organization · Organization |
| T4 | `group.created` | trigger: groups INSERT | success | group · Organization, Group |
| T5 | `group.renamed` | trigger: UPDATE OF name (changed) | success | group · Organization, Group |
| T6 | `group.slug_changed` | trigger: UPDATE OF slug (changed) | success | group · Organization, Group |
| T7 | `membership.granted` | trigger: role_bindings INSERT | success | role_binding · Organization, Group? · subject, role |
| T8 | `membership.role_changed` | trigger: UPDATE OF role (changed) | success | role_binding · Organization, Group? · subject, role, previous role |
| T9 | `membership.removed` | trigger: role_bindings DELETE | success | role_binding · Organization, Group? · subject, role |
| T10 | `business_profile.created` | trigger: business_profiles INSERT | success | business_profile · Organization |
| T11 | `business_profile.updated` | trigger: UPDATE OF the 5 context columns (row changed) | success | business_profile · Organization |
| T12 | `business_profile.linked` | trigger: group_business_profiles INSERT | success | business_profile · Organization, Group |
| T13 | `business_profile.unlinked` | trigger: group_business_profiles DELETE | success | business_profile · Organization, Group |
| T14 | `user.created` | trigger: users INSERT | success | user · global · subject |
| T15 | `user.disabled` | trigger: status active → disabled (revocation priority) | success (system in the lifecycle producer) | user · global · subject |
| T16 | `user.enabled` | trigger: status disabled → active (strict) | success (system in the lifecycle producer) | user · global · subject |
| T17 | `session.issued` | trigger: sessions INSERT | success | session · global · subject |
| T18 | `session.revoked` | statement trigger: revoked_at NULL → set, one row per revoked session (expired-but-unrevoked included) | success · `logout`, `logout_all`, `provider_logout`, `disabled` (revocation priority), `superseded` (strict) | session · global · subject |
| T19 | `session.provider_revocation_recorded` | trigger: UPDATE OF provider_revocation (changed) | success (provider reported `confirmed`), failed (`failed`): provider-reported, not verified provider state | session · global · subject |
| T20 | `object.upload_started` | trigger: stored_objects INSERT | success | stored_object · Organization, Group? |
| T21 | `object.ready` | trigger: state pending → ready | success | stored_object · Organization, Group? |
| T22 | `object.delete_started` | trigger: state pending/ready → deleting (logical deletion) | success | stored_object · Organization, Group? |
| T23 | `object.deleted` | trigger: state deleting → deleted (absence confirmed; tombstone) | success | stored_object · Organization, Group? |
| T24 | `mcp.connection_established` | trigger: mcp_connections state pending → active (strict: activation fails without its event) | success | mcp_connection · Organization · subject (the owner) |
| T25 | `mcp.context_changed` | trigger: UPDATE OF selected_group_id (changed; strict) | success | mcp_connection · Organization, Group? (the new selection; none when cleared) · subject (the owner) |
| T26 | `organization.slug_changed` | trigger: UPDATE OF slug (changed) | success | organization · Organization |
| T27 | `domain.claimed` | trigger: organization_domains INSERT | success | organization_domain · Organization |
| T28 | `domain.state_changed` | trigger: UPDATE OF state (changed) | success · the new state: `verified`, `active`, `suspended`, `removed` | organization_domain · Organization |
| T29 | `domain.checked` | trigger: UPDATE OF last_check_at (changed, outcome set) | success (`ok`, `withdrawn`), failed · `mismatch`, `nxdomain`, `transient`, `tls_error`, `withdraw_error`, `withdraw_unknown` | organization_domain · Organization |
| T30 | `organization.branding_changed` | trigger: INSERT or UPDATE of app.organization_branding | success · `updated`, `reset`, `logo_set`, `logo_removed` (derived from the old and new rows; no value recorded) | organization · Organization |
| T31 | `invitation.issued` | trigger: app.invitations INSERT | success · the invitable role; never the address or token | invitation · Organization, Group? · intended subject when addressed to a user |
| T32 | `invitation.revoked` | trigger: UPDATE setting `revoked_at` | success · the invitable role | invitation · Organization, Group? |
| T33 | `invitation.consumed` | trigger: UPDATE setting `consumed_at`, inside a claim body | success · the invitable role; subject required | invitation · Organization, Group? · consumer |
| T34 | `enrollment.policy_changed` | trigger: UPDATE OF enrollment_mode (changed) — the Organization owner's guarded policy write, or a readiness close of an open Organization | success | organization · Organization |
| T35 | `enrollment.readiness_changed` | trigger: UPDATE OF the readiness pair (changed), only inside the superuser operator procedures `app.enrollment_readiness_set` / `app.enrollment_history_backfill` or the replay marker | success · the basis (`operator_verified_absent`, `operator_backfilled`) or `history_unresolved` on a close; actor unknown; the operator reference as request_id | organization · Organization |
| T36 | `enrollment.restriction_backfilled` | trigger: app.enrollment_restrictions INSERT with `recorded_via = 'backfill'` (operator backfill only) | success · `operator_backfilled`; actor unknown; the operator reference as request_id | organization · Organization · backfilled subject |
| T37 | `enrollment.requested` | trigger: app.enrollment_requests INSERT, inside the claim writer's `app.enrollment_request` | success; no reason; actor = the applicant | enrollment_request · Organization · applicant |
| T38 | `enrollment.decided` | trigger: UPDATE OF status to `approved` or `denied` by an Organization manager | success; no reason (approval and denial are read from the request row, not the event) | enrollment_request · Organization · applicant |
| T39 | `enrollment.revoked` | trigger: UPDATE OF status `approved` → `revoked` of an unconsumed approval by an Organization manager | success; no reason | enrollment_request · Organization · applicant |
| T40 | `mcp.connection_renamed` | trigger: UPDATE OF display_name on app.mcp_connections (changed: set, changed or cleared; strict: the rename fails without its event) | success; no reason; never the name, the previous name or the new name | mcp_connection · Organization · subject (the owner) |
| T41 | `label.created` | trigger: app.label_definitions INSERT | success; no reason; never the name | label (the definition id) · Organization, Group = the owning Group of a Group-owned definition (none for an Organization-owned one) |
| T42 | `label.updated` | trigger: UPDATE OF name or visibility on app.label_definitions, one event per changed column; a no-op update records nothing | success · `renamed`, `visibility_changed`; never the old or new name | label (the definition id) · Organization, Group = the owning Group of a Group-owned definition |
| T43 | `label.deleted` | trigger: app.label_definitions DELETE | success; no reason | label (the definition id) · Organization, Group = the owning Group of a Group-owned definition |
| T44 | `label.assigned` | trigger: app.label_assignments INSERT | success; no reason; never the target type or a non-Group target id | label (the definition id) · Organization, Group = the target Group (required) |
| T45 | `label.unassigned` | trigger: app.label_assignments DELETE, including the assignments removed by a definition deletion | success; no reason | label (the definition id) · Organization, Group = the target Group (required) |
| G1 | `auth.sign_in` | core: success in sign-in's transaction; denied/failed independently after the one-use login transaction was consumed | success · `signed_in`, `session_rotated`; denied · `transaction_expired`, `login_canceled`, `code_rejected`, `impersonation_refused`, `mfa_unverified`, `stale_authentication`, `identity_conflict`, `user_disabled`; failed · `provider_error`, `provider_timeout`, `identity_namespace_mismatch`, `identity_unavailable`; success · `signup_authenticated` (root signup callback), `tenant_created` (pinned tenant callback); denied · `signup_origin_rejected`, `signup_user_mismatch` | global · subject when verified |
| G2 | `access.denied` | core: independently, after the denied transaction | denied · `organization_not_permitted`, `group_not_selectable`, `target_not_found`, `target_kind_mismatch`, `target_outside_context`, `role_not_permitted`, `no_linked_path`, `edit_policy_disabled`, `not_creator`, `object_not_ready`, `capability_unknown`, `fact_inconsistent`, `resource_mismatch` (only with `mcp.request`, global, no target), `enrollment_closed`, `enrollment_restricted`, `approval_revoked`, `invitation_mismatch`, `recipient_ineligible`, `session_invalid`, `host_changed` (only with an `enrollment.*` operation, global, no target), `signup_invalid`, `intent_used`, `slug_taken`, `continuation_invalid`, `signup_context_mismatch` and `host_changed` (only with a `signup.*` operation, global, no target); one of the 80 operations below | target only when established (forbidden); Organization only when established, else global |
| G3 | `object.url_issued` | core: audit transaction after the offline presign, before the URL is returned | success · TTL | stored_object · Organization, Group? |
| G4 | `object.upload_failed` | core: independently, after a post-T1 failure and its compensation attempt | failed/unknown · `compensated`, `compensation_failed`, `not_compensated` | stored_object · Organization, Group? (none after a later policy denial) |
| G5 | `object.deletion_pending` | core: independently, when physical removal or the tombstone is unconfirmed | failed/unknown · `provider_unavailable`, `tombstone_failed`, `tombstone_unknown` | stored_object · Organization, Group? |
| G6 | `revocation.unaudited` | core (identity): only after a revocation committed with its gap marker set | success · `logout`, `logout_all`, `provider_logout`, `disabled` | session (`logout`, `provider_logout`) or user (`logout_all`, `disabled`) · global · subject |
| G7 | `access.override_used` | core: in the mutating transaction before COMMIT, only when the explicit Group path denies and the Organization override allows an ordinary Group administration write; an append failure refuses the write | success; no reason; the operation (Group rename/slug, onboarding completion, member role change or removal, invitation issue or revoke, profile update) | group (Group id as resource and scope) or business_profile (Group only for an existing-profile completion) · Organization · actor (Organization owner or admin) |

The 80 operations of `access.denied`: `access.resolve`, `organization.get`, `organization.get_policy`,
`organization.set_policy`, `group.get`, `group.rename`, `group.change_slug`, `business_profile.get`,
`business_profile.update`, `object.put`, `object.get`, `object.metadata`, `object.sign`, `object.delete`, `audit.read`,
`mcp.request`, every protected MCP call, `domain.list`, `domain.claim`, `domain.rotate`,
`domain.verify`, `domain.recheck`, `domain.remove`, `domain.retry_withdrawal` and `organization.set_slug`,
`branding.get`, `branding.update`, `branding.reset`, `branding.set_logo` and `branding.remove_logo`, and
`group.suggest_slug`, `group.create`, `group.list`, `group.onboarding.get`, `group.onboarding.complete_new`,
`group.onboarding.complete_existing` and `business_profile.list`, and `membership.list`, `membership.self`,
`membership.change_role`, `membership.remove`, `membership.leave`, `ownership.promote`, `ownership.demote`,
`ownership.transfer`, `invitation.list`, `invitation.mine`, `invitation.candidates`, `invitation.issue`,
`invitation.revoke`, `invitation.get`, `invitation.accept`, `invitation.consume` and `retirement.eligibility`, and
`enrollment.start`, `enrollment.public`, `enrollment.status`, `enrollment.request`, `enrollment.admit`,
`enrollment.abandon`, `enrollment.policy.get`, `enrollment.policy.update`, `enrollment.requests.list`,
`enrollment.requests.decide` and `enrollment.requests.revoke`, and `organization.set_group_override`, and
`signup.start`, `signup.status`, `signup.slug`, `signup.create`, `signup.challenge`, `signup.prove`, `signup.complete`
and `signup.abandon`, and `label.definition.list`, `label.definition.create`, `label.definition.update`,
`label.definition.delete`, `label.assignment.list`, `label.assign` and `label.unassign`.

**Labels ([TENANCY.md](TENANCY.md#labels)).** Migration 00020 adds the five trigger-owned actions T41–T45
(45 trigger and 7 core actions, 52 in all), captured by the strict triggers `audit_capture_label_definitions`
(`AFTER INSERT OR UPDATE OF name, visibility OR DELETE ON app.label_definitions`) and `audit_capture_label_assignments`
(`AFTER INSERT OR DELETE ON app.label_assignments`) through the owner-owned invoker functions
`audit.capture_label_definitions`/`audit.capture_label_assignments`; the reasons 79 → 81 (`renamed`,
`visibility_changed`, only on `label.updated`), the seven `label.*` operations above (73 → 80) and the resource type
`label` (11 → 12). Every event names the label definition id as its resource and the Organization; `group_id` is the
owning Group of a Group-owned definition (none for an Organization-owned one) or, for `label.assigned`/`unassigned`,
the target Group (required by the shape rule). No name, target type, non-Group target id or payload enters the row, and
a no-op update records nothing. The replaced `audit.events_attribute` admits the five actions as trigger-owned under
that shape (`success`, Organization required, no role, subject or TTL) and `label` as an `access.denied` resource type;
no label operation appends `access.override_used`. A duplicate assign or a refused write records no row. The down
migration refuses while any label row or any label action, operation, reason or resource is recorded.

**Connection names ([MCP.md](MCP.md)).** Migration 00019 adds one trigger-owned action, T40
`mcp.connection_renamed` (40 trigger and 7 core actions, 47 in all), captured by the new strict trigger
`audit_capture_name` (`AFTER UPDATE OF display_name ON app.mcp_connections`, only when the value changes) through the
replaced `audit.capture_mcp_connections`; `audit.events_attribute` admits it with the `mcp.*`
shape (Organization, no Group, no reason). A set, a change and a clear are all one `renamed` event; a same-value rename
writes none. No name, previous or new value enters the row. No reason, operation or resource type is added (the
`access.decision` operation `mcp.connection.rename` is a diagnostic record, not an audit operation).

**Platform root signup ([IDENTITY.md](IDENTITY.md#platform-root-signup)).** Migration 00018 adds no action
(still 39 trigger and 7 core actions, 46 in all) and no resource type (11); the reasons 70 → 79 (the four `auth.sign_in`
reasons and the five signup refusals above; `host_changed` is reused) and the eight `signup.*` operations (65 → 73), with
the `events_attribute` body narrowing each new reason to its producer and shape (the signup refusals only with a
`signup.*` operation, no Organization, Group or target; `host_changed` there only with `signup.challenge` or
`signup.complete`). A root creation records `organization.created` (T1) and `membership.granted` (T7, the creator's
`org_owner` binding) through the existing capture triggers in the provision writer's creation statements, attributed to
the creator, under the provision writer's only audit authority `events_insert_provision` (trigger depth ≥ 1, those two
actions only, actor = the context user, an Organization set); an identical retry records nothing new, and a failed
creation retains no row and no event. Signup refusals are recorded independently afterwards (unattributed for the
tenant hops). The down migration refuses while an 00018 reason or operation is recorded.

**Explicit Group authority ([AUTHORIZATION.md](AUTHORIZATION.md)).** Migration 00017 adds the core action G7
`access.override_used` (39 trigger actions, 7 core actions, 46 in all), the reasons 68 → 70
(`group_management_override_enabled`, `group_management_override_disabled` on `organization.policy_changed`) and the
operation `organization.set_group_override` (64 → 65); resource types stay 11. The Organization owner's guarded write
of the default-false `app.organizations.group_management_override` is recorded by the `organizations` capture as
`organization.policy_changed` with the old/new state in its reason; a refused write (any non-owner, a forged or generic
policy patch, a direct runtime UPDATE stopped by the owner-only guard) records no change. `access.override_used` names
the actual mutated resource and is accepted by `events_attribute` only from an Organization owner or admin while the
override is currently on; reads log the decision reason `allowed_override` in `access.decision` only. The down migration
refuses while any Organization has the override on or an 00017 action, reason or operation is recorded.

**Organization enrollment.** Migration 00016 adds the six trigger-owned `enrollment.*` actions T34–T39 (33 +
6 = 39 trigger actions, 6 core actions, 45 in all), the `enrollment_request` resource type (10 → 11), the reasons 58 → 68 (the six new enrollment refusals of G2 —
`host_changed` is reused —, `recipient_proof_unavailable` on `auth.sign_in` denied, and the readiness reasons `operator_verified_absent`,
`operator_backfilled`, `history_unresolved`), the eleven `enrollment.*` operations above (53 → 64) and the
`events_attribute` body with the enrollment shapes. `enrollment.policy_changed` is the owner's guarded policy write
(`UpdateEnrollmentPolicy`); its refusal of `open` while membership history is unresolved (`enrollment_not_ready`) writes
no audit row. `enrollment.readiness_changed` and `enrollment.restriction_backfilled` come only from the superuser
operator procedures (actor unknown, operator reference as `request_id`). The three request captures are appended in
the writing transaction (a refused row rolls the request or decision back): `enrollment.requested` by the claim writer
under its own `events_insert_enrollment` policy, `enrollment.decided`/`enrollment.revoked` by the deciding runtime
session; the consumption of an approval is recorded by the admission's `membership.granted` only. The down migration
refuses while an 00016 action, reason, operation or resource type is recorded.

**Membership, ownership and invitations ([TENANCY.md](TENANCY.md#membership-ownership-and-invitations)).**
Migration 00015 adds the three `invitation.*` actions (T31–T33, 30 + 3 trigger actions, 6 core actions), the
`invitation` resource type (9 → 10), the seventeen operations above (36 → 53) and the `events_attribute` body with the
invitation shape (the 00013 body otherwise byte-identical; `TestMembershipCatalog` hash-diffs it). Role changes,
promotions, demotions, transfers, removals and departures are recorded by the existing `role_bindings` captures
(`membership.role_changed`, `membership.removed`); a consumption records `membership.granted` or
`membership.role_changed` and `invitation.consumed` in the claim's transaction through `events_insert_claim`, the claim
writer's only audit authority (trigger-issued captures of those three actions, attributed to the context subject). No
reason or outcome is added. The down migration refuses while any invitation or invitation-era audit row exists.

**Group onboarding ([TENANCY.md](TENANCY.md#group-onboarding)).** No action, reason, resource type or
outcome is added; the existing capture triggers record the row changes with the request's context user as actor. A
guarded Group creation records `group.created` (T4) and then `membership.granted` (T7) for the creator's `group_owner`
binding, both within the one Group INSERT statement (the capture triggers sort before the bootstrap trigger, so T4
precedes T7; the attribution trigger-source rule admits both unchanged); a new-profile completion records `business_profile.created` (T10) then `business_profile.linked`
(T12); an existing-profile completion records `business_profile.linked` only. An audit fault refuses the change (no
orphan profile is committed). Migration 00014 widens only `events_operation_valid` with the seven operations above (one
validating scan of `audit.events`); its down refuses while any event carries one of them. Recordable denials of the
seven operations follow the rule below (for example `access.denied` `group.create` `organization_not_permitted` after
the subject's Organization membership was removed); slug and association conflicts are operational only.

**Domains and tenant-host sign-in ([DOMAINS.md](DOMAINS.md)).** Every domain row change and slug change is
recorded by the T26–T29 capture triggers in its own transaction; hostnames and challenge values are never stored in
an event (the resource is the row UUID). A tenant-host sign-in refused by the issuance guard is G1 denied with
`organization_not_member` (no current membership of the resolved Organization, D-2(a)) or `host_changed` (the host no
longer routes to the login's Organization); the login transaction is consumed and no session is issued.

**MCP connections ([MCP.md](MCP.md)).** Activation and every Group selection or clear are recorded in their own
transaction by the T24/T25 capture triggers; an audit fault refuses the change (strict). Termination (the logout tool,
web revoke, sign-out everywhere, disable, expiry) is the existing `session.revoked` of the connection's `mcp` session.
A policy refusal of a protected MCP call is `access.denied` with operation `mcp.request`. A verified, mapped subject
presenting a token at a connection, host or client its own connection is not bound to is `access.denied` with reason
`resource_mismatch`: global, no target, actor the connection's owner. An audience that matches no connection is refused
during token verification and is operational only. No token, subject, client identifier or connection URL is stored.

**Recordable denial:** a policy refusal after a valid binding check (`context_denied` with
`organization_not_permitted` or `group_not_selectable`, or `not_found` / `forbidden`). Unauthenticated calls, binding
misuse, invalid input, conflicts and unavailable or internal failures are operational only. A rejected foreign selector
never becomes a recorded UUID: the target is recorded only for `forbidden` (readable on the same facts), so unknown and
foreign targets record identically. A storage T2 or D4 policy refusal after an allowed T1/D1 is not a new denial; its
G4/G5 event carries no Organization.

**Not recorded:** pre-consume authentication refusals, successful `Resolve`, successful reads, renewal, protocol
bookkeeping (login transactions, logout references), display-name refreshes, the namespace binding, idempotent no-ops
(no row changed) and simulated crash seams.

**Retry identity.** No idempotency keys and no automatic retries. A trigger event exists exactly when its row change
committed; a core event is attempted once; a caller's retry is a new operation with new events. `occurred_at` is the
append time; for independent records it follows the decision by at most the record budget.

## Attribution and trust

- **Actor** is derived by `events_attribute`, never supplied. Lifecycle actions (`user.disabled`, `user.enabled`,
  `session.revoked` and `revocation.unaudited` with reason `disabled`) are `system` only inside the lifecycle producer
  (`DisableUser`/`EnableUser`, which run in the target's IDENTITY shape); other events are `user` with the context user,
  else `unknown` (privileged or bootstrap SQL without a context, anonymous authentication failures). No actor is
  invented from missing context.
- **Correlation**: the context statement ([DATABASE-ISOLATION.md](DATABASE-ISOLATION.md)) also sets the transaction-local `audit.request_id` and `audit.trace_id`
  (diagnostic only, never compared or used in a decision) and resets the revocation gap marker. They are outside the
  prior-context guard and confer no authority.
- **Direct appends** (core events): the Organization is accepted only when it equals the context Organization and
  `app.rls_org_role()` is non-NULL now; a Group only when `app.groups` (read under the recording transaction's row-level
  security) has it in that Organization. With a Group, the current Organization role is proven by that lookup itself:
  `groups_select` requires `app.rls_org_role()` to be non-NULL, so no separate reader call is made. This equivalence holds
  for the runtime role and the provisioned owner under FORCE row-level security only; a superuser, a BYPASSRLS role or an
  administrator altering the policies is outside the claim, as are forged runtime contexts (the database-isolation trust limit, [DATABASE-ISOLATION.md](DATABASE-ISOLATION.md)). A
  Group named without a current Organization role is refused as `audit_event_group`. Organization records use the
  TENANT-ORG shape without a selected Group;
  Organization-less records the IDENTITY shape (EMPTY for anonymous authentication failures). Trigger rows take
  Organization and Group from the changed row.
- **Guaranteed by the database:** no runtime UPDATE, DELETE, TRUNCATE, DDL or GRANT; no runtime-supplied actor, time, id
  or correlation; trigger-owned events only from real row changes; closed shapes; Organization attribution consistent
  with the established context; a direct-append Group that belongs to the event Organization; the revocation exception
  limited to the five pinned transitions in owner-owned trigger code, with no runtime trigger-disable or bypass.
- **Not claimed:** that a superuser, the owner, or a holder of the runtime credential with arbitrary SQL cannot fabricate
  or suppress events (the database-isolation trust limit): such a holder can set any context and append core-owned actions for it. A
  raw-runtime self re-enable would be recorded as `system`; a self un-revocation (`revoked_at` set → NULL) is not an
  existing path and is not captured. A privileged session that revokes during an audit fault commits with no reader of
  the gap marker. Privileged administration is outside the claim.

## Failure semantics

| Mechanism | Events | Audit-append failure | Unknown COMMIT |
|---|---|---|---|
| M1 trigger | T1–T23 except the M5 transitions | the statement fails and the transaction rolls back: no unaudited success (facade `internal`/`unavailable`, identity `identity_unavailable`) | change and event both durable or neither; no retry |
| M2 core, same transaction | G1 success | sign-in fails `identity_unavailable`; no session, no events | unchanged identity semantics; no retry |
| M3 core, independent | G1 denied/failed, G2, G4, G5 | the caller's outcome is unchanged; one `audit.record_failed`; no durable row is claimed | the event may or may not exist; no retry |
| M4 issuance | G3 | no URL is returned (`unavailable`, `storage.operation` stage `audit`) | no URL; at most one `object.url_issued` ("issued, not delivered") |
| M5 revocation priority | T15 active → disabled; T18 `logout`, `logout_all`, `provider_logout`, `disabled`; then G6 | only the audit insert rolls back inside a bounded exception block; the revocation commits; the transaction-local gap marker is read before COMMIT; after COMMIT one `revocation.unaudited` is attempted and one `audit.gap` (`independent`: `recorded` or `failed`) is written; the call returns its normal success | a failed or unknown COMMIT with the marker set: `identity_unavailable`, no G6, `audit.gap` `not_attempted` |

- Sign-in (including `superseded` rotation), enable, grants and every other mutation stay strictly atomic. The session
  revocation trigger reads its transition tables only and executes no audit statement unless a revocation transition is
  present, so an audit fault never aborts a logout or disable before the exception block, and renewal executes no audit
  statement. `query_canceled` is never caught.
- Provider logout's later bookkeeping (`session.provider_revocation_recorded`) is strict: an audit fault fails only that
  step (`identity_unavailable`, as for any failure of it) and never undoes the committed revocation.
- Independent records run on a context detached from the caller's cancellation, bounded by a 2 s record budget and a
  2 s cleanup budget (an expired cleanup closes the held connection, which the pool destroys). They are written after
  the business transaction released its connection, so no request holds two pooled connections, and before the call's
  operational record. If the database is unavailable, the diagnostic is the only trace.
- No provider or network I/O inside an audit-holding transaction; no audit statement takes row locks on `app` tables.

## Schema, grants and inspection

- Provisioning (`infra/postgres/init/10-roles.sh`, `provision-roles.sql`) creates the empty schema `audit` owned by
  `<db>_owner` (existing volumes: re-run `provision-roles.sql` before migrating). Deploy order: provision → migrate to 8
  → deploy code. Old code against version 8 works (triggers record; correlation NULL); new code against version 7
  fails closed at sign-in.
- `ENABLE` and `FORCE ROW LEVEL SECURITY`. `events_select`: the context Organization and `app.rls_org_role() =
  'org_owner'`. `events_insert`: a `user` actor equals the context user, and the Organization is NULL or the context
  Organization. No UPDATE or DELETE policy.
- Runtime: `USAGE` on `audit`, `SELECT` (filtered), `INSERT` of 12 descriptive columns. Nothing else; no EXECUTE in
  `audit`, no PUBLIC entry. `<db>_retention` holds no audit privilege.
- **Global rows** (no Organization) match no runtime or owner policy. Inspecting them is a privileged operator action:
  a superuser, or the owner after an explicit, reviewed `ALTER TABLE audit.events NO FORCE ROW LEVEL SECURITY`
  maintenance step that is reverted afterwards. There is no platform-admin role.
- Down (development and test only) refuses while any row exists. A disposable test database purges history through
  its superuser (`DISABLE TRIGGER events_refuse_truncate`, `TRUNCATE`, `ENABLE TRIGGER`), never in production.

## Read API

`ReadAuditEvents(ctx, tc, AuditReadRequest{Since, Until, Limit, Cursor})` runs through the guard (`audit.read`,
REPEATABLE READ READ ONLY): the time window `[Since, Until)` (zero: no bound), page size 0 → 50, 1..100
(`invalid_page_size`), `Until` after `Since` (`invalid_time_range`), and an opaque cursor bound to the Organization and
window (`invalid_cursor`). Events are newest first by `(occurred_at, id)` through `events_tenant_time_idx`. A selected
Group grants nothing new: an owner in selected-Group context reads the whole Organization; Group roles read nothing. A
page is a snapshot: an event committed later with an earlier time may be absent from an ongoing traversal and appears
in a fresh one. Successful reads are not audited; denials are (`audit.read`).

## Performance

Audited row changes add no
round trip; a revocation adds one (the gap-marker read) and writes one statement-level insert of N rows for N revoked
sessions (expired-but-unrevoked included); each independent record and each signed URL adds a 4-statement transaction
on a second, briefly held pooled connection; a direct append with a Group makes at most 3 reader calls (two
Organization-reader init-plans and at most one Group reader inside the RLS Group lookup, no outer call); every sessions
UPDATE runs the statement trigger's aggregate over its transition tables. Audit growth is bounded only by the
operator's explicit retention policy ([RECOVERY.md](RECOVERY.md)); there is no automatic job.
E-AUDIT-1..3 (B-AUDIT-IT) record the counts and the tenant-read plan; unexpected scans, fan-out, reader growth or lock
changes print `E-AUDIT-TRIGGER` lines and return to review.

## Validation

| Check | Proves |
|---|---|
| B-AUDIT-CATALOG | TestAuditCatalog: the exact relation, constraints, index, RLS, policies, every ACL, the 18 + 3 trigger definitions, the 11-function allowlist and bodies, runtime and retention refusals, owner FORCE, and eight rolled-back missing-protection detectors |
| B-AUDIT-MIGRATION | TestAuditMigration (empty → 8, a populated version-7 world → 8 with every row and ACL identical, refused unclean schemas, the down refusal with history, empty down/up and reapply, a rolled-back up) and persistence across a PostgreSQL restart |
| B-AUDIT-IT | TestAudit: every one of the 29 actions observed from real calls (`AUDIT-INV`), exact fields and refused forgeries, same-transaction atomicity, strict faults, revocation priority and the privilege fault, identity fan-out, storage outcomes, owner-only reads, page bounds and continuation, no canary and no deletion path, E-AUDIT-1/2 |

The inherited suites (B-POLICY-*, B-RLS-*, B-OBJ-* including the storage attack suite) run at version 8 with the triggers
active.

## Retention contract

`<db>_retention` holds no audit privilege at version 8. Migration `00009_retention_maintenance.sql` (version 9) adds,
together: `USAGE` on `audit`, SELECT of `id`, `occurred_at` and `organization_id`, DELETE, and the policies
`events_retention_select` and `events_retention_delete` `TO <db>_retention` only, which admit exactly the rows the
transaction's own declaration makes eligible (`app.retention_eligible`: scope, a cutoff strictly in the past and holds on
the row's Organization, Group, actor, subject and resource); and an `events_refuse` exception that returns the row only
for a DELETE by that role of an eligible row. Every other UPDATE, DELETE and TRUNCATE by every role, the owner and
superusers included, still raises append-only. B-AUDIT-CATALOG and B-AUDIT-MIGRATION keep their exact version-8 meaning
on the migration set through 00008; B-RETENTION-CATALOG and B-RETENTION-IT prove the version-9 delta with the actual
retention login. No audit action was added: backup, restore, replay and maintenance evidence is the operator receipt, and
a replayed disablement's audit gap is recorded through the existing `revocation.unaudited` action
([RECOVERY.md](RECOVERY.md)). The 29-action catalog is unchanged. Legal holds and periods are owner decisions.
