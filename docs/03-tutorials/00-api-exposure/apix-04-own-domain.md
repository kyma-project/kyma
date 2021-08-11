---
title: Use a custom domain to expose a service
---

This tutorial shows how to set up your custom domain and prepare a certificate to use the domain for exposing a service.

To learn how to expose a service, go to [this](..//apix-01-expose-service-apigateway.md) tutorial.

## Prerequisites

Depending on your cluster provider, meet the relevant prerequisites.

<div tabs name="cluster-provider" group="cluster-provider">
  <details>
  <summary label="Gardener">
  Gardener cluster
  </summary>

If you use a Gardener cluster, add an annotation for the `DNSProvider` and `DNSEntry` Custom Resources (CRs). The annotation allows the DNS source controllers to watch resources on the default cluster and create DNS entries on the target cluster. The annotation should look as follows:

```bash
  annotations:
    dns.gardener.cloud/class: garden
```

For more details see the [official Gardener documentation](https://gardener.cloud/docs/concepts/networking/dns-managment/#automatic-creation-of-dns-entries-for-services-and-ingresses).

  </details>
  <details>
  <summary label="non-Gardener">
  Non-Gardener cluster
  </summary>

If you use a cluster not managed by Gardener, install the required components manually. Follow these steps:

1. Create a Namespace. Run:

   ```bash
   kubectl create ns {NAMESPACE_NAME}
   ```

2. Download the [External DNS Management](https://github.com/gardener/external-dns-management) and [Certificate Management](https://github.com/gardener/cert-management) projects locally.

3. Enable vertical Pod autoscaling on your cluster. For example, for Google Cloud Platform, run:

   ```bash
   gcloud beta container clusters update {PROJECT_NAME} --enable-vertical-pod-autoscaling
   ```

4. Install the `external-dns-management` component in your Namespace:

   ```bash
   helm install external-dns {PATH_TO_EXTERNAL_DNS_MANAGEMENT}/charts/external-dns-management --namespace={NAMESPACE_NAME} --set configuration.identifier=external-dns-identifier
   ```

5. Install the `cert-management` component in your Namespace:

   ```bash
   helm install cert-manager {PATH_TO_CERT_MANAGEMENT}/charts/cert-management --namespace={NAMESPACE_NAME} --set configuration.identifier=cert-manager-identifier
   ```

</details>
</div>

## Steps

Follow these steps to set up your custom domain and 

1. Install Kyma 2.0 or higher.

2. Create a Namespace. Run:

   ```bash
   kubectl create ns {NAMESPACE_NAME}
   ```

3. Create a Secret. See the [official External DNS Management docuemntation](https://github.com/gardener/external-dns-management) and choose your DNS cloud service provider for the relevant guidelines. Once the Secret Custom Resource (CR) is ready, run the following command:

   ```bash
   kubectl apply -n {NAMESPACE_NAME} -f {SECRET}.yaml
   ```

4. Set up the `external-dns-management` component.

- Create a DNSProvider Custom Resource Definition (CRD). See the following example and modify values for the **namespace**, [**spec.type**](https://github.com/gardener/external-dns-management#external-dns-management), **spec.secretRef.name** and **spec.domains.include** parameters. For the **spec.secretRef.name** parameter, use the value from `{SECRET}.yaml`.

   ```bash
   apiVersion: dns.gardener.cloud/v1alpha1
   kind: DNSProvider
   metadata:
     name: dns-provider
     namespace: NAMESPACE_NAME # Edit this value
     # Uncomment when using on a Gardener cluster
   #  annotations:
   #    dns.gardener.cloud/class: garden
   spec:
     type: your-provider-dns # Edit this value
     secretRef:
       name: secret-name # Edit this value
     domains:
       include:
         # Replace with a domain of the hosted zone
         - mydomain.com # Edit this value
   ```

   Once the DNSProvider CRD is ready, create the DNSProvider CR. Run:

   ```bash
   kubectl apply -n {NAMESPACE_NAME} -f {DNS-PROVIDER}.yaml
   ```

- Create a DNSEntry CRD. See the following example and modify values for the **namespace**, **spec.dnsName**, **spec.ttl**, and **spec.targets.IP** parameters:

   ```bash
   apiVersion: dns.gardener.cloud/v1alpha1
   kind: DNSEntry
   metadata:
     name: dns-entry
     namespace: NAMESPACE_NAME # Edit this value
     # Uncomment when using on a Gardener cluster
   #  annotations:
   #    dns.gardener.cloud/class: garden
   spec:
     dnsName: "private.sub1.mydomain.com" # Edit this value
     ttl: 600
     targets:
       - IP # Edit this value
   ```

   Once the DNSEntry CRD is ready, create the DNSEntry CR. Run:

   ```bash
   kubectl apply -n {NAMESPACE_NAME} -f {DNS-ENTRY}.yaml
   ```

   >**NOTE:** You can create many DNSEntry CRs for one DNSProvider depending on the number of subdomains you want to use.

5. Create an Issuer CRD. See the following example and modify values of the **spec.acme.email**, **spec.domains.include**, and **spec.domains.exclude** parameters. As the value for the **spec.domains.include** parameter, use the subdomain from `{DNS-ENTRY}.yaml`.

```bash
   apiVersion: cert.gardener.cloud/v1alpha1
   kind: Issuer
   metadata:
     name: letsencrypt-staging
     namespace: default
   spec:
     acme:
       server: https://acme-staging-v02.api.letsencrypt.org/directory
       email: YOUR_EMAIL_ADDRESS # Edit this value
       autoRegistration: true

       # Name of the Secret used to store the ACME account private key
       # If does not exist, a new one is created
       privateKeySecretRef:
         name: letsencrypt-staging-secret
         namespace: default

       # Optionally, restrict domain ranges for which certificates can be requested
       domains:
         include:
           - private.sub1.mydomain.com # Edit this value
   #     exclude:
   #       - private.sub2.mydomain.com # Edit this value
   ```

   Once the DNSEntry CRD is ready, create the DNSEntry CR. Run:

   ```bash
   kubectl apply -n {NAMESPACE-NAME} -f {ISSUER}.yaml
   ```

6. Create a Certificate CRD. See the following example and modify values of the **spec.secretName**, **spec.commonName**, , and **spec.issuerRef.name** parameters. As the value for the **spec.issuerRef.name** parameter, use the Issuer name from `{ISSUER}.yaml`.

   ```bash
   apiVersion: cert.gardener.cloud/v1alpha1
   kind: Certificate
   metadata:
     name: httpbin
     namespace: istio-system
   spec:
     secretName: httpbin-tls-credentials # Name of the created Secret. Edit this value
     commonName: mydomain.com # Edit this value
     issuerRef:
       # Name of the Issuer created previously
       name: letsencrypt-staging
       namespace: default
     dnsNames:
       - private.sub.mydomain.com # Edit this value
   ```

   Once the Certificate CRD is ready, create the Certificate CR. Run:

   ```bash
   kubectl apply -n {NAMESPACE_NAME} -f {CERTIFICATE}.yaml
   ```

When you finish the setup, go to [this](..//apix-01-expose-service-apigateway.md) tutorial to learn how to expose a service.