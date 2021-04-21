---
title: Sudden container failure
type: Troubleshooting
---

The container can suddenly fail when you use the `kyma run function` command with these flags:
    - `runtime=Nodejs12` or `runtime=Nodejs10`
    - `debug=true`
    - `hot-deploy=true`

In such a case, you can see the `[nodemon] app crashed` message in the container's logs.

If you use Kyma in Kubernetes, Kubernetes itself should run the Function in the container.
If you use Kyma without Kubernetes, you have to rerun the container yourself. 
