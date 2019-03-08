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
|controller|A [Kubernetes Controller](https://kubernetes.io/docs/concepts/workloads/controllers/) which reacts on a standard Kubernetes resource or manages [Custom Resource Definition](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/) resources. The component's name reflects the name of the primary resource it controls.|namespace-controller|
|controller-manager|A daemon that embeds all [Kubernetes Controller](https://kubernetes.io/docs/concepts/workloads/controllers/)s of a domain. Such approach can bring benefits in operations in contrast to shipping all controllers individual. A `controller-managaer` is named by the domain it s belonging to. |assetstore-controller-manager|
|operator|is a [Kubernetes Operator](https://coreos.com/operators/) covering  application specific logic for operatiion of the application, like steps to upscale a stateful applicatin. It usually reacts on changes to resources of a [Custom Resource Definition](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/). It uses the name of the application it operates. |application-operator|
|job| A [Kubernetes Job](https://kubernetes.io/docs/tasks/job/) which performs a task once or periodically. It uses the name of the task it performs. |istio-patch-job (not renamed yet)|
|proxy| Acts as a proxy for an existing component, usually introducing a security model for this component. It uses the component's name. | apiserver-proxy|
|service| Serves an HTTP/S-based API, usually securely exposed to the public. It uses the domain name and the API it serves.|connector-service|
|broker| Implements the [OpenServiceBroker](https://www.openservicebrokerapi.org/) specification to enrich the Kyma Service Catalog with the services of a provider. It uses the name of the provider it integrates with.|azure-broker|
|configurer| A one-time task which usually runs as an [Init Container](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/) in order to configure the application.|ark-plugins-configurer (not migrated yet)|
