# SKR test

This test covers SKR (SAP Kyma Runtime).

## Usage

Prepare the `.env` file based on the `.env.template`. Run the following command to set up the environment variables in your system:

```bash
export $(xargs < .env)
```

Run the test scenario:

```bash
npm run test-skr
```

## Environment variables
`AL_SERVICE_KEY` needs to be a cloud foundry service key. With the info about `UAA (User Account and Authentication)`. For more info in service key refer [here](https://docs.cloudfoundry.org/devguide/services/service-keys.html)
