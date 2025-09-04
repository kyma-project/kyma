# Istio Module
Use the Istio module to manage and configure the Istio service mesh.

## What Is Istio?

Istio is an open-source service mesh that provides a uniform way to manage, connect, and secure microservices. It helps to manage traffic, enhance security capabilities, and provide telemetry data for understanding service behavior. See the [open-source Istio documentation](https://istio.io/latest/docs/).

The Istio module installs and manages Istio in your Kyma cluster. By default, the Istio module is automatically added when you create a Kyma runtime instance.

## Features

The Istio module offers the following features:
- **Management of Istio installation and upgrades**: The module installs Istio and simplifies the process of managing its installation, reducing the complexity and time required for maintenance.
- **Default Istio configuration**: You can quickly have Istio installed with default settings.
- **Fine-tuning capabilities**: You can optimize Istio and fine-tune its settings according to specific performance or operational requirements.
- **Synchronization of the data plane with the Istio control plane**: This ensures that changes you make to the control plane are consistently reflected in the data plane, ensuring that the network operates consistently and reliably without any discrepancies.
- **Support for the X-Forwarded-For (XFF) header**: You can configure the XFF header to manage and track the source of incoming requests.

## Architecture

![Istio Operator Architecture](../assets/istio-controller-overview-user.svg)

### Istio Operator

Within the Istio module, Istio Operator handles the management and configuration of the Istio service mesh. It contains one controller, referred to as Istio Controller.

### Istio Controller

Istio Controller manages Istio and its resources as defined in the Istio custom resource (CR). Istio Controller is responsible for:
- Installing, upgrading, and uninstalling Istio
- Restarting workloads that have Istio sidecar proxy injected to ensure that these workloads are using the correct version of Istio

## API / Custom Resource Definitions

The `istios.operator.kyma-project.io` CustomResourceDefinition (CRD) describes the Istio CR that Istio Controller uses to manage the installation of Istio. See [Istio Custom Resource](https://kyma-project.io/#/istio/user/04-00-istio-custom-resource?id=istio-custom-resource).

## Resource Consumption

To learn more about the resources used by the Istio module, see [Kyma's Modules Sizing](https://help.sap.com/docs/btp/sap-business-technology-platform/kyma-modules-sizing?version=Cloud#istio).
