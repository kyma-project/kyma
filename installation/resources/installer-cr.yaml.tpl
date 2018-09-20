apiVersion: "installer.kyma-project.io/v1alpha1"
kind: Installation
metadata:
  name: kyma-installation
  labels:
    action: install
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
    - name: "prometheus-operator"
      namespace: "kyma-system"
    - name: "provision-bundles"
    - name: "dex"
      namespace: "kyma-system"
    - name: "core"
      namespace: "kyma-system"
    - name: "remote-environments"
      namespace: "kyma-integration"
      release: "ec-default"
    - name: "remote-environments"
      namespace: "kyma-integration"
      release: "hmc-default"
    - name: "updatefun"
      namespace: "kyma-system"
