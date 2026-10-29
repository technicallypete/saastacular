> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G018 Result — Group creation and create-or-select Business Profile onboarding

## Status

**Initial Result (R0 prepared; no implementation started).** 2026-10-17 UTC. Accepted design: Plan r5
(`management/v0/plans/G018-group-business-profile-onboarding-plan.md`, SHA-256 `4fe4642f…75fa`; Codex ACCEPT —
provisional technical design — in `management/v0/results/G018-plan-codex-review-r5.md`, amended SHA-256
`5184febe…ad11`). Current-target R0 reconciliation against the merged G017 (`3cc18228`, source-identical to
`origin/staging` `89c7e8ab`): `management/v0/plans/G018-group-business-profile-onboarding-plan-r0-addendum.md`
(§A gate change, §B drift table — **no material drift**, one non-material seed fixture reconciliation —, §C re-pinned
counts, §D hashes, §E scope unchanged at 43 patterns, §F resources, §G D-1/exclusions, §H gates).

Gate state at this Result: G017 merged (MR !40) with staging pipeline `2958742766` running and **no final ACCEPT**;
per the owner amendment (staging `89c7e8ab`) that ACCEPT gates G018 **integration and final acceptance**, not its
start. Still required before implementation: Codex confirmation of the R0 bindings below and the recorded execution
release. Budget **B7, checkpoint after attempt 4: 0 of 7 spent.** G019/G058 holds unchanged.

Worker record. R0: task G018, role plan/R0; native isolated worktree
`<source-checkout>` (generated
`agent-<id>` name, the accepted AGENTS.md fallback; task/role/path recorded here), branch `g018-r0` from
`origin/staging` `0bc68f0f` with `89c7e8ab` merged; model Fable 5.1, planning effort "high (requested)", actual value
unknown. No tests, builds, probes, containers or application/test edits were run or made.

## R0 binding (Git-first; prepared for Codex to confirm)

| Label | Bound bytes |
|---|---|
| Target `origin/staging` at the R0 fetch, 2026-10-17 (contains the G017 merge `3cc1822826754ff70cf6346809f62b417080832c`; the three later commits are management-only, `git diff --stat 3cc18228 89c7e8ab -- . ':!management'` empty) | `89c7e8abc7e9294bdb6fd5631db1b43529db18c7` (tree `92b2ce35398621a1b9377f324a414ba9f898ef52`) |
| Plan r5 (P5) / archives r4 / r3 / r2 / r1 / r0 | `4fe4642ff0d5457db156ff3d5d31e723891dad1b4dd303a86ef0f70fb6dd75fa` / `55bc7821a7b723c28fd47d7ee2579ddf66b2ee9affc035162e1bd046ff00b583` / `963a97cabbe86de16017430cfedbf9fd12a19c53c1d834fcd860d4aa0d348c3e` / `120625d108637ddc0766da0bceb234400523b96be91e722e483ba2d4bac030c4` / `e9ac8c83c0780abca4b99a154276238c0514ae5dd3a4a5148ae7d6ef042e05eb` / `c7a64649f35692e3b985d292026cc4e21649d293ad8e9ec0d86eb98718574c59` |
| R0 addendum (PA), this branch's commit | recorded at hand-back (the commit on `g018-r0`) |
| Scope current (SC) `results/G018-scope-allowlist.txt`, 43 executable patterns (the r5 proposal `plans/G018-scope-allowlist.txt` `5a94f3ac6cc585bfd437e3b5cf7b577bec147ca0c2d3d816d9e8e8d267a84a7e` unchanged; header line only) | recorded at hand-back |
| Goal (G) / Spec (S), both with the 2026-10-17 owner amendment | `36f4e91f8a42990e127845c755ebe06482c4b450d391b00f3f5c297cd3210bf4` / `93e03e9f525c627855f21e8313ccccbeb46ae0400443169833016bcad5ae9f8a` |
| Reviews r5 (amended) / r4 / r3 / r2 / r1 / r0 | `5184febe95b79c38fd1c217858c1bc940c8c390ed5060cd5aebf1ea28fc7ad11` / `2f741b02d433654e8012acfcd4edd8f0675a724cfb1f27b9f03c134e7afc6532` / `5567da3ca2db4242298fd3b8bce6fe4feefec90ea72fb94616a879d4ad5b2e86` / `423f8107a980f65c1a1b04e0121b9bc57b8bc54facb8dbd6b6ef0a72a4926ebc` / `ec3855ed39cca99c7e10e49209781ca260e42b6763dc6c411f2b244f835f3b45` / `b545ebc02b290f314daf462babe4c3a318b345d4e80be3dcf3c03e635f1ee663` |
| G017 Result (X17; G017 not finally accepted) / PROGRESS (PR) / ROADMAP (RM) | `1b79d03606c09d12a256eeee500fa82f3ace1c0159e6521c8fbd57e5b0dfcebc` / `79d5450ce9d8ee807f87d2079f1495d38939124dffe087f5c6af6210c2e4823c` / `ba42d247c15aacc9775c3b4c2e0eaec1dc1bfcdac4a795854d7a5ac58fc193b0` |
| AGENTS (AG) / CLAUDE (CL) | `5f1f06a0465508839cbaf735a75a698d7fe0b82615b5c24551ce2b1e6ad13af8` / `bbcffcb4acfceeea6705d7981b1530b731f6f79e1373c0fe5263f0c2186a882b` |
| Architecture (AR) / Security (SE) / Performance (PF) / Decisions (DE) | `d25bc7028d585aefd50663af1487aa8b77b835b18cdf03a2b7aff5dc3079f0f4` / `143de1d2f5278817a363dfeaf995ab807e2d1899f1623aa2fcd7becddb94cf00` / `d49ec8e049e0cae12af0b4dea60904369073ed4015fd7b7c7bae19ca05d0f3cf` / `6c713d58682caceb4bdfcca45a63bcef89f976355c92bbdf46459372199324d4` |
| Foundation notes TENANCY / AUTHORIZATION / DATABASE-ISOLATION / AUDIT / MIGRATIONS / WEB-SHELL / FOLLOW-UP / RECOVERY | `e94605ae3cc7b4a63d841c94d48ad3c1f1cf52b0e5eb9bbe15b889e4627f2d3d` / `34ca5dad9ba22fc6c583b09844ac3ea6ebe30c0cd357d744eacc21f6a5acce12` / `4284f7b70bc6eae782d5e2906845091f7c206295627e1692b388864d10d7ca54` / `724166cd81867ee2804f0be01a509e600292f32709587107c39d1c863fe33529` / `0653d020563f6c8ba93955e80ebfbeada9fe9e0d06fa7b2cf363c0e8ea4fcdf8` / `038d24b09ac42a3dd2c3932fbccba5469592de3921b43f01536b88194709cb1c` / `b65e2d07fa13c95b534fd4a5cfbe14ce60442cfea8dc75777ed0d6e1b300eafd` / `7534448f30e5960546da6de6538c7fb268c8daa39ed63ab6cb1ba930f4ff1458` |
| Harness `run.sh` (RS) / `required-checks.tsv` (RC, 121 rows) / `lib.sh` (LB) / `validate-controls.sh` (VC) / `profile-paths.tsv` (PP) | `75cba0a732c2206852f25da4d5db2e0327350efe7daac35cac9a0867c5d87007` / `00bcd691b4e6db7572242a24a999b4a05fd01afa2248e933f5943ace2a5f3292` / `748b25aebed4f8ecd01d0d45c38c65e810490c047b2051e2f40b7935966e6b58` / `8b4933ae70157e43593f48040deb8329e3e135fb8db3273f75806da81b4a4bb0` / `d78868bd9b396a0e9640d9418802f3ec82323ccff94f1554302206f9e751ef5b` |
| Log/route enums `completion.schema.json` (CS) / `events.json` (EV) | `1552e38381d607bab50542c8b5e34722f53514c5c1fbaece621ab25b4c9e67a2` / `e6b05710bc6fcf418174b090fa013ef50e6098ec6b6029fa3a1118136bbcefc9` |
| Seed (SD) / Playwright (PW) / last migration 00013 (M13) / Vestive skill (VS) | `9e9988a95257d5675045e6fbd871754c342d0dcd0eeb67f3763692ce7bfbdc35` / `462cbf5a92a767c582c4b709b2883dab65dd3d968ce42c425cb19984a3cd28cc` / `4a1ff388df58b649645f227eff40791d6adf4aae34a0187409d28a9c95928656` / `7dfac5faecadb0b297a1be5641b1b86140dd28e6b4d32c78b140bc1cb65c670e` |
| D-5 copy source `branding/api/http.go` (BA) / `web/lib/session.ts` (SS) / `app-shell.tsx` (AS) | `c65b5127a2fe54188ea56455fd149016d1929099e83c7c899450c225dd7fbaba` / `14a2e350371a1ffe573e9edd17838983739655020d8c8d2e34310e609bd71c62` / `e824bcc540430d85a3113404e9bf2065a1139d723f3ddb785590abc7f260293d` |

**Bound paths.** P5/P4…P0 `management/v0/plans/G018-group-business-profile-onboarding-plan{,-r4,-r3,-r2,-r1,-r0}.md`;
PA `management/v0/plans/G018-group-business-profile-onboarding-plan-r0-addendum.md`; SC
`management/v0/results/G018-scope-allowlist.txt`; G `management/v0/goals/G018-group-business-profile-onboarding.md`;
S `management/v0/specs/G018-group-business-profile-onboarding-spec.md`; reviews
`management/v0/results/G018-plan-codex-review-r{5,4,3,2,1,0}.md`; X17
`management/v0/results/G017-organization-branding-logged-out-shell-result.md`; PR/RM `management/v0/{PROGRESS,ROADMAP}.md`;
AR/SE/PF/DE `docs/{ARCHITECTURE,SECURITY,PERFORMANCE,DECISIONS}.md`; foundation notes `docs/foundation/<NAME>.md`;
RS/RC/LB `infra/validation/{run.sh,required-checks.tsv,lib.sh}`; VC/PP `infra/ci/{validate-controls.sh,profile-paths.tsv}`;
CS/EV `infra/validation/tools/internal/logrecord/{completion.schema.json,events.json}`; SD `infra/identity/mcp-seed.sql`;
PW `web/playwright.config.ts`; M13 `core/migrations/00013_organization_branding.sql`; VS
`web/.claude/skills/vestive-design/SKILL.md`; BA `core/internal/branding/api/http.go`; SS `web/lib/session.ts`; AS
`web/components/layout/app-shell.tsx`.

## R0 reconciliation summary (addendum §A–§H)

- **Interfaces (§B):** P-1..P-11 re-cited at `89c7e8ab`; counts unchanged (13 capabilities; 36/30/58/29 audit; 11
  `inContextTx`; 14 production files; definer allowance 00013 = 1; `ShellNavItem.href` anchors only; `readForm`
  single parameter; `MAX_FORM_BYTES` 16 KiB; identity projects at `playwright.config.ts:102-105`). Line drift only in
  `matrix_test.go` (`:850`), `integration_test.go` (`:1420`) and the `run.sh` version-13 pins. `required-checks.tsv`
  is 121 as r5 anticipated. **Migration `00014` is free** (no file, no reference outside `management/`).
- **Seed (§B P-10, non-material):** G017 made carol Organization A's `org_owner`; r5 §3's "Carol stays `org_member`"
  is reconciled — carol keeps the G017 role, the plain-member browser actor is another seeded `org_member`, the +4/+3/+2
  seed additions stand. No interface change; no Plan revision.
- **Scope (§E):** 43 patterns, no additions; examined-and-rejected candidates listed.
- **Resources (§F):** single writer on `g018-group-business-profile-onboarding`, `g018-core-check:b<N>` /
  `g018-web-check:b<N>` images, `g018-focus` database project, native worktree; implementation Opus high.
- **D-1 (§G):** superseding decision (owner/admin/member create) still the Spec's and Plan's rule; no public
  enrollment, G019 or G058 work included.

## Execution record

Budget B7, checkpoint after attempt 4. **Attempts spent: 0.** No implementation, validation batch, manual feedback or
MR exists. The first attempt opens only after Codex records the execution release against the bindings above.

## Execution release and implementer record (attempt 1)

**Release:** `management/v0/results/G018-r0-execution-release-codex-review.md` (SHA-256 `6e08f838c1d8752ec2e3695314e27373add2b8e482344ced384f0efe3aefc1b8`),
Codex ACCEPT of the R0 binding and execution release, 2026-10-17. Governance delta recorded per the release: it
supersedes the R0 table's AGENTS/CLAUDE hashes with AGENTS `069013bdb49acc577198301293ab45606c8aad70023f3b3ee5cc5dc630d7ed10`
and CLAUDE `46dde951a98c2afc7c4792f390b3756989679e01b098b4bfb2307f9150178a7e` (integration `b4d4be7d`: governance,
portable skills/templates and management only; no application/harness delta). Release clarifications applied: matrix
primary cells **286 → 330** (two capabilities × 22 primary cells each, computed from the real loop at implementation);
`web/components/layout/app-shell.tsx` navigation is in scope as planned (only the specifically listed G017 shell
components stay untouched); carol keeps her G017 Organization A `org_owner` role and the plain-member actor is an
actual seeded ordinary `org_member`; no public enrollment, G019 or G058 work.

**Implementer record (verified before the first write, 2026-10-17 UTC):** task G018, role implement, sole writer.
Native isolated worktree `<source-checkout>`
(generated `agent-<id>` name, the accepted AGENTS.md fallback; not renamed); branch
`g018-group-business-profile-onboarding` created at `origin/staging` `d1fd674b6c4336c9cac381e437f10f04e3178d9d`
(no prior remote branch of that name; the planning branches `g018-plan-r1/r2/r4`, `g018-r0` are untouched). Model
Opus 5.5; implementation effort "high (requested)", actual value unknown. Bindings re-verified at `d1fd674b` by
`sha256sum`: release `6e08f838…`, Plan r5 `4fe4642f…75fa`, R0 addendum `829cfe4f…4dc84`, initial Result `b23b3fb2…6596`,
scope `76f1878f…5d11` (43 executable patterns), Goal `36f4e91f…0bf4`, Spec `93e03e9f…f8a`, AGENTS `069013bd…7ed10`,
CLAUDE `46dde951…a7e` — all equal. Resources at start: no `g018-*` container, image or compose project exists; the
running `g017-b4` stack belongs to another task and is not touched. Shared harness lock
`<home>/.local/state/platform-evidence/local-harness.lock`; scripts under
`<home>/.local/state/platform-evidence/g018-scripts/`, temporary evidence under `…/platform-evidence/g018/`.

Budget: B7, checkpoint after attempt 4; **attempt 1 open** (0 spent before it).

### Slice 1 — stopped before application edits: audit operation constraint gap (hand-back (c))

**Finding (design contradiction inside the approved risky migration, discovered while reading the slice-1 inputs at
`d1fd674b`; no code written).** Plan r5 §5.2/§11.2 and the R0 addendum §B P-3/§C add **seven `access.denied`
operations** (`group.suggest_slug`, `group.create`, `group.list`, `group.onboarding.get`,
`group.onboarding.complete_new`, `group.onboarding.complete_existing`, `business_profile.list`; 29 → 36 in
`core/internal/audit/audit.go`) while r5 §3/§4.6 and the addendum state that **00014 touches no audit constraint and
`TestCatalogParity` keeps reading 00013**. Both cannot hold:

- `audit.events.operation` carries the CHECK constraint `events_operation_valid`, last replaced by
  `core/migrations/00013_organization_branding.sql:303` onward (the 29 version-13 operations); every guarded denial is
  written with its operation (`core/internal/tenancy/access/audit.go:115-128`, `recordDenial`). An `access.denied` row
  for any of the seven new operations is refused by the constraint; `record` then only logs `audit.record_failed`, so
  every new-operation denial would silently go unaudited (AT-INV and the AUDIT.md invariant fail).
- `core/internal/audit/audit_test.go:69-95` (`TestCatalogParity`) requires `audit.Operations` to equal the
  `events_operation_valid` list read from 00013 and pins `len(Operations) == 29`; adding the seven operations while
  reading 00013 fails b1.

Checked and **not** affected: actions/reasons/resource types/outcomes (unchanged by G018); `audit.events_attribute`
(T4/T7/T10/T12 from the bootstrap and capture triggers reach it at depth ≥ 2, so the trigger-source rule admits them
unchanged); B-AUDIT-CATALOG (`run.sh:3115` compares actions only); `core/internal/migrate/audit_integration_test.go`
(pins constraint names, not lists — out of scope and unchanged under the recommended disposition).

**Smallest disposition proposed (no allow-list change; every file is already inside the 43 patterns):** 00014 also
replaces `events_operation_valid` with the version-13 list plus the seven G018 operations (the 00011/00013
drop-and-add constraint pattern; one validating scan of `audit.events` inside the migration transaction); actions,
reasons, resource types, outcomes and `audit.events_attribute` stay byte-unchanged; 00014's down refuses while any
`audit.events` row carries one of the seven operations (history is never discarded — the 00013 down pattern) and
otherwise restores the version-13 constraint; `TestCatalogParity` reads the operation list from 00014 (the other five
lists keep reading 00013) and pins 36; CT-ONB-1 asserts the widened constraint and the down refusal; AUDIT.md and
MIGRATIONS.md record it. Rejected alternatives: recording the new denials under existing operations (misattribution)
or not recording them (breaks AT-INV). Supersedes only the r5/addendum sentences "00014 touches no audit constraint" /
"`TestCatalogParity` keeps reading 00013"; no other design element changes.

**Read-only readiness probe (precondition 1 of r5 §4, informational):** `flock …/local-harness.lock
…/g018-scripts/probe-pgcontrol.sh` after `bash infra/ci/docker-desktop-mounts.sh` (13,676 of 100,000 mounts, 13%):
throwaway container `g018-probe-pgcontrol` from the pinned `pgvector/pgvector:0.8.7-pg17` (PostgreSQL 17.11), no host
port, no project resources — `pg_control_system()` and `pg_current_xact_id()` carry the default PUBLIC ACL and a fresh
NOLOGIN role has EXECUTE on both (`t|t`). Container stopped and removed (verified `docker ps -a --filter name=g018`
empty). No migration, test, build or harness row ran.

**Attempt state:** attempt 1 opened at the release; paused before any application, test, harness or doc edit, pending
Codex's disposition of the finding above (whether this pause consumes attempt 1 is for Codex's ledger reconciliation).
Outstanding-findings checklist: (1) audit operation constraint — awaiting disposition.

