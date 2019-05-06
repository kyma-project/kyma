apiVersion: "installer.kyma-project.io/v1alpha1"
kind: Installation
metadata:
  name: kyma-installation
  labels:
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
    # - name: "knative-serving"
    #   namespace: "knative-serving"
    # - name: "knative-eventing"
    #   namespace: "knative-eventing"
    # - name: "prometheus-operator"
    # namespace: "kyma-system"
    # - name: "dex"
    #   namespace: "kyma-system"
    # - name: "service-catalog"
    #   namespace: "kyma-system"
    # - name: "service-catalog-addons"
    #   namespace: "kyma-system"
    # - name: "helm-broker"
    #   namespace: "kyma-system"
    # - name: "nats-streaming"
    #   namespace: "natss"
    # - name: "assetstore"
    #   namespace: "kyma-system"
    # - name: "cms"
    #   namespace: "kyma-system"
    # - name: "core"
    #   namespace: "kyma-system"
    # - name: "knative-provisioner-natss"
    #   namespace: "knative-eventing"
    # - name: "event-bus"
    #   namespace: "kyma-system"
    # - name: "application-connector-ingress"
    #   namespace: "kyma-system"
    # - name: "application-connector-helper"
    #   namespace: "kyma-integration"
    # - name: "application-connector"
    #   namespace: "kyma-integration"
    #- name: "ark"
    #  namespace: "heptio-ark"
    #- name: "logging"
    #  namespace: "kyma-system"
    #- name: "jaeger"
    #  namespace: "kyma-system"
    #- name: "monitoring"
    #  namespace: "kyma-system"
