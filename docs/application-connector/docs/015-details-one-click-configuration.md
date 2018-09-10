---
title: Automatic connection configuration
type: Details
---

Kyma Application Connector allows to authenticate and securely communicate with different external solutions. Kyma provides an easy way to set up such connection through the automatic connection configuration mechanism.

## Flow description

The automatic connection configuration flow is presented in this diagram:
![Automatic Configuration Flow](./assets/002-automatic-configuration.png)

This example assumes that a new Remote Environment intended to connect the external solution already exists and is in the `disconnected` state, which means that there are no external solutions connected to it.

On the diagram, the administrator on the Kyma side and on the external system side is the same person.

1. The admin requests for a token using the CLI or the UI and receives a link with the token, which is valid for a limited period of time.
2. The admin passes the token to the external system, which requests for information regarding the Kyma installation. In the response, it receives the following information:
    - the URL to which a third-party solution sends its Certificate Signing Request (CSR)
    - URLs of the available APIs
    - information required to generate a CSR
3. The external system generates a CSR based on the information provided by Kyma and sends the CSR to the designated URL. In the response, the external system receives a signed certificate. It can use the certificate to authenticate and safely communicate with Kyma.

## Configuration steps

Follow these steps to configure the automatic connection between the Kyma Application Connector and an external solution:

1. Get the configuration address URL with a valid token.

  Using the UI:

   - Go to the Kyma console UI.
   - Select **Administration**.
   - Select the **Remote Environments** from the **Integration** section.
   - Choose the Remote Environment to which you want to connect the external solution.
   - Click **Connect Remote Environment**.
   - Copy the token by clicking **Copy to clipboard**.


  >**NOTE:** When you connect an external solution to a local Kyma deployment, you must set NodePort of the `core-nginx-ingress-controller` for the Gateway Service and for the Event Service.
  To get the NodePort, run:
    ```
    kubectl -n kyma-system get svc core-nginx-ingress-controller -o 'jsonpath={.spec.ports[?(@.port==443)].nodePort}'
    ```
  Set it for the Gateway Service and the Event Service using these calls:
    ```
    curl https://gateway.kyma.local:{NODE_PORT}/ec-default/v1/metadata/services --cert ec-default.crt --key ec-default.key -k
    ```
    ```
    curl https://gateway.kyma.local:{NODE_PORT}/ec-default/v1/ec-default/v1/events --cert ec-default.crt --key ec-default.key -k
    ```

  Alternatively, get the configuration URL with a valid token using `kubectl port-forward` or `kubectl proxy`.

    - Request:

      First, run:
      ```
      kubectl -n=kyma-integration port-forward svc/connector-service-internal-api 8080:8080
      ```
      Send the request in a new terminal window:
      ```
      curl -X POST http://localhost:8080/v1/remoteenvironments/{remote-environment-name}/tokens
      ```
    - Response:
      ```json
      {
          "url":"{CONFIGURATION_URL_WITH_TOKEN}",
          "token":"example-token-123"
      }
      ```

2. Use the provided link to fetch information about the Kyma URLs and CSR configuration.

    - Request:
    ```
    curl {CONFIGURATION_URL_WITH_TOKEN}
    ```
    >**NOTE:** The URL you call in this step contains a token that is valid for a single call. If you need to get the configuration details once again, generate a new configuration URL with a valid token and call it again. You get a code `403` error if you call the same configuration URL more than once.

    - Response:
    ```json
    {
        "csrUrl": "https://connector-service.CLUSTER_NAME.kyma.cluster.cx/v1/remoteenvironments/{remote-environment-name}/client-certs?token=example-token-456",
        "api":{
            "metadataUrl":      "https://gateway.CLUSTER_NAME.kyma.cluster.cx/{remote-environment-name}/v1/metadata/services",
            "eventsUrl":        "https://gateway.CLUSTER_NAME.kyma.cluster.cx/{remote-environment-name}/v1/events",
            "certificatesUrl":  "https://connector-service.CLUSTER_NAME.kyma.cluster.cx/v1/remoteenvironments/{remote-environment-name}",
        },
        "certificate":{
            "subject":"OU=Test,O=Test,L=Blacksburg,ST=Virginia,C=US,CN=ec-default",
            "extensions": "",
            "key-algorithm": "rsa2048",
        }
    }
    ```

3. Use values received in the `certificate.subject` field to create a CSR.

      ```
      openssl req -new -out test.csr -key ec-default.key -subj "/OU=OrgUnit/O=Organization/L=Waldorf/ST=Waldorf/C=DE/CN=ec-default"
      ```

  After you create the CSR, make the following call:

    - Request:

      ```
      curl -H "Content-Type: application/json" -d '{"csr":"BASE64_ENCODED_CSR_HERE"}' https://connector-service.CLUSTER_NAME.kyma.cluster.cx/v1/remoteenvironments/{remote-environment-name}/client-certs?token=example-token-456
      ```

    - Response:

      ```
      {
          "crt":"BASE64_ENCODED_CRT"
      }
      ```

4. The `crt` field contains a valid base64-encoded PEM block of a certificate signed by the Kyma CA.

5. The external solution can now use the created certificate to securely authenticate and communicate with the Application Connector.
