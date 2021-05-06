# Application Operator Tests

## Overview

This project contains the acceptance tests that you can run as part of the Kyma Application Connector testing process.
The tests are written in Go. Run them as standard Go tests.

## Usage

This section provides information on building and versioning the Docker image, as well as configuring Kyma.

Environment parameters used by the tests:

| Name | Required | Default | Description | Possible values |
|------|----------|---------|-------------|-----------------|
| **NAMESPACE** | Yes | `kyma-integration` | Namespace in which the test Application will operate. |  |
| **HELM_DRIVER** | Yes | `secret` | Backend storage driver used by Helm 3 to store release data. | `configmap`, `secret`, `memory` |
| **INSTALLATION_TIMEOUT_SECONDS** | No | 180 | Timeout for the release installation, provided in seconds. |  |
| **GATEWAY_DEPLOYED_PER_NAMESPACE** | No | `false` | Flag that specifies whether Application Gateway should be deployed once per Namespace basing on ServiceInstance or for every Application. | `true`, `false` |
| **CENTRAL_APPLICATION_CONNECTIVITY_VALIDATOR** | No | `true` | Flag that specifies whether Central Application Connectivity Validator is used i.e. Validator is not deployed for every Application. | `true`, `false` |

### Configuring Kyma

After building and pushing the Docker image, set the proper tag in the `resources/core/charts/application-connector/charts/application-operator/values.yaml` file, in the **tests.image.tag** property.
