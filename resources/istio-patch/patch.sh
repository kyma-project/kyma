#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

function run_patch() {
  local type=$1
  local name=$2
  local result=$?
  local patch=$(cat $DIR/$name.$type.patch.json)
  if [ $result != 0 ]; then
    echo $patch
    return $result
  fi
  kubectl patch $type -n istio-system $name --patch "$patch" --type json
}

function run_all_patches() {
  for f in $(find $DIR -name '*\.patch\.json' | xargs basename); do
    local type=$(cut -d. -f2 <<< $f)
    local name=$(cut -d. -f1 <<< $f)
    run_patch $type $name
  done
}

function remove_not_used() {
  while read line; do
    local type=$(cut -d' ' -f1 <<< $line)
    local name=$(cut -d' ' -f2 <<< $line)
    kubectl delete $type $name -n istio-system
  done <$DIR/delete
}

function configure_sidcar_injector() {
  local configmap=$(kubectl -n istio-system get configmap istio-sidecar-injector -o jsonpath='{.data.config}')
  # Disable autoinjecting
  configmap=$(sed 's/policy: enabled/policy: disabled/' <<< "$configmap")
  
  # Set limits for sidecar. Our namespaces have resourcequota set thus every container needs to have limits defined. 
  # Add limits to already existing resources sections
  configmap=$(sed 's|    resources:|    resources:\'$'\n      limits: { memory: 50Mi }|' <<< "$configmap")
  # In case there is no limits section add one at the begining of container definition/ It serves as default. 
  configmap=$(sed 's|  - name: istio-\(.*\)|  - name: istio-\1\'$'\n    resources: { limits: { memory: 50Mi } }|' <<< "$configmap")

  # Escape new lines and double quotes for kubectl
  configmap=$(sed -e ':a' -e 'N' -e '$!ba' -e 's/\n/\\n/g' <<< "$configmap")
  configmap=$(sed 's/"/\\"/g' <<< "$configmap")

  kubectl patch -n istio-system configmap istio-sidecar-injector --type merge -p '{"data": {"config":"'"$configmap"'"}}'
}

function apply_all() {
  for f in $(find $DIR -name '*\.yaml' | xargs basename); do
    kubectl apply -f $DIR/$f
  done
}

function open_ingress_ports() {
  if [[ $IS_LOCAL_INSTALLATION == "true" ]]; then
    kubectl patch -n istio-system deployment istio-ingressgateway --type json -p '
      [
        {
          "op": "add",
          "path": "/spec/template/spec/containers/0/ports/0/hostPort",
          "value": 80
        },{
          "op": "add",
          "path": "/spec/template/spec/containers/0/ports/1/hostPort",
          "value": 443
        }
      ]
    '
  fi
}

function set_external_load_balancer() {
  if [[ -n $EXTERNAL_PUBLIC_IP ]]; then
    kubectl patch -n istio-system service istio-ingressgateway --type json -p '
      [
        {
          "op": "replace",
          "path": "/spec/loadBalancerIP",
          "value": "'$EXTERNAL_PUBLIC_IP'"
        }
      ]
    '
  fi
}

run_all_patches
remove_not_used
apply_all
configure_sidcar_injector
open_ingress_ports
set_external_load_balancer
