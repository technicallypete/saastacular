> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G009 implementation review

## Attempt 1 pre-integration review — 2026-10-14 UTC

**CORRECTION LOOP — one bounded P2 finding.** No MR or implementation
acceptance yet. Authorize correction pass 1 as **attempt 2 of B5**; checkpoint
after attempt 3 remains. Scope and fourteen ACs are unchanged.

Reviewed clean candidate `787ca4e2ba9a3036a3acdda469fdff6aba5f630b`, branch
`g009-tenant-object-storage`, on staging
`f629847238e0516dc0292c624eaf217bbb16f196`. Result SHA-256:
`b0de505bfcf8869ce49114b1d295f2bed18d1ac967a359940c83fcb90e2aef64`.
Independent oracle r2 remains
`e1d26df13f845de4787986e08e32195bc92c91d3c8b0d13f28ed5fa17f9ab96f`.
Codex inspected the actual diff and local reports, with two bounded read-only
source reviewers (storage lifecycle and RLS/authorization). No tests or resource
mutations were performed by Codex reviewers.

### Required correction C1 — cancellation races with digest snapshot

At this candidate, `core/internal/tenancy/access/objects.go:536` reads
`xr.sum()` unconditionally after closing/reaping the owned source.
`objects_io.go:46` deregisters an underlying Body.Read before the surrounding
exactReader.Read performs its hash.Write (line 210 or 265). An asynchronous
transport can return on cancellation while a conforming source releases a
data-bearing read. waitReleased proves that Body.Read has returned, not that
the outer reader's hash update is complete. Consequently hash.Sum and hash.Write
can run concurrently. This is a static source finding, not a reproduced failure.
It violates R7/AC11 cancellation safety.

Prevent concurrent digest access. The smallest sufficient correction may read
the digest only after observing verdictExact: the final hash write precedes
that verdict's publication under the reader mutex, and other paths do not need
a digest. Consistent synchronization of hash mutation/snapshot is another
permitted implementation choice. Do not add an unbounded transport join, weaken
deadlines/compensation, or change accepted residual X6 to resolve this race.

Minimum validation: a deterministic asynchronous-store/cancellation regression
with a conforming source returning data as it releases; exercise the snapshot
window, verify no premature publication and preserve close-once/bounded-return
behavior. Run the targeted regression under the race detector and the covering
upload-reader/cancellation/fault checks, including adjacent exact-success digest
cases. Record the original finding and new evidence; escalate an unavailable
race-tool prerequisite rather than building an unrelated validation framework.

Fetch/reconcile before batches. Rerun affected coverage and reuse unrelated
passing evidence with exact delta provenance; no repeat of unchanged CI controls,
browser or other suites solely for this correction. Final hosted gates remain.
Append the corrected Result and safekeep the branch before returning for review.

### Reviewed areas without blockers

- Migration 00007, UUID/generated key/same-Organization FK, exact grants,
  FORCE RLS, creator-only transitions, bounded current-fact/session decisions
  and independent oracle/detectors conform to the approved Plan and rulings.
- Storage adapter, staged finalization/delete, ambiguous-COMMIT reconciliation,
  pre-upload key check, signing/HTTPS/privacy and documented X1–X7 residuals
  otherwise conform within this review. No new architecture decision is required.
- Reported deviations 1–4 and 6–7 are accepted as bounded reconciliation:
  R2 notes in OBJECT-STORAGE.md, constant location, version-6 subset runners
  retaining inherited tests, boundary parsing, APPLICABLE additions and final
  report-only delta. The fixed decision-context bug remains recorded history.

### Validation and remaining gates

Verified b18 (`9487675`, target f629847): 95 PASS, two substantive failed IDs
(B-PROTECTED and B-TOGGLE-OFF), plus the failing aggregate B-REQUIRED-CHECKS.
It remains a failed run. b19 (`6466f6f`, same target) passed 24 checks, including
both corrected IDs and aggregate. The intervening delta is only APPLICABLE and
Result bytes; b19 covers the affected suites. Across the reconciled batches,
98 distinct B-* IDs have passing evidence (full plus unit). b18 includes all five
B-OBJ-* and 441 CI controls. `6466f6f` to submission changes only the Result.
This supports reuse, not a claim that b18 itself passed.

