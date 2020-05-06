helm3 delete cluster-essentials  -n kyma-system 
helm3 delete testing  -n kyma-system 
kubectl delete -f kyma-yaml/cert-manager.yaml
helm3 delete istio-kyma-patch  -n istio-system 
helm3 delete istio  -n istio-system 

helm3 delete ingress-dns-cert  -n istio-system  
helm3 delete knative-serving-init -n knative-serving 
helm3 delete knative-serving -n knative-serving 
helm3 delete knative-eventing -n knative-eventing

helm3 delete dex  -n kyma-system 
helm3 delete ory -n kyma-system 
helm3 delete api-gateway -n kyma-system 
helm3 delete rafter rafter -n kyma-system 
helm3 delete service-catalog -n kyma-system 
helm3 delete service-catalog-addons -n kyma-system 
# helm3 delete helm-broker helm-broker -n kyma-system 
helm3 delete nats-streaming -n natss 

helm3 delete core -n kyma-system 
helm3 delete cluster-users -n kyma-system 
helm3 delete apiserver-proxy -n kyma-system 
helm3 delete serverless -n kyma-system 
helm3 delete knative-provisioner-natss -n knative-eventing 
helm3 delete event-sources -n kyma-system 
helm3 delete application-connector -n kyma-integration 

# Create installer deployment scaled to 0 to get console running:
kubectl delete -f kyma-yaml/installer-local.yaml

kubectl delete ns kyma-system
kubectl delete ns kyma-integration
kubectl delete ns knative-serving
kubectl delete ns knative-eventing
kubectl delete ns natss
kubectl delete ns istio-system
