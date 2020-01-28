apiVersion: "installer.kyma-project.io/v1alpha1"
kind: Installation
metadata:
  name: kyma-installation
  labels:
    action: install
    kyma-project.io/installation: ""
  finalizers:
    - finalizer.installer.kyma-project.io
spec:
  version: "__VERSION__"
  url: "__URL__"
  components:
    - name: "cluster-essentials"
      namespace: "kyma-system"
    - name: "testing"
      namespace: "kyma-system"
    - name: "istio-init"
      namespace: "istio-system"
    - name: "istio"
      namespace: "istio-system"
    - name: "xip-patch"
      namespace: "kyma-installer"
    - name: "istio-kyma-patch"
      namespace: "istio-system"
    - name: "knative-serving-init"
      namespace: "knative-serving"
    - name: "knative-serving"
      namespace: "knative-serving"
   # - name: "knative-build-init"
   #   namespace: "knative-build"
   # - name: "knative-build"
   #   namespace: "knative-build"
    - name: "knative-eventing"
      namespace: "knative-eventing"
    - name: "dex"
      namespace: "kyma-system"
    - name: "ory"
      namespace: "kyma-system"
    - name: "api-gateway"
      namespace: "kyma-system"
    - name: "service-catalog"
      namespace: "kyma-system"
    - name: "service-catalog-addons"
      namespace: "kyma-system"
    - name: "helm-broker"
      namespace: "kyma-system"
    - name: "nats-streaming"
      namespace: "natss"
    - name: "rafter"
      namespace: "kyma-system"
    - name: "core"
      namespace: "kyma-system"
    - name: "permission-controller"
      namespace: "kyma-system"
    - name: "knative-provisioner-natss"
      namespace: "knative-eventing"
    - name: "event-bus"
      namespace: "kyma-system"
    - name: "event-sources"
      namespace: "kyma-system"
    - name: "application-connector-ingress"
      namespace: "kyma-system"
    - name: "application-connector-helper"
      namespace: "kyma-integration"
    - name: "application-connector"
      namespace: "kyma-integration"
    - name: "backup-init"
      namespace: "kyma-system"
    - name: "backup"
      namespace: "kyma-system"
    - name: "logging"
      namespace: "kyma-system"
    - name: "jaeger"
      namespace: "kyma-system"
    - name: "monitoring"
      namespace: "kyma-system"
    - name: "kiali"
      namespace: "kyma-system"
    #- name: "compass"
    #  namespace: "compass-system"
    #- name: "compass-runtime-agent"
    #  namespace: "compass-system"
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
