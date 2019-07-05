---
title: Tutorials
---

By default, the Helm Broker fetches bundles listed in the `index.yaml` file from the `bundles` repository [release](https://github.com/kyma-project/bundles/releases). This tutorial shows how to configure the Helm Broker to fetch cluster-wide and Namespace-scoped bundle definitions from other remote HTTPS servers.

## Steps

Follow these steps to configure the Helm Broker to fetch bundle definitions from other remote HTTPS servers.
<div tabs>
  <details>
  <summary>
  Cluster-wide bundles
  </summary>

  1. [Create a repository](#details-create-a-bundles-repository) with your bundles. To complete this tutorial step by step, use the existing [bundles](https://github.com/kyma-project/bundles/tree/master/bundles) repository.
  2. [Install Kyma](/root/kyma/#installation-installation) locally or on a cluster.
  3. Create the [ClusterAddonsConfiguration](#custom-resource-clusteraddonsconfiguration) CR which contains URLs to your bundles.

  ```yaml
  kubectl create -f https://kyma-project.io/assets/docs/master/helm-broker/docs/assets/cluster-addon.yaml
  ```
  4. The Helm Broker triggers the Service Catalog synchronization automatically. New Service Classes appear after a few seconds.

  </details>
  <details>
  <summary>
  Namespace-scoped bundles
  </summary>

  1. [Create a repository](#details-create-a-bundles-repository) with your bundles. To complete this tutorial step by step, use the existing [bundles](https://github.com/kyma-project/bundles/tree/master/bundles) repository.
  2. [Install Kyma](/root/kyma/#installation-installation) locally or on a cluster.
  3. Create the `hodor` Namespace where you want to enable the Helm Broker:
  ```
  kubectl create namespace hodor
  ```

  4. Create the [AddonsConfiguration](#custom-resource-addonsconfiguration) CR which contains URLs to your bundles:

  ```yaml
  kubectl create -f https://kyma-project.io/assets/docs/master/helm-broker/docs/assets/namespaced-addon.yaml
  ```

  5. The Helm Broker triggers the Service Catalog synchronization automatically. New Service Classes appear after a few seconds.

   </details>
</div>
