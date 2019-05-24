---
title: Use your own Kyma Installer image
type: Installation
---

When you install Kyma from the release, you use the release artifacts that already contain the Kyma Installer - a Docker image containing the combined binary of the Installer and the component charts from the `/resources` folder.
If you install Kyma from sources and use the latest `master` branch, you must build the image yourself to prepare the configuration file for Kyma installation on a GKE or AKS cluster. You also require a new image if you add to the installation components and custom Helm charts that are not included in the `/resources` folder.

In addition to the tools required to install Kyma on a custer, you also need:
- [Docker](https://www.docker.com/)
- [Docker Hub](https://hub.docker.com/) account or any other Docker registry

>**NOTE:** Follow these steps both for your own and the `xip.io` default domain.

1. Clone the repository using the `git clone https://github.com/kyma-project/kyma.git` command and navigate to the root folder.

2. Build a Kyma-Installer image that is based on the current Installer binary and includes the current installation configurations and resources charts. Run:

    ```
    docker build -t kyma-installer -f tools/kyma-installer/kyma.Dockerfile .
    ```

3. Push the image to your Docker Hub. Run:
    ```
    docker tag kyma-installer:latest {YOUR_DOCKER_LOGIN}/kyma-installer
    docker push {YOUR_DOCKER_LOGIN}/kyma-installer
    ```

4. Prepare the Kyma deployment file.
Run this command:

```
(cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) > my-kyma.yaml
```

5. The output of this operation is the `my_kyma.yaml` file.
Find the following section in `my_kyma.yaml` and modify the file by changing the `image` attribute to the new value: "{YOUR_DOCKER_LOGIN}/kyma-installer":
```
spec:
  template:
    metadata:
      labels:
        name: kyma-installer
    spec:
      serviceAccountName: kyma-installer
      containers:
      - name: kyma-installer-container
        image: {YOUR_DOCKER_LOGIN}/kyma-installer
        imagePullPolicy: IfNotPresent
```

6. Install Kyma using the custom Kyma-Installer image.
Select the installation scenario that fits your requirements - GKE or AKS, with the XIP.io or your own DNS domain.
Remember: During the installation, in the step `Deploy Kyma.`, apply your modified Kyma deployment file `my-kyma.yaml` instead of the released one! Execute:
```
kubectl apply -f my-kyma.yaml
```
