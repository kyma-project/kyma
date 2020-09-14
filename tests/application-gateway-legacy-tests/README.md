# Application Gateway Tests

## Overview

This project contains the acceptance tests for the Kyma Application Gateway in the legacy mode.

## Prerequisites

The project requires Go 1.8 or higher.

## Usage

Environment parameters used by the tests:

| Name | Required | Description | Default | Example value |
|------|----------|---------|-------------|-----------------|
| **APPLICATION** | Yes | Name of the Application to test | None | `my-application` |
| **NAMESPACE** | Yes | Namespace in which the test Application will operate | None | `kyma-integration` |
| **MOCK_SERVICE_PORT** | Yes | Number of the port used by the mock service created by the test | None | `8080` |
| **TEST_EXECUTOR_IMAGE** | No | Name of the test executor image created by the Helm test | Version matching the Helm test image | `user/my-image:1.0.0` |

## Development

The project consists of two applications: Helm Test and Test Executor.
The Test Executor Pod is managed by Helm Test.
