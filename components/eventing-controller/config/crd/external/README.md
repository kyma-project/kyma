# External Custom Resource Definitions

## Overview

This directory contains the external custom resource definitions used for testing purposes.

## Custom Resource Definitions

| Kind | API Version | Description | Reference |
| --------| -------------------------------- |------------------------------------------------------------------------- | --------- |
| APIRule | gateway.kyma-project.io/v1alpha1 | The APIRule instance allows exposing services to outside of the cluster. | [APIRule CRD](https://github.com/kyma-incubator/api-gateway/blob/master/config/crd/bases/gateway.kyma-project.io_apirules.yaml) |
| OAuth2Client | hydra.ory.sh/v1alpha1 | The OAuth2Client instance allows creation of OAuth2 credentials. | [OAuth2Client CRD](https://github.com/ory/hydra-maester/blob/master/config/crd/bases/hydra.ory.sh_oauth2clients.yaml) |
