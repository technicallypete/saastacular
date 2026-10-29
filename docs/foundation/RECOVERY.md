# Recovery, retention and incident procedures

The recovery tooling is one operator command set, `infra/recovery/recovery.sh`, one additive privilege migration
(`core/migrations/00009_retention_maintenance.sql`) and a paired synthetic drill in the `migrations` validation suite
(B-RETENTION-CATALOG, B-RETENTION-IT, B-RECOVERY-DRILL, B-MAINTENANCE-DRILL). It demonstrates a recoverable,
authorization-safe PostgreSQL and object-store checkpoint and bounded maintenance on disposable synthetic environments.
**A successful local drill is not production backup qualification, an RPO/RTO commitment or a privacy-compliance
claim.** Production cadence, location, encryption, retention periods and response duties stay operator and deployment
decisions (see [Production responsibilities](#production-responsibilities)).

The tool runs on an operator host with Docker. It locates the project's containers by their Compose labels, reads
values from the operator's `.env`-style files and passes them to containers only through `--env-file` process
substitutions: never on a command line, never in output. Archives travel on stdin and stdout; nothing is bind-mounted.
Raw backup sets, the restore-input ledger, attestations, receipts and env files are private material: keep them
outside the repository with mode 0700/0600, never in CI artifacts, fixtures or evidence. Only the tool's sanitized
lines (statuses, counts, sizes, durations, reason codes, backup UUIDs) are publishable.

## Data classes and responsibilities

| Class | Authority | In the checkpoint | Recovery dependency | Private material | Retention, hold, expiry | Operator | Production prerequisite |
|---|---|---|---|---|---|---|---|
| Identity and sessions (`users`, `identity_principals`, `sessions`, `login_transactions`, `identity_provider_binding`; the platform root signup's `signup_sessions`) | core PostgreSQL | yes | provisioning, migration ledger; sessions and signup credentials invalidated before reopening | recoverable PKCE verifiers and CSRF values for unexpired rows; hashes; provider identifiers; names ([SECURITY.md](../SECURITY.md) §3) | sessions expire; login transactions at most 10 min; signup credentials at most 30 min (continuation secrets 120 s); restore expires all sessions and signup credentials and deletes login transactions | platform operator | backup encryption and access control; one WorkOS namespace |
| MCP connections (`mcp_connections` and each connection's `mcp` session row; [MCP.md](MCP.md)) | core PostgreSQL | yes | the session invalidation below | identifiers, host, connection URL, client ID and selected Group; no token, refresh value or hash of one | a restore ends every connection (its session expires); post-checkpoint connections are absent and cannot be recreated from a bearer token | platform operator | none beyond the identity row; provider revocations are not undone by a restore |
| Tenant records (Organizations, Groups, bindings, BusinessProfiles, links, policy flag) | core PostgreSQL | yes | replay of later restrictions | names and descriptions | no automatic deletion | platform operator | application-specific real-data decisions |
| Audit (`audit.events`) | core PostgreSQL, append-only for runtime | yes, exactly | none (history preserved) | identifiers, no payloads ([AUDIT.md](AUDIT.md)) | the retention role with an explicit policy and holds | platform operator | legal retention periods and holds |
| Object metadata (`stored_objects`) | core PostgreSQL | yes | object bytes and binding | identifiers only | tombstones purged only after provider absence and ledger preservation | platform operator | as audit |
| Object bytes and binding metadata | object provider (MinIO locally, R2 planned) | yes, with size, SHA-256 and binding | metadata rows | tenant content | follow the metadata lifecycle | platform operator | provider versioning/immutability, encryption, location (R2 not qualified) |
| Redis | ephemeral (readiness only) | **no** | starts empty | none persisted | none | platform operator | keep ephemeral ([SECURITY.md](../SECURITY.md) §7) |
| Operational evidence (logs, receipts, the status sink, CI artifacts) | operator, GitLab | no | none | sanitized only | CI 90 days; local evidence temporary | platform operator | log retention policy |
| Backup sets | operator private storage | n/a | manifest, ledger, provisioning | everything above | backup policy with holds and quarantine | platform operator | cadence, retention, location, encryption, immutability, RPO/RTO |
| Restore-input ledger and attestation | operator private storage, separate from backups | no (held independently) | none | identifiers only | never pruned by the tool; must outlive every recoverable backup | incident owner | durable separate storage |

Product, RAG and AI data are application-specific; the foundation has no early schema for them.

## Backup: the capture fence and the trusted administrator window

```sh
bash infra/recovery/recovery.sh backup --env-file SOURCE.env --backup-root DIR --ledger LEDGER --admin-window REF [--use-fence]
bash infra/recovery/recovery.sh fence   --env-file SOURCE.env --backup-root DIR --admin-window REF
bash infra/recovery/recovery.sh unfence --env-file SOURCE.env --backup-root DIR [--fence-id ID]
```

A paired database/object checkpoint is consistent only while nothing writes either store. The tool establishes an
exclusion and drain, held for the whole capture:

- **Fence record first.** Before any change it writes `DIR/fences/<fence-id>.tsv` (0600), flushes and reads it back:
  the source identity (project, containers, volumes, PostgreSQL system identifier), the exact prior database ACL, the
  prior status of the object service user and the administrator-window record. Each later step appends one line after
  it completed and was verified. A record that never completed its read-back means no change was made.
- **X1** stops `core`, `web` and `intelligence` (graceful, bounded) and waits until they exited.
- **X2** revokes CONNECT on the database from exactly the recorded non-superuser grantees (runtime, migrator,
  retention) and terminates their backends; their open transactions roll back. A database where PUBLIC holds CONNECT
  is refused at X0, and after the revoke no non-superuser login may still be able to connect (directly, through PUBLIC
  or through an inherited role), or the fence aborts.
- **X3** a labelled superuser holder session takes `SHARE` locks on every table of `app` and `audit` (lock timeout 30 s:
  an administrator write in progress refuses `admin_write_in_progress`) and idles in its transaction for the fence's
  maximum duration. It drains writes already in progress and blocks new ones while reads and `pg_dump` proceed. It is a
  bounded safeguard, not a claim against every privileged catalog change.
- **X4** disables the application's object service user (fixture MinIO administrator) and verifies it is refused.
- **X5** restarts MinIO (graceful stop until exited, start, ready): accepted provider operations either committed
  before or were abandoned; none can commit during capture.

**Trusted administrator-exclusive window.** Administrator exclusivity across both stores, PostgreSQL catalog activity
included, is a recorded operational prerequisite, not something the tool proves. For the whole window the operator
must be the only administrator of the source PostgreSQL cluster and MinIO deployment. `--admin-window REF` (a
non-confidential change or incident reference) is recorded in the fence record and the manifest; without it the tool
refuses `admin_window_unavailable` before any change. Another superuser session at X0 or at the final check, a
changed holder or a lost lock aborts `window_changed`. These checks detect only some contradictions; their absence
proves nothing. The guarantee holds within the recorded window; it is not claimed against a concurrent privileged
administrator who violates it (privileged administration is outside the runtime guarantee). No MinIO trace is used.

**Capture.** C1 ledger sync and mark; C2 data fingerprint (per table row count and ordered row digest) and catalog
fingerprint (relations, columns, policies, function bodies, triggers, constraints, indexes, default privileges and the
project's roles, with ACLs normalized: a restore replays the same privileges in another item order); C3 object listing;
C4 `pg_dump --format=custom` of exactly the project database as the superuser, streamed on stdout, its table of
contents checked (only schemas `app` and `audit`, exactly the fingerprinted tables); C5 the bucket's bytes as one tar
and the binding metadata, read by the fixture administrator; C6 the window checks, then the supplementary re-check
(fingerprints and listing unchanged, no other superuser session); C7 per-object size and SHA-256 recomputed from the
tar. `manifest.tsv` (with its fixed backup format line) records all of it; `COMPLETE` holds the manifest's SHA-256 and is written
last. **`COMPLETE` means capture completeness only.** Fence release is a separate source state: the fence record gains
`released` and the backup receipt says `release: verified`, `failed` (exit 4, `unfence_failed`; the set stays complete)
or `deferred` (`--use-fence`). A failed, timed-out, signalled or inaccessible capture writes `INCOMPLETE` with a closed
reason, releases a fence it established, and never writes `COMPLETE`; an unexpected exit during fencing or capture
(any failed step) does the same through the tool's exit guard, whose release does not depend on writing the marker
or a log line; it then reports the fence's actual state (released, or still held: run `unfence`). Under `--use-fence`
the fence is the operator's, not the capture's: a signal, a timeout, a refusal or an unexpected exit during that capture
never releases it. The record stays without `released`, CONNECT stays revoked, the service user stays disabled and
`status` reports `backup.fence: held`; the set is `INCOMPLETE` and the exit non-zero, and only an explicit `unfence`
restores the prior state. One fence or backup runs per backup root at a time
(`fence_busy` otherwise), and `--max-duration` of `fence` is 1-7200 seconds.

`unfence` converges every element to the recorded prior state whatever progress lines exist: it terminates the holder's
backend (removing its container alone does not end the server backend), removes its containers, sets the service user
to its recorded prior status (a user disabled before the fence stays disabled) and re-grants CONNECT only to the
recorded prior grantees, then verifies and appends `admin_window_end` and `released`. Every step is idempotent: run it again after any
interruption, including of an earlier `unfence`. Until it succeeds the source stays closed to writers (fail-closed),
and `status` reports `backup.fence: held`. Application services are restarted by the operator afterwards.

Optional environment of the tool: `RECOVERY_STEP_TIMEOUT` (seconds per step, default 300), `RECOVERY_LABEL` (an extra
label on every container the tool creates) and `RECOVERY_CID_DIR` (a directory for the distinct, initially absent
cidfiles of its detached holders; the validation harness uses both for its ownership registry).
`RECOVERY_DRILL_PAUSE_AFTER_X2` (a bounded pause between X2 and X3), `RECOVERY_DRILL_PAUSE_IN_REPLAY` (a bounded pause
after replay read the ledger head) and `RECOVERY_DRILL_FAIL_CAPTURE=marker` (a capture failure whose INCOMPLETE marker
cannot be written) are validation-drill seams, inert unless set, and are never used in operation. A fence refuses `fence_held` while an unreleased fence record of the same source exists.

## Restore and verification

```sh
bash infra/recovery/recovery.sh verify  --backup DIR/ID
bash infra/recovery/recovery.sh restore --backup DIR/ID --target-env-file TARGET.env --receipts RCPT --backup-policy POLICY \
  [--source-env-file SOURCE.env] [--steps all|database|objects]
```

Preflight refuses, before any target change: an incomplete, corrupt, missing or mixed component; an expired or
quarantined set (or one past `backup_retain_for_s` and not held); an unknown format, another server major version, a
schema version above this checkout or another database name; a target that aliases the source (project, container,
volume or system identifier, from the manifest and, when given, the live source); a non-fresh target for a first
restore; a running target `core`, `web` or `intelligence`; and, for a first restore, a receipts directory that already
holds invalidation, replay or reopen receipts (`receipts_reused`: every restore uses a fresh receipts directory). The target is a fresh volume provisioned by
`infra/postgres/init/10-roles.sh` with fresh values; its roles must equal the set's role section. The database step
renders the table of contents minus the two provisioned schema entries with `pg_restore` and applies it with `psql
--single-transaction -v ON_ERROR_STOP=1` (data loads before the post-data triggers and policies, so no audit trigger
fires and nothing is disabled). The same transaction first withdraws the provisioned default sequence grant of the
owner, so a restored sequence (the migration ledger's) does not gain it; the set's own default-privilege entries then
restore the source's exact defaults. The step passes only when the data and catalog fingerprints equal the manifest.
The catalog fingerprint treats a NULL ACL as its default and compares expression texts without parentheses and
whitespace, because a restore re-deparses nested AND/OR without changing their meaning (a difference only in grouping
is not detected by the fingerprint). The object step uploads each
object with its content type, `Content-Disposition: attachment` and binding, and reads it back. A provider failure
leaves the receipt `partial`; `--steps objects` continues only on the exact receipt-bound target and checkpoint, and
only when every object already on the target has the set's name, size, bytes and binding (`continuation_mismatch`
otherwise; a mismatched object is never overwritten). `restore.tsv` says `restored-closed` only when every step passed. There is no
reset or overwrite command: discarding a failed target is a separate operator action.

## Restored authentication, later restrictions and reopening

```sh
bash infra/recovery/recovery.sh invalidate-auth --target-env-file TARGET.env --receipts RCPT
bash infra/recovery/recovery.sh ledger init|append|sync|verify|attest ...
bash infra/recovery/recovery.sh replay --backup DIR/ID --ledger LEDGER --target-env-file TARGET.env --receipts RCPT
bash infra/recovery/recovery.sh reopen --backup DIR/ID --ledger LEDGER --receipts RCPT --sign-off REF --runtime-verification pass
```

`invalidate-auth` runs one superuser transaction on the closed target: every unrevoked session expires now and gets a
fresh CSRF value, every unreturned provider-logout reference expires, every login transaction is deleted, with
postconditions (zero live sessions, login transactions and unexpired references) or rollback. No `revoked_at` changes,
so no `session.revoked` event is forged. The platform root signup state is folded into the same `infra/recovery/sql/invalidate.sql`
(`recovery.sh` needs no signup-specific step): every unrevoked root signup credential and its current continuation generation expire
now (a generation-0 row keeps its NULL continuation expiry), counted as `signup_sessions_expired` in the receipt, with
the postcondition that no unrevoked signup row stays live; a restored signup credential, reference, challenge, nonce or
proof is therefore refused, and signup login transactions and pinned handoffs are deleted with the other login
transactions. Organizations created by signup before the checkpoint restore as ordinary tenant records. Every MCP connection owns a session row, so pending and active connections are
terminal after the invalidation (TestMCPRecovery runs this exact SQL), and the owners create new ones through the
setup page. No MCP replay kind is needed. Re-enabling a user after a restore is not unconditional: the MCP client containment
runbook's prior-state, explicit-clearance and other-hold checks still govern, and an already-disabled or otherwise held
user stays disabled ([OPERATIONS.md](OPERATIONS.md), "Compromised MCP client containment"). Not claimed: revocation of
WorkOS sessions or Connect tokens, or of signed URLs issued by the
source (valid against the source provider for at most 300 s). The invalidation receipt records the restore binding
(a digest of the restore receipt's identity lines) and the target identity; replay and reopen accept only receipts
bound to the current restore (`receipt_unbound` otherwise), and earlier replay gaps carry forward only from a replay
receipt of the same restore.

The **restore-input ledger** (with its fixed ledger format line) is a hash-chained, operator-held file in storage separate from
backups: sequence, previous-line digest, UTC time, kind, canonical references, origin (`sync` or `operator`) and a
non-confidential reference. Only restrictions are recorded: `object.delete`, `user.disable` and `policy.disable` by
`ledger sync` from authoritative state (every backup syncs before its mark); `membership.remove`,
`membership.downgrade` and `profile.unlink` by `ledger append` (the privileged primitives have no recorder yet, F-6).
References are strict, at append and again at replay: exactly the kind's keys, canonical lowercase UUIDs, a closed-set
scope type and roles of that scope, an Organization binding scoped to its Organization, and a downgrade that lowers the
role (owner > admin > member); anything else is refused at append (exit 2) and replays as `conflict`.
Object deletion is terminal and recorded once; reversible restrictions are re-recorded after the latest backup's mark.
`ledger attest --basis final-sync|last-sync` records the operator's completeness attestation; `last-sync` needs
`--owner-risk REF`. The chain makes edits detectable and the manifest anchors it at the mark; a malicious operator can
still omit entries or attest falsely (outside the mechanical guarantee; the incident owner's responsibility).

`replay` applies the entries after the backup's mark, one superuser transaction each, idempotently (`applied`,
`already_applied`, `conflict`, `partial`); `already_applied` requires every referenced entity to exist in the restored
checkpoint with the link or binding absent (or the recorded state reached), and the receipt names the executing
operator. A replay covers exactly the ledger head it validated at its start: the receipt records the ledger identity,
through-sequence and head digest, entries appended later are not applied, and a head that changed before the end
makes the run `incomplete` (`ledger_drift`). `reopen` requires the replay receipt to cover the current attested head
(`replay_stale` otherwise): a new entry needs a new replay, not only a new attestation. Object deletions run in the recorded deleter's context (the storage guard
requires an actor) and remove the provider bytes; a provider failure (including a stopped target MinIO) leaves the row
`deleting`, never readable or signable, and a rerun resumes. A replayed `user.disable` reads the audit gap marker before COMMIT, never rolls the
disablement back, and on a gap makes one bounded independent `revocation.unaudited` attempt; the receipt says `gap` with
`independent=recorded|failed|not_attempted`, and earlier gaps survive reruns. `reopen` blocks on a missing ledger or
attestation, a chain or mark failure, a stale ledger, an attestation that does not cover the head, `last-sync` without an
owner reference, an incomplete replay, a restore not `restored-closed`, missing invalidation or a failed runtime
verification. It writes `reopen.tsv` with the acknowledged gaps and prints the start command; until then the target's
application services stay stopped.

## Retention and object maintenance

```sh
bash infra/recovery/recovery.sh retention audit|objects --policy POLICY --env-file ENV --preview PREVIEW [--ledger LEDGER]
bash infra/recovery/recovery.sh retention audit|objects --policy POLICY --env-file ENV --preview PREVIEW --apply --receipts RCPT [--ledger LEDGER]
```

Migration 00009 gives the existing `<db>_retention` login exactly: USAGE on schema `audit`; SELECT on `audit.events`
(`id`, `occurred_at`, `organization_id`) and `app.stored_objects` (`id`, `organization_id`, `state`, `deleted_at`);
DELETE on both; EXECUTE on `app.retention_eligible`. Four policies (SELECT and DELETE, retention only) admit only rows
the transaction's own declaration makes eligible: `retention.scope` (`organization:<uuid>` or `global`),
`retention.cutoff` (UTC, strictly before the transaction time) and `retention.holds` (`none` or up to 256 UUIDs). Any
absent or malformed setting raises 22023. `audit.events_refuse` lets only that login delete an eligible row; every
other UPDATE, DELETE and TRUNCATE by every role, owner and superuser included, still raises append-only. The login
cannot read other rows or columns, insert or alter history, update or truncate anything, touch other tables, change
roles or bypass row-level security; scope and hold decisions are the trusted policy's.

The policy file is strict `KEY=VALUE` (unknown, duplicate or missing keys exit 2; no default period or hold):
`policy_id`, `policy_label`, `scope`, `global_authorization` (exactly with `scope=global`), one of `audit_cutoff` /
`audit_retain_for_s` and of `object_cutoff` / `object_retain_for_s`, `holds`, `batch_size` (1–500), `max_batches`
(1–1000), `statement_timeout_ms`, `lock_timeout_ms`, `run_timeout_s`. A dry run resolves the cutoff once from the
database clock, counts and writes the preview; nothing changes. `--apply` requires the preview, refuses a changed
policy or a grown eligible count, reuses the resolved cutoff and deletes in bounded batches (each its own transaction
with its timeouts; the database re-checks eligibility at DELETE time), stopping at `max_batches`, the run timeout or a
signal after the current batch. The receipt says `complete`, `partial` or `failed`, never a false completion: eligible
rows left at the stop make it `partial`.

**Availability residual (accepted limitation; migration 00009).** PostgreSQL lets any holder of DELETE on a table
take every `LOCK TABLE` mode on it for the rest of its transaction, and row-level security does not limit table locks;
there is no separately revocable LOCK privilege. The table-level DELETE grants therefore let the retention
login (or anyone holding its value) take, for example, ACCESS EXCLUSIVE on `audit.events` or `app.stored_objects` and
hold it in an idle transaction. That blocks, across every Organization, each audited write (every write path that
records audit history) and every object read, upload or delete until the session ends. It cannot read, change or
delete anything beyond its policies. The command's statement and lock timeouts bound normal operation but are no
guarantee against a compromised credential. Controls: the value stays in the operator's private env file and is never
given to core, web or intelligence. Detection: inspect blocking sessions, a lock granted to `<db>_retention` on those
tables or an idle-in-transaction retention backend outside a running `retention` command:

```sql
SELECT a.pid, a.usename, a.state, a.xact_start, l.mode, l.relation::regclass
FROM pg_locks l JOIN pg_stat_activity a ON a.pid = l.pid
WHERE a.usename LIKE '%\_retention' AND l.locktype = 'relation' AND l.granted;
```

Recovery: terminate those backends (`pg_terminate_backend`) and rotate or withdraw the retention value (SECURITY.md
section 6).

Tombstone purge additionally skips `not_preserved` (no `object.delete` ledger entry), `bytes_present` (repair by an
authorized repeat delete) and `provider_unavailable` (also when the source MinIO is stopped); it needs an Organization
scope. Pending and deleting rows are
never selected: pending age never proves an upload inert. They are reported by `status` and repaired only through the
existing guarded delete by an authorized actor ([OBJECT-STORAGE.md](OBJECT-STORAGE.md)).

**Bounded head-of-line blocking (accepted limitation).** Purge selects candidates oldest first
(`deleted_at`, then `id`) in batches of `batch_size`, and a skipped row is not deleted, so it stays at the head of every
later batch. An invocation therefore ends at the first batch that deletes nothing: when the oldest batch is fully
skipped (`not_preserved`, `bytes_present` or `provider_unavailable`), younger eligible tombstones behind it are not
reached in that run. The receipt is truthful about it: `partial` (exit non-zero) with the skip counts and the
remaining eligible count, never a claim that every eligible row was visited. An unchanged retry selects the same head
and makes no progress. The operator resolves the actual blockers first, only through the existing paths: for
`not_preserved`, synchronize the trusted ledger or append the missing `object.delete` preservation entries; for
`bytes_present`, complete the authorized repeat delete through the guarded delete; for `provider_unavailable`,
restore the source MinIO and the scoped object values; then run a new dry run and apply. Never delete or alter blockers
to force progress, never erase ledger or replay evidence, and never read a `partial` receipt as completion. No keyset
paging or starvation-free traversal is claimed.

## Backup expiry and quarantine

```sh
bash infra/recovery/recovery.sh expire --backup-root DIR --backup-policy POLICY [--apply]
bash infra/recovery/recovery.sh quarantine --backup DIR/ID --reason CODE
```

The backup policy is strict `KEY=VALUE`: `policy_id`, `backup_max_age_s`, `backup_retain_for_s`, `backup_holds` (may be
empty, never absent) and `pending_review_after_s`. A set is eligible when older than `backup_retain_for_s`, not held
and not quarantined. Apply writes `EXPIRED` first and removes the components; anything left gives `EXPIRED-PARTIAL`
with the remaining items, and a rerun continues. When every component is gone the manifest and `COMPLETE` are removed
too, so only the `EXPIRED` record remains. Only the local copy is managed: provider, replicated or immutable
copies are unknown. Restore refuses expired, partially expired and quarantined sets; restoring a copy found elsewhere is
outside the tool and needs an explicit operator decision, replay and a retention dry run on that target before reopening.

## Status hook

```sh
bash infra/recovery/recovery.sh status --backup-root DIR --backup-policy POLICY --receipts RCPT [--env-file ENV] [--ledger LEDGER] [--sink FILE]
```

One JSON line per check (`time`, `check`, `status`, closed `reason`, optional backup UUID; no path, tenant identifier,
value or row content) and a summary. Checks: `policy` (`ok`, `unknown`), `backup.latest` (`ok`, `missing`, `stale`),
`backup.integrity` (`ok`, `failed`, `unknown`), `backup.fence` (`ok`, `held`), `restore.last` and `replay.last` (`ok`,
`failed`, `partial`, `none`; `replay.last` also `gap` until a reopen sign-off covers it), `maintenance.last` and, with
`--env-file`, `maintenance.pending_review` (`ok`, `pending`: counts only). `none` (never run) is healthy. Exit 0 only
when every check is `ok` or `none`; 1 otherwise; 2 for invalid input. `--sink` appends the same lines to a mode-0600
local file: the deterministic local alert destination. No metrics stack, daemon, pager or notification account.

## Legacy ownerless records

Migration 00015 refuses, naming counts only (`legacy owner preflight: N organizations / M groups without an owner
binding`), while any Organization lacks an `org_owner` or any Group a `group_owner` binding (`source = 'membership'`).
Nothing is promoted or deleted automatically. The restricted procedure, run by the trusted administrator in the
quiescent migration window:

1. List the ownerless records with the attested superuser (the same counting predicate as the preflight; the output
   stays in the operator's private evidence, never in a ticket or log).
2. For each record, confirm with the tenant which existing member becomes the owner; prefer a current admin of the
   same scope and, for a Group, a current member of its Organization.
3. Write that owner binding through the superuser repair path (the authority guard returns early for the superuser,
   so the change is not a runtime path; the existing `role_bindings` capture records it as `membership.granted` or
   `membership.role_changed`). An abandoned record is removed only under the existing restore/removal procedures, never
   as a side effect of the migration.
4. Re-run `migrate up`; the preflight passes once no ownerless record remains.

After version 15 the guard keeps at least one owner for every Organization and Group. A guarded downgrade or removal
after a backup is a later restriction of the existing `membership.downgrade`/`membership.remove` ledger kinds (replayed
after a restore; no new kind); a mistaken change is reversed by an ordinary guarded promotion or invitation, never by
restoring a binding. All-owner unavailability (every owner disabled or retired) still needs this restricted,
verified superuser path; emergency suspension keeps the ownership binding without usable access.

## Incident runbook

1. **Detect.** `status` non-zero (or its sink line), a failed receipt, readiness 503, `audit.gap` or
   `audit.record_failed` diagnostics (AUDIT.md).
2. **Own and escalate.** The platform operator is the incident owner. Legal, customer-notification and deadline
   decisions belong to the deploying organization's accountable owner; the tooling never decides them.
3. **Contain, scoped.** Stop writers (`fence`), keep restored targets closed, `quarantine` suspect sets, rotate service
   values per [SECURITY.md](../SECURITY.md) §6 when exposure is suspected. Never modify the source during backup or
   restore; a target can always be discarded.
4. **Private material.** Raw sets, ledgers, attestations and receipts stay private; evidence is redacted to the
   tool's sanitized lines.
5. **Triage.** `verify` names the failed component; receipts name the failed step.
6. **Recover.** `restore` a complete, unquarantined set into a fresh owned target; `invalidate-auth`; `ledger sync`
   (against the source after writers stopped, if reachable) and `ledger attest`; `replay`; the runtime verification of
   the restored target through the real guarded facade; `reopen --sign-off`. The procedure never requires a successful
   audit append to undo a security revocation.
7. **After reopening.** Take a fresh checkpoint of the reopened environment; `status` returns `ok`; redo any later
   re-grants the restore lost (replay is over-restrictive by design for reversed restrictions).

## Observed drill and production responsibilities

The validation drill records its sizes, window, durations and the recovery point (the checkpoint's window end: changes
after it are lost except replayed restrictions) in `recovery-drill.tsv`. These are observations of
a local Docker Desktop/WSL or hosted DinD run, not RPO/RTO targets.

### Production responsibilities

Not implemented, explicit operator/deployment prerequisites: backup cadence, RPO/RTO, alerting destinations, backup
location, encryption at rest and key custody, immutability and versioning, processor responsibilities, legal retention
periods and holds, and incident notification duties; a managed-PostgreSQL administrator able to revoke CONNECT,
terminate backends and hold table locks; for R2 (no per-user disable or restart), an attested precondition (every write
value revoked, then a settle interval) plus a separate read-only capture value; dedicated read-only provider values for
backup and purge checks and write values for restore; and, as locally, a recorded trusted administrator window across
both stores. The tool refuses when no administrator window is recorded.

**Organization branding.** `app.organization_branding` is ordinary Organization data captured and restored with the database;
logo bytes are stored objects restored with the object store. After a restore, a branding row whose logo object is
absent shows the platform-default mark (the management view reports the logo as `absent`) until an owner or admin
removes or replaces it.

**Group onboarding (version 14).** No recovery change: Groups, bindings, BusinessProfiles and associations created by
onboarding are ordinary tenancy rows captured and restored with the database, under the existing drill. The two
`business_profiles` provenance columns (`created_xid`, `created_sysid`) are written verbatim by `pg_restore` (their
defaults are not evaluated and the BEFORE UPDATE immutability trigger does not fire on restore), are covered by the
data fingerprint on both sides, and the catalog fingerprint compares two version-14 schemas; `infra/recovery/**` and
the fingerprint SQL are unchanged. Because a restore target must have a different system identifier from its source,
a restored profile can never satisfy the eligibility reader's "inserted by this transaction in this cluster" branch in
the target: it becomes attachable to another Group only through a current `group_owner`/`group_admin` path of a Group
it is already attached to (or by an Organization owner/admin). `TestOnboardingCatalog` R-ONB-1 proves the colliding
transaction-id case deterministically; the real version-14 source→target restore is exercised by B-RECOVERY-DRILL in
the `migrations` suite. A `pg_resetwal`
that moves a cluster's transaction counter backwards is outside the supported recovery path (recorded residual).
