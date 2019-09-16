---
title: Enable Compass in Kyma
type: Installation
---

To enable Compass in Kyma, follow the instructions for the [custom component installation](/root/kyma#configuration-custom-component-installation) and enable the `compass` and `compass-runtime-agent` modules.
You can also [install Kyma on a cluster](/root/kyma#installation-install-kyma-on-a-cluster) with the ready-to-use configurations for different Compass modes.

## Default Kyma installation

This is a preconfigured single-tenant and single-Runtime mode which will eventually become part of the default Kyma installation. It provides the complete cluster Kyma installation with all components, including both Compass and Agent. To enable this mode, follow the cluster Kyma installation and use the [`installer-cr-cluster-with-compass.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-with-compass.yaml.tpl) configuration file.

## Kyma with Compass only

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
  # The parameter that enables the Compass gateway, as the default Kyma gateway is disabled in this installation mode.
  gateway.gateway.enabled: "true"
  # The name of the currently used gateway
  global.istio.gateway.name: compass-istio-gateway
  # The Namespace of the currently used gateway
  global.istio.gateway.namespace: compass-system
  # The parameter that disables preconfiguration for the Compass Agent
  global.agentPreconfiguration: "false"
  # The Namespace with a Secret that contains a certificate for the Connector Service
  global.connector.secrets.ca.namespace: compass-system
```

## Kyma as a Runtime

This is a single-tenant mode that provides complete cluster Kyma installation with Agent only. To enable this mode, follow the cluster Kyma installation and use the  [`installer-cr-cluster-runtime.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-runtime.yaml.tpl) configuration file.
