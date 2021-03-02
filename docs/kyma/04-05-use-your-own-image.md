---
title: Use your own Kyma Installer image
type: Installation
---

When you install Kyma from a release, you use the release artifacts that already contain the Kyma Installer - a Docker image containing the combined binary of the Kyma Operator and the component charts from the `/resources` folder.
If you  want to install Kyma from sources, you must build the image yourself. You also require a new image if you add components and custom Helm charts that are not included in the `/resources` folder to the installation.

Alternatively, you can also install Kyma from the latest master or any previous master commit using Kyma CLI. See different [installation source options](https://github.com/kyma-project/cli/blob/master/docs/gen-docs/kyma_install.md#flags).

In addition to the tools required to install Kyma on a cluster, you also need:

- [Docker](https://www.docker.com/)
- [Docker Hub](https://hub.docker.com/) account or any other Docker registry

1. Clone the Kyma repository to your machine using either HTTPS or SSH. Run this command to clone the repository and change your working directory to `kyma`:

    <div tabs name="use-your-own-kyma-installer-image">
      <details>
      <summary label="https">
      HTTPS
      </summary>

      ```bash
      git clone https://github.com/kyma-project/kyma.git ; cd kyma
      ```
  
      </details>
      <details>
      <summary label="ssh">
      SSH
      </summary>

      ```bash
      git clone git@github.com:kyma-project/kyma.git ; cd kyma
      ```

      </details>
    </div>

2. Build a Kyma-Installer image that is based on the current Kyma Operator binary and includes the current installation configurations and resources charts. Run:

   ```bash
   docker build -t kyma-installer -f tools/kyma-installer/kyma.Dockerfile .
   ```

3. Push the image to your Docker Hub. Run:

   ```bash
   docker tag kyma-installer:latest {YOUR_DOCKER_LOGIN}/kyma-installer
   docker push {YOUR_DOCKER_LOGIN}/kyma-installer
   ```

4. Install Kyma using your image. Run this command:

   ```bash
   kyma install -s {YOUR_DOCKER_LOGIN}/kyma-installer
   ```
