---
title: Configure Helm Broker
type: Configuration
---

The Helm Broker fetches bundle definitions from an HTTP server defined in the `values.yaml` file. The **config.repository.URL** attribute defines the HTTP server URL.

### Configuring the Helm Broker externally

Follow these steps to change the configuration and make the Helm Broker fetch bundles from a custom HTTP server:

1. Create a remote bundles repository. Your remote bundle repository must include the following resources:
    - A `yaml` file which defines available bundles, for example `bundles.yaml`.
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

2. In the `values.yaml` provide your server's URLs in the **repository.URLs** attribute as a list of URLs separated by a semicolon:

  ```yaml
    repository:
      URLs: "http://custom.bundles-repository/bundles.yaml;http://another.bundles-repository/bundles.yaml"
  ```
  > **NOTE:** You can skip the `yaml` filename in the URL if the name of the file is `index.yaml`. In that case, your URL should be equal to `http://custom.bundles-repository/`.

3. Install Kyma on Minikube. See [this](/docs/root/kyma#installation-install-kyma-locally-from-the-release) document to learn how.

### Configure repository URLs in the runtime

Follow these steps to add the repository URL:

1. Add a new bundle repository URL:

 ```bash
 URLS=$(kubectl get -n kyma-system deployment/core-helm-broker --output=jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="APP_REPOSITORY_URLS")].value}')

 kubectl set env -n kyma-system deployment/core-helm-broker -e APP_REPOSITORY_URLS="$URLS;http://custom.bundles-repository/bundles.yaml"

 # At the moment, there is a bug in Minikube. As a result, for the local installation you need to manually edit the broker and bump the relistRequests attribute.

 kubectl edit clusterservicebrokers core-helm-broker
 ```

2. Wait for the Helm Broker to run using the following command:

 ```bash
 kubectl get pod -n kyma-system -l app=core-helm-broker
 ```

Running the Helm Broker triggers the Service Catalog synchronization automatically. New ClusterServiceClasses appear after a half-minute.
