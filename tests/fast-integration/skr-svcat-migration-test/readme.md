# SKR SVCAT migration test

This test covers SKR (SAP Kyma Runtime) SVCAT migration.

## Usage

Prepare the `.env` file based on the `.env.template`. Run the following command to set up the environment variables in your system:

```bash
export $(xargs < .env)
```

Then, run the test scenario:

```bash
npm run test-skr-svcat-migration
```
