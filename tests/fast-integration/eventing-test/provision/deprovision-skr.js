const {expect} = require('chai');
const {
  deprovisionSKR,
} = require('../../kyma-environment-broker');

const {
  keb,
} = require('../../skr-test');

const {
  getEnvOrThrow,
  debug,
} = require('../../utils');
const {KCPWrapper, KCPConfig} = require('../../kcp/client');
const {
  cleanupTestingResources,
  slowTime,
} = require('../utils');

const instanceId = getEnvOrThrow('INSTANCE_ID');

describe('Clean the testing resources and de-provision SKR cluster', function() {
  this.timeout(60 * 60 * 1000); // 1h
  this.slow(slowTime);
  const kcp = new KCPWrapper(KCPConfig.fromEnv());

  it('Compass scenario and the test namespaces should be deleted', async function() {
    await cleanupTestingResources();
  });

  it('Should trigger KEB to de-provision SKR', async function() {
    debug(`De-provision SKR with runtime ID: ${instanceId}`);
    const operationID = await deprovisionSKR(keb, kcp, instanceId, null, false);

    expect(operationID).to.not.be.empty;
  });
});
