---
title: ?
type: Details
---

This is end-to-end flow describing how to set the Helm Broker to fetch cluster-wide or Namespace-scoped addon definitions from your remote server.

  1. [Create a repository](#details-addons-repository-structure) with your addons.
  3. [Install Kyma](/root/kyma/#installation-installation) locally or on a cluster.
  4. Create the right CR which contains URLs to your addons repositories:
      * For cluster-wide addons, create the [ClusterAddonsConfiguration](#custom-resource-clusteraddonsconfiguration) CR.
      * For Namespace-scoped addons, create a Namespace where you want to enable the Helm Broker and then create the [AddonsConfiguration](#custom-resource-addonsconfiguration) CR in that Namespace.
  5. The Helm Broker triggers the Service Catalog synchronization automatically. New Service Classes appear after a few seconds.
