---
title: Expose and secure a service with JWT
---

This tutorial shows how to expose and secure services or Functions using API Gateway Controller. The controller reacts to an instance of the API Rule custom resource (CR) and creates an Istio Virtual Service and [Oathkeeper Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules) according to the details specified in the CR. To interact with the secured services, the tutorial uses a JWT client registered through the Hydra Maester controller.

It may be a follow-up to the [Use a custom domain to expose a service](./apix-01-own-domain.md) tutorial.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Deploy a service](./apix-02-deploy-service.md) tutorial.

## Expose, secure, and access your resources

1. Expose the service and secure it by creating an API Rule CR in your Namespace. If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-system/kyma-gateway`. Run:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: httpbin
  namespace: mst
spec:
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  rules:
    - accessStrategies:
        - config:
            jwks_urls:
            - https://kymagoattest.accounts400.ondemand.com/oauth2/certs
          handler: jwt
      methods:
        - GET
      path: /.*
  service:
    name: httpbin
    port: 8000
    host: mst.dt-test.goatz.shoot.canary.k8s-hana.ondemand.com
```

>**NOTE:** If you are running Kyma on k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

2. Call the secured service using a JTW access token:

```bash
curl -ik https://mst.dt-test.goatz.shoot.canary.k8s-hana.ondemand.com/headers -H "Authorization: Bearer $ACCESS_TOKEN"
```
