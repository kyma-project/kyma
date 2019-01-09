---
title: Update Kyma Installation
type: Installation
---

This Installation guide describes how to update Kyma installation in both local and cluster scenarios.

## Prerequisites

- Working Kyma installation
- Access to the Kyma installation cluster with kubectl tool
- Docker
- Access to a Docker Registry - only for cluster installation

## Overview

Kyma consists of multiple components, installed as [Helm](https://github.com/helm/helm/tree/master/docs) releases.

Update of an existing installation can include:
- changes in charts
- changes in overrides
- adding new releases

> **NOTE:** In case of dependency conflicts or major changes between components versions, some updates may not be possible. Further steps assume that the update of all involved components is not obstructed.

> **NOTE:** Currently there's no support for components removal during update.

## Update procedure

The update procedure consists of three main steps:
- Prepare the update
- Update Kyma-Installer
- Trigger the update process

### Prepare the update

For existing components: Make all required changes to Helm charts in the `resources` subdirectory.

For new components: add a top-level Helm chart for each component. Use `kubectl edit installation kyma-installation` to include the component on the Installation CR components list, as described in the **Installation Custom Resource** document.

For changes in overrides: update existing ConfigMaps and Secrets. Add new ConfigMaps and Secrets if required. See the **Installation overrides** document for further reference.


### Update Kyma-Installer

#### Update Kyma-Installer - local scenario

- Build new image for Kyma-Installer with updated charts using the following script:  
```
./installation/scripts/build-kyma-installer.sh
```  
> **NOTE:** If the existing installation was started using `installation/cmd/run.sh` script with a `--vm-driver <value>` parameter, provide the same parameter to the `build-kyma-installer.sh` script.

- Redeploy the Kyma-Installer Pod:  
```
kubectl delete pod -n kyma-installer <Installer-Pod-name>
```


#### Update Kyma-Installer - cluster scenario

- Build new image for Kyma-Installer with updated charts using the following command:  
```
docker build -t <your_image_name>:<your_tag> -f tools/kyma-installer/kyma.Dockerfile .
```

- Push the image to your Docker registry.

- Redeploy Kyma-Installer Pod using the new image.  
Run: `kubectl edit deployment kyma-installer -n kyma-installer`  
An editor opens, allowing to manually edit Kyma-Installer Deployment configuration. Change the "image" and "imagePullPolicy" attributes in the following section:  
```  
     spec:
       containers:
       - image: <your_image_name>:<your_tag>
         imagePullPolicy: Always
```  
> **NOTE:** If the desired image name and imagePullPolicy is already in the deployment configuration, just remove the Pod with: `kubectl delete pod -n kyma-installer <Installer-Pod-name>`

### Trigger the update process

 Execute the following command to trigger the update process:

```
kubectl label installation/kyma-installation action=install
```
