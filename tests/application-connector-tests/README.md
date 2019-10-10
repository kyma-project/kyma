# Application Connector Tests

## Overview

This project contains the acceptance tests that you can run as part of the Kyma Application Connector testing process.

## Usage

Environment parameters used by the tests:

| Name | Required | Default | Description | Possible values |
|------|----------|---------|-------------|-----------------|
| **NAMESPACE** | Yes | None | The Namespace in which the test Application runs. | `kyma-integration` |
| **CENTRAL** | No | false | Determines if the Connector Service operates in the central mode.  | true | 
| **SKIP_SSL_VERIFY** | No | false | Determines if the TLS should be skipped. | true |


### Run locally

If the test can't find `InClusterConfig`, it uses the local `kubeconfig` file.

To run tests locally, export the required environment variables:
```
export NAMESPACE=kyma-integration
export CENTRAL=false
export SKIP_SSL_VERIFY=true
```

Use `go test` to run the tests:
```
go test ./test/... -v
```
