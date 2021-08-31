---
title: Authorization in Kyma
---

## User authorization

Kyma uses roles and user groups to manage access to the cluster. If you want to access the system through Kyma Dashboard or using kubectl, you must be authenticated with a JWT token. This token collects user information such as username, email, or `groups` claim for the system to determine whether you have access to perform certain operations.

### Cluster-wide authorization

Roles in Kyma are defined as ClusterRoles and use the Kubernetes mechanism of aggregation, which allows you to combine multiple ClusterRoles into a single ClusterRole. Use the aggregation mechanism to efficiently manage access to Kubernetes and Kyma-specific resources.

>**NOTE:** Read the [Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) to learn more about the aggregation mechanism used to define Kyma roles.

The predefined roles are:

| Role | Default group | Description |
| --- | --- | --- |
| **kyma-essentials** | `runtimeDeveloper` | The basic role required to allow the user to access Kyma Dashboard of the cluster. This role doesn't give the user rights to modify any resources. |
| **kyma-namespace-admin-essentials** | `runtimeNamespaceAdmin` | The role that allows the user to access Kyma Dashboard and create Namespaces, built on top of the **kyma-essentials** role. |
| **kyma-view** | `runtimeOperator` | The role for listing Kubernetes and Kyma-specific resources. |
| **kyma-edit** | None | The role for editing Kyma-specific resources. It's [aggregated](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) by other roles. |
| **kyma-developer** | None | The role created for developers who build implementations using Kyma. It allows you to list and edit Kubernetes and Kyma-specific resources. You need to bind it manually to a user or a group in the Namespaces of your choice. Use the `runtimeDeveloper` group when you run Kyma with the default `cluster-users` chart configuration. |
| **kyma-admin** | `runtimeAdmin` | The role with the highest permission level which gives access to all Kubernetes and Kyma resources and components with administrative rights. |
| **kyma-namespace-admin** | `runtimeNamespaceAdmin` | The role that has the same rights as the **kyma-admin** role, except for the write access to [AddonsConfigurations](../../05-technical-reference/00-custom-resources/smgt-04-hb-addonsconfiguration.md).

To learn more about the default roles and how they are constructed, see the [`rbac-roles.yaml`](https://github.com/kyma-project/kyma/blob/master/resources/cluster-users/templates/rbac-roles.yaml) file.

### Role binding

You can assign any of the predefined roles to a user or to a group of users in the context of:

- The entire cluster by creating a [ClusterRoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)
- A specific Namespace by creating a [RoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)

>**TIP:** The **Global permissions** view in the **Settings** section of Kyma Dashboard allows you to manage cluster-level bindings between user groups and roles. To manage bindings between user groups and roles in a Namespace, select the Namespace and go to the **Configuration** section of the **Permissions** view.

>**TIP:** To ensure proper Namespace separation, use [RoleBindings](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding) to give users access to the cluster. This way a group or a user can have different permissions in different Namespaces.

The RoleBinding or ClusterRoleBinding must have a group specified as their subject. See the [Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) to learn how to manage Roles and RoleBindings.

## Service-to-service authorization

Kyma uses the native [Istio Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/). The Authorization Policy enables access control on workloads in the mesh.

## User-to-service authorization

Kyma uses a custom [API Gateway](../../01-overview/main-areas/api-exposure/apix-01-api-gateway.md) component that is built on top of [ORY Oathkeeper](https://www.ory.sh/oathkeeper/docs/). The API Gateway allows exposing user applications within the Kyma environment and secures them if necessary.