b18's informational build-ledger UNMET is the predicted web-check build missing
when the disabled variant failed. The corrected off batch's ledger is MET with
12/12 rows and no unknowns. Preserve both observations; no new ledger project.

After C1 source/evidence re-review, use the authorized MR workflow: reconcile
latest staging, submit reviewed content, dispose Duo findings, then eligible
auto-merge and applicable staging disposition. Cleanup/closeout and final AC14
remain to be verified. G009 is not completed; G010 and G021 are not released.

## Correction pass 1 / attempt 2 pre-integration review — 2026-10-14 UTC

**ACCEPT — implementation source and reconciled local evidence for MR preparation.**
This is pre-integration acceptance, not final Goal completion. C1 is resolved.
Budget remains 2 used of B5; checkpoint after attempt 3.

Reviewed clean candidate `5605a6c41a9b9ea5edb298fdf473755ace63ca32` on staging
`6ed5363f9b9519a6f42357dfec776f418f9e0802`. Result SHA-256:
`97826521e15a8154121f79cff2ae6f2560cfdaf2031263cbb78959f071ab5fe6`.
The exact Result is imported into staging with this record. Oracle r2 remains
byte-identical at the SHA recorded above. The original storage reviewer
re-reviewed the bounded correction and recommends ACCEPT; no new blockers.

### Verified correction and evidence

- Only three core files change from the rebased prior candidate `80a918c`:
  objects.go, objects_io.go and objects_test.go. The transport goroutine alone
  touches the hash. After the final write it computes the digest, then publishes
  digest and verdictExact together under the mutex. upload reads that result;
  pending/error paths return no digest. No unbounded join or deadline,
  compensation, authorization, oracle or X6 change.
- Regression covers the held data-bearing-read window, cancellation, close-once,
  bounded return and successful complete digests. b24 targeted race-detector
  run (`-count=3`) exits 0 with zero race warnings. b25's isolated reverted-code
  negative control exits 1 with six race reports at the original Sum/Write
  boundary; this is an expected failed control, not a candidate failure.
- b26 passes the storage integration/ordering/fault/footprint batch. b27 on clean
  checkpoint `a6fae5e2c0c79ec145b7d120037b6af4a5bde4a6`, target 6ed5363, exits 0:
  41 PASS, all 40 required IDs pass, no FAIL/UNMET. It covers lint, unit and
  migrations, including all five B-OBJ-* checks. Submission differs from that
  tested checkpoint only in its Result report.
- Reuse unchanged passing b18/b19 evidence under the previous exact-delta review.
  No additional full local, browser or CI-control rerun is required for C1.
  The same-version pinned Debian Go image supplies the race detector's C toolchain
  as a bounded test tool; no repository dependency or runtime image change.

### Integration authorization and remaining gates

Claude may reconcile the branch onto latest origin/staging, push and open the
feature MR to staging under standing owner authority. Preserve the reviewed
application bytes; assess any relevant delta before reuse. The management-only
review/Result import does not independently require application revalidation.

Keep source-branch deletion enabled. Record the reviewed source SHA and obtain
required fast/MR evidence and Duo disposition. Do not arm auto-merge while
review/Duo findings or other pre-merge holds remain; report findings for Codex.
Once eligible, use the standing authorized auto-merge route without another
routine owner-permission question. Required staging coverage, failed-check status
rules, actual integrated-content reconciliation and cleanup remain binding.

G009 remains pending final hosted disposition/closeout and Codex final ACCEPT.
PG-003 stays open with its recorded observation. G010 and G021 remain unreleased.

## MR !24 Duo disposition and conditional integration — 2026-10-14 UTC

**ACCEPT — no implementation correction required for the one Low finding.**
MR !24 (not included)
source `4c7569e95d9801a9c2e1360e0611e266b6c9cce1` differs from reviewed
`5605a6c` only in PROGRESS and this Codex review (imported staging history).
Application, tests and harness bytes remain reviewed/tested content.
Pipeline `2946524336` succeeded: required fast job `17168951532` succeeded
in 427 seconds; the optional allow-failure manual gate was not run. This is
the intended MR profile, not comprehensive hosted qualification.

