---
title: Overview
type: Overview
---

The security model in Kyma uses the [Service Mesh](../../service-mesh/docs/001-overview.md) component to enforce authorization through [Kubernetes Role Based Authentication](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) (RBAC) in the cluster. The identity federation is managed through [Dex](https://github.com/coreos/dex), which is an open-source, OpenID Connect identity provider.

## Authentication

Dex implements a system of connectors that allow you to delegate authentication to external OpenID Connect and SAML2-compliant Identity Providers and use their user stores. See [this](./005-details-add-connector.md) document to learn how to enable authentication with an external Identity Provider by using a Dex connector.

Out of the box, Kyma comes with its own static user store used by Dex to authenticate users. This solution is designed for use with local Kyma deployments as it keeps the predefined users' credentials in an easily available ConfigMap file.
See [this](./006-details-manage-static-users.md) document to learn how to manage users in the static store used by Dex.

## Authorization

Kyma uses a group-based approach to managing authorizations.
To give users that belong to a group access to resources in Kyma, you must create:

- Role and RoleBinding - for resources in a given Namespace or Environment.
- ClusterRole and ClusterRoleBinding - for resources available in the entire cluster.

The RoleBinding or ClusterRoleBinding must have a group specified as their subject. 
See [this](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) document to learn how to manage Roles and RoleBindings.

>**NOTE:** You cannot define groups for the static user store. Instead, bind the user directly to a role or a cluster role by setting the user as the subject of a RoleBinding or ClusterRoleBinding.

The system creates two default roles in every Environment:
- `kyma-admin-role` - this role gives the user full access to the Environment.
- `kyma-reader-role` - this role gives the user the right to read all resources in the given Environment.

For more details about Environments, see [this](../../kyma/docs/005-environments.md) document.

>**NOTE:** The **Global permissions** section in the **Administration** view of the Kyma Console UI allows you to manage user-group bindings.
