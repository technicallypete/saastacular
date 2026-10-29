> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G055 Implementation Plan — Single-skill Restruct SDLC consolidation

Revision r1 (bounded correction of r0), 2026-10-16 UTC. Author: Claude Code (G055 planner, the single Plan
writer; task G055, role plan). Status: **provisional; proposed for Codex bounded re-review. Not approvable or
executable until the §1 gates close.** Nothing was implemented, installed, exported, published, bound, started or
validated during planning; no test, build, install, provider call, settings change, session or runtime change was
made. r1 answers the Codex r0 review (not included) (CORRECTION LOOP, planning only;
SHA-256 `2d5090eeaad4a724331a1d28d57102294c58dc03baaba327ea8d71e75da19653`): the four corrections are mapped in
§14 and the Q1–Q8 dispositions are recorded in §12. The reviewed r0 is preserved byte-exact as
plan-r0 (not included) (SHA-256
`f61c3f85d58697a2d182f0b8282d2c894d5ca07aa451e453ce7ab65119632538`, 621 lines), published by Codex on staging
`94604dac120fde7e850b58fa138b807974251cf2`, onto which this branch is rebased.

- Authority: the owner instruction "Replan 054 and 055" recorded in the
  planning refresh dispatch (not included), the [Goal](../goals/G055-restruct-sdlc-consolidation.md)
  (r1 plus its dated amendments, fourteen ACs) and the [Work Spec](../specs/G055-restruct-sdlc-consolidation-spec.md)
  (r1 plus the selective-adoption amendment and later policy inheritances).
- Tier: architecture/security/tooling (Goal Execution Policy): concise Plan, explicit Codex Plan ACCEPT before
  implementation. Budget **B5, checkpoint after attempt 3**; zero attempts used; planning consumes none (§11).
- Planning checkout: native isolated worktree `.claude/worktrees/agent-<id>`, branch `g055-plan`
  from origin/staging `834cde1c81e0fd6085bfcafee57d10138d953892` (the generated name is allowed by the standing
  native naming fallback). The coordinator safekeeps the branch; the planner pushes only this Plan. During
  planning, staging advanced to `8714ca5cf680a8c8df838d46a71b4248deedffe9`, which changes only PROGRESS.md,
  imports G054 Plan r2 (archiving r1 as `-r1.md`) and adds the G054 r2 Codex review; every other §2 row is
  byte-identical there (recorded, not re-planned).
- Effort: dispatched as Fable 5.1 **high**; not natively verifiable from inside this worker (recorded as
  dispatched/unverified). The parts given the deepest analysis, where the dispatch allowed selective xhigh, are the
  executing-code trust chain and home binding (§5.4), the legacy compatibility route (§5.3) and the stopped-observer
  cutover (§5.6). Implementation default: high (Opus high per the model-aware refinement); the owner may raise the
  R4/R5 phases.
- G054 is being re-planned in parallel. Every G054-derived assumption below is marked **P-G054-n** (§3) and is
  reconciled against G054's refreshed Plan before this Plan is finalized, and against the finally accepted,
  integrated and export-reconciled G054 before Plan ACCEPT (§1).

## 1. Gates

| Gate | Requirement | State at planning | Closes before |
|---|---|---|---|
| G1 | G054 final Codex ACCEPT (including AC7 live pair, F1 and hosted staging disposition) | open; G054 HELD at M1/AC7 (`6b82ac5`, focused source review accepted only) | Plan ACCEPT |
| G2 | G054 committed integration into staging (observer/helper/skill/template changes, status-marker amendment carried) | open | Plan ACCEPT |
| G3 | G054 source/export reconciliation recorded as Codex disposed it (r2 review Q2, **stage-only**): a disposable G054 export compared with the bound sibling; the sibling, its 21-path owner-held overlay, MANIFEST, index, README, tags and publication state are never mutated by G054. G055 records the sibling state (HEAD, tags, overlay set and hashes) at that time | open; sibling main `44699df` is 31 untagged owner syncs past `1.0.1` (verbatim files equal staging) and carries the preserved, never-swept overlay; 1.1.0 is a future recommendation only (Q5) | Plan ACCEPT |
| G4 | P-G054 reconciliation of this Plan against the refreshed G054 Plan and the actual accepted G054 source (bounded Plan update, or recorded "unchanged") | **first pass done in this revision** against G054 Plan r2 (`f1e0afc`, design-ACCEPTED by the r2 review (not included) imported on staging `8714ca5`); the rows stay provisional until G054's final ACCEPT, then a second pass against the accepted, integrated source. G054 has **no** implementation release | Plan ACCEPT |
| G5 | Codex Plan ACCEPT of the reconciled revision; exact input binding (§2 re-hashed on the then-current staging) | open | execution release |
| G6 | Execution release through the verified coordinator under automatic reassessment and execution release (not included); one implementer at high | open | attempt 1 |
| G7 | Required manual feedback (AC13 real-discovery evidence rows, §8) and operator readiness before F1 | open | F1/MR |
| G8a | **Live-code safety before integration.** The owner-authorized, PID-safe stop of the live Beginify observer (which executes the legacy path that integration replaces with a forwarder), with recorded readiness, **before the first operation that replaces live code in the main checkout, i.e. before the staging merge of this Goal**. Without this specific authorization the candidate is **HELD** before integration; integration and final acceptance never bypass it (§5.6) | open | conditional merge |
| G8b | The verified canonical restart after integration (§5.6 steps 4–5), under the same owner authority; no second observer, no replay | open | closeout |
| G9 | Owner authority for sibling commit/push/tag/release (Restruct SDLC 2.0.0, §5.9) and any home installation on the owner's machine | open | publication only |

G9 is not an implementation gate; a release-ready candidate is not a release. G8a/G8b are not test gates either
(every batch uses disposable repositories), but G8a **is** an integration gate: the live pair is stopped only by
the owner's authorization, never by this Goal's tests, and nothing is cut over now.

## 2. Bound inputs (staging `834cde1`)

Git commit and path are the default binding; SHA-256 values are listed so S0 can verify with `sha256sum -c`.

