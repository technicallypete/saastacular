#!/usr/bin/env bash
# Compute Engine startup script of one staging VM (G070, module gcp-staging-vm). Runs as root at every boot.
# The module prepends `export STAGING_ENV=<environment_name>` and stores the result in the instance's plain
# startup-script metadata key.
#
# It prepares the host only, idempotently:
# - the metadata guard, before Docker is installed (below);
# - the retained data disk (by-id device google-staging-data; ext4 filesystem label staging-data) is formatted
#   once when it carries no filesystem or partition table, listed in /etc/fstab and mounted at /var/lib/docker;
#   docker.service requires that mount, so Docker never starts on the boot disk;
# - the pinned Docker Engine and Compose plugin from Docker's apt repository (signing key fingerprint checked),
#   held against upgrades, with the classic image store so images also live on the data disk;
# - /srv/<env> (deployed files) and /run/<env> (the private env file directory), refusing unless /run/<env> is
#   on tmpfs (findmnt).
# It creates no secret, reads no Secret Manager value, and never pulls, deploys or starts the application.
#
# Metadata guard. The VM service account reads every staging secret in Secret Manager and the Artifact Registry
# repository, and any process that reaches the metadata server (169.254.169.254, fd20:ce::254,
# metadata.google.internal) can obtain its token: the token is a property of the VM, so no IAM narrowing keeps it
# from the containers. The host scripts need it (remote/lib.sh md()); no container does. The oneshot unit
# staging-metadata-guard.service runs /usr/local/sbin/staging-metadata-guard, which inserts DROP rules for both
# metadata addresses into the raw table's PREROUTING chain (comment staging-metadata-guard). raw/PREROUTING sees only
# packets that ARRIVE on an interface - container bridge traffic, forwarded or addressed to the host itself - before
# routing, NAT and Docker's own chains; host-originated packets take OUTPUT and are untouched. The rule depends on no
# Docker chain, so neither a Docker restart nor a restart-policy start can race it, and docker.service Requires= and
# is After= the unit (a drop-in): Docker never starts without the rules (fail closed). DNS (port 53) stays open:
# Docker's embedded resolver forwards a container's queries to the host's resolver, which on Compute Engine is the
# metadata address, from the container's namespace; DNS carries no credential. A missing IPv6 raw table (no
# ip6tables, IPv6 disabled) skips only the IPv6 rules; any IPv4 failure fails the unit. The unit is written, enabled
# and started before the Docker packages are installed, because the Docker package starts the daemon and a retained
# data disk brings back containers with a restart policy. iptables comes from Debian's archive (not version-pinned,
# unlike Docker). remote/lib.sh metadata_denied proves the boundary from every Compose network before app and edge
# are reported ready.
#
# Offline rehearsal of the package pins only (no disk, no directories, no tmpfs guard, guard files written but no
# systemctl and no rules), in debian:12:
#   bash -s -- --no-disk < startup.sh
# The guard rules alone, in the current network namespace (needs CAP_NET_ADMIN; the packet-level proof):
#   startup.sh --metadata-guard-rules
# Usage: startup.sh [--no-disk] [--env NAME] | --metadata-guard-rules   (--env overrides STAGING_ENV)
set -euo pipefail

DOCKER_VERSION=29.8.2
COMPOSE_VERSION=5.5.1
DOCKER_APT_SUFFIX="-1~debian.12~bookworm"
DOCKER_APT_FINGERPRINT=9DC858229FC7DD38854AE2D88D81803C0EBFCD88
DATA_DEVICE=/dev/disk/by-id/google-staging-data
DATA_LABEL=staging-data
DOCKER_ROOT=/var/lib/docker
GUARD_BIN=/usr/local/sbin/staging-metadata-guard
GUARD_UNIT=staging-metadata-guard.service

