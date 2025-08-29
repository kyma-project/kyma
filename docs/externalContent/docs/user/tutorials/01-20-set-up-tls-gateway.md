# Set Up a TLS Gateway in Simple Mode

This tutorial shows how to set up a TLS Gateway in simple mode.

## Prerequisites

* [Set up your custom domain](./01-10-setup-custom-domain-for-workload.md).

## Steps


<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to **Istio > Gateways** and select **Create**.
2. Provide the following configuration details:
    - **Name**: `example-gateway`
    - Add a server with the following values:
      - **Port Number**: `443`
      - **Name**: `https`
      - **Protocol**: `HTTPS`
      - **TLS Mode**: `SIMPLE`
      - **Credential Name** is the name of the Secret that contains the credentials.
    - Use `{SUBDOMAIN}.{CUSTOM_DOMAIN}` as a host.

3. Select **Create**.

#### **kubectl**

1. Export the following values as environment variables:

    ```bash
    export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
    export NAMESPACE={YOUR_NAMESPACE}
    export GATEWAY=$NAMESPACE/example-gateway
    ```

2. To create a TLS Gateway in simple mode, run:

    ```bash
    cat <<EOF | kubectl apply -f -
    ---
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
      name: example-gateway
      namespace: $NAMESPACE
    spec:
      selector:
        istio: ingressgateway
      servers:
        - port:
            number: 443
            name: https
            protocol: HTTPS
          tls:
            mode: SIMPLE
            credentialName: $TLS_SECRET
          hosts:
            - "*.$DOMAIN_TO_EXPOSE_WORKLOADS"
    EOF
    ```

<!-- tabs:end -->