| Path | SHA-256 |
|---|---|
| management/v0/goals/G055-restruct-sdlc-consolidation.md | `fd6c4e22950fc275c1df504bf091d87787d6f3b6ecc264f72649289a84d2bc8f` |
| management/v0/specs/G055-restruct-sdlc-consolidation-spec.md | `7d9f3bea82155754770b339c15e3d8db953c6e8692c15674357bc33b31d4b141` |
| management/v0/results/G054-G055-planning-refresh-dispatch.md | `6154266a93aa4661be33344ec782d8a5190d37ad84c6af4f4a872ce796829128` |
| management/v0/results/G055-specification-review-r1.md | `d5ab66491437a089f1aa16705489512f7b878f90a0278d37685b0bf549fb5362` |
| management/v0/goals/G054-claude-only-supervision.md | `d836e445f26190396a0d772c3eabaccd9dbb10c01985b133f39732eee35bf84d` |
| management/v0/specs/G054-claude-only-supervision-spec.md | `ad7ce99887a0ce1b26c3a27c30bfb2adf8a0f27664190de5b9e15a40808c5925` |
| management/v0/plans/G054-claude-only-supervision-plan.md (r1, accepted) | `ae4724176722a7057fe4ff556b2d583b6ff3ffacd985a92f3e88dc1683807952` |
| G054 Plan r2 (provisional design, Codex design-ACCEPT; `origin/g054-replan` `f1e0afc476509cb724870d76b6aac29d5a8c2f0e`, imported at the same path on staging `8714ca5`) | `250bf0c54dff3ec50cae95800cc404bcc0c90ee9f157277aa0560183e90d07e7` |
| management/v0/results/G054-plan-codex-review-r2.md (staging `8714ca5`; Q1–Q5 dispositions override conflicting r2 proposal clauses) | `42596bbeaa1a94e290b51a30bc19d5e224f0887fd975a559fd5d89b77dc34367` |
| management/v0/results/G054-claude-only-supervision-result.md (imported submission) | `107c807bba630132fd277089ee47b6ef1f1e210e6ab2bfdc5579c0d8d71f1cf4` |
| management/v0/results/G054-focused-codex-review.md | `722027861b9b32c5c9091806dc85156bd42cd96d48ca25ad9c2089dac6997b04` |
| management/v0/results/G054-publication-follow-up.md | `d37eccea34c64b70d985336135197583333cc53c29f0933fdbc6f914e699c4ab` |
| management/v0/results/ONEOFF-coordinator-status-markers.md | `64b761a17c3217a7eb668203c54cd2bf8407b43c44d671ed2bd76e313b563ea7` |
| management/v0/results/ONEOFF-transport-capability-compatibility.md | `18f320f2731a11009bcfbdcf75a9db982d7f8b49f54c38691dbfa6d38456569a` |
| management/v0/results/ONEOFF-observer-daemon-0163-compatibility.md | `c7217e54c226739073ea15c604625777b4f5973f36175185b00a528edefad34b` |
| AGENTS.md | `5d5d818519bf16c0ecd2e6eef67426cc0e0fc25b36ccc7ec6ec1d93c9c90d468` |
| CLAUDE.md | `629bb40acbcc7f04d9f6d4790dd272aa47dcf2f4be6e6794822e223e332629ec` |
| docs/ARCHITECTURE.md | `c056ce5f5418bac1870dfbc05b521632d9b66c66ceb9906d3d7144f588f1d30f` |
| docs/SECURITY.md | `38aa0d1df17678dcb2a241cd739693f41eb66d1cdf4d7b23239b434776ef6b37` |
| docs/DECISIONS.md | `1223d089a32fac985e2a9d4e0cca6ee3ca5577752ae2dbc82c489cfff2110971` |
| management/v0/ROADMAP.md | `b8ee52d9201eeb8abcf897646ed69939b60e60c2643eb685cf14f74cabc53995` |
| management/v0/PROGRESS.md | `7ca3d109fb7045607f3527293fec0fcc6869b9ca51aec29432ccc483a761aff2` |
| management/BUDGET_POLICY.md | `0115fb585cbde0ffaa067507c477ec7d5dcd28c6fd37c2d47c700c9dbd657cd0` |
| management/SUPERVISOR-HANDOFF.md (Codex-owned; unchanged by G055) | `7c1345703c093f3c36a031bf41c9496252baeafc2cede1250c3bb12be41716c5` |
| docs/agents/CLAUDE-SESSION-HANDOFF.md | `0e34f34e234861e5e8f068fea3802d7205b1e7db9bf2aa163ce408c88915f4d1` |
| docs/agents/GITLAB-BRANCH-WORKFLOW.md | `9b2909825db7316156f209c8f1c7bd2a103a4093c34b007d373ce26b86fa3655` |
| infra/ci/profile-paths.tsv | `d78868bd9b396a0e9640d9418802f3ec82323ccff94f1554302206f9e751ef5b` |
| infra/ci/select-profile.sh | `335d0f41f3c2618b1863c0e44ddb90a07bc4f34f8262e89fdddfd080d0f3823a` |
| infra/validation/run.sh (B-SKILL-UNCHANGED, B-PROTECTED) | `e62f76127501fb6659cf4ae7cab14da80f968437f52c04e5776ea0834142ad44` |
| .gitlab-ci.yml | `15c81dfaad3d959024f58ddaa02a504560f94e4abcb5be70d4f496503ef905d9` |
| .gitignore | `e5ba1d203c3f31b6c397802a2f0da142bafc86ae1ba312f6e470632cde59639f` |
| skills/sdlc-management/SKILL.md | `0db2b5e492d65eaccf6a5da3b36d24c37bfd80859861bd43b9633be6388bf6d2` |
| skills/sdlc-management/scripts/check_sdlc.py | `e65ce7dd301cb0b1f238b33d5537b5bc69519cc40bdc608a7ffc9adfb7829725` |
| skills/sdlc-management/scripts/test_check_sdlc.py | `1dbde082babd217d4a4936a69ac4e01db471694ec7e62eeb1a1d7bfd0ef467c8` |
| skills/sdlc-management/references/adoption.md | `dc20ebf0957a7e287c43bf236d3978b651078ddb8987846bcedc9d7b33396ceb` |
| .agents/skills/claude-session-handoff/SKILL.md | `297663debd3d8e67f45f57fb6a419ebbf116be718306fff40f786730d2d00085` |
| .agents/skills/claude-session-handoff/references/local-session-handoff.md | `58620c03a35b4a4726e91d78a011e7756acf91e92d302f2af3fa5b9c467ba8d9` |
| .agents/skills/claude-session-handoff/scripts/handoff.py | `fa570adc6153c0dced2fbaaed6091367da8475993ddbd523eb75db6a7ee276ab` |
| .agents/skills/claude-session-handoff/scripts/watch.py | `79c86047437b6a6d3a281d9d49391dae4f62af35c581e35a4cd01e795dd9cb48` |
| .agents/skills/claude-session-handoff/scripts/test_handoff.py | `6d823601969256f27b18dee2b7a2238a4758ddde000af76979c4e5e935ca7051` |
| .agents/skills/claude-session-handoff/scripts/test_watch.py | `802a38c26a9a825beb2dc01bccc5964311afe3c208136fbe456baccded62770f` |
| .agents/skills/repository-supervisor/SKILL.md | `e2c06472d63527b1bc7c8aaa50d4f51310a8851663782c566a3cb89b65f2676c` |
| ../../restruct-sdlc HEAD (branch `main`, 21 dirty paths: 20 modified, 1 untracked, G054 overlay; read-only) | commit `44699df40996031c3c6ea457a9894242aacf61af` |
| ../../restruct-sdlc MANIFEST.tsv at HEAD (43 rows) | `8473a00bb5d36a888e5de8bbf05cdd4e0c5091697c35077f8e3c0f441671fc92` |
| ../../restruct-sdlc README.md at HEAD = working tree (owner prose) | `ec39ef94f6f7a05be3580afeaaf439181c19310c3707ce7a2361259ab12b0763` |
| Held G054 branch `origin/g054-claude-only-supervision` (read-only baseline) | commit `6b82ac57fbb6fde8f24ab426707fb57380b0ad36` (base `d1f4402`; 22 files, +2217/−40) |

Other inputs inspected read-only: the six sdlc references, templates, `agents/openai.yaml` files, G053 Plan r1/Result
(export procedure P1–P7, install-loop collision fixtures, gitleaks-under-lock), the 1.0.1 publication record,
`infra/ci/validate-controls.sh` (F-hand-*/F-agent-* selector fixtures), `claude --version` = 2.1.299, Python
3.12.3, Git 2.43.0, and this session's own skill listing (§5.2). The observer runtime/target state directory was
not inspected.

r1 note: at staging `94604da` the review found 39 of 41 directly checked rows matching; the G054 r1 Plan row now
lives at `G054-claude-only-supervision-plan-r1.md` (same bytes) and PROGRESS differs by the recorded G054/G055
management imports. These are rebound at the readiness check, not re-planned.

