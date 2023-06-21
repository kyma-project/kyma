## What is Kyma

Kyma is an opinionated set of Kubernetes-based modular building blocks, including all necessary capabilities to develop and run enterprise-grade cloud-native applications.
It is the open path to the SAP ecosystem supporting business scenarios end-to-end.

Kyma is an actively maintained open-source project supported by SAP. The Kyma project is also a foundation of Kubernetes Runtime which is a part of SAP Business Technology Platform (BTP). You can use Kyma modules in your own Kubernetes cluster, or try the managed version from SAP BTP with a ready-to-use Kubernetes cluster powered by Gardener. 

![overview](./assets/kyma-overview.svg)

## Modular Kyma

Classic Kyma offered a fixed set of preconfigured components whose development rhythm is synchronized and determined by the release schedule. With the modular approach, Kyma components become modules, each providing one functionality developed independently of the other ones. Each module has its own custom resource that holds the desired configuration and the operator that reconciles the configuration. Kyma project is currently in the transition phase. Some components are already independent modules, but others are still part of the big Kyma release and are installed with `kyma deploy` command. With each successive release, fewer components will be available within the preconfigured Kyma, but more and more will be offered as independent modules.

### Kyma modules 

> **NOTE:** The entries marked with "*" are still components that will be modularized soon.

| Module | Purpose |
|---|---|
| [BTP Operator]((https://kyma-docs.netlify.app//?basePath=https://raw.githubusercontent.com/kyma-project/btp-manager/aa50848013372806eaf2e707c217b8bed4eb09cb/docs/user/&homepage=README.md&sidebar=true&loadSidebar=_sidebar.md&browser-tab-title=BTP%20Operator%20Documentation#/)) |Within the BTP Operator module, BTP Manager installs SAP BTP Service Operator that allows you to consume SAP BTP services from your Kubernetes cluster using Kubernetes-native tools. |  |
| [Istio*](https://github.com/kyma-project/istio) | Istio is a service mesh with Kyma-specific configuration. | |
| [Serverless](https://github.com/kyma-project/serverless-manager) | With the Serverless module, you can define simple code snippets (Funtions) with minimal implementation effort. |  |
| [Telemetry*](https://github.com/kyma-project/telemetry-manager) | Enable telemetry agents to easily collect application logs and distributed traces for your application and dispatch them to backends.|   |
| [Eventing*](https://github.com/kyma-project/eventing-manager) | Eventing provides functionality to publish and subscribe to CloudEvents. <br> At the moment, the SAP Event Mesh default plan and NATS (provided by the NATS module) are supported. |  |
| [NATS*](https://github.com/kyma-project/nats-manager) |NATS deploys a NATS cluster within the Kyma cluster. You can use it as a backend for Kyma Eventing. |  |
| [Application Connector*](https://github.com/kyma-project/application-connector-manager) | Application Connector allows you to connect with external solutions. No matter if you want to integrate an on-premise or a cloud system, the integration process doesn't change, which allows you to avoid any configuration or network-related problems. | 
| API Gateway* |API Gateway provides functionalities that allow you to expose and secure APIs. |

### Kyma Beta modules:
> **NOTE:** The entries marked with "*" are still components that will be modularized soon.

| Module | Purpose |
|---|---|
| [Keda](https://kyma-docs.netlify.app//?basePath=https://raw.githubusercontent.com/kyma-project/keda-manager/main/docs/user/&homepage=README.md&sidebar=true&loadSidebar=_sidebar.md&browser-tab-title=Keda%20module%20Documentation#/) | The Keda module comes with Keda Manager, an extension to Kyma that allows you to install KEDA (Kubernetes Event Driven Autoscaler). |


## Kyma main areas

![areas](./assets/kyma-areas.svg)

To learn more about specific Kyma areas and functionalities, go to the respective sections.

Kyma is built upon leading cloud-native, open-source projects, such as Istio, NATS, Serverless, and Prometheus. The features developed by Kyma are the unique “glue” that holds them together, so you can connect and extend your applications easily and intuitively. To learn how to do that, head over to the [Get Started](./02-get-started) section where you can find step-by-step instructions to get your environment up and running.

The extensions and customizations you create are decoupled from your core applications, which adds to these general advantages of using Kyma:

![advantages](./assets/kyma-advantages.svg)
