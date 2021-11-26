---
title: Migration Guide 1.25-2.0
---

Once you upgrade to Kyma 2.0, perform the manual steps described in the Migration Guide.

## Security

### Native Kubernetes authentication in Kyma

Kyma 2.0 simplifies things and switches to the native Kubernetes authentication. As a result, we no longer support the following authentication and authorization components:

- API Server Proxy
- Console Backend Service
- Dex
- IAM Kubeconfig Service
- Permission Controller
- UAA Activator

To use the native Kubernetes authentication in Kyma, you need to remove the deprecated components manually.

After the successful upgrade to Kyma 2.0, run the following [script](.assets/1.25-2.0-remove-deprecated-resources.sh), which uninstalls and deletes the unsupported items.

### ORY Oathkeeper without Dex

With Kyma 2.0, the Dex component becomes deprecated. Existing API Rules that have a JWT access strategy defined must be enriched with an individual **jwks_url** pointing to a custom OpenID Connect-compliant identity provider. Follow these steps to migrate your API Rule custom resources (CRs):

1. List all the API Rule CRs having a JWT access strategy defined. Run:

   ```bash
   kubectl get apirules -A -o=json | jq '.items[]|select(any( .spec.rules[].accessStrategies[]; .handler=="jwt"))|.metadata'
   ```

2. Go through the list and in each of the API Rule CRs, change the value of the **jwks_url** parameter from `http://dex-service.kyma-system.svc.cluster.local:5556/keys` to the relevant URL of your custom identity provider. Run:

   ```bash
   kubectl edit {RESOURCE} n-{NAMESPACE}
   ```

## Observability

### Grafana, Kiali, and Jaeger without Dex

With the 2.0 release, Kyma does not expose Grafana, Kiali, and Jaeger UIs by default. After the upgrade, for each servic, you can set up an identity provider of your choice. Do it securely and follow the steps outlined in the [Access and Expose Kiali, Grafana, and Jaeger](https://kyma-project.io/docs/kyma/latest/04-operation-guides/security/sec-06-access-expose-kiali-grafana) document.
