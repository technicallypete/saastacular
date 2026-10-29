> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G020 — Generic labels and visibility — Result

Status: implementation released but sequenced after RC7; coordinator should preserve any started work and pause at a safe boundary. S0/source binding is required before source edits.

Goal: [G020](../goals/G020-generic-labels-visibility.md)
Spec: [G020 Work Spec](../specs/G020-generic-labels-visibility-spec.md)
Plan: [accepted r2](../plans/G020-generic-labels-visibility-plan.md)
Plan review: Codex r2 ACCEPT (not included)
Execution release: Codex confirmation (not included)
Input binding: 2026-10-21 snapshot (not included)
Scope: 28-pattern allow-list (not included)

## Execution binding

- Integration baseline: `staging` `8edf047716337344182dd94d85a94f64441f6d9e`.
- Budget: B5, checkpoint after attempt 3; implementation attempts spent: 0.
- Scope and exact current input hashes are in the execution release and binding snapshot. The snapshot records 49 changed hashes out of 115 versus the planning inventory; this is a change inventory, not semantic approval.
- G063 final ACCEPT remains a gate for G020 staging integration/final acceptance. Isolated implementation is released now.

## S0 status

Pending coordinator handoff and completion before application-source edits:

- Classify design-relevant input changes against current accepted source and policies.
- Verify the next migration slot and re-pin affected migration, catalog, route, audit, browser and harness selectors.
- Confirm DB-pool/build-resource availability and preserve lock/sequence rules.

## S0 record — 2026-10-21 (writer; read-only inspection, no source edit)

**Outcome: S0 clear, no material architecture/security/contract conflict.** One minor scope reconciliation for Codex
(Q-S0-1 below); implementation **paused** for the RC7 qualification sequence by owner re-sequencing. Implementation
attempts spent: **0** (S0 is not an attempt).

Writer: Opus 5.5 (high requested; runtime effort not observable), isolated worktree
`.claude/worktrees/agent-<id>`, branch `g020-impl` created at release commit
`65f397e0eeb5b2165fb46075702004f9eefb59d0` (HEAD and branch asserted before any edit).

**Baseline proof.** `git diff --name-only 8edf0477 65f397e0` = exactly the five management files (PROGRESS, this
Result, the binding JSON, the release record, the allow-list). `git rev-parse <rev>:<tree>` is equal at `8edf0477` and
`65f397e0` for `core` `e7317e11`, `web` `8ce067e4`, `intelligence` `c5d9ecc3`, `infra` `3f403c05`, `docs` `0de6bc2f`,
`contracts` `1c19fb7f`, `.gitlab-ci.yml` `81370def`. **Later staging delta:** `b0722a9a` and the current tip
`76e796cd` are management-only (same seven trees equal to `8edf0477`; changed paths: PROGRESS, ROADMAP, G020/G063 Goals,
G062 Result, this Result, the release record) — not rebased; source baseline stays `8edf0477`; implementation is paused
for the RC7 qualification sequence (Goal amendment "RC7 before G020", 2026-10-21).

### S0.1 Changed-input classification (49 of 115)

Every one of the 49 Plan-r2 hashes resolves to its bytes at the planning base `c1e724c6`; the 66 unchanged inputs include
every load-bearing access/tenancy source the Plan cites (`policy.go`, `facade.go`, `objects_policy.go`, `authority.go`,
`invitations.go`, `log.go`, `audit.go`, `errors.go`, `matrix_test.go`, `policy_test.go`, `rls_expected_test.go`,
`attack_expected_test.go`, `facts.go`, `store.go`, `names.go`, `tenancy/boundary_test.go`), migrations `00006`–`00008`,
`00014`–`00017`, `catalog.go`/`tools.go`/`catalog_fixture_test.go`, the membership/onboarding/enrollment adapters, the
web pages/components/clients the Plan patterns on (`groups/[id]/page.tsx`, `organization/members/page.tsx`,
`group-authority/*`, `state-badge.tsx`), `membership.spec.ts`, `.gitlab-ci.yml` and `mcp-seed.sql`. The design source
is therefore unchanged; the changed inputs below only move pins.

**Unaffected (governance/process or other Goals; no G020 design, source, selector, privacy or acceptance effect):**
AGENTS.md, CLAUDE.md, ROADMAP, PROGRESS, G060 Result/review/allow-list (G060 accepted; its source is in the
unchanged set), the four G055 Goal/Spec/Plan/Result files, `GITLAB-BRANCH-WORKFLOW.md`, `CLAUDE-PERMISSIONS.md`,
`docs/DECISIONS.md` (G062 signup decision), `infra/ci/lib.sh` (profile counts only; recounted at the harness step),
`infra/identity/emulator-seed.yaml` (G062 signup client).

**Affected — reconciled (reuse current accepted contracts):**

| Input | Delta (since `c1e724c6`) | G020 disposition |
|---|---|---|
| G020 Goal / Spec | context-policy and scheduling amendments (G062, G063, RC7 sequence); Spec MCP clarification | no requirement change; MCP clarification is already Plan D-1 (five production tools; fixture evidence proof-only) |
| `docs/ARCHITECTURE.md`, `docs/SECURITY.md` | G055 tooling and G062 root-signup sections | no label text touched; G020's §53.8 status note and SECURITY "Labels — G020" note append as planned |
| `docs/foundation/{AUTHORIZATION,DATABASE-ISOLATION,TENANCY,AUDIT,MCP,WEB-SHELL,MIGRATIONS,IDENTITY}.md` | G062 provision writer (4th helper role, signup policies, version 18 = 79 relations) and G063 `display_name` (version 19, 79 relations, `mcp.connection_renamed`) | G020 adds version-20 notes after the version-19 ones; DATABASE-ISOLATION's helper inventory is now four helpers (policy reader, auth reader, claim writer, provision writer) — G020 adds none; IDENTITY `identity:browser` line moved L325 → L440 (now includes `--project=signup-light/dark`); MIGRATIONS still lacks row 16 (D-11 observation stands) |
| `contracts/mcp/control-plane.md` | G063 rename semantics; "Production serves exactly the five tools" L111 and C-6 unchanged | untouched (D-1) |
| `core/migrations/migrations_test.go`, `migrate/runner_test.go`, `migrate/integration_test.go` | set 00001–00019; `metadataThroughVersion18/19`; relations 79; `current_version 19` | +00020 everywhere; new `metadataThroughVersion20` (79 + 9 = **88** relations; + three owner-owned app functions; audit functions + `capture_label_definitions`, `capture_label_assignments`) |
| `core/internal/audit/audit.go`, `audit_test.go` | 40 trigger + 7 core = 47 actions; 79 reasons; 73 operations; 11 resource types; parity reads 00019 (actions), 00018 (reasons/operations), 00016 (resource types) | +5 trigger actions → 45 + 7 = **52**; reasons **81**; operations **80**; resource types **12**; parity source for all four → `00020_labels.sql` with the "00019 list plus exactly the label values" guard pattern |
| `core/internal/tenancy/access/mcp.go`, `mcp_test.go`, `mcp_integration_test.go`, `boundary_test.go`, `audit_integration_test.go`, `integration_test.go`, `recovery_integration_test.go` | G063 `RenameMCPConnection` (+1 `inContextTx` site, now 19), AT-INV 47 actions, `current_version 19` | label operations run through `guard` → `inContextTx` sites stay **19**; AT-INV 47 → **52**; `current_version` 19 → **20**; MCP fixture tools are the Plan's test-only path (no production catalog change) |
| `core/cmd/server/main.go` | G062 signup + G063 rename registration; `MAIN_GO_G016_SHA=73d8baa2…` (run.sh:180) | register the labels handler; re-pin the digest |
| `web/lib/session.ts` | G062 signup cookies; `readForm` unchanged | reuse `readForm` as planned |
| `web/app/account/page.tsx` | G062/G063 moved the staff links into `web/components/account/account-sections.tsx` (L75-91, `manages` gate) | **Q-S0-1** (below) |
| `web/playwright.config.ts` | + `signup-light/dark` | add `labels-light/dark` (D-10) |
| `infra/validation/run.sh`, `lib.sh`, `required-checks.tsv` | route pins L184-202 (+`CORE_SIGNUP_ROUTES`, MCP rename), `WEB_ROUTES` L742, migration-set/CLI/catalog/persist/toggle/identity literals at version 19 / 79, B-AUDIT-CATALOG/IT 47 from 00019, browser matrix L644-650 (+signup row), 129 rows | see S0.2 pins |
| `infra/ci/profile-paths.tsv` | G062/G063 foundation rows (`core/internal/signup/`, `web/app/signup/`, `web/components/account/`) | add `foundation dir core/internal/labels/` (Plan) and, by the same precedent, `web/app/organization/labels/` and `web/components/labels/` (`web/app/groups/` is already foundation) |
| `logrecord/events.json`, `logrecord_test.go`, `completion.schema.json` | G062 `signup.request`, G063 `mcp.request`/rename; access.decision operations 67, capabilities 19; `/internal/v1/organization/group-authority` **is** present in the completion route enums (L55, L122) | +`labels.request`; operations 67 → **74**; capabilities 19 → **22**; label routes added to both enums. D-11's completion-schema observation is resolved (no G020 action) |

