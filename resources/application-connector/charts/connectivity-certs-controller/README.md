# Connectivity Certs Controller

## Overview
The Connectivity Certs Controller fetches the client certificate and the root CA from the central Connector Service and saves them to Secrets.

## Configuration
You can set the following parameters in the Connectivity Certs Controller chart:
- **appName** is the name of the Controller used during registration. The default value is `connectivity-certs-controller`.
- **namespace** is the namespace in which the Controller creates the Secrets that store certificates. The default Namespace is `kyma-integration`.
- **clusterCertificatesSecret** is the namespace and the name of the Secret which stores the client certificate and key. Requires the `Namespace/secret name` format. The default value is `kyma-integration/cluster-client-certificates`.
- **caCertificatesSecret** is the namespace and the name of the Secret which stores the CA certificate. Requires the `Namespace/secret name` format. The default value is `istio-system/application-connector-certs`.
- **controllerSyncPeriod** is the time period between resyncing existing resources. The default value is 5 minutes.
- **minimalConnectionSyncPeriod** is the minimal time between trying to synchronize with Central Connector Service. The default value is 5 minutes.
 
