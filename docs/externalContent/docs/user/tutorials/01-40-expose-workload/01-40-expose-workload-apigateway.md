# Expose a Workload

This tutorial shows how to expose an unsecured instance of the HTTPBin Service and call its endpoints.

> [!WARNING]
>  Exposing a workload to the outside world is a potential security vulnerability, so be careful. In a production environment, always secure the workload you expose with [JWT](../01-50-expose-and-secure-a-workload/v2alpha1/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* You have the Istio and API Gateway modules added.
* You have a deployed workload. 
  > [!NOTE] 
  > To expose a workload using APIRule in version `v2`, the workload must be a part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection?id=enable-istio-sidecar-proxy-injection).
* To use CLI instructions, you must install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) and [curl](https://curl.se/). Alternatively, you can use Kyma dashboard.
* You have [set up your custom domain](../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.
  
  > [!NOTE]
  > Because the default Kyma domain is a wildcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment.

  > [!TIP]
  > To learn what the default domain of your Kyma cluster is, run `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'`.

## Steps

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Discovery and Network > API Rules** and select **Create**.
2. Provide the name of the APIRule CR.
3. Add the name and port of the service you want to expose.
4. Add a Gateway.
5. Add a rule with the following configuration:
    - **Path**: `/*`
    - **Handler**: `No Auth`
    - **Methods**: `GET`
6. Add one more rule with the following configuration:
    - **Path**: `/post`
    - **Handler**: `No Auth`
    - **Methods**: `POST`
7. Choose **Create**.

#### **kubectl**

To expose your workload, create an APIRule CR. You can adjust the configuration, if needed.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: {APIRULE_NAME}
  namespace: {APIRULE_NAMESPACE}
spec:
  hosts:
    - {SUBDOMAIN}.{DOMAIN_NAME}
  service:
    name: {SERVICE_NAME}
    namespace: {SERVICE_NAMESPACE}
    port: {SERVICE_PORT}
  gateway: {NAMESPACE/GATEWAY}
  rules:
    - path: /*
      methods: ["GET"]
      noAuth: true
    - path: /post
      methods: ["POST"]
      noAuth: true
EOF
```

<!-- tabs:end -->


### Access Your Workload

- Send a `GET` request to the exposed workload:

  ```bash
  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/ip
  ```
  If successful, the call returns the `200 OK` response code.

- Send a `POST` request to the exposed workload:

  ```bash
  curl -ik -X POST https://{SUBDOMAIN}.{DOMAIN_NAME}/post -d "test data"
  ```
  If successful, the call returns the `200 OK` response code.
