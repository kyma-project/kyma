# Kyma Modules

Classic Kyma offered a fixed set of preconfigured components whose development rhythm was synchronized and determined by the release schedule.

With the modular approach, Kyma components became independent modules, each providing one functionality developed independently of the other ones. Each module has its own custom resource that holds the desired configuration and the operator that reconciles the configuration.

You can enable modules at any time. If you decide that some of them are not needed for your use case, you can disable them and free the resources. To learn how to install a module, visit [Quick install](../02-get-started/01-quick-install.md). To learn how to quickly uninstall or upgrade Kyma with specific modules, visit [Uninstall and upgrade Kyma with a module](../02-get-started/08-uninstall-upgrade-kyma-module.md).

| Module | Purpose |
|---|---|
| [Istio](https://github.com/kyma-project/istio) | Istio is a service mesh with the Kyma-specific configuration. |
| [SAP BTP Operator](https://kyma-project.io/#/btp-manager/user/README.md) | Within the SAP BTP Operator module, BTP Manager installs SAP BTP Service Operator that allows you to consume SAP BTP services from your Kubernetes cluster using Kubernetes-native tools. |
| [Application Connector](https://github.com/kyma-project/application-connector-manager) | Application Connector allows you to connect with external solutions. No matter if you want to integrate an on-premise or a cloud system, the integration process doesn't change, which allows you to avoid any configuration or network-related problems. |
| [Keda](https://kyma-project.io//#/keda-manager/user/README.md) | The Keda module comes with Keda Manager, an extension to Kyma that allows you to install [KEDA (Kubernetes Event Driven Autoscaler)](https://keda.sh/). |
| [Serverless](https://kyma-project.io/#/serverless-manager/user/README.md) | With the Serverless module, you can define simple code snippets (Functions) with minimal implementation effort. |
| [Telemetry](https://kyma-project.io/#/telemetry-manager/user/README.md) | Enable telemetry agents to collect application logs, distributed traces, and metrics for your application and dispatch them to backends.|
| [NATS](https://kyma-project.io/#/nats-manager/user/README.md) | NATS deploys a NATS cluster within the Kyma cluster. You can use it as a backend for Kyma Eventing. |
| [Eventing](https://github.com/kyma-project/eventing-manager) | Eventing provides functionality to publish and subscribe to CloudEvents. <br> At the moment, the SAP Event Mesh default plan and NATS (provided by the NATS module) are supported. |
| [API Gateway](https://github.com/kyma-project/api-gateway) | API Gateway provides functionalities that allow you to expose and secure APIs. |
