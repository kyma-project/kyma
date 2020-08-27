set -e

if [[ -z "${ENV_TYPE}" || "${ENV_TYPE}" == "gardener" ]]; then
    echo "---> Gardener env detected"
else
    echo "---> Other env already detected, exitting"
    exit 0
fi

export IS_GARDENER=false
export DOMAIN="${GLOBAL_DOMAIN}"

if [[ -n "$(kubectl -n ${SHOOT_INFO_NAMESPACE} get cm ${SHOOT_INFO_NAME} --ignore-not-found)" ]]; then
	echo "---> Gardener env detected"
	IS_GARDENER=true
	DOMAIN="$(kubectl -n ${SHOOT_INFO_NAMESPACE} get cm ${SHOOT_INFO_NAME} -o jsonpath='{.data.domain}')"
	if [[ -z "$DOMAIN" ]]; then
		echo "---> Could not get domain, exitting"
		exit 1
	fi
else
	echo "---> Gardener not detected, exitting"
	exit 0
fi

PATCH_YAML=$(cat << EOF
---
data:
  global.ingress.domainName: "${DOMAIN}"
  global.environment.type: "gardener"
  global.environment.gardener: "true"
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
  modules.gardener.enabled: "true"
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
