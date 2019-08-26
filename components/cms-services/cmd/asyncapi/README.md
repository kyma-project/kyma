# CMS AsyncAPI Service

## Overview

The CMS AsyncAPI Service is an HTTP server that exposes the AsyncAPI processing functionality. It contains HTTP endpoints which accept `multipart/form-data` forms.  

## Prerequisites

Use the following tools to set up the project:

- [Go distribution](https://golang.org)
- [Docker](https://www.docker.com/)

## Usage

Read how to run and use the service.

### API

See the [OpenAPI specification](openapi.yaml) for the full API documentation. You can use the [Swagger Editor](https://editor.swagger.io/) to preview and test the API service.

### Run a local version

To run the local version of the CMS AsyncAPI Service without building the binary, run this command:

```bash
go run main.go
```

The service listens on port `3000`.

### Build a production version

To build the production Docker image, run this command:

```bash
make -C ../../ docker-build
```

### Environment variables

Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| **APP_SERVICE_HOST** | No | `127.0.0.1` | The host on which the HTTP server listens |
| **APP_SERVICE_PORT** | No | `3000` | The port on which the HTTP server listens |
| **APP_VERBOSE** | No | `false` | The toggle used to enable detailed logs in the application |
