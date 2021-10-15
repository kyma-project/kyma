# SKR test

This test covers deprovisioning of previous nightly SKR (SAP Kyma Runtime) cluster, provisioning a new one and executing standard set of tests on it.

## Usage

Prepare the `.env` file based on the `.env.template`. Run the following command to set up the environment variables in your system:

```bash
export $(xargs < .env)
```

Run the test scenario:

```bash
npm run nightly-skr
```

## Environment variables
`AL_SERVICE_KEY` must be a cloud foundry service key with the info about `UAA (User Account and Authentication)`. Learn more about [Managing Service Keys in Cloud Foundry](https://docs.cloudfoundry.org/devguide/services/service-keys.html).
