---
title: Helm Broker chart
---

To configure the Helm Broker chart, override the default values of its `values.yaml` file. This document describes parameters that you can configure.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](todo)
>* [Top-level charts overrides](todo)

## Configurable parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| **ctrl.resources.limits.cpu** | Defines limits for CPU resources. | `100m` |
| **ctrl.resources.limits.memory** | Defines limits for memory resources. During the clone action, the Git binary loads the whole repository into memory. You may need to adjust this value if you want to clone a bigger repository.| `76Mi` |
| **ctrl.resources.requests.cpu** | Defines requests for CPU resources. | `80m` |
| **ctrl.resources.requests.memory** | Defines requests for memory resources. | `32Mi` |
| **ctrl.tmpDirSizeLimit** | Specifies a size limit on the `tmp` directory in the Helm Pod. This directory is used to store processed addons. Eviction manager monitors the disk space used by the Pod and evicts it when the usage exceeds the limit. Then, the Pod is marked as `Evicted`. The limit is enforced with a time delay, usually about 10s. | `1Gi` |
| **global.cfgReposUrlName** | Specifies the name of the default ConfigMap which provides the URLs of addons repositories. | `helm-repos-urls` |
| **global.isDevelopMode** | Defines whether to accept URL prefixes from the **global.urlRepoPrefixes.additionalDevelopMode** list. If set to `true`, the Helm Broker accepts the prefixes from the list. | `false` |
| **global.urlRepoPrefixes.default** | Defines a list of accepted prefixes for repository URLs. | `'https://', 'git::', 'github.com/', 'bitbucket.org/'` |
| **global.urlRepoPrefixes.additionalDevelopMode** | Defines a list of accepted prefixes for repository URLs when develop mode is enabled. | `'http://'` |
| **additionalAddonsRepositories.myRepo** | Provides a map of additional ClusterAddonsConfiguration repositories to create by default. The key is used as a name and the value is used as a URL for the repository. | `github.com/myOrg/myRepo//addons/index.yaml` |
