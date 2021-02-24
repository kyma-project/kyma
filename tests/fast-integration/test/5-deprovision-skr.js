const {
    KEBClient,
    deprovisionSKR,
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
    it("Send deprovisioning call to KEB", async function(){
    operationID =  await deprovisionSKR(kebClient, instanceID, planID, name)
    });

    it("Wait for the SKR to deprovision", async function(){
        await ensureOperationSucceeded(kebClient, instanceID, operationID)
    }).timeout(3600000);;
});