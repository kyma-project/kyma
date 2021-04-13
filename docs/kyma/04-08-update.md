---
title: Update Kyma
type: Installation
---

This guide describes how to update Kyma deployed locally or on a cluster.

>**NOTE:** Updating Kyma means introducing changes to a running deployment. If you want to upgrade to a newer version, read the [installation document](#installation-upgrade-kyma).

## Prerequisites

- [Kyma CLI]((https://github.com/kyma-project/cli))
- [Docker](https://www.docker.com/)
- Access to a Docker Registry - only for cluster update

## Overview

Kyma consists of multiple components, installed as [Helm](https://helm.sh/docs/) releases.

Update of an existing deployment can include:

- Changes in charts
- Changes in overrides
- Adding new Helm releases

The update procedure consists of three main steps:

- Prepare the update
- Trigger the update process

In case of dependency conflicts or major changes between components versions, some updates may not be possible.

> **CAUTION:** Currently Kyma doesn't support removing components as a part of the update process.

## Prepare the update

- If you update an existing component, make all required changes to the Helm charts of the component located in the [`resources`](https://github.com/kyma-project/kyma/tree/main/resources) directory.

- If you add a new component to your Kyma deployment, add a top-level Helm chart for that component. Additionally, download the current [Installation custom resource](#custom-resource-installation) from the cluster and add the new component to the components list:

   ```bash
   kubectl -n default get installation kyma-installation -o yaml > installation.yaml
   ```

- If you introduce changes in the overrides, create a file with your changes as ConfigMaps or Secrets. See the [configuration document](#configuration-helm-overrides-for-kyma-installation) for more information on overrides.

## Perform the update

If your changes involve any modifications in the `/resources` folder that includes component chart configurations, perform the steps under the **Update with resources modifications** tab. If you only modify installation artifacts, for example by adding or removing components in the installation files or adding or removing overrides in the configuration files, perform the steps under the **Update without resources modifications** tab.

Read about each update step in the following sections.

<div tabs name="perform-the-update">
   <details>
   <summary label="update-with-resources-modifications">
   Update with resources modifications
   </summary>

   1. Check which version you're currently running. Run this command:

      ```bash
      kyma version
      ```

   2. Provide the same version of the current cluster to the upgrade command. Provide also an image name and a tag so that Kyma CLI will build a Docker image with your local changes and push it to the registry. It will also trigger the update process. If you have changes for the overrides or the components list, you can also pass them using the `-o` and `-c` flags.

      ```bash
      kyma upgrade -s local --custom-image {IMAGE_NAME}:{IMAGE_TAG}
      ```

   </details>
   <details>
   <summary label="update-without-resources-modifications">
   Update without resources modifications
   </summary>

   1. Check which version you're currently running. Run this command:

      ```bash
      kyma version
      ```

   2. Provide the same version of the current cluster to the upgrade command. Pass the path of the overrides file using the `-o` flag and/or the path of the installation file using the `-c` flag:

      ```bash
      kyma upgrade -s {VERSION} -o {OVERRIDES_FILE_PATH} -c {INSTALLATION_FILE_PATH}
      ```

   </details>
</div>
