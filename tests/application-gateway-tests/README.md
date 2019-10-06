# Application Gateway Tests

## Overview

This project contains the acceptance tests for the Kyma Application Gateway.

## Prerequisites

The project requires Go 1.8 or higher.

## Usage

Environment parameters used by the tests:

| Name | Required | Default | Description | Possible values |
|------|----------|---------|-------------|-----------------|
| **APPLICATION** | Yes | None | The name of the Application to test | `my-application` |
| **NAMESPACE** | Yes | None | The Namespace in which the test Application will operate | `kyma-integration` |
| **MOCK_SERVICE_PORT** | Yes | None |  Port number used by the mock service created by the test | `8080` |
| **TEST_EXECUTOR_IMAGE** | No | Same version as the Helm test image  | Image name of the test executor created by the Helm test  | `user/my-image:1.0.0` |


## Development

The project consists of two applications: Helm Test and Test Executor.
The Test Executor Pod is managed by Helm Test.
