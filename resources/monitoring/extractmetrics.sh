#!/usr/bin/env bash

set -o errexit                                      # stop script when an error occurs
set -o pipefail                                     # exit status of the last command that threw a non-zero exit code is returned
set -o nounset                                      # exit when your script tries to use undeclared variables.
[[ "${DEBUG:-false}" == 'true' ]] && set -o xtrace  # prints every expression before executing it

declare _allmetrics
declare NAME
declare URL
declare RULES_PATH

while [[ $# -gt 0 ]]
do
    key="$1"
    shift
    case ${key} in
        --name|-n)
            NAME="$1"
            shift
            ;;
        --url|-u)
            URL="$1"
            shift
            ;;
        --rules-path|-r)
            RULES_PATH="$1"
            shift
            ;;
        -*)
            echo "Unknown flag ${key}"
            exit 1
            ;;
    esac
done

#### DO NOT EDIT BELOW ####
readonly __rootpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly __rulespath="${RULES_PATH}"

readonly _prometheus_pod_spec=$(kubectl get pod -l app=prometheus --all-namespaces -o json | jq -r '.items[0] | .metadata.namespace + "|" + .spec.serviceAccountName' | tail -n1)
readonly _pod_ns=$(echo "${_prometheus_pod_spec}" | cut -d"|" -f1)
readonly _pod_sa=$(echo "${_prometheus_pod_spec}" | cut -d"|" -f2)
readonly _pod_name="metrics-fetcher"

function cleanup {
  printf "\nCleaning up temporary pod."
  kubectl delete pod --namespace "${_pod_ns}" "${_pod_name}" >/dev/null 2>&1
  while kubectl get pod --namespace "${_pod_ns}" "${_pod_name}" >/dev/null 2>&1; do
    printf "."
  done
  printf "[Done]\n"
}
trap cleanup EXIT

function init() {
  mkdir -p "${__rulespath}"

  if [[ $(find -L "${__rulespath}" -type f | wc -l) -eq 0 ]]; then
    printf "ERROR: rules folder \"%s\" is empty.\n" "${__rulespath}" 
    printf "Here are some examples you can \"git clone\" or copy the files you want into %s :\n" "${__rulespath}"
    printf "\t- https://github.com/istio/installer/tree/master/istio-telemetry/grafana/dashboards\n"
    printf "\t- https://github.com/helm/charts/tree/master/stable/prometheus-operator/templates/grafana/dashboards\n"
    printf "\t- https://github.com/helm/charts/tree/master/stable/prometheus-operator/templates/prometheus/rules\n"
    printf "\t- https://github.com/kubernetes-monitoring/kubernetes-mixin\n"
    exit 1
  fi

  if ! kubectl get pod --namespace "${_pod_ns}" "${_pod_name}" >/dev/null 2>&1; then
    printf "Creating temporary pod"
    kubectl run "${_pod_name}" \
      --namespace "${_pod_ns}" \
      --serviceaccount="${_pod_sa}" \
      --generator=run-pod/v1 \
      --restart=Never \
      --overrides='{"apiVersion":"v1","metadata":{"annotations":{"sidecar.istio.io/inject":"false"}}}' \
      --image=tutum/curl:latest -- sleep 1d > /dev/null 2>&1

    while ! kubectl get pod --namespace "${_pod_ns}" "${_pod_name}" -o json | jq -er 'select(.status.phase=="Running")' >/dev/null 2>&1; do
      printf "."
    done
    printf "[Done]\n\n"
  fi
}

function gettype() {
  local mname="${1}"
  local mtype=""
  mtype=$(echo "${_allmetrics}" | awk '/#\s*TYPE\s*'"${mname}"'\s/{print $0}')
  if [[ "${mtype}" != "" ]]; then
    echo "${mtype}" | awk '{print $4}'
  fi
}

function getdesc() {
  local mname="${1}"
  local mdesc=""
  echo ${mname}
  mdesc=$(echo "${_allmetrics}" | awk '/#\s*HELP\s*'"${mname}"'\s/{print $0}')
  if [[ "${mdesc}" != "" ]]; then
    echo "${mdesc}" | cut -d" " -f4-
  fi
}

function write-metrics() {
  local name="${1}"
  local url="${2}"

  printf "Fetching metrics for %s\n" "${name}"

  # shellcheck disable=SC2016
  _allmetrics=$(kubectl exec --namespace "${_pod_ns}" -it "${_pod_name}" -- \
    sh -c 'TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token); curl --insecure --silent --header "Authorization: Bearer ${TOKEN}" '"${url}" | \
    tr -dc '[:print:]\n')

  # extract all uniq metrics names and return only the one used in rules and dashboards
  _umetrics=$(echo "${_allmetrics}" | sed -E -e '/^#.*$/d' -e 's/\{.*$//' -e 's/^([^ ]*).*$/\1/' | sort -u )
  _sortedmetrics=$(grep -hRio ${__rulespath} -f <(echo "${_umetrics}") | sort -u)
  # generate the metric relabeling config regex
  _relabel_regex=$(echo "${_sortedmetrics}" | tr '\n' '|' | sed 's/|$//')

  # generate readme

  cat << EOL

### metric relabeling config for ${name} ###
metricRelabeligs:
- sourceLabels: [ __name__ ]
  regex: ^(${_relabel_regex})$
  action: keep

EOL
}

init
write-metrics "${NAME}" "${URL}"

