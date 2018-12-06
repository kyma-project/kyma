# Acceptance Tests

## Overview

This project contains the acceptance tests that you can run as part of the Kyma testing process.
The tests are written in Go. Run them as standard Go tests.

## Usage
This section provides information on building and versioning of the Docker image, as well as configuring the Kyma.

### Configuring Kyma

After building and pushing the Docker image, set the proper tag in the `resources/core/charts/api-controller/values.yaml` file, in the`tests.image.version` property.
