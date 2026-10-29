> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G020 final Goal acceptance — 2026-10-24

Outcome: **FINAL ACCEPT**.

## Reviewed delivery

- Implementation MR !95 (not included) merged as
  `504fe8fc7138111ff566c1a07902bfe550d5e30c`. The reviewed source was
  `a7b93b110cf23c000bdc9f968e59193a1346e7f3`; its tree matches the merge tree
  `c0f3882cf2ce40b0597399a13e7b950cab3b36e9`. Hosted pipeline `2983877204`
  succeeded. Duo's MR !95 comment disclosed that its diff view was truncated;
  Codex's full source review is the review of record.
- The post-merge affected qualification on `504fe8fc` completed with exit 0:
  all 529 `ci-controls` passed; all 15 required IDs and the final row passed;
  there were no FAIL or UNMET result rows, build-ledger exceptions, or leftover
  inventory. Docker stayed at the recorded baseline (15 containers, 11 volumes,
  6 networks, 52 image rows). The retained evidence is
  `<home>/.local/state/platform-evidence/g020-postmerge/`.
- Source coverage from the approved implementation runs and the exact merged
  tree is recorded in the [G020 Result](G020-generic-labels-visibility-result.md),
  including the migration, catalog, authority/privacy, RLS, audit, HTTP/MCP,
  web and browser evidence. Earlier failed runs remain recorded as failures;
  later passes do not rewrite that history. No blanket full local application
  suite was required by the accepted selection.
- Result closeout MR !96 merged before Codex reviewed its text. This process-order
  deviation, its authorship correction, and the supplemental archived-ref
  correction are recorded in the Result and were not erased. MR !97 records the
  correction. MR !98 added the missing AC8 observation; it merged as
  `1f59302361e6645fe9473c71b95c34c14323a61f` from source
  `83f13547f90f946e1db4c901e0368bdf211b4812`, with identical tree
  `c30201cd9ac79ea909e05840455e688d5d02bad0`. Duo approved !98 and the
  Result-only MR had no pipeline.

## Spec acceptance

| Criterion | Disposition |
|---|---|
| AC1 — Model and lifecycle | Accepted from the reviewed version-20 migration, lifecycle, constraint, migration and catalog evidence in the Result. |
| AC2 — Current authority | Accepted from the Organization/Group role positives and denials, including direct HTTP/MCP enforcement, in the Result. |
| AC3 — Privacy | Accepted from default-private Organization-label checks and current-authority/visibility evidence; no G021 or member-visibility expansion is claimed. |
| AC4 — Target boundary | Accepted from same-tenant positives and foreign, cross-Group, unknown and unsupported-target negatives; the synthetic target remains test-only. |
| AC5 — Storage isolation | Accepted from forced-RLS, direct-SQL and cross-scope evidence in the Result. |
| AC6 — Atomicity and audit | Accepted from rollback-on-audit-failure, event-shape and catalog-parity evidence in the Result. |
| AC7 — Surfaces | Accepted from the reviewed web, HTTP/MCP parity, direct-call denial and focused UI/browser evidence. |
| AC8 — Delivery and workflow | Accepted from the required MR/qualification/cleanup disposition and the new factual G055 usage observation in Result §“G055 consolidated tooling — bounded usage observation.” |

The AC8 observation describes what the coordinator reported doing: repository
discovery through `CLAUDE.md` and `AGENTS.md`, the roles/handoffs used, and
cleanup guidance followed. It explicitly states that the parent/child skill
files were not separately loaded and makes no G055 requalification or G021
barrier claim. The bounded independent review found no material AC8 defect;
it did not independently audit the underlying Claude transcript.

## Qualification and cleanup limits

The retained post-merge launcher required the exact merged SHA and a clean
checkout before starting. The run logs do not contain a captured
`git rev-parse HEAD` line, so the checkout SHA is not independently visible in
the logs; the launcher guard and matching 75-path `B-SCOPE` count corroborate
the binding. This evidence limitation is disclosed in the Result and is not
represented as a stronger capture than exists.

Closeout verification found no G020 local/remote refs, worktrees, temporary
paths or owned Docker resources. The new AC8 writer branch
`worktree-agent-<id>` was unattached and pointed exactly at
`origin/main`; it was compare-and-deleted. The 112 unrelated owner-held
`worktree-agent-*` branches remain. The live Claude coordinator worktree remains
protected. The archived G020 bundles and evidence tarball were checksum-checked;
both bundles verify. Historical merge-order and authorization-attribution
deviations remain visible in the Result.

No tests were launched for the docs-only closeout. The repository's next
scheduled action is to cut RC11 from the then-current staging tip, followed by
the combined staging preview. This acceptance does not authorize main promotion
or waive the G021 foundation barrier.
