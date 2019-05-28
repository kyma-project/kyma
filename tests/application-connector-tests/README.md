# Application Connector Tests

## Overview

This project contains the acceptance tests that you can run as part of the Kyma Application Connector testing process.

## Usage

Environment parameters used by the tests:

| Name | Required | Default | Description | Possible values |
|------|----------|---------|-------------|-----------------|
| **NAMESPACE** | Yes | - | The Namespace in which the test Application will operate | `kyma-integration` |
| **CENTRAL** | No | false | Determines if the Connector Services acts as a Central | true | 
| **SKIP_SSL_VERIFY** | No | false | Determines if the TLS should be skipped | true | 


### Run locally

If the test is not able to find `InClusterConfig`, it will try to use the local `kubeconfig` file.

To run tests locally export required environment variables:
```
export NAMESPACE=kyma-integration
export CENTRAL=false
export SKIP_SSL_VERIFY=true
```
 
And use `go test` to run the tests:
```
go test ./test/... -v
```
