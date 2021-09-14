---
title: Configuration parameters
---

Kyma uses Helm's chart templating mechanism with `values.yaml` files that contain the configuration parameters.

A global chart can override the subcharts that complement it. Learn more about the syntax and semantic of [global charts and subcharts in Helm](https://helm.sh/docs/chart_template_guide/subcharts_and_globals/).

If you want to [change the Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md), you simply modify the values of the respective chart or subchart, and deploy them.
