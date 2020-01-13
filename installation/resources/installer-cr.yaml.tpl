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
      source:
        url: https://github.com/kyma-project/kyma.git//resources/cluster-essentials
    - name: "testing"
      namespace: "kyma-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip-padu//kyma-release-1.9/resources/testing
    - name: "istio-init"
      namespace: "istio-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/istio-init
    - name: "istio"
      namespace: "istio-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/istio
    - name: "xip-patch"
      namespace: "kyma-installer"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/xip-patch
    - name: "istio-kyma-patch"
      namespace: "istio-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/istio-kyma-patch
    - name: "knative-serving-init"
      namespace: "knative-serving"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/knative-serving-init
    - name: "knative-serving"
      namespace: "knative-serving"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/knative-serving
#    - name: "knative-build-init"
#      namespace: "knative-build"
#    - name: "knative-build"
#      namespace: "knative-build"
    - name: "knative-eventing"
      namespace: "knative-eventing"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/knative-eventing
    - name: "dex"
      namespace: "kyma-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/dex
    - name: "ory"
      namespace: "kyma-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/ory
    - name: "api-gateway"
      namespace: "kyma-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/api-gateway
    - name: "service-catalog"
      namespace: "kyma-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/service-catalog
    - name: "service-catalog-addons"
      namespace: "kyma-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/service-catalog-addons
    - name: "helm-broker"
      namespace: "kyma-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/helm-broker
    - name: "nats-streaming"
      namespace: "natss"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/nats-streaming
    - name: "rafter"
      namespace: "kyma-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/rafter
    - name: "core"
      namespace: "kyma-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/core
    - name: "knative-provisioner-natss"
      namespace: "knative-eventing"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/knative-provisioner-natss
    - name: "event-bus"
      namespace: "kyma-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/event-bus
    - name: "event-sources"
      namespace: "kyma-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/event-sources
    - name: "application-connector-ingress"
      namespace: "kyma-system"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/application-connector-ingress
    - name: "application-connector-helper"
      namespace: "kyma-integration"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/application-connector-helper
    - name: "application-connector"
      namespace: "kyma-integration"
      source:
        url: https://github.com/kyma-project/kyma/archive/release-1.9.zip//kyma-release-1.9/resources/application-connector
#    - name: "backup-init"
#      namespace: "kyma-system"
#    - name: "backup"
#      namespace: "kyma-system"
#    - name: "logging"
#      namespace: "kyma-system"
#    - name: "jaeger"
#      namespace: "kyma-system"
#    - name: "monitoring"
#      namespace: "kyma-system"
#    - name: "kiali"
#      namespace: "kyma-system"
#    - name: "compass"
#      namespace: "compass-system"
#    - name: "compass-runtime-agent"
#      namespace: "compass-system"
