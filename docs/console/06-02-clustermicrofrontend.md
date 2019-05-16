---
title: ClusterMicroFrontend
type: Custom Resource
---

The `clustermicrofrontends.ui.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to extend the Kyma Console. It allows you to extend the Console for the entire Cluster. The cluster micro frontend  is added to Console automatically based on the `yaml` file. To avoid naming conflicts with the core system, the root node receives the `cmf-` prefix in the URL. Additionally,  **navigationContext** and **viewGroup**  [node configuration](https://github.com/kyma-project/luigi/blob/master/docs/navigation-parameters-reference.md#node-parameters) parameters are set to allow simple navigation. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd clustermicrofrontends.ui.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample CR that extends the Console.

```yaml
apiVersion: ui.kyma-project.io/v1alpha1
kind: ClusterMicroFrontend
metadata:
  name: sample-microfrontend
spec:
  version: 0.0.1
  category: category-name
  viewBaseUrl: https://sample-microfrontend-url.com
  placement: cluster
  navigationNodes:
    - label: Sample List
      navigationPath: items
      viewUrl: /
      requiredPermissions:
      - apiGroup: foo.bar.io
        resource: items
        verbs:
          - list
    - label: Details
      navigationPath: items/:id
      showInNavigation: false
      viewUrl: /:id
      requiredPermissions:
      - apiGroup: foo.bar.io
        resource: items
        verbs:
          - update
          - delete
```

This table lists all the possible parameters of a given resource together with their descriptions:


| Field   |      Mandatory      |  Description |
|----------|:-------------:|------|
| **metadata.name** | **YES** | Specifies the name of the CR. |
| **spec.version** | **NO** | Specifies the version of the cluster micro front-end. |
| **spec.category** | **NO** | Defines the category name under which the cluster micro front-end appears in the navigation. |
| **spec.viewBaseUrl** | **YES** | Specifies the address of the cluster micro front-end. The address has to begin with `https://`.  |
| **spec.placement** | **NO** |  Specifies if the cluster micro front-end should be visible in the Namespace navigation or settings navigation. The placement value has to be either `namespace` or `cluster`. |
| **spec.navigationNodes** | **YES** | The list of navigation nodes specified for the cluster micro front-end. |
| **spec.navigationNodes.label** | **YES** | Specifies the name used to display the cluster micro front-end's node in the Console UI. |
| **spec.navigationNodes.navigationPath** | **NO** | Specifies the path that is used for routing within the Console. |
| **spec.navigationNodes.viewUrl** | **NO** | Specifies the URL used to display the content of the cluster micro-front end. |
| **spec.navigationNodes.externalLink** | **NO** | Specifies the URL used to display the content of the cluster micro-front end in the new browser tab. |
| **spec.navigationNodes.showInNavigation** | **NO** | The Boolean that specifies if the cluster micro front-end's node is visible in the navigation or not. |
| **spec.navigationNodes.requiredPermissions** | **NO** | Specifies the list of permissions (RBAC rules) that determine if the navigation node should be shown for the current user.  |
