const {
  gatherOptions,
} = require('./index');
const {getOrProvisionSKR} = require('./provision/provision-skr');

const provisioningTimeout = 1000 * 60 * 30; // 30m
let globalTimeout = 1000 * 60 * 70; // 70m
const slowTime = 5000;

describe('Provision SKR instance', function() {
  globalTimeout += provisioningTimeout;

  this.timeout(globalTimeout);
  this.slow(slowTime);

  let skr;
  let options = undefined;
  let shootInfo = undefined;

  before('Gather default options', async function() {
    options = gatherOptions(); // with default values
  });

  it('should provision SKR cluster', async function() {
    this.timeout(provisioningTimeout);
    console.log(`SKR Instance ID: ${options.instanceID}`);
    skr = await getOrProvisionSKR(options, false, provisioningTimeout);
    options = skr.options;
    shootInfo = skr.shoot;
  });

  after('Print Shoot Info', async function() {
    // Print data out for spinnaker.
    // It is used in spinnaker to pass data to next stages.
    // More info: https://spinnaker.io/docs/guides/user/kubernetes-v2/run-job-manifest/#spinnaker_property_

    if (options && options.instanceID) {
      console.log(`SPINNAKER_PROPERTY_INSTANCE_ID=${options.instanceID}`);
    }

    if (shootInfo) {
      console.log(`SPINNAKER_PROPERTY_PROVISIONED=true`);
    }
  });
});