Current accepted contracts reused: G063's name CHECK control-class precedent (`(name COLLATE "C") !~
'[\u0001-\u001f\u007f-\u009f\u2028\u2029]'`, 1..80 code points, `= btrim`) for `label_definitions_name_valid`
(Plan D-5 bound, same grammar); G063's precondition shape (`md5(prosrc)` of the bodies replaced) with the **00019**
`audit.events_attribute` body as the base G020 edits and the down migration restores; G062/G063 helper inventories
(policy reader 9, claim writer 7, auth reader 8, provision writer 3) asserted unchanged in the 00020 postcondition.

**Q-S0-1 (minor scope reconciliation, not blocking):** Plan §1.5 links `/organization/labels` from `/account`, but the
owner/admin link block now lives in `web/components/account/account-sections.tsx`, outside the 28-pattern allow-list.
Options: (a) Codex adds the exact pattern `^web/components/account/account-sections[.]tsx$` (one `<a>` beside
`account-group-authority-link`, same `manages` gate); or (b) drop the `/account` link and keep the `/organization/members`
`staffLinks` entry, which is in scope. The writer will not edit outside the allow-list; default is (b) unless Codex
picks (a).

**Q-S0-1 resolved (Codex `ed0dd821`, 2026-10-21):** option (a). The allow-list gains exactly
`^web/components/account/account-sections[.]tsx$` (29 patterns; header and existing lines intact, dated amendment
comment). This is an implementation-location adjustment for the already approved `/account` → `/organization/labels`
link, not new scope. The proposed `profile-paths.tsv` foundation rows are accepted under the in-scope file.

**Release update (coordinator relay of Codex handoff `1d93425b`, 2026-10-21):** the owner superseded the RC7-first
hold; implementation starts now, ahead of RC7 qualification, from the paused state (`g020-impl` at `535f8f77`, no
reset, baseline `8edf0477`, no rebase). G020 staging integration/MR merge and final acceptance remain after RC7
qualification **and** G063 final ACCEPT. Attempts spent: 0 until the first implementation round completes.

### S0.2 Migration slot and pins (from actual source at `8edf0477`)

- Head is `00019_mcp_connection_names.sql`; no `000NN_labels.sql` exists; no locally known `origin/*` ref carries any
  `0002x` migration; **`00020_labels.sql` is the next free slot** and matches `^core/migrations/000[0-9][0-9]_labels[.]sql$`.

| Pin | Plan r2 figure | Actual now | New expected |
|---|---|---|---|
| migration set / latest version | 17 → NN | 19 | **20** (`migrations_test.go:44-55`, `runner_test.go:128-135,311`, run.sh B-MIGRATE-SET L1365/1377, CLI L2523-2535, root-migrate L1872/1891/1910, B-PERSIST L4813/4819, B-TOGGLE-OFF L4998/5027, B-IDENTITY-CATALOG L5149) |
| `current_version` test pins | 17 | 19 (access/integration_test:236, recovery:92, mcp/env:253, migrate/integration:992/1063/1074/1085/1212) | **20**; down 20 → 19 |
| relations | 71 → 80 | 79 | **88** (migrate/integration, B-MIGRATE-CATALOG `roles 79` rows, B-IDENTITY-CATALOG "79 in all") |
| capabilities | 19 → 22 | 19 (`matrix_test.go:878`, `policy_test.go`, events.json) | **22** |
| oracle cells | 627 → actual | 627 | re-pinned from the new tables at implementation |
| access operations (`log.go` constants) | 45 → 52 | 45 (+`opMCPRename` lives in `mcp.go`) | **52**; events.json access.decision operations 67 → **74** |
| audit actions / trigger | 46 → 51 / 39 → 44 | 47 / 40 | **52 / 45** (AT-INV, B-AUDIT-CATALOG, B-AUDIT-IT, AUDIT.md) |
| reasons / operations / resource types | 70→72 / 65→72 / 11→12 | 79 / 73 / 11 | **81 / 80 / 12** |
| helper inventories | 9 / 7 | policy reader 9, claim writer 7, auth reader 8, provision writer 3 | unchanged |
| `inContextTx` sites | 18 | 19 | **19** (unchanged) |
| tenancy production files / statements | 19 → 20; 86/24 → re-pin | 19; 86 (24 writes) | **20**; statements re-pinned |
| core routes | +7 pairs / 6 paths | `CORE_*_ROUTES` L184-202 | new `CORE_LABELS_ROUTES` (7 pairs) |
| `MAIN_GO_G016_SHA` | re-pin | `73d8baa2…` (L180) | re-pinned after `main.go` edit |
| `WEB_ROUTES` | + pages | L742 | + `groups/[id]/labels/action/route.ts`, `groups/[id]/labels/page.tsx`, `organization/labels/action/route.ts`, `organization/labels/page.tsx` |
| `required-checks.tsv` | 127 + 3 | 129 rows (profiles fast 35, integration 75, full 122) | **132** rows; profile counts recounted |
| browser matrix / projects | +1 row / +2 | 6 identity rows (L644-650); IDENTITY.md L440 | + `labels-light labels-dark` row; +2 `--project=` |

### S0.3 Capacity and resources (recorded, not re-tested)

At most two concurrent private DB/browser runs repository-wide (third only with owner approval); one image build at a
time; a G063 merged-SHA full qualification holds `local-harness.lock` in another checkout. The writer starts no Compose
stack, database, browser or image build and does not touch the lock; it runs only isolated stack-free unit/static checks
(plain native docker, read-only source, disposable GOCACHE, `-mod=readonly`, `GOPROXY=off`, `--network none`). DB,
browser, ops and harness batches (Plan §5.3 b2/b4 and the required final coverage) are coordinator-run.

### S0.4 Writer-1 context checkpoint and transfer state (2026-10-21)

Writer 1 reached the 200k handoff threshold after S0 and the scope amendment (`ce0f82a7`), before any application
source edit; handed back at this safe boundary for a fresh S1 writer (no attempt spent, no restart of S0). Transfer
notes for the successor (all read-only findings at `8edf0477`):

- 00020 template sources: table/constraint shape `00015:134-185` (`app.invitations`); keys-immutable and bound
  trigger bodies `00015:436-499` (owner-owned invoker, `format(...%L)` role names, `FOR NO KEY UPDATE` then count,
  `integrity_constraint_violation`); policy macros `00017:140-150` (policies `TO rt, ow`); precondition/postcondition
  and down-refusal shape `00019:33-111,419-487`; current catalog lists: actions `00019:118-136`, reasons/operations
  `00018:190-250`, resource types `00016:299-302`; `audit.events_attribute` base body `00019:140-380` (G020 edits it by
  script: + the five `label.*` actions in `trig`, a `label.*` shape branch, `label` in the `access.denied` resource
  list; down restores the 00019 bytes and the precondition pins `md5(prosrc)` of the 00019 body).
- Escapes: the file tools rewrite backslash-u sequences in content into raw characters (observed again in this Result,
  repaired by script); generate SQL/Go vectors containing them by script and verify bytes with `grep`/`od`.
- Scripts in the scratch dir `<home>/.claude/jobs/580cdb2f/tmp/g020w/`: `allow-check.sh` (29-pattern check;
  outputs `allow-check-*.txt`), `tree-compare.sh`, `staging-delta.sh`, `s0-diff.py`, `fix-escapes.py`.

## Work and validation ledger

No application source changes, implementation tests, builds, database/browser checks, CI runs, or cleanup actions had been performed before round 1. Record each implementation round, exact command, exit status/counts, failures and corrections below without deleting earlier evidence.

### Round 1 — S1 (schema, policy, registry, RLS oracles), writer 2, 2026-10-21

Writer 2 (Opus 5.5; successor to writer 1 at its S0.4 transfer state), branch `g020-impl-w2` from `a8955e16` (HEAD
asserted before edit), pushed fast-forward to `origin/g020-impl`. Bounded delegation (CLAUDE.md delegation; one write
owner): three sub-writers in isolated worktrees on disjoint file sets from the same Plan r2 + S0.2 pins; writer 2
reviewed, cherry-picked and integrated (`61c78ebd` P, `6ea43f0f` R, `ab8eac31` M, then the integration fix below).

- **P** (capabilities/oracles/registry/store): `label.organization.manage` (Organization kind, owner/admin),
  `label.group.manage` (own Group owner/admin), `label.group.read` (any own Group role); orgAdmin/orgMgr/override never
  consulted for `label.group.*`; `readCapability`: Group label kinds read through `label.group.read` (staff without a
  Group role get `not_found`). Oracle cells 627 → **726**; seven `opLabel*` (45 → 52); closed registry `{group}` in
  `labels_targets.go` with a nil `labelHooks` hook point (no `fixture` in production files); `tenancy/labels.go`
  statements (tenancy statements 86/24 → **97/30**, production files 20, locking +2, writeTargets +2, `inContextTx` 19).
- **M** (migration/catalog): `00020_labels.sql` up/down; relations 79 → **88**; five triggers; catalog actions 47 → **52**
  (trigger 40 → 45), reasons **81**, operations **80**, resource types **12**; `events_attribute` md5 00019
  `ce481e26…` → 00020 `cff36882…`, down restores the 00019 bytes; helper inventories 9/7/8/3 asserted unchanged;
  version pins 19 → 20; new `TestLabelsCatalog` CT-LAB-1..4 and `TestLabelsMigration` MG-LAB-1..4
  (`labels_integration_test.go`).
- **R** (T3): `TestRLSIsolation` label fixture rows, RLS-SCOPE/RLS-ORG extensions, RLS-LABEL reads (16) / writes and the
  isolated FK cases (i)–(iv), converse 42501/23514, paired positive, RESTRICT cleanup; `TestIsolationAttacks` step G1
  (G1-01..21). Subject mapping onto the access fixtures: carol=ownerA, frank=adminA, bob=crossUser (A3 owner),
  dave=gOwnerA1 (A3 member), Dana=ownerA4, Grace=fixture-only graceA (A admin + A3 admin, inserted/removed in RLS-LABEL).
- **Integration fix (writer 2):** access `targetScope` uses `audit.ResourceLabel` (P's local constant removed).

Deviations (recorded for review; none changes the authority/privacy model): (1) **seven** policies, not "8" (Plan §3
step 6 miscount; §3 step 4 names seven; assignments have no UPDATE); (2) the assignments SELECT EXISTS qualifies the
outer columns (`label_assignments.label_id`/`.organization_id`) — the Plan's unqualified text would bind to `d` and be
always true; the S2 fixture disjunct must qualify likewise; (3) `label_definitions_visibility_valid` also closes the
value set; (4) the label shape branch requires `group_id` on `label.assigned/unassigned`; (5) `*_id_nonzero` CHECKs
added (invitations precedent); (6) capture trigger names `audit_capture_label_*`; (7) label SQL lives in
`tenancy/labels.go` (TestSQLBoundary merges it), constraint mapping in `tenancy/errors.go` (`classify` precedent); (8)
adapter `resolve` also takes `Facts`; `admits(owner)` added; (9) `ValidateLabelName` also refuses U+2028/U+2029 (matches
the SQL class); (10) label RLS read/write oracles in sibling tables `rlsExpectedLabelReads/Writes` (the ok/42501
vocabulary cannot express zero-row/23503/23514/23505); carol's Organization-owned INSERT with A3 selected is pinned ok
(§3 Organization branch has no selected-Group filter). Predicted deparse pins (target shape, assignments SELECT, name
CHECK, trigger/FK defs) and multi-check first-refusal codes await the first b2 run (correctable under standing authority).

b1 stack-free (`gotest.sh`: `golang:1.27.3-alpine`, `--network none`, source `:ro`, module cache `:ro`,
`-mod=readonly`, `GOPROXY=off`, disposable GOCACHE; `b1.sh s1`, outputs `b1-s1-*.txt` in the writer-2 scratch dir), all
exit 0: access selector 13/13 top-level pass; tenancy 9/9; audit 2/2; migrations 2/2; migrate 2/2; mcp 3/3; `go test
./...` 23 packages ok; `go vet ./...` and `go vet -tags integration ./...` clean; gofmt 0 files. Byte check: `git grep`
finds the escaped class in `00020_labels.sql` only; raw U+2028/2029 count 0 in every G020-changed file and this Result
(pre-existing raw characters in `mcp_test.go`, `invitations_test.go` and the G063 Result are untouched). Allow-list: 29
patterns, 28 paths, 0 outside (`allow-check-s1.txt`).

**Not run:** every DB/integration test (`TestMigrationLifecycle`, `TestLabelsCatalog`, `TestLabelsMigration`,
`TestRLSIsolation`, `TestIsolationAttacks`, `TestAccessPolicy`, `TestAudit`, `TestMCPDiscovery`), web, browser, harness.
Known pending: AT-INV loops over `audit.TriggerActions` and fails until S2 adds label producers; label fixture rows
commit `label.*` events into the shared access test DB (AT-INV observed set, S2); `run.sh`/`events.json` literals are
stale until S2 (log vectors) and S4 (harness pins). Implementation rounds completed: **1** (attempts spent: 1 of B5).


## Owner sequencing update — 2026-10-21

Latest order: G063 final ACCEPT → RC7 full candidate qualification → G020 implementation/integration → combined staging preview → RC8 full candidate qualification → G057 → G021. This supersedes the original immediate-start timing in the execution release while preserving its scope and budget. Do not begin/resume G020 implementation until RC7 qualification completes. Preserve any work already started. The preview follows G020 and precedes RC8.


## Reconciliation onto staging `9761b89d` — writer 2, 2026-10-22

Resumption relayed by the coordinator (owner direction via Codex; same B5 ledger, round 1 / attempt 1 continues, no
reset). Note for Codex: the G020 Goal at `9761b89d` still ends with the G064-before-G020 pause and the order
G063 → G064 → RC7 → G020; the writer acted on the relayed release and found no dated resumption record on staging.

- Read-only delta `65f397e0..9761b89d` (merge base `65f397e0`): application/test paths only
  `core/internal/tenancy/access/objects_integration_test.go` (G063), `infra/ci/validate.sh`,
  `infra/ci/validate-controls.sh`, `docs/PERFORMANCE.md`, `docs/foundation/OPERATIONS.md`; the rest is
  governance/management/skills. No migration, no allow-listed source path; only overlap with G020 is this Result
  (conflict resolved keeping the round-1 ledger and the Codex sequencing section verbatim). No S0.2 pin moves:
  `objects_integration_test.go` carries no version/relation/action/capability literal; `infra/ci/*` is outside the
  allow-list and Plan r2 (G064 owns its count pins). No Plan assumption invalidated.
- Rebased `g020-impl-w2` onto `9761b89d` (committer `<owner-email>`); branch/HEAD asserted; staging is an
  ancestor. Post-rebase stack-free: `go vet -tags integration ./...` exit 0; `go test ./...` exit 0 (23 packages ok).
- G064 (MR !80) will change `infra/validation/{run.sh,lib.sh,required-checks.tsv}`, validate-controls pins and three
  migrate integration tests: a second rebase is expected; S4 harness pins must be re-derived after it.

### S0.5 Writer-2 checkpoint and transfer state (2026-10-22)

Writer 2 passed the 150k checkpoint at this safe boundary (S1 complete and rebased; S2 not started). For the S2 writer:
Plan r2 §4.1 (guarded operations in new `access/labels.go`: `guard`/`guardEvaluated` per §1.3 lock order; Organization
create via `policyWriteQuery`; Organization-owned delete needs Organization-wide context), §1.4 test-only fixture adapter
(`labels_fixture_test.go`, integration tag, registers through the nil `labelHooks` in `labels_targets.go`; the widened
SELECT disjunct must qualify outer columns as `label_assignments.*`; teardown byte-compares the exported
`LabelsWantTargetShape`/`LabelsWantAssignmentsSelect`/`LabelsWantAssignmentColumns` from
`migrate/labels_integration_test.go`), T4 `TestLabels` LAB-1..17, T5 (a)–(g), T6 `TestLabelsOrdering` + OR-8, T8 AT-LAB
and the AT-INV observed set (label fixture rows in `TestRLSIsolation` commit `label.*` events), T9
`core/internal/labels/api` (copy `membership/api` helpers, `TestLabelsCopyIsExact`), T10 MCP fixture tools, log vectors
(`events.json` `labels.request`, operations 67 → 74, capabilities 19 → 22; completion-schema routes), `main.go`
registration. New exported facade decls go into `access/boundary_test.go` `pinnedExportedDecls/Methods`. Audit constants
already exist: `audit.Label{Created,Updated,Deleted,Assigned,Unassigned}`, `ReasonLabelRenamed`,
`ReasonLabelVisibilityChanged`, `OpLabel*`, `ResourceLabel`. Tenancy statements/codes are in `tenancy/labels.go` and
`tenancy/errors.go`. Effective delegation pattern used in S1: disjoint-file sub-writers in isolated worktrees branched
from the current HEAD, cherry-picked by the writer (brief: `.../tmp/g020w2/BRIEF-common.md`, update its base SHA).


### Round 2 — S2, S3, S4 (writer 3), 2026-10-22

Writer 3 (Opus 5.5; successor to writer 2 at its S0.5 checkpoint), branch `g020-impl-w3` from `origin/g020-impl`
`938fdfe9` (branch/HEAD asserted before any edit), pushed fast-forward to `origin/g020-impl` after each slice
(`2a23494a` S2, `1df0782d` S3 + docs, then the S4 head below). Same B5 ledger; this round runs under attempt 1's
ongoing budget (no reset; see the attempt line at the end). Bounded delegation (CLAUDE.md delegation; one write
owner): six disjoint-file sub-writers in isolated worktrees branched from `938fdfe9` (P from `1df0782d`), coding
against a fixed facade API contract stub (`.../tmp/g020w3/labels_facade_stub.go`, compile-only, never committed),
plus one read-only security review. Writer 3 reviewed, cherry-picked and integrated every commit.

| Commit | Slice | Content |
|---|---|---|
| `0c396d53` | S2 F | `access/labels.go`: the seven guarded operations (Plan §4.1) through `guard`/`guardEvaluated` (`inContextTx` sites stay 19); `TestLabelInputs`, `TestLabelDecisions`; boundary pins (+`labels.go`, 5 types, 7 methods) |
| `e08c1454` | S2 H | `core/internal/labels/api` (7 pairs / 6 paths, strict bodies, `labels.request`, `TestLabelsHTTP/RefusalMapping/BodyShapes/RoutesPinned/CopyIsExact` — 21 helper declarations byte-pinned to `membership/api`); `main.go` registration; logrecord `labels.request`, access.decision operations 67 → 74, capabilities 19 → 22, label routes in both completion enums |
| `33747c5e` | S2 fix | `TestProductionImporters` admits `core/internal/labels/api` (importable by `cmd/server` only); events.json lets the three assignment operations record `group.read` (the facade's second decision on the exact target Group; a denial there, e.g. a Group outside the selected context, records `group.read`); +1 accept vector |
| `434fc01a` | S2 I2 | `TestLabelsOrdering` O-LAB-1..4 and OR-8 (`rls_ordering_integration_test.go`), `TestAudit` AT-LAB + AT-INV (45 trigger / 52 actions), MCP T10: three test-only label fixture tools (`useLabelFixtures`), `TestToolCatalog` still 5 production / 0 capability tools, `TestMCPDiscovery` CAT-6 |
| `2a23494a` | S2 I1 | `labels_fixture_test.go` (test-only `fixture` kind via `labelHooks`, §1.4 r2 widening install/teardown with byte comparison against the captured bodies and the copied migrate pins, plus a source check that the copies equal `migrate/labels_integration_test.go`); `TestLabels` LAB-1..17; `TestRLSIsolation` RLS-LABEL-FIXTURE = T5 (a)–(g) |
| `59fc3195` | S3 W | `web/lib/labels-client.ts`, `components/labels/*`, `/organization/labels`, `/groups/[id]/labels`, Group-page Labels section, members `staffLinks` and `/account` (`account-sections.tsx`) links, `labels-light/dark` projects, `labels-client` (17) / `labels-forms` (13) unit tests, `labels.spec.ts` E-LAB-1..6 + S-LANDMARKS/S-NAV |
| `1df0782d` | S4 D | supersession notes (Plan §9) in ARCHITECTURE, SECURITY and eight foundation docs |
| `2fdbbc72` | S2 review | comment-only corrections from the security review (findings 1 and 3) |
| `266a0d32` | S4 P | harness pins: run.sh / lib.sh / required-checks.tsv / profile-paths.tsv |

**S4 harness pins (P, mechanical, derived from the actual tests/source at `1df0782d`):** migration set + `00020`,
version 19 → 20 and relations 79 → 88 across B-MIGRATE-SET/CLI/CATALOG (unique rows P, 0, 79, 88; rollback/inspect stay
79 at version 19), root migrate, B-PERSIST, B-TOGGLE-OFF, B-DBRENAME, B-RECOVERY-DRILL, B-IDENTITY-CATALOG (`LABEL_OBJECTS`,
88 in all) and the version-20 texts of the IT checks; B-AUDIT-CATALOG source `00020_labels.sql`, 52/45; B-AUDIT-IT 10
subtests, AUDIT-INV 52/52; `CORE_LABELS_ROUTES` (7 pairs), labels route-pattern and importer checks, both access-importer
lists, `MAIN_GO_IMPORTS`, `MAIN_GO_G016_SHA` → `561045f7…11bc30` (re-verified by writer 3: `sha256sum core/cmd/server/main.go`);
B-TENANCY-HTTP label probes; subtest counts TestRLSIsolation 10 → 21, TestRLSOrdering 7 → 8, TestMCPDiscovery 5 → 6, and
(S1 carry-over) TestIsolationAttacks 22 → 24 with ISO-ORACLE 142 → 163; new `labels_checks` B-LABELS-CATALOG (4 + 4),
B-LABELS-IT (17), B-LABELS-ORDERING (4); required-checks 129 → **132** rows (fast 35, integration 75, full 125, staging 5);
`WEB_ROUTES` +4; request-context importer list + `lib/labels-client.ts`; lib.sh browser row `labels-light labels-dark|E-LAB-1..6
S-LANDMARKS S-NAV S-ERRORS`; profile-paths foundation rows `core/internal/labels/`, `web/app/organization/labels/`,
`web/components/labels/`. None of these harness checks has been executed.

**Deviations (for review; none changes the authority/privacy model):** (1) `TestLabelsOrdering` lives in
`rls_ordering_integration_test.go`, not `labels_integration_test.go` (file ownership during parallel delegation; same name and
b2 selector). (2) Group-owned update/delete/assign re-read the owning Group `FOR SHARE` (the `invitationAdmin` pattern; no
upgrade when that Group is the selected one), not `FOR NO KEY UPDATE` as Plan §1.3/§4.1 word it; owner scope is immutable and
the definition row lock is unchanged (security review: no impact). (3) In update/delete the manage decision precedes the
`visibility_not_allowed`/`organization_context_required` checks, so the record carries the capability (only managers can read
an Organization-owned definition). (4) Invalid scope field is `invalid_scope` (Plan §4.1). (5) Assignment rows are re-checked
against the decided owner scope on top of RLS; an unreadable Group-owned target is a recorded `not_found`. (6) The widened
fixture SELECT keeps every production conjunct and widens only `target_type = 'group'` to `(target_type = 'group' OR
(target_type = 'fixture' AND EXISTS …))` with outer columns qualified (a whole-body `production OR fixture` would skip the
definition check). (7) The widening is installed per suite (TestLabels after LAB-12; inside RLS-LABEL-FIXTURE), not in `itSetup`;
T5 (e)'s 23514 is observed through a rolled-back superuser insert because RLS refuses the runtime insert first (42501). (8) The
fixture MCP tools use their own catalog seam so CAT-1..5 stay unchanged. (9) Web row byte budget 640 B (assignment row 560 B),
not 400 B: Go's JSON encoder escapes `<`, so a valid 80-code-point name can reach 482 B (worst-case row 631 B; arithmetic in the
code comment, G019 precedent); `listGroupLabels` is unpaged (≤ 200 rows). (10) Web forms send both name and visibility on a
Group-owned update (absolute set). (11) E-LAB-4's Group-role removal runs on a Group bob creates in the spec (seed intact for the
dark project); the override is reset in `finally`. (12) Disabled-variant file list in run.sh gains `core/internal/labels
web/lib/labels-client.ts web/app/organization/labels` (main.go now imports `labels/api`; not named in Plan §2). (13) The
stub's doc example "fixture" was dropped from `access/labels.go` (no non-test file names it).

**Security review (read-only, Plan §6, at `2a23494a`): no stop finding.** No path by which a Group role reveals an
Organization-owned label or its assignments, or Organization authority (orgAdmin, orgMgr/override) reads a Group-owned
definition; no production reachability of the fixture adapter; no name/target id in `labels.request`, access.decision or audit.
Findings and dispositions: (1) owning-Group lock wording → deviation (2) + comment fix `2fdbbc72`; (3) Group-owned paths resolve
the target before the manage decision, safe because resolve reads under the invoker's RLS → registration contract item 1 now
requires that (`2fdbbc72`); (4) `organization_context_required` is reachable only through the facade: the BFF resolves
Organization-wide (`http.go` resolve with the Organization id only), so the web status string is the safe, currently unused
direction and LAB-17 is facade-level evidence; (5) test-only fixture policy has no `ctx_org` conjunct (relies on
`rls_group_role`, Plan §1.4 wording) — no production effect. **For Codex (N1, disposition requested, no code change made):**
an Organization owner reads all Organization audit rows (`audit.read`), including `label.created/updated/deleted` of Group-owned
definitions (definition id, owning Group, actor, time, reason; never the name) and `access.denied` rows for `label.group.*`;
this is the accepted D-4 shape and the `object.*` precedent, but it is existence/timing metadata of Group content. Options:
accept and state it (TENANCY/AUDIT), or omit `resource_id` on Group-owned label events (a migration change, new scope).

**Stack-free checks (writer 3, integrated HEAD `266a0d32`; `b1.sh final`, outputs `b1-final-*.txt` in `.../tmp/g020w3/`;
`golang:1.27.3-alpine`, `--network none`, source `:ro`, module cache `:ro`, `-mod=readonly`, `GOPROXY=off`, disposable
GOCACHE), all exit 0:** access selector (+`TestLabelInputs|TestLabelDecisions`) 15/15; tenancy 9/9; audit 2/2; migrations 2/2;
migrate 2/2; labels/api 5/5; mcp 3/3; `go test ./...` 24 packages ok; `go vet ./...` and `go vet -tags integration ./...`
clean; gofmt 0 files. logrecord module (`infra/validation/tools`, same docker pattern): `go test ./internal/logrecord/` ok.
`bash -n run.sh lib.sh` ok. Allow-list: 29 patterns, 73 paths since `9761b89d`, 0 outside. Byte check: raw U+2028/2029 count 0
in every round-2 file (the one in `core/internal/mcp/mcp_test.go` predates G020: 1 at `9761b89d`, 1 now); no backslash-u
escape text added. Web unit tests: the official run needs `node_modules` (absent in worktrees; `server-only` not found), so
B-UNIT-WEB is coordinator-run; unofficial evidence on a disposable copy with an empty `server-only` stub: labels-client 17/17,
labels-forms 13/13, membership-client + group-authority-forms + membership-forms 35/35. tsc/lint/next build not run.

**Not run (coordinator batches; predicted outcomes to confirm, pins correctable under standing authority):**
b2 `./internal/migrate -run '^(TestMigrationLifecycle|TestLabelsCatalog|TestLabelsMigration|TestAuditCatalog|TestMembershipCatalog)$'`;
`./internal/tenancy/access -run '^(TestLabels|TestAccessPolicy|TestRLSIsolation|TestIsolationAttacks|TestAudit)$'`;
`./internal/tenancy/access -run '^(TestLabelsOrdering|TestRLSOrdering)$'` (OR-8 and O-LAB-1..4);
`./internal/mcp -run '^(TestMCPDiscovery|TestMCPTransport)$'`. Predicted SQLSTATEs (T5): 23503 `label_assignments_fixture_target_fkey` ×3,
`label_assignments_target_group_fkey`, `label_assignments_label_fkey`; 23514 `label_assignments_target_shape` ×2; 42501 ×2.
Predicted denial split to confirm: Group roles on Organization labels — `forbidden` for list/create/list-assignments (the
Organization is readable), `not_found` for update/delete/assign by definition id (unreadable row). b3 web unit/lint/typecheck;
b4 browser `--project=labels-light --project=labels-dark --grep 'E-LAB|S-LANDMARKS|S-NAV'` (plus E-ONB-8 Group-page
assertions); harness B-LABELS-*, B-MIGRATE-*, B-AUDIT-*, B-MCP-SOURCE, B-TENANCY-HTTP, B-WEB-BOUNDARY, B-IDENTITY-E2E,
B-SCOPE, B-REQUIRED-CHECKS.

**Staging delta:** after S4, `origin/staging` (`git diff --stat 9761b89d origin/staging`) carries the G064 merge (26 files: run.sh
805 and lib.sh 441 lines changed, required-checks.tsv, `infra/ci/validate*.sh`, three `migrate/*_integration_test.go`,
`tenancy/integration_test.go`, DATABASE-ISOLATION/IDENTITY/MIGRATIONS docs, gotest report tooling). Read-only merge-tree
of `origin/staging` with `266a0d32`: content conflicts in exactly `core/internal/tenancy/integration_test.go`,
`infra/validation/required-checks.tsv` and `infra/validation/run.sh`; the three docs and lib.sh auto-merge. This round did **not** rebase (a rebase of already-pushed history needs a
non-fast-forward push, outside writer 3's authority); the second rebase and the S4 harness-pin re-derivation remain due
before b2. Implementation rounds completed: **2** (S1; S2–S4) under attempt **1** of B5 (no reset).


### Round 3 — rebase onto G064 staging (writer 4), 2026-10-22

Writer 4 (Opus 5.5; rebase writer, single write owner), isolated worktree `agent-<id>`, local branch
`g020-impl-w4` from `origin/g020-impl` `35dd2b5f` (branch/HEAD asserted; local safety ref `g020-impl-pre-g064-35dd2b5f`,
not pushed). The 18 commits were rebased from `9761b89d` onto `origin/staging` `6ea4cc31` (G064 merged), then again onto
`10638c23` when staging advanced during the round (management-only: `PROGRESS.md` and
`G020-audit-metadata-codex-disposition-20261022.md`; replay clean). Author and committer of every commit: <owner>
<<owner-email>>. No source behaviour or assertion was changed beyond the merge below. Same B5 ledger, attempt 1 (no
reset).

**Conflict resolutions (exactly the three files the round-2 merge-tree predicted):**

| File (commit) | Resolution |
|---|---|
| `core/internal/tenancy/integration_test.go` (S1 `eca9725f`) | G064's TENANCY-SETUP starting-version check first (top version derived from the embedded set, now 20; start must be top or 5), then G020's version-20 branch (purge label rows and audit history, 20 → 19), then the existing version-19 ladder. G064's `defer c.Close` in the purge helpers kept. No assertion changed |
| `infra/validation/required-checks.tsv` (S4 `266a0d32`) | both rows kept: `B-IT-ISOLATION` then `B-LABELS-CATALOG/IT/ORDERING` (migrations); header terms `+ 1 G064 + 3 G020`; counts re-derived from the merged rows (below) |
| `infra/validation/run.sh` (S4 `266a0d32`) | (1) G064 hoisted the top-version relation pin into `MIG_RELATIONS`; it is now **88** (version 20), used by mig_step 3-apply/4-reapply/7-apply, B-MIGRATE-CATALOG and (through lib.sh) `db_pristine_verify`/T-1; the version-19 rollback/inspect steps keep the literal 79 and B-MIGRATE-CATALOG expects the unique rows P, 0, 79, `$MIG_RELATIONS`; G020's 00020 pins and texts otherwise as in round 2. (2) `it_isolation_record` stays immediately after `signup_checks` (its T-2r repeat); `labels_checks` follows, before `recovery_drill`. (3) `rls_it` keeps G064's `it_fresh` before `rls-identity-it` with G020's TestRLSIsolation 21; `mcp_it` keeps `it_fresh` before `mcp-facade-it` with TestMCPDiscovery 6 |

**G064 conventions adopted in G020's own harness code (not a conflict; required by G064's reporter):** each of the three
`labels_checks` invocations now starts on its own fresh template copy (`it_fresh … labels-catalog|labels-it|labels-ordering`
before `rls_go`), and the three names are declared in `GO_IT_migrations` (`labels-catalog|B-LABELS-CATALOG`,
`labels-it|B-LABELS-IT`, `labels-ordering|B-LABELS-ORDERING`); without the declaration validate.sh fails an undeclared
invocation. The labels_checks comment no longer says "on the signup_checks template". No G020 Go test needed a fixture
change: G064 changed only the tenancy `setup` helper (above) and three migrate PersistSeed tests; `TestLabelsCatalog` and
`TestLabelsMigration` already `reset` first, and TestLabels/TestLabelsOrdering/TestRLSIsolation T5 use the unchanged
access fixtures. lib.sh and the three G064-touched docs auto-merged with byte-identical G020 hunks (per-file patch
comparison old-base→old-head vs new-base→HEAD: only the resolutions above differ).

**Re-derived counts (from the merged file, `infra/ci/lib.sh` profiles + `final`):** required-checks **133** rows (130
staging + 3 G020), header per-Goal sum 133, no duplicate or malformed row; profiles fast 35, integration 75, **full 126**,
staging 5. Migration pins unchanged from round 2 (version 20, relations 88 at the top, 79 at version 19).
`MAIN_GO_G016_SHA` `561045f7…11bc30` = `sha256sum core/cmd/server/main.go`.

**Stack-free checks (rebased HEAD before this entry; `golang:1.27.3-alpine`, `--pull never --network none`, source and
module cache `:ro`, `-mod=readonly`, `GOPROXY=off`, disposable GOCACHE; scripts and outputs `.../tmp/g020rb/b1*.sh`,
`b1-r3-*.txt`), all exit 0:** access selector 15/15; tenancy 9/9; audit 2/2; migrations 2/2; migrate 2/2; labels/api
5/5; mcp 3/3; `go test ./...` 24 packages ok; `go vet ./...` and `go vet -tags integration ./...` clean; gofmt 0 files;
tools module: logrecord ok, gotest (G064 reporter) ok, `go vet ./...` clean. `bash -n` ok on run.sh, lib.sh,
validate.sh, validate-controls.sh and the three G064 driver scripts; shellcheck v0.11.0 B-SHELL pass (`-S warning`, and
`-S info --include=SC2251`) on run.sh + lib.sh rc 0. Static parity (`invcheck.sh`): every literal `go_it`/`rls_go`
name is declared (the undeclared-literal set is empty; rd-*, obj-sign-host, rls-provisioned-* are dynamic); each labels
`it_fresh` purpose equals its invocation; each B-LABELS-* and B-IT-ISOLATION ID has exactly one row. `git diff --check`
clean. Allow-list: 29 patterns, 73 paths since `10638c23`, 0 outside.

**N1 reconciliation (Codex disposition `G020-audit-metadata-codex-disposition-20261022.md`, ACCEPT):** the bounded label
audit envelope (action, outcome/reason, Organization and owning/target Group attribution, resource type `label`, definition
id; never a name, target type or target id) stays readable only through the existing Organization-owner `audit.read`; no
source change or rerun follows from the ruling, and no new access is implied. Recorded here as the requested Result
reconciliation; it is not G020 acceptance.

**Not run (coordinator batches; nothing above executed a database, Compose stack, browser or harness suite):** b2 DB batch
as listed in round 2 (now on fresh template copies: TestLabelsCatalog/TestLabelsMigration, TestLabels, TestLabelsOrdering,
TestRLSIsolation incl. T5, TestMCPDiscovery CAT-6, TestAudit, TestIsolationAttacks, the tenancy setup ladder from 20);
harness `migrations` suite for B-MIGRATE-CLI/CATALOG (88/79), B-IT-ISOLATION T-1 at 88, B-LABELS-*, B-REQUIRED-CHECKS
(133); B-UNIT-WEB, web lint/typecheck, browser E-LAB. Unverified until those run: the T-1 pristine catalog at 88 and the
template copy at version 20, and the tenancy setup's 20 → 5 ladder from a fresh top-version copy. Implementation rounds
completed: **3** (S1; S2–S4; rebase) under attempt **1** of B5 (no reset).


### Round 4 — merge of current staging (writer 5), 2026-10-23

Writer 5 (Opus 5.5; single write owner), isolated worktree `agent-<id>`, `g020-impl` reset to
`origin/g020-impl` `79069c76` (HEAD and clean tree asserted). Stack-free only: no Docker, Compose, harness, control,
build, shellcheck or browser run. Same B5 ledger, attempt 1 (no reset).

**Reconciliation:** `origin/staging` `66f00679` (160 commits after merge base `10638c23`: G064 corrections and
reuse stages, G065, G066, Next 16.4.1 / eslint-config-next 16.4.1, opentelemetry-api 1.45.1, braces exception removed,
docs/management) merged by merge commit `f289eb3a` (history kept; fast-forward push). Staging changed no `core/`,
`contracts/`, migration, `required-checks.tsv`, `profile-paths.tsv`, `infra/ci/lib.sh` or `web/` source file (only
`web/package.json` and its lockfile), so slot `00020`, version 20, relations 88/79, catalog 52/45/81/80/12, capability 22,
oracle and subtest pins and 133 required rows (fast 35, integration 75, full 126, staging 5) stand unchanged.

| File | Resolution |
|---|---|
| `infra/validation/run.sh` migrations batch (conflict) | `labels_checks` stays right after `it_isolation_record` and now precedes staging's G064 §C3 `go_it_runner_end` guard around the recovery drill (the three labels invocations run inside the B-MIG batch on the runner); staging's held-drill branch kept verbatim |
| `infra/validation/run.sh` `persist()` (conflict) | staging's G065 K3 body (failed restart records B-PERSIST FAIL and returns; `|| ok=0` on the reads) plus G020's `00020_labels` history pin and `current_version 20` |
| `infra/validation/run.sh` B-TENANCY-HTTP (semantic, `268a1f5c`) | G020's label-route POST probe was the only raw `docker run` left after G064 routed every creator through `docker_run_owned`; now `docker_run_owned wget-probe …` like its G018/G019/G058 siblings; paths, statuses and record text unchanged |
| `lib.sh`, `docs/ARCHITECTURE.md`, `docs/SECURITY.md` | auto-merged |

Verification (all exit 0 / empty): G020's added/removed lines in `lib.sh`, `run.sh`, ARCHITECTURE and SECURITY are
byte-identical before and after the merge (only the probe fix differs since); `git diff --stat 79069c76 f289eb3a` over
`core contracts infra/validation/tools required-checks.tsv profile-paths.tsv web/app web/lib web/components web/tests
web/playwright.config.ts .gitlab-ci.yml` is empty, so round 3's Go/tools stack-free evidence stays attributable (no
rerun); no shell function used by a G020 line was removed or redefined on staging except `docker_run_owned` (now used);
no raw `docker run|create` remains in run.sh; `GO_IT_migrations` still declares the three labels invocations; `bash -n`
run.sh, lib.sh, validate.sh, validate-controls.sh; `git diff --check`; `required-checks.tsv` 133 rows, 2 columns, no
duplicate ID; `profile-paths.tsv` 79 rows, 3 columns; `events.json`, `completion.schema.json` parse. Allow-list: 29
patterns, 73 paths vs `origin/staging`, 0 outside. Diff vs staging: 73 files (+11658/−195 at the merge, same as round 3).

**Remaining work:** none source-side. Every approved slice (S1–S4) is coded; only real-daemon validation can advance.

**Real-daemon selectors for the coordinator (dependency order; private runs per CLAUDE.md, at most two concurrent,
one build at a time; estimates unmeasured unless noted):**

1. Build (one build slot): core check image and web check image from this HEAD (staging's Next 16.4.1 lockfile;
   G020 web source unchanged since round 2).
2. b2 private PostgreSQL (one DB slot, `core/`, `-tags integration`, pinned MinIO for TestAudit/TestIsolationAttacks):
   `./internal/migrate -run '^(TestMigrationLifecycle|TestLabelsCatalog|TestLabelsMigration|TestAuditCatalog|TestMembershipCatalog)$'`;
   `./internal/tenancy/access -run '^(TestLabels|TestAccessPolicy|TestRLSIsolation|TestIsolationAttacks|TestAudit)$'`;
   `./internal/tenancy/access -run '^(TestLabelsOrdering|TestRLSOrdering)$'`;
   `./internal/mcp -run '^(TestMCPDiscovery|TestMCPTransport)$'`; `./internal/tenancy -run '^TestTenancyModel$'` (setup
   ladder from version 20; B-TENANCY-IT). First real run of the T5 SQLSTATE/deparse predictions (round 2), correctable under standing authority.
3. b3 web (no DB; parallel with 2): `unit-web` suite (B-UNIT-WEB: labels-client 17, labels-forms 13), `lint-web`
   (B-WEB-BOUNDARY, lint, typecheck).
4. Harness suites, `infra/validation/run.sh --base 66f00679 --allow-file management/v0/results/G020-scope-allowlist.txt`
   after `bash infra/ci/docker-desktop-mounts.sh`: `lint lint-base lint-core unit-core` (B-SHELL, B-MIGRATE-SET,
   B-MCP-SOURCE, B-UNIT-CORE); `migrations` (B-MIGRATE-CLI/CATALOG/IT, B-IT-ISOLATION T-1 at 88, B-LABELS-CATALOG/IT/
   ORDERING, B-RLS-*, B-AUDIT-*, B-MCP-IT, B-IDENTITY-CATALOG, B-RECOVERY-DRILL; the longest DB batch); `root`
   (B-TENANCY-HTTP label probes through `docker_run_owned`). Coverage reason: migration 00020 changes the schema every
   DB-bound check reads and run.sh/lib.sh pins move in these suites.
5. b4 browser (private identity Compose stack, one test slot, reserved loopback ports): `ops` (B-IDENTITY-E2E with
   `--project=labels-light --project=labels-dark`, E-LAB-1..6 S-LANDMARKS S-NAV S-ERRORS, E-ONB-8 Group page; B-PERSIST
   version 20); then `rename` (B-DBRENAME) and `off` (B-TOGGLE-OFF, disabled-variant list incl. labels paths).
6. `ci-controls` / `controls` (run.sh/lib.sh changed: G064 controls copy them) and final B-SCOPE, B-REQUIRED-CHECKS (133).
   Steps 4–6 together approximate the full profile; a broad run needs the supervisor's recorded disposition. Exact merged-SHA
   local qualification stays post-merge per CLAUDE.md.

**Not run (honest):** every item in the list above; no Go test was rerun this round (inputs unchanged, round-3 evidence
reused); web unit/lint/typecheck/build never run officially (Next 16.4.1 now unverified against the labels pages);
shellcheck B-SHELL on the merged run.sh not run. Implementation rounds completed: **4** (S1; S2–S4; rebase; staging merge)
under attempt **1** of B5 (no reset).

### Round 5 — first real-daemon run: diagnosis and correction (writer 6), 2026-10-23

Writer 6 (Fable 5.1; single write owner), isolated worktree `agent-<id>`, `g020-impl` at `c6a7351b` (HEAD
and clean tree asserted). Stack-free only (no Docker, Compose, harness, control, build, shellcheck or browser run); the
coordinator ran the real daemon. Same B5 ledger, attempt 1: this bounded correction is part of the attempt (no reset).

**Failing evidence** (coordinator run 2026-10-23T21:41–21:46Z, preserved read-only at
`~/.local/state/platform-evidence/g020-r4/`):

| Batch | Ledger | First failure and cascade |
|---|---|---|
| A `migrations` suite (`reportA/results.tsv`, 57 rows) | 48 FAIL; 4 PASS (B-MIGRATE-ENV, B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE); 1 UNMET (B-MAINTENANCE-DRILL); 4 INFO | B-MIGRATE-CLI: `keep-A/mig-3-apply.log` and `mig-7-apply.log` end with `up failed: partial migration error (type:sql,version:20): ERROR: syntax error at end of input (SQLSTATE 42601)`; 00001–00019 applied (`catalog-p.log` 5-rollback/6-inspect at `app_objects_beyond_metadata=79`). The other 47 (B-MIGRATE-CATALOG/-IT, every *-CATALOG/-MIGRATION/-IT/-ORDERING row, B-IT-ISOLATION T-2r, B-LABELS-CATALOG/-IT/-ORDERING, B-RECOVERY-DRILL, B-REQUIRED-CHECKS) are cascades: no version-20 template database exists, so G064 `db_fresh` refused each invocation (`diagnostics/labels-catalog-it.log`: "no template record … beginify_tpl.rec"). |
| B `unit-web lint-web` (`reportB/results.tsv`, 14 rows) | 2 FAIL (B-UNIT-WEB; derived B-REQUIRED-CHECKS); 8 PASS (B-WEB-BOUNDARY, -TOKENS, -COMPONENTS, -DESIGN, -ASSET-CONTROLS, B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE) | `reportB/diagnostics/unit-web-check.log`: `eslint --max-warnings=0 .` → `tests/unit/labels-forms.test.mjs 78:7 warning 'GROUP_PATH' is assigned a value but never used`; `npm run check` stops there, so typecheck and `test:unit` did not run. |

**Diagnosis 1 — migration 00020 (confidence high, ~90%; stack-free).** Not a goose split. A stack-free reproduction of
goose v3.28's `sqlparser` rules (Up/Down; a StatementBegin/End block is one statement; otherwise a statement ends at a
line whose last non-comment word ends with `;`) yields 26 Up and 5 Down statements for 00020, each balanced (dollar
tags, parentheses, quotes) and `;`-terminated; the same splitter over 00001–00019 (which apply) flags nothing. The
defect is PL/pgSQL: `read_sql_construct` reads an `IF … THEN` condition up to the **first `THEN` token at parenthesis
depth 0** (it tracks parentheses only, not `CASE … END`). Lines 106–107 (precondition, Up statement 1) and 814–815
(postcondition, Up statement 26) compare the `::text` count `IS DISTINCT FROM CASE r.conname WHEN 'events_action_valid'
THEN 47 …` with a bare `CASE`, so the condition is cut at `… IS DISTINCT FROM CASE r.conname WHEN 'events_action_valid'`,
which the SQL parser rejects with exactly `syntax error at end of input` (42601) on the first `FOR r IN … LOOP` iteration
(DO blocks are not syntax-checked at compile time, so the preceding owner/role/version/relation preconditions ran and
passed). Consistent with 00019 applying (no bare CASE in an IF), steps 3 and 7 failing identically on fresh databases, no
`RAISE` text in the error, and 00017's precedent, which parenthesizes every CASE inside an IF (`IF … (CASE ex_role WHEN
'group_owner' THEN 3 … END)`, lines 443/497/552). Residual 10%: no PostgreSQL ran here. If the rerun still fails at
B-MIGRATE-CLI, the minimal real-daemon probe is one `psql -v ON_ERROR_STOP=1` of
`DO $$ DECLARE x int := 1; BEGIN IF x IS DISTINCT FROM CASE x WHEN 1 THEN 1 ELSE 2 END THEN RAISE NOTICE 'y'; END IF; END $$`
on any private PostgreSQL (expected: the same 42601), then `goose up` of 00020 alone with the error's `Where:` context.

**Fix 1 (`core/migrations/00020_labels.sql`, +3/−2):** parenthesize the two `CASE … END` expressions
(`IS DISTINCT FROM (CASE r.conname … END)`) with a one-line comment at the first; semantics, counts (47/79/73/11 and
52/81/80/12), check order and messages unchanged. The `CASE` at 816–817 is already inside `position(…)` (depth 1) and is
untouched. No pin moves: the `audit.events_attribute()` prosrc (`ce481e26…` version 19 / `cff36882…` version 20) is not
edited; `migBody` extracts `AS $$ … $$` function bodies only; only 00006–00009 are sha256-pinned (`migrations_test.go`);
the Down section is unchanged. No other in-scope SQL (run.sh/lib.sh, Go tests, 00001–00019) embeds the shape.

**Diagnosis/fix 2 (`web/tests/unit/labels-forms.test.mjs`, −1):** `GROUP_PATH` was dead (`groupPost` builds its path
from its `id` argument); removed. No assertion or lint-rule change. The failing eslint run covered the whole `web/` tree
and reported exactly one problem, so the other G020 web files are lint-clean by that run; typecheck and `test:unit`
remain unproven.

**Stack-free checks run (all exit 0):** splitter on patched 00020 up (26 statements, 0 flagged) and down (5, 0 flagged)
and on 00001–00019 both directions (0 flagged); `node --check` on the test file; `git diff --check`; both changed paths
match allow-list patterns 1 and 19 (29 patterns, 0 outside). Scratch (`/tmp/g020-r5-scratch/`) removed after handback.

**Real-daemon selectors to rerun (coordinator; failed batches first):**

1. `migrations` suite (`infra/validation/run.sh --base 66f00679 --allow-file
   management/v0/results/G020-scope-allowlist.txt` after `bash infra/ci/docker-desktop-mounts.sh`): B-MIGRATE-CLI steps
   3/7 are the proof of fix 1; then its cascades (B-MIGRATE-CATALOG/-IT, B-IT-ISOLATION at 88, B-LABELS-CATALOG/-IT/
   -ORDERING, the *-CATALOG/-MIGRATION/-IT/-ORDERING rows, B-RECOVERY-DRILL). The first real run of the round-2 T5
   predictions is still ahead.
2. `unit-web lint-web` (B-UNIT-WEB lint → typecheck → `test:unit` labels-client 17 / labels-forms 13; B-WEB-BOUNDARY).
3. Then round 4's list unchanged: b2 Go `-tags integration` selectors (`./internal/migrate -run
   '^(TestMigrationLifecycle|TestLabelsCatalog|TestLabelsMigration|TestAuditCatalog|TestMembershipCatalog)$'`;
   `./internal/tenancy/access -run '^(TestLabels|TestAccessPolicy|TestRLSIsolation|TestIsolationAttacks|TestAudit)$'` and
   `'^(TestLabelsOrdering|TestRLSOrdering)$'`; `./internal/mcp -run '^(TestMCPDiscovery|TestMCPTransport)$'`;
   `./internal/tenancy -run '^TestTenancyModel$'`); `lint lint-base lint-core unit-core`; `root`; `ops`
   (`labels-light`/`labels-dark`), `rename`, `off`; `ci-controls`/`controls`; final B-SCOPE and B-REQUIRED-CHECKS (133).
   Round 4's attributable passes (B-WEB-* ×5, B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE, B-MIGRATE-ENV) need no rerun
   while their inputs are unchanged.

**Not run (honest):** nothing real-daemon; no Go test, eslint, tsc, unit test, build, shellcheck or browser run; the
PostgreSQL behaviour is argued from `pl_gram.y` semantics and the 00017 precedent, not observed. Implementation rounds
completed: **5** under attempt **1** of B5 (no reset).

### Round 6 — reconciliation onto staging bf94dc6d via ca68ddc9 and f811246b (writer 7), 2026-10-24

Writer 7 (Opus 5.5; single write owner), isolated worktree `agent-<id>`, branch `g020-merge-r6` from
`origin/g020-impl` `ae80b3a9` (HEAD asserted). Mechanical reconciliation only: no Docker daemon (fake `docker` exiting 1
first in PATH, `DOCKER_HOST=tcp://127.0.0.1:1`), no test suite. Same B5 ledger, attempt 1 (no reset).

**(a) Target/source SHAs.** Staging advanced twice during this round, so it was reconciled in three merges (history
kept; one fast-forward push at the end). The final target is `bf94dc6d`:

| Step | Staging tip merged | Merge commit | Delta |
|---|---|---|---|
| 1 (intermediate) | `ca68ddc9` (13 commits after the last merged staging `66f00679`) | `c4ae77cb` (parents `ae80b3a9`, `ca68ddc9`) | below |
| 2 (intermediate) | `f811246b` (1 commit after `ca68ddc9`; Codex "Record RC10 qualification and release G020 hold") | `d852a1c1` (parents `c4ae77cb`, `f811246b`) | `PROGRESS.md`, `ROADMAP.md` only |
| 3 (target) | `bf94dc6d` (1 commit after `f811246b`; "Record RC10 promotion and verification") | `99339d64` (parents `ab7c6d4e`, `bf94dc6d`) | `PROGRESS.md`, `ROADMAP.md`, `results/ONEOFF-rc020rc10-candidate-20261023.md` only |

`ab7c6d4e` is the first Result commit for this section, written at the `f811246b` stage. The Result commit after
`99339d64` updates it to the final target. `g020-impl` before: `ae80b3a9`; merge base with staging before: `66f00679`.
All three merges were conflict-free (git ort; `infra/validation/run.sh` auto-merged). Ancestor proof: `git merge-base
--is-ancestor bf94dc6d HEAD` and `… ae80b3a9 HEAD` both exit 0.

Executable delta `66f00679..bf94dc6d` (`f811246b..bf94dc6d` adds no executable file): `infra/ci/validate.sh` (`progress_drain` prints compact `N/TOTAL ICON STATUS ID`
via printf; CI-only), `infra/validation/run.sh` (one new line after the libraries are sourced:
`if [[ -n "${PLATFORM_PROGRESS_FILE:-}" ]]; then export -n PLATFORM_PROGRESS_FILE; fi`), and one format sentence in
`docs/foundation/OPERATIONS.md`. The rest is management records: `PROGRESS.md`, `ROADMAP.md`,
`MAINT-progress-nested-run-result.md`, `ONEOFF-compact-validation-progress-20261023.md`,
`ONEOFF-rc020rc10-candidate-20261023.md`, `ONEOFF-rc020rc9-candidate-20261023.md`.

**(b) Changed management inputs.** At target `bf94dc6d` (recomputed with `git show origin/staging:<path> | sha256sum`;
matches the merged tree):

| File | sha256 at `bf94dc6d` |
|---|---|
| `management/v0/PROGRESS.md` | `1c8365ad2ee24e82f109459a5d1d25c7c762b21c71b7958c5fad8214ef7761e6` |
| `management/v0/ROADMAP.md` | `24c2afe7013dcc5a57328ea1f9557de37d0c492f3a77d53dffd1b46076e82d53` |
| `management/v0/results/ONEOFF-rc020rc10-candidate-20261023.md` (RC10 Result) | `c1c6f0c8d285f4f98b8f664ff6f341ba6d4a75e2d98fdc89a25df30964be0221` |

Superseded intermediate values:
- At `f811246b`: PROGRESS `5496ebf76d3ed6a87bf54e7dfd6ef6dbfe703de10853226e24710d2276128e99`, ROADMAP
  `fc04ed1ae3e09317b9c11199c20126bf33e5cbf8f3ca61eca2f28f2ed734f1cb`.
- At `ca68ddc9`: PROGRESS `8b43eceaa4599f9a9287aad773b8ca05cc581cf09bc9d6f1abee72f273a3dd86`, ROADMAP
  `b87939d608fc99a83235fb5a2707a55b11ed7e23241a59b5714ccc230610e19e`.

Effect: records only. The order is now G066 → RC9 promoted → MR !93 progress MAINT → RC10 full hosted qualification →
G020 → RC11 → preview → G057 → G021 → RC12.
- In `f811246b`, Codex recorded that RC10's qualification was complete and released the G020 hold. The "RC10 running"
  and "G020 resume hold" entries became "RC10 full hosted qualification PASS" and "G020 resumption released".
- In `bf94dc6d`, that PASS entry became "RC10 PROMOTED": the owner directed the merge of !94, and it merged to `main` as
  `181733f2` with `release-verify.sh merged` PASS. `ROADMAP.md` gained the matching promotion text (no stable release
  authorized; the G020 hold is cleared and work resumes after current-input reconciliation). The RC10 Result records
  the promotion and its verification.

G020 scope, Plan r2, Goal/Spec, allow-list and dependencies are unchanged.

Of the S0-bound inputs, only three appear in `66f00679..bf94dc6d`: `PROGRESS.md`, `ROADMAP.md` (reconciled above) and
`infra/validation/run.sh`, which G020 already edits; its effect is in (d). The other six changed files are not bound
inputs. The S0 binding JSON was not edited.

**(c) RC10 facts (from the staging records).** Annotated tag `0.2.0-rc.10`, object `368bba2c…`, at staging `79be8388`
(tree `d8a97395`); promotion MR !94; pipeline `2983225317` SUCCESS (fast `validate:checks` 35/35, full
`validate:integration` 93/93); release-verify qualified (124 required IDs) and promotable PASS. Live compact output was
reviewed: 93 unique sequential `N/93` rows, 0 failures, no fixture leakage. Promoted to `main` as `181733f2` (MR !94
merged; release-verify merged PASS). No stable release is authorized.

**(d) Effect of the harness delta on G020.**
- The compact `validate.sh` format is CI-only and is not an input to any G020 check.
- The run.sh `export -n` line affects only the live-progress display and nested launches.
- All of G020's run.sh edits are retained: the `labels_checks` placement, the `00020_labels`/`current_version 20` pins,
  and the B-TENANCY-HTTP `docker_run_owned wget-probe` label probe (10 `wget-probe` occurrences). The `export -n` line
  sits at line 43.

Verification (all exit 0 or empty):
- `bash -n infra/validation/run.sh infra/ci/validate.sh`.
- `git diff --check ae80b3a9 HEAD`.
- `git diff ae80b3a9 HEAD -- core/migrations web/tests` is empty, so round 5's fixes stand. Migration 00020 keeps both
  parenthesized `IS DISTINCT FROM (CASE r.conname … END)` IF conditions (lines 108 and 816), and `GROUP_PATH` stays
  removed.
- Allow-list check: 29 patterns, 73 paths in `git diff origin/staging..HEAD --name-only` against `bf94dc6d`, 0 outside.

**(e) Attempt ledger.** Unchanged: still B5 attempt 1, budget not reset. Round 5's diagnosis stands (parenthesized CASE
in 00020; unused `GROUP_PATH` removed). The real-daemon reruns are still pending, now with `--base bf94dc6d`, in this
order:
1. `migrations`, then `unit-web lint-web`.
2. The b2 Go `-tags integration` selectors.
3. `lint lint-base lint-core unit-core`.
4. `root`.
5. `ops` (`labels-light`/`labels-dark`), `rename`, `off`.
6. `ci-controls`/`controls`, then final B-SCOPE and B-REQUIRED-CHECKS (133).

Implementation rounds completed: **6** under attempt **1** of B5.

**(f) Cleanup inventory.** This writer used no daemon and created no Docker objects. Scratch `/tmp/g020r6bin/` (fake
`docker`, message files) is removed at handback. The frozen checkout `g020-frozen` (`c6a7351b`) is retained until the
new frozen checkout replaces it. This worktree is reclaimed by the coordinator.

**Not run (honest):** every real-daemon check above; no Go test, eslint, tsc, unit test, build, shellcheck or browser
run this round.

### Round 7 — migration apply failure (smallint = smallint[]): diagnosis and correction (writer 8), 2026-10-24

Source: branch `g020-r7` from `origin/g020-impl` `aaeb1530c0f9badb4d1904cf973cd9023d902a2b` (contains staging
`bf94dc6d5520b5dc47edb9c04011856070536b13`); fix commit `fb28bd97800e0054538468f25390bf56ee8f7267`, this Result the
following commit. Allow-list: only `core/migrations/00020_labels.sql` and this Result changed.

**(a) Evidence.** Round-6 real-daemon run A (`migrations`, `--base bf94dc6d`, source `aaeb1530`), read-only:
`<home>/.local/state/platform-evidence/g020-r6/consoleA.txt` lines 7 and 16 (B-MIGRATE-CLI steps 3-apply and
7-apply), `reportA/migrate-it.log`, `reportA/required-checks.tsv`. Exact error:
`up failed: partial migration error (type:sql,version:20): ERROR: operator does not exist: smallint = smallint[] (SQLSTATE 42883)`.
B-MIGRATE-CLI FAIL; 48 further checks cascade from that single cause (consoleA line 240: every *-CATALOG/-MIGRATION/
-IT/-ORDERING check, B-OBJ-FAULTS, B-IT-ISOLATION and B-RECOVERY-DRILL FAIL, B-MAINTENANCE-DRILL UNMET) because the G064
template is not created without a PASS B-MIGRATE-CLI. Run B (`unit-web lint-web`) PASS 9/9 (`consoleB.txt`, `reportB/`).

**(b) Cause (high confidence).** `core/migrations/00020_labels.sql`, postcondition DO block, former lines 763-765:
`a.attnum = ANY ((SELECT c.conkey FROM pg_constraint c WHERE ... conname = 'label_assignments_label_fkey'))`. PostgreSQL's
grammar matches `ANY` followed by a parenthesized SELECT, even doubly parenthesized, as `select_with_parens`, i.e. the
subquery form `expr = ANY (subquery)`: each returned row's column (`conkey smallint[]`) is compared with `attnum smallint`,
for which no `=` operator exists, so expression preparation raises 42883 and the goose transaction fails. The IF at former
lines 761-769 is the first expression after the constraint-list assertion (749-759); everything before it (preconditions,
tables, triggers, grants, policies, catalog swap, attribution body, and the postconditions for columns, column privileges,
grants/FORCE, 88 relations and the constraint list) therefore applied successfully in round 6 inside the rolled-back
transaction.

**(c) Fix** (one hunk, now lines 763-766; the assertion's meaning is unchanged: no nullable column of
`app.label_assignments` participates in `label_assignments_label_fkey`). Before:
```
AND a.attnum = ANY ((SELECT c.conkey FROM pg_constraint c WHERE c.conrelid = 'app.label_assignments'::regclass
                       AND c.conname = 'label_assignments_label_fkey')))
```
After:
```
-- IN (SELECT unnest(conkey)): a parenthesized SELECT after ANY is the subquery form (smallint = smallint[]).
AND a.attnum IN (SELECT unnest(c.conkey) FROM pg_constraint c WHERE c.conrelid = 'app.label_assignments'::regclass
                   AND c.conname = 'label_assignments_label_fkey'))
```
`unnest` over a catalog array is the idiom this file and 00010-00017 already use for `polroles`. With the constraint
absent both forms are false (the preceding constraint-list assertion already requires its presence).

**(d) Audit of the rest of 00020 (static; read against the PostgreSQL catalog documentation and 00016-00019).**
- Not-yet-executed Up code (former lines 766-830): the `pg_get_constraintdef((SELECT oid ...))` scalar subqueries
  (00019 l.453 idiom); the policy inventory (`unnest(p.polroles) x`, `x = rt` oid/regrole, `x::regrole::text`,
  `COLLATE "C"`: the 00017 l.1020 idiom; expected `ow+rt` order and polcmd letters r/a/w/d checked); `position(... IN
  a || b)` is a valid `b_expr`; the trigger inventory (`tgenabled = 'O'`, `NOT tgisinternal`; the five names in C
  order verified); the function loop (`proconfig IS DISTINCT FROM ARRAY['search_path=pg_catalog, pg_temp']`,
  `has_function_privilege('public', oid, 'EXECUTE')`, `proowner <> ow`: the 00019 l.450-452 idioms, and 00019's
  postcondition already asserts the same for `audit.events_attribute()`, so the CREATE OR REPLACE'd body keeps a passing
  ACL); the `aclexplode(q.proacl)` grantee check matches 00018/00019.
- md5 constants recomputed from the file text (Python, the dollar-quoted body exactly as `prosrc` stores it): the 00020
  Up body = `cff3688235276166e16d3ba9fcdf24b8` (line 807 matches); the Down body = `ce481e2677b791e5834a5f3da1af8956`
  = the 00019 Up body (lines 121 and 1182 match).
- Catalog counts recomputed by counting quoted values: Up 52/81/80/12 (line 816 matches), Down 47/79/73/11 (line 108
  matches).
- Down (lines 835-1188): NO FORCE/FORCE inside the transaction, DROP order (assignments before definitions, functions
  after the tables), restored catalog and body: nothing found.
- Searched the whole file for: other `ANY ((SELECT`/`ALL (`/`IN` over array-returning subqueries (none), array-vs-scalar
  operators (none), regclass/regrole/regprocedure casts (all on literals or `to_reg*`), `IS DISTINCT FROM` precedence
  (CASE operands parenthesized at 108/816; row comparisons parenthesized), ambiguous columns (every catalog query is
  single-table or aliased), catalog column names (`conkey`, `confdeltype`, `convalidated`, `polcmd`, `polpermissive`,
  `polroles`, `tgenabled`, `tgisinternal`, `proconfig`, `prosecdef`, `proacl`, `attgenerated`, `attnotnull`: all
  correct), PL/pgSQL IF/THEN (every IF ... THEN/END IF balanced; multi-line conditions end with THEN outside
  parentheses). Nothing else found. Lines 40-759 are additionally proven by round-6 execution and are unchanged.
- Go tests (`core/internal/migrate/labels_integration_test.go`, `core/internal/tenancy/access/labels_*test.go`,
  `core/migrations/migrations_test.go`) reference object names, `label_assignments_target_shape`, `attnum`-ordered column
  lists and function bodies via `migBody` (function bodies only); none pins the changed DO-block text or `conkey`. No test
  change needed.

**(e) Static checks run.** `git diff --check` clean. `git diff aaeb1530 fb28bd97 --stat`: one file, +3/-2 (one comment
line, two lines rewritten). No SQL parser is available on the host (no `psql`/`pg_query`), so no parse beyond the manual
read; no database, Docker daemon, Go, eslint, tsc, unit test, build or browser run this round (fake `docker` first in
PATH, `DOCKER_HOST=tcp://127.0.0.1:1`).

**(f) Next selectors for the coordinator** (new head, all with `--base bf94dc6d`, in this order):
1. `migrations` (B-MIGRATE-CLI first, then its cascade: B-MIGRATE-CATALOG/-IT, every *-CATALOG/-MIGRATION/-IT/-ORDERING,
   B-OBJ-FAULTS, B-IT-ISOLATION, B-RECOVERY-DRILL, B-MAINTENANCE-DRILL, B-LABELS-CATALOG/-IT/-ORDERING).
2. `unit-web lint-web` only if a web or lint-web input changed: none did (`git diff aaeb1530 HEAD -- web` is empty), so
   round-6 run B (9/9 PASS) remains attributable.
3. The b2 Go `-tags integration` selectors.
4. `lint lint-base lint-core unit-core`.
5. `root`.
6. `ops` (`labels-light`/`labels-dark`), `rename`, `off`.
7. `ci-controls`/`controls`, then final B-SCOPE and B-REQUIRED-CHECKS.

**(g) Attempt ledger.** Unchanged: still B5 attempt 1, budget not reset. Implementation rounds completed: **7** under
attempt **1** of B5.

**(h) Cleanup inventory.** This writer used no daemon and created no Docker object. Scratch `/tmp/g020r7-fakebin/` (fake
`docker`) and `/tmp/g020r7-msg/` (commit messages) are removed at handback; the `g020-r7` worktree is left for the
coordinator to reclaim.

### Round 8 — first full real-daemon migrations run (7 failures): diagnosis and correction (writer 9), 2026-10-24

Source: branch `g020-r8` from `origin/g020-impl` `5e9134848d22a6deeb689e2f0f0d848db85257eb` (contains staging
`bf94dc6d5520b5dc47edb9c04011856070536b13`); fix commits `b75f81f8` (matrix pin), `36ea19de` (fixture teardown),
`4fa83bf7` (attack G1 actor), `a6edfb29` (LAB-5); this Result the following commit. Allow-list: only
`core/internal/tenancy/access/{integration_test,labels_fixture_test,attack_expected_test,attack_integration_test,labels_integration_test}.go`
and this Result changed. No daemon, database, Go toolchain, SQL parser or browser was used (fake `docker` first in PATH,
`DOCKER_HOST=tcp://127.0.0.1:1`); every finding below is read from the preserved logs and the source.

**(a) Evidence (read-only).** Round-7 run A (`migrations`, `--base bf94dc6d`, source `5e913484`, frozen checkout
`.claude/worktrees/g020-frozen7`, launched by `runA.sh` without `--owned-registry`), `rc=1`:
`<home>/.local/state/platform-evidence/g020-r7/consoleA.txt`, `reportA/results.tsv` (7 non-PASS rows plus
B-REQUIRED-CHECKS FAIL and B-MAINTENANCE-DRILL UNMET as before), `reportA/go-tests/{identity-catalog,policy-model,rls-it,
obj-it-attack,domains-catalog,labels-it}.log`, `reportA/go-tests.tsv`, `reportA/diagnostics/it-isolation.log`,
`reportA/db-lifecycle.tsv` (one `db_fresh` per invocation; `domains-catalog` 00:41:35.9, `domains-it` 00:41:42.5: no
overlap). Run B (`lint lint-base lint-core unit-core` 18/18) and round 6's `unit-web lint-web` 9/9 stay attributable:
no input of theirs changed except gofmt's view of the five integration-tagged test files (see (h)). The run is
preserved as FAILED; nothing below re-labels it.

**(b) Per-failure table.** Class: (a) pinned expectation G020 legitimately changes; (b) defect in G020 migration/facade/
HTTP/test code; (c) pre-existing or environment/order dependence.

| Check | Subtest | Exact assertion (log) | Class | Cause | Confidence | Disposition |
|---|---|---|---|---|---|---|
| B-IDENTITY-CATALOG | `TestIdentityCatalog/objects_and_owner` | `core/internal/identity/store_test.go:77-78`: `tables=... invitations,label_assignments,label_definitions,login_transactions,... rls=20` vs `want` (19 tables) and `rls != 18` | (a) | 00020 adds two forced-RLS tables to schema `app`; the exact-set pin (bumped by G017/G019/G058/G062 the same way) does not include them. The forbidden-column check (lines 86-94) runs over `identityTables` only and is unaffected. | high | **BLOCKER: `core/internal/identity/store_test.go` is not on the allow-list.** Exact change in (d). |
| B-POLICY-IT | `TestAccessPolicy/IT-MATRIX-1` | `integration_test.go:1499: asserted 484 cells, want 418` | (a) | The loop asserts 6+6+5+5 = 22 cells per capability (flag x selected, org_member unasserted when selected); three label capabilities add 66 = 418 -> 484, exactly the observed count; no cell disagreed. The "FAIL/4 footprint" in results.tsv is the wrapper counting the failed package's subtests: `TestAccessFootprint` is 4/4 PASS in the log. | high | Fixed `b75f81f8`: pin 418 -> 484 with the G018/G019 comment extended. |
| B-RLS-IT | `TestRLSIsolation/RLS-LABEL-FIXTURE .../(g) teardown` | `labels_fixture_test.go:300: ... ERROR: cannot drop table app.label_fixture_targets because other objects depend on it (SQLSTATE 2BP01)` | (b) | The widened `label_assignments_select` policy's `EXISTS (SELECT 1 FROM app.label_fixture_targets ...)` creates a `pg_depend` entry; the teardown dropped the table before restoring the policy. "FAIL/7 pool" is the same wrapper artefact: `TestRLSPool` 7/7 PASS, `TestRLSFootprint` PASS. | high | Fixed `36ea19de`: teardown order = reverse of install (FK, column, policy, CHECK, table). T5 (g) comparisons unchanged. |
| B-LABELS-IT | `TestLabels` (parent; all 17 subtests PASS) | same `labels_fixture_test.go:300` 2BP01 in the parent's `t.Cleanup` | (b) | Same cause as the B-RLS-IT row (one fixture teardown shared by both suites). | high | Same fix `36ea19de`. |
| B-LABELS-IT | `TestLabels/LAB-5 input refusals ...` | `labels_integration_test.go:353: invalid name: <nil> (field ""), want invalid_input "invalid_name"` twice (the `" Leading"` and `"Trailing "` inputs) | (b) test expectation | `tenancy.ValidateLabelName` reuses `ValidateName` and returns the trimmed stored form (Plan r2 E18 "reuse ... trimmed"; unit tests `TestLabelNameTrim` and the "bound applies after trimming" case, accepted in unit-core 18/18; `labels.go:112-116` doc; TENANCY.md:23 convention). The DB CHECK `name = btrim(name)` admits the stored form. LAB-5's refusal expectation contradicts the approved rule, not the facade. TENANCY.md:402 "validated, never normalized" is the sentence on the database constraint; the web client's `value.trim() !== value` refusal (`labels-client.ts:102`) is stricter client-side validation and is not the core contract. | medium-high (the artefacts conflict; the accepted Plan and unit tests decide) | Fixed `a6edfb29`: refusal list = `""`, `"   "`, 81 code points, embedded tab, `"a b"`, plus a no-write check; a padded name is asserted to succeed and be stored as its trimmed form, then deleted. If Codex intends the stricter core contract instead, the alternative is a one-line `t != s` refusal in `ValidateLabelName` plus flipping the two unit cases; not done here. |
| B-OBJ-IT | `TestIsolationAttacks/G/G1` | `attack_integration_test.go:1534: ISO-ORACLE-DISAGREEMENT G1/G1-07: visible [], oracle [lG1 lG1tm]` (ISO-ORACLE 163 cases, 1 disagreement; "attack FAIL/24" = 24 subtests, 1 failing) | (b) test expectation | G runs after C3; `C3-06 RemoveOrganizationMembership(gOwner, orgA)`. With no Organization binding `rls_org_role()` is NULL and `label_definitions_select` (00020 l.340-343: `organization_id = O AND R IS NOT NULL AND ...`) admits nothing. The G1 rows (new in G020) ignored the world mutation that AMB-24 already records ("gOwner is exercised only by inherited tests before C3"). The policy itself is correct: G1-11 (gMember2, same shape) PASSED, and the Group-owned branch without a selected Group is `(G IS NULL OR group_id = G) AND (GMGR OR ...)`. | high | Fixed `4fa83bf7`: G1-07..G1-10 use `gAdmin` (org_member + group_admin of g1; bindings intact through G, used by E3-03) with the same Visible/Effect values; `atkShape` user swapped accordingly. The assertion's meaning (a Group manager with no selected Group reads exactly its Group's definitions and never Organization-owned rows/assignments) is preserved. |
| B-DOMAINS-CATALOG | `TestDomainsCatalog/CT-DOM-12` | `domains_integration_test.go:510: down after revocation and invalidation: partial migration error (type:sql,version:12): ERROR: refusing to roll back 00012: live legacy browser sessions would become usable by the global reader (SQLSTATE P0001) ""` ("FAIL/4 subtests" = 4 subtests, 1 failing; the `:162`/`:246` "detector ... got" lines are `t.Logf` output of the PASSING CT-DOM-1/CT-DOM-NEG) | **undecided**, (c)-leaning | See (e). | low | No edit (file outside the allow-list anyway); reproducer and diagnostic in (e). |
| B-IT-ISOLATION | T-9, T-10 | `it-isolation.log`: `T-9 UNMET not exercised: runner unavailable (no-registry); eligible invocations take the per-invocation container path` (T-10 identical); T-1..T-8 PASS | (c) environment | `run.sh:2565-2573` `_iti_runner`: `it_runner_precondition` (`lib.sh:1159-1175`) needs `OWNED_REG` (a registry file with `cid/`, token, repo) plus the local durable marker `STATE_ROOT/g003-pending/local-*.tsv`; `runA.sh` launched `run.sh` without `--owned-registry`, so outside GitLab CI the two runner controls are UNMET by design ("unexecuted coverage is never PASS"), and B-REQUIRED-CHECKS fails on any UNMET required ID (`run.sh:6690`). Not a knock-on of the Go failures. | high | No code change. Coordinator remedy in (f). |
| B-REQUIRED-CHECKS | aggregate | `required checks missing, duplicated or not PASS: ...` | — | Aggregate of the rows above. | — | Resolves with them. |

**(c) Fixes (static re-read for compile-level correctness; no Go toolchain).**
- `integration_test.go` (+3/-2): `cells != 484` and the comment; no identifier change.
- `labels_fixture_test.go` (+8/-6): the seven teardown statements reordered (strings only), docstring updated.
- `attack_expected_test.go` (+9/-7): `Actor`/`Context`/`Source`/`Note` strings of four rows; field alignment kept as the
  neighbours (gofmt-style verified by eye: `Visible:`/`Source:`/`Note:` two-space, `Effect:`/`Pair:`/`Source:` one-space
  as before). `attack_integration_test.go` (+3/-1): `u["gOwner"]` -> `u["gAdmin"]` (already used at lines 967/980) and a
  two-line comment. No citation (`objects:`/`rls:`) added, so ISO-CITATIONS is unaffected; the oracle case count (163)
  is unchanged.
- `labels_integration_test.go` (+16/-1): `p, err := svc.CreateLabelDefinition(...)` (new `p`, existing `err`),
  `LabelDefinitionView.Name`/`.ID` (labels.go:26-32), `e.labCount` and `c := e.canaries` as in the surrounding
  subtests; the ` ` case is the Go escape (a literal U+2028 produced by the editor was replaced; `git diff`
  carries no non-ASCII byte).

**(d) BLOCKER needing a Codex scope amendment: `core/internal/identity/store_test.go` (not on the allow-list).**
Required change (class (a) pin, the same bump G017/G019/G058/G062 made), lines 73-77:
```
-			// and from version 18 the G062 signup_sessions table, forced as well (TestSignupCatalog).
+			// and from version 18 the G062 signup_sessions table, forced as well (TestSignupCatalog), and from version 20 the
+			// G020 label_assignments and label_definitions tables, forced as well (TestLabelsCatalog).
			want := "business_profiles,enrollment_requests,enrollment_restrictions,enrollment_sessions,group_business_profiles,groups," +
-				"identity_principals,identity_provider_binding,invitations,login_transactions,mcp_connections,organization_branding," +
+				"identity_principals,identity_provider_binding,invitations,label_assignments,label_definitions,login_transactions," +
+				"mcp_connections,organization_branding," +
				"organization_domains,organizations,role_bindings,sessions,signup_sessions,stored_objects,users"
-			if tables != want || owners != it.cfg.OwnerRole() || rls != 18 {
+			if tables != want || owners != it.cfg.OwnerRole() || rls != 20 {
```
Justification: the observed `tables=`/`rls=20` string in `identity-catalog.log:3` is exactly this set; the other three
subtests PASS; the forbidden-column pattern runs over `identityTables` only. Until this lands, B-IDENTITY-CATALOG (and
therefore B-REQUIRED-CHECKS) stays FAIL on every `migrations` run.

**(e) Undecided: CT-DOM-12 (`core/internal/migrate/domains_integration_test.go`, also outside the allow-list).**
Facts established from the log and source: the subtest `reset`s, migrates 0 -> 11, inserts exactly three unbound browser
sessions as the superuser (l.332-336), migrates to 12, is correctly refused once (l.385), revokes `legacyRevoke` as the
runtime (1 row, audited +1), invalidates the two unrevoked rows as the superuser (`RowsAffected == 2`, l.433-436) and
verifies, on the same three ids and with the same predicate as the 00012 guard, `0 live, 1 revoked` (l.438-442). The
probes between l.447-505 are each rolled back and none can produce an unbound live browser row (the only unbound browser
INSERT is expected and observed refused; the admitted one is `kind = 'mcp'`; the bound control has `origin` set, which
would trip guard 1, not guard 2). The second down (l.508) is nevertheless refused by guard 2
(`00012` l.277-279: `kind='browser' AND origin IS NULL AND revoked_at IS NULL AND idle_expires_at > now() AND
absolute_expires_at > now()`, evaluated by the owner under `NO FORCE`). I could not derive, from the preserved evidence,
either a fourth row or a `now()` earlier than the invalidation (goose v3.28 begins a fresh transaction per migration on
its session-lock connection; `rlsReset` commits its `BeginTx`; `setup` is autocommit). G020 contributes nothing that
executes in this subtest except `reset`'s `labelDrops` (seven `DROP ... IF EXISTS` on the version-20 clone, before the
existing drops) and the version-20 template itself; the harness selector (`it_fresh` + `^TestDomainsCatalog$`) is
byte-identical on staging `bf94dc6d` (run.sh l.4100). The test also failed in rounds 4 and 6, but there as a cascade of
the unapplied migration, so this is its first observation on an applied version-20 template. Minimum reproducer for the
coordinator (both on the merged `g020-impl` head and on `bf94dc6d` with the same template harness): the `migrations`
suite's `domains-catalog` invocation alone, or `go test -tags integration -run '^TestDomainsCatalog$/CT-DOM-12' ./internal/migrate/`
through `it_fresh`; if it fails on both trees it is pre-existing under the G064 template (RC10's hosted full
qualification passed on its tag source, `ONEOFF-rc020rc10-candidate-20261023.md` l.39, which I did not cross-check at
the job-log level). A one-statement diagnostic in a scratch copy of the test, run before l.508,
`SELECT id, kind, origin, revoked_at, idle_expires_at, absolute_expires_at, now() FROM app.sessions`, decides between
the two hypotheses; it is not committed here (out of scope).

**(f) B-IT-ISOLATION remedy (no code change).** Launch the next `migrations` run with the G064 runner registry so T-9/
T-10 execute: either through `infra/ci/validate.sh` local mode (it creates `$EV/owned` and the
`STATE_ROOT/g003-pending/local-<job>.tsv` marker and passes `--owned-registry "$REG/owned.tsv"`, l.828), or by passing
`--owned-registry <registry.tsv>` to `run.sh` together with a valid local marker (`lib.sh` l.1159-1175: `job`
`local-<12 hex>`, `token` 32 hex matching `<dir>/token`, `repo` = the real checkout path, `<dir>/cid/`, mode 700).
Without it the suite cannot reach B-REQUIRED-CHECKS PASS locally whatever the Go results are.

**(g) Static checks run.** `git diff --check` clean after every edit; `git diff --stat` 5 files, +40/-17 before the
Result; no non-ASCII byte added. No gofmt/vet/compile/test/database/daemon/browser run this round (see the header).

**(h) Next selectors for the coordinator** (new head, all with `--base bf94dc6d`, in this order):
1. `migrations` again, launched with the owned registry per (f) (all seven failures are in it). Expected residuals:
   B-IDENTITY-CATALOG FAIL until (d) lands (scope amendment), and CT-DOM-12 per (e); everything else should turn PASS.
2. The b2 Go `-tags integration` selectors (`./internal/tenancy/access/` `^TestAccessPolicy$`, `^TestRLSIsolation$`,
   `^TestIsolationAttacks$`, `^TestLabels$` are the ones whose files changed).
3. `root`, then `ops` (`labels-light`/`labels-dark`), `rename`, `off`.
4. `ci-controls`/`controls`, then final B-SCOPE and B-REQUIRED-CHECKS.
5. `lint lint-base lint-core unit-core`: `unit-core` must rerun (core-check's gofmt sees the five edited integration-tagged
   files); `lint-core` (static import boundary) and `lint`/`lint-base` have no changed input but ride along in the same
   batch (B-SECRET-CANARY needs `lint lint-base` together). `unit-web lint-web`: no web input changed; round-6 run B
   remains attributable.

**(i) Attempt ledger.** Unchanged: still B5 attempt 1, budget not reset. Implementation rounds completed: **8** under
attempt **1** of B5.

**(j) Cleanup inventory.** This writer used no daemon and created no Docker object. Scratch `/tmp/g020r8-fakebin/` (fake
`docker`) and `/tmp/g020r8-msg/` (commit messages) are removed at handback; the `g020-r8` worktree
(`.claude/worktrees/agent-<id>`) is left for the coordinator to reclaim.

### Round 9 — Codex rulings on Round 8 and the store_test pin (writer 10), 2026-10-24

Source: branch `g020-r9` from `origin/g020-impl` `bfcca2a3a2b12055180c97d8b3757fe3c1076b68` (contains staging
`bf94dc6d5520b5dc47edb9c04011856070536b13`); source + allow-list commit `12d380be`; this Result the following commit.
No daemon, database, Go toolchain or browser was used (fake `docker` first in PATH, `DOCKER_HOST=tcp://127.0.0.1:1`).

**(a) Scope amendment (Codex ruling, handoff `ae55426b`, 2026-10-24).** Narrow migration-derived test-pin amendment
for the existing authorized G020 task, following the `ed0dd821` precedent (dated amendment comment appended; header
and existing lines intact). The allow-list gains exactly `^core/internal/identity/store_test[.]go$` (30 patterns),
permitting only the round 8 (d) pin for the two new forced-RLS tables `label_assignments` and `label_definitions`; no
other identity file is authorized. Change applied exactly as shown in round 8 (d), in
`core/internal/identity/store_test.go` (+5/-3): the version comment extended (version 20, G020, TestLabelsCatalog), the
`want` table set gains `label_assignments,label_definitions` between `invitations` and `login_transactions`, and
`rls != 18` -> `rls != 20`. The forbidden-column inventory and every other assertion are unchanged. The new expected
string equals, character for character, the observed `tables=` string in round-7 run A
`<home>/.local/state/platform-evidence/g020-r7/reportA/go-tests/identity-catalog.log` line 3 (compared by
script; byte-wise sorted, as returned by `string_agg`). Allow-list SHA-256: before
`ee59433fe7aca0351cb785d5227efcaef308202d8a128e48708399f3f4ad5304`, after
`34e0b05eb4b04b327ad02b5ddc142c115573cf2f63e497576a383e1cb37dceac`. The bound input binding
`management/v0/results/G020-implementation-input-binding-20261021.json` is untouched. Static checks: `git diff --check`
clean; `git diff --name-only bfcca2a3..12d380be` = `core/internal/identity/store_test.go`,
`management/v0/results/G020-scope-allowlist.txt`, both matched by non-comment allow-list patterns (negative control
`core/internal/identity/store.go` does not match); no non-ASCII byte added. No gofmt/compile/test run this round.

**(b) CT-DOM-12 ruling.** Not edited and not waived (cause unproven; `core/internal/migrate/domains_integration_test.go`
is outside the allow-list). Its status remains FAIL/not-PASS as recorded in round-7 run A. A comparison on the same
harness and template (G020 head vs staging baseline `bf94dc6d`) is pending and will be recorded by the coordinator. If
it fails only on G020, the G020-specific cause and the minimal in-scope remedy are due; if it also fails on the
baseline, it is a pre-existing failure needing a separate qualification disposition.

**(c) LAB-5 ruling.** Trimming is accepted: Plan section 1.1 E18 reuses `ValidateName`, the stored form is trimmed and
the 80-code-point bound applies; `ValidateLabelName` and `TestLabelNameTrim` implement it. The round-8 test change
(`a6edfb29`) verifying that padded input is stored trimmed stays. The core contract and the web validator are not
changed.

**(d) Attempt ledger.** Unchanged: still B5 attempt 1, budget not reset. Implementation rounds completed: **9** under
attempt **1** of B5.

**(e) Cleanup inventory.** This writer used no daemon and created no Docker object. Scratch `/tmp/g020r9-fakebin/`
(fake `docker`) and `/tmp/g020r9-msg/` (commit messages, path/pattern lists) are removed at handback; the `g020-r9`
worktree (`.claude/worktrees/agent-<id>`) is left for the coordinator to reclaim.

### Round 10 — round-8 run A residuals, harness pin correction and staging 36467b87/3980f12a (writer 11), 2026-10-24

Source: branch `g020-r10` from `origin/g020-impl` `934e3e9567d27705739e49c44b315ceeb1751660`; merge of staging
`36467b87896ec5236a9999a36539abf5386c2b41` `e6e775ec`; harness pin commit `e8404eb3`; merge of the final staging target
`3980f12ad5c8241597fe6607a3b3deadf4c77f70` `aac195d5`; this Result the following commit.
No daemon, database, Go toolchain or browser was used (fake `docker` first in PATH, `DOCKER_HOST=tcp://127.0.0.1:1`).

**(a) Round-8 runs (coordinator; evidence `<home>/.local/state/platform-evidence/g020-r8/`).** Both launched
2026-10-24T01:22:24Z from the frozen checkout `.claude/worktrees/g020-frozen8` at `bfcca2a3` with
`--base bf94dc6d5520b5dc47edb9c04011856070536b13 --allow-file management/v0/results/G020-scope-allowlist.txt`.

| Run | Selectors | Outcome | Evidence |
|---|---|---|---|
| A | `migrations` (with `--owned-registry .../g020-r8/owned/owned.tsv`, see (e)) | rc=1, end 01:46:28Z: **FAILED** | `reportA/results.tsv`, `reportA/go-tests/`, `consoleA.txt` |
| B | `lint lint-base lint-core unit-core` | rc=0, end 01:28:57Z: every check PASS, B-REQUIRED-CHECKS PASS | `reportB/results.tsv`, `consoleB.txt` |

Run A, from `reportA/results.tsv`: every migrations check PASS except exactly two FAIL rows, and B-REQUIRED-CHECKS FAIL
solely because of them (`required checks missing, duplicated or not PASS: B-IDENTITY-CATALOG(1,FAIL)
B-LABELS-IT(1,FAIL)`):
1. **B-IDENTITY-CATALOG** (`identity catalog`): the known `store_test.go` pin (round 8 (d)), fixed in `12d380be`
   (round 9 (a)), contained in `934e3e95`; not yet re-run.
2. **B-LABELS-IT** (`TestLabels: 15 subtest results (logcheck: ok)`): the Go test itself passed
   (`reportA/go-tests/labels-it.log`: `--- PASS: TestLabels (1.41s)` with 15 passing subtests, none skipped or failed);
   the harness pin required 17. Corrected in (c).

PASS in run A, among them: B-POLICY-IT, B-RLS-IT, B-OBJ-IT, B-LABELS-CATALOG (4/4 CT-LAB, 4/4 MG-LAB),
B-LABELS-ORDERING (4/4 O-LAB), B-IT-ISOLATION (T-9 runner residue and T-10 runner fault executed with the owned
registry; in round-7 run A it was UNMET, `required controls not exercised: T-9 T-10`), B-DOMAINS-CATALOG,
B-DOMAINS-IT, B-RECOVERY-DRILL, B-MAINTENANCE-DRILL, and B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE. The run is preserved
as FAILED (rc=1); no row is rewritten.

**(b) CT-DOM-12 two-run record and Codex disposition.** B-DOMAINS-CATALOG (TestDomainsCatalog, CT-DOM-12) FAILED in
round-7 run A (`g020-r7/reportA/results.tsv`: `domain catalog (FAIL/4 subtests)`; round 8 (e)) and PASSED 4/4 in
round-8 run A on head `bfcca2a3` (`g020-r8/reportA/results.tsv`; selector `^TestDomainsCatalog$`, template base
`bf94dc6d`). Its test file `core/internal/migrate/domains_integration_test.go` is outside the G020 allow-list and
unchanged by G020. The difference is intermittent and unexplained: the round-7 run had no owned registry and ran
concurrently with a lint batch; the round-8 run had the registry and also ran concurrently with a lint batch. Both
outcomes are preserved. **Codex disposition (relayed by the coordinator, 2026-10-24):** no G020-versus-staging
baseline comparison is required; the round-8 PASS is a successful non-reproduction; the round-7 run stays FAILED; the
cause is unknown; this is not a waiver and not evidence that the failure is pre-existing; no code edit and no extra
baseline test. (Staging `3980f12a` records this check in PROGRESS.md.)

**(c) Harness pin correction (`e8404eb3`, `infra/validation/run.sh` +3/-3, labels_checks()).**
`core/internal/tenancy/access/labels_integration_test.go` has 15 `t.Run` subtests covering the 17 IDs LAB-1..17:
LAB-1..5, one subtest `LAB-6..8` (assign/unassign, already_assigned, not_assigned: the three IDs are asserted inside
it), LAB-9..17. The harness's `it_count` counts `--- PASS: TestLabels/` lines, so the correct pin is 15:
- before: `[[ "$n" == 17 ]]`; PASS message `... runtime login: 17/17 subtests (LAB-1/2 ...`; comment above
  labels_checks `17 LAB subtests (LAB-1..17)`;
- after: `[[ "$n" == 15 ]]`; PASS message `... runtime login: 15/15 subtests covering LAB-1..17, LAB-6..8 being one
  subtest (LAB-1/2 ...` (rest unchanged); comment `15 LAB subtests covering LAB-1..17 (LAB-6..8 are one subtest)`.
The test file and its assertions are unchanged; the FAIL message is unchanged. Other pins re-verified against the
sources and the preserved round-8 logs: TestLabelsCatalog 4 (CT-LAB-1..4) and TestLabelsMigration 4 (MG-LAB-1..4)
in `core/internal/migrate/labels_integration_test.go` and `reportA/go-tests/labels-catalog.log`; TestLabelsOrdering 4
(O-LAB-1..4) in `core/internal/tenancy/access/rls_ordering_integration_test.go` and `labels-ordering.log`; no nested
subtests in any of them. No other count or ID pin of the TestLabels family exists under `infra/`
(`required-checks.tsv` names B-LABELS-CATALOG/IT/ORDERING without counts; `profile-paths.tsv` a comment only).
Static: `bash -n infra/validation/run.sh` OK; `git diff --check` clean; no non-ASCII byte added;
`git diff --name-only origin/staging..e8404eb3` (74 paths) all matched by the 30 non-comment allow-list patterns; rechecked against `3980f12a` at the final head.
shellcheck is unavailable here.

**(d) Merges of staging `36467b87` (`e6e775ec`) and `3980f12a` (`aac195d5`), both without conflicts.** Each is one
management-only commit changing only `management/v0/PROGRESS.md`: `36467b87` (round-eight rulings record, one line
added after `bf94dc6d`; intermediate PROGRESS.md SHA-256
`ea4e1da5a2356b91a77bba51310a7057dac6f2b4c78ee4a9394a2ed7a49afd6b`) and the final target `3980f12a` ("Record G020
nonreproducing domain check", one line changed; PROGRESS.md
`f3df5e4d54c71e7506e5624d16996efecaef8f274317d3f0a1ee392d43589c70`). Unchanged across both: ROADMAP.md
`24c2afe7013dcc5a57328ea1f9557de37d0c492f3a77d53dffd1b46076e82d53` (unchanged since `bf94dc6d`) and the RC10 Result
`management/v0/results/ONEOFF-rc020rc10-candidate-20261023.md` `c1c6f0c8d285f4f98b8f664ff6f341ba6d4a75e2d98fdc89a25df30964be0221`.
No source, harness or test input changed in either merge, so no earlier attributable outcome is invalidated by them.
The next runs use `--base 3980f12ad5c8241597fe6607a3b3deadf4c77f70`.

**(e) Owned-registry and marker procedure used for round-8 run A (so T-9/T-10 execute).** The coordinator created a
mode-700 registry directory owned by the operator (`g020-r8/owned/`, with `cid/`, `token` and `owned.tsv` whose
single `job` (`local-<12 hex>`), `token` (32 hex, equal to `owned/token`) and `repo` (the real checkout path) rows
satisfy `it_runner_precondition` in `infra/validation/lib.sh`), passed `--owned-registry .../owned/owned.tsv` to
`run.sh`, and wrote the durable local marker `STATE_ROOT/g003-pending/<job>.tsv`
(`STATE_ROOT` = `~/.local/state/platform-evidence/g003-local`; here `g003-pending/local-b8f3383fb094.tsv`, path in
`g020-r8/marker-path.txt`) with `registry` (the registry directory), `stage` (`launching`/`launched`) and `repo` rows.
**The marker file must be removed after the run**; it was removed (verified absent at this writing). The same
procedure applies to the next `migrations` run.

**(f) Attempt ledger and next selectors.** Unchanged: still B5 attempt 1, budget not reset. Implementation rounds
completed: **10** under attempt **1** of B5. Next selectors (new head, `--base 3980f12a`, in this order):
1. `migrations` again, with the owned registry and marker per (e) (turns B-IDENTITY-CATALOG and B-LABELS-IT; watch
   B-DOMAINS-CATALOG per (b)).
2. `lint lint-base lint-core unit-core` (`core/internal/identity/store_test.go` and `infra/validation/run.sh` changed
   since run B).
3. `root`, then `ops` (`labels-light`/`labels-dark`), `rename`, `off`.
4. `ci-controls`/`controls`, then final B-SCOPE and B-REQUIRED-CHECKS.

**(g) Cleanup inventory.** This writer used no daemon and created no Docker object. Scratch `/tmp/g020r10-fakebin/`
(fake `docker`) and `/tmp/g020r10-*.txt` (commit messages, diff, path/pattern lists) are removed at handback; the
`g020-r10` worktree (`.claude/worktrees/agent-<id>`) is left for the coordinator to reclaim. The
round-8 evidence directory was read only.

### Round 11 — run D (`ops`) E-LAB-5 failure: diagnosis and test correction (writer 12), 2026-10-24

Source: branch `g020-r11` from `origin/g020-impl` `27097759fbf5e072df266889c0c2fd3320f43169` (staging target
`3980f12ad5c8241597fe6607a3b3deadf4c77f70` verified unchanged at start); the correction is one commit on it, this
Result in the same commit. No daemon, database, Go or Node toolchain, browser or network was used.

**(a) Failure record (coordinator run D; evidence `<home>/.local/state/platform-evidence/g020-r10/`, read
only).** `runD.sh`: frozen checkout `.claude/worktrees/g020-frozen9` at `27097759`, `run.sh --base 3980f12a
--allow-file management/v0/results/G020-scope-allowlist.txt ops`, started 2026-10-24T02:18:49Z, rc=1 at 02:38:24Z.
`reportD/results.tsv`: exactly two non-PASS rows, `B-IDENTITY-E2E FAIL identity browser scenarios` and
`B-REQUIRED-CHECKS FAIL required checks missing, duplicated or not PASS: B-IDENTITY-E2E(1,FAIL)`; every other row
PASS (among them B-IDENTITY-IT 95/95, B-IDENTITY-PROBES, B-MCP-E2E 6/6, B-IDENTITY-PRIVACY, B-IDENTITY-CONFIG, the
ops/outage/persistence variants). `consoleD.txt`: `browser stats (expected skipped flaky unexpected): 169 0 0 2`; the
two unexpected tests are the same test in both labels projects, `[labels-light]` and `[labels-dark] ›
tests/browser/labels.spec.ts:438 › [E-LAB-5] forged label writes are refused as Not permitted and change nothing`,
first failing assertion `labels.spec.ts:452 expect(assign.status).toBe(403)` — **Expected 403, Received 404** —
after alice (plain Organization member) POSTed `action=assign&label_id=<Organization label>` with her own valid CSRF
token to `/groups/<research>/labels/action`. The run stopped at that first expect, so the later steps of the test
never executed. E-LAB-1..4 and E-LAB-6, S-LANDMARKS, S-NAV passed in both projects. **Run D stays recorded as
FAILED (rc=1); no row is rewritten.** This was the first execution of the labels browser specs in G020.

**(b) Root cause: the test asserted the wrong refusal shape; core and web answered the contract.** Traced from code,
not reproduced (no daemon):
1. Fixture: alice holds only `org_member` of Organization A (`infra/identity/mcp-seed.sql:72-80`, the bulk
   `org_member` INSERT; the research bindings are dave `group_member` :69, bob `group_owner` :85 and Grace
   `group_admin` :96). The spec's own header says so (:12) and E-LAB-3 already asserts her fixed 404 on
   `/groups/<research>/labels` (:382-386).
2. Web: `app/groups/[id]/labels/action/route.ts:12-13` → `groupLabelActionRoute` → `actionRoute`
   (`lib/labels-client.ts:467-509`) → `assignLabel` → core `POST /internal/v1/groups/{id}/labels/assign`; a refusal is
   rendered by `labelsRefusal` (:432-454): `forbidden` → 403 "Not permitted" / "Your role does not permit this
   change." / "Nothing was changed."; `not_found` → 404 "Not found" / "That item does not exist or is not available
   to you. Nothing was changed.". Nothing in web decides.
3. Core HTTP: `core/internal/labels/api/http.go:637-647 assignLabel` → `acc.AssignLabel`; `failure()` :274-294 maps
   `access.CodeNotFound` → `not_found` (404, :126-129).
4. Facade: `core/internal/tenancy/access/labels.go:297-357 labelAssignment` reads the definition first (:311
   `tenancy.LabelDefinitionTx`; :292 "the definition is read first (invisible: not_found)") inside the subject's
   RLS context (`facade.go:50-85 guardedTx` → `inContextTx` with user/organization/group). The
   `label_definitions_select` policy (`core/migrations/00020_labels.sql:340-343`) admits an Organization-owned row
   only under the Organization-manage predicate, which alice does not satisfy → zero rows → `CodeLabelNotFound`
   (`tenancy/labels.go:262-268`) → `CodeNotFound` (`labels_test.go:114`). The capability decision
   (`labelGroupTarget` :388-404, `label.organization.manage` then `group.read`) is never reached; had the
   definition been readable, `denial()` (`facade.go:93-102`) with `readCapability` = `organization.read`
   (`policy.go:140-153`) would give `forbidden` — that is bob's case in step 2 of the test, not alice's.
5. Contract: Plan r2 §4.2 (`plans/G020-generic-labels-visibility-plan.md:501-502`): "`not_found` for an
   unknown/foreign/outside-context/unreadable definition, Group or target (the same-facts read rule); `forbidden`
   for a readable one whose capability denies"; `docs/foundation/AUTHORIZATION.md:374-375` ("an unknown, foreign or
   unreadable definition is `not_found`"), :352-353 ("a Group role never reads them"), :657-658 and :664 (the
   general `not_found`/`forbidden` rule); Spec AC3 (`specs/...-spec.md:131`): Organization-private labels absent
   from member "errors". The Go rows encode exactly this: `labels_integration_test.go:509-519` (bob, dave, dana,
   mia: assign an Organization label → `CodeNotFound`; create one → `CodeForbidden`), `labels_test.go:78-80`.
   So the contractually correct, non-disclosing answer for alice is the fixed 404; **(a) applies: the test's
   expectation was wrong**, taken from the Plan's T12 shorthand (`plan.md:574`: "forged POSTs (alice assigns; bob
   creates an Organization label; frank creates a research label with OVR on) → 'Not permitted'"), which is
   inconsistent with the Plan's own §4.2 rule and with AUTHORIZATION.md. The Plan bytes are not changed (approved
   artifact); this Result records the deviation for Codex.
6. The rest of the test, checked for the same class: **step 3 (frank, Organization admin with the override on,
   creating a research label) would have failed identically.** `CreateLabelDefinition` in the Group scope
   (`access/labels.go:113-117`) guards `label.group.manage` on research; the override is never consulted for
   `label.group.*` (`AUTHORIZATION.md:341-342`, `policy.go:302-306`); the denial maps through
   `readCapability(label.group.manage)` = `label.group.read` (`policy.go:142-146`), which frank lacks (no research
   role) → `CodeNotFound` → 404 (`AUTHORIZATION.md:355-357`: "Organization staff without a Group role, with the
   override on or off, get no list. `readCapability` maps both Group label kinds to `label.group.read`, so such a
   caller's refusal is `not_found`"; `labels_test.go:71-76`; `labels_integration_test.go:578-597`, carol/frank with
   and without the override: "creates in research" → `CodeNotFound`). **Step 2 (bob creating an Organization
   label) is correct as written**: `label.organization.manage` denied on the readable Organization →
   `forbidden` → 403 "Not permitted" (`policy.go:140-141,153`; `labels_test.go:78-80`;
   `labels_integration_test.go:513-514`). The closing no-state-change checks (carol's Group page, the Organization
   labels page, bob's research labels page) are unaffected.

**(c) Correction (`web/tests/browser/labels.spec.ts` only, +20/-8).** alice's step asserts 404, "Not found",
"Nothing was changed." and that the page does not contain the label's name; frank's step asserts 404, "Not found",
"Nothing was changed." and that the page does not contain the submitted name; bob's step is unchanged (403, "Not
permitted"); the trailing no-state-change assertions, the override restore and the finally-cleanup are unchanged;
the title keeps the `[E-LAB-5]` id ("… refused (Not permitted, or the fixed 404 for an invisible label or Group) and
change nothing") and the header comment states the two refusal shapes with their rule. Nothing is loosened: no
assertion accepts both codes, the non-disclosure checks are added, no production code, core contract, policy,
migration or harness pin changes. Not edited (outside the minimum rerun set): the example in
`web/tests/unit/labels-forms.test.mjs:236`'s title names "Organization staff without a Group role creating a Group
label" as a "Not permitted" example; that mock-based unit test is correct for a `forbidden` answer, but the example
subject would receive `not_found` from core. Cosmetic; left for a later wording pass.

**(d) Secondary, informational: run D's `BUILD-LEDGER` `unknown 3 … ac5_status UNMET`.** `reportD/ledger-summary.tsv`
rows 1 (core: missing step build 2/6..6/6), 2 (intelligence: missing step base 2/7..7/7) and 5 (web: missing step
deps 2/4..4/4, source 1/1, build 1/1, runtime 2/5..4/5) are the three service images of the ops copy built by
`docker compose --env-file .env -f infra/compose/local.yml up -d --build --wait` (`reportD/diagnostics/ops-start.log.cmds`).
Their plain progress (`ops-start.log`: core `#10 [core build 1/6] FROM …`, `#9 [core runtime 1/2] FROM …`,
`#35 [core runtime 2/2] COPY --from=build … CACHED`, `#37 [core] exporting to image … manifest list
sha256:5610619e…` = `builds.tsv` row 1 `image_id`; web `#41 [web deps 1/4] FROM`, `#43 [web runtime 5/5] COPY …
CACHED`, `#44` export; intelligence `#20 base 1/7 FROM`, `#21 uv 1/1 FROM`, `#23 runtime 1/1 COPY src ./src CACHED`,
`#24` export) shows BuildKit 27.5.1 matching each final stage's last step against a stored cache record without
starting the dependency stages' vertices, so those steps are never printed; the ledger's row model requires every
step of every reachable stage to appear (`lib.sh` LEDGER_ROW_AWK comment and `why("missing step …")` :2872) and
marks the row `unknown`. Each export completed and the recorded image ids match. The same three Dockerfiles (same
`dockerfile_sha256`) parsed fully in the same run's identity-overlay builds (rows 11/12/15/16/17/22, `parsed`,
0 executed) and in run F (`reportF/builds.tsv`, 19 rows `parsed`, `unknown 0`), so the difference is daemon cache
state (which cache keys were already stored when the lazy lookup ran), not source. Not attributable to G020:
`git diff 3980f12a..HEAD -- infra/docker infra/compose` is empty, the `lib.sh` delta is the browser-matrix row
only, and no `run.sh` hunk touches ledger code; the class predates G020 (G009 Result:35 "INFO BUILD-LEDGER …
unknown 1 … ac5_status UNMET is an informational G051 ledger line"; G051 Result:1599). It failed no required check
(B-REQUIRED-CHECKS named only B-IDENTITY-E2E). Recommended handling: keep it as this INFO record; no harness change
in G020; the `ops` rerun may or may not reproduce it. If determinism is wanted, that is a separate G051-scope item
(e.g. accept rows whose only missing steps belong to stages consumed by a step that printed CACHED with a completed
export, or disable the lazy path for ledger builds).

**(e) Rerun scope and other selectors (from code).** `labels.spec.ts` is matched only by the Playwright projects
`labels-light`/`labels-dark` (`web/playwright.config.ts:127-128`, `testMatch /labels\.spec\.ts/`), run by
`identity_e2e` (`run.sh:6187`), called only from `variants_identity` (:5963) inside `suite_ops` (:5417); `rename`
and `off` do not run it. The matrix pins scenario ids by title prefix (`lib.sh:1897`, `E-LAB-5` kept) and requires
`stats` `^[1-9][0-9]* 0 0 0$` (:1905) — no fixed test count (169 is not pinned anywhere; `run.sh:6220` counts
rows at run time). Hosted: `infra/ci/profile-paths.tsv:89,117-121` map the labels pages to `identity:browser`,
full profile only. No Go, unit-web, migration or harness input changed. **Minimum rerun: `ops` (B-IDENTITY-E2E,
the labels-light/labels-dark projects; the whole identity environment is rebuilt by that selector).** Every other
attributable pass of this round (runs A, C, F) is retained.

**(f) Static checks run (no daemon, no toolchain).** `git diff --check` clean; changed paths
(`web/tests/browser/labels.spec.ts`, this Result) both match the 30 allow-list patterns, the binding JSON untouched;
no non-ASCII byte added to the spec (this Result keeps the document's existing typography); bracket/paren balance of
the spec unchanged by the edit. TypeScript compilation and the Playwright run are left to the daemon run
(`web-browser` image, `ops`): no local `tsc`, `node_modules` or Node 22-or-later exists here, so no type or syntax
check beyond the above was executed.

**(g) Attempt ledger and cleanup.** Still B5 attempt 1, budget not reset; implementation rounds completed: **11**
under attempt 1. This writer created no Docker object; scratch `/tmp/g020r11-*.txt` (commit message) is removed at
handback; the `g020-r11` worktree is left for the coordinator to reclaim. The run-D evidence directory was read only.

### Round 12 — run G (`controls ci-controls`) six-control failure: diagnosis and scope ruling request (writer 13), 2026-10-24

Source: branch `g020-r12` from `origin/g020-impl` `537a26956323bf11877b16c068024f342b96c449` (staging target
`3980f12ad5c8241597fe6607a3b3deadf4c77f70` verified unchanged at start); this Result is the only commit. No daemon,
database, toolchain, browser or network was used; no Docker command was issued. **No source file was edited this
round** (class (b), see (c)): the file that holds the failing pins is outside the allow-list.

**(a) Failure record (coordinator run G; evidence `<home>/.local/state/platform-evidence/g020-r10/`, read
only).** `runG.sh`: frozen checkout `.claude/worktrees/g020-frozen9` at `27097759`, `run.sh --base 3980f12a
--allow-file management/v0/results/G020-scope-allowlist.txt controls ci-controls`, started 2026-10-24T02:40:34Z, rc=1 at
03:51:03Z (`BACKGROUND INFO controls joined (exit 0, 4225 s)`). `reportG/results.tsv`: B-DEPS-CONTROLS PASS,
B-WRAPPER-CONTROLS PASS, B-DOCKER-ONLY PASS, B-SCOPE PASS (74 changed paths, all within the allow-list),
`B-CI-CONTROLS FAIL CI controls (exit 1, 523 passing)`, `B-SECRET-CANARY UNMET no secrets were generated`,
`B-REQUIRED-CHECKS FAIL … B-CI-CONTROLS(1,FAIL) B-SECRET-CANARY(1,UNMET)`. `reportG/ci-controls.log`: `529 selected;
529 defined`, 523 PASS, exactly six FAIL; first failure line 36. The six, verbatim (line: message):

| Log line | Control | Message |
|---|---|---|
| 36 | VA7 | `summary.json required_count is not 93;` |
| 41 | VU1 | `rows fast/gate/union 35/96/127, want 35/93/124;` |
| 50 | VU10 | `rows fast/gate/union 35/96/127, want 35/93/124;` |
| 143 | J8 | `the remainder result has 96 check rows, want 93;` |
| 483 | R3a | `the PASS line lacks the eligibility and the recomputed coverage: [release-verify: PASS qualified 0.1.0-rc.1: eligible (owner promotion instruction still required): MR !7 pipeline 1001 (jobs 4504, 4505) project <gitlab-project-id>; commit a85fd1b2… tree af07a012…; full fast 35 + remainder 96 -> 127 required IDs incl. ci-controls, areas none];` |
| 520 | R8a | `the PASS line lacks the evidence-only statement or the recomputed coverage: [release-verify: PASS merged 0.1.0-rc.1: merged evidence (verification only; no merge mode, actor or authority inferred): … full fast 35 + remainder 96 -> 127 required IDs incl. ci-controls, areas none (recorded)];` |

Round 11 changed only `web/tests/browser/labels.spec.ts` and this Result after `27097759`; neither is read by the
`controls` or `ci-controls` suite, so run G's evidence applies unchanged to `537a2695`. **Run G stays FAILED.**

**(b) Root cause (from code, one cause for all six).** Each of the six controls hard-codes the pre-G020 required-ID
counts of the full pipeline in `infra/ci/validate-controls.sh`, while the values they are compared with are recomputed
at run time from the live registry `infra/validation/required-checks.tsv` and the profile lists of `infra/ci/lib.sh`:

- *Registry arithmetic.* G020 S4 (`7d0b216c`) added `B-LABELS-CATALOG`, `B-LABELS-IT` and `B-LABELS-ORDERING` (suite
  `migrations`: 46 → 49 rows; 130 → 133 rows; header already says `full 126`). Fast partition = lint 9 + lint-base 4 +
  lint-web 5 + lint-core 1 + provenance 5 + contracts 6 + unit 1 + final 4 = **35** (unchanged: `migrations` is not in
  `CI_SUITES_fast`, `lib.sh:28`). Gate remainder = `CI_SUITES_full` minus fast (`lib.sh:30`: controls 2 + ci-controls 1
  + root 17 + migrations 49 + ops 14 + rename 5 + off 4) + final 4 = **96** (was 93). Union = 35 + 96 − the 4 final IDs
  present in both = **127** (was 124). These are exactly the observed values; the additions are the ones the Plan
  required (`required-checks.tsv` +3 rows, "full 126"; re-derived in rounds 3/4, Result lines 383 and 424) and RC10's
  124 on `main` `181733f2` is the pre-G020 value, consistent with this.
- **VA7** `validate-controls.sh:1813` `sum_has required_count 93` (and the `:1816` detail "93 required IDs"):
  `validate.sh` writes `required_count` from the executed suites plus final of the real registry
  (`validate.sh:1841-1883`) → 96.
- **VU1** `:1935` `vu_split full 35 93 124` and **VU10** `:1987` `vu_split rc 35 93 124` (`:1993` detail): `vu_split`
  (`:1920-1934`) counts the PASS rows of the fixture publications that `validate.sh` produced with the stub harness from
  the live registry (`vu_ids` `:1916-1919`) and compares them with its FAST/GATE/UNION arguments → `35/96/127`.
  `pipeline-union.sh` itself passed in both (the message is the count assertion, `:1928`, not a union refusal).
- **J8** `:3463` `(( n != 93 ))`: `jexpect gate` counts the published `results.tsv` check rows of the remainder → 96.
- **R3a** `:6802` and **R8a** `:7018`: `release-verify.sh` recomputes the coverage from the candidate's own registry and
  `lib.sh` (`release-verify.sh:426-435`, `Q_COV="full fast $(count_of fs) + remainder $(count_of rs) -> $(count_of us)
  required IDs incl. ci-controls, …"`) and printed `fast 35 + remainder 96 -> 127`; both controls match the PASS line
  against the literal substring `fast 35 + remainder 93 -> 124 required IDs incl. ci-controls`. The quoted PASS lines
  show the eligibility / evidence-only statements present and the receipt checks passed, so this is **not a
  release-verify defect**: only the pinned substring differs (`:6803`'s detail also says "35 + 93 -> 124").
- The literals `93`/`124` occur in `validate-controls.sh` only in these nine lines (`:1813`, `:1816`, `:1935`, `:1987`,
  `:1993`, `:3463`, `:6802`, `:6803`, `:7018`); every other `124` is the deadline exit code (`:669`, `:678`, `:2333`,
  `:3937`, `:3940`). No other control reads those lines; `infra/ci/lib.sh` holds no count; no digest of
  `validate-controls.sh` is pinned anywhere (`run-approved-controls.sh`, `unchanged.allow`, `.gitlab-ci.yml`,
  `required-checks.tsv` do not name it).
- *Precedent.* The same six controls failed for the same reason when G064 added `B-IT-ISOLATION` and were moved
  92 → 93 / 123 → 124 in `6dce6e2f` ("G064 S6: validate-controls.sh count pins for the B-IT-ISOLATION required row",
  8 lines, "mechanical pin parity only"); G062 did the same in `33306fec`. Both Goals' allow-lists carried
  `^infra/ci/validate-controls[.]sh$`.

**(c) Correction class: (b) — the pins live outside the allow-list; ruling requested.** `infra/ci/validate-controls.sh`
matches none of the 30 patterns of `management/v0/results/G020-scope-allowlist.txt`. Plan r2 §11 excluded
"`infra/ci/**` other than `profile-paths.tsv`" on purpose, and round 6 (Result, "Reconciliation onto staging 9761b89d")
recorded the file as outside the allow-list with G064 owning its count pins. The Plan did not foresee that its own
`required-checks.tsv` +3 rows move these six count pins (the G062/G064 mechanical consequence). The pins must move
**to** the registry-derived values 96/127; weakening the controls (dropping the count assertions) is not proposed.

Minimal amendment proposal for Codex, following the `ed0dd821` / `ae55426b` precedent (dated comment appended; header
and existing lines intact; allow-list SHA-256 before `34e0b05eb4b04b327ad02b5ddc142c115573cf2f63e497576a383e1cb37dceac`):

```
# Amendment 2026-10-24 (Codex ruling, <handoff>, narrow ci-controls count-pin amendment): validate-controls.sh count pins for the three B-LABELS-* required rows only (Result round 12 (c)); 31 patterns.
^infra/ci/validate-controls[.]sh$
```

authorizing exactly this edit of `infra/ci/validate-controls.sh` (9 lines, 11 substitutions, no control logic, no new
or removed assertion; prepared and checked on a scratch copy only, see (f); file SHA-256 before
`6802f8498bcce2638d948bccf48838cef9c0b6a04fc4d5a764e91dfa7b50e796`, after
`c9a4d849bb245ccdb7d32afd59a4942f41ecda5d1d1d61c110f344db4d248a62`):

```
1813s/required_count 93/required_count 96/
1816s/93 required IDs/96 required IDs/
1935s/vu_split full 35 93 124/vu_split full 35 96 127/
1935s/gate 93 -> 124 distinct/gate 96 -> 127 distinct/
1987s/vu_split rc 35 93 124/vu_split rc 35 96 127/
1993s/gate 93 -> 124 distinct/gate 96 -> 127 distinct/
3463s/n != 93/n != 96/
3463s/want 93/want 96/
6802s/remainder 93 -> 124/remainder 96 -> 127/
6803s/35 + 93 -> 124/35 + 96 -> 127/
7018s/remainder 93 -> 124/remainder 96 -> 127/
```

Every assertion is preserved: VA7 still requires `required_count` to equal the remainder count (now 96), VU1/VU10 still
require the exact fast/gate/union partition and the exact 4-ID overlap (now 35/96/127), J8 still requires the exact
remainder row count (96), R3a/R8a still require the eligibility / evidence-only statement **and** the recomputed
coverage (now "35 + remainder 96 -> 127"); the detail strings move with them. `B-SCOPE` on the current tree stays PASS
because nothing outside the allow-list was touched. The bound input binding
`management/v0/results/G020-implementation-input-binding-20261021.json` is untouched
(`90055f12fcd44ba0afc31bcac09dc865606c25a93a1c5db53b4e76a20bc040cf`).

**(d) `B-SECRET-CANARY UNMET` in run G: a selection artifact, no change.** `run.sh:6668` records UNMET when the run's
secrets file is empty. Its writers are `make_env_file` (`:314`, every stack-starting suite), the lint secret-scan
canaries (`ss_canary` `:500`, `static_secret_source` `:620`, both under `suite_lint`), `new_canary`/root (`:2020`,
`:2168`, `:2483`), migrations (`:3732`) and ops (`:5261`); none is reachable from `suite_controls` (`:389`:
`static_deps_controls`, `run-approved-controls.sh`) or `suite_ci_controls` (`:396`: `static_ci_controls`). Round-8 run B
(`lint lint-base lint-core unit-core`, same class of source) passed it; any final selection that includes `lint` (with
`lint-base`, per the round-8 note) produces the PASS row.

**(e) Minimal attributable rerun (after the amendment is ruled and the edit applied).**

1. Focused proof of the six, direct script run from the frozen checkout root (no `run.sh`; `--only` is a comma list,
   `validate-controls.sh:82`, `:205-215`; the `vu_make full`/`rc` fixtures are built on demand):
   `bash infra/ci/validate-controls.sh --only VA7,VU1,VU10,J8,R3a,R8a --log "$EVIDENCE/ci-controls-summary.log"
   >"$EVIDENCE/ci-controls.log" 2>&1` (optionally `CI_CONTROLS_OWNED_REG=<this run's registry>` as `run.sh:1544`
   passes). Expected last line: `ci-controls: 6 passed, 0 failed (of 6 selected; 529 defined)`. Cost from run G's
   `ci-controls-timings.tsv`: VA7 9 s, VU1 19 s, VU10 19 s, J8 13 s, R3a 7 s, R8a 7 s = 74 s plus the startup guard and
   template (≈ 2–5 min). The script itself runs `docker-desktop-mounts.sh` first.
2. `lint lint-base` through `run.sh` (B-SHELL shellchecks `infra/ci/validate-controls.sh`, `run.sh:332-348`;
   `lint-base` alongside for the canary pairing): the only suite that reads the changed file (≈ 7 min; round-8 run B
   timing).
3. Not affected, no rerun: `controls` (B-DEPS-CONTROLS and B-WRAPPER-CONTROLS read `web/package*.json` and
   `run-approved-controls.sh`, not `validate-controls.sh`), every Go, web, `migrations`, `ops`, `rename`, `off`
   selector; no pin names the file.
4. **The `B-CI-CONTROLS` required row is all-or-nothing through `run.sh`:** `static_ci_controls` (`run.sh:1545-1553`)
   runs the script without `--only` and requires the PASS set to equal `--list` (≈ 70 min, run G: 4225 s). Run G's 523
   PASS are attributable to every control other than the six (none of the other 523 reads the nine changed lines),
   so run G (523) + step 1 (6) cover all 529 controls on the corrected source; the full `ci-controls` suite run is
   still needed once for the required row and is in any case part of the post-merge merged-SHA affected qualification
   (`run.sh`, `lib.sh` and `required-checks.tsv` changed). Whether to run it pre-merge as well is the coordinator's
   scheduling decision; run G remains recorded as FAILED either way.

**(f) Static checks run (no daemon, no toolchain).** `git diff --check` clean; `git diff --name-only 3980f12a HEAD` =
74 paths (73 source/docs + this Result), every one matched by a non-comment allow-list pattern, the negative control
`infra/ci/validate-controls.sh` not matched; the binding JSON untouched. The proposed edit was applied with the eleven
`sed` expressions above to a scratch copy under `/tmp/g020r12-scratch/` (never to the repository): `bash -n` exit 0,
`diff -u` = exactly the 9 changed lines, and the residual literal scan left only the `:3940` deadline detail
(`harness 124`). No shellcheck (pinned image; daemon) and no control was executed.

**(g) Attempt ledger and cleanup.** Still B5 attempt 1, budget not reset; implementation rounds completed: **12** under
attempt 1. This writer created no Docker object; scratch `/tmp/g020r12-scratch/` and `/tmp/g020r12-msg/` are removed at
handback; the `g020-r12` worktree (`.claude/worktrees/agent-<id>`) is left for the coordinator to reclaim.
The run-G evidence directory was read only.

### Round 13 — Codex-approved count-pin amendment applied (writer 14), 2026-10-24

Source: branch `g020-r13` from `origin/g020-impl` `d3902c92fe64a59808f2452be6b5d1fe5adacbe5` (staging target
`3980f12ad5c8241597fe6607a3b3deadf4c77f70` verified at start). No daemon, database, toolchain, browser or Docker
command was used.

**(a) Codex ruling (relayed by the coordinator; a scope/validation ruling only, not an implementation-acceptance or
merge authorization).** APPROVED, narrowly: add `^infra/ci/validate-controls[.]sh$` as the 31st allow-list pattern,
authorizing ONLY the eleven stated literal substitutions across the nine count-pin lines of
`infra/ci/validate-controls.sh` (remainder 93 -> 96, full union 124 -> 127), with no assertion or control-logic change.
Qualification timing: before the MR, the six focused controls (VA7, VU1, VU10, J8, R3a, R8a) plus `lint lint-base`;
the full `B-CI-CONTROLS` suite runs after the authorized merge on the exact merged staging SHA. Run G stays recorded
as FAILED; the focused proof does not rewrite it. Nothing else is authorized.

**(b) Allow-list.** The round 12 (c) amendment comment and pattern appended (header and existing lines intact; the
`<handoff>` placeholder recorded as the coordinator relay): **31** patterns. SHA-256 before
`34e0b05eb4b04b327ad02b5ddc142c115573cf2f63e497576a383e1cb37dceac`, after
`92b71190897c069ecc6a9c7203d80bf3006f16ed0a50712acd8e988401c5bb04`. The binding
`management/v0/results/G020-implementation-input-binding-20261021.json` is untouched
(`90055f12fcd44ba0afc31bcac09dc865606c25a93a1c5db53b4e76a20bc040cf`).

**(c) Edit.** The eleven round 12 (c) `sed` expressions applied verbatim to `infra/ci/validate-controls.sh` (+9/-9):
lines 1813, 1816, 1935, 1987, 1993, 3463, 6802, 6803, 7018; each changed line differs from the original only in the
93 -> 96 / 124 -> 127 literals. SHA-256 before `6802f8498bcce2638d948bccf48838cef9c0b6a04fc4d5a764e91dfa7b50e796`,
after `c9a4d849bb245ccdb7d32afd59a4942f41ecda5d1d1d61c110f344db4d248a62` (equals the round 12 prepared value).
`bash -n` exit 0. Residual `93`/`124` literals: none in the six controls; only the deadline exit-code lines 669, 678,
2333, 3937, 3940 (unchanged, not count pins).

**(d) Focused rerun the coordinator will run (from round 12 (e)), from the frozen checkout root:**
1. `bash infra/ci/validate-controls.sh --only VA7,VU1,VU10,J8,R3a,R8a --log "$EVIDENCE/ci-controls-summary.log"
   >"$EVIDENCE/ci-controls.log" 2>&1` — expected last line `ci-controls: 6 passed, 0 failed (of 6 selected; 529
   defined)`.
2. `infra/validation/run.sh --base <staging SHA> --allow-file management/v0/results/G020-scope-allowlist.txt lint
   lint-base` (B-SHELL shellchecks the changed file; `lint-base` for the secret-canary pairing).
3. After the authorized merge: full `ci-controls` on the exact merged staging SHA (the `B-CI-CONTROLS` required row).

**(e) Static checks.** `git diff --check` clean; `git diff --name-only 3980f12a HEAD` matched against the 31
non-comment patterns: zero paths outside. Attempt ledger: still B5 attempt 1; implementation rounds completed: **13**.
Cleanup: no Docker object; scratch `/tmp/g020r13-msg/` removed at handback; the worktree is left for the coordinator.

#### Staging reconciliation (4b333da6)

Staging advanced to `4b333da64790c9feebf36af37f7c6ba83a8df821` (`3980f12a` verified as its ancestor). The delta
`3980f12a..4b333da6` is exactly `management/v0/PROGRESS.md` (Codex-owned; records the Codex Round 12 ruling), SHA-256
at staging `380f5b3f7daa29982a7926ecc12681c10773d39b72b0b0a0329ece29f14169e0`; the branch had not changed that file.
Merged `--no-ff` into `g020-r13` as `9d83701f570ebf2a28aac3252a96b846dd6a3365` without conflict. No source, test or
harness input changed, so no rerun beyond the Codex-ruled focused controls in (d) is needed; the scope base for the
focused `lint lint-base` run becomes `4b333da6`.

### Round 14 - consolidated pre-MR local qualification record (final source 02e59439)

Writer 15 (docs only; branch `g020-r14` from `origin/g020-impl` `02e59439964e48eb83ca4a7c23bb8a121b82e821`, tree
`b50b01a5`; staging `4b333da64790c9feebf36af37f7c6ba83a8df821` verified). No daemon or toolchain was used; every number
below is read from `<home>/.local/state/platform-evidence/<dir>/` (`run.txt`, `DONE-*`, `run*.sh`,
`report*/results.tsv`, consoles). All runs: coordinator, real daemon (docker 27.5.1), plain `infra/validation/run.sh
--base <staging> --allow-file management/v0/results/G020-scope-allowlist.txt` from a frozen checkout (except J, a direct
`validate-controls.sh` run). Scope base: `bf94dc6d` (r6-r8), `3980f12a` (r9-r11), `4b333da6` (r13). Counts are
results.tsv rows; "IDs" is the B-REQUIRED-CHECKS count. Failed runs stay FAILED; no row is rewritten.

| Dir / run | Source | Selectors | rc | Rows | Outcome |
|---|---|---|---|---|---|
| g020-r6 B | `aaeb1530` | `unit-web lint-web` | 0 | 10 PASS, 4 INFO; 9 IDs | PASS |
| g020-r6 A | `aaeb1530` | `migrations` | 1 | 48 FAIL, 4 PASS, 1 UNMET, 4 INFO | **FAILED** (00020 apply error, cascade; Round 7) |
| g020-r7 B | `5e913484` | `lint lint-base lint-core unit-core` | 0 | 19 PASS, 5 INFO; 18 IDs | PASS |
| g020-r7 A | `5e913484` | `migrations` | 1 | 45 PASS, 7 FAIL (6 checks + derived), 1 UNMET (B-IT-ISOLATION T-9/T-10), 4 INFO | **FAILED** (7 IDs not PASS; Round 8) |
| g020-r8 B | `bfcca2a3` | `lint lint-base lint-core unit-core` | 0 | 19 PASS, 5 INFO; 18 IDs | PASS |
| g020-r8 A | `bfcca2a3` | `migrations` (owned registry) | 1 | 50 PASS, 3 FAIL, 5 INFO | **FAILED**: B-IDENTITY-CATALOG, B-LABELS-IT (+ derived); pins fixed in `934e3e95` / Round 10 |
| g020-r9 A | `27097759` | `migrations` (owned registry + durable marker) | 0 | 53 PASS, 5 INFO; 52 IDs | PASS |
| g020-r9 B | `27097759` | `lint lint-base lint-core unit-core` | 0 | 19 PASS, 5 INFO; 18 IDs | PASS |
| g020-r9 C | `27097759` | `root` | 0 | 21 PASS, 4 INFO; 20 IDs | PASS |
| g020-r10 F | `27097759` | `rename off` | 0 | 13 PASS, 4 INFO; 12 IDs | PASS |
| g020-r10 D | `27097759` | `ops` | 1 | 16 PASS, 2 FAIL, 4 INFO | **FAILED**: B-IDENTITY-E2E (+ derived) |
| g020-r10 G | `27097759` | `controls ci-controls` | 1 | 4 PASS, 2 FAIL, 1 UNMET, 5 INFO | **FAILED**: B-CI-CONTROLS (+ derived) |
| g020-r11 H | `537a2695` | `ops` | 0 | 18 PASS, 4 INFO; 17 IDs | PASS |
| g020-r13 J | `02e59439` | `validate-controls.sh --only VA7,VU1,VU10,J8,R3a,R8a` | 0 | `6 passed, 0 failed (of 6 selected; 529 defined)` | PASS (focused) |
| g020-r13 K | `02e59439` | `lint lint-base` | 0 | 17 PASS, 5 INFO; 16 IDs | PASS |
| g020-r13 L | `02e59439` | `lint-web` | 1 | 7 PASS, 1 FAIL (derived), 1 UNMET, 4 INFO | **not a clean PASS** (see below) |
| g020-r13 M | `02e59439` | `unit-web` | 0 | 5 PASS, 4 INFO; 4 IDs | PASS |

Key rows:
- **g020-r9 A** (01:52-02:18Z): B-IDENTITY-CATALOG PASS (version 20, the 14 reviewed identity objects); B-LABELS-IT PASS
  (15/15 subtests covering LAB-1..17); B-LABELS-CATALOG 4/4 and B-LABELS-ORDERING 4/4 PASS; B-DOMAINS-CATALOG 4/4 PASS;
  B-IT-ISOLATION PASS with T-9 (runner residue) and T-10 (runner fault) executed; B-SECRET-CANARY PASS (39 secrets);
  B-SCOPE PASS (74 paths since `3980f12a`); B-REQUIRED-CHECKS PASS (52 IDs). **B** and **C**: B-SCOPE 74 paths,
  B-REQUIRED-CHECKS 18 and 20 IDs; BUILD-LEDGER unknown 0 in A, B and C.
- **g020-r10 D** (02:18-02:38Z): `consoleD.txt` lines 31-32 `Expected: 403` / `Received: 404` at
  `labels.spec.ts:452` (line 41); browser stats `169 0 0 2` (line 57): 169 passed, the two unexpected are `[E-LAB-5]`
  in `labels-light` and `labels-dark` (`reportD/diagnostics/idn-browser.log`: `2 failed`, `169 passed`). Its
  `BUILD-LEDGER INFO expected 25, rows 25, unknown 3 ... ac5_status UNMET` is BuildKit cache-elided dependency
  stages, not G020-caused (Round 11 (d)); the rerun H shows `unknown 0 ... MET`.
- **g020-r10 G** (02:40-03:51Z; controls joined after 4225 s): B-DEPS-CONTROLS and B-WRAPPER-CONTROLS PASS;
  `ci-controls.log` `523 passed, 6 failed (of 529 selected; 529 defined)`: VA7, VU1, VU10, J8, R3a, R8a (Round 12);
  B-SECRET-CANARY UNMET only because the selection had no `lint`; evidence `reportG`, `ci-controls.log`.
- **g020-r11 H** (02:52-03:13Z): B-IDENTITY-E2E PASS, "171 tests in identity-light ... labels-light and labels-dark
  ... 0 skipped/flaky/unexpected" (E-LAB-1..6 in both labels projects); BUILD-LEDGER `unknown 0 ... ac5_status MET`;
  B-SECRET-CANARY PASS (349 secrets); B-SCOPE 74 paths.
- **g020-r13 J/K** (04:06-04:10Z): J's six PASS lines show the recomputed `96` / `35 + 96 -> 127` values. K: B-SHELL
  PASS (shellchecks `infra/ci`), B-SECRET-CANARY PASS (19 secrets), B-SCOPE PASS (75 paths since `4b333da6`),
  B-REQUIRED-CHECKS PASS (16 IDs).
- **g020-r13 L** (rc=1): the five B-WEB-* rows, B-DOCKER-ONLY and B-SCOPE PASS; `B-SECRET-CANARY UNMET no secrets
  were generated` and therefore `B-REQUIRED-CHECKS FAIL ... B-SECRET-CANARY(1,UNMET)`. A selection artifact, but L is
  **not** a clean wrapper PASS. For the required rows it is superseded by M and by K's canary PASS on the same tree.
- **g020-r13 M** (04:12-04:13Z): B-UNIT-WEB PASS (web-check = eslint over `web/`, `tsc --noEmit` over `**/*.ts`
  including the edited `tests/browser/labels.spec.ts`, `test:unit`; plus secret-scan); B-SECRET-CANARY PASS
  (8 secrets); B-SCOPE 75 paths; B-REQUIRED-CHECKS PASS (4 IDs).

**(a) Coverage carried to `02e59439`.** `git diff --name-only 27097759 02e59439` = `infra/ci/validate-controls.sh`,
`management/v0/PROGRESS.md`, `management/v0/results/G020-generic-labels-visibility-result.md`,
`management/v0/results/G020-scope-allowlist.txt`, `web/tests/browser/labels.spec.ts`. Round 11 (`27097759..537a2695`)
changed only `labels.spec.ts` (and this Result): covered by H (`ops`) and M (`unit-web`). Rounds 12-13
(`537a2695..02e59439`) changed only the nine `validate-controls.sh` count-pin lines, the allow-list, this Result and
the PROGRESS.md-only staging merge `4b333da6`: covered by J (the six controls) and K (B-SHELL, B-SCOPE, canary). No
input of g020-r9 A/B/C or g020-r10 F changed afterwards, so those passes carry over unchanged.

**(b) Not yet executed (Codex ruling, Round 13 (a)).** The full `ci-controls` selector (the B-CI-CONTROLS required row,
529 controls) runs after the authorized merge on the exact merged staging SHA. Run G remains FAILED. J is a focused
proof of the six controls; it does not rewrite G and is not a B-CI-CONTROLS PASS. B-CI-CONTROLS has not passed.

**(c) Resources and cleanup receipt.** Each `run.txt` records the pre-run inventory 15 containers / 11 volumes /
6 networks / 52 images (r6-r13, also before G). The return to that baseline after each run, and the only transient
excursion (images 56 during run G), are coordinator observations not stored in an evidence file. Leaked fixed-name
volume `project-archetype-render-cache`: created 2026-10-24T02:20:49Z by `skills/saas-foundation/generator/render.sh`
during run F, unlabeled (`Labels: null`), no container used it; removed by the coordinator after D and F ended (receipt
`g020-r10/leaked-volume-receipt.txt`, `leaked-volume-inspect.json`). Durable local markers: the r9 marker
(`g003-pending/local-89316f3384bc.tsv`) was removed after run A; J's self-registered marker is also gone; verified now:
`platform-evidence/g003-local/` is empty (no `g003-pending/`). Finished writer worktrees and old frozen checkouts
(`agent-<id>`, `agent-<id>`, `agent-<id>`, `g020-frozen9`, `g020-frozen11`) are
reclaimed (absent from `git worktree list`); their evidence is kept. `g020-frozen13` (`02e59439`) is retained as the
frozen proof checkout until the MR merges; release condition: after the merged-SHA run starts. The held `g064-*` and
`worktree-agent-*` branches are untouched (not approved for deletion).

**(d) Budget.** Still B5 attempt 1, budget not reset. Rounds 11-13 were mechanical/test-contract corrections under
standing authority and Codex rulings (recorded above). This round adds only this Result section; scratch
`/tmp/g020r14-msg/` is removed at handback and the worktree is left for the coordinator.

### Round 15 — Codex source-review finding (P2): web stored-name predicate vs core's trim set (writer 16), 2026-10-24

Source: branch `g020-r15` from `origin/g020-impl` `539c0edf84ccf4733523c95a78768704a86acccd` (Codex source review of
`539c0edf`: CORRECTION REQUIRED). Stack-free: no daemon, database, Go or web toolchain; Node v20.19.0 on the host was
used only for a standalone proof of the predicate (no `node_modules`, no TypeScript stripping, so the unit files
themselves were not executed here).

**(a) Finding (Codex, P2).** `tenancy.ValidateLabelName` delegates to Go `strings.TrimSpace`, which does not trim U+FEFF,
so core (and `label_definitions_name_valid`) accept and store a name with a boundary U+FEFF; the web stored-name
predicate `isLabelName` used JavaScript `trim()`, which does trim U+FEFF, so the create/rename answer and every later
list answer carrying that stored name were read as `malformed`: the stored label/page became unusable (the unknown-outcome
page on the mutation, a malformed list afterwards). Required: align the web predicate with core's accepted grammar, with
regressions in the existing allow-listed client and core unit tests; the server contract stays unchanged.

**(b) Root cause and the exact code-point difference.** Core: `core/internal/tenancy/names.go:19` (`strings.TrimSpace`),
then `:21-26` (C0, DEL, C1 refused in the trimmed value), `labels.go:117-131` (U+2028/U+2029 refused anywhere before
trimming, 1-80 code points after), `core/migrations/00020_labels.sql:160-161` (`btrim`, 1-80 `char_length`, the same
class). Web: `web/lib/labels-client.ts:102` (old) `value.trim() !== value`, reached from `label()` (`:177`, create/rename
`row()` and `listLabels`) and `groupLabel()` (`:185`, `listGroupLabels`). Go `unicode.IsSpace` is exactly Unicode
White_Space: U+0009-000D, U+0020, U+0085, U+00A0, U+1680, U+2000-200A, U+2028, U+2029, U+202F, U+205F, U+3000 (25).
JavaScript `trim()` is WhiteSpace (TAB, VT, FF, SP, U+FEFF and the 17 Zs: U+0020, U+00A0, U+1680, U+2000-200A, U+202F,
U+205F, U+3000) plus LineTerminator (LF, CR, U+2028, U+2029): also 25. Exhaustive comparison over the whole code space
(Node 20): the sets differ in exactly **U+FEFF (JavaScript only) and U+0085 (Go only)**; nothing else (U+180E is Cf in
both). U+0085 is unreachable in a stored name: core trims it at a boundary and refuses it inside (control); the web
`NAME_FORBIDDEN` class (`\p{Cc}`) refuses it anywhere, so the only observable mismatch was U+FEFF. The input side has no
mismatch: web never validates or trims a submitted name (`readLabelAction`, `session.ts:309` `field`, the form has no
`maxLength`); core trims and decides, so the earlier "stricter input validator" ruling is not exercised.

**(c) Correction (web only; +6/-1 in `web/lib/labels-client.ts`).** `NAME_EDGE_SPACE = /^\p{White_Space}|\p{White_Space}$/u`
(`:48`) replaces `value.trim() !== value` in `isLabelName` (`:107`): by definition the same set `strings.TrimSpace`
removes, so the predicate accepts exactly what core can store (a boundary U+FEFF, counted among the 80 code points by
`Array.from`, as `utf8.RuneCountInString` counts it) and still refuses every name core never stores (a boundary
White_Space code point, any control/separator, empty, over 80 code points). Same approach as the accepted G063 mirror in
`web/lib/mcp-client.ts:244-258` ("JavaScript's trim() is not used (it also trims U+FEFF)"). Nothing changes in core, the
migration, the DTO key sets, byte bounds, the CopyIsExact-pinned helpers, the refusal reading, escaping (names remain
rendered as escaped text, never logged or put in a URL) or any authorization path: core still decides every read/write.

**(d) Regression tests (existing files extended; no new suite).** `web/tests/unit/labels-client.test.mjs:268` "accepts a
stored name with a boundary U+FEFF on the create, rename and both list paths; refuses a boundary White_Space": create
(`createLabel`), rename (`updateLabel`), `listLabels` and `listGroupLabels` accept `U+FEFF Priority U+FEFF`, a lone U+FEFF
and an 80-code-point U+FEFF-padded name; the 17 Zs code points leading and trailing (34 rows), U+0085 at either boundary
and an 81-code-point name are `malformed` on both list paths; a space-padded create answer and a U+3000-suffixed rename
answer are `malformed`. `web/tests/unit/labels-forms.test.mjs:157` "answers the closed status when core's stored name
begins or ends with U+FEFF": the Organization route answers 303 `?status=created` / `?status=renamed` (not the
unknown-outcome page) and sends the name as submitted. `core/internal/tenancy/labels_test.go:63` `TestLabelNameTrimSet`
pins current core behavior unchanged: `unicode.IsSpace` holds for exactly the 25 listed code points; each at both
boundaries trims to `x` (U+2028/U+2029 refused before trimming); boundary U+FEFF names (incl. 80 code points) are returned
unchanged and an 81-code-point U+FEFF name is `invalid_name`; `x U+0085` trims to `x`. On the old predicate the three
U+FEFF acceptance assertions fail (`trim()` strips the BOM, so `label()`/`groupLabel()` returned `malformed`, proven by the
standalone Node evaluation: old `false`, new `true`; every refused vector `false` under both).

**(e) Pins, reruns, static checks.** No harness pin reads these files or their test counts (`unit_service` in
`infra/validation/run.sh:445-478` runs the README `web-check`/`core-check` lines without a count; `validate-controls.sh`,
`required-checks.tsv` and `profile-paths.tsv` name none of them): no pin moved. Nothing else reads `isLabelName`
(`labels.spec.ts` drives the UI only; no label name with U+FEFF is seeded). Minimum reruns on the final source:
`infra/validation/run.sh --base <staging SHA> --allow-file management/v0/results/G020-scope-allowlist.txt unit-web` (eslint,
`tsc --noEmit`, `test:unit`) and `... unit-core` (gofmt, vet, `go test` incl. `TestLabelNameTrimSet`); `lint lint-base`
only if the coordinator wants the canary pairing. `git diff --check` clean; `git diff --name-only 3980f12a HEAD` matched
against the 31 patterns: only `management/v0/PROGRESS.md` outside (staging's own). Allow-list unchanged
(`92b71190...c5bb04`), binding untouched (`90055f12...c040cf`). Attempt ledger: still B5 attempt 1; rounds completed:
**15**. All prior failed outcomes (runs A r6-r8, D, G, L) stand as recorded. Cleanup: no Docker object; scratch
`/tmp/g020r15-msg/` removed at handback; the worktree is left for the coordinator.

#### Staging reconciliation (4c63cd5d)

Staging advanced to `4c63cd5de796d108d091f1399025dd72b7a10215` (`4b333da6` verified as its ancestor). The delta
`4b333da6..4c63cd5d` is exactly `management/v0/PROGRESS.md` (Codex-owned; records the Codex source-review disposition of
`539c0edf`: CORRECTION REQUIRED, the boundary-U+FEFF core/web mismatch corrected above), SHA-256 at staging
`1547bf451254e3a1487dce90835fa073cefa40d66010f17ba4ed57d13b8049d5`; the branch had not changed that file beyond
staging's own content. Merged `--no-ff` into `g020-r15` after the round 15 commits, without conflict. No source, test or
harness input changed, so the minimum reruns in (e) are unchanged; the scope base becomes `4c63cd5d`.

#### Round 15 evidence correction

The Round 15 text above ("the unit files themselves were not executed here") refers only to the writer's stack-free
worktree. On the exact frozen source `ae96e75bd91b9483d9ecbff86080959b347d6b11` (checkout `g020-frozen15`, scope base
`4c63cd5d`, allow-list of 31 patterns, SHA-256 `92b71190897c069ecc6a9c7203d80bf3006f16ed0a50712acd8e988401c5bb04`) the
coordinator ran both focused selectors, and both PASSED.

- **Prior runs N/O** (`<home>/.local/state/platform-evidence/g020-r15/`): N = `unit-web` rc=0
  (04:32:47Z-04:33:48Z): B-UNIT-WEB, B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE (75 changed paths since `4c63cd5d`) and
  B-REQUIRED-CHECKS PASS. O = `unit-core` rc=0 (ended 04:34:12Z): the same five checks (B-UNIT-CORE in place of
  B-UNIT-WEB) PASS. N and O were launched concurrently, which violates the sequential image-build rule: both wrappers
  started at 04:32:47Z, their build ledgers showed one and two executed steps respectively, and `infra/validation/run.sh`
  has no cross-run serialization. N's cleanup snapshot (`reportN/inventory-leftover.txt`) observed the concurrent core
  container `g001-unit-core-1063a7-core-check-run-046de8e0b4a6`; that container is absent now. N and O are preserved as
  history (not deleted or rewritten) and are **not** counted as the sequential proof.
- **Sequential rerun P/Q** (`<home>/.local/state/platform-evidence/g020-r15b/`, one wrapper script `runseq.sh`,
  strictly sequential, same frozen source, base and allow-list): P = `unit-web` rc=0 (04:39:24Z-04:39:56Z), then
  Q = `unit-core` rc=0 (04:39:56Z-04:40:38Z). Each recorded the same five checks PASS (B-UNIT-WEB / B-UNIT-CORE,
  B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE 75 paths, B-REQUIRED-CHECKS with 4 required IDs each recorded once);
  `inventory-leftover.txt` is empty for both. Docker before/after: 15 containers / 11 volumes / 6 networks / 52 images
  (baseline; no owned leftovers).
- **What the PASS covers.** The wrapper PASS covers `web-check` (eslint, `tsc --noEmit`, `test:unit`) and `core-check`
  (gofmt, vet, `go test`), which include the new regressions in `labels-client.test.mjs`, `labels-forms.test.mjs` and
  `TestLabelNameTrimSet`. The harness keeps no per-test logs, so named per-test output for those regressions is not
  separately retained.

No source code changed in this correction; no full qualification and no `ci-controls` run was made. The six focused
controls (VA7, VU1, VU10, J8, R3a, R8a; PASS on `02e59439`, round 13 run J) and the full affected `ci-controls`
selector remain required after the authorized merge on the exact merged staging SHA, as previously ruled. All earlier
failed runs (r10 D and G; r13 L, not a clean PASS) stay FAILED as recorded.

#### Staging reconciliation (d6d95dd7)

Staging advanced to `d6d95dd747b300b339d7507abaccf83e4523cf8e` (`4c63cd5d` verified as its ancestor). The delta
`4c63cd5d..d6d95dd7` is exactly `management/v0/PROGRESS.md` (Codex's one-line round 15 status), SHA-256 at staging
`0781676bb79433ea6ee500db98530229a3744ad4fdd3a3135181796f38beb8fe`. Merged `--no-ff` after the evidence-correction
commit, without conflict; `management/v0/PROGRESS.md` on the branch equals staging's byte for byte. No application,
test or harness input changed, so no rerun is needed; the tested source remains `ae96e75b`.

#### Staging reconciliation (48330587)

Staging advanced to `48330587b4dbbb525fb9d51940381131878e48b0` (`d6d95dd7` verified as its ancestor). The delta `d6d95dd7..48330587` is exactly
`management/v0/PROGRESS.md` (Codex's source-ACCEPT progress record), SHA-256 at staging `5942f0e510b8d6a9f6c4d50e536c6c150387bea145b61616d1b065e3e818aaee`. Merged `--no-ff`
without conflict; `management/v0/PROGRESS.md` on the branch equals staging's byte for byte. No application, test or
harness input changed, so no rerun is needed; the tested source remains `ae96e75b`.

### Final merge, post-merge qualification and cleanup receipt

**Merge.** MR !95 `G020: generic labels and visibility`, source `g020-impl` @
`a7b93b110cf23c000bdc9f968e59193a1346e7f3` (tested source `ae96e75b`, tree `c88d806b`; the later commits change only
the Result and PROGRESS), Codex source ACCEPT round 15. Hosted MR pipeline `2983877204` (`merge_request_event`, sha
`a7b93b11`): success. Duo was requested and replied "found nothing to comment on", but stated that its review used
truncated tool output (about 15% of the diff), so it is advisory only; Codex's full source review is the review of
record. Unresolved threads: 0. Merged with `glab mr merge 95 --auto-merge --sha a7b93b11… --remove-source-branch` as
merge commit `504fe8fc7138111ff566c1a07902bfe550d5e30c`, parents `48330587b4dbbb525fb9d51940381131878e48b0` and
`a7b93b110cf23c000bdc9f968e59193a1346e7f3`; the merged tree `c0f3882cf2ce40b0597399a13e7b950cab3b36e9` equals the source
tree, and the remote source branch was deleted. Staging later advanced to `69ff6c5a` (PROGRESS.md-only Codex status).

**Post-merge qualification on the exact merged SHA `504fe8fc`** (per Codex's round-12 ruling: full affected
`ci-controls`; no full local suite):

```
bash infra/validation/run.sh --base 48330587b4dbbb525fb9d51940381131878e48b0 \
  --allow-file management/v0/results/G020-scope-allowlist.txt \
  --report-dir …/g020-postmerge/reportPM lint controls ci-controls
```

- Window 05:02:35Z–06:12:48Z, 2026-10-24; rc=0. `lint` was added to the approved `controls ci-controls` selection
  solely so that B-SECRET-CANARY is not UNMET (the selection artifact seen in run G).
- Results: 15 required IDs plus `final`, each PASS (16 PASS rows in `results.tsv`, none FAIL/UNMET), including
  B-CI-CONTROLS (529 controls pass), B-DEPS-CONTROLS, B-WRAPPER-CONTROLS, B-SECRET-CANARY (19 generated secrets),
  B-SCOPE (75 changed paths since `48330587`, all within the allow-list) and B-REQUIRED-CHECKS (15 required IDs).
- Build ledger unknown/duplicate/exception rows: 0; inventory-leftover: empty; Docker before/after: 15 containers /
  11 volumes / 6 networks / 52 images.
- Evidence: `<home>/.local/state/platform-evidence/g020-postmerge/` (`reportPM`, `consolePM.txt`, `run.txt`,
  `runPM.sh`).
- Source binding, stated honestly: the run executed in the detached checkout `g020-merged-504fe8fc` (named in the
  retained `runPM.sh`), created with `git worktree add --detach … 504fe8fc7138111ff566c1a07902bfe550d5e30c`. The
  launcher script `g020-startPM.sh` refused to start unless `git rev-parse HEAD` equalled that SHA and
  `git status --porcelain` was empty, but the retained logs do not capture the `git rev-parse HEAD` output, so there is
  no independent captured HEAD record. Corroboration only: B-SCOPE counted 75 changed paths since `48330587`, equal to
  `git diff --name-only 48330587 504fe8fc` (75 paths, re-counted at closeout).
- Earlier failures stay as recorded (round 10 run G FAILED; round 13 run L `lint-web` selection artifact; round 10 run
  D failed). The earlier round-15 concurrent runs N/O are history and are not counted.

**Cleanup receipt (2026-10-24).**

- Worktrees: all G020 worker and frozen worktrees were removed (writers, `g020-frozen9/11/13/15`, `g020-merge16`,
  `g020-merged-504fe8fc`); only the integration checkout and the coordinator worktree remain. Unique commits were
  preserved before any removal.
- Local branches: 10 finished branches that are ancestors of staging were deleted (`g020-impl-w4`, `g020-merge-r6`,
  `g020-r7` through `g020-r14`; plus, earlier, `g020-impl`, `g020-r15`, `g020-r15b`, `g020-merge16`). Remote
  `origin/g020-impl` was deleted by the merge; the stale tracking entry was pruned.
- Bundle: the 16 local `g020-*` branches not contained in staging (`g020-impl-pre-g064-35dd2b5f`, `g020-impl-w2`,
  `g020-impl-w3`, `g020-plan`, `g020-plan-r1`, `g020-plan-r2`, `g020-s1-mig`, `g020-s1-pol`, `g020-s1-rls`,
  `g020-s2-facade`, `g020-s2-http`, `g020-s2-it-labels`, `g020-s2-it-order`, `g020-s3-web`, `g020-s4-docs`,
  `g020-s4-harness`) and the remote `g020-plan` @ `d021cf68` (Plan r2 history) were first preserved in a verified git
  bundle, `<home>/.local/state/platform-evidence/g020-closeout/g020-unmerged-refs-20261024.bundle`: 17 heads (16
  local plus `origin/g020-plan`), `git bundle verify` OK (its 3 prerequisite commits are all in staging), SHA-256
  `a6aec77266f64da522439222a22fb8cd0d9a04f6e05f9c8ea5538cf08804f50e`; ref list in `bundled-refs.txt`.
- Authorization attribution: the deletions were performed on the answer "Delete G020 refs + evidence" recorded in the
  coordinator session's AskUserQuestion tool result. Codex subsequently stated that it selected that option itself,
  relaying the owner's standing instruction in the Codex session (cleanup is non-negotiable and tracked continuously;
  dangling containers/worktrees/tmp files are unacceptable; unique history must be preserved in verified archives), per
  handoffs `<redacted-uuid>` and `<redacted-uuid>`. This is a relayed
  standing authorization, not a new answer typed by the owner in the coordinator session; the coordinator acted on the
  tool result before that attribution was clarified. The original tool result stays in the session history unchanged.
  Everything deleted is recoverable from the verified bundle and tarball (checksums above and below).
- Local and remote refs: before deletion, the bundle and tarball SHA-256 values were re-checked (`sha256sum -c` OK) and the bundle
  re-verified. The 16 unmerged local `g020-*` branches listed above were deleted after confirming that each tip SHA
  equals its bundled head. The remote `g020-plan` @ `d021cf68` was deleted from origin
  (`git push origin --delete g020-plan`) after confirming that its SHA equals the bundled `origin/g020-plan` head, and
  `git fetch --prune` was run. Remaining: 0 remote `g020-*` refs, and 0 local `g020-*` refs apart from this closeout
  branch `g020-result-closeout`, which its MR deletes. The unique history now lives only in the bundle, which retains
  the plan history; release condition: none needed unless the owner wants a ref restored via
  `git fetch <bundle> <ref>`. The held G064 refs and the 112 local `worktree-agent-*` branches are untouched pending the
  owner (record `platform-evidence/g064-a1/HELD-REFS-20261023.md`).
- Evidence: failed-run proof and all run evidence were compacted without loss into
  `<home>/.local/state/platform-evidence/g020-closeout/g020-evidence-20261024.tar.gz` (2008 entries; dirs
  `g020-r4`, `r6`, `r7`, `r8`, `r9`, `r10`, `r11`, `r13`, `r15`, `r15b`, `postmerge`), SHA-256
  `0a856267c669089227ea34515bab1e6d596d9ec0b91675d0277d17c794bc1a29`. The original `g020-*` evidence directories (r4 190 files, r6 176, r7 292, r8 606, r9 440, r10 97, r11 25, r13 43, r15 22,
  r15b 24, postmerge 17: 1932 files) were removed after confirming that every directory is in the verified tarball. The
  tarball is the retained compacted failure/run proof. Paths in this Result that name
  `<home>/.local/state/platform-evidence/g020-*/` now refer to entries inside that tarball. No secrets directory
  exists for G020.
- Docker: 15 containers / 11 volumes / 6 networks / 52 images (baseline; 0 G020-owned). No G020 volumes, markers or
  leaked fixed-name volumes remain (the one leaked `project-archetype-render-cache` of round 10 was removed earlier,
  with its receipt in `g020-r10`).

This closeout records evidence only and changes no source, test or harness input.

#### Closeout correction — merge-order deviation and supplemental ref archive

- Merge-order deviation: the Result-only MR !96 (head `9333b0d9bb3cb5784d024751b78235e67e4d9d15`, tree
  `734ef48b030eb04030fd91862e098d373b2c5860`) was merged to staging as `63915cbf4ae283aa727d7cbd1c518bf8113139cf`
  (parents `69ff6c5a44ea9ac6b3fc6c8615e6f3844199313b` and `9333b0d9`; the merged tree equals the head tree) by the
  coordinator under the standing auto-merge rule (Duo found nothing, 0 unresolved threads, no pipeline for a Result-only
  MR) BEFORE Codex's closeout review was sent. It was not a pre-merge Codex ACCEPT of that text. Codex's later review
  found the text otherwise append-only and accurately qualified, and found one cleanup discrepancy (next bullet). This
  is recorded as a process-order deviation.
- Cleanup discrepancy found by Codex review: the receipt above says that no local G020 refs remain except the closeout
  branch and that the unique history lives only in the bundle. That was wrong. The local ref
  `refs/cleanup/20261022/g020-qual` still existed at `df2c46f486480ad13a61ef4c4b8f5716836e74ae`, with seven commits not
  in staging (S0 input classification `535f8f77`, scope amendment `ce0f82a7`, writer-1 checkpoint `a8955e16`, S1-P
  `61c78ebd`, S1 T3 `6ea43f0f`, S1 00020 migration `ab8eac31`, S1 audit integration `df2c46f4`), and it is not an
  ancestor of any head in the 17-head bundle `g020-unmerged-refs-20261024.bundle` (those heads were `refs/heads/g020-*`
  and `origin/g020-plan` only; the audit enumerated heads only, not `refs/cleanup`).
- Supplemental archive: `<home>/.local/state/platform-evidence/g020-closeout/g020-qual-ref-20261024.bundle`
  contains the single ref `refs/cleanup/20261022/g020-qual` -> `df2c46f486480ad13a61ef4c4b8f5716836e74ae`
  (`git bundle verify` okay; one prerequisite ref, in staging), SHA-256
  `2d5b3594d23dcaf3f8e291cdb383cf2f41cbcd9a254634778cd3e4af49e4294d` (checksum file `qualref.sha256`, verify output
  `qualref-verify.txt`).
- Retention, stated honestly: the ref `refs/cleanup/20261022/g020-qual` has NOT been deleted. Codex relayed the standing
  cleanup instruction, but the coordinator treats deletion of unmerged unique history as requiring the owner's own
  in-session approval (and the earlier cleanup deletions rested on an answer Codex selected; see the attribution bullet
  above). Retained; reason: unique unmerged history, now also preserved in the supplemental bundle; release condition:
  the owner's explicit in-session approval to delete it. The other refs under `refs/cleanup/20261022/` (`g064-*`,
  `agent-<id>`, `rc7-qual`), the held G064 refs and the 112 `worktree-agent-*` branches are untouched.
- Corrected statements (as of the first commit of this correction): other than this correction's own source branch
  `g020-result-correction` (deleted with its MR), 0 `refs/heads/g020-*` local refs and 0 remote `g020-*` refs
  remained; 1 local `refs/cleanup/20261022/g020-*` ref remained (retained as in the previous bullet; see the Update
  below for its later deletion). The unique history of the 16 deleted branches and of the remote `g020-plan` lives in
  the 17-head bundle; the `g020-qual` history lived in both the then-live ref and the supplemental bundle.

This correction records evidence only and changes no source, test, harness or allow-list input.

##### Update — local cleanup ref deleted after verified archival

- Deletion: after the correction above was committed, Codex deleted the local ref `refs/cleanup/20261022/g020-qual`
  (tip `df2c46f486480ad13a61ef4c4b8f5716836e74ae`) itself, as a compare-and-delete against that exact SHA, and reported
  that it had independently verified the supplemental bundle first. Coordinator re-verification afterwards: the ref is
  absent; `git bundle verify` of `g020-qual-ref-20261024.bundle` is okay and its sole head is
  `refs/cleanup/20261022/g020-qual` -> `df2c46f486480ad13a61ef4c4b8f5716836e74ae`; its SHA-256
  `2d5b3594d23dcaf3f8e291cdb383cf2f41cbcd9a254634778cd3e4af49e4294d` still matches (`sha256sum -c` OK); the main
  bundle (`a6aec77266f64da522439222a22fb8cd0d9a04f6e05f9c8ea5538cf08804f50e`) and the evidence tarball
  (`0a856267c669089227ea34515bab1e6d596d9ec0b91675d0277d17c794bc1a29`) also still match; the commit object `df2c46f4`
  is still present in the local object store.
- Attribution: the deletion was Codex's application of the owner's existing standing cleanup-and-preserve instruction
  (cleanup is non-negotiable; unique history is preserved in verified archives). It was performed by Codex, not by the
  coordinator, and it is NOT a fresh or owner-typed reply in the coordinator session; the coordinator had asked the
  owner and recorded no reply. As of this update, this supersedes the "has NOT been deleted" / retained-pending-the-owner
  statement in the retention bullet above.
- Final ref counts at this update (as observed by the coordinator before this update's writer started): 0
  `refs/cleanup/20261022/g020-*` refs; `refs/cleanup/20261022/` retains 9 refs, all untouched: the 7 `g064-*` refs
  (`g064-full`, `g064-full2`, `g064-s1`, `g064-s1b`, `g064-s3a`, `g064-s3b`, `g064-s4`) plus `agent-<id>`
  and `rc7-qual`; local `refs/heads/g020-*`: only the correction branch of this MR (deleted with its source branch on
  merge); remote `g020-*`: only that same branch; `worktree-agent-*` local branches: 114, i.e. the 112 held at the time
  of the first closeout plus the auto-branches of the two closeout writers, all untouched and held; G064 refs untouched.
  This update's own writer added its auto-branch `worktree-agent-<id>` (115 `worktree-agent-*` branches
  when it committed) and its local branch `g020-result-correction-b`, alongside `g020-result-correction`.

This update records evidence only and changes no source, test, harness or allow-list input.

#### G055 consolidated tooling — bounded usage observation (Spec AC8 / Plan §8)

- Cutover status: G055 had cut over before this use. G055 is closed (MR !63 merged at `5b813e32`, per PROGRESS). The
  canonical package is `skills/restruct-sdlc/` (parent `SKILL.md` plus the child skills `sdlc-management`,
  `repository-supervisor` and `claude-session-handoff`, and `agents/`), with the harness aliases
  `.claude/skills/restruct-sdlc` and `.agents/skills/restruct-sdlc`. The legacy standalone layout was not used.
- Entrypoint/discovery as actually used in G020: the coordinator session did not invoke the `restruct-sdlc` skill
  through the Skill tool; the only Skill-tool call recorded in its transcript was an unrelated `vestive-design` read.
  Discovery was through the repository's `CLAUDE.md` and `AGENTS.md`, which link into
  `skills/restruct-sdlc/sdlc-management/references/` (on staging `2e3285c3`, `restruct-sdlc` appears on 26 lines of
  `CLAUDE.md` and 38 lines of `AGENTS.md`). The coordinator followed those linked references (`coordination`,
  `validation-integration`, `closeout-risks`, `adoption`) as governance. No separate load of the child `SKILL.md` files
  is recorded, and this section does not claim one.
- Roles/handoffs observed: Codex as supervisor/reviewer (source reviews, and rulings relayed by cross-session handoff
  messages carrying handoff IDs); the Claude coordinator; and isolated stack-free writers (Fable for diagnosis and
  corrections, Opus for implementation and docs). Coordinator handoffs ended with the `COORDINATOR_STATUS:` marker from
  `coordination.md#coordinator-final-status-line`, using the classes review-ready/blocked/question/waiting/completed/
  working/decision. Every Result round, evidence correction, MR and merge went through that path.
