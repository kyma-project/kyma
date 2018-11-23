---
title: TokenFetch
type: Custom Resource
---

The `tokenfetch.connectorservice.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to request token for RemoteEnvironment from the Connector Service. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd tokenrequest.connectorservice.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource in which the TokenFetch requests for the token for `test` Remote Environment:

```
apiVersion: connectorservice.kyma-project.io/v1alpha1
kind: TokenFetch
metadata:
  name: test
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR and the Remote Environment to fetch token for. |
