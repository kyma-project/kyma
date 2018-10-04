```
 _   _ ___      _    ____ ___ 
| | | |_ _|    / \  |  _ \_ _|
| | | || |    / _ \ | |_) | | 
| |_| || |   / ___ \|  __/| | 
 \___/|___| /_/   \_\_|  |___|
 
```

## Security

Security in GraphQL is based on Istio Authentication Policies, Istio Role-Based Access Control (RBAC) used for authorization, 
and a custom Envoy filter. 

### Authentication

The Authentication Policy is configured with the default Kyma OpenID Connect provider, Dex. This means that you must use an ID token issued by Dex to access GraphQL.
The Authentication Policy is defined in [this](./templates/authentication.yaml) file.

### Custom Envoy filter

The Envoy filter extracts security-related attributes from GraphQL queries. These attributes are added to the request as additional 
 headers and are sent to the target service. These are the additional request headers:

- **kyma-graphql-parsed** - this header is set if the filter parsed the query.

- **kyma-graphql-resources** - this header contains the names of the queried resources. The names are comma-separated
  and sorted in alphabetical order. The list is enclosed with curly braces.

The Envoy filter is registered in the pilot's discovery container as a webhook. The source code of the Envoy filter, in the form of a LUA script, is located in [this](../../../istio/charts/webhook/) directory.

### Authorization

Authorization relies on Istio Role-Based Access Control (RBAC).
RBAC in Istio is based on these concepts:

- **Authorization Request Context** defines the attributes that are extracted from the request and sent to the
  RBAC Engine, which is implemented as Istio Policy/Mixer adapter. Headers set by the custom Envoy filter 
  are sent to the RBAC Engine. This enables to define constraints on queries in a Service Role.
  The Authorization Request Context is defined in [this](./templates/authorization.yaml) file.

- **RBAC Rule** Mixer adapter rule used to enable RBAC for GraphQL, and to associate Authorization Request Context,
  which should be used for authorization.
  RBAC Rule is defined in [this](./templates/authorization.yaml) file.

- **Service Role** defines the role required to access the services in the mesh. This requirement must be fulfilled to secure the GraphQL query:

  - **metadata.name** defines the name of the role and is required to access the query results.

  - **metadata.namespace** is the Kyma installation Namespace (by default: `kyma-system`).

  - **rules.services** is the fully-qualified address of the Core UI API service (by default: `["core-ui-api.kyma-system.svc.cluster.local"]`).
  
  - **spec.rules[0].paths** is the GraphQL path (by default: `/graphql`).
  
  - **spec.rules[0].methods** indicates the secured HTTP methods. As a minimum you must secure the `POST` method.
  
  - **spec.rules[0].constraints** indicates the constraints:
  
     - **kymaGraphQlResources** - The value should contain the names of the queried resources. List the resources in alphabetical order.
       Enclose the list in curly brackets. Value extracted from the **kyma-graphql-resources** header.

  For example, to define a role for access to the query with `apis` resource you must define the following Service Role:

  ```yaml
  apiVersion: "config.istio.io/v1alpha2"
  kind: ServiceRole
  metadata:
    name: SERVICE_ROLE_REQUIRED_TO_ACCESS_QUERY
  spec:
    rules:
    - services: ["core-ui-api.kyma-system.svc.cluster.local"]
      paths: ["/graphql"]
      methods: ["*"]
      constraints:
      - key: "kymaGraphQlResources"
        values: ["{apis}"]
  ```

- **Service Role Binding** grants a role to subjects (e.g., a user, a group, a service).

  For example, to add a permission to the query which requires `SERVICE_ROLE_REQUIRED_TO_ACCESS_QUERY` Service Role
  for the `USERNAME` user, you must define the following Service Role Binding:  

  ```yaml
  apiVersion: "config.istio.io/v1alpha2"
  kind: ServiceRoleBinding
  metadata:
    name: BINDING_NAME
  spec:
    subjects:
    - user: "USERNAME"
    roleRef:
      kind: ServiceRole
      name: "SERVICE_ROLE_REQUIRED_TO_ACCESS_QUERY"
  ```

For more information see the [Istio RBAC](https://istio.io/docs/concepts/security/rbac/) documentation.

### Secure the query

By default all queries are restricted. For details, see the [Current state and migration](#current-state-and-migration) section.
To enable access to the query, one has to create a Service Role.

### Grant permissions to the query

To grant permissions to the query for a given user, group, or a service, you must create a Service Role Binding.

### Current state and migration

To enable backward compatibility a global `graphql-manage-all` Service Role is defined and assigned to all users by default.
This configuration allows all users to execute all queries in the GraphQL.

To restrict access to the queries, you must mplement these changes:

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

To secure query by queried resources, use the `kymaGraphQlResources` constraint as follows:

```bash
cat <<EOF | kubectl create -f -
---
apiVersion: "config.istio.io/v1alpha2"
kind: ServiceRole
metadata:
  name: graphql-groups-read
  namespace: kyma-system
spec:
  rules:
  - services: ["core-ui-api.kyma-system.svc.cluster.local"]
    paths: ["/graphql"]
    methods: ["*"]
    constraints:
    - key: "kymaGraphQlResources"
      values: ["{remoteEnvironments}"]
EOF
```

To grant the specified user or group permissions to the query, create appropriate Service Role Binding.
The following Service Role Binding grants `graphql-groups-read` role to the `admin@kyma.cx` user.

```bash
cat <<EOF | kubectl create -f -
---
apiVersion: "config.istio.io/v1alpha2"
kind: ServiceRoleBinding
metadata:
  name: admin-graphql-groups-read
  namespace: kyma-system
spec:
  subjects:
  - user: "admin@kyma.cx"
  roleRef:
    kind: ServiceRole
    name: "graphql-groups-read"
EOF
```