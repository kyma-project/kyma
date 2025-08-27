# Import Image Into Kyma's Internal Docker Registry

> [!NOTE]
> To use the following `image-import` command, you must [install the Docker Registry module](https://github.com/kyma-project/docker-registry?tab=readme-ov-file#install) on your Kyma runtime

```sh
docker pull kennethreitz/httpbin

kyma alpha registry image-import kennethreitz/httpbin:latest
```

Run a Pod from a locally hosted image

```sh
kubectl run my-pod --image=localhost:32137/kennethreitz/httpbin:latest --overrides='{ "spec": { "imagePullSecrets": [ { "name": "dockerregistry-config" } ] } }'
```
