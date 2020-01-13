apiVersion: v1
kind: ConfigMap
metadata:
  name: ory-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: ory
    kyma-project.io/installation: ""
data:
  postgresql.enabled: "true"
  hydra.hydra.autoMigrate: "true"