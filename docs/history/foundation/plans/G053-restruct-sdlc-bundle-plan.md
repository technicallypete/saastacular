> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G053 Implementation Plan — Reusable SDLC management skill and Restruct bundle

Revision r2, 2026-10-12. Author: Claude Code (G053 planner). Status: **proposed for Codex Plan review.** No
implementation, installation, export, sibling creation, CI change, push or validation run was performed.

r2 answers the Codex r1 review (not included) (CORRECTION LOOP; SHA-256
`d84e8d7f7f33564ff5334a2047c783b5cf0bba51752c41a430cc38a4471d061b`):
- **C1** link classes: §3.2 Link layout, §5.1 P3/P3t/A2, §6 B1/B3/B4, §9 AC1/AC2/AC10;
- **C2** relocation deferral: §3.2 items 8–9, §9.1 (g1)–(g4), §11.

Q1, Q2 and Q4 are resolved by that review (§11). The reviewed r1 is preserved byte-for-byte as
plan-r1 (not included) (SHA-256
`0d78814a873d2af3b97225edf881846cc4c045d8f46e9d17a9bfc913f993f84b`). Bindings (§2) and B3 are unchanged.

- Inputs: [Goal G053](../goals/G053-restruct-sdlc-bundle.md) and [Spec r1](../specs/G053-restruct-sdlc-bundle-spec.md).
- Tier: ordinary bounded tooling/documentation packaging. Budget: **B3, checkpoint after attempt 2** (§10).
- Planning checkout: the generated isolated-agent worktree `.claude/worktrees/agent-<id>`, with this
  planner as the single writer for G053. Branch `g053-plan`. The path is kept until safe handback under the
  previously accepted exception, which the 2026-10-12 AGENTS.md naming tightening names explicitly. It is never
  renamed or moved.
- Base: `origin/staging` `a54d0245a355d44b5c8c4f28a1053dcf4f775c74`, tree `869efc882ed6589ed1d38971beb7acbd90510a1f`.
  Planning started on `21c4d07` and was rebased three times:
  - to `33ff757` (Codex handoff `bd094a1a`): companion/runbook prose for a waiting-state prompt check,
    SUPERVISOR-HANDOFF.md and PROGRESS.md; the Goal and Spec were unchanged;
  - to `d6125e4` (Codex handoff `2dc2bcda`): the owner's worker-naming dispatch gate, in AGENTS.md, CLAUDE.md, the
    handoff SKILL.md, the **Goal** (`e970f23b…` → `09b15789…`) and the **Spec** (`06e713f7…` → `49a03133…`; R5, the
    naming-exception sentence and history; the AC table was unchanged);
  - to `a54d024` (Codex handoff `2adf608b`): the owner's final naming decision, in AGENTS.md, CLAUDE.md, the
    handoff SKILL.md, SUPERVISOR-HANDOFF.md, PROGRESS.md, the **Goal** (`09b15789…` → `3c9fc42a…`; In Scope,
    Execution Policy and history) and the **Spec** (`49a03133…` → `c8401052…`; R5, the AC5 verification text and
    history). The decision: `0-agent-<id>` is reserved for the single active persistent coordinator, workers use
    task prefixes, a name grants no authority, and a live path is never relocated.

## 1. Approach

- Author one new skill, `skills/sdlc-management/`, as the source of record in this repository. It contains a
  SKILL.md of at most 200 lines, six linked references, parameterized templates, and one stdlib package/adoption
  checker with its tests.
- Assemble the local `../restruct-sdlc` bundle from the reviewed allow-list in §3.3:
  - the new skill, verbatim;
  - exported copies of `claude-session-handoff` and `repository-supervisor`, carrying only the sanitizing
    deltas in §4.1;
  - a root README, LICENSE and MANIFEST.
- Every existing source stays byte-identical. That covers the companion skills, helper and tests, the runbooks,
  adopted governance, the application and the generator.
- **No repository CI, harness, profile-map or gate change** (§4.5). The new paths select **full** under the current
  map, so qualification is the existing full profile: locally before the MR, and hosted on staging after
  conditional integration.
- **No** database, schema or migration change. **No** public API/MCP contract, dependency, provider, permission,
  model or application change.

## 2. Bound inputs at `a54d024`

| Path | SHA-256 |
|---|---|
| management/v0/goals/G053-restruct-sdlc-bundle.md | `3c9fc42a1cc8da02b3ce6e36db647fd328c5cc3a430327bd075c7f940ccbb456` |
| management/v0/specs/G053-restruct-sdlc-bundle-spec.md | `c8401052345b6b45ec07b59b826a7a2e4962dab3f7bf0ab5b2d2db271d2f63dc` |
| management/v0/PROGRESS.md | `ec58a86b699e182d7a7ed3a73a999045820ebebcddb07f00a298a6f16d575a27` |
| management/v0/ROADMAP.md | `625ef4485ba984a748ad0f6284d42a1427c3d45607b44852a36694176f857be8` |
| management/BUDGET_POLICY.md | `eb9d458ed5bfda365c6ca4dcef215dbe8b38679dede0d4d53c13d3cd7e210988` |
| management/SUPERVISOR-HANDOFF.md | `8401118d3b5e71916ccc88c92777c87bee5b39ebff87e6344f27065610921df4` |
| AGENTS.md | `c1671d4f66b2683f9271787d1caacc00508a7c2fa45cdeffffc417f860ea2c99` |
| CLAUDE.md | `98a540879a733946d4a2ea9298aeb207359b89571baa59246b390c57b6658d00` |
| docs/ARCHITECTURE.md | `365a998ed4485e32a06b5fe572eab5446bada5c671261f43996310887cdaa975` |
| docs/SECURITY.md (empty; not a waiver) | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| docs/DECISIONS.md | `b23522208c3f8aece05a6f5c930b688b6d44ad181dd632cfeb7ce5c7f40492f8` |
| docs/agents/CLAUDE-SESSION-HANDOFF.md | `4ba737946e229e3ad3a2c3081d4746b7de16add1fad4546e07edea3ccd1e01cb` |
| docs/agents/GITLAB-BRANCH-WORKFLOW.md | `1f5b55c8cde04cf743f7660e0b42a40ce478b8fe18ae3e9619229f73ae28b1d1` |
| .agents/skills/claude-session-handoff/SKILL.md | `34b4ed8f9ded7ddfd3394fbb3e43e42e4b753270b27b6b46e28d6bc7aba00426` |
| .agents/skills/claude-session-handoff/agents/openai.yaml | `c623fdb58e3be79443e5d93f633a4ec1f7783cf63116a2f57ea2736676499107` |
| .agents/skills/claude-session-handoff/references/local-session-handoff.md | `dbb72c27f5670b61d04ba3db7f3a427e29bf931e0664bf1a594371f46367c0dc` |
| .agents/skills/claude-session-handoff/scripts/handoff.py | `e2bdce82472125fe006efa9e5e70dec003c428e386fa971a3442e732724f9d77` |
| .agents/skills/claude-session-handoff/scripts/test_handoff.py | `aa52b1292c2e8bcb9404d392ce90cf73764f499571ca33a4afce6bbff0916f74` |
| .agents/skills/repository-supervisor/SKILL.md | `d6b03c14646147fae643de692e2fe20307154e2e81a7d862b21f0528e481f7dc` |
| .agents/skills/repository-supervisor/agents/openai.yaml | `587f23e27064559066153252d8a56f50d6dc167606d8d1dec12857d4f9ae58e8` |
| skills/saas-foundation/LICENSE (MIT text source) | `af5fb2212e1e79a81443ff705570a4f84c6c587143c84fe4fa590b4aae3667c6` |
| infra/ci/profile-paths.tsv | `bb94a9001b71a0e191a31cb107b427c1cec73358f03e71d50e9ad37932cf0d89` |
| infra/ci/select-profile.sh | `caf5a875936f4f3c56e669567ad3ae092589f8e9c8cbefe2ddb2dfd8f85cca7e` |
| infra/ci/lib.sh | `e34f29ed8d99609132edc390cc13bfb3922754a63e1944a013b9e69a60f507d6` |
| .gitlab-ci.yml | `79b2d2371ac02495f4fe7b95ee386035b6687818e7139f908bf731510af52f84` |
| infra/validation/run.sh | `d357c4ea95a5a6b1ef4e2d3971a2553a300b44b425755086b105775d710ad462` |
| infra/validation/required-checks.tsv | `e2a64f2fe23eba7e90dc25d3ec0b9fdedf4844a375a8979533643086721f3303` |
| infra/security/gitleaks.toml | `df25add66bcf6819e0e94066d318f8692adcdbc692989b26087ceb4b483ece45` |
| .gitignore | `cefaa07d656a45904de0c4b0b68549262b70c68775728746080533dc5b052fb2` |

