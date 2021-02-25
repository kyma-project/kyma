const {
    KEBClient,
    KEBConfig,
    provisionSKR,
    ensureOperationSucceeded,
} = require("../keb");
const { debug, genRandom } = require("../utils");


describe("Provisioning SKR", function () {

    let config = new KEBConfig()
    var operationID;

    const kebClient = new KEBClient(config)
    it("Send provisioning call to KEB", async function(){
    operationID =  await provisionSKR(kebClient, instanceID, planID, name)
    });

    it("Wait for the SKR to provision", async function(){
        await ensureOperationSucceeded(kebClient, "kj-trial11",  "a9c046b7-f760-4dc9-9d7b-7586ab023b46")
    }).timeout(3600000);
});