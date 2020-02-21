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
  global.ory.hydra.persistence.enabled: "true"
  global.ory.hydra.persistence.postgresql.enabled: "true"
  global.ory.hydra.persistence.gcloud.enabled: "false"
  hydra.hydra.autoMigrate: "true"
  oathkeeper.deployment.resources.limits.cpu: "800m"
  oathkeeper.deployment.resources.requests.cpu: "200m"
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
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: core-cbs-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: core
    kyma-project.io/installation: ""
data:
  console-backend-service.replicaCount: 2
