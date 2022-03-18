# Telemetry Operator

## Overview

The telemetry operator contains a logging controller that generates a Fluent Bit configuration from one or more `LogPipeline` custom resources. The controller ensures that all Fluent Bit Pods run the current configuration by deleting Pods after the configuration has changed. See all [CRD attributes](api/v1alpha1/logpipeline_types.go) and an [example](config/samples/telemetry_v1alpha1_logpipeline.yaml).

For now, creating Fluent Bit Pods is out of scope of the operator. An existing Fluent Bit Daemon Set is expected.

The generated Config Map (by default, `telemetry-fluent-bit-sections` in the `kyma-system` namespace) must be mounted to the Fluent Bit Pods and consumed by an `@INCLUDE` statement in an existing [configuration file](https://docs.fluentbit.io/manual/administration/configuring-fluent-bit/classic-mode/configuration-file). Fluent Bit parsers, file references, and environment variables are available in an additional Config Map or Secret.

See the flags that configure all Config Maps, Secret and Daemon Set names in [main.go](main.go).

The operator has been bootstrapped with [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) 3.1.0. Additional APIs can also be [added by Kubebuilder](https://book.kubebuilder.io/cronjob-tutorial/new-api.html).

## Development

### Prerequisites
- Install [kubebuilder 3.2.0](https://github.com/kubernetes-sigs/kubebuilder) which is the base framework for this controller
- Install [kustomize](https://github.com/kubernetes-sigs/kustomize) which lets you customize raw, template-free `yaml` files during local development
- Install [Golang 1.17](https://golang.org/dl/) or newer (for local execution)
- Install [Docker](https://www.docker.com/get-started)

### Available Commands

For development, you can use the following commands:

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
make install-local
```

- Uninstall CRDs to cluster in current kubeconfig context

```bash
make uninstall-local
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
