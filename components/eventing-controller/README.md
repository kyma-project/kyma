# Eventing Controller

## Overview

This component contains controllers for various CustomResourceDefinitions related to eventing in Kyma. The controller comes with 2 containers:
- [controller](https://github.com/kubernetes-sigs/controller-runtime)
- [kube-rbac-proxy](https://github.com/brancz/kube-rbac-proxy)


## Prerequisites
- Install [ko](https://github.com/google/ko) which is used to build and deploy the controller during local development
- Install [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) which is the base framework for this controller
- Install [kustomize](https://github.com/kubernetes-sigs/kustomize) which lets you customize raw, template-free YAML files during local development

### Installation

- To deploy the controller inside a cluster, make sure you have `ko` installed and configured according to the [usage instructions](https://github.com/google/ko#usage), then run:

    ```sh
    make deploy-local

    ## To verify all the manifests after the processing by Kustomize without applying to the cluster use make target deploy-local-dry-run    
    make deploy-local-dry-run
    ```

## Usage 

This section explains how to use the Eventing Controller.

- The Eventing Controller comes with the following command line argument flags:

    | Flag    | Default Value | Description                                                                                   |
    | ----------------------- | ------------- |---------------------------------------------------------------------------------------------- |
    | metrics-addr            | :8080          | The address the metric endpoint binds to.
    | enable-leader-election            | false          | Enables leader-election in controller and ensures there is only one active controller. 
    | lease-duration            | 15s          | The duration the non-leader candidates will wait to force acquire leadership. Valid time units are ns, us (or µs), ms, s, m, h. 

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

- Add new apis using [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) CLI followed by generating boiler-plate code by executing the following script:

    ```sh
    kubebuilder create api --group batch --version v1 --kind CronJob

    make manifests
    ```

- Update fields in the `spec` of an existing CustomResourceDefinition by modifying the go file for the type i.e. `api/version/<crd>_types.go`. E.g. `api/v1alpha1/subscriptions_types.go` for Subscriptions CRD. After that, execute the following command to generate boiler-plate code:

    ```sh
    make manifests
    ```

- Add the necessary changes manually in the sample CustomResources after updating fields for an existing CustomResourceDefinition inside the folder `config/samples/`. For example, for subscriptions, update the fields manually in `config/samples/eventing_v1alpha1_subscriptioncomponents/eventing-controller/config/crd/bases/eventing.kyma-project.io_subscriptions.yaml.yaml`

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
