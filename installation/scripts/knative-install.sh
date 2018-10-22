#!/usr/bin/env bash
set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KN_BUILD_RELEASE=https://raw.githubusercontent.com/knative/serving/master/third_party/config/build/release.yaml
KN_SERVING_RELEASE=https://storage.googleapis.com/knative-releases/serving/latest/release-lite.yaml

source $CURRENT_DIR/utils.sh

#curl -L https://raw.githubusercontent.com/knative/serving/master/third_party/istio-1.0.2/istio.yaml \
#  | sed 's/LoadBalancer/NodePort/' \
#  | kubectl apply --filename -
#
#
#kubectl label namespace default istio-injection=enabled
#
#echo "Is istio ready?"
#
#READ YES
#
#curl -L https://github.com/knative/serving/releases/download/v0.1.1/release-lite.yaml \
#  | sed 's/LoadBalancer/NodePort/' \
#  | kubectl apply --filename -

#kubectl apply -f https://raw.githubusercontent.com/knative/serving/3022b58ef16afab56e9754243f70e90fc336b1f1/third_party/istio-1.0.2/istio-crds.yaml
#while [ $(kubectl get crd gateways.networking.istio.io -o jsonpath='{.status.conditions[?(@.type=="Established")].status}') != 'True' ]; do
#  echo "Waiting on Istio CRDs"; sleep 1
#done
#kubectl apply -f https://raw.githubusercontent.com/knative/serving/3022b58ef16afab56e9754243f70e90fc336b1f1/third_party/istio-1.0.2/istio.yaml

#kubectl apply -f https://raw.githubusercontent.com/knative/serving/master/third_party/config/build/release.yaml



TMPDIR=`mktemp -d "${CURRENT_DIR}/../../temp-XXXXXXXXXX"`
pushd $TMPDIR

######### istio ########

log "Download Istio" yellow
curl -L -J -O https://github.com/istio/istio/releases/download/1.0.2/istio-1.0.2-osx.tar.gz
tar xzf istio-1.0.2-osx.tar.gz
cd istio-1.0.2/install/kubernetes/helm/
log "Installing Istio CRDs" yellow
kubectl create namespace istio-system
kubectl apply -f istio/templates/crds.yaml
while [ $(kubectl get crd gateways.networking.istio.io -o jsonpath='{.status.conditions[?(@.type=="Established")].status}') != 'True' ]; do
  log "Waiting on Istio CRDs" yellow; sleep 1
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

log "Waiting for Istio to run" yellow
${CURRENT_DIR}/is-ready.sh istio-system istio galley
${CURRENT_DIR}/is-ready.sh istio-system istio pilot
${CURRENT_DIR}/is-ready.sh istio-system istio mixer
${CURRENT_DIR}/is-ready.sh istio-system istio citadel
${CURRENT_DIR}/is-ready.sh istio-system app prometheus
log "Istio installed successfully" green

######### knative build ########
log "Installing Knative Build" yellow
kubectl apply -f $KN_BUILD_RELEASE

log "Verifying Knative Build installation" yellow
sleep 2
until kubectl get -f $KN_BUILD_RELEASE > /dev/null 2>&1
do
    log "Knative Build CRDs not yet synced, re-applying." yellow
    kubectl apply -f $KN_BUILD_RELEASE
    sleep 2
done

${CURRENT_DIR}/is-ready.sh knative-build app build-controller
${CURRENT_DIR}/is-ready.sh knative-build app build-webhook

log "Knative Build installation verified" green

######### knative serving ########

log "Installing Knative Serving" yellow

curl -L $KN_SERVING_RELEASE \
| sed 's/LoadBalancer/NodePort/' \
| tee knative-serving.yaml \
| kubectl apply -f -


log "Verifying Knative Serving installation" yellow
sleep 2
until kubectl get -f $KN_SERVING_RELEASE > /dev/null 2>&1
do
    log "Knative Serving CRDs not yet synced, re-applying." yellow
    kubectl apply -f knative-serving.yaml
    sleep 2
done

${CURRENT_DIR}/is-ready.sh knative-serving app activator
${CURRENT_DIR}/is-ready.sh knative-serving app autoscaler
${CURRENT_DIR}/is-ready.sh knative-serving app controller
${CURRENT_DIR}/is-ready.sh knative-serving app webhook
${CURRENT_DIR}/is-ready.sh knative-monitoring app grafana
${CURRENT_DIR}/is-ready.sh knative-monitoring app kube-state-metrics
${CURRENT_DIR}/is-ready.sh knative-monitoring app node-exporter
${CURRENT_DIR}/is-ready.sh knative-monitoring app prometheus

log "Knative Serving installation verified" green

popd
rm -rf $TMPDIR