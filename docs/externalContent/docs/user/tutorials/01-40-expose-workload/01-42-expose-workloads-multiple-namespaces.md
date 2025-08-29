# Expose Workloads in Multiple Namespaces With a Single APIRule Definition

Learn how to expose Service endpoints in multiple namespaces.

> [!WARNING]
>  Exposing a workload to the outside world causes a potential security vulnerability, so be careful. In a production environment, secure the workload you expose with [JWT](../01-50-expose-and-secure-a-workload/01-52-expose-and-secure-workload-jwt.md).


##  Prerequisites

* You have the Istio and API Gateway modules added.
* You have deployed two workloads in different namespaces.
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

1. Create a namespace with the Istio sidecar proxy injection enabled.
2. In the created namespace, go to **Discovery and Network > API Rules** and choose **Create**.
3. Switch to the `YAML` section.
4. Paste the following APIRule custom resource (CR) and replace the placeholders:
    ```YAML
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
          service:
            name: {FIRST_SERVICE_NAME}
            namespace: {FIRST_SERVICE_NAMESPACE}
            port: {FIRST_SERVICE_PORT}
          noAuth: true
        - path: /get
          methods: ["GET"]
          service:
            name: {SECOND_SERVICE_NAME}
            namespace: {SECOND_SERVICE_NAMESPACE}
            port: {SECOND_SERVICE_PORT}
          noAuth: true
    ```
5. Choose **Create**.

#### **kubectl**

1. Create a separate namespace for the APIRule CR with enabled Istio sidecar proxy injection.
    ```bash
    export NAMESPACE={NAMESPACE_NAME}
    kubectl create ns $NAMESPACE
    kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
    ```
2. Expose the Services in their respective namespaces by creating an APIRule custom resource (CR) in its own namespace. Run:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: {APIRULE_NAME}
      namespace: $NAMESPACE
    spec:
      hosts:
        - {SUBDOMAIN}.{DOMAIN_NAME}
      gateway: {GATEWAY_NAMESPACE}/{GATEWAY_NAME}
      rules:
        - path: /headers
          methods: ["GET"]
          service:
            name: {FIRST_SERVICE_NAME}
            namespace: {FIRST_SERVICE_NAMESPACE}
            port: {FIRST_SERVICE_PORT}
          noAuth: true
        - path: /get
          methods: ["GET"]
          service:
            name: {SECOND_SERVICE_NAME}
            namespace: {SECOND_SERVICE_NAMESPACE}
            port: {SECOND_SERVICE_PORT}
          noAuth: true
    EOF
    ```
<!-- tabs:end -->

### Access Your Workloads

To call the endpoints, send `GET` requests to the exposed Services:

  ```bash
  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/headers

  curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/get
  ```
  
If successful, the calls return the `200 OK` response code.
