---
title: Cannot connect to a service exposed by an APIRule
type: Troubleshooting
---

##  Basic troubleshooting

API Gateway is a controller. It adds a status to the rules it processes. For basic troubleshooting, you can check the APIRule status:

   ```
   kubectl describe apirules.gateway.kyma-project.io -n {NAMESPACE} {NAME}
   ```

If status is `Error`, edit the APIRule and fix issues described in `.Status.APIRuleStatus.Desc`. If there is no status, the APIRule has not been processed yet. Make sure the API Gateway is running.

## 401 Unauthorized

If you reach your service and get `401 Unauthorized` in response, make sure that:

- You are using an access token with proper scopes and it is active:

  1. Export the credentials of your Oauth2Client as environmental variables:
      ```
      export CLIENT_ID="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_id}' | base64 --decode)"
      export CLIENT_SECRET="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_secret}' | base64 --decode)"
      ```
  2. Encode your client credentials and export them as an environment variable:
      ```
      export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
      ```
  3. Check access token status:
      ```
      curl -X POST "https://oauth2.{DOMAIN}/oauth2/introspect" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "token={ACCESS_TOKEN}"
      ```
  4. Generate a [new access token](/components/api-gateway-v2/#tutorials-expose-and-secure-a-service-register-an-o-auth2-client-and-get-tokens) if needed.
  
- Make sure your client from Oauth2Client resource is registered properly in Hydra OAuth2 and OpenID Connect server. You need to call Hydra `/client` endpoint, but it is an administrative endpoint, so you need to call it from the inside of the cluster:

  1. Prepare a Pod with curl installed. For example, you can use [this](https://hub.docker.com/r/curlimages/curl) image.
  2. Fetch Client ID from secret specified in the Oauth2Client resource:
      ```
      kubectl get secrets {SECRET_NAME} -n {SECRET_NAMESPACE} -o jsonpath='{ .data.client_id }' | base64 --decode
      ```
  3. Call Hydra:
      ```
      kubectl exec {POD_NAME} -n {NAMESPACE} -it -- curl http://ory-hydra-admin.kyma-system:4445/clients
      ```
      If Client ID from step 2 is not available on the clients list, make sure Hydra Maester has connection to the database and/or restart the Hydra Measter Pod.
      
## 404 Not Found

If you reach your service and get `404 Not Found` in response, make sure that:

- Proper Oathkeeper Rule has been created:
  ```
  kubectl get rules.oathkeeper.ory.sh -n {NAMESPACE}
  ```
  Name of the Rule consists of name of the APIRule and a random suffix.
- Proper VirtualService has been created:
  ```
  kubectl get virtualservices.networking.istio.io -n {NAMESPACE}
  ```
  Name of the VirtualService consists of name of the APIRule and a random suffix.

## 502 Bad Gateway

If you reach your service and get `502 Bad Gateway` in response, make sure that:

- Name of the Service specified in the APIRule is proper

## 503 Service Unavailable

If you reach your exposed service and get `503 Service Unavailable` in response, make sure that:

- Oatkeeper authorization & authentication proxy is running
- Application to which Service points to is running
