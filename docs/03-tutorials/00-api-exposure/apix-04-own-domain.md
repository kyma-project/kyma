---
title: Use a custom domain to expose a service
---

This tutorial shows how to set up your custom domain and prepare a certificate to use the domain for exposing a service. The components used are Gardener [External DNS Management](https://gardener.cloud/docs/concepts/networking/dns-managment/#external-dns-management) and [Certificate Management](https://gardener.cloud/docs/concepts/networking/cert-managment/).

To learn how to expose a service, go to [this](./apix-01-expose-service-apigateway.md) tutorial.

## Prerequisites

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

When you meet the prerequisites, go directly to step 2.

## Steps

Follow these steps to set up your custom domain and prepare a certificate required to expose a service.

1. Install Kyma 2.0 or higher.

2. Create a Namespace. Run:

   ```bash
   kubectl create ns {NAMESPACE_NAME}
   ```

3. Create a Secret containing credentails for your DNS cloud service provider account. See the [official External DNS Management docuemntation](https://github.com/gardener/external-dns-management/blob/master/README.md#external-dns-management), choose your DNS cloud service provider, and follow the relevant guidelines. Once the YAML file with the required parameters is ready, create a Secret Custom Resource (CR). Run the following command:

   ```bash
   kubectl apply -n {NAMESPACE_NAME} -f {SECRET}.yaml
   ```

4. Set up the `external-dns-management` component.

- Create a DNSProvider CR. See the following example and modify values for the **metadata.namespace**, [**spec.type**](https://github.com/gardener/external-dns-management#using-the-dns-controller-manager), **metadata.spec.secretRef.name** and **metadata.spec.domains.include** parameters. As the value of **spec.type**, use the relevant provider type. See the [official Gardener examples](https://github.com/gardener/external-dns-management/tree/master/examples) of the DNSProvider CR. For the **spec.secretRef.name** parameter, use the **metadata.name** value from your `{SECRET}.yaml`. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: dns.gardener.cloud/v1alpha1
   kind: DNSProvider
   metadata:
     name: dns-provider
     namespace: NAMESPACE_NAME # Edit this value
     annotations:
       dns.gardener.cloud/class: garden
   spec:
     type: your-provider-dns-type # Edit this value
     secretRef:
       name: secret-name # Edit this value
     domains:
       include:
         - mydomain.com # The main domain. Edit this value
   EOF  
   ```

- Create a DNSEntry CR. See the following example and modify values for the **metadata.namespace**, **metadata.spec.dnsName**, and **metadata.spec.targets.IP** parameters. Optionally, you can also change the value of **metadata.spec.ttl**. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: dns.gardener.cloud/v1alpha1
   kind: DNSEntry
   metadata:
     name: dns-entry
     namespace: NAMESPACE_NAME # Edit this value
     annotations:
       dns.gardener.cloud/class: garden
   spec:
     dnsName: "my.sub1.mydomain.com" # Your subdomain. Edit this value
     ttl: 600
     targets:
       - 1.2.3.4 # IP address. Edit this value
   EOF
   ```

    >**NOTE:** To check the Ingress Gateway IP, run:
    >```
    >kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
    >```

   >**NOTE:** You can create many DNSEntry CRs for one DNSProvider, depending on the number of subdomains you want to use.

5. Create an Issuer CR. See the following example and modify values of the **metadata.spec.acme.email**, **metadata.spec.domains.include**, and **metadata.spec.domains.exclude** parameters. As the values for the **metadata.spec.domains.include** parameter, use the main domain, the subdomain from the **spec.dnsName** parameter of the DNSEntry CR, and a wildcard DNS record. Run:

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
       email: YOUR_EMAIL_ADDRESS # Edit this value
       autoRegistration: true
       # Name of the Secret used to store the ACME account private key
       # If doesn't exist, a new one is created
       privateKeySecretRef:
         name: letsencrypt-staging-secret
         namespace: default
       domains:
         include:
           - mydomain.com # The main domain. Edit this value
           - my.sub1.mydomain.com # The subdomain provided in the DNS Entry created in the previous step. Edit this value
           - "*.mydomain.com" # A wildcard DNS record. Edit this value
         # Optionally, restrict domain ranges for which certificates can be requested
   #     exclude:
   #       - my.sub2.mydomain.com # Edit this value
   EOF
   ```

6. Create a Certificate CR. See the following example and modify values of the **spec.secretName**, **spec.commonName**, and **spec.issuerRef.name**, and **spec.dnsNames** parameters. As the value for the **spec.issuerRef.name** parameter, use the value from te **metadata.name** parameter of the Issuer CR. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: cert.gardener.cloud/v1alpha1
   kind: Certificate
   metadata:
     name: httpbin-cert
     namespace: istio-system
   spec:
     secretName: httpbin-tls-credentials # Name of the Secret created using this CR. Edit this value
     commonName: mydomain.com # The main domain. Edit this value
     issuerRef:
       name: letsencrypt-staging # Name of the Issuer created in the previous step. Edit this value
       namespace: default
     dnsNames:
       - my.sub1.mydomain.com # Edit this value
   EOF
   ```

7. Create a Gateway CR. See the following example and modify values of the **spec.servers.tls.credentialName** and **spec.servers.hosts** parameters. For the **spec.servers.tls.credentialName** parameter, use the **spec.secretName** value of the Certificate CR. As the value of **spec.servers.hosts**, use the subdomain from the **spec.dnsName** parameter of the DNSEntry CR. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.istio.io/v1alpha3
   kind: Gateway
   metadata:
     name: httpbin-gateway
   spec:
     selector:
       istio: ingressgateway # Use Istio Ingress Gateway as deafult
     servers:
       - port:
           number: 443
           name: https-httpbin
           protocol: HTTPS
         tls:
           mode: SIMPLE
           credentialName: secret-name # Name of the Secret created in the Certificate CR in the previous step. Edit this value
         hosts:
           - "my.sub1.mydomain.com" # Edit this value
   EOF
   ```

8. Add your subdomain(s) at the end of the API Gateway **domain.allowlist** parameter. See the following example: `--domain-allowlist=my.sub1.mydomain.com`. Use a comma as a separator. Run the following command, to edit the deployment:

   ```bash
  kubectl edit deployment -n kyma-system api-gateway
  ```

   Press `i` to enter and `esc` to exit the interactive mode. Save the changes and exit the editor by pressing `:wq`.

9. Restart API Gateway. Run:
   
   ```bash
   kubectl rollout restart deployment {DEPLOYMENT_NAME}
   ```

   >**TIP:** To check the deployment name, run: `kubectl get deployments -A`

When you finish the setup, go to [this](./apix-01-expose-service-apigateway.md) tutorial to learn how to expose a service.
