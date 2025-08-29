# Set Up a Custom Identity Provider

Kyma sits on top of Kubernetes and leverages [authentication strategies](https://kubernetes.io/docs/reference/access-authn-authz/authentication/) from it. The purpose of all of those authentication strategies is to associate the identity of the caller with the request to the API server and evaluate access based on roles (RBAC).

One of the strategies enables you to use your own identity provider. This is very convenient because you can delegate the verification of who the users are to a separate user management entity and even reuse it in different systems.

> [!NOTE]
>  Kubernetes supports OpenID Connect (OIDC) JWT Access tokens, so make sure your identity provider is OIDC-compliant.

## Prerequisites

* Generate a `kubeconfig` file for the Kubernetes cluster that hosts the Kyma instance.
* Use an OIDC-compliant identity provider.

## Steps

### Configure Your Identity Provider

> [!NOTE]
>  If you don't have access to the identity provider, you can sign up for a free tier plan at [Auth0](https://auth0.com/).

Configure a dedicated client (often referred to as an application) in your identity provider.

1. Save these details of your application at your identity provider:

- `issuerUrl`
- `clientId`
- `clientSecret`

2. Add `http://localhost:8000` to the allowed redirect URLs that are required for the OIDC login plugin.
3. Configure the name of the **username** and **group** claims.
4. Enable the Proof Key for Code Exchange (PKCE) authentication flow.

### Configure Your Identity Provider as the OIDC Server

Add flags to the API server as described in the [Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#configuring-the-api-server). The ways of adding the flags to the API server differ depending on the Kubernetes distribution you use.
For example, if you want to use k3d, you need to pass the additional `--k3s-server-arg` flags containing the OIDC server configuration when creating the cluster. See the [specification](https://k3d.io/v5.1.0/usage/commands/k3d_cluster_create/) of the `k3d cluster create` command:

```bash
k3d cluster create kyma \
--k3s-server-arg "--kube-apiserver-arg=oidc-issuer-url=<your-ipd-issuer-url>" \
--k3s-server-arg "--kube-apiserver-arg=oidc-username-claim=<username-claim-at-your-ipd>" \
--k3s-server-arg "--kube-apiserver-arg=oidc-client-id=<your-ipd-client-id>" \
--k3s-server-arg "--kube-apiserver-arg=oidc-groups-claim=<group-claim-at-your-ipd>" \
```

For managed Kubernetes, see the documentation related to your provider.
For example, if you use Gardener as a managed Kubernetes offering, see the [OIDC Preset](https://github.com/gardener/gardener/blob/master/docs/usage/security/openidconnect-presets.md) documentation.

### Configure Role-Based Access to Identities Provided by Your OIDC Server

Including the JWT token in the call to the API server enables the API server to validate and extract the associated identity from the **username** and **group** claims of the JWT token.

Now, define which individuals or groups should have access to which Kyma resources. The default setup does not provide access to any. You need to model permissions using the [RBAC concept](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

You can combine user-level and group-level permissions in one binding. Run `kubectl create clusterrolebinding --help` in your terminal to see more options.

### Configure kubectl Access

With this step, you will set up the OIDC provider in the `kubeconfig` file to enforce authentication flow when accessing Kyma using `kubectl`.

1. Install the [kubelogin](https://github.com/int128/kubelogin) plugin.
2. Copy your current `kubeconfig` file into a new file.
3. In the new `kubeconfig` file, define a new OIDC user and set up the OIDC provider.

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
    ```
    > [!NOTE]
    > `--oidc-client-secret=YOUR_CLIENT_SECRET` is not required if your OICS server supports the PKCE authentication flow.

4. To enforce the OIDC login, set the OIDC user as a default user in the context.

    ```yaml
    context:
        cluster: {YOUR_CLUSTER_NAME}
        user: oidc
    ```

5. Now, you can share the modified kubeconfig file with the members of your team or organization. When they use it, your identity provider will handle the authentication. The Kubernetes API server will make sure they have access to resources according to the roles bound to them as individuals or group members.
