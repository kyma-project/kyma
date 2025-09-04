# Expose Workloads Using oauth2-proxy

Learn how to configure [oauth2-proxy](https://github.com/oauth2-proxy/manifests/tree/main/helm/oauth2-proxy) external authorization provider in the Istio custom resource (CR).

## Prerequisites
* You have the Istio module added. If you want to use a Kyma domain to expose a workload, also the API Gateway module must be added.
* You have [set up your custom domain](https://kyma-project.io/#/api-gateway/user/tutorials/01-10-setup-custom-domain-for-workload). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.
  
  > [!NOTE]
  > Because the default Kyma domain is a wildcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment.

  >[!TIP]
  > To get the default domain of your Kyma cluster, run the following command:
  >```yaml
  >kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'
  >```

* You have installed [Helm](https://helm.sh/docs/intro/install/).

## Context
The Istio CR allows configuring external authorization providers that operate over HTTP. To set up the oauth2-proxy external authorization provider in the Istio CR and expose your workload using an Istio VirtualService, follow these steps.

## Procedure
1. Export the name of the namespace in which you want to deploy the HTTPBin Service:
    ```bash
    export NAMESPACE={NAMESPACE_NAME}
    ```  
2. Create a sample HTTPBin workload:
    ```bash
    kubectl create ns $NAMESPACE
    kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
    kubectl create -n $NAMESPACE -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
    ```
3. Export the name of your domain and Gateway:
   ```bash
    export DOMAIN_NAME={YOUR_DOMAIN_NAME}
    export GATEWAY={YOUR_GATEWAY}
    ```   
4. Create a VirtualService to expose the workload:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: ext-authz
      namespace: $NAMESPACE
    spec:
      hosts:
      - "httpbin.$DOMAIN_NAME"
      gateways:
      - $GATEWAY
      http:
      - match:
        - uri:
            prefix: /
        route:
        - destination:
            port:
              number: 80
            host: httpbin.$NAMESPACE.svc.cluster.local
    EOF
    ```

5. Export the following configuration values as environment variables:

    ```bash
    export CLIENT_ID={YOUR_CLIENT_ID}
    export CLIENT_SECRET={YOUR_CLIENT_SECRET}
    export COOKIE_SECRET={YOUR_COOKIE_SECRET}
    export OIDC_ISSUER_URL={YOUR_OIDC_ISSUER_URL}
    ```

    Option | Description |
    ---------|----------|
    CLIENT_ID | 	The unique identifier for the client application that is registered with the external authorizer. |
    CLIENT_SECRET | A secret key known only to the client and the external authorizer. It is used to authenticate the client when communicating with the external authorizer. To generate this value, you can use the command `openssl rand -base64 32 \| head -c 32 \| base64` |
    CLIENT_COOKIE | A secret key used to sign and encrypt the cookies that are used for session management and user authentication. To generate this value, you can use the command `openssl rand -base64 32 \| head -c 32 \| base64` |
    OIDC_ISSUER_URL | This is the URL of the OpenID Connect (OIDC) issuer. Typically, you can find the issuer at `https://{YOUR_IDENTITY_PROVIDER_INSTANCE}/.well-known/openid-configuration`. |

6. Create a `values.yaml` file with the oauth2-proxy configuration for your authorization server:

    >[!TIP]
    > You can adapt this configuration to better suit your needs. See the [additional configuration parameters](https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview/#config-options).

    ```bash
    cat <<EOF > values.yaml
    config:
      clientID: $CLIENT_ID
      clientSecret: $CLIENT_SECRET
      cookieSecret: $COOKIE_SECRET
    extraArgs:
      provider: oidc
      cookie-secure: false
      cookie-domain: "$DOMAIN_TO_EXPOSE_WORKLOADS"
      cookie-samesite: lax
      set-xauthrequest: true
      whitelist-domain: "*.$DOMAIN_TO_EXPOSE_WORKLOADS:*"
      reverse-proxy: true
      pass-access-token: true
      set-authorization-header: true
      pass-authorization-header: true
      scope: "openid email"
      upstream: static://200
      skip-provider-button: true
      redirect-url: "https://oauth2-proxy.$DOMAIN_TO_EXPOSE_WORKLOADS/oauth2/callback"
      oidc-issuer-url: https://$OIDC_ISSUER_URL
      code-challenge-method: S256
    EOF
    ```

7. To install oauth2-proxy with the defined configuration, use [oauth2-proxy helm chart](https://github.com/oauth2-proxy/manifests):

    ```bash
    helm repo add oauth2-proxy https://oauth2-proxy.github.io/manifests
    helm install custom oauth2-proxy/oauth2-proxy -f values.yaml
    ```

8. Register oauth2-proxy as an authorization provider in the Istio module:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: operator.kyma-project.io/v1alpha1
    kind: Istio
    metadata:
      name: default
      namespace: kyma-system
    spec:
      config:
        numTrustedProxies: 1
        authorizers:
        - name: "oauth2-proxy"
          service: "custom-oauth2-proxy.$NAMESPACE.svc.cluster.local"
          port: 80
          headers:
            inCheck:
              include: ["authorization", "cookie"]
              add:
                x-auth-request-redirect: "https://%REQ(:authority)%%REQ(x-envoy-original-path?:path)%"
            toUpstream:
              onAllow: ["authorization", "path", "x-auth-request-user", "x-auth-request-email", "x-auth-request-access-token"]
            toDownstream:
              onAllow: ["set-cookie"]
              onDeny: ["content-type", "set-cookie"]
    EOF
    ```
9.  Create an AuthorizationPolicy CR with the CUSTOM action and the oauth2-proxy provider:
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: security.istio.io/v1
    kind: AuthorizationPolicy
    metadata:
        name: ext-authz
        namespace: test
    spec:
      action: CUSTOM
      provider:
        name: oauth2-proxy
      rules:
      - to:
        - operation:
            paths:
            - /headers
      selector:
        matchLabels:
          app: httpbin
    EOF
    ```

10. Create a `DestinationRule` resource with a traffic policy for the external authorization provider:
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1
    kind: DestinationRule
    metadata:
      name: external-authz-https
      namespace: istio-system
    spec:
      host: $OIDC_ISSUER_URL
      trafficPolicy:
        tls:
          mode: SIMPLE
    EOF
    ```


### Results
When you access the URL of the exposed service `https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers`, you are redirected to the authorization provider's page.
