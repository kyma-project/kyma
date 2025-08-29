# Expose and Secure a Workload with Istio

This tutorial shows how to expose and secure a workload using Istio's built-in security features. You will expose the workload by creating a [VirtualService](https://istio.io/latest/docs/reference/config/networking/virtual-service/). Then, you will secure access to your workload by adding the JWT validation verified by the Istio security configuration with [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) and [Request Authentication](https://istio.io/latest/docs/reference/config/security/request_authentication/).

## Prerequisites

* [Deploy a sample HTTPBin Service](../../01-00-create-workload.md).
* [Obtain a JSON Web Token (JWT)](../01-51-get-jwt.md).
* [Set Up Your Custom Domain](../../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.
  
  > [!NOTE]
  > Because the default Kyma domain is a wildcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment.

  > [!TIP]
  > To learn what the default domain of your Kyma cluster is, run `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'`.

## Steps

### Expose Your Workload

<!-- tabs:start -->
  #### **Kyma dashboard**

  1. Go to **Istio > Virtual Services** and select **Create**.
  2. Provide the following configuration details:
      - **Name**: `httpbin`
      - Go to **HTTP > Matches > Match** and provide URI of the type **prefix** and value `/`.
      - Go to **HTTP > Routes > Route > Destination**. Replace `{NAMESPACE}` with the name of the HTTPBin Service's namespace and add the following fields:
        - **Host**: `httpbin.{NAMESPACE}.svc.cluster.local`
        - **Port Number**: `8000`
  3. To create the VirtualService, select **Create**.

  #### **kubectl**

  1. Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:

      <!-- tabs:start -->
      #### **Custom Domain**

      ```bash
      export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
      export GATEWAY=$NAMESPACE/httpbin-gateway
      ```
      #### **Kyma Domain**

      ```bash
      export DOMAIN_TO_EXPOSE_WORKLOADS={KYMA_DOMAIN_NAME}
      export GATEWAY=kyma-system/kyma-gateway
      ```
      <!-- tabs:end -->

  2. To expose your workload, create a VirtualService:

      ```bash
      cat <<EOF | kubectl apply -f -
      apiVersion: networking.istio.io/v1alpha3
      kind: VirtualService
      metadata:
        name: httpbin
        namespace: $NAMESPACE
      spec:
        hosts:
        - "httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS"
        gateways:
        - $GATEWAY
        http:
        - match:
          - uri:
              prefix: /
          route:
          - destination:
              port:
                number: 8000
              host: httpbin.$NAMESPACE.svc.cluster.local
      EOF
      ```
<!-- tabs:end -->

### Secure Your Workload

To secure the HTTPBin workload using a JWT, create a Request Authentication with Authorization Policy. Workloads with the **matchLabels** parameter specified require a JWT for all requests. Follow the instructions:

<!-- tabs:start -->
  #### **Kyma Dashboard**
  1. Go to **Configuration > Custom Resources > RequestAuthentications**.
  2. Select **Create** and paste the following configuration into the editor:
      ```yaml
      apiVersion: security.istio.io/v1beta1
      kind: RequestAuthentication
      metadata:
        name: jwt-auth-httpbin
        namespace: {NAMESPACE}
      spec:
        selector:
          matchLabels:
            app: httpbin
        jwtRules:
        - issuer: {ISSUER}
          jwksUri: {JWKS_URI}
      ```
  3. Replace the placeholders:
    - `{NAMESPACE}` is the name of the namespace in which you deployed the HTTPBin Service.
    - `{ISSUER}` is the issuer of your JWT.
    - `{JWKS_URI}` is your JSON Web Key Set URI.
  4. Select **Create**.
  5. Go to **Istio > Authorization Policies**.
  6. Select **Create**, switch to the `YAML` tab and paste the following configuration into the editor:
      ```yaml
      apiVersion: security.istio.io/v1beta1
        kind: AuthorizationPolicy
        metadata:
          name: httpbin
          namespace: {NAMESPACE}
        spec:
          selector:
            matchLabels:
              app: httpbin
          rules:
          - from:
            - source:
                requestPrincipals: ["*"]
      EOF
      ```
  7. Replace `{NAMESPACE}` with the name of the namespace in which you deployed the HTTPBin Service.
  8. Select **Create**.

  #### **kubectl**

  Create the Request Authentication and Authorization Policy resources:

  ```bash
  cat <<EOF | kubectl apply -f -
  apiVersion: security.istio.io/v1beta1
  kind: RequestAuthentication
  metadata:
    name: jwt-auth-httpbin
    namespace: $NAMESPACE
  spec:
    selector:
      matchLabels:
        app: httpbin
    jwtRules:
    - issuer: $ISSUER
      jwksUri: $JWKS_URI
  ---
  apiVersion: security.istio.io/v1beta1
  kind: AuthorizationPolicy
  metadata:
    name: httpbin
    namespace: $NAMESPACE
  spec:
    selector:
      matchLabels:
        app: httpbin
    rules:
    - from:
      - source:
          requestPrincipals: ["*"]
  EOF
  ```
<!-- tabs:end -->
### Access the Secured Resources

To access your HTTPBin Service, use [curl](https://curl.se).

1. To call the endpoint, send a `GET` request to the HTTPBin Service.

    ```bash
    curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200
    ```
    You get the code `401 Unauthorized` error.

2. Now, access the secured workload using the correct JWT.

    ```bash
    curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/status/200 --header "Authorization:Bearer $ACCESS_TOKEN"
    ```
    You get the `200 OK` response code.