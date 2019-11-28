---
title: Get the client certificate
type: Tutorials
---

After you create an Application, connect it to an external solution to consume the solution's APIs and event catalogs in Kyma. To accomplish this, get the client certificate for the external solution and register its services.

This guide shows you how to get the client certificate.

>**NOTE:** The client certificate is valid for 92 days. See [this](#tutorials-renew-the-client-certificate) tutorial to learn how to renew the client certificate. 
You can also revoke the client certificate, which prevents it from being renewed. See [this](#tutorials-revoke-the-client-certificate) tutorial to learn how to do this.

## Prerequisites

- [OpenSSL toolkit](https://www.openssl.org/docs/man1.0.2/apps/openssl.html) to create a Certificate Signing Request (CSR), keys, and certificates which meet high security standards

## Get the configuration URL with a token

To get the configuration URL which allows you to fetch the required configuration details, create a TokenRequest custom resource (CR). The controller which handles this CR kind adds the **status** section to the created CR. The **status** section contains the required configuration details.

- Create a TokenRequest CR. The CR name must match the name of the Application for which you want to get the configuration details. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: applicationconnector.kyma-project.io/v1alpha1
   kind: TokenRequest
   metadata:
     name: {APP_NAME}
   EOF
   ```

- Fetch the TokenRequest CR you created to get the configuration details from the **status** section. Run:

   ```bash
   kubectl get tokenrequest.applicationconnector.kyma-project.io {APP_NAME} -o yaml
   ```
   
   >**NOTE:** If the response doesn't contain the **status** section, wait for a few moments and fetch the CR again.

   A successful call returns the following response:
  
   ```yaml
   apiVersion: applicationconnector.kyma-project.io/v1alpha1
   kind: TokenRequest
   metadata:
     name: {APP_NAME}
   status:
     expireAfter: 2018-11-22T18:38:44Z
     application: {APP_NAME}
     state: OK
     token: h31IwJiLNjnbqIwTPnzLuNmFYsCZeUtVbUvYL2hVNh6kOqFlW9zkHnzxYFCpCExBZ_voGzUo6IVS_ExlZd4muQ==
     url: https://connector-service.kyma.local/v1/applications/signingRequests/info?token=h31IwJiLNjnbqIwTPnzLuNmFYsCZeUtVbUvYL2hVNh6kOqFlW9zkHnzxYFCpCExBZ_voGzUo6IVS_ExlZd4muQ==
   ```

## Get the CSR information and configuration details from Kyma

Use the link you got in the previous step to fetch the CSR information and configuration details required to connect your external solution. Run:

```bash
curl {CONFIGURATION_URL_WITH_TOKEN}
```
>**NOTE:** The URL you call in this step contains a token that is valid for 5 minutes or for a single call. You get a code `403` error if you call the same configuration URL more than once, or if you call a URL with an expired token.

A successful call returns the following response:

```json
{
    "csrUrl": "{CSR_SIGNING_URL_WITH_TOKEN}",
    "api":{
        "metadataUrl":      "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services",
        "eventsUrl":        "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events",
        "infoUrl":          "https://gateway.{CLUSTER_DOMAIN}/v1/applications/management/info",
        "certificatesUrl":  "https://connector-service.{CLUSTER_DOMAIN}/v1/applications/certificates",
    },
    "certificate":{
        "subject":"OU=Test,O=TestOrg,L=Waldorf,ST=Waldorf,C=DE,CN={APP_NAME}",
        "extensions": "",
        "key-algorithm": "rsa2048",
    }
}
```

> **NOTE:** The response contains URLs to the Application Registry API and the Events Service API, however, it is not recommended to use them. You should call the `metadata` endpoint URL, which is provided in `api.infoUrl` property, to fetch correct URLs to the Application Registry API and to the Events Service API, and other configuration details.

## Generate a CSR and send it to Kyma

Generate a CSR using the certificate subject data obtained in the previous step:

```bash
openssl genrsa -out generated.key 2048
openssl req -new -sha256 -out generated.csr -key generated.key -subj "/OU=Test/O=TestOrg/L=Waldorf/ST=Waldorf/C=DE/CN={APP_NAME}"
openssl base64 -in generated.csr
```

Send the encoded CSR to Kyma. Run:

```bash
curl -H "Content-Type: application/json" -d '{"csr":"BASE64_ENCODED_CSR_HERE"}' {CSR_SIGNING_URL_WITH_TOKEN}
```

The response contains a valid client certificate signed by the Kyma Certificate Authority (CA).

```json
{
    "crt":"BASE64_ENCODED_CRT_CHAIN",
    "clientCrt":"BASE64_ENCODED_CLIENT_CRT",
    "caCrt":"BASE64_ENCODED_CA_CRT"
}
```

After you receive the certificate, decode it and use it in your Application. 

## Call the metadata endpoint

Call the `metadata` endpoint with the generated certificate to get URLs to the following:

- the Application Registry API
- the Event Service API
- the `certificate renewal` endpoint
- the `certificate revocation` endpoint

The URL to the `metadata` endpoint is returned in the response body from the configuration URL. Use the value of the `api.infoUrl` property to get the URL. Run:

```bash
curl https://gateway.{CLUSTER_DOMAIN}/v1/applications/management/info --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
```

A successful call returns the following response:

```json
{
  "clientIdentity": {
    "application": "{APP_NAME}"
  },
  "urls": {
    "metadataUrl": "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services",
    "eventsUrl": "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events",
    "renewCertUrl": "https://gateway.{CLUSTER_DOMAIN}/v1/applications/certificates/renewals",
    "revokeCertUrl": "https://gateway.{CLUSTER_DOMAIN}/v1/applications/certificates/revocations"
  },
  "certificate": {
    "subject": "OU=Test,O=Test,L=Blacksburg,ST=Virginia,C=US,CN={APP_NAME}",
    "extensions": "string",
    "key-algorithm": "rsa2048"
  }
}
```

Use `urls.metadataUrl` and `urls.eventsUrl` to get the URLs to the Application Registry API and to the Event Service API.

## Call the Application Registry and Event Service on local deployment

Since Kyma installation on Minikube uses the self-signed certificate by default, skip TLS verification.

Call the Application Registry with this command:

```bash
curl https://gateway.kyma.local/{APP_NAME}/v1/metadata/services --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
```

Use this command to call the Event Service:

```bash
curl https://gateway.kyma.local/{APP_NAME}/v1/events --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
```
