
# Expose and Secure a Workload with ExtAuth

Learn how to expose and secure services using APIGateway Controller and OAuth2.0 Client Credentials flow. For this purpose, this tutorial uses [`oauth2-proxy`](https://oauth2-proxy.github.io/oauth2-proxy/) with an OAuth2.0 complaint authorization server supporting OIDC discovery. APIGateway Controller reacts to an instance of the APIRule custom resource (CR) and creates an Istio [VirtualService](https://istio.io/latest/docs/reference/config/networking/virtual-service/) and [AuthorizationPolicy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) with action type `CUSTOM`.

## Prerequisites

* You have the Istio and API Gateway modules added.
* You have a deployed workload.
  > [!NOTE] 
  > To expose a workload using APIRule in version `v2`, the workload must be a part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection?id=enable-istio-sidecar-proxy-injection).
* To use CLI instructions, you must install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) and [curl](https://curl.se/).
* You have [set up your custom domain](../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.
  
  > [!NOTE]
  > Because the default Kyma domain is a wildcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment.

  > [!TIP]
  > To learn what the default domain of your Kyma cluster is, run `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'`.

* You have a JSON Web Token (JWT). See [Obtain a JWT](./01-51-get-jwt.md).

## Steps

### Expose and Secure Your Workload

1. Replace the placeholders and define the `oauth2-proxy` configuration for your authorization server.
  
    >[!TIP]
    > To generate `COOKIE_SECRET` and `CLIENT_SECRET`, you can use the command `openssl rand -base64 32 | head -c 32 | base64`.
  
    >[!TIP]
    >You can adapt this configuration to better suit your needs. See the [additional configuration parameters](https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview/#config-options).

    ```bash
    cat <<EOF > values.yaml
    config:
      clientID: {CLIENT_ID}
      clientSecret: {CLIENT_SECRET}
      cookieName: ""
      cookieSecret: {COOKIE_SECRET}
    
    extraArgs: 
      auth-logging: true
      cookie-domain: "{DOMAIN_TO_EXPOSE_WORKLOADS}"
      cookie-samesite: lax
      cookie-secure: false
      force-json-errors: true
      login-url: static://401
      oidc-issuer-url: {OIDC_ISSUER_URL}
      pass-access-token: true
      pass-authorization-header: true
      pass-host-header: true 
      pass-user-headers: true
      provider: oidc
      request-logging: true
      reverse-proxy: true
      scope: "{TOKEN_SCOPES}"
      set-authorization-header: true
      set-xauthrequest: true
      skip-jwt-bearer-tokens: true
      skip-oidc-discovery: false
      skip-provider-button: true
      standard-logging: true
      upstream: static://200
      whitelist-domain: "*.{DOMAIN_TO_EXPOSE_WORKLOADS}:*"
    EOF
    ```

2. To install `oauth2-proxy` with your configuration, use [oauth2-proxy helm chart](https://github.com/oauth2-proxy/manifests):

    ```bash
    kubectl create namespace oauth2-proxy
    helm repo add oauth2-proxy https://oauth2-proxy.github.io/manifests
    helm upgrade --install oauth2-proxy oauth2-proxy/oauth2-proxy -f values.yaml -n oauth2-proxy
    ```


3. Register `oauth2-proxy` as an authorization provider in the Istio module:

    ```bash
    kubectl patch istio -n kyma-system default --type merge --patch '{"spec":{"config":{"authorizers":[{"name":"oauth2-proxy","port":80,"service":"oauth2-proxy.oauth2-proxy.svc.cluster.local","headers":{"inCheck":{"include":["x-forwarded-for", "cookie", "authorization"]}}}]}}}'
    ```

4. To expose and secure the Service, create the following APIRule:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2alpha1
    kind: APIRule
    metadata:
      name: {APIRULE_NAME}
      namespace: {APIRULE_NAMESPACE}
    spec:
      hosts: 
        - {SUBDOMAIN}.{DOMAIN_TO_EXPOSE_WORKLOADS}
      service:
        name: {SERVICE_NAME}
        port: {SERVICE_PORT}
      gateway: {GATEWAY_NAME}/{GATEWAY_NAMESPACE}
      rules:
        - extAuth:
            authorizers:
              - oauth2-proxy
          methods:
            - GET
          path: /*
    EOF
    ```

### Access the Secured Resources

To access your HTTPBin Service use [curl](https://curl.se).

1. To call the endpoint, send a `GET` request to the HTTPBin Service.

    ```bash
    curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_TO_EXPOSE_WORKLOADS}/headers
    ```
    You get the error `401 Unauthorized`.

2. Now, access the secured workload using the correct JWT.

    ```bash
    curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_TO_EXPOSE_WORKLOADS}/headers --header "Authorization:Bearer {ACCESS_TOKEN}"
    ```
    You get the `200 OK` response code.