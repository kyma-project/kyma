const {
    KEBClient,
    provisionSKR,
    ensureOperationSucceeded,
} = require("../keb");
const { debug, genRandom } = require("../utils");


describe("Provisioning SKR", function () {

    const kebHost = process.env["KEB_HOST"] || "";
    const clientID = process.env["KEB_CLIENT_ID"] || "";
    const clientSecret = process.env["KEB_CLIENT_SECRET"] || "";
    const globalAccountID = process.env["KEB_GLOBALACCOUNT_ID"] || "";
    const subAccountID = process.env["KEB_SUBACCOUNT_ID"] || "";
    const planID = process.env["KEB_PLAN_ID"] || "";
    const name = process.env["KEB_SKR_NAME"] || "";
    const instanceID = process.env["KEB_INSTANCE_ID"] || "";
    var operationID;

    const kebClient = new KEBClient(kebHost,clientID,clientSecret,globalAccountID,subAccountID)
    it("Send provisioning call to KEB", async function(){
    operationID =  await provisionSKR(kebClient, instanceID, planID, name)
    });

    it("Wait for the SKR to provision", async function(){
        await ensureOperationSucceeded(kebClient, "kj-trial11",  "a9c046b7-f760-4dc9-9d7b-7586ab023b46")
    }).timeout(3600000);
});