**Target advance reconciled (safe boundary, before the safekeeping push):** `origin/staging` advanced
`d1fd674b` → `23abbbeb87020d79ab78e9ad26031af8f6ff9970` ("Keep predecessor responsible for staging repairs during
successor implementation"): governance, skills, management only (`git diff --stat d1fd674b 23abbbeb -- . ':!management'
':!skills' ':!AGENTS.md' ':!docs/agents'` touches only `.agents/skills/*`); no application, harness, test or migration
delta. Branch rebased onto `23abbbeb` (two Result-only commits, no conflict); ancestry verified. Owner clarification
recorded (process record `management/v0/results/PROCESS-implementation-overlap-20261017.md`
`6be9d5e324533705f07bdab22b151a9083eaa5161a4e7d20a483620ad5b41936`): G017 staging diagnosis/repairs stay with the
G017 owner on its own branch/ledger; this worker stays on G018 only; only unsafe affected work pauses; G017 final
ACCEPT still gates G018 integration/final acceptance. Rebound governance at `23abbbeb`: AGENTS
`186f4edd6d252b81039ff872ef64e467b655e47bb1cbaa7f657c395b3bc346d0` (was `069013bd…`), CLAUDE `46dde951…a7e`
(unchanged), Goal `0cef83762027801eb40bc382493b11ea38c7615d6ea19de223ca5cc46cc9f9ee`, Spec
`408095ef5803c2f42f6843dcdf0bbce2cfc303c5f0d8d87266828fcd6eaf3678` (each +13 lines: the same owner clarification
appended), release review `3827fad2399855cbfbf59405b8aecfbbd4921f0ce1889a2d5b87f584e54c9bfc` (same clarification
appended). Plan r5 `4fe4642f…`, addendum `829cfe4f…` and scope `76f1878f…` unchanged.

### Audit-operation amendment accepted; slice 1 resumed (attempt 1) — WIP checkpoint

**Disposition:** `management/v0/results/G018-audit-operation-amendment-codex-review.md` (SHA-256
`1fde7e79bdc7efbe25bdc24746b1dcd5e4b48abb4db27f4d0c9a4b4f0032996e`, at `origin/staging` `e8187f57f68fca0e21ed54235c8b5d584ddaf655`)
accepts the proposal above (Result source `f7cd5e2f`, Result `f08a5e28…01d8`; initial binding Result kept as
`G018-group-business-profile-onboarding-result-initial-binding.md`). It supersedes only the r5/R0 sentences "00014
touches no audit constraint" / "`TestCatalogParity` keeps reading 00013"; r5 and the R0 addendum stay byte-immutable.
The pre-code pause is not a failed round or a new attempt. Same writer, worktree and branch; `origin/staging`
`e8187f57` (management-only since `23abbbeb`) merged at `9baf74a4`, ancestry verified; AGENTS `186f4edd…46d0`, CLAUDE
`46dde951…a7e` unchanged since the last reconciliation. G019 planning authority change acknowledged (management only).

**Implemented in this WIP checkpoint (unvalidated — no batch has run):**

- `core/migrations/00014_group_onboarding.sql` — r5 §4 in full plus the amendment: precondition `DO` (owner, version-13
  shape, no runtime INSERT/policy-reader SELECT on the four tables, the eight 00004 `business_profiles` columns,
  `pg_control_system()`/`pg_current_xact_id()` executable by runtime and policy reader); `created_xid`/`created_sysid`
  `ADD COLUMN` with defaults; replaced `business_profiles_keys_immutable()` (four-column comparison); the
  `events_operation_valid` single-statement DROP/ADD with the 29 + 7 operations; `groups_bootstrap_owner()` SECURITY
  INVOKER trigger (role names fixed at migration time, no-op for other roles or without `ctx_user()`), AFTER INSERT
  trigger on `app.groups`; column-limited runtime INSERT grants (`groups (id, organization_id, name, slug)`,
  `role_bindings` seven columns, `business_profiles (id, organization_id, display_name, legal_name, description, industry,
  website)`, `group_business_profiles (organization_id, group_id, business_profile_id)`); policy-reader column grants
  and the two helper policies; `app.rls_profile_eligible(uuid)` created as the policy reader (context first, then
  branch (a) insertion provenance or branch (b) owner/admin path), EXECUTE to runtime and owner only; the four INSERT
  policies (`groups_insert` with `R IS NOT NULL`, `role_bindings_insert_bootstrap` with `pg_trigger_depth() > 0`,
  `business_profiles_insert`, `group_business_profiles_insert` with the reader); postconditions (four-table ACL string,
  six policies, two functions, `app.groups` triggers, provenance columns/defaults, replaced body). Down: refuses while
  any `audit.events` row (FORCE lifted inside the transaction only, restored on success, rolled back on refusal) carries
  a G018 operation, then drops policies/reader/trigger/function, revokes grants, restores the 00004 body, drops both
  columns and restores the exact version-13 operation constraint. **The postcondition ACL string is computed from
  00006/00011 and not yet executed — expect to correct it at the first b2 run.**
- `core/internal/audit/audit.go` — seven operations (`OpGroupSuggestSlug` … `OpBusinessProfileList`), 29 → 36.
- `core/internal/audit/audit_test.go` — `TestCatalogParity` reads the operation list from 00014 and pins 36; other
  lists unchanged.

**Known-red at this checkpoint (expected until the remaining slice-1 items land):** every version-13 pin (r5 §3 /
addendum P-4: `migrations_test.go`, `runner_test.go`, `migrate/integration_test.go`, `tenancy/integration_test.go`,
`access/integration_test.go`, `recovery_integration_test.go`, `mcp/env_integration_test.go:253`, `run.sh` hits) and the
audit-catalog-dependent checks until their pins move to version 14.

**Remaining slice-1 list (precise, for the successor):**

1. Validate/correct 00014 (first b2 run of `^TestOnboardingCatalog$`); write `core/internal/migrate/onboarding_integration_test.go`
   `TestOnboardingCatalog` CT-ONB-1 (grants, policies, trigger/reader ownership/search_path/volatility, provenance
   columns/defaults/backfill, immutability refusal per column as owner, `has_function_privilege` pins, insertion predicate
   true/false cases, reader false without context, **amendment: widened constraint shape (36 operations), unknown-operation
   refusal, a history-preserving down refusal and a clean down/up**) and R-ONB-1 neg/pos subcases (r5 §6.2).
2. Version-14 pins (the list above) incl. `metadataThroughVersion14Only` and the `run.sh` version-13 greps.
3. `access/policy.go`: `group.create` (Organization kind), `group.complete_onboarding` (Group kind) and matrix rows;
   `matrix_test.go`/`policy_test.go`/`integration_test.go` cells 286 → 330 computed from the real loop.
4. `tenancy/store.go` constants (`sqlInsertGroupKeyed`, `sqlInsertBusinessProfileKeyed`, `sqlSelectGroupByID`);
   `tenancy/onboarding.go` (`NewGroupID`, `CreateGroupTx`, `CreateBusinessProfileTx`, `LinkGroupBusinessProfileTx`);
   `tenancy/facts.go` ascending-id partition of other candidates around a locked target (r5 §1.4); `boundary_test.go`
   (15 files, +5 functions, new statements in the 50/13 count).
5. `access/onboarding.go`: guarded `CreateGroup`, `CompleteGroupOnboardingWithNewProfile`, `CompleteGroupOnboardingWithProfile`,
   `GetGroupOnboarding`, `onboardingQuery(g, p)`; `errors.go` `association_exists`→`conflict`; `log.go` operations;
   `access/boundary_test.go` (`inContextTx` count, 00014 definer allowance = 1, exported methods).
6. Tests ONB-1..9, ONB-13 (`TestOnboarding` in `access/onboarding_integration_test.go`, incl. **an actual new-operation
   denial persisted through the normal runtime path** — amendment), O-ONB-1..3/D/E in `ordering_integration_test.go`,
   `rls_expected_test.go` r3 rows, attack rows; the r5 §6.3 subcase-selection grep recorded here.
7. b1 then b2 with the r5 §6.3 focused selectors (g018-core-check image, g018-focus project, lock route), fix, rerun.
8. Then slice 2 (lists, slug suggestion, HTTP, pins) and hand back (b) for the draft MR.

### Successor #2 implementer record (attempt 1 continued; verified before the first write, 2026-10-17 UTC)

Task G018, role implement, sole writer (fresh successor #2). Model Opus 5.5; effort "high (requested)", actual value
unknown. Native isolated worktree
`<source-checkout>` (generated `agent-<id>`
name, the accepted AGENTS.md fallback; not renamed); local branch `g018-impl2` tracking
`origin/g018-group-business-profile-onboarding`, created at the WIP checkpoint `3da9dfa3bdb28b25156991a5e677bb467de152f4`;
`origin/staging` `b1a2565660a1fd90abe306dea5e3379f7b61b91b` (management only: SUPERVISOR-HANDOFF, PROGRESS) merged at
HEAD `896e8de72665160d6f1b33dd64414d6bc477b951` before any write. Bindings re-verified by `sha256sum` there: Result at
`3da9dfa3` `581c81f5…bd2c`, scope `76f1878f…5d11` (43 executable patterns), release review `3827fad2…9bfc`, audit
amendment `1fde7e79…996e`, Plan r5 `4fe4642f…75fa`, R0 addendum `829cfe4f…4dc84` — all equal. Budget B7 (checkpoint
after attempt 4); attempt 1 continues, no reset.

### Successor #2 checkpoint — 00014 validated; remaining slice-1 items 1–2 done (attempt 1, hand-back (c))

**Target reconciliation:** `origin/staging` advanced `b1a25656` → `4afbb3f6ef77cc193811e06dffd879e8201ab18c` ("Set
autopilot endpoint before G020 planning including G019 and G058 delivery": management only — SUPERVISOR-HANDOFF,
PROGRESS, ROADMAP, G019/G058 Goal/Spec); merged at `7b47bcf9` before the first batch; ancestry re-verified before each
batch (`g018-scripts/prebatch.sh`); Docker Desktop mounts 13,676 of 100,000 (13%) each time.

**Changes (all inside the 43 patterns):**

- `core/migrations/00014_group_onboarding.sql` — one correction found by the first run: postcondition 3 concatenated
  `pg_proc.provolatile` ("char") without a cast (42725 `text || "char"`); now `provolatile::text`. Every other
  postcondition, including the four-table ACL string the predecessor computed, held on the first executed run.
- `core/internal/migrate/onboarding_integration_test.go` (new) — `TestOnboardingCatalog` on a populated version-13 →
  14 upgrade: **CT-ONB-1** (runtime INSERT/UPDATE column sets, policy-reader columns, no table-level non-SELECT
  privilege, no runtime INSERT/UPDATE on `created_xid`/`created_sysid`/`created_at`, `has_function_privilege` of
  `pg_control_system()`/`pg_current_xact_id()` for runtime and policy reader, the six policies' roles and load-bearing
  predicates, both functions' owner/SECURITY DEFINER/volatility/search_path/ACL, `md5(prosrc)` of the reader and of the
  replaced immutability body against the migration text, the bootstrap body and trigger, provenance columns/defaults,
  backfill `created_xid <> current` and `created_sysid = live identifier`, owner refusal 23000 for `created_xid`,
  `created_sysid` and `id`; runtime: the keyed Group insert without RETURNING as `org_owner` **and** `org_member` with
  the trigger-written owner binding read back by S-G2; 42501 for a Group insert in selected-Group context, by a
  Group-role-only subject, into a foreign Organization and without context, for depth-0 owner-binding inserts (other
  subject and own), and runtime DELETE on all four tables; insertion predicate true for a same-transaction insert,
  false for a committed row, for a committed row UPDATEd in the transaction and without context; profile insert
  without context 42501), **R-ONB-1-neg/-pos** (r5 §6.2 fixture exactly: superuser restore-shaped insert with
  `created_xid = pg_current_xact_id()` and a foreign `created_sysid`, collision asserted real, one-row UPDATE, then
  `SET LOCAL ROLE` runtime with the member's Organization-only context: reader false, association 42501; positive:
  S-P1 + reader true + S-P2 one row, committed, provenance checked on the superuser connection), **CT-ONB-AUD**
  (constraint lists exactly 36 operations incl. the seven; `group.unknown` 23514; each new operation admitted; probe
  rows written with triggers disabled for the superuser fixture only, rolled back), **CT-ONB-UPDOWN** (clean down →
  `rolled_back 00014 / current_version 13`, `metadataThroughVersion13Only`, catalog equal to the pre-upgrade version-13
  snapshot; up → catalog equal to the version-14 snapshot; with one committed `access.denied`/`group.create` row the
  down refuses with the history message, catalog unchanged, the row still present, `audit.events` FORCE restored).
- Version-14 pins: `migrations/migrations_test.go` (set 00001–00014), `migrate/runner_test.go` (reviewed set, 14
  sources), `migrate/integration_test.go` (`metadataThroughVersion14Only`; lifecycle up/synthetic down/reapply/down
  chain 14→13→12…/up again; `reset` drops `app.rls_profile_eligible` as the policy reader and
  `app.groups_bootstrap_owner()` after the tenancy tables), `tenancy/integration_test.go` (purge audit history, then
  14 → 13 before the existing chain), `access/integration_test.go`, `access/recovery_integration_test.go`,
  `mcp/env_integration_test.go:253` (literal only), `infra/validation/run.sh` (B-MIGRATE-SET set and message, root
  migrate block pending/applied/applied-status lines and versions, B-MIGRATE-CLI steps — reapply 00014, rollback
  00014 → 13 at 50 relations, inspect 00013 applied/00014 pending, apply 00014 → 14, production version 14 — and
  message, B-MIGRATE-CATALOG comment/message, B-RECOVERY-DRILL version, B-PERSIST history grep/message, B-DBRENAME,
  B-TOGGLE-OFF apply/inspect/message, B-IDENTITY-CATALOG message; `bash -n` clean; B-AUDIT-CATALOG still reads actions
  from 00013, unchanged by G018).

**Focused outcomes** (image `g018-core-check:b1` from `infra/docker/core.Dockerfile` target `check`; project
`g018-focus` via `compose run --use-aliases`, no host port; every Docker command under `flock …/local-harness.lock`;
logs under `platform-evidence/g018/`):

| Selector | Result |
|---|---|
| compile: gofmt -l; vet `./internal/tenancy/... ./internal/audit/ ./internal/migrate/ ./migrations/ ./cmd/server/`; vet -tags integration `./internal/tenancy/... ./internal/migrate/ ./internal/mcp/` | 0 files; pass; pass (first run: duplicate `sortedPair` helper — removed, the package's existing one reused) |
| unit `^(TestCatalogParity\|TestMigrationSet\|TestNormalSetIsTheReviewedSet)$` (audit, migrations, migrate) | 3/3 PASS |
| `./internal/migrate/ ^TestOnboardingCatalog$` | run 1 FAIL (00014 42725, fixed); run 2 FAIL (test: `format('%s', bool)` renders `t`; reserved slug `new` → `fresh-group`); **run 3 PASS 6/6** |
| `./internal/migrate/ ^TestMigrationLifecycle$` | PASS 11/11 |
| `./internal/tenancy/ ^(TestTenancyModel\|TestBusinessProfileModel)$` | PASS (72 results) |
| `./internal/mcp/ ^TestMCPTransport$` | PASS 8/8 |
| `./internal/migrate/ ^TestRLSCatalog$/^(CT-RLS-2\|CT-RLS-3\|CT-RLS-5)` | PASS 4/4 |
| `./internal/tenancy/access/ ^TestRecoverySeed$` | version-14 pin passed (migrate up reached 14 and the seed ran); the test then stops at `RECOVERY_OLD_DIR is required` — a B-RECOVERY-DRILL-only setting (r5 §6.3: the drill is not rerun locally; hosted migrations suite evidence). Earlier attempts: missing object-store settings (rerun with the focus MinIO), `identity_namespace_mismatch` on a database already bound by TestMCPTransport (rerun on a fresh focus database) |

**Cleanup:** `focus.sh down` — containers 0, volumes 0, networks 0, secret files 0; `docker ps -a --filter name=g018`
empty. Kept for reuse: images `g018-core-check:b1`, `g018-focus-minio`, `g018-focus-minio-init`.

**Remaining slice-1 list (precise, for the successor):**

3. `access/policy.go`: `group.create` (Organization kind), `group.complete_onboarding` (Group kind) and matrix rows;
   `matrix_test.go`/`policy_test.go`/`integration_test.go` cells 286 → 330 computed from the real loop.
4. `tenancy/store.go` constants (`sqlInsertGroupKeyed`, `sqlInsertBusinessProfileKeyed`, `sqlSelectGroupByID`);
   `tenancy/onboarding.go` (`NewGroupID`, `CreateGroupTx`, `CreateBusinessProfileTx`, `LinkGroupBusinessProfileTx`);
   `tenancy/facts.go` ascending-id partition (r5 §1.4); `tenancy/boundary_test.go`.
5. `access/onboarding.go` guarded ops (`CreateGroup` keyed, no RETURNING before the owner insert — the SQL sequence is
   now proven by CT-ONB-1; `Complete*`; `GetGroupOnboarding`; `onboardingQuery(g, p)`), `errors.go`, `log.go`,
   `access/boundary_test.go`.
6. Tests ONB-1..9, ONB-13 (incl. one new-operation denial persisted through the normal runtime path), O-ONB-1..3/D/E,
   `rls_expected_test.go` r3 rows, attack rows; the r5 §6.3 subcase grep recorded here.
7. b1 then b2 with the r5 §6.3 focused selectors (scripts `g018-scripts/{prebatch,compile,unit,focus,pins-it}.sh`;
   `focus.sh it NAME PKG REGEX [obj]`; TestRecoverySeed/TestMCPTransport need a fresh database each).
8. Slice 2, then hand back (b) for the draft MR.

### Successor #3 implementer record (attempt 1 continued; verified before the first write, 2026-10-17 UTC)

Task G018, role implement, sole writer (fresh successor #3). Model Opus 5.5; effort "high (requested)", actual value
unknown. Native isolated worktree
`<source-checkout>` (generated `agent-<id>`
name, the accepted AGENTS.md fallback; not renamed); local branch `g018-impl3` tracking
`origin/g018-group-business-profile-onboarding`, created at the checkpoint `70290ff1a383a808962c1079f62bf63408d5bf34`.
`origin/staging` `4afbb3f6ef77cc193811e06dffd879e8201ab18c` is already an ancestor (no merge needed). Bindings
re-verified by `sha256sum` at `70290ff1`: Result `1f5fc64e…633c`, scope `76f1878f…5d11` (43 executable patterns),
release review `3827fad2…9bfc`, audit amendment `1fde7e79…996e`, Plan r5 `4fe4642f…75fa`, R0 addendum
`829cfe4f…4dc84` — all equal. Budget B7 (checkpoint after attempt 4); attempt 1 continues, no reset.

### Successor #3 checkpoint — remaining slice-1 items 3–5 done, item 6 partly (attempt 1, hand-back (c))

**Target reconciliation:** `origin/staging` `4afbb3f6` unchanged and an ancestor of HEAD before each batch
(`g018-scripts/prebatch.sh`, scripts repointed to this worktree); Docker Desktop mounts 13,680 of 100,000 (13%) each
time. All 25 changed paths against `origin/staging` match the 43 patterns (`grep -v -E -f` over the executable
patterns: no unmatched path).

**Changes (all inside the 43 patterns):**

- Item 3 — `access/policy.go`: `group.create` (Organization kind; every valid Organization role; a selected-Group
  context is refused with `role_not_permitted`, so the readable Organization makes it `forbidden`) and
  `group.complete_onboarding` (Group kind; the `group.update` rule: all for Organization owner/admin, own for
  group_owner/group_admin, restricted to the selected Group); `eligible` admits the Organization owner/admin path for
  `group.complete_onboarding`. `matrix_test.go`: two oracle capabilities and rows in the four primary tables
  (`A A A A A A`/`D D X D D D` and `A A D A A D`/`A A X A A D`), the context-difference pin extended by the six
  `group.create` columns, 15 capabilities, **330 primary cells** (from the real loop), six `group.complete_onboarding`
  scope cases (S01/S02/S04/S05/S07/S13), and `group.create` excluded with `organization.read` from the two org_member
  negative loops (`TestOracleUnknownInputs`, `TestOracleCrossOrganization`: org_member's own path).
  `policy_test.go`: both capabilities in `allCapabilities`, the target-gate loops and two `TestPolicyScope` cases.
  `integration_test.go`: IT-MATRIX-1 pin 330.
- Item 4 — `tenancy/store.go`: a separate G018 `const` block with `sqlInsertGroupKeyed` and
  `sqlInsertBusinessProfileKeyed` (the S-G2 read reuses the existing `sqlSelectGroup`, which keeps the Organization
  predicate `TestSQLBoundary` requires — deviation from the predecessor's planned `sqlSelectGroupByID`, which has none).
  `tenancy/onboarding.go` (new): `CreateGroupTx(ctx, tx, org, id, name, slug)` (keyed insert, no RETURNING, then the
  read; empty read → `unavailable`), `CreateBusinessProfileTx(ctx, tx, org, id, c)` (keyed, no RETURNING),
  `LinkGroupBusinessProfileTx` (`sqlInsertAssociation` verbatim); a 23505 on `groups_pkey`/`business_profiles_pkey` →
  `unavailable`. `tenancy/facts.go`: the ascending-id partition (with Lock and a target: others below the target
  `FOR SHARE`, the target `FOR NO KEY UPDATE`, others above `FOR SHARE`; no-target and plain-read shapes unchanged) and
  a `FactQuery.Associations` option (associations of the read Groups without any profile read — needed by the
  new-profile completion and `GetGroupOnboarding`; `ReadFacts` signature unchanged). `tenancy/boundary_test.go`: 15
  production files, +3 package functions (slice 2 adds the two page readers), statements 52 (15 writes).
- Item 5 — `access/onboarding.go` (new): `GroupOnboarding`, `CreateGroup`, `GetGroupOnboarding`,
  `CompleteGroupOnboardingWithNewProfile`, `CompleteGroupOnboardingWithProfile`, `onboardingQuery(g, p)`; ids are
  generated in `access` (UUID v4, the `newObjectID` pattern) and passed to the tenancy writers, so no
  `tenancy.NewGroupID` export is added (deviation: −1 tenancy export). Every operation runs through the existing
  `guard` (so `inContextTx` call sites stay 11 in slice 1). The existing-profile path decides
  `select_for_onboarding` inside `act` on the same locked facts, with `denial`/`markDenied` (recorded under
  `group.onboarding.complete_existing`); the locked state check answers `conflict` field `association_exists`; create
  `slug_conflict` → `conflict` field `slug`. `errors.go`: `association_exists` → `conflict`. `log.go`: the seven
  operation constants. `access/boundary_test.go`: `onboarding.go` in the file list, +5 exported methods, `type
  GroupOnboarding`, and the 00014 SECURITY DEFINER allowance (exactly one, `app.rls_profile_eligible`).
- Item 6 (part) — `access/onboarding_integration_test.go` (new) `TestOnboarding`: ONB-1 (org owner, admin **and member**
  create; exactly one trigger-written `group_owner/membership` binding of the creator; `group.created` then
  `membership.granted` attributed to the creator; no Organization binding change; read-back), ONB-2 (selected-Group
  context → `forbidden`/`role_not_permitted`, recorded in the Organization; **Organization binding removed after Resolve
  → `context_denied`, one `access.denied`/`group.create`/`organization_not_permitted` row persisted through the normal
  runtime path** (IDENTITY shape, actor = the subject; the amendment's requirement); zero state change; input refusals),
  ONB-4 (same slug → `conflict` field `slug`, zero change; same slug in another Organization), ONB-5 (member-created
  Group completed with a new profile, flag off and on, and by the Organization admin; one association;
  `business_profile.created` then `business_profile.linked` by the actor; read-backs), ONB-8 (second completion of a
  complete Group, both branches → `conflict`/`association_exists`, zero change), ONB-9 (unlinked `not_found`, foreign
  `not_found`, member-only `forbidden`/`no_linked_path`, new Group alone `not_found`; four `access.denied` rows; zero
  associations; then attached-Group owner selects pA and the Organization admin selects the unlinked pU; another
  member's Group unreadable and not completable), ONB-13 (member creates, completes new, then selects its own attached
  profile for a second Group; a profile without a path is refused).

**Focused outcomes** (image `g018-core-check:s3a`, project `g018-focus` via `compose run --use-aliases`, no host
port, every Docker command under `flock …/local-harness.lock`; logs under `platform-evidence/g018/`):

| Selector | Result |
|---|---|
| compile: gofmt -l; vet; vet -tags integration (the `compile.sh` packages) | run 1 gofmt `store.go` (const alignment; fixed with a separate block); vet-it runs 1–2: duplicate `asError` (the `mcp_test.go` helper reused), `BusinessProfile.Context.DisplayName`; **final: 0 files, pass, pass** |
| b1 `./internal/tenancy/... ./internal/audit/...` `^(TestServiceRejectsBeforeSQL\|TestPackageBoundary\|TestSQLBoundary\|TestMembershipLockException\|TestAccessBoundary\|TestProductionImporters\|TestRLSBoundary\|TestOracleTables\|TestOracleMatrix\|TestOracleScope\|TestOracleManageable\|TestOracleGates\|TestOracleUnknownInputs\|TestOracleCrossOrganization\|TestToolEligibility\|TestPolicyGates\|TestPolicyScope\|TestCatalogParity\|TestMigrationParity)$` | run 1: 19 FAIL (access file pin; org_member loops asserting `group.create` deny — oracle fixed as above); **run 2: PASS 1,624, FAIL 0** |
| `./internal/tenancy/access/ ^TestOnboarding$` | **PASS 8/8 on the first executed run** |
| `./internal/tenancy/access/ ^TestAccessPolicy$/^IT-MATRIX-1` | PASS (330 cells on DB-loaded facts) |
| `./internal/tenancy/access/ ^TestAccessOrdering$/^(O-13\|O-14\|E-PG-3)` | PASS 5/5 (the changed `ReadFacts` lock path) |

**Cleanup:** `focus.sh down` — containers 0, volumes 0, networks 0, secret files 0; `docker ps -a --filter name=g018`
empty. Images kept: `g018-core-check:s3a` (this checkpoint), `g018-core-check:b1`, `g018-focus-minio`,
`g018-focus-minio-init`.

**Note for slice 2:** `access.GroupPage` already exists (the G014 MCP selectable-Group page, `mcp.go:73`); the r5 §5.1
`GroupPage` for `ListGroups` needs another name (proposed `OnboardingGroupPage`; the planned `ProfilePage` is free).

**Remaining list (precise, for the successor):**

6. Rest of the tests: ONB-3 (hookDB fault inside S-G1 → neither row; deterministic bootstrap race), ONB-6 (fault at
   S-P2 → no orphan profile; audit-store fault), ONB-7 + O-ONB-1..3 (pre-commit revocations on the hookDB seams),
   **O-ONB-D** (combined-facts lock order, both interleavings, no 40P01) and **O-ONB-E** (UPDATE-then-revocation 42501,
   mirror positive, equal-timestamp branch) in `ordering_integration_test.go`; ONB-2(d)/ONB-9(e) fixed-actor runtime SQL
   and reader calls; `rls_expected_test.go` r3 rows (+ `TestRLSIsolation` subcases); attack rows B2/C3; the r5 §6.3
   subcase-selection grep recorded here.
7. b1/b2 with the r5 §6.3 selectors after item 6 (TestRecoverySeed and TestMCPTransport each on a fresh focus DB).
8. Slice 2 (ListGroups/ListBusinessProfiles paged, SuggestGroupSlug, `core/internal/onboarding/api`, server registration,
   logrecord enums incl. the two capabilities, importer/route pins, `inContextTx` count, `onboarding_checks()`), then
   hand back (b) for the draft MR.

### Successor #4 implementer record (attempt 1 continued; verified before the first write, 2026-10-17 UTC)

Task G018, role implement, sole writer (fresh successor #4). Model Opus 5.5; effort "high (requested)", actual value
unknown. Native isolated worktree
`<source-checkout>` (generated `agent-<id>`
name, the accepted AGENTS.md fallback; not renamed); local branch `g018-impl4` tracking
`origin/g018-group-business-profile-onboarding`, created at the checkpoint `ed4eb051fd0b06a20a4992606733f5ae131217a7`.
`origin/staging` `4afbb3f6ef77cc193811e06dffd879e8201ab18c` is already an ancestor (no merge needed). Bindings
re-verified by `sha256sum` at `ed4eb051`: Result `7956a638…176d`, scope `76f1878f…5d11` (43 executable patterns),
Plan r5 `4fe4642f…75fa`, R0 addendum `829cfe4f…4dc84` — all equal. Scripts in `platform-evidence/g018-scripts/`
repointed to this worktree. Budget B7 (checkpoint after attempt 4); attempt 1 continues, no reset.

### Successor #4 checkpoint — item 6 mostly done (attempt 1, hand-back (c))

**Target reconciliation:** `origin/staging` `4afbb3f6` unchanged and an ancestor of HEAD before each batch
(`g018-scripts/prebatch.sh`, repointed to this worktree); Docker Desktop mounts 13,684 of 100,000 (13%). All 28
changed paths against `origin/staging` match the 43 patterns (`grep -v -E -f` over the executable patterns: none
unmatched). Production code unchanged in this round (tests and the Result only).

**Changes (all inside the 43 patterns):**

- `access/integration_test.go` — three hookDB steps for the keyed onboarding inserts: `IG` (S-G1
  `INSERT INTO app.groups (id,`), `IP` (S-P1 `INSERT INTO app.business_profiles (id,`), `IA` (S-P2
  `INSERT INTO app.group_business_profiles`). Previously all three were `?`; no existing test pauses or faults on an
  insert into these tables (the `setFail("?")` users are stored-object inserts) and no insert carries a locking clause,
  so the footprint classification is unchanged (E-PG-1/2 rerun below).
- `access/onboarding_integration_test.go` `TestOnboarding` +4 subtests (12 in total): **ONB-3** (a) fault injected
  after S-G1 (the S-G2 read fails) → `unavailable`, state unchanged; (b) deterministic bootstrap race — creator 1
  pauses after its keyed insert, the uncommitted Group is not observable and no committed Group lacks an owner binding;
  different slugs: creator 2 commits while 1 is paused; same slug: creator 2 waits on the unique key (`waitWaiters`) and
  answers `conflict` field `slug`; each winner has exactly one trigger-written `group_owner/membership` binding.
  **ONB-6** validation (no transaction), fault at S-P2 after S-P1 ran → `unavailable`, no orphan profile, Group
  incomplete; strict audit path — a planted `audit.events` CHECK refusing only `business_profile.linked` → the
  completion fails with no orphan profile, then succeeds once the fault is removed. **ONB-2(d)** fixed-actor runtime SQL
  (owner and member contexts): own owner-binding INSERT on another owner's Group, on its own existing Group and on a
  Group of another Organization → 42501 each; keyed Group INSERT into another Organization, in selected-Group context
  and by a Group-role-only subject → 42501; zero state change. **ONB-9(e)** direct `app.rls_profile_eligible` calls as
  the runtime role: legitimate context true (the only positive); no context false; selected Group of another
  Organization false; unbound selected Group false; Organization binding removed with the gA binding and a selected-gA
  context retained (attack C2 shape) false, and the association INSERT → 42501 in selected-gA, selected-target and
  Organization-only contexts.
- `access/ordering_integration_test.go` `TestAccessOrdering`: `prepared.step` (the order-A pause statement, default
  `W`; existing cases unchanged); cases **O-ONB-1** RemoveGroupMembership vs new-profile completion (ABC,
  `not_found`), **O-ONB-2** ChangeGroupRole admin→member vs new-profile completion (ABC, `forbidden`), **O-ONB-3**
  DisableUser vs existing-profile completion (ABC, `unauthenticated`), **O-ONB-3s** logout-all vs new-profile
  completion (AB, `unauthenticated`), **O-ONB-3p** removal of the last qualifying path vs existing-profile completion
  (ABC, `not_found`); target = the Group's associations and the Organization's profile count (so B/C also prove no
  orphan profile). **O-ONB-D** (g1 < g2 < g3 sorted by id; g3 attached to p3; A: g1 ← p3, B: g2 ← p3; the first pauses
  before S-P2 holding its locks, the second waits, both orders): both commit, no 40P01, g1/g2/g3 each with exactly one
  association to p3. **O-ONB-E** on the attested runtime login: (a) T1 UPDATEs p1 (one row), reader true, trusted
  revocation of u's gA binding commits without blocking, reader false, association to gB → 42501; (b) mirror: T1
  inserts pN after the identical revocation and attaches it to gB → one row; (c) the owner sets `p_eq.created_at` to
  T1's `transaction_timestamp()` (asserted equal inside T1), T1 UPDATEs, revocation, reader false, association → 42501.
- `access/rls_expected_test.go` **oracle r3**: `rlsInsert` and `rlsExpectedInserts()` — 31 attempts (8 Group
  inserts, 3 direct owner-binding inserts, 9 profile inserts, 9 association inserts incl. two same-transaction new
  profiles, 2 association DELETEs) with Basis; r3 header records the sources and limitations. **Deviation
  (independence):** r0–r2 were authored without reading the implementation; these r3 rows were written by the
  implementer, who had read migration 00014 (recorded in the header for Codex). The fixture is unchanged (the
  eligibility positive through an attached Group and the UPDATEd-committed refusal are not expressible in it; they are
  covered by ONB-9/9e, O-ONB-E and CT-ONB-1/R-ONB-1).
- `access/rls_integration_test.go` `TestRLSIsolation`: new subtest **RLS-SCOPE-INSERT** (selected by `^RLS-SCOPE`)
  executing the r3 rows, each in its own rolled-back runtime transaction, state unchanged afterwards; **RLS-ORG**
  updated for version 14: adminA's Group INSERT into its own Organization is now admitted (one row; was 42501 before
  group.create) and a Group INSERT into Organization B → 42501 (new row).

**r5 §6.3 subcase-selection grep** (RLS-ATTEST/RLS-AUTH, `TestRLSPool`, `TestRLSFootprint` statements for INSERT/DELETE
on the four tables or the helper/reader functions): RLS-ATTEST/AUTH 1 hit — a superuser fixture `INSERT INTO
app.role_bindings` on `e.setup` (RLS bypassed, not a policy input); `TestRLSPool` 0; `TestRLSFootprint` 1 hit — it
counts `rls_org_role`/`rls_group_role` calls in plans, whose bodies 00014 does not change (the new reader is reached
only by the association INSERT policy, which that test does not run). **Selection: none of the three rerun.**

**Focused outcomes** (images `g018-core-check:s4a/s4b/s4c` from `infra/docker/core.Dockerfile` target `check`; project
`g018-focus` via `compose run --use-aliases`, no host port; every Docker command under `flock …/local-harness.lock`;
logs `platform-evidence/g018/s4-*.log`):

| Selector | Result |
|---|---|
| compile (gofmt -l; vet; vet -tags integration over the `compile.sh` packages), s4a/s4b/s4c | 0 files, pass, pass each time (first build clean) |
| `./internal/tenancy/access/ ^TestOnboarding$` | run 1: 10 PASS, ONB-2(d) FAIL (test: the state snapshot was taken before the Group-role-only fixture setup; fixed by re-snapshotting); **run 2: PASS 12/12** |
| `./internal/tenancy/access/ ^TestAccessOrdering$/^(O-ONB\|O-13\|O-14\|E-PG-3)` | **PASS on the first run** (12 PASS lines: the parent and 11 subtests — O-ONB-1/2/3/3s/3p with every listed order, O-ONB-D both interleavings, O-ONB-E (a)/(b)/(c), O-13, O-14, E-PG-3a/b) |
| `./internal/tenancy/access/ ^TestRLSIsolation$/^(RLS-SCOPE\|RLS-CTX\|RLS-ORG\|RLS-IMMUT\|RLS-PROJ)` | run 1: RLS-SCOPE-INSERT PASS (31 rows agree), RLS-ORG FAIL (the pre-G018 negative for an own-Organization Group INSERT — the intended version-14 change; updated as above); **run 2: PASS 8/8** |
| `./internal/tenancy/access/ ^TestAccessFootprint$/^(E-PG-1\|E-PG-2)` | PASS (hookDB step change) |
| `./internal/tenancy/access/ ^TestAccessPolicy$/^(IT-CTX-6\|IT-MATRIX-1\|IT-PROFILE-\|IT-FLAG-4\|IT-FACADE-)` | PASS (15 results) |

Not rerun (no changed input in this round): b1 unit selectors (no non-integration file changed; last PASS 1,624 at
s3a), TestOnboardingCatalog, TestRLSCatalog subcases and the version pins (migrations and production code unchanged
since their recorded passes).

**Cleanup:** `focus.sh down` — containers 0, volumes 0, networks 0, secret files 0; `docker ps -a --filter name=g018`
empty. Images kept: `g018-core-check:s4c` (this checkpoint), `s3a`, `b1`, `g018-focus-minio`, `g018-focus-minio-init`
(`s4a`/`s4b` superseded, kept until the supervisor's image cleanup).

**Remaining list (precise, for the successor):**

6. Rest of item 6: **attack rows B2/C3** in `attack_expected_test.go` (r2 header) and `TestIsolationAttacks` (B2: a
   new Group selecting a profile without a path — facade `not_found`/`forbidden` and the runtime association INSERT
   42501, forged creator ignored, direct owner-binding insert 42501; C3: the C2 retained selected-Group context after
   the Organization membership removal — `CompleteGroupOnboardingWithProfile` denied, `rls_profile_eligible` false,
   association INSERT 42501; completed-Group unlink 42501; citations resolved by ISO-CITATIONS); the
   **`TestAccessFootprint` E-PG rows** for create/complete (both) (PG-009; lists in slice 2); `TestAudit` AT-INV/AT-ATOM
   with the new operations if their catalog loop needs real calls.
7. b1 (unit) and b2 with the r5 §6.3 selectors after the attack/footprint additions: attacks
   `^TestIsolationAttacks$/^(ISO-ATTEST|B|C|ISO-CITATIONS)` and `^TestAudit$/^(AT-FMT|AT-ATOM|AT-INV)` (both need the focus
   MinIO, `focus.sh minio` + `obj`); the rows passed here need no rerun unless their inputs change.
8. Slice 2 (unchanged from the successor #3 list; name the list type `OnboardingGroupPage`), then hand back (b).

### Successor #5 implementer record (attempt 1 continued; verified before the first write, 2026-10-17 UTC)

Task G018, role implement, sole writer (fresh successor #5). Model Opus 5.5; effort "high (requested)", actual value
unknown. Native isolated worktree
`<source-checkout>` (generated `agent-<id>`
name, the accepted AGENTS.md fallback; not renamed); local branch `g018-impl5` tracking
`origin/g018-group-business-profile-onboarding`, created at the checkpoint `81a7a4f848183faea319cf0b1ee7cf5a218f4b55`.
`origin/staging` had advanced to `c598c1ac32ab37080e65c5a3b12eaa29fbb44a2e` with management-only changes
(`management/v0/PROGRESS.md`, `management/v0/results/G017-implementation-codex-review.md`); merged without conflict
(merge `38cd25b4`), staging now an ancestor of HEAD. Bindings re-verified by `sha256sum` at `81a7a4f8`: Result
`94812879…f281`, scope `76f1878f…5d11` (43 executable patterns), Plan r5 `4fe4642f…75fa`, R0 addendum `829cfe4f…4dc84`
— all equal. Scripts in `platform-evidence/g018-scripts/` repointed to this worktree (new `inv.sh` for the image
inventory). Superseded images `g018-core-check:s4a/s4b` removed. Budget B7 (checkpoint after attempt 4); attempt 1
continues, no reset.

### Successor #5 round 1 — item 6 done, item 7 (b2 attack/audit/footprint) passed (attempt 1)

**Target reconciliation:** `g018-scripts/prebatch.sh` before the batch: `origin/staging` `c598c1ac` an ancestor of HEAD
`017422dd`; Docker Desktop mounts 13,692 of 100,000 (13%); g018 containers 0. Production code unchanged in this round
(tests and the Result only).

**Changes (inside the 43 patterns):**

- `access/attack_expected_test.go` **oracle r2** (additions only): steps **B2** (11 rows: gMember creates gB2, reads its
  onboarding, then completion with p1 → `forbidden` (member-only path), p2/p3/pB → `not_found`; runtime association
  INSERT → 42501; direct owner-binding INSERT on g2 and a forged owner binding for gMember2 on gB2 → 42501; visible
  associations `{l:g1-p1}` and role bindings `{rb:gMember:orgA, rb:gMember:g1, rb:gMember:gB2}`) and **C3** (11 rows:
  gOwner creates gC3; reader true in selected g1; rolled-back association INSERT (gC3, p1) admitted; completed-Group
  unlink DELETE → 42501; trusted Organization removal; retained Organization-only context → both completions
  `context_denied`; reader false in the retained selected-g1 shape; association INSERT → 42501; visible associations
  empty); `atkEffect.Reader`; AMB-24..26. **Deviation (independence):** as for rls r3, the r2 rows were written by the
  implementer (recorded in the r2 header). Citations `rlsins:<name>` cite rls r3 insert rows.
- `access/attack_integration_test.go`: `create`/`sqlWrite`/`reader` glue, B2 after B1 and C3 after C2, the `reader`
  observation field, and ISO-CITATIONS resolving `rlsins:` against `rlsExpectedInserts()`.
- `access/footprint_integration_test.go`: subtest **"E-PG-2 onboarding writes (G018)"** (PG-009): per subject (s1, s50,
  X admin) in Organization-only context, two creates and both completions; asserts the keyed inserts exactly once, no
  locking clause on an unrecognized statement or an insert, and equal statement counts for s1 and s50 (existing-profile
  completion may differ by the §1.5 straddle lock statement).
- `TestAudit` AT-INV/AT-ATOM: **no change needed** (AT-INV inventories audit actions, not operations; the onboarding
  writes add no action, and ONB-6/ONB-13 already cover the strict path and a persisted new-operation denial).

**E-PG observations (PG-009, statements per write; locked rows):** create 7 (`C U S O OB IG Gt`; U/S/O/OB 1 each, same
for s1, s50 and the X admin); new-profile completion 10 (`C U S O OB GB Gt A IP IA`; GB 1, Gt 1); existing-profile
completion 11 or 12 (`… GB G [Gt] G A P IA`; s1 12 with the straddle lock statement, s50 and the X admin 11; G 51 and
GB 52 locked rows for 50 bindings — subject-scoped, no per-row round trip).

**Focused outcomes** (image `g018-core-check:s5a`; project `g018-focus` via `compose run --use-aliases`, no host port,
focus MinIO; every Docker command under the `local-harness.lock` flock; logs `platform-evidence/g018/s5-*.log`):

| Selector | Result |
|---|---|
| compile (gofmt -l; vet; vet -tags integration over the `compile.sh` packages), s5a | 0 files, pass, pass |
| `./internal/tenancy/access/ ^TestIsolationAttacks$` (obj) | **PASS** first run; `ISO-ORACLE {"cases":104,"disagreements":0}`; B2, C3 and ISO-CITATIONS PASS |
| `./internal/tenancy/access/ ^TestAccessFootprint$/^(E-PG-1\|E-PG-2)` (obj) | **PASS** (E-PG-1, E-PG-2, E-PG-2 onboarding writes) |
| `./internal/tenancy/access/ ^TestAudit$/^(AT-FMT\|AT-ATOM\|AT-INV)` (obj) | AT-FMT/AT-ATOM PASS; AT-INV FAIL — 5 actions never observed (`object.*`, `session.provider_revocation_recorded`) because their producers are other TestAudit subtests the subset excludes (selector limitation, not a defect) |
| `./internal/tenancy/access/ ^TestAudit$` (obj) | **PASS 9/9**; `AUDIT-INV {"actions":36,"observed":36}` |

**Selector deviation (r5 §6.3):** the pinned subcase selectors `^TestIsolationAttacks$/^(ISO-ATTEST|B|C|ISO-CITATIONS)`
and `^TestAudit$/^(AT-FMT|AT-ATOM|AT-INV)` cannot pass alone: the attack oracle's `finish` fails every unchecked
expectation (A/D/E rows) and AT-INV needs actions produced by other TestAudit subtests. Each was run as its whole single
test function instead (still one focused test each; no umbrella). b1 not rerun: no non-integration file changed since
its s3a PASS.

**Cleanup:** `focus.sh down` — containers 0, volumes 0, networks 0, secret files 0.

**Allow-list check:** all 31 changed paths against `origin/staging` `c598c1ac` match the 43 executable patterns
(`grep -v -E -f`: none unmatched). Hand-back (c) after items 6/7 (successor #5 context near its bound); images kept:
`g018-core-check:s5a` (this checkpoint), `s4c`, `s3a`, `b1`, `g018-focus-minio`, `g018-focus-minio-init`.

**Remaining list (precise, for the successor):**

8. Slice 2 (unchanged): paged `ListGroups`/`ListBusinessProfiles` (G014 keyset pattern; list type
   `OnboardingGroupPage`, `ProfilePage` free) and `SuggestGroupSlug`; `core/internal/onboarding/api` (pinned copy of the
   branding/api scope helper, D-5); `cmd/server/main.go` registration; logrecord enums (2 capabilities, 7 operations);
   importer and route pins, the `inContextTx` count and `onboarding_checks()` in `run.sh`; the list E-PG rows (extend
   "E-PG-2 onboarding writes", Plan r5 §1.5: facts + 1 page + 1 associations statement / facts + 1 page statement);
   focused checks; then hand back (b) for the draft MR.

### Successor #6 implementer record (attempt 1 continued, 2026-10-17 UTC)

Task G018, role implement, sole writer (fresh successor #6). Model Opus 5.5; effort "high (requested)", actual value
unknown. Native isolated worktree
`<source-checkout>` (generated `agent-<id>`
name, the accepted AGENTS.md fallback; not renamed); local branch `g018-impl6` tracking
`origin/g018-group-business-profile-onboarding`, created at the checkpoint `42a22cc15201121858fb37db836e7ce51c286bdc`.
`origin/staging` `c598c1ac` already an ancestor (no merge needed). Bindings re-verified by `sha256sum` at `42a22cc1`:
Result `ab35513b…693b`, scope `76f1878f…5d11` (43 executable patterns), Plan r5 `4fe4642f…75fa`, R0 addendum
`829cfe4f…4dc84` — all equal. Scripts in `platform-evidence/g018-scripts/` repointed to this worktree. **Ordering note:**
this record was written after the first slice-2 source edits of this session and before any validation batch (the
bindings above were verified before those edits). Budget B7 (checkpoint after attempt 4); attempt 1 continues.

### Successor #6 round 1 — slice 2 core items done (attempt 1; hand-back (b) for the draft MR)

**Target reconciliation:** `origin/staging` advanced four times during the round with management/governance-only
changes (`b8b6c380`, `a1e1d938`, `61cdd431` and the later G017 DNS/allow-list review commits; no file outside
`management/`, `AGENTS.md`, `CLAUDE.md`, `docs/agents/`, `.agents/skills/`, `skills/`); each was merged without
conflict before the next batch (`g018-scripts/prebatch.sh`: ancestry checked, Docker Desktop mounts 13,704–13,780 of
100,000, g018 containers 0). The 2026-10-17 owner "mechanical correction autonomy" Goal amendment (staging `c5d1b177`)
covers the pin/bookkeeping reconciliations recorded below.

**Changes (all inside the 43 patterns):**

- `tenancy/onboarding.go` + `store.go`: `GroupsPage` (one keyset page statement `sqlFactGroupsPage`, limit+1 rows,
  plus one associations statement; the `slug` filter reuses `sqlSelectBySlug`), `ProfilesPage` (`sqlFactProfilesPage`),
  `MaxOnboardingPage = 100`, and `SuggestGroupSlugTx` (the transaction-scoped suggestion over the RLS-visible slugs).
  `TestPackageBoundary` +3 functions (G018 total **+6**: the plan said +5 — `SuggestGroupSlugTx` is the extra one, needed
  because the raw `Service.SuggestGroupSlug` opens its own transaction), `TestSQLBoundary` 52 → **54** statements
  (15 writes unchanged).
- `access/onboarding.go`: `ListGroups` (`group.read` per row; selected-Group context answers from its facts only),
  `ListBusinessProfiles` (`business_profile.select_for_onboarding` per row on the page rows; summaries; `Next` = last
  scanned id), `SuggestGroupSlug` (`group.create`), `OnboardingGroupPage`/`BusinessProfileSummary`/`ProfilePage`, the
  shared `pageInput` (`invalid_page_size`/`invalid_cursor`) and one list runner `listed` (bind, one read-only
  transaction: liveness, facts, context check, per-row decisions; the same access.decision/access.denied records as
  `guard`). **Deviations (interfaces, for review):** (1) `GetBusinessProfileView` (the `can_update` hint: a second
  `decide` of `business_profile.profile.update` on the same snapshot, recorded as `business_profile.get`) — an 8th new
  exported method, because `GetBusinessProfile`'s signature is used by MCP; (2) exported aliases `GroupID`,
  `BusinessProfileID`, `BusinessContext` and `ParseGroupID`/`ParseBusinessProfileID`, because the importer boundary
  (`TestProductionImporters`, B-MCP-SOURCE) forbids `onboarding/api` from importing `tenancy` (the `type Facts` alias is
  the precedent); (3) `inContextTx` call sites 11 → **12** (one shared list site), not the plan's 17 — slice 1 reused
  `guard`, which adds no site. `TestAccessBoundary` pins updated accordingly; `TestProductionImporters` admits
  `internal/onboarding/api` as an access importer and only `cmd/server` as its importer.
- `core/internal/onboarding/api/http.go` (new): the eight pairs on six paths (§5.3), `begin`/`scope`/`originHost`
  byte-identical to `branding/api`, closed codes (11, incl. `method_not_allowed` and `onboarding_unavailable`), strict
  bodies (`DisallowUnknownFields`, one value), whole-payload budgets `maxBody = 8 << 10` / `maxProfileBody = 96 << 10`
  read before any decode, a closed JSON 405 for the other methods of each served path and a closed JSON 404 below the
  two roots; one `onboarding.request` event per request (operation, outcome, reason, host tag). **Deviations:** no
  `csrf_token` in the list answers (the byte-pinned `begin` does not expose the session info; web takes the CSRF token
  from its identity session, as the branding pages do); a non-JSON content type is `invalid_request` (the plan's closed
  set has no `unsupported_media_type`); a malformed path id is `not_found`.
- `http_test.go` (new): `TestOnboardingHTTP` (every pair: no session → 401, foreign/`http` origin → 403, missing,
  repeated or unknown parameters → 400, non-JSON, query on a mutation, a forged `owner_user_id`, trailing data and
  malformed JSON → 400; every other method → 405 closed JSON; five sub-path probes → 404 closed JSON; only the
  resolver statement ran), `TestOnboardingRefusalMapping`, `TestOnboardingBodyBudgets` (ONB-H9: budgets recomputed
  from the tenancy maxima = 82,780 and 3,487; the max-field all-astral payload raw and fully `\u`-escaped decodes intact,
  validates and reaches the session check; the 4,001-code-point description is refused by `ValidateBusinessContext`
  with `invalid_description`; `maxProfileBody + 1` and `maxBody + 1` → 400, a body exactly at the budget → 401),
  `TestOnboardingCopyIsExact` (D-5), `TestOnboardingNewRefusesIncompleteConfiguration`. **Gap (disclosed):** the
  authenticated HTTP cases (CSRF failure, Group-role-only 403, `org_member` 201) need a real session; as for
  branding/api they are covered at the facade (ONB-1/2/13) and belong to the browser slice (E-ONB-4).
- `access/onboarding_test.go` (new): `TestOnboardingPageBounds` (unit page input and the §1.5 response arithmetic).
- `cmd/server/main.go`: registers the onboarding handler after branding (new digest
  `f4deee748112769925996724bd9b7674e8332348ee8d0046f9da10e4f061165d`, import set + `core/internal/onboarding/api`).
- logrecord: `events.json` (+7 operations and +2 capabilities in `access.decision`, four operation → capability rules,
  the new `onboarding.request` event), `completion.schema.json` (the six paths in both route enums), `logrecord_test.go`
  (catalogue +1 event; accepted records for the six routes, the event and all seven operations; five rejects).
- `run.sh`: `MAIN_GO_G016_SHA`/`MAIN_GO_IMPORTS`, access importers + `core/internal/onboarding/api`, a new
  `CORE_ONBOARDING_ROUTES` pin (B-MCP-SOURCE), B-TENANCY-HTTP (the two served core paths retired from the absence probes,
  20 → 18; a positive closed-refusal pin: GET → 400 `ambiguous_parameter`, POST {} → 403 on `/groups` and 405 on
  `/business-profiles`, no model payload — **deviation:** the plan's "401 without a session" needs a resolvable origin,
  which `TestOnboardingHTTP` pins; the web probes stay until slice 3 serves the web pages), `onboarding_checks()` with
  **B-ONBOARDING-CATALOG** (`it_count` 5) and **B-ONBOARDING-IT** (`it_count` 14) after `branding_checks` and in the
  fallback FAIL list; `required-checks.tsv` +2 rows (121 → 123, header arithmetic). `bash -n` clean. The harness rows
  themselves were not executed (focused selectors only, as instructed).
- `access/onboarding_integration_test.go`: **ONB-10** (a fresh Organization with 7 Groups and 7 profiles, limit 3:
  complete strictly ascending traversal for the owner, the member sees only its Group, owner profiles in 3 pages, a Group
  owner who is also a member elsewhere sees only its attached profile over 2 pages at limit 1, a member none;
  `invalid_page_size`/`invalid_cursor` on both lists; the exact `slug` filter; no foreign row), **ONB-11** (retried
  create with the same or a changed name → `conflict` field `slug`; `ListGroups(slug=)` finds the committed Group;
  retried completion → `conflict`, the committed profile unchanged), **ONB-12** (`can_update` equals the matrix for
  owner/admin/Group owner flag off and on/Group member; hint true, binding revoked, `PUT` refused, no change).
  **Not covered here:** the selected-Group list restriction (facade code path present; no selected-Group fixture in
  `TestOnboarding`) — listed for the next round.
- `access/footprint_integration_test.go` "E-PG-2 onboarding writes": the two list reads, asserting no locked row and
  equal statement counts.

**E-PG observations (PG-009):** `group.list` 9 statements (`C U S O OB GB G ? A`: facts + 1 page + 1 associations),
`business_profile.list` 9 (`C U S O OB GB G A ?`: facts + 1 page), no locked rows, identical for s1, s50 and the X admin.

**Focused outcomes** (images `g018-core-check:s6a/s6b/s6d/s6e`, `g018-vtools-check:s6c`; project `g018-focus` via
`compose run --use-aliases`, no host port, focus MinIO for the footprint row; every Docker command under the
`local-harness.lock` flock; logs `platform-evidence/g018/{compile,unit}-s6*` and `s6-*.log`):

| Selector | Result |
|---|---|
| compile s6a | gofmt flagged one line of `access/onboarding.go` (fixed); vet, vet -tags integration pass |
| compile s6b/s6d/s6e (gofmt; vet incl. `./internal/onboarding/... ./internal/branding/...`; vet -tags integration) | 0 files, pass, pass |
| b1 unit (s6b) `./internal/tenancy/... ./internal/audit/... ./internal/onboarding/... ./internal/branding/...` with the §6.3 selector list + the onboarding/api tests | **PASS**, 711 PASS lines, 0 FAIL (the five onboarding/api tests PASS) |
| logrecord (s6c) `go test ./internal/logrecord/` (whole package, module `validationtools`) | **PASS** 10/10 |
| `./internal/tenancy/access/ ^TestOnboarding$` (s6d) | run 1: 13 PASS, ONB-10 FAIL (test expectation: row-level security already restricts the profile scan to readable rows, so the Group owner's traversal was 1 page; fixture changed to give it a member-only readable profile); **run 2 (s6e): PASS, 14/14 subtests** |
| `./internal/tenancy/access/ ^TestAccessFootprint$/^(E-PG-1\|E-PG-2)` (s6e, obj) | **PASS** (E-PG-1, E-PG-2, E-PG-2 onboarding writes incl. the list reads) |

Not rerun (no changed input): `TestOnboardingCatalog` and the RLS/attack/ordering/audit rows (no migration, policy or
writer change); `TestOnboardingPageBounds` was added after the b1 run (compiled and vetted in s6d/s6e; its first
execution is pending — next b1).

**Cleanup:** `focus.sh down` — containers 0, volumes 0, networks 0, secret files 0; `docker ps -a --filter name=g018`
empty. Images kept: `g018-core-check:s6e` (this checkpoint), `s5a`, `s4c`, `s3a`, `b1`, `g018-vtools-check:s6c`,
`g018-focus-minio`, `g018-focus-minio-init` (`s6a/s6b/s6d` superseded, kept for the supervisor's image cleanup).

**Remaining (precise):** run `TestOnboardingPageBounds` (b1); a selected-Group case for both lists; the authenticated
HTTP cases in the browser slice; web probes of B-TENANCY-HTTP with slice 3; then slices 3–5 per Plan r5 §8.

### Successor #7 implementer record (attempt 1 continued; verified before the first write, 2026-10-17 UTC)

Task G018, role implement, sole writer (fresh successor #7). Model Opus 5.5; effort "high (requested)", actual value
unknown. Native isolated worktree
`<source-checkout>` (generated `agent-<id>`
name, the accepted AGENTS.md fallback; not renamed); local branch `g018-s7` from
`origin/g018-group-business-profile-onboarding` `f32d049b1d9b1160d4eb46ccf4d726247d781683` (draft MR !41 source).
`origin/staging` had advanced to `988ed02094b216f7e310674bb0dc350ed8959523` (management-only: `PROGRESS.md`, G017 DNS
review records); merged without conflict (merge `7e666b79`), staging an ancestor of HEAD. Bindings re-verified by
`sha256sum` at `f32d049b` before any edit: Result `055f2f6b…3383`, scope `76f1878f…5d11` (43 executable patterns), Plan
r5 `4fe4642f…75fa`, R0 addendum `829cfe4f…4dc84`, audit amendment review `1fde7e79…996e` (staging `e8187f57`), execution
release review `3827fad2…bfc`, PROCESS mechanical-corrections rule `c1dbd4ce…49ab` — all as bound. Owner D-1
(superseding) applies: org owner, admin and member may create their own Group. Scripts in `platform-evidence/g018-scripts/`
repointed to this worktree. Budget B7 (checkpoint after attempt 4); attempt 1 continues, no reset. Scope of this
session: slice-2 tail (`TestOnboardingPageBounds` b1 run; selected-Group list case), then slice 3 (Plan r5 §8.3).

### Successor #7 mechanical correction — MR !41 pipeline `2958919772` (head `f32d049b`) checks failure (attempt 1)

Coordinator report, confirmed by the Codex status review (`G018-r0-execution-release-codex-review.md`, "Draft MR checks
failure", staging `3a3c2f39`, merged at `21e2c744` before this batch; ancestry verified; Docker Desktop mounts 13,768 of
100,000):

- **B-CONTRACT-TOOLS (G018, mechanical — PROCESS rule `c1dbd4ce…`):** slice 2 left one line of
  `infra/validation/tools/internal/logrecord/logrecord_test.go` unformatted (a missing space after `"core",` in the
  "access guarded operation without a capability" case; line 318). Fixed by `gofmt` (container `g018-core-check:s7a`,
  stdin, under the `local-harness.lock` flock); assertions unchanged. Focused check: `gofmt -l` on the file — no output;
  the validation-tools **check stage** (`infra/docker/validation-tools.Dockerfile --target check`, image
  `g018-vtools-check:s7b`, `--network none`; gofmt, vet, tests) **exit 0**, `internal/logrecord` ok (log
  `platform-evidence/g018/vtools-s7b.log`).
- **B-UNIT-CHECKS — external known failure, not G018:** `TestDNSSystemChecker` (`core/internal/domains`) returned
  `custom.example.net.` instead of the edge for `long.example.com` — the known G016 DNS adapter concurrency defect owned
  by the held G017 DNS repair. Not touched, not retried; a passing retry would not resolve it.
- **B-REQUIRED-CHECKS:** the cascade of the two above.

### Successor #7 round 1 — slice-2 tail done; slice 3 started (attempt 1; hand-back (c), context bound)

**Target reconciliation:** before each batch `origin/staging` was fetched and verified an ancestor of HEAD (`988ed020`
merged at `7e666b79`; `3a3c2f39` merged at `21e2c744`, both management-only); `bash infra/ci/docker-desktop-mounts.sh`
13,764–13,768 of 100,000 (13%); g018 containers 0. Every Docker command ran under the `local-harness.lock` flock.

**Slice-2 tail (done):**

- `access/onboarding_integration_test.go` ONB-10: **selected-Group list case** for both lists — the Organization owner
  selected in `gs[0]` lists only `gs[0]` (complete, `next` null), nothing after `gs[0]`, and the slug filter of `gs[3]`
  finds nothing; the Group owner selected in its owned `gs[1]` lists `gs[1]` and only `ps[1]`; selected in `gs[0]`
  (member there) it lists `gs[0]` and no selectable profile. No production change.

| Selector (image) | Result |
|---|---|
| compile `s7a` (gofmt -l; vet; vet -tags integration over the `compile.sh` packages) | 0 files, pass, pass |
| b1 `./internal/tenancy/access/ ^TestOnboardingPageBounds$` (s7a) | **PASS** (its first execution) |
| `./internal/tenancy/access/ ^TestOnboarding$` (s7a, `g018-focus`) | **PASS first run**, 14/14 subtests incl. the extended ONB-10 |
| validation-tools check stage (`g018-vtools-check:s7b`) | **PASS** (see the mechanical correction above) |

Cleanup: `focus.sh down` — containers 0, volumes 0, networks 0, secret files 0.

**Slice 3 (started; Plan r5 §1.6, §5.4, §8.3):**

- `web/lib/session.ts`: `readForm(request, options?: { maxBytes })` and `readBounded(request, maxBytes =
  MAX_FORM_BYTES)` — absent options keep the 16 KiB default for every existing caller; `session.test.mjs` +1 case
  (default unchanged for `undefined`/`{}`/`{maxBytes: undefined}`; 96 KiB honoured at the boundary; a smaller explicit
  budget refused).
- `web/lib/onboarding-client.ts` (new, not yet wired to pages): the eight calls `listGroups`, `suggestSlug`,
  `createGroup`, `getGroup`, `completeOnboarding`, `listProfiles`, `getProfile`, `updateProfile` with strict shapes
  (exact keys, UUIDs, slug grammar, code-point maxima, derived `complete` ⇔ profile attached, strictly ascending page
  ids, `next` not below the last row, slug-filtered list ≤ 1 matching row) and per-endpoint success bounds (§1.5: Group
  row 1,500, list 150,227; profile row 3,200, list 320,227; single profile 42 KiB; refusals 64 KiB); closed refusal
  reading (`{"error"}` or, for `invalid_request`/`conflict`, `{"error","field"}`); `PROFILE_FORM_BYTES = 96 KiB` with
  its derivation constants; pure form helpers (`readProfileInput` with CRLF → LF in the description, `fieldError`,
  `fieldErrorId`, `FIELD_LABELS`) and the route status pages (`createConflictPage`, `unknownOutcomePage`,
  `refusalPage`, `outcomeUnknown`) whose conflict/unknown wording never claims creation.
- b3 (`g018-web-check:s7c/s7d`, `web.sh`): s7c lint failed on one unused loop variable in `codePoints` (fixed); **s7d
  `npm run check` (lint, typecheck, unit) exit 0; `npm run test:unit` 165/165 pass.**

**Allow-list:** all 42 changed paths against `origin/staging` `3a3c2f39` match the 43 executable patterns (`grep -v -E
-f`: none unmatched).

**Design decisions taken for the remainder of slice 3 (for the successor; within Plan r5, flagged for review):**
(1) shell area links as an `AppShell` "Workspace areas" navigation rendered only on a tenant host for a signed-in
member (`brand.indicator.kind === "member"`), so the generic frame and the Primary anchor navigation (S-NAV's three `#`
links) stay unchanged — the Plan's "add an `areas` slot" option; (2) pages read the session (own core call, CSRF token
— core's list answers carry none, successor #6 deviation) and call onboarding with `lookupContext`, as the branding
page; (3) a validation refusal redirects back with `?error=<field code>` and the page renders `fieldError` next to the
field (`aria-describedby`, `aria-invalid`, focus) without echoing submitted values (create keeps the validated name
when the refusal is the slug); completion conflict redirects to the Group page with a closed status that shows the
current state; create conflict re-reads `GET /groups?slug=` in the route and answers `createConflictPage`; no
`maxLength` on profile inputs (UTF-16 units would refuse valid astral values; core validates code points).

**Remaining (precise):** slice 3 — `web/components/ui/{input,textarea}.tsx` (kit `input.tsx` sha256 `2db8906d…07a3`,
`textarea.tsx` `62e91211…5c2a`; replace the kit's `bg-surface`, which is not a token here, with `bg-background` and the
branding focus outline), `web/components/onboarding/*.tsx` (frame/request helper, profile fields, state badge), the nine
pages/routes of §1.6, `app-shell.tsx` areas, `onboarding-client.test.mjs` and `onboarding-forms.test.mjs`, `run.sh`
(`WEB_ROUTES` +9 sorted; `TENANCY_PROBE_WEB` −`/groups` −`/business-profiles` (15 → 13) with an optional positive GET
pin; B-WEB-COMPONENTS record text "seven" → "nine"), `lib.sh` (`UI_FILES` + `input.tsx textarea.tsx`, `UI_SLOTS` +
`input textarea`), `profile-paths.tsv` (`foundation dir web/app/groups/`, `web/app/business-profiles/`); b1 (no core
change expected), b3. Then slices 4–5 per Plan r5 §8; the authenticated HTTP cases belong to slice 4 (E-ONB-4).

### Successor #8 implementer record (attempt 1 continued; verified before the first write, 2026-10-17 UTC)

Task G018, role implement, sole writer (fresh successor #8). Model Opus 5.5; effort "high (requested)", actual value
unknown. Native isolated worktree
`<source-checkout>` (generated `agent-<id>`
name, the accepted AGENTS.md fallback; not renamed); local branch `g018-s8` from
`origin/g018-group-business-profile-onboarding` `1a6c134389d7e91777d8eed781fd59f1a76d18e5` (draft MR !41 source).
`origin/staging` `3a3c2f3928467db35c947b6795c58089c789dfbd` already an ancestor of HEAD (no merge needed). Bindings
re-verified by `sha256sum` at `1a6c1343` before any edit: Result `35fe0dd3…0c39`, scope `76f1878f…5d11` (43 executable
patterns), Plan r5 `4fe4642f…75fa`, R0 addendum `829cfe4f…4dc84`, execution release review `1c0a1cfe…c123` (staging
`3a3c2f39`), audit amendment review `1fde7e79…996e`, PROCESS mechanical-corrections rule `c1dbd4ce…49ab` — all as bound.
Owner D-1 (superseding) applies: org owner, admin and member may create their own Group. Kit sources re-verified:
`input.tsx` `2db8906d…07a3`, `textarea.tsx` `62e91211…5c2a`. Repository Vestive skill loaded
(`web/.claude/skills/vestive-design`, kit §8–§9 adaptations: field boundary on `border-input`, app focus outline).
Budget B7 (checkpoint after attempt 4); attempt 1 continues, no reset. Scope of this session: the rest of slice 3 per
successor #7's "Remaining" list and design decisions (1)–(3).

### Successor #8 round 1 — slice 3 web done; MR !41 B-WEB-BOUNDARY mechanical correction (attempt 1)

**Target reconciliation:** before each batch `prebatch.sh` fetched `origin/staging` (`3a3c2f39`, unchanged all round)
and verified it an ancestor of HEAD; `bash infra/ci/docker-desktop-mounts.sh` 13,768 of 100,000 (13%) every time;
g018 containers 0. Every Docker/harness command ran under the `local-harness.lock` flock (no refusal). Scripts in
`platform-evidence/g018-scripts/` repointed to this worktree.

**MR !41 pipeline `2958938657` (head `1a6c1343`) — B-WEB-BOUNDARY (G018, mechanical — PROCESS rule `c1dbd4ce…`):**
`lib/onboarding-client.ts` imports `./request-context`, outside the B-WEB-BOUNDARY request-context allow-list. Plan
r5 does not inventory this list (its §3 names only `run.sh` `WEB_ROUTES`, probe lists and pins; `run.sh` is an
allowed path) — an inventory gap of the same class as G017's branding entries, reconciled here. The allow-list
(`run.sh` `static_web_boundary`) now names exactly `lib/onboarding-client.ts`, `app/groups/create/route.ts`,
`app/groups/\[id\]/onboarding/complete/route.ts` and `app/business-profiles/\[id\]/update/route.ts` (no globs); the
onboarding pages and components do not import it (the frame types its lookup context from `lookupContext`). The
FAIL log text, the comment and the PASS record text name the G018 client and three routes; the routes-differ log
and the PASS route list name the nine G018 routes.

**Slice 3 (Plan r5 §1.6, §5.4, §8.3) — done:**

- `components/ui/input.tsx`, `textarea.tsx`: adapted from the kit (lineage header with the kit SHA-256s); `bg-background`
  for the kit's surface class (not a token here), resting boundary `border-input` (the 3:1 role; kit §9.9), the
  shell's focus outline (kit §9.7 pattern), `font-sans`, `data-slot` `input`/`textarea`, `aria-invalid` boundary.
- `components/onboarding/`: `onboarding-frame.tsx` (tenant-host-only frame and session read, fresh lookup request ID
  per onboarding call, fixed refusal pages: `not_found`/`origin_rejected` → 404, `not_available`/`forbidden` → "Not
  available", anything else → "unavailable"; runtime configuration passed in by each page, since B-WEB-BOUNDARY bars
  components from `@/lib/config`), `form-field.tsx` (native label, hint, error by `aria-describedby`, `aria-invalid`,
  focus on the refused field, no `maxLength`), `profile-fields.tsx` (the five fields), `state-badge.tsx`.
- The nine routes of §1.6: `/groups` (keyset page, state, Open / Resume onboarding, New Group, First/Next page);
  `/groups/new` (step 1 GET name; step 2 core's suggested editable slug, name echoed only after core accepted it);
  `POST /groups/create` (16 KiB default; field refusal → `?error=<code>` (+ `name` for a slug refusal); conflict →
  slug-filtered re-read with a fresh lookup ID → `createConflictPage`; empty slug conflict → fixed "Address unavailable"
  page; possibly committed → `unknownOutcomePage`); `/groups/<uuid>` (status `completed`/`already_complete`, profile
  summary + Edit/View per `can_update`); `/groups/<uuid>/onboarding` (create-new form and select-existing radio form from
  one `ListBusinessProfiles` page, Next/First; completed Group shows its state, no forms); `POST
  …/onboarding/complete` and `POST /business-profiles/<uuid>/update` (`PROFILE_FORM_BYTES`; conflict → Group page
  `already_complete`; profile field refusal → `?error=<code>`; `forbidden` → "Not permitted"); `/business-profiles` and
  `/business-profiles/<uuid>` (editor or read-only view per `can_update`; last-save-wins disclosure; no associations).
- `app-shell.tsx`: optional `area` prop; a "Workspace areas" navigation (Groups, Business Profiles; `aria-current` on
  the page's area) after the header only when `brand.indicator.kind === "member"` (decision (1)); generic frame and
  Primary navigation unchanged.
- Tests: `onboarding-client.test.mjs` (exact outbound requests of all eight calls, strict shapes, slug filter, bounds
  arithmetic and over-bound answers, closed refusals, transport) and `onboarding-forms.test.mjs` (budget arithmetic;
  max-field all-astral submission through both profile routes accepted and forwarded intact; `PROFILE_FORM_BYTES + 1`
  → 400 with nothing forwarded; create keeps 16 KiB; helper and route outcomes) — 28 new cases.
- Harness: `run.sh` `WEB_ROUTES` 22 → 31 (C-sorted, as `find` lists them); `TENANCY_PROBE_WEB` 15 → 13 (−`/groups`,
  −`/business-profiles`; the optional positive web GET pin not added); B-WEB-COMPONENTS text "seven" → "nine";
  `lib.sh` `UI_FILES` + `input.tsx textarea.tsx`, `UI_SLOTS` + `input textarea`; `profile-paths.tsv` `foundation dir`
  `web/app/groups/`, `web/app/business-profiles/`.

| Batch (image / command) | Result |
|---|---|
| b3 `web.sh s8a` (`g018-web-check:s8a`: `npm run check`) | **FAIL** lint: `no-html-link-for-pages` on `/groups/new` (rule mistakes it for `/groups/[id]`), `react/display-name` on the frame closure |
| b3 `web.sh s8b` | lint, typecheck pass; unit **189/193** — 4 test-expectation errors in the new client test (own traceparent span; an all-digit UUID upper-cased is unchanged) |
| b3 `web.sh s8c` | **PASS** `npm run check` exit 0; `npm run test:unit` **193/193** |
| `run.sh --base 3a3c2f39 --allow-file …G018-scope-allowlist.txt lint-web` (report `g018/lintweb-s8a`) | B-WEB-BOUNDARY FAIL (component imported `@/lib/config`), B-WEB-TOKENS FAIL (comment named the kit's `bg-surface`); COMPONENTS, DESIGN, ASSET-CONTROLS, SCOPE (62 paths) PASS |
| same, `lintweb-s8b` after the fixes | **B-WEB-BOUNDARY PASS**, B-WEB-TOKENS, B-WEB-COMPONENTS, B-WEB-DESIGN, B-WEB-ASSET-CONTROLS, B-SCOPE (62 paths), B-DOCKER-ONLY PASS; B-SECRET-CANARY UNMET → B-REQUIRED-CHECKS FAIL, the known focused-suite artifact (no secrets generated without `lint`), not a check failure |
| b3 `web.sh s8d` (retest after the fixes) | **PASS** `npm run check` exit 0; unit **193/193** |

b1 not run: no core change this round. Known external: `TestDNSSystemChecker` (B-UNIT-CHECKS) not touched or run.
Cleanup: images `g018-web-check:s8a`–`s8d` removed; containers 0. **Allow-list:** 62 changed paths against
`origin/staging` `3a3c2f39` all match the 43 patterns (`grep -v -E -f`: none unmatched; B-SCOPE PASS).

**Decisions for Codex (within Plan r5; flagged):** (a) `web/THIRD_PARTY_NOTICES.md` is outside the allow-list: the
input/textarea headers point to its shadcn/ui MIT licence text, but its file list still names only button and
dialog — an allow-list addition or a follow-up is needed for the list. (b) The Workspace areas navigation shows on
every tenant page for a member (including branding and the tenant home), not only onboarding pages; slice-4 browser
pins of member pages (landmarks, navigation) must account for it. (c) One `eslint-disable-next-line
@next/next/no-html-link-for-pages` with its reason (precedent: two `no-img-element` disables). (d) A
`not_available`/`forbidden` onboarding read renders a "Not available" page instead of the branding page's 404, because
any Organization member may use onboarding (D-1). (e) Plan r5 §5.3's "16 core and 10 web absence probes" predates R0;
the probe list now holds 13 web paths, as instructed.

**Still open:** slice 3 docs (`docs/foundation/WEB-SHELL.md` lineage of input/textarea and the areas navigation)
belong with slice 5; slices 4–5 per Plan r5 §8 (browser `onboarding.spec.ts`, shell/mcp spec and support pins,
authenticated HTTP cases E-ONB-4, B-TENANCY-HTTP run, docs, the full local qualification).

### Successor #9 implementer record (attempt 1 continued; verified before the first write, 2026-10-17 UTC)

Task G018, role implement, sole writer (fresh successor #9). Model Opus 5.5; effort "high (requested)", actual value
unknown. Native isolated worktree
`<source-checkout>` (generated `agent-<id>`
name, the accepted AGENTS.md fallback; not renamed); local branch `g018-s9` from
`origin/g018-group-business-profile-onboarding` `763f850800cbbef154baf85e45dff813056f5735` (draft MR !41 source).
`origin/staging` `3a3c2f3928467db35c947b6795c58089c789dfbd` already an ancestor of HEAD (no merge needed). Bindings
re-verified by `sha256sum` at `763f8508` before any edit: Result `aca4f5d5…fa30`, scope `76f1878f…5d11` (43 executable
patterns), Plan r5 `4fe4642f…75fa`, R0 addendum `829cfe4f…4dc84`, execution release review `1c0a1cfe…c123` (staging
`3a3c2f39`), audit amendment review `1fde7e79…996e`, PROCESS mechanical-corrections rule `c1dbd4ce…49ab` — all as bound.
Owner D-1 (superseding) applies: org owner, admin and member may create their own Group. Successor #8's decisions
(a)–(e) carried: `web/THIRD_PARTY_NOTICES.md` left untouched (pending Codex); browser landmark/navigation pins allow
for the member Workspace areas navigation on every tenant page. Scripts in `platform-evidence/g018-scripts/` repointed
to this worktree. Budget B7 (checkpoint after attempt 4); attempt 1 continues, no reset. Scope of this session: slice 4
(Plan r5 §8 item 4, §6.2 E-ONB-1..7, §6.3 b4): seed additions, `onboarding.spec.ts`, Playwright projects, browser
support/matrix pins, S-LANDMARKS/S-NAV, the B-TENANCY-HTTP web probes.

**Successor #9 rebind — scope amendment 01 (staging `ead5d2ce`).** At the first safe boundary (no batch running; no
check had started this session) `origin/staging` `ead5d2ce398b9f27fa202af817f4bc47f5b6c647` (management-only: PROGRESS,
execution release review, archived allow-list, new allow-list) was merged without conflict at `6e9dace9`; staging an
ancestor of HEAD. Re-hashed: scope `3287ebca1765b787a930e6c894bebbc7a9e339cbe35d3ddd2f92ff5897c76b90` (**44**
executable patterns; the prior 43, `76f1878f…5d11`, archived at `G018-scope-allowlist-before-amendment01.txt`),
execution release review `fcfc0a1541dae5874fc2cc5075b119c4a9544a82ee5457fb3ec7638c9025556f` (appended "Scope
amendment01 and slice3 reconciliation" dispositions read), Plan r5 `4fe4642f…75fa` unchanged. Per the amendment,
`web/THIRD_PARTY_NOTICES.md` now names `components/ui/input.tsx` and `textarea.tsx` (G018) in the existing shadcn/ui
attribution paragraph; MIT copyright, licence text and unrelated notices unchanged; no notice-only test (diff review).
Decisions (b)–(e) and the request-context entries accepted; attempt 1/B7/checkpoint 4 unchanged; not a replan.

### Successor #9 round 1 — slice 4 browser coverage done (attempt 1; hand-back (c) for R-B review)

**Target reconciliation:** before each batch `prebatch.sh` fetched `origin/staging` and verified it an ancestor of HEAD
(`ead5d2ce` merged at `6e9dace9`, see the rebind above; `9cf9beed` — management-only G019 plan review — merged at
`31f4747d` before the harness lint batch; bindings re-hashed unchanged: scope `3287ebca…6b90`, release review
`fcfc0a15…556f`). `bash infra/ci/docker-desktop-mounts.sh` 13,768–13,840 of 100,000 (13%). Every Docker/harness command
ran under the `local-harness.lock` flock (no refusal). G017's running preview `g017-b4` (standard+100 ports, 543) was
not touched: the disposable stack `g018-b4` (`g018-scripts/b4.sh`, the G017 b4 script with project `g018-b4`, its
published ports moved to 2xxxx and the proxy on a random loopback port) ran IDENTITY.md's identity:start/migrate, the
harness's mcp_seed step and the identity:browser command narrowed to named projects.

**Slice 4 (Plan r5 §8 item 4, §6.2 E-ONB-1..7, §6.3 b4) — done:**

- `infra/identity/mcp-seed.sql` (`5b30dbf3…f84a`), exact rows: **frank `org_admin` of A** (replacing his `org_member`
  row of A; the E-ONB-1 creator), **bob `group_owner` of research**, **dave `group_member` of a new Group planning (C)
  with no Organization C binding** (the Group-role-only subject), Group planning (C); profiles Research Labs (A,
  attached to research), Design Studio (A, attached to design), Operations Partner (B, unattached); 2 associations.
  Carol stays A's owner; alice stays a plain member of A (E-ONB-4's plain-member creator).
- `web/tests/browser/onboarding.spec.ts` (`b347d11c…3179`, new): E-ONB-1 (admin: suggested address contains the
  name, edited address, a core-refused website associated by `aria-describedby`/`aria-invalid`/focus without echo,
  new-profile completion, Edit link), E-ONB-2 (owner: incomplete state without false completion, Resume from the list,
  only Research Labs selectable, completion, completed onboarding page without forms), E-ONB-3 (both entry points reach
  the same record; dave read-only: View link, no textbox, forged update `403 Not permitted`, profile unchanged, not in
  his list), E-ONB-4 (D-1 plain member creates/completes its own Group and sees neither research nor design; another
  member's Group/onboarding/profile and absent ids all `404` without object data; forged completion of design refused
  with nothing changed; forged-CSRF create `403`, no Group; Group-role-only dave refused at C sign-in, no session, no
  New Group, no areas nav, direct create `401`, and the probe address is then free in C), E-ONB-5 (keyboard-only E-ONB-1
  plus axe WCAG 2.2 AA and no horizontal scroll at 1280/320 on the sign-in-first, list, both New Group steps,
  onboarding, field-error, completed Group, editor and list pages), E-ONB-6 (same step-2 form in two tabs: one Group;
  the second answer is "Group address already in use", "Nothing was created", link to the created Group), E-ONB-7
  (dave's MCP connection rows unchanged across the areas links, Group, profile and New Group step 2 — no Group created,
  because B-MCP-E2E pins dave's selectable Groups of A to research alone), and member-page **S-LANDMARKS**/**S-NAV**
  (decision (b): banner, Primary and Workspace areas navigation, main, contentinfo, `aria-current` only on the page's
  area, none on `/account`; heading order; Primary anchors and both area links by keyboard).
- `web/playwright.config.ts`: identity projects `onboarding-light`/`onboarding-dark` beside the mcp projects.
- `web/tests/browser/mcp.spec.ts`: M-NOAUTH no areas nav for anonymous; M-SETUP the member's areas nav (Groups,
  Business Profiles, no `aria-current`). `web/tests/browser/shell.spec.ts`: S-LANDMARKS no areas nav on the frame.
