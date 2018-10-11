---
title: Create a new Remote Environment
type: Getting Started
---

The Remote Environment Controller provisions and de-provisions the necessary deployments for the created Remote Environment (RE). The controller creates all Remote Environments in the `kyma-integration` Namespace.

>**NOTE:** A Remote Environment represents a single connected external solution.

To create a new RE, run this command:

```
cat <<EOF | kubectl apply -f -
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: RemoteEnvironment
metadata:
  name: {RE_NAME}
spec:
  description: {RE_DESCRIPTION}
  labels:
    region: us
    kind: production
EOF
```