Duo discussion `b9bfc4b13b2dd0ca7f74b489b19a4015e5b5d643`, note `4011903885`,
observes that setting only SIGNED_URL_MAX_TTL below 60 while leaving DEFAULT_TTL
unset causes configuration refusal, and suggests clamping the implicit default.
The observation is accurate; it is not a violation of the reviewed contract.
Plan r1 §4.7 fixes the default at 60 and requires default <= maximum; the
documented settings table states the same relationship. An unset value uses its
documented default, and the resulting pair must be valid. For maximum 30,
operators must explicitly configure DEFAULT_TTL <=30 as well. The error names
both dependent settings and fails closed. Clamping would be a changed defaulting
policy, not a necessary correctness/security fix. Retain the current behavior;
the configuration ergonomics observation is nonblocking. No test or production
change is requested solely to accept a different preference for defaulting.

Duo disclosed truncated coverage; its summary is not a complete security review.
The independent Codex source reviews above cover the migration, model, storage,
authorization and harness portions in addition to that automated review.

Claude may reply with this rationale and resolve the finding, confirm current
source/target and no other holds, then arm SHA-bound auto-merge to staging with
source-branch deletion under standing owner authority. No extra manual MR full
run or repeated unchanged local tests. A relevant new source/target change must
be reconciled. Qualifying staging disposition, integrated-content verification
and cleanup remain required before final G009 ACCEPT or dependent execution.

## Failed staging gate / bounded correction authorization — 2026-10-14 UTC

**CORRECTION LOOP — hosted acceptance remains blocked.**
Pipeline `2946564278`, merge `e087f2c`, required gate job `17169253060` failed
after 2527.7 seconds. Fast job `17169253059` succeeded. Codex independently read
the job status, trace and published artifacts: B-WEB-ERROR-OUTPUT failed (WE4
`vacuous`); the required-check aggregate failed accordingly. All five B-OBJ-*
checks passed. Owned cleanup reports verified, four resources removed and zero
remaining. Preserve the failed run; unrelated-path evidence does not waive it.

WE4 deliberately throws a planted canary through an unwrapped test-only route,
then requires its presence in Docker logs. The artifact records its presence
test as failed; the retained artifact cannot establish why. The current check
uses a two-second sleep followed by `docker logs | grep -qF` under pipefail.
Delayed capture is a hypothesis. Early grep exit causing upstream SIGPIPE is
another possibility to inspect; neither cause is asserted as confirmed.

Authorize a bounded **post-integration correction pass 2 / attempt 3 of B5**:
Claude investigates this one control, reproduces or otherwise characterizes its
failure and makes only the minimum justified correction in
`infra/validation/run.sh`'s WE4 control block. This extends the previous section-
specific scope for this exact purpose, not general harness work. Production web
code, fixture semantics, security assertions, other checks and CI profiles remain
unchanged. Any additional path or broader change returns for review.

Required guarantee: a successful request still has status 500 and the exact
canary must actually reach the captured container logs. Capture/read failures
and an absent canary must still fail. Avoid early-consumer pipeline ambiguity
and, if warranted, use bounded synchronization rather than an unbounded wait or
weaker expectation. Keep canary-bearing raw material ephemeral as today.

Use one implementation owner and existing fixtures/tools. Run the focused WE4
positive and absent-canary/read-failure covering cases, then affected web-error
checks; reuse unchanged source evidence. No blind full-job retry, new test
framework or broad local requalification before the focused diagnosis/fix is
review-ready. Preserve the diagnosis, exact delta and outcomes in the Result.
Checkpoint after attempt 3 is due with the hand-back.

Use a temporary correction branch from current staging, reconcile before testing,
safekeep and return for Codex diff review. The normal MR/Duo/conditional-integration
route and required staging disposition apply to a code correction; this is not a
direct staging-write exception. If no fix is supported, return the evidence for
a retry decision rather than patching speculatively. G010 Plan is accepted but
its implementation remains held; G021 remains closed.

## WE4 correction pass 2 / attempt 3 pre-integration review — 2026-10-14 UTC

**ACCEPT — bounded correction source and focused local evidence for integration.**
Reviewed candidate `7f8da8a5232424e0bb9b33763111ba28cb42aecb` on branch
`g009-we4-correction`, base `6bcddc11aae123687fddc1a66ae99bd7e76c4cb2`.
Result SHA-256 `83e604408307637c13232b6bc19bb4766a9e1d4164976b12973e6dea6ad24c5f`
is imported exactly with this record. A bounded independent read-only reviewer
and Codex inspected the two-path diff and focused evidence. No Codex tests ran.

