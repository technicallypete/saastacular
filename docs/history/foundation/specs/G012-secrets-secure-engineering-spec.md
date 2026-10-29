> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G012 — Secrets and secure engineering controls work specification

Goal: [G012](../goals/G012-secrets-secure-engineering.md)  
Epic: EPIC-004 — Audit and compliance foundation  
State: r0 + privacy inheritance amendment — definition; predecessor G011 implementation active

## Rationale, authority and inspection baseline

Architecture §56 calls for secrets abstraction, access lifecycle controls,
encryption, dependency scanning and vendor/change evidence. Several already
exist. G012 closes specific gaps and describes operating responsibilities;
neither scanner success nor configuration declarations certify security.

Inspection baseline: staging `a65f5ae`, after final G010 ACCEPT/EPIC-003 closeout.
G011 Plan r2 (`72042f5c…`) and execution release are approved; its implementation
is not yet accepted. This is not an execution snapshot. Reassess the final G011
implementation, audit producers, migrations and review residuals before binding.

### Existing controls to reuse

- `core/internal/config/config.go`: SecretProvider/EnvSecrets and secret-value
  redaction; WorkOS environment/file loading and production identity refusals.
  Provider errors currently enter `Load` errors verbatim: sanitize this boundary.
- `intelligence/src/models/settings.py`: environment settings with secrets
  excluded from representation; web configuration is server-only. Preserve the
  existing service separation rather than distribute provider/datastore keys.
- `infra/security/gitleaks.toml`, Compose secret-scan, existing service lint,
  generated-secret canaries and CI publication/privacy controls.
- G004–G010 current-authority revocation/isolation; G011's final audit contracts
  after acceptance. No new user administration surface or role model is needed.
- `docs/foundation/VENDORS.md` already exists; extend it, not a second inventory.
  `docs/SECURITY.md` is currently empty and can hold the engineering policy.

## Scope and architectural/security invariants

In scope: requirements R1–R7 below. Production code changes are limited to
necessary current secret/configuration conformance; tooling and documentation
extend accepted validation paths. No migration or new durable entity is planned.
Unexpected schema needs return for scope review.

Core retains identity, policy and audit authority; web remains a thin BFF;
intelligence remains internal. UUID canonical identity, Organization/Group
isolation, derived BusinessProfile authority and current session checks remain.
Do not introduce product schema or features before G021. Optional MFA and
unknown per-session assurance remain unchanged (ADR-003).

No managed secret/KMS provider, runtime vendor, infrastructure deployment,
automatic rotation daemon, new secret HTTP/MCP endpoint, public contract or
general security/CI framework. G013 owns recovery, retention/deletion and incident
exercises. Container/OS vulnerability scanning and full Git-history remediation
are follow-up scope, not implied by application-dependency scanning.

## Requirements

### R1 — Current-secret inventory and safe resolution

Inventory current secrets and sensitive credentials by configuration **name**,
service/consumer, supported source, required/optional conditions, trust boundary,
reload/restart behavior, rotation owner and exposure restrictions. Cover database
role passwords, Redis/MinIO credentials, WorkOS key, internal service token and
optional gateway key; classify public identifiers separately. Include tool/CI
credentials in the operating inventory, never their values. Distinguish disabled
optional services from missing required credentials.

Extend/harden existing SecretProvider rather than create a second core resolver.
Demonstrate a replacement provider through the existing injection boundary, with
no caller dependence on provider-specific errors or types. Local environment
loading remains supported; retain WorkOS's existing environment/file contract.
The Plan identifies any additional file-source support actually needed; do not
invent a cross-language secret RPC service or require a cloud adapter.

Missing required, empty, denied, unavailable or ambiguous secret sources fail
closed before the affected client/service can serve privileged work. A provider
failure cannot silently fall back to stale/environment/default credentials.
Optional absence retains its documented disabled behavior. Diagnostics identify
only allowed configuration names/reason codes, never raw provider text, secret
values, credential paths or credential-bearing URLs. Returned partial config
cannot be mistaken for a valid credential set. Include newly loaded secret
material in existing redaction mechanisms. Secret acquisition stays at the
documented configuration/reload boundary, not on every request.

### R2 — Least privilege, rotation and deprovisioning

Map each current service/tool to its minimum database/storage/identity/internal
credential use, accepted grants and operator removal procedure. Runtime services
must not gain owner/migration/maintenance credentials; browsers/web bundles
must not acquire datastore, WorkOS or gateway secrets. Reuse existing isolation
tests/evidence; change grants only for a demonstrated reviewed defect.

Document and exercise **one** representative service credential rotation and
withdrawal in a disposable existing datastore/provider fixture. Record consumer
reload/restart, old credential denial on new authentication, new credential
success, and affected consumer recovery. State the fate of pre-existing
authenticated connections; if disconnect/restart is required, include that step
and verify the old connection can no longer operate afterwards. Removing a
source value alone must not be described as instant revocation of loaded values
or open connections. No real credentials/accounts are altered.

