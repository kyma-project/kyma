#!/usr/bin/env bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

if [[ -z ${REQUIRED_ISTIO_VERSION} ]]; then
    echo "Please set REQUIRED_ISTIO_VERSION variable!"
    exit 1
fi

if [[ -z ${CONFIG_DIR} ]]; then
    CONFIG_DIR=${DIR}
fi

function require_istio_version() {
    local version
    version=$(kubectl -n istio-system get deployment istio-pilot -o jsonpath='{.spec.template.spec.containers[0].image}' | awk -F: '{print $2}')
    if [[ "$version" != ${REQUIRED_ISTIO_VERSION} ]]; then
        echo "Istio must be in version: $REQUIRED_ISTIO_VERSION!"
        exit 1
    fi
}

function require_istio_system() {
    kubectl get namespace istio-system >/dev/null
}

function require_mtls_disabled() {
    local mTLS=$(kubectl get meshpolicy default -o jsonpath='{.spec.peers}' --ignore-not-found=true | grep "mtls")
    if [[ ${mTLS} != "" ]]; then
        echo "mTLS must be disabled"
        exit 1
    fi
}


function run_all_patches() {
  echo "--> Patch resources"
  for f in $(find ${CONFIG_DIR} -name '*\.patch\.json' -maxdepth 1); do
    local type=$(basename ${f} | cut -d. -f2)
    local name=$(basename ${f} | cut -d. -f1)
    echo "    Patch $type $name from: ${f}"
    local patch=$(cat ${f})
    set +e
    local out
    out=$(kubectl patch ${type} -n istio-system ${name} --patch "$patch" --type json)
    local result=$?
    set -e
    echo "$out"
    if [[ ${result} -ne 0 ]] && [[ ! "$out" = *"NotFound"* ]]; then
        exit ${result}
    fi
  done
}

function remove_not_used() {
  echo "--> Delete resources"
  while read line; do
    echo "    Delete $line"
    local type=$(cut -d' ' -f1 <<< ${line})
    local name=$(cut -d' ' -f2 <<< ${line})
    set +e
    kubectl delete ${type} ${name} -n istio-system
    set -e
  done <${CONFIG_DIR}/delete
}

function configure_sidecar_injector() {
  echo "--> Configure sidecar injector"
  local configmap=$(kubectl -n istio-system get configmap istio-sidecar-injector -o jsonpath='{.data.config}')
  local alreadyEnabled=$(grep "policy: enabled" <<< "$configmap")
  if [[ -n ${alreadyEnabled} ]]; then
    # Disable automatic injecting
    configmap=$(sed 's/policy: enabled/policy: disabled/' <<< "$configmap")
    configmap=$(sed 's/\[\[ \.ProxyConfig\.ZipkinAddress \]\]/zipkin.kyma-system:9411/g' <<< "$configmap")

    # Set limits for sidecar. Our namespaces have resource quota set thus every container needs to have limits defined.
    # Add limits to already existing resources sections
    configmap=$(sed 's|    resources:|    resources:\n      limits: { memory: 128Mi, cpu: 100m }\n      requests: { memory: 128Mi, cpu: 10m }|' <<< "$configmap")
    # In case there is no limits section add one at the beginning of container definition. It serves as default.
    configmap=$(sed 's|  - name: istio-\(.*\)|  - name: istio-\1\'$'\n    resources: { limits: { memory: 128Mi, cpu: 100m }, requests: { memory: 128Mi, cpu: 10m } }|' <<< "$configmap")

    # Escape new lines and double quotes for kubectl
    configmap=$(sed -e ':a' -e 'N' -e '$!ba' -e 's/\n/\\n/g' <<< "$configmap")
    configmap=$(sed 's/"/\\"/g' <<< "$configmap")

    kubectl patch -n istio-system configmap istio-sidecar-injector --type merge -p '{"data": {"config":"'"$configmap"'"}}'
  fi
}

function check_requirements() {
  while read crd; do
    echo "    Require CRD crd $crd"
    kubectl get crd ${crd}
    if [[ $? -ne 0 ]]; then
        echo "Cannot find required CRD $crd"
    fi
  done <${CONFIG_DIR}/required-crds
}

require_istio_system
require_istio_version
require_mtls_disabled
check_requirements
configure_sidecar_injector
run_all_patches
remove_not_used
