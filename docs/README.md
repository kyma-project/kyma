## What is Kyma

Kyma is an opinionated set of Kubernetes-based modular building blocks, including all necessary capabilities to develop and run enterprise-grade cloud-native applications.
It is the open path to the SAP ecosystem supporting business scenarios end-to-end.

![overview](assets/modular-kyma.png)

Kyma is an actively maintained open-source project supported by SAP. The Kyma project is also a foundation of SAP BTP, Kyma runtime which is a part of SAP Business Technology Platform (BTP). You can use Kyma modules in your own Kubernetes cluster, or try the managed version from SAP BTP with a ready-to-use Kubernetes cluster powered by Gardener. 



## Kyma modules 

Classic Kyma offered a fixed set of preconfigured components whose development rhythm is synchronized and determined by the release schedule. 

With the modular approach, Kyma components become modules, each providing one functionality developed independently of the other ones. Each module has its own custom resource that holds the desired configuration and the operator that reconciles the configuration. 

The Kyma project is currently in the transition phase. Some components are already independent modules, but others are still part of the big Kyma release and are installed with the `kyma deploy` command. With each successive release, fewer components will be available within preconfigured Kyma, but more and more will be offered as independent modules.

You can enable modules at any time. Give them a try! If you decide that some of them are not needed for your use case, you can disable them and free the resources. Learn how to [enable, disable, and upgrade a module](./04-operation-guides/operations/08-enable-disable-upgrade-kyma-module.md).

The table of Kyma modules:

> **NOTE:** The entries marked with "*" are still components that will be modularized soon.

| Module | Purpose |
|---|---|
| [BTP Operator](https://kyma-project.github.io/kyma/#/btp-manager/README.md) | Within the BTP Operator module, BTP Manager installs SAP BTP Service Operator that allows you to consume SAP BTP services from your Kubernetes cluster using Kubernetes-native tools. |
| [Keda](https://kyma-project.github.io/kyma/#/keda-manager/user/README.md) | The Keda module comes with Keda Manager, an extension to Kyma that allows you to install [KEDA (Kubernetes Event Driven Autoscaler)](https://keda.sh/). |
| [Istio*](https://github.com/kyma-project/istio) | Istio is a service mesh with the Kyma-specific configuration. |
| [Serverless*](https://kyma-project.github.io/kyma/#/serverless-manager/user/README.md) | With the Serverless module, you can define simple code snippets (Functions) with minimal implementation effort. |
| [Telemetry*](https://kyma-project.github.io/kyma/#/telemetry-manager/user/README.md) | Enable telemetry agents to easily collect application logs and distributed traces for your application and dispatch them to backends.|
| [Eventing*](https://github.com/kyma-project/eventing-manager) | Eventing provides functionality to publish and subscribe to CloudEvents. <br> At the moment, the SAP Event Mesh default plan and NATS (provided by the NATS module) are supported. |
| [NATS*](https://github.com/kyma-project/nats-manager) | NATS deploys a NATS cluster within the Kyma cluster. You can use it as a backend for Kyma Eventing. |
| [Application Connector*](https://github.com/kyma-project/application-connector-manager) | Application Connector allows you to connect with external solutions. No matter if you want to integrate an on-premise or a cloud system, the integration process doesn't change, which allows you to avoid any configuration or network-related problems. | 
| [API Gateway*](https://github.com/kyma-project/api-gateway) | API Gateway provides functionalities that allow you to expose and secure APIs. |

## Kyma's strengths
Kyma is built upon leading cloud-native, open-source projects and open standards, such as Istio, NATS, Cloud Events, Open Telemetry, and Prometheus. We created an opinionated set of modules you can easily enable in your Kubernetes cluster to speed up cloud application development and operations. With Kyma, you save the time to pick the right tools and the effort to keep them secure and up to date. Also, you can use the modules you need from Kyma and complement them with other Kubernetes tools.

Kyma is a Kubernetes-based application runtime with several extensions, not a full-blown platform. The extensions make Kyma more attractive for developers who want to focus on business logic and limit investment in technical services and infrastructure. Kyma is part of SAP Business Technology Platform and offers easy integration with BTP services and other SAP systems. 
 
Kyma has been open-source since 2018 and part of SAP BTP since 2019. We believe that openness and vendor independence is a valuable proposition. Even though Kyma is a part of the commercial product, SAP BTP, Kyma runtime, it will remain an open project. We also believe that offering an open-source project as a commercial product only benefits both parties. Open-source users get the confidence that the project won't be abandoned anytime soon, and customers see the quality and technical details of the product. Apart from that, SAP strongly supports the open-source community. For more information, visit [SAP Open Source](https://community.sap.com/topics/open-source).

## Kyma and SAP BTP, Kyma runtime
There's a difference between the open-source Kyma project and SAP BTP, Kyma runtime (SKR). SKR is a bundle of a Kubernetes cluster powered by Gardener and Kyma modules provided as a managed service. All the components are regularly updated, and the availability is monitored and guaranteed by Service-Level Agreement (SLA). SKR is also preconfigured to easily connect to other SAP services and systems. Using SKR, you can face some limitations in configuring Kyma components because some settings are managed centrally and overwrite user changes. But still, you get the admin access to the cluster, that is, the **cluster-admin** role. If you use Kyma open-source components, you have more control and flexibility over installation, configuration, and upgrade processes but more operations-related responsibilities.
