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
