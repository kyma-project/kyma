# Application Operator Tests

## Overview

This project contains the acceptance tests that you can run as part of the Kyma Application Connector testing process.
The tests are written in Go. Run them as standard Go tests.

## Usage

This section provides information on building and versioning the Docker image, as well as configuring Kyma.

### Configuring Kyma

After building and pushing the Docker image, set the proper tag in the `resources/core/charts/application-connector/charts/application-operator/values.yaml` file, in the`tests.image.tag` property.
