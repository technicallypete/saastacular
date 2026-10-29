> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G020 Work Spec — Generic labels and visibility

> Current phase: provisional Claude planning authorized, 2026-10-20; see the superseding owner amendment at the end. Implementation remains held.

Definition draft, 2026-10-18 UTC. Owner authorized Goal/Spec only; no Plan dispatch,
implementation, migration reservation or qualification run. Read with
[Goal](../goals/G020-generic-labels-visibility.md), Architecture §53.8 and current
accepted authorization, audit, RLS and MCP contracts. Later Plan reconciles the
accepted G019/G058/G055 source; no pending implementation is treated as accepted.

## 1. Scope and model

A LabelDefinition belongs to exactly one Organization and one owner scope:
Organization (owner ID = that Organization) or Group (a Group of it). It contains
canonical ID, bounded nonempty display name and closed visibility. An assignment
references one definition plus a canonical target type/ID; enforce uniqueness of
(label, target type, target ID) and matching Organization. Owner scope and target
are immutable; moving either requires an independently authorized operation and
is outside this Goal. Reject ambiguous polymorphic selectors/unknown enum values.

The Plan proposes concrete name normalization, length/list limits, scoped name
uniqueness/collision behavior, IDs/indexes and exact errors using existing repo
conventions. Apply bounds server-side; no unbounded lists or SQL identifiers from
input. Names are text, never HTML. Label metadata grants no membership or access.

Initial production target: **Group**, only for an Organization-owned definition in
that same Organization. Target existence and current target authority must be
verified before assigning/removing/reading. Do not label Organizations, users,
shared Business Profiles, objects or future product records in this Goal.

Group-owned definitions are supported now. Their future assignments go through
a closed typed core target-validation contract: target type, canonical Organization,
exact owning Group, current caller/context authority and existence. A registered
adapter must derive these facts from authoritative storage in the same transaction,
not caller-supplied ownership or a preflight cache. Unknown/unregistered target
refuses. No Group-owned production assignment type is initially registered:
prove the extension with isolated test-only target fixtures, never reachable by
HTTP/MCP or production startup. Document this intentional limit in UI/API/Result;
do not present a fake working product-object label picker.

## 2. Authority and visibility

Reuse current RoleBindings/current context, with finite label capabilities mapped
in the Plan. No second authority store. Organization owner/admin may manage
Organization-owned definitions and assignments on their authorized Groups.
Group owner/admin may manage definitions owned by their exact Group and future
registered targets there. Owner direction, 2026-10-18: Groups are labeled by
Organization admins; Group objects are labeled by that Group's admins. Owners
have their corresponding admin powers. Organization-level administration alone
is not an additional Group-owned object-label management path; require a current
owner/admin binding in the exact owning Group for that path. Respect selected-
Group context limits. Ordinary members
cannot create/edit/delete/assign labels or change visibility. A Group role alone
never admits an Organization member or grants Organization label administration.

Visibility is a closed two-state semantic contract; the Plan may use repository-
consistent enum spelling. Group-owned private-by-default is a conservative draft
choice, to confirm at Plan review; no accepted broader reader policy is overridden:

| Setting | Read outcome |
| --- | --- |
| Owner-scope private (default) | Only authorized label managers of that owner scope; ordinary Group members cannot read Organization-private labels. |
| Target members (explicit opt-in) | Readers must also hold current read authority on the exact target. This option applies only to Group-owned definitions/assignments within that Group's current readers and valid registered targets. Organization-owned labels on Groups do not permit target-member visibility. |

Visibility never widens target access. Reading Group-owned opted-in definitions
requires current read access to that owning Group; assignment reads additionally
validate the registered target. A member of Group A gets no Group B label metadata
because a definition is shared, names match or profiles are linked. Organization-owned definitions and assignments are always Organization-admin/owner
private: reject target-member visibility for that owner scope. Group owners/admins
and members cannot discover labels on their Group through their Group roles.
A subject who independently holds Organization admin/owner authority may read
through that Organization role; do not confuse this with Group-role access. Managers may list their authorized scope's definitions normally.

