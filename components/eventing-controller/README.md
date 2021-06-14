# Eventing Controller

## Overview

This component contains controllers for various CustomResourceDefinitions related to Eventing in Kyma. The following controller comes with this container:

- [`controller`](https://github.com/kyma-project/kyma/blob/main/components/eventing-controller/cmd/eventing-controller/main.go) which lays down the Eventing infrastructure in Business Event Bus (BEB) or [NATS](https://docs.nats.io/nats-concepts/intro).

## Prerequisites

- Install [ko](https://github.com/google/ko) which is used to build and deploy the controller during local development
- Install [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) which is the base framework for this controller
- Install [kustomize](https://github.com/kubernetes-sigs/kustomize) which lets you customize raw, template-free `yaml` files during local development

### Installation

- To deploy the controllers inside a cluster, make sure you have `ko` installed and configured according to the [instructions](https://github.com/google/ko#setup). Then run:

```sh
make deploy-local
```

- To verify all the manifests after the processing by Kustomize without applying to the cluster, use make target `deploy-local-dry-run`:

```sh
make deploy-local-dry-run
```

## Usage

This section explains how to use the Eventing Controller. It expects the following environment variables:

| Environment Variable          | Description                                                                         |
| ----------------------------- | ----------------------------------------------------------------------------------- |
| **For both**                  |                                                                                     |
| `APP_LOG_FORMAT`              | The format of the Application logs.                                                 |
| `APP_LOG_LEVEL`               | The level of the Application logs.                                                  |
| `BACKEND_CR_NAMESPACE`        | The namespace of the Backend Resource (CR).                                         |
| `BACKEND_CR_NAME`             | The name of the Backend Resource (CR).                                              |
| `PUBLISHER_IMAGE`             | The image of the Event Publisher Proxy.                                             |
| `PUBLISHER_IMAGE_PULL_POLICY` | The pull-policy of the Event Publisher Proxy.                                       |
| `PUBLISHER_PORT_NUM`          | The port number of the Event Publisher Proxy itself.                                |
| `PUBLISHER_METRICS_PORT_NUM`  | The port number of the Event Publisher Proxy metrics.                               |
| `PUBLISHER_SERVICE_ACCOUNT`   | The service account of the Event Publisher Proxy.                                   |
| `PUBLISHER_REPLICAS`          | The number of replicas of the Event Publisher Proxy.                                |
| `PUBLISHER_REQUESTS_CPU`      | The CPU requests of the Event Publisher Proxy.                                      |
| `PUBLISHER_REQUESTS_MEMORY`   | The memory requests of the Event Publish Proxy.                                     |
| `PUBLISHER_LIMITS_CPU`        | The CPU limits of the Event Publisher Proxy.                                        |
| `PUBLISHER_LIMITS_MEMORY`     | The memory limits of the Event Publisher Proxy.                                     |
| **For NATS**                  |                                                                                     |
| `NATS_URL`                    | The URL for the NATS server.                                                        |
| `EVENT_TYPE_PREFIX`           | The Event Type Prefix for the NATS backend.                                         |
| `MAX_IDLE_CONNS`              | The maximum number of idle connecttions for the HTTP transport of the NATS backend. |
| `MAX_CONNS_PER_HOST`          | The maximum connections per host for the HTTP transport of the NATS backend.        |
| `MAX_IDLE_CONNS_PER_HOST`     | The maximum idle connections per host for the HTTP transport of the NATS backend.   |
| `IDLE_CONN_TIMEOUT`           | The idle timeout duration for the HTTP transport of the NATS backend.               |
| **For BEB**                   |                                                                                     |
| `TOKEN_ENDPOINT`              | The Authentication Server Endpoint to provide Access Tokens.                        |
| `WEBHOOK_ACTIVATION_TIMEOUT`  | The timeout duration used for webhook activation to acquire Access Tokens for Kyma. |
| `WEBHOOK_CLIENT_ID`           | The Client ID used by webhooks to acquire Access Tokens from Kyma.                  |
| `WEBHOOK_CLIENT_SECRET`       | The Client Secret used by webhooks to acquire Access Tokens from Kyma.              |
| `WEBHOOK_TOKEN_ENDPOINT`      | The Kyma public endpoint to provide Access Tokens.                                  |
| `EXEMPT_HANDSHAKE`            | The exemption handshake switch of the subscription protocol settings.               |
| `QOS`                         | The quality of service setting of the subscription protocol settings.               |
| `CONTENT_MODE`                | The content mode of the subscription protocol settings.                              |
| `DOMAIN`                      | The Kyma cluster public domain.                                                     |

The additional command line arguments are:

| Flag                     | Description                                               | Default Value | Backend |
| ------------------------ | --------------------------------------------------------- | ------------- | ------- |
| `metrics-addr`           | The TCP address that the controller binds to for serving Prometheus metrics.  | `:8080` | Both  |
| `health-probe-bind-addr` | The TCP address that the controller binds to for serving health probes.       | `:8080` | Both  |
| `ready-check-endpoint`   | The endpoint of the readiness probe.                      | `readyz`        | Both    |
| `health-check-endpoint`  | The endpoint of the health probe.                         | `healthz`       | Both    |
| `reconcile-period`       | The period between triggering of reconciling calls (BEB). | 10 minutes    | BEB     |
| `max-reconnects`         | The maximum number of reconnection attempts (NATS).       | 10            | NATS    |
| `reconnect-wait`         | Wait time between reconnection attempts (NATS).           | 1 second      | NATS    |

- To install the CustomResourceDefinitions in a cluster, run:

```sh
make install
```

- To uninstall the CustomResourceDefinitions in a cluster, run:

```sh
make uninstall
```

- To install the sample CustomResources in a cluster, run:

```sh
make install-samples
```

- To uninstall the sample CustomResources in a cluster, run:

```sh
make uninstall-samples
```

## Development

### Project setup

Before running the component, execute the following command once to pull software dependencies and run tests:

```sh
make test
## To download dependencies only
make resolve-local
```

### Generate code during local development

> More details on scaffolding code using kubebuilder can be found [here](https://github.com/kubernetes-sigs/kubebuilder/blob/master/designs/simplified-scaffolding.md).

- Add new APIs using [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) CLI followed by generating boilerplate code by executing the following script:

```sh
kubebuilder create api --group batch --version v1 --kind CronJob

make manifests
```

- Update fields in the `spec` of an existing CustomResourceDefinition by modifying the Go file for the type i.e. `api/version/<crd>_types.go`. For example, `api/v1alpha1/subscriptions_types.go` for Subscriptions CRD. After that, execute the following command to generate boilerplate code:

```sh
make manifests
```

- Add the necessary changes manually in the sample CustomResources after updating fields for an existing CustomResourceDefinition inside the folder `config/samples/`. For example, for subscriptions, update the fields manually in `config/samples/eventing_v1alpha1_subscriptioncomponents/eventing-controller/config/crd/bases/eventing.kyma-project.io_subscriptions.yaml.yaml`

- The kubebuilder bootstrapped files have been reduced to the bare minimum. If at a later point one of theses files are required (e.g. for a webhook), get them either from [this PR](https://github.com/kyma-project/kyma/pull/9510/commits/6ce5b914c5ef175dea45c27ccca826becb1b5818) or create a sample kubebuilder project and copy all required files from there:

```sh
kubebuilder init --domain kyma-project.io
```

### Set up the environment

#### Start the controller locally

> Currently running the controller in local developer mode is broken and needs adoptions of the latest changes.

1. Export the following mandatory environment variables:

| ENV VAR                  | Description                                                            | Default Value          |
| ------------------------ | ---------------------------------------------------------------------- | ---------------------- |
| `KUBECONFIG`             | Path to a local kubeconfig file.                                       | ~/.kube/config         |
| `NATS_URL`               | URL of the NATS server.                                                | nats://127.0.0.1:4222  |
| `EVENT_TYPE_PREFIX`      | Path to a local kubeconfig file.                                       | sap.kyma.custom        |
| `WEBHOOK_CLIENT_ID`      | The Client ID used by webhooks to acquire Access Tokens from Kyma.     | WEBHOOK_CLIENT_ID      |
| `WEBHOOK_CLIENT_SECRET`  | The Client Secret used by webhooks to acquire Access Tokens from Kyma. | WEBHOOK_CLIENT_SECRET  |
| `WEBHOOK_TOKEN_ENDPOINT` | The Kyma public endpoint to provide Access Tokens.                     | WEBHOOK_TOKEN_ENDPOINT |
| `DOMAIN`                 | Domain.                                                                | example.com            |

2. Build the binary:

```sh
make manager
```

3. Run the controller:

```sh
make run
```
