# Central Application Gateway

## Overview

This is the repository for the Central Application Gateway.

## Prerequisites

The Central Application Gateway requires Go 1.8 or higher.

## Installation

To install the Central Application Gateway, follow these steps:

1. Clone the repository to your local machine:
   ```bash
   git clone git@github.com:kyma-project/kyma.git
   ```
2. Navigate to the directory with the Central Application Gateway:
   ```bash
   cd kyma/components/central-application-gateway
   ```
3. Build the component:
   ```bash
   CGO_ENABLED=0 go build ./cmd/applicationgateway
   ```

## Usage

This section explains how to use the Central Application Gateway.

### Start the Central Application Gateway

To start the Central Application Gateway, run this command:

```
./applicationgateway
```

The Central Application Gateway has the following parameters:
- **disableLegacyConnectivity** is the flag for disabling the default legacy mode and enabling the Compass mode. The default value is `false`.
- **proxyPort** is the port that acts as a proxy for the calls from services and Functions to an external solution in the default standalone (legacy) mode. The default port is `8080`.
- **proxyPortCompass** is the port that acts as a proxy for the calls from services and Functions to an external solution in the Compass mode. The default port is `8082`.
- **externalAPIPort** is the port that exposes the API which allows checking the component status. The default port is `8081`.
- **namespace** is the Namespace in which the Central Application Gateway is deployed. The default Namespace is `kyma-system`.
- **requestTimeout** is the timeout for requests sent through the Central Application Gateway, expressed in seconds. The default value is `1`.
- **skipVerify** is the flag for skipping the verification of certificates for the proxy targets. The default value is `false`.
- **requestLogging** is the flag for logging incoming requests. The default value is `false`.
- **proxyTimeout** is the timeout for requests sent through the proxy, expressed in seconds. The default value is `10`.
- **proxyCacheTTL** is the time to live of the remote API information stored in the proxy cache, expressed in seconds. The default value is `120`.


## API
The Central Application Gateway exposes:
- an external API implementing a health endpoint for liveness and readiness probes
- an internal API implementing a proxy handler accessible via a service of type ClusterIP

### Standalone (legacy) mode
If  **disableLegacyConnectivity** is `false`, the proxy API exposes the following endpoint:
```bash
{APPLICATION_NAME}/{SERVICE_NAME}/{TARGET_API_PATH}
``` 

For instance, if the user registered the `cc-occ-commerce-webservices` service in the `ec` application using Application Registry, they can send a request to the following URL: 
```bash
http://central-application-gateway:8080/ec/cc-occ-commerce-webservices/basesites
```

As a result, the Central Application Gateway:
1. Looks for the `cc-occ-commerce-webservices` service in the `ec` Application CRD and extracts the target URL path along with the authentication configuration
2. Modifies the request to include the authentication data
3. Sends the request to the following path:
   ```bash
   {TARGET_URL_EXTRACTED_FROM_APPLICATION_CRD}/basesites

### Compass mode
If **disableLegacyConnectivity** is `true`, the proxy API exposes the following endpoint:
```bash
{APPLICATION_NAME}/{API_BUNDLE_NAME}/{API_DEFINITION_NAME}/{TARGET_API_PATH}
``` 

For instance, if the user registered the `cc-occ` API bundle with the `commerce-webservices` API definition in the `ec` application, they can send a request to the following URL:
```bash
http://central-application-gateway:8082/ec/cc-occ/commerce-webservices/basesites
``` 

As a result, the Central Application Gateway:
1. Looks for the `cc-occ` service and the `commerce-webservices` entry in the `ec` Application CRD and extracts the target URL path along with the authentication configuration
2. Modifies the request to include the authentication data
3. Sends the request to the following path: 
   ```bash
   {TARGET_URL_EXTRACTED_FROM_APPLICATION_CRD}/basesites

## Development

This section explains the development process.

### Generate mocks

Prerequisites:

 - [Mockery](https://github.com/vektra/mockery) 2.0 or higher

To generate mocks, run:

```sh
go generate ./...
```

When adding a new interface to be mocked or when a mock of an existing interface is not being generated, add the following line directly above the interface declaration:

```
//go:generate mockery --name {INTERFACE_NAME}
```

### Tests

This section outlines the testing details.

#### Unit tests

To run the unit tests, run this command:

```
go test./...
```

### Contribution

To learn how you can contribute to this project, see the [Contributing](/CONTRIBUTING.md) document.
