---
title: Enable the Azure Broker for local deployment
type: Configuration
---
>**NOTE:** To enable the Azure Broker, you need a [Microsoft Azure](https://azure.microsoft.com/en-us) subscription.

By default, the Azure Broker is disabled for local installation and does not install along with other Kyma core components.
To enable the installation of the Azure Broker, export these Azure Broker-specific environment variables before you install Kyma:  

- `AZURE_BROKER_TENANT_ID`
- `AZURE_BROKER_SUBSCRIPTION_ID`
- `AZURE_BROKER_CLIENT_ID`
- `AZURE_BROKER_CLIENT_SECRET`

Export these variables using the details of your [Microsoft Azure](https://azure.microsoft.com/en-us) subscription, for example:
```
export AZURE_BROKER_TENANT_ID='{YOUR_TENANT_ID}'
```
