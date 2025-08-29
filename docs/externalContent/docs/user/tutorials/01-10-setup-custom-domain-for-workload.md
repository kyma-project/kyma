# Set Up a Custom Domain for a Workload

This tutorial shows how to set up a custom domain and prepare a certificate required for exposing a workload. It uses the Gardener [External DNS Management](https://github.com/gardener/external-dns-management) and [Certificate Management](https://github.com/gardener/cert-management) components.

## Prerequisites

* You have a custom domain.
* If you use a cluster not managed by Gardener, install the [External DNS Management](https://github.com/gardener/external-dns-management#quick-start) and [Certificate Management](https://github.com/gardener/cert-management) components manually in a dedicated namespace. SAP BTP, Kyma runtime clusters are managed by Gardener, so you are not required to install any additional components.

## Steps

### Create a Secret with Credentials

Create a Secret containing credentials for the DNS cloud service provider account in your namespace. To learn how to do it, follow the [External DNS Management guidelines](https://github.com/gardener/external-dns-management/blob/master/README.md#external-dns-management).

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Select the namespace you want to use.
2. Go to **Configuration > Secrets**.
3. Choose **Create** and provide your configuration details.
4. Choose **Create**.

#### **kubectl**
Use `kubectl apply` to create a Secret containing the credentials and export its name as an environment variable:

```bash
export SECRET={SECRET_NAME}
```
<!-- tabs:end -->

### Create a DNSProvider Custom Resource (CR)

<!-- tabs:start -->
  #### **Kyma Dashboard**

1. Go to **Configuration > DNS Providers**.
2. Choose **Create** and provide the details:
    - **Name**: `dns-provider`
    - **Type**: is the type of your DNS cloud service provider.
    - Add the following annotation:
      - **dns.gardener.cloud/class**: `garden`
    - In the `Secret Reference` section, add these fields:
      - **Namespace** is the name of the namespace in which you created the Secret containing the credentials. 
      - **Name** is the name of the Secret.
    - In the `Include Domains` section, add your custom domain.
3. Choose **Create**.

#### **kubectl**

1. Export the following values as environment variables: the type of your DNS cloud service provider, the name of your custom domain, and the name of the namespace you want to use.

    ```bash
    export PROVIDER_TYPE={YOUR_PROVIDER_TYPE}
    export DOMAIN_TO_EXPOSE_WORKLOADS={YOUR_DOMAIN_NAME}
    export NAMESPACE={YOUR_NAMESPACE}
    ````
2. To create a DNSProvider CR, run:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSProvider
    metadata:
      name: dns-provider
      namespace: $NAMESPACE
      annotations:
        dns.gardener.cloud/class: garden
    spec:
      type: $PROVIDER_TYPE
      secretRef:
        name: $SECRET
      domains:
        include:
          - $DOMAIN_TO_EXPOSE_WORKLOADS
    EOF
    ```
<!-- tabs:end -->

### Create a DNSEntry CR

<!-- tabs:start -->
#### **Kyma Dashboard**
1. In the `istio-system` namespace, go to **Discovery and Network > Services**. Select the `istio-ingressgateway` Service and copy its external IP address.
2. In the namespace of your HTTPBin Deployment, go to **Configuration > DNS Entries**.
3. Choose **Create** and provide the details:
    - **Name**:`dns-entry`
    - Add the annotation:
      - **dns.gardener.cloud/class**: `garden`
    - For **DNSName**, use `*.{DOMAIN_TO_EXPOSE_WORKLOADS}`. Replace `{DOMAIN_TO_EXPOSE_WORKLOADS}` with the name of your custom domain.
    - Paste the external IP address of the `istio-ingressgateway` Service in the **Target** field.
4. Choose **Create**.

#### **kubectl**

1. Export the following values as environment variables:

    ```bash
    export IP=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}') # Assuming only one LoadBalancer with external IP
    ```
    > [!NOTE]
    > For some cluster providers you need to replace the `ip` with the `hostname`, for example, in AWS, set `jsonpath='{.status.loadBalancer.ingress[0].hostname}'`.

2. To create a DNSEntry CR, run:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: dns.gardener.cloud/v1alpha1
    kind: DNSEntry
    metadata:
      name: dns-entry
      namespace: $NAMESPACE
      annotations:
        dns.gardener.cloud/class: garden
    spec:
      dnsName: "*.$DOMAIN_TO_EXPOSE_WORKLOADS"
      ttl: 600
      targets:
        - $IP
    EOF
    ```
<!-- tabs:end -->

### Create a Certificate CR

> [!NOTE]
> While using the default configuration, certificates with the Let's Encrypt Issuer are valid for 90 days and automatically renewed 30 days before their validity expires. For more information, read the documentation on [Gardener Certificate Management](https://github.com/gardener/cert-management#requesting-a-certificate) and [Gardener extensions for certificate Services](https://gardener.cloud/docs/extensions/others/gardener-extension-shoot-cert-service/).

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Go to the `istio-system` namespace.
2. Go to **Configuration > Certificates**.
3. Choose **Create** and provide the details:
    - **Name**:`my-cert`
    - **Secret Name** is the name of your TLS Secret.
    - **Common Name** is the name of your custom domain.
4. Choose **Create**.

#### **kubectl**

1. Export the name of the TLS Secret that you would like to create:

    ```bash
    export TLS_SECRET={TLS_SECRET_NAME}
    ```

2. To create a Certificate CR, run:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: cert.gardener.cloud/v1alpha1
    kind: Certificate
    metadata:
      name: my-cert
      namespace: istio-system
    spec:
      secretName: $TLS_SECRET
      commonName: $DOMAIN_TO_EXPOSE_WORKLOADS
      issuerRef:
        name: garden
    EOF
    ```

3. To check the certificate status, run:

    ```bash
    kubectl get certificate my-cert -n istio-system
    ```
<!-- tabs:end -->

### Next Steps
[Set up a TLS Gateway](./01-20-set-up-tls-gateway.md) or [set up an mTLS Gateway](./01-30-set-up-mtls-gateway.md).

For more examples of CRs for Services and Ingresses, see the [Gardener external DNS management documentation](https://github.com/gardener/external-dns-management/tree/master/examples).
