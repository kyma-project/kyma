# Requirements:
# kubectl
# helm3
# k3d


# Install k3d:
# curl -s https://raw.githubusercontent.com/rancher/k3d/master/install.sh | bash

START_TIME=$SECONDS
function print_installed_time_for_component() {
  local start_time=$1
  local stop_time=$2
  local component=$3
  local ELAPSED_TIME=$(($stop_time - $start_time))
  echo "${component} is installed in $(($ELAPSED_TIME / 60)) minutes and $(($ELAPSED_TIME % 60)) seconds."
}
# Create Kyma cluster
k3d create --publish 80:80 --publish 443:443 --enable-registry --registry-volume local_registry --registry-name registry.localhost --server-arg --no-deploy --server-arg traefik -n kyma -t 60

# Delete cluster with keep-registry-volume to cache docker images
# k3d delete --keep-registry-volume -n kyma

export KUBECONFIG="$(k3d get-kubeconfig -n='kyma')"
export DOMAIN=local.kyma.pro
export GARDENER=false
export OVERRIDES=global.isLocalEnv=true,global.ingress.domainName=$DOMAIN,global.environment.gardener=$GARDENER,global.domainName=$DOMAIN,global.minikubeIP=127.0.0.1,global.tlsCrt=ZHVtbXkK
export ORY=global.ory.hydra.persistence.enabled=false,global.ory.hydra.persistence.postgresql.enabled=false,hydra.hydra.autoMigrate=false
export LOCALREGISTRY="docker-registry.enabled=false,containers.manager.envs.functionDockerAddress.value=registry.localhost:5000,containers.manager.envs.functionDockerExternalAddress.value=registry.localhost:5000,global.ingress.domainName=$DOMAIN"
export REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /k3d-registry)
sed "s/REGISTRY_IP/$REGISTRY_IP/" kyma-yaml/coredns-patch.tpl >kyma-yaml/coredns-patch.yaml


# This file will be created by cert-manager (not needed anymore):
rm ./core/charts/gateway/templates/kyma-gateway-certs.yaml || true

# Create namespaces
kubectl create ns kyma-system
kubectl create ns istio-system
kubectl create ns kyma-integration
kubectl create ns cert-manager
kubectl create ns knative-serving
kubectl create ns knative-eventing
kubectl create ns natss

# Disable istio injection for some ns
kubectl label ns istio-system istio-injection=disabled --overwrite
kubectl label ns cert-manager istio-injection=disabled --overwrite

start_time=$SECONDS
helm upgrade -i cluster-essentials cluster-essentials --set $OVERRIDES -n kyma-system
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "cluster-essentials"

start_time=$SECONDS
helm upgrade -i testing testing --set $OVERRIDES -n kyma-system
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "testing"

kubectl apply -f kyma-yaml/cert-manager.yaml
kubectl -n kube-system patch cm coredns --patch "$(cat kyma-yaml/coredns-patch.yaml)"

start_time=$SECONDS
helm upgrade -i istio istio -n istio-system --set $OVERRIDES
while [[ $(kubectl get pods -n istio-system -l istio=sidecar-injector -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "waiting for istio" && sleep 5; done
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "istio"

start_time=$SECONDS
helm upgrade -i ingress-dns-cert ingress-dns-cert --set $OVERRIDES -n istio-system
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "ingress-dns-cert"

start_time=$SECONDS
helm upgrade -i istio-kyma-patch istio-kyma-patch -n istio-system --set $OVERRIDES
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "istio-kyma-patch"

start_time=$SECONDS
helm upgrade -i dex dex --set $OVERRIDES -n kyma-system
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "dex"

start_time=$SECONDS
helm upgrade -i ory ory --set $OVERRIDES --set $ORY -n kyma-system
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "ory"

start_time=$SECONDS
helm upgrade -i api-gateway api-gateway --set $OVERRIDES -n kyma-system
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "api-gateway"

start_time=$SECONDS
helm upgrade -i core core --set $OVERRIDES -n kyma-system
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "core"

start_time=$SECONDS
helm upgrade -i console console --set $OVERRIDES -n kyma-system
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "console"

start_time=$SECONDS
helm upgrade -i cluster-users cluster-users --set $OVERRIDES -n kyma-system
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "cluster-users"

start_time=$SECONDS
helm upgrade -i apiserver-proxy apiserver-proxy --set $OVERRIDES -n kyma-system
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "apiserver-proxy"

start_time=$SECONDS
helm upgrade -i serverless serverless --set $LOCALREGISTRY -n kyma-system
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "serverless"

start_time=$SECONDS
helm upgrade -i application-connector application-connector -n kyma-integration --set $OVERRIDES
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "application-connector"

start_time=$SECONDS
helm upgrade -i rafter rafter -n kyma-system --set $OVERRIDES
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "rafter"

start_time=$SECONDS
helm upgrade -i service-catalog service-catalog -n kyma-system --set $OVERRIDES
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "service-catalog"

start_time=$SECONDS
helm upgrade -i service-catalog-addons service-catalog-addons -n kyma-system --set $OVERRIDES
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "service-catalog-addons service-catalog-addons"


# Install knative-eventing and knative-serving
start_time=$SECONDS
helm upgrade -i knative-serving knative-serving -n knative-serving --set $OVERRIDES
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "knative-serving"

start_time=$SECONDS
helm upgrade -i knative-eventing knative-eventing -n knative-eventing
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "knative-eventing"

start_time=$SECONDS
helm upgrade -i knative-provisioner-natss knative-provisioner-natss -n knative-eventing
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "knative-provisioner-natss"

start_time=$SECONDS
helm upgrade -i nats-streaming nats-streaming -n natss
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "nats-streaming"

start_time=$SECONDS
helm upgrade -i event-sources event-sources -n kyma-system
stop_time=$SECONDS
print_installed_time_for_component "${start_time}" "${stop_time}" "event-sources"


# Install Application connector

# Create installer deployment scaled to 0 to get console running:
kubectl apply -f kyma-yaml/installer-local.yaml

# Compute time taken to install
ELAPSED_TIME=$(($SECONDS - $START_TIME))

# Download the certificate:
kubectl get secret kyma-gateway-certs -n istio-system -o jsonpath='{.data.tls\.crt}' | base64 --decode > kyma.crt

# Patching istio pods to consume less resources
kubectl patch deploy -n istio-system $(kubectl get deploy -n istio-system -lapp=istio-mixer -ojsonpath='{.items[*].metadata.name}') --patch="$(cat istio-resources-patch.yaml)"

# Import the certificate:
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain kyma.crt

echo 'Kyma Console Url:'
echo `kubectl get virtualservice console-web -n kyma-system -o jsonpath='{ .spec.hosts[0] }'`
echo 'User admin@kyma.cx, password:'
echo `kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode`

echo "$(($ELAPSED_TIME / 60)) minutes and $(($ELAPSED_TIME % 60)) seconds elapsed."
