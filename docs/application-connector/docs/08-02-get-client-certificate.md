---
title: Get the client certificate
type: Tutorials
---

After you create an Application (App), connect it to an external solution to consume the solution's APIs and Event catalogs in Kyma. To accomplish this, get the client certificate for the external solution and register its services.

This guide shows you how to get the client certificate.

## Prerequisites

- [OpenSSL toolkit](https://www.openssl.org/docs/man1.0.2/apps/openssl.html) to create a Certificate Signing Request (CSR), keys, and certificates which fulfil high security standards.

## Get the configuration URL with a token

To get the configuration URL which allows you to fetch the required configuration details, create a TokenRequest custom resource (CR). The controller which handles this CR kind adds the **status** section to the created CR. The **status** section contains the required configuration details.

- Create a TokenRequest CR. The CR name must match the name of the App for which you want to get the configuration details. Run:
  ```
  cat <<EOF | kubectl apply -f -
  apiVersion: applicationconnector.kyma-project.io/v1alpha1
  kind: TokenRequest
  metadata:
    name: {APP_NAME}
  EOF
  ```

- Fetch the TokenRequest CR you created to get the configuration details from the **status** section. Run:
  ```
  kubectl get tokenrequest.applicationconnector.kyma-project.io {APP_NAME} -o yaml
  ```
  >**NOTE:** If the response doesn't contain the **status** section, wait for a few moments and fetch the CR again.

A successful call returns the following response:
  ```
  apiVersion: applicationconnector.kyma-project.io/v1alpha1
  kind: TokenRequest
  metadata:
    name: {APP_NAME}
  status:
    expireAfter: 2018-11-22T18:38:44Z
    application: {APP_NAME}
    state: OK
    token: h31IwJiLNjnbqIwTPnzLuNmFYsCZeUtVbUvYL2hVNh6kOqFlW9zkHnzxYFCpCExBZ_voGzUo6IVS_ExlZd4muQ==
    url: https://connector-service.kyma.local/v1/applications/test/info?token=h31IwJiLNjnbqIwTPnzLuNmFYsCZeUtVbUvYL2hVNh6kOqFlW9zkHnzxYFCpCExBZ_voGzUo6IVS_ExlZd4muQ==
  ```

## Get the CSR information and configuration details from Kyma

Use the link you got in the previous step to fetch the CSR information and configuration details required to connect your external solution. Run:

```
curl {CONFIGURATION_URL_WITH_TOKEN}
```
>**NOTE:** The URL you call in this step contains a token that is valid for 5 minutes or for a single call. You get a code `403` error if you call the same configuration URL more than once, or if you call an URL with an expired token.

A successful call returns the following response:
```
{
    "csrUrl": "{CSR_SIGNING_URL_WITH_TOKEN}",
    "api":{
        "metadataUrl":      "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services",
        "eventsUrl":        "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events",
        "certificatesUrl":  "https://connector-service.{CLUSTER_DOMAIN}/v1/applications/{APP_NAME}",
    },
    "certificate":{
        "subject":"OU=Test,O=Test,L=Blacksburg,ST=Virginia,C=US,CN={APP_NAME}",
        "extensions": "",
        "key-algorithm": "rsa2048",
    }
}
```

When you connect an external solution to a local Kyma deployment, you must set the NodePort of the `application-connector-nginx-ingress-controller` for the Application Registry and for the Event Service.

- To get the NodePort, run:
  ```
  kubectl -n kyma-system get svc application-connector-nginx-ingress-controller -o 'jsonpath={.spec.ports[?(@.port==443)].nodePort}'
  ```
- Set it for the Application Registry and the Event Service using these calls:
  ```
  curl https://gateway.kyma.local:{NODE_PORT}/{APP_NAME}/v1/metadata/services --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
  ```
  ```
  curl https://gateway.kyma.local:{NODE_PORT}/{APP_NAME}/v1/events --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
  ```

## Generate a CSR and send it to Kyma

Generate a CSR using the values obtained in the previous step:
```
openssl genrsa -out generated.key 2048
openssl req -new -sha256 -out generated.csr -key generated.key -subj "/OU=OrgUnit/O=Organization/L=Waldorf/ST=Waldorf/C=DE/CN={APP_NAME}"
openssl base64 -in generated.csr
```

Send the encoded CSR to Kyma. Run:

```
curl -H "Content-Type: application/json" -d '{"csr":"BASE64_ENCODED_CSR_HERE"}' https://connector-service.{CLUSTER_DOMAIN}/v1/applications/{APP_NAME}/client-certs?token=example-token-456
```

The response contains a valid client certificate signed by the Kyma Certificate Authority:
```
{
    "crt":"BASE64_ENCODED_CRT"
}
```

After you receive the certificate, decode it and use it in your application. Register the services of your external solution through the Application Registry.
