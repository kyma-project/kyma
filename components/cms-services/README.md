# CMS Services

## Overview

The CMS Services project collects the source code for services that extend the functionality of Headless CMS and Asset Store. Currently, it only contains the [CMS AsyncAPI Service](cmd/asyncapi/README.md).

## Prerequisites

Use the following tools to set up the project:

- [Go](https://golang.org)
- [Docker](https://www.docker.com/)

## Development

Read how to develop, test, and validate the project.

### Install dependencies

This project uses `dep` as a dependency manager. To install all required dependencies, use the following command:

```bash
dep ensure --vendor-only --v
```

### Run tests

To run all unit tests, execute the following command:

```bash
go test ./...
```

### Verify the code

To check if the code is correct and you can push it, run the `before-commit.sh` script. It builds the application, runs tests, and checks the status of the vendored libraries. It also runs the static code analysis and ensures that code formatting is correct.
