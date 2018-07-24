# Application Connector Tests

## Overview

This project contains the acceptance tests that you can run as part of the Kyma Application Connector testing process.
The tests are written in Go. Run them as standard Go tests.
Each component or group of scenarios has a separate folder, like `metadata`.

Project contains tests for following components:
 - Metadata 

## Usage

This section provides information on building and versioning of the Docker image, as well as configuring the Kyma.

### Configuring the Kyma

After building and pushing the Docker image, set the proper tag in the `resources/core/charts/application-connector/metadata/values.yaml` file, in the`tests.image.tag` property.
