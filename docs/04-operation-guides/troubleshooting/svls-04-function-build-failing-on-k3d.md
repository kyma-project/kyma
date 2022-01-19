---
title: Functions failing to build on k3d
---


## Symptom

There are rare cases, for some k3d versions and configurations, where users experience Functions failing to be built.

If you see that your Function cannot be built,
```
$ kubectl get functions.serverless.kyma-project.io nyfun
NAME    CONFIGURED   BUILT   RUNNING   RUNTIME    VERSION   AGE
myfun   True         False             nodejs14   1         3h15m
```
and the Function build job shows the following error, meaning that your host k3d environment is likely to experience the problem.
```
$ kubectl logs myfun-build-zqhk8-7xl6h
kaniko should only be run inside of a container, run with the --force flag if you are sure you want to continue
```

## Cause

This problem originates in kaniko - the container image build tool used in Kyma. kaniko is designed to be run in a container and this is how it is executed in Kyma (as build jobs).
kaniko has a detection mechanism to verify whether the build is actually executed in a container and fails in case it is not.
This detection mechanism has issues and in some circumstances (i.e hosts with cgroups in version 2 or other, not yet clearly identified) it shows false positive result. 

Related issues:
 - https://github.com/kyma-project/kyma/issues/13051
 - https://github.com/GoogleContainerTools/kaniko/issues/1592
 
## Remedy

kaniko can be executed with `--force` flag that skips the verification. To do so, please override the kaniko execution arguments by passing `--force` flag.

Introduce a file with overrides, i.e `my-overrides.yaml`
```yaml
serverless:
  containers:
    manager:
      envs:
        functionBuildExecutorArgs:
          value: --insecure,--skip-tls-verify,--skip-unused-stages,--log-format=text,--cache=true,--force
```

Use the file to override default configuration while deploying kyma on your k3d instance:
```bash
kyma deploy --values-file my-overrides.yaml
```

When used, the build would succeed, but kaniko would assume that it is run outside a container and would produce this log:
```
$ k logs myfun-build-gmzhz-9rnmm
time="2022-01-14T08:25:16Z" level=warning msg="kaniko is being run outside of a container. This can have dangerous effects on your system"
```
