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
  url: ""
  components:
    - name: "cluster-essentials"
      namespace: "kyma-system"
    - name: "updatefun"
      namespace: "kyma-system"
