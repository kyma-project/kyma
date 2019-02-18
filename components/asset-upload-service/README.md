# Asset Upload Service

## Overview

Asset Upload Service is a HTTP server, that exposes file upload capability to Minio. It contains a simple HTTP endpoint, which expects multipart form data input. 

## Prerequisites

Use the following tools to set up the project:

- [Go distribution](https://golang.org)
- [Docker](https://www.docker.com/)

## Usage

### Run a local version

To run the application without building the binary, against local Kyma installation on Minikube, run this command:

```bash
APP_KUBECONFIG_PATH=/Users/$USER/.kube/config APP_VERBOSE=true APP_UPLOAD_ACCESS_KEY={accessKey} APP_UPLOAD_SECRET_KEY={secretKey} go run main.go
```

Replace values in curly braces with proper details, where:
- `{accessKey}` is the access key required to sign in to the content storage server
- `{secretKey}` is the secret key required to sign in to the content storage server

The service listens on port `3000`.

### Access on cluster

In order to use Asset Upload Service on cluster, run the command:

```bash
kubectl port-forward deployment/assetstore-asset-upload-service 3000:3000 -n kyma-system
```

The service will be exposed on `3000` port.

### Build a production version

To build the production Docker image, run this command:

```bash
docker build {image_name}:{image_tag}
```

The variables are:

- `{image_name}` - name of the output image (default: `asset-upload-service`)
- `{image_tag}` - tag of the output image (default: `latest`)

### Upload files

To upload files, send a Multipart form POST request to `/upload` endpoint. The endpoint recognizes the following field names:

- `private` - array of files, which should be uploaded to private system bucket.  
- `private` - array of files, which should be uploaded to public read-only system bucket.  
- `directory` - optional directory, where the uploaded files are put. If it is not specified, it will be randomized.

To do the multipart request using `curl`, run the following command in this repository:

```bash
curl -v -F directory='example' -F private=@main.go -F private=@Gopkg.toml -F public=@Dockerfile http://localhost:3000/upload
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

### Configure logger verbosity level

This application uses `glog` to log messages. Pass command line arguments described in the [glog.go](https://github.com/golang/glog/blob/master/glog.go) document to customize the log, such as log level and output.

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

To check if the code is correct and you can push it, run the `before-commit.sh` script. It builds the application, runs tests, checks the status of the vendored libraries, runs the static code analysis, and ensures that the formatting of the code is correct.
