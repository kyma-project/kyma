# Cluster-users-test

## Content
- Contains the `Dockerfile` for the image used in Kyma cluster-users tests.
- Contains the `sar-test.sh` script that runs tests for a chart.

## Details
The test logs in to Dex using the static user's emails and creates calls to the IAM Kubeconfig Service to generate a `kubeconfig` file for each call. Afterwards, `kubectl` tool is used to perform SAR (SubjectAccessReview) calls to the api-server.

### Test Scenario
The tests ask the api-server if a specific user can perform specific operations on a specific object, and asserts the results:

```bash
EMAIL=${DEVELOPER_EMAIL} PASSWORD=${DEVELOPER_PASSWORD} getConfigFile
export KUBECONFIG="${PWD}/kubeconfig"

echo "--> ${DEVELOPER_EMAIL} should be able to get Deployments in ${NAMESPACE}" testPermissions "get" "deployment" "${NAMESPACE}" "yes"
```
