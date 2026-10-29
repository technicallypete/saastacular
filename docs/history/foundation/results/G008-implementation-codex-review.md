> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G008 implementation review — 2026-10-14 UTC

**CORRECTION LOOP** — one bounded input-classification correction.
Reviewed candidate `33f4517cd6155ccc5942aedd352cea1a3f02e53d`, base `e80077d`.
Exact Result SHA-256 `98a88593f62a3bbe53985765859beb64f140b42d3b6504ee85b96d9dc974fccb`.
Read-only delegated source review found no blocking RLS/grant/context/pool defect.
Verified hosted gate job `17160250730`, pipeline `2945315031`, passed on `f0ed177`;
latest candidate differs only in Result. Latest fast and Duo review remain separate.

## Required correction

Duo's low finding is confirmed in `core/internal/identity/store.go` candidate
lines 666–709: DisableUser/EnableUser reject only the zero UUID. Noncanonical
or malformed context IDs reach the strict SQL context parser and map to
ErrUnavailable. Validate the required canonical nonzero UUID form before opening
a transaction; return ErrInvalidRequest for invalid inputs. Preserve valid UUID
behavior, transaction-local context, revocation/ordering and unavailable classification
for genuine storage failures. No API/security-model redesign or new dependency.

Minimum verification: both operations cover empty/malformed/uppercase/zero input
and demonstrate rejection before database work; retain canonical valid-ID operation
coverage. Run focused identity unit and covering integration/ordering tests. Fetch/
reconcile before the batch under the accepted procedure; keep prior failures and
passing full-gate evidence. Reuse unaffected controls with exact content provenance
and reconcile required final/hosted coverage; no identical CI-control rerun solely
for this guard. Append correction evidence to Result and disposition the Duo thread.
Stay within the existing shared B7 allowance, counting the correction honestly.

## Nonblocking residuals and gates

The b34 setup-sign-in ordering failure remains unexplained and disclosed; subsequent
20 isolated passes, four ordering passes and hosted PASS support treating it as a
nonblocking residual without a new full diagnostic campaign. Preserve PG-002's
owner-accepted measured risk, source/export history and item-specific cleanup holds.

No G008 final ACCEPT or merge granted. Owner-authorized G009 provisional planning
starts only after MR checks pass and Duo findings are resolved/dispositioned.
G009 implementation still requires final accepted/integrated G008 and its own
approved Plan/binding. G021 remains closed.

## Correction pass 1 re-review — 2026-10-14 UTC

**ACCEPT for correction/source review; final Goal ACCEPT pending staging.**
Reviewed `d0e8d400ecde7c993bec086110ab39f394dea03b` against `33f4517`.
The guard matches the strict nonzero lowercase UUID context parser, rejects
invalid IDs before database work, and preserves valid-ID transaction/locking/
revocation and genuine storage-error behavior. The test covers thirteen invalid
forms for both methods with zero DB calls, plus the canonical positive path.
Read-only delegated correction review found no source defect. Result §14's
focused b42/b43 and b44 unit/migrations/ops evidence is proportionate; no tests
were rerun by Codex. Clarification: its byte-identity reuse claim covers unchanged
application/schema/harness inputs outside identity, not changed management files.

MR fast job `17161473752` in pipeline `2945514655` passed on `d0e8d40`.
Duo thread `a215cc094fc0a618015b7b24b73b2006941aa5c1` is resolved; its
subsequent review confirms the fix and reports no further finding, with its
truncated-output limitation retained. Earlier full gate evidence stays attributable
to `f0ed177`, not represented as a fresh full run on the correction.

GitLab reports MR !23 merged at 06:33:44Z as `81b186c`, before this correction
re-review; this record does not retroactively claim pre-merge review completion.
The temporary remote source branch is gone. Integrated application/infra/
foundation-doc bytes match `d0e8d40`; local staging was fast-forwarded.
Staging pipeline `2945531965` is running and does not yet satisfy final acceptance.
Preserve existing bounded closeout obligations; do not remove live/proof-uncertain
resources. No remaining source correction requested.

Owner's conditions for provisional G009 planning (passing MR checks and disposed
Duo findings) are now met. Planning may start on integrated `81b186c`; G009
implementation remains held for G008 final ACCEPT and its own approved Plan/binding.

## Final implementation acceptance — 2026-10-14 UTC

**ACCEPT — G008 is completed and integrated.**
Accepted correction candidate `d0e8d400ecde7c993bec086110ab39f394dea03b`,
integrated by MR !23 as `81b186c3872ea49ea571761b46f612007fb909ce`.
Current submitted Result SHA-256 `58d193fbf8b814dcf87d3f39df24414c19cf39761d5ad3dfef6f95d987f758f2`.
Prior source/correction reviews and fourteen-criterion evidence stand; no
application/schema/harness delta exists between the tested merge and current
staging at this decision. Later management/skill updates do not claim new tests.

### Verified hosted disposition

Pipeline 2945531965 (not included)
is **success** on `81b186c`. Fast `17161589052` and gate `17161589053`
are both successful required jobs (neither allow-failure). Gate duration
2641.60 s, about 44 minutes; total pipeline duration 3070 s.
Gate trace records harness exit 0, B-RLS-CATALOG/MIGRATION/IT/ORDERING PASS,
B-CI-CONTROLS PASS and B-REQUIRED-CHECKS PASS; owned cleanup verified
(4 resources removed), final context=0 and guard=ok.

The existing saved gate results contain 64 PASS/6 INFO, no FAIL/UNMET;
63 required gate IDs are reported once each and 441/441 CI controls pass.
Gate is the full-profile remainder, not a standalone fast-job substitute;
both successful jobs form the required pipeline coverage. Saved evidence:
- `results.tsv` SHA-256 `020fef8514b56e300858e3e98ceecd1441d30ffe25e2200fed7dfddf16a39cad`
- `summary.json` SHA-256 `703161232d7602f14b7875a22eb911f2c61318dd4b489297b45f3df3621394e8`
- `selection.tsv` SHA-256 `0e4284cbe91bd3c4c5aba7e0dd91e01564371eb6d1ec768b59acb49fa57c1968`

Duo's UUID finding is corrected/resolved and the correction source review is
ACCEPT. Preserve previous failed/interrupted runs, unexplained b34 setup-sign-in
residual and owner-accepted PG-002 risk; none becomes a pass or a production
performance guarantee. B7 usage remains the reported shared attempt 1, with
correction pass 1 and inner batches retained; no checkpoint-4 requirement arose.

### Closeout and downstream gates

Remote temporary source deletion is verified. Authorize existing bounded local
closeout for terminal G008 planning/implementation worktrees, local task branches,
markers and disposable owned focus/test resources after verifying no live/unique
work or needed evidence. Retain compact acceptance/evidence and required inputs
before purging bulk output. Preserve uncertain items, native permission holds,
G009/G054 workers, coordinator checkout and unrelated databases. Cleanup is
separate from acceptance; report its compact receipt/held items without a new
framework or validation campaign.

PROGRESS compaction assessed: G009 planning is active against recent management
inputs and EPIC-003 remains open through G010. Defer moving its live chronology
to the next safe Goal-closeout window; the G010 epic closeout is a mandatory
compaction checkpoint. Retain this concrete deferral rather than rewriting
active inputs.

G009's G008 dependency is satisfied. Continue its corrected Plan review and
fresh execution binding before implementation under the owner's unattended
loop authority. G010/G011 and G021 are not accepted or released by G008.
