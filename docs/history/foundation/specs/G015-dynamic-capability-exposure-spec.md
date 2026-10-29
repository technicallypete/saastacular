> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G015 — Dynamic capability exposure work specification

Goal: [G015](../goals/G015-dynamic-capability-exposure.md)

Epic: EPIC-005 — MCP control plane

State: Definition ready; automatic provisional planning at G014 implementation MR creation; execution gated

## Purpose, authority and baseline

Implement the discovery/staleness/concurrency portion of
[ADR-005](../../../adr/ADR-005-tenant-bound-mcp-connections.md) and the
[connection-context contract](../../../../contracts/mcp/connection-context.md).
G014 owns external authentication, connection lifecycle and the five controls;
G015 composes their accepted boundary with current capability discovery and
independent invocation authorization. G021's product barrier remains closed.

Owner “Do it” (2026-10-15) authorized this definition. The owner's later automatic-
workflow clarification removes the extra G015 approval hold: automatically begin
provisional planning when G014's implementation MR first opens (draft included),
without waiting for CI/Duo or merge; reassess after its final
ACCEPT, and release execution only when the normal review/dependency/input/budget/
resource gates close. Explicit owner stops still override automatic continuation;
this adds no G016+ automatic-specification mandate.

Inspected integration checkout: `staging` at
`d1dc1080a3c190e6968c61d6d41e69ccad3efe10`, 2026-10-15 UTC. The dormant
`core/internal/mcp/server.go` constructs an empty official SDK server; `core/go.mod`
pins `github.com/modelcontextprotocol/go-sdk v1.8.0`. G014 is not implemented or
finally accepted. Its r5 Plan at `85d6993` was an unaccepted proposal, not an
execution input or proof of delivery. This is a provisional specification:
reconcile actual integrated/accepted G014 before approving a G015 Plan/execution.
Do not reserve a migration or treat proposed predecessor APIs as shipped.

Canonical inputs: AGENTS/CLAUDE, Architecture §52/§53.9, SECURITY, DECISIONS,
ADR-003/004/005/006, ROADMAP/PROGRESS, G014 Goal/Spec and eventual accepted
Plan/Result, foundation AUTHORIZATION/DATABASE-ISOLATION/AUDIT/IDENTITY and
PERFORMANCE. Architecture and accepted decisions win over historical examples;
unresolved contradictions escalate.

## Prerequisites and retained boundaries

G014 must supply its accepted authenticated SDK HTTP transport, exact registered
endpoint/resource and local user/client/Organization binding, fresh request scope,
connection-local Group/lifecycle state, reviewed recovery and mutation ordering,
and qualified WorkOS contract. Final G013 recovery and accepted G004–G012
controls remain inherited prerequisites. G015 does not redo provider experiments,
revocation qualification, setup UI, Group controls or hostname provisioning.

WorkOS remains the global identity and credential issuer under owner option B.
Independent local selection/lifecycle/catalog state does **not** mean independent
provider refresh families. Direct A-audience bearer/session replay at B refuses;
a separately issued exact B-audience token still needs B's own local subject,
client, Organization and liveness agreement. No credential-level isolation promise,
global logout behavior or provider change is introduced here.

## Required behavior

### R1 — Finite permission-filtered discovery

`tools/list` authenticates and validates the exact connection before returning
private tool metadata. Its result is the permitted catalog for one coherent
connection context and current authority, with no foreign-tool metadata leakage.
Unknown capability mappings fail closed. Tool visibility follows accepted core
policy, including Organization-only versus selected-Group inheritance, existing
roles, profile links and the current Organization policy flag. An argument,
client catalog/version marker or model assertion cannot supply authority.

Production exposes only G014's five approved controls:
`account.get_current_user`, `account.get_current_context`, `context.list_groups`,
`context.switch_group` and `logout`, under their accepted eligibility rules.
These may share eligibility for a valid connection; do not add production tools
merely to create different role catalogs. Demonstrate generic capability mapping
with a small set of test-only in-process tool fixtures using existing capabilities
and real policy facts. They compose only in tests, never production route
registration, configuration-driven arbitrary discovery or a plugin registry.

Pin how tool-level eligibility handles capabilities requiring exact targets;
visibility never grants every argument/resource accepted by a schema. Keep
production schemas/descriptions bounded and non-sensitive. The finite five-tool
catalog needs no pagination. If pagination or catalog limits are actually added,
the Plan must cover bounds, cursor/context freshness and complete authorized
enumeration; no generic cursor framework is required.

