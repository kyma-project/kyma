# Asset Store Controller Manager

## Overview

Asset Store is a Kubernetes-native solution for storing assets, such as documentation, images, API specifications, and client-side applications. It consists of the Asset Controller and the Bucket Controller.

## Prerequisites

Use the following tools to set up the project:

* [Go distribution](https://golang.org)
* [Docker](https://www.docker.com/)
* [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)

## Usage

### Run a local version

To run the application outside the cluster, run this command:

```bash
make run
```

### Build a production version

To build the production Docker image, run this command:

```bash
IMG={image_name}:{image_tag} make docker-build
```

The variables are:

* `{image_name}` which is the name of the output image. Use `assetstore-controller-manager` for the image name.
* `{image_tag}` which is the tag of the output image. Use `latest` for the tag name.


## Configuration

This document describes configuration details of the application.

### Environmental Variables

Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| APP_CLUSTER_ASSET_RELIST_INTERVAL | No | `5m` | The period of time after which the controller refreshes the ClusterAsset statuses. |
| APP_ASSET_RELIST_INTERVAL | No | `5m` | The period of time after which the controller refreshes the Asset statuses. |
| APP_BUCKET_RELIST_INTERVAL | No | `5m` | The period of time after which the controller refreshes the Bucket statuses. |
| APP_CLUSTER_BUCKET_RELIST_INTERVAL | No | `5m` | The period of time after which the controller refreshes the ClusterBucket statuses. |
| APP_STORE_ENDPOINT | No | `minio.kyma.local` | The address of the content storage server. |
| APP_STORE_EXTERNAL_ENDPOINT | No | `https://minio.kyma.local` | The external address of the content storage server. |
| APP_STORE_ACCESS_KEY | Yes |  | The access key required to sign in to the content storage server. |
| APP_STORE_SECRET_KEY | Yes |  | The secret key required to sign in to the content storage server. |
| APP_STORE_USE_SSL | No | `true` | Use HTTPS for the connection with the content storage server. |
| APP_WEBHOOK_MUTATION_TIMEOUT | No | `1m` | The period of time after which mutation will be canceled. |
| APP_WEBHOOK_VALIDATION_TIMEOUT | No | `1m` | The period of time after which validation will be canceled. |
| APP_LOADER_TEMPORARY_DIRECTORY | No | `/tmp` | The path directory used for temporary storing data. |

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
