---
title: Get the client certificate
type: Getting Started
---

After you create a Remote Environment (RE) in Kyma, it's time to connect it with an external solution, which allows to consume external APIs and Event catalogs of this solution. To accomplish this you must get the client certificate for the external solution and register its services.

This guide shows you how to get the client certificate.

## Prerequisites

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
  curl -X POST http://localhost:8080/v1/remoteenvironments/{RE_NAME}/tokens
  ```

A successful call returns the following response:
  ```
  {
    "url":"{CONFIGURATION_URL_WITH_TOKEN}",
    "token":"example-token-123"
  }
  ```

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
        "metadataUrl":      "https://gateway.{CLUSTER_DOMAIN}/{RE_NAME}/v1/metadata/services",
        "eventsUrl":        "https://gateway.{CLUSTER_DOMAIN}/{RE_NAME}/v1/events",
        "certificatesUrl":  "https://connector-service.{CLUSTER_DOMAIN}/v1/remoteenvironments/{RE_NAME}",
    },
    "certificate":{
        "subject":"OU=Test,O=Test,L=Blacksburg,ST=Virginia,C=US,CN={RE_NAME}",
        "extensions": "",
        "key-algorithm": "rsa2048",
    }
}
```

When you connect an external solution to a local Kyma deployment, you must set the NodePort of the `application-connector-nginx-ingress-controller` for the Metadata Service and for the Event Service.

- To get the NodePort, run:
  ```
  kubectl -n kyma-system get svc application-connector-nginx-ingress-controller -o 'jsonpath={.spec.ports[?(@.port==443)].nodePort}'
  ```
- Set it for the Metadata Service and the Event Service using these calls:
  ```
  curl https://gateway.kyma.local:{NODE_PORT}/{RE_NAME}/v1/metadata/services --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
  ```
  ```
  curl https://gateway.kyma.local:{NODE_PORT}/{RE_NAME}/v1/events --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
  ```

## Generate a CSR and send it to Kyma

Generate a CSR using the values obtained in the previous step:
```
openssl genrsa -out generated.key 2048
openssl req -new -sha256 -out generated.csr -key generated.key -subj "/OU=OrgUnit/O=Organization/L=Waldorf/ST=Waldorf/C=DE/CN={RE_NAME}"
openssl base64 -in generated.csr
```

Send the encoded CSR to Kyma. Run:

```
curl -H "Content-Type: application/json" -d '{"csr":"BASE64_ENCODED_CSR_HERE"}' https://connector-service.{CLUSTER_DOMAIN}/v1/remoteenvironments/{RE_NAME}/client-certs?token=example-token-456
```

The response contains a valid client certificate signed by the Kyma Certificate Authority:
```
{
    "crt":"BASE64_ENCODED_CRT"
}
```

After you receive the certificate, decode it and use it in your application. Register the services of your external solution through the Metadata Service.
