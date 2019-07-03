# Dex Integration Tests

## Overview

This project contains the integration tests for `dex` component

## Usage

To test your changes and build an image, use the `make build build-image` command.

### Add new test

To add a new test:

1. Add a new package.
2. Modify the Dockerfile and build.sh script to compile the test binary to pkg.test.
3. Add execution of the test to the `entrypoint.sh` script.
4. Add deletion of the binary to Makefile's cleanup step.

### Configure Kyma

After building and pushing the Docker image, set the proper directory and tag in the `resources/dex/values.yaml` file, in the `dex_integration_tests` property.
