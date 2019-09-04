---
title: Error when calling a registered service
type: Troubleshooting
---

If you call a registered service and receive an error, follow these steps to detect the source of the problem:


1. Check the logs

    Check the logs from Application Gateway Pod to verify that the call reached Application Gateway.   
    To fetch these logs, run this command:
    ```
    kubectl -n kyma-integration logs -l app={APP_NAME}-application-gateway -c {APP_NAME}-application-gateway
    ```
    The request that reached the Pod is logged by Application Gateway.
  
2. Check for Access Service

    If the call you tried to make is not in the logs, check if an [Access Service](#architecture-application-connector-components-access-service) exists for the service you are trying to call.
    ```
    kubectl -n kyma-integration get svc app-{APP_NAME}-{SERVICE_ID}
    ```
3. Re-register the service

    If Access Service does not exist, run this command to deregister the service you tried to call:

    <div tabs name="deregistration">
      <details>
      <summary>
      With a trusted certificate
      </summary>

      ```
      curl -X DELETE https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{SERVICE_ID} --cert {CERTIFICATE_FILE} --key {KEY_FILE}
      ```
      </details>
      <details>
      <summary>
      Without a trusted certificate
      </summary>

      ```
      curl -X DELETE https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{SERVICE_ID} --cert {CERTIFICATE_FILE} --key {KEY_FILE} -k
      ```
      </details>
    </div>

    Then, register the service and try calling again. Registering the service again recreates the Access Service.  
    To register a service, see [this tutorial](components/application-connector/#tutorials-register-a-service-register-a-service).


4. Check the API URL

    If your call reaches the Application Gateway and the Access Service exists, but you still receive an error, check if the API URL in the service definition matches the API URL of the actual service you are trying to call.  
    To check the target URL of the API, fetch the Service definition from Application Registry:

    <div tabs name="verification">
      <details>
      <summary>
      With a trusted certificate
      </summary>

      ```
      curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{SERVICE_ID} --cert {CERTIFICATE_FILE} --key {KEY_FILE}
      ```
      </details>
      <details>
      <summary>
      Without a trusted certificate
      </summary>

      ```
      curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{SERVICE_ID} --cert {CERTIFICATE_FILE} --key {KEY_FILE} -k
      ```
      </details>
    </div>

    A successful call returns a `json` response with the service definition that contains the target URL.  
    Call the target URL directly to verify that the value of `api.targetUrl` is correct.
