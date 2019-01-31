---
title: GraphQL
type: Details
---

Kyma uses a custom GraphQL implementation in the UI API Layer and deploys an RBAC-based logic to control access to the GraphQL endpoint.

The authorization in GraphQL uses RBAC, which means that:
  - All of the Roles, RoleBindings, ClusterRoles and CluserRoleBindings that you create and assign are effective and give the same permissions when the users interact with the cluster resources both through the CLI and the GraphQL endpoints.
  - To give users access to specific queries you must create appropriate Roles and bindings in your cluster.

The implementation assigns GraphQL actions to specific Kubernetes verbs:

| GraphQL action | Kubernetes verb(s) |
|---|---|
| query | GET (for a single resource), LIST (for multiple resources) |
| mutation | CREATE, DELETE |
| subscription | WATCH |

> **NOTE:** Due to the nature of Kubernetes, you can secure specific resources specified by their name only for queries and mutations. Subscriptions work only for entire resource groups, such as kinds, and therefore don't allow for such level of granularity.

## Available GraphQL actions

To access cluster resources through GraphQL, an action securing given resource must be defined and implemented in the cluster.
See the [GraphQL schema](/kyma/components/ui-api-layer/internal/gqlschema/schema.graphql) file to see the list of actions implemented in every Kyma cluster by default.

This is an example GraphQL query implemented in Kyma out of the box. This query secures the access to IDPPreset custom resources. When used by a user whose permissions match these defined by the [directive](https://graphql.org/learn/queries/#directives), the query returns a single IDPPreset custom resource with the specified name which exists in the defined Namespace.
```
IDPPreset(name: String!): IDPPreset @HasAccess(attributes: {resource: "IDPPreset", verb: "get", apiGroup: "authentication.kyma-project.io", apiVersion: "v1alpha1"})
```
