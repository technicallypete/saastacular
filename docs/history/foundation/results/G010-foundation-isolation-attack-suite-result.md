> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G010 — Foundation isolation attack suite: Result

## Status

**Integrated; hosted staging qualification PASS (retry); awaiting Codex final ACCEPT.** This is not final acceptance
and does not close EPIC-003. This record is compact and durable. Under the 2026-10-14 owner direction, this Result,
the Codex records and the hosted CI artifacts are the durable evidence. Local raw output was temporary.

| Milestone | Record |
|---|---|
| Plan r1 ACCEPT | G010-plan-r1-codex-review.md (not included) (Plan SHA-256 `ca6f2176cc7019dff8fc8912b90e53951ed5210efcb82ac3566957a551b76a78`, accepted commit `0d4cfe3`; r0 archive `487e500c…`) |
| Execution binding | prepared on staging `325c136` after G009 FINAL ACCEPT (`841e1357…`); Codex mechanical ACCEPT on `908b84c` (same review file) |
| Attempt-1 oracle rulings | same review file, staging `71dd019` |
| Pre-integration ACCEPT | [G010-implementation-codex-review.md](G010-implementation-codex-review.md), candidate `71c8c47`; this Result's previous version (`2a732fab…`) was imported byte-exact |
| Integration | MR !26 → merge `c961e562f1041b51d5f45d0e9bbf172813c030d1` (G010 commits `19bace5`, `e1088b9`, `7a749e9` after the final rebase; the MR/Duo disposition, the conditional integration, the G7d disposition and the retry verification are all in the implementation review file) |
| Hosted staging gate | pipeline `2947802084`: first gate job `17179052627` FAILED (one CI control, below); the authorized same-source retry, job `17180444863`, **PASSED** |

## Scope delivered

- `core/internal/tenancy/access/attack_expected_test.go`: the independent oracle **r1**, SHA-256
  `e50f30cdbe50eed7bae6cf9574aabcd3c9ccedce58b99ab0a00afa2be25c7582`, by the G010 oracle delegate. It holds 82 literal
  expectations for the ten composition steps, a full R1 world, and `Inherited`/`Source` fields. r0 (`0635567d…`) was
  replaced per the Codex rulings: E3-02 returns `ok` with `deletion_pending`, and D1-17 cites the dynamic RLS-SCOPE
  last-path subtest.
- `core/internal/tenancy/access/attack_integration_test.go`: `TestIsolationAttacks`. It has ISO-ATTEST, chains A–E
  (A1, A2, B1, C1, C2, D1, D2, E1, E2, E3) and ISO-CITATIONS. Every expectation is checked exactly once, by ID, with
  the Plan §1 denial evidence:
  - the exact code;
  - a paired control that was checked earlier and agreed;
  - zero provider calls;
  - unchanged Head presence, object rows and tenancy rows.

  E1 and E3 instead assert their staged effects plus a working control. E3 runs on the distinct seeded live object,
  and no provider call follows its revocation.
- `infra/validation/run.sh` `obj_it`: a second named invocation inside B-OBJ-IT. It pins 17 subtests and
  `ISO-ORACLE {"cases":82,"disagreements":0}`, and checks the access log of both logs. No new check ID.
- `docs/foundation/AUTHORIZATION.md` gains the "Isolation attack suite (G010)" section. `OBJECT-STORAGE.md` and
  `FOLLOW-UP.md` get one-line updates.
- No production code, schema, contract, capability, dependency, CI framework or allow-list change. The allow-list is
  G010-scope-allowlist.txt (not included), `72b31a4a…`, with 4 patterns. Architecture deviations: none.

## Key commands and outcomes

Every local batch first ran `git fetch` and the exact-delta/ancestry check, then
`bash infra/ci/docker-desktop-mounts.sh`, then took the harness flock. The harness form was
`bash infra/validation/run.sh --base <target> --allow-file management/v0/results/G010-scope-allowlist.txt
--report-dir <dir> <suites>`.

| Evidence | Candidate / target | Outcome |
|---|---|---|
| Focused `go test -tags integration -run '^TestIsolationAttacks$'` (oracle r0) | 908b84c+wip / 908b84c | 81/82. The two findings went to Codex rulings; this failure is kept as history |
| Same, oracle r1 | cc771f8 / 71dd019 | PASS: ISO-ORACLE 82/0, 17/17 |
| Security review (read-only, `s-review` SHA-256 `b862ed25939f22ebd1070ffcf984ca5dbf17c5457c683dccc79165bad3285525`) | e484cc7 / 71dd019 | FINDINGS, highest Low. Four Low findings and one Info corrected in the glue and `obj_it`; focused rerun PASS 82/0, 17/17 |
| AC8 control: one disposable, never-pushed commit flipping C1-04 `forbidden` → `ok`; `run.sh … migrations` | f8f412a / 71dd019 | **FAILED as required**, exit 1. Only B-OBJ-IT (oracle 82/1) and its dependent aggregate failed; the focused output named only `C1/C1-04`. Restoration verified (HEAD, clean tree, oracle bytes) |
| b4 `run.sh … all unit` | 3f1f3f2 / 71dd019 | 96 PASS. B-OBJ-IT PASS including TestIsolationAttacks. B-CI-CONTROLS failed at 440/441 (control VS2), and B-REQUIRED-CHECKS failed with it |
| Failed-suite rerun `lint lint-base provenance ci-controls` | 10eb021 / d56665f | VS2 passed, VC7 failed once. VS2 and VC7 each passed in focused runs |
| Failed-suite rerun `lint lint-base provenance ci-controls` | 36e1322 / 31963ae | **PASS**: 21 PASS, B-CI-CONTROLS 441/441, B-REQUIRED-CHECKS 20/20 |
| Hosted gate job `17179052627` | c961e56 | FAILED: B-CI-CONTROLS 440/441, control **G7d** ("the guard was not identified exactly once"); cause unproven |
| Focused G7d diagnosis ×3 (`validate-controls.sh --only G7d`) | c961e56 | 3/3 PASS, `killed` count 1 each time; see the Codex G7d records in the implementation review |
| Hosted gate retry job `17180444863` (19:09:30–20:01:37Z, 3126.8 s) | c961e56 | **PASS**, status 0: 69 PASS, 0 FAIL; B-CI-CONTROLS 441/441 including G7d; B-REQUIRED-CHECKS 68 IDs; B-OBJ-IT PASS with TestIsolationAttacks |

