apiVersion: "installer.kyma-project.io/v1alpha1"
kind: Installation
metadata:
  name: kyma-installation
  namespace: default
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
   # - name: "knative-eventing-channel-kafka"
   #   namespace: knative-eventing
   #   source:
   #     # NOTE: this is a go-getter compatible url
   #     url: github.com/kyma-incubator/knative-kafka.git//resources/knative-kafka?ref=master
   # - name: "knative-eventing-channel-kafka-tests"
   #   namespace: knative-eventing
    - name: "dex"
      namespace: "kyma-system"
    - name: "ory"
      namespace: "kyma-system"
    - name: "api-gateway"
      namespace: "kyma-system"
    - name: "rafter"
      namespace: "kyma-system"
    - name: "service-catalog"
      namespace: "kyma-system"
    - name: "service-catalog-addons"
      namespace: "kyma-system"
    - name: "helm-broker"
      namespace: "kyma-system"
    - name: "nats-streaming"
      namespace: "natss"
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
