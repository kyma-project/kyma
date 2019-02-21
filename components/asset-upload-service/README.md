# Asset Upload Service

## Overview

Asset Upload Service is a HTTP server, that exposes file upload capability to Minio. It contains a simple HTTP endpoint, which expects multipart form data input. 

## Prerequisites

Use the following tools to set up the project:

- [Go distribution](https://golang.org)
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

To upload files, send a Multipart form POST request to `/upload` endpoint. The endpoint recognizes the following field names:

- `private` - array of files, which should be uploaded to private system bucket.  
- `public` - array of files, which should be uploaded to public read-only system bucket.  
- `directory` - optional directory, where the uploaded files are put. If it is not specified, it will be randomized.

To do the multipart request using `curl`, run the following command in this repository:

```bash
curl -v -F directory='example' -F private=@main.go -F private=@Gopkg.toml -F public=@Dockerfile http://localhost:3000/v1/upload
```

The result is:

```json
{
   "uploadedFiles": [
      {
         "fileName": "Gopkg.toml",
         "remotePath": "https://minio.kyma.local/private-1b0sjap35m9o0/example/Gopkg.toml",
         "bucket": "private-1b0sjap35m9o0",
         "size": 212
      },
      {
         "fileName": "Dockerfile",
         "remotePath": "https://minio.kyma.local/public-1b0sjaq6t6jr8/example/Dockerfile",
         "bucket": "public-1b0sjaq6t6jr8",
         "size": 630
      },
      {
         "fileName": "main.go",
         "remotePath": "https://minio.kyma.local/private-1b0sjap35m9o0/example/main.go",
         "bucket": "private-1b0sjap35m9o0",
         "size": 4414
      }
   ]
}
```

See the [Swagger specification](../../docs/asset-store/docs/assets/asset-upload-service-swagger.yaml) to read full API documentation. You can use the [Swagger Editor](https://editor.swagger.io) to preview and test the API service.

## Configuration

This section describes how to configure the application.

### Environmental Variables

Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| APP_HOST | No | `127.0.0.1` | The host on which the HTTP server listens. |
| APP_PORT | No | `3000` | The port on which the HTTP server listens. |
| APP_KUBECONFIG_PATH | No |  | The path to the `kubeconfig` file, needed for running an application outside of the cluster. |
| APP_VERBOSE | No | No | The parameter which shows detailed logs in the application. |
| APP_UPLOAD_TIMEOUT | No | `30m` | Timeout for uploading files. |
| APP_MAX_UPLOAD_WORKERS | No | `10` | The maximum number of concurrent upload workers. |
| APP_UPLOAD_ENDPOINT | No | `minio.kyma.local` | The address of the content storage server. |
| APP_UPLOAD_PORT | No | `443` | The port on which the content storage server listens. |
| APP_UPLOAD_ACCESS_KEY | Yes |  | The access key required to sign in to the content storage server. |
| APP_UPLOAD_SECRET_KEY | Yes |  | The secret key required to sign in to the content storage server. |
| APP_UPLOAD_SECURE | No | `true` | Use HTTPS for the connection with the content storage server. |
| APP_UPLOAD_EXTERNAL_ENDPOINT | No | `https://minio.kyma.local` | The external address of the content storage server. If not set, the system uses the `APP_UPLOAD_ENDPOINT` variable. |
| APP_BUCKET_PRIVATE_PREFIX | No | `private` | The prefix of the private system bucket. |
| APP_BUCKET_PUBLIC_PREFIX | No | `public` | The prefix of the public system bucket. |
| APP_BUCKET_PUBLIC_PREFIX | No | `us-east-1` | The region of the system buckets. |
| APP_CONFIG_ENABLED | No | `true` | Toggle for config save and load using ConfigMap resource |
| APP_CONFIG_NAME | No | `asset-upload-service` | ConfigMap resource name |
| APP_CONFIG_NAMESPACE | No | `kyma-system` | ConfigMap resource namespace |

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
