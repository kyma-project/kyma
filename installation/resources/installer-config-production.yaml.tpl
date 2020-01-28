apiVersion: v1
kind: ConfigMap
metadata:
  name: ory-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: ory
    kyma-project.io/installation: ""
data:
  postgresql.enabled: "true"
  hydra.hydra.autoMigrate: "true"
  global.ory.hydra.persitance.enabled: "true"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: istio
    kyma-project.io/installation: ""
data:
  global.proxy.resources.requests.cpu: "100m"
  global.proxy.resources.requests.memory: "128Mi"
  global.proxy.resources.limits.cpu: "500m"
  global.proxy.resources.limits.memory: "1024Mi"
  
  gateways.istio-ingressgateway.resources.requests.cpu: "100m"
  gateways.istio-ingressgateway.resources.requests.memory: "128Mi" 
  gateways.istio-ingressgateway.resources.limits.cpu: "2000m" 
  gateways.istio-ingressgateway.resources.limits.memory: "1024Mi"

  mixer.telemetry.resources.requests.cpu: "1000m"
  mixer.telemetry.resources.requests.memory: "1G"
  mixer.telemetry.resources.limits.cpu: "4800m"
  mixer.telemetry.resources.limits.memory: "4G"

  mixer.policy.resources.requests.memory: "256Mi"
  mixer.policy.resources.limits.memory: "512Mi"
  mixer.policy.resources.requests.cpu: "100m"
  mixer.policy.resources.limits.cpu: "500m"

  pilot.resources.requests.cpu: "500m"
  pilot.resources.requests.memory: "2048Mi"
  pilot.resources.limits.memory: "4G"
  pilot.resources.limits.cpu: "1000m"