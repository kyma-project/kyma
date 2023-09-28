# Runtime Agent

## Overview

This is the repository for Runtime Agent.

The main responsibilities of Runtime Agent are:
- Establishing a trusted connection between the Runtime and Compass
- Renewing the trusted connection between the Runtime and Compass
- Configuring the Runtime

## Prerequisites

Runtime Agent requires Go 1.8 or higher.

## Unit testing

To execute unit tests for controller functionality, Runtime Agent requires an additional `etcd` binary to be installed on your local machine at this location: `/usr/local/kubebuilder/bin`.

See [this link](https://book.kubebuilder.io/reference/envtest.html) for installation instructions.

## Usage

To start Runtime Agent, run this command:

```bash
./compass-runtime-agent
```

## Parameters and environment variables

Runtime Agent uses these environment variables:
- **APP_AGENT_CONFIGURATION_SECRET** specifies the Namespace and the Name of the Secret containing the Runtime Agent Configuration.
- **APP_CONTROLLER_SYNC_PERIOD** specifies the time period between resynchronizing existing resources.
- **APP_MINIMAL_COMPASS_SYNC_TIME** specifies the minimal time between synchronizing the configuration.
- **APP_CERT_VALIDITY_RENEWAL_THRESHOLD** specifies when the certificate must be renewed based on the remaining validity time of the current certificate.
- **APP_CLUSTER_CERTIFICATES_SECRET** specifies the Namespace and the Name of the Secret in which to store the client certificate and the key.
- **APP_CA_CERTIFICATES_SECRET** specifies the Namespace and the Name of the Secret in which to store the CA certificate.
- **APP_SKIP_COMPASS_TLS_VERIFY** specifies whether to communicate with Connector and Director with disabled TLS verification.
- **APP_SKIP_APPS_TLS_VERIFY** specifies whether to set up Applications synchronized from Compass to communicate with external systems with disabled TLS verification.
- **APP_GATEWAY_PORT** specifies the Application Gateway port.
- **APP_QUERY_LOGGING** specifies whether to log GraphQL queries.
- **APP_RUNTIME_EVENTS_URL** specifies the Events URL of the cluster that Runtime Agent runs on.
- **APP_RUNTIME_CONSOLE_URL** specifies the Console URL of the cluster that Runtime Agent runs on. <!-- TODO: To be removed after it's been removed from the code. See https://github.com/kyma-project/kyma/pull/13984, the comment: discussion_r861476457. -->
- **APP_HEALTH_PORT** specifies the health check port.
- **APP_CA_CERT_SECRET_TO_MIGRATE** specifies the Namespace and the name of the Secret which stores the CA certificate to be renamed. Requires the `{NAMESPACE}/{SECRET_NAME}` format. 
- **APP_CA_CERT_SECRET_KEYS_TO_MIGRATE** specifies the list of keys to be copied when migrating the old Secret specified in **APP_CA_CERT_SECRET_TO_MIGRATE** to the new one specified in **APP_CA_CERTIFICATES_SECRET**. Requires the JSON table format.

## Renaming Secrets

To rename the Secret containing the CA cert, you must specify these environment variables:
- **APP_CA_CERTIFICATES_SECRET**
- **APP_CA_CERT_SECRET_TO_MIGRATE**
- **APP_CA_CERT_SECRET_KEYS_TO_MIGRATE**

## Generating custom resource client

Because Go Modules do not support code generation, you must run the code generator inside a Docker container.
To generate a custom resource client and deep copy Functions, run this command from the component directory:

```bash
./hack/code-gen-in-docker.sh
```
