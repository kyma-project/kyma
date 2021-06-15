---
title: Available presets
---

Function's resources and replicas as well as resources for image-building Jobs are based on presets. A preset is a predefined group of values. There are three groups of presets defined for a Function CR and include the presents for:

- Function's resources
- Function's replicas
- Image-building Job's resources

## Configuration

To add a new preset to the Serverless configuration for the defaulting webhook to set it on all Function CRs, update the `values.yaml` file in the Serverless chart. To do it, change the configuration for the **webhook.values.function.replicas.presets**, **webhook.values.function.resources.presets** or **webhook.values.buildJob.resources.presets** parameters. Read the [Serverless chart configuration](./05-configuration-parameters/svls-01-serverless-chart.md) to find out more.

## Usage

If you want to apply values from a preset to a single Function, override the existing values for a given preset in the Function CR. To do it, first remove the relevant fields from the Function CR and then add the relevant preset labels. For example, to modify the default values for **buildResources**, remove all its entries from the Function CR and add an appropriate **serverless.kyma-project.io/build-resources-preset: {PRESET}** label to the Function CR.

### Function's replicas

| Preset | Minimum number | Maximum number |
| - | - | - |
| `S` | 1 | 1 |
| `M` | 1 | 2 |
| `L` | 1 | 5 |
| `XL` | 1 | 10 |

To apply values ​​from a given preset, use the **serverless.kyma-project.io/replicas-preset: {PRESET}** label in the Function CR.

### Function's resources

| Preset | Request CPU | Request memory | Limit CPU | Limit memory |
| - | - | - | - | - |
| `XS` | `10m` | `16Mi` | `25m` | `32Mi` |
| `S` | `25m` | `32Mi` | `50m` | `64Mi` |
| `M` | `50m` | `64Mi` | `100m` | `128Mi` |
| `L` | `100m` | `128Mi` | `200m` | `256Mi` |
| `XL` | `200m` | `256Mi` | `400m` | `512Mi` |

To apply values ​​from a given preset, use the **serverless.kyma-project.io/function-resources-preset: {PRESET}** label in the Function CR.

### Build Job's resources

| Preset | Request CPU | Request memory | Limit CPU | Limit memory |
| - | - | - | - | - |
| `local-dev` | `200m` | `200Mi` | `400m` | `400Mi` |
| `slow` | `400m` | `400Mi` | `700m` | `700Mi` |
| `normal` | `700m` | `700Mi` | `1100m` | `1100Mi`|
| `fast` | `1100m` | `1100Mi` | `1700m` | `1700Mi`|

To apply values ​​from a given preset, use the **serverless.kyma-project.io/build-resources-preset: {PRESET}** label in the Function CR.
