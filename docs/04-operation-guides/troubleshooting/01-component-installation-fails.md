# Component Doesn't Work After Successful Installation {docsify-ignore-all}

## Symptom

The installation was successful but a component does not behave in the expected way.

## Cause

A Pod might be not running.

## Remedy

1. To check if all deployed Pods are running, run:

   ```bash
   kubectl get pods --all-namespaces
   ```

   The command retrieves all Pods from all namespaces, the status of the Pods, and their instance numbers.

2. Check if the status is `Running` for all Pods.
3. If any of the Pods that you need was not started successfully, install Kyma again.

If all Pods were started successfully but the problem persists, investigate the reason with the following commands:

- To get a detailed view of the installation process, use the `--verbose` flag.
- To tweak the values on a component level, use `deploy --components`: Pass a components list that includes only the components you want to test and try out the settings that work for your installation.
- To understand which component failed during deployment, *deactivate* the default atomic deployment: `--atomic=false`.
   With atomic deployment active, any component that hasn't been installed successfully is rolled back, which may make it hard to find out what went wrong. By disabling the flag, the failed components are not rolled back.