At execution start, S0 re-verifies these hashes with `sha256sum -c`. A changed row is reconciled before any batch
and recorded in the Result:
- Re-read only the changed sections.
- A changed companion file re-derives its §4.1 delta mechanically.
- A helper behavior change re-runs the export tests.

Inputs are never refreshed silently. Applicable ADRs: none changes. ADR-002, the hosted runner, is reused unchanged.

Superseded bindings from this Plan's earlier drafts are kept for provenance:
- at `33ff757`: Goal `e970f23b…`; Spec `06e713f7…`; AGENTS.md `25ba6e88…`; CLAUDE.md `9ff89a66…`; handoff
  SKILL.md `943d56f5…`;
- at `d6125e4`: Goal `09b15789…`; Spec `49a03133…`; AGENTS.md `cffd307e…`; CLAUDE.md `8d83aa52…`; handoff
  SKILL.md `a557d851…`; PROGRESS.md `679ad92d…`; SUPERVISOR-HANDOFF.md `9d37d8b9…`.

Every other row is byte-identical across `33ff757..a54d024`.

## 3. Files

### 3.1 Repository changes: the complete allow-list against the bound base

- **Add** `skills/sdlc-management/**`, exactly the files in §3.2.
- **Add** `management/v0/results/G053-restruct-sdlc-bundle-result.md`.

No other path changes; the §6 batch B5 scope check enforces this. Plan revisions remain planning artifacts in
`management/v0/plans/`. The skill is not installed into this repository's `.agents/skills/` or any personal
location.

### 3.2 `skills/sdlc-management/`

| Path | Purpose |
|---|---|
| `SKILL.md` | At most 200 lines. Frontmatter `name: sdlc-management`, `description`, `license: MIT`, `metadata.version: "0.1.0"`. It covers purpose and boundaries, when to use the skill, inputs, the adoption procedure, a lifecycle summary, when to load the companions, authority precedence, supported limits and package verification. Every reference and template is linked. |
| `LICENSE` | MIT text, byte-identical to `skills/saas-foundation/LICENSE`. |
| `agents/openai.yaml` | Codex interface metadata (display name, short description, default prompt), in the companions' format. |
| `references/roles-lifecycle.md` | R2: the ownership matrix (owner, supervisor, coordinator/implementer, workers) and the eleven Goal fields. Also the Spec, Plan and Result contents; stable IDs and preserved revision history; the four distinct states (Plan approval, readiness, conditional integration, final ACCEPT); precedence (harness and owner direction over skills and templates); and recorded scope/authority for policy changes. |
| `references/budgets.md` | R3: B3/B5/B7 plus optional B15 and their checkpoints. Goal-level attempts (not commands, batches or inner loops); workers share the allowance. Explicit extensions keep used attempts and record reason, remaining scope and approver; counters never reset. Effort is a configurable recommendation (planning xhigh, implementation high), never a verified runtime claim. |
| `references/adoption.md` | R4: explicit parameters and risk tiers (authorized direct presentation work, concise ordinary Plans, stronger security/data/architecture review). Existing governance is never silently overwritten; existing files receive a `<path>.sdlc-proposed` proposal for an owner merge decision. Bundle storage is distinct from harness discovery, and the helper's committed-path assumptions (§4.1) are documented. Unsupported cases are listed. |
| `references/coordination.md` | R5/R6: exactly one active, owner-selected persistent coordinator. One verified owner per task, non-overlapping file ownership and explicit multi-session routing. Survivors are recovered before replacements; unknown discovery authorizes nothing. Scoped service instructions are read. Receipt, acknowledgment and execution are distinct; peer messages cannot grant approval or change permissions/providers. Transport is delegated to `claude-session-handoff` by link; there is no transport code here. Also the **naming, dispatch and relocation gate**, below this table. |
| `references/validation-integration.md` | R7: the generic policy (fetch/reconcile/rebase before batches and MR; finish a batch, fix, rerun, advance; manual feedback before final qualification; evidence reuse needs unchanged content/inputs). Safekeeping pushes; reviewed MR integration; automated review findings inspected before merge; promotion sources retained; only verified merged temporary branches deleted; supervisor assessment of target advances. Final validation needs success by default. The solo-operator placement and canceled-pipeline rule are explicit options. A concrete GitLab/glab adapter; other providers are adaptation points. |
| `references/closeout-risks.md` | R8: acceptance separated from cleanup. Verify integrated content, unique work and live ownership before removing worktrees, branches or markers, or closing disposable MRs. Session records are archived privately before pruning, through the helper's `compact`. A live coordinator's directory is never deleted. PROGRESS is compacted into a dated archive with working links and preserved failures, and bound-input owners are notified. Performance concerns (query/index/locking) go in RISKS.md without a vendor-specific architecture. |
| `templates/parameters.example.json` | Every parameter with a description and a generic default (§4.4). |
| `templates/root/AGENTS.md.tmpl`, `templates/root/CLAUDE.md.tmpl` | Root governance and the implementer role. Each links the references instead of duplicating them. |
| `templates/root/gitignore.lines` | `/management/**/SESSION-HANDOFFS.jsonl`, which matches the helper's default record path. |
| `templates/management/BUDGET_POLICY.md.tmpl`, `templates/management/SUPERVISOR-HANDOFF.md.tmpl` | The shared budget policy and the release-independent supervisor reference. |
| `templates/management/release/{ROADMAP,PROGRESS,BACKLOG}.md.tmpl` | Release index files, instantiated under `management/{{RELEASE}}/`. |
| `templates/management/release/goals/GOAL.md.tmpl` | The eleven fields in governance order. |
| `templates/management/release/{specs/SPEC,plans/PLAN,results/RESULT,results/REVIEW}.md.tmpl` | The Spec, Plan (with AC mapping), Result (with an Execution Budget ledger, failed/skipped/interrupted checks, and an outstanding-findings checklist) and supervisor review records. A review instance is named `<ID>-implementation-codex-review.md` (§4.1). |
| `templates/management/release/archive/PROGRESS-ARCHIVE.md.tmpl` | The header for `archive/PROGRESS-YYYY-MM-DD.md`: source SHA, original hashes, and preserved text. |
| `templates/docs/RISKS.md.tmpl` | The risk register, with a performance section. |
| `templates/docs/agents/BRANCH-WORKFLOW.md.tmpl` | The integration runbook: generic rules plus a GitLab/glab section. |
| `templates/docs/agents/CLAUDE-SESSION-HANDOFF.md.tmpl` | The local verification record (CLI version, platform, date) plus a pointer to the companion's bundled reference, so protocol text is not duplicated. |
| `scripts/check_sdlc.py` | The stdlib package and adoption checker (§5.1). |
| `scripts/test_check_sdlc.py` | stdlib `unittest` fixtures in temporary directories (§6). |

