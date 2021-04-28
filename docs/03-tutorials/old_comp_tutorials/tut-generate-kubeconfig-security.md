---
title: Get the kubeconfig file
type: Tutorials
---

The IAM Kubeconfig Service is a proprietary tool that generates a `kubeconfig` file which allows the user to access the Kyma cluster through the Command Line Interface (CLI), and to manage the connected cluster within the permission boundaries of the user.

The service is a publicly exposed service. You can access it directly under the `https://configurations-generator.{YOUR_CLUSTER_DOMAIN}` address. The service requires a valid ID token issued by Dex to return a code `200` result.

## Steps

Follow these steps to get the `kubeconfig` file and configure the CLI to connect to the cluster:

1. Access the Console UI of your Kyma cluster.
2. Click the user icon in the upper right corner of the screen.
3. Click the **Get Kubeconfig** button to download the `kubeconfig` file to a selected location on your machine.
4. Open a terminal window.
5. Export the **KUBECONFIG** environment variable to point to the downloaded `kubeconfig`. Run this command:

   ```bash
   export KUBECONFIG={KUBECONFIG_FILE_PATH}
   ```

   >**NOTE:** Drag and drop the `kubeconfig` file in the terminal to easily add the path of the file to the `export KUBECONFIG` command you run.

6. Run `kubectl cluster-info` to check if the CLI is connected to the correct cluster.

   >**NOTE:** Exporting the **KUBECONFIG** environment variable works only in the context of the given terminal window. If you close the window in which you exported the variable, or if you switch to a new terminal window, you must export the environment variable again to connect the CLI to the desired cluster.

   Alternatively, get the `kubeconfig` file by sending a `GET` request with a valid ID token issued for the user to the `/kube-config` endpoint of the `https://configurations-generator.{YOUR_CLUSTER_DOMAIN}` service. For example:

   ```
   curl GET https://configurations-generator.{YOUR_CLUSTER_DOMAIN}/kube-config -H "Authorization: Bearer {VALID_ID_TOKEN}"
   ```
