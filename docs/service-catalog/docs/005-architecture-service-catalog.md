---
title: Architecture
---

The diagram and steps describe the Service Catalog workflow and the roles of specific cluster and Environment-wide resources in this process:

![Service Catalog flow](./assets/service-catalog-flow.svg)

1. The Kyma installation results in the registration of the default Service Brokers in the Kyma cluster. The Kyma administrator can manually register other ClusterServiceBrokers in the Kyma cluster. The Kyma user can also register a Service Broker in a given Environment.

2. Inside the cluster, each ClusterServiceBroker exposes services that are ClusterServiceClasses in their different variations called ClusterServicePlans. Similarly, the ServiceBroker registered in a given Environment exposes ServiceClasses and ServicePlans only in this specific Environment.

3. In the Console UI or CLI, the Kyma user lists all exposed cluster-wide and Environment-specific services and requests to create instances of those services in the Environment.

4. The Kyma user creates bindings to the ServiceInstances to allow the given applications to access the provisioned services.
