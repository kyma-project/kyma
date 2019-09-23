---
title: Enable Compass in Kyma
type: Installation
---

To enable Compass in Kyma, follow the instructions for the [custom component installation](/root/kyma#configuration-custom-component-installation) and enable the `compass` and `compass-runtime-agent` modules. You can also [install Kyma on a cluster](/root/kyma#installation-install-kyma-on-a-cluster) with the ready-to-use configurations for different modes. There are two modes in which you can enable Compass in Kyma: default Kyma installation and Compass as a Central Management Plane.

## Default Kyma installation

This is a preconfigured single-tenant and single-Runtime mode which will eventually become part of the default Kyma installation. It provides the complete cluster Kyma installation with all components, including both Compass and Agent. This mode allows you to register external applications in Kyma. To enable this mode, follow the cluster Kyma installation and use the [`installer-cr-cluster-with-compass.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-with-compass.yaml.tpl) configuration file.

![Kyma mode1](./assets/kyma-mode1.svg)


## Compass as a Central Management Plane

This is a multiple-cluster mode in which you need one cluster with Compass and at least one cluster with Kyma Runtime, which you can connect and manage using Compass. This mode allows you to manage your Runtimes in one central place and integrate them with applications.

![Kyma mode2](./assets/kyma-mode2.svg)


### Kyma with Compass

This is a multi-tenant and multi-Runtime mode that provides cluster Kyma installation with Compass and only the selected Kyma components that Compass uses. To enable this mode, create this ConfigMap and then perform the cluster Kyma installation using the
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
  # The parameter that enables the Compass gateway, as the default Kyma gateway is disabled in this installation mode
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

### Kyma Runtime

This is a single-tenant mode that provides complete cluster Kyma installation with Agent only. To enable this mode, follow the cluster Kyma installation and use the [`installer-cr-cluster-runtime.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-runtime.yaml.tpl) configuration file.
