apiVersion: "installer.kyma.cx/v1alpha1"
kind: Installation
metadata:
  name: kyma-installation
  labels:
    action: install
  finalizers:
    - finalizer.installer.kyma.cx
spec:
  version: "__VERSION__"
  url: "__URL__"
  components:
    - name: "cluster-prerequisites"
    - name: "cluster-essentials"
    - name: "istio"
    - name: "prometheus-operator"
    - name: "provision-bundles"
    - name: "dex"
    - name: "core"
    - name: "remote-environments" 
