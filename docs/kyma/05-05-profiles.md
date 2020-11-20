---
title: Profiles
type: Configuration
---

The Kyma Operator allows you to use pre-defined profiles to install or upgrade Kyma.
Currently supported profiles are: 
- Evaluation - a profile with limited resources that you can use for trial purposes
- Production - a profile with full resource that you can use on production


The profile is a subset of the chart's `values.yaml` defined in the `profile-{name}.yaml` file at the component chart root folder.  For example, the `profile-evaluation.yaml` file defines settings for the `evaluation` profile. Values from the profile override settings from `values.yaml`. A profile can override not only a section but also the whole file, if required. If the profile file is not present, the Kyma Operator will use the default set of values from the `values.yaml` file.

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

>**NOTE:** If the `profile-{name}.yaml` file doesn't exist in the component chart, you can simply create it.

User-provided overrides have precedence over the applied profile. For more information, see the [overrides section](#configuration-helm-overrides-for-kyma-installation).