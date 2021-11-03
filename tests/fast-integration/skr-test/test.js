const {
    skrTest,
    GatherOptions,
} = require('./');
const {provisionSKR, deprovisionSKR} = require("../kyma-environment-broker");
const {keb, gardener, director} = require("./helpers");
const {initializeK8sClient} = require("../utils");
const {unregisterKymaFromCompass} = require("../compass");

describe(`Execute SKR test`, function () {
    this.timeout(60 * 60 * 1000 * 3); // 3h
    this.slow(5000);
    let options = GatherOptions();
    let skr;

    describe('Provision SKR', function () {
        it(`Provision SKR with ID ${options.runtimeID}`, async function () {
            const customParams = {
                oidc: options.oidc0,
            };
            skr = await provisionSKR(keb, gardener,
                options.runtimeID,
                options.runtimeName,
                null,
                null,
                customParams);
            initializeK8sClient({ kubeconfig: skr.shoot.kubeconfig });
        });
    });
    skrTest(skr, options);
    describe(`Deprovision SKR`, function () {
        it("Deprovision SKR", async function () {
            await deprovisionSKR(keb, options.runtimeID);
        });

        it("Unregister SKR resources from Compass", async function () {
            await unregisterKymaFromCompass(director, options.scenarioName);
        });
    });
});
