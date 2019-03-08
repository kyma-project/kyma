# Resources                                                                                  

## Overview

Resources are all components in Kyma that are available for local and cluster installation. You can find more details about each component in the corresponding README.md files.

Resources currently include, but are not limited to, the following:

- Elements which are essential for the installation of `core` components in Kyma, such as certificates, users, and permissions
- Examples of the use of specific components
- Scripts for the installation of Helm, Istio deployment, as well as scripts for validating Pods, starting Kyma, and testing

## Development

Each component, test or tool from Kyma repository whose image is used in Kyma contains Makefile. 
Makefile is used to build image component and push it to the external repository. 
Makefile has to also contains command which return path to `values.yaml` file localization with version of the actual image is used in Kyma.
Here is example of command in Makefile for `service-binding-usage-controller`

```
.PHONY: path-to-referenced-charts
path-to-referenced-charts:
    @echo "resources/service-catalog-addons"
```

The command shows path to `values.yaml` file where version of component used in Kyma is set:
Here is example of file `resources/service-catalog-addons/`

```
global:
  containerRegistry:
    path: eu.gcr.io/kyma-project
  istio:
    gateway:
      name: kyma-gateway
  service_binding_usage_controller:
    dir: develop/
    version: d1930a3d
```

Version is localized under the key `global.<name_of_component>.version`. 
`name_of_component` is the directory name of component where dashes are replace by undercourses. 
For exmaple component directory `service-binding-usage-controller` is changed to `service_binding_usage_controller`
