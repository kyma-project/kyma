---
title: GraphQL
type: Details
---

Kyma uses a custom GraphQL implementation in the UI API Layer to secure interactions between frontends and Kubernetes resources of the cluster.

The authorization in GraphQL uses the Kyma implementation of RBAC, which means that:
  - All of the Roles, RoleBindings, ClusterRoles and CluserRoleBindings that you create and assign are effective and give the same permissions when the users interact with the cluster resources both through the CLI and the GraphQL-secured frontends.
  - To give users access to specific queries you must create appropriate Roles and bindings in your cluster.

> **NOTE:** Although Kyma GraphQL implementation uses RBAC, you can't use it to secure access to specific resources specified by their name.  

Kyma GraphQL implementation assigns actions to specific Kubernetes verbs:

| GraphQL action | Kubernetes verb(s) |
|---|---|
| query | GET (for a single resource), LIST (for multiple resources) |
| mutation | CREATE, DELETE |
| subscription | WATCH |

## Available GraphQL actions

To access cluster resources through GraphQL, an action securing given resource must be defined in the cluster.
To see the list of actions available in your cluster, run this command:
```
kubectl PLCHLDR
```

Alternatively, see the [GraphQL schema](/kyma/components/ui-api-layer/internal/gqlschema/schema.graphql) file to see the list of actions implemented in every Kyma cluster by default.
