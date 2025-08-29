# Migrating APIRule `v1beta1` of Type **jwt** to Version `v2`

Learn how to migrate an APIRule created in version `v1beta1` using the **jwt** handler to version `v2`.

## Context

APIRule in version `v1beta1` is deprecated and scheduled for removal. Once the APIRule custom resource definition (CRD) stops serving version `v1beta1`, the API server will no longer respond to requests for APIRules in this version. As a result, you will encounter errors when attempting to access the APIRule custom resource using the deprecated `v1beta1` version. Therefore, you must migrate to version `v2`.

## Prerequisites

* You have read [Changes Introduced in APIRule `v2`](../custom-resources/apirule/04-70-changes-in-apirule-v2.md), which details the updates implemented in the new version of APIRule. If any of these changes affect your setup, you must consider them when migrating to APIRule `v2` and make the necessary adjustments.
* You have the Istio and API Gateway modules added.
* You have installed [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) and [curl](https://curl.se/).
* You have a deployed workload exposed by an APIRule in the deprecated `v1beta1` version. The APIRule uses the **jwt** handler.
  > [!NOTE] 
  > The workload exposed by the APIRule in version `v2` must be a part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection?id=enable-istio-sidecar-proxy-injection).

## Steps

This example demonstrates a migration from an APIRule `v1beta1` with the **jwt** handler to an APIRule `v2` with the **jwt** handler.
The example uses an HTTPBin service, exposing the `/anything` and `/.*` endpoints. The HTTPBin service is deployed in its own namespace, with Istio enabled, ensuring the workload is part of the Istio service mesh.

1. Retrieve a configuration of the APIRule in version `v1beta1` and save it for further modifications. For instructions, see [Retrieve the Complete **spec** of an APIRule in Version `v1beta1`](./01-81-retrieve-v1beta1-spec.md). See a sample of the retrieved **spec** in the YAML format:
    The following configuration uses the **jwt** handler to expose the HTTPBin service's `/anything` and `/.*` endpoints.
    
    ```yaml
    host: httpbin.local.kyma.dev
    service:
      name: httpbin
      namespace: test
      port: 8000
    gateway: kyma-system/kyma-gateway
    rules:
      - path: /anything
        methods:
          - POST
        accessStrategies:
          - handler: jwt
            config:
              jwks_urls:
                -  https://{IAS_TENANT}.accounts.ondemand.com/oauth2/certs
      - path: /.*
        methods:
          - GET
        accessStrategies:
          - handler: jwt
            config:
              jwks_urls:
                -  https://{IAS_TENANT}.accounts.ondemand.com/oauth2/certs
    ```

2. Adjust the retrieved configuration to align with the **jwt** configuration of APIRule `v2`. 

    To ensure the APIRule specification is compatible with version `v2`, you must include a mandatory field named **issuer** in the **jwt** handler's configuration.
    You can find the **issuer** URL in the OIDC well-known configuration of your tenant, located at `https://{YOUR_TENANT}.accounts.ondemand.com/.well-known/openid-configuration`. Additionally, note that the value of the `jwks_urls` field is now stored in the `jwksUri` field.  Tokens do not need to be reissued unless they have expired. 
    See an example of the adjusted APIRule configuration for version `v2`:

    ```yaml
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: httpbin
      namespace: test
    spec:
      hosts:
        - httpbin
      service:
        name: httpbin
        namespace: test
        port: 8000
      gateway: kyma-system/kyma-gateway
      rules:
        - jwt:
            authentications:
              - issuer: https://{YOUR_TENANT}.accounts.ondemand.com
                jwksUri: https://{YOUR_TENANT}.accounts.ondemand.com/oauth2/certs
          methods:
            - POST
          path: /anything
        - jwt:
            authentications:
              - issuer: https://{YOUR_TENANT}.accounts.ondemand.com
                jwksUri: https://{YOUR_TENANT}.accounts.ondemand.com/oauth2/certs
          methods:
            - GET
          path: /{**}
    ```
    > [!NOTE]
    > The **hosts** field accepts a short host name (without a domain). Additionally, the path `/.*` has been changed to `/{**}` because APIRule `v2` does not support regular expressions in the **spec.rules.path** field.
    >
    > For more information, see [Changes Introduced in APIRule `v2`](../custom-resources/apirule/04-70-changes-in-apirule-v2.md).

3. Update the APIRule to version `v2` by applying the adjusted configuration. 

   To verify the version of the applied APIRule, check the value of the `gateway.kyma-project.io/original-version` annotation in the APIRule spec. A value of `v2` indicates that the APIRule has been successfully migrated. To see the value, run:
    ```bash 
    kubectl get apirules.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
    ```
    The following output indicates that the APIRule has been successfully migrated to version `v2`:
    ```yaml
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      annotations:
        gateway.kyma-project.io/original-version: v2
    ...
    ```
    > [!WARNING] Do not manually change the `gateway.kyma-project.io/original-version` annotation. This annotation is automatically updated when you apply your APIRule in version `v2`. Modifying the annotation's value manually causes your APIRule v1beta1 to be handled and configured as version v2, potentially leading to reconciliation errors.

4. To preserve the internal traffic policy from the APIRule `v1beta1`, you must apply the following AuthorizationPolicy. 

    In APIRule `v2`, internal traffic is blocked by default. Without this AuthorizationPolicy, attempts to connect internally to the workload cause an `RBAC: access denied` error. Ensure that the selector label is updated to match the target workload.

    | Option  | Description  |
    |---|---|
    |**{NAMESPACE}**  | The namespace to which the AuthorizationPolicy applies. This namespace must include the target workload for which you allow internal traffic. The selector matches workloads in the same namespace as the AuthorizationPolicy. |
    |**{LABEL_KEY}**: **{LABEL_VALUE}** | To further restrict the scope of the AuthorizationPolicy, specify label selectors that match the target workload. Replace these placeholders with the actual key and value of the label. The label indicates a specific set of Pods to which a policy should be applied. The scope of the label search is restricted to the configuration namespace in which the AuthorizationPolicy is present. <br>For more information, see [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/).|
    ```yaml
    apiVersion: security.istio.io/v1
    kind: AuthorizationPolicy
    metadata:
      name: allow-internal
      namespace: {NAMESPACE}
    spec:
      selector:
        matchLabels:
          {LABEL_KEY}: {LABEL_VALUE} 
      action: ALLOW
      rules:
      - from:
        - source:
            notPrincipals: ["cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"]
    ```

5. To retain the CORS configuration from the APIRule `v1beta1`, update the APIRule in version `v2` to include the same CORS settings. 

   For preflight requests to work correctly, you must explicitly add the `"OPTIONS"` method to the **rules.methods** field of your APIRule `v2`. For guidance, see the [APIRule `v2` examples](../custom-resources/apirule/04-10-apirule-custom-resource.md#sample-custom-resource).

### Access Your Workload

- Send a `GET` request to the exposed workload using JWT authentication:

  ```bash
  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/ip --header "Authorization:Bearer $ACCESS_TOKEN"
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the exposed workload using JWT authentication:

  ```bash
  curl -ik -X POST https://{SUBDOMAIN}.{DOMAIN_NAME}/anything -d "test data" --header "Authorization:Bearer $ACCESS_TOKEN"
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the exposed workload without JWT authentication:

  ```bash
  curl -ik -X POST https://{SUBDOMAIN}.{DOMAIN_NAME}/anything -d "test data" 
  ```
  The call returns the `403` error code.