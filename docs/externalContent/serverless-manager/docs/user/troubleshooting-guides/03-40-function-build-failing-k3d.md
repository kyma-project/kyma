# Functions Failing to Build on k3d

## Symptom

There are rare cases, for some k3d versions and configurations, where users experience Functions failing to be built.

Your Function cannot be built and you get the following output:

   ```bash
   $ kubectl get functions.serverless.kyma-project.io nyfun
   NAME    CONFIGURED   BUILT   RUNNING   RUNTIME    VERSION   AGE
   myfun   True         False             nodejs20   1         3h15m
   ```

Additionally, the Function build job shows the following error, meaning that your host k3d environment is likely to experience the problem:

   ```bash
   $ kubectl logs myfun-build-zqhk8-7xl6h
   kaniko should only be run inside of a container, run with the --force flag if you are sure you want to continue
   ```

## Cause

This problem originates in [kaniko](https://github.com/GoogleContainerTools/kaniko) - the container image build tool used in Kyma. kaniko is designed to be run in a container and this is how it is executed in Kyma (as build jobs).
kaniko has a detection mechanism to verify whether the build is actually executed in a container and fails in case it is not.
This detection mechanism has issues and in some circumstances (that is, hosts with cgroups in version 2 or other, not yet clearly identified) it shows a false positive result.

Related issues:

- https://github.com/kyma-project/kyma/issues/13051
- https://github.com/GoogleContainerTools/kaniko/issues/1592

## Remedy

You can force kaniko to skip the verification by overriding the kaniko execution arguments with the `--force` flag.

Introduce a file with overrides, for example `my-overrides.yaml`.

   ```yaml
   serverless:
     containers:
       manager:
         envs:
           functionBuildExecutorArgs:
             value: --insecure,--skip-tls-verify,--skip-unused-stages,--log-format=text,--cache=true,--use-new-run,--compressed-caching=false,--force
   ```

Use the file to override the default configuration while deploying Kyma on your k3d instance:

   ```bash
   kyma deploy --values-file my-overrides.yaml
   ```
