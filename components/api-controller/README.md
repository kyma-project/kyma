```
    _    ____ ___    ____            _             _ _
   / \  |  _ \_ _|  / ___|___  _ __ | |_ _ __ ___ | | | ___ _ __
  / _ \ | |_) | |  | |   / _ \| '_ \| __| '__/ _ \| | |/ _ \ '__|
 / ___ \|  __/| |  | |__| (_) | | | | |_| | | (_) | | |  __/ |
/_/   \_\_|  |___|  \____\___/|_| |_|\__|_|  \___/|_|_|\___|_|
```


## Overview

The Kyma API Controller is a core component that manages Istio authentication policies and VirtualServices, and allows to expose services using the Kyma Console or API resources. It is implemented according to the [Kubernetes Operator](https://coreos.com/blog/introducing-operators.html) principles and operates on `api.gateway.kyma.cx` Custom Resource Definition (CRD) resources.

This [Helm chart](/resources/core/charts/api-controller/Chart.yaml) defines the component's installation.

## Prerequisites

You need these tools to work with the API Controller:

- [Go distribution](https://golang.org)
- [Docker](https://www.docker.com/)


## Details

This section describes how to run the controller locally, how to build the Docker image for the production environment, how to use the environment variables, and how to test the Kyma API Controller.

### Run the component locally

Run Minikube with Istio to use the API Controller locally. Run this command to run the application without building the binary:

```bash
$ go run cmd/controller/main.go
```

### Use environment variables

Use these environment variables to configure the application:

| Name | Required | Default | Description |
|-----|---------|--------|------------|
| API_CONTROLLER_LOG_LEVEL | No | `info` | Show detailed logs in the application.
| DEFAULT_ISSUER | Yes | - | Used to set default issuer in NetworkPolicy
| DEFAULT_JWKS_URI | Yes | - | Used to set default jwksUri in NetworkPolicy
| GATEWAY_FQDN | Yes | - | Used to set gateway in VirtualServices spec


### Test

Run all tests:

```bash
$ go test -v ./...
```

Run all tests with coverage:

```bash
$ go test -coverprofile=coverage_report.out -v ./...
```

Run unit tests only:

```bash
$ go test -short -v ./...
```

Run unit tests with coverage:

```bash
go test -short -coverprofile=coverage_report.out -v ./...
```

Run integration tests only:

```bash
$ go test -run Integration -v ./...
```
