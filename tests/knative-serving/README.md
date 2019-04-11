# Acceptance Tests

## Overview

This project contains the acceptance tests that you can run as part of the Kyma testing process.
The tests are written in Go. Run them as standard Go tests.

## Usage
This section provides information on building and versioning the Docker image, as well as configuring Kyma.

### Building Docker image

To build a Docker image, run:
```
make build
```

To push the built image, run:
```
DOCKER_PUSH_REPOSITORY={address of repository} make push-image
```

### Configuring Kyma

After building and pushing the Docker image, set the proper tag in the `resources/kantive/values.yaml` file, in the **global.test_knative_serving_acceptance.version** property.
