# Application Connector Module

## What is Application Connectivity in Kyma?

Application Connectivity in Kyma simplifies the interaction between external systems and your Kyma workloads. The main benefits are:

* Smooth and loosely coupled integration of external systems with Kyma workloads using [Kyma Eventing](https://kyma-project.io/#/eventing-manager/user/README)
* Easy consumption of SAP BTP services by supporting the SAP Extensibility approach
* Establishing high-security standards for any interaction between systems by using trusted communication channels and authentication methods
* Reducing configuration changes for Kyma workloads through encapsulating configuration details of external API endpoints
* Monitoring and tracing capabilities to facilitate operational aspects

## Application Connector Module

The Application Connector module bundles all features of Application Connectivity in Kyma. You can install and manage the module using Kyma dashboard.

The module includes Kubernetes operators and is fully configurable over its own Kubernetes custom resources (CRs). For each external system, a dedicated configuration is used. This allows for individual configuration of security aspects (like encryption and authentication) per system.

Besides proxying any ingress and egress requests to external systems and dealing with security concerns, it also includes full integration with SAP BTP Unified Customer Landscape (UCL) to simplify the consumption of SAP BTP services.

## Features

The Application Connector module provides the following features:

* Easy installation of Kyma's Application Connectivity capabilities by enabling the Application Connector module in your Kyma runtime.
* Simple configuration using Kubernetes CRs and easy management with Kyma dashboard.
* Full integration of SAP BTP's UCL service, which implements the SAP Extensibility concept. This allows for the automated integration of external systems registered in the UCL service.
* Dispatching of incoming requests from external systems to Kyma workloads (for example, a [Kyma Serverless Function](https://kyma-project.io/#/serverless-manager/user/resources/06-10-function-cr)) by using an Istio Gateway with mTLS and the [Kyma Eventing module](https://kyma-project.io/#/eventing-manager/user/README).
* Proxying outgoing requests to external APIs and transparently covering security requirements like encryption and authentication (like OAuth 2.0 + mTLS, Basic Auth, and client certificates).
* Metering of throughput and exposing monitoring metrics.

### Options for Integrating External Systems

#### Automatically by UCL

If an external system is registered for Kyma runtime in SAP BTP's UCL, the Application Connector module automatically configures it and can send requests to Kyma workloads. The Application Connector module includes [`Runtime Agent`](./technical-reference/runtime-agent/README.md) and acts as a client of the UCL backend. It automatically retrieves the configuration of each external system and integrates it into Kyma.

#### Manually

It is always possible to integrate any external system into Kyma by applying the configuration by hand. The steps for configuring and integrating a new external system in your Kyma runtime are described in the [Integrate an External System with Kyma](tutorials/01-00-integrate-external-system.md).
