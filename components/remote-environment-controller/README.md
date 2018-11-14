# Remote Environment Controller

## Overview

Remote Environment Controller detects changes in Remote Environment custom resources and acts accordingly.


## Performed operations

Remote Environment Controller performs different operations as a result of the following events:

 - Remote Environment created - Controller installs Helm chart containing all the necessary Kubernetes resources required for the RE to work.
 - Remote Environment updated - Status of RE release update
 - Remote Environment deleted - Controller deletes Helm chart corresponding to the given RE.

 
 ## Usage
 
 The Remote Environment Controller has the following parameters:
 - **appName** - This is the name used in controller registration. The default value is `remote-environment-controller`.
 - **domainName** - Domain name of the cluster. Default domain name is `kyma.local`.
 - **namespace** - Namespace where the Remote Environment charts will be deployed. The default namespace is `kyma-integration`.
 - **tillerUrl** - Tiller release server url. The default is `tiller-deploy.kube-system.svc.cluster.local:44134`.
 - **syncPeriod** - Time period between resyncing existing resources. The default value is `30` seconds.
 - **installationTimeout** - Time after the release installation will time out. The default value is `240` seconds.
 - **proxyServiceImage** - The image of the Proxy Service that will be used in Remote Environment Chart.
 - **eventServiceImage** - The image of the Event Service that will be used in Remote Environment Chart.
 - **eventServiceTestsImage** - The image of the Event Service Tests that will be used in Remote Environment Chart.
