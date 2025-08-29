# Expose and Secure a Workload with JWT

This tutorial shows how to expose and secure Services using APIGateway Controller. The Controller reacts to an instance of the APIRule custom resource (CR) and creates an Istio VirtualService and [Oathkeeper Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules) according to the details specified in the CR. To interact with the secured workloads, the tutorial uses a JWT token.

## Prerequisites

* [Deploy a sample HTTPBin Service](../../01-00-create-workload.md).
* [Obtain a JSON Web Token (JWT)](../01-51-get-jwt.md).
* [Set up your custom domain](../../01-10-setup-custom-domain-for-workload.md) or use a Kyma domain instead.

## Steps

### Expose and Secure Your Workload

1. Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
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

2. To expose and secure the Service, create the following APIRule:
    
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: httpbin
      namespace: $NAMESPACE
    spec:
      host: httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS   
      service:
        name: httpbin
        port: 8000
      gateway: $GATEWAY
      rules:
        - accessStrategies:
          - handler: jwt
            config:
              jwks_urls:
              - $JWKS_URI
          methods:
            - GET
          path: /.*
    EOF
    ```

### Access the Secured Resources

To access your HTTPBin Service, use [curl](https://curl.se).

1. To call the endpoint, send a `GET` request to the HTTPBin Service.

    ```bash
    curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers
    ```
    You get the error `401 Unauthorized`.

2. Now, access the secured workload using the correct JWT.

    ```bash
    curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers --header "Authorization:Bearer $ACCESS_TOKEN"
    ```
    You get the `200 OK` response code.