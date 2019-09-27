---
title: Overview
---

>**CAUTION:** This implementation replaces the API Gateway that based on the Api custom resource. The services you exposed and secured using the deprecated implementation require no action, as the API Gateway Controller co-exist with the API Controller in the cluster. Expose and secure new services and functions using the `v2` implementation described in this documentation topic.

To make your service accessible outside the Kyma cluster, expose it using the Kyma API Gateway Controller, which listens for the custom resource (CR) objects that follow the `apirule.gateway.kyma-project.io` CustomResourceDefinition (CRD). Creating a valid CR triggers the API Gateway Controller to create an Istio Virtual Service. Optionally, you can specify the **rules** attribute of the CR to secure the exposed service with Oathkeeper Access Rules.