Apply filtering to names, IDs, assignment lists, pagination totals/cursors,
filters, search and error output. Foreign/unknown/inaccessible identifiers must
follow the existing non-enumerating contract. Visibility narrowing, role removal,
departure and session revocation take effect on the next request/tool execution;
no stale cache/catalog grants. Authentication/applicant credentials alone grant
no label access. Test genuine elevated authority separately from Group-member
privacy cases.

## 3. Persistence, consistency and audit

Plan one bounded migration against the actual free version at planning time.
Reuse the approved runtime/owner role split, FORCE RLS, current context and narrowly
scoped existing helper conventions. No PUBLIC grants, runtime DDL/BYPASSRLS or
unreviewed SECURITY DEFINER owner/helper. Enforce valid owner scope, matching
Organization, foreign keys where possible and uniqueness in storage. RLS must
protect both definition and assignment reads/writes, including private labels;
authenticated application-only checks are insufficient.

Application authorization, target validation, writes and mandatory audit commit
atomically under the existing transaction/lock order. A racing target deletion,
authority change or duplicate assignment cannot leave an orphan/cross-scope row
or report success after rollback. Specify actual locking/isolation guarantees,
FK/deletion cleanup, label-deletion assignment cleanup and downgrade refusal
where rows/new audit catalog values prevent safe down migration. No production
data repair or irreversible blanket cascade across unrelated entities.

Use existing audit infrastructure for successful definition create/update/delete,
visibility change and assignment add/remove, and existing denied-action semantics.
Plan exact minimal catalog actions, shape/attribution, counts and changed consumers
before coding. No label names or arbitrary target content in logs. Preserve
existing transaction outcome/error contracts: unknown COMMIT is unavailable,
not a claimed rollback or blind compensating write. Repeated assignment/removal
must have explicit idempotent/conflict behavior without duplicate state or audit.

## 4. Interfaces and UI

Core owns policy, target validation, transactions and label rules. HTTP and MCP are
thin context-authenticated adapters with the same finite capability policy and
closed target registry. Update contracts/catalog schemas and context eligibility;
hidden tools must reject direct execution. Preserve tenant-host/session/CSRF,
input validation, trusted context and secret-safe error/logging boundaries.

Provide minimal label administration/assignment on existing Organization Group
management views, plus Group definition management where currently authorized.
Read-only opted-in Group-owned labels may appear on authorized target views; Organization-owned labels must remain absent from Group-role views. Private labels
must not be serialized into member HTML/JSON and hidden with CSS. Empty/unsupported
Group-owned target assignment stays explicit. Use current approved Vestive skill
location/components, small existing controls and accessible light/dark layouts;
no new dashboard, separate product design or taxonomy builder. No public signup
or live provider action. Any owner visual check uses one bounded coherent preview.

## 5. Acceptance and required evidence

| AC | Required outcome and focused evidence |
| --- | --- |
| AC1 — Model/lifecycle | Both owner-scope definitions; scoped uniqueness/bounds; duplicate assignment/removal and deletion semantics; migration catalog/up/down/refusal evidence on real PostgreSQL. |
| AC2 — Current authority | Owner/admin positives and ordinary-member denials across Organization/Group scopes, selected context and direct HTTP/MCP invocation; ordinary membership/label never grants elevated authority. |
| AC3 — Privacy | Default Organization-private labels absent from member payloads, lists/counts/filters/errors; Group owner/admin/member roles cannot see Organization-owned labels, and attempts to opt them into target-member visibility refuse; Group-owned opt-in intersects current target read; narrowing/revocation takes effect immediately. |
| AC4 — Target boundary | Same-tenant Group positive; foreign/unknown/cross-Group/unsupported substitutions rejected; Group-owned test adapter proves real owner+target checks/transaction semantics without registering a fake production type. |
| AC5 — Storage isolation | Runtime/RLS direct-SQL owner-scope/visibility negatives and paired positives, no cross-tenant/orphan assignment; approved grants/helper inventory and migration legacy-path checks. |
| AC6 — Atomicity/audit | Audit failure rolls back mutation, concurrent duplicate/delete/authority cases meet actual guarantee, correct attribution/action/shape and unknown-COMMIT handling; catalog parity and historical consumers updated. |
| AC7 — Surfaces | Minimal accessible existing-shell UI, HTTP/MCP parity and hidden direct-call denial; parser/bounds/CSRF/context and member payload checks; focused light/dark/mobile evidence for changed UI. |
| AC8 — Delivery/workflow | Current-head required MR/Duo/staging disposition, all new rows actually selected by hosted harness, scoped resource cleanup and final Codex ACCEPT. Record G055 consolidated entrypoint/discovery/role/handoff/cleanup use in Result without requalifying G055 or claiming the G021 barrier passed. |

