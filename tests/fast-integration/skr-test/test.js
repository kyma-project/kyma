const {
    GatherOptions,
} = require('./');
const {provisionSKR, deprovisionSKR} = require("../kyma-environment-broker");
const {keb, gardener, director} = require("./helpers");
const {initializeK8sClient} = require("../utils");
const {unregisterKymaFromCompass, addScenarioInCompass, assignRuntimeToScenario} = require("../compass");
const {OIDCE2ETest, CommerceMockTest} = require("./skr-test");
const uuid = require("uuid");
describe(`Execute SKR test`, function () {
    this.timeout(60 * 60 * 1000 * 3); // 3h
    this.slow(5000);
    const instanceID = uuid.v4();
    let options = GatherOptions();
    let skr;

    before('Provision SKR', async function () {
        const customParams = {
            oidc: options.oidc0,
        };
        skr = await provisionSKR(keb, gardener,
            instanceID,
            options.runtimeName,
            null,
            null,
            customParams);
        initializeK8sClient({ kubeconfig: skr.shoot.kubeconfig });
        await addScenarioInCompass(director, options.scenarioName);
        await assignRuntimeToScenario(director, skr.shoot.compassID, options.scenarioName);
    });
    describe('Execute tests', function () {
        OIDCE2ETest(skr, instanceID, options);
        CommerceMockTest(skr, options);
    });
    after(`Deprovision SKR`, async function () {
        await deprovisionSKR(keb, instanceID);
        await unregisterKymaFromCompass(director, options.scenarioName);
    });
});
