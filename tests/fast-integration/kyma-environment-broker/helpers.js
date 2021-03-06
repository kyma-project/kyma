const fs = require("fs");
const { 
  wait,
  debug,
  getEnvOrThrow,
} = require("../utils");
const { 
  expect 
} = require("chai");

const SHOOT_KUBECONFIG_PATH = getEnvOrThrow("SHOOT_KUBECONFIG");

async function provisionSKR(keb, gardener, instanceID, name) {
  const resp = await keb.provisionSKR(name, instanceID);
  expect(resp).to.have.property("operation");

  const operationID = resp.operation
  const shootName = resp.dashboard_url.split(".")[1];
  debug(`Operation ID ${operationID}`, `Shoot name ${shootName}`);

  await ensureOperationSucceeded(keb, instanceID, operationID);

  const shoot = await gardener.getShoot(shootName);
  debug(`Compass ID ${shoot.compassID}`);

  fs.writeFileSync(SHOOT_KUBECONFIG_PATH, shoot.kubeconfig);
  
  return {
    operationID, 
    shoot,
  };
}

async function deprovisionSKR(keb, instanceID) {
  const resp = await keb.deprovisionSKR(instanceID);
  expect(resp).to.have.property("operation");

  const operationID = resp.operation;
  debug(`Operation ID ${operationID}`);

  await ensureOperationSucceeded(keb, instanceID, operationID);

  return operationID;
}

async function ensureOperationSucceeded(keb, instanceID, operationID) {
  const res = await wait(
    () => keb.getOperation(instanceID, operationID),
    (res) => res && res.state && (res.state === "succeeded" || res.state === "failed"),
    1000 * 60 * 60 * 2, // 2h
    1000 * 30 // 30 seconds
  );
  
  debug("KEB operation:", res);
  expect(res).to.have.property("state", "succeeded");

  return res;
}

async function getShootName(keb, instanceID){
  const resp = await keb.getRuntime(instanceID)
  expect(resp.data).to.be.lengthOf(1)

  return resp.data[0].shootName
}

module.exports = {
  SHOOT_KUBECONFIG_PATH,
  provisionSKR,
  deprovisionSKR,
  ensureOperationSucceeded,
  getShootName
};
