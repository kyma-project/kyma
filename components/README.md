# Components

## Overview

The components folder contains the sources of all Kyma components.
A Kyma component is any Pod/container/image deployed with and referenced in a Kyma module/chart to provide the module's functionality.
Each subfolder in the _components_ directory defines one component.

## Details

Every Kyma component has a dedicated folder containing its sources and a README.md containing further instructions on how to build and develop the component.

The component name and with that the folder name should be suffixed with its component type. The types are defined as following:

| type|description|example|
|--|--|--|
|controller|is a [Kubernetes Controller](https://kubernetes.io/docs/concepts/workloads/controllers/) reacting on standard Kubernetes resource or managing resources of a [Custom Resource Definition](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/). The name of the component reflects name of the primary resource which it controls.|namespace-controller|
|operator|is a [Kubernetes Operator](https://coreos.com/operators/) covering  application specific logic for operatiion of the application, like steps to upscale a stateful applicatin. It usually reacts on changes to resources of a [Custom Resource Definition](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/). It uses the name of the application it operates. |application-operator|
|job|is a [Kubernetes Job](https://kubernetes.io/docs/tasks/job/), performing a task once or periodically. It uses the name of the task it performs. |istio-patch-job (not renamed yet)|
|proxy| proxies an existing component, usually introducing a security model for the proxied component. It uses the component name. | apiserver-proxy|
|service| serves an HTTP/S-based API, usually exposed securely to the public. It uses the domain name and the API it serves.|connector-service|
|broker| is implementing the [OpenServiceBroker](https://www.openservicebrokerapi.org/) specification to enrich the Kyma Service Catalog with services of a provider. It uses the name of the provider it integrates with.|azure-broker|
|configurer|a one time task usually executed as an [init container](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/) in order to configure an application|ark-plugins-configurer (not migrated yet)|