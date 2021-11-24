const {
    gardener, GatherOptions, director,
} = require('./helpers')
const {getEnvOrDefault, initializeK8sClient} = require("../utils");
const {CommerceMockTest} = require("./skr-test");
const {addScenarioInCompass, assignRuntimeToScenario} = require("../compass");

const shootName = getEnvOrDefault("SHOOT","something");

describe(`Execute tests on existing SKR`, function () {
    this.timeout(60 * 60 * 1000 * 3); // 3h
    this.slow(5000);
    before(async function () {
        this.options = GatherOptions();
        try {
            this.shoot = await gardener.getShoot(shootName);
            await addScenarioInCompass(director, this.options.scenarioName);
            await assignRuntimeToScenario(director, this.shoot.compassID, this.shoot.scenarioName);
            initializeK8sClient({kubeconfig: this.shoot.kubeconfig});
        } catch (e) {
            throw e
        }

    })
    CommerceMockTest();
})
