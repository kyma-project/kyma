apiVersion: v1
kind: Secret
metadata:
  name: cluster-certificate
  namespace: kyma-installer
type: Opaque
data:
  tls_cert: __TLS_CERT__
  tls_key: __TLS_KEY__
