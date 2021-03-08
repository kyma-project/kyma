# Eventing Controller

## Overview

This component contains controllers for various CustomResourceDefinitions related to eventing in Kyma. The following controllers come with this container:
- [`controller`](https://github.com/kyma-project/kyma/blob/master/components/eventing-controller/cmd/eventing-controller/main.go) which lays down the eventing infrastructure in Business Event Bus (BEB).
- [`nats-controller`](https://github.com/kyma-project/kyma/blob/master/components/eventing-controller/cmd/eventing-controller-nats/main.go) which lays down the eventing infrastructure in [NATS](https://docs.nats.io/nats-concepts/intro).

## Prerequisites
- Install [ko](https://github.com/google/ko) which is used to build and deploy the controller during local development
- Install [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) which is the base framework for this controller
- Install [kustomize](https://github.com/kubernetes-sigs/kustomize) which lets you customize raw, template-free `yaml` files during local development

### Installation

- To deploy the controllers inside a cluster, make sure you have `ko` installed and configured according to the [instructions](https://github.com/google/ko#setup).

- For `controller`, run:

    ```sh
    make deploy-local

    ## To verify all the manifests after the processing by Kustomize without applying to the cluster, use make target deploy-local-dry-run    
    make deploy-local-dry-run
    ```

- For `nats-controller`, run:

    ```sh
    make deploy-eventing-controller-nats-local

    ## To verify all the manifests processed by Kustomize, without applying them to the cluster, use the make target called "deploy-eventing-controller-nats-local-dry-run".    
    make deploy-eventing-controller-nats-local-dry-run
    ```

## Usage 

This section explains how to use the Eventing Controller.

- `controller` comes with the following command line argument flags:

    | Flag                    | Description                                           | Default Value   |
    | ----------------------- | ----------------------------------------------------- | --------------- |
    | `metrics-addr`            | The address the metric endpoint binds to.           | `:8080`         |
    | `reconcile-period`        | The period between triggering of reconciling calls. | `10 minutes`    |
    | `enable-debug-logs`       | Enable debug logs.                                  | `false`         |

- `nats-controller` comes with the following command line argument flags:

    | Flag                    | Description                                           | Default Value |
    | ----------------------- | ----------------------------------------------------- | ------------- |
    | `metrics-addr`            | The address the metric endpoint binds to.           | `:8080`       |
    | `reconcile-period`        | The period between triggering reconciling calls.    | `10 minutes`  |
    | `enable-debug-logs`       | The parameter that enables debug logs.              | `false`       |
    | `max-reconnects`          | The maximum number of reconnection attempts.        | `10 `         |
    | `reconnect-wait`         | Wait time between reconnection attempts.            | `1 second`    |

- `controller` expects the following environment variables:

    | Environment variable       | Description                                                                     |
    | -------------------------- | ------------------------------------------------------------------------------- |
    | **CLIENT_ID**              | The Client ID used to acquire Access Tokens from the Authentication server.     |
    | **CLIENT_SECRET**          | The Client Secret used to acquire Access Tokens from the Authentication server. |
    | **TOKEN_ENDPOINT**         | The Authentication Server Endpoint to provide Access Tokens.                    |
    | **WEBHOOK_CLIENT_ID**      | The Client ID used by webhooks to acquire Access Tokens from Kyma.              |
    | **WEBHOOK_CLIENT_SECRET**  | The Client Secret used by webhooks to acquire Access Tokens from Kyma.          |
    | **WEBHOOK_TOKEN_ENDPOINT** | The Kyma public endpoint to provide Access Tokens.                              |
    | **DOMAIN**                 | The Kyma cluster public domain.                                                 |

- `nats-controller` expects the following environment variables:

    | Environment variable   | Description                      |
    | ---------------------- | -------------------------------- |
    | **NATS_URL**           | The URL for the NATS server.     |

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

1. Export the following mandatory environment variables:

    * **KUBECONFIG** - path to a local kubeconfig file, if different from the default OS location.

2. Build the binary:

    ```sh
    make manager
    ```

3. Run the controller:

    ```sh
    make run
    ```
