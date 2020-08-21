apiVersion: v1
kind: Secret
metadata:
  name: application-connector-certificate-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
type: Opaque
data:
  global.applicationConnectorCa: ""
  global.applicationConnectorCaKey: ""
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: installation-config-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
data:
  global.isLocalEnv: "true"
  global.domainName: "kyma.local"
  global.adminPassword: ""
  global.minikubeIP: ""
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
  pilot.resources.limits.memory: 1024Mi
  pilot.resources.limits.cpu: 500m
  pilot.resources.requests.memory: 512Mi
  pilot.resources.requests.cpu: 250m
  pilot.autoscaleEnabled: "false"

  mixer.policy.resources.limits.memory: 2048Mi
  mixer.policy.resources.limits.cpu: 500m
  mixer.policy.resources.requests.memory: 512Mi
  mixer.policy.resources.requests.cpu: 300m

  mixer.telemetry.resources.limits.memory: 2048Mi
  mixer.telemetry.resources.limits.cpu: 500m
  mixer.telemetry.resources.requests.memory: 512Mi
  mixer.telemetry.resources.requests.cpu: 300m
  mixer.loadshedding.mode: disabled

  mixer.policy.autoscaleEnabled: "false"
  mixer.telemetry.autoscaleEnabled: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: helm-broker-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: helm-broker
    kyma-project.io/installation: ""
data:
  global.isDevelopMode: "true" # global, because subcharts also use it
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: dex-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: dex
    kyma-project.io/installation: ""
data:
  telemetry.enabled: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: application-connector-tests
  namespace: kyma-installer
  labels:
    installer: overrides
    component: application-connector
    kyma-project.io/installation: ""
data:
  application-operator.tests.enabled: "false"
  application-registry.tests.enabled: "false"
  connector-service.tests.enabled: "false"
  tests.application_connector_tests.enabled: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: core-tests
  namespace: kyma-installer
  labels:
    installer: overrides
    component: core
    kyma-project.io/installation: ""
data:
  console.test.acceptance.enabled: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: compass-runtime-agent-tests
  namespace: kyma-installer
  labels:
    installer: overrides
    component: compass-runtime-agent
    kyma-project.io/installation: ""
data:
  compassRuntimeAgent.tests.enabled: "false"
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
  global.ory.hydra.persistence.enabled: "false"
  global.ory.hydra.persistence.postgresql.enabled: "false"
  hydra.hydra.autoMigrate: "false"
  hydra.deployment.resources.requests.cpu: "50m"
  hydra.deployment.resources.limits.cpu: "150m"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: tracing-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: tracing
    kyma-project.io/installation: ""
data:
  jaeger.spec.strategy: "allInOne"
  jaeger.spec.storage.type: "memory"
  jaeger.spec.storage.options.memory.max-traces: "10000"
  jaeger.spec.resources.limits.memory: "150Mi"
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
  alertmanager.alertmanagerSpec.resources.limits.cpu: "50m"
  alertmanager.alertmanagerSpec.resources.limits.memory: "100Mi"
  alertmanager.alertmanagerSpec.resources.requests.cpu: "20m"
  alertmanager.alertmanagerSpec.resources.requests.memory: "50Mi"
  alertmanager.alertmanagerSpec.retention: "1h"
  prometheus.prometheusSpec.resources.limits.cpu: "150m"
  prometheus.prometheusSpec.resources.limits.memory: "800Mi"
  prometheus.prometheusSpec.resources.requests.cpu: "100m"
  prometheus.prometheusSpec.resources.requests.memory: "200Mi"
  prometheus.prometheusSpec.retention: "2h"
  prometheus.prometheusSpec.retentionSize: "500MB"
  prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage: "1Gi"
  grafana.persistence.enabled: "false"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: serverless-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: serverless
    kyma-project.io/installation: ""
data:
  containers.manager.envs.buildRequestsCPU.value: "100m"
  containers.manager.envs.buildRequestsMemory.value: "200Mi"
  containers.manager.envs.buildLimitsCPU.value: "200m"
  containers.manager.envs.buildLimitsMemory.value: "400Mi"
  # TODO: Solve a problem with DNS
  tests.enabled: "false"

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: knative-serving-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: knative-serving
    kyma-project.io/installation: ""
data:
  networking_istio.resources.requests.cpu: "10m"
  networking_istio.resources.requests.memory: "100Mi"
  activator.resources.requests.cpu: "100m"
  activator.resources.requests.memory: "100Mi"
  autoscaler.resources.requests.cpu: "10m"
  autoscaler.resources.requests.memory: "100Mi"
  autoscaler_hpa.resources.requests.cpu: "10m"
  autoscaler_hpa.resources.requests.memory: "100Mi"
  controller.resources.requests.cpu: "30m"
  controller.resources.requests.memory: "100Mi"
