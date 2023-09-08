# Central Application Gateway

## Overview

This is the repository for Central Application Gateway.

## Prerequisites

Central Application Gateway requires Go 1.8 or higher.

## Installation

To install Central Application Gateway, follow these steps:

1. Clone the repository to your local machine:
   ```bash
   git clone git@github.com:kyma-project/kyma.git
   ```
2. Navigate to the directory with Central Application Gateway:
   ```bash
   cd kyma/components/central-application-gateway
   ```
3. Build the component:
   ```bash
   CGO_ENABLED=0 go build ./cmd/applicationgateway
   ```

## Usage

This section explains how to use Central Application Gateway.

### Start Central Application Gateway

To start Central Application Gateway, run this command:

```bash
./applicationgateway 
```

Central Application Gateway has the following parameters:

- **apiServerURL** - The address of the Kubernetes API server. Overrides any value in a kubeconfig. Only required if out-of-cluster.
- **applicationSecretsNamespace** - Namespace where Application secrets used by the Application Gateway exist. The default is `kymasystem`
- **externalAPIPort** - Port that exposes the API which allows checking the component status and exposes log configuration. The default is `8081`
- **kubeConfig** - Path to a kubeconfig. Only required if out-of-cluster
- **logLevel** - Log level: `panic` | `fatal` | `error` | `warn` | `info` | `debug`. Can't be lower than `info`. The default is  `zapInfoLevel`
- **proxyCacheTTL** - TTL, in seconds, for proxy cache of Remote API information. The default is `120`
- **proxyPort** - Port that acts as a proxy for the calls from services and Functions to an external solution in the default standalone mode or Compass bundles with a single API definition. The default is `8080`
- **proxyPortCompass** - Port that acts as a proxy for the calls from services and Functions to an external solution in the Compass mode. The default is `8082`
- **proxyTimeout** - Timeout for requests sent through the proxy, expressed in seconds. The default is `10`
- **requestTimeout** - Timeout for requests sent through Central Application Gateway, expressed in seconds. The defaultis `1`

## API

Central Application Gateway exposes:
- an external API implementing a health endpoint for liveness and readiness probes
- 2 internal APIs implementing a proxy handler accessible via a service of type `ClusterIP`

Application Gateway also supports redirects for the request flows in which the URL host remains unchanged. For more details, see [Response rewriting](../../docs/05-technical-reference/ac-01-application-gateway-details.md#response-rewriting).

### Standalone mode

The proxy API exposes the following endpoint:
```bash
{APPLICATION_NAME}/{SERVICE_NAME}/{TARGET_API_PATH}
``` 

For instance, if there's a `cc-occ-commerce-webservices` service in the `ec` Application CR, the user can send a request to the following URL: 
```bash
http://central-application-gateway.kyma-system:8080/ec/cc-occ-commerce-webservices/basesites
```

As a result, Central Application Gateway:
1. Looks for the `cc-occ-commerce-webservices` service in the `ec` Application CR and extracts the target URL path along with the authentication configuration.
2. Modifies the request to include the authentication data.
3. Sends the request to the following path:
   ```bash
   {TARGET_URL_EXTRACTED_FROM_APPLICATION_CR}/basesites
   ```

#### Standalone mode for Compass - simplified API

The standalone mode can also be used for Compass bundles with a single API definition.
In this case, `{API_DEFINITION_NAME}` is removed from the URL and the pattern looks as follows:
```bash
{APPLICATION_NAME}/{API_BUNDLE_NAME}/{TARGET_API_PATH}
```
> **NOTE:** Invocation of service bundles configured with multiple API definitions results in a `400 Bad Request` failure.

### Compass mode

The proxy API exposes the following endpoint:
```bash
{APPLICATION_NAME}/{API_BUNDLE_NAME}/{API_DEFINITION_NAME}/{TARGET_API_PATH}
```

For instance, if the user registered the `cc-occ` API bundle with the `commerce-webservices` API definition in the `ec` application, they can send a request to the following URL:
```bash
http://central-application-gateway.kyma-system:8082/ec/cc-occ/commerce-webservices/basesites
```

As a result, Central Application Gateway:
1. Looks for the `cc-occ` service and the `commerce-webservices` entry in the `ec` Application CR and extracts the target URL path along with the authentication configuration.
2. Modifies the request to include the authentication data.
3. Sends the request to the following path: 
   ```bash
   {TARGET_URL_EXTRACTED_FROM_APPLICATION_CRD}/basesites
   ```

#### Handling ambiguous API definition names

A combination of `{API_BUNDLE_NAME}` and `{API_DEFINITION_NAME}` which are extracted from an Application CR must be unique for a given application.
Invocation of endpoints with duplicate names results in a `400 Bad Request` failure. In such a case, you must change one of the names to avoid ambiguity.

### Status codes for errors returned by Application Gateway

- `404 Not Found` - returned when the Application specified in the path doesn't exist.
- `400 Bad Request` - returned when an Application, service, or entry for the [Compass mode](https://kyma-project.io/#/01-overview/application-connectivity/README) is not specified in the path.
- `504 Gateway Timeout` - returned when a call to the target API times out.

## Development

This section explains the development process.

### Generate mocks

Prerequisites:

 - [Mockery](https://github.com/vektra/mockery) 2.0 or higher

To generate mocks, run:

```bash
go generate ./...
```

When adding a new interface to be mocked or when a mock of an existing interface is not being generated, add the following line directly above the interface declaration:

```bash
//go:generate mockery --name {INTERFACE_NAME}
```

### Tests

This section outlines the testing details.

#### Unit tests

To run the unit tests, run this command:

```bash
go test./...
```

### Contribution

To learn how you can contribute to this project, see the [Contributing](/CONTRIBUTING.md) document.
