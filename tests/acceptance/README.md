# Acceptance Tests

## Overview

This project contains the acceptance tests that you can run as part of Kyma testing process.
The tests are written in Go. Each component or group of scenarios has a separate folder, like `servicecatalog`.

## Usage

To test your changes and build an image, use the `make build build-image` command.

### Add new test

To add a new test:

1. Add a new package.
2. Modify the Dockerfile and build.sh script to compile the test binary to pkg.test.
3. Add execution of the test to the `entrypoint.sh` script.
4. Add deletion of the binary to Makefile's cleanup step.

### Configure Kyma

After building and pushing the Docker image, set the proper tag in the `resources/core/values.yaml` file, in the`acceptanceTest.imageTag` property.
