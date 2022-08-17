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
  let options;
  let shootInfo;

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
    if (shootInfo != undefined) {
      console.log(options);

      // print data out for spinnaker
      console.log('***Print out data for spinnaker***');
      console.log(`SPINNAKER_PROPERTY_PROVISIONED=true`);
      console.log(`SPINNAKER_PROPERTY_INSTANCE_ID=${options.instanceID}`);
    }
  });
});
