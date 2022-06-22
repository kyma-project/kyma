---
title: Migration Guide 2.3-2.4
---

In Kyma 2.0, we introduced a [new way to reach registered services and some Application Connectivity improvements](https://kyma-project.io/blog/2021/12/7/release-notes-20/#application-connectivity), which were related to [Service Catalog removal](https://kyma-project.io/blog/2021/12/7/release-notes-20/#service-catalog-deprecation-update).
We then also informed you we'd be removing the support for the old way soon and encouraged you to switch to the new flow making use of Central Application Gateway. In this release, we're following up on this promise and switching off the support for the old flow. 

Due to the removal of the deprecated components (such as Application Operator, Application Broker, Rafter, and Service Catalog), certain resources which were previously created in your cluster are now obsolete and must be deleted if you made use of the old flow using these components.
When you upgrade from Kyma 2.3 to 2.4, either run the script [`2.3-2.4-cleanup-orphaned-svcat-resources.sh`](assets/2.3-2.4-cleanup-orphaned-svcat-resources.sh) or perform the required steps from that script manually.

This script removes the obsolete resources, but it leaves the deprecated Application Gateways intact. However, due to the removal of Application Operator, new Application Gateways will not be created. For new services to work properly, you must make use of the currently supported flow using Central Application Gateway.