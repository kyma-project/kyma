const {
    gardener, GatherOptions,
} = require('./helpers')
const {getEnvOrDefault, initializeK8sClient} = require("../utils");
const {CommerceMockTest} = require("./skr-test");

const shootName = getEnvOrDefault("SHOOT","something");

describe(`Execute tests on existing SKR`, function () {
    this.timeout(60 * 60 * 1000 * 3); // 3h
    this.slow(5000);
    before(async function () {
        this.options = GatherOptions();
        this.shoot = await gardener.getShoot(shootName);
        initializeK8sClient({kubeconfig: this.shoot.kubeconfig});
    })
    CommerceMockTest();
})
