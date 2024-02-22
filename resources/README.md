# Resources

## Overview

Resources are all components in Kyma that are available for local and cluster installation. You can find more details about each component in the corresponding README.md files.

Resources currently include, but are not limited to, the following:

- Elements which are essential for the installation of `core` components in Kyma, such as certificates, users, and permissions
- Examples of the use of specific components
- Scripts for the installation of Helm, Istio deployment, as well as scripts for validating Pods, starting Kyma, and testing

## Development

Every component, test, or tool in the `kyma` repository contains a Makefile. A Makefile is used to build an image of a given component and to push it to the external repository. Every time you create a new component, test, or tool, ensure that its Makefile contains a path to the `values.yaml` file which informs about the actual image version used in Kyma.
To do so, add this entry to the Makefile:

```
.PHONY: path-to-referenced-charts
path-to-referenced-charts:
    @echo "{path to the referenced charts}"
```

The version of the actual component image is located under the **global.{name_of_component}.version** property.
**{name_of_component}** is a directory name of the component where dashes are replaced by underscores.
