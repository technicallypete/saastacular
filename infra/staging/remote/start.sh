#!/usr/bin/env bash
# VM side of `staging.sh start` (docs/foundation/STAGING.md); see lib.sh for the layout, secret and access notes.
#
#   start.sh --env ENV [--dry-run]
#
# Explicit resume after a stop, an expiry (Compute Engine's deferred stop: abrupt, no `compose stop`) or a reboot.
# Docker's restart policy (unless-stopped) may already have resumed containers before this runs; that autonomous
# restart is recorded as observed and is never recovery evidence. Sequence: record what is running -> stop app and
# edge by their Docker labels (no secret needed; a failed stop HOLDs) -> only then render the env from the pinned
# secret versions (a slow or failed fetch leaves app and edge stopped) -> Compose confirms the stop -> the PostgreSQL
# data volume must exist -> stores up (PostgreSQL crash recovery included) -> migrate status, pending migrations only
# -> the dataset marker must be present -> no container network reaches the metadata server -> only then app and edge
# up -> metadata boundary, HTTPS readiness, certificates, running digests, marker -> READY.
# Any failed check HOLDs: app and edge stay stopped, a HOLD receipt is written, nothing is reseeded or reset.
# Sourcing this file (a test) defines main without running it.
#
# Exit: 0 ready; 2 refused (wrong VM, no deploy yet); 3 HOLD.
set -euo pipefail
umask 077
# shellcheck source=lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

main() {
  local env="" observed line rc
  while (($#)); do
    case $1 in
      --env) (($# >= 2)) || die 2 "--env needs a value"; env=$2; shift 2 ;;
      --dry-run) DRY_RUN=1; shift ;;
      *) die 2 "usage: start.sh --env ENV [--dry-run]" ;;
    esac
  done
  remote_init "$env"
  load_meta "$SRV/meta"
  vm_preflight
  use_release "$SRV/current"

  capture observed docker ps --all --filter "label=com.docker.compose.project=$ENV_NAME" \
    --format '{{.Label "com.docker.compose.service"}} {{.State}} {{.Status}}'
  while IFS= read -r line; do [[ -z $line ]] || say observed-before-hold "$line"; done <<<"$observed"

  stop_by_label edge web intelligence core || hold "app and edge could not be stopped by label before the secrets were fetched"
  say app-edge-stopped "edge web intelligence core (by label, before any secret fetch)"
  compose_env pinned || hold "the pinned secret versions are unavailable or malformed (deploy again with valid versions)"
  stop_app_edge || hold "app and edge could not be stopped before the recovery checks"

  volume_exists "${ENV_NAME}_postgres-data" || hold "the PostgreSQL data volume is missing (no reseed)"
  stores_up || hold "the stores did not become healthy (PostgreSQL recovery or MinIO failed)"
  rc=0
  migrate_current || rc=$?
  ((rc == 0)) || hold "the migration status was inconsistent or migrate up failed"
  marker_label
  [[ -n $MARKER ]] || hold "existing dataset without its marker (no reseed; data is discarded only by reset --confirm)"
  say recovery-checks "passed: stores healthy, migrations current, marker $MARKER"
  metadata_denied || hold "a container network reaches the metadata server or could not be probed (staging-metadata-guard)"

  app_up || hold "app and edge did not become healthy"
  readiness || hold "not ready after start (see the steps above)"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then main "$@"; fi
