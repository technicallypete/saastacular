> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G013 — Recovery, retention and incident evidence: Implementation Plan

Revision: r3, 2026-10-15 UTC. Status: **review-ready for Codex Plan re-review**. This is not approval, not an execution
binding and not implementation. The G012 predecessor gate is satisfied (final ACCEPT). Execution waits for explicit
Codex Plan ACCEPT, the owner's PG-005 disposition where Codex requires it, and the Git-first execution binding with
input, budget and resource release (§0). Tier: architecture/security/data (Goal Execution Policy).

r3 is the one trust-boundary clarification for the r2 re-review (not included)
(CORRECTION LOOP; review file digest `a5d23177c99334bea9acb8d2c44f7f1ef97f6a749ca1edb5e80380a6000160da`). History is
preserved byte-identically:

- r2: G013-recovery-retention-incident-evidence-plan-r2.md (not included),
  digest `d659406ae20723179365cb8cc265dd18f9081b615f19378833c3b9a50eec8c9e` (its change map answers the
  r1 re-review (not included));
- r1: G013-recovery-retention-incident-evidence-plan-r1.md (not included),
  digest `7b80791b19c36a52694bf892220451835d4abbf0b8f96fd954b9052d0d8ad006` (answers the
  r0 review (not included));
- r0: G013-recovery-retention-incident-evidence-plan-r0.md (not included),
  digest `4822a1430de442391588e9ef745c8c24d6ccc4f1fee66af965ca34d7caaaab36`.

Only the administrator trust boundary of the capture changes, with the wording, manifest, drill assertions and records
that depended on it. The concrete application and provider drain, the fence record written before the first change,
convergent interrupted release, the receipt-bound restore, replay gaps and Q7, the fixture administrator capture, the
query decisions and Q8 (COMPLETE means capture completeness only) are resolved and unchanged. The ten ACs, the four
check IDs, B7 with the checkpoint after attempt 4 and every other reviewed choice are unchanged. No attempt is spent.

**r3 change map:**

| r2 re-review item | Change | Where |
|---|---|---|
| Trusted administrator window | Administrator exclusivity across both stores, PostgreSQL catalog activity included, is a **recorded trusted operational prerequisite**: the operator declares the administrator-exclusive maintenance window with a reference, and capture refuses `admin_window_unavailable` when it is not recorded or is visibly contradicted. The guarantee is stated within that boundary, not against a concurrent privileged administrator violating it. The trace becomes optional supplementary diagnostics, never a completeness or exclusivity proof; the SHARE-lock holder stays as a bounded safeguard only. The drill case that used the trace as proof is replaced by the ordinary precondition refusal | §1, §3.1, §3.2, §3.3, §3.4, §10.2 RD-B, §10.3, §11.2, §14, §15 |
| Dispositions | Q8 accepted; no question open | §15 |

- **Goal, Spec and definition review (bound by SHA-256; unchanged from `5e1e987` to `df3d1b0`):**
  - [Goal](../goals/G013-recovery-retention-incident-evidence.md)
    `62de3c7524def3a7517c46441fbb59b828a5c80f238c58ac85ef49cece97db40`;
  - [Spec](../specs/G013-recovery-retention-incident-evidence-spec.md) (scanner-safe wording correction)
    `c43f8c9746d7b0c072e870699e064dad84b3f0c05ec5f20b0a9376aad0657ba8`;
  - definition review (not included)
    `99827f7bd0b50386a2090038ab902ca4f9e7dba75cbb3fa43c82d93b171cb0b2`.
- **Accepted final G011 (binding inputs):** merge `1af469634225e6e6527aca485b40aa55953227a8`;
  [final review](../results/G011-implementation-codex-review.md)
  `481afdb96a5ff91152f238233967b9026650d04311667aa2996cfb6b0bd73219`;
  [Result](../results/G011-durable-audit-events-result.md)
  `e62069daa0107d82c08bf0f15f63bba47a5e0d8a39c7f3382689e1e0e3b36572`;
  closeout receipt (not included) `95e0cd46…`; Plan r2 `72042f5c…`;
  [AUDIT.md](../../../foundation/AUDIT.md)
  `33a9bac93b867342e81a949681e742b90406d06cdf5471f9e495e4c889c44760`;
  `core/migrations/00008_audit_events.sql`
  `b1dedf1f56f28c97a9f8c59796ff1b615c1ed79e6b8ab8f2cfb602098c597233`; ADR-006 `dc7eda77…`;
  `core/migrations/00007_stored_objects.sql` `28f0a073…`.
- **Accepted final G012 (binding inputs; r1):** MR !28 merge `5e1e98735f6224888889d516ef2397a0ea8dc89a`;
  [implementation review with final acceptance](../results/G012-implementation-codex-review.md)
  `f76a050bd6010121b99c662ef43d29b5f95283e516e71bafa998477a815bdd57`;
  [final Result](../results/G012-secrets-secure-engineering-result.md)
  `e9224e6197daa4d23ac594caaef6ab029bb1422f38a5e2887738030334b6f161`;
  closeout receipt (not included) `f9964e9b…`;
  D0 disposition (not included) `c7dde64f…`; accepted Plan r1 `cd90a790…`;
  execution binding (not included) `260d370d…`;
  [SECURITY.md](../../../SECURITY.md)
  `e6bb1a70ec98225cd4146d9a463d2d5f7fe1ec6f6bd54f0c931f33a1b87d555f`.
- **Planner:** task G013, role plan. Native isolated worktree `.claude/worktrees/agent-<id>` under the
  standing owner fallback for native names (AGENTS.md "Task-prefixed Claude worktrees"); branch `g013-plan`. Effort
  xhigh as instructed; the runtime setting cannot be verified from inside the worker (unknown). **Implementation
  default: high**; the operator may choose xhigh for the migration and replay work.
- **Baseline (Git-first):** r3 is rebased onto `origin/staging` `df3d1b060472e10e554294c78081c1344ab1f0f2` (the r2
  re-review). Its delta from r0's base `5e1e987` is `9f7464c` (the r0 review, PG-005 in docs/PERFORMANCE.md,
  SUPERVISOR-HANDOFF, PROGRESS, a G012 review section), `728165e` (the G012 final review, final Result, closeout
  receipt, SUPERVISOR-HANDOFF, PROGRESS), `b59fcda` (the r1 re-review, a PG-005 clarification, SUPERVISOR-HANDOFF,
  PROGRESS) and `df3d1b0` (the r2 re-review, SUPERVISOR-HANDOFF, PROGRESS): management plus docs records, no code,
  migration, harness or G013 input change. Read-only inspection of files and Git history only: no install, build, scan,
  test, Docker command, probe or harness run, and no subagent was used for planning (r0 to r3).
- **Read:** AGENTS.md, CLAUDE.md, ARCHITECTURE §21/21.1, §40, §56/56.1, §59, §62, §63, DECISIONS, ADR-006 (and the
  ADR index), ROADMAP (EPIC-004 and the privacy handoff), PROGRESS (active mandate and G014 stop), SUPERVISOR-HANDOFF
  (G013 definition and autopilot sections), BUDGET_POLICY, PERFORMANCE (PG-001..004; r1 adds PG-005 and the r0 review
  in full, plus the G012 final review section and final Result changes; r2 adds the r1 re-review and the PG-005 r1
  clarification), the G011 Plan r2 (database, validation, delegation, escalation and handoff sections) and its final
  review, the G012 Plan r1 in full, its two Plan
  reviews, binding confirmation, implementation review, D0 disposition and Result, docs/SECURITY.md, foundation AUDIT,
  DATABASE-ISOLATION, OBJECT-STORAGE, MIGRATIONS, OPERATIONS and FOLLOW-UP, and the code named below. No
  `docs/RISKS.md` exists; PERFORMANCE.md is the register (§14 E8). Core has no service-specific CLAUDE.md or skill; no UI
  work, so no Vestive skill applies.

## 0. Gates and predecessor reconciliation

| Gate | Status at r1 |
|---|---|
| G012 integrated on staging | satisfied: merge `5e1e987` (MR !28) |
| G012 final ACCEPT (staging pipeline, scoped cleanup, final Codex review) | **satisfied** (r1): staging `2949510391` SUCCESS, final review `f76a050b…`, final Result `e9224e61…`, closeout receipt `f9964e9b…` |
| Codex Plan ACCEPT | pending (this r3; r0, r1 and r2 received CORRECTION LOOP) |
| Git-first binding: allow-list (§11.8) and initial Result imported by Codex; current-target reconciliation | pending |
| Owner decisions | PG-005 (Codex recorded it; the owner was asked whether to proceed with bounded measurements and no speculative index): **answer pending, no risk approval** (§14 E8, Q1). No other owner decision is needed by this Plan |
| G014 | out of scope; owner hard stop after G013 (PROGRESS mandate). Nothing here prepares G014 work |

**G011 assumptions (A-G011-n), bound to the accepted final G011:**

- A-G011-1: schema version 8 is the head. `audit.events` (three relations: table, primary key, partial tenant-time
  index), two policies (`events_select`, `events_insert`), three triggers on it, eleven owner-owned invoker functions in
  `audit`; 38 relations, 29 policies and 15 functions in `app`; schema `audit` grants runtime USAGE only.
- A-G011-2: the AUDIT.md "G013 contract" is the starting point: `<db>_retention` holds no audit privilege; G013's
  reviewed migration may add, together, USAGE on `audit`, an expired-row DELETE grant and policy for that role, and an
  `events_refuse` exception limited to that role and eligible rows. Runtime and owner stay refused.
- A-G011-3: the 29-action catalog is closed and unchanged. G013 adds no audit action. Restore, replay and maintenance
  evidence is the operator receipt (the G012 precedent for service-credential rotation), not a catalog event (Q3).
- A-G011-4: capture triggers make every row change of the eight audited tables durable with its event (M1, strict);
  the session revocation trigger records only `revoked_at` transitions; expiry, renewal and login-transaction
  bookkeeping are not audited; `stored_objects` has no DELETE capture trigger and no foreign key points at it.

**G012 assumptions (A-G012-n), bound to the accepted final G012 in r1** (r0 labelled them provisional). The final
acceptance changed no application, migration, harness or SECURITY.md byte after the merge: the merged tree equals the
reviewed source and every later commit is management only, so each assumption holds as written.

- A-G012-1: SECURITY.md as integrated (above hash) is the base for G013's edits: §3 inventory (the retention row and
  the recoverable versus hash-only authentication fields), §5 least privilege (retention "until G013"), §6 rotation,
  §7 at-rest matrix (backups blocked-unknown), §8 F-7, §9 G013 handoff. A later change to these sections is
  reconciled by rebase before any batch.
- A-G012-2: required-check totals after G012: 110 rows; profiles fast 35, integration 74, full 103; full pipeline fast
  35 plus gate 73, 104 distinct IDs (validate-controls lines 1112, 1115, 1206, 1948; OPERATIONS.md lines 93 and 125),
  as the final review's staging evidence confirms. §11.5 recomputes from the then-current totals.
- A-G012-3: G012 adds no migration, role or grant; B-SECRET-ROTATION stays last in `ops`; G013 adds nothing to `ops`.
- A-G012-4: the braces exception expires 2026-10-29 (inclusive). After that date B-DEPSCAN fails every profile until a
  fix or a renewed owner disposition. G013 never edits the exception; a lapse during G013 validation is reported as an
  external blocker for Codex/owner disposition (§14 R-e).

## 1. Implementation approach

One additive privilege migration for the existing retention role, one operator command set built from existing tools,
one paired synthetic drill and the matching documentation. No new runtime service, endpoint, role, image, vendor,
scheduler, daemon or framework.

