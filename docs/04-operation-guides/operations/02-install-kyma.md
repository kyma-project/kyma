---
title: Install Kyma
---

You can simply use the default Kyma installation, or modify it as it fits your purposes.

Meet the prerequisites, provision a k3d cluster, and use the `deploy` command to run Kyma locally.

## Prerequisites

>**CAUTION:** As of version 1.20, [Kubernetes deprecated Docker](https://kubernetes.io/blog/2020/12/02/dont-panic-kubernetes-and-docker/) as a container runtime in favor of [containerd](https://containerd.io/). Due to a different way in which containerd handles certificate authorities, Kyma's built-in Docker registry does not work correctly on clusters running with a self-signed TLS certificate on top of Kubernetes installation where containerd is used as a container runtime. If that is your case, either upgrade the cluster to use Docker instead of containerd, generate a valid TLS certificate for your Kyma instance or [configure an external Docker registry](https://kyma-project.io/docs/kyma/latest/03-tutorials/00-serverless/svls-07-set-external-registry/).

- [Kubernetes](https://kubernetes.io/docs/setup/) (supported version 1.26)
  - [k3d](https://k3d.io) (for local installation only, v5.0.0 or higher)
- [Kyma CLI](https://github.com/kyma-project/cli)
- Minimum Docker resources: 4 CPUs and 8 GB RAM (learn how to adjust the values on [Mac](https://docs.docker.com/desktop/settings/mac/#resources), [Windows](https://docs.docker.com/desktop/settings/windows/#resources), or [Linux](https://docs.docker.com/desktop/settings/linux/#resources)).

## Provision and install

> **CAUTION:** Installation on a local k3d cluster currently does not work on Apple M1 SoC.

You can either use an out-of-the-box k3d cluster or choose any other cluster provider. To quickly provision a k3d cluster, run:

  ```bash
  kyma provision k3d
  ```

  But you can do more. To define the name of your k3d cluster and pass arguments to the Kubernetes API server, for example, to log to stderr, run:

  ```bash
  kyma provision k3d --name='{CUSTOM_NAME}' --k3s-arg='--alsologtostderr@server:0'
  ```

> **NOTE:** If you're on Linux and provisioning k3d fails, follow the [troubleshooting guide](../troubleshooting/01-k3d-fails-on-linux.md).

## Default installation

Use the `deploy` command to install Kyma.

  ```bash
  kyma deploy
  ```

With Kyma installed on a local k3d cluster, access Kyma Dashboard using `kyma dashboard`. The command opens a browser and takes you to localhost with the web-based administrative UI for Kyma. Use the same command to access Kyma installed on a remote cluster.

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
  {load_balancer_IP} grafana.kyma.example.com oauth2.kyma.example.com registry.kyma.example.com connector-service.kyma.example.com gateway.kyma.example.com
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

> **NOTE:** To learn how to enable a Kyma module go to [Enable, disable and upgrade a Kyma module](./08-enable-disable-upgrade-kyma-module.md#enable-a-kyma-module).
