---
title: Network error when accessing Busola
---

## Symptom

You're trying to access the Kyma Dashboard of a cluster Kyma deployment and your browser shows a 'Network Error'.

## Cause

Your local machine doesn't have the Kyma self-signed TLS certificate added to the system trusted certificate list.

## Remedy

To fix this, follow one of these two approaches:

### Configure your OS

If you have administrative access to your OS, add the Kyma certificate to the trusted certificates list of your OS.
After installing Kyma on your GKE or AKS cluster, run:

  ```bash
  tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
  && kubectl get secret kyma-gateway-certs -n istio-system -o jsonpath='{.data.tls\.crt}' | base64 --decode > $tmpfile \
  && sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
  && rm $tmpfile
  ```

### Configure your browser

If you don't have administrative access to the OS, trust the certificate in your browser.

- For Chrome, follow [this guide](https://stackoverflow.com/questions/7580508/getting-chrome-to-accept-self-signed-localhost-certificate).
- For Firefox, follow [this guide](https://javorszky.co.uk/2019/11/06/get-firefox-to-trust-your-self-signed-certificates/)

You must trust the certificate for the `apiserver.foo.bar` address.
