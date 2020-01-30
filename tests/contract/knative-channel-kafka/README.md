# Knative Kafka channel acceptance tests

## Overview

This project contains the acceptance tests for the [Knative Kafka channel](https://github.com/kyma-incubator/knative-kafka). It can run as part of the Kyma testing process. The tests are written in Go, so you can run them as standard Go tests.

## Usage
This section provides information on building and versioning the Docker image.

### Building Docker image

To build a Docker image, run:
```
make resolve-local build-image
```

To push the built image, run:
```
DOCKER_PUSH_REPOSITORY={address of repository}
make push-image
```

### Configuring Kyma

After building and pushing the Docker image, set the proper tag in the `resources/knative-eventing-channel-kafka-tests/values.yaml` file, in the **test.version** property.
