---
title: Overview
type: Overview
---

To make your service accessible outside the Kyma cluster, expose it using the Kyma API Controller, which listens for the Custom Resource (CR) objects that follow the `api.gateway.kyma.cx` Custom Resource Definition (CRD). Creating a valid CR triggers the API Controller to create an Istio Ingress for the service. Optionally, you can specify the **authentication** attribute of the CR to secure the exposed service and create an Istio Authentication Policy for it.
