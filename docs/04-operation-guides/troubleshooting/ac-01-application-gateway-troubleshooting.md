---
title: Error when calling a registered service
---

## Symptom

You call a registered service and receive an error.

## Remedy

Follow these steps to detect the source of the problem:

1. Check the logs.

    Check the logs from the Application Gateway Pod to verify that the call reached Application Gateway.
    To fetch these logs, run this command:
    
    ```bash
    kubectl -n kyma-system logs -l app=central-application-gateway -c central-application-gateway
    ```
   
    The request that reached the Pod is logged by Application Gateway.

2. Check the API URL.

    If your call reaches Application Gateway, but you still receive an error, check if the API URL in the service definition matches the API URL of the actual service you are trying to call.
    To check the target URL of the API, fetch the Service definition from Application Registry:

    <div tabs name="verification" group="error-when-calling-a-registered-service">
      <details>
      <summary label="with-a-trusted-certificate">
      With a trusted certificate
      </summary>

      ```bash
      curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{SERVICE_ID} --cert {CERTIFICATE_FILE} --key {KEY_FILE}
      ```
      </details>
      <details>
      <summary label="without-a-trusted-certificate">
      Without a trusted certificate
      </summary>

      ```bash
      curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{SERVICE_ID} --cert {CERTIFICATE_FILE} --key {KEY_FILE} -k
      ```
      </details>
    </div>

    A successful call returns a `json` response with the service definition that contains the target URL.
    Call the target URL directly to verify that the value of `api.targetUrl` is correct.
