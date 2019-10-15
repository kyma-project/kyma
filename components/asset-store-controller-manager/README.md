# Asset Store Controller Manager

## Overview

Asset Store is a Kubernetes-native solution for storing assets, such as documentation, images, API specifications, and client-side applications. It consists of the Asset Controller and the Bucket Controller.

## Prerequisites

Use the following tools to set up the project:

* [Go](https://golang.org)
* [Docker](https://www.docker.com/)
* [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)

## Usage

### Run a local version

To run the application outside the cluster, run this command:

```bash
GO111MODULE=on make run
```

### Build a production version

To build the production Docker image, run this command:

```bash
IMG={image_name}:{image_tag} make docker-build
```

The variables are:

* `{image_name}` which is the name of the output image. Use `asset-store-controller-manager` for the image name.
* `{image_tag}` which is the tag of the output image. Use `latest` for the tag name.

### Environmental Variables

Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| **APP_CLUSTER_ASSET_RELIST_INTERVAL** | No | `30s` | The period of time after which the controller refreshes the status of a ClusterAsset CR |
| **APP_CLUSTER_ASSET_MAX_CONCURRENT_RECONCILES** | No | `1` | The maximum number of cluster asset reconciles that can run in parallel |
| **APP_ASSET_RELIST_INTERVAL** | No | `30s` | The period of time after which the controller refreshes the status of an Asset CR |
| **APP_ASSET_MAX_CONCURRENT_RECONCILES** | No | `1` | The maximum number of asset reconciles that can run in parallel |
| **APP_BUCKET_RELIST_INTERVAL** | No | `30s` | The period of time after which the controller refreshes the status of a Bucket CR |
| **APP_BUCKET_MAX_CONCURRENT_RECONCILES** | No | `1` | The maximum number of bucket reconciles that can run in parallel |
| **APP_CLUSTER_BUCKET_RELIST_INTERVAL** | No | `30s` | The period of time after which the controller refreshes the status of a ClusterBucket |
| **APP_CLUSTER_BUCKET_MAX_CONCURRENT_RECONCILES** | No | `1` | The maximum number of cluster bucket reconciles that can run in parallel |
| **APP_STORE_ENDPOINT** | No | `minio.kyma.local` | The address of the content storage server |
| **APP_STORE_EXTERNAL_ENDPOINT** | No | `https://minio.kyma.local` | The external address of the content storage server |
| **APP_STORE_ACCESS_KEY** | Yes | None | The access key required to sign in to the content storage server |
| **APP_STORE_SECRET_KEY** | Yes | None | The secret key required to sign in to the content storage server |
| **APP_STORE_USE_SSL** | No | `true` | The variable that enforces the use of HTTPS for the connection with the content storage server |
| **APP_STORE_UPLOAD_WORKERS_COUNT** | No | `10` | The number of workers used in parallel to upload files to the storage bucket |
| **APP_WEBHOOK_MUTATION_TIMEOUT** | No | `1m` | The period of time after which mutation is canceled |
| **APP_WEBHOOK_MUTATION_WORKERS_COUNT** | No | `10` | The number of workers used in parallel to mutate files |
| **APP_WEBHOOK_METADATA_EXTRACTION_TIMEOUT** | No | `1m` | The period of time after which metadata extraction is canceled |
| **APP_WEBHOOK_VALIDATION_TIMEOUT** | No | `1m` | The period of time after which validation is canceled |
| **APP_WEBHOOK_VALIDATION_WORKERS_COUNT** | No | `10` | The number of workers used in parallel to validate files |
| **APP_LOADER_VERIFY_SSL** | No | `true` | The variable that verifies SSL certificate before downloading source files |
| **APP_LOADER_TEMPORARY_DIRECTORY** | No | `/tmp` | The path to the directory used to temporarily store data |

## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:

```bash
make resolve
```

### Run tests

To run all unit tests, use the following command:

```bash
make test
```
