# Application Connector

## Overview

The Application Connector connects an external solution to Kyma.

## Details

The Application Connector Helm chart contains all the global components:
- Applicaiton broker
- Application operator
- Application registry
- Connection token handler
- Connector service

## Connector Service

The Connector Service is responsible for generating and sending back client certificates based on a Certificate Signing Request (CSR). It supports two modes: standalone and central. In the standalone mode, the Connector Service works on and for the cluster in which it is deployed. In the central mode, Connector Service supports multiple clusters. By default, the Connector Service is installed as a standalone component.

### Install Connector Service as central

To install the Connector Service configured in the central mode, you must override several values of the Kyma Installer. Run:

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: connector-service-central-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
data:
  connector-service.deployment.args.central: "true"
  connector-service.tests.central: "true"
EOF
```

### Install Application Connector without the Connector Service

You can install the Application Connector (AC) without the Connector Service. Use this approach when preparing an environment for working with a Connector Service deployed as a standalone component.
To install the AC without the Connector Service, you must override several values of the Kyma Installer. Run:
>**NOTE:** Installing AC without the Connector Service disables creating new Applications through the Console UI.

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: connector-service-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
data:
  connector_service.enabled: "false"
EOF
```
