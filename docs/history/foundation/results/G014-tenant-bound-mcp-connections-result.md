> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G014 — Tenant-bound MCP connections and optional Group selection: Result

## Status

**Integrated, with closeout complete for final review; final ACCEPT belongs to Codex.** MR !31 merged as `a8d146c`
(exact head `8c56943`). Staging pipeline 2953945251 qualifies: the first gate run failed on B-OBJ-IT (cause unproven),
and Codex's single retry, job 17225023104, succeeded. See "Integration and closeout" at the end. The attempt-2 and
attempt-1 statuses below are kept as history.

**Attempt 2 (Duo follow-up pagination correction) implemented and validated for review; not accepted.** See
"Attempt 2" at the end: correction commit `fca7f2e`, rebased without source change onto staging `dc1a409`; focused
checks, the focused MCP integration run and the single approved `ops` run all PASS; cleanup verified. The attempt-1
status below is kept as history.

**Implementation complete for review; not accepted.** Review candidate: the head that carries this Result (branch
`g014-tenant-bound-mcp-connections`, draft MR !31; no merge authorized). Implementation attempt 1 of B7 (checkpoint after
attempt 4). Local evidence: focused batches b1–b3, the S-review and F3 corrections with their focused checks, and the
final run, recorded below with exact outcomes; the final's B-CI-CONTROLS row stays failed (VS3/VR1, environmental,
reran and passed in isolation) and its cleanup stays rc 18 (one unproven hold). Hosted MR fast, Duo and staging gates
are unchanged. The initial execution-readiness record follows unchanged as history.

## Initial record (2026-10-15, before execution)

### Initial status

**Initial Result; binding prepared for Codex; execution not started; not accepted.**
- This is the execution-readiness record that Plan r8 §0.1 and §12 ask for, ahead of Codex's recorded execution
  release. It is not implementation, not review-ready and not acceptance.
- Nothing has run: no S0, build, test, harness, Docker, provider call or migration. No delegate has been dispatched and
  no MR has been opened.
- Implementation starts only after Codex records the execution release.

## Execution binding (Git-first)

Hash rows carry neutral labels; the paths are listed under "Bound paths".

| Label | Bound bytes |
|---|---|
| Target: origin/staging at fetch, 2026-10-15 (contains the G013 ACCEPT commit `aae75c1`) | `ef0676a8ed830b776c8e46e37169a842377f1063` |
| Plan r8 (P8) | `e53728103993c6c545abbf7c778564d4194bff2b7c4813002551be5afc0412dc` |
| Plan r7 archive (P7), provisional design ACCEPT at `66a1b6f` | `e1186f4cc6b5a07623ebdc03cb5d18bf862aac87bb3fea1385d93e6b72fcc4ab` |
| Plan r7 review (R7) | `dd5e824aef7e592ec1beaea3d3dfaa2f4f945d8133c617f75552fd03ac727fbf` |
| Goal (G) | `e35a9e5dbf503c22b81a07653c17841d354818159fab6a05947268c4888fc255` |
| Spec (S) | `c5bc60c7c4bfc3ae92a0e63711e41a57eea6d375b6b78139bf2599eb68021a44` |
| ADR-005 (A5) | `db31ffd657d168796773bd21c8257085e24daca0d4947200facdf9ffac573b79` |
| Connection contract (C) | `23edbe6404694fa15866929775d224f9456a43f6d6f6df0e070faa0228ba2140` |
| SECURITY.md (D1) | `66e80909c686de57700c754443bbf12e287fb6850548e18130e9d9d7f4b04229` |
| Provider qualification review (Q1) | `d4bea7bd794f2c669138e859d998d5fe0a377402d555012f7947358926cd42e4` |
| Revocation qualification record, QP-B (Q2) | `9cd7709c3eade3fce4118496d4b91045345fda8e0cd26bd7311f402bbe44be79` |
| G013 implementation review with final ACCEPT and later evidence-folder records (E1) | `51658846ca380fa635b9c63da49e0f21bbd56322e77a46df186445042081ae0d` |
| G013 Result (E2) | `e6921e3688c047d195947d6c89a59b11c40184bdc718d6711795d4ce6d96a0a8` |
| RECOVERY.md (D2) | `12670f744937ea98e62d44ca3177f88d22d6d86f3446f3512497e0ba4a7b405e` |
| Restore invalidation SQL (D3) | `f198f41d0a15b0d4444b153a3db638deb90b9911d4bb0ee70444d41d9b5b2f52` |

**Provenance**
- Every input binds to the target and its path. The remaining Plan inputs (33 rows) are bound in Plan r8's
  bound-input table, verified against the target.
- The r6 to r0 archives sit beside the Plan, byte-exact as imported by Codex.
- This Result cannot carry its own hash; the hand-back reports it.
- The B-SCOPE allow-list text is Plan r8 §11.8. Codex imports it as `G014-scope-allowlist.txt` at binding.

**Bound paths**
- P8 `management/v0/plans/G014-tenant-bound-mcp-connections-plan.md`; P7 and the earlier archives are the `-plan-rN.md`
  files beside it.
