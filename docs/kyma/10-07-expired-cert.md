---
title: Expired SSL certificate
type: Troubleshooting
---

The self-signed SSL certificate used in Kyma instances deployed with xip.io is valid for 30 days. If the self-signed certificate expired for your cluster and you can't, for example, log in to the Kyma Console, follow these steps to regenarte the certificate.

1. Delete the ConfigMap that stores the current SSL certificate. Run:

  ```
  kublectl delete cm -n kyma-installer net-global-overrides
  ```

2. Trigger the update process to generate a new certificate and propagate it to the components that use it. Run:

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

3. Add the newly generated certificate to your system's trusted certificates. Run:

  ```
  tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
  && kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\.ingress\.tlsCrt}' | base64 --decode > $tmpfile \
  && sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
  && rm $tmpfile
  ```

4. Access your cluster's Console under the `https://console.{CLUSTER_DOMAIN}` address.
