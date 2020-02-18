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
    - name: "dex"
      namespace: "kyma-system"
    - name: "ory"
      namespace: "kyma-system"
    - name: "compass"
      namespace: "compass-system"
    - name: "monitoring"
      namespace: "kyma-system"
    - name: "logging"
      namespace: "kyma-system"