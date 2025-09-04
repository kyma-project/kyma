# API Gateway Module

## What Is API Gateway?

API Gateway is a Kyma module with which you can expose and secure APIs.

To use the API Gateway module, you must also add the Istio module. Moreover, to expose a workload using the APIRule custom resource, the workload must be part of the Istio service mesh. 

By default, both the API Gateway and Istio modules are automatically added when you create a Kyma runtime instance. 

## Features

The API Gateway module offers the following features:

- Ory Oathkeeper installation: The module simplifies and manages the installation of Ory Oathkeeper.
- API Exposure: The module combines ORY Oathkeeper and Istio capabilities to offer the APIRule CustomResourceDefinition. By creating APIRule custom resources, you can easily and securely expose your workloads.
- Kyma Gateway installation: The module installs the default simple TLS Kyma Gateway.

## Architecture

![Kyma API Gateway Operator Overview](../assets/operator-overview.svg)

### API Gateway Operator

Within the API Gateway module, API Gateway Operator manages the application of API Gateway's configuration and handles resource reconciliation. It contains the following controllers: APIGateway Controller, APIRule Controller, and RateLimit Controller.


### APIGateway Controller

APIGateway Controller manages the installation of [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and handles the configuration of Kyma Gateway and the resources defined in the APIGateway custom resource (CR). The controller is responsible for:
- Installing, upgrading, and uninstalling Ory Oathkeeper
- Configuring Kyma Gateway
- Managing Certificate and DNSEntry resources

### APIRule Controller

APIRule Controller uses [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper) and [Istio](https://istio.io/) resources to expose and secure APIs.

### RateLimit Controller

RateLimit Controller manages the configuration of local rate limiting on the Istio service mesh layer. By creating a RateLimit custom resource (CR), you can limit the number of requests targeting an exposed application in a unit of time, based on specific paths and headers.

## API/Custom Resource Definitions

The `apigateways.operator.kyma-project.io` CustomResourceDefinition (CRD) describes the APIGateway CR that APIGateway Controller uses to manage the module and its resources. See [APIGateway Custom Resource](./custom-resources/apigateway/04-00-apigateway-custom-resource.md).

The `apirules.operator.kyma-project.io` CRD describes the APIRule CR that APIRule Controller uses to expose and secure APIs. See [APIRule Custom Resource](./custom-resources/apirule/README.md).

The `ratelimits.gateway.kyma-project.io` CRD describes the kind and the format of data that RateLimit Controller uses to configure request rate limits for applications. See [RateLimit Custom Resource](./custom-resources/ratelimit/04-00-ratelimit.md).

## Resource Consumption

To learn more about the resources used by the Istio module, see [Kyma Modules' Sizing](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/kyma-modules-sizing?locale=en-US&state=DRAFT&version=Internal&comment_id=22217515&show_comments=true#api-gateway).
