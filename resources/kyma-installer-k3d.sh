#!/usr/bin/env bash

# Requirements:
# kubectl
# curl
# k3d
# openssl

# Install k3d:
# curl -s https://raw.githubusercontent.com/rancher/k3d/master/install.sh | bash

START_TIME=$SECONDS
# Delete the old cluster if it exists
k3d delete --keep-registry-volume -n kyma || true
# Create Kyma cluster
k3d create --publish 80:80 --publish 443:443 --enable-registry --registry-volume local_registry --registry-name registry.localhost --server-arg --no-deploy --server-arg traefik -n kyma -t 60

# Delete cluster with keep-registry-volume to cache docker images
# k3d delete --keep-registry-volume -n kyma

export KUBECONFIG="$(k3d get-kubeconfig -n='kyma')"
export DOMAIN=local.kyma.pro
export GARDENER=false
export REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /k3d-registry)
sed "s/REGISTRY_IP/$REGISTRY_IP/" kyma-yaml/coredns-patch.tpl >kyma-yaml/coredns-patch.yaml
KYMA_SCRIPTS_DIR="../installation/scripts"
KYMA_RESOURCES_DIR="../installation/resources"
KYMA_INSTALLER_IMAGE="${KYMA_INSTALLER_IMAGE:-eu.gcr.io/kyma-project/kyma-operator:98e02519}"
INSTALLER_YAML="${KYMA_RESOURCES_DIR}/installer.yaml"

# This file will be created by cert-manager (not needed anymore):
rm ./core/charts/gateway/templates/kyma-gateway-certs.yaml || true

# Patch coredns with local domains
kubectl -n kube-system patch cm coredns --patch "$(cat kyma-yaml/coredns-patch.yaml)"

# Create kyma-installer ns and label
kubectl create ns kyma-installer
kubectl label ns kyma-installer istio-injection=disabled --overwrite
kubectl label ns kyma-installer kyma-project.io/installation="" --overwrite


# Having dummy values as cert-manager
TLS_CERT=ZHVtbXkK
TLS_KEY=ZHVtbXkK

# Create overrides for net-global
kubectl create configmap net-global-overrides \
      --from-literal global.isLocalEnv="true" \
      --from-literal global.domainName="$DOMAIN" \
      --from-literal global.minikubeIP="127.0.0.1" \
      --from-literal global.ingress.domainName="$DOMAIN" \
      --from-literal global.ingress.tlsCrt=$TLS_CERT \
      --from-literal global.ingress.tlsKey=$TLS_KEY \
      --from-literal global.environment.gardener="$GARDENER" \
      -n kyma-installer
kubectl label cm net-global-overrides -n kyma-installer installer=overrides --overwrite
kubectl label cm net-global-overrides -n kyma-installer kyma-project.io/installation= --overwrite

# Create overrides for Ory
kubectl create configmap ory-overrides \
      --from-literal global.ory.hydra.persistence.enabled=false \
      --from-literal global.ory.hydra.persistence.postgresql.enabled=false \
      --from-literal hydra.hydra.autoMigrate=false \
      -n kyma-installer
kubectl label cm ory-overrides -n kyma-installer installer=overrides
kubectl label cm ory-overrides -n kyma-installer component=ory

# Create overrides for Serverless
kubectl create configmap serverless-overrides \
      --from-literal dockerRegistry.enableInternal=false \
      --from-literal dockerRegistry.serverAddress=registry.localhost:5000 \
      --from-literal dockerRegistry.registryAddress=registry.localhost:5000 \
      --from-literal global.ingress.domainName="$DOMAIN" \
      -n kyma-installer
kubectl label cm serverless-overrides -n kyma-installer installer=overrides
kubectl label cm serverless-overrides -n kyma-installer component=serverless

echo "Manual concatenating yamls"
# shellcheck disable=SC2002
cat "${INSTALLER_YAML}" \
  | sed -e 's;image: eu.gcr.io/kyma-project/.*/installer:.*$;'"image: ${KYMA_INSTALLER_IMAGE};" \
  | kubectl apply -f-

# Apply installation CR
cat << EOF | kubectl apply -f -
apiVersion: "installer.kyma-project.io/v1alpha1"
kind: Installation
metadata:
  name: kyma-installation
  namespace: default
  labels:
    action: install
    kyma-project.io/installation: ""
  finalizers:
    - finalizer.installer.kyma-project.io
spec:
  components:
    - name: "cluster-essentials"
      namespace: "kyma-system"
    - name: "testing"
      namespace: "kyma-system"
    - name: "cert-manager"
      namespace: "cert-manager"
    - name: "istio"
      namespace: "istio-system"
    - name: "ingress-dns-cert"
      namespace: "istio-system"
    - name: "istio-kyma-patch"
      namespace: "istio-system"
    - name: "dex"
      namespace: "kyma-system"
    - name: "ory"
      namespace: "kyma-system"
    - name: "api-gateway"
      namespace: "kyma-system"
    - name: "core"
      namespace: "kyma-system"
    - name: "console"
      namespace: "kyma-system"
    - name: "cluster-users"
      namespace: "kyma-system"
    - name: "apiserver-proxy"
      namespace: "kyma-system"
    - name: "serverless"
      namespace: "kyma-system"
    - name: "application-connector"
      namespace: "kyma-integration"
    - name: "rafter"
      namespace: "kyma-system"
    - name: "service-catalog"
      namespace: "kyma-system"
    - name: "service-catalog-addons"
      namespace: "kyma-system"
    - name: "knative-serving"
      namespace: "knative-serving"
    - name: "knative-eventing"
      namespace: "knative-eventing"
    - name: "nats-streaming"
      namespace: "natss"
    - name: "knative-provisioner-natss"
      namespace: "knative-eventing"
    - name: "event-sources"
      namespace: "kyma-system"
EOF

while true; do \
  result=$(kubectl -n default get installation/kyma-installation -o jsonpath="{'Status: '}{.status.state}{', description: '}{.status.description}"); echo "${result}"; \
  if [[ "${result}" == *"Kyma installed"* ]]; then
    break
  fi
  sleep 2;
done
# Compute time taken to install
ELAPSED_TIME=$(($SECONDS - $START_TIME))

# Download the certificate:
kubectl get secret kyma-gateway-certs -n istio-system -o jsonpath='{.data.tls\.crt}' | base64 --decode > kyma.crt

# Import the certificate:
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain kyma.crt

echo 'Kyma Console Url:'
echo `kubectl get virtualservice console-web -n kyma-system -o jsonpath='{ .spec.hosts[0] }'`
echo 'User admin@kyma.cx, password:'
echo `kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode`

echo "$(($ELAPSED_TIME / 60)) minutes and $(($ELAPSED_TIME % 60)) seconds elapsed."
