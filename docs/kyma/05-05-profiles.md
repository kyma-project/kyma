---
title: Profiles
type: Configuration
---

The Kyma Operator allows you to use pre-defined profiles to install or upgrade Kyma. To install Kyma using a profile, you must specify the **spec.profile** attribute in the [Installation custom resource](#custom-resource-installation) (CR).

Profile is a set of helm values defined in `profile-{name}.yaml` file at the component chart folder. If the profile file is not present, the Kyma Operator will use default set of values from the `values.yaml` file.

//example of profile file

On top of the applied profile the Kyma Operator will put user-defined overrides.

//link to overrides