# Expose Multiple Workloads on the Same Host

Learn how to expose multiple workloads on different paths by defining a Service at the root level and by defining Services on each path separately.

> [!WARNING]
>  Exposing a workload to the outside world is always a potential security vulnerability, so be careful. In a production environment, remember to secure the workload you expose with [JWT](../../01-50-expose-and-secure-a-workload/v2alpha1/01-52-expose-and-secure-workload-jwt.md).

## Prerequisites

* You have the Istio and API Gateway modules added.
* You have deployed two workloads in one namespace.
  > [!NOTE] 
  > To expose a workload using APIRule in version `v2alpha1`, the workload must be a part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection?id=enable-istio-sidecar-proxy-injection).
* You must install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) and [curl](https://curl.se/).
* You have [set up your custom domain](../../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.
  
  > [!NOTE]
  > Because the default Kyma domain is a wildcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment.

  > [!TIP]
  > To learn what the default domain of your Kyma cluster is, run `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'`.

## Define Multiple Services on Different Paths

## Context
Learn how to expose two Services on different paths at the `spec.rules` level without a root Service defined.

## Steps

Replace the placeholders and run the following command:

  ```bash
  cat <<EOF | kubectl apply -f -
  apiVersion: gateway.kyma-project.io/v2alpha1
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

## Define a Service at the Root Level

## Context

You can also define a Service at the root level. Such a definition is applied to all the paths specified at **spec.rules** that do not have their own Services defined. Services defined at the **spec.rules** level have precedence over Service definition at the **spec.service** level.

## Steps

Replace the placeholders and run the following command:
```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2alpha1
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

## Access Your Workloads

To call the endpoints, send `GET` requests to the exposed Services:

  ```bash
  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/headers

  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/get
  ```
If successful, the calls return the `200 OK` response code.