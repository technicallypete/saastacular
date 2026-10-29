> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G055 post-merge full-run disposition — 2026-10-20

## Independent failure inventory

Reviewed the completed run's `results.tsv`, `depscan-results.tsv` and `depscan.log` from private evidence
`~/.local/state/platform-evidence/g055-postmerge-5b813e32/live-copy/report/`.
Run baseline: merged staging `5b813e32`. Results: 118 PASS, 7 INFO, 2 FAIL.

| Failed control | Diagnosis | Disposition |
|---|---|---|
| B-DEPSCAN | Current dependency policy blocks PyPI `certifi@2026.7.22`, Critical `GHSA-6q48-6c7w-9gw2`, patched in `2026.10.19`. Advisory downloads, scanner build and synthetic fixtures pass. | Real dependency finding; retain failure. Request a narrow repair proposal before outside-scope implementation. |
| B-REQUIRED-CHECKS | Derived solely from failed B-DEPSCAN. | Retain aggregate failure; no separate production fix. |

All other listed controls passed. The existing npm braces finding is already excepted by the scan policy;
this disposition creates no new exception and changes no severity threshold.

## Advisory verification

The [GitHub-reviewed certifi advisory](https://github.com/advisories/GHSA-6q48-6c7w-9gw2), published 2026-10-20,
confirms the affected range includes the locked version and names `2026.10.19` as patched. It concerns compromised
CA intermediates and trust-store removal. This is unrelated to the stdlib-only Restruct export; the already
published bundle/tag remain attributable to their separately reviewed package and secret-scan evidence.

## Supervisor decision

**G055 final ACCEPT remains held.** Preserve the failed full run; do not blindly rerun it, exempt the advisory,
change test expectations, or reinterpret publication as qualification.

Coordinator to propose a bounded corrective dependency change: exact lockfile paths, direct/mirrored template
parity if present, certifi-only patch update and selected validation commands. No unrelated upgrades or new
infrastructure. This is a request for a repair proposal, not an implementation release or a new Goal.
Codex will review the exact scope and qualification strategy before execution. Reuse unaffected passed controls
when justified; no replacement full-suite run is authorized by this disposition.

## Owner disposition superseding the hold — 2026-10-20

The owner subsequently directed: “we can qualify and close out O five five.” Codex records this as a task-specific
qualification exception for the independent certifi advisory and **FINAL ACCEPTS G055**. The original run remains
FAILED with both controls and counts unchanged; no advisory exception is added and no test result is changed. The
certifi-only dependency maintenance is separate Beginify staging work and does not reopen G055. See [final review](G055-final-codex-review-20261020.md).
