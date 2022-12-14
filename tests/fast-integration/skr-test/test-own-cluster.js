const {gatherOptions} = require('./index');
const {getOrProvisionSKR} = require('./provision/provision-skr');
const {GardenerClient, GardenerConfig} = require('../gardener');

const {getSecret} = require('../utils');

const {expect} = require('chai');

const {toBase64} = require('../utils');

const {genRandom} = require('../utils');

const provisioningTimeout = 1000 * 60 * 30; // 30m
const deprovisioningTimeout = 1000 * 60 * 1; // 1m
const globalTimeout = 1000 * 60 * 70; // 70m
const slowTime = 5000;

const gardener = new GardenerClient(GardenerConfig.fromEnv());

describe('SKR own_cluster test', function() {
  this.timeout(globalTimeout);
  this.slow(slowTime);

  let options = gatherOptions(); // with default values
  let skr;

  const shootName = genRandom(6);

  before('Shoot should be ready', async function() {
    this.timeout(provisioningTimeout);
    console.log(`Creating a shoot ${shootName}`);
    await gardener.createShoot(shootName);
  });

  before('Ensure SKR is provisioned', async function() {
    console.log(`Waiting for a shoot ${shootName}`);
    const shoot = await gardener.getShoot(shootName);

    this.timeout(provisioningTimeout);
    options.customParams.kubeconfig = toBase64(shoot.kubeconfig);
    options.customParams.shootDomain = shoot.shootDomain;
    options.customParams.shootName = shootName;

    console.log(`Waiting for provisioned SKR`);
    skr = await getOrProvisionSKR(options,
        false, provisioningTimeout);
    options = skr.options;
  });

  // Check if the secret with BTP credentials are created properly
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

  after('Cleanup the resources', async function() {
    console.log('Cleaning up');
    this.timeout(deprovisioningTimeout);
    await gardener.deleteShoot(shootName);
  });
});
