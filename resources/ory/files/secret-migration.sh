#!/usr/bin/env bash

readonly SECRET_SYSTEM_KEY="secretsSystem"
readonly SECRET_COOKIE_KEY="secretsCookie"
readonly DSN_KEY="dsn"
readonly SERVICE_ACCOUNT_KEY="gcp-sa.json"
readonly NAMESPACE='{{ .Release.Namespace }}'
readonly TARGET_SECRET_NAME='{{ include "ory.fullname" . }}-hydra-credentials'
readonly LOCAL_SECRETS_DIR="/etc/secrets"
readonly DSN_OPTS="sslmode=disable&max_conn_lifetime=10s"

function get_from_file () {
  cat "${LOCAL_SECRETS_DIR}/${1}" 2> /dev/null || { >&2 echo "File ${1} not found. This value will be generated unless it is required!" && return 1; }
}

function get_from_file_or_die () {
  get_from_file "${1}" || { >&2 echo "error: ${1} is required but it does not exist! Exiting..." && exit 1; }
}

function generateRandomString () {
  cat /dev/urandom | LC_ALL=C tr -dc 'a-z0-9' | fold -w ${1} | head -n 1
}

function communicate_missing_override() {
  echo "${1} not provided via overrides. Looking for value in existing secrets..."
}

{{- if .Values.global.ory.hydra.persistence.enabled }}
  {{- if .Values.global.ory.hydra.persistence.postgresql.enabled }}
DB_TYPE="postgres"
DB_USER="{{ .Values.global.postgresql.postgresqlUsername }}"
DB_URL="ory-postgresql.{{ .Release.Namespace }}.svc.cluster.local:5432"
DB_NAME="{{ .Values.global.postgresql.postgresqlDatabase }}"
PASSWORD="{{ .Values.global.postgresql.postgresqlPassword }}"
PASSWORDR="{{ .Values.global.postgresql.replicationPassword }}"
PASSWORD_R_KEY="postgresql-replication-password"
  if [[ -z "${PASSWORDR}" ]]; then
	    communicate_missing_override "${PASSWORD_R_KEY}"
	      PASSWORDR=$(get_from_file "${PASSWORD_R_KEY}" || generateRandomString 10)
  fi
PASSWORD_KEY="postgresql-password"
if [[ -z "${PASSWORD}" ]]; then
  communicate_missing_override "${PASSWORD_KEY}"
  PASSWORD=$(get_from_file "${PASSWORD_KEY}" || generateRandomString 10)
fi
  {{- else }}
DB_TYPE="{{ .Values.global.ory.hydra.persistence.dbType }}"
DB_USER="{{ .Values.global.ory.hydra.persistence.user }}"
DB_URL="{{ .Values.global.ory.hydra.persistence.dbUrl }}"
DB_NAME="{{ .Values.global.ory.hydra.persistence.dbName }}"
PASSWORD="{{ .Values.global.ory.hydra.persistence.password }}"
PASSWORD_KEY="dbPassword"
if [[ -z "${PASSWORD}" ]]; then
  communicate_missing_override "${PASSWORD_KEY}"
  PASSWORD=$(get_from_file_or_die "${PASSWORD_KEY}")
fi
  {{- end }}
DSN=${DB_TYPE}://${DB_USER}:${PASSWORD}@${DB_URL}/${DB_NAME}?${DSN_OPTS}
{{- else }}
DSN=memory
{{- end }}

{{- if .Values.global.ory.hydra.persistence.gcloud.enabled }}
SERVICE_ACCOUNT="{{ .Values.global.ory.hydra.persistence.gcloud.saJson | b64enc }}"
if [[ -z "${SERVICE_ACCOUNT}" ]]; then
  communicate_missing_override "${SERVICE_ACCOUNT_KEY}"
  SERVICE_ACCOUNT=$(get_from_file_or_die "${SERVICE_ACCOUNT_KEY}" | base64 -w 0)
fi
{{- end }}

SYSTEM="{{ .Values.hydra.hydra.config.secrets.system }}"
if [[ -z "${SYSTEM}" ]]; then
  communicate_missing_override "${SECRET_SYSTEM_KEY}"
  SYSTEM=$(get_from_file "${SECRET_SYSTEM_KEY}" || generateRandomString 32)
fi

COOKIE="{{ .Values.hydra.hydra.config.secrets.cookie }}"
if [[ -z "${COOKIE}" ]]; then
  communicate_missing_override "${SECRET_COOKIE_KEY}"
  COOKIE=$(get_from_file "${SECRET_COOKIE_KEY}" || generateRandomString 32)
fi

DATA=$(cat << EOF
  ${DSN_KEY}: $(echo -n "${DSN}" | base64 -w 0)
  ${SECRET_SYSTEM_KEY}: $(echo -n "${SYSTEM}" | base64 -w 0)
  ${SECRET_COOKIE_KEY}: $(echo -n "${COOKIE}" | base64 -w 0)
  {{- if .Values.global.ory.hydra.persistence.enabled }}
  ${PASSWORD_KEY}: $(echo -n "${PASSWORD}" | base64 -w 0)
    {{- if .Values.global.ory.hydra.persistence.postgresql.enabled }}
  ${PASSWORD_R_KEY}: $(echo -n "${PASSWORDR}" | base64 -w 0)
    {{- end}}
  {{- end }}
  {{- if .Values.global.ory.hydra.persistence.gcloud.enabled }}
  ${SERVICE_ACCOUNT_KEY}: $(echo -n "${SERVICE_ACCOUNT}")
  {{- end }}
EOF
)

echo "Data to be applied:"
echo "${DATA}"

SECRET=$(cat <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: ${TARGET_SECRET_NAME}
  namespace: ${NAMESPACE}
  labels:
    app.kubernetes.io/name: {{ include "ory.name" . }}
type: Opaque
data:
${DATA}
EOF
)

echo "Applying database secret"
set -e
echo "${SECRET}" | kubectl apply -f -
