# Redis Cluster

Use the Cloud Manager module to provision a Redis cluster.

The Cloud Manager module allows you to provision a cloud provider-managed Redis cluster in cluster mode within your cluster network.

> [!NOTE]
> Using the Cloud Manager module and enabling Redis, introduces additional costs. For more information, see [Calculation with the Cloud Manager Module](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/commercial-information-sap-btp-kyma-runtime?state=DRAFT&version=Internal#loioc33bb114a86e474a95db29cfd53f15e6__section_cloud_manager).

## Cloud Providers

When you create a Redis cluster in SAP BTP, Kyma runtime, you depend on the cloud provider of your Kyma cluster. The cloud provider in use determines the exact implementation.

The Cloud Manager module supports the Redis cluster feature of the following cloud providers:

* Microsoft Azure [Azure Cache for Redis](https://azure.microsoft.com/en-us/products/cache)

You can configure Cloud Manager's Redis clusters using a dedicated Redis cluster custom resource (CR) corresponding with the cloud provider for your Kyma cluster, namely AzureRedisCluster CR. For more information, see [Redis Resources](./resources/README.md#redis-resources).

### Tiers

When you provision a Redis cluster, you must use the Premium tier. 

## Prerequisites

To instantiate Redis cluster, an IpRange CR must exist in the Kyma cluster. IpRange defines network address space reserved for your cloud provider's Redis resources. If you don't create the IpRange CR manually, Cloud Manager creates a default IpRange CR with the default address space and Classless Inter-Domain Routing(CIDR) selected. For more information, see [IpRange Custom Resoucre](./resources/04-10-iprange.md).

## Lifecycle

AzureRedisCluster is namespace-level CR. Once you create any of the Redis cluster resources, the following resources are also created automatically:

* IpRange CR
  * IpRange is a cluster-level CR.
  * Only one IpRange CR can exist per cluster.
  * If you don't want the default IpRange to be used, create one manually.
* Secret CR
  * The Secret is a namespace-level CR.
  * The Secret's name is the same as the name of the respective Redis cluster CR.
  * The Secret holds values and information used to access the Redis cluster.
