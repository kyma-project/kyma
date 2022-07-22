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
TEST_SUFFIX=                 # Required for every run on SKR, must be 4 characters long
KUBECONFIG=                  # Kyma cluster kubeconfig file path
EVENTMESH_SECRET_FILE=       # Event Mesh Credentials JSON file path
COMPASS_HOST=                # Only required if KYMA_TYPE=SKR
COMPASS_CLIENT_ID=           # Only required if KYMA_TYPE=SKR
COMPASS_CLIENT_SECRET=       # Only required if KYMA_TYPE=SKR
COMPASS_TENANT=              # Only required if KYMA_TYPE=SKR
GARDENER_KUBECONFIG=         # Only required if KYMA_TYPE=SKR
```
>**IMPORTANT:** The `TEST_SUFFIX` is required for every test run for SKR cluster. It needs to be 4 characters long, as it is the name for the compass scenario.
> The eventing tests add a scenario with the `"test-${TEST_SUFFIX}"` name.

>**NOTE:** The Event Mesh Credentials JSON file can be downloaded from the BTP Cockpit under your subaccount instances.

3. Run the following command to set up the environment variables in your system:
```bash
export $(xargs < .env)
```

4. Execute the Eventing tests locally:

- **inside the OSS cluster:**
```bash
npm run test-eventing
```
- **inside the SKR cluster:**
```bash
npm run test-eventing TEST_SUFFIX=abcd
```
>**NOTE:** The `at least once` delivery test for JetStream is only run when the `STREAM_RETENTION_POLICY` is set to `limits` and the `CONSUMER_DELIVER_POLICY` is set to `all`.

## Troubleshooting ##

If the `TEST_SUFFIX` environment variable was not set during the test execution or a scenario with that suffix already exists, you will get the following error:
>_Update of API/Event is not supported yet for the compass scenario_.

To avoid this error, first get the list of the already existing scenarios:
```bash
 npm run eventing-get-registered-scenarios
```
Then, for the `TEST_SUFFIX` environment variable, assign a value that is not listed in the list of existing scenarios.
