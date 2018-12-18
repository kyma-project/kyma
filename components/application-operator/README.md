# Application Operator

## Overview

Application Operator detects changes in Remote Environment custom resources and acts accordingly.


## Performed operations

Application Operator performs different operations as a result of the following events:

 - Remote Environment created - Controller installs Helm chart containing all the necessary Kubernetes resources required for the RE to work.
 - Remote Environment updated - Status of RE release update
 - Remote Environment deleted - Controller deletes Helm chart corresponding to the given RE.


## Usage

 The Application Operator has the following parameters:
 - **appName** - This is the name used in controller registration. The default value is `application-operator`.
 - **domainName** - Domain name of the cluster. The default domain name is `kyma.local`.
 - **namespace** - Namespace where the Remote Environment charts will be deployed. The default namespace is `kyma-integration`.
 - **tillerUrl** - Tiller release server URL. The default is `tiller-deploy.kube-system.svc.cluster.local:44134`.
 - **syncPeriod** - Time period between resyncing existing resources. The default value is `30` seconds.
 - **installationTimeout** - Time after the release installation will time out. The default value is `240` seconds.
 - **proxyServiceImage** - Proxy Service image version to be used in the Remote Environment chart.
 - **eventServiceImage** - Event Service image version to be used in the Remote Environment chart.
 - **eventServiceTestsImage** - Event Service Tests image version to be used in the Remote Environment chart.

## Testing on a local deployment

When you develop the Application Connector components, you can test the changes you introduced on a local Kyma deployment before you push them to a production cluster.
To test the component you modified, run the `run-with-local-tests.sh` script located in the `scripts` directory.

Running the script builds the Docker image of the component, pushes it to the Minikube registry, and updates the component deployment in the Minikube cluster. It then triggers the `run-local-tests.sh` script, which builds the image of the acceptance tests to the Minikube registry, creates a Pod with the tests, and fetches the logs from that Pod.

Alternatively, you can run only the `run-local-tests.sh` script for the given component to build the image of the component's acceptance tests to the Minikube registry, create a Pod with the tests, and fetch the logs from that Pod.
