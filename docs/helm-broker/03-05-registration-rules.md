---
title: Registration rules
type: Details
---

This document presents the rules you must follow when you configure the Helm Broker. It also describes the Helm Broker's behavior in case of conflicts.

## Using HTTP URLs

On your non-local clusters, you can use only servers with TLS enabled. All incorrect or unsecured URLs will be omitted. Find the information about the rejected URLs in the Helm Broker logs. You can use unsecured URLs only on your local cluster. To use URLs without TLS enabled, set the **global.isDevelopMode** environment variable in the [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/helm-broker/values.yaml) file to `true`.

## Registering the same ID multiple times

>**NOTE:** This section does not cover global problems with conflicting IDs between ClusterServiceClasses or ServiceClasses. There can still be a situation where few different brokers register ClusterServiceClasses or ServiceClasses with the same ID. In such a case, those classes are visible in the Service Catalog but provisioning action is blocked with the `Found more that one class` message.

* When both AddonsConfiguration and ClusterAddonsConfiguration register addons with the same ID, then both are visible in the Service Catalog and they do not have a conflict with each other.
* When there is more than one addon with the same ID specified under the **repositories** parameter in one ClusterAddonsConfiguration, the whole CR is marked as failed. The **status** field of the CR contains information about the conflicted addons. The same rule applies to AddonsConfiguration CRs.
* When many ClusterAddonsConfigurations register addons with the same IDs, only the first CR is registered and all others are marked as failed. In the **status** entry, you can find information about the conflicted addons. The same rule applies to AddonsConfiguration CRs.  
