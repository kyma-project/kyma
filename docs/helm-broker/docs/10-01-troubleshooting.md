---
title: Troubleshooting
---

### Possible `FAILED` status for created ServiceInstances

If your ServiceInstance creation was successful and yet the release is marked as `FAILED` on the releases list when running the `helm list` command, it means that there is an error on the Helm's side that was not passed on to the Helm Broker. To get the error details, check Tiller's logs.
