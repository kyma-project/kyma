const uuid = require('uuid');
const {
    provisionSKR,
    saveKubeconfig,
} = require('../../kyma-environment-broker');

const {
    KCPConfig,
    KCPWrapper,
} = require('../../kcp/client');

const {
    gardener,
    keb,
} = require('../../skr-test');

const {
    getEnvOrThrow,
    debug
} = require('../../utils');

const instanceId = process.env.INSTANCE_ID || uuid.v4()
const runtimeName = getEnvOrThrow('RUNTIME_NAME');
const kymaVersion = getEnvOrThrow('KYMA_VERSION');
const kymaOverridesVersion = process.env.KYMA_OVERRIDES_VERSION || ""
const kcp = new KCPWrapper(KCPConfig.fromEnv());

let skr;

describe('Provision SKR cluster', function () {
    this.timeout(60 * 60 * 1000 * 2); // 2h
    this.slow(5000);
    before('Provision new SKR', async function () {
        // login to kcp, required by provisionSKR method
        const version = await kcp.version([]);
        debug('Login to KCP. Version: ', version)
        await kcp.login();

        // define params required by provisionSKR method
        const provisioningTimeout = 1000 * 60 * 60 // 1h

        const customParams = { "kymaVersion": kymaVersion };
        if (kymaOverridesVersion) {
            customParams["overridesVersion"] = kymaOverridesVersion;
        }

        debug(`Parameters:\n`,
            `runtime ID: ${instanceId}\n`,
            `runtime name: ${runtimeName}\n`,
            `kyma version: ${kymaVersion}\n`,
            `custom params: ${JSON.stringify(customParams)}\n`,
        );

        // Finally, call the provision SKR method with the provided params
        skr = await provisionSKR(
            keb,
            kcp,
            gardener,
            instanceId,
            runtimeName,
            null,
            null,
            customParams,
            provisioningTimeout
        );
    });

    describe('Check provisioned SKR', function () {
        it('Should get Runtime Status after provisioning', async function () {
            let runtimeStatus = await kcp.getRuntimeStatusOperations(instanceId);
            debug(`\nRuntime status: ${runtimeStatus}`)
        });
        it(`Should save kubeconfig for the SKR to ~/.kube/config`, async function() {
            await saveKubeconfig(skr.shoot.kubeconfig);
        });
    })
});
