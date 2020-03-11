---
title: Manage static users in Dex
type: Tutorials
---

## Create a new static user

To create a static user in Dex, create a Secret with the **dex-user-config** label set to `true`. Run:

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name:  {SECRET_NAME}
  namespace: {SECRET_NAMESPACE}
  labels:
    "dex-user-config": "true"
data:
  email: {BASE64_USER_EMAIL}
  username: {BASE64_USERNAME}
  password: {BASE64_USER_PASSWORD}  
type: Opaque
EOF
```
>**NOTE:** If you don't specify the Namespace in which you want to create the Secret, the system creates it in the `default` Namespace.

The following table describes the fields that are mandatory to create a static user. If you don't include any of these fields, the user is not created.

|Field | Description |
|---|---|
| data.email | Base64-encoded email address used to sign-in to the console UI. Must be unique. |
| data.username | Base64-encoded username displayed in the console UI. |
| data.password | Base64-encoded user password. There are no specific requirements regarding password strength, but it is recommended to use a password that is at least 8-characters-long. |

Create the Secrets in the cluster before Dex is installed. The Dex init-container with the tool that configures Dex generates user configuration data basing on properly labeled Secrets, and adds the data to the ConfigMap.

If you want to add a new static user after Dex is installed, restart the Dex Pod. This creates a new Pod with an updated ConfigMap.

## Bind a user to a Role or a ClusterRole

A newly created static user has no access to any resources of the cluster as there is no Role or ClusterRole bound to it.  
By default, Kyma comes with the following ClusterRoles:

- **kyma-admin**: gives full admin access to the entire cluster
- **kyma-namespace-admin**: gives full admin access except for the write access to AddonsConfigurations
- **kyma-edit**: gives full access to all Kyma-managed resources
- **kyma-developer**: gives full access to Kyma-managed resources and basic Kubernetes resources
- **kyma-view**: allows viewing and listing all of the resources of the cluster
- **kyma-essentials**: gives a set of minimal view access right to use the Kyma Console

To bind a newly created user to the **kyma-view** ClusterRole, run this command:
```
kubectl create clusterrolebinding {BINDING_NAME} --clusterrole=kyma-view --user={USER_EMAIL}
```

To check if the binding is created, run:
```
kubectl get clusterrolebinding {BINDING_NAME}
```
