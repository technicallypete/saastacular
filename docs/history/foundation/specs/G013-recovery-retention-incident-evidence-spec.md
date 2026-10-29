> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G013 — Recovery, retention and incident evidence work specification

Goal: [G013](../goals/G013-recovery-retention-incident-evidence.md)
Epic: EPIC-004 — Audit and compliance foundation
State: r0 — definition; actual G011/G012 delivery must be reconciled

## Rationale, baseline and authority

Backup files alone do not prove recovery: data, object bytes, privileges and
current restrictions must agree before restored access resumes. G013 provides
one bounded operational proof for the generic foundation, not production
infrastructure or a second validation framework.

Inspection baseline: staging `113dc0b`. Accepted root migrations currently end
at G009 version 7; G011's unaccepted implementation adds audit version 8 under
its approved Plan r2 and scoped attribution disposition (not included).
G012 is specified, not implemented. These are provisional planning assumptions;
final execution binds actual accepted G012/G011 source and review residuals.
Do not reserve a migration number or copy unfinished application code now.

Reuse ADR-006's separate audit ownership, G008's runtime/migrator/retention/helper
roles and RLS, G009's server-derived object identity and staged lifecycle, G012's
secret/encryption/vendor responsibilities, existing logging/CI/cleanup. There is
no recovery/retention command or production backup facility accepted today.
G011's approved maintenance handoff gives the retention role no audit privilege
until a separately reviewed G013 migration adds its narrow access.

## Scope and invariants

Only the existing generic stores and recovery/maintenance interfaces are in
scope. Durable canonical UUIDs and Organization/Group authority remain intact;
BusinessProfile is not a third tenancy layer. Core owns policy and audit; no
new application-administration identity, public endpoint or product schema.
Redis is ephemeral and starts empty after recovery, not a backup source of truth.

Do not deploy cloud backup, schedules, replication/PITR, whole-stack encryption,
new vendors, a job/queue platform or production incident integration. Operator
commands and concise procedures using existing tools are sufficient. Minimal
maintenance SQL/Go/tooling and an additive migration may be proposed in the Plan;
any new privilege is explicitly inventoried/reviewed. No blanket administrator
grant, trigger-disable bypass or extension of ordinary runtime authority.

## Requirements

### R1 — Finite recovery and retention contract

Inventory the existing data classes: application identity/session and tenant
records, audit records, object metadata/bytes, Redis ephemeral state, operational
evidence and backup copies. For each name its authority, included/excluded backup
content, recovery dependency, whether it holds secrets, its retention, hold and expiry behavior,
operator and remaining production prerequisite. Future product/RAG data is a
handoff, not an early schema.

The Plan pins a finite synthetic fixture: two Organizations, separate Groups,
one shared profile within an Organization, differing roles/global identities,
an existing disabled/revoked case, audit records and representative ready,
deleting/deleted object states with known bytes/digests. Reuse existing setup
helpers and positive/negative access cases; no new exhaustive oracle framework.
Use fresh private test credentials, never production/provider keys or real data.

Production backup cadence, retention durations/holds, RPO/RTO targets, notification
obligations, locations and encryption are owner/deployment decisions. Exercise
explicitly labelled fixture values; no invented universal defaults or SLOs.
Document what is configured, unknown or blocked. Local exercise success cannot
clear unknown production requirements or constitute GDPR/SOC 2 compliance.

### R2 — Consistent paired backup and private handling

Provide an operator-invoked, Docker-first checkpoint of PostgreSQL plus the
matching private object bytes/required metadata. A quiesced window is sufficient:
stop writers, drain in-flight uploads/deletions, and keep the fixture quiet until
both captures finish. The Plan names how this is established and detected, with
an explicit abort/hold if consistency cannot be established. A database snapshot
alone is not an atomic database/object-store snapshot.

Record a compact versioned backup description: source identity, schema/tool
versions, checkpoint/time window, component completion status, object key/size/
digest expectations and integrity hashes sufficient to detect mixed/missing/
corrupt inputs. Keep it private where identifiers are sensitive. Publish a
successful checkpoint only after every required component and integrity check
succeeds. A failed command in a pipeline, timeout, cancellation or inaccessible
source/destination cannot become a successful empty/partial backup. Incomplete
outputs stay marked incomplete and are eligible only for scoped cleanup/retry.

Reconstruct required cluster roles/grants through the accepted provisioning
contract; do not assume a database dump includes all cluster state. Never restore
old role passwords or grant backup authority to runtime just for convenience.
The Plan identifies the privileged operator boundary required to capture all
RLS-protected rows and verifies completeness; a runtime-filtered subset is not
a full database backup. Capture no unrelated database/project.

