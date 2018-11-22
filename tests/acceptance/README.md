# Acceptance Tests

## Overview

This project contains the acceptance tests that you can run as part of Kyma testing process.
The tests are written in Go. Each component or group of scenarios has a separate folder, like `dex` or `servicecatalog`.

## Usage

This section provides information on building and versioning of the Docker image, as well as configuring Kyma.
Use `make build build-image` to test your changes and build image.

### Adding new tests

1. Add a new package
2. Modify the Dockerfile to compile the test binary to pkg.test
3. Add execution of the test to the `entrypoint.sh` script
4. Add deletion of the binary to Makefile's cleanup step.

### Configuring Kyma

After building and pushing the Docker image, set the proper tag in the `resources/core/values.yaml` file, in the`acceptanceTest.imageTag` property.