- R7 G014-plan-r7-codex-review.md (not included).
- G `management/v0/goals/G014-tenant-bound-mcp-connections.md`;
  S `management/v0/specs/G014-tenant-bound-mcp-connections-spec.md`.
- A5 `docs/adr/ADR-005-tenant-bound-mcp-connections.md`; C `contracts/mcp/connection-context.md`; D1 `docs/SECURITY.md`.
- Q1 G014-provider-qualification-codex-review.md (not included);
  Q2 ONEOFF-G014-revocation-qualification.md (not included).
- E1 [G013-implementation-codex-review.md](G013-implementation-codex-review.md);
  E2 [G013-recovery-retention-incident-evidence-result.md](G013-recovery-retention-incident-evidence-result.md).
- D2 `docs/foundation/RECOVERY.md`; D3 `infra/recovery/sql/invalidate.sql`.

## Gates

| Gate | Status |
|---|---|
| Owner direction and option B | approved (dated Goal, Spec, ADR-005, Architecture and contract amendments) |
| Owner decisions OD-1, OD-3, OD-4, OD-6 | answered; OD-6 was "Continue with workos test" |
| QP (§11.0) | run; QP-1/QP-2 passed; QP-3 failed, kept as history and superseded as a requirement by option B |
| RV-1 provider stage and QP-4 namespace (QP-B) | passed and accepted (`33b6750`); limited to the actual user, pre-registered public client and configuration |
| Final G013 ACCEPT | done (`aae75c1`) |
| R0 (Plan r8 §0.2) | done; nothing material; Codex confirmations Q7 and Q8 requested |
| Plan | r7 provisional design ACCEPT; r8 R0 revision awaits Codex review |
| Execution release and binding | ready for Codex; not yet recorded |

## Budget and resources

- **Budget:** B7, with the checkpoint after attempt 4. **0 attempts used.** Planning, assessments, QP and QP-B spent
  none. Forecast: 2 attempts likely, 3 plausible (not a guarantee).
- **Effort:** implementation defaults to high. The model follows the recorded routing policy; the actual setting is
  recorded at start.
- **Validation resources:**
  - the coordinator's harness lock, taken at each batch start (none held now);
  - the existing port sets `PORTS_MIG` (`infra/validation/run.sh:161`) and `PORTS_IDN` (`:4451`);
  - existing harness projects, with no new port set, image, volume type or per-run bind mount;
  - fixed fixture paths and 0600 private inputs;
  - an evidence folder under `~/.local/state/platform-evidence/`, created at execution.
- **Provider resources:** none for implementation. Tests use the test trust anchor (C10). OT-1 runs only if a
  third-party client is claimed, which needs its own pin (OD-2) and authorization.
- **Leases:** none held. The worktree is the task's native planning worktree, branch `g014-plan`.

## Implementation record (2026-10-16)

**Status:** implemented and validated as recorded below; not accepted. Draft MR !31 (opened by the coordinator,
Codex-authorized; auto-merge not armed).

### Execution context

- Worktree `agent-<id>` (generated-name fallback; task G014, role implement), branch
  `g014-tenant-bound-mcp-connections`. Model Opus 5.5, effort high (instructed; the runtime setting cannot be verified
  from inside the worker).
- Release: `G014-plan-r8-codex-review.md` (handoff ada76dfa; release staging b04ec59). Budget B7: **1 attempt used**
  (implementation, focused batches, S-review fixes); the checkpoint stays after attempt 4.
- Evidence (temporary): `~/.local/state/platform-evidence/g014-impl-b04ec59/` (batch log, per-batch reports, owned
  registries, the S-review record).

### Binding reconciliation (mechanical rebinds, coordinator-confirmed; no G014 design, scope or interface change)

| Row | Old | New | Delta |
|---|---|---|---|
| E1 G013 implementation review | `51658846…ae0d` | `16b9af53…f9e6` | "Raw evidence hold resolved" appendix (73aba54) |
| ROADMAP | `15cd15c2…4407` | `cf436b34…96c3` | G016 Goal links (c438445) |
| ROADMAP | `cf436b34…96c3` | `b8ee52d9…3995` | G016–G018 rows (5e1c96b) |
| ARCHITECTURE | `e6993bcc…a67011` | `4ebd8e54…80bd8` | shared infrastructure section (75b73a5) |
| Allow-list (amendment 01) | `f0f7b487…` (38) | `55fb9e53…150b` (40) | `store_test.go`, `TENANCY.md` |
| Allow-list (amendment 02) | `55fb9e53…150b` | `fbdbe7b2…d022` (42) | `completion.schema.json`, `logrecord_test.go` |
| Plan r8 review (release) | `31c92b17…1cdf` | `a2d0dd44…b9db` | early draft MR appendix (37aa23a) |
| ARCHITECTURE | `4ebd8e54…80bd8` | `f2fb679e…9b3e` | §27.1 sqlc direction (65b10f0) |

