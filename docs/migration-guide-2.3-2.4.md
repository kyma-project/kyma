---
title: Migration Guide 2.3-2.4
---

## Service Catalog cleanup script

In Kyma 2.0, we introduced a [new way to reach registered services and some Application Connectivity improvements](https://kyma-project.io/blog/2021/12/7/release-notes-20/#application-connectivity), which were related to the [Service Catalog removal](https://kyma-project.io/blog/2021/12/7/release-notes-20/#service-catalog-deprecation-update).
We then also informed you we'd be removing the support for the old way soon and encouraged you to switch to the new flow making use of Central Application Gateway. In this release, we're following up on this promise and switching off the support for the old flow.

Due to the removal of the deprecated components (such as Application Operator, Application Broker, Rafter, and Service Catalog), certain resources which were previously created in your cluster are now obsolete.
If you want to delete these obsolete resources, **after** you upgrade from Kyma 2.3 to 2.4, either run the cleanup script [`2.3-2.4-cleanup-orphaned-svcat-resources.sh`](https://github.com/kyma-project/kyma/blob/release-2.4/docs/assets/2.3-2.4-cleanup-orphaned-svcat-resources.sh) or perform the steps from that script manually.

This script removes the obsolete resources, but it leaves the deprecated Application Gateways intact. However, due to the removal of Application Operator, new Application Gateways will not be created. For new services to work properly, you must make use of the currently supported flow using Central Application Gateway.

> **NOTE:** By default, we do not remove any components from your cluster. If, for some reason, you would like to keep the Service Catalog-related resources, then you can choose to skip this migration guide and just not run the provided script.

> **CAUTION:** Note that although we no longer support the old flow, you should run the cleanup consciously **after** you've switched to the new flow. If you continue to use service instances in your workloads, for example in Functions, **running this script will remove these service instances, cause data loss, and break your application**!
 
## Kiali cleanup script

Multiple resources of the Kiali component were renamed due to the changes in the upstream project. You can delete the obsolete resources by running the cleanup script [`2.3-2.4-cleanup-kiali-istio-upgrade.sh`](https://github.com/kyma-project/kyma/blob/release-2.4/docs/assets/2.3-2.4-cleanup-kiali-istio-upgrade.sh).