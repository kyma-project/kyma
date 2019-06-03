---
title: Experimental features
type: Details
---

The Service Catalog requires its own instance of api-server and etcd, which increases the complexity of the cluster configuration and maintenance costs.
In case of api-server downtime, all Service Catalog resources are unavailable.
For this reason, Kyma developers contribute to the Service Catalog project to remove the dependency on these external components and replace them
with a native Kubernetes solution - CustomResourceDefinitions (CRDs).

### Enable CRDs

To enable the CRDs feature in the Service Catalog, override the **service-catalog-apiserver.enabled** and **service-catalog-crds.enabled** parameters
in the installation file:
- For the local installation, modify the `service-catalog-overrides` ConfigMap in the [installer-config-local.yaml](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-config-local.yaml.tpl#L73) file:
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
- For the cluster installation, add the `service-catalog-overrides` ConfigMap to the to the cluster before the installation starts. Execute:
    ```
    kubectl create configmap service-catalog-overrides -n kyma-installer --from-literal=service-catalog-apiserver.enabled=false --from-literal=service-catalog-crds.enabled=true \
    && kubectl label configmap service-catalog-overrides -n kyma-installer installer=overrides component=service-catalog
    ```