Human/tenant deprovisioning uses existing disable/session-revoke/role-removal
paths and current-authority checks, with a runbook mapping steps to G011's final
events. Reference accepted lifecycle coverage; rerun affected cases if G012
changes them. Preserve ADR-006: an audit-only failure must not undo committed
logout/disablement; no new general audit-bypass privilege or event catalog.

### R3 — Secret and static security checks

Retain existing Gitleaks and static language checks; close only demonstrated
coverage gaps. A local ignored runtime `.env` exception must not hide a tracked
secret-bearing file from the submitted-source scan. Pin the actual source set:
candidate tracked content plus relevant submitted additions; no ambient home
directory, live credential files or unrelated worktrees. Existing evidence scans
continue to cover outputs before publication.

A generated disposable credential-shaped canary in a normally scanned file and
in a tracked-file exclusion edge case must cause failure; an untracked ignored
local secret file remains outside submitted-source evidence. Keep canary values
out of reports. Scanner unavailable/error/invalid output is a failed/unmet check,
never an empty clean report. Reuse accepted output-publication failure tests;
do not build another scanner-of-scanners framework.

Retain Go vet, web lint/type checks and Python security-enabled lint. The Plan
maps actual security rules and coverage limits, adding a static analyzer only
for a named current-code gap. No general SAST rollout is required by this Spec.

### R4 — Dependency vulnerability policy and checks

Provide containerized, read-only vulnerability checks over the resolved Go,
npm and Python dependency graphs, including application and development/test
dependencies. Use authoritative lock/module inputs and report transitive
coverage, toolchain coverage and any unsupported category explicitly. Do not
claim an OS/container-image or full supply-chain audit. No package install
scripts, application startup, live provider requests or automatic dependency
fix command are needed to scan a graph.

The Plan pins scanner versions/images and advisory sources, maps tool-specific
result/exit semantics, and records scan time plus available database revision or
fetch timestamp. Advisory data changes independently of source: do not label it
byte-reproducible. Bound network timeouts; unavailable feeds, missing/invalid
inputs, truncated/malformed/incomplete results and scanner failures cannot pass.
Scanners receive no runtime/CI/provider credentials. Any transmitted data is
limited to the approved dependency metadata; no source, `.env` or evidence upload
to a new SaaS scanner is authorized.

Blocking policy: High/Critical known dependency findings block, as do known
reachable Go vulnerabilities regardless of missing severity. Unclassified known
findings require disposition rather than being assumed below threshold. Lower
severity findings are reported with owner/action or explicit reviewed follow-up.
Reachability can inform disposition but cannot silently omit the dependency
inventory. A finding exception requires a supervisor-reviewed exact advisory,
package/version, rationale, named owner, expiry/review condition and compensating
control; security-model/risk acceptance changes escalate to the product owner.
No blanket baseline, wildcard suppression, test-dependency exemption, automatic
upgrade or weakened threshold is preapproved. A stale exception cannot pass.

Verify each scanner's real successful invocation and representative finding
handling using small disposable known-input/advisory fixtures, not intentional
vulnerable application dependencies. Test clean/finding/tool-failure/incomplete
result behavior at the changed boundary; consolidate equivalent cases. New real
blocking findings must be corrected within reviewed scope or explicitly disposed
before ACCEPT; lack of a fix is not permission to suppress them.

### R5 — Encryption configuration and honest readiness

Maintain a finite matrix for browser→web, web→core, core→PostgreSQL, migration→
PostgreSQL, core→Redis, core→object-store/data and signed-download endpoints,
core→identity provider, core→intelligence and intelligence→gateway, plus persistent
PostgreSQL/object storage/backups at rest. Mark disabled optional paths explicitly.

For each: current actual mechanism, local exception, supported production
configuration, trust/certificate verification, owner and proof source or unmet
production prerequisite. Local Compose is a development fixture, not evidence
of production encryption. Current plaintext/implicit PostgreSQL, Redis and
internal HTTP paths must be identified. Existing object-store/WorkOS HTTPS
refusals remain enforced.

Add or reuse a bounded configuration/readiness check that refuses a
production-ready result for plaintext, certificate-verification bypass,
unsupported transport or missing required at-rest evidence. Unknown external
state stays unknown/blocked; a self-declared boolean is not verified encryption.
Configuration-only evidence must be labelled as such. Synthetic valid/invalid
fixtures test the checker; a valid fixture is not a deployed-platform attestation.
Development/test allowances must not carry into the production result.

