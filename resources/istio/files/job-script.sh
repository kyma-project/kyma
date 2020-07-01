#!/bin/bash -e
#printf "\n=== print IstioOperator configuration ==========================================\n"
#cat /etc/istio/istioOperator-draft.yaml
##printf "\n=== generate manifest from IstioOperator configuration =========================\n"
##echo 'istioctl manifest generate -f /etc/istio/istioOperator-draft.yaml --set "values.global.jwtPolicy=first-party-jwt" --set "values.mixer.telemetry.loadshedding.mode=disabled" --set "values.mixer.policy.autoscaleEnabled=false" --set "components.policy.k8s.resources.limits.cpu=500m" --set "components.policy.k8s.resources.limits.memory=2048Mi" --set "components.policy.k8s.resources.requests.cpu=300m" --set "components.policy.k8s.resources.requests.memory=512Mi" --set "values.mixer.telemetry.autoscaleEnabled=false" --set "components.telemetry.k8s.resources.limits.cpu=500m" --set "components.telemetry.k8s.resources.limits.memory=2048Mi" --set "components.telemetry.k8s.resources.requests.cpu=300m" --set "components.telemetry.k8s.resources.requests.memory=512Mi" --set "values.pilot.autoscaleEnabled=false" --set "components.pilot.k8s.resources.limits.cpu=500m" --set "components.pilot.k8s.resources.limits.memory=1024Mi" --set "components.pilot.k8s.resources.requests.cpu=250m" --set "components.pilot.k8s.resources.requests.memory=512Mi" > padu.yaml'
##istioctl manifest generate -f /etc/istio/istioOperator-draft.yaml       --set "values.global.jwtPolicy=first-party-jwt" --set "values.mixer.telemetry.loadshedding.mode=disabled" --set "values.mixer.policy.autoscaleEnabled=false" --set "components.policy.k8s.resources.limits.cpu=500m" --set "components.policy.k8s.resources.limits.memory=2048Mi" --set "components.policy.k8s.resources.requests.cpu=300m" --set "components.policy.k8s.resources.requests.memory=512Mi" --set "values.mixer.telemetry.autoscaleEnabled=false" --set "components.telemetry.k8s.resources.limits.cpu=500m" --set "components.telemetry.k8s.resources.limits.memory=2048Mi" --set "components.telemetry.k8s.resources.requests.cpu=300m" --set "components.telemetry.k8s.resources.requests.memory=512Mi" --set "values.pilot.autoscaleEnabled=false" --set "components.pilot.k8s.resources.limits.cpu=500m" --set "components.pilot.k8s.resources.limits.memory=1024Mi" --set "components.pilot.k8s.resources.requests.cpu=250m" --set "components.pilot.k8s.resources.requests.memory=512Mi" > padu.yaml
#echo 'istioctl manifest generate -f /etc/istio/istioOperator-draft.yaml --set "values.global.jwtPolicy=first-party-jwt"' > padu.yaml
#istioctl manifest generate -f /etc/istio/istioOperator-draft.yaml       --set "values.global.jwtPolicy=first-party-jwt" > padu.yaml
#printf "\n=== display new manifest =======================================================\n"
#cat padu.yaml
#printf "\n=== apply new manifest =========================================================\n"
#kubectl apply -f padu.yaml
#printf "\n=== apply new manifest again after a few seconds (so apiserver is ready to handle crds) =========================================================\n"
#sleep 3
#kubectl apply -f padu.yaml

istioctl manifest apply -f /etc/istio/istioOperator-draft.yaml

#while [ "$(kubectl get po -n istio-system -l app=sidecarInjectorWebhook -o jsonpath='{ .items[0].status.phase}')" != "Running" ]
#do
#  echo "sidecar injector still not running. Waiting..."
#  sleep 1s
#done
#echo "sidecar injector is running"
#echo "patching api-server destination rule"
#kubectl patch destinationrules.networking.istio.io -n istio-system api-server --type merge --patch '{"spec": {"trafficPolicy": { "connectionPool" : { "tcp": {"connectTimeout": "30s"}}}}}'

echo "Apply custom kyma manifests"
kubectl apply -f /etc/manifests
