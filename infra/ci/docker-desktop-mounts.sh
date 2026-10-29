#!/usr/bin/env bash
# Docker Desktop mount-table check (docs/foundation/OPERATIONS.md, "Docker Desktop bind-mount leak").
# Docker Desktop on WSL keeps every bind-mounted host path mounted in its VM until WSL itself shuts down;
# deleting the directory, removing the container or restarting Docker Desktop does not release it. When
# the VM's mount table reaches fs.mount-max, every container start fails with "failed to mount ...
# no space left on device" although the disk has space. Run this before a long local validation batch.
#
#   bash infra/ci/docker-desktop-mounts.sh [PERCENT]
#
# PERCENT (default 80) is the share of fs.mount-max at which the check fails.
# Exit: 0 below PERCENT, or not Docker Desktop on WSL (nothing to check); 1 at or above PERCENT; 2 usage.
set -euo pipefail

pct="${1:-80}"
if [[ ! "$pct" =~ ^[1-9][0-9]?$ ]]; then echo "usage: docker-desktop-mounts.sh [PERCENT 1-99]" >&2; exit 2; fi
if command -v wsl.exe >/dev/null 2>&1; then :; else
  echo "docker-desktop-mounts: not running under WSL; skipped"; exit 0
fi
# wsl.exe warns when the current directory has no Windows path; run it from / and drop that warning.
got="$(cd / && timeout 30 wsl.exe -d docker-desktop -- sh -c 'wc -l </proc/mounts; cat /proc/sys/fs/mount-max' 2>/dev/null \
  | tr -d '\r')" || got=""
used="$(sed -n 1p <<<"$got")" max="$(sed -n 2p <<<"$got")"
if [[ ! "$used" =~ ^[0-9]+$ || ! "$max" =~ ^[1-9][0-9]*$ ]]; then
  echo "docker-desktop-mounts: the docker-desktop WSL distro is not readable; skipped"; exit 0
fi
share=$(( used * 100 / max ))
if (( share >= pct )); then
  echo "docker-desktop-mounts: $used of $max mounts in use ($share%, limit $pct%)." >&2
  echo "  Quit Docker Desktop, run 'wsl --shutdown' from Windows, then start Docker Desktop again." >&2
  echo "  This stops every WSL distro and container and clears /tmp; named volumes are kept." >&2
  exit 1
fi
echo "docker-desktop-mounts: $used of $max mounts in use ($share%)"
