---
title: Use your own Kyma Installer image
type: Installation
---

When you install Kyma from a release, you use the release artifacts that already contain the Kyma Installer - a Docker image containing the combined binary of the Installer and the component charts from the `/resources` folder.
If you install Kyma from sources and use the latest `master` branch, you must build the image yourself to prepare the configuration file for Kyma installation on a GKE or AKS cluster. You also require a new image if you add components and custom Helm charts that are not included in the `/resources` folder to the installation.

In addition to the tools required to install Kyma on a cluster, you also need:
- [Docker](https://www.docker.com/)
- [Docker Hub](https://hub.docker.com/) account or any other Docker registry

>**CAUTION:** These instructions are valid starting with Kyma 1.2. If you want to install older releases, refer to the respective documentation versions. 

1. Clone the Kyma repository to your machine using either HTTPS or SSH. Run this command to clone the repository and change your working directory to `kyma`:
    <div tabs>
      <details>
      <summary>
      HTTPS
      </summary>

      ```
      git clone https://github.com/kyma-project/kyma.git ; cd kyma
      ```
      </details>
      <details>
      <summary>
      SSH
      </summary>

      ```
      git clone git@github.com:kyma-project/kyma.git ; cd kyma
      ```
      </details>
    </div>

2. Build a Kyma-Installer image that is based on the current Installer binary and includes the current installation configurations and resources charts. Run:
    ```
    docker build -t kyma-installer -f tools/kyma-installer/kyma.Dockerfile .
    ```

3. Push the image to your Docker Hub. Run:
    ```
    docker tag kyma-installer:latest {YOUR_DOCKER_LOGIN}/kyma-installer
    docker push {YOUR_DOCKER_LOGIN}/kyma-installer
    ```

4. Prepare the Kyma deployment file. Run this command:
    ```
    (cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) > my-kyma.yaml
    ```

5. The output of this operation is the `my_kyma.yaml` file.
Find the following section in `my_kyma.yaml` and modify it to fetch the image you prepared. Change `image` attribute value to `{YOUR_DOCKER_LOGIN}/kyma-installer`:
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

6. Use the `my-kyma.yaml` file to deploy Kyma. Choose the desired [installation option](#installation-overview) and run this command after you prepared the cluster:  
    ```
    kubectl apply -f my-kyma.yaml
    ```