**Naming, dispatch and relocation gate.** This is required wording, not an option. It appears in
`references/coordination.md` and in the AGENTS.md and CLAUDE.md templates, generalized from the owner's 2026-10-12
rules (the dispatch gate and the final naming decision):

1. **Coordinator name.** `0-agent-<id>` is reserved for the single active persistent coordinator. Its existing agent
   ID is kept when available, and no worker or second session may use the `0-` prefix.
2. **One active coordinator.** Exactly one coordinator is active. This is verified through complete session
   discovery and the owner-selected pin (the helper's `status` must be `pinned-live-verified`). An unknown or
   ambiguous discovery result never justifies creating another coordinator.
3. **A name grants no authority.** Directory names are organizational only. They never establish session identity,
   select a routing target, grant authority, or justify a new or replacement coordinator. Routing and ownership
   come from discovery, the verified pin, and recorded task ownership.
4. **Worker names.** Every new worker needs a task-prefixed isolated worktree: `<task-id>-agent-<id>` in lowercase,
   or another task-prefixed name.
5. **Before dispatch,** the coordinator records the task ID and intended path, and selects a creation route that
   can establish that name.
6. **Before any task write,** the worker verifies and reports the actual Git worktree path and branch (from
   `git worktree list --porcelain`) and its owner (session/agent ID). An unprefixed path fails the check.
7. **Unsupported naming.** If tooling cannot establish the name safely, that dispatch stops. Continuing needs a
   recorded, task-specific exception from `{{SUPERVISOR_AGENT}}` that states the tooling limitation and the actual
   path and owner. A generated name is never accepted silently. Active bound paths are preserved.
8. **Relocation refusal (r2, Codex C2).** No worktree is ever moved beneath a live process or session, and no
   session is restarted or duplicated just for naming. The default is **defer**. Relocating an existing coordinator
   to `0-agent-<id>`, or any rename, may proceed only when **every** precondition below is *positively* verified;
   anything unknown, ambiguous, inaccessible or unverified defers the move.
   - **Complete discovery.** The helper's existing `status` reports complete discovery, never `discovery-unknown`.
   - **Positively verified stopped ownership.** `status` is `pinned-not-live` for the pinned UUID, and the CLI
     listing shows that UUID only as a stopped/done row. No live registration (under any UUID, including a copy or
     fork) has its checkout at or under the path.
     - Not proof of stopped ownership: `pinned-ambiguous`, `pinned-unverified`, `pin-unsafe`, `unpinned`,
       `discovery-unknown`, an invisible PID or restricted namespace, or the mere absence of a
       `pinned-live-verified` result.
     - A missing positive stopped row also defers.
   - **No cwd users and no open-file users.** These are separate facts, and each must be established:
     - *cwd:* readlink `/proc/<pid>/cwd` and `/proc/<pid>/root` for every process;
     - *open files and mappings:* `/proc/<pid>/fd/*` and `/proc/<pid>/maps`, or `lsof +D <path>` where installed.
       A cwd scan alone never establishes the absence of open-file users.

     Either check is unproven, and the move defers, if any process belongs to another user, any entry is
     unreadable, a PID is invisible, a container could mount the path, or `lsof` reports access errors. These are
     standard inspection commands written into the procedure; there is no process-scanner code.
   - **Quiet state.** No validation run or other helper invocation is active, and work and bindings are clean or
     preserved.
   - **Supported CLI.** The current CLI's same-UUID restart from the moved path has been verified locally and
     recorded in the local session-handoff verification record.
9. **Relocation procedure,** only after item 8 passes in full:
   1. Run `git worktree move`, then verify with `git worktree list --porcelain`.
   2. Restart the same UUID from the new path through the recorded supported route.
   3. Reconcile the session path, and record pin provenance: the helper's `status` reports the checkout change
      since the pin, and the pin is re-verified or re-pinned with the replaced provenance recorded.

   Any failure stops and is reported, with no further moves, restarts or session creation.

This is a procedural gate, not a claim that the CLI enforces it. It adds no naming framework, lock bypass,
process-scanner code, helper change or new checker command. Verification uses Git, standard `/proc`/`lsof`
inspection and the helper's existing `status`.

All templates use the `.tmpl` suffix. Agents working under `skills/` therefore never read them as
directory-scoped `AGENTS.md`/`CLAUDE.md` instructions. Templates are distilled anew from the governance in §2; dated
owner-direction narratives and history are not copied.

**Link layout (r2, Codex C1).** Each link is resolved against a declared root. Three classes exist, and none is
exempt from checking.
- **L1 — intra-package links** in shipped documents (SKILL.md, references and the bundle README) resolve inside
  their own skill directory, or for the README, inside the bundle root.
- **L2 — sibling-skill links** take the form `../<skill>/…`, normalized from the skill root. `<skill>` must be one of
  the three bundle skills. They resolve against a declared skills root:
  - the bundle's `skills/`;
  - the adopting repository's installed `.agents/skills/` (companion loading, as `repository-supervisor` already
    does);
  - in this repository, `--skills-root .agents/skills` for the companion sources.

  An unknown skill, a missing target or an escape from the declared root fails.
- **L3 — adoption links** in `.tmpl` files are never resolved in template storage.
  - *Target layout.* Each template's destination follows a fixed rule: strip `templates/`; `root/` maps to the
    repository root; `release/` maps to `management/{{RELEASE}}/`; drop `.tmpl`. Instance templates (GOAL, SPEC,
    PLAN, RESULT, REVIEW, PROGRESS-ARCHIVE) keep their directory and take an ID-based filename.
  - *Root governance targets.* AGENTS.md and CLAUDE.md link to `.agents/skills/sdlc-management/references/<x>.md`,
    `.agents/skills/{claude-session-handoff,repository-supervisor}/SKILL.md`, `management/BUDGET_POLICY.md`,
    `management/SUPERVISOR-HANDOFF.md`, `management/{{RELEASE}}/…` and `docs/agents/…`.
  - *Before rendering,* the checker validates link syntax and parameters.
  - *After rendering,* A2 resolves the links in the disposable adopting repository.

### 3.3 Bundle allow-list: `../restruct-sdlc` (§4.3 resolves the path)

