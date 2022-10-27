const {
  gatherOptions,
  oidcE2ETest,
} = require('./index');
const {getOrProvisionSKR} = require('./provision/provision-skr');
const {deprovisionAndUnregisterSKR} = require('./provision/deprovision-skr');
const {getSecret} = require('../utils');
const {expect} = require('chai');

const skipProvisioning = process.env.SKIP_PROVISIONING === 'true';
const provisioningTimeout = 1000 * 60 * 30; // 30m
const deprovisioningTimeout = 1000 * 60 * 95; // 95m
let globalTimeout = 1000 * 60 * 70; // 70m
const slowTime = 5000;

describe('SKR test', function() {
  if (!skipProvisioning) {
    globalTimeout += provisioningTimeout + deprovisioningTimeout;
  }
  this.timeout(globalTimeout);
  this.slow(slowTime);

  let options = gatherOptions(); // with default values
  let skr;
  const getShootInfoFunc = function() {
    return skr.shoot;
  };
  const getShootOptionsFunc = function() {
    return options;
  };

  before('Ensure SKR is provisioned', async function() {
    this.timeout(provisioningTimeout);
    skr = await getOrProvisionSKR(options, skipProvisioning, provisioningTimeout);
    options = skr.options;
  });

  // Check if the secret with BTP credentials is created properly
  it('should check if secret exists and contains the expected data', async function() {
    console.log(`Checking the contents of the "sap-btp-manager" secret`);
    const secretToCheck = await getSecret('sap-btp-manager', 'kyma-system');
    expect(secretToCheck).to.not.be.empty;
    expect(secretToCheck.metadata.labels['app.kubernetes.io/managed-by']).to.equal('kcp-kyma-environment-broker');
    expect(secretToCheck.data).to.have.property('clientid');
    expect(secretToCheck.data).to.have.property('clientsecret');
    expect(secretToCheck.data).to.have.property('sm_url');
    expect(secretToCheck.data).to.have.property('tokenurl');
    expect(secretToCheck.data).to.have.property('cluster_id');
  });

  // Run the OIDC tests
  oidcE2ETest(getShootOptionsFunc, getShootInfoFunc);

  after('Cleanup the resources', async function() {
    this.timeout(deprovisioningTimeout);
    await deprovisionAndUnregisterSKR(options, deprovisioningTimeout, skipProvisioning, false);
  });
});
