---
title: Log into a private package registry using credentials from a Secret
type: Tutorials
---

This tutorial shows how you can login in to private package registry through define credentials in Secret. 

## Steps

### Create a Secret 

<div tabs name="override" group="external-packages-registry">
  <details>
  <summary label="node">
  Node.js
  </summary>

1. Export these variables:

 ```bash
 export REGISTRY={ADDRESS_TO_REGISTRY}
 export TOKEN={TOKEN_TO_REGISTRY}
 export NAMESPACE={FUNCTION_NAMESPACE}
 ```

2. Create a Secret:

 ```yaml
 cat <<EOF | kubectl apply -f -
 apiVersion: v1
 kind: Secret
 metadata:
   name: serverless-package-registry-config
   namespace: {NAMESPACE}
 type: Opaque
 stringData:
   .npmrc: |
       registry=https://{REGISTRY}
       //{REGISTRY}:_authToken={TOKEN}
EOF
 ```

  </details>
  <details>
  <summary label="python">
  Python
  </summary>

1. Export these variables:

 ```bash
 export REGISTRY={ADDRESS_TO_REGISTRY}
 export NAMESPACE={FUNCTION_NAMESPACE}
 export USERNAME={USERNAME_TO_REGISTRY}
 export PASSWORD={PASSWORD_TO_REGISTRY}
 ```

2. Create a Secret:

 ```yaml
 cat <<EOF | kubectl apply -f -
 apiVersion: v1
 kind: Secret
 metadata:
   name: serverless-package-registry-config
   namespace: {NAMESPACE}
 type: Opaque
 stringData:
   pip.conf: |
     [global]
     extra-index-url = {USERNAME}:{PASSWORD}@{REGISTRY}
EOF
 ```

  </details>
  <details>
  <summary label="node-python">
  Node.js & Python
  </summary>

1. Export these variables:

 ```bash
 export REGISTRY={ADDRESS_TO_REGISTRY}
 export TOKEN={TOKEN_TO_REGISTRY}
 export NAMESPACE={FUNCTION_NAMESPACE}
 export USERNAME={USERNAME_TO_REGISTRY} 
 export PASSWORD={PASSWORD_TO_REGISTRY}
 ```

2. Create a Secret:

 ```yaml
 cat <<EOF | kubectl apply -f -
 apiVersion: v1
 kind: Secret
 metadata:
   name: serverless-package-registry-config
   namespace: {NAMESPACE}
 type: Opaque
 stringData:
   .npmrc: |
       registry=https://{REGISTRY}
       //{REGISTRY}:_authToken={TOKEN}
   pip.conf: |
       [global]
       extra-index-url = {USERNAME}:{PASSWORD}@{REGISTRY}
EOF
 ```

  </details>
</div>


### Test the package registry switch

[Create a Function](#tutorials-create-an-inline-function) with dependencies from external registry and check if your Function was created and all conditions are set to `True`:

```bash
kubectl get functions -n $NAMESPACE
```

You should get a result similar to the following example:

```bash
NAME            CONFIGURED   BUILT     RUNNING   RUNTIME    VERSION   AGE
test-function   True         True      True      nodejs12   1         96s
```

> **CAUTION:** If you want to create a cluster-wide Secret, you must create it in the `kyma-system` Namespace and add the `serverless.kyma-project.io/config: credentials` label. Read more about [requirements for Secret CRs](#details-switching-registries-at-runtime).
