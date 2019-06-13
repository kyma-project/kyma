---
title: Experimental features
type: Details
---

The Service Catalog requires its own instance of api-server and etcd, which increases the complexity of the cluster configuration and maintenance costs.
In case of api-server downtime, all Service Catalog resources are unavailable.
For this reason, Kyma developers contribute to the Service Catalog project to remove the dependency on these external components and replace them
with a native Kubernetes solution - CustomResourceDefinitions (CRDs).

### Enable CRDs

To enable the CRDs feature in the Service Catalog, override the **global.serviceCatalogApiserver.enabled** and **global.serviceCatalogCrds.enabled** parameters
in the installation file:
- For the local installation, modify the `installation-config-overrides` ConfigMap in the [installer-config-local.yaml](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-config-local.yaml.tpl) file:
    ```
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: installation-config-overrides
      namespace: kyma-installer
      labels:
        installer: overrides
        kyma-project.io/installation: ""
    data:
      global.isLocalEnv: "true"
      global.domainName: "kyma.local"
      global.adminPassword: ""
      nginx-ingress.controller.service.loadBalancerIP: ""
      global.serviceCatalogApiserver.enabled: "true"
      global.serviceCatalogCrds.enabled: "false"
    ```
- For the cluster installation, add the `service-catalog-versions-overrides` ConfigMap to the cluster before the installation starts. Run:
    ```
    kubectl create configmap service-catalog-overrides-versions -n kyma-installer --from-literal=global.serviceCatalogApiserver.enabled=false --from-literal=global.serviceCatalogCrds.enabled=true \
    && kubectl label configmap service-catalog-overrides-versions -n kyma-installer installer=overrides
    ```

