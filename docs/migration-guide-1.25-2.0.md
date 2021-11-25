---
title: Migration Guide 1.25-2.0
---

Once you upgrade to Kyma 2.0, perform the manual steps described in the Migration Guide.

## Security

### Native Kubernetes authentication in Kyma

Kyma 2.0 does not support the following authentication and authorization components:

- API Server Proxy
- Console Backend Service
- Dex
- IAM Kubeconfig Service
- Permission Controller
- UAA Activator

To use the native Kubernetes authentication in Kyma, you need to remove the deprecated components manually. Follow these steps:

[STEPS]

For more details, read the [Kyma 2.0 release notes].

### ORY Oathkeeper without Dex

With Kyma 2.0 the Dex component becomes deprecated. Existing API Rules that have a JWT access strategy defined must be enriched with an individual **jwks_url** pointing to a custom OpenID Connect-compliant identity provider. Follow these step to migrate you API Rules:

[STEPS]
