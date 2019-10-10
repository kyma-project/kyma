---
title: Install Kyma through the GCP Marketplace
type: Installation
---

Follow these steps to quickly install Kyma through the GCP Marketplace:

1. Access **project Kyma** on the [Google Cloud Platform (GCP) Marketplace](https://console.cloud.google.com/marketplace/details/sap-public/kyma?q=kyma%20project) and click **CONFIGURE**.

2. When the pop-up box appears, select the project in which you want to create a Kubernetes cluster and deploy Kyma.

3. To create a Kubernetes cluster for your Kyma installation, select a cluster zone from the drop-down menu and click **Create cluster**. Wait for a few minutes for the Kubernetes cluster to provision.

4. Adjust the basic settings of the Kyma deployment or use the default values:

   | Field   |      Default value     |
   |----------|-------------|
   | **Namespace** | `default` |
   | **App instance name** | `kyma-1` |
   | **Cluster Admin Service Account** | `Create a new service account` |

5. Accept the GCP Marketplace Terms of Service to continue.

6. Click **Deploy** to install Kyma.

   >**NOTE:** The installation can take several minutes to complete.

7. After you click **Deploy**, you're redirected to the **Applications** page under **Kubernetes Engine** in the GCP Console where you can check the installation status. When you see a green checkmark next to the application name, Kyma is installed. Follow the instructions from the **Next steps** section in **INFO PANEL** to add the Kyma self-signed TLS certificate to the trusted certificates of your OS.

8. Access the cluster using the link and login details provided in the **Kyma info** section on the **Application details** page.

   >**TIP:** Watch [this](https://www.youtube.com/watch?v=hxVhQqI1B5A) video for a walkthrough of the installation process.
