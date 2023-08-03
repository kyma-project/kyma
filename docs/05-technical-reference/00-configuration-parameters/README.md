---
title: Configuration parameters
---

Kyma uses Helm's chart templating mechanism with `values.yaml` files that contain the configuration parameters.

A global chart can override the subcharts that complement it. Learn more about the syntax and semantics of [global charts and subcharts in Helm](https://helm.sh/docs/chart_template_guide/subcharts_and_globals/).

If you want to [change the Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md), simply redefine them in your custom `values.yaml` file and supply them as input to the `kyma deploy` command, similar to Helm commands.

The following documents describe charts and configurable parameters in Kyma:

- [Application Connector chart](ac-01-application-connector-chart.md)
- [Ory limitations](apix-01-ory-limitations.md)
- [ORY chart](/05-technical-reference/00-configuration-parameters/apix-02-ory-chart.md)
- [Observability charts](obsv-01-configpara-observability.md)
- [Connection with Compass](ra-01-connection-with-compass.md)
- [Cluster Users chart](sec-01-cluster-users.md)
- [Istio chart](smsh-01-istio-chart.md)
- [Istio Resources chart](smsh-02-istio-resources-chart.md)
- [Serverless chart](svls-01-serverless-chart.md)
- [Environment variables](svls-02-environment-variables.md)
