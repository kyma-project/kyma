# Authorization in Kyma

## User Authorization

Kyma uses the Kubernetes concept of roles. Assign roles to individual users or user groups to manage access to the cluster. If you want to access the system through Kyma dashboard or using kubectl, you need a `kubeconfig` file with user context. User permissions are recognized depending on roles that are bound to this user and known from the `kubeconfig` context.

> [!NOTE]
> Read the [Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) to learn more about the aggregation mechanism used to define Kyma roles.

### Role Binding

Assigning roles in Kyma is based on the Kubernetes RBAC concept. You can assign any of the predefined roles to a user or to a group of users in the context of:

- The entire cluster by creating a [ClusterRoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)
- A specific namespace by creating a [RoleBinding](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)

You can use your own Identity Provider (IdP) using OpenID Connect to authenticate. Using a custom IdP enables assigning roles to a group of users. Custom IdP allows you to define user groups and assign roles to them in Kyma. In this case, a group claim from the access token is used to recognize permissions.

> [!TIP]
> The **ClusterRoles** and **ClusterRoleBindings** view in the **Configuration** section of Kyma dashboard allow you to manage cluster-level bindings between user groups and roles. To manage bindings between user groups and roles in a namespace, select the namespace and go to **Roles** and **Role Bindings** in the **Configuration** section.

> [!TIP]
> To ensure proper namespace separation, use [RoleBindings](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding) to give users access to the cluster. This way a group or a user can have different permissions in different namespaces.

## Service-To-Service Authorization

Kyma uses the native [Istio Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/). The Authorization Policy enables access control on workloads in the mesh.

## User-To-Service Authorization

The [API Gateway module](https://kyma-project.io/#/api-gateway/user/README), which is built on top of [Ory Oathkeeper](https://www.ory.sh/oathkeeper/docs/), allows exposing user applications within the Kyma environment and secures them if necessary.