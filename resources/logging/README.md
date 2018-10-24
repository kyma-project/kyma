# Logging

## Overview
This document explains how Kyma installs `OK Log` and `Logspout` in the `kyma-system` Namespace, and how to use it to check logs in Kyma.


## Troubleshooting
- Make sure that `Logspout` is pulling logs from docker containers:
  1. Start a shell in the Logspout Pod
  2. A HTTP GET call to endpoint `http://localhost:80/logs` should print all the logs from the current containers
- Check Logspout logs to make sure it is configured correctly to feed the logs to the `ingest-fast` port of OK Log
```bash
kubectl -n kyma-system logs <Logspout-pod-name>
```

## References
- Read more [details on OK Log](https://github.com/oklog/oklog).
- Read more [details on Logspout](https://github.com/gliderlabs/logspout).
