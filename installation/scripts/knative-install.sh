#!/usr/bin/env bash
set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KN_SERVING_URL=https://storage.googleapis.com/knative-releases/serving/latest/release-lite.yaml
KN_EVENTING_URL=https://storage.googleapis.com/knative-releases/eventing/latest/release.yaml

source $CURRENT_DIR/utils.sh


POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --serving-url)
            KN_SERVING_URL=$2
            shift
            shift
            ;;
        --eventing-url)
            KN_EVENTING_URL=$2
            shift
            shift
            ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters


TMPDIR=`mktemp -d "${CURRENT_DIR}/../../temp-XXXXXXXXXX"`
pushd $TMPDIR

######### istio ########

log "Downloading Istio..." yellow
curl -L -J -O https://github.com/istio/istio/releases/download/1.0.2/istio-1.0.2-osx.tar.gz
tar xzf istio-1.0.2-osx.tar.gz
cd istio-1.0.2/install/kubernetes/helm/
log "Installing Istio CRDs..." yellow
kubectl create namespace istio-system
kubectl apply -f istio/templates/crds.yaml
while [ $(kubectl get crd gateways.networking.istio.io -o jsonpath='{.status.conditions[?(@.type=="Established")].status}') != 'True' ]; do
  log "Waiting on Istio CRDs..." yellow; sleep 1
done

log "Installing Istio..."
helm template -n istio --namespace istio-system \
 --set security.enabled=true \
 --set global.proxy.includeIPRanges="10.0.0.1/8" \
 --set gateways.istio-ingressgateway.service.externalPublicIp="" \
 --set gateways.istio-ingressgateway.type="NodePort" \
 --set pilot.resources.limits.memory=1024Mi \
 --set pilot.resources.limits.cpu=100m \
 --set pilot.resources.requests.memory=256Mi \
 --set pilot.resources.requests.cpu=100m \
 --set mixer.resources.limits.memory=256Mi \
 --set mixer.resources.requests.memory=128Mi \
  istio | kubectl apply -f -

kubectl label namespace default istio-injection=enabled

log "Waiting for Istio to run..." yellow
${CURRENT_DIR}/is-ready.sh istio-system istio galley
${CURRENT_DIR}/is-ready.sh istio-system istio pilot
${CURRENT_DIR}/is-ready.sh istio-system istio mixer
${CURRENT_DIR}/is-ready.sh istio-system istio citadel
${CURRENT_DIR}/is-ready.sh istio-system app prometheus
log "Istio is installed successfully." green

######### Knative build & serving ########

log "Installing Knative build and serving ..." yellow

curl -L $KN_SERVING_URL \
| sed 's/LoadBalancer/NodePort/' \
| tee knative-serving.yaml \
| kubectl apply -f -


log "Verifying Knative build and serving installation..." yellow
sleep 2
until kubectl get -f $KN_SERVING_URL > /dev/null 2>&1
do
    log "Knative CRDs not yet synced, re-applying..." yellow
    kubectl apply -f knative-serving.yaml
    sleep 2
done

${CURRENT_DIR}/is-ready.sh knative-build app build-controller
${CURRENT_DIR}/is-ready.sh knative-build app build-webhook
${CURRENT_DIR}/is-ready.sh knative-serving app activator
${CURRENT_DIR}/is-ready.sh knative-serving app autoscaler
${CURRENT_DIR}/is-ready.sh knative-serving app controller
${CURRENT_DIR}/is-ready.sh knative-serving app webhook
${CURRENT_DIR}/is-ready.sh knative-monitoring app grafana
${CURRENT_DIR}/is-ready.sh knative-monitoring app kube-state-metrics
${CURRENT_DIR}/is-ready.sh knative-monitoring app node-exporter
${CURRENT_DIR}/is-ready.sh knative-monitoring app prometheus

log "Knative build and serving installation verified" green

log "Installing Knative eventing..." yellow

kubectl apply -f $KN_EVENTING_URL

log "Verifying Knative eventing installation..." yellow

${CURRENT_DIR}/is-ready.sh knative-eventing app eventing-controller
${CURRENT_DIR}/is-ready.sh knative-eventing app webhook

log "Knative eventing installation verified" green


popd
rm -rf $TMPDIR