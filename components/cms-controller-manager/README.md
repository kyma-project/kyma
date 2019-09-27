# Content Management System (CMS) Controller Manager

## Overview

CMS is a Kubernetes-native solution for managing content. It is built on top of the Asset Store and consists of the DocsTopic Controller and the ClusterDocsTopic Controller.

## Prerequisites

Use the following tools to set up the project:

* [Go](https://golang.org)
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

* `{image_name}` which is the name of the output image. Use `cms-controller-manager` for the image name.
* `{image_tag}` which is the tag of the output image. Use `latest` for the tag name.

### Environment Variables

Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| **APP_CLUSTER_DOCS_TOPIC_RELIST_INTERVALL** | No | `5m` | The period of time after which the controller refreshes the status of a ClusterDocsTopic CR. |
| **APP_CLUSTER_BUCKET_REGION** | No | None | Specifies the location of the region in which the controller creates a ClusterBucket CR. If the field is empty, the controller creates the bucket under the default location. |
| **APP_DOCS_TOPIC_RELIST_INTERVALL** | No | `5m` | The period of time after which the controller refreshes the status of a DocsTopic CR |
| **APP_BUCKET_REGION** | No | None | Specifies the location of the region in which the controller creates a Bucket CR. If the field is empty, the controller creates the bucket under the default location. |
| **APP_WEBHOOK_CFG_MAP_NAME** | No | webhook-configmap | The name of the ConfigMap that contains webhook definitions. |
| **APP_WEBHOOK_CFG_MAP_NAMESPACE** | No | webhook-configmap | The namespace of the ConfigMap that contains webhook definitions. |

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
