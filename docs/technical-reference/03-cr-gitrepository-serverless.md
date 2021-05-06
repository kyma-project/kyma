---
title: GitRepository
type: Custom Resources
---

The `gitrepositories.serverless.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define and manage Git repositories that store the Function's source code and dependencies. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd gitrepositories.serverless.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample custom resource that creates a GitRepository object pointing to a Git repository with the Function's source code and dependencies. This resource specifies that the repository requires an SSH key to authenticate to it and points to the Secret that stores these credentials.

```yaml
apiVersion: serverless.kyma-project.io/v1alpha1
kind: GitRepository
metadata:
  name: sample-with-auth
spec:
  url: "git@github.com:sample-organization/sample-repo.git"
  auth:
    type: key
    secretName: kyma-git-creds
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter | Required for HTTP(S) | Required for SSH | Description |
|-----------|-------------|-------------|---------------|
| **spec.url** | Yes | Yes | Provides the address to the Git repository with the Function's code and dependencies. Depending on whether the repository is public or private and what authentication method is used to access it, the URL must start with the `http(s)`, `git`, or `ssh` prefix, and end with the `.git` suffix. |
| **spec.auth** | No | Yes | Specifies that you must authenticate to the Git repository. |
| **spec.auth.type** | No | Yes  | Defines if you must authenticate to the repository with a password or token (`basic`), or an SSH key (`key`). For SSH, this parameter must be set to `key`. |
| **spec.auth.secretName** | No | Yes | Specifies the name of the Secret with credentials used by the Function Controller to authenticate to the Git repository in order to fetch the Function's source code and dependencies. This Secret must be stored in the same Namespace as the GitRepository CR. The **spec.auth.secretName** parameter is required if you provide **spec.auth**. |

## Related resources and components

These are the resources related to this CR:

| Custom resource           | Description                   |
| ------------------- | ------------------------------------------------------------------------------------------------------------ |
| [Function](#custom-resource-function)     | Stores the Function's source code and dependencies on a cluster.  |

These components use this CR:

| Component           | Description                              |
| ------------------- | ------------------------------------------------------------------------------------------------------------ |
| Function Controller | Uses the GitRepository CR to locate the Function's source code and dependencies in a Git repository.
