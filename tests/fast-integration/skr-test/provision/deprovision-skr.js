const {deprovisionSKR} = require('../../kyma-environment-broker');
const {keb, kcp} = require('../helpers');


async function deprovisionSKRInstance(options, timeout, ensureSuccess=true) {
  try {
    await deprovisionSKR(keb, kcp, options.instanceID, timeout, ensureSuccess);
  } catch (e) {
    throw new Error(`De-provisioning failed: ${e.toString()}`);
  } finally {
    const runtimeStatus = await kcp.getRuntimeStatusOperations(options.instanceID);
    console.log(`\nRuntime status after de-provisioning: ${runtimeStatus}`);
    await kcp.reconcileInformationLog(runtimeStatus);
  }
}

module.exports = {
  deprovisionSKRInstance,
};
