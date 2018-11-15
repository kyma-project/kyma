---
apiVersion: v1
kind: ConfigMap
metadata:
  name: installer-feature-gates
  namespace: kyma-installer
  labels:
    installer: feature-gates
data:
  features: "__FEATURE_GATES__"