This Goal implements checks and records responsibilities, **not** whole-stack
TLS, KMS, cloud disk encryption or production deployment. Those remaining
production prerequisites may stay blocked at G012 ACCEPT if fully identified;
that is no waiver for eventual deployment/G021 readiness review. New transport
support requires an approved bounded scope decision and certificate-chain/host/
untrusted-peer negatives. Do not silently repair every transport here.

### R6 — Existing vendor and change-evidence records

Update `docs/foundation/VENDORS.md` for actual runtime/test/CI/advisory-source
integrations. Record component/version source, purpose, service/environment,
data categories sent, credential owner, privilege, processor/license review
status, deprovisioning responsibility and production prerequisites. Use public
documentation/config references; no personal account/secret dump or invented
contract, residency, license approval or live-provider qualification.

Populate `docs/SECURITY.md` with this finite control policy, reporting/triage
route, finding/exception handling and links to operating procedures. Reuse Git,
MR/Codex/Duo dispositions, required CI results and G011 audit evidence; no
new evidence database, retention daemon or organizational compliance program.
G013 receives the encryption/credential/vendor matrix and audit maintenance
boundary for its recovery/retention/incident work.

Carry Architecture §56.1 (not included)
in these same records: identify whether an integration receives personal data,
its purpose/data categories, the operator responsible for controller/processor
classification, known processing locations/subprocessors, retention/deletion and
data-use/training terms, and transfer/agreement review status. Record unknown
or unreviewed matters as production prerequisites; do not invent legal approval
or require completed vendor negotiations to define the local foundation.
In particular, record identity/AI providers and advisory metadata transmission
according to actual current behavior, not imagined live traffic. Reuse the
existing inventory and AC8 documentation review; no second privacy registry.

G012 supplies these generic technical/operating hooks; G013 consumes retention/
incident responsibilities, and future G056 owns assembled-product rights work.
No rights portal, consent workflow, new personal-data schema or GDPR compliance
certification is part of G012. G056 is not a predecessor or new G012 execution
gate. Reusable foundation documentation must distinguish shipped controls from
adopter-owned legal/deployment and application-specific privacy work.

### R7 — Proportionate validation, provenance and cleanup

Use the accepted harness/report/CI entry points. The Plan maps each new check to
its owning suites and actual path trigger, avoids duplicate service scans in
both job partitions, and distinguishes source/security checks from CI meta-tests.
Run CI-controls only for relevant changed CI/harness behavior under current
selection; no additional unchanged-control campaign. Documentation-only updates
retain the existing pipeline bypass. Expected infra/config changes select full
under the current matrix; record the actual final selection, not an affected-only
promise. A bounded mapping edit is allowed; runner/profile redesign is not.

Develop with focused stack-free secret/scanner/config tests and the one rotation
fixture. Finish safe failing batches, fix and rerun the failed/affected batch,
then continue, reusing unchanged passing coverage. Required manual feedback,
if any, precedes final qualification. Fetch/reconcile before each batch using
the accepted delta rule; record source/target, inputs, tool/advisory versions,
commands, outcomes and evidence reuse. Do not repeat a broad suite merely to
obtain another all-green report or because a management commit changed SHA.

Result/CI artifacts and compact Git reviews are durable evidence; local bulk is
temporary. Record owned resource IDs at creation; remove owned throwaway
credentials/resources on normal and interrupted completion, with one compact
receipt and named uncertain holds. No global prune, unrelated-volume removal,
secret-bearing output or new approval/cleanup framework.

## Acceptance criteria and verification

