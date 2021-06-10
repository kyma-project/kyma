---
title: Install Kyma
---

You can simply use the default Kyma installation, or modify it as it fits your purposes:

## Default installation

If you use the `deploy` command without any flags, Kyma provides a default domain. 
For example, if you install Kyma on a local cluster, the default URL is `https://console.local.kyma.dev`.

  ```
  kyma deploy
  ```

## Choose resource consumption

By default, Kyma is installed with the default chart values defined in the `values.yaml` files. You can also control the allocation of resources, such as memory and CPU, that the components consume by installing Kyma with the following pre-defined profiles:

- **Evaluation** needs limited resources and is suited for trial purposes.
- **Production** is configured for high availability and scalability. It requires more resources than the evaluation profile but is a better choice for production workload.

For example, to install Kyma with the evaluation profile, run the following command:

  ```
  kyma deploy -p evaluation
  ```

>**NOTE:** You can check the values used for each component in respective folders of the [`resources`](https://github.com/kyma-project/kyma/tree/main/resources) directory. The `profile-evaluation.yaml` and `profile-production.yaml` files contain values used for the evaluation and production profiles respectively. If the component doesn't have files for the given profiles, the profile values are the same as default chart values defined in the `values.yaml` file.

A profile is defined globally for the whole Kyma installation. It's not possible to install a profile only for the selected components. However, you can [change the settings](#03-change-kyma-config-values) by overriding the values set for the profile. The profile values have precedence over the default chart values, and override values have precedence over the applied profile.

## Install specific configuration values

- To install Kyma with different configuration values, use the `--values-file` and the `--value` flags. For details, see [Change Kyma settings](#03-change-kyma-config-values).

## Install with custom domain

To install Kyma using your own domain name, you must provide the certificate and key as files. 
If you don't have a certificate yet, you can create a self-signed certificate and key:

  ```
  openssl req -x509 -newkey rsa:4096 -keyout key.pem -out crt.pem -days 365
  ```

  When prompted, provide your credentials, such as your name and your domain, as wildcard: `*.$DOMAIN`.

  Then, pass the certificate files to the `deploy` command:

  ```
  kyma deploy --domain {DOMAIN} --tls-cert crt.pem --tls-key key.pem
  ```

## Install from a specific source

Optionally, you can specify from which source you want to deploy Kyma. For example, you can choose the `main` branch (or any other branch on the Kyma repository), a specific PR, or a release version. For more details, see the documentation for the `deploy` command.<br>
For example, to install Kyma from a specific version, such as `1.19.1`, run:

  ```
  kyma deploy --source=1.19.1
  ```

- Alternatively, to build Kyma from your local sources and deploy it on a remote cluster, run:

  ```
  kyma deploy --source=local
  ```
  > **NOTE:** By default, Kyma expects to find local sources in the `$GOPATH/src/github.com/kyma-project/kyma` folder. To adjust the path, set the `-w ${PATH_TO_KYMA_SOURCES}` parameter.

## Install specific components

To deploy Kyma with only specific components, run:

  ```
  kyma deploy --components-file {COMPONENTS_FILE_PATH}
  ```

  `{COMPONENTS_FILE_PATH}` is the path to a YAML file containing the desired component list to be installed. In the following example, only six components are deployed on the cluster:

  ```
  prerequisites:
    - name: "cluster-essentials"
    - name: "istio"
      namespace: "istio-system"
  components:
    - name: "testing"
    - name: "xip-patch"
    - name: "istio-kyma-patch"
    - name: "dex"
  ```

- Alternatively, you can specify single components instead of a file:
  
  ```
  kyma deploy --component {COMPONENT_NAME@NAMESPACE}
  ```

  If no Namespace is provided, the `default` Namespace is used. For example, to install the `testing` component in the `default` Namespace and the `application-connector` component in the `kyma-integration` Namespace, run:
  
  ```
  kyma deploy --component testing --component application-connector@kyma-integration
  ```
