---
title: Configure Helm Broker
type: Configuration
---

The Helm Broker fetches bundle definitions from an HTTP server defined in the `values.yaml` file. The **config.repository.baseURL** attribute defines the HTTP server URL. By default, the Helm Broker contains an embedded HTTP server which serves bundles from the Kyma `bundles` directory.


### Configuring the Helm Broker on the embedded HTTP server

By default, the Helm Broker contains an embedded HTTP server which serves bundles from the `bundles` directory. Deploying Kyma automatically populates the bundles.

To add a yBundle to the Helm Broker, place your yBundle directory in the `bundles` folder.
> **NOTE:** The name of the yBundle directory in the `bundles` folder must follow this pattern: \<bundle_name>\-\<bundle_version>\.


### Configuring the Helm Broker externally

Follow these steps to change the configuration and make the Helm Broker fetch bundles from a remote HTTP server:

1. Create a remote bundles repository. Your remote bundle repository must include the following resources:
    - An `index.yaml` file which defines available bundles.
      This file must have the following structure:

      ```text
      apiVersion: v1
      entries:
        <bundle_name>:
          - name: <bundle_name>
            description: <bundle_description>
            version: <bundle_version>
      ```
      This is an example of an `index.yaml` file for the Redis bundle:
      ```text
      apiVersion: v1
      entries:
        redis:
          - name: redis
            description: Redis service
            version: 0.0.1
      ```

    - A `<bundle_name>-<bundle_version>.tgz` file for each bundle version defined in the `index.yaml` file. The `.tgz` file is an archive of your bundle's directory.

2. In the [values.yaml](../../../resources/core/charts/helm-broker/values.yaml) file, set the **embeddedRepository.provision** attribute to `false` to disable the embedded server. Provide your server's URL in the **config.repository.baseURL** attribute:

  ```yaml
embeddedRepository:
  # Defines whether to provision the embedded bundle repository.
  # To provision, specify this value to true
  provision: true

  config:
    repository:
      baseURL: "http://custom.bundles-repository"
  ```

3. Install Kyma on Minikube. See the [local Kyma installation](../../kyma/docs/031-gs-local-installation.md) document to learn how.
