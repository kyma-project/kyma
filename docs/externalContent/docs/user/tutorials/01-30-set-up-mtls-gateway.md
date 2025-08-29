# Set Up an mTLS Gateway

Learn how to set up an mTLS Gateway in Istio.

## Context

<!-- markdown-link-check-disable-next-line -->
According to the official [CloudFlare documentation](https://cloudflare.com/learning/access-management/what-is-mutual-tls/):
>Mutual TLS, or mTLS for short, is a method for mutual authentication. mTLS ensures that the parties at each end of a network connection are who they claim to be by verifying that they both have the correct private key. The information within their respective TLS certificates provides additional verification.

To establish a working mTLS connection, several things are required:

1. A working DNS entry pointing to the Istio Gateway IP
2. A valid Root CA certificate and key
3. Generated client and server certificates with a private key
4. Istio and API-Gateway installed on a Kubernetes cluster

The procedure of setting up a working mTLS Gateway is described in the following steps. The tutorial uses a Gardener shoot cluster and its API. The mTLS Gateway is exposed under your domain with a valid DNS `A` record.

## Prerequisites

* You have the Istio and API Gateway modules added.
* You have [set up your custom domain](./01-10-setup-custom-domain-for-workload.md).

## Steps

### Set Up an mTLS Gateway

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Create a DNS Entry and generate a wildcard certificate.

    > [!NOTE]
    > How to perform this step heavily depends on the configuration of a hyperscaler. Always consult the official documentation of each cloud service provider.

    For Gardener shoot clusters, including SAP BTP, Kyma runtime, follow [Set Up a Custom Domain For a Workload](01-10-setup-custom-domain-for-workload.md).

2. Generate a Root CA and a client certificate.

    This step is required for mTLS validation, which allows Istio to verify the authenticity of a client host.

    For a detailed step-by-step guide on how to generate a self-signed certificate, follow [Prepare Self-Signed Root Certificate Authority and Client Certificates](01-60-security/01-61-mtls-selfsign-client-certicate.md).

3. Set up Istio Gateway in mutual mode. 
    1. Go to **Istio > Gateways** and choose **Create**. 
    2. Add the name `kyma-mtls-gateway`.
    3. Add a server with the following configuration:
      - **Port Number**: `443`
      - **Name**: `mtls`
      - **Protocol**: `HTTPS`
      - **TLS Mode**: `MUTUAL`
      - **Credential Name**: `kyma-mtls-certs`
      - Add a host `*.{DOMAIN_NAME}`. Replace `{DOMAIN_NAME}` with the name of your custom domain.
    4. Choose **Create**.

4. Create a Secret containing the Root CA certificate.

    In order for the `MUTUAL` mode to work correctly, you must apply a Root CA in a cluster. This Root CA must follow the [Istio naming convention](https://istio.io/latest/docs/reference/config/networking/gateway/#ServerTLSSettings) so Istio can use it.
    Create an Opaque Secret containing the previously generated Root CA certificate in the `istio-system` namespace.

    1. Go to **Configuration > Secrets** and choose **Create**. 
    2. Provide the following configuration details:
      - **Name**: `kyma-mtls-certs`
      - **Type**: `Opaque`
      - In the `Data` section, choose **Read value from file**. Select the file that contains your Root CA certificate.

#### **kubectl**

1. Create a DNS Entry and generate a wildcard certificate.

    > [!NOTE]
    > How to perform this step heavily depends on the configuration of a hyperscaler. Always consult the official documentation of each cloud service provider.

    For Gardener shoot clusters, including SAP BTP, Kyma runtime, follow [Set Up a Custom Domain For a Workload](01-10-setup-custom-domain-for-workload.md).

2. Generate a Root CA and a client certificate.

    This step is required for mTLS validation, which allows Istio to verify the authenticity of a client host.

    For a detailed step-by-step guide on how to generate a self-signed certificate, follow [Prepare Self-Signed Root Certificate Authority and Client Certificates](01-60-security/01-61-mtls-selfsign-client-certicate.md).

3. To set up Istio Gateway in mutual mode, apply the Gateway resource.

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
      name: kyma-mtls-gateway
      namespace: default
    spec:
      selector:
        app: istio-ingressgateway
        istio: ingressgateway
      servers:
        - port:
            number: 443
            name: mtls
            protocol: HTTPS
          tls:
            mode: MUTUAL
            credentialName: kyma-mtls-certs
          hosts:
            - "*.{DOMAIN_NAME}"
    EOF
    ```

4. Create a Secret containing the Root CA certificate.

    In order for the `MUTUAL` mode to work correctly, you must apply a Root CA in a cluster. This Root CA must follow the [Istio naming convention](https://istio.io/latest/docs/reference/config/networking/gateway/#ServerTLSSettings) so Istio can use it.
    Create an Opaque Secret containing the previously generated Root CA certificate in the `istio-system` namespace. 

    Run the following command:

    ```bash
    kubectl create secret generic -n istio-system kyma-mtls-certs --from-file=cacert=cacert.crt
    ```
<!-- tabs:end -->
