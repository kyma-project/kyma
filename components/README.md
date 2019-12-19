# Components

## Overview

The `components` directory contains the sources of all Kyma components.
A Kyma component is any Pod, container, or image deployed with and referenced in a Kyma module or chart to provide the module's functionality.
Each subdirectory in the `components` directory defines one component.

## Details

Every Kyma component resides in a dedicated folder which contains its sources and a `README.md` file. This file provides instructions on how to build and develop the component.

The component's name consists of a term describing the component, followed by the **component type**. The first part of the name may differ depending on the component's purpose.
This table lists the available types:

| type|description|example|
|--|--|--|
|controller|A [Kubernetes Controller](https://kubernetes.io/docs/concepts/workloads/controllers/) which reacts on a standard Kubernetes resource or manages [Custom Resource Definition](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/) resources. The component's name reflects the name of the primary resource it controls.|connectivity-certs-controller|
|controller-manager|A daemon that embeds all [Kubernetes Controllers](https://kubernetes.io/docs/concepts/workloads/controllers/) of a domain. Such an approach brings operational benefits in comparison to shipping all controllers separately. A `controller-manager` takes the name of the domain it belongs to. | - |
|operator|is a [Kubernetes Operator](https://coreos.com/operators/) which covers the application-specific logic behind the operation of the application, such as steps to upscale a stateful application. It reacts on changes made to custom resources derived from a given [CustomResourceDefinition](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/). It uses the name of the application it operates. |application-operator|
|job| A [Kubernetes Job](https://kubernetes.io/docs/tasks/job/) which performs a task once or periodically. It uses the name of the task it performs. |istio-patch-job (not renamed yet)|
|proxy| Acts as a proxy for an existing component, usually introducing a security model for this component. It uses the component's name. | apiserver-proxy|
|service| Serves an HTTP/S-based API, usually securely exposed to the public. It uses the domain name and the API it serves.|connector-service|
|broker| Implements the [OpenServiceBroker](https://www.openservicebrokerapi.org/) specification to enrich the Kyma Service Catalog with the services of a provider. It uses the name of the provider it integrates with.|azure-broker|
|configurer| A one-time task which usually runs as an [Init Container](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/) in order to configure the application.|dex-static-user-configurer

## Development

Follow [this](https://github.com/kyma-project/kyma/blob/master/resources/README.md) development guide when you add a new component to the `kyma` repository.
