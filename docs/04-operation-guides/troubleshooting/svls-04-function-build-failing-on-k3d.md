---
title: Functions failing to build on k3d
---


## Symptom

There are rare cases for some k3d versions and configurations where users experienced functions failing to be built.

If you see that your function cannot be built,
```
$ kubectl get functions.serverless.kyma-project.io nyfun
NAME    CONFIGURED   BUILT   RUNNING   RUNTIME    VERSION   AGE
myfun   True         False             nodejs14   1         3h15m
```
and you  see an error log from kaniko build execution from the build job than your host k3d environment is probably affected
```
$ kubectl logs myfun-build-zqhk8-7xl6h
kaniko should only be run inside of a container, run with the --force flag if you are sure you want to continue
```

## Cause

This problem originates in kaniko - the container image build tool used in kyma. Kaniko is designed to be run in container. Kyma complies with the recomendation and executes kaniko in container ( as build jobs ).
Kaniko has a detection mechanism to verify wheather the build is actually executed in a container and fails in case it is not.
This detection mechanism has some issues and for some circumstances (i.e hosts with cgroups in version 2 or other .. not yet clearly identidfied ) it results with false positive result. 

Related Issues:
 - https://github.com/kyma-project/kyma/issues/13051
 - https://github.com/GoogleContainerTools/kaniko/issues/1592
 
## Remedy

Kaniko can be executed with `--force` flag that skips the verification. To do so, please override the kaniko execution arguments by passing `--force` flag additionally.

```
kyma provision k3d
kyma deploy --value containers.manager.envs.functionBuildExecutorArgs="--insecure,--skip-tls-verify,--skip-unused-stages,--log-format=text,--cache=true,--force"
```

When used, the build would suceed, but Kaniko would assume it is run outside container and woould produce this log:
```
$ k logs myfun-build-gmzhz-9rnmm
time="2022-01-14T08:25:16Z" level=warning msg="kaniko is being run outside of a container. This can have dangerous effects on your system"
```
