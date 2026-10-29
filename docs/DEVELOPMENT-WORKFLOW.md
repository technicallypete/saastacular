# Development workflow

SaaStacular's default development workflow is **Restruct SDLC**, an agent-driven software delivery workflow bundled
at `distributions/skills/restruct-sdlc/` as the repository's default development dependency. It is distributed under
its own MIT license ([distributions/skills/restruct-sdlc/LICENSE](../distributions/skills/restruct-sdlc/LICENSE),
copyright Vitruvian Technology, Corp.), separately from the repository's Apache-2.0 license, and it is not
relicensed.

The workflow governs agent-driven development, not the application. Configuration, builds, tests, migrations, local
startup and runtime work without an AI account, an agent session or a watcher; every command in the
[README](../README.md#quick-start) runs without agent tooling. The presence of the files scaffolds no management
records, installs no personal skills, creates no agent sessions and activates no watcher. This bundle is a pinned
copy; its upstream project remains its source of truth, and updating the bundled copy is an explicit, reviewed
version change.

## What ships

- **The package.** One entrypoint, [`distributions/skills/restruct-sdlc/SKILL.md`](../distributions/skills/restruct-sdlc/SKILL.md),
  routes to exactly one of three complete child skills:

  | Child skill | Capabilities |
  |---|---|
  | `sdlc-management` | adopt the workflow; bootstrap management directories and project guidance; write Goal, Spec, Plan, Result and review records; apply budget, validation, integration, release and cleanup policies |
  | `repository-supervisor` | resume supervision from committed records; review Plans and implementation evidence; manage dependencies, acceptance and progress |
  | `claude-session-handoff` | discover and verify local Claude Code sessions; send authorized handoffs and check receipts; observe milestone notices |

  [`distributions/skills/restruct-sdlc/README.md`](../distributions/skills/restruct-sdlc/README.md) describes the
  workflow and its installation in detail.
- **The live rules.** [AGENTS.md](../AGENTS.md) and [CLAUDE.md](../CLAUDE.md) carry the workflow's full governance for
  this repository: roles and ownership, Goal/Spec/Plan/Result formats, planning, implementation, review outcomes,
  budgets, parallel scheduling, validation selection, integration, evidence, resource accountability and cleanup.
  The runbooks under [docs/agents/](agents/) hold the GitLab branch workflow, Claude Code session handoffs and
  project permission placement.
- **Codex project defaults.** [`.codex/config.toml`](../.codex/config.toml) sets the repository's model defaults for
  Codex sessions (main sessions `gpt-6.1-sol` at `low`; spawned agents `gpt-6.1-sol` at `high`). It is
  developer-tool configuration, not an application or runtime dependency; explicit session settings override it.

## Discovery

The repository provides two discovery aliases, both relative symlinks to the same bundle:

- `.agents/skills/restruct-sdlc` → `../../distributions/skills/restruct-sdlc` (Codex and other agents that scan
  `.agents/skills`);
- `.claude/skills/restruct-sdlc` → `../../distributions/skills/restruct-sdlc` (Claude Code project sessions).

A personal skill with the same name may take precedence over the project skill; keep personal copies identical or
remove them so the repository source is used. A native Windows checkout without symlink support gets a small text
file in place of each alias; Linux and WSL are the supported runtime.

## Use the default workflow

Ask your agent to use `restruct-sdlc` to bootstrap your management records, starting from
`distributions/skills/restruct-sdlc/sdlc-management/references/adoption.md` and
`distributions/skills/restruct-sdlc/sdlc-management/templates/parameters.example.json`. Record your owner and agent
roles, canonical documents, branches, validation commands and budgets; the adoption guide creates only missing
files and proposes changes beside existing ones (AGENTS.md and CLAUDE.md already hold the rules, so review any
proposal against them rather than overwriting). Point its validation commands at this repository's existing entry
points (`infra/ci/validate.sh` profiles and the README blocks) rather than adding a parallel harness. A watcher or
coordinator binding is activated only when you explicitly establish and verify it. The rules keep the dated owner
directions and examples of the project they were developed in; treat their role names, branch names, budgets,
worker targets, model and effort choices and test-run concurrency as defaults, and record your own choices in
AGENTS.md and CLAUDE.md.

## Use another workflow, or opt out fully

1. **Disable discovery.** Delete both aliases, `.agents/skills/restruct-sdlc` and `.claude/skills/restruct-sdlc`
   (and `.agents/skills/`, `.agents/`, `.claude/skills/` or `.claude/` if they are then empty). Removing an alias
   disables discovery by that agent only; it does not remove the workflow instructions.
2. **Replace the rules.** Replace or update [AGENTS.md](../AGENTS.md) and [CLAUDE.md](../CLAUDE.md) with your own
   rules, and remove or rewrite `docs/agents/`. Keep the neutral repository orientation (commands and architecture
   constraints) if it is useful to you.
3. **Optionally delete the bundle.** Delete `distributions/skills/restruct-sdlc/` (and `distributions/skills/` and
   `distributions/` if they are then empty), then remove the bundle's row from
   [THIRD_PARTY_NOTICES.md](../THIRD_PARTY_NOTICES.md) and its mention in [NOTICE](../NOTICE). Keep the MIT notices
   for as long as any of its files remain.
4. **Commit the change as its own commit.** The harness's skill-root check (`B-SKILL-UNCHANGED`, `provenance` suite)
   compares the skill directories and discovery aliases with the `--base` commit, so run the developer checks with a
   base that already contains the removal, or expect that one change to be reported as a skill-root change.

The validation harness, the application and its tests do not use Restruct. Opting out changes no application code,
migration, configuration, test or runtime behaviour. Delete `.codex/config.toml` too if you do not want its Codex
model defaults.
