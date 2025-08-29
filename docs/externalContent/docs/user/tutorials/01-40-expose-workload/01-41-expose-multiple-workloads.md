# Expose Multiple Workloads on the Same Host

Learn how to expose multiple workloads on different paths by defining a Service at the root level and by defining Services on each path separately.

> [!WARNING]
>  Exposing a workload to the outside world is always a potential security vulnerability, so be careful. In a production environment, remember to secure the workload you expose with [JWT](../01-50-expose-and-secure-a-workload/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* You have the Istio and API Gateway modules added.
* You have deployed two workloads in one namespace.
  > [!NOTE] 
  > To expose a workload using APIRule in version `v2`, the workload must be a part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection?id=enable-istio-sidecar-proxy-injection).
* To use CLI instructions, you must install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) and [curl](https://curl.se/). Alternatively, you can use Kyma dashboard.
* You have [set up your custom domain](../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.
  
  > [!NOTE]
  > Because the default Kyma domain is a wildcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment.

  > [!TIP]
  > To learn what the default domain of your Kyma cluster is, run `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'`.

## Define Multiple Services on Different Paths

## Context
Learn how to expose two Services on different paths at the `spec.rules` level without a root Service defined.

## Steps

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Discovery and Network > API Rules** and choose **Create**.
2. Provide the name of the APIRule CR.
3. Add a Gateway.
4. Add a rule with the configuration details of the first Service.
5. Add another rule with the configuration details of the second Service.
6. Choose **Create**.

#### **kubectl**
Replace the placeholders and run the following command:

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
    gateway: {GATEWAY_NAMESPACE}/{GATEWAY_NAME}
    rules:
    - path: /headers
      methods: ["GET"]
      noAuth: true
      service:
        name: {FIRST_SERVICE_NAME}
        port: {FIRST_SERVICE_PORT}
    - path: /get
      methods: ["GET"]
      noAuth: true
      service:
        name: {SECOND_SERVICE_NAME}
        port: {SECOND_SERVICE_PORT}
  EOF
  ```
<!-- tabs:end -->

## Define a Service at the Root Level

## Context

You can also define a Service at the root level. Such a definition is applied to all the paths specified at **spec.rules** that do not have their own Services defined. Services defined at the **spec.rules** level have precedence over Service definition at the **spec.service** level.

## Steps

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Discovery and Network > API Rules** and choose **Create**.
2. Provide the name of the APIRule CR.
3. Add a Gateway.
4. Define a Service in the `Service` section.
5. Add one rule wihout a Service definition. Use the following configuration:
  - **Path**: `/headers`
  - **Handler**: `No Auth`
  - **Methods**: `GET`
6. Add another rule with the Service definition. Use the following configuration:
  - **Path**: `/get`
  - **Handler**: `No Auth`
  - **Methods**: `POST`
  - Add the name and namespace of the Second namespace.
7. Choose **Create**.

#### **kubectl**
Replace the placeholders and run the following command:
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
  gateway: {GATEWAY_NAMESPACE}/{GATEWAY_NAME}
  service:
    name: {FIRST_SERVICE_NAME}
    port: {FIRST_SERVICE_PORT}
  rules:
    - path: /headers
      methods: ["GET"]
      noAuth: true
    - path: /get
      methods: ["GET"]
      noAuth: true
      service:
        name: {SECOND_SERVICE_NAME}
        port: {SECOND_SERVICE_PORT}
EOF
```
<!-- tabs:end -->

## Access Your Workloads

To call the endpoints, send `GET` requests to the exposed Services:

  ```bash
  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/headers

  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/get
  ```
If successful, the calls return the `200 OK` response code.