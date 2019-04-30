# Connectivity Certs Controller

## Overview
The Connectivity Certs Controller fetches the client certificate and the root CA from the central Connector Service and saves them to Secrets.

## Configuration
You can set the following parameters in the Connectivity Certs Controller chart:
- **appName** - Name of the Controller used during registration. The default value is `connectivity-certs-controller`.
- **namespace** - Namespace in which the Controller creates the Secrets that store certificates. The default Namespace is `kyma-integration`.
- **clusterCertificatesSecret** - Name of the Secret which stores the client certificate and key. The default name is `cluster-client-certificates`.
- **caCertificatesSecret** - Name of the Secret which stores the CA certificate. The default name is `nginx-auth-ca`.
- **syncPeriod** - Time period between resyncing existing resources. The default value is 5 minutes.
- **minimalSyncTime** - Minimal time between trying to synchronize with Central Connector Service. The default value is 5 minutes.
