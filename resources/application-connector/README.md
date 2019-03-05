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

The Connector Service is responsible for generating and sending back client certificates based on Certificate Signing Request (CSR). It supports two modes: standalone and central. In standalone mode, the Connector Service works on and for the cluster in which it is deployed. In the central mode, Connector Service supports multiple clusters. By default, the installation of the Connector Service installation is set to standalone.

### Install Connector Service as central

To install Connector Service configured in the central mode you must override values as presented in the following example:

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

### Install without Connector Service

You can install Application Connector without Connector Service. For example in preparing an environment for working with central Connector Service. Installation without Connector Service will also disable the possibility to create Application from the UI. To perform the installation without Connector Service you must override values as presented in the following example:

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