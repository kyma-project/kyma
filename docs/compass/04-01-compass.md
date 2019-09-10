---
title: Enable Compass in Kyma
type: Installation
---

To enable Compass in Kyma, follow the instructions for the [custom component installation](/root/kyma#configuration-custom-component-installation) and enable the `compass` and `compass-runtime-agent` modules.
You can also [install Kyma on a cluster](/root/kyma#installation-install-kyma-on-a-cluster) with the ready-to-use configurations for different Compass modes.

## Default Kyma installation

This is a preconfigured single-tenant and single-Runtime mode which will eventually become part of the default Kyma installation. It provides complete cluster Kyma installation with both Compass and Agent components. To enable this mode, follow the cluster Kyma installation and use the [`installer-cr-cluster-with-compass.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-with-compass.yaml.tpl) configuration file.

## Kyma as Compass

This is a multi-tenant and multi-runtimes mode that provides cluster Kyma installation with Compass only. This configuration includes only the selected Kyma components that Compass uses. To enable this mode, create this ConfigMap and then perform the cluster Kyma installation using the
 [`installer-cr-cluster-compass.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-compass.yaml.tpl) configuration file:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: compass-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: compass
    kyma-project.io/installation: ""
data:
  # Enables Compass gateway
  gateway.gateway.enabled: "true"
  # Disables configuration for Compass Agent
  global.agentPreconfiguration: "false"
  #
  global.connector.secrets.ca.namespace: compass-system
  # This is the name of the actually used gateway
  global.istio.gateway.name: compass-istio-gateway
  # This is the Namespace of the actually used gateway
  global.istio.gateway.namespace: compass-system
```

## Kyma as a Runtime

This is a single-tenant mode that provides complete cluster Kyma installation with Agent only. To enable this mode, follow the cluster Kyma installation and use the  [`installer-cr-cluster-runtime.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-runtime.yaml.tpl) configuration file.
