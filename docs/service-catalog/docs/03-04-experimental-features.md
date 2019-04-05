---
title: Experimental features
type: Details
---

Currently Service Catalog requires its own instance of api-server and etcd. This adds additional complexity to the cluster configuration and increases
maintenance costs. 

### Enable CRDs

To enable the CRDs feature you have to override parameters `service-catalog-apiserver.enabled` and `service-catalog-crds.enabled`
in the installer-config file. Modify or add the `service-catalog-overrides` config map:  
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
