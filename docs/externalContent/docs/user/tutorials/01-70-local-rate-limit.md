# Configuring Local Rate Limiting

The RateLimit custom resource (CR) allows you to apply local rate limit configuration for specific paths of an exposed application.

> [!NOTE]
> Local rate limits apply to the traffic that is directed toward a workload. If configured improperly, an attacker can exhaust all tokens and cause a Denial-of-Service attack, making the service inaccessible.

## Prerequisites

* You have the Istio and API Gateway modules added.
* You have [set up your custom domain](./01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.
  
  > [!NOTE]
  > Because the default Kyma domain is a wildcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment.

  > [!TIP]
  > To learn what the default domain of your Kyma cluster is, run '`kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'`.


## Deploying a Sample Service

1. Create a test namespace and enable Istio sidecar injection:
    ```bash
    kubectl create namespace test
    kubectl label namespace test istio-injection=enabled
    ```

2. Deploy and expose a simple HTTPBin Service:
    ```bash
    kubectl run httpbin --namespace test --image=kennethreitz/httpbin --labels app=httpbin
    kubectl expose --namespace test pod httpbin --port 80
    ```

3. Create an APIRule to expose the previously created workload:

    >[!NOTE]
    > `httpbin.local.kyma.dev` domain will always resolve to `127.0.0.1`.
    > Make sure that istio-ingressgateway is accessible under that IP.
    
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2alpha1
    kind: APIRule
    metadata:
      name: httpbin
      namespace: test
    spec:
      hosts:
        - httpbin.local.kyma.dev
      gateway: kyma-system/kyma-gateway
      rules:
        - path: /*
          service:
            name: httpbin
            port: 80
          methods: ["GET","POST"]
          noAuth: true
    EOF
    ```

    To verify the connection to the HTTPBin workload, run:
    ```bash
    curl -Lk https://httpbin.local.kyma.dev/ip
    ```

    If successful, you get the response:
    ```
    {
       "origin": "127.0.0.1"
    }
    ```

## Deploying Path-Based Rate Limit Configuration

The following example sets up a local rate limit for all endpoints exposed by the HTTPBin Service.
Additionally, it configures separate rate limit configuration for the `/ip` path.

Make sure that the **enableResponseHeaders** field is set to `true`. This enables the **x-ratelimit-limit** and **x-ratelimit-remaining** response headers, which can help confirm that the rate limits are working.

> [!NOTE]
> The token limit must be a multiple of the token bucket fill timer. 
> If the configuration is incorrect, the RateLimit CR is in the `Error` state, and the rate limit is not applied.

Apply the following RateLimit CR into the cluster:
```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1alpha1
kind: RateLimit
metadata:
  labels:
    app: httpbin
  name: ratelimit-path-sample
  namespace: test
spec:
  selectorLabels:
    app: httpbin
  enableResponseHeaders: true
  local:
    defaultBucket:
      maxTokens: 5
      tokensPerFill: 5
      fillInterval: 60s
    buckets:
      - path: /ip
        bucket:
          maxTokens: 10
          tokensPerFill: 5
          fillInterval: 60m
EOF
```

To check if the rate limit configuration is applied, run:
```bash
kubectl get ratelimits --namespace test ratelimit-path-sample
```

If successful, you get the following response:
```
NAME                    STATUS   AGE
ratelimit-path-sample   Ready    1s
```

To check if the rate limit configuration is working, run:
```bash
curl -kLv https://httpbin.local.kyma.dev/ip
```

If successful, the response contains the **x-ratelimit-limit** and **x-ratelimit-remaining** headers:
```
(...)
* Request completely sent off
< HTTP/2 200 
< server: istio-envoy
< date: ***
< content-type: application/json
< content-length: 29
< x-envoy-upstream-service-time: 1
< x-ratelimit-limit: 10
< x-ratelimit-remaining: 9
< 
{
  "origin": "127.0.0.1"
}
* Connection #0 to host httpbin.local.kyma.dev left intact
```

You have configured the path-based rate limits.
You can apply only one RateLimit CR in the cluster. To follow the next example, remove the RateLimit CR from a cluster.
```bash
kubectl delete ratelimits -n test ratelimit-path-sample
```

## Deploying Header-Based Rate Limit Configuration

The following example sets up a local rate limit for all endpoints exposed by the HTTPBin Service.
Additionally, it configures a separate rate limit for requests with the header **X-Rate-Limited** set to `true`.

Make sure that the **enableResponseHeaders** field is set to `true`. This enables the **x-ratelimit-limit** and **x-ratelimit-remaining** response headers, which can help confirm that the rate limits are working.

> [!NOTE]
> The token limit must be a multiple of the token bucket fill timer. 
> If the configuration is incorrect, the RateLimit CR is in the `Error` state, and the rate limit is not applied.

Apply the following RateLimit CR:
```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1alpha1
kind: RateLimit
metadata:
  labels:
    app: httpbin
  name: ratelimit-header-sample
  namespace: test
spec:
  selectorLabels:
    app: httpbin
  enableResponseHeaders: true
  local:
    defaultBucket:
      maxTokens: 1
      tokensPerFill: 1
      fillInterval: 30s
    buckets:
      - headers:
          X-Rate-Limited: "true"
        bucket:
          maxTokens: 10
          tokensPerFill: 5
          fillInterval: 30s
EOF
```

To check if the rate limit configuration is applied, run:
```bash
kubectl get ratelimits --namespace test ratelimit-header-sample
```

If successful, you get the following response:
```
NAME                      STATUS   AGE
ratelimit-header-sample   Ready    1s
```

To check if the rate limit configuration is working, run:
```bash
curl -kLv https://httpbin.local.kyma.dev/headers
```

If successful, the response contains the **x-ratelimit-limit** and **x-ratelimit-remaining** headers:
```
(...)
* Request completely sent off
< HTTP/2 200 
< server: istio-envoy
< date: ***
< content-type: application/json
< content-length: 529
< x-envoy-upstream-service-time: 17
< x-ratelimit-limit: 1
< x-ratelimit-remaining: 0
< 
{
  "headers": {
    "Accept": "*/*", 
    "Host": "httpbin.local.kyma.dev", 
    "User-Agent": "curl/8.7.1", 
    "X-Envoy-Attempt-Count": "1", 
    "X-Envoy-Expected-Rq-Timeout-Ms": "180000", 
    "X-Envoy-Internal": "true", 
    "X-Forwarded-Host": "httpbin.local.kyma.dev"
  }
}
* Connection #0 to host httpbin.local.kyma.dev left intact
```

If you provide an `X-Rate-Limited: "true"` header in a request, it is projected with different rate limits.
To check if the header-based rate limit is configured, run:
```bash
curl -H "X-Rate-Limited: true" -kLv https://httpbin.local.kyma.dev/headers
```

If successful, the response contains the **x-ratelimit-limit** and **x-ratelimit-remaining** headers:
```
(...)
* Request completely sent off
< HTTP/2 200 
< server: istio-envoy
< date: ***
< content-type: application/json
< content-length: 560
< x-envoy-upstream-service-time: 2
< x-ratelimit-limit: 10
< x-ratelimit-remaining: 9
< 
{
  "headers": {
    "Accept": "*/*", 
    "Host": "httpbin.local.kyma.dev", 
    "User-Agent": "curl/8.7.1", 
    "X-Envoy-Attempt-Count": "1", 
    "X-Envoy-Expected-Rq-Timeout-Ms": "180000", 
    "X-Envoy-Internal": "true", 
    "X-Forwarded-Host": "httpbin.local.kyma.dev", 
    "X-Rate-Limited": "true"
  }
}
* Connection #0 to host httpbin.local.kyma.dev left intact
```

You have configured the header-based rate limits.
You can apply only one RateLimit CR in the cluster. To follow the next example, remove the RateLimit CR from a cluster.
```bash
kubectl delete ratelimits -n test ratelimit-header-sample
```

## Path and header-based rate limit configuration

You can also set a rate limit to control the number of connections per path.
That means both `path` and `headers` fields can be used freely.

The following example sets up a local rate limit for all endpoints exposed by HTTPBin Service.
Additionally, it configures a separate rate limit configuration for the `/headers` path, which is applied only if the request contains the `X-Rate-Limited: true` header.

Make sure that the **enableResponseHeaders** field is set to `true`. This enables the **x-ratelimit-limit** and **x-ratelimit-remaining** response headers, which can help confirm that the rate limits are working.

> [!NOTE]
> The token limit must be a multiple of the token bucket fill timer. 
> If the configuration is incorrect, the RateLimit CR is in the `Error` state, and the rate limit is not applied.

Apply the following RateLimit CR:
```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1alpha1
kind: RateLimit
metadata:
  labels:
    app: httpbin
  name: ratelimit-path-header-sample
  namespace: test
spec:
  selectorLabels:
    app: httpbin
  enableResponseHeaders: true
  local:
    defaultBucket:
      maxTokens: 1
      tokensPerFill: 1
      fillInterval: 30s
    buckets:
      - headers:
          X-Rate-Limited: "true"
        path: /headers
        bucket:
          maxTokens: 10
          tokensPerFill: 5
          fillInterval: 30s
EOF
```

To check if the rate limit configuration is applied, run:
```bash
kubectl get ratelimits --namespace test ratelimit-path-header-sample
```

If successful, you get the response:
```
NAME                           STATUS   AGE
ratelimit-path-header-sample   Ready    1s
```

To check if the rate limit configuration is working, run the following command. This call uses tokens from the default bucket:
```bash
curl -kLv https://httpbin.local.kyma.dev/headers
```

If successful, the response contains the **x-ratelimit-limit** and **x-ratelimit-remaining** headers:
```
(...)
* Request completely sent off
< HTTP/2 200 
< server: istio-envoy
< date: ***
< content-type: application/json
< content-length: 529
< x-envoy-upstream-service-time: 17
< x-ratelimit-limit: 1
< x-ratelimit-remaining: 0
< 
{
  "headers": {
    "Accept": "*/*", 
    "Host": "httpbin.local.kyma.dev", 
    "User-Agent": "curl/8.7.1", 
    "X-Envoy-Attempt-Count": "1", 
    "X-Envoy-Expected-Rq-Timeout-Ms": "180000", 
    "X-Envoy-Internal": "true", 
    "X-Forwarded-Host": "httpbin.local.kyma.dev"
  }
}
* Connection #0 to host httpbin.local.kyma.dev left intact
```

If you provide the `X-Rate-Limited: "true"` header in a request to the `/headers` endpoint, it is projected with different rate limits.
To check if the header-based rate limit is configured, run:
```bash
curl -H "X-Rate-Limited: true" -kLv https://httpbin.local.kyma.dev/headers
```

If successful, you get the following response:
```
(...)
* Request completely sent off
< HTTP/2 200 
< server: istio-envoy
< date: ***
< content-type: application/json
< content-length: 560
< x-envoy-upstream-service-time: 2
< x-ratelimit-limit: 10
< x-ratelimit-remaining: 9
< 
{
  "headers": {
    "Accept": "*/*", 
    "Host": "httpbin.local.kyma.dev", 
    "User-Agent": "curl/8.7.1", 
    "X-Envoy-Attempt-Count": "1", 
    "X-Envoy-Expected-Rq-Timeout-Ms": "180000", 
    "X-Envoy-Internal": "true", 
    "X-Forwarded-Host": "httpbin.local.kyma.dev", 
    "X-Rate-Limited": "true"
  }
}
* Connection #0 to host httpbin.local.kyma.dev left intact
```

If you send the request with the header to a different endpoint, the token from the default bucket is still used. To verify that access to the `/ip` endpoint is also rate-limited, run:

```bash
curl -H "X-Rate-Limited: true" -kLv https://httpbin.local.kyma.dev/ip
```

You get the `HTTP/2 429` status code, which confirms that the rate limit has been exceeded:
```
(...)
> X-Rate-Limited: true
> 
* Request completely sent off
< HTTP/2 429 
< content-length: 18
< content-type: text/plain
< x-ratelimit-limit: 1
< x-ratelimit-remaining: 0
< date: Wed, 22 Jan 2025 14:07:10 GMT
< server: istio-envoy
< x-envoy-upstream-service-time: 2
< 
* Connection #0 to host httpbin.local.kyma.dev left intact
local_rate_limited
```

You have configured the path-based and header-based rate limits. Remove the RateLimit CR from a cluster.
```bash
kubectl delete ratelimits -n test ratelimit-path-header-sample
```