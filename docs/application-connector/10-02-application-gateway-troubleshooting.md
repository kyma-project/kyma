---
title: Error when calling a registered Service
type: Troubleshooting
---

If you call a registered service and receive an error:

- Verify that the call reached Application Gateway.  
  To do that, fetch logs from Application Gateway Pod:
  ```
  kubectl -n kyma-integration logs -l app={APP_NAME}-application-gateway -c {APP_NAME}-application-gateway
  ```
  If the request reaches the Pod, it is logged by Application Gateway.
  
  If the call is not in the logs, check if [Access Service](components/application-connector/#architecture-application-connector-components-access-service) exists.
  ```
  kubectl -n kyma-integration get svc app-{APP_NAME}-{SERVICE_ID}
  ```
  If Access Service does not exist, run this command to deregister the Service:

  <div tabs name="deregistration">
    <details>
    <summary>
    With trusted certificate
    </summary>

    ```
    curl -X DELETE https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{SERVICE_ID} --cert {CERTIFICATE_FILE} --key {KEY_FILE}
    ```
    </details>
    <details>
    <summary>
    Without trusted certificate
    </summary>

    ```
    curl -X DELETE https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{SERVICE_ID} --cert {CERTIFICATE_FILE} --key {KEY_FILE} -k
    ```
    </details>
  </div>

  Then register the Service.  
  To register a Service, see [this tutorial](components/application-connector/#tutorials-register-a-service-register-a-service) again.


- If you still receive an error, check the target URL of the API that you have registered and verify that it is correct.  
  To do that, fetch the Service definition from Application Registry:

  <div tabs name="verification">
    <details>
    <summary>
    With trusted certificate
    </summary>

    ```
    curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{SERVICE_ID} --cert {CERTIFICATE_FILE} --key {KEY_FILE}
    ```
    </details>
    <details>
    <summary>
    Without trusted certificate
    </summary>

    ```
    curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{SERVICE_ID} --cert {CERTIFICATE_FILE} --key {KEY_FILE} -k
    ```
    </details>
  </div>

  You should receive a `json` response with the Service definition that contains the target URL.  
  Access the target URL directly to verify that the value of `api.targetUrl` is correct.
