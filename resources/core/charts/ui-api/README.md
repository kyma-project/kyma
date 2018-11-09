# UI API

## Security

Security in GraphQL is based on Istio Authentication Policies, Istio Role-Based Access Control (RBAC) used for authorization,
and a custom Envoy filter.

### Authentication

The Authentication Policy is configured with the default Kyma OpenID Connect provider, Dex. This means that you must use an ID token issued by Dex to access GraphQL.
The Authentication Policy is defined in [this](./templates/authentication.yaml) file.

### Authorization

Authorization relies on Istio Role-Based Access Control (RBAC).
RBAC in Istio is based on these concepts:

- **RBAC Config** defines for which services the Istio RBAC should be enabled.

- **Service Role** defines the role required to access the services in the mesh. This requirement must be fulfilled to secure the GraphQL query:

  - **metadata.name** defines the name of the role and is required to access the query results.

  - **metadata.namespace** is the Kyma installation Namespace (by default: `kyma-system`).

  - **rules.services** is the fully-qualified address of the Core UI API service (by default: `["core-ui-api.kyma-system.svc.cluster.local"]`).

  - **spec.rules[0].paths** is the GraphQL path (by default: `/graphql`).

  - **spec.rules[0].methods** indicates the secured HTTP methods. As a minimum you must secure the `POST` method.

  - **spec.rules[0].constraints** allows to specify additional conditions which can be found [here](https://istio.io/docs/reference/config/authorization/constraints-and-properties/#constraints).

- **Service Role Binding** grants a role to subjects (e.g., a user, a group, a service).

  For example, to add a permission to the query which requires `SERVICE_ROLE_REQUIRED_TO_ACCESS_QUERY` Service Role
  for the `USER_EMAIL` user, you must define the following Service Role Binding:  

  ```yaml
  apiVersion: "rbac.istio.io/v1alpha1"
  kind: ServiceRoleBinding
  metadata:
    name: BINDING_NAME
  spec:
    subjects:
    - properties:
      request.auth.claims[email]: `USER_EMAIL`
    roleRef:
      kind: ServiceRole
      name: "SERVICE_ROLE_REQUIRED_TO_ACCESS_QUERY"
  ```

For more information see the [Istio RBAC](https://istio.io/docs/concepts/security/rbac/) documentation.

### Secure the query

By default all queries are allowed. For details, see the [Current state and migration](#current-state-and-migration) section.

### Grant permissions to the query

To grant permissions to the query for a given user, group, or a service, you must create a Service Role Binding.

### Current state and migration

To enable backward compatibility a global `graphql-manage-all` Service Role is defined and assigned to all users by default.
This configuration allows all users to execute all queries in the GraphQL.

To restrict access to the queries, you must implement these changes:

1. Define more fine-grained Service Roles.

1. Assign particular Service Roles to the appropriate users, groups or services (define Service Role Bindings).

1. Delete [all-users--graphql-manage-all](./templates/servicerolebinding-manage-all-for-all-users.yaml) Service Role Binding.

1. Delete [graphql-manage-all](./templates/servicerole-manage-all.yaml) Service Role.

### Example

This example shows how to set up the authorization for queries in GraphQL.

The secured query is as follows:

```
query GetApis {
  apis(environment: "default") {
    name
  }
}
```

To create the role for the query, use the following command:

```bash
cat <<EOF | kubectl create -f -
---
apiVersion: "rbac.istio.io/v1alpha1"
kind: ServiceRole
metadata:
  name: graphql-groups-read
  namespace: kyma-system
spec:
  rules:
  - services: ["core-ui-api.kyma-system.svc.cluster.local"]
    paths: ["/graphql"]
    methods: ["*"]
EOF
```

To grant the specified user or group permissions to the query, create appropriate Service Role Binding.
The following Service Role Binding grants `graphql-groups-read` role to the `admin@kyma.cx` user.

```bash
cat <<EOF | kubectl create -f -
---
apiVersion: "rbac.istio.io/v1alpha1"
kind: ServiceRoleBinding
metadata:
  name: admin-graphql-groups-read
  namespace: kyma-system
spec:
  subjects:
    - properties:
      request.auth.claims[email]: `admin@kyma.cx`
  roleRef:
    kind: ServiceRole
    name: "graphql-groups-read"
EOF
```
