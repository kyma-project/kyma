---
title: Custom Resources
type: Details
---

Kubeless uses custom resource definitions (CRD) to:

* define the information required for the configuration of custom resources
* create functions
* create objects

The function CRD ships by default with Kubeless.

See the content of the `kubeless-crd.yaml` file: 

````
apiVersion: apiextensions.k8s.io/v1beta1
description: Kubernetes Native Serverless Framework
kind: CustomResourceDefinition
metadata:
  name: {{ .Values.function.customResourceDefinition.metadata.name | quote }}
  labels:
{{ include "labels.standard" . | indent 4 }}
spec:
  group: {{ .Values.function.customResourceDefinition.spec.group | quote }}
  names:
    kind: {{ .Values.function.customResourceDefinition.names.kind | quote }}
    plural: {{ .Values.function.customResourceDefinition.names.plural | quote }}
    singular: {{ .Values.function.customResourceDefinition.names.singular | quote }}
  scope: Namespaced
  version: v1beta1

````

Use the `.yaml` file to create the custom resource using the following command:

```
kubectl create -f <filename>
```

Kubeless creates a new namespaced endpoint that you can use to create and manage custom objects. Learn how to use CRDs to create objects in the Kubeless documentation on the [Kubeless website](https://kubeless.io/).
