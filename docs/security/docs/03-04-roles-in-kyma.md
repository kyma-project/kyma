---
title: Roles in Kyma
type: Details
---

Kyma uses roles and groups to manage access in the cluster. Every cluster comes with five predefined roles which give the assigned users different level of permissions suitable for different purposes.
These roles are defined as ClusterRoles and use the Kubernetes mechanism of aggregation which allows you to combine multiple ClusterRoles into a single ClusterRole. Use the aggregation mechanism to efficiently manage access to Kubernetes and Kyma-specific resources.

>**NOTE:** Read [this](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) Kubernetes documentation to learn more about the aggregation mechanism used to define Kyma roles.

You can assign any of the predefined roles to a user or to a group of users in the context of:  
  - The entire cluster by creating a [ClusterRoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)
  - A specific Namespace by creating a [RoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)

The predefined roles, arranged in the order of increasing access level, are:

| Role | Description |
| --- | --- |
| **kyma-essentials** | The basic role required to allow the users to see the Console UI of the cluster. This role doesn't give the user rights to modify any resources. |
| **kyma-view** | The role for listing Kubernetes and Kyma-specific resources. |
| **kyma-edit** | The role for to editing Kyma-specific resources.  |
| **kyma-developer** | The role created for developers who build implementations using Kyma. It allows you to list and edit Kubernetes and Kyma-specific resources. |
| **kyma-admin** | The role with the highest permission level which gives access to all Kubernetes and Kyma resources and components with administrative rights. |

>**NOTE:** To learn more about the default roles and how they are constructed, see [this](https://github.com/kyma-project/kyma/blob/master/resources/core/charts/cluster-users/templates/rbac-roles.yaml) file.