| Bundle path | Origin | Source |
|---|---|---|
| `README.md` | new | What the bundle is, its three skills, flat installation, requirements, verification, adoption, limits and unsupported cases, the license, and its status as a publication candidate (not published). This is the only place with Restruct branding. |
| `LICENSE` | new | MIT text, "Copyright (c) 2026 Vitruvian Technology, Corp." |
| `MANIFEST.tsv` | new | §4.2 format; it lists every other file. |
| `skills/sdlc-management/**` | verbatim | The §3.2 files at the tested commit. |
| `skills/claude-session-handoff/SKILL.md` | adapted | §4.1 |
| `skills/claude-session-handoff/references/local-session-handoff.md` | adapted | §4.1 |
| `skills/claude-session-handoff/scripts/test_handoff.py` | adapted | §4.1 (three fixture constants) |
| `skills/claude-session-handoff/scripts/handoff.py` | verbatim | Byte-identical. |
| `skills/claude-session-handoff/agents/openai.yaml` | verbatim | Byte-identical. |
| `skills/claude-session-handoff/LICENSE` | new | The root LICENSE bytes. |
| `skills/repository-supervisor/SKILL.md` | adapted | §4.1 |
| `skills/repository-supervisor/agents/openai.yaml` | verbatim | Byte-identical. |
| `skills/repository-supervisor/LICENSE` | new | The root LICENSE bytes. |

Nothing else enters the bundle: no management records, Results, evidence, transcripts, pins, logs, `.git`, Vestive
skill, generator, application or governance files.

## 4. Decisions

### 4.1 Export/install compatibility

**Companion deltas (exhaustive).** Line numbers are at `a54d024`. The rules are:
1. Remove originating-project names, history labels and live identifiers.
2. Replace release-specific paths with `management/<release>/`.
3. Make no procedure, refusal, authority, version-claim or behavior change. The waiting-state section, the
   coordinator-naming paragraph and the dispatch-gate paragraph are carried verbatim.

| File | Delta |
|---|---|
| `claude-session-handoff/SKILL.md` | L12: "adapted from the Beginify runbook" becomes "adapted from the originating repository's runbook". L24–25: `G006` becomes `GNNN`. L76, L150, L198, L270, L271: `management/v0/` becomes `management/<release>/`. L189: "staging disposition" becomes "integration-branch disposition". The coordinator-naming paragraph (L119–123) and the dispatch-gate paragraph (L125–132) are carried verbatim. |
| `claude-session-handoff/references/local-session-handoff.md` | L1: the provenance comment drops the source project name. L118: "GitLab/provider credentials" becomes "Git-hosting/provider credentials". |
| `repository-supervisor/SKILL.md` | L10–11: "an explicitly delegated GLM or DeepSeek session" becomes "another explicitly delegated agent session". L64–65: the project invariants (UUID identities, tenancy semantics, G021) become "the adopting repository's canonical invariants, authority semantics and any declared delivery barrier". |
| `claude-session-handoff/scripts/test_handoff.py` | L10–12 hold live identifiers: the current coordinator's session UUID, its predecessor and a real handoff ID. They become the synthetic `00000000-0000-4000-8000-000000000001`, `…002` and `…003`. These are canonical UUIDs that the helper accepts. Test logic is unchanged. |

**Helper path assumptions: documented, not changed.** The helper's behavior constrains adopters as follows:
- `compact` verifies the helper committed at `.agents/skills/claude-session-handoff/scripts/handoff.py`
  (`HELPER_RELATIVE`, read with `git show HEAD:`).
- Task labels match `G\d{3}`.
- Document references must lie under `management/`.
- Closure records match `management/<release>/results/<ID>-implementation-codex-review*.md`, with
  `--verified-by codex|owner`.
- The sender label is `codex-supervisor`.
- Messaging is verified on Claude Code 2.1.290/2.1.291/2.1.296, and the creation gate only on 2.1.296.
- The helper needs Linux `/proc`.

The templates adopt these as defaults: `G###` IDs, root `management/`, `management/<release>/`, and the review
record name. Generalizing any of them would change source behavior, which is out of G053's scope (§11 E3).

**Supported installation.** Copy the three flat skill directories into the target repository's `.agents/skills/`
and commit them. This is the layout tested here: Codex discovers repository skills there, and Claude sessions read
SKILL.md explicitly as the adopted AGENTS.md requires. The bundle's `skills/` directory is storage, not a discovery
location. The following are not promised:
- other install paths (`compact` refuses them);
- native Claude `.claude/skills/` discovery;
- discovery after `cd`;
- personal installation;
- macOS or native Windows;
- other CLIs or versions, which the helper fails closed on.

**Independence.** `sdlc-management` loads alone. The companions are loaded only for session messaging or supervision;
there is no mandatory reload of all three skills. `sdlc-management` contains no transport code and assigns no role.

**Retained comment.** `handoff.py` stays byte-identical, including the L23 provenance comment that cites the internal
task label of the 2.1.296 reverification. Identical executable bytes are the strongest evidence that the safeguards
are retained, and the comment holds no live identifier or private data. Q2 is resolved by Codex: keep the comment
and export the file byte-identical.

### 4.2 Privacy, licensing and notices

**Private literal list.** Generated at run time in the evidence directory; it never enters the bundle or the
repository. It holds:
- the archetype identity values: manifest inputs `product_name`, `project_slug`, `platform_base_domain`,
  `custom_domain_cname_target` and `database_name`;
- the Git user name and email tokens;
- the `origin` host and path;
- the absolute main-checkout path and `$HOME`;
- every full UUID found under `management/` and in the companion tests, with its 8-hex prefix;
- the project-specific terms `G021`, `WorkOS`, `Vestive`, `identity-qualification` and `OS001`.

**Scans.**
- Checker P5–P7 (§5.1), using its built-in patterns plus the list above.
- The accepted gitleaks image `zricethezav/gitleaks:v8.30.1`, with `infra/security/gitleaks.toml`. It runs through
  the same tar-on-stdin form as `validate.sh`'s evidence scan: no bind mount and no allowlist edit. A finding is
  fixed in the content, never allowlisted.
- A notice inventory (`Copyright|SPDX|License`).

**Rights.** All exported content was authored in this repository for Vitruvian Technology, Corp. At the base, the
companions contain no third-party notice lines. The established MIT text comes from `skills/saas-foundation/LICENSE`.
Any third-party-derived text or uncertain right stops the export (Goal stop condition).

**MANIFEST.tsv.** Rows are `sha256<TAB>origin<TAB>path`, with `origin` one of new, verbatim or adapted. The file
holds no source commit, source path or private hash. Private provenance stays in the Result: source commit and
tree, source/export hashes, and the exact unified diff for each adapted file.

### 4.3 Destination ownership

- **Path.** DEST is the main checkout's parent plus `/restruct-sdlc`, derived from
  `git rev-parse --path-format=absolute --git-common-dir`, not from the worktree. That is
  `<home>/Code/VitruvianTech/Beginify/restruct-sdlc`. A read-only check on 2026-10-12 found it absent.
- **Who and when.** The G053 implementer creates it in step D1, after B1–B3 pass. Planning creates nothing.
- **Creation.**
  1. Assemble the bundle in a private stage, `$EVID/stage/restruct-sdlc`, and validate it.
  2. Run `mkdir "$DEST"` without `-p`, so it fails if DEST already exists.
  3. Copy the stage; verify the manifest at DEST.
  4. Record `$EVID/destination.json`: path, pre-existence, UTC creation time and manifest SHA-256.
