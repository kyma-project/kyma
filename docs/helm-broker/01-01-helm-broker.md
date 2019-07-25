---
title: Overview
---

The Helm Broker is a [Service Broker](/components/service-catalog/#service-brokers-overview) which exposes Helm charts as Service Classes in the [Service Catalog](/components/service-catalog/). To do so, the Helm Broker uses the concept of addons. An addon is an abstraction layer over a Helm chart which provides all information required to convert the chart into a Service Class.

The Helm Broker fetches addons which contain a set of specific [files](#details-create-addons). You must place your addons in a repository of an appropriate [format](#details-create-addons-repository). The Helm Broker fetches default cluster-wide addons defined by the [helm-repos-urls](https://github.com/kyma-project/kyma/blob/master/resources/helm-broker/templates/default-addons-cfg.yaml) custom resource (CR). This CR contains URLs that point to the release of  [`addons`](https://github.com/kyma-project/addons/releases) repository compatible with a given [Kyma release](https://github.com/kyma-project/kyma/releases). You can also [configure](#details-fetch-addons-from-https-servers) the Helm Broker to fetch addons definitions from any remote HTTP or HTTPS server.

In Kyma, you can use addons to install the following Service Brokers:

* [Azure Service Broker](/components/service-catalog/#service-brokers-azure-service-broker)
* [AWS Service Broker](/components/service-catalog/#service-brokers-aws-service-broker)

To see all addons that the Helm Broker provides, go to the [`addons`](https://github.com/kyma-project/addons) repository.

The Helm Broker implements the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/profile.md#service-metadata) (OSB API).
To be compliant with the Service Catalog version used in Kyma, the Helm Broker supports only the following OSB API versions:
- v2.13
- v2.12
- v2.11

> **NOTE:** The Helm Broker does not implement the OSB API update operation.
