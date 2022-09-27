---
title: Available presets
---

Function's resources and replicas as well as resources for image-building Jobs are based on presets. A preset is a predefined group of values. There are two groups of presets defined for a Function CR and include the presents for:

- Function's resources
- Image-building Job's resources

## Configuration

To add a new preset to the Serverless configuration for the defaulting webhook to set it on all Function CRs, update the `values.yaml` file in the Serverless chart. To do it, change the configuration for the **webhook.values.function.resources.presets** or **webhook.values.buildJob.resources.presets** parameters. Read the [Serverless chart configuration](./00-configuration-parameters/svls-01-serverless-chart.md) to find out more.

## Usage

If you want to apply values from a preset to a single Function, override the existing values for a given preset in the Function CR. To do it, first remove the relevant fields from the Function CR and then add the relevant preset labels. For example, to modify the default values for **buildResources**, remove all its entries from the Function CR and add an appropriate **serverless.kyma-project.io/build-resources-preset: {PRESET}** label to the Function CR.

### Function's resources

| Preset | Request CPU | Request memory | Limit CPU | Limit memory |
| - | - | - | - | - |
| `XS` | `50m` | `64Mi` | `100m` | `128Mi` |
| `S` | `100m` | `128Mi` | `200m` | `256Mi` |
| `M` | `200m` | `256Mi` | `400m` | `512Mi` |
| `L` | `400m` | `512Mi` | `800m` | `1024Mi` |
| `XL` | `800m` | `1024Mi` | `1600m` | `2048Mi` |

To apply values ​​from a given preset, use the **serverless.kyma-project.io/function-resources-preset: {PRESET}** label in the Function CR.

### Build Job's resources

| Preset | Request CPU | Request memory | Limit CPU | Limit memory |
| - | - | - | - | - |
| `local-dev` | `200m` | `200Mi` | `400m` | `400Mi` |
| `slow` | `200m` | `200Mi` | `700m` | `700Mi` |
| `normal` | `700m` | `700Mi` | `1100m` | `1100Mi`|
| `fast` | `1100m` | `1100Mi` | `1700m` | `1100Mi`|

To apply values ​​from a given preset, use the **serverless.kyma-project.io/build-resources-preset: {PRESET}** label in the Function CR.
