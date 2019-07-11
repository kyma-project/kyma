---
title: Overview
---

The Helm Broker is a [Service Broker](/components/service-catalog/#service-brokers-overview) which exposes Helm charts as Service Classes in the [Service Catalog](/components/service-catalog/). To do so, the Helm Broker uses the concept of addons. Addons are abstraction layers over Helm charts which provide all necessary information to convert the charts into Service Classes.

The Helm Broker fetches addons which contain a set of specific [files](#details-create-addons). You must place your addons in a repository of an appropriate [format](#details-create-addons-repository). By default, the Helm Broker fetches cluster-wide addons defined in the [ClusterAddonsConfiguration](#custom-resource-clusteraddonsconfiguration) custom resource (CR). This CR contains URLs that point to the release of the Kyma [`bundles`](https://github.com/kyma-project/bundles/releases) repository. You can also [configure](#tutorials-tutorials) the Helm Broker to fetch addons definitions from any remote HTTP or HTTPS server.

In Kyma, you can use addons to install the following Service Brokers:

* [Google Cloud Platform (GCP) Broker](/components/service-catalog/#service-brokers-gcp-broker)
* [Azure Service Broker](/components/service-catalog/#service-brokers-azure-service-broker)
* [AWS Service Broker](/components/service-catalog/#service-brokers-aws-service-broker)

To get all addons that the Helm Broker provides, go to the [`bundles`](https://github.com/kyma-project/bundles) repository.

The Helm Broker implements the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/profile.md#service-metadata) (OSB API).
To be compliant with the Service Catalog version used in Kyma, the Helm Broker supports only the following OSB API versions:
- v2.13
- v2.12
- v2.11

> **NOTE:** The Helm Broker does not implement the OSB API update operation.
