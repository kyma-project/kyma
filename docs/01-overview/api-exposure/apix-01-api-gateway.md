---
title: API Gateway
---

To make your service accessible outside the Kyma cluster, expose it using Kyma API Gateway Controller, which listens for the custom resource (CR) objects that follow the `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD). Creating a valid CR triggers API Gateway Controller to create an Istio VirtualService. Optionally, you can specify the **rules** attribute of the CR to secure the exposed service with Oathkeeper Access Rules.

API Gateway Controller allows you to secure the exposed services using JWT tokens issued by an OpenID Connect-compliant identity provider, or OAuth2 tokens issued by the Kyma OAuth2 server. You can secure the entire service, or secure the selected endpoints. Alternatively, you can leave the service unsecured.

>**CAUTION:** Since Kyma 2.2, Ory stack has been deprecated, and Ory Hydra was removed with Kyma 2.19. For more information, read the blog posts explaining the [new architecture](https://blogs.sap.com/2023/02/10/sap-btp-kyma-runtime-api-gateway-future-architecture-based-on-istio/) and [Ory Hydra migration](https://blogs.sap.com/2023/06/06/sap-btp-kyma-runtime-ory-hydra-oauth2-client-migration/). See the [deprecation note](https://github.com/kyma-project/website/blob/main/content/blog-posts/2022-05-04-release-notes-2.2/index.md#ory-stack-deprecation-note).
