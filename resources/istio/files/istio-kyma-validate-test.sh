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

if [[ -z ${REQUIRED_ISTIO_VERSION} ]]; then
    log "Please set REQUIRED_ISTIO_VERSION variable!" red
    exit 1
fi

function require_istio_version() {
    local version
    version=$(kubectl -n istio-system get deployment istiod -o jsonpath='{.spec.template.spec.containers[0].image}' | awk -F: '{print $2}')
    if [[ "$version" != "${REQUIRED_ISTIO_VERSION}" ]]; then
        log "Istio must be in version: $REQUIRED_ISTIO_VERSION!" red
        exit 1
    fi
}

function require_istio_system() {
    kubectl get namespace istio-system >/dev/null
}

function check_mtls_enabled() {
  log "--> Check global mTLS"
  local mTLS=$(kubectl get PeerAuthentication -n istio-system default -o jsonpath='{.spec.mtls.mode}')
  local status=$?
  if [[ $status != 0 ]]; then
    log "----> PeerAuthentication istio-system/default not found!" red
    exit 1
  fi
  if [[ "${mTLS}" != "STRICT" ]]; then
    log "----> mTLS must be \"STRICT\"" red
    exit 1
  fi
  log "----> mTLS is enabled" green
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

function require_ingressgateway_hpa() {
    echo "--> Checking istio-ingresgateway HPA"
    local targetMemHPA=$(kubectl -n istio-system get istiooperators.install.istio.io installed-state -o jsonpath="{.spec.components.ingressGateways[0].k8s.hpaSpec.metrics[?(@.resource.name=='memory')].resource.targetAverageUtilization}")
    if [[ ${targetMemHPA} != "80" ]]; then
       echo "   Memory based HPA needs to be set and targetAverageUtilization is 80" red
       exit 1
    fi
    echo " Memory based HPA is has targetAverageUtilization set to 80" green
}

require_istio_system
require_istio_version
require_ingressgateway_hpa
check_mtls_enabled
check_sidecar_injector
log "Istio is configured to run Kyma!" green
