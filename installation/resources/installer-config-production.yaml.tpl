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
  oathkeeper.oathkeeper.deployment.resources.limits.cpu: "800m"
  oathkeeper.oathkeeper.deployment.resources.requests.cpu: "200m"
  hpa.oathkeeper.minReplicas: "3"
  hpa.oathkeeper.maxReplicas: "10"
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
  global.proxy.resources.requests.cpu: "300m"
  global.proxy.resources.requests.memory: "128Mi"
  global.proxy.resources.limits.cpu: "500m"
  global.proxy.resources.limits.memory: "1024Mi"

  gateways.istio-ingressgateway.autoscaleMin: "3" 
  gateways.istio-ingressgateway.autoscaleMax: "10"
