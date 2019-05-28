---
title: Update Kyma
type: Installation
---

This guide describes how to update Kyma deployed locally or on a cluster.

If you are unsure of what version of Kyma you're running, check the name of the Installer image used for deployment. Run:
```
kubectl -n kyma-installer get deploy kyma-installer -o jsonpath='{.spec.template.spec.containers[].image}
```

## Prerequisites

- [Docker](https://www.docker.com/)
- Access to a Docker Registry - only for cluster installation

## Overview

Kyma consists of multiple components, installed as [Helm](https://github.com/helm/helm/tree/master/docs) releases.

Update of an existing deployment can include:
- Changes in charts
- Changes in overrides
- Adding new releases

The update procedure consists of three main steps:
- Prepare the update
- Update the Kyma Installer
- Trigger the update process

In case of dependency conflicts or major changes between components versions, some updates may not be possible.

> **CAUTION:** Currently Kyma doesn't support removing components as a part of the update process.

## Prepare the update

- If you update an existing component, make all required changes to the Helm charts of the component located in the [`resource`](https://github.com/kyma-project/kyma/tree/master/resources) directory.

- If you add a new component to your Kyma deployment, add a top-level Helm chart for that component. Additionally, run this command to edit the Installation custom resource and add the new component to the installed components list:
  ```
  kubectl edit installation kyma-installation
  ```

> **NOTE:** Read [this](#custom-resource-installation) document to learn more about the Installation custom resource.


- If you introduced changes in overrides, update the existing ConfigMaps and Secrets. Add new ConfigMaps and Secrets if required. See [this](#configuration-helm-overrides-for-kyma-installation) document for more information on overrides.

## Perform the update

If your changes involve any modifications in the `/resources` folder that includes component chart configurations, perform the whole update process that includes updating the Kyma Installer and triggering the update. If you only modify installation artifacts, for example by adding or removing components in the installation files or adding or removing overrides in the configuration files, only trigger the update process.

Read about each update step in the following sections.

### Update the Kyma Installer on a local deployment

- Build a new image for the Kyma Installer:  
  ```
  ./installation/scripts/build-kyma-installer.sh
  ```  
  > **NOTE:** If you started Kyma with the `run.sh` script with a `--vm-driver {value}` parameter, provide the same parameter to the `build-kyma-installer.sh` script.

- Restart the Kyma Installer Pod:  
  ```
  kubectl delete pod -n kyma-installer {INSTALLER_POD_NAME}
  ```

### Update the Kyma Installer on a cluster deployment

- Build a new image for the Kyma Installer:
  ```
  docker build -t {IMAGE_NAME}:{IMAGE_TAG} -f tools/kyma-installer/kyma.Dockerfile .
  ```

- Push the image to your Docker registry.

- Redeploy the Kyma Installer Pod using the new image. Run this command to edit the Deployment configuration:
  ```
  kubectl edit deployment kyma-installer -n kyma-installer
  ```
  Change the `image` and `imagePullPolicy` attributes in this section:  
    ```  
         spec:
           containers:
           - image: <your_image_name>:<your_tag>
             imagePullPolicy: Always
    ```  
> **NOTE:** If the desired image name and `imagePullPolicy` is already set in the deployment configuration, restart the Pod by running `kubectl delete pod -n kyma-installer {INSTALLER_POD_NAME}`.

### Trigger the update process

Execute the following command to trigger the update process:

```
kubectl label installation/kyma-installation action=install
```
