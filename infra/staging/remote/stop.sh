#!/usr/bin/env bash
# VM side of `staging.sh stop` (docs/foundation/STAGING.md); see lib.sh for the layout, secret and access notes.
#
#   stop.sh --env ENV [--dry-run]
#
# `docker compose stop` of every service (PostgreSQL gets its fast shutdown; a stopped container is not restarted by
# unless-stopped), with the env rendered from the pinned secret versions. When that env cannot be rendered, or the
# Compose stop fails, the project's containers are stopped by their Compose labels instead (no interpolation needed),
# so PostgreSQL still gets its shutdown before the caller stops the VM. Containers, volumes, images and the edge's
# certificates stay; so does each container's injected environment in its configuration on the retained, encrypted
# data disk (VM-only; deleting the env file revokes nothing). The staging.sh caller then stops the VM.
# Sourcing this file (a test) defines main without running it.
#
# Exit: 0 stopped; 2 refused (wrong VM); 4 a container could not be listed or stopped by label (after the Compose stop
# failed or without a rendered env); nothing is reported STOPPED then.
set -euo pipefail
umask 077
# shellcheck source=lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

main() {
  local env="" rc
  while (($#)); do
    case $1 in
      --env) (($# >= 2)) || die 2 "--env needs a value"; env=$2; shift 2 ;;
      --dry-run) DRY_RUN=1; shift ;;
      *) die 2 "usage: stop.sh --env ENV [--dry-run]" ;;
    esac
  done
  remote_init "$env"
  load_meta "$SRV/meta"
  vm_preflight
  use_release "$SRV/current"

  rc=0
  compose_env pinned || rc=$?
  if ((rc == 0)) && dc stop --timeout 60; then
    say stop "every service stopped (compose stop)"
  elif ((rc == 0)); then
    say stop "compose stop failed; stopping by Compose labels"
    stop_by_label edge web intelligence core minio redis postgres ||
      die 4 "compose stop failed and a container could not be listed or stopped by label; check docker ps"
  else
    say stop "the pinned env could not be rendered; stopping by Compose labels"
    stop_by_label edge web intelligence core minio redis postgres || die 4 "a container could not be listed or stopped by label; check docker ps"
  fi
  ((DRY_RUN)) || rm -f "$ENV_FILE"
  say retained "database, objects, edge-data/edge-config and the containers (with their injected environment, on the encrypted data disk, VM-only)"
  printf 'result\tSTOPPED\n'
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then main "$@"; fi