- Cleanup evidence: continuous resource accountability and mandatory task-acceptance cleanup from `closeout-risks.md`
  were applied (per-run Docker baseline checks at 15/11/6/52, finished worktrees removed, verified bundle and tarball
  before removing refs and evidence, checkpoint notes). The receipt is the
  `### Final merge, post-merge qualification and cleanup receipt` section above.
- Observed gaps (factual; not a requalification): (1) the first cleanup audit enumerated `refs/heads` and the remote
  only, and missed `refs/cleanup/20261022/g020-qual`, which Codex's review found; (2) the coordinator merged the
  Result-only MR !96 before Codex's closeout review (recorded above as a process-order deviation); (3) a native
  AskUserQuestion answer was selected by a peer session, not typed by the owner (recorded above); (4) this AC8
  observation was itself missing from the first closeout and is added by a correction MR.
- Not claimed: this is an observation of use. It is not a G055 requalification and not a G021 barrier pass.

##### Update — closeout writer branches removed by Codex

- After the preceding update, Codex removed four unattached G020-closeout `worktree-agent-*` branches, after verifying
  that each was exactly `origin/main` `181733f2340fcc373d03baaadf5fd4a4fe6c06f6` with no unique commit and no attached
  worktree (Codex's report). Coordinator re-observation afterwards: 112 local `worktree-agent-*` branches remain, equal
  to the original 112 held refs, all untouched; `origin/main` is `181733f2`. This supersedes the 114/115
  `worktree-agent-*` counts in the earlier closeout text, which counted those auto-branches. No product, source or test
  change.

This section records evidence only and changes no source, test, harness or allow-list input.
