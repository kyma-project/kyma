# Cluster-users-test

## Content
- Contains the `Dockerfile` for the image used in Kyma cluster-users tests.
- Contains the `sar-test.sh` script that runs tests for a chart.

## Details
The test login into `Dex` using the static users emails and create calls to `iam-kubeconfig-service` to generate a *kubeconfig* for each of them. Afterwards `kubectl` tool is used to perform SAR (SubjectAccessReview) calls to the api-server.

**Test Scenario**:
The tests ask the api-server if a specific user can perform specific operations on a specific object, an asserts the resuls:

```bash
EMAIL=${DEVELOPER_EMAIL} PASSWORD=${DEVELOPER_PASSWORD} getConfigFile
export KUBECONFIG="${PWD}/kubeconfig"

echo "--> ${DEVELOPER_EMAIL} should be able to get Deployments in ${NAMESPACE}" testPermissions "get" "deployment" "${NAMESPACE}" "yes"
```
