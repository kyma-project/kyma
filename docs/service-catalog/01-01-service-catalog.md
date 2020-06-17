---
title: Service Catalog
type: Overview
---

Service Catalog groups reusable integrated services from [Service Brokers](#overview-service-brokers) registered in Kyma. Its purpose is to provide you with an easy way to access services that the Service Brokers manage and use them in your application. For example, you can easily consume services provided by third-party cloud platforms, such as Azure, AWS, or GCP.

Due to the fact that Kyma runs on Kubernetes, you can easily instantiate a service provided and maintained by a third party. You can consume it without extensive knowledge on clustering of such a service, and without worrying about its upgrades and maintenance. You can also easily provision an instance of the software offering exposed by a Service Broker registered in Kyma and bind it to an application running in the Kyma cluster.

Using the Service Catalog, you can perform the following operations:

- Expose consumable services by listing them with all their details, including documentation and consumption plans.
- Consume services by provisioning them in a given Namespace.
- Bind services to applications.

>**NOTE:** Service Catalog used in Kyma is based on the [Service Catalog provided by Kubernetes](https://github.com/kubernetes-sigs/service-catalog).
