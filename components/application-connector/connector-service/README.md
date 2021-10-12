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
- **appName** is the name of the application used by Kubernetes deployments and services. The default value is `connector-service`.
- **externalAPIPort** is the port that exposes the Connector Service API to an external solution. The default port is `8081`.
- **internalAPIPort** is the port that exposes the Connector Service within Kubernetes cluster. The default port is `8080`.
- **namespace** is the Namespace in which the Connector Service is deployed. The default Namespace is `kyma-integration`.
- **tokenLength** is the length of registration tokens. The default value is `64`.
- **appTokenExpirationMinutes** is the time after which tokens for Applications expire and are no longer valid. The default value is `5` minutes.
- **runtimeTokenExpirationMinutes** is the time after which tokens for Runtimes expire and are no longer valid. The default value is `10` minutes.
- **caSecretName** is the Namespace and the name of the Secret which stores the certificate and the key used to sign client certificates. Requires the `Namespace/secret name` format. The default value is `kyma-integration/nginx-auth-ca`.
- **rootCACertificateSecretName** is the Namespace and the name of the Secret which stores the root CA (Certificate Authority) Certificate in case the certificates are signed by the intermediate CA. Requires the `Namespace/secret name` format. Empty by default. 
- **requestLogging** is the flag for logging incoming requests. It is set to `False` by default.
- **connectorServiceHost** is the host under which this service is accessible. It is used for generating the URL. The default host is `cert-service.wormhole.cluster.kyma.cx`.
- **gatewayBaseURL** is the base URL of the Gateway Service.
- **certificateProtectedHost** is the host secured with the client certificate, used for the certificate renewal. The default host is `gateway.wormhole.cluster.kyma.cx`.
- **appsInfoURL** is the URL at which the management information for Applications is available. If not provided, it bases on `connectorServiceHost`.
- **runtimesInfoURL** is the URL at which the management information for runtimes is available. If not provided, it bases on `connectorServiceHost`.
- **appCertificateValidityTime** is the time until which the certificates that the service issues for Applications are valid. The default value is `90` days.
- **runtimeCertificateValidityTime** is the time until which the certificates that the service issues for Runtimes are valid. The default value is `90` days.
- **central** is the flag that determines whether the Connector Service works in the central mode.
- **revocationConfigMapName** is the name of the ConfigMap containing the revoked certificates list.
- **lookupEnabled** is the flag that determines if the Connector should make a call to get the gateway endpoint. The default value is `False`.
- **lookupConfigMapPath** is the path in the Pod where ConfigMap for cluster lookup is stored. The default value is `/etc/config/config.json`. Used only when **lookupEnabled** is set to `True`.

Connector Service also uses the following environment variables for CSR-related information config:
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