Plan maps every AC to named executable selectors and expected meaningful outputs.
Inventory changed migration-version, action/reason/operation, fixture, row-count,
logrecord, contract and disabled-variant consumers before the first batch. Use
change-impact selection and existing test/attack/oracle fixtures, not a new suite
family or exhaustive role permutation. Run focused local checks; required hosted
CI remains. Preserve first failures, diagnose the affected chain and continue only
safe independent work. Reuse unchanged predecessor evidence with exact source
attribution; no full-suite rerun for documentation or SHA-only drift.

## 6. Readiness, dependencies and phase holds

G019/G058/G055 final ACCEPT/integration and actual source reconciliation precede
execution; G055 final ACCEPT also gates planning under the owner's specific hold.
G057 is not a G020 execution dependency but remains mandatory before G021.
This owner instruction is **specification only**, not a broader roadmap mandate
or authorization to dispatch Claude once the document exists. Codex must record
later applicable phase authority, Plan ACCEPT, current input/scope/budget/resource
binding and release. Provisional B5/checkpoint3; one implementation writer,
300k checkpoint/350k safe handoff, event-driven concise updates. Preserve earlier
phase and owner-held MR integration boundaries. No new infrastructure, skill
installer, email delivery, product-specific label policy or G054 work.

At Plan review resolve exact capability/catalog names, adapter transaction contract,
list bounds/normalization, deletion/rollback handling, minimal UI placement and
focused selectors. These are bounded design choices within this definition;
any proposed different authority/visibility model, real extra target type or
security-model change requires escalation rather than silent Plan expansion.

## Definition review and validation — 2026-10-18 UTC

Codex reviewed this bounded definition against Architecture §53.8, current
AUTHORIZATION, audit/RLS conventions and the foundation barrier. Bounded read-only
architecture assistance confirmed the initial Group target/future-target adapter
boundary and warned against treating shared Business Profiles as Group-owned.
Root owns this draft disposition; this is not Claude Plan or execution ACCEPT.
Group-owned default privacy remains the explicit conservative draft choice above.
Local document links and whitespace checked; no application tests or pipeline
launched for specification-only publication. Reconcile actual accepted predecessors
and resolve the named Plan choices before any later execution release.

## Owner privacy clarification — 2026-10-18 UTC

“Group owners admins don't see group labels” means Organization-owned labels
**on Groups** are private to Organization admins/owners. This supersedes the
draft option allowing explicit target-member visibility for Organization-owned
labels. Neither Group ownership nor Group administration exposes these labels,
definitions, assignments, counts or filters. Only an independently held current
Organization admin/owner role qualifies. Group-owned labels on Group objects
retain the separate scoped authority/visibility rules; no phase release follows.

## G060 dependency reconciliation — 2026-10-19

[G060](../goals/G060-explicit-group-authority.md) is a required accepted predecessor before G055
completion and this task's planning. Reconcile its metadata/content separation:
Organization staff may manage Organization-private Group labels under existing
Organization authority; Group-owned object labels still require actual Group
roles and cannot inherit the ordinary Group-administration override. No G020
planning/execution release; existing eight criteria and privacy scope remain.


