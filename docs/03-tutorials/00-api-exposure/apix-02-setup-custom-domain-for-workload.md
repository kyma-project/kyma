---
title: Set up a custom domain for a workload
---

This tutorial shows how to set up your custom domain and prepare a certificate for exposing a workload. It uses Gardener [External DNS Management](https://github.com/gardener/external-dns-management) and [Certificate Management](https://github.com/gardener/cert-management) components.
  >**NOTE:** Feel free to skip this tutorial if you use a Kyma domain instead of your custom domain.

## Prerequisites

If you use a cluster not managed by Gardener, install the [External DNS Management](https://github.com/gardener/external-dns-management) and [Certificate Management](https://github.com/gardener/cert-management) components manually in a dedicated Namespace.

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create those workloads, follow the [Create a workload](./apix-01-create-workload.md) tutorial.

## Steps

Follow these steps to set up your custom domain and prepare a certificate required to expose a workload.

1. Create a Secret containing credentials for your DNS cloud service provider account in your Namespace.

  See the [official External DNS Management documentation](https://github.com/gardener/external-dns-management/blob/master/README.md#external-dns-management), choose your DNS cloud service provider, and follow the relevant guidelines to create a secret in your Namespace. Then export the following value as an environment variable:

   ```bash
   export SECRET={SECRET_NAME}
   ```

2. Create a DNSProvider and a DNSEntry custom resource (CR).

   - Export the following values as environment variables and run the command provided.
  
   As the **SPEC_TYPE**, use the relevant provider type. See the [official Gardener examples](https://github.com/gardener/external-dns-management/tree/master/examples) of the DNSProvider CR.

   ```bash
   export SPEC_TYPE={PROVIDER_TYPE}
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME} 
   ```
   >**NOTE:** `DOMAIN_NAME` is the domain that you own, for example, mydomain.com


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
          - $DOMAIN_TO_EXPOSE_WORKLOADS
    EOF
    ```

   - Export the following values as environment variables and run the command provided:

   ```bash
   export IP=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}') # Assuming only one LoadBalancer with external IP
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
      dnsName: "*.$DOMAIN_TO_EXPOSE_WORKLOADS"
      ttl: 600
      targets:
        - $IP
    EOF
    ```

3. Create an Issuer CR.

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
           - $DOMAIN_TO_EXPOSE_WORKLOADS
           - "*.$DOMAIN_TO_EXPOSE_WORKLOADS"
   EOF
   ```

4. Create a Certificate CR.

  Export the following values as environment variables and run the command provided.

   ```bash
   export TLS_SECRET={TLS_SECRET_NAME} # The name of the TLS Secret that will be created in this step, for example, httpbin-tls-credentials
   export ISSUER={ISSUER_NAME} # The name of the Issuer CR, for example,letsencrypt-staging
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
     commonName: $DOMAIN_TO_EXPOSE_WORKLOADS
     issuerRef:
       name: $ISSUER
       namespace: $NAMESPACE
   EOF
   ```
   >**NOTE:** Run the following command to check the certificate status: `kubectl get certificate httpbin-cert -n istio-system `

5. Create a Gateway CR. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.istio.io/v1alpha3
   kind: Gateway
   metadata:
     name: httpbin-gateway
     namespace: $NAMESPACE
   spec:
     selector:
       istio: ingressgateway # Use Istio Ingress Gateway as default
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

## Next steps

- [Expose a workload](./apix-03-expose-workload-apigateway.md)
- [Expose multiple workloads on the same host](./apix-04-expose-multiple-workloads.md)
- [Expose and secure a workload with OAuth2](./apix-05-expose-and-secure-workload-oauth2.md)
- [Expose and secure a workload with Istio](./apix-07-expose-and-secure-workload-istio.md)
- [Expose and secure a workload with JWT](./apix-08-expose-and-secure-workload-jwt.md)