---
title: Error when calling a registered service
type: Troubleshooting
---

If you call a registered service and receive an error, follow these steps to detect the source of the problem:


1. Check the logs.

    Check the logs from the Application Gateway Pod to verify that the call reached the Application Gateway.
    To fetch these logs, run this command:
    
    ```bash
    kubectl -n kyma-integration logs -l app={APP_NAME}-application-gateway -c {APP_NAME}-application-gateway
    ```
   
    The request that reached the Pod is logged by the Application Gateway.

2. Check for the Access Service.

    If the call you tried to make is not in the logs, check if the [Access Service](#architecture-application-connector-components-access-service) for the service you are trying to call exists.
    ```
    kubectl -n kyma-integration get svc {APP_NAME}-{SERVICE_ID}
    ```
3. Re-register the service.

    If the Access Service does not exist, deregister the service you tried to call.

    <div tabs name="deregistration" group="error-when-calling-a-registered-service">
      <details>
      <summary label="with-a-trusted-certificate">
      With a trusted certificate
      </summary>

      ```bash
      curl -X DELETE https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{SERVICE_ID} --cert {CERTIFICATE_FILE} --key {KEY_FILE}
      ```
      </details>
      <details>
      <summary label="without-a-trusted-certificate">
      Without a trusted certificate
      </summary>

      ```bash
      curl -X DELETE https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{SERVICE_ID} --cert {CERTIFICATE_FILE} --key {KEY_FILE} -k
      ```
      </details>
    </div>

    Then, register the service and try calling again. Registering the service again recreates the Access Service.
    To register a service, see [this tutorial](#tutorials-register-a-service-register-a-service).


4. Check the API URL.

    If your call reaches the Application Gateway and the Access Service exists, but you still receive an error, check if the API URL in the service definition matches the API URL of the actual service you are trying to call.
    To check the target URL of the API, fetch the Service definition from the Application Registry:

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
