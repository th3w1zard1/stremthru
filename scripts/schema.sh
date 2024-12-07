#!/usr/bin/env bash

set -euo pipefail

echo_err() {
  >&2 echo "$@"
}

declare -r cmd="${1:-"help"}"
shift || true

declare atlas="atlas"

if ! type $atlas >/dev/null 2>&1; then
  atlas="${TMPDIR}/atlas"
fi
if ! type $atlas >/dev/null 2>&1; then
  echo_err "Trying to download Atlas..."
  curl -sSf https://atlasgo.sh | sh -s -- --community --no-install --output "${atlas}" --yes
  chmod u+x "${atlas}"
fi
if ! type $atlas >/dev/null 2>&1; then
  echo_err "Could not resolve: atlas"
  exit 1
fi

echo_err "$($atlas version)"
echo_err

declare db_url="${STREMTHRU_DATABASE_URI}"
declare schema_file="schema.hcl"

case "${db_url}" in
postgresql://*)
  schema_file="schema.postgres.hcl"
  ;;
esac

case "${cmd}" in
inspect)
  echo_err \$ atlas schema inspect -u "\${STREMTHRU_DATABASE_URI}" "${@}"
  echo_err
  $atlas schema inspect -u "${db_url}" "${@}"
  ;;
migrate)
  echo_err \$ schema apply -u "\${STREMTHRU_DATABASE_URI}" --file "${schema_file}" "${@}"
  echo_err
  $atlas schema apply -u "${db_url}" --file "${schema_file}" "${@}"
  ;;
help)
  cat <<EOF
./scripts/schema.sh <COMMAND>

COMMANDS:
  inspect
  migrate
EOF
  ;;
*)
  exit 1
  ;;
esac