- `infra/validation/lib.sh` `BROWSER_MATRIX_IDENTITY_E2E` + `onboarding-light onboarding-dark|E-ONB-1..7 S-LANDMARKS
  S-NAV S-ERRORS`; `run.sh` `identity_e2e` appends `--project=onboarding-light --project=onboarding-dark` to the
  IDENTITY.md identity:browser line (PASS text names them).
- `web/THIRD_PARTY_NOTICES.md`: amendment 01 (above).

| Batch (image / command) | Result |
|---|---|
| b3 `web.sh s9a` (`npm run check`; covers the spec's lint/typecheck) | **PASS** exit 0; unit 193/193 |
| b4 `b4.sh test b4-s9a` (onboarding-light, onboarding-dark; all tests) | **18 passed** (9 × 2), first run |
| `PROJECTS=mcp-light,mcp-dark,domain-platform,domain-custom,identity-light b4.sh test b4-s9b 'M-NOAUTH\|M-NONMEMBER\|M-SETUP\|M-A11Y\|E-BRD\|E-ATTR'` (seed and nav-pin impact) | 14 passed, **1 failed: E-BRD-5** scrollWidth 324 at 320px |
| diagnostic rerun `b4-s9c` E-BRD-5 alone; read-only probe `probe/overflow.cjs` (carol, 320px) | FAIL reproduced; overflow is G017's "Reset to platform defaults" submit (left 20, width 304) — **external**, see below |
| focused B-TENANCY-HTTP `g018-scripts/tenancy-http.sh` (run.sh's arrays and checks on `g018-b4`) | **PASS** 66 probes: 18 core ×2, the 2 served core paths ×2 (400 `ambiguous_parameter` / 403 / 405, no model payload), 13 web ×2 → 404 |
| `run.sh --base 9cf9beed --allow-file …G018-scope-allowlist.txt lint lint-web` (`g018/lint-s9a`) | **exit 0; B-REQUIRED-CHECKS PASS (17 IDs)**; B-SCOPE 68 paths, B-SHELL, B-CI-CONFIG, B-WEB-BOUNDARY/-COMPONENTS/-TOKENS/-DESIGN/-ASSET-CONTROLS, B-SECRET-CANARY/-SOURCE PASS |
| b4 `b4-s9d` `--grep E-ONB` after the E-ONB-7 change; b3 `web.sh s9b` | **14 passed**; **PASS** unit 193/193 |

Known external failures, not touched: **E-BRD-5** — G017's Reset submit overflow, diagnosed and fixed by G017 on
`origin/g017-staging-repair` (`0a0fc999`, `branding-form.tsx`), not yet in staging (G017 Result row "B-IDENTITY-E2E
(E-BRD-5 scrollWidth 324)"); not caused by G018 (the overflowing element's width is intrinsic; G018 changes no
branding file). `TestDNSSystemChecker` (B-UNIT-CHECKS) not run. Cleanup: `b4.sh down` (containers 0, volumes 0), the
11 images this session built removed (`rmi-s9.sh`); g018 containers/volumes/networks 0. **Allow-list:** 68 changed
paths against `origin/staging` `9cf9beed` all match the 44 patterns (none unmatched; B-SCOPE PASS).

**Decisions for Codex (flagged):** (f) the admin actor is **frank, not alice** (R0 addendum §B said alice
`org_admin`): `org_admin` holds `organization.manage_branding`, and G017's E-BRD-5 pins alice as the non-manager of A,
so alice stays a plain member and serves as the E-ONB-4 member. (g) The Group-role-only browser subject cannot hold a
tenant session (D-2(a) refuses its sign-in), so E-ONB-4 covers it as refused sign-in + no create action + direct POST
`401` + the address still free; the authenticated Group-role-only negative stays core-level (ONB-2, ONB-H). (h) The
onboarding projects join the harness by `run.sh` appending them to the IDENTITY.md identity:browser line, because
`docs/foundation/IDENTITY.md` is outside the allow-list; slice 5 may instead add them to that docs line if Codex adds
the path. (i) `web/tests/browser/support/{assets,modes}.ts` unchanged: the seeded users are read in the spec, as in
`mcp.spec.ts`/`identity.spec.ts`. (j) Plan r5's b4 selector `--project=shell-light/-dark` cannot run under
`IDENTITY_E2E=1` (shell projects are fixture-only), so member-page S-LANDMARKS/S-NAV live in the onboarding projects and
the shell spec gets only the frame's no-areas pin; the shell projects themselves were not run this round (one added
assertion; they belong to the final affected coverage).

**Still open:** R-B review (Codex); slice 5 per Plan r5 §8 (docs incl. WEB-SHELL lineage and areas navigation, M1
operator preview, reconcile, required final affected local coverage incl. the full identity profile with the shell
projects and B-MCP-E2E, Result, MR ready after Codex pre-integration acceptance). E-BRD-5 will pass only once G017's
staging repair is in staging.

### Successor #10 implementer record (attempt 1 continued; verified before the first write, 2026-10-17 UTC)

Task G018, role implement, sole writer (fresh successor #10, replacing #9 under the context policy). Model Opus 5.5;
effort "high (requested)", actual value unknown. Native isolated worktree
`<source-checkout>` (generated `agent-<id>`
name, the accepted AGENTS.md fallback; not renamed); local branch `g018-s10` from
`origin/g018-group-business-profile-onboarding` `97fd8c5a14efe3e60acac4e623e0432d491a6787` (MR !41 source).
**Scope-02 rebind:** at the first safe boundary (no batch running or started this session) `origin/staging`
`f69531963b60f74ea41323ce177f045ecc0aeedb` (management-only: SUPERVISOR-HANDOFF, PROGRESS, imported Result, archived
allow-list, new allow-list, R-B review) was merged without conflict at `4eb051b8`. Bindings re-verified by `sha256sum`
after the merge and before any edit: R-B review `194c1390…1a1a` (CORRECTION LOOP; read in full), scope
`747816831f313e2bea9c63fb9d2110c98270920e9d7b83727ab155c210d56e65` (**45** executable patterns; adds only
`^docs/foundation/IDENTITY[.]md$`; prior 44 archived as `G018-scope-allowlist-before-amendment02.txt`), Result
`f7334d00…c0e0` (the imported bytes, unchanged by the merge), Plan r5 `4fe4642f…75fa`, R0 addendum `829cfe4f…4dc84`,
execution release review `fcfc0a15…556f` (current on staging), audit amendment review `1fde7e79…996e`, PROCESS
mechanical-corrections rule `c1dbd4ce…49ab` — all as bound. R-B dispositions on #9's decisions (f)–(j) carried as
accepted/reconciled (frank admin / alice plain member; Group-only refused sign-in as the browser boundary, not an
authenticated 403 claim; onboarding-project landmarks; E-ONB-7 wording limited to setup-row metadata; THIRD_PARTY
notice). Budget **B7, checkpoint after attempt 4**; attempt 1 continues as a correction round, no reset or enlargement.
Scope of this session: the R-B blocker (completion notice gated on core state; E-ONB-2 extension), scope-02 relocation
of the onboarding projects into IDENTITY.md, slice-5 docs and M1 preview preparation; no full identity/ops/profile
rerun (not released). Known external failures noted only: `TestDNSSystemChecker` (held G017 DNS repair) and E-BRD-5
(G017 fix `0a0fc999` not yet on staging).

### Successor #10 round 1 — R-B correction, scope-02 relocation, slice-5 docs and M1 preparation (attempt 1; hand-back (c))

**Target reconciliation:** `prebatch.sh` (fetch, staging ancestry, `bash infra/ci/docker-desktop-mounts.sh`) before
each batch; every Docker/harness command under the `local-harness.lock` flock (no refusal). Staging `f6953196` merged
at `4eb051b8` (the scope-02 rebind, above). Before the lint batch staging had advanced to
`1bb052bd6d19f7229d550f9a4af4a6f9ba4d5cac` (management-only: G019 plan review r1, PROGRESS, SUPERVISOR-HANDOFF); merged
without conflict at `294cd332`; scope `74781683…6e56` and R-B review `194c1390…1a1a` re-hashed unchanged. Mount table
13,804–13,868 of 100,000 (13%).

**1. R-B BLOCKER fixed** (`d170009e`). `web/app/groups/[id]/page.tsx`: the closed `STATUS` notices are selected only
when core's `group.complete` is true (`group.complete && status !== undefined && Object.hasOwn(STATUS, status)`), with
a comment citing Spec R2/AC5; no new flash framework, API or copy. `web/tests/browser/onboarding.spec.ts` E-ONB-2:
while bob's new Group is incomplete it visits `?status=completed` and `?status=already_complete` and asserts
`group-state` "Incomplete", the "Resume onboarding" link, no `status-message`, no completion/already-complete text and
no `group-profile`; after the real completion it now also asserts the `completed` notice text (the gate does not
suppress the truthful notice). Not changed (outside R-B): the Business Profile page's `?status=updated` notice is
likewise URL-driven; it claims no Group completion — noted for Codex only.

**2. Scope 02** (`eb84747e`). `docs/foundation/IDENTITY.md` `identity:browser` now ends `--project=mcp-dark
--project=onboarding-light --project=onboarding-dark`; `run.sh` `identity_e2e` no longer appends them (two lines
removed; the B-IDENTITY-E2E PASS text still names the projects). The resulting executable line is byte-identical to the
former docs line plus append. Verified statically with `docs_block`'s awk on the new file: one line, each of the eight
`--project=` values exactly once; no other append remains in `run.sh`. IDENTITY.md also gains the G018 seed rows in
"Test overlay", a **Browser projects** bullet (the four project pairs and the E-ONB-1..7/S-LANDMARKS/S-NAV scenarios)
and `onboarding.spec.ts` in the L2 emulator evidence row. No auth/provider/recovery contract text changed.

**3. Slice 5 docs** (`db8e4dfb`; Plan r5 §8 item 5). Drafted by one bounded general-purpose delegate (model inherited,
Opus 5.5; effort unknown) in this worktree with exclusive ownership of the ten doc files and no VCS/Docker/test
actions; reviewed and integrated here (WEB-SHELL, SECURITY and PERFORMANCE diffs read in full; the `input`/`textarea`
kit SHA-256 lineage values checked against the files' headers and the skill kit; nine `components/ui` files). TENANCY
(`## Group onboarding (G018)` and model/privilege/boundary notes), AUTHORIZATION (15 capabilities, `### Guarded Group
onboarding operations (G018)`, 36 operations, ordering step 6, attack B2/C3, mutation-path inventory),
DATABASE-ISOLATION (version 14), AUDIT (36 operations incl. G017's five previously unlisted; G018 paragraph),
MIGRATIONS (version 14), WEB-SHELL (`### Form controls (G018)` lineage, onboarding projects in the browser-projects
table, `## Group onboarding pages (G018)` incl. the core-gated notices and the Workspace areas navigation), FOLLOW-UP
(STAGE-ONBOARDING item, G018 deferrals), RECOVERY (version 14; drill from the hosted staging gate), SECURITY (§5 runtime
row; "Tenancy mutations over HTTP (G018)"), PERFORMANCE (`## PG-009`, counts from the E-PG-2 onboarding rows;
plans/latency unknown). Pre-existing stale counts left alone: AUDIT "35 actions" header, B-AUDIT-IT "29 actions",
TENANCY's version-3/4 grant table.

**Mechanical harness pin correction (found during the docs pass; `db8e4dfb`).** G018 extended four whole tests whose
`run.sh` subtest/oracle pins were never updated, so their harness rows (run by the hosted staging gate) would FAIL:
B-OBJ-IT `TestIsolationAttacks` 17→**19**, `ISO-ORACLE` 82→**104** cases (matches `g018/s5-attacks.log` at the last
change `5424eca4`: 19 subtests, `{"cases":104,"disagreements":0}`); B-POLICY-IT `TestAccessFootprint` 2→**3** (`E-PG-2
onboarding writes (G018)`; `s6-fp.log` shows the three); B-POLICY-ORDERING `TestAccessOrdering` 17→**24** (+O-ONB-1,
-2, -3, -3s, -3p, -D, -E, the seven names in `s4-ord.log`); B-RLS-IT `TestRLSIsolation` 8→**9** (+`RLS-SCOPE-INSERT`;
`RLS-SCOPE` stays 36 cases, `s4-rls2.log`). PASS texts updated to the same counts. `TestAccessPolicy` 39 kept: G018 adds
no `t.Run` site to it (static diff), but its whole-test count was not observed. These counts are static/partial-run
derivations; whole-test confirmation is in the proposed union below.

**Escalation flag for Codex (not changed).** AUTHORIZATION "Availability: denied exclusive-write requests" says any
adapter exposing a guarded write that takes an exclusive row lock before the decision must first provide per-subject
admission or rate limits and bounded request/transaction lifetimes. G018's completions lock the target Group `FOR NO
KEY UPDATE` before the decision and are now reachable through the BFF; no such limit exists. A denied completion
changes nothing but can delay writers of that Group. Plan r5 §10 does not address it. Disposition needed (recorded
residual like PG-001, deferral to the pre-deployment/G021 work, or a required bound) — a security/availability
decision, so no code or doc claim was added.

**4. M1 operator preview — prepared, running.** Compose project **`g018-m1`** (`g018-scripts/m1.sh` = `b4.sh` with
project `g018-m1`, state `platform-evidence/g018/m1/`; services on 2xxxx loopback ports, never standard+100), built at
`eb84747e` (web/core sources identical to HEAD `294cd332`: later commits are docs, `run.sh` pins and management only),
`identity:migrate` version 14, seed `mcp-seed.sql` `5b30dbf3…f84a`; E-ONB-2 ran on it (so it holds a few test Groups).
Run CA `platform-evidence/g018/m1/ca.pem`; seeded users `platform-evidence/g018/m1/users.json` (0600; values never in
chat/logs). Smoke via `curl --cacert ca.pem --connect-to <host>:543:127.0.0.1:<port>`: `https://org-a.localhost:543/`,
`/sign-in`, `/groups` and `https://org-c.localhost:543/` all 200. The origins' port is fixed at 543 and G017's `g017-b4`
preview holds `127.0.0.1:543` (not touched), so the proxy currently publishes on a random loopback port. Owner steps:
(1) pause `g017-b4` when G017 permits (its own `b4.sh stop`); (2) `bash platform-evidence/g018-scripts/m1.sh publish`
(refuses while 543 is in use; recreates web and the proxy from the existing image on `127.0.0.1:543`, keeping
database, users, CA and seed); (3) trust `ca.pem` only in a dedicated browser profile and open
`https://org-a.localhost:543/`; sign in as carol (A owner), frank (A admin), alice (A plain member), bob (research
owner) or dave (research member; Group-only in C) from `users.json`; (4) Plan r5 §6.4 items: New Group with
suggested/edited address, a duplicate address, resume, both completion branches, paging, Groups/Business Profiles
areas, the shared editor from both entry points, member read-only, Light/Dark/System at 1280 and 320, keyboard-only
completion, error association on an invalid website. Cleanup after M1: `m1.sh down` (`down -v`) and the image receipt.
Readiness is not confirmed until the owner says so; M1 feedback pending.

| Batch (image / command) | Result |
|---|---|
| b3 `web.sh s10a` (`npm run check` = eslint `--max-warnings=0`, typecheck, `npm run test:unit`; then `test:unit`) at `eb84747e` | **PASS** exit 0; unit 193/193 (logs `web-s10a-check.log` `85c13c29…bec29`, `-unit.log` `22b381fa…8e2471`) |
| `m1.sh up` / `seed` (g018-m1) | start 0, migrate 0 (version 14), users 0, seed 0 |
| `m1.sh test b4-s10a 'E-ONB-2'` (onboarding-light, onboarding-dark) | **2 passed** (1.7 s / 1.5 s), first run (`b4-s10a/run.log` `08918488…3a7d`) |
| `run.sh --base 1bb052bd --allow-file management/v0/results/G018-scope-allowlist.txt --report-dir g018/lint-s10a lint lint-web` at `294cd332` | **exit 0, RESULT PASSED**; B-REQUIRED-CHECKS PASS (17 IDs); B-SCOPE PASS (79 paths, all within the 45 patterns); B-SHELL, B-CI-CONFIG, B-WEB-BOUNDARY/-COMPONENTS/-TOKENS/-DESIGN/-ASSET-CONTROLS, B-SECRET-CANARY/-SOURCE PASS; leftover inventory empty (`lint-s10a.log` `b95bc3aa…6097`) |

No negative control (the old page against the new E-ONB-2) was run. Known external failures, not touched or run:
`TestDNSSystemChecker` (held G017 DNS repair) and E-BRD-5 (G017 fix `0a0fc999` not yet on staging).

**Proposed final affected coverage (for Codex; nothing below was run except as recorded).** Union of Plan r5 §6.3
affected selectors, reusing attributable passing evidence:
- *Run (changed input not yet covered):* (a) `shell-light`/`shell-dark` `--grep S-LANDMARKS` in fixture mode (no
  `IDENTITY_E2E`; #9's one added frame assertion, never executed); (b) whole-test confirmation of the four corrected
  pins with their harness selectors, counts only: `./internal/tenancy/access/ '^(TestAccessPolicy|TestAccessFootprint)$'`,
  `'^TestAccessOrdering$'`, `'^(TestRLSIsolation|TestRLSPool|TestRLSFootprint)$'`, `'^TestIsolationAttacks$'` (one
  focus-DB run each; the alternative is to let the hosted staging gate confirm them, at the risk of a post-merge
  failure); (c) `lint lint-web` once more at the final candidate SHA (B-SCOPE/B-REQUIRED-CHECKS final), only if the
  candidate differs from `294cd332` beyond this Result.
- *Reuse:* B-UNIT-WEB `web-s10a` (web unchanged since); E-ONB-2 `b4-s10a`; the other E-ONB-1, 3–7 and onboarding
  S-LANDMARKS/S-NAV `b4-s9a` (18) and `b4-s9d` (14) — the page change alters output only for an incomplete Group with
  a status marker; mcp/identity nav-pin impact `b4-s9b` (14 pass, E-BRD-5 external); B-TENANCY-HTTP (66 probes, #9);
  B-WEB-COMPONENTS `lint-s10a`; core rows B-UNIT-CORE, B-ONBOARDING-CATALOG/IT, B-RLS-CATALOG/B-RLS-IT subcases, the
  access rows, `TestMCPTransport` and the version pins per #2–#7 (core source unchanged since `e1065c86`).
- *Not run locally:* identity/ops/full profiles, B-RECOVERY-DRILL (hosted migrations suite). Hosted: MR fast pipeline
  (automatic); the path-based staging gate after an authorized conditional merge.

Cleanup and resources: g018-m1 left running for M1 (9 containers running; volumes and its images kept until
`m1.sh down` and the image receipt); image `g018-web-check:s10a` kept with them; the harness run left no leftovers.
Predecessors' kept images (`g018-core-check:*`, focus images) untouched. G017 (`g017-b4`) and G019 resources untouched.
Budget B7 (checkpoint after attempt 4); attempt 1 continues (correction round), no reset or enlargement.

**Still open:** Codex review of this correction round, the pin correction, the availability flag and the proposed
union; owner M1 feedback (preview running); then the approved final affected coverage, final source/security/evidence
review, MR !41 ready, Duo, Codex pre-integration acceptance; G017 final ACCEPT gates integration.

### Successor #10 round 2 — M1 preview on 543, availability addendum, mechanical count/denominator corrections (attempt 1)

**Inputs.** Codex disposition on staging `cd0f7241`: `G018-rb-delta-availability-codex-review.md`
`566a5de103c64b95c4c411ea9108e2cd937944c11bd6c85a9883adb5531a5653` (R-B delta at `d27be289` ACCEPTED; Result `e52d0696`
imported exactly; availability is a requirement, not a residual) and the dated Spec amendment "Completion-adapter
availability — 2026-10-17 security review" (`G018-group-business-profile-onboarding-spec.md`
`6076fd8f3125c03f44efed89db77103d8a71b48d2a55b25a9cced7d0bacc5fdb`). Staging `cd0f7241` merged at `a0fa16a8`, then
`29e84643` (management-only G019 review r3) at `c4b87e24`; staging an ancestor of HEAD before each batch; scope
`74781683…6e56` unchanged. Mount table 13,836 of 100,000. All Docker commands under the `local-harness.lock` flock (no
refusal); `m1.sh publish` was run without an outer flock because each Docker command inside it already takes the same
lock (an outer hold would block them).

**1. M1 preview ready on 543.** The coordinator stopped `g017-b4` (project-scoped `stop`, nothing removed).
`m1.sh publish`: web=0, proxy=0; `g018-m1-web-1` now publishes `127.0.0.1:543` (and `127.0.0.1:23100`). Readiness
(`curl --cacert m1/ca.pem --resolve <host>:543:127.0.0.1`): `https://org-a.localhost:543/`, `/sign-in`, `/groups`,
`/business-profiles`, `https://org-c.localhost:543/` and `/sign-in` all **200** (rechecked after the later batch).
Owner handoff: `platform-evidence/g018/m1/M1-HANDOFF.md` (steps, dedicated-profile CA trust, the five seeded users by
role with credentials only in `users.json`, the Plan r5 §6.4 checklist, cleanup). M1 feedback pending.

**2. Availability addendum (not implemented).**
`management/v0/plans/G018-group-business-profile-onboarding-plan-r5-availability-addendum.md` (SHA-256
`e2969ff13e31114f8a62d688387c1343801e9e98624c399bf3321a40c2715860`, pushed at `5172faa9`, unchanged since): per-subject non-queueing admission keyed on `s.tc.Subject()`
after the copied `begin` and before both `Complete*` calls (bounded set, release on every outcome, closed
`onboarding_unavailable` 503); 10 s request/read bound on the completion route (read deadline + `WithTimeout`,
earlier caller deadline kept); 5 s decision / 5 s rollback cleanup via a completion-only `guardBounded` (storageTx
pattern) with `guarded` gaining a cleanup parameter and every other operation passing `cleanupUnbounded`; denial audit
2 s + 2 s accounted (≤ 19 s worst case per admitted completion); unknown-COMMIT unchanged; exact files (no scope
addition), four tests and focused selectors; residual flagged: admission is per core process. Awaiting Codex review.

**3. Mechanical corrections (`c2454967`).** Failure texts synchronized with the corrected pins: B-POLICY-IT `$f/3`
footprint, B-POLICY-ORDERING `$n/24`, B-RLS-IT `$n/9` isolation, B-OBJ-IT `attack $an/19` (the oracle text prints the
actual JSON). No expectation weakened. **Whole-suite count confirmation runs withdrawn** (not approved); replaced by
static source/log comparison (`g018-scripts/runcount.py` counts `t.Run` sites inside the function body; staging bytes
from `origin/staging`):

| Row / test | Old pin | Staging static | HEAD static | Log evidence | New pin |
|---|---|---|---|---|---|
| B-OBJ-IT `TestIsolationAttacks` | 17 / oracle 82 | 17 sites | 19 sites (+B2, +C3) | `s5-attacks.log` (last change `5424eca4`): 19 PASS subtests, `ISO-ORACLE {"cases":104,"disagreements":0}` | 19 / 104 |
| B-POLICY-IT `TestAccessPolicy` | 39 | 39 sites | 39 sites | (no change) | 39 |
| B-POLICY-IT `TestAccessFootprint` | 2 | 2 sites | 3 sites (+E-PG-2 onboarding writes) | `s6-fp.log`: the 3 subtests PASS | 3 |
| B-POLICY-ORDERING `TestAccessOrdering` | 17 | 7 sites (one table loop) + 11 table cases (O-1..4, O-6..12) = 17 | 9 sites + 16 table cases (+O-ONB-1, -2, -3, -3s, -3p) = 24 | `s4-ord.log`: all seven O-ONB names PASS with O-13/O-14/E-PG-3a/b | 24 |
| B-RLS-IT `TestRLSIsolation` | 8 (scope 36) | 8 sites | 9 sites (+RLS-SCOPE-INSERT) | `s4-rls2.log`: RLS-SCOPE-INSERT PASS, `RLS-SCOPE {"cases":36,…}` | 9 (36) |

`TestRLSPool`/`TestRLSFootprint` have no changed input (pins 7/pass unchanged). Whole-test counts are confirmed only by
the required hosted qualification; no full local pass is claimed.

**Fixture-mode S-LANDMARKS (approved focused case).** First attempt in the `g018-m1` runner namespace (`m1.sh fixture
shell-s10a`, log `ca126534…`): **2 failed — environment, not product**: the identity overlay's web answers the
generic-frame request at `127.0.0.1:3000` with 404 (console 404, no frame), because the overlay web is host-gated; the
shell projects need README's plain `platform:browser` environment. Rerun on a disposable plain stack `g018-sh`
(`g018-scripts/sh.sh`: README `platform:start`/`migrate` with `local.yml` only, ports 3xxxx) at `c4b87e24`:
`sh.sh test shell-s10b 'S-LANDMARKS'` (`--project=shell-light --project=shell-dark`, no `IDENTITY_E2E`) → **2 passed**
(417 ms / 326 ms; log `a93e1f7d…`). Cleanup: `sh.sh down` → 0 containers, 0 volumes, 0 networks; its 7 images removed
(`rmi-s10.sh`, 0 left).

| Batch | Result |
|---|---|
| `m1.sh publish` (g018-m1 → 127.0.0.1:543) | web 0, proxy 0; six readiness URLs 200 |
| `m1.sh fixture shell-s10a 'S-LANDMARKS'` (overlay namespace) | 2 failed — environment (host-gated overlay web 404), superseded |
| `sh.sh up`; `sh.sh test shell-s10b 'S-LANDMARKS'` (plain stack, fixture mode) | start 0, migrate 0; **2 passed** |
| `sh.sh down`; `rmi-s10.sh` | 0 containers/volumes/networks; 0 g018-sh images |

No lint batch: the only source change this round is `run.sh` message text (B-SHELL syntax checked with `bash -n`); per
the review, lint-s10a stays valid for identical blobs and a fast `lint` belongs with the next source change.
Resources: `g018-m1` running for M1 (9 containers, on 543); `g017-b4` stopped by the coordinator (not touched here).
Budget B7, checkpoint 4; attempt 1 continues.

**Open:** Codex review of the availability addendum, then its implementation with the addendum's selectors; owner M1
feedback; final source/security/evidence review; MR !41 ready; Duo; predecessor (G017) ACCEPT gates integration.

### Successor #11 implementer record (attempt 1 continued; verified before the first write, 2026-10-17 UTC)

Task G018, role implement, sole writer (fresh successor #11, replacing #10 under the context policy). Model Opus 5.5;
effort "high (requested)", actual value unknown. Native isolated worktree
`<source-checkout>` (generated `agent-<id>`
name, the accepted AGENTS.md fallback; not renamed); local branch `g018-s11` from
`origin/g018-group-business-profile-onboarding` `0f925d3b31a8e0816ee7f74bcebdd4a432066466` (MR !41 source).
`origin/staging` `1561fd754057f6401dca3a07bafe925d76f9920f` (management-only: SUPERVISOR-HANDOFF, PROGRESS, the
availability addendum review, the Spec clarification) merged without conflict before any edit and before any batch;
the bound files are byte-identical to staging. Bindings re-verified by `sha256sum` after the merge: availability
addendum review `4932c5b3ba028e737c1f59857d542df2d61811768bfe9329e8d089f272acedb7` (ACCEPT with implementation
requirements; read in full), Spec with its dated clarification `b2656c42…9213`, addendum `e2969ff1…5860`, Plan r5
`4fe4642f…75fa`, R0 addendum `829cfe4f…4dc84`, execution release review `fcfc0a15…556f`, PROCESS mechanical-corrections
`c1dbd4ce…49ab`, scope `747816831f313e2bea9c63fb9d2110c98270920e9d7b83727ab155c210d56e65` (45 patterns, unchanged),
Result `1a77a7f4c43d409e301940f9725c28eeffcdd2c64321758450edeb0b10a0dcbc` (the #10 records read). Budget **B7, checkpoint
after attempt 4**; attempt 1 continues, no reset or enlargement. Scope of this session: implement the accepted
availability addendum with the review's binding points (decision context into every completion act callback and its
failure classification; admission keyed on an unnamed UUID array, no tenancy import in the adapter; 256-entry
process-local nonqueueing cap; 10 s request/read, 5 s decision, 5 s cleanup, denial audit 2 s + 2 s unchanged; existing
unavailable and unknown-COMMIT semantics; admission after `begin`, before the completion transaction), tested with the
approved focused selectors only. No identity/ops/browser/full-suite rerun, no constant/AST mirror test. Known external
failures noted only: `TestDNSSystemChecker` (held G017 DNS) and E-BRD-5 (G017 fix not on staging).

### Successor #11 round 1 — availability addendum implemented (attempt 1; hand-back (c) for source/security review)

**Inputs and reconciliation.** As recorded above (staging `1561fd75` merged before any edit; bindings verified).
`prebatch.sh` (fetch, staging ancestry, `bash infra/ci/docker-desktop-mounts.sh`) before every batch: staging stayed
`1561fd75`, an ancestor of HEAD each time; mount table 13,836–13,844 of 100,000 (13%). Every Docker/harness command ran
under the `local-harness.lock` flock (directly, or per command inside the g018 scripts); no permission or lock refusal.
The worktree-isolation guard rejected some compound shell forms (chained git/hash commands, a sed with a variable
path, a python edit script, a heredoc append chained with a git commit, a `docker ps --format` listing); they were
re-issued as plain single commands as the guard instructs (or replaced by the Edit tool); the listing was not retried
(not needed). Scripts: `platform-evidence/g018-scripts/s11/` (#10's `prebatch`, `compile`, `unit`, `focus` with this
worktree's path).

**Implementation (`d0ff602d`; 8 files, all within the 45 patterns).**
- *Decision context into every act callback (the review's binding point).* `onboarding.go`: new unexported
  `guardBounded` derives `dctx := context.WithTimeout(ctx, completionDecisionBudget)` (5 s; an earlier deadline is
  kept), runs the guarded transaction under `dctx` with `completionCleanupBudget` (5 s) for the deferred rollback, and
  calls `act(dctx, tx, f)`. Both completion closures now take `ctx context.Context` as their first parameter, which
  shadows the caller's context, so `CreateBusinessProfileTx`, `LinkGroupBusinessProfileTx` and `writeFailure`'s
  classification all run on the decision-bounded context; the session/facts statements, `failureReason` and inTx's
  BEGIN/COMMIT/`commitFailure` already run on it. `recordGuarded` and `logGuarded` keep the caller's context (the
  denial audit's detached 2 s + 2 s unchanged). Only the two completions call `guardBounded`.
- *Cleanup parameter, scoped deviation from the addendum's wording.* The addendum said `guarded` gains a `cleanup`
  parameter. `guarded` has a third caller, `branding.go` (G017 `SetOrganizationBranding`), which is outside the G018
  scope, so `guarded` keeps its signature and becomes a one-line wrapper: `guardedTx(…, cleanupUnbounded, act)`. The
  former body moved unchanged into `guardedTx` (`facade.go`), with `cleanup` passed to its single `inContextTx` call.
  inContextTx call sites stay **12** (`TestAccessBoundary` passes); every existing operation keeps `cleanupUnbounded`.
  The same behaviour, with no unlisted file touched.
- *Admission (`http.go`).* `admission`: a mutex-guarded `map[[16]byte]struct{}` (an unnamed UUID array, no tenancy
  import; key `[16]byte(s.tc.Subject())`, the verified canonical subject, so the sessions of one user share it),
  `maxCompletionsInFlight = 256`. `try` never waits; its release is idempotent (`sync.Once`), so a stale release cannot
  free a later holder's entry. `completeGroup` acquires it after the unchanged `begin` and its body checks and before
  either facade call, `defer release()`, refusal = the existing `onboarding_unavailable` 503. The package comment
  states the per-process limit (N processes: up to N per subject).
- *Request/read deadline.* `bounded(d, next)` wraps only `RouteCompleteGroup` with `completionRequestDeadline = 10 s`:
  `context.WithTimeout(r.Context(), d)` (earlier caller deadline kept) and the connection read deadline set to the same
  instant via `http.NewResponseController(w).SetReadDeadline` before the body is read (the MCP handler's pattern;
  unsupported writers ignored), canceled when the handler returns. `begin` is byte-identical
  (`TestOnboardingCopyIsExact` passes); the other seven routes are unchanged.
- *Unchanged:* the unavailable mapping, unknown-COMMIT classification, lock order, SQL, error codes and the web.
- *Docs.* AUTHORIZATION.md `### G018 onboarding completions (the first adapter-exposed case)` under the availability
  rule: what is bounded, admission after `begin` (not before every database operation), the N-process limit, the
  ~19 s worst case with the 2 s + 2 s audit, the unchanged outcomes. PERFORMANCE.md PG-009 gains a "Completion
  lifetime bounds" bullet (fixed safety bounds, no latency claim).
- *Pins (mechanical).* `run.sh` B-POLICY-ORDERING `[[ "$n" == 25 ]]`, PASS text `25/25 … O-ONB-E and O-ONB-T`, FAIL
  text `$n/25`; no expectation changed.

**Tests (behavioural; no constant or AST mirror).**
- `TestOnboardingCompletionAdmission` (new): two "sessions" of one subject key, the second refused without waiting
  (a 1 s guard) while the first holds; another subject admitted; readmission after release; a stale repeated release
  does not free the new holder; release on panic and on cancellation, each followed by readmission; 256 distinct
  subjects fill the set, another is refused, one release admits it. Placement through the real mux: with the
  handler's set full, an unauthenticated completion is still `401 unauthenticated` (not 503) and takes no entry, so
  admission does not precede `begin`.
- `TestOnboardingCompletionDeadline` (new): the fakeDB fixture now records each statement's context. Over all eight
  pairs, only the completion route's `begin` statement runs under a deadline (≤ 10 s, > 9 s left), and that context is
  canceled once the request returns; with a 2 s caller deadline the statement sees ≤ 2 s. Read deadline: a real
  `httptest` server with `bounded(300ms, completion mutation)` and a raw connection that sends headers and a partial
  body, then stalls, gets `400 invalid_request` after ≥ 300 ms, and no statement runs.
- `TestOnboardingCompletionBudgets` not added: the review rules out a constant/AST pin, and O-ONB-T exercises the
  callback deadline against the database.
- `O-ONB-T` (new `TestAccessOrdering` subtest). (1) A model rename paused before COMMIT holds the target Group; a
  completion with a 1 s caller deadline ends `unavailable` within [1 s, 4 s]; the holder then commits; nothing changed;
  a later completion of the same Group succeeds. (2) For both `new` and `existing`, a model link paused before COMMIT
  holds the Group's association key (it is invisible to the completion's facts, so the decision allows). The act
  callback's association INSERT waits on it, observed as a lock wait, with **no caller deadline**. The completion ends
  `unavailable` within [5 s, 10 s]. Lock waiters and runtime-role open transactions settle to 0. The profile and
  association counts are unchanged while the holder is still open (the new mode's profile is rolled back). The holder
  then commits its association.
- **Negative control** (built from a `/tmp` copy, not the worktree): `guardBounded` passing the caller's `ctx` to the
  act callback (the captured-context defect). O-ONB-T **fails**: "act write (new|existing): still waiting after 15s (the
  wait is not bounded)". Its Group-lock phase passes, as the review predicted for a test that waits only at the
  pre-decision lock.

| Batch | Result |
|---|---|
| `s11/compile.sh s11a` (core check image; `gofmt -l .`; `go vet` and `go vet -tags integration` of the touched packages) at the implementation tree | build 0, gofmt 0 (no files), vet 0, vet-it 0 |
| b1 `unit.sh s11a b1-api ./internal/onboarding/api/ '^(TestOnboardingCompletionAdmission\|TestOnboardingCompletionDeadline\|TestOnboardingCopyIsExact\|TestOnboardingHTTP\|TestOnboardingRefusalMapping)$'` | **PASS** exit 0, 5/5 (Deadline 0.81 s; log `36b4cfcb…`) |
| b1 `unit.sh s11a b1-access ./internal/tenancy/access/ '^(TestOnboardingCompletionBudgets\|TestAccessBoundary\|TestRLSBoundary)$'` | **PASS** exit 0 (TestAccessBoundary incl. the 12 inContextTx sites, TestRLSBoundary; log `b4db421b…`) |
| `focus.sh up` (g018-focus postgres, no host port) | healthy |
| b2 `focus.sh it s11-ord ./internal/tenancy/access/ '^TestAccessOrdering$/^O-ONB-(T\|D\|E)'` (image s11a) | **PASS** exit 0: O-ONB-D, O-ONB-E, O-ONB-T (11.12 s) (log `adfb0fe3…`) |
| b2 `focus.sh it s11-onb ./internal/tenancy/access/ '^TestOnboarding$/^ONB-(5\|6\|8\|9)'` | **PASS** exit 0: ONB-5, -6, -8, -9, -9e (log `69b048e1…`) |
| negative control `focus.sh it s11-neg-ord … '^TestAccessOrdering$/^O-ONB-T'` (image s11neg) | **FAIL as intended** exit 1 (both act-write modes unbounded; log `a6b3f7f8…`) |
| `focus.sh down`; `docker rmi g018-core-check:s11neg` and `:s11a` | 0 containers/volumes/networks/secrets; both images removed |
| `run.sh --base 1561fd75 --allow-file management/v0/results/G018-scope-allowlist.txt --report-dir g018/lint-s11a lint` at `d0ff602d` | **exit 0, RESULT PASSED**: B-SCOPE PASS (80 changed paths since 1561fd75, all within the allow-list), B-SHELL, B-REQUIRED-CHECKS (12 IDs), B-CI-CONFIG, B-SECRET-CANARY/-SOURCE and the other lint rows PASS (log `8838242b…`) |

Not run (not authorized or unchanged): identity, ops, browser, full-profile and count campaigns. Unchanged UI and
S-LANDMARKS evidence (#10 `b4-s10a`, `shell-s10b`, `web-s10a`) remain attributable to unchanged web blobs. The whole-test
B-POLICY-ORDERING count 25 is from the static pin update (one subtest added), confirmed only by hosted qualification.
Known external failures, not touched: `TestDNSSystemChecker`, E-BRD-5.

Resources: g018-focus fully removed; s11 images removed; `g018-m1` (owner M1 on 543), `g017-b4` and G017/G019
resources untouched. Budget B7, checkpoint 4; attempt 1 continues (one implementation round, no reset).

**Open:** Codex final source/security review of `d0ff602d` (the `guardedTx` scoping choice; the admission/deadline
evidence); owner M1 feedback; MR !41 ready; Duo; pre-merge evidence; G017 final ACCEPT gates integration;
post-integration staging disposition.

### Correction #12 implementer record (attempt 2 of B7; verified before the first write, 2026-10-17 UTC)

Task G018, role bounded post-review correction, sole writer (correction owner #12). Model Opus 5.5; effort
"high (requested)", actual value unknown. Native isolated worktree
`<source-checkout>` (generated `agent-<id>`
name, the accepted AGENTS.md fallback; not renamed); local branch `g018-correction-12` from
`origin/g018-group-business-profile-onboarding` `4278f7b72b9d65d08b968ddd8fe542d3adfea44e` (MR !41 source).
`origin/staging` `95531f11dc1dc641d887b0c254fb1e89f6a08fe5` (management-only: SUPERVISOR-HANDOFF, PROGRESS, the Codex
source review, the preserved prior Result) merged without conflict before any edit and before any batch. Bindings
verified by `sha256sum` after the merge: Codex source review `management/v0/results/G018-implementation-codex-review.md`
`9b4815e99926d278fd6bea69ea82fbe606fd54c53c09a99462ebd9b0debf06bb` (CORRECTION LOOP, C1 and C2; read in full), scope
`management/v0/results/G018-scope-allowlist.txt` `747816831f313e2bea9c63fb9d2110c98270920e9d7b83727ab155c210d56e65`
(45 patterns, unchanged), this Result before this record `77321aa4c8ad6ef09bb60ada888e005833f985de023cd3eb5b8dc31a98219b7b`.
Budget **B7, checkpoint after attempt 4**: the review-ready submission ended attempt 1; this bounded post-review
correction is **attempt 2 of B7**; no reset or enlargement. Scope: exactly C1 (the Group detail card's industry `Text`
gains `[overflow-wrap:anywhere]`, `web/app/groups/[id]/page.tsx`) and C2 (the onboarding mutation decode requires
`io.EOF` from a second decode instead of `dec.More()`, `core/internal/onboarding/api/http.go`, with the trailing `]`, `}`
and whitespace cases added to `TestOnboardingHTTP`), all in already-scoped files. Validation: focused b1
(`TestOnboardingHTTP` with the copy/refusal checks), gofmt/vet, web check only if C1 needs it. No browser, identity,
ops, full-profile, availability or ordering rerun; unchanged evidence reused.

### Correction #12 round 1 — C1 and C2 (attempt 2 of B7; hand-back for narrow delta review)

**Inputs and reconciliation.** As recorded above (staging `95531f11` merged as `ea134792` before any edit; bindings
verified). `g018-scripts/s12/prebatch.sh` (fetch, staging ancestry, `bash infra/ci/docker-desktop-mounts.sh`) before the
batch: staging `95531f11` an ancestor of HEAD `e2ad07a5`; mounts exit 0, 13,844 of 100,000 (13%). Every Docker command
ran under `flock -w 300 <home>/.local/state/platform-evidence/local-harness.lock`; no permission or lock refusal.
The worktree-isolation guard refused two compound shell forms (a chained git/hash command, a sed loop over a variable);
they were re-issued as plain single commands, as the guard instructs.

**Implementation (`e2ad07a5`; 3 source files, all within the 45 patterns).**
- *C1* `web/app/groups/[id]/page.tsx`: the Group detail card's industry `<Text variant="meta">` gains
  `className="[overflow-wrap:anywhere]"`, the same class and multi-line form as `web/app/business-profiles/page.tsx`
  (the choices in `web/app/groups/[id]/onboarding/page.tsx` wrap through their `[overflow-wrap:anywhere]` parent span).
  Static verification only; no browser run (as instructed). The Vestive skill was not loaded for a one-class change that
  copies an existing renderer.
- *C2* `core/internal/onboarding/api/http.go` `mutation`: `dec.Decode(&in); err != nil || dec.More()` became
  `err != nil || !errors.Is(dec.Decode(new(json.RawMessage)), io.EOF)`. A trailing `]` or `}` is a syntax error on the
  second decode and any further value decodes without error, so both are refused with the existing closed
  `invalid_request`; trailing whitespace yields `io.EOF` and is accepted. Body budgets (`LimitReader` + length check),
  `DisallowUnknownFields` and the response are unchanged; branding and other adapters untouched; no shared parser.
- *Tests* `core/internal/onboarding/api/http_test.go` `TestOnboardingHTTP`, for every mutation pair next to the existing
  trailing-data/malformed cases: tails `]`, `}`, ` ]`, `\n}` → `400 invalid_request` (decode refused before the session
  check); tail ` \r\n\t` → `401 unauthenticated` (the decode accepted it and the request reached the next check, the
  session). Bodies are copied before appending so `ok` is not aliased. Not run as a negative control: with `dec.More()`
  the `]`/`}` tails would pass decode and answer 401 (stdlib `More` returns false at `]`/`}`), which these cases reject.

| Batch | Result |
|---|---|
| `s12/compile.sh s12a` (core check image; `gofmt -l .`; `go vet` and `go vet -tags integration` of the touched packages) at `e2ad07a5` | build 0, gofmt 0 (no files), vet 0, vet-it 0 |
| b1 `s12/unit.sh s12a b1-api ./internal/onboarding/api/ '^(TestOnboardingHTTP\|TestOnboardingCopyIsExact\|TestOnboardingRefusalMapping)$'` | **PASS** exit 0, 3/3 (`ok core/internal/onboarding/api`; log `g018/unit-s12a/b1-api.log` `e13bb24055be9106c6338bfc8b1bfc9033ca9401171ed1e44db6a522181ea047`) |
| `docker rmi g018-core-check:s12a` | removed |

Not run (instructed; unchanged inputs): web `npm run check` (C1 adds only a class string; ESLint/typecheck are
unaffected; MR fast checks run them), browser, identity, ops, full-profile, availability (`TestOnboardingCompletionAdmission`,
`TestOnboardingCompletionDeadline`, O-ONB-T) and ordering campaigns, and the scope lint (the three paths were already
in the 45 patterns and B-SCOPE passed for them at `d0ff602d`; MR checks rerun it). #11's s11a evidence stays attributable
to unchanged inputs except `http.go`'s decode line, which the b1 rerun above covers.

Resources: only `g018-core-check:s12a` created, then removed; `g018-m1` (owner M1 on 543), `g017-b4` and G017/G019
resources untouched. Budget B7, checkpoint 4: attempt 2 (this bounded post-review correction), no reset.

**Open:** Codex narrow delta review of `e2ad07a5` (C1, C2); refresh MR !41's slice-2 description (asked by the review,
not done in this round, which was bounded to C1/C2); owner M1 feedback; MR !41 ready; pre-merge evidence; G017 final
ACCEPT gates integration; post-integration staging disposition.

### Correction attempt 3 implementer record (attempt 3 of B7, checkpoint 4; verified before the first write, 2026-10-17 UTC)

Task G018, role M1 UX correction writer (items 1, 7, 8), sole writer. Model Opus 5.5; effort "high (requested)",
actual value unknown. Native isolated worktree
`<source-checkout>` (generated `agent-<id>`
name, the accepted AGENTS.md fallback; not renamed); local branch `g018-wt-ab83` from
`origin/g018-group-business-profile-onboarding` `f575a44cbe4f80f27dfba48df3fb1d157dfcd4ff` (MR !41 source, draft).
`origin/staging` `aff1bded170d25625604cd4790eeba64d1adbe6a` merged as `8a48a7ca` before any edit and before any batch;
one conflict, `infra/validation/run.sh` B-IDENTITY-CATALOG record text (G018 "version 14" against staging's G017
"2 G017 branding relations (50 in all)"), resolved by keeping both (version 14, 50 relations); no other hunk changed.
Bindings verified by `sha256sum` after the merge:

| Input | SHA-256 |
|---|---|
| Release review `management/v0/results/G018-m1-ux-proposal-codex-review.md` (read in full; binding refinements applied) | `6bf214298c7db58c4d80b83f5f73ebabeb10387f04fe63db399fb5a358c118fe` |
| Proposal `management/v0/plans/G018-m1-ux-proposal.md` (staging copy) = `<home>/.local/state/platform-evidence/g018/m1-ux-proposal/PROPOSAL.md` | `0b0877408ddfaf84bb5fb10a0f711afd15137b76f90175df892946347726cfcf` |
| Scope `management/v0/results/G018-scope-allowlist.txt` (amendment03, 47 patterns) | `9dc53f4951696fa55851819aaf7c1a29d16cc733a91a7a544c6164f8f953cc8a` |
| Goal (dated "Owner M1 UX amendment — 2026-10-18 UTC") | `05c2a11370540e13dd999d9072944f27292f91137795f45f22af33b3bef99331` |
| Spec (same dated amendment) | `0ef4c0c0e66e6ff67e57c0a32ebdbc1d97be7eaae7664c920566785c25046af1` |
| Plan r5 (preserved, unchanged) | `4fe4642ff0d5457db156ff3d5d31e723891dad1b4dd303a86ef0f70fb6dd75fa` |
| `management/v0/results/PROCESS-mechanical-corrections-20261017.md` | `c1dbd4ceb1e63c44caa489ed1ced35d1e08da7a1852da8188c43f3b5aec049ab` |
| This Result before this record | `b5f86df18b98fe91ea867cb833db3db2e87a6f836ecfd9f05c20df2f348b711e` |

Budget **B7, checkpoint 4**: the proposal spent no attempt; this release is **correction attempt 3 of B7**; no reset or
enlargement. Scope: accepted D1 (home "Sign in" as the existing `/auth/login` POST with PendingSubmit), D4 (onboarding
"Sign in first" notice, same form), D3 (definitive core `conflict` → inline PRG to the create form with a fresh
authorized holder read; the unknown-COMMIT 503 page unchanged) and D2 Option B (`GET /groups/address-check?slug=`,
closed `{state: in_use|not_seen|invalid}`, client blur advisory), with the review's binding refinements, the harness
pins (WEB_ROUTES, run.sh record text, request-context allow-list) and no invented `web-boundary` selector. Evidence
paths stay home-based (`<home>/.local/state/platform-evidence/g018/...`). Validation once on the final delta:
focused Docker Node onboarding-forms tests (incl. one stale-response case), E-ONB-1/4/6 in onboarding-light/dark on a
disposable g018-named stack (not g018-m1), one static `run.sh lint lint-web`. The live owner stack `g018-m1` and
forwarder `g018-m1-emu-fwd` are preserved.

### Correction attempt 3 round 1 — M1 items 1, 7, 8 (attempt 3 of B7, checkpoint 4; hand-back for scoped delta review)

**Inputs and reconciliation.** Staging `aff1bded` merged as `8a48a7ca` before any edit (recorded above). Before the
first Docker batch staging had advanced to `e31bfb47152bf97bd9e80f02c461b6b786a17c00` (management/docs only: PROGRESS,
ROADMAP, OS001 Goal/Spec, ARCHITECTURE); merged without conflict as `eb4051f7`, then ancestry verified. Every batch was
preceded by `git fetch`, the staging-ancestry check (`e31bfb47` an ancestor of HEAD each time) and
`bash infra/ci/docker-desktop-mounts.sh` (exit 0; 13,852–13,920 of 100,000 mounts, 13%). Every Docker command ran as a
plain single command prefixed with `flock <home>/.local/state/platform-evidence/local-harness.lock`; no
permission or classifier refusal. The worktree-isolation guard refused several compound or computed forms (chained
git/hash commands, heredoc edits, an inline `node -e` script, an `awk -v` loop, a `docker ps --format` template, a
`--grep` regex after `npx playwright test`, `--rmi local` as two words, a reference-filter glob); each was re-issued in
the plain form the guard asks for (Edit/Write tools, separate commands, `--rmi=local`, the Playwright command in an
in-container script `g018-scripts/b7/onb.sh`, the emulator query in `g018-scripts/b7/users.cjs`, a plain image list).

**Implementation (`9b159282`, test fix `91e798c2`; 12 source/test/doc files, all within the 47 patterns; B-SCOPE PASS).**
- *Item 1 (D1, D4)* `web/app/page.tsx`: the tenant entry's logged-out action is the sign-in page's own form
  `<form method="post" action="/auth/login">` with `<PendingSubmit label="Sign in" pendingLabel="Signing in…"
  data-testid="entry-sign-in">` (no fields; the route's exact-Origin, login-binding and closed-outcome guards apply
  unchanged; post-login target the route's default `/`). Signed-in viewers keep the Account link.
  `web/components/onboarding/onboarding-frame.tsx`: the "Sign in first" notice uses the same form (its `Button` import
  dropped). `/sign-in`, its status notices, the logout return and the minimal refusal pages' links are unchanged.
- *Item 7 (D3)* `web/app/groups/create/route.ts`: core's `conflict` is one `303` to
  `/groups/new?error=slug_taken&name=<name>[&slug=<slug>]` — the name under the existing non-empty/≤200-code-point
  rule (now `keepName`, shared with the field-refusal branch), the slug only when `isSlug`; no page rendering and no
  re-read in the route. A core `invalid_request` carrying the web-local code `slug_taken` is not a known refusal (400).
  The `outcomeUnknown` → `unknownOutcomePage("group")` 503 path is unchanged. `web/lib/onboarding-client.ts`:
  `FIELD_ERRORS.slug_taken` (field `slug`, "That address is already in use in this workspace. Choose another one."),
  `CREATE_CONFLICT` and `ADDRESS_ADVISORY` copy; `createConflictPage` removed (dead). `web/app/groups/new/page.tsx`: on
  `slug_taken` with step 2 the slug field shows the echoed grammar-valid slug (else the fresh suggestion) with the field
  error (`aria-invalid`, `aria-describedby`, autofocus as before), the summary "Nothing was created by this request."
  and, only when a fresh `listGroups(r.lookup(), …, { slug, limit: 1 })` returns a readable holder, the line "A Group
  you can see already uses it: <name>. Open <name> (→ /groups/<uuid>) If an earlier attempt was interrupted, this may
  be that Group." (React-escaped; name/id are core's validated answer). `form-field.tsx` gains `describedBy` (an extra
  id, listed last) and `children` (rendered after the error).
- *Item 8 (D2 Option B)* `web/app/groups/address-check/route.ts` (new): `GET` only (others 405 `Allow: GET`); profile,
  session cookie (else 401) and `ownOrigin` (else 401); exactly one non-empty `slug` parameter of ≤63 characters and no
  other parameter (else 400 `invalid_request`); a grammar-valid slug → core's existing guarded
  `GET /internal/v1/groups?origin&limit=1&slug=` with a fresh lookup request ID → `200 {state:"in_use"|"not_seen",
  slug}`; any other value → no Group lookup, core's session read for this origin → `200 {state:"invalid", slug}` for a
  member (403 for a non-member). Errors are fixed: 401 `unauthenticated`, 403 `not_available`, 503 `unavailable` (every
  upstream failure or malformed answer; never `not_seen`). `jsonResponse` (no-store, nosniff, no-referrer); no CORS
  headers; no names, ids or counts. `web/components/onboarding/address-check.tsx` (new, `"use client"`, imports only
  React): wraps the server-rendered slug field; native `input` → clear the hint, bump the generation, abort the request
  in flight; native `blur` → ask about exactly the field value (untrimmed) unless already answered; empty → nothing;
  longer than the server-supplied `maxSlug` → `invalid` without a request; one request at a time (relative URL, `GET`,
  `cache: "no-store"`, `credentials: "same-origin"`, `redirect: "error"`, 5 s abort timer); an answer applies only when
  its generation is current, the field still holds the asked value and the bounded (≤256 bytes, fatal UTF-8) JSON is
  exactly `{state, slug}` with a known state and `slug` equal to that value; anything else clears the hint. The hint is
  a `role="status"` line (`#slug-advisory`, `text-text-2`) that the field's `aria-describedby` lists; it never sets
  `aria-invalid`, disables submit, replaces the field error, polls, retries or writes. Copy comes from server props.
- *Harness* `infra/validation/run.sh`: `WEB_ROUTES` + `groups/address-check/route.ts` (C-sorted), the route-inventory
  log text, the request-context allow-list `app/groups/(create|address-check)/route\.ts` with its comment and failure
  text, and the B-WEB-BOUNDARY PASS text. `infra/ci/validate-controls.sh` derives from `WEB_ROUTES` (unchanged). No
  `web-boundary` selector was invented. *Docs* `docs/foundation/WEB-SHELL.md`: the logged-out entry sentence, the
  `/groups/create` row and a new `/groups/address-check` row.
- *Tests* `web/tests/unit/onboarding-forms.test.mjs`: PRG cases (grammar-valid, empty and non-slug values, over-long
  name, a forged `slug_taken` field code), copy pins (no "was created."/"successfully"/"recovered"/"availab"), the
  address-check route (exact `{state, slug}` for in_use/not_seen, one exact-slug lookup with its own request ID,
  security headers and no CORS; invalid via the session read and the non-member 403; no cookie, bad host, missing,
  empty, over-long, repeated and extra parameters refused before core; 401/403/503 mapping incl. a malformed and a
  thrown upstream; 405s) and the hint logic (the component transpiled with the project's TypeScript in a load hook):
  **the stale-response case** (a held answer for the old value, arriving after an edit and a newer check, is dropped;
  the edit aborts the request; only the current exact answer applies), echo/value mismatch, refusal/shape/oversize/
  content-type and timeout clears, over-long values ask nothing. `web/tests/browser/onboarding.spec.ts`:
  `credentials(…, from "/")`; E-ONB-1 signs in from the tenant entry's Sign in, checks `aria-describedby` lists
  `slug-advisory`, holds the `research` check with `page.route`, edits and blurs with its own slug (`not_seen`),
  releases the held answer (still `not_seen`), then a readable `in_use` (field not invalid, submit enabled), an edit
  clears it, then creates; E-ONB-4 (alice) sees `not_seen` for the unreadable `research`, submits it and gets the
  inline conflict without a holder, then creates with the suggestion; the C notice's Sign in is the login form (no
  link); the address check without a C session is `401 {error:"unauthenticated"}`; E-ONB-6 submits the second copy
  from a **script-disabled** context with the same session (native submit, PRG) and gets the inline conflict with the
  readable holder's Open link and `aria-describedby` = error, hint and advisory ids.

| Batch | Result |
|---|---|
| `docker build -f infra/docker/web.Dockerfile --target check -t g018-web-check:b7a web` at `9b159282` (+ staging merge `eb4051f7`) | build 0 |
| `docker run --rm --network none g018-web-check:b7a node --conditions=react-server --test tests/unit/onboarding-forms.test.mjs` | **21/22** — one test expectation too strict (a failure clears the hint twice: at the start of the check and on failure); batch finished, then fixed |
| `docker run --rm --network none g018-web-check:b7a sh -c "npm run lint && npm run typecheck"` | **PASS** exit 0 (ESLint `--max-warnings=0` over the web tree, `tsc --noEmit`) |
| rebuild `g018-web-check:b7b` at `91e798c2`; the same focused unit command | **PASS** exit 0, **22/22** (`g018/web-b7b/unit-onboarding-forms.log` `c44ff7c92c541757953f5099508a36f8ee7a949ce9e069b15c3f126964ad9503`) |
| `docker run --rm --network none g018-web-check:b7b npx eslint --max-warnings=0 tests/unit/onboarding-forms.test.mjs` (retest of the fixed file) | **PASS** exit 0 |
| Disposable stack `g018-b7` (env from `.env.example` with generated values, 0600, ports 3xxxx; default random proxy publish): `docker compose -p g018-b7 … -f infra/compose/local.yml -f infra/compose/identity.yml up -d --build --wait`; `… -f infra/compose/migrate.yml run --rm --build migrate up`; emulator user IDs via `g018-scripts/b7/users.cjs` in the browser runner; `infra/identity/mcp-seed.sql` through `psql` as the PostgreSQL superuser | start 0; migrate 0 (current_version 14); users 0; seed 0 |
| `… --profile browser run --rm --no-deps -e IDENTITY_E2E=1 -v identity-run:/identity:ro … web-browser` running `npx playwright test --project=onboarding-light --project=onboarding-dark --grep 'E-ONB-[146]\]'` at `91e798c2` | **PASS 6/6** (E-ONB-1/4/6 × light/dark; `g018/b7-onb/run.log` `d5d06e8191a26ed2fdd179806919ea033c83db548cd403c856a9b1fbc9cdac04`) |
| `bash infra/validation/run.sh --base e31bfb47 --allow-file management/v0/results/G018-scope-allowlist.txt --report-dir g018/lint-b7 lint lint-web` at `91e798c2` | **RESULT: PASSED**; B-SCOPE 83 paths within the allow-list; B-REQUIRED-CHECKS 17 IDs; B-WEB-BOUNDARY (new route inventory and request-context entry), B-WEB-TOKENS/-COMPONENTS/-DESIGN/-ASSET-CONTROLS, B-SHELL, B-HADOLINT, B-COMPOSE, B-CI-CONFIG, B-SECRET-SOURCE/-CANARY, B-DOCKER-ONLY PASS (`g018/lint-b7.log` `9f0ef42cecc41b4dcb72dd210030319de3fec7216b27a5c1bc4cbb17a2b0ddf1`) |

Not run (as released): E-ONB-2/3/5/7, the other onboarding/identity/shell browser projects (G017's E-BRD entry
assertions use only the unchanged `entry-sign-in` test id), core/identity/ops/root/controls suites and the full web
unit suite (lint and typecheck covered the whole web tree). No axe run with the hint visible (E-ONB-5 not in scope).

**Deviations.** (1) The 200 JSON is `{state, slug}` rather than `{state}` alone: the slug echo implements the binding
"echoes the exact submitted value" refinement and carries only the caller's own ≤63-character value (closed, ≤93
bytes). (2) The invalid path calls core's session read (the proposal said "no core call") so every answered state is
resolved by core for this origin, per the release ("resolve current session/origin in core, not cookie presence
alone"). (3) Non-member and other refusals map to a fixed 403 `not_available` besides the proposal's 401/503. (4) The
hint-logic unit test loads the TSX component through a TypeScript transpile hook in the existing test file (no new
tool or package).

**Resources and cleanup.** Created and removed: stack `g018-b7` (`down -v --remove-orphans --rmi=local`: 13
containers, 3 volumes, 3 networks, 9 images), images `g018-web-check:b7a`/`b7b` and the env file; afterwards no
`g018-b7` container, volume or image remains. `g018-m1` (10 containers incl. `g018-m1-emu-fwd`) untouched and running.
The worktree `agent-<id>` and its local branch `g018-wt-ab83` remain for the coordinator to remove after
this push.

Before the push staging advanced again to `b554a868f9e0efe9d44183766de4a4697517430c` (management-only: G054/G055
consolidation, PROGRESS, ROADMAP, SUPERVISOR-HANDOFF; no G018 input or web/infra/docs file); merged without conflict.
No check was repeated for it: every recorded result stays attributable to unchanged source, test and harness inputs.

Budget B7, checkpoint 4: attempt 3 (this correction), no reset. **Open:** Codex scoped delta review; owner M1 recheck
of items 1/7/8 (the live `g018-m1` preview still runs the earlier source until an intentional update); M1 items 2/6
under owner clarification; MR !41 refresh and ready; normal MR/Duo/staging gates.

## Pre-integration status: M1 closed, provenance repair folded, candidate CI (coordinator record, 2026-10-18 UTC)

Recorded by the coordinator (Opus 5.5, high requested; actual unknown). Management-only: no application, test or harness file changed.

**Owner M1 is complete.**
- The live `g018-m1` preview was moved to reviewed source `452f096d` by rebuilding only core, core-replica, web and proxy with `--no-deps`.
  - The emulator, postgres, redis, MinIO, intelligence, CA, volumes and the loopback emulator forwarder `g018-m1-emu-fwd` were preserved; their start times were unchanged.
- **Items 1/7/8:** on that preview, the owner rechecked direct sign-in, the inline duplicate-address conflict and the advisory address hint on blur. They answered "all good" (Codex verified the timestamp, 01:00:01.386Z).
- **Item 2:** the contrast refusal message was displayed, and a higher-contrast colour saved correctly.
- **Item 6:** a local `SEC_ERROR_UNKNOWN_ISSUER` on org-c. The preview certificate covers org-c and chains to the trusted test CA. Possible local ESET interception; no root cause is claimed.
- **Items 3/4/5/9:** work as approved.
- The org-a logo upload was verified read-only, from the database through the S3 key to the served bytes (hashes match), with org-b refused (404).

**Provenance repair folded.**
- Staging c722fa4c (supervisor repo-local storage policy) had made B-RENDER-PLAN/REPEAT fail on MR !41 (pipeline 2959365999).
- The accepted repair `2b39fcfe` (three render-customization registrations, count 28→31) was folded into this branch under scope amendment 04 (48 patterns, `123513ae…`).
- The handback was imported as `PROCESS-local-storage-provenance-repair-handback.md` (`758f9e16…`).

**Candidate CI.**
- MR !41 pipeline 2959388809 on `b7b50ef1` succeeded (`validate:checks`); `validate:integration` is manual for MRs.
- Codex's final pre-integration source ACCEPT is recorded on staging 0f4432e6 for exact `b7b50ef1`.
- This record reconciles staging 0f4432e6 (management-only) into the branch, so the candidate's application, test and harness blobs equal `b7b50ef1`.

**Open:**
- MR !41 ready, Duo and its disposition, required checks on the reconciled candidate, auto-merge;
- the post-merge staging gate (AC10);
- closeout cleanup, including the `g018-m1` preview, its forwarder and the preview worktrees;
- final ACCEPT.

Budget B7, checkpoint 4, attempt 3, unchanged.

## Staging repair (attempt 4)

Writer: Opus 5.5 (effort high requested by dispatch; actual unknown), isolated worktree
`.claude/worktrees/agent-<id>` (tool-generated name, not renamed), branch `g018-staging-repair`.
Authority: Codex CORRECTION LOOP `G018-staging-repair-codex-review.md` (SHA-256
`77dd9c2bc638392ff06718373c3bfa4cbec843ec69bca760a24fbf78fef73114`, handoff c53acee4) for failed staging pipeline 2959418555 (job
17260628234, merge 7aa0f5d8). Diagnosis evidence reused read-only from `platform-evidence/g018-gate-2959418555/`.

**Base/candidate.**
- Created from staging `4d65f1d2`; repair commit `59f3f747` (pushed).
- Pre-off fetch: staging advanced to `3fc2e828` (process/status-prose docs only; no G018 input, harness or web/core file).
  Rebased without conflict → `8ee87d37` (force-with-lease push of own temporary branch). Staging ancestor verified.
- Script blobs identical before/after the rebase: `run.sh` `4ba2919b`, `validate-controls.sh` `a179a74b`; the
  shellcheck, replay and controls evidence on `59f3f747` therefore applies unchanged to `8ee87d37`.

**Repairs (exactly two files).**
- **R1** `infra/validation/run.sh:2333`: B-MIGRATE-CATALOG summary predicate drops only the stale `roles…48` row;
  expects P, roles 0, roles 50. 14-line count, seven `mig_step` calls, owner/ACL/runtime/migrator flags, version and
  production-refusal assertions untouched. Comment `:2297-2299` (steps run at versions 0, 14, 13; the 48 row was the
  retired version-12 rollback) and PASS text `:2334` (0 at version 0; 50 at 14/13 = 35+3+4+6+2 relations) corrected.
- **R2** `run.sh:248-251` APPLICABLE: `core/internal/onboarding web/lib/onboarding-client.ts web/app/groups
  web/app/business-profiles`, G017-style comment.
- **R3** `run.sh:252-254` APPLICABLE: `infra/docker/core.Dockerfile.dockerignore infra/docker/web.Dockerfile.dockerignore`;
  intelligence ignore file not added. Attribution: corrects the omission in the upstream supervisor repo-local storage
  change `c722fa4c` and its registration repair `2b39fcfe` (comment in source); not a G018 product defect.
- **R4** `infra/ci/validate-controls.sh`: VA7 `required_count 86` + detail text (`:1120`, `:1123`); VU1
  `vu_split full 35 86 117` + detail (`:1214`); J8 `n != 86` + message (`:2061`). VU9 and selection semantics unchanged.

**Validation (minimum set only).** Evidence root `platform-evidence/g018-repair-a4/` (mode 0700).

| Command | Outcome | Evidence |
|---|---|---|
| `docker run … koalaman/shellcheck:v0.11.0 -S warning -x -P SCRIPTDIR infra/validation/run.sh infra/ci/validate-controls.sh` | exit 0, no findings | — |
| `docker run … koalaman/shellcheck:v0.11.0 -S info --include=SC2251 -x -P SCRIPTDIR infra/ci/validate-controls.sh` | exit 0 | — |
| `bash -n` on each script | exit 0 both | — |
| `catalog-replay.sh old <hosted catalog-p.log>` (`8953991…c50`) | **NOMATCH**, exit 1 (14 lines, 3 unique rows: objects 0×2, 50×5) | `catalog-replay-old.out` `df003ba3…f8d5`; script `6b4e99d4…a73` |
| `catalog-replay.sh new <same log>` | **MATCH**, exit 0 | `catalog-replay-new.out` `9db63b63…b3b8` |
| `flock …/local-harness.lock bash infra/ci/validate-controls.sh --only VA7,VU1,J8 --log …` (first try) | exit 2 usage: `--log` parent dir missing; no control ran | — |
| same, after `mkdir controls` | **3 passed, 0 failed** (of 3; 441 defined): VA7 86 IDs, VU1 35+86→117, J8 86 cases | `controls/ci-controls-summary.log` `bfd2edce…7eeb` |
| pre-batch `git fetch`; ancestor check (failed → rebase above → passed); `bash infra/ci/docker-desktop-mounts.sh` (13%) | ready | — |
| `flock …/local-harness.lock bash infra/validation/run.sh --base 3fc2e828 --allow-file management/v0/results/G018-scope-allowlist.txt (123513ae…) --report-dir …/off --secrets-file …/off-secrets off` at `8ee87d37` (run aeaf9f) | **exit 0, RESULT: PASSED**: B-TOGGLE-OFF-ADDITIONS (93 applicable additions byte-identical), B-OBS-OFF, B-BROWSER-OFF (144 tests, 0 skipped/flaky/unexpected), B-TOGGLE-OFF (start/check/stop, migrate to 14), B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE (2 paths), B-REQUIRED-CHECKS (7 IDs) PASS; build ledger 12/12 MET | `off.log` `0907d7d6…0277`; `off/results.tsv` `7c7a2777…be2a` |

Not run (forbidden/unneeded): full 441 controls, local migrations suite, root/ops/all, second off, MR manual
integration job, any pipeline retry/cancel. `g018-m1` stack, volumes and forwarder untouched.

**Cleanup.** Off run: leftover inventory empty; no `g001.run=aeaf9f` container remains. Run secrets file
`off-secrets` (mode 600) kept in the private evidence dir pending closeout cleanup (no deletion without owner approval).

**Open:** Codex source review of `8ee87d37`+Result; coordinator MR to staging, required checks/Duo, standing merge
policy; post-merge staging gate; G018 final ACCEPT.

Budget: **attempt 4 of B7, checkpoint after 4** (no reset/extension); checkpoint due before any further correction.

## Final integration, attempt ledger and closeout receipt (coordinator, 2026-10-18 UTC)

**Status:** technical and integration qualification complete. Codex accepted the repaired staging qualification and
authorized closeout (`G018-implementation-codex-review.md`, staging `16acdf19`, sha256 `df96d2dc…49b0`). Final Goal
ACCEPT is Codex's, after it verifies this record.

**Integration ledger**

| Step | Exact identity | Outcome |
|---|---|---|
| MR !41 (source `g018-group-business-profile-onboarding`) | merge `7aa0f5d8a1b4feeed7425457b49fc615ac3bf675` | merged after source/M1/Duo/Codex gates |
| Staging gate on `7aa0f5d8` | pipeline 2959418555 | **failed** (`validate:integration`: harness pins only; historical, not relabeled) |
| Staging repair R1–R4 (attempt 4) | `5f156047a4295b3e441b479e6177e91e998baa7a` (rebased from reviewed `ed0f12d8`, script blobs identical) | Codex source ACCEPT |
| MR !44 (source `g018-staging-repair`, deleted on merge) | MR pipeline 2959570970 success; Duo approved, no comments; merged pinned to `5f156047` as `3682446b42e031ba526e0c56c7413d96fe1bfc1b` | merged under standing authority |
| Staging gate on `3682446b` | pipeline **2959582812**: `validate:checks` 17261472501 success 435.78 s; `validate:integration` 17261472502 success 3385.19 s | **success**; Codex root verified artifacts by job ID: 35+86 PASS, union 117, 0 FAIL |

The optional MR manual integration job was not launched on either MR. No pipeline was retried or cancelled.

**Attempt ledger (B7, checkpoint after 4):**
- Attempt 1: implementation slices.
- Attempt 2: correction #12 (C1/C2).
- Attempt 3: M1 items 1/7/8 correction.
- Attempt 4: staging repair R1–R4.

No reset or extension; 4 of 7 used. The checkpoint after attempt 4 was assessed by Codex within B7.

**Closeout receipt:** `~/.local/state/platform-evidence/g018-closeout-20261018/` (0700, 60K, retained).

| File | SHA-256 |
|---|---|
| `evidence-manifest.tsv` (per-entry files/bytes/tree hash, recorded before removal) | `eec5d00e299268d501693018d420f00f32aefae956ef9b39d457af69e801b1d3` |
| `removed.tsv` (16 entries, all `removed`) | `5edc436860ad89ea4ab693d2c9792f6435fb0f7ba4073bc1101841f3ffb01898` |
| `g018-plan-history.bundle` (verified; refs g018-plan, -r1, -r2, -r4, g018-r0) | `75cde1a5d56774ad55f9e7737f618b8aa2b382c5cc6407af9b5e60326c9bc95f` |
| `docker-containers.tsv` / `docker-images.tsv` | `bb5e48f2…a6d4` / `3e24a6e4…8a8e` |
| `docker-volumes.txt` / `docker-networks.tsv` | `44b6f35c…49bc` / `c52ad9dd…653d` |

**Removed:**
- **Preview:** `g018-m1` compose project, through `docker compose -p g018-m1 down --volumes --remove-orphans`.
  - 11 containers;
  - volumes `identity-run`, `minio-data`, `postgres-data`;
  - networks `data`, `inference`, `edge`. `edge` was removed separately after the forwarder.
- **Forwarder:** container `g018-m1-emu-fwd`.
- **Images:** 9 `g018-m1-*:latest`, by exact tag.
- **Worktrees:**
  - `agent-<id>` (detached `452f096d`, clean);
  - `agent-<id>` (`g018-s10` `0f925d3b`, clean).
- **Local branches:** 18 `g018-*`. 13 are ancestors of staging. The 5 planning branches are covered by the bundle; every planning blob except one superseded interim r1 draft (`f3f1ed1e`) is already archived on staging.
- **Remote branches:** `g018-plan-r1`, `g018-plan-r2`, `g018-plan-r4`, `g018-r0`, all in the bundle.
- **Evidence:** 16 entries, 321 files, 11,840,065 bytes:
  - `g018`;
  - `g018-gate-2959418555`;
  - `g018-gate-2959582812` (Codex root proof, hashed above, now disposable);
  - `g018-repair-a4`, including `off-secrets`;
  - `g018-scripts`;
  - six provisional Plan copies and five provisional scope copies, all byte-identical to staging archives.

**Retained:**
- shared `project-archetype-render-cache` (not task evidence);
- G019's `g019-a1-gomod` volume, worktree and evidence;
- the coordinator checkpoint;
- G053/G054/G014 holds.
- No blanket prune.

**Post-check:**
- No `g018*` container, image, volume, network, local or remote branch, or worktree remains.
- The only `g018*` evidence entry left is the receipt directory.

**Owner-local cleanup (supported owner handling; not done by Claude):**
- Remove the G018 run test CA (`ca.pem`, now deleted from disk) from the dedicated browser profile it was trusted in. Or delete that profile if it was created only for M1.
- Remove any hosts-file line or ESET exception that was added solely for the `org-*.localhost:543` preview.
- Claude has no record of such hosts or ESET entries beyond the Item 6 note (possible ESET interception), so nothing is claimed removed.
- Unrelated trust settings are not affected.