## Owner authorizes provisional planning now — 2026-10-20

Owner agreed “Ya i agree” to starting G020's provisional Plan now rather than waiting for G055 qualification. This supersedes this document's earlier G055/G060 acceptance-based **planning** hold and specification-only phase stop. Codex releases the verified existing coordinator to assign one planner (reuse any existing G020 planner; do not duplicate it). Read-only repository inspection and Plan/input/history artifacts only: no implementation, migration reservation, installs, builds, DB/browser probes or shared test resources.

Plan against current integrated G060 source and current approved G055 design, marking pending final qualification/discovery/cutover as provisional. Reconcile actual accepted G060/G055 and any later relevant deltas before implementation release; do not infer acceptance from their merges or from this planning permission. Preserve explicit Group authority, private Organization labels and actual Group-role requirements for Group object labels; G060's ordinary management override does not grant object-label access. Use current canonical local-after-merge qualification and permission/resource rules rather than older hosted-staging wording in historical AC8 notes. Keep G055 as the first normal workflow exercise when qualified, without requalifying it here.

G019/G058 are closed. G060/G055 final acceptance, reviewed Plan, exact input/scope/budget/resource readiness and supervisor execution release remain implementation gates. Provisional B5/checkpoint3 unchanged; planning consumes no implementation attempt. No new roadmap range, G061 dispatch, product feature or live deployment authority. Earlier holds remain as history, superseded only for this planning phase.


Current owner context amendment,2026-10-20:150k checkpoint/200k safe handoff supersedes earlier300k/350k task defaults at the next safe boundary. Existing planner/phase/budget/bindings and no-runtime planning restriction remain; no restart or additional review round solely for policy carry-forward. Coordinator `/compact` stays owner-only.


Owner superseding amendment,2026-10-20: Claude/Codex self-compaction at safe points is now authorized through supported native controls; the earlier owner-only rule above is historical. Keep150k/200k assessment, save state, preserve active batches/children/budgets and verify actual completion. If the live interface has no agent-accessible control, report the limitation once; no synthetic input or replacement-session workaround. G055 carries the policy into consolidated guidance/templates/export; no task restart or added validation campaign.


## MCP foundation clarification — provisional Plan review,2026-10-20

Current G020 preserves the established five production MCP control tools. Prove finite label-capability eligibility, current-context refusal and hidden direct-call behavior using existing test-only fixtures; this is not production label-operation delivery. A new production tenant-data tool requires separately reviewed public-contract scope. Core label rules remain authoritative and test adapters cannot reach production startup or HTTP/MCP. No change to accepted label privacy, target/RLS requirements or implementation holds.


## Owner scheduling amendment — G062 before G020 execution — 2026-10-20

Owner clarified the delivery sequence as G060 → G055 → G062 → G020 → G057 →
G021 → OS001 → OS002. G020's accepted Plan and technical requirements remain
unchanged. G062 final ACCEPT precedes G020 implementation release as an owner
schedule gate, not a newly assessed technical dependency. Before release,
reconcile G062's accepted source and G055's integrated state with the existing
Plan and exact bindings. No Plan re-review or replan is required solely because
of this order change.

## Owner scheduling amendment — G063 next before G020 — 2026-10-21

Owner now prioritizes [G063](../goals/G063-account-mcp-connections.md), the shared
My Account MCP manager and friendly names, next after G062. Current order is
**G062 → G063 → G020 → G057 → G021 → OS001 → OS002**. G063 final ACCEPT
gates G020 integration/final acceptance; isolated overlap needs normal supervisor
release.
This supersedes earlier immediate-G020 sequencing, adds a scheduling gate only,
and changes no G020 requirements or accepted Plan bytes. Reconcile inputs and
readiness before the existing supervisor execution release; retain prior history.
