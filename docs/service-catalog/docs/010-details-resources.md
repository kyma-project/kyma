---
title: Resources
type: Details
---

This document includes an overview of resources that the Kyma Service Catalog provides.

>**NOTE:** The "Cluster" prefix in front of resources means they are cluster-wide. Resources without that prefix refer to the Environment scope.   

* **ClusterServiceBroker** is an endpoint for a set of managed services that a third party offers and maintains.

* **ClusterServiceClass** is a managed service exposed by a given ClusterServiceBroker. When a cluster administrator registers a new Service Broker in the Service Catalog, the Service Catalog controller obtains new services exposed by the Service Broker and renders them in the cluster as ClusterServiceClasses. A ClusterServiceClass is synonymous with a service in the Service Catalog.

* **ClusterServicePlan** is a variation of a ClusterServiceClass that offers different levels of quality, configuration options, and the cost of a given service. Contrary to the ClusterServiceClass, which is purely descriptive, the ClusterServicePlan provides technical information to the ClusterServiceBroker on this part of the service that the ClusterServiceBroker can expose.

* **ServiceBroker** is a Remote Environment Broker registered in a given Environment. It exposes service classes and service plans only in this Environment.

* **ServiceClass**  is an Environment-wide representation of a ClusterServiceClass. Similarly to the ClusterServiceClass, it is synonymous with a service in the Service Catalog.

* **ServicePlan** is an Environment-wide representation of a ClusterServicePlan.

* **ServiceInstance** is a provisioned instance of a ClusterServiceClass to use in one or more cluster applications.

* **ServiceBinding** is a link between a ServiceInstance and an application that cluster users create to obtain access credentials for their applications.

* **Secret** is a basic resource to transfer logins and passwords to the Deployment. The service binding process leads to the creation of a Secret.

* **ServiceBindingUsage** is a Kyma custom resource that allows the ServiceBindingUsage controller to inject Secrets into a given application.

* **UsageKind** is a Kyma custom resource that allows you to bind a ServiceInstance to any resource.