Staging moves reconciled by rebase: 73aba54 → c438445 (browser-report ownership fix) → 4b7929f → 75b73a5 → 5e1c96b →
b703d64 (amendment 01) → 3e0e554 (amendment 02) → 5bc1bee (scope-metadata CI exemption: run.sh B-CI-CONFIG rules,
validate-controls.sh, profile-paths.tsv, OPERATIONS.md; Codex's hunks kept intact; it extends an existing F-row control,
so no G014 count pin moved) → 37aa23a → 51e4df0 → 2ffff8b (F3 correction accepted) → 01f5fd6 → 65b10f0 → 8a5e57f →
37c6143. No conflict at any step; Codex's hunks kept intact.

### Scope amendments (Codex-accepted)

- **01** (`G014-scope-amendment-01-codex-review.md`, `72cf8ac7…dcbc4`): `core/internal/identity/store_test.go`
  (TestIdentityCatalog's exact table set gains `mcp_connections`, RLS count 11) and `docs/foundation/TENANCY.md`
  (Boundary paragraph). Found by a version-10 sweep before b2.
- **02** (`G014-scope-amendment-02-codex-review.md`, `a0619f4a…71a1`): `completion.schema.json` (the two internal setup
  route paths in the global and core-only route enums) and `logrecord_test.go` (two positives, prefix and trailing-slash
  negatives; later the three MCP event names in TestCatalogueMatchesTheBrief, pre-approved).

### Implementation summary

- **Schema** `core/migrations/00010_mcp_connections.sql`: `sessions.kind`, `app.mcp_connections` (constraints, guard
  trigger, three Organization-bound policies, minimal grants), E1 resolves browser sessions only, `sessions_insert`
  admits `mcp` rows only in TENANT-ORG with a current role, two trigger-owned audit actions with widened CHECKs and
  attribution, exact v9 preconditions and v10 postconditions, refusing down.
- **Identity** (`identity/mcp*.go`, `identity.go`, `store.go`): connection records, M1 insert, M2 activation, M5
  renewal, revocation (scope `mcp`), E2 subject mapping; `workos/connect.go` Connect validator (bounded key set, or the
  C10 test anchor).
- **Facade** (`tenancy/access/mcp.go`, `access.go`, `context.go`): verified request scope and the guarded MCP
  operations; `tenancy/mcp.go` and `store.go` selectable-Group pages.
- **Config** `config/mcp.go`: `MCP_*` variables, bounds, registry, C10 refused outside test.
- **Transport** `core/internal/mcp`: registry, stateless metadata, admission (Origin, body, slots, rate), token and
  binding checks, the stateless streamable handler with per-request servers, five tools, internal setup routes,
  operational events. `cmd/server/main.go` mounts it only when configured.
- **Web**: `/connections/mcp` page and create/revoke routes; server-only `lib/mcp-client.ts` (strict answers, URL
  bound to the request origin).
- **Test overlay**: `identity.yml` MCP environment and registry mount, `proxy.conf` MCP-only routes to core,
  `emulator-seed.yaml` dave, `mcp-hosts.tsv`, `mcp-seed.sql`, identitysetup Connect test signer.
- **Harness**: B-MCP-CATALOG, B-MCP-IT (migrations), B-MCP-E2E (ops, incl. E2E-MCP-6 refusals with no bind mount),
  the MCP seed and mcp-* projects in the identity project, the B-MCP-SOURCE successor, B-MCP-HTTP wording,
  B-WEB-BOUNDARY, the C10 guard (N11), B-SKILL-UNCHANGED reviewed delta, version-10 pins, required checks (117; 33/14),
  control pins (gate 80; VU1 35/80/111; VU9 35/44/75), one profile-paths row, events catalogue, customization reasons,
  VENDORS rows pin.
- **Docs**: `contracts/mcp/control-plane.md`, `docs/foundation/MCP.md`, SECURITY §6.2 and §9, IDENTITY,
  AUTHORIZATION, DATABASE-ISOLATION, AUDIT, MIGRATIONS, OPERATIONS (RV-1 runbook with the Q7 caveat), RECOVERY (data
  class, Q7 wording), FOLLOW-UP, VENDORS, TENANCY; skill references and fixture headers (C4).
- 86+ files; see the branch diff against staging.

### Validation (local; finish, fix, rerun)

| Batch | Head / base | Suites | Duration | Outcome |
|---|---|---|---|---|
| b1 run1 | b7fe90d / 3e0e554 | lint lint-base lint-web lint-core provenance unit-core unit-web | 430 s | FAIL B-SHELL (SC2155), B-UNIT-WEB (foreign-host URL accepted); rest PASS |
| b1 run2 | d125cdd | lint lint-web unit-web | 261 s | B-SHELL PASS; B-UNIT-WEB FAIL (pending answer with a client accepted) |
| b1 run3 | f86052b | lint-web unit-web | 72 s | PASS |
| b2 run1 | f86052b / 3e0e554 | lint-base migrations | 1034 s | FAIL B-MCP-IT record check only (failed/write_conflict outside the catalogue); tests all PASS |
| b2 run2 | d4d2f1c / 5bc1bee | migrations | 914 s | PASS (37 rows incl. B-MIGRATE-CLI, B-IDENTITY-CATALOG v10, B-AUDIT-IT 31, B-MCP-CATALOG, B-MCP-IT, both drills) |
| b3 run1 | d4d2f1c / 5bc1bee | ops | 885 s | FAIL B-IDENTITY-E2E (M-LIST), B-MCP-E2E (E2E-MCP-7 threshold), B-IDENTITY-PRIVACY (duplicate request ID); rest PASS |
| b3 run2 | c8f8f9a / 37aa23a | ops | 861 s | PASS (18 rows incl. B-IDENTITY-E2E with mcp-light/mcp-dark, B-MCP-E2E 6/6, B-IDENTITY-PRIVACY) |
| F3 focused | cc14d2e, 38f8a71 / 2ffff8b, 01f5fd6 | go test on a fresh focus database | — | PASS: TestMCPCatalog 3/3 (CT-MCP-GUARD), TestMCPMigration 5/5, TestMCPOrdering 8/8 (O-MCP-5, O-MCP-6, O-MCP-2/2b), TestMCPFootprint, TestMCPTransport/Lifecycle/Audit/Recovery, TestMCPIdentity*, core gofmt/vet/unit |

Every batch b1–b3: fetch, ancestry check and the Docker Desktop mount check first; the shared harness lock; a per-batch
owned registry; finalize cleanup rc 0; `inventory-leftover.txt` empty. The final and its rerun follow below.

### S-review 01 (delegate 1, read-only, d4d2f1c): PASS WITH FINDINGS

| ID | Fix |
|---|---|
| F1 medium (slow body holds a slot) | body read under a 10 s read deadline before an in-flight slot is claimed; TestStalledBodyHoldsNoSlot |
| F2 medium (Plan §6.2 events) | `mcp.request`, `mcp.connection`, `mcp.setup` emitted with closed fields only; events.json schemas; catalogue pin (amendment 02) |
| F3 low (switch overtaken) | first the snapshot-expected compare (O-MCP-2b); Codex request 03 CORRECTION LOOP (Spec R5); correction proposal accepted (`G014-f3-correction-codex-review.md`, `9b0b374b…bb5`; proposal `management/v0/plans/G014-f3-correction-proposal.md`, `6ec8784d…`): one Organization-only write locking the previous and target Group facts; the guard validates visible same-Organization previous and target Groups; S1–S3 refinements (cc14d2e); O-MCP-5/6 and CT-MCP-GUARD |
| F4 low (hard-coded 30 min) | neutral wording |

Delta review of c8f8f9a: PASS (L1 revoke comment and not-found text, L2 AUTHORIZATION sentence: fixed at 78afb7b).
F3 implementation review of cc14d2e: PASS (T1 guard test requires one affected row in its success cases; T2 O-MCP-5
asserts one durable access.denied, reason `group_not_selectable`: fixed at 38f8a71).

### Deviations and decisions

- **F3 corrected (no deviation remains).** Codex request 03 (`G014-request-03-codex-review.md`, `861cabc3…1a2b`): Spec
  R5 allows only a clear or logout after the selected Group is lost; my first claim that G008 made the lock impossible
  was wrong (own Group bindings are visible in the Organization-only shape; the conflict was G014's own guard). The
  accepted correction is implemented at cc14d2e: the selection write runs in the Organization-only shape with the Plan's
  lock order (previous and target Group facts locked), the guard requires that shape and validates visible
  same-Organization previous and target Groups, and raced refusals are recorded denials.
- **Internal list answer** carries the session's CSRF token (one core call per page; the request-ID uniqueness rule):
  **accepted** by Codex (request 03) as a bounded internal-contract addition; documented in the contract.
- **Own revoke** of a terminal connection answers `not_found`.
- **A concurrent selection change** is a conflict (`access.decision` deny/`write_conflict`); the tool still reports
  `unavailable`.
- `scopes_supported` omitted (C7 alternative); `DisableLocalhostProtection=true` (S0; registry Host check first);
  audience mismatches are refused at verification before mapping, so only stored-binding/client mismatches record
  `resource_mismatch`; `core.exit` config keys do not list `MCP_*` (observability/exit.go outside the allow-list).
- `skills/saas-foundation/SKILL.md` still describes the archetype switching model (lines 40 and 108; outside the
  allow-list): follow-up.
- B-MCP-E2E uses the mcp-dark handoff (the later project); minted tokens reach the registry through stdout lines that
  are removed from the saved log.

### Final validation and reconciliation

| Run | Head / base | Suites | Duration | Outcome |
|---|---|---|---|---|
| final | 38f8a71 / 01f5fd6 | `run.sh all unit` (ci-controls auto) | 3682 s | rc 1: **109 PASS, 2 FAIL rows**: B-CI-CONTROLS (439 of 441 controls PASS; VS3 and VR1 FAIL, both the real-daemon storage-observation assertion "an available observation is not ok; a size is not a Docker CLI size", validate-controls.sh:3213-3214, run as the background child while the root suite built images) and B-REQUIRED-CHECKS (only that row). Every G014 check PASS (B-MCP-CATALOG, B-MCP-IT, B-MCP-E2E, B-MCP-SOURCE, B-MCP-HTTP, B-IDENTITY-*, B-WEB-BOUNDARY, B-SKILL-UNCHANGED, B-MIGRATE-*, B-AUDIT-*, both drills). Finalize cleanup **rc 18** (removed 0, refused 0, unproven 1, remaining 0, failed 0): see the unproven hold below. `inventory-leftover.txt` empty |
| final-rerun | 9541a03 / 65b10f0 | `ci-controls provenance` (isolated) | started 04:32:52Z; stopped 05:05:27Z | provenance complete and PASS (B-RENDER-ID, B-RENDER-PLAN, B-RENDER-REPEAT, B-PROTECTED incl. the rebased docs/ARCHITECTURE.md, B-SKILL-UNCHANGED). ci-controls **cancelled** after Codex's targeted-validation disposition: 410 controls PASS at the stop, no FAIL, **VS3 PASS and VR1 PASS**; SIGTERM to the harness's top process (its EXIT trap stopped the background child: "FAIL BACKGROUND ci-controls stopped before join"); rc 143; finalize cleanup rc 0 (removed 0, refused 0, unproven 0, remaining 0, failed 0); `inventory-leftover.txt` empty. The cancelled wrapper is recorded as cancelled, not as a pass |

**Local control coverage (Codex targeted-validation disposition, `G014-targeted-validation-disposition.md`,
`672d9f72…d9c9`):** the 441-control union is the original run's 439 PASS (38f8a71 / 01f5fd6, 2026-10-16T03:28–04:29Z)
plus the two actual passing reruns of VS3 and VR1 (9541a03 / 65b10f0, 2026-10-16T04:32–05:05Z). The executable tree is
unchanged between the two heads (the delta is documentation and management only). This is attributed coverage, not a
synthesized full PASS; the original B-CI-CONTROLS and B-REQUIRED-CHECKS rows stay failed as recorded, and the cause of
VS3/VR1 in the original run is unproven (probably contention on `docker system df` during concurrent builds).

Preserved evidence (temporary folder `preserve/`): the original final's `ci-controls.log` `d317f0bb…6996`, `results.tsv`
`656ff24f…9f96`, `required-checks.tsv` `2d136fea…c686`, `meta.txt` `21e9aa83…cef9`, `cleanup.txt` `923742ea…b0c`; the
rerun's partial `ci-controls.log` (copied at 381 PASS, VS3/VR1 PASS rows included) `ce4c8ef7…f83f`, `results.tsv`
`f4d13966…7157`, `timings.tsv` `6e3bc95a…9069`, `ci-controls-timings.tsv` `2c6b421a…a35a`.

**Unproven hold (cleanup rc 18 stays a failure):** image `beginify-core-check:latest` (`sha256:73bc165c…`; Compose
labels project `beginify`, service `core-check`), absent from the final's start inventory and present after it, with no
build or capture row in the owned registry, so the gating cleanup could not prove ownership and kept it. Probable cause:
the root suite's README `platform:check` line builds it without a registry row (a possible backlog item, not a G014
change). Held, not deleted, per the coordinator (an earlier image under that tag, `2db75c3e…`, is held for the same
reason).

**Open, unproven environmental item:** the MR !31 fast job on 9541a03 (pipeline 2953677058)
failed the scanner fixture suite B-SECRET-SOURCE and its tally; B-SECRET-CANARY passed (no leak). The same check passed on the five earlier MR heads and in the local final; the
38f8a71→9541a03 delta is documentation only. The coordinator is diagnosing the scanner fixture; no patch or retry here.

**Staging reconciliation after the final:** rebased onto 65b10f0 (documentation-only sqlc direction: ARCHITECTURE §27.1,
MIGRATIONS.md lines near the top merged with the G014 version-10 row, BACKLOG, PROGRESS, G016–G018 Specs); ARCHITECTURE
rebound `4ebd8e54…80bd8` → `f2fb679e…9b3e` (not MCP content); then onto 8a5e57f and 37c6143 (management, governance and
review records only). The only check that reads the changed protected docs (B-PROTECTED) passed in the rerun; the
MIGRATIONS.md docs blocks executed by B-MIGRATE-CLI and B-TOGGLE-OFF are unchanged.

**Preliminary source review** (`G014-preliminary-source-review.md`, `f9d6d56d…f52e`): no production-code blocker; one
bounded contract-wording correction, folded into this commit: `contracts/mcp/control-plane.md` now distinguishes a
definitive refusal or known abort (previous context unchanged) from `unavailable` after a possible commit (a lost COMMIT
reply or a failed post-write context read: outcome unknown; the client rereads the context before retrying). No new
error, schema or diagnostic; wording only, no rerun (no harness check reads that file).

### Acceptance criteria

Evidence map per Plan r8 §10.7. AC1–AC12 each have passing evidence in the final run (every G014 check PASS) and the
focused runs above; the final's failed B-CI-CONTROLS row is environmental and attributed as recorded. Acceptance is
Codex's.

### Claims

Evidence covers the tested client, user and configuration only. No claim about arbitrary MCP clients, no revocation
SLA, no immediate JWT invalidation. OT-1 not run (no third-party client claimed).

### Iteration history

- Attempt 1 (of B7; checkpoint after 4): implementation; b1 (3 runs), b2 (2 runs), b3 (2 runs); S-review 01 fixes;
  delta review L1/L2; Codex request 03 (CSRF field accepted, F3 correction loop); F3 proposal accepted and implemented;
  F3 review T1/T2; the final; the isolated, cancelled ci-controls/provenance rerun; the contract wording correction.
- Scope amendments 01 and 02; no budget extension.

### Suggested follow-up

- `skills/saas-foundation/SKILL.md` lines 40 and 108 (archetype switching model; outside G014's allow-list).
- Root-suite README `platform:check` image build without an owned-registry row (`beginify-core-check`, cleanup rc 18).
- MR !31 fast B-SECRET-SOURCE non-reproduction (coordinator diagnosing).
- G015 R0 reconciles its no-change wording with the `unavailable`-after-possible-commit distinction.

## Attempt 2 — Duo follow-up pagination correction (2026-10-16)

**Status:** implemented and validated; not accepted. Attempt **2 of B7** (checkpoint after attempt 4); the proposal and
addendum reviews spent none.

### Context and release

- Worker: native worktree `agent-<id>` (generated-name fallback), task G014, role implement (correction),
  the sole write owner after the attempt-1 implementer stopped. Model Opus 5.5, effort high (instructed; the runtime
  setting cannot be verified from inside the worker).
- Finding: Duo follow-up, CORRECTION LOOP, in `G014-implementation-codex-review.md` (`a7c325a4…9677`). The setup
  listing returned the oldest 100 live connections while creation is uncapped, so a newer connection was never listed
  and could not be revoked from the page.
- Design inputs (byte-exact Codex imports): proposal `plans/G014-pagination-proposal.md` (`d2427e91…9c9b2a`) and
  addendum 01 `plans/G014-pagination-proposal-addendum-01.md` (`b2bca697…be06cf`).
- Release: `G014-pagination-proposal-codex-review.md`, "Addendum01 review and correction execution release" (record
  `70329bd9…f71c` at staging 5e4af61). Its later appendices are the frozen-source preliminary review of `fca7f2e` (no
  source blocker) and the "Single ops exception" (record `c66fdb0b…9f15` at staging dc1a409).
- Allow-list unchanged (`fbdbe7b2…d022`, 42 patterns); every changed path matches it, so no scope amendment.

### Delta (commit `fca7f2e`, 13 files, +706/−55)

- **Core.**
  - `identity/mcp_store.go` and `identity/mcp.go`:
    - two constant statements share the unchanged owner, Organization, session and pending-expiry filters;
    - order `created_at DESC, id DESC` with `LIMIT 101`; the next page adds `(c.created_at, c.id) < ($4, $5)`;
    - new types `MCPListPage` (100), `MCPListPosition` and `MCPConnectionPage`.
  - `tenancy/access/mcp.go`: `ListMCPConnections(tc, after)`. Every page re-runs bind, live session, membership and
    the list in one snapshot transaction.
  - `mcp/connections.go`:
    - the query is exactly `{origin}` or `{origin, cursor}`, each with one value;
    - the cursor is decoded only after `begin` (session, origin, membership); any fault answers `invalid_request`;
    - cursor `s1|<µs>|<anchor UUID>|<tag>`, base64url (strict, no padding), at most 128 characters;
    - `tag` = the first 16 bytes of SHA-256 over user, Organization and the trusted registry `Origin`. It is a context
      tag, not a MAC, and the position is untrusted;
    - accepted time domain: µs 1..253402300799999999. An unencodable anchor fails closed (`mcp_unavailable`);
    - `next_cursor` is always present.
- **Contract and docs.** The GET row and an untrusted-position paragraph in `contracts/mcp/control-plane.md`; the
  facade row in `docs/foundation/AUTHORIZATION.md`.
- **Web.**
  - `lib/mcp-client.ts`:
    - an optional cursor, refused as `invalid_request` without a fetch when its shape is wrong;
    - a strict 3-key body; a non-null `next_cursor` needs exactly 100 rows;
    - the list-success-only bound `MCP_LIST_MAX_RESPONSE_BYTES` = 1,277,222. Create, revoke and every error answer
      keep 64 KiB. Streaming cancellation and the timeout are unchanged.
  - `app/connections/mcp/page.tsx`: Older/Newest links, a cursor-page empty state and a closed "Page link no longer
    valid" page. The create and revoke routes are unchanged: create lands on page 1, where the new connection is
    first, and revoke works from any page by connection ID.
- **Tests (existing functions only; subtest counts unchanged).**
  - `TestCursor`: setup codec, contexts (including an alias origin), the time-domain edges, overflow and the encoder,
    plus the real-`writeJSON` worst-case page size of exactly 1,277,222 bytes, with no hostile client ID above it.
  - `TestPackageBoundary`: allows `crypto/sha256`, `crypto/subtle`, `encoding/hex` and `strconv` (standard library,
    needed by the cursor).
  - `access/mcp_test.go`: stub signature only.
  - IT-MCP-3 body (`setupListingPages`):
    - 101 rows with a 3-row timestamp tie across the page boundary, then the exact traversal order, a stable second
      traversal, the anchor revoked and an older page-2 row revoked;
    - the 100/101 boundary and new-URL visibility;
    - malformed cursors and query shapes; another user, another Organization's origin and an alias origin of the same
      Organization (the alias handler gives 400, and the same cursor on its own origin gives 200);
    - a non-member and an anonymous caller with a cursor;
    - edited positions that list only own live rows, at the domain maximum (exactly page 1) and minimum (empty);
    - membership loss and restore, and an ended browser session, between pages.
  - Web unit: cursor request and refusal, next-cursor shape, the exact-maximum page accepted and one byte over refused,
    and over-64-KiB create, revoke and list-error answers refused.
  - `[M-LIST]`: newest first, no older link, and the invalid-page notice leading back to page 1.

### Reconciliation

- Rebased 1da97cd → staging 5e4af61 (management/governance only). Then, at safe boundaries, 5e4af61 → d926012 →
  dc1a409 (management/governance only).
- Source is unchanged by these rebases: `git diff fca7f2e 38ca1d6` lists only the 13 staging management, governance
  and skill files.
- Heads: `fca7f2e` (on 5e4af61) → `e44828c` (on d926012) → `38ca1d6` (on dc1a409). Each was pushed with a lease. The
  ancestor check passed before every batch.
- Rebind: ARCHITECTURE `f2fb679e…9b3e` → `8cc90850…04b1` (G016 sqlc timing amendment; no MCP content).
- Governance read: AGENTS.md and CLAUDE.md gained the worker-context handoff (350k) and the focused-selection default
  (d926012, `validation-integration.md`). They are not G014 bound inputs. The focused-selection policy was applied
  from the a2-run1 stop onward.

### Validation (focused; finish, fix, rerun)

| Check | Source | Command | Outcome |
|---|---|---|---|
| Core gofmt/vet/unit (dev image) | working tree = `fca7f2e` code | `docker build --target check -f infra/docker/core.Dockerfile core`; then, network none: `gofmt -l .`, `go vet ./...`, `go vet -tags integration ./internal/{mcp,tenancy/access,identity}/`, `go vet -tags mcpe2e ./internal/mcp/`, `go test -count=1 ./internal/mcp/ ./internal/tenancy/access/ ./internal/identity/` | vet and tests PASS (TestCursor, including the wire bound). gofmt flagged alignment in 2 files, fixed before the commit and re-covered by B-UNIT-CHECKS below |
| Web lint/typecheck/unit (dev image) | same | `docker build --target check -f infra/docker/web.Dockerfile web`, run with network none (`npm run check`) | PASS, unit 133/133 |
| a2-run1: `lint lint-base unit migrations ops` | `fca7f2e` | `run.sh --base 5e4af61 --allow-file …G014-scope-allowlist.txt --owned-registry … --report-dir …` under the harness flock, 05:55:28Z | **Cancelled** at 06:06:40Z (683 s, rc 143) by SIGTERM to the harness top process, after Codex narrowed the selection. Completed rows all PASS: lint 9 rows (B-SHELL … B-SECRET-SOURCE); lint-base B-LITERALS, B-NAMING, B-IDENTITY-GUARDS, B-DEPSCAN; unit B-UNIT-CHECKS (core gofmt/vet/unit, web lint/typecheck/unit, intelligence); and, unrelated and before the stop, B-MIGRATE-*, B-TENANCY-*, B-BUSINESS-PROFILE-*, B-POLICY-*, B-RLS-*, B-OBJ-*, B-AUDIT-CATALOG, B-AUDIT-MIGRATION. Owned cleanup rc 0 (removed 0, refused 0, unproven 0); leftover empty |
| B-MCP-IT equivalent (focused) | `e44828c` (code = `fca7f2e`) | disposable focus database: Compose project `g014w-focus` (postgres only, from `infra/compose/local.yml`, generated 0600 env). The dev core-check image on its data network with the run.sh IT env shape: `go test -tags integration -count=1 -timeout 30m -v -run '^(TestMCPTransport\|TestMCPLifecycle\|TestMCPAudit\|TestMCPRecovery)$' ./internal/mcp/`, with `MCP_INVALIDATE_SQL_B64` set; then `down -v` | PASS, counts 7/6/1/1 as the run.sh pin (IT-MCP-3 with the pagination extension, 0.75 s). 06:09:52–06:10:08Z, 16 s. Project containers, volumes and networks: none left. The test migrates to version 10 itself |
| a2-ops1: `ops` (the single approved exception) | `38ca1d6` (code = `fca7f2e`) | `run.sh --base dc1a409 --allow-file management/v0/results/G014-scope-allowlist.txt --owned-registry <reg>/owned.tsv --secrets-file <reg>/secrets --report-dir <ev>/report ops` under the harness flock | **PASS**, rc 0, 886 s (06:10:54–06:25:40Z). B-IDENTITY-E2E: 68 tests including M-LIST and M-A11Y in mcp-light/mcp-dark, 0 skipped/flaky. B-MCP-E2E: TestMCPE2E 6/6. Also B-IDENTITY-IT/PROBES/PRIVACY/CONFIG, B-OBS-*, B-OUTAGE, B-PERSIST, B-SECRET-ROTATION, B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE and B-REQUIRED-CHECKS all PASS. Owned cleanup rc 0 (0/0/0); leftover empty |

- **Ops exception:** "Single ops exception", `G014-pagination-proposal-codex-review.md` (`c66fdb0b…9f15`). Only
  `ops` reaches variants_identity, which owns the HTTPS/emulator environment, fixtures, handoff and MCP driver. The
  unaffected ops coverage is acknowledged, not called focused. There was one run and no wholesale retry.
- **Not repeated** (reused, unchanged): migrations/control suites, full `all`, CI controls, and the attempt-1 security,
  privacy and auth evidence for unchanged code.
- **Cleanup:** my labelled dev images (`g014-dev-core-check`, `g014-dev-web-check`) were removed. No labelled
  container is left, and the focus project left no container, volume or network.

### Acceptance (correction)

| Requirement | Evidence |
|---|---|
| Newest-first stable traversal with a 100-row page; new connection visible | IT-MCP-3: exact order, page 1 first row; `[M-LIST]` |
| Older rows reachable and revocable; anchor deletion tolerated | IT-MCP-3: anchor and page-2 revokes |
| No product cap, no order-only fix | 101+ connections listed across pages |
| Cursor: opaque, untrusted, context-bound (user, Organization, trusted origin); strict parser; time domain | TestCursor; IT-MCP-3 foreign, alias, edited and domain cases |
| Authority on every page | IT-MCP-3: membership loss, ended session, non-member, anonymous |
| List-success byte bound 1,277,222; 64 KiB elsewhere | TestCursor wire bound (real encoder); web unit maximum and over-bound |
| Unchanged auth, RLS, CSRF, no-store, events and codes | unchanged code paths; a2-ops1 identity/MCP privacy and E2E rows |

### Iteration history (attempt 2)

- The proposal, addendum 01 and the review loop (no attempt spent), then one implementation commit and the focused
  checks above.
- a2-run1 was cancelled by direction (selection too broad). No test failure in attempt 2 and no correction pass.

### Suggested follow-up (attempt 2)

- A bounded harness selector for the identity/MCP environment (variants_identity without variants_ops), as the ops
  exception records. Not in this correction.

## Integration and closeout (2026-10-16; coordinator)

### Integration
- **Duo:** on MR !31, Duo's listing thread was answered with the attempt-2 evidence and resolved. Duo's review of the
  current head found nothing new.
- **Fast:** pipeline 2953927234 passed on `8c56943`.
- **Merge:** auto-merge was armed with the exact SHA `8c56943a7964ee3641afee31750f8c49884d3617` and source deletion.
  It merged as `a8d146c656a40efe24d0233a0ec3aab830dedb60`. The staging tree equals the head, and the source branch
  is deleted.

### Staging gate (pipeline 2953945251)
- **Fast job 17223548435:** success.
- **Gate job 17223548436: FAILED.**
  - Only B-OBJ-IT failed: one TestObjects subtest, with the service/RLS oracles at 297/0 and 85/0, the attack suite
    17/17 and the footprint passing. The B-REQUIRED-CHECKS tally failed with it.
  - The hosted job doesn't publish the object log, so the failed subtest and its cause are **unknown**.
  - One authorized focused reproduction (`-count=3`, existing fixture) passed 3/3 (30/30 subtests).
  - The leading hypothesis, unproven, is the OBJ-SIGN 2 s pre-expiry window on a loaded runner.
  - Records: `staging-gate-2953945251-diagnosis.md` (abc601bd…) and `obj-it-diagnosis.md` (477ce2ec…).
- **Retry job 17225023104, run once by Codex: SUCCESS.** 80 PASS / 0 FAIL / 6 INFO, 79 required IDs, B-OBJ-IT PASS,
  B-CI-CONTROLS 441, and the leftover inventory is the baseline `project-archetype-render-cache` only. The pipeline
  is success. Disposition is on staging `f45ad71`.

### Budget
- Attempts 1 and 2 are G014's. The non-blocking OBJ-LOG diagnostic follow-up is shared attempt 3, with its own Result
  (`G014-object-diagnostic-publication-result.md`); its MR !32 auto-merge is armed.
- **Total: 3 of B7 used, under the checkpoint of 4.**

### Closeout
- **Receipt:** `~/.local/state/platform-evidence/g014-closeout/RECEIPT.md`, SHA-256
  `382df43b4bf4268df9310c78f4a3e88016aeb238b5bca6709a5f5d7c030b0323`.
- **Removed:**
  - G014's bulk run-evidence folders (about 20 MB);
  - two clean worktrees (the correction worker and the G014 planner);
  - three merged or equal local branches.
  - No G014 Docker resource remains.
- **Held:**
  - The implementer-1 worktree. Non-force removal refused, but it has 0 untracked files and all 1,271 files are
    byte-identical to `1da97cd`, an ancestor of staging. It needs an authorized force-remove.
  - Active OBJ-LOG and G015 resources.
  - Small record files.
  - The unproven `beginify-core-check:latest` image.
  - `origin/g014-plan` (history).
