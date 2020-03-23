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

UNLABEL_RUNTIME_MUTATION='mutation {
  setRuntimeLabel(runtimeID:"'${RUNTIME_ID}'", key: "scenarios", value: ["DEFAULT"]) {
      key
	    value
  }
}'

UNLABEL_RUNTIME_MUTATION=$(echo $UNLABEL_RUNTIME_MUTATION | sed 's/"/\\"/g')

BODY="{
    \"query\": \"$UNLABEL_RUNTIME_MUTATION\"
}"

curl -v -X POST ${URL} -H "Content-Type: application/json" -H "Authorization: Bearer $DEX_TOKEN" -H "Tenant: $TENANT" -d "$BODY" -k


#REMOVE_SCENARIO='mutation {
#  result: updateLabelDefinition( in: {
#		key: "scenarios",
#		schema: "{\"type\":\"array\",\"minItems\":1,\"uniqueItems\":true,\"items\":{\"type\":\"string\",\"enum\":[\"DEFAULT\"]}}""",
#	} ) {
#    key
#    schema
#  }
#}'
#
#REMOVE_SCENARIO=$(echo $REMOVE_SCENARIO | sed 's/"/\\"/g')
#
#BODY="{
#    \"query\": \"$REMOVE_SCENARIO\"
#}"
#
#curl -v -X POST ${URL} -H "Content-Type: application/json" -H "Authorization: Bearer $DEX_TOKEN" -H "Tenant: $TENANT" -d "$BODY" -k
#
