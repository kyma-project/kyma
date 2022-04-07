# SKR test

This test covers SKR (SAP Kyma Runtime).

## File structure
- **provision** folder contains the scripts for provisioning and de-provisioning the SKR cluster using KEB client
- **oidc** folder contains the OIDC-related tests
- **commerce-mock** folder consists of two scripts:
  - **index.js** contains the commerce mock tests suite
  - **prep.js** can be called in stand-alone mode for Commerce Mock resources preparation

## Usage modes

You can use the SKR test in two modes - with and without provisioning.

### With provisioning

In this mode, the test executes the following steps:

1. Provision SKR cluster and register compass resources
2. Run OIDC Test
3. Run Commerce Mock Test
4. De-provision the SKR instance and clean the compass resources.

### Without Provisioning.

In this mode the test additionally needs the following environment variables:
- `SKIP_PROVISIONING`, set to `true`
- `INSTANCE_ID` the uuid of the provisioned SKR instance
- `TEST_SUFFIX` determines the compass scenario and the app assigned to it. If not set, it's randomly generated.

>**IMPORTANT:** The `TEST_SUFFIX must be 4 characters long and is required for every commerce-mock test run. If the compass scenario with the given suffix already exists, the test will try to reuse it.

In this mode, the test executes the following steps:
1. Ensure SKR exists and register compass resources
2. Run OIDC Test
3. Run Commerce Mock Test
4. Clean the compass resources
 
**NOTE:** The SKR test additionally contains a stand-alone script, which you can use to register the Commerce Mock resources.

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

COMPASS_HOST=
COMPASS_CLIENT_ID=
COMPASS_CLIENT_SECRET=
COMPASS_TENANT=

KCP_KEB_API_URL=
KCP_GARDENER_NAMESPACE=
KCP_OIDC_ISSUER_URL=
KCP_OIDC_CLIENT_ID=
KCP_OIDC_CLIENT_SECRET=
KCP_TECH_USER_LOGIN=
KCP_TECH_USER_PASSWORD=
KCP_MOTHERSHIP_API_URL=
KCP_KUBECONFIG_API_URL=

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

### Commerce Mock resources preparation

If you want to prepare the Commerce Mock resources

1. Set all the environment variables as explained for the mode without provisioning.
2. Run the following command:
    ```bash
    npm run test-commercemock-prepare
    ```

