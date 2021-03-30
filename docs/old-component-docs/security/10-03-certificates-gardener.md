---
title: Issues with certificates on Gardener
type: Troubleshooting
---

During installation on Gardener, Kyma requests domain SSL certificates using the Gardener's [`Certificate`](https://gardener.cloud/documentation/guides/administer_shoots/request_cert/#request-a-certificate-via-certificate) custom resource to ensure secure communication through both Kyma UI and Kubernetes CLI.

This process can result in the following issues:

- `xip-patch` or `apiserver-proxy` installation takes too long.
- `Certificate is still not ready, status is {STATUS}. Exiting...` error occurs.
- Certificates are no longer valid.

If any of these issues appears, follow these steps:

1. Check the status of the Certificate resource:

    ```bash
    kubectl get certificates.cert.gardener.cloud --all-namespaces
    ```

2. If the status of any Certificate is `Error`, run:

    ```bash
    kubectl get certificates -n {CERTIFICATE_NAMESPACE} {CERTIFICATE_NAME} -o jsonpath='{ .status.message }'
    ```

The result describes the reason for the failure of issuing a domain SSL certificate. Depending on the moment when the error occurred, you can perform different actions.

<div tabs>
  <details>
  <summary>
  Error during the installation
  </summary>

1. Make sure the domain name provided in the `net-global-overrides` ConfigMap is proper and it meets the Gardener requirements.
2. Check if the `istio-ingressgateway` Service in the `istio-system` Namespace contains proper annotations:

    ```yaml
    dns.gardener.cloud/class=garden
    dns.gardener.cloud/dnsnames=*.{DOMAIN}
    ```
   
3. Check if the `apiserver-proxy-ssl` Service in the `kyma-system` Namespace contains proper annotations:
    
    ```yaml
    dns.gardener.cloud/class=garden
    dns.gardener.cloud/dnsnames=apiserver.{DOMAIN}
    ```

  </details>
  <details>
  <summary>
  Error after the installation
  </summary>

You can create a new Certificate resource applying suggestions from the error message to request a new domain SSL certificate. Follow these steps:

1. Make sure the Secret connected to the Certificate resource is not present on the cluster. To find its name and Namespace, run:

    ```bash
    kubectl get certificates -n {CERTIFICATE_NAMESPACE} {CERTIFICATE_NAME} -o jsonpath='{ .spec.secretRef }'
    ```

2. Delete the incorrect Certificate from the cluster.

3. Apply the fixed Certificate.

>**NOTE:** If you upgrade Kyma, you may need to perform steps from **Error during the installation** tab.

  </details>
</div>
