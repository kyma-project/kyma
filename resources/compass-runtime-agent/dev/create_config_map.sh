#!/usr/bin/env bash

# This config map is used by the Compass Runtime Agent it will be created by job in Compass chart

discoverUnsetVar=false

for var in TOKEN RUNTIME_ID; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [ "${discoverUnsetVar}" = true ] ; then
    exit 1
fi

DOMAIN=$(kubectl -n compass-system get vs compass-gateway -o 'jsonpath={.spec.hosts[0]}')

cat <<EOF | kubectl -n compass-system apply -f -
apiVersion: v1
data:
  CONNECTOR_URL: https://${DOMAIN}/connector/graphql
  RUNTIME_ID: ${RUNTIME_ID}
  TENANT: 3e64ebae-38b5-46a0-b1ed-9ccee153a0ae
  TOKEN: ${TOKEN}
kind: ConfigMap
metadata:
  name: compass-agent-configuration
  namespace: compass-system
EOF
