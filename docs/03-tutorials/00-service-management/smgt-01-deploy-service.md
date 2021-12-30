---
title: Deploy an SAP BTP service in your Kyma cluster
---

This tutorial describes how you can deploy an SAP BTP service in your Kyma cluster using the SAP BTP service operator.

## Prerequisites

- [Kyma cluster](https://kyma-project.io/docs/kyma/latest/04-operation-guides/operations/02-install-kyma/) running on Kubernetes version 1.19 or higher
- SAP BTP [Global Account](https://help.sap.com/products/BTP/65de2977205c403bbc107264b8eccf4b/d61c2819034b48e68145c45c36acba6e.html?locale=en-US) and [Subaccount](https://help.sap.com/products/BTP/65de2977205c403bbc107264b8eccf4b/55d0b6d8b96846b8ae93b85194df0944.html?locale=en-US)
- Service Management Control ([SMCTL](https://help.sap.com/viewer/09cc82baadc542a688176dce601398de/Cloud/en-US/0107f3f8c1954a4e96802f556fc807e3.html)) command line interface
- [kubectl](https://kubernetes.io/docs/tasks/tools/) v1.17 or higher
- [helm](https://helm.sh/) v3.0 or higher

>**CAUTION:** For the BTP service operator to work, you must disable Istio sidecars that are enabled on the Kyma clusters by default. To do so, run these commands:

## Steps

1. Install the [SAP BTP service operator](https://github.com/SAP/sap-btp-service-operator) in your Kyma cluster.
2.
