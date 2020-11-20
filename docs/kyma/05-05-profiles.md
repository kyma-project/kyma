---
title: Profiles
type: Configuration
---

The Kyma Operator allows you to use pre-defined profiles to install or upgrade Kyma. A profile is defined globally for the whole Kyma installation or upgrade. It's not possible to set a profile per single component.
The currently supported profiles are: 
- Evaluation - a profile with limited resources that you can use for trial purposes
- Production - a profile with full resources that you can use on production

A profile is a subset of the chart's `values.yaml` defined in the `profile-{name}.yaml` file at the component chart root folder. For example, the `profile-evaluation.yaml` file defines settings for the `evaluation` profile. Values from the profile override settings from `values.yaml`. A profile can override not only a section but also the whole file if necessary. If the profile file is not present, the Kyma Operator uses the default set of values from the `values.yaml` file.

For example, the following fragment of the `values.yaml` file defines the **limitRange** settings:

```yaml
limitRange:
  max:
    memory: 4Gi
  default:
    memory: 96Mi
  defaultRequest:
    memory: 32Mi
```

To override **limitRange.max.memory** for the `production` profile, add the following to the `profile-production.yaml` file: 

```yaml
limitRange:
  max:
    memory: 5Gi
```

>**NOTE:** If the `profile-{name}.yaml` file doesn't exist in the component chart, you can simply create it. However, there can be only one file for a given profile. 

User-provided overrides have precedence over the applied profile. For more information, see the [overrides section](#configuration-helm-overrides-for-kyma-installation).
