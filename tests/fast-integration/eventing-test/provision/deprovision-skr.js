const {expect} = require('chai');
const {
  deprovisionSKR,
  KEBClient,
  KEBConfig,
} = require('../kyma-environment-broker');

const {
  getEnvOrThrow,
  debug,
} = require('../../utils');
const keb = new KEBClient(KEBConfig.fromEnv());
const {KCPWrapper, KCPConfig} = require('../kcp/client');
const {slowTime} = require('../utils');

const instanceId = getEnvOrThrow('INSTANCE_ID');

describe('Clean the testing resources and de-provision SKR cluster', function() {
  this.timeout(60 * 60 * 1000); // 1h
  this.slow(slowTime);
  const kcp = new KCPWrapper(KCPConfig.fromEnv());

  it('Should trigger KEB to de-provision SKR', async function() {
    debug(`De-provision SKR with runtime ID: ${instanceId}`);
    const operationID = await deprovisionSKR(keb, kcp, instanceId, null, false);

    expect(operationID).to.not.be.empty;
  });
});
