# Permission Controller

## Overview
The Permission Controller listens for new Namespaces and creates a RoleBinding for the users of the specified group to the **kyma-admin** role within these Namespaces. The Controller uses a blacklist mechanism, which defines the Namespaces in which the users of the defined group are not assigned the **kyma-admin** role. When the Controller is deployed in a cluster, it checks all existing Namespaces and assigns the roles accordingly.

Click [here](/resources/permission-controller) to access the Helm chart that defines the component's installation.

## Prerequisites

- working Kubernetes cluster
- go 1.12 or higher
- docker

## Development

The development process uses the formulae declared in the [generic](/common/makefiles/generic-make-go.mk) and [component-specific](./Makefile) Makefiles.

### Verify your changes and run tests
Before each commit, use the `verify` formula to test your changes:
  ```bash
  make verify
  ```

### Build a Docker image

Use the the `build-image` formula to build a Docker image of the controller:
  ```bash
  make build-image
  ```

### Run on a Kubernetes cluster using local sources

Use the `run` formula to run the controller using local sources:
  ```bash
  make run EXCLUDED_NAMESPACES={EXCLUDED_NAMESPACES} SUBJECT_GROUPS={SUBJECT_GROUPS} STATIC_CONNECTOR={STATIC_CONNECTOR}
  ```
  
See [this file](/resources/permission-controller/README.md#configuration) to learn how to use the environment variables.