no_disk=0
guard_rules=0
env_name="${STAGING_ENV:-}"
work_dir=""
trap '[[ -z "$work_dir" ]] || rm -rf "$work_dir"' EXIT
while (($#)); do
  case "$1" in
    --no-disk) no_disk=1 ;;
    --metadata-guard-rules) guard_rules=1 ;;
    --env) env_name="${2:?--env needs a value}"; shift ;;
    *) echo "staging-startup: unknown argument $1" >&2; exit 2 ;;
  esac
  shift
done

log() { printf 'staging-startup: %s\n' "$*"; }
die() { printf 'staging-startup: ERROR: %s\n' "$*" >&2; exit 1; }

# metadata_guard_script: the body of /usr/local/sbin/staging-metadata-guard, the single source of the guard rules (the
# unit's ExecStart and --metadata-guard-rules both run it). Idempotent: each rule is checked (-C) before it is
# inserted (-I) and checked again afterwards.
metadata_guard_script() {
  cat <<'GUARD'
#!/bin/bash
# staging-metadata-guard: written by the G070 staging startup script at every boot (edit startup.sh, not this file).
# Drops container traffic to the Compute Engine metadata server in raw/PREROUTING; host traffic (OUTPUT) and DNS
# (port 53) are untouched. Any IPv4 failure is fatal; the IPv6 rules are skipped only without an IPv6 raw table.
set -euo pipefail
guard() { # guard iptables|ip6tables ADDRESS
  local proto
  for proto in tcp udp; do
    "$1" -w -t raw -C PREROUTING -d "$2" -p "$proto" ! --dport 53 -m comment --comment staging-metadata-guard -j DROP 2>/dev/null ||
      "$1" -w -t raw -I PREROUTING -d "$2" -p "$proto" ! --dport 53 -m comment --comment staging-metadata-guard -j DROP
    "$1" -w -t raw -C PREROUTING -d "$2" -p "$proto" ! --dport 53 -m comment --comment staging-metadata-guard -j DROP
  done
}
guard iptables 169.254.169.254/32
if ip6tables -w -t raw -S >/dev/null 2>&1; then
  guard ip6tables fd20:ce::254/128
  echo "staging-metadata-guard: 169.254.169.254 and fd20:ce::254 dropped in raw PREROUTING (tcp/udp except port 53)"
else
  echo "staging-metadata-guard: no IPv6 raw table (ip6tables missing or IPv6 disabled); IPv4 rules only"
fi
GUARD
}

if ((guard_rules)); then
  ((no_disk == 0)) || die "--metadata-guard-rules takes no other argument"
  bash -c "$(metadata_guard_script)" staging-metadata-guard </dev/null
  exit 0
fi

# shellcheck source=/dev/null
. /etc/os-release
[[ "${ID:-}" == debian && "${VERSION_CODENAME:-}" == bookworm ]] || die "Debian 12 (bookworm) required, found ${ID:-?} ${VERSION_CODENAME:-?}"
if ((no_disk == 0)); then
  [[ "$env_name" =~ ^[a-z][a-z0-9-]{1,20}$ ]] || die "STAGING_ENV (or --env) must match ^[a-z][a-z0-9-]{1,20}\$"
fi