A database backup can contain sensitive authentication material even with
synthetic data. Raw dumps, bytes, credential files and restore-input records are
private operational material, **not CI evidence artifacts or committed fixtures**.
Use owned private storage, no credential-bearing arguments/logs, no broad
publication scan exemptions. Only sanitized summaries/check outcomes enter
Result/CI artifacts. Remove synthetic raw copies after verified closeout.
Production encryption/access/durability requirements remain explicit under G012.

### R3 — Isolated fail-closed restore and verification

Restore into a separately identified, fresh/proven-owned disposable target with
distinct server/storage resources and freshly provisioned service credentials.
Verify resolved resource identity, not merely a different project label. Reject
source/target aliasing, unrelated/nonempty targets, wrong checkpoint/components,
incompatible schema/tool input and corruption before unsafe mutation. No default
overwrite/reset of the real source or arbitrary caller-provided resources.

Keep ordinary target access and live external-provider activity disabled during
restore/reconciliation. Stop on partial restore/catalog/integrity failure and
leave the target unavailable; cleanup or a fresh retry is explicit. A restore
exit code or health endpoint by itself is not a reopen decision.

Verify exact expected fixture rows/relationships/UUIDs and object contents,
sizes/digests, required provider binding metadata and lifecycle semantics, plus migration ledger, schema ownership,
functions/triggers, role attributes/memberships/grants and RLS. Global/tenant
audit records persist subject to explicit retention/replay, with no forged
success attribution from restore. Use the actual restored runtime role for
representative allowed reads/mutations and refused cross-Organization/Group
access, prohibited audit mutation and privilege escalation. Superuser-only
row counts do not satisfy recovery. Backup/restore commands leave the source
unchanged, apart from explicitly planned synthetic post-checkpoint restriction/
deletion fixtures used in R4.

### R4 — Authentication and post-checkpoint restrictions

Fresh database credentials do not invalidate restored application sessions.
Before reopening, a reviewed **recovery-only** procedure invalidates restored
application sessions and pending login/logout transactions/capabilities that
could resume old authentication. Preserve canonical users/identity mappings and
legitimately retained audit history. Do not weaken normal identity/audit paths
or claim that local restore revokes a WorkOS session or an issued bearer URL.
Test an old application credential/callback refusal and a fresh synthetic login
against the restored current state; no live identity qualification is required.

A historical checkpoint must be reconciled with authoritative later deletion,
disablement, role restriction and other applicable scope restrictions before
ordinary access. Provide a minimal operator-controlled restore-input/receipt
procedure, held independently of that old checkpoint, with trusted provenance
and exact canonical subject/object/scope references. An existing trusted record
or bounded instruction file is sufficient; no new event bus/ledger service.

These instructions must remain available at least while affected backups remain
recoverable; audit retention/tombstone expiry must not erase their sole source.
Missing/stale/untrusted instructions or uncertain completeness leave reopening
blocked, not presumed safe. Pin that operator trust boundary and the completeness
decision in the Plan; no autonomous legal decision from an audit event or age.

Exercise one post-checkpoint user/role restriction and one object deletion,
including a case where the original source tombstone is no longer retained.
Restore the old data, apply the retained instruction set and verify denied old
authority, no newly issued read/signed access to the deleted object, removal of
its restored obsolete bytes, and an unrelated-tenant positive unchanged. Replay
must be retry-safe with truthful partial failure. No membership removal implies
deletion of shared BusinessProfiles or all global-user data.

### R5 — Restricted bounded retention and object repair

Default maintenance invocation to dry-run; mutation requires an explicit apply
mode and a complete operator-owned eligibility policy/scope. Fixture policies
cover expired/recent/held records and exact Organization or separately authorized
global scope. The Plan pins bounded batch size, cutoff/clock semantics, ordering,
timeouts, hold/restriction inputs and revalidation between preview and mutation.
Do not infer a lawful retention period or absent hold from missing configuration.

For audit events, use G011's dedicated retention-role handoff: only reviewed
schema usage, eligible-row DELETE policy/grant and the corresponding narrow
immutability-trigger exception together. Preserve runtime/ordinary-owner audit
mutation refusal, append/read authority, NOINHERIT/NOBYPASSRLS and forbidden
UPDATE/TRUNCATE/DDL/role escalation. Eligibility must be enforced at the reviewed
database boundary, not only in a shell filter. Scope/hold decisions belong to
the trusted maintenance policy; the Plan identifies what a malicious privileged
operator remains able to do. No general audit-disable switch or new app role.

For object maintenance, reuse guarded staged deletion where an authorized actor
operates; separately restricted maintenance cannot turn a credential into an
ordinary-user permission bypass. Purge metadata only for eligible, unheld,
confirmed-deleted objects after confirmed provider absence and preservation of
required restore-replay instructions. Do not erase the sole reconciliation
record for remaining bytes or cascade-delete audit history.

