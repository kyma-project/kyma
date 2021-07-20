---
title: Helm Broker
---

Helm Broker is a [Service Broker](./smgt-02-brokers-overview.md) that exposes Helm charts as Service Classes in [Service Catalog](./smgt-01-sc-overview.md). To do so, Helm Broker uses the concept of addons. An addon is an abstraction layer over a Helm chart which provides all information required to convert the chart into a Service Class.

Helm Broker fetches addons, which contain a set of specific [files](./smgt-11-hb-create-addons.md). You must place your addons in a repository of an appropriate [format](./smgt-14-hb-create-addons-repo.md). Helm Broker fetches default cluster-wide addons defined by the [helm-repos-urls](https://github.com/kyma-project/kyma/blob/main/resources/helm-broker/templates/default-addons-cfg.yaml) custom resource (CR). This CR contains URLs that point to the release of  [`addons`](https://github.com/kyma-project/addons/releases) repository compatible with a given [Kyma release](https://github.com/kyma-project/kyma/releases). You can also configure Helm Broker to fetch addons definitions from other addons repositories.

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
