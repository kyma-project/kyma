const {
    GatherOptions,
} = require('./');
const {provisionSKR, deprovisionSKR} = require('../kyma-environment-broker');
const {keb, gardener, director} = require('./helpers');
const {initializeK8sClient} = require('../utils');
const {unregisterKymaFromCompass, addScenarioInCompass, assignRuntimeToScenario} = require('../compass');
const {OIDCE2ETest, CommerceMockTest} = require('./skr-test');

describe(`Execute SKR test`, function () {
    this.timeout(60 * 60 * 1000 * 3); // 3h
    this.slow(5000);
    before('Provision SKR', async function () {
        try {
            this.options = GatherOptions();
            console.log(`Provision SKR with instance ID ${this.options.instanceID}`);
            const customParams = {
                oidc: this.options.oidc0,
            };
            let skr = await provisionSKR(keb, gardener,
                this.options.instanceID,
                this.options.runtimeName,
                null,
                null,
                customParams);
            this.shoot = skr.shoot;
            await addScenarioInCompass(director, this.options.scenarioName);
            await assignRuntimeToScenario(director, this.shoot.compassID, this.options.scenarioName);
            initializeK8sClient({kubeconfig: this.shoot.kubeconfig});
        } catch (e) {
            throw new Error(`before hook failed: ${e.toString()}`);
        }
    });
    OIDCE2ETest();
    CommerceMockTest();
    after(`Deprovision SKR`, async function () {
        await deprovisionSKR(keb, this.options.instanceID);
        await unregisterKymaFromCompass(director, this.options.scenarioName);
    });
});
