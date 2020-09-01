set -e

export DOMAIN="${GLOBAL_DOMAIN}"
export IP="${GLOBAL_IP}"

if [[ -z "${ENV_TYPE}" || "${ENV_TYPE}" == "xip" ]]; then
    echo "---> xip.io env detected"
else
    echo "---> Other env already detected, exitting"
    exit 0
fi

PATCH_YAML=$(cat << EOF
---
data:
  global.ingress.domainName: "${DOMAIN}"
  global.loadBalancerIP: "${IP}"
  global.environment.xip: "true"
EOF
)

echo "---> Patching cm ${CLUSTER_INFO_NS}/${CLUSTER_INFO_CM}"
set +e
msg=$(kubectl patch cm ${CLUSTER_INFO_CM} --patch "${PATCH_YAML}" -n ${CLUSTER_INFO_NS} 2>&1)
status=$?
set -e

if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
    echo "$msg"
    exit $status
fi

PATCH_YAML=$(cat << EOF
---
data:
  modules.manager.enabled: "true"
  modules.xip.enabled: "true"
EOF
)

echo "---> Patching cm kyma-installer/certificates-overrides"
set +e
msg=$(kubectl patch cm certificates-overrides --patch "${PATCH_YAML}" -n kyma-installer 2>&1)
status=$?
set -e

if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
    echo "$msg"
    exit $status
fi
