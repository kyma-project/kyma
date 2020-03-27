---
title: Cannot connect to a service exposed by an APIRule
type: Troubleshooting
---

##  Basic troubleshooting

API Gateway is a controller. It adds a status to the rules it processes. For basic troubleshooting, check the APIRule status:

   ```bash
   kubectl describe apirules.gateway.kyma-project.io -n {NAMESPACE} {NAME}
   ```

If the status is `Error`, edit the APIRule and fix issues described in `.Status.APIRuleStatus.Desc`. If you still encounter issues, make sure the API Gateway, Hydra and Oathkeeper are running or take a look at other, more specific troubleshooting guides.

## 401 Unauthorized or 403 Forbidden

If you reach your service and get `401 Unauthorized` or `403 Forbidden` in response, make sure that:

- You are using an access token with proper scopes and it is active:

  1. Export the credentials of your OAuth2Client as environmental variables:
  
      ```bash
      export CLIENT_ID="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_id}' | base64 --decode)"
      export CLIENT_SECRET="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_secret}' | base64 --decode)"
      ```
     
  2. Encode your client credentials and export them as an environment variable:
  
      ```bash
      export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
      ```
     
  3. Check access token status:
  
      ```bash
      curl -X POST "https://oauth2.{DOMAIN}/oauth2/introspect" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "token={ACCESS_TOKEN}"
      ```
     
  4. Generate a [new access token](/components/api-gateway-v2/#tutorials-expose-and-secure-a-service-register-an-o-auth2-client-and-get-tokens) if needed.
  
- Your client from OAuth2Client resource is registered properly in Hydra OAuth2 and OpenID Connect server. You need to call the Hydra administrative endpoint `/client` from inside of the cluster. Follow this steps:

  1. Expose Hydra to your local environment:
  
      ```bash
      kubectl port-forward -n kyma-system service/ory-hydra-admin 4445
      ```
  
  2. Fetch the Client ID from Secret specified in the OAuth2Client resource:
  
      ```bash
      kubectl get secrets {SECRET_NAME} -n {SECRET_NAMESPACE} -o jsonpath='{ .data.client_id }' | base64 --decode
      ```
     
  3. Call Hydra:
  
      ```bash
      curl localhost:4445/clients
      # Or if you have jq installed
      curl localhost:4445/clients | jq '.'
      ```
     
  4. If the Client ID from step 2 is not available on the clients list, make sure Hydra Maester has access to the database and/or restart the Hydra Measter Pod.
      
## 404 Not Found

If you reach your service and get `404 Not Found` in response, make sure that:

- Proper Oathkeeper Rule has been created:

  ```bash
  kubectl get rules.oathkeeper.ory.sh -n {NAMESPACE}
  ```
  
  >**TIP:** Name of the Rule consists of the name of the APIRule and a random suffix.

- Proper VirtualService has been created:

  ```bash
  kubectl get virtualservices.networking.istio.io -n {NAMESPACE}
  ```
  
  >**TIP:** Name of the VirtualService consists of the name of the APIRule and a random suffix.
