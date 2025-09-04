# Integrate an External System with Kyma

## Introduction

The document describes the steps needed to connect an external system (for example, HTTPBin) with Kyma runtime.

In this example, Kyma sends authenticated requests to an external service.

## Prerequisites

1. Create Kyma runtime either by using the [SAP BTP cockpit](https://help.sap.com/docs/btp/sap-business-technology-platform/create-kyma-environment-instance) or by following the [Kyma Quick Install](https://kyma-project.io/#/02-get-started/01-quick-install) tutorial.
2. Besides the Kyma default modules like Istio and API Gateway, you must enable the following Kyma modules:
    * Application Connector: it includes the Application Gateway for proxying requests to external systems.
    * Serverless: used to run a Function that sends an HTTP request to an external system.
    If you created Kyma runtime using SAP BTP cockpit, follow this tutorial for [Adding and Deleting a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module?#add-and-delete-a-kyma-module-using-kyma-dashboard). Alternatively, continue with the steps described in the [Kyma Quick Install](https://kyma-project.io/#/02-get-started/01-quick-install?id=steps) tutorial.

## Integrate an External System with Kyma Runtime

1. [Create an Application custom resource (CR)](./01-10-create-application.md). The Application CR represents an external system and contains all information about exposed endpoints and their security configuration etc.
2. [Register a service for the external system](./01-20-register-manage-services.md) in the Application CR. The service is an abstraction of the external system. Kyma workloads can send their requests to the service URL, and the Application Gateway proxies these requests to the external system and handles all security-related aspects transparently (for example, establishing a trusted connection or authentication).
    * [Register a secured API](./01-30-register-secured-api.md). The Application Connector module supports many different authentication methods. In this step, you can find an example for each of them.
    * [Disable TLS certificate validation](./01-50-disable-tls-certificate-verification.md). For testing purposes, disabling the TLS certificate validation can be helpful. By default, the Application Connector module validates TLS certificates when establishing a secure connection.
3. [Call the external API](./01-40-call-registered-service-from-kyma.md). In this step, you can learn how to call an external system using a Kyma Function.
