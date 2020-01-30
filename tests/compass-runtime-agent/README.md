# Compass Runtime Agent Tests

## Overview

This project contains the acceptance tests that you can run Kyma Compass Runtime Agent testing process.
The tests are written in Go. Run them as standard Go tests.

## Usage

This section provides information on versioning of the Docker image, as well as configuring Kyma to use new Compass Runtime Agent image.

### Configure Kyma

After building and pushing the Docker image, set the proper values in the `resources/compass-runtime-agent/values.yaml` file under the **global.images.runtimeAgentTests** property for your newly created image.


## Development

To run tests against current cluster context, run the following commands from the tests root directory:
```
export DOCKER_PUSH_REPOSITORY={DOCKER_REPOSITORY}
export DOCKER_TAG={DOCKER_TAG}
./scripts/run_tests.sh
```
