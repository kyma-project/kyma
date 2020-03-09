---
title: Expose the lambda with an API Rule
type: Tutorials
---

This tutorial shows how you can expose a lambda function to access it outside the cluster, through an HTTP proxy. To expose it, use an APIRule custom resource (CR) managed by the in-house API Gateway Controller. This controller reacts to an instance of the APIRule CR and, based on its details, it creates an Istio Virtual Service and Oathkeeper Access Rules that specify your permissions for the exposed function.

When you complete this tutorial, you get a lambda that:

- Is available under an unsecured endpoint (**handler** set to `noop` in the APIRule CR).
- Accepts `GET`, `POST`, `PUT`, and `DELETE` methods.

>**NOTE:** To learn more about securing your lambda, see [this](/components/api-gateway-v2/#tutorials-expose-and-secure-a-service-deploy-expose-and-secure-the-sample-resources) tutorial.

## Prerequisites

This tutorial is based on an existing lambda. To create one, follow the [Create a lambda](#tutorials-create-a-lambda) tutorial.

## Steps

Follows these steps:

1. Export these variables:

    ```bash
    export DOMAIN={DOMAIN_NAME}
    export NAME={LAMBDA_NAME}
    export NAMESPACE=serverless
    ```
    
    >**NOTE:** Lambda takes the name from the Function CR name. The APIRule CR can have a different name but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable.

2. Create an APIRule CR for your lambda. It is exposed on port `80` that is the default port of the [Service Placeholder](#architecture-architecture).

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1alpha1
    kind: APIRule
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      gateway: kyma-gateway.kyma-system.svc.cluster.local
      rules:
      - path: /.*
        accessStrategies:
        - config: {}
          handler: noop
        methods:
        - GET
        - POST
        - PUT
        - DELETE
      service:
        host: $NAME.$DOMAIN
        name: $NAME
        port: 80
    EOF
    ```
    
3. Check if the API Rule was created successfully and has the `OK` status:

    ```bash
    kubectl get apirules $NAME -n $NAMESPACE -o=jsonpath='{.status.APIRuleStatus.code}'
    ```

4. Access the lambda's external address:

    ```bash
    curl https://$NAME.$DOMAIN
    ```
