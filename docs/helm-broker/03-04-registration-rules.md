---
title: Registration rules
type: Details
---

This document presents the rules you must follow when you configure the Helm Broker. It also describes the Helm Broker's behavior in case of conflicts.

### Using unsecured URLs

On your non-local clusters, you can use only servers with TLS enabled. All incorrect or unsecured URLs will be omitted. Find the information about the rejected URLs in the Helm Broker logs. You can use unsecured URLs only on your local cluster. To use URLs without TLS enabled, set the **global.isDevelopMode** environment variable in the [values.yaml](https://github.com/kyma-project/kyma/blob/master/resources/helm-broker/values.yaml) file to `true`.

### Registering the same ID multiple times

>**NOTE:** This section does not cover global problems with conflicting IDs between ClusterServiceClasses or ServiceClasses. There can be still a situation where few different brokers will register ClusterServiceClasses/ServiceClasses with the same ID. In such case, those classes are visible in Service Catalog but provisioning action is blocked with _"Found more that one class"_ reason.

* When both AddonsConfiguration and ClusterAddonsConfiguration register bundles with the same ID, then both are visible in the Service Catalog and they do not have a conflict with each other.
* When there are more that one bundles with the same ID specified under the **repositories** parameter in the same ClusterAddonsConfiguration, it is marked as failed. In the **status** entry, you can find all information about the conflicted bundles. The same rule applies to AddonsConfiguration.
* When many ClusterAddonsConfigurations register bundles with the same IDs, then only the first CR is registered and all others are marked as failed. In the **status** entry, you can find all information about conflicted bundles. The same role applies to AddonsConfiguration.  
