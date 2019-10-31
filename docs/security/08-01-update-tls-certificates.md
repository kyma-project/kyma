---
title: Update TLS certificate
type: Tutorials
---

The TLS certificate is a vital security element. Follow this tutorial to update the TLS certificate in Kyma.

>**TIP:** The self-signed certificate used in Kyma instances deployed with `xip.io` is valid for 30 days. If the self-signed certificate expired for your cluster and you can't, for example, log in to the Kyma Console, follow this tutorial to regenerate the certificate.

## Steps

>**CAUTION:** When you regenerate the SSL certificate for Kyma, the kubeconfig file generated through the Console UI becomes invalid. To complete these steps, use the admin kubeconfig file generated for the Kubernetes cluster that hosts the Kyma instance you're working on.

1. Delete the ConfigMap and the Secret that stores the expired Kyma SSL certificate. Run:

  ```
  kubectl delete cm -n kyma-installer net-global-overrides ; kubectl delete secret -n kyma-system apiserver-proxy-tls-cert
  ```

2. Trigger the update process to generate a new certificate. Run:

  ```
  kubectl label installation/kyma-installation action=install
  ```

  To watch the progress of the update, run:

  ```
  while true; do \
  kubectl -n default get installation/kyma-installation -o jsonpath="{'Status: '}{.status.state}{', description: '}{.status.description}"; echo; \
  sleep 5; \
  done
  ```

  The process is complete when you see the `Kyma installed` message.

3. Restart the Console Backend Service to propagate the new certificate. Run:

  ```
  kubectl delete pod -n kyma-system -l app=console-backend-service
  ```

4. Add the newly generated certificate to the trusted certificates of your OS. For MacOS, run:

  ```
  tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
  && kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\.ingress\.tlsCrt}' | base64 --decode > $tmpfile \
  && sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
  && rm $tmpfile
  ```
