---
title: Group
type: Custom Resource
---

The `groups.authentication.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format that represents user groups available in the ID provider in the Kyma cluster. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd groups.authentication.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample CR that represents an user group available in the ID provider in the Kyma cluster.

```yaml
apiVersion: authentication.kyma-project.io/v1alpha1
kind: Group
metadata:
    name: "sample-group"
spec:
    name: "admins"
    idpName: "github"
    description: "'admins' represents the group of users with administrative privileges in the organization."
```

This table analyses the elements of the sample CR and the information it contains:

| Field   |      Required      |  Description |
|----------|:-------------:|------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **spec.name** | Yes | Specifies the name of the group. |
| **spec.idpName** | Yes | Specifies the name of the ID provider in which the group exists. |
| **spec.description** | No | Description of the group available in the ID provider. |
