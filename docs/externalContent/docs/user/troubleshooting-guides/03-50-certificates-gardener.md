# Issues with Certificates on Gardener

## Symptom & Cause

During installation on Gardener, Kyma requests domain SSL certificates using the Gardener's [Certificate](https://github.com/gardener/cert-management#requesting-a-certificate) custom resource (CR) to ensure secure communication through both Kyma UI and Kubernetes CLI.

This process can result in the following issues:

- Certificates installation takes too long.
- `Certificate is still not ready, status is {STATUS}. Exiting...` error occurs.
- Certificates are no longer valid.

## Solution

If any of these issues appears, follow these steps:

1. Check the status of the Certificate CR:

    ```bash
    kubectl get certificates.cert.gardener.cloud --all-namespaces
    ```

2. If the status of any Certificate is `Error`, run:

    ```bash
    kubectl get certificates -n {CERTIFICATE_NAMESPACE} {CERTIFICATE_NAME} -o jsonpath='{ .status.message }'
    ```

The result describes the reason for the failure of issuing a domain SSL certificate. Depending on the moment when the error occurred, you can perform different actions.

<!-- tabs:start -->
#### Error During the Installation

1. Make sure the provided domain name is proper and meets the Gardener requirements.

2. Check if the `istio-ingressgateway` Service in the `istio-system` namespace contains proper annotations:

    ```yaml
    dns.gardener.cloud/class=garden
    dns.gardener.cloud/dnsnames=*.{DOMAIN}
    ```

#### Error After the Installation

You can create a new Certificate resource applying suggestions from the error message to request a new domain SSL certificate. Follow these steps:

1. Make sure the Secret connected to the Certificate resource is not present in the cluster. To find its name and namespace, run:

    ```bash
    kubectl get certificates -n {CERTIFICATE_NAMESPACE} {CERTIFICATE_NAME} -o jsonpath='{ .spec.secretRef }'
    ```

2. Delete the incorrect Certificate from the cluster.

3. Apply the fixed Certificate.   
<!-- tabs:end -->