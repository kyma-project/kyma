---
title: Obtaining the Certificate
type: Getting Started
---

After you have created Remote Environment within Kyma it's time to connect your
external system, this process consists of two steps:
- Obtaining the certificate
- Registering services

In this Getting Started section we will focus on the first step.

Metadata Service (a component responsible for managing registered services)
is secured with mutual TLS authentication and thus - the client needs a proper
certificate in order to be able to consume it's API.

## Prerequisites

During this scenario you will need to create a proper CSR (Certificate Signing
Request), keys and certificates, to fulfil security standards we highly recommend
you to use the
[OpenSSL toolkit](https://www.openssl.org/docs/man1.0.2/apps/openssl.html)
to do so.

First of all you will need a private RSA key, if your external service does not already
have one you can create it by using following command:

`openssl genrsa -out generated.key 4096`

After you've got your key you can proceed to the next step.

## Retrieving token

>**NOTE:** In this step we are beginning so-called 'one click integration', you can check the whole flow with all the details described in [here.](TODO)

You can get the required token in two ways, first one makes the use of Kyma
console UI:
- From home page navigate to **Administration**
- In **Integration** section select **Remote Environments**
- Choose the Remote Environment that you want to connect your external system to
- Click **Connect Remote Environment**
- Copy the token by clicking **Copy to clipboard**

The second way enables you to fetch the token with the use of `kubectl port-forward`:
- First, you need to expose Connector Service outside of Kubernetes:
  ```
  kubectl -n=kyma-integration port-forward svc/connector-service-internal-api 8080:8080
  ```

- You should now be able to make a POST request to the following endpoint
  ```
  curl -X POST http://localhost:8080/v1/remoteenvironments/{remote-environment-name}/tokens
  ```

  This call should yield response of following structure:
  ```
  {
    "url":"{CONFIGURATION_URL_WITH_TOKEN}",
    "token":"example-token-123"
  }
  ```

>**NOTE:** This token that is valid for a single CSR information call. If you need to get the configuration details once again, just follow above steps to fetch a new token.

## Fetching CSR information from Kyma

You can now use provided link to fetch information about the Kyma URLs and CSR
configuration that you need to provide, to do so - make the following request:
```
curl {CONFIGURATION_URL_WITH_TOKEN}
```

The response should look like:
```
{
    "csrUrl": "{CSR_SIGNING_URL_WITH_TOKEN}",
    "api":{
        "metadataUrl":      "https://gateway.{CLUSTER_NAME}.kyma.cluster.cx/{REMOTE-ENVIRONMENT-NAME}/v1/metadata/services",
        "eventsUrl":        "https://gateway.{CLUSTER_NAME}.kyma.cluster.cx/{REMOTE-ENVIRONMENT-NAME}/v1/events",
        "certificatesUrl":  "https://connector-service.{CLUSTER_NAME}.kyma.cluster.cx/v1/remoteenvironments/{remote-environment-name}",
    },
    "certificate":{
        "subject":"OU=Test,O=Test,L=Blacksburg,ST=Virginia,C=US,CN={REMOTE-ENVIRONMENT-NAME}",
        "extensions": "",
        "key-algorithm": "rsa2048",
    }
}
```

## Generating CSR and sending it to Kyma

You can now generate CSR with values provided by Connector Service:
```
openssl req -new -out generated.csr -key generated.key -subj "/OU=OrgUnit/O=Organization/L=Waldorf/ST=Waldorf/C=DE/CN={REMOTE-ENVIRONMENT-NAME}"
```

After the CSR is created, encode it with Base64 and then send it to Kyma:
```
curl -H "Content-Type: application/json" -d '{"csr":"{BASE64_ENCODED_CSR_HERE}"}' {CSR_SIGNING_URL_WITH_TOKEN}
```

In the response there is a valid certificate signed by Kyma's Certificate Authority:
```
{
    "crt":"BASE64_ENCODED_CRT"
}
```

After you've received your certificate you can start registering your services in
Metadata Service API, you can find details on this topic in the next section of
this Getting Started doc.