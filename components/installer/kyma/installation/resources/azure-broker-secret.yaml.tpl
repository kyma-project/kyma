apiVersion: v1
kind: Secret
metadata:
  name: azure-broker-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
type: Opaque
data:
  azure-broker.subscription_id: "__AZURE_BROKER_SUBSCRIPTION_ID__"
  azure-broker.tenant_id: "__AZURE_BROKER_TENANT_ID__"
  azure-broker.client_id: "__AZURE_BROKER_CLIENT_ID__"
  azure-broker.client_secret: "__AZURE_BROKER_CLIENT_SECRET__"