# Connector Service Acceptance Tests

## Overview

This project contains the acceptance tests for Kyma Connector Service.

## Prerequisites

The Connector Service Acceptance Tests requires Go 1.8 or higher.

## Build

To install the Metadata Service components, follow these steps:

1. `git clone git@github.com:kyma-project/kyma.git`
2. `cd /tests/connector-service-tests`
3. `CGO_ENABLED=0 go build ./test/apitests/connector_test.go`

## Usage

### Environment parameters

* **INTERNAL_API_URL**  - The URL of Connector Service internal API 
* **EXTERNAL_API_URL** - The URL of Connector Service external API 
* **GATEWAY_URL** - The URL of Remote Environment Gateway API
* **SKIP_SSL_VERIFY** - A flag for skipping SSL certificate validation

### Running

1. Provide all required environment parameters
2. Execute tests
    
    ```bash
    go test ./...
    ```


