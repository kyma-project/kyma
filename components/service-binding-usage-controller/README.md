# Service Binding Usage Controller

## Overview

Service Binding Usage Controller injects **Service Bindings** into a given application using the **Service Binding Usage** custom resource, which allows for binding this application to a given Service Instance. The Service Binding Usage is a Kubernetes custom resource which is Namespace-scoped.

For the custom resource definition, see the [Service Binding Usage CRD file](../../installation/resources/crds/service-catalog/servicebindingusages.servicecatalog.crd.yaml). For more information on the Service Binding Usage Controller, see the [docs](./docs) folder in this repository. You can also refer to the corresponding [Service Bindng Usage documentation](https://kyma-project-old.netlify.app/docs/components/service-catalog/#custom-resource-service-binding-usage) on the website.

## Prerequisites

To set up the project, download these tools:

* [Go](https://golang.org/dl/) 1.11.4
* [Dep](https://github.com/golang/dep) v0.5.0
* [Docker](https://www.docker.com/)

These Go and Dep versions are compliant with the `buildpack` used by Prow. For more details, read [the Buildpack Golang Docker Image README](https://github.com/kyma-project/test-infra/blob/main/prow/images/buildpack-golang/README.md).

## Usage

This section explains how to use the Service Binding Usage Controller.

### Run a local version

To run the application without building the binary file, run this command:

```bash
APP_KUBECONFIG_PATH=/Users/$User/.kube/config APP_LOGGER_LEVEL=debug go run cmd/controller/main.go
```

For the description of the available environment variables, see the **Use environment variables** section.

### Use environment variables

Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|-----|---------|--------|------------|
| **APP_PORT** | No | `3000` | The port on which the HTTP server listens. |
| **APP_LOGGER_LEVEL** | No | `info` | Show detailed logs in the application. |
| **APP_KUBECONFIG_PATH** | No | None | The path to the `kubeconfig` file that you need to run an application outside of the cluster. |

## Development

Use the `make verify` command to test your changes before each commit. To build an image, use the `make build-image` command with the **DOCKER_PUSH_REPOSITORY** and **DOCKER_PUSH_DIRECTORY** variables. For example:
```
DOCKER_PUSH_REPOSITORY=eu.gcr.io DOCKER_PUSH_DIRECTORY=/kyma-project/develop make build-image
```