- **Re-export after a fix.** DEST is owned only if that record exists and DEST matches it exactly: no extra or
  changed files, and no symlink in DEST or its parent. An owned tree moves to `$EVID/superseded/<UTC>/` and is
  recreated as above.
- **Collisions.** Any other state stops the export and escalates: no record, empty but unowned, modified, or
  symlinked. Unowned content is never overwritten, merged into or deleted.
- **Not done:** `git init`, remote creation, push, tag or publication (Q4, resolved by Codex: no sibling Git).

### 4.4 Portable policy and templates

**Parameters.** Templates use `{{NAME}}` placeholders. The parameters, with generic defaults, are:
- `PROJECT_NAME`, `RELEASE` (default `v0`);
- `OWNER_ROLE`, `ARCHITECTURE_AUTHORITY`, `SUPERVISOR_AGENT`, `IMPLEMENTER_AGENT`;
- `CANONICAL_SOURCES[]`, `SERVICE_ROOTS[]` (default empty);
- `INTEGRATION_BRANCH`, `RELEASE_BRANCH`, `PROTECTED_BRANCHES[]`;
- `CI_PROVIDER` (`gitlab` or `other`), `MR_TOOL`, `FAST_CHECK_COMMAND`, `FULL_CHECK_COMMAND`;
- `PRE_BATCH_ENV_CHECK` (optional);
- `FINAL_VALIDATION_PLACEMENT` (`mr-hosted` by default, or `local-before-mr`);
- `CANCELED_PIPELINE_SATISFIES` (default `false`);
- `DEFAULT_BUDGET` (`B5`), `PLANNING_EFFORT` (`xhigh`), `IMPLEMENTATION_EFFORT` (`high`);
- `DELIVERY_BARRIER` (optional; default empty).

**Not generic defaults:**
- the G021 barrier and the product, tenancy, MCP, identity-provider and UI-kit rules;
- dated owner directions and G051 timings or one-off acceptances;
- the G047→G003 contract and identity-qualification branch;
- standard+100 ports;
- Beginify branch and actor names;
- the Docker Desktop mount rule, which becomes the optional `PRE_BATCH_ENV_CHECK`.

**Adoption.** Adoption creates only missing files. An existing file gets a `.sdlc-proposed` sibling plus a recorded
owner merge decision. The checker's adoption mode rejects unresolved placeholders.

**Literals and branding.** The B-LITERALS intent is enforced by the checker on both the skill and the bundle. Only
the bundle README carries Restruct branding.

### 4.5 Validation and CI: proportionate, reusing accepted controls

- **Current classification, stated honestly.** `skills/` is class `foundation`, so every `skills/sdlc-management/**`
  addition selects **full**; Markdown under `skills/` gets no documentation bypass. `management/**/*.md` is docs. The
  change's union is therefore **full**. Workflow admits MR and staging pipelines. No waiver is assumed.
- **Decision: no CI wiring.** No AC requires registering a CI check:
  - The package checks run as named local batches against both the skill and the bundle, and the bundle ships the
    checker for adopters.
  - The existing full profile covers the committed tree. The repository-wide gitleaks `secret-scan` sees
    `skills/**` through B-CHECKS-ROOT and B-UNIT-CHECKS. B-PROTECTED and B-SKILL-UNCHANGED (`validate.sh` passes
    `--base HEAD`) prove a run changed nothing.

  Registering a new required check ID would change existing control expectations, which the Spec does not
  authorize ("no existing check replacement, profile/gate change"). The sites that would change:
  - `infra/ci/validate-controls.sh`: ctl_VP1 `required_count 33`, the remainder case's `required_count 60`,
    ctl_VU1 `33 60 89`, ctl_VU9 `33 42 71`, and any affected-profile count whose suites include the host suite;
  - `docs/foundation/OPERATIONS.md`: the count tables;
  - `infra/validation/required-checks.tsv`: the header.

  It would also need a container runtime for the checker. The result is **zero existing-code change**. This is the
  minimum reading of the Goal's "when required by the approved Plan". Q1 is resolved by Codex: no CI wiring is
  required. Focused package tests plus the mandatory path-selected full qualification are acceptable.
- **Guards unchanged.** B-SKILL-UNCHANGED and B-PROTECTED are not modified. Goal scope is verified by the read-only
  diff/hash check in batch B5, not by `run-approved.sh` with the pre-G053 base. That form would, correctly, report
  the new skill as a skill change.
- **Host Python.** The checker uses host `python3` with the standard library only, the same accepted exception as the
  handoff helper. It is agent tooling, not an application runtime. It always runs with `-B` (no `__pycache__`); an
  untracked file under a skill root would fail B-SKILL-UNCHANGED.

## 5. Interfaces and contracts

None is public: no API/MCP, schema or provider contract.

### 5.1 `check_sdlc.py`

Output is a JSON summary. Exit 0 means pass, 1 findings, 2 usage error.

**`package PATH [--skills-root DIR] [--literals FILE]`** checks a skill directory or a bundle root. The skills
root defaults to `PATH/skills` for a bundle and to PATH's parent for a skill directory:
- **P1:** frontmatter delimited; `name` matches the directory and `^[a-z0-9-]+$`; non-empty `description`.
- **P2:** `sdlc-management/SKILL.md` is at most 200 lines.
- **P3 (L1/L2, shipped documents only; `.tmpl` excluded):**
  - L1 targets resolve inside the containing skill directory, or for the bundle README, the bundle root.
  - L2 targets resolve against the skills root and name only the three bundle skills.
  - A missing target, an escape from the declared root, or an absolute filesystem link fails.
- **P3t (L3, before rendering, for each `.tmpl`):**
  - Every `{{NAME}}` is a key in `templates/parameters.example.json`.
  - Link targets are relative, with no scheme and no absolute path.
  - Substitute the example values and apply the destination rule (§3.2). The normalized target must stay inside
    the adopting-repository root, or a template link escapes and fails.
  - The target must lie in the documented target layout: a path that adoption creates, or
    `.agents/skills/<bundle skill>/<path>`, where `<path>` exists in the package or skills root. A missing target
    fails.
- **P4:** every `sdlc-management` reference and template is reachable from SKILL.md.
- **P5:** no symlink, `.git`, bytecode or special file.
- **P6:** forbidden content: home and user paths (`/home/<name>/`, `/Users/`, Windows user-profile paths), `/run/user/<digits>`,
  non-synthetic full UUIDs, private-key headers, and `gitlab.com/`, `github.com/` or `git@` remotes, plus the
  `--literals` entries, matched case-insensitively.
- **P7 (bundle only):** README.md, LICENSE and exactly the three skill directories. Every file other than
  MANIFEST.tsv is listed exactly once with a matching hash and a valid origin.

**`adoption REPO --release NAME [--literals FILE]`** checks an instantiated repository:
- **A1:** the root governance files, `management/BUDGET_POLICY.md`, `management/SUPERVISOR-HANDOFF.md`, the
  release ROADMAP, PROGRESS and BACKLOG, and the `goals/`, `specs/`, `plans/`, `results/` and `archive/`
  directories all exist.
- **A2 (rendered L3 links):** every relative link in the root governance and under `management/` and `docs/`
  resolves inside REPO, including links into the installed `.agents/skills/`. A missing target or an escape from
  REPO fails. The installed skills are then checked with P3 using `--skills-root REPO/.agents/skills`.
- **A3:** no unresolved `{{…}}` outside `.agents/skills/`.
- **A4:** each Goal file has the eleven fields in order.
- **A5:** `.gitignore` has the session-log line.
- **A6:** the literal scan.

### 5.2 Other contracts

- The template parameters are as listed in §4.4, and MANIFEST.tsv is as in §4.2.
- Link roots and the template destination rule are as in §3.2, Link layout.
- Installation is flat under `.agents/skills/` (§4.1).
- No helper CLI, refusal code, state path or lock changes.

## 6. Validation plan (implementation time; nothing runs during planning)

**Before every batch:**
1. `git fetch origin`.
2. Run `git merge-base --is-ancestor origin/staging HEAD`; if it fails, rebase and reconcile the §2 bindings.
3. Run `bash infra/ci/docker-desktop-mounts.sh`; if it fails, do not launch, and recover per OPERATIONS.md.
4. Run `sha256sum -c` against the binding file.
5. Record the target and candidate SHAs.

**During and after batches:**
- Docker batches run under the accepted shared lock,
  `flock "$HOME/.local/state/platform-evidence/local-harness.lock" …`.
- Evidence goes to `$HOME/.local/state/platform-evidence/g053-<checkout>/` (`$EVID`), outside both the checkout and
  DEST.
- No repository edits during any run (B-RENDER-PLAN).
- After each batch: `git status --porcelain --ignored` is clean, and owned cleanup is verified.
- Batches follow finish → diagnose/fix → rerun failed and affected → advance. The Result keeps an
  outstanding-findings checklist.

| Batch | Runs | Docker/lock |
|---|---|---|
| S0 (prep, not a test) | Before any write, verify the §7 dispatch gate: the intended-path record against the actual path, branch and owner. Bind §2. Check DEST absence (`test ! -e && ! -L`). Record python3, git, docker and `claude --version`. | no |
| B1 package | `python3 -B -m unittest discover -s skills/sdlc-management/scripts -p 'test_check_sdlc.py' -v`: fixtures for P1–P7, P3t and A1–A6, positive and negative. The C1 minimum must pass or fail as stated: a template rendered into the documented layout passes; a missing target, an unresolved or undeclared parameter, and an escaping target each fail, in both shipped (P3) and template (P3t/A2) form. Then `check_sdlc.py package skills/sdlc-management --skills-root .agents/skills --literals $EVID/literals.txt`, and `wc -l` on SKILL.md. Source-helper baseline: `python3 -B -m unittest discover -s .agents/skills/claude-session-handoff/scripts -p 'test_handoff.py' -v`, read-only (59 tests at the base). | no |
| B2 export | Assemble the stage from §3.3 and apply the §4.1 deltas. Run `check_sdlc.py package $STAGE`, using the bundle's `skills/` as the skills root. Run the exported helper tests on the stage, expecting the same count to pass. `cmp` the verbatim files; `diff -u` the adapted files, with the `test_handoff.py` diff confined to L10–12. Run the gitleaks stream scan and the notice inventory. | gitleaks only, under the lock |
| B3 adoption | Disposable repos A and B in `mktemp` (§9 AC2–AC8 walkthroughs), with a local bare "remote" for AC7. The bundle skills are installed under each repo's `.agents/skills/`. Run `check_sdlc.py adoption`, which resolves rendered links in the repo, and the literal scans. Rendered-link negatives are copies of repo A with a deleted target, a leftover `{{X}}`, and an edited `../../` escape; each must fail. Then the AC5 relocation fixtures and the defer walkthrough (non-mutating). | no |
| D1 publish | §4.3 creation, then `check_sdlc.py package "$DEST"`; the manifest hash equals the stage's. | no |
| B4 relocation | `cp -R "$DEST"` to a second temporary path and run the package check. Install into disposable repo C's `.agents/skills/` and run `adoption` and P3 against the installed skills root. Run the installed helper tests, bootstrap from `sdlc-management`, and run the supervisor-resume exercise (§9 AC10). | no |
| B5 final | After fetch/rebase/reconcile and on a clean tree: `bash infra/ci/validate.sh --profile full` (README `platform:ci`). Scope check: `git diff --name-status <base>..HEAD` ⊆ §3.1, and `git diff --exit-code <base> HEAD -- AGENTS.md CLAUDE.md docs .agents skills/saas-foundation infra .gitlab-ci.yml .gitignore .archetype core web intelligence contracts management/BUDGET_POLICY.md management/SUPERVISOR-HANDOFF.md`. B1 reruns only if its inputs changed; otherwise it is reused with provenance. DEST's `skills/sdlc-management` matches the tested HEAD's blobs. | yes, under the lock |

- **Manual and operator feedback.** The Spec requires none. Stage R1, the behavioral review (§8), is agent-performed
  with isolated data. It runs with B4, before B5. If Codex requires operator feedback, it takes R1's position, and
  B5 waits for it.
- **Result commit after B5.** It changes only `management/**/*.md`, which is docs class. B5 evidence is reused with
  provenance: the tested commit, plus a delta that touches only the Result.
- **Hosted.**
  - The MR `fast` job runs automatically. The MR `gate` is manual and is not started.
  - After the owner authorizes the merge, the staging push runs `fast` plus the full remainder in `gate`
    automatically. Under the owner status rule, `success` or `canceled` qualifies, a failure blocks, and
    pending/running does not qualify.
  - Temporary-branch safekeeping pushes run no pipeline.
  - GitLab quota exhaustion is a resource stop, reported as such.

## 7. Sequence

0. After Codex Plan ACCEPT, the coordinator dispatches one implementer under the naming, dispatch and relocation
   gate (§3.2, items 4–7), with no generated-name fallback:
   - Before dispatch, record task G053 and the intended path `.claude/worktrees/g053-agent-<id>`, and select a
     creation route that establishes it.
   - Before any task write, the worker reports `git worktree list --porcelain` for its path, the branch
     `g053-sdlc-bundle` (from `origin/staging`), and its session/agent ID.
   - An unprefixed path stops that dispatch until Codex records a G053-specific exception.
   - G053 never touches, renames or relocates the coordinator's checkout. Moving it to `0-agent-5c6f436` is the
     owner's separate operational step at a supported safe point.

   Implementation effort defaults to **high**; remind the operator, and record actual effort or "unknown". Then run
   S0.
1. Author `skills/sdlc-management` (references, templates, checker and tests). Run B1.
2. Stage the export (copies, §4.1 deltas, README, LICENSE, MANIFEST). Run B2.
3. Run the adoption and walkthrough exercises, B3.
4. D1: publish to DEST under §4.3.
5. Run B4, plus the optional R1 reviewer. Fix findings and rerun the affected batches; a bundle change re-exports
   under §4.3.
6. Commit and push a safekeeping checkpoint (standing authority). Fetch, rebase and reconcile, then run B5.
7. Write the Result and commit it; push the checkpoint. Hand back to Codex; attempt 1 ends.
8. After Codex's pre-integration ACCEPT:
   - Re-check the target; if it advanced, rebase and rerun only the affected checks.
   - Push, then run `glab mr create` to `staging`, keeping source-branch deletion (standing authority).
   - Merge only with owner authorization.
   - Record the staging disposition. Codex gives final ACCEPT.
   - Close out per the runbooks: verify branch deletion, then remove the worktree.

## 8. Delegation

