---
title: Helm Broker
type: Overview
---

The Helm Broker is an implementation of a service broker which runs in the Kyma cluster and deploys Kubernetes native resources using [Helm](https://github.com/kubernetes/helm) and Kyma bundles. A bundle is an abstraction layer over a Helm chart which allows you to represent it as a ClusterServiceClass in the Service Catalog. For example, a bundle can provide plan definitions or binding details. The Helm Broker fetches bundle definitions from an HTTP server. By default, the Helm Broker fetches bundles from the newest release with the [Semantic Versioning](https://semver.org/) pattern available at the [`bundles`](https://github.com/kyma-project/bundles/releases) repository.

Using bundles, you can also install other brokers, such as the [Google Cloud Platform Service Broker](https://github.com/kyma-project/bundles/tree/master/bundles/gcp-broker-provider-0.0.1).

For more information about the Service Brokers, see the **Service Brokers overview** document.

The Helm Broker implements the [Open Service Broker API][https://github.com/openservicebrokerapi/servicebroker/blob/v2.14/profile.md#service-metadata] (OSB API).
To be compliant with the Service Catalog version used in Kyma, the Helm Broker supports only the following versions of the OSB API:
- v2.13
- v2.12
- v2.11

> **NOTE:** The Helm Broker does not implement the OSB API update operation.
