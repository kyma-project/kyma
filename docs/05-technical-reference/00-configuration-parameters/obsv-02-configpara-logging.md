---
title: Logging charts
---

You find the configurable parameters for Logging in the [Logging values.yaml](https://github.com/kyma-project/kyma/blob/main/resources/logging/values.yaml) file and in the following subcharts:

- [Fluent Bit](https://github.com/kyma-project/kyma/blob/main/resources/logging/charts/fluent-bit/values.yaml)
- [Log UI](https://github.com/kyma-project/kyma/blob/main/resources/logging/charts/logui/values.yaml)
- [Loki](https://github.com/kyma-project/kyma/blob/main/resources/logging/charts/loki/values.yaml)

The `values.yaml` files are fully documented and provide details on the configurable parameters and their customization.

To override the configuration, modify the default values of the `values.yaml` file and [deploy them](../../04-operation-guides/operations/03-change-kyma-config-values.md).
