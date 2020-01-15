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
  global.ory.hydra.persitance.enabled: "true"
  # global.ory.hydra.persitance.user: "someUser"
  # global.ory.hydra.persitance.secretName: "my-secret"
  # global.ory.hydra.persitance.secretKey: "password"
  # global.ory.hydra.persitance.dbUrl: "mydb.mynamespace.svc.cluster.local:1234"
  # global.ory.hydra.persitance.dbName: "db4hydra"
