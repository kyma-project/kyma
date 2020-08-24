---
title: Authorization in Kyma
type: Details
---

## User authorization
Kyma uses roles and user groups to manage access in the cluster. Each user accessing the system (whether by the Console or CLI) must introduce themselves with a JWT token collecting their user information (username, email, groups or claims) which is used to determine whether the user has access to perform certain operations.

You can assign any of the predefined roles to a user or to a group of users in the context of:  
  - The entire cluster by creating a [ClusterRoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)
  - A specific Namespace by creating a [RoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)

Every cluster comes with predefined roles which give the assigned users different level of permissions suitable for different purposes:
  - **kyma-essentials**
  - **kyma-view**
  - **kyma-namespace-admin-essentials**
  - **kyma-edit**
  - **kyma-developer**
  - **kyma-admin**
  - **kyma-namespace-admin**

Read more about [roles in Kyma](#details-roles-in-kyma).

>**NOTE:** The **Global permissions** view in the **Settings** section of the Kyma Console UI allows you to manage cluster-level bindings between user groups and roles. To manage bindings between user groups and roles in a Namespace, select the Namespace and go to the **Configuration** section of the **Permissions** view.

>**TIP:** To ensure proper Namespace separation, use [RoleBindings](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding) to give users access to the cluster. This way a group or a user can have different permissions in different Namespaces.

The RoleBinding or ClusterRoleBinding must have a group specified as their subject.
See the [Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) to learn how to manage Roles and RoleBindings.

>**NOTE:** You cannot define groups for the static user store. Instead, bind the user directly to a role or a cluster role by setting the user as the subject of a RoleBinding or ClusterRoleBinding.

### Cluster wide
Roles in Kyma are defined as ClusterRoles and use the Kubernetes mechanism of aggregation which allows you to combine multiple ClusterRoles into a single ClusterRole. Use the aggregation mechanism to efficiently manage access to Kubernetes and Kyma-specific resources.

>**NOTE:** Read the [Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) to learn more about the aggregation mechanism used to define Kyma roles.

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

### Namespace wide
To ensure that each clients namespace in the system is accessible for the user, a Permission Controller is deployed within the cluster.

The Permission Controller is a Kubernetes controller which listens for new Namespaces and creates RoleBindings for the users of the specified group to the **kyma-namespace-admin** role within these Namespaces. The Controller uses a blacklist mechanism, which defines the Namespaces in which the users of the defined group are not assigned the **kyma-admin** role.

When the Controller is deployed in a cluster, it checks all existing Namespaces and assigns the roles accordingly.

By default, the controller binds users of the **runtimeNamespaceAdmin** group to the **kyma-namespace-admin** role in the Namespaces they create. Additionally, the controller creates a RoleBinding for the static `namespace.admin@kyma.cx` user to the **kyma-admin** role in every Namespace that is not blacklisted.

>**NOTE:** This allows clients to manage their namespaces and create additional bindings to suit their needs. 

## Service to Service authorization
As Kyma is build on top of the Istio Service Mesh, we support the native [Istio RBAC](https://archive.istio.io/v1.4/docs/reference/config/security/istio.rbac.v1alpha1/) mechanism provided by the mesh. The RBAC enabled the creation of `ServiceRoles` and `ServiceRoleBindings` which enable a fine grained method of restricting access to services inside the kubernetes cluster. 
For more details on Istio RBAC please read [this section](/components/service-mesh/#details-istio-rbac-configuration)

## Authorization in API Gateway
Kyma uses a custom Api-Gateway component, which is build on top of [ORY Oathkeeper](https://www.ory.sh/oathkeeper/docs/). It is used to streamline the process of exposing user applications within the Kyma environment, and securing them if necessary. 

For more details on the Api-Gateway please read [this section](components/api-gateway#details-details)
