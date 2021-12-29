const {} = require("../../compass");
const {
    getEnvOrThrow,
} = require("../../utils");
const {
    KCPConfig,
    KCPWrapper,
} = require("../../kcp/client")
const {debug} = require("../../utils");
const {inspect} = require("util");
const {
    skrInstanceId
} = require("./utils");

// Mocha root hook
process.env.KCP_KEB_API_URL = `https://kyma-env-broker.` + keb.host;
process.env.KCP_GARDENER_NAMESPACE = `garden-kyma-dev`;
process.env.KCP_OIDC_ISSUER_URL = `https://kymatest.accounts400.ondemand.com`;
process.env.KCP_MOTHERSHIP_API_URL = 'https://mothership-reconciler.cp.dev.kyma.cloud.sap/v1';
process.env.KCP_KUBECONFIG_API_URL = 'https://kubeconfig-service.cp.dev.kyma.cloud.sap';
const kcp = new KCPWrapper(KCPConfig.fromEnv());
const kymaUpgradeVersion = getEnvOrThrow(process.env.KYMA_SOURCE)

describe(`Upgrade the skr cluster with instanceID ${skrInstanceId} to kyma version ${kymaUpgradeVersion}`, function () {
    it(`Perform Upgrade`, async function () {
        let kcpUpgradeStatus = await kcp.upgradeKyma(skrInstanceId, kymaUpgradeVersion)
        debug("Upgrade Done!")
    });

    it(`Get Runtime Status`, async function () {
        let runtimeStatus = await kcp.runtimes({instanceID: skrInstanceId})
        debug(inspect(runtimeStatus, false, null, false))
    });
});