- There is one implementation owner and no implementation subagents.
- One optional read-only reviewer may run stage R1. It reviews privacy, portability and authority boundaries, and
  performs the AC10 relocation and supervisor-resume exercise using only the published bundle copy and a disposable
  repository.
- The reviewer writes only in its own temporary directory, reports findings to the implementer, shares the current
  attempt, and has no acceptance authority.
- The reviewer is dispatched without a Git worktree and makes no repository writes. If tooling would create a
  worktree for it, the same dispatch gate applies: `g053-review-<id>` with verification, or a recorded Codex
  exception.
- The implementer integrates and validates everything. Reviewer effort is inherited; verify it where possible,
  otherwise record "unknown".

## 9. Acceptance-criteria mapping

r2 keeps this mapping. It extends AC1, AC2 and AC10 with the C1 link-class evidence, and AC5 with the C2 defer and
open-file cases. The other eight rows are unchanged.

All twelve ACs are remapped against Spec `c8401052…`. Only AC5's verification text changed. It now requires the
documented `0-agent-<id>` coordinator and task-prefixed worker examples, one active coordinator, safe-relocation
refusal for a live checkout, and that naming alone never selects or authorizes a session. R5 adds the matching
export requirement. The other eleven rows are unchanged in the Spec. AC1, AC2 and AC10 gain inspections of the
naming gate; the rest are re-checked and unaffected.

| AC | Deliverables | Evidence |
|---|---|---|
| AC1 | SKILL.md (≤200 lines), six references, three distinct entry points | P1–P4 and P3t on the skill (with `--skills-root .agents/skills`) and on the bundle (skills root `skills/`), using the r2 link classes; `wc -l`; reference inventory. Boundary review: no transport code or role assignment in `sdlc-management`; companions link rather than duplicate; no mandatory reload of all three skills. The naming gate is consistent across `coordination.md`, the templates and the carried companion paragraphs, and no text treats a name as authority. |
| AC2 | The §3.2 templates | Repo A adoption, with the skills installed under `.agents/skills/`: A1–A5. Rendered links resolve in repo A (A2). The C1 negatives (a missing target, an unresolved parameter, an escaping target) fail, both in the checker tests and in copies of repo A. Inspect the created paths and links, the supervisor reference outside the release directory, the gitignored session log, all eleven Goal fields, and the naming, dispatch and relocation wording (items 1–9) in the instantiated AGENTS.md and CLAUDE.md. The literal scan shows no project records. |
| AC3 | `budgets.md`; the RESULT template's ledger | Synthetic G001 (B3) in repo A. Attempt 1: a failed batch and a worker contribution, in the same attempt. Attempt 2: a correction. Attempt 3: exhausted. Then an extension record (previous and new budget, used attempts, reason, remaining work, approver). The ledger keeps every failure and counts no commands. |
| AC4 | `adoption.md`; parameters | Repo B with a different project, `RELEASE=r1`, and `develop`/`release` branches, plus a pre-edited AGENTS.md. Its hash is unchanged and a `.sdlc-proposed` file exists. A2 links resolve. No barrier or vendor literals. |
| AC5 | `coordination.md` and the template wording (gate items 1–9); the exported helper with its carried naming and dispatch paragraphs | The exported helper tests pass, matching the source baseline. **Documented examples:** a `0-agent-<id>` coordinator and `g002-agent-<id>` workers appear in `coordination.md` and the instantiated AGENTS.md. Fixtures (a)–(h), listed in §9.1 below. |
| AC6 | `coordination.md` | The helper's refusal and receipt tests (newer-peer block, `delivery-unconfirmed`, `queued_observed` is not a receipt). A walkthrough: a queued delivery, and a peer request to change permissions that is refused. No live configuration change and no session creation. |
| AC7 | `validation-integration.md`; the BRANCH-WORKFLOW template | Repo A with a local bare remote. It records target and candidate SHAs and the rebase, a failed batch followed by fix, rerun and advance, and manual feedback before final qualification. The review → MR → conditional integration → promotion route uses glab commands that are printed, not run. A merged temporary branch is deleted; long-lived branches are kept. No hosted actions. |
| AC8 | `closeout-risks.md`; the archive and RISKS templates | Repo A fixtures. Closed G001 is archived before compaction, with links verified by A2. Active G002 and a dirty active worktree are preserved; only a clean merged worktree is removed. RISKS.md records a query/index concern. The helper's `compact` tests cover archive-before-prune. |
| AC9 | MANIFEST, LICENSE files, the §4.2 scans | P5–P7; gitleaks returns `[]`; the notice inventory. No `.git`, local state or symlink. The B5 scope and protection diff is empty outside §3.1. Companion hashes equal §2. B-PROTECTED and B-SKILL-UNCHANGED pass in the full run. No existing-code change. |
| AC10 | README installation and limits | B4 from a second path: package check, flat install, P3 and A2 resolution against the installed `.agents/skills/` layout, installed helper tests, bootstrap including the naming, dispatch and relocation procedure, and supervisor resume (the reviewer reconstructs the active Goal and next gate from repo C records alone). Records the platform (Linux/WSL), the Python and Git versions, the verified CLI versions and the unsupported cases from §4.1. |
| AC11 | — | B1 and B2 (checker tests; exported helper tests); B5, the full local profile on the final candidate; MR `fast`; the staging full gate disposition. Failures, exact inputs and cleanup are preserved. No CI exemption is implied. |
| AC12 | The Result | The AC map, commands and statuses, source/export hashes and diffs, the attempt ledger, limitations, and the DEST record. Codex reviews the actual diff and bundle. |

### 9.1 AC5 fixtures

**Fixtures in repo A**, using real `git worktree add` / `git worktree list --porcelain`: (a) intended path `g002-agent-x` is recorded; the actual path, branch and owner are verified; writes proceed. (b) An unprefixed `agent-y` fails verification; dispatch stops with no task write, and an exception request is recorded. (c) A recorded task-specific exception lets that one dispatch proceed with its actual path recorded. (d) A worker worktree named with the `0-` prefix is rejected as reserved. (e) **One coordinator:** a second "coordinator" candidate, or `discovery-unknown`, creates no coordinator and dispatches nothing; the exported helper's creation-gate and routing tests (`reusable-live-coordinator`, `ambiguous-candidates`, `discovery-unknown`, default route only on `pinned-live-verified`) support this. (f) **Name ≠ authority:** a worktree renamed to `0-agent-z` without a verified pin is not routed to or treated as coordinator; routing follows the pin and discovery only.

(g) **Live-path refusal (r2):**
- **(g1) live cwd:** a background process has its cwd inside `0-agent-old`. The cwd inspection detects it, and the move is refused.
- **(g2) open-file holder:** a background process has its cwd elsewhere but an open descriptor on a file under the path (`exec 3<file; sleep`). The cwd inspection alone passes; the fd/`lsof` inspection detects the holder, and the move is refused.
- **(g3) non-mutating defer walkthrough:** each of these inputs leads to *defer*, and `git worktree list --porcelain` is unchanged before and after: `discovery-unknown`, `pinned-unverified`, `pinned-ambiguous`, `unpinned`, a `pinned-not-live` result with no positive stopped CLI row, an unreadable or other-user `/proc` entry, and `lsof` unavailable together with incomplete `/proc` access. The status inputs are recorded synthetic `status` JSON shapes taken from the exported helper's test fixtures; no real session is queried or changed.
- **(g4) allowed move:** only with every item-8 precondition positively satisfied. The status input is a synthetic fixture state with a verified stopped row; the process checks are real and clean; the supported-CLI record is synthetic and labelled as such. Then `git worktree move` runs and is verified, and pin-provenance reconciliation is recorded. Evidence for that reconciliation is a scratch probe in `$EVID` that reuses the exported test fixtures to show `status` reporting `changes.checkout`. If the fixtures cannot be reused, record an inspection of `handoff.py` L720–725 instead.

