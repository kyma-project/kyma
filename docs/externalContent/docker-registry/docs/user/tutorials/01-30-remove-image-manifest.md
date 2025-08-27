# Remove Image from Registry

This tutorial shows how to remove a previously pushed image to the registry using the pure registry API or [skopeo](https://github.com/containers/skopeo).

## Prerequsities

* [kubectl](https://kubernetes.io/docs/tasks/tools/)
* [Kyma CLI v3](https://github.com/kyma-project/cli/)
* [skopeo](https://github.com/containers/skopeo/) (for the skopeo scenario)
* [curl](https://curl.se/) (for the registry API scenario)

## Steps

1. Enable the image manifests deletion functionality by changing the **.spec.storage.deleteEnabled** flag to `true`:

```bash
kubectl apply -n kyma-system -f - <<EOF
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
    name: default
    namespace: kyma-system
spec:
    storage:
        deleteEnabled: true
EOF
```

Once the DockerRegistry CR becomes `Ready`, you see the updated **.status.deleteEnabled** field with a new value.

```yaml
...
status:
    deleteEnabled: true
```

2. Push the image to the registry:

```bash
kyma alpha registry image-import <IMAGE_NAME>:<IMAGE_TAG>
```

3. Port-forward the registry service to another terminal:

```bash
kubectl port-forward -n kyma-system svc/dockerregistry 5000:5000
```

4. Export registry credentials:

```bash
export DR_USERNAME=$(kubectl get secret -n kyma-system dockerregistry-config -o jsonpath="{.data.username}" | base64 -d)
export DR_PASSWORD=$(kubectl get secret -n kyma-system dockerregistry-config -o jsonpath="{.data.password}" | base64 -d)
```

<Tabs>
<Tab name="Registry API">

5. Verify that the image was pushed to the registry and exists with the given tag:

```bash
curl -u "$DR_USERNAME:$DR_PASSWORD" -sS localhost:5000/v2/<IMAGE_NAME>/tags/list
```

6. Get the tag digest:

```bash
curl -u "$DR_USERNAME:$DR_PASSWORD" -o /dev/null -w '%header{Docker-Content-Digest}' -H 'Accept: application/vnd.docker.distribution.manifest.v2+json' -sS localhost:5000/v2/<IMAGE_NAME>/manifests/0.1
```

7. Remove the tag using digest from the previous step:

```bash
curl -u "$DR_USERNAME:$DR_PASSWORD" -X DELETE localhost:5000/v2/dr/manifests/<DIGEST>
```
</Tab>
<Tab name="skopeo">

5. Verify that the image was pushed to the registry and exists with given tag:

```bash
skopeo list-tags --creds "$DR_USERNAME:$DR_PASSWORD" --tls-verify=false docker://localhost:5000/<IMAGE_NAME>
```

6. Remove the tag using digest from previous step:

```bash
skopeo delete --creds "$DR_USERNAME:$DR_PASSWORD" --tls-verify=false docker://localhost:5000/<IMAGE_NAME>:<IMAGE_TAG>
```
</Tab>
</Tabs>
