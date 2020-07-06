# Requirements:
# kubectl 
# helm3 
# k3d


# Install k3d:
# curl -s https://raw.githubusercontent.com/rancher/k3d/master/install.sh | bash

# Create Kyma cluster
k3d create --publish 80:80 --publish 443:443 --enable-registry --registry-volume local_registry --registry-name registry.localhost --server-arg --no-deploy --server-arg traefik -n kyma -t 60 

# Delete cluster with keep-registry-volume to cache docker images
# k3d delete --keep-registry-volume -n kyma

export KUBECONFIG="$(k3d get-kubeconfig -n='kyma')"
export DOMAIN=local.kyma.pro
export GARDENER=false
export OVERRIDES=global.isLocalEnv=true,global.ingress.domainName=$DOMAIN,global.environment.gardener=$GARDENER,global.domainName=$DOMAIN,global.minikubeIP=127.0.0.1,global.tlsCrt=ZHVtbXkK
export ORY=global.ory.hydra.persistence.enabled=false,global.ory.hydra.persistence.postgresql.enabled=false,hydra.hydra.autoMigrate=false
export LOCALREGISTRY="dockerRegistry.enableInternal=false,dockerRegistry.serverAddress=registry.localhost:5000,dockerRegistry.registryAddress=registry.localhost:5000,global.ingress.domainName=$DOMAIN"
export REGISTRY_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' /k3d-registry)
sed "s/REGISTRY_IP/$REGISTRY_IP/" kyma-yaml/coredns-patch.tpl >kyma-yaml/coredns-patch.yaml

# This file will be created by cert-manager (not needed anymore):
rm ./core/charts/gateway/templates/kyma-gateway-certs.yaml

# Create namespaces
kubectl create ns kyma-system
kubectl create ns istio-system
kubectl create ns kyma-integration
kubectl create ns cert-manager
kubectl create ns knative-serving
kubectl create ns knative-eventing
kubectl create ns natss

kubectl label ns istio-system istio-injection=disabled --overwrite
kubectl label ns cert-manager istio-injection=disabled --overwrite

helm3 upgrade -i cluster-essentials cluster-essentials --set $OVERRIDES -n kyma-system 
helm3 upgrade -i testing testing --set $OVERRIDES -n kyma-system
kubectl apply -f kyma-yaml/cert-manager.yaml
kubectl -n kube-system patch cm coredns --patch "$(cat kyma-yaml/coredns-patch.yaml)"
helm3 upgrade -i istio istio -n istio-system --set $OVERRIDES 

while [[ $(kubectl get pods -n istio-system -l istio=sidecar-injector -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "waiting for istio" && sleep 5; done

helm3 upgrade -i ingress-dns-cert ingress-dns-cert --set $OVERRIDES -n istio-system  
helm3 upgrade -i istio-kyma-patch istio-kyma-patch -n istio-system --set $OVERRIDES 
# helm3 upgrade -i knative-serving knative-serving -n knative-serving --set $OVERRIDES 
# helm3 upgrade -i knative-eventing knative-eventing -n knative-eventing --set $OVERRIDES 

helm3 upgrade -i dex dex --set $OVERRIDES -n kyma-system 
helm3 upgrade -i ory ory --set $OVERRIDES --set $ORY -n kyma-system 
helm3 upgrade -i api-gateway api-gateway --set $OVERRIDES -n kyma-system 

helm3 upgrade -i rafter rafter --set $OVERRIDES -n kyma-system 
helm3 upgrade -i service-catalog service-catalog --set $OVERRIDES -n kyma-system 
helm3 upgrade -i service-catalog-addons service-catalog-addons --set $OVERRIDES -n kyma-system 
helm3 upgrade -i helm-broker helm-broker --set $OVERRIDES -n kyma-system 

helm3 upgrade -i core core --set $OVERRIDES -n kyma-system 
helm3 upgrade -i console console --set $OVERRIDES -n kyma-system 
helm3 upgrade -i cluster-users cluster-users --set $OVERRIDES -n kyma-system 
helm3 upgrade -i apiserver-proxy apiserver-proxy --set $OVERRIDES -n kyma-system 
helm3 upgrade -i serverless serverless --set $LOCALREGISTRY -n kyma-system 

helm3 upgrade -i application-connector application-connector --set $OVERRIDES -n kyma-integration 

# Install knative-eventing and knative-serving
helm3 upgrade -i knative-serving knative-serving --set $OVERRIDES -n knative-serving 
helm3 upgrade -i knative-eventing knative-eventing -n knative-eventing
helm3 upgrade -i knative-provisioner-natss knative-provisioner-natss -n knative-eventing
helm3 upgrade -i nats-streaming nats-streaming -n natss
helm3 upgrade -i event-sources event-sources -n kyma-system


# Create installer deployment scaled to 0 to get console running:
kubectl apply -f kyma-yaml/installer-local.yaml

# Download the certificate: 
kubectl get secret kyma-gateway-certs -n istio-system -o jsonpath='{.data.tls\.crt}' | base64 --decode > kyma.crt
# Import the certificate: 
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain kyma.crt

echo 'Kyma Console Url:'
echo `kubectl get virtualservice console-web -n kyma-system -o jsonpath='{ .spec.hosts[0] }'`
echo 'User admin@kyma.cx, password:'
echo `kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode`
