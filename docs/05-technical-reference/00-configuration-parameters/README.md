---
title: Configuration parameters
---

Kyma uses Helm's chart templating mechanism with `values.yaml` files that contain the configuration parameters.

A global chart can override the subcharts that complement it. Learn more about the syntax and semantics of [global charts and subcharts in Helm](https://helm.sh/docs/chart_template_guide/subcharts_and_globals/).

If you want to [change the Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md), simply redefine them in your custom `values.yaml` file and supply them as input to the `kyma deploy` command, similar to Helm commands.
