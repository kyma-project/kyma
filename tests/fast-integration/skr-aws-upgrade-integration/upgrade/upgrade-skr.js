const {kcp} = require('../../!skr-test/helpers');

async function upgradeSKRInstance(options, kymaUpgradeVersion, timeout) {
  try {
    await kcp.upgradeKyma(options.instanceID, kymaUpgradeVersion, timeout);
    console.log('Upgrade Done!');
  } catch (e) {
    throw new Error(`Upgrade failed: ${e.toString()}`);
  } finally {
    const runtimeStatus = await kcp.getRuntimeStatusOperations(options.instanceID);
    console.log(`\nRuntime status after upgrade: ${runtimeStatus}`);
    await kcp.reconcileInformationLog(runtimeStatus);
  }
}


module.exports = {
  upgradeSKRInstance,
};

