set -e

export DOMAIN="${GLOBAL_DOMAIN}"
export IP="${GLOBAL_IP}"

if [[ -z "${ENV_TYPE}" || "${ENV_TYPE}" == "xip" ]]; then
    echo "---> xip.io env detected"
else
    echo "---> Other env already detected, exitting"
    exit 0
fi

if [ -z "${GLOBAL_IP}" ]; then
    echo "---> No GLOBAL_IP has been set, defaulting to xip.io"
    EXTERNAL_PUBLIC_IP=$(kubectl get service -n "${ISTIO_GATEWAY_NAMESPACE}" "${ISTIO_GATEWAY_NAME}" -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
    if [[ -z "$EXTERNAL_PUBLIC_IP" ]]; then
        echo "---> Could not get IP, exitting"
        exit 1
    fi
    DOMAIN="${EXTERNAL_PUBLIC_IP}.xip.io"
    IP="${EXTERNAL_PUBLIC_IP}"
fi

PATCH_YAML=$(cat << EOF
---
data:
  global.ingress.domainName: "${DOMAIN}"
  global.loadBalancerIP: "${IP}"
  global.isXipEnv: "true"
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
