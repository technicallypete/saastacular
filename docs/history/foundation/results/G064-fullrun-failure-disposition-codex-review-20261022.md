> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G064 merged-SHA full-run failure disposition — 2026-10-22

## Decision

Authorize one bounded G064 correction for `B-SECRET-CANARY`: after
`mcp_e2e` has extracted and registered the observed bearer values, remove the
raw `mcp-e2e.log.events` intermediate before any evidence retention or scan can
preserve it. The converted per-test log remains the diagnostic artifact and
must remain redacted. Do not place the raw bearer values in this record, a
Result, a console message, or a test assertion output.

This is a confirmed G064 evidence-privacy defect. In the reviewed staging code,
`mcp_e2e` registers values from the converted `$log` and scrubs the converted
`$WORK/go-tests/mcp-e2e.log`; it does not remove the source `$log.events` stream
created by `go_it`. The latter contains the raw Go JSON test stream and can
therefore retain the bearer output. The fix is limited to the G064 harness path;
it changes neither product authentication nor test assertions.

Require a focused regression using a planted bearer value to prove: the value
is registered before the raw intermediate is removed; no raw `.events` file
survives the completed MCP invocation or enters retained/exported evidence;
the converted log is redacted; and the secret canary passes. Preserve a
negative control showing the old path leaves the planted value in the raw
stream. Keep the test output redacted.

## Separate failure retained

The same reported full run also failed `B-CI-CONTROLS` at V18m under load, with
`B-REQUIRED-CHECKS` derived. This ruling does not waive or attribute that
failure to the MCP defect. Preserve the exact run/SHA/job and V18m evidence in
the G064 Result; diagnose V18m separately and rerun only the affected control
as permitted by the accepted validation policy. Do not represent the full run
as passed. G064 final acceptance remains pending until the required validation
disposition is complete.

## Scope and gates

Claude may implement this single correction, add its focused regression, and
submit the source for Codex review. It is not source acceptance, MR merge
approval, waiver of the failed full run, or final Goal acceptance. Preserve the
failed run and all other passing evidence. No broader redaction refactor or
unrelated changes are authorized.

## Evidence reviewed and limits

- `infra/validation/run.sh`: `go_it` captures the raw stream as `$log.events`;
  `mcp_e2e` registers observed values and removes them from the converted log,
  but does not remove `$log.events`.
- `infra/validation/lib.sh`: the exported per-test log is the converted log and
  is redacted by `go_tests_export`; raw `.events` files are intended to remain
  only in the private work directory.
- Coordinator notice dated 2026-10-22 14:07 UTC reports
  `B-SECRET-CANARY`, V18m under load, and derived `B-REQUIRED-CHECKS`. The exact
  run SHA and retained artifacts were not present in the staging Result when
  this disposition was written; Claude must add them without exposing secrets.
