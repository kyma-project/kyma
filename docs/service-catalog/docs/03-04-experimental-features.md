---
title: Experimental features
type: Details
---

Currently Service Catalog requires its own instance of api-server and etcd. This adds additional complexity to the cluster configuration and increases
maintenance costs. In case of api-server downtime all Service Catalog resources are not available (even simple pod restart can result in such a behaviour).
Kyma developers are contributing to the Service Catalog project to remove dependency to these external components and replace them 
with native K8S solution - Custom Resource Definition.

### Enable CRDs

To enable the CRDs feature you have to override parameters `service-catalog-apiserver.enabled` and `service-catalog-crds.enabled`
in the installer-config file:
- for a local installation modify the `service-catalog-overrides` config map in [installer-config-local.yaml](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-config-local.yaml.tpl#L73):
    ```
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: service-catalog-overrides
      namespace: kyma-installer
      labels:
        installer: overrides
        component: service-catalog
        kyma-project.io/installation: ""
    data:
      etcd-stateful.etcd.resources.limits.memory: 256Mi
      etcd-stateful.replicaCount: "1"
      service-catalog-apiserver.enabled: "false"
      service-catalog-crds.enabled: "true"
    ```
- for a cluster installation add the `service-catalog-overrides` config map to [installer-config-cluster.yaml.tpl](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-config-cluster.yaml.tpl): 
    ```
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: service-catalog-overrides
      namespace: kyma-installer
      labels:
        installer: overrides
        component: service-catalog
        kyma-project.io/installation: ""
    data:
      service-catalog-apiserver.enabled: "false"
      service-catalog-crds.enabled: "true"
    ```
