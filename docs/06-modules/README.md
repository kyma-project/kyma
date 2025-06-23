# Kyma Modules

> [!WARNING]
> This is the open-source documentation for the Kyma project. If you are a managed offering user of **SAP BTP, Kyma runtime**, see [Kyma Environment](https://help.sap.com/docs/BTP/65de2977205c403bbc107264b8eccf4b/468c2f3c3ca24c2c8497ef9f83154c44.html) on SAP Help Portal.

With Kyma’s modular approach, you can install just the modules you need, instead of a predefined set of components. Each module has its own custom resource that holds the desired configuration and the operator that reconciles the configuration.

You can add modules at any time. If you decide that some of them are not needed for your use case, you can delete them and free the resources. To learn how to install a module, visit [Quick install](../02-get-started/01-quick-install.md). To learn how to quickly uninstall or upgrade Kyma with specific modules, visit [Uninstall and upgrade Kyma with a module](../02-get-started/08-uninstall-upgrade-kyma-module.md).

| Module | Purpose |
|---|---|
| [Istio](https://kyma-project.io/#/istio/user/README) | Istio is a service mesh with the Kyma-specific configuration. |
| [API Gateway](https://kyma-project.io/#/api-gateway/user/README) | API Gateway provides functionalities that allow you to expose and secure APIs. |
| [SAP BTP Operator](https://kyma-project.io/#/btp-manager/user/README.md) | Within the SAP BTP Operator module, BTP Manager installs SAP BTP Service Operator that allows you to consume SAP BTP services from your Kubernetes cluster using Kubernetes-native tools. |
| [Application Connector](https://kyma-project.io/#/application-connector-manager/user/README) | Application Connector allows you to connect with external solutions. No matter if you want to integrate an on-premise or a cloud system, the integration process doesn't change, which allows you to avoid any configuration or network-related problems. |
| [Cloud Manager](https://kyma-project.io/#/cloud-manager/user/README)| Cloud Manager brings hyperscaler products and resources into the Kyma cluster in a secure way. |
| [Docker Registry](https://kyma-project.io/#/docker-registry/user/README)| The Docker Registry module provides a lightweight, open-source Docker registry for storing and distributing container images in the Kubernetes environment. |
| [Eventing](https://kyma-project.io/#/eventing-manager/user/README) | Eventing provides functionality to publish and subscribe to CloudEvents. <br> At the moment, the SAP Event Mesh default plan and NATS (provided by the NATS module) are supported. |
| [Keda](https://kyma-project.io//#/keda-manager/user/README.md) | The Keda module comes with Keda Manager, an extension to Kyma that allows you to install [KEDA (Kubernetes Event Driven Autoscaler)](https://keda.sh/). |
| [NATS](https://kyma-project.io/#/nats-manager/user/README.md) | NATS deploys a NATS cluster within the Kyma cluster. You can use it as a backend for Kyma Eventing. |
| [Serverless](https://kyma-project.io/#/serverless-manager/user/README.md) | With the Serverless module, you can define simple code snippets (Functions) with minimal implementation effort. |
| [Telemetry](https://kyma-project.io/#/telemetry-manager/user/README.md) | Enable telemetry agents to collect application logs, distributed traces, and metrics for your application and dispatch them to backends.|
