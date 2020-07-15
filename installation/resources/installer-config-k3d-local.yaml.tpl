apiVersion: v1
kind: ConfigMap
metadata:
  name: net-global-overrides
  labels:
    installer: overrides
    kyma-project.io/installation: ""
data:
  global.domainName: local.kyma.pro
  global.environment.gardener: "false"
  global.ingress.domainName: local.kyma.pro
  global.ingress.tlsCrt: ZHVtbXkK
  global.ingress.tlsKey: ZHVtbXkK
  global.isLocalEnv: "true"
  global.minikubeIP: 127.0.0.1
  creationTimestamp: null
---
apiVersion: v1
data:
  global.ory.hydra.persistence.enabled: "false"
  global.ory.hydra.persistence.postgresql.enabled: "false"
  hydra.hydra.autoMigrate: "false"
kind: ConfigMap
metadata:
  name: ory-overrides
  labels:
    component: ory
    installer: overrides
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: serverless-overrides
  namespace: kyma-installer
  labels:
    component: serverless
    installer: overrides
data:
  dockerRegistry.enableInternal: "false"
  dockerRegistry.registryAddress: registry.localhost:5000
  dockerRegistry.serverAddress: registry.localhost:5000
  global.ingress.domainName: local.kyma.pro
