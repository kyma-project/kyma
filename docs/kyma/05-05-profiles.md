---
title: Profiles
type: Configuration
---

The Kyma Operator allows you to use pre-defined profiles to install or upgrade Kyma. To install Kyma using a profile, you must specify the **spec.profile** attribute in the [Installation custom resource](#custom-resource-installation) (CR).

Profile is a set of helm values defined in `profile-{name}.yaml` file at the component chart folder. If the profile file is not present, the Kyma Operator will use default set of values from the `values.yaml` file. The following fragment of the `values.yaml` shows definition limitRange settings:

```yaml
limitRange:
  max:
    memory: 4Gi
  default:
    memory: 96Mi
  defaultRequest:
    memory: 32Mi
```

To override, for example `limitRange.max.memory` put following in the `profile-{name}.yaml` file: 

```yaml
limitRange:
  max:
    memory: 5Gi
```

>**NOTE:** If the `profile-{name}.yaml` file doesn't exist in a component chart simply create it.

On top of the applied profile the Kyma Operator will put user-defined overrides. For more information, see the [overrides section](#configuration-helm-overrides-for-kyma-installation).