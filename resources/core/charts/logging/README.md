
```
  _                       _             
 | |                     (_)            
 | |     ___   __ _  __ _ _ _ __   __ _ 
 | |    / _ \ / _` |/ _` | | '_ \ / _` |
 | |___| (_) | (_| | (_| | | | | | (_| |
 |______\___/ \__, |\__, |_|_| |_|\__, |
               __/ | __/ |         __/ |
              |___/ |___/         |___/ 
```

## Overview
This document explains how Kyma installs `OK Log` and `Logspout` in the `kyma-system` namespace, and how to use it to check logs in Kyma.

## Details
The chart installs the following resources in `kyma-system` namespace:
* Configmap
* Daemonset
* Service
* Statefulset

**OK Log** is a distributed and coordination-free log management system. It's a light weight system with a simple UI.

**Logspout** is a log router for Docker containers that runs as a daemonset. It attaches to all containers on a node, then routes their logs to **OKLog**. 


#### Access logs in OK Log UI
1. Port-forward port **7650**

```bash
kubectl -n kyma-system port-forward svc/core-logging-oklog 7650:7650
```
2. Access OK Log UI in [here](http://localhost:7650/ui).
3. Use a plaintext or regex to pull up logs. E.g.  `error`


## Troubleshooting
- Check whether `Logspout` is pulling logs from docker containers
  1. Start a shell in the Logspout pod
  2. A HTTP GET call to endpoint `http://localhost:80/logs` should print all the logs from the current containers
- Check Logspout logs to make sure it is configured correctly to feed the logs to the `ingest-fast` port of OK Log
```bash
kubectl -n kyma-system logs <Logspout-pod-name>
```

## References
- More details on OK Log is [here](https://github.com/oklog/oklog)
- More details on Logspout is [here](https://github.com/gliderlabs/logspout)