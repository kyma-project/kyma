const { retryPromise, debug, expectNoAxiosErr } = require("../utils");

const { expect } = require("chai");

class KEBConfig {
  constructor() {
    this.kebHost = process.env["KEB_HOST"] || "";
    this.clientID = process.env["KEB_CLIENT_ID"] || "";
    this.clientSecret = process.env["KEB_CLIENT_SECRET"] || "";
    this.globalAccountID = process.env["KEB_GLOBALACCOUNT_ID"] || "";
    this.subAccountID = process.env["KEB_SUBACCOUNT_ID"] || "";
    this.planID = process.env["KEB_PLAN_ID"] || "";
    this.name = process.env["KEB_SKR_NAME"] || "";
    this.instanceID = process.env["KEB_INSTANCE_ID"] || "";
  }
}

async function provisionSKR(keb, instanceID, planID, name) {
  const response = await keb.provisionSKR(planID, name, instanceID);
  expect(response).to.have.property("operation");

  return response.operation;
}

async function deprovisionSKR(keb, instanceID, planID) {
  const response = await keb.deprovisionSKR(instanceID, planID);
  expect(response).to.have.property("operation");

  return response.operation;
}

async function ensureOperationSucceeded(keb, instanceID, operationID) {
  await retryPromise(
    async () => {
      let res = await keb.getOperation(instanceID, operationID);
      debug(res);
      expect(res).to.have.property("state", "succeeded");
    },
    60,
    36000000
  ).catch(expectNoAxiosErr);
}

async function getProvisionerID(keb, instanceID){
  const response = await keb.getRuntime(instanceID)
  expect(response.data).to.be.lengthOf(1)

  return response.data[0].runtimeID
}

module.exports = {
  KEBConfig,
  provisionSKR,
  deprovisionSKR,
  ensureOperationSucceeded,
  getProvisionerID
};
