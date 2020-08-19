---
title: Roles in Kyma
type: Details
---

Kyma uses roles and groups to manage access in the cluster. Every cluster comes with predefined roles which give the assigned users different level of permissions suitable for different purposes.
These roles are defined as ClusterRoles and use the Kubernetes mechanism of aggregation which allows you to combine multiple ClusterRoles into a single ClusterRole. Use the aggregation mechanism to efficiently manage access to Kubernetes and Kyma-specific resources.

>**NOTE:** Read the [Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) to learn more about the aggregation mechanism used to define Kyma roles.

You can assign any of the predefined roles to a user or to a group of users in the context of:  
  - The entire cluster by creating a [ClusterRoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)
  - A specific Namespace by creating a [RoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)

>**TIP:** To ensure proper Namespace separation, use [RoleBindings](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding) to give users access to the cluster. This way a group or a user can have different permissions in different Namespaces.

The predefined roles are:

| Role | Default group | Description |
| --- | --- | --- |
| **kyma-essentials** | `runtimeDeveloper` | The basic role required to allow the users to access the Console UI of the cluster. This role doesn't give the user rights to modify any resources. |
| **kyma-namespace-admin-essentials** | `runtimeNamespaceAdmin` | The role that allows the user to access the Console UI and create Namespaces, built on top of the **kyma-essentials** role. Used to give the members of selected groups the ability to create Namespaces in which the [Permission Controller](#details-permission-controller) binds them to the **kyma-namespace-admin** role. |
| **kyma-view** | `runtimeOperator` | The role for listing Kubernetes and Kyma-specific resources. |
| **kyma-edit** | None | The role for editing Kyma-specific resources. It's [aggregated](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) by other roles. |
| **kyma-developer** | None | The role created for developers who build implementations using Kyma. It allows you to list and edit Kubernetes and Kyma-specific resources. You need to bind it manually to a user or a group in the Namespaces of your choice. Use the `runtimeDeveloper` group when you run Kyma with the default `cluster-users` chart configuration. |
| **kyma-admin** | `runtimeAdmin` | The role with the highest permission level which gives access to all Kubernetes and Kyma resources and components with administrative rights. |
| **kyma-namespace-admin** | `runtimeNamespaceAdmin` | The role that has the same rights as the **kyma-admin** role, except for the write access to [AddonsConfigurations](https://kyma-project.io/docs/master/components/helm-broker#custom-resource-addons-configuration). The [Permission Controller](#details-permission-controller) automatically creates a RoleBinding to the `runtimeNamespaceAdmin` group in all non-system Namespaces. |

>**CAUTION:** To give a user the **kyma-developer** role permissions in a Namespace, create a RoleBinding to the **kyma-developer** cluster role in that Namespace. You can define a subject of the RoleBinding by specifying either a **Group**, or a **User**. If you decide to specify a **User**, provide a user email.  

To learn more about the default roles and how they are constructed, see the [`rbac-roles.yaml`](https://github.com/kyma-project/kyma/blob/master/resources/cluster-users/templates/rbac-roles.yaml) file.
