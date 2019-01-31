---
title: GraphQL
type: Details
---

Kyma uses a custom GraphQL implementation in the UI API Layer and deploys an RBAC-based logic to control access to the GraphQL endpoint.

The authorization in GraphQL uses RBAC, which means that:
  - All of the Roles, RoleBindings, ClusterRoles and CluserRoleBindings that you create and assign are effective and give the same permissions when the users interact with the cluster resources both through the CLI and the GraphQL endpoints.
  - To give users access to specific queries you must create appropriate Roles and bindings in your cluster.

> **NOTE:** Although Kyma GraphQL implementation uses RBAC, you can't use it to secure access to specific resources based on their name. You can secure all custom resources of the IDPPreset kind, but you cannot secure all custom resources with a name "supereagle".

The implementation assigns GraphQL actions to specific Kubernetes verbs:

| GraphQL action | Kubernetes verb(s) |
|---|---|
| query | GET (for a single resource), LIST (for multiple resources) |
| mutation | CREATE, DELETE |
| subscription | WATCH |

## Available GraphQL actions

To access cluster resources through GraphQL, an action securing given resource must be defined and implemented in the cluster.
See the [GraphQL schema](/kyma/components/ui-api-layer/internal/gqlschema/schema.graphql) file to see the list of actions implemented in every Kyma cluster by default.
