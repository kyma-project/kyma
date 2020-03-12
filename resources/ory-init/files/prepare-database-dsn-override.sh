#!/usr/bin/env bash

if [[ -s /etc/database/${DB_SECRET_KEY} ]]; then
	DB_PASSWORD=$(cat /etc/database/${DB_SECRET_KEY})
	DSN="${DB_TYPE}://${DB_USER}:${DB_PASSWORD}@${DB_URL}/${DB_NAME}?sslmode=disable"
else
	DSN="memory"
fi

## create override
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: ory-dsn-override
  namespace: kyma-installer
  labels:
    installer: overrides
    component: ory
    kyma-project.io/installation: ""
data:
  hydra.hydra.config.dsn: "$(echo "${DSN}" | tr -d '\n' | base64 -w 0)"
EOF