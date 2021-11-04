---
title: Use a custom domain to expose a service
---

This tutorial shows how to set up your custom domain and prepare a certificate for exposing a service. The components used are Gardener [External DNS Management](https://github.com/gardener/external-dns-management) and [Certificate Management](https://github.com/gardener/cert-management).

Once you finish the steps, learn how to [expose a service](./apix-02-expose-service-apigateway.md) or how to [expose and secure a service](./apix-03-expose-and-sercure-service.md).

## Prerequisites

To follow this tutorial, use Kyma 2.0 or higher.

If you use a cluster not managed by Gardener, install the [External DNS Management](https://github.com/gardener/external-dns-management) and [Certificate Management](https://github.com/gardener/cert-management) components manually in a dedicated Namespace.

> **NOTE:** This tutorial uses the External DNS Management v.0.10.4 and the Certificate Management v0.8.3.

## Steps

Follow these steps to set up your custom domain and prepare a certificate required to expose a service.

1. Create a Namespace. Run:

  ```bash
   kubectl create ns {NAMESPACE_NAME}
   ```

2. Create a Secret containing credentials for your DNS cloud service provider account in your Namespace.

  See the [official External DNS Management documentation](https://github.com/gardener/external-dns-management/blob/master/README.md#external-dns-management), choose your DNS cloud service provider, and follow the relevant guidelines. Then run:

  ```bash
  kubectl apply -n {NAMESPACE_NAME} -f {SECRET}.yaml
  ```

3. Create a DNSProvider and a DNSEntry CRs.

  > **CAUTION:** Bear in mind that the **metadata.annotation** parameter, may be either not needed or subject to change depending on the External DNS Management configuration provided during the component installation.

   - Export the following values as environment variables and run the command provided. 
  
   As the **SPEC_TYPE**, use the relevant provider type. See the [official Gardener examples](https://github.com/gardener/external-dns-management/tree/master/examples) of the DNSProvider CR.

   ```bash
   export NAMESPACE={NAMESPACE_NAME}
   export SPEC_TYPE={PROVIDER_TYPE}
   export SECRET={SECRET_NAME}
   export DOMAIN={CLUSTER_DOMAIN} # The domain that you own, e.g. mydomain.com.
   ```

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
      type: $SPEC_TYPE
      secretRef:
        name: $SECRET
      domains:
        include:
          - $DOMAIN
    EOF
    ```

   - Export the following values as environment variables and run the command provided:

   ```bash
   export WILDCARD={WILDCRAD_SUBDOMAIN} #e.g. *.api.mydomain.com
   export IP=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}') # assuming only one LoadBalancer with external IP
   ```

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
      dnsName: "$WILDCARD"
      ttl: 600
      targets:
        - $IP
    EOF
    ```

  >**NOTE:** You can create many DNSEntry CRs for one DNSProvider, depending on the number of subdomains you want to use. To simplify your setup, consider using a wildcard subdomain if all your DNSEntry objects share the same subdomain and resolve to the same IP, for example: `*.api.mydomain.com`. Remember that such a wildcard entry results in DNS configuration that doesn't support the following hosts: `api.mydomain.com` and `mydomain.com`. We don't use these hosts in this tutorial, but you can add DNS Entries for them if you need.

4. Create an Issuer CR.

  Export the following values as environment variables and run the command provided.

   ```bash
   export EMAIL={YOUR_EMAIL_ADDRESS}
   export DOMAIN={CLUSTER_DOMAIN} #e.g. mydomain.com
   export SUBDOMAIN={YOUR_SUBDOMAIN} #e.g. api.mydomain.com
   export WILDCARD={WILDCARD_SUBDOMAIN} #e.g. *.api.mydomain.com
   ```

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: cert.gardener.cloud/v1alpha1
   kind: Issuer
   metadata:
     name: letsencrypt-staging
     namespace: default
   spec:
     acme:
       server: https://acme-staging-v02.api.letsencrypt.org/directory
       email: $EMAIL
       autoRegistration: true
       # Name of the Secret used to store the ACME account private key
       # If doesn't exist, a new one is created
       privateKeySecretRef:
         name: letsencrypt-staging-secret
         namespace: default
       domains:
         include:
           - $DOMAIN
           - $SUBDOMAIN
           - "$WILDCARD"
         # Optionally, restrict domain ranges for which certificates can be requested
   #     exclude:
   #       - my.sub2.mydomain.com # Edit this value
   EOF
   ```

   > **NOTE:** The Issuer CR must be created in the `default` Namespace. It is a global CR which many Certificate CRs may refer to.

5. Create a Certificate CR.

  Export the following values as environment variables and run the command provided.

   ```bash
   export TLS_SECRET={SECRET_NAME} #e.g. httpbin-tls-credentials
   export ISSUER={ISSUER_NAME} # The name of the Issuer CR, e.g.letsencrypt-staging.
   ```

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: cert.gardener.cloud/v1alpha1
   kind: Certificate
   metadata:
     name: httpbin-cert
     namespace: istio-system
   spec:  
     secretName: $TLS_SECRET
     commonName: $DOMAIN
     issuerRef:
       name: $ISSUER
       namespace: default
     dnsNames:
       - "$WILDCARD"
       - $SUBDOMAIN
   EOF
   ```

   > **NOTE:** Istio requires the Certificate CR containing the Secret to be created in the `istio-system` Namespace.

## Next tutorial

Proceed with [this](./apix-02-expose-service-apigateway.md) tutorial to expose a service using your custom domain.
