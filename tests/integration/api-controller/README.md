# API Controller Integration Tests

## Overview

This folder contains the integration tests for the `api-controller` component.
The tests are written in Go. Run them as standard Go tests.

## Usage

To test your changes and build the image, run the `make build build-image` command.

## Configuring Kyma

After building and pushing the Docker image, set the proper directory and tag in the `resources/core/charts/api-controller/values.yaml` file, in the `tests.image.version` property.
