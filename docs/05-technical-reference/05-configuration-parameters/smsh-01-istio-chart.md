---
title: Istio <!--Operator--> chart
---

To configure the Istio <!--Operator--> chart and, override the default values of its [`values.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/istio-operator/values.yaml) file. This document describes parameters that you can configure.

The Istio installation in Kyma uses the [IstioOperator](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) API.
Kyma provides the default IstioOperator configurations for production and evaluation profiles, <!--but you can add a custom IstioOperator definition that overrides the default settings. Nie wystawiamy calej konfiguracji istio operator file’a-->

>**TIP:** See how to [change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md).

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter |  Description | Default value |
|-------|-------|:--------:|
