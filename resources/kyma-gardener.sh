# Requirements:
# kubectl configured with gardener cluster
# helm3 
# 


export DOMAIN="$(kubectl -n kube-system get configmap shoot-info -o jsonpath='{.data.domain}')"
export GARDENER=true
export OVERRIDES=global.isLocalEnv=false,global.ingress.domainName=$DOMAIN,global.environment.gardener=$GARDENER,global.domainName=$DOMAIN,global.tlsCrt=ZHVtbXkK
export ORY=global.ory.hydra.persistence.enabled=false,global.ory.hydra.persistence.postgresql.enabled=false,hydra.hydra.autoMigrate=false

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

if [[ -f kyma-yaml/kyma-gateway-certs.yaml ]]
then
    echo "Restore certificate from the backup (to avoid lets encrypt rate limit)"
    kubectl apply -n istio-system -f kyma-yaml/kyma-gateway-certs.yaml
fi
helm3 upgrade -i --cleanup-on-fail cluster-essentials cluster-essentials --set $OVERRIDES -n kyma-system 
helm3 upgrade -i --cleanup-on-fail testing testing --set $OVERRIDES -n kyma-system
kubectl apply -f kyma-yaml/cert-manager.yaml
#kubectl -n kube-system patch cm coredns --patch "$(cat kyma-yaml/coredns-patch.yaml)"
helm3 upgrade -i --cleanup-on-fail istio istio -n istio-system --set $OVERRIDES 

while [[ $(kubectl get pods -n istio-system -l istio=sidecar-injector -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "waiting for istio" && sleep 5; done

helm3 upgrade -i --cleanup-on-fail ingress-dns-cert ingress-dns-cert --set $OVERRIDES -n istio-system  
helm3 upgrade -i --cleanup-on-fail istio-kyma-patch istio-kyma-patch -n istio-system --set $OVERRIDES 
helm3 upgrade -i --cleanup-on-fail knative-serving-init knative-serving-init -n knative-serving --set $OVERRIDES
helm3 upgrade -i --cleanup-on-fail knative-serving knative-serving -n knative-serving --set $OVERRIDES 
helm3 upgrade -i --cleanup-on-fail knative-eventing knative-eventing -n knative-eventing --set $OVERRIDES 

helm3 upgrade -i --cleanup-on-fail dex dex --set $OVERRIDES -n kyma-system 
helm3 upgrade -i --cleanup-on-fail ory ory --set $OVERRIDES --set $ORY -n kyma-system 
helm3 upgrade -i --cleanup-on-fail api-gateway api-gateway --set $OVERRIDES -n kyma-system 
helm3 upgrade -i --cleanup-on-fail rafter rafter --set $OVERRIDES -n kyma-system 
helm3 upgrade -i --cleanup-on-fail service-catalog service-catalog --set $OVERRIDES -n kyma-system 
helm3 upgrade -i --cleanup-on-fail service-catalog-addons service-catalog-addons --set $OVERRIDES -n kyma-system 
# helm3 upgrade -i --cleanup-on-fail helm-broker helm-broker --set $OVERRIDES -n kyma-system 
helm3 upgrade -i --cleanup-on-fail nats-streaming nats-streaming --set $OVERRIDES -n natss 

helm3 upgrade -i --cleanup-on-fail core core --set $OVERRIDES -n kyma-system 
helm3 upgrade -i --cleanup-on-fail cluster-users cluster-users --set $OVERRIDES -n kyma-system 
helm3 upgrade -i --cleanup-on-fail apiserver-proxy apiserver-proxy --set $OVERRIDES -n kyma-system 
helm3 upgrade -i --cleanup-on-fail serverless serverless --set $OVERRIDES -n kyma-system 
helm3 upgrade -i --cleanup-on-fail knative-provisioner-natss knative-provisioner-natss --set $OVERRIDES -n knative-eventing 
helm3 upgrade -i --cleanup-on-fail event-sources event-sources --set $OVERRIDES -n kyma-system 
helm3 upgrade -i --cleanup-on-fail application-connector application-connector --set $OVERRIDES -n kyma-integration 

# Create installer deployment scaled to 0 to get console running:
kubectl apply -f kyma-yaml/installer-local.yaml
if [[ ! -f kyma-yaml/kyma-gateway-certs.yaml ]]
then
    echo "Backup certificate from the cluster (lets encrypt rate limit)"
    kubectl get secret -n istio-system kyma-gateway-certs -oyaml > kyma-yaml/kyma-gateway-certs.yaml
fi

echo 'Kyma Console Url:'
echo `kubectl get virtualservice core-console -n kyma-system -o jsonpath='{ .spec.hosts[0] }'`
echo 'User admin@kyma.cx, password:'
echo `kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode`
