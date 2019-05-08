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

2. Build an image that is based on the current Installer image and includes the current installation and resources charts. Run:

    ```
    docker build -t kyma-installer -f tools/kyma-installer/kyma.Dockerfile .
    ```

3. Push the image to your Docker Hub. Run:
    ```
    docker tag kyma-installer:latest {YOUR_DOCKER_LOGIN}/kyma-installer
    docker push {YOUR_DOCKER_LOGIN}/kyma-installer
    ```

4. Prepare the deployment file.

<div tabs>
  <details>
  <summary>
  GKE - xip.io
  </summary>


Run this command:

```
(cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__.*__//g" > my-kyma.yaml
```

Alternatively, run this command if you deploy Kyma with GKE version 1.12.6-gke.X and above:

```
(cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__.*__//g" > my-kyma.yaml
```

  </details>
  <details>
  <summary>
  GKE - own domain
  </summary>


Run this command:

```
(cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__DOMAIN__/$DOMAIN/g" |sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
```

Alternatively, run this command if you deploy Kyma with GKE version 1.12.6-gke.X and above:

```
(cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__DOMAIN__/$DOMAIN/g" |sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
```


  </details>
  <details>
  <summary>
  AKS - xip.io
  </summary>


Run this command:

```
(cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__.*__//g" > my-kyma.yaml
```

Alternatively, run this command if you deploy Kyma with Kubernetes version 1.14 and above:

```
(cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__.*__//g" > my-kyma.yaml
```

  </details>
  <details>
  <summary>
  AKS - own domain
  </summary>


Run this command:

```
(cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__DOMAIN__/$SUB_DOMAIN.$DNS_DOMAIN/g" | sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
```

Alternatively, run this command if you deploy Kyma with Kubernetes version 1.14 and above:

```
(cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__DOMAIN__/$SUB_DOMAIN.$DNS_DOMAIN/g" | sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
```


  </details>
</div>

5. The output of this operation is the `my_kyma.yaml` file. Modify it to fetch the proper image with the changes you made ({YOUR_DOCKER_LOGIN}/kyma-installer). Use the modified file to deploy Kyma on your GKE cluster.
