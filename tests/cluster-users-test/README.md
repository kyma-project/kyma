# Cluster-users-test

## Content
- Contains the `Dockerfile` for the image used in Kyma cluster-users tests.
- Contains the `sar-test.sh` script that runs tests for a chart.

## Details
This tests use the `kubectl` tool to perform SAR (SubjectAccessReview) calls to the api-server. As the tests verify if a user can access resources it is supposed to, we perform the whole login procedure using Dex. 

**Test Scenario**:
The tests ask the api-server if a specific user can perform specific operations on a specific object, an asserts the resuls:

```bash
echo "--> developer@kyma.cx should be able to get specific CRD"
testPermissions "developer@kyma.cx" "get" "crd/installations.installer.kyma-project.io" "yes"

echo "--> developer@kyma.cx should NOT be able to list ClusterRole"
testPermissions "developer@kyma.cx" "list" "clusterrole" "no"
```