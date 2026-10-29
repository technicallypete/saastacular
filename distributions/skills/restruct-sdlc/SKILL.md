---
name: restruct-sdlc
description: Restruct SDLC bundle entrypoint. Routes to exactly one of three preserved child skills — sdlc-management (adopt or apply the Goal, Spec, Plan and Result management workflow and its templates), repository-supervisor (resume supervision from repository records) or claude-session-handoff (check local Claude Code sessions, deliver authorized handoffs, run the milestone observer). Use when any of those tasks is needed; read only the child the task needs.
license: MIT
---

# Restruct SDLC

This bundle packages three complete, independently maintained skills under one directory. This entrypoint only
routes; it adds no policy, role, helper or workflow of its own. **Loading it assigns no role and starts nothing.**

## Routing

Read the **one** child `SKILL.md` the current task needs, through its contained relative link, and follow it. Do
not load all three bodies for routine status or messaging.

| Task | Read |
|---|---|
| Adopt the management workflow, instantiate templates, or apply Goal/Spec/Plan/Result, budget, validation, integration and closeout rules | [sdlc-management/SKILL.md](sdlc-management/SKILL.md) |
| Resume or continue supervision of a Goal from the repository's management records | [repository-supervisor/SKILL.md](repository-supervisor/SKILL.md) |
| Check a local Claude Code session, deliver an authorized handoff, or run the milestone observer | [claude-session-handoff/SKILL.md](claude-session-handoff/SKILL.md) |

Children keep their own names, frontmatter, notices, references, scripts and templates; a child may link to a
sibling child for a specific rule. Resolve repository paths named in a child (`AGENTS.md`, `management/`, helper
commands) against the **target checkout**, never against the skill storage directory.

## Layout

- **Adopting repository (supported default):** copy the whole `restruct-sdlc/` directory, as real files, to
  `.agents/skills/restruct-sdlc/` and commit it. Codex discovers the parent there. An optional Claude discovery
  alias is the relative symlink `.claude/skills/restruct-sdlc -> ../../.agents/skills/restruct-sdlc`. Children are
  not installed or discovered separately. See
  [installing the bundle](sdlc-management/references/adoption.md#installing-the-skills).
- **Source-maintainer repository:** the canonical tree lives at `distributions/skills/restruct-sdlc/`, with the two relative
  discovery aliases `.agents/skills/restruct-sdlc` and `.claude/skills/restruct-sdlc` pointing to
  `../../distributions/skills/restruct-sdlc`.

Installing or discovering the bundle creates no management tree, fills no project decision, starts no agent and
activates no observer. Adoption is a separate, explicit, owner-authorized procedure described by sdlc-management.

## Verification

Run from the repository root with `python3 -B` (no bytecode in skill storage):

```sh
python3 -B <bundle>/sdlc-management/scripts/check_sdlc.py package <bundle>
python3 -B <bundle>/sdlc-management/scripts/check_sdlc.py package <bundle>/sdlc-management --skills-root <bundle>
python3 -B -m unittest discover -s <bundle>/sdlc-management/scripts -p 'test_check_sdlc.py'
python3 -B -m unittest discover -s <bundle>/claude-session-handoff/scripts -p 'test_*.py'
```

`<bundle>` is `distributions/skills/restruct-sdlc` in the source-maintainer layout and `.agents/skills/restruct-sdlc` in an
adopting repository. A distribution root (`README.md`, `LICENSE`, `MANIFEST.tsv`, `restruct-sdlc/`) is checked
with `check_sdlc.py package <distribution-root>`.
