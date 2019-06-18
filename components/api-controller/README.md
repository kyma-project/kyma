# API Controller

## Overview

The Kyma API Controller is a core component that manages Istio authentication policies and VirtualServices, and allows to expose services using the Kyma Console or API resources. It is implemented according to the [Kubernetes Operator](https://coreos.com/blog/introducing-operators.html) principles and operates on `api.gateway.kyma-project.io` CustomResourceDefinition (CRD) resources.

This [Helm chart](/resources/core/charts/api-controller/Chart.yaml) defines the component's installation.

## Prerequisites

You need these tools to work with the API Controller:

- [Go distribution](https://golang.org)
- [Docker](https://www.docker.com/)


## Details

This section describes how to run the controller locally, how to build the Docker image for the production environment, how to use the environment variables, and how to test the Kyma API Controller.

### Run the component locally

Run Minikube with Istio to use the API Controller locally. Run this command to run the application without building the binary:

```bash
$ go run cmd/controller/main.go
```

### Use environment variables

Use these environment variables to configure the application:

| Name | Required | Default | Description | Possible values |
|------|----------|---------|-------------|-----------------|
| **API_CONTROLLER_LOG_LEVEL** | No | `info` | Show detailed logs in the application. | `info`, `debug`
| **DEFAULT_ISSUER** | Yes | - | Used to set default issuer in the Policy. | any string |
| **DEFAULT_JWKS_URI** | Yes | - | Used to set default jwksUri in the Policy. | any string |
| **GATEWAY_FQDN** | Yes | - | Used to set gateway in the Virtual Service specification. | any string |
| **DOMAIN_NAME** | Yes | - | Used to set a hostname in the Virtual Service specification if a short version of the hostname is provided. | any string |
| **CORS_ALLOW_ORIGIN** | No | `""` | Used to set the `corsPolicy.allowOrigin` field in the specification of the Virtual Service. | `"*"` or comma-separated list of origins |
| **CORS_ALLOW_METHODS** | No | `""` | Used to set the `corsPolicy.allowMethods` field in the specification of the Virtual Service. | comma-separated list of methods |
| **CORS_ALLOW_HEADERS** | No | `""` | Used to set the `corsPolicy.allowHeaders` field in the specification of the Virtual Service. | `"*"` or comma-separated list of headers |
| **BLACKLISTED_SERVICES** | No | `"kubernetes.default,istio-citadel.istio-system,istio-galley.istio-system,istio-ingressgateway.istio-system,istio-pilot.istio-system,istio-policy.istio-system,istio-sidecar-injector.istio-system,istio-telemetry.istio-system,apiserver-proxy.kyma-system, apiserver-proxy-ssl.kyma-system"` | Used to defined a list of services for which the API Controller doesn't create Virtual Services or Policies. | Comma-separated list of services and their source Namespace in the `{SERVICE_NAME}.{NAMESPACE}` format.


### Test

Run all tests:

```bash
$ go test -v ./...
```

Run all tests with coverage:

```bash
$ go test -coverprofile=coverage_report.out -v ./...
```

Run unit tests only:

```bash
$ go test -short -v ./...
```

Run unit tests with coverage:

```bash
go test -short -coverprofile=coverage_report.out -v ./...
```

Run integration tests only:

```bash
$ go test -run Integration -v ./...
```
