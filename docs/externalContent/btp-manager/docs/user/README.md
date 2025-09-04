# SAP BTP Operator Module

Use the SAP BTP Operator module to enable service management and consume SAP BTP services from your Kyma cluster.

## What is SAP BTP Operator?

The SAP BTP Operator module provides service management, which allows you to consume [SAP BTP services](https://discovery-center.cloud.sap/protected/index.html#/viewServices) from your Kyma cluster using Kubernetes-native tools.
Within the SAP BTP Operator module, [BTP Manager](https://github.com/kyma-project/btp-manager) installs the [SAP BTP service operator](https://github.com/SAP/sap-btp-service-operator/blob/main/README.md).
The SAP BTP service operator enables provisioning and managing service instances and service bindings of SAP BTP services. Consequently, your Kubernetes-native applications can access and use the services from your cluster.

## Features

The SAP BTP Operator module provides the following features:

* Credentials and access preconfiguration: Your Secret is available on Kyma runtime creation. See [Preconfigured Credentials and Access](03-10-preconfigured-secret.md).
* Customization of the default credentials and access: You can customize the default `sap-btp-manager` Secret. See [Customize the Default Credentials and Access](03-11-customize_secret.md).
* Multitenancy: You can configure multiple subaccounts in a single cluster. See [Working with Multiple Subaccounts](03-20-multitenancy.md).
* Lifecycle management of service instances and service bindings: You can create, update, and delete service instances and service bindings. See [Create Service Instances and Service Bindings](03-30-create-instances-and-bindings.md), [Update Service Instances](03-31-update-service-instances.md), and [Delete Service Bindings and Service Instances](03-32-delete-bindings-and-instances.md).
* Service binding rotation: You can enhance security by automatically rotating the credentials associated with your service bindings. See [Rotating Service Bindings](03-40-service-binding-rotation.md).
* Service binding Secret formatting: You can use different attributes in your ServiceBinding resource to generate different formats of your Secret resources. See [Formatting Service Binding Secrets](03-50-formatting-service-binding-secret.md).

## Scope

By default, the SAP BTP Operator module consumes SAP BTP services from the subaccount in which you created it. To consume the services from multiple subaccounts in a single Kyma cluster, configure a dedicated Secret for each additional subaccount.

## Architecture

The SAP BTP Operator module provides and retrieves the initial credentials that your application needs to use an SAP BTP service.

![SAP BTP Operator architecture](../assets/module_architecture.drawio.svg)

Depending on the number of configured Secrets, SAP BTP Operator can have access to multiple subaccounts within your cluster.

![Access configuration](../assets/access_configuration.drawio.svg)

For more information on multitenancy, see [Working with Multiple Subaccounts](03-20-multitenancy.md).

### SAP BTP, Kyma Runtime

SAP BTP, Kyma runtime runs on a Kubernetes cluster and wraps the SAP BTP Operator module, API server, and one or more applications. The application or multiple applications are the actual consumers of SAP BTP services.

### BTP Manager

BTP Manager is an operator based on the [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) framework. It extends the Kubernetes API by providing the [BtpOperator Custom Resource Definition (CRD)](https://github.com/kyma-project/btp-manager/blob/main/config/crd/bases/operator.kyma-project.io_btpoperators.yaml). 
BTP Manager performs the following operations:

* Manages the lifecycle of the SAP BTP service operator, including provisioning of the latest version, updating, and deprovisioning.
* Configures the SAP BTP service operator.

### SAP BTP Service Operator

The SAP BTP service operator is an open-source component that allows you to connect SAP BTP services to your cluster and manage them using Kubernetes-native tools. It is responsible for communicating with SAP Service Manager. The operator's resources (service instances and service bindings) come from the `services.cloud.sap.com` API group.

### SAP Service Manager

[SAP Service Manager](https://help.sap.com/docs/service-manager/sap-service-manager/sap-service-manager?locale=en-US) is the central registry for service brokers and platforms in SAP BTP, which enables you to do the following:

* Consume platform services in any connected runtime environment.
* Track the creation and management of service instances.
* Share services and service instances between different runtimes.

SAP Service Manager uses [Open Service Broker API](https://www.openservicebrokerapi.org/) (OSB API) to communicate with service brokers.

### Service Brokers

Service brokers manage the lifecycle of services. SAP Service Manager interacts with service brokers using OSB API to provision and manage service instances and service bindings.

## API/Custom Resource Definitions

The `btpoperators.operator.kyma-project.io` Custom Resource Definition (CRD) describes the kind and the data format that SAP BTP Operator uses to configure resources.

See the documentation related to the BtpOperator custom resource (CR):

* [SAP BTP Operator Custom Resource](./resources/02-10-sap-btp-operator-cr.md)
* [Service Instance Custom Resource](./resources/02-20-service-instance-cr.md)
* [Service Binding Custom Resource](./resources/02-30-service-binding-cr.md)


## Resource Consumption

To learn more about the resources used by the SAP BTP Operator module, see [Kyma Modules' Sizing](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/kyma-modules-sizing?locale=en-US&state=DRAFT&version=Internal#sap-btp-operator).
