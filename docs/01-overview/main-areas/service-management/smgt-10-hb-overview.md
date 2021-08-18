---
title: Helm Broker
---

Helm Broker is a [Service Broker](./smgt-02-brokers-overview.md) that exposes Helm charts as Service Classes in [Service Catalog](./smgt-01-sc-overview.md). To do so, Helm Broker uses the concept of addons. An addon is an abstraction layer over a Helm chart which provides all information required to convert the chart into a Service Class.

By default, Helm Broker fetches cluster-wide addons defined by the [helm-repos-urls](https://github.com/kyma-project/kyma/blob/main/resources/helm-broker/templates/default-addons-cfg.yaml) custom resource (CR). This CR contains URLs that point to the release of  [`addons`](https://github.com/kyma-project/addons/releases) repository compatible with a given [Kyma release](https://github.com/kyma-project/kyma/releases). You can also [create your own addons](../../../03-tutorials/00-service-management/smgt-11-hb-create-addons.md) and configure Helm Broker to fetch addons definitions from other addons repositories.

In Kyma, you can use addons to install the following Service Brokers:

* Azure Service Broker
* AWS Service Broker
* GCP Service Broker

To see all addons that Helm Broker provides, go to the [`addons`](https://github.com/kyma-project/addons) repository.

Helm Broker implements the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/profile.md#service-metadata) (OSB API).
To be compliant with the Service Catalog version used in Kyma, Helm Broker supports only the following OSB API versions:
- v2.13
- v2.12
- v2.11

> **NOTE:** Helm Broker does not implement the OSB API update operation.
