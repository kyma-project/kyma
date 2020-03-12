# IAM Kubeconfig Service

## Overview

This project is a generator of configurations used in Kyma.

## Prerequisites

The following tools are required to set up the project:

- [Go](https://golang.org)
- [Docker](https://www.docker.com/)

## Installation

For installation use the dedicated [Helm chart](../../resources/iam-kubeconfig-service).

## Usage

### Configuration

Use the following arguments to configure the application:

| Name | Required | Default | Description |
| -----|---------|--------|------------ |
| port | No | `8000` | Application port. |
| kube-config-cluster-name | Yes | None |  Name of the Kubernetes cluster. |
| kube-config-url | Yes | None | URL of the Kubernetes Apiserver. |
| kube-config-ca-file | Yes | None | Path of the file with Certificate Authority of the Kubernetes cluster. |
| kube-config-ns | No | None | Default namespace of the Kubernetes context. |
| oidc-issuer-url | Yes | None | The URL of the OpenID issuer. Used to verify the OIDC JSON Web Token (JWT). |
| oidc-client-id | Yes | None | The client ID for the OpenID Connect client. |
| oidc-username-claim | No | `email` | Identifier of the user in JWT claim. |
| oidc-groups-claim | No | `groups` | Identifier of groups in JWT claim. |
| oidc-username-prefix | No | None | If provided, all users will be prefixed with this value to prevent conflicts with other authentication strategies. |
| oidc-groups-prefix | No | None | If provided, all groups will be prefixed with this value to prevent conflicts with other authentication strategies. |

### Run a local version

In order to run a local version, a running minikube is required.

To run the application without building the binary, execute the following commands:

```bash
go run cmd/generator/main.go \
  -kube-config-cluster-name=minikube \
  -kube-config-url=:8443 \
  -kube-config-ca-file=~/.minikube/ca.crt \
  -oidc-issuer-url="https://dex.kyma.local" \
  -oidc-client-id="kyma-client"
```

## Development

### Testing

Run tests:

```bash
go test -v ./...
```

Run tests with coverage:

```bash
go test -coverprofile=coverage_report.out -v ./...
```
