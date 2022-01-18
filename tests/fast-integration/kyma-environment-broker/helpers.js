const { wait, debug } = require("../utils");
const { expect } = require("chai");
const fs = require("fs");
const os = require("os");

async function provisionSKR(
  keb,
  kcp,
  gardener,
  instanceID,
  name,
  platformCreds,
  btpOperatorCreds,
  customParams,
  timeout
) {
  const resp = await keb.provisionSKR(
    name,
    instanceID,
    platformCreds,
    btpOperatorCreds,
    customParams
  );
  expect(resp).to.have.property("operation");

  const operationID = resp.operation;
  const shootName = resp.dashboard_url.split(".")[1];
  debug(`Operation ID ${operationID}`, `Shoot name ${shootName}`);

  await ensureOperationSucceeded(keb, kcp, instanceID, operationID, timeout);

  const shoot = await gardener.getShoot(shootName);
  debug(`Compass ID ${shoot.compassID}`);

  return {
    operationID,
    shoot,
  };
}

function ensureValidShootOIDCConfig(shoot, targetOIDCConfig) {
  expect(shoot).to.have.nested.property("oidcConfig.clientID", targetOIDCConfig.clientID);
  expect(shoot).to.have.nested.property("oidcConfig.issuerURL", targetOIDCConfig.issuerURL);
  expect(shoot).to.have.nested.property("oidcConfig.groupsClaim", targetOIDCConfig.groupsClaim);
  expect(shoot).to.have.nested.property("oidcConfig.usernameClaim", targetOIDCConfig.usernameClaim);
  expect(shoot).to.have.nested.property(
    "oidcConfig.usernamePrefix",
    targetOIDCConfig.usernamePrefix
  );
  expect(shoot.oidcConfig.signingAlgs).to.eql(targetOIDCConfig.signingAlgs);
}

async function deprovisionSKR(keb, kcp, instanceID, timeout, ensureSuccess=true) {
  const resp = await keb.deprovisionSKR(instanceID);
  expect(resp).to.have.property("operation");

  const operationID = resp.operation;
  debug(`Operation ID ${operationID}`);

  if (ensureSuccess) {
    await ensureOperationSucceeded(keb, kcp, instanceID, operationID, timeout);
  }

  return operationID;
}

async function updateSKR(keb, kcp, gardener, instanceID, shootName, customParams, timeout, btpOperatorCreds = null, isMigration = false) {
  const resp = await keb.updateSKR(instanceID, customParams, btpOperatorCreds, isMigration);
  expect(resp).to.have.property("operation");

  const operationID = resp.operation;
  debug(`Operation ID ${operationID}`);

  await ensureOperationSucceeded(keb, kcp, instanceID, operationID, timeout);

  const shoot = await gardener.getShoot(shootName);

  return {
    operationID,
    shoot,
  };
}

async function ensureOperationSucceeded(keb, kcp, instanceID, operationID, timeout) {
  const res = await wait(
    () => keb.getOperation(instanceID, operationID),
    (res) => res && res.state && (res.state === "succeeded" || res.state === "failed"),
    timeout,
    1000 * 30 // 30 seconds
  ).catch(async (err) => {
    let runtimeStatus = await kcp.getRuntimeStatusOperations(instanceID)
    throw(`${err}\nRuntime status: ${runtimeStatus}`)
  });

  debug("KEB operation:", res);
  if(res.state !== "succeeded") {
    let runtimeStatus = await kcp.getRuntimeStatusOperations(instanceID)
    throw(`operation didn't succeed in time: ${JSON.stringify(res, null, `\t`)}\nRuntime status: ${runtimeStatus}`);
  }
  return res;
}

async function getShootName(keb, instanceID) {
  const resp = await keb.getRuntime(instanceID);
  expect(resp.data).to.be.lengthOf(1);

  return resp.data[0].shootName;
}

async function ensureValidOIDCConfigInCustomerFacingKubeconfig(keb, instanceID, oidcConfig) {
  let kubeconfigContent;
  try {
    kubeconfigContent = await keb.downloadKubeconfig(instanceID);
  } catch (err) {}

  var issuerMatchPattern = "\\b" + oidcConfig.issuerURL + "\\b";
  var clientIDMatchPattern = "\\b" + oidcConfig.clientID + "\\b";
  expect(kubeconfigContent).to.match(new RegExp(issuerMatchPattern, "g"));
  expect(kubeconfigContent).to.match(new RegExp(clientIDMatchPattern, "g"));
}

async function saveKubeconfig(kubeconfig) {
  const directory = `${os.homedir()}/.kube`;
  if (!fs.existsSync(directory)) {
    fs.mkdirSync(directory, {recursive: true});
  }

  fs.writeFileSync(`${directory}/config`, kubeconfig);
}

module.exports = {
  provisionSKR,
  deprovisionSKR,
  saveKubeconfig,
  updateSKR,
  ensureOperationSucceeded,
  getShootName,
  ensureValidShootOIDCConfig,
  ensureValidOIDCConfigInCustomerFacingKubeconfig,
};
