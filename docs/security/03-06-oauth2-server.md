---
title: OAuth2 and OpenID Connect server
type: Details
---

By default, every Kyma deployment comes with an OAuth2 authorization server solution from [ORY](https://www.ory.sh/). The `ory` [component](https://github.com/kyma-project/kyma/tree/master/resources/ory) consists of three elements:

- [Hydra](https://github.com/ory/hydra) OAuth2 and OpenID Connect server which issues access, refresh, and ID tokens to registered clients which call services in a Kyma cluster.
- [Oathkeeper](https://github.com/ory/oathkeeper) authorization & authentication proxy which authenticates and authorizes incoming requests basing on the list of defined [Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules).
- [Oathkeeper Maester](https://github.com/ory/oathkeeper-maester) Kubernetes controller which feeds Access Rules to the Oathkeeper proxy by creating or updating the Oathkeeper ConfigMap and populating it with rules found in instances of the `rules.oathkeeper.ory.sh/v1alpha1` custom resource.

Out of the box, the Kyma implementation of the ORY stack supports the [OAuth 2.0 Client Credentials Grant](https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/).

>**NOTE:** The implementation of the ORY Oauth2 server in Kyma is still in the early stages and is subject to changes and development. Read [this](https://kyma-project.io/blog/2019/7/31/kyma-collaboration-with-ory/) blog post to get a better understanding of the target integration of the ORY stack in Kyma.

## Register an OAuth2 client

To interact with the Kyma OAuth2 server, you must register an OAuth2 client. Follow these steps:

1. Export the required client application details and your cluster domain as environment variables. Run:
  ```
  export CLIENT_ID={YOUR_CLIENT_ID}
  export CLIENT_SECRET={YOUR_CLIENT_SECRET}
  export DOMAIN={YOUR_CLUSTER_DOMAIN}
  ```

2. Send a CURL request to the `/clients` endpoint of the OAuth2 server to register your client with read and write scopes. Run:
  ```
  curl -ik -X POST "https://oauth2-admin.$DOMAIN/clients" -d '{"grant_types":["client_credentials"], "client_id":"'$CLIENT_ID'", "client_secret":"'$CLIENT_SECRET'", "scope":"read write"}'
  ```

## OAuth2 server in action

After you register an OAuth2 client, go to [this](https://github.com/kyma-incubator/examples/tree/master/ory-hydra/scenarios/client-credentials) Kyma Incubator repository to try a Client Credentials Grant example that showcases the integration.

You can also interact with the OAuth2 server using its REST API. Read the official [ORY documentation](https://www.ory.sh/docs/hydra/sdk/api) to learn more about the available endpoints.

>**TIP:** If you have any questions about the ORY-Kyma integration, you can ask them on the **#security** [Slack channel](http://slack.kyma-project.io/) and get answers directly from Kyma developers.   
