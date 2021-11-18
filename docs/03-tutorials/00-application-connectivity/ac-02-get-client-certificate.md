---
title: Get the client certificate
---

After you create an Application, connect it to an external solution to consume the solution's APIs and event catalogs in Kyma. To accomplish this, get the client certificate for the external solution and register its services.

This guide shows you how to get the client certificate.

>**NOTE:** The client certificate is valid for 92 days. See how to [renew the client certificate](ac-06-renew-client-cert.md), and 
how to [revoke the client certificate](../../03-tutorials/00-application-connectivity/ac-07-revoke-client-cert.md), which prevents it from being renewed.

## Prerequisites

- [OpenSSL toolkit](https://www.openssl.org/source/) to create a Certificate Signing Request (CSR), keys, and certificates which meet high security standards
- The [jq](https://stedolan.github.io/jq/download/) tool to prettify the JSON output
- Your [Application name exported](ac-01-create-application.md#prerequisites) as an environment variable

> **CAUTION:** On a local Kyma deployment, skip SSL certificate verification when making a `curl` call, by adding the `-k` flag to it. Alternatively, add the Kyma certificates to your local certificate storage on your machine using the `kyma import certs` command.

## Get the configuration URL with a token

To get the configuration URL which allows you to fetch the required configuration details, create a TokenRequest custom resource (CR). The controller which handles this CR kind adds the **status** section to the created CR. The **status** section contains the required configuration details.

1. Create a TokenRequest CR. The CR name must match the name of the Application for which you want to get the configuration details. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: applicationconnector.kyma-project.io/v1alpha1
   kind: TokenRequest
   metadata:
     name: $APP_NAME
   EOF
   ```

2. Fetch the TokenRequest CR you created to get the configuration details from the **status** section. Run:

   ```bash
   kubectl get tokenrequest.applicationconnector.kyma-project.io $APP_NAME -o yaml
   ```

   >**NOTE:** If the response doesn't contain the **status** section, wait for a few moments and fetch the CR again.

   A successful call returns the following response:

   ```yaml
   apiVersion: applicationconnector.kyma-project.io/v1alpha1
   context: {}
   kind: TokenRequest
   metadata:
     ...
     name: {APP_NAME}
     ...
   status:
     application: {APP_NAME}
     expireAfter: 2018-11-22T18:38:44Z
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
        "eventsInfoUrl":    "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events/subscribed",
        "eventsUrl":        "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events",
        "metadataUrl":      "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services",
        "infoUrl":          "https://gateway.{CLUSTER_DOMAIN}/v1/applications/management/info",
        "certificatesUrl":  "https://connector-service.{CLUSTER_DOMAIN}/v1/applications/certificates",
    },
    "certificate":{
        "subject": "O=Organization,OU=OrgUnit,L=Waldorf,ST=Waldorf,C=DE,CN={APP_NAME}",
        "extensions": "",
        "key-algorithm": "rsa2048",
    }
}
```

> **NOTE:** The response contains URLs to the Application Registry API and the Events Publisher API, however, using them is not recommended. To fetch correct URLs to these APIs, as well as other configuration details, call the `metadata` endpoint URL. It is provided in the **api.infoUrl** property.

## Generate a CSR and send it to Kyma

1. Export the names of the generated CSR, client certificate and key, and your [cluster domain](../../02-get-started/01-quick-install.md#export-your-cluster-domain) as environment variables:

   ```bash
   export CSR_FILE_NAME=generated
   export CLIENT_CERT_FILE_NAME=generated
   export KEY_FILE_NAME=generated
   export CLUSTER_DOMAIN=local.kyma.dev
   ```
   
2. Generate a CSR using the certificate subject data obtained in the previous step:

   ```bash
   openssl genrsa -out generated.key 2048
   openssl req -new -sha256 -out $CSR_FILE_NAME.csr -key $KEY_FILE_NAME.key -subj "/OU=OrgUnit/O=Organization/L=Waldorf/ST=Waldorf/C=DE/CN=$APP_NAME"
   openssl base64 -in $CSR_FILE_NAME.csr -A
   ```
   
3. Send the encoded CSR to Kyma. Run:

   ```bash
   curl -X POST -H "Content-Type: application/json" -d '{"csr":"BASE64_ENCODED_CSR_HERE"}' {CSR_SIGNING_URL_WITH_TOKEN}
   ```

   The response contains a valid client certificate signed by the Kyma Certificate Authority (CA), a CA certificate, and a certificate chain.
   
   ```json
   {
       "crt":"BASE64_ENCODED_CRT_CHAIN",
       "clientCrt":"BASE64_ENCODED_CLIENT_CRT",
       "caCrt":"BASE64_ENCODED_CA_CRT"
   }
   ```

4. After you receive the certificates, decode them and save them as files.

   <div tabs name="Decode and save the certificates" group="generate-csr-and-send-to-kyma">
     <details open>
     <summary label="macOS">
     macOS
     </summary>
   
   To do that, select the JSON response you got, copy it, and run:
   ```bash
   pbpaste | jq -r ".crt" | base64 -D > generated.crt
   pbpaste | jq -r ".clientCrt" | base64 -D > generated_client.crt
   pbpaste | jq -r ".caCrt" | base64 -D > generated_ca.crt
   ```
   
     </details>
     <details>
     <summary label="other">
     other
     </summary>
   
      Decode the certificates manually and save them as `generated.crt`, `generated_client.crt`, and `generated_ca.crt` respectively. 
   
     </details>
   </div>

Now you can use the decoded certificate chain in your Application.

## Call the metadata endpoint

Call the `metadata` endpoint with the generated certificate to get URLs to the following:

- the Application Registry API
- the Event Publisher API
- the certificate renewal endpoint
- the certificate revocation endpoint

The URL to the `metadata` endpoint is returned in the response body from the configuration URL. Use the value of the **api.infoUrl** property to get the URL. Run:

```bash
curl https://gateway.$CLUSTER_DOMAIN/v1/applications/management/info --cert $CLIENT_CERT_FILE_NAME.crt --key $KEY_FILE_NAME.key
```

A successful call returns the following response:

```json
{
  "clientIdentity": {
    "application": "{APP_NAME}"
  },
  "urls": {
    "eventsInfoUrl": "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events/subscribed",
    "eventsUrl": "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events",
    "metadataUrl": "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services",
    "renewCertUrl": "https://gateway.{CLUSTER_DOMAIN}/v1/applications/certificates/renewals",
    "revokeCertUrl": "https://gateway.{CLUSTER_DOMAIN}/v1/applications/certificates/revocations"
  },
  "certificate": {
    "subject": "O=Organization,OU=OrgUnit,L=Waldorf,ST=Waldorf,C=DE,CN={APP_NAME}",
    "extensions": "string",
    "key-algorithm": "rsa2048"
  }
}
```

Use **urls.metadataUrl** and **urls.eventsUrl** to get the URLs to the Application Registry API and to the Event Publisher API.

## Call Event Publisher

Use this command to call Event Publisher:

```bash
curl -X POST -H "Content-Type: application/json" https://gateway.$CLUSTER_DOMAIN/$APP_NAME/v1/events --cert $CLIENT_CERT_FILE_NAME.crt --key $KEY_FILE_NAME.key -d '{
  "event-type": "ExampleEvent",
  "event-type-version": "v1",
  "event-id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "event-time": "2018-10-16T15:00:00Z",
  "data": "some data"
  }'
```