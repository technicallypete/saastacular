> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G009 Work Spec — Tenant-aware object storage boundary

Revision: r1, 2026-10-13 UTC
Status: G008 accepted/integrated; corrected Plan approval and execution binding pending
Goal: [G009](../goals/G009-tenant-object-storage.md)

## Rationale, baseline and authority

Architecture §40 (not included) places
binary payloads in S3-compatible storage and ownership/relationships in PostgreSQL.
An authorized UUID must resolve an exact stored object; storage prefixes confer
no authority. This Goal owns the generic foundation, before product upload workflows.

Inspection reference: clean staging `1c3bb0e9a04854e859e9ddee6db79daa79696c56`,
source inspection only. Core has S3 configuration and a HeadBucket readiness
probe, existing pinned AWS SDK v2/S3 dependencies and private-bucket provisioning
in `infra/minio/init.sh`. It has no owned ObjectStore or object metadata yet.
G007 is accepted. [G008 Plan r3](../plans/G008-postgresql-rls-plan.md) is approved;
its proposed RLS/role/catalog changes are not accepted implementation evidence.

Read canonical governance/security/architecture/decisions, accepted ADRs, budget,
roadmap/progress, G008's actual final reviews/Results and the foundation
[storage contract](../../../../skills/saas-foundation/references/storage-audit.md).
ADR-004/current G007 supersede historical BusinessEntity grants. Preserve
Organization→Group tenancy, UUID identity, no independent BusinessProfile roles,
opaque PostgreSQL-authoritative sessions, optional MFA and G021. Slugs, keys,
hostnames and profile links are never substitute authority. Verified uploader UUID
permits only R3's own-upload deletion within a currently authorized Group; it is
not a standalone access or ownership grant.

## Scope boundaries

In scope: five internal storage operations, minimal scoped metadata/RLS,
provider adapter, exact capability mapping, finite payload/signed-access bounds,
safe cross-system lifecycle, local MinIO evidence and minimum associated docs/checks.
No new public web/API/MCP route, attachment domain, onboarding, advisor workflow,
object listing UI, profile-owned object tier, live cloud deployment or durable
AuditEvent schema. Signed GET only; uploads pass through the guarded core boundary.
R2 compatibility is documented, not cloud-qualified here.

## Requirements

### R1 — Owned contracts and caller boundary

Core owns ObjectStore operations Put, Get, Delete, SignedGetURL and Metadata;
provider details stay behind its adapter. Exact Go package/interface/API signatures
belong to the Plan. The guarded service takes authenticated request/context and
canonical UUID selectors; it never accepts caller roles, allow decisions, bucket,
provider endpoint, raw key/path or a reusable authority snapshot. Raw adapter
access is restricted to this trusted boundary and separately classified readiness/
test setup. Preserve existing readiness and service import directions.

### R2 — Metadata, immutable identity and RLS

A new minimal stored-object relation has a canonical nonzero UUID, required
Organization, optional Group constrained by same-Organization composite FK,
server-generated opaque unique key, creator attribution tied to the authenticated creator and server-verified
byte/metadata facts. Caller-supplied attribution is rejected; creator identity stays
immutable and supports only R3's scoped own-upload deletion, without creating an
owner role or independent membership. Enough lifecycle state exists to prevent incomplete/deleted
objects being returned or signed. UUID, owning Organization/Group and key are
immutable; replacement/reparenting is excluded. The provider reference must be
checkably bound to this exact metadata UUID and ownership, so a forged metadata
key cannot redirect an otherwise authorized operation to another object. Creator
equality never bypasses current identity, membership, context or scope checks.
Group resources do not require a BusinessProfile ID.

ENABLE/FORCE RLS and exact command/column privileges use the actual accepted
G008 context/helper patterns. Metadata is protected tenant data, not an exempt
infrastructure directory. Raw runtime SQL without application predicates must
retain the action/scope boundary. Preserve privileged-role separation, exact
owner FORCE proof, invoker behavior, composite integrity and no runtime helper
membership. No copied effective-role/authority rows. Additive migration only;
old migrations stay byte-identical. Plan inventories all added functions/policies/
roles and lifecycle privileges before approval.

