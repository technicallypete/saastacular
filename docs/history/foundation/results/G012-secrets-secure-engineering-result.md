> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G012 — Secrets and secure engineering controls: Result

## Status

**Integrated; hosted staging qualification PASS; awaiting Codex final ACCEPT.** This is not final acceptance. Codex
accepted the source and local evidence (after two correction passes) and gave conditional integration ACCEPT; MR !28
auto-merged at the reviewed head and the hosted staging gate passed (section "Integration and hosted staging gate").
This Result, the Codex records and the hosted job artifacts are the durable evidence; local raw output was temporary.
The paragraph and rows below this one keep their attempt 1 wording as history. Every
approved part of Plan r1 is delivered with the D0 dispositions D1 to D4 applied. The read-only security review returned
NO-BLOCKING-FINDINGS; its three Medium findings and five of its seven Low findings are fixed, and the other two Low
findings are recorded deviations. The final local batch (`all unit`) is recorded under "Commands and validation".

| Item | Value |
|---|---|
| Candidate (source tested by the final batch) | `c9f874e4cffb515f33da3215ea278f7df0c2bb78` |
| Target (origin/staging at every pre-batch fetch since the last rebase) | `34e036258b486b079867ad8e277393b25e376ea3` |
| Final local batch | `all unit` on `c9f874e`: RESULT PASSED, rc 0; every one of the 104 required IDs PASS exactly once (B-REQUIRED-CHECKS counts the 103 others); 2026-10-15 03:21 to 04:14 UTC (53 min); no leftover resource |
| Branch | `g012-secrets-secure-engineering` (pushed; this Result is committed after the final batch, on top of the candidate) |
| Effort | high as instructed; actual runtime setting unknown (not verifiable from inside the worker) |
| Budget | B5: 3 of 5 attempts used (attempt 1, correction 1 as attempt 2, correction 2 as attempt 3); the checkpoint after attempt 3 is recorded in "Correction 2" |
| Integration | MR !28 auto-merged (armed by the coordinator under Codex's conditional integration ACCEPT, handoff 6692422e) at the reviewed head `9c2798612e4916771e45aa69238c9c5cf185aa00`; merge `5e1e98735f6224888889d516ef2397a0ea8dc89a`, whose tree `8cbb5d2e74c395fe6d2212f42f23429603606a45` equals the reviewed head's tree; remote source and review branches deleted |
| Hosted staging gate | pipeline `2949510391` **success**: fast job `17191203943` success; gate job `17191203944` **PASS** (73 PASS, 0 FAIL); success qualifies the CI acceptance gate under the owner status rule |

## Execution release and input reconciliation

- **Release:** execution-binding confirmation (not included) on staging `c59c8d8`
  (handoff 420a0bd1). Per its provenance clarification, the eleven planning inputs bind to base `4c240b1`, while the
  original allow-list and the initial Result (below) bind to the prepared candidate `023d394`.
- **Rebases (each management-only, each followed by the README secret-scan line on the rebased tree):**
  - onto `c59c8d8` (release; `023d394` dropped as already upstream);
  - onto `ed57368` (D0 dispositions D1, D3 and D4, D2 held; the 19-pattern allow-list amendment);
  - onto `34e0362` (owner decision on D2).
- **Disposition:** G012-D0-codex-disposition.md (not included) at `34e0362`.
- **Bound bytes at the target** (neutral labels on hash lines):

| Label | SHA-256 |
|---|---|
| D0 disposition with the owner decision section | `c7dde64fe0d6dfabd141574d5283ac40ef44e3de1c4c5dd707135db1e2bc16bc` |
| Execution-binding confirmation | `260d370d6fc56969815db73cd74c6475f9a684ccf30225cd81393369229f0026` |
| Allow-list, 19 patterns (amended) | `30d2af3e9a686afad88709d757ae8e84eeb762fef34e981c6ade3236d2680f58` |
| Allow-list, 15 patterns (original, historical) | `19134a9a89caab0c8588b71a7edec4fe3e968044695432d60e987df07d2a491d` |
| Initial Result (historical, kept below) | `3ec474b428a8776c604dc0f1e70050ed8baeba70d223677378efa83efff2bad4` |
| Plan r1 (unchanged) | `cd90a79099bf320fadd9aad39ef4501d1971fb339d6053a533efe69f7f70d51a` |

All 54 changed paths are inside the 19-pattern allow-list (B-SCOPE). `docs/SECURITY.md` is a protected file bound by
`--expect-hash` in each batch (its value per batch is in the evidence map).

## D0 baseline and dispositions

D0 ran the pinned tools only, with the §5.2 route, on `c59c8d8` (attempt 1 began here).

- **Advisory fetch:** all four objects passed every check (HTTP 200, size equal to Content-Length and under the cap,
  MD5 equal to the published value, zip test, Last-Modified within 72 h; Go database `modified` 2026-10-08, within 30
  days). Sizes: Go 12.3 MB, npm 218.4 MB, PyPI 36.0 MB, Go database 3.4 MB.
- **Contract mismatch (D1):** osv-scanner v2.6.0 reads its offline databases from `<cache>/osv-scalibr/<ecosystem>/`
  rather than the documented `osv-scanner/` subdirectory; with the planned layout and the read-only mount it exited
  127. A diagnostic with the corrected layout loaded all three databases (exit 1, parseable output, about 10 s).
  Codex approved the corrected path (no pin, flag or failure change). The original failure and the diagnostic are kept
  as evidence; no repeat campaign.
- **Identity baseline:** Go 36/36, npm 472/472 identities (478 locations), PyPI 29/29 (the virtual root
  `intelligence` excluded); no unsupported record.
- **Findings:** 13 Go standard-library advisories in go1.27.1, all fixed in 1.27.2 (nine called: GO-2026-6603, 6605,
  6607, 6608, 6610, 6611, 6612, 6613, 6617; imported only: 6604, 6609; module only: 6599, 6600), and npm `braces`
  3.0.3 (GHSA-vfj7-8cjw-p6xm, alias CVE-2026-93687, High 8.7, development only, no fixed release).
- **D3 (approved):** Go 1.27.1 to 1.27.2 across `core/go.mod`, `infra/validation/tools/go.mod`, the core, migrate,
  MinIO and validation-tools Dockerfiles and the image inventory in `run.sh` and `lib.sh` (index digest
  `f92b6ef8…7e12d`, linux/amd64 `3b15ba43…35f8`). The core and MinIO Dockerfiles are registered customizations
  (register 26 to 28) and carried to the disabled variant; generator, skill templates and original proposals stay at
  1.27.1. With Go 1.27.2 **all thirteen standard-library findings are cleared**; no Go exception exists.
- **D4 (confirmed):** called, imported-only and module-only Go findings without a severity are Unclassified and block.
- **D2 (owner-approved, conditional):** the exact braces exception is in place after both conditions were established
  (next section). The finding stays visible as EXCEPTED in every B-DEPSCAN report; it is never reported as fixed.
- **Disclosed deviation (once):** the D0 driver called `govulncheck -version` during phase A while the network was
  allowed, which made one extra metadata request to `vuln.go.dev` beyond the four fixed objects. No package, graph,
  source or report was sent. `infra/security/dependency-scan.sh` never makes this call; version provenance comes from
  the offline `config` message of the scan itself.
- **Cost:** a complete local B-DEPSCAN takes about one minute plus image pulls (fetch 24 s for about 280 MB, phase A
  12 s, fixtures about 6 s, current tree about 12 s), inside the accepted +3 to 6 min forecast and the unchanged
  20-minute fast cap. Hosted timing is measured by the MR fast job.

## D2 braces evidence (durable record)

Collected on candidate `87ffc7108c9f227658386bbce7794094b4839423` over target `34e0362`, with three untracked
in-progress paths (`infra/security/dependency-exceptions.tsv`, `infra/security/dependency-scan.sh`,
`infra/security/testdata/`) and no change under `web/`; G012 changes nothing under `web/`, so the web tree is identical
in the final candidate. Images were built from `web/` with `infra/docker/web.Dockerfile` and removed afterwards.

| Label | Value |
|---|---|
| Production web runtime image (target `runtime`) | `sha256:0b8e335f5c67a3e33679f9fffc09f6704b24cb24768feb9fcc3f57d14ce2976f` |
| Lint toolchain image (target `deps`) | `sha256:87205a281bcffc1ecebd51d15bcf84e32a20fa3d0cb0daa0736e96552ded328f` |
| Evidence script 1 (images, package search, npm ls) | `cf6105753435e870276f5ee1da4c23844b6fd54197715335ab7fc8c9cde64581` |
| Evidence script 2 (mention context, call path, declared dependents) | `28a252443e7687cae003591a787f1adfcfa659490698a58e706c03388b67f7cb` |
| Evidence script 3 (resolved ESLint settings, target `source`) | `d820d03e36ef29c5ae46cc8826e58e35b3cd8cdb1a103f71c25d152a269423ba` |

Conclusions:
1. **Absent from production artifacts.** The runtime image (Next standalone output on `node:24-slim`) holds no `braces`
   directory, no `package.json` named `braces`, `micromatch` or `fast-glob`, and the base image's global npm bundle
   holds none of them. The standalone `node_modules` top level is `@img @next @swc client-only detect-libc next react
   react-dom semver sharp styled-jsx`. The only text mentions are a React development warning string, Next's vendored
   `picomatch` (a different package with its own brace counters) and a comment in `csrf-protection.js`.
2. **One development path, one loader, trusted input only.** `npm ls braces`: `eslint-config-next@16.3.8` to
   `@next/eslint-plugin-next@16.3.8` to `fast-glob@3.3.1` to `micromatch@4.0.8` to `braces@3.0.3`; no other package
   declares any of them. The only source file that loads `fast-glob` is
   `@next/eslint-plugin-next/dist/utils/get-root-dirs.js`, used only by the `no-html-link-for-pages` rule. Its only glob
   input is `settings.next.rootDir`; the resolved ESLint configuration for `app/page.tsx` has no `settings.next`, so the
   rule uses the working directory and never passes a pattern. ESLint runs only in the web check stage over repository
   files; no request or customer input reaches it.

The exception row binds npm `braces` 3.0.3 and GHSA-vfj7-8cjw-p6xm (the alias matches through the finding), names
<owner> (product owner), expires 2026-10-29 (inclusive), cites the committed disposition and states the
review and revoke conditions (a fixed release, a version change, production inclusion, or exposed pattern input). Since
the security-review fix M2, vulnpolicy also fails the check automatically if an excepted npm or PyPI identity is not
development-only, so production inclusion revokes the exception without a manual step.

## Implementation summary and files changed

54 paths changed since `34e0362` (`git diff --stat`: 4272 insertions, 51 deletions); commits `73cef1d`, `87ffc71`,
`c43f039`, `e8b33dd`, `8e5d4b0`, `c9f874e` (safekeeping pushes after each round).

- **R1:** `core/internal/config/secrets.go` (sentinels, `resolveSecret`) and `secrets_test.go` (U-SEC-1 to U-SEC-6);
  `config.go`: `EnvSecrets` wraps `ErrSecretNotSet` and the secret closure calls `resolveSecret`.
- **R5:** `core/internal/config/readiness.go`, `readiness_test.go` (U-RDY), `core/cmd/readiness/main.go`, `main_test.go`.
- **R3:** `infra/security/secret-scan.sh`, `infra/compose/secret-scan.yml`, the README `platform:check` scan line and
  one sentence; B-SECRET-SOURCE in `run.sh` (`lint`).
- **R4:** `infra/security/dependency-scan.sh`, `infra/security/dependency-exceptions.tsv` (header plus the braces row),
  `infra/security/testdata/{govulncheck,osv,exceptions}/…` (synthetic data only),
  `infra/validation/tools/cmd/vulnpolicy/{main,parse,policy,exceptions}.go` and `policy_test.go`; B-DEPSCAN in
  `run.sh` (`lint-base`), with a `DEPSCAN-PINS` INFO row.
- **R2:** B-SECRET-ROTATION in `run.sh` (`ops`, last on the disposable copy).
- **D3:** the seven Go-patch paths; `render-customizations.tsv` (28 rows; reasons for `core/go.mod`, `config.go`,
  `VENDORS.md`, `FOLLOW-UP.md`, two new Dockerfile rows); `APPLICABLE` gains the new core files and both Dockerfiles.
- **R6:** `docs/SECURITY.md`; `docs/foundation/VENDORS.md` (the reviewed run extended from 5 to 33 lines, pinned in
  `run.sh`), `OPERATIONS.md` (security checks and counts), `FOLLOW-UP.md` (STAGE-OPS status), `DATABASE-ISOLATION.md`
  (no-rotation pointer).
- **Harness wiring and pins:** `required-checks.tsv` (110 rows; lint 9, lint-base 4, ops 13), `validate-controls.sh`
  count pins, B-SHELL list, render-register count, VENDORS constants, toggle-off PASS text.

**Count-pin changes** (inventory at binding plus the G012 delta): `validate-controls.sh` 948/952 fast 33 to 35;
1070/1079 affected web 19 to 20; 1112/1115 remainder 72 to 73; 1206 VU1 33/72/101 to 35/73/104; 1207 VU9 33/42/71 to
35/43/74; 1937 19 to 20; 1948 72 to 73. OPERATIONS.md lines 90 to 93, 99, 100, 120, 121 (fast 35; affected web 20, core
16, intelligence 20; integration 74; full 103; lint nine; lint-base with B-DEPSCAN; 35/73/104; 35/43/74).
`run.sh` render register 26 to 28.

## Migrations, schema, contracts

- Database, schema, migrations, roles, grants, policies, audit events: **none**. The rehearsal alters only the
  disposable copy's role password and is torn down with the copy.
- Public API, MCP and BFF contracts: **none**. New internal Go API: the config sentinels and `AssessReadiness`; the
  `core/cmd/readiness` tool is in no runtime image and serves no route.

## Tests added

- Core: U-SEC-1 to U-SEC-6 (`secrets_test.go`, including the acquisition boundary: one `.Secret(` call in
  `resolveSecret`, one `resolveSecret` call in `Load`, one `config.Load` call in the server main); U-RDY-1 to U-RDY-6
  (`readiness_test.go`); command exit codes 0/3/2 (`main_test.go`).
- vulnpolicy: 12 test groups (class boundaries, merge by alias, reachability at every level, exceptions including
  stale, unused, mismatched, wildcard, range, group, duplicate and the development-only rule, identity reconciliation
  with equal-count omission and substitution negatives, unsupported records, the govulncheck contract, malformed
  outputs and exit codes, ungrouped vulnerabilities, fixture-db).
- Harness: B-SECRET-SOURCE (S0 to S7 including the gitlink and RE2-only cases), B-DEPSCAN (fixed offline fixtures plus
  the current tree), B-SECRET-ROTATION (R0 to R6).

## Commands and validation (attempt 1)

Every batch: `git fetch origin` with the exact-delta or rebase rule, `bash infra/ci/docker-desktop-mounts.sh` (10 %
each time) and the local harness flock (`local-harness.lock`), through a wrapper that records candidate and target
SHAs and the dirty-file list. Raw output is temporary local evidence under
`<home>/.local/state/platform-evidence/g012-impl-4c240b1/` (`<EVID>`), with script hashes in `scripts.sha`
and the batch table in `batch-log.md`. Harness form: `bash infra/validation/run.sh --base 34e036258b486b079867ad8e277393b25e376ea3
--allow-file management/v0/results/G012-scope-allowlist.txt --expect-hash docs/SECURITY.md=<sha> --report-dir <EVID>/<batch>/report <suites>`.

Commit labels used in the table (hash-free rows): C0 `c59c8d8` (release), C1 `a1d0945`, C2 `87ffc71`,
C3 `c43f039`, C4 `e8b33dd`, C5 `8e5d4b0`, C6 `c9f874e` (the final candidate), T `34e0362` (the target).

| Batch | Candidate / target | Commands | Outcome |
|---|---|---|---|
| pre-d0-scan | C0 / C0 | README secret-scan line (then the Compose line) | rc 0, no leaks |
| d0 | C0 (clean) | D0 driver: fetch, osv-scanner, govulncheck phases A and B | fetch rc 0; osv-scanner rc 127 (layout); govulncheck rc 0 with 13 findings; cleanup verified |
| d0-diag-layout, d0-layout | C0 | layout diagnostic; D0 with the corrected layout | the scanner creates `osv-scalibr`; corrected layout: rc 1 with parseable findings, 10 s |
| pre-rebase scans | C1, C2 | README scan line on each rebased tree | first: 2 generic-api-key hits in my own new test literals (rewritten before commit, no value printed); then clean |
| b1-core-01 | C0 plus R1/R5 work | README core-check line | rc 0, every core package ok |
| b1-vtools-01 | C1 plus work tree | validation-tools check stage | rc 0 |
| depscan-dev-01 | C2 plus work tree | `dependency-scan.sh scan` | fetch, phase A, tools and both fixture parts PASS; current FAIL on braces only (before the exception); all 13 Go findings cleared; 57 s |
| braces-evidence 1 to 3 | C2 plus work tree | section "D2 braces evidence" | both owner conditions established |
| b1-01 | C3 / T | `lint lint-base unit contracts provenance` | **FAILED** rc 1: B-SECRET-CANARY found the B-SECRET-SOURCE fixture canaries left under the harness work directory, plus the consequential B-REQUIRED-CHECKS; every other check PASS (B-DEPSCAN, B-SECRET-SOURCE, B-UNIT-CHECKS, B-CONTRACT-TOOLS, B-RENDER-PLAN with 28 customizations, B-PROTECTED, B-SCOPE, B-SHELL, B-DEPS); 6.5 min |
| b1-02 | C4 / T | `lint` (the fix removes the fixture trees) | rc 0, 12 required IDs |
| readiness-doc-run, run2 | C4, C5 | the documented readiness commands | first form (go run) reported exit 1 and mixed build output into stdout, so the documentation changed; second form: sample exit 3 (blocked), secret-named input exit 2 without the value |
| b2-01 | C5 / T | `ops off` | rc 0: 21 PASS (B-SECRET-ROTATION, B-OBS-*, B-OUTAGE, B-PERSIST, B-IDENTITY-IT/E2E/PROBES/PRIVACY/CONFIG, B-TOGGLE-OFF-ADDITIONS, B-OBS-OFF, B-BROWSER-OFF, B-TOGGLE-OFF and the final checks); 20 min; no leftover |
| S-review | C5 | read-only security review (Delegate 1, dispatched by the coordinator) | NO-BLOCKING-FINDINGS; 3 Medium, 7 Low |
| sr-vtools | C5 plus fixes | validation-tools check stage | rc 0; vulnpolicy 12 test groups PASS |
| sr-focus | C6 / T | `lint lint-base unit` | rc 0, 17 required IDs (B-SECRET-SOURCE with the RE2 case, B-DEPSCAN, B-UNIT-CHECKS with core `go test ./...`); 5 min |
| pre-final-scan | C6 | README secret-scan line (the new script) | rc 0, 1130 regular files and 1 link, no leaks |
| final-01 | C6 / T | `all unit` (ci-controls auto-selected) | **PASSED** rc 0: 104 required IDs PASS, 0 FAIL, including B-CI-CONTROLS (441 controls), B-SECRET-CANARY (220 generated secrets), B-SCOPE (54 paths), B-PROTECTED (549 files, one expected hash), B-RENDER-PLAN (28 customizations), B-DEPSCAN, B-SECRET-SOURCE, B-SECRET-ROTATION, B-CHECKS-ROOT, B-UNIT-CHECKS and every inherited identity, audit, RLS, object and toggle check; 53 min; inventory leftover empty, no owned container or volume left |

B-SECRET-ROTATION in b2-01: R0 ready with one runtime backend; R2 the pre-rotation connection still served readiness
200; R3 the old value failed password authentication and the new value logged in; R4 one backend terminated, 503 after
0 s; R5 source updated and core recreated, 200 after 5 s on one post-rotation backend; R6 `PASSWORD NULL`, login
failed, one backend terminated, 503 after 0 s. The final batch reproduced the same R0 to R6 values.

**Evidence hashes** (local, purged at closeout): b1-01 results `9ff66717…1444`, run log `22e32ee1…5405`; b1-02 results
`f2f275ee…c2c0`; b2-01 results `1daf09fb…f143`, run log `a492d23f…6a21`; depscan-dev-01 run log `1dccd00b…19d6`;
sr-focus results `9db41276…7001`, run log `846735dc…5096`; S-review report `ac403a8d…8c18`. final-01 results `b03fd702…53cb`, run log `dfce3e26…2fe4`, B-DEPSCAN step results `f4efaf11…5363`.
`docs/SECURITY.md` expected values: `a24cb157…5559` (b1), `145f0156…4724` (b2), `e6bb1a70…555f` (sr-focus, final).

## Security review dispositions (Plan §12, Delegate 1)

Reviewer: native read-only agent dispatched by the coordinator on `8e5d4b0` (report `<EVID>/s-review.md`). Verdict
NO-BLOCKING-FINDINGS.

| Finding | Disposition |
|---|---|
| M1 exemption guard could fail open (grep error read as no match; RE2 read as ERE) | **Fixed** (C6): grep exit 0 gives 3, 1 continues, any other exit gives 2; RE2-only syntax (flags, non-capturing groups, Perl classes, lazy quantifiers) is refused with 2; new B-SECRET-SOURCE S7 case with an RE2-only pattern (exit 2), passed in sr-focus |
| M2 production inclusion was not detected automatically | **Fixed**: vulnpolicy reports a problem when an npm or PyPI exception covers an identity that is not development-only; unit test on npm and PyPI runtime identities; fixture exceptions now cover development-only identities |
| M3 braces evidence not durable | **Fixed** in this Result (section "D2 braces evidence") together with the one-time D0 disclosure |
| L1 Compose failures mapped to exit 1 | **Fixed**: exit 1 only when the overlay printed its findings line; otherwise 2 |
| L2 silent truncation of readiness input; empty origin list | **Fixed**: input over 64 KiB is refused; an origins value with no origin is blocked-unknown; tests added |
| L3 boundary test too loose | **Fixed**: exactly one `resolveSecret` call (in `Load`) and one `config.Load` call (in the server main) |
| L4 volume reuse and stale `--out` | **Fixed**: existing volumes and a non-empty `--out` are refused before anything is created; the size cap after download stays (informational; the 300 s bound and the cap check still fail it) |
| L5 ungrouped osv-scanner vulnerabilities | **Fixed**: incomplete result; unit test |
| L6 submitted set staged in a host temporary directory | **Deviation recorded**: the set is copied to a 0700 `mktemp -d` directory (removed by the trap) and streamed from there, because one tar must also carry the link-target listing and the counts; no bind mount and no network, as planned |
| L7 Go containers run as root | **Deviation recorded**: the owned module-cache volume is created root-owned and making it writable for another user would need CAP_CHOWN; the containers keep `--cap-drop ALL`, `no-new-privileges`, no Docker socket, no environment secret and network none for every scan; osv-scanner runs as 65534 |

## Acceptance criteria

| AC | Evidence | State |
|---|---|---|
| AC1 inventory and provider substitution | `docs/SECURITY.md` section 3 checked against config, Compose, call sites and the identity schema (recoverable versus hash-only fields); U-SEC-2 replacement provider and U-SEC-1 `EnvSecrets` through the same `Load`; U-SEC-5/6 startup-only acquisition; disabled paths labelled; no new runtime service | met (final batch) |
| AC2 fail-closed acquisition without disclosure | U-SEC-3/4 (closed texts, canary-bearing provider errors never disclosed, `Config{}`); existing WorkOS file-failure tests; B-SECRET-CANARY | met |
| AC3 rotation and withdrawal demonstrated | B-SECRET-ROTATION R0 to R6 (b2-01 and final-01); synthetic values; the copy torn down by the existing `down -v` | met |
| AC4 least privilege and deprovisioning intact | `docs/SECURITY.md` sections 5 and 6.1; unchanged accepted proofs in the final batch (B-IDENTITY-IT, B-AUDIT-IT, B-POLICY-IT, B-RLS-IT, B-OBJ-IT including the G010 attack suite), B-COMPOSE, B-MIGRATE-ENV, B-IDENTITY-GUARDS | met (final-01 PASS) |
| AC5 submitted-source checks fail closed | README scan through B-UNIT-CHECKS and B-CHECKS-ROOT on the real tree with its tracked link; B-SECRET-SOURCE S0 to S7; gitleaks configuration and lint unchanged | met |
| AC6 enforceable dependency checks for three graphs | B-DEPSCAN: identity reconciliation of all three graphs (roots excluded, development membership, go1.27.2 standard library), pinned tool and advisory provenance, fixed offline fixtures, vulnpolicy units including equal-count negatives; D0 triage; every real finding corrected (13 Go) or disposed (braces, owner-approved, conditions proven, development-only enforced) | met |
| AC7 encryption readiness rejects unsafe posture | `docs/SECURITY.md` section 7; U-RDY negatives and the labelled synthetic positive; the documented command blocked on the current shape; no encryption claim | met |
| AC8 vendor and security records complete | VENDORS run (current integrations, scanners, advisory data, data-handling section with unknowns as production prerequisites), SECURITY.md, F-1 to F-8, the private-backup handoff to G013, G013/G056 and adopter responsibilities; no schema, provider, product surface, MFA, audit bypass or compliance claim | met |
| AC9 attributable, bounded CI evidence | this evidence map; actual selection full plus ci-controls; the single final local batch; failure paths and output privacy preserved; failed and reused evidence recorded; hosted: MR fast job `17189970959` FAILED (cidfile defect, corrected in attempt 3), job `17190291257` and final MR pipeline `2949481827` success, staging pipeline `2949510391` success with gate job `17191203944` 73 PASS (section "Integration and hosted staging gate") | met |
| AC10 review and closeout | this Result, the S-review and its dispositions, the B5 ledger, integration and the hosted staging PASS; Codex final ACCEPT, the cleanup receipt, PROGRESS and the G013 handoff remain ("Closeout owed") | pending Codex final ACCEPT |

## Architecture and Plan deviations

No architecture deviation. Plan deviations, each within approved scope:
1. D1 (approved): the osv-scanner database path is `osv-scalibr`, not the documented `osv-scanner`.
2. D3 (approved): Go 1.27.2 across the seven approved paths, two new registered customizations (26 to 28).
3. D2 (owner-approved): the exceptions file holds one exact braces row instead of the header only.
4. `config.go` changes in two adjacent hunks, not one (accepted by the security review).
5. The scanner pins are reported in a `DEPSCAN-PINS` INFO row and the B-DEPSCAN PASS text, not in the IMAGES INFO list:
   the harness computes that list identically in `lib.sh`, whose amendment covers only the Go patch.
6. The govulncheck fixture uses synthetic advisories on standard-library symbols instead of a `replace`d local module,
   because the policy treats `replace` directives as unsupported records.
7. B-SECRET-SOURCE adds a gitlink case to S6 (the Codex watch item) and an RE2-only case to S7 (M1).
8. The documented readiness invocation builds once and runs the compiled command (`go run` hides the exit code).
9. vulnpolicy requires development-only identities for npm and PyPI exceptions (M2; strengthening, not widening).
10. L6 and L7 as recorded above.
11. The D0 `govulncheck -version` request, disclosed above.

## Unresolved issues, blockers and history

- No blocker. Hosted MR fast (first hosted B-DEPSCAN timing) and the staging gate remain after Codex review.
- The braces exception expires 2026-10-29: after that date B-DEPSCAN fails until a fixed release or a renewed explicit
  owner disposition; it is never renewed automatically.
- Escalations raised and resolved: the D0 contract mismatch and real blocking findings (handed back; D1 to D4).
- Correction cycles: none yet.

## Execution Budget

Budget class: B5
Maximum goal-level attempts: 5 (checkpoint after attempt 3)
Attempts used: 1 (active until Codex review)

### Attempt Log

1. Started at D0 on 2026-10-15 (release `c59c8d8`). D0 found the layout mismatch and blocking findings; escalated and
   resolved by D1 to D4 while independent work continued; implementation of R1 to R6, focused batches and their fixes
   (the b1-01 canary leftover), the read-only security review and its fixes, and the final batch. Inner failed-batch
   loops shared the attempt. Submitted review-ready.

## Suggested follow-up (not authorized scope)

- Verified TLS for core and migrate to PostgreSQL and for core to Redis (F-1), an `https` check for
  `CORE_INTERNAL_URL` (F-2), argv-free minio-init credentials (F-4), an operator deprovisioning entry point (F-6),
  image and supply-chain scanning (F-8).
- Re-evaluate the braces exception before 2026-10-29.

## Binding preparation record (historical)

The initial Result as committed at `023d394` and imported by Codex, kept for provenance; its headings are demoted one
level. Statements in it describe the binding-prep state and are superseded by the sections above.

### Status

**Binding prepared; execution not started; not accepted.** This record is the Git-first execution binding prepared
for Codex's confirmation (Plan r1 §0, §12 sequence step 1). It is not implementation readiness, not review-ready and
not acceptance. No D0, install, scan, build, dependency, harness, Docker, probe or application/test/config/CI command
has run. The security reviewer has not been dispatched. No MR has been opened. Implementation (D0 onward) starts only
after Codex confirms this binding and the input/resource release is relayed.

### Execution binding (Git-first)

Hash lines below deliberately carry neutral labels; the artifact paths are listed under "Bound paths".

| Label | Bound bytes |
|---|---|
| Base staging (origin/staging at fetch, 2026-10-15) | `4c240b1133b40c4d77c707f2aaf52012183adbbf` |
| Plan r1 (P1), accepted source `2b19576` | `cd90a79099bf320fadd9aad39ef4501d1971fb339d6053a533efe69f7f70d51a` |
| Plan r0 archive (P2) | `21d73dff2e62a1e5315d22f94a09670eb0342743bb1205e962e01a2a3808c82f` |
| Plan r1 Codex review, ACCEPT (R1) | `ecc8d24d2a640b88c0fdacada1d3f7d8d12c5e6e4bca6b96ca8c12edeadcac2b` |
| Plan r0 Codex review, historical CORRECTION LOOP (R0) | `bf2fb1f4a10ced0ebcc2713ac621d5842eedff6d6bd81cafd07d83fa95ec7620` |
| Goal (G) | `fb6becf9ebfaeeffb3d2b5833f0894a8dbb9c649d5be103d924710be45ce4174` |
| Spec with the privacy inheritance amendment (S) | `a48d865b3bc462b86785406c6991f4e35b323a73d21af3d8079cf49ec6e9c493` |
| Specification review (SR) | `e7a13253ea081e23b32c4670f59e9e0bcc27f5412e9a46c1cf221abb5c7e41d1` |
| G011 implementation review with final ACCEPT (F1) | `481afdb96a5ff91152f238233967b9026650d04311667aa2996cfb6b0bd73219` |
| G011 final Result (F2) | `e62069daa0107d82c08bf0f15f63bba47a5e0d8a39c7f3382689e1e0e3b36572` |
| G011 closeout receipt (F3) | `95e0cd46ef819e2a45c10435f376e837d052e2d828ee2578cc0b40f41940ce02` |
| G011 merge / merged tree | `1af469634225e6e6527aca485b40aa55953227a8` / `c312478966715a729d1e5431dd567f8c275b3cb6` |
| AUDIT.md (A) | `33a9bac93b867342e81a949681e742b90406d06cdf5471f9e495e4c889c44760` |
| B-SCOPE allow-list (L), 28 lines, 15 patterns | `19134a9a89caab0c8588b71a7edec4fe3e968044695432d60e987df07d2a491d` |

**Bound paths** (bound by commit `4c240b1` and path; AGENTS.md Git-first provenance, no per-file inventory):
- P1 `management/v0/plans/G012-secrets-secure-engineering-plan.md`;
  P2 `management/v0/plans/G012-secrets-secure-engineering-plan-r0.md`.
- R1 G012-plan-r1-codex-review.md (not included); R0 G012-plan-r0-codex-review.md (not included).
- G `management/v0/goals/G012-secrets-secure-engineering.md`;
  S `management/v0/specs/G012-secrets-secure-engineering-spec.md`;
  SR G012-specification-codex-review.md (not included).
- F1 [G011-implementation-codex-review.md](G011-implementation-codex-review.md);
  F2 [G011-durable-audit-events-result.md](G011-durable-audit-events-result.md);
  F3 G011-closeout-receipt.md (not included); A `docs/foundation/AUDIT.md`.
- L G012-scope-allowlist.txt (not included).
- Governance: `AGENTS.md`, `CLAUDE.md`, `management/BUDGET_POLICY.md`, `docs/ARCHITECTURE.md`, `docs/DECISIONS.md`,
  ADR-002/003/004/006, `docs/PERFORMANCE.md`, `docs/foundation/*`. Implementation sources and harness: the tree at
  `4c240b1`.

Every SHA-256 above was computed at `4c240b1` and equals the value cited by the Plan header and the r1 review. The
staging Plan bytes equal the accepted source (`git diff 2b19576 4c240b1 -- management/v0/plans/` is empty).
`1af4696` and `bf462f6` are ancestors of `4c240b1`.

**Delta since the Plan's base `bf462f6`.** Exactly one commit, `4c240b1` ("Accept G012 plan r1 and import reviewed
plan history"): the Plan r1 and r0 archive imports, the r1 review, SUPERVISOR-HANDOFF and one PROGRESS line. No core,
infra, test, contract or foundation-doc byte changed. It is not a G012 design input change. A-G011-1..4 hold as
written (the final G011 tree and review are the ones the Plan binds).

**Allow-list derivation.** The 15 EREs are copied exactly from Plan r1 §11.7. The block from `# Create` to the end of
the file is byte-identical to the Plan's fenced block (`diff` empty). The §11.7 conditional entries are not listed:
`core/go.mod`, `core/go.sum`, `web/package.json`, `web/package-lock.json`, `intelligence/pyproject.toml`,
`intelligence/uv.lock`, the matching B-DEPS edits in `infra/validation/lib.sh` and register rows, and
`infra/ci/{lib.sh,profile-paths.tsv}` for the fast-budget fallback. Each needs a Codex-reviewed amendment. Read-only
self-checks over the comment-stripped patterns:
- every pattern compiles (`python3 -I` `re.compile`; `grep -E -f` exit status not 2);
- `grep -E -f` matched all 29 concrete planned paths of Plan §8/§11.7 (example testdata paths included);
- it matched none of 24 representative excluded paths: the six conditional lockfile/manifest paths,
  `infra/validation/lib.sh`, `infra/ci/{lib.sh,profile-paths.tsv,validate.sh}`, `core/cmd/server/main.go`,
  `core/internal/config/{identity_config.go,config_test.go}`, `infra/security/gitleaks.toml`, `infra/compose/local.yml`,
  `.env.example`, `.gitlab-ci.yml`, GETTING-STARTED, AUDIT, PROGRESS, SUPERVISOR-HANDOFF, the Plan, an
  out-of-set testdata directory and an upper-case `vulnpolicy` file.

This is not a harness run. Any further path needs a Codex-reviewed allow-list amendment.

**Current-tree facts at binding (read-only).** None of the nine create paths exists; all eleven modify paths exist;
`docs/SECURITY.md` is empty (0 bytes). The Git index has one link, `.agents/skills/vestive-design` (mode 120000), and
no gitlink (mode 160000).

**Count-pin inventory at binding (Plan §11.3; read-only).** `required-checks.tsv` has 107 rows (lint 8, lint-base 3,
ops 12). In `infra/ci/validate-controls.sh`, the cited lines carry the expected values: 948/952 (fast 33), 1070/1079
(affected web 19), 1112/1115 (remainder 72), 1206 (VU1 33/72/101), 1207 (VU9 33/42/71), 1937 (19), 1948 (72).
OPERATIONS.md lines 90/92/93/99/100/120/121 carry 33, 71, 100, "eight", the three lint-base checks, 33/72/101 and
33/42/71. The Plan's `grep -nE` term pass over `validate-controls.sh` and OPERATIONS.md returns 66 hit lines; each
change is recorded here at implementation.

**Harness binding.** Plan §11.1 standard form: `bash infra/validation/run.sh --base <target> --allow-file
management/v0/results/G012-scope-allowlist.txt [--expect-hash docs/SECURITY.md=<sha>] --report-dir <EVID>/<batch>
<suites>`. Batch evidence goes under `<home>/.local/state/platform-evidence/g012-impl-4c240b1/<batch>`. During
validation this Result and the allow-list stay at their base bytes (B-PROTECTED); `docs/SECURITY.md` uses
`--expect-hash`.

### Binding conditions (from the r1 review)

1. **Download-only advisories.** Only the four fixed public advisory objects of Plan §5.1/§5.2 are fetched; scans run
   with network none. No online OSV query, no dependency-graph, source, report or `.env` upload, and no fallback
   upload. Existing Go proxy/checksum traffic keeps its documented scope.
2. **Fast-job cost.** About 270 MB of public data per fast run and a forecast +3–6 min are a design estimate, not a
   measured SLA, against the **unchanged 20-minute fast cap**. D0 and b1 record actual cost.
3. **Mismatch returns to Codex.** A D0/b1 budget overrun or a tool-contract mismatch (flag, exit code, schema, offline
   behavior) stops affected work and returns to Codex for a narrow placement/scope disposition. No pin change, relaxed
   failure, cache or advisory-distribution platform, runner migration, benchmark campaign, optional failing scanner or
   stale-data pass.
4. **Security dispositions.** Blocking findings are corrected within reviewed scope or receive an explicit reviewed
   disposition; lockfile remediation needs an allow-list amendment (Plan §5.5).

### Implementation watch item (r1 review, item 2)

The submitted-set listing must identify Git index mode **160000 (gitlink)** before any filesystem directory skip, so
a submodule entry is refused (exit 2) as the Plan already requires and is never counted as a skipped untracked
directory. This is implementation guidance for the approved guarantee, not a Plan revision; B-SECRET-SOURCE evidence
records how it is met.

### Ownership

| Item | Value |
|---|---|
| Task / role | G012 / impl (the one implementation owner of all files, Plan §12) |
| Agent id | `agent-<id>` |
| Path | `<source-checkout>` (native name under the standing owner fallback, AGENTS.md "Task-prefixed Claude worktrees"; no worktree created or switched, no rename) |
| Repository | `git@gitlab.example:<namespace>/<project>.git` (SSH; no embedded authentication) |
| Branch | `g012-secrets-secure-engineering`, created from the base staging commit above (upstream to staging unset before the first push) |
| Worktree lock at dispatch | `locked claude agent agent-<id> (pid 68375 start 10178190)` |
| Effort | high as instructed; actual runtime setting unknown (not verifiable from inside the worker) |
| File ownership | every path in the allow-list; no delegate edits files |
| Dispatch record | `<home>/.local/state/platform-evidence/g012-impl-4c240b1/dispatch-record-impl.md` (outside Git) |
| Coordinator | `0-agent-5c6f436` (dispatch intent in the same directory) |

The read-only security reviewer (Plan §12, after b2) is not yet dispatched. Its task/role/agent/path/effort will be
recorded here when it is.

### Execution Budget

Budget class: B5 (Plan §13; approved in the r1 review)
Maximum goal-level attempts: 5, checkpoint after attempt 3
Attempts used: **0** (binding prep is management records only and is not an attempt)

### Attempt Log

1. Not started (begins with D0 after the binding is confirmed and released).

### Open gates (execution held)

1. **Codex confirmation of this binding**: base, allow-list, this Result and the ownership record.
2. **Input/resource release** relayed by the coordinator. Before D0, the target is refetched and the exact-delta rule
   applied.

Then the per-batch gates of Plan §11.1 apply: fetch and exact-delta (a management-only delta is still a secret-scan
input), `bash infra/ci/docker-desktop-mounts.sh`, the harness flock owned by the coordinator, and target/candidate SHAs.

### Implementation summary, files changed, migrations, contracts, tests

None yet. Only this Result and the allow-list are committed. Planned: no migration, schema, role, grant, audit event,
public API or MCP contract change (Plan §9).

### Commands and validation

Read-only only:
- `git fetch origin`; `git checkout -b g012-secrets-secure-engineering origin/staging`; `git branch --unset-upstream`;
- `git merge-base --is-ancestor` (`4c240b1`, `1af4696`, `bf462f6` against HEAD);
- `git worktree list --porcelain`; `git log` and `git diff --stat` (`bf462f6..origin/staging`; `2b19576 4c240b1` on
  plans); `git ls-files -s` (link and gitlink modes);
- `sha256sum` of the bound inputs and the allow-list; `awk`/`diff` of the allow-list block against Plan §11.7;
  `grep -E -f` and `python3 -I` regex self-checks;
- `grep`/`awk` reads of `required-checks.tsv`, `validate-controls.sh` and OPERATIONS.md.

No D0, scan, test, build, harness, Docker or probe command ran.

### Acceptance criteria

Mapping: Plan §10.

| AC | Status |
|---|---|
| AC1 inventory and provider substitution | pending |
| AC2 fail-closed acquisition without disclosure | pending |
| AC3 disposable rotation and withdrawal | pending |
| AC4 least privilege and deprovisioning intact | pending |
| AC5 submitted-source scan fails closed | pending |
| AC6 enforceable dependency checks for three graphs | pending |
| AC7 encryption readiness rejects unsafe posture | pending |
| AC8 vendor/security records complete | pending |
| AC9 attributable, bounded CI evidence | pending |
| AC10 review and closeout | pending |

### Architecture deviations

None.

### Unresolved issues and blockers

- Execution is blocked only by the open gates above. No technical blocker was found in binding prep.

### Suggested follow-up

None.

## Correction 1 (attempt 2): fail-closed parser fixes

Codex returned CORRECTION LOOP on attempt 1 ([implementation review](G012-implementation-codex-review.md) on staging
`3a852ba`, handoff beafe7eb): two bounded fail-closed defects in `vulnpolicy`. This is correction pass 1, Goal-level
attempt 2 of B5 (checkpoint after attempt 3); no scope, Plan, threshold or owner change. The sections above are the
attempt 1 record and stay as submitted.

**Reconciliation.** Fetched and rebased onto `3a852ba` (management only: the review, the byte-exact import of the
attempt 1 Result, PROGRESS and SUPERVISOR-HANDOFF); the attempt 1 Result commit dropped as already upstream. The README
secret-scan line on the rebased tree found no leaks (`c2-pre-scan`).

**Fixes** (commit `67ef90c`; the only non-management change since the attempt 1 candidate is these four files in
`infra/validation/tools/cmd/vulnpolicy/`, 107 insertions and 2 deletions):

| Blocker | Fix | New negatives (`policy_test.go`) |
|---|---|---|
| 1 Line-scanner truncation: the exceptions, go.mod and uv.lock readers returned a parsed prefix without checking the scanner error | Each reader returns an input error on any `bufio.Scanner` failure (for example an over-limit line) before any parsed data is used; the CLI exits 2 | `TestScannerErrorsRejectParsedPrefix`: exceptions with a valid row, an over-limit line and a hidden expired row; go.mod with an over-limit line before a hidden requirement; uv.lock with an over-limit line before a hidden unsupported package; each refused by the parser and by the CLI with exit 2 |
| 2 OSV references checked one way only; `Decoder.More` is not an EOF check | Every group ID must resolve to a non-empty advisory record of the same package, every record needs a non-empty ID, and the reverse check stays; after the one result document a second `Decode` must return `io.EOF` | `TestOSVReferencesAndSingleDocument`: a group ID without its record, an empty advisory ID, a trailing closing brace, a trailing closing bracket and a second document (all refused; incomplete inputs evaluate to exit 2, the trailing bracket also through the CLI); trailing whitespace stays valid |

All earlier clean, finding and exception cases keep their outcomes (the existing test groups, the B-DEPSCAN fixtures
and the current tree, which still reports braces as EXCEPTED).

**Focused evidence** (wrapper as before: mounts 10 %, flock, candidate and target recorded):

| Batch | Candidate / target | Commands | Outcome |
|---|---|---|---|
| c2-vtools | correction work tree on the rebased branch | validation-tools check stage (gofmt, vet, `go test ./...`) | rc 0; vulnpolicy 14 test groups PASS, including the two new ones |
| c2-pre-scan | the correction commit (below) | README scan line | rc 0, no leaks |
| c2-focus | commit `67ef90c` over target `3a852ba` | `run.sh --base 3a852ba… --expect-hash docs/SECURITY.md=e6bb1a70… lint lint-base unit contracts provenance` | **PASSED** rc 0 in 6 min: 29 PASS, 0 FAIL; leftover inventory empty |

c2-focus results file `67344c9b…743e`, run log `4e59449a…4743`. Suite choice: `lint-base` owns B-DEPSCAN (vulnpolicy
built from source and run over the fixtures and the current tree); `contracts` owns B-CONTRACT-TOOLS (the validation-tools
check stage, where the vulnpolicy units run); `lint` holds B-SHELL and B-SECRET-SOURCE and is needed for
B-SECRET-CANARY; `unit` runs the README check block, including the submitted-source scan over the rebased tree with its
management delta; `provenance` holds B-RENDER-*, B-PROTECTED (550 files, one expected hash) and B-SKILL-UNCHANGED; the
final checks give B-DOCKER-ONLY, B-SECRET-CANARY (24 generated secrets), B-SCOPE (54 paths since the target) and
B-REQUIRED-CHECKS (28 others).

**Reuse mapping for the 104 required IDs** (the correction changes only the validation-tools module, whose only
consumers are B-DEPSCAN and B-CONTRACT-TOOLS; no runtime, browser, migration, rotation, toggle or CI-control input
changed):
- From c2-focus (`67ef90c` / `3a852ba`), 29 IDs: B-SHELL, B-HADOLINT, B-COMPOSE, B-ENV-EXAMPLE, B-DEPS, B-MIGRATE-SET,
  B-DOCKERIGNORE, B-CI-CONFIG, B-SECRET-SOURCE, B-LITERALS, B-NAMING, B-IDENTITY-GUARDS, B-DEPSCAN, B-UNIT-CHECKS,
  B-CONTRACT-TOOLS, B-CONTRACT-GO, B-CONTRACT-PY, B-CONTRACT-AGREE, B-CONTRACT-PEER, B-CONTRACT-NEGATIVE, B-RENDER-ID,
  B-RENDER-PLAN, B-RENDER-REPEAT, B-PROTECTED, B-SKILL-UNCHANGED, B-DOCKER-ONLY, B-SECRET-CANARY, B-SCOPE,
  B-REQUIRED-CHECKS.
- Reused from final-01 (`c9f874e` / `34e0362`, results `b03fd702…53cb`), 75 IDs: the `lint-web` (5) and `lint-core`
  (B-MCP-SOURCE) guards, `controls` (B-DEPS-CONTROLS, B-WRAPPER-CONTROLS), `ci-controls` (B-CI-CONTROLS, 441 controls),
  `root` (17, including B-CHECKS-ROOT and the browser suite), `migrations` (27, including the identity, tenancy,
  policy, RLS, object and audit catalog, migration and integration checks), `ops` (13, including B-SECRET-ROTATION and
  the identity checks), `rename` (5) and `off` (4, including B-TOGGLE-OFF-ADDITIONS).

**Execution Budget (updated)**

Budget class: B5
Maximum goal-level attempts: 5 (checkpoint after attempt 3)
Attempts used: 2 (second active until Codex re-review)

1. Attempt 1: as recorded above; submitted review-ready; Codex CORRECTION LOOP (two parser defects).
2. Attempt 2 (correction pass 1): the two fixes, their regressions, the focused batch and this record; submitted for
   automatic re-review.

Status after correction 1: review-ready again, not accepted; no MR. Codex re-review, then MR fast, Duo, the staging
gate and cleanup remain. The braces exception (owner approval, 2026-10-29 expiry) is unchanged.

## Correction 2 (attempt 3): hosted cidfile collision

**Trigger.** After Codex accepted correction 1 for integration, MR !28's automatic fast job failed on hosted CI (job
17189970959, head `2758235` over base `a8eb787`). B-DEPSCAN failed with "fetch PASS; phase-a PASS; vulnpolicy-build PASS;
fixtures-govulncheck FAIL; fixtures-osv FAIL; current PASS", followed by the consequential B-REQUIRED-CHECKS failure;
every other fast check passed (suite `lint-base` took 86 s on hosted; the whole fast harness 428 s of 1162 s). The
hosted per-evaluation outputs were not published, and this record does not claim them. Codex recorded a hosted hold and
then authorized this correction as attempt 3 of B5 (review sections "Hosted MR qualification hold" and "Hosted
cidfile diagnosis and correction authorization" on the current staging head).

**Cause.** Hosted `infra/ci/validate.sh` runs the harness with `--owned-registry`, so `static_depscan` passes
`--cid-dir` to `dependency-scan.sh`. `drun` named each container's ID file `g012-<name>-<tag>.cid`, which is the
same for every run of one step name, and `evaluate` runs three times per fixture step (none, valid, expired). Docker
refused the second and third runs with exit 125 ("container ID file found"); `current` evaluates once and passed. My
local batches never passed `--owned-registry`, so the hosted integration path was not exercised locally.

**Reproduction (before the fix).** `dependency-scan.sh … --cid-dir <EVID>/repro-cid/cid scan` on the clean tree at
`2758235` (fresh volumes, empty `--out`) gave exactly the hosted step results; the failing assertions were
"exceptions go-valid: vulnpolicy exit runner-125, want 0; exceptions go-expired: vulnpolicy exit runner-125, want 1;
…" and the same for osv-valid and osv-expired; the first (none) evaluation of each step passed. Results file
`067cb109…6508`, run log `fd2300d6…f024`.

**Fixes** (commit `d7ab607`; two files, both inside the 19-pattern allow-list):
1. `infra/security/dependency-scan.sh` `drun`: one distinct, initially absent ID file per actual invocation, named with
   the invoking subshell's PID and a random suffix (the `lib.sh` `docker_run_owned` pattern), so it also holds inside
   command substitutions and pipelines. A missing suffix or an existing file fails closed (exit 125, never a pass).
   Earlier ID files stay in the registry directory for the owned cleanup; ownership rules are unchanged.
2. `infra/validation/run.sh` `static_depscan`: a failing B-DEPSCAN summary now carries the failing steps' own
   assertion text (step names, expectations and exit codes; never a value) through `depscan_fail_detail`, with control
   characters removed and capped at 800 characters. The FAIL status, exit handling and every scan are unchanged.

**Validation** (wrapper as before: mounts 10 %, flock, candidate and target recorded; reconciled onto `964dd81`,
management only):

| Batch | Candidate / target | Commands | Outcome |
|---|---|---|---|
| c3-pre-scan | the correction commit over the target | README scan line on the rebased tree, plus the redacted gitleaks diagnostic | no leaks, 0 findings |
| c3-cid-run | the correction commit | `dependency-scan.sh … --cid-dir <fresh dir> scan` (fresh volumes, empty `--out`) | rc 0: all six steps PASS; fixture exceptions none/valid/expired give 1/0/1 for both families; `current` PASS with braces EXCEPTED; 18 ID files, 18 distinct, none empty, no "container ID file found" message; no owned container or volume left |
| fail-detail-demo | run.sh function text | `depscan_fail_detail` applied to the failing reproduction results, an over-long detail with control characters, an all-PASS file and a missing file | failing detail kept (487-character message naming both failing steps and their exit codes); cap holds (819 characters with the prefix, 0 control characters); empty for PASS and missing input |
| c3-hosted-path | the correction commit over the target | registry made as `validate.sh` `make_registry` does (mode 700, `cid/`, token, header/job/token/repo rows, Docker start inventory), then `run.sh --owned-registry … --secrets-file … lint lint-base`, then `cleanup-owned.sh --mode finalize` | **PASSED** rc 0 in 4 min: 17 required IDs PASS (B-SHELL over both edited scripts, B-SECRET-SOURCE, B-DEPSCAN on the hosted `--cid-dir` path, B-SECRET-CANARY, B-SCOPE, B-REQUIRED-CHECKS); 18 distinct ID files; finalize cleanup exit 0 with an empty inventory; no owned container or volume left |

Evidence hashes (local, purged at closeout): c3-cid-run log `71ecd544…f315`, results `0baf00cb…d553`;
fail-detail-demo log `9aaa834b…c03b`; c3-hosted-path log `4d6b54dd…f175`, results `4de2a096…e53b`.

**Reuse mapping for the 104 required IDs** (the correction changes only `dependency-scan.sh`, whose only consumer is
B-DEPSCAN, and one failure message in `run.sh`):
- From c3-hosted-path, 17 IDs: B-SHELL, B-HADOLINT, B-COMPOSE, B-ENV-EXAMPLE, B-DEPS, B-MIGRATE-SET, B-DOCKERIGNORE,
  B-CI-CONFIG, B-SECRET-SOURCE, B-LITERALS, B-NAMING, B-IDENTITY-GUARDS, B-DEPSCAN, B-DOCKER-ONLY, B-SECRET-CANARY,
  B-SCOPE, B-REQUIRED-CHECKS.
- From c2-focus (correction 1 over `3a852ba`), 12 IDs: B-UNIT-CHECKS, B-CONTRACT-TOOLS, B-CONTRACT-GO, B-CONTRACT-PY,
  B-CONTRACT-AGREE, B-CONTRACT-PEER, B-CONTRACT-NEGATIVE, B-RENDER-ID, B-RENDER-PLAN, B-RENDER-REPEAT, B-PROTECTED,
  B-SKILL-UNCHANGED (their inputs are unchanged by this correction; the management delta since then was scanned by
  c3-pre-scan).
- From final-01, 75 IDs, as mapped in correction 1.
- Hosted proof: a new MR fast run after Codex's bounded review.

**Execution Budget (updated): B5 checkpoint after attempt 3**

Budget class: B5
Maximum goal-level attempts: 5 (checkpoint after attempt 3)
Attempts used: 3 (third active until Codex review)

1. Attempt 1: implementation through review-ready; Codex CORRECTION LOOP (two parser defects).
2. Attempt 2 (correction pass 1): the parser fixes; Codex accepted source and local evidence for integration.
3. Attempt 3 (correction pass 2): the hosted cidfile collision found by MR fast, its reproduction, the two fixes and
   the focused validation above.

**Checkpoint record.** Progress: every Plan r1 requirement is implemented; the source fixes from both corrections are
validated locally, including the hosted `--owned-registry` path that was previously untested. Remaining risk: the
hosted environment (fresh daemon, network to the four advisory objects and the Go module proxy) is proven only by the
next MR fast run; the earlier hosted run already showed the download and timing path working (86 s for `lint-base`).
Forecast: no further source change expected; attempts 4 and 5 stay in reserve for a hosted-only defect. Any material
new failure goes back to Codex under the existing budget.

Status after correction 2: review-ready, not accepted. The correction commits are not on the MR branch yet (see the
hand-back); no MR fast was started, auto-merge stays off, and the braces exception (owner approval, 2026-10-29 expiry)
is unchanged.

## Integration and hosted staging gate

Facts verified by the coordinator from the GitLab API and the job artifacts (2026-10-15).

- **Integration.** Codex gave conditional integration ACCEPT (handoff 6692422e) after accepting correction 2 (attempt
  3). The coordinator armed auto-merge on the exact head; MR !28 merged at `9c2798612e4916771e45aa69238c9c5cf185aa00`
  into merge commit `5e1e98735f6224888889d516ef2397a0ea8dc89a`, whose tree `8cbb5d2e74c395fe6d2212f42f23429603606a45`
  equals the reviewed head's tree. The remote source branch was deleted; the temporary review branch was deleted after
  its content was verified on the MR branch.
- **MR fast jobs:**
  - `17189970959` on `2758235`: **FAILED** (B-DEPSCAN fixture steps, the cidfile collision recorded in "Correction 2",
    and the consequential B-REQUIRED-CHECKS);
  - `17190291257` on `4305d11`: success, 35 PASS, B-DEPSCAN with all six steps, suite `lint-base` 104 s on hosted;
  - final MR pipeline `2949481827` on `9c27986`: success (GitLab merged on success).
- **GitLab Duo:** review sessions 9507542 and 9508253 only started; a late note (4017884656) reported "Something went
  wrong and the review request was stopped". Codex disposition: Duo unavailable for this MR, neither approved nor clean.
  The coordinator recorded this in MR note 4017854576.
- **Staging pipeline `2949510391`: success.**
  - fast job `17191203943`: success;
  - gate job `17191203944`: success, 06:19:02 to 06:56:35 UTC (2299.3 s), profile gate; executed suites controls,
    ci-controls, root, migrations, ops, rename, off; head `5e1e987`, tree `8cbb5d2e…`, comparison base `797b190`;
    required count 73, stragglers 0, harness stop none;
  - results: 73 PASS, 0 FAIL, 6 INFO; B-CI-CONTROLS 441/441; B-SECRET-ROTATION PASS; B-AUDIT-IT PASS; 40 E-RLS trigger
    lines (the baseline);
  - the owner status rule applies: success qualifies the CI acceptance gate.
- **Gate artifact digests** (SHA-256, neutral labels):

| Artifact | SHA-256 |
|---|---|
| results.tsv | `4782248930a503dfd0a941bbb672663cc5b557e7fe3691df2f5f0191516e4240` |
| summary.json | `db4ac91601e3cdad7a2f726b1269bfea41c259c96cdb2263523b74e45dd77ac3` |
| ci-controls.log | `ea307fe0e896144e721b1874aea18e82447fa81a78fcf008c7f68e1214ee396e` |
| required-checks.tsv | `321135abe74ec8e4a109522743b09ae9968d76791148df37c69545ddd80815b8` |

## Evidence retention

**Necessary proof holds: none.** This Result, the Codex review records and the hosted job artifacts carry the proof.
The local folders are temporary and eligible for purge at closeout after a compact receipt:
- `<home>/.local/state/platform-evidence/g012-impl-4c240b1/` (implementation batches, D0, the braces evidence
  runs, the S-review report, the correction reproductions and batch log); its `readme.env` holds generated local
  values and is purged with the folder;
- `<home>/.local/state/platform-evidence/g012-plan-795b95a/` (planning).

**Open external expiry:** the braces 3.0.3 exception (GHSA-vfj7-8cjw-p6xm, owner-approved) stays in force until
2026-10-29 inclusive; after that date B-DEPSCAN fails by design until a fixed release or a renewed explicit owner
disposition, and production inclusion revokes it earlier.

## Closeout owed

- Codex final ACCEPT, the PROGRESS update and the G013 handoff (Codex).
- The coordinator's bounded cleanup of the resources this implementer created:
  - the worktree `<source-checkout>`;
  - its local branches: the closeout branch `g012-closeout`, the former feature branch of this Goal (same name as the
    deleted remote MR branch), and the native worktree branch
    `worktree-agent-<id>`; no local review branch exists (the review branch was only pushed);
  - images pulled or built: `ghcr.io/google/osv-scanner@sha256:afd838850ac1a0fcc15ff4a041dc9ba11123c3f0d2666217a5f0fcf9222b55fa`
    (pulled for D0), `golang:1.27.2-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d`
    (pulled for the Go patch), and `beginify-core-check:latest` (the README project's check image, rebuilt by the
    documented readiness commands; verify its current ID before removal, since another local run may have rebuilt it);
    every other image I built (`g012-impl-vtools-check`, `g012-impl-web-runtime`, `g012-impl-web-deps`,
    `g012-impl-web-src`) was removed by its script;
  - containers and volumes: none remain; every container and volume I created carried `g012.owner=impl` (with
    `g012.run=<run>` or `g012.depscan=<tag>`) or a harness `g001.run=<id>` label and was removed by its script or the
    harness (volumes `g012-advisory-*`, `g012-gocache-*`, `g012-impl-adv-*`, `g012-impl-goc-*`);
  - the two evidence folders above, and the local registry folder `c3-reg` inside the implementation folder;
  - `/tmp` files from binding prep: `/tmp/g012-plan-block.txt`, `/tmp/g012-patterns.txt`, `/tmp/g012-pos.txt`,
    `/tmp/g012-neg.txt` (harness and scan work directories under `/tmp` were removed by their traps).
- G021 stays closed; the G014 Goal/Spec human-input stop remains.
- The B5 ledger stays at 3 used.
