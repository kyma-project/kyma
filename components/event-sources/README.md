# Knative Event Sources for Kyma

## Overview

This component contains controllers and adapters for custom Knative sources.

Available event sources:

* `HTTPSource` - used for applications emitting HTTP events.

## Usage

### Project setup

Before running the component, execute the following command once to pull software dependencies, apply mandatory patches, and compile all binaries:

```console
$ make
```

### Run the controller inside the cluster

To deploy the controller inside a cluster, make sure you have `ko` installed and configured according to the [usage instructions](https://github.com/google/ko#usage), then run:

```console
$ ko apply -f config/
```

### Run the controller locally

#### Set up the environment

Follow these steps to set up the environment:

>**NOTE:** You only need to do this once.

1. Create the CustomResourceDefinitions for event source types:

    ```console
    $ kubectl create -f config/ -l kyma-project.io/crd-install=true
    ```

2. Create the `kyma-system` Namespace if it does not exist. The controller watches ConfigMaps inside it.

    ```console
    $ kubectl create ns kyma-system
    ```

3. Create ConfigMaps for logging and observability in the `kyma-system` Namespace.

    ```console
    $ kubectl create -f config/400-config-logging.yaml
    $ kubectl create -f config/400-config-observability.yaml
    ```

#### Start the controller

1. Export the following mandatory environment variables:

    * **KUBECONFIG** - path to a local kubeconfig file, if different from the default OS location.
    * **HTTP_ADAPTER_IMAGE** - container image of the HTTP adapter.

2. Build the binary:

    ```console
    $ make controller-manager
    ```

3. Run the controller:

    ```console
    $ ./controller-manager
    ```

## Development

### Custom types

The client code for custom API objects is generated using [code generators](https://github.com/kubernetes/code-generator/). This includes native REST clients, listers and informers, as well as injection code for Knative.

The client code must be re-generated whenever the API for custom types (`apis/.../types*.go`) changes:

```console
$ make codegen
```

For details, see [Generate code](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md#generate-code) section of the Kubernetes API changes guide.