### R3 — Read and write action policy

Each operation authenticates/revalidates current session/user and trusted active
Organization/selected Group, loads exact metadata under RLS and authorizes that
object's owning scope before any provider effect or disclosure. Organization
objects read through `organization.read`; Group objects through `group.read`.
Selected Group restricts Group targets; Organization authority remains independent
as the accepted context contract permits. A profile shared with a visible Group
never reveals an object owned by another Group. Foreign/inaccessible/missing
object IDs have the same external category and cannot leak provider detail.

**Owner-approved matrix, 2026-10-13:** upload, read and delete are separate actions.

| Object scope | Read | Put | Delete |
|---|---|---|---|
| Organization | Current Organization members allowed by organization.read | Organization owner/admin only | Organization owner/admin only |
| Group | Current callers allowed by group.read for that exact Group | Organization owner/admin or that Group's owner/admin/member, within selected scope | Organization owner/admin or that Group's owner/admin may delete any object in scope; that Group's ordinary member may delete only an object they uploaded |

Member own-upload deletion requires the stored immutable creator UUID to equal
the authenticated actor UUID **and** current authorized membership/access to that
exact owning Group. Derive creator from the trusted actor at Put, not request input.
Another user's upload in the same Group is denied to an ordinary member. Being the
uploader cannot grant access to another Organization, sibling/unselected Group,
or a Group the user has left or been removed from. The same global UUID across
tenants does not waive this boundary. Organization-scoped objects retain the
admin-only mutation rule, including objects a former admin uploaded.

An admin/owner downgraded to ordinary member may retain own-upload deletion only
if current Group access remains; deleting others' uploads must deny. Revoking
Group access or the session removes the member's own-delete path too. No copied
uploader ACL, additional owner role or BusinessProfile authority store is introduced.
The Plan's independent action oracle and runtime RLS proofs cover this exact matrix.

Storage actions are an explicitly reviewed internal
mapping; do not borrow `organization.manage_members/manage_policy` for ordinary
object writes or silently change existing G007 capabilities. A key/prefix or
caller-supplied creator never grants mutation. Later product attachment/retention
workflows belong to their owning Goals and must explicitly reconcile any narrower
rules; this foundation adds no product routes, uploads UI or parent-domain schema.

### R4 — Payload and selector bounds

Validate UUID/scope, declared and streamed byte counts, metadata fields/content
length/type and finite operation deadlines before/through side effects. The Plan
pins positive configured maximum bytes and metadata limits/defaults, with exact
error behavior for absent/negative/overflow/oversized/inconsistent lengths and
unknown-length streams. Enforce a bounded reader; no unbounded buffering or
unlimited provider wait. Caller key/path/traversal/encoded separator attempts do
not reach the adapter. Content/type/name is untrusted data, never authority,
server-fetched URL or executable content. No file scanning/rendering workflow.

### R5 — Signed GET and endpoint identity

Issue a GET URL only for a current authorized READY object. Default TTL **60 s**;
request/config bounds **1–300 s**, with a hard maximum **300 s**. Reject invalid,
zero/negative/overflow and above-cap values; no caller-controlled method, key,
bucket or response headers. Bind exact key/method/host and safe download response
metadata. Treat the resulting URL/query as a bearer secret; never log or retain
it in repository/evidence. URL reuse within its TTL is permitted, not single-use.

Real consumer access must work from the documented local host/test context.
The existing `http://minio:9000` backend address is not assumed browser-resolvable;
pin trusted signing/consumer endpoint configuration that matches the request host/
port without rewriting a signed URL or trusting inbound Host/forwarded headers.
Loopback defaults and configurable identity/ports remain supported.

