---
title: Console access network error
type: Troubleshooting
---

If you try to access the Console of a local or a cluster Kyma deployment and your browser shows a 'Network Error', your machine doesn't have the Kyma self-signed TLS certificate added to your system's trusted certificate list.
Fix this by following one of these two approaches:

- Add the Kyma certificate to the trusted certificates list of your OS:

    <div tabs>
      <details>
      <summary>
      Minikube on MacOS
      </summary>

      ```
      sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain {PATH_TO_CERT}
      ```
      </details>
      <details>
      <summary>
      Minikube on Linux
      </summary>

      ```
      certutil -d sql:$HOME/.pki/nssdb -A -t "P,," -n {CERT_DISPLAYNAME} -i {PATH_TO_CERT}
      ```
      </details>
      <details>
      <summary>
      Cluster installation with xip.io
      </summary>

      Run this command after you install Kyma on your GKE or AKS cluster:
      ```
      tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
      && kubectl get configmap cluster-certificate-overrides -n kyma-installer -o jsonpath='{.data.global\.tlsCrt}' | base64 --decode > $tmpfile \
      && sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
      && rm $tmpfile
      ```
      </details>
    </div>

- Trust the certificate in your browser. Follow [this guide](https://stackoverflow.com/questions/7580508/getting-chrome-to-accept-self-signed-localhost-certificate) for Chrome or [this guide](https://origin-symwisedownload.symantec.com/resources/webguides/sslv/sslva_first_steps/Content/Topics/Configure/ssl_firefox_cert.htm) for Firefox. You must trust the certificate for the following addresses: `apiserver.foo.bar`, `console.foo.bar`, `dex.foo.bar`, `console-backend.foo.bar`.
  >**TIP:** This solution is suitable for users who don't have administrative access to the OS.
