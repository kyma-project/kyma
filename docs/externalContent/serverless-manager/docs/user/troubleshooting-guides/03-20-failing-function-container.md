# Failing Function Container

## Symptom

The container suddenly fails when you use the `kyma run function` command with these flags:

- `runtime=nodejs20`
- `debug=true`
- `hot-deploy=true`

In such a case, you can see the `[nodemon] app crashed` message in the container's logs.

## Solution

If you use Kyma in Kubernetes, Kubernetes itself should run the Function in the container.
If you use Kyma without Kubernetes, you have to rerun the container yourself.
