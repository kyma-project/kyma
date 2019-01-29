# Asset Store Controller Manager

## Overview

Asset Store is a Kubernetes-native solution for storing assets, such as documentation, images, API specifications, and client-side applications. It consists of the Asset Controller and the Bucket Controller.

## Prerequisites

Use the following tools to set up the project:

* [Go distribution](https://golang.org)
* [Docker](https://www.docker.com/)

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


## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:
```bash
dep ensure -vendor-only
```

### Run tests

To run all unit tests, use the following command:

```bash
make test
```