### R2 — Freshness without reconnect or token refresh

Successful select/clear invalidates the targeted connection's prior catalog
before its next request can consume that catalog. Failed selection preserves the
previous selection and its still-valid catalog state. A relevant committed
membership/role, Group–profile association, policy, user/session or connection
lifecycle change affects later protected calls and lists without reconnecting or
refreshing credentials. Grants as well as removals appear on the next permitted
list; restoring a role cannot revive terminal/expired authorization.

An invalid selected Group must not silently clear or fall back to Organization
authority. Preserve G014's exact recovery/select/clear rules under current
Organization authorization and its narrow same-binding termination exception
after membership loss. Neither catalog filtering nor a generic guard may make
those accepted recovery paths unreachable. Otherwise expired, revoked, disabled,
membership-lost, unknown or inconsistent contexts fail closed without private
metadata, protected data or effects. Store/verifier failure cannot authorize from
old catalog state.

Use a fresh request scope and current accepted core policy for discovery and
execution. Store only necessary non-authoritative connection/catalog metadata;
no persistent TenantContext, request envelope, full role/fact/allow cache or Redis
authority copy. The Plan chooses the smallest supported freshness mechanism.

### R3 — Independent direct-call and effect authorization

A hidden direct `tools/call`, including a previously visible tool now absent,
returns the same non-disclosing external tool-unavailable behavior as an unknown
tool and never invokes its handler. Public error details cannot expose hidden
schema, target or membership facts. A caller need not list first, and an earlier
list/version never authorizes invocation. Visible tools still perform the exact
guarded operation's current resource/capability checks and preserve inherited
non-disclosing error semantics. Filtering is no substitute for effect authorization.

Retain G007/G008 snapshot and write-linearization semantics: list/read requests
starting after a relevant authority change commits observe it. An already
authorized in-flight read may complete on its accepted snapshot. A guarded write
orders before the conflicting authority change or observes it and denies; there
is no adapter-check-to-effect gap. Do not promise cancellation of authorized
reads or undo of committed writes. Reuse existing RLS, lock order, audit,
unknown-commit and retry contracts; extend only a demonstrated adapter gap.

### R4 — Coherent concurrent and independent connection state

Pin observable ordering for concurrent select/clear, list and call on one logical
connection. A result uses one valid selection/context revision or fails safely;
it cannot combine one Group's catalog with another's authority, publish an old
computed catalog as the new selection, or use a stale selection to authorize an
effect. Reuse G014's control-mutation ordering and G007 guards; no new generalized
concurrency system or serializable-read promise is implied.

Two distinct connection URLs for the same user/Organization and a connection to
another Organization retain independent selection and catalog state. The same
logical URL continues to share its own G014 state across supported transports;
an SDK session identifier is not a new authority boundary. Browser navigation,
reconnect/resume and ordinary credential renewal cannot copy another connection's
context or resurrect a terminal record. Global authority changes legitimately
affect each connection's next request under its own current facts.

### R5 — Supported notification and re-list contract

Inherit G014's accepted protocol revision and SDK; no protocol upgrade is required.
Use supported `notifications/tools/list_changed` behavior for live sessions where
applicable, advertising `tools.listChanged` truthfully. The Plan pins delivery,
coalescing, observation timing and transport-loss behavior with a deterministic
testable rule. Notification payloads disclose no catalog, hidden metadata or
authority; a notification is a hint to re-list.

The bounded minimum is request-bound observation: local successful select/clear
invalidates before the next request; committed external authority changes are
observed on the next protected request/list. When an effective catalog delta is
observed, notify a connected session supporting the reviewed delivery path under
the pinned rule. No detection or push while an idle connection makes no request
is promised. Every subsequent permitted list is fresh independently of whether a
client receives/acts on the hint; disconnected/no-delivery cases recover by
supported re-list/reconnect. No background polling, event bus, global push-latency
SLO or client UI guarantee. Do not claim unchanged effective catalogs changed.

