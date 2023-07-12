---
title: Enable Kyma with Runtime Agent
---

[Runtime Agent](../../01-overview/application-connectivity/ra-01-runtime-agent-overview.md) is a component that is used in the [Compass mode](../../01-overview/application-connectivity/README.md) of Application Connectivity. 
By default, Kyma uses the [standalone Application Connectivity mode](../../01-overview/application-connectivity/README.md), which does not support integration with Compass.
Because of this, on installation, you must:
- Disable the components used in the standalone mode by setting the **global.disableLegacyConnectivity** value to `true`, and 
- Add the `compass-runtime-agent` module in the `kyma-system` Namespace to the [list of components](https://github.com/kyma-project/kyma/blob/main/installation/resources/components.yaml)

```yaml
kyma deploy --value global.disableLegacyConnectivity="true" --components-file {YOUR_COMPONENTS_FILE_PATH}
```

>**TIP:** Read more about how to [change Kyma settings](03-change-kyma-config-values.md) and [install Kyma with specific components](02-install-kyma.md#install-specific-components).