- The only non-report change is the authorized WE4 block. Complete Docker log
  capture precedes file-based grep, removing early-consumer SIGPIPE ambiguity.
  Exact canary presence and independent status 500 remain required; Docker read
  failure stops fail-closed. The private capture is removed on normal outcomes.
- Twenty non-following log captures with half-second delays cap retry attempts;
  this is approximately ten seconds of polling, not a hard ten-second end-to-end
  guarantee when Docker itself is slow. Existing outer execution limits remain.
- c2-b08 on final harness bytes passes present/absent/read-failure and delayed-
  producer control cases. The old pipeline can report `141 0` despite a match;
  the new capture succeeds for that same mechanism. No retained capture remains.
- c2-b09 on clean `8b5108ed7da2e7b6308481f8eb3cfe8276ae2770` exits 0:
  32 PASS, all 31 required lint/lint-base/root IDs pass, including WE4 and canary
  scans. Candidate changes since that checkpoint are report-only.
- Original hosted cause remains unproven: local timing observations do not rule
  out every hosted delay, and the exact old `grep -qF` pipeline did not fail in
  the local 60-run sample. The early-exit mechanism and its correction are
  demonstrated, without claiming recovered evidence from the failed runner.

### Attempt-3 checkpoint disposition

The scoped strategy is converging. The previously reviewed storage/authorization
criteria and their hosted checks passed; the only failed staging obligation was
the inherited WE4 control, now corrected and locally covered. Failures and the
shellcheck correction remain in the Result. Remaining work is MR/Duo disposition,
qualifying staging evidence, integrated-content reconciliation and bounded final
closeout. No scope split, budget extension or new harness project is justified.
Three attempts used of B5, two remaining. Claude should append/reference this
assessment in its Result checkpoint record before MR handoff; no new test run
is needed for that report-only completion.

Claude may reconcile, push and open the correction MR to staging. Preserve source-
branch deletion; dispose Duo findings and other holds before eligible SHA-bound
auto-merge under standing authority. Reuse unchanged passing local evidence;
no broad local duplicate or optional manual MR gate is requested. The minimum
harness edit still selects required hosted full/CI-controls coverage. Verify
its actual result before final G009 ACCEPT; the failed pipeline remains failed.
G010 execution stays held until that final acceptance and its binding release.

## Second staging gate / single hosted retry — 2026-10-14 UTC

G009 remains **held, not finally accepted**. The WE4 correction merged as
`944ecf3`; its run.sh bytes equal the reviewed `7f8da8a` correction. Relative
to G009 merge e087f2c, all non-management changes are that WE4 block.

Codex independently verified pipeline `2946943462`: required fast job
`17172092974` succeeded; gate `17172092975` failed after 2201.8 seconds. Published
results identify B-DOCS-EXEC-ROOT `start block` plus the missing-check aggregate.
The root suite's later checks, including WE4, were not reached; this run is not
evidence that WE4 failed again. Preserve all missing/failed outcomes honestly.

The published build ledger lacks completed MinIO server steps/export and one
minio-init step, without a captured build error. This is consistent with a build
interruption, but does not establish a network/runner cause. No deterministic
source failure is identified. Trace reports verified cleanup (four resources,
zero owned remaining). The same root setup passed before the WE4-only change.

Authorize **one retry of gate job 17172092975 in the same pipeline on exact
commit 944ecf3**. Do not create a new pipeline, repeat successful fast or start
a local full run. Record new job ID and both attempts' outcomes. No speculative
source patch or extra validation framework is authorized. If the retry reveals
the same build failure or another required failure, return exact evidence for
the next bounded disposition; this is not an unlimited retry policy.

This unchanged-source qualification retry does not consume a new Goal attempt:
G009 remains 3 used of B5, checkpoint already reviewed. Final staging disposition,
integrated-content/cleanup confirmation and Codex final ACCEPT remain required.
G010 Plan stands accepted, but implementation stays held. This management-only
record is not an application validation-input change or a reason to cancel the
authorized retry.

