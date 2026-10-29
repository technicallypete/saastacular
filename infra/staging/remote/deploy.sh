#!/usr/bin/env bash
# VM side of `staging.sh deploy` (docs/foundation/STAGING.md); see lib.sh for the layout, secret and access notes.
#
#   deploy.sh --env ENV --incoming DIR [--pin-secret NAME=VERSION|latest|none]... [--dry-run]
#
# DIR holds release/ (the candidate's infra subtree, git archive) and meta/ (remote.env, manifest.tsv), uploaded by
# staging.sh. Sequence: promote the release and meta -> resolve the Secret Manager versions (latest unless pinned; an
# optional secret without an enabled version is unset; the identity triple all or none) and record their numbers ->
# render the private env -> pull the manifest digests -> stores up -> migrate status; app and edge stop when a
# migration or an image change is pending; pending migrations only -> seed when the PostgreSQL data volume did not
# exist before this run's first `up`; an existing dataset with its marker is never seeded; one without its marker is
# seeded again only while it holds no application data (an interrupted first deploy: migrate up or the seed failed
# after the volume was created; seed.sql is idempotent), and HOLDs when any table holds a row or the count cannot be
# read (never a silent reseed) -> no existing container network reaches the metadata server -> app and edge up ->
# metadata boundary from every network (a first deploy creates the edge and inference networks only now; a failure
# HOLDs with app and edge stopped), HTTPS readiness, certificates, running digests = manifest, marker. Never down,
# never migrate down or reapply. A re-run with the same candidate converges: after a completed seed without reseeding;
# after an interrupted first deploy only while no application data exists (otherwise reset --confirm). A malformed
# CF_API_TOKEN is refused before anything changes. Sourcing this file (a test) defines main without running it.
#
# Rotation: add a Secret Manager version, deploy the same candidate (Compose recreates the containers whose
# environment changed and removes the replaced ones), then disable the superseded version. G057 cleanup restores the
# earlier identity configuration by pinning those versions (or none) here BEFORE disabling the versions it added:
# disabling a version alone does not revoke a value already injected into a container.
#
# Exit: 0 ready; 2 refused before any change; 3 HOLD (app and edge left stopped); 4 failed after a change started.
set -euo pipefail
umask 077
# shellcheck source=lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

