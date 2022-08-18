const {
  gatherOptions,
} = require('./index');
const {deprovisionAndUnregisterSKR} = require('./provision/deprovision-skr');
const {getEnvOrThrow} = require('./helpers');

const deprovisioningTimeout = 1000 * 60 * 95; // 95m
let globalTimeout = 1000 * 60 * 70; // 70m
const slowTime = 5000;

describe('De-provision SKR instance', function() {
  globalTimeout += deprovisioningTimeout;

  this.timeout(globalTimeout);
  this.slow(slowTime);

  let options;
  const instanceID = getEnvOrThrow('INSTANCE_ID');

  before('Gather options', async function() {
    options = gatherOptions(); // with default values
    options.instanceID = instanceID;
  });

  it('should de-provision SKR cluster', async function() {
    this.timeout(deprovisioningTimeout);
    console.log(`SKR Instance ID: ${options.instanceID}`);
    await deprovisionAndUnregisterSKR(options, deprovisioningTimeout, false, false);
  });
});