Pending age alone does not prove an upload inert. Hold live/ambiguous pending
rows unless the quiesced window/reconciliation proves no late writer can publish
them. Deleting rows may be retried through the reviewed recovery path; provider
failure/unknown outcome stays pending, never falsely complete. Verify timeout/
provider failure/retry and preservation of a ready object, recent row, hold and
foreign/out-of-scope record. Reuse G009 unknown-commit guarantees rather than
replaying its entire failure matrix. No background sweep/scheduler/bucket scan.

Backup expiry follows the explicit policy and holds; deletion of a local file
is not proof a provider/immutable backup copy disappeared. Demonstrate expiry
eligibility and an expired backup's refusal/reconciliation behavior on owned
synthetic material. Record partial deletion and retained copies honestly.

### R6 — Operational signals, incident drill and recovery objectives

Reuse existing readiness/logging/diagnostic records. Provide the smallest
operator-invoked status/failure hook for backup completion/freshness/integrity,
restore/replay failure and maintenance pending/failure. Missing policy/unknown
state is not healthy. Bound checks and keep output secret/content-free.
No metrics stack, collector, daemon, pager vendor or actual notification account.

Exercise one representative failure end-to-end: detect a bad/missing checkpoint
or failed reconciliation, emit a sanitized non-success signal to a local test/
operator evidence destination, follow containment/triage and recover safely.
Documentation alone is not exercised alert evidence, but no external message is
required. Pin any operator feedback before final CI; a deterministic local sink
can demonstrate the hook without an interactive stage when the Plan says so.

Runbooks name detection, incident owner/escalation, scoped containment, secret/
credential handling, evidence access/redaction, restoration/replay verification,
rollback or quarantine, and reopen sign-off. Map to actual G011 events/diagnostics
without requiring successful audit append to undo security revocation. Legal/
customer notification decisions and deadlines remain qualified owner obligations.

Record data/byte counts, checkpoint window, backup/restore/verification durations,
recovery-point/data-loss meaning and environment for the single drill. Separate
observations from proposed production RPO/RTO/cadence and unconfigured alert/
storage/encryption/processor responsibilities. No load/benchmark campaign or
arbitrary target must be met. Passing the fixture does not qualify live R2,
managed PostgreSQL, production scale or the operator's future response time.

### R7 — Proportionate validation and durable evidence

Start with focused script/config/maintenance tests and one owned source/target
fixture. One paired happy-path recovery plus bounded negative cases can cover
multiple criteria; do not run a full foundation stack for every corruption,
scope or policy input. Reuse accepted unchanged permission/backup-tool controls;
add only the minimum current harness/required-check mapping and changed-behavior
coverage. Source checks remain required until separately reviewed retirement.

Actual infra/migration/authorization paths select full under the current map;
CI-controls run when CI/harness changes require them. The Plan declares actual
local/MR/staging selection, resource limits and focused commands. No runner/
profile redesign, general orchestration, duplicate full runs or separate control
campaign is a prerequisite. Follow failed-batch fix/rerun and safe pre-batch
target/binding reconciliation, reusing unaffected passing evidence.

Record source/target commits, tools/image digests, synthetic fixture/policy and
backup-input provenance, outcomes (failed/interrupted/skipped explicit), resource
ownership and relevant query/index/helper/lock observations. Flag new costs to
the owner through PERFORMANCE.md/RISKS.md; no silent tuning or unreviewed risk.
Pin deadlines and batch bounds in the Plan before execution.

Raw backup/test operational material is temporary/private. Keep compact Git
Results/reviews and sanitized CI artifacts; clean owned source/restore resources,
raw backup bundles, private instruction/credential files and local caches under
existing closeout policy. Interrupted work preserves only named necessary
unfinished inputs with owner/reason/release condition. No unrelated prune.

## Acceptance criteria and verification

