---
title: Enable Kyma with Runtime Agent
---

To enable Kyma with Runtime Agent, follow the [cluster Kyma installation with specific components](02-install-kyma.md#install-specific-components) and add the `compass-runtime-agent` module to the list of components. The default [legacy mode](../../05-technical-reference/00-architecture/ac-01-application-connector-components.md#application-operator) used in Kyma does not support integration with Compass. For that reason, before you start the installation, apply the following ConfigMap which disables components used in the legacy mode, such as Application Registry and Connector Service:

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
