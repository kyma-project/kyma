# Apiserver-Proxy Integration Tests

## Overview

This folder contains the integration tests for `apiserver-proxy` component.

## Details
- Contains the `Dockerfile` for the image used in Kyma apiserver-proxy tests.
- Contains the `fetch-token` application used for fetching authentication token from Dex.
- Contains the `test.sh` script that runs tests for the chart.

## Usage

To test your changes and build an image, use the `make build build-image` command.

## Configure Kyma

After building and pushing the Docker image, set the proper directory and tag in the `resources/core/values.yaml` file, in the `apiserver_proxy_integration_tests` property.

