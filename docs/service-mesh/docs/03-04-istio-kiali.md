---
title: Istio kiali 
type: Details
---

Kiali is the utility to visualize Istio's service-mesh. 

**NOTE** If you enable Kiali, enable `prometheus-operator` and `monitoring` components first. The Kiali installer generates a random password for the `admin` user and stores it in a Secret. Use the `admin` user to login to the Kiali UI. Run this command to get the password:
```
$ kubectl -n istio-system get secret kiali -o json | jq -r .data.passphrase | base64 -D | awk '{print "Password:", $1}' 
```
