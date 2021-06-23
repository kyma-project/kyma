---
title: Set up Custom Identity Provider in Kyma
type: Tutorials
---

Kyma sits on top of kubernetes and therefore it leverages [authentication strategies](https://kubernetes.io/docs/reference/access-authn-authz/authentication/) from kubernetes. The purpose for all of them is to associate idendity of the caller with the request to the API server and evalueate access based on RBAC.

One of the strategies allows you to use your own identity provider. This is very convinient because it allows you to delegate the verification of who the users are to a separate user management entity and even re-use it in different systems.

> **NOTE:** Kubernetes supports OpenID Connect JWT Access tokens so please make sure your identity provider is OpenID connect compliant.

## Prerequisites

1. Kubeconfig file to your kyma cluster
2. OIDC compliant Identity provider

## Steps

### Configure your Identity Provider

> **NOTE:** If you dont have access to Identity Provider you can sign up for a free tier plan at [Auth0](https://auth0.com/)

Configure a dedicated client ( often refered as application ) at your identity provider.

1. Note down the `issuerUrl` of your application at your identity provider
2. Note down the `clientId` of the application at your identity provider
3. Note down the `clientSecret` of the application at your identity provider
4. Add `http://localhost:8000` to allowed redirect URIs ( required for oidc-login plugin )
5. Configure the name of the `username` claim
6. Configure the name of the `group` claim
7. Enable PKCE authentication flow 

### Configure your Identity Provider as OIDC server

In general you have to add flags to the api server, as described [here](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#configuring-the-api-server). You will do it in a different way depending on your kubernetes distribution.
For example, if you want to use `k3d` you need to pass extra `--k3s-server-arg` flags containing oidc server configuration at the cluster creation. See [specs](https://k3d.io/usage/commands/k3d_cluster_create/) of the `k3d cluster create` command, i.e:

```bash
k3d cluster create kyma \
--k3s-server-arg "--kube-apiserver-arg=oidc-issuer-url=<your-ipd-issuer-url>" \
--k3s-server-arg "--kube-apiserver-arg=oidc-username-claim=<username-claim-at-your-ipd>" \
--k3s-server-arg "--kube-apiserver-arg=oidc-client-id=<your-ipd-client-id>" \
--k3s-server-arg "--kube-apiserver-arg=oidc-groups-claim=<group-claim-at-your-ipd>" \
```

Now, whenever api server is called with a JWT token, the api server will be able to validate and extract the associated identity ( from the `username` claim ) and groups ( from the `group` claim) from the token.


### Configure Role Based Access to identies provided by your OIDC server

Now, after having completed the steps so far, every request to api server with JWT token attached is automatically associated with identity string of the user and optionally a set of groups to which the user belongs. You shoud now define which individuals or groups should have access to which kyma resources because by default they won't have access to anything. You need to modelpermissions using the concept of Role Based Access Control ([RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)).

By default, Kyma comes with the following ClusterRoles:

- **kyma-admin**: gives full admin access to the entire cluster
- **kyma-namespace-admin**: gives full admin access except for the write access to [AddonsConfigurations](/components/helm-broker#custom-resource-addons-configuration)
- **kyma-edit**: gives full access to all Kyma-managed resources
- **kyma-developer**: gives full access to Kyma-managed resources and basic Kubernetes resources
- **kyma-view**: allows viewing and listing all of the resources of the cluster
- **kyma-essentials**: gives a set of minimal view access right to use the Kyma Console

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

You can combine user and group level permission in one binding. Run `kubectl create clusterrolebinding --help` in your terminal for more options.

### Configure kubectl access

With this step you will set up the OIDC provider in  kubeconfig file to enforce authentication flow.

1. Install [kubelogin](https://github.com/int128/kubelogin) plugin
2. Copy your current kubeconfig file into a new file
3. In the new kubeconfig file define a new `oidc` user and setup OIDC provider, as follows: 

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
        #- --oidc-client-secret=YOUR_CLIENT_SECRET this is not required if your OICS server supports PKCE authentication flow
    ```
4. To enforce oidc login set the `oidc` user as a default user in the context
    ```yaml
    context:
        cluster: {your cluster name}
        user: oidc
    ```
5. Now you can share the modified kubeconfig file to members of your team / organisation. When they will use it, your identity provider will do the authentication for you and the kubernetes api server will make sure they will have access to resources according to the role bound to them as individuals or group members.     