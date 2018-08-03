# Binding Usage Controller

## Overview

The Binding Usage Controller injects the **ServiceBindings** into a given application using the **ServiceBindingUsage** resource, which allows this application to bind to a given ServiceInstance. The ServiceBindingUsage is a Kubernetes custom resource which is Namespace-scoped. For the custom resource definition, see the [ServiceBindingUsage CRD file](../../resources/cluster-essentials/templates/service-binding-usage.crd.yaml). For more detailed information on the Binding Usage Controller, see the [docs](./docs) folder in this repository.

## Prerequisites

To set up the project, use these tools:
* Version 1.9 or higher of [Go](https://golang.org/dl/)
* The latest version of [Docker](https://www.docker.com/)
* The latest version of [Dep](https://github.com/golang/dep)

## Usage

This section explains how to use the Binding Usage Controller.

### Run a local version
To run the application without building the binary file, run this command:

```bash
APP_KUBECONFIG_PATH=/Users/{User}/.kube/config APP_LOGGER_LEVEL=debug go run cmd/controller/main.go
```

For the description of the available environment variables, see the **Use environment variables** section.

### Use environment variables
Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|-----|---------|--------|------------|
| **APP_PORT** | No | `3000` | The port on which the HTTP server listens. |
| **APP_LOGGER_LEVEL** | No | `info` | Show detailed logs in the application. |
| **APP_KUBECONFIG_PATH** | No |  | The path to the `kubeconfig` file that you need to run an application outside of the cluster. |
| **APP_PLUGGABLE_SBU** | No | false | The feature flag that enables pluggable binding usage by **UsageKind** resources. 

## Development

Use the `before-commit.sh` script to test your changes before each commit.
