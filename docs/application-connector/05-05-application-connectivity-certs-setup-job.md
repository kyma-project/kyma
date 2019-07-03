---
title: Application Connectivity Certs Setup Job
type: Configuration
---

To configure the Application Connectivity Certs Setup Job, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents: 
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Sub-charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-sub-chart-overrides)


## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **global.applicationConnectorCaKey** | Specifies the private key for the Application Connector. If not specified it will be generated. | `` |
| **global.applicationConnectorCa** | Specifies the certificate for the Application Connector. If not specified it will be generated | `` |
| **application_connectivity_certs_setup_job.secrets.connectorCertificateSecret.name** | Specifies the Secret name to which the certificate and key for the Connector Service should be saved | `connector-service-app-ca` |
| **application_connectivity_certs_setup_job.secrets.connectorCertificateSecret.namespace** | Specifies the Secret namespace to which the certificate and key for the Connector Service should be saved | `kyma-integration` |
| **application_connectivity_certs_setup_job.secrets.caCertificateSecret.name** | Specifies the Secret name to which the CA certificate should be saved | `application-connector-certs` |
| **application_connectivity_certs_setup_job.secrets.caCertificateSecret.namespace** | Specifies the Secret namespace to which the CA certificate should be saved | `istio-system` |
| **application_connectivity_certs_setup_job.certificate.validityTime** | Specifies how long the generated certificate should be valid. | `92d` | 

