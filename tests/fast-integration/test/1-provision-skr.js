const {
    KEBClient,
    KEBConfig,
    provisionSKR,
    ensureOperationSucceeded,
} = require("../keb");
const { debug, genRandom } = require("../utils");


describe("Provisioning SKR", function () {

    let config = new KEBConfig()
    let operationID;

    const kebClient = new KEBClient(config)
    it("Send provisioning call to KEB", async function(){
    operationID =  await provisionSKR(kebClient, config.instanceID, config.planID, config.name)
    });

    it("Wait for the SKR to provision", async function(){
        await ensureOperationSucceeded(kebClient, config.instanceID,  operationID)
    }).timeout(3600000);
});