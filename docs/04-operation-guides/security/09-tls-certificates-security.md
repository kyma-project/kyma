---
title: Update TLS certificate
type: Security
---

The TLS certificate is a vital security element. Follow this tutorial to update the TLS certificate in Kyma.

>**NOTE:** This procedure can interrupt the communication between your cluster and the outside world for a limited period of time.

## Prerequisites

 - New TLS certificate and key for custom domain deployments, base64-encoded
 - `kubeconfig` file generated for the Kubernetes cluster that hosts the Kyma instance

## Steps

<div tabs>
  <details>
  <summary>
  Custom domain certificate
  </summary>

  >**CAUTION:** When you regenerate the TLS certificate for Kyma, the `kubeconfig` file generated through the Console UI becomes invalid. To complete these steps, use the admin `kubeconfig` file generated for the Kubernetes cluster that hosts the Kyma instance you're working on.

  1. Delete the `net-global-overrides` ConfigMap.
      ```bash
      kubectl delete cm -n kyma-installer net-global-overrides
      ```

  2. Export the Kyma version, your domain, new certificate and key as the environment variables.
      ```bash
      export KYMA_VERSION={KYMA_RELEASE_VERSION}
      export DOMAIN={YOUR_DOMAIN}
      export TLS_CERT={YOUR_NEW_CERTIFICATE}
      export TLS_KEY={YOUR_NEW_KEY}
      ```

  3. Trigger the update process. Run:
      ```bash
      kyma upgrade -s $KYMA_VERSION --domain $DOMAIN --tls-cert $TLS_CERT --tls-key $TLS_KEY
      ```
    
      The process is complete when you see the `Kyma installed` message.

  4. Restart the Console Backend Service to propagate the new certificate. Run:
      ```bash
      kubectl delete pod -n kyma-system -l app=backend
      ```

  </details>

  <details>
  <summary>
  Self-signed certificate
  </summary>

  The self-signed TLS certificate used in Kyma instances deployed with `xip.io` is valid for 30 days. If the self-signed certificate expired for your cluster and you can't, for example, log in to the Kyma Console, regenerate the self-signed certificate.

  >**CAUTION:** When you regenerate the TLS certificate for Kyma, the `kubeconfig` file generated through the Console UI becomes invalid. To complete these steps, use the admin `kubeconfig` file generated for the Kubernetes cluster that hosts the Kyma instance you're working on.

  1. Delete the ConfigMap and the Secret that stores the expired Kyma TLS certificate. Run:
      ```bash
      kubectl delete cm -n kyma-installer net-global-overrides ; kubectl delete secret -n kyma-system apiserver-proxy-tls-cert
      ```

  2. Trigger the update process. Run:
      ```bash
      kubectl -n default label installation/kyma-installation action=install
      ```

      To watch the progress of the update, run:
      ```bash
      while true; do \
      kubectl -n default get installation/kyma-installation -o jsonpath="{'Status: '}{.status.state}{', description: '}{.status.description}"; echo; \
      sleep 5; \
      done
      ```

      The process is complete when you see the `Kyma installed` message.

  3. Restart the Console Backend Service to propagate the new certificate. Run:
      ```bash
      kubectl delete pod -n kyma-system -l app=backend
      ```

  4. Add the newly generated certificate to the trusted certificates of your OS. For MacOS, run:
      ```bash
      tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
      && kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\.ingress\.tlsCrt}' | base64 --decode > $tmpfile \
      && sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
      && rm $tmpfile
      ```

  </details>

</div>