The G010 files are byte-identical across the b4 candidate, the b16 candidate, the reviewed `71c8c47` and the merge
`c961e56`. Every staging advance during attempt 1 was management, governance or docs only and was reconciled by
rebase.

**Hosted evidence hashes.**
- Failed gate: `ci-controls.log` `b1864548ee5e2c8660eabeab11ebc1729b3bb4e810d58dfa93eefa6f5f9c7df6`, `results.tsv`
  `f46bc32232d606063f7e0fc6b590685d0005533c052b29092901186730b74177`.
- Retry: gate log `fee314214d8d4d8170dccc625c6e407bad83a6bca413aecee878a40daaf58758`, `results.tsv`
  `3704cbc97b8926ff5f7641ac536c5f2bdbafe20120dbadef671e9059fc7617ae`, `summary.json`
  `28e227a251b7c3ad1d21a5dcdbe5f2363e290f4f18d3d89f8f3ad8c6560a03e0`.
- The GitLab job artifacts are the durable copies. Artifact expiry is recorded, not mirrored.

**Unproven CI-control failures.** VS2, VC7 and G7d each failed once in a full run and passed in focused reproductions
and in another full run. G010 changes none of the scripts they exercise. Their causes are not established, and the
failures are not rewritten as passes.

## Acceptance criteria

| AC | Evidence |
|---|---|
| AC1 | oracle r1 (every step, literals, `Inherited`, `Source`, authorship, actor roster with cited role coverage); ISO-CITATIONS PASS; AUTHORIZATION.md section |
| AC2 | A1-01..14 and A2-01..04 agree, including exact SQL sets for both shapes, the 0-row cross-tenant UPDATE and zero BeginTx for binding faults |
| AC3 | B1-01..11 agree: the shared profile gives no sibling path through the facade, storage or SQL |
| AC4 | C1-01..05 and C2-01..14 agree: the creator is the subject; own delete allowed, another member's denied, admin delete allowed; Group-only access is `context_denied` and every tenant relation is empty |
| AC5 | D1-01..17 and D2-01..05 agree: committed changes on a retained scope; one backend with the same PID and empty settings |
| AC6 | E1, E2 and E3 agree: staged downgrade compensated and never readable; URL 200 while logically deleted and non-2xx after physical removal before expiry; `deletion_pending` repaired by a repeat delete |
| AC7 | ISO-ATTEST; privileged setup kept separate; spy counts; access-log checks on both logs; B-RLS/B-OBJ catalog and unit boundary PASS (b4, hosted retry) |
| AC8 | the C1-04 reversal failed the existing gate on exactly that expectation; restored |
| AC9 | the table above; hosted qualification PASS on the retry |
| AC10 | B5 ledger below; cleanup and the EPIC-003 closeout disposition are owed (below) |

## Attempt ledger (B5)

| Attempt | State |
|---|---|
| Budget | B5, checkpoint after attempt 3. Planning and binding prep spent none |
| 1 | **Used.** Covers implementation, the oracle r0→r1 rulings, the S-review corrections, the AC8 control, local coverage, integration, the G7d diagnosis and the hosted retry. Rulings, diagnosis and the retry spent no further attempt |

## Ownership

| Role | Identity |
|---|---|
| Implementer | task G010 / impl, agent `a20bcd2a68bc4b58d`, native worktree `.claude/worktrees/agent-<id>` (G010-only exception), effort high as instructed (actual unknown) |
| Oracle delegate | task G010 / oracle, native Agent, no worktree, effort high |
| Security reviewer | read-only native Agent (type Plan), effort high |

## Local evidence and proof holds

**Necessary proof holds: none.** This Result, the Codex review records and the hosted job artifacts carry the proof.
The local folders are temporary and eligible for purge at closeout: `g010-impl-325c136`, `g010-oracle-908b84c`,
`g010-plan-83a5784`, and the coordinator's hosted-gate copies. Before purging, take a compact receipt that includes
the oracle r0 archive hash `0635567d…`.

## Closeout owed

- Codex final ACCEPT and the mandatory EPIC-003 epic-closeout PROGRESS compaction disposition, including the
  G008-deferred movement (Plan §8). Claude does not edit PROGRESS.
- Bounded cleanup of the implementer's owned resources: the coordinator's list. G021 stays closed.
- Handoffs per Plan §8: G011 audit mapping and read isolation; G013 tombstones and pending rows; G014/G015 scope
  trust; G018/G019 membership administration. No new scope is authorized here.
