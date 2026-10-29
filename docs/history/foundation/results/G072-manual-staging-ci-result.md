> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G072 — Manual staging and preview CI start/deploy and stop: Result (source slice, attempt 1 of B3)

Date: 2026-10-28 UTC. Author: Claude coordinator (Sonnet 5.5) with four Opus 5.5 high writers. **Source implementation and the
focused offline proofs only.** No GitLab setting, WIF/IAM, GCP project, DNS, credential or live staging/preview operation was
performed or released (L-G0 … L-H2 stay separately gated; the G057 hold applies to every staging item). Not self-accepted.

## 1. Inputs and rebinding

| Item | Value |
|---|---|
| Accepted Plan r1 | `management/v0/plans/G072-manual-staging-ci-plan.md` sha256 `ef530b36b81d8575d067123369e7dd33889c23292a14211d5d96e4bb6dbcf509` (Codex ACCEPT, release `G072-plan-r1-codex-review-20261028.md`) |
| Spec r2 / Goal | sha256 `bbcbbc5e516686ea95ae43ae5659efd0f1707dad22b809544492df84e4aef726` / `d42f40abdd87ba07d6340e08339f8cd4df3f887a21a5ec632f4fecf1dcda311b` |
| Implementation base | staging `1523442dbd4dae49b09064914f6d97ae67508cf9` (already carries the accepted Plan bytes) |
| **Safe rebind recorded** | Plan baseline `0c66332d` → `1523442d`: the only non-management delta is `.codex/config.toml` (Luna `xhigh` preserved). Later reconciliation to staging `f5465109`: management-only (G057 Spec/Goal, PROGRESS, ROADMAP); no bound G072/G070 source changed. |
| Tested source | `g072-impl` `a4246d01ffd698d4d0e3d919930a7a283e77a510` (36 files, +7686/−54 vs staging `f5465109`) |

Writers (each in an isolated worktree from `g072-impl`, exclusive paths, no overlap verified before merging):
W1 `g072-w1` `bd4ecd74` (federation/IAM bootstrap + test), W2 `g072-w2` `290c7c91` (lease/fence/metadata library, adapter, tests),
W3 `g072-w3` `1ca4e873` (`.gitlab-ci.yml`, launcher, publication filter, CI context image, `run.sh` B-CI-CONFIG/B-SHELL),
W4 `g072-w4` `0198f4cf` (docs, hostile-build proof, `check.sh` rows). Merged with `--no-ff`; model Opus 5.5 high.

## 2. What was built

- **Lease, fence, metadata (W2):** `infra/staging/meta-lib.sh`; additive `staging.sh`/`lib.sh` changes (lease on provision/deploy/start/stop/reset/retire,
  `lease --show|--release|--break`, schema fence written before upload in `deploy_candidate`, `build --source`, `GCLOUD_IMPERSONATE_SA`).
  Legacy G070 local semantics preserved (D12); dry-run parity proven.
- **CI adapter (W2):** `ci.sh`, `ci-lib.sh`, `ci-environments.txt` (`staging`, `preview-1`), `examples/ci-settings.example`; actions prepare/check/build/deploy/stop and the
  local `lease`, `meta`, `lineage init|reconcile`.
- **Federation (W1):** `infra/terraform/bootstrap/ci-federation.sh` (`--dry-run`, `--verify`, `--teardown`), `ci-meta-permissions.tsv`, README section.
- **Routing/launcher (W3):** `.gitlab-ci.yml` typed inputs (no `images`), jobs `staging:prepare/check/build/deploy/stop`; `infra/ci/staging-action.sh`;
  `infra/docker/staging-ci.Dockerfile`; `infra/staging/ci-publish.sh` (`publish_filter`, `canary_scan`, `summary_render`); `run.sh` B-CI-CONFIG
  (two-document parse + 32 mutation controls), B-SHELL list, B-HADOLINT.
- **Docs/proofs (W4):** `STAGING.md` "CI start/deploy and stop", `OPERATIONS.md`; `tests/ci-build-boundary.sh` + fixtures; `check.sh` rows
  CI-FEDERATION, CI-ADAPTER, ENV-LEASE, CI-LAUNCHER, `--cases ROW=SEL`, LIFECYCLE-DRYRUN lease/§10.9 cases, STAGING-GREP raw-log/lease rules,
  SHELLCHECK-DIRECT list = B-SHELL staging list.
