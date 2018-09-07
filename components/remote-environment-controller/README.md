# Remote Environment Controller

## Overview

Remote Environment Controller detects changes in Remote Environment custom resources and acts accordingly.


## Performed operations

Remote Environment Controller performs different operations as a result of the following events:

 - Remote Environment created - Controller installs Helm chart containing all the necessary Kubernetes resources.
 - Remote Environment deleted - Controller deletes Helm chart corresponding to the given RE.
 