| ID | Required outcome | Objective verification |
|---|---|---|
| AC1 | R1 inventory matches actual current secret consumers and provider substitution works. | Compare configuration/Compose/service call sites to the finite inventory; real EnvSecrets success plus injected replacement provider success through the same core loader; disabled optional paths labelled. No new runtime service/vendor or request-time lookup. |
| AC2 | Secret acquisition fails closed without disclosure. | Focused missing/empty/denied/unavailable/ambiguous-source tests, including provider error text containing a canary; invalid config cannot construct an affected client. Existing supported WorkOS file failures remain sanitized. Scan permitted output; a planted leak must be detected. |
| AC3 | R2 service credential lifecycle is demonstrated. | One disposable old→new rotation and withdrawal: new login works, old new-login fails, documented consumer reload/restart recovers; old open connection's actual lifetime and required termination observed. Scoped teardown receipt, no real credential changes. |
| AC4 | Least privilege and existing user deprovisioning remain intact. | Secret-to-service/grant matrix; unchanged accepted role/session/policy proofs referenced, affected tests run if modified. Browser/web secret exclusion and no runtime owner/migration credentials; runbook maps revocation to final G011 events and preserves audit-failure exception. |
| AC5 | Submitted-source secret/static checks retain coverage and fail closed. | Existing lint/secret checks run; synthetic scanned-file and tracked-exclusion canaries fail, ignored untracked local file not published/scanned as submitted source; scanner error cannot pass; no weakened lint or broad exclusions. |
| AC6 | All three resolved dependency graphs receive enforceable vulnerability checks. | Current-tree real scanner reports; direct/transitive and dev/test inclusion recorded, tools/advisory provenance pinned. Small finding/clean/error/incomplete-result fixtures exercise selected tool semantics; blocking and expired-exception cases cannot pass. All real findings corrected or explicitly disposed under R4. |
| AC7 | Encryption readiness describes and rejects unsafe production posture. | Complete R5 path/at-rest matrix; checker negatives for plaintext, verification bypass, unsupported/unknown requirements, and local-only exception in production. Synthetic qualifying configuration positive labelled configuration-only; no claim that current local volumes/links are encrypted. |
| AC8 | Vendor/security operating records and inherited boundaries are complete. | VENDORS/SECURITY/procedures cover each actual current integration, R6 privacy/data-handling fields and named unknowns/production blockers; compare to manifests/config and R1/R5 inventories. No new schema/provider, product surface, mandatory MFA, audit bypass or compliance/deployment claim. G013/G056 responsibilities and adopter-owned privacy work explicit, without making G056 a predecessor. |
| AC9 | CI/change evidence is attributable and coverage is bounded. | Plan/Result show owning suites and actual trigger/partition; required final local/MR/staging evidence under current policy, failure/error paths and output privacy preserved. No successful skip/incomplete scan, repeated unchanged qualification or general harness redesign. Exact source/target/tools/feed and failed/interrupted/reused evidence recorded. |
| AC10 | Review and closeout meet governance. | Result maps AC1–AC9 and deviations, B5/checkpoint-3 ledger, read-only security review and scoped cleanup receipt; Codex actual-diff ACCEPT plus required staging disposition; PROGRESS assessment and G013 handoff. G021 stays closed. |

## Expected artifacts, Plan gates and completion

Deliver the finite secret/privilege/encryption inventories and procedures in the
existing security/foundation documentation; update VENDORS rather than duplicate
it. Claude chooses the smallest file layout. Include narrowly changed resolver/
configuration code, pinned scanner configuration/invocations, focused tests,
minimum accepted harness wiring and a compact Result. No new snapshot registry.

Before Plan approval, resolve:

1. Actual accepted G011 interfaces/residuals and audit coupling; no speculative
   migration or event expansion. Any contradictory final delivery escalates.
2. Exact credential source/reload semantics and the single representative
   rotation fixture, including already-authenticated connections.
3. Scanner choices, supported lock/toolchain versions, sources/network behavior
   and severity/reachability mapping. Record any new scan baseline as unknown
   until the first authorized implementation batch; do not require pre-approval
   scans. That batch collects current findings. Blocking findings need bounded
   remediation or reviewed disposition before ACCEPT; scope-changing fixes stop
   for review. Read-only planning inspection is not authority for installs,
   scanner execution or credential-bearing probes.
4. Configuration-checker inputs and measured-versus-declared evidence rules;
   unsupported production encryption remains a named block. No production
   deployment/vendor choice is necessary to write the Plan.
5. Concrete changed files, AC mapping, actual CI selection and bounded timing;
   no duplicate broad validation or procedural framework as a prerequisite.

No deviation that weakens isolation, privacy, scan failure behavior, advisory
policy or production-readiness truth is preapproved. Minor file/tool choices
within these boundaries belong in the reviewed Plan. A finding waiver, scope
expansion or new vendor needs its stated review/owner decision.

Completion requires AC1–AC10, accepted diff/evidence, required staging disposition
and cleanup. G011 must be finally accepted and the Plan/binding released before
implementation. This defines G012 only; it creates no G013 Spec or application
implementation. Standing automatic planning applies after G011 integration.

## Tool semantics references for planning

Tool selection remains Claude's reviewed Plan decision. Distinguish report
contents from process thresholds: [npm audit](https://docs.npmjs.com/cli/commands/npm-audit/)
documents that audit-level changes its failure threshold, not report filtering.
[Go vulnerability management](https://go.dev/security/vuln/) distinguishes known
dependency vulnerabilities from reachable code. These references do not pin a
future executable or replace actual selected-version tests.

## Privacy inheritance amendment — owner direction, 2026-10-14

Owner asks that G012 recognize the privacy/GDPR direction and that Saastacular
retain reusable privacy foundations. R6 and existing AC8 now explicitly carry
privacy-relevant vendor/data-handling fields and unknowns in the same inventory.
No additional acceptance criterion, scanner, legal determination, schema, rights
workflow or predecessor is added. Ten ACs/B5, G011 dependency and Plan/binding
gates are unchanged; no G012 implementation/planning has begun.
