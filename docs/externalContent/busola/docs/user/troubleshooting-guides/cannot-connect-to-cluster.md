# Cannot Connect to Cluster

## Symptom

<!-- Describe the problem from the user's perspective. Provide the undesirable condition or symptom that the user may want to correct. This could be an error message or an undesirable state.
-->

When I connect to the Kubernetes cluster by Busola, I get a connection error.

## Cause

Connection problems can be caused by:

- Incorrect kubeconfig
- The Busola backend is not allowed to connect to the cluster

## Solution

1. Check the correctness of the provided kubeconfig using [kubectl](https://kubernetes.io/docs/reference/kubectl/) by executing any command.
2. Make sure that the cluster can accept a connection from the Busola domain.