prepare_data_disk() {
  local waited=0 fstype label pttype
  until [[ -b "$DATA_DEVICE" ]] || ((waited >= 60)); do
    sleep 1
    waited=$((waited + 1))
  done
  [[ -b "$DATA_DEVICE" ]] || die "data disk $DATA_DEVICE is not attached"

  fstype="$(blkid -p -o value -s TYPE "$DATA_DEVICE" || true)"
  if [[ -z "$fstype" ]]; then
    pttype="$(blkid -p -o value -s PTTYPE "$DATA_DEVICE" || true)"
    [[ -z "$pttype" ]] || die "data disk carries a $pttype partition table and no filesystem; refusing to format it"
    log "formatting the empty data disk (ext4, label $DATA_LABEL)"
    mkfs.ext4 -q -m 0 -L "$DATA_LABEL" -E lazy_itable_init=0,lazy_journal_init=0,discard "$DATA_DEVICE"
  fi
  fstype="$(blkid -p -o value -s TYPE "$DATA_DEVICE" || true)"
  label="$(blkid -p -o value -s LABEL "$DATA_DEVICE" || true)"
  [[ "$fstype" == ext4 && "$label" == "$DATA_LABEL" ]] || die "data disk holds '$fstype' labelled '$label', expected ext4 labelled $DATA_LABEL; refusing to mount it"

  if ! grep -qE "^LABEL=${DATA_LABEL}[[:space:]]" /etc/fstab; then
    printf 'LABEL=%s %s ext4 defaults,nofail,discard 0 2\n' "$DATA_LABEL" "$DOCKER_ROOT" >>/etc/fstab
  fi
  install -d -m 0755 /etc/systemd/system/docker.service.d
  printf '[Unit]\nRequiresMountsFor=%s\n' "$DOCKER_ROOT" >/etc/systemd/system/docker.service.d/10-data-disk.conf
  systemctl daemon-reload

  if ! mountpoint -q "$DOCKER_ROOT"; then
    if [[ -d "$DOCKER_ROOT" && -n "$(ls -A "$DOCKER_ROOT")" ]]; then
      die "$DOCKER_ROOT on the boot disk is not empty; refusing to hide it under the data disk"
    fi
    install -d -m 0710 "$DOCKER_ROOT"
    mount "$DOCKER_ROOT"
  fi
  [[ "$(findmnt -rno LABEL,FSTYPE --mountpoint "$DOCKER_ROOT")" == "$DATA_LABEL ext4" ]] || die "$DOCKER_ROOT is not the $DATA_LABEL ext4 filesystem"
  log "data disk mounted at $DOCKER_ROOT"
}

# install_metadata_guard: iptables, the guard script, its unit and the docker.service drop-in (Requires=, After=),
# rewritten at every boot; the unit is enabled and started, then the rules are re-asserted directly (restarting the unit
# would restart Docker through Requires=). --no-disk writes the files only (no systemd, no CAP_NET_ADMIN there).
install_metadata_guard() {
  if ! command -v iptables >/dev/null 2>&1 || ! command -v ip6tables >/dev/null 2>&1; then
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -q
    apt-get install -y -q --no-install-recommends iptables
  fi
  install -d -m 0755 /usr/local/sbin /etc/systemd/system/docker.service.d
  metadata_guard_script >"$GUARD_BIN.new"
  chmod 0755 "$GUARD_BIN.new"
  mv -f "$GUARD_BIN.new" "$GUARD_BIN"
  printf '%s\n' '[Unit]' 'Description=Drop container traffic to the Compute Engine metadata server (G070 staging)' \
    'Before=docker.service' '' '[Service]' 'Type=oneshot' 'RemainAfterExit=yes' "ExecStart=$GUARD_BIN" '' \
    '[Install]' 'WantedBy=multi-user.target' >"/etc/systemd/system/$GUARD_UNIT"
  printf '%s\n' '[Unit]' "Requires=$GUARD_UNIT" "After=$GUARD_UNIT" >/etc/systemd/system/docker.service.d/20-metadata-guard.conf
  if ((no_disk)); then
    log "rehearsal: metadata guard files written; systemctl and the rules skipped"
    return 0
  fi
  systemctl daemon-reload
  systemctl enable --now "$GUARD_UNIT" >/dev/null
  "$GUARD_BIN"
  log "metadata guard active before Docker"
}

pins_installed() {
  local dv cv
  command -v docker >/dev/null 2>&1 || return 1
  dv="$(docker --version 2>/dev/null || true)"
  cv="$(docker compose version --short 2>/dev/null || true)"
  [[ "$dv" == "Docker version $DOCKER_VERSION,"* && "${cv#v}" == "$COMPOSE_VERSION" ]]
}

