---
title: Logging in to private Npm registry using credentials from Secret
type: Tutorials
---

This tutorial shows how you can login in to private Npm registry through define credentials in Secret. 


## Steps

These sections will lead you through the whole installation, configuration, and synchronization process. You will first install k3d and create a cluster for your custom resources (CRs). Then, you will need to apply the necessary Custom Resource Definitions (CRDs) from Kyma to be able to create Functions and triggers. Finally, you will install Flux and authorize it with the `write` access to your GitHub repository in which you store the resource files. Flux will automatically synchronize any new changes pushed to your repository with your k3d cluster.

### Install and configure a k3d cluster

1. Export these variables:

 ```bash
 export REGISTRY={ADDRESS_TO_NPM_REGISTRY}
 export TOKEN={TOKEN_TO_NPM_REGISTRY}
 export NAMESPACE={FUNCTION_NAMESPACE}
 ```

2. Create a Secret:

 ```yaml
 cat <<EOF | kubectl apply -f -
 apiVersion: v1
 kind: Secret
 metadata:
   name: serverless-npm-registry-config
   namespace: {NAMESPACE}
 type: Opaque
 stringData:
   .npmrc: |
       registry=https://{REGISTRY}
       //{REGISTRY}:_authToken={TOKEN}
 EOF
 ```
 