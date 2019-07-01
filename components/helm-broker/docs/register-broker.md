# Registering the Helm Broker


By default Helm Broker is registered in Kyma as the [ClusterServiceBroker](TBD). Registration is done in automated way by the Helm Broker.

When user creates an ClusterAddonsConfiguration (CAC) custom resource then Helm Broker creates Service and register itself in Service Catalog as ClusterServiceBroker. There is always only one Service and Cluster Service Broker , even if there are more CACs.

When user creates an AddonsConfiguration (AC) custom resource, then Helm broker creates Service and register ifself in ServiceCatalog as ServiceBroker inside the Namespace in which the AC is created. There is always only one Service and Service Broker per Namespace, even if there are more ACs.
