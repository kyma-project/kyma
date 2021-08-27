---
title: Use a custom domain to expose a service
---

This tutorial shows how to set up your custom domain and prepare a certificate to use the domain for exposing a service. The components used are Gardener [External DNS Management](https://gardener.cloud/docs/concepts/networking/dns-managment/#external-dns-management) and [Certificate Management](https://gardener.cloud/docs/concepts/networking/cert-managment/).

To learn how to expose a service, go to the [**Expose a service**](./apix-01-expose-service-apigateway.md) tutorial.

TO follow this tutorial, use Kyma 2.0 or higher.

## Prerequisites

If you use a cluster not managed by Gardener, install the required components manually. Follow these steps:

1. Create a Namespace. Run:

   ```bash
   kubectl create ns gardener-components
   ```

2. Download the [External DNS Management](https://github.com/gardener/external-dns-management) and [Certificate Management](https://github.com/gardener/cert-management) projects locally. Run:

   ```bash
   git clone --branch v0.10.4 https://github.com/gardener/external-dns-management.git
   git clone --branch v0.8.3 https://github.com/gardener/cert-management.git
   ```

   > **NOTE:** This tutorial uses the External DNS Management v.0.10.4 and the Certificate Management v0.8.3.

3. Enable vertical Pod autoscaling on your cluster. For example, for Google Cloud Platform, run:

   ```bash
   gcloud container clusters update {CLUSTER_NAME} --enable-vertical-pod-autoscaling --zone="{CLUSTER_ZONE}" --project="{PROJECT_NAME}"
   ```

4. Install the `external-dns-management` component in the `gardener-components` Namespace:

   ```bash
   helm install external-dns {PATH_TO_EXTERNAL_DNS_MANAGEMENT}/charts/external-dns-management --namespace=gardener-components --set configuration.identifier=external-dns-identifier --set configuration.disableNamespaceRestriction=true
   ```

5. Install the `cert-management` component in the `gardener-components` Namespace:

   ```bash
   helm install cert-manager {PATH_TO_CERT_MANAGEMENT}/charts/cert-management --namespace=gardener-components --set configuration.identifier=cert-manager-identifier
   ```

6. Add the following RBAC rules to allow the Certificate Management component to configure objects. Create an `istio-systen` Namespace, a ClutserRole and a RoleBinding:

    ```bash
    kubectl create ns istio-system
    ```

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: cert-controller-manager-dnsentries
    rules:
    - apiGroups:
      - dns.gardener.cloud
      resources:
      - dnsentries
      verbs:
      - get
      - list
      - update
      - watch
      - create
      - delete
    EOF
    ```

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: rbac.authorization.k8s.io/v1
    kind: RoleBinding
    metadata:
      name: cert-controller-manager-dnsentries
      namespace: istio-system
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: cert-controller-manager-dnsentries
    subjects:
    - kind: ServiceAccount
      name: cert-controller-manager
      namespace: gardener-components
    EOF
    ```

## Steps

Follow these steps to set up your custom domain and prepare a certificate required to expose a service.

1. [Install Kyma](../../04-operation-guides/operations/01-install-kyma.md) 2.0 or higher.

2. Create a Namespace. Run:

   ```bash
   kubectl create ns {NAMESPACE_NAME}
   ```

3. Create a Secret containing credentials for your DNS cloud service provider account. See the [official External DNS Management docuemntation](https://github.com/gardener/external-dns-management/blob/master/README.md#external-dns-management), choose your DNS cloud service provider, and follow the relevant guidelines. Remember to edit the **metadata.namespace** parameter and provide your {NAMESPACE_NAME} as the required value.  Once the YAML file with the relevant parameters is ready, create a Secret Custom Resource (CR). Run the following command:

   ```bash
   kubectl apply -n {NAMESPACE_NAME} -f {SECRET}.yaml
   ```

4. Set up the `external-dns-management` component.

- Create a DNSProvider CR. Export values of the **metadata.namespace**, [**spec.type**](https://github.com/gardener/external-dns-management#using-the-dns-controller-manager), **metadata.spec.secretRef.name** and **metadata.spec.domains.include** parameters as environment variables. As the value of **spec.type**, use the relevant provider type. See the [official Gardener examples](https://github.com/gardener/external-dns-management/tree/master/examples) of the DNSProvider CR. For the **spec.secretRef.name** parameter, use the **metadata.name** value from your `{SECRET}.yaml`. The domain you provide is the one that you own, for example `mydomain.com`. In the next steps you provide a subdomain of this domain, for example `api.mydomain.com`. Run:

  ```bash
  export NAMESPACE={NAMESPACE_NAME}
  export SPEC_TYPE={PROVIDER_TYPE}
  export SECRET={SECRET_NAME}
  export DOMAIN={CLUSTER_DOMAIN} #e.g. mydomain.com
  ```
  
  Create a DNSProvider CR. Run:

  <div tabs>
  <details>
  <summary>
  Gardener cluster
  </summary>

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

  </details>
  <details>
  <summary>
  Non-Gardener cluster
  </summary>

  ```bash
  cat <<EOF | kubectl apply -f -
  apiVersion: dns.gardener.cloud/v1alpha1
  kind: DNSProvider
  metadata:
    name: dns-provider
    namespace: $NAMESPACE
  spec:
    type: $SPEC_TYPE
    secretRef:
      name: $SECRET
    domains:
      include:
        - $DOMAIN
  EOF
  ```

  </details>
  </div>

- Create a DNSEntry CR. Export values of the **metadata.namespace**, **metadata.spec.dnsName**, and **metadata.spec.targets.IP** parameters as environment variables. Optionally, you can also change the value of **metadata.spec.ttl**. Run:

  ```bash
  export NAMESPACE={NAMESPACE_NAME}
  export WILDCARD={WILDCRAD_SUBDOMAIN} #e.g. *.api.mydomain.com
  export IP=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}') # assuming only one LoadBalancer with external IP
  ```

 Create a DNSEntry CR. Run:

  <div tabs>
  <details>
  <summary>
  Gardener cluster
  </summary>

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

  </details>
  <details>
  <summary>
  Non-Gardener cluster
  </summary>

  ```bash
  cat <<EOF | kubectl apply -f -
  apiVersion: dns.gardener.cloud/v1alpha1
  kind: DNSEntry
  metadata:
    name: dns-entry
    namespace: $NAMESPACE
  spec:
    dnsName: "$WILDCARD"
    ttl: 600
    targets:
      - $IP
  EOF
  ```

  </details>
  </div>

  >**NOTE:** You can create many DNSEntry CRs for one DNSProvider, depending on the number of subdomains you want to use. To simplify your setup, consider using a wildcard subdomain if all your DNSEntry objects share the same subdomain and resolve to the same IP, for example: `*.api.mydomain.com`. Remember that such a wildcard entry results in DNS configuration that doesn't support the following hosts: `api.mydomain.com` and `mydomain.com`. We don't use these hosts in this tutorial, but you can add DNS Entries for them if you need.

5. Create an Issuer CR. Export values of the **metadata.spec.acme.email**, **metadata.spec.domains.include**, and **metadata.spec.domains.exclude** parameters as environment variables. As the values for the **metadata.spec.domains.include** parameter, use the main domain, the subdomain from the **spec.dnsName** parameter of the DNSEntry CR, and a wildcard DNS record. Run:

   ```bash
   export EMAIL={YOUR_EMAIL_ADDRESS}
   export DOMAIN={CLUSTER_DOMAIN} #e.g. mydomain.com
   export SUBDOMAIN={YOUR_SUBDOMAIN} #e.g. api.mydomain.com
   export WILDCARD={WILDCARD_SUBDOMAIN} #e.g. *.api.mydomain.com
   ```

   Create an Issuer CR. Run:

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

   > **NOTE:** The Issuer CR  must be created in the `default` Namespace.

6. Create a Certificate CR. Export values of the **metadata.namespace**, **spec.secretName**, **spec.commonName**, and **spec.issuerRef.name**, and **spec.dnsNames** parameters as environment variables. As the value for the **spec.issuerRef.name** parameter, use the value from te **metadata.name** parameter of the Issuer CR. Run:

   ```bash
   export TLS_SECRET={SECRET_NAME} #e.g. httpbin-tls-credentials
   export DOMAIN={CLUSTER_DOMAIN} #e.g. mydomain.com
   export ISSUER={ISSUER_NAME}
   export WILDCARD={WILDCARD_SUBDOMAIN} #e.g. *.api.mydomain.com
   ```

   Create a Certificate CR. Run:

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
   EOF
   ```

   > **NOTE:** The Certificte CR and the Secret with the TLS certificate are created in the same Namespace. Additionally, Istio requires the Secret to be stored in the `istio-system` Namespace so that the Secret could be used for HTTPS traffic. As a result the Certificate must be also created in the `istio-system` Namespace.

7. Create a Gateway CR. Export values of the **spec.servers.tls.credentialName** and **spec.servers.hosts** parameters as anvironment variables. For the **spec.servers.tls.credentialName** parameter, use the **spec.secretName** value of the Certificate CR. As the value of **spec.servers.hosts**, use the wildcard DNS record. Run:

   ```bash
   export NAMESPACE={NAMESPACE_NAME}
   export TLS_SECRET={SECRET_NAME}
   export WILDCARD={WILDCARD_SUBDOMAIN} # e.g. *api.mydomain.com
   ```

   Create a Gateway CR. Run:

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
           - "$WILDCARD"
   EOF
   ```

8. Add your subdomain(s) at the end of the API Gateway **domain.allowlist** parameter. See the following example: `--domain-allowlist=api.mydomain.com`. Use a comma as a separator. Run the following command, to edit the deployment:

   ```bash
   kubectl edit deployment -n kyma-system api-gateway
   ```

   >**TIP:** To avoid adding every subdomain to the API Gateway **domain.allowlist** parameter, disable the allowlist mechanism. Override the value of the **config.enableDomainAllowList** parameter in the API Gateway chart by changing its value to `false`. For more details on overrides, see how to [change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md).

When you finish the setup, go to [this](./apix-01-expose-service-apigateway.md) tutorial to learn how to expose a service.

If you want to expose a different workload using a different domain, repeat steps from 2 to 8 with the new details.
