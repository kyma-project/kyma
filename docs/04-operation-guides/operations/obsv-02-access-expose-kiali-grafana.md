---
title: Access and Expose Kiali, Grafana, and Jaeger
---

By default, Kyma does not expose Kiali, Grafana, and Jaeger. However, you can still access them using port forwarding. If you want to expose Kiali, Grafana, and Jaeger securely, use an identity provider of your choice.

## Prerequisites

- You have defined the kubeconfig file for your cluster as default (see [Kubernetes: Organizing Cluster Access Using kubeconfig Files](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)).
- To expose the services securely with OAuth, you must have a registered OAuth application with one of the [supported providers](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/oauth_provider#github-auth-provider).

## Access Kiali, Grafana, and Jaeger

### Steps

1. To forward a local port to a port on the service's Pod, run the following command:

<div tabs>
  <details>
  <summary>
  Kiali
  </summary>

  ```bash
  kubectl -n kyma-system port-forward svc/kiali-server 20001:20001
  ```

  </details>
  <details>
  <summary>
  Grafana
  </summary>

  ```bash
  kubectl -n kyma-system port-forward svc/monitoring-grafana 3000:80
  ```

  </details>
  <details>
  <summary>
  Jaeger
  </summary>

  ```bash
  kubectl -n kyma-system port-forward svc/tracing-jaeger-query 16686:16686
  ```

  </details>
</div>

>**NOTE:** kubectl port-forward does not return. You will have to cancel it with Ctrl+C if you want to stop port forwarding.

2. To access the respective service's UI, open `http://localhost:20001` (for Kiali), `http://localhost:3000` (for Grafana), or `http://localhost:16686` (for Jaeger) in your browser.

## Expose Kiali, Grafana, and Jaeger securely

Kyma manages an [OAuth2 Proxy](https://oauth2-proxy.github.io/oauth2-proxy/) instance to secure access to Kiali, Grafana and Jaeger. To make the services accessible, configure OAuth2 Proxy by creating a Kubernetes `Secret` with your identity provider credentials.

### Steps

The following example shows how to use an OpenID Connect (OIDC) compliant identity provider for Kiali, Grafana and Jaeger.

>**NOTE:** The OAuth2 Proxy supports a wide range of other well-known authentication services or OpenID Connect for custom solutions. To find instructions for other authentication services, see the [list of supported providers](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/oauth_provider).

1. Create a new OpenID Connect application for your identity provider. Set the callback URL to the `/oauth2/callback` path of your service, for example, `https://kiali.kyma.example.com/oauth2/callback`. Your identity provider will return a client ID, a client secret, and a token issuer URL.

2. Create a `Secret` for the OAuth2 Proxy configuration [environment variables](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/overview/#environment-variables).

   - For an OpenID Connect compliant provider, adapt the client ID, secret and token issuer to the values that were provided while creating the application.

   - To limit access to specific user groups, configure this with the `OAUTH2_PROXY_ALLOWED_GROUPS` variable and ensure that `OAUTH2_PROXY_OIDC_GROUPS_CLAIM` points to the groups attribute name that is used by your authentication service (`groups` is the default). To get the configuration flags required for other identity provider types, see [OAuth2 Proxy docs](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/oauth_provider/).

<div tabs>
  <details>
  <summary>
  Kiali
  </summary>

  ```bash
  kubectl -n kyma-system create secret generic kiali-auth-proxy-user \
    --from-literal="OAUTH2_PROXY_CLIENT_ID=<my-client-id>" \
    --from-literal="OAUTH2_PROXY_CLIENT_SECRET=<my-client-secret>" \
    --from-literal="OAUTH2_PROXY_OIDC_ISSUER_URL=<my-token-issuer>" \
    --from-literal="OAUTH2_PROXY_PROVIDER=oidc" \
    --from-literal="OAUTH2_PROXY_SCOPE=openid" \
    --from-literal="OAUTH2_PROXY_ALLOWED_GROUPS=<my-groups>" \
    --from-literal="OAUTH2_PROXY_SKIP_PROVIDER_BUTTON=true"
  ```

  </details>
  <details>
  <summary>
  Grafana
  </summary>

  ```bash
  kubectl -n kyma-system create secret generic monitoring-auth-proxy-grafana-user \
    --from-literal="OAUTH2_PROXY_CLIENT_ID=<my-client-id>" \
    --from-literal="OAUTH2_PROXY_CLIENT_SECRET=<my-client-secret>" \
    --from-literal="OAUTH2_PROXY_OIDC_ISSUER_URL=<my-token-issuer>" \
    --from-literal="OAUTH2_PROXY_PROVIDER=oidc" \
    --from-literal="OAUTH2_PROXY_SCOPE=openid" \
    --from-literal="OAUTH2_PROXY_ALLOWED_GROUPS=<my-groups>" \
    --from-literal="OAUTH2_PROXY_SKIP_PROVIDER_BUTTON=true"
  ```

  </details>
  <details>
  <summary>
  Jaeger
  </summary>

  ```bash
  kubectl -n kyma-system create secret generic tracing-auth-proxy-grafana-user \
    --from-literal="OAUTH2_PROXY_CLIENT_ID=<my-client-id>" \
    --from-literal="OAUTH2_PROXY_CLIENT_SECRET=<my-client-secret>" \
    --from-literal="OAUTH2_PROXY_OIDC_ISSUER_URL=<my-token-issuer>" \
    --from-literal="OAUTH2_PROXY_PROVIDER=oidc" \
    --from-literal="OAUTH2_PROXY_SCOPE=openid" \
    --from-literal="OAUTH2_PROXY_ALLOWED_GROUPS=<my-groups>" \
    --from-literal="OAUTH2_PROXY_SKIP_PROVIDER_BUTTON=true"
  ```

  </details>
</div>

>**NOTE:** By default, you are redirected to the documentation. To go to the service's UI instead, disable the OAuth2 Proxy provider button by setting `OAUTH2_PROXY_SKIP_PROVIDER_BUTTON=true`.

3. Restart the OAuth2 Proxy pod:

<div tabs>
  <details>
  <summary>
  Kiali
  </summary>

  ```bash
  kubectl -n kyma-system delete pod -l app=kiali-auth-proxy
  ```

  </details>
  <details>
  <summary>
  Grafana
  </summary>

  ```bash
  kubectl -n kyma-system delete pod -l app.kubernetes.io/name=auth-proxy,app.kubernetes.io/instance=monitoring
  ```

  </details>
  <details>
  <summary>
  Jaeger
  </summary>

  ```bash
  kubectl -n kyma-system delete pod -l app.kubernetes.io/name=auth-proxy,app.kubernetes.io/instance=tracing
  ```

  </details>
</div>
