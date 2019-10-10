# Connector Service Acceptance Tests

## Overview

This project contains the acceptance tests for the Kyma Connector Service.

## Prerequisites

The project requires Go 1.8 or higher.

## Installation

To install the Connector Service components, follow these steps:

1. `git clone git@github.com:kyma-project/kyma.git`
2. `cd /tests/connector-service-tests`
3. `CGO_ENABLED=0 go build ./test/apitests/connector_test.go`

## Usage

Set the environment parameters:

| Name | Required | Default | Description | Possible values |
|------|----------|---------|-------------|-----------------|
| **INTERNAL_API_URL** | Yes | None | The URL of Connector Service internal API | `http://localhost:8080` | 
| **EXTERNAL_API_URL** | Yes | None | The URL of Connector Service external API | `https://connector-service.kyma.local` |
| **GATEWAY_URL** | Yes | None |  The URL of Application Gateway API | `https://gateway.kyma.local:30218` |
| **SKIP_SSL_VERIFY** | No | `false` | A flag for skipping SSL certificate validation | `true` |

Run this command to execute tests:

    ```bash
    go test ./...
    ```
