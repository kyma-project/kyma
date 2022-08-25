# Telemetry Operator

## Overview

The telemetry operator contains a logging controller that generates a Fluent Bit configuration from one or more LogPipeline and LogParser Custom Resources. The controller ensures that all Fluent Bit Pods run the current configuration by deleting Pods after the configuration has changed. See all [CRD attributes](apis/telemetry/v1alpha1/logpipeline_types.go) and some [examples](config/samples).

For now, creating Fluent Bit Pods is out of scope of the operator. An existing Fluent Bit DaemonSet is expected.

The generated ConfigMap (by default, `telemetry-fluent-bit-sections` in the `kyma-system` namespace) must be mounted to the Fluent Bit Pods and consumed by an `@INCLUDE` statement in an existing [configuration file](https://docs.fluentbit.io/manual/administration/configuring-fluent-bit/classic-mode/configuration-file). Fluent Bit parsers, file references, and environment variables are available in an additional ConfigMap or Secret.

See the flags that configure all ConfigMaps, Secret and DaemonSet names in [main.go](main.go).

The operator has been bootstrapped with [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) 3.6.0. Additional APIs can also be [added by Kubebuilder](https://book.kubebuilder.io/cronjob-tutorial/new-api.html).

## Development

### Prerequisites
- Install [kubebuilder 3.6.0](https://github.com/kubernetes-sigs/kubebuilder) which is the base framework for this controller
- Install [kustomize](https://github.com/kubernetes-sigs/kustomize) which lets you customize raw, template-free `yaml` files during local development (see `kustomize` make target)
- Install [Golang 1.19](https://golang.org/dl/) or newer (for local execution)
- Install [Docker](https://www.docker.com/get-started)
- Install [OpenSSL](https://www.openssl.org/) to generate webhook certificate for local execution

### Available Commands

For development, you can use the following commands:

- Run all tests and validation

```bash
make
```

- Regenerate YAML manifests (CRDs and ClusterRole)

```bash
make manifests-local
```

- Copy CRDs and ClusterRole to installation directory

```bash
make copy-manifests-local
```

- Install CRDs to cluster in current kubeconfig context

```bash
make install-crds-local
```

- Uninstall CRDs to cluster in current kubeconfig context

```bash
make uninstall-crds-local
```

- Run the operator locally (uses current kubeconfig context)

```bash
kubectl -n kyma-system scale deployment telemetry-operator --replicas=0 # Scale down in-cluster telemetry-operator
make run-local
```

- Build container image and deploy to cluster in current kubeconfig context. Deploy telemetry chart first, as described before. Then run the following commands to deploy your own operator image.

```bash
make build-image DOCKER_PUSH_DIRECTORY=<my container repo>
make push-image DOCKER_PUSH_DIRECTORY=<my container repo> DOCKER_TAG=latest
kubectl -n kyma-system set image deployment telemetry-operator manager=<my container repo>/telemetry-operator:latest
```
