---
title: Set up a custom domain for a workload
---

This tutorial shows how to set up a custom domain and prepare a certificate required for exposing a workload. Skip this tutorial if you want to use a Kyma domain instead of your custom domain.

## Prerequisites

* [A sample HttpBin service deployment and a sample Function](./apix-01-create-workload.md)
* If you use a cluster not managed by Gardener, install the [External DNS Management](https://github.com/gardener/external-dns-management#quick-start) and [Certificate Management](https://github.com/gardener/cert-management) components manually in a dedicated Namespace.

## Steps

1. Create a Secret containing credentials for the DNS cloud service provider account in your Namespace.

  * Choose your DNS cloud service provider and create a secret in your Namespace. To learn how to do it, follow [the guidelines](https://github.com/gardener/external-dns-management/blob/master/README.md#external-dns-management) provided in the External DNS Management documentation. 
  * Export the name of a secret as an environment variable:

    ```bash
    export SECRET={SECRET_NAME}
    ```

2. Create a `DNSProvider` custom resource (CR).

  * Export the following values as environment variables. As the `SPEC_TYPE`, use the relevant provider type. The `DOMAIN_NAME` value specifies the name of a domain that you own, for example, `mydomain.com`. 

    ```bash
    export SPEC_TYPE={PROVIDER_TYPE}
    export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME} 
    ````
  
  * To create a `DNSProvider` CR, run: 

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
  
3. Create a `DNSEntry` CR.
   
  * Export the following values as environment variables:

    ```bash
    export IP=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}') # Assuming only one LoadBalancer with external IP
    ```
   **NOTE:** For some cluster providers you need to replace the `ip` with the `hostname`, for example, in AWS, set `jsonpath='{.status.loadBalancer.ingress[0].hostname}'`.

  * To create a `DNSEntry` CR, run:

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

4. Create Issuer CR.

  * Export the following values as environment variables:

    ```bash
    export EMAIL={YOUR_EMAIL_ADDRESS}
    ```
  * To create Issuer CR, run: 

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

5. Create a Certificate CR.

  * Export the following values as environment variables. The `TLS_SECRET` value signifies the name of the TLS Secret, and the `ISSUER` value signifies the name of the Issuer CR.

    ```bash
    export TLS_SECRET={TLS_SECRET_NAME} # The name of the TLS Secret that will be created in this step, for example, httpbin-tls-credentials
    export ISSUER={ISSUER_NAME} # The name of the Issuer CR, for example,letsencrypt-staging
    ```

  * To create a Certificate CR, run:

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
  * Run the following command to check the certificate status: 
  
    ```bash
    `kubectl get certificate httpbin-cert -n istio-system`
    ```

6. Create Gateway CR. Run:

   >**NOTE:** Skip this step, if you're creating the mTLS gateway. 

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

To see more examples of custom resources for services and ingresses, visit the [Gardener external DNS management documentation](https://github.com/gardener/external-dns-management/tree/master/examples).