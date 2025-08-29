# Expose and Secure a Workload with a Certificate

This tutorial shows how to expose and secure a workload with mutual authentication using TLS Gateway.

## Prerequisites

* You have the Istio and API Gateway module added.
* You have a deployed workload.
  > [!NOTE] 
  > To expose a workload using APIRule in version `v2`, the workload must be a part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection?id=enable-istio-sidecar-proxy-injection).
* You have [set Up Your Custom Domain](../01-10-setup-custom-domain-for-workload.md).
* [Set up a mutual TLS Gateway](../01-30-set-up-mtls-gateway.md) and export the bundle certificates.
* Prepare a Client Root CA and certificate. For non-production environments, you can [create your own self-signed Client Root CA and certificate](../01-60-security/01-61-mtls-selfsign-client-certicate.md).

## Procedure

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to the namespace in which you want to create an APIRule CR.

2. Go to **Discovery and Network > APIRule** and select **Create**. 
3. Provide the following configuration details:
    - Add a name for your APIRule CR.
    - Add the name and namespace of the Gateway you want to use.
    - Specify the host.
4. Add a Rule with the following configuration:
    - **Path**:`/*`
    - **Methods**:`GET`
    - **Access Strategy**: `No Auth`
    - In **Requests > Headers**, add the following key-value pairs: 
      - **X-CLIENT-SSL-CN**: `%DOWNSTREAM_PEER_SUBJECT%`
      - **X-CLIENT-SSL-SAN**: `%DOWNSTREAM_PEER_URI_SAN%`
      - **X-CLIENT-SSL-ISSUER**: `%DOWNSTREAM_PEER_ISSUER%`
    - Add the name and port of the Service you want to expose.
5. Choose **Create**.

#### **kubectl**

1. Export the following values as environment variables:

  ```bash
  export CLIENT_ROOT_CA_CRT_FILE={CLIENT_ROOT_CA_CRT_FILE}
  export CLIENT_CERT_CN={COMMON_NAME}
  export CLIENT_CERT_ORG={ORGANIZATION}
  export CLIENT_CERT_CRT_FILE={CLIENT_CERT_CRT_FILE}
  export CLIENT_CERT_KEY_FILE={CLIENT_CERT_KEY_FILE}
  ```

2. Create an APIRule CR that adds the **X-CLIENT-SSL** headers to incoming requests.
   
   > [!NOTE] The namespace that you use for creating an APIRule must have Istio sidecar injection enabled. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection).

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2
    kind: APIRule
    metadata:
      name: {APIRULE_NAME}
      namespace: {APIRULE_NAMESPACE}
    spec:
      gateway: {GATEWAY_NAMESPACE}/{GATEWAY_NAME}
      hosts:
        - {SUBDOMAIN}.{DOMAIN}
      rules:
        - methods:
            - GET
          noAuth: true
          path: /*
          timeout: 300
          request:
            headers:
              X-CLIENT-SSL-CN: '%DOWNSTREAM_PEER_SUBJECT%'
              X-CLIENT-SSL-ISSUER: '%DOWNSTREAM_PEER_ISSUER%'
              X-CLIENT-SSL-SAN: '%DOWNSTREAM_PEER_URI_SAN%'
      service:
        name: {SERVICE_NAME}
        port: {SERVICE_PORT}
    EOF
    ```

<!-- tabs:end -->

## Access the Secured Resources

Call the secured endpoints of the HTTPBin Service.

In the following command, replace the name of the workload's subdomain and domain. Send a `GET` request to the Service with the client certificates that you used to create mTLS Gateway:

```bash
curl --key ${CLIENT_CERT_KEY_FILE} \
      --cert ${CLIENT_CERT_CRT_FILE} \
      --cacert ${CLIENT_ROOT_CA_CRT_FILE} \
      -ik -X GET https://{SUBDOMAIN}.{DOMAIN}/headers
```

If successful, the call returns the `200 OK` response code. If you call the Service without the proper certificates or with invalid ones, you get the error `403 Forbidden`.