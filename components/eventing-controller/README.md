# Eventing Controller

## Overview

The Eventing Controller components contain reconcilers for various CRDs related to Eventing in Kyma:

- Subscription reconciler, which lays down the Eventing infrastructure in Business Event Bus (BEB) or [NATS](https://docs.nats.io/nats-concepts/intro)
- Backend reconciler, which updates the status of the currently active Eventing backend

For eventing in NATS mode, Eventing Controller also acts as a dispatcher. It converts messages received from NATS to POST requests and forwards them to the configured sinks.

## Prerequisites

- Install [ko](https://github.com/ko-build/ko/blob/main/docs/install.md), which is used to build and deploy the controller during local development.
- Install [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), which is the base framework for this controller (currently used version: 3.1).
- Install [Kustomize](https://github.com/kubernetes-sigs/kustomize), which lets you customize raw, template-free `yaml` files during local development.
- Install lint on the local environment:
  
  ```bash
  curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b $GOPATH/bin
  ```

### Installation

1. Delete any existing deployment for Eventing Controller from the Kyma cluster, because it could interfere with the reconciliation process.

1. Make sure the environment variables are set.
   The make target `set-up-local-env` uses default values. Change them as needed.
   If you want to push your images to Docker Hub, set the env variable `KO_DOCKER_REPO` to `index.docker.io/<docker_id>`.

1. To verify all the manifests after the processing by Kustomize without applying them to the cluster, use the make target `deploy-dry-run`:

   ```sh
   make DOMAIN=custom-domain.com deploy-dry-run
   ```

1. When everything is set up as you need it, run:

   ```sh
   make DOMAIN=custom-domain.com deploy
   ```

## Usage

### Environment Variables

The Eventing Controller expects the following environment variables:

| Environment Variable              | Description                                                                                    |
|-----------------------------------|------------------------------------------------------------------------------------------------|
| **For both**                      |                                                                                                |
| `APP_LOG_FORMAT`                  | The format of the Application logs.                                                            |
| `APP_LOG_LEVEL`                   | The level of the Application logs.                                                             |
| `BACKEND_CR_NAMESPACE`            | The namespace of the Backend Resource (CR).                                                    |
| `BACKEND_CR_NAME`                 | The name of the Backend Resource (CR).                                                         |
| `PUBLISHER_IMAGE`                 | The image of the Event Publisher Proxy.                                                        |
| `PUBLISHER_IMAGE_PULL_POLICY`     | The pull-policy of the Event Publisher Proxy.                                                  |
| `PUBLISHER_PORT_NUM`              | The port number of the Event Publisher Proxy itself.                                           |
| `PUBLISHER_METRICS_PORT_NUM`      | The port number of the Event Publisher Proxy metrics.                                          |
| `PUBLISHER_SERVICE_ACCOUNT`       | The service account of the Event Publisher Proxy.                                              |
| `PUBLISHER_REPLICAS`              | The number of replicas of the Event Publisher Proxy.                                           |
| `PUBLISHER_REQUESTS_CPU`          | The CPU requests of the Event Publisher Proxy.                                                 |
| `PUBLISHER_REQUESTS_MEMORY`       | The memory requests of the Event Publisher Proxy.                                              |
| `PUBLISHER_LIMITS_CPU`            | The CPU limits of the Event Publisher Proxy.                                                   |
| `PUBLISHER_LIMITS_MEMORY`         | The memory limits of the Event Publisher Proxy.                                                |
| **For NATS**                      |                                                                                                |
| `NATS_URL`                        | The URL for the NATS server.                                                                   |
| `EVENT_TYPE_PREFIX`               | The event type prefix for the NATS and BEB backend.                                            |
| `MAX_IDLE_CONNS`                  | The maximum number of idle connections for the HTTP transport of the NATS backend.             |
| `MAX_CONNS_PER_HOST`              | The maximum connections per host for the HTTP transport of the NATS backend.                   |
| `MAX_IDLE_CONNS_PER_HOST`         | The maximum idle connections per host for the HTTP transport of the NATS backend.              |
| `IDLE_CONN_TIMEOUT`               | The idle timeout duration for the HTTP transport of the NATS backend.                          |
| `DEFAULT_MAX_IN_FLIGHT_MESSAGES`  | The maximum idle "in-flight messages" sent by NATS to the sink without waiting for a response. |
| `DEFAULT_DISPATCHER_RETRY_PERIOD` | The retry period for resending an event to a sink, if the sink doesn't return 2XX.             |
| `DEFAULT_DISPATCHER_MAX_RETRIES`  | The maximum number of retries to send an event to a sink in case of errors.                    |
| **For NATS JetStream**            |                                                                                                |
|  `JS_STREAM_NAME`                 | The name of the stream where all events are stored.                                            |
|  `JS_STREAM_STORAGE_TYPE`         | The storage type of the stream: `memory` or `file`.                                            |
|  `JS_STREAM_RETENTION_POLICY`     | The policy to delete events from the stream: `limits` or `interest`. See [NATS: Stream Limits, Retention, and Policy](https://docs.nats.io/using-nats/developer/develop_jetstream/model_deep_dive#stream-limits-retention-and-policy). |
|  `JS_STREAM_MAX_MSGS`             | The maximum number of messages in the stream. Used only when storage policy is set to `limits`. |
|  `JS_STREAM_MAX_BYTES`            | The maximum size of the stream in bytes. Used only when storage policy is set to `limits`.     |
|  `JS_CONSUMER_DELIVER_POLICY`     | The policy to deliver events to consumers from the stream. Supported values are: `all`, `last`, `last_per_subject`, and `new`. See [NATS: DeliverPolicy](https://docs.nats.io/nats-concepts/jetstream/consumers#deliverpolicy).      |
| **For BEB**                       |                                                                                                |
| `TOKEN_ENDPOINT`                  | The Authentication Server Endpoint to provide Access Tokens.                                   |
| `WEBHOOK_ACTIVATION_TIMEOUT`      | The timeout duration used for webhook activation to acquire Access Tokens for Kyma.            |
| `WEBHOOK_TOKEN_ENDPOINT`          | The Kyma public endpoint to provide Access Tokens.                                             |
| `EXEMPT_HANDSHAKE`                | The exemption handshake switch of the subscription protocol settings.                          |
| `QOS`                             | The quality of service setting of the subscription protocol settings.                          |
| `CONTENT_MODE`                    | The content mode of the subscription protocol settings.                                        |
| `DOMAIN`                          | The Kyma cluster public domain.                                                                |

### Command Line Arguments

The additional command line arguments are:

| Flag                     | Description                                                                  | Default Value | Backend |
|--------------------------|------------------------------------------------------------------------------|---------------|---------|
| `metrics-addr`           | The TCP address that the controller binds to for serving Prometheus metrics. | `:8080`       | Both    |
| `health-probe-bind-addr` | The TCP address that the controller binds to for serving health probes.      | `:8080`       | Both    |
| `ready-check-endpoint`   | The endpoint of the readiness probe.                                         | `readyz`      | Both    |
| `health-check-endpoint`  | The endpoint of the health probe.                                            | `healthz`     | Both    |
| `reconcile-period`       | The period between triggering of reconciling calls (BEB).                    | 10 minutes    | BEB     |
| `max-reconnects`         | The maximum number of reconnection attempts (NATS).                          | 10            | NATS    |
| `reconnect-wait`         | Wait time between reconnection attempts (NATS).                              | 1 second      | NATS    |

### Commands

- To install the CustomResourceDefinitions in a cluster, run:

  ```sh
  make install
  ```

- To uninstall the CustomResourceDefinitions in a cluster, run:

  ```sh
  make uninstall
  ```

- To install the sample custom resources in a cluster, run:

  ```sh
  make install-samples
  ```

- To uninstall the sample custom resources in a cluster, run:

  ```sh
  make uninstall-samples
  ```

## Development

Before a commit, check the code quality:

```bash
$ make check-code
```

### Project Setup

Before running the Eventing Controller component, run the following command once to pull software dependencies and run tests:

```sh
make test
## To download dependencies only
make resolve-local
```

### Generate Code During Local Development

If you want to know more about scaffolding code with Kubebuilder, read [Simplified Builder-Based Scaffolding](https://github.com/kubernetes-sigs/kubebuilder/blob/master/designs/simplified-scaffolding.md).

1. To add new APIs using [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) CLI followed by generating boilerplate code, run the following script:

   ```sh
   kubebuilder create api --group batch --version v1 --kind CronJob

   make manifests
   ```

1. Update fields in the **spec** of an existing CRD by modifying the Go file for the type, that is, `api/version/<crd>_types.go`. For example, to adjust the Subscriptions CRD, modify `api/v1alpha1/subscriptions_types.go`.
   After that, run the following command to generate boilerplate code:

   ```sh
   make manifests
   ```

1. To use the newly generated CRDs, copy them to installation folders of Kyma:

   ```sh
   make copy-crds
   ```

1. After updating fields for an existing CRD, add the necessary changes manually in the sample custom resources inside the folder `config/samples/`. For example, for subscriptions, update the fields manually in `config/samples/eventing_v1alpha1_subscriptioncomponents/eventing-controller/config/crd/bases/eventing.kyma-project.io_subscriptions.yaml`

1. The Kubebuilder bootstrapped files have been reduced to the bare minimum. If you need one of these files later (for example, for a webhook), get them either from [this PR](https://github.com/kyma-project/kyma/pull/9510/commits/6ce5b914c5ef175dea45c27ccca826becb1b5818) or create a sample Kubebuilder project and copy all required files from there:

   ```sh
   kubebuilder init --domain kyma-project.io
   ```

### Set Up the Environment

#### Start the Controller Locally

> **CAUTION:** Running the Eventing Controller in local developer mode is currently broken and needs adoption of the latest changes.

1. Set up port forwarding for the in-cluster NATS instance:

   ```sh
   kubectl port-forward -n kyma-system svc/eventing-nats 4222
   ```

2. Export the following environment variables:

   | ENV VAR                  | Description                                        | Optional | Default Value               |
   |--------------------------|----------------------------------------------------|----------|-----------------------------|
   | `KUBECONFIG`             | Path to a local kubeconfig file.                   | yes      | ~/.kube/config              |
   | `NATS_URL`               | URL of the NATS server.                            | no       | nats.nats.svc.cluster.local |
   | `EVENT_TYPE_PREFIX`      | The event type prefix for the NATS and BEB backend.        | yes      | sap.kyma.custom             |
   | `WEBHOOK_TOKEN_ENDPOINT` | The Kyma public endpoint to provide Access Tokens. | yes      | WEBHOOK_TOKEN_ENDPOINT      |
   | `DOMAIN`                 | Domain.                                            | yes      | example.com                 |

   ```sh
   export  NATS_URL=nats://localhost:4222
   ```

3. Run the Eventing Controller:

   ```sh
   make run
   ```

4. To run the controller using your IDE, specify the buildtag `local`.

   > **NOTE:** We currently support the buildtag `local` to avoid setting incorrect OwnerRefs in the PublisherProxy deployment when running the controller on a developer's machine. Essentially, the PublisherProxy deployment remains in the cluster although the controller is removed due to no OwnerRef in the PublisherProxy deployment.
