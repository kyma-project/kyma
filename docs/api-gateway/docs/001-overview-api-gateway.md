---
title: Overview
---

To make your service accessible outside the Kyma cluster, expose it using the Kyma API Controller, which listens for the custom resource (CR) objects that follow the `api.gateway.kyma-project.io` CustomResourceDefinition (CRD). Creating a valid CR triggers the API Controller to create an Istio Virtual Service. Optionally, you can specify the **authentication** attribute of the CR to secure the exposed service and create an Istio Authentication Policy for it.
