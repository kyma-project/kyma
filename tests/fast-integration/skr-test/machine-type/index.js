const {expect} = require('chai');
const {updateSKR} = require('../../kyma-environment-broker');
const {keb, kcp} = require('../helpers');
const {GardenerClient, GardenerConfig} = require('../../gardener');
const {getEnvOrThrow} = require('../../utils');

const updateTimeout = 1000 * 60 * 30; // 30m

function machineTypeE2ETest(getShootOptionsFunc, getShootInfoFunc) {
  describe('Machine type update test', function() {
    let shoot = undefined;
    let options = undefined;
    const gardener = new GardenerClient(GardenerConfig.fromEnv());
    const updateMachineType = getEnvOrThrow('MACHINE_TYPE_UPDATE');

    before('Get provisioned Shoot Info', async function() {
      shoot = getShootInfoFunc();
      options = getShootOptionsFunc();
    });

    it('Should check default machine type after provisioning', async function() {
      const machineType = await getMachineType(gardener, shoot.name);
      console.log(`Default machine type ${machineType}`);
    });

    it('Should update SKR service instance with machine type', async function() {
      this.timeout(updateTimeout);
      const customParams = {
        machineType: updateMachineType,
      };
      const skr = await updateSKR(keb,
          kcp,
          gardener,
          options.instanceID,
          shoot.name,
          customParams,
          updateTimeout,
          null,
          false);
      shoot = skr.shoot;
    });

    it('Should check machine type after update', async function() {
      const machineType = await getMachineType(gardener, shoot.name);
      expect(machineType).to.equal(updateMachineType);
    });
  });
}

async function getMachineType(gardener, shoot) {
  await gardener.waitForShoot(shoot, 'Reconcile');
  const sh = await gardener.getShoot(shoot);
  return sh.spec.provider.workers[0].machine.type;
}

module.exports = {
  machineTypeE2ETest,
};
