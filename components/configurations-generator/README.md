```
  ____             __ _                       _   _
 / ___|___  _ __  / _(_) __ _ _   _ _ __ __ _| |_(_) ___  _ __  ___
| |   / _ \| '_ \| |_| |/ _` | | | | '__/ _` | __| |/ _ \| '_ \/ __|
| |__| (_) | | | |  _| | (_| | |_| | | | (_| | |_| | (_) | | | \__ \
 \____\___/|_| |_|_| |_|\__, |\__,_|_|  \__,_|\__|_|\___/|_| |_|___/
                        |___/
  ____                           _
 / ___| ___ _ __   ___ _ __ __ _| |_ ___  _ __
| |  _ / _ \ '_ \ / _ \ '__/ _` | __/ _ \| '__|
| |_| |  __/ | | |  __/ | | (_| | || (_) | |
 \____|\___|_| |_|\___|_|  \__,_|\__\___/|_|

```

## Overview

This project is a generator of configurations used in Kyma.

## Prerequisites

The following tools are required to set up the project:
- [Go distribution](https://golang.org)
- [Docker](https://www.docker.com/)

## Installation

For installation use dedicated [Helm chart](https://github.com/kyma-project/kyma/tree/master/resources/core/charts/configurations-generator)

## Usage

### Configuration

Use the following arguments to configure the application:

| Name | Required | Default | Description |
| -----|---------|--------|------------ |
| port | No | 8000 | Application port. |
| kube-config-custer-name | Yes | |  Name of the Kubernetes cluster. |
| kube-config-url | Yes | | URL of the Kubernetes Apiserver. |
| kube-config-ca | Yes, if kube-config-ca-file not specified | | Certificate Authority of the Kubernetes cluster. |
| kube-config-ca-file | Yes, if kube-config-ca not specified | | File with Certificate Authority of the Kubernetes cluster. |
| kube-config-ns | No | | Default namespace of the Kubernetes context. |

### Run a local version

In order to run a local version, a running minikube is required.

To run the application without building the binary, execute the following commands:

```bash
go run cmd/generator/main.go \
  -kube-config-custer-name=minikue \
  -kube-config-url=:8443 \
  -kube-config-ca-file=~/.minikube/ca.crt
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
