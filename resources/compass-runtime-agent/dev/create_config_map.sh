#!/usr/bin/env bash

# This config map is used by the Compass Runtime Agent it will be created by job in Compass chart

DOMAIN=$(kubectl -n compass-system get vs compass-gateway -o 'jsonpath={.spec.hosts[0]}')

cat <<EOF | kubectl -n compass-system apply -f -
apiVersion: v1
data:
  DIRECTOR_URL: https://$DOMAIN/director/graphql
  RUNTIME_ID: 854f2e79-3266-47c0-8ccb-8ed7b6b8f48f
  TENANT: "$TENANT"
kind: ConfigMap
metadata:
  name: compass-agent-configuration
  namespace: compass-system
EOF
