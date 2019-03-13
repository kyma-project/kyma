# Application Gateway Tests

## Overview

This project contains the acceptance tests for the Kyma Application Gateway.

## Prerequisites

The project requires Go 1.8 or higher.

## Usage

Environment parameters used by the tests:

| Name | Required | Default | Description | Possible values |
|------|----------|---------|-------------|-----------------|
| **APPLICATION** | Yes | - | The name of the Application to test | `my-application` | 
| **NAMESPACE** | Yes | - | The namespace on which the test application will operate | `kyma-integration` |
| **MOCK_SERVICE_PORT** | Yes | - |  Port number used by the mock service created by the test | `8080` |
| **TEST_EXECUTOR_IMAGE** | No | Same version as the Helm test image  | Image name of the test executor created by the Helm test  | `user/my-image:1.0.0` |


## Testing process

As the Helm test pod can not be injected with Istio Proxy sidecar, which is required to communicate with the Application Gateway, the test creates a pod which carries out the test (`executor`).



