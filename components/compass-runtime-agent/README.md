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
- **APP_CONNECTION_CONFIG_MAP** specifies the Namespace and the Name of the Config Map containing Runtime Agent Configuration. 
- **APP_CONTROLLER_SYNC_PERIOD** specifies the time period between resynchronizing existing resources.
- **APP_MINIMAL_COMPASS_SYNC_TIME** specifies the minimal time between synchronizing the configuration.
- **APP_CERT_VALIDITY_RENEWAL_THRESHOLD** specifies when the certificate must be renewed based on the remaining validity time of the current certificate. 
- **APP_CLUSTER_CERTIFICATES_SECRET** specifies the Namespace and the Name of the Secret in which to store the client certificate and the key.
- **APP_CA_CERTIFICATES_SECRET** specifies the Namespace and the Name of the Secret in which to store the CA certificate.
- **APP_INSECURE_CONNECTOR_COMMUNICATION** specifies whether to communicate with the Connector Service with disabled TLS verification.
- **APP_INTEGRATION_NAMESPACE** specifies the Namespace in which to create the resources.
- **APP_GATEWAY_PORT** specifies the Application Gateway port.
- **APP_INSECURE_CONFIGURATION_FETCH** specifies whether to fetch the configuration with disabled TLS verification.
- **APP_UPLOAD_SERVICE_URL** specifies the URL of the upload service.
- **APP_QUERY_LOGGING** specifies whether to log GraphQL queries.
- **APP_METRICS_LOGGING_TIME_INTERVAL** specifies the time interval between the cluster metrics logging.


## Generating Custom Resource client

Because Go Modules do not support code generation, you must run the code generator inside a Docker container.
To generate a Custom Resource client and deep copy functions, run this command from the component directory:
```
./hack/code-gen-in-docker.sh
```
