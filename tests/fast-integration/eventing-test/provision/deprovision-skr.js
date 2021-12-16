const {
    deprovisionSKR,
} = require("../../kyma-environment-broker");

const {
    keb,
} = require("../../skr-test");

const {
    getEnvOrThrow,
} = require("../../utils");

const instanceId = getEnvOrThrow("INSTANCE_ID")

describe("De-provision SKR cluster", function () {
    this.timeout(3600000 * 1); // 1h
    this.slow(5000);

    it(`De-provision SKR`, async function () {
        console.log(`De-provision SKR with runtime ID ${instanceId}`);
        await deprovisionSKR(keb, instanceId);
    })
});
