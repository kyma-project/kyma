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

function require_mtls_enabled() {
    # TODO: rethink how that should be done
    local mTLS=$(kubectl get meshpolicy default -o jsonpath='{.spec.peers[0].mtls.mode}')
    if [[ "${mTLS}" != "STRICT" ]] && [[ "${mTLS}" != "" ]]; then
        echo "mTLS must be \"STRICT\""
        exit 1
    fi
}

function configure_policy_checks(){
  echo "--> Enable policy checks if not enabled"
  local istioConfigmap="$(kubectl -n istio-system get cm istio -o jsonpath='{@.data.mesh}')"
  local policyChecksDisabled=$(grep "disablePolicyChecks: true" <<< "$istioConfigmap")
  if [[ -n ${policyChecksDisabled} ]]; then
    istioConfigmap=$(sed 's/disablePolicyChecks: true/disablePolicyChecks: false/' <<< "$istioConfigmap")

    # Escape commented escaped newlines
    istioConfigmap=$(sed 's/\\n/\\\\n/g' <<< "$istioConfigmap")

    # Escape new lines and double quotes for kubectl
    istioConfigmap=$(sed -e ':a' -e 'N' -e '$!ba' -e 's/\n/\\n/g' <<< "$istioConfigmap")
    istioConfigmap=$(sed 's/"/\\"/g' <<< "$istioConfigmap")

    set +e
    local out
    out=$(kubectl patch -n istio-system configmap istio --type merge -p '{"data": {"mesh":"'"$istioConfigmap"'"}}')
    local result=$?
    set -e
    echo "$out"
    if [[ ${result} -ne 0 ]] && [[ ! "$out" = *"not patched"* ]]; then
      exit ${result}
    fi
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
    if [[ ${result} -ne 0 ]] && [[ ! "$out" = *"NotFound"* ]] && [[ ! "$out" = *"not patched"* ]]; then
        exit ${result}
    fi
  done
}

function remove_not_used() {
  echo "--> Delete resources"
  while read line; do
    if [[ "$line" == "" ]]; then
        continue
    fi
    echo "    Delete $line"
    local type=$(cut -d' ' -f1 <<< ${line})
    local name=$(cut -d' ' -f2 <<< ${line})
    kubectl delete ${type} ${name} -n istio-system --ignore-not-found=true
  done <${CONFIG_DIR}/delete
}

function label_namespaces(){
  echo "--> Add 'istio-injection' label to namespaces"
  while read line; do
    local name
    name=$(cut -d' ' -f1 <<< "${line}")
    local switch
    switch=$(cut -d' ' -f2 <<< "${line}")
    set +e
    kubectl label namespace "${name}" "istio-injection=${switch}" --overwrite
    set -e
  done <"${CONFIG_DIR}"/injection-in-namespaces
}

function configure_sidecar_injector() {
  echo "--> Configure sidecar injector"
  local configmap=$(kubectl -n istio-system get configmap istio-sidecar-injector -o jsonpath='{.data.config}')
  local policyDisabled=$(grep "policy: disabled" <<< "$configmap")
  if [[ -n ${policyDisabled} ]]; then
    # Force automatic injecting
    configmap=$(sed 's/policy: disabled/policy: enabled/' <<< "$configmap")
  fi

  configmap=$(sed 's/\[\[ .ProxyConfig.GetTracing.GetZipkin.GetAddress \]\]/zipkin.kyma-system:9411/g' <<< "$configmap")

  # Set limits for sidecar. Our namespaces have resource quota set thus every container needs to have limits defined.
  # In case there is no limits section add one at the beginning of container definition. It serves as default.
  CONTAINERS="istio-init istio-proxy"
  for CONTAINER in $CONTAINERS; do
    INSERTED=$(sed -n "/- name: ${CONTAINER}/,/image:/p" <<< "$configmap" | wc -l)
    if [[ "$INSERTED" -gt 2 ]]; then
      echo "Patch already applied for ${CONTAINER}"
    else
      configmap=$(sed "s|  - name: ${CONTAINER}|  - name: ${CONTAINER}\n    resources: { limits: { memory: 128Mi, cpu: 100m }, requests: { memory: 128Mi, cpu: 10m } }|" <<< "$configmap")
    fi
  done

  # Escape new lines and double quotes for kubectl
  configmap=$(sed -e ':a' -e 'N' -e '$!ba' -e 's/\n/\\n/g' <<< "$configmap")
  configmap=$(sed 's/"/\\"/g' <<< "$configmap")

  set +e
  local out
  out=$(kubectl patch -n istio-system configmap istio-sidecar-injector --type merge -p '{"data": {"config":"'"$configmap"'"}}')
  local result=$?
  set -e
  echo "$out"
  if [[ ${result} -ne 0 ]] && [[ ! "$out" = *"not patched"* ]]; then
    exit ${result}
  fi
}

function restart_sidecar_injector() {
  INJECTOR_POD_NAME=$(kubectl get pods -n istio-system -l istio=sidecar-injector -o=name)
  kubectl delete "${INJECTOR_POD_NAME}"
}

function check_requirements() {
  while read crd; do
    echo "Require CRD ${crd}"
    kubectl get customresourcedefinitions "${crd}"
    if [[ $? -ne 0 ]]; then
        echo "Cannot find required CRD ${crd}"
    fi
  done <${CONFIG_DIR}/required-crds
}

function check_ingress_ports() {
  while true; do
    local pod=$(kubectl get pod -l app=istio-ingressgateway -n istio-system | grep "istio-ingressgateway" | awk '{print $1}')
    local state=$(kubectl get pod ${pod} -n istio-system -o jsonpath="{.status.phase}")
    if [[ "$state" == "Running" ]]; then
      break
    else
      sleep 10s
    fi
  done
  local out=$(kubectl exec -t ${pod} -n istio-system -- netstat -lptnu)
  if echo "${out}" | grep "443" ; then
    echo "OPEN"
  else
    echo "CLOSED"
  fi
}

function restart_ingress_pod() {
  echo "---> Checking istio-ingressgateway ports"
  while true; do
    status=$(check_ingress_ports)
    if [[ "$status" == "OPEN" ]]; then
      break
    else
      sleep 5s
    fi
  done
}

require_istio_system
require_istio_version
require_mtls_enabled
check_requirements
configure_policy_checks
configure_sidecar_injector
restart_sidecar_injector
run_all_patches
remove_not_used
label_namespaces
restart_ingress_pod
