---
title: Update TLS certificate
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

  1. Export the Kyma version, your domain, new certificate and key as the environment variables.

      ```bash
      export KYMA_VERSION={KYMA_RELEASE_VERSION}
      export DOMAIN={YOUR_DOMAIN}
      export TLS_CERT={YOUR_NEW_CERTIFICATE}
      export TLS_KEY={YOUR_NEW_KEY}
      ```

  2. Trigger the update process. Run:

      ```bash
      kyma deploy -s $KYMA_VERSION --domain $DOMAIN --tls-cert $TLS_CERT --tls-key $TLS_KEY
      ```
    
      The process is complete when you see the `Kyma installed` message.

  </details>

  <details>
  <summary>
  Self-signed certificate
  </summary>

  The self-signed TLS certificate used in Kyma instances deployed with the default `kyma.example.com` domain is valid for 30 days. If the self-signed certificate expired for your cluster, regenerate the self-signed certificate. 

    1. Export the Kyma version as the environment variable.

      ```bash
      export KYMA_VERSION={KYMA_RELEASE_VERSION}
      ```

    2. Delete the Secret that stores the expired Kyma TLS certificate. Run:

      ```bash
      kubectl delete secret -n istio-system kyma-gateway-certs
      ```

    3. Trigger the update process. Run:

      ```bash
      kyma deploy -s $KYMA_VERSION
      ```

      The process is complete when you see the `Kyma installed` message.

    4. Add the newly generated certificate to the trusted certificates of your OS. For MacOS, run:
  
      ```bash
      tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
      && kubectl get secret kyma-gateway-certs -n istio-system -o jsonpath='{.data.tls\.crt}' | base64 --decode > $tmpfile \
      && sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
      && rm $tmpfile
      ```

  </details>

</div>
