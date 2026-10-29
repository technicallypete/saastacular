#!/usr/bin/env bash
# VM side of `staging.sh status` and `staging.sh receipt` (docs/foundation/STAGING.md). Read-only: Docker labels and
# read queries only, no Compose invocation, so no env file and no secret is fetched. Prints key<TAB>value lines.
#
#   status.sh --env ENV [--dry-run]
#
# Exit: 0; 2 refused (wrong VM, no deploy yet).
set -euo pipefail
umask 077
# shellcheck source=lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

main() {
  local env="" out line
  while (($#)); do
    case $1 in
      --env) (($# >= 2)) || die 2 "--env needs a value"; env=$2; shift 2 ;;
      --dry-run) DRY_RUN=1; shift ;;
      *) die 2 "usage: status.sh --env ENV [--dry-run]" ;;
    esac
  done
  remote_init "$env"
  load_meta "$SRV/meta"
  vm_preflight
  use_release "$SRV/current"

  printf 'environment\t%s\ncompose_project\t%s\ncandidate\t%s\nrelease\t%s\n' "$ENV_NAME" "${RCONF[COMPOSE_PROJECT_NAME]}" \
    "${RCONF[CANDIDATE]}" "${REL##*/}"
  capture out docker ps --all --filter "label=com.docker.compose.project=$ENV_NAME" \
    --format '{{.Label "com.docker.compose.service"}} {{.State}} {{.Image}}'
  while IFS= read -r line; do [[ -z $line ]] || printf 'service\t%s\n' "${line// /$'\t'}"; done <<<"$out"
  runs_manifest_images || true
  marker_label
  printf 'marker\t%s\n' "${MARKER:-absent}"
  cid_of edge
  if ((DRY_RUN)) || [[ -n $CID ]]; then cert_lines "${RCONF[PLATFORM_BASE_DOMAIN]}" || printf 'cert\tunavailable\n'; fi
  if ((!DRY_RUN)) && [[ -f $SRV/state/secret-versions.tsv ]]; then
    while IFS=$'\t' read -r line out; do printf 'secret\t%s\t%s\n' "$line" "$out"; done <"$SRV/state/secret-versions.tsv"
  fi
  capture out docker volume ls --filter "label=com.docker.compose.project=$ENV_NAME" --format '{{.Name}}'
  while IFS= read -r line; do [[ -z $line ]] || printf 'volume\t%s\n' "$line"; done <<<"$out"
  capture out findmnt -no SOURCE,FSTYPE --target /var/lib/docker
  printf 'docker_root\t%s\n' "$out"
  capture out docker system df --format '{{.Type}}={{.TotalCount}}/{{.Size}}'
  printf 'docker_df\t%s\n' "${out//$'\n'/ }"
  if ((!DRY_RUN)); then
    printf 'env_file_present\t%s\n' "$([[ -e $ENV_FILE ]] && echo yes || echo no)"
    printf 'holds\t%s\n' "$(find "$SRV/state" -maxdepth 1 -name 'hold-*.tsv' 2>/dev/null | wc -l)"
  fi
  printf 'result\tSTATUS\n'
}

main "$@"
