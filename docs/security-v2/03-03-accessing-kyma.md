---
title: Accessing Kyma
type: Details
---

Kyma can be accessed in two different ways:
- Console User Interface: Simple webUI, which allows the user to view, create and manage their resources. 
- Kubernetes native CLI (kubectl): Kyma uses a custom proxy to handle all api-server connections, therefore one can distinguish two types of config files:
    + Cluster config file: Obtained directly from your cloud provider, allows direct access to the k8s api-server, usually as the admin user. This config is not kyma managed.
    + Kyma generated config file: Obtained from the webUI, this config uses the kyma api-server and predetermined user configuration (access and restrictions). 

![Kyma access diagram](assets/access-kyma.svg)

## Console UI
To read more about the console, please visit [this page](components/console/#overview-overview)


## CLI access with kubectl

