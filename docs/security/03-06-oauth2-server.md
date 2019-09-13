---
title: OAuth2 and OpenID Connect server
type: Details
---

By default, every Kyma deployment comes with an OAuth2 authorization server solution from [ORY](https://www.ory.sh/). The `ory` [component](https://github.com/kyma-project/kyma/tree/master/resources/ory) consists of four elements:

- [Hydra](https://github.com/ory/hydra) OAuth2 and OpenID Connect server which issues access, refresh, and ID tokens to registered clients which call services in a Kyma cluster.
- [Oathkeeper](https://github.com/ory/oathkeeper) authorization & authentication proxy which authenticates and authorizes incoming requests basing on the list of defined [Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules).
- [Oathkeeper Maester](https://github.com/ory/oathkeeper-maester) Kubernetes controller which feeds Access Rules to the Oathkeeper proxy by creating or updating the Oathkeeper ConfigMap and populating it with rules found in instances of the `rules.oathkeeper.ory.sh/v1alpha1` custom resource.
- [Hydra Maester](https://github.com/ory/hydra-maester) Kubernetes controller which manages OAuth2 clients by communicating the data found in instances of the `oauth2clients.hydra.ory.sh` custom resource to the ORY Hydra API.

Out of the box, the Kyma implementation of the ORY stack supports the [OAuth 2.0 Client Credentials Grant](https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/).

>**NOTE:** The implementation of the ORY Oauth2 server in Kyma is still in the early stages and is subject to changes and development. Read [this](https://kyma-project.io/blog/2019/7/31/kyma-collaboration-with-ory/) blog post to get a better understanding of the target integration of the ORY stack in Kyma.

## Register an OAuth2 client

To interact with the Kyma OAuth2 server, you must register an OAuth2 client. For each client, you can either specify your own ID and password or let Hydra OAuth2 server allocate random credentials.

>**NOTE:** By default, you can create clients only in the `kyma-system` and `default` Namespaces. Read [this](https://github.com/ory/k8s/blob/master/docs/helm/hydra-maester.md#configuration) document to learn how to enable creating clients in other Namespaces. See the ORY Hydra Maester [Github page](https://github.com/ory/hydra-maester) to learn more about the `oauth2clients.hydra.ory.sh` custom resource. 

### Use your own credentials

1. Create a Kubernetes Secret that contains an ID and password you want to use to create a client:

```
apiVersion: v1
kind: Secret
metadata:
  name: {NAME_OF_SECRET}
  namespace: {CLIENT_NAMESPACE}
type: Opaque
data:
  client_id: {BASE64_ENCODED_ID}
  client_secret: {BASE64_ENCODED_PASSWORD}
```

2. Create a custom resource with the `secretName` property set to the name of Kubernetes Secret you created. Run this command to trigger the creation of a client:

```
cat <<EOF | kubectl apply -f -
apiVersion: hydra.ory.sh/v1alpha1
kind: OAuth2Client
metadata:
  name: {NAME_OF_CLIENT}
  namespace: {CLIENT_NAMESPACE}
spec:
  grantTypes:
    - "client_credentials"
  scope: "read write"
  secretName: {NAME_OF_SECRET}
EOF
```

>**NOTE:** Each instance of the `oauth2clients.hydra.ory.sh` custom resource and the Secret that stores the credentials of the corresponding client must exist in the same Namespace.

Creating this custom resource triggers the Hydra Maester controller which sends a client registration request to the OAuth2 server. You can interact with the client using the credentials specified in the secret.

### Use Hydra-generated credentials

Run this command to create a custom resource that triggers the creation of a client:

```
cat <<EOF | kubectl apply -f -
apiVersion: hydra.ory.sh/v1alpha1
kind: OAuth2Client
metadata:
  name: {NAME_OF_CLIENT}
  namespace: {CLIENT_NAMESPACE}
spec:
  grantTypes:
    - "client_credentials"
  scope: "read write"
  secretName: {NAME_OF_KUBERNETES_SECRET}
EOF
```

Creating this custom resource triggers the Hydra Maester controller which sends a client registration request to the OAuth2 server and creates a new Kubernetes Secret using the name specified in the `secretName` property. This Secret contains the credentials of the registered client.

>**NOTE:** Each instance of the `oauth2clients.hydra.ory.sh` custom resource and the Secret that stores the credentials of the corresponding client share the Namespace.

Run this command to get the credentials of the registered OAuth2:
```
kubectl get secret -n {CLIENT_NAMESPACE} {NAME_OF_KUBERNETES_SECRET} -o yaml
```


## OAuth2 server in action

After you register an OAuth2 client, go to [this](https://github.com/kyma-incubator/examples/tree/master/ory-hydra/scenarios/client-credentials) Kyma Incubator repository to try a Client Credentials Grant example that showcases the integration.

You can also interact with the OAuth2 server using its REST API. Read the official [ORY documentation](https://www.ory.sh/docs/hydra/sdk/api) to learn more about the available endpoints.

>**TIP:** If you have any questions about the ORY-Kyma integration, you can ask them on the **#security** [Slack channel](http://slack.kyma-project.io/) and get answers directly from Kyma developers.   