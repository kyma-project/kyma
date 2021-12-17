const uuid = require("uuid");
const {
    provisionSKR,
    saveKubeconfig,
} = require("../../kyma-environment-broker");

const {
    KCPConfig,
    KCPWrapper,
} = require("../../kcp/client")

const {
    OIDCE2ETest,
    GatherOptions,
    gardener,
    keb,
} = require("../../skr-test");

const {
    getEnvOrThrow,
    debug
} = require("../../utils");

// Mocha root hook
process.env.KCP_KEB_API_URL = `https://kyma-env-broker.` + keb.host;
process.env.KCP_GARDENER_NAMESPACE = `garden-kyma-dev`;
process.env.KCP_OIDC_ISSUER_URL = `https://kymatest.accounts400.ondemand.com`;
process.env.KCP_MOTHERSHIP_API_URL = 'https://mothership-reconciler.cp.dev.kyma.cloud.sap/v1';
process.env.KCP_KUBECONFIG_API_URL = 'https://kubeconfig-service.cp.dev.kyma.cloud.sap';

const kcp = new KCPWrapper(KCPConfig.fromEnv());
const instanceId = process.env.INSTANCE_ID || uuid.v4()
const runtimeName = getEnvOrThrow("RUNTIME_NAME")
const kymaVersion = getEnvOrThrow("KYMA_VERSION")
const kymaOverridesVersion = getEnvOrThrow("KYMA_OVERRIDES_VERSION")
const kymaProfile = process.env.EXECUTION_PROFILE || "evaluation"

describe("Provision SKR cluster", function () {
    this.timeout(3600000 * 2); // 2h
    this.slow(5000);
    before(`Provision new SKR`, async function () {
        try {
            this.options = GatherOptions();

            console.log('Login to KCP...');
            let version = await kcp.version([])
            debug(version)
            await kcp.login();

            console.log(`Provision SKR with:\n runtime ID: ${instanceId},\n runtime name: ${runtimeName},\n kyma version: ${kymaVersion},\n overrides version: ${kymaOverridesVersion}\n`);
            const customParams = {
                oidc: this.options.oidc0,
                "kymaVersion": kymaVersion,
                "overridesVersion": kymaOverridesVersion,
                "purpose": kymaProfile,
            };

            let skr = await provisionSKR(keb, gardener, instanceId, runtimeName, null, null, customParams);
            this.shoot = skr.shoot;

            console.log(`Saving kubeconfig to ~/.kube/config`);
            await saveKubeconfig(skr.shoot.kubeconfig);
        }
        catch (e) {
            throw new Error(`SKR provisioning failed: ${e.toString()}`);
        }
    });

    // Run OIDC e2e tests
    OIDCE2ETest();
});
