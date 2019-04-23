---
title: Use your own image
type: Installation
---

You can use your own image to prepare the installation file when installing Kyma on a GKE or AKS cluster.

>**NOTE:** Follow these steps both for your own and the `xip.io` default domain.

1. Check out [kyma-project](https://github.com/kyma-project/kyma) and enter the root folder.

2. Build an image that is based on the current Installer image and includes the current installation and resources charts. Run:

    ```
    docker build -t kyma-installer:latest -f tools/kyma-installer/kyma.Dockerfile .
    ```

3. Push the image to your Docker Hub. Run:
    ```
    docker tag kyma-installer:latest {YOUR_DOCKER_LOGIN}/kyma-installer:latest
    docker push {YOUR_DOCKER_LOGIN}/kyma-installer:latest
    ```

4. Prepare the deployment file.

<div tabs>
  <details>
  <summary>
  GKE
  </summary>


    - Run this command if you use the `xip.io` default domain:
    ```
    (cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

    - Run this command if you use your own domain:
    ```
    (cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__DOMAIN__/$DOMAIN/g" |sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```
    > **NOTE:** If you deploy Kyma with GKE version 1.12.6-gke.X and above, follow these steps to prepare the deployment file.

    - Run this command if you use the xip.io default domain:
    ```
    (cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

    - Run this command if you use your own domain:
    ```
    (cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__DOMAIN__/$DOMAIN/g" |sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

  </details>
  <details>
  <summary>
  AKS
  </summary>


    - Run this command if you use the `xip.io` default domain:
    ```
    (cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

    - Run this command if you use your own domain:
    ```
    (cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__DOMAIN__/$SUB_DOMAIN.$DNS_DOMAIN/g" | sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```
    > **NOTE:** If you deploy Kyma with Kubernetes version 1.14 and above, follow these steps to prepare the deployment file.

    - Run this command if you use the xip.io default domain:
    ```
    (cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

    - Run this command if you use your own domain:
    ```
    (cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__DOMAIN__/$SUB_DOMAIN.$DNS_DOMAIN/g" | sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

  </details>
</div>

5. The output of this operation is the `my_kyma.yaml` file. Modify it to fetch the proper image with the changes you made ([YOUR_DOCKER_LOGIN]/kyma-installer:latest). Use the modified file to deploy Kyma on your GKE cluster.
