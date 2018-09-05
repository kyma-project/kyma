---
title: Group
type: Custom Resource
---

The `groups.authentication.kyma-project.io` Custom Resource Definition (CRD) is a detailed description of the kind of data and the format representing user group available in the ID provider. To get the up-to-date CRD and show
the output in the `yaml` format, run this command:

```
kubectl get crd groups.authentication.kyma-project.io -o yaml
```

## Sample Custom Resource

This is a sample CR that represents user group available in the identity provider.

```
apiVersion: authentication.kyma-project.io/v1alpha1
kind: Group
metadata:
    name: sample-group
spec:    
    name: admins      
    idpName: github
    description: "admins" represent group of users with administrative privilages in the organization.
```

This table analyses the elements of the sample CR and the information it contains:


| Field   |      Mandatory?      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **spec.name** | **YES** | Specifies the name of the group. |
| **spec.idpName** | **YES** | Specifies the name of the identity provider in which provided group exists. |
| **spec.description** | **NO** | Description of the group available in the ID provider. |
