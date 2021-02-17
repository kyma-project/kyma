# Eventing Controller

## Overview

This component contains controllers for various CustomResourceDefinitions related to eventing in Kyma. The controller comes with this container:
- [controller](https://github.com/kubernetes-sigs/controller-runtime)

## Prerequisites
- Install [ko](https://github.com/google/ko) which is used to build and deploy the controller during local development
- Install [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) which is the base framework for this controller
- Install [kustomize](https://github.com/kubernetes-sigs/kustomize) which lets you customize raw, template-free `yaml` files during local development

### Installation

- To deploy the controller inside a cluster, make sure you have `ko` installed and configured according to the [usage instructions](https://github.com/google/ko#usage), then run:

    ```sh
    make deploy-local

    ## To verify all the manifests after the processing by Kustomize without applying to the cluster, use make target deploy-local-dry-run    
    make deploy-local-dry-run
    ```

## Usage 

This section explains how to use the Eventing Controller.

- The Eventing Controller comes with the following command line argument flags:

    | Flag    | Default Value | Description                                                                                   |
    | ----------------------- | ------------- |---------------------------------------------------------------------------------------------- |
    | metrics-addr            | :8080          | The address the metric endpoint binds to.

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
