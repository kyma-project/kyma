# Core

## Overview

According to the [Manifesto](https://kyma-project.github.io/community/), Kyma is a product with batteries included. The `core` directory contains all components required to run Kyma. For more details about each core component, see the corresponding README.md files.

## Details

This section describes how to add a new core component. It also describes how to configure or disable a component that already exists.

### Add a new core component

If you develop a new core component, add a new sub-chart to the `core` directory. 
Update the [`requirements.yaml`](requirements.yaml) file by adding the **name** and **condition** attributes for the created component. 
To learn more about the **condition** attribute, see the [tags and condition fields in helm charts](https://github.com/kubernetes/helm/blob/release-2.7/docs/charts.md#tags-and-condition-fields-in-requirementsyaml) documentation.

### Inject sensitive data into a core component

To inject sensitive data into a core component during the Kyma installation, follow these steps:
1. Create the `secrets.yaml` file locally. In the file, include the name of the component to inject sensitive data to:

  ```
  helm-broker:
    config:
      basic_auth_password: p4ssw0rd
  ```

  Use the same `secrets.yaml` file for all core components. The structure of the **config** section is different for each component. For more details, see the `values.yaml` files associated with specific components.

2. Start a container during the [installation](../../docs/kyma/04-02-local-installation.md), and mount the `secrets.yaml` file in the `run.sh` script with the following command:

  ```
  ./run.sh -s ${PATH_TO_DIRECTORY_WITH_THE_SECRET_YAML_FILE}/secrets.yaml
  ```
