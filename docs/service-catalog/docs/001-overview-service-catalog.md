---
title: Overview
type: Overview
---

The Service Catalog is a grouping of reusable, integrated services from all Service Brokers registered in Kyma. Its purpose is to provide an easy way for Kyma users to access services that the Service Brokers manage and use them in their applications.

Due to the fact that Kyma runs on Kubernetes, you can easily run, in Kyma, a service that a third party provides and maintains, such as a database. Without extensive knowledge about the clustering of such a datastore service and the responsibility for its upgrades and maintenance, you can easily provision an instance of the software offering that a Service Broker registered in Kyma exposes, and bind it with an application running in the Kyma cluster.

## Operations

You can perform the following cluster-wide operations in the Service Catalog:
- Expose the consumable services by listing them with all the details, including the documentation and the consumption plans.
- Consume the services by provisioning them in a given Environment, which is Kyma's representation of the Kubernetes Namespace.
- Bind the services to the applications through Secrets.
