---
title: Fetch addons from HTTPS servers
type: Details
---

By default, the Helm Broker fetches addons listed in the `index.yaml` file from the `addons` repository [release](https://github.com/kyma-project/bundles/releases) compatible with a given Kyma version. Follow these steps to configure the Helm Broker to fetch cluster-wide or Namespace-scoped addon definitions from remote HTTPS servers.

  1. [Create a repository](#details-create-addons-repository) with your addons.
  2. [Install Kyma](/root/kyma/#installation-installation) locally or on a cluster.
  3. Create the proper CR which contains URLs to your addons repositories:
  
    * For cluster-wide addons, create the [ClusterAddonsConfiguration](#custom-resource-clusteraddonsconfiguration) CR.
    * For Namespace-scoped addons, create a Namespace where you want to enable the Helm Broker and then create the [AddonsConfiguration](#custom-resource-addonsconfiguration) CR.
  4. The Helm Broker triggers the Service Catalog synchronization automatically. New Service Classes appear after a few seconds.
