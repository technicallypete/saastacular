# Development workflow

SaaStacular uses **Restruct SDLC** as its default development dependency. It is an
unaltered GitHub submodule at `.agents/skills/restruct-sdlc/`, pinned to an exact
commit by Git and registered in [`.gitmodules`](../.gitmodules). Restruct retains
its MIT licenses; SaaStacular's Apache-2.0 license does not replace them.

The full workflow stays in [AGENTS.md](../AGENTS.md), [CLAUDE.md](../CLAUDE.md)
and [docs/agents/](agents/): roles, Goal/Spec/Plan/Result records, planning,
implementation, review, budgets, parallel scheduling, validation, integration,
evidence, resource accountability and cleanup. Installation starts no agent or
watcher and creates no management records automatically.

## Initialize for agent work

Clone with `git clone --recurse-submodules` or initialize an existing checkout:

```sh
git submodule update --init -- .agents/skills/restruct-sdlc
```

The entrypoint at [`.agents/skills/restruct-sdlc/SKILL.md`](../.agents/skills/restruct-sdlc/SKILL.md)
routes to three complete child skills:

| Child | Responsibility |
|---|---|
| `sdlc-management` | Adopt the workflow, bootstrap records, and apply delivery policies and templates. |
| `repository-supervisor` | Recover supervision, review Plans/Results, and manage dependencies and acceptance. |
| `claude-session-handoff` | Verify local Claude sessions, deliver authorized handoffs, and observe milestones. |

Claude discovers the same submodule through the relative symlink
`.claude/skills/restruct-sdlc` → `../../.agents/skills/restruct-sdlc`.
Linux and WSL support this layout; native Windows requires symlink support.
Use the pinned project version rather than an unrelated personal copy.

Start adoption with the submodule's `sdlc-management/references/adoption.md` and
`sdlc-management/templates/parameters.example.json`. Record owner/agent roles,
canonical documents, branches, validation commands and budgets. Reconcile its
proposals with the full rules already in AGENTS.md and CLAUDE.md; do not overwrite
them automatically. Use this repository's existing validation commands.

[`.codex/config.toml`](../.codex/config.toml) supplies optional project model
defaults; explicit session settings override them.

## Application checks and dependency updates

Application configuration, builds, tests, migrations and runtime work without
initializing Restruct. Parent CI explicitly skips automatic submodule checkout.
The parent scanner scans `.gitmodules` as ordinary source metadata and excludes
this independent submodule without inspecting it. Application snapshots and
Docker contexts exclude the development dependency. CI and local application
checks do not validate Restruct's pin, contents or Claude symlink and do not run
Restruct tests. They work when Restruct is absent or uninitialized.

Submodule initialization uses the recorded commit, never the latest branch tip.
To update Restruct, fetch its upstream commits, check out a reviewed commit in
`.agents/skills/restruct-sdlc`, and commit the changed gitlink in SaaStacular.
Reconcile live workflow guidance when upstream policies change. Such updates
retain the normal review rules; the application does not acquire a runtime
dependency on Restruct.

The original export manifests describe the initial publication and remain in that
commit's history. Current maintained repository content is identified by its Git
tree and the explicit Restruct gitlink, rather than stale export file lists.

## Use another workflow

1. Preserve any local Restruct edits, then use `git submodule deinit -- .agents/skills/restruct-sdlc`
   and `git rm .agents/skills/restruct-sdlc .claude/skills/restruct-sdlc`.
   Remove `.gitmodules` if no entries remain. Removing discovery alone does not
   remove the written workflow rules.
2. Replace or update AGENTS.md, CLAUDE.md and `docs/agents/` with your own rules.
   Keep useful repository commands and architecture constraints.
3. Reconcile the Restruct notices in [NOTICE](../NOTICE) and
   [THIRD_PARTY_NOTICES.md](../THIRD_PARTY_NOTICES.md). Retain MIT notices for any
   Restruct content you keep. Commit the intentional removal as its own change.

No application test or CI change is needed to remove Restruct. Delete
`.codex/config.toml` too if its model defaults are unwanted.
