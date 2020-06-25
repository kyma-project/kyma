---
title: OAuth2 and OpenID Connect server
type: Details
---

By default, every Kyma deployment comes with an OAuth2 authorization server solution from [ORY](https://www.ory.sh/). The `ory` [component](https://github.com/kyma-project/kyma/tree/master/resources/ory) consists of four elements:

- [Hydra](https://github.com/ory/hydra) OAuth2 and OpenID Connect server which issues access, refresh, and ID tokens to registered clients which call services in a Kyma cluster.
  + By default, Hydra is deployed with a database backend for persistent data management. The database used is the [official Bitnami Postgres Package](https://github.com/helm/charts/tree/master/stable/postgresql), however, we use the [Official Postgres Docker images](https://hub.docker.com/_/postgres?tab=description) instead of those provided by Bitnami. This is because Postgres provides Alpine based images, which are lighter and have a reduced attack surface.
- [Oathkeeper](https://github.com/ory/oathkeeper) authorization & authentication proxy which authenticates and authorizes incoming requests basing on the list of defined [Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules).
- [Oathkeeper Maester](https://github.com/ory/oathkeeper-maester) Kubernetes controller which feeds Access Rules to the Oathkeeper proxy by creating or updating the Oathkeeper ConfigMap and populating it with rules found in instances of the `rules.oathkeeper.ory.sh/v1alpha1` custom resource.
- [Hydra Maester](https://github.com/ory/hydra-maester) Kubernetes controller which manages OAuth2 clients by communicating the data found in instances of the `oauth2clients.hydra.ory.sh` custom resource to the ORY Hydra API.

Out of the box, the Kyma implementation of the ORY stack supports the [OAuth 2.0 Client Credentials Grant](https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/).

>**NOTE:** The implementation of the ORY Oauth2 server in Kyma is still in the early stages and is subject to changes and development. Read the [blog post](https://kyma-project.io/blog/2019/7/31/kyma-collaboration-with-ory/) to get a better understanding of the target integration of the ORY stack in Kyma.

## Register an OAuth2 client

To interact with the Kyma OAuth2 server, you must register an OAuth2 client. To register a client, create an instance of the OAuth2Client custom resource (CR) which triggers the Hydra Maester controller to send a client registration request to the OAuth2 server.

>**CAUTION:** If you run Kyma on a Minikube cluster, Hydra stores client data in an in-memory database. This configuration is prone to data loss and may cause erratic behavior of the Oauth2 server. By default, Hydra Maester reconciles the database every 10 hours, but you can resolve any discrepancies manually by deleting the Hydra Maester Pod.

When you register an OAuth2 client, you can set its redirect URI used in user-facing flows. To add a redirect URI for the client you register, use the optional **spec.redirectUris** property. For more details, see the full ORY [OAuth2Client Custom Resource Definition](https://github.com/ory/hydra-maester/blob/master/config/crd/bases/hydra.ory.sh_oauth2clients.yaml)(CRD).   

For each client, you can provide client ID and secret. If you don't provide the credentials, Hydra generates a random client ID and secret pair.
Client credentials are stored as Kubernetes Secret in the same Namespace as the CR instances of the corresponding clients.

### Use your own credentials

1. Create a Kubernetes Secret that contains the client ID and secret you want to use to create a client:

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

2. Create a CR with the **secretName** property set to the name of Kubernetes Secret you created. Run this command to trigger the creation of a client:

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
     redirectUris: ["{URI1}" , "{URI2}"]
   EOF
   ```

   >**NOTE:** This sample OAuth2Client CR has a redirect URI defined through the optional **spec.redirectUris** property. See the [CRD](https://github.com/ory/hydra-maester/blob/master/config/crd/bases/hydra.ory.sh_oauth2clients.yaml) for more details.  

### Use Hydra-generated credentials

Run this command to create a CR that triggers the creation of a client. The OAuth2 server generates a client ID and secret pair and saves it to a Kubernetes secret with the name specified in the **secretName** property.

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

### Get the registered client credentials

Run this command to get the credentials of the registered OAuth2 client:

```
kubectl get secret -n {CLIENT_NAMESPACE} {NAME_OF_KUBERNETES_SECRET} -o yaml
```

### Update the OAuth2 client credentials

If the credentials of your OAuth2 client are compromised, follow these steps to change them to a new pair:

1. Create a new Kubernetes Secret with a new client ID and client secret.
2. Edit the instance of the client's corresponding `oauth2clients.hydra.ory.sh/v1alpha1` CR by replacing the value of the **SecretName** property with the name of the newly created Secret.

>**TIP:** When you complete these steps, remember to delete the Secret that stores the old client credentials.

## OAuth2 server in action

To see the OAuth2 server in action, complete the [tutorial](/components/api-gateway/#tutorials-tutorials) which shows you how to expose a service, secure it with OAuth2 tokens, and interact with it using the registered client.  

You can also interact with the OAuth2 server using its REST API. Read the official [ORY documentation](https://www.ory.sh/docs/hydra/sdk/api) to learn more about the available endpoints.

>**TIP:** If you have any questions about the ORY–Kyma integration, you can ask them on the **#security** [Slack channel](http://slack.kyma-project.io/) and get answers directly from Kyma developers.

## Configuration guidelines

### OAuth2 client data persistence

To prevent data loss, the OAuth2 server stores the registered client data in a database. By default, Kyma comes with a pre-configured in-cluster PostgreSQL database that requires no manual setup. This configuration is not, however, considered production-ready and we recommend using an external database. [This section](#configuration-o-auth2-server-profiles) provides guidance on migrating your OAuth2 server to the persistence mode of your choice, and describes the migration mechanism itself.

### The `ory-hydra-credentials` Secret

To establish a connection with a database, Hydra needs a set of credentials provided by the user as Helm overrides. Depending on the desired persistence mode, some of those values are also required to configure optional ORY sub-charts, i.e. the PostgreSQL database and Gcloud proxy mechanism. To reduce the number of in-cluster Kubernetes Secrets and to avoid confusion, the components involved follow the single Secret policy. Namely, they all use the `ory-hydra-credentials` Secret as the only source of credentials. Being an ORY-related object, the Secret resides in the `kyma-system` Namespace.

>**TIP:** We strongly recommend backing up this Secret after every installation or upgrade procedure.

### Reaping the parameters

To ensure that the OAuth2 server is configured properly, Helm runs a preliminary job prior to the ORY chart installation or upgrade. This job combines the overrides containing credentials into one Kubernetes Secret accessible to all components involved in the persistence mechanism. The job is also responsible for identifying missing overrides, if any. If a required override has not been specified, the job either falls back to an alternative source of credentials, or fails and logs the missing override key, thus interrupting the installation or upgrade procedure.

#### Prioritization of parameters

This list presents the priority of parameters in descending order:

1. Use the overrides provided by the user before the installation or update process.

2. Reuse the parameters stored in existing Kubernetes Secrets and accessible to the job's container through a volume mount. This option is available only for the update process.

   >**CAUTION:** The initial implementation of OAuth2 client persistence in Kyma doesn't follow the single Secret policy, but rather distributes the credentials per component. This mechanism is backward-compatible. However, we recommend removing the deprecated Secrets after upgrading. Deprecated Secrets are: `ory-hydra`, `ory-postgres`, and `ory-gcloud-sqlproxy`.

3. Generate random values or fail, depending on the nature of a given value.

The following table lists all the possible keys aggregated in the `ory-hydra-credentials` Secret, along with their fallback policies.

| Secret | Override | Fallback policy |
|------- |----------|-----------------|
| `dsn` | This value is constructed | User provided database parameters or default to in-memory settings |
| `secretsSystem` | `hydra.hydra.config.secrets.system` | Generate random string |
| `secretsCookie` | `hydra.hydra.config.secrets.cookie` | Generate random string |
| `gcp-sa.json` | `global.ory.hydra.persistence.gcloud.saJson` | Interrupt the procedure |

### Helm Overrides

For required overrides and their descriptions, see [this document](#configuration-o-auth2-server-profiles-persistence-modes-for-the-production-profile).
