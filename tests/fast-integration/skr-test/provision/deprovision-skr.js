const {deprovisionSKR} = require('../../kyma-environment-broker');
const {keb, kcp} = require('../provision/provision-skr');


async function deprovisionSKRInstance(options, timeout) {
  try {
    await deprovisionSKR(keb, kcp, options.instanceID, timeout);
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
