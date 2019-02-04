---
title: GraphQL
type: Details
---

Kyma uses a custom [GraphQL](https://graphql.org/) implementation in the UI API Layer and deploys an RBAC-based logic to control access to the GraphQL endpoint.

The authorization in GraphQL uses RBAC, which means that:
  - All of the Roles, RoleBindings, ClusterRoles and CluserRoleBindings that you create and assign are effective and give the same permissions when the users interact with the cluster resources both through the CLI and the GraphQL endpoints.
  - To give users access to specific queries you must create appropriate Roles and bindings in your cluster.

The implementation assigns GraphQL actions to specific Kubernetes verbs:

| GraphQL action | Kubernetes verb(s) |
|:---|:---|
| **query** | GET (for a single resource), LIST (for multiple resources) |
| **mutation** | CREATE, DELETE |
| **subscription** | WATCH |

> **NOTE:** Due to the nature of Kubernetes, you can secure specific resources specified by their name only for queries and mutations. Subscriptions work only with entire resource groups, such as kinds, and therefore don't allow for such level of granularity.

## Available GraphQL actions

To access cluster resources through GraphQL, an action securing given resource must be defined and implemented in the cluster.
See the [GraphQL schema](https://github.com/kyma-project/kyma/blob/master/components/ui-api-layer/internal/gqlschema/schema.graphql) file to see the list of actions implemented in every Kyma cluster by default.

### Structure of a defined GraphQL action

This is an example GraphQL query implemented in Kyma out of the box. This query secures the access to IDPPreset custom resources. When used by a user whose permissions match these defined by the [directive](https://graphql.org/learn/queries/#directives), the query returns a single IDPPreset custom resource with the specified name.
```
IDPPreset(name: String!): IDPPreset @HasAccess(attributes: {resource: "IDPPreset", verb: "get", apiGroup: "authentication.kyma-project.io", apiVersion: "v1alpha1"})
```

| Defined GraphQL action element | Description |
|:----------|:------|
| `IDPPreset(name: String!)` |  Name of the query followed by a string that specifies the name of the queried resource. |
| `: IDPPreset` | Defines the type of object returned by the query. In this case it's a single IDPPreset custom resource. |
| `@HasAccess(attributes:` | Defines the GraphQL directive that secures the access to the resource. |
| `resource: "IDPPreset"` | Defines the type of secured Kubernetes resource, in this case all IDPPreset custom resources. |
| `verb: "get"` | Defines the secured interaction type. Defines the GraphQL action type, which in this case is "query". |
| `apiGroup: "authentication.kyma-project.io"` | Defines the apiGroup to which the user must have access to get the result of this query. |
| `apiVersion: "v1alpha1"` | Specifies the apiVersion of the query subject. |

### Secure a defined GraphQL action

To allow access to the example query, create an RBAC role in the cluster and bind it to a user or a client. This role allows access specifically to the example query:

```
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: kyma-idpp-query-example
  labels:
    app: kyma
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
  annotations:
    "helm.sh/hook-weight": "0"
rules:
- apiGroups: ["authentication.kyma-project.io"]
  resources: ["IDPPresets"]
  verbs: ["get"]
```
