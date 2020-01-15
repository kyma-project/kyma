#!/bin/bash
DB_PASSWORD=$(cat /etc/database/${DB_SECRET_KEY})
DSN="postgres://${DB_USER}:${DB_PASSWORD}@${DB_URL}/${DB_NAME}?sslmode=disable"

PATCH=$(cat << EOF
---
data:
  dsn: "$(echo "${DSN}" | tr -d '\n' | base64 -w 0)"
EOF
)

set +e
msg=$(kubectl patch secret "${HYDRA_SECRET_NAME}" --patch "${PATCH}" -n "${HYDRA_SECRET_NAMESPACE}" 2>&1)
status=$?
set -e
if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
  echo "$msg"
  exit $status
fi
echo "---> Secret ${HYDRA_SECRET_NAME} patched"