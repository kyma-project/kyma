---
title: Basic troubleshooting
type: Troubleshooting
---

## 403 Forbidden in the Console

If you log to the Console and get 403 Forbidden error, do the following:

  1. Fetch the ID Token. For example, use the [Chrome Developer Tools](https://developers.google.com/web/tools/chrome-devtools) and search for the token in sent requests.
  2. Decode the ID Token. For example, use the [jwt.io](https://jwt.io/) page.
  3. Check if the token contains groups claims:
      ```
      "groups": [
        "example-group"
      ]
      ```
  4. Make sure the group you are assigned to has [permissions](https://kyma-project.io/docs/components/security/#details-roles-in-kyma) to view resources you requested.
    
## Problems with certificates on Gardener

During installation on Gardener Kyma requests domain SSL certificates using Gardener's custom resource `Certificate` to grant a secure communication using both Kyma UI and Kubernetes CLI. When you experience:
 - `xip-patch` installation takes too long or an error occurs `Certificate is still not ready, status is {STATUS}. Exiting...`,
 - you notice any issues regarding certificates validity,
  
  follow the steps:

1. Check the status of the Certificate resource. Run:

    ```kubectl get certificates.cert.gardener.cloud --all-namespaces```

2. If status of any Certificate is `Error`, run:

    ```kubectl get certificates -n {CERTIFICATE_NAMESPACE} {CERTIFICATE_NAME} -o jsonpath='{ .status.message }'```

The result describes the reason of failure of issuing a domain SSL certificate. You can create a new Certificate resource applying suggestion from the error message to request a new domain SSL certificate. Follow this steps:
    
1. Make sure the secret connected to the Certificate resource is not present on the cluster. To find its name and namespace, run:

    ```kubectl get certificates -n {CERTIFICATE_NAMESPACE} {CERTIFICATE_NAME} -o jsonpath='{ .spec.secretRef }'```
    
2. Delete the Certificate from the cluster.

3. Apply fixed Certificate.

