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
