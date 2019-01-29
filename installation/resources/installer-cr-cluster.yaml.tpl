apiVersion: "installer.kyma-project.io/v1alpha1"
kind: Installation
metadata:
  name: kyma-installation
  finalizers:
    - finalizer.installer.kyma-project.io
spec:
  version: "__VERSION__"
  url: "__URL__"
  components: 
    - name: "cluster-essentials"
      namespace: "kyma-system"
    - name: "istio"
      namespace: "istio-system"
    - name: "nats-streaming"
      namespace: "natss"
    - name: "knative"
      namespace: "istio-system" # KNative comes with namespaces hardcoded so this one is only for installer
    - name: "istio-kyma-patch"
      namespace: "istio-system"
    - name: "prometheus-operator"
      namespace: "kyma-system"
    - name: "dex"
      namespace: "kyma-system"
    - name: "service-catalog"
      namespace: "kyma-system"
    - name: "core"
      namespace: "kyma-system"
    - name: "application-connector"
      namespace: "kyma-system"
    - name: "ark"
      namespace: "heptio-ark"
    - name: "logging"
      namespace: "kyma-system"
    - name: "jaeger"
      namespace: "kyma-system"
    - name: "monitoring"
      namespace: "kyma-system"
