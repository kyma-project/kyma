#!/usr/bin/env bash

readonly SECRET_SYSTEM_KEY="secretsSystem"
readonly SECRET_COOKIE_KEY="secretsCookie"
readonly SERVICE_ACCOUNT_KEY="gcp-sa.json"
readonly NAMESPACE='{{ .Release.Namespace }}'
readonly TARGET_SECRET_NAME='{{ include "ory.fullname" . }}-hydra-credentials'
readonly LOCAL_SECRETS_DIR="/etc/secrets"

trap "echo 'error' && exit 1" TERM
export TOP_PID=$$

function get_from_file () {
  cat "${LOCAL_SECRETS_DIR}/${1}" 2> /dev/null || { >&2 echo "File ${1} not found!" && return 1; }
}

function get_from_file_or_die () {
  get_from_file "${1}" || { >&2 echo "${1} is required but it does not exist!" && kill -s TERM ${TOP_PID}; }
}

function generateRandomString () {
  cat /dev/urandom | LC_ALL=C tr -dc 'a-z0-9' | fold -w ${1} | head -n 1
}

function communicate_missing_override() {
  echo "${1} not provided via overrides. Looking for value in existing secrets..."
}

{{ if .Values.global.ory.hydra.persistence.postgresql.enabled }}
  PASSWORD=$(echo -n "{{ .Values.global.postgresql.postgresqlPassword }}" | base64 --decode)
  PASSWORD_KEY="postgresql-password"
  if [[ -z "${PASSWORD}" ]]; then
    communicate_missing_override "${PASSWORD_KEY}"
    PASSWORD=$(get_from_file "${PASSWORD_KEY}" || generateRandomString 10)
  fi
{{ else }}
  PASSWORD=$(echo -n "{{ .Values.global.ory.hydra.persistence.password }}"  | base64 --decode)
  PASSWORD_KEY="dbPassword"
  if [[ -z "${PASSWORD}" ]]; then
    communicate_missing_override "${PASSWORD_KEY}"
    PASSWORD=$(get_from_file_or_die "${PASSWORD_KEY}")
  fi
{{ end }}

{{ if .Values.global.ory.hydra.persistence.gcloud.enabled }}
  SERVICE_ACCOUNT=$(echo -n "{{ .Values.global.ory.hydra.persistence.gcloud.saJson }}" | base64 --decode)
  if [[ -z "${SERVICE_ACCOUNT}" ]]; then
    communicate_missing_override "${SERVICE_ACCOUNT_KEY}"
    SERVICE_ACCOUNT=$(get_from_file_or_die "${SERVICE_ACCOUNT_KEY}")
  fi
{{ end }}

SYSTEM=$(echo -n "{{ .Values.hydra.hydra.config.secrets.system }}" | base64 --decode)
if [[ -z "${SYSTEM}" ]]; then
  communicate_missing_override "${SECRET_SYSTEM_KEY}"
  SYSTEM=$(get_from_file "${SECRET_SYSTEM_KEY}" || generateRandomString 32)
fi

COOKIE=$(echo -n "{{ .Values.hydra.hydra.config.secrets.cookie }}" | base64 --decode)
if [[ -z "${COOKIE}" ]]; then
  communicate_missing_override "${SECRET_COOKIE_KEY}"
  COOKIE=$(get_from_file "${SECRET_COOKIE_KEY}" || generateRandomString 32)
fi

STRING_DATA=

SECRET=$(cat <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: "${TARGET_SECRET_NAME}"
  namespace: "${NAMESPACE}"
  labels:
{{ include "ory.labels" . | indent 4 }}
type: Opaque
stringData:
  "${PASSWORD_KEY}": "${PASSWORD}"
  "${SECRET_SYSTEM_KEY}": "${SYSTEM}"
  "${SECRET_COOKIE_KEY}": "${COOKIE}"
  {{ if .Values.global.ory.hydra.persistence.gcloud.enabled }}
  "${SERVICE_ACCOUNT_KEY}": "${SERVICE_ACCOUNT}"
  {{ end }}
EOF
)

echo "Applying database secret"
echo "${SECRET}" | kubectl apply -f - --dry-run -o yaml
echo "${SECRET}" | kubectl apply -f -
