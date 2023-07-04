---
title: Authorization in Kyma
---

## User authorization

Kyma uses the Kubernetes concept of roles. Assign roles to individual users or user groups to manage access to the cluster. If you want to access the system through Kyma Dashboard or using kubectl, you need a `kubeconfig` file with user context. User permissions are recognized depending on roles that are bound to this user and known from the `kubeconfig` context.

### Cluster-wide authorization

Roles in Kyma are defined as ClusterRoles and use the Kubernetes mechanism of aggregation, which allows you to combine multiple ClusterRoles into a single ClusterRole. Kyma comes with a set of roles that are aggregated to the main end-user roles. You can use the aggregation mechanism to efficiently manage access to Kubernetes and Kyma-specific resources.

>**NOTE:** Read the [Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) to learn more about the aggregation mechanism used to define Kyma roles.

The predefined end-user roles are:

| Role | Description |
| --- | --- |
| **kyma-essentials** | The basic role required to allow the user to access Kyma Dashboard of the cluster. This role doesn't give the user rights to modify any resources. **Note that with Kyma 2.0, the kyma-essentials role becomes deprecated.** |
| **kyma-namespace-admin-essentials** | The role that allows the user to access Kyma Dashboard and create Namespaces, built on top of the **kyma-essentials** role. |
| **kyma-view** | The role for listing Kubernetes and Kyma-specific resources. |
| **kyma-edit** | The role for editing Kyma-specific resources. It's [aggregated](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) by other roles. |
| **kyma-snapshots** | The role for managing VolumeSnapshot CR for backups. |
| **kyma-developer** | The role created for developers who build implementations using Kyma. It allows you to list, edit, and create Kubernetes and Kyma-specific resources. |
| **kyma-namespace-admin** | The role which gives access to a specific Namespace with administrative rights. |

To learn more about the default roles and how they are constructed, see the [`rbac-roles.yaml`](https://github.com/kyma-project/kyma/blob/master/resources/cluster-users/templates/rbac-roles.yaml) file.

After creating a Kyma cluster, you become an admin of this instance and the Kubernetes **cluster-admin** role is assigned to you by default. It is the role with the highest permission level which gives access to all Kubernetes and Kyma resources and components with administrative rights. As the **cluster-admin**, you can assign roles to other users.

### Role binding

Assigning roles in Kyma is based on the Kubernetes RBAC concept. You can assign any of the predefined roles to a user or to a group of users in the context of:

- The entire cluster by creating a [ClusterRoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)
- A specific Namespace by creating a [RoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)

You can use your own Identity Provider (IdP) using OpenID Connect to authenticate. Using a custom IdP enables assigning roles to a group of users. Custom IdP allows you to define user groups and assign roles to them in Kyma. In this case, a group claim from the access token is used to recognize permissions.

>**TIP:** The **ClusterRoles** and **ClusterRoleBindings** view in the **Configuration** section of Kyma Dashboard allow you to manage cluster-level bindings between user groups and roles. To manage bindings between user groups and roles in a Namespace, select the Namespace and go to **Roles** and **Role Bindings** in the **Configuration** section.

>**TIP:** To ensure proper Namespace separation, use [RoleBindings](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding) to give users access to the cluster. This way a group or a user can have different permissions in different Namespaces.

## Service-to-service authorization

Kyma uses the native [Istio Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/). The Authorization Policy enables access control on workloads in the mesh.

## User-to-service authorization

Kyma uses a custom [API Gateway](../../01-overview/api-exposure/apix-01-api-gateway.md) component that is built on top of [ORY Oathkeeper](https://www.ory.sh/oathkeeper/docs/). The API Gateway allows exposing user applications within the Kyma environment and secures them if necessary.
