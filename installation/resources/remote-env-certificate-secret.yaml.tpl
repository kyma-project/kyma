apiVersion: v1
kind: Secret
metadata:
  name: remote-env-certificate
  namespace: kyma-installer
type: Opaque
data:
  remote_env_ca: __REMOTE_ENV_CA__
  remote_env_ca_key: __REMOTE_ENV_CA_KEY__
