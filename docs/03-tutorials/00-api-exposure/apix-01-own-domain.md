---
title: Use a custom domain to expose a service
---

This tutorial shows how to set up your custom domain and prepare a certificate for exposing a service. The components used are Gardener [External DNS Management](https://github.com/gardener/external-dns-management) and [Certificate Management](https://github.com/gardener/cert-management).

Once you finish the steps, learn how to [expose a service](./apix-02-expose-service-apigateway.md) or how to [expose and secure a service](./apix-03-expose-and-sercure-service.md).

## Prerequisites

To follow this tutorial, use Kyma 2.0 or higher.

If you use a cluster not managed by Gardener, install the [External DNS Management](https://github.com/gardener/external-dns-management) and [Certificate Management](https://github.com/gardener/cert-management) components manually in a dedicated Namespace.

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

3. Create a DNS Provider and a DNS Entry CRs.

   - Export the following values as environment variables and run the command provided.
  
   As the **SPEC_TYPE**, use the relevant provider type. See the [official Gardener examples](https://github.com/gardener/external-dns-management/tree/master/examples) of the DNS Provider CR.

   ```bash
   export NAMESPACE={NAMESPACE_NAME}
   export SPEC_TYPE={PROVIDER_TYPE}
   export SECRET={SECRET_NAME}
   export DOMAIN={DOMAIN_NAME} # The domain that you own, e.g. mydomain.com.
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

4. Create an Issuer CR.

  Export the following values as environment variables and run the command provided.

   ```bash
   export EMAIL={YOUR_EMAIL_ADDRESS}
   ```

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: cert.gardener.cloud/v1alpha1
   kind: Issuer
   metadata:
     name: letsencrypt-staging
     namespace: $NAMESPACE
   spec:
     acme:
       server: https://acme-staging-v02.api.letsencrypt.org/directory
       email: $EMAIL
       autoRegistration: true
       privateKeySecretRef:
         name: letsencrypt-staging-secret
         namespace: $NAMESPACE
       domains:
         include:
           - "$WILDCARD"
   EOF
   ```

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
   EOF
   ```

## Next steps

Proceed with the [Deploy a service](./apix-02-deploy-service.md) tutorial to deploy an instance of the HttpBin service or a sample Function.

Once you have your service deployed, you can continue by choosing one of the following tutorials to:

- [Expose a service](./apix-02-expose-service-apigateway.md)
- [Expose and secure a service with OAuth2](./apix-03-expose-and-secure-service-oauth2.md)
- [Expose and secure a service with JTW](./apix-04-expose-and-secure-service-jwt.md)
