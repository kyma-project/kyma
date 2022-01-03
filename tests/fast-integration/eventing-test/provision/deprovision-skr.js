const {expect} = require("chai");
const {
    deprovisionSKR,
} = require("../../kyma-environment-broker");

const {
    keb,
} = require("../../skr-test");

const {
    getEnvOrThrow,
    debug,
} = require("../../utils");

const instanceId = getEnvOrThrow("INSTANCE_ID")

describe("De-provision SKR cluster", function () {
    this.timeout(60 * 60 * 1000 * 1); // 1h
    this.slow(5000);

    it(`Should trigger KEB to de-provision SKR`, async function () {
        debug(`De-provision SKR with runtime ID: ${instanceId}`);
        const operationID = await deprovisionSKR(keb, null, instanceId, null, false);

        expect(operationID).to.not.be.empty;
    })
});
