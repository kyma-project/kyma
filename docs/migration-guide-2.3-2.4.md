---
title: Migration Guide 2.3-2.4
---

In Kyma 2.0, we introduced a [new way to reach registered services and some Application Connectivity improvements](https://kyma-project.io/blog/2021/12/7/release-notes-20/#application-connectivity), which were related to [Service Catalog removal](https://kyma-project.io/blog/2021/12/7/release-notes-20/#service-catalog-deprecation-update).
We then also informed you we'd be removing the support for the old way (based on service instances) soon and encouraged you to switch to the new flow making use of Central Application Gateway.
In this release, we're following up on this promise and switching off the support for the old flow. 

Due to the removal of the deprecated components (such as Application Operator, Application Broker, Rafter, and Service Catalog), certain resources which were previously created in your cluster are now obsolete.

To clean up the obsolete resources, we provide a script that removes them.
This script removes the obsolete resources, but it leaves the deprecated Application Gateways intact. However, due to the removal of Application Operator, new Application Gateways will not be created. For new services to work properly, you must make use of the currently supported flow using Central Application Gateway.

The cleanup is optional, however, if you do decide to run it, **you must perform the cleanup consciously** and only **after you've switched to the new flow**.

> **CAUTION:** If you continue to use service instances in your workloads, for example in Functions, running this script will break your application. 

**After** you upgrade from Kyma 2.3 to 2.4, either run the script [`2.3-2.4-cleanup-orphaned-svcat-resources.sh`](assets/2.3-2.4-cleanup-orphaned-svcat-resources.sh) or perform the required steps from that script manually.



> **NOTE:** By default, we do not remove any components from your cluster. If, for some reason, you would like to keep the Service Catalog-related resources, then you can choose to skip this migration guide and just not run the provided script.


