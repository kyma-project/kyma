---
title: Configure the Helm Broker
type: Configuration
---

Follow these steps to change the configuration and make the Helm Broker fetch bundles from a custom HTTP servers:

1. Create a remote repository with your bundles definitions.
2. Install Kyma on Minikube. See [this](/root/kyma#installation-install-kyma-locally-from-the-release) document to learn how.

3. Create a ConfigMap which contains an URL to the repository:

1) trzeba oblabelowac
 ```bash
kubectl create configmap my-helm-repos-urls -n kyma-system --from-literal=URLs=https://github.com/kyma-project/bundles/releases/download/latest/index-testing.yaml
 ```
>**NOTE:** If you want to fetch bundles from many HTTP servers, use `\n` to separate the URLs.

**NOTE:** helm broker fetches repositories from congifmaps labelled with `helm-broker-repo=true`. To Add the label to your configmap, run:
 ```bash
kubectl label configmap my-helm-repos-urls -n kyma-system helm-broker-repo=true
 ```

2) kubectl apply -f {file name} Use the following example to create a valid ConfigMap with many URLs from the `yaml` file:
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


Helm Broker triggers the Service Catalog synchronization automatically. New ClusterServiceClasses appear after a few seconds.
