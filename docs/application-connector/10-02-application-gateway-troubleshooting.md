---
title: Application Gateway troubleshooting
type: Troubleshooting
---

If you call a registered service and receive an error:

- Verify that the call reached Application Gateway.  
  To do that fetch logs from Application Gateway Pod:
  ```
  kubectl -n kyma-integration logs -l app={APP_NAME}-application-gateway -c {APP_NAME}-application-gateway
  ```
  If the request reached the Pod, it should be logged by Application Gateway.
  
  If the call is not in the logs, check if Access Service exists.
  ```
  kubectl -n kyma-integration get svc app-{APP_NAME}-{SERVICE_ID}
  ```
  If it doesn't, try to deregister the Service using the following command:

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

  and register it again.

- Verify that the target URL is correct. 
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

  You should receive the `json` response with the service definition.

  Access the target URL directly to verify that the value of `api.targetUrl` is correct.
