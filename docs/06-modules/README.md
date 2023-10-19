## Kyma modules

Classic Kyma offered a fixed set of preconfigured components whose development rhythm is synchronized and determined by the release schedule.

With the modular approach, Kyma components become modules, each providing one functionality developed independently of the other ones. Each module has its own custom resource that holds the desired configuration and the operator that reconciles the configuration.

The Kyma project is currently in the transition phase. Some components are already independent modules, but others are still part of the big Kyma release and are installed with the `kyma deploy` command. With each successive release, fewer components will be available within preconfigured Kyma, but more and more will be offered as independent modules.

You can enable modules at any time. Give them a try! If you decide that some of them are not needed for your use case, you can disable them and free the resources. Learn how to [install, uninstall, and upgrade Kyma with a module](../02-get-started/08-install-uninstall-upgrade-kyma-module.md).

The table of Kyma modules:

> **NOTE:** The entries marked with "*" are still components that will be modularized soon.

| Module | Purpose |
|---|---|
| [SAP BTP Operator](https://kyma-project.io/#/btp-manager/user/README.md) | Within the SAP BTP Operator module, BTP Manager installs SAP BTP Service Operator that allows you to consume SAP BTP services from your Kubernetes cluster using Kubernetes-native tools. |
| [Keda](https://kyma-project.io//#/keda-manager/user/README.md) | The Keda module comes with Keda Manager, an extension to Kyma that allows you to install [KEDA (Kubernetes Event Driven Autoscaler)](https://keda.sh/). |
| [Serverless](https://kyma-project.io/#/serverless-manager/user/README.md) | With the Serverless module, you can define simple code snippets (Functions) with minimal implementation effort. |
| [Telemetry](https://kyma-project.io/#/telemetry-manager/user/README.md) | Enable telemetry agents to easily collect application logs and distributed traces for your application and dispatch them to backends.|
| [NATS](https://kyma-project.io/#/nats-manager/user/README.md) | NATS deploys a NATS cluster within the Kyma cluster. You can use it as a backend for Kyma Eventing. |
| [Eventing*](https://github.com/kyma-project/eventing-manager) | Eventing provides functionality to publish and subscribe to CloudEvents. <br> At the moment, the SAP Event Mesh default plan and NATS (provided by the NATS module) are supported. |
| [Application Connector*](https://github.com/kyma-project/application-connector-manager) | Application Connector allows you to connect with external solutions. No matter if you want to integrate an on-premise or a cloud system, the integration process doesn't change, which allows you to avoid any configuration or network-related problems. | 
| [API Gateway*](https://github.com/kyma-project/api-gateway) | API Gateway provides functionalities that allow you to expose and secure APIs. |
| [Istio*](https://github.com/kyma-project/istio) | Istio is a service mesh with the Kyma-specific configuration. |