| ID | Required outcome | Objective evidence |
|---|---|---|
| AC1 | R1/R2 define and produce a complete consistent checkpoint. | Finite store/fixture matrix and explicit consistency window; synthetic two-Organization database/object capture with version, completeness and integrity records. Partial component/command failure cannot report success; backup/restore leave the source unchanged apart from planned R4 fixture mutations. |
| AC2 | Restore is isolated and rejects unsafe/incomplete inputs. | Separate proven-owned target; source/target collision, wrong/mixed/missing/corrupt component and partial restore failure cases fail safely with target access blocked. No source/unrelated target overwrite; only sanitized proof published. |
| AC3 | Restored content and authority match the intended state. | Expected UUID relationships, migration/catalog/roles/RLS/audit records, object digests and required provider binding metadata verified; representative real-runtime positives, cross-tenant/Group negatives and audit/privilege refusals. Fresh service credentials; old application credential/pending authentication refused and fresh synthetic login works. |
| AC4 | Post-checkpoint restrictions/deletions cannot be silently resurrected. | Trusted independent replay inputs survive source tombstone expiry; after restore a post-cut user/role restriction and object deletion are reapplied before access, old authority/new signed issuance denied and obsolete bytes absent. Retry/partial failure truthful; missing/untrusted/incomplete replay keeps target unavailable; unrelated tenant positive unchanged. |
| AC5 | Audit retention is restricted and bounded. | Dry-run changes nothing; explicit eligible apply removes expired/unheld in-scope records only; recent/held/out-of-scope preserved. Actual retention credential succeeds only through reviewed eligibility boundary, runtime/ordinary owner and forbidden UPDATE/TRUNCATE/DDL/role escalation remain refused. Catalog and populated migration checks for any new grants/functions/trigger exception. |
| AC6 | Object maintenance and backup expiry preserve truthful lifecycle/holds. | Confirmed-deleted eligible metadata purge only after byte absence and replay preservation; ready/recent/held/foreign/live-or-ambiguous pending controls unchanged. Deletion failure stays pending, retry succeeds; backup policy expiry/hold demonstrated without falsely claiming all copies erased. No audit-history cascade. |
| AC7 | Recovery/maintenance failures reach an exercised operational signal. | One bounded failure/triage/recovery drill with non-success status and sanitized local sink output; unknown/stale/missing-policy status not healthy. Containment/reopen steps followed; logs/audit signals match actual effects. No live external notification or invented alert deployment. |
| AC8 | Runbooks distinguish observed recovery from production responsibilities. | One drill's sizes/window/durations and recovery-point meaning recorded; operator roles, RPO/RTO/cadence decisions, credentials, encryption/location/vendor/backup retention, legal/privacy duties and rollout blockers explicit. No cloud/production/compliance claim. |
| AC9 | Validation/provenance/privacy cover the change without a new framework. | Focused plus actual path-selected final coverage under current local/MR/staging rules; exact tested commits/inputs/tools, failure history and evidence reuse. Raw sensitive bundles never published; scan existing outputs, inspect required-check execution and minimal wiring. No skipped required behavior, duplicate broad proof campaign or unrelated CI work. |
| AC10 | Review, cleanup and epic closeout are complete. | Result maps AC1–AC9, actual diff/read-only security review, B7/checkpoint4, scoped normal/interrupted cleanup receipt, required staging disposition and Codex final ACCEPT. EPIC-004 PROGRESS compaction outcome/deferral recorded; G014/G021/G056/OS001 handoffs explicit. |

## Artifacts, Plan decisions and gates

Claude's Plan selects minimal command/procedure files, restricted maintenance
changes/tests, existing harness additions and documentation location (for example
one RECOVERY guide linked from OPERATIONS/MIGRATIONS/AUDIT/OBJECT-STORAGE).
No prescribed new framework/directory tree, retention service or snapshot registry.
Reconcile G012's actual final operating documentation before creating duplicates.

Resolve before Plan approval:

1. Final accepted predecessor state, maintenance-role/trigger contract and schema
   inventory; label provisional assumptions if G012 integration is awaiting CI.
2. Backup consistency/quiescence and required role/ownership restoration,
   target-identity/refusal checks and raw material isolation. Existing provider
   tools are preferred; pin compatible versions and exact commands.
3. Finite maintenance policy/hold inputs, eligibility privilege enforcement,
   batch/cancellation bounds and any additive migration/guard exceptions.
4. Restored-auth invalidation and trusted later-restriction/replay procedure,
   source of completeness and failure behavior, including tombstone expiry.
   No speculation that old audit rows reconstruct all later decisions.
5. Test fixture values, deterministic alert destination, manual feedback if any,
   AC reuse/mapping, actual CI triggers and remaining production owner decisions.

No major deviation, blanket privileged access or required evidence deferral is
preapproved. Fixture policies/measurement choices are not production legal/RPO/
RTO decisions. Completion requires all ten ACs and accepted integration evidence;
planning/implementation obey G012 and Plan/input gates. G013 closes EPIC-004.
G014 receives recovery/session/replay semantics; G021 verifies reusable controls;
G056 later supplies assembled-product rights; OS001 exports only accepted G021
generic mechanisms. No detailed G014 Spec or product work is authorized here.

## Technical references for planning

[PostgreSQL 17 pg_dump](https://www.postgresql.org/docs/17/app-pgdump.html) documents
consistent database capture and its RLS/security behavior; it does not provide a
paired object-store snapshot. [pg_dumpall](https://www.postgresql.org/docs/17/app-pg-dumpall.html)
documents cluster-global roles/settings and password handling. These inform the
reviewed procedure, not an instruction to export password hashes or use an
unreviewed privileged connection. Exact accepted-version behavior is tested
during authorized execution.
