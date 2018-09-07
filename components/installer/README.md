# Installer

Installer is a tool for installing all Kyma components.
The project is based on the Kubernetes operator pattern. It tracks changes of the `Installation` type instance of the custom resource. It also installs, uninstalls, and updates Kyma accordingly.

## Prerequisites

- Minikube 0.26.1
- kubectl 1.9.0
- Docker
- jq

## Development

Before each commit, use the [before-commit.sh](./before-commit.sh) script to test your changes:
```
./before-commit.sh
```

### Build a Docker image

Run the [build.sh](./scripts/build.sh) script to build a Docker image:

```
./scripts/build.sh
```

### Run on Minikube

#### Run with local Kyma resources
```
export KYMA_PATH={PATH_TO_KYMA_SOURCES}
```
This environment variable is used to trigger shell scripts located in Kyma.
```
./scripts/run.sh --local --cr {PATH_TO_CR}
```

It will run Minikube, set up the installer, and install Kyma from local sources. If Minikube was started before, it will be restarted.

See the [Custom Resource file](#custom-resource-file) section in this document to learn how to generate a Custom Resource file.

To track progress of the installation, run:

```
../../installation/scripts/is-installed.sh
```

#### Rerun without restarting Minikube

This scenario is useful when you want to reuse cached Docker images.

Run the following script to clear running Minikube:
```
../../installation/scripts/clean-up.sh
```

Execute the `run.sh` script with the `--skip-minikube-start` flag to rerun Kyma installation without stopping your Minikube:
```
./scripts/run.sh --skip-minikube-start
```

### Update the Kyma cluster

Connect to the cluster that hosts your Kyma installation. Prepare the URL to the updated Kyma `tar.gz` package. Run the following command to edit the installation CR:
```
kubectl edit installation/{CR_NAME}
```
Change the `url` property in **spec** to `{URL TO KYMA TAR GZ}`. Trigger the update by overriding the **action** label in CR:
```
kubectl label installation/{CR_NAME} action=update --overwrite
```

### Update the local Kyma installation

Prepare local changes in Kyma sources. Run the following command to copy the updated sources to the Installer Pod and trigger the update action:
```
../../installation/scripts/update.sh --local --cr-name {CR_NAME}
```

> **NOTE:** You do not have to restart Minikube.

### Uninstall Kyma

Run the following command to completely uninstall Kyma:
```
kubectl label installation/kyma-installation action=uninstall
```

### Custom Resource file

The Custom Resource file for installer provides the basic information for Kyma installation.

The required properties iclude:

- `url` which is the URL address to a Kyma charts package. Only `tar.gz` is supported and it is required for non-local installations only.
- `version` which is the version of Kyma.
- `components` which is the list of Kyma components.


### Generate the Custom Resource file for installer

Generate a Custom Resource file using the [create-cr.sh](../../installation/scripts/create-cr.sh) script. It accepts the following arguments:

- `--output` is a mandatory parameter which indicates the location of the Custom Resource output file.
- `--url` is the URL to the Kyma package.
- `--version` is the version of Kyma.

Example:
```
../../installation/scripts/create-cr.sh --output installer-cr.yaml --url {URL TO KYMA TAR GZ} --version 0.0.1
```

To run the Kyma with the Azure Broker enabled, mark `azure-broker` subcomponent as enabled using `manage-component.sh` script and specify the Azure credentials in the `azure-broker-secret.yaml` Secret file which template is in `installation/resources` directory. Edit the file by providing the following properties along with their values encoded as base64 in the **data** definition: `azure-broker.subscription_id`, `azure-broker.tenant_id`, `azure-broker.client_id` and `azure-broker.client_secret`.

## Static overrides for cluster installations

You can define cluster-specific overrides for each root chart. In the cluster deployment scenario, the installer reads the `cluster.yaml` file in each root chart and appends its content to the overrides applied during the 
Helm installation.

## Install selected components only

This tool installs components specified in `Installation` CR. See the [installer-cr.yaml.tpl](../../installation/resources/installer-cr.yaml.tpl) file for more details. 

To enable installation of a component, specify its name and Namespace. If you want the release name to be different from the component's name, provide the release parameter.

Example:

```
spec:
    ...
    components:
    ...
    - name: "remote-environments"
      namespace: "kyma-integration"
      release: "hmc-default"
```

In the example, the `remote-environments` component is installed in the `kyma-integration` Namespace, using `hmc-default` as the release name. The **name** and **namespace** fields are mandatory. The name of the component is also the name of the component subdirectory in the `resources` directory. The Installer assumes that the component subdirectory is a valid Helm chart.
