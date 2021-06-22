---
title: Overview
---

To make your service accessible outside the Kyma cluster, expose it using the Kyma API Gateway Controller, which listens for the custom resource (CR) objects that follow the `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD). Creating a valid CR triggers the API Gateway Controller to create an Istio Virtual Service. Optionally, you can specify the **rules** attribute of the CR to secure the exposed service with Oathkeeper Access Rules.

The API Gateway Controller allows you to secure the exposed services using JWT tokens issued by OpenID Connect-compliant identity provider, or OAuth2 tokens issued by the Kyma OAuth2 server. You can secure the entire service, or secure the selected endpoints. Alternatively, you can leave the service unsecured.

> **NOTE:** To learn more, read about the [Kyma OAuth2 server](/components/security/#details-o-auth2-and-open-id-connect-server).
