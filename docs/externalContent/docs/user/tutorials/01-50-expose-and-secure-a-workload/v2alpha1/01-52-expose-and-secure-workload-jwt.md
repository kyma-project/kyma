# Expose and Secure a Workload with JWT

This tutorial shows how to expose and secure Services using APIGateway Controller. The Controller reacts to an instance of the APIRule custom resource (CR) and creates an Istio [VirtualService](https://istio.io/latest/docs/reference/config/networking/virtual-service/), [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) and [Request Authentication](https://istio.io/latest/docs/reference/config/security/request_authentication/) according to the details specified in the CR. To interact with the secured workloads, the tutorial uses a JWT token.

## Prerequisites

* You have the Istio and API Gateway modules added.
* You have a deployed workload.
  > [!NOTE] 
  > To expose a workload using APIRule in version `v2alpha1`, the workload must be a part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection?id=enable-istio-sidecar-proxy-injection).
* You have [set up your custom domain](../../01-10-setup-custom-domain-for-workload.md). Alternatively, you can use the default domain of your Kyma cluster and the default Gateway `kyma-system/kyma-gateway`.
  
  > [!NOTE]
  > Because the default Kyma domain is a wildcard domain, which uses a simple TLS Gateway, it is recommended that you set up your custom domain for use in a production environment.

  > [!TIP]
  > To learn what the default domain of your Kyma cluster is, run `kubectl get gateway -n kyma-system kyma-gateway -o jsonpath='{.spec.servers[0].hosts}'`.

* You have a JSON Web Token. See [Obtain a JWT](../01-51-get-jwt.md).
* You must install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) and [curl](https://curl.se/).


## Steps

To expose and secure your Service, create the following APIRule:

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
  service:
    name: {SERVICE_NAME}
    port: {SERVICE_PORT}
  gateway: {GATEWAY_NAME}/{GATEWAY_NAMESPACE}
  rules:
    - jwt:
        authentications:
          - issuer: {ISSUER}
            jwksUri: {JWKS_URI}
      methods:
        - GET
      path: /*
EOF
```


### Access the Secured Resources

1. To call the endpoint, send a `GET` request to the HTTPBin Service.

    ```bash
    curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/headers
    ```
    You get the error `401 Unauthorized`.

2. Now, access the secured workload using the correct JWT.

    ```bash
    curl -ik -X GET https://{SUBDOMAIN}.{DOMAIN_NAME}/headers --header "Authorization:Bearer $ACCESS_TOKEN"
    ```
    You get the `200 OK` response code.