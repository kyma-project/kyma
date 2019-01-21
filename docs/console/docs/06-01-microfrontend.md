---
title: MicroFrontend
type: Custom Resource
---

The `microfrontend.ui.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to extend the Kyma Console. It allows you to extend the Console for the specific Namespace. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd microfrontends.ui.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample CR that extends the Console.

```yaml
apiVersion: ui.kyma-project.io/v1alpha1
kind: MicroFrontend
metadata:
  name: sample-microfrontend
  namespace: production
spec:
  version: 0.0.1
  category: Sample Category
  viewBaseUrl: https://sample-microfrontend-url.com
  navigationNodes:
    - label: Sample List
      navigationPath: items
      viewUrl: /
    - label: Details
      navigationPath: items/:id
      showInNavigation: false
      viewUrl: /:id
```

This table lists all the possible parameters of a given resource together with their descriptions:


| Field   |      Mandatory?      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** | **YES** | Specifies the name of the CR. |
| **metadata.namespace** | **YES** | Specifies the target Namespace for the CR. |
| **spec.version** | **NO** | Specifies the version of the micro front-end. |
| **spec.category** | **NO** | Specifies the category name under which the micro front-end appears in the navigation. |
| **spec.viewBaseUrl** | **YES** |  Specifies the address of the micro front-end. The address has to begin with `https://`.  |
| **spec.navigationNodes** | **YES** | The list of navigation nodes specified for the micro front-end. |
| **spec.navigationNodes.label** | **YES** | Specifies the name used to display the micro front-end's node in the Console UI. |
| **spec.navigationNodes.navigationPath** | **NO** | Specifies the path used for routing within the Console. |
| **spec.navigationNodes.viewUrl** | **NO** | Specifies the URL used to display the content of a micro front-end. |
| **spec.navigationNodes.showInNavigation** | **NO** | The Boolean that specifies if the micro front-end's node is visible in the navigation or not. |
