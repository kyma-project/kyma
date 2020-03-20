#!/usr/bin/env bash

trap "echo 'skrypt wyjebalo w ciul' && exit 1" TERM
export TOP_PID=$$

PASSWORD="${1}"
FILE="${2}"
COOKIE="${3}"
FILE2="${4}"

LOCAL_SECRETS_DIR="/Users/i348140/Desktop/dupa"

function get_from_file () {
  cat "${LOCAL_SECRETS_DIR}/${1}" 2> /dev/null || { >&2 echo "File ${1} not found!" && return 1; }
}

function get_from_file_or_die () {
  get_from_file "${1}" || { >&2 echo "${1} is required but it does not exist!" && kill -s TERM ${TOP_PID}; }
}

function get_from_file_or_generate_random () {
  local pwd=$(get_from_file "${1}" || cat /dev/urandom | LC_ALL=C tr -dc 'a-z0-9' | fold -w ${2} | head -n 1)
  echo -n "${pwd}" | base64
}

function communicate_missing_override() {
  echo "${1} not provided via overrides. Looking for value in existing secrets..."
}

if [[ -z "${PASSWORD}" ]]; then
  communicate_missing_override "${FILE}"
  PASSWORD=$(get_from_file_or_generate_random "${FILE}" 10)
fi


if [[ -z "${COOKIE}" ]]; then
  communicate_missing_override "${FILE2}"
  COOKIE=$(get_from_file_or_die "${FILE2}")
fi

SECRET=$(cat <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: tajne
  namespace: goat
type: Opaque
data:
  "pswKry": "${PASSWORD}"
  "cookieKey": "${COOKIE}"
EOF
)

echo "Applying database secret"
echo "${SECRET}" | kubectl apply -f - --dry-run -o yaml
echo "${SECRET}" | kubectl apply -f - 
