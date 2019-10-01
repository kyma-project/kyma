---
title: Overview
---

>**CAUTION:** This implementation replaces the API Gateway that is based on the Api custom resource. The services you exposed and secured using the deprecated implementation require no action, as the API Gateway Controller co-exists with the API Controller in the cluster. Expose and secure new services and functions secured with OAuth2 using the `v2` implementation described in this documentation topic.

To make your service accessible outside the Kyma cluster, expose it using the Kyma API Gateway Controller, which listens for the custom resource (CR) objects that follow the `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD). Creating a valid CR triggers the API Gateway Controller to create an Istio Virtual Service. Optionally, you can specify the **rules** attribute of the CR to secure the exposed service with Oathkeeper Access Rules.

The API Gateway Controller allows you to secure the exposed services using JWT tokens issued by Kyma Dex, or OAuth2 tokens issued by the Kyma OAuth2 server. You can secure the entire service, or secure the selected endpoints. Alternatively, you can leave the service unsecured.

>**NOTE:** Read [this](/components/security/#details-o-auth2-and-open-id-connect-server) document to learn more about the Kyma OAuth2 server.
