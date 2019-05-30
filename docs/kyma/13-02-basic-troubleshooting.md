---
title: Basic troubleshooting
type: Troubleshooting
---

## Console UI password

If you don't set the password for the **admin@kyma.cx** user using the `--password` parameter or you forget the password you set, you can get it from the `admin-user` Secret located in the `kyma-system` Namespace. Run this command:

```
kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode
```

## Installer doesn't respond as expected

If the Installer does not respond as expected, check the installation status using the `is-installed.sh` script with the `--verbose` flag added. Run:

```
scripts/is-installed.sh --verbose
```

## Installation successful, component not working

If the installation is successful but a component does not behave in an expected way, see if all deployed Pods are running. Run this command:

```
kubectl get pods --all-namespaces
```
The command retrieves all Pods from all Namespaces, the status of the Pods, and their instance numbers. Check if the STATUS column shows Running for all Pods. If any of the Pods that you require do not start successfully, perform the installation again.

## Can't login to the Console after hibernating Minikube cluster

If you put your local running cluster into hibernation or use `minikube stop` and `minikube start` the date and time settings of Minikube get out of sync with the system date and time settings. As a result, the access token used to log in cannot be properly validated by Dex and you cannot log in to the console. To fix that, set the date and time used by your machine in Minikube. Run:

```
minikube ssh -- docker run -i --rm --privileged --pid=host debian nsenter -t 1 -m -u -n -i date -u $(date -u +%m%d%H%M%Y)
```

## Errors after restarting Kyma on Minikube

If you restart Kyma using unauthorized methods, such as triggering the installation when a Minikube cluster with Kyma is already running, the cluster might become unresponsive which can be fixed by reinstalling Kyma.
To prevent such behavior, stop and restart Kyma using only the method described [here](#installation-install-kyma-locally-stop-and-restart-kyma-without-reinstalling).
