const { retryPromise, debug } = require("../utils");
const { genRandom,
    expectNoAxiosErr } = require("../utils");

    const { expect } = require("chai");

async function provisionSKR(keb,instanceID, planID, name) {

  const response = await keb.provisionSKR(planID,name,instanceID)
  expect(response).to.have.property("operation")

  return response.operation
  
}

async function deprovisionSKR(keb,instanceID) {

  const response = await keb.deprovisionSKR(instanceID)
  expect(response).to.have.property("operation")

  return response.operation
  
}

async function ensureOperationSucceeded(keb, instanceID, operationID ) {
  await retryPromise(
   async () =>{
     let res = await keb.getSKRState(instanceID, operationID);
     debug(res)
     expect(res).to.have.property("state", "succeeded");
  },
    60,
    36000000
  ).catch(expectNoAxiosErr);
}

module.exports = {
  provisionSKR,
  deprovisionSKR,
  ensureOperationSucceeded
};
