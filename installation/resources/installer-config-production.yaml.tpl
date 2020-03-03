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
  name: monitoring-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: monitoring
    kyma-project.io/installation: ""
data:
  prometheus.prometheusSpec.retentionSize: "15GB"
  prometheus.prometheusSpec.retention: "30d"
  prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage: "20Gi"
  prometheus.prometheusSpec.resources.limits.cpu: "600m"
  prometheus.prometheusSpec.resources.limits.memory: "2Gi"
  prometheus.prometheusSpec.resources.requests.cpu: "300m"
  prometheus.prometheusSpec.resources.requests.memory: "1Gi"
  alertmanager.alertmanagerSpec.retention: "240h"
