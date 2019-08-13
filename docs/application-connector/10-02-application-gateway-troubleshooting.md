---
title: Application Gateway troubleshooting
type: Troubleshooting
---

In case user calls registered service and receives an error:
- Verify that the call reached Application Gateway. 
  To do that fetch logs from Application Gateway pod:
  ```
  kubectl -n kyma-integration logs -l app={APP_NAME}-application-gateway -c {APP_NAME}-application-gateway
  ```
  If the request reached the pod, it should be logged by Application Gateway.
  
  If the call is not in the logs, check if Access Service exists.
  ```
  kubectl -n kyma-integration get svc app-{APP_NAME}-{SERVICE_ID}
  ```
  If it doesn't, try to deregister the Service using the following command

  <div tabs name="installation">
    <details>
    <summary>
    From release
    </summary>

    When you install Kyma locally from a release, follow [this](https://kyma-project.io/docs/master/root/kyma/#installation-install-kyma-locally) guide.
    Ensure that you created the local Kubernetes cluster with `10240Mb` memory and `30Gb` disk size.
    ```
    ./scripts/minikube.sh --domain "kyma.local" --vm-driver "hyperkit" --memory 10240Mb --disk-size 30g
    ```

    Run the following command before triggering the Kyma installation process:
    ```
    kubectl -n kyma-installer patch configmap installation-config-overrides -p '{"data": {"global.knative": "true", "global.kymaEventBus": "false", "global.natsStreaming.clusterID": "knative-nats-streaming"}}'
    ```
    </details>
    <details>
    <summary>
    From sources
    </summary>

    When you install Kyma locally from sources, add the `--knative` argument to the `run.sh` script. Run this command:

    ```
    ./run.sh --knative
    ```
    </details>
  </div>

  and register it again.

- Verify that the target URL is correct. 
  To do that, you can fetch the Service definition from Application Registry:

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

  Access the target url directly to verify that the value of `api.targetUrl` is correct.