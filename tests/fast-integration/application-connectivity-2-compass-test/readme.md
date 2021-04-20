# Kyma Application Connectivity 2.0 with Compass test

This test covers Kyma with Application Connectivity 2.0 running on k3s with Compass.

## Usage

Prepare the `.env` file based on the `.env.template`. Run the following command to set up the environment variables in your system:

```bash
export $(xargs < .env)
```

Run the test scenario:

```bash
npm run test-application-connectivity-2-compass-test
```