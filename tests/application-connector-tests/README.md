# Application Connector Tests

## Overview

This project contains the acceptance tests that you can run as part of the Kyma Application Connector testing process.
The tests are written in Go. Run them as standard Go tests.

## Usage

Environment parameters used by the tests:

| Name | Required | Default | Description | Possible values |
|------|----------|---------|-------------|-----------------|
| **CENTRAL** | No | false | Determines if the Connector Services acts as a Central | true | 
| **NAMESPACE** | Yes | - | The Namespace in which the test Application will operate | `kyma-integration` |