# Classic image store: images and containers stay under /var/lib/docker (the data disk); bounded container logs.
write_daemon_config() {
  local want='{"features":{"containerd-snapshotter":false},"log-driver":"json-file","log-opts":{"max-size":"10m","max-file":"5"}}'
  install -d -m 0755 /etc/docker
  if [[ "$(cat /etc/docker/daemon.json 2>/dev/null || true)" != "$want" ]]; then
    printf '%s\n' "$want" >/etc/docker/daemon.json
    if ((no_disk == 0)) && systemctl is-active --quiet docker; then
      log "daemon.json changed while Docker runs; it applies at the next Docker restart"
    fi
  fi
}

install_docker() {
  local key fpr arch
  if pins_installed; then
    log "Docker $DOCKER_VERSION and Compose $COMPOSE_VERSION already installed"
    return 0
  fi
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -q
  apt-get install -y -q --no-install-recommends ca-certificates curl gnupg

  work_dir="$(mktemp -d)"
  key="$work_dir/docker.asc"
  install -d -m 0700 "$work_dir/gnupg"
  curl -fsSL https://download.docker.com/linux/debian/gpg -o "$key"
  fpr="$(GNUPGHOME="$work_dir/gnupg" gpg --batch --show-keys --with-colons "$key" | awk -F: '$1 == "fpr" { print $10; exit }')"
  [[ "$fpr" == "$DOCKER_APT_FINGERPRINT" ]] || die "Docker apt key fingerprint '$fpr' differs from $DOCKER_APT_FINGERPRINT"
  install -d -m 0755 /etc/apt/keyrings
  install -m 0644 "$key" /etc/apt/keyrings/docker.asc

  arch="$(dpkg --print-architecture)"
  printf 'deb [arch=%s signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian bookworm stable\n' "$arch" \
    >/etc/apt/sources.list.d/docker.list
  apt-get update -q
  apt-get install -y -q --no-install-recommends \
    "docker-ce=5:${DOCKER_VERSION}${DOCKER_APT_SUFFIX}" \
    "docker-ce-cli=5:${DOCKER_VERSION}${DOCKER_APT_SUFFIX}" \
    "docker-compose-plugin=${COMPOSE_VERSION}${DOCKER_APT_SUFFIX}" \
    containerd.io
  apt-mark hold docker-ce docker-ce-cli docker-compose-plugin containerd.io >/dev/null
  pins_installed || die "installed Docker/Compose do not report $DOCKER_VERSION/$COMPOSE_VERSION"
  log "installed Docker $DOCKER_VERSION and Compose $COMPOSE_VERSION"
}

prepare_runtime_dirs() {
  local fs
  install -d -m 0750 -o root -g root "/srv/$env_name"
  install -d -m 0700 -o root -g root "/run/$env_name"
  fs="$(findmnt -no FSTYPE --target "/run/$env_name")"
  [[ "$fs" == tmpfs ]] || die "/run/$env_name is on '$fs', not tmpfs; the private env file must stay on tmpfs"
  log "/srv/$env_name ready; /run/$env_name is tmpfs (mode 0700)"
}

# Every step reads /dev/null, never this script's own stdin (the rehearsal feeds the script on stdin).
if ((no_disk)); then
  install_metadata_guard </dev/null
  write_daemon_config </dev/null
  install_docker </dev/null
  log "rehearsal: $(docker --version); compose $(docker compose version --short)"
  exit 0
fi

prepare_data_disk </dev/null
install_metadata_guard </dev/null
write_daemon_config </dev/null
install_docker </dev/null
systemctl enable --now docker </dev/null >/dev/null
[[ "$(docker info --format '{{.DockerRootDir}}' </dev/null)" == "$DOCKER_ROOT" ]] || die "Docker root is not $DOCKER_ROOT"
prepare_runtime_dirs </dev/null
log "ready: Docker $DOCKER_VERSION, Compose $COMPOSE_VERSION, root $DOCKER_ROOT on $DATA_LABEL; application not started"
