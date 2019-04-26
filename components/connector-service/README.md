# Connector Service

## Overview

This is the repository for the Kyma Connector Service.

## Prerequisites

The Connector Service requires Go 1.8 or higher.

## Usage

This section explains how to use the Connector Service.

### Start the Connector Service
To start the Connector Service, run this command:

```
./connectorservice
```

The Connector Service has the following parameters:
- **appName** - This is the name of the application used by Kubernetes deployments and services. The default value is `connector-service`.
- **externalAPIPort** - This port exposes the Connector Service API to an external solution. The default port is `8081`.
- **internalAPIPort** - This port exposes the Connector Service within Kubernetes cluster. The default port is `8080`.
- **namespace** - Namespace where Connector Service is deployed. The default Namespace is `kyma-integration`.
- **tokenLength** - Length of registration tokens. The default value is `64`.
- **appTokenExpirationMinutes** - Time after which tokens for Applications expire and are no longer valid. The default value is `5` minutes.
- **runtimeTokenExpirationMinutes** - Time after which tokens for Runtimes expire and are no longer valid. The default value is `10` minutes.
- **caSecretName** - Name of the secret which contains the root Certificate Authority (CA). The default value is `nginx-auth-ca`.
- **requestLogging** - Flag for logging incoming requests. It is set to `False` by default.
- **connectorServiceHost** - Host under which this service is accessible. It is used for generating the URL. The default host is `cert-service.wormhole.cluster.kyma.cx`.
- **gatewayBaseURL** - Base URL of the gateway service.
- **certificateProtectedHost** - Host secured with the client certificate, used for the certificate renewal. The default host is `gateway.wormhole.cluster.kyma.cx`.
- **appsInfoURL** - URL at which the management information for applications is available. If not provided, it bases on `connectorServiceHost`.
- **runtimesInfoURL** - URL at which the management information for runtimes is available. If not provided, it bases on `connectorServiceHost`.
- **appCertificateValidityTime** - Time until which the certificates that the service issues for Applications are valid. The default value is 90 days.
- **runtimeCertificateValidityTime** - Time until which the certificates that the service issues for Runtimes are valid. The default value is 90 days.
- **central** - Determines whether the Connector Service works in the central mode.
- **revocationConfigMapName** - Name of the ConfigMap containing the revoked certificates list.

Connector Service also uses following environmental variables for CSR - related information config:
- **COUNTRY** (two-letter-long country code)
- **ORGANIZATION**
- **ORGANIZATIONALUNIT**
- **LOCALITY**
- **PROVINCE**

## Testing on local deployment

When you develop the Application Connector components, you can test the changes you introduced on a local Kyma deployment before you push them to a production cluster.
To test the component you modified, run the `run-with-local-tests.sh` script located in the `scripts` directory.

Running the script builds the Docker image of the component, pushes it to the Minikube registry, and updates the component deployment in the Minikube cluster. It then triggers the `run-local-tests.sh` script, which builds the image of the acceptance tests to the Minikube registry, creates a Pod with the tests, and fetches the logs from that Pod.

Alternatively, you can run only the `run-local-tests.sh` script for the given component to build the image of the component's acceptance tests to the Minikube registry, create a Pod with the tests, and fetch the logs from that Pod.
