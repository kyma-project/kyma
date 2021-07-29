---
title: Register your addons in Service Catalog
---

## Register your addons in Service Catalog

After you successfully create your addons, you can register them in Service Catalog. To do so, create a [ClusterAddonsConfiguration](../../05-technical-reference/06-custom-resources/smgt-03-hb-clusteraddonsconfiguration.md) or [AddonsConfiguration](../../05-technical-reference/06-custom-resources/smgt-04-hb-addonsconfiguration.md) CR for cluster-wide or Namespace-scoped addons respectively.

Service Catalog will then fetch the addons you provided in the ClusterAddonsConfiguration or AddonsConfiguration CRs and will create respective ServiceClasses for them. To learn more about the technicalities behind this process, read [Helm Broker basic architecture](../../05-technical-reference/03-architecture/smgt-10-hb.md) and [Helm Broker architecture deep-dive](../../05-technical-reference/03-architecture/smgt-11-hb-deep-dive.md).


## Registration rules

### Addons URLs must be secured with TLS

On your non-local clusters, you can use only servers with TLS enabled. All incorrect or unsecured URLs will be omitted. Find the information about the rejected URLs in Helm Broker logs. To check logs from Helm Broker, run these commands:

```
export HELM_BROKER_POD_NAME=kubectl get pod -n kyma-system -l app=helm-broker
kubectl logs -n kyma-system $HELM_BROKER_POD_NAME helm-broker
```

You can use unsecured URLs only on your local cluster. To use URLs without TLS enabled, set the **global.isDevelopMode** environment variable in the [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/helm-broker/values.yaml) file to `true`.

### Addons IDs of the same type must be unique

Addons provided in one ClusterAddonsConfiguration or AddonsConfiguration CR must have unique IDs. Otherwise:

* When both AddonsConfiguration and ClusterAddonsConfiguration register addons with the same ID, then both are visible in Service Catalog and they do not have a conflict with each other.
* When there is more than one addon with the same ID specified under the **repositories** parameter in one ClusterAddonsConfiguration, the whole CR is marked as failed. The **status** field of the CR contains information about the conflicted addons. The same rule applies to AddonsConfiguration CRs.
* When many ClusterAddonsConfigurations register addons with the same IDs, only the first CR is registered and all others are marked as failed. In the **status** entry, you can find information about the conflicted addons. The same rule applies to AddonsConfiguration CRs.  
