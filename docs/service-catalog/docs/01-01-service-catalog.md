---
title: Overview
---

The Service Catalog is a grouping of reusable, integrated services from all Service Brokers registered in Kyma. Its purpose is to provide an easy way for Kyma users to access services that the Service Brokers manage and use them in their applications.

Due to the fact that Kyma runs on Kubernetes, you can easily instantiate a service that a third party provides and maintains, such as a database. You can consume it from Kyma without extensive knowledge about the clustering of such a datastore service and the responsibility for its upgrades and maintenance. You can also easily provision an instance of the software offering that a Service Broker registered in Kyma exposes, and bind it to an application running in the Kyma cluster.

You can perform the following operations in the Service Catalog:

- Expose the consumable services by listing them with all the details, including the documentation and the consumption plans.
- Consume the services by provisioning them in a given Namespace.
- Bind the services to the applications through Secrets.
