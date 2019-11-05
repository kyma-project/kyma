# Knative event sources for Kyma

Available event sources:

* `HTTPSource`: for applications emitting HTTP events.

## Running the controller

### Inside a cluster

```console
$ ko apply -f config/
```

### Locally

#### Environment setup (performed once)

Create the CustomResourceDefinitions for event source types.

```console
$ kubectl create -f config/ -l kyma-project.io/crd-install=true
```

Create the `kyma-system` system namespace if it does not exist. The controller watches ConfigMaps inside it.

```console
$ kubectl create ns kyma-system
```

Create ConfigMaps for logging and observability inside the system namespace (`kyma-system`).

```console
$ kubectl create -f config/400-config-logging.yaml
$ kubectl create -f config/400-config-observability.yaml
```

#### Controller startup

Export the following mandatory environment variables:

* **KUBECONFIG**: path to a local kubeconfig file (if different from the default OS location)
* **HTTP_ADAPTER_IMAGE**: container image of the HTTP receiver adapter

Build the binary.

```console
$ make cmd/controller-manager
```

Run the controller.

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

See also [API changes > Generate Code](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md#generate-code) (Kubernetes Developer Guide).
