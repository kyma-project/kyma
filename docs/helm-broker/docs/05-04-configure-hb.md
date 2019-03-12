---
title: Configuration
---

By default, the Helm Broker fetches bundles from the newest release of the [`bundles`](https://github.com/kyma-project/bundles/releases) repository. You can also configure the Helm Broker to fetch bundle definitions from any remote HTTP server. Follow these steps:

1. Create a remote repository with your bundles definitions.
2. Install Kyma on Minikube. See [this](/root/kyma#installation-install-kyma-locally-from-the-release) document to learn how.
3. Create a ConfigMap which contains an URL to the repository. You can either:

  * Create a ConfigMap using the `kubectl create` command:
  ```bash
  kubectl create configmap my-helm-repos-urls -n kyma-system --from-literal=URLs=https://github.com/kyma-project/bundles/releases/download/latest/index-testing.yaml
  ```

    >**NOTE:** If you want to fetch bundles from many HTTP servers, use `\n` to separate the URLs.

    If you use this method, you must label your ConfigMap with `helm-broker-repo=true`. To add the label to your ConfigMap, run:

    ```bash
    kubectl label configmap my-helm-repos-urls -n kyma-system helm-broker-repo=true
    ```

 * Create a valid ConfigMap from the `yaml` file. Follow this example:
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

    Then, run:
    ```bash
    kubectl apply {configmap_name}
    ```

The Helm Broker triggers the Service Catalog synchronization automatically. New ClusterServiceClasses appear after a few seconds.
