---
title: Enable Kyma with Runtime Agent
---

To enable Kyma with Runtime Agent, follow the cluster Kyma installation using the [`installer-cr-cluster-runtime.yaml.tpl`](https://github.com/kyma-project/kyma/blob/master/installation/resources/installer-cr-cluster-runtime.yaml.tpl) configuration file and enable the `compass-runtime-agent` module. The default [legacy mode](../../05-technical-reference/03-architecture/ac-01-application-connector-components.md#application-operator) used in Kyma does not support integration with Compass. For that reason, before you start the installation, apply the following ConfigMap which disables components used in the legacy mode, such as Application Registry and Connector Service:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: disable-legacy-connectivity-override
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
data:
  global.disableLegacyConnectivity: "true"
EOF
```
