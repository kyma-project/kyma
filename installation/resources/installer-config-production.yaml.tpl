---
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
  oathkeeper.deployment.resources.limits.cpu: "800m"  
  oathkeeper.deployment.resources.requests.cpu: "200m"  
  hpa.oathkeeper.minReplicas: "3" 
  hpa.oathkeeper.maxReplicas: "10"  
  hydra.replicaCount: "2"
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
  kyma_istio_operator: |-
    apiVersion: install.istio.io/v1alpha1
    kind: IstioOperator
    metadata:
      namespace: istio-system
    spec:
      components:
        ingressGateways:
        - name: istio-ingressgateway
          k8s:
            hpaSpec:
              maxReplicas: 10
              minReplicas: 3
      values:
        global:
          proxy:
            resources:
              requests:
                cpu: 150m
                memory: 128Mi
              limits:
                cpu: 500m
                memory: 1024Mi
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
  prometheus.prometheusSpec.resources.limits.cpu: "1"
  prometheus.prometheusSpec.resources.limits.memory: "4Gi"
  prometheus.prometheusSpec.resources.requests.cpu: "300m"
  prometheus.prometheusSpec.resources.requests.memory: "1Gi"
  prometheus-istio.server.resources.limits.memory: "4Gi"
  alertmanager.alertmanagerSpec.retention: "240h"

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: logging-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: logging
    kyma-project.io/installation: ""
data:
  loki.resources.limits.memory: "512Mi"
  fluent-bit.resources.limits.memory: "256Mi"