#!/usr/bin/env bash

function log() {
    local exp=$1;
    local color=$2;
    local style=$3;
    local NC='\033[0m'
    if ! [[ ${color} =~ '^[0-9]$' ]] ; then
       case $(echo ${color} | tr '[:upper:]' '[:lower:]') in
        black) color='\e[30m' ;;
        red) color='\e[31m' ;;
        green) color='\e[32m' ;;
        yellow) color='\e[33m' ;;
        blue) color='\e[34m' ;;
        magenta) color='\e[35m' ;;
        cyan) color='\e[36m' ;;
        white) color='\e[37m' ;;
        nc|*) color=${NC} ;; # no color or invalid color
       esac
    fi
    if ! [[ ${style} =~ '^[0-9]$' ]] ; then
        case $(echo ${style} | tr '[:upper:]' '[:lower:]') in
        bold) style='\e[1m' ;;
        underline) style='\e[4m' ;;
        inverted) style='\e[7m' ;;
        *) style="" ;; # no style or invalid style
       esac
    fi
    printf "${color}${style}${exp}${NC}\n"
}

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

if [[ -z ${REQUIRED_ISTIO_VERSION} ]]; then
    log "Please set REQUIRED_ISTIO_VERSION variable!" red
    exit 1
fi

if [[ -z ${CONFIG_DIR} ]]; then
    CONFIG_DIR=${DIR}
fi

function require_istio_version() {
    local version
    version=$(c)
    if [[ "$version" != ${REQUIRED_ISTIO_VERSION} ]]; then
        log "Istio must be in version: $REQUIRED_ISTIO_VERSION!" red
        exit 1
    fi
}

function require_istio_system() {
    kubectl get namespace istio-system >/dev/null
}

function check_mtls_enabled() {
    # TODO: rethink how that should be done
    local mTLS=$(kubectl get meshpolicy default -o jsonpath='{.spec.peers[0].mtls.mode}')
    if [[ "${mTLS}" != "STRICT" ]] && [[ "${mTLS}" != "" ]]; then
        log "mTLS must be \"STRICT\"" red
        exit 1
    fi
}

function check_policy_checks(){
  log "--> Check policy checks"
  local istioConfigmap="$(kubectl -n istio-system get cm istio -o jsonpath='{@.data.mesh}')"
  local policyChecksDisabled=$(grep "disablePolicyChecks: true" <<< "$istioConfigmap")
  if [[ -n ${policyChecksDisabled} ]]; then
      log "    disablePolicyChecks must be FALSE" red
      exit 1
  fi
  log "    Policy checks are enabled" green
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

function check_sidecar_injector() {
  echo "--> Check sidecar injector"
  local configmap=$(kubectl -n istio-system get configmap istio-sidecar-injector -o jsonpath='{.data.config}')
  local policyDisabled=$(grep "policy: disabled" <<< "$configmap")
  if [[ -n ${policyDisabled} ]]; then
    # Force automatic injecting
    log "    Automatic injection policy must be ENABLED" red
    exit 1
  fi
  log "    Automatic injection policy is enabled" green
}

function restart_sidecar_injector() {
  INJECTOR_POD_NAME=$(kubectl get pods -n istio-system -l istio=sidecar-injector -o=name)
  kubectl delete "${INJECTOR_POD_NAME}"
}

function check_requirements() {
  while read crd; do
    log "Require CRD ${crd}"
    kubectl get customresourcedefinitions "${crd}"
    if [[ $? -ne 0 ]]; then
        log "Cannot find required CRD ${crd}" red
        exit 1
    fi
    log "CRD ${crd} present" green
  done <${CONFIG_DIR}/required-crds
}

require_istio_system
require_istio_version
check_mtls_enabled
check_requirements
check_policy_checks
check_sidecar_injector
# restart_sidecar_injector
# run_all_patches
# remove_not_used
# label_namespaces
log "Istio is configured to run Kyma!" green
