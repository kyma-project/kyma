---
title: Resources
type: Architecture
---

Service Catalog uses a set of custom resources provided either by Kubernetes or by Kyma itself.
These are the native [Kubernetes resources](https://kubernetes.io/docs/concepts/extend-kubernetes/service-catalog/#api-resources) that Service Catalog uses:

>**NOTE:** The "Cluster" prefix in front of resources means they are cluster-wide. Resources without that prefix refer to the Namespace scope.   

* **ClusterServiceBroker** is an endpoint for a set of managed services that a third party offers and maintains.

* **ClusterServiceClass** is a managed service exposed by a given ClusterServiceBroker. When a cluster administrator registers a new Service Broker in the Service Catalog, the Service Catalog controller obtains new services exposed by the Service Broker and renders them in the cluster as ClusterServiceClasses. A ClusterServiceClass is synonymous with a service in the Service Catalog.

* **ClusterServicePlan** is a variation of a ClusterServiceClass that offers different levels of quality, configuration options, and the cost of a given service. Contrary to the ClusterServiceClass, which is purely descriptive, the ClusterServicePlan provides technical information to the ClusterServiceBroker on this part of the service that the ClusterServiceBroker can expose.

* **ServiceBroker** is any Service Broker registered in a given Namespace where it exposes ServiceClasses and ServicePlans that are available only in that Namespace.

* **ServiceClass**  is a Namespace-scoped representation of a ClusterServiceClass. Similarly to the ClusterServiceClass, it is synonymous with a service in the Service Catalog.

* **ServicePlan** is a Namespace-scoped representation of a ClusterServicePlan.

* **ServiceInstance** is a provisioned instance of a ClusterServiceClass to use in one or more cluster applications.

* **ServiceBinding** is a link between a ServiceInstance and an application that a cluster user creates to request credentials or configuration details for a given ServiceInstance.

* **Secret** is a basic resource to transfer credentials or configuration details that the application uses to consume a ServiceInstance. The service binding process leads to the creation of a Secret.


These are the Service Catalog resources that Kyma provides:

* [**ServiceBindingUsage**](#custom-resource-service-binding-usage) is a Kyma custom resource that allows Secret injection into a given application.

* [**UsageKind**](#custom-resource-usage-kind) is a Kyma custom resource that defines which resources you can bind to the ServiceBinding and how to bind them.
