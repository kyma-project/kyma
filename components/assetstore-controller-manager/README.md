# Asset Store Controller Manager

## Overview

Asset Store is a generic, Kubernetes-native solution for storing documentation, images, API specifications, and client-side applications. It consists of two controllers: Asset Controller and Bucket Controller.

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
docker build {image_name}:{image_tag}
```

The variables are:

* `{image_name}` - name of the output image (recommended: `assetstore-controller-manager`)
* `{image_tag}` - tag of the output image (recommended: `latest`)


## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:
```bash
dep ensure -vendor-only
```

### Run tests

To run all unit tests, execute the following command:

```bash
make test
```