| Req | Approach (detail in §2–§7) |
|---|---|
| R1 | Finite data-class inventory and responsibility matrix in a new `docs/foundation/RECOVERY.md`; a pinned two-Organization fixture seeded through existing test helpers and real G004/G009 services; labelled fixture policy values |
| R2 | `infra/recovery/recovery.sh backup`: an exclusion and drain held for the whole capture (r1: application services stopped until exited, a database fence revoking CONNECT from every non-superuser login role with their backends terminated, an object fence disabling the application's object service user; r2: the fence record persisted before any change, a SHARE-lock holder on every captured table as a bounded safeguard, a provider restart; r3: a recorded trusted administrator-exclusive window across both stores as the prerequisite, with an optional supplementary trace; refusal when any part cannot be established; endpoint samples and fingerprints supplementary only), superuser `pg_dump` custom archive plus object bytes and binding metadata through the project's built MinIO client image, a versioned private manifest; the completion marker, written last, certifies capture completeness only, and fence release is recorded and reported separately (r2) |
| R3 | `recovery.sh restore`: fail-closed preflight (backup integrity, format/schema/tool compatibility, target identity distinct from the source by resolved containers, volumes and PostgreSQL system identifier, a fresh target for a first restore or, r1, the exact receipt-bound target for a staged continuation, app services stopped), single-transaction `pg_restore` into a target provisioned by the accepted role script with fresh service values, object upload with binding metadata, then catalog/data/object verification. Runtime positives and negatives through the real guarded facade (Go integration phase) |
| R4 | `recovery.sh invalidate-auth` (recovery-only SQL on the closed target); a hash-chained, operator-held **restore-input ledger** (sync from authoritative state plus operator appends), an operator completeness attestation, `recovery.sh replay` (finite restriction kinds, retry-safe; r1: revocation gaps of replayed disablement read before COMMIT and receipted truthfully) and `recovery.sh reopen` (gate) |
| R5 | Migration `00009_retention_maintenance.sql` (retention-only policies, column grants and DELETE on audit events and deleted object tombstones; narrow `events_refuse` exception); `recovery.sh retention` (dry-run default, preview record, apply in bounded batches); tombstone purge only after provider absence and ledger preservation; pending and deleting rows repaired only through the existing guarded delete; `recovery.sh expire` and `quarantine` for backup sets |
| R6 | `recovery.sh status` (bounded, content-free JSON lines, non-zero for anything not healthy) with a deterministic local sink; one end-to-end incident exercise (corrupt latest checkpoint); runbook in RECOVERY.md; observed drill sizes and durations in the Result |
| R7 | Four new required IDs in the existing `migrations` suite, the version-9 reconciliation of accepted pins (the G011 §10.2 pattern), focused batches and one broad final run; staging full gate |

**Rejected alternatives (recorded):**
- *A Go maintenance or recovery command.* It would need a second importer of `core/internal/objectstore` (an accepted
  boundary pin), a new build target or image, and a production deployment story. The operations are provider-tool
  operations (pg_dump, pg_restore, psql, mc) plus SQL whose eligibility is enforced by the database itself (Spec R5:
  "at the reviewed database boundary"). Bash plus SQL follows the G012 `infra/security` precedent.
- *Schema-from-migrations plus data-only restore.* It needs `--disable-triggers` (a superuser trigger bypass on
  restore), ledger exclusions and a second source of catalog truth. A full logical restore creates triggers and
  policies after the data load (post-data section), so no trigger fires and none is disabled.
- *`pg_dumpall` or restoring role passwords.* Cluster roles come from the accepted provisioning contract
  (`10-roles.sh`, or `provision-roles.sql` without init hooks) with fresh values; no globals dump is taken (Spec R2).
- *Deriving replay from audit rows.* Excluded by the Spec ("no speculation that old audit rows reconstruct all later
  decisions"); audit retention may remove them and M5 revocations can be unaudited.
- *Revoking restored sessions with an existing `revoke_reason`.* It would write `session.revoked` events with a
  misleading reason. Expiry invalidates without forging events (§5.1).
- *A recovery-specific audit action or a new admin identity.* A catalog or identity change (Q3); receipts suffice.
- *A new harness suite or a separate source/target copy of the repository.* A new suite touches `SUITES_KNOWN`, every
  profile and many control pins; a new copy adds per-run bind-mount paths. The drill runs in `migrations` and the target
  reuses the migrations copy with its own env file (§11.4).
- *Dedicated new indexes for retention.* Write amplification on every audit insert; bounded measurement instead
  (PG-005; the owner's answer is pending and no cost is accepted, Q1).

**Spec "Resolve before Plan approval" items:**

| # | Spec item | Resolved in |
|---|---|---|
| 1 | Accepted predecessor state, maintenance-role/trigger contract, schema inventory; provisional labels | §0 (A-G011-1..4 and, in r1, A-G012-1..4 bound to the accepted finals); §6.1 |
| 2 | Consistency/quiescence, role/ownership restoration, target identity and refusals, raw-material isolation; tools, versions, commands | §3, §4; tools pinned in §3.5 |
| 3 | Finite policy and hold inputs, eligibility enforcement, batch/cancellation bounds, migration and guard exception | §6 |
| 4 | Invalidation of restored sign-ins and the replay of later restrictions; completeness source and failure; tombstone expiry | §5 |
| 5 | Fixture values, deterministic alert destination, manual feedback, AC mapping, CI triggers, production owner decisions | §2.2, §7, §10, §11, §2.1 |

## 2. R1 — Data classes, fixture and fixture policy

### 2.1 Data-class inventory and responsibility matrix (recorded in RECOVERY.md)

| Class | Authority | In the checkpoint | Recovery dependency | Holds private material | Retention, hold, expiry | Operator | Production prerequisite (owner/deployment) |
|---|---|---|---|---|---|---|---|
| Identity and sessions (`users`, `identity_principals`, `sessions`, `login_transactions`, `identity_provider_binding`) | core PostgreSQL | yes | provisioning, migration ledger; sessions invalidated before reopen | recoverable PKCE verifiers and CSRF values for unexpired rows; hashes; provider identifiers; names (SECURITY.md §3) | sessions expire; login transactions at most 10 min; restore expires all and deletes login transactions | platform operator | backup encryption and access control; WorkOS namespace unchanged |
| Tenant records (Organizations, Groups, bindings, BusinessProfiles, links, policy flag) | core PostgreSQL | yes | replay of later restrictions | names and descriptions | no automatic deletion (product workflows later) | platform operator | real-data decisions (G056) |
| Audit (`audit.events`) | core PostgreSQL, append-only for runtime | yes, exactly | none (history preserved) | identifiers, no payloads (AUDIT.md) | retention role, explicit policy, holds (§6) | platform operator | legal retention periods and holds |
| Object metadata (`stored_objects`) | core PostgreSQL | yes | object bytes and binding | none beyond identifiers | tombstones purged only after absence and ledger preservation (§6.5) | platform operator | as audit |
| Object bytes and binding metadata | object provider (MinIO locally, R2 planned) | yes, with size, digest and binding | metadata rows | tenant content | follow metadata lifecycle | platform operator | provider versioning/immutability, encryption, location (R2 not qualified) |
| Redis | ephemeral (readiness only today) | **no** | starts empty | none persisted | none | platform operator | keep ephemeral (SECURITY.md §7) |
| Operational evidence (logs, receipts, sink, CI artifacts) | operator, GitLab | no | none | sanitized only | CI 90 days; local evidence temporary (AGENTS.md) | platform operator | log retention policy |
| Backup sets | operator private storage | n/a | manifest, ledger, provisioning | everything above | explicit backup policy with holds and quarantine (§6.6) | platform operator | cadence, retention, location, encryption, immutability, RPO/RTO |
| Restore-input ledger and attestation | operator private storage, separate from backups | no (held independently) | none | identifiers only | never pruned in G013; must outlive every recoverable backup | incident owner | durable separate storage |

Product, RAG and AI data are a handoff to their owning Goals (G029–G035, G056), not an early schema.

### 2.2 Pinned synthetic fixture and fixture policy

Seeded by `TestRecoverySeed` through the existing access-package helpers (`setupObjects`-style environment with a
recovery variant that keeps the identity namespace row, `itEnv` primitives, `audPut`, the fake WorkOS through real G004
services, the private MinIO through the real G009 facade):

| Item | Content |
|---|---|
| Organizations, Groups | A with Groups A1 and A2; B with Group B1 |
| BusinessProfiles | P in A linked to A1 and A2 (shared within one Organization); Q in B linked to B1 |
| Users (global identities) | `ownerA` (org_owner A), `memberA1` (group_member A1), `gadminA2` (group_admin A2), `memberA2` (group_member A2), `ownerB` (org_owner B), `cross` (group_member A1 and B1), `disabled0` (disabled before the checkpoint), `disable1` (active at the checkpoint) |
| Authentication | live sessions for the active users; one session revoked by logout before the checkpoint; one live session of `memberA2` and one pending login transaction kept as **old authentication**, written only to a run-owned named volume (never logs or host files) |
| Objects (known bytes and digests) | ready: A1 (by `memberA1`), A2 (by `memberA2`), A Organization-level (by `ownerA`), B1; one tombstone in A1 and one in B1 (out of scope for A); one deleting row in A1 with bytes present (provider delete failure seam); one inert pending row in A1 without bytes |
| Audit | trigger and core events from the above, split by a recorded database timestamp `T_mid` (older events before, sign-ins, denials and URL issuance after); Organization and global rows |
| Clock | fake WorkOS issuer fixed for the run (`SetIssuer`), so restored principals and the namespace row stay valid across processes |

Fixture policy values (all labelled "fixture", never production or legal decisions): audit cutoff `T_mid`; object
cutoff `T_obj` (recorded during the maintenance phase); holds by user (`cross`) and by one object; batch size 2,
max batches 2 then 1000; statement timeout 5 s; lock timeout 1 s; run timeout 120 s; backup max age 3600 s; backup
retention 1 s with an explicit hold on the main checkpoint. Generated values come from `configure-env.sh` per project;
no production or provider value is used.

## 3. R2 — Consistent paired backup and private handling

### 3.1 Command form

`bash infra/recovery/recovery.sh backup --env-file FILE --backup-root DIR --ledger FILE --admin-window REF [--use-fence] [--timeout SECONDS]`

`bash infra/recovery/recovery.sh fence --env-file FILE --backup-root DIR --admin-window REF` and
`bash infra/recovery/recovery.sh unfence --env-file FILE --backup-root DIR [--fence-id ID]` (r1; r2 keeps every fence
record under `DIR/fences/<fence-id>.tsv` so that `status` finds a held fence)

`FILE` is the source project's `.env`-style file (project name, database name, the superuser value, the MinIO
administrator values of the local fixture and the object service user name are read from it; no value appears in argv
or output). `DIR` is an absolute operator-private directory outside the repository; the command creates
`DIR/<backup-id>/` with mode 0700 and refuses an existing set directory. By default `backup` establishes the fence
itself and releases it at the end; `--use-fence` instead requires a fence already established by `fence` (operator
maintenance windows) and leaves it for `unfence`. `unfence` is also the documented recovery after any interruption that
left a fence fully or partly in place (fail-closed: the source then stays unavailable to writers until released).
`--admin-window REF` (r3) records the operator's trusted administrator-exclusive window (§3.2); `REF` is a
non-confidential incident or change reference.

### 3.2 Writer exclusion, drain and capture (r1; drains and lifecycle r2; trust boundary r3)

**Who can write.** In the database: only login roles holding CONNECT on it, which provisioning limits to
`<db>_runtime` (core), `<db>_migrator` (migrate) and `<db>_retention` (this Goal's maintenance), plus the cluster
superuser; PUBLIC holds no CONNECT. In the object store: the application's object service user, whose values only core
and the exited `minio-init` receive (B-COMPOSE), plus the MinIO administrator. Signed URLs are GET only (G009), so no
client writes bypass core. The fence below excludes the non-administrator paths, drains work already accepted by either
store and adds a bounded table-lock safeguard; other administrators are excluded by the recorded trusted window (r3).

**Trusted administrator-exclusive window (r3).** Administrator exclusivity across both stores is an operational
prerequisite, not something the tooling proves. For the whole window, from X0 to the end of capture (with
`--use-fence`, until `backup` has finished), the operator running the command must be the only administrator of the
source PostgreSQL cluster and of the source MinIO deployment: no other superuser or administrator session, job or tool,
no other PostgreSQL catalog activity (DDL, role, grant or setting changes) and no other MinIO administration or write.
The operator records the window with `--admin-window REF`; the reference, the operator and the start and end times are
written to the fence record and the manifest. Capture refuses `admin_window_unavailable` when no reference is recorded,
and aborts `window_changed` when the window is visibly contradicted: another superuser client backend at X0 or at C6, a
failed lock-safeguard check, or another administrator's activity seen by the optional trace or a supplementary sample.
These checks only detect some contradictions; their absence proves nothing. The guarantee holds within the recorded
window; it is not claimed against a concurrent privileged administrator who violates it (privileged administration is
already outside the runtime guarantee, ADR-006).

**Fence record first (r2).** Before any change, X0 writes `DIR/fences/<fence-id>.tsv` (mode 0600), syncs it to disk
and reads it back: fence ID, source identity (Compose project; PostgreSQL container ID, volume and `system_identifier`;
MinIO container ID and volume), the exact prior `pg_database.datacl`, the prior status of the object service user, the
administrator-window record (r3), and the planned steps. Each later step appends one progress line only after it has
completed and verified (for example `db_revoked`, `locks_held <pid> <backend_start> <xact_start> <holder container>`,
`user_disabled`, `provider_restarted <old start> <new start>`, `trace_started <container> <start>` when the optional
trace is used, `window_open`). The original state is therefore never held only in process memory. A record that never
completed its read-back means no change was made.

**Exclusion and drain (established before capture, held until release):**

1. **X1 drain the application:** `docker compose stop core web intelligence` for the source project (graceful
   SIGTERM), then wait until every such container is `exited` (bound 120 s; exit codes recorded). This ends the
   application's own requests; it does not by itself drain work a store already accepted, which X3 and X5 do. A
   container that does not exit refuses `writers_not_stopped`.
2. **X2 database fence:** `REVOKE CONNECT ON DATABASE <db> FROM` exactly the non-superuser grantees present in the
   recorded ACL (normally `<db>_runtime`, `<db>_migrator`, `<db>_retention`); `pg_terminate_backend` every backend of
   those roles (their open transactions roll back; nothing partial commits); wait until none remains (bound 30 s).
3. **X3 database drain and write lock (r2):** a dedicated holder session (a labelled, detached PostgreSQL client
   container on the project data network, connected as the cluster superuser) opens one transaction and takes
   `LOCK TABLE … IN SHARE MODE` on every table of `app` and `audit`, with `lock_timeout` 30 s. SHARE conflicts with the
   row-change and DDL locks, so the grant shows that every write transaction already in progress on those tables, an
   administrator's included, has committed or aborted (drain); while it is held it blocks inserts, updates, deletes,
   truncation and DDL on those tables by any role, while reads, pg_dump's ACCESS SHARE locks and the fingerprints
   proceed. A lock wait that times out refuses `admin_write_in_progress`. The holder then idles in its transaction,
   bounded by the fence's maximum duration. It is a bounded additional safeguard (r3), not a claim to prevent every
   possible privileged catalog change; the trusted window covers those.
4. **X4 object fence:** disable the service user through the MinIO administrator of the local fixture
   (`mc admin user disable`); verify a request with the service values is refused.
5. **X5 provider drain (r2):** restart the source MinIO server (graceful stop until the process has exited, then start
   and wait until healthy). Every operation the provider accepted before the stop has either committed, and is then
   visible to the capture, or was abandoned with the stopped process; none can commit during capture. Because the
   service user is already disabled, a client cannot resume with the application's values after the restart.
6. **X6 optional supplementary trace (r2; diagnostics only in r3):** the command may start a labelled, detached trace
   container (`mc admin trace --json` as the fixture administrator; its output stays in the container log, nothing is
   bind-mounted) and record its container ID and start time. The trace is not an exhaustive write history: a running
   container, a received probe or a settle interval does not show that every call was observed. An observed write or
   administration call aborts the capture as a contradiction of the trusted window; a missing or unreadable trace is
   recorded and does not, by itself, refuse or qualify the capture.

Any step X0–X5 that cannot be established refuses `fence_not_established` and releases what was done.

**Window checks (r2; safeguard and detection only in r3).** Checked at C6 (and, with `--use-fence`, by `backup` at
its start and at C6):

- (a) **PostgreSQL safeguard:** the holder is the same backend (PID and `backend_start` unchanged), its transaction is
  the same (`xact_start` unchanged) and every SHARE lock is still granted (`pg_locks`); otherwise abort
  `window_changed`.
- (b) **MinIO, when the optional trace is kept:** any recorded call outside the read allow-list (list, head, get and the
  command's probes) aborts `window_changed`.

Neither check proves administrator exclusivity; that rests on the recorded trusted window. Other administrators' reads
are not excluded by the window; they cannot change the checkpoint.

**Capture (inside the fence):**

1. **C1 ledger sync and mark:** `ledger sync` (§5.2); the mark is the ledger head (ledger UUID, sequence and entry
   digest).
2. **C2 fingerprint P1:** per table of `app` and `audit`, row count and an ordered row digest; a catalog fingerprint
   (schemas, relations, columns and ACLs, policies, function bodies by digest and settings, triggers, constraints,
   indexes, default privileges, the migration ledger rows); role attributes and memberships.
3. **C3 object listing L1:** name, size and provider tag of every object in the bucket.
4. **C4 database capture:** `pg_dump --format=custom` of exactly `DATABASE_NAME` as the cluster superuser (pg_dump sets
   `row_security=off`; a non-bypassing role makes it fail rather than dump a filtered subset), streamed on stdout to
   `database.dump`; its table of contents must hold only that database's objects (schemas `app` and `audit`, their
   ACLs and the owner's default privileges), exactly the relations of P1; no other database or project is captured.
5. **C5 object capture:** a one-shot container of the project's built `minio-init` image on the project data network,
   acting as the fixture's MinIO administrator because the service user is fenced, copies the bucket into its own
   temporary directory, records per-object Content-Type and binding metadata (`object-id`, `organization-id`,
   `group-id`), and streams one tar on stdout to `objects.tar` plus `objects.meta`.
6. **C6 window checks and supplementary re-check:** checks (a) and (b); then, supplementary only, P2 equals P1, L2
   equals L1 and no client backend other than the command's and the holder's; else abort `window_changed`.
7. **C7 integrity:** sizes and SHA-256 of both components, per-object SHA-256 recomputed from the tar against L1 sizes.

**One COMPLETE meaning (r2).** `COMPLETE` certifies **capture completeness only**: every component written, integrity
verified, and the drains and the fence held with the window checks passing from X0 through C6, within the recorded
trusted administrator window (r3). It is written right after C7, inside the fence, and does
not depend on release, because the captured data is consistent whatever happens to the source afterwards. Fence
release is a separate source state, recorded only in the fence record (`released` with the verified prior state) and in
the backup receipt (`release: verified`, `failed` or `deferred` for `--use-fence`). `status` reports
`backup.fence: held` (not healthy) for every fence record without a verified release, so a held or failed-release
source is never hidden behind a complete set. The integrated `backup` releases after `COMPLETE`; if its release fails
it exits non-zero with `unfence_failed` while the set stays complete. With `--use-fence`, X0–X6 are the separate `fence`
command, `backup` refuses unless the fence record shows `window_open` and the administrator-window record and checks
(a) and (b) pass at its start, and release is the separate `unfence` command.

**Release and `unfence` (r2).** Release reads the fence record, verifies the source identity, and converges each element
to the **recorded prior state**, independent of which progress lines exist, appending a progress line after each:
stop the optional trace container (or any labelled trace container of this fence ID); end the holder's transaction and
remove its container (any labelled holder of this fence ID); set the service user's status to the recorded prior status (a user
disabled before the fence stays disabled); set the database ACL to exactly the recorded prior ACL (re-granting only
grantees present before, never adding one). It then verifies both equal the record and appends `released`. Every step is
idempotent, so `unfence` after an interruption at any point, including during an earlier `unfence`, completes the
release. The provider restart is not reverted (it changes no configuration or data).

**Supplementary checks are not proof.** Endpoint samples (backends, listings, fingerprints) cannot reveal a client that
connected and exited between samples or a change reverted before the second sample. The guarantees come from the drains,
the fence and the recorded trusted administrator window; the lock safeguard, the optional trace and the samples only add
detection.

**Production prerequisites (not implemented):** a managed-PostgreSQL administrator able to revoke CONNECT, terminate
backends and hold table locks; for R2, which has neither a per-user disable in its S3 interface nor a restart, a
conservative, attested operator precondition (every write value revoked, then a settle interval no shorter than the
longest accepted request) plus a separate read-only capture value; and, as locally, a recorded trusted administrator
window across both stores. The command refuses when neither the drain nor such a precondition is available, or when no
administrator window is recorded.

Any failed command (`set -o pipefail`, every `PIPESTATUS` checked), timeout (`timeout` on each step and on the whole
command), signal (trap) or inaccessible source or destination stops the command with a non-zero exit, releases any
fence it established and writes `status=incomplete` with a closed reason; a kill that prevents the trap leaves the
fence record for `unfence`. Pinned bounds: 120 s for X1, 30 s for X2 and for the X3 lock wait, 120 s for the X5
restart, 300 s per other step and 1800 s per backup or restore command by default (`--timeout` lowers or raises the
command bound up to 7200 s); a fence's maximum duration is the command bound (the holder ends its transaction then);
the drill uses 60 s per step and 600 s per command; replay 60 s per entry; status 30 s per probe.

A database snapshot alone is not treated as a paired snapshot: the pair is consistent because every non-administrator
writer is excluded, already-accepted work is drained in both stores, and the whole capture runs inside the recorded
trusted administrator window (Spec R2).

### 3.3 Manifest and publication

`manifest.tsv` (private, versioned, first line `format<TAB>g013-backup/1`): backup UUID, creation time, source identity
(Compose project, PostgreSQL and MinIO container IDs, their volume names, PostgreSQL `system_identifier` from
`pg_control_system()`, database name, server version, bucket), schema version (highest applied migration), tool
versions and image digests, window start and end (database clock and host UTC), the fence ID, the trusted
administrator-window record (r3: reference, operator, start and end), the drain and safeguard evidence (holder backend
and transaction identity with the granted locks, provider restart times; the optional trace's identity and outcome when
used, as diagnostics), supplementary-check outcomes, ledger ID and mark, component files with size and SHA-256,
per-table counts and digests, the catalog fingerprint digest, and one row per object (name, size, SHA-256, content type,
binding fields). The manifest does not record release (r2): release lives in the fence record and the backup receipt
(§3.2). `COMPLETE` is written last and holds the manifest's SHA-256; it certifies capture completeness only. A set
without a valid `COMPLETE` is incomplete and eligible only for scoped cleanup or retry.

### 3.4 Private handling and the privileged boundary

- Raw sets, ledgers, attestations, receipts and env files are private operational material: mode 0700/0600, never CI
  artifacts, committed fixtures or evidence; only sanitized summaries (statuses, counts, sizes, durations, reason codes,
  backup UUIDs) are published. No scanner exemption is added.
- Archives and tars travel only through stdin and stdout; no host directory is bind-mounted (OPERATIONS.md bind-mount
  rule). Container-local temporary copies end with the `--rm` container.
- **Privileged boundary:** capture, restore, invalidation, replay and ledger sync use the PostgreSQL cluster superuser
  (`POSTGRES_SUPERUSER_PASSWORD`, the existing provisioning-administrator boundary; locally via `docker exec` on the
  container's local socket or TCP inside the project network); the X3 lock holder is a superuser session too. In the
  local fixture the object fence, the provider restart, the optional trace and the capture inside the fence use the existing
  MinIO administrator values (r1/r2: the service user is disabled during capture; accepted for the fixture only);
  restore, replay and purge checks use the project's existing scoped object service values (Q2, accepted). Production
  prerequisites (not implemented): an administrator with superuser or BYPASSRLS on managed PostgreSQL; dedicated
  read-only provider values for backup and purge checks, write values for restore, and a way to fence the
  application's write value and to drain and observe the provider (§3.2). Nothing is given to core, web or intelligence.
- Every Docker invocation that records a cidfile uses a distinct, initially absent file (the G012 attempt-3 lesson).

### 3.5 Tools (existing, pinned by the repository)

`pg_dump`, `pg_restore` and `psql` 17 from the compose PostgreSQL image `pgvector/pgvector:0.8.7-pg17` (same major
version as the server; digest recorded per run); `mc` from the project's built `infra/docker/minio.Dockerfile`
`server`/`init` images (digest recorded); host bash 4.4+, coreutils `sha256sum`, `tar`, `timeout`. No new image or
dependency. Exact flag behavior of these versions is confirmed in the first authorized batch; a mismatch with this
Plan stops affected work and returns to Codex (G012 D0 precedent), never relaxes a check.

## 4. R3 — Isolated fail-closed restore and verification

### 4.1 Command form and preflight (before any target mutation)

`bash infra/recovery/recovery.sh restore --backup DIR/<id> --target-env-file FILE [--source-env-file FILE]
--receipts DIR --backup-policy FILE [--steps all|database|objects]`

`--steps` stages a restore so an operator can continue after a provider failure; the receipt stays `partial` until both
steps pass. **First restore versus continuation (r1):**

- A **first restore** (`all` or `database`) requires a fresh target (row "target not fresh" below).
- A **continuation** (`objects`) is accepted only on the exact receipt-bound target and checkpoint: the receipt in
  `--receipts` names this backup (identifier and manifest digest) and this target (container IDs, volume names and
  `system_identifier`, all re-resolved now and equal); its database step is `done` with the database and catalog
  verification passed; the command recomputes the target's data and catalog fingerprints and they still equal the
  manifest; and the target bucket holds only objects of this manifest, each matching its recorded size, digest and
  binding (an earlier partial upload). Verified objects are kept, missing ones are uploaded. Any mismatch refuses
  `continuation_mismatch`. An unrelated non-empty target, or one without a matching receipt, stays refused.
- There is no general resume framework: only this one staged step, bound to one receipt.

Refusals (closed reasons, exit non-zero, target untouched):

| Case | Check |
|---|---|
| incomplete, corrupt, missing or mixed component | `COMPLETE` present and equal to the manifest digest; every component present with the recorded size and SHA-256; tar members equal to the manifest rows and digests |
| expired or quarantined set | backup policy and the set's `EXPIRED`/`QUARANTINED` marker |
| incompatible input | manifest format unknown; server major version differs; manifest schema version above the highest migration in the checkout running the command; target `DATABASE_NAME` differs |
| source/target alias | target project name equal to the source; any target container or volume ID equal to a source one (manifest and, when given, the live source); target `system_identifier` equal to the source's |
| target not fresh (first restore) | any relation in `app` or `audit`, a migration ledger, or any object in the target bucket |
| continuation not bound (objects step) | no receipt, or a receipt for another backup or target, a database step not `done`/verified, recomputed fingerprints not equal to the manifest, or any bucket object outside the manifest or not matching it |
| ordinary access possible | any target `core`, `web` or `intelligence` container running |

There is no reset, overwrite or `down -v` subcommand; disposing of a failed target is an explicit, separate operator
action on a target the operator created (the harness does it only for its own registered projects).

### 4.2 Restore and verification

1. **Roles from provisioning:** the target is a fresh volume initialized by `10-roles.sh` with freshly generated values
   (no old role password). The command verifies role attributes and memberships equal the manifest's role section.
2. **Database:** the archive is streamed into a one-shot PostgreSQL client container; its table of contents minus the
   two schema-creation entries for `app` and `audit` (provisioning created both, empty, owner-owned) is restored with
   `pg_restore --single-transaction --exit-on-error` as superuser. Data load precedes post-data triggers and policies,
   so no audit trigger fires and nothing is disabled. Failure rolls back the whole restore. The database step ends with
   its own verification (data fingerprint per table, catalog fingerprint and migration ledger rows equal to the
   manifest) and is recorded `done` only when that passes (r1).
3. **Objects:** each object is uploaded with its Content-Type, `Content-Disposition: attachment` and the three binding
   fields, then read back (size, binding, SHA-256). A failure after the database step leaves the receipt `partial`;
   a later `--steps objects` continuation (§4.1) finishes on the same target.
4. **Verification:** the database-step results above; object listing and digests equal the manifest; Redis is a fresh
   empty container.
5. **Receipt** `restore.tsv` (private): backup ID, target identity, each step with status and duration. Status
   `restored-closed` only when every step passed; `failed` or `partial` otherwise. The reopen gate refuses anything else.

Runtime verification (the actual restored runtime role through the real guarded facade, §10.2 RD-G `TestRecoveryVerify`)
runs before the reopen decision; superuser fingerprints prove completeness only (Spec R3).

## 5. R4 — Restored authentication and post-checkpoint restrictions

### 5.1 Recovery-only invalidation (`recovery.sh invalidate-auth`)

One superuser transaction on the closed target, refused unless the target's identity matches the restore receipt and
differs from the source:

- every unrevoked session: `idle_expires_at` and `absolute_expires_at` set to the least of their value and the
  database time of invalidation (E1 refuses expired sessions; renewal is unreachable); `csrf_token` replaced with fresh
  random bytes (minimization; the backup copy still holds the old values and stays private);
- every unreturned provider-logout reference: `provider_logout_ref_expires_at` set to the same time (E3 refuses it);
- `DELETE FROM app.login_transactions` (removes recoverable PKCE verifiers);
- postconditions in the same transaction: zero live sessions, zero login transactions, zero unexpired unreturned
  references; else rollback and `failed`.

No `revoked_at` changes, so no `session.revoked` event is forged; expiry and login bookkeeping are unaudited by design
(A-G011-4). Canonical users, principals, bindings and audit history are untouched. The receipt records counts. Not
claimed: revocation of WorkOS sessions or of signed URLs issued by the source (those stay valid against the source
provider until they expire, at most 300 s; the target's fresh provider values do not accept them).

### 5.2 Restore-input ledger (operator-held bounded instruction file)

`ledger.tsv`, held by the incident owner in storage separate from backups. Header: format `g013-ledger/1`, ledger UUID
(the environment's restriction history; it continues after a restore, when the restored target becomes the
environment), database name, creation time. Every backup manifest records the ledger UUID and its mark. Each entry:
sequence (contiguous from 1), SHA-256 of the previous line (a hash chain), recorded time (UTC), kind, exact canonical
UUID references, origin (`sync` or `operator`), and a non-confidential operator record reference.

| Kind | References | Origin | Replay on the closed target (superuser; one transaction per entry) |
|---|---|---|---|
| `object.delete` | Organization, object, recorded deleter | sync from `stored_objects` in `deleting`/`deleted` | IDENTITY-shape context of the recorded deleter (the G009 guard requires an actor); ready or pending → deleting; provider delete and absence check; deleting → deleted |
| `user.disable` | user | sync from `users.status = 'disabled'` | `active` → `disabled` with the recorded time; the revocation-gap marker is reset first and read before COMMIT (r1, below) |
| `policy.disable` | Organization | sync from the policy flag being false | flag set false |
| `membership.remove` | Organization, subject, scope type and ID, role | operator append (privileged primitive today, F-6) | the primitive's exact DELETE predicate (`sqlRemoveBinding`) |
| `membership.downgrade` | as above plus from-role and to-role (lower) | operator append | the primitive's UPDATE predicate with the expected from-role |
| `profile.unlink` | Organization, Group, BusinessProfile | operator append (privileged primitive) | the primitive's DELETE predicate |

Only restrictions are recorded; grants, enables and upgrades never are. Unknown kinds fail closed.

**Sync rule:** `recovery.sh ledger sync` (privileged read) appends entries for currently restrictive states. Object
deletion is terminal, so each object is recorded once (any entry for that object counts). User disablement and the
policy flag are reversible, so they are re-recorded unless an identical entry exists after the latest complete,
unquarantined backup's mark (read from the backup root for the same ledger UUID). Every backup runs sync before its
mark. Hence for any retained backup B: an object recorded after its deletion has an entry after B's mark whenever B
holds it alive, and every reversible restriction observed by a sync after B's window has an entry after B's mark.

**Replay rule:** apply entries with sequence greater than the backup's mark, in order. Entries at or below the mark are
already reflected (the restriction was performed before it was recorded, and the mark is read inside the fenced
capture window); an object recorded while still `deleting` comes back `deleting`, never readable or signable, and is
finished by the §6.4 repair path. Each entry is idempotent: already applied → `already_applied`; a conflicting state
(binding with another role, deleter absent from the checkpoint, unexpected object state) → `conflict`, which blocks
reopening for an operator decision.
A provider failure leaves the object `deleting` (never readable or signable) and the entry `partial`; a rerun resumes.
Replayed changes are recorded by the existing capture triggers: `unknown` actor for privileged SQL, and the recorded
deleter for object transitions (the guard's actor rule; the receipt states that the restore operator executed them).

**Revocation gaps of replayed disablement (r1).** `user.disable` is the only replay kind on G011's revocation-priority
path (M5): the users capture trigger catches an audit-insert fault, adds `user.disabled` to the transaction-local
`audit.revocation_gap` marker and lets the disablement commit. Direct replay does not run `Store.DisableUser`'s
after-revocation step, so the replay transaction carries the marker itself:

1. reset the marker (`set_config('audit.revocation_gap', '', true)`) as its first statement;
2. apply the disablement UPDATE;
3. read `current_setting('audit.revocation_gap', true)` before COMMIT;
4. COMMIT: the disablement stands whatever the marker holds; it is never rolled back to obtain an event;
5. marker empty: confirm the `user.disabled` row exists and receipt `audit: recorded`. Marker set: after COMMIT,
   attempt once, in a separate transaction bounded like G011's independent record (2 s), the existing
   `revocation.unaudited` action in its existing G6 shape (outcome success, reason `disabled`, resource type `user` with
   the resource ID equal to the subject, no Organization or Group, actor `unknown`, since no lifecycle context exists;
   Q7 accepted), then receipt `audit: gap` with `independent: recorded` or `failed`, and a sanitized `replay.last: gap`
   status line. A failed or unknown COMMIT with the marker set gives `independent: not_attempted`; the rerun then reads
   the user's actual state (`already_applied` when disabled).

An earlier gap or unknown-commit record is carried into every later receipt of the same restore: a rerun's
`already_applied` never erases it (accepted disposition, recorded in r2). The receipt never claims the missing
`user.disabled` event exists. The other kinds are strict (M1): an audit fault rolls
the whole entry back, which is reported `failed` with nothing applied, so no gap can arise there. The 29-action catalog
is unchanged.

Post-checkpoint re-grants, re-enables and upgrades are lost by the restore (they are post-checkpoint data); a later
reversal of a recorded restriction makes replay over-restrictive, which is the fail-safe direction, and is listed in the
receipt for the operator to redo after reopening.

### 5.3 Trust boundary and completeness decision (pinned)

- **Trust boundary:** the ledger and attestation are operator-held. The chain makes edits detectable; the backup
  manifest anchors the chain at its mark, so a substituted or forked ledger fails. A malicious operator can still omit
  entries recorded after a mark or attest falsely; that is outside the mechanical guarantee and is the incident owner's
  responsibility.
- **Completeness:** `recovery.sh ledger attest --through-seq N --basis final-sync|last-sync --by REF` writes an
  attestation with the ledger head digest. `final-sync` means sync ran against the source after writers stopped at
  incident start; `last-sync` means the source was unavailable and changes after the last sync are unknown.
- **Reopen blocks** (no autonomous decision): missing ledger or attestation; chain or header failure; mark digest not
  matching the manifest; ledger shorter than the mark (stale); attestation head not equal to the ledger head;
  `last-sync` without an explicit owner risk-acceptance reference; any replay entry not applied. An audit event or age
  never decides completeness.
- **Tombstone expiry:** a tombstone may be purged only when the ledger (UUID matching the retained backups) holds an
  `object.delete` entry for it (§6.5). Sync records objects only once they are `deleting` or `deleted`, so that entry
  follows the mark of every backup that still holds the object alive, and the ledger remains the source after the
  tombstone and its audit events are gone.

### 5.4 Reopen gate (`recovery.sh reopen --sign-off REF`)

Requires the restore receipt `restored-closed`, the invalidation receipt, a complete replay receipt (or zero entries
after the mark) with a valid attestation, and the runtime verification outcome supplied by the operator (the harness
passes the Go phase result). An applied entry with an audit gap does not block reopening (the restriction is in force,
as G011's revocation-priority decision requires); the gate lists every gap and `reopen.tsv` records it under the
sign-off (r1). It writes `reopen.tsv` (time, sign-off reference, receipts digests, acknowledged gaps) and prints the
documented start command. Until then the target's application services stay stopped; starting them anyway is a
privileged-operator action outside the guarantee (recorded residual).

## 6. R5 — Restricted bounded retention and object repair

### 6.1 Migration `core/migrations/00009_retention_maintenance.sql` (explicitly flagged; §14 E1, E2)

Additive privileges, policies and one function; no table, column, index, data or role change. One goose transaction.

- **Preconditions:** the exact version-8 inventories (00008 postcondition strings copied, as 00008 copied 00007's).
- **Function** `app.retention_eligible(p_organization uuid, p_at timestamptz, p_ids uuid[]) RETURNS boolean`:
  owner-owned, SECURITY INVOKER, STABLE, `search_path = pg_catalog, pg_temp`, PUBLIC EXECUTE revoked, EXECUTE to
  `<db>_retention` only. It parses three transaction-local settings on every call and raises `22023` when any is absent
  or malformed: `retention.scope` (`organization:<uuid>` or `global`), `retention.cutoff` (timestamptz strictly before
  `transaction_timestamp()`), `retention.holds` (comma-separated canonical UUIDs, at most 256, the key required even
  when empty). Eligible exactly when the scope matches (`organization_id` equal, or NULL for `global`),
  `p_at < cutoff` and no element of `p_ids` is held.
- **Policies, `TO <db>_retention` only (permissive, command-specific):**
  - `events_retention_select` FOR SELECT and `events_retention_delete` FOR DELETE on `audit.events`, USING
    `app.retention_eligible(organization_id, occurred_at, ARRAY[organization_id, group_id, actor_user_id,
    subject_user_id, resource_id])`;
  - `stored_objects_retention_select` and `stored_objects_retention_delete` on `app.stored_objects`, USING
    `state = 'deleted' AND app.retention_eligible(organization_id, deleted_at, ARRAY[organization_id, group_id, id,
    creator_user_id, deleted_by])`.
- **Grants to `<db>_retention`:** USAGE on schema `audit`; column SELECT on the columns its statements reference
  (`audit.events`: `id`, `occurred_at`, `organization_id`; `app.stored_objects`: `id`, `organization_id`, `state`,
  `deleted_at`), plus any policy-referenced column PostgreSQL 17 proves necessary (pinned by CT-RET-1, recorded in the
  Result); DELETE on both tables. Nothing else: no INSERT, UPDATE, TRUNCATE, REFERENCES, TRIGGER, DDL, membership,
  BYPASSRLS or other table.
- **`events_refuse` body replaced** (same name, owner, ACL and three triggers): it returns `OLD` only when
  `TG_OP = 'DELETE'`, `current_user` equals the literal `<db>_retention` (embedded at migration time with `format`) and
  `app.retention_eligible(...)` holds for `OLD`; every other UPDATE, DELETE and TRUNCATE by every role, owner and
  superuser included, still raises `append-only`.
- **Postconditions:** the exact version-9 inventories (relations unchanged; four new policies; the function allowlist
  plus one; ACLs; the refuse-function body digest; schema privileges `audit` = runtime and retention USAGE).
- **Down (development and test only):** drops the four policies and the function, revokes the grants, restores the
  version-8 `events_refuse` body; allowed with rows present (no data is involved); postconditions are the exact
  version-8 inventories.
- **Risk:** `CREATE POLICY` takes a brief ACCESS EXCLUSIVE lock on `audit.events` and `app.stored_objects`: migrate in
  a quiescent window (deploy order provision → migrate → code, as for 00006–00008). 00001–00008 stay byte-identical
  (B-MIGRATE-SET adds the 00008 digest).

**What the retention login can and cannot do (malicious-operator residual):** within one transaction it may declare any
single scope, any past cutoff and any hold set, and then delete audit rows and `deleted` tombstones matching that
declaration, reading only the minimal columns of those rows. It cannot read other rows, insert or alter audit events,
update or truncate anything, touch other tables, change roles or grants, or bypass row-level security; runtime and the
owner remain unable to delete. A superuser or the owner after a privileged policy change stays outside the claim
(ADR-006). Scope and hold decisions are the trusted policy's; the database guarantees deletion never exceeds the
declared policy and never happens without one.

### 6.2 Policy file (operator-owned, strict `KEY=VALUE`)

Unknown, duplicate or missing keys are refused (exit 2); there are no defaults for periods or holds.

| Key | Meaning | Bound |
|---|---|---|
| `policy_id`, `policy_label` | identifiers (labels such as `fixture` are informational) | required |
| `scope` | `organization:<uuid>` or `global` | required |
| `global_authorization` | owner authorization reference | required exactly when `scope=global` |
| `audit_cutoff` or `audit_retain_for_s` | absolute UTC time, or seconds before the database clock at run start | exactly one when audit retention runs |
| `object_cutoff` or `object_retain_for_s` | as above for tombstones (`deleted_at`) | exactly one when object purge runs |
| `holds` | comma-separated UUIDs (Organizations, Groups, users, objects, resources); may be empty, never absent | ≤ 256 |
| `batch_size`, `max_batches` | per-run bounds | 1–500; 1–1000 |
| `statement_timeout_ms`, `lock_timeout_ms`, `run_timeout_s` | timeouts | 100–30000; 100–10000; 1–3600 |

### 6.3 Audit retention (`recovery.sh retention audit --policy FILE --env-file FILE [--apply --preview FILE]`)

- **Dry-run (default):** resolves the cutoff once from the database clock, sets the three settings with `set_config(...,
  true)` inside one read-only transaction as the retention login (TCP, value passed by name through `docker exec -e`),
  counts eligible rows and writes a private preview record (policy SHA-256, resolved cutoff, scope, eligible count).
  Nothing changes.
- **Apply:** requires the preview record; refuses when the policy file's digest changed; reuses the preview's resolved
  cutoff (eligibility can only shrink); each batch is its own transaction with `SET LOCAL statement_timeout` and
  `lock_timeout`, deleting `WHERE id IN (SELECT id … ORDER BY occurred_at, id LIMIT batch_size)`; the database
  re-evaluates eligibility at DELETE time (revalidation). It stops at `max_batches`, at the run timeout or on a signal
  after the current batch, and refuses when the current eligible count exceeds the preview count.
- **Receipt:** batches, deleted counts, remaining eligible count, outcome `complete`, `partial` (bound reached or
  failure) or `failed`, never a false completion. Hold, recent and out-of-scope rows are not visible to the role.
- **Query shapes (r1).** The statements carry explicit scope and time predicates, bound to the same validated values as
  the transaction-local settings. The policy's `app.retention_eligible` remains the security boundary (it re-checks
  scope, cutoff and holds on every row); the explicit predicates only shape the plan and can never widen deletion,
  because both must hold.

  ```sql
  -- Organization scope (dry-run counts use the same WHERE clause)
  DELETE FROM audit.events WHERE id IN (
    SELECT id FROM audit.events
    WHERE organization_id = $org AND occurred_at < $cutoff
    ORDER BY occurred_at, id LIMIT $batch);
  -- Global scope (separately authorized)
  DELETE FROM audit.events WHERE id IN (
    SELECT id FROM audit.events
    WHERE organization_id IS NULL AND occurred_at < $cutoff
    ORDER BY occurred_at, id LIMIT $batch);
  ```

  **Index expectations (labelled, not assumed).** `events_tenant_time_idx (organization_id, occurred_at, id)` is partial
  (`WHERE organization_id IS NOT NULL`). The Organization statement's strict equality on `organization_id` is what
  allows the planner to consider it; whether it is chosen, given the row-level-security barrier and the non-leakproof
  eligibility function evaluated as a filter, is **unknown until E-RET-1**. The global statement cannot use that index
  (its rows are excluded), so a sequential scan per batch is expected (E-RET-2): bounded batches limit work per
  invocation, not the total scan cost, which grows with table size times batches. Held rows are filtered after they are
  read, so heavily held scopes read more rows per deleted row. No new index is proposed (PG-005, owner answer pending).

- Pending age never proves an upload inert: maintenance never selects pending or deleting rows; `status` reports their
  counts (privileged read) as `pending_review`.
- Repair is the existing guarded `DeleteObject` by an authorized actor (G009): an authorized repeat delete re-runs
  removal for `deleting` and `deleted` rows; a provider failure or unknown outcome stays `deletion_pending`. G009's
  unknown-commit guarantees are reused, not replayed.

### 6.5 Tombstone purge (`recovery.sh retention objects …`, same dry-run/preview/apply contract)

Per batch, outside any database transaction: (1) select eligible tombstones as the retention login; (2) skip
`not_preserved` unless the ledger holds an `object.delete` entry for the object (§5.3); (3) provider check with
the project's object service values (local fixture): `404` → absent; present bytes → skip `bytes_present` (repair by an
authorized repeat delete); any other outcome → skip `provider_unavailable`; (4) DELETE only the confirmed IDs (eligibility
re-evaluated). Ready, pending, deleting, recent, held and out-of-scope rows are never visible. No audit row is deleted by
object purge and nothing cascades (no foreign keys point at `stored_objects` or from `audit.events`).

Query shape (r1), with the same boundary rule as §6.3: candidates are selected with
`WHERE organization_id = $org AND state = 'deleted' AND deleted_at < $cutoff ORDER BY deleted_at, id LIMIT $batch`,
and the DELETE names only the confirmed IDs (`WHERE id = ANY($confirmed)`, eligibility re-checked by the policy).
`stored_objects_organization_group_idx (organization_id, group_id)` may serve the Organization predicate, with
`state`/`deleted_at` filtered and a sort on `deleted_at`; the actual plan is **unknown until E-RET-3**.

### 6.6 Backup expiry, holds and quarantine

The backup policy file uses the same strict `KEY=VALUE` rules: `policy_id`, `backup_max_age_s`, `backup_retain_for_s`,
`backup_holds` (backup UUIDs; may be empty, never absent) and `pending_review_after_s`, all required.

- `recovery.sh expire --backup-root DIR --backup-policy FILE [--apply]`: a set is eligible when older than
  `backup_retain_for_s`, not listed in `backup_holds` and not quarantined. Apply deletes the set's files in a fixed
  order and writes `EXPIRED` first; any failure leaves `EXPIRED-PARTIAL` with the remaining items listed; the receipt
  always states that only the local copy is managed and that provider, replicated or immutable copies are unknown.
- `recovery.sh quarantine --backup DIR/<id> --reason CODE`: marks a set; restore refuses it; expiry keeps it.
- Restore refuses expired, partially expired and quarantined sets. Restoring a copy found elsewhere is outside the tool
  and needs an owner decision; the runbook requires replay and a retention dry-run on that target before reopening.

## 7. R6 — Signals, incident exercise and runbook

### 7.1 Status hook (`recovery.sh status --backup-root DIR --backup-policy FILE --receipts DIR [--env-file FILE] [--sink FILE]`)

One JSON line per check (`time`, `check`, `status`, closed `reason`, optional backup UUID; no path, name, identifier
of tenants, value or row content) and a summary line:

| Check | Status values |
|---|---|
| `backup.latest` | `ok`, `missing`, `stale` (older than `backup_max_age_s`) |
| `backup.integrity` | `ok`, `failed` (latest complete, unquarantined set re-verified), `unknown` |
| `backup.fence` (r1) | `ok`, `held` (r2: any record under `DIR/fences/` without a verified `released`, including a partly established or partly released fence) |
| `restore.last`, `replay.last` | `ok`, `failed`, `partial`, `none`; `replay.last` also `gap` (r1: an applied disablement with an audit gap not yet covered by a reopen sign-off) |
| `maintenance.last` | `ok`, `partial`, `failed`, `none` |
| `maintenance.pending_review` | `ok`, `pending` (deleting rows, pending rows older than `pending_review_after_s`, tombstones without ledger entries; counts only; requires `--env-file`) |
| `policy` | `ok`, `unknown` (missing or invalid policy) |

`none` (never run) is a known state and counts as healthy for restore, replay and maintenance; `pending_review` runs only
with `--env-file`. Exit 0 only when every check is `ok` or `none`; 1 otherwise; 2 for invalid input. Every probe is
bounded (30 s). `--sink` appends the same lines to a mode-0600 local file: the deterministic local destination (Spec
R6). No metrics stack, collector,
daemon, pager or notification account.

### 7.2 Incident exercise (one, end to end, automated)

Detect: after the main checkpoint B1 and the post-checkpoint changes, a second checkpoint B2 is taken and one byte of
its database archive is flipped (simulated storage corruption); `status` reports `backup.integrity failed` (exit 1)
and the sink receives the non-success line. Contain: `quarantine` B2; `restore` from B2 is refused. Triage:
`verify` names the failed component. Recover: restore B1 into the target, invalidate, replay the entries after B1's
mark, verify, reopen with a sign-off reference; take a fresh checkpoint B3 of the reopened target; `status` returns
`ok` and the sink records it. The audit table shows no event from backup, status, quarantine or invalidation, and only
the replay events from replay, with the injected disablement gap receipted rather than claimed (§5.2; logs and audit
match actual effects). No external message is sent.

### 7.3 Runbook content (RECOVERY.md)

Detection (status, receipts, `audit.gap`/`audit.record_failed` diagnostics, readiness 503); incident owner and
escalation (platform operator; the product owner for legal and customer notification decisions and deadlines, which
the tooling never decides); scoped containment (stop writers, keep targets closed, quarantine sets, rotate service
values per SECURITY.md §6 when exposure is suspected); private-material handling and evidence access with redaction;
restoration, replay and runtime verification (gate checklist); rollback and quarantine (the source is never modified
by backup or restore; a target can be discarded); reopen sign-off; post-incident fresh checkpoint. The procedure maps
to actual G011 events and diagnostics and never requires a successful audit append to undo a security revocation.

### 7.4 Observed measurements (one drill, recorded in the Result)

Rows per table, archive and tar bytes, object count, window start and end, backup, restore, verification, invalidation,
replay and reopen durations, the recovery point (B1's window end) and its data-loss meaning (post-B1 changes are lost
except replayed restrictions), and the environment (local Docker Desktop/WSL or hosted DinD, image digests). They are
observations, not RPO/RTO targets; production cadence, RPO/RTO, alerting, backup location, encryption and processor
responsibilities stay explicit owner/deployment decisions.

## 8. Files and modules

| Path | Change |
|---|---|
| `core/migrations/00009_retention_maintenance.sql` | **new** (§6.1) |
| `core/internal/migrate/retention_integration_test.go` | **new**: TestRetentionCatalog (CT-RET), TestRetentionMigration (MG-RET) |
| `core/internal/tenancy/access/retention_integration_test.go` | **new**: TestRetention (RT, E-RET) with the actual retention login |
| `core/internal/tenancy/access/recovery_integration_test.go` | **new**: the drill's Go phases (seed, post-checkpoint changes, verify, repair) reusing the package's helpers; a recovery environment variant that keeps the namespace row |
| `infra/recovery/recovery.sh`, `infra/recovery/lib.sh` | **new**: subcommands `fence`, `unfence` (r1), `backup`, `verify`, `restore`, `invalidate-auth`, `ledger {init,append,sync,verify,attest}`, `replay`, `reopen`, `retention {audit,objects}`, `status`, `quarantine`, `expire` |
| `infra/recovery/sql/*.sql` | **new**: fence and release statements, fingerprints, supplementary checks, invalidation, replay statements (incl. the revocation-gap read), retention statements with explicit predicates, status counts |
| `infra/recovery/testdata/…` | **new**: stack-free policy and ledger fixtures (synthetic, no values) |
| `infra/validation/run.sh` | four new check functions in `suite_migrations`; `PORTS_REC`; B-SHELL list; `APPLICABLE` (+ `infra/recovery`, RECOVERY.md); retention value in the IT env file; version-9 reconciliation (§11.3) |
| `infra/validation/required-checks.tsv` | +4 IDs; header totals |
| `infra/validation/render-customizations.tsv` | FOLLOW-UP.md reason text (row count unchanged) |
| `infra/ci/validate-controls.sh` | count pins only (§11.5) |
| version-9 reconciliation | `core/migrations/migrations_test.go`; `core/internal/migrate/{integration,audit_integration,runner}_test.go`; `core/internal/tenancy/integration_test.go`; `core/internal/tenancy/access/integration_test.go` (§11.3) |
| `docs/foundation/RECOVERY.md` | **new** guide and runbook |
| `docs/SECURITY.md` | §3 retention consumer and recovery operator values; §5 retention row; §7 backups at rest; §8 F-7 disposition; §9 handoffs |
| `docs/foundation/{AUDIT,OBJECT-STORAGE,DATABASE-ISOLATION,MIGRATIONS,OPERATIONS,FOLLOW-UP}.md` | G013 contract delivered; tombstone/pending handoff delivered; retention role row; version 9; new checks, counts and the recovery pointer; STAGE-OPS/STAGE-AUDIT status |

**Unchanged:** production Go code, `core/cmd/*`, `core/internal/objectstore` and its boundary pins, provisioning
scripts, compose files, `.env.example`, `configure-env.sh`, Dockerfiles, VENDORS.md (no new integration),
`.gitlab-ci.yml`, `infra/ci/{lib.sh,profile-paths.tsv,select-profile.sh,validate.sh}`, PERFORMANCE.md (Codex recorded
PG-005 in r1's base), ARCHITECTURE.md and every Codex-owned management file.

## 9. Schema, contracts and interfaces

- **Migration:** 00009 only (§6.1): privileges, four policies, one function, one function body; reversible; no data.
- **Public API, MCP, BFF and admin UI: none.** No route, no product schema, no new identity.
- **Operator interface (new, developer-facing):** `infra/recovery/recovery.sh` subcommands, the policy file, the ledger,
  attestation, receipts and status lines (formats versioned and documented in RECOVERY.md).
- **Database contract change for a role:** `<db>_retention` gains a consumer and the §6.1 privileges (closes F-7).

## 10. Tests and acceptance-criteria mapping

### 10.1 Retention boundary (Go, real PostgreSQL 17)

- **TestRetentionCatalog** (B-RETENTION-CATALOG): CT-RET-1 the exact version-9 delta (function, policies, column and
  table ACLs, refuse-function body digest, schema privileges), everything else equal to version 8; CT-RET-2 retention
  attributes (NOINHERIT, NOBYPASSRLS, no membership) and runtime, owner and helper ACLs unchanged; CT-RET-NEG rolled-back
  detectors, each yielding its code with the catalog unchanged after: a retention UPDATE grant, a retention INSERT grant
  on `audit.events`, a retention TRUNCATE grant, a retention policy also `TO` runtime, a refuse function without the role
  check, a SECURITY DEFINER eligibility function, PUBLIC EXECUTE on it, a retention policy on another `app` table.
- **TestRetentionMigration** (B-RETENTION-CATALOG): MG-RET-1 empty to 9; MG-RET-2 a populated version-8 world to 9 with
  every row and non-delta ACL identical; MG-RET-3 down to 8 with rows present restores the exact version-8 catalog and
  rows; MG-RET-4 up again and reapply; MG-RET-5 a rolled-back up leaves version 8.
- **TestRetention** (B-RETENTION-IT), actual retention login on a world with old, recent, held, foreign and global rows:
  RT-1 absent settings raise on SELECT and DELETE; RT-2 malformed scope, future or present cutoff, oversized or invalid
  holds raise; RT-3 Organization scope: visible count equals the expected eligible set, DELETE removes exactly it, and
  recent, held, foreign and global rows remain (superuser check); RT-4 global scope deletes only eligible global rows;
  RT-5 UPDATE, TRUNCATE, INSERT, ALTER, CREATE, GRANT, SET ROLE, trigger disable and `row_security=off` refused with rows
  and catalog unchanged; RT-6 runtime and owner DELETE, UPDATE and TRUNCATE still refused or zero rows, and a superuser
  DELETE still raises `append-only` (the exception is role-bound); RT-7 tombstones: only eligible `deleted` rows are
  visible and deletable; ready, pending, deleting, held, recent and foreign rows untouched; runtime still has no DELETE;
  RT-8 revalidation: a hold added between count and delete preserves the row; RT-9 E-RET-1..3 (§11.6) on a pinned
  synthetic distribution (r1), generated by trigger-owned row changes, the global anonymous sign-in-failure shape
  already used by AT-PRIV and superuser-seeded tombstone rows without bytes, then `ANALYZE audit.events,
  app.stored_objects`: Organization A about 4,000 events before `T_mid` and 1,000 after, about 200 of them naming a
  held user; Organization B about 1,000; global about 2,000 before and 500 after; about 200 tombstones in A (50 held,
  50 after the cutoff) and 50 in B. The exact counts are recorded with each plan; this is a fixture for observing
  plans, not a benchmark.

### 10.2 Operator drills (harness functions in `suite_migrations`, after the existing checks)

**B-RECOVERY-DRILL** (source = the migrations project reset to fresh volumes and migrated to 9; target = a second
project on the same copy with its own env file):

| Step | Content and required outcome |
|---|---|
| RD-S | stack-free tool cases: policy parser refusals (missing `holds`, unknown and duplicate keys, both cutoff forms, future cutoff, oversized holds) exit 2; ledger chain verify passes, tampered entry, truncated (stale) ledger and forked header fail |
| RD-B | seed; then fence and capture cases. (1) **Administrator write in progress (r2):** a superuser transaction holding an uncommitted UPDATE on a captured table makes the X3 lock wait time out; `fence` refuses `admin_write_in_progress` and releases to the exact recorded prior state. (2) **In-flight writes drained:** a runtime session holding an uncommitted write is terminated by `fence` (r1: rolled back, absent from the checkpoint and the source); and (r2) a PUT with the application's object values whose body the harness holds open is in flight when `fence` runs: after X4 and the X5 restart it has not committed, the object is absent from the source and from B1. (3) **While fenced:** a runtime login and a PUT with the application's object values are refused; (r2) a superuser UPDATE on a captured table blocks behind the SHARE-lock safeguard until its own statement timeout, changing nothing; `status` reports `backup.fence: held` (exit 1). (4) **Capture and release (r2 COMPLETE rule):** `backup --use-fence` produces B1 with `COMPLETE`, its receipt says `release: deferred` and `status` still reports `backup.fence: held`; `unfence` restores the exact prior ACL and service-user status, the record shows `released`, `status` reports `backup.fence: ok`, and a runtime login and a PUT succeed again; source fingerprint unchanged by the backup. (5) **Trusted window prerequisite (r3; replaces the r2 trace-as-proof case):** `fence` and the integrated `backup` without `--admin-window` refuse `admin_window_unavailable` before any change (no fence record, ACL and service user unchanged, no `COMPLETE`). Negative backups each exit non-zero with `incomplete` and no `COMPLETE`, releasing any fence they set: `--use-fence` without an established fence, stopped source PostgreSQL (inaccessible source), an existing or unwritable set directory (inaccessible destination), a paused source MinIO (X4 step timeout, `fence_not_established`: the database elements are released at once, the MinIO element by `unfence` after the pause ends), TERM while the capture runs (cancellation, fence released), and a capture container that fails while its stream consumer succeeds (pipeline failure, via an invalid bucket in a copy of the env file) |
| RD-K | **Fence lifecycle interruptions (r2):** (1) `fence` is killed with SIGKILL (no trap) while blocked at X4 against a paused MinIO, after the record shows `locks_held`; the holder container is still running; after the pause ends, `unfence` stops the holder, leaves the service user at its recorded prior status, restores the exact prior ACL and records `released`, and the prior state is verified exactly. (2) After a full `fence`, `unfence` is killed with SIGKILL while blocked restoring the service user against a paused MinIO (trace and holder already stopped); `status` reports `backup.fence: held`; a second `unfence` completes and the prior state is verified exactly. (3) **Prior state preserved:** with the service user already disabled and the retention role's CONNECT already revoked before the fence, `fence` then `unfence` leave the user disabled and the grant absent (nothing enabled or broadened); the harness then restores its pre-case state |
| RD-P | post-checkpoint changes on the source (all six kinds: remove `memberA1` from A1, downgrade `gadminA2`, unlink P from A2, disable `disable1`, set A's policy flag false, `ownerA` deletes the A Organization-level object); operator appends; sync; tombstone purge of that object on the source (ledger-preserved); B2 |
| RD-I | incident (§7.2): B2 corrupted, `status` exit 1 with the sink line, quarantine, restore refused |
| RD-R | restore refusals before mutation: alias (target = source), non-empty target, missing component, corrupt component, mixed set, unknown format, incomplete set, quarantined set; the target stays empty |
| RD-F | staged restore of B1 on the same owned target (r1): `--steps database` succeeds with its database and catalog verification; the objects step against a stopped target MinIO fails, receipt `partial`, gate refuses; continuation with a receipt naming another checkpoint (B2), and an objects step aimed at a non-empty target the receipt does not name (the source), are refused (`continuation_mismatch` or alias, nothing changed); MinIO restarted, `--steps objects` continues on the same target, re-verifies the receipt binding and fingerprints, uploads the missing objects and ends `restored-closed` |
| RD-V | the RD-F target: fingerprints, catalog, ledger rows and objects equal the manifest; source fingerprint unchanged; source runtime value refused on the target; Redis empty |
| RD-A | `invalidate-auth` receipt and postconditions |
| RD-L | gate refusals: before replay; missing ledger; tampered chain; stale ledger; missing attestation; `last-sync` without owner reference |
| RD-Y | first replay with the target MinIO stopped and an **injected audit fault (r1):** a test-only BEFORE INSERT trigger on the disposable target's `audit.events` that raises for `user.disabled` only, created by the harness as superuser; `disable1` is disabled and stays disabled, its entry is receipted `audit: gap` with `independent: recorded` (one `revocation.unaudited`, actor `unknown`), `status` reports `replay.last: gap`; the object entry is `partial` and the gate refuses. The fault trigger is dropped and the catalog fingerprint is re-verified equal to the manifest; MinIO restarted, replay again: all applied, earlier entries `already_applied` |
| RD-G | `TestRecoveryVerify` on the target (actual runtime role, real facade): V1 restored relationships; V2 runtime positives incl. signed fetch and binding; V3 cross-Organization and cross-Group negatives incl. shared-profile non-transitivity; V4 owner audit read of restored history, runtime audit UPDATE/DELETE/TRUNCATE and trigger-owned inserts refused; V5 SET ROLE, CREATE and ALTER refused; V6 old session, pending login callback and logout reference refused, `disabled0` denied, the revoked session stays revoked; V7 fresh synthetic login works; V8 each replayed restriction denies its old authority, the deleted object is neither readable nor signable and its bytes are absent, B and `cross`'s B access unchanged, P and the removed user's global row remain; V9 replay events present with the documented attribution (actual restore operator versus recorded original deleter) and no other new events; for `disable1` no `user.disabled` event is claimed or present, the gap receipt and the `revocation.unaudited` record are (r1) |
| RD-O | reopen decision written with the sign-off reference, listing the acknowledged `disable1` gap; B3 of the reopened target (same ledger UUID; integrated fence, `COMPLETE`, then receipt `release: verified`, r2); `status` without the database check exits 0 (`replay.last: ok` once the sign-off covers the gap, `backup.fence: ok`) and the sink records recovery |
| RD-M | `recovery-drill.tsv` (sanitized measurements, §7.4) published to the report directory |

If the recovery drill fails before RD-O, it removes its private material at once (MD-C's assertions) and the
maintenance drill is recorded UNMET.

**B-MAINTENANCE-DRILL** (on the reopened target; UNMET, never PASS, when the drill did not reach RD-O):

| Step | Content and required outcome |
|---|---|
| MD-A | audit: dry-run leaves the fingerprint unchanged; apply with `max_batches=2` ends `partial`; a changed policy file refuses the preview; apply completes; held (`cross`), recent, B and global rows preserved; global scope without `global_authorization` refused, with it deletes only eligible global rows; a held row lock (superuser session) makes a batch fail on `lock_timeout` truthfully, retry succeeds |
| MD-R | `TestRecoveryRepair` (Go, authorized actor through the facade): repeat delete of the seeded deleting row with a provider fault stays `deletion_pending`, without the fault becomes `deleted`; the inert pending row is deleted by an authorized delete; a ready object is deleted without a ledger entry (tombstone T) |
| MD-O | purge before sync: T `not_preserved`; after sync with the target MinIO stopped: `provider_unavailable`, nothing removed; bytes planted under the restored seed tombstone: `bytes_present` until an authorized repeat delete (a second Go repair step); then the eligible tombstones are purged, while the held tombstone, a tombstone deleted after the resolved cutoff (recent), B1's tombstone (out of scope) and every ready row remain; pending and deleting rows are never selected (RT-7) |
| MD-E | expiry: held B1 and quarantined B2 kept; hold released, B1 expired with the "other copies unknown" statement; an obstruction (a directory in place of a component file) gives `EXPIRED-PARTIAL`, restore refused; rerun completes; restore of the expired set refused |
| MD-S | `status --env-file` reports `pending_review: pending` before MD-R (seeded deleting and pending rows) and `ok` after MD-R/MD-O; a missing policy gives `unknown`, a one-second maximum age gives `stale` and an empty backup root gives `missing` (each exit 1); the last receipts are reported truthfully |
| MD-C | the private directory and the named volume are removed; no backup, ledger, attestation, receipt or env file remains under the work directory |

### 10.3 Acceptance criteria

| AC | Evidence |
|---|---|
| AC1 | §2.1 matrix; §2.2 fixture; RD-B (the writer exclusion and drain held for the whole capture and released exactly, the terminated in-flight writer and the refused late writers; r2: the drained in-flight provider PUT and the administrator write in progress refused by the lock safeguard; r3: capture refused without a recorded trusted administrator window; complete B1 with version, completeness and integrity records under the single COMPLETE meaning; partial and failed captures never complete; source unchanged apart from RD-P's planned changes); RD-K (fence record persisted before change, exact recovery after interrupted fence and unfence, prior state never broadened) |
| AC2 | RD-R refusals before mutation, RD-F partial failure with the target closed and the receipt-bound continuation completing on the same owned target while mismatched continuations are refused, distinct owned target, sanitized evidence only |
| AC3 | RD-V fingerprints, catalog, roles from provisioning, ledger rows, object digests and binding; RD-G V1–V7 (actual runtime positives, cross-tenant and Group negatives, audit and privilege refusals, fresh service values, old authentication refused, fresh login) |
| AC4 | RD-P/RD-L/RD-Y/RD-G V8–V9: ledger survives the purged source tombstone, six restriction kinds reapplied before reopening, old authority and new signed issuance denied, obsolete bytes absent, retry and partial truthful, the injected replay audit fault leaves the user disabled with a truthful gap receipt, missing/untrusted/incomplete inputs keep the target closed, unrelated tenant unchanged |
| AC5 | B-RETENTION-CATALOG (CT-RET, MG-RET), B-RETENTION-IT (RT-1..9 with the actual retention login), MD-A (dry-run, bounded apply, preview revalidation, holds, scope, timeouts) |
| AC6 | MD-O, MD-R, MD-E: purge only after absence and ledger preservation, controls preserved, pending/deleting repair truthful, expiry with holds and honest partial deletion, no audit cascade |
| AC7 | RD-I and RD-O: non-success status and sink line, containment, triage, recovery, reopen sign-off; `unknown`, `stale` and `missing` are not healthy (MD-S); `replay.last: gap` until signed off (RD-Y) and `backup.fence: held` while a fence is unreleased or partly released (RD-B, RD-K) are not healthy; audit and logs match effects |
| AC8 | RECOVERY.md matrix and runbook; RD-M measurements and recovery-point meaning in the Result; production responsibilities explicit; no compliance claim |
| AC9 | §11: focused plus path-selected final coverage (full + ci-controls), exact commits and image digests per batch, failure history, raw material never published (MD-C, B-SECRET-CANARY, evidence scans), B-REQUIRED-CHECKS, minimal wiring |
| AC10 | S-review, Result with AC map, B7 ledger and checkpoint 4, scoped normal and interrupted cleanup receipt, staging disposition, Codex final ACCEPT; EPIC-004 compaction is Codex's; handoffs §16 |

## 11. Validation

### 11.1 Before every batch

`git fetch origin`; apply the exact-delta rule (record SHAs and reason) or rebase and reconcile, including G012's final
ACCEPT delta. A management-only delta is still an input to the repository-wide secret scans: run the README secret-scan
line on the rebased tree before carrying evidence forward. Then `bash infra/ci/docker-desktop-mounts.sh` (stop on
failure), take the harness flock, record target and candidate SHAs. The Result and allow-list stay at base bytes while
validating (B-PROTECTED); `docs/SECURITY.md` uses `--expect-hash`. Standard form: `bash infra/validation/run.sh --base
<target> --allow-file management/v0/results/G013-scope-allowlist.txt --expect-hash docs/SECURITY.md=<sha>
--report-dir <EVID>/<batch> <suites>`.

### 11.2 Batches

| Batch | Content | Purpose | Wall (forecast) |
|---|---|---|---|
| b0 tools | first authorized use of pg_dump/pg_restore/psql/mc forms on a focus pair (owned, in the run inventory): flags, archive TOC filtering, metadata on upload, and the fence primitives (r1: CONNECT revoke and backend termination, `mc admin user disable` and its exact restore; r2: the SHARE-lock holder and its `pg_locks` evidence, the graceful MinIO restart, and, if the optional trace is kept, its `mc admin trace --json` record format); no repository change | confirm §3.5 contract; mismatch returns to Codex | 10–15 min |
| b1 static | core check image (gofmt, vet incl. integration tag, unit: migration set, runner), shellcheck through `run.sh lint` (B-SHELL, B-MIGRATE-SET, B-SECRET-SOURCE), `provenance` | inner loop | ~10 min |
| b2 focus-db | TestRetentionCatalog/Migration/Retention plus the reconciled TestAuditCatalog/Migration and access ITs at version 9 (`-run` selections) on a focus stack | inner loop | 4–8 min per run, 3–6 runs |
| b3 focus-drill | the drill phases on a focus source/target pair driven by a temporary local driver under `<EVID>` (not committed), then `run.sh migrations` once to prove the wiring | inner loop | 10–15 min per drill run, 3–6 runs; ~35 min for the suite |
| S-review | Delegate 1 (§12) | fix; rerun affected b1–b3 only | — |
| final | reconcile, then `all unit` (ci-controls auto-selected) | the **only** broad run; every required ID PASS once | 65–80 min |

**No manual or operator-feedback stage is required:** no UI, no live provider; the incident exercise uses the
deterministic local sink (Spec R6 permits this when the Plan says so). Total local machine time without failures: about
3–5 h. Unchanged passing batches are not repeated for a SHA-only or management-only delta, except the scan-input rule.

### 11.3 Version-9 reconciliation of accepted checks (guarantees preserved; G011 §10.2 pattern)

| Location | Change |
|---|---|
| `run.sh` `static_migration_set` | set + 00009; add `MIGRATION_00008_SHA` |
| root migrate block | 00009 lines; `current_version` 0/9/9 |
| B-MIGRATE-CLI/CATALOG | reapply and rollback act on 00009; `app` relation count stays 38 |
| B-PERSIST, B-DBRENAME, B-TOGGLE-OFF, B-IDENTITY-CATALOG, B-OBJ-IT, B-AUDIT-IT texts | version 9 |
| `migrations_test.go`, `runner_test.go` | set through 00009; 00008 byte-identical pin |
| `migrate/integration_test.go` | latest-version outputs; `reset()` also drops the version-9 objects |
| `migrate/audit_integration_test.go` | `subsetFS(t, 8)` so the G011 catalog and migration evidence keeps its version-8 meaning (CT-AUD-3/5 "retention nothing") |
| `tenancy/integration_test.go` | step-down chain 9→8→7→… (00009 down needs no purge) |
| `access/integration_test.go` | latest version 9: B-POLICY/B-RLS/B-OBJ/B-AUDIT ITs and G010's attack suite run with the retention policies present (inherited regression) |
| `run.sh` IT env file | adds the retention value for TestRetention (mode 600, removed after use) |

The implementer inventories every pin at binding with `grep -nE '00008|version 8|current_version.t8|_retention'` over
`infra/validation`, `infra/ci` and the core tests, records each change in the Result, and treats any hit outside the
allow-list as a conditional amendment for Codex review (for example a G010 attack-suite pin), never a silent edit.

### 11.4 Drill resources and bounds

- Source: the migrations project (`g001-mig-<run>`), reset with `down -v` after every existing check, then fresh volumes
  and migration to 9. Target: `g013-rec-<run>`, the same copy with its own generated env file and `PORTS_REC=(3700 8780
  8700 6132 7079 9700 9701)` (picked up by `runner-preflight.sh`), so no new per-run host path is bind-mounted.
- Old-authentication handoff: a run-owned named volume (`registry_row volume` at creation), removed at the end.
- Private material: `$WORK/recovery/private` (0700), shared by both drills, removed by trap on any failure and at the
  end of the maintenance drill, and asserted gone (MD-C); never in `--report-dir`.
- Every long step has a `timeout`; the whole drill is bounded at 30 min and the maintenance drill at 15 min; owned
  containers, networks and volumes are cleaned by the existing ownership registry; peak disk well under 1 GiB.

### 11.5 Harness wiring and count pins

New IDs, all in `migrations`: `B-RETENTION-CATALOG`, `B-RETENTION-IT`, `B-RECOVERY-DRILL`, `B-MAINTENANCE-DRILL`.

| Pin | Now (A-G012-2) | After |
|---|---|---|
| `required-checks.tsv` rows / migrations | 110 / 27 | 114 / 31 |
| profiles fast / integration / full | 35 / 74 / 103 | 35 / 74 / 107 |
| VA7 and the remainder-row check (validate-controls 1112, 1115, 1948) | 73 | 77 |
| VU1 full split (1206) | 35/73/104 | 35/77/108 |
| VU9 integration split, affected counts | unchanged | unchanged |
| OPERATIONS.md 93 and 125 | 103; 35/73/104 | 107; 35/77/108 |

CI meta-tests run because `infra/` changed (current selection), not as an extra campaign.

### 11.6 Performance observations (no benchmark)

Run in TestRetention as the actual retention login, on the pinned synthetic distribution of §10.1 after `ANALYZE`, with
the exact §6.3/§6.5 statements (r1). Each line records the statement shape, the fixture counts, the plan nodes, the
index or scan actually used, rows read versus deleted, buffers and timing:

- E-RET-1: EXPLAIN (ANALYZE, BUFFERS) of the Organization-scope audit batch and its count; whether
  `events_tenant_time_idx` is used is **unknown until this observation** and is not inferred for production volumes;
- E-RET-2: the global-scope batch and count (a sequential scan is expected because the partial index excludes these
  rows);
- E-RET-3: the tombstone candidate query and the confirmed-ID DELETE, plus the lock inventory during each DELETE
  (row-exclusive table locks and row locks on deleted rows only, no other `app` row lock).

An unexpected scan, helper work or lock prints `E-RET-TRIGGER` and returns to review (PG-005 stop rule). The drill
records durations and sizes. No index is added and no benchmark is run; PG-005's owner answer is pending.

### 11.7 CI selection (actual paths) and hosted gates

`core/migrations/`, `core/internal/{migrate,tenancy}/` and `infra/` select **full**; `infra/*` also selects
**ci-controls**. Docs alone would be fast. No `profile-paths.tsv` change; no affected-only hosted run is promised. The
four new IDs run only in the gate partition (`migrations`); the new scripts are linted by B-SHELL in the fast job.
Hosted: after Codex's pre-integration review, reconcile, push, `glab` MR to `staging` with source-branch deletion;
MR fast runs automatically; the optional manual gate is not requested; the post-merge staging full gate is required
(success or owner-canceled satisfies it, canceled recorded as canceled; any failed job fails; pending does not qualify).
Expected staging gate growth: about 10–15 min.

### 11.8 Provenance and the proposed B-SCOPE allow-list

Git-first. Per batch: candidate/target commits, commands, outcomes (failed, interrupted and skipped explicit), image
IDs and digests (core check, PostgreSQL, MinIO server and init), tool versions, CI links. Durable proof lives in the
Result and CI artifacts; raw output is temporary and purged at closeout. The Result is a scanned file and is checked with
the README scan line before each commit.

Proposed `management/v0/results/G013-scope-allowlist.txt` (imported by Codex at binding):

```text
# Create
^core/migrations/00009_retention_maintenance\.sql$
^core/internal/migrate/retention_integration_test\.go$
^core/internal/tenancy/access/(retention|recovery)_integration_test\.go$
^infra/recovery/(recovery|lib)\.sh$
^infra/recovery/sql/[a-z0-9-]+\.sql$
^infra/recovery/testdata/[A-Za-z0-9._/-]+$
^docs/foundation/RECOVERY\.md$
^management/v0/results/G013-(scope-allowlist\.txt|recovery-retention-incident-evidence-result\.md)$
# Modify: version-9 reconciliation
^core/migrations/migrations_test\.go$
^core/internal/migrate/(integration|audit_integration|runner)_test\.go$
^core/internal/tenancy/integration_test\.go$
^core/internal/tenancy/access/integration_test\.go$
# Modify: harness, catalogs and counts
^infra/validation/(run\.sh|required-checks\.tsv|render-customizations\.tsv)$
^infra/ci/validate-controls\.sh$
# Modify: docs
^docs/SECURITY\.md$
^docs/foundation/(AUDIT|OBJECT-STORAGE|DATABASE-ISOLATION|MIGRATIONS|OPERATIONS|FOLLOW-UP)\.md$
```

Conditional entries (only with a Codex-reviewed amendment): `infra/validation/lib.sh` (if a shared helper is needed),
reconciliation hits outside this list, and `core/internal/tenancy/access/attack_*_test.go` if a version-9 pin there moves.

## 12. Delegation and sequence

- **Implementer:** one owner of all files (task G013, role impl; `g013-agent-<id>-impl` or the native-name fallback,
  recorded with path, branch and HEAD). Default effort high. No file split: the migration, its tests, the operator
  command and the drill share the retention and storage boundary (Goal: no overlapping split).
- **Delegate 1 (required), read-only security review** after b2/b3: the migration (policy expressions, settings parser,
  role-bound refuse exception, grants, down), the malicious-operator residual, the capture fence and its exact release
  (r1), invalidation completeness, replay trust and attribution with the revocation-gap handling (r1), target identity
  refusals and the receipt-bound continuation (r1), value handling in scripts (no argv, env files 0600, no output of
  values), private-material lifecycle, sink sanitization, and the reconciliation's non-weakening. Findings are fixed by
  the implementer.
- **Delegate 2 (optional), read-only recovery analysis** after the first drill run: pg_dump/pg_restore 17 behavior
  against §3–§4, the consistency-window detection, failure-path completeness and AC mapping gaps. Both share B7 and
  record task, role and native worktree; decisions stay with Codex and the owner.
- The coordinator owns the shared harness lock and stack scheduling.

**Sequence:** (1) binding after G012 final ACCEPT: reconciliation (A-G012-n), migration number check, allow-list and
initial Result import; (2) b0 tools; (3) 00009 + TestRetentionCatalog/Migration + reconciliation; (4) TestRetention;
(5) `recovery.sh` backup/verify/restore + fingerprints; (6) invalidation, ledger, replay, reopen; (7) retention and purge
commands, expiry, quarantine, status; (8) Go drill phases; (9) harness wiring, pins, `APPLICABLE`, register reason;
(10) docs; (11) b1–b3 loops → S-review → fixes and affected reruns → reconcile → final → Result; (12) safekeeping push
after each round (no MR) → Codex review → MR.

## 13. Budget, forecast and scope

**B7**, shared attempts, checkpoint after attempt 4; planning spends none. **Forecast: 2 likely, 3 plausible.**

| Attempt | Expected content |
|---|---|
| 1 | b0 through the final run, review-ready |
| 2 | one bounded Codex correction (likely for a security/data Goal of this size) |
| 3 | hosted DinD drift in the drill (networks, image builds, timing) or a late G012 reconciliation |
| 4+ | not expected; the checkpoint after 4 reassesses convergence and scope |

**Context and scope:** one bounded cycle; estimated diff 5–6k lines (migration ~400, Go tests ~1.8k, scripts and SQL
~1.6k, harness ~0.5k, docs ~0.7k). It is the largest EPIC-004 Goal but needs no new platform. **No split is
recommended.** If the checkpoint shows risk, the split is G013a (00009, retention tests, retention and purge commands,
B-RETENTION-*, the maintenance drill on a seeded source) and G013b (backup, restore, invalidation, ledger, replay,
incident, B-RECOVERY-DRILL), with no weakened check.

## 14. Escalation items for explicit review; risks

| # | Item | Class |
|---|---|---|
| E1 | Migration 00009: additive retention policies, grants and function, and a replaced `events_refuse` body with a role-bound eligible-row DELETE exception; reversible; brief ACCESS EXCLUSIVE locks on two tables (quiescent window). Anticipated by the AUDIT.md G013 contract; not destructive to data | security-model change to accepted G011/G009 catalogs; migration |
| E2 | `<db>_retention` gains privileges (no new role): audit USAGE, minimal column SELECT and DELETE limited by policy to declared-eligible audit rows and `deleted` tombstones, EXECUTE on the eligibility function; it can read minimal columns of eligible global audit rows. §6.1 states the malicious-operator residual. No BYPASSRLS, membership, INSERT, UPDATE, TRUNCATE or DDL; no value reaches applications (B-COMPOSE) | new privileges |
| E3 | Privileged operator boundary: the cluster superuser for capture, restore, invalidation, replay and ledger sync; in the local fixture the MinIO administrator for the object fence and the capture inside it (r1), and the existing scoped object values for restore, replay and purge checks (Q2, accepted). Production equivalents are prerequisites. No RLS disable, `NO FORCE`, trigger disable or BYPASSRLS grant | privileged maintenance |
| E3a | **Capture fence (r1; r2; r3):** for the capture only, the source database's CONNECT grants to the non-superuser login roles are revoked and their backends terminated, a superuser holder takes SHARE locks on every captured table as a bounded safeguard, the application's object service user is disabled, and the source MinIO is restarted to drain accepted operations. Administrator exclusivity across both stores, PostgreSQL catalog activity included, is a recorded trusted operational window (`--admin-window`); capture refuses without it, and the guarantee is not claimed against a privileged administrator violating it; an optional trace is diagnostics only. The exact prior ACL and user status are recorded before any change and restored and verified on release; an interruption leaves the source fenced (fail-closed) until `unfence`, which converges to the recorded prior state and never enables or grants beyond it | temporary source change (exclusion), reverted |
| E4 | Restore-time invalidation by expiry with zero audit events (receipt evidence); replay recorded by existing triggers as `unknown` actor, and object deletions as the recorded original deleter (the G009 guard requires an actor; no system identity exists), with the actual restore operator stated in the receipt; a replayed disablement's revocation gap is read before COMMIT, never rolled back and receipted truthfully, with one bounded `revocation.unaudited` attempt (r1, Q7) | audit attribution |
| E5 | Restore-input ledger: operator-held hash-chained file, sync from authoritative state plus operator appends, completeness by operator attestation, reopen blocks, purge precondition, never pruned in G013; fail-safe over-restriction for later reversals | anti-resurrection design and trust boundary |
| E6 | Destructive assertions only on harness-created projects and sets; the tool has no reset; target refusals before mutation; source unchanged except RD-P's planned post-checkpoint fixture changes and purge and the reverted capture fence; the RD-Y fault trigger exists only on the disposable target and is dropped and verified gone | destructive-action safety |
| E7 | +4 required IDs, version-9 reconciliation of accepted pins (§11.3), `PORTS_REC`, B-SHELL list, `APPLICABLE`, register reason, count pins | validation-scope change |
| E8 | **PG-005 (recorded by Codex; owner answer pending, no risk approval):** r1 characterization: the Organization batch has explicit scope and time predicates and *may* use the partial `events_tenant_time_idx`, but index selection is unknown until E-RET-1 (RLS barrier, non-leakproof eligibility filter); the global batch has no index path and scans per batch, so total cost grows with table size times batches; held rows are read and then filtered; the tombstone query may use the `(organization_id, group_id)` index plus a sort (unknown until E-RET-3); delete WAL, autovacuum and row locks add load; backup fingerprints scan and sort twice inside the fenced window, which is closed to writers for its whole length (r2: the SHARE locks block rather than refuse later writers, and the provider restart adds its start-up time); single-transaction restore length and WAL; sync scans of `stored_objects` and `users`. Approach, pending the owner: **bounded measurements, no speculative index** (a partial index would add write cost to every global audit insert or object transition), E-RET-1..3 on the pinned distribution and drill durations; no benchmark. CI cost: about +10–15 min on the staging gate, +12–15 min on the local final run | performance policy |
| E9 | **None:** public HTTP/MCP/BFF/admin contract, new vendor, image or runtime service, production data or provider access, scheduler or daemon, new role or identity, Redis persistence, product schema, audit catalog change, architecture document change | explicit none |
| E10 | Canonical documents: `docs/SECURITY.md` edited under B-PROTECTED with `--expect-hash`; new RECOVERY.md | governance documents |
| E11 | The drill fixes the fake WorkOS issuer across processes so restored principals stay valid (test-only; production keeps one WorkOS namespace) | test fixture |

**Risks:**
- R-a: pg_restore TOC filtering or `mc` metadata flags behave differently from §3–§4. Mitigation: b0 confirms before
  wiring; a mismatch returns to Codex; the fallback is the equivalent documented mechanism with the same guarantee.
- R-b: hosted DinD differences (container networking, image builds, ownership capture) in the drill. Mitigation: the
  drill uses only patterns already used by `migrations` and `ops`; failures are diagnosed with bounded detail in the
  published summary (the G012 attempt-3 lesson), never retried blindly.
- R-c: the version-9 reconciliation exposes an interaction in inherited ITs or G010's attack suite. Mitigation: b2 runs
  them early; a pin moves only with its guarantee preserved and a conditional amendment.
- R-d: generated Plan, Result or docs text trips gitleaks `generic-api-key`. Mitigation: neutral labels on digest lines,
  backticked variable names, no slash-joined word lists after key-like words; README scan before each commit.
- R-e: the braces exception lapses on 2026-10-29 (A-G012-4). Mitigation: report as an external blocker; no G013 edit.
- R-f (r1; r2; r3): the fence primitives (CONNECT revoke with backend termination, the SHARE-lock holder, `mc admin
  user disable`, the graceful MinIO restart and, if kept, the optional trace's record format) behave differently from
  §3.2 on the pinned images, or an interruption leaves a fence held. Mitigation: b0 confirms them before wiring; the
  record is written before any change and release is verified against it; `unfence` and `status`
  (`backup.fence: held`) cover interruptions (RD-K); a mismatch returns to Codex, never a weaker drain or fence. An
  unreadable optional trace is recorded as unavailable diagnostics, not as evidence either way.

## 15. Open questions and dispositions

Dispositions from the r0 review, applied in r1:

- Q1 (PG-005): Codex recorded PG-005 and asked the owner whether to proceed with bounded measurements and no
  speculative index; **the answer is pending and no performance risk is accepted.** r1 is consistent with that
  approach (§6.3, §11.6, E8). Execution waits for the owner's answer where Codex requires it.
- Q2: reuse of the existing scoped synthetic object values: **accepted**; production values stay prerequisites. r1 adds
  only the fixture's existing MinIO administrator for the capture fence (§3.2, E3).
- Q3: operator receipts, no new audit action: **accepted**, together with the r1 gap handling (§5.2).
- Q4: four new IDs: **accepted**.
- Q5: Codex records PG-005: **done**.
- Q6: finite restriction kinds, later authority-reducing flows joining the handoff: **accepted**.

Dispositions from the r1 re-review, recorded in r2:

- Q7: **accepted** — one independent post-COMMIT attempt, bounded to 2 s, of the existing `revocation.unaudited` in its
  existing G6 shape, by the privileged recovery tooling; the disablement is kept, the receipt stays truthful
  (`recorded`, `failed` or `not_attempted`), earlier gaps survive retries (§5.2); no new action, runtime privilege or
  change to normal revocation.
- Restore continuation, gap receipts, the local MinIO administrator for capture (fixture only) and the explicit
  query/index assumptions: **resolved**, not reopened.
- PG-005 (Q1): the owner's answer is **still pending**; no performance risk is accepted.

Dispositions from the r2 re-review, recorded in r3:

- Q8: **accepted** — `COMPLETE` means capture completeness only; release is separate, recorded as verified, failed or
  deferred; a held fence stays unhealthy even for a complete set; a failed release stays a non-zero exit (§3.2).
- The concrete drain, the fence record before the first change and convergent interrupted release: **resolved**.
- Administrator exclusivity: the recorded trusted window across both stores (§3.2), applied in r3; the trace is
  optional diagnostics and the lock holder a bounded safeguard.
- PG-005 (Q1): the owner's answer is **still pending**; no performance risk is accepted.

Open: none.

## 16. Closeout and handoffs

- Bounded cleanup: one scoped preflight, one removal pass, a compact receipt; drill projects, volumes and private
  material recorded at creation; raw sets, ledgers, attestations and env files removed; uncertain items held; native
  refusals reported once.
- G013 closes EPIC-004: the PROGRESS compaction disposition is Codex's at closeout.
- **G014** receives (definition by Codex, no G014 work here): restore invalidates application sessions, login
  transactions and logout references; future MCP sessions or connection bindings must join that invalidation and the
  replay kinds; the retention boundary and maintenance trust limits; receipts as operator evidence.
- **G021** verifies the reusable recovery, retention and incident controls; it stays closed. **G056** later supplies
  assembled-product rights and real retention decisions; G013 neither waits for it nor implements it. **OS001** exports
  only accepted G021 generic mechanisms (RECOVERY.md stays neutral).

## Revision history

- r0, 2026-10-15 UTC: initial provisional Plan on staging `5e1e987` (G012 merged by MR !28; G012 final ACCEPT and
  staging disposition pending). Read-only inspection only; no subagent, probe, install, build, scan, test, Docker or
  harness command was run.
- r1, 2026-10-15 UTC: bounded revision for the r0 review (not included) CORRECTION LOOP,
  rebased onto staging `728165e` (management-only deltas `9f7464c`, which also records PG-005 in PERFORMANCE.md, and
  `728165e`, G012 final ACCEPT). Four corrections (change map at top):
  - an actual writer exclusion and drain for capture (§3.2);
  - the receipt-bound staged continuation with a same-target provider retry (§4.1, §4.2, RD-F);
  - truthful revocation-gap handling for replayed disablement (§5.2, §5.4, RD-Y, RD-G V9);
  - explicit retention query predicates with index use labelled unknown until E-RET-1..3 on a pinned distribution
    (§6.3, §6.5, §10.1, §11.6, E8).

  The review dispositions are applied (Q2, Q3, Q4, Q6 accepted; Q5 done; Q1 owner answer pending, no risk approval),
  Q7 is added, and A-G012-1..4 are bound to the accepted final G012. r0 is archived byte-identically (`4822a143…`).
  Read-only inspection only; no subagent, probe, install, build, scan, test, Docker or harness command was run. The ten
  ACs and B7 with the checkpoint after attempt 4 are unchanged.
- r2, 2026-10-15 UTC: small capture-lifecycle revision for the r1 re-review (not included)
  CORRECTION LOOP, rebased onto staging `b59fcda` (management-only delta with a PG-005 clarification). Three
  corrections (change map at top): concrete drains of already-accepted work in both stores (SHARE-lock holder, provider
  restart) with whole-window proofs and administrator exclusivity across PostgreSQL and MinIO (§3.2, RD-B); the fence
  record persisted before the first change with exact convergent release and interruption recovery (§3.1, §3.2, RD-K);
  one COMPLETE meaning, capture completeness only, with release recorded and reported separately (§3.2, §3.3, §7.1,
  RD-B, RD-O). Accepted dispositions recorded (Q7 shape, earlier gaps survive retries); Q8 added. r1 archived
  byte-identically (`7b80791b…`), r0 archive untouched (`4822a143…`). Read-only inspection only; no subagent, probe,
  install, build, scan, test, Docker or harness command was run. The ten ACs, the four check IDs and B7 with the
  checkpoint after attempt 4 are unchanged.
- r3, 2026-10-15 UTC: one trust-boundary clarification for the r2 re-review (not included)
  CORRECTION LOOP, rebased onto staging `df3d1b0` (management-only delta). Administrator exclusivity across both stores,
  PostgreSQL catalog activity included, is restated as a recorded trusted operational window (`--admin-window`), with
  refusal when it is not recorded and the guarantee stated within that boundary; the trace is optional diagnostics, not
  proof; the lock holder is a bounded safeguard; the manifest, RD-B (the trace-as-proof case replaced by the ordinary
  precondition refusal), AC1, b0, E3a, R-f and §15 aligned; Q8 accepted. The drain, the fence record before change,
  interrupted release and every other section are unchanged. r2 archived byte-identically (`d659406a…`); r1 and r0
  archives untouched. Read-only inspection only; no subagent, probe, install, build, scan, test, Docker or harness
  command was run. The ten ACs, the four check IDs and B7 with the checkpoint after attempt 4 are unchanged.
