# Apiserver-Proxy Integration Tests

## Overview

This folder contains the integration tests for the `apiserver-proxy` component.

## Details
- Contains the Dockerfile for the image used in Kyma API Server Proxy tests.
- Contains the `fetch-token` application used for fetching authentication tokens from Dex.
- Contains the `test.sh` script that runs tests for the chart.

## Usage

To test your changes and build the image, run the `make build build-image` command.

## Configure Kyma

After building and pushing the Docker image, set the proper directory and tag in the `resources/apiserver-proxy/values.yaml` file, in the `apiserver_proxy_integration_tests` property.

