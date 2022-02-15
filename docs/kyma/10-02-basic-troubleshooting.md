---
title: Basic troubleshooting
type: Troubleshooting
---

## Console UI password

If you forget the password for the **admin@kyma.cx**, you can get it from the `admin-user` Secret located in the `kyma-system` Namespace. Run this command:

```bash
kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode
```

## Kyma Installer doesn't respond as expected

If the Installer does not respond as expected, check the installation status using the `is-installed.sh` script with the `--verbose` flag added. Run:

```bash
scripts/is-installed.sh --verbose
```

## Installation successful, component not working

If the installation is successful but a component does not behave in the expected way, inspect Helm releases for more details on all of the installed components.

Run this command to list all of the available Helm releases:

```bash
helm list --all-namespaces --all
```

Run this command to get more detailed information about a given release:

```bash
helm status {RELEASE_NAME} -n {RELEASE_NAMESPACE}
```

>**NOTE:** Names of Helm releases correspond to names of Kyma components.


Additionally, see if all deployed Pods are running. Run this command:

```bash
kubectl get pods --all-namespaces
```

The command retrieves all Pods from all Namespaces, the status of the Pods, and their instance numbers. Check if the status is `Running` for all Pods. If any of the Pods that you require do not start successfully, install Kyma again.

## Can't log in to the Console after hibernating the Minikube cluster

If you put a local cluster into hibernation or use `minikube stop` and `minikube start` the date and time settings of Minikube get out of sync with the system date and time settings. As a result, the access token used to log in cannot be properly validated by Dex and you cannot log in to the console. To fix that, set the date and time used by your machine in Minikube. Run:

```bash
minikube ssh -- docker run -i --rm --privileged --pid=host debian nsenter -t 1 -m -u -n -i date -u $(date -u +%m%d%H%M%Y)
```

## Errors after restarting Kyma on Minikube

If you restart Kyma using unauthorized methods, such as triggering the installation when a Minikube cluster with Kyma is already running, the cluster might become unresponsive which can be fixed by reinstalling Kyma.
To prevent such behavior, [stop and restart Kyma](#installation-install-kyma-locally-stop-and-restart-kyma-without-reinstalling) using only the method described.

## Can't deprovision Gardener cluster

If you are unable to deprovision a Gardener cluster, you might receive the following error:

```bash
Flow "Shoot cluster deletion" encountered task errors: [task "Cleaning extended API groups" failed: 1 error occurred:
retry failed with context deadline exceeded, last error: remaining objects are still present: [*v1beta1.CustomResourceDefinition /installations.installer.kyma-project.io]
```

If this happens, you must remove the finalizer from the `kyma-installation` CR before you deprovision the cluster. Run this command: 

```bash
kubectl patch installation kyma-installation --type=merge -p '{"metadata":{"finalizers":null}}'
```
