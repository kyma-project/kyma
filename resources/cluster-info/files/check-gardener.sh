set -ex

CERT_TYPE=$(cat ${CONFIG_MAP_DIR}/global.certificates.type)
echo $CERT_TYPE " --- " $MANAGER_ENABLED

GARDEN_CERTS=false
if kubectl api-versions | grep -c cert.gardener.cloud ; then
  echo "--> Gardener Certificate CR present"
  GARDEN_CERTS=true
fi

if [[ $GARDEN_CERTS && ! $MANAGER_ENABLED ]]; then
  echo "--> Gardener CR found, manager not requested. Use gardener certs and stop"
  PATCH_YAML=$(cat << EOF
---
data:
  global.certificates.type: "gardener"
EOF
)
fi

if [[ ! $GARDEN_CERTS && $MANAGER_ENABLED ]]; then
  echo "--> Gardener certs are not present, manager requested. Enable cert-manager and move one"
fi

# PATCH_YAML=$(cat << EOF
# ---
# data:
#   global.ingress.domainName: "${DOMAIN}"
#   global.apiserver.domainName: "${DOMAIN}"
#   global.environment.type: "gardener"
#   global.environment.gardener: "true"
# EOF
# )

# echo "---> Patching cm ${CLUSTER_INFO_NS}/${CLUSTER_INFO_CM}"
# set +e
# msg=$(kubectl patch cm ${CLUSTER_INFO_CM} --patch "${PATCH_YAML}" -n ${CLUSTER_INFO_NS} 2>&1)
# status=$?
# set -e

# if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
#     echo "$msg"
#     exit $status
# fi

# PATCH_YAML=$(cat << EOF
# ---
# data:
#   modules.manager.enabled: "false"
#   modules.gardener.enabled: "true"
# EOF
# )

# echo "---> Patching cm kyma-installer/certificates-overrides"
# set +e
# msg=$(kubectl patch cm certificates-overrides --patch "${PATCH_YAML}" -n kyma-installer 2>&1)
# status=$?
# set -e

# if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
#     echo "$msg"
#     exit $status
# fi
