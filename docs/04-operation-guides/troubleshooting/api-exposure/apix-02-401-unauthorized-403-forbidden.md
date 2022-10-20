---
title: Cannot connect to a service exposed by an API Rule - 401 Unauthorized or 403 Forbidden
---

## Symptom

When you try to reach your service, you get `401 Unauthorized` or `403 Forbidden` in response.

## Remedy

Make sure that the following conditions are met:

- You are using an access token with proper scopes and it is active:

  1. Export the credentials of your OAuth2Client as environment variables:

      > **NOTE:** Export the **CLIENT_NAMESPACE** and **CLIENT_NAME** variables before you proceed with step 1.
      
      ```bash
      export CLIENT_ID="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_id}' | base64 --decode)"
      export CLIENT_SECRET="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_secret}' | base64 --decode)"
      ```

  2. Encode your client credentials and export them as an environment variable:

      ```bash
      export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
      ```

  3. Check the access token status:

      ```bash
      curl -X POST "https://oauth2.{CLUSTER_DOMAIN}/oauth2/introspect" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "token={ACCESS_TOKEN}"
      ```

  4. Generate a [new access token](../../../03-tutorials/00-api-exposure/apix-05-expose-and-secure-workload-oauth2.md#register-an-oauth2-client-and-get-tokens) if needed.

- Your client from the OAuth2Client resource is registered properly in Hydra OAuth2 and the OpenID Connect server. You need to call the Hydra administrative endpoint `/client` from inside of the cluster. Follow these steps:

  1. Fetch the Client ID from the Secret specified in the OAuth2Client resource:

      ```bash
      kubectl get secrets {SECRET_NAME} -n {SECRET_NAMESPACE} -o jsonpath='{ .data.client_id }' | base64 --decode
      ```

  2. Create a simple curl Pod:

      ```bash
      cat <<EOF | kubectl apply -f -
      apiVersion: v1
      kind: Pod
      metadata:
        labels:
          app: ory-curl
        name: ory-curl
        namespace: {SECRET_NAMESPACE}
      spec:
        containers:
        - name: curl
          image: alpine
          terminationMessagePolicy: "FallbackToLogsOnError"
          command:
            - /bin/sh
            - -c
            - |
              apk add curl jq
              curl ory-hydra-admin.kyma-system.svc.cluster.local:4445/clients | jq '.'
      EOF
      ```

  3. Check logs from the `ory-curl` Pod:

      ```bash
      kubectl logs -n {SECRET_NAMESPACE} ory-curl curl
      ```

  4. If the Client ID from step 1 is not available on the client list, make sure Hydra has access to the database and/or restart the Hydra Measter Pod. You can check the logs using the following commands:

     ```bash
      # Check logs from the Hydra-Maester controller application
      kubectl logs -n kyma-system -l "app.kubernetes.io/name=hydra-maester" -c hydra-maester
      # Example output
      2020-05-04T12:19:04.472Z  INFO  controller-runtime.controller Starting EventSource  {"controller": "oauth2client", "source": "kind source: /, Kind="}2020-05-04T12:19:04.472Z  INFO  setup starting manager
      2020-05-04T12:19:04.573Z  INFO  controller-runtime.controller Starting Controller {"controller": "oauth2client"}
      2020-05-04T12:19:04.673Z  INFO  controller-runtime.controller Starting workers  {"controller": "oauth2client", "worker count": 1}
      2020-05-04T12:26:30.819Z  INFO  controllers.OAuth2Client  using default client
      2020-05-04T12:26:30.835Z  INFO  controllers.OAuth2Client  using default client
      # This log entry informs us that a client has been created, and should be visible within Hydra
      2020-05-04T12:26:31.468Z  DEBUG controller-runtime.controller Successfully Reconciled {"controller": "oauth2client", "request": "test-ns/test-client"}
      ```

      ```bash
      # Check logs from the Hydra application
      kubectl logs -n kyma-system -l "app.kubernetes.io/name=hydra" -c hydra
      ```
