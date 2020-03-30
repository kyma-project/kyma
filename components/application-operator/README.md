# Application Operator

## Overview

The Application Operator (AO) can work in two modes.
By default, it detects changes in [Application](../../docs/application-connector/06-01-application.md) custom resources and acts accordingly. In this mode, Application Gateway is created for each Application.
In the alternative mode, it detects changes in [ServiceInstance](../../docs/service-catalog/03-01-resources.md) custom resources and acts accordingly. In this mode, Application Gateway is created per Namespace.


## Performed operations

The Application Operator performs different operations as a result of the following events.

<!--- when gatewayOncePerNamespace=false (default)  -->
In the default Gateway-per-Application mode:
 - Application created - the AO installs the Helm chart that contains all the necessary Kubernetes resources required for the Application to work.
 - Application updated - the AO updates the Status of the Application Helm Release.
 - Application deleted - the AO deletes Helm chart corresponding to the given Application.

<!--- when gatewayOncePerNamespace=true -->
In the Gateway-per-Namespace mode:
 - First ServiceInstance created in a given Namespace - the AO installs the Helm chart that contains all the necessary Kubernetes resources required for the Application Gateway to work.
 - Last ServiceInstance from a given Namespace is deleted - the AO deletes the Gateway Helm chart.


## Usage

The Application Operator has the following parameters:
 - **appName** is the name used in controller registration. The default value is `application-operator`.
 - **domainName** is the domain name of the cluster. The default domain name is `kyma.local`.
 - **namespace** is the Namespace where the AO deploys the charts of the Application. The default Namespace is `kyma-integration`.
 - **tillerUrl** is the Tiller release server URL. The default value is `tiller-deploy.kube-system.svc.cluster.local:44134`.
 - **helmTLSKeyFile** is the path to the TLS key used for secure communication with Tiller. The default value is `/etc/certs/tls.key`.
 - **helmTLSCertificateFile** is the path to the TLS certificate used for secure communication with Tiller. The default value is `/etc/certs/tls.crt`.
 - **tillerTLSSkipVerify** disables TLS verification in communication with Tiller. The default value is `true`.
 - **syncPeriod** is the time period between resyncing existing resources. The default value is `30` seconds.
 - **installationTimeout** is the time after which the release installation will time out. The default value is `240` seconds.
 - **applicationGatewayImage** is the Application Gateway image version to use in the Application chart.
 - **applicationGatewayTestsImage** is the Application Gateway Tests image version to use in the Application chart.
 - **eventServiceImage** is the Event Service image version to use in the Application chart.
 - **eventServiceTestsImage** is the Event Service Tests image version to use in the Application chart.
 - **applicationConnectivityValidatorImage** is the Application Connectivity Validator image version to use in the Application chart.
 - **gatewayOncePerNamespace** is a flag that specifies whether Application Gateway should be deployed once per Namespace based on ServiceInstance or for every Application. The default value is `false`.
 - **strictMode** is a toggle used to enable or disable Istio authorization policy for validator and HTTP source adapter. The default value is `disabled`.
## Testing on a local deployment

When you develop the Application Connector components, you can test the changes you introduced on a local Kyma deployment before you push them to a production cluster.
To test the component you modified, run the `run-with-local-tests.sh` script located in the `scripts` directory.

Running the script builds the Docker image of the component, pushes it to the Minikube registry, and updates the component deployment in the Minikube cluster. It then triggers the `run-local-tests.sh` script, which builds the image of the acceptance tests to the Minikube registry, creates a Pod with the tests, and fetches the logs from that Pod.

Alternatively, you can run only the `run-local-tests.sh` script for the given component to build the image of the component's acceptance tests in the Minikube registry, create a Pod with the tests, and fetch the logs from that Pod.