(h) An active bound path is neither renamed nor moved. **Recovery records:** a surviving worker is reused; explicit multi-session routing is retained.

None of these changes the helper or its tests, starts a CLI compatibility campaign, or relocates a real coordinator.

## 10. Budget

B3: at most three shared Goal-level attempts, with a **checkpoint after attempt 2** (BUDGET_POLICY §7). One or two
attempts are expected. Batches, inner fix loops and the reviewer do not consume attempts. Planning consumed none.
Exhaustion stops work for Codex. The work fits one attempt and one context; no split is recommended.

## 11. Risks, escalations and open questions

- **Q1 (CI enforcement): RESOLVED** by the Codex r1 review. No CI wiring is required; focused package tests plus
  the mandatory path-selected full qualification are acceptable. The r1 alternative site is withdrawn and not
  planned.
- **Q2 (helper comment): RESOLVED.** Keep the comment and the executable bytes; export `handoff.py` byte-identical.
- **E3 (helper constraints).** The `G\d{3}`, `codex-review` record name, `codex-supervisor` and install-path
  constraints stay as documented. Generalizing them is a source behavior change and needs separate scope.
- **Q4 (sibling Git): RESOLVED.** No `git init` in DEST. The review is reproduced from MANIFEST and the Result.
- **Implementer naming.** If the coordinator's isolated-agent tooling cannot establish `g053-agent-<id>` safely, the
  G053 implementation dispatch stops before any write. Codex must record a G053-specific exception first (§7 step
  0). This planner's own path is already covered and is not moved.
- **Coordinator relocation (C2, r2).** The exported procedure (§3.2 items 8–9) defers by default. Relocation needs
  all of:
  - complete discovery;
  - positively verified stopped ownership (`pinned-not-live` plus a stopped CLI row; any unknown, ambiguous,
    inaccessible or unverified state defers);
  - separately proven absence of cwd users and of open-file users (a cwd scan alone is insufficient; an unproven
    inspection defers);
  - a quiet state;
  - a locally recorded, supported same-UUID restart from the moved path.

  The CLI restart has not been verified during planning, and G053 relocates no real coordinator, so the bundle
  makes no compatibility promise for it.
- **Destination and source changes.**
  - An unowned or changed DEST stops the export (§4.3).
  - A further companion or source change before execution is reconciled mechanically (§2). Any need to change a
    source's behavior stops for scope review.
- **Execution risks.**
  - Shared-lock contention: queue; never cancel another task's run.
  - Docker Desktop mount exhaustion: stop and recover.
  - A gitleaks finding: fix the content, never the allowlist.
  - Hosted quota: a resource stop.
- **Not escalated.** No architecture, security-model, migration, public-contract, vendor or permission change is
  planned. The G021 barrier, G008 and OS001 are unaffected.

## 12. Revision history

- r1, 2026-10-12: the initial Claude Plan. It was drafted on `21c4d07`, reconciled at `33ff757` (handoff
  `bd094a1a`; prose-only companion change, no new delta), and reconciled again at `d6125e4` (handoff `2dc2bcda`)
  before handback:
  - Goal `e970f23b…` → `09b15789…`; Spec `06e713f7…` → `49a03133…`;
  - AGENTS.md, CLAUDE.md and the handoff SKILL.md re-bound;
  - the dispatch gate added to the exported procedures (§3.2) and to G053's own dispatch (§7, §8);
  - the companion delta line numbers shifted;
  - AC5 remapped, with AC2 and AC10 evidence extended.

  That state was committed as `f0a65c2` on `d6125e4` and reported to the coordinator, with Plan SHA-256
  `f33f727865836334086041c045bbc71ad3ec8aac681dba32a09c257595a4857d`. It was not reviewed by Codex. Rebasing onto
  `a54d024` carried it unchanged as commit `3d009bc`.
- r1 reconciliation, 2026-10-12: reconciled at `a54d024` (handoff `2adf608b`), before Codex review:
  - Goal `09b15789…` → `3c9fc42a…`; Spec `49a03133…` → `c8401052…`;
  - AGENTS.md, CLAUDE.md, the handoff SKILL.md, SUPERVISOR-HANDOFF.md and PROGRESS.md re-bound;
  - the gate became the naming, dispatch and relocation gate: `0-agent-<id>` for the single coordinator, exactly
    one active coordinator, name ≠ authority, and live-path relocation refusal with a fail-closed safe point;
  - the superseded `COORDINATOR_NAME` parameter (a stable role-based name) was removed;
  - the companion delta line numbers shifted, and the new coordinator-naming paragraph is carried verbatim;
  - all twelve ACs remapped, with AC5 rebuilt as fixtures (a)–(h);
  - G053 does not relocate the coordinator (§7, §11).

  The earlier committed state remains in Git history as `3d009bc`, the preceding commit on `g053-plan`, so no
  revision file is archived. No separate planning inventory is needed: the binding table is §2.

  That reconciled state was committed as `6159e14` (Plan SHA-256
  `0d78814a873d2af3b97225edf881846cc4c045d8f46e9d17a9bfc913f993f84b`). It is the reviewed r1, archived
  byte-for-byte as plan-r1 (not included).
- 2026-10-12 — Codex Plan r1 review: **CORRECTION LOOP**
  (review (not included), SHA-256
  `d84e8d7f7f33564ff5334a2047c783b5cf0bba51752c41a430cc38a4471d061b`, handoff `79eee591`). It found two bounded Plan
  corrections, C1 and C2, and resolved Q1 (no CI wiring), Q2 (keep the helper's bytes) and Q4 (no sibling Git). The
  planning correction consumes no implementation attempt; B3 is unchanged.
- r2, 2026-10-12, at the same base `a54d024` with unchanged §2 bindings:
  - **C1:** three declared link classes (§3.2 Link layout). L1 intra-package and L2 sibling-skill links resolve
    against their package, bundle or installed skills root. L3 `.tmpl` adoption links are validated before
    rendering (declared parameters, relative syntax, the destination rule, target layout, no escape) and resolved
    in the adopting repository after rendering (A2).
  - **C1 evidence:** the positive render, plus missing-target, unresolved-parameter and escaping-target negatives,
    in the checker tests and in repo A copies.
  - **C2:** relocation defers by default. It requires complete discovery, positively verified stopped ownership,
    separately proven absence of cwd and open-file users, a quiet state and a recorded supported same-UUID restart.
    Unknown, ambiguous, inaccessible or unverified states defer.
  - **C2 evidence:** (g1) live-cwd refusal kept; (g2) open-file holder refusal added; (g3) non-mutating defer
    walkthrough added; (g4) fully verified synthetic move.
  - Q1, Q2 and Q4 are marked resolved.
  - The AC5 fixtures move to §9.1 so the AC table stays valid Markdown.

  There is no new framework, helper change, process-scanner code, live relocation or pre-approval test run.
