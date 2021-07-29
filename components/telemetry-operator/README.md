# Telemetry Operator

## Overview

The telemetry operator contains a logging controller that generates a Fluent Bit configuration from one or more `LoggingConfiguration` custom resources. The controller ensures that all Fluent Bit pods run the current configuration by deleting pods after the configuration has changed. Find all CRD attributes [here](api/v1alpha1/loggingconfiguration_types.go) and an example [here](config/samples/telemetry_v1alpha1_loggingconfiguration.yaml).

For now, creating Fluent Bit pods is out of scope of the operator. An existing Fluent Bit DaemonSet is expected.

The generated ConfigMap (by default, `logging-fluent-bit-sections` in the `kyma-system` namespace) must be mounted to the Fluent Bit pods and consumed by an `@INCLUDE` statement in an existing [configuration file](https://docs.fluentbit.io/manual/administration/configuring-fluent-bit/configuration-file). File references and environment variables are available in an additional ConfigMap or Secret.

See the flags that configure all ConfigMaps, Secret and DaemonSet names in [main.go](main.go).

The operator has been bootstrapped with [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) 3.1.0. Additional APIs can also be [added by Kubebuilder](https://book.kubebuilder.io/cronjob-tutorial/new-api.html).

## Development

### Prerequisites
- Install [kubebuilder 3.1.0](https://github.com/kubernetes-sigs/kubebuilder) which is the base framework for this controller
- Install [kustomize](https://github.com/kubernetes-sigs/kustomize) which lets you customize raw, template-free `yaml` files during local development
- Install [Golang 1.16](https://golang.org/dl/) or newer (for local execution)
- Install [Docker](https://www.docker.com/get-started)

### Usage

You can do the following:

- Run all tests and validation

```bash
make
```

- Regenerate YAML manifests

```bash
make manifests
```

- Install CRDs to cluster in current kubeconfig context

```bash
make install
```

- Uninstall CRDs to cluster in current kubeconfig context

```bash
make uninstall
```

- Run the operator locally (uses current kubeconfig context)

```bash
make run-local
```

- Build container image and deploy to cluster in current kubeconfig context

```bash
make build-image IMG_NAME=<my container repo>
make push-image IMG_NAME=<my container repo> TAG=latest
make deploy-local IMG_NAME=<my container repo> TAG=latest
```

- Remove controller from cluster in current kubeconfig context

```bash
make undeploy-local
```
