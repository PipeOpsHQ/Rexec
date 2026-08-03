#!/usr/bin/env bash
# Register / update pipeopshq/rexec on Packagist via API.
#
# Prerequisites (one-time, from https://packagist.org/profile/):
#   export PACKAGIST_USERNAME=your-username
#   export PACKAGIST_TOKEN=your-api-token
#
# Usage:
#   ./scripts/register-packagist.sh              # create package if missing
#   ./scripts/register-packagist.sh --update      # force update/crawl
set -euo pipefail

REPO_URL="${PACKAGIST_REPO_URL:-https://github.com/PipeOpsHQ/rexec-php}"
PACKAGE="${PACKAGIST_PACKAGE:-pipeopshq/rexec}"

if [[ -z "${PACKAGIST_USERNAME:-}" || -z "${PACKAGIST_TOKEN:-}" ]]; then
  echo "Set PACKAGIST_USERNAME and PACKAGIST_TOKEN from https://packagist.org/profile/"
  echo "Then re-run: $0"
  exit 1
fi

AUTH="username=${PACKAGIST_USERNAME}&apiToken=${PACKAGIST_TOKEN}"

create_package() {
  echo "Creating Packagist package from ${REPO_URL}..."
  RESP=$(curl -sS -X POST \
    -H 'Content-Type: application/json' \
    "https://packagist.org/api/create-package?${AUTH}" \
    -d "{\"repository\":{\"url\":\"${REPO_URL}\"}}")
  echo "$RESP"
  echo "$RESP" | grep -qi 'success\|already' && return 0
  # "Package already exists" is fine
  if echo "$RESP" | grep -qi 'already exists\|Package exists'; then
    echo "Package already registered."
    return 0
  fi
  return 1
}

update_package() {
  echo "Triggering Packagist update for ${PACKAGE}..."
  RESP=$(curl -sS -X POST \
    -H 'Content-Type: application/json' \
    "https://packagist.org/api/update-package?${AUTH}" \
    -d "{\"repository\":{\"url\":\"${REPO_URL}\"}}")
  echo "$RESP"
}

if [[ "${1:-}" == "--update" ]]; then
  update_package
else
  create_package || update_package
fi

echo
echo "Verify: https://packagist.org/packages/${PACKAGE}"
curl -sS -o /dev/null -w "packagist HTTP %{http_code}\n" "https://packagist.org/packages/${PACKAGE}.json"