Codex initiated the authorized retry through GitLab: new gate job
`17173244141`, same pipeline `2946943462`, exact commit
`944ecf3e4e295c196a406f56b55f35226b511c35`, initially pending. No additional
retry is authorized while it runs; coordinator monitors and records the outcome.

## Hosted retry independently verified — 2026-10-14 UTC

Gate job 17173244141 (not included)
in pipeline `2946943462` succeeded on exact integrated commit
`944ecf3e4e295c196a406f56b55f35226b511c35`, finished 14:37:17 UTC,
2680.23 seconds (44m 40s). The unchanged successful fast job remains applicable.
Codex inspected the API, published results/summary and final trace: 69 PASS,
zero FAIL/UNMET, root startup and B-WEB-ERROR-OUTPUT passed, all five B-OBJ
checks passed, 441 CI controls passed, required-check aggregate passed.
The summary identifies its pre-publication scope; the final authoritative job
trace reports status 0 with all final phases 0, cleanup_status=0 and
owned_remaining=0. No source change or additional test run was required.

Current staging differs from this tested commit only in management artifacts.
Both earlier hosted failures remain recorded; their causes are not retroactively
proven by this pass. Three B5 attempts used, checkpoint unchanged. Claude is
finishing the report-only Result and local cleanup receipt; final ACCEPT awaits
that AC14 closeout. G010 Plan remains accepted, execution still held.

A bounded independent read-only reviewer confirmed the hosted API/artifacts,
WE4 output and management-only delta. Report-only closeout Result imported
byte-exact from `83be598238cf1a42f1fdcba3afb818cb61a8e49c`; its hosted history
and artifact hashes agree. Local ownership/cleanup confirmation is the sole
remaining AC14 item; do not rerun or reopen unchanged accepted source.

## Final implementation acceptance — 2026-10-14 UTC

**ACCEPT — G009 implementation and AC1–AC14 complete.**

The accepted implementation integrated through MR !24 (`e087f2c`) and the
reviewed WE4 correction through MR !25 (`944ecf3`). Source/correction reviews,
independent security/source review, applicable local evidence, disposed Duo
finding and the successful staging qualification above satisfy the criteria.
The passing fast plus retry gate `17173244141` validate exact integrated commit
`944ecf3e4e295c196a406f56b55f35226b511c35`; later staging changes through this
acceptance are management only. The Result imported from `83be598` is unchanged.
Both earlier failed hosted jobs remain failed historical evidence.

### AC14 closeout verification

Codex read the existing compact `g009-impl-bd31633/CLOSEOUT-RECEIPT.md`, SHA-256
`0e6f219510d28d2375187a05a3b6995077fddea24e8094f964865958b6e5e8f8`, and verified
current worktree/branch listings, Docker listings and absence of the exact
local test-password file. G009 implementation/planning worktrees and six local
branches are removed; no G009-named containers, volume or images remain. The
receipt records removal of the owned cache volume and two focus MinIO images.
Hosted cleanup separately reports status 0 and owned_remaining=0.

The shared official Go image, explicitly retained evidence/history and remote
planning/report branches are held as identified records/shared material, not
unreported live stacks. Preserve G010/G054/live coordinator and unrelated or
unproven resources. No global prune, new cleanup framework or additional tests
were required. B5 closes at three used attempts; checkpoint reviewed, no extension.

### Inherited risks and next action

PG-003 remains a recorded bounded query-cost observation, not a production SLO
or an owner waiver; documented signed-URL/in-flight/compensation residuals remain.
G010 inherits the exact accepted storage, RLS and policy contracts for its
approved integrated attack chains. G011 inherits operational storage issuance
records as inputs to its future durable-audit scope; no G011 implementation or
audit guarantee is introduced by G009. No storage guarantee is deferred to G010.

G009's dependency hold on G010 is released. Under existing owner unattended-loop
authority and accepted G010 Plan r1 (`ca6f2176…`), Claude may prepare its Git-first
execution scope/initial Result and return exact bindings for mechanical Codex
confirmation before application/test implementation. Reuse accepted controls;
no design reopening or additional validation campaign. G021 remains closed.
PROGRESS compaction assessed: this is not epic closeout; retain active G010 and
recent G009 history. The mandatory epic-closeout checkpoint remains with G010.
