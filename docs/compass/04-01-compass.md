---
title: Enable Compass in Kyma
type: Installation
---

To enable Compass in Kyma, follow the instructions for the [custom component installation](/root/kyma#configuration-custom-component-installation) and enable the `compass` and `compass-runtime-agent` modules.

You can also [install Kyma](/root/kyma#installation-install-kyma-on-a-cluster) with these ready-to-use configurations for different Compass modes:

| Mode | Configuration file | Description |
|----------------|----------------|------|
|Kyma with both Compass and Agent| [`installer-cr-cluster-with-compass.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-with-compass.yaml.tpl) | Provides complete cluster Kyma installation with both Compass and Agent. These components will eventually become part of the default Kyma installation.  |
|Kyma with Compass only| [`installer-cr-cluster-compass.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-compass.yaml.tpl) | Provides cluster Kyma installation with Compass. This configuration includes only the selected Kyma components that Compass uses. |
|Kyma Runtime with Agent only| [`installer-cr-cluster-runtime.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-runtime.yaml.tpl) | Provides complete cluster Kyma installation with Agent only. |

>**CAUTION:** If you want to install Kyma with the Compass module only, you have to create this ConfigMap before the installation:
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
  gateway.gateway.enabled: "true"
  global.agentPreconfiguration: "false"
  global.connector.secrets.ca.namespace: compass-system
  global.istio.gateway.name: compass-istio-gateway
  global.istio.gateway.namespace: compass-system
```
