---
title: Test addons
type: Details
---

If you want to test whether your addons are correct, you can use the dry run mode to check the generated manifests of the chart without installing it.

The `--debug` option prints the generated manifests. As a prerequisite, you must install [Helm](https://github.com/kubernetes/helm) on your machine to run this command:

```bash
 helm install --dry-run {path-to-chart} --debug
```

You can also use this option for the basic debugging purposes, when your addons are installed but you fail to provision ServiceInstances from them. For more details, read the Helm [official documentation](https://helm.sh/docs/chart_template_guide/debugging/).
