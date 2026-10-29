> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G053 — Implementation review

Outcome: **CORRECTION LOOP** — three bounded adoption/documentation corrections.
Reviewed 2026-10-12 UTC. No implementation ACCEPT or MR integration approval.

## Reviewed candidate and evidence

Candidate `88e07f55bb29f17000e19c3ec524c17e621c1c77`, base
`3a3f4b87528028d71b1f067e1657ff14f8b49cfc`; remote delivery branch
`g053-restruct-sdlc-bundle` matches. Result SHA-256
`95a6e762a6aa107e68b2680772fd5d1c861df3649c63fa3779477694b1dff1f4`,
imported byte-identically. A Codex read-only subagent reviewed actual skill,
templates, checker and local bundle; supervisor independently reviewed scope,
protected-source bytes, publication tip, manifest and full-run evidence.

Only the new skill's 29 files and Result change. The bundle's 40 manifest rows
match its exact files/hashes; new skill equals candidate bytes. Helper executable
and protected source samples remain unchanged. The full local run on
`5034d06d8098ab4aea9bd00ba46e49b6456ff593` ends status 0, all phases 0,
88 PASS/5 INFO, cleanup 0, owned_remaining 0 and registry removed. Scope/protection
checks pass. The candidate's subsequent delta is Result-only. Historical failed
inner runs remain recorded. Hosted gates are pending; no MR exists.

## C1 — Existing-repository adoption cannot scan unrelated app content

`skills/sdlc-management/scripts/check_sdlc.py` A3 (lines 539–545) scans all
non-skill source for template delimiters; A6 (line 571) applies package privacy
patterns to the entire adopting repository. A valid app template containing
`{{ customer.name }}` or README with a public GitHub source URL fails adoption.
Real UUIDs/public Git-host links in ordinary project records are also legitimate.
This conflicts with R4/AC4's existing-repository adoption.

Separate package/export privacy from target-repository adoption. Keep P6 strict
on public package content. Constrain unresolved SDLC placeholders to declared
adopted artifacts; adoption privacy checks must target genuine source-origin
leaks/private state, not reject arbitrary existing app code or legitimate target
project URLs/identifiers. Do not simply exempt all management content or weaken
checks on installed/exported skill files. Explain the scope in checker docs.
This supersedes the overly broad adopting-repository scan described in Plan r2;
it changes no public-package privacy guarantee.

Minimum checks: unrelated existing app delimiters, a public source URL and a
legitimate UUID survive adoption unchanged; an unresolved adopted parameter still
fails; a planted prohibited export/source-origin value still fails in its proper
scope. Extend the covering fixtures, not the test framework.

## C2 — Installation must preserve existing same-named skills

Bundle README's recursive copy command (lines 27–31), adoption reference step 2
and SKILL.md step 2 install into existing destinations without a collision check.
`cp -R` can silently overwrite an edited installed helper/skill, violating R4.

Add a documented pre-copy check: create only absent safe destinations; identical
installed content may remain unchanged; differing content must be preserved for
a recorded owner merge/replacement decision. Reject unsafe symlink destinations.
Keep README, entrypoint and adoption reference consistent. A small procedure is
sufficient; no installer framework or helper change is requested.

Minimum check: an existing edited skill remains byte-identical and the procedure
records its collision/decision rather than overwriting it. Existing fresh-install
and relocation walkthroughs still pass.

## C3 — Direct execution must not require a separate Plan approval

`templates/root/CLAUDE.md.tmpl` step 1 exempts an explicitly authorized direct
tier, then step 2 unconditionally says to stop for Plan ACCEPT. R4 and the
adoption risk-tier section authorize direct execution without a separate Plan.

Make step 2 conditional on the Goal requiring a Plan. Retain explicit authorized
Goal/Spec scope for direct execution and stronger architecture/security/data
review; agents cannot self-select a lighter tier.

Minimum check: a rendered direct-tier walkthrough proceeds under authorized
Goal/Spec, while an ordinary Goal still stops for Plan ACCEPT.

## Correction authority and remaining gates

Authorize **attempt 2 of B3** for these corrections. Maximum remains 3, checkpoint
after attempt 2; preserve attempt 1 and all failures. Fix only the new skill and
exported copies/README, re-export through the accepted ownership procedure and
update the Result. Do not rewrite accepted Plan/history or existing foundation/
application/companion/CI sources.
Continue with the existing verified G053 owner/retained checkout; its recorded
naming exception does not authorize another generated worker or duplicate owner.

Run covering checker tests, affected package/privacy/install/adoption exercises
and required final coverage under the unchanged path/Plan policy. Reuse unaffected
helper/application evidence only with exact provenance; do not repeat debugging
against a broad suite or add qualification campaigns. Fetch/reconcile before
batches; preserve cleanup, failed/interrupted evidence and actual CI placement.

The coordinator's safekeeping push is independently verified at `88e07f5`.
Append truthful publication status to the Result, whose worker-authored text
predates that push; preserve the original explanation. Push completed correction
checkpoints under standing authority, then return for automatic review.
MR readiness, owner merge authority, hosted disposition and final ACCEPT remain.
