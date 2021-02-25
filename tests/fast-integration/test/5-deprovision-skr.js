const {
    KEBClient,
    KEBConfig,
    deprovisionSKR,
    ensureOperationSucceeded,
} = require("../keb");
const { debug, genRandom } = require("../utils");


describe("Deprovisioning SKR", function () {

    let config = new KEBConfig()
    let operationID;

    const kebClient = new KEBClient(config)
    it("Send deprovisioning call to KEB", async function(){
    operationID =  await deprovisionSKR(kebClient, config.instanceID, config.planID)
    });

    it("Wait for the SKR to deprovision", async function(){
        await ensureOperationSucceeded(kebClient, config.instanceID, operationID)
    }).timeout(3600000);;
});