# Compass Runtime Agent

## Overview

This is the repository for the Compass Runtime Agent.

The main responsibilities of the Compass Runtime Agent are:
- Establishing a trusted connection between the Runtime and Compass
- Renewing the trusted connection between the Runtime and Compass
- Configuring the Runtime


## Prerequisites

The Compass Runtime Agent requires Go 1.8 or higher.

## Usage

To start the Compass Runtime Agent, run this command:

```
./compass-runtime-agent
```

## Parameters and environment variables

The Compass Runtime Agent uses these environment variables:
- **APP_CONFIG_FILE** - specifies the path to the configuration file.
- **APP_CONTROLLER_SYNC_PERIOD** - specifies the time period between resyncing existing resources.
- **APP_MINIMAL_COMPASS_SYNC_TIME** - specifies the minimal time between synchronizing the configuration.
- **APP_CERT_VALIDITY_RENEWAL_THRESHOLD** - specifies when the certificate should be renewed based on remaining validity time of current certificate. 
- **APP_CLUSTER_CERTIFICATES_SECRET** - specifies the Namespace and Name of the secret in which client certificate and key should be stored.
- **APP_CA_CERTIFICATES_SECRET** - specifies the Namespace and Name of the secret in which CA certificate should be stored.
- **APP_INSECURE_CONNECTOR_COMMUNICATION** - specifies whether to communicate with Connector Service with disabled TLS verification.
- **APP_INTEGRATION_NAMESPACE** - specifies the Namespace in which the resources are created.
- **APP_GATEWAY_PORT** - specifies the Application Gateway port.
- **APP_INSECURE_CONFIGURATION_FETCH** - specifies whether to fetch the configuration with disabled TLS verification.
- **APP_UPLOAD_SERVICE_URL** - specifies the URL of the upload service.
- **APP_QUERY_LOGGING** - specifies whether GraphQL queries should be logged.


## Generating Custom Resource client

Code generation is not yet supported with Go Modules, therefor the code generator needs to be run inside Docker container.
To generate Custom Resource client and deep copy functions, run this command from component directory:
```
./hack/code-gen-in-docker.sh
```
