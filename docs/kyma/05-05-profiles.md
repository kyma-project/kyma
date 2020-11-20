---
title: Profiles
type: Configuration
---

The Kyma Operator allows you to use pre-defined profiles to install or upgrade Kyma. To install Kyma using a profile, you must specify the **spec.profile** attribute in the [Installation custom resource](#custom-resource-installation) (CR).

The profile is a subset of the chart's `values.yaml` defined in `profile-{name}.yaml` file at the component chart root folder.  For example, the file `profile-evaluation.yaml` defines settings for `evaluation` profile. Values from the profile override settings from `values.yaml`. A profile can override only a section, or even the whole file if required. If the profile file is not present, the Kyma Operator will use default set of values from the `values.yaml` file.
Currently supported profiles are: 
- evaluation
- production

The following fragment of the `values.yaml` shows definition limitRange settings:

```yaml
limitRange:
  max:
    memory: 4Gi
  default:
    memory: 96Mi
  defaultRequest:
    memory: 32Mi
```

To override `limitRange.max.memory` for the `production` profile put the following in the `profile-production.yaml` file: 

```yaml
limitRange:
  max:
    memory: 5Gi
```

>**NOTE:** If the `profile-{name}.yaml` file doesn't exist in a component chart you can simply create it.

On top of the applied profile the Kyma Operator will put user-defined overrides. For more information, see the [overrides section](#configuration-helm-overrides-for-kyma-installation).
