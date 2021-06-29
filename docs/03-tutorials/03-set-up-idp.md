---
title: Set up a custom identity provider
type: Tutorials
---

Kyma sits on top of Kubernetes and leverages [authentication strategies](https://kubernetes.io/docs/reference/access-authn-authz/authentication/) from it. The purpose of all of those authentication strategies is to associate the identity of the caller with the request to the API server and evaluate access based on roles (RBAC).

One of the strategies allows you to use your own identity provider. This is very convenient because you can delegate the verification of who the users are to a separate user management entity and even re-use it in different systems.

> **NOTE:** Kubernetes supports OpenID Connect (OIDC) JWT Access tokens, so make sure your identity provider is OIDC-compliant.

## Prerequisites

1. Kubeconfig file to your Kyma cluster
2. OIDC-compliant identity provider

## Steps

### Configure your identity provider

> **NOTE:** If you don't have access to the identity provider, you can sign up for a free tier plan at [Auth0](https://auth0.com/).

Configure a dedicated client (often referred to as an application) at your identity provider.

1. Note down these details of your application at your identity provider:

- `issuerUrl`
- `clientId` 
- `clientSecret` 

2. Add `http://localhost:8000` to allowed redirect URIs that are required for the OIDC login plugin.
3. Configure the name of the `username` and `group` claims.
4. Enable the Proof Key for Code Exchange (PKCE) authentication flow.

### Configure your identity provider as the OIDC server

In general, you have to add flags to the API server as described in the [Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#configuring-the-api-server). You will do it in different ways depending on your Kubernetes distribution.
For example, if you want to use  k3d, you need to pass additional `--k3s-server-arg` flags containing the OIDC server configuration when creating the cluster. See the [specification](https://k3d.io/usage/commands/k3d_cluster_create/) of the `k3d cluster create` command:

```bash
k3d cluster create kyma \
--k3s-server-arg "--kube-apiserver-arg=oidc-issuer-url=<your-ipd-issuer-url>" \
--k3s-server-arg "--kube-apiserver-arg=oidc-username-claim=<username-claim-at-your-ipd>" \
--k3s-server-arg "--kube-apiserver-arg=oidc-client-id=<your-ipd-client-id>" \
--k3s-server-arg "--kube-apiserver-arg=oidc-groups-claim=<group-claim-at-your-ipd>" \
```

If you use managed Kubernetes, you will do it differently depending on your provider.
For example, if you use Gardener as a managed Kubernetes offering, you will probably want to look at the [OIDC Preset](https://github.com/gardener/gardener/blob/master/docs/usage/openidconnect-presets.md) resources that will help you with the task.



### Configure role-based access to identities provided by your OIDC server

Now, whenever api server is called with a JWT token, the api server will be able to validate and extract the associated identity ( from the `username` and `group` claims of the JWT token).

You must now define which individuals or groups should have access to which Kyma resources because the default setup does not provide access to any. You need to model permissions using the [RBAC concept](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

By default, Kyma comes with the following ClusterRoles:

- **kyma-admin**: gives full admin access to the entire cluster
- **kyma-namespace-admin**: gives full admin access except for the write access to [AddonsConfigurations](/components/helm-broker#custom-resource-addons-configuration)
- **kyma-edit**: gives full access to all Kyma-managed resources
- **kyma-developer**: gives full access to Kyma-managed resources and basic Kubernetes resources
- **kyma-view**: allows viewing and listing all of the resources in the cluster
- **kyma-essentials**: gives a set of minimal view access right to use in Kyma Dashboard

To bind a user to the **kyma-admin** ClusterRole, run this command:

```
kubectl create clusterrolebinding {BINDING_NAME} --clusterrole=kyma-admin --user={USERNAME AS IDENTIFIED AT YOUR IDP}
```

To check if the binding is created, run:

```
kubectl get clusterrolebinding {BINDING_NAME}
```

To bind a group of users to the **kyma-admin** ClusterRole, run this command:

```
kubectl create clusterrolebinding {BINDING_NAME} --clusterrole=kyma-admin --group={GROUPNAME}
```

You can combine user and group level permissions in one binding. Run `kubectl create clusterrolebinding --help` in your terminal for more options.

### Configure kubectl access

With this step, you will set up the OIDC provider in the kubeconfig file to enforce authentication flow when accessing Kyma using `kubectl`.

1. Install the [kubelogin](https://github.com/int128/kubelogin) plugin.
2. Copy your current kubeconfig file into a new file.
3. In the new kubeconfig file, define a new OIDC user and set up the OIDC provider. 

    ```yaml
    users:
    - name: oidc
    user:
        exec:
        apiVersion: client.authentication.k8s.io/v1beta1
        command: kubectl
        args:
        - oidc-login
        - get-token
        - --oidc-issuer-url=ISSUER_URL
        - --oidc-client-id=YOUR_CLIENT_ID
        #- --oidc-client-secret=YOUR_CLIENT_SECRET this is not required if your OICS server supports the PKCE authentication flow
    ```
4. To enforce the OIDC login, set the OIDC user as a default user in the context.
    ```yaml
    context:
        cluster: {YOUR_CLUSTER_NAME}
        user: oidc
    ```
5. Now you can share the modified kubeconfig file with the members of your team or organization. When they use it, your identity provider will handle the authentication. The Kubernetes API server will make sure they will have access to resources according to the role bound to them as individuals or group members.     
