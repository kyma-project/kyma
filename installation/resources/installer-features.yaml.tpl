---
apiVersion: v1
kind: ConfigMap
metadata:
  name: installer-features
  namespace: kyma-installer
  labels:
    installer: features
data:
  features: "__FEATURES__"