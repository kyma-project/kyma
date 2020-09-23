# Eventing controller

## Overview

This component contains controllers for various CustomResourceDefinitions related to eventing in Kyma.

## Usage

### Project setup

Before running the component, execute the following command once to pull software dependencies, apply mandatory patches, and compile all binaries:

```shell script
$ make resolve-local
```

### Run the controller inside the cluster

To deploy the controller inside a cluster, make sure you have `ko` installed and configured according to the [usage instructions](https://github.com/google/ko#usage), then run:

```shell script
$ make deploy
```

#### Set up the environment

Follow these steps to set up the environment:

>**NOTE:** You only need to do this once.


#### Start the controller

1. Export the following mandatory environment variables:

    * **KUBECONFIG** - path to a local kubeconfig file, if different from the default OS location.

2. Build the binary:

    ```console
    $ make controller-manager
    ```

3. Run the controller:

    ```console
    $ ./controller-manager
    ```

## Development

- Create the CustomResourceDefinitions:

    ```shell script
    $ make install
    ```

### Custom types

The client code for custom API objects is generated using [code generators](https://github.com/kubernetes/code-generator/). This includes native REST clients, listers and informers, as well as injection code for Knative.

The client code must be re-generated whenever the API for custom types (`apis/.../types*.go`) changes:

```console
$ make codegen
```

For details, see [Generate code](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md#generate-code) section of the Kubernetes API changes guide.
