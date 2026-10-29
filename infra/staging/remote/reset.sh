#!/usr/bin/env bash
# VM side of `staging.sh reset --confirm <env>` (docs/foundation/STAGING.md); see lib.sh for the layout, secret and
# access notes.
#
#   reset.sh --env ENV --confirm ENV [--dry-run]
#
# The only data-discarding path: an explicit owner decision recorded with its run. It addresses exactly this VM's
# Compose project: `docker compose down` (containers and networks; no volume option) -> remove exactly
# <env>_postgres-data and <env>_minio-data -> stores up (first initialization creates the roles) -> migrate up ->
# seed (a new marker label) -> metadata boundary -> app and edge up -> readiness. edge-data and edge-config are never
# touched (the ACME account and certificates survive; no new issuance). Removing the containers also removes the
# injected environment they held; the next containers receive the pinned secret versions again.
#
# Exit: 0 ready; 2 refused before any change; 3 HOLD (a container network reaches the metadata server; app and edge
# stopped); 4 failed after the reset started.
set -euo pipefail
umask 077
# shellcheck source=lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

main() {
  local env="" confirm="" v rc
  while (($#)); do
    case $1 in
      --env | --confirm) (($# >= 2)) || die 2 "$1 needs a value" ;;&
      --env) env=$2; shift 2 ;;
      --confirm) confirm=$2; shift 2 ;;
      --dry-run) DRY_RUN=1; shift ;;
      *) die 2 "usage: reset.sh --env ENV --confirm ENV [--dry-run]" ;;
    esac
  done
  remote_init "$env"
  [[ $confirm == "$ENV_NAME" ]] || die 2 "reset needs --confirm $ENV_NAME"
  load_meta "$SRV/meta"
  [[ ${RCONF[COMPOSE_PROJECT_NAME]} == "$ENV_NAME" ]] || die 2 "COMPOSE_PROJECT_NAME is not $ENV_NAME"
  vm_preflight
  use_release "$SRV/current"
  compose_env pinned || die 2 "the pinned secret versions are unavailable; nothing was changed"

  say reset "discarding ${ENV_NAME}_postgres-data and ${ENV_NAME}_minio-data; edge-data and edge-config kept"
  dc down || die 4 "compose down failed"
  for v in "${ENV_NAME}_postgres-data" "${ENV_NAME}_minio-data"; do
    if ((DRY_RUN)) || volume_exists "$v"; then run docker volume rm "$v" || die 4 "could not remove $v"; fi
  done
  stores_up || die 4 "the stores did not become healthy after the reset"
  rc=0
  migrate_current || rc=$?
  ((rc == 0)) || die 4 "migrate failed after the reset"
  seed_dataset || die 4 "the seed failed after the reset"
  metadata_denied || hold "a container network reaches the metadata server or could not be probed (staging-metadata-guard)"
  app_up || die 4 "app and edge did not become healthy after the reset"
  readiness || die 4 "reset but not ready (see the steps above)"
}

main "$@"
