# Asset Metadata Service

## Overview

The Asset Metadata Service is an HTTP server that exposes the functionality for extracting metadata from files. It contains a simple HTTP endpoint which accepts `multipart/form-data` forms. The service extracts front matter YAML metadata from text files of all extensions. 

## Prerequisites

Use the following tools to set up the project:

- [Go distribution](https://golang.org)
- [Docker](https://www.docker.com/)

## Usage

### Run a local version

To run the application against the local Kyma installation on Minikube without building the binary, run this command:

```bash
APP_VERBOSE=true go run main.go
```

The service listens on port `3000`.

### Access the service on a cluster

To use the Asset Metadata Service on a cluster, run the command:

```bash
kubectl port-forward deployment/assetstore-asset-metadata-service 3000:3000 -n kyma-system
```

You can access the service on port `3000`.

### Build a production version

To build the production Docker image, run this command:

```bash
docker build {image_name}:{image_tag}
```

The variables are:

- `{image_name}` that is the name of the output image. The default name is `asset-metadata-service`.
- `{image_tag}` that is the tag of the output image. The default tag is `latest`.

### Extract metadata from files

For the full API documentation, including OpenAPI specification, see the [Asset Store docs](../../docs/asset-store/docs/03-04-asset-metadata-service.md).

### Environment variables

Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|------|:----------:|---------|-------------|
| **APP_HOST** | No | `127.0.0.1` | The host on which the HTTP server listens |
| **APP_PORT** | No | `3000` | The port on which the HTTP server listens |
| **APP_VERBOSE** | No | No | The toggle used to enable detailed logs in the application |
| **APP_PROCESS_TIMEOUT** | No | `10m` | The file process timeout |
| **APP_MAX_WORKERS** | No | `10` | The maximum number of concurrent metadata extraction workers |

### Configure the logger verbosity level

This application uses `glog` to log messages. Pass command line arguments described in the [`glog.go`](https://github.com/golang/glog/blob/master/glog.go) file to customize the log parameters, such as the log level and output.

For example:
```bash
go run main.go --stderrthreshold=INFO -logtostderr=false
```

## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:
```bash
dep ensure -vendor-only
```

### Run tests

To run all unit tests, execute the following command:

```bash
go test ./...
```

### Verify the code

To check if the code is correct and you can push it, run the `before-commit.sh` script. It builds the application, runs tests, and checks the status of the vendored libraries. It also runs the static code analysis and ensures that the formatting of the code is correct.
