# Connectivity Certs Controller

## Overview
The Connectivity Certs Controller is responsible for fetching client certificate and root CA from Connector Service.

## Configuration
The Connectivity Certs Controller has the following parameters, that can be set through the chart:
- **appName** - Name used in controller registration. The default value is `connectivity-certs-controller`.
- **namespace** - Namespace in which secrets are created. The default is `kyma-integration`.
- **clusterCertificatesSecret** - Secret name where cluster client certificate and key are kept. The default is `cluster-client-certificates`.
- **caCertificatesSecret** - Secret name where CA certificate is kept. The default is `nginx-auth-ca`.
