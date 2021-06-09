---
title: Access and Expose Kiali and Grafana
---

By default, Kyma does not expose Kiali and Grafana. However, you can still access them using port forwarding. If you want to expose Kiali and Grafana securely, use an identity provider of your choice.

## Prerequisites

- You have defined the kubeconfig file for your cluster as default (see [Kubernetes: Organizing Cluster Access Using kubeconfig Files](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)).
- To expose the services securely with OAuth, you must have a registered OAuth application with one of the [supported providers](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/oauth_provider#github-auth-provider).

## Access Kiali and Grafana

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

</div>


>**NOTE:** kubectl port-forward does not return. You will have to cancel it with Ctrl+C if you want to stop port forwarding.

2. To access the respective service's UI, open `http://localhost:20001` (for Kiali) or `http://localhost:3000` (for Grafana) in your browser.

## Expose Kiali and Grafana securely

To make Kiali and Grafana permanently accessible, expose the services securely using [OAuth2 Proxy](https://oauth2-proxy.github.io/oauth2-proxy/).

### Steps

The following example shows how to use Github as authentication provider for Kiali and Grafana. You create an `oauth2_proxy` `Deployment` to achieve this, and expose it as a `VirtualService` via Kyma's Istio Gateway.

>**NOTE:** The `oauth2_proxy` supports a wide range of other well-known authentication services or OpenID Connect for custom solutions. See the [list of supported providers](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/oauth_provider) to find instructions for other authentication services.

1. Chose a domain for the exposed service under the Kyma base domain. For example, if your Kyma cluster is reachable under `kyma.example.com`, use `kiali.kyma.example.com` or `grafana.kyma.example.com`, respectively.

2. Create a new Github application under https://github.com/settings/apps. Set the callback URL to `https://kiali.kyma.example.com/oauth2/callback`. Ensure at least read-only access to email addresses for the Github application. Copy the client ID and secret.

3. Create a Kubernetes Secret for the client ID and secret:

<div tabs>
  <details>
  <summary>
  Kiali
  </summary>

  ```bash
  kubectl create secret generic oauth2-kiali-secret -n kyma-system --from-literal="OAUTH2_PROXY_CLIENT_ID=<client-id>" --from-literal="OAUTH2_PROXY_CLIENT_SECRET=<client-secret>" --from-literal="OAUTH2_PROXY_COOKIE_SECRET=``openssl rand -hex 16``"
  ```

  </details>
  <details>
  <summary>
  Grafana
  </summary>

  ```bash
  kubectl create secret generic oauth2-grafana-secret -n kyma-system --from-literal="OAUTH2_PROXY_CLIENT_ID=<client-id>" --from-literal="OAUTH2_PROXY_CLIENT_SECRET=<client-secret>" --from-literal="OAUTH2_PROXY_COOKIE_SECRET=``openssl rand -hex 16``"
  ```

  </details>
</div>

4. Create the `oauth2_proxy` Deployment. Adjust the `args` for the [Github auth provider](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/oauth_provider#github-auth-provider) depending on your own requirements:

<div tabs>
  <details>
  <summary>
  Kiali
  </summary>

  ```yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: oauth2-kiali
    labels:
      app: oauth2-kiali
      target: oauth2-kiali
  spec:
    replicas: 1
    selector:
      matchLabels:
        app: oauth2-kiali
    template:
      metadata:
        labels:
          app: oauth2-kiali
      spec:
        containers:
        - name: oauth2-proxy
          image: quay.io/oauth2-proxy/oauth2-proxy:v7.1.3
          imagePullPolicy: IfNotPresent
          args:
          - --provider=github
          - --email-domain="*"
          - --http-address=0.0.0.0:3000
          - --upstream=http://kiali-server.kyma-system.svc:20001
          - --cookie-name=kiali_oauth2_proxy
          - --proxy-prefix=/oauth2
          - --ping-path=/oauth2/healthy
          - --silence-ping-logging=true
          - --reverse-proxy=true
          - --skip-provider-button=true
          - --cookie-secure
          envFrom:
          - secretRef:
              name: oauth2-kiali-secret
          ports:
          - name: http
            containerPort: 3000
            protocol: TCP
          livenessProbe:
            httpGet:
              path: /oauth2/healthy
              port: http
            initialDelaySeconds: 3
            timeoutSeconds: 2
          readinessProbe:
            httpGet:
              path: /oauth2/healthy
              port: http
            initialDelaySeconds: 3
            timeoutSeconds: 2
        securityContext:
          fsGroup: 65534
          runAsNonRoot: true
          runAsUser: 65534
  ```

  </details>
  <details>
  <summary>
  Grafana
  </summary>

  ```yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: oauth2-grafana
    labels:
      app: oauth2-grafana
      target: oauth2-grafana
  spec:
    replicas: 1
    selector:
      matchLabels:
        app: oauth2-grafana
    template:
      metadata:
        labels:
          app: oauth2-grafana
      spec:
        containers:
        - name: oauth2-proxy
          image: quay.io/oauth2-proxy/oauth2-proxy:v7.1.3
          imagePullPolicy: IfNotPresent
          args:
          - --provider=github
          - --email-domain="*"
          - --http-address=0.0.0.0:3000
          - --upstream=http://monitoring-grafana.kyma-system.svc:80
          - --cookie-name=grafana_oauth2_proxy
          - --proxy-prefix=/oauth2
          - --ping-path=/oauth2/healthy
          - --silence-ping-logging=true
          - --reverse-proxy=true
          - --skip-provider-button=true
          - --cookie-secure
          envFrom:
          - secretRef:
              name: oauth2-grafana-secret
          ports:
          - name: http
            containerPort: 3000
            protocol: TCP
          livenessProbe:
            httpGet:
              path: /oauth2/healthy
              port: http
            initialDelaySeconds: 3
            timeoutSeconds: 2
          readinessProbe:
            httpGet:
              path: /oauth2/healthy
              port: http
            initialDelaySeconds: 3
            timeoutSeconds: 2
        securityContext:
          fsGroup: 65534
          runAsNonRoot: true
          runAsUser: 65534
  ```

  </details>
</div>


5. Create a Service for the `oauth2_proxy`:

<div tabs>
  <details>
  <summary>
  Kiali
  </summary>

  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: oauth2-kiali
    labels:
      app: oauth2-kiali
  spec:
    type: ClusterIP
    ports:
    - port: 3000
      name: http
      protocol: TCP
      targetPort: http
    selector:
      app: oauth2-kiali
  ```

  </details>
  <details>
  <summary>
  Grafana
  </summary>

  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: oauth2-grafana
    labels:
      app: oauth2-grafana
  spec:
    type: ClusterIP
    ports:
    - port: 3000
      name: http
      protocol: TCP
      targetPort: http
    selector:
      app: oauth2-grafana
  ```

  </details>
</div>

6. To expose the Service using Istio, create a VirtualService. Set the domain in the `hosts` list to your desired name:

<div tabs>
  <details>
  <summary>
  Kiali
  </summary>

  ```yaml
  apiVersion: networking.istio.io/v1alpha3
  kind: VirtualService
  metadata:
    name: oauth2-kiali
  spec:
    hosts:
    - kiali.kyma.example.com
    gateways:
    - kyma-system/kyma-gateway
    http:
    - match:
      - uri:
          regex: /.*
      route:
      - destination:
          port:
            number: 3000
          host: oauth2-kiali
  ```

  </details>
  <details>
  <summary>
  Grafana
  </summary>

  ```yaml
  apiVersion: networking.istio.io/v1alpha3
  kind: VirtualService
  metadata:
    name: oauth2-grafana
  spec:
    hosts:
    - grafana.kyma.example.com
    gateways:
    - kyma-system/kyma-gateway
    http:
    - match:
      - uri:
          regex: /.*
      route:
      - destination:
          port:
            number: 3000
          host: oauth2-grafana
  ```

  </details>
</div>
