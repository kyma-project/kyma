---
title: Enable Compass in Kyma
type: Installation
---

To enable Compass in Kyma, follow the instructions for the [custom component installation](/root/kyma#configuration-custom-component-installation) and enable the `compass` and `compass-runtime-agent` modules.

You can also [install Kyma](/root/kyma#installation-install-kyma-on-a-cluster) with these ready-to-use configurations for different Compass modes:

| Configuration | Description |
|----------------|------|
| [`installer-cr-cluster-with-compass.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-with-compass.yaml.tpl) | Provides complete cluster Kyma installation with both Compass and Agent. These components will eventually become part of the default Kyma installation.  |
| [`installer-cr-cluster-compass.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-compass.yaml.tpl) | Provides cluster Kyma installation with Compass. This configuration includes only the selected Kyma components that Compass uses. |
| [`installer-cr-cluster-runtime.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-runtime.yaml.tpl) | Provides complete cluster Kyma installation with Agent only. |
