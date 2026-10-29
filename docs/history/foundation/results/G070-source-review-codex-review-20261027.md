> **Editorial note (public export).** Historical source-project record, retained as foundation history evidence; it is not this repository's active management state and grants no authority. Machine paths, session and account identifiers, personal addresses and live provider bindings are replaced with `<placeholders>`; links to records not included are marked "(not included)". Git commit/tree hashes identify the original source project and may not resolve here. See the [foundation history index](../INDEX.md).

# G070 source review — Codex

Date: 2026-10-27 UTC
Disposition: **CORRECTION LOOP; do not open the implementation MR yet**
Reviewed candidate: `origin/g070` `e04876f87d0a6c741ad056887dd1a68b7742d4e0`; executable source `7560f094`
Base: staging `c3e4a627c6550455b1fb3411049bf06282db1343`
Plan: accepted G070 Plan r1 with recorded corrections. G071 Plan r4 handoff also applies.

## Review boundary and evidence

I reviewed the final source, Goal/Spec/Plan, Result and referenced offline evidence. The reported final offline/static rows are consistent with the source changes and the recorded `SECRETS_FILE` driver omission is appropriately attributed to the driver. I did not run tests, contact providers, allocate cloud resources, or alter the candidate. A separate read-only source review independently raised the findings below.

The `tf-work` directory (1.6 GB) is recorded removed and verified absent in the task inventory. The 75 MB `tf-mirror` remains with a named G070 owner, offline-rerun purpose and final-acceptance release condition. Keep that retention in the inventory and remove it at the recorded condition. The Result reports no G070 containers, volumes or networks and preserves the earlier cleanup receipts.

## Required corrections

### G070-C1 — P1: isolate the VM metadata service from application containers

`infra/terraform/modules/gcp-staging-vm/instance.tf` attaches the VM service account with the `cloud-platform` scope. `infra/terraform/bootstrap/bootstrap.sh` grants that account `secretmanager.secretAccessor` on the environment's secrets, while `infra/staging/remote/lib.sh` obtains tokens from `metadata.google.internal`. I found no host-firewall or `DOCKER-USER` rule in the candidate that blocks Compose networks from the GCE metadata server. As written, a compromised application container may be able to obtain the VM identity token and read staging secrets, including database and identity credentials. The offline proofs do not establish a container-to-metadata boundary.

Add and persist a host-level metadata egress block for all application/container networks while keeping host-side lifecycle scripts able to use metadata. Prove from each relevant Compose service/network that metadata token access is denied, and prove the host deploy/start scripts retain only the access they need. Consider narrowing VM IAM to the minimum required secrets and permissions as defense in depth.

Google's guidance notes that processes on a Compute Engine VM can obtain the attached service account token by default and recommends host firewall restrictions for selected workloads: [service-account best practices](https://docs.cloud.google.com/iam/docs/best-practices-service-accounts#limit_metadata_server_access_to_selected_users_and_processes).

### G070-C2 — P1: readiness must not require apex and wildcard certificates to share a serial

`infra/staging/remote/lib.sh:https_ready` (`infra/staging/remote/lib.sh:469–492`) requires the certificate serial for the apex domain to equal the serial served for a random tenant label. Apex and wildcard names are distinct certificate identifiers; Caddy may correctly serve separate certificates, and a wildcard certificate does not itself cover the apex. This equality is not a valid readiness invariant and can reject a healthy deployment.

Validate the apex and random-label handshakes independently: trusted chain (or the explicitly accepted rehearsal issuer), hostname/SAN coverage, and expiry. Keep the random-label probe to establish wildcard routing, but do not require equal serials. Add an offline/unit proof for distinct valid certificates and a focused runtime proof for both hostnames.