# promote INCOMING: the release becomes /srv/<env>/releases/<sha12> (an existing one must be byte-identical), meta
# replaces /srv/<env>/meta, current points at the release; releases other than current and its predecessor are pruned.
promote() {
  local in=$1 sha12=${RCONF[CANDIDATE]:0:12} have want old d
  if ((DRY_RUN)); then
    printf '[vm] (promote %s/release to %s/releases/<sha12>, %s/meta to %s/meta; %s/current -> releases/<sha12>)\n' "$in" "$SRV" "$in" "$SRV" "$SRV"
    return 0
  fi
  [[ -d $in/release/infra/compose && -f $in/release/infra/staging/seed.sql ]] || die 2 "$in/release is not a staging release"
  install -d -m 0750 "$SRV" "$SRV/releases" "$SRV/meta" "$SRV/state"
  if [[ -d $SRV/releases/$sha12 ]]; then
    want="$(cd "$in/release" && find . -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum | sha256sum)"
    have="$(cd "$SRV/releases/$sha12" && find . -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum | sha256sum)"
    [[ $want == "$have" ]] || die 2 "releases/$sha12 exists with different content"
    rm -rf "$in/release"
  else
    mv "$in/release" "$SRV/releases/$sha12"
  fi
  install -m 0640 "$in/meta/remote.env" "$SRV/meta/remote.env.new"
  install -m 0640 "$in/meta/manifest.tsv" "$SRV/meta/manifest.tsv.new"
  mv -f "$SRV/meta/remote.env.new" "$SRV/meta/remote.env"
  mv -f "$SRV/meta/manifest.tsv.new" "$SRV/meta/manifest.tsv"
  old="$(readlink "$SRV/current" 2>/dev/null || true)"
  ln -sfn "releases/$sha12" "$SRV/current.new"
  mv -Tf "$SRV/current.new" "$SRV/current"
  rm -rf "$in"
  for d in "$SRV"/releases/*; do
    [[ -d $d && $d != "$SRV/releases/$sha12" && $d != "$SRV/${old:-none}" ]] || continue
    rm -rf "$d"
  done
  say promote "release $sha12 (candidate ${RCONF[CANDIDATE]}); previous ${old:-none}"
}

main() {
  local env="" incoming="" name rc fresh
  while (($#)); do
    case $1 in
      --env | --incoming | --pin-secret) (($# >= 2)) || die 2 "$1 needs a value" ;;&
      --env) env=$2; shift 2 ;;
      --incoming) incoming=$2; shift 2 ;;
      --pin-secret)
        [[ $2 =~ ^([A-Z][A-Z0-9_]*)=([0-9]+|latest|none)$ ]] || die 2 "--pin-secret takes NAME=VERSION|latest|none"
        name=${BASH_REMATCH[1]}
        [[ " ${SECRETS_REQUIRED[*]} ${SECRETS_OPTIONAL[*]} " == *" $name "* ]] || die 2 "--pin-secret: unknown secret $name"
        PIN[$name]=${BASH_REMATCH[2]}
        shift 2
        ;;
      --dry-run) DRY_RUN=1; shift ;;
      *) die 2 "usage: deploy.sh --env ENV --incoming DIR [--pin-secret NAME=VERSION|latest|none]... [--dry-run]" ;;
    esac
  done
  remote_init "$env"
  [[ -n $incoming ]] || die 2 "missing --incoming"
  load_meta "$incoming/meta"
  vm_preflight

  # The seeding rule (Plan r1.3): decided before any `up`, from the volume's existence; an existing volume without its
  # marker is seeded again only while it holds no application data (below).
  fresh=0
  if ((DRY_RUN)); then
    volume_exists "${ENV_NAME}_postgres-data" || true
    fresh=1
  elif ! volume_exists "${ENV_NAME}_postgres-data"; then
    fresh=1
  fi
  say dataset "postgres-data existed before this run: $( ((fresh)) && echo no || echo yes)"

  compose_env deploy || die 2 "secret versions unavailable; nothing was changed"
  promote "$incoming"
  write_secret_versions
  use_release "$SRV/current"
  pull_images
  stores_up || die 4 "the stores did not become healthy"
  rc=0
  migrate_current || rc=$?
  ((rc != 1)) || hold "the migration status could not be read consistently"
  ((rc == 0)) || die 4 "stopping app/edge or migrate up failed; app and edge may be stopped"
  if ((fresh)); then
    seed_dataset || die 4 "the seed failed on the fresh dataset"
  else
    marker_label
    if [[ -z $MARKER ]]; then
      # An interrupted first deploy (migrate up or the seed failed after the volume was created) left no marker and
      # no application data: seed.sql is idempotent, so the seed is retried. Any row, or a count that cannot be
      # read, HOLDs.
      rc=0
      dataset_empty || rc=$?
      ((rc != 2)) || hold "existing dataset without its marker whose application data could not be counted (no reseed; data is discarded only by reset --confirm)"
      ((rc == 0)) || hold "existing dataset without its marker (no reseed; data is discarded only by reset --confirm)"
      say dataset "existing postgres-data without its marker and without application data (interrupted first deploy): retrying the seed"
      seed_dataset || die 4 "the seed failed on the existing dataset without application data"
    fi
  fi
  metadata_denied || hold "a container network reaches the metadata server or could not be probed (staging-metadata-guard)"
  app_up || die 4 "app and edge did not become healthy"
  readiness || die 4 "deployed but not ready (see the steps above)"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then main "$@"; fi
