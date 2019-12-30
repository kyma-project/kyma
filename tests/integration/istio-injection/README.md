# API Controller Integration Tests

## Overview

This folder contains the integration tests for the `api-controller` component.
The tests are written in Go. Run them as standard Go tests.

## Usage

To test your changes and build the image, run the `make build build-image` command.

## Configuring Kyma

After building and pushing the Docker image, configure it to be used in Kyma by setting proper values for `dir` and `version` in the `global.api_controller_integration_tests` property in the `resources/core/values.yaml` file.