Caddy documents wildcard certificates as covering names below the wildcard label, not the apex: [Automatic HTTPS — wildcard certificates](https://caddyserver.com/docs/automatic-https#wildcard-certificates).

### G070-C3 — P1: prevent a malformed Cloudflare token from appearing in diagnostics

The edge image pins `caddy-dns/cloudflare` v0.2.4. Its `Provision()` error path can include the raw API token when client initialization fails. `infra/staging/remote/lib.sh` currently rejects empty values and quote/line-break characters but otherwise forwards the value to Caddy. A malformed token can therefore place the secret in Caddy startup stderr and the retained deploy/start logs.

Validate the token before rendering or starting the edge, and ensure all failure paths emit a generic error without the value. Add a canary proof that a deliberately malformed secret is absent from command output, container logs, and persisted receipts. Do not print rendered Compose configuration or the raw token while diagnosing.

The pinned plugin's v0.2.4 [`Provision()` error path](https://raw.githubusercontent.com/caddy-dns/cloudflare/v0.2.4/cloudflare.go) includes the token in the initialization error.

### G070-C4 — P1: stop recovered app/edge services before secret retrieval

`infra/staging/remote/start.sh:34–44` captures the current state, then calls `compose_env pinned` before `stop_app_edge`. On a VM reboot, `restart: unless-stopped` may already have restarted the public edge and app. A slow or unavailable metadata/Secret Manager request can leave those services serving while recovery waits, contrary to the accepted Plan's stop/hold-first sequence and the script's own header.

After recording the observed state, stop and hold app and edge before calling `compose_env`. If secret retrieval or any later recovery check fails, verify they remain stopped and the HOLD receipt is written. Add a deterministic slow/failing secret-fetch regression for that ordering.

### G070-C5 — P2: carry G071's accepted root settings through staging configuration

G071 Plan r4 §6.1 requires G070 W2 to pass optional non-secret `DEFAULT_TENANT_SLUG` and `PLATFORM_ROOT_PAGE` through `staging.conf`/`remote.env` into the core service, with unset/empty semantics documented. `infra/staging/remote/lib.sh:35–39` omits both from `REMOTE_ENV_KEYS`; the staging env examples and `infra/compose/staging.yml` core environment also omit them. The G071 behavior therefore cannot be selected in the G070 deployment.

Add the two optional keys end to end, preserving unset/empty behavior, update the relevant examples/docs and static Compose/config assertions, and add a focused proof for both unset defaults and explicit values.

## Disposition and next gate

Resolve C1–C5 in the existing G070 implementation branch, preserve the failure/history record, and return a fresh exact-head source review. Do not start live/provider work or open the implementation MR before the security and recovery corrections are accepted. After the corrections, reconcile the candidate with any management-only staging publication without changing frozen evidence inputs; re-run only checks affected by those source changes, then continue the Plan's remaining hosted and live gates. This review is not source ACCEPT, MR approval, provider authorization, merge authority or final Goal ACCEPT.

## Follow-up source review — candidate `6b5474f4` (CORRECTION LOOP)

Date: 2026-10-27 UTC. Reviewed exact pushed head `6b5474f4b448b9ef50c133f360a1c51a442f6bca` on staging base `9f5a0b653649ef371e4b30deca786b33dbcf8e24`; no tests or provider calls were run by Codex. The updated Result preserves the offline-5 failure and offline-6, static-9 and static-10 outcomes. A separate read-only review independently confirmed the disposition below. Any subsequent uncommitted worktree edits are outside this exact-head review.

G070-C2 through C5 are closed in source. C1 remains open for two reasons:

1. `infra/staging/remote/lib.sh:584–600` still runs the runtime metadata probe with BusyBox `wget -T 3`. The preserved offline-5 evidence records exit 139 for that option. The helper treats any nonzero, empty-output status as DENIED except Docker CLI statuses 125–127, so a wget crash or other unexpected probe failure may falsely pass the live metadata check. The `8a9e3cfd` change removed `-T` only from the offline packet-proof client; it did not change the production helper. Remove the problematic option/use a verified outer timeout and make every unexpected probe exit UNPROVEN/HOLD. Add a focused helper regression for exit 139 with empty output, an expected denied response, and a reachable positive control.
2. The same `metadata_denied` helper creates throwaway containers with direct `docker run --rm` but no explicit G070 task/run ownership label or allocation/removal accounting. Add ownership labels and unique probe identity in the existing inventory/receipt, and verify cleanup on success and failure/cancellation. Auto-removal alone does not satisfy the repository's continuous resource-accountability requirement.

Disposition remains **CORRECTION LOOP; no implementation MR or live/provider action**. After the bounded C1 fixes and attributable focused checks, review a fresh exact head. Preserve the failed probe and all existing evidence.

## Follow-up source review — candidate `bc29d8ba` (CORRECTION LOOP)

Date: 2026-10-27 UTC. Reviewed exact `origin/g070` head `bc29d8baf6660ac11b2d87175dc927d1b6aaf943` against staging `6fc95ca77d6fb032123ef998e0bf332d81fcd4c7`. The executable fix is `1e80c1a5`; `bc29d8ba` adds the Result record. No tests, Docker actions, provider calls or edits to the candidate were performed by Codex. A separate read-only reviewer confirmed the disposition below.

The fail-closed classification part of C1 is closed: `metadata_denied` no longer uses `wget -T`; unexpected statuses including 139, 137 and Docker errors are UNPROVEN/HOLD; and METADATA-PROBE reports coverage for expected denial, reachable responses, abnormal exits and a failed tool control. The Result preserves the earlier offline-5 failure and reports offline-7 (16/16) and static-11 PASS at the executable head. These are coordinator-reported results; Codex did not rerun them.

One bounded P2 remains in the cleanup-accounting part of C1:

1. `infra/staging/remote/lib.sh:622,637` records only the generated name before `docker run`. The foreground `--rm` invocation removes the probe container without first recording its exact container ID and immutable image identity in the existing run inventory.
2. `probe_reap` (`lib.sh:587–590`) removes IDs returned by a run-label query without checking each candidate against a recorded container ID and its task/environment/run/name/image identity. It also does not reconcile a recorded abandoned probe from an earlier interrupted run before allocating another probe.

This does not meet `AGENTS.md`'s existing resource-accountability rule: record exact IDs/tags at allocation, verify label-discovered resource identity against the run inventory before removal, and reconcile after interruption or reboot. Close this within the existing G070 inventory/receipt flow, without another manifest/framework. Focused coverage should demonstrate exact container/image identity recorded before removal, verified-owner-only reaping, and safe handling of interrupted/stale or mismatched resources. Hold uncertain or unowned matches; preserve unrelated/shared resources. The current Result's event descriptions do not establish that allocation/removal identity chain.

**Disposition: CORRECTION LOOP.** Do not open the implementation MR or begin live/provider work yet. Preserve all current evidence and failures. The review was relayed to the verified coordinator as handoff `<redacted-uuid>`; its receipt is confirmed, but execution is unconfirmed, so no second delivery was attempted.

## Follow-up source review — candidate `8c0d2bf6` (CORRECTION LOOP)

Date: 2026-10-27 UTC. Reviewed exact `origin/g070` head `8c0d2bf6d64ae9a58793cff768b46a85742b3cfe` on staging base `a391c9fe7578bcde41c7244732a121169a718f19`. The executable fix is `4bdeb239`; `8c0d2bf6` adds the Result record. No tests, Docker actions, provider calls or candidate edits were performed by Codex. A separate read-only reviewer independently confirmed the findings.

The fail-closed classification, exact container/image capture, ownership labels, identity checking and stale-resource discovery are materially improved. The preserved local evidence for offline-8 and static-12 matches the Result's reported PASS outcomes. It also preserves the `probe-only-1` fixture failure and the passing rerun. These are verified existing evidence files; Codex did not rerun the checks.

Three bounded cleanup gates remain in `infra/staging/remote/lib.sh`:

1. `metadata_denied` (`:701–703`) sets `bad=1` when the initial `probe_reap` fails, then continues to allocate new probe containers. An unresolved cleanup must stop before any new resource-creating command. Existing held/mismatch fixtures assert eventual exit 1 and that the seeded resource was not deleted, but do not assert that no new `docker run` occurred.
2. `probe_record allocate` (`:607`) is unchecked. Since `metadata_denied` is called in `|| hold`, Bash errexit is suppressed through nested functions; an inventory write failure can still reach the Docker creator. Require successful inventory/CID-file setup before starting the creator, and fail closed if any required allocation record cannot be written. Add a focused write-failure fixture.
3. Successful stale-container reaping and stale-record closure (`:655–673`) leave `/srv/<env>/state/probe-cids/<name>` behind. Normal removal deletes its CID file, but recovery paths do not. Delete the CID file only after verified container absence and test both recovery paths; retain the append-only TSV history.

Keep the fixes in the existing G070 state/inventory flow. Preserve offline-8/static-12 and the earlier `probe-only-1` failure. No implementation MR or live/provider work is released.

**Disposition: CORRECTION LOOP.** This review was relayed to the verified coordinator as handoff `<redacted-uuid>`; the helper receipt is confirmed, but delivery/execution is unconfirmed. No automatic retry was made.

## Follow-up source review — candidate `556c0651` (CORRECTION LOOP)

Date: 2026-10-27 UTC. Reviewed exact `origin/g070` head `556c0651d699e9ff7cf8b31c697c6a966e336199`, whose executable source is `1a4701fa8bf5962850efd4fe6f86c3cd9dfbbe3c`, on staging base `8503d211980dd666e785d6752b4d7ce3f56382bf`. The latest commit appends the Result record. An independent read-only review confirmed the finding below. Codex ran no tests or Docker/provider actions and made no changes to the candidate.

The three prior cleanup corrections are present: preflight reconciliation now blocks new probes, the allocation row is checked before Docker starts, and cidfiles are removed only after verified container absence. The latest Result's exact-source `offline-9` record reports 16/16 PASS, including `METADATA-PROBE`; `static-13` reports `lint lint-base provenance` PASS. I inspected these existing records without rerunning them.

### G070-C1 residual — P1: stop allocating probes while a per-probe cleanup/accounting failure is unresolved

`infra/staging/remote/lib.sh:745–750` marks a failed tool-control probe UNPROVEN and continues to the next network. In the endpoint loop at `:759–767`, a lifecycle/accounting error such as `PROBE_RC=125` is also marked UNPROVEN but the loop proceeds to later endpoints and networks. `probe_exec` uses 125 for missing identity, failed inventory writes and unverified removal; the final `probe_reap` is not called until after all these loops. A probe container that remains after failed removal can therefore coexist with newly allocated probes.

This violates the existing cleanup rule: unresolved cleanup/accounting must block further resource-consuming work until reconciled. On any lifecycle/accounting failure, stop creating probes immediately; use the existing exact-owner reconciliation/cleanup path, and keep readiness held if absence and inventory state cannot be verified. Do not resume allocation while any prior probe remains unresolved. Add a focused regression with multiple networks and a targeted endpoint-removal failure; assert that no later `docker run` occurs before reconciliation. The current `removal-stuck` case uses one network and does not assert the subsequent run count, so it misses this continuation.

**Disposition: CORRECTION LOOP.** Keep the current candidate and evidence; return a new exact-head review after the bounded fix and focused proof. No implementation MR, live/provider action or implementation ACCEPT is released.

## Follow-up source review — candidate `79788616` (CORRECTION LOOP)

Date: 2026-10-27 UTC. Reviewed exact `origin/g070` head `797886164706a84848d68303f52065ae8182b46a`, whose executable source is `739baead00176c8c1cd8426eeea4f13d073b39d1`, on staging base `2884a74acbea77591e1f19709f79a067aa2f483f`. The latest commit appends the Result. A separate read-only reviewer confirmed the remaining finding. Codex ran no tests or Docker actions and made no candidate edits.

The prior per-probe stop finding is closed: both the tool-control and endpoint paths stop on `PROBE_RC=125`, call the exact-owner reaper, and return failure before another probe can run. The new two-network endpoint-removal fixture checks the two-run ceiling and that the second network is not probed; the mutation run fails when that stop is removed. Existing evidence records offline-10 at `739baead` as 16/16 PASS and static-14 (`lint lint-base provenance`, B-SCOPE against `2884a74a`) as PASS. Codex inspected these records without rerunning them. Read-only Docker listings found no G070 containers, networks or volumes; the named proof images and `tf-mirror` remain retained as the Result records.

### G070-C1 residual — P2: stale-record recovery can falsely close a record when cidfile deletion fails

In `infra/staging/remote/lib.sh:695–697`, after confirming no matching container, `probe_reap` writes the terminal `stale-closed` inventory event and then runs `rm -f "$SRV/state/probe-cids/$nm"` without checking its result. Because callers invoke `probe_reap` in conditional contexts, errexit does not make this failure fatal. The function then reports the stale record closed; subsequent recovery ignores it because only final `allocate`, `created` and `remove-failed` events are retried. A failed unlink can therefore leave a cidfile while the check reports cleanup success.

Only record `stale-closed` after cidfile deletion and absence are verified. If unlink fails, keep the record retryable, report unresolved cleanup and return failure so readiness remains held. Add a focused unlink-failure fixture asserting nonzero/held status, no terminal close event and successful retry once removal works. The existing stale-record case proves successful deletion but does not inject unlink failure. The normal `probe_remove` path already propagates unlink failure; this finding is specific to stale-record closure.

**Disposition: CORRECTION LOOP.** Preserve the exact-head evidence and prior failures. No implementation MR, live/provider action or implementation ACCEPT is released pending the focused fix and a fresh review.

## Final source review — candidate `ca916211` (ACCEPT for MR opening)

Date: 2026-10-27 UTC. Reviewed exact pushed source `ca916211c2a203c2a45fc7d6b675977f0f0d3145` with Result append `a50f23e7cf7e48f2f0df45b3856de09934269835`, based on staging `8ab7bb846c222ba39ff246f610a7a4315054d076`. The branch is clean and matches `origin/g070`; staging is an ancestor. A bounded independent reviewer reached the same disposition. Codex ran no tests or provider actions and made no candidate edits.

The remaining stale-cidfile finding is closed. In `probe_reap`, unlink is attempted and the path is verified absent before the terminal `stale-closed` row is written. An unlink failure emits an unresolved receipt step, leaves the allocate record retryable, and returns failure so preflight stops before any new probe. A later successful retry closes the record once. The new fixture simulates an unlink failure with a non-empty directory and asserts nonzero/HOLD, no terminal close, zero probe starts, then successful one-time closure after the obstruction is removed.

The Result records `offline-11` PASS (16/16) and `static-15` PASS (`lint lint-base provenance`, B-SCOPE against `8ab7bb84`, 43 paths, 21 required IDs). Codex inspected those existing receipts only. The source still uses bounded `timeout` rather than `wget -T`, and probe containers have explicit owner labels and identity accounting, closing the earlier source-review findings. A read-only Docker inventory found no G070 containers, networks or volumes.

**Disposition: ACCEPT for opening the implementation MR to staging.** Applicable hosted CI, Duo disposition and integration gates remain. This is not final Goal ACCEPT; no real pilot, paid allocation, DNS or provider action is released. Those remain subject to the Goal's separately recorded live-action conditions.

## Duo-fix checkpoint — source `9ce10e1f` (CORRECTION LOOP)

Date: 2026-10-27 UTC. Reviewed exact source `9ce10e1ffb732f7550c0d11bb1bf3763c7137566` and Result append `e7972b25e584ba38dfe03b4468b1274c7280979e`: commits `54248b11` (stop fallback), `0092e70b` (first-deploy retry), and `9ce10e1f` (exact environment names). This is the submitted Duo-fix checkpoint only; the in-progress runtime-interface writer is excluded. Supervisor and independent read-only review agree on the final finding below. Codex ran no tests/provider calls and made no candidate edits.

The Compose-stop fallback closes the original failure path and does not report STOPPED when the fallback fails. First-deploy recovery retries only an exactly empty markerless database and HOLDs on existing rows or an unreadable count; seed.sql is transactional. The exact-name design fixes the sibling-environment attribution problem while still discovering labelled strays. These changes are accepted in substance.

Existing `offline-12` records all 16 rows PASS and `static-16` records `lint lint-base provenance` PASS on the exact source (B-SCOPE against `836d58ed`, 43 paths, 21 required IDs). The Result preserves the diagnostic mutations as intended failures. Those receipts verify the executed fixtures, not real PostgreSQL XML or Google Cloud CLI behavior; those provider paths remain disclosed live-pilot checks.

### Residual — quote the regex in gcloud's filter language

`infra/staging/lib.sh:269` returns grouped expressions such as `^(staging-edge|staging-iap)$`. `live_names` at lines 339/341 and the dry-run `reconcile` paths at lines 354/356 interpolate these into `--filter "name ~ $(env_name_filter ...)"`. The shell passes one argument, but that argument contains no quotes around the regex for gcloud's own parser. The [official filter grammar](https://docs.cloud.google.com/sdk/gcloud/reference/topic/filters) states that unquoted literals end at `(` or `)`. Thus the generated grouped regex is not a single quoted filter value and the lifecycle's real resource-list/reconciliation commands can be rejected before proceeding.

The fake gcloud fixture strips `name ~ ` and sends the remainder directly to grep; it bypasses the grammar boundary and therefore accepts this invalid serialization. This is a source/documented-grammar finding, not an observed cloud failure.

Preserve inner quotes in the actual argument, for example `name ~ '^(staging-edge|staging-iap)$'`, at every affected real/dry-run callsite. Extend the existing fixture to reject unquoted grouped values and interpret the quoted value; retain the sibling-environment and stray-resource assertions. No new CLI installation or test framework is required.

**Disposition: CORRECTION LOOP, one bounded syntax/fixture correction.** The runtime writer may include this in its already-owned `lib.sh`/`check.sh` work, preserving one writer per file. Validate the affected `check.sh run --only LIFECYCLE-DRYRUN` row and changed-script ShellCheck, then include the correction and attributable evidence in the combined runtime source handback. Reuse unaffected passes; no full application/ci-controls campaign or separate new qualification run is added. MR !113 remains unmerged pending the combined source and applicable hosted/Duo gates. Live actions and current G070 budgets retain their existing limits.

## Combined runtime source review — candidate `9d563441` (CORRECTION LOOP)

Date: 2026-10-27 UTC. Reviewed exact `origin/g070` head `9d563441dcf73a1cc778ceb4dbe2502fdb971682`, with executable source `94ff4a8d988c4ec79ed7ccc83b11fe0d9744bb04`, based on staging `a9a2d9eb5367af55f8e3e23ef682837218221c0e`. The Result records offline-14 (16/16), static-17 (`lint lint-base provenance`) PASS and 19 intentional diagnostic mutations. Codex inspected these existing records without rerunning tests or taking provider actions. A bounded independent read-only reviewer reached the same correction disposition; no source edits were made.

The combined source closes the gcloud filter-language quoting issue: the actual filter argument now contains inner quotes, and the lifecycle fixture distinguishes quoted from unquoted grouped expressions. The coordinator's `94ff4a8d` correction also fixes the offline-13 `BASH_REMATCH`/`--override session_max_seconds` refusal and `SESSION_VAR` ShellCheck findings. The Result accurately preserves offline-13 as failed and offline-14 as the passing rerun.

### Residual 1 — P2: uncertain VM lookup can be mistaken for a first allocation

`infra/staging/lib.sh:219–226` sets `VMSTATE=ABSENT` for every failed `gcloud compute instances describe`, without distinguishing confirmed NOT_FOUND from authentication, permission, API or transient errors. The new `cmd_provision` path (`infra/staging/staging.sh:191–208`) treats ABSENT as first allocation and selects the configured default duration. If a lookup fails for an existing VM with a custom duration, provision can plan the full Terraform apply with the default instead of preserving the instance's observed value. With `--yes`, it can apply a duration change; the same false ABSENT result in retirement can skip the intended graceful VM-side stop.

Require positive confirmation that the exact VM is absent. Any uncertain lookup must be represented as an error/HOLD before a mutating plan proceeds; do not infer first allocation or a default session length from a failed read. Add a focused lifecycle regression proving an existing VM with a failed/uncertain read does not select the default or run a Terraform plan/apply, deploy, start or destroy path, while the confirmed-absent first-allocation case still works.

### Residual 2 — P2 validation gap: exercise the saved-plan JSON extractor

`infra/staging/lib.sh:413–428` adds the `tf show -json` → jq → TSV extraction used by `duration_scope_check` to decide whether the saved Terraform plan may be applied. The `LIFECYCLE-DRYRUN` fixtures currently replace `plan_scope_rows` itself with pre-extracted TSV (`infra/staging/check.sh:913`), so they exercise the classifier and its mutations but never run this new JSON extraction program.

Add a bounded offline fixture through the real extractor and classifier: accept a valid duration-only update, and refuse at least an additional changed attribute/resource or replacement. This is an uncovered validation boundary, not a demonstrated jq defect. Re-run the affected lifecycle row and ShellCheck for changed scripts; retain offline-14's unrelated unchanged passes and avoid a broad qualification campaign.

**Disposition: CORRECTION LOOP.** The two findings were sent to the verified G070 coordinator as handoff `<redacted-uuid>`; the helper receipt is confirmed, execution/acknowledgment is not. MR !113 remains unmerged. This review does not release final Goal ACCEPT, live/pilot, DNS or provider activity. No additional G072 planning or execution is released.

## Combined runtime source review — candidate `44aa9c74` (CORRECTION LOOP)

Date: 2026-10-27 UTC. Reviewed exact `origin/g070` Result head `44aa9c74496948420075066f9918b279496dc27c`, executable source `95269940418ae7941714ca75f46c25cc88cb2664`, based on staging `0e2cb275ff986242e3f0d435d7fe29e40b212743`. Codex did not run tests or provider actions. The existing Result records offline-15 (3 selected rows of 16, all passing), static-18 (`lint lint-base provenance`) PASS, and 11 diagnostic mutations failing as intended. An independent read-only review confirmed the finding below and the two submitted fixes.

The prior VM-absence finding is closed: `vm_read` accepts ABSENT only after an exact-name list positively confirms absence and otherwise returns UNREADABLE; focused fixtures cover the error and empty-status paths. The plan-extraction validation gap is also closed: `LIFECYCLE-DRYRUN` invokes the actual pinned jq `plan_scope_rows` program against Terraform-shaped JSON and then runs `duration_scope_check`, including unknown leaves and no-op filtering.

### Residual — existing VM with an unreadable session duration falls back to default

`infra/staging/lib.sh:281–285` in `session_var` initializes the configured default and replaces it only when `SESSION_HAVE` is numeric. If an existing VM's `vm_session` command succeeds but emits an empty or malformed duration, `status` and `retire` silently keep the default and proceed instead of HOLDing. Also, `IFS=$'\t' read` treats tabs as IFS whitespace and collapses consecutive tabs; output such as `RUNNING<TAB><TAB><terminationTimestamp>` shifts the timestamp into `SESSION_HAVE`. The current `session-unread-status` and `session-unread-retire` fixtures cover only a failing describe command, not successful malformed output.

Require a valid numeric duration whenever the VM exists, parse the response without collapsing an empty middle field, and add focused status/retire fixtures for an empty duration, a missing duration before a termination timestamp, and malformed/non-numeric duration. Each must HOLD before any Terraform plan or write; confirmed absence must continue to use the configured default.

**Disposition: CORRECTION LOOP.** This finding was sent to the verified G070 coordinator as handoff `<redacted-uuid>`; the helper receipt is confirmed, but acknowledgment/execution is not yet confirmed. MR !113 remains unmerged. No broad/full-suite or live/provider action is released.

## Session-duration correction review — candidate `cd2ef479` (ACCEPT for MR progression)

Date: 2026-10-27 UTC. Reviewed exact `origin/g070` Result head `cd2ef4799b78a9f3e1de87206c4e842e88427f12`, executable source `2ef1c46b3eabd6bde0dd3820bd28031f21136e1d`, based on staging `756df412daecd3079c7303601dd5cc2e03b29a0c`. Codex inspected the source diff and existing evidence without running tests or contacting providers. A bounded independent read-only review reached the same disposition.

The prior finding is closed: `session_var` now requires a whole positive-seconds value for an existing VM and HOLDs before Terraform planning when the query fails or the duration is missing/malformed; only a positively confirmed absent VM uses the configured default. `vm_session` splits the first output line by tab with parameter expansion, preserving empty middle fields rather than allowing an IFS-whitespace `read` to shift the termination timestamp into the duration.

Evidence matches the exact candidate: offline-16 selected 3 of 16 rows (`LIFECYCLE-DRYRUN`, `STAGING-GREP`, `SHELLCHECK-DIRECT`) and all passed; static-19 (`lint lint-base provenance`, B-SCOPE against `756df412`, 43 paths and 21 required IDs) passed; five session-duration mutations failed as intended. Fixtures cover empty, missing, nonnumeric, suffixed and zero durations for status/retire as scoped, plus direct parser examples. These are reviewed existing receipts, not Codex reruns. The real gcloud `value()` output when `maxRunDuration` is absent remains for the separately gated pilot.

**Disposition: ACCEPT for MR progression.** The coordinator may update/advance MR !113 at this source under its existing hosted-fast, Duo, rebase and integration gates. This is not merge authority, final Goal ACCEPT or release of live/pilot/provider actions.


## Rehearsal repair source review — candidate `6ffd53dd` (ACCEPT for MR progression)

Date: 2026-10-27 UTC. Reviewed exact `g070-closeout-1` head `6ffd53dd9`, based on staging `6cdb1ae0f`; the executable/source repair commits are `ae8b7cecd` and `e5f713e1e`. `git diff --check` is clean. The reviewed delta is limited to staging release/token handling, its focused control, `docs/foundation/STAGING.md`, and the append-only Result update. Codex reviewed the source and existing receipts; no tests or provider actions were rerun.

The earlier broad token-prefix check is corrected: `resolve_secrets` admits the legacy 40-character form or `cfut_`/`cfat_` plus 48 alphanumerics, matching the Cloudflare formats accepted by the pinned DNS plugin; unsupported `cfxx_` canaries are rejected before env rendering or Docker/Compose calls, with a fixed non-secret error. The recorded `CF-TOKEN-CANARY`, `SHELLCHECK-DIRECT`, `STAGING-GREP` and `LIFECYCLE-DRYRUN` receipts pass at `e5f713e1`. The release permission correction extracts only repository files under a normal umask, leaving mounted `minio-init` files readable; `umask-1` passed at `ae8b7cec`.

`l1-deploy-3` reports READY on candidate `2f5e7ae4` with rehearsal-only, intentionally untrusted certificates. This proves that rehearsal candidate, not the later narrowed validator in `e5f713e1`, and does not close the production-certificate, smoke, lifecycle, pilot, retirement, cleanup or final-acceptance gates.

**Disposition: ACCEPT for normal MR opening.** Preserve the `2f5e7ae4` rehearsal manifest and evidence. After merge, prepare a new manifest bound to the exact merged staging SHA while reusing the same seven already-pushed image IDs/digests; no image rebuild is needed because the image inputs are unchanged. Deploy the merged source SHA for the remaining live gates. Hosted fast checks and Duo disposition apply; auto-merge only after the existing staging-MR gates clear. This is not final G070 ACCEPT or separate permission for out-of-scope provider actions.


## L-5 start latency disposition — bounded source correction authorized

Date: 2026-10-27 UTC. MR !115 is merged as `717bc9f8`. The coordinator reports L-4 PASS (marker/certificate checks, unchanged plan and stop time at start + 7200 s); Codex has not independently inspected that run's receipt. L-5 is blocked: the planned 600 s session cannot cover a reported ~27-minute start, with metadata probes reportedly taking ~130 s per hung connection.

The merged `probe_cmd` already invokes BusyBox `timeout 5 wget`, so simply increasing the session window or changing the timeout without diagnosing the observed delay is not accepted. **Codex authorizes preparation of one bounded source correction within G070's existing metadata-boundary/start-readiness scope.** Preserve the 600 s minimum and 7200 s default, all three metadata endpoint checks on every project network, fail-closed REACHABLE/UNPROVEN handling, and exact-owner probe allocation/identity/ledger/reaping. If the fix can interrupt `docker run`, it must reconcile the exact cidfile/ledger identity and HOLD on uncertain cleanup. Do not retry the 600 s live start before source review.

Required next evidence: identify and reproduce the latency in the current probe path, then prove a bounded per-probe and total start-time on focused offline fixtures while retaining the security/cleanup invariants. Review the source before another live L-5. A source change goes through a normal repair MR and its hosted fast/Duo gates; after merge, bind the exact merged SHA to the same seven existing image IDs/digests without rebuilding. This ruling is not source ACCEPT, a change to the accepted Plan, or final G070 ACCEPT; if the implementation changes L-5's accepted sequence/criteria, append a Plan amendment for review before execution.


## L-5 metadata timeout fix source review — candidate `f5ff36f3` (ACCEPT for MR opening)

Date: 2026-10-27 UTC. Reviewed source commit `f5ff36f355d73eac5c0bdceed4d80db742caea99` against parent `057ba176a`; the Result is at `a5cdb70ba` on `g070-closeout-2`. The worktree is clean and the source diff is limited to `infra/staging/remote/lib.sh` and `infra/staging/check.sh`; `git diff --check` is clean. A bounded independent read-only review reached the same disposition.

`probe_cmd` adds Docker `--init`: docker-init becomes PID 1 and BusyBox `timeout 5` is an ordinary child, so the already-required timeout can deliver its signal. The metadata endpoint/network set, response handling, current 124/143 timeout classification, reachable/unproven decisions, no-output handling, ownership labels, cidfile, image identity, allocation/created/removed ledger, exact-ID removal, stale reaping and fail-closed stops are unchanged. `METADATA-PROBE` requires `--init` on all four normal-case probe containers; the existing control and status cases remain.

Existing evidence is attributable to the exact change: `probe-init-1` records `METADATA-PROBE`, `SHELLCHECK-DIRECT`, `STAGING-GREP` and `LIFECYCLE-DRYRUN` PASS on `f5ff36f3`; `probe-init-neg-2` records the expected `METADATA-PROBE` failure with only `--init` removed. The `l5-live-1` receipt reports rc 0, 53.9 seconds for 12 probes across the three staging networks, 12/12/12 allocate/created/removed events and no remaining probe containers/cidfiles. These are reviewed implementer receipts, not a Codex-run validation or independent VM call. The reported VM was RUNNING with its 7200 s expiry armed; Codex did not access or change the VM.

**Disposition: ACCEPT for normal repair-MR opening.** Run ordinary hosted fast/Duo review and follow the standing staging auto-merge gates. This evidence closes the probe-latency defect only; it does not prove a complete `start --duration 600` / L-5 pass. After merge, bind the exact merged source SHA to the same seven image IDs/digests without rebuilding, then rerun the accepted L-5 sequence. This is not final G070 ACCEPT or a new live-action release.
