# SKR test

This test covers SKR (SAP Kyma Runtime).

## File structure
- **provision** folder contains the scripts for provisioning and de-provisioning the SKR cluster using KEB client
- **oidc** folder contains the OIDC-related tests

## Usage modes

You can use the SKR test in two modes - with and without provisioning.

### With provisioning

In this mode, the test executes the following steps:

1. Provision SKR cluster
2. Run OIDC Test
4. De-provision the SKR instance and clean the resources.

### Without Provisioning.

In this mode the test additionally needs the following environment variables:
- `SKIP_PROVISIONING`, set to `true`
- `INSTANCE_ID` the uuid of the provisioned SKR instance

In this mode, the test executes the following steps:
1. Ensure SKR exists
2. Run OIDC Test
4. Clean the resources
 
**NOTE:** The SKR test additionally contains a stand-alone script, which you can use to register the resources.

## Test execution

1. Before you run the test, prepare the `.env` file based on the following `.env.template`:
```
INSTANCE_ID
SKIP_PROVISIONING
KEB_HOST=
KEB_CLIENT_ID=
KEB_CLIENT_SECRET=
KEB_GLOBALACCOUNT_ID=
KEB_SUBACCOUNT_ID=
KEB_USER_ID=
KEB_PLAN_ID=
GARDENER_KUBECONFIG=

KCP_KEB_API_URL=
KCP_GARDENER_NAMESPACE=
KCP_OIDC_ISSUER_URL=
KCP_OIDC_CLIENT_ID=
KCP_OIDC_CLIENT_SECRET=
KCP_TECH_USER_LOGIN=
KCP_TECH_USER_PASSWORD=
KCP_MOTHERSHIP_API_URL=
KCP_KUBECONFIG_API_URL=

BTP_OPERATOR_CLIENTID=
BTP_OPERATOR_CLIENTSECRET=
BTP_OPERATOR_URL=
BTP_OPERATOR_TOKENURL=

AL_SERVICE_KEY= #must be a cloud foundry service key with the info about `UAA (User Account and Authentication)`. Learn more about [Managing Service Keys in Cloud Foundry](https://docs.cloudfoundry.org/devguide/services/service-keys.html).
```

2. To set up the environment variables in your system, run:

```bash
export $(xargs < .env)
```

3. Choose whether you want to run the test with or without provisioning.
   - To run the test **with** provisioning, call the following target:

    ```bash
    npm run test-skr
    #or
    make test-skr
    ```
    - To run the SKR test **without** provisioning, use the following command:

    ```bash
    make test-skr SKIP_PROVISIONING=true
    #or
    npm run test-skr # when all env vars are exported
    ```

