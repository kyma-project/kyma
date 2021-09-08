# SKR test for Preview-Plan

This test covers SKR (SAP Kyma Runtime).

## Usage

Prepare the `.env` file based on the `.env.template`. Run the following command to set up the environment variables in your system:

```bash
export $(xargs < .env)
```

Run the test scenario:

```bash
npm run ci-kyma-preview
```

## Environment variables
`KEB_PLAN_ID` must be a the Preview-PlanID.
