# Expose Workloads in Multiple Namespaces With a Single APIRule Definition

This tutorial shows how to expose Service endpoints in multiple namespaces using APIGateway Controller.

> [!WARNING]
>  Exposing a workload to the outside world causes a potential security vulnerability, so tread carefully. In a production environment, secure the workload you expose with [OAuth2](../../) or [JWT](../../01-50-expose-and-secure-a-workload/v1beta1-deprecated/01-52-expose-and-secure-workload-jwt.md).


##  Prerequisites

Create three namespaces. Deploy two instances of the HTTPBin Service, each in a separate namespace. To learn how to do it, follow the [Create a Workload](../../01-00-create-workload.md) tutorial. Reserve the third namespace for creating an APIRule.

> [!NOTE]
> Remember to enable automatic Istio sidecar proxy injection in each namespace. See [Enable Sidecar Injection for a Namespace](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection?id=enable-sidecar-injection-for-a-namespace).

## Steps

### Expose Your Workloads

1. Export the namespaces' and Services' names as environment variables:

    ```bash
    export FIRST_SERVICE={SERVICE_NAME}
    export SECOND_SERVICE={SERVICE_NAME}
    export NAMESPACE_FIRST_SERVICE={NAMESPACE_NAME}
    export NAMESPACE_SECOND_SERVICE={NAMESPACE_NAME}
    export NAMESPACE_APIRULE={NAMESPACE_NAME}
    ```
  
2. Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
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

3. Expose the HTTPBin Services in their respective namespaces by creating an APIRule custom resource (CR) in its own namespace. Run:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: httpbin-services
      namespace: $NAMESPACE_APIRULE
    spec:
      host: httpbin-services.$DOMAIN_TO_EXPOSE_WORKLOADS
      gateway: $GATEWAY
      rules:
        - path: /headers
          methods: ["GET"]
          service:
            name: $FIRST_SERVICE
            namespace: $NAMESPACE_FIRST_SERVICE
            port: 8000
          accessStrategies:
            - handler: no_auth
        - path: /get
          methods: ["GET"]
          service:
            name: $SECOND_SERVICE
            namespace: $NAMESPACE_SECOND_SERVICE
            port: 8000
          accessStrategies:
            - handler: no_auth
    EOF
    ```

    > [!NOTE]
    > If you are using k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.


### Access Your Workloads
To access your HTTPBin Services, use [curl](https://curl.se).

To call the endpoints, send `GET` requests to the HTTPBin Services:

  ```bash
  curl -ik -X GET https://httpbin-services.$DOMAIN_TO_EXPOSE_WORKLOADS/headers

  curl -ik -X GET https://httpbin-services.$DOMAIN_TO_EXPOSE_WORKLOADS/get
  ```
If successful, the calls return the `200 OK` response code.
