---
title: '"Failed to pull image" error'
type: Troubleshooting
---

When you try to install Kyma locally on Minikube, the installation may fail at a very early stage logging this error:

``` bash
POD: tiller-certs-job-cpl67: ERROR: Failed to pull image "eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20200121-4f3202bd": rpc error: code = Unknown desc = Error response from daemon: Get https://eu.gcr.io/v2/: dial tcp: lookup eu.gcr.io on 192.168.64.1:53: read udp 192.168.64.5:55778->192.168.64.1:53: read: connection refused
error
POD: tiller-deploy-f88ff7869-z5cmx: ERROR: MountVolume.SetUp failed for volume "tiller-certs" : secrets "tiller-secret" not found
```

This message shows that the installation fails because the required Docker image can't be downloaded from a Google Container Registry address. Minikube can't download the image because its DNS server can't resolve the image's address.

If you get this error, check if any process is listening on port `53`. Run:

``` bash
sudo lsof -i tcp:53
```

If the port is taken by a process other than Minikube, the output of this command will point you to the software causing the issue.

To fix this problem, try adjusting the configuration of the software that's blocking the port. In some cases, you might have to uninstall the software to free port `53`.

For example, [dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html) users can add `listen-address=192.168.64.1` to `dnsmasq.conf` to run dnsmasq and Minikube at the same time.

For more details, read [this](https://github.com/kubernetes/minikube/issues/3036) issue.
