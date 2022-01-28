# Eventing Test


## Overview 

This test covers the end-to-end flow for Eventing. It is divided in three parts:
1. `eventing-test-prep.js` - prepares all the resources, mocks and assets required for tests to be executed
2. `eventing-test.js` - the actual tests
3. `eventing-test-cleanup.js` - removes the test resources and namespaces from the cluster

## Usage
To run Eventing-specific tests locally, follow these steps:

1. Install dependencies:
```bash
cd kyma/tests/fast-integration
npm install
```

2. Prepare the `.env` file based on the `.env.template`. 
```
KYMA_TYPE=OSS                # OSS or SKR
KUBECONFIG=                  # Kyma cluster kubeconfig file path
EVENTMESH_SECRET_FILE=       # Event Mesh Credentials JSON file path
COMPASS_HOST=                # Only required if KYMA_TYPE=SKR
COMPASS_CLIENT_ID=           # Only required if KYMA_TYPE=SKR
COMPASS_CLIENT_SECRET=       # Only required if KYMA_TYPE=SKR
COMPASS_TENANT=              # Only required if KYMA_TYPE=SKR
GARDENER_KUBECONFIG=         # Only required if KYMA_TYPE=SKR
```
>**NOTE:** The Event Mesh Credentials JSON file can be downloaded from the BTP Cockpit under your subaccount instances.

3. Run the following command to set up the environment variables in your system:
```bash
export $(xargs < .env)
```

4.  Execute the Eventing tests:
```bash
npm run test-eventing
```