Protocol references for Plan verification: official
[MCP tools specification](https://modelcontextprotocol.io/specification/2025-11-25/server/tools)
and installed [Go SDK server](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/server.go).
Pin the actual inherited negotiated revision and SDK behavior during planning.

### R6 — Minimal integration, privacy and query assessment

Core owns discovery policy and execution guards; web/intelligence must not copy
authorization. Preserve production import/runtime-exposure boundaries and accepted
MCP audit/recovery behavior. Reuse existing decision/lifecycle audit mappings;
catalog hints alone do not require a new durable event stream. No credentials,
provider/refresh values, hidden catalogs or sensitive role/profile facts in logs,
notifications or management evidence. Test fixtures must stay outside production.

Assess actual new query/round-trip, set-size, fan-out and lock changes against
[PERFORMANCE](../PERFORMANCE.md), including relevant PG-001/PG-006
limits. Reuse existing small fixtures for bounded observations only where a new
query/ordering path needs evidence. Flag material cost/locking uncertainty for
review; no benchmark target, upfront index/schema change or authorization cache.
If accepted existing primitives cannot meet this Spec, report the smallest gap
before expanding scope.

## Acceptance criteria and verification

| ID | Required outcome | Objective verification |
|---|---|---|
| AC1 | Production discovery returns only the five approved controls eligible under the current exact connection, with no test-only or unknown capability exposure. | Real SDK HTTP list asserts exact production names/metadata and authenticated eligibility; source/import review proves test-only fixtures cannot enter production registration. Missing/invalid connection authority reveals no private catalog. |
| AC2 | Generic capability filtering follows existing Organization/Group/role/link/policy semantics. | Reused matrix expectations plus test-only existing-capability fixtures demonstrate allowed/hidden differences for Organization-only, selected Group, role downgrade, profile link removal and policy off/on; no new capability/role or production tool. |
| AC3 | Successful select/clear refreshes only its logical connection; failed selection preserves prior state. | HTTP list before/after authorized select and clear matches expected state; denied/foreign/unknown selection changes neither selection nor valid catalog. Exact G014 recovery/clear and termination exceptions remain reachable. |
| AC4 | Committed authority/lifecycle changes govern the next list/call without reconnect or refresh. | Same logical connection and valid provider credential across role/membership/link/policy changes and expiry/revocation/disable: next permitted list reflects grants/removals, invalid contexts deny with no fallback. Store-failure case never serves cached authority. Re-grant cannot revive terminal state. |
| AC5 | Hidden, guessed and stale direct calls cannot execute or disclose hidden metadata; visible calls independently guard their exact effects. | Real SDK HTTP hidden/unknown/stale-call negatives have the pinned indistinguishable tool-unavailable envelope and zero handler/effect; paired authorized call succeeds. Wrong target or lost capability refuses through current accepted guards despite earlier listing. |
| AC6 | Concurrent select/list/call obey the declared coherent-context and effect-ordering rule. | Bounded deterministic barriers demonstrate a selection/catalog race, a stale publication race where relevant, and authorization-change versus invocation ordering; resulting catalog/context and effects match the Plan's allowed outcomes. Reuse unchanged G007 read/write proofs; add database interleavings only for actual adapter gaps. |
| AC7 | Independent local connections never share selection/catalog authority. | Two distinct A connections and B for one user retain their own expected context/catalog after targeted selection, browser navigation and supported reconnect/renewal; same-URL reuse follows G014. Exact endpoint/audience/local binding regressions remain green; no refresh-family independence claim. |
| AC8 | Catalog-change signaling and re-list work under the pinned supported protocol. | Real SDK HTTP evidence for local delta, externally committed delta observed on a later request, no effective delta and disconnected/no-delivery re-list. Assert truthful capability advertisement and pinned notification/coalescing rule; fresh re-list succeeds without client-UI or idle-push assertions. |
| AC9 | Security/audit/recovery and production boundaries remain intact with a bounded query footprint. | Actual diff and focused inherited regressions show fresh request scopes/guards/RLS, recovery and logout exceptions, sanitized errors/logs/notifications and no production fixture registration. Record actual query delta, applicable bounded observations and escalation disposition; no speculative schema/index/cache. |
| AC10 | Required evidence, integration and cleanup are attributable and complete. | Result maps AC1–AC9 to exact tested/reviewed source and relevant content/input provenance, command outcomes and inherited evidence; required accepted path-selected local/MR/staging gates and scoped cleanup receipt are recorded. Codex final ACCEPT remains separate from submission or merge. |

## Bounded validation and provenance

Use five scenario groups in the existing fixture/matrix: (1) finite catalogs and
policy-derived test fixtures; (2) select/clear and committed authority changes;
(3) hidden/stale direct-call denial and effect guards; (4) connection isolation,
reconnect and deterministic ordering; (5) supported notifications and re-list.
Combine overlapping cases rather than taking a Cartesian product of every role,
tool, change and connection. Unit checks cover pure mapping/filter outcomes;
real SDK HTTP checks cover adapter composition; targeted real PostgreSQL ordering
checks cover only new authority/effect gaps. Mocks alone do not establish adapter
authorization, while notifications do not require a second full stack per case.

Reuse G014's provider/HTTP fixture and G007–G013 accepted policy/RLS/audit/recovery
evidence with exact unchanged content provenance. No repeated WorkOS browser
qualification for this catalog-only increment, no new oracle/general harness or
large permutation campaign. If auth/provider behavior changes, stop for review
rather than treating this reuse as proof of the new behavior.

Apply accepted Git-first reconciliation, batch sequencing, input bindings,
checkpoint publication, cleanup and GitLab branch/CI trigger profiles by reference.
The Plan maps actual changed paths to required checks; a full profile is required
when the accepted path map selects it, not solely because this Spec has ten ACs.
No repeat of unchanged passing suites for documentation/SHA-only advances; keep
reused evidence distinct from newly run evidence. Bind exact relevant source,
configuration, fixture and protocol/SDK inputs through the accepted procedure;
do not create frozen management-document hash manifests or another registry.
Retain failure/interruption history and exact resource ownership; close owned
resources under bounded cleanup and preserve uncertain/shared items explicitly.

## Plan decisions, permitted deviations and exit gates

The concise Claude Plan must identify the actual G014 delivery and any changed
assumptions; the closed tool-to-existing-capability mapping and target eligibility;
the smallest freshness/catalog-marker strategy; selection/list/call ordering;
the supported notification capability, observation/delivery/coalescing rule; the
five scenario groups' minimum checks and actual query/CI path delta. These are
implementation proposals for review, not new owner provider-choice questions.

Internal factoring, non-authoritative marker representation, no-pagination for
five controls, and test-only fixture composition may vary within these outcomes.
No production tool, new capability/role, weakened authority check, changed G014
recovery/logout exception, provider contract, major dependency, schema/index or
infrastructure expansion is a permitted deviation. Material changes return to
Codex and owner-owned decisions escalate with preserved history.

Before execution: actual required dependency ACCEPTs and G014 integration/
reconciliation are recorded; Plan ACCEPT and current
input/budget/resource release are complete. Before final acceptance: all ten ACs,
actual diff review, applicable local/MR/staging disposition and cleanup are
complete with no unresolved security/architecture findings. B7/checkpoint4 and
G021 remain binding. Planning uses the first verified implementation-MR trigger,
and no implementation is released by the existence of this Spec alone.

## Revision — automatic continuation clarification, 2026-10-15

The initial `9396bdc` definition's extra owner phase hold is superseded by the
owner's confirmation of the standing automated workflow. Planning may start
automatically at G014 integration while predecessor acceptance is pending;
execution requires final acceptance and the recorded normal release. Scope,
AC1–AC10, five focused scenarios and B7/checkpoint4 are unchanged. The inspected
r5/baseline description above is historical; current readiness includes accepted
r7 provisional design and QP-B provider qualification, not delivered G014 code.

## Revision — earlier automatic planning trigger, 2026-10-15

Owner clarification supersedes the previous integration-only trigger: begin the
G015 provisional Plan at first verified G014 implementation MR creation, including
a draft with implementation content. Record MR/source/target/candidate, reuse its
existing planner and do not wait for CI/Duo/merge. Planning-only/placeholder/probe
MRs do not qualify. Reconcile relevant changes and final accepted G014 before
execution; all normal acceptance/release gates and owner stops remain.


## Supervisor clarification — G015 Duo target-existence note (2026-10-16)

For capability-mapped discovery, Organization owner/admin paths are role-level
hints after the current context gate: they do not assert an existing Group or
BusinessProfile target. Owner-only privileged-profile eligibility remains owner-only.
These names may be listed in an otherwise empty Organization. Group-derived paths
still need their applicable current candidate/link/policy facts. Discovery never
grants an effect, proves target existence or replaces exact-target authorization;
unknown/foreign targets retain the accepted non-disclosing refusal at call time.
This pins R1's target semantics consistently with the accepted Organization-level
path without new queries, target-enumeration claims or production tools. Add a
literal empty-Organization case and qualify contradictory "some target" wording.
