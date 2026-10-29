#!/usr/bin/env bash
# Creates the local .env from .env.example, filling every empty secret with a
# locally generated value. Usage, from anywhere: infra/local/configure-env.sh
#
# - Never overwrites an existing .env.
# - Secrets are 48 hexadecimal characters (URL-safe) from /dev/urandom.
# - .env is created with mode 600; values are never printed.
# - Optional secrets (the AI gateway key) stay empty.
# Requires only bash and coreutils.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
example="$root/.env.example"
target="$root/.env"
optional=" OPENROUTER_API_KEY "

if [[ -e "$target" || -L "$target" ]]; then
  echo "configure-env: .env already exists; leaving it unchanged." >&2
  echo "To start over, move it away first. Role passwords and DATABASE_NAME apply only to a new database volume." >&2
  exit 1
fi
[[ -f "$example" ]] || { echo "configure-env: .env.example not found" >&2; exit 1; }

umask 077
tmp="$(mktemp "$root/.env.tmp.XXXXXX")"
trap 'rm -f "$tmp"' EXIT
trap 'rm -f "$tmp"; exit 130' INT TERM HUP

generated=()
while IFS= read -r line || [[ -n "$line" ]]; do
  if [[ "$line" =~ ^([A-Z0-9_]+)=$ ]] && [[ "$optional" != *" ${BASH_REMATCH[1]} "* ]]; then
    printf '%s=%s\n' "${BASH_REMATCH[1]}" "$(head -c 24 /dev/urandom | od -An -tx1 | tr -d ' \n')"
    generated+=("${BASH_REMATCH[1]}")
  else
    printf '%s\n' "$line"
  fi
done <"$example" >"$tmp"
chmod 600 "$tmp"

# -n never replaces a .env that appeared meanwhile; the temporary file then
# remains and is removed on exit.
mv -n "$tmp" "$target"
if [[ -e "$tmp" ]]; then
  echo "configure-env: .env appeared while generating; leaving it unchanged." >&2
  exit 1
fi
trap - EXIT INT TERM HUP
echo "configure-env: created .env (mode 600) with generated values for: ${generated[*]}"
