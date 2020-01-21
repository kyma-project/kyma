# Permission-controller

## Overview
Permission-controller provides a mechanism for granting admin privileges within custom Namespaces to selected user groups. Under the hood, the component is a Kubernetes controller that watches for instances of `Namespace core/v1` objects and ensures desired RBAC configuration by creating and updating objects of type `rolebindings.rbac.authorization.k8s.io`.

Click [here](/resources/permission-controller) to access the Helm chart that defines the component's installation.

## Prerequisites

- working Kubernetes cluster
- go 1.12 or higher
- docker

## Development

The development process uses formulae declared in the [generic](/common/makefiles/generic-make-go.mk) and [component-specific](./Makefile) Makefiles.

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