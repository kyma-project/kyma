---
title: Roles in Kyma
type: Details
---

Kyma uses roles and groups to manage access in the cluster. Every cluster comes with five predefined roles which give the assigned users different level of permissions suitable for different purposes.
These roles are defined as ClusterRoles and use the Kubernetes mechanism of aggregation, which allows to combine multiple ClusterRoles into a single ClusterRole. Using the aggregation mechanism allows to efficiently manage access to Kubernetes resources and Kyma-specific resources.

>**NOTE:** Read [this](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) Kubernetes documentation to learn more about the aggregation mechanism used to define Kyma roles.

You can assign any of the predefined roles to a user or to a group of users in the context of:  
  - the entire cluster, creating a [ClusterRoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)
  - a specific Namespace, creating a [RoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)

The predefined roles, arranged in the order of increasing access level, are:

| Role | Description |
| --- | --- |
| **kyma-essentials** | The basic role required to allow the users to see the Console UI of the cluster. This role doesn't give the user rights to manipulate any resources. |
| **kyma-view** | The role allowing to get and list Kubernetes resources and Kyma-specific resources. |
| **kyma-edit** | The role allowing to edit Kyma-specific resources.  |
| **kyma-developer** | The role created for developers who build implementations using Kyma. Allows to edit, get, and list Kubernetes resources and Kyma-specific resources. |
| **kyma-admin** | The role with the highest permission level allowing to access all Kubernetes and Kyma resources and components with administrative rights. |

>**NOTE:** To learn more about the default roles and how they are constructed, see [this](https://github.com/kyma-project/kyma/blob/master/resources/core/charts/cluster-users/templates/rbac-roles.yaml) file.
