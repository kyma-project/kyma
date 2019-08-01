#!/usr/bin/env bash

DOMAIN=$(kubectl -n compass-system get vs compass-gateway -o 'jsonpath={.spec.hosts[0]}')

cat <<EOF | kubectl -n compass-system apply -f -
apiVersion: v1
data:
  DIRECTOR_URL: $DOMAIN/director/graphql
  RUNTIME_ID: 854f2e79-3266-47c0-8ccb-8ed7b6b8f48f
  TENANT: 'demo1'
kind: ConfigMap
metadata:
  name: compass-agent-configuration
  namespace: compass-system
EOF
