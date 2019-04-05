---
title: Overview
---

The Service Catalog groups reusable, integrated services from all [Service Brokers](#service-brokers-overview) registered in Kyma. Its purpose is to provide an easy way for Kyma users to access services that the Service Brokers manage and use them in their applications.

Due to the fact that Kyma runs on Kubernetes, you can easily instantiate a service that a third party provides and maintains, such as a database. You can consume it from Kyma without extensive knowledge about the clustering of such a datastore service and the responsibility for its upgrades and maintenance. You can also easily provision an instance of the software offering that a Service Broker registered in Kyma exposes, and bind it to an application running in the Kyma cluster.

You can perform the following operations in the Service Catalog:

- Expose the consumable services by listing them with all the details, including the documentation and the consumption plans.
- Consume the services by provisioning them in a given Namespace.
- Bind the services to the applications through Secrets.

There are two versions of Service Catalog that you can use:

- mainstream (official) version that uses its own instance of apiserver and etcd - this is the default choice
- experimental version that uses Custom Resource Definitions

To enable the experimental version you have to override parameters `service-catalog-apiserver.enabled` and `service-catalog-crds.enabled`
in the installer-config file. Modify or add the `service-catalog-overrides` config map:  
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: service-catalog-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: service-catalog
    kyma-project.io/installation: ""
data:
  service-catalog-apiserver.enabled: "false"
  service-catalog-crds.enabled: "true"
```
