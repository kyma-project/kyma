---
title: Install Kyma
---

You can simply use the default Kyma installation, or modify it as it fits your purposes:

## Default installation

Meet the prerequisites, provision a k3d cluster, and use the `deploy` command to run Kyma locally.

### Prerequisites

- [Kyma CLI](https://github.com/kyma-project/cli)
- [Docker](https://docs.docker.com/get-docker/)
- [k3d](https://k3d.io/#installation)

### Provision and install

You can either use an out-of-the-box k3d cluster or choose any other cluster provider. To quickly provision a k3d cluster, run:

  ```bash
  kyma provision k3d
  ```

  > **TIP:** If you want to define the name of your k3d cluster and pass arguments to the Kubernetes API server (for example, to log to stderr), run:
  >
  > ```bash
  > kyma provision k3d --name='{CUSTOM_NAME}' --server-args='--alsologtostderr'
  > ```

## Default installation

Use the `deploy` command to install Kyma.

  ```bash
  kyma deploy
  ```

With Kyma installed on a local k3d cluster, access Kyma Dashboard using `kyma dashboard --local`. The command opens a browser and takes you to localhost with the web-based administrative UI for Kyma. With Kyma installed on a remote cluster, access the Dashboard at [`https://dashboard.kyma-project.io`](https://dashboard.kyma-project.io).

<!-- `kyma dashboard --local` and the URL aren't working yet, see 02-get-started/01-quick-install.md -->

## Choose resource consumption

By default, Kyma is installed with the default chart values defined in the `values.yaml` files. You can also control the allocation of resources, such as memory and CPU, that the components consume by installing Kyma with the following pre-defined profiles:

- **Evaluation** needs limited resources and is suited for trial purposes.
- **Production** is configured for high availability and scalability. It requires more resources than the evaluation profile but is a better choice for production workload.

For example, to install Kyma with the evaluation profile, run the following command:

  ```bash
  kyma deploy -p evaluation
  ```

>**NOTE:** You can check the values used for each component in respective folders of the [`resources`](https://github.com/kyma-project/kyma/tree/main/resources) directory. The `profile-evaluation.yaml` and `profile-production.yaml` files contain values used for the evaluation and production profiles respectively. If the component doesn't have files for the given profiles, the profile values are the same as default chart values defined in the `values.yaml` file.

A profile is defined globally for the whole Kyma installation. It's not possible to install a profile only for the selected components. However, you can [change the settings](03-change-kyma-config-values.md) by overriding the values set for the profile. The profile values have precedence over the default chart values, and override values have precedence over the applied profile.

## Install specific configuration values

- To install Kyma with different configuration values, use the `--values-file` and the `--value` flags. For details, see [Change Kyma settings](03-change-kyma-config-values.md).

## Install with custom domain

If you install Kyma on a remote cluster, you can use the out-of-the box `kyma.example.com` domain. All you need to do is get your load balancer IP address and add the following line to the `hosts` file:

  ```bash
  {load_balancer_IP} kiali.kyma.example.com grafana.kyma.example.com oauth2.kyma.example.com registry.kyma.example.com jaeger.kyma.example.com connector-service.kyma.example.com gateway.kyma.example.com
  ```

To install Kyma using your own domain name, you must provide the certificate and key as files. If you don't have a certificate yet, you can create a self-signed certificate and key:

  ```bash
  openssl req -x509 -newkey rsa:4096 -keyout key.pem -out crt.pem -days 365
  ```

  When prompted, provide your credentials, such as your name and your domain, as wildcard: `*.$DOMAIN`.

  Then, pass the certificate files to the `deploy` command:

  ```bash
  kyma deploy --domain {DOMAIN} --value global.ingress.domainName={DOMAIN} --tls-crt crt.pem --tls-key key.pem
  ```

## Install from a specific source

Optionally, you can specify from which source you want to deploy Kyma. For example, you can choose the `main` branch (or any other branch on the Kyma repository), a specific PR, or a release version. For more details, see the documentation for the `deploy` command.

For example, to install Kyma from a specific version, such as `1.19.1`, run:

  ```bash
  kyma deploy --source=1.19.1
  ```

- Alternatively, to build Kyma from your local sources and deploy it on a remote cluster, run:

  ```bash
  kyma deploy --source=local
  ```

  >**NOTE:** By default, Kyma expects to find local sources in the `$GOPATH/src/github.com/kyma-project/kyma` folder. To adjust the path, set the `-w ${PATH_TO_KYMA_SOURCES}` parameter.

## Install specific components

To deploy Kyma with only specific components, run:

  ```bash
  kyma deploy --components-file {COMPONENTS_FILE_PATH}
  ```

  `{COMPONENTS_FILE_PATH}` is the path to a YAML file containing the desired component list to be installed. In the following example, only eight components are deployed on the cluster:

  ```yaml
prerequisites:
  - name: "cluster-essentials"
  - name: "istio"
    namespace: "istio-system"
  - name: "certificates"
    namespace: "istio-system"
components:
  - name: "logging"
  - name: "tracing"
  - name: "kiali"
  - name: "monitoring"
  - name: "eventing"
  ```

- Alternatively, you can specify single components instead of a file:

  ```bash
  kyma deploy --component {COMPONENT_NAME@NAMESPACE}
  ```

  If you provide no Namespace, the default Namespace called `kyma-system` is used. For example, to install the `eventing` component in the default Namespace and the `istio` component in the `istio-system` Namespace, run:
  
  ```bash
  kyma deploy --component eventing --component istio@istio-system
  ```

>**TIP:** To see a complete list of all Kyma components go to the [`components.yaml`](https://github.com/kyma-project/kyma/blob/main/installation/resources/components.yaml) file.
