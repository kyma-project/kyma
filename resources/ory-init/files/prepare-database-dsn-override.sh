#!/usr/bin/env bash

DB_PASSWORD=$(cat /etc/database/${DB_SECRET_KEY})
DSN="${DB_TYPE}://${DB_USER}:${DB_PASSWORD}@${DB_URL}/${DB_NAME}?sslmode=disable"\
ENCODED_DSN=$(echo "${DSN}" | tr -d '\n' | base64 -w 0)

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
  hydra.hydra.config.dsn: "${ENCODED_DSN}"
EOF