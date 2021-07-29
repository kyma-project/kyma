---
title: Console access network error
type: Troubleshooting
---

If you try to access the Console of a cluster Kyma deployment and your browser shows a 'Network Error', your local machine doesn't have the Kyma self-signed TLS certificate added to the system trusted certificate list.
To fix this, follow one of these two approaches:

1. Add the Kyma certificate to the trusted certificates list of your OS. Run this command after you install Kyma on your GKE or AKS cluster:

      ```
      tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
      && kubectl get secret kyma-gateway-certs -n istio-system -o jsonpath='{.data.tls\.crt}' | base64 --decode > $tmpfile \
      && sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
      && rm $tmpfile
      ```
   
2. Trust the certificate in your browser. Follow [this guide](https://stackoverflow.com/questions/7580508/getting-chrome-to-accept-self-signed-localhost-certificate) for Chrome or [this guide](https://javorszky.co.uk/2019/11/06/get-firefox-to-trust-your-self-signed-certificates/) for Firefox. You must trust the certificate for the `apiserver.foo.bar` address.

    >**TIP:** This solution is suitable for users who don't have administrative access to the OS.
