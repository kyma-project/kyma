---
title: Basic architecture
type: Architecture
---

The Helm Broker is installed alongside other Kyma components and it automatically registers itself in the Service Catalog as a ClusterServiceBroker. The installation provides the default [helm-repos-urls](#link) ClusterAddonsConfiguration (CAC) custom resource (CR). It contains URLs from which Helm Broker fetches addons. You can also [add your own addons](#details-fetch-addons-from-https-servers) with URLs that point to any remote HTTPS server.

If you want the Helm Broker to act as a Namespace-scoped ServiceBroker, create the [AddonsConfiguration](#custom-resource-addonsconfiguration) (AC) custom resource. In such a case, the Helm Broker creates a service and registers itself in the Service Catalog as a ServiceBroker inside the Namespace in which the CR is created.

The Helm Broker workflow starts with the registration process, during which the Helm Broker fetches addons from URLs provided in the ClusterAddonsConfiguration or AddonsConfiguration CRs, and registers them as Service Classes in the Service Catalog.

## Cluster-wide addons flow

1. The Helm Broker watches for ClusterAddonsConfigurations in a given cluster.
2. The user creates the ClusterAddonsConfiguration custom resource.
3. The Helm Broker fetches and parses the data of all addon repositories defined in the ClusterAddonsConfiguration.
4. The Helm Broker creates a ClusterServiceBroker. There is always only one ClusterServiceBroker, even if there are more ClusterAddonsConfigurations.
5. The Service Catalog fetches services that the ClusterServiceBroker exposes.
6. The Service Catalog creates a ClusterServiceClass for each service received from the ClusterServiceBroker.

![Helm Broker cluster](./assets/hb-cluster.svg)

## Namespace-scoped addons flow

1. The Helm Broker watches for AddonsConfigurations in all Namespaces.
2. The user creates the AddonsConfiguration custom resource in a given Namespace.
3. The Helm Broker fetches and parses the data of all addon repositories defined in the AddonsConfiguration.
4. The Helm Broker creates a Service Broker (SB) inside the Namespace in which the AddonsConfiguration is created. There is always only one ServiceBroker per Namespace, even if there are more AddonsConfigurations.
5. The Service Catalog fetches services that the Service Broker exposes.
6. The Service Catalog creates a ServiceClass for each service received from the Service Broker.

![Helm Broker cluster](./assets/hb-namespaced.svg)

## Provisioning and binding

After you register your addons in the Service Catalog, you can provision and bind Service Classes that your addons provide.

1. Select a given addon Service Class from the Service Catalog.
2. Provision this Service Class by creating its ServiceInstance in a given Namespace.
3. Bind your ServiceInstance to a service or lambda.
4. The service or lambda calls a given addon.

![Helm Broker architecture](./assets/hb-architecture.svg)
