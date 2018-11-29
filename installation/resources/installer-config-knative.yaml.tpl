---
apiVersion: v1
kind: ConfigMap
metadata:
  name: knative-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
data:
  global.knative: "true"