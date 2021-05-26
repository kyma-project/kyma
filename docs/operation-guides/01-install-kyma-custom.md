---
title: Custom Kyma Installation
---
<!-- to be reviewed by ...?  -->

Besides the default installation, there are several ways to install Kyma:

## Installation
<!-- the default variant is surely mentioned in Basic Tasks/Get Started, too -->
You can simply use the `deploy` command without any flags, and Kyma provides a default domain. 
For example, if you install Kyma on a local cluster, the default URL is `https://console.local.kyma.dev`.

  ```
  kyma deploy
  ```

If you install Kyma locally, use the evaluation profile:

  ```
  kyma deploy -p evaluation
  ```

## Installation with custom domain

To install Kyma using your own domain name, you must provide the certificate and key as files. 
If you don't have a certificate yet, you can create a self-signed certificate and key:

  ```
  openssl req -x509 -newkey rsa:4096 -keyout key.pem -out crt.pem -days 365
  ```

  When prompted, provide your credentials, such as your name and your domain (as wildcard: `*.$DOMAIN`).

  Then, pass the certificate files to the deploy command:

  ```
  kyma deploy --domain {DOMAIN} --tls-cert crt.pem --tls-key key.pem
  ```

## Installation from a specific source

Optionally, you can specify from which source you want to deploy Kyma, such as the `main` branch (or any other branch on the Kyma repository), a specific PR, or a release version. For more details, see the documentation for the `deploy` command.<br>
For example, to install Kyma from a specific version, such as `1.19.1`, run:

  ```
  kyma deploy --source=1.19.1
  ```

- Alternatively, to build Kyma from your local sources and deploy it on a remote cluster, run:

  ```
  kyma deploy --source=local
  ```
  > **NOTE:** By default, Kyma expects to find local sources in the `$GOPATH/src/github.com/kyma-project/kyma` folder. To adjust the path, set the `-w ${PATH_TO_KYMA_SOURCES}` parameter.

## Installation of specific components

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

  If no Namespace is provided, then the default Namespace is used. For example, to install the `testing` component in the default Namespace and the `application-connector` component in the `kyma-integration` Namespace, run:
  
  ```
  kyma deploy --component testing --component application-connector@kyma-integration
  ```

## Installation with specific configuration values

- You can also install Kyma with different configuration values than the default settings. To do this, you use the `--values-file` and the `--value` flag. For details, see [Change Kyma settings](#change-kyma-settings).