- **Coordinator:** `staging-ci.Dockerfile` google-cloud-cli pinned by digest. Image `…/google-cloud-cli:524.0.0-slim` (W3's guess) does not exist; pulled
  `590.0.1-slim` once, verified bash 5.2, git 2.47, tar, ssh, ssh-keygen, sha256sum, curl, gcloud, coreutils present (A7), pinned
  `@sha256:2a5368c4f3e25f7e202ccbd7b63e4128380d8fd683266f7f015c5d0afcc8f3be`.

## 3. Selected proofs (executable selectors; evidence in `.local/evidence/g072-a1/`)

| Proof | Selector | Source | Outcome |
|---|---|---|---|
| EP-1b | `bash infra/staging/tests/env-lease.sh --work W --report R`; `bash infra/staging/tests/ci-adapter.sh --only fence --work W --report R` | `fd3c354b` | **PASS** 36/36; **PASS** 9/9 (`ep1b-1/`) |
| EP-1 (A5) | `bash infra/staging/tests/ci-build-boundary.sh --work W --report R` (holds the single build slot) | `fd3c354b` | **PASS** 6/6: fixture, scan-controls, compose-refused (exit 2 before any build), baseline-measure, dockerfile-clean (hostile `migrate.Dockerfile`, no canary in 7 image exports/inspect/history/output), mutant-detected. **Artifacts 371,132,823 bytes ≤ 1024 MiB: A5 does not trigger** (`ep1-1/`) |
| S1 | `bash infra/validation/run.sh --base f5465109… --allow-file <G072 allowlist> --report-dir E/static-3 lint lint-base provenance` | `a4246d01` | **PASS** all rows, incl. B-SHELL, B-HADOLINT, B-CI-CONFIG (32 of 32 mutations refused), B-SCOPE (36 paths in the allow-list), B-SECRET-CANARY, B-REQUIRED-CHECKS (`static-3/`) |
| S2 | `bash infra/staging/check.sh run --work W --report E/offline-2 --edge-image g070a1-edge:p0-r1 --only CI-FEDERATION,CI-ADAPTER,ENV-LEASE,CI-LAUNCHER,LIFECYCLE-DRYRUN,STAGING-GREP,SHELLCHECK-DIRECT` | `71b410f3` | **PASS** 7/7 (`offline-2/`) |
| S4 | `bash infra/ci/validate-controls.sh --only F-push-main,F-push-staging,F-agent-push-staging,F-agent-push-main,RR1,RR2` | `71b410f3` | **PASS** 6 of 6 (530 defined; no full suite) |

Attribution: S2/S4 ran on `71b410f3`; the later commit `a4246d01` only renames a local array in `run.sh`'s B-CI-CONFIG mutation block
(9 lines, no behaviour change, shellcheck of `run.sh`/`lib.sh` clean) and is covered by the S1 rerun on `a4246d01`. S3 (full image build proof beyond EP-1) was not selected: EP-1
covers the changed build boundary; the G070 image-ID preservation is exercised inside EP-1 baseline-measure.
Writer-side results before integration: ci-federation 88/0 (26 mutations killed), ci-adapter 74/74 (7 mutations), env-lease 36/36 (4 mutations), ci-launcher 80/80 (27 mutations).

## 4. Preserved failures (executed and failed; kept distinct)

1. **First S2 run (`c9873d21`)**: SHELLCHECK-DIRECT FAIL (real SC2178/SC2128/SC2179/SC2034/SC2120/SC2318 findings in `check.sh`, `ci-lib.sh`, `ci-adapter.sh`,
   `env-lease.sh`, `ci-build-boundary.sh`, `ci-federation.sh`; writers had no shellcheck on the host) and ENV-LEASE FAIL `dryrun-parity retire`
   (a **test defect**: the base and new dry-runs straddled a wall-clock second, so inventory timestamps and plan names differed; passed standalone by luck).
   Corrected in `71b410f3` (timestamp normalisation, name/array renames, split `local`, file-level SC2034/SC2120 directives where state is shared through
   sourcing/dynamic scope); second run green.
2. **First S1 run (`c9873d21`)**: B-SHELL FAIL and B-REQUIRED-CHECKS FAIL: W3's B-CI-CONFIG `local -a name` made shellcheck read an unrelated `name` at `run.sh` line 2415
   as an array. Renamed `mutname` in `a4246d01`; S1 green.
3. Writer-time: `ci-launcher.sh` tripped the existing no-unformatted-inspect rule (a quoted grep pattern); W4 added a narrow exception rather than edit W3's file.
Repair attempts: one correction pass each (within the single B3 attempt 1).

## 5. AC map (source evidence; live evidence pending)

| AC | Source evidence | Live |
|---|---|---|
| AC1 routing | B-CI-CONFIG (inputs, five jobs' exact keys, 32 mutations), CI-LAUNCHER (80), S4 controls | F-1a (not released) |
| AC2 candidate fidelity | CI-ADAPTER cand/reuse groups, LIFECYCLE-DRYRUN `build --source`, EP-1 (ID preservation) | L-H1 |
| AC3 identity/secrets | CI-FEDERATION (88), ENV-LEASE IAM cycle, CI-LAUNCHER publication, CI-ADAPTER cred group, EP-1 canary (raw/base64/gzip/tar) | F-1a, §10.3 denial probes, L-W1p |
| AC4 start/deploy | CI-ADAPTER life/order/fence groups, EP-1b | L-H1, L-H2 |
| AC5 stop/recovery | ENV-LEASE (both-direction concurrency, expired CI break, break-glass, parity), CI-ADAPTER stop/supersession/rc 255 cases | F-1a/F-1b, L-H1 |
| AC6 closeout | cleanup at final ACCEPT (below) | — |

## 6. Deviations, assumptions and Codex decisions needed

- **W1:** custom role ID `ciStop<env>` renders `ciStoppreview_1` (role IDs allow no hyphens); `roles/iap.tunnelResourceAccessor` is granted at project scope (no per-instance
  gcloud command): **Codex to confirm**; `osAdminLogin` instance binding is lost on re-provision (rerun the script); GitLab namespace ID is owner-supplied.
- **W3:** the launcher's `after_script` cannot release the lease or remove the OS Login key (no adapter entry point for it); it stops the context with a 60 s grace, writes a fallback
  summary and relies on lease expiry/the break rule. The adapter's own EXIT trap releases the lease and key on normal exit paths. Publication cases use a stand-in adapter in W3's test.
- **W2 (restrictive settlements):** every CI deploy is refused until `lineage init`; identity pins all versions or all `none`; jobs keep `SESSION_MAX_SECONDS=7200`; stop under a held local lease
  writes only its stop request; the bucket check confirms only the derived name; receipts stored as `receipts/<id>-<sha12>.tsv` (G070 receipt IDs can collide between jobs). The adapter
  unsets planted `CLOUDFLARE_API_TOKEN`/`STAGING_CI_ID_TOKEN` instead of refusing; build passes every `DOCKER_*` variable (Plan S8).
- **W4:** Plan §10.9 item 5 (`STAGING_SSH_KEY_EXPIRE_AFTER`) not implemented (the Plan allows). `tests/fakes/*` are not linted by either shell list (extend later if wanted). `ci-launcher.sh` prints no `RESULT` line (its row parses its summary line).
  `env-lease.sh --only presence-403` selects 0 cases though the case exists (group selection works).
- **Digest pin:** google-cloud-cli `590.0.1-slim` replaces the nonexistent `524.0.0-slim`.
- No candidate publisher credentials or fallback were added (A5 not triggered).

## 7. Cannot be checked offline (live items verify these)

Real `gcloud` argument forms (`--attribute-mapping` with `==`, `--condition`, `--verify` read formats, custom-role recreation); `gcloud storage` 404/412 wording and the `#generation` URL form; goose lineage query against the real store;
`curl` in the context image at run time; OS Login key TTL; GitLab accepting typed inputs/interpolation, `artifacts:access: none`, the runner credential helper for the prepare fetch, `environment_protected` /
`environment_action` claims; D9 tier/feature availability (L-G0). B-CI-CONFIG ran with the real `yq` image on `a4246d01`; W3's earlier Python stand-in is superseded.

## 8. Estimates versus actuals

| Category | Plan | Actual (known) |
|---|---|---|
| Coding | 23–34 agent-hours; wall 11–15 h | W2 ≈ 54 min wall (≈ 568k tokens), W4 phase 2 ≈ 15 min; W1/W3/W4 phase 1 wall time not captured; all four writers finished the same session |
| Build/test | S1 ≈ 6 min; S2 5–8 min; EP-1b 10–20 min; EP-1 45–75 min | S1 ≈ 6 min/run (3 runs); S2 ≈ 2 min/run (2 runs); EP-1b, EP-1 and S4 not separately timed (EP-1 ran in the background with cached bases); 2 correction reruns |
| Review/queue | unknown | pending |

## 9. Remaining and cleanup

Source review by Codex, MR to `staging` with hosted fast checks and Duo, merge only after gates, then attributable post-merge rows on the frozen merged SHA. Live items L-G0 … L-H2 each need
their own recorded release; the G057 retention hold and Part B are unchanged. Cleanup at final ACCEPT: worktrees `agent-<id>`, `agent-<id>`, `agent-<id>`,
`agent-<id>`, `agent-<id>`, `agent-<id>`, `agent-<id>` and their branches (`g072-plan`, `g072-w1…w4`, `g072-impl` after merge), the pulled
`google-cloud-cli:590.0.1-slim` image if no longer pinned in the merged tree, and `.local/{evidence,secrets,tmp}/g072-a1` after Codex disposition. At the time of this Result no G072-labelled container or image remains.

## 10. Source review round 1: Codex REVISE of `026b27b1` and corrections (append-only; supersedes §6 where it differs)

The hosted pipeline on `026b27b1` (2997259685) passed (`validate:checks` success; `validate:integration` manual, not run) and was preserved: no push to
`g072-impl` was made until it was terminal. Codex found three P1 and one P2 source defects; each was corrected by its writer on a branch, then merged:

| Finding | Correction | New offline evidence |
|---|---|---|
| P1 denial probes failed open (`ci-lib.sh`: every curl failure counted as denial; token command unchecked) | valid non-empty token required; denial only on 403/404 or a 200 granting nothing; transport, 5xx, 401/auth, parse and empty/failed-token outcomes fail closed (exit 6, zero writes, no reuse decision); an impersonation failure must carry a permission denial | 11 cases (incl. granted 200, empty token, fail-open mutation control) |
| P1 stop blocked before provider fallback (`oslogin_key` required; key failure exited) | key generation, key registration and guest stop are best-effort for the stop action; provider stop always attempted; success only after TERMINATED is observed; provider-stop failure exits 4; deploy keeps the key strict | 5 cases (3 failures reaching the fallback, provider failure, control) |
| P1 metadata bucket ownership unchecked (`meta-lib.sh`) | bucket describe returns `projectNumber`; compared with the project number taken from the profile's `CI_WIF_PROVIDER` (created in the environment project against its pinned number; CI identities cannot read resourcemanager), exported as `STAGING_META_PROJECT_NUMBER`; a local operator without it uses `gcloud projects describe` as the deployer, and a failure there refuses (exit 2, tells the operator to set the variable); foreign, missing or malformed number refused with zero writes; a confirmed 404 stays absent | 7 cases + `owner-local` in ENV-LEASE |
| P2 `roles/iap.tunnelResourceAccessor` at project scope (W1) | the stop identity's tunnel grant is a project binding with IAM condition `ci-stop-iap-instance`: `resource.name == "projects/<number>/iap_tunnel/zones/<zone>/instances/<env>-vm"`; setup, `--dry-run`, `--verify`, `--teardown` use it; the script refuses an unconditioned grant | ci-federation 93/0: exact-condition case, project-wide `--verify` tamper, 3 mutations killed |

Validation of the corrections (executed and failed first, preserved):
1. **S2 rerun on `31069b65`:** LIFECYCLE-DRYRUN FAIL (my `check.sh` expectation of the new describe line was wrong twice: first omitted `--raw` and the projectNumber field, then the dry-run's
   `printf %q` renders the comma as `\,`); fixed in `31069b65` and `f08b9f91`; PASS after.
2. **S2 rerun on `31069b65`:** CI-ADAPTER FAIL on one case, `cand-branch-pinned` (STAGING_REUSE missing). It did not reproduce in 3 standalone full runs or 3 group runs, so it was investigated rather than rerun. W2 found a
   **real code defect**: `meta_class` read a bare `412`/`403`/`404` anywhere in gcloud's error text, including inside the object URL, so a commit SHA containing `412` (`…f412c…`) turned a registry 404 into a precondition failure; the check job died with
   exit 2 and wrote no `reuse.env`. Run-to-run SHA differences made it intermittent. Fixed in `c1514fb2`: `gs://` URLs stripped before classifying, only standalone status codes accepted, registry-404 and impersonation-denial checks tightened,
   two `grep -q` under `pipefail` checks made safe, fixture commit dates fixed, failing jobs' stdout/stderr/calls kept in `REPORT/jobs/`; cases `class-meta` and `control-class`. Loops: before 15 full + 10 group runs (parallel, niced) no repro; after 10 full runs 99/99.
3. Hosted pipeline on `31069b65` (2997291529): `validate:checks` success.

Remaining unverifiable offline (added): whether `gcloud storage buckets describe --raw` returns `projectNumber`; whether the deployer has `resourcemanager.projects.get` (local runs); whether IAP matches the instance name or only the numeric ID in
the conditioned resource name (a mismatch denies, never widens); gcloud's parsing of `==` in `--condition`; remove-binding matching on the condition.

## 11. Source review round 2: Codex REVISE (P2) and correction (append-only)

Hosted pipeline 2997354302 on `6e8e212d` passed (`validate:checks` success; `validate:integration` manual, not run) before any further push.

| Finding | Correction | New offline evidence |
|---|---|---|
| P2 `ci-lib.sh` accepted any HTTP 200 body without validating it as JSON: a truncated `{"kind":"storage#testIamPermissionsResponse",` passed as a denial | `probe_verdict` parses the body with jq (`PROBE_JQ`): exactly one `testIamPermissionsResponse` object whose `permissions` is absent or an empty array is a denial; a non-empty string array is a grant; every other shape (truncated, wrong kind, non-array or non-string permissions, empty body, two values, null, array) fails closed (exit 6, no reuse decision, zero writes). Merged from `g072-w2` `1d644a0f` | ci-adapter 107/107 (8 new cases incl. mutation control `control-probe-grep`), env-lease 37/37, shellcheck clean |

The offline fixture runs `PROBE_JQ` through a stand-in `jq`, so I also ran the exact program with the real pinned jq (`ghcr.io/jqlang/jq:1.7.1@sha256:096b83865ad59b5b02841f103f85c51318394331bf1995e187ea3be937432`, network none) on 12 body shapes: all verdicts matched the expected ones.
Unverifiable offline (added): that the hosted/staging CI job image provides a jq at that version; the real-jq check above used the pinned jq image on the host only.

Proof on merged head `77355196` (base `f5465109`): S2 7/7 rows PASS (ledger `offline-6`; CI-FEDERATION 93/0, CI-ADAPTER 107/0, ENV-LEASE 37/0, CI-LAUNCHER 80/0, LIFECYCLE-DRYRUN, STAGING-GREP, SHELLCHECK-DIRECT), S1 (`lint lint-base provenance`) all PASS (`static-6`).
S4 and EP-1 were not rerun: S4 last ran on `71b410f3` and the later deltas (`ci-lib.sh`, `ci-adapter.sh` tests/fakes) touch neither `.gitlab-ci.yml` nor `validate-controls.sh`; EP-1 last passed on `6e8e212d` (`ep1-3`) with the artifact and Dockerfile inputs unchanged.

## 12. Duo review of `aa62e1d4` and correction (append-only; changes the accepted executable tree)

Duo (requested on the draft MR, first review, no prior threads) requested changes with one valid minor finding: `show_summary` in `infra/ci/staging-action.sh` filtered with `^[a-z_]+: `, so the `receipt_sha256` summary line (a key containing digits) never reached the job log, although the docs say the log carries it; the launcher test only checked the artifact file. Verified by reading the filter and the `SUMMARY_KEYS` list.
Disposition: **correction**, not a documented risk. Commit `3c267994` allows digits in the key pattern and extends `summary_ok` in `infra/staging/tests/ci-launcher.sh` to require every summary key in the job log. Focused proof on `3c267994`: CI-LAUNCHER 80/80 and SHELLCHECK-DIRECT PASS (ledger `duo1-fix`); control with the old filter (`--allow-dirty`, restored afterwards, tree clean): CI-LAUNCHER FAIL, 3 of 80 cases (`p-deploy`, `p-no-summary`, `p-stop`), SHELLCHECK-DIRECT PASS (`duo1-control`).
Duo states its diff was truncated: `.gitlab-ci.yml`, `staging.sh`, `ci-federation.sh` and the Terraform files were not reviewed by Duo; that coverage is not claimed here. The executable tree is now `3c267994` (two files changed against the accepted `77355196`); the other offline rows were not rerun because no other executable path changed.

## 13. Source review round 3: Codex REVISE at executable head `3c267994` (Result head `e9a51344`) (append-only)

Codex cleared the `3c267994` delta (summary regex and launcher assertion) by source inspection, then reviewed the paths Duo omitted. `infra/terraform/bootstrap/ci-federation.sh` has three P1 fail-closed/ownership gaps (the other omitted paths: no finding reported); MR !119 stays draft, no auto-merge. I confirmed the cited sites in the source before dispatch (`rd()` at line 247, `storage ls ... || true` at 516, adoption by `buckets describe` at 411, teardown at 553):

| # | Finding | Required correction |
|---|---|---|
| 1 | Metadata bucket ownership is not verified before mutation: setup adopts any accessible existing global `gs://$meta_bucket` and updates settings/labels; teardown recursively deletes it. The script validates the configured `project_number` but never compares the bucket's owner. Violates Spec R11 and Plan §2.4/§10.3/§10.5 | owner `projectNumber` compared with the pinned project number before setup updates, before teardown deletion and in verification; foreign/unknown owner fails closed; foreign-owner fixture proving no mutation or deletion and a matching-owner case |
| 2 | Teardown suppresses inventory read errors (`storage ls ... \|\| true`): an unreadable bucket counts as empty and teardown continues to remove bindings, accounts, roles, pool and bucket contents | failed list stops before any teardown mutation; confirmed absent distinguished from failed/forbidden; list-failure fixture (nonzero, zero mutations) and live-object confirmation cases |
| 3 | Verification reads fail open: `rd()` turns every gcloud read error into empty successful output, so expected-empty user-managed-key and state-bucket IAM checks and a failed secrets list can report `ok` | read errors preserved; a failed required read fails verification and no `ok` receipt; fixtures for failed key listing, failed expected-empty IAM read, failed secrets listing |

Correction owner: a fresh W1 writer (`g072-w1b`, Opus 5.5 high) owns `infra/terraform/bootstrap/ci-federation.sh` and `infra/staging/tests/ci-federation.sh` (fake gcloud only additively). The correction record and its evidence follow in §14 after it is merged and proven.

## 14. Round 3 corrections (append-only)

Hosted pipeline 2997523484 on `e9a51344` passed (`validate:checks` success; `validate:integration` manual, not run) before any further push. W1b (`g072-w1b`, head `7b8b28af`, based on `e9a51344`, two files changed) was merged into `g072-impl` as `516c9adb` (executable tree).

| # | Correction (`infra/terraform/bootstrap/ci-federation.sh`) | New cases (`infra/staging/tests/ci-federation.sh`) |
|---|---|---|
| 1 | `meta_owner` runs `storage buckets describe --raw --format=value(name,projectNumber)`; the name must be the bucket and the owner the pinned `--project-number`. A foreign, missing or malformed owner, another name and any describe failure other than a confirmed 404 stop the run: before setup's first change, before teardown's first change and again before the recursive delete; `--verify` compares it (`meta-bucket owner`). A confirmed 404 is absent. | `owner.setup-foreign`, `setup-missing`, `setup-malformed`, `setup-describe-denied`, `setup-matching`, `teardown-foreign`, `teardown-describe-failed`, `teardown-absent`, `verify.owner-foreign` |
| 2 | `meta_live` replaces `storage ls … \|\| true`: only gcloud's "One or more URLs matched no objects" is empty; any other failed listing stops teardown before any change, also with `--confirm-meta`. Existence reads (`exists`, `present`, `value_of`) treat only a not-found answer as absent and stop on anything else (a denial is checked before a not-found wording). | `teardown.list-denied`, `list-404`, `list-denied-unconfirmed`, `policy-unreadable`, `setup.pool-unreadable` (live-object confirmation cases kept) |
| 3 | `rd` records each failed read; the item is UNREADABLE, verification fails and no `ok` receipt is printed (`read_ok`); the failed secrets list has its own item; the `\|\| true` at the services list only covers grep's no-match. | `verify.keys-unreadable`, `state-policy-unreadable`, `secrets-unreadable` |

Evidence: federation 93/0 on the base, 93/0 after the script change alone, 116/0 final (17 new cases); six mutation controls killed (`owner-blind`, `owner-unverified`, `describe-failure-absent`, `list-failure-empty`, `verify-reads-tolerated`, `value-reads-tolerated`; transcripts show the old behaviour returning: `ok`, bucket `removed`, `absent`). The writer's evidence is in its task directory. I read the delta and ran S2 on the merged head `516c9adb` (base `f5465109`; ledger `offline-7`): CI-FEDERATION 116/0, CI-ADAPTER 107/0, ENV-LEASE 37/0, CI-LAUNCHER 80/0, LIFECYCLE-DRYRUN, STAGING-GREP and SHELLCHECK-DIRECT (pinned `koalaman/shellcheck:v0.11.0`) PASS; S1 `lint lint-base provenance` all PASS (`static-7`). The shared fake `gcloud` is unchanged, so the other suites only reran as regression.
Unverifiable offline (added): the real gcloud not-found wording (`404`/`NOT_FOUND`/"not found", denials checked first), the real empty-listing text "One or more URLs matched no objects" (the fake models the same text), `buckets describe --raw` returning `projectNumber`.

## 15. Round 4: Codex P1 on `729ae0a7` (not_found substring parser) and correction (append-only)

Hosted pipeline 2997619280 on `729ae0a7` passed (`validate:checks` success; `validate:integration` manual, not run) before the next push.

**Finding (Codex, P1, source review of `729ae0a7`):** `not_found()` in `infra/terraform/bootstrap/ci-federation.sh` scanned all stderr for a standalone `404`, not the API status. The accepted project-ID regex allows IDs such as `demo-404-preview`, so the derived metadata bucket name contains `-404-`; a transient `HTTPError 503` echoing the `gs://…` URL was classified as not-found. In teardown the first owner check then set `absent`, which skipped the live-object listing and the `--confirm-meta` guard, and if the second owner check succeeded the recursive delete could remove live metadata without confirmation. The round-3 tests missed it (fixed project `example`; the describe 503 fault omitted the resource URI).

**Correction (W1b, `g072-w1b` `d21df8bb`, merged as `318f0b40`):** `gerr_class` classifies by gcloud's status marker only, after removing every URL, `projects/` path and every name the script derives (project, number, bucket, accounts, pool, roles, secret, repository, instance, env). A denial and any other status code or token win; not-found needs an explicit marker (HTTPError/HTTP/status/code 404, `[404]`, `(404)`, "not found: 404", `NOT_FOUND`, or compute's "The resource '…' was not found"); `meta_live`'s empty-bucket decision uses the same classifier (`nomatch`). Teardown also refuses when the bucket was absent at the first owner check but is present before the delete (its live objects were never listed).

New cases (8 named; the suite's pass count rose by 10): `teardown.owner-recheck`, `p404.project-id` (the legal project id with `-404-` is accepted), `p404.teardown-503-uri` (first-describe-only 503 including the bucket URI: refusal before any write, live objects preserved without `--confirm-meta`), `p404.setup-503-uri`, `p404.verify-503-uri`, `p404.teardown-policy-503`, `p404.teardown-404-absent` (a genuine 404 with such a name is still absent), `p404.teardown-nomatch-uri`. The test fake gives realistic not-found errors and can limit a fault to the first N calls. Counts: federation 116/0 after the classifier change alone, 126/0 final. Controls: `substring-parser` (old `not_found` swapped back in) kills the teardown-503, setup-503 and policy-503 cases; `owner-recheck-removed` kills `teardown.owner-recheck` (the recheck alone also stops the delete under the old parser).

Proof on merged head `318f0b40` (base `f5465109`, ledger `offline-8`): CI-FEDERATION 126/0, CI-ADAPTER 107/0, ENV-LEASE 37/0, CI-LAUNCHER 80/0, LIFECYCLE-DRYRUN, STAGING-GREP, SHELLCHECK-DIRECT (pinned `koalaman/shellcheck:v0.11.0`) PASS; S1 `lint lint-base provenance` all PASS (`static-8`). Shared fakes and `meta-lib.sh` are unchanged.
Unverifiable offline (added): the exact gcloud error wording (compute's not-found carries no status code, so its sentence is relied on); a name of three or more letters that also appears in a genuine error can only turn a real not-found into a refusal, never the reverse; the service-account e-mail is not stripped.

## 16. Round 5: Codex P2 on `318f0b40` (bucket ownership race in setup) and correction (append-only)

Hosted pipeline 2997682334 on `3b6601ff` passed (`validate:checks` success; `validate:integration` manual, not run) before the next push.

**Finding (Codex, P2, source review of `318f0b40`):** `setup()` called `meta_owner absent` once, then performed API, WIF, account and role writes, and `meta_bucket_create` trusted the cached `META_STATE=present` when it updated `gs://$meta_bucket` and granted bucket IAM. If the bucket was deleted and its global name reused by another project in that interval, setup could update and grant on a bucket it does not own. Required: recheck `projectNumber` immediately before updating an existing bucket and before the bucket IAM bindings; verify a created bucket and refuse when create reports it already exists; a fake-gcloud regression where the owner is correct first and flips during an intervening call, with a mutation control removing the late check.

**Correction (W1b, `g072-w1b` `cf5f24bf`, merged as `d9909865`):** `run()` calls a new `meta_guard` before every non-dry-run write that names the metadata bucket, except its creation: a fresh owner check immediately before the settings update, each IAM grant and the deletion. A bucket that is gone or owned by another project exits 1. A failed create (name taken meanwhile) is refused, not adopted, and a newly created bucket is owner-checked right after. Only `infra/terraform/bootstrap/ci-federation.sh` and `infra/staging/tests/ci-federation.sh` changed.

New cases: `race.before-update` (owner flips before the settings update: refused, no update, no grant), `race.before-grants` (flip after the update: refused before any grant), `race.after-create`, `race.matching` (matching owner succeeds: 8 grants, 10 owner checks, exit 0), `race.create-taken`. The test fake can change a bucket's owner when a chosen call runs. Counts: federation 126/0 after the script change alone, 132/0 final. Control `late-owner-checks-removed` kills `race.before-update`, `race.before-grants` and `race.after-create` (setup updated and granted on the foreign bucket each time); `race.create-taken` has no control because the old code also stopped on a failed create.

Proof on merged head `d9909865` (base `f5465109`, ledger `offline-9`): CI-FEDERATION 132/0, CI-ADAPTER 107/0, ENV-LEASE 37/0, CI-LAUNCHER 80/0, LIFECYCLE-DRYRUN, STAGING-GREP, SHELLCHECK-DIRECT PASS; S1 `lint lint-base provenance` all PASS (`static-9`). Executable tree is now `d9909865`.
Unverifiable offline (added): ownership change on the real gcloud API, and the cost of one extra `buckets describe` per bucket write. Management-only staging `ec40db25` (Codex status record) not yet reconciled into this branch; the imported Result snapshot is `3b6601ff` and this Result is newer.

## 17. RC `0.2.0-rc.23` B-BROWSER-ROOT failure, locally reproduced defect and repair (append-only; dated 2026-10-28; revised after Codex review)

**RC23 outcome (unchanged, unqualified):** annotated tag object `7f83a2106fde0689df8b1d5905c4535f0f32b3a6`, peeled `d89931363ff7f3169bf8df4eca7997320fd225b9`; none of them moved. Promotion MR !125 was closed and `release/0.2.0-rc.23` deleted. Pipeline 2999310235 was canceled by the owner account after `validate:integration` (job 17540404890) failed B-BROWSER-ROOT at row 12/97; rows 13 to 21 had passed before the cancel. The job's only GitLab artifact is the trace `job.log` (15,893 bytes); the browser/JUnit/report archive was not produced. The final trace is retained locally at `.local/evidence/rc23/validate-integration-17540404890-final.log` (sha256 `e4d772d778f1ef0e25014c036b5f23ca1b1d4e7ee8a9caa51459e89914e2d20b`); it records the row failure but not the browser assertion. The hosted assertion was therefore not captured, and RC23 is retained as canceled and unqualified, never reported as passing.

**Locally reproduced defect; likely hosted cause, attribution unproven.** The reproduction below proves that S-APPEARANCE-CONTROL has a transient-highlight race under 1-CPU matrix load on the exact RC23 candidate tree. Because the hosted job's browser assertion/report was not captured, it does not prove that this was the assertion behind the hosted B-BROWSER-ROOT failure; the attribution rests on the matching starved wall time and the absence of any other failure at 1 CPU.

Private reproduction on the exact candidate tree `d8993136` (README `platform:browser` command with the harness's report injection, private compose stack): unconstrained run 172/172 (325 s); browser container limited to 1 CPU (hosted wall time was a comparable ~9m40s) 171 passed, 1 failed in 546 s: `[S-APPEARANCE-CONTROL]` `choose(page, "system", "dark")`, `appearance.spec.ts:99`, "the highlight moves to dark", Expected 1, Received 2. The helper waited only for *some* option to be highlighted and counted its arrow presses from that first reading; under a starved frame the open list reports a transient highlight first and then settles on the stored mode, so the press landed on the wrong option. This is the transient-highlight arrow-count race, **distinct from RC22's S-MOTION** (animationstart delivery, fixed by !124). The unfixed test alone passes 40/40 at 1 and 0.5 CPU, so it needs full-matrix load. A 0.5 CPU run of the unfixed tree (extreme starvation, reference only) had 170 passed and 2 failed, both S-CONTRAST 120 s timeouts (a different symptom that does not occur at 1 CPU).

**Repair:** MR !126, source `e3f6cbaa742eef5f76b42180e2878ba9d271c157`, only `web/tests/browser/appearance.spec.ts`: after the list opens, wait until the already-asserted stored mode (`from`) is the highlighted option before counting presses; assertions and loop unchanged, all five `choose()` callers covered. Hosted fast pipeline 2999778426 PASS (`validate:checks` success; `validate:integration` manual, not run); Duo clean with zero unresolved discussions; Codex source review ACCEPT. Merged as `21cc8f6357ca6c83028c351274a0dcd1ea32f171` (tree `7e25fd62…` identical to the source tree).

**Affected coverage:** README `platform:browser` matrix, 1 CPU, private stack: fixed source `e3f6cbaa` twice, 172/172 (681 s, 551 s); **exact merged SHA `21cc8f63`: 172/172 passed in 558 s, skipped 0**, cleanup containers 0 and images 0. Not the harness `root` suite: its default host ports collide with an unrelated container on 5432 here, so the same README command and report injection ran on private reserved ports. Evidence (private): `.local/evidence/rc23-browser-root/{cpu1,cpu0.5,appearance-old,fixed-cpu1-1,fixed-cpu1-2,merged-21cc8f63-cpu1}`.

**Not done, by direction:** no full integration run, no new RC tag, no promotion MR. Active task bindings unchanged. Remaining cleanup (worktrees `g072-rc23-frozen`, `g072-rc23-merged`, local branch `maint-browser-root-rc23`, image `g072-rc23-motion:browser`, evidence directories) is held for final ACCEPT; the appearance reproducer image was removed.


## 18. L-P0 preview project and budget — 2026-10-28 (Codex operational receipt)

Owner reports completing project creation and billing attachment; Codex independently
read back `saas-foundation-preview-1`, project number `<gcp-project-number>`, ACTIVE, folder
`101398290430`, and enabled billing account `<billing-account-id>`. The owner
approved matching staging's $25 USD monthly alert at 50%, 90% and 100%. Codex
created and independently described budget `<redacted-uuid>`:
MONTH, filter only `projects/<gcp-project-number>`, CURRENT_SPEND thresholds 0.5/0.9/1.0,
default billing IAM recipients enabled. No custom notification channel, automatic
shutdown or spend cap is configured by this budget. L-P0 is complete; G1a, preview
bootstrap, federation, CI live proofs and WorkOS/G057 work remain separate steps.
The project and budget are retained for preview-1 under the campaign owner; no
preview VM or deployment was created by these actions.


## 19. GitLab G1a partial setup — 2026-10-28 (Codex direct execution)

Owner explicitly selected Codex-only setup today and authorized proceeding with
GitLab setup. Codex created/read back environments `staging` (42179007) and
`preview-1` (42179026), each protected with exactly one Maintainers deploy access
entry (`access_level=40`). The existing protected staging branch and project
pipeline-variable override restriction `no_one_allowed` were verified. Existing
`identity-staging` environment and WorkOS variables were preserved.

Three protected/raw file variables scoped exactly to `staging` were created and
read back: `STAGING_CI_CONF`, `STAGING_CI_BACKEND`, `STAGING_CI_TFVARS`. CONF comes
from the retained G070 profile, omits EVIDENCE_DIR/PRIVATE_DIR/GOOGLE_ADC_FILE and
explicitly selects SESSION_MAX_SECONDS=7200; backend/tfvars match operator bytes.
Values were submitted via request stdin and compared in memory, never printed.
Sanitized readback: `.local/evidence/g072-setup/gitlab-g1a.json`.

Initial JSON environment creation lacked an explicit Content-Type and returned
HTTP 415; the old CLI returned zero, and response parsing failed on missing id.
Codex reconciled actual environment/protection state, corrected Content-Type and
HTTP/error handling, then completed idempotent creation/readback without duplicate
environments. No deployment pipeline or GCP mutation occurred in this step.

G1a is partial: preview's three profile variables await its concrete G070 config.
G1b `STAGING_CI_SETTINGS` in both scopes awaits verified federation provider and
identity pins; no placeholder provider or fabricated pins were uploaded. Resource
groups await the first authorized proof pipelines. This is configuration proof,
not a successful CI deploy or final G072 ACCEPT.


## 20. Preview bootstrap preparation — 2026-10-28 (Codex)

Owner requested preparation of the preview configuration and bootstrap plan.
Three 0600 private profiles are prepared under `.local/secrets/g072-preview-1/`,
with independent backend/project/accounts/registry/domain. The existing bootstrap
and first-allocation provision dry-runs passed on source `f49d978c`; no live
bootstrap/provision, image build/push, DNS or token access was performed. The
missing candidate image manifest remains explicit. Local ownership, input hashes
and proof hashes are in `.local/evidence/g072-setup/preview-1/preparation-inventory.json`.
The concrete bootstrap proposal (not included)
names the owner-run command, resource/IAM set, two DNS token containers, costs,
rollback and later provisioning holds. It is ready for the owner's bootstrap
step; preparation does not claim live readiness or Goal acceptance.


## 21. Preview bootstrap and DNS secret metadata verified — 2026-10-28

Owner completed bootstrap, separate Terraform Cloudflare secret creation and
removal of the default Compute account's automatically granted Editor role.
Codex independently read back the expected project-owned state-bucket settings,
accounts/IAM, registry, nine APIs and 14 secret containers; all eight generated
application secrets and the two DNS token secrets have enabled version 1. No
secret payload was accessed. No preview VM exists. Root checkout remains clean
after publication. The deployment proposal (not included)
records the unchanged seven-image build reuse and next bounded VM/DNS release;
image publication/provisioning are not yet performed.


## 22. Both federation profiles configured; WorkOS next — 2026-10-28

Owner changed campaign sequencing to configuration first and a coordinated live
validation session at the end. Required separate login, start/stop, denial, reuse
and cross-environment checks remain; this is not a coverage waiver. VM startup
is held while identity configuration is prepared. The accepted L-W1p early
federation path is used; no deployment is inferred from configuration.

Owner ran the reviewed preview and staging federation commands. Codex's full
staging `ci-federation.sh --verify` passed (exit 0, final `staging ok`), including
the existing stopped VM's policy. Preview verification exited 1 solely on the
instance IAM read for the deliberately uncreated preview VM (404/UNREADABLE);
all other reported provider, account/key, role, bucket, project, registry and
secret binding rows were ok. Preserve that failed overall verification and
repeat the VM-dependent/full proof during live validation. Proofs are retained
under `.local/evidence/g072-setup/{staging,preview-1}/federation-verify.txt`.

Both ACTIVE providers are read back at projects/<gcp-project-number> and
projects/<gcp-project-number>, pools staging-ci and preview-1-ci, provider gitlab, with
exact environment-specific start/stop/read audiences. Staging remains TERMINATED.
No secret payload or WorkOS credential was accessed. Metadata reads show all
three WorkOS secrets in both projects have no versions; staging's retained
receipts also explicitly pin all three to none.

Codex uploaded and byte/metadata-readback verified all four protected/raw file
variables per environment: STAGING_CI_CONF/BACKEND/TFVARS/SETTINGS, exactly scoped
to staging or preview-1. SETTINGS contains each real verified provider and the
complete current all-none identity triple. Supervisor selected only the known
sibling state/CI metadata bucket names as read-only denial-probe targets; the
future live checks must prove those permissions are denied. Sanitized proof:
`.local/evidence/g072-setup/gitlab-complete-profiles.json`; private settings
files are beside each environment's retained operator profile. G1a/G1b profile
configuration is complete for the current unauthenticated baseline; numeric
identity triples will replace none only after the WorkOS setup gates. No live
GitLab federation authentication/deploy proof is claimed from these readbacks.
Resource-group process settings await their creation by final proof pipelines.


## 23. WorkOS issuers verified; complete numeric deployment pins — 2026-10-28

Owner configured separate Staging and Preview 1 WorkOS environments, application
Deploy in each, preserving staging's original Local application. Owner reports
callback/sign-out/default settings saved and entered each Deploy client ID/API key
through hidden-input stdin into its own GCP secrets (enabled version 1). Both
issuer containers were initially empty. Owner then explicitly authorized the
small API-only issuer-discovery operation, including temporary user/session
creation, verification, cleanup and saving only verified issuer values.

Codex reused byte-identical production token.go/jwks.go in a standalone helper,
with the existing synthetic-user/TOTP routines and no application DB. Compilation
and both sequential live checks passed. For each environment: exactly one owned
example.org user; token obtained from fixed WorkOS TLS endpoint; trusted client
JWKS and production RS256/client/subject/time/impersonation validation; owned
session revoked; user deleted and GET 404 confirmed; only then issuer saved as
enabled numeric version 1. Client IDs, API keys and issuers compared in memory
are distinct across environments. No credential values are printed or committed.

Both GitLab SETTINGS variables now pin the complete CLIENT_ID/ISSUER/API_KEY
triple to 1/1/1; exact scoped/protected file-variable readback passed. These six
secret versions are owner deployment configuration retained through qualification,
not test cleanup targets. Actual serving deployment/receipt and browser/CI live
proofs are still pending. Staging has not been started/redeployed by this check.

Proof and ownership: `.local/evidence/g057-issuer-discovery/` holds sanitized
per-environment results, lifecycle journals, exact helper source and inventory
with source/binary hashes. API keys and tokens were stdin/process memory only.
Temporary private issuer files were deleted after Secret Manager write/readback.
Four exactly inventoried containers (two compiles, two live runs) are verified
absent; no image/volume/network was created; shared Go image retained. Helper
binary/scratch and empty private issuer directory were removed. No raw API body,
JWT, password or TOTP secret is retained. This is issuer/configuration evidence,
not a full G057 qualification or end-to-end GitLab deployment proof.


## 24. Preview provisioned; first GitLab live actions exposed launcher gap — 2026-10-28

Codex executed the owner-approved preview release after the owner refreshed ADC.
The first plan attempt stopped on missing local `jq`, before apply; a checksum-
verified official jq 1.7.1 binary was installed under `.local/tools/`. The next
saved plan was independently inspected: 11 creates, no updates/destruction,
confined to the approved preview VM/disks/network/firewalls/address/operator IAM
and three DNS-only A records. All 11 resources applied in `saas-foundation-preview-1`.
The seven existing candidate `98b0c538` image IDs were pushed without rebuilding;
the canonical manifest now records the registry digests.

Provision's automatic deploy used numeric WorkOS pins 1/1/1, by sourcing the
unchanged operator functions and initializing their existing PINS array before
main (the provision CLI itself has no pin option). Readiness returned READY at
18:49:54Z, all seven runtime digests matched, and the retained DB marker was
recorded. HTTPS root redirects to `/signup`, signup and auth-host sign-in return
200, and browser-trusted Let's Encrypt TLS validates. This is serving/readiness
proof, not completed user signup or G057 Part B. Receipt
`preview-1-98b0c5382f89-1` sha256
`4d6beb851e42e6f00471aa1d27ebfbe7b9d9aab36246a994128b3cca0e297f2f`.
Observed migration lineage reconciled as verified (22 versions); its one fence
entry was cleared. Preview metadata manifest published. Staging metadata imports
published its existing pushed manifests/receipts without starting its VM.
Both initial lineage setups remain accurately recorded as assumed before this
preview observation.

The previously absent preview CI stop-account instance OS Login binding was added
once the VM existed. Full federation verification now passes for preview and
staging. `staging-env-staging` resource group is `oldest_first`; preview/build
resource-group modes remain to be initialized/read back. Project token-free
checkout flag was explicitly set true (it was already requested by job YAML).

Two staging stop pipelines, 3000743934 and 3000747562, both failed before a context
or cloud action: the runner left a password-bearing `.git/config` remote despite
its token-free flag. No federation/start/stop PASS is inferred. The validation
launcher already sanitizes these runner-created remote URLs, but the staging
launcher omitted that step. Focused repair MR !128 at `ca7b02d1` adds the same
sanitizer before the unchanged credential-absence scan. Unsafe section URLs and
stray token files still refuse. CI-LAUNCHER 83/83, warning-level ShellCheck and
shell syntax pass. First fixture run was 82/83 because the new unsafe-URL case
mistook expected cleanup Docker calls for a started context; the assertion was
corrected and the full focused set rerun. Duo and hosted fast are pending.

Preview is independently observed TERMINATED at <staging-external-ip> after the initial
proof; its 7200-second STOP scheduling, disks, certificates, DNS and images are
retained. Staging remains stopped and unchanged. Private proof/inventory resides
under `.local/evidence/g072-setup/`, including plan/apply, HTTPS, federation,
metadata/lineage, serving receipt, failed CI traces and focused-repair results.
The isolated repair checkout is `.codex/worktrees/maint-staging-launcher-credentials`;
its branch and reviewed source are retained until merge and live proof. Disposable
fixture trees were reclaimed by their harness; the labelled ShellCheck container
was removed, with IDs/events recorded. No full suite was selected for setup.

Next: clear !128's ordinary source/CI/Duo gates, merge, then repeat the actual
GitLab action proofs. Remaining G072 CI reuse/concurrency/denial proofs and G057
browser qualification are not reported as completed.


### §24 source acceptance — 2026-10-28

Codex ACCEPTs the two-file repair source `ca7b02d15b513fe8c7baae06c68f29c0ba7b4376`,
subject to its ordinary MR gates. A separate read-only Codex review found no
blocking correctness/security findings: the sanitizer matches the existing
validation-launcher algorithm, preserves username/host/path, and runs before
the unchanged literal-token scan. Unsafe section names and credential/scan
failures still refuse before context creation. Recorded focused proof is 83/83;
ShellCheck at repository warning severity is clean. The first unrestricted
ShellCheck invocation emitted existing informational diagnostics; it was not
misreported as a clean warning-severity run. No application image rebuild or
full controls suite is justified by this repair. Duo/hosted fast remain pending;
live CI actions remain unproven until the merged control is exercised.


### §24 owner action-label amendment — 2026-10-28

Owner requested the exact user-facing action choices **Validate**, **Deploy
environment** and **Stop environment**, folded into existing repair MR !128.
Source `d67dbe1b` updates the typed options/default, matching workflow/job routing,
the exact B-CI-CONFIG expectations/mutation input and operator documentation.
Internal `staging:*`/`validate:*` job names, environment actions, WIF audiences,
permissions and deployment behavior are unchanged. API callers use the same
new case-sensitive strings as the dropdown. YAML routing assertions, shell
syntax and GitLab CI lint pass; the changed MR head restarts its normal fast/Duo
reviews. Previous acceptance of `ca7b02d1` covers the credential repair only;
this small label delta is separately reconciled before merge. No cloud action
or full suite was added by this amendment.


### §24 failed fast attempt and owner-directed timeout repair — 2026-10-28

Pipeline 3000826432 / job 17552383718 on `be097fbf` failed (exit 124,
1242-second job duration). It emitted B-DEPSCAN FAIL at 13/35, then reached
30/35 before its internal 1158-second harness deadline. The supervisor's live
updates incorrectly missed that earlier failure; this correction supersedes
those “no failures” statements. The complete trace and artifact archive are
preserved in `.local/evidence/g072-setup/mr128-fast-failure/` and the final trace.

Artifact `ci-evidence/diagnostics/depscan.log` attributes the scan failure to
phase A: core module download/verification and pinned govulncheck installation
hit the 300-second `gova` limit. No vulnerability finding was produced. Owner
explicitly directed fixing it and increasing the fast deadline to 25 minutes.
Repair `bc0bc05f` in the same MR raises that phase to 600 seconds, preserving
scanner versions, hash verification, offline limits, findings policy and refusal
on expiry. An actual fresh-cache phase-A proof at 2 CPUs passed in 26 seconds;
this did not reproduce the hosted 300-second stall and is not a full B-DEPSCAN
PASS. Its two containers/two volumes were removed and absence verified; shared
pinned images retained. Proof: `depscan-phase-a-proof.json` and matching driver,
logs and CID files under `.local/evidence/g072-setup/`.

Fast/affected harness caps are now 1500 seconds. The fast job's outer 40-minute
and script 35-minute allowances accommodate startup and the unchanged five-minute
reserve while preserving the 25-minute harness cap and four-minute after-script.
Integration/full/staging caps and reserves are unchanged. Exact CI assertions and
OPERATIONS agree. Focused V37 passed all 18 budget/time cases, including the new
configured fast/affected cap check after two minutes of startup; its two
end-to-end malformed/exhausted-budget refusals also pass. Separate read-only
budget review ACCEPTed. Syntax and GitLab CI lint pass. A ShellCheck invocation
without SCRIPTDIR emitted existing cross-file variable warnings; matching the
repository's source lookup is the authoritative rerun and passed without diagnostics.
Codex ACCEPTs combined source `bc0bc05f` subject to the ordinary MR gates.
Hosted fast and Duo on
the combined repair remain gates; no merge, new VM start or full suite is claimed.


## 25. Repair merged; real GitLab stop proofs pass — 2026-10-28

MR !128 merged as `3059650e122fcfe956ebbc6eb825ae3e48f2f79d` after combined
source `bc0bc05f` received Codex acceptance, clean/approved Duo and hosted fast
pipeline 3000908234 / job 17553034717 PASS (35/35, including B-DEPSCAN).
The log confirms the new 1500-second harness deadline. The prior failed attempt
remains failed. No full integration run was selected. Root staging fast-forwarded
to the merge; executable/non-management files equal the tested source exactly.
Remote source-branch deletion was independently verified; its clean worktree and
local branch were removed after source/evidence preservation.

On merged control `3059650e`, real GitLab **Stop environment** pipelines passed:
- staging 3000953363 / job 17553405885: observed TERMINATED, exit 0; receipt
  `staging-000000000000-1`, sha256
  `32d183ca8af5b7c60fff35b1490736b3698e8330ae918a8e8191818de9ed4ad2`.
- preview-1 3000953407 / job 17553406177: observed TERMINATED, exit 0; receipt
  `preview-1-000000000000-1`, sha256
  `d34afe7196685cf9d04c9eafbe4450a86e98f967bdc1d7dd0733a6eb4c143248`.

Both launchers independently report zero password URLs/stored job credentials/
ID-token files in the project before context startup, and verified private-directory
cleanup. The real federation/stop path and its required denial probes executed
successfully. The initial sparse metadata legitimately has no candidate (receipt's
zero placeholder); this does not claim a serving deploy or overwrite its history.
Both environment resource groups are read back `oldest_first`; build group still
awaits its first deploy pipeline. Private summaries/artifacts and pipeline records
are under `.local/evidence/g072-setup/{staging,preview-1}/`.

Second consecutive preview stop 3000967137 is running. A temporary same-repository
candidate branch `g072-preview-live-proof` is created at the exact merged SHA for
the approved branch-selection proof, inventoried for deletion after deploy/reuse
proofs. Next: its deploy with custom duration, the staging deploy, simultaneous
readiness/isolation and preview same-SHA reuse/stop proofs. No G057 Part B browser
qualification or new final Goal ACCEPT is inferred from these setup checks.


## 26. Live deploy HOLD and owner-directed coordinator return — 2026-10-28

Second preview stop 3000967137 passed. Deploy pipelines 3000979883 (preview) and
3000980003 (staging) both passed prepare/check/build with candidate `3059650e`.
Preview's seven images were pushed/published, but job17553610404 failed HOLD
(exit3): its live duration changed to5400 s while Terraform remained7200 s, and
strict refresh scope refused additional resource/default-field drift. No refresh
apply or new serving release is claimed. Exact diagnostic is preserved in
`.local/evidence/g072-setup/preview-1/ci-deploy-1-published.log`; cause is unproven.
Preview is independently TERMINATED. Staging job17553611689 remained running at
handoff, with staging VM independently TERMINATED/max7200 s.

The owner explicitly directed bringing Claude back because foreground Codex
monitoring was blocking the main chat. This revokes the Codex-only setup hold.
The verified existing coordinator resumes diagnosis, scoped repairs and live
setup monitoring/execution. Its exact resume instructions, profiles, evidence,
remaining gates and resource retention are consolidated in the existing campaign
checklist's current-execution section; no second coordinator or new scope is added.


### §26 native diagnostic question — 2026-10-28

A waiting notice was inspected through the verified coordinator's existing
`claude attach` terminal. Its multi-select requested (1) read-only IAP SSH to
staging container states/logs and (2) loading the stored Terraform Cloudflare
token into process memory for a private read-only preview refresh plan.
Both are necessary within the owner's existing live-setup/repair authorization
and explicit return-to-Claude instruction. Codex selected those two routine
choices and verified the native displayed tool-result, then detached with Ctrl+Z
without stopping the coordinator. This was an owner-authorized answer entered
by Codex, not the owner personally clicking, and grants no new state apply,
reset/retirement or permission bypass. The attribution was also sent to the
verified inbox (queued at the time checked); no duplicate send was made.
Coordinator-reported new state: staging app/edge health failed after desired
WorkOS pins, remote deploy reverted to the prior release, VM RUNNING with
22:20Z automatic stop. This is reported pending independent diagnostic evidence;
preview's separate refresh HOLD remains. Claude owns diagnosis and monitoring.


## 27. Staging build-permission repair source ACCEPT — 2026-10-28

Codex ACCEPTs MR !129 source `d4a4d9a9012039b96ce58391dd74fa21f735c75b`.
The runtime failure is attributable to candidate worktree checkout under umask077:
Docker COPY preserves 0600 source files, which intelligence's uid10001 cannot
read. The change scopes umask022 to `git worktree add` in a subshell; the enclosing
private directory remains0700 and the caller remains077. No credential, network,
image-user or application behavior is weakened. A separate read-only reviewer
found no blocking findings.

Independently inspected regression evidence under
`.local/evidence/g072-setup/build-umask/`: old source fails `reuse-build-path`
with mode600; corrected source passes at644 and the reuse group passes9/9.
API verifies exact-head fast pipeline3001083397 success, Duo approval and zero
open findings. Standing staging merge authority applies after final exact-head/
input reconciliation. Claude owns merge verification, the fresh candidate/image
build and subsequent live deployment proof. Do not reuse the already-published
bad intelligence image/manifest from candidate3059650e or overwrite its failure
history. Select a new accepted candidate SHA and rebuild through corrected CI;
the old candidate's infrastructure also predates this compatibility change.

This is source acceptance, not serving qualification or final G072 ACCEPT.
The separate preview drift repair !130 remains under its own review/findings
and gates; no approval of it is implied here.


## 28. Preview refresh repair review — bounded REVISE — 2026-10-28

Reviewed MR !130 head `f7f2976be39f8e51495f5e6cd484fffa6a48b2fa`.
Exact-head pipeline3001150683 succeeded, Duo approved and reported documentation
findings are resolved. The new recorded-state allowances need one correction
before source ACCEPT: `benign_drift` restricts the attachment VM basename but
uses arbitrary project/zone regexes. With VM=preview-1-vm and configuration
saas-foundation-preview-1/us-east1-b, a users attachment in unrelated-project /
europe-west1-b with the same basename is accepted. A pure-function reproduction
is saved as `.local/evidence/g072-setup/refresh-cross-project-probe.txt`; no cloud
call/mutation was needed.

REVISE (P2): bind both users and users[0] attachment forms to the exact configured
project, zone and VM, refusing missing context. Add targeted wrong-project and
wrong-zone refusal cases (including the same valid VM basename). Keep all current
field/value restrictions and the saved-plan/no-normal-apply behavior. Reuse
unchanged evidence and run the affected focused lifecycle selectors plus normal
MR gates; no blanket CI-controls/full suite is requested. Claude remains the
implementation owner; this review grants no state apply or broader allowances.


## 29. Preview refresh correction source ACCEPT — 2026-10-28

Codex ACCEPTs MR !130 exact source
`047d0891dd458d7254e1dce537c1d10592da0081`. The sole P2 from §28 is closed:
attachment list/item values are exact JSON strings built from the configured
project, zone and VM, and missing context refuses. A separate bounded read-only
review found no new blockers. Independently executed pure-function probes pass
6/6: exact references accepted in both forms; the same VM basename in another
project or zone refused in both. Evidence:
`.local/evidence/g072-setup/refresh-exact-codex-final-probe.json`.

Independently inspected `refresh-lag-4/report/ledger.tsv`: LIFECYCLE-DRYRUN and
STAGING-GREP PASS; the ledger accurately names four accepted/23 refused scope
row sets. API verifies exact-head hosted pipeline3001218337 success and Duo
approval. The two stale documentation/count threads were independently checked
against corrected source and resolved by Codex. The saved refresh-only plan gate
still refuses other attributes, unknown values, remote membership/role changes,
replacements and non-no-op resource actions. No broad state apply is authorized.

Standing staging merge authority applies after final exact-head/reconciliation.
Claude owns merge verification, the new combined candidate after !129/!130,
fresh corrected image build and actual live deploy/session-recovery/reuse proof.
Failed candidates/pipelines remain failed. Source acceptance does not claim that
the preview now starts or that G057 browser qualification/final G072 acceptance
is complete. The new candidate must retain !129's readable build context; do not
reuse the known-bad3059650e image manifest.


## 30. Owner staging walkthrough: profile name and MCP failures — 2026-10-28

The owner reports two authenticated STAGING issues (not preview): account name
shows the unexpected fixture-like value “MQ A”, and the AI client connections
panel reports “Connections are unavailable right now. Your account and sign-out
still work. Please try again later.” The screenshots show a signed-in session;
this does not establish that account binding or the MCP backend is correct.
Screenshots/email/tenant/session specifics are retained privately, not committed.

The verified existing coordinator received both reports and owns read-only
diagnosis/normal bounded repairs. Verify reported email's WorkOS issuer/subject
and local principal binding before classifying the name as cosmetic; distinguish
provider profile data/test residue from a wrong-principal mapping. Do not reset,
reassign tenant ownership or rename blindly. The MCP wording maps to the generic
`unavailable` result in connection-manager.tsx; inspect actual listConnections
status/classification and safe server logs, rather than inferring ineligibility
from the UI. Preserve the owner's active walkthrough/session and notify before
an intentional stop. Existing setup diagnostic access persists.

The source/transport handoff is private
`.local/tmp/g072-owner-staging-account-mcp-report.txt`; inbox receipt confirmed,
execution/outcome not yet verified. These are open defects, not G057/MCP/browser
PASS. The isolated issuer-discovery helper used SyntheticIssuer and generated
example.org addresses, with verified deletion; no causal attribution to that
helper or any prior test is made without account evidence.


### §30 owner deferral — 2026-10-28

The owner subsequently directed: “We don't have to fix now, just record issues
ahead of 021.” This supersedes the immediate diagnosis/repair request in §30.
Both findings are recorded in BACKLOG with a pre-G021 roadmap review pointer;
no current app repair, new Goal/MR or test campaign is released. The verified
coordinator was sent the deferral, including parking only any newly started
work on these two findings at a safe boundary. Existing authorized environment
setup continues. Original screenshots/failure history remain private/preserved,
and no MCP or account-binding PASS is inferred from this deferral.


### §30 deferred findings: read-only diagnostic note — 2026-10-28

Inspected the coordinator's sanitised `owner-mcp-findings-1.md` under the existing
staging evidence directory. It reports Core's matching account-request trace
returned unmatched404 for /internal/v1/mcp/connections and no MCP_* configuration
was present (names only inspected). Canonical MCP docs/config explicitly define
all-unset as disabled. This supports the reported deployment-configuration cause
for the generic UI error, not an account-eligibility conclusion. Live trace/env
observations remain coordinator-reported rather than independently repeated.

The note reports the account's first signup and two workspace creations in the
walkthrough; the persisted/provider source of “MQ A” remains unproven. No DB or
WorkOS account read was performed. BACKLOG now records this diagnostic detail;
owner deferral remains in force. No app repair, new MR, broader investigation or
MCP/signup qualification is triggered. Claude confirms setup work continues
separately and these findings are parked.


### §30 owner clarification: investigation may continue — 2026-10-28

The owner clarified: “Claude can continue investigating but we don't need a fix
urgently.” This lifts the investigation pause for the two owner staging findings,
while preserving nonurgent/deferred repair scheduling. Bounded read-only,
owner-account-only investigation may continue at low priority relative to active
G072/G070 setup, with no blind account changes or secret/other-user exposure.
BACKLOG and the pre-G021 review pointer now reflect this clarification. It is not
an urgent repair implementation/MR release or a qualification PASS. The verified
coordinator was sent the updated disposition; prior directives/results remain
historical and the current setup work continues independently.


## 31. Preview interrupted-duration recovery lacks token — 2026-10-28

Owner reports preview deploy failed. API independently verifies pipeline3001317261
job17556221357 failed exit4; prepare/check/build passed. Its published sanitised
log was read and preserved as
`.local/evidence/g072-setup/preview-1/ci-deploy-latest-failure.log`.
The recorded source candidate is8e09ec61; images were pushed and manifest published.
Start then detects live5400s / Terraform7200s and enters recovery, but refuses
because CLOUDFLARE_API_TOKEN is absent. No new plan-drift cause is inferred; the
immediate blocker is missing token provision to the state-recovery path.

The verified coordinator received the concrete diagnostic and owns the bounded
CI recovery-condition repair under normal source/verification/MR gates. Focused
coverage must distinguish actual VM duration from recorded Terraform duration,
including live=desired while Terraform lags. No blind retry, scope-gate weakening,
reset or retirement is authorized. Preserve the failed run and reusable build.
Existing setup authorization remains; the separate owner account/MCP findings
remain low-priority investigation with nonurgent repair, not this setup blocker.


## 32. Interrupted-duration token repair source ACCEPT — 2026-10-28

Codex ACCEPTs MR !131 exact source
`6300d18702230cd5acafce120671a9d3a1f490c2`, subject to normal hosted fast/Duo and
exact-input gates. Separate bounded read-only review also ACCEPTs with no blockers.
The adapter retries once only when staging.sh start returns2, no token is loaded,
and its raw log contains the exact missing-token refusal. The real start recovery
requires that token before its scheduling/refresh/start writes, so the first
refusal is not a duplicate VM mutation. The retry retains the same lease and
requested duration; other failures and a second refusal still fail. Token
retrieval stays in the existing scoped federated path/private canary, is passed
only to start and cleared before subsequent steps.

Independently inspected preserved evidence in
`.local/evidence/g072-setup/recovery-token/`: unfixed code fails the new
life-stopped-recovery case; corrected lifecycle group passes13/13. The case uses
split VM5400s / Terraform7200s state, actual adapter/start fixture path, one secret
read, one provider start, no repeated set-scheduling, correct5400s summary and no
token in later Terraform. Reported ShellCheck/literal checks are recorded in the
implementer's MR; hosted fast/Duo remain coordinator-owned gates at this review.

The overall failed deployment had already published images/manifest; “before any
write” applies only to the refused start operation, not the entire pipeline.
Claude should keep that distinction in final documentation. Standing staging
merge authority applies once remaining gates clear. Claude owns merge/input
reconciliation, live recovery/deploy/reuse verification and cleanup in background.
This is source acceptance only; no new serving/final-G072/G057 PASS is inferred.


## 33. Optional MCP deployment configuration source ACCEPT — 2026-10-28

Owner superseded the deferred MCP repair hold and requested the diagnosed missing
configuration be fixed now in staging and preview. The account-name investigation
remains separate; no account edits or reset are authorized by this repair.

Codex ACCEPTs MR !132 exact source `c9d85499182fd037a378cbc6eb00a2f86e063210`
against `aeb71ccd9136e3ecab2b14bdff342df6cbd0e1be`, subject to hosted fast,
applicable Duo disposition and exact-input reconciliation. Independent bounded
read-only source review found no blockers. Optional issuer/origins propagate from
staging.conf through local validation, remote metadata and rendering to core only.
Empty defaults retain the disabled state; origins without an issuer are refused.
Core identity, token validation and tenant boundaries remain authoritative.
No credentials or test trust anchors are introduced.

Codex inspected the evidence head and four PASS ledger rows in
`.local/evidence/g072-setup/mcp-config/`: LIFECYCLE-DRYRUN, COMPOSE-STATIC,
SHELLCHECK-DIRECT and STAGING-GREP, all bound to the reviewed source. These are
preserved implementer runs, not independently rerun checks. Minor nonblocking
wording correction: five malformed MCP cases were added, not six as stated in
the MR and test outcome prose.

Claude owns remaining fast/Duo gates, normal standing staging integration,
canonical private profile and GitLab file-variable updates at a safe deployment
boundary, and corrected deployments in both environments. Both owner-supplied
AuthKit origins were independently verified against public OIDC discovery issuer
and JWKS metadata. Endpoint availability, connection create/list/revoke and actual
client authorization are distinct live proofs; this source ACCEPT claims none of
them and is not final G072/G057 acceptance. Preserve earlier failed deployments,
resource ownership, session deadlines and cleanup.


## 34. Owner name resolution and MCP rollout status — 2026-10-28

Owner corrected the WorkOS profile and confirmed the staging display-name issue
resolved. Codex relayed closure to Claude (receipt confirmed); no further account
investigation, reset or name repair is requested. Resolution is owner-reported;
no independent historical principal-binding proof is claimed.

MR !132 source `c9d85499182fd037a378cbc6eb00a2f86e063210` merged as
`c66dee7b0edb3cb71ef26137ee69af19d44be33b` at 23:05:43Z; Codex independently
checked GitLab's merged state. Both supplied public AuthKit domains returned exact
matching issuer and JWKS discovery values. Source acceptance remains in §33.

Claude reported a native auto-mode classifier refusal of the private profile and
environment-scoped STAGING_CI_CONF updates as secret-store writes. The coordinator
subsequently reported the owner's direct in-session instruction, “I want it
updated,” approving the two profiles and two scoped variables. This is recorded
as reported direct approval, not a bypass or proof of execution. No secret values
are included here. Configuration readback, corrected deployments and live MCP
connection/authentication proofs remain pending confirmation. Existing runtime,
resource ownership, failure preservation and cleanup policies continue.


## 35. Owner preview MCP confirmation — 2026-10-28

Owner reports: “Mcp worked on preview.” Record preview MCP functionality as
owner-confirmed. No independent test, specific client, create/list/revoke matrix
or serving SHA is inferred from this report. Staging MCP verification remains
separate and is not established by the preview result. Claude should reconcile
this observation with its existing deployment receipt without repeating the
owner's successful walkthrough solely for documentation.


## 36. Owner acceptance and closeout direction — 2026-10-28

Owner confirms “Everything is perfect,” accepts the delivered G070/G072 outcome,
and directs preview stop and retirement, preservation of running staging at its
current deployed candidate, direct staging publication of Result/acceptance
records, then RC25 tag and promotion MR. Codex relayed this to the existing
coordinator; delivery receipt is confirmed, execution remains to be verified.
Preview's held same-SHA reuse proof is not a prerequisite to the owner-requested
retirement and must remain unexecuted if not performed. No unexecuted check is
converted to PASS. Staging must not be retired or reset; preserve its data,
certificates and credentials for the owner's G057 sittings. Keep-up follows the
supported session lifecycle; report its exact deadline rather than implying
indefinite uptime. Management-only publication does not require another deployment.

Owner subsequently schedules G057 continuation for tomorrow or Friday
(2026-10-29 or 2026-10-30). Do not start its tests now. This deferral does not
cancel preview retirement, current closeout or RC25. Final cleanup disposition,
actual preview absence and release cut remain pending coordinator receipts;
RC25 must include completed records and receives normal full hosted qualification.
No promotion merge or stable release tag is authorized by this instruction.


## 37. RC25 abandoned by owner; retirement repair first — 2026-10-28

Owner supersedes the earlier open-RC instruction: abandon RC25 and finish the
retirement repair first. RC25 had already been created as promotion MR !133,
source `463c4a0620417ec2ede45f5859ad5a026d90fd83`. Codex closed !133, requested
cancellation of pipeline 3001655226, and removed temporary source branch
`release/0.2.0-rc.25`. Cancellation request returned while the pipeline still
reported running, so terminal cancellation remains to be checked. Preserve the
immutable RC tag and creation history; no qualification, main promotion or
replacement RC is claimed or authorized. Claude owns the ongoing retirement
repair and final cancellation/cleanup reconciliation. The earlier queued RC25
instructions are superseded; staging retention and deferred G057 remain.


## 38. Retirement IAM-path repair source ACCEPT; RC26 successor — 2026-10-28

Codex ACCEPTs MR !134 source `63526e3794480c993f7538c3ac1fa5a0f4de3e87`
against `463c4a0620417ec2ede45f5859ad5a026d90fd83`, subject to current
fast/Duo and exact-input reconciliation gates. Bounded independent read-only
review also found no blockers. Terraform's instance IAM member uses a full
instance path; normalization now requires the configured project and zone
before the existing exact environment object-name check. The own canonical
path is accepted; foreign project, zone and sibling instance remain refused.
The existing saved-plan scope gate is retained rather than bypassed.

Codex inspected exact-head evidence and the three selected PASS ledger rows in
`.local/evidence/g072-setup/retire-iam-scope/`: LIFECYCLE-DRYRUN,
SHELLCHECK-DIRECT and STAGING-GREP. These are implementer runs, not independent
reruns. Source acceptance does not establish actual retirement or absence.
Claude owns merge, authorized preview retirement through the supported lifecycle,
absence and cleanup receipt, preserving staging. A native refusal must be
reported without changing permissions or retrying via another route.

Owner requests RC26 when done: after the repair and final closeout records are
on staging, cut a fresh immutable `0.2.0-rc.26` and open its main-promotion MR
with full hosted qualification. Do not recreate RC25, move its tag or merge a
promotion without owner instruction. Preview retirement and acceptance receipts
must be reconciled before that cut; G057 remains deferred to the owner's next
sitting. Staging's reported deadline is 2026-10-29T01:44:57Z; retention does not
imply unlimited uptime or permission to retire/reset its dataset.


## Final ACCEPT — owner outcome disposition and retirement proof, 2026-10-29

Codex grants FINAL ACCEPT for the delivered G070/G072 environment setup under the
owner's explicit acceptance and closeout directions. Earlier failures and source
reviews remain historical. MR !134 merge `c1fc1833fe18dbd733371990f6c156910ba8b0dc`
contains the accepted retirement repair. Codex independently checked the GitLab
merge, inspected the preserved apply output (11 resources destroyed), absence
output and receipt hash `ffe6bb9891461166dfcd087eb0442af0cf00163477223ed3990ac40719fe020e`.
These are coordinator-run cloud observations, not new independent cloud queries.
A bounded read-only closeout review found no blocking source/security discrepancy.
The coordinator receipt (not included) is imported
byte-for-byte (SHA-256 `e8741409c952c4a5e8afcefc7a12c0dcd69e27a412bff343048142d5fdd9dfba`).

Preview VM, disk, address, scoped network/firewall and three DNS records are gone.
Preview project/bootstrap, secrets, federation, metadata/state and registry
images/manifests intentionally remain available for future reprovisioning; owner:
existing environment coordinator/operator, release condition: separately authorized
bootstrap decommissioning. Private profiles remain protected. Staging is retained
on serving candidate `ec7c4b3b`, data/certificates/WorkOS credentials intact;
reported auto-stop remains 2026-10-29T01:44:57Z. No indefinite uptime is claimed.

Operator-machine image aliases, task worktrees/temp and minimal evidence are not
claimed absent merely because no preview containers remain. Supervisor retains
these under the existing coordinator/run inventory through RC26 qualification
and receipt reconciliation; owner: existing Claude coordinator. Reclaim exact
finished task-owned objects at that safe boundary, preserving shared images,
private profiles and minimal failure proof. This is explicit retained-resource
accounting, not an all-resource CLEAN claim or permission for host-wide pruning.

Owner acceptance and immediate retirement supersede waiting for the held preview
same-SHA reuse test; it remains UNEXECUTED, not PASS. New retirement fixtures were
not run against the old code; that negative proof remains reading-only. No full
client enrollment matrix, reset proof or G057 Part B PASS is inferred. Owner
confirmed preview MCP works and the overall walkthrough is satisfactory. G057
continues on the owner's next sitting, October 29 or 30, and its own outstanding
gates/foundation barrier remain.

Closeout compaction assessed: replace the superseded G070/G072 current-index
entries now and preserve their exact original text in the dated archive.
Broader historical-index compaction is deferred to the post-RC26 reconciliation
boundary because other active goals retain bound references; no tests are rerun
for this documentation update. RC26 may now be cut after this publication from
the final staging tip, with full hosted qualification and no main-merge authority.


## Post-acceptance resource cleanup authorization — 2026-10-28 owner session

Owner explicitly requests cleanup of accepted G070/G072 worktrees, local and
remote branches, images and Docker resources. Codex sent the bounded cleanup
instruction to the existing coordinator; receipt is confirmed, execution and
removal counts are not yet verified. This releases the earlier local-artifact
retention through RC qualification where objects are finished and exclusively
task-owned; active RC27 inputs and shared images remain protected.

Use existing exact-owner inventories, preserve unique branch content before
removal, verify absence afterwards and report each resource category. Preserve
running staging, its data/certificates/credentials, cloud bootstrap, canonical
private profiles, GitLab/federation configuration, immutable RC tags and required
minimal failure/acceptance evidence. The coordinator's live worktree remains
retained until a safe supported handoff; no deletion beneath a running session.
No host-wide prune or new test campaign is authorized. Actual native refusals
must be reported without bypass. Append the verified removed/retained/unresolved
receipt here when available; this entry claims authorization, not completion.
Cleanup-only publication does not require moving a tag or cutting another RC.


## RC27 abandoned; RC28 waits for verified cleanup publication — 2026-10-28 owner session

Owner directs abandonment/closure of RC27 and deletion of its temporary branch,
then RC28 only after cleanup is finished and its Result is pushed to staging.
Codex closed promotion MR !135, requested cancellation of pipeline 3001710770
(response still running at request time), and successfully deleted temporary
branch `release/0.2.0-rc.27`. Preserve the immutable tag and history; no RC27
qualification or promotion is claimed. Claude must reconcile terminal cancellation
and continue cleanup. RC28 cut is conditional on verified cleanup receipt import
and publication; no further routine owner prompt is needed then. Use the final
staging tip including that receipt, normal full hosted qualification, and separate
owner authority for main promotion. Earlier RC27 instructions are superseded.


## Verified post-acceptance cleanup publication — 2026-10-29

Codex accepts the scoped cleanup receipt, imported unchanged as
G070/G072 cleanup receipt (not included), SHA-256
`eafab7cda739b21969d99ddcae4e883f533ea65a6a445e0df6ffa00f84cec893`.
Coordinator reports removal of 33 local branches, 9 remote branches, 12 worktrees
and 39 image tags, with zero task containers/volumes/networks remaining.

Independent bounded read-only review verified absence of the 33 listed local
branches, preservation of deleted unmerged tips, and a valid 16-ref bundle with
SHA-256 `e9808de2971c6d3b51a51e3abeecc8f2f77256b8086ecb0dc56849192c90f6e4`.
Codex checked remaining worktrees, remote heads (only main, staging,
identity-qualification and the two G057 planning branches), and current Docker
name/tag listings with no G070/G072/preview-1 matches. Deleted-image counts and
volume/network absence are coordinator-reported inventory observations, not
independent retrospective proof. No whole-host CLEAN claim is made.

The receipt's last RC27 uncertainty is now resolved: GitLab independently reports
pipeline 3001710770 canceled. RC27 MR remains closed, source branch absent and tag
preserved. Retained unique cleanup refs/bundle, active coordinator tree, G071/M1
and G057 trees, unrelated/uncertain resources, shared images, durable profiles,
bootstrap and minimal evidence remain explicitly accounted for in the receipt.
They are not silently deleted or reported absent. Existing coordinator owns the
live tree/scratch until safe handoff; G057 belongs to its deferred owner campaign;
uncertain older resources remain outside this cleanup's proven scope.

G070/G072 acceptance stands. Final cleanup records are now published; RC28 cut is
released from the resulting staging tip with full hosted qualification. No main
promotion merge, stable tag or staging reset/retirement is authorized.
