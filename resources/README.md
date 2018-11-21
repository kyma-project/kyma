# Resources                                                                                  

## Overview

Resources are all components in Kyma that are available for local and cluster installation. You can find more details about each component in the corresponding README.md files.

Resources currently include, but are not limited to, the following:

- Elements which are essential for the installation of `core` components in Kyma, such as certificates, users, and permissions
- Examples of the use of specific components
- Scripts for the installation of Helm, Istio deployment, as well as scripts for validating Pods, starting Kyma, and testing


## Installation
`Monitoring` chart is not installed by default on minikube installation. Its present by default for cluster installations.

To Install `monitoring` chart on minikube cluster one can run following command:

```bash
helm install monitoring --set global.isLocalEnv=true --set global.alertTools.credentials.victorOps.apikey="" --set global.alertTools.credentials.victorOps.routingkey="" --set global.alertTools.credentials.slack.channel="" --set global.alertTools.credentials.slack.apiurl="" -n kyma-system
```

