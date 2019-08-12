# Application Operator

## Overview

Application Operator detects changes in Application custom resources and acts accordingly.


## Performed operations

Application Operator (AO) performs different operations as a result of the following events:

 - Application created - the AO installs Helm chart containing all the necessary Kubernetes resources required for the Application to work.
 - Application updated - the AO updates the Status of the Application release.
 - Application deleted - the AO deletes Helm chart corresponding to the given Application.


## Usage

 The Application Operator has the following parameters:
 - **appName** is the name used in controller registration. The default value is `application-operator`.
 - **domainName** is the domain name of the cluster. The default domain name is `kyma.local`.
 - **namespace** is the namespace where the Application charts will be deployed. The default namespace is `kyma-integration`.
 - **tillerUrl** is the tiller release server URL. The default is `tiller-deploy.kube-system.svc.cluster.local:44134`.
 - **helmTLSKeyFile** is the path to the TLS key used for Tiller communication. The default is `/etc/certs/tls.key`.
 - **helmTLSCertificateFile** is the path to the TLS certificate used for Tiller communication. The default is `/etc/certs/tls.crt`.
 - **tillerTLSSkipVerify** disables TLS verification in communication with Tiller. The default is `true`.
 - **syncPeriod** is the time period between resyncing existing resources. The default value is `30` seconds.
 - **installationTimeout** is the time after which the release installation will time out. The default value is `240` seconds.
 - **applicationGatewayImage** is the Application Gateway image version to use in the Application chart.
 - **applicationGatewayTestsImage** is the Application Gateway Tests image version to use in the Application chart.
 - **eventServiceImage** is the Event Service image version to use in the Application chart.
 - **eventServiceTestsImage** is the Event Service Tests image version to use in the Application chart.
 - **applicationConnectivityValidatorImage** is the Application Connectivity Validator image version to use in the Application chart.

## Testing on a local deployment

When you develop the Application Connector components, you can test the changes you introduced on a local Kyma deployment before you push them to a production cluster.
To test the component you modified, run the `run-with-local-tests.sh` script located in the `scripts` directory.

Running the script builds the Docker image of the component, pushes it to the Minikube registry, and updates the component deployment in the Minikube cluster. It then triggers the `run-local-tests.sh` script, which builds the image of the acceptance tests to the Minikube registry, creates a Pod with the tests, and fetches the logs from that Pod.

Alternatively, you can run only the `run-local-tests.sh` script for the given component to build the image of the component's acceptance tests to the Minikube registry, create a Pod with the tests, and fetch the logs from that Pod.
