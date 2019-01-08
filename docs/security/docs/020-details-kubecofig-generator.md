---
title: Kubeconfig generator
type: Details
---

The Kubeconfig generator is a proprietary tool that generates a `kubeconfig` file which allows the user to access the Kyma cluster through the Command Line Interface (CLI), and to manage the connected cluster within the permission boundaries of the user.

The Kubeconfig generator rewrites the ID token issued for the user by Dex into the generated `kubeconfig` file. The time to live (TTL) of the ID token is 24 hours, which effectively means that the TTL of the generated `kubeconfig` file is 24 hours as well.

The generator is a publicly exposed service. You can access it directly under the `https://configurations-generator.{YOUR_CLUSTER_DOMAIN}` address. The service requires a valid ID token issued by Dex to return a code `200` result.

## Get the kubeconfig file and configure the CLI

Follow these steps to get the `kubeconfig` file and configure the CLI to connect to the cluster:

1. Access the Console UI of your Kyma cluster.
2. Click **Administration**.
3. Click the **Download config** button to download the `kubeconfig` file to a selected location on your machine.
4. Open a terminal window.
5. Export the `KUBECONFIG` environment variable to point to the downloaded `kubeconfig`. Run this command:
  ```
  export KUBECONFIG={path_to_downloaded_kubeconfig}
  ```
  >**NOTE:** Drag and drop the `kubeconfig` file in the terminal to easily add the path of the file to the `export KUBECONFIG` command you run.

6. Run `kubectl cluster-info` to check if the CLI is connected to the correct cluster.

>**NOTE:** Exporting the `KUBECONFIG` environment variable works only in the context of the given terminal window. If you close the window in which you exported the variable, or if you switch to a new terminal window, you must export the environment variable again to connect the CLI to the desired cluster.

Alternatively, get the `kubeconfig` file by sending a `GET` request with a valid ID token issued for the user to the `/kube-config` endpoint of the `https://configurations-generator.{YOUR_CLUSTER_DOMAIN}` service. For example:
```
curl GET https://configurations-generator.{YOUR_CLUSTER_DOMAIN}/kube-config -H "Authorization: Bearer {VALID_ID_TOKEN}"
```
