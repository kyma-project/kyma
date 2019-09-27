---
title: Overview
---

>**CAUTION:** This implementation of the API Gateway is **deprectaed**. The services you exposed and secured using this implementation require no action, as the API Controller co-exist with the API Gateway Controller in the cluster. Expose and secure new services and functions using the `v2` implementation which uses the [APIRule](/components/api-gateway-v2#custom-resource-api-rule) custom resource. Read [this](/components/api-gateway-v2#overview-overview) documentation to learn more.

To make your service accessible outside the Kyma cluster, expose it using the Kyma API Controller, which listens for the custom resource (CR) objects that follow the `api.gateway.kyma-project.io` CustomResourceDefinition (CRD). Creating a valid CR triggers the API Controller to create an Istio Virtual Service. Optionally, you can specify the **authentication** attribute of the CR to secure the exposed service and create an Istio Authentication Policy for it.