Revocation prevents subsequent issuance/guarded access after the pinned decision
point. A previously issued URL may remain usable until its expiry; an already
started download may finish after expiry. Do not promise immediate role-aware
URL invalidation. Test new-request expiry against real MinIO and record residuals.
[Signed-URL semantics](https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-presigned-url.html).

### R6 — Provider security and portability

Use the existing non-root core credential limited to its private bucket. Core
never receives MinIO root credentials; web/intelligence receive no storage secret.
Anonymous bucket/object reads/writes and bucket administration by the core
credential are denied; prefixes do not replace service authorization. No public
ACL, new bucket per tenant, cloud purchase or credential rotation is implied.

Preserve endpoint/region/bucket/path-style/credential separation. Document R2's
initial production role and required endpoint/region/signing/credential settings
behind the same interface, with actual compatibility limits from the
[R2 API reference](https://developers.cloudflare.com/r2/api/s3/api/). No live R2
access or full feature-parity claim is required. Local MinIO remains local-only
under the accepted vendor inventory; no production promotion of its image.

### R7 — Database/provider failure and interruption

PostgreSQL and object storage are not one atomic transaction. The Plan pins exact
states/transitions, success/unknown/error meanings and retry/reconciliation paths.
Do not make metadata READY or expose bytes/signing before the declared durable
upload outcome. Re-authorize any later publish/finalize after provider work;
revocation after preparation cannot publish through cached authority.

Deletion must prevent new guarded reads/signing once its logical deletion commits,
with physical removal/result separately truthful. Existing bearer URLs may still
work while provider deletion is pending, bounded by R5's issuance TTL; do not
claim metadata state invalidates them. Repeated delete/retry behavior is pinned.
A failed provider delete cannot be reported as confirmed physical deletion.

Handle SQL/Put/Get/Delete/metadata/cancellation and process interruption at each
cross-system boundary. Preserve uncertain-COMMIT semantics: unavailable may
accompany committed state. Never destructively compensate/delete bytes solely
because a commit reply was lost. Establish authoritative owned state before
compensation; failed compensation retains attributable, inaccessible pending
state and a bounded repair/retry procedure. No new background queue/daemon or
unbounded bucket scan. Cleanup targets only proven owned namespaces/objects.

### R8 — Current authority, ordering and cost

Pin authorization and metadata-state linearization for every operation against
accepted G007/G008 writers. Every new operation after committed role downgrade,
membership/session revoke, disable or scope mismatch must deny if its allow path
is gone. Define racing/in-flight reads/mutations and signed issuance honestly;
do not claim transfer already in progress can always be recalled.

Provider I/O must not silently keep an authorization transaction/row lock open
for an unbounded transfer. Pin finite deadline/connection/lock use and staged
reauthorization where needed, preserving ordered database decisions. Assess
metadata point lookups, indexes and helper/query fan-out under
[PG-003](../PERFORMANCE.md). Record concrete unexpected scans/locking/
provider delays and escalate before mitigation; no preapproval benchmark,
unrelated index tuning or invented numeric performance guarantee.

### R9 — Operational evidence and staged audit

Record bounded structured allowed/denied/error and signed-issuance events with
actor/object/Organization/optional Group UUIDs, action/outcome and TTL as relevant.
Keep URLs/signatures/credentials, raw provider errors, payloads, filenames/private
metadata and auth material out of logs/traces/errors/retained evidence. Use current
redaction/canary controls and an exact event schema/call-site inventory.

This issuance accounting is operational evidence. G011 owns durable transactional
AuditEvent persistence and completeness; do not create its schema or claim that
logs satisfy durable audit/compliance. Expose a documented integration point and
hand off issuance/failure categories without weakening G009's required evidence.

## Acceptance criteria and verification

| ID | Required outcome | Verification/evidence |
|---|---|---|
| AC1 | Five owned operations and guarded provider boundary satisfy R1 without a public/product surface. | Interface/unit/import/wiring inspection; raw adapter bypass negatives; route/schema comparison with baseline. |
| AC2 | Metadata UUID/scope/key/lifecycle facts satisfy R2. | Actual runtime constraint/immutable-column tests, same-Organization Group FK, direct-SQL creator/key-binding forgery negatives, duplicate-key behavior and independent catalog inventory. |
| AC3 | Metadata/Get/signing follow exact read capability and current context. | Independent Organization/Group role matrix; selected/foreign/sibling/shared-profile IDs; forged/stale/expired/revoked context; paired positives and no denied provider calls. |
| AC4 | Put/Delete match R3's separate actions, including member Group uploads and scoped own-upload deletion. | Independent role/scope oracle; member Put and own Delete positive; another uploader's same-Group object, foreign/sibling/unselected Group, revoked membership/session and Organization mutation deny; admin-to-member own/other distinction; forged creator/key/prefix attempts; unchanged DB/provider state and no denied I/O. |
| AC5 | New relation RLS/context/roles compose with accepted G008. | Direct runtime-login SELECT/INSERT/UPDATE/DELETE without app filters, empty/malformed/poisoned pooled context, FORCE owner positive/foreign denial, exact grants/helpers and deliberate missing-protection detector. |
| AC6 | Authorized local byte and metadata round-trips succeed. | Real private MinIO Put→Get→Metadata using scoped credential; byte/content-length/type comparison and exact object identity, not fake-only provider evidence. |
| AC7 | Delete/retry/missing semantics and denied-delete isolation are truthful. | Real physical deletion and ready-state removal; repeat/missing outcome; Get/sign afterward denied; foreign denial leaves both systems unchanged. |
| AC8 | Signed GET is exact-object, bounded and usable without anonymous access. | Real consumer URL fetch, 60-second default/300 maximum and invalid TTL cases; altered key/method/host/signature denial, anonymous denial, short real expiry and reuse-within-TTL behavior. |
| AC9 | Revocation/order/races meet R5/R8's stated guarantees. | Deterministic current-actor revoke/downgrade/disable/session-expiry before decision; pinned in-flight ordering; former URL residual verified through expiry with no new issuance. |
| AC10 | Input/stream/deadline limits reject invalid requests safely. | Invalid UUID/scope/key/path/metadata/length/TTL and oversized/unknown stream cases; bounded memory/reader behavior and no unauthorized state/provider change. |
| AC11 | Cross-system faults/unknown replies/interruption meet R7. | Inject faults at every pinned transition, actual DB/MinIO state, cancellation, ambiguous commit, compensation failure and retry/repair. Confirm possibly committed valid bytes are not destroyed; pending/deleting objects never become new guarded access. |
| AC12 | Migrations/persistence and inherited behavior remain supported. | Empty/populated upgrade, versioned exact catalog/grants, restart/persisted object round-trip; owned disposable rollback/reapply; old migrations unchanged and applicable auth/RLS/root/off/rename checks. |
| AC13 | Privacy/issuance records and provider documentation satisfy R6/R9. | Event schema/call sites; credential/content/signed-query canaries and planted-leak detector; scoped credential administration denial; safe consumer endpoint/R2/TTL/revocation/failure docs with no live-cloud claim. |
| AC14 | Final evidence/provenance/budget/cleanup/integration are complete. | Exact source/target/input/scope hashes, commands/results, real versus injected evidence, failures/unmet/interrupted history, B5/checkpoint3 and no owned leftovers; final actual-diff review and applicable staging disposition; G010/G011 handoff without G021 release. |

## Expected artifacts and validation strategy

Claude Plan: reconcile actual accepted G008, owner write disposition, exact
metadata/grant/function/operation inventory, state diagram, bounded settings,
endpoint signing model, action/error/linearization contracts, failure scenarios,
query/index cost and AC mapping. Minimum code/migration/tests/docs plus Result
and source/input inventories. Proposed `docs/foundation/OBJECT-STORAGE.md`
records the usable commands/settings and guarantees; exact module names belong
to the Plan. Reuse the existing AWS SDK unless a justified reviewed pin change
is needed; no new major vendor or framework.

Real PostgreSQL runtime login and scoped local MinIO are mandatory for happy path,
isolation, signed access, deletion and persistence. Pair every denial/fault with
a working authorized path and inspect both systems' owned state. Use spies/fault
hooks for deterministic unknown/error boundaries, not as substitutes for real
provider success or expiry. Setup uses separately identified owned privileged
fixtures; assertion paths never inherit root/owner credentials.

Design scenarios before approval; implement them during execution, not as a
parallel preapproval harness. Reuse wrapper/input/drift controls, fetch/rebase
before batches, shared-stack scheduling, mount safeguards and owned Docker/object
cleanup. No broad bucket deletion, session rebind or secret exposure. Required
local full and MR/staging selection follow actual paths/current governance;
metadata migrations/shared validation paths select full. A canceled staging run
is canceled, never proof that tests passed; failed required checks block acceptance.
No duplicate unchanged provider qualification, unrelated browser campaign or live
R2 benchmark. Required operator feedback, if proposed, precedes final qualification.

## Provenance, deviations and completion gates

Record exact final accepted G008 baseline/Results/deviations, committed inputs,
source/candidate/target/catalog versions, scope allow-list, nonsecret configuration,
provider/tool versions and artifact hashes. Separate real root/disposable/injected
results; retain failures, skipped/unmet and interrupted runs. Never commit signed
URLs, secret keys, payloads or private host/session records. Copy required evidence
out of volatile locations under existing cleanup procedures.

Permitted implementation choices: internal packages, opaque key layout, minimal
lifecycle representation and bounded normalization/retry design, if reviewed and
all criteria remain. No weaker matrix, RLS, bearer-TTL, privacy, current-fact,
unknown-outcome or cleanup fallback. New cloud/repair/public/product authority,
a new major dependency or incompatible G008 integration escalates.

Before implementation: G008 final ACCEPT/integration, actual-state reconciliation,
this owner-approved write matrix, explicit Claude Plan ACCEPT and exact execution binding.
After implementation: AC1–AC14, Result, actual-diff Codex ACCEPT and qualifying
post-integration disposition. G010 can then proceed under its own authorization;
G011 receives durable-audit obligations. No earlier product features through G021.

## Open decisions and revision history

- **Owner decision resolved:** members may upload within an authorized Group and
  delete their own uploads while Group access remains. Others' uploads and
  Organization mutations retain R3's owner/admin rules. See
  r1 disposition (not included).
- **Plan pins:** accepted G008 interfaces/policies/migration number; metadata
  lifecycle and ambiguous-commit repair; payload limits; endpoint model and
  authorization/lock linearization; TTL exposure and query/index/provider costs.
  These are not permission to weaken a requirement or start implementation.
- r0, 2026-10-13: drafted at owner direction while G008 execution preparation is
  pending. Existing application/Goal bindings and G054 manual hold unchanged.
- r1, 2026-10-13: owner approves Group-member uploads and own-upload deletion.
  Supersedes the pending conservative write proposal and creator-attribution-only
  wording, only for this bounded storage action. Exact r0 archive (not included)
  and its review remain historical. Fourteen ACs/B5 and G008/G021 gates unchanged;
  no implementation or Plan dispatch follows from this revision.

## Actual-state planning reconciliation — 2026-10-14 UTC

Provisional G009 planning is authorized using G008 merged source `81b186c`
(MR !23, correction candidate `d0e8d40`, Codex correction/source ACCEPT).
Migration `00006_row_level_security.sql` and DATABASE-ISOLATION.md now exist
in that source; inspect their actual catalog/grants/transaction patterns instead
of the older planned-only inspection baseline above. Required G008 staging
pipeline `2945531965` is still running at this release of planning authority.
The earlier baseline is historical; G008 final ACCEPT and actual-state
reconciliation remain prerequisites to G009 Plan approval/execution.
Requirements, fourteen ACs, B5/checkpoint 3 and G021 remain unchanged.
