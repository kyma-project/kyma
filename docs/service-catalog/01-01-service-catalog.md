---
title: Overview
---

The Service Catalog groups reusable, integrated services from the [Service Brokers](#service-brokers-overview) registered in Kyma. Its purpose is to provide you an easy way to access services that the Service Brokers manage, and use them in your applications. For example, you can easily consume services provided by the third-party cloud platforms, such as [Azure](#service-brokers-azure-service-broker), [AWS](#service-brokers-aws-service-broker), or [GCP](#service-brokers-gcp-service-broker). 

Due to the fact that Kyma runs on Kubernetes, you can easily instantiate a service that a third party provides and maintains, such as a database. You can consume it from Kyma without extensive knowledge about the clustering of such a datastore service and the responsibility for its upgrades and maintenance. You can also easily provision an instance of the software offering that a Service Broker registered in Kyma exposes, and bind it to an application running in the Kyma cluster.

You can perform the following operations in the Service Catalog:

- Expose the consumable services by listing them with all the details, including the documentation and the consumption plans.
- Consume the services by provisioning them in a given Namespace.
- Bind the services to the applications through Secrets.

>**NOTE:** Kyma uses the Service Catalog based on the one provided by [Kubernetes](https://github.com/kubernetes-sigs/service-catalog).