**Reconciliation rule.** Execution starts from the then-current origin/staging after G1–G3. S0 re-verifies this
table; a PROGRESS-only or Result-import change is recorded, a change to any helper/observer/checker/skill/template
row re-derives its delta before any batch (G054's integrated bytes replace the staging rows above), a changed Goal
or Spec stops work for Codex, and no input is refreshed silently.

## 3. P-G054 provisional interfaces

Every item is taken from G054's accepted r1 Plan, its held implementation and, since the coordinator's
2026-10-16 assumption message, G054 Plan r2 (`f1e0afc`, §2; "r2" below). Each stays provisional until G054's
final ACCEPT; G055 consumes the **accepted** form only. "Confirmed" means r2 states the same interface.

| ID | Provisional interface G055 depends on | r2 reconciliation | G055 use | If it changes |
|---|---|---|---|---|
| P-G054-1 | `handoff.py pin|unpin --role {coordinator,supervisor}`; `supervisor.json` (`claude-session-handoff/supervisor/v1`) beside `pin.json`; no cross-writes | confirmed | carried unchanged; schemas and the repo-keyed private namespace preserved (§5.6) | re-pin §5.5/§5.6 state tables |
| P-G054-2 | `watch.py bind --backend {codex,claude}`; `binding.json`/`state.json` carry `backend` only in Claude mode; `deliver_claude` ladder | confirmed | carried unchanged; the cutover (§5.6) is backend-agnostic | rerun V2 cutover cases only |
| P-G054-3 | `send --sender {codex-supervisor,claude-supervisor}` with `/proc` ancestry; frame labels `codex-supervisor`, `claude-supervisor`, `coordinator-observer` | confirmed; **AC7 live proof still unverified** | carried; home mode and forwarders add no sender; G055 neither depends on nor re-proves the live transport | if AC7 changes labels, update §5.7 text only |
| P-G054-4 | closure line `Supervisor backend: claude|codex`; `compact --verified-by {codex,owner,claude-supervisor}`; review name `<ID>-implementation-<backend>-review*.md` | confirmed | carried; the checker still never constrains review names | update §5.8 legacy-format rows |
| P-G054-5 | `SUPERVISOR_BACKEND` (codex default) and `SUPERVISOR_WORKER_ROOT`; coherent four-value adoption set | confirmed | templates move with the package; values unchanged | re-render AC3 fixtures |
| P-G054-6 | `test_claude_only.py` `Disposable` fixture | confirmed, with r2 **C1**: version `2.1.300`/protocol 1 now sends once with `compatibility_warnings`; only protocol 2 sends nothing | **reused** (post-C1 form) as the base of the AC13/AC14 fixture (§8) | adapt the import in `test_consolidation.py` |
| P-G054-7 | sibling state and write guard | **replaced**: main `44699df`, 31 untagged owner syncs past `1.0.1`, verbatim files equal staging, a preserved 21-path dirty overlay (20 modified, 1 untracked) that is never swept, 43-row overlay manifest. Codex r2 review **Q2: stage-only** — G054's corrections are candidate-only; D6 is not adopted; G054 performs no sibling write and no publication. **Q5:** 1.1.0 is a future recommendation only | G055 assumes **no sibling write path from G054**: its export is **stage-only (S)** by default; W (§5.9) exists only under a separately authorized owner sibling update that leaves the tree clean at a recorded commit, and never touches an overlay path | re-pin §5.9 only if the owner authorizes a sibling update |
| P-G054-8 | runbook G054 section and AC7 row; r2 **C3** wording to capability/baseline semantics | confirmed | path references updated by §6; content untouched | none |
| P-G054-9 | live-state guard (temp HOME only; `/proc` identity of the live observer recorded per batch); evidence under `g054-impl-r2/` | confirmed | reused verbatim as the G055 batch guard; G055 evidence under `g055-impl/` | none |
| P-G054-10 | `check_sdlc.py` unchanged by G054 | confirmed (r2 "Not changed") | all checker changes are G055 scope (§5.8); G055 carries G054's three ported checker tests | if G054 changes the checker after all, rebase §5.8 on it |
| P-G054-11 | status markers; backend-agnostic classification | confirmed; already on staging | accepted input, not provisional | n/a |
| P-G054-12 | transport: integer `peerProtocol` 1, versions as baselines producing `compatibility_warnings`, one pre-write reconnect, never a retry after `sendall`, creation gate version-restricted, `KNOWN_UNSENT = {version, codex-capabilities, supervisor-unverified}` (r2 C2/C4) | new in r2 | carried unchanged; G055 adds **no** known-unsent class: forwarder/home/alias refusals happen before durable intent and leave markers unconsumed | rerun V2 known-unsent cases only |
| P-G054-13 | sdlc `SKILL.md` at the P2 cap (200 lines; r2 **C5** moves the backend sentence into roles-lifecycle.md) | new in r2 | the G055 entrypoint is rewritten as routing with an explicit line budget (§5.1); every policy sentence of the old entrypoint lands in a reference, none is dropped | re-budget §5.1 |
| P-G054-14 | r2 **S1**: the candidate is re-derived as `g054-claude-only-supervision-r2` on staging (cherry-pick of `76821b0`, overlay as merge oracle); the held `6b82ac5` stays history | new in r2 | G055's base is the **integrated** commit, never the held branch; G054's attempt-1 evidence is history only | re-hash §2 at S0 |
| P-G054-15 | live Beginify stays Codex-supervised with its current pair, pin and observer | confirmed | §5.6: migration preserves `pin.json`, `supervisor.json`, `observer/*` schemas and the namespace key; the cutover changes only `process.json` argv under G8 | none |

Items already on staging (markers, transport compatibility, daemon 0.163 baseline, 350k/focused-selection/targeted-
rerun policies) are accepted inputs, not provisional.

## 4. Approach

- **One package, three modules, no new engine.** `skills/restruct-sdlc/` is the only mutable source copy. The
  three existing bodies become one 200-line entrypoint plus targeted references; helpers, templates, checker and
  tests move with `git mv` (history preserved). No transport, policy, installer, plugin or configuration framework
  is added; capability selection is routing text plus one checker option (§5.5).
- **Discovery by alias, executing code by real path.** `.agents/skills/restruct-sdlc -> ../../skills/restruct-sdlc`
  is a committed relative symlink (the G049 Vestive alias precedent). The helper binds the canonical regular files;
  an alias is resolved and containment-checked, never trusted as a committed blob (§5.4).
- **Minimum helper surgery.** Path assumptions change at the enumerated sites only (§5.4); the complete executing-
  code trust chain (self = committed = working = writer, process scan, main-checkout, lock, writer identity) is kept
  and extended to the forwarders and the home inventory. No check is weakened; new failures fail closed.
- **Additive private state.** Nothing under `~/.local/state/claude-session-handoff/<repo-key>/` is renamed, reset
  or migrated; only the observer `process.json` argv changes at an owner-authorized stopped cutover (§5.6).
- **Reuse tests; extend fixtures.** The 64 handoff, 82 observer, 35 checker and 5 Claude-only cases are reused;
  focused additions cover only the changed boundaries (§8). No broad default suite, no permutation campaign.
- **Export is a staged candidate.** Source/export reconciliation produces a verified stage and, only under the W
  preconditions, an uncommitted sibling write. Publication is 2.0.0 (breaking layout) under separate owner authority.

## 5. Design

### 5.1 Canonical package and module routing (R1, R2, R9)

```
skills/restruct-sdlc/
  SKILL.md                      name: restruct-sdlc; <= 200 physical lines incl. frontmatter; version 2.0.0
  LICENSE                       MIT (moved from sdlc-management)
  agents/openai.yaml            one UI metadata: display "Restruct SDLC", short description, default prompt
  references/
    roles-lifecycle.md budgets.md adoption.md coordination.md validation-integration.md closeout-risks.md
                                  (moved unchanged; anchors preserved; management module)
    supervision.md              repository-supervisor SKILL.md body (supervision module)
    handoff.md                  claude-session-handoff SKILL.md body (coordination module)
    local-session-handoff.md    bundled runbook reference (moved unchanged)
    installation.md             NEW: repository/home installation, selective adoption, legacy migration/path map,
                                  code binding, precedence, upgrade/rollback, cutover procedure
  scripts/
    handoff.py watch.py check_sdlc.py           (moved; minimum path adaptations, §5.4, §5.8)
    test_handoff.py test_watch.py test_check_sdlc.py test_claude_only.py   (moved; additions only)
    test_consolidation.py                        NEW: installation/migration/home/capability cases (§8)
  compat/legacy/claude-session-handoff/scripts/{handoff.py,watch.py}
                                  forwarder sources (§5.3); installed only at a legacy path during migration
  templates/                    moved unchanged (parameters.example.json, root/, management/, docs/)
```

Entrypoint contents (routing, not policy): frontmatter; a capability table (handoff only / management only /
supervision / full workflow → which references and scripts, which private state and approvals, what is **not**
activated); a task-to-resource table (adopt or scaffold → adoption.md + templates; lifecycle/budget/validation
rules → the matching reference; resume supervision/review → supervision.md; check/handoff/pin/compact → handoff.md;
observer → handoff.md observer section; install/migrate/home → installation.md); authority precedence; the
composition rule with `saas-foundation` (management process precedence only, unchanged); supported limits; the
three verification commands. "Loading assigns no role and grants no authority" stays on the first screen. A
routine status request loads `handoff.md` only; management templates and the supervision text are not required.

**Line budget (P-G054-13).** The current sdlc entrypoint sits at the 200-line P2 cap, so the new `SKILL.md` is
written fresh as routing, not by appending: frontmatter and boundaries ≤ 15, capability table ≤ 30, task-to-
resource table ≤ 35, installation/migration/home summary ≤ 25, authority precedence and composition ≤ 15,
supported limits ≤ 15, verification ≤ 15, dated policy pointers ≤ 20; target ≤ 170 with a 30-line reserve. The
old entrypoint's lifecycle summary, defaults and dated policy paragraphs move to the anchors that already hold
them (roles-lifecycle.md, validation-integration.md, closeout-risks.md, coordination.md, budgets.md) or, for the
observer/effort/fallback items, to `handoff.md`; the Result lists each moved sentence with its destination so
nothing is silently dropped. `check_sdlc.py package` P2 and `wc -l` prove the cap in V4.

R2 content (ownership matrix, Goal IDs, tiers, budgets/ledger, escalation, review outcomes, relay, publication
options, batch/rebase, manual-feedback order, MR/cleanup/archive, epic-closeout compaction checkpoint, Git-first
provenance, bounded cleanup, targeted reruns, 350k handoffs, focused selection, context compaction) moves
byte-for-byte inside the references; only intra-package link prefixes change. `saas-foundation` is not modified.

### 5.2 Discovery and placement (R1)

| Harness / placement | Discovery path | Evidence |
|---|---|---|
| Codex, repository | `.agents/skills/restruct-sdlc` (relative symlink to `../../skills/restruct-sdlc`) | symlinked skill directories are the accepted G049 pattern (`.agents/skills/vestive-design`); AC13 records the actual Codex listing (Q6) |
| Claude Code, repository | explicit read of `.agents/skills/restruct-sdlc/SKILL.md` as AGENTS.md instructs; no automatic discovery | **verified at planning:** this Claude Code 2.1.299 session lists `~/.claude/skills/*` and nested `*/.claude/skills/*` only; `.agents/skills/claude-session-handoff` and `.agents/skills/repository-supervisor` (both with `SKILL.md`) are not listed |
| Codex, personal/home | `~/.agents/skills/restruct-sdlc/` (real directory) | AC13 disposable HOME plus the owner's actual listing row (Q6) |
| Claude Code, personal/home | `~/.claude/skills/restruct-sdlc` as a **relative symlink** to `../../.agents/skills/restruct-sdlc` when the installed CLI lists it; otherwise an **identical copy** that is discovery-only (its scripts refuse to execute unless it is the bound home root, §5.4) | probe at execution with the owner's real session listing (Q2); fake transport never proves discovery |

**Home-only rendered references (r1 correction 3).** Rendered governance (AGENTS.md, CLAUDE.md,
SUPERVISOR-HANDOFF.md, docs/agents/CLAUDE-SESSION-HANDOFF.md) must reach the skill's resources from a repository
that holds no `.agents/skills/restruct-sdlc`. The smallest supported method is one existing-style template
parameter, `SKILL_LOCATION` with the two values `repository` (default; references render as today's relative
Markdown links `.agents/skills/restruct-sdlc/...`) and `home` (the same references render as plain code spans
`~/.agents/skills/restruct-sdlc/<resource>` and commands `python3 -B ~/.agents/skills/restruct-sdlc/scripts/<tool>
--repo .`). `~` is the user's home, not a machine-specific path, and no absolute path enters committed or public
content. No resolver, duplicated executable fallback or second template set is added: the template files are the
same, the parameter selects the rendering of each reference (conditional values, not conditional templates). The
checker's A2 resolves relative links only; in `home` mode the rendered code spans are checked by
`adoption --package ~/.agents/skills/restruct-sdlc` (each referenced resource must exist under PATH, read-only).
The AC13 walkthrough (§8) renders repo A in `home` mode and executes every rendered command and opens every
referenced resource from the rendered text, recording the actual access results; explicit `--repo`, immutable
external code and the collision refusals are unchanged.

Same-name resolution at install: an existing `~/.agents/skills/restruct-sdlc`, `~/.claude/skills/restruct-sdlc`
or `.agents/skills/restruct-sdlc` is never overwritten (IDENTICAL/COLLISION/STOP, §5.3). The owner's existing
docs-only `~/.agents/skills/claude-session-handoff` (SKILL.md, agents, references; no scripts) is inventoried and
left in place; its retirement is an owner decision (Q3). Relative resource references resolve from the selected
skill location; target-repository paths resolve only from `--repo` or the verified Git root.

### 5.3 Installation, legacy migration and compatibility (R3)

**Install loop (one skill, two placements).** The documented loop keeps the accepted G053 semantics (absent →
INSTALLED, identical → IDENTICAL, differing → COLLISION kept byte-identical, symlinked `.agents`/`.agents/skills`/
destination → STOP) for `skills/restruct-sdlc` into `TARGET/.agents/skills/restruct-sdlc` (repository) or
`$HOME/.agents/skills/restruct-sdlc` (home), plus the optional Claude home alias/copy. A repository that keeps the
Beginify layout (real `skills/restruct-sdlc/` plus alias) is documented as the source-maintainer layout; adopters
use the flat real directory. Bootstrap writes only owner-authorized target content; it never installs, trusts or
upgrades executable code (home mode needs the committed binding, §5.4). Interrupted installs are recoverable by
re-running the loop (idempotent; partial directories report COLLISION and are never completed silently).

**Legacy migration (three skills → one): order and format rule (r1 correction 1).** Two directory kinds are
distinguished by the checker and the procedure:
- a **legacy SKILL-bearing package**: `.agents/skills/<sdlc-management|claude-session-handoff|repository-supervisor>/`
  containing a `SKILL.md` (discoverable);
- a **pinned scripts-only compatibility directory**: `.agents/skills/claude-session-handoff/scripts/` holding only
  `handoff.py` and `watch.py` whose bytes equal the package's `compat/legacy/.../scripts/` files, with no `SKILL.md`,
  `agents/`, `references/` or tests (not discoverable; approved forwarders, §5.3 table).

Mixed = a canonical install plus any legacy SKILL-bearing package → refused by the final adoption check. A
compatibility directory is **not** legacy: canonical adoption accepts it and reports `compat-forwarders: verified
(2 files)` or fails A2 with `compat-forwarder-modified|extra-file`. Partial migration (any step below interrupted)
is refused by the final check and preserves every file; nothing is deleted or completed silently.

`installation.md` procedure, in this order:
1. **Preflight (non-activating; nothing retired, nothing written to the target):** inventory the legacy
   directories and hash them against the published manifest; record owner customizations as diffs; run the existing
   candidate/package checks that do not depend on the target's layout: `check_sdlc.py package <canonical package>`
   (P1–P7) and `check_sdlc.py adoption . --release <r> --format legacy` on the current install (the legacy format is
   still valid at this point). No final (canonical) adoption check runs here, so the migration is not rejected by
   its own strictness.
2. **Install** the canonical package with the loop above (COLLISION/STOP rules unchanged).
3. **Live code gate:** if `watch.py --repo . status` (run from the legacy path) reports a running observer, the
   owner-authorized PID-safe stop (§5.6 steps 1–3) must precede step 4; without that authority the procedure stops
   here and reports `migration-held-live-observer` (the canonical package stays installed, legacy discovery stays
   active, nothing is replaced).
4. **Retire legacy discovery:** remove the legacy `SKILL.md`, `agents/`, `references/` and test files; replace the
   legacy `scripts/handoff.py`/`scripts/watch.py` with the forwarders; remove `repository-supervisor/` and the
   storage copy `skills/sdlc-management/` (source-maintainer layout only).
5. **Final adoption validation:** `check_sdlc.py adoption . --release <r>` (canonical format; compatibility
   directory verified; mixed/partial refused).
6. **Commit**; canonical restart of the observer when applicable (§5.6 steps 4–5). Rollback is `git revert` of
   that commit (private state is untouched, so no state rollback exists or is needed); a reverted repository is
   again a legacy install and passes `--format legacy`.

Owner customizations are re-applied as a recorded merge decision, never automatically. Retired instructions are
not copied anywhere: the path map record `management/v0/results/G055-legacy-path-map.md` lists old path → new path
with the source commit and blob of each retired file; closed Goal/Plan/Result evidence keeps its old paths. The
migration fixtures extend the existing G053 install-loop and `test_check_sdlc` adoption fixtures (§8 AC4/AC9); no
installer or transaction framework is added.

**Pinned legacy command set (compatibility route).** Supported through forwarders at the legacy path, only in a
repository that migrated from a legacy install (Beginify included):

| Legacy invocation | Route | Output/exit |
|---|---|---|
| `python3 -B .agents/skills/claude-session-handoff/scripts/handoff.py <discover|check|send|receipt|pin|status|unpin|compact> …` | forwarder `os.execv`s the same interpreter on `<root>/skills/restruct-sdlc/scripts/handoff.py` (or the flat installed path) with identical argv | identical to the canonical command (the process is replaced; stdout/stderr/exit status are the canonical ones) |
| `python3 -B .agents/skills/claude-session-handoff/scripts/watch.py --repo PATH <bind|start|stop|status|unbind|smoke|poll-once> …` | same, to `scripts/watch.py` | identical |
| `python3 -B skills/sdlc-management/scripts/check_sdlc.py …` | **not forwarded** (no process/state ties; documented path change) | retired; the path map names the new command |

Forwarder rules: at most 20 lines, standard library, no argument parsing; it resolves `<root>` as the Git root of
its own real path, requires the canonical target to be a regular file inside that root (no symlink component), and
otherwise exits 2 with `legacy-forwarder-target-missing`. It never falls back to a home copy, another checkout or
a supplied path, and it never bypasses a canonical refusal (exec means the canonical code decides everything).
Deprecation: forwarders stay until the Beginify cutover (§5.6) is complete and one public release has shipped; a
later bounded change removes them. The legacy directory keeps no `SKILL.md`, so it is not a discoverable skill.

### 5.4 Executable-code binding and writer quiescence (R4)

**Repository-local mode (unchanged chain, new paths).** `HELPER_RELATIVE` becomes
`skills/restruct-sdlc/scripts/handoff.py` for the source-maintainer layout, with the flat installed path
`.agents/skills/restruct-sdlc/scripts/handoff.py` accepted when it is a real directory in `HEAD`; the helper picks
the committed regular path, never the alias (`git show HEAD:<alias path>` is a symlink blob and is refused). The
executing file must resolve inside the selected repository's canonical scripts directory
(`executing-path-outside-canonical` otherwise). Changed sites, and only these:

| Site (`handoff.py` at `834cde1`) | Change |
|---|---|
| `HELPER_RELATIVE` (L32) | canonical path; a second constant for the flat installed path; selection by `HEAD` content |
| `helper_dirs` (L982–991) | add the legacy repository scripts directory (forwarders), the canonical/flat directories of every worktree, and the home roots `~/.agents/skills/restruct-sdlc/scripts`, `~/.claude/skills/restruct-sdlc/scripts`, `~/.agents/skills/claude-session-handoff/scripts` (legacy personal) |
| `helper_argument` (L998–1016) | also match `watch.py` and the forwarder paths |
| `committed_helper_sha` / `prerequisite_evidence` (L1108–1150) | compare self/committed/working/writer on the canonical file; add the forwarder check (each legacy forwarder present in any worktree must equal `HEAD` and the package `compat/` bytes, else `legacy-forwarder-modified`); add the home-binding verification below when the executing file is outside the repository |
| `watch.py` `child_argv` (L477) | unchanged code; the recorded argv naturally carries the canonical path after cutover |
| new global `--repo PATH` on `handoff.py` (if absent at S0; `watch.py` already has it) | required when the executing file is outside the target repository; never inferred from the skill directory |

**Home mode (new, bounded).** A target repository approves external executing code with one committed, tracked,
clean file `management/RESTRUCT-SDLC-CODE-BINDING.json` (schema `restruct-sdlc/code-binding/v1`):
`{package_version, inventory: [{path, sha256}] for scripts/handoff.py, scripts/watch.py, scripts/check_sdlc.py and
every module they import from the package (currently `handoff` only), review: {path, sha256} of the reviewed
approval record, approved_at}`. No absolute path, user name or machine identity is stored. The helper in home mode:
1. resolves its own real path and requires it under exactly one supported home root (`~/.agents/skills/restruct-sdlc`
   or `~/.claude/skills/restruct-sdlc`, after symlink resolution); if both roots exist as distinct real directories
   with differing bytes → `ambiguous-home-copies`; if the target repository has its own committed canonical helper →
   `repository-install-present` with the exact local command printed (repository precedence; no silent switch);
2. reads the binding from the **target repository** (`--repo`), requires it tracked, clean (HEAD blob = working
   bytes) and well-formed; missing → `home-code-binding-missing`, dirty → `home-code-binding-dirty`;
3. hashes the actual executing files (itself, `watch.py`, imported modules) and compares with the inventory;
   mismatch or version drift → `home-code-binding-mismatch`.
Read-only `discover`, `check`, `status` and `watch.py status` run without a binding and print the explanation;
`send`, `pin`, `unpin`, `compact`, `watch.py bind/start/smoke` and any state write refuse before acting. A
command-line hash, the package's own manifest and `MANIFEST.tsv` never authorize anything. `handoff.py binding
propose --repo PATH` prints the JSON for review and commitment; it writes nothing. Upgrade = new inventory committed
by the repository after review; rollback = `git revert` of that commit; no auto-refresh or rebaseline.

**Shared home install and quiescence.** Before replacing code under a home root, the documented procedure runs the
existing `process_scan` over the home directories: any live, held or uninspectable process executing from that root
(an observer bound to any repository, a forwarder, a compaction) blocks replacement (`shared-code-in-use`, with the
PIDs and repositories derivable from each observer's `process.json` only through that repository's own
`watch.py --repo R status`). The bounded stop procedure is per repository (`watch.py --repo R stop`, PID-verified);
other repositories' approvals stay as they are and must commit a new binding before their sends resume. No cross-
repository repin, cleanup or state access exists. All supported paths share the existing maintenance lock and
writer-identity protections; a modified/uncommitted canonical file, a modified forwarder, a stale supplied hash, an
escaping symlink or an unsafe state directory cannot prune or archive delivery records (existing refusals).

### 5.5 Selective capabilities (R9)

| Selection | Resources loaded | Private state / approvals that may be created | Never created or activated |
|---|---|---|---|
| Handoff only | `handoff.md`, `local-session-handoff.md`, `scripts/handoff.py` | `~/.local/state/claude-session-handoff/<repo-key>/pin.json` (only on `pin`), the ignored delivery record `management/SESSION-HANDOFFS.jsonl` (only on `send`), home mode: the committed code binding | Goals/Specs/Plans/Results, templates, `parameters.json`, supervisor role record (unless `pin --role supervisor` is explicitly run), progress publication, observer (`bind` and `start` are separate explicit commands that require verified pair bindings) |
| Management only | `adoption.md`, lifecycle references, `templates/`, `check_sdlc.py` | the rendered management files the owner authorizes; `.gitignore` lines | pin, messaging, observer, supervisor role |
| Supervision | `supervision.md` plus the referenced lifecycle references, existing records | none (reads records); a review record only when authorized | scaffolding missing records, binding sessions, enabling messaging: missing dependencies are explained and refused |
| Full workflow | all | as adopted today (`OBSERVER_AUTO_ACTIVATE`, `SUPERVISOR_AUTO_PUBLISH` defaults unchanged) | nothing beyond the accepted defaults |

Selection is natural (the task names the capability) or the existing adoption metadata (`parameters.json` present
→ management adopted). The only mechanism is one checker option, `check_sdlc.py adoption REPO --release R
--capabilities handoff|management|supervision|full` (default `full`, byte-compatible with today): package/link/code
integrity checks always run completely; management checks (A1, A3, A4, A5) run only for `management`/`full`;
`supervision` verifies the referenced records exist and otherwise reports `dependency-missing` without scaffolding;
unadopted artifacts are reported `not-adopted` rather than missing or adopted. Changing selection never deletes,
rewrites or disables existing files, bindings, state or repository policy. The installed package holds no mutable
target data; `parameters.json`, records and state live in the target repository and its private namespace.

### 5.6 Session, backend and private-state continuity; stopped-observer cutover (R5)

Unchanged and verified by reuse: Codex-default and explicit Claude backends, UUID/repository/role verification,
truthful senders, native-permission refusal, receipt vs acknowledgment vs dispatch, closure verification, legacy
Codex closure records, Codex-unavailable trap in Claude mode, the 30-second default, the 100 MiB advisory, bounded
transcript parsing, singleton `run.lock`, PID-safe stop, retirement suspension, known-unsent recovery, uncertain
no-retry. The private namespace keeps its name (`claude-session-handoff`) and schemas; renaming the package renames
nothing there. No session is created or binding replaced during migration.

**Live code replacement and cutover (Beginify live pair; r1 correction 4).** The live observer executes
`.agents/skills/claude-session-handoff/scripts/watch.py` from the main checkout. Integrating this Goal into the
checkout that the observer runs from replaces that file with a forwarder while the old process may still be
running (its modules are in memory, but the accepted rule is that no code is replaced under a live writer, and no
immutable-old-code arrangement exists: the main checkout has one file per path and the process scan treats the
old path as a helper invocation). Therefore the **stop precedes the first operation that replaces live code**:
- **Before the staging merge (G8a):** the owner authorizes and the operator performs steps 1–3 below in the main
  checkout; readiness (PID, start, argv, timestamp, `process.json` absent, state hashes) is recorded in the Result
  and runbook. If the authorization or a clean step 3 is missing, the candidate is **HELD before integration**;
  the MR may exist but is not merged, and no later gate (review, integration, final ACCEPT) bypasses this.
- **After integration (G8b):** steps 4–5 from the canonical path of the main checkout.
- The observer stays inactive between G8a and G8b; that window is owner-chosen and reported, never hidden.
- Disposable-repository evidence (§8 AC7) is separate from, and never substitutes for, the authorization to stop
  the actual Beginify observer. No cutover happens now.

Pinned source/installed path: after G8b the observer runs from the canonical `skills/restruct-sdlc/scripts/watch.py`
of the main checkout. Steps, executed by the operator, recorded with timestamps and `/proc` identities (no UUIDs):
1. `watch.py --repo <main> status` from the **legacy path** (same process) — record PID/start/argv;
2. `stop` (PID and start-time verified; blocked → stop here, report, no further step, candidate held);
3. verify no observer process and no `process.json`; hash `binding.json`, `state.json`, dedup/cursor files;
4. `start` from the canonical path (default fresh start: tail, dedup retained, uncertain intents still block,
   no backlog replay; `--resume` only on explicit owner instruction);
5. verify exactly one process, `process.json` argv names the canonical path, state hashes unchanged except the
   cursor/process records; `status` shows the same supervisor fingerprint and backend.
Interrupted cutover: after step 2 the pair is simply inactive and reported; step 4 can be repeated; a failed step 4
leaves no second process (`run.lock`), and the old path is never restarted while the canonical one runs. A live
transcript or worktree is never moved. Beginify's pair, pin, observer and permissions are untouched by every test.

### 5.7 Reporting and role separation (R6)

The `COORDINATOR_STATUS: CLASS | TASK-ID | SUMMARY` contract, classes, working silence, marker priority, native
question notification, legacy prose detection, source/echo exclusions and the mixed-task priority rule stay in the
companion classifier (`watch.py`) and are taught once in `coordination.md` plus the templates; `handoff.md` links,
never duplicates. Aliases, discovery, forwarders and home placement assign no role and switch no actor; sender
labels and owner attribution stay as accepted (P-G054-3).

### 5.8 Checker, links and privacy (R7)

`check_sdlc.py` changes (minimum): `BUNDLE_SKILLS` becomes format-aware: **canonical** (`restruct-sdlc` only) is
the default; **legacy** (the three names) is supported only with explicit `--format legacy` for 1.x bundles and
installations (migration checks); a tree containing both `restruct-sdlc` and any legacy **SKILL-bearing**
package, or an incomplete legacy set, fails P7/A2 `mixed-format` under either format, whereas the pinned
scripts-only compatibility directory (§5.3) is verified, not treated as legacy, and the preflight/final ordering
of §5.3 decides which format is checked when. `adoption` accepts `.agents/skills/restruct-sdlc` as a
real directory or as a relative symlink that resolves inside the repository to a real directory (the resolved
package is checked with P1–P6; an escaping or dangling link fails A2). `--package PATH` lets `adoption` check a
home placement: package integrity at PATH, plus the repository's code binding against PATH's inventory
(`binding-ok|missing|dirty|mismatch`), without requiring a copied repository-local executable. `--capabilities`
(§5.5). A6's installed-prefix list follows the format. `INSTANCE_NAMES`, required Goal fields, ignore lines,
collision rules and P5 remain. Public-identity treatment: `package --public-remote URL` accepts a hosted-remote
match only when it is exactly URL and occurs in the bundle-root `README.md`; every other hosted-remote, home-path,
private-key or non-synthetic UUID finding still fails (the scan is not disabled). Links: P3/P4 verify every
entrypoint/reference/template link at the package, the alias, a flat install and a home placement in disposable
trees (reuse of the `test_check_sdlc` fixtures). Licenses: one MIT `LICENSE` in the package plus the bundle root.

### 5.9 Source/export coordination (R8)

- **Mapping.** Source `skills/restruct-sdlc/**` → sibling `skills/restruct-sdlc/**`; the sibling's three legacy
  skill directories are removed in the same export; bundle root keeps `LICENSE`, `README.md`, `MANIFEST.tsv`.
  Adapted files (synthetic UUIDs in `test_handoff.py` L10–12; the public-identity README) keep the G053/G054
  adaptation-diff invariants; everything else is verbatim (`cmp`). The manifest is rebuilt with unique paths,
  origins `new|verbatim|adapted` and verifying hashes; the Result records the full before/after table and the
  old-layout → new-layout map.
- **README (owner prose preserved).** Only the Skills, Requirements, Installation, Verify, Adopt and Limits
  sections change to the one-skill layout, the home placement, the selective capabilities, the legacy migration
  pointer and the updated limits; "Why Restruct?", the intro, status line and license text are untouched. The
  text is proposed in the Result and applied in the stage; owner wording decisions are recorded.
- **Compatibility and release.** The layout change breaks the published install loop and legacy paths: SemVer
  **2.0.0** recommended, with migration notes; G054's 1.1.0 is itself only a future recommendation (r2 review
  Q5), so the release line (1.1.0 then 2.0.0, or one 2.0.0) is an owner decision (Q1). No tag, release, force
  push, remote change or sibling commit follows from acceptance.
- **Stage and write.** Stage in `$EVID/export-stage/` from committed blobs (G053 §5 procedure: P1–P7 with the
  private literal list, four exported suites, verbatim `cmp`, gitleaks tar-on-stdin under the shared lock, notice
  inventory, collision fixtures). Sibling write **W** (uncommitted) only if: HEAD equals the post-G054 commit
  recorded at G3, tags `1.0.0`/`1.0.1` (and any G054 tag) unchanged, `status --porcelain` empty (the G054 overlay
  has been committed through its own route; G055 never writes, compares against or "resolves" an overlay path),
  remote unchanged (redacted), manifest verifying, no symlink component, every realpath inside the sibling root, no
  pending owner work; otherwise stage only **S**, reported. G054 is stage-only (r2 review Q2; D6 not adopted), so
  the overlay persists until the owner acts: the **default G055 outcome is S**, with the disposable stage compared
  against the bound sibling (equal/differing path list, no writes), exactly as G054's Q2 procedure. Dual-repository completion (commit/push both, verify both remotes) is the owner's
  publication step (G9); the Result reports partial state honestly and never claims both current.

## 6. Files, modules and proposed file scope

### 6.1 Repository allow-list (`G055-scope-allowlist`; the F1 scope diff must be a subset)

```
# add (git mv from the paths under "remove" where the file exists today; history preserved)
skills/restruct-sdlc/SKILL.md
skills/restruct-sdlc/LICENSE
skills/restruct-sdlc/agents/openai.yaml
skills/restruct-sdlc/references/{roles-lifecycle,budgets,adoption,coordination,validation-integration,closeout-risks}.md
skills/restruct-sdlc/references/{supervision,handoff,local-session-handoff,installation}.md
skills/restruct-sdlc/scripts/{handoff,watch,check_sdlc}.py
skills/restruct-sdlc/scripts/{test_handoff,test_watch,test_check_sdlc,test_claude_only,test_consolidation}.py
skills/restruct-sdlc/compat/legacy/claude-session-handoff/scripts/{handoff,watch}.py
skills/restruct-sdlc/templates/**            (moved unchanged)
.agents/skills/restruct-sdlc                 (relative symlink -> ../../skills/restruct-sdlc)
.agents/skills/claude-session-handoff/scripts/{handoff,watch}.py   (replaced by the forwarders)
management/v0/results/G055-restruct-sdlc-consolidation-result.md
management/v0/results/G055-legacy-path-map.md
management/v0/results/G055-scope-allowlist.txt
# modify (link-path updates and the dated G055 sections only; Q7 disposition: mechanical, for Codex review)
AGENTS.md CLAUDE.md                          (every active link/name to the three retired skills -> skills/restruct-sdlc/...; 16 + 7 hits at planning)
docs/agents/CLAUDE-SESSION-HANDOFF.md        (8 hits incl. helper/skill paths; one dated G055 section incl. the G8a/G8b record)
docs/agents/GITLAB-BRANCH-WORKFLOW.md        (6 hits incl. L257 "skill's closeout procedure" -> skills/restruct-sdlc/references/handoff.md)
# templates and references that move: their 26 + 18 + 7 + 1 + 4 + 4 + 1 hits are retargeted in place (intra-package)
# remove (after git mv)
skills/sdlc-management/**
.agents/skills/repository-supervisor/**
.agents/skills/claude-session-handoff/{SKILL.md,agents/openai.yaml,references/local-session-handoff.md}
.agents/skills/claude-session-handoff/scripts/{test_handoff,test_watch,test_claude_only}.py
```

**Active-link inventory (r1 correction 3).** Step 5 of §9 greps the active governance, runbooks, templates and
references for `claude-session-handoff`, `repository-supervisor` and `sdlc-management` (counts above, taken at
planning) and retargets every active hit; closed Goals/Specs/Plans/Results/reviews and archives are historical and
untouched. V4's P3/P4/A2 plus a final zero-hit grep over the active set (excluding `management/v0/{goals,specs,
plans,results,archive}` and the path-map record, which cites the old paths by design) are the evidence.

**Unchanged** (verified by an empty protection diff in F1): `.gitlab-ci.yml`, `infra/**` (including
`profile-paths.tsv`, whose three exact handoff rows become inert, see §13 R3), `.gitignore`, `skills/saas-foundation/**`,
`.agents/skills/vestive-design`, `management/BUDGET_POLICY.md`, `management/SUPERVISOR-HANDOFF.md`, the Goal, Spec,
ROADMAP and PROGRESS, `core/ web/ intelligence/ contracts/`, `README.md` (no skill paths), `agents/openai.yaml`.
No Beginify code binding is created (Beginify stays repository-local).

### 6.2 Sibling (conditional, uncommitted; §5.9)

`skills/restruct-sdlc/**` (new tree), `MANIFEST.tsv`, `README.md` (listed sections), deletion of
`skills/{sdlc-management,claude-session-handoff,repository-supervisor}/**`; `LICENSE` unchanged.

### 6.3 Private evidence

`$EVID = $HOME/.local/state/platform-evidence/g055-impl/`, outside both checkouts; it never touches
`g053-observer-*`, `g054-*` or the observer state directory. Disposable HOMEs are under `$TMPDIR` only (P-G054-9).

## 7. Schema, contracts and interfaces

- **Database/schema/migrations:** none. Private state: no schema change; `process.json` argv differs after the
  cutover; the new committed binding file exists only in home-mode target repositories (never in Beginify now).
- **Public API/MCP:** none. No application, CI, provider, model, permission or settings change.
- **CLI additions (optional; defaults unchanged):** `handoff.py --repo PATH` (global; required in home mode);
  `handoff.py binding propose`; `check_sdlc.py package --format {canonical,legacy} --public-remote URL`;
  `check_sdlc.py adoption --format … --package PATH --capabilities …`. The two forwarders.
- **New refusal codes (all fail closed, no record, no send):** `executing-path-outside-canonical`,
  `legacy-forwarder-modified`, `legacy-forwarder-target-missing` (forwarder exit 2), `ambiguous-home-copies`,
  `repository-install-present`, `home-code-binding-missing|dirty|mismatch`, `shared-code-in-use`; checker findings
  `mixed-format`, `binding-missing|dirty|mismatch`, `dependency-missing`, status `not-adopted`.
- **Frames, labels, receipt semantics, markers:** unchanged (P-G054-3, P-G054-11).

## 8. Tests, validation and acceptance mapping

Existing coverage inspected: `test_handoff.py` (64 cases; t1–t15 state, pin, route, receipt, creation, compact,
process scan, transport), `test_watch.py` (82; polling, markers, sources, identity, queue, advisory, lifecycle,
recovery), `test_check_sdlc.py` (35; P1–P7, P3t, A1–A6), `test_claude_only.py` (5, P-G054-6; disposable
installation with real subprocesses), the G053 install-loop collision fixtures and export checks (P1–P7, gitleaks,
notice inventory), and `validate.sh` full (B-SKILL-UNCHANGED/B-PROTECTED compare skill roots with `--base`, i.e.
committed content; they need no change when every change is committed before a run). New tests are limited to the
changed boundaries listed below; each is justified by a gap.

| AC | Change | Evidence: reuse / extend / new (path, case or check IDs) | Batch |
|---|---|---|---|
| AC1 | §5.1 package, §5.2 alias | **Reuse** `test_check_sdlc.test_real_skill_passes_all_checks`, `test_p1_*`, `test_p2_line_limit`, `test_p4_unreachable_reference_fails`, `test_bundle_passes`; `check_sdlc.py package skills/restruct-sdlc --literals $EVID/literals.txt`; `wc -l SKILL.md` ≤ 200. **Extend** `test_check_sdlc`: alias accepted (`test_a2_alias_resolves_inside_repo`), escaping/dangling alias fails, no companion install required (`test_fresh_install_single_entrypoint`). | V4 |
| AC2 | §5.1 routing, §5.5 | **Extend** `test_consolidation.test_routing_walkthroughs`: from the disposable repo, a handoff-only status run reads only `handoff.md`-listed resources (recorded file-access set via `strace`-free method: the walkthrough asserts the commands and references it used, not headings), a supervisor-resume walkthrough (reuse the G053 AC10 resume script on `supervision.md`), a coordinator-status walkthrough with the marker classifier (**reuse** `test_watch.test_working_marker_suppresses_history_and_recovery_prose`, `test_explicit_attention_classes_notify_once`). | V3, V2 |
| AC3 | templates moved unchanged | **Reuse** `test_claude_only.test_claude_supervisor_walkthrough` and `test_default_codex_walkthrough_and_legacy_record` (Goal→Plan→Result→independent review→closeout on repos A/B; fields, budgets, ownership, barriers, publication defaults, no self-acceptance), `test_check_sdlc.test_rendered_adoption_passes`, `test_a4_goal_fields_in_order`, `test_existing_application_content_survives_adoption`; `saas-foundation` tree hash unchanged before/after; the epic-closeout compaction checkpoint and Git-first provenance text verified present in the rendered governance (grep assertions added to the walkthrough). | V3, V4 |
| AC4 | §5.3 install/migration/forwarders | **Reuse** G053 install-loop collision fixtures (absent/identical/COLLISION/symlink STOP) on the new package. **New** `test_consolidation.test_install_and_migrate`: fresh single-directory install, second-path adoption, identical rerun, differing file preserved, symlink escape refused, partial migration refused and preserved, interrupted install recovery (rerun), legacy command parity (forwarder argv/stdout/exit equal to canonical for `status`, `check`, `watch.py status`), forwarder refusals (missing target, modified bytes), archived-instruction path map resolves. Gap: no existing test covers installation of a one-skill package or forwarders. | V3 |
| AC5 | §5.4 trust chain | **Reuse** `test_handoff.test_t14_*` (12 cases: legacy processes, renamed/module invocations, uninspectable, linked worktree, real process holding the log, invocation path upgrade) and `test_t15_*` (10 archive/replace cases) against the canonical path. **Extend** t14: alias invocation positive, canonical committed valid positive, forwarder present-and-equal positive; negatives: modified forwarder, uncommitted canonical, stale writer hash, symlinked alias target outside root, executing path outside canonical, held/uninspectable writer; home: binding-approved positive, missing/dirty/mismatch/self-approved-manifest/version-drift negatives, ambiguous copies, repository-install-present; unchanged log/state bytes after every refusal (existing assertion helper). | V1 |
| AC6 | backends/roles unchanged | **Reuse** `test_claude_only.test_claude_only_operations_never_invoke_codex`, `test_unverified_version_or_protocol_sends_nothing`, `test_sender_outside_the_supervisor_ancestry_is_refused_and_not_rerouted`; `test_handoff.test_t4_rejections`, `test_t5_*`, `test_t8_refusals_and_zero_spawns`, `test_t10_*`; G054 closure fixtures. **Extend** one case each through the forwarder route and the home route (wrong/equal/foreign/stale/unbound role → no send/cleanup; status/receipt semantics compared). | V3, V1 |
| AC7 | §5.6 cutover | **Reuse** `test_watch.test_start_duplicate_stop_and_cleanup`, `test_stop_refuses_reused_pid`, `test_fresh_start_skips_saved_backlog_and_keeps_dedup`, `test_fresh_start_does_not_clear_uncertain_intent`, `test_restart_keeps_dedup_and_offsets`, `test_run_loop_is_single_instance_and_stops_on_signal`, `test_default_interval_is_thirty_and_override_is_honoured`, `test_default_threshold_is_100_mib_and_status_reports_it`. **Extend** `test_watch.test_stopped_cutover_preserves_state`: seeded pin/supervisor/binding/history/cursor/uncertainty hashed before/after; stop from legacy path, start from canonical, argv recorded, no replay/duplicate/retry; live relocation (start canonical while legacy runs) refused by the lock; interrupted cutover resumes; second observer refused. Both backends. | V2 |
| AC8 | unchanged | **Reuse** `test_watch` marker/source/echo cases (`test_incoming_sidechain_and_echo_markers_never_notify`, `test_notice_echo_does_not_retrigger_then_genuine_milestone_notifies_once`, `test_mixed_task_working_summary_cannot_hide_review_handoff`, `test_native_question_is_not_hidden_by_working_marker`, `test_invalid_marker_does_not_suppress_a_real_prose_failure`) in both backends via P-G054-6; one assertion added that an alias/forwarder/home invocation changes no role or label. | V2, V3 |
| AC9 | §5.8 checker | **Reuse** `test_check_sdlc` P3/P3t/P5/P6/P7/A2/A6 cases and `test_bundle_readme_link_must_stay_inside`. **Extend**: canonical passes; `--format legacy` passes a 1.x bundle; mixed/incomplete fails under both formats; `--public-remote` exact-URL positive in root README and negatives (other URL, same URL elsewhere, private remote); `--package` home check with binding states; license inventory (two MIT files). Secret/private-state canaries: the private literal list (G053 §5) plus gitleaks on the stage. | V4, V5 |
| AC10 | §5.9 export | **Reuse** G053 B2/V5 procedure: manifest unique/verifying, adaptation-diff invariants, verbatim `cmp`, P1–P7 with literals, gitleaks `[]`, notice inventory, collision fixtures on the staged bundle; sibling guards and post-checks (HEAD, tags, remote, LICENSE, unrelated files, no pending work staged); release/migration notes distinguish publication authority. | V5 |
| AC11 | all | V1–V5 attributable passes on the final candidate (executed counts recorded: 64+n, 82+n, 35+n, 5, new file; every selector run records `Ran N tests`, N > 0). **Compat run (r1 correction 2):** base test files from G054's integrated commit run against the candidate scripts; **before execution** the Result lists, by test name, each base assertion that is path-only (asserts the old `.agents/skills/claude-session-handoff` literal) with its expected outcome (`fail-path-only`) and the named replacement check that carries the same runtime guarantee; every other case has expected outcome `pass`. The run is recorded as it happened: the expected old-layout failures stay **failures/history** (never relabelled a pass), an unexpected failure is a defect, and no behavior assertion is removed to make the run green. F1 scope/protection diffs; MR `fast` automatic; staging gate (**full**, since `skills/` is foundation) after an authorized conditional merge under the owner status rule; no local full default; failed/interrupted history retained. | F1 |
| AC12 | Result | Result maps AC1–AC14 to commits, §2 hashes, commands and statuses; B5 ledger and checkpoint; limitations; source/export table; cleanup receipt; Codex actual-diff ACCEPT and integration disposition; no hidden release or G021 waiver. | F1 |
| AC13 | §5.2 home, §5.4 binding, isolation | **Reuse** P-G054-6 `Disposable` (temp HOME guard, fake CLI/inboxes, trap). **New** `test_consolidation.test_home_two_repositories`: one home install (Codex path plus Claude alias/copy), two Git repositories A and B (B with a linked worktree): discover/check/pin/send/receipt, scaffold (management), supervision resume, record maintenance (`compact` dry run and apply) with approved bindings; package bytes unchanged after every step; writes confined to the selected repository and its private namespace; refusals: missing/invalid `--repo`, foreign role/pin/closure, ambiguous copies, unapproved upgrade; linked worktree shares only its repository's state; shared-code replacement with B's live observer refused, stopped approved upgrade preserves both repositories' records and needs separate binding commits. **Manual evidence (G7):** the owner's real Codex and Claude Code listings showing `restruct-sdlc` from the supported paths on a disposable HOME, recorded as real evidence distinct from the fake controls (Q2, Q6). | V3, M1 |
| AC14 | §5.5 | **Extend** `test_consolidation` within the AC13 fixture: handoff-only on repo A (discover/check/authorized send/receipt) with no Goals/Specs/Plans/Results, no templates, no observer, no supervisor role, no progress publication; only `pin.json`, the ignored delivery record and (home) the binding appear (directory snapshot diff); management-only, supervision-of-existing-records and full routing without silent activation; `adoption --capabilities handoff` passes without management artifacts and marks them `not-adopted`; switching selection preserves files/state/policy; missing dependency refuses without side effects; the home handoff-only target is repo A of AC13 (no new campaign). | V3, V4 |

**Batches** (focused first, inspect command expansion, then the covering suite; fetch/rebase/reconcile and
`bash infra/ci/docker-desktop-mounts.sh` before each; owned cleanup and a clean `git status --porcelain --ignored`
after each; helper tests use host `python3 -B`, the accepted agent-tooling exception):

| Batch | Selector |
|---|---|
| S0 | dispatch gate (path/branch/owner/role), §2 `sha256sum -c`, P-G054 reconciliation record, `python3`/`git`/`claude --version`, sibling preconditions, live observer `/proc` identity (P-G054-9) |
| V1 | `python3 -B -m unittest discover -s skills/restruct-sdlc/scripts -p 'test_handoff.py' -k test_t14_ -k test_t15_ -v` (repeated `-k` values are OR-ed substrings; `Ran N tests` with N > 0 is required and recorded, N = 0 fails the batch); the whole file runs **once** on the final candidate as the attributable covering pass, not as a progression step |
| V2 | `… -p 'test_watch.py' -k TestStoppedCutover -k test_start_ -k test_stop_ -k test_fresh_start_ -k test_known_unsent -v` (new cutover cases live in the class `TestStoppedCutover`; positive count required); whole file once on the final candidate; then the compat run (§8 AC11) |
| V3 | `… -p 'test_claude_only.py' -v` (reuse, unchanged expectations); `… -p 'test_consolidation.py' -k TestInstallMigrate -k TestRouting -v`, then `-k TestHomeTwoRepositories -k TestCapabilities -v` (one class per boundary; positive counts recorded); whole file once on the final candidate; traces to `$EVID/v3/` |
| V4 | `… -p 'test_check_sdlc.py' -v`; `check_sdlc.py package skills/restruct-sdlc --literals …`; `check_sdlc.py adoption . --release v0 --capabilities full`; `wc -l skills/restruct-sdlc/SKILL.md`; disposable repos A (codex) and B (claude) adoption plus `--capabilities handoff` |
| V5 | export stage, checks, gitleaks under `flock "$HOME/.local/state/platform-evidence/local-harness.lock"`, conditional W |
| R1 (optional) | read-only migration/security review (§10) |
| M1 (operator) | AC13 real discovery rows (G7); no session creation; Beginify untouched |
| F1 | fetch/rebase/reconcile; V1–V5 attributable passes on the final candidate (rerun only what changed inputs); scope diff `git diff --name-status <base>..HEAD` ⊆ §6.1; protection diff `git diff --exit-code <base> HEAD -- .gitlab-ci.yml infra core web intelligence contracts skills/saas-foundation .agents/skills/vestive-design management/BUDGET_POLICY.md management/SUPERVISOR-HANDOFF.md .gitignore` empty. **No local `validate.sh --profile full` by default** (r2 review Q1 applies: the harness executes none of the four Python suites; its only `skills/`-touching controls, B-SKILL-UNCHANGED/B-PROTECTED, compare committed content and are exercised by the hosted staging gate). A broad local run needs a recorded technical reason and supervisor disposition before launch |

**CI selection** (`profile-paths.tsv`): `skills/` → foundation (full); `.agents/skills/**` additions/deletions/
renames → agent (fast) or deletion-or-rename (full); `docs/**`, `management/**`, `AGENTS.md`, `CLAUDE.md` → docs.
Hosted union **full**: the MR `fast` job runs automatically, the manual MR gate is not started, and the staging
gate after an authorized conditional merge supplies the required hosted full evidence under the owner status
rule (success/canceled qualifies, failure fails, pending does not qualify). No CI file changes.

## 9. Sequence

0. After G1–G6: the coordinator dispatches one implementer (native isolated worktree, branch
   `g055-restruct-sdlc-consolidation` from current origin/staging; effort high, recorded or unknown). S0.
1. `git mv` the package into `skills/restruct-sdlc/`; write `SKILL.md`, `supervision.md`, `handoff.md`,
   `installation.md`, `openai.yaml`; add the alias; update intra-package links. V4 focused (P1–P4, line count).
2. `handoff.py`/`watch.py` path sites, forwarders, `--repo`, home binding, refusal codes; tests. V1, V2 (incl. compat run).
3. `check_sdlc.py` formats/alias/package/capabilities/public-remote; tests. V4.
4. `test_consolidation.py` (install/migration/routing/home/capabilities). V3. Safekeeping commit and push.
5. Governance/runbook link updates, path map record, scope allowlist; `installation.md` migration/cutover text.
   Docs-only; V4 link checks rerun.
6. V5 stage, checks, conditional W. Safekeeping push.
7. R1 (optional); fix findings; rerun affected batches only.
8. M1 (G7) and operator feedback; resolve findings.
9. F1, Result, commit, push; hand back (attempt 1 ends).
10. After Codex pre-integration ACCEPT: re-check target, rerun affected checks, `glab mr create --target-branch
    staging` (source-branch deletion). **Before the merge: G8a** (owner-authorized PID-safe stop of the live
    Beginify observer, readiness recorded; otherwise HELD, MR open but unmerged). Owner-authorized merge; **G8b**
    canonical restart from the main checkout; staging disposition; final ACCEPT; closeout cleanup.
11. Separately authorized: 2.0.0 publication (§5.9, G9). No live home installation is part of this sequence.

## 10. Delegation

- One implementation/export owner (high); no implementation subagents; no competing writers to source or sibling.
- Optional read-only reviewers (shared B5, no acceptance authority, temp-directory writes only; a native worktree, if
  created, keeps its generated name and is recorded): (a) migration/security review of §5.3–§5.4 (forwarder trust
  chain, binding refusals, precedence, privacy of exported content); (b) disposable-adoption review that replays the
  AC13/AC14 walkthrough from `installation.md` alone. Neither selects roles, approves migration/publication or
  accepts the Goal. The implementer integrates and validates all findings.

## 11. Budget

**B5, checkpoint after attempt 3** (BUDGET_POLICY §7). Forecast **2–3 attempts**: attempt 1 through F1 with the
repository-local consolidation, migration, checker and export; the home-mode binding and multi-repository fixture
are the convergence risk. Inner batches, reviewers and corrections share the allowance; planning consumed none; no
autonomous extension, reset or weaker fallback. **No split is pre-approved (Q8 disposition):** all fourteen ACs,
including AC13/home functionality, stay in scope. If checkpoint 3 shows nonconvergence, the implementer proposes
a bounded split in the Result for explicit Codex Goal/Spec disposition; no partial Goal is accepted by this Plan.
Context: one implementer context is expected to suffice; the work is checkpointed at the end of steps 4 and 6 for
a 350k handoff if needed.

## 12. Owner questions and their dispositions (Codex r0 review; none blocks Plan review now)

| Q | r0 question (summary) | Disposition (r0 review) | Carried into |
|---|---|---|---|
| Q1 | Release line (1.1.0 then 2.0.0 vs one 2.0.0); keep `--format legacy`? | Actual release sequencing is deferred to publication; explicit legacy checker mode is within the reviewed migration scope; 2.0.0 remains a recommendation | §5.8 `--format legacy` kept; §5.9 release line stays a publication-time decision |
| Q2 | Claude home placement: alias vs identical discovery-only copy | Both already permitted by the Spec; preserve executing-code binding and collision refusal; **no live home install authorized** | §5.2 unchanged; §9 step 11 |
| Q3 | Existing personal `~/.agents/skills/claude-session-handoff` | Leave untouched; retirement is outside this dispatch | §5.2 (inventoried only) |
| Q4 | Forwarder set and retention | The two helper/observer forwarders and bounded retention are acceptable once correction 1 is applied; the documented `check_sdlc.py` path change stays explicit | §5.3 (format rule, compatibility directory) |
| Q5 | Binding file `management/RESTRUCT-SDLC-CODE-BINDING.json`, approver | Within the Spec's required repository-owned binding; supervisor review plus existing scoped commit authority suffice; no extra approval framework, no self-trusting manifest | §5.4 unchanged |
| Q6 | Read-only Codex/Claude discovery evidence method | Identify a supported read-only actual harness-discovery method **before F1**; a fixture listing is not real evidence; no provider call, new owner session, settings change or probing campaign during planning | §8 AC13 M1 row (G7): the method is recorded in the Result before F1; absent, AC13's real-discovery row is reported blocked, never substituted |
| Q7 | Mechanical governance/runbook link updates on the task branch | Allowed for review; preserves authority, policy and permissions; grants no implementation before the normal release | §6.1 modify list, active-link inventory |
| Q8 | Pre-approved split fallback | **Not** pre-approved; no removal of AC13/home functionality or partial Goal; at checkpoint 3 propose a bounded split for explicit Goal/Spec disposition | §11 |

No open owner question remains for Plan review; publication-time (Q1) and pre-F1 (Q6) items are tracked above.

## 13. Risks and stops (each stops the affected work and returns it; none is resolved by widening trust)

- **R1 Alias and committed-file checks.** If any accepted helper check cannot be satisfied with the symlink alias
  plus canonical real path (for example a harness that resolves `.agents/skills/<name>` only as a regular directory),
  escalate before copying the package or changing the check (Spec R1/R4).
- **R2 Live observer.** Any contact with Beginify's live pair, pin, observer process or state during tests stops
  work immediately (P-G054-9 guard). Live code is replaced only after the owner-authorized PID-safe stop (G8a,
  §5.6); a missing authorization or a blocked `stop` holds the candidate before integration, and no review,
  integration or acceptance step bypasses that hold.
- **R3 CI path map.** The three exact handoff-file rows in `profile-paths.tsv` and the `.gitlab-ci.yml` omission
  globs become inert when those files are removed; no change is planned. If S0 finds a control that requires
  mapped files to exist, propose the one-row minimal CI change for Codex review before any batch.
- **R4 Compat run deltas.** Base tests that assert the old helper path fail by design under the new path; each is
  listed by name with its replacement. An unexplained compat failure is a defect, not an accepted delta.
- **R5 Home discovery.** Claude Code or Codex not listing the personal placement is reported as blocked for AC13's
  real-discovery row; fake controls never substitute. No settings, permission or CLI change.
- **R6 Shared-code quiescence.** Uninspectable or namespaced processes block home-code replacement; report, do not
  force-stop or skip.
- **R7 Sibling drift.** A sibling HEAD, tag, remote or pending-work change versus G3 means S only.
- **R8 Scope.** Any need for a new installer, runtime, trust service, package manager, session manager, transport,
  permission mechanism, CI redesign or state-schema reset is out of scope: stop and propose separate scope.
- **R9 Staging advances / G054 changes.** Reconcile under §2; a changed accepted G054 interface re-opens the P-G054
  row only. A changed Goal or Spec stops work for Codex.
- **R10 AC7 independence.** G054's live transport proof is unverified; no G055 criterion depends on it and no
  G055 batch re-proves it. If a G055 walkthrough would need a live Claude pair, that step is recorded as out of
  scope, not substituted by a fake.
- **Execution risks.** Shared-lock contention (queue; never cancel another run); Docker Desktop mount exhaustion
  (recover, then launch); gitleaks finding (fix content, never the allowlist); quota (resource stop); native
  classifier refusals (report, do not work around).

**Not escalated.** No architecture, security-model, data-model, public-contract, vendor, provider, model or
permission change is planned. G008/G009 inputs, OS001 and the G021 barrier are unaffected.

## 14. Planning record and revision history

**Reads (all read-only):** governance (AGENTS.md, CLAUDE.md, docs/ARCHITECTURE.md, docs/SECURITY.md, docs/DECISIONS.md),
ROADMAP/PROGRESS/BUDGET_POLICY, the G055 Goal/Spec/review, the dispatch record, G054 Goal/Spec/Plan r1/Result/focused
review/publication follow-up, the status-marker and transport one-offs, the three skills with references, helpers,
checker, tests and templates, the runbooks, `profile-paths.tsv`, `select-profile.sh`, `validate-controls.sh`
fixtures, `run.sh` provenance controls, the sibling checkout (status, HEAD, tags, manifest, README, diff stat; nothing
modified, staged, stashed or cleaned), the personal skill directories (listing only), `claude --version`.

**Not done:** no test, build, install, export, runtime, settings, session, permission or sibling change; no Codex
invocation; no inspection of `~/.local/state/platform-evidence/g053-observer-target/` or any observer state.

**Revision history.**
- r0, 2026-10-16: first provisional Plan on base `834cde1`, branch `g055-plan`. Before its first commit, the
  coordinator's G054 refreshed-assumption message (G054 Plan r2, `f1e0afc`) was reconciled in place: §1 G3/G4, §2
  (r2 row), §3 (P-G054-7 replaced; P-G054-12 to P-G054-15 added; r2 corrections C1–C5 noted), §5.1 line budget,
  §5.9 W preconditions and §13 R10. The G1–G3 gates and a second reconciliation pass against the finally accepted
  G054 source remain before Codex Plan review and ACCEPT.
- r0 (second in-place fold, same day, as a follow-up commit to `6f79f87`): the coordinator reported Codex's
  design-ACCEPT of G054 r2 (review on staging `8714ca5`). Folded: Q2 stage-only (no sibling write path from G054;
  G055 default S; §1 G3, §3 P-G054-7, §5.9), Q5 (1.1.0 future recommendation only; §5.9 release line), Q1
  (focused V1–V5 plus hosted gates, no local full default; §8 F1, CI selection, AC11), the staging advance note
  in the header and the review row in §2. G054 has no implementation release; its final ACCEPT, integration and
  source/export reconciliation still gate G055 Plan ACCEPT and execution. Reviewed as r0 at `d8dbd2a9…`.
- r1, 2026-10-16: bounded correction answering the Codex r0 review (not included)
  (SHA-256 `2d5090ee…9653`; CORRECTION LOOP, planning only, no attempt consumed), on branch `g055-plan` rebased
  onto staging `94604da`; r0 preserved byte-exact as `G055-restruct-sdlc-consolidation-plan-r0.md`
  (`f61c3f85…2538`, 621 lines). Mapping of the four corrections:
  1. migration order and format checks → §5.3 (preflight with non-activating checks, install, live-code gate,
     retirement, final canonical adoption, commit; legacy SKILL-bearing package vs pinned scripts-only
     compatibility directory; mixed/partial refusals; extended existing fixtures);
  2. executable selectors and compatibility truth → §8 batch table (repeated `-k` named selectors, positive
     executed counts, whole file once on the final candidate) and AC11 (pre-enumerated path-only assertions with
     expected outcomes and replacements; expected failures stay failures/history);
  3. home-only rendered references and active legacy links → §5.2 (`SKILL_LOCATION` rendering of references,
     AC13 verification of rendered links/resource access) and §6.1 (active-link inventory incl.
     `docs/agents/GITLAB-BRANCH-WORKFLOW.md:257`; historical records untouched);
  4. live code replacement vs cutover → §1 G8a/G8b, §5.6 (owner-authorized PID-safe stop before the first
     live-code replacement, i.e. before the staging merge; canonical restart after; otherwise HELD; no bypass),
     §9 step 10, §13 R2.
  Q1–Q8 dispositions recorded in §12; §11 no longer pre-approves a split (Q8). Design, scope, AC coverage,
  budget and gates are otherwise unchanged; no test, build or install was run.

## Owner carry-forward — explicit branch and image reconciliation, 2026-10-20

The owner requires acceptance cleanup to audit task-owned remote branches
explicitly; automatic MR source-branch deletion alone is insufficient. Check
all remote task refs, MR state, live use and unique-content reconciliation
before deletion. Reconcile local refs/worktrees independently, and record exact
task-owned Docker image IDs while retaining shared/reused images. The portable
policy is in `docs/agents/GITLAB-BRANCH-WORKFLOW.md`,
`skills/sdlc-management/references/closeout-risks.md`, the adoption guidance and
generated branch-workflow template. G055 carries it into Restruct at the next
safe boundary; no restart, scope/budget change or extra validation campaign.
