# Connectivity Certs Controller

## Overview
The Connectivity Certs Controller is responsible for fetching client certificate and root CA from Connector Service.

## Configuration
The Connectivity Certs Controller has the following parameters, that can be set through the chart:
- **appName** - Name used in controller registration. The default value is `connectivity-certs-controller`.
- **namespace** - Namespace in which secrets are created. The default is `kyma-integration`.
- **clusterCertificatesSecret** - Name of the Secret which stores where the client certificate and key are. The default name is `cluster-client-certificates`.
- **caCertificatesSecret** - Name of the Secret which stores the CA certificate. The default name is `nginx-auth-ca`.
