#!/usr/bin/env bash

## This scripts sets DEFAULT scenario label on Runtime removing all other scenarios
## You can use this in case tests panicked and did not cleanup testing scenario from Runtime

discoverUnsetVar=false

for var in DEX_TOKEN; do
    if [[ -z "${!var}" ]] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done
if [[ "${discoverUnsetVar}" = true ]] ; then
    exit 1
fi

RUNTIME_ID="$(kubectl -n compass-system get cm compass-agent-configuration -o jsonpath='{.data.RUNTIME_ID}')"
TENANT="$(kubectl -n compass-system get cm compass-agent-configuration -o jsonpath='{.data.TENANT}')"
HOST="$(kubectl -n compass-system get vs compass-gateway -o 'jsonpath={.spec.hosts[0]}')"
URL="https://$HOST/director/graphql"

MUTATION='mutation {
  setRuntimeLabel(runtimeID:"'${RUNTIME_ID}'", key: "scenarios", value: ["DEFAULT"]) {
      key
	    value
  }
}'

MUTATION=$(echo $MUTATION | sed 's/"/\\"/g')

BODY="{
    \"query\": \"$MUTATION\"
}"

echo $BODY

curl -v -X POST ${URL} -H "Content-Type: application/json" -H "Authorization: Bearer $DEX_TOKEN" -H "Tenant: $TENANT" -d "$BODY" -k

