# Resources                                                                                  

## Overview

Resources are all components in Kyma that are available for local and cluster installation. You can find more details about each component in the corresponding README.md files.

Resources currently include, but are not limited to, the following:

- Elements which are essential for the installation of `core` components in Kyma, such as certificates, users, and permissions
- Examples of the use of specific components
- Scripts for the installation of Helm, Istio deployment, as well as scripts for validating Pods, starting Kyma, and testing


## Installation
Monitoring and Logging charts are not installed by default when you install Kyma on Minikube. They are available by default for cluster installations.

To install Monitoring chart on the Minikube cluster, run the following command inside resources directory:

```bash
helm install monitoring --set global.isLocalEnv=true --set global.alertTools.credentials.victorOps.apikey="" --set global.alertTools.credentials.victorOps.routingkey="" --set global.alertTools.credentials.slack.channel="" --set global.alertTools.credentials.slack.apiurl="" -n kyma-system
```

To install Logging chart on the Minikube cluster, run the following command inside resources directory:

```bash
helm install logging --set global.isLocalEnv=true
```
