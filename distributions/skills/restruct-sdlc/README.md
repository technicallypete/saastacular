# Restruct SDLC

An agent-driven software delivery workflow for repositories using Codex and Claude Code. Restruct SDLC bundles
planning, implementation review, progress tracking and local session coordination so work can continue from
repository records instead of relying on a long chat history.

The human owner sets product direction and authority. A supervisor maintains Goals and Specs and reviews Plans
and Results. A coordinator organizes implementation workers, validation and handoffs. Role assignments are
explicit: the owner can choose Claude as supervisor and use Codex for specifications and architecture reviews.

## What's included

One `restruct-sdlc` entrypoint routes to the child skill needed for the current task. The three existing skills
remain complete children of the bundle, with their own instructions, references, templates and scripts.

| Skill | Capabilities |
|---|---|
| `sdlc-management` | Adopt the workflow; bootstrap management directories and project guidance; write Goal, Spec, Plan, Result and review artifacts; apply budgets, validation, integration, release and cleanup policies. |
| `repository-supervisor` | Recover supervision from committed records; review Plans and implementation evidence; manage dependencies, acceptance and progress; relay decisions to the coordinator. |
| `claude-session-handoff` | Discover and verify existing local Claude sessions; pin the intended coordinator; send authorized handoffs and check receipts; observe milestone notices; archive terminal-task handoff records. |

## Workflow features

- **Repository-based continuity.** Roadmap, progress, backlog, task artifacts and dated archives keep decisions,
  failures and acceptance history available across sessions and agents.
- **Clear ownership and bounded autonomy.** Separate product decisions, specification, planning, implementation
  and review responsibilities. Goals define scope, budgets, stop conditions and validation requirements.
- **Risk-based delivery.** Small presentation changes can use an authorized direct tier. Features use concise
  Plans; architecture, identity, authorization, migrations and public contracts retain stronger review gates.
- **Evidence-based acceptance.** Reviews examine the actual diff and validation evidence. A merge or passing
  test alone does not complete a Goal. Repair records preserve the full failure inventory and its disposition.
- **Integration and releases.** Parameterized branch and validation rules cover focused checks, review, merges,
  coordinator-owned affected post-merge suites, owner-requested or supervisor-recommended full local staging runs,
  full hosted RC qualification and owner-controlled promotion. The concrete
  hosted adapter documents GitLab; other providers require explicit adaptation.
- **Efficient coordination.** Milestone observation, explicit owner-decision notices, bounded context and safe handoffs reduce
  repeated status reads. Supervisors send needed decisions promptly; routine coordinator polling belongs to
  the observer rather than the supervisor.
- **Planned parallel implementation.** Goals/Specs record priority from owner direction or supervisor judgment,
  normally high. Plans target 3–5 implementation workers for high or 1–3 for low whenever feasible within
  ownership, dependency, native/resource and budget limits. Kickoff follows the accepted allocation.
- **Task cleanup.** Track task-owned worktrees, branches, containers, images and evidence; remove accepted-task
  leftovers while preserving active work and explicitly held proof.
- **Project resource attribution.** Require project/run labels for exclusively owned Docker resources created
  by Compose or direct commands; use explicit Compose project names and retain compact exact-ID/recovery records
  for anonymous/shared exceptions. Label discovery adds no separate cleanup manifest or removal authority.
- **Project-local configuration.** Adoption guidance covers shared and local Claude permissions, optional Codex
  model defaults, ignored private files under the integration checkout's `.local/`, and supported home-state
  fallbacks for isolated workers.
- **Workflow maintenance.** Changes to reusable rules also update the adopting repository's corresponding live
  guidance, so the current project receives the change as well as future adopters.

## Distribution layout

```text
sdlc-skill/
  README.md
  LICENSE
  MANIFEST.tsv
  restruct-sdlc/
    SKILL.md
    agents/
    sdlc-management/
    repository-supervisor/
    claude-session-handoff/
```

`MANIFEST.tsv` records packaged file hashes and origins. The bundle's parent skill is the discovery entrypoint;
the children are loaded through it rather than installed separately.

## Install and adopt

1. Copy the entire `restruct-sdlc/` directory into your repository at `.agents/skills/restruct-sdlc/` and commit
   it. Keep its contained files together. If an installation already exists, compare it and agree on an update
   before replacing anything; do not copy through a symlinked destination.
2. For Claude discovery, optionally create the relative alias
   `.claude/skills/restruct-sdlc -> ../../.agents/skills/restruct-sdlc` where no conflicting entry exists.
3. Ask your agent to use `restruct-sdlc` to adopt the workflow. Start with the bundled
   `sdlc-management/references/adoption.md` and `sdlc-management/templates/parameters.example.json`.
   Choose your roles, canonical documents, release directory, branches, validation commands and budgets.
4. Review the adoption changes and run the supplied checker. Adoption creates missing files and proposes
   changes for existing files instead of silently overwriting project governance.

Adoption can establish:

```text
AGENTS.md
CLAUDE.md
docs/agents/
management/
  SUPERVISOR-HANDOFF.md
  BUDGET_POLICY.md
  <release>/
    ROADMAP.md
    PROGRESS.md
    BACKLOG.md
    goals/
    specs/
    plans/
    results/
    archive/
```

The templates also cover ignore rules and an optional `.codex/config.toml`. Existing project documents and
settings are reconciled explicitly. Installation alone does not bootstrap a project, assign an agent role,
start sessions or enable an observer.

## Requirements and limits

The supplied helpers use Python 3 and its standard library. Git and an agent environment that can read local
skills are required for the repository workflow. Configure your project's own build, test and CI commands;
the bundle does not provide an application test suite or install application dependencies.

The Claude session helper additionally requires Linux process visibility and a compatible local Claude Code
interface. Its transport uses a version-sensitive local interface; consult its bundled compatibility and
handoff documentation before enabling it. macOS and native Windows session transport are not supported by
this package. Session creation, permission changes, merges and publication still require their applicable
owner authority and native tool permissions.

Earlier installations with three separate top-level skill directories should migrate to the whole bundle,
retarget project references, and retire old copies under an explicit adoption decision. Personal skill
installation is not the supported default; the committed repository copy keeps the team on the same version.

## License

MIT. See the distribution's `LICENSE` and the notices retained in the bundled skills.
