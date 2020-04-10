---
title: Overview
---

>**CAUTION:** This implementation replaces the API Gateway that is based on the Api custom resource. The services you exposed and secured using the deprecated implementation are automatically converted to use the new implementation. Read [this](/components/api-gateway-v2#details-migration-from-the-previous-api-resources) document to learn how to check if the service was properly migrated.

To make your service accessible outside the Kyma cluster, expose it using the Kyma API Gateway Controller, which listens for the custom resource (CR) objects that follow the `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD). Creating a valid CR triggers the API Gateway Controller to create an Istio Virtual Service. Optionally, you can specify the **rules** attribute of the CR to secure the exposed service with Oathkeeper Access Rules.

The API Gateway Controller allows you to secure the exposed services using JWT tokens issued by Kyma Dex, or OAuth2 tokens issued by the Kyma OAuth2 server. You can secure the entire service, or secure the selected endpoints. Alternatively, you can leave the service unsecured.

>**NOTE:** Read [this](/components/security/#details-o-auth2-and-open-id-connect-server) document to learn more about the Kyma OAuth2 server.
