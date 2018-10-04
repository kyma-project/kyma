---
title: Get the client certificate
type: Getting Started
---

After you create a Remote Environment (RE) in Kyma, it's time to connect it with an external solution, which allows to consume external APIs and Event catalogs of this solution. To accomplish this you must get the client certificate for the external solution and register its services.

This guide shows you how to get the client certificate.

## Prerequisites

- You need a private RSA key for your external solution. If your solution doesn't have one, generate it using this command:
  ```
  openssl genrsa -out generated.key 4096
  ```

- [OpenSSL toolkit](https://www.openssl.org/docs/man1.0.2/apps/openssl.html) to create a Certificate Signing Request (CSR), keys, and certificates which fulfil high security standards.

## Get the configuration URL with a token

Get the configuration URL with a token which allows you to get Kyma CSR configuration and URLs in Kyma required to connect your external solution to a created Remote Environment.
Follow this steps to get it using the CLI:

- Expose the Connector Service outside of Kubernetes using `kubectl port-forward`:
  ```
  kubectl -n=kyma-integration port-forward svc/connector-service-internal-api 8080:8080
  ```

- Make a POST request to the `tokens` endpoint:
  ```
  curl -X POST http://localhost:8080/v1/remoteenvironments/{remote-environment-name}/tokens
  ```

A successful call returns the following response:
  ```
  {
    "url":"{CONFIGURATION_URL_WITH_TOKEN}",
    "token":"example-token-123"
  }
  ```

Alternatively, use the UI:

  - Go to the Kyma console UI.
  - Select **Administration**.
  - Select the **Remote Environments** from the **Integration** section.
  - Choose the Remote Environment to which you want to connect the external solution.
  - Click **Connect Remote Environment**.
  - Copy the token by clicking **Copy to clipboard**.

## Get the CSR information and configuration details from Kyma

Use the link you got in the previous step to fetch the CSR information and configuration details required to connect your external solution. Run:

```
curl {CONFIGURATION_URL_WITH_TOKEN}
```
>**NOTE:** The URL you call in this step contains a token that is valid for a single call. If you need to get the configuration details once again, generate a new configuration URL with a valid token and call it again. You get a code `403` error if you call the same configuration URL more than once.

A successful call returns the following response:
```
{
    "csrUrl": "{CSR_SIGNING_URL_WITH_TOKEN}",
    "api":{
        "metadataUrl":      "https://gateway.{CLUSTER_NAME}.kyma.cluster.cx/{REMOTE-ENVIRONMENT-NAME}/v1/metadata/services",
        "eventsUrl":        "https://gateway.{CLUSTER_NAME}.kyma.cluster.cx/{REMOTE-ENVIRONMENT-NAME}/v1/events",
        "certificatesUrl":  "https://connector-service.{CLUSTER_NAME}.kyma.cluster.cx/v1/remoteenvironments/{RE_NAME}",
    },
    "certificate":{
        "subject":"OU=Test,O=Test,L=Blacksburg,ST=Virginia,C=US,CN={RE_NAME}",
        "extensions": "",
        "key-algorithm": "rsa2048",
    }
}
```

## Generate a CSR and send it to Kyma

Generate a CSR using the values obtained in the previous step:
```
openssl req -new -out generated.csr -key generated.key -subj "/OU=OrgUnit/O=Organization/L=Waldorf/ST=Waldorf/C=DE/CN={RE_NAME}"
```

After the CSR is created, encode it with Base64. Send the encoded CSR to Kyma. Run:
```
curl -H "Content-Type: application/json" -d '{"csr":"{BASE64_ENCODED_CSR_HERE}"}' {CSR_SIGNING_URL_WITH_TOKEN}
```

The response contains a valid client certificate signed by the Kyma Certificate Authority:
```
{
    "crt":"BASE64_ENCODED_CRT"
}
```

After you receive the certificate, register the services of your external solution using the Metadata API.
