---
title: Configure Helm Broker
type: Configuration
---

The Helm Broker fetches bundle definitions from HTTP servers defined in the `helm-repos-urls` ConfigMap.

### Add a new bundle repository

Follow these steps to change the configuration and make the Helm Broker fetch bundles from a custom HTTP servers:

1. Create a remote repository with the definition of your bundles. Your remote bundle repository must include the following resources:
    - A `yaml` file which defines available bundles, for example `index.yaml`.
      This file must have the following structure:
      ```text
      apiVersion: v1
      entries:
        {bundle_name}:
          - name: {bundle_name}
            description: {bundle_description}
            version: {bundle_version}
      ```
      This is an example of a `yaml` file for the Redis bundle:
      ```text
      apiVersion: v1
      entries:
        redis:
          - name: redis
            description: Redis service
            version: 0.0.1
      ```
    - A `{bundle_name}-{bundle_version}.tgz` file for each bundle version defined in the `yaml` file. The `.tgz` file is an archive of your bundle's directory.

2. Install Kyma on Minikube. See [this](/root/kyma#installation-install-kyma-locally-from-the-release) document to learn how.

3. Create a ConfigMap which contains an URL to the repository:
 ```bash
kubectl create configmap my-helm-repos-urls -n kyma-system --from-literal=URLs=https://github.com/kyma-project/bundles/releases/download/latest/index-testing.yaml
 ```
>**NOTE:** If you want to fetch bundles from many HTTP servers, use `\n` to separate the URLs.

4. Label the newly created ConfigMap:
 ```bash
kubectl label configmap my-helm-repos-urls -n kyma-system helm-broker-repo=true
 ```
 
Helm Broker triggers the Service Catalog synchronization automatically. New ClusterServiceClasses appear after a few seconds.

Use the following example to create a valid ConfigMap with many URLs from the `yaml` file:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: helm-repos-configs
  labels:
    helm-broker-repo: "true"
data:
  URLs: |-
    https://github.com/kyma-project/bundles/releases/download/0.3.0/index-testing.yaml
    https://github.com/kyma-project/bundles/releases/download/0.3.0/
```
