# Knative event sources for Kyma

Available event sources:

* `HTTPSource`: for applications emitting HTTP events.

## Running the controller inside a cluster

```console
$ ko apply -f config/
```

## Running the controller locally

### Environment setup (performed once)

Create the CustomResourceDefinitions for event source types.

```console
$ kubectl create -f config/ -l kyma-project.io/crd-install=true
```

Create the `kyma-system` system namespace. The controller watches ConfigMaps inside it.

```console
$ kubectl create -f config/100-namespace.yaml
```

Create ConfigMaps for logging and observability inside the system namespace (`kyma-system`)

```console
$ kubectl create -f config/400-config-logging.yaml
$ kubectl create -f config/400-config-observability.yaml
```

### Controller startup

Export the following mandatory environment variables:

* **KUBECONFIG**: path to a local kubeconfig file (if different from the default OS location)
* **SYSTEM_NAMESPACE**: set to "kyma-system" (see above)
* **METRICS_DOMAIN**: domain of the exposed Prometheus metrics. Arbitrary value (e.g. "kyma-project.io/event-sources")
* **HTTP_ADAPTER_IMAGE**: container image of the HTTP receiver adapter

Build the binary

```console
$ make cmd/controller-manager
```

Run the controller

```console
$ ./controller-manager
```
