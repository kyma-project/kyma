# Asset Upload Service

## Overview

The Asset Upload Service is an HTTP server that exposes the file upload functionality for MinIO. It contains a simple HTTP endpoint which accepts `multipart/form-data` forms. It can upload files to the private and public system buckets.

## Prerequisites

Use the following tools to set up the project:

- [Go](https://golang.org)
- [Docker](https://www.docker.com/)

## Usage

### Run a local version

To run the application against the local Kyma installation on Minikube without building the binary, run this command:

```bash
APP_KUBECONFIG_PATH=/Users/$USER/.kube/config APP_VERBOSE=true APP_UPLOAD_ACCESS_KEY={accessKey} APP_UPLOAD_SECRET_KEY={secretKey} go run main.go
```

Replace values in curly braces with proper details, where:

- `{accessKey}` is the access key required to sign in to the content storage server.
- `{secretKey}` is the secret key required to sign in to the content storage server.

The service listens on port `3000`.

### Access on a cluster

To use the Asset Upload Service on a cluster, run the command:

```bash
kubectl port-forward deployment/assetstore-asset-upload-service 3000:3000 -n kyma-system
```

You can access the service on port `3000`.

### Build a production version

To build the production Docker image, run this command:

```bash
docker build {image_name}:{image_tag}
```

The variables are:

- `{image_name}` that is the name of the output image. The default name is `asset-upload-service`.
- `{image_tag}` that is the tag of the output image. The default tag is `latest`.

### Upload files

For the full API documentation, including OpenAPI specification, see the [Asset Store docs](https://kyma-project.io/docs/master/components/asset-store#details-asset-upload-service).

### Environmental variables

Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| **APP_HOST** | No | `127.0.0.1` | The host on which the HTTP server listens |
| **APP_PORT** | No | `3000` | The port on which the HTTP server listens |
| **APP_KUBECONFIG_PATH** | No | None | The path to the `kubeconfig` file, needed for running an application outside of the cluster |
| **APP_VERBOSE** | No | None | The toggle used to enable detailed logs in the application |
| **APP_UPLOAD_TIMEOUT** | No | `30m` | The file upload timeout |
| **APP_MAX_UPLOAD_WORKERS** | No | `10` | The maximum number of concurrent upload workers |
| **APP_UPLOAD_ENDPOINT** | No | `minio.kyma.local` | The address of the content storage server |
| **APP_UPLOAD_PORT** | No | `443` | The port on which the content storage server listens |
| **APP_UPLOAD_ACCESS_KEY** | Yes | None | The access key required to sign in to the content storage server |
| **APP_UPLOAD_SECRET_KEY** | Yes | None | The secret key required to sign in to the content storage server |
| **APP_UPLOAD_SECURE** | No | `true` | The HTTPS connection with the content storage server |
| **APP_UPLOAD_EXTERNAL_ENDPOINT** | No | `https://minio.kyma.local` | The external address of the content storage server. If not set, the system uses the `APP_UPLOAD_ENDPOINT` variable. |
| **APP_BUCKET_PRIVATE_PREFIX** | No | `private` | The prefix of the private system bucket |
| **APP_BUCKET_PUBLIC_PREFIX** | No | `public` | The prefix of the public system bucket |
| **APP_BUCKET_PUBLIC_PREFIX** | No | `us-east-1` | The region of the system buckets |
| **APP_CONFIG_ENABLED** | No | `true` | The toggle used to save and load the configuration using the ConfigMap resource |
| **APP_CONFIG_NAME** | No | `asset-upload-service` | ConfigMap resource name |
| **APP_CONFIG_NAMESPACE** | No | `kyma-system` | ConfigMap resource namespace |

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
