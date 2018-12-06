# Docs Builder

## Overview

This project is used for building and pushing documentation-related Docker images. They are built from sources (`/docs` directory on this repository).

> **Note:** This is a temporary solution, before migration to the new Documentation Delivery concept. 

## Prerequisites

Use the following tools to set up the project:

* [Go distribution](https://golang.org)
* [Docker](https://www.docker.com/)

## Development

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:
```bash
dep ensure -vendor-only
```

### Run

```bash

APP_DOCKER_IMAGE_PREFIX={imagePrefix} APP_DOCKER_IMAGE_TAG={imageTag} go run main.go

```

Replace values in curly braces with proper details, where:
- `{imagePrefix}` is the prefix, which will be set for all built Docker images
- `{imageTag}` is the tag, which will be set for all built Docker images


### Verify the code

To check if the code is correct and you can push it, run the `before-commit.sh` script. It builds the application, runs tests, checks the status of the vendored libraries, runs the static code analysis, and ensures that the formatting of the code is correct.
