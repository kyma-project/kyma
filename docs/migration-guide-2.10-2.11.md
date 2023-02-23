---
title: Migration Guide 2.10-2.11
---

## Observability

### Kiali deprecation

Kiali was deprecated with Kyma 2.8 and has been removed with Kyma 2.11. After upgrading to Kyma 2.11, run the script [2.10-2.11-cleanup-kiali.sh](https://github.com/kyma-project/kyma/blob/main/docs/assets/2.10-2.11-cleanup-kiali.sh) to remove Kiali. 
> **NOTE** Users who want to continue using Kiali should deploy the [Kiali example](https://github.com/kyma-project/examples/tree/main/kiali).

## Service Management

### PodPreset cleanup

The PodPreset component was deprecated in [Kyma 2.4](https://kyma-project.io/blog/2022/6/30/release-notes-24#pod-preset-deprecation-note) and removed from [Kyma 2.10](https://github.com/kyma-project/kyma/pull/16647). Run the [cleanup script](./assets/2.10-2.11-cleanup-podpreset.bash) to remove any residual resources related to PodPresets. Follow [the guide](https://kyma-project.io/blog/2022/6/30/release-notes-24#pod-preset-deprecation-note) to transform the usage of `Secrets` from Kyma `